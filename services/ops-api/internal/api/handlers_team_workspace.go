package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/authcontext"
)

type TeamWorkspaceHandlers struct {
	pool *pgxpool.Pool
}

func NewTeamWorkspaceHandlers(pool *pgxpool.Pool) *TeamWorkspaceHandlers {
	return &TeamWorkspaceHandlers{pool: pool}
}

func (h *TeamWorkspaceHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/team-workspace", h.GetWorkspace)
	r.Post("/api/v1/team-workspace/teams", h.CreateTeam)
	r.Post("/api/v1/team-workspace/tasks", h.CreateTask)
}

type teamWorkspaceResponse struct {
	Teams []humanAITeamResponse `json:"teams"`
	Tasks []aiTaskResponse      `json:"tasks"`
}

type humanAITeamResponse struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Status         string          `json:"status"`
	CadenceMinutes int32           `json:"cadence_minutes"`
	ApprovalPolicy json.RawMessage `json:"approval_policy"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

type aiTaskResponse struct {
	ID               string          `json:"id"`
	TeamID           string          `json:"team_id"`
	Title            string          `json:"title"`
	Status           string          `json:"status"`
	Priority         string          `json:"priority"`
	InputContext     json.RawMessage `json:"input_context"`
	ProposedOutput   json.RawMessage `json:"proposed_output"`
	RequiresApproval bool            `json:"requires_approval"`
	DueAt            string          `json:"due_at,omitempty"`
	StartedAt        string          `json:"started_at,omitempty"`
	CompletedAt      string          `json:"completed_at,omitempty"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

func (h *TeamWorkspaceHandlers) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	tenantUUID, ok := requiredTenantUUID(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "failed to open workspace query", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())
	if err := setWorkspaceRLS(r.Context(), tx, tenantUUID, orgUUID); err != nil {
		http.Error(w, "failed to set tenant context", http.StatusInternalServerError)
		return
	}

	teams, err := listHumanAITeams(r.Context(), tx, tenantUUID, orgUUID)
	if err != nil {
		http.Error(w, "failed to list teams", http.StatusInternalServerError)
		return
	}
	tasks, err := listAITasks(r.Context(), tx, tenantUUID, orgUUID)
	if err != nil {
		http.Error(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teamWorkspaceResponse{Teams: teams, Tasks: tasks})
}

type createTeamRequest struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	CadenceMinutes int32           `json:"cadence_minutes"`
	ApprovalPolicy json.RawMessage `json:"approval_policy"`
}

