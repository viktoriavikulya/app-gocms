package rest

import (
	"net/http"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
)

func (h Handler) requirePrincipal(r *http.Request) (appauthn.Principal, int, any, bool) {
	principal, ok := h.principal(r)
	if !ok {
		status, payload := unauthorized()
		return appauthn.Principal{}, status, payload, false
	}
	return principal, 0, nil, true
}
