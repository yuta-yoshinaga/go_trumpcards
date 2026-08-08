//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// オマハフェーズ定数 (Holdemと共通)
const (
	OmahaPhaseInit     = HoldemPhaseInit
	OmahaPhasePreFlop  = HoldemPhasePreFlop
	OmahaPhaseFlop     = HoldemPhaseFlop
	OmahaPhaseTurn     = HoldemPhaseTurn
	OmahaPhaseRiver    = HoldemPhaseRiver
	OmahaPhaseShowdown = HoldemPhaseShowdown
	OmahaPhaseEnd      = HoldemPhaseEnd
	OmahaPhaseRebuy    = HoldemPhaseRebuy
)

// オマハアクション定数 (Holdemと共通)
const (
	OmahaActionFold  = HoldemActionFold
	OmahaActionCheck = HoldemActionCheck
	OmahaActionCall  = HoldemActionCall
	OmahaActionBet   = HoldemActionBet
	OmahaActionRaise = HoldemActionRaise
	OmahaActionAllIn = HoldemActionAllIn
)

// リバイフェーズ種別定数 (Holdemと共通)
const (
	OmahaRebuyPhaseNone  = HoldemRebuyPhaseNone
	OmahaRebuyPhaseRebuy = HoldemRebuyPhaseRebuy
	OmahaRebuyPhaseAddon = HoldemRebuyPhaseAddon
)

// Omaha オマハホールデムクラス
//
// `hiLo` が true の場合は Omaha 8 or Better (Hi-Lo スプリットポット)
// として動作する。ショーダウン時にハイハンド (既存の `EvalBestHand`) と
// 8 or Better のローハンド (`EvalBestLowHand`) を並行して評価し、
// 各サイドポットを 50/50 で分割する (奇数チップは Hi 側へ)。
// qualified なローが 1 人もいない場合はハイ側が全額獲得する。
type Omaha struct {
	communityCardBettingBase
	trumpCards      *TrumpCards
	players         []*OmahaPlayer
	communityCards  []*Card
	sidePots        []SidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	hiLo            bool // Omaha 8 or Better (Hi-Lo) モード
	holeCards       int  // ホールカード配布枚数 (0 は既定の4枚扱い; Big O は5枚)
	config          OmahaConfig
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
}

// NewOmaha コンストラクタ
func NewOmaha(trumpCards *TrumpCards, players []*OmahaPlayer, config OmahaConfig) *Omaha {
	o := &Omaha{
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
		phase:           OmahaPhaseInit,
	}
	o.initTournamentState(len(players))
	return o
}

// NewDefaultOmaha returns Omaha with the default table size and DefaultOmahaConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultOmaha() *Omaha {
	cfg := DefaultOmahaConfig()
	return NewOmaha(NewTrumpCards(0), NewOmahaPlayersForTable(cfg.TableSize), cfg)
}

// NewOmahaHiLo returns Omaha configured as Omaha 8 or Better (Hi-Lo split pot).
func NewOmahaHiLo(trumpCards *TrumpCards, players []*OmahaPlayer, config OmahaConfig) *Omaha {
	o := NewOmaha(trumpCards, players, config)
	o.hiLo = true
	return o
}

// NewDefaultOmahaHiLo returns Omaha 8 or Better with the default table size
// and DefaultOmahaConfig. Single source of truth for CUI, Web, and Worker
// construction sites for the Hi-Lo variant.
func NewDefaultOmahaHiLo() *Omaha {
	cfg := DefaultOmahaConfig()
	return NewOmahaHiLo(NewTrumpCards(0), NewOmahaPlayersForTable(cfg.TableSize), cfg)
}

// GetIsHiLo returns true when the game is configured as Omaha 8 or Better.
func (o *Omaha) GetIsHiLo() bool { return o.hiLo }

// holeCardCount はホールカード配布枚数を返す。未設定 (0) のゲーム
// (既存のオマハや古いシリアライズ状態) は既定の4枚として扱う。
func (o *Omaha) holeCardCount() int {
	if o.holeCards <= 0 {
		return 4
	}
	return o.holeCards
}

// GetHoleCardCount はホールカード配布枚数を返す (Big O では5)。
func (o *Omaha) GetHoleCardCount() int { return o.holeCardCount() }

