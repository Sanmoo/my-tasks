package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetProjects(ctx context.Context, namesOrAliases []string) ([]*Project, error) {
	args := m.Called(ctx, namesOrAliases)
	return args.Get(0).([]*Project), args.Error(1)
}

func (m *MockRepository) GetAllProjects(ctx context.Context) ([]*Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Project), args.Error(1)
}

func TestNewService(t *testing.T) {
	mockRepo := &MockRepository{}
	service := NewService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
}

func TestService_GetProjectsByNamesOrAliases(t *testing.T) {
	ctx := context.Background()

	t.Run("successful project retrieval", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewService(mockRepo)

		// Create mock projects
		project1, err := NewProject("Project A")
		require.NoError(t, err)

		project2, err := NewProject("Project B")
		require.NoError(t, err)

		expectedProjects := []*Project{project1, project2}

		// Set up mock expectations
		mockRepo.On("GetProjects", ctx, []string{"project-a", "project-b"}).Return(expectedProjects, nil)

		// Call the method
		projects, err := service.GetProjectsByNamesOrAliases(ctx, []string{"project-a", "project-b"})

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, expectedProjects, projects)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewService(mockRepo)

		// Set up mock to return error
		mockRepo.On("GetProjects", ctx, []string{"project-a"}).Return([]*Project{}, assert.AnError)

		// Call the method
		projects, err := service.GetProjectsByNamesOrAliases(ctx, []string{"project-a"})

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, projects)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty project names", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewService(mockRepo)

		// Set up mock expectations for empty slice
		mockRepo.On("GetProjects", ctx, []string{}).Return([]*Project{}, nil)

		// Call the method
		projects, err := service.GetProjectsByNamesOrAliases(ctx, []string{})

		// Assertions
		assert.NoError(t, err)
		assert.Empty(t, projects)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetAllProjects(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval of all projects", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewService(mockRepo)

		// Create mock projects
		project1, err := NewProject("Project A")
		require.NoError(t, err)

		project2, err := NewProject("Project B")
		require.NoError(t, err)

		expectedProjects := []*Project{project1, project2}

		// Set up mock expectations
		mockRepo.On("GetAllProjects", ctx).Return(expectedProjects, nil)

		// Call the method
		projects, err := service.GetAllProjects(ctx)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, expectedProjects, projects)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewService(mockRepo)

		// Set up mock to return error
		mockRepo.On("GetAllProjects", ctx).Return([]*Project{}, assert.AnError)

		// Call the method
		projects, err := service.GetAllProjects(ctx)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, projects)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no projects found", func(t *testing.T) {
		mockRepo := &MockRepository{}
		service := NewService(mockRepo)

		// Set up mock to return empty slice
		mockRepo.On("GetAllProjects", ctx).Return([]*Project{}, nil)

		// Call the method
		projects, err := service.GetAllProjects(ctx)

		// Assertions
		assert.NoError(t, err)
		assert.Empty(t, projects)
		mockRepo.AssertExpectations(t)
	})
}
