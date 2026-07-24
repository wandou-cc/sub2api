package service

import (
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestBuildDailySeriesUsesSelectedTimezoneAndFillsMissingDays(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	start := time.Date(2026, 7, 20, 0, 0, 0, 0, location)
	end := time.Date(2026, 7, 23, 0, 0, 0, 0, location)
	firstPaidAt := time.Date(2026, 7, 19, 16, 30, 0, 0, time.UTC)
	lastPaidAt := time.Date(2026, 7, 22, 15, 59, 0, 0, time.UTC)

	series := buildDailySeries([]*dbent.PaymentOrder{
		{PayAmount: 10.125, PaidAt: &firstPaidAt},
		{PayAmount: 20, PaidAt: &lastPaidAt},
	}, start, end)

	require.Equal(t, []DailyStats{
		{Date: "2026-07-20", Amount: 10.13, Count: 1},
		{Date: "2026-07-21"},
		{Date: "2026-07-22", Amount: 20, Count: 1},
	}, series)
}

func TestBuildTopUsersAggregatesAllUsersAndSortsByAmount(t *testing.T) {
	orders := make([]*dbent.PaymentOrder, 0, 13)
	for userID := int64(1); userID <= 12; userID++ {
		orders = append(orders, &dbent.PaymentOrder{
			UserID:    userID,
			UserEmail: "user@example.com",
			PayAmount: float64(userID),
		})
	}
	orders = append(orders, &dbent.PaymentOrder{UserID: 1, UserEmail: "user@example.com", PayAmount: 20})

	users := buildTopUsers(orders)

	require.Len(t, users, 12)
	require.Equal(t, TopUserStat{UserID: 1, Email: "user@example.com", Amount: 21}, users[0])
	require.Equal(t, int64(12), users[1].UserID)
	require.Equal(t, int64(2), users[len(users)-1].UserID)
}
