package projectcache

import (
	"context"
	"fmt"
	"sync"
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
	c := &C{
		projectLister: projectLister,
		log:           log.WithName("projectcache"),
		projectsByID:  make(map[string]*v1.Project),
	}

	c.readyCond = sync.NewCond(&c.mu)

	return c
}

// C is a cache for project data.
type C struct {
	projectLister projectLister
	log           logr.Logger

	projectsByID       map[string]*v1.Project
	ready              bool
	lastSuccessfulSync time.Time

	mu        sync.Mutex
	readyCond *sync.Cond
}

// Run starts the project cache.
func (c *C) Run(ctx context.Context, interval time.Duration) error {
	c.log.Info("Starting project cache...")

	if err := c.sync(ctx); err != nil {
		return fmt.Errorf("sync project cache: %s", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.log.Info("Stopping project cache...")
			return ctx.Err()
		case <-ticker.C:
			if err := c.sync(ctx); err != nil {
				// Gracefully handle the error.
				c.log.Error(err, "Failed to sync project cache")
			}
		}
	}
}

func (c *C) sync(ctx context.Context) error {
	resp, err := c.projectLister.ListProjects(ctx, &uv1.ListProjectsRequest{})
	if err != nil {
		return fmt.Errorf("list projects: %s", err)
	}

	projectsByID := make(map[string]*v1.Project)
	for _, p := range resp.Projects {
		converted := convertProject(p)
		projectsByID[converted.Id] = converted
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.projectsByID = projectsByID
	c.lastSuccessfulSync = time.Now()
	c.ready = true

	c.readyCond.Broadcast()

	c.log.Info("Projects synced", "count", len(c.projectsByID))

	return nil
}

// GetProject retrieves a project by its ID from the cache.
func (c *C) GetProject(projectID string) (*v1.Project, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for !c.ready {
		c.log.Info("Waiting for project cache to be ready...")
		c.readyCond.Wait()
		c.log.Info("Project cache is ready")
	}

	p, ok := c.projectsByID[projectID]
	if !ok {
		return nil, fmt.Errorf("project with ID %s not found", projectID)
	}
	return p, nil
}

func convertProject(p *uv1.Project) *v1.Project {
	return &v1.Project{
		Id: p.Id,

		Assignments:         convertAssignments(p.Assignments),
		KubernetesNamespace: p.KubernetesNamespace,
	}
}

func convertAssignments(as []*uv1.ProjectAssignment) []*v1.ProjectAssignment {
	converted := make([]*v1.ProjectAssignment, len(as))
	for i, a := range as {
		converted[i] = &v1.ProjectAssignment{
			ClusterId:      a.ClusterId,
			Namespace:      a.Namespace,
			KueueQueueName: a.KueueQueueName,
			NodeSelector:   convertNodeSelector(a.NodeSelector),
		}
	}
	return converted
}

func convertNodeSelector(ns []*uv1.ProjectAssignment_NodeSelector) []*v1.ProjectAssignment_NodeSelector {
	converted := make([]*v1.ProjectAssignment_NodeSelector, len(ns))
	for i, n := range ns {
		converted[i] = &v1.ProjectAssignment_NodeSelector{
			Key:   n.Key,
			Value: n.Value,
		}
	}

	return converted
}
