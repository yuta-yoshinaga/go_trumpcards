//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// DramahaHoleCards は配られるホールカードの枚数。**常に 5 枚。**
// Omaha の 4 枚と違い、この 5 枚がそのままドロー側の役になるので、枚数は
// 卓設定ではなくゲームの定義そのもの。
const DramahaHoleCards = 5

// ドラマハフェーズ定数 (Holdemと共通)
const (
	DramahaPhaseInit     = HoldemPhaseInit
	DramahaPhasePreFlop  = HoldemPhasePreFlop
	DramahaPhaseFlop     = HoldemPhaseFlop
	DramahaPhaseTurn     = HoldemPhaseTurn
	DramahaPhaseRiver    = HoldemPhaseRiver
	DramahaPhaseShowdown = HoldemPhaseShowdown
	DramahaPhaseEnd      = HoldemPhaseEnd
	DramahaPhaseRebuy    = HoldemPhaseRebuy

	// DramahaPhaseDraw はフロップのベッティングを終えた後のドローラウンド。
	//
	// **クローン元 (Omaha) には無いフェーズ。** Holdem 系の 0..7 と衝突しない
	// 値を使う —— これらの定数は Holdem のものを別名で参照しているだけなので、
	// 既存の値を踏むとフェーズ判定が別ゲームの分岐に落ちる。
	//
	// **フロップの後に置く。** 何を捨てるかはボードを見て決めるものなので、
	// フロップ前に引かせると判断材料が無く、ターン後だと引いた札を活かす
	// ベッティングラウンドが 1 つしか残らない。
	DramahaPhaseDraw = 8
)

// ドラマハアクション定数 (Holdemと共通)
const (
	DramahaActionFold  = HoldemActionFold
	DramahaActionCheck = HoldemActionCheck
	DramahaActionCall  = HoldemActionCall
	DramahaActionBet   = HoldemActionBet
	DramahaActionRaise = HoldemActionRaise
	DramahaActionAllIn = HoldemActionAllIn
)

// リバイフェーズ種別定数 (Holdemと共通)
const (
	DramahaRebuyPhaseNone  = HoldemRebuyPhaseNone
	DramahaRebuyPhaseRebuy = HoldemRebuyPhaseRebuy
	DramahaRebuyPhaseAddon = HoldemRebuyPhaseAddon
)

// Dramaha ドラマハホールデムクラス
//
// として動作する。ショーダウン時にハイハンド (既存の `EvalBestHand`) と
// 8 or Better のローハンド (`EvalBestLowHand`) を並行して評価し、
// 各サイドポットを 50/50 で分割する (奇数チップは Hi 側へ)。
// qualified なローが 1 人もいない場合はハイ側が全額獲得する。
type Dramaha struct {
	communityCardBettingBase
	trumpCards     *TrumpCards
	players        []*DramahaPlayer
	communityCards []*Card
	sidePots       []SidePot
	dealerIdx      int
	currentTurn    int
	phase          int
	holeCards      int // ホールカード配布枚数 (0 は既定の4枚扱い; Big O は5枚)
	// preflopCommunity はプリフロップ前に表向きにするコミュニティの枚数。
	//
	// **0 が通常のドラマハ。** Courchevel だけが 1 で、フロップの 1 枚目を
	// 賭ける前に見せる ── 公開する枚数を前倒しするだけで、総枚数も役の
	// 作り方も変わらない。
	preflopCommunity int
	config           DramahaConfig
	roundResults     []HoldemResult
	cpuActions       []HoldemCpuAction
	startingChips    []int
	vpipTracked      []bool
	pfrTracked       []bool
	threeBetTracked  []bool
	// drawnFlags はドローラウンドで交換を済ませた席。降りた席は最初から true。
	drawnFlags     []bool
	tournamentBase // handCount / rebuyCounts / addonUsed (issue #1463)
	lastCpuError   error
	rebuyPhaseType int
	actionLogBase
	humanProfile    *BettingHumanProfile
	lastHumanPlayMs int
}

// NewDramaha コンストラクタ
func NewDramaha(trumpCards *TrumpCards, players []*DramahaPlayer, config DramahaConfig) *Dramaha {
	o := &Dramaha{
		communityCardBettingBase: communityCardBettingBase{
			actedFlags: make([]bool, len(players)),
		},
		trumpCards:      trumpCards,
		players:         players,
		communityCards:  make([]*Card, 0),
		sidePots:        make([]SidePot, 0),
		roundResults:    make([]HoldemResult, 0),
		cpuActions:      make([]HoldemCpuAction, 0),
		startingChips:   make([]int, len(players)),
		vpipTracked:     make([]bool, len(players)),
		pfrTracked:      make([]bool, len(players)),
		threeBetTracked: make([]bool, len(players)),
		config:          config,
		phase:           DramahaPhaseInit,
		holeCards:       DramahaHoleCards,
	}
	o.initTournamentState(len(players))
	return o
}

// NewDefaultDramaha returns Dramaha with the default table size and DefaultDramahaConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultDramaha() *Dramaha {
	cfg := DefaultDramahaConfig()
	return NewDramaha(NewTrumpCards(0), NewDramahaPlayersForTable(cfg.TableSize), cfg)
}

// holeCardCount はホールカード配布枚数を返す。未設定 (0) のゲーム
// (既存のドラマハや古いシリアライズ状態) は既定の4枚として扱う。
// holeCardCount は 1 席に配るホールカードの枚数。
//
// **ドラマハは常に 5 枚。** クローン元 (Omaha) は 4 枚が既定で Big O だけが
// 5 枚だったが、こちらは 5 枚がそのままドロー側の役になるので、枚数を卓設定に
// 委ねる余地が無い。**未設定 (0) を 4 に落とすと、復元した卓だけ 4 枚配りに
// なってドロー側の評価が丸ごと成立しなくなる。**
func (o *Dramaha) holeCardCount() int {
	return DramahaHoleCards
}

// GetHoleCardCount はホールカード配布枚数を返す (Big O では5)。
func (o *Dramaha) GetHoleCardCount() int { return o.holeCardCount() }

