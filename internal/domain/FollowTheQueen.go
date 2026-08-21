//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// フェーズ定数
const (
	FollowTheQueenPhaseInit          = 0 // 初期状態
	FollowTheQueenPhaseThirdStreet   = 1 // サードストリート (2 down + 1 up + betting)
	FollowTheQueenPhaseFourthStreet  = 2 // フォースストリート (1 up + betting)
	FollowTheQueenPhaseFifthStreet   = 3 // フィフスストリート (1 up + betting)
	FollowTheQueenPhaseSixthStreet   = 4 // シックススストリート (1 up + betting)
	FollowTheQueenPhaseSeventhStreet = 5 // セブンスストリート (1 down + betting)
	FollowTheQueenPhaseShowdown      = 6 // ショーダウン
	FollowTheQueenPhaseEnd           = 7 // ゲーム終了
	FollowTheQueenPhaseRebuy         = 8 // リバイ/アドオン待ち
)

// アクション定数 (共通定数のエイリアス)
const (
	FollowTheQueenActionFold  = bettingActionFold
	FollowTheQueenActionCheck = bettingActionCheck
	FollowTheQueenActionCall  = bettingActionCall
	FollowTheQueenActionBet   = bettingActionBet
	FollowTheQueenActionRaise = bettingActionRaise
	FollowTheQueenActionAllIn = bettingActionAllIn
)

// FollowTheQueenResult ショーダウン結果
type FollowTheQueenResult struct {
	PlayerIdx int     // プレイヤーインデックス
	HandRank  int     // ハンドランク
	HandName  string  // ハンド名
	BestHand  []*Card // ベスト5枚
	Kickers   []int   // キッカーカード値
	WonAmount int     // 獲得チップ
	Mucked    bool    // マックしたかどうか
}

// FollowTheQueenCpuAction CPU行動記録
type FollowTheQueenCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// リバイフェーズ種別定数
const (
	FollowTheQueenRebuyPhaseNone  = 0
	FollowTheQueenRebuyPhaseRebuy = 1
	FollowTheQueenRebuyPhaseAddon = 2
)

// FollowTheQueen フォロー・ザ・クイーン
type FollowTheQueen struct {
	trumpCards      *TrumpCards
	players         []*FollowTheQueenPlayer
	communityCard   *Card // カード不足時の共有カード
	pot             int
	sidePots        []SidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	config          FollowTheQueenConfig
	gameEndFlag     bool
	lastBet         int
	minRaise        int
	raiseCount      int
	actedFlags      []bool
	roundResults    []FollowTheQueenResult
	cpuActions      []FollowTheQueenCpuAction
	startingChips   []int
	vpipTracked     []bool
	pfrTracked      []bool
	threeBetTracked []bool
	tournamentBase  // handCount / rebuyCounts / addonUsed (issue #1463)
	lastCpuError    error
	// wildRank は「第2のワイルド」のランク。0 は未設定。表向きの Q の**次に
	// 配られた 1 枚**で決まり、次の表向き Q が出るたびに決め直される。
	wildRank int
	// queenPending は直前の表向き札が Q だった状態。次の 1 枚が wildRank を
	// 決める。**Q が最後の表向き札なら、この状態のまま終わって第2ワイルドは
	// 無いまま**になる ── それが規則。
	queenPending   bool
	rebuyPhaseType int
	actionLogBase
	humanProfile     *BettingHumanProfile
	lastHumanPlayMs  int
	bringInPlayerIdx int // ブリングインプレイヤーインデックス
}

