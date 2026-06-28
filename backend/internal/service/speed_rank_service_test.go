package service

import (
	"context"
	"testing"
	"time"
)

type speedRankRepoStub struct {
	entries        []SpeedRankEntry
	history        []SpeedRankEntry
	insertedByRank map[int]bool
	start          time.Time
	end            time.Time
	limit          int
	grants         []SpeedRankRewardRecord
}

// GetDailyRanking records the query window and returns configured rows.
func (s *speedRankRepoStub) GetDailyRanking(_ context.Context, start, end time.Time, limit int) ([]SpeedRankEntry, error) {
	s.start = start
	s.end = end
	s.limit = limit
	return s.entries, nil
}

// GetDailyWinners returns configured historical winners.
func (s *speedRankRepoStub) GetDailyWinners(context.Context, int) ([]SpeedRankEntry, error) {
	return s.history, nil
}

// GrantReward records reward attempts and returns configured idempotency result.
func (s *speedRankRepoStub) GrantReward(_ context.Context, reward SpeedRankRewardRecord) (bool, error) {
	s.grants = append(s.grants, reward)
	return s.insertedByRank[reward.Rank], nil
}

type speedRankAuthInvalidatorStub struct {
	userIDs []int64
}

// InvalidateAuthCacheByKey is unused by speed rank rewards.
func (s *speedRankAuthInvalidatorStub) InvalidateAuthCacheByKey(context.Context, string) {}

// InvalidateAuthCacheByUserID records rewarded users whose auth cache was invalidated.
func (s *speedRankAuthInvalidatorStub) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.userIDs = append(s.userIDs, userID)
}

// InvalidateAuthCacheByGroupID is unused by speed rank rewards.
func (s *speedRankAuthInvalidatorStub) InvalidateAuthCacheByGroupID(context.Context, int64) {}

func TestSpeedRankServiceGetTodayRankingRewards(t *testing.T) {
	now := time.Date(2026, 6, 28, 14, 30, 0, 0, time.Local)
	repo := &speedRankRepoStub{entries: []SpeedRankEntry{
		{Rank: 1, UserID: 11, InputTokens: 100, OutputTokens: 80, TotalTokens: 180},
		{Rank: 2, UserID: 12, InputTokens: 90, OutputTokens: 70, TotalTokens: 160},
		{Rank: 3, UserID: 13, InputTokens: 80, OutputTokens: 60, TotalTokens: 140},
	}}
	svc := NewSpeedRankService(repo, nil, nil)

	result, err := svc.GetTodayRanking(context.Background(), now)
	if err != nil {
		t.Fatalf("GetTodayRanking returned error: %v", err)
	}

	wantStart := time.Date(2026, 6, 28, 0, 0, 0, 0, time.Local)
	if !repo.start.Equal(wantStart) || !repo.end.Equal(wantStart.AddDate(0, 0, 1)) || repo.limit != 3 {
		t.Fatalf("query window mismatch: start=%s end=%s limit=%d", repo.start, repo.end, repo.limit)
	}
	if result.RankingDate != "2026-06-28" || !result.NextRewardAt.Equal(wantStart.AddDate(0, 0, 1)) {
		t.Fatalf("response dates mismatch: date=%s next=%s", result.RankingDate, result.NextRewardAt)
	}
	rewards := []float64{result.Entries[0].Reward, result.Entries[1].Reward, result.Entries[2].Reward}
	if rewards[0] != 5 || rewards[1] != 3 || rewards[2] != 2 {
		t.Fatalf("reward mapping mismatch: %#v", rewards)
	}
}

func TestSpeedRankServiceIssueRewardsInvalidatesOnlyInsertedRewards(t *testing.T) {
	rewardDate := time.Date(2026, 6, 27, 18, 0, 0, 0, time.Local)
	repo := &speedRankRepoStub{
		entries: []SpeedRankEntry{
			{Rank: 1, UserID: 21, InputTokens: 300, OutputTokens: 200, TotalTokens: 500},
			{Rank: 2, UserID: 22, InputTokens: 200, OutputTokens: 100, TotalTokens: 300},
		},
		insertedByRank: map[int]bool{1: true, 2: false},
	}
	auth := &speedRankAuthInvalidatorStub{}
	svc := NewSpeedRankService(repo, auth, nil)

	if err := svc.IssueRewardsForDay(context.Background(), rewardDate); err != nil {
		t.Fatalf("IssueRewardsForDay returned error: %v", err)
	}

	if len(repo.grants) != 2 {
		t.Fatalf("grant attempts = %d, want 2", len(repo.grants))
	}
	if repo.grants[0].RewardAmount != 5 || repo.grants[1].RewardAmount != 3 {
		t.Fatalf("grant rewards mismatch: %#v", repo.grants)
	}
	if len(auth.userIDs) != 1 || auth.userIDs[0] != 21 {
		t.Fatalf("invalidated users = %#v, want [21]", auth.userIDs)
	}
}