// Reset ゲーム初期化
func (o *Dramaha) Reset() error {
	o.phase = DramahaPhaseInit
	o.pot = 0
	o.sidePots = make([]SidePot, 0)
	o.communityCards = make([]*Card, 0)
	o.gameEndFlag = false
	o.lastBet = 0
	o.minRaise = o.config.BigBlind
	o.raiseCount = 0
	o.actedFlags = make([]bool, len(o.players))
	o.roundResults = make([]HoldemResult, 0)
	o.cpuActions = make([]HoldemCpuAction, 0)
	o.rebuyPhaseType = DramahaRebuyPhaseNone
	o.actionLog = nil
	o.lastHumanPlayMs = 0

	// メタAI: プロファイル初期化
	if o.config.CpuMetaAI {
		if o.humanProfile != nil {
			o.humanProfile.GamesPlayed++
		} else {
			o.humanProfile = &BettingHumanProfile{}
		}
	}

	o.trumpCards.Shuffle()
	for _, p := range o.players {
		p.Reset()
		p.SetFolded(false)
		p.SetAllIn(false)
		p.SetCurrentBet(0)
		p.handRank = 0
		p.bestHand = nil
		p.drawBestHand = nil
		p.drawRank = PokerHandHighCard
		if p.GetChips() <= 0 && !o.config.RebuyEnabled {
			p.SetChips(o.config.InitChips)
		}
		p.IncrementTotalHands()
	}

	o.vpipTracked = make([]bool, len(o.players))
	o.pfrTracked = make([]bool, len(o.players))
	o.threeBetTracked = make([]bool, len(o.players))

	if o.config.TournamentMode && o.config.BlindLevelHands > 0 && o.handCount > 0 && o.handCount%o.config.BlindLevelHands == 0 {
		o.config.SmallBlind = o.config.SmallBlind * o.config.BlindMultiplier / 100
		o.config.BigBlind = o.config.BigBlind * o.config.BlindMultiplier / 100
		if o.config.SmallBlind < 1 {
			o.config.SmallBlind = 1
		}
		if o.config.BigBlind < 2 {
			o.config.BigBlind = 2
		}
	}
	o.handCount++

	if o.config.RebuyEnabled && o.handCount <= o.config.RebuyPeriodHands {
		needHumanRebuy := false
		for i, p := range o.players {
			if p.GetChips() <= 0 && o.rebuyCounts[i] < o.config.RebuyMaxCount {
				if p.GetIsHuman() {
					needHumanRebuy = true
				} else {
					p.AddChips(o.config.RebuyChips)
					o.rebuyCounts[i]++
				}
			}
		}
		if needHumanRebuy {
			o.phase = DramahaPhaseRebuy
			o.rebuyPhaseType = DramahaRebuyPhaseRebuy
			return nil
		}
	}

	if o.checkAndTransitionAddon() {
		return nil
	}

	return o.continueReset()
}

