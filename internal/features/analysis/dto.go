package analysis

type AnalyzeDTO struct {
	Speed                 string  `json:"speed"`
	Winrate               float64 `json:"winrate"`
	WinrateLast10Days     float64 `json:"winrate_last10_days"`
	MostPopularDebutWhite string  `json:"most_popular_debut_white"`
	MostWinrateDebutWhite string  `json:"most_winrate_debut_white"`
	MostPopularDebutBlack string  `json:"most_popular_debut_black"`
	MostWinrateDebutBlack string  `json:"most_winrate_debut_black"`
	AvgAccuracy           float64 `json:"avg_accuracy"`
	AvgAccuracyLast10Days float64 `json:"avg_accuracy_last10_days"`
	TiltFactor            float64 `json:"tilt_factor"`
}
