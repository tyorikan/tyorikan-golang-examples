package main

import (
	"bytes"
	"context"
	
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type MockStorer struct {
	CreateFunc func(ctx context.Context, p *ProductRequest) error
	UpdateFunc func(ctx context.Context, p *ProductRequest) error
	DeleteFunc func(ctx context.Context, p *ProductRequest) error
}

func (m *MockStorer) CreateProduct(ctx context.Context, p *ProductRequest) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, p)
	}
	return nil
}
func (m *MockStorer) UpdateProduct(ctx context.Context, p *ProductRequest) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, p)
	}
	return nil
}
func (m *MockStorer) DeleteProduct(ctx context.Context, p *ProductRequest) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, p)
	}
	return nil
}

type MockPublisher struct {
	PublishedMessages []*AuditLogMessage
	PublishFunc       func(ctx context.Context, msg *AuditLogMessage) error
}

func (m *MockPublisher) PublishEvent(ctx context.Context, msg *AuditLogMessage) error {
	m.PublishedMessages = append(m.PublishedMessages, msg)
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, msg)
	}
	return nil
}

// --- Test Cases ---

func TestProductHandler_CreateSuccess(t *testing.T) {
	// Setup
	mockStore := &MockStorer{}
	mockPublisher := &MockPublisher{}
	server := NewServer(mockStore, mockPublisher)

	body := `{"operation": "CREATE", "product_id": "p-123", "product_name": "Test Item"}`
	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	// Execute
	server.Router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"success"`)
	assert.Contains(t, rr.Body.String(), `"operation":"CREATE"`)
	require.Len(t, mockPublisher.PublishedMessages, 1)
	assert.Equal(t, "CREATE", mockPublisher.PublishedMessages[0].Operation)
	assert.Equal(t, "p-123", mockPublisher.PublishedMessages[0].ProductID)
}

func TestProductHandler_UpdateSuccess(t *testing.T) {
	// Setup
	mockStore := &MockStorer{}
	mockPublisher := &MockPublisher{}
	server := NewServer(mockStore, mockPublisher)

	body := `{"operation": "UPDATE", "product_id": "p-123", "product_name": "Updated Item"}`
	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	// Execute
	server.Router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, mockPublisher.PublishedMessages, 1)
	assert.Equal(t, "UPDATE", mockPublisher.PublishedMessages[0].Operation)
}

func TestProductHandler_DeleteSuccess(t *testing.T) {
	// Setup
	mockStore := &MockStorer{}
	mockPublisher := &MockPublisher{}
	server := NewServer(mockStore, mockPublisher)

	body := `{"operation": "DELETE", "product_id": "p-123"}`
	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	// Execute
	server.Router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, mockPublisher.PublishedMessages, 1)
	assert.Equal(t, "DELETE", mockPublisher.PublishedMessages[0].Operation)
}

func TestProductHandler_ValidationFailure(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{"Missing Operation", `{"product_id": "p-123", "product_name": "Test Item"}`},
		{"Invalid Operation", `{"operation": "READ", "product_id": "p-123"}`},
		{"Missing ProductID", `{"operation": "CREATE", "product_name": "Test Item"}`},
		{"Missing ProductName for CREATE", `{"operation": "CREATE", "product_id": "p-123"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockStore := &MockStorer{}
			mockPublisher := &MockPublisher{}
			server := NewServer(mockStore, mockPublisher)
			req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(tc.body))
			rr := httptest.NewRecorder()

			// Execute
			server.Router.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, rr.Code)
			assert.Contains(t, rr.Body.String(), `"status":"error"`)
			assert.Empty(t, mockPublisher.PublishedMessages) // Nothing should be published
		})
	}
}

func TestProductHandler_DatabaseFailure(t *testing.T) {
	// Setup
	dbError := errors.New("db is down")
	mockStore := &MockStorer{
		CreateFunc: func(ctx context.Context, p *ProductRequest) error {
			return dbError
		},
	}
	mockPublisher := &MockPublisher{}
	server := NewServer(mockStore, mockPublisher)

	body := `{"operation": "CREATE", "product_id": "p-123", "product_name": "Test Item"}`
	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	// Execute
	server.Router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Empty(t, mockPublisher.PublishedMessages)
}

func TestProductHandler_PublisherFailure(t *testing.T) {
	// Setup
	mockStore := &MockStorer{}
	mockPublisher := &MockPublisher{
		PublishFunc: func(ctx context.Context, msg *AuditLogMessage) error {
			return errors.New("pubsub is down")
		},
	}
	server := NewServer(mockStore, mockPublisher)

	body := `{"operation": "CREATE", "product_id": "p-123", "product_name": "Test Item"}`
	req := httptest.NewRequest("POST", "/products", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	// Execute
	server.Router.ServeHTTP(rr, req)

	// Assert: The request should still succeed even if the audit log fails
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"success"`)
	// The message was still added to the mock slice before the error was returned
	assert.Len(t, mockPublisher.PublishedMessages, 1)
}