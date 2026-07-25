//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"

	entdialect "entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestRechargeLotteryRuleBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount    float64
		maxRarity string
		rewardMin float64
		rewardMax float64
	}{
		{amount: 4.99, maxRarity: rechargeLotteryRarityRare, rewardMin: 0.10, rewardMax: 0.30},
		{amount: 5, maxRarity: rechargeLotteryRarityEpic, rewardMin: 0.20, rewardMax: 1},
		{amount: 19.99, maxRarity: rechargeLotteryRarityEpic, rewardMin: 0.20, rewardMax: 1},
		{amount: 20, maxRarity: rechargeLotteryRarityEpicPlus, rewardMin: 0.50, rewardMax: 4},
		{amount: 99.99, maxRarity: rechargeLotteryRarityEpicPlus, rewardMin: 0.50, rewardMax: 4},
		{amount: 100, maxRarity: rechargeLotteryRarityLegendary, rewardMin: 2, rewardMax: 18},
		{amount: 499.99, maxRarity: rechargeLotteryRarityLegendary, rewardMin: 2, rewardMax: 18},
		{amount: 500, maxRarity: rechargeLotteryRarityLegendary, rewardMin: 8, rewardMax: 60},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(strconv.FormatFloat(tt.amount, 'f', -1, 64), func(t *testing.T) {
			rule := rechargeLotteryRuleForAmount(tt.amount)
			require.Equal(t, tt.maxRarity, rule.MaxRarity)
			require.Equal(t, tt.rewardMin, rule.RewardMin)
			require.Equal(t, tt.rewardMax, rule.RewardMax)
		})
	}
}

func TestRechargeLotteryTierProbabilityBoundaries(t *testing.T) {
	t.Parallel()

	for _, rule := range rechargeLotteryRules {
		total := 0
		for _, tier := range rule.Tiers {
			total += tier.Probability
		}
		require.Equal(t, 100, total)
	}

	amount := 100.0
	require.Equal(t, rechargeLotteryRarityCommon, rechargeLotteryTierForRoll(amount, 39).Rarity)
	require.Equal(t, rechargeLotteryRarityRare, rechargeLotteryTierForRoll(amount, 40).Rarity)
	require.Equal(t, rechargeLotteryRarityRare, rechargeLotteryTierForRoll(amount, 69).Rarity)
	require.Equal(t, rechargeLotteryRarityEpic, rechargeLotteryTierForRoll(amount, 70).Rarity)
	require.Equal(t, rechargeLotteryRarityEpic, rechargeLotteryTierForRoll(amount, 87).Rarity)
	require.Equal(t, rechargeLotteryRarityEpicPlus, rechargeLotteryTierForRoll(amount, 88).Rarity)
	require.Equal(t, rechargeLotteryRarityEpicPlus, rechargeLotteryTierForRoll(amount, 96).Rarity)
	require.Equal(t, rechargeLotteryRarityLegendary, rechargeLotteryTierForRoll(amount, 97).Rarity)
}

func TestRechargeLotteryRewardIsCentAccurateAndCappedByRecharge(t *testing.T) {
	t.Parallel()

	tier := RechargeLotteryTier{Rarity: rechargeLotteryRarityRare, RewardMin: 0.19, RewardMax: 0.30}
	require.Equal(t, 0.19, rechargeLotteryRewardForOffset(5, tier, 0))
	require.Equal(t, 0.30, rechargeLotteryRewardForOffset(5, tier, 11))
	require.Equal(t, 0.20, rechargeLotteryRewardForOffset(0.20, tier, 11))
}

