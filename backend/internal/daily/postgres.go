package daily

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore implements DailyStore using pgx/v5.
type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

// ── Users ─────────────────────────────────────────────────────────────────────

func (s *PostgresStore) UpsertUser(ctx context.Context, u *User) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			email        = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url   = EXCLUDED.avatar_url
	`, u.ID, u.Email, u.DisplayName, u.AvatarURL)
	return err
}

// ── Groups ────────────────────────────────────────────────────────────────────

func (s *PostgresStore) CreateGroup(ctx context.Context, g *Group) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO groups (id, name, invite_code, created_by)
		VALUES ($1, $2, $3, $4)
	`, g.ID, g.Name, g.InviteCode, g.CreatedBy)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)
	`, g.ID, g.CreatedBy)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	var g Group
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, invite_code, created_by, created_at
		FROM groups WHERE id = $1
	`, groupID).Scan(&g.ID, &g.Name, &g.InviteCode, &g.CreatedBy, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &g, err
}

func (s *PostgresStore) GetGroupByInviteCode(ctx context.Context, code string) (*Group, error) {
	var g Group
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, invite_code, created_by, created_at
		FROM groups WHERE invite_code = $1
	`, code).Scan(&g.ID, &g.Name, &g.InviteCode, &g.CreatedBy, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &g, err
}

func (s *PostgresStore) GetUserGroups(ctx context.Context, userID string) ([]Group, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT g.id, g.name, g.invite_code, g.created_by, g.created_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = $1
		ORDER BY g.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.InviteCode, &g.CreatedBy, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *PostgresStore) JoinGroup(ctx context.Context, groupID, userID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, groupID, userID)
	return err
}

func (s *PostgresStore) GetGroupMembers(ctx context.Context, groupID string) ([]GroupMember, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id, u.display_name, COALESCE(u.avatar_url, '')
		FROM users u
		JOIN group_members gm ON gm.user_id = u.id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []GroupMember
	for rows.Next() {
		var m GroupMember
		if err := rows.Scan(&m.ID, &m.DisplayName, &m.AvatarURL); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// ── Questions + Answers ───────────────────────────────────────────────────────

func (s *PostgresStore) GetTodayQuestion(ctx context.Context) (*DailyQuestion, error) {
	var q DailyQuestion
	err := s.pool.QueryRow(ctx, `
		SELECT id, question, active_date
		FROM daily_questions
		WHERE active_date = CURRENT_DATE
	`).Scan(&q.ID, &q.Question, &q.ActiveDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &q, err
}

func (s *PostgresStore) GetUserAnswer(ctx context.Context, userID, questionID string) (*Answer, error) {
	var a Answer
	err := s.pool.QueryRow(ctx, `
		SELECT id, question_id, user_id, answer_text, submitted_at
		FROM daily_answers
		WHERE user_id = $1 AND question_id = $2
	`, userID, questionID).Scan(&a.ID, &a.QuestionID, &a.UserID, &a.AnswerText, &a.SubmittedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (s *PostgresStore) SubmitAnswer(ctx context.Context, a *Answer) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO daily_answers (id, question_id, user_id, answer_text)
		VALUES ($1, $2, $3, $4)
	`, a.ID, a.QuestionID, a.UserID, a.AnswerText)
	return err
}

