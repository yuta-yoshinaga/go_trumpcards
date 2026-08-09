//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// フェーズ定数
const (
	SevenCardStudPhaseInit          = 0 // 初期状態
	SevenCardStudPhaseThirdStreet   = 1 // サードストリート (2 down + 1 up + betting)
	SevenCardStudPhaseFourthStreet  = 2 // フォースストリート (1 up + betting)
	SevenCardStudPhaseFifthStreet   = 3 // フィフスストリート (1 up + betting)
	SevenCardStudPhaseSixthStreet   = 4 // シックススストリート (1 up + betting)
	SevenCardStudPhaseSeventhStreet = 5 // セブンスストリート (1 down + betting)
	SevenCardStudPhaseShowdown      = 6 // ショーダウン
	SevenCardStudPhaseEnd           = 7 // ゲーム終了
	SevenCardStudPhaseRebuy         = 8 // リバイ/アドオン待ち
)

// アクション定数 (共通定数のエイリアス)
const (
	SevenCardStudActionFold  = bettingActionFold
	SevenCardStudActionCheck = bettingActionCheck
	SevenCardStudActionCall  = bettingActionCall
	SevenCardStudActionBet   = bettingActionBet
	SevenCardStudActionRaise = bettingActionRaise
	SevenCardStudActionAllIn = bettingActionAllIn
)

// SevenCardStudResult ショーダウン結果
type SevenCardStudResult struct {
	PlayerIdx int     // プレイヤーインデックス
	HandRank  int     // ハンドランク
	HandName  string  // ハンド名
	BestHand  []*Card // ベスト5枚
	Kickers   []int   // キッカーカード値
	WonAmount int     // 獲得チップ
	Mucked    bool    // マックしたかどうか
	// LowQualifies は 8-or-better のローが成立したか (Hi-Lo のみ)。
	LowQualifies bool
	// LowBestHand はローのベスト5枚 (Hi-Lo のみ)。
	LowBestHand []*Card
	// WonLow はローとして獲得したチップ (Hi-Lo のみ)。WonAmount はハイとローの合計。
	WonLow int
}

// SevenCardStudCpuAction CPU行動記録
type SevenCardStudCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// リバイフェーズ種別定数
const (
	SevenCardStudRebuyPhaseNone  = 0
	SevenCardStudRebuyPhaseRebuy = 1
	SevenCardStudRebuyPhaseAddon = 2
)

// SevenCardStud セブンカードスタッド
type SevenCardStud struct {
	trumpCards      *TrumpCards
	players         []*SevenCardStudPlayer
	communityCard   *Card // カード不足時の共有カード
	pot             int
	sidePots        []SidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	config          SevenCardStudConfig
	gameEndFlag     bool
	lastBet         int
	minRaise        int
	raiseCount      int
	actedFlags      []bool
	roundResults    []SevenCardStudResult
	cpuActions      []SevenCardStudCpuAction
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
	bringInPlayerIdx int  // ブリングインプレイヤーインデックス
	lowball          bool // ローボール (Razz) モード
	hiLo             bool // Hi-Lo (8 or Better) スプリットモード
}

// NewSevenCardStud コンストラクタ
func NewSevenCardStud(trumpCards *TrumpCards, players []*SevenCardStudPlayer, config SevenCardStudConfig) *SevenCardStud {
	n := len(players)
	s := &SevenCardStud{
		trumpCards:       trumpCards,
		players:          players,
		sidePots:         make([]SidePot, 0),
		actedFlags:       make([]bool, n),
		roundResults:     make([]SevenCardStudResult, 0),
		cpuActions:       make([]SevenCardStudCpuAction, 0),
		startingChips:    make([]int, n),
		vpipTracked:      make([]bool, n),
		pfrTracked:       make([]bool, n),
		threeBetTracked:  make([]bool, n),
		config:           config,
		phase:            SevenCardStudPhaseInit,
		bringInPlayerIdx: -1,
	}
	s.initTournamentState(n)
	return s
}

