package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/crypto/ethsecp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestIsSupportedKeys(t *testing.T) {
	testCases := []struct {
		name        string
		pk          cryptotypes.PubKey
		isSupported bool
	}{
		{
			"nil key",
			nil,
			false,
		},
		{
			"ethsecp256k1 key",
			&ethsecp256k1.PubKey{},
			true,
		},
		{
			"ed25519 key",
			&ed25519.PubKey{},
			true,
		},
		{
			"multisig key - no pubkeys",
			&multisig.LegacyAminoPubKey{},
			false,
		},
		{
			"multisig key - valid pubkeys",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &ed25519.PubKey{}}),
			true,
		},
		{
			"multisig key - nested multisig",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &multisig.LegacyAminoPubKey{}}),
			false,
		},
		{
			"multisig key - invalid pubkey",
			multisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{&ed25519.PubKey{}, &ed25519.PubKey{}, &secp256k1.PubKey{}}),
			false,
		},
		{
			"cosmos secp256k1",
			&secp256k1.PubKey{},
			false,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.isSupported, IsSupportedKey(tc.pk), tc.name)
	}
}

func TestGetAccAddressFromBech32(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("cosmos", "cosmospub")

	testCases := []struct {
		name       string
		address    string
		expAddress string
		expError   bool
	}{
		{
			"blank bech32 address",
			" ",
			"",
			true,
		},
		{
			"invalid bech32 address",
			"evmos",
			"",
			true,
		},
		{
			"invalid address bytes",
			"cosmos1123",
			"",
			true,
		},
		{
			"evmos address",
			"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
			"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
			false,
		},
		{
			"cosmos address",
			"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
			"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
			false,
		},
		{
			"osmosis address",
			"osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2",
			"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
			false,
		},
	}

	for _, tc := range testCases {
		addr, err := GetAccAddressFromBech32(tc.address)
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expAddress, addr.String(), tc.name)
		}
	}
}

func TestEvmosCoinDenom(t *testing.T) {
	testCases := []struct {
		name     string
		denom    string
		expError bool
	}{
		{
			"valid denom - native coin",
			"aatom",
			false,
		},
		{
			"valid denom - ibc coin",
			"ibc/7B2A4F6E798182988D77B6B884919AF617A73503FDAC27C916CD7A69A69013CF",
			false,
		},
		{
			"valid denom - ethereum address (ERC-20 contract)",
			"erc20:0x52908400098527886e0f7030069857D2E4169EE7",
			false,
		},
		{
			"invalid denom - only one character",
			"a",
			true,
		},
		{
			"invalid denom - too large (> 127 chars)",
			"ibc/7B2A4F6E798182988D77B6B884919AF617A73503FDAC27C916CD7A69A69013CF7B2A4F6E798182988D77B6B884919AF617A73503FDAC27C916CD7A69A69013CF",
			true,
		},
		{
			"invalid denom - starts with 0 but not followed by 'x'",
			"0a52908400098527886E0F7030069857D2E4169EE7",
			true,
		},
		{
			"invalid denom - hex address but 19 bytes long",
			"0x52908400098527886E0F7030069857D2E4169E",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			err := sdk.ValidateDenom(tc.denom)
			if tc.expError {
				require.Error(t, err, tc.name)
			} else {
				require.NoError(t, err, tc.name)
			}
		})
	}
}

func TestAccAddressFromBech32(t *testing.T) {
	testCases := []struct {
		address      string
		bech32Prefix string
		expErr       bool
		errContains  string
	}{
		{
			"",
			"",
			true,
			"empty address string is not allowed",
		},
		{
			"cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5",
			"stride",
			true,
			"invalid Bech32 prefix; expected stride, got cosmos",
		},
		{
			"cosmos1xv9tklw7d82sezh9haa573wufgy59vmw5",
			"cosmos",
			true,
			"decoding bech32 failed: invalid checksum",
		},
		{
			"stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
			"stride",
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc //nolint:copyloopvar // Needed to work correctly with concurrent tests

		t.Run(tc.address, func(t *testing.T) {
			t.Parallel()

			_, err := CreateAccAddressFromBech32(tc.address, tc.bech32Prefix)
			if tc.expErr {
				require.Error(t, err, "expected error while creating AccAddress")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			} else {
				require.NoError(t, err, "expected no error while creating AccAddress")
			}
		})
	}
}

func TestAddressConversion(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("cosmos", "cosmospub")

	hex := "0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E"
	bech32 := "cosmos10jmp6sgh4cc6zt3e8gw05wavvejgr5pwsjskvv"

	require.Equal(t, bech32, Bech32StringFromHexAddress(hex))
	gotAddr, err := HexAddressFromBech32String(bech32)
	require.NoError(t, err)
	require.Equal(t, hex, gotAddr.Hex())
}

func TestGetIBCDenomAddress(t *testing.T) {
	testCases := []struct {
		name        string
		denom       string
		expErr      bool
		expectedRes string
	}{
		{
			"",
			"test",
			true,
			"does not have 'ibc/' prefix",
		},
		{
			"",
			"ibc/",
			true,
			"is not a valid IBC voucher hash",
		},
		{
			"",
			"ibc/qqqqaaaaaa",
			true,
			"invalid denomination for cross-chain transfer",
		},
		{
			"",
			"ibc/DF63978F803A2E27CA5CC9B7631654CCF0BBC788B3B7F0A10200508E37C70992",
			false,
			"0x631654CCF0BBC788b3b7F0a10200508e37c70992",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			address, err := GetIBCDenomAddress(tc.denom)
			if tc.expErr {
				require.Error(t, err, "expected error while get ibc denom address")
				require.Contains(t, err.Error(), tc.expectedRes, "expected different error")
			} else {
				require.NoError(t, err, "expected no error while get ibc denom address")
				require.Equal(t, address.Hex(), tc.expectedRes)
			}
		})
	}
}

// TestBytes32ToString tests the Bytes32ToString helper function
func TestBytes32ToString(t *testing.T) {
	testCases := []struct {
		name     string
		input    [32]byte
		expected string
	}{
		{
			name:     "Full string - no null bytes",
			input:    [32]byte{'M', 'a', 'k', 'e', 'r', ' ', 'T', 'o', 'k', 'e', 'n'},
			expected: "Maker Token",
		},
		{
			name:     "Short string - with null bytes",
			input:    [32]byte{'M', 'K', 'R'},
			expected: "MKR",
		},
		{
			name:     "Empty string",
			input:    [32]byte{},
			expected: "",
		},
		{
			name:     "Single character",
			input:    [32]byte{'A'},
			expected: "A",
		},
		{
			name:     "String with special characters",
			input:    [32]byte{'T', 'e', 's', 't', '-', '1', '2', '3'},
			expected: "Test-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Bytes32ToString(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
