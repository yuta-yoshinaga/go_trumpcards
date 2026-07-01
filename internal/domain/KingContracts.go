//go:build !js || !wasm || extra

package domain

// KingContracts.go はコントラクトごとの得点計算 (Strategy パターン) を担う。
// 6 つの負のコントラクトは獲得トリックの内容から、King (Trump) は取ったトリック数から、
// それぞれ 1 ディール分の得点 (プレイヤー別) を算出する。

// kingContractDef は 1 コントラクトのメタデータと得点関数。
type kingContractDef struct {
	id    int
	score func(g *King) map[int]int
}

// kingContractDefs は全 7 コントラクトの定義表 (id 昇順)。
var kingContractDefs = []kingContractDef{
	{id: KingContractNoTricks, score: kingScoreNoTricks},
	{id: KingContractNoHearts, score: kingScoreNoHearts},
	{id: KingContractNoQueens, score: kingScoreNoQueens},
	{id: KingContractKingHeart, score: kingScoreKingHeart},
	{id: KingContractNoLastTwo, score: kingScoreNoLastTwo},
	{id: KingContractNoMen, score: kingScoreNoMen},
	{id: KingContractKingTrump, score: kingScoreKingTrump},
}

// scoreDeal は現在のコントラクトの得点関数を実行し、結果内訳を返す。
func (g *King) scoreDeal() *KingDealDetail {
	gained := map[int]int{}
	for i := range g.players {
		gained[i] = 0
	}
	if def := kingContractDefFor(g.currentContract); def != nil {
		gained = def.score(g)
	}
	return &KingDealDetail{
		Contract:  g.currentContract,
		TrumpSuit: g.trumpSuit,
		DealerIdx: g.dealerIdx,
		Gained:    gained,
	}
}

// kingContractDefFor は id に対応するコントラクト定義を返す (なければ nil)。
func kingContractDefFor(id int) *kingContractDef {
	for i := range kingContractDefs {
		if kingContractDefs[i].id == id {
			return &kingContractDefs[i]
		}
	}
	return nil
}

// kingScoreNoTricks: 取ったトリック 1 つにつき減点。
func kingScoreNoTricks(g *King) map[int]int {
	res := map[int]int{}
	for i, p := range g.players {
		res[i] = -p.GetTrickCount() * KingNoTrickPenalty
	}
	return res
}

// kingScoreNoHearts: 取ったハート 1 枚につき減点。
func kingScoreNoHearts(g *King) map[int]int {
	res := map[int]int{}
	for i, p := range g.players {
		res[i] = -p.CapturedHearts() * KingHeartPenalty
	}
	return res
}

// kingScoreNoQueens: 取った Q 1 枚につき減点。
func kingScoreNoQueens(g *King) map[int]int {
	res := map[int]int{}
	for i, p := range g.players {
		res[i] = -p.CapturedQueens() * KingQueenPenalty
	}
	return res
}

// kingScoreKingHeart: K♥ を取ったプレイヤーが大幅減点。
func kingScoreKingHeart(g *King) map[int]int {
	res := map[int]int{}
	for i, p := range g.players {
		if p.HasKingOfHearts() {
			res[i] = -KingKingHeartPenalty
		} else {
			res[i] = 0
		}
	}
	return res
}

// kingScoreNoLastTwo: 最後の 2 トリック (12・13 番目) を取ったプレイヤーが減点。
// 各プレイヤーが取った「最後の 2 トリック」の本数に応じて減点する。
func kingScoreNoLastTwo(g *King) map[int]int {
	res := map[int]int{}
	for i := range g.players {
		res[i] = 0
	}
	for i, p := range g.players {
		lastTwo := 0
		for _, rank := range p.GetTrickRanks() {
			if rank >= KingHandSize-1 {
				lastTwo++
			}
		}
		res[i] = -lastTwo * KingLastTwoPenalty
	}
	return res
}

// kingScoreNoMen: 取った J / K 1 枚につき減点。
func kingScoreNoMen(g *King) map[int]int {
	res := map[int]int{}
	for i, p := range g.players {
		res[i] = -p.CapturedMen() * KingMenPenalty
	}
	return res
}

// kingScoreKingTrump: 取ったトリック 1 つにつき加点 (正のコントラクト)。
func kingScoreKingTrump(g *King) map[int]int {
	res := map[int]int{}
	for i, p := range g.players {
		res[i] = p.GetTrickCount() * KingTrumpReward
	}
	return res
}
