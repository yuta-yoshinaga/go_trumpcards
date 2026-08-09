//go:build !js || !wasm || extra

package domain

import "math/rand"

// CinchCPU.go は CPU の意思決定 (ビッド / 切り札宣言 / トリックプレイ) を担う。
// Easy 難易度は乱数を使う (テストは Easy を避けるか retry ループで両分岐を網羅する)。
// Normal / Hard は決定的なヒューリスティックでテストの再現性を担保する。

// --- ビッド ---

// cpuSelectBid は CPU がビッドを選択する。
func (g *Cinch) cpuSelectBid(playerIdx int) int {
	switch g.config.CpuDifficulty {
	case CinchDifficultyHard:
		return g.bidWithRules(playerIdx, g.estimateBid(playerIdx, true))
	case CinchDifficultyNormal:
		return g.bidWithRules(playerIdx, g.estimateBid(playerIdx, false))
	default:
		return g.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy はランダムに pass / 小さめのビッドを選ぶ。
func (g *Cinch) cpuBidEasy(playerIdx int) int {
	choices := []int{CinchPassBid, CinchMinBid, CinchMinBid + 1, CinchMinBid + 2}
	bid := choices[rand.Intn(len(choices))]
	if bid > 0 && bid <= g.currentBid {
		bid = CinchPassBid
	}
	return g.coercePassWhenForbidden(playerIdx, bid)
}

// estimateBid は手札から獲得可能ポイントの見積り (=最大ビッド) を返す。
// 各スートを切り札と仮定したときの期待ポイントを評価し、最大値を採用する。
// strict=true では控えめに評価する。
func (g *Cinch) estimateBid(playerIdx int, strict bool) int {
	best := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		pts := g.estimateSuitPoints(playerIdx, suit)
		if pts > best {
			best = pts
		}
	}
	if strict {
		best -= 2
	}
	if best < 0 {
		best = 0
	}
	if best > CinchMaxBid {
		best = CinchMaxBid
	}
	return best
}

// estimateSuitPoints は指定スートを切り札と仮定したときの期待獲得ポイント (粗い見積り)。
func (g *Cinch) estimateSuitPoints(playerIdx, suit int) int {
	p := g.players[playerIdx]
	trumpCount := 0
	hasAce, hasKing, hasTen, hasJack := false, false, false, false
	hasRightPedro, hasLeftPedro := false, false
	sameColor := cinchSameColorSuit(suit)
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if cinchIsTrump(c, suit) {
			trumpCount++
		}
		if c.GetDesign() == suit {
			switch c.GetValue() {
			case 1:
				hasAce = true
			case 13:
				hasKing = true
			case 10:
				hasTen = true
			case 11:
				hasJack = true
			case 5:
				hasRightPedro = true
			}
		}
		if c.GetValue() == 5 && c.GetDesign() == sameColor {
			hasLeftPedro = true
		}
	}
	pts := 0
	// A (High) はほぼ確実。
	if hasAce {
		pts++
	}
	// K は A/長さがあれば取りやすい。
	if hasKing && (hasAce || trumpCount >= 3) {
		pts++
	}
	// 10 (Game) と J は長い手ほど取りやすい。
	if hasTen && trumpCount >= 3 {
		pts++
	}
	if hasJack && trumpCount >= 3 {
		pts++
	}
	// Pedro は 5 点。長い切り札があれば守り切れる (自分で取る/庇う) 見込み。
	if hasRightPedro && trumpCount >= 4 {
		pts += 5
	}
	if hasLeftPedro && trumpCount >= 4 {
		pts += 5
	}
	return pts
}

// bidWithRules は見積りにビッドルール (>currentBid) を適用する。
func (g *Cinch) bidWithRules(playerIdx, suggested int) int {
	if suggested < CinchMinBid || suggested <= g.currentBid {
		return g.coercePassWhenForbidden(playerIdx, CinchPassBid)
	}
	if suggested > CinchMaxBid {
		suggested = CinchMaxBid
	}
	return suggested
}

// coercePassWhenForbidden は親で全員パスの場合に強制 stuck (CinchMinBid) を返す。
func (g *Cinch) coercePassWhenForbidden(playerIdx, bid int) int {
	if bid == CinchPassBid && playerIdx == g.dealerIdx && g.currentBid == 0 &&
		g.bidsCompleted() == CinchPlayerCnt-1 {
		return CinchMinBid
	}
	return bid
}

// --- 切り札宣言 ---

// cpuSelectTrump は CPU が宣言する切り札スートを返す (最も期待ポイントの高いスート)。
func (g *Cinch) cpuSelectTrump(playerIdx int) int {
	best := CardDesignSpade
	bestPts := -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		pts := g.estimateSuitPoints(playerIdx, suit)
		if pts > bestPts {
			bestPts = pts
			best = suit
		}
	}
	return best
}

// --- トリックプレイ ---

