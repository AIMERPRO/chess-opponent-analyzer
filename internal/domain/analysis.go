package domain

import "time"

type Analysis struct {
	Speed                 string
	Winrate               float64
	WinrateLast10Days     float64
	MostPopularDebutWhite string
	MostWinrateDebutWhite string
	MostPopularDebutBlack string
	MostWinrateDebutBlack string
	AvgAccuracy           float64
	AvgAccuracyLast10Days float64
	TiltFactor            float64
	CachedAt              time.Time
}
