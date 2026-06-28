package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const speedRankLimit = 3
const speedRankHistoryLimit = 14

var speedRankRewardByRank = map[int]float64{
	1: 5,
	2: 3,
	3: 2,
}

type SpeedRankEntry struct {
	Rank         int     `json:"rank"`
	UserID       int64   `json:"user_id"`
	Email        string  `json:"email"`
	Username     string  `json:"username"`
	RewardDate   string  `json:"reward_date,omitempty"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	Reward       float64 `json:"reward"`
}

type SpeedRankResponse struct {
	Entries       []SpeedRankEntry `json:"entries"`
	History       []SpeedRankEntry `json:"history"`
	NextRewardAt  time.Time        `json:"next_reward_at"`
	GeneratedAt   time.Time        `json:"generated_at"`
	RankingDate   string           `json:"ranking_date"`
	RewardAmounts map[int]float64  `json:"reward_amounts"`
}

type SpeedRankRewardRecord struct {
	RewardDate   time.Time
	Rank         int
	UserID       int64
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	RewardAmount float64
}

type SpeedRankRepository interface {
	GetDailyRanking(ctx context.Context, start, end time.Time, limit int) ([]SpeedRankEntry, error)
	GetDailyWinners(ctx context.Context, limit int) ([]SpeedRankEntry, error)
	GrantReward(ctx context.Context, reward SpeedRankRewardRecord) (bool, error)
}

type SpeedRankService struct {
	rankRepo             SpeedRankRepository
	authCacheInvalidator APIKeyAuthCacheInvalidator
	billingCache         *BillingCacheService
	stopCh               chan struct{}
	doneCh               chan struct{}
	startOnce            sync.Once
	stopOnce             sync.Once
}

// NewSpeedRankService creates the daily usage leaderboard service.
func NewSpeedRankService(
	rankRepo SpeedRankRepository,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	billingCache *BillingCacheService,
) *SpeedRankService {
	return &SpeedRankService{
		rankRepo:             rankRepo,
		authCacheInvalidator: authCacheInvalidator,
		billingCache:         billingCache,
		stopCh:               make(chan struct{}),
		doneCh:               make(chan struct{}),
	}
}

// GetTodayRanking returns today's top three users and the next reward time.
func (s *SpeedRankService) GetTodayRanking(ctx context.Context, now time.Time) (*SpeedRankResponse, error) {
	start := dayStart(now)
	end := start.AddDate(0, 0, 1)
	entries, err := s.rankRepo.GetDailyRanking(ctx, start, end, speedRankLimit)
	if err != nil {
		return nil, fmt.Errorf("get speed rank: %w", err)
	}
	history, err := s.rankRepo.GetDailyWinners(ctx, speedRankHistoryLimit)
	if err != nil {
		return nil, fmt.Errorf("get speed rank history: %w", err)
	}
	for i := range entries {
		entries[i].Reward = speedRankRewardByRank[entries[i].Rank]
	}
	return &SpeedRankResponse{
		Entries:      entries,
		History:      history,
		NextRewardAt: end,
		GeneratedAt:  now,
		RankingDate:  start.Format("2006-01-02"),
		RewardAmounts: map[int]float64{
			1: speedRankRewardByRank[1],
			2: speedRankRewardByRank[2],
			3: speedRankRewardByRank[3],
		},
	}, nil
}

// Start begins the midnight reward loop.
func (s *SpeedRankService) Start() {
	s.startOnce.Do(func() {
		go s.loop()
	})
}

// Stop ends the midnight reward loop.
func (s *SpeedRankService) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		<-s.doneCh
	})
}

// IssueRewardsForDay grants rewards for one completed local day.
func (s *SpeedRankService) IssueRewardsForDay(ctx context.Context, rewardDate time.Time) error {
	start := dayStart(rewardDate)
	end := start.AddDate(0, 0, 1)
	entries, err := s.rankRepo.GetDailyRanking(ctx, start, end, speedRankLimit)
	if err != nil {
		return fmt.Errorf("get speed rank reward candidates: %w", err)
	}
	for i := range entries {
		reward := speedRankRewardByRank[entries[i].Rank]
		if reward <= 0 {
			continue
		}
		inserted, err := s.rankRepo.GrantReward(ctx, SpeedRankRewardRecord{
			RewardDate:   start,
			Rank:         entries[i].Rank,
			UserID:       entries[i].UserID,
			InputTokens:  entries[i].InputTokens,
			OutputTokens: entries[i].OutputTokens,
			TotalTokens:  entries[i].TotalTokens,
			RewardAmount: reward,
		})
		if err != nil {
			return fmt.Errorf("insert speed rank reward: %w", err)
		}
		if !inserted {
			continue
		}
		s.invalidateBalanceCaches(ctx, entries[i].UserID)
	}
	return nil
}

func (s *SpeedRankService) loop() {
	defer close(s.doneCh)
	for {
		next := dayStart(time.Now()).AddDate(0, 0, 1)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-timer.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			if err := s.IssueRewardsForDay(ctx, next.AddDate(0, 0, -1)); err != nil {
				slog.Error("issue speed rank rewards failed", "date", next.AddDate(0, 0, -1).Format("2006-01-02"), "error", err)
			}
			cancel()
		case <-s.stopCh:
			timer.Stop()
			return
		}
	}
}

// invalidateBalanceCaches refreshes API key auth and balance eligibility after a reward.
func (s *SpeedRankService) invalidateBalanceCaches(ctx context.Context, userID int64) {
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	if s.billingCache != nil {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.billingCache.InvalidateUserBalance(cacheCtx, userID); err != nil {
				slog.Error("invalidate user balance cache failed", "user_id", userID, "error", err)
			}
		}()
	}
}

// dayStart returns the local midnight bucket for usage ranking.
func dayStart(t time.Time) time.Time {
	year, month, day := t.In(time.Local).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}
