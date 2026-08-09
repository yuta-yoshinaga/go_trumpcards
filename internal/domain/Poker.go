//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ポーカーゲームのフェーズ定数
const (
	PokerPhaseInit      = 0 // 初期状態
	PokerPhaseDeal      = 1 // カード配布後 (第1ベッティングラウンド)
	PokerPhaseExchange  = 2 // カード交換フェーズ
	PokerPhaseSecondBet = 3 // 第2ベッティングラウンド
	PokerPhaseEnd       = 4 // ゲーム終了
)

// アクション定数 (共通定数のエイリアス)
const (
	PokerActionFold  = bettingActionFold  // フォールド
	PokerActionCheck = bettingActionCheck // チェック
	PokerActionCall  = bettingActionCall  // コール
	PokerActionBet   = bettingActionBet   // ベット
	PokerActionRaise = bettingActionRaise // レイズ
	PokerActionAllIn = bettingActionAllIn // オールイン
)

// pokerDefaultMaxRaises Fixed/PotLimit時のデフォルト最大レイズ回数
const pokerDefaultMaxRaises = bettingMaxRaisesPerRound

// PokerResult ショーダウン結果
type PokerResult struct {
	PlayerIdx int    // プレイヤーインデックス
	HandRank  int    // ハンドランク
	HandName  string // ハンド名
	Kickers   []int  // キッカーカード値
	WonAmount int    // 獲得チップ
}

// PokerCpuAction CPU行動記録
type PokerCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// PokerCpuExchange CPUカード交換記録
type PokerCpuExchange struct {
	PlayerIdx     int // プレイヤーインデックス
	ExchangeCount int // 交換枚数
}

// straightDrawCardInfo ストレートドロー探索用のカード情報
type straightDrawCardInfo struct {
	idx   int
	value int
}

// findOpenEndedDraw sorted cardsからskip位置を除外し、残り4枚が連続かつ条件を満たすか判定する
func findOpenEndedDraw(cards []straightDrawCardInfo, check func(remaining []int) bool) int {
	for skip := 0; skip < len(cards); skip++ {
		remaining := make([]int, 0, 4)
		for j, c := range cards {
			if j != skip {
				remaining = append(remaining, c.value)
			}
		}
		if len(remaining) != 4 {
			continue
		}
		isConsecutive := true
		for k := 1; k < len(remaining); k++ {
			if remaining[k] != remaining[k-1]+1 {
				isConsecutive = false
				break
			}
		}
		if isConsecutive && check(remaining) {
			return cards[skip].idx
		}
	}
	return -1
}

// pokerRoundState ラウンドごとにリセットされる状態
type pokerRoundState struct {
	phase         int
	pot           int
	currentTurn   int
	lastBet       int
	minRaise      int
	raiseCount    int
	actedFlags    []bool
	sidePots      []SidePot
	startingChips []int
	roundResults  []PokerResult
	cpuActions    []PokerCpuAction
	cpuExchanges  []PokerCpuExchange
	actionLogBase
	gameEndFlag     bool
	lastCpuError    error // CPU行動エラーの最後のフォールバック記録 (テスト検出用)
	lastHumanPlayMs int
}

// Poker ポーカークラス (5枚ドローポーカー・マルチプレイヤー)
type Poker struct {
	trumpCards   *TrumpCards
	players      []*PokerPlayer
	config       PokerConfig
	dealerIdx    int
	humanProfile *BettingHumanProfile
	round        pokerRoundState
}

// NewPoker コンストラクタ
func NewPoker(trumpCards *TrumpCards, players []*PokerPlayer, config PokerConfig) *Poker {
	return &Poker{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: pokerRoundState{
			phase:         PokerPhaseInit,
			sidePots:      make([]SidePot, 0),
			actedFlags:    make([]bool, len(players)),
			roundResults:  make([]PokerResult, 0),
			cpuActions:    make([]PokerCpuAction, 0),
			cpuExchanges:  make([]PokerCpuExchange, 0),
			startingChips: make([]int, len(players)),
		},
	}
}

// NewDefaultPoker returns Poker with the standard 4-player setup (1 human balanced,
// 3 CPU with mixed styles) and DefaultPokerConfig. Used as the single source of truth
// for CUI, Web, and Worker construction sites.
func NewDefaultPoker() *Poker {
	config := DefaultPokerConfig()
	players := []*PokerPlayer{
		NewPokerPlayer(true, PokerStyleBalanced),
		NewPokerPlayer(false, PokerStyleConservative),
		NewPokerPlayer(false, PokerStyleAggressive),
		NewPokerPlayer(false, PokerStyleBluffer),
	}
	return NewPoker(NewTrumpCards(config.JokerCount), players, config)
}

