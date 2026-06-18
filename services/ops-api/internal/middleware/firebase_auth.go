package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/synq/pkg/authcontext"
)

// ContextKey is a custom type to prevent context collisions
type ContextKey string

const (
	IsAgentKey ContextKey = "is_agent"
)

// AuthMiddleware validates Identity Platform JWTs (Humans) and SPIFFE Tokens (Agents).
func AuthMiddleware(authClient *auth.Client, dbpool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: Missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			ctx := r.Context()

			// TODO: Implement proper SPIFFE/DPoP verification for agents.
			// The fake SPIFFE bypass has been removed for security reasons.

			// Fallback to Human Identity Verification (Firebase)
			token, err := authClient.VerifyIDToken(r.Context(), tokenString)
			
			var tokenUID string
			var tokenClaims map[string]interface{}
			var tenantID string

			if err != nil {
				// In local development, the Next.js frontend (localhost:9099) and the Go backend (192.168.1.6:9099)
				// often have mismatched JWT Issuers, causing verification to fail. We will manually parse the JWT payload.
				if os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") != "" && os.Getenv("ENV") != "production" {
					parts := strings.Split(tokenString, ".")
					if len(parts) == 3 {
						payload, err := base64.RawURLEncoding.DecodeString(parts[1])
						if err == nil {
							var claims map[string]interface{}
							if json.Unmarshal(payload, &claims) == nil {
								tokenUID, _ = claims["user_id"].(string)
								tokenClaims = claims
							}
						}
					}
					
					if tokenUID == "" {
						http.Error(w, "Unauthorized: Invalid token and manual decode failed", http.StatusUnauthorized)
						return
					}
				} else {
					http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
					return
				}
			} else {
				tokenUID = token.UID
				tokenClaims = token.Claims
				if token.Firebase.Tenant != "" {
					tenantID = token.Firebase.Tenant
				}
			}

			// 1. Extract tenantid
			if tenantID == "" {
				if tid, ok := tokenClaims["tenant_id"].(string); ok {
					tenantID = tid
				}
			}

			// Local Dev Bypass
			if tenantID == "" && os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") != "" && os.Getenv("ENV") != "production" {
				var dbOrgID, dbTenantID string
				err := dbpool.QueryRow(ctx, "SELECT org_id, id FROM tenants LIMIT 1").Scan(&dbOrgID, &dbTenantID)
				if err == nil {
					// Intelligently sync the emulator user to the local Postgres database so their email is recognized
					userEmail := "admin@aurea.dev"
					if emailRaw, ok := tokenClaims["email"]; ok {
						if parsed, ok := emailRaw.(string); ok && parsed != "" {
							userEmail = parsed
						}
					}

					var dbUserID string
					err = dbpool.QueryRow(ctx, "SELECT id::text FROM users WHERE email = $1 LIMIT 1", userEmail).Scan(&dbUserID)
					if err != nil {
						// User doesn't exist, create them locally to fix the "mess"
						dbpool.QueryRow(ctx, `INSERT INTO users (org_id, tenant_id, email, role) VALUES ($1, $2, $3, 'admin') RETURNING id::text`, dbOrgID, dbTenantID, userEmail).Scan(&dbUserID)
					}

					ctx = authcontext.WithTenantID(ctx, dbTenantID)
					ctx = authcontext.WithOrgID(ctx, dbOrgID)
					ctx = authcontext.WithRole(ctx, "admin")
					if dbUserID != "" {
						ctx = authcontext.WithUserID(ctx, dbUserID)
					} else {
						ctx = authcontext.WithUserID(ctx, tokenUID)
					}
					ctx = context.WithValue(ctx, IsAgentKey, false)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			if tenantID == "" {
				http.Error(w, "Forbidden: Missing tenantid claim", http.StatusForbidden)
				return
			}
			ctx = authcontext.WithTenantID(ctx, tenantID)

			// 2. Extract org_id
			orgID, ok := token.Claims["org_id"].(string)
			if !ok || orgID == "" {
				http.Error(w, "Forbidden: Missing orgid claim", http.StatusForbidden)
				return
			}
			ctx = authcontext.WithOrgID(ctx, orgID)

			// 3. Extract role
			role, ok := token.Claims["role"].(string)
			if !ok || role == "" {
				http.Error(w, "Forbidden: Missing role claim", http.StatusForbidden)
				return
			}
			ctx = authcontext.WithRole(ctx, role)

			// 4. Extract db_uid
			if dbUID, ok := token.Claims["db_uid"].(string); ok {
				ctx = authcontext.WithUserID(ctx, dbUID)
			}
			ctx = context.WithValue(ctx, IsAgentKey, false)

			// Call the next handler with the enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
