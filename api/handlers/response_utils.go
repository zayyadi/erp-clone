package handlers

import (
	"encoding/json"
	"erp-system/pkg/errors" // Assuming this is your custom errors package
	"erp-system/pkg/logger"
	"net/http"
)

// --- Generic API Response Structures ---

// SuccessResponse wraps a successful API response.
type SuccessResponse struct {
	Status  string      `json:"status"` // e.g., "success"
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse wraps an error API response.
type ErrorResponse struct {
	Status  string      `json:"status"` // e.g., "error"
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"` // e.g., validation errors map[string]string or error code
}

// PaginatedResponse is a generic structure for paginated list responses.
// This was previously defined within inventory_handlers.go, moving it here.
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total int64       `json:"total"`
}


// --- Shared Handler Utility Functions ---

// respondWithError sends a JSON error response.
func respondWithError(w http.ResponseWriter, err error) {
	logger.ErrorLogger.Printf("API Error: %v", err) // Log the original error
	var statusCode int
	var apiError ErrorResponse // Use the shared ErrorResponse

	switch e := err.(type) {
	case *errors.NotFoundError:
		statusCode = http.StatusNotFound
		apiError = ErrorResponse{Status: "error", Message: e.Error()}
	case *errors.ValidationError:
		statusCode = http.StatusBadRequest
		// If e.Field is relevant, include it in details or message
		details := e.Field
		if details == "" { // If field is not set, just use message
			apiError = ErrorResponse{Status: "error", Message: e.Message}
		} else {
			apiError = ErrorResponse{Status: "error", Message: e.Error(), Details: details}
		}
	case *errors.ConflictError:
		statusCode = http.StatusConflict
		apiError = ErrorResponse{Status: "error", Message: e.Error()}
	case *errors.UnauthorizedError:
		statusCode = http.StatusUnauthorized
		apiError = ErrorResponse{Status: "error", Message: e.Error()}
	case *errors.ForbiddenError:
		statusCode = http.StatusForbidden
		apiError = ErrorResponse{Status: "error", Message: e.Error()}
	case *errors.InternalServerError: // Handle our custom internal server error
		statusCode = http.StatusInternalServerError
		apiError = ErrorResponse{Status: "error", Message: "An internal server error occurred."}
		// Optionally include e.Error() in Details for debugging if not sensitive
		// apiError.Details = e.Error()
	default: // Handles standard Go errors or other unclassified errors
		statusCode = http.StatusInternalServerError
		apiError = ErrorResponse{Status: "error", Message: "An unexpected error occurred."}
		// Log the original error string for server logs if it's a generic error
		logger.ErrorLogger.Printf("Defaulting to 500 for unhandled error type: %T, value: %v", err, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(apiError); encodeErr != nil {
		logger.ErrorLogger.Printf("Failed to encode error JSON response: %v", encodeErr)
		// Fallback if encoding the error itself fails
		http.Error(w, `{"status":"error","message":"Failed to encode error response"}`, http.StatusInternalServerError)
	}
}

// respondWithJSON sends a JSON success response.
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response := SuccessResponse{ // Use the shared SuccessResponse
		Status: "success",
		Data:   payload,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.ErrorLogger.Printf("Failed to encode JSON success response: %v", err)
		// Fallback error response if JSON encoding fails for the success response
		// This situation should be rare.
		respondWithError(w, errors.NewInternalServerError("failed to encode success response", err))
	}
}
