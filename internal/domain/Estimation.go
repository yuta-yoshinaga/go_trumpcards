//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// EstimationPhase エスティメーションのゲームフェーズ
type EstimationPhase int

// Estimation のフェーズ定数
const (
	// EstimationPhaseTrump 親が切り札スートを決める
	EstimationPhaseTrump EstimationPhase = iota
	// EstimationPhaseBid 各自が獲得予定トリック数を宣言する
	EstimationPhaseBid
	// EstimationPhasePlay プレイ中
	EstimationPhasePlay
	// EstimationPhaseRoundEnd ラウンド終了
	EstimationPhaseRoundEnd
	// EstimationPhaseGameEnd ゲーム終了
	EstimationPhaseGameEnd
)

// EstimationPlayerCnt プレイヤー数（4 人固定・個人戦）
const EstimationPlayerCnt = 4

// EstimationHandSize 各プレイヤーの手札枚数
const EstimationHandSize = 13

// EstimationTricksPerRound 1 ラウンドのトリック数
const EstimationTricksPerRound = EstimationHandSize

// EstimationBaseScore 宣言の当否に乗る基礎点
const EstimationBaseScore = 10

// EstimationDashScore Dash Call（0 宣言）の得点。成功で +、失敗で -。
const EstimationDashScore = 23

// estimationMaxSliceLen caps slice sizes during deserialisation.
const estimationMaxSliceLen = 1000

// Estimation エスティメーション ゲームクラス。
//
// 中東・湾岸諸国で家庭的に遊ばれる予想制トリックテイキング。4 人個人戦、52 枚を
// 13 枚ずつ。**獲得トリック数をぴたりと言い当てることだけが得点になる**——
// 多すぎても少なすぎても同じだけ減点される、Oh Hell 系の中でも振れ幅の大きい形。
//
// このゲームに固有なのは宣言の *種類* で得点の桁が変わることで、実装したのは
// 2 つ:
//
//	Dash Call : 0 を宣言する。13 枚持って 1 つも取らないので ±23 と大きい
//	Risk      : そのラウンドで最も高く宣言した人。当否の得点が 2 倍になる
//
// **「宣言の合計が 13 になってはいけない」制約は既存の OhHell.go
// (`GetRestrictedBid`) が同じものを実装している。** issue はこれを新規と
// 書いているが、実際に新しいのは上の 2 つだけ。ここでも最後の宣言者
// （親の右隣）が合計 13 を避ける義務を負う。
//
// **得点表は出典が割れるので、写さずに一貫した形を選んである:**
//
//	通常 n    : 的中 +(10+n) / 外し -(10+n)
//	Dash (0) : 的中 +23      / 外し -23
//	Risk     : 上記の 2 倍
//
// 過不足の *量* ではなく当否だけで振れるので、高く宣言するほど賭けが大きい。
// Risk が最高宣言者に付くのはそのためで、切り札を選ぶ親の優位と釣り合う。
type Estimation struct {
	trumpCards *TrumpCards
	players    []*EstimationPlayer
	config     EstimationConfig

	phase       EstimationPhase
	roundNumber int
	trickNumber int
	trumpSuit   int

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int
	// bidPlayerIdx は宣言の手番。プレイの手番とは別に持つ。
	bidPlayerIdx int

	gameEndFlag bool
	winnerIdx   int

	actionLogBase
}

// NewEstimation コンストラクタ
func NewEstimation(trumpCards *TrumpCards, players []*EstimationPlayer, config EstimationConfig) *Estimation {
	return &Estimation{trumpCards: trumpCards, players: players, config: config, winnerIdx: -1}
}

// NewDefaultEstimation 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultEstimation() *Estimation {
	players := make([]*EstimationPlayer, 0, EstimationPlayerCnt)
	for i := range EstimationPlayerCnt {
		players = append(players, NewEstimationPlayer(i == 0))
	}
	return NewEstimation(NewTrumpCards(0), players, DefaultEstimationConfig())
}

// Reset ゲーム全体を初期化する
func (e *Estimation) Reset() {
	e.roundNumber = 1
	e.dealerIdx = 0
	e.gameEndFlag = false
	e.winnerIdx = -1
	e.actionLog = nil
	for _, p := range e.players {
		p.ResetGame()
	}
	e.dealRound()
}