func (h *TeamWorkspaceHandlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	role, err := authcontext.GetRole(r.Context())
	if err != nil || (strings.ToUpper(role) != "ADMIN" && strings.ToUpper(role) != "MANAGER") {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}
	tenantUUID, ok := requiredTenantUUID(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}
	userUUID := optionalUserUUID(r)

	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.CadenceMinutes <= 0 {
		req.CadenceMinutes = 30
	}
	if len(req.ApprovalPolicy) == 0 {
		req.ApprovalPolicy = json.RawMessage(`{}`)
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "failed to open workspace transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())
	if err := setWorkspaceRLS(r.Context(), tx, tenantUUID, orgUUID); err != nil {
		http.Error(w, "failed to set tenant context", http.StatusInternalServerError)
		return
	}

	var team humanAITeamResponse
	err = tx.QueryRow(r.Context(), `
		INSERT INTO human_ai_teams (org_id, tenant_id, owner_user_id, name, description, cadence_minutes, approval_policy)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, name, COALESCE(description, ''), status, cadence_minutes, approval_policy, created_at::text, updated_at::text
	`, orgUUID, tenantUUID, userUUID, req.Name, nullableString(req.Description), req.CadenceMinutes, req.ApprovalPolicy).Scan(
		&team.ID, &team.Name, &team.Description, &team.Status, &team.CadenceMinutes,
		&team.ApprovalPolicy, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "failed to create team", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "failed to commit team", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

type createAITaskRequest struct {
	TeamID           string          `json:"team_id"`
	Title            string          `json:"title"`
	Priority         string          `json:"priority"`
	InputContext     json.RawMessage `json:"input_context"`
	RequiresApproval bool            `json:"requires_approval"`
	DueAt            string          `json:"due_at"`
}

func (h *TeamWorkspaceHandlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	role, err := authcontext.GetRole(r.Context())
	if err != nil || (strings.ToUpper(role) != "ADMIN" && strings.ToUpper(role) != "MANAGER" && strings.ToUpper(role) != "EDITOR") {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}
	tenantUUID, ok := requiredTenantUUID(w, r)
	if !ok {
		return
	}
	orgUUID, ok := requiredOrgUUID(w, r)
	if !ok {
		return
	}
	userUUID := optionalUserUUID(r)

	var req createAITaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	teamID, err := uuid.Parse(strings.TrimSpace(req.TeamID))
	if err != nil {
		http.Error(w, "valid team_id is required", http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	req.Priority = strings.ToUpper(strings.TrimSpace(req.Priority))
	if req.Priority == "" {
		req.Priority = "NORMAL"
	}
	if len(req.InputContext) == 0 {
		req.InputContext = json.RawMessage(`{}`)
	}
	var dueAt pgtype.Timestamptz
	if req.DueAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.DueAt)
		if err != nil {
			http.Error(w, "due_at must be RFC3339", http.StatusBadRequest)
			return
		}
		dueAt = pgtype.Timestamptz{Time: parsed, Valid: true}
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "failed to open task transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())
	if err := setWorkspaceRLS(r.Context(), tx, tenantUUID, orgUUID); err != nil {
		http.Error(w, "failed to set tenant context", http.StatusInternalServerError)
		return
	}

	var task aiTaskResponse
	err = tx.QueryRow(r.Context(), `
		INSERT INTO ai_tasks (org_id, tenant_id, team_id, created_by, title, priority, input_context, requires_approval, due_at)
		SELECT $1, $2, id, $4, $5, $6, $7, $8, $9
		FROM human_ai_teams
		WHERE id = $3 AND tenant_id = $2 AND org_id = $1 AND deleted_at IS NULL
		RETURNING id::text, team_id::text, title, status, priority, input_context, proposed_output, requires_approval,
			COALESCE(due_at::text, ''), COALESCE(started_at::text, ''), COALESCE(completed_at::text, ''), created_at::text, updated_at::text
	`, orgUUID, tenantUUID, pgtype.UUID{Bytes: teamID, Valid: true}, userUUID, req.Title, req.Priority, req.InputContext, req.RequiresApproval, dueAt).Scan(
		&task.ID, &task.TeamID, &task.Title, &task.Status, &task.Priority,
		&task.InputContext, &task.ProposedOutput, &task.RequiresApproval,
		&task.DueAt, &task.StartedAt, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "team not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "failed to commit task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func listHumanAITeams(ctx context.Context, tx pgx.Tx, tenantID pgtype.UUID, orgID pgtype.UUID) ([]humanAITeamResponse, error) {
	rows, err := tx.Query(ctx, `
		SELECT id::text, name, COALESCE(description, ''), status, cadence_minutes, approval_policy, created_at::text, updated_at::text
		FROM human_ai_teams
		WHERE tenant_id = $1 AND org_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teams := []humanAITeamResponse{}
	for rows.Next() {
		var team humanAITeamResponse
		if err := rows.Scan(&team.ID, &team.Name, &team.Description, &team.Status, &team.CadenceMinutes, &team.ApprovalPolicy, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func listAITasks(ctx context.Context, tx pgx.Tx, tenantID pgtype.UUID, orgID pgtype.UUID) ([]aiTaskResponse, error) {
	rows, err := tx.Query(ctx, `
		SELECT id::text, team_id::text, title, status, priority, input_context, proposed_output, requires_approval,
			COALESCE(due_at::text, ''), COALESCE(started_at::text, ''), COALESCE(completed_at::text, ''), created_at::text, updated_at::text
		FROM ai_tasks
		WHERE tenant_id = $1 AND org_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 100
	`, tenantID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []aiTaskResponse{}
	for rows.Next() {
		var task aiTaskResponse
		if err := rows.Scan(
			&task.ID, &task.TeamID, &task.Title, &task.Status, &task.Priority,
			&task.InputContext, &task.ProposedOutput, &task.RequiresApproval,
			&task.DueAt, &task.StartedAt, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func setWorkspaceRLS(ctx context.Context, tx pgx.Tx, tenantID pgtype.UUID, orgID pgtype.UUID) error {
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", uuid.UUID(tenantID.Bytes).String()); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, "SELECT set_config('app.current_org', $1, true)", uuid.UUID(orgID.Bytes).String())
	return err
}

func optionalUserUUID(r *http.Request) pgtype.UUID {
	userID, err := authcontext.GetUserID(r.Context())
	if err != nil {
		return pgtype.UUID{}
	}
	parsed, err := uuid.Parse(userID)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}
}

func nullableString(value string) pgtype.Text {
	if strings.TrimSpace(value) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
