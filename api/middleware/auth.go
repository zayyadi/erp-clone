package middleware

import (
	"context"
	"net/http"
	"strings"
	// "erp-system/pkg/errors" // For custom error types
	// "erp-system/pkg/logger" // For logging
	// "erp-system/configs"    // For JWT secret or other auth configs
)

// UserContextKey is a custom type for context key to avoid collisions.
type UserContextKey string

const (
	// ContextUserKey is the key used to store user information in the request context.
	ContextUserKey UserContextKey = "user"
)

// Authenticate is a middleware that checks for a valid token (e.g., JWT).
// This is a placeholder and needs to be implemented with actual token validation logic.
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// logger.InfoLogger.Println("Auth middleware: processing request for", r.URL.Path)

		// 1. Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// logger.WarnLogger.Println("Auth middleware: Authorization header missing")
			// http.Error(w, errors.NewUnauthorizedError("Authorization header required").Error(), http.StatusUnauthorized)
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			// logger.WarnLogger.Println("Auth middleware: Invalid Authorization header format")
			// http.Error(w, errors.NewUnauthorizedError("Invalid Authorization header format").Error(), http.StatusUnauthorized)
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// 2. Validate the token (e.g., parse JWT, check signature, expiry)
		// This is where you would use your JWT library and secret key from config.
		// For example, if using "github.com/golang-jwt/jwt/v5":
		/*
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				// cfg := configs.GetConfig() // Assuming JWTSecret is in your config
				// return []byte(cfg.JWTSecretKey), nil
				return []byte("your_jwt_secret_placeholder"), nil // Replace with actual secret
			})

			if err != nil {
				logger.ErrorLogger.Printf("Auth middleware: Invalid token: %v", err)
				http.Error(w, errors.NewUnauthorizedError("Invalid token").Error(), http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// 3. If valid, extract user information and add it to the request context.
				// You might want to store user ID, roles, etc.
				// For example, getting a "sub" (subject, usually user ID) claim:
				userID, ok := claims["sub"].(string)
				if !ok {
					logger.ErrorLogger.Println("Auth middleware: 'sub' claim missing or not a string in token")
					http.Error(w, errors.NewInternalServerError("Error processing token claims", nil).Error(), http.StatusInternalServerError)
					return
				}

				// Create a user object or struct to store in context
				// For simplicity, storing userID directly. You might have a User struct.
				ctx := context.WithValue(r.Context(), ContextUserKey, userID)
				logger.InfoLogger.Printf("Auth middleware: User %s authenticated", userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				logger.ErrorLogger.Println("Auth middleware: Token claims invalid or token is not valid")
				http.Error(w, errors.NewUnauthorizedError("Invalid token claims").Error(), http.StatusUnauthorized)
				return
			}
		*/

		// Placeholder: For now, let's assume the token is valid if it's "test_token"
		if tokenString == "test_token" {
			// Store some dummy user info in context
			ctx := context.WithValue(r.Context(), ContextUserKey, "test_user_id")
			// logger.InfoLogger.Println("Auth middleware: Test user authenticated")
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// logger.WarnLogger.Println("Auth middleware: Invalid token provided")
			// http.Error(w, errors.NewUnauthorizedError("Invalid token").Error(), http.StatusUnauthorized)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	})
}

// GetUserFromContext retrieves user information from the request context.
// The type of the returned value depends on what was stored (e.g., user ID, User struct).
func GetUserFromContext(ctx context.Context) (interface{}, bool) {
	user, ok := ctx.Value(ContextUserKey).(interface{}) // Or specific type like string, models.User
	return user, ok
}

// Authorize is a placeholder for role-based access control (RBAC) or permission checks.
// It would typically take required roles/permissions as arguments.
func Authorize(requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// user, ok := GetUserFromContext(r.Context())
			// if !ok {
			// 	logger.WarnLogger.Println("Authorize middleware: User not found in context, authentication might have failed or is missing")
			// 	http.Error(w, errors.NewUnauthorizedError("User not authenticated").Error(), http.StatusUnauthorized)
			// 	return
			// }

			// Perform permission check based on the user and requiredPermission
			// This is highly dependent on how your user model and permissions are structured.
			// Example:
			/*
				userRoles, err := fetchUserRoles(user.ID) // Fetch roles for the user
				if err != nil {
					http.Error(w, errors.NewInternalServerError("Could not fetch user roles", err).Error(), http.StatusInternalServerError)
					return
				}
				if !hasPermission(userRoles, requiredPermission) {
					http.Error(w, errors.NewForbiddenError("Insufficient permissions").Error(), http.StatusForbidden)
					return
				}
			*/

			// Placeholder: For now, allow all authenticated users to pass authorization.
			// logger.InfoLogger.Printf("Authorize middleware: User has required permission (%s) for %s", requiredPermission, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}