func (s *PostgresStore) GetGroupAnswerProgress(ctx context.Context, groupID, questionID string) (AnswerProgress, error) {
	var p AnswerProgress
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(da.id)::int     AS answered,
			COUNT(gm.user_id)::int AS total
		FROM group_members gm
		LEFT JOIN daily_answers da
			ON da.user_id = gm.user_id AND da.question_id = $2
		WHERE gm.group_id = $1
	`, groupID, questionID).Scan(&p.Answered, &p.Total)
	return p, err
}

func (s *PostgresStore) GetGroupAnswers(ctx context.Context, groupID, questionID string) ([]Answer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT da.id, da.question_id, da.user_id, da.answer_text, da.submitted_at
		FROM daily_answers da
		JOIN group_members gm ON gm.user_id = da.user_id
		WHERE gm.group_id = $1 AND da.question_id = $2
		ORDER BY da.submitted_at ASC
	`, groupID, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []Answer
	for rows.Next() {
		var a Answer
		if err := rows.Scan(&a.ID, &a.QuestionID, &a.UserID, &a.AnswerText, &a.SubmittedAt); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

// ── Guesses + Results ─────────────────────────────────────────────────────────

func (s *PostgresStore) HasUserSubmittedGuesses(ctx context.Context, guesserID, groupID, questionID string) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM daily_guesses
		WHERE guesser_id = $1 AND group_id = $2 AND question_id = $3
	`, guesserID, groupID, questionID).Scan(&count)
	return count > 0, err
}

func (s *PostgresStore) SubmitGuesses(ctx context.Context, guesses []Guess) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, g := range guesses {
		_, err := tx.Exec(ctx, `
			INSERT INTO daily_guesses
				(id, question_id, group_id, guesser_id, answer_id, guessed_user_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), g.QuestionID, g.GroupID, g.GuesserID, g.AnswerID, g.GuessedUserID)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) GetGroupResults(ctx context.Context, guesserID, groupID, questionID string) ([]GuessResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			da.id                  AS answer_id,
			da.answer_text,
			da.user_id             AS actual_user_id,
			u_actual.display_name  AS actual_name,
			dg.guessed_user_id,
			u_guessed.display_name AS guessed_name,
			(dg.guessed_user_id = da.user_id) AS correct
		FROM daily_guesses dg
		JOIN daily_answers da     ON da.id = dg.answer_id
		JOIN users u_actual       ON u_actual.id = da.user_id
		JOIN users u_guessed      ON u_guessed.id = dg.guessed_user_id
		WHERE dg.guesser_id = $1 AND dg.group_id = $2 AND dg.question_id = $3
		ORDER BY da.submitted_at ASC
	`, guesserID, groupID, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []GuessResult
	for rows.Next() {
		var r GuessResult
		if err := rows.Scan(
			&r.AnswerID, &r.AnswerText,
			&r.ActualUserID, &r.ActualName,
			&r.GuessedUserID, &r.GuessedName,
			&r.Correct,
		); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *PostgresStore) GetLeaderboard(ctx context.Context, groupID string) ([]LeaderboardEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			u.id,
			u.display_name,
			COALESCE(u.avatar_url, ''),
			COALESCE(SUM(
				CASE WHEN dg.guessed_user_id = da.user_id
				      AND da.user_id != dg.guesser_id
				THEN 1 ELSE 0 END
			), 0)::int AS score
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		LEFT JOIN daily_guesses dg
			ON dg.guesser_id = gm.user_id AND dg.group_id = $1
		LEFT JOIN daily_answers da ON da.id = dg.answer_id
		WHERE gm.group_id = $1
		GROUP BY u.id, u.display_name, u.avatar_url
		ORDER BY score DESC, u.display_name ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.DisplayName, &e.AvatarURL, &e.Score); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ── Chat ──────────────────────────────────────────────────────────────────────

func (s *PostgresStore) GetChatMessages(ctx context.Context, groupID, questionID string) ([]ChatMessage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.id, m.group_id, m.question_id, m.user_id,
		       u.display_name, COALESCE(u.avatar_url, ''),
		       m.message, m.created_at
		FROM group_chat_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.group_id = $1 AND m.question_id = $2
		ORDER BY m.created_at ASC
	`, groupID, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(
			&m.ID, &m.GroupID, &m.QuestionID, &m.UserID,
			&m.DisplayName, &m.AvatarURL,
			&m.Message, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (s *PostgresStore) SendChatMessage(ctx context.Context, m *ChatMessage) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO group_chat_messages (id, group_id, question_id, user_id, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, m.ID, m.GroupID, m.QuestionID, m.UserID, m.Message).Scan(&m.CreatedAt)
}
