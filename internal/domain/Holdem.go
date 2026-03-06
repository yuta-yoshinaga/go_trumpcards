package domain

import (
	"fmt"
	"math/rand"
)

// テキサスホールデムプレイヤー数
const HoldemPlayerCnt = 4

// フェーズ定数
const (
	HoldemPhaseInit     = 0 // 初期状態
	HoldemPhasePreFlop  = 1 // プリフロップ
	HoldemPhaseFlop     = 2 // フロップ
	HoldemPhaseTurn     = 3 // ターン
	HoldemPhaseRiver    = 4 // リバー
	HoldemPhaseShowdown = 5 // ショーダウン
	HoldemPhaseEnd      = 6 // ゲーム終了
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
	WonAmount int     // 獲得チップ
}

// HoldemCpuAction CPU行動記録
type HoldemCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// Holdem テキサスホールデムクラス
type Holdem struct {
	trumpCards     *TrumpCards
	players        []*HoldemPlayer
	communityCards []*Card
	pot            int
	sidePots       []HoldemSidePot
	dealerIdx      int
	currentTurn    int
	phase          int
	config         HoldemConfig
	gameEndFlag    bool
	lastBet        int
	minRaise       int
	raiseCount     int
	actedFlags     []bool
	roundResults   []HoldemResult
	cpuActions     []HoldemCpuAction
	startingChips  []int
	vpipTracked    []bool // 当該ハンドでVPIP済みかどうか
	pfrTracked     []bool // 当該ハンドでPFR済みかどうか
	handCount      int    // ハンド数 (トーナメントモード用)
	lastCpuError   error  // CPU行動エラーの最後のフォールバック記録 (テスト検出用)
}

// NewHoldem コンストラクタ
func NewHoldem(trumpCards *TrumpCards, players []*HoldemPlayer, config HoldemConfig) *Holdem {
	return &Holdem{
		trumpCards:     trumpCards,
		players:        players,
		communityCards: make([]*Card, 0),
		sidePots:       make([]HoldemSidePot, 0),
		actedFlags:     make([]bool, len(players)),
		roundResults:   make([]HoldemResult, 0),
		cpuActions:     make([]HoldemCpuAction, 0),
		startingChips:  make([]int, len(players)),
		vpipTracked:    make([]bool, len(players)),
		pfrTracked:     make([]bool, len(players)),
		config:         config,
		phase:          HoldemPhaseInit,
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

	h.trumpCards.Shuffle()
	for _, p := range h.players {
		p.Reset()
		p.SetFolded(false)
		p.SetAllIn(false)
		p.SetCurrentBet(0)
		p.handRank = 0
		p.bestHand = nil
		if p.GetChips() <= 0 {
			p.SetChips(h.config.InitChips)
		}
		p.IncrementTotalHands()
	}

	// HUDスタッツ追跡フラグをリセット
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))

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
	// プリフロップスタッツ追跡 (Holdem固有)
	h.trackPreFlopStats(playerIdx, action)

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
	// 4人プレイでは20アクションを超えることは稀。4ベッティングラウンドでも80アクション程度のため、200は十分な安全マージン。
	const maxIterations = 200
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
	switch h.config.BettingLimit {
	case BettingLimitPotLimit:
		maxRaises = holdemDefaultMaxRaises
		maxBetAmount = h.pot + h.lastBet
	case BettingLimitNoLimit:
		maxRaises = 0
		maxBetAmount = 0
	default: // Fixed
		maxRaises = holdemDefaultMaxRaises
		maxBetAmount = 0
	}
	return
}

// cpuDecide CPUプレイヤーの意思決定
func (h *Holdem) cpuDecide(idx int) (int, int) {
	p := h.players[idx]
	style := p.GetPlayStyle()
	callAmount := h.lastBet - p.GetCurrentBet()

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
