package domain

import (
	"fmt"
	"math/rand"
)

// フェーズ定数
const (
	HoldemPhaseInit     = 0 // 初期状態
	HoldemPhasePreFlop  = 1 // プリフロップ
	HoldemPhaseFlop     = 2 // フロップ
	HoldemPhaseTurn     = 3 // ターン
	HoldemPhaseRiver    = 4 // リバー
	HoldemPhaseShowdown = 5 // ショーダウン
	HoldemPhaseEnd      = 6 // ゲーム終了
	HoldemPhaseRebuy    = 7 // リバイ/アドオン待ち
)

// アクション定数 (共通定数のエイリアス)
const (
	HoldemActionFold  = bettingActionFold  // フォールド
	HoldemActionCheck = bettingActionCheck // チェック
	HoldemActionCall  = bettingActionCall  // コール
	HoldemActionBet   = bettingActionBet   // ベット
	HoldemActionRaise = bettingActionRaise // レイズ
	HoldemActionAllIn = bettingActionAllIn // オールイン
)

// holdemDefaultMaxRaises Fixed/PotLimit時のデフォルト最大レイズ回数
const holdemDefaultMaxRaises = bettingMaxRaisesPerRound

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

// HoldemSidePot サイドポット (共通SidePot型のエイリアス)
type HoldemSidePot = SidePot

// HoldemResult ショーダウン結果
type HoldemResult struct {
	PlayerIdx int     // プレイヤーインデックス
	HandRank  int     // ハンドランク
	HandName  string  // ハンド名
	BestHand  []*Card // ベスト5枚
	Kickers   []int   // キッカーカード値
	WonAmount int     // 獲得チップ
}

// HoldemCpuAction CPU行動記録
type HoldemCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// リバイフェーズ種別定数
const (
	HoldemRebuyPhaseNone  = 0 // なし
	HoldemRebuyPhaseRebuy = 1 // リバイ待ち
	HoldemRebuyPhaseAddon = 2 // アドオン待ち
)

// Holdem テキサスホールデムクラス
type Holdem struct {
	trumpCards      *TrumpCards
	players         []*HoldemPlayer
	communityCards  []*Card
	pot             int
	sidePots        []HoldemSidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	config          HoldemConfig
	gameEndFlag     bool
	lastBet         int
	minRaise        int
	raiseCount      int
	actedFlags      []bool
	roundResults    []HoldemResult
	cpuActions      []HoldemCpuAction
	startingChips   []int
	vpipTracked     []bool // 当該ハンドでVPIP済みかどうか
	pfrTracked      []bool // 当該ハンドでPFR済みかどうか
	threeBetTracked []bool // 当該ハンドで3Bet追跡済みかどうか
	handCount       int    // ハンド数 (トーナメントモード用)
	lastCpuError    error  // CPU行動エラーの最後のフォールバック記録 (テスト検出用)
	rebuyCounts     []int  // プレイヤーごとのリバイ回数
	addonUsed       []bool // プレイヤーごとのアドオン使用フラグ
	rebuyPhaseType  int    // 0=none, 1=rebuy pending, 2=addon pending
}

// NewHoldem コンストラクタ
func NewHoldem(trumpCards *TrumpCards, players []*HoldemPlayer, config HoldemConfig) *Holdem {
	return &Holdem{
		trumpCards:      trumpCards,
		players:         players,
		communityCards:  make([]*Card, 0),
		sidePots:        make([]HoldemSidePot, 0),
		actedFlags:      make([]bool, len(players)),
		roundResults:    make([]HoldemResult, 0),
		cpuActions:      make([]HoldemCpuAction, 0),
		startingChips:   make([]int, len(players)),
		vpipTracked:     make([]bool, len(players)),
		pfrTracked:      make([]bool, len(players)),
		threeBetTracked: make([]bool, len(players)),
		rebuyCounts:     make([]int, len(players)),
		addonUsed:       make([]bool, len(players)),
		config:          config,
		phase:           HoldemPhaseInit,
	}
}

