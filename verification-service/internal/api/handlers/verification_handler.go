package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"verification-service/internal/app/services"
	"verification-service/pkg/errors"
)

// VerificationHandler handles HTTP requests for verifications
type VerificationHandler struct {
	verificationService *services.VerificationService
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(verificationService *services.VerificationService) *VerificationHandler {
	return &VerificationHandler{
		verificationService: verificationService,
	}
}

// InitiateVerification handles the request to initiate a new verification
func (h *VerificationHandler) InitiateVerification(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Initiate verification
	verification, err := h.verificationService.InitiateVerification(r.Context(), requestBody)
	if err != nil {
		if _, ok := err.(*errors.ValidationError); ok {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to initiate verification")
		return
	}

	// Set Location header
	w.Header().Set("Location", "/api/v1/verification/"+verification.VerificationID)

	// Respond with 202 Accepted
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"verificationId": verification.VerificationID,
		"verificationAt": verification.VerificationAt,
		"status":         verification.Status,
		"message":        "Verification has been successfully initiated.",
	})
}

// GetVerification handles the request to get a verification
func (h *VerificationHandler) GetVerification(w http.ResponseWriter, r *http.Request) {
	// Get verification ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get verification
	verification, err := h.verificationService.GetVerification(r.Context(), id)
	if err != nil {
		if err.Error() == "verification not found: "+id {
			respondWithError(w, http.StatusNotFound, "Verification not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get verification")
		return
	}

	// If verification is still processing, return 202 Accepted
	if verification.Status == "PROCESSING" || 
	   verification.Status == "INITIALIZED" || 
	   verification.Status == "IMAGES_FETCHED" || 
	   verification.Status == "SYSTEM_PROMPT_READY" || 
	   verification.Status == "TURN1_PROMPT_READY" || 
	   verification.Status == "TURN1_PROCESSING" || 
	   verification.Status == "TURN1_COMPLETED" || 
	   verification.Status == "TURN2_PROMPT_READY" || 
	   verification.Status == "TURN2_PROCESSING" {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"verificationId": verification.VerificationID,
			"status":         verification.Status,
			"message":        "Your verification is still processing. Please check back shortly.",
		})
		return
	}

	// Respond with verification
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(verification)
}

// ListVerifications handles the request to list verifications
func (h *VerificationHandler) ListVerifications(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters := make(map[string]interface{})
	
	if vendingMachineID := r.URL.Query().Get("vendingMachineId"); vendingMachineID != "" {
		filters["vendingMachineId"] = vendingMachineID
	}
	
	if status := r.URL.Query().Get("verificationStatus"); status != "" {
		filters["verificationStatus"] = status
	}
	
	if fromDate := r.URL.Query().Get("fromDate"); fromDate != "" {
		filters["fromDate"] = fromDate
	}
	
	if toDate := r.URL.Query().Get("toDate"); toDate != "" {
		filters["toDate"] = toDate
	}
	
	// Parse pagination parameters
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	
	// List verifications
	verifications, total, err := h.verificationService.ListVerifications(r.Context(), filters, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list verifications")
		return
	}
	
	// Calculate next offset
	nextOffset := offset + len(verifications)
	if nextOffset >= total {
		nextOffset = 0
	}
	
	// Respond with verifications
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": verifications,
		"pagination": map[string]interface{}{
			"total":      total,
			"limit":      limit,
			"offset":     offset,
			"nextOffset": nextOffset,
		},
	})
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}