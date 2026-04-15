package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/Harsh-Bansal-13/taskflow-harsh/backend/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string, fields map[string]string) {
	resp := models.ErrorResponse{Error: msg, Fields: fields}
	writeJSON(w, status, resp)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}

func paginatedResponse(data interface{}, page, limit, total int) models.PaginatedResponse {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return models.PaginatedResponse{
		Data:       data,
		Page:       page,
		Limit:      limit,
		TotalCount: total,
		TotalPages: totalPages,
	}
}
