package auth

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/config"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/middleware"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/response"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	log     *zap.Logger
}

func NewHandler(service Service, log *zap.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, cfg *config.Config) {
	// public routes - w/o auth middleware
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/refresh", h.RefreshToken)
	mux.HandleFunc("POST /auth/logout", h.Logout)

	// secure routes - with auth middleware
	mux.Handle("POST /auth/logout-all", middleware.AuthMiddleware(cfg, http.HandlerFunc(h.LogoutFromAllDevices)))
	mux.Handle("GET /users/{id}", middleware.AuthMiddleware(cfg, http.HandlerFunc(h.GetUser)))
	mux.Handle("PATCH /users/{id}", middleware.AuthMiddleware(cfg, http.HandlerFunc(h.UpdateUser)))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("decoding request failed", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokenPair, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.log.Error("failed to login user", zap.Error(err))
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, tokenPair)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("decoding request failed", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokenPair, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.log.Error("failed to register user", zap.Error(err))
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, tokenPair)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req TokenRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("decoding request failed", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokenPair, err := h.service.RefreshToken(r.Context(), req)
	if err != nil {
		h.log.Error("failed to refresh token", zap.Error(err))
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, tokenPair)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req TokenRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("decoding request failed", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.service.Logout(r.Context(), req.RefreshToken)

	if err != nil {
		h.log.Error("failed to logout", zap.Error(err))
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}

func (h *Handler) LogoutFromAllDevices(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	err := h.service.LogoutFromAllDevices(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to logout from all devices", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.log.Error("invalid user ID", zap.String("id", userIDStr))
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to get user", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.log.Error("invalid user ID", zap.String("id", userIDStr))
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req UpdateUserDTO
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("decoding request failed", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		h.log.Error("failed to update user", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	response.JSON(w, http.StatusOK, user)
}
