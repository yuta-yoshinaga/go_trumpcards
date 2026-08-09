//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// フェーズ定数
const (
	FiveCardStudPhaseInit         = 0 // 初期状態
	FiveCardStudPhaseSecondStreet = 1 // セカンドストリート (1 down + 1 up + bring-in + betting)
	FiveCardStudPhaseThirdStreet  = 2 // サードストリート (1 up + betting)
	FiveCardStudPhaseFourthStreet = 3 // フォースストリート (1 up + betting)
	FiveCardStudPhaseFifthStreet  = 4 // フィフスストリート (1 up + betting)
	FiveCardStudPhaseShowdown     = 5 // ショーダウン
	FiveCardStudPhaseEnd          = 6 // ゲーム終了
	FiveCardStudPhaseRebuy        = 7 // リバイ/アドオン待ち
)

// アクション定数 (共通定数のエイリアス)
const (
	FiveCardStudActionFold  = bettingActionFold
	FiveCardStudActionCheck = bettingActionCheck
	FiveCardStudActionCall  = bettingActionCall
	FiveCardStudActionBet   = bettingActionBet
	FiveCardStudActionRaise = bettingActionRaise
	FiveCardStudActionAllIn = bettingActionAllIn
)

// FiveCardStudResult ショーダウン結果
type FiveCardStudResult struct {
	PlayerIdx int     // プレイヤーインデックス
	HandRank  int     // ハンドランク
	HandName  string  // ハンド名
	BestHand  []*Card // ベスト5枚
	Kickers   []int   // キッカーカード値
	WonAmount int     // 獲得チップ
	Mucked    bool    // マックしたかどうか
}

// FiveCardStudCpuAction CPU行動記録
type FiveCardStudCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// リバイフェーズ種別定数
const (
	FiveCardStudRebuyPhaseNone  = 0
	FiveCardStudRebuyPhaseRebuy = 1
	FiveCardStudRebuyPhaseAddon = 2
)

// errFiveCardStudInvalid is the shared sentinel error returned for all
// UnmarshalJSON validation failures (kept single to minimise worker binary size).
var errFiveCardStudInvalid = fmt.Errorf("fivecardstud: invalid game state")

// FiveCardStud ファイブカードスタッド
type FiveCardStud struct {
	trumpCards      *TrumpCards
	players         []*FiveCardStudPlayer
	communityCard   *Card // カード不足時の共有カード
	pot             int
	sidePots        []SidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	config          FiveCardStudConfig
	gameEndFlag     bool
	lastBet         int
	minRaise        int
	raiseCount      int
	actedFlags      []bool
	roundResults    []FiveCardStudResult
	cpuActions      []FiveCardStudCpuAction
	startingChips   []int
	vpipTracked     []bool
	pfrTracked      []bool
	threeBetTracked []bool
	tournamentBase  // handCount / rebuyCounts / addonUsed (issue #1463)
	lastCpuError    error
	rebuyPhaseType  int
	actionLogBase
	humanProfile     *BettingHumanProfile
	lastHumanPlayMs  int
	bringInPlayerIdx int // ブリングインプレイヤーインデックス
}

// NewFiveCardStud コンストラクタ
func NewFiveCardStud(trumpCards *TrumpCards, players []*FiveCardStudPlayer, config FiveCardStudConfig) *FiveCardStud {
	n := len(players)
	s := &FiveCardStud{
		trumpCards:       trumpCards,
		players:          players,
		sidePots:         make([]SidePot, 0),
		actedFlags:       make([]bool, n),
		roundResults:     make([]FiveCardStudResult, 0),
		cpuActions:       make([]FiveCardStudCpuAction, 0),
		startingChips:    make([]int, n),
		vpipTracked:      make([]bool, n),
		pfrTracked:       make([]bool, n),
		threeBetTracked:  make([]bool, n),
		config:           config,
		phase:            FiveCardStudPhaseInit,
		bringInPlayerIdx: -1,
	}
	s.initTournamentState(n)
	return s
}

// NewDefaultFiveCardStud returns FiveCardStud with the default table size and
// DefaultFiveCardStudConfig. Used as the single source of truth for CUI, Web,
// and Worker construction sites.
func NewDefaultFiveCardStud() *FiveCardStud {
	cfg := DefaultFiveCardStudConfig()
	return NewFiveCardStud(NewTrumpCards(0), NewFiveCardStudPlayersForTable(cfg.TableSize), cfg)
}

