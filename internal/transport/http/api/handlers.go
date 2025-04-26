package api

import (
	"AuthService/internal/models"
	"AuthService/internal/errdefs"
	"AuthService/internal/interfaces"
	"AuthService/internal/logger"

	"context"
	"encoding/json"
	"strings"
	"net/http"
	"fmt"

    "go.uber.org/zap"
	"github.com/google/uuid"
)

type Handler struct {
	authService interfaces.IAuthService
	ctx context.Context
}

func NewHandler(ctx context.Context, authService interfaces.IAuthService) *Handler {
	return &Handler{
		ctx: ctx,
		authService: authService,
	}
}

func (h *Handler) GenerateRequestID(ctx context.Context) context.Context {
    return context.WithValue(ctx, logger.RequestID, uuid.New().String())
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNotSupported
	}
	
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", http.ErrNotSupported
	}
	
	return authHeader[7:], nil
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := h.GenerateRequestID(h.ctx)
	logger.GetLoggerFromCtx(ctx).Info(ctx, "HandleLogin invoked")

	userID := r.URL.Query().Get("userID")
	if userID == "" {
        logger.GetLoggerFromCtx(ctx).Error(ctx, "Missing userID parameter for login")
        http.Error(w, "userID parameter is required", http.StatusBadRequest)
        return
    }

    ip := getClientIP(r)
	
	access, refresh, err := h.authService.Login(r.Context(), userID, ip)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, 
			fmt.Sprintf("error generate toekns: %w", err),
		)
		http.Error(w, "Error generate toekns", http.StatusInternalServerError)
		return
	}
	
	response := models.LoginResponse{
		Access: access,
		Refresh: refresh,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
//
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := h.GenerateRequestID(h.ctx)
	logger.GetLoggerFromCtx(ctx).Info(ctx, "HandleRefreshToken invoked")

	var req struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, 
			fmt.Sprintf("error parse body: %w", err),
		)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clientIP := getClientIP(r)

	newAccess, newRefresh, err := h.authService.RefreshToken(
		r.Context(), 
		req.AccessToken, 
		req.RefreshToken, 
		clientIP,
	)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, 
			fmt.Sprintf("error generate toekns: %w", err),
		)
		http.Error(w, "Error generate toekns", http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Access: newAccess,
		Refresh: newRefresh,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := h.GenerateRequestID(h.ctx)
	logger.GetLoggerFromCtx(ctx).Info(ctx, "HandleRegister invoked")

	var regReq models.RegistrationRequest
	
	if err := json.NewDecoder(r.Body).Decode(&regReq); err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, 
			fmt.Sprintf("Failed to decode registration request: %w", err),
		)
		http.Error(w, "Invalid request body", http.StatusInternalServerError)
		return
	}
	
	user := &models.User{
		Email:    regReq.Email,
	}

	ip := getClientIP(r)
	
	userID, access, refresh, err := h.authService.Register(r.Context(), user, ip)
	if err != nil {
		if errdefs.Is(err, errdefs.ErrInvalidCredentials) {
			logger.GetLoggerFromCtx(ctx).Error(ctx, 
				fmt.Sprintf("Invalid registration data: %w", err),
			)
			http.Error(w, "Invalid registration data", http.StatusInternalServerError)
			return
		}
		logger.GetLoggerFromCtx(ctx).Error(ctx, 
			fmt.Sprintf("User registration failed: %w", err),
		)
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}
	

	logger.GetLoggerFromCtx(ctx).Info(ctx, "User successfully registered", zap.String("userID", userID))

	response := models.RegistrationResponse{
		UserID: userID,
		Access:   access,
		Refresh:  refresh,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getClientIP(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}