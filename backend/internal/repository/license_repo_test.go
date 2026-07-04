package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// TestLicenseRepositorySetCodeStatusRefundUsesSeparateReasonParameter verifies the SQL parameter contract for status and reason.
func TestLicenseRepositorySetCodeStatusRefundUsesSeparateReasonParameter(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 7, 3, 13, 21, 0, 0, time.UTC)
	createdAt := now.Add(-time.Hour)
	columns := []string{
		"id", "code_id", "code", "license_id", "product", "product_batch", "features", "status",
		"usb_fingerprint", "activated_at", "last_verified_at", "expires_at", "revoked_at",
		"revoked_reason", "created_at", "updated_at",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,\s+activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at\s+FROM license_codes\s+WHERE code_id = \$1\s+FOR UPDATE`).
		WithArgs("code_test").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			int64(1), "code_test", "UCLAW-TEST-0001", nil, "uclaw-usb", "dev-2026-06", "{openmontage}",
			service.LicenseStatusActive, "fp-one", now, nil, nil, nil, "", createdAt, createdAt,
		))
	mock.ExpectQuery(`UPDATE license_codes\s+SET status = \$2, revoked_at = \$3, revoked_reason = \$4, updated_at = \$3\s+WHERE id = \$1\s+RETURNING id, code_id, code, license_id, product, product_batch, features, status, usb_fingerprint,\s+activated_at, last_verified_at, expires_at, revoked_at, revoked_reason, created_at, updated_at`).
		WithArgs(int64(1), service.LicenseStatusRefunded, sqlmock.AnyArg(), service.LicenseStatusRefunded).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			int64(1), "code_test", "UCLAW-TEST-0001", nil, "uclaw-usb", "dev-2026-06", "{openmontage}",
			service.LicenseStatusRefunded, "fp-one", now, nil, nil, now, service.LicenseStatusRefunded, createdAt, now,
		))
	mock.ExpectCommit()

	repo := NewLicenseRepository(db)
	code, err := repo.SetCodeStatus(context.Background(), "code_test", service.LicenseStatusRefunded, now)

	require.NoError(t, err)
	require.Equal(t, service.LicenseStatusRefunded, code.Status)
	require.Equal(t, service.LicenseStatusRefunded, code.RevokedReason)
	require.NoError(t, mock.ExpectationsWereMet())
}
