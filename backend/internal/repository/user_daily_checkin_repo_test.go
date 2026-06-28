//go:build unit

package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// TestClaimDailyCheckin_RecordsClaimBeforeAddingBalance verifies successful claims update balance once.
func TestClaimDailyCheckin_RecordsClaimBeforeAddingBalance(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := newUserRepositoryWithSQL(nil, db)
	today := time.Date(2026, 6, 28, 12, 0, 0, 0, time.Local)
	checkinDay := checkinDate(today)
	claimedAt := time.Date(2026, 6, 28, 4, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`
WITH user_row AS (
    SELECT id
    FROM users
    WHERE id = $1 AND deleted_at IS NULL
    FOR UPDATE
),
claimed AS (
    INSERT INTO user_daily_checkins (user_id, checkin_date, reward)
    SELECT id, $3::date, $2
    FROM user_row
    ON CONFLICT (user_id, checkin_date) DO NOTHING
    RETURNING created_at
),
updated AS (
    UPDATE users
    SET balance = balance + $2,
        updated_at = NOW()
    WHERE id = (SELECT id FROM user_row)
      AND EXISTS (SELECT 1 FROM claimed)
    RETURNING balance
)
SELECT updated.balance, claimed.created_at
FROM updated, claimed
`)).
		WithArgs(int64(42), 0.5, checkinDay).
		WillReturnRows(sqlmock.NewRows([]string{"balance", "created_at"}).AddRow(10.5, claimedAt))

	status, err := repo.ClaimDailyCheckin(ctx, 42, today, 0.5)
	require.NoError(t, err)
	require.True(t, status.Claimed)
	require.Equal(t, claimedAt, *status.ClaimedAt)
	require.InDelta(t, 10.5, status.Balance, 0.000001)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestClaimDailyCheckin_ReturnsClaimedWhenDailyRecordAlreadyExists verifies replayed claims do not add balance.
func TestClaimDailyCheckin_ReturnsClaimedWhenDailyRecordAlreadyExists(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := newUserRepositoryWithSQL(nil, db)
	today := time.Date(2026, 6, 28, 12, 0, 0, 0, time.Local)
	checkinDay := checkinDate(today)
	claimedAt := time.Date(2026, 6, 28, 4, 0, 0, 0, time.UTC)

	mock.ExpectQuery("WITH user_row AS \\(").
		WithArgs(int64(42), 0.5, checkinDay).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT uc.id IS NOT NULL, uc.created_at, u.balance
FROM users u
LEFT JOIN user_daily_checkins uc ON uc.user_id = u.id AND uc.checkin_date = $2::date
WHERE u.id = $1 AND u.deleted_at IS NULL
`)).
		WithArgs(int64(42), checkinDay).
		WillReturnRows(sqlmock.NewRows([]string{"claimed", "created_at", "balance"}).AddRow(true, claimedAt, 10.5))

	status, err := repo.ClaimDailyCheckin(ctx, 42, today, 0.5)
	require.Error(t, err)
	require.True(t, errors.Is(err, service.ErrDailyCheckinClaimed))
	require.True(t, status.Claimed)
	require.Equal(t, claimedAt, *status.ClaimedAt)
	require.InDelta(t, 10.5, status.Balance, 0.000001)
	require.NoError(t, mock.ExpectationsWereMet())
}