// NewRazz Razz (A-5 ローボール) コンストラクタ
func NewRazz(trumpCards *TrumpCards, players []*SevenCardStudPlayer, config SevenCardStudConfig) *SevenCardStud {
	s := NewSevenCardStud(trumpCards, players, config)
	s.lowball = true
	return s
}

// NewSevenCardStudHiLo は Seven Card Stud Hi-Lo (8 or Better) を生成する。
//
// ハイは通常のスタッドと同じ。ローは **8 以下 5 枚・ペア無し** が成立した人の
// あいだで争い、成立者が 1 人もいなければハイがポットを総取りする。既存の
// Razz エンジンではなく通常のハイ評価に乗せているのは、Hi-Lo の主役があくまで
// ハイで、ローは条件付きの半分だからである。
func NewSevenCardStudHiLo(trumpCards *TrumpCards, players []*SevenCardStudPlayer, config SevenCardStudConfig) *SevenCardStud {
	s := NewSevenCardStud(trumpCards, players, config)
	s.hiLo = true
	return s
}

// NewDefaultSevenCardStudHiLo returns Seven Card Stud Hi-Lo (8 or Better) with
// the default table size. Used as the single source of truth for CUI, Web, and
// Worker construction sites.
func NewDefaultSevenCardStudHiLo() *SevenCardStud {
	cfg := DefaultSevenCardStudConfig()
	return NewSevenCardStudHiLo(NewTrumpCards(0), NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
}

// NewDefaultSevenCardStud returns SevenCardStud with the default table size and
// DefaultSevenCardStudConfig. Used as the single source of truth for CUI, Web,
// and Worker construction sites.
func NewDefaultSevenCardStud() *SevenCardStud {
	cfg := DefaultSevenCardStudConfig()
	return NewSevenCardStud(NewTrumpCards(0), NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
}

// NewDefaultRazz returns Razz (A-5 lowball) with the default table size and
// DefaultRazzConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultRazz() *SevenCardStud {
	cfg := DefaultRazzConfig()
	return NewRazz(NewTrumpCards(0), NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
}

// GetIsLowball ローボールモードかどうか
func (s *SevenCardStud) GetIsLowball() bool { return s.lowball }

// GetIsHiLo は Hi-Lo (8 or Better) スプリットかどうかを返す。
func (s *SevenCardStud) GetIsHiLo() bool { return s.hiLo }

// Reset ゲーム初期化
func (s *SevenCardStud) Reset() error {
	s.phase = SevenCardStudPhaseInit
	s.pot = 0
	s.sidePots = make([]SidePot, 0)
	s.communityCard = nil
	s.gameEndFlag = false
	s.lastBet = 0
	s.minRaise = s.config.SmallBet
	s.raiseCount = 0
	s.actedFlags = make([]bool, len(s.players))
	s.roundResults = make([]SevenCardStudResult, 0)
	s.cpuActions = make([]SevenCardStudCpuAction, 0)
	s.rebuyPhaseType = SevenCardStudRebuyPhaseNone
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
			s.phase = SevenCardStudPhaseRebuy
			s.rebuyPhaseType = SevenCardStudRebuyPhaseRebuy
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
func (s *SevenCardStud) continueReset() error {
	// ハンド開始時のチップを記録
	s.startingChips = make([]int, len(s.players))
	for i, p := range s.players {
		s.startingChips[i] = p.GetChips()
	}

	// アンティ投入
	s.postAntes()

	// サードストリート配布: 2枚伏せ + 1枚表
	for round := 0; round < 3; round++ {
		for j := 0; j < len(s.players); j++ {
			idx := (s.dealerIdx + 1 + j) % len(s.players)
			card := s.trumpCards.DrawCard()
			if card == nil {
				break
			}
			if round < 2 {
				s.players[idx].AddHoleCard(card)
			} else {
				s.players[idx].AddDoorCard(card)
			}
		}
	}

	s.phase = SevenCardStudPhaseThirdStreet

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
func (s *SevenCardStud) postAntes() {
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
// 通常: 最も低いドアカードを持つプレイヤー
// Razz: 最も高いドアカードを持つプレイヤー
// 同値の場合はスートランキングで決定 (クラブ < ダイヤ < ハート < スペード)
func (s *SevenCardStud) determineBringIn() int {
	bestIdx := 0
	bestVal := -1
	bestSuit := -1
	if !s.lowball {
		bestVal = 999
		bestSuit = 999
	}

	for i, p := range s.players {
		if len(p.GetDoorCards()) == 0 {
			continue
		}
		door := p.GetDoorCards()[0]
		val := door.GetValue()
		if val == 1 && !s.lowball {
			val = 14 // Ace is high (normal stud only; in Razz ace stays low)
		}
		suit := SuitRank(door.GetDesign())

		if s.lowball {
			// Razz: 最も高いカードがブリングイン
			if val > bestVal || (val == bestVal && suit > bestSuit) {
				bestIdx = i
				bestVal = val
				bestSuit = suit
			}
		} else {
			// 通常: 最も低いカードがブリングイン
			if val < bestVal || (val == bestVal && suit < bestSuit) {
				bestIdx = i
				bestVal = val
				bestSuit = suit
			}
		}
	}
	return bestIdx
}

// postBringIn ブリングインを投入する
func (s *SevenCardStud) postBringIn() {
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
func (s *SevenCardStud) PlayerAction(action, amount, humanPlayMs int) error {
	if s.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if s.phase < SevenCardStudPhaseThirdStreet || s.phase > SevenCardStudPhaseSeventhStreet {
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
			s.humanProfile.RecordFoldToBet(action == SevenCardStudActionFold)
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
func (s *SevenCardStud) bettingPlayers() []BettingPlayer {
	return toBettingPlayers(s.players)
}

// executeAction 指定プレイヤーのアクション実行
func (s *SevenCardStud) executeAction(playerIdx, action, amount int) error {
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
func (s *SevenCardStud) currentBetSize() int {
	if s.phase >= SevenCardStudPhaseFifthStreet {
		return s.config.BigBet
	}
	return s.config.SmallBet
}

// advanceTurn 次のプレイヤーに進める
func (s *SevenCardStud) advanceTurn() {
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
func (s *SevenCardStud) isBettingRoundComplete() bool {
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
func (s *SevenCardStud) advancePhase() {
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
	case SevenCardStudPhaseThirdStreet:
		s.phase = SevenCardStudPhaseFourthStreet
		s.minRaise = s.config.SmallBet
		s.dealStreetCard(true) // 表向き
		s.appendLog(-1, "deal", "dealt fourth street", nil)
	case SevenCardStudPhaseFourthStreet:
		s.phase = SevenCardStudPhaseFifthStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(true)
		s.appendLog(-1, "deal", "dealt fifth street", nil)
	case SevenCardStudPhaseFifthStreet:
		s.phase = SevenCardStudPhaseSixthStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(true)
		s.appendLog(-1, "deal", "dealt sixth street", nil)
	case SevenCardStudPhaseSixthStreet:
		s.phase = SevenCardStudPhaseSeventhStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(false) // 伏せ札
		s.appendLog(-1, "deal", "dealt seventh street", nil)
	case SevenCardStudPhaseSeventhStreet:
		s.phase = SevenCardStudPhaseShowdown
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
		s.phase = SevenCardStudPhaseShowdown
		s.resolveShowdown()
		return
	}

	// 4th Street以降: 最も強い表向き手を持つプレイヤーから開始
	s.currentTurn = s.determineBettingLeader()
}

// dealStreetCard 各アクティブプレイヤーにカードを1枚配る
func (s *SevenCardStud) dealStreetCard(faceUp bool) {
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
func (s *SevenCardStud) dealRemainingStreets() {
	// 現在のフェーズから7th streetまでのカードを配る
	for phase := s.phase; phase <= SevenCardStudPhaseSeventhStreet; phase++ {
		switch phase {
		case SevenCardStudPhaseFourthStreet, SevenCardStudPhaseFifthStreet, SevenCardStudPhaseSixthStreet:
			s.dealStreetCard(true)
		case SevenCardStudPhaseSeventhStreet:
			s.dealStreetCard(false)
		}
	}
}

// determineBettingLeader ベッティングリーダーを返す
// 通常: 最も強い表向き手を持つアクティブプレイヤー
// Razz: 最も弱い (低い) 表向き手を持つアクティブプレイヤー
func (s *SevenCardStud) determineBettingLeader() int {
	bestIdx := -1
	for i, p := range s.players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if bestIdx == -1 {
			bestIdx = i
			continue
		}
		if s.lowball {
			if CompareVisibleHandsLow(p, s.players[bestIdx]) > 0 {
				bestIdx = i
			}
		} else {
			if CompareVisibleHands(p, s.players[bestIdx]) > 0 {
				bestIdx = i
			}
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
func (s *SevenCardStud) countActivePlayers() int {
	return countPlayers(s.players, func(p *SevenCardStudPlayer) bool { return !p.GetFolded() })
}

// bettingLimits ベッティングリミット設定
func (s *SevenCardStud) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(s.config.BettingLimit, s.pot, s.lastBet)
}

// --- HUDスタッツ ---

// trackPreFlopStats サードストリートのHUDスタッツを追跡
func (s *SevenCardStud) trackPreFlopStats(playerIdx, action int) {
	if s.phase != SevenCardStudPhaseThirdStreet {
		return
	}
	isVPIPAction := false
	isPFRAction := false
	switch action {
	case SevenCardStudActionCall:
		isVPIPAction = true
	case SevenCardStudActionBet, SevenCardStudActionRaise, SevenCardStudActionAllIn:
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
		if action == SevenCardStudActionRaise || action == SevenCardStudActionAllIn {
			s.players[playerIdx].IncrementThreeBet()
		}
		s.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats 4th Street以降のAFスタッツを追跡
func (s *SevenCardStud) trackPostFlopStats(playerIdx, action int) {
	if s.phase < SevenCardStudPhaseFourthStreet || s.phase > SevenCardStudPhaseSeventhStreet {
		return
	}
	switch action {
	case SevenCardStudActionBet, SevenCardStudActionRaise, SevenCardStudActionAllIn:
		s.players[playerIdx].IncrementPostFlopBetRaise()
	case SevenCardStudActionCall:
		s.players[playerIdx].IncrementPostFlopCall()
	}
}

// --- ショーダウン ---

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (s *SevenCardStud) resolveLastPlayer() {
	for i, p := range s.players {
		if !p.GetFolded() {
			p.AddChips(s.pot)
			s.roundResults = []SevenCardStudResult{{
				PlayerIdx: i,
				WonAmount: s.pot,
			}}
			s.pot = 0
			break
		}
	}
	s.phase = SevenCardStudPhaseEnd
	s.gameEndFlag = true
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (s *SevenCardStud) resolveShowdown() {
	// ハンド評価 (共有カードがある場合はそれも含める)
	for _, p := range s.players {
		if !p.GetFolded() {
			if s.communityCard != nil {
				p.AddHoleCard(s.communityCard)
			}
			if s.lowball {
				p.EvalBestHandRazz()
			} else {
				p.EvalBestHand()
			}
			if s.hiLo {
				// ハイとは独立にローを評価する。ハイのベスト5枚とローの
				// ベスト5枚は同じ7枚から別々に選んでよい。
				p.EvalBestLowHandEightOrBetter()
			}
			if s.communityCard != nil {
				// 一時的に追加した共有カードを除去
				p.holeCards = p.holeCards[:len(p.holeCards)-1]
			}
		}
	}

	bp := s.bettingPlayers()
	s.sidePots = CalculateSidePots(bp, s.pot, s.startingChips)
	var wonAmounts map[int]int
	var wonLow map[int]int
	switch {
	case s.hiLo:
		wonAmounts, wonLow = s.distributeStudHiLoPots(bp)
	case s.lowball:
		wonAmounts = DistributePotsWithWinnerFunc(bp, s.sidePots, FindPotWinnersRazz)
	default:
		wonAmounts = DistributePots(bp, s.sidePots)
	}

	s.roundResults = make([]SevenCardStudResult, 0)
	humanLost := false
	for i, p := range s.players {
		if p.GetFolded() {
			continue
		}
		handName := s.getHandName(p.GetHandRank())
		if s.lowball {
			handName = getRazzHandName(p.GetHandRank(), p.GetBestHand())
		}
		result := SevenCardStudResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  handName,
			BestHand:  p.GetBestHand(),
			Kickers:   ExtractKickers(p.GetBestHand(), p.GetHandRank()),
			WonAmount: wonAmounts[i],
		}
		if s.hiLo {
			result.LowQualifies = p.GetLowQualifies()
			result.LowBestHand = p.GetLowBestHand()
			result.WonLow = wonLow[i]
			// WonAmount はハイとローの合計にする。片方だけを表示すると
			// 「勝ったのにチップが合わない」画面になる。
			result.WonAmount += wonLow[i]
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

// distributeStudHiLoPots は各サイドポットをハイ/ロー 50:50 で分配する。
//
// **qualifying なローが 1 人もいなければハイが全額を取る** —— これが 8 or Better
// の肝で、ローを取りに行った人が空振りするとポットが丸ごとハイへ行く。
// 奇数チップはハイ側に寄せる (ポーカー慣例)。
//
// 分配そのものは Omaha Hi-Lo と同じ helper (distributeAmongWinners) を通す。
func (s *SevenCardStud) distributeStudHiLoPots(bp []BettingPlayer) (hi, lo map[int]int) {
	hi = make(map[int]int)
	lo = make(map[int]int)
	for _, sp := range s.sidePots {
		hiWinners := FindPotWinners(bp, sp.EligiblePlayers)
		if len(hiWinners) == 0 {
			continue
		}
		loWinners := s.findStudLowWinners(sp.EligiblePlayers)

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

// findStudLowWinners は対象プレイヤーのうち有効なロー (8 or Better) を持つ
// 最良の人を返す。1 人もいなければ nil。同点はスプリット。
func (s *SevenCardStud) findStudLowWinners(eligible []int) []int {
	var winners []int
	var bestCards []*Card
	for _, idx := range eligible {
		if idx < 0 || idx >= len(s.players) {
			continue
		}
		p := s.players[idx]
		if p.GetFolded() || !p.GetLowQualifies() {
			continue
		}
		cards := p.GetLowBestHand()
		if bestCards == nil {
			bestCards = cards
			winners = []int{idx}
			continue
		}
		switch cmp := compareRazzCards(cards, bestCards); {
		case cmp < 0:
			bestCards = cards
			winners = []int{idx}
		case cmp == 0:
			winners = append(winners, idx)
		}
	}
	return winners
}

// finalizeShowdown ショーダウンを完了しENDフェーズに遷移する
func (s *SevenCardStud) finalizeShowdown() {
	s.phase = SevenCardStudPhaseEnd
	s.gameEndFlag = true
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
}

// Muck 人間プレイヤーがハンドをマックする
func (s *SevenCardStud) Muck() error {
	if s.phase != SevenCardStudPhaseShowdown {
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
func (s *SevenCardStud) ShowHand() error {
	if s.phase != SevenCardStudPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	s.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (s *SevenCardStud) IsMuckAvailable() bool {
	if s.phase != SevenCardStudPhaseShowdown {
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
func (s *SevenCardStud) getHandName(rank int) string {
	return pokerHandName(rank)
}

// getRazzHandName Razz用ハンド名を返す (例: "8-Low", "Wheel", "One Pair")
func getRazzHandName(rank int, bestHand []*Card) string {
	if rank != PokerHandHighCard {
		if rank >= 0 && rank < len(PokerHandNames) {
			return PokerHandNames[rank]
		}
		return "Unknown"
	}
	if len(bestHand) == 0 {
		return "High Card"
	}
	// HighCard: 最高カード値で名前をつける (Ace=1)
	highVal := 0
	for _, c := range bestHand {
		v := c.GetValue()
		if v > highVal {
			highVal = v
		}
	}
	// A-2-3-4-5 = Wheel
	if highVal == 5 {
		vals := make(map[int]bool)
		for _, c := range bestHand {
			vals[c.GetValue()] = true
		}
		if vals[1] && vals[2] && vals[3] && vals[4] && vals[5] {
			return "Wheel"
		}
	}
	return fmt.Sprintf("%d-Low", highVal)
}

// --- 棋譜 ---

func (s *SevenCardStud) logAction(playerIdx, action, amount int) {
	switch action {
	case SevenCardStudActionFold:
		s.appendLog(playerIdx, "fold", "fold", nil)
	case SevenCardStudActionCheck:
		s.appendLog(playerIdx, "check", "check", nil)
	case SevenCardStudActionCall:
		s.appendLog(playerIdx, "call", fmt.Sprintf("call %d", s.players[playerIdx].GetCurrentBet()), nil)
	case SevenCardStudActionBet:
		s.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case SevenCardStudActionRaise:
		s.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case SevenCardStudActionAllIn:
		s.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", s.players[playerIdx].GetCurrentBet()), nil)
	}
}

// --- ゲッター ---

// GetPhase フェーズ取得
func (s *SevenCardStud) GetPhase() int { return s.phase }

// GetPlayers プレイヤー一覧取得
func (s *SevenCardStud) GetPlayers() []*SevenCardStudPlayer { return s.players }

// GetPlayer 指定プレイヤー取得
func (s *SevenCardStud) GetPlayer(i int) *SevenCardStudPlayer {
	if i >= 0 && i < len(s.players) {
		return s.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (s *SevenCardStud) GetPlayerCnt() int { return len(s.players) }

// GetCommunityCard 共有カード取得 (カード不足時のみ)
func (s *SevenCardStud) GetCommunityCard() *Card { return s.communityCard }

// GetPot ポット取得
func (s *SevenCardStud) GetPot() int { return s.pot }

// GetSidePots サイドポット取得
func (s *SevenCardStud) GetSidePots() []SidePot { return s.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (s *SevenCardStud) GetDealerIdx() int { return s.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (s *SevenCardStud) GetCurrentTurn() int { return s.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *SevenCardStud) GetGameEndFlag() bool { return s.gameEndFlag }

// GetLastBet 最後のベット取得
func (s *SevenCardStud) GetLastBet() int { return s.lastBet }

// GetMinRaise 最小レイズ額取得
func (s *SevenCardStud) GetMinRaise() int { return s.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (s *SevenCardStud) GetRaiseCount() int { return s.raiseCount }

// GetRoundResults ラウンド結果取得
func (s *SevenCardStud) GetRoundResults() []SevenCardStudResult { return s.roundResults }

// GetCpuActions CPU行動記録取得
func (s *SevenCardStud) GetCpuActions() []SevenCardStudCpuAction { return s.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得
func (s *SevenCardStud) GetLastCpuError() error { return s.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (s *SevenCardStud) GetHumanProfile() *BettingHumanProfile { return s.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (s *SevenCardStud) ResetProfile() { s.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする
func (s *SevenCardStud) ExportProfile() interface{} {
	if s.humanProfile == nil {
		return nil
	}
	d := s.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (s *SevenCardStud) ImportProfile(data []byte) error {
	p, err := importBettingProfile(data)
	if err != nil || p == nil {
		return err
	}
	s.humanProfile = p
	return nil
}

// GetConfig 設定取得
func (s *SevenCardStud) GetConfig() SevenCardStudConfig { return s.config }

// SetConfig 設定変更
func (s *SevenCardStud) SetConfig(cfg SevenCardStudConfig) { s.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (s *SevenCardStud) IsHumanTurn() bool {
	return isHumanTurn(s.players, s.currentTurn)
}

// GetActedFlags actedフラグ取得
func (s *SevenCardStud) GetActedFlags() []bool {
	return copyOf(s.actedFlags)
}

// GetHandCount ハンド数取得
func (s *SevenCardStud) GetHandCount() int { return s.handCount }

// GetBringInPlayerIdx ブリングインプレイヤーインデックス取得
func (s *SevenCardStud) GetBringInPlayerIdx() int { return s.bringInPlayerIdx }

// Resize プレイヤースライスを差し替え
func (s *SevenCardStud) Resize(players []*SevenCardStudPlayer) {
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

// sevenCardStudJSON is the JSON wire format.
type sevenCardStudJSON struct {
	TrumpCards       *TrumpCards              `json:"tc"`
	Players          []*SevenCardStudPlayer   `json:"pl"`
	CommunityCard    *Card                    `json:"cc,omitempty"`
	Pot              int                      `json:"pt"`
	SidePots         []SidePot                `json:"sp"`
	DealerIdx        int                      `json:"di"`
	CurrentTurn      int                      `json:"ct"`
	Phase            int                      `json:"ph"`
	Config           SevenCardStudConfig      `json:"cf"`
	GameEndFlag      bool                     `json:"ge"`
	LastBet          int                      `json:"lb"`
	MinRaise         int                      `json:"mr"`
	RaiseCount       int                      `json:"rc"`
	ActedFlags       []bool                   `json:"af"`
	RoundResults     []SevenCardStudResult    `json:"rr"`
	CpuActions       []SevenCardStudCpuAction `json:"ca"`
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
	Lowball          bool                     `json:"lw,omitempty"`
	HiLo             bool                     `json:"hl,omitempty"`
}

const sevenCardStudMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (s *SevenCardStud) MarshalJSON() ([]byte, error) {
	j := sevenCardStudJSON{
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
		Lowball:          s.lowball,
		HiLo:             s.hiLo,
	}
	if s.humanProfile != nil {
		d := s.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SevenCardStud) UnmarshalJSON(data []byte) error {
	var j sevenCardStudJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > sevenCardStudMaxSliceLen || len(j.SidePots) > sevenCardStudMaxSliceLen ||
		len(j.ActedFlags) > sevenCardStudMaxSliceLen || len(j.RoundResults) > sevenCardStudMaxSliceLen ||
		len(j.CpuActions) > sevenCardStudMaxSliceLen || len(j.StartingChips) > sevenCardStudMaxSliceLen ||
		len(j.ActionLog) > sevenCardStudMaxSliceLen {
		return fmt.Errorf("sevencardstud: input array exceeds maximum allowed size")
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.players = j.Players
	if s.players == nil {
		s.players = make([]*SevenCardStudPlayer, 0)
	}
	s.communityCard = j.CommunityCard
	s.pot = j.Pot
	s.sidePots = j.SidePots
	if s.sidePots == nil {
		s.sidePots = make([]SidePot, 0)
	}
	s.dealerIdx = j.DealerIdx
	s.currentTurn = j.CurrentTurn
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
		s.roundResults = make([]SevenCardStudResult, 0)
	}
	s.cpuActions = j.CpuActions
	if s.cpuActions == nil {
		s.cpuActions = make([]SevenCardStudCpuAction, 0)
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
	s.lowball = j.Lowball
	s.hiLo = j.HiLo
	if j.Profile != nil {
		s.humanProfile = &BettingHumanProfile{}
		s.humanProfile.Import(*j.Profile)
	}
	return nil
}