// continueReset ディール以降のリセット処理
func (o *Dramaha) continueReset() error {
	o.startingChips = make([]int, len(o.players))
	for i, p := range o.players {
		o.startingChips[i] = p.GetChips()
	}

	o.postBlinds()

	// ホールカード配布 (ドラマハ=4枚, Big O=5枚)
	for i := 0; i < o.holeCardCount(); i++ {
		for j := 0; j < len(o.players); j++ {
			idx := (o.dealerIdx + 1 + j) % len(o.players)
			card := o.trumpCards.DrawCard()
			if card != nil {
				o.players[idx].AddCard(card)
			}
		}
	}

	// **フロップの 1 枚目を先に見せるバリアントがある (Courchevel)。**
	// 既定は 0 枚なので、通常のドラマハと Big O の進行は変わらない。
	for i := 0; i < o.preflopCommunity; i++ {
		if card := o.trumpCards.DrawCard(); card != nil {
			o.communityCards = append(o.communityCards, card)
		}
	}
	if o.preflopCommunity > 0 {
		o.appendLog(-1, "deal", "exposed the first flop card", o.communityCards)
	}

	o.phase = DramahaPhasePreFlop
	o.currentTurn = (o.dealerIdx + 3) % len(o.players)

	if err := o.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postBlinds ブラインド投入
func (o *Dramaha) postBlinds() {
	postBlindsFor(o.players, o.dealerIdx, o.config.SmallBlind, o.config.BigBlind, &o.pot, &o.lastBet, o.actedFlags, o)
}

// PlayerAction 人間プレイヤーのアクション実行
// humanPlayMs: 迷い時間(ms, 0=計測なし)
func (o *Dramaha) PlayerAction(action, amount, humanPlayMs int) error {
	if o.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if o.phase < DramahaPhasePreFlop || o.phase > DramahaPhaseRiver {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !o.players[o.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// メタAI: 人間アクション記録
	o.lastHumanPlayMs = humanPlayMs
	if o.config.CpuMetaAI && o.humanProfile != nil {
		pl := o.players[o.currentTurn]
		handRank := pl.EvalBestHand(o.communityCards)
		o.humanProfile.RecordAction(handRank, action)
		o.humanProfile.RecordHesitation(humanPlayMs)
		if o.lastBet > pl.GetCurrentBet() {
			o.humanProfile.RecordFoldToBet(action == DramahaActionFold)
		}
	}

	err := o.executeAction(o.currentTurn, action, amount)
	if err != nil {
		return err
	}

	o.advanceTurn()
	return o.runCpuActions()
}

// trackPreFlopStats プリフロップのHUDスタッツを追跡
func (o *Dramaha) trackPreFlopStats(playerIdx, action int) {
	if o.phase != DramahaPhasePreFlop {
		return
	}

	isVPIPAction := false
	isPFRAction := false

	switch action {
	case DramahaActionCall:
		isVPIPAction = true
	case DramahaActionBet, DramahaActionRaise, DramahaActionAllIn:
		isVPIPAction = true
		isPFRAction = true
	}

	if isVPIPAction && !o.vpipTracked[playerIdx] {
		o.players[playerIdx].IncrementVPIP()
		o.vpipTracked[playerIdx] = true
	}

	if isPFRAction && !o.pfrTracked[playerIdx] {
		o.players[playerIdx].IncrementPFR()
		o.pfrTracked[playerIdx] = true
	}

	if o.raiseCount >= 1 && !o.threeBetTracked[playerIdx] {
		o.players[playerIdx].IncrementThreeBetOpportunity()
		if action == DramahaActionRaise || action == DramahaActionAllIn {
			o.players[playerIdx].IncrementThreeBet()
		}
		o.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats ポストフロップのAFスタッツを追跡
func (o *Dramaha) trackPostFlopStats(playerIdx, action int) {
	if o.phase < DramahaPhaseFlop || o.phase > DramahaPhaseRiver {
		return
	}

	switch action {
	case DramahaActionBet, DramahaActionRaise, DramahaActionAllIn:
		o.players[playerIdx].IncrementPostFlopBetRaise()
	case DramahaActionCall:
		o.players[playerIdx].IncrementPostFlopCall()
	}
}

// executeAction 指定プレイヤーのアクション実行
func (o *Dramaha) executeAction(playerIdx, action, amount int) error {
	o.trackPreFlopStats(playerIdx, action)
	o.trackPostFlopStats(playerIdx, action)

	bp := toBettingPlayers(o.players)
	state := o.bettingState()
	maxRaises, maxBetAmount := o.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, o.config.BigBlind, maxRaises, maxBetAmount)
	o.syncBettingState(state)
	if err != nil {
		return err
	}

	o.logAction(playerIdx, action, amount)

	if o.countActivePlayers() == 1 {
		o.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (o *Dramaha) advanceTurn() {
	if o.gameEndFlag {
		return
	}

	bp := toBettingPlayers(o.players)
	if o.isBettingRoundComplete(bp) {
		o.advancePhase()
		return
	}

	if next := o.findNextActiveTurn(o.currentTurn, bp); next >= 0 {
		o.currentTurn = next
		return
	}

	o.advancePhase()
}

// advancePhase 次のフェーズに進める
// Draw は指定した位置のホールカードを捨てて引き直す。
//
// **1 ラウンドに 1 回だけ。** 何度も引けると、ドロー側の役はチップの続く限り
// 作り直せてしまい、二分されたポットの片側が事実上の掛け金勝負になる。
// 空の indices は「交換しない」。
func (o *Dramaha) Draw(playerIdx int, indices []int) error {
	if o.phase != DramahaPhaseDraw {
		return NewDomainError(ErrWrongPhase, "Draw is only allowed during the draw round.")
	}
	if playerIdx < 0 || playerIdx >= len(o.players) {
		return NewDomainError(ErrInvalidPlay, "No such seat.")
	}
	if o.drawnFlags[playerIdx] {
		return NewDomainError(ErrInvalidPlay, "This seat has already drawn.")
	}
	p := o.players[playerIdx]
	if len(indices) > DramahaHoleCards {
		return NewDomainError(ErrInvalidPlay, "Too many cards to exchange.")
	}
	seen := make(map[int]bool, len(indices))
	for _, i := range indices {
		if i < 0 || i >= p.GetCardsSize() {
			return NewDomainError(ErrInvalidPlay, "Card index out of range.")
		}
		if seen[i] {
			return NewDomainError(ErrInvalidPlay, "The same card cannot be exchanged twice.")
		}
		seen[i] = true
	}

	for _, i := range indices {
		// **山が尽きたら替えない。** nil を手札に入れると役の評価が落ちる。
		if c := o.trumpCards.DrawCard(); c != nil {
			p.ReplaceCard(i, c)
		}
	}
	o.drawnFlags[playerIdx] = true
	o.appendLog(playerIdx, "draw", fmt.Sprintf("exchanged %d card(s)", len(indices)), nil)

	if o.allDrawn() {
		o.advancePhase()
		// **CPU を動かすのを忘れない。** PlayerAction は最後に runCpuActions を
		// 呼ぶが、Draw は advancePhase までしかやっていなかった。ターンに入った
		// 直後の手番は CPU なので、誰も動かさないまま人間の入力を待ち、
		// 「あなたの番ではありません」から永久に抜けられなくなる。
		return o.runCpuActions()
	}
	return nil
}

// allDrawn は全員がドローを終えたかを返す。
func (o *Dramaha) allDrawn() bool {
	for _, done := range o.drawnFlags {
		if !done {
			return false
		}
	}
	return true
}

// autoDrawForCPUs は CPU 席のドローを自動で済ませる。
func (o *Dramaha) autoDrawForCPUs() {
	for i, p := range o.players {
		if o.drawnFlags[i] || p.GetIsHuman() {
			continue
		}
		_ = o.Draw(i, dramahaCPUDiscards(p.HoleCardsCopy()))
	}
}

// dramahaCPUDiscards は CPU が捨てる位置を返す。
//
// **役に絡まない札だけを捨てる。** ペア以上を構成している札は残し、何も無ければ
// 高い 2 枚を残して 3 枚引く —— 5 枚を丸ごと引き直すと、噛み合っている札まで
// 手放すことになる。
func dramahaCPUDiscards(cards []*Card) []int {
	if len(cards) != DramahaHoleCards {
		return nil
	}
	freq := make(map[int]int, DramahaHoleCards)
	for _, c := range cards {
		if c != nil {
			freq[c.GetValue()]++
		}
	}
	var discards []int
	for i, c := range cards {
		if c != nil && freq[c.GetValue()] >= 2 {
			continue
		}
		discards = append(discards, i)
	}
	if len(discards) == DramahaHoleCards {
		ranked := make([]int, DramahaHoleCards)
		for i := range ranked {
			ranked[i] = i
		}
		sort.SliceStable(ranked, func(a, b int) bool {
			return dramahaCardRank(cards[ranked[a]]) > dramahaCardRank(cards[ranked[b]])
		})
		discards = append([]int(nil), ranked[2:]...)
		sort.Ints(discards)
	}
	return discards
}

// dramahaCardRank は A を 14 として扱った札の高さ。
func dramahaCardRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return 14
	}
	return c.GetValue()
}

func (o *Dramaha) advancePhase() {
	for _, p := range o.players {
		p.SetCurrentBet(0)
	}
	o.lastBet = 0
	o.minRaise = o.config.BigBlind
	o.raiseCount = 0
	o.actedFlags = make([]bool, len(o.players))
	for i, p := range o.players {
		if p.GetFolded() || p.GetAllIn() {
			o.actedFlags[i] = true
		}
	}

	switch o.phase {
	case DramahaPhasePreFlop:
		o.phase = DramahaPhaseFlop
		// **フロップは「3 枚まで足す」。** 先に見せた枚数を引かずに 3 枚
		// 配ると、Courchevel の場だけ 6 枚になって役の作り方が変わる。
		for i := len(o.communityCards); i < 3; i++ {
			card := o.trumpCards.DrawCard()
			if card != nil {
				o.communityCards = append(o.communityCards, card)
			}
		}
		o.appendLog(-1, "deal", "dealt flop", o.communityCards)
	case DramahaPhaseFlop:
		// フロップのベッティングが終わったらドローラウンドへ。ターンはその後。
		o.phase = DramahaPhaseDraw
		o.drawnFlags = make([]bool, len(o.players))
		for i, p := range o.players {
			if p.GetFolded() {
				o.drawnFlags[i] = true
			}
		}
		o.appendLog(-1, "draw", "draw round", nil)
		// **ここで return する。** autoDrawForCPUs は最後の CPU が引いた時点で
		// Draw() 経由で advancePhase を**入れ子に**呼び、ターンを配って下の
		// activeCnt ブロックまで走り切る。その後この外側のフレームが switch を
		// 抜けて同じブロックをもう一度実行すると、resolveShowdown が二度走って
		// **ポットが二重に配られる** —— 実測で卓上の総額が 4000 → 4015 に増えた。
		// フェーズを進めた側のフレームだけが後片付けをする。
		o.autoDrawForCPUs()
		// **全員オールインでも人間のドローは飛ばさない。** 下の activeCnt <= 1
		// の短絡は残りのボードを配って決着させてしまうが、そこへ落ちると人間
		// だけ引かないままショーダウンに行く (CPU は autoDrawForCPUs で引き
		// 終えている)。ドローが済んでいなければ、ここで人間を待つ。
		return
	case DramahaPhaseDraw:
		o.phase = DramahaPhaseTurn
		card := o.trumpCards.DrawCard()
		if card != nil {
			o.communityCards = append(o.communityCards, card)
		}
		o.appendLog(-1, "deal", "dealt turn", o.communityCards[3:])
	case DramahaPhaseTurn:
		o.phase = DramahaPhaseRiver
		card := o.trumpCards.DrawCard()
		if card != nil {
			o.communityCards = append(o.communityCards, card)
		}
		o.appendLog(-1, "deal", "dealt river", o.communityCards[4:])
	case DramahaPhaseRiver:
		o.phase = DramahaPhaseShowdown
		o.appendLog(-1, "showdown", "showdown", nil)
		o.resolveShowdown()
		return
	}

	activeCnt := 0
	for _, p := range o.players {
		if !p.GetFolded() && !p.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		o.dealRemainingCommunity()
		o.phase = DramahaPhaseShowdown
		o.resolveShowdown()
		return
	}

	o.currentTurn = o.findNextActive(o.dealerIdx)
}

// dealRemainingCommunity 残りのコミュニティカードを全て配る
func (o *Dramaha) dealRemainingCommunity() {
	// **走り切る前にドローを済ませる。** 全員オールインだとベッティングが
	// 成立しないのでボードを一気に配って決着させるが、**ドローは賭けではない**
	// —— 引かないままショーダウンに行くと、ポットの半分 (ドロー側) を「誰も
	// 交換していない手」で決めることになる。オールインの席は操作できないので、
	// 人間ぶんも CPU と同じ方針で自動的に引く。
	o.runOutDraw()
	dealUpTo(&o.communityCards, o.trumpCards, 5)
}

// runOutDraw は走り切りのときに、まだ引いていない席のドローを自動で済ませる。
func (o *Dramaha) runOutDraw() {
	if len(o.drawnFlags) != len(o.players) {
		o.drawnFlags = make([]bool, len(o.players))
	}
	for i, p := range o.players {
		if o.drawnFlags[i] || p.GetFolded() {
			o.drawnFlags[i] = true
			continue
		}
		for _, idx := range dramahaCPUDiscards(p.HoleCardsCopy()) {
			if c := o.trumpCards.DrawCard(); c != nil {
				p.ReplaceCard(idx, c)
			}
		}
		o.drawnFlags[i] = true
	}
}

// findNextActive 指定インデックスの次のアクティブプレイヤーを探す
func (o *Dramaha) findNextActive(fromIdx int) int {
	return findNextActive(o.players, fromIdx)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (o *Dramaha) countActivePlayers() int {
	return countPlayers(o.players, func(p *DramahaPlayer) bool { return !p.GetFolded() })
}

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (o *Dramaha) resolveLastPlayer() {
	for i, p := range o.players {
		if !p.GetFolded() {
			p.AddChips(o.pot)
			o.roundResults = []HoldemResult{{
				PlayerIdx: i,
				WonAmount: o.pot,
			}}
			o.pot = 0
			break
		}
	}
	o.phase = DramahaPhaseEnd
	o.gameEndFlag = true
	o.dealerIdx = (o.dealerIdx + 1) % len(o.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (o *Dramaha) resolveShowdown() {
	for _, p := range o.players {
		if !p.GetFolded() {
			p.EvalBestHand(o.communityCards)
			// **ポットは必ず二分される。** 片方だけ評価する分岐は無い ——
			// Omaha 側とドロー側は常に両方成立する。
			p.EvalDrawHand()
		}
	}

	bp := toBettingPlayers(o.players)
	o.sidePots = CalculateSidePots(bp, o.pot, o.startingChips)

	omahaAmounts, drawAmounts := o.distributeDramahaPots(bp)

	o.roundResults = make([]HoldemResult, 0)
	humanLost := false
	for i, p := range o.players {
		if p.GetFolded() {
			continue
		}
		hi := omahaAmounts[i]
		lo := drawAmounts[i]
		result := HoldemResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  o.getHandName(p.GetHandRank()),
			BestHand:  p.GetBestHand(),
			Kickers:   ExtractKickers(p.GetBestHand(), p.GetHandRank()),
			WonAmount: hi + lo,
		}
		result.HiWonAmount = hi
		result.LowWonAmount = lo
		result.LowQualifies = true // ドロー側は 5 枚あれば必ず成立する
		result.LowBestHand = p.GetDrawBestHand()
		o.roundResults = append(o.roundResults, result)
		if p.GetIsHuman() && (hi+lo) == 0 {
			humanLost = true
		}
	}

	if humanLost {
		return
	}

	o.finalizeShowdown()
}

// distributeDramahaPots は各サイドポットを Omaha 側 / ドロー側で 50:50 に
// 分配する。
//
// **ドロー側は必ず勝者が居る。** クローン元の Hi-Lo は「8 or Better のローが
// 成立しなければハイ側が総取り」だったが、5 枚あればどんな手でも役として順位が
// 付くので、その分岐は起きない。半分ずつが常に出ていく。
// 奇数チップは Omaha 側に寄せる (ポーカー慣例)。
func (o *Dramaha) distributeDramahaPots(bp []BettingPlayer) (omaha, draw map[int]int) {
	omaha = make(map[int]int)
	draw = make(map[int]int)
	for _, sp := range o.sidePots {
		omahaWinners := FindPotWinners(bp, sp.EligiblePlayers)
		if len(omahaWinners) == 0 {
			continue
		}
		drawWinners := o.findDramahaDrawWinners(sp.EligiblePlayers)

		drawPot := sp.Amount / 2
		omahaPot := sp.Amount - drawPot // 奇数チップは Omaha 側
		if len(drawWinners) == 0 {
			// 到達しない想定 (5 枚あれば必ず役が付く) だが、降りた席しか
			// 残っていないような盤面で 0 人になったら、チップを卓に置き去りに
			// せず Omaha 側へ寄せる。
			omahaPot = sp.Amount
			drawPot = 0
		}

		distributeAmongWinners(bp, omahaWinners, omahaPot, omaha)
		distributeAmongWinners(bp, drawWinners, drawPot, draw)
	}
	return omaha, draw
}

// findDramahaDrawWinners は対象プレイヤーの中でドロー側の役が最も強い者を
// 返す。同点はスプリット。
//
// **比較は役位が先、同位なら札の高さ。** 役位だけで決めると、同じ「ワンペア」
// 同士でエースのペアと 2 のペアが引き分けになる。
func (o *Dramaha) findDramahaDrawWinners(eligible []int) []int {
	var winners []int
	bestRank := -1
	var bestCards []*Card
	for _, idx := range eligible {
		p := o.players[idx]
		if p.GetFolded() {
			continue
		}
		rank, cards := p.GetDrawRank(), p.GetDrawBestHand()
		if cards == nil {
			continue
		}
		switch {
		case rank > bestRank:
			winners = []int{idx}
			bestRank, bestCards = rank, cards
		case rank == bestRank:
			switch cmp := compareHighCardsSlice(cards, bestCards); {
			case cmp > 0:
				winners = []int{idx}
				bestCards = cards
			case cmp == 0:
				winners = append(winners, idx)
			}
		}
	}
	return winners
}

// finalizeShowdown ショーダウンを完了し、END フェーズに遷移する
func (o *Dramaha) finalizeShowdown() {
	// **配り終えたポットは 0 にする。** 理由は Holdem 側と同じ。
	o.pot = 0
	o.phase = DramahaPhaseEnd
	o.gameEndFlag = true
	o.dealerIdx = (o.dealerIdx + 1) % len(o.players)
}

// Muck 人間プレイヤーがハンドをマックする
func (o *Dramaha) Muck() error {
	if o.phase != DramahaPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Muck is not available now.")
	}
	for i := range o.roundResults {
		if o.players[o.roundResults[i].PlayerIdx].GetIsHuman() {
			o.roundResults[i].Mucked = true
			break
		}
	}
	o.finalizeShowdown()
	return nil
}

// ShowHand 人間プレイヤーがハンドを公開する
func (o *Dramaha) ShowHand() error {
	if o.phase != DramahaPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	o.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (o *Dramaha) IsMuckAvailable() bool {
	if o.phase != DramahaPhaseShowdown {
		return false
	}
	for _, r := range o.roundResults {
		if o.players[r.PlayerIdx].GetIsHuman() && r.WonAmount == 0 {
			return true
		}
	}
	return false
}

// getHandName ハンドランクから名前を返す
func (o *Dramaha) getHandName(rank int) string {
	return pokerHandName(rank)
}

// runCpuActions CPUプレイヤーのアクションを実行
func (o *Dramaha) runCpuActions() error {
	if o.gameEndFlag {
		return nil
	}
	const maxIterations = 500
	iterations := 0
	// **ドローフェーズはこのループの対象外。** ベッティングのラウンドではない
	// ので `PreFlop..River` の範囲に入らず (Draw = 8)、ここで抜けるのが正しい
	// —— 抜けた先で人間のドローを待つ。CPU のドローは advancePhase が
	// autoDrawForCPUs で先に済ませてある。
	for !o.gameEndFlag && o.phase >= DramahaPhasePreFlop && o.phase <= DramahaPhaseRiver {
		iterations++
		if iterations > maxIterations {
			return fmt.Errorf("maxIterations reached in runCpuActions, possible infinite loop")
		}
		if o.players[o.currentTurn].GetIsHuman() {
			return nil
		}
		if o.players[o.currentTurn].GetFolded() || o.players[o.currentTurn].GetAllIn() {
			o.advanceTurn()
			continue
		}
		action, amount := o.cpuDecide(o.currentTurn)
		o.cpuActions = append(o.cpuActions, HoldemCpuAction{
			PlayerIdx: o.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := o.executeAction(o.currentTurn, action, amount)
		if err != nil {
			o.handleCpuActionError(o.currentTurn, action, err)
		}
		if o.gameEndFlag {
			return nil
		}
		o.advanceTurn()
	}
	return nil
}

// handleCpuActionError CPUアクション失敗時のフォールバック処理
func (o *Dramaha) handleCpuActionError(playerIdx, action int, err error) {
	o.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", playerIdx, action, err)
	callAmt := o.lastBet - o.players[playerIdx].GetCurrentBet()
	if callAmt > 0 {
		_ = o.executeAction(playerIdx, DramahaActionFold, 0)
	} else {
		_ = o.executeAction(playerIdx, DramahaActionCheck, 0)
	}
}

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (o *Dramaha) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(o.config.BettingLimit, o.pot, o.lastBet)
}

// cpuDecide CPUプレイヤーの意思決定
func (o *Dramaha) cpuDecide(idx int) (int, int) {
	p := o.players[idx]
	style := p.GetPlayStyle()
	callAmount := o.lastBet - p.GetCurrentBet()

	// GTO
	if style == HoldemStyleGTO {
		var action, amount int
		if o.phase == DramahaPhasePreFlop {
			action, amount = o.cpuDecidePreFlopGTO(idx, callAmount)
		} else {
			action, amount = o.cpuDecidePostFlopGTO(idx, callAmount)
		}
		maxRaises, maxBetAmount := o.bettingLimits()
		if maxBetAmount > 0 && amount > maxBetAmount {
			amount = maxBetAmount
		}
		if maxRaises > 0 && o.raiseCount >= maxRaises {
			if action == DramahaActionRaise || action == DramahaActionBet {
				if callAmount > 0 {
					return DramahaActionCall, 0
				}
				return DramahaActionCheck, 0
			}
		}
		return action, amount
	}

	params, ok := holdemStyleParamsMap[style]
	if !ok {
		return o.cpuCallOrCheck(callAmount)
	}

	// メタAI: ブラフ率を調整
	if o.config.CpuMetaAI && o.humanProfile != nil {
		adjusted := o.humanProfile.AdjustedBluffChance(float64(params.bluffRate))
		params.bluffRate = int(math.Round(adjusted))
	}

	var action, amount int
	if o.phase == DramahaPhasePreFlop {
		action, amount = o.cpuDecidePreFlop(idx, params, callAmount)
	} else {
		action, amount = o.cpuDecidePostFlop(idx, params, callAmount)
	}

	// メタAI: 人間のベット/レイズに対してコール確率を調整
	if o.config.CpuMetaAI && o.humanProfile != nil && o.lastHumanPlayMs > 0 {
		if action == DramahaActionFold && callAmount > 0 {
			handRank := p.EvalBestHand(o.communityCards)
			bracket := bettingHandBracket(handRank)
			adjustedCall := o.humanProfile.AdjustedCallChance(0.0, bracket, o.lastHumanPlayMs)
			if adjustedCall > 0 && rand.Float64() < adjustedCall {
				action = DramahaActionCall
				amount = 0
			}
		}
	}

	maxRaises, maxBetAmount := o.bettingLimits()

	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}

	if maxRaises > 0 && o.raiseCount >= maxRaises {
		if action == DramahaActionRaise || action == DramahaActionBet {
			if callAmount > 0 {
				return DramahaActionCall, 0
			}
			return DramahaActionCheck, 0
		}
	}
	return action, amount
}

// cpuFoldOrCheck コール額がある場合はフォールド、なければチェック
func (o *Dramaha) cpuFoldOrCheck(callAmount int) (int, int) {
	return CpuFoldOrCheck(callAmount)
}

// cpuCallOrCheck コール額がある場合はコール、なければチェック
func (o *Dramaha) cpuCallOrCheck(callAmount int) (int, int) {
	return CpuCallOrCheck(callAmount)
}

// cpuRaiseOrBet コール額がある場合はレイズ、なければベット
func (o *Dramaha) cpuRaiseOrBet(p *DramahaPlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(p.GetChips(), callAmount, raiseAmt)
}

// cpuBetOrAllIn ベットする (チップ不足時はオールイン)
func (o *Dramaha) cpuBetOrAllIn(p *DramahaPlayer, betAmt int) (int, int) {
	if betAmt > p.GetChips() {
		return DramahaActionAllIn, 0
	}
	return DramahaActionBet, betAmt
}

// cpuPotBet ポット比率ベースのベット額を計算
func (o *Dramaha) cpuPotBet(potPct int) int {
	return potBet(o.pot, potPct, o.config.BigBlind, o.minRaise)
}

// cpuDecidePreFlop プリフロップのCPU意思決定
func (o *Dramaha) cpuDecidePreFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := o.players[idx]
	strength := o.evalPreFlopStrength(idx)

	if strength < params.preFlopFoldThreshold {
		if params.preFlopFoldCompound {
			if callAmount > o.config.BigBlind*params.preFlopFoldCallMult {
				return DramahaActionFold, 0
			}
		} else {
			return o.cpuFoldOrCheck(callAmount)
		}
	}

	if params.aggressive {
		if strength >= params.preFlopRaiseThreshold || rand.Intn(100) < params.bluffRate {
			return o.cpuRaiseOrBet(p, callAmount, o.cpuPotBet(params.preFlopRaisePotPct))
		}
		return o.cpuCallOrCheck(callAmount)
	}

	if callAmount > 0 {
		return DramahaActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return o.cpuBetOrAllIn(p, o.cpuPotBet(params.preFlopBluffPotPct))
	}
	return DramahaActionCheck, 0
}

// cpuDecidePostFlop フロップ以降のCPU意思決定
func (o *Dramaha) cpuDecidePostFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := o.players[idx]
	handRank := p.EvalBestHand(o.communityCards)

	if params.aggressive {
		if handRank >= params.postFlopRaiseRank || rand.Intn(100) < params.bluffRate {
			return o.cpuRaiseOrBet(p, callAmount, o.cpuPotBet(params.postFlopRaisePotPct))
		}
		if params.postFlopFallbackFold {
			if params.postFlopCondCallRank >= 0 && handRank >= params.postFlopCondCallRank && callAmount > 0 {
				return DramahaActionCall, 0
			}
			return o.cpuFoldOrCheck(callAmount)
		}
		if callAmount > 0 {
			if handRank <= params.postFlopAggrFoldRank && callAmount > o.config.BigBlind*params.postFlopAggrFoldMult {
				return DramahaActionFold, 0
			}
			return DramahaActionCall, 0
		}
		return DramahaActionCheck, 0
	}

	if callAmount > 0 {
		if handRank <= params.postFlopPassFoldRank {
			if params.postFlopPassFoldMult < 0 || callAmount > o.config.BigBlind*params.postFlopPassFoldMult {
				return DramahaActionFold, 0
			}
		}
		return DramahaActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return o.cpuBetOrAllIn(p, o.cpuPotBet(params.postFlopBluffPotPct))
	}
	return DramahaActionCheck, 0
}

// cpuDecidePreFlopGTO GTOプリフロップ意思決定
func (o *Dramaha) cpuDecidePreFlopGTO(idx, callAmount int) (int, int) {
	p := o.players[idx]
	strength := o.evalPreFlopStrength(idx)
	dist := gtoPreFlopTable[gtoPreFlopIndex(strength)]
	decision := gtoRollAction(dist)

	switch decision {
	case 0:
		return o.cpuFoldOrCheck(callAmount)
	case 2:
		betAmt := o.cpuPotBet(gtoPreFlopBetPct)
		return o.cpuRaiseOrBet(p, callAmount, betAmt)
	default:
		return o.cpuCallOrCheck(callAmount)
	}
}

// cpuDecidePostFlopGTO GTOポストフロップ意思決定
func (o *Dramaha) cpuDecidePostFlopGTO(idx, callAmount int) (int, int) {
	p := o.players[idx]
	handRank := p.EvalBestHand(o.communityCards)
	category := classifyGTOHand(handRank)
	bt := evalBoardTexture(o.communityCards)

	wetIdx := 0
	if bt.wet {
		wetIdx = 1
	}
	dist := gtoPostFlopTable[category][wetIdx]
	decision := gtoRollAction(dist)

	potPct := gtoDryBoardBetPct
	if bt.wet {
		potPct = gtoWetBoardBetPct
	}

	if bt.paired && decision == 2 && category <= gtoHandMedium && rand.Intn(100) < 30 {
		return o.cpuCallOrCheck(callAmount)
	}

	if bt.highCards >= 3 && decision == 2 && category <= gtoHandWeak && rand.Intn(100) < 40 {
		return o.cpuCallOrCheck(callAmount)
	}

	switch decision {
	case 0:
		return o.cpuFoldOrCheck(callAmount)
	case 2:
		betAmt := o.cpuPotBet(potPct)
		return o.cpuRaiseOrBet(p, callAmount, betAmt)
	default:
		return o.cpuCallOrCheck(callAmount)
	}
}

// evalPreFlopStrength プリフロップハンド強度評価 (0-100)。
// ホールカード4枚 (Dramaha) でも5枚 (Big O) でも同じ採点ロジックで動作する。
// 4枚の場合の挙動は従来と完全に一致する。
func (o *Dramaha) evalPreFlopStrength(idx int) int {
	p := o.players[idx]
	n := p.GetCardsSize()
	if n < 4 {
		return 0
	}

	vals := make([]int, n)
	designs := make([]int, n)
	for i := 0; i < n; i++ {
		vals[i] = p.GetCard(i).GetValue()
		designs[i] = p.GetCard(i).GetDesign()
		if vals[i] == 1 {
			vals[i] = 14
		}
	}

	score := 0

	// ペアボーナス
	pairCount := 0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if vals[i] == vals[j] {
				pairCount++
				if vals[i] >= 10 {
					score += 10
				} else {
					score += 5
				}
			}
		}
	}

	// ハイカード値 (上位2枚)
	sorted := make([]int, n)
	copy(sorted, vals)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if sorted[j] > sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	score += sorted[0] + sorted[1]

	// スートボーナス (同スートペアの数)
	suitCount := make(map[int]int)
	for _, d := range designs {
		suitCount[d]++
	}
	for _, cnt := range suitCount {
		if cnt == 2 {
			score += 8 // single suited pair
		}
		if cnt >= 3 {
			// 3+ same suit is actually bad (waste of outs)
			score -= 5
		}
	}

	// コネクタボーナス: 全ホールカードの隣接ペアを評価
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			gap := sorted[i] - sorted[j]
			if gap < 0 {
				gap = -gap
			}
			switch gap {
			case 1:
				score += 4
			case 2:
				score += 2
			}
		}
	}

	// ラップ (連番) ボーナス: 任意の4枚のスパンで判定 (5カードドラマハに対応)
	isWrap := false
	for i := 0; i <= n-4; i++ {
		if sorted[i]-sorted[i+3] <= 4 {
			isWrap = true
			break
		}
	}
	if isWrap {
		score += 8
	}

	// ハイカードボーナス
	highCount := 0
	for _, v := range vals {
		if v >= 12 {
			highCount++
		}
	}
	if highCount >= 2 {
		score += 8
	}

	return clamp(score, 0, 100)
}

