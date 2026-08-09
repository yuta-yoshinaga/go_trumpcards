//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// パイナップルフェーズ定数 (Holdemと共通 + ディスカードフェーズ)
const (
	PineapplePhaseInit     = HoldemPhaseInit
	PineapplePhasePreFlop  = HoldemPhasePreFlop
	PineapplePhaseFlop     = HoldemPhaseFlop
	PineapplePhaseTurn     = HoldemPhaseTurn
	PineapplePhaseRiver    = HoldemPhaseRiver
	PineapplePhaseShowdown = HoldemPhaseShowdown
	PineapplePhaseEnd      = HoldemPhaseEnd
	PineapplePhaseRebuy    = HoldemPhaseRebuy
	PineapplePhaseDiscard  = 8 // フロップ後のディスカードフェーズ
)

// パイナップルアクション定数 (Holdemと共通)
const (
	PineappleActionFold  = HoldemActionFold
	PineappleActionCheck = HoldemActionCheck
	PineappleActionCall  = HoldemActionCall
	PineappleActionBet   = HoldemActionBet
	PineappleActionRaise = HoldemActionRaise
	PineappleActionAllIn = HoldemActionAllIn
)

// リバイフェーズ種別定数 (Holdemと共通)
const (
	PineappleRebuyPhaseNone  = HoldemRebuyPhaseNone
	PineappleRebuyPhaseRebuy = HoldemRebuyPhaseRebuy
	PineappleRebuyPhaseAddon = HoldemRebuyPhaseAddon
)

// Pineapple パイナップルポーカー (Crazy Pineapple / Irish Poker) クラス
// ホールカードを initialDealCount 枚配り、フロップ後に 2 枚になるまでディスカードする
type Pineapple struct {
	communityCardBettingBase
	trumpCards      *TrumpCards
	players         []*PineapplePlayer
	communityCards  []*Card
	sidePots        []SidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	config          PineappleConfig
	roundResults    []HoldemResult
	cpuActions      []HoldemCpuAction
	startingChips   []int
	vpipTracked     []bool
	pfrTracked      []bool
	threeBetTracked []bool
	tournamentBase  // handCount / rebuyCounts / addonUsed (issue #1463)
	lastCpuError    error
	rebuyPhaseType  int
	actionLogBase
	humanProfile    *BettingHumanProfile
	lastHumanPlayMs int
	discardDone     []bool // ディスカード済みフラグ
	// discardAfterFlopBetting true にすると「Crazy Pineapple / Irish Poker」モードとなり、
	// ディスカードのタイミングがフロップ公開直後ではなくフロップベッティング終了後になる。
	discardAfterFlopBetting bool
	initialDealCount        int // 初期配布枚数 (Pineapple/CrazyPineapple=3, IrishPoker=4)
}

// NewPineapple コンストラクタ
func NewPineapple(trumpCards *TrumpCards, players []*PineapplePlayer, config PineappleConfig) *Pineapple {
	p := &Pineapple{
		communityCardBettingBase: communityCardBettingBase{
			actedFlags: make([]bool, len(players)),
		},
		trumpCards:       trumpCards,
		players:          players,
		communityCards:   make([]*Card, 0),
		sidePots:         make([]SidePot, 0),
		roundResults:     make([]HoldemResult, 0),
		cpuActions:       make([]HoldemCpuAction, 0),
		startingChips:    make([]int, len(players)),
		vpipTracked:      make([]bool, len(players)),
		pfrTracked:       make([]bool, len(players)),
		threeBetTracked:  make([]bool, len(players)),
		discardDone:      make([]bool, len(players)),
		config:           config,
		phase:            PineapplePhaseInit,
		initialDealCount: 3,
	}
	p.initTournamentState(len(players))
	return p
}

// NewDefaultPineapple returns Pineapple with the default table size and
// DefaultPineappleConfig. Used as the single source of truth for CUI, Web,
// and Worker construction sites.
func NewDefaultPineapple() *Pineapple {
	cfg := DefaultPineappleConfig()
	return NewPineapple(NewTrumpCards(0), NewPineapplePlayersForTable(cfg.TableSize), cfg)
}

// NewCrazyPineapple コンストラクタ (Crazy Pineapple モード)。
// 通常の Pineapple と同じ構造だが、ディスカードのタイミングがフロップ
// ベッティング終了後（ターン配布前）に移動する。
func NewCrazyPineapple(trumpCards *TrumpCards, players []*PineapplePlayer, config PineappleConfig) *Pineapple {
	p := NewPineapple(trumpCards, players, config)
	p.discardAfterFlopBetting = true
	return p
}

// NewDefaultCrazyPineapple returns a Crazy Pineapple game with the default
// table size and DefaultPineappleConfig. Counterpart to NewDefaultPineapple.
func NewDefaultCrazyPineapple() *Pineapple {
	cfg := DefaultPineappleConfig()
	return NewCrazyPineapple(NewTrumpCards(0), NewPineapplePlayersForTable(cfg.TableSize), cfg)
}

// NewIrishPoker コンストラクタ (Irish Poker モード)。
// ホールカードを 4 枚配り、フロップベッティング終了後に 2 枚をディスカードする。
func NewIrishPoker(trumpCards *TrumpCards, players []*PineapplePlayer, config PineappleConfig) *Pineapple {
	p := NewPineapple(trumpCards, players, config)
	p.discardAfterFlopBetting = true
	p.initialDealCount = 4
	return p
}

// NewDefaultIrishPoker returns an Irish Poker game with the default table size
// and DefaultPineappleConfig.
func NewDefaultIrishPoker() *Pineapple {
	cfg := DefaultPineappleConfig()
	return NewIrishPoker(NewTrumpCards(0), NewPineapplePlayersForTable(cfg.TableSize), cfg)
}

// IsDiscardAfterFlopBetting reports whether this instance is running in
// Crazy Pineapple or Irish Poker mode (discard after flop betting round).
func (p *Pineapple) IsDiscardAfterFlopBetting() bool {
	return p.discardAfterFlopBetting
}