// dealRound 13 枚ずつ配り、親が切り札を決めるところから始める
func (e *Estimation) dealRound() {
	e.phase = EstimationPhaseTrump
	e.trickNumber = 0
	e.currentTrick = nil
	e.trumpSuit = 0
	for _, p := range e.players {
		p.ResetRound()
	}

	e.trumpCards = NewTrumpCards(0)
	e.trumpCards.Shuffle()
	for range EstimationHandSize {
		for i := range EstimationPlayerCnt {
			idx := (e.dealerIdx + 1 + i) % EstimationPlayerCnt
			if c := e.trumpCards.DrawCard(); c != nil {
				e.players[idx].AddCard(c)
			}
		}
	}
	e.sortAllHands()
	// 切り札を決めるのは親。宣言もそこから始まる。
	e.currentPlayerIdx = e.dealerIdx
	e.bidPlayerIdx = e.dealerIdx
	e.leadPlayerIdx = (e.dealerIdx + 1) % EstimationPlayerCnt
	e.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d を開始", e.roundNumber), nil)
}

// sortAllHands 手札をスート・ランク順に並べ替える
func (e *Estimation) sortAllHands() {
	for _, p := range e.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return estimationRank(ci) < estimationRank(cj)
		})
	}
}

// estimationRank 札の強さ。A が最強、以下 K,Q,J,10..2。
func estimationRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// SelectTrump 人間の親が切り札スートを決める
func (e *Estimation) SelectTrump(suit int) error {
	if e.gameEndFlag {
		return errors.New("game has ended")
	}
	if e.phase != EstimationPhaseTrump {
		return errors.New("not the trump-selection phase")
	}
	if e.dealerIdx != 0 {
		return errors.New("only the dealer selects the trump suit")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", suit)
	}
	e.acceptTrump(suit)
	return nil
}

// CpuSelectTrump CPU の親が切り札スートを決める
func (e *Estimation) CpuSelectTrump() {
	if e.gameEndFlag || e.phase != EstimationPhaseTrump || e.dealerIdx == 0 {
		return
	}
	e.acceptTrump(e.longestSuit(e.dealerIdx))
}

// acceptTrump 切り札を確定させ、宣言フェーズに入る
func (e *Estimation) acceptTrump(suit int) {
	e.trumpSuit = suit
	e.phase = EstimationPhaseBid
	e.bidPlayerIdx = e.dealerIdx
	e.appendLog(e.dealerIdx, "trump", fmt.Sprintf("切り札を %d に決めた", suit), nil)
}

// longestSuit いちばん枚数の多いスート
func (e *Estimation) longestSuit(idx int) int {
	p := e.players[idx]
	counts := map[int]int{}
	for i := range p.GetCardsSize() {
		counts[p.GetCard(i).GetDesign()]++
	}
	best, bestN := CardDesignSpade, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestN {
			best, bestN = suit, counts[suit]
		}
	}
	return best
}

// PlayerBid 人間プレイヤーが宣言する
func (e *Estimation) PlayerBid(bid int) error {
	if e.gameEndFlag {
		return errors.New("game has ended")
	}
	if e.phase != EstimationPhaseBid {
		return errors.New("not the bidding phase")
	}
	if e.bidPlayerIdx != 0 {
		return errors.New("not your turn to bid")
	}
	if bid < 0 || bid > EstimationHandSize {
		return fmt.Errorf("invalid bid: %d", bid)
	}
	if r := e.GetRestrictedBid(); r >= 0 && bid == r {
		return fmt.Errorf("the last bidder cannot make the total %d", EstimationHandSize)
	}
	e.acceptBid(0, bid)
	return nil
}

// CpuBid 手番の CPU が宣言する
func (e *Estimation) CpuBid() {
	if e.gameEndFlag || e.phase != EstimationPhaseBid || e.bidPlayerIdx == 0 {
		return
	}
	idx := e.bidPlayerIdx
	bid := e.estimateTricks(idx)
	// **最後の宣言者は合計 13 を避ける義務がある。** 1 つずらして回避する。
	if r := e.GetRestrictedBid(); r >= 0 && bid == r {
		if bid > 0 {
			bid--
		} else {
			bid++
		}
	}
	e.acceptBid(idx, bid)
}

