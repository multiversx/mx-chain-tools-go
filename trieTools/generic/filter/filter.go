package filter

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Field string
type Comparator string

const (
	Address Field = "address"
	Balance Field = "balance"
	Nonce   Field = "nonce"
	Pair    Field = "pair"
	Token   Field = "token"

	Equal              Comparator = "eq"
	NotEqual           Comparator = "ne"
	GreaterThan        Comparator = "gt"
	LessThan           Comparator = "lt"
	GreaterOrEqualThan Comparator = "ge"
	LessOrEqualThan    Comparator = "le"
)

type AccountDetails struct {
	Nonce   uint64              `json:"nonce"`
	Address string              `json:"address"`
	Balance *big.Int            `json:"balance"`
	Pairs   map[string]string   `json:"pairs"`
	Tokens  map[string]struct{} `json:"tokens"`
}

type Operation interface {
	ApplyFilter(accDetails *AccountDetails) bool
}

type operation struct {
	field      Field
	comparator Comparator
	value      string
}

type addressOperation struct {
	operation
}

func (a *addressOperation) ApplyFilter(accDetails *AccountDetails) bool {
	return evaluateStringExpression(accDetails.Address, a.comparator, a.value)
}

type balanceOperation struct {
	operation
}

func (b *balanceOperation) ApplyFilter(accDetails *AccountDetails) bool {
	n := new(big.Int)
	n, _ = n.SetString(b.value, 10)
	return evaluateBigIntExpression(accDetails.Balance, b.comparator, n)
}

type nonceOperation struct {
	operation
}

func (n *nonceOperation) ApplyFilter(accDetails *AccountDetails) bool {
	v, _ := strconv.ParseUint(n.value, 10, 64)
	return evaluateUnsignedExpression(accDetails.Nonce, n.comparator, v)
}

type pairOperation struct {
	operation
}

func (p *pairOperation) ApplyFilter(accDetails *AccountDetails) bool {
	store := strings.Split(p.value, ":")
	key := store[0]
	value := store[1]

	return evaluateStringExpression(accDetails.Pairs[key], p.comparator, value)
}

type tokenOperation struct {
	operation
}

func (t *tokenOperation) ApplyFilter(accDetails *AccountDetails) bool {
	for k, _ := range accDetails.Tokens {
		if evaluateStringExpression(k, t.comparator, t.value) {
			return true
		}
	}

	return false
}

func newFilterOperation(field Field, cmp Comparator, value string) (Operation, error) {
	if field == "" {
		return nil, errors.New("field_type cannot be empty")
	}

	if cmp == "" {
		return nil, errors.New("comparator_type cannot be empty")
	}

	if value == "" {
		return nil, errors.New("value_type cannot be empty")
	}

	switch field {
	case Address:
		if cmp != Equal &&
			cmp != NotEqual {
			return nil, fmt.Errorf("comparator [%s] is not supported as a filtering option for field address", cmp)
		}

		if len(value) != 32 {
			return nil, fmt.Errorf("value [%v] is not supported as a filering option for field address: value is not 32 characters long", value)
		}

		op := operation{field, cmp, value}

		return &addressOperation{op}, nil

	case Balance:
		n := new(big.Int)
		_, ok := n.SetString(value, 10)
		if !ok {
			return nil, fmt.Errorf("value [%v] is not supported as a filering option for field balance", value)
		}

		op := operation{field, cmp, value}

		return &balanceOperation{op}, nil

	case Nonce:
		_, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("value [%v] is not supported as a filering option for field balance: %v", value, err)
		}
		op := operation{field, cmp, value}

		return &nonceOperation{op}, nil

	case Pair:
		if cmp != Equal &&
			cmp != NotEqual {
			return nil, fmt.Errorf("comparator [%s] is not supported as a filtering option for field pair", cmp)
		}

		if len(strings.Split(value, ":")) != 2 {
			return nil, fmt.Errorf("value [%v] is not supported as a filtering option for field parid", value)
		}

		op := operation{field, cmp, value}

		return &pairOperation{op}, nil

	case Token:
		if cmp != Equal &&
			cmp != NotEqual {
			return nil, fmt.Errorf("comparator [%s] is not supported as a filtering option for field token", cmp)
		}

		op := operation{field, cmp, value}

		return &tokenOperation{op}, nil

	default:
		return nil, fmt.Errorf("field [%s] is not supported as filtering option", field)

	}
}

func evaluateStringExpression(address string, cmp Comparator, value string) bool {
	switch cmp {
	case Equal:
		return address == value

	case NotEqual:
		return address != value

	default:
		return false
	}
}

func evaluateBigIntExpression(balance *big.Int, operator Comparator, value *big.Int) bool {
	switch operator {
	case Equal:
		return balance.Cmp(value) == 0

	case NotEqual:
		return balance.Cmp(value) != 0

	case GreaterThan:
		return balance.Cmp(value) == 1

	case LessThan:
		return balance.Cmp(value) == -1

	case GreaterOrEqualThan:
		return balance.Cmp(value) == 1 || balance.Cmp(value) == 0

	case LessOrEqualThan:
		return balance.Cmp(value) == -1 || balance.Cmp(value) == 0

	default:
		return false
	}
}

func evaluateUnsignedExpression(nonce uint64, operator Comparator, value uint64) bool {
	switch operator {
	case Equal:
		return nonce == value

	case NotEqual:
		return nonce != value

	case GreaterThan:
		return nonce > value

	case LessThan:
		return nonce < value

	case GreaterOrEqualThan:
		return nonce >= value

	case LessOrEqualThan:
		return nonce <= value

	default:
		return false
	}
}
