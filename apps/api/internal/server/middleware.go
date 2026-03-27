package server

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mantrixflow/go-api/pkg/response"
)

const ctxUserID = "userID"
const ctxEmail = "userEmail"

// AuthJWT verifies HS256 Supabase JWT and sets user id + email on context.
func (s *State) AuthJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "No authentication token provided")
		}
		raw := strings.TrimSpace(h[7:])
		tok, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(s.Cfg.SupabaseJWTSecret), nil
		})
		if err != nil || !tok.Valid {
			return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
		}
		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims")
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token missing sub")
		}
		email, _ := claims["email"].(string)
		c.Locals(ctxUserID, sub)
		c.Locals(ctxEmail, email)
		return c.Next()
	}
}

func UserID(c *fiber.Ctx) string {
	v := c.Locals(ctxUserID)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// RequireOrgMember ensures :organizationId path param matches an active membership for the user.
func (s *State) RequireOrgMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := c.Params("organizationId")
		uid := UserID(c)
		if orgID == "" || uid == "" {
			return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "organizationId required")
		}
		var n int64
		err := s.DB.Raw(`
			SELECT COUNT(*) FROM organization_members
			WHERE organization_id = ?::uuid AND user_id = ?::uuid
			AND status IN ('active','accepted')`,
			orgID, uid,
		).Scan(&n).Error
		if err != nil || n == 0 {
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "Not a member of this organization")
		}
		return c.Next()
	}
}

// InternalToken matches Nest internal/callback auth.
func (s *State) InternalToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		got := c.Get("X-Callback-Token")
		if got == "" {
			got = c.Get("X-Internal-Token")
		}
		exp := s.Cfg.CallbackToken
		if exp == "" {
			exp = s.Cfg.InternalToken
		}
		if exp == "" || got != exp {
			return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid callback token")
		}
		return c.Next()
	}
}