// acceptBid 宣言を記録し、次の席へ回す
func (e *Estimation) acceptBid(idx, bid int) {
	p := e.players[idx]
	p.SetBid(bid)
	if bid == 0 {
		p.SetCallType(EstimationCallDash)
		e.appendLog(idx, "dash", "Dash Call（0 宣言）", nil)
	} else {
		e.appendLog(idx, "bid", fmt.Sprintf("%d トリックを宣言", bid), nil)
	}

	// **残りを数えるのは SetBid の *後*。** `> 1` にすると 3 人目を記録した
	// 時点で締め切ってしまい、4 人目が一度も宣言しない——合計 13 禁止の制約が
	// 掛かる席がそもそも来なくなる。
	if e.bidsRemaining() > 0 {
		e.bidPlayerIdx = (idx + 1) % EstimationPlayerCnt
		return
	}
	e.closeBidding()
}

// bidsRemaining まだ宣言していない人数
func (e *Estimation) bidsRemaining() int {
	n := 0
	for _, p := range e.players {
		if p.GetBid() < 0 {
			n++
		}
	}
	return n
}

// closeBidding 宣言を締め切り、Risk を確定させてプレイに入る。
//
// **Risk は最高宣言者。** 同値なら先に宣言した人に付く（親から時計回り）。
// Dash Call は 0 なので、全員が Dash でない限り Risk にはならない。
func (e *Estimation) closeBidding() {
	riskIdx, best := -1, -1
	for i := range EstimationPlayerCnt {
		idx := (e.dealerIdx + i) % EstimationPlayerCnt
		if bid := e.players[idx].GetBid(); bid > best {
			riskIdx, best = idx, bid
		}
	}
	if riskIdx >= 0 && best > 0 {
		e.players[riskIdx].SetCallType(EstimationCallRisk)
		e.appendLog(riskIdx, "risk", fmt.Sprintf("Risk（最高宣言 %d）", best), nil)
	}
	e.phase = EstimationPhasePlay
	e.leadPlayerIdx = (e.dealerIdx + 1) % EstimationPlayerCnt
	e.currentPlayerIdx = e.leadPlayerIdx
}

// GetRestrictedBid 最後の宣言者が選べない宣言値を返す (-1 = 制限なし)。
//
// **合計が 13 ちょうどになると全員が宣言どおり取れてしまい、賭けが成り立たない。**
// OhHell.GetRestrictedBid と同じ制約で、こちらは最後に宣言する席に掛かる。
func (e *Estimation) GetRestrictedBid() int {
	if e.phase != EstimationPhaseBid || e.bidsRemaining() != 1 {
		return -1
	}
	total := 0
	for _, p := range e.players {
		if p.GetBid() >= 0 {
			total += p.GetBid()
		}
	}
	restricted := EstimationHandSize - total
	if restricted < 0 || restricted > EstimationHandSize {
		return -1
	}
	return restricted
}

