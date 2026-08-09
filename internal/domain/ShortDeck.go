//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// ショートデックフェーズ定数 (Holdemと共通)
const (
	ShortDeckPhaseInit     = HoldemPhaseInit
	ShortDeckPhasePreFlop  = HoldemPhasePreFlop
	ShortDeckPhaseFlop     = HoldemPhaseFlop
	ShortDeckPhaseTurn     = HoldemPhaseTurn
	ShortDeckPhaseRiver    = HoldemPhaseRiver
	ShortDeckPhaseShowdown = HoldemPhaseShowdown
	ShortDeckPhaseEnd      = HoldemPhaseEnd
	ShortDeckPhaseRebuy    = HoldemPhaseRebuy
)

// ショートデックアクション定数 (Holdemと共通)
const (
	ShortDeckActionFold  = HoldemActionFold
	ShortDeckActionCheck = HoldemActionCheck
	ShortDeckActionCall  = HoldemActionCall
	ShortDeckActionBet   = HoldemActionBet
	ShortDeckActionRaise = HoldemActionRaise
	ShortDeckActionAllIn = HoldemActionAllIn
)

// リバイフェーズ種別定数 (Holdemと共通)
const (
	ShortDeckRebuyPhaseNone  = HoldemRebuyPhaseNone
	ShortDeckRebuyPhaseRebuy = HoldemRebuyPhaseRebuy
	ShortDeckRebuyPhaseAddon = HoldemRebuyPhaseAddon
)