// Reset ゲーム初期化
func (p *Poker) Reset() error {
	p.round = pokerRoundState{
		phase:         PokerPhaseInit,
		minRaise:      p.config.MinBet,
		sidePots:      make([]SidePot, 0),
		actedFlags:    make([]bool, len(p.players)),
		roundResults:  make([]PokerResult, 0),
		cpuActions:    make([]PokerCpuAction, 0),
		cpuExchanges:  make([]PokerCpuExchange, 0),
		startingChips: make([]int, len(p.players)),
	}

	// メタAI: プロファイル初期化
	if p.config.CpuMetaAI {
		if p.humanProfile != nil {
			p.humanProfile.GamesPlayed++
		} else {
			p.humanProfile = &BettingHumanProfile{}
		}
	}

	// Lowballモードではジョーカーを強制的に0にする
	if p.config.IsLowball {
		p.config.JokerCount = 0
	}

	// デッキをジョーカー枚数に合わせて再生成しシャッフル
	p.trumpCards = NewTrumpCards(p.config.JokerCount)
	p.trumpCards.Shuffle()

	// アクティブ席数 = human(1) + CpuCount (上限: players配列長)
	activeSeatCount := p.config.CpuCount + 1
	if activeSeatCount > len(p.players) {
		activeSeatCount = len(p.players)
	}
	if activeSeatCount < 1 {
		activeSeatCount = 1
	}

	for i, pl := range p.players {
		pl.Reset()
		pl.SetFolded(false)
		pl.SetAllIn(false)
		pl.SetCurrentBet(0)
		pl.SetExchangeCount(0)
		pl.handRank = 0
		if pl.GetChips() <= 0 {
			pl.SetChips(p.config.InitChips)
		}
		// 席数を超えるプレイヤーは即フォールド扱い
		if i >= activeSeatCount {
			pl.SetFolded(true)
		}
	}

	// ハンド開始時のチップを記録 (サイドポット計算用)
	for i, pl := range p.players {
		p.round.startingChips[i] = pl.GetChips()
	}

	// アンティ徴収
	p.collectAntes()

	// カード配布 (アクティブプレイヤーのみ、ディーラーの左から5枚ずつ)
	for c := 0; c < 5; c++ {
		for j := 0; j < len(p.players); j++ {
			idx := (p.dealerIdx + 1 + j) % len(p.players)
			if p.players[idx].GetFolded() {
				continue
			}
			card := p.trumpCards.DrawCard()
			if card != nil {
				p.players[idx].AddCard(card)
			}
		}
	}

	p.round.phase = PokerPhaseDeal
	// ディーラーの左からアクティブプレイヤーを開始
	p.round.currentTurn = p.findNextActive(p.dealerIdx)

	// CPU第1ベットアクション実行
	p.runCpuActions()
	return nil
}

// collectAntes アンティ徴収
func (p *Poker) collectAntes() {
	for i, pl := range p.players {
		if pl.GetFolded() {
			continue
		}
		ante := p.config.Ante
		if pl.GetChips() < ante {
			ante = pl.GetChips()
		}
		pl.SubtractChips(ante)
		p.round.pot += ante
		p.appendLog(i, "ante", fmt.Sprintf("ante %d chips", ante), nil)
	}
}

