//go:build test

package domain

// Basic-strategy solver used to generate the Spanish 21 table (#4705).
//
// **表を出典から写さない。**スパニッシュ21の戦略表は出典によってセルが食い違い
// (ハード12 vs 6 は「ヒット」と「スタンド」で割れる)、Wizard of Odds の表は画像で
// 機械可読でない。さらに市販の表はほぼ H17 用だが、このゲームは
// DefaultBlackJackConfig() が S17。**このゲームの規則そのものから解く。**
//
// このソルバの正しさは「標準デッキで走らせて、リポジトリが既に持っている標準の
// 基本戦略表を再現すること」で検証する (TestSolver_ReproducesStandardTable)。
// 再現できるなら、デッキ構成とバリアント設定だけ差し替えた出力も信用できる。
//
// モデル: 無限デッキ近似 (基本戦略表の標準的な計算法)。ディーラーはピークする
// (checkNaturalBlackJack が即終了させる) ので、プレイヤーの手番では
// **ディーラーがナチュラルでないことが確定している**。その条件付き分布で解く。

import (
	"fmt"
	"sort"
	"strings"
)

// solverRules は解く対象のゲーム規則。ドメインの設定から写す。
type solverRules struct {
	// probs[v] は BJ 値 v (1=A, 10=10/J/Q/K) を引く確率。
	probs [11]float64
	// dealerHitsSoft17 は config.DealerHitsSoft17。
	dealerHitsSoft17 bool
	// player21AlwaysWins / playerBJBeatsDealerBJ は variant のフラグ。
	player21AlwaysWins    bool
	playerBJBeatsDealerBJ bool
	// bonuses が真なら spanish21BonusEval 相当のボーナスを支払う。
	bonuses bool
	// surrender はレイトサレンダーが使えるか (SurrenderRule == BJSurrenderLate)。
	surrender bool
	// canDouble は最初の2枚でダブルダウンできるか。
	canDouble bool
}

// deckProbs はカード値の出現数からBJ値の確率を作る。
func deckProbs(counts map[int]int) [11]float64 {
	var out [11]float64
	total := 0
	for _, n := range counts {
		total += n
	}
	for v, n := range counts {
		bv := v
		if bv >= 10 {
			bv = 10
		}
		out[bv] += float64(n) / float64(total)
	}
	return out
}

// standardRules は標準ブラックジャック (52枚, S17, ボーナス無し)。ソルバ検証用。
func standardRules() solverRules {
	counts := map[int]int{}
	for v := 1; v <= 13; v++ {
		counts[v] = 4
	}
	return solverRules{
		probs:     deckProbs(counts),
		surrender: true,
		canDouble: true,
	}
}

// spanish21Rules は NewSpanish21BlackJack() が作るゲームの規則。
func spanish21Rules() solverRules {
	counts := map[int]int{}
	for _, v := range SpanishDeckValues {
		counts[v] = 4
	}
	cfg := DefaultBlackJackConfig()
	variant := Spanish21Variant()
	return solverRules{
		probs:                 deckProbs(counts),
		dealerHitsSoft17:      cfg.DealerHitsSoft17,
		player21AlwaysWins:    variant.Player21AlwaysWins,
		playerBJBeatsDealerBJ: variant.PlayerBJBeatsDealerBJ,
		bonuses:               variant.BonusEval != nil,
		surrender:             cfg.SurrenderRule == BJSurrenderLate,
		canDouble:             true,
	}
}

// dealerOutcome はディーラーの最終結果。bust は 22 で表す。
const dealerBust = 22

// dealerDist はアップカード u から、**ナチュラルでないことを条件に**
// ディーラーの最終スコア分布を返す (17..21, dealerBust)。
func (r solverRules) dealerDist(u int) map[int]float64 {
	// ホールカードを1枚引く。ピークにより、u+hole がナチュラルになる組は起こりえない。
	holes := map[int]float64{}
	var norm float64
	for c := 1; c <= 10; c++ {
		p := r.probs[c]
		if p == 0 {
			continue
		}
		if (u == 1 && c == 10) || (u == 10 && c == 1) {
			continue // ナチュラル: プレイヤーの手番は来ない
		}
		holes[c] = p
		norm += p
	}

	out := map[int]float64{}
	memo := map[[2]int]map[int]float64{}
	for c, p := range holes {
		total, soft := addCard(u, false, c)
		if u == 1 || c == 1 {
			soft = true
			total = u + c + 10
			if total > 21 {
				total -= 10
				soft = false
			}
		}
		for score, q := range r.dealerFrom(total, soft, memo) {
			out[score] += q * p / norm
		}
	}
	return out
}

