package filter

import (
	"errors"
	"fmt"
	"strings"
)

func ParseFlag(input string) (Operation, error) {
	var (
		field Field
		cmp   Comparator
		value string
	)

	// Trim the string before splitting after comma.
	s := strings.TrimSpace(input)
	filterContents := strings.Split(s, ",")

	// If it does not contain 3 elements, then the provided input is not following the required format.
	if len(filterContents) != 3 {
		return nil, errors.New("filters must be provided in the following format: --filters field=<field_type>,comparator=<comparator_type>,value=<string>")
	}

	// Iterate over each key-value pair
	for _, fc := range filterContents {

		// Split by =
		keyValues := strings.Split(fc, "=")

		// If it does not contain exactly 2 elements, then the provided input is not following the required format.
		if len(keyValues) != 2 {
			return nil, errors.New("filters must be provided in the following format: --filters field=<field_type>,comparator=<comparator_type>,value=<string>")
		}

		// Check if the key-value pairs can be parsed into intelligible operations.
		switch keyValues[0] {
		case "field":
			if keyValues[1] != string(Address) &&
				keyValues[1] != string(Balance) &&
				keyValues[1] != string(Nonce) &&
				keyValues[1] != string(Pair) &&
				keyValues[1] != string(Token) {
				return nil, fmt.Errorf("field [%s] not supported as a filtering option", keyValues[1])
			}

			field = Field(keyValues[1])

		case "comparator":
			if keyValues[1] != string(Equal) &&
				keyValues[1] != string(NotEqual) &&
				keyValues[1] != string(GreaterThan) &&
				keyValues[1] != string(LessThan) &&
				keyValues[1] != string(GreaterOrEqualThan) &&
				keyValues[1] != string(LessOrEqualThan) {
				return nil, fmt.Errorf("comparator [%s] not supported as a filtering option", keyValues[1])
			}

			cmp = Comparator(keyValues[1])

		case "value":
			if keyValues[1] == "" {
				return nil, fmt.Errorf("value [%s] not supported as a filtering option", keyValues[1])
			}

			value = keyValues[1]

		default:
			return nil, errors.New("filters must be provided in the following format: --filters field=<field_type>,comparator=<comparator_type>,value=<string>")
		}
	}

	// Create a filter operation from the parsed input.
	fo, err := newFilterOperation(field, cmp, value)
	if err != nil {
		return nil, err
	}

	return fo, nil
}