// --- リバイ/アドオン ---

func (o *Dramaha) checkAndTransitionAddon() bool {
	if o.config.AddonEnabled && o.handCount == o.config.AddonAfterHand {
		needHumanAddon := false
		for i, p := range o.players {
			if !o.addonUsed[i] {
				if p.GetIsHuman() {
					needHumanAddon = true
				} else {
					p.AddChips(o.config.AddonChips)
					o.addonUsed[i] = true
				}
			}
		}
		if needHumanAddon {
			o.phase = DramahaPhaseRebuy
			o.rebuyPhaseType = DramahaRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (o *Dramaha) Rebuy() error {
	if o.phase != DramahaPhaseRebuy || o.rebuyPhaseType != DramahaRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	for i, p := range o.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && o.rebuyCounts[i] < o.config.RebuyMaxCount {
			p.AddChips(o.config.RebuyChips)
			o.rebuyCounts[i]++
			o.appendLog(i, "rebuy", "rebuy", nil)
			break
		}
	}
	o.rebuyPhaseType = DramahaRebuyPhaseNone
	if o.checkAndTransitionAddon() {
		return nil
	}
	return o.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (o *Dramaha) SkipRebuy() error {
	if o.phase != DramahaPhaseRebuy || o.rebuyPhaseType != DramahaRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	o.rebuyPhaseType = DramahaRebuyPhaseNone
	for _, p := range o.players {
		if p.GetIsHuman() && p.GetChips() <= 0 {
			o.phase = DramahaPhaseEnd
			o.gameEndFlag = true
			return nil
		}
	}
	if o.checkAndTransitionAddon() {
		return nil
	}
	return o.continueReset()
}

// Addon 人間プレイヤーがアドオンを実行する
func (o *Dramaha) Addon() error {
	if o.phase != DramahaPhaseRebuy || o.rebuyPhaseType != DramahaRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, p := range o.players {
		if p.GetIsHuman() && !o.addonUsed[i] {
			p.AddChips(o.config.AddonChips)
			o.addonUsed[i] = true
			break
		}
	}
	o.rebuyPhaseType = DramahaRebuyPhaseNone
	return o.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (o *Dramaha) SkipAddon() error {
	if o.phase != DramahaPhaseRebuy || o.rebuyPhaseType != DramahaRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	o.rebuyPhaseType = DramahaRebuyPhaseNone
	return o.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (o *Dramaha) IsRebuyAvailable() bool {
	return rebuyAvailable(o.config.RebuyEnabled, o.handCount, o.config.RebuyPeriodHands, o.players, o.rebuyCounts, o.config.RebuyMaxCount)
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (o *Dramaha) IsAddonAvailable() bool {
	return addonAvailable(o.config.AddonEnabled, o.handCount, o.config.AddonAfterHand, o.players, o.addonUsed)
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (o *Dramaha) GetRebuyCounts() []int {
	return copyOf(o.rebuyCounts)
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (o *Dramaha) GetAddonUsed() []bool {
	return copyOf(o.addonUsed)
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (o *Dramaha) GetRebuyPhaseType() int { return o.rebuyPhaseType }

// GetEquity エクイティ計算結果を返す
func (o *Dramaha) GetEquity() *HoldemEquityResult {
	if o.phase < DramahaPhasePreFlop || o.phase > DramahaPhaseRiver {
		return nil
	}
	var humanPlayer *DramahaPlayer
	for _, p := range o.players {
		if p.GetIsHuman() {
			humanPlayer = p
			break
		}
	}
	if humanPlayer == nil || humanPlayer.GetFolded() {
		return nil
	}
	humanCards := make([]*Card, humanPlayer.GetCardsSize())
	for i := 0; i < humanPlayer.GetCardsSize(); i++ {
		humanCards[i] = humanPlayer.GetCard(i)
	}
	activePlayers := 0
	for _, p := range o.players {
		if !p.GetIsHuman() && !p.GetFolded() {
			activePlayers++
		}
	}
	result := calcDramahaEquityWithHoleCount(humanCards, o.communityCards, activePlayers, omahaEquitySimulations, nil, o.holeCardCount())
	return &result
}

// GetPotOdds ポットオッズを返す
func (o *Dramaha) GetPotOdds() float64 {
	if o.phase < DramahaPhasePreFlop || o.phase > DramahaPhaseRiver {
		return 0.0
	}
	humanCurrentBet := 0
	for _, p := range o.players {
		if p.GetIsHuman() {
			humanCurrentBet = p.GetCurrentBet()
			break
		}
	}
	callAmount := o.lastBet - humanCurrentBet
	if callAmount < 0 {
		callAmount = 0
	}
	return CalcPotOdds(o.pot, callAmount)
}

// --- ゲッター ---

// GetPhase フェーズ取得
func (o *Dramaha) GetPhase() int { return o.phase }

// GetPlayers プレイヤー一覧取得
func (o *Dramaha) GetPlayers() []*DramahaPlayer { return o.players }

// GetPlayer 指定プレイヤー取得
func (o *Dramaha) GetPlayer(i int) *DramahaPlayer {
	if i >= 0 && i < len(o.players) {
		return o.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (o *Dramaha) GetPlayerCnt() int { return len(o.players) }

// GetCommunityCards コミュニティカード取得
func (o *Dramaha) GetCommunityCards() []*Card { return o.communityCards }

// GetPot ポット取得
func (o *Dramaha) GetPot() int { return o.pot }

// GetSidePots サイドポット取得
func (o *Dramaha) GetSidePots() []SidePot { return o.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (o *Dramaha) GetDealerIdx() int { return o.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (o *Dramaha) GetCurrentTurn() int { return o.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (o *Dramaha) GetGameEndFlag() bool { return o.gameEndFlag }

// GetLastBet 最後のベット取得
func (o *Dramaha) GetLastBet() int { return o.lastBet }

// GetMinRaise 最小レイズ額取得
func (o *Dramaha) GetMinRaise() int { return o.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (o *Dramaha) GetRaiseCount() int { return o.raiseCount }

// GetRoundResults ラウンド結果取得
func (o *Dramaha) GetRoundResults() []HoldemResult { return o.roundResults }

// GetCpuActions CPU行動記録取得
func (o *Dramaha) GetCpuActions() []HoldemCpuAction { return o.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得
func (o *Dramaha) GetLastCpuError() error { return o.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (o *Dramaha) GetHumanProfile() *BettingHumanProfile { return o.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (o *Dramaha) ResetProfile() { o.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする (プロファイルがない場合はnil)
func (o *Dramaha) ExportProfile() interface{} {
	if o.humanProfile == nil {
		return nil
	}
	d := o.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (o *Dramaha) ImportProfile(data []byte) error {
	p, err := importBettingProfile(data)
	if err != nil || p == nil {
		return err
	}
	o.humanProfile = p
	return nil
}

// GetConfig 設定取得
func (o *Dramaha) GetConfig() DramahaConfig { return o.config }

// SetConfig 設定変更
func (o *Dramaha) SetConfig(cfg DramahaConfig) { o.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (o *Dramaha) IsHumanTurn() bool {
	return isHumanTurn(o.players, o.currentTurn)
}

// GetActedFlags actedフラグ取得
func (o *Dramaha) GetActedFlags() []bool {
	return copyOf(o.actedFlags)
}

// GetHandCount ハンド数取得
func (o *Dramaha) GetHandCount() int { return o.handCount }

// logAction ベッティングアクションを棋譜に記録する
func (o *Dramaha) logAction(playerIdx, action, amount int) {
	switch action {
	case DramahaActionFold:
		o.appendLog(playerIdx, "fold", "fold", nil)
	case DramahaActionCheck:
		o.appendLog(playerIdx, "check", "check", nil)
	case DramahaActionCall:
		o.appendLog(playerIdx, "call", fmt.Sprintf("call %d", o.players[playerIdx].GetCurrentBet()), nil)
	case DramahaActionBet:
		o.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case DramahaActionRaise:
		o.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case DramahaActionAllIn:
		o.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", o.players[playerIdx].GetCurrentBet()), nil)
	}
}

// dramahaJSON is the JSON wire format for Dramaha.
type dramahaJSON struct {
	TrumpCards       *TrumpCards              `json:"tc"`
	Players          []*DramahaPlayer         `json:"pl"`
	CommunityCards   []*Card                  `json:"cc"`
	Pot              int                      `json:"pt"`
	SidePots         []SidePot                `json:"sp"`
	DealerIdx        int                      `json:"di"`
	CurrentTurn      int                      `json:"ct"`
	Phase            int                      `json:"ph"`
	Config           DramahaConfig            `json:"cf"`
	GameEndFlag      bool                     `json:"ge"`
	LastBet          int                      `json:"lb"`
	MinRaise         int                      `json:"mr"`
	RaiseCount       int                      `json:"rc"`
	ActedFlags       []bool                   `json:"af"`
	RoundResults     []HoldemResult           `json:"rr"`
	CpuActions       []HoldemCpuAction        `json:"ca"`
	StartingChips    []int                    `json:"sc"`
	VPIPTracked      []bool                   `json:"vt"`
	PFRTracked       []bool                   `json:"ft"`
	ThreeBetTracked  []bool                   `json:"tt"`
	HandCount        int                      `json:"hc"`
	RebuyCounts      []int                    `json:"rb"`
	AddonUsed        []bool                   `json:"au"`
	RebuyPhaseType   int                      `json:"rp"`
	ActionLog        []*ActionLogEntry        `json:"al"`
	Profile          *BettingHumanProfileData `json:"pf,omitempty"`
	LastHumanPlayMs  int                      `json:"hm"`
	DrawnFlags       []bool                   `json:"dwf,omitempty"`
	HoleCards        int                      `json:"hcn,omitempty"`
	PreflopCommunity int                      `json:"pfc,omitempty"`
}

// dramahaMaxSliceLen caps slice sizes during deserialisation.
const dramahaMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (o *Dramaha) MarshalJSON() ([]byte, error) {
	j := dramahaJSON{
		TrumpCards:       o.trumpCards,
		Players:          o.players,
		CommunityCards:   o.communityCards,
		Pot:              o.pot,
		SidePots:         o.sidePots,
		DealerIdx:        o.dealerIdx,
		CurrentTurn:      o.currentTurn,
		Phase:            o.phase,
		Config:           o.config,
		GameEndFlag:      o.gameEndFlag,
		LastBet:          o.lastBet,
		MinRaise:         o.minRaise,
		RaiseCount:       o.raiseCount,
		ActedFlags:       o.actedFlags,
		RoundResults:     o.roundResults,
		DrawnFlags:       o.drawnFlags,
		CpuActions:       o.cpuActions,
		StartingChips:    o.startingChips,
		VPIPTracked:      o.vpipTracked,
		PFRTracked:       o.pfrTracked,
		ThreeBetTracked:  o.threeBetTracked,
		HandCount:        o.handCount,
		RebuyCounts:      o.rebuyCounts,
		AddonUsed:        o.addonUsed,
		RebuyPhaseType:   o.rebuyPhaseType,
		ActionLog:        o.actionLog,
		LastHumanPlayMs:  o.lastHumanPlayMs,
		HoleCards:        o.holeCards,
		PreflopCommunity: o.preflopCommunity,
	}
	if o.humanProfile != nil {
		d := o.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Dramaha) UnmarshalJSON(data []byte) error {
	var j dramahaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > dramahaMaxSliceLen || len(j.CommunityCards) > dramahaMaxSliceLen ||
		len(j.SidePots) > dramahaMaxSliceLen || len(j.ActedFlags) > dramahaMaxSliceLen ||
		len(j.RoundResults) > dramahaMaxSliceLen || len(j.CpuActions) > dramahaMaxSliceLen ||
		len(j.StartingChips) > dramahaMaxSliceLen || len(j.ActionLog) > dramahaMaxSliceLen {
		return fmt.Errorf("dramaha: input array exceeds maximum allowed size")
	}
	o.trumpCards = j.TrumpCards
	if o.trumpCards == nil {
		o.trumpCards = NewTrumpCards(0)
	}
	o.players = j.Players
	if o.players == nil {
		o.players = make([]*DramahaPlayer, 0)
	}
	o.communityCards = j.CommunityCards
	if o.communityCards == nil {
		o.communityCards = make([]*Card, 0)
	}
	o.pot = j.Pot
	o.sidePots = j.SidePots
	if o.sidePots == nil {
		o.sidePots = make([]SidePot, 0)
	}
	o.dealerIdx = j.DealerIdx
	o.currentTurn = j.CurrentTurn
	o.phase = j.Phase
	o.config = j.Config
	o.gameEndFlag = j.GameEndFlag
	o.lastBet = j.LastBet
	o.minRaise = j.MinRaise
	o.raiseCount = j.RaiseCount
	o.actedFlags = j.ActedFlags
	if o.actedFlags == nil {
		o.actedFlags = make([]bool, 0)
	}
	o.roundResults = j.RoundResults
	// **ドロー済みフラグは席数ぶん必ず用意する。** 往復で落とすと、ドロー中に
	// 保存した卓を戻したとき空スライスになり、Draw() が添字で panic する
	// (Worker はリクエストごとに卓を戻すので、これは必ず起きる経路)。
	o.drawnFlags = j.DrawnFlags
	if len(o.drawnFlags) != len(o.players) {
		o.drawnFlags = make([]bool, len(o.players))
		for i, p := range o.players {
			if p.GetFolded() {
				o.drawnFlags[i] = true
			}
		}
	}
	if o.roundResults == nil {
		o.roundResults = make([]HoldemResult, 0)
	}
	o.cpuActions = j.CpuActions
	if o.cpuActions == nil {
		o.cpuActions = make([]HoldemCpuAction, 0)
	}
	o.startingChips = j.StartingChips
	if o.startingChips == nil {
		o.startingChips = make([]int, 0)
	}
	o.vpipTracked = j.VPIPTracked
	if o.vpipTracked == nil {
		o.vpipTracked = make([]bool, 0)
	}
	o.pfrTracked = j.PFRTracked
	if o.pfrTracked == nil {
		o.pfrTracked = make([]bool, 0)
	}
	o.threeBetTracked = j.ThreeBetTracked
	if o.threeBetTracked == nil {
		o.threeBetTracked = make([]bool, 0)
	}
	o.handCount = j.HandCount
	o.rebuyCounts = j.RebuyCounts
	if o.rebuyCounts == nil {
		o.rebuyCounts = make([]int, 0)
	}
	o.addonUsed = j.AddonUsed
	if o.addonUsed == nil {
		o.addonUsed = make([]bool, 0)
	}
	o.rebuyPhaseType = j.RebuyPhaseType
	o.actionLog = j.ActionLog
	if o.actionLog == nil {
		o.actionLog = make([]*ActionLogEntry, 0)
	}
	o.lastHumanPlayMs = j.LastHumanPlayMs
	o.holeCards = j.HoleCards
	o.preflopCommunity = j.PreflopCommunity
	if j.Profile != nil {
		o.humanProfile = &BettingHumanProfile{}
		o.humanProfile.Import(*j.Profile)
	}
	return nil
}

// Resize プレイヤースライスを差し替え、プレイヤー数依存スライスを再初期化する
func (o *Dramaha) Resize(players []*DramahaPlayer) {
	o.players = players
	n := len(players)
	o.actedFlags = make([]bool, n)
	o.startingChips = make([]int, n)
	o.vpipTracked = make([]bool, n)
	o.pfrTracked = make([]bool, n)
	o.threeBetTracked = make([]bool, n)
	o.initTournamentState(n)
}
