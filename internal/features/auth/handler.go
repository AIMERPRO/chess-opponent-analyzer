package auth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/apperrors"
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

// Login godoc
// @Summary      Login
// @Description  Authenticate by username/password and receive an access/refresh token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequestDTO  true  "Login credentials"
// @Success      200      {object}  TokenResponseDTO
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[LoginRequestDTO](r, w, h.log)
	if !ok {
		return
	}

	tokenPair, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.log.Error("failed to login user", zap.Error(err))
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	response.JSON(w, http.StatusOK, tokenPair)
}

// Register godoc
// @Summary      Register
// @Description  Create a new account and receive an access/refresh token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequestDTO  true  "Registration data"
// @Success      200      {object}  TokenResponseDTO
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[RegisterRequestDTO](r, w, h.log)
	if !ok {
		return
	}

	tokenPair, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.log.Error("failed to register user", zap.Error(err))
		if errors.Is(err, apperrors.ErrConflict) {
			response.Error(w, http.StatusConflict, "username already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	response.JSON(w, http.StatusOK, tokenPair)
}

// RefreshToken godoc
// @Summary      Refresh tokens
// @Description  Exchange a valid refresh token for a new access/refresh token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      TokenRequestDTO  true  "Refresh token"
// @Success      200      {object}  TokenResponseDTO
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /auth/refresh [post]
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[TokenRequestDTO](r, w, h.log)
	if !ok {
		return
	}

	tokenPair, err := h.service.RefreshToken(r.Context(), req)
	if err != nil {
		h.log.Error("failed to refresh token", zap.Error(err))
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "token not found")
			return
		}
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, tokenPair)
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidate a single refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      TokenRequestDTO  true  "Refresh token"
// @Success      204      "No Content"
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[TokenRequestDTO](r, w, h.log)
	if !ok {
		return
	}

	err := h.service.Logout(r.Context(), req.RefreshToken)

	if err != nil {
		h.log.Error("failed to logout", zap.Error(err))
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}

// LogoutFromAllDevices godoc
// @Summary      Logout from all devices
// @Description  Invalidate every refresh token of the authenticated user
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/logout-all [post]
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

// GetUser godoc
// @Summary      Get user by ID
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponseDTO
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{id} [get]
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
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	response.JSON(w, http.StatusOK, user)
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Update username and/or lichess username of the user
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int            true  "User ID"
// @Param        request  body      UpdateUserDTO  true  "Fields to update"
// @Success      200      {object}  UserResponseDTO
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /users/{id} [patch]
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.log.Error("invalid user ID", zap.String("id", userIDStr))
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	req, ok := decodeAndValidate[UpdateUserDTO](r, w, h.log)
	if !ok {
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		h.log.Error("failed to update user", zap.Error(err))
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	response.JSON(w, http.StatusOK, user)
}
