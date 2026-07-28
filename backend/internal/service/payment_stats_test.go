//go:build unit

package service

import (
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

// TestBuildDailySeriesUsesSelectedTimezoneAndFillsMissingDays verifies range boundaries in the selected timezone.
func TestBuildDailySeriesUsesSelectedTimezoneAndFillsMissingDays(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	start := time.Date(2026, 7, 20, 0, 0, 0, 0, location)
	end := time.Date(2026, 7, 23, 0, 0, 0, 0, location)
	firstPaidAt := time.Date(2026, 7, 19, 16, 30, 0, 0, time.UTC)
	lastPaidAt := time.Date(2026, 7, 22, 15, 59, 0, 0, time.UTC)

	series := buildDailySeries([]*dbent.PaymentOrder{
		paymentStatsTestOrder(1, "user@example.com", "CNY", 10.125, &firstPaidAt),
		paymentStatsTestOrder(1, "user@example.com", "CNY", 20, &lastPaidAt),
	}, start, end)

	require.Equal(t, []DailyStats{
		{Date: "2026-07-20", Amount: CurrencyAmounts{"CNY": 10.13}, Count: 1},
		{Date: "2026-07-21", Amount: CurrencyAmounts{}},
		{Date: "2026-07-22", Amount: CurrencyAmounts{"CNY": 20}, Count: 1},
	}, series)
}

// TestBuildTopUsersAggregatesAllUsersAndSortsByAmount verifies the complete per-currency ranking.
func TestBuildTopUsersAggregatesAllUsersAndSortsByAmount(t *testing.T) {
	orders := make([]*dbent.PaymentOrder, 0, 13)
	for userID := int64(1); userID <= 12; userID++ {
		orders = append(orders, paymentStatsTestOrder(userID, "user@example.com", "CNY", float64(userID), nil))
	}
	orders = append(orders, paymentStatsTestOrder(1, "user@example.com", "CNY", 20, nil))

	users := buildTopUsers(orders)["CNY"]

	require.Len(t, users, 12)
	require.Equal(t, TopUserStat{UserID: 1, Email: "user@example.com", Amount: 21}, users[0])
	require.Equal(t, int64(12), users[1].UserID)
	require.Equal(t, int64(2), users[len(users)-1].UserID)
}

// TestComputeBasicStatsGroupsAmountsByCurrency verifies totals, averages, and unique users across currencies.
func TestComputeBasicStatsGroupsAmountsByCurrency(t *testing.T) {
	t.Parallel()

	todayStart := time.Date(2026, time.July, 25, 0, 0, 0, 0, time.UTC)
	yesterday := todayStart.Add(-time.Hour)
	today := todayStart.Add(time.Hour)
	orders := []*dbent.PaymentOrder{
		paymentStatsTestOrder(1, "alice@example.com", "CNY", 10, &today),
		paymentStatsTestOrder(2, "bob@example.com", "USD", 10, &today),
		paymentStatsTestOrder(1, "alice@example.com", "CNY", 5, &yesterday),
	}

	stats := &DashboardStats{}
	computeBasicStats(stats, orders, todayStart)

	require.Equal(t, CurrencyAmounts{"CNY": 15, "USD": 10}, stats.TotalAmount)
	require.Equal(t, CurrencyAmounts{"CNY": 10, "USD": 10}, stats.TodayAmount)
	require.Equal(t, CurrencyAmounts{"CNY": 7.5, "USD": 10}, stats.AvgAmount)
	require.Equal(t, 3, stats.TotalCount)
	require.Equal(t, 2, stats.TodayCount)
	require.Equal(t, 2, stats.UserCount)
}

// TestPaymentDashboardBreakdownsGroupAmountsAndRankingsByCurrency verifies every dashboard breakdown independently groups currencies.
func TestPaymentDashboardBreakdownsGroupAmountsAndRankingsByCurrency(t *testing.T) {
	t.Parallel()

	firstDay := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)
	secondDay := firstDay.AddDate(0, 0, 1)
	orders := []*dbent.PaymentOrder{
		paymentStatsTestOrder(1, "alice@example.com", "CNY", 5.555, &firstDay),
		paymentStatsTestOrder(2, "bob@example.com", "CNY", 10, &firstDay),
		paymentStatsTestOrder(1, "alice@example.com", "USD", 20, &secondDay),
		paymentStatsTestOrder(2, "bob@example.com", "USD", 10, &secondDay),
	}
	orders[0].PaymentType = "stripe"
	orders[1].PaymentType = "stripe"
	orders[2].PaymentType = "stripe"
	orders[3].PaymentType = "alipay"

	daily := buildDailySeries(orders, firstDay, firstDay.AddDate(0, 0, 2))
	require.Equal(t, []DailyStats{
		{Date: "2026-07-24", Amount: CurrencyAmounts{"CNY": 15.56}, Count: 2},
		{Date: "2026-07-25", Amount: CurrencyAmounts{"USD": 30}, Count: 2},
	}, daily)

	methods := buildMethodDistribution(orders)
	require.Equal(t, []PaymentMethodStat{
		{Type: "alipay", Amount: CurrencyAmounts{"USD": 10}, Count: 1},
		{Type: "stripe", Amount: CurrencyAmounts{"CNY": 15.56, "USD": 20}, Count: 3},
	}, methods)

	users := buildTopUsers(orders)
	require.Equal(t, TopUsersByCurrency{
		"CNY": {
			{UserID: 2, Email: "bob@example.com", Amount: 10},
			{UserID: 1, Email: "alice@example.com", Amount: 5.56},
		},
		"USD": {
			{UserID: 1, Email: "alice@example.com", Amount: 20},
			{UserID: 2, Email: "bob@example.com", Amount: 10},
		},
	}, users)
}

// paymentStatsTestOrder creates an order with the persisted provider currency snapshot used by dashboard aggregation.
func paymentStatsTestOrder(userID int64, email, currency string, amount float64, paidAt *time.Time) *dbent.PaymentOrder {
	return &dbent.PaymentOrder{
		UserID:           userID,
		UserEmail:        email,
		PayAmount:        amount,
		PaidAt:           paidAt,
		ProviderSnapshot: map[string]any{"currency": currency},
	}
}
