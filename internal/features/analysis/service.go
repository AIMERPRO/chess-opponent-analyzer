package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/domain"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/infrastructure/lichess"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	Analyze(ctx context.Context, username string, speed string) (domain.Analysis, error)
}

type service struct {
	lichessClient *lichess.Client
	redis         *redis.Client
}

func NewService(lichessClient *lichess.Client, redis *redis.Client) Service {
	return &service{
		lichessClient: lichessClient,
		redis:         redis,
	}
}

func (s *service) Analyze(ctx context.Context, username string, speed string) (domain.Analysis, error) {
	redisKey := username + "_" + speed
	cachedAnalysis, err := s.redis.Get(ctx, redisKey).Result()
	if err == nil {
		var analysis domain.Analysis
		if err = json.Unmarshal([]byte(cachedAnalysis), &analysis); err != nil {
		} else {
			analysis.CachedAt = time.Now()
			return analysis, nil
		}
	}

	analysis, err := s.analyzeGames(ctx, username, speed)
	if err != nil {
		return domain.Analysis{}, err
	}

	analysisJSON, err := json.Marshal(analysis)
	if err == nil {
		s.redis.Set(ctx, redisKey, analysisJSON, 2*time.Hour)
	}

	return analysis, nil
}

func (s *service) analyzeGames(ctx context.Context, username string, speed string) (domain.Analysis, error) {
	games, err := s.lichessClient.GetUserGames(ctx, username, speed, 100)
	if err != nil {
		return domain.Analysis{}, err
	}

	openingsCounterWhite := make(map[string]int)
	openingsCounterBlack := make(map[string]int)
	winsByOpeningWhite := make(map[string]int)
	winsByOpeningBlack := make(map[string]int)

	var tiltCounter int

	var winCounter int
	var winCounterLast10 int
	var gamesCountLast10 int

	accuracyList := make([]float64, 0, len(games))
	var avgAccuracy float64
	accuracyListLast10 := make([]float64, 0)
	var avgAccuracyLast10 float64

	tenDaysAgo := time.Now().AddDate(0, 0, -10).UnixMilli()

	var processedGames int
	for _, game := range games {
		userBlackOrWhite, gameErr := s.checkIfUserBlackOrWhite(game, username)
		if gameErr != nil {
			continue
		}
		processedGames++

		if userBlackOrWhite == "Black" {
			if game.Opening != nil {
				openingsCounterBlack[game.Opening.Name]++
			}
			if game.Players.Black.Analysis != nil {
				accuracyList = append(accuracyList, float64(game.Players.Black.Analysis.Accuracy))
			}
		} else {
			if game.Opening != nil {
				openingsCounterWhite[game.Opening.Name]++
			}
			if game.Players.White.Analysis != nil {
				accuracyList = append(accuracyList, float64(game.Players.White.Analysis.Accuracy))
			}
		}

		if game.Winner != nil {
			userColor := strings.ToLower(userBlackOrWhite)
			if *game.Winner == userColor {
				winCounter++
			} else if game.Status == "resign" {
				tiltCounter++
			}
		}

		if game.Winner != nil && game.Opening != nil {
			userColor := strings.ToLower(userBlackOrWhite)
			if *game.Winner == userColor {
				if userBlackOrWhite == "Black" {
					winsByOpeningBlack[game.Opening.Name]++
				} else {
					winsByOpeningWhite[game.Opening.Name]++
				}
			}
		}

		isLast10Days := game.CreatedAt >= tenDaysAgo
		if isLast10Days {
			gamesCountLast10++
			if game.Winner != nil {
				userColor := strings.ToLower(userBlackOrWhite)
				if *game.Winner == userColor {
					winCounterLast10++
				}
			}
			if userBlackOrWhite == "Black" {
				if game.Players.Black.Analysis != nil {
					accuracyListLast10 = append(accuracyListLast10, float64(game.Players.Black.Analysis.Accuracy))
				}
			} else {
				if game.Players.White.Analysis != nil {
					accuracyListLast10 = append(accuracyListLast10, float64(game.Players.White.Analysis.Accuracy))
				}
			}
		}
	}

	var winRate float64
	if processedGames > 0 {
		winRate = float64(winCounter) / float64(processedGames) * 100
	}

	var tiltFactor float64
	losses := processedGames - winCounter
	if losses > 0 {
		tiltFactor = float64(tiltCounter) / float64(losses) * 100
	}

	for _, acc := range accuracyList {
		avgAccuracy += acc
	}
	if len(accuracyList) > 0 {
		avgAccuracy /= float64(len(accuracyList))
	}

	mostPopularDebutBlack := s.mostPopularDebut(openingsCounterBlack)
	mostPopularDebutWhite := s.mostPopularDebut(openingsCounterWhite)

	winrateByOpeningWhite := make(map[string]float64)
	for opening, wins := range winsByOpeningWhite {
		total := openingsCounterWhite[opening]
		winrateByOpeningWhite[opening] = float64(wins) / float64(total) * 100
	}

	winrateByOpeningBlack := make(map[string]float64)
	for opening, wins := range winsByOpeningBlack {
		total := openingsCounterBlack[opening]
		winrateByOpeningBlack[opening] = float64(wins) / float64(total) * 100
	}

	mostWinrateDebutWhite := s.mostWinrateDebut(winrateByOpeningWhite)
	mostWinrateDebutBlack := s.mostWinrateDebut(winrateByOpeningBlack)

	var winrateLast10 float64
	if gamesCountLast10 > 0 {
		winrateLast10 = float64(winCounterLast10) / float64(gamesCountLast10) * 100
	}

	for _, acc := range accuracyListLast10 {
		avgAccuracyLast10 += acc
	}
	if len(accuracyListLast10) > 0 {
		avgAccuracyLast10 /= float64(len(accuracyListLast10))
	}

	return domain.Analysis{
		Speed:                 speed,
		Winrate:               winRate,
		MostPopularDebutWhite: mostPopularDebutWhite,
		MostPopularDebutBlack: mostPopularDebutBlack,
		MostWinrateDebutWhite: mostWinrateDebutWhite,
		MostWinrateDebutBlack: mostWinrateDebutBlack,
		WinrateLast10Days:     winrateLast10,
		AvgAccuracyLast10Days: avgAccuracyLast10,
		AvgAccuracy:           avgAccuracy,
		TiltFactor:            tiltFactor,
	}, nil
}

func (s *service) checkIfUserBlackOrWhite(game lichess.GameLichess, username string) (string, error) {
	if game.Players.Black.User != nil && game.Players.Black.User.Name == username {
		return "Black", nil
	}
	if game.Players.White.User != nil && game.Players.White.User.Name == username {
		return "White", nil
	}
	return "", fmt.Errorf("user %s not found in game", username)
}

func (s *service) mostPopularDebut(counter map[string]int) string {
	maxCount := 0
	result := ""
	for name, count := range counter {
		if count > maxCount {
			maxCount = count
			result = name
		}
	}
	return result
}

func (s *service) mostWinrateDebut(counter map[string]float64) string {
	var maxCount float64
	result := ""
	for name, count := range counter {
		if count > maxCount {
			maxCount = count
			result = name
		}
	}
	return result
}