func TestMarkCompletedIssuesOnlyBalanceRechargeOpportunity(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := &PaymentService{entClient: client}

	balanceOrder := createRechargeLotteryOrder(t, ctx, client, payment.OrderTypeBalance, OrderStatusRecharging, 100)
	balanceLease := &paymentFulfillmentLease{version: balanceOrder.UpdatedAt}
	require.NoError(t, svc.markCompleted(ctx, balanceOrder, balanceLease, "RECHARGE_SUCCESS"))
	require.NoError(t, svc.markCompleted(ctx, balanceOrder, balanceLease, "RECHARGE_SUCCESS"))
	balanceDraw, err := client.RechargeLotteryDraw.Query().Only(ctx)
	require.NoError(t, err)
	require.Equal(t, balanceOrder.ID, balanceDraw.OrderID)
	require.Equal(t, rechargeLotteryRarityLegendary, balanceDraw.MaxRarity)

	subscriptionOrder := createRechargeLotteryOrder(t, ctx, client, payment.OrderTypeSubscription, OrderStatusRecharging, 20)
	subscriptionLease := &paymentFulfillmentLease{version: subscriptionOrder.UpdatedAt}
	require.NoError(t, svc.markCompleted(ctx, subscriptionOrder, subscriptionLease, "SUBSCRIPTION_SUCCESS"))
	count, err := client.RechargeLotteryDraw.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestDrawRechargeLotteryCreditsBalanceOnce(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	order := createRechargeLotteryOrder(t, ctx, client, payment.OrderTypeBalance, OrderStatusCompleted, 20)
	_, err := client.RechargeLotteryDraw.Create().
		SetUserID(order.UserID).
		SetOrderID(order.ID).
		SetRechargeAmount(order.Amount).
		SetMaxRarity(rechargeLotteryRarityEpicPlus).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	first, err := svc.DrawRechargeLottery(ctx, order.UserID, order.ID)
	require.NoError(t, err)
	require.True(t, first.Claimed)
	require.GreaterOrEqual(t, first.RewardAmount, 0.50)
	require.LessOrEqual(t, first.RewardAmount, 4.0)

	userAfterFirst, err := client.User.Get(ctx, order.UserID)
	require.NoError(t, err)
	require.Equal(t, first.RewardAmount, userAfterFirst.Balance)
	require.NotNil(t, first.ClaimedAt)
	require.NotNil(t, first.BalanceAfter)
	require.Equal(t, userAfterFirst.Balance, *first.BalanceAfter)

	second, err := svc.DrawRechargeLottery(ctx, order.UserID, order.ID)
	require.NoError(t, err)
	require.Equal(t, first.OrderID, second.OrderID)
	require.Equal(t, first.RechargeAmount, second.RechargeAmount)
	require.Equal(t, first.MaxRarity, second.MaxRarity)
	require.Equal(t, first.Claimed, second.Claimed)
	require.Equal(t, first.Rarity, second.Rarity)
	require.Equal(t, first.RewardAmount, second.RewardAmount)
	require.NotNil(t, second.ClaimedAt)
	require.True(t, first.ClaimedAt.Equal(*second.ClaimedAt))
	require.Equal(t, first.BalanceAfter, second.BalanceAfter)
	userAfterSecond, err := client.User.Get(ctx, order.UserID)
	require.NoError(t, err)
	require.Equal(t, userAfterFirst.Balance, userAfterSecond.Balance)
}

func TestDrawRechargeLotteryRejectsCrossUserReplay(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	ownerOrder := createRechargeLotteryOrder(t, ctx, client, payment.OrderTypeBalance, OrderStatusCompleted, 20)
	draw, err := client.RechargeLotteryDraw.Create().
		SetUserID(ownerOrder.UserID).
		SetOrderID(ownerOrder.ID).
		SetRechargeAmount(ownerOrder.Amount).
		SetMaxRarity(rechargeLotteryRarityEpicPlus).
		Save(ctx)
	require.NoError(t, err)
	attackerOrder := createRechargeLotteryOrder(t, ctx, client, payment.OrderTypeBalance, OrderStatusCompleted, 20)

	svc := &PaymentService{entClient: client}
	_, err = svc.DrawRechargeLottery(ctx, attackerOrder.UserID, ownerOrder.ID)
	require.Error(t, err)
	require.Equal(t, "RECHARGE_LOTTERY_NOT_AVAILABLE", infraerrors.Reason(err))

	reloadedDraw, err := client.RechargeLotteryDraw.Get(ctx, draw.ID)
	require.NoError(t, err)
	require.Nil(t, reloadedDraw.ClaimedAt)
	owner, err := client.User.Get(ctx, ownerOrder.UserID)
	require.NoError(t, err)
	attacker, err := client.User.Get(ctx, attackerOrder.UserID)
	require.NoError(t, err)
	require.Zero(t, owner.Balance)
	require.Zero(t, attacker.Balance)
}

func TestDrawRechargeLotteryRejectsReplayAfterRefund(t *testing.T) {
	ctx := context.Background()
	client := newRechargeLotteryLockTestClient(t)
	order := createRechargeLotteryOrder(t, ctx, client, payment.OrderTypeBalance, OrderStatusCompleted, 20)
	_, err := client.RechargeLotteryDraw.Create().
		SetUserID(order.UserID).
		SetOrderID(order.ID).
		SetRechargeAmount(order.Amount).
		SetMaxRarity(rechargeLotteryRarityEpicPlus).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	first, err := svc.DrawRechargeLottery(ctx, order.UserID, order.ID)
	require.NoError(t, err)
	_, err = client.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusRefunded).Save(ctx)
	require.NoError(t, err)

	_, err = svc.DrawRechargeLottery(ctx, order.UserID, order.ID)
	require.Error(t, err)
	require.Equal(t, "RECHARGE_LOTTERY_NOT_AVAILABLE", infraerrors.Reason(err))
	userAfterReplay, err := client.User.Get(ctx, order.UserID)
	require.NoError(t, err)
	require.Equal(t, first.RewardAmount, userAfterReplay.Balance)
}

