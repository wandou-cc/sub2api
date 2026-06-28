//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckPasswordSupportsBcrypt(t *testing.T) {
	user := &User{}
	require.NoError(t, user.SetPassword("new-password"))

	require.True(t, user.CheckPassword("new-password"))
	require.False(t, user.CheckPassword("wrong-password"))
}

func TestCheckPasswordSupportsLegacyScrypt(t *testing.T) {
	legacyHash := "scrypt:00112233445566778899aabbccddeeff:bd9b734f3d858cc223e87f6ca02147b2956aa5b7d839490ac24ef02e8422bebf0086b0e99fd084e2682cc7b7f92f168a57a71574feab1749ab709f91f2d4ee97"

	require.True(t, checkPassword("old-password", legacyHash))
	require.False(t, checkPassword("wrong-password", legacyHash))
}

func TestCheckPasswordRejectsMalformedLegacyScrypt(t *testing.T) {
	require.False(t, checkPassword("old-password", "scrypt:00112233445566778899aabbccddeeff"))
	require.False(t, checkPassword("old-password", "scrypt:00112233445566778899aabbccddeeff:not-hex"))
}
