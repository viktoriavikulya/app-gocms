package authn

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fastygo/platform/pkg/contracts"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLoginLocked        = errors.New("login temporarily locked")
	ErrForbidden          = errors.New("missing capability")
	ErrAppTokenFailed     = errors.New("app token is invalid or unavailable")
)

type Principal struct {
	PrincipalID  contracts.PrincipalID
	Roles        []string
	Capabilities map[contracts.CapabilityID]struct{}
}

func (p Principal) ID() contracts.PrincipalID {
	return p.PrincipalID
}

func (p Principal) Has(capability contracts.CapabilityID) bool {
	if capability == "" {
		return true
	}
	_, ok := p.Capabilities[capability]
	return ok
}

type User struct {
	ID           contracts.PrincipalID
	Email        string
	PasswordHash string
	Roles        []string
	Active       bool
}

type AppToken struct {
	ID           string
	UserID       contracts.PrincipalID
	Prefix       string
	Hash         string
	Capabilities []contracts.CapabilityID
	ExpiresAt    time.Time
	RevokedAt    *time.Time
}

type LoginAttempt struct {
	Identifier string
	RemoteAddr string
	Success    bool
	CreatedAt  time.Time
}

type MemoryStore struct {
	mu       sync.RWMutex
	users    map[contracts.PrincipalID]User
	tokens   map[string]AppToken
	attempts []LoginAttempt
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:  map[contracts.PrincipalID]User{},
		tokens: map[string]AppToken{},
	}
}

func NewSeededMemoryStore() (*MemoryStore, error) {
	store := NewMemoryStore()
	service := NewService(store)
	for _, seed := range []struct {
		id       contracts.PrincipalID
		email    string
		password string
		roles    []string
	}{
		{id: "admin", email: "admin@example.test", password: "admin", roles: []string{RoleAdmin}},
		{id: "editor", email: "editor@example.test", password: "editor", roles: []string{RoleEditor}},
		{id: "author", email: "author@example.test", password: "author", roles: []string{RoleAuthor}},
		{id: "viewer", email: "viewer@example.test", password: "viewer", roles: []string{RoleViewer}},
	} {
		hash, err := service.hasher.Hash(seed.password)
		if err != nil {
			return nil, err
		}
		store.users[seed.id] = User{ID: seed.id, Email: seed.email, PasswordHash: hash, Roles: seed.roles, Active: true}
	}
	return store, nil
}

func (s *MemoryStore) UserByIdentifier(identifier string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	for _, user := range s.users {
		if strings.ToLower(string(user.ID)) == identifier || strings.ToLower(user.Email) == identifier {
			return user, true
		}
	}
	return User{}, false
}

func (s *MemoryStore) User(id contracts.PrincipalID) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

func (s *MemoryStore) SaveAttempt(attempt LoginAttempt) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts = append(s.attempts, attempt)
}

func (s *MemoryStore) RecentFailures(identifier string, since time.Time) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	failures := 0
	for _, attempt := range s.attempts {
		if attempt.Identifier == identifier && !attempt.Success && !attempt.CreatedAt.Before(since) {
			failures++
		}
	}
	return failures
}

func (s *MemoryStore) SaveToken(token AppToken) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token.Prefix] = token
}

func (s *MemoryStore) Token(prefix string) (AppToken, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	token, ok := s.tokens[prefix]
	return token, ok
}

type Service struct {
	store  *MemoryStore
	hasher PasswordHasher
	now    func() time.Time
}

func NewService(store *MemoryStore) Service {
	return Service{store: store, hasher: DefaultPasswordHasher(), now: time.Now}
}

func (s Service) AuthenticatePassword(_ context.Context, identifier string, password string, remoteAddr string) (Principal, error) {
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	if identifier == "" || strings.TrimSpace(password) == "" {
		return Principal{}, ErrInvalidCredentials
	}
	if s.store.RecentFailures(identifier, s.now().Add(-24*time.Hour)) >= 3 {
		return Principal{}, ErrLoginLocked
	}
	user, ok := s.store.UserByIdentifier(identifier)
	if !ok || !user.Active || !s.hasher.Verify(password, user.PasswordHash) {
		s.store.SaveAttempt(LoginAttempt{Identifier: identifier, RemoteAddr: remoteAddr, CreatedAt: s.now().UTC()})
		return Principal{}, ErrInvalidCredentials
	}
	s.store.SaveAttempt(LoginAttempt{Identifier: identifier, RemoteAddr: remoteAddr, Success: true, CreatedAt: s.now().UTC()})
	return principalForUser(user), nil
}

