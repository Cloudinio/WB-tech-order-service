package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"
	"github.com/go-chi/chi/v5"
	"github.com/Cloudinio/wb-tech-order-service/internal/usecase"
	repopg "github.com/Cloudinio/wb-tech-order-service/internal/repository/postgres"
)

type Handler struct {
	repo usecase.OrderRepository
}

func NewHandler(repo usecase.OrderRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) GetOrderByUID(w nethttp.ResponseWriter, r *nethttp.Request) {
	orderUID := chi.URLParam(r, "order_uid")
	if orderUID == "" {
		writeJSON(w, nethttp.StatusBadRequest, map[string]string{
			"error": "order_uid is required",
		})
		return
	}

	order, err := h.repo.GetByUID(r.Context(), orderUID)
	if err != nil {
		if errors.Is(err, repopg.ErrOrderNotFound) {
			writeJSON(w, nethttp.StatusNotFound, map[string]string{
				"error": "order not found",
			})
			return
		}

		writeJSON(w, nethttp.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	resp := NewOrderResponse(order)
	writeJSON(w, nethttp.StatusOK, resp)
}

func writeJSON(w nethttp.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
	}
}