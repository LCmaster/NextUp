package services

import (
	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/ws"
)

// ProjectService encapsulates business logic for projects.
// Currently it is a thin wrapper kept for architectural consistency and to
// provide a seam for future logic (e.g. ownership validation, soft-delete).
type ProjectService struct {
	queries *db.Queries
	hub     *ws.Hub
}

// NewProjectService creates a ProjectService.
func NewProjectService(queries *db.Queries, hub *ws.Hub) *ProjectService {
	return &ProjectService{queries: queries, hub: hub}
}