// ShortDeck ショートデックホールデムクラス
type ShortDeck struct {
	communityCardBettingBase
	trumpCards      *TrumpCards
	players         []*ShortDeckPlayer
	communityCards  []*Card
	sidePots        []SidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	config          ShortDeckConfig
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

// NewShortDeck コンストラクタ
func NewShortDeck(trumpCards *TrumpCards, players []*ShortDeckPlayer, config ShortDeckConfig) *ShortDeck {
	sd := &ShortDeck{
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
		phase:           ShortDeckPhaseInit,
	}
	sd.initTournamentState(len(players))
	return sd
}

// NewDefaultShortDeck returns ShortDeck with the default table size and
// DefaultShortDeckConfig using the short (6+) deck. Used as the single source
// of truth for CUI, Web, and Worker construction sites.
func NewDefaultShortDeck() *ShortDeck {
	cfg := DefaultShortDeckConfig()
	return NewShortDeck(NewTrumpCardsShortDeck(), NewShortDeckPlayersForTable(cfg.TableSize), cfg)
}

// Reset ゲーム初期化
func (sd *ShortDeck) Reset() error {
	sd.phase = ShortDeckPhaseInit
	sd.pot = 0
	sd.sidePots = make([]SidePot, 0)
	sd.communityCards = make([]*Card, 0)
	sd.gameEndFlag = false
	sd.lastBet = 0
	sd.minRaise = sd.config.BigBlind
	sd.raiseCount = 0
	sd.actedFlags = make([]bool, len(sd.players))
	sd.roundResults = make([]HoldemResult, 0)
	sd.cpuActions = make([]HoldemCpuAction, 0)
	sd.rebuyPhaseType = ShortDeckRebuyPhaseNone
	sd.actionLog = nil
	sd.lastHumanPlayMs = 0

	// メタAI: プロファイル初期化
	if sd.config.CpuMetaAI {
		if sd.humanProfile != nil {
			sd.humanProfile.GamesPlayed++
		} else {
			sd.humanProfile = &BettingHumanProfile{}
		}
	}

	sd.trumpCards.Shuffle()
	for _, p := range sd.players {
		p.Reset()
		p.SetFolded(false)
		p.SetAllIn(false)
		p.SetCurrentBet(0)
		p.handRank = 0
		p.bestHand = nil
		if p.GetChips() <= 0 && !sd.config.RebuyEnabled {
			p.SetChips(sd.config.InitChips)
		}
		p.IncrementTotalHands()
	}

	sd.vpipTracked = make([]bool, len(sd.players))
	sd.pfrTracked = make([]bool, len(sd.players))
	sd.threeBetTracked = make([]bool, len(sd.players))

	if sd.config.TournamentMode && sd.config.BlindLevelHands > 0 && sd.handCount > 0 && sd.handCount%sd.config.BlindLevelHands == 0 {
		sd.config.SmallBlind = sd.config.SmallBlind * sd.config.BlindMultiplier / 100
		sd.config.BigBlind = sd.config.BigBlind * sd.config.BlindMultiplier / 100
		if sd.config.SmallBlind < 1 {
			sd.config.SmallBlind = 1
		}
		if sd.config.BigBlind < 2 {
			sd.config.BigBlind = 2
		}
	}
	sd.handCount++

	if sd.config.RebuyEnabled && sd.handCount <= sd.config.RebuyPeriodHands {
		needHumanRebuy := false
		for i, p := range sd.players {
			if p.GetChips() <= 0 && sd.rebuyCounts[i] < sd.config.RebuyMaxCount {
				if p.GetIsHuman() {
					needHumanRebuy = true
				} else {
					p.AddChips(sd.config.RebuyChips)
					sd.rebuyCounts[i]++
				}
			}
		}
		if needHumanRebuy {
			sd.phase = ShortDeckPhaseRebuy
			sd.rebuyPhaseType = ShortDeckRebuyPhaseRebuy
			return nil
		}
	}

	if sd.checkAndTransitionAddon() {
		return nil
	}

	return sd.continueReset()
}

// continueReset ディール以降のリセット処理
func (sd *ShortDeck) continueReset() error {
	sd.startingChips = make([]int, len(sd.players))
	for i, p := range sd.players {
		sd.startingChips[i] = p.GetChips()
	}

	sd.postBlinds()

	// ホールカード配布 (2枚: ショートデックはホールデムと同じ)
	for i := 0; i < 2; i++ {
		for j := 0; j < len(sd.players); j++ {
			idx := (sd.dealerIdx + 1 + j) % len(sd.players)
			card := sd.trumpCards.DrawCard()
			if card != nil {
				sd.players[idx].AddCard(card)
			}
		}
	}

	sd.phase = ShortDeckPhasePreFlop
	sd.currentTurn = (sd.dealerIdx + 3) % len(sd.players)

	if err := sd.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postBlinds ブラインド投入
func (sd *ShortDeck) postBlinds() {
	sbIdx := (sd.dealerIdx + 1) % len(sd.players)
	bbIdx := (sd.dealerIdx + 2) % len(sd.players)

	sbAmount := sd.config.SmallBlind
	if sd.players[sbIdx].GetChips() < sbAmount {
		sbAmount = sd.players[sbIdx].GetChips()
	}
	sd.players[sbIdx].SubtractChips(sbAmount)
	sd.players[sbIdx].SetCurrentBet(sbAmount)
	sd.pot += sbAmount
	sd.appendLog(sbIdx, "blind", fmt.Sprintf("posts small blind %d", sbAmount), nil)

	bbAmount := sd.config.BigBlind
	if sd.players[bbIdx].GetChips() < bbAmount {
		bbAmount = sd.players[bbIdx].GetChips()
	}
	sd.players[bbIdx].SubtractChips(bbAmount)
	sd.players[bbIdx].SetCurrentBet(bbAmount)
	sd.pot += bbAmount
	sd.appendLog(bbIdx, "blind", fmt.Sprintf("posts big blind %d", bbAmount), nil)

	sd.lastBet = bbAmount

	if sd.players[sbIdx].GetChips() == 0 {
		sd.players[sbIdx].SetAllIn(true)
		sd.actedFlags[sbIdx] = true
	}
	if sd.players[bbIdx].GetChips() == 0 {
		sd.players[bbIdx].SetAllIn(true)
		sd.actedFlags[bbIdx] = true
	}
}

// PlayerAction 人間プレイヤーのアクション実行
// humanPlayMs: 迷い時間(ms, 0=計測なし)
func (sd *ShortDeck) PlayerAction(action, amount, humanPlayMs int) error {
	if sd.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if sd.phase < ShortDeckPhasePreFlop || sd.phase > ShortDeckPhaseRiver {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !sd.players[sd.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// メタAI: 人間アクション記録
	sd.lastHumanPlayMs = humanPlayMs
	if sd.config.CpuMetaAI && sd.humanProfile != nil {
		pl := sd.players[sd.currentTurn]
		handRank := pl.EvalBestHand(sd.communityCards)
		sd.humanProfile.RecordAction(handRank, action)
		sd.humanProfile.RecordHesitation(humanPlayMs)
		if sd.lastBet > pl.GetCurrentBet() {
			sd.humanProfile.RecordFoldToBet(action == ShortDeckActionFold)
		}
	}

	err := sd.executeAction(sd.currentTurn, action, amount)
	if err != nil {
		return err
	}

	sd.advanceTurn()
	return sd.runCpuActions()
}

// trackPreFlopStats プリフロップのHUDスタッツを追跡
func (sd *ShortDeck) trackPreFlopStats(playerIdx, action int) {
	if sd.phase != ShortDeckPhasePreFlop {
		return
	}

	isVPIPAction := false
	isPFRAction := false

	switch action {
	case ShortDeckActionCall:
		isVPIPAction = true
	case ShortDeckActionBet, ShortDeckActionRaise, ShortDeckActionAllIn:
		isVPIPAction = true
		isPFRAction = true
	}

	if isVPIPAction && !sd.vpipTracked[playerIdx] {
		sd.players[playerIdx].IncrementVPIP()
		sd.vpipTracked[playerIdx] = true
	}

	if isPFRAction && !sd.pfrTracked[playerIdx] {
		sd.players[playerIdx].IncrementPFR()
		sd.pfrTracked[playerIdx] = true
	}

	if sd.raiseCount >= 1 && !sd.threeBetTracked[playerIdx] {
		sd.players[playerIdx].IncrementThreeBetOpportunity()
		if action == ShortDeckActionRaise || action == ShortDeckActionAllIn {
			sd.players[playerIdx].IncrementThreeBet()
		}
		sd.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats ポストフロップのAFスタッツを追跡
func (sd *ShortDeck) trackPostFlopStats(playerIdx, action int) {
	if sd.phase < ShortDeckPhaseFlop || sd.phase > ShortDeckPhaseRiver {
		return
	}

	switch action {
	case ShortDeckActionBet, ShortDeckActionRaise, ShortDeckActionAllIn:
		sd.players[playerIdx].IncrementPostFlopBetRaise()
	case ShortDeckActionCall:
		sd.players[playerIdx].IncrementPostFlopCall()
	}
}

// executeAction 指定プレイヤーのアクション実行
func (sd *ShortDeck) executeAction(playerIdx, action, amount int) error {
	sd.trackPreFlopStats(playerIdx, action)
	sd.trackPostFlopStats(playerIdx, action)

	bp := toBettingPlayers(sd.players)
	state := sd.bettingState()
	maxRaises, maxBetAmount := sd.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, sd.config.BigBlind, maxRaises, maxBetAmount)
	sd.syncBettingState(state)
	if err != nil {
		return err
	}

	sd.logAction(playerIdx, action, amount)

	if sd.countActivePlayers() == 1 {
		sd.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (sd *ShortDeck) advanceTurn() {
	if sd.gameEndFlag {
		return
	}

	bp := toBettingPlayers(sd.players)
	if sd.isBettingRoundComplete(bp) {
		sd.advancePhase()
		return
	}

	if next := sd.findNextActiveTurn(sd.currentTurn, bp); next >= 0 {
		sd.currentTurn = next
		return
	}

	sd.advancePhase()
}

// advancePhase 次のフェーズに進める
func (sd *ShortDeck) advancePhase() {
	for _, p := range sd.players {
		p.SetCurrentBet(0)
	}
	sd.lastBet = 0
	sd.minRaise = sd.config.BigBlind
	sd.raiseCount = 0
	sd.actedFlags = make([]bool, len(sd.players))
	for i, p := range sd.players {
		if p.GetFolded() || p.GetAllIn() {
			sd.actedFlags[i] = true
		}
	}

	switch sd.phase {
	case ShortDeckPhasePreFlop:
		sd.phase = ShortDeckPhaseFlop
		for i := 0; i < 3; i++ {
			card := sd.trumpCards.DrawCard()
			if card != nil {
				sd.communityCards = append(sd.communityCards, card)
			}
		}
		sd.appendLog(-1, "deal", "dealt flop", sd.communityCards)
	case ShortDeckPhaseFlop:
		sd.phase = ShortDeckPhaseTurn
		card := sd.trumpCards.DrawCard()
		if card != nil {
			sd.communityCards = append(sd.communityCards, card)
		}
		sd.appendLog(-1, "deal", "dealt turn", sd.communityCards[3:])
	case ShortDeckPhaseTurn:
		sd.phase = ShortDeckPhaseRiver
		card := sd.trumpCards.DrawCard()
		if card != nil {
			sd.communityCards = append(sd.communityCards, card)
		}
		sd.appendLog(-1, "deal", "dealt river", sd.communityCards[4:])
	case ShortDeckPhaseRiver:
		sd.phase = ShortDeckPhaseShowdown
		sd.appendLog(-1, "showdown", "showdown", nil)
		sd.resolveShowdown()
		return
	}

	activeCnt := 0
	for _, p := range sd.players {
		if !p.GetFolded() && !p.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		sd.dealRemainingCommunity()
		sd.phase = ShortDeckPhaseShowdown
		sd.resolveShowdown()
		return
	}

	sd.currentTurn = sd.findNextActive(sd.dealerIdx)
}

// dealRemainingCommunity 残りのコミュニティカードを全て配る
func (sd *ShortDeck) dealRemainingCommunity() {
	for len(sd.communityCards) < 5 {
		card := sd.trumpCards.DrawCard()
		if card == nil {
			break
		}
		sd.communityCards = append(sd.communityCards, card)
	}
}

// findNextActive 指定インデックスの次のアクティブプレイヤーを探す
func (sd *ShortDeck) findNextActive(fromIdx int) int {
	return findNextActive(sd.players, fromIdx)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (sd *ShortDeck) countActivePlayers() int {
	return countPlayers(sd.players, func(p *ShortDeckPlayer) bool { return !p.GetFolded() })
}

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (sd *ShortDeck) resolveLastPlayer() {
	for i, p := range sd.players {
		if !p.GetFolded() {
			p.AddChips(sd.pot)
			sd.roundResults = []HoldemResult{{
				PlayerIdx: i,
				WonAmount: sd.pot,
			}}
			sd.pot = 0
			break
		}
	}
	sd.phase = ShortDeckPhaseEnd
	sd.gameEndFlag = true
	sd.dealerIdx = (sd.dealerIdx + 1) % len(sd.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (sd *ShortDeck) resolveShowdown() {
	for _, p := range sd.players {
		if !p.GetFolded() {
			p.EvalBestHand(sd.communityCards)
		}
	}

	bp := toBettingPlayers(sd.players)
	sd.sidePots = CalculateSidePots(bp, sd.pot, sd.startingChips)
	wonAmounts := DistributePots(bp, sd.sidePots)

	sd.roundResults = make([]HoldemResult, 0)
	humanLost := false
	for i, p := range sd.players {
		if p.GetFolded() {
			continue
		}
		result := HoldemResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  sd.getHandName(p.GetHandRank()),
			BestHand:  p.GetBestHand(),
			Kickers:   ExtractKickers(p.GetBestHand(), p.GetHandRank()),
			WonAmount: wonAmounts[i],
		}
		sd.roundResults = append(sd.roundResults, result)
		if p.GetIsHuman() && wonAmounts[i] == 0 {
			humanLost = true
		}
	}

	if humanLost {
		return
	}

	sd.finalizeShowdown()
}

// finalizeShowdown ショーダウンを完了し、END フェーズに遷移する
func (sd *ShortDeck) finalizeShowdown() {
	sd.phase = ShortDeckPhaseEnd
	sd.gameEndFlag = true
	sd.dealerIdx = (sd.dealerIdx + 1) % len(sd.players)
}

// Muck 人間プレイヤーがハンドをマックする
func (sd *ShortDeck) Muck() error {
	if sd.phase != ShortDeckPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Muck is not available now.")
	}
	for i := range sd.roundResults {
		if sd.players[sd.roundResults[i].PlayerIdx].GetIsHuman() {
			sd.roundResults[i].Mucked = true
			break
		}
	}
	sd.finalizeShowdown()
	return nil
}

// ShowHand 人間プレイヤーがハンドを公開する
func (sd *ShortDeck) ShowHand() error {
	if sd.phase != ShortDeckPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	sd.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (sd *ShortDeck) IsMuckAvailable() bool {
	if sd.phase != ShortDeckPhaseShowdown {
		return false
	}
	for _, r := range sd.roundResults {
		if sd.players[r.PlayerIdx].GetIsHuman() && r.WonAmount == 0 {
			return true
		}
	}
	return false
}

// getHandName ハンドランクから名前を返す (ショートデック用)
func (sd *ShortDeck) getHandName(rank int) string {
	if rank >= 0 && rank < len(ShortDeckHandNames) {
		return ShortDeckHandNames[rank]
	}
	return "Unknown"
}

// runCpuActions CPUプレイヤーのアクションを実行
func (sd *ShortDeck) runCpuActions() error {
	if sd.gameEndFlag {
		return nil
	}
	const maxIterations = 500
	iterations := 0
	for !sd.gameEndFlag && sd.phase >= ShortDeckPhasePreFlop && sd.phase <= ShortDeckPhaseRiver {
		iterations++
		if iterations > maxIterations {
			return fmt.Errorf("maxIterations reached in runCpuActions, possible infinite loop")
		}
		if sd.players[sd.currentTurn].GetIsHuman() {
			return nil
		}
		if sd.players[sd.currentTurn].GetFolded() || sd.players[sd.currentTurn].GetAllIn() {
			sd.advanceTurn()
			continue
		}
		action, amount := sd.cpuDecide(sd.currentTurn)
		sd.cpuActions = append(sd.cpuActions, HoldemCpuAction{
			PlayerIdx: sd.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := sd.executeAction(sd.currentTurn, action, amount)
		if err != nil {
			sd.handleCpuActionError(sd.currentTurn, action, err)
		}
		if sd.gameEndFlag {
			return nil
		}
		sd.advanceTurn()
	}
	return nil
}

// handleCpuActionError CPUアクション失敗時のフォールバック処理
func (sd *ShortDeck) handleCpuActionError(playerIdx, action int, err error) {
	sd.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", playerIdx, action, err)
	callAmt := sd.lastBet - sd.players[playerIdx].GetCurrentBet()
	if callAmt > 0 {
		_ = sd.executeAction(playerIdx, ShortDeckActionFold, 0)
	} else {
		_ = sd.executeAction(playerIdx, ShortDeckActionCheck, 0)
	}
}

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (sd *ShortDeck) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(sd.config.BettingLimit, sd.pot, sd.lastBet)
}

// cpuDecide CPUプレイヤーの意思決定
func (sd *ShortDeck) cpuDecide(idx int) (int, int) {
	p := sd.players[idx]
	style := p.GetPlayStyle()
	callAmount := sd.lastBet - p.GetCurrentBet()

	// GTO
	if style == HoldemStyleGTO {
		var action, amount int
		if sd.phase == ShortDeckPhasePreFlop {
			action, amount = sd.cpuDecidePreFlopGTO(idx, callAmount)
		} else {
			action, amount = sd.cpuDecidePostFlopGTO(idx, callAmount)
		}
		maxRaises, maxBetAmount := sd.bettingLimits()
		if maxBetAmount > 0 && amount > maxBetAmount {
			amount = maxBetAmount
		}
		if maxRaises > 0 && sd.raiseCount >= maxRaises {
			if action == ShortDeckActionRaise || action == ShortDeckActionBet {
				if callAmount > 0 {
					return ShortDeckActionCall, 0
				}
				return ShortDeckActionCheck, 0
			}
		}
		return action, amount
	}

	params, ok := shortDeckStyleParamsMap[style]
	if !ok {
		return sd.cpuCallOrCheck(callAmount)
	}

	// メタAI: ブラフ率を調整
	if sd.config.CpuMetaAI && sd.humanProfile != nil {
		adjusted := sd.humanProfile.AdjustedBluffChance(float64(params.bluffRate))
		params.bluffRate = int(math.Round(adjusted))
	}

	var action, amount int
	if sd.phase == ShortDeckPhasePreFlop {
		action, amount = sd.cpuDecidePreFlop(idx, params, callAmount)
	} else {
		action, amount = sd.cpuDecidePostFlop(idx, params, callAmount)
	}

	// メタAI: 人間のベット/レイズに対してコール確率を調整
	if sd.config.CpuMetaAI && sd.humanProfile != nil && sd.lastHumanPlayMs > 0 {
		if action == ShortDeckActionFold && callAmount > 0 {
			handRank := p.EvalBestHand(sd.communityCards)
			bracket := bettingHandBracket(handRank)
			adjustedCall := sd.humanProfile.AdjustedCallChance(0.0, bracket, sd.lastHumanPlayMs)
			if adjustedCall > 0 && rand.Float64() < adjustedCall { //nolint:gosec // non-crypto random for game AI
				action = ShortDeckActionCall
				amount = 0
			}
		}
	}

	maxRaises, maxBetAmount := sd.bettingLimits()

	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}

	if maxRaises > 0 && sd.raiseCount >= maxRaises {
		if action == ShortDeckActionRaise || action == ShortDeckActionBet {
			if callAmount > 0 {
				return ShortDeckActionCall, 0
			}
			return ShortDeckActionCheck, 0
		}
	}
	return action, amount
}

// cpuFoldOrCheck コール額がある場合はフォールド、なければチェック
func (sd *ShortDeck) cpuFoldOrCheck(callAmount int) (int, int) {
	return CpuFoldOrCheck(callAmount)
}

// cpuCallOrCheck コール額がある場合はコール、なければチェック
func (sd *ShortDeck) cpuCallOrCheck(callAmount int) (int, int) {
	return CpuCallOrCheck(callAmount)
}

// cpuRaiseOrBet コール額がある場合はレイズ、なければベット
func (sd *ShortDeck) cpuRaiseOrBet(p *ShortDeckPlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(p.GetChips(), callAmount, raiseAmt)
}

// cpuBetOrAllIn ベットする (チップ不足時はオールイン)
func (sd *ShortDeck) cpuBetOrAllIn(p *ShortDeckPlayer, betAmt int) (int, int) {
	if betAmt > p.GetChips() {
		return ShortDeckActionAllIn, 0
	}
	return ShortDeckActionBet, betAmt
}

// cpuPotBet ポット比率ベースのベット額を計算
func (sd *ShortDeck) cpuPotBet(potPct int) int {
	bet := sd.pot * potPct / 100
	if bet < sd.config.BigBlind {
		bet = sd.config.BigBlind
	}
	if bet < sd.minRaise {
		bet = sd.minRaise
	}
	return bet
}

// cpuDecidePreFlop プリフロップのCPU意思決定
func (sd *ShortDeck) cpuDecidePreFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := sd.players[idx]
	strength := sd.evalPreFlopStrength(idx)

	if strength < params.preFlopFoldThreshold {
		if params.preFlopFoldCompound {
			if callAmount > sd.config.BigBlind*params.preFlopFoldCallMult {
				return ShortDeckActionFold, 0
			}
		} else {
			return sd.cpuFoldOrCheck(callAmount)
		}
	}

	if params.aggressive {
		if strength >= params.preFlopRaiseThreshold || rand.Intn(100) < params.bluffRate {
			return sd.cpuRaiseOrBet(p, callAmount, sd.cpuPotBet(params.preFlopRaisePotPct))
		}
		return sd.cpuCallOrCheck(callAmount)
	}

	if callAmount > 0 {
		return ShortDeckActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return sd.cpuBetOrAllIn(p, sd.cpuPotBet(params.preFlopBluffPotPct))
	}
	return ShortDeckActionCheck, 0
}

// cpuDecidePostFlop フロップ以降のCPU意思決定
func (sd *ShortDeck) cpuDecidePostFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := sd.players[idx]
	handRank := p.EvalBestHand(sd.communityCards)

	if params.aggressive {
		if handRank >= params.postFlopRaiseRank || rand.Intn(100) < params.bluffRate {
			return sd.cpuRaiseOrBet(p, callAmount, sd.cpuPotBet(params.postFlopRaisePotPct))
		}
		if params.postFlopFallbackFold {
			if params.postFlopCondCallRank >= 0 && handRank >= params.postFlopCondCallRank && callAmount > 0 {
				return ShortDeckActionCall, 0
			}
			return sd.cpuFoldOrCheck(callAmount)
		}
		if callAmount > 0 {
			if handRank <= params.postFlopAggrFoldRank && callAmount > sd.config.BigBlind*params.postFlopAggrFoldMult {
				return ShortDeckActionFold, 0
			}
			return ShortDeckActionCall, 0
		}
		return ShortDeckActionCheck, 0
	}

	if callAmount > 0 {
		if handRank <= params.postFlopPassFoldRank {
			if params.postFlopPassFoldMult < 0 || callAmount > sd.config.BigBlind*params.postFlopPassFoldMult {
				return ShortDeckActionFold, 0
			}
		}
		return ShortDeckActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return sd.cpuBetOrAllIn(p, sd.cpuPotBet(params.postFlopBluffPotPct))
	}
	return ShortDeckActionCheck, 0
}

// cpuDecidePreFlopGTO GTOプリフロップ意思決定
func (sd *ShortDeck) cpuDecidePreFlopGTO(idx, callAmount int) (int, int) {
	p := sd.players[idx]
	strength := sd.evalPreFlopStrength(idx)
	dist := gtoPreFlopTable[gtoPreFlopIndex(strength)]
	decision := gtoRollAction(dist)

	switch decision {
	case 0:
		return sd.cpuFoldOrCheck(callAmount)
	case 2:
		betAmt := sd.cpuPotBet(gtoPreFlopBetPct)
		return sd.cpuRaiseOrBet(p, callAmount, betAmt)
	default:
		return sd.cpuCallOrCheck(callAmount)
	}
}

// cpuDecidePostFlopGTO GTOポストフロップ意思決定
func (sd *ShortDeck) cpuDecidePostFlopGTO(idx, callAmount int) (int, int) {
	p := sd.players[idx]
	handRank := p.EvalBestHand(sd.communityCards)
	category := classifyGTOHand(handRank)
	bt := evalBoardTexture(sd.communityCards)

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
		return sd.cpuCallOrCheck(callAmount)
	}

	if bt.highCards >= 3 && decision == 2 && category <= gtoHandWeak && rand.Intn(100) < 40 {
		return sd.cpuCallOrCheck(callAmount)
	}

	switch decision {
	case 0:
		return sd.cpuFoldOrCheck(callAmount)
	case 2:
		betAmt := sd.cpuPotBet(potPct)
		return sd.cpuRaiseOrBet(p, callAmount, betAmt)
	default:
		return sd.cpuCallOrCheck(callAmount)
	}
}

// evalPreFlopStrength プリフロップハンド強度評価 (0-100) for 2-card ShortDeck hands
func (sd *ShortDeck) evalPreFlopStrength(idx int) int {
	p := sd.players[idx]
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

// --- リバイ/アドオン ---

func (sd *ShortDeck) checkAndTransitionAddon() bool {
	if sd.config.AddonEnabled && sd.handCount == sd.config.AddonAfterHand {
		needHumanAddon := false
		for i, p := range sd.players {
			if !sd.addonUsed[i] {
				if p.GetIsHuman() {
					needHumanAddon = true
				} else {
					p.AddChips(sd.config.AddonChips)
					sd.addonUsed[i] = true
				}
			}
		}
		if needHumanAddon {
			sd.phase = ShortDeckPhaseRebuy
			sd.rebuyPhaseType = ShortDeckRebuyPhaseAddon
			return true
		}
	}
	return false
}

// Rebuy 人間プレイヤーがリバイを実行する
func (sd *ShortDeck) Rebuy() error {
	if sd.phase != ShortDeckPhaseRebuy || sd.rebuyPhaseType != ShortDeckRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	for i, p := range sd.players {
		if p.GetIsHuman() && p.GetChips() <= 0 && sd.rebuyCounts[i] < sd.config.RebuyMaxCount {
			p.AddChips(sd.config.RebuyChips)
			sd.rebuyCounts[i]++
			sd.appendLog(i, "rebuy", "rebuy", nil)
			break
		}
	}
	sd.rebuyPhaseType = ShortDeckRebuyPhaseNone
	if sd.checkAndTransitionAddon() {
		return nil
	}
	return sd.continueReset()
}

// SkipRebuy 人間プレイヤーがリバイを辞退する
func (sd *ShortDeck) SkipRebuy() error {
	if sd.phase != ShortDeckPhaseRebuy || sd.rebuyPhaseType != ShortDeckRebuyPhaseRebuy {
		return NewDomainError(ErrWrongPhase, "Rebuy is not available now.")
	}
	sd.rebuyPhaseType = ShortDeckRebuyPhaseNone
	for _, p := range sd.players {
		if p.GetIsHuman() && p.GetChips() <= 0 {
			sd.phase = ShortDeckPhaseEnd
			sd.gameEndFlag = true
			return nil
		}
	}
	if sd.checkAndTransitionAddon() {
		return nil
	}
	return sd.continueReset()
}

// Addon 人間プレイヤーがアドオンを実行する
func (sd *ShortDeck) Addon() error {
	if sd.phase != ShortDeckPhaseRebuy || sd.rebuyPhaseType != ShortDeckRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	for i, p := range sd.players {
		if p.GetIsHuman() && !sd.addonUsed[i] {
			p.AddChips(sd.config.AddonChips)
			sd.addonUsed[i] = true
			break
		}
	}
	sd.rebuyPhaseType = ShortDeckRebuyPhaseNone
	return sd.continueReset()
}

// SkipAddon 人間プレイヤーがアドオンを辞退する
func (sd *ShortDeck) SkipAddon() error {
	if sd.phase != ShortDeckPhaseRebuy || sd.rebuyPhaseType != ShortDeckRebuyPhaseAddon {
		return NewDomainError(ErrWrongPhase, "Addon is not available now.")
	}
	sd.rebuyPhaseType = ShortDeckRebuyPhaseNone
	return sd.continueReset()
}

// IsRebuyAvailable 人間プレイヤーがリバイ可能かどうか
func (sd *ShortDeck) IsRebuyAvailable() bool {
	return rebuyAvailable(sd.config.RebuyEnabled, sd.handCount, sd.config.RebuyPeriodHands, sd.players, sd.rebuyCounts, sd.config.RebuyMaxCount)
}

// IsAddonAvailable 人間プレイヤーがアドオン可能かどうか
func (sd *ShortDeck) IsAddonAvailable() bool {
	return addonAvailable(sd.config.AddonEnabled, sd.handCount, sd.config.AddonAfterHand, sd.players, sd.addonUsed)
}

// GetRebuyCounts プレイヤーごとのリバイ回数取得
func (sd *ShortDeck) GetRebuyCounts() []int {
	return copyOf(sd.rebuyCounts)
}

// GetAddonUsed プレイヤーごとのアドオン使用フラグ取得
func (sd *ShortDeck) GetAddonUsed() []bool {
	return copyOf(sd.addonUsed)
}

// GetRebuyPhaseType リバイフェーズ種別取得
func (sd *ShortDeck) GetRebuyPhaseType() int { return sd.rebuyPhaseType }

// GetEquity エクイティ計算結果を返す
func (sd *ShortDeck) GetEquity() *HoldemEquityResult {
	if sd.phase < ShortDeckPhasePreFlop || sd.phase > ShortDeckPhaseRiver {
		return nil
	}
	var humanPlayer *ShortDeckPlayer
	for _, p := range sd.players {
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
	for _, p := range sd.players {
		if !p.GetIsHuman() && !p.GetFolded() {
			activePlayers++
		}
	}
	result := CalcShortDeckEquity(humanCards, sd.communityCards, activePlayers, shortDeckEquitySimulations, nil)
	return &result
}

// GetPotOdds ポットオッズを返す
func (sd *ShortDeck) GetPotOdds() float64 {
	if sd.phase < ShortDeckPhasePreFlop || sd.phase > ShortDeckPhaseRiver {
		return 0.0
	}
	humanCurrentBet := 0
	for _, p := range sd.players {
		if p.GetIsHuman() {
			humanCurrentBet = p.GetCurrentBet()
			break
		}
	}
	callAmount := sd.lastBet - humanCurrentBet
	if callAmount < 0 {
		callAmount = 0
	}
	return CalcPotOdds(sd.pot, callAmount)
}

// --- ゲッター ---

// GetPhase フェーズ取得
func (sd *ShortDeck) GetPhase() int { return sd.phase }

// GetPlayers プレイヤー一覧取得
func (sd *ShortDeck) GetPlayers() []*ShortDeckPlayer { return sd.players }

// GetPlayer 指定プレイヤー取得
func (sd *ShortDeck) GetPlayer(i int) *ShortDeckPlayer {
	if i >= 0 && i < len(sd.players) {
		return sd.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (sd *ShortDeck) GetPlayerCnt() int { return len(sd.players) }

// GetCommunityCards コミュニティカード取得
func (sd *ShortDeck) GetCommunityCards() []*Card { return sd.communityCards }

// GetPot ポット取得
func (sd *ShortDeck) GetPot() int { return sd.pot }

// GetSidePots サイドポット取得
func (sd *ShortDeck) GetSidePots() []SidePot { return sd.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (sd *ShortDeck) GetDealerIdx() int { return sd.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (sd *ShortDeck) GetCurrentTurn() int { return sd.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (sd *ShortDeck) GetGameEndFlag() bool { return sd.gameEndFlag }

// GetLastBet 最後のベット取得
func (sd *ShortDeck) GetLastBet() int { return sd.lastBet }

// GetMinRaise 最小レイズ額取得
func (sd *ShortDeck) GetMinRaise() int { return sd.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (sd *ShortDeck) GetRaiseCount() int { return sd.raiseCount }

// GetRoundResults ラウンド結果取得
func (sd *ShortDeck) GetRoundResults() []HoldemResult { return sd.roundResults }

// GetCpuActions CPU行動記録取得
func (sd *ShortDeck) GetCpuActions() []HoldemCpuAction { return sd.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得
func (sd *ShortDeck) GetLastCpuError() error { return sd.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (sd *ShortDeck) GetHumanProfile() *BettingHumanProfile { return sd.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (sd *ShortDeck) ResetProfile() { sd.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする (プロファイルがない場合はnil)
func (sd *ShortDeck) ExportProfile() interface{} {
	if sd.humanProfile == nil {
		return nil
	}
	d := sd.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (sd *ShortDeck) ImportProfile(data []byte) error {
	p, err := importBettingProfile(data)
	if err != nil || p == nil {
		return err
	}
	sd.humanProfile = p
	return nil
}

// GetConfig 設定取得
func (sd *ShortDeck) GetConfig() ShortDeckConfig { return sd.config }

// SetConfig 設定変更
func (sd *ShortDeck) SetConfig(cfg ShortDeckConfig) { sd.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (sd *ShortDeck) IsHumanTurn() bool {
	return isHumanTurn(sd.players, sd.currentTurn)
}

// GetActedFlags actedフラグ取得
func (sd *ShortDeck) GetActedFlags() []bool {
	return copyOf(sd.actedFlags)
}

// GetHandCount ハンド数取得
func (sd *ShortDeck) GetHandCount() int { return sd.handCount }

// logAction ベッティングアクションを棋譜に記録する
func (sd *ShortDeck) logAction(playerIdx, action, amount int) {
	switch action {
	case ShortDeckActionFold:
		sd.appendLog(playerIdx, "fold", "fold", nil)
	case ShortDeckActionCheck:
		sd.appendLog(playerIdx, "check", "check", nil)
	case ShortDeckActionCall:
		sd.appendLog(playerIdx, "call", fmt.Sprintf("call %d", sd.players[playerIdx].GetCurrentBet()), nil)
	case ShortDeckActionBet:
		sd.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case ShortDeckActionRaise:
		sd.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case ShortDeckActionAllIn:
		sd.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", sd.players[playerIdx].GetCurrentBet()), nil)
	}
}

// shortDeckJSON is the JSON wire format for ShortDeck.
type shortDeckJSON struct {
	TrumpCards      *TrumpCards              `json:"tc"`
	Players         []*ShortDeckPlayer       `json:"pl"`
	CommunityCards  []*Card                  `json:"cc"`
	Pot             int                      `json:"pt"`
	SidePots        []SidePot                `json:"sp"`
	DealerIdx       int                      `json:"di"`
	CurrentTurn     int                      `json:"ct"`
	Phase           int                      `json:"ph"`
	Config          ShortDeckConfig          `json:"cf"`
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
}

// shortDeckMaxSliceLen caps slice sizes during deserialisation.
const shortDeckMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (sd *ShortDeck) MarshalJSON() ([]byte, error) {
	j := shortDeckJSON{
		TrumpCards:      sd.trumpCards,
		Players:         sd.players,
		CommunityCards:  sd.communityCards,
		Pot:             sd.pot,
		SidePots:        sd.sidePots,
		DealerIdx:       sd.dealerIdx,
		CurrentTurn:     sd.currentTurn,
		Phase:           sd.phase,
		Config:          sd.config,
		GameEndFlag:     sd.gameEndFlag,
		LastBet:         sd.lastBet,
		MinRaise:        sd.minRaise,
		RaiseCount:      sd.raiseCount,
		ActedFlags:      sd.actedFlags,
		RoundResults:    sd.roundResults,
		CpuActions:      sd.cpuActions,
		StartingChips:   sd.startingChips,
		VPIPTracked:     sd.vpipTracked,
		PFRTracked:      sd.pfrTracked,
		ThreeBetTracked: sd.threeBetTracked,
		HandCount:       sd.handCount,
		RebuyCounts:     sd.rebuyCounts,
		AddonUsed:       sd.addonUsed,
		RebuyPhaseType:  sd.rebuyPhaseType,
		ActionLog:       sd.actionLog,
		LastHumanPlayMs: sd.lastHumanPlayMs,
	}
	if sd.humanProfile != nil {
		d := sd.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (sd *ShortDeck) UnmarshalJSON(data []byte) error {
	var j shortDeckJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > shortDeckMaxSliceLen || len(j.CommunityCards) > shortDeckMaxSliceLen ||
		len(j.SidePots) > shortDeckMaxSliceLen || len(j.ActedFlags) > shortDeckMaxSliceLen ||
		len(j.RoundResults) > shortDeckMaxSliceLen || len(j.CpuActions) > shortDeckMaxSliceLen ||
		len(j.StartingChips) > shortDeckMaxSliceLen || len(j.ActionLog) > shortDeckMaxSliceLen {
		return fmt.Errorf("shortdeck: input array exceeds maximum allowed size")
	}
	sd.trumpCards = j.TrumpCards
	if sd.trumpCards == nil {
		sd.trumpCards = NewTrumpCards(0)
	}
	sd.players = j.Players
	if sd.players == nil {
		sd.players = make([]*ShortDeckPlayer, 0)
	}
	sd.communityCards = j.CommunityCards
	if sd.communityCards == nil {
		sd.communityCards = make([]*Card, 0)
	}
	sd.pot = j.Pot
	sd.sidePots = j.SidePots
	if sd.sidePots == nil {
		sd.sidePots = make([]SidePot, 0)
	}
	sd.dealerIdx = j.DealerIdx
	sd.currentTurn = j.CurrentTurn
	sd.phase = j.Phase
	sd.config = j.Config
	sd.gameEndFlag = j.GameEndFlag
	sd.lastBet = j.LastBet
	sd.minRaise = j.MinRaise
	sd.raiseCount = j.RaiseCount
	sd.actedFlags = j.ActedFlags
	if sd.actedFlags == nil {
		sd.actedFlags = make([]bool, 0)
	}
	sd.roundResults = j.RoundResults
	if sd.roundResults == nil {
		sd.roundResults = make([]HoldemResult, 0)
	}
	sd.cpuActions = j.CpuActions
	if sd.cpuActions == nil {
		sd.cpuActions = make([]HoldemCpuAction, 0)
	}
	sd.startingChips = j.StartingChips
	if sd.startingChips == nil {
		sd.startingChips = make([]int, 0)
	}
	sd.vpipTracked = j.VPIPTracked
	if sd.vpipTracked == nil {
		sd.vpipTracked = make([]bool, 0)
	}
	sd.pfrTracked = j.PFRTracked
	if sd.pfrTracked == nil {
		sd.pfrTracked = make([]bool, 0)
	}
	sd.threeBetTracked = j.ThreeBetTracked
	if sd.threeBetTracked == nil {
		sd.threeBetTracked = make([]bool, 0)
	}
	sd.handCount = j.HandCount
	sd.rebuyCounts = j.RebuyCounts
	if sd.rebuyCounts == nil {
		sd.rebuyCounts = make([]int, 0)
	}
	sd.addonUsed = j.AddonUsed
	if sd.addonUsed == nil {
		sd.addonUsed = make([]bool, 0)
	}
	sd.rebuyPhaseType = j.RebuyPhaseType
	sd.actionLog = j.ActionLog
	if sd.actionLog == nil {
		sd.actionLog = make([]*ActionLogEntry, 0)
	}
	sd.lastHumanPlayMs = j.LastHumanPlayMs
	if j.Profile != nil {
		sd.humanProfile = &BettingHumanProfile{}
		sd.humanProfile.Import(*j.Profile)
	}
	return nil
}

// Resize プレイヤースライスを差し替え、プレイヤー数依存スライスを再初期化する
func (sd *ShortDeck) Resize(players []*ShortDeckPlayer) {
	sd.players = players
	n := len(players)
	sd.actedFlags = make([]bool, n)
	sd.startingChips = make([]int, n)
	sd.vpipTracked = make([]bool, n)
	sd.pfrTracked = make([]bool, n)
	sd.threeBetTracked = make([]bool, n)
	sd.initTournamentState(n)
}