// Reset ゲーム初期化
func (s *FiveCardStud) Reset() error {
	s.phase = FiveCardStudPhaseInit
	s.pot = 0
	s.sidePots = make([]SidePot, 0)
	s.communityCard = nil
	s.gameEndFlag = false
	s.lastBet = 0
	s.minRaise = s.config.SmallBet
	s.raiseCount = 0
	s.actedFlags = make([]bool, len(s.players))
	s.roundResults = make([]FiveCardStudResult, 0)
	s.cpuActions = make([]FiveCardStudCpuAction, 0)
	s.rebuyPhaseType = FiveCardStudRebuyPhaseNone
	s.actionLog = nil
	s.lastHumanPlayMs = 0
	s.bringInPlayerIdx = -1

	// メタAI
	if s.config.CpuMetaAI {
		if s.humanProfile != nil {
			s.humanProfile.GamesPlayed++
		} else {
			s.humanProfile = &BettingHumanProfile{}
		}
	}

	s.trumpCards.Shuffle()
	for _, p := range s.players {
		p.ClearCards()
		p.Reset()
		p.SetFolded(false)
		p.SetAllIn(false)
		p.SetCurrentBet(0)
		p.handRank = 0
		p.bestHand = nil
		if p.GetChips() <= 0 && !s.config.RebuyEnabled {
			p.SetChips(s.config.InitChips)
		}
		p.IncrementTotalHands()
	}

	// HUDスタッツ追跡フラグリセット
	s.vpipTracked = make([]bool, len(s.players))
	s.pfrTracked = make([]bool, len(s.players))
	s.threeBetTracked = make([]bool, len(s.players))

	// トーナメントモード: アンティエスカレーション
	if s.config.TournamentMode && s.config.AnteLevelHands > 0 && s.handCount > 0 && s.handCount%s.config.AnteLevelHands == 0 {
		s.config.Ante = s.config.Ante * s.config.AnteMultiplier / 100
		s.config.BringIn = s.config.BringIn * s.config.AnteMultiplier / 100
		s.config.SmallBet = s.config.SmallBet * s.config.AnteMultiplier / 100
		s.config.BigBet = s.config.BigBet * s.config.AnteMultiplier / 100
		if s.config.Ante < 1 {
			s.config.Ante = 1
		}
		if s.config.BringIn < 1 {
			s.config.BringIn = 1
		}
		if s.config.SmallBet < 1 {
			s.config.SmallBet = 1
		}
		if s.config.BigBet < 1 {
			s.config.BigBet = 1
		}
	}
	s.handCount++

	// リバイチェック
	if s.config.RebuyEnabled && s.handCount <= s.config.RebuyPeriodHands {
		needHumanRebuy := false
		for i, p := range s.players {
			if p.GetChips() <= 0 && s.rebuyCounts[i] < s.config.RebuyMaxCount {
				if p.GetIsHuman() {
					needHumanRebuy = true
				} else {
					p.AddChips(s.config.RebuyChips)
					s.rebuyCounts[i]++
				}
			}
		}
		if needHumanRebuy {
			s.phase = FiveCardStudPhaseRebuy
			s.rebuyPhaseType = FiveCardStudRebuyPhaseRebuy
			return nil
		}
	}

	// アドオンチェック
	if s.checkAndTransitionAddon() {
		return nil
	}

	return s.continueReset()
}

