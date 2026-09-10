//go:build !js || !wasm || extra2

package domain

import "errors"

const (
	TehonbikiMinChips     = 100
	TehonbikiMaxChips     = 100000
	TehonbikiDefaultChips = 1000
	TehonbikiMinBet       = 10
	TehonbikiMaxBet       = 500
	TehonbikiDefaultBet   = 50
)

// TehonbikiBetType identifies one house-table wager.
type TehonbikiBetType string

const (
	TehonbikiBetSingle TehonbikiBetType = "single"
	TehonbikiBetDouble TehonbikiBetType = "double"
	TehonbikiBetTriple TehonbikiBetType = "triple"
	TehonbikiBetHalf   TehonbikiBetType = "half"
)

// TehonbikiPayout contains the adopted house-table profit ratio.
type TehonbikiPayout struct{ MultiplierNum, MultiplierDen int }

var TehonbikiPayouts = map[TehonbikiBetType]TehonbikiPayout{TehonbikiBetSingle: {9, 2}, TehonbikiBetDouble: {9, 5}, TehonbikiBetTriple: {9, 10}, TehonbikiBetHalf: {9, 10}}

// TehonbikiConfig is the table configuration.
type TehonbikiConfig struct {
	InitialChips int
	DefaultBet   int
}

func DefaultTehonbikiConfig() TehonbikiConfig {
	return TehonbikiConfig{TehonbikiDefaultChips, TehonbikiDefaultBet}
}
func (c TehonbikiConfig) Validate() error {
	if c.InitialChips < TehonbikiMinChips || c.InitialChips > TehonbikiMaxChips {
		return errors.New("tehonbiki: initial chips out of range")
	}
	if c.DefaultBet < TehonbikiMinBet || c.DefaultBet > TehonbikiMaxBet {
		return errors.New("tehonbiki: default bet out of range")
	}
	return nil
}
