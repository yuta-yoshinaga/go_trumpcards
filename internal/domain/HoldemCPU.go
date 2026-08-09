//go:build !js || !wasm || casino

package domain

import (
	"fmt"
	"math"
	"math/rand"
)

// cpuStyleParams CPU意思決定パラメータ
type cpuStyleParams struct {
	aggressive bool // true=Aggressive(TAG/LAG), false=Passive(LAP/TAP)
	bluffRate  int  // ブラフ率(%)

	// PreFlop
	preFlopFoldThreshold  int  // strength < this → fold評価
	preFlopFoldCompound   bool // true: callAmount条件付きFold(Loose), false: foldOrCheck(Tight)
	preFlopFoldCallMult   int  // Loose only: callAmount > BB*this でFold
	preFlopRaiseThreshold int  // Aggressive only: strength >= this → raise
	preFlopRaisePotPct    int  // Aggressive only: raise = pot*this/100
	preFlopBluffPotPct    int  // Passive only: bluff bet = pot*this/100

	// PostFlop Aggressive
	postFlopRaiseRank    int  // handRank >= this → raise
	postFlopRaisePotPct  int  // raise = pot*this/100
	postFlopCondCallRank int  // handRank >= this → Call (-1=skip)
	postFlopFallbackFold bool // true=foldOrCheck(TAG), false=callブロック+条件付きFold(LAG)
	postFlopAggrFoldRank int  // fold if handRank <= this
	postFlopAggrFoldMult int  // fold if callAmount > BB*this

	// PostFlop Passive
	postFlopPassFoldRank int // fold if handRank <= this
	postFlopPassFoldMult int // fold if callAmount > BB*this (-1=any callAmount)
	postFlopBluffPotPct  int // bluff bet = pot*this/100
}

// holdemStyleParamsMap スタイルごとのパラメータ
var holdemStyleParamsMap = map[HoldemPlayStyle]cpuStyleParams{
	HoldemStyleTAG: {
		aggressive: true, bluffRate: 15,
		preFlopFoldThreshold: 40, preFlopFoldCompound: false,
		preFlopRaiseThreshold: 70, preFlopRaisePotPct: 75,
		postFlopRaiseRank: PokerHandTwoPair, postFlopRaisePotPct: 66,
		postFlopCondCallRank: PokerHandOnePair, postFlopFallbackFold: true,
	},
	HoldemStyleLAP: {
		aggressive: false, bluffRate: 5,
		preFlopFoldThreshold: 15, preFlopFoldCompound: true, preFlopFoldCallMult: 2,
		preFlopBluffPotPct:   50,
		postFlopPassFoldRank: PokerHandHighCard, postFlopPassFoldMult: 3,
		postFlopBluffPotPct: 33,
	},
	HoldemStyleTAP: {
		aggressive: false, bluffRate: 5,
		preFlopFoldThreshold: 30, preFlopFoldCompound: false,
		preFlopBluffPotPct:   50,
		postFlopPassFoldRank: PokerHandHighCard, postFlopPassFoldMult: -1,
		postFlopBluffPotPct: 33,
	},
	HoldemStyleLAG: {
		aggressive: true, bluffRate: 30,
		preFlopFoldThreshold: 15, preFlopFoldCompound: true, preFlopFoldCallMult: 3,
		preFlopRaiseThreshold: 50, preFlopRaisePotPct: 100,
		postFlopRaiseRank: PokerHandOnePair, postFlopRaisePotPct: 100,
		postFlopFallbackFold: false,
		postFlopAggrFoldRank: PokerHandHighCard, postFlopAggrFoldMult: 4,
	},
}

