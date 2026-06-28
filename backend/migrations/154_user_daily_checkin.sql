CREATE TABLE IF NOT EXISTS user_daily_checkins (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    checkin_date DATE NOT NULL,
    reward       DECIMAL(20, 8) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_daily_checkins_user_date
    ON user_daily_checkins(user_id, checkin_date);

CREATE INDEX IF NOT EXISTS idx_user_daily_checkins_user_created
    ON user_daily_checkins(user_id, created_at DESC);
