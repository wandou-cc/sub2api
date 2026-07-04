-- Add standalone license service storage.
CREATE TABLE IF NOT EXISTS license_codes (
    id BIGSERIAL PRIMARY KEY,
    code_id VARCHAR(64) NOT NULL UNIQUE,
    code VARCHAR(64) NOT NULL UNIQUE,
    license_id VARCHAR(64) UNIQUE,
    product VARCHAR(100) NOT NULL,
    product_batch VARCHAR(100) NOT NULL,
    features TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    status VARCHAR(20) NOT NULL DEFAULT 'unused',
    usb_fingerprint TEXT NOT NULL DEFAULT '',
    activated_at TIMESTAMPTZ,
    last_verified_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT license_codes_status_check CHECK (status IN ('unused', 'active', 'disabled', 'expired', 'revoked', 'refunded'))
);

CREATE INDEX IF NOT EXISTS idx_license_codes_status ON license_codes(status);
CREATE INDEX IF NOT EXISTS idx_license_codes_expires_at ON license_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_license_codes_product_batch ON license_codes(product, product_batch);
