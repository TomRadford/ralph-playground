package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
)

type server struct {
	db *sql.DB
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

func NewServer(db *sql.DB) (ServerInterface, error) {
	s := &server{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *server) GetHealth(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

func (s *server) ListPlayers(ctx *echo.Context) error {
	rows, err := s.db.Query(`
		SELECT id, name, created_at
		FROM players
		ORDER BY created_at ASC
	`)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}
	defer rows.Close()

	items := make([]player, 0)
	for rows.Next() {
		var (
			p             player
			createdAtText string
		)
		if err := rows.Scan(&p.ID, &p.Name, &createdAtText); err != nil {
			return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
		}

		createdAt, err := time.Parse(time.RFC3339Nano, createdAtText)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
		}
		p.CreatedAt = createdAt
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, playersResponse{Items: items})
}

func (s *server) CreatePlayer(ctx *echo.Context) error {
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

	if _, err := s.db.Exec(`
		INSERT INTO players (id, name, created_at)
		VALUES (?, ?, ?)
	`, p.ID, p.Name, p.CreatedAt.Format(time.RFC3339Nano)); err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}

	return ctx.JSON(http.StatusCreated, p)
}

func (s *server) ListMatches(ctx *echo.Context) error {
	rows, err := s.db.Query(`
		SELECT id, left_player1_id, left_player2_id, right_player1_id, right_player2_id,
		       left_score, right_score, played_at, created_at
		FROM matches
		ORDER BY played_at DESC, created_at DESC
	`)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}
	defer rows.Close()

	items := make([]match, 0)
	for rows.Next() {
		m, err := scanMatchRow(rows)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, matchesResponse{Items: items})
}

func (s *server) CreateMatch(ctx *echo.Context) error {
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

	if err := s.validateMatchSides(req.Left, req.Right); err != nil {
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

	leftPlayer2 := toNullableString(m.Left.Player2ID)
	rightPlayer2 := toNullableString(m.Right.Player2ID)
	if _, err := s.db.Exec(`
		INSERT INTO matches (
			id, left_player1_id, left_player2_id, right_player1_id, right_player2_id,
			left_score, right_score, played_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		m.ID,
		m.Left.Player1ID,
		leftPlayer2,
		m.Right.Player1ID,
		rightPlayer2,
		m.LeftScore,
		m.RightScore,
		m.PlayedAt.Format(time.RFC3339Nano),
		m.CreatedAt.Format(time.RFC3339Nano),
	); err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}

	return ctx.JSON(http.StatusCreated, m)
}

func (s *server) GetLeaderboard(ctx *echo.Context) error {
	players, err := s.listPlayersForStats()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}

	stats := make(map[string]*leaderboardEntry, len(players))
	for _, p := range players {
		stats[p.ID] = &leaderboardEntry{
			PlayerID:   p.ID,
			PlayerName: p.Name,
		}
	}

	matches, err := s.listMatchesForStats()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}
	for _, m := range matches {
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
	uniqueIDs := make([]string, 0, 4)

	sides := []matchSide{left, right}
	for _, side := range sides {
		playerIDs := collectPlayerIDs(side)
		if len(playerIDs) == 0 {
			return errors.New("each side must have at least one player")
		}
		for _, playerID := range playerIDs {
			if _, exists := allIDs[playerID]; exists {
				return errors.New("a player cannot appear multiple times in a match")
			}
			allIDs[playerID] = struct{}{}
			uniqueIDs = append(uniqueIDs, playerID)
		}
	}

	existing, err := s.existingPlayerIDs(uniqueIDs)
	if err != nil {
		return errors.New("internal server error")
	}
	if len(existing) != len(uniqueIDs) {
		return errors.New("all players in a match must already exist")
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

func (s *server) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS matches (
		id TEXT PRIMARY KEY,
		left_player1_id TEXT NOT NULL,
		left_player2_id TEXT,
		right_player1_id TEXT NOT NULL,
		right_player2_id TEXT,
		left_score INTEGER NOT NULL,
		right_score INTEGER NOT NULL,
		played_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(left_player1_id) REFERENCES players(id),
		FOREIGN KEY(left_player2_id) REFERENCES players(id),
		FOREIGN KEY(right_player1_id) REFERENCES players(id),
		FOREIGN KEY(right_player2_id) REFERENCES players(id)
	);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return err
	}
	return nil
}

func (s *server) existingPlayerIDs(ids []string) (map[string]struct{}, error) {
	if len(ids) == 0 {
		return map[string]struct{}{}, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT id
		FROM players
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]struct{}, len(ids))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *server) listPlayersForStats() ([]player, error) {
	rows, err := s.db.Query(`
		SELECT id, name, created_at
		FROM players
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]player, 0)
	for rows.Next() {
		var (
			p             player
			createdAtText string
		)
		if err := rows.Scan(&p.ID, &p.Name, &createdAtText); err != nil {
			return nil, err
		}
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtText)
		if err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *server) listMatchesForStats() ([]match, error) {
	rows, err := s.db.Query(`
		SELECT id, left_player1_id, left_player2_id, right_player1_id, right_player2_id,
		       left_score, right_score, played_at, created_at
		FROM matches
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]match, 0)
	for rows.Next() {
		m, err := scanMatchRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanMatchRow(s scanner) (match, error) {
	var (
		m           match
		leftP2      sql.NullString
		rightP2     sql.NullString
		playedAtRaw string
		createdRaw  string
	)

	err := s.Scan(
		&m.ID,
		&m.Left.Player1ID,
		&leftP2,
		&m.Right.Player1ID,
		&rightP2,
		&m.LeftScore,
		&m.RightScore,
		&playedAtRaw,
		&createdRaw,
	)
	if err != nil {
		return match{}, err
	}

	if leftP2.Valid && leftP2.String != "" {
		value := leftP2.String
		m.Left.Player2ID = &value
	}
	if rightP2.Valid && rightP2.String != "" {
		value := rightP2.String
		m.Right.Player2ID = &value
	}

	playedAt, err := time.Parse(time.RFC3339Nano, playedAtRaw)
	if err != nil {
		return match{}, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdRaw)
	if err != nil {
		return match{}, err
	}

	m.PlayedAt = playedAt
	m.CreatedAt = createdAt
	return m, nil
}

func toNullableString(value *string) interface{} {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}