// PlayerAction 人間プレイヤーのアクション実行
// humanPlayMs: 迷い時間(ms, 0=計測なし)
func (p *Poker) PlayerAction(action, amount, humanPlayMs int) error {
	if p.round.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if p.round.phase != PokerPhaseDeal && p.round.phase != PokerPhaseSecondBet {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !p.players[p.round.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// メタAI: 人間アクション記録
	p.round.lastHumanPlayMs = humanPlayMs
	if p.config.CpuMetaAI && p.humanProfile != nil {
		pl := p.players[p.round.currentTurn]
		pl.EvalHand()
		handRank := pl.GetHandRank()
		p.humanProfile.RecordAction(handRank, action)
		p.humanProfile.RecordHesitation(humanPlayMs)
		if p.round.lastBet > pl.GetCurrentBet() {
			p.humanProfile.RecordFoldToBet(action == PokerActionFold)
		}
	}

	err := p.executeAction(p.round.currentTurn, action, amount)
	if err != nil {
		return err
	}

	p.advanceTurn()
	p.runCpuActions()

	// ベッティングラウンド完了で交換フェーズに移行した場合、CPU交換を実行
	if p.round.phase == PokerPhaseExchange {
		p.advanceExchangePhase()
	}
	return nil
}

// PlayerExchange プレイヤーカード交換
func (p *Poker) PlayerExchange(indices []int) error {
	if p.round.phase != PokerPhaseExchange {
		return NewDomainError(ErrWrongPhase, "Exchange is not allowed now.")
	}
	if !p.players[p.round.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// 指定カードを交換
	for _, idx := range indices {
		newCard := p.trumpCards.DrawCard()
		if newCard != nil {
			p.players[p.round.currentTurn].ExchangeCard(idx, newCard)
		}
	}
	p.players[p.round.currentTurn].SetExchangeCount(len(indices))
	p.appendLog(p.round.currentTurn, "exchange", fmt.Sprintf("exchange %d card(s)", len(indices)), nil)
	p.round.actedFlags[p.round.currentTurn] = true

	// 残りのCPU交換を実行
	p.advanceTurn()
	p.advanceExchangePhase()

	return nil
}

// PlayerStand カード交換なし
func (p *Poker) PlayerStand() error {
	if p.round.phase != PokerPhaseExchange {
		return NewDomainError(ErrWrongPhase, "Stand is not allowed now.")
	}
	if !p.players[p.round.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	p.players[p.round.currentTurn].SetExchangeCount(0)
	p.appendLog(p.round.currentTurn, "exchange", "exchange 0 card(s)", nil)
	p.round.actedFlags[p.round.currentTurn] = true

	// 残りのCPU交換を実行
	p.advanceTurn()
	p.advanceExchangePhase()

	return nil
}

// bettingPlayers BettingPlayerスライスを生成
func (p *Poker) bettingPlayers() []BettingPlayer {
	return toBettingPlayers(p.players)
}

// executeAction 指定プレイヤーのアクション実行
func (p *Poker) executeAction(playerIdx, action, amount int) error {
	bp := p.bettingPlayers()
	// ActedFlags はスライス参照を共有: ExecuteBettingAction 内の変更が p.round.actedFlags に直接反映される
	state := &BettingState{
		Pot: p.round.pot, LastBet: p.round.lastBet, MinRaise: p.round.minRaise,
		RaiseCount: p.round.raiseCount, ActedFlags: p.round.actedFlags,
	}
	maxRaises, maxBetAmount := p.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, p.config.MinBet, maxRaises, maxBetAmount)
	p.round.pot = state.Pot
	p.round.lastBet = state.LastBet
	p.round.minRaise = state.MinRaise
	p.round.raiseCount = state.RaiseCount
	if err != nil {
		return err
	}

	// アクションログ記録
	p.logBettingAction(playerIdx, action, amount)

	// フォールドでアクティブプレイヤーが1人になったらチェック
	if p.countActivePlayers() == 1 {
		p.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (p *Poker) advanceTurn() {
	if p.round.gameEndFlag {
		return
	}

	// ベッティングラウンド終了チェック
	if p.round.phase == PokerPhaseDeal || p.round.phase == PokerPhaseSecondBet {
		if p.isBettingRoundComplete() {
			p.advancePhase()
			return
		}
	}

	// 交換フェーズでの終了チェック
	if p.round.phase == PokerPhaseExchange {
		if p.isExchangeComplete() {
			return
		}
	}

	// 次のアクティブプレイヤーを探す
	for i := 1; i <= len(p.players); i++ {
		next := (p.round.currentTurn + i) % len(p.players)
		if !p.players[next].GetFolded() && !p.players[next].GetAllIn() && !p.round.actedFlags[next] {
			p.round.currentTurn = next
			return
		}
	}
}

// isRoundComplete 全アクティブプレイヤーが行動済みかチェック
func (p *Poker) isRoundComplete() bool {
	for i, pl := range p.players {
		if pl.GetFolded() || pl.GetAllIn() {
			continue
		}
		if !p.round.actedFlags[i] {
			return false
		}
	}
	return true
}

// isBettingRoundComplete ベッティングラウンドが完了したかチェック
func (p *Poker) isBettingRoundComplete() bool {
	return p.isRoundComplete()
}

// isExchangeComplete 交換フェーズが完了したかチェック
func (p *Poker) isExchangeComplete() bool {
	return p.isRoundComplete()
}

// advancePhase 次のフェーズに進める
func (p *Poker) advancePhase() {
	switch p.round.phase {
	case PokerPhaseDeal:
		// ベッティングラウンド1完了 → 交換フェーズへ
		p.round.phase = PokerPhaseExchange
		// ラウンドベットリセット
		for _, pl := range p.players {
			pl.SetCurrentBet(0)
		}
		p.round.lastBet = 0
		p.round.minRaise = p.config.MinBet
		p.round.raiseCount = 0
		p.round.actedFlags = make([]bool, len(p.players))
		for i, pl := range p.players {
			if pl.GetFolded() || pl.GetAllIn() {
				p.round.actedFlags[i] = true
			}
		}
		// ディーラーの左から開始
		p.round.currentTurn = p.findNextActive(p.dealerIdx)

	case PokerPhaseSecondBet:
		// ベッティングラウンド2完了 → ショーダウン
		p.resolveShowdown()
	}
}

// startSecondBettingRound 第2ベッティングラウンド開始
func (p *Poker) startSecondBettingRound() {
	p.round.phase = PokerPhaseSecondBet
	// ラウンドベットリセット
	for _, pl := range p.players {
		pl.SetCurrentBet(0)
	}
	p.round.lastBet = 0
	p.round.minRaise = p.config.MinBet
	p.round.raiseCount = 0
	p.round.actedFlags = make([]bool, len(p.players))
	for i, pl := range p.players {
		if pl.GetFolded() || pl.GetAllIn() {
			p.round.actedFlags[i] = true
		}
	}

	// アクティブプレイヤーが0-1人ならショーダウンへ
	activeCnt := 0
	for _, pl := range p.players {
		if !pl.GetFolded() && !pl.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		p.resolveShowdown()
		return
	}

	// ディーラーの左から開始
	p.round.currentTurn = p.findNextActive(p.dealerIdx)

	// CPU第2ベットアクション実行
	p.runCpuActions()
}

// findNextActive 指定インデックスの次のアクティブプレイヤーを探す
func (p *Poker) findNextActive(fromIdx int) int {
	return findNextActive(p.players, fromIdx)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (p *Poker) countActivePlayers() int {
	cnt := 0
	for _, pl := range p.players {
		if !pl.GetFolded() {
			cnt++
		}
	}
	return cnt
}

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (p *Poker) resolveLastPlayer() {
	for i, pl := range p.players {
		if !pl.GetFolded() {
			pl.AddChips(p.round.pot)
			p.round.roundResults = []PokerResult{{
				PlayerIdx: i,
				WonAmount: p.round.pot,
			}}
			p.round.pot = 0
			break
		}
	}
	p.round.phase = PokerPhaseEnd
	p.round.gameEndFlag = true
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (p *Poker) resolveShowdown() {
	// ハンド評価
	for i, pl := range p.players {
		if !pl.GetFolded() {
			pl.EvalHand()
			cards := make([]*Card, pl.GetCardsSize())
			for j := 0; j < pl.GetCardsSize(); j++ {
				cards[j] = pl.GetCard(j)
			}
			p.appendLog(i, "showdown", fmt.Sprintf("showdown: %s", pl.GetHandName()), cards)
		}
	}

	// サイドポット計算・配分
	bp := p.bettingPlayers()
	p.round.sidePots = CalculateSidePots(bp, p.round.pot, p.round.startingChips)
	var wonAmounts map[int]int
	if p.config.IsLowball {
		wonAmounts = DistributePotsWithWinnerFunc(bp, p.round.sidePots, FindPotWinnersLowball)
	} else {
		wonAmounts = DistributePots(bp, p.round.sidePots)
	}

	// 結果を構築
	p.round.roundResults = make([]PokerResult, 0)
	for i, pl := range p.players {
		if pl.GetFolded() {
			continue
		}
		result := PokerResult{
			PlayerIdx: i,
			HandRank:  pl.GetHandRank(),
			HandName:  pl.GetHandName(),
			Kickers:   ExtractKickers(pl.GetComparisonCards(), pl.GetHandRank()),
			WonAmount: wonAmounts[i],
		}
		p.round.roundResults = append(p.round.roundResults, result)
	}

	p.round.phase = PokerPhaseEnd
	p.round.gameEndFlag = true
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
}

// runCpuActions CPUプレイヤーのアクションを実行
func (p *Poker) runCpuActions() {
	if p.round.gameEndFlag {
		return
	}
	for !p.round.gameEndFlag && (p.round.phase == PokerPhaseDeal || p.round.phase == PokerPhaseSecondBet) {
		if p.players[p.round.currentTurn].GetIsHuman() {
			return
		}
		if p.players[p.round.currentTurn].GetFolded() || p.players[p.round.currentTurn].GetAllIn() {
			p.advanceTurn()
			continue
		}
		action, amount := p.cpuDecide(p.round.currentTurn)
		p.round.cpuActions = append(p.round.cpuActions, PokerCpuAction{
			PlayerIdx: p.round.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := p.executeAction(p.round.currentTurn, action, amount)
		if err != nil {
			p.round.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", p.round.currentTurn, action, err)
			callAmt := p.round.lastBet - p.players[p.round.currentTurn].GetCurrentBet()
			if callAmt > 0 {
				_ = p.executeAction(p.round.currentTurn, PokerActionFold, 0)
			} else {
				_ = p.executeAction(p.round.currentTurn, PokerActionCheck, 0)
			}
		}
		if p.round.gameEndFlag {
			return
		}
		p.advanceTurn()
	}
}

// advanceExchangePhase runs remaining CPU exchanges and, if all players have
// exchanged, starts the second betting round.
func (p *Poker) advanceExchangePhase() {
	if p.round.gameEndFlag {
		return
	}
	p.runCpuExchanges()
	if p.isExchangeComplete() {
		p.startSecondBettingRound()
	}
}

// runCpuExchanges CPUプレイヤーのカード交換を実行
func (p *Poker) runCpuExchanges() {
	if p.round.gameEndFlag {
		return
	}
	for p.round.phase == PokerPhaseExchange {
		if p.isExchangeComplete() {
			return
		}
		if p.players[p.round.currentTurn].GetIsHuman() {
			return
		}
		if p.players[p.round.currentTurn].GetFolded() || p.players[p.round.currentTurn].GetAllIn() {
			p.round.actedFlags[p.round.currentTurn] = true
			p.advanceTurn()
			continue
		}

		// CPU交換AI
		var indices []int
		if p.config.IsLowball {
			indices = p.cpuDecideExchangeLowball(p.round.currentTurn)
		} else {
			indices = p.cpuDecideExchange(p.round.currentTurn)
		}
		for _, idx := range indices {
			newCard := p.trumpCards.DrawCard()
			if newCard != nil {
				p.players[p.round.currentTurn].ExchangeCard(idx, newCard)
			}
		}
		p.players[p.round.currentTurn].SetExchangeCount(len(indices))
		p.appendLog(p.round.currentTurn, "exchange", fmt.Sprintf("exchange %d card(s)", len(indices)), nil)
		p.round.cpuExchanges = append(p.round.cpuExchanges, PokerCpuExchange{
			PlayerIdx:     p.round.currentTurn,
			ExchangeCount: len(indices),
		})
		p.round.actedFlags[p.round.currentTurn] = true
		p.advanceTurn()
	}
}

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (p *Poker) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(p.config.BettingLimit, p.round.pot, p.round.lastBet)
}

// cpuDecide CPUプレイヤーの意思決定
func (p *Poker) cpuDecide(idx int) (int, int) {
	pl := p.players[idx]
	style := pl.GetPlayStyle()
	callAmount := p.round.lastBet - pl.GetCurrentBet()

	params, ok := pokerStyleParamsMap[style]
	if !ok {
		return p.cpuCallOrCheck(callAmount)
	}

	// メタAI: ブラフ率を調整
	if p.config.CpuMetaAI && p.humanProfile != nil {
		adjusted := p.humanProfile.AdjustedBluffChance(float64(params.bluffRate))
		params.bluffRate = int(math.Round(adjusted))
	}

	pl.EvalHand()
	handRank := pl.GetHandRank()

	// Lowballモードではハンドランクを反転 (弱いハンドほど高く評価)
	if p.config.IsLowball {
		handRank = PokerHandFiveOfAKind - handRank
	}

	// 交換枚数読み: 他プレイヤーの交換枚数が少ない場合に警戒
	exchangeWarning := p.calcExchangeWarning(idx, params.exchangeReadWeight)

	var action, amount int
	if p.round.phase == PokerPhaseDeal {
		action, amount = p.cpuDecideFirstBet(idx, params, callAmount, handRank)
	} else {
		action, amount = p.cpuDecideSecondBet(idx, params, callAmount, handRank, exchangeWarning)
	}

	// メタAI: 人間のベット/レイズに対してコール確率を調整
	if p.config.CpuMetaAI && p.humanProfile != nil && p.round.lastHumanPlayMs > 0 {
		if action == PokerActionFold && callAmount > 0 {
			bracket := bettingHandBracket(handRank)
			adjustedCall := p.humanProfile.AdjustedCallChance(0.0, bracket, p.round.lastHumanPlayMs)
			if adjustedCall > 0 && rand.Float64() < adjustedCall {
				action = PokerActionCall
				amount = 0
			}
		}
	}

	maxRaises, maxBetAmount := p.bettingLimits()

	// PotLimit: CPUベット額をポットサイズに制限
	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}

	// レイズ上限に達したら変更
	if maxRaises > 0 && p.round.raiseCount >= maxRaises {
		if action == PokerActionRaise || action == PokerActionBet {
			if callAmount > 0 {
				return PokerActionCall, 0
			}
			return PokerActionCheck, 0
		}
	}
	return action, amount
}

// calcExchangeWarning 他プレイヤーの交換枚数から警戒度を計算 (0-100)
func (p *Poker) calcExchangeWarning(idx, weight int) int {
	if p.round.phase != PokerPhaseSecondBet {
		return 0
	}
	minExchange := 5
	for i, pl := range p.players {
		if i == idx || pl.GetFolded() {
			continue
		}
		ec := pl.GetExchangeCount()
		if ec < minExchange {
			minExchange = ec
		}
	}
	// 交換枚数0 = 強い手の可能性高い → 警戒度Max
	// 交換枚数が少ないほど警戒度が高い
	if minExchange >= 3 {
		return 0
	}
	// warning = (3 - minExchange) * weight / 3
	return (3 - minExchange) * weight / 3
}

// cpuDecideFirstBet 第1ベッティングラウンドのCPU意思決定
func (p *Poker) cpuDecideFirstBet(idx int, params pokerCpuStyleParams, callAmount, handRank int) (int, int) {
	pl := p.players[idx]

	// フォールド評価
	if handRank <= params.firstFoldThreshold {
		if !params.aggressive {
			if callAmount > p.config.MinBet*params.firstCallMaxMult {
				return PokerActionFold, 0
			}
		} else {
			// アグレッシブスタイル: ブラフ率を先にチェック
			if rand.Intn(100) < params.bluffRate {
				betAmt := p.config.MinBet * params.firstBetMult
				return p.cpuRaiseOrBet(pl, callAmount, betAmt)
			}
			return p.cpuFoldOrCheck(callAmount)
		}
	}

	// ベット/レイズ
	if handRank >= params.firstBetThreshold || rand.Intn(100) < params.bluffRate {
		betAmt := p.config.MinBet * params.firstBetMult
		return p.cpuRaiseOrBet(pl, callAmount, betAmt)
	}

	return p.cpuCallOrCheck(callAmount)
}

// cpuDecideSecondBet 第2ベッティングラウンドのCPU意思決定
func (p *Poker) cpuDecideSecondBet(idx int, params pokerCpuStyleParams, callAmount, handRank, exchangeWarning int) (int, int) {
	pl := p.players[idx]

	// 交換読み警戒が高い場合、フォールド閾値を上げる
	adjustedFoldThreshold := params.secondFoldThreshold
	if exchangeWarning > 50 {
		adjustedFoldThreshold = params.secondFoldThreshold + 1
	}

	// フォールド評価
	if handRank <= adjustedFoldThreshold {
		if callAmount > p.config.MinBet*params.secondCallMaxMult {
			return PokerActionFold, 0
		}
		if callAmount > 0 {
			return p.cpuCallOrCheck(callAmount)
		}
	}

	// ベット/レイズ
	if handRank >= params.secondBetThreshold || rand.Intn(100) < params.bluffRate {
		betAmt := p.config.MinBet * params.secondBetMult
		return p.cpuRaiseOrBet(pl, callAmount, betAmt)
	}

	return p.cpuCallOrCheck(callAmount)
}

// cpuFoldOrCheck コール額がある場合はフォールド、なければチェック
func (p *Poker) cpuFoldOrCheck(callAmount int) (int, int) {
	return CpuFoldOrCheck(callAmount)
}

// cpuCallOrCheck コール額がある場合はコール、なければチェック
func (p *Poker) cpuCallOrCheck(callAmount int) (int, int) {
	return CpuCallOrCheck(callAmount)
}

// cpuRaiseOrBet レイズまたはベット (チップ不足時はオールイン)
func (p *Poker) cpuRaiseOrBet(pl *PokerPlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(pl.GetChips(), callAmount, raiseAmt)
}

// cpuDecideExchange CPUカード交換AI
func (p *Poker) cpuDecideExchange(idx int) []int {
	pl := p.players[idx]
	pl.EvalHand()
	rank := pl.GetHandRank()

	if rank >= PokerHandTwoPair {
		return []int{}
	}

	// フラッシュドロー判定
	if rank < PokerHandOnePair {
		discardIdx := p.findFlushDrawDiscard(idx)
		if discardIdx >= 0 {
			return []int{discardIdx}
		}
	}

	// ストレートドロー判定
	if rank < PokerHandOnePair {
		discardIdx := p.findStraightDrawDiscard(idx)
		if discardIdx >= 0 {
			return []int{discardIdx}
		}
	}

	// スタンドパットブラフ: 弱い手でも交換しないことで強い手を装う
	// ドローハンドより後に配置し、有望なドローを無駄にしない
	params := pokerStyleParamsMap[pl.GetPlayStyle()]
	if params.standPatBluffRate > 0 && rand.Intn(100) < params.standPatBluffRate {
		return []int{}
	}

	if rank == PokerHandOnePair {
		// ワンペアならペア以外の3枚を交換
		valueCounts := make(map[int][]int)
		for i := 0; i < pl.GetCardsSize(); i++ {
			v := pl.GetCard(i).GetValue()
			valueCounts[v] = append(valueCounts[v], i)
		}
		indices := []int{}
		for _, idxList := range valueCounts {
			if len(idxList) == 1 {
				indices = append(indices, idxList[0])
			}
		}
		return indices
	}

	// ハイカードなら最も低い3枚を交換
	// Note: ジョーカーがある場合は必ずOnePair以上になるためここに到達しない
	type cardIdx struct {
		idx   int
		value int
	}
	cards := make([]cardIdx, pl.GetCardsSize())
	for i := 0; i < pl.GetCardsSize(); i++ {
		v := pl.GetCard(i).GetValue()
		if v == 1 {
			v = 14
		}
		cards[i] = cardIdx{i, v}
	}
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].value < cards[j].value
	})
	result := []int{}
	for i := 0; i < 3 && i < len(cards); i++ {
		result = append(result, cards[i].idx)
	}
	return result
}

// cpuDecideExchangeLowball Lowball用CPUカード交換AI
// ペアがあれば片方を捨てる。8以上の高いカード(Ace=14)を捨てる。最大3枚交換。
func (p *Poker) cpuDecideExchangeLowball(idx int) []int {
	pl := p.players[idx]

	// カード値を収集 (Ace=14)
	type cardInfo struct {
		idx   int
		value int
	}
	cards := make([]cardInfo, pl.GetCardsSize())
	for i := 0; i < pl.GetCardsSize(); i++ {
		v := pl.GetCard(i).GetValue()
		if v == 1 {
			v = 14
		}
		cards[i] = cardInfo{i, v}
	}

	// ペアを見つけて余分なカードを交換候補にする (優先度高)
	valueCounts := make(map[int][]int)
	for _, c := range cards {
		valueCounts[c.value] = append(valueCounts[c.value], c.idx)
	}
	pairDiscards := []int{}
	isPairCard := make(map[int]bool)
	for _, idxList := range valueCounts {
		if len(idxList) >= 2 {
			for _, ci := range idxList {
				isPairCard[ci] = true
			}
			pairDiscards = append(pairDiscards, idxList[1:]...)
		}
	}

	// 8以上の高いカードでペアの一部でないものを候補に追加 (優先度低)
	highCardDiscards := []int{}
	for _, c := range cards {
		if c.value >= 8 && !isPairCard[c.idx] {
			highCardDiscards = append(highCardDiscards, c.idx)
		}
	}

	// ペア解消を優先し、残り枠を高いカードで埋める (最大3枚)
	// 5枚ハンドではペア交換候補は最大3枚 (4-of-a-kind)
	indices := pairDiscards
	if len(indices) < 3 {
		sort.Slice(highCardDiscards, func(i, j int) bool {
			return cards[highCardDiscards[i]].value > cards[highCardDiscards[j]].value
		})
		needed := 3 - len(indices)
		if len(highCardDiscards) > needed {
			highCardDiscards = highCardDiscards[:needed]
		}
		indices = append(indices, highCardDiscards...)
	}

	return indices
}

// findFlushDrawDiscard 4枚フラッシュドローの外れカード位置を返す
func (p *Poker) findFlushDrawDiscard(playerIdx int) int {
	pl := p.players[playerIdx]
	suitCounts := make(map[int]int)
	for i := 0; i < pl.GetCardsSize(); i++ {
		if pl.GetCard(i).GetDesign() != CardDesignJoker {
			suitCounts[pl.GetCard(i).GetDesign()]++
		}
	}
	for suit, count := range suitCounts {
		if count == 4 {
			for i := 0; i < pl.GetCardsSize(); i++ {
				if pl.GetCard(i).GetDesign() != suit && pl.GetCard(i).GetDesign() != CardDesignJoker {
					return i
				}
			}
		}
	}
	return -1
}

// findStraightDrawDiscard 4枚ストレートドローの外れカード位置を返す
func (p *Poker) findStraightDrawDiscard(playerIdx int) int {
	pl := p.players[playerIdx]
	cards := make([]straightDrawCardInfo, 0, pl.GetCardsSize())
	for i := 0; i < pl.GetCardsSize(); i++ {
		if pl.GetCard(i).GetDesign() == CardDesignJoker {
			continue // ジョーカーはストレートドロー計算から除外
		}
		v := pl.GetCard(i).GetValue()
		if v == 1 {
			v = 14
		}
		cards = append(cards, straightDrawCardInfo{i, v})
	}
	if len(cards) < 5 {
		return -1 // ジョーカーがある場合はスキップ
	}
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].value < cards[j].value
	})

	idx := findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] > 1 && r[3] < 14
	})
	if idx >= 0 {
		return idx
	}

	for i := range cards {
		if cards[i].value == 14 {
			cards[i].value = 1
		}
	}
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].value < cards[j].value
	})

	return findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] == 1 && r[3] <= 5
	})
}

