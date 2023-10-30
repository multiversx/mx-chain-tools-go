package process

import (
	"fmt"
	"math/big"
)

func beautifyAmount(val string) string {
	bf, _ := big.NewFloat(0).SetString(val)

	den, _ := big.NewFloat(0).SetString("1000000000000000000")

	bf.Quo(bf, den)

	flt, _ := bf.Float64()
	res := fmt.Sprintf("%.2f EGLD", flt)

	return res
}
