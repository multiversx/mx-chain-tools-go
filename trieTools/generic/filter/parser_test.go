package filter

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Run("wrong input", func(t *testing.T) {
		got, err := ParseFlag("This is some random string")
		nErr := errors.New("filters must be provided in the following format: --filters field=<field_type>,comparator=<comparator_type>,value=<string>")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("wrong input good format", func(t *testing.T) {
		got, err := ParseFlag("a=b,c=d,e=f")
		nErr := errors.New("filters must be provided in the following format: --filters field=<field_type>,comparator=<comparator_type>,value=<string>")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("empty field", func(t *testing.T) {
		got, err := ParseFlag("field=,comparator=eq,value=abcd")
		nErr := errors.New("field [] not supported as a filtering option")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("empty comparator", func(t *testing.T) {
		got, err := ParseFlag("field=address,comparator=,value=abcd")
		nErr := errors.New("comparator [] not supported as a filtering option")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("empty value", func(t *testing.T) {
		got, err := ParseFlag("field=address,comparator=eq,value=")
		nErr := errors.New("value [] not supported as a filtering option")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("wrong separator key-value", func(t *testing.T) {
		got, err := ParseFlag("field-balance,comparator-eq,field-3")
		nErr := errors.New("filters must be provided in the following format: --filters field=<field_type>,comparator=<comparator_type>,value=<string>")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("not enough types provided", func(t *testing.T) {
		got, err := ParseFlag("comparator=eq")
		nErr := errors.New("filters must be provided in the following format: --filters field=<field_type>,comparator=<comparator_type>,value=<string>")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("no field types provided", func(t *testing.T) {
		got, err := ParseFlag("comparator=eq,comparator=eq,comparator=eq")
		nErr := errors.New("field_type cannot be empty")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("only field types provided", func(t *testing.T) {
		got, err := ParseFlag("field=address,field=address,field=address")
		nErr := errors.New("comparator_type cannot be empty")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("no value type provided", func(t *testing.T) {
		got, err := ParseFlag("field=address,comparator=eq,field=address")
		nErr := errors.New("value_type cannot be empty")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("no comparator type provided", func(t *testing.T) {
		got, err := ParseFlag("field=address,value=eq,field=address")
		nErr := errors.New("comparator_type cannot be empty")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})
}

func TestAddressFieldParser(t *testing.T) {
	t.Run("eq comparator", func(t *testing.T) {
		got, err := ParseFlag("field=address,comparator=eq,value=Thisisa32characterlongstring1234")
		require.NoError(t, err)
		expected := &addressOperation{operation{field: Address, comparator: Equal, value: "Thisisa32characterlongstring1234"}}
		require.Equal(t, got, expected)
	})

	t.Run("ne comparator", func(t *testing.T) {
		got, err := ParseFlag("field=address,comparator=ne,value=Thisisa32characterlongstring1234")
		require.NoError(t, err)
		expected := &addressOperation{operation{field: Address, comparator: NotEqual, value: "Thisisa32characterlongstring1234"}}
		require.Equal(t, got, expected)
	})

	t.Run("gt comparator", func(t *testing.T) {
		got, err := ParseFlag("field=address,comparator=gt,value=Thisisa32characterlongstring1234")
		nErr := fmt.Errorf("comparator [%s] is not supported as a filtering option for field address", GreaterThan)
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})

	t.Run("wrong address format", func(t *testing.T) {
		got, err := ParseFlag("field=address,comparator=ne,value=x")
		nErr := fmt.Errorf("value [%v] is not supported as a filering option for field address: value is not 32 characters long", "x")
		require.Equal(t, nErr, err)
		require.Equal(t, got, nil)
	})
}

func TestBalanceFieldParser(t *testing.T) {
	t.Run("no errors eq comparator", func(t *testing.T) {
		got, err := ParseFlag("field=balance,comparator=eq,value=3")
		require.NoError(t, err)
		expected := &balanceOperation{operation{field: Balance, comparator: Equal, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors ne comparator", func(t *testing.T) {
		got, err := ParseFlag("field=balance,comparator=ne,value=3")
		require.NoError(t, err)
		expected := &balanceOperation{operation{field: Balance, comparator: NotEqual, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors gt comparator", func(t *testing.T) {
		got, err := ParseFlag("field=balance,comparator=gt,value=3")
		require.NoError(t, err)
		expected := &balanceOperation{operation{field: Balance, comparator: GreaterThan, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors lt comparator", func(t *testing.T) {
		got, err := ParseFlag("field=balance,comparator=lt,value=3")
		require.NoError(t, err)
		expected := &balanceOperation{operation{field: Balance, comparator: LessThan, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors ge comparator", func(t *testing.T) {
		got, err := ParseFlag("field=balance,comparator=ge,value=3")
		require.NoError(t, err)
		expected := &balanceOperation{operation{field: Balance, comparator: GreaterOrEqualThan, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors le comparator", func(t *testing.T) {
		got, err := ParseFlag("field=balance,comparator=le,value=3")
		require.NoError(t, err)
		expected := &balanceOperation{operation{field: Balance, comparator: LessOrEqualThan, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("wrong value type provided", func(t *testing.T) {
		got, err := ParseFlag("field=balance,comparator=gt,value=abcd")
		nErr := errors.New("value [abcd] is not supported as a filering option for field balance")
		require.Equal(t, err, nErr)
		require.Equal(t, got, nil)
	})
}

func TestNonceFieldParser(t *testing.T) {
	t.Run("no errors eq comparator", func(t *testing.T) {
		got, err := ParseFlag("field=nonce,comparator=eq,value=3")
		require.NoError(t, err)
		expected := &nonceOperation{operation{field: Nonce, comparator: Equal, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors ne comparator", func(t *testing.T) {
		got, err := ParseFlag("field=nonce,comparator=ne,value=3")
		require.NoError(t, err)
		expected := &nonceOperation{operation{field: Nonce, comparator: NotEqual, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors gt comparator", func(t *testing.T) {
		got, err := ParseFlag("field=nonce,comparator=gt,value=3")
		require.NoError(t, err)
		expected := &nonceOperation{operation{field: Nonce, comparator: GreaterThan, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors lt comparator", func(t *testing.T) {
		got, err := ParseFlag("field=nonce,comparator=lt,value=3")
		require.NoError(t, err)
		expected := &nonceOperation{operation{field: Nonce, comparator: LessThan, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors ge comparator", func(t *testing.T) {
		got, err := ParseFlag("field=nonce,comparator=ge,value=3")
		require.NoError(t, err)
		expected := &nonceOperation{operation{field: Nonce, comparator: GreaterOrEqualThan, value: "3"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors le comparator", func(t *testing.T) {
		got, err := ParseFlag("field=nonce,comparator=le,value=3")
		require.NoError(t, err)
		expected := &nonceOperation{operation{field: Nonce, comparator: LessOrEqualThan, value: "3"}}
		require.Equal(t, got, expected)
	})
}

func TestPairFieldParser(t *testing.T) {
	t.Run("no errors eq comparator", func(t *testing.T) {
		got, err := ParseFlag("field=pair,comparator=eq,value=abcd:Thisisa32characterlongstring1234")
		require.NoError(t, err)
		expected := &pairOperation{operation{field: Pair, comparator: Equal, value: "abcd:Thisisa32characterlongstring1234"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors ne comparator", func(t *testing.T) {
		got, err := ParseFlag("field=pair,comparator=ne,value=abcd:Thisisa32characterlongstring1234")
		require.NoError(t, err)
		expected := &pairOperation{operation{field: Pair, comparator: NotEqual, value: "abcd:Thisisa32characterlongstring1234"}}
		require.Equal(t, got, expected)
	})

	t.Run("ge comparator not valid", func(t *testing.T) {
		got, err := ParseFlag("field=pair,comparator=ge,value=abcd:abcd")
		require.Equal(t, err, fmt.Errorf("comparator [%s] is not supported as a filtering option for field pair", "ge"))
		require.Equal(t, got, nil)
	})
}

func TestTokenFieldParser(t *testing.T) {
	t.Run("no errors eq comparator", func(t *testing.T) {
		got, err := ParseFlag("field=token,comparator=eq,value=some_token")
		require.NoError(t, err)
		expected := &tokenOperation{operation{field: Token, comparator: Equal, value: "some_token"}}
		require.Equal(t, got, expected)
	})

	t.Run("no errors ne comparator", func(t *testing.T) {
		got, err := ParseFlag("field=token,comparator=ne,value=some_token")
		require.NoError(t, err)
		expected := &tokenOperation{operation{field: Token, comparator: NotEqual, value: "some_token"}}
		require.Equal(t, got, expected)
	})

	t.Run("ge comparator not valid", func(t *testing.T) {
		got, err := ParseFlag("field=token,comparator=ge,value=some_token")
		require.Equal(t, err, fmt.Errorf("comparator [%s] is not supported as a filtering option for field token", "ge"))
		require.Equal(t, got, nil)
	})
}