// --- ゲッター ---

// GetPhase フェーズ取得
// pokerEquitySimulations はエクイティ計算のモンテカルロ試行回数。
const pokerEquitySimulations = 2000

// GetEquity はベッティングフェーズでの人間の勝率を返す。それ以外では nil。
//
// **Holdem 系は EquityDisplay でこれを出しているのに、5 カードドローには
// 仕組み自体が無く、2巡目ベットで call/raise/fold を判断する材料が
// 交換確率パネルしか無かった (#4678)。**
func (p *Poker) GetEquity() *HoldemEquityResult {
	if p.round.phase != PokerPhaseDeal && p.round.phase != PokerPhaseSecondBet {
		return nil
	}
	human := p.findHumanPlayer()
	if human == nil || human.GetFolded() {
		return nil
	}
	cards := make([]*Card, human.GetCardsSize())
	for i := range cards {
		cards[i] = human.GetCard(i)
	}
	active := 0
	for _, pl := range p.players {
		if !pl.GetIsHuman() && !pl.GetFolded() {
			active++
		}
	}
	result := CalcPokerEquity(cards, active, pokerEquitySimulations, nil)
	return &result
}

// GetPotOdds はコールに必要な額に対するポットオッズを返す (0-100)。
func (p *Poker) GetPotOdds() float64 {
	if p.round.phase != PokerPhaseDeal && p.round.phase != PokerPhaseSecondBet {
		return 0.0
	}
	human := p.findHumanPlayer()
	if human == nil {
		return 0.0
	}
	callAmount := p.round.lastBet - human.GetCurrentBet()
	if callAmount < 0 {
		callAmount = 0
	}
	return CalcPotOdds(p.round.pot, callAmount)
}

