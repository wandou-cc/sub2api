package service

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/rechargelotterydraw"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	rechargeLotteryRarityCommon    = "common"
	rechargeLotteryRarityRare      = "rare"
	rechargeLotteryRarityEpic      = "epic"
	rechargeLotteryRarityEpicPlus  = "epic_plus"
	rechargeLotteryRarityLegendary = "legendary"
)

// RechargeLotteryTier describes one transparent rarity probability and reward range.
type RechargeLotteryTier struct {
	Rarity      string  `json:"rarity"`
	Probability int     `json:"probability"`
	RewardMin   float64 `json:"reward_min"`
	RewardMax   float64 `json:"reward_max"`
}

// RechargeLotteryRule describes one recharge amount band. MaxRecharge equal to zero means no upper limit.
type RechargeLotteryRule struct {
	MinRecharge float64               `json:"min_recharge"`
	MaxRecharge float64               `json:"max_recharge"`
	MaxRarity   string                `json:"max_rarity"`
	RewardMin   float64               `json:"reward_min"`
	RewardMax   float64               `json:"reward_max"`
	Tiers       []RechargeLotteryTier `json:"tiers"`
}

// RechargeLotteryOpportunity is the user-facing opportunity and immutable draw result.
type RechargeLotteryOpportunity struct {
	OrderID        int64      `json:"order_id"`
	RechargeAmount float64    `json:"recharge_amount"`
	MaxRarity      string     `json:"max_rarity"`
	Claimed        bool       `json:"claimed"`
	Rarity         string     `json:"rarity"`
	RewardAmount   float64    `json:"reward_amount"`
	BalanceAfter   *float64   `json:"balance_after,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ClaimedAt      *time.Time `json:"claimed_at,omitempty"`
}

// RechargeLotteryOverview contains all pending opportunities and recent results.
type RechargeLotteryOverview struct {
	PendingCount  int                          `json:"pending_count"`
	Opportunities []RechargeLotteryOpportunity `json:"opportunities"`
}

var rechargeLotteryRules = []RechargeLotteryRule{
	{
		MinRecharge: 0, MaxRecharge: 5, MaxRarity: rechargeLotteryRarityRare, RewardMin: 0.10, RewardMax: 0.30,
		Tiers: []RechargeLotteryTier{
			{Rarity: rechargeLotteryRarityCommon, Probability: 80, RewardMin: 0.10, RewardMax: 0.18},
			{Rarity: rechargeLotteryRarityRare, Probability: 20, RewardMin: 0.19, RewardMax: 0.30},
		},
	},
	{
		MinRecharge: 5, MaxRecharge: 20, MaxRarity: rechargeLotteryRarityEpic, RewardMin: 0.20, RewardMax: 1,
		Tiers: []RechargeLotteryTier{
			{Rarity: rechargeLotteryRarityCommon, Probability: 65, RewardMin: 0.20, RewardMax: 0.40},
			{Rarity: rechargeLotteryRarityRare, Probability: 28, RewardMin: 0.41, RewardMax: 0.70},
			{Rarity: rechargeLotteryRarityEpic, Probability: 7, RewardMin: 0.71, RewardMax: 1},
		},
	},
	{
		MinRecharge: 20, MaxRecharge: 100, MaxRarity: rechargeLotteryRarityEpicPlus, RewardMin: 0.50, RewardMax: 4,
		Tiers: []RechargeLotteryTier{
			{Rarity: rechargeLotteryRarityCommon, Probability: 50, RewardMin: 0.50, RewardMax: 1},
			{Rarity: rechargeLotteryRarityRare, Probability: 32, RewardMin: 1.01, RewardMax: 1.80},
			{Rarity: rechargeLotteryRarityEpic, Probability: 15, RewardMin: 1.81, RewardMax: 2.80},
			{Rarity: rechargeLotteryRarityEpicPlus, Probability: 3, RewardMin: 2.81, RewardMax: 4},
		},
	},
	{
		MinRecharge: 100, MaxRecharge: 500, MaxRarity: rechargeLotteryRarityLegendary, RewardMin: 2, RewardMax: 18,
		Tiers: []RechargeLotteryTier{
			{Rarity: rechargeLotteryRarityCommon, Probability: 40, RewardMin: 2, RewardMax: 4},
			{Rarity: rechargeLotteryRarityRare, Probability: 30, RewardMin: 4.01, RewardMax: 7},
			{Rarity: rechargeLotteryRarityEpic, Probability: 18, RewardMin: 7.01, RewardMax: 10.50},
			{Rarity: rechargeLotteryRarityEpicPlus, Probability: 9, RewardMin: 10.51, RewardMax: 14.50},
			{Rarity: rechargeLotteryRarityLegendary, Probability: 3, RewardMin: 14.51, RewardMax: 18},
		},
	},
	{
		MinRecharge: 500, MaxRecharge: 0, MaxRarity: rechargeLotteryRarityLegendary, RewardMin: 8, RewardMax: 60,
		Tiers: []RechargeLotteryTier{
			{Rarity: rechargeLotteryRarityCommon, Probability: 20, RewardMin: 8, RewardMax: 14},
			{Rarity: rechargeLotteryRarityRare, Probability: 28, RewardMin: 14.01, RewardMax: 22},
			{Rarity: rechargeLotteryRarityEpic, Probability: 27, RewardMin: 22.01, RewardMax: 32},
			{Rarity: rechargeLotteryRarityEpicPlus, Probability: 18, RewardMin: 32.01, RewardMax: 45},
			{Rarity: rechargeLotteryRarityLegendary, Probability: 7, RewardMin: 45.01, RewardMax: 60},
		},
	},
}

// rechargeLotteryRuleForAmount returns the immutable rule snapshot for a credited recharge amount.
func rechargeLotteryRuleForAmount(amount float64) RechargeLotteryRule {
	for _, rule := range rechargeLotteryRules {
		if rule.MaxRecharge == 0 || amount < rule.MaxRecharge {
			return rule
		}
	}
	panic("recharge lottery rules do not cover the credited amount")
}

// rechargeLotteryTierForRoll maps a 0-99 probability roll into the advertised tier.
func rechargeLotteryTierForRoll(amount float64, roll int64) RechargeLotteryTier {
	rule := rechargeLotteryRuleForAmount(amount)
	cumulative := int64(0)
	for _, tier := range rule.Tiers {
		cumulative += int64(tier.Probability)
		if roll < cumulative {
			return tier
		}
	}
	panic("recharge lottery tier probabilities do not cover the roll")
}

// rechargeLotteryRewardForOffset returns a cent-accurate reward and never awards more than the recharge itself.
func rechargeLotteryRewardForOffset(amount float64, tier RechargeLotteryTier, offset int64) float64 {
	minCents := decimal.NewFromFloat(tier.RewardMin).Mul(decimal.NewFromInt(100)).IntPart()
	rewardCents := minCents + offset
	rechargeCents := decimal.NewFromFloat(amount).Mul(decimal.NewFromInt(100)).Round(0).IntPart()
	if rewardCents > rechargeCents {
		rewardCents = rechargeCents
	}
	if rewardCents < 1 {
		rewardCents = 1
	}
	return decimal.NewFromInt(rewardCents).Div(decimal.NewFromInt(100)).InexactFloat64()
}

// drawRechargeLotteryReward uses cryptographic randomness for both rarity and amount selection.
func drawRechargeLotteryReward(amount float64) (string, float64, error) {
	rarityRoll, err := cryptorand.Int(cryptorand.Reader, big.NewInt(100))
	if err != nil {
		return "", 0, fmt.Errorf("generate lottery rarity roll: %w", err)
	}
	tier := rechargeLotteryTierForRoll(amount, rarityRoll.Int64())
	minCents := decimal.NewFromFloat(tier.RewardMin).Mul(decimal.NewFromInt(100)).IntPart()
	maxCents := decimal.NewFromFloat(tier.RewardMax).Mul(decimal.NewFromInt(100)).IntPart()
	amountRoll, err := cryptorand.Int(cryptorand.Reader, big.NewInt(maxCents-minCents+1))
	if err != nil {
		return "", 0, fmt.Errorf("generate lottery amount roll: %w", err)
	}
	return tier.Rarity, rechargeLotteryRewardForOffset(amount, tier, amountRoll.Int64()), nil
}

// GetRechargeLotteryOverview lists every pending opportunity plus the ten most recent results.
func (s *PaymentService) GetRechargeLotteryOverview(ctx context.Context, userID int64) (*RechargeLotteryOverview, error) {
	pending, err := s.entClient.RechargeLotteryDraw.Query().
		Where(
			rechargelotterydraw.UserIDEQ(userID),
			rechargelotterydraw.ClaimedAtIsNil(),
			rechargelotterydraw.HasOrderWith(paymentorder.StatusEQ(OrderStatusCompleted)),
		).
		Order(dbent.Desc(rechargelotterydraw.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending recharge lottery draws: %w", err)
	}
	recent, err := s.entClient.RechargeLotteryDraw.Query().
		Where(rechargelotterydraw.UserIDEQ(userID), rechargelotterydraw.ClaimedAtNotNil()).
		Order(dbent.Desc(rechargelotterydraw.FieldClaimedAt)).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recent recharge lottery draws: %w", err)
	}

	opportunities := make([]RechargeLotteryOpportunity, 0, len(pending)+len(recent))
	for _, draw := range append(pending, recent...) {
		opportunities = append(opportunities, rechargeLotteryOpportunityFromEntity(draw))
	}
	return &RechargeLotteryOverview{
		PendingCount:  len(pending),
		Opportunities: opportunities,
	}, nil
}

// DrawRechargeLottery claims one order-bound opportunity and credits the reward exactly once.
func (s *PaymentService) DrawRechargeLottery(ctx context.Context, userID, orderID int64) (*RechargeLotteryOpportunity, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin recharge lottery transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.PaymentOrder.Query().
		Where(
			paymentorder.IDEQ(orderID),
			paymentorder.UserIDEQ(userID),
			paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
			paymentorder.StatusEQ(OrderStatusCompleted),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.BadRequest("RECHARGE_LOTTERY_NOT_AVAILABLE", "this order has no available recharge lottery draw")
		}
		return nil, fmt.Errorf("lock recharge lottery order: %w", err)
	}

	draw, err := tx.RechargeLotteryDraw.Query().
		Where(rechargelotterydraw.OrderIDEQ(orderID), rechargelotterydraw.UserIDEQ(userID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.BadRequest("RECHARGE_LOTTERY_NOT_AVAILABLE", "this order has no available recharge lottery draw")
		}
		return nil, fmt.Errorf("lock recharge lottery draw: %w", err)
	}
	if draw.ClaimedAt != nil {
		result := rechargeLotteryOpportunityFromEntity(draw)
		return &result, nil
	}

	rarity, rewardAmount, err := drawRechargeLotteryReward(draw.RechargeAmount)
	if err != nil {
		return nil, err
	}
	updatedUser, err := tx.User.UpdateOneID(userID).AddBalance(rewardAmount).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("credit recharge lottery reward: %w", err)
	}
	claimedAt := time.Now()
	draw, err = tx.RechargeLotteryDraw.UpdateOneID(draw.ID).
		SetRarity(rarity).
		SetRewardAmount(rewardAmount).
		SetBalanceAfter(updatedUser.Balance).
		SetClaimedAt(claimedAt).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("save recharge lottery result: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit recharge lottery transaction: %w", err)
	}

	if s.redeemService != nil {
		s.redeemService.invalidateBalanceCaches(ctx, userID)
	}
	s.writeAuditLog(ctx, orderID, "RECHARGE_LOTTERY_CLAIMED", fmt.Sprintf("user:%d", userID), map[string]any{
		"rarity":       rarity,
		"rewardAmount": rewardAmount,
		"balanceAfter": updatedUser.Balance,
	})
	result := rechargeLotteryOpportunityFromEntity(draw)
	return &result, nil
}

// rechargeLotteryOpportunityFromEntity maps the persisted draw contract to its API contract.
func rechargeLotteryOpportunityFromEntity(draw *dbent.RechargeLotteryDraw) RechargeLotteryOpportunity {
	return RechargeLotteryOpportunity{
		OrderID:        draw.OrderID,
		RechargeAmount: draw.RechargeAmount,
		MaxRarity:      draw.MaxRarity,
		Claimed:        draw.ClaimedAt != nil,
		Rarity:         draw.Rarity,
		RewardAmount:   draw.RewardAmount,
		BalanceAfter:   draw.BalanceAfter,
		CreatedAt:      draw.CreatedAt,
		ClaimedAt:      draw.ClaimedAt,
	}
}

// rechargeLotteryRefundDeduction returns the proportional reward reversal for a refunded recharge.
func (s *PaymentService) rechargeLotteryRefundDeduction(ctx context.Context, client *dbent.Client, orderID int64, refundAmount, orderAmount float64) (float64, error) {
	draw, err := client.RechargeLotteryDraw.Query().
		Where(rechargelotterydraw.OrderIDEQ(orderID), rechargelotterydraw.ClaimedAtNotNil()).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("get recharge lottery reward for refund: %w", err)
	}
	if math.Abs(refundAmount-orderAmount) <= amountToleranceCNY {
		return draw.RewardAmount, nil
	}
	return decimal.NewFromFloat(draw.RewardAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(2).
		InexactFloat64(), nil
}