// GetInitialDealCount returns the number of hole cards dealt at the start.
func (p *Pineapple) GetInitialDealCount() int {
	if p.initialDealCount == 0 {
		return 3
	}
	return p.initialDealCount
}

// Reset ゲーム初期化
func (p *Pineapple) Reset() error {
	p.phase = PineapplePhaseInit
	p.pot = 0
	p.sidePots = make([]SidePot, 0)
	p.communityCards = make([]*Card, 0)
	p.gameEndFlag = false
	p.lastBet = 0
	p.minRaise = p.config.BigBlind
	p.raiseCount = 0
	p.actedFlags = make([]bool, len(p.players))
	p.roundResults = make([]HoldemResult, 0)
	p.cpuActions = make([]HoldemCpuAction, 0)
	p.rebuyPhaseType = PineappleRebuyPhaseNone
	p.actionLog = nil
	p.lastHumanPlayMs = 0
	p.discardDone = make([]bool, len(p.players))

	// メタAI: プロファイル初期化
	if p.config.CpuMetaAI {
		if p.humanProfile != nil {
			p.humanProfile.GamesPlayed++
		} else {
			p.humanProfile = &BettingHumanProfile{}
		}
	}

	p.trumpCards.Shuffle()
	for _, pl := range p.players {
		pl.Reset()
		pl.SetFolded(false)
		pl.SetAllIn(false)
		pl.SetCurrentBet(0)
		pl.handRank = 0
		pl.bestHand = nil
		if pl.GetChips() <= 0 && !p.config.RebuyEnabled {
			pl.SetChips(p.config.InitChips)
		}
		pl.IncrementTotalHands()
	}

	// HUDスタッツ追跡フラグをリセット
	p.vpipTracked = make([]bool, len(p.players))
	p.pfrTracked = make([]bool, len(p.players))
	p.threeBetTracked = make([]bool, len(p.players))

	// トーナメントモード: ブラインドエスカレーション
	if p.config.TournamentMode && p.config.BlindLevelHands > 0 && p.handCount > 0 && p.handCount%p.config.BlindLevelHands == 0 {
		p.config.SmallBlind = p.config.SmallBlind * p.config.BlindMultiplier / 100
		p.config.BigBlind = p.config.BigBlind * p.config.BlindMultiplier / 100
		if p.config.SmallBlind < 1 {
			p.config.SmallBlind = 1
		}
		if p.config.BigBlind < 2 {
			p.config.BigBlind = 2
		}
	}
	p.handCount++

	// リバイチェック
	if p.config.RebuyEnabled && p.handCount <= p.config.RebuyPeriodHands {
		needHumanRebuy := false
		for i, pl := range p.players {
			if pl.GetChips() <= 0 && p.rebuyCounts[i] < p.config.RebuyMaxCount {
				if pl.GetIsHuman() {
					needHumanRebuy = true
				} else {
					pl.AddChips(p.config.RebuyChips)
					p.rebuyCounts[i]++
				}
			}
		}
		if needHumanRebuy {
			p.phase = PineapplePhaseRebuy
			p.rebuyPhaseType = PineappleRebuyPhaseRebuy
			return nil
		}
	}

	// アドオンチェック
	if p.checkAndTransitionAddon() {
		return nil
	}

	return p.continueReset()
}

