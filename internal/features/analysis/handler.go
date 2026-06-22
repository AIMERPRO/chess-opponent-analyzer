package analysis

import (
	"net/http"

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
	// secure routes - with auth middleware
	mux.Handle("GET /analyze/{username}", middleware.AuthMiddleware(cfg, http.HandlerFunc(h.AnalyzeUser)))
}

func (h *Handler) AnalyzeUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	speed := r.URL.Query().Get("speed")
	if speed == "" {
		speed = "blitz"
	}
	
	var userAnalysisDTO AnalyzeDTO
	analysis, err := h.service.Analyze(r.Context(), username, speed)
	if err != nil {
		h.log.Error("failed to analyze user", zap.String("username", username), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userAnalysisDTO = AnalyzeDTO{
		Speed:                 speed,
		Winrate:               analysis.Winrate,
		WinrateLast10Days:     analysis.WinrateLast10Days,
		MostPopularDebutBlack: analysis.MostPopularDebutBlack,
		MostPopularDebutWhite: analysis.MostPopularDebutWhite,
		MostWinrateDebutBlack: analysis.MostWinrateDebutBlack,
		MostWinrateDebutWhite: analysis.MostWinrateDebutWhite,
		AvgAccuracy:           analysis.AvgAccuracy,
		AvgAccuracyLast10Days: analysis.AvgAccuracyLast10Days,
		TiltFactor:            analysis.TiltFactor,
	}

	response.JSON(w, http.StatusOK, userAnalysisDTO)

}
