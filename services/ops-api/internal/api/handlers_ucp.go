package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/synq/pkg/authcontext"
)

type UCPHandler struct{}

func NewUCPHandler() *UCPHandler {
	return &UCPHandler{}
}

func (h *UCPHandler) RegisterRoutes(r chi.Router) {
	// The AI agent hits this static URL, but our middleware will ensure it's authenticated
	// and has the 4 Identity Pillars (TenantID, OrgID, UserID, Role) injected into the context.
	r.Get("/api/v1/ucp/catalog.json", h.HandleUCPRedirect)
}

// HandleUCPRedirect securely routes the AI crawler to the highly-available static GCS file
func (h *UCPHandler) HandleUCPRedirect(w http.ResponseWriter, r *http.Request) {
	// 1. Secure Context Extraction
	// The LLM crawler (or Merchant Center) must provide an API Key that our auth middleware
	// uses to definitively resolve the Tenant ID.
	tenantID, err := authcontext.GetTenantID(r.Context())
	if err != nil || tenantID == "" {
		http.Error(w, "Unauthorized: AI Agent must provide a valid API key linked to a Tenant", http.StatusUnauthorized)
		return
	}

	// 2. Resolve the GCS Bucket path
	// E.g., https://storage.googleapis.com/ucp-feeds/tenant_XYZ_ucp_catalog.json
	bucketName := os.Getenv("UCP_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "ucp-feeds"
	}

	objectName := fmt.Sprintf("tenant_%s_ucp_catalog.json", tenantID)

	var gcsURL string
	if emulatorHost := os.Getenv("STORAGE_EMULATOR_HOST"); emulatorHost != "" {
		// Local development mapping
		gcsURL = fmt.Sprintf("http://%s/%s/%s", emulatorHost, bucketName, objectName)
	} else {
		// Production GCS mapping
		gcsURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
	}

	// 3. Issue the 302 Found Redirect
	// This instructs the AI crawler to download the massive JSON-LD file from the edge CDN,
	// keeping our primary API and Database completely insulated from the bandwidth load.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // The redirect URL is dynamic, do not cache the 302
	http.Redirect(w, r, gcsURL, http.StatusFound)
}
