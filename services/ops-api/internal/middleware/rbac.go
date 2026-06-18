package middleware

import (
	"net/http"
	"strings"

	"github.com/synq/pkg/authcontext"
)

// RequireRole checks if the authenticated user has one of the allowed roles.
// It assumes AuthMiddleware has already successfully run and populated the context.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userRole, err := authcontext.GetRole(r.Context())
			if err != nil || userRole == "" {
				http.Error(w, "Forbidden: No role assigned", http.StatusForbidden)
				return
			}

			roleAllowed := false
			for _, allowedRole := range allowedRoles {
				if strings.EqualFold(userRole, allowedRole) {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