// estimateTricks CPU の宣言。A/K/切り札の枚数から取れそうな数を見積もる。
func (e *Estimation) estimateTricks(idx int) int {
	p := e.players[idx]
	score := 0.0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		isTrump := c.GetDesign() == e.trumpSuit
		switch {
		case c.GetValue() == 1: // A
			score += 1.0
		case c.GetValue() == 13: // K
			score += 0.7
		case c.GetValue() == 12: // Q
			score += 0.4
		case isTrump:
			score += 0.3
		}
	}
	bid := int(score)
	if bid > EstimationHandSize {
		bid = EstimationHandSize
	}
	if bid < 0 {
		bid = 0
	}
	return bid
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (e *Estimation) PlayerPlay(cardIndex int) error {
	if e.gameEndFlag {
		return errors.New("game has ended")
	}
	if e.phase != EstimationPhasePlay {
		return errors.New("not the play phase")
	}
	if e.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return e.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (e *Estimation) CpuPlay() {
	if e.gameEndFlag || e.phase != EstimationPhasePlay || e.currentPlayerIdx == 0 {
		return
	}
	_ = e.play(e.currentPlayerIdx, e.chooseCpuCard(e.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (e *Estimation) play(playerIdx, cardIndex int) error {
	p := e.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !e.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	e.currentTrick = append(e.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	e.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(e.currentTrick) < EstimationPlayerCnt {
		e.currentPlayerIdx = (playerIdx + 1) % EstimationPlayerCnt
		return nil
	}
	e.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか
func (e *Estimation) canPlay(playerIdx int, card *Card) bool {
	if len(e.currentTrick) == 0 {
		return true
	}
	leadSuit := e.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := e.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (e *Estimation) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(e.players) {
		return nil
	}
	p := e.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if e.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決する
func (e *Estimation) resolveTrick() {
	winner := e.trickWinner()
	cards := make([]*Card, 0, len(e.currentTrick))
	for _, tc := range e.currentTrick {
		cards = append(cards, tc.Card)
	}
	e.players[winner].AddTrick(cards)

	e.trickNumber++
	e.currentTrick = nil
	e.leadPlayerIdx = winner
	e.currentPlayerIdx = winner

	if e.trickNumber >= EstimationTricksPerRound {
		e.finishRound()
	}
}

// trickWinner 現在のトリックの勝者
func (e *Estimation) trickWinner() int {
	if len(e.currentTrick) == 0 {
		return e.leadPlayerIdx
	}
	leadSuit := e.currentTrick[0].Card.GetDesign()
	bestIdx, best := e.currentTrick[0].PlayerIdx, e.currentTrick[0].Card
	for _, tc := range e.currentTrick[1:] {
		if e.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// beats challenger が currentBest に勝つか
func (e *Estimation) beats(challenger, currentBest *Card, leadSuit int) bool {
	cTrump := challenger.GetDesign() == e.trumpSuit
	bTrump := currentBest.GetDesign() == e.trumpSuit
	if cTrump != bTrump {
		return cTrump
	}
	if challenger.GetDesign() != currentBest.GetDesign() {
		return challenger.GetDesign() == leadSuit
	}
	return estimationRank(challenger) > estimationRank(currentBest)
}

// finishRound 宣言の当否で得点を確定させる
func (e *Estimation) finishRound() {
	for i, p := range e.players {
		got := p.GetTrickCount()
		score := EstimationScoreFor(p.GetBid(), got, p.GetCallType())
		p.SetRoundScore(score)
		p.AddTotalScore(score)
		e.appendLog(i, "score", fmt.Sprintf("宣言%d 獲得%d: %+d", p.GetBid(), got, score), nil)
	}

	if e.roundNumber >= e.config.Rounds {
		e.finishGame()
		return
	}
	e.phase = EstimationPhaseRoundEnd
}

// EstimationScoreFor 宣言 bid・獲得 got・宣言種別 call の増減を返す。
//
//	通常 n    : 的中 +(10+n) / 外し -(10+n)
//	Dash (0) : 的中 +23      / 外し -23
//	Risk     : 上記の 2 倍
//
// **過不足の量では変わらない。** 1 つ足りなくても 5 つ多くても同じ減点で、
// 「ぴたりと当てる」以外に価値が無いのがこのゲームの性格。
func EstimationScoreFor(bid, got int, call EstimationCallType) int {
	base := EstimationBaseScore + bid
	if call == EstimationCallDash {
		base = EstimationDashScore
	}
	if call == EstimationCallRisk {
		base *= 2
	}
	if got == bid {
		return base
	}
	return -base
}

// NextRound 次のラウンドを開始する
func (e *Estimation) NextRound() {
	if e.gameEndFlag || e.phase != EstimationPhaseRoundEnd {
		return
	}
	e.roundNumber++
	e.dealerIdx = (e.dealerIdx + 1) % EstimationPlayerCnt
	e.dealRound()
}

// finishGame 累計得点の最も高いプレイヤーの勝ち
func (e *Estimation) finishGame() {
	e.phase = EstimationPhaseGameEnd
	e.gameEndFlag = true
	bestIdx, best, tied := -1, 0, false
	for i, p := range e.players {
		switch {
		case bestIdx < 0 || p.GetTotalScore() > best:
			bestIdx, best, tied = i, p.GetTotalScore(), false
		case p.GetTotalScore() == best:
			tied = true
		}
	}
	if tied {
		e.winnerIdx = -1
	} else {
		e.winnerIdx = bestIdx
	}
	e.appendLog(-1, "result", fmt.Sprintf("最終得点 %d/%d/%d/%d",
		e.players[0].GetTotalScore(), e.players[1].GetTotalScore(),
		e.players[2].GetTotalScore(), e.players[3].GetTotalScore()), nil)
}

// chooseCpuCard CPU の手。**宣言に足りなければ取りに行き、足りていれば逃げる。**
// このゲームは取りすぎも取らなすぎも同じだけ減点されるので、CPU も「ちょうど」を
// 目指して打つ。
func (e *Estimation) chooseCpuCard(playerIdx int) int {
	valid := e.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := e.players[playerIdx]
	wantsMore := p.GetTrickCount() < p.GetBid()

	if len(e.currentTrick) == 0 {
		return e.pickExtreme(p, valid, wantsMore)
	}
	if wantsMore {
		if idx, ok := e.pickCheapestWinner(p, valid); ok {
			return idx
		}
		// 取れないなら一番弱い札を捨てる。
		return e.pickExtreme(p, valid, false)
	}
	// 取りたくない: 勝たない札のうち一番強いものを吐く。
	bestIdx, bestRank := -1, -1
	for _, i := range valid {
		c := p.GetCard(i)
		if e.wouldWin(c) {
			continue
		}
		if r := estimationRank(c); r > bestRank {
			bestIdx, bestRank = i, r
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	// 全部勝ってしまうなら、せめて一番弱い札で勝つ。
	return e.pickExtreme(p, valid, false)
}

// pickExtreme valid のうち最強 (high) または最弱の札を選ぶ
func (e *Estimation) pickExtreme(p *EstimationPlayer, valid []int, high bool) int {
	bestIdx, bestRank := valid[0], estimationRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := estimationRank(p.GetCard(i))
		if (high && r > bestRank) || (!high && r < bestRank) {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx
}

// pickCheapestWinner トリックを取れる札のうち一番弱いもの
func (e *Estimation) pickCheapestWinner(p *EstimationPlayer, valid []int) (int, bool) {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !e.wouldWin(c) {
			continue
		}
		if r := estimationRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx, bestIdx >= 0
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (e *Estimation) wouldWin(c *Card) bool {
	if c == nil || len(e.currentTrick) == 0 {
		return true
	}
	leadSuit := e.currentTrick[0].Card.GetDesign()
	best := e.currentTrick[0].Card
	for _, tc := range e.currentTrick[1:] {
		if e.beats(tc.Card, best, leadSuit) {
			best = tc.Card
		}
	}
	return e.beats(c, best, leadSuit)
}

// EstimationHint ヒント情報
type EstimationHint struct {
	// CardIndex 推奨する手札のインデックス（宣言・切り札選択中は nil）
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
	// Value 切り札スート、または推奨する宣言数
	Value int
}

// GetHint 人間プレイヤーへの推奨手を返す
func (e *Estimation) GetHint() *EstimationHint {
	if e.gameEndFlag {
		return nil
	}
	if e.phase == EstimationPhaseTrump && e.dealerIdx == 0 {
		return &EstimationHint{Reason: "estimationSelectTrump", Value: e.longestSuit(0)}
	}
	if e.phase == EstimationPhaseBid && e.bidPlayerIdx == 0 {
		bid := e.estimateTricks(0)
		if r := e.GetRestrictedBid(); r >= 0 && bid == r {
			if bid > 0 {
				bid--
			} else {
				bid++
			}
			return &EstimationHint{Reason: "estimationAvoidRestricted", Value: bid}
		}
		if bid == 0 {
			return &EstimationHint{Reason: "estimationDashCall", Value: 0}
		}
		return &EstimationHint{Reason: "estimationBid", Value: bid}
	}
	if !e.IsHumanTurn() || e.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := e.chooseCpuCard(0)
	reason := "estimationDuck"
	if e.players[0].GetTrickCount() < e.players[0].GetBid() {
		reason = "estimationWinTrick"
	}
	return &EstimationHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (e *Estimation) GetPhase() EstimationPhase { return e.phase }

// GetConfig 現在の設定
func (e *Estimation) GetConfig() EstimationConfig { return e.config }

// SetConfig 設定を差し替える
func (e *Estimation) SetConfig(c EstimationConfig) { e.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (e *Estimation) GetRoundNumber() int { return e.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (e *Estimation) GetTrickNumber() int { return e.trickNumber }

// GetTrumpSuit 切り札のスート（未決定は 0）
func (e *Estimation) GetTrumpSuit() int { return e.trumpSuit }

// GetCurrentTrick 現在のトリック
func (e *Estimation) GetCurrentTrick() []*TrickCard { return e.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (e *Estimation) GetCurrentPlayerIdx() int { return e.currentPlayerIdx }

// GetBidPlayerIdx 宣言の手番
func (e *Estimation) GetBidPlayerIdx() int { return e.bidPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (e *Estimation) GetLeadPlayerIdx() int { return e.leadPlayerIdx }

// GetDealerIdx ディーラー
func (e *Estimation) GetDealerIdx() int { return e.dealerIdx }

// GetPlayerCnt プレイヤー数
func (e *Estimation) GetPlayerCnt() int { return len(e.players) }

// GetPlayer 指定インデックスのプレイヤー
func (e *Estimation) GetPlayer(i int) *EstimationPlayer {
	if i < 0 || i >= len(e.players) {
		return nil
	}
	return e.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (e *Estimation) GetGameEndFlag() bool { return e.gameEndFlag }

// GetWinnerIdx 勝利プレイヤー (-1: 未確定/同点)
func (e *Estimation) GetWinnerIdx() int { return e.winnerIdx }

// IsHumanTurn 人間の手番か
func (e *Estimation) IsHumanTurn() bool {
	return !e.gameEndFlag && e.phase == EstimationPhasePlay && e.currentPlayerIdx == 0
}

// IsHumanBidTurn 人間が宣言する番か
func (e *Estimation) IsHumanBidTurn() bool {
	return !e.gameEndFlag && e.phase == EstimationPhaseBid && e.bidPlayerIdx == 0
}

// IsHumanTrumpTurn 人間が切り札を決める番か
func (e *Estimation) IsHumanTrumpTurn() bool {
	return !e.gameEndFlag && e.phase == EstimationPhaseTrump && e.dealerIdx == 0
}

// GiveUp 投了する
func (e *Estimation) GiveUp() {
	if e.gameEndFlag {
		return
	}
	e.phase = EstimationPhaseGameEnd
	e.gameEndFlag = true
	e.winnerIdx = 1
	e.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (e *Estimation) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	e.appendLogAt(e.trickNumber, playerIdx, actionType, detail, cards)
}

// estimationJSON is the KV snapshot format for Estimation.
type estimationJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*EstimationPlayer `json:"pl"`
	Config           EstimationConfig    `json:"cf"`
	Phase            EstimationPhase     `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	TrumpSuit        int                 `json:"ts"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	CurrentPlayerIdx int                 `json:"cp"`
	BidPlayerIdx     int                 `json:"bp"`
	LeadPlayerIdx    int                 `json:"lp"`
	DealerIdx        int                 `json:"di"`
	GameEndFlag      bool                `json:"ge"`
	WinnerIdx        int                 `json:"wi"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (e *Estimation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&estimationJSON{
		TrumpCards:       e.trumpCards,
		Players:          e.players,
		Config:           e.config,
		Phase:            e.phase,
		RoundNumber:      e.roundNumber,
		TrickNumber:      e.trickNumber,
		TrumpSuit:        e.trumpSuit,
		CurrentTrick:     e.currentTrick,
		CurrentPlayerIdx: e.currentPlayerIdx,
		BidPlayerIdx:     e.bidPlayerIdx,
		LeadPlayerIdx:    e.leadPlayerIdx,
		DealerIdx:        e.dealerIdx,
		GameEndFlag:      e.gameEndFlag,
		WinnerIdx:        e.winnerIdx,
		ActionLog:        e.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。値域を検証する。
func (e *Estimation) UnmarshalJSON(data []byte) error {
	var j estimationJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < EstimationPhaseTrump || j.Phase > EstimationPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.TrickNumber < 0 || j.TrickNumber > EstimationTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if len(j.ActionLog) > estimationMaxSliceLen {
		return errors.New("estimation: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > EstimationPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"bid player":     j.BidPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= EstimationPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= EstimationPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.TrumpCards != nil {
		e.trumpCards = j.TrumpCards
	}
	if len(j.Players) == EstimationPlayerCnt {
		e.players = j.Players
	}
	e.config = j.Config
	e.phase = j.Phase
	e.roundNumber = j.RoundNumber
	e.trickNumber = j.TrickNumber
	e.trumpSuit = j.TrumpSuit
	e.currentTrick = j.CurrentTrick
	e.currentPlayerIdx = j.CurrentPlayerIdx
	e.bidPlayerIdx = j.BidPlayerIdx
	e.leadPlayerIdx = j.LeadPlayerIdx
	e.dealerIdx = j.DealerIdx
	e.gameEndFlag = j.GameEndFlag
	e.winnerIdx = j.WinnerIdx
	e.actionLog = j.ActionLog
	return nil
}