// findHumanPlayer は人間プレイヤーを返す (見つからなければ nil)。
func (p *Poker) findHumanPlayer() *PokerPlayer {
	for _, pl := range p.players {
		if pl.GetIsHuman() {
			return pl
		}
	}
	return nil
}

func (p *Poker) GetPhase() int { return p.round.phase }

// GetPlayers プレイヤー一覧取得
func (p *Poker) GetPlayers() []*PokerPlayer { return p.players }

// GetPot ポット取得
func (p *Poker) GetPot() int { return p.round.pot }

// GetSidePots サイドポット取得
func (p *Poker) GetSidePots() []SidePot { return p.round.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (p *Poker) GetDealerIdx() int { return p.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (p *Poker) GetCurrentTurn() int { return p.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (p *Poker) GetGameEndFlag() bool { return p.round.gameEndFlag }

// GetLastBet 最後のベット取得
func (p *Poker) GetLastBet() int { return p.round.lastBet }

// GetMinRaise 最小レイズ額取得
func (p *Poker) GetMinRaise() int { return p.round.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (p *Poker) GetRaiseCount() int { return p.round.raiseCount }

// GetAnte アンティ取得
func (p *Poker) GetAnte() int { return p.config.Ante }

// GetRoundResults ラウンド結果取得
func (p *Poker) GetRoundResults() []PokerResult { return p.round.roundResults }

// GetCpuActions CPU行動記録取得
func (p *Poker) GetCpuActions() []PokerCpuAction { return p.round.cpuActions }

// GetCpuExchanges CPU交換記録取得
func (p *Poker) GetCpuExchanges() []PokerCpuExchange { return p.round.cpuExchanges }

// GetConfig 設定取得
func (p *Poker) GetConfig() PokerConfig { return p.config }

// SetConfig 設定変更
func (p *Poker) SetConfig(cfg PokerConfig) { p.config = cfg }

// GetLastCpuError 最後のCPUアクションエラー取得 (テスト・デバッグ用)
func (p *Poker) GetLastCpuError() error { return p.round.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (p *Poker) GetHumanProfile() *BettingHumanProfile { return p.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (p *Poker) ResetProfile() { p.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする (プロファイルがない場合はnil)
func (p *Poker) ExportProfile() interface{} {
	if p.humanProfile == nil {
		return nil
	}
	d := p.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (p *Poker) ImportProfile(data []byte) error {
	prof, err := importBettingProfile(data)
	if err != nil || prof == nil {
		return err
	}
	p.humanProfile = prof
	return nil
}

// GetActionLog 棋譜を取得する
func (p *Poker) GetActionLog() []*ActionLogEntry { return p.round.actionLog }

// appendLog 棋譜にエントリを追加する
func (p *Poker) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.round.appendLog(playerIdx, actionType, detail, cards)
}

// pokerRoundStateJSON is the JSON wire format for pokerRoundState.
type pokerRoundStateJSON struct {
	Phase           int                `json:"ph"`
	Pot             int                `json:"pt"`
	CurrentTurn     int                `json:"ct"`
	LastBet         int                `json:"lb"`
	MinRaise        int                `json:"mr"`
	RaiseCount      int                `json:"rc"`
	ActedFlags      []bool             `json:"af"`
	SidePots        []SidePot          `json:"sp"`
	StartingChips   []int              `json:"sc"`
	RoundResults    []PokerResult      `json:"rr"`
	CpuActions      []PokerCpuAction   `json:"ca"`
	CpuExchanges    []PokerCpuExchange `json:"ce"`
	ActionLog       []*ActionLogEntry  `json:"al"`
	GameEndFlag     bool               `json:"ge"`
	LastHumanPlayMs int                `json:"hm"`
}

// pokerJSON is the JSON wire format for Poker.
type pokerJSON struct {
	TrumpCards *TrumpCards              `json:"tc"`
	Players    []*PokerPlayer           `json:"pl"`
	Config     PokerConfig              `json:"cf"`
	DealerIdx  int                      `json:"di"`
	Profile    *BettingHumanProfileData `json:"pf,omitempty"`
	Round      pokerRoundStateJSON      `json:"rd"`
}

// pokerMaxSliceLen caps slice sizes during deserialisation.
const pokerMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (p *Poker) MarshalJSON() ([]byte, error) {
	j := pokerJSON{
		TrumpCards: p.trumpCards,
		Players:    p.players,
		Config:     p.config,
		DealerIdx:  p.dealerIdx,
		Round: pokerRoundStateJSON{
			Phase:           p.round.phase,
			Pot:             p.round.pot,
			CurrentTurn:     p.round.currentTurn,
			LastBet:         p.round.lastBet,
			MinRaise:        p.round.minRaise,
			RaiseCount:      p.round.raiseCount,
			ActedFlags:      p.round.actedFlags,
			SidePots:        p.round.sidePots,
			StartingChips:   p.round.startingChips,
			RoundResults:    p.round.roundResults,
			CpuActions:      p.round.cpuActions,
			CpuExchanges:    p.round.cpuExchanges,
			ActionLog:       p.round.actionLog,
			GameEndFlag:     p.round.gameEndFlag,
			LastHumanPlayMs: p.round.lastHumanPlayMs,
		},
	}
	if p.humanProfile != nil {
		d := p.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Poker) UnmarshalJSON(data []byte) error {
	var j pokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > pokerMaxSliceLen || len(j.Round.ActedFlags) > pokerMaxSliceLen ||
		len(j.Round.SidePots) > pokerMaxSliceLen || len(j.Round.StartingChips) > pokerMaxSliceLen ||
		len(j.Round.RoundResults) > pokerMaxSliceLen || len(j.Round.CpuActions) > pokerMaxSliceLen ||
		len(j.Round.CpuExchanges) > pokerMaxSliceLen || len(j.Round.ActionLog) > pokerMaxSliceLen {
		return fmt.Errorf("poker: input array exceeds maximum allowed size")
	}
	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCards(0)
	}
	p.players = j.Players
	if p.players == nil {
		p.players = make([]*PokerPlayer, 0)
	}
	p.config = j.Config
	p.dealerIdx = j.DealerIdx
	if j.Profile != nil {
		p.humanProfile = &BettingHumanProfile{}
		p.humanProfile.Import(*j.Profile)
	}
	p.round = pokerRoundState{
		phase:           j.Round.Phase,
		pot:             j.Round.Pot,
		currentTurn:     j.Round.CurrentTurn,
		lastBet:         j.Round.LastBet,
		minRaise:        j.Round.MinRaise,
		raiseCount:      j.Round.RaiseCount,
		actedFlags:      j.Round.ActedFlags,
		sidePots:        j.Round.SidePots,
		startingChips:   j.Round.StartingChips,
		roundResults:    j.Round.RoundResults,
		cpuActions:      j.Round.CpuActions,
		cpuExchanges:    j.Round.CpuExchanges,
		actionLogBase:   actionLogBase{actionLog: j.Round.ActionLog},
		gameEndFlag:     j.Round.GameEndFlag,
		lastHumanPlayMs: j.Round.LastHumanPlayMs,
	}
	// Nil-guard slices
	if p.round.actedFlags == nil {
		p.round.actedFlags = make([]bool, 0)
	}
	if p.round.sidePots == nil {
		p.round.sidePots = make([]SidePot, 0)
	}
	if p.round.startingChips == nil {
		p.round.startingChips = make([]int, 0)
	}
	if p.round.roundResults == nil {
		p.round.roundResults = make([]PokerResult, 0)
	}
	if p.round.cpuActions == nil {
		p.round.cpuActions = make([]PokerCpuAction, 0)
	}
	if p.round.cpuExchanges == nil {
		p.round.cpuExchanges = make([]PokerCpuExchange, 0)
	}
	if p.round.actionLog == nil {
		p.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// logBettingAction ベッティングアクションをログに記録する
func (p *Poker) logBettingAction(playerIdx, action, amount int) {
	switch action {
	case PokerActionFold:
		p.appendLog(playerIdx, "fold", "fold", nil)
	case PokerActionCheck:
		p.appendLog(playerIdx, "check", "check", nil)
	case PokerActionCall:
		p.appendLog(playerIdx, "call", fmt.Sprintf("call %d", p.players[playerIdx].GetCurrentBet()), nil)
	case PokerActionBet:
		p.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", p.players[playerIdx].GetCurrentBet()), nil)
	case PokerActionRaise:
		p.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", p.players[playerIdx].GetCurrentBet()), nil)
	case PokerActionAllIn:
		p.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", p.players[playerIdx].GetCurrentBet()), nil)
	}
}
