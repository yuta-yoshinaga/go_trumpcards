//go:build test

package domain

// TrappolaFindDeclarationsForTest は 1 席の手札から成立する役を返す (テスト用)。
func TrappolaFindDeclarationsForTest(playerIdx int, p *TrappolaPlayer) []TrappolaDeclaration {
	return trappolaFindDeclarations(playerIdx, p)
}

// GetDeckForTest は山札を返す (テスト用)。
func (g *Trappola) GetDeckForTest() *TrumpCards { return g.trumpCards }

// TrappolaStrengthForTest は札位の強さを返す (テスト用)。
func TrappolaStrengthForTest(value int) int { return trappolaStrength(value) }

// TrappolaThirdsForTest はカード点 (1/3 点単位) を返す (テスト用)。
func TrappolaThirdsForTest(value int) int { return trappolaThirds(value) }
