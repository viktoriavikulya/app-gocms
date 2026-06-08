package app

import (
	"context"
	"sync"

	"github.com/fastygo/platform/pkg/bff"
	"github.com/fastygo/platform/pkg/contracts"
)

type memoryActionAuditor struct {
	mu     sync.Mutex
	events []contracts.AuditEvent
}

func (a *memoryActionAuditor) RecordActionAudit(_ context.Context, event contracts.AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, event)
	return nil
}

var cmsActionAuditor = &memoryActionAuditor{}

func cmsActionAudit() bff.AuditRecorder {
	return cmsActionAuditor
}
