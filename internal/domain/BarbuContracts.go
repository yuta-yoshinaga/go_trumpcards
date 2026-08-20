//go:build !js || !wasm || extra4

package domain

// BarbuContracts.go はコントラクトごとの得点計算 (Strategy パターン) を担う。
// 6 つのトリックコントラクトは獲得トリックの内容から、Dominoes は上がり順位
// から、それぞれ 1 ディール分の得点 (プレイヤー別) を算出する。

// barbuContractDef は 1 コントラクトのメタデータと得点関数。
type barbuContractDef struct {
	id    int
	score func(b *Barbu) map[int]int
}

// barbuContractDefs は全 7 コントラクトの定義表 (id 昇順)。
var barbuContractDefs = []barbuContractDef{
	{id: BarbuContractNoTricks, score: scoreNoTricks},
	{id: BarbuContractNoHearts, score: scoreNoHearts},
	{id: BarbuContractNoQueens, score: scoreNoQueens},
	{id: BarbuContractKingHeart, score: scoreKingHeart},
	{id: BarbuContractNoLastTrick, score: scoreNoLastTrick},
	{id: BarbuContractTrumps, score: scoreTrumps},
	{id: BarbuContractDominoes, score: scoreDominoes},
}

// scoreDeal は現在のコントラクトの得点関数を実行し、結果内訳を返す。
func (b *Barbu) scoreDeal() *BarbuDealDetail {
	gained := map[int]int{}
	for i := range b.players {
		gained[i] = 0
	}
	if def := barbuContractDefFor(b.currentContract); def != nil {
		gained = def.score(b)
	}
	return &BarbuDealDetail{
		Contract:  b.currentContract,
		TrumpSuit: b.trumpSuit,
		DealerIdx: b.dealerIdx,
		Gained:    gained,
	}
}

// barbuContractDefFor は id に対応するコントラクト定義を返す (なければ nil)。
func barbuContractDefFor(id int) *barbuContractDef {
	for i := range barbuContractDefs {
		if barbuContractDefs[i].id == id {
			return &barbuContractDefs[i]
		}
	}
	return nil
}

// scoreNoTricks: 取ったトリック 1 つにつき減点。
func scoreNoTricks(b *Barbu) map[int]int {
	g := map[int]int{}
	for i, p := range b.players {
		g[i] = -p.GetTrickCount() * BarbuNoTrickPenalty
	}
	return g
}

// scoreNoHearts: 取ったハート 1 枚につき減点。
func scoreNoHearts(b *Barbu) map[int]int {
	g := map[int]int{}
	for i, p := range b.players {
		g[i] = -p.CapturedHearts() * BarbuHeartPenalty
	}
	return g
}

// scoreNoQueens: 取った Q 1 枚につき減点。
func scoreNoQueens(b *Barbu) map[int]int {
	g := map[int]int{}
	for i, p := range b.players {
		g[i] = -p.CapturedQueens() * BarbuQueenPenalty
	}
	return g
}

// scoreKingHeart: K♥ を取ったプレイヤーが大幅減点。
func scoreKingHeart(b *Barbu) map[int]int {
	g := map[int]int{}
	for i, p := range b.players {
		if p.HasKingOfHearts() {
			g[i] = -BarbuKingHeartPenalty
		} else {
			g[i] = 0
		}
	}
	return g
}

// scoreNoLastTrick: 最後の (13 番目) トリックを取ったプレイヤーが減点。
func scoreNoLastTrick(b *Barbu) map[int]int {
	g := map[int]int{}
	for i := range b.players {
		g[i] = 0
	}
	if b.lastTrickWinner >= 0 && b.lastTrickWinner < len(b.players) {
		g[b.lastTrickWinner] = -BarbuLastTrickPenalty
	}
	return g
}

// scoreTrumps: 取ったトリック 1 つにつき加点 (正のコントラクト)。
func scoreTrumps(b *Barbu) map[int]int {
	g := map[int]int{}
	for i, p := range b.players {
		g[i] = p.GetTrickCount() * BarbuTrumpReward
	}
	return g
}

// scoreDominoes: 上がり順位に応じて加点 / 減点。
func scoreDominoes(b *Barbu) map[int]int {
	g := map[int]int{}
	for i, p := range b.players {
		rank := p.GetDominoRank()
		if rank >= 1 && rank <= BarbuPlayerCnt {
			g[i] = BarbuDominoScores[rank-1]
		} else {
			// 未上がり (手札が残ったまま) は最下位扱い。
			g[i] = BarbuDominoScores[BarbuPlayerCnt-1]
		}
	}
	return g
}