// cpuSelectPlayCard は CPU がプレイするカードのインデックスを返す。
func (g *Cinch) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == CinchDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart は Normal/Hard 共通の決定的プレイ。取りに行くべきトリックは勝てる最小の
// カードで、そうでなければ最も安全な (ポイントを与えない) カードを捨てる。
func (g *Cinch) cpuPlaySmart(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	wantWin := g.cpuWantsWin(playerIdx)
	if len(g.currentTrick) == 0 {
		if wantWin {
			return g.pickStrongLead(p, valid)
		}
		return g.pickSafeLead(p, valid)
	}
	winning := g.currentWinningCard()
	leadSuit := g.currentTrick[0].Card.GetDesign()
	trickHasPoints := g.trickCarriesPoints()
	if wantWin || trickHasPoints {
		// 勝てる最小のカードを選ぶ。
		bestIdx, bestRank := -1, 1<<30
		for _, idx := range valid {
			c := p.GetCard(idx)
			if g.cardBeats(c, winning, leadSuit) {
				r := g.cinchStrength(c)
				if r < bestRank {
					bestRank, bestIdx = r, idx
				}
			}
		}
		if bestIdx >= 0 {
			return bestIdx
		}
	}
	// 勝てない/取りたくない: ポイントの無い最弱カードを捨てる。
	return g.pickSafeDiscard(p, valid)
}

// cpuWantsWin はこのプレイヤーがトリックを取りに行くべきか。
func (g *Cinch) cpuWantsWin(playerIdx int) bool {
	if playerIdx == g.bidWinnerIdx {
		return true // ビッダーは常に取りに行く
	}
	// non-bidder はビッダーをセットしたいのでポイント絡みは取りに行く。
	return false
}

// trickCarriesPoints は現在進行中のトリックにポイントカードが含まれるか。
func (g *Cinch) trickCarriesPoints() bool {
	for _, tc := range g.currentTrick {
		if cinchPointValue(tc.Card, g.trumpSuit) > 0 {
			return true
		}
	}
	return false
}

// currentWinningCard は進行中トリックで現在勝っているカードを返す (空なら nil)。
func (g *Cinch) currentWinningCard() *Card {
	return currentTrickWinnerCard(g.currentTrick, g)
}

// cinchStrength はカードの絶対的な強さ (切り札は +100 でオフ切り札より必ず強い)。
func (g *Cinch) cinchStrength(c *Card) int {
	if cinchIsTrump(c, g.trumpSuit) {
		return 100 + cinchTrumpRank(c, g.trumpSuit)
	}
	return cinchRankValue(c.GetValue())
}

// pickStrongLead は取りに行くリード: 最強の切り札 (無ければ最強札)。
func (g *Cinch) pickStrongLead(p *CinchPlayer, valid []int) int {
	best, bestVal := valid[0], -1
	for _, idx := range valid {
		if v := g.cinchStrength(p.GetCard(idx)); v > bestVal {
			best, bestVal = idx, v
		}
	}
	return best
}

// pickSafeLead はポイントを与えたくないリード: ポイントの無い最弱カード。
func (g *Cinch) pickSafeLead(p *CinchPlayer, valid []int) int {
	return g.pickSafeDiscard(p, valid)
}

// pickSafeDiscard はポイントを含まない最弱カードを選ぶ。全てポイント札なら最弱ポイント札。
func (g *Cinch) pickSafeDiscard(p *CinchPlayer, valid []int) int {
	best, bestScore := -1, 1<<30
	for _, idx := range valid {
		c := p.GetCard(idx)
		score := g.cinchStrength(c)
		// ポイント札は温存したい (捨てたくない) ので大きく加点。
		score += cinchPointValue(c, g.trumpSuit) * 50
		if score < bestScore {
			bestScore, best = score, idx
		}
	}
	if best < 0 {
		return valid[0]
	}
	return best
}

// --- Hint ---

// GetHint は人間プレイヤーの手番における推奨アクションを返す。
func (g *Cinch) GetHint() *CinchHint {
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 {
		return nil
	}
	switch g.phase {
	case CinchPhaseBid:
		if g.bidPlayerIdx != humanIdx {
			return nil
		}
		bid := g.bidWithRules(humanIdx, g.estimateBid(humanIdx, true))
		reason := "bid_pass"
		if bid >= CinchMinBid {
			reason = "bid_strong"
		}
		return &CinchHint{Bid: &bid, Reason: reason}
	case CinchPhaseNameTrump:
		if g.bidWinnerIdx != humanIdx {
			return nil
		}
		suit := g.cpuSelectTrump(humanIdx)
		return &CinchHint{TrumpSuit: &suit, Reason: "name_trump"}
	case CinchPhasePlay:
		if g.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := g.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(humanIdx, valid)
		return &CinchHint{CardIndices: []int{idx}, Reason: g.playHintReason(humanIdx, idx)}
	default:
		return nil
	}
}

// playHintReason はプレイヒントの理由キーを返す。
func (g *Cinch) playHintReason(playerIdx, chosenIdx int) string {
	p := g.players[playerIdx]
	card := p.GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		return "lead_strong"
	}
	if cinchIsTrump(card, g.trumpSuit) {
		return "trump_cut"
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	return "discard_low"
}
