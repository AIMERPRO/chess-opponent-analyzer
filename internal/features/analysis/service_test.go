package analysis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/infrastructure/lichess"
)

const heroName = "hero"

// newGamesServer returns a test server that streams the given games as ndjson,
// mimicking the lichess "export games" endpoint.
func newGamesServer(t *testing.T, games []lichess.GameLichess) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		enc := json.NewEncoder(w) // Encode writes one JSON object + newline per call
		for _, g := range games {
			if err := enc.Encode(g); err != nil {
				t.Errorf("encoding game failed: %v", err)
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// heroGame builds a game from the hero's point of view.
// heroColor is "white" or "black"; winner is "white"/"black".
func heroGame(opening, heroColor, winner, status string, heroAccuracy int, createdAt int64) lichess.GameLichess {
	hero := lichess.Player{
		User:     &lichess.User{Name: heroName},
		Analysis: &lichess.PlayerAnalysis{Accuracy: heroAccuracy},
	}
	opp := lichess.Player{User: &lichess.User{Name: "villain"}}

	players := lichess.Players{}
	if heroColor == "white" {
		players.White, players.Black = hero, opp
	} else {
		players.White, players.Black = opp, hero
	}

	w := winner
	return lichess.GameLichess{
		Speed:     "blitz",
		Status:    status,
		Winner:    &w,
		Opening:   &lichess.Opening{Name: opening},
		Players:   players,
		CreatedAt: createdAt,
	}
}

func newTestService(srv *httptest.Server) *service {
	return &service{lichessClient: lichess.NewClient(srv.URL + "/")}
}

func approx(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-9
}

func TestService_AnalyzeGames(t *testing.T) {
	now := time.Now()
	recent := now.UnixMilli()
	old := now.AddDate(0, 0, -20).UnixMilli()

	games := []lichess.GameLichess{
		// recent
		heroGame("Italian Game", "white", "white", "mate", 90, recent),   // win
		heroGame("Italian Game", "white", "black", "resign", 70, recent), // loss by resign -> tilt
		// old
		heroGame("Sicilian Defense", "black", "black", "mate", 80, old),   // win
		heroGame("Sicilian Defense", "black", "white", "resign", 60, old), // loss by resign -> tilt
	}

	srv := newGamesServer(t, games)
	s := newTestService(srv)

	res, err := s.analyzeGames(context.Background(), heroName, "blitz")
	if err != nil {
		t.Fatalf("analyzeGames() error = %v", err)
	}

	if res.Speed != "blitz" {
		t.Errorf("Speed = %q, want blitz", res.Speed)
	}
	if !approx(res.Winrate, 50) {
		t.Errorf("Winrate = %v, want 50", res.Winrate)
	}
	// losses = 2, both by resign -> tilt = 2/2 * 100
	if !approx(res.TiltFactor, 100) {
		t.Errorf("TiltFactor = %v, want 100", res.TiltFactor)
	}
	// accuracy over all 4 games = (90+70+80+60)/4
	if !approx(res.AvgAccuracy, 75) {
		t.Errorf("AvgAccuracy = %v, want 75", res.AvgAccuracy)
	}
	if res.MostPopularDebutWhite != "Italian Game" {
		t.Errorf("MostPopularDebutWhite = %q, want Italian Game", res.MostPopularDebutWhite)
	}
	if res.MostPopularDebutBlack != "Sicilian Defense" {
		t.Errorf("MostPopularDebutBlack = %q, want Sicilian Defense", res.MostPopularDebutBlack)
	}
	if res.MostWinrateDebutWhite != "Italian Game" {
		t.Errorf("MostWinrateDebutWhite = %q, want Italian Game", res.MostWinrateDebutWhite)
	}
	if res.MostWinrateDebutBlack != "Sicilian Defense" {
		t.Errorf("MostWinrateDebutBlack = %q, want Sicilian Defense", res.MostWinrateDebutBlack)
	}

	// only the two recent games count for the 10-day window: 1 win / 2 games
	if !approx(res.WinrateLast10Days, 50) {
		t.Errorf("WinrateLast10Days = %v, want 50", res.WinrateLast10Days)
	}
	// recent accuracies = (90+70)/2
	if !approx(res.AvgAccuracyLast10Days, 80) {
		t.Errorf("AvgAccuracyLast10Days = %v, want 80", res.AvgAccuracyLast10Days)
	}
}

func TestService_AnalyzeGames_LichessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	s := newTestService(srv)

	_, err := s.analyzeGames(context.Background(), heroName, "blitz")
	if err == nil {
		t.Fatal("expected error when lichess returns non-200, got nil")
	}
}

func TestService_checkIfUserBlackOrWhite(t *testing.T) {
	s := &service{}

	blackGame := lichess.GameLichess{
		Players: lichess.Players{
			White: lichess.Player{User: &lichess.User{Name: "villain"}},
			Black: lichess.Player{User: &lichess.User{Name: heroName}},
		},
	}
	got, err := s.checkIfUserBlackOrWhite(blackGame, heroName)
	if err != nil {
		t.Fatalf("checkIfUserBlackOrWhite() error = %v", err)
	}

	if got != "Black" {
		t.Errorf("checkIfUserBlackOrWhite() = %q, want Black", got)
	}

	whiteGame := lichess.GameLichess{
		Players: lichess.Players{
			White: lichess.Player{User: &lichess.User{Name: heroName}},
			Black: lichess.Player{User: &lichess.User{Name: "villain"}},
		},
	}

	got, err = s.checkIfUserBlackOrWhite(whiteGame, heroName)
	if err != nil {
		t.Fatalf("checkIfUserBlackOrWhite() error = %v", err)
	}

	if got != "White" {
		t.Errorf("checkIfUserBlackOrWhite() = %q, want White", got)
	}

	nilUserGame := lichess.GameLichess{}
	_, err = s.checkIfUserBlackOrWhite(nilUserGame, heroName)
	if err == nil {
		t.Errorf("checkIfUserBlackOrWhite() with nil user = %q, want Error", got)
	}
}

func TestService_mostPopularDebut(t *testing.T) {
	s := &service{}

	counter := map[string]int{"Italian Game": 5, "French Defense": 2}
	if got := s.mostPopularDebut(counter); got != "Italian Game" {
		t.Errorf("mostPopularDebut() = %q, want Italian Game", got)
	}

	if got := s.mostPopularDebut(map[string]int{}); got != "" {
		t.Errorf("mostPopularDebut(empty) = %q, want empty string", got)
	}
}

func TestService_mostWinrateDebut(t *testing.T) {
	s := &service{}

	counter := map[string]float64{"Italian Game": 80.0, "French Defense": 40.0}
	if got := s.mostWinrateDebut(counter); got != "Italian Game" {
		t.Errorf("mostWinrateDebut() = %q, want Italian Game", got)
	}

	if got := s.mostWinrateDebut(map[string]float64{}); got != "" {
		t.Errorf("mostWinrateDebut(empty) = %q, want empty string", got)
	}
}