// Reset ゲーム初期化
func (h *Holdem) Reset() error {
	h.phase = HoldemPhaseInit
	h.pot = 0
	h.sidePots = make([]HoldemSidePot, 0)
	h.communityCards = make([]*Card, 0)
	h.gameEndFlag = false
	h.lastBet = 0
	h.minRaise = h.config.BigBlind
	h.raiseCount = 0
	h.actedFlags = make([]bool, len(h.players))
	h.roundResults = make([]HoldemResult, 0)
	h.cpuActions = make([]HoldemCpuAction, 0)
	h.rebuyPhaseType = HoldemRebuyPhaseNone

	h.trumpCards.Shuffle()
	for _, p := range h.players {
		p.Reset()
		p.SetFolded(false)
		p.SetAllIn(false)
		p.SetCurrentBet(0)
		p.handRank = 0
		p.bestHand = nil
		if p.GetChips() <= 0 && !h.config.RebuyEnabled {
			p.SetChips(h.config.InitChips)
		}
		p.IncrementTotalHands()
	}

	// HUDスタッツ追跡フラグをリセット
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))

	// トーナメントモード: ブラインドエスカレーション
	if h.config.TournamentMode && h.config.BlindLevelHands > 0 && h.handCount > 0 && h.handCount%h.config.BlindLevelHands == 0 {
		h.config.SmallBlind = h.config.SmallBlind * h.config.BlindMultiplier / 100
		h.config.BigBlind = h.config.BigBlind * h.config.BlindMultiplier / 100
		if h.config.SmallBlind < 1 {
			h.config.SmallBlind = 1
		}
		if h.config.BigBlind < 2 {
			h.config.BigBlind = 2
		}
	}
	h.handCount++

	// リバイチェック: リバイ有効 & リバイ期間内
	if h.config.RebuyEnabled && h.handCount <= h.config.RebuyPeriodHands {
		needHumanRebuy := false
		for i, p := range h.players {
			if p.GetChips() <= 0 && h.rebuyCounts[i] < h.config.RebuyMaxCount {
				if p.GetIsHuman() {
					needHumanRebuy = true
				} else {
					// CPU自動リバイ
					p.AddChips(h.config.RebuyChips)
					h.rebuyCounts[i]++
				}
			}
		}
		if needHumanRebuy {
			h.phase = HoldemPhaseRebuy
			h.rebuyPhaseType = HoldemRebuyPhaseRebuy
			return nil
		}
	}

	// アドオンチェック: アドオン有効 & アドオンハンド番号に到達
	if h.config.AddonEnabled && h.handCount == h.config.AddonAfterHand {
		needHumanAddon := false
		for i, p := range h.players {
			if !h.addonUsed[i] {
				if p.GetIsHuman() {
					needHumanAddon = true
				} else {
					// CPU自動アドオン
					p.AddChips(h.config.AddonChips)
					h.addonUsed[i] = true
				}
			}
		}
		if needHumanAddon {
			h.phase = HoldemPhaseRebuy
			h.rebuyPhaseType = HoldemRebuyPhaseAddon
			return nil
		}
	}

	return h.continueReset()
}

