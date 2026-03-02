package domain

import (
	"fmt"
	"math/rand"
	"sort"
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

// アクション定数
const (
	HoldemActionFold  = 0 // フォールド
	HoldemActionCheck = 1 // チェック
	HoldemActionCall  = 2 // コール
	HoldemActionBet   = 3 // ベット
	HoldemActionRaise = 4 // レイズ
	HoldemActionAllIn = 5 // オールイン
)

// CPU AI 閾値
const holdemMaxRaisesPerRound = 4 // 1ラウンドの最大レイズ回数

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

// HoldemSidePot サイドポット
type HoldemSidePot struct {
	Amount          int   // ポット額
	EligiblePlayers []int // 受取対象プレイヤーインデックス
}

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
	if h.config.TournamentMode && h.handCount > 0 && h.handCount%h.config.BlindLevelHands == 0 {
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
	// VPIP: Call/Bet/Raise/AllIn → ハンドにつき1回のみ
	if !h.vpipTracked[playerIdx] {
		switch action {
		case HoldemActionCall, HoldemActionBet, HoldemActionRaise, HoldemActionAllIn:
			h.players[playerIdx].IncrementVPIP()
			h.vpipTracked[playerIdx] = true
		}
	}
	// PFR: Bet/Raise/AllIn → ハンドにつき1回のみ
	if !h.pfrTracked[playerIdx] {
		switch action {
		case HoldemActionBet, HoldemActionRaise, HoldemActionAllIn:
			h.players[playerIdx].IncrementPFR()
			h.pfrTracked[playerIdx] = true
		}
	}
}

// executeAction 指定プレイヤーのアクション実行
func (h *Holdem) executeAction(playerIdx, action, amount int) error {
	p := h.players[playerIdx]

	// プリフロップスタッツ追跡
	h.trackPreFlopStats(playerIdx, action)

	switch action {
	case HoldemActionFold:
		p.SetFolded(true)
		h.actedFlags[playerIdx] = true

	case HoldemActionCheck:
		if h.lastBet > p.GetCurrentBet() {
			return NewDomainError(ErrInvalidPlay, "Cannot check with outstanding bet.")
		}
		h.actedFlags[playerIdx] = true

	case HoldemActionCall:
		diff := h.lastBet - p.GetCurrentBet()
		if diff <= 0 {
			return NewDomainError(ErrInvalidPlay, "Nothing to call.")
		}
		if p.GetChips() <= diff {
			// オールイン (チップ不足)
			allInAmount := p.GetChips()
			p.SubtractChips(allInAmount)
			p.SetCurrentBet(p.GetCurrentBet() + allInAmount)
			h.pot += allInAmount
			p.SetAllIn(true)
		} else {
			p.SubtractChips(diff)
			p.SetCurrentBet(p.GetCurrentBet() + diff)
			h.pot += diff
		}
		h.actedFlags[playerIdx] = true

	case HoldemActionBet:
		if h.raiseCount >= holdemMaxRaisesPerRound {
			return NewDomainError(ErrInvalidPlay, "Maximum number of raises for this round has been reached.")
		}
		if h.lastBet > 0 {
			return NewDomainError(ErrInvalidPlay, "Cannot bet when there is an outstanding bet. Use raise.")
		}
		if amount < h.config.BigBlind {
			return NewDomainError(ErrInvalidAmount, "Bet must be at least the big blind.")
		}
		if amount > p.GetChips() {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
		}
		p.SubtractChips(amount)
		p.SetCurrentBet(p.GetCurrentBet() + amount)
		h.pot += amount
		h.lastBet = p.GetCurrentBet()
		h.minRaise = amount
		h.raiseCount++
		// 他の全員のactedフラグをリセット (ベットしたプレイヤー以外)
		h.resetActedExcept(playerIdx)
		if p.GetChips() == 0 {
			p.SetAllIn(true)
		}

	case HoldemActionRaise:
		if h.raiseCount >= holdemMaxRaisesPerRound {
			return NewDomainError(ErrInvalidPlay, "Maximum number of raises for this round has been reached.")
		}
		diff := h.lastBet - p.GetCurrentBet()
		if diff < 0 {
			diff = 0
		}
		if amount < h.minRaise {
			return NewDomainError(ErrInvalidAmount, "Raise must be at least the minimum raise.")
		}
		totalNeeded := diff + amount
		if totalNeeded >= p.GetChips() {
			return h.executeAction(playerIdx, HoldemActionAllIn, 0)
		}
		p.SubtractChips(totalNeeded)
		p.SetCurrentBet(p.GetCurrentBet() + totalNeeded)
		h.pot += totalNeeded
		h.lastBet = p.GetCurrentBet()
		h.minRaise = amount
		h.raiseCount++
		h.resetActedExcept(playerIdx)
		if p.GetChips() == 0 {
			p.SetAllIn(true)
		}

	case HoldemActionAllIn:
		allInAmount := p.GetChips()
		if allInAmount <= 0 {
			return NewDomainError(ErrInsufficientChips, "No chips to go all-in.")
		}
		p.SubtractChips(allInAmount)
		newBet := p.GetCurrentBet() + allInAmount
		p.SetCurrentBet(newBet)
		h.pot += allInAmount
		p.SetAllIn(true)
		if newBet > h.lastBet {
			raiseAmount := newBet - h.lastBet
			h.lastBet = newBet
			h.raiseCount++
			// ショートオールイン (最低レイズ額未満) はアクションを再開しない
			if raiseAmount >= h.minRaise {
				h.minRaise = raiseAmount
				h.resetActedExcept(playerIdx)
			} else {
				h.actedFlags[playerIdx] = true
			}
		} else {
			h.actedFlags[playerIdx] = true
		}

	default:
		return NewDomainError(ErrInvalidPlay, "Unknown action.")
	}

	// フォールドでアクティブプレイヤーが1人になったらチェック
	if h.countActivePlayers() == 1 {
		h.resolveLastPlayer()
		return nil
	}

	return nil
}

// resetActedExcept 指定プレイヤー以外のactedフラグをリセット (フォールド・オールイン除く)
func (h *Holdem) resetActedExcept(exceptIdx int) {
	for i := range h.actedFlags {
		if i == exceptIdx {
			h.actedFlags[i] = true
			continue
		}
		if h.players[i].GetFolded() || h.players[i].GetAllIn() {
			continue
		}
		h.actedFlags[i] = false
	}
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

	// サイドポット計算
	h.calculateSidePots()

	// 各ポットの配分
	h.roundResults = make([]HoldemResult, 0)
	wonAmounts := make(map[int]int)

	for _, sp := range h.sidePots {
		winners := h.findPotWinners(sp.EligiblePlayers)
		share := sp.Amount / len(winners)
		remainder := sp.Amount % len(winners)
		for i, wIdx := range winners {
			won := share
			if i == 0 {
				won += remainder
			}
			h.players[wIdx].AddChips(won)
			wonAmounts[wIdx] += won
		}
	}

	// 結果を構築
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

// calculateSidePots サイドポット計算
func (h *Holdem) calculateSidePots() {
	type playerContrib struct {
		idx    int
		amount int
	}

	// 各プレイヤーの合計投入額を計算
	contribs := make([]playerContrib, 0)
	for i, p := range h.players {
		// 投入額 = ハンド開始時チップ - 現在のチップ
		invested := h.startingChips[i] - p.GetChips()
		if invested < 0 {
			invested = 0
		}
		contribs = append(contribs, playerContrib{idx: i, amount: invested})
	}

	// オールインプレイヤーがいない場合はシンプルなメインポット
	hasAllIn := false
	for _, p := range h.players {
		if p.GetAllIn() && !p.GetFolded() {
			hasAllIn = true
			break
		}
	}

	if !hasAllIn {
		eligible := make([]int, 0)
		for i, p := range h.players {
			if !p.GetFolded() {
				eligible = append(eligible, i)
			}
		}
		h.sidePots = []HoldemSidePot{{Amount: h.pot, EligiblePlayers: eligible}}
		return
	}

	// オールインがある場合: 各オールイン額でポットを分割
	type allInLevel struct {
		amount int
		idx    int
	}
	levels := make([]allInLevel, 0)
	for _, c := range contribs {
		if h.players[c.idx].GetAllIn() && !h.players[c.idx].GetFolded() {
			levels = append(levels, allInLevel{amount: c.amount, idx: c.idx})
		}
	}
	// 投入額の昇順でソート
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].amount < levels[j].amount
	})

	h.sidePots = make([]HoldemSidePot, 0)
	prevLevel := 0
	remaining := h.pot

	for _, lv := range levels {
		if lv.amount <= prevLevel {
			continue
		}
		layerAmount := lv.amount - prevLevel
		potAmount := 0
		eligible := make([]int, 0)
		for _, c := range contribs {
			contribution := c.amount - prevLevel
			if contribution <= 0 {
				continue
			}
			if contribution > layerAmount {
				contribution = layerAmount
			}
			potAmount += contribution
			if !h.players[c.idx].GetFolded() {
				eligible = append(eligible, c.idx)
			}
		}
		if potAmount > 0 {
			h.sidePots = append(h.sidePots, HoldemSidePot{Amount: potAmount, EligiblePlayers: eligible})
			remaining -= potAmount
		}
		prevLevel = lv.amount
	}

	// 残りのポット (非オールインプレイヤー分)
	if remaining > 0 {
		eligible := make([]int, 0)
		for i, p := range h.players {
			if !p.GetFolded() && !p.GetAllIn() {
				eligible = append(eligible, i)
			}
		}
		if len(eligible) == 0 {
			// 全員オールインの場合は全未フォールドプレイヤーが対象
			for i, p := range h.players {
				if !p.GetFolded() {
					eligible = append(eligible, i)
				}
			}
		}
		h.sidePots = append(h.sidePots, HoldemSidePot{Amount: remaining, EligiblePlayers: eligible})
	}
}

