package server

import "context"

// NewWithError composes the existing engine without terminating the caller on
// startup failure, allowing the portable runtime to release its owned database.
func NewWithError(ctx context.Context, m *Manager, skillDir, dataDir, keyDir string) (*Server, error) {
	return newServer(ctx, m, skillDir, dataDir, keyDir)
}