// continueReset ディール以降のリセット処理 (リバイ/アドオン判定後に実行)
func (p *Pineapple) continueReset() error {
	// ハンド開始時のチップを記録 (サイドポット計算用)
	p.startingChips = make([]int, len(p.players))
	for i, pl := range p.players {
		p.startingChips[i] = pl.GetChips()
	}

	// ブラインド投入
	p.postBlinds()

	// ホールカード配布
	for i := 0; i < p.initialDealCount; i++ {
		for j := 0; j < len(p.players); j++ {
			idx := (p.dealerIdx + 1 + j) % len(p.players)
			card := p.trumpCards.DrawCard()
			if card != nil {
				p.players[idx].AddCard(card)
			}
		}
	}

	p.phase = PineapplePhasePreFlop
	// UTG (ビッグブラインドの次) から開始
	p.currentTurn = (p.dealerIdx + 3) % len(p.players)

	// CPUプリフロップアクション実行
	if err := p.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postBlinds ブラインド投入
func (p *Pineapple) postBlinds() {
	sbIdx := (p.dealerIdx + 1) % len(p.players)
	bbIdx := (p.dealerIdx + 2) % len(p.players)

	sbAmount := p.config.SmallBlind
	if p.players[sbIdx].GetChips() < sbAmount {
		sbAmount = p.players[sbIdx].GetChips()
	}
	p.players[sbIdx].SubtractChips(sbAmount)
	p.players[sbIdx].SetCurrentBet(sbAmount)
	p.pot += sbAmount
	p.appendLog(sbIdx, "blind", fmt.Sprintf("posts small blind %d", sbAmount), nil)

	bbAmount := p.config.BigBlind
	if p.players[bbIdx].GetChips() < bbAmount {
		bbAmount = p.players[bbIdx].GetChips()
	}
	p.players[bbIdx].SubtractChips(bbAmount)
	p.players[bbIdx].SetCurrentBet(bbAmount)
	p.pot += bbAmount
	p.appendLog(bbIdx, "blind", fmt.Sprintf("posts big blind %d", bbAmount), nil)

	p.lastBet = bbAmount

	if p.players[sbIdx].GetChips() == 0 {
		p.players[sbIdx].SetAllIn(true)
		p.actedFlags[sbIdx] = true
	}
	if p.players[bbIdx].GetChips() == 0 {
		p.players[bbIdx].SetAllIn(true)
		p.actedFlags[bbIdx] = true
	}
}

// PlayerAction 人間プレイヤーのアクション実行
// humanPlayMs: 迷い時間(ms, 0=計測なし)
func (p *Pineapple) PlayerAction(action, amount, humanPlayMs int) error {
	if p.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if p.phase < PineapplePhasePreFlop || p.phase > PineapplePhaseRiver {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !p.players[p.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// メタAI: 人間アクション記録
	p.lastHumanPlayMs = humanPlayMs
	if p.config.CpuMetaAI && p.humanProfile != nil {
		pl := p.players[p.currentTurn]
		handRank := pl.EvalBestHand(p.communityCards)
		p.humanProfile.RecordAction(handRank, action)
		p.humanProfile.RecordHesitation(humanPlayMs)
		if p.lastBet > pl.GetCurrentBet() {
			p.humanProfile.RecordFoldToBet(action == PineappleActionFold)
		}
	}

	err := p.executeAction(p.currentTurn, action, amount)
	if err != nil {
		return err
	}

	p.advanceTurn()
	return p.runCpuActions()
}

// DiscardCard 人間プレイヤーが手札から1枚をディスカードする
func (p *Pineapple) DiscardCard(cardIdx int) error {
	return p.DiscardCards([]int{cardIdx})
}

// DiscardCards 人間プレイヤーが手札から複数枚を一括でディスカードする。
// Irish Poker は 4 枚から 2 枚を一度に捨てるため複数インデックスを受け取り、
// 高い順に取り除く (取り除きごとにインデックスがずれないようにするため)。
// インデックスは 1 枚以上・重複不可・範囲内でなければならない。
func (p *Pineapple) DiscardCards(cardIdxs []int) error {
	if p.phase != PineapplePhaseDiscard {
		return NewDomainError(ErrWrongPhase, "Discard is not allowed now.")
	}

	// 人間プレイヤーを探す
	humanIdx := -1
	for i, pl := range p.players {
		if pl.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return NewDomainError(ErrNotHumanTurn, "No human player found.")
	}
	if p.discardDone[humanIdx] {
		return NewDomainError(ErrWrongPhase, "Already discarded.")
	}

	pl := p.players[humanIdx]
	if len(cardIdxs) == 0 {
		return NewDomainError(ErrInvalidCard, "No card index supplied.")
	}
	// インデックス検証 (範囲・重複)
	seen := make(map[int]bool, len(cardIdxs))
	for _, idx := range cardIdxs {
		if idx < 0 || idx >= pl.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "Invalid card index.")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "Duplicate card index.")
		}
		seen[idx] = true
	}

	// 高い順に取り除く (低インデックスを後回しにしてズレを防ぐ)
	sorted := append([]int(nil), cardIdxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, idx := range sorted {
		p.removeCard(humanIdx, idx)
		p.appendLog(humanIdx, "discard", "discard", nil)
	}

	if pl.GetCardsSize() <= 2 {
		p.discardDone[humanIdx] = true
	}

	// 全員ディスカード済みならディスカード後のベッティングへ
	// (通常 Pineapple = フロップベッティング、Crazy Pineapple/Irish Poker = ターンベッティング)
	if p.allDiscardDone() {
		p.startBettingAfterDiscard()
	}

	return p.runCpuActions()
}

// removeCard 指定プレイヤーの指定インデックスのカードを取り除く
func (p *Pineapple) removeCard(playerIdx, cardIdx int) {
	pl := p.players[playerIdx]
	cards := make([]*Card, 0, pl.GetCardsSize()-1)
	for i := 0; i < pl.GetCardsSize(); i++ {
		if i != cardIdx {
			cards = append(cards, pl.GetCard(i))
		}
	}
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}

// allDiscardDone 全てのアクティブプレイヤーがディスカード済みかチェック
func (p *Pineapple) allDiscardDone() bool {
	for i, pl := range p.players {
		if pl.GetFolded() || pl.GetAllIn() {
			continue
		}
		if !p.discardDone[i] {
			return false
		}
	}
	return true
}

// startBettingAfterDiscard ディスカード完了後のベッティングラウンドを開始する。
// 通常 Pineapple ではフロップベッティングへ、Crazy Pineapple ではターンを
// 配ってターンベッティングへ進む（ディスカードがフロップベッティングの
// 後に行われたため）。
func (p *Pineapple) startBettingAfterDiscard() {
	if p.discardAfterFlopBetting {
		p.phase = PineapplePhaseTurn
		if card := p.trumpCards.DrawCard(); card != nil {
			p.communityCards = append(p.communityCards, card)
		}
		p.appendLog(-1, "deal", "dealt turn", p.communityCards[3:])
	} else {
		p.phase = PineapplePhaseFlop
	}

	// ベッティングラウンド初期化
	for _, pl := range p.players {
		pl.SetCurrentBet(0)
	}
	p.lastBet = 0
	p.minRaise = p.config.BigBlind
	p.raiseCount = 0
	p.actedFlags = make([]bool, len(p.players))
	for i, pl := range p.players {
		if pl.GetFolded() || pl.GetAllIn() {
			p.actedFlags[i] = true
		}
	}

	// ディーラーの次のアクティブプレイヤーから開始
	p.currentTurn = p.findNextActive(p.dealerIdx)
}

// trackPreFlopStats プリフロップのHUDスタッツを追跡
func (p *Pineapple) trackPreFlopStats(playerIdx, action int) {
	if p.phase != PineapplePhasePreFlop {
		return
	}

	isVPIPAction := false
	isPFRAction := false

	switch action {
	case PineappleActionCall:
		isVPIPAction = true
	case PineappleActionBet, PineappleActionRaise, PineappleActionAllIn:
		isVPIPAction = true
		isPFRAction = true
	}

	if isVPIPAction && !p.vpipTracked[playerIdx] {
		p.players[playerIdx].IncrementVPIP()
		p.vpipTracked[playerIdx] = true
	}

	if isPFRAction && !p.pfrTracked[playerIdx] {
		p.players[playerIdx].IncrementPFR()
		p.pfrTracked[playerIdx] = true
	}

	if p.raiseCount >= 1 && !p.threeBetTracked[playerIdx] {
		p.players[playerIdx].IncrementThreeBetOpportunity()
		if action == PineappleActionRaise || action == PineappleActionAllIn {
			p.players[playerIdx].IncrementThreeBet()
		}
		p.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats ポストフロップのAFスタッツを追跡
func (p *Pineapple) trackPostFlopStats(playerIdx, action int) {
	if p.phase < PineapplePhaseFlop || p.phase > PineapplePhaseRiver {
		return
	}

	switch action {
	case PineappleActionBet, PineappleActionRaise, PineappleActionAllIn:
		p.players[playerIdx].IncrementPostFlopBetRaise()
	case PineappleActionCall:
		p.players[playerIdx].IncrementPostFlopCall()
	}
}

// executeAction 指定プレイヤーのアクション実行
func (p *Pineapple) executeAction(playerIdx, action, amount int) error {
	p.trackPreFlopStats(playerIdx, action)
	p.trackPostFlopStats(playerIdx, action)

	bp := toBettingPlayers(p.players)
	state := p.bettingState()
	maxRaises, maxBetAmount := p.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, p.config.BigBlind, maxRaises, maxBetAmount)
	p.syncBettingState(state)
	if err != nil {
		return err
	}

	p.logAction(playerIdx, action, amount)

	if p.countActivePlayers() == 1 {
		p.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (p *Pineapple) advanceTurn() {
	if p.gameEndFlag {
		return
	}

	bp := toBettingPlayers(p.players)
	if p.isBettingRoundComplete(bp) {
		p.advancePhase()
		return
	}

	if next := p.findNextActiveTurn(p.currentTurn, bp); next >= 0 {
		p.currentTurn = next
		return
	}

	p.advancePhase()
}

// advancePhase 次のフェーズに進める
func (p *Pineapple) advancePhase() {
	// ラウンドベットリセット
	for _, pl := range p.players {
		pl.SetCurrentBet(0)
	}
	p.lastBet = 0
	p.minRaise = p.config.BigBlind
	p.raiseCount = 0
	p.actedFlags = make([]bool, len(p.players))
	for i, pl := range p.players {
		if pl.GetFolded() || pl.GetAllIn() {
			p.actedFlags[i] = true
		}
	}

	switch p.phase {
	case PineapplePhasePreFlop:
		// フロップカード3枚を配る
		for i := 0; i < 3; i++ {
			card := p.trumpCards.DrawCard()
			if card != nil {
				p.communityCards = append(p.communityCards, card)
			}
		}
		p.appendLog(-1, "deal", "dealt flop", p.communityCards)
		if p.discardAfterFlopBetting {
			// Crazy Pineapple: ディスカードはフロップベッティング後。
			// ここではフロップベッティングを開始する。
			p.phase = PineapplePhaseFlop
			break
		}
		// 通常 Pineapple: フロップ公開直後にディスカードへ。
		p.enterDiscardPhase()
		return
	case PineapplePhaseFlop:
		if p.discardAfterFlopBetting {
			// Crazy Pineapple: フロップベッティング終了後にディスカード。
			p.enterDiscardPhase()
			return
		}
		p.phase = PineapplePhaseTurn
		card := p.trumpCards.DrawCard()
		if card != nil {
			p.communityCards = append(p.communityCards, card)
		}
		p.appendLog(-1, "deal", "dealt turn", p.communityCards[3:])
	case PineapplePhaseTurn:
		p.phase = PineapplePhaseRiver
		card := p.trumpCards.DrawCard()
		if card != nil {
			p.communityCards = append(p.communityCards, card)
		}
		p.appendLog(-1, "deal", "dealt river", p.communityCards[4:])
	case PineapplePhaseRiver:
		p.phase = PineapplePhaseShowdown
		p.appendLog(-1, "showdown", "showdown", nil)
		p.resolveShowdown()
		return
	}

	// アクティブプレイヤーが0-1人ならショーダウンへ
	activeCnt := 0
	for _, pl := range p.players {
		if !pl.GetFolded() && !pl.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		p.dealRemainingCommunity()
		p.phase = PineapplePhaseShowdown
		p.resolveShowdown()
		return
	}

	// ディーラーの次のアクティブプレイヤーから開始
	p.currentTurn = p.findNextActive(p.dealerIdx)
}

// enterDiscardPhase ディスカードフェーズに入る
func (p *Pineapple) enterDiscardPhase() {
	p.phase = PineapplePhaseDiscard
	p.discardDone = make([]bool, len(p.players))

	// フォールド済み・オールイン済みプレイヤーは自動ディスカード
	for i, pl := range p.players {
		if pl.GetFolded() {
			p.discardDone[i] = true
			continue
		}
		if pl.GetAllIn() {
			for pl.GetCardsSize() > 2 {
				discardIdx := p.cpuDiscard(i)
				p.removeCard(i, discardIdx)
			}
			p.discardDone[i] = true
		}
	}

	// CPUプレイヤーの自動ディスカード
	for i, pl := range p.players {
		if p.discardDone[i] || pl.GetIsHuman() {
			continue
		}
		for pl.GetCardsSize() > 2 {
			discardIdx := p.cpuDiscard(i)
			p.removeCard(i, discardIdx)
			p.appendLog(i, "discard", "discard", nil)
		}
		p.discardDone[i] = true
	}

	// 全員ディスカード済み (人間なし or 人間がフォールド/オールイン) なら
	// ディスカード後のベッティングへ
	// (通常 Pineapple = フロップベッティング、Crazy Pineapple = ターンベッティング)
	if p.allDiscardDone() {
		p.startBettingAfterDiscard()
	}
}

// IsDiscardPhase ディスカードフェーズかどうか
func (p *Pineapple) IsDiscardPhase() bool {
	return p.phase == PineapplePhaseDiscard
}

// dealRemainingCommunity 残りのコミュニティカードを全て配る
func (p *Pineapple) dealRemainingCommunity() {
	dealUpTo(&p.communityCards, p.trumpCards, 5)
}

// findNextActive 指定インデックスの次のアクティブプレイヤーを探す
func (p *Pineapple) findNextActive(fromIdx int) int {
	return findNextActive(p.players, fromIdx)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (p *Pineapple) countActivePlayers() int {
	return countPlayers(p.players, func(pl *PineapplePlayer) bool { return !pl.GetFolded() })
}

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (p *Pineapple) resolveLastPlayer() {
	for i, pl := range p.players {
		if !pl.GetFolded() {
			pl.AddChips(p.pot)
			p.roundResults = []HoldemResult{{
				PlayerIdx: i,
				WonAmount: p.pot,
			}}
			p.pot = 0
			break
		}
	}
	p.phase = PineapplePhaseEnd
	p.gameEndFlag = true
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (p *Pineapple) resolveShowdown() {
	for _, pl := range p.players {
		if !pl.GetFolded() {
			pl.EvalBestHand(p.communityCards)
		}
	}

	bp := toBettingPlayers(p.players)
	p.sidePots = CalculateSidePots(bp, p.pot, p.startingChips)
	wonAmounts := DistributePots(bp, p.sidePots)

	p.roundResults = make([]HoldemResult, 0)
	humanLost := false
	for i, pl := range p.players {
		if pl.GetFolded() {
			continue
		}
		result := HoldemResult{
			PlayerIdx: i,
			HandRank:  pl.GetHandRank(),
			HandName:  p.getHandName(pl.GetHandRank()),
			BestHand:  pl.GetBestHand(),
			Kickers:   ExtractKickers(pl.GetBestHand(), pl.GetHandRank()),
			WonAmount: wonAmounts[i],
		}
		p.roundResults = append(p.roundResults, result)
		if pl.GetIsHuman() && wonAmounts[i] == 0 {
			humanLost = true
		}
	}

	if humanLost {
		return
	}

	p.finalizeShowdown()
}

// finalizeShowdown ショーダウンを完了し、END フェーズに遷移する
func (p *Pineapple) finalizeShowdown() {
	p.phase = PineapplePhaseEnd
	p.gameEndFlag = true
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
}

// Muck 人間プレイヤーがハンドをマックする
func (p *Pineapple) Muck() error {
	if p.phase != PineapplePhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Muck is not available now.")
	}
	for i := range p.roundResults {
		if p.players[p.roundResults[i].PlayerIdx].GetIsHuman() {
			p.roundResults[i].Mucked = true
			break
		}
	}
	p.finalizeShowdown()
	return nil
}

// ShowHand 人間プレイヤーがハンドを公開する
func (p *Pineapple) ShowHand() error {
	if p.phase != PineapplePhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	p.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (p *Pineapple) IsMuckAvailable() bool {
	if p.phase != PineapplePhaseShowdown {
		return false
	}
	for _, r := range p.roundResults {
		if p.players[r.PlayerIdx].GetIsHuman() && r.WonAmount == 0 {
			return true
		}
	}
	return false
}

// getHandName ハンドランクから名前を返す
func (p *Pineapple) getHandName(rank int) string {
	return pokerHandName(rank)
}

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (p *Pineapple) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(p.config.BettingLimit, p.pot, p.lastBet)
}

// GetEquity エクイティ計算結果を返す
func (p *Pineapple) GetEquity() *HoldemEquityResult {
	if p.phase < PineapplePhasePreFlop || p.phase > PineapplePhaseRiver {
		return nil
	}
	var humanPlayer *PineapplePlayer
	for _, pl := range p.players {
		if pl.GetIsHuman() {
			humanPlayer = pl
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
	for _, pl := range p.players {
		if !pl.GetIsHuman() && !pl.GetFolded() {
			activePlayers++
		}
	}
	result := CalcEquity(humanCards, p.communityCards, activePlayers, holdemEquitySimulations, nil)
	return &result
}

// PineappleDiscardPreview は捨て札候補1枚ぶんの「これを捨てたら残る手」。
type PineappleDiscardPreview struct {
	// CardIdx はホールカードのインデックス (0 始まり)。
	CardIdx int
	// HandRank はこの札を捨てたときに残る5枚の役 (PokerHand*)。
	HandRank int
	// Recommended は最も強い役が残る捨て札に付く。同点なら全部に付く。
	Recommended bool
}

// GetHumanDiscardPreviews は人間の3枚のホールカードそれぞれについて、
// 「その1枚を捨てたら残る2枚がボードと作る最強の役」を返す。Crazy Pineapple の
// 「3枚のうちどれを捨てるか」を横並びで比べられるようにするためのもの (#4686)。
//
// **ボードを見て役を名乗れるときしか返さない。**
//   - プレーンな Pineapple のディスカードはフロップ前なので nil。残る2枚だけでは
//     役が決まらない。代わりにスーテッド/コネクターの手掛かりが出る (#4685)。
//   - Irish Poker は2枚捨てなので「1枚捨てたら残る手」の前提が成り立たず nil。
//
// 判定は CPU の捨て方 (cpuDiscard) と同じ bestRankWithBoard を通す。別実装に
// すると、CPU 自身が選ばない捨て方を人間に勧めることになる。
func (p *Pineapple) GetHumanDiscardPreviews() []PineappleDiscardPreview {
	if p.phase != PineapplePhaseDiscard {
		return nil
	}
	var human *PineapplePlayer
	for _, pl := range p.players {
		if pl.GetIsHuman() {
			human = pl
			break
		}
	}
	// 手札3枚ちょうどのときだけ。捨て終わった後は2枚なのでここで弾かれる。
	if human == nil || human.GetFolded() || human.GetCardsSize() != 3 {
		return nil
	}
	if len(p.communityCards) < 3 {
		return nil
	}

	previews := make([]PineappleDiscardPreview, human.GetCardsSize())
	best := -1
	for k := range previews {
		keep := make([]*Card, 0, len(previews)-1)
		for i := 0; i < human.GetCardsSize(); i++ {
			if i != k {
				keep = append(keep, human.GetCard(i))
			}
		}
		rank := p.bestRankWithBoard(keep)
		previews[k] = PineappleDiscardPreview{CardIdx: k, HandRank: rank}
		if rank > best {
			best = rank
		}
	}
	for i := range previews {
		previews[i].Recommended = previews[i].HandRank == best
	}
	return previews
}

// PineappleDiscardPairPreview は「この2枚を捨てたら残る手」。
// Irish Poker のように2枚まとめて捨てる変種で使う。
type PineappleDiscardPairPreview struct {
	// DiscardIdx0 / DiscardIdx1 は捨てるホールカードのインデックス。
	DiscardIdx0 int
	DiscardIdx1 int
	// HandRank は残る2枚がボードと作る最強の役 (PokerHand*)。
	HandRank int
	// Recommended は最も強い役が残る組み合わせに付く。同点なら全部に付く。
	Recommended bool
}

// GetHumanDiscardPairPreviews は4枚配りで「どの2枚を捨てるか」の C(4,2)=6 通りを
// すべて評価して返す (#4687)。
//
// **Web は1枚目を選んだ後の3択しか出せない**（残り3択の絞り込み表示）。CUI には
// 選択途中という状態が無いので、代わりに6通りを最初から並べる。暗算の負荷を
// 消すという目的は同じで、情報量はこちらが多い。
//
// 3枚配り (Pineapple / Crazy Pineapple) では nil。そちらは1枚捨てなので
// GetHumanDiscardPreviews の担当。
func (p *Pineapple) GetHumanDiscardPairPreviews() []PineappleDiscardPairPreview {
	if p.phase != PineapplePhaseDiscard {
		return nil
	}
	var human *PineapplePlayer
	for _, pl := range p.players {
		if pl.GetIsHuman() {
			human = pl
			break
		}
	}
	if human == nil || human.GetFolded() || human.GetCardsSize() != 4 {
		return nil
	}
	if len(p.communityCards) < 3 {
		return nil
	}

	previews := make([]PineappleDiscardPairPreview, 0, 6)
	best := -1
	for i := 0; i < human.GetCardsSize(); i++ {
		for j := i + 1; j < human.GetCardsSize(); j++ {
			keep := make([]*Card, 0, 2)
			for k := 0; k < human.GetCardsSize(); k++ {
				if k != i && k != j {
					keep = append(keep, human.GetCard(k))
				}
			}
			rank := p.bestRankWithBoard(keep)
			previews = append(previews, PineappleDiscardPairPreview{
				DiscardIdx0: i, DiscardIdx1: j, HandRank: rank,
			})
			if rank > best {
				best = rank
			}
		}
	}
	for i := range previews {
		previews[i].Recommended = previews[i].HandRank == best
	}
	return previews
}

// GetPotOdds ポットオッズを返す
func (p *Pineapple) GetPotOdds() float64 {
	if p.phase < PineapplePhasePreFlop || p.phase > PineapplePhaseRiver {
		return 0.0
	}
	humanCurrentBet := 0
	for _, pl := range p.players {
		if pl.GetIsHuman() {
			humanCurrentBet = pl.GetCurrentBet()
			break
		}
	}
	callAmount := p.lastBet - humanCurrentBet
	if callAmount < 0 {
		callAmount = 0
	}
	return CalcPotOdds(p.pot, callAmount)
}

// --- リバイ/アドオン ---

func (p *Pineapple) checkAndTransitionAddon() bool {
	if p.config.AddonEnabled && p.handCount == p.config.AddonAfterHand {
		needHumanAddon := false
		for i, pl := range p.players {
			if !p.addonUsed[i] {
				if pl.GetIsHuman() {
					needHumanAddon = true
				} else {
					pl.AddChips(p.config.AddonChips)
					p.addonUsed[i] = true
				}
			}
		}
		if needHumanAddon {
			p.phase = PineapplePhaseRebuy
			p.rebuyPhaseType = PineappleRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (p *Pineapple) Rebuy() error {
	if p.phase != PineapplePhaseRebuy || p.rebuyPhaseType != PineappleRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	for i, pl := range p.players {
		if pl.GetIsHuman() && pl.GetChips() <= 0 && p.rebuyCounts[i] < p.config.RebuyMaxCount {
			pl.AddChips(p.config.RebuyChips)
			p.rebuyCounts[i]++
			p.appendLog(i, "rebuy", "rebuy", nil)
			break
		}
	}
	p.rebuyPhaseType = PineappleRebuyPhaseNone
	if p.checkAndTransitionAddon() {
		return nil
	}
	return p.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (p *Pineapple) SkipRebuy() error {
	if p.phase != PineapplePhaseRebuy || p.rebuyPhaseType != PineappleRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	p.rebuyPhaseType = PineappleRebuyPhaseNone
	for _, pl := range p.players {
		if pl.GetIsHuman() && pl.GetChips() <= 0 {
			p.phase = PineapplePhaseEnd
			p.gameEndFlag = true
			return nil
		}
	}
	if p.checkAndTransitionAddon() {
		return nil
	}
	return p.continueReset()
}

// Addon 人間プレイヤーがアドオンを実行する
func (p *Pineapple) Addon() error {
	if p.phase != PineapplePhaseRebuy || p.rebuyPhaseType != PineappleRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, pl := range p.players {
		if pl.GetIsHuman() && !p.addonUsed[i] {
			pl.AddChips(p.config.AddonChips)
			p.addonUsed[i] = true
			break
		}
	}
	p.rebuyPhaseType = PineappleRebuyPhaseNone
	return p.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (p *Pineapple) SkipAddon() error {
	if p.phase != PineapplePhaseRebuy || p.rebuyPhaseType != PineappleRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	p.rebuyPhaseType = PineappleRebuyPhaseNone
	return p.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (p *Pineapple) IsRebuyAvailable() bool {
	if !p.config.RebuyEnabled || p.handCount > p.config.RebuyPeriodHands {
		return false
	}
	for i, pl := range p.players {
		if pl.GetIsHuman() && pl.GetChips() <= 0 && p.rebuyCounts[i] < p.config.RebuyMaxCount {
			return true
		}
	}
	return false
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (p *Pineapple) IsAddonAvailable() bool {
	if !p.config.AddonEnabled || p.handCount != p.config.AddonAfterHand {
		return false
	}
	for i, pl := range p.players {
		if pl.GetIsHuman() && !p.addonUsed[i] {
			return true
		}
	}
	return false
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (p *Pineapple) GetRebuyCounts() []int {
	return copyOf(p.rebuyCounts)
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (p *Pineapple) GetAddonUsed() []bool {
	return copyOf(p.addonUsed)
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (p *Pineapple) GetRebuyPhaseType() int { return p.rebuyPhaseType }

// --- ゲッター ---

// GetPhase フェーズ取得
func (p *Pineapple) GetPhase() int { return p.phase }

// GetPlayers プレイヤー一覧取得
func (p *Pineapple) GetPlayers() []*PineapplePlayer { return p.players }

// GetPlayer 指定プレイヤー取得
func (p *Pineapple) GetPlayer(i int) *PineapplePlayer {
	if i >= 0 && i < len(p.players) {
		return p.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (p *Pineapple) GetPlayerCnt() int { return len(p.players) }

// GetCommunityCards コミュニティカード取得
func (p *Pineapple) GetCommunityCards() []*Card { return p.communityCards }

// GetPot ポット取得
func (p *Pineapple) GetPot() int { return p.pot }

// GetSidePots サイドポット取得
func (p *Pineapple) GetSidePots() []SidePot { return p.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (p *Pineapple) GetDealerIdx() int { return p.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (p *Pineapple) GetCurrentTurn() int { return p.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (p *Pineapple) GetGameEndFlag() bool { return p.gameEndFlag }

// GetLastBet 最後のベット取得
func (p *Pineapple) GetLastBet() int { return p.lastBet }

// GetMinRaise 最小レイズ額取得
func (p *Pineapple) GetMinRaise() int { return p.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (p *Pineapple) GetRaiseCount() int { return p.raiseCount }

// GetRoundResults ラウンド結果取得
func (p *Pineapple) GetRoundResults() []HoldemResult { return p.roundResults }

// GetCpuActions CPU行動記録取得
func (p *Pineapple) GetCpuActions() []HoldemCpuAction { return p.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得
func (p *Pineapple) GetLastCpuError() error { return p.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (p *Pineapple) GetHumanProfile() *BettingHumanProfile { return p.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (p *Pineapple) ResetProfile() { p.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする
func (p *Pineapple) ExportProfile() interface{} {
	if p.humanProfile == nil {
		return nil
	}
	d := p.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (p *Pineapple) ImportProfile(data []byte) error {
	prof, err := importBettingProfile(data)
	if err != nil || prof == nil {
		return err
	}
	p.humanProfile = prof
	return nil
}

// GetConfig 設定取得
func (p *Pineapple) GetConfig() PineappleConfig { return p.config }

// SetConfig 設定変更
func (p *Pineapple) SetConfig(cfg PineappleConfig) { p.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (p *Pineapple) IsHumanTurn() bool {
	return isHumanTurn(p.players, p.currentTurn)
}

// GetActedFlags actedフラグ取得
func (p *Pineapple) GetActedFlags() []bool {
	return copyOf(p.actedFlags)
}

// GetHandCount ハンド数取得
func (p *Pineapple) GetHandCount() int { return p.handCount }

// GetDiscardDone ディスカード済みフラグ取得
func (p *Pineapple) GetDiscardDone() []bool {
	result := make([]bool, len(p.discardDone))
	copy(result, p.discardDone)
	return result
}

// logAction ベッティングアクションを棋譜に記録する
func (p *Pineapple) logAction(playerIdx, action, amount int) {
	switch action {
	case PineappleActionFold:
		p.appendLog(playerIdx, "fold", "fold", nil)
	case PineappleActionCheck:
		p.appendLog(playerIdx, "check", "check", nil)
	case PineappleActionCall:
		p.appendLog(playerIdx, "call", fmt.Sprintf("call %d", p.players[playerIdx].GetCurrentBet()), nil)
	case PineappleActionBet:
		p.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case PineappleActionRaise:
		p.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case PineappleActionAllIn:
		p.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", p.players[playerIdx].GetCurrentBet()), nil)
	}
}

// pineappleJSON is the JSON wire format for Pineapple.
type pineappleJSON struct {
	TrumpCards              *TrumpCards              `json:"tc"`
	Players                 []*PineapplePlayer       `json:"pl"`
	CommunityCards          []*Card                  `json:"cc"`
	Pot                     int                      `json:"pt"`
	SidePots                []SidePot                `json:"sp"`
	DealerIdx               int                      `json:"di"`
	CurrentTurn             int                      `json:"ct"`
	Phase                   int                      `json:"ph"`
	Config                  PineappleConfig          `json:"cf"`
	GameEndFlag             bool                     `json:"ge"`
	LastBet                 int                      `json:"lb"`
	MinRaise                int                      `json:"mr"`
	RaiseCount              int                      `json:"rc"`
	ActedFlags              []bool                   `json:"af"`
	RoundResults            []HoldemResult           `json:"rr"`
	CpuActions              []HoldemCpuAction        `json:"ca"`
	StartingChips           []int                    `json:"sc"`
	VPIPTracked             []bool                   `json:"vt"`
	PFRTracked              []bool                   `json:"ft"`
	ThreeBetTracked         []bool                   `json:"tt"`
	HandCount               int                      `json:"hc"`
	RebuyCounts             []int                    `json:"rb"`
	AddonUsed               []bool                   `json:"au"`
	RebuyPhaseType          int                      `json:"rp"`
	ActionLog               []*ActionLogEntry        `json:"al"`
	Profile                 *BettingHumanProfileData `json:"pf,omitempty"`
	LastHumanPlayMs         int                      `json:"hm"`
	DiscardDone             []bool                   `json:"dd"`
	DiscardAfterFlopBetting bool                     `json:"cz,omitempty"`
	InitialDealCount        int                      `json:"idc,omitempty"`
}

// pineappleMaxSliceLen caps slice sizes during deserialisation.
const pineappleMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (p *Pineapple) MarshalJSON() ([]byte, error) {
	j := pineappleJSON{
		TrumpCards:              p.trumpCards,
		Players:                 p.players,
		CommunityCards:          p.communityCards,
		Pot:                     p.pot,
		SidePots:                p.sidePots,
		DealerIdx:               p.dealerIdx,
		CurrentTurn:             p.currentTurn,
		Phase:                   p.phase,
		Config:                  p.config,
		GameEndFlag:             p.gameEndFlag,
		LastBet:                 p.lastBet,
		MinRaise:                p.minRaise,
		RaiseCount:              p.raiseCount,
		ActedFlags:              p.actedFlags,
		RoundResults:            p.roundResults,
		CpuActions:              p.cpuActions,
		StartingChips:           p.startingChips,
		VPIPTracked:             p.vpipTracked,
		PFRTracked:              p.pfrTracked,
		ThreeBetTracked:         p.threeBetTracked,
		HandCount:               p.handCount,
		RebuyCounts:             p.rebuyCounts,
		AddonUsed:               p.addonUsed,
		RebuyPhaseType:          p.rebuyPhaseType,
		ActionLog:               p.actionLog,
		LastHumanPlayMs:         p.lastHumanPlayMs,
		DiscardDone:             p.discardDone,
		DiscardAfterFlopBetting: p.discardAfterFlopBetting,
		InitialDealCount:        p.initialDealCount,
	}
	if p.humanProfile != nil {
		d := p.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Pineapple) UnmarshalJSON(data []byte) error {
	var j pineappleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > pineappleMaxSliceLen || len(j.CommunityCards) > pineappleMaxSliceLen ||
		len(j.SidePots) > pineappleMaxSliceLen || len(j.ActedFlags) > pineappleMaxSliceLen ||
		len(j.RoundResults) > pineappleMaxSliceLen || len(j.CpuActions) > pineappleMaxSliceLen ||
		len(j.StartingChips) > pineappleMaxSliceLen || len(j.ActionLog) > pineappleMaxSliceLen {
		return fmt.Errorf("pineapple: input array exceeds maximum allowed size")
	}
	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCards(0)
	}
	p.players = j.Players
	if p.players == nil {
		p.players = make([]*PineapplePlayer, 0)
	}
	p.communityCards = j.CommunityCards
	if p.communityCards == nil {
		p.communityCards = make([]*Card, 0)
	}
	p.pot = j.Pot
	p.sidePots = j.SidePots
	if p.sidePots == nil {
		p.sidePots = make([]SidePot, 0)
	}
	p.dealerIdx = j.DealerIdx
	p.currentTurn = j.CurrentTurn
	p.phase = j.Phase
	p.config = j.Config
	p.gameEndFlag = j.GameEndFlag
	p.lastBet = j.LastBet
	p.minRaise = j.MinRaise
	p.raiseCount = j.RaiseCount
	p.actedFlags = j.ActedFlags
	if p.actedFlags == nil {
		p.actedFlags = make([]bool, 0)
	}
	p.roundResults = j.RoundResults
	if p.roundResults == nil {
		p.roundResults = make([]HoldemResult, 0)
	}
	p.cpuActions = j.CpuActions
	if p.cpuActions == nil {
		p.cpuActions = make([]HoldemCpuAction, 0)
	}
	p.startingChips = j.StartingChips
	if p.startingChips == nil {
		p.startingChips = make([]int, 0)
	}
	p.vpipTracked = j.VPIPTracked
	if p.vpipTracked == nil {
		p.vpipTracked = make([]bool, 0)
	}
	p.pfrTracked = j.PFRTracked
	if p.pfrTracked == nil {
		p.pfrTracked = make([]bool, 0)
	}
	p.threeBetTracked = j.ThreeBetTracked
	if p.threeBetTracked == nil {
		p.threeBetTracked = make([]bool, 0)
	}
	p.handCount = j.HandCount
	p.rebuyCounts = j.RebuyCounts
	if p.rebuyCounts == nil {
		p.rebuyCounts = make([]int, 0)
	}
	p.addonUsed = j.AddonUsed
	if p.addonUsed == nil {
		p.addonUsed = make([]bool, 0)
	}
	p.rebuyPhaseType = j.RebuyPhaseType
	p.actionLog = j.ActionLog
	if p.actionLog == nil {
		p.actionLog = make([]*ActionLogEntry, 0)
	}
	p.lastHumanPlayMs = j.LastHumanPlayMs
	p.discardDone = j.DiscardDone
	if p.discardDone == nil {
		p.discardDone = make([]bool, 0)
	}
	p.discardAfterFlopBetting = j.DiscardAfterFlopBetting
	p.initialDealCount = j.InitialDealCount
	if p.initialDealCount == 0 {
		p.initialDealCount = 3
	}
	if j.Profile != nil {
		p.humanProfile = &BettingHumanProfile{}
		p.humanProfile.Import(*j.Profile)
	}
	return nil
}

// Resize プレイヤースライスを差し替え、プレイヤー数依存スライスを再初期化する
func (p *Pineapple) Resize(players []*PineapplePlayer) {
	p.players = players
	n := len(players)
	p.actedFlags = make([]bool, n)
	p.startingChips = make([]int, n)
	p.vpipTracked = make([]bool, n)
	p.pfrTracked = make([]bool, n)
	p.threeBetTracked = make([]bool, n)
	p.discardDone = make([]bool, n)
	p.initTournamentState(n)
}
