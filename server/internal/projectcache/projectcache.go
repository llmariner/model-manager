package projectcache

import (
	"context"

	v1 "github.com/llmariner/model-manager/api/v1"
)

// New creates a new project cache instance.
func New() *C {
	return &C{}
}

// C is a cache for project data.
type C struct {
}

// Run starts the project cache.
func (c *C) Run(ctx context.Context) error {
	// TODO(kenji): Implement.
	return nil
}

// WaitForInitialSync waits for the initial sync to complete.
func (c *C) WaitForInitialSync(ctx context.Context) error {
	// TODO(kenji): Implement.
	return nil
}

// GetProject retrieves a project by its ID from the cache.
func (c *C) GetProject(projectID string) (*v1.Project, error) {
	// TODO(kenji): Implement.
	return nil, nil
}
