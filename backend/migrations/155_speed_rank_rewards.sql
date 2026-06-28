CREATE TABLE IF NOT EXISTS speed_rank_rewards (
    id              BIGSERIAL PRIMARY KEY,
    reward_date     DATE NOT NULL,
    rank            INT NOT NULL,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    input_tokens    BIGINT NOT NULL DEFAULT 0,
    output_tokens   BIGINT NOT NULL DEFAULT 0,
    total_tokens    BIGINT NOT NULL DEFAULT 0,
    reward_amount   DECIMAL(20, 8) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_speed_rank_rewards_date_rank
    ON speed_rank_rewards(reward_date, rank);
CREATE UNIQUE INDEX IF NOT EXISTS idx_speed_rank_rewards_date_user
    ON speed_rank_rewards(reward_date, user_id);
CREATE INDEX IF NOT EXISTS idx_speed_rank_rewards_user_created
    ON speed_rank_rewards(user_id, created_at DESC);