// NewFollowTheQueen コンストラクタ
func NewFollowTheQueen(trumpCards *TrumpCards, players []*FollowTheQueenPlayer, config FollowTheQueenConfig) *FollowTheQueen {
	n := len(players)
	s := &FollowTheQueen{
		trumpCards:       trumpCards,
		players:          players,
		sidePots:         make([]SidePot, 0),
		actedFlags:       make([]bool, n),
		roundResults:     make([]FollowTheQueenResult, 0),
		cpuActions:       make([]FollowTheQueenCpuAction, 0),
		startingChips:    make([]int, n),
		vpipTracked:      make([]bool, n),
		pfrTracked:       make([]bool, n),
		threeBetTracked:  make([]bool, n),
		config:           config,
		phase:            FollowTheQueenPhaseInit,
		bringInPlayerIdx: -1,
	}
	s.initTournamentState(n)
	return s
}

// NewDefaultFollowTheQueen returns FollowTheQueen with the default table size and
// DefaultFollowTheQueenConfig. Used as the single source of truth for CUI, Web,
// and Worker construction sites.
func NewDefaultFollowTheQueen() *FollowTheQueen {
	cfg := DefaultFollowTheQueenConfig()
	return NewFollowTheQueen(NewTrumpCards(0), NewFollowTheQueenPlayersForTable(cfg.TableSize), cfg)
}

// Reset ゲーム初期化
func (s *FollowTheQueen) Reset() error {
	s.phase = FollowTheQueenPhaseInit
	s.pot = 0
	s.sidePots = make([]SidePot, 0)
	s.communityCard = nil
	s.gameEndFlag = false
	s.lastBet = 0
	s.minRaise = s.config.SmallBet
	s.raiseCount = 0
	s.actedFlags = make([]bool, len(s.players))
	s.roundResults = make([]FollowTheQueenResult, 0)
	s.cpuActions = make([]FollowTheQueenCpuAction, 0)
	s.rebuyPhaseType = FollowTheQueenRebuyPhaseNone
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
			s.phase = FollowTheQueenPhaseRebuy
			s.rebuyPhaseType = FollowTheQueenRebuyPhaseRebuy
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
func (s *FollowTheQueen) continueReset() error {
	// ハンド開始時のチップを記録
	s.startingChips = make([]int, len(s.players))
	for i, p := range s.players {
		s.startingChips[i] = p.GetChips()
	}

	// アンティ投入
	s.postAntes()

	// **前のハンドのワイルドを持ち越さない。**残ると、配る前から特定ランクが
	// 強い盤になる。配りの途中で Q が出れば、その場で改めて設定される。
	s.wildRank = 0
	s.queenPending = false
	s.publishWildRank()

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
				s.noteUpCard(card)
			}
		}
	}

	s.phase = FollowTheQueenPhaseThirdStreet

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
func (s *FollowTheQueen) postAntes() {
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

// determineBringIn ブリングインプレイヤーを決定する。
// 最も低いドアカードを持つプレイヤーが払う。同値の場合はスートランキングで
// 決定する (クラブ < ダイヤ < ハート < スペード)。
//
// **ワイルドはここでは効かない。** ブリングインはドアカードの「見た目の低さ」で
// 決まる強制ベットで、そのカードがワイルドかどうかは関係しない。ワイルドの
// クイーンを見せている席が最低ランクを理由に払う、という場面は起こりうる。
func (s *FollowTheQueen) determineBringIn() int {
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
			val = 14 // エースはハイ
		}
		suit := SuitRank(door.GetDesign())

		// 最も低いドアカードがブリングイン
		if val < bestVal || (val == bestVal && suit < bestSuit) {
			bestIdx = i
			bestVal = val
			bestSuit = suit
		}
	}
	return bestIdx
}

