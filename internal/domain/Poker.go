package domain

import (
	"fmt"
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

// CPU AI 閾値
const pokerMaxRaisesPerRound = bettingMaxRaisesPerRound

// PokerSidePot サイドポット (共通SidePot型のエイリアス)
type PokerSidePot = SidePot

// PokerResult ショーダウン結果
type PokerResult struct {
	PlayerIdx int    // プレイヤーインデックス
	HandRank  int    // ハンドランク
	HandName  string // ハンド名
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

// Poker ポーカークラス (5枚ドローポーカー・マルチプレイヤー)
type Poker struct {
	trumpCards    *TrumpCards
	players       []*PokerPlayer
	config        PokerConfig
	phase         int
	pot           int
	dealerIdx     int
	currentTurn   int
	lastBet       int
	minRaise      int
	raiseCount    int
	actedFlags    []bool
	sidePots      []PokerSidePot
	startingChips []int
	roundResults  []PokerResult
	cpuActions    []PokerCpuAction
	cpuExchanges  []PokerCpuExchange
	gameEndFlag   bool
	lastCpuError  error // CPU行動エラーの最後のフォールバック記録 (テスト検出用)
}

// NewPoker コンストラクタ
func NewPoker(trumpCards *TrumpCards, players []*PokerPlayer, config PokerConfig) *Poker {
	return &Poker{
		trumpCards:    trumpCards,
		players:       players,
		config:        config,
		phase:         PokerPhaseInit,
		sidePots:      make([]PokerSidePot, 0),
		actedFlags:    make([]bool, len(players)),
		roundResults:  make([]PokerResult, 0),
		cpuActions:    make([]PokerCpuAction, 0),
		cpuExchanges:  make([]PokerCpuExchange, 0),
		startingChips: make([]int, len(players)),
	}
}

// Reset ゲーム初期化
func (p *Poker) Reset() error {
	p.phase = PokerPhaseInit
	p.pot = 0
	p.sidePots = make([]PokerSidePot, 0)
	p.gameEndFlag = false
	p.lastBet = 0
	p.minRaise = p.config.MinBet
	p.raiseCount = 0
	p.actedFlags = make([]bool, len(p.players))
	p.roundResults = make([]PokerResult, 0)
	p.cpuActions = make([]PokerCpuAction, 0)
	p.cpuExchanges = make([]PokerCpuExchange, 0)

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
	p.startingChips = make([]int, len(p.players))
	for i, pl := range p.players {
		p.startingChips[i] = pl.GetChips()
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

	p.phase = PokerPhaseDeal
	// ディーラーの左からアクティブプレイヤーを開始
	p.currentTurn = p.findNextActive(p.dealerIdx)

	// CPU第1ベットアクション実行
	p.runCpuActions()
	return nil
}

// collectAntes アンティ徴収
func (p *Poker) collectAntes() {
	for _, pl := range p.players {
		if pl.GetFolded() {
			continue
		}
		ante := p.config.Ante
		if pl.GetChips() < ante {
			ante = pl.GetChips()
		}
		pl.SubtractChips(ante)
		p.pot += ante
	}
}

// PlayerAction 人間プレイヤーのアクション実行
func (p *Poker) PlayerAction(action, amount int) error {
	if p.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if p.phase != PokerPhaseDeal && p.phase != PokerPhaseSecondBet {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !p.players[p.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	err := p.executeAction(p.currentTurn, action, amount)
	if err != nil {
		return err
	}

	p.advanceTurn()
	p.runCpuActions()
	return nil
}

// PlayerExchange プレイヤーカード交換
func (p *Poker) PlayerExchange(indices []int) error {
	if p.phase != PokerPhaseExchange {
		return NewDomainError(ErrWrongPhase, "Exchange is not allowed now.")
	}
	if !p.players[p.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// 指定カードを交換
	for _, idx := range indices {
		newCard := p.trumpCards.DrawCard()
		if newCard != nil {
			p.players[p.currentTurn].ExchangeCard(idx, newCard)
		}
	}
	p.players[p.currentTurn].SetExchangeCount(len(indices))
	p.actedFlags[p.currentTurn] = true

	// 残りのCPU交換を実行
	p.advanceTurn()
	p.runCpuExchanges()

	// 交換フェーズ完了判定
	if p.isExchangeComplete() {
		p.startSecondBettingRound()
	}

	return nil
}

// PlayerStand カード交換なし
func (p *Poker) PlayerStand() error {
	if p.phase != PokerPhaseExchange {
		return NewDomainError(ErrWrongPhase, "Stand is not allowed now.")
	}
	if !p.players[p.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	p.players[p.currentTurn].SetExchangeCount(0)
	p.actedFlags[p.currentTurn] = true

	// 残りのCPU交換を実行
	p.advanceTurn()
	p.runCpuExchanges()

	// 交換フェーズ完了判定
	if p.isExchangeComplete() {
		p.startSecondBettingRound()
	}

	return nil
}

// bettingPlayers BettingPlayerスライスを生成
func (p *Poker) bettingPlayers() []BettingPlayer {
	bp := make([]BettingPlayer, len(p.players))
	for i, pl := range p.players {
		bp[i] = pl
	}
	return bp
}

// executeAction 指定プレイヤーのアクション実行
func (p *Poker) executeAction(playerIdx, action, amount int) error {
	bp := p.bettingPlayers()
	// ActedFlags はスライス参照を共有: ExecuteBettingAction 内の変更が p.actedFlags に直接反映される
	state := &BettingState{
		Pot: p.pot, LastBet: p.lastBet, MinRaise: p.minRaise,
		RaiseCount: p.raiseCount, ActedFlags: p.actedFlags,
	}
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, p.config.MinBet)
	p.pot = state.Pot
	p.lastBet = state.LastBet
	p.minRaise = state.MinRaise
	p.raiseCount = state.RaiseCount
	if err != nil {
		return err
	}

	// フォールドでアクティブプレイヤーが1人になったらチェック
	if p.countActivePlayers() == 1 {
		p.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (p *Poker) advanceTurn() {
	if p.gameEndFlag {
		return
	}

	// ベッティングラウンド終了チェック
	if p.phase == PokerPhaseDeal || p.phase == PokerPhaseSecondBet {
		if p.isBettingRoundComplete() {
			p.advancePhase()
			return
		}
	}

	// 交換フェーズでの終了チェック
	if p.phase == PokerPhaseExchange {
		if p.isExchangeComplete() {
			return
		}
	}

	// 次のアクティブプレイヤーを探す
	for i := 1; i <= len(p.players); i++ {
		next := (p.currentTurn + i) % len(p.players)
		if !p.players[next].GetFolded() && !p.players[next].GetAllIn() && !p.actedFlags[next] {
			p.currentTurn = next
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
		if !p.actedFlags[i] {
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
	switch p.phase {
	case PokerPhaseDeal:
		// ベッティングラウンド1完了 → 交換フェーズへ
		p.phase = PokerPhaseExchange
		// ラウンドベットリセット
		for _, pl := range p.players {
			pl.SetCurrentBet(0)
		}
		p.lastBet = 0
		p.minRaise = p.config.MinBet
		p.raiseCount = 0
		p.actedFlags = make([]bool, len(p.players))
		for i, pl := range p.players {
			if pl.GetFolded() || pl.GetAllIn() {
				p.actedFlags[i] = true
			}
		}
		// ディーラーの左から開始
		p.currentTurn = p.findNextActive(p.dealerIdx)

	case PokerPhaseSecondBet:
		// ベッティングラウンド2完了 → ショーダウン
		p.resolveShowdown()
	}
}

// startSecondBettingRound 第2ベッティングラウンド開始
func (p *Poker) startSecondBettingRound() {
	p.phase = PokerPhaseSecondBet
	// ラウンドベットリセット
	for _, pl := range p.players {
		pl.SetCurrentBet(0)
	}
	p.lastBet = 0
	p.minRaise = p.config.MinBet
	p.raiseCount = 0
	p.actedFlags = make([]bool, len(p.players))
	for i, pl := range p.players {
		if pl.GetFolded() || pl.GetAllIn() {
			p.actedFlags[i] = true
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
	p.currentTurn = p.findNextActive(p.dealerIdx)

	// CPU第2ベットアクション実行
	p.runCpuActions()
}

// findNextActive 指定インデックスの次のアクティブプレイヤーを探す
func (p *Poker) findNextActive(fromIdx int) int {
	for i := 1; i <= len(p.players); i++ {
		next := (fromIdx + i) % len(p.players)
		if !p.players[next].GetFolded() && !p.players[next].GetAllIn() {
			return next
		}
	}
	return (fromIdx + 1) % len(p.players)
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
			pl.AddChips(p.pot)
			p.roundResults = []PokerResult{{
				PlayerIdx: i,
				WonAmount: p.pot,
			}}
			p.pot = 0
			break
		}
	}
	p.phase = PokerPhaseEnd
	p.gameEndFlag = true
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (p *Poker) resolveShowdown() {
	// ハンド評価
	for _, pl := range p.players {
		if !pl.GetFolded() {
			pl.EvalHand()
		}
	}

	// サイドポット計算・配分
	bp := p.bettingPlayers()
	p.sidePots = CalculateSidePots(bp, p.pot, p.startingChips)
	wonAmounts := DistributePots(bp, p.sidePots)

	// 結果を構築
	p.roundResults = make([]PokerResult, 0)
	for i, pl := range p.players {
		if pl.GetFolded() {
			continue
		}
		result := PokerResult{
			PlayerIdx: i,
			HandRank:  pl.GetHandRank(),
			HandName:  pl.GetHandName(),
			WonAmount: wonAmounts[i],
		}
		p.roundResults = append(p.roundResults, result)
	}

	p.phase = PokerPhaseEnd
	p.gameEndFlag = true
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
}

// runCpuActions CPUプレイヤーのアクションを実行
func (p *Poker) runCpuActions() {
	if p.gameEndFlag {
		return
	}
	for !p.gameEndFlag && (p.phase == PokerPhaseDeal || p.phase == PokerPhaseSecondBet) {
		if p.players[p.currentTurn].GetIsHuman() {
			return
		}
		if p.players[p.currentTurn].GetFolded() || p.players[p.currentTurn].GetAllIn() {
			p.advanceTurn()
			continue
		}
		action, amount := p.cpuDecide(p.currentTurn)
		p.cpuActions = append(p.cpuActions, PokerCpuAction{
			PlayerIdx: p.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := p.executeAction(p.currentTurn, action, amount)
		if err != nil {
			p.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", p.currentTurn, action, err)
			callAmt := p.lastBet - p.players[p.currentTurn].GetCurrentBet()
			if callAmt > 0 {
				p.executeAction(p.currentTurn, PokerActionFold, 0)
			} else {
				p.executeAction(p.currentTurn, PokerActionCheck, 0)
			}
		}
		if p.gameEndFlag {
			return
		}
		p.advanceTurn()
	}
}

// runCpuExchanges CPUプレイヤーのカード交換を実行
func (p *Poker) runCpuExchanges() {
	if p.gameEndFlag {
		return
	}
	for p.phase == PokerPhaseExchange {
		if p.isExchangeComplete() {
			return
		}
		if p.players[p.currentTurn].GetIsHuman() {
			return
		}
		if p.players[p.currentTurn].GetFolded() || p.players[p.currentTurn].GetAllIn() {
			p.actedFlags[p.currentTurn] = true
			p.advanceTurn()
			continue
		}

		// CPU交換AI
		indices := p.cpuDecideExchange(p.currentTurn)
		for _, idx := range indices {
			newCard := p.trumpCards.DrawCard()
			if newCard != nil {
				p.players[p.currentTurn].ExchangeCard(idx, newCard)
			}
		}
		p.players[p.currentTurn].SetExchangeCount(len(indices))
		p.cpuExchanges = append(p.cpuExchanges, PokerCpuExchange{
			PlayerIdx:     p.currentTurn,
			ExchangeCount: len(indices),
		})
		p.actedFlags[p.currentTurn] = true
		p.advanceTurn()
	}
}

// cpuDecide CPUプレイヤーの意思決定
func (p *Poker) cpuDecide(idx int) (int, int) {
	pl := p.players[idx]
	style := pl.GetPlayStyle()
	callAmount := p.lastBet - pl.GetCurrentBet()

	params, ok := pokerStyleParamsMap[style]
	if !ok {
		return p.cpuCallOrCheck(callAmount)
	}

	pl.EvalHand()
	handRank := pl.GetHandRank()

	// 交換枚数読み: 他プレイヤーの交換枚数が少ない場合に警戒
	exchangeWarning := p.calcExchangeWarning(idx, params.exchangeReadWeight)

	var action, amount int
	if p.phase == PokerPhaseDeal {
		action, amount = p.cpuDecideFirstBet(idx, params, callAmount, handRank)
	} else {
		action, amount = p.cpuDecideSecondBet(idx, params, callAmount, handRank, exchangeWarning)
	}

	// レイズ上限に達したら変更
	if p.raiseCount >= pokerMaxRaisesPerRound {
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
	if p.phase != PokerPhaseSecondBet {
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
func (p *Poker) GetPhase() int { return p.phase }

// GetPlayers プレイヤー一覧取得
func (p *Poker) GetPlayers() []*PokerPlayer { return p.players }

// GetPot ポット取得
func (p *Poker) GetPot() int { return p.pot }

// GetSidePots サイドポット取得
func (p *Poker) GetSidePots() []PokerSidePot { return p.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (p *Poker) GetDealerIdx() int { return p.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (p *Poker) GetCurrentTurn() int { return p.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (p *Poker) GetGameEndFlag() bool { return p.gameEndFlag }

// GetLastBet 最後のベット取得
func (p *Poker) GetLastBet() int { return p.lastBet }

// GetMinRaise 最小レイズ額取得
func (p *Poker) GetMinRaise() int { return p.minRaise }

// GetAnte アンティ取得
func (p *Poker) GetAnte() int { return p.config.Ante }

// GetRoundResults ラウンド結果取得
func (p *Poker) GetRoundResults() []PokerResult { return p.roundResults }

// GetCpuActions CPU行動記録取得
func (p *Poker) GetCpuActions() []PokerCpuAction { return p.cpuActions }

// GetCpuExchanges CPU交換記録取得
func (p *Poker) GetCpuExchanges() []PokerCpuExchange { return p.cpuExchanges }

// GetConfig 設定取得
func (p *Poker) GetConfig() PokerConfig { return p.config }

// SetConfig 設定変更
func (p *Poker) SetConfig(cfg PokerConfig) { p.config = cfg }

// GetLastCpuError 最後のCPUアクションエラー取得 (テスト・デバッグ用)
func (p *Poker) GetLastCpuError() error { return p.lastCpuError }

// --- テスト用セッター ---

// SetPhase フェーズ設定（テスト用）
func (p *Poker) SetPhase(phase int) { p.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (p *Poker) SetCurrentTurn(turn int) { p.currentTurn = turn }

// SetPot ポット設定（テスト用）
func (p *Poker) SetPot(pot int) { p.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (p *Poker) SetDealerIdx(idx int) { p.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (p *Poker) SetGameEndFlag(flag bool) { p.gameEndFlag = flag }

// SetActedFlags actedフラグ設定（テスト用）
func (p *Poker) SetActedFlags(flags []bool) { p.actedFlags = flags }

// SetLastBet 最後のベット設定（テスト用）
func (p *Poker) SetLastBet(bet int) { p.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (p *Poker) SetMinRaise(raise int) { p.minRaise = raise }

// SetRaiseCount レイズ回数設定（テスト用）
func (p *Poker) SetRaiseCount(count int) { p.raiseCount = count }

// SetRoundResults ラウンド結果設定（テスト用）
func (p *Poker) SetRoundResults(results []PokerResult) { p.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (p *Poker) SetCpuActions(actions []PokerCpuAction) { p.cpuActions = actions }

// SetCpuExchanges CPU交換記録設定（テスト用）
func (p *Poker) SetCpuExchanges(exchanges []PokerCpuExchange) { p.cpuExchanges = exchanges }

// SetSidePots サイドポット設定（テスト用）
func (p *Poker) SetSidePots(pots []PokerSidePot) { p.sidePots = pots }

// SetStartingChips ハンド開始時チップ設定（テスト用）
func (p *Poker) SetStartingChips(chips []int) { p.startingChips = chips }

// GetStartingChips ハンド開始時チップ取得（テスト用）
func (p *Poker) GetStartingChips() []int { return p.startingChips }

// GetActedFlags actedフラグ取得（テスト用）
func (p *Poker) GetActedFlags() []bool {
	result := make([]bool, len(p.actedFlags))
	copy(result, p.actedFlags)
	return result
}