// runCpuActions CPUプレイヤーのアクションを実行
func (h *Holdem) runCpuActions() error {
	if h.gameEndFlag {
		return nil
	}
	// maxIterationsはCPUアクションの無限ループを防ぐための安全策。
	// 1ラウンドのアクションは最大でも「ベット→レイズ→リレイズ→キャップ」の4レイズ + 各プレイヤーのコールとなり、
	// 9人プレイでは1ラウンド最大約45アクション、4ベッティングラウンドで約360アクション程度のため、500は十分な安全マージン。
	const maxIterations = 500
	iterations := 0
	for !h.gameEndFlag && h.phase >= HoldemPhasePreFlop && h.phase <= HoldemPhaseRiver {
		iterations++
		if iterations > maxIterations {
			return fmt.Errorf("maxIterations reached in runCpuActions, possible infinite loop")
		}
		if h.players[h.currentTurn].GetIsHuman() {
			return nil
		}
		if h.players[h.currentTurn].GetFolded() || h.players[h.currentTurn].GetAllIn() {
			h.advanceTurn()
			continue
		}
		action, amount := h.cpuDecide(h.currentTurn)
		h.cpuActions = append(h.cpuActions, HoldemCpuAction{
			PlayerIdx: h.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := h.executeAction(h.currentTurn, action, amount)
		if err != nil {
			h.handleCpuActionError(h.currentTurn, action, err)
		}
		if h.gameEndFlag {
			return nil
		}
		h.advanceTurn()
	}
	return nil
}

// handleCpuActionError CPUアクション失敗時のフォールバック処理
func (h *Holdem) handleCpuActionError(playerIdx, action int, err error) {
	h.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", playerIdx, action, err)
	callAmt := h.lastBet - h.players[playerIdx].GetCurrentBet()
	if callAmt > 0 {
		_ = h.executeAction(playerIdx, HoldemActionFold, 0)
	} else {
		_ = h.executeAction(playerIdx, HoldemActionCheck, 0)
	}
}

// --- GTO (Game Theory Optimal) AI ---

// GTO ベットサイズ定数 (ポット比率 %)
const (
	gtoPreFlopBetPct  = 66 // プリフロップ: 2/3ポット
	gtoDryBoardBetPct = 66 // ドライボード: 2/3ポット
	gtoWetBoardBetPct = 75 // ウェットボード: 3/4ポット
)

// gtoHandCategory GTOハンドカテゴリ
type gtoHandCategory int

const (
	gtoHandTrash  gtoHandCategory = 0 // ゴミ
	gtoHandWeak   gtoHandCategory = 1 // 弱い
	gtoHandMedium gtoHandCategory = 2 // 中程度
	gtoHandStrong gtoHandCategory = 3 // 強い
	gtoHandNuts   gtoHandCategory = 4 // ナッツ級
)

// classifyGTOHand ハンドランクからGTOカテゴリに分類
func classifyGTOHand(handRank int) gtoHandCategory {
	switch {
	case handRank >= PokerHandFourOfAKind: // FourOfAKind, StraightFlush, RoyalFlush, FiveOfAKind
		return gtoHandNuts
	case handRank >= PokerHandFlush: // Flush, FullHouse
		return gtoHandStrong
	case handRank >= PokerHandThreeOfAKind: // ThreeOfAKind, Straight
		return gtoHandMedium
	case handRank >= PokerHandOnePair: // OnePair, TwoPair
		return gtoHandWeak
	default: // HighCard
		return gtoHandTrash
	}
}

// gtoActionDist GTOアクション確率分布 (合計100)
type gtoActionDist struct {
	foldPct  int
	checkPct int
	betPct   int
}

// gtoPreFlopTable プリフロップGTOアクション分布テーブル
// strength: 0-19=trash, 20-39=weak, 40-59=medium, 60-79=strong, 80-100=premium
var gtoPreFlopTable = [5]gtoActionDist{
	{foldPct: 70, checkPct: 20, betPct: 10}, // trash (0-19)
	{foldPct: 40, checkPct: 35, betPct: 25}, // weak (20-39)
	{foldPct: 15, checkPct: 35, betPct: 50}, // medium (40-59)
	{foldPct: 5, checkPct: 20, betPct: 75},  // strong (60-79)
	{foldPct: 0, checkPct: 10, betPct: 90},  // premium (80-100)
}

// gtoPostFlopTable ポストフロップGTOアクション分布テーブル [handCategory][0=dry,1=wet]
var gtoPostFlopTable = [5][2]gtoActionDist{
	// trash
	{{foldPct: 60, checkPct: 30, betPct: 10}, {foldPct: 75, checkPct: 20, betPct: 5}},
	// weak
	{{foldPct: 25, checkPct: 45, betPct: 30}, {foldPct: 35, checkPct: 40, betPct: 25}},
	// medium
	{{foldPct: 5, checkPct: 30, betPct: 65}, {foldPct: 10, checkPct: 25, betPct: 65}},
	// strong
	{{foldPct: 0, checkPct: 20, betPct: 80}, {foldPct: 0, checkPct: 15, betPct: 85}},
	// nuts
	{{foldPct: 0, checkPct: 15, betPct: 85}, {foldPct: 0, checkPct: 10, betPct: 90}},
}

// gtoPreFlopIndex プリフロップ強度からテーブルインデックスを返す
func gtoPreFlopIndex(strength int) int {
	switch {
	case strength >= 80:
		return 4
	case strength >= 60:
		return 3
	case strength >= 40:
		return 2
	case strength >= 20:
		return 1
	default:
		return 0
	}
}

// gtoRollAction 確率分布に基づいてアクションを決定
func gtoRollAction(dist gtoActionDist) int {
	roll := rand.Intn(100)
	if roll < dist.foldPct {
		return 0 // fold
	}
	if roll < dist.foldPct+dist.checkPct {
		return 1 // check/call
	}
	return 2 // bet/raise
}

// cpuDecidePreFlopGTO GTOプリフロップ意思決定
func (h *Holdem) cpuDecidePreFlopGTO(idx, callAmount int) (int, int) {
	p := h.players[idx]
	strength := h.evalPreFlopStrength(idx)
	dist := gtoPreFlopTable[gtoPreFlopIndex(strength)]
	decision := gtoRollAction(dist)

	switch decision {
	case 0: // fold
		return h.cpuFoldOrCheck(callAmount)
	case 2: // bet/raise
		betAmt := h.cpuPotBet(gtoPreFlopBetPct)
		return h.cpuRaiseOrBet(p, callAmount, betAmt)
	default: // check/call
		return h.cpuCallOrCheck(callAmount)
	}
}

// cpuDecidePostFlopGTO GTOポストフロップ意思決定
func (h *Holdem) cpuDecidePostFlopGTO(idx, callAmount int) (int, int) {
	p := h.players[idx]
	handRank := p.EvalBestHand(h.communityCards)
	category := classifyGTOHand(handRank)
	bt := evalBoardTexture(h.communityCards)

	wetIdx := 0
	if bt.wet {
		wetIdx = 1
	}
	dist := gtoPostFlopTable[category][wetIdx]
	decision := gtoRollAction(dist)

	// ベットサイズ: ドライ=2/3ポット, ウェット=3/4ポット
	potPct := gtoDryBoardBetPct
	if bt.wet {
		potPct = gtoWetBoardBetPct
	}

	// ペアボードではベット頻度を下げる (トラップ重視)
	if bt.paired && decision == 2 && category <= gtoHandMedium && rand.Intn(100) < 30 {
		return h.cpuCallOrCheck(callAmount)
	}

	// ハイカードが多いボードではブラフを抑制
	if bt.highCards >= 3 && decision == 2 && category <= gtoHandWeak && rand.Intn(100) < 40 {
		return h.cpuCallOrCheck(callAmount)
	}

	switch decision {
	case 0: // fold
		return h.cpuFoldOrCheck(callAmount)
	case 2: // bet/raise
		betAmt := h.cpuPotBet(potPct)
		return h.cpuRaiseOrBet(p, callAmount, betAmt)
	default: // check/call
		return h.cpuCallOrCheck(callAmount)
	}
}

// cpuDecide CPUプレイヤーの意思決定
func (h *Holdem) cpuDecide(idx int) (int, int) {
	p := h.players[idx]
	style := p.GetPlayStyle()
	callAmount := h.lastBet - p.GetCurrentBet()

	// GTO: 独自の混合戦略ロジックを使用
	if style == HoldemStyleGTO {
		var action, amount int
		if h.phase == HoldemPhasePreFlop {
			action, amount = h.cpuDecidePreFlopGTO(idx, callAmount)
		} else {
			action, amount = h.cpuDecidePostFlopGTO(idx, callAmount)
		}
		maxRaises, maxBetAmount := h.bettingLimits()
		if maxBetAmount > 0 && amount > maxBetAmount {
			amount = maxBetAmount
		}
		if maxRaises > 0 && h.raiseCount >= maxRaises {
			if action == HoldemActionRaise || action == HoldemActionBet {
				if callAmount > 0 {
					return HoldemActionCall, 0
				}
				return HoldemActionCheck, 0
			}
		}
		return action, amount
	}

	params, ok := holdemStyleParamsMap[style]
	if !ok {
		return h.cpuCallOrCheck(callAmount)
	}

	// メタAI: ブラフ率を調整
	if h.config.CpuMetaAI && h.humanProfile != nil {
		adjusted := h.humanProfile.AdjustedBluffChance(float64(params.bluffRate))
		params.bluffRate = int(math.Round(adjusted))
	}

	var action, amount int
	if h.phase == HoldemPhasePreFlop {
		action, amount = h.cpuDecidePreFlop(idx, params, callAmount)
	} else {
		action, amount = h.cpuDecidePostFlop(idx, params, callAmount)
	}

	// メタAI: 人間のベット/レイズに対してコール確率を調整
	if h.config.CpuMetaAI && h.humanProfile != nil && h.lastHumanPlayMs > 0 {
		if action == HoldemActionFold && callAmount > 0 {
			handRank := p.EvalBestHand(h.communityCards)
			bracket := bettingHandBracket(handRank)
			adjustedCall := h.humanProfile.AdjustedCallChance(0.0, bracket, h.lastHumanPlayMs)
			if adjustedCall > 0 && rand.Float64() < adjustedCall { //nolint:gosec // non-crypto random for game AI
				action = HoldemActionCall
				amount = 0
			}
		}
	}

	maxRaises, maxBetAmount := h.bettingLimits()

	// PotLimit: CPUベット額をポットサイズに制限
	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}

	// レイズ上限に達したら、レイズ/ベットをコール/チェックに変更
	if maxRaises > 0 && h.raiseCount >= maxRaises {
		if action == HoldemActionRaise || action == HoldemActionBet {
			if callAmount > 0 {
				return HoldemActionCall, 0
			}
			return HoldemActionCheck, 0
		}
	}
	return action, amount
}

// cpuFoldOrCheck コール額がある場合はフォールド、なければチェック
func (h *Holdem) cpuFoldOrCheck(callAmount int) (int, int) {
	return CpuFoldOrCheck(callAmount)
}

// cpuCallOrCheck コール額がある場合はコール、なければチェック
func (h *Holdem) cpuCallOrCheck(callAmount int) (int, int) {
	return CpuCallOrCheck(callAmount)
}

// cpuRaiseOrBet コール額がある場合はレイズ、なければベット (チップ不足時はオールイン)
func (h *Holdem) cpuRaiseOrBet(p *HoldemPlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(p.GetChips(), callAmount, raiseAmt)
}

// cpuBetOrAllIn ベットする (チップ不足時はオールイン)
func (h *Holdem) cpuBetOrAllIn(p *HoldemPlayer, betAmt int) (int, int) {
	if betAmt > p.GetChips() {
		return HoldemActionAllIn, 0
	}
	return HoldemActionBet, betAmt
}

// cpuPotBet ポット比率ベースのベット額を計算 (最低BB, 最低minRaise)
func (h *Holdem) cpuPotBet(potPct int) int {
	return potBet(h.pot, potPct, h.config.BigBlind, h.minRaise)
}

// cpuDecidePreFlop プリフロップのCPU意思決定
func (h *Holdem) cpuDecidePreFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := h.players[idx]
	strength := h.evalPreFlopStrength(idx)

	// Fold評価
	if strength < params.preFlopFoldThreshold {
		if params.preFlopFoldCompound {
			if callAmount > h.config.BigBlind*params.preFlopFoldCallMult {
				return HoldemActionFold, 0
			}
		} else {
			return h.cpuFoldOrCheck(callAmount)
		}
	}

	if params.aggressive {
		// Aggressive: raise or call
		if strength >= params.preFlopRaiseThreshold || rand.Intn(100) < params.bluffRate {
			return h.cpuRaiseOrBet(p, callAmount, h.cpuPotBet(params.preFlopRaisePotPct))
		}
		return h.cpuCallOrCheck(callAmount)
	}

	// Passive: call, bluff bet, or check
	if callAmount > 0 {
		return HoldemActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return h.cpuBetOrAllIn(p, h.cpuPotBet(params.preFlopBluffPotPct))
	}
	return HoldemActionCheck, 0
}