func TestRechargeLotteryRewardIsReversedProportionallyOnRefund(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	order := createRechargeLotteryOrder(t, ctx, client, payment.OrderTypeBalance, OrderStatusCompleted, 20)
	claimedAt := time.Now()
	_, err := client.RechargeLotteryDraw.Create().
		SetUserID(order.UserID).
		SetOrderID(order.ID).
		SetRechargeAmount(order.Amount).
		SetMaxRarity(rechargeLotteryRarityEpicPlus).
		SetRarity(rechargeLotteryRarityEpic).
		SetRewardAmount(8).
		SetBalanceAfter(28).
		SetClaimedAt(claimedAt).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient: client,
		userRepo:  &mockUserRepo{getByIDUser: &User{ID: order.UserID, Balance: 100}},
	}
	plan := &RefundPlan{OrderID: order.ID, Order: order, RefundAmount: 10}
	require.NoError(t, svc.refreshBalanceRefundDeduction(ctx, plan))
	require.Equal(t, 14.0, plan.BalanceToDeduct)

	plan.RefundAmount = 20
	require.NoError(t, svc.refreshBalanceRefundDeduction(ctx, plan))
	require.Equal(t, 28.0, plan.BalanceToDeduct)
}

type rechargeLotteryLockTestDriver struct {
	entdialect.Driver
}

// Dialect enables the production locking query while the test wrapper removes its unsupported SQLite suffix.
func (rechargeLotteryLockTestDriver) Dialect() string {
	return entdialect.Postgres
}

// Query removes only the locking suffix before executing against the single-connection test database.
func (d rechargeLotteryLockTestDriver) Query(ctx context.Context, query string, args, rows any) error {
	return d.Driver.Query(ctx, strings.ReplaceAll(query, " FOR UPDATE", ""), args, rows)
}

// Tx wraps transaction queries with the same test-only locking suffix removal.
func (d rechargeLotteryLockTestDriver) Tx(ctx context.Context) (entdialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return rechargeLotteryLockTestTx{Tx: tx}, nil
}

type rechargeLotteryLockTestTx struct {
	entdialect.Tx
}

// Query removes only the locking suffix inside a test transaction.
func (tx rechargeLotteryLockTestTx) Query(ctx context.Context, query string, args, rows any) error {
	return tx.Tx.Query(ctx, strings.ReplaceAll(query, " FOR UPDATE", ""), args, rows)
}

// newRechargeLotteryLockTestClient creates a SQLite client that can exercise PostgreSQL lock paths serially.
func newRechargeLotteryLockTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	baseDriver := entsql.OpenDB(entdialect.SQLite, db)
	schemaClient := dbent.NewClient(dbent.Driver(baseDriver))
	require.NoError(t, schemaClient.Schema.Create(context.Background()))
	client := dbent.NewClient(dbent.Driver(rechargeLotteryLockTestDriver{Driver: baseDriver}))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// createRechargeLotteryOrder creates the minimal persisted order required by lottery service tests.
func createRechargeLotteryOrder(t *testing.T, ctx context.Context, client *dbent.Client, orderType, status string, amount float64) *dbent.PaymentOrder {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	user, err := client.User.Create().
		SetEmail("lottery-" + strconv.FormatInt(now.UnixNano(), 10) + "@example.com").
		SetPasswordHash("hash").
		SetUsername("lottery-user").
		Save(ctx)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(amount).
		SetPayAmount(amount).
		SetFeeRate(0).
		SetRechargeCode("LOTTERY-" + strconv.FormatInt(now.UnixNano(), 10)).
		SetOutTradeNo("sub2_lottery_" + strconv.FormatInt(now.UnixNano(), 10)).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("lottery-trade").
		SetOrderType(orderType).
		SetStatus(status).
		SetExpiresAt(now.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetUpdatedAt(now).
		Save(ctx)
	require.NoError(t, err)
	return order
}