func (s Service) Principal(id contracts.PrincipalID) (Principal, bool) {
	user, ok := s.store.User(id)
	if !ok || !user.Active {
		return Principal{}, false
	}
	return principalForUser(user), true
}

func (s Service) Require(principal Principal, capability contracts.CapabilityID) error {
	if !principal.Has(capability) {
		return ErrForbidden
	}
	return nil
}

func (s Service) CreateAppToken(ctx context.Context, userID contracts.PrincipalID, capabilities []contracts.CapabilityID, ttl time.Duration) (string, AppToken, error) {
	if _, ok := s.Principal(userID); !ok {
		return "", AppToken{}, fmt.Errorf("user %q not found", userID)
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	id, err := randomToken(12)
	if err != nil {
		return "", AppToken{}, err
	}
	secret, err := randomToken(24)
	if err != nil {
		return "", AppToken{}, err
	}
	raw := id + "." + secret
	token := AppToken{
		ID:           id,
		UserID:       userID,
		Prefix:       id,
		Hash:         tokenHash(raw),
		Capabilities: capabilities,
		ExpiresAt:    s.now().UTC().Add(ttl),
	}
	s.store.SaveToken(token)
	return raw, token, ctx.Err()
}

func (s Service) AuthenticateAppToken(_ context.Context, raw string) (Principal, error) {
	prefix, _, ok := strings.Cut(strings.TrimSpace(raw), ".")
	if !ok || prefix == "" {
		return Principal{}, ErrAppTokenFailed
	}
	token, ok := s.store.Token(prefix)
	if !ok || token.RevokedAt != nil || s.now().UTC().After(token.ExpiresAt) || token.Hash != tokenHash(raw) {
		return Principal{}, ErrAppTokenFailed
	}
	principal, ok := s.Principal(token.UserID)
	if !ok {
		return Principal{}, ErrAppTokenFailed
	}
	if len(token.Capabilities) > 0 {
		filtered := Principal{PrincipalID: principal.PrincipalID, Roles: principal.Roles, Capabilities: map[contracts.CapabilityID]struct{}{}}
		for _, capability := range token.Capabilities {
			if principal.Has(capability) {
				filtered.Capabilities[capability] = struct{}{}
			}
		}
		return filtered, nil
	}
	return principal, nil
}

func BuiltInRoles() map[string][]contracts.CapabilityID {
	return map[string][]contracts.CapabilityID{
		RoleAdmin: allAdminCapabilities(),
		RoleEditor: {
			capabilityAdminAccess,
			capabilityContentRead,
			capabilityContentWrite,
			capabilityContentCreate,
			capabilityContentEdit,
			capabilityContentEditOwn,
			capabilityContentEditOthers,
			capabilityContentPublish,
			capabilityContentSchedule,
			capabilityContentArchive,
			capabilityContentDelete,
			capabilityContentRestore,
			capabilityContentManageRevisions,
			capabilityContentPrivate,
			capabilityMediaUpload,
			capabilityMediaEdit,
			capabilityMediaDelete,
			capabilityTaxonomyAssign,
			capabilityRESTAccess,
			capabilityRESTAccessPrivate,
			capabilityRESTWrite,
		},
		RoleAuthor: {
			capabilityAdminAccess,
			capabilityContentRead,
			capabilityContentCreate,
			capabilityContentEditOwn,
			capabilityContentPublish,
			capabilityContentSchedule,
			capabilityContentPrivate,
			capabilityMediaUpload,
			capabilityRESTAccess,
			capabilityRESTAccessPrivate,
			capabilityRESTWrite,
		},
		RoleContributor: {
			capabilityAdminAccess,
			capabilityContentRead,
			capabilityContentCreate,
			capabilityContentEditOwn,
			capabilityRESTAccess,
			capabilityRESTWrite,
		},
		RoleViewer: {
			capabilityAdminAccess,
			capabilityContentRead,
			capabilityRESTAccess,
		},
	}
}

func allAdminCapabilities() []contracts.CapabilityID {
	return []contracts.CapabilityID{
		capabilityAdminAccess,
		capabilityContentRead,
		capabilityContentWrite,
		capabilityContentCreate,
		capabilityContentEdit,
		capabilityContentEditOwn,
		capabilityContentEditOthers,
		capabilityContentPublish,
		capabilityContentSchedule,
		capabilityContentArchive,
		capabilityContentDelete,
		capabilityContentRestore,
		capabilityContentManageRevisions,
		capabilityContentPrivate,
		capabilityMediaUpload,
		capabilityMediaEdit,
		capabilityMediaDelete,
		capabilityMediaReadPrivate,
		capabilityTaxonomyManage,
		capabilityTaxonomyAssign,
		capabilityThemesView,
		capabilityThemesActivate,
		capabilityThemesManageSettings,
		capabilityPluginsView,
		capabilityPluginsInstall,
		capabilityPluginsActivate,
		capabilityPluginsDeactivate,
		capabilityPluginsUninstall,
		capabilityPluginsManageSettings,
		capabilityUsersView,
		capabilityUsersCreate,
		capabilityUsersEdit,
		capabilityUsersDelete,
		capabilityUsersManage,
		capabilityRolesView,
		capabilityRolesManage,
		capabilitySettingsView,
		capabilitySettingsManage,
		capabilityRESTAccess,
		capabilityRESTAccessPrivate,
		capabilityRESTWrite,
	}
}

// Capability IDs mirror pkg/module/capabilities.go; kept local to avoid pulling panel/render deps into application/authn.
const (
	capabilityAdminAccess            contracts.CapabilityID = "admin.access"
	capabilityContentRead            contracts.CapabilityID = "content.read"
	capabilityContentWrite           contracts.CapabilityID = "content.write"
	capabilityContentCreate          contracts.CapabilityID = "content.create"
	capabilityContentEdit            contracts.CapabilityID = "content.edit"
	capabilityContentEditOwn         contracts.CapabilityID = "content.edit_own"
	capabilityContentEditOthers      contracts.CapabilityID = "content.edit_others"
	capabilityContentPublish         contracts.CapabilityID = "content.publish"
	capabilityContentSchedule        contracts.CapabilityID = "content.schedule"
	capabilityContentArchive         contracts.CapabilityID = "content.archive"
	capabilityContentDelete          contracts.CapabilityID = "content.delete"
	capabilityContentRestore         contracts.CapabilityID = "content.restore"
	capabilityContentManageRevisions contracts.CapabilityID = "content.manage_revisions"
	capabilityContentPrivate         contracts.CapabilityID = "content.read_private"
	capabilityMediaUpload            contracts.CapabilityID = "media.upload"
	capabilityMediaEdit              contracts.CapabilityID = "media.edit"
	capabilityMediaDelete            contracts.CapabilityID = "media.delete"
	capabilityMediaReadPrivate       contracts.CapabilityID = "media.read_private"
	capabilityTaxonomyManage         contracts.CapabilityID = "taxonomies.manage"
	capabilityTaxonomyAssign         contracts.CapabilityID = "taxonomies.assign"
	capabilityThemesView             contracts.CapabilityID = "themes.view"
	capabilityThemesActivate         contracts.CapabilityID = "themes.activate"
	capabilityThemesManageSettings   contracts.CapabilityID = "themes.manage_settings"
	capabilityPluginsView            contracts.CapabilityID = "plugins.view"
	capabilityPluginsInstall         contracts.CapabilityID = "plugins.install"
	capabilityPluginsActivate        contracts.CapabilityID = "plugins.activate"
	capabilityPluginsDeactivate      contracts.CapabilityID = "plugins.deactivate"
	capabilityPluginsUninstall       contracts.CapabilityID = "plugins.uninstall"
	capabilityPluginsManageSettings contracts.CapabilityID = "plugins.manage_settings"
	capabilityUsersView              contracts.CapabilityID = "users.view"
	capabilityUsersCreate            contracts.CapabilityID = "users.create"
	capabilityUsersEdit              contracts.CapabilityID = "users.edit"
	capabilityUsersDelete            contracts.CapabilityID = "users.delete"
	capabilityUsersManage            contracts.CapabilityID = "users.manage"
	capabilityRolesView              contracts.CapabilityID = "roles.view"
	capabilityRolesManage            contracts.CapabilityID = "roles.manage"
	capabilitySettingsView           contracts.CapabilityID = "settings.view"
	capabilitySettingsManage         contracts.CapabilityID = "settings.manage"
	capabilityRESTAccess             contracts.CapabilityID = "rest.access"
	capabilityRESTAccessPrivate      contracts.CapabilityID = "rest.access_private"
	capabilityRESTWrite              contracts.CapabilityID = "rest.write"
)

const (
	RoleAdmin       = "admin"
	RoleEditor      = "editor"
	RoleAuthor      = "author"
	RoleContributor = "contributor"
	RoleViewer      = "viewer"
)

func principalForUser(user User) Principal {
	principal := Principal{PrincipalID: user.ID, Roles: slices.Clone(user.Roles), Capabilities: map[contracts.CapabilityID]struct{}{}}
	roles := BuiltInRoles()
	for _, role := range user.Roles {
		for _, capability := range roles[role] {
			principal.Capabilities[capability] = struct{}{}
		}
	}
	return principal
}

func randomToken(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func tokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
