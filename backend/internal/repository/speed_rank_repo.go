package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type speedRankRepository struct {
	sql sqlExecutor
}

// NewSpeedRankRepository creates the usage leaderboard repository.
func NewSpeedRankRepository(sqlDB *sql.DB) service.SpeedRankRepository {
	return &speedRankRepository{sql: sqlDB}
}

// GetDailyRanking returns the top users by daily input+output token usage.
func (r *speedRankRepository) GetDailyRanking(ctx context.Context, start, end time.Time, limit int) ([]service.SpeedRankEntry, error) {
	rows, err := r.sql.QueryContext(ctx, `
SELECT
    ROW_NUMBER() OVER (ORDER BY SUM(ul.input_tokens + ul.output_tokens) DESC, SUM(ul.output_tokens) DESC, MIN(u.id) ASC) AS rank,
    u.id,
    u.email,
    COALESCE(u.username, ''),
    SUM(ul.input_tokens)::bigint,
    SUM(ul.output_tokens)::bigint,
    SUM(ul.input_tokens + ul.output_tokens)::bigint
FROM usage_logs ul
JOIN users u ON u.id = ul.user_id AND u.deleted_at IS NULL
WHERE ul.created_at >= $1
  AND ul.created_at < $2
  AND (ul.input_tokens > 0 OR ul.output_tokens > 0)
GROUP BY u.id, u.email, u.username
ORDER BY SUM(ul.input_tokens + ul.output_tokens) DESC, SUM(ul.output_tokens) DESC, u.id ASC
LIMIT $3
`, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	entries := make([]service.SpeedRankEntry, 0, limit)
	for rows.Next() {
		var entry service.SpeedRankEntry
		if err := rows.Scan(
			&entry.Rank,
			&entry.UserID,
			&entry.Email,
			&entry.Username,
			&entry.InputTokens,
			&entry.OutputTokens,
			&entry.TotalTokens,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetDailyWinners returns the latest daily first-place reward records.
func (r *speedRankRepository) GetDailyWinners(ctx context.Context, limit int) ([]service.SpeedRankEntry, error) {
	rows, err := r.sql.QueryContext(ctx, `
SELECT
    1 AS rank,
    srr.user_id,
    u.email,
    COALESCE(u.username, ''),
    srr.reward_date::text,
    srr.input_tokens,
    srr.output_tokens,
    srr.total_tokens,
    srr.reward_amount::float8
FROM speed_rank_rewards srr
JOIN users u ON u.id = srr.user_id AND u.deleted_at IS NULL
WHERE srr.rank = 1
ORDER BY srr.reward_date DESC
LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	entries := make([]service.SpeedRankEntry, 0, limit)
	for rows.Next() {
		var entry service.SpeedRankEntry
		if err := rows.Scan(
			&entry.Rank,
			&entry.UserID,
			&entry.Email,
			&entry.Username,
			&entry.RewardDate,
			&entry.InputTokens,
			&entry.OutputTokens,
			&entry.TotalTokens,
			&entry.Reward,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// GrantReward records one daily rank reward and credits the user's balance atomically.
func (r *speedRankRepository) GrantReward(ctx context.Context, reward service.SpeedRankRewardRecord) (inserted bool, err error) {
	db, ok := r.sql.(*sql.DB)
	if !ok {
		return false, fmt.Errorf("speed rank repository requires *sql.DB")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
INSERT INTO speed_rank_rewards (
    reward_date,
    rank,
    user_id,
    input_tokens,
    output_tokens,
    total_tokens,
    reward_amount
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (reward_date, rank) DO NOTHING
`, reward.RewardDate, reward.Rank, reward.UserID, reward.InputTokens, reward.OutputTokens, reward.TotalTokens, reward.RewardAmount)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read speed rank reward insert result: %w", err)
	}
	if affected == 0 {
		if err = tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}

	result, err = tx.ExecContext(ctx, `
UPDATE users
SET balance = balance + $1,
    total_recharged = total_recharged + $1,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL
`, reward.RewardAmount, reward.UserID)
	if err != nil {
		return false, err
	}
	affected, err = result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read speed rank balance update result: %w", err)
	}
	if affected != 1 {
		return false, service.ErrUserNotFound
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
