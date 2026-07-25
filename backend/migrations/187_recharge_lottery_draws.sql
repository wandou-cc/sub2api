CREATE TABLE IF NOT EXISTS recharge_lottery_draws (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id        BIGINT NOT NULL REFERENCES payment_orders(id) ON DELETE CASCADE,
    recharge_amount DECIMAL(20, 2) NOT NULL,
    max_rarity      VARCHAR(20) NOT NULL,
    rarity          VARCHAR(20) NOT NULL DEFAULT '',
    reward_amount   DECIMAL(20, 2) NOT NULL DEFAULT 0,
    balance_after   DECIMAL(20, 8),
    claimed_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_recharge_lottery_draws_order_id
    ON recharge_lottery_draws(order_id);

CREATE INDEX IF NOT EXISTS idx_recharge_lottery_draws_user_id
    ON recharge_lottery_draws(user_id);

CREATE INDEX IF NOT EXISTS idx_recharge_lottery_draws_user_claimed_at
    ON recharge_lottery_draws(user_id, claimed_at);

CREATE INDEX IF NOT EXISTS idx_recharge_lottery_draws_created_at
    ON recharge_lottery_draws(created_at);