// postBringIn ブリングインを投入する
func (s *FollowTheQueen) postBringIn() {
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
func (s *FollowTheQueen) PlayerAction(action, amount, humanPlayMs int) error {
	if s.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if s.phase < FollowTheQueenPhaseThirdStreet || s.phase > FollowTheQueenPhaseSeventhStreet {
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
			s.humanProfile.RecordFoldToBet(action == FollowTheQueenActionFold)
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
func (s *FollowTheQueen) bettingPlayers() []BettingPlayer {
	return toBettingPlayers(s.players)
}

// executeAction 指定プレイヤーのアクション実行
func (s *FollowTheQueen) executeAction(playerIdx, action, amount int) error {
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
func (s *FollowTheQueen) currentBetSize() int {
	if s.phase >= FollowTheQueenPhaseFifthStreet {
		return s.config.BigBet
	}
	return s.config.SmallBet
}

// advanceTurn 次のプレイヤーに進める
func (s *FollowTheQueen) advanceTurn() {
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
func (s *FollowTheQueen) isBettingRoundComplete() bool {
	return bettingRoundComplete(s.players, s.actedFlags)
}

// advancePhase 次のフェーズに進める
func (s *FollowTheQueen) advancePhase() {
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
	case FollowTheQueenPhaseThirdStreet:
		s.phase = FollowTheQueenPhaseFourthStreet
		s.minRaise = s.config.SmallBet
		s.dealStreetCard(true) // 表向き
		s.appendLog(-1, "deal", "dealt fourth street", nil)
	case FollowTheQueenPhaseFourthStreet:
		s.phase = FollowTheQueenPhaseFifthStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(true)
		s.appendLog(-1, "deal", "dealt fifth street", nil)
	case FollowTheQueenPhaseFifthStreet:
		s.phase = FollowTheQueenPhaseSixthStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(true)
		s.appendLog(-1, "deal", "dealt sixth street", nil)
	case FollowTheQueenPhaseSixthStreet:
		s.phase = FollowTheQueenPhaseSeventhStreet
		s.minRaise = s.config.BigBet
		s.dealStreetCard(false) // 伏せ札
		s.appendLog(-1, "deal", "dealt seventh street", nil)
	case FollowTheQueenPhaseSeventhStreet:
		s.phase = FollowTheQueenPhaseShowdown
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
		s.phase = FollowTheQueenPhaseShowdown
		s.resolveShowdown()
		return
	}

	// 4th Street以降: 最も強い表向き手を持つプレイヤーから開始
	s.currentTurn = s.determineBettingLeader()
}

// dealStreetCard 各アクティブプレイヤーにカードを1枚配る
func (s *FollowTheQueen) dealStreetCard(faceUp bool) {
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
			s.noteUpCard(card)
		} else {
			s.players[idx].AddHoleCard(card)
		}
	}
}

// dealRemainingStreets 残りのストリートのカードを全て配る
func (s *FollowTheQueen) dealRemainingStreets() {
	// 現在のフェーズから7th streetまでのカードを配る
	for phase := s.phase; phase <= FollowTheQueenPhaseSeventhStreet; phase++ {
		switch phase {
		case FollowTheQueenPhaseFourthStreet, FollowTheQueenPhaseFifthStreet, FollowTheQueenPhaseSixthStreet:
			s.dealStreetCard(true)
		case FollowTheQueenPhaseSeventhStreet:
			s.dealStreetCard(false)
		}
	}
}

// determineBettingLeader ベッティングリーダー (最も強い表向き手を持つ
// アクティブプレイヤー) を返す。
func (s *FollowTheQueen) determineBettingLeader() int {
	bestIdx := -1
	for i, p := range s.players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if bestIdx == -1 {
			bestIdx = i
			continue
		}
		if followTheQueenCompareVisibleHands(p, s.players[bestIdx]) > 0 {
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
func (s *FollowTheQueen) countActivePlayers() int {
	return countPlayers(s.players, func(p *FollowTheQueenPlayer) bool { return !p.GetFolded() })
}

// bettingLimits ベッティングリミット設定
func (s *FollowTheQueen) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(s.config.BettingLimit, s.pot, s.lastBet)
}

// --- HUDスタッツ ---

// trackPreFlopStats サードストリートのHUDスタッツを追跡
func (s *FollowTheQueen) trackPreFlopStats(playerIdx, action int) {
	if s.phase != FollowTheQueenPhaseThirdStreet {
		return
	}
	isVPIPAction := false
	isPFRAction := false
	switch action {
	case FollowTheQueenActionCall:
		isVPIPAction = true
	case FollowTheQueenActionBet, FollowTheQueenActionRaise, FollowTheQueenActionAllIn:
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
		if action == FollowTheQueenActionRaise || action == FollowTheQueenActionAllIn {
			s.players[playerIdx].IncrementThreeBet()
		}
		s.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats 4th Street以降のAFスタッツを追跡
func (s *FollowTheQueen) trackPostFlopStats(playerIdx, action int) {
	if s.phase < FollowTheQueenPhaseFourthStreet || s.phase > FollowTheQueenPhaseSeventhStreet {
		return
	}
	switch action {
	case FollowTheQueenActionBet, FollowTheQueenActionRaise, FollowTheQueenActionAllIn:
		s.players[playerIdx].IncrementPostFlopBetRaise()
	case FollowTheQueenActionCall:
		s.players[playerIdx].IncrementPostFlopCall()
	}
}

// --- ショーダウン ---

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (s *FollowTheQueen) resolveLastPlayer() {
	for i, p := range s.players {
		if !p.GetFolded() {
			p.AddChips(s.pot)
			s.roundResults = []FollowTheQueenResult{{
				PlayerIdx: i,
				WonAmount: s.pot,
			}}
			s.pot = 0
			break
		}
	}
	s.phase = FollowTheQueenPhaseEnd
	s.gameEndFlag = true
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (s *FollowTheQueen) resolveShowdown() {
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

	s.roundResults = make([]FollowTheQueenResult, 0)
	humanLost := false
	for i, p := range s.players {
		if p.GetFolded() {
			continue
		}
		handName := s.getHandName(p.GetHandRank())
		result := FollowTheQueenResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  handName,
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
func (s *FollowTheQueen) finalizeShowdown() {
	// **配り終えたポットは 0 にする。** 理由は Holdem 側と同じ。
	s.pot = 0
	s.phase = FollowTheQueenPhaseEnd
	s.gameEndFlag = true
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
}

// Muck 人間プレイヤーがハンドをマックする
func (s *FollowTheQueen) Muck() error {
	if s.phase != FollowTheQueenPhaseShowdown {
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
func (s *FollowTheQueen) ShowHand() error {
	if s.phase != FollowTheQueenPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	s.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (s *FollowTheQueen) IsMuckAvailable() bool {
	if s.phase != FollowTheQueenPhaseShowdown {
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
func (s *FollowTheQueen) getHandName(rank int) string {
	return pokerHandName(rank)
}

// --- 棋譜 ---

func (s *FollowTheQueen) logAction(playerIdx, action, amount int) {
	switch action {
	case FollowTheQueenActionFold:
		s.appendLog(playerIdx, "fold", "fold", nil)
	case FollowTheQueenActionCheck:
		s.appendLog(playerIdx, "check", "check", nil)
	case FollowTheQueenActionCall:
		s.appendLog(playerIdx, "call", fmt.Sprintf("call %d", s.players[playerIdx].GetCurrentBet()), nil)
	case FollowTheQueenActionBet:
		s.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case FollowTheQueenActionRaise:
		s.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case FollowTheQueenActionAllIn:
		s.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", s.players[playerIdx].GetCurrentBet()), nil)
	}
}

// --- ゲッター ---

// FollowTheQueenQueenValue クイーンのランク。常時ワイルドで、表向きに出ると
// 「次の 1 枚」を第2のワイルドに指名する。
const FollowTheQueenQueenValue = CardValueMax - 1

// noteUpCard は表向きに配られた 1 枚を見て、ワイルドの状態を進める。
//
// **表向きの札だけがワイルドを動かす。**伏せ札はここを通らない。順序が要点で、
// 「Q の直後の 1 枚」を決めるために、まず保留中かどうかを見てから Q かどうかを
// 見る ── 逆にすると Q 自身が自分の後継になってしまう。
func (s *FollowTheQueen) noteUpCard(card *Card) {
	if card == nil {
		return
	}
	if card.GetValue() == FollowTheQueenQueenValue {
		// **新しい Q は前のワイルドを取り消す。**同時に 2 つのランクが
		// ワイルドになることはない。次の 1 枚が来るまでは第2ワイルド無し。
		s.wildRank = 0
		s.queenPending = true
		s.publishWildRank()
		return
	}
	if s.queenPending {
		s.wildRank = card.GetValue()
		s.queenPending = false
	}
	s.publishWildRank()
}

// publishWildRank は現在のワイルドを全プレイヤーへ配る。
//
// **評価はプレイヤー側で起きる。**ゲームだけが知っていると、ショーダウンも CPU の
// 判断も表示も、ワイルドを見ないまま進む ── 規則が飾りになる。
func (s *FollowTheQueen) publishWildRank() {
	for _, p := range s.players {
		p.SetWildRank(s.wildRank)
	}
}

// GetWildRank は第2のワイルドのランクを返す（0 なら未設定）。
func (s *FollowTheQueen) GetWildRank() int { return s.wildRank }

// IsWild はその札がワイルドかを返す。**Q は常に、加えて現在の wildRank が。**
// 表向きか伏せかは関係しない ── 動かすのは表向きの Q だけだが、ワイルドである
// ことは手札のどこにあっても同じ。
//
// `s.wildRank != 0` は**防御的なガードで、テストでは観測できない** —— 52 枚の
// デッキに値 0 のカードは無いので、外しても答えは変わらない。「テスト済み」では
// なく意図の表明としてここに置いてある。
func (s *FollowTheQueen) IsWild(card *Card) bool {
	if card == nil {
		return false
	}
	v := card.GetValue()
	return v == FollowTheQueenQueenValue || (s.wildRank != 0 && v == s.wildRank)
}

// GetPhase フェーズ取得
func (s *FollowTheQueen) GetPhase() int { return s.phase }

// GetPlayers プレイヤー一覧取得
func (s *FollowTheQueen) GetPlayers() []*FollowTheQueenPlayer { return s.players }

// GetPlayer 指定プレイヤー取得
func (s *FollowTheQueen) GetPlayer(i int) *FollowTheQueenPlayer {
	if i >= 0 && i < len(s.players) {
		return s.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (s *FollowTheQueen) GetPlayerCnt() int { return len(s.players) }

// GetCommunityCard 共有カード取得 (カード不足時のみ)
func (s *FollowTheQueen) GetCommunityCard() *Card { return s.communityCard }

// GetPot ポット取得
func (s *FollowTheQueen) GetPot() int { return s.pot }

// GetSidePots サイドポット取得
func (s *FollowTheQueen) GetSidePots() []SidePot { return s.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (s *FollowTheQueen) GetDealerIdx() int { return s.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (s *FollowTheQueen) GetCurrentTurn() int { return s.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *FollowTheQueen) GetGameEndFlag() bool { return s.gameEndFlag }

// GetLastBet 最後のベット取得
func (s *FollowTheQueen) GetLastBet() int { return s.lastBet }

// GetMinRaise 最小レイズ額取得
func (s *FollowTheQueen) GetMinRaise() int { return s.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (s *FollowTheQueen) GetRaiseCount() int { return s.raiseCount }

// GetRoundResults ラウンド結果取得
func (s *FollowTheQueen) GetRoundResults() []FollowTheQueenResult { return s.roundResults }

// GetCpuActions CPU行動記録取得
func (s *FollowTheQueen) GetCpuActions() []FollowTheQueenCpuAction { return s.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得
func (s *FollowTheQueen) GetLastCpuError() error { return s.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (s *FollowTheQueen) GetHumanProfile() *BettingHumanProfile { return s.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (s *FollowTheQueen) ResetProfile() { s.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする
func (s *FollowTheQueen) ExportProfile() interface{} {
	if s.humanProfile == nil {
		return nil
	}
	d := s.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (s *FollowTheQueen) ImportProfile(data []byte) error {
	p, err := importBettingProfile(data)
	if err != nil || p == nil {
		return err
	}
	s.humanProfile = p
	return nil
}

// GetConfig 設定取得
func (s *FollowTheQueen) GetConfig() FollowTheQueenConfig { return s.config }

// SetConfig 設定変更
func (s *FollowTheQueen) SetConfig(cfg FollowTheQueenConfig) { s.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (s *FollowTheQueen) IsHumanTurn() bool {
	return isHumanTurn(s.players, s.currentTurn)
}

// GetActedFlags actedフラグ取得
func (s *FollowTheQueen) GetActedFlags() []bool {
	return copyOf(s.actedFlags)
}

// GetHandCount ハンド数取得
func (s *FollowTheQueen) GetHandCount() int { return s.handCount }

// GetBringInPlayerIdx ブリングインプレイヤーインデックス取得
func (s *FollowTheQueen) GetBringInPlayerIdx() int { return s.bringInPlayerIdx }

// Resize プレイヤースライスを差し替え
func (s *FollowTheQueen) Resize(players []*FollowTheQueenPlayer) {
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

// followTheQueenJSON is the JSON wire format.
type followTheQueenJSON struct {
	TrumpCards       *TrumpCards               `json:"tc"`
	Players          []*FollowTheQueenPlayer   `json:"pl"`
	CommunityCard    *Card                     `json:"cc,omitempty"`
	Pot              int                       `json:"pt"`
	SidePots         []SidePot                 `json:"sp"`
	DealerIdx        int                       `json:"di"`
	CurrentTurn      int                       `json:"ct"`
	Phase            int                       `json:"ph"`
	Config           FollowTheQueenConfig      `json:"cf"`
	GameEndFlag      bool                      `json:"ge"`
	LastBet          int                       `json:"lb"`
	MinRaise         int                       `json:"mr"`
	RaiseCount       int                       `json:"rc"`
	ActedFlags       []bool                    `json:"af"`
	RoundResults     []FollowTheQueenResult    `json:"rr"`
	CpuActions       []FollowTheQueenCpuAction `json:"ca"`
	StartingChips    []int                     `json:"sc"`
	VPIPTracked      []bool                    `json:"vt"`
	PFRTracked       []bool                    `json:"ft"`
	ThreeBetTracked  []bool                    `json:"tt"`
	HandCount        int                       `json:"hc"`
	RebuyCounts      []int                     `json:"rb"`
	AddonUsed        []bool                    `json:"au"`
	RebuyPhaseType   int                       `json:"rp"`
	ActionLog        []*ActionLogEntry         `json:"al"`
	Profile          *BettingHumanProfileData  `json:"pf,omitempty"`
	LastHumanPlayMs  int                       `json:"hm"`
	BringInPlayerIdx int                       `json:"bi"`
}

const followTheQueenMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (s *FollowTheQueen) MarshalJSON() ([]byte, error) {
	j := followTheQueenJSON{
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
func (s *FollowTheQueen) UnmarshalJSON(data []byte) error {
	var j followTheQueenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > followTheQueenMaxSliceLen || len(j.SidePots) > followTheQueenMaxSliceLen ||
		len(j.ActedFlags) > followTheQueenMaxSliceLen || len(j.RoundResults) > followTheQueenMaxSliceLen ||
		len(j.CpuActions) > followTheQueenMaxSliceLen || len(j.StartingChips) > followTheQueenMaxSliceLen ||
		len(j.ActionLog) > followTheQueenMaxSliceLen {
		return fmt.Errorf("followthequeen: input array exceeds maximum allowed size")
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.players = j.Players
	if s.players == nil {
		s.players = make([]*FollowTheQueenPlayer, 0)
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
		s.roundResults = make([]FollowTheQueenResult, 0)
	}
	s.cpuActions = j.CpuActions
	if s.cpuActions == nil {
		s.cpuActions = make([]FollowTheQueenCpuAction, 0)
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