// continueReset ディール以降のリセット処理 (リバイ/アドオン判定後に実行)
func (h *Holdem) continueReset() error {
	// ハンド開始時のチップを記録 (サイドポット計算用)
	h.startingChips = make([]int, len(h.players))
	for i, p := range h.players {
		h.startingChips[i] = p.GetChips()
	}

	// ブラインド投入
	h.postBlinds()

	// ホールカード配布
	for i := 0; i < 2; i++ {
		for j := 0; j < len(h.players); j++ {
			idx := (h.dealerIdx + 1 + j) % len(h.players)
			card := h.trumpCards.DrawCard()
			if card != nil {
				h.players[idx].AddCard(card)
			}
		}
	}

	h.phase = HoldemPhasePreFlop
	// UTG (ビッグブラインドの次) から開始
	h.currentTurn = (h.dealerIdx + 3) % len(h.players)

	// CPUプリフロップアクション実行
	if err := h.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postBlinds ブラインド投入
func (h *Holdem) postBlinds() {
	sbIdx := (h.dealerIdx + 1) % len(h.players)
	bbIdx := (h.dealerIdx + 2) % len(h.players)

	sbAmount := h.config.SmallBlind
	if h.players[sbIdx].GetChips() < sbAmount {
		sbAmount = h.players[sbIdx].GetChips()
	}
	h.players[sbIdx].SubtractChips(sbAmount)
	h.players[sbIdx].SetCurrentBet(sbAmount)
	h.pot += sbAmount

	bbAmount := h.config.BigBlind
	if h.players[bbIdx].GetChips() < bbAmount {
		bbAmount = h.players[bbIdx].GetChips()
	}
	h.players[bbIdx].SubtractChips(bbAmount)
	h.players[bbIdx].SetCurrentBet(bbAmount)
	h.pot += bbAmount

	h.lastBet = bbAmount

	// チップが0になったらオールイン
	if h.players[sbIdx].GetChips() == 0 {
		h.players[sbIdx].SetAllIn(true)
		h.actedFlags[sbIdx] = true
	}
	if h.players[bbIdx].GetChips() == 0 {
		h.players[bbIdx].SetAllIn(true)
		h.actedFlags[bbIdx] = true
	}
}

// PlayerAction 人間プレイヤーのアクション実行
func (h *Holdem) PlayerAction(action, amount int) error {
	if h.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if h.phase < HoldemPhasePreFlop || h.phase > HoldemPhaseRiver {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !h.players[h.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	err := h.executeAction(h.currentTurn, action, amount)
	if err != nil {
		return err
	}

	h.advanceTurn()
	return h.runCpuActions()
}

// trackPreFlopStats プリフロップのHUDスタッツを追跡
func (h *Holdem) trackPreFlopStats(playerIdx, action int) {
	if h.phase != HoldemPhasePreFlop {
		return
	}

	isVPIPAction := false
	isPFRAction := false

	switch action {
	case HoldemActionCall:
		isVPIPAction = true
	case HoldemActionBet, HoldemActionRaise, HoldemActionAllIn:
		isVPIPAction = true
		isPFRAction = true
	}

	if isVPIPAction && !h.vpipTracked[playerIdx] {
		h.players[playerIdx].IncrementVPIP()
		h.vpipTracked[playerIdx] = true
	}

	if isPFRAction && !h.pfrTracked[playerIdx] {
		h.players[playerIdx].IncrementPFR()
		h.pfrTracked[playerIdx] = true
	}

	// 3Bet追跡: raiseCount >= 1 (既にレイズがある) かつ未追跡
	if h.raiseCount >= 1 && !h.threeBetTracked[playerIdx] {
		h.players[playerIdx].IncrementThreeBetOpportunity()
		if action == HoldemActionRaise || action == HoldemActionAllIn {
			h.players[playerIdx].IncrementThreeBet()
		}
		h.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats ポストフロップのAFスタッツを追跡
func (h *Holdem) trackPostFlopStats(playerIdx, action int) {
	if h.phase < HoldemPhaseFlop || h.phase > HoldemPhaseRiver {
		return
	}

	switch action {
	case HoldemActionBet, HoldemActionRaise, HoldemActionAllIn:
		h.players[playerIdx].IncrementPostFlopBetRaise()
	case HoldemActionCall:
		h.players[playerIdx].IncrementPostFlopCall()
	}
}

// bettingPlayers BettingPlayerスライスを生成
func (h *Holdem) bettingPlayers() []BettingPlayer {
	bp := make([]BettingPlayer, len(h.players))
	for i, pl := range h.players {
		bp[i] = pl
	}
	return bp
}

// executeAction 指定プレイヤーのアクション実行
func (h *Holdem) executeAction(playerIdx, action, amount int) error {
	// HUDスタッツ追跡 (Holdem固有)
	h.trackPreFlopStats(playerIdx, action)
	h.trackPostFlopStats(playerIdx, action)

	bp := h.bettingPlayers()
	// ActedFlags はスライス参照を共有: ExecuteBettingAction 内の変更が h.actedFlags に直接反映される
	state := &BettingState{
		Pot: h.pot, LastBet: h.lastBet, MinRaise: h.minRaise,
		RaiseCount: h.raiseCount, ActedFlags: h.actedFlags,
	}
	maxRaises, maxBetAmount := h.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, h.config.BigBlind, maxRaises, maxBetAmount)
	h.pot = state.Pot
	h.lastBet = state.LastBet
	h.minRaise = state.MinRaise
	h.raiseCount = state.RaiseCount
	if err != nil {
		return err
	}

	// フォールドでアクティブプレイヤーが1人になったらチェック
	if h.countActivePlayers() == 1 {
		h.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (h *Holdem) advanceTurn() {
	if h.gameEndFlag {
		return
	}

	// ベッティングラウンド終了チェック
	if h.isBettingRoundComplete() {
		h.advancePhase()
		return
	}

	// 次のアクティブプレイヤーを探す
	for i := 1; i <= len(h.players); i++ {
		next := (h.currentTurn + i) % len(h.players)
		if !h.players[next].GetFolded() && !h.players[next].GetAllIn() && !h.actedFlags[next] {
			h.currentTurn = next
			return
		}
	}

	// 全員行動済みならフェーズ進行
	h.advancePhase()
}

// isBettingRoundComplete ベッティングラウンドが完了したかチェック
func (h *Holdem) isBettingRoundComplete() bool {
	for i, p := range h.players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if !h.actedFlags[i] {
			return false
		}
	}
	return true
}

// advancePhase 次のフェーズに進める
func (h *Holdem) advancePhase() {
	// ラウンドベットリセット
	for _, p := range h.players {
		p.SetCurrentBet(0)
	}
	h.lastBet = 0
	h.minRaise = h.config.BigBlind
	h.raiseCount = 0
	h.actedFlags = make([]bool, len(h.players))
	// フォールド・オールインプレイヤーはacted扱い
	for i, p := range h.players {
		if p.GetFolded() || p.GetAllIn() {
			h.actedFlags[i] = true
		}
	}

	switch h.phase {
	case HoldemPhasePreFlop:
		h.phase = HoldemPhaseFlop
		for i := 0; i < 3; i++ {
			card := h.trumpCards.DrawCard()
			if card != nil {
				h.communityCards = append(h.communityCards, card)
			}
		}
	case HoldemPhaseFlop:
		h.phase = HoldemPhaseTurn
		card := h.trumpCards.DrawCard()
		if card != nil {
			h.communityCards = append(h.communityCards, card)
		}
	case HoldemPhaseTurn:
		h.phase = HoldemPhaseRiver
		card := h.trumpCards.DrawCard()
		if card != nil {
			h.communityCards = append(h.communityCards, card)
		}
	case HoldemPhaseRiver:
		h.phase = HoldemPhaseShowdown
		h.resolveShowdown()
		return
	}

	// アクティブプレイヤーが0-1人ならショーダウンへ
	activeCnt := 0
	for _, p := range h.players {
		if !p.GetFolded() && !p.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		// 全員オールインまたはフォールド → 残りのコミュニティカードを配ってショーダウン
		h.dealRemainingCommunity()
		h.phase = HoldemPhaseShowdown
		h.resolveShowdown()
		return
	}

	// ディーラーの次のアクティブプレイヤーから開始
	h.currentTurn = h.findNextActive(h.dealerIdx)
}

// dealRemainingCommunity 残りのコミュニティカードを全て配る
func (h *Holdem) dealRemainingCommunity() {
	for len(h.communityCards) < 5 {
		card := h.trumpCards.DrawCard()
		if card == nil {
			break
		}
		h.communityCards = append(h.communityCards, card)
	}
}

// findNextActive 指定インデックスの次のアクティブ (フォールド・オールインでない) プレイヤーを探す
func (h *Holdem) findNextActive(fromIdx int) int {
	for i := 1; i <= len(h.players); i++ {
		next := (fromIdx + i) % len(h.players)
		if !h.players[next].GetFolded() && !h.players[next].GetAllIn() {
			return next
		}
	}
	return (fromIdx + 1) % len(h.players)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (h *Holdem) countActivePlayers() int {
	cnt := 0
	for _, p := range h.players {
		if !p.GetFolded() {
			cnt++
		}
	}
	return cnt
}

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (h *Holdem) resolveLastPlayer() {
	for i, p := range h.players {
		if !p.GetFolded() {
			p.AddChips(h.pot)
			h.roundResults = []HoldemResult{{
				PlayerIdx: i,
				WonAmount: h.pot,
			}}
			h.pot = 0
			break
		}
	}
	h.phase = HoldemPhaseEnd
	h.gameEndFlag = true
	h.dealerIdx = (h.dealerIdx + 1) % len(h.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (h *Holdem) resolveShowdown() {
	// ハンド評価
	for _, p := range h.players {
		if !p.GetFolded() {
			p.EvalBestHand(h.communityCards)
		}
	}

	// サイドポット計算・配分
	bp := h.bettingPlayers()
	h.sidePots = CalculateSidePots(bp, h.pot, h.startingChips)
	wonAmounts := DistributePots(bp, h.sidePots)

	// 結果を構築
	h.roundResults = make([]HoldemResult, 0)
	for i, p := range h.players {
		if p.GetFolded() {
			continue
		}
		result := HoldemResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  h.getHandName(p.GetHandRank()),
			BestHand:  p.GetBestHand(),
			Kickers:   ExtractKickers(p.GetBestHand(), p.GetHandRank()),
			WonAmount: wonAmounts[i],
		}
		h.roundResults = append(h.roundResults, result)
	}

	h.phase = HoldemPhaseEnd
	h.gameEndFlag = true
	h.dealerIdx = (h.dealerIdx + 1) % len(h.players)
}

// getHandName ハンドランクから名前を返す
func (h *Holdem) getHandName(rank int) string {
	if rank >= 0 && rank < len(PokerHandNames) {
		return PokerHandNames[rank]
	}
	return "Unknown"
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

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (h *Holdem) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(h.config.BettingLimit, h.pot, h.lastBet)
}

// --- GTO (Game Theory Optimal) AI ---

// GTO ベットサイズ定数 (ポット比率 %)
const (
	gtoPreFlopBetPct  = 66 // プリフロップ: 2/3ポット
	gtoDryBoardBetPct = 66 // ドライボード: 2/3ポット
	gtoWetBoardBetPct = 75 // ウェットボード: 3/4ポット
)

// boardTexture ボードテクスチャ分析結果
type boardTexture struct {
	paired       bool // ペアボード (同じ数字が2枚以上)
	flushDraw    bool // フラッシュドロー (同スート3枚以上)
	straightDraw bool // ストレートドロー (5幅ウィンドウ内に3枚以上)
	wet          bool // ウェットボード (flushDraw || straightDraw)
	highCards    int  // ハイカード数 (10以上 or A)
}

// evalBoardTexture コミュニティカードからボードテクスチャを分析
func evalBoardTexture(communityCards []*Card) boardTexture {
	bt := boardTexture{}
	if len(communityCards) == 0 {
		return bt
	}

	// ペア判定
	valCount := make(map[int]int)
	suitCount := make(map[int]int)
	for _, c := range communityCards {
		valCount[c.GetValue()]++
		suitCount[c.GetDesign()]++
	}
	for _, cnt := range valCount {
		if cnt >= 2 {
			bt.paired = true
			break
		}
	}

	// フラッシュドロー判定
	for _, cnt := range suitCount {
		if cnt >= 3 {
			bt.flushDraw = true
			break
		}
	}

	// ストレートドロー判定: 5幅ウィンドウ内に3枚以上
	vals := make(map[int]bool)
	for _, c := range communityCards {
		v := c.GetValue()
		vals[v] = true
		if v == 1 {
			vals[14] = true // A=14としても扱う
		}
	}
	for base := 1; base <= 10; base++ {
		cnt := 0
		for v := base; v < base+5; v++ {
			if vals[v] {
				cnt++
			}
		}
		if cnt >= 3 {
			bt.straightDraw = true
			break
		}
	}

	bt.wet = bt.flushDraw || bt.straightDraw

	// ハイカード数 (10以上 or A)
	for _, c := range communityCards {
		v := c.GetValue()
		if v >= 10 || v == 1 {
			bt.highCards++
		}
	}

	return bt
}

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

	var action, amount int
	if h.phase == HoldemPhasePreFlop {
		action, amount = h.cpuDecidePreFlop(idx, params, callAmount)
	} else {
		action, amount = h.cpuDecidePostFlop(idx, params, callAmount)
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
	bet := h.pot * potPct / 100
	if bet < h.config.BigBlind {
		bet = h.config.BigBlind
	}
	if bet < h.minRaise {
		bet = h.minRaise
	}
	return bet
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

// clamp 値を範囲内に収める
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// --- リバイ/アドオン ---

// checkAndTransitionAddon はアドオン判定を行い、人間にアドオン決定を促す場合は true を返す。
// CPU は自動でアドオンを実行する。
func (h *Holdem) checkAndTransitionAddon() bool {
	if h.config.AddonEnabled && h.handCount == h.config.AddonAfterHand {
		needHumanAddon := false
		for i, p := range h.players {
			if !h.addonUsed[i] {
				if p.GetIsHuman() {
					needHumanAddon = true
				} else {
					p.AddChips(h.config.AddonChips)
					h.addonUsed[i] = true
				}
			}
		}
		if needHumanAddon {
			h.rebuyPhaseType = HoldemRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (h *Holdem) Rebuy() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	for i, p := range h.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && h.rebuyCounts[i] < h.config.RebuyMaxCount {
			p.AddChips(h.config.RebuyChips)
			h.rebuyCounts[i]++
			break
		}
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	if h.checkAndTransitionAddon() {
		return nil
	}
	return h.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (h *Holdem) SkipRebuy() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	// バスト中の人間がリバイを辞退 → ゲーム終了
	for _, p := range h.players {
		if p.GetIsHuman() && p.GetChips() <= 0 {
			h.phase = HoldemPhaseEnd
			h.gameEndFlag = true
			return nil
		}
	}
	// 人間にチップが残っている場合 (通常ありえないが安全策)
	if h.checkAndTransitionAddon() {
		return nil
	}
	return h.continueReset()
}

// Addon 人間プレイヤーがアドオンを実行する
func (h *Holdem) Addon() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, p := range h.players {
		if p.GetIsHuman() && !h.addonUsed[i] {
			p.AddChips(h.config.AddonChips)
			h.addonUsed[i] = true
			break
		}
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	return h.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (h *Holdem) SkipAddon() error {
	if h.phase != HoldemPhaseRebuy || h.rebuyPhaseType != HoldemRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	return h.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (h *Holdem) IsRebuyAvailable() bool {
	if !h.config.RebuyEnabled || h.handCount > h.config.RebuyPeriodHands {
		return false
	}
	for i, p := range h.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && h.rebuyCounts[i] < h.config.RebuyMaxCount {
			return true
		}
	}
	return false
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (h *Holdem) IsAddonAvailable() bool {
	if !h.config.AddonEnabled || h.handCount != h.config.AddonAfterHand {
		return false
	}
	for i, p := range h.players {
		if p.GetIsHuman() && !h.addonUsed[i] {
			return true
		}
	}
	return false
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (h *Holdem) GetRebuyCounts() []int {
	result := make([]int, len(h.rebuyCounts))
	copy(result, h.rebuyCounts)
	return result
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (h *Holdem) GetAddonUsed() []bool {
	result := make([]bool, len(h.addonUsed))
	copy(result, h.addonUsed)
	return result
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (h *Holdem) GetRebuyPhaseType() int { return h.rebuyPhaseType }

// --- ゲッター ---

// GetPhase フェーズ取得
func (h *Holdem) GetPhase() int { return h.phase }

// GetPlayers プレイヤー一覧取得
func (h *Holdem) GetPlayers() []*HoldemPlayer { return h.players }

// GetPlayer 指定プレイヤー取得
func (h *Holdem) GetPlayer(i int) *HoldemPlayer {
	if i >= 0 && i < len(h.players) {
		return h.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (h *Holdem) GetPlayerCnt() int { return len(h.players) }

// GetCommunityCards コミュニティカード取得
func (h *Holdem) GetCommunityCards() []*Card { return h.communityCards }

// GetPot ポット取得
func (h *Holdem) GetPot() int { return h.pot }

// GetSidePots サイドポット取得
func (h *Holdem) GetSidePots() []HoldemSidePot { return h.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (h *Holdem) GetDealerIdx() int { return h.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (h *Holdem) GetCurrentTurn() int { return h.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (h *Holdem) GetGameEndFlag() bool { return h.gameEndFlag }

// GetLastBet 最後のベット取得
func (h *Holdem) GetLastBet() int { return h.lastBet }

// GetMinRaise 最小レイズ額取得
func (h *Holdem) GetMinRaise() int { return h.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (h *Holdem) GetRaiseCount() int { return h.raiseCount }

// GetRoundResults ラウンド結果取得
func (h *Holdem) GetRoundResults() []HoldemResult { return h.roundResults }

// GetCpuActions CPU行動記録取得
func (h *Holdem) GetCpuActions() []HoldemCpuAction { return h.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得 (テスト・デバッグ用)
func (h *Holdem) GetLastCpuError() error { return h.lastCpuError }

// GetConfig 設定取得
func (h *Holdem) GetConfig() HoldemConfig { return h.config }

// SetConfig 設定変更
func (h *Holdem) SetConfig(cfg HoldemConfig) { h.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (h *Holdem) IsHumanTurn() bool {
	if h.currentTurn >= 0 && h.currentTurn < len(h.players) {
		return h.players[h.currentTurn].GetIsHuman()
	}
	return false
}

// GetActedFlags actedフラグ取得
func (h *Holdem) GetActedFlags() []bool {
	result := make([]bool, len(h.actedFlags))
	copy(result, h.actedFlags)
	return result
}

// GetHandCount ハンド数取得
func (h *Holdem) GetHandCount() int { return h.handCount }

// Resize プレイヤースライスを差し替え、プレイヤー数依存スライスを再初期化する
func (h *Holdem) Resize(players []*HoldemPlayer) {
	h.players = players
	n := len(players)
	h.actedFlags = make([]bool, n)
	h.startingChips = make([]int, n)
	h.vpipTracked = make([]bool, n)
	h.pfrTracked = make([]bool, n)
	h.threeBetTracked = make([]bool, n)
	h.rebuyCounts = make([]int, n)
	h.addonUsed = make([]bool, n)
	h.handCount = 0
}