// continueReset ディール以降のリセット処理
func (s *FiveCardStud) continueReset() error {
	// ハンド開始時のチップを記録
	s.startingChips = make([]int, len(s.players))
	for i, p := range s.players {
		s.startingChips[i] = p.GetChips()
	}

	// アンティ投入
	s.postAntes()

	// セカンドストリート配布: 1枚伏せ + 1枚表
	for round := 0; round < 2; round++ {
		for j := 0; j < len(s.players); j++ {
			idx := (s.dealerIdx + 1 + j) % len(s.players)
			card := s.trumpCards.DrawCard()
			if card == nil {
				break
			}
			if round < 1 {
				s.players[idx].AddHoleCard(card)
			} else {
				s.players[idx].AddDoorCard(card)
			}
		}
	}

	s.phase = FiveCardStudPhaseSecondStreet

	// ブリングイン決定
	s.bringInPlayerIdx = s.determineBringIn()
	s.postBringIn()

	// ブリングインプレイヤーの次から開始
	s.currentTurn = (s.bringInPlayerIdx + 1) % len(s.players)

	// CPUアクション実行
	if err := s.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postAntes 全プレイヤーにアンティを投入させる
func (s *FiveCardStud) postAntes() {
	for i, p := range s.players {
		ante := s.config.Ante
		if p.GetChips() < ante {
			ante = p.GetChips()
		}
		p.SubtractChips(ante)
		s.pot += ante
		s.appendLog(i, "ante", fmt.Sprintf("posts ante %d", ante), nil)
		if p.GetChips() == 0 {
			p.SetAllIn(true)
			s.actedFlags[i] = true
		}
	}
}

// determineBringIn ブリングインプレイヤーを決定する
// 最も低いドアカードを持つプレイヤー。
// 同値の場合はスートランキングで決定 (クラブ < ダイヤ < ハート < スペード)。
func (s *FiveCardStud) determineBringIn() int {
	bestIdx := 0
	bestVal := 999
	bestSuit := 999

	for i, p := range s.players {
		if len(p.GetDoorCards()) == 0 {
			continue
		}
		door := p.GetDoorCards()[0]
		val := door.GetValue()
		if val == 1 {
			val = 14 // Ace is high
		}
		suit := SuitRank(door.GetDesign())

		// 通常: 最も低いカードがブリングイン
		if val < bestVal || (val == bestVal && suit < bestSuit) {
			bestIdx = i
			bestVal = val
			bestSuit = suit
		}
	}
	return bestIdx
}

// postBringIn ブリングインを投入する
func (s *FiveCardStud) postBringIn() {
	p := s.players[s.bringInPlayerIdx]
	bringIn := s.config.BringIn
	if p.GetChips() < bringIn {
		bringIn = p.GetChips()
	}
	p.SubtractChips(bringIn)
	p.SetCurrentBet(bringIn)
	s.pot += bringIn
	s.lastBet = bringIn
	s.appendLog(s.bringInPlayerIdx, "bringin", fmt.Sprintf("brings in %d", bringIn), nil)

	if p.GetChips() == 0 {
		p.SetAllIn(true)
		s.actedFlags[s.bringInPlayerIdx] = true
	}
	// ブリングインプレイヤーは行動済み
	s.actedFlags[s.bringInPlayerIdx] = true
}

// PlayerAction 人間プレイヤーのアクション実行
func (s *FiveCardStud) PlayerAction(action, amount, humanPlayMs int) error {
	if s.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if s.phase < FiveCardStudPhaseSecondStreet || s.phase > FiveCardStudPhaseFifthStreet {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !s.players[s.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// メタAI
	s.lastHumanPlayMs = humanPlayMs
	if s.config.CpuMetaAI && s.humanProfile != nil {
		pl := s.players[s.currentTurn]
		handRank := pl.EvalBestHand()
		s.humanProfile.RecordAction(handRank, action)
		s.humanProfile.RecordHesitation(humanPlayMs)
		if s.lastBet > pl.GetCurrentBet() {
			s.humanProfile.RecordFoldToBet(action == FiveCardStudActionFold)
		}
	}

	err := s.executeAction(s.currentTurn, action, amount)
	if err != nil {
		return err
	}

	s.advanceTurn()
	return s.runCpuActions()
}

// bettingPlayers BettingPlayerスライスを生成
func (s *FiveCardStud) bettingPlayers() []BettingPlayer {
	bp := make([]BettingPlayer, len(s.players))
	for i, pl := range s.players {
		bp[i] = pl
	}
	return bp
}

// executeAction 指定プレイヤーのアクション実行
func (s *FiveCardStud) executeAction(playerIdx, action, amount int) error {
	s.trackPreFlopStats(playerIdx, action)
	s.trackPostFlopStats(playerIdx, action)

	bp := s.bettingPlayers()
	state := &BettingState{
		Pot: s.pot, LastBet: s.lastBet, MinRaise: s.minRaise,
		RaiseCount: s.raiseCount, ActedFlags: s.actedFlags,
	}
	maxRaises, maxBetAmount := s.bettingLimits()
	betSize := s.currentBetSize()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, betSize, maxRaises, maxBetAmount)
	s.pot = state.Pot
	s.lastBet = state.LastBet
	s.minRaise = state.MinRaise
	s.raiseCount = state.RaiseCount
	if err != nil {
		return err
	}

	s.logAction(playerIdx, action, amount)

	if s.countActivePlayers() == 1 {
		s.resolveLastPlayer()
	}
	return nil
}

// currentBetSize 現在のストリートのベットサイズを返す
func (s *FiveCardStud) currentBetSize() int {
	if s.phase >= FiveCardStudPhaseFourthStreet {
		return s.config.BigBet
	}
	return s.config.SmallBet
}

// advanceTurn 次のプレイヤーに進める
func (s *FiveCardStud) advanceTurn() {
	if s.gameEndFlag {
		return
	}

	if s.isBettingRoundComplete() {
		s.advancePhase()
		return
	}

	for i := 1; i <= len(s.players); i++ {
		next := (s.currentTurn + i) % len(s.players)
		if !s.players[next].GetFolded() && !s.players[next].GetAllIn() && !s.actedFlags[next] {
			s.currentTurn = next
			return
		}
	}

	s.advancePhase()
}

// isBettingRoundComplete ベッティングラウンドが完了したかチェック
func (s *FiveCardStud) isBettingRoundComplete() bool {
	for i, p := range s.players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if !s.actedFlags[i] {
			return false
		}
	}
	return true
}

// advancePhase 次のフェーズに進める
func (s *FiveCardStud) advancePhase() {
	// ラウンドベットリセット
	for _, p := range s.players {
		p.SetCurrentBet(0)
	}
	s.lastBet = 0
	s.raiseCount = 0
	s.actedFlags = make([]bool, len(s.players))
	for i, p := range s.players {
		if p.GetFolded() || p.GetAllIn() {
			s.actedFlags[i] = true
		}
	}

	switch s.phase {
	case FiveCardStudPhaseSecondStreet:
		s.phase = FiveCardStudPhaseThirdStreet
		s.minRaise = s.config.SmallBet
		s.dealStreetCard(true) // 表向き
		s.appendLog(-1, "deal", "dealt third street", nil)
	case FiveCardStudPhaseThirdStreet:
		s.phase = FiveCardStudPhaseFourthStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(true)
		s.appendLog(-1, "deal", "dealt fourth street", nil)
	case FiveCardStudPhaseFourthStreet:
		s.phase = FiveCardStudPhaseFifthStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(true)
		s.appendLog(-1, "deal", "dealt fifth street", nil)
	case FiveCardStudPhaseFifthStreet:
		s.phase = FiveCardStudPhaseShowdown
		s.appendLog(-1, "showdown", "showdown", nil)
		s.resolveShowdown()
		return
	}

	// アクティブプレイヤーが0-1人ならショーダウンへ
	activeCnt := 0
	for _, p := range s.players {
		if !p.GetFolded() && !p.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		// 残りのカードを配ってショーダウン
		s.dealRemainingStreets()
		s.phase = FiveCardStudPhaseShowdown
		s.resolveShowdown()
		return
	}

	// 3rd Street以降: 最も強い表向き手を持つプレイヤーから開始
	s.currentTurn = s.determineBettingLeader()
}

// dealStreetCard 各アクティブプレイヤーにカードを1枚配る
func (s *FiveCardStud) dealStreetCard(faceUp bool) {
	activePlayers := 0
	for _, p := range s.players {
		if !p.GetFolded() {
			activePlayers++
		}
	}

	remaining := s.trumpCards.GetRemainingCount()
	if remaining < activePlayers && remaining > 0 {
		// カード不足: 共有カードとして配る
		card := s.trumpCards.DrawCard()
		if card != nil {
			s.communityCard = card
		}
		return
	}

	for j := 0; j < len(s.players); j++ {
		idx := (s.dealerIdx + 1 + j) % len(s.players)
		if s.players[idx].GetFolded() {
			continue
		}
		card := s.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if faceUp {
			s.players[idx].AddDoorCard(card)
		} else {
			s.players[idx].AddHoleCard(card)
		}
	}
}

// dealRemainingStreets 残りのストリートのカードを全て配る
func (s *FiveCardStud) dealRemainingStreets() {
	// 現在のフェーズから5th streetまでのカードを配る (全て表向き)
	for phase := s.phase; phase <= FiveCardStudPhaseFifthStreet; phase++ {
		switch phase {
		case FiveCardStudPhaseThirdStreet, FiveCardStudPhaseFourthStreet, FiveCardStudPhaseFifthStreet:
			s.dealStreetCard(true)
		}
	}
}

// determineBettingLeader ベッティングリーダーを返す
// 最も強い表向き手を持つアクティブプレイヤー。
func (s *FiveCardStud) determineBettingLeader() int {
	bestIdx := -1
	for i, p := range s.players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if bestIdx == -1 {
			bestIdx = i
			continue
		}
		if compareVisibleHandsFcs(p, s.players[bestIdx]) > 0 {
			bestIdx = i
		}
	}
	if bestIdx == -1 {
		// フォールバック: フォールドしていない最初のプレイヤー
		for i, p := range s.players {
			if !p.GetFolded() {
				return i
			}
		}
		return 0
	}
	return bestIdx
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (s *FiveCardStud) countActivePlayers() int {
	cnt := 0
	for _, p := range s.players {
		if !p.GetFolded() {
			cnt++
		}
	}
	return cnt
}

// bettingLimits ベッティングリミット設定
func (s *FiveCardStud) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(s.config.BettingLimit, s.pot, s.lastBet)
}

