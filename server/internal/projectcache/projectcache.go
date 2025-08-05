package projectcache

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	v1 "github.com/llmariner/model-manager/api/v1"
	uv1 "github.com/llmariner/user-manager/api/v1"
	"google.golang.org/grpc"
)

type projectLister interface {
	ListProjects(context.Context, *uv1.ListProjectsRequest, ...grpc.CallOption) (*uv1.ListProjectsResponse, error)
}

// New creates a new project cache instance.
func New(projectLister projectLister, log logr.Logger) *C {
	return &C{
		projectLister: projectLister,
		log:           log.WithName("projectcache"),
	}
}

// C is a cache for project data.
type C struct {
	projectLister projectLister
	log           logr.Logger
}

// Run starts the project cache.
func (c *C) Run(ctx context.Context, interval time.Duration) error {
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