// dealerFrom はディーラーが (total, soft) から引き切ったときの最終スコア分布。
func (r solverRules) dealerFrom(total int, soft bool, memo map[[2]int]map[int]float64) map[int]float64 {
	softKey := 0
	if soft {
		softKey = 1
	}
	key := [2]int{total, softKey}
	if v, ok := memo[key]; ok {
		return v
	}

	stand := total >= 17 && (total != 17 || !soft || !r.dealerHitsSoft17)
	if total > 21 {
		res := map[int]float64{dealerBust: 1}
		memo[key] = res
		return res
	}
	if stand {
		res := map[int]float64{total: 1}
		memo[key] = res
		return res
	}

	res := map[int]float64{}
	for c := 1; c <= 10; c++ {
		p := r.probs[c]
		if p == 0 {
			continue
		}
		nt, ns := addCard(total, soft, c)
		for score, q := range r.dealerFrom(nt, ns, memo) {
			res[score] += q * p
		}
	}
	memo[key] = res
	return res
}

// addCard は (total, soft) にBJ値 c を足した結果を返す。
func addCard(total int, soft bool, c int) (int, bool) {
	if c == 1 {
		if total+11 <= 21 {
			return total + 11, true
		}
		total++
	} else {
		total += c
	}
	if total > 21 && soft {
		return total - 10, false
	}
	return total, soft
}

// handState はプレイヤーのハンド。ranks は3枚未満のときだけ意味を持ち、
// 6-7-8 / 7-7-7 のトリオボーナス判定に使う。
type handState struct {
	total   int
	soft    bool
	n       int
	ranks   [2]int // n<=2 のときの実カード値 (0=未設定)
	doubled bool
	// fromSplit ならナチュラルBJ扱いしない (judgeHandCore の fromSplit)。
	fromSplit bool
}

// bonusMultiplier は勝利時のボーナス倍率 (利益/ベット) を返す。0 ならボーナス無し。
// spanish21BonusEval + spanishSuitBonus を写したもの。
// スート一致は分からないので **常に mixed (3:2)** として扱う — 同スートを仮定すると
// EV を過大評価する。
func (r solverRules) bonusMultiplier(h handState) float64 {
	if !r.bonuses || h.total != 21 || h.doubled {
		return 0
	}
	if h.n == 3 && h.ranks[0] != 0 {
		trio := []int{h.ranks[0], h.ranks[1], h.thirdRank()}
		sort.Ints(trio)
		if trio[0] == 6 && trio[1] == 7 && trio[2] == 8 {
			return 1.5
		}
		if trio[0] == 7 && trio[1] == 7 && trio[2] == 7 {
			return 1.5
		}
	}
	switch {
	case h.n == 5:
		return 1.5
	case h.n == 6:
		return 2
	case h.n >= 7:
		return 3
	}
	return 0
}

// thirdRank は3枚目の実カード値。ranks に入り切らないので total から逆算する。
func (h handState) thirdRank() int {
	return h.total - h.ranks[0] - h.ranks[1]
}

// isNaturalBJ は2枚21のナチュラル。
func (h handState) isNaturalBJ() bool { return h.n == 2 && h.total == 21 && !h.fromSplit }

// evStand は現在のハンドでスタンドしたときの純益 (ベット1単位あたり)。
func (r solverRules) evStand(h handState, dealer map[int]float64) float64 {
	unit := 1.0
	if h.doubled {
		unit = 2
	}
	// **バストはベット全額の負け。**ダブル後なら2単位。ここを 1 に固定すると
	// ソルバはハード12をダブルし始める (実際に踏んだ)。
	if h.total > 21 {
		return -unit
	}

	var ev float64
	for dscore, p := range dealer {
		res := r.judge(h, dscore)
		switch res {
		case GameResultLose:
			ev += p * -unit
		case GameResultDraw:
			// 0
		case GameResultWin:
			switch {
			case h.isNaturalBJ():
				ev += p * 1.5
			default:
				if m := r.bonusMultiplier(h); m > 0 {
					ev += p * m
				} else {
					ev += p * unit
				}
			}
		}
	}
	return ev
}

// judge は judgeHandCore を写したもの。ディーラーはナチュラルではない
// (ピーク済み) ので、そのケースは現れない。
func (r solverRules) judge(h handState, dscore int) GameResult {
	if h.total > 21 {
		return GameResultLose
	}
	if dscore == dealerBust {
		return GameResultWin
	}
	if h.total == 21 && r.player21AlwaysWins {
		return GameResultWin
	}
	if h.total > dscore {
		return GameResultWin
	}
	if dscore > h.total {
		return GameResultLose
	}
	// 同点。ディーラーはナチュラルでないので、プレイヤーがナチュラルなら勝ち。
	if h.isNaturalBJ() {
		return GameResultWin
	}
	return GameResultDraw
}

// evHit はヒットしてその後最善に打つときの純益。
func (r solverRules) evHit(h handState, dealer map[int]float64, memo map[string]float64) float64 {
	key := h.key() + "|hit"
	if v, ok := memo[key]; ok {
		return v
	}
	var ev float64
	for c := 1; c <= 10; c++ {
		p := r.probs[c]
		if p == 0 {
			continue
		}
		nh := h.draw(c)
		if nh.total > 21 {
			ev += p * -1
			continue
		}
		best := r.evStand(nh, dealer)
		if hitEV := r.evHit(nh, dealer, memo); hitEV > best {
			best = hitEV
		}
		ev += p * best
	}
	memo[key] = ev
	return ev
}

