package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fastygo/app-gocms/pkg/module/codex"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func result(status int, payload any, err error) (int, any) {
	if err != nil {
		return errorResponse(err)
	}
	return status, payload
}

func errorResponse(err error) (int, any) {
	message := err.Error()
	switch {
	case strings.Contains(message, "not found"):
		return notFound(message)
	case strings.Contains(message, "required"), strings.Contains(message, "invalid"), strings.Contains(message, "unsupported"):
		return validationError(message)
	default:
		return serverError(err)
	}
}

func notFound(message string) (int, any) {
	return http.StatusNotFound, apiError(http.StatusNotFound, "not_found", message)
}

func unauthorized() (int, any) {
	return http.StatusUnauthorized, apiError(http.StatusUnauthorized, "unauthorized", "authorization required")
}

func forbidden() (int, any) {
	return http.StatusForbidden, apiError(http.StatusForbidden, "forbidden", "missing capability")
}

func validationError(message string) (int, any) {
	return http.StatusBadRequest, apiError(http.StatusBadRequest, "validation_error", message)
}

func serverError(err error) (int, any) {
	return http.StatusInternalServerError, apiError(http.StatusInternalServerError, "internal_error", err.Error())
}

func apiError(status int, code string, message string) codex.ErrorEnvelope {
	return codex.ErrorEnvelope{Error: codex.ErrorBody{Code: code, Message: message, Status: status}}
}
