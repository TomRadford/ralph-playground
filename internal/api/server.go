package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type server struct {
	mu         sync.RWMutex
	players    []player
	playersByID map[string]player
	matches    []match
}

type errorResponse struct {
	Message string `json:"message"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type playersResponse struct {
	Items []player `json:"items"`
}

type player struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type createPlayerRequest struct {
	Name string `json:"name"`
}

type matchSide struct {
	Player1ID string  `json:"player1Id"`
	Player2ID *string `json:"player2Id,omitempty"`
}

type createMatchRequest struct {
	Left       matchSide `json:"left"`
	Right      matchSide `json:"right"`
	LeftScore  int       `json:"leftScore"`
	RightScore int       `json:"rightScore"`
	PlayedAt   string    `json:"playedAt"`
}

type match struct {
	ID         string    `json:"id"`
	Left       matchSide `json:"left"`
	Right      matchSide `json:"right"`
	LeftScore  int       `json:"leftScore"`
	RightScore int       `json:"rightScore"`
	PlayedAt   time.Time `json:"playedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

type matchesResponse struct {
	Items []match `json:"items"`
}

type leaderboardEntry struct {
	PlayerID       string `json:"playerId"`
	PlayerName     string `json:"playerName"`
	Matches        int    `json:"matches"`
	Wins           int    `json:"wins"`
	Losses         int    `json:"losses"`
	GoalsFor       int    `json:"goalsFor"`
	GoalsAgainst   int    `json:"goalsAgainst"`
	GoalDifference int    `json:"goalDifference"`
}

type leaderboardResponse struct {
	Items []leaderboardEntry `json:"items"`
}

func NewServer() ServerInterface {
	return &server{
		playersByID: make(map[string]player),
		players:     make([]player, 0),
		matches:     make([]match, 0),
	}
}

func (s *server) GetHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

func (s *server) ListPlayers(ctx echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]player, len(s.players))
	copy(items, s.players)
	return ctx.JSON(http.StatusOK, playersResponse{Items: items})
}

func (s *server) CreatePlayer(ctx echo.Context) error {
	var req createPlayerRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse{Message: "invalid request body"})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 100 {
		return ctx.JSON(http.StatusBadRequest, errorResponse{Message: "name must be between 1 and 100 characters"})
	}

	now := time.Now().UTC()
	p := player{
		ID:        newUUID(),
		Name:      name,
		CreatedAt: now,
	}

	s.mu.Lock()
	s.players = append(s.players, p)
	s.playersByID[p.ID] = p
	s.mu.Unlock()

	return ctx.JSON(http.StatusCreated, p)
}

func (s *server) ListMatches(ctx echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]match, len(s.matches))
	copy(items, s.matches)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].PlayedAt.After(items[j].PlayedAt)
	})

	return ctx.JSON(http.StatusOK, matchesResponse{Items: items})
}

func (s *server) CreateMatch(ctx echo.Context) error {
	var req createMatchRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse{Message: "invalid request body"})
	}

	if req.LeftScore < 0 || req.RightScore < 0 {
		return ctx.JSON(http.StatusBadRequest, errorResponse{Message: "scores must be >= 0"})
	}

	playedAt, err := time.Parse(time.RFC3339, req.PlayedAt)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse{Message: "playedAt must be RFC3339 date-time"})
	}

	s.mu.Lock()
	if err := s.validateMatchSides(req.Left, req.Right); err != nil {
		s.mu.Unlock()
		return ctx.JSON(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	m := match{
		ID:         newUUID(),
		Left:       req.Left,
		Right:      req.Right,
		LeftScore:  req.LeftScore,
		RightScore: req.RightScore,
		PlayedAt:   playedAt.UTC(),
		CreatedAt:  time.Now().UTC(),
	}

	s.matches = append(s.matches, m)
	s.mu.Unlock()

	return ctx.JSON(http.StatusCreated, m)
}

func (s *server) GetLeaderboard(ctx echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]*leaderboardEntry, len(s.players))
	for _, p := range s.players {
		stats[p.ID] = &leaderboardEntry{
			PlayerID:   p.ID,
			PlayerName: p.Name,
		}
	}

	for _, m := range s.matches {
		leftPlayers := collectPlayerIDs(m.Left)
		rightPlayers := collectPlayerIDs(m.Right)

		for _, playerID := range leftPlayers {
			entry := stats[playerID]
			entry.Matches++
			entry.GoalsFor += m.LeftScore
			entry.GoalsAgainst += m.RightScore
		}
		for _, playerID := range rightPlayers {
			entry := stats[playerID]
			entry.Matches++
			entry.GoalsFor += m.RightScore
			entry.GoalsAgainst += m.LeftScore
		}

		if m.LeftScore > m.RightScore {
			for _, playerID := range leftPlayers {
				stats[playerID].Wins++
			}
			for _, playerID := range rightPlayers {
				stats[playerID].Losses++
			}
		} else if m.RightScore > m.LeftScore {
			for _, playerID := range rightPlayers {
				stats[playerID].Wins++
			}
			for _, playerID := range leftPlayers {
				stats[playerID].Losses++
			}
		}
	}

	items := make([]leaderboardEntry, 0, len(stats))
	for _, entry := range stats {
		entry.GoalDifference = entry.GoalsFor - entry.GoalsAgainst
		items = append(items, *entry)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Wins != items[j].Wins {
			return items[i].Wins > items[j].Wins
		}
		if items[i].GoalDifference != items[j].GoalDifference {
			return items[i].GoalDifference > items[j].GoalDifference
		}
		if items[i].GoalsFor != items[j].GoalsFor {
			return items[i].GoalsFor > items[j].GoalsFor
		}
		return strings.ToLower(items[i].PlayerName) < strings.ToLower(items[j].PlayerName)
	})

	return ctx.JSON(http.StatusOK, leaderboardResponse{Items: items})
}

func (s *server) validateMatchSides(left matchSide, right matchSide) error {
	allIDs := make(map[string]struct{}, 4)

	sides := []matchSide{left, right}
	for _, side := range sides {
		playerIDs := collectPlayerIDs(side)
		if len(playerIDs) == 0 {
			return errors.New("each side must have at least one player")
		}
		for _, playerID := range playerIDs {
			if _, exists := s.playersByID[playerID]; !exists {
				return errors.New("all players in a match must already exist")
			}
			if _, exists := allIDs[playerID]; exists {
				return errors.New("a player cannot appear multiple times in a match")
			}
			allIDs[playerID] = struct{}{}
		}
	}

	return nil
}

func collectPlayerIDs(side matchSide) []string {
	ids := make([]string, 0, 2)
	if side.Player1ID != "" {
		ids = append(ids, side.Player1ID)
	}
	if side.Player2ID != nil && *side.Player2ID != "" {
		ids = append(ids, *side.Player2ID)
	}
	return ids
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	buf := make([]byte, 36)
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])

	return string(buf)
}
