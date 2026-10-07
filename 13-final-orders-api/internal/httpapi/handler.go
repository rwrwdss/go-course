// Package httpapi — HTTP-слой: разбор запросов, вызов сервиса, маппинг ошибок в статусы.
package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"gocourse/13-final-orders-api/internal/domain"
	"gocourse/13-final-orders-api/internal/service"
)

// Handler — HTTP API заказов.
type Handler struct {
	svc *service.OrderService
	mux *http.ServeMux
}

func NewHandler(svc *service.OrderService) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /health", h.health)
	h.mux.HandleFunc("GET /orders", h.listOrders)
	h.mux.HandleFunc("POST /orders", h.createOrder)
	h.mux.HandleFunc("GET /orders/{id}", h.getOrder)
	h.mux.HandleFunc("POST /orders/{id}/status", h.changeStatus)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

// writeServiceError переводит ошибку сервиса в HTTP-ответ {"error": "..."}:
//
//	domain.ErrValidation        → 422, текст ошибки
//	domain.ErrNotFound          → 404, текст ошибки
//	domain.ErrInvalidTransition → 409, текст ошибки
//	всё остальное               → 500 "internal error" (детали только в лог, клиенту — нельзя!)
//
// 🐞 ЗАДАНИЕ 4: сейчас всё превращается в 500, а внутренний текст ошибки утекает клиенту.
func writeServiceError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.List())
}

type createOrderRequest struct {
	Customer string              `json:"customer"`
	Items    []service.ItemInput `json:"items"`
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	order, err := h.svc.CreateOrder(req.Customer, req.Items)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Location", "/orders/"+strconv.Itoa(order.ID))
	writeJSON(w, http.StatusCreated, order)
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	order, err := h.svc.Get(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

type changeStatusRequest struct {
	Status domain.Status `json:"status"`
}

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req changeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	order, err := h.svc.ChangeStatus(id, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}
