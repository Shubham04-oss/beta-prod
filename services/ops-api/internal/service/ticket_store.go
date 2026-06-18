package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/synq/pkg/db"
)

type TicketClaims struct {
	TenantID string
	OrgID    string
	UserID   string
	Role     string
}

type WSTicketStore struct {
	queries *db.Queries
}

func NewWSTicketStore(q *db.Queries) *WSTicketStore {
	return &WSTicketStore{queries: q}
}

func (s *WSTicketStore) CreateTicket(ctx context.Context, claims TicketClaims) (string, error) {
	var tenantID pgtype.UUID
	if claims.TenantID != "" {
		_ = tenantID.Scan(claims.TenantID)
	}

	var orgID pgtype.UUID
	if claims.OrgID != "" {
		_ = orgID.Scan(claims.OrgID)
	}

	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(60 * time.Second),
		Valid: true,
	}

	id, err := s.queries.CreateWSTicket(ctx, db.CreateWSTicketParams{
		TenantID:  tenantID,
		OrgID:     orgID,
		UserID:    claims.UserID,
		Role:      claims.Role,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return UUIDToString(id), nil
}

func (s *WSTicketStore) ConsumeTicket(ctx context.Context, ticketIDStr string) (*TicketClaims, error) {
	var id pgtype.UUID
	err := id.Scan(ticketIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket id format")
	}

	row, err := s.queries.ConsumeWSTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	return &TicketClaims{
		TenantID: UUIDToString(row.TenantID),
		OrgID:    UUIDToString(row.OrgID),
		UserID:   row.UserID,
		Role:     row.Role,
	}, nil
}

func UUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
