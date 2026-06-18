package importexport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
)

// Eventarc Payload for google.cloud.storage.object.v1.finalized
type StorageObjectData struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

type EventarcHandler struct {
	manipulator   *Manipulator
	storageClient *storage.Client
}

func NewEventarcHandler(m *Manipulator) (*EventarcHandler, error) {
	ctx := context.Background()
	// Initialize GCS client. For local dev, this might need GOOGLE_APPLICATION_CREDENTIALS
	// or fallback to a mock.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("Warning: Failed to create storage client (local dev?): %v", err)
		// We don't return error here so local dev can boot without GCS credentials
	}

	return &EventarcHandler{
		manipulator:   m,
		storageClient: client,
	}, nil
}

// HandleImportWebhook is invoked by Google Cloud Eventarc when a new CSV is uploaded.
// We expect standard CloudEvents over HTTP.
func (h *EventarcHandler) HandleImportWebhook(w http.ResponseWriter, r *http.Request) {
	// Eventarc sends CloudEvents headers. We can check Ce-Type.
	ceType := r.Header.Get("Ce-Type")
	if ceType != "google.cloud.storage.object.v1.finalized" && ceType != "" {
		log.Printf("Ignored event type: %s", ceType)
		w.WriteHeader(http.StatusOK)
		return
	}

	var data StorageObjectData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	log.Printf("Received Eventarc notification for gs://%s/%s", data.Bucket, data.Name)

	if h.storageClient == nil {
		http.Error(w, "Storage client not initialized", http.StatusInternalServerError)
		return
	}

	// 1. Open the file stream from Cloud Storage
	ctx := r.Context()
	obj := h.storageClient.Bucket(data.Bucket).Object(data.Name)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		log.Printf("Failed to read from GCS: %v", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// 2. Extract tenant routing info from path: imports/{org_id}/{tenant_id}/filename
	parts := strings.Split(data.Name, "/")
	if len(parts) < 4 || parts[0] != "imports" {
		log.Printf("Invalid file path format: %s", data.Name)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	orgID := parts[1]
	tenantID := parts[2]
	userID := "system"

	// 3. Process the file via Data Manipulator (Streaming)
	if err := h.manipulator.ProcessProductsCSV(ctx, tenantID, orgID, userID, reader); err != nil {
		log.Printf("Failed to process CSV: %v", err)
		http.Error(w, "Failed to process CSV", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
