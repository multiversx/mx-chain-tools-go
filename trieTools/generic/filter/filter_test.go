package filter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddressApplyFilter(t *testing.T) {
	t.Run("address filter eq", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Address, Equal, "Thisisa32characterlongstring1234")
		require.NoError(t, err)

		accDetails := &AccountDetails{Address: "Thisisa32characterlongstring1234"}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("address filter ne", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Address, NotEqual, "Thisisa32characterlongstring1234")
		require.NoError(t, err)

		accDetails := &AccountDetails{Address: "Thisisa32characterlongstring1233"}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})
}

func TestBalanceApplyFilter(t *testing.T) {
	t.Run("balance filter eq", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, Equal, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("3", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter ne", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, NotEqual, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("4", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter gt", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, GreaterThan, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("4", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter lt", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, LessThan, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("2", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter ge equal", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, GreaterOrEqualThan, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("3", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter ge greater", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, GreaterOrEqualThan, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("4", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter le equal", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, LessOrEqualThan, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("3", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter le less", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Balance, LessOrEqualThan, "3")
		require.NoError(t, err)

		n := new(big.Int)
		n, _ = n.SetString("2", 10)
		accDetails := &AccountDetails{Balance: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})
}

func TestNonceApplyFilter(t *testing.T) {
	t.Run("nonce filter eq", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, Equal, "3")
		require.NoError(t, err)

		n := uint64(3)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("nonce filter ne", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, NotEqual, "3")
		require.NoError(t, err)

		n := uint64(4)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("nonce filter gt", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, GreaterThan, "3")
		require.NoError(t, err)

		n := uint64(4)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("nonce filter lt", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, LessThan, "3")
		require.NoError(t, err)

		n := uint64(2)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("nonce filter ge equal", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, GreaterOrEqualThan, "3")
		require.NoError(t, err)

		n := uint64(3)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter ge greater", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, GreaterOrEqualThan, "3")
		require.NoError(t, err)

		n := uint64(4)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter le equal", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, LessOrEqualThan, "3")
		require.NoError(t, err)

		n := uint64(3)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("balance filter le less", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Nonce, LessOrEqualThan, "3")
		require.NoError(t, err)

		n := uint64(2)
		accDetails := &AccountDetails{Nonce: n}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})
}

func TestPairApplyFilter(t *testing.T) {
	t.Run("pair filter eq", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Pair, Equal, "abcd:Thisisa32characterlongstring1234")
		require.NoError(t, err)

		accDetails := &AccountDetails{Pairs: map[string]string{"abcd": "Thisisa32characterlongstring1234"}}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("pair filter ne", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Pair, NotEqual, "abcd:Thisisa32characterlongstring5678")
		require.NoError(t, err)

		accDetails := &AccountDetails{Pairs: map[string]string{"abcd": "Thisisa32characterlongstring1234"}}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})
}

func TestTokenApplyFilter(t *testing.T) {
	t.Run("pair filter eq", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Token, Equal, "some_token")
		require.NoError(t, err)

		accDetails := &AccountDetails{Tokens: map[string]struct{}{"some_token": {}}}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})

	t.Run("pair filter ne", func(t *testing.T) {
		filterOperation, err := newFilterOperation(Token, NotEqual, "some_token")
		require.NoError(t, err)

		accDetails := &AccountDetails{Tokens: map[string]struct{}{"some_other_token": {}}}
		result := filterOperation.ApplyFilter(accDetails)
		require.Equal(t, true, result)
	})
}