// findPotWinners 対象プレイヤーから最強ハンドのプレイヤーを返す (複数ならスプリット)
func (h *Holdem) findPotWinners(eligible []int) []int {
	bestRank := -1
	var bestCards []*Card
	var winners []int

	for _, idx := range eligible {
		p := h.players[idx]
		if p.GetFolded() {
			continue
		}
		rank := p.GetHandRank()
		if rank > bestRank {
			bestRank = rank
			bestCards = p.GetBestHand()
			winners = []int{idx}
		} else if rank == bestRank {
			cmp := compareHighCardsSlice(p.GetBestHand(), bestCards)
			if cmp > 0 {
				bestCards = p.GetBestHand()
				winners = []int{idx}
			} else if cmp == 0 {
				winners = append(winners, idx)
			}
		}
	}

	return winners
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
			// CPUの決定したアクションが失敗した場合 (例: チップ不足でレイズ不可)、
			// エラーを記録し、フォールバック: コール額がある場合はフォールド、なければチェック。
			h.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", h.currentTurn, action, err)
			callAmt := h.lastBet - h.players[h.currentTurn].GetCurrentBet()
			if callAmt > 0 {
				h.executeAction(h.currentTurn, HoldemActionFold, 0)
			} else {
				h.executeAction(h.currentTurn, HoldemActionCheck, 0)
			}
		}
		if h.gameEndFlag {
			return nil
		}
		h.advanceTurn()
	}
	return nil
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

	// レイズ上限に達したら、レイズ/ベットをコール/チェックに変更
	if h.raiseCount >= holdemMaxRaisesPerRound {
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
	if callAmount > 0 {
		return HoldemActionFold, 0
	}
	return HoldemActionCheck, 0
}

