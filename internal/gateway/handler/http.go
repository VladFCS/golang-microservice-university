package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/domain"
	"github.com/vlad/microservices-grpc-kubernetes/internal/gateway/service"
)

type HTTPHandler struct {
	service *service.Service
	logger  *slog.Logger
}

func NewHTTPHandler(service *service.Service, logger *slog.Logger) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		logger:  logger,
	}
}

func (h *HTTPHandler) Register(router chi.Router) {
	router.Get("/products/{id}", h.getProduct)
}

func (h *HTTPHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimSpace(chi.URLParam(r, "id"))
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": domain.ErrProductIDRequired.Error()})
		return
	}

	product, err := h.service.GetProductDetails(r.Context(), productID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to fetch product details",
			slog.String("product_id", productID),
			slog.Any("error", err),
		)
		writeGatewayError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func writeGatewayError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProductIDRequired):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrProductNotFound), errors.Is(err, domain.ErrStockNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, context.DeadlineExceeded):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "downstream request timed out"})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch downstream data"})
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(append(body, '\n'))
}
