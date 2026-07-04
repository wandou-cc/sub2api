package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type licenseRepository struct {
	db *sql.DB
}

// NewLicenseRepository creates repository access for standalone license records.
func NewLicenseRepository(db *sql.DB) service.LicenseRepository {
	return &licenseRepository{db: db}
}

// CreateCodes inserts generated activation codes.
func (r *licenseRepository) CreateCodes(ctx context.Context, codes []service.LicenseCode) error {
	if len(codes) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin license create codes: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for i := range codes {
		code := &codes[i]
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO license_codes (code_id, code, product, product_batch, features, status, expires_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		`, code.CodeID, code.Code, code.Product, code.ProductBatch, pq.Array(code.Features), code.Status, code.ExpiresAt, code.CreatedAt); err != nil {
			if isUniqueViolation(err) {
				return service.ErrLicenseCodeConflict.WithCause(err)
			}
			return fmt.Errorf("insert license code: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit license create codes: %w", err)
	}
	return nil
}

// ListCodes returns standalone license records in reverse creation order.
func (r *licenseRepository) ListCodes(ctx context.Context) ([]service.LicenseCode, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,
		       activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at
		FROM license_codes
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list license codes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var codes []service.LicenseCode
	for rows.Next() {
		code, err := scanLicenseCode(rows)
		if err != nil {
			return nil, err
		}
		codes = append(codes, *code)
	}
	return codes, rows.Err()
}

// Activate binds an activation code to one USB fingerprint.
func (r *licenseRepository) Activate(ctx context.Context, input service.LicenseActivateInput, licenseID string, now time.Time) (*service.LicenseCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin license activate: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	code, err := queryLicenseCodeForUpdate(ctx, tx, "code = $1", input.ActivationCode)
	if err != nil {
		return nil, translateLicenseNotFound(err, service.ErrLicenseCodeNotFound)
	}
	expired, err := expireLicenseCodeIfNeeded(ctx, tx, code, now)
	if err != nil {
		return nil, err
	}
	if expired {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit license expiration: %w", err)
		}
		return nil, service.ErrLicenseNotActive
	}
	if code.Product != input.Product || code.ProductBatch != input.ProductBatch {
		return nil, service.ErrLicenseProductMismatch
	}
	if code.Status == service.LicenseStatusDisabled || code.Status == service.LicenseStatusExpired ||
		code.Status == service.LicenseStatusRevoked || code.Status == service.LicenseStatusRefunded {
		return nil, service.ErrLicenseNotActive
	}
	if code.USBFingerprint != "" && code.USBFingerprint != input.USBFingerprint {
		if err := revokeLicenseCodeTx(ctx, tx, code, now, "fingerprint_mismatch"); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit fingerprint revoke: %w", err)
		}
		return nil, service.ErrLicenseFingerprintMismatch
	}
	if code.LicenseID == "" {
		row := tx.QueryRowContext(ctx, `
			UPDATE license_codes
			SET license_id = $2, status = $3, usb_fingerprint = $4, activated_at = $5, updated_at = $5
			WHERE id = $1
			RETURNING id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,
			          activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at
		`, code.ID, licenseID, service.LicenseStatusActive, input.USBFingerprint, now)
		code, err = scanLicenseCode(row)
		if err != nil {
			return nil, fmt.Errorf("activate license code: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit license activate: %w", err)
	}
	return code, nil
}

// Verify validates an active license binding and refreshes last_verified_at.
func (r *licenseRepository) Verify(ctx context.Context, input service.LicenseVerifyInput, now time.Time) (*service.LicenseCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin license verify: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	code, err := queryLicenseCodeForUpdate(ctx, tx, "code_id = $1 AND license_id = $2", input.CodeID, input.LicenseID)
	if err != nil {
		return nil, translateLicenseNotFound(err, service.ErrLicenseNotFound)
	}
	expired, err := expireLicenseCodeIfNeeded(ctx, tx, code, now)
	if err != nil {
		return nil, err
	}
	if expired {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit license expiration: %w", err)
		}
		return nil, service.ErrLicenseNotActive
	}
	if code.Product != input.Product {
		return nil, service.ErrLicenseProductMismatch
	}
	if code.Status != service.LicenseStatusActive {
		return nil, service.ErrLicenseNotActive
	}
	if code.USBFingerprint != input.USBFingerprint {
		if err := revokeLicenseCodeTx(ctx, tx, code, now, "fingerprint_mismatch"); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit fingerprint revoke: %w", err)
		}
		return nil, service.ErrLicenseFingerprintMismatch
	}
	row := tx.QueryRowContext(ctx, `
		UPDATE license_codes
		SET last_verified_at = $2, updated_at = $2
		WHERE id = $1
		RETURNING id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,
		          activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at
	`, code.ID, now)
	code, err = scanLicenseCode(row)
	if err != nil {
		return nil, fmt.Errorf("update license verification: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit license verify: %w", err)
	}
	return code, nil
}

// Deactivate revokes one license binding.
func (r *licenseRepository) Deactivate(ctx context.Context, input service.LicenseDeactivateInput, now time.Time) (*service.LicenseCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin license deactivate: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	code, err := queryLicenseCodeForUpdate(ctx, tx, "code_id = $1 AND license_id = $2", input.CodeID, input.LicenseID)
	if err != nil {
		return nil, translateLicenseNotFound(err, service.ErrLicenseNotFound)
	}
	if err := revokeLicenseCodeTx(ctx, tx, code, now, "deactivated"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit license deactivate: %w", err)
	}
	code.Status = service.LicenseStatusRevoked
	code.RevokedAt = &now
	code.RevokedReason = "deactivated"
	code.UpdatedAt = now
	return code, nil
}

// SetCodeStatus updates a code status for standalone admin operations.
func (r *licenseRepository) SetCodeStatus(ctx context.Context, codeID, status string, now time.Time) (*service.LicenseCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin license status update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	code, err := queryLicenseCodeForUpdate(ctx, tx, "code_id = $1", codeID)
	if err != nil {
		return nil, translateLicenseNotFound(err, service.ErrLicenseCodeNotFound)
	}
	if status == "enable" {
		newStatus := service.LicenseStatusUnused
		if code.LicenseID != "" {
			newStatus = service.LicenseStatusActive
		}
		row := tx.QueryRowContext(ctx, `
			UPDATE license_codes
			SET status = $2, revoked_at = NULL, revoked_reason = '', updated_at = $3
			WHERE id = $1
			RETURNING id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,
			          activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at
		`, code.ID, newStatus, now)
		code, err = scanLicenseCode(row)
	} else {
		row := tx.QueryRowContext(ctx, `
			UPDATE license_codes
			SET status = $2, revoked_at = $3, revoked_reason = $4, updated_at = $3
			WHERE id = $1
			RETURNING id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,
			          activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at
		`, code.ID, status, now, status)
		code, err = scanLicenseCode(row)
	}
	if err != nil {
		return nil, fmt.Errorf("update license code status: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit license status update: %w", err)
	}
	return code, nil
}

// UpdateCodeFeatures replaces the feature set on one activation code.
func (r *licenseRepository) UpdateCodeFeatures(ctx context.Context, codeID string, features []string, now time.Time) (*service.LicenseCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin license features update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	code, err := queryLicenseCodeForUpdate(ctx, tx, "code_id = $1", codeID)
	if err != nil {
		return nil, translateLicenseNotFound(err, service.ErrLicenseCodeNotFound)
	}
	row := tx.QueryRowContext(ctx, `
		UPDATE license_codes
		SET features = $2, updated_at = $3
		WHERE id = $1
		RETURNING id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,
		          activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at
	`, code.ID, pq.Array(features), now)
	code, err = scanLicenseCode(row)
	if err != nil {
		return nil, fmt.Errorf("update license code features: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit license features update: %w", err)
	}
	return code, nil
}

// RevokeLicense revokes a code by license_id.
func (r *licenseRepository) RevokeLicense(ctx context.Context, licenseID string, now time.Time) (*service.LicenseCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin license revoke: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	code, err := queryLicenseCodeForUpdate(ctx, tx, "license_id = $1", licenseID)
	if err != nil {
		return nil, translateLicenseNotFound(err, service.ErrLicenseNotFound)
	}
	if err := revokeLicenseCodeTx(ctx, tx, code, now, "license_revoked"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit license revoke: %w", err)
	}
	code.Status = service.LicenseStatusRevoked
	code.RevokedAt = &now
	code.RevokedReason = "license_revoked"
	code.UpdatedAt = now
	return code, nil
}

// queryLicenseCodeForUpdate loads one license row with a row lock.
func queryLicenseCodeForUpdate(ctx context.Context, tx *sql.Tx, where string, args ...any) (*service.LicenseCode, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,
		       activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at
		FROM license_codes
		WHERE `+where+`
		FOR UPDATE
	`, args...)
	return scanLicenseCode(row)
}

// expireLicenseCodeIfNeeded marks expired unused or active codes.
func expireLicenseCodeIfNeeded(ctx context.Context, tx *sql.Tx, code *service.LicenseCode, now time.Time) (bool, error) {
	if code.ExpiresAt == nil || code.ExpiresAt.After(now) {
		return false, nil
	}
	if code.Status != service.LicenseStatusUnused && code.Status != service.LicenseStatusActive {
		return false, nil
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE license_codes
		SET status = $2, revoked_at = $3, revoked_reason = 'expired', updated_at = $3
		WHERE id = $1
	`, code.ID, service.LicenseStatusExpired, now)
	if err != nil {
		return false, fmt.Errorf("expire license code: %w", err)
	}
	code.Status = service.LicenseStatusExpired
	code.RevokedAt = &now
	code.RevokedReason = "expired"
	code.UpdatedAt = now
	return true, nil
}

// revokeLicenseCodeTx revokes the locked license row.
func revokeLicenseCodeTx(ctx context.Context, tx *sql.Tx, code *service.LicenseCode, now time.Time, reason string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE license_codes
		SET status = $2, revoked_at = $3, revoked_reason = $4, updated_at = $3
		WHERE id = $1
	`, code.ID, service.LicenseStatusRevoked, now, reason)
	if err != nil {
		return fmt.Errorf("revoke license code: %w", err)
	}
	return nil
}

// translateLicenseNotFound maps sql no-row errors to license domain errors.
func translateLicenseNotFound(err error, notFound error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}
	return err
}