// --- HUDスタッツ ---

// trackPreFlopStats セカンドストリートのHUDスタッツを追跡
func (s *FiveCardStud) trackPreFlopStats(playerIdx, action int) {
	if s.phase != FiveCardStudPhaseSecondStreet {
		return
	}
	isVPIPAction := false
	isPFRAction := false
	switch action {
	case FiveCardStudActionCall:
		isVPIPAction = true
	case FiveCardStudActionBet, FiveCardStudActionRaise, FiveCardStudActionAllIn:
		isVPIPAction = true
		isPFRAction = true
	}
	if isVPIPAction && !s.vpipTracked[playerIdx] {
		s.players[playerIdx].IncrementVPIP()
		s.vpipTracked[playerIdx] = true
	}
	if isPFRAction && !s.pfrTracked[playerIdx] {
		s.players[playerIdx].IncrementPFR()
		s.pfrTracked[playerIdx] = true
	}
	if s.raiseCount >= 1 && !s.threeBetTracked[playerIdx] {
		s.players[playerIdx].IncrementThreeBetOpportunity()
		if action == FiveCardStudActionRaise || action == FiveCardStudActionAllIn {
			s.players[playerIdx].IncrementThreeBet()
		}
		s.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats 3rd Street以降のAFスタッツを追跡
func (s *FiveCardStud) trackPostFlopStats(playerIdx, action int) {
	if s.phase < FiveCardStudPhaseThirdStreet || s.phase > FiveCardStudPhaseFifthStreet {
		return
	}
	switch action {
	case FiveCardStudActionBet, FiveCardStudActionRaise, FiveCardStudActionAllIn:
		s.players[playerIdx].IncrementPostFlopBetRaise()
	case FiveCardStudActionCall:
		s.players[playerIdx].IncrementPostFlopCall()
	}
}

// --- ショーダウン ---

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (s *FiveCardStud) resolveLastPlayer() {
	for i, p := range s.players {
		if !p.GetFolded() {
			p.AddChips(s.pot)
			s.roundResults = []FiveCardStudResult{{
				PlayerIdx: i,
				WonAmount: s.pot,
			}}
			s.pot = 0
			break
		}
	}
	s.phase = FiveCardStudPhaseEnd
	s.gameEndFlag = true
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (s *FiveCardStud) resolveShowdown() {
	// ハンド評価 (共有カードがある場合はそれも含める)
	for _, p := range s.players {
		if !p.GetFolded() {
			if s.communityCard != nil {
				p.AddHoleCard(s.communityCard)
			}
			p.EvalBestHand()
			if s.communityCard != nil {
				// 一時的に追加した共有カードを除去
				p.holeCards = p.holeCards[:len(p.holeCards)-1]
			}
		}
	}

	bp := s.bettingPlayers()
	s.sidePots = CalculateSidePots(bp, s.pot, s.startingChips)
	wonAmounts := DistributePots(bp, s.sidePots)

	s.roundResults = make([]FiveCardStudResult, 0)
	humanLost := false
	for i, p := range s.players {
		if p.GetFolded() {
			continue
		}
		result := FiveCardStudResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  s.getHandName(p.GetHandRank()),
			BestHand:  p.GetBestHand(),
			Kickers:   ExtractKickers(p.GetBestHand(), p.GetHandRank()),
			WonAmount: wonAmounts[i],
		}
		s.roundResults = append(s.roundResults, result)
		if p.GetIsHuman() && wonAmounts[i] == 0 {
			humanLost = true
		}
	}

	if humanLost {
		return
	}
	s.finalizeShowdown()
}

// finalizeShowdown ショーダウンを完了しENDフェーズに遷移する
func (s *FiveCardStud) finalizeShowdown() {
	s.phase = FiveCardStudPhaseEnd
	s.gameEndFlag = true
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
}

// Muck 人間プレイヤーがハンドをマックする
func (s *FiveCardStud) Muck() error {
	if s.phase != FiveCardStudPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Muck is not available now.")
	}
	for i := range s.roundResults {
		if s.players[s.roundResults[i].PlayerIdx].GetIsHuman() {
			s.roundResults[i].Mucked = true
			break
		}
	}
	s.finalizeShowdown()
	return nil
}

// ShowHand 人間プレイヤーがハンドを公開する
func (s *FiveCardStud) ShowHand() error {
	if s.phase != FiveCardStudPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	s.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (s *FiveCardStud) IsMuckAvailable() bool {
	if s.phase != FiveCardStudPhaseShowdown {
		return false
	}
	for _, r := range s.roundResults {
		if s.players[r.PlayerIdx].GetIsHuman() && r.WonAmount == 0 {
			return true
		}
	}
	return false
}

// getHandName ハンドランクから名前を返す
func (s *FiveCardStud) getHandName(rank int) string {
	if rank >= 0 && rank < len(PokerHandNames) {
		return PokerHandNames[rank]
	}
	return "Unknown"
}

// --- 棋譜 ---

func (s *FiveCardStud) logAction(playerIdx, action, amount int) {
	switch action {
	case FiveCardStudActionFold:
		s.appendLog(playerIdx, "fold", "fold", nil)
	case FiveCardStudActionCheck:
		s.appendLog(playerIdx, "check", "check", nil)
	case FiveCardStudActionCall:
		s.appendLog(playerIdx, "call", fmt.Sprintf("call %d", s.players[playerIdx].GetCurrentBet()), nil)
	case FiveCardStudActionBet:
		s.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case FiveCardStudActionRaise:
		s.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case FiveCardStudActionAllIn:
		s.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", s.players[playerIdx].GetCurrentBet()), nil)
	}
}

// --- ゲッター ---

// GetPhase フェーズ取得
func (s *FiveCardStud) GetPhase() int { return s.phase }

// GetPlayers プレイヤー一覧取得
func (s *FiveCardStud) GetPlayers() []*FiveCardStudPlayer { return s.players }

// GetPlayer 指定プレイヤー取得
func (s *FiveCardStud) GetPlayer(i int) *FiveCardStudPlayer {
	if i >= 0 && i < len(s.players) {
		return s.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (s *FiveCardStud) GetPlayerCnt() int { return len(s.players) }

// GetCommunityCard 共有カード取得 (カード不足時のみ)
func (s *FiveCardStud) GetCommunityCard() *Card { return s.communityCard }

// GetPot ポット取得
func (s *FiveCardStud) GetPot() int { return s.pot }

// GetSidePots サイドポット取得
func (s *FiveCardStud) GetSidePots() []SidePot { return s.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (s *FiveCardStud) GetDealerIdx() int { return s.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (s *FiveCardStud) GetCurrentTurn() int { return s.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *FiveCardStud) GetGameEndFlag() bool { return s.gameEndFlag }

// GetLastBet 最後のベット取得
func (s *FiveCardStud) GetLastBet() int { return s.lastBet }

// GetMinRaise 最小レイズ額取得
func (s *FiveCardStud) GetMinRaise() int { return s.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (s *FiveCardStud) GetRaiseCount() int { return s.raiseCount }

// GetRoundResults ラウンド結果取得
func (s *FiveCardStud) GetRoundResults() []FiveCardStudResult { return s.roundResults }

// GetCpuActions CPU行動記録取得
func (s *FiveCardStud) GetCpuActions() []FiveCardStudCpuAction { return s.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得
func (s *FiveCardStud) GetLastCpuError() error { return s.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (s *FiveCardStud) GetHumanProfile() *BettingHumanProfile { return s.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (s *FiveCardStud) ResetProfile() { s.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする
func (s *FiveCardStud) ExportProfile() interface{} {
	if s.humanProfile == nil {
		return nil
	}
	d := s.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (s *FiveCardStud) ImportProfile(data []byte) error {
	p, err := importBettingProfile(data)
	if err != nil || p == nil {
		return err
	}
	s.humanProfile = p
	return nil
}

// GetConfig 設定取得
func (s *FiveCardStud) GetConfig() FiveCardStudConfig { return s.config }

// SetConfig 設定変更
func (s *FiveCardStud) SetConfig(cfg FiveCardStudConfig) { s.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (s *FiveCardStud) IsHumanTurn() bool {
	if s.currentTurn >= 0 && s.currentTurn < len(s.players) {
		return s.players[s.currentTurn].GetIsHuman()
	}
	return false
}

// GetActedFlags actedフラグ取得
func (s *FiveCardStud) GetActedFlags() []bool {
	result := make([]bool, len(s.actedFlags))
	copy(result, s.actedFlags)
	return result
}

// GetHandCount ハンド数取得
func (s *FiveCardStud) GetHandCount() int { return s.handCount }

// GetBringInPlayerIdx ブリングインプレイヤーインデックス取得
func (s *FiveCardStud) GetBringInPlayerIdx() int { return s.bringInPlayerIdx }

// Resize プレイヤースライスを差し替え
func (s *FiveCardStud) Resize(players []*FiveCardStudPlayer) {
	s.players = players
	n := len(players)
	s.actedFlags = make([]bool, n)
	s.startingChips = make([]int, n)
	s.vpipTracked = make([]bool, n)
	s.pfrTracked = make([]bool, n)
	s.threeBetTracked = make([]bool, n)
	s.initTournamentState(n)
}

// --- JSON ---

// fiveCardStudJSON is the JSON wire format.
type fiveCardStudJSON struct {
	TrumpCards       *TrumpCards              `json:"tc"`
	Players          []*FiveCardStudPlayer    `json:"pl"`
	CommunityCard    *Card                    `json:"cc,omitempty"`
	Pot              int                      `json:"pt"`
	SidePots         []SidePot                `json:"sp"`
	DealerIdx        int                      `json:"di"`
	CurrentTurn      int                      `json:"ct"`
	Phase            int                      `json:"ph"`
	Config           FiveCardStudConfig       `json:"cf"`
	GameEndFlag      bool                     `json:"ge"`
	LastBet          int                      `json:"lb"`
	MinRaise         int                      `json:"mr"`
	RaiseCount       int                      `json:"rc"`
	ActedFlags       []bool                   `json:"af"`
	RoundResults     []FiveCardStudResult     `json:"rr"`
	CpuActions       []FiveCardStudCpuAction  `json:"ca"`
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
	BringInPlayerIdx int                      `json:"bi"`
}

const fiveCardStudMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (s *FiveCardStud) MarshalJSON() ([]byte, error) {
	j := fiveCardStudJSON{
		TrumpCards:       s.trumpCards,
		Players:          s.players,
		CommunityCard:    s.communityCard,
		Pot:              s.pot,
		SidePots:         s.sidePots,
		DealerIdx:        s.dealerIdx,
		CurrentTurn:      s.currentTurn,
		Phase:            s.phase,
		Config:           s.config,
		GameEndFlag:      s.gameEndFlag,
		LastBet:          s.lastBet,
		MinRaise:         s.minRaise,
		RaiseCount:       s.raiseCount,
		ActedFlags:       s.actedFlags,
		RoundResults:     s.roundResults,
		CpuActions:       s.cpuActions,
		StartingChips:    s.startingChips,
		VPIPTracked:      s.vpipTracked,
		PFRTracked:       s.pfrTracked,
		ThreeBetTracked:  s.threeBetTracked,
		HandCount:        s.handCount,
		RebuyCounts:      s.rebuyCounts,
		AddonUsed:        s.addonUsed,
		RebuyPhaseType:   s.rebuyPhaseType,
		ActionLog:        s.actionLog,
		LastHumanPlayMs:  s.lastHumanPlayMs,
		BringInPlayerIdx: s.bringInPlayerIdx,
	}
	if s.humanProfile != nil {
		d := s.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *FiveCardStud) UnmarshalJSON(data []byte) error {
	var j fiveCardStudJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > fiveCardStudMaxSliceLen || len(j.SidePots) > fiveCardStudMaxSliceLen ||
		len(j.ActedFlags) > fiveCardStudMaxSliceLen || len(j.RoundResults) > fiveCardStudMaxSliceLen ||
		len(j.CpuActions) > fiveCardStudMaxSliceLen || len(j.StartingChips) > fiveCardStudMaxSliceLen ||
		len(j.ActionLog) > fiveCardStudMaxSliceLen {
		return errFiveCardStudInvalid
	}
	if err := j.Config.Validate(); err != nil {
		return errFiveCardStudInvalid
	}
	if j.Phase < FiveCardStudPhaseInit || j.Phase > FiveCardStudPhaseRebuy {
		return errFiveCardStudInvalid
	}
	if j.Pot < 0 || j.LastBet < 0 {
		return errFiveCardStudInvalid
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.players = j.Players
	if s.players == nil {
		s.players = make([]*FiveCardStudPlayer, 0)
	}
	s.communityCard = j.CommunityCard
	s.pot = j.Pot
	s.sidePots = j.SidePots
	if s.sidePots == nil {
		s.sidePots = make([]SidePot, 0)
	}
	s.dealerIdx = j.DealerIdx
	s.currentTurn = j.CurrentTurn
	if len(s.players) > 0 && (s.currentTurn < 0 || s.currentTurn >= len(s.players)) {
		s.currentTurn = 0
	}
	s.phase = j.Phase
	s.config = j.Config
	s.gameEndFlag = j.GameEndFlag
	s.lastBet = j.LastBet
	s.minRaise = j.MinRaise
	s.raiseCount = j.RaiseCount
	s.actedFlags = j.ActedFlags
	if s.actedFlags == nil {
		s.actedFlags = make([]bool, 0)
	}
	s.roundResults = j.RoundResults
	if s.roundResults == nil {
		s.roundResults = make([]FiveCardStudResult, 0)
	}
	s.cpuActions = j.CpuActions
	if s.cpuActions == nil {
		s.cpuActions = make([]FiveCardStudCpuAction, 0)
	}
	s.startingChips = j.StartingChips
	if s.startingChips == nil {
		s.startingChips = make([]int, 0)
	}
	s.vpipTracked = j.VPIPTracked
	if s.vpipTracked == nil {
		s.vpipTracked = make([]bool, 0)
	}
	s.pfrTracked = j.PFRTracked
	if s.pfrTracked == nil {
		s.pfrTracked = make([]bool, 0)
	}
	s.threeBetTracked = j.ThreeBetTracked
	if s.threeBetTracked == nil {
		s.threeBetTracked = make([]bool, 0)
	}
	s.handCount = j.HandCount
	s.rebuyCounts = j.RebuyCounts
	if s.rebuyCounts == nil {
		s.rebuyCounts = make([]int, 0)
	}
	s.addonUsed = j.AddonUsed
	if s.addonUsed == nil {
		s.addonUsed = make([]bool, 0)
	}
	s.rebuyPhaseType = j.RebuyPhaseType
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	s.lastHumanPlayMs = j.LastHumanPlayMs
	s.bringInPlayerIdx = j.BringInPlayerIdx
	if j.Profile != nil {
		s.humanProfile = &BettingHumanProfile{}
		s.humanProfile.Import(*j.Profile)
	}
	return nil
}
