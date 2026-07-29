CREATE TABLE IF NOT EXISTS carpool_plans (
    id BIGSERIAL PRIMARY KEY,
    total_amount DECIMAL(20,2) NOT NULL CHECK (total_amount > 0),
    target_members INTEGER NOT NULL CHECK (target_members > 0),
    note TEXT NOT NULL CHECK (BTRIM(note) <> ''),
    revision INTEGER NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS carpool_plans_target_members ON carpool_plans(target_members);
CREATE INDEX IF NOT EXISTS carpool_plans_created_at ON carpool_plans(created_at);

INSERT INTO carpool_plans (total_amount, target_members, note)
SELECT 1600.00, 4, '4 人共享一个独立账号，不提供账号密码。首位用户付款后 48 小时内满 4 人成团；未成团将由管理员全额原路退款。成团后由管理员采购账号并开通专属分组，服务期从实际开通起计算 30 天。严禁生成违法违规内容；因用户违规导致风控或封号，一律不退款。非用户违规导致的官方风控或封号，是否退款由官方决定；官方审核通过后按实付金额和剩余服务天数退款。'
WHERE NOT EXISTS (
    SELECT 1 FROM carpool_plans WHERE total_amount = 1600.00 AND target_members = 4
);

INSERT INTO carpool_plans (total_amount, target_members, note)
SELECT 1600.00, 2, '2 人共享一个独立账号，不提供账号密码。首位用户付款后 48 小时内满 2 人成团；未成团将由管理员全额原路退款。成团后由管理员采购账号并开通专属分组，服务期从实际开通起计算 30 天。严禁生成违法违规内容；因用户违规导致风控或封号，一律不退款。非用户违规导致的官方风控或封号，是否退款由官方决定；官方审核通过后按实付金额和剩余服务天数退款。'
WHERE NOT EXISTS (
    SELECT 1 FROM carpool_plans WHERE total_amount = 1600.00 AND target_members = 2
);

INSERT INTO carpool_plans (total_amount, target_members, note)
SELECT 1400.00, 1, '1 人独立购买，无需等待成团，可由管理员提供账号密码。付款后由管理员采购账号并开通专属分组，服务期从实际开通起计算 30 天。严禁生成违法违规内容；因用户违规导致风控或封号，一律不退款。非用户违规导致的官方风控或封号，是否退款由官方决定；官方审核通过后按实付金额和剩余服务天数退款。'
WHERE NOT EXISTS (
    SELECT 1 FROM carpool_plans WHERE total_amount = 1400.00 AND target_members = 1
);

CREATE TABLE IF NOT EXISTS carpool_groups (
    id BIGSERIAL PRIMARY KEY,
    carpool_plan_id BIGINT NOT NULL,
    carpool_plan_revision INTEGER NOT NULL CHECK (carpool_plan_revision > 0),
    target_members INTEGER NOT NULL CHECK (target_members > 0),
    total_amount DECIMAL(20,2) NOT NULL CHECK (total_amount > 0),
    price_per_member DECIMAL(20,2) NOT NULL CHECK (price_per_member > 0),
    plan_note TEXT NOT NULL,
    member_count INTEGER NOT NULL DEFAULT 0 CHECK (member_count >= 0 AND member_count <= target_members),
    status VARCHAR(30) NOT NULL,
    open_key VARCHAR(32) UNIQUE,
    status_reason TEXT,
    deadline_at TIMESTAMPTZ,
    formed_at TIMESTAMPTZ,
    opened_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS carpool_groups_carpool_plan_id ON carpool_groups(carpool_plan_id);
CREATE INDEX IF NOT EXISTS carpool_groups_status ON carpool_groups(status);
CREATE INDEX IF NOT EXISTS carpool_groups_deadline_at ON carpool_groups(deadline_at);
CREATE INDEX IF NOT EXISTS carpool_groups_expires_at ON carpool_groups(expires_at);
CREATE INDEX IF NOT EXISTS carpool_groups_created_at ON carpool_groups(created_at);

ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS carpool_size INTEGER;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS carpool_plan_id BIGINT;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS carpool_plan_revision INTEGER;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS carpool_total_amount DECIMAL(20,2);
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS carpool_plan_note TEXT;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS carpool_group_id BIGINT REFERENCES carpool_groups(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS payment_orders_carpool_group_id_user_id_key
    ON payment_orders(carpool_group_id, user_id);

DROP INDEX IF EXISTS payment_orders_active_carpool_order_key;
CREATE UNIQUE INDEX payment_orders_active_carpool_order_key
    ON payment_orders(user_id, carpool_plan_id)
    WHERE order_type = 'carpool'
      AND carpool_group_id IS NULL
      AND (
          status IN ('PENDING', 'PAID', 'RECHARGING')
          OR (status = 'FAILED' AND paid_at IS NOT NULL)
      );
