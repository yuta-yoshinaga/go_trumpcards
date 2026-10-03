//go:build !js || !wasm || extra6

package domain

// JassFamilyCardPoints returns a card's trick points in the Jass family
// (Belote, Coinche, Jass, Klaberjass, Tarabish).
// Trump suit: J=20, 9=14, A=11, 10=10, K=4, Q=3.
// Other suits: A=11, 10=10, K=4, Q=3, J=2.
// Every other card, and a nil card, scores 0.
func JassFamilyCardPoints(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11:
			return 20
		case 9:
			return 14
		case 1:
			return 11
		case 10:
			return 10
		case 13:
			return 4
		case 12:
			return 3
		}
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	}
	return 0
}
