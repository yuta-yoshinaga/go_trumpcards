//go:build !js || !wasm || extra2

package domain

// TehonbikiPlayer is the child player's chip balance.
type TehonbikiPlayer struct{ chips ChipHolder }

func NewTehonbikiPlayer(n int) *TehonbikiPlayer     { p := &TehonbikiPlayer{}; p.SetChips(n); return p }
func (p *TehonbikiPlayer) GetChips() int            { return p.chips.GetChips() }
func (p *TehonbikiPlayer) SetChips(n int)           { p.chips.SetChips(n) }
func (p *TehonbikiPlayer) AddChips(n int)           { p.chips.AddChips(n) }
func (p *TehonbikiPlayer) SubtractChips(n int) bool { return p.chips.SubtractChips(n) }
