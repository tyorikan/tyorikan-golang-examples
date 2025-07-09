
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/spanner"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"google.golang.org/api/option"
)

// --- Data Structures ---

// ProductRequest defines the structure for incoming API requests.
type ProductRequest struct {
	Operation   string          `json:"operation" validate:"required,oneof=CREATE UPDATE DELETE"`
	ProductID   string          `json:"product_id" validate:"required"`
	ProductName string          `json:"product_name" validate:"required_if=Operation CREATE,required_if=Operation UPDATE"`
	Description string          `json:"description"`
	Attributes  json.RawMessage `json:"attributes"`
}

// AuditLogMessage is the structure for messages sent to Pub/Sub.
type AuditLogMessage struct {
	Operation   string      `json:"operation"`
	ProductID   string      `json:"product_id"`
	RequestData any         `json:"request_data"`
	Timestamp   time.Time   `json:"timestamp"`
}

// --- Interfaces for Testability ---

// Storer defines the interface for database operations.
type Storer interface {
	CreateProduct(ctx context.Context, p *ProductRequest) error
	UpdateProduct(ctx context.Context, p *ProductRequest) error
	DeleteProduct(ctx context.Context, p *ProductRequest) error
}

// Publisher defines the interface for publishing messages.
type Publisher interface {
	PublishEvent(ctx context.Context, msg *AuditLogMessage) error
}

// --- Spanner Implementation ---

// SpannerStore implements the Storer interface for Google Cloud Spanner.
type SpannerStore struct {
	Client *spanner.Client
}

func (s *SpannerStore) CreateProduct(ctx context.Context, p *ProductRequest) error {
	_, err := s.Client.Apply(ctx, []*spanner.Mutation{
		spanner.Insert("Products",
			[]string{"ProductId", "ProductName", "Description", "Attributes", "CreatedAt", "UpdatedAt"},
			[]any{p.ProductID, p.ProductName, p.Description, spanner.NullJSON{Value: p.Attributes, Valid: true}, spanner.CommitTimestamp, spanner.CommitTimestamp},
		),
	})
	return err
}

func (s *SpannerStore) UpdateProduct(ctx context.Context, p *ProductRequest) error {
	_, err := s.Client.Apply(ctx, []*spanner.Mutation{
		spanner.Update("Products",
			[]string{"ProductId", "ProductName", "Description", "Attributes", "UpdatedAt"},
			[]any{p.ProductID, p.ProductName, p.Description, spanner.NullJSON{Value: p.Attributes, Valid: true}, spanner.CommitTimestamp},
		),
	})
	return err
}

func (s *SpannerStore) DeleteProduct(ctx context.Context, p *ProductRequest) error {
	_, err := s.Client.Apply(ctx, []*spanner.Mutation{
		spanner.Delete("Products", spanner.Key{p.ProductID}),
	})
	return err
}

// --- Pub/Sub Implementation ---

// PubSubPublisher implements the Publisher interface for Google Cloud Pub/Sub.
type PubSubPublisher struct {
	Topic *pubsub.Topic
}

func (pub *PubSubPublisher) PublishEvent(ctx context.Context, msg *AuditLogMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal audit message: %w", err)
	}

	result := pub.Topic.Publish(ctx, &pubsub.Message{Data: data})
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// --- Server and Handlers ---

// Server holds the dependencies for the HTTP server.
type Server struct {
	DB        Storer
	Publisher Publisher
	Validator *validator.Validate
	Router    *chi.Mux
}

func NewServer(db Storer, pub Publisher) *Server {
	s := &Server{
		DB:        db,
		Publisher: pub,
		Validator: validator.New(),
		Router:    chi.NewRouter(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.Router.Use(middleware.Logger)
	s.Router.Post("/products", s.handleProductRequest())
}

func (s *Server) handleProductRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.Validator.Struct(&req); err != nil {
			s.sendErrorResponse(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
			return
		}

		var dbErr error
		ctx := r.Context()

		switch req.Operation {
		case "CREATE":
			dbErr = s.DB.CreateProduct(ctx, &req)
		case "UPDATE":
			dbErr = s.DB.UpdateProduct(ctx, &req)
		case "DELETE":
			dbErr = s.DB.DeleteProduct(ctx, &req)
		default:
			s.sendErrorResponse(w, "Invalid operation", http.StatusBadRequest)
			return
		}

		if dbErr != nil {
			log.Printf("Database operation failed: %v", dbErr)
			s.sendErrorResponse(w, "Database operation failed", http.StatusInternalServerError)
			return
		}

		auditMsg := &AuditLogMessage{
			Operation:   req.Operation,
			ProductID:   req.ProductID,
			RequestData: req,
			Timestamp:   time.Now().UTC(),
		}
		if err := s.Publisher.PublishEvent(ctx, auditMsg); err != nil {
			// Log the error but don't fail the request, as the primary operation succeeded.
			log.Printf("WARN: Failed to publish audit event: %v", err)
		}

		s.sendSuccessResponse(w, req.Operation, req.ProductID)
	}
}

func (s *Server) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": message})
}

func (s *Server) sendSuccessResponse(w http.ResponseWriter, operation, productID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK for all successful operations
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"operation":  operation,
		"product_id": productID,
	})
}

// --- Main Application Setup ---

func main() {
	ctx := context.Background()

	// Env variables
	projectID := os.Getenv("GCP_PROJECT_ID")
	instanceID := os.Getenv("SPANNER_INSTANCE_ID")
	databaseID := os.Getenv("SPANNER_DATABASE_ID")
	topicID := os.Getenv("PUBSUB_TOPIC_ID")

	if projectID == "" || instanceID == "" || databaseID == "" || topicID == "" {
		log.Fatal("GCP_PROJECT_ID, SPANNER_INSTANCE_ID, SPANNER_DATABASE_ID, and PUBSUB_TOPIC_ID environment variables must be set.")
	}

	// Spanner client
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)
	spannerClient, err := spanner.NewClient(ctx, dbPath, option.WithUserAgent("spanner-demo/2.0"))
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer spannerClient.Close()
	store := &SpannerStore{Client: spannerClient}

	// Pub/Sub client
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer pubsubClient.Close()
	topic := pubsubClient.Topic(topicID)
	publisher := &PubSubPublisher{Topic: topic}

	// Server setup
	server := NewServer(store, publisher)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, server.Router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