// cpuCallOrCheck コール額がある場合はコール、なければチェック
func (h *Holdem) cpuCallOrCheck(callAmount int) (int, int) {
	if callAmount > 0 {
		return HoldemActionCall, 0
	}
	return HoldemActionCheck, 0
}

// cpuRaiseOrBet コール額がある場合はレイズ、なければベット (チップ不足時はオールイン)
func (h *Holdem) cpuRaiseOrBet(p *HoldemPlayer, callAmount, raiseAmt int) (int, int) {
	if raiseAmt > p.GetChips() {
		return HoldemActionAllIn, 0
	}
	if callAmount > 0 {
		if raiseAmt+callAmount > p.GetChips() {
			return HoldemActionAllIn, 0
		}
		return HoldemActionRaise, raiseAmt
	}
	return HoldemActionBet, raiseAmt
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
	if gap == 1 {
		score += 10
	} else if gap == 2 {
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

// --- テスト用セッター ---

// SetPhase フェーズ設定（テスト用）
func (h *Holdem) SetPhase(phase int) { h.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (h *Holdem) SetCurrentTurn(turn int) { h.currentTurn = turn }

// SetCommunityCards コミュニティカード設定（テスト用）
func (h *Holdem) SetCommunityCards(cards []*Card) { h.communityCards = cards }

// SetPot ポット設定（テスト用）
func (h *Holdem) SetPot(pot int) { h.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (h *Holdem) SetDealerIdx(idx int) { h.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (h *Holdem) SetGameEndFlag(flag bool) { h.gameEndFlag = flag }

// SetActedFlags actedフラグ設定（テスト用）
func (h *Holdem) SetActedFlags(flags []bool) { h.actedFlags = flags }

// SetLastBet 最後のベット設定（テスト用）
func (h *Holdem) SetLastBet(bet int) { h.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (h *Holdem) SetMinRaise(raise int) { h.minRaise = raise }

// SetRaiseCount レイズ回数設定（テスト用）
func (h *Holdem) SetRaiseCount(count int) { h.raiseCount = count }

// SetRoundResults ラウンド結果設定（テスト用）
func (h *Holdem) SetRoundResults(results []HoldemResult) { h.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (h *Holdem) SetCpuActions(actions []HoldemCpuAction) { h.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (h *Holdem) SetSidePots(pots []HoldemSidePot) { h.sidePots = pots }

// SetStartingChips ハンド開始時チップ設定（テスト用）
func (h *Holdem) SetStartingChips(chips []int) { h.startingChips = chips }

// GetStartingChips ハンド開始時チップ取得（テスト用）
func (h *Holdem) GetStartingChips() []int { return h.startingChips }

// GetHandCount ハンド数取得
func (h *Holdem) GetHandCount() int { return h.handCount }

// SetHandCount ハンド数設定（テスト用）
func (h *Holdem) SetHandCount(count int) { h.handCount = count }
