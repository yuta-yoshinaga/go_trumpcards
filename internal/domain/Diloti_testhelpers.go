//go:build test

package domain

// SetTableForTest は場札と宣言を差し替える (テスト用)。
//
// **場は配りで決まるので、狙った盤面は組めない。** 捕獲規則や宣言の見え方を
// 確かめるにはここで固定するしかない。
func (d *Diloti) SetTableForTest(cards []*Card, decls []*DilotiDeclaration) {
	d.table = cards
	if decls == nil {
		decls = make([]*DilotiDeclaration, 0)
	}
	d.decls = decls
}

// SetGameEndRoundForTest は固定得点から終局する局末状態を作る。
func (d *Diloti) SetGameEndRoundForTest() {
	for _, p := range d.players {
		p.ResetRound()
	}
	d.players[0].AddScore(55)
	d.players[0].AddCaptured([]*Card{NewCard(CardDesignDiamond, 10, true)})
	d.table = make([]*Card, 0)
	d.decls = make([]*DilotiDeclaration, 0)
	d.finishRound()
}
