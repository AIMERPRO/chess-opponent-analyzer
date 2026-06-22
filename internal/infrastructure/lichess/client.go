package lichess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type GameLichess struct {
	ID        string   `json:"id"`
	Variant   string   `json:"variant"`
	Speed     string   `json:"speed"`
	Status    string   `json:"status"`
	Winner    *string  `json:"winner"`
	Opening   *Opening `json:"opening"`
	Players   Players  `json:"players"`
	CreatedAt int64    `json:"createdAt"`
}

type Opening struct {
	ECO  string `json:"eco"`
	Name string `json:"name"`
}

type Players struct {
	White Player `json:"white"`
	Black Player `json:"black"`
}

type PlayerAnalysis struct {
	Inaccuracy int `json:"inaccuracy"`
	Mistake    int `json:"mistake"`
	Blunder    int `json:"blunder"`
	Acpl       int `json:"acpl"`
	Accuracy   int `json:"accuracy"`
}

type Player struct {
	User     *User           `json:"user"`
	Rating   int             `json:"rating"`
	Analysis *PlayerAnalysis `json:"analysis"`
}
type User struct {
	Name string `json:"name"`
}

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		http:    &http.Client{},
		baseURL: baseURL,
	}
}

func (c *Client) GetUserGames(ctx context.Context, username string, speed string, limit int) ([]GameLichess, error) {
	lichessGetGamesURL := c.baseURL

	lichessGetGamesURL += username

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lichessGetGamesURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/x-ndjson")
	q := req.URL.Query()
	q.Add("speed", speed)
	q.Add("max", strconv.Itoa(limit))
	q.Add("opening", "true")  // чтобы получить данные о дебюте
	q.Add("accuracy", "true") // чтобы получить точность
	q.Add("variant", "standard")
	q.Add("rated", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lichess API returned status: %d", resp.StatusCode)
	}

	gamesLichess := make([]GameLichess, 0)

	respGames := bufio.NewScanner(resp.Body)
	for respGames.Scan() {
		line := respGames.Bytes()
		if len(line) == 0 {
			continue
		}

		var game GameLichess
		if err = json.Unmarshal(line, &game); err != nil {
			return nil, fmt.Errorf("failed to parse game: %w", err)
		}

		gamesLichess = append(gamesLichess, game)
	}

	if err = respGames.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan response: %w", err)
	}

	return gamesLichess, nil
}