type licenseCodeScanner interface {
	Scan(dest ...any) error
}

// scanLicenseCode maps a database row into the license service model.
func scanLicenseCode(row licenseCodeScanner) (*service.LicenseCode, error) {
	var code service.LicenseCode
	var licenseID sql.NullString
	var activatedAt sql.NullTime
	var lastVerifiedAt sql.NullTime
	var expiresAt sql.NullTime
	var revokedAt sql.NullTime
	if err := row.Scan(
		&code.ID,
		&code.CodeID,
		&code.Code,
		&licenseID,
		&code.Product,
		&code.ProductBatch,
		pq.Array(&code.Features),
		&code.Status,
		&code.USBFingerprint,
		&activatedAt,
		&lastVerifiedAt,
		&expiresAt,
		&revokedAt,
		&code.RevokedReason,
		&code.CreatedAt,
		&code.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if licenseID.Valid {
		code.LicenseID = licenseID.String
	}
	if activatedAt.Valid {
		code.ActivatedAt = &activatedAt.Time
	}
	if lastVerifiedAt.Valid {
		code.LastVerifiedAt = &lastVerifiedAt.Time
	}
	if expiresAt.Valid {
		code.ExpiresAt = &expiresAt.Time
	}
	if revokedAt.Valid {
		code.RevokedAt = &revokedAt.Time
	}
	return &code, nil
}