// evDouble は1枚だけ引いて倍賭けで降りるときの純益。
func (r solverRules) evDouble(h handState, dealer map[int]float64) float64 {
	var ev float64
	for c := 1; c <= 10; c++ {
		p := r.probs[c]
		if p == 0 {
			continue
		}
		nh := h.draw(c)
		nh.doubled = true
		ev += p * r.evStand(nh, dealer)
	}
	return ev
}

// draw はBJ値 c を1枚加えたハンドを返す。
func (h handState) draw(c int) handState {
	nh := h
	nh.total, nh.soft = addCard(h.total, h.soft, c)
	nh.n = h.n + 1
	if nh.n > 7 {
		nh.n = 7
	}
	if h.n >= 3 {
		nh.ranks = [2]int{} // 4枚目以降はトリオ判定に無関係
	}
	return nh
}

// key はメモ用のキー。
func (h handState) key() string {
	return fmt.Sprintf("%d/%t/%d/%d,%d/%t/%t", h.total, h.soft, h.n, h.ranks[0], h.ranks[1], h.doubled, h.fromSplit)
}

// evSplit はペアを割ったときの純益 (再スプリット無し、片方あたり×2の近似)。
func (r solverRules) evSplit(pv int, dealer map[int]float64, memo map[string]float64) float64 {
	var one float64
	for c := 1; c <= 10; c++ {
		p := r.probs[c]
		if p == 0 {
			continue
		}
		h := newHand(pv, c)
		h.fromSplit = true
		best := r.evStand(h, dealer)
		if v := r.evHit(h, dealer, memo); v > best {
			best = v
		}
		if r.canDouble {
			if v := r.evDouble(h, dealer); v > best {
				best = v
			}
		}
		one += p * best
	}
	return 2 * one
}

// newHand は実カード値 a, b の2枚ハンドを作る。
func newHand(a, b int) handState {
	h := handState{n: 2, ranks: [2]int{a, b}}
	t, s := addCard(0, false, bjOf(a))
	t, s = addCard(t, s, bjOf(b))
	h.total, h.soft = t, s
	return h
}

// bjOf は実カード値をBJ値へ。
func bjOf(v int) int {
	if v >= 10 {
		return 10
	}
	return v
}

// solveCell は1マスの最善手を返す。
func (r solverRules) solveCell(h handState, upcard int, isPair bool, pairValue int) BJSuggestedAction {
	dealer := r.dealerDist(upcard)
	memo := map[string]float64{}

	standEV := r.evStand(h, dealer)
	hitEV := r.evHit(h, dealer, memo)

	best := BJSuggestHit
	bestEV := hitEV
	if standEV > bestEV {
		best, bestEV = BJSuggestStand, standEV
	}
	if r.canDouble {
		if dv := r.evDouble(h, dealer); dv > bestEV {
			// スタンドの方がヒットより良い局面での D は「不可ならスタンド」。
			if standEV > hitEV {
				best = BJSuggestDoubleStand
			} else {
				best = BJSuggestDouble
			}
			bestEV = dv
		}
	}
	if isPair {
		if sv := r.evSplit(pairValue, dealer, memo); sv > bestEV {
			best, bestEV = BJSuggestSplit, sv
		}
	}
	if r.surrender && -0.5 > bestEV {
		best = BJSuggestSurrender
	}
	return best
}

// actionLetter は表の表示用。
func actionLetter(a BJSuggestedAction) string {
	switch a {
	case BJSuggestHit:
		return "H"
	case BJSuggestStand:
		return "S"
	case BJSuggestDouble:
		return "D"
	case BJSuggestDoubleStand:
		return "Ds"
	case BJSuggestSplit:
		return "Sp"
	case BJSuggestSurrender:
		return "Rh"
	}
	return "?"
}

// upcards は表の列。dealerIdx と同じ並び (2..9, 10, A)。
var solverUpcards = []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 1}

// renderRow は1行分を "{A, B, ...}" 形式で返す。
func renderRow(cells []BJSuggestedAction) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = actionLetter(c)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// evOfAction は指定アクションの純益を返す (比較用)。
func (r solverRules) evOfAction(h handState, upcard int, a BJSuggestedAction, isPair bool, pairValue int) float64 {
	dealer := r.dealerDist(upcard)
	memo := map[string]float64{}
	switch a {
	case BJSuggestStand:
		return r.evStand(h, dealer)
	case BJSuggestHit:
		return r.evHit(h, dealer, memo)
	case BJSuggestDouble, BJSuggestDoubleStand:
		return r.evDouble(h, dealer)
	case BJSuggestSplit:
		return r.evSplit(pairValue, dealer, memo)
	case BJSuggestSurrender:
		return -0.5
	}
	return 0
}
