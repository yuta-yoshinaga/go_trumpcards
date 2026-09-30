package domain

// AceTenCardPoints returns a card's trick points in the ace-ten family
// (Schnapsen, Bezique, Pinochle, Marias, Tysiac, ...): A=11, 10=10, K=4,
// Q=3, J=2 and every other rank 0. A nil card scores 0.
func AceTenCardPoints(c *Card) int {
	if c == nil {
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
	default:
		return 0
	}
}