// cpuDecidePostFlop フロップ以降のCPU意思決定
func (h *Holdem) cpuDecidePostFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := h.players[idx]
	handRank := p.EvalBestHand(h.communityCards)

	if params.aggressive {
		// Aggressive: raise → conditional call/fold
		if handRank >= params.postFlopRaiseRank || rand.Intn(100) < params.bluffRate {
			return h.cpuRaiseOrBet(p, callAmount, h.cpuPotBet(params.postFlopRaisePotPct))
		}
		if params.postFlopFallbackFold {
			// TAG: conditional call then foldOrCheck
			if params.postFlopCondCallRank >= 0 && handRank >= params.postFlopCondCallRank && callAmount > 0 {
				return HoldemActionCall, 0
			}
			return h.cpuFoldOrCheck(callAmount)
		}
		// LAG: call block with conditional fold
		if callAmount > 0 {
			if handRank <= params.postFlopAggrFoldRank && callAmount > h.config.BigBlind*params.postFlopAggrFoldMult {
				return HoldemActionFold, 0
			}
			return HoldemActionCall, 0
		}
		return HoldemActionCheck, 0
	}

	// Passive: call block with conditional fold → bluff bet → check
	if callAmount > 0 {
		if handRank <= params.postFlopPassFoldRank {
			if params.postFlopPassFoldMult < 0 || callAmount > h.config.BigBlind*params.postFlopPassFoldMult {
				return HoldemActionFold, 0
			}
		}
		return HoldemActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return h.cpuBetOrAllIn(p, h.cpuPotBet(params.postFlopBluffPotPct))
	}
	return HoldemActionCheck, 0
}

// evalPreFlopStrength プリフロップハンド強度評価 (0-100)
func (h *Holdem) evalPreFlopStrength(idx int) int {
	p := h.players[idx]
	if p.GetCardsSize() < 2 {
		return 0
	}
	v1 := p.GetCard(0).GetValue()
	v2 := p.GetCard(1).GetValue()
	d1 := p.GetCard(0).GetDesign()
	d2 := p.GetCard(1).GetDesign()

	// エースを14として扱う
	if v1 == 1 {
		v1 = 14
	}
	if v2 == 1 {
		v2 = 14
	}

	suited := d1 == d2
	score := 0

	// ペア
	if v1 == v2 {
		score = 50 + v1*3
		if v1 >= 10 {
			score += 15
		}
		return clamp(score, 0, 100)
	}

	// ハイカード値
	high := v1
	low := v2
	if v2 > v1 {
		high = v2
		low = v1
	}

	score = high*2 + low
	if suited {
		score += 10
	}

	// コネクタ (連続数字)
	gap := high - low
	switch gap {
	case 1:
		score += 10
	case 2:
		score += 5
	}

	// ハイカードボーナス
	if high >= 12 {
		score += 10
	}

	return clamp(score, 0, 100)
}
