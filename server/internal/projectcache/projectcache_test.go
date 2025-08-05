package projectcache

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	uv1 "github.com/llmariner/user-manager/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProjectCache(t *testing.T) {
	lister := &fakeProjectLister{
		projects: []*uv1.Project{
			{
				Id: "proj0",
				Assignments: []*uv1.ProjectAssignment{
					{
						NodeSelector: []*uv1.ProjectAssignment_NodeSelector{
							{
								Key:   "key0",
								Value: "value0",
							},
						},
					},
				},
			},
			{
				Id: "proj1",
				Assignments: []*uv1.ProjectAssignment{
					{
						NodeSelector: []*uv1.ProjectAssignment_NodeSelector{
							{
								Key:   "key1",
								Value: "value1",
							},
						},
					},
				},
			},
		},
	}

	c := New(lister, testr.New(t))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	done := make(chan struct{})
	go func() {
		if err := c.Run(ctx, 1*time.Second); err != nil {
			assert.ErrorIs(t, err, context.Canceled)
		}
		close(done)
	}()

	err := c.WaitForInitialSync(ctx)
	assert.NoError(t, err)

	got, err := c.GetProject("proj0")
	assert.NoError(t, err)
	ns := got.Assignments[0].NodeSelector[0]
	assert.Equal(t, "key0", ns.Key)
	assert.Equal(t, "value0", ns.Value)

	got, err = c.GetProject("proj1")
	assert.NoError(t, err)
	ns = got.Assignments[0].NodeSelector[0]
	assert.Equal(t, "key1", ns.Key)
	assert.Equal(t, "value1", ns.Value)

	_, err = c.GetProject("proj2")
	assert.Error(t, err)

	cancel()
	<-done
}

func TestProjectCache_SyncError(t *testing.T) {
	lister := &fakeProjectLister{
		hasErr: true,
	}

	c := New(lister, testr.New(t))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	done := make(chan struct{})
	go func() {
		if err := c.Run(ctx, 1*time.Second); err != nil {
			assert.ErrorIs(t, err, context.Canceled)
		}
		close(done)
	}()

	err := c.WaitForInitialSync(ctx)
	assert.Error(t, err)

	cancel()
	<-done
}

type fakeProjectLister struct {
	projects []*uv1.Project
	hasErr   bool
}

func (l *fakeProjectLister) ListProjects(context.Context, *uv1.ListProjectsRequest, ...grpc.CallOption) (*uv1.ListProjectsResponse, error) {
	if l.hasErr {
		return nil, status.Errorf(codes.Internal, "fake error")
	}

	return &uv1.ListProjectsResponse{
		Projects: l.projects,
	}, nil
}