// Reset ゲーム初期化
func (o *Omaha) Reset() error {
	o.phase = OmahaPhaseInit
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
	o.rebuyPhaseType = OmahaRebuyPhaseNone
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
		p.lowBestHand = nil
		p.lowQualifies = false
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
			o.phase = OmahaPhaseRebuy
			o.rebuyPhaseType = OmahaRebuyPhaseRebuy
			return nil
		}
	}

	if o.checkAndTransitionAddon() {
		return nil
	}

	return o.continueReset()
}

// continueReset ディール以降のリセット処理
func (o *Omaha) continueReset() error {
	o.startingChips = make([]int, len(o.players))
	for i, p := range o.players {
		o.startingChips[i] = p.GetChips()
	}

	o.postBlinds()

	// ホールカード配布 (オマハ=4枚, Big O=5枚)
	for i := 0; i < o.holeCardCount(); i++ {
		for j := 0; j < len(o.players); j++ {
			idx := (o.dealerIdx + 1 + j) % len(o.players)
			card := o.trumpCards.DrawCard()
			if card != nil {
				o.players[idx].AddCard(card)
			}
		}
	}

	o.phase = OmahaPhasePreFlop
	o.currentTurn = (o.dealerIdx + 3) % len(o.players)

	if err := o.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postBlinds ブラインド投入
func (o *Omaha) postBlinds() {
	sbIdx := (o.dealerIdx + 1) % len(o.players)
	bbIdx := (o.dealerIdx + 2) % len(o.players)

	sbAmount := o.config.SmallBlind
	if o.players[sbIdx].GetChips() < sbAmount {
		sbAmount = o.players[sbIdx].GetChips()
	}
	o.players[sbIdx].SubtractChips(sbAmount)
	o.players[sbIdx].SetCurrentBet(sbAmount)
	o.pot += sbAmount
	o.appendLog(sbIdx, "blind", fmt.Sprintf("posts small blind %d", sbAmount), nil)

	bbAmount := o.config.BigBlind
	if o.players[bbIdx].GetChips() < bbAmount {
		bbAmount = o.players[bbIdx].GetChips()
	}
	o.players[bbIdx].SubtractChips(bbAmount)
	o.players[bbIdx].SetCurrentBet(bbAmount)
	o.pot += bbAmount
	o.appendLog(bbIdx, "blind", fmt.Sprintf("posts big blind %d", bbAmount), nil)

	o.lastBet = bbAmount

	if o.players[sbIdx].GetChips() == 0 {
		o.players[sbIdx].SetAllIn(true)
		o.actedFlags[sbIdx] = true
	}
	if o.players[bbIdx].GetChips() == 0 {
		o.players[bbIdx].SetAllIn(true)
		o.actedFlags[bbIdx] = true
	}
}

// PlayerAction 人間プレイヤーのアクション実行
// humanPlayMs: 迷い時間(ms, 0=計測なし)
func (o *Omaha) PlayerAction(action, amount, humanPlayMs int) error {
	if o.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if o.phase < OmahaPhasePreFlop || o.phase > OmahaPhaseRiver {
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
			o.humanProfile.RecordFoldToBet(action == OmahaActionFold)
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
func (o *Omaha) trackPreFlopStats(playerIdx, action int) {
	if o.phase != OmahaPhasePreFlop {
		return
	}

	isVPIPAction := false
	isPFRAction := false

	switch action {
	case OmahaActionCall:
		isVPIPAction = true
	case OmahaActionBet, OmahaActionRaise, OmahaActionAllIn:
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
		if action == OmahaActionRaise || action == OmahaActionAllIn {
			o.players[playerIdx].IncrementThreeBet()
		}
		o.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats ポストフロップのAFスタッツを追跡
func (o *Omaha) trackPostFlopStats(playerIdx, action int) {
	if o.phase < OmahaPhaseFlop || o.phase > OmahaPhaseRiver {
		return
	}

	switch action {
	case OmahaActionBet, OmahaActionRaise, OmahaActionAllIn:
		o.players[playerIdx].IncrementPostFlopBetRaise()
	case OmahaActionCall:
		o.players[playerIdx].IncrementPostFlopCall()
	}
}

// executeAction 指定プレイヤーのアクション実行
func (o *Omaha) executeAction(playerIdx, action, amount int) error {
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
func (o *Omaha) advanceTurn() {
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
func (o *Omaha) advancePhase() {
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
	case OmahaPhasePreFlop:
		o.phase = OmahaPhaseFlop
		for i := 0; i < 3; i++ {
			card := o.trumpCards.DrawCard()
			if card != nil {
				o.communityCards = append(o.communityCards, card)
			}
		}
		o.appendLog(-1, "deal", "dealt flop", o.communityCards)
	case OmahaPhaseFlop:
		o.phase = OmahaPhaseTurn
		card := o.trumpCards.DrawCard()
		if card != nil {
			o.communityCards = append(o.communityCards, card)
		}
		o.appendLog(-1, "deal", "dealt turn", o.communityCards[3:])
	case OmahaPhaseTurn:
		o.phase = OmahaPhaseRiver
		card := o.trumpCards.DrawCard()
		if card != nil {
			o.communityCards = append(o.communityCards, card)
		}
		o.appendLog(-1, "deal", "dealt river", o.communityCards[4:])
	case OmahaPhaseRiver:
		o.phase = OmahaPhaseShowdown
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
		o.phase = OmahaPhaseShowdown
		o.resolveShowdown()
		return
	}

	o.currentTurn = o.findNextActive(o.dealerIdx)
}

// dealRemainingCommunity 残りのコミュニティカードを全て配る
func (o *Omaha) dealRemainingCommunity() {
	for len(o.communityCards) < 5 {
		card := o.trumpCards.DrawCard()
		if card == nil {
			break
		}
		o.communityCards = append(o.communityCards, card)
	}
}

// findNextActive 指定インデックスの次のアクティブプレイヤーを探す
func (o *Omaha) findNextActive(fromIdx int) int {
	for i := 1; i <= len(o.players); i++ {
		next := (fromIdx + i) % len(o.players)
		if !o.players[next].GetFolded() && !o.players[next].GetAllIn() {
			return next
		}
	}
	return (fromIdx + 1) % len(o.players)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (o *Omaha) countActivePlayers() int {
	cnt := 0
	for _, p := range o.players {
		if !p.GetFolded() {
			cnt++
		}
	}
	return cnt
}

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (o *Omaha) resolveLastPlayer() {
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
	o.phase = OmahaPhaseEnd
	o.gameEndFlag = true
	o.dealerIdx = (o.dealerIdx + 1) % len(o.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (o *Omaha) resolveShowdown() {
	for _, p := range o.players {
		if !p.GetFolded() {
			p.EvalBestHand(o.communityCards)
			if o.hiLo {
				p.EvalBestLowHand(o.communityCards)
			}
		}
	}

	bp := toBettingPlayers(o.players)
	o.sidePots = CalculateSidePots(bp, o.pot, o.startingChips)

	var hiAmounts, lowAmounts map[int]int
	if o.hiLo {
		hiAmounts, lowAmounts = o.distributeHiLoPots(bp)
	} else {
		hiAmounts = DistributePots(bp, o.sidePots)
	}

	o.roundResults = make([]HoldemResult, 0)
	humanLost := false
	for i, p := range o.players {
		if p.GetFolded() {
			continue
		}
		hi := hiAmounts[i]
		lo := lowAmounts[i] // nil map read returns 0
		result := HoldemResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  o.getHandName(p.GetHandRank()),
			BestHand:  p.GetBestHand(),
			Kickers:   ExtractKickers(p.GetBestHand(), p.GetHandRank()),
			WonAmount: hi + lo,
		}
		if o.hiLo {
			result.HiWonAmount = hi
			result.LowWonAmount = lo
			result.LowQualifies = p.GetLowQualifies()
			result.LowBestHand = p.GetLowBestHand()
		}
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

// distributeHiLoPots は各サイドポットをハイ/ロー 50:50 で分配する。
// qualified なローが居ない場合はハイ側が全額獲得する。
// 奇数チップは Hi 側に寄せる (ポーカー慣例)。複数勝者間の余りは
// 既存 DistributePotsWithWinnerFunc と同じく winners[0] が引き取る。
func (o *Omaha) distributeHiLoPots(bp []BettingPlayer) (hi, lo map[int]int) {
	hi = make(map[int]int)
	lo = make(map[int]int)
	for _, sp := range o.sidePots {
		hiWinners := FindPotWinners(bp, sp.EligiblePlayers)
		if len(hiWinners) == 0 {
			continue
		}
		loWinners := o.findOmahaLowWinners(sp.EligiblePlayers)

		hiPot := sp.Amount
		loPot := 0
		if len(loWinners) > 0 {
			loPot = sp.Amount / 2
			hiPot = sp.Amount - loPot // 奇数チップは Hi 側に寄せる
		}

		distributeAmongWinners(bp, hiWinners, hiPot, hi)
		distributeAmongWinners(bp, loWinners, loPot, lo)
	}
	return hi, lo
}

// findOmahaLowWinners は対象プレイヤーの中から有効なロー (8 or Better)
// を持つベストプレイヤーを返す。qualified なローが 1 人もいなければ nil。
// 同点はスプリット (compareRazzCards == 0)。
func (o *Omaha) findOmahaLowWinners(eligible []int) []int {
	var winners []int
	var bestCards []*Card
	for _, idx := range eligible {
		p := o.players[idx]
		if p.GetFolded() || !p.GetLowQualifies() {
			continue
		}
		cards := p.GetLowBestHand()
		if bestCards == nil {
			bestCards = cards
			winners = []int{idx}
			continue
		}
		cmp := compareRazzCards(cards, bestCards)
		if cmp < 0 {
			bestCards = cards
			winners = []int{idx}
		} else if cmp == 0 {
			winners = append(winners, idx)
		}
	}
	return winners
}

// distributeAmongWinners は amount を winners 間で均等配分し、
// 余りは winners[0] に寄せる。amount==0 または winners==nil は no-op。
// チップ加算と won マップ更新を同時に行う。
func distributeAmongWinners(bp []BettingPlayer, winners []int, amount int, won map[int]int) {
	if amount <= 0 || len(winners) == 0 {
		return
	}
	share := amount / len(winners)
	remainder := amount % len(winners)
	for i, w := range winners {
		got := share
		if i == 0 {
			got += remainder
		}
		bp[w].AddChips(got)
		won[w] += got
	}
}

// finalizeShowdown ショーダウンを完了し、END フェーズに遷移する
func (o *Omaha) finalizeShowdown() {
	o.phase = OmahaPhaseEnd
	o.gameEndFlag = true
	o.dealerIdx = (o.dealerIdx + 1) % len(o.players)
}

// Muck 人間プレイヤーがハンドをマックする
func (o *Omaha) Muck() error {
	if o.phase != OmahaPhaseShowdown {
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
func (o *Omaha) ShowHand() error {
	if o.phase != OmahaPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	o.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (o *Omaha) IsMuckAvailable() bool {
	if o.phase != OmahaPhaseShowdown {
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
func (o *Omaha) getHandName(rank int) string {
	if rank >= 0 && rank < len(PokerHandNames) {
		return PokerHandNames[rank]
	}
	return "Unknown"
}

// runCpuActions CPUプレイヤーのアクションを実行
func (o *Omaha) runCpuActions() error {
	if o.gameEndFlag {
		return nil
	}
	const maxIterations = 500
	iterations := 0
	for !o.gameEndFlag && o.phase >= OmahaPhasePreFlop && o.phase <= OmahaPhaseRiver {
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
func (o *Omaha) handleCpuActionError(playerIdx, action int, err error) {
	o.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", playerIdx, action, err)
	callAmt := o.lastBet - o.players[playerIdx].GetCurrentBet()
	if callAmt > 0 {
		_ = o.executeAction(playerIdx, OmahaActionFold, 0)
	} else {
		_ = o.executeAction(playerIdx, OmahaActionCheck, 0)
	}
}

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (o *Omaha) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(o.config.BettingLimit, o.pot, o.lastBet)
}

// cpuDecide CPUプレイヤーの意思決定
func (o *Omaha) cpuDecide(idx int) (int, int) {
	p := o.players[idx]
	style := p.GetPlayStyle()
	callAmount := o.lastBet - p.GetCurrentBet()

	// GTO
	if style == HoldemStyleGTO {
		var action, amount int
		if o.phase == OmahaPhasePreFlop {
			action, amount = o.cpuDecidePreFlopGTO(idx, callAmount)
		} else {
			action, amount = o.cpuDecidePostFlopGTO(idx, callAmount)
		}
		maxRaises, maxBetAmount := o.bettingLimits()
		if maxBetAmount > 0 && amount > maxBetAmount {
			amount = maxBetAmount
		}
		if maxRaises > 0 && o.raiseCount >= maxRaises {
			if action == OmahaActionRaise || action == OmahaActionBet {
				if callAmount > 0 {
					return OmahaActionCall, 0
				}
				return OmahaActionCheck, 0
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
	if o.phase == OmahaPhasePreFlop {
		action, amount = o.cpuDecidePreFlop(idx, params, callAmount)
	} else {
		action, amount = o.cpuDecidePostFlop(idx, params, callAmount)
	}

	// メタAI: 人間のベット/レイズに対してコール確率を調整
	if o.config.CpuMetaAI && o.humanProfile != nil && o.lastHumanPlayMs > 0 {
		if action == OmahaActionFold && callAmount > 0 {
			handRank := p.EvalBestHand(o.communityCards)
			bracket := bettingHandBracket(handRank)
			adjustedCall := o.humanProfile.AdjustedCallChance(0.0, bracket, o.lastHumanPlayMs)
			if adjustedCall > 0 && rand.Float64() < adjustedCall {
				action = OmahaActionCall
				amount = 0
			}
		}
	}

	maxRaises, maxBetAmount := o.bettingLimits()

	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}

	if maxRaises > 0 && o.raiseCount >= maxRaises {
		if action == OmahaActionRaise || action == OmahaActionBet {
			if callAmount > 0 {
				return OmahaActionCall, 0
			}
			return OmahaActionCheck, 0
		}
	}
	return action, amount
}

// cpuFoldOrCheck コール額がある場合はフォールド、なければチェック
func (o *Omaha) cpuFoldOrCheck(callAmount int) (int, int) {
	return CpuFoldOrCheck(callAmount)
}

// cpuCallOrCheck コール額がある場合はコール、なければチェック
func (o *Omaha) cpuCallOrCheck(callAmount int) (int, int) {
	return CpuCallOrCheck(callAmount)
}

// cpuRaiseOrBet コール額がある場合はレイズ、なければベット
func (o *Omaha) cpuRaiseOrBet(p *OmahaPlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(p.GetChips(), callAmount, raiseAmt)
}

// cpuBetOrAllIn ベットする (チップ不足時はオールイン)
func (o *Omaha) cpuBetOrAllIn(p *OmahaPlayer, betAmt int) (int, int) {
	if betAmt > p.GetChips() {
		return OmahaActionAllIn, 0
	}
	return OmahaActionBet, betAmt
}

// cpuPotBet ポット比率ベースのベット額を計算
func (o *Omaha) cpuPotBet(potPct int) int {
	bet := o.pot * potPct / 100
	if bet < o.config.BigBlind {
		bet = o.config.BigBlind
	}
	if bet < o.minRaise {
		bet = o.minRaise
	}
	return bet
}

// cpuDecidePreFlop プリフロップのCPU意思決定
func (o *Omaha) cpuDecidePreFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := o.players[idx]
	strength := o.evalPreFlopStrength(idx)

	if strength < params.preFlopFoldThreshold {
		if params.preFlopFoldCompound {
			if callAmount > o.config.BigBlind*params.preFlopFoldCallMult {
				return OmahaActionFold, 0
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
		return OmahaActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return o.cpuBetOrAllIn(p, o.cpuPotBet(params.preFlopBluffPotPct))
	}
	return OmahaActionCheck, 0
}

// cpuDecidePostFlop フロップ以降のCPU意思決定
func (o *Omaha) cpuDecidePostFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := o.players[idx]
	handRank := p.EvalBestHand(o.communityCards)

	if params.aggressive {
		if handRank >= params.postFlopRaiseRank || rand.Intn(100) < params.bluffRate {
			return o.cpuRaiseOrBet(p, callAmount, o.cpuPotBet(params.postFlopRaisePotPct))
		}
		if params.postFlopFallbackFold {
			if params.postFlopCondCallRank >= 0 && handRank >= params.postFlopCondCallRank && callAmount > 0 {
				return OmahaActionCall, 0
			}
			return o.cpuFoldOrCheck(callAmount)
		}
		if callAmount > 0 {
			if handRank <= params.postFlopAggrFoldRank && callAmount > o.config.BigBlind*params.postFlopAggrFoldMult {
				return OmahaActionFold, 0
			}
			return OmahaActionCall, 0
		}
		return OmahaActionCheck, 0
	}

	if callAmount > 0 {
		if handRank <= params.postFlopPassFoldRank {
			if params.postFlopPassFoldMult < 0 || callAmount > o.config.BigBlind*params.postFlopPassFoldMult {
				return OmahaActionFold, 0
			}
		}
		return OmahaActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return o.cpuBetOrAllIn(p, o.cpuPotBet(params.postFlopBluffPotPct))
	}
	return OmahaActionCheck, 0
}

// cpuDecidePreFlopGTO GTOプリフロップ意思決定
func (o *Omaha) cpuDecidePreFlopGTO(idx, callAmount int) (int, int) {
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
func (o *Omaha) cpuDecidePostFlopGTO(idx, callAmount int) (int, int) {
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
// ホールカード4枚 (Omaha) でも5枚 (Big O) でも同じ採点ロジックで動作する。
// 4枚の場合の挙動は従来と完全に一致する。
func (o *Omaha) evalPreFlopStrength(idx int) int {
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

	// ラップ (連番) ボーナス: 任意の4枚のスパンで判定 (5カードオマハに対応)
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

func (o *Omaha) checkAndTransitionAddon() bool {
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
			o.phase = OmahaPhaseRebuy
			o.rebuyPhaseType = OmahaRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (o *Omaha) Rebuy() error {
	if o.phase != OmahaPhaseRebuy || o.rebuyPhaseType != OmahaRebuyPhaseRebuy {
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
	o.rebuyPhaseType = OmahaRebuyPhaseNone
	if o.checkAndTransitionAddon() {
		return nil
	}
	return o.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (o *Omaha) SkipRebuy() error {
	if o.phase != OmahaPhaseRebuy || o.rebuyPhaseType != OmahaRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	o.rebuyPhaseType = OmahaRebuyPhaseNone
	for _, p := range o.players {
		if p.GetIsHuman() && p.GetChips() <= 0 {
			o.phase = OmahaPhaseEnd
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
func (o *Omaha) Addon() error {
	if o.phase != OmahaPhaseRebuy || o.rebuyPhaseType != OmahaRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, p := range o.players {
		if p.GetIsHuman() && !o.addonUsed[i] {
			p.AddChips(o.config.AddonChips)
			o.addonUsed[i] = true
			break
		}
	}
	o.rebuyPhaseType = OmahaRebuyPhaseNone
	return o.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (o *Omaha) SkipAddon() error {
	if o.phase != OmahaPhaseRebuy || o.rebuyPhaseType != OmahaRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	o.rebuyPhaseType = OmahaRebuyPhaseNone
	return o.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (o *Omaha) IsRebuyAvailable() bool {
	if !o.config.RebuyEnabled || o.handCount > o.config.RebuyPeriodHands {
		return false
	}
	for i, p := range o.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && o.rebuyCounts[i] < o.config.RebuyMaxCount {
			return true
		}
	}
	return false
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (o *Omaha) IsAddonAvailable() bool {
	if !o.config.AddonEnabled || o.handCount != o.config.AddonAfterHand {
		return false
	}
	for i, p := range o.players {
		if p.GetIsHuman() && !o.addonUsed[i] {
			return true
		}
	}
	return false
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (o *Omaha) GetRebuyCounts() []int {
	result := make([]int, len(o.rebuyCounts))
	copy(result, o.rebuyCounts)
	return result
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (o *Omaha) GetAddonUsed() []bool {
	result := make([]bool, len(o.addonUsed))
	copy(result, o.addonUsed)
	return result
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (o *Omaha) GetRebuyPhaseType() int { return o.rebuyPhaseType }

// GetEquity エクイティ計算結果を返す
func (o *Omaha) GetEquity() *HoldemEquityResult {
	if o.phase < OmahaPhasePreFlop || o.phase > OmahaPhaseRiver {
		return nil
	}
	var humanPlayer *OmahaPlayer
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
	result := calcOmahaEquityWithHoleCount(humanCards, o.communityCards, activePlayers, omahaEquitySimulations, nil, o.holeCardCount())
	return &result
}

// GetPotOdds ポットオッズを返す
func (o *Omaha) GetPotOdds() float64 {
	if o.phase < OmahaPhasePreFlop || o.phase > OmahaPhaseRiver {
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
func (o *Omaha) GetPhase() int { return o.phase }

// GetPlayers プレイヤー一覧取得
func (o *Omaha) GetPlayers() []*OmahaPlayer { return o.players }

// GetPlayer 指定プレイヤー取得
func (o *Omaha) GetPlayer(i int) *OmahaPlayer {
	if i >= 0 && i < len(o.players) {
		return o.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (o *Omaha) GetPlayerCnt() int { return len(o.players) }

// GetCommunityCards コミュニティカード取得
func (o *Omaha) GetCommunityCards() []*Card { return o.communityCards }

// GetPot ポット取得
func (o *Omaha) GetPot() int { return o.pot }

// GetSidePots サイドポット取得
func (o *Omaha) GetSidePots() []SidePot { return o.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (o *Omaha) GetDealerIdx() int { return o.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (o *Omaha) GetCurrentTurn() int { return o.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (o *Omaha) GetGameEndFlag() bool { return o.gameEndFlag }

// GetLastBet 最後のベット取得
func (o *Omaha) GetLastBet() int { return o.lastBet }

// GetMinRaise 最小レイズ額取得
func (o *Omaha) GetMinRaise() int { return o.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (o *Omaha) GetRaiseCount() int { return o.raiseCount }

// GetRoundResults ラウンド結果取得
func (o *Omaha) GetRoundResults() []HoldemResult { return o.roundResults }

// GetCpuActions CPU行動記録取得
func (o *Omaha) GetCpuActions() []HoldemCpuAction { return o.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得
func (o *Omaha) GetLastCpuError() error { return o.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (o *Omaha) GetHumanProfile() *BettingHumanProfile { return o.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (o *Omaha) ResetProfile() { o.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする (プロファイルがない場合はnil)
func (o *Omaha) ExportProfile() interface{} {
	if o.humanProfile == nil {
		return nil
	}
	d := o.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (o *Omaha) ImportProfile(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	d, err := ImportBettingHumanProfileJSON(data)
	if err != nil {
		return err
	}
	o.humanProfile = &BettingHumanProfile{}
	o.humanProfile.Import(d)
	return nil
}

// GetConfig 設定取得
func (o *Omaha) GetConfig() OmahaConfig { return o.config }

// SetConfig 設定変更
func (o *Omaha) SetConfig(cfg OmahaConfig) { o.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (o *Omaha) IsHumanTurn() bool {
	if o.currentTurn >= 0 && o.currentTurn < len(o.players) {
		return o.players[o.currentTurn].GetIsHuman()
	}
	return false
}

// GetActedFlags actedフラグ取得
func (o *Omaha) GetActedFlags() []bool {
	result := make([]bool, len(o.actedFlags))
	copy(result, o.actedFlags)
	return result
}

// GetHandCount ハンド数取得
func (o *Omaha) GetHandCount() int { return o.handCount }

// logAction ベッティングアクションを棋譜に記録する
func (o *Omaha) logAction(playerIdx, action, amount int) {
	switch action {
	case OmahaActionFold:
		o.appendLog(playerIdx, "fold", "fold", nil)
	case OmahaActionCheck:
		o.appendLog(playerIdx, "check", "check", nil)
	case OmahaActionCall:
		o.appendLog(playerIdx, "call", fmt.Sprintf("call %d", o.players[playerIdx].GetCurrentBet()), nil)
	case OmahaActionBet:
		o.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case OmahaActionRaise:
		o.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case OmahaActionAllIn:
		o.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", o.players[playerIdx].GetCurrentBet()), nil)
	}
}

// omahaJSON is the JSON wire format for Omaha.
type omahaJSON struct {
	TrumpCards      *TrumpCards              `json:"tc"`
	Players         []*OmahaPlayer           `json:"pl"`
	CommunityCards  []*Card                  `json:"cc"`
	Pot             int                      `json:"pt"`
	SidePots        []SidePot                `json:"sp"`
	DealerIdx       int                      `json:"di"`
	CurrentTurn     int                      `json:"ct"`
	Phase           int                      `json:"ph"`
	Config          OmahaConfig              `json:"cf"`
	GameEndFlag     bool                     `json:"ge"`
	LastBet         int                      `json:"lb"`
	MinRaise        int                      `json:"mr"`
	RaiseCount      int                      `json:"rc"`
	ActedFlags      []bool                   `json:"af"`
	RoundResults    []HoldemResult           `json:"rr"`
	CpuActions      []HoldemCpuAction        `json:"ca"`
	StartingChips   []int                    `json:"sc"`
	VPIPTracked     []bool                   `json:"vt"`
	PFRTracked      []bool                   `json:"ft"`
	ThreeBetTracked []bool                   `json:"tt"`
	HandCount       int                      `json:"hc"`
	RebuyCounts     []int                    `json:"rb"`
	AddonUsed       []bool                   `json:"au"`
	RebuyPhaseType  int                      `json:"rp"`
	ActionLog       []*ActionLogEntry        `json:"al"`
	Profile         *BettingHumanProfileData `json:"pf,omitempty"`
	LastHumanPlayMs int                      `json:"hm"`
	HiLo            bool                     `json:"hl,omitempty"`
	HoleCards       int                      `json:"hcn,omitempty"`
}

// omahaMaxSliceLen caps slice sizes during deserialisation.
const omahaMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (o *Omaha) MarshalJSON() ([]byte, error) {
	j := omahaJSON{
		TrumpCards:      o.trumpCards,
		Players:         o.players,
		CommunityCards:  o.communityCards,
		Pot:             o.pot,
		SidePots:        o.sidePots,
		DealerIdx:       o.dealerIdx,
		CurrentTurn:     o.currentTurn,
		Phase:           o.phase,
		Config:          o.config,
		GameEndFlag:     o.gameEndFlag,
		LastBet:         o.lastBet,
		MinRaise:        o.minRaise,
		RaiseCount:      o.raiseCount,
		ActedFlags:      o.actedFlags,
		RoundResults:    o.roundResults,
		CpuActions:      o.cpuActions,
		StartingChips:   o.startingChips,
		VPIPTracked:     o.vpipTracked,
		PFRTracked:      o.pfrTracked,
		ThreeBetTracked: o.threeBetTracked,
		HandCount:       o.handCount,
		RebuyCounts:     o.rebuyCounts,
		AddonUsed:       o.addonUsed,
		RebuyPhaseType:  o.rebuyPhaseType,
		ActionLog:       o.actionLog,
		LastHumanPlayMs: o.lastHumanPlayMs,
		HiLo:            o.hiLo,
		HoleCards:       o.holeCards,
	}
	if o.humanProfile != nil {
		d := o.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Omaha) UnmarshalJSON(data []byte) error {
	var j omahaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > omahaMaxSliceLen || len(j.CommunityCards) > omahaMaxSliceLen ||
		len(j.SidePots) > omahaMaxSliceLen || len(j.ActedFlags) > omahaMaxSliceLen ||
		len(j.RoundResults) > omahaMaxSliceLen || len(j.CpuActions) > omahaMaxSliceLen ||
		len(j.StartingChips) > omahaMaxSliceLen || len(j.ActionLog) > omahaMaxSliceLen {
		return fmt.Errorf("omaha: input array exceeds maximum allowed size")
	}
	o.trumpCards = j.TrumpCards
	if o.trumpCards == nil {
		o.trumpCards = NewTrumpCards(0)
	}
	o.players = j.Players
	if o.players == nil {
		o.players = make([]*OmahaPlayer, 0)
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
	o.hiLo = j.HiLo
	o.holeCards = j.HoleCards
	if j.Profile != nil {
		o.humanProfile = &BettingHumanProfile{}
		o.humanProfile.Import(*j.Profile)
	}
	return nil
}

// Resize プレイヤースライスを差し替え、プレイヤー数依存スライスを再初期化する
func (o *Omaha) Resize(players []*OmahaPlayer) {
	o.players = players
	n := len(players)
	o.actedFlags = make([]bool, n)
	o.startingChips = make([]int, n)
	o.vpipTracked = make([]bool, n)
	o.pfrTracked = make([]bool, n)
	o.threeBetTracked = make([]bool, n)
	o.initTournamentState(n)
}
