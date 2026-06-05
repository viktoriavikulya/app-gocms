package authn_test

import (
	"testing"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/capcheck"
)

func TestEditorCanPublishAuthorCannotArchive(t *testing.T) {
	store, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	service := appauthn.NewService(store)

	editor, ok := service.Principal("editor")
	if !ok {
		t.Fatal("missing editor user")
	}
	if !capcheck.CanPublish(editor) {
		t.Fatal("editor should publish")
	}

	author, ok := service.Principal("author")
	if !ok {
		t.Fatal("missing author user")
	}
	if !capcheck.CanPublish(author) {
		t.Fatal("author should publish own drafts")
	}
	if capcheck.CanArchive(author) {
		t.Fatal("author should not archive")
	}
}

func TestAuthorEditOwnNotOthers(t *testing.T) {
	store, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	service := appauthn.NewService(store)
	author, ok := service.Principal("author")
	if !ok {
		t.Fatal("missing author user")
	}
	if !capcheck.CanEdit(author, "author") {
		t.Fatal("author should edit own content")
	}
	if capcheck.CanEdit(author, "someone-else") {
		t.Fatal("author should not edit others content")
	}
}

func TestBuiltInRolesExposeGranularCapabilities(t *testing.T) {
	roles := appauthn.BuiltInRoles()
	if len(roles[appauthn.RoleContributor]) == 0 {
		t.Fatal("contributor role missing")
	}
	foundCreate := false
	for _, cap := range roles[appauthn.RoleContributor] {
		if cap == modulecms.CapabilityContentCreate {
			foundCreate = true
		}
		if cap == modulecms.CapabilityContentPublish {
			t.Fatal("contributor must not publish")
		}
	}
	if !foundCreate {
		t.Fatal("contributor missing content.create")
	}
}
