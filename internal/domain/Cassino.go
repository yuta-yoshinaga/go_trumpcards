package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// CassinoPlayerCnt カシノのプレイヤー数 (4人固定)
const CassinoPlayerCnt = 4

// CassinoInitialHandSize 各プレイヤーへの初期配札枚数 (1 パック)
const CassinoInitialHandSize = 4

// CassinoInitialTableSize 最初のラウンド開始時に場へ置くカード枚数
const CassinoInitialTableSize = 4

// Cassino ゲームのフェーズ
const (
	// CassinoPhaseDealing 配札中
	CassinoPhaseDealing = "dealing"
	// CassinoPhasePlayerTurn プレイヤーターン (人間または CPU)
	CassinoPhasePlayerTurn = "playerTurn"
	// CassinoPhaseRoundEnd ラウンド終了 (捕獲の締め処理)
	CassinoPhaseRoundEnd = "roundEnd"
	// CassinoPhaseGameEnd ゲーム終了 (誰かが TargetScore に到達)
	CassinoPhaseGameEnd = "gameEnd"
)

// CassinoActionType カシノでの行動種別
type CassinoActionType string

// CassinoActionType 定数
const (
	// CassinoActionTake 場札を取る
	CassinoActionTake CassinoActionType = "take"
	// CassinoActionBuild ビルドを作成 / 拡張する
	CassinoActionBuild CassinoActionType = "build"
	// CassinoActionTrail 場に置くだけ
	CassinoActionTrail CassinoActionType = "trail"
)

// CassinoAction はプレイヤー (人間/CPU) 1 ターン分の行動記録。
type CassinoAction struct {
	PlayerIdx     int               // 行動したプレイヤーインデックス
	Type          CassinoActionType // 行動種別
	PlayedCard    *Card             // 出した手札 1 枚
	CapturedCards []*Card           // 捕獲したカード (take 時のみ)
	BuildValue    int               // ビルド宣言値 (build 時のみ)
	IsSweep       bool              // スイープ発生
}

// cassinoRoundState はラウンドごとにリセットされる状態。
type cassinoRoundState struct {
	phase          string           // 現在のフェーズ
	currentTurn    int              // 現在の手番
	tableCards     []*Card          // 場の単独カード (ビルド除く)
	builds         []*CassinoBuild  // 場のビルド
	lastCaptureIdx int              // 最後に捕獲したプレイヤー (-1 = なし)
	humanAction    *CassinoAction   // 人間の最後の行動
	cpuActions     []*CassinoAction // 人間ターン後の CPU 行動履歴
	actionLogBase
	packsDealt      int  // これまでに配ったパック数 (1 回の配布 = 4 枚/人)
	gameEndFlag     bool // ゲーム終了フラグ (TargetScore 到達)
	roundWinners    []int
	lastRoundScores map[int]int // 前ラウンドの内訳得点 (プレイヤー別)
	lastRoundDetail *CassinoScoreDetail
}

// CassinoScoreDetail は 1 ラウンドの得点内訳。
type CassinoScoreDetail struct {
	Cards           map[int]int // プレイヤー別捕獲枚数
	Spades          map[int]int // プレイヤー別スペード数
	Aces            map[int]int // プレイヤー別エース数
	HasBigCasino    int         // 10♦ を取ったプレイヤー (-1 = なし)
	HasLittleCasino int         // 2♠ を取ったプレイヤー (-1 = なし)
	Sweeps          map[int]int // プレイヤー別スイープ数
	Gained          map[int]int // プレイヤー別にこのラウンドで得た点数
}

// Cassino はカシノゲームの状態を保持する集約ルート。
type Cassino struct {
	trumpCards *TrumpCards
	players    []*CassinoPlayer
	config     CassinoConfig
	round      cassinoRoundState
}

// NewCassino コンストラクタ。
func NewCassino(trumpCards *TrumpCards, players []*CassinoPlayer, config CassinoConfig) *Cassino {
	return &Cassino{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: cassinoRoundState{
			phase:          CassinoPhaseDealing,
			lastCaptureIdx: -1,
		},
	}
}

// NewDefaultCassino returns a Cassino with the standard 4-player setup
// (1 human, 3 CPU) and DefaultCassinoConfig.
func NewDefaultCassino() *Cassino {
	config := DefaultCassinoConfig()
	players := make([]*CassinoPlayer, CassinoPlayerCnt)
	players[0] = NewCassinoPlayer(true)
	for i := 1; i < CassinoPlayerCnt; i++ {
		players[i] = NewCassinoPlayer(false)
	}
	return NewCassino(NewTrumpCards(0), players, config)
}

// Reset は新しい「ゲーム」を開始する。累計得点もクリアする。
func (c *Cassino) Reset() {
	for _, p := range c.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
		p.ResetSweepCount()
		p.ResetTotalScore()
	}

	c.trumpCards = NewTrumpCards(0)
	c.trumpCards.Shuffle()

	c.round = cassinoRoundState{
		phase:          CassinoPhaseDealing,
		lastCaptureIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}

	c.startRound(true)
}

// NextRound は次のラウンドを開始する。
// 捕獲札、スイープ数はラウンドごとにクリア (累計得点は維持)。
func (c *Cassino) NextRound() {
	if c.round.gameEndFlag {
		return
	}
	for _, p := range c.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
		p.ResetSweepCount()
	}
	c.trumpCards = NewTrumpCards(0)
	c.trumpCards.Shuffle()
	c.round.phase = CassinoPhaseDealing
	c.round.currentTurn = 0
	c.round.tableCards = nil
	c.round.builds = nil
	c.round.lastCaptureIdx = -1
	c.round.humanAction = nil
	c.round.cpuActions = nil
	c.round.packsDealt = 0
	c.startRound(false)
}

// startRound は配札と初期場札を設定する。
func (c *Cassino) startRound(isFirstRound bool) {
	c.dealNextPack()
	if isFirstRound {
		// 初ラウンドの先頭: 場に 4 枚配る
		for i := 0; i < CassinoInitialTableSize; i++ {
			card := c.trumpCards.DrawCard()
			if card == nil {
				break
			}
			c.round.tableCards = append(c.round.tableCards, card)
		}
		c.appendLog(-1, "deal", fmt.Sprintf("dealt %d table cards", len(c.round.tableCards)), c.round.tableCards)
	}
	c.round.phase = CassinoPhasePlayerTurn
}

// dealNextPack は各プレイヤーに CassinoInitialHandSize 枚配る。
// 山札が尽きたら部分的に配って終わる。
func (c *Cassino) dealNextPack() {
	for k := 0; k < CassinoInitialHandSize; k++ {
		for i := 0; i < len(c.players); i++ {
			card := c.trumpCards.DrawCard()
			if card == nil {
				c.round.packsDealt++
				return
			}
			c.players[i].AddCard(card)
		}
	}
	c.round.packsDealt++
}

// allHandsEmpty は全員の手札が空か。
func (c *Cassino) allHandsEmpty() bool {
	return allHandsEmpty(c.players)
}

// PlayerTake は人間プレイヤーが take を実行する。
func (c *Cassino) PlayerTake(handIdx int, tableIdxs []int, buildIdxs []int) error {
	if err := c.guardHumanTurn(); err != nil {
		return err
	}
	c.round.cpuActions = nil
	return c.applyTake(c.round.currentTurn, handIdx, tableIdxs, buildIdxs, c.setHumanAction)
}

// PlayerBuild は人間プレイヤーが build を実行する。
func (c *Cassino) PlayerBuild(handIdx int, tableIdxs []int, declaredValue int) error {
	if err := c.guardHumanTurn(); err != nil {
		return err
	}
	c.round.cpuActions = nil
	return c.applyBuild(c.round.currentTurn, handIdx, tableIdxs, declaredValue, c.setHumanAction)
}

// PlayerTrail は人間プレイヤーが trail を実行する。
func (c *Cassino) PlayerTrail(handIdx int) error {
	if err := c.guardHumanTurn(); err != nil {
		return err
	}
	c.round.cpuActions = nil
	return c.applyTrail(c.round.currentTurn, handIdx, c.setHumanAction)
}

// CpuPlay は CPU のターンを 1 回進める。
func (c *Cassino) CpuPlay() {
	if c.round.gameEndFlag || c.round.phase != CassinoPhasePlayerTurn {
		return
	}
	if c.players[c.round.currentTurn].GetIsHuman() {
		return
	}
	playerIdx := c.round.currentTurn
	plan := c.chooseCpuAction(playerIdx)
	switch plan.Type {
	case CassinoActionTake:
		_ = c.applyTake(playerIdx, plan.handIdx, plan.tableIdxs, plan.buildIdxs, c.appendCpuAction)
	case CassinoActionBuild:
		_ = c.applyBuild(playerIdx, plan.handIdx, plan.tableIdxs, plan.buildValue, c.appendCpuAction)
	default:
		_ = c.applyTrail(playerIdx, plan.handIdx, c.appendCpuAction)
	}
}

// guardHumanTurn は人間ターンかつゲーム進行中か確認。
func (c *Cassino) guardHumanTurn() error {
	if c.round.gameEndFlag {
		return ErrGameEnded
	}
	if c.round.phase != CassinoPhasePlayerTurn {
		return NewDomainError(ErrWrongPhase, "not in player turn phase")
	}
	if !c.players[c.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// setHumanAction は human の action を記録。
func (c *Cassino) setHumanAction(a *CassinoAction) { c.round.humanAction = a }

// appendCpuAction は cpu の action を記録。
func (c *Cassino) appendCpuAction(a *CassinoAction) {
	c.round.cpuActions = append(c.round.cpuActions, a)
}

// applyTake は take アクションを実行する共通処理。
func (c *Cassino) applyTake(playerIdx, handIdx int, tableIdxs, buildIdxs []int, record func(*CassinoAction)) error {
	player := c.players[playerIdx]
	handCard := player.GetCard(handIdx)
	if handCard == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}

	// 場札の検証
	var takenTableIdxGroups [][]int
	var takenTable []*Card
	if len(tableIdxs) > 0 {
		ok, groups := isValidTakeSelection(handCard, c.round.tableCards, tableIdxs)
		if !ok {
			return NewDomainError(ErrInvalidPlay, "selected table cards do not form valid captures")
		}
		takenTableIdxGroups = groups
		for _, idx := range tableIdxs {
			takenTable = append(takenTable, c.round.tableCards[idx])
		}
	}

	// ビルドの検証: 指定されたビルドは全て値が handCard と一致
	var takenBuilds []*CassinoBuild
	if len(buildIdxs) > 0 {
		for _, bi := range buildIdxs {
			if bi < 0 || bi >= len(c.round.builds) {
				return NewDomainError(ErrInvalidPlay, fmt.Sprintf("build index %d out of range", bi))
			}
			b := c.round.builds[bi]
			if CassinoCardValue(handCard) != b.Value {
				return NewDomainError(ErrInvalidPlay, fmt.Sprintf("hand card does not match build value %d", b.Value))
			}
			takenBuilds = append(takenBuilds, b)
		}
	}

	if len(takenTable) == 0 && len(takenBuilds) == 0 {
		return NewDomainError(ErrInvalidPlay, "take requires at least one target")
	}

	// 変更を確定
	_ = player.RemoveCard(handIdx)
	// 場から取る (降順で index 削除)
	c.removeTableCardsByIndex(tableIdxs)
	// ビルドから取る
	c.removeBuildsByIndex(buildIdxs)

	captured := make([]*Card, 0)
	captured = append(captured, handCard)
	captured = append(captured, takenTable...)
	for _, b := range takenBuilds {
		captured = append(captured, b.AllCards()...)
	}
	player.AddCaptured(captured)
	c.round.lastCaptureIdx = playerIdx

	isSweep := len(c.round.tableCards) == 0 && len(c.round.builds) == 0 && !c.lastTakeInRound()
	if isSweep && c.config.SweepBonusEnabled {
		player.IncrementSweep()
	}

	action := &CassinoAction{
		PlayerIdx:     playerIdx,
		Type:          CassinoActionTake,
		PlayedCard:    handCard,
		CapturedCards: captured[1:],
		IsSweep:       isSweep,
	}
	record(action)
	c.appendLog(playerIdx, string(CassinoActionTake), fmt.Sprintf("took %d card(s)", len(captured)-1), captured)
	// 内部状態を前進
	_ = takenTableIdxGroups // 検証用途のみ (将来 UI 表示で使う可能性)
	c.postActionAdvance()
	return nil
}

// applyBuild は build アクションを実行する共通処理。
func (c *Cassino) applyBuild(playerIdx, handIdx int, tableIdxs []int, declaredValue int, record func(*CassinoAction)) error {
	player := c.players[playerIdx]
	handCard := player.GetCard(handIdx)
	if handCard == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	if declaredValue < 2 || declaredValue > 10 {
		return NewDomainError(ErrInvalidPlay, "build value must be between 2 and 10")
	}
	if CassinoIsFaceCard(handCard) {
		return NewDomainError(ErrInvalidPlay, "cannot build with face cards")
	}
	if CassinoCardValue(handCard) > declaredValue {
		return NewDomainError(ErrInvalidPlay, "hand card exceeds declared build value")
	}
	// 手札に declaredValue と同じ値のカードが別に残っている必要あり
	if !c.playerHasCaptureCard(player, handIdx, declaredValue) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("you must hold another %d to capture this build", declaredValue))
	}
	// 選択した場札は絵札を含まない
	if len(tableIdxs) == 0 {
		return NewDomainError(ErrInvalidPlay, "build requires at least one table card")
	}
	seen := make(map[int]bool)
	var selectedCards []*Card
	sum := CassinoCardValue(handCard)
	for _, idx := range tableIdxs {
		if idx < 0 || idx >= len(c.round.tableCards) || seen[idx] {
			return NewDomainError(ErrInvalidPlay, "invalid table card selection")
		}
		seen[idx] = true
		tc := c.round.tableCards[idx]
		if CassinoIsFaceCard(tc) {
			return NewDomainError(ErrInvalidPlay, "cannot include face cards in a build")
		}
		selectedCards = append(selectedCards, tc)
		sum += CassinoCardValue(tc)
	}
	if sum != declaredValue {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("sum %d does not match declared build value %d", sum, declaredValue))
	}

	// 確定
	_ = player.RemoveCard(handIdx)
	groupCards := make([]*Card, 0, len(selectedCards)+1)
	groupCards = append(groupCards, handCard)
	groupCards = append(groupCards, selectedCards...)
	c.removeTableCardsByIndex(tableIdxs)

	// 既存ビルドに値一致のものがあり、自分が複合ビルドを許可する設定なら合流
	merged := false
	if c.config.MultiBuildEnabled {
		for _, b := range c.round.builds {
			if b.Value == declaredValue && b.OwnerIdx == playerIdx {
				b.AddGroup(groupCards)
				merged = true
				break
			}
		}
	}
	if !merged {
		c.round.builds = append(c.round.builds, NewCassinoBuild(playerIdx, declaredValue, groupCards))
	}

	action := &CassinoAction{
		PlayerIdx:  playerIdx,
		Type:       CassinoActionBuild,
		PlayedCard: handCard,
		BuildValue: declaredValue,
	}
	record(action)
	c.appendLog(playerIdx, string(CassinoActionBuild), fmt.Sprintf("built value %d", declaredValue), groupCards)
	c.postActionAdvance()
	return nil
}

// applyTrail は trail アクションを実行する共通処理。
func (c *Cassino) applyTrail(playerIdx, handIdx int, record func(*CassinoAction)) error {
	player := c.players[playerIdx]
	handCard := player.GetCard(handIdx)
	if handCard == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	// 自分が保有するビルドがあると trail できない
	if c.playerOwnsBuild(playerIdx) {
		return NewDomainError(ErrInvalidPlay, "must resolve your existing build before trailing")
	}
	_ = player.RemoveCard(handIdx)
	c.round.tableCards = append(c.round.tableCards, handCard)
	action := &CassinoAction{
		PlayerIdx:  playerIdx,
		Type:       CassinoActionTrail,
		PlayedCard: handCard,
	}
	record(action)
	c.appendLog(playerIdx, string(CassinoActionTrail), "trailed", []*Card{handCard})
	c.postActionAdvance()
	return nil
}

// postActionAdvance はアクション後の共通進行処理。
// ラウンド終了判定を行い、そうでなければ次の手番に進める。
func (c *Cassino) postActionAdvance() {
	if c.isRoundOver() {
		c.finishRound()
		return
	}
	// 手番を次へ
	c.round.currentTurn = (c.round.currentTurn + 1) % len(c.players)
	// 手札切れ + 山札あり → 全員へ次のパックを配る
	if c.allHandsEmpty() && c.trumpCards.GetRemainingCount() > 0 {
		c.dealNextPack()
	}
}

// isRoundOver は現在のラウンドが終了しているか。
// 手札 0 + 山札 0 の時点で終了。
func (c *Cassino) isRoundOver() bool {
	return c.allHandsEmpty() && c.trumpCards.GetRemainingCount() == 0
}

// finishRound はラウンド終了処理: 残りの場札を最後に捕獲したプレイヤーに渡し、スコア計算。
func (c *Cassino) finishRound() {
	c.round.phase = CassinoPhaseRoundEnd
	// 残りの場札 + ビルドを最後の捕獲者に渡す (スイープにはしない)
	leftover := make([]*Card, 0)
	leftover = append(leftover, c.round.tableCards...)
	for _, b := range c.round.builds {
		leftover = append(leftover, b.AllCards()...)
	}
	c.round.tableCards = nil
	c.round.builds = nil
	if c.round.lastCaptureIdx >= 0 && len(leftover) > 0 {
		c.players[c.round.lastCaptureIdx].AddCaptured(leftover)
		c.appendLog(c.round.lastCaptureIdx, "lastTake", fmt.Sprintf("last-take: %d card(s)", len(leftover)), leftover)
	}
	// スコア計算
	detail := c.scoreRound()
	c.round.lastRoundDetail = detail
	c.round.lastRoundScores = detail.Gained
	for i, p := range c.players {
		p.AddScore(detail.Gained[i])
	}
	// ゲーム終了判定
	maxScore := 0
	for _, p := range c.players {
		if p.GetTotalScore() > maxScore {
			maxScore = p.GetTotalScore()
		}
	}
	if maxScore >= c.config.TargetScore {
		c.round.gameEndFlag = true
		c.round.phase = CassinoPhaseGameEnd
		// ラウンド勝者リストを作る
		winners := make([]int, 0)
		for i, p := range c.players {
			if p.GetTotalScore() == maxScore {
				winners = append(winners, i)
			}
		}
		c.round.roundWinners = winners
		c.appendLog(-1, "gameEnd", fmt.Sprintf("game ended at %d points", maxScore), nil)
	} else {
		c.appendLog(-1, "roundEnd", "round ended", nil)
	}
}

// scoreRound はラウンドの得点内訳を計算する。
func (c *Cassino) scoreRound() *CassinoScoreDetail {
	det := &CassinoScoreDetail{
		Cards:           make(map[int]int),
		Spades:          make(map[int]int),
		Aces:            make(map[int]int),
		Sweeps:          make(map[int]int),
		Gained:          make(map[int]int),
		HasBigCasino:    -1,
		HasLittleCasino: -1,
	}
	for i, p := range c.players {
		det.Cards[i] = p.CapturedCount()
		det.Sweeps[i] = p.GetSweepCount()
		for _, card := range p.GetCapturedCards() {
			if CassinoIsSpade(card) {
				det.Spades[i]++
			}
			if CassinoIsAce(card) {
				det.Aces[i]++
			}
			if CassinoIsBigCasino(card) {
				det.HasBigCasino = i
			}
			if CassinoIsLittleCasino(card) {
				det.HasLittleCasino = i
			}
		}
	}
	// 最多カード / 最多スペードを決定 (同点なら誰ももらえない)
	mostCardsIdx := uniqueMaxIndex(det.Cards)
	mostSpadesIdx := uniqueMaxIndex(det.Spades)
	for i := range c.players {
		score := 0
		if i == mostCardsIdx {
			score += CassinoScoreMostCards
		}
		if i == mostSpadesIdx {
			score += CassinoScoreMostSpades
		}
		if det.HasBigCasino == i {
			score += CassinoScoreBigCasino
		}
		if det.HasLittleCasino == i {
			score += CassinoScoreLittleCasino
		}
		score += det.Aces[i] * CassinoScoreAce
		if c.config.SweepBonusEnabled {
			score += det.Sweeps[i] * CassinoScoreSweep
		}
		det.Gained[i] = score
	}
	return det
}

// uniqueMaxIndex はマップの中で最大値かつ単独のキーを返す。同点または空なら -1。
func uniqueMaxIndex(m map[int]int) int {
	best := -1
	bestVal := 0
	tie := false
	for k, v := range m {
		if best == -1 || v > bestVal {
			best = k
			bestVal = v
			tie = false
		} else if v == bestVal {
			tie = true
		}
	}
	if tie || bestVal == 0 {
		return -1
	}
	return best
}

// lastTakeInRound は「場に残っているものがこのターンですべて空になり、かつラウンド終了と重なる」か。
// スイープにはラウンド末の "last take" を含めないため、この判定で弾く。
func (c *Cassino) lastTakeInRound() bool {
	// postActionAdvance 直前の判定なのでこの時点では仮に true / false いずれでも良いが、
	// 「手札が全員 0 かつ山札 0」であれば以降のカードプレイはなくなり、
	// その最終 take をスイープに含めないという規則を表現。
	return c.allHandsEmpty() && c.trumpCards.GetRemainingCount() == 0
}

// playerHasCaptureCard は player が handIdx 以外に declaredValue と一致する値のカードを持っているか。
func (c *Cassino) playerHasCaptureCard(player *CassinoPlayer, handIdx, declaredValue int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if i == handIdx {
			continue
		}
		card := player.GetCard(i)
		if card == nil || CassinoIsFaceCard(card) {
			continue
		}
		if CassinoCardValue(card) == declaredValue {
			return true
		}
	}
	return false
}

// playerOwnsBuild は player がビルドを保有しているか。
func (c *Cassino) playerOwnsBuild(playerIdx int) bool {
	for _, b := range c.round.builds {
		if b.OwnerIdx == playerIdx {
			return true
		}
	}
	return false
}

// sortIndicesDescending returns a copy of idxs sorted in descending order.
// Used to delete by index from the back so earlier deletes don't shift later
// targets.
func sortIndicesDescending(idxs []int) []int {
	out := make([]int, len(idxs))
	copy(out, idxs)
	sort.Sort(sort.Reverse(sort.IntSlice(out)))
	return out
}

// removeTableCardsByIndex は降順に並び替えてから tableCards を削除する。
func (c *Cassino) removeTableCardsByIndex(idxs []int) {
	c.round.tableCards = removeIndices(c.round.tableCards, idxs)
}

// removeBuildsByIndex は降順に並び替えてから builds を削除する。
func (c *Cassino) removeBuildsByIndex(idxs []int) {
	if len(idxs) == 0 {
		return
	}
	for _, idx := range sortIndicesDescending(idxs) {
		if idx >= 0 && idx < len(c.round.builds) {
			c.round.builds = append(c.round.builds[:idx], c.round.builds[idx+1:]...)
		}
	}
}

// appendLog 棋譜にエントリを追加する。
func (c *Cassino) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	c.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn 現在の手番が人間かどうか
func (c *Cassino) IsHumanTurn() bool {
	if c.round.gameEndFlag {
		return false
	}
	return c.players[c.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (c *Cassino) GetCurrentTurn() int { return c.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (c *Cassino) GetGameEndFlag() bool { return c.round.gameEndFlag }

// GetTableCards 場の単独カード取得。
func (c *Cassino) GetTableCards() []*Card { return c.round.tableCards }

// GetBuilds 場のビルド取得。
func (c *Cassino) GetBuilds() []*CassinoBuild { return c.round.builds }

// GetPlayer プレイヤー取得
func (c *Cassino) GetPlayer(idx int) *CassinoPlayer {
	return getPlayer(c.players, idx)
}

// GetPlayerCnt プレイヤー数取得
func (c *Cassino) GetPlayerCnt() int { return len(c.players) }

// GetCpuActions CPU ターンの行動履歴取得
func (c *Cassino) GetCpuActions() []*CassinoAction { return c.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (c *Cassino) GetHumanAction() *CassinoAction { return c.round.humanAction }

// GetConfig ローカルルール設定取得
func (c *Cassino) GetConfig() CassinoConfig { return c.config }

// SetConfig ローカルルール設定を変更
func (c *Cassino) SetConfig(config CassinoConfig) { c.config = config }

// GetActionLog 棋譜取得
func (c *Cassino) GetActionLog() []*ActionLogEntry { return c.round.actionLog }

// GetPhase 現在のフェーズ取得
func (c *Cassino) GetPhase() string { return c.round.phase }

// GetLastRoundDetail 直前ラウンドの得点詳細取得 (nil の場合もあり得る)
func (c *Cassino) GetLastRoundDetail() *CassinoScoreDetail { return c.round.lastRoundDetail }

// GetLastCaptureIdx 最後に捕獲したプレイヤー (-1 = なし)
func (c *Cassino) GetLastCaptureIdx() int { return c.round.lastCaptureIdx }

// GetRoundWinners ゲーム終了時の勝者リスト。
func (c *Cassino) GetRoundWinners() []int { return c.round.roundWinners }

// GetRemainingDeck 山札の残り枚数。
func (c *Cassino) GetRemainingDeck() int { return c.trumpCards.GetRemainingCount() }

// GetPacksDealt これまでに配布されたパック数。
func (c *Cassino) GetPacksDealt() int { return c.round.packsDealt }

// --- JSON Serialization ---

// cassinoActionJSON is the JSON wire format for CassinoAction.
type cassinoActionJSON struct {
	PlayerIdx     int               `json:"pi"`
	Type          CassinoActionType `json:"ty"`
	PlayedCard    *Card             `json:"pc"`
	CapturedCards []*Card           `json:"cc"`
	BuildValue    int               `json:"bv"`
	IsSweep       bool              `json:"sw"`
}

// MarshalJSON implements json.Marshaler.
func (a *CassinoAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(cassinoActionJSON{
		PlayerIdx:     a.PlayerIdx,
		Type:          a.Type,
		PlayedCard:    a.PlayedCard,
		CapturedCards: a.CapturedCards,
		BuildValue:    a.BuildValue,
		IsSweep:       a.IsSweep,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *CassinoAction) UnmarshalJSON(data []byte) error {
	var j cassinoActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.Type = j.Type
	a.PlayedCard = j.PlayedCard
	a.CapturedCards = j.CapturedCards
	a.BuildValue = j.BuildValue
	a.IsSweep = j.IsSweep
	return nil
}

// cassinoScoreDetailJSON is the JSON wire format for CassinoScoreDetail.
type cassinoScoreDetailJSON struct {
	Cards           map[int]int `json:"cd"`
	Spades          map[int]int `json:"sp"`
	Aces            map[int]int `json:"ac"`
	HasBigCasino    int         `json:"bc"`
	HasLittleCasino int         `json:"lc"`
	Sweeps          map[int]int `json:"sw"`
	Gained          map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *CassinoScoreDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(cassinoScoreDetailJSON{
		Cards:           d.Cards,
		Spades:          d.Spades,
		Aces:            d.Aces,
		HasBigCasino:    d.HasBigCasino,
		HasLittleCasino: d.HasLittleCasino,
		Sweeps:          d.Sweeps,
		Gained:          d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *CassinoScoreDetail) UnmarshalJSON(data []byte) error {
	var j cassinoScoreDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Cards = j.Cards
	d.Spades = j.Spades
	d.Aces = j.Aces
	d.HasBigCasino = j.HasBigCasino
	d.HasLittleCasino = j.HasLittleCasino
	d.Sweeps = j.Sweeps
	d.Gained = j.Gained
	return nil
}

// cassinoJSON is the JSON wire format for Cassino.
type cassinoJSON struct {
	TrumpCards      *TrumpCards         `json:"tc"`
	Players         []*CassinoPlayer    `json:"pl"`
	Config          CassinoConfig       `json:"cf"`
	Phase           string              `json:"ph"`
	CurrentTurn     int                 `json:"ct"`
	TableCards      []*Card             `json:"tb"`
	Builds          []*CassinoBuild     `json:"bl"`
	LastCaptureIdx  int                 `json:"lc"`
	HumanAction     *CassinoAction      `json:"ha"`
	CpuActions      []*CassinoAction    `json:"ca"`
	ActionLog       []*ActionLogEntry   `json:"al"`
	PacksDealt      int                 `json:"pd"`
	GameEndFlag     bool                `json:"ge"`
	RoundWinners    []int               `json:"rw"`
	LastRoundScores map[int]int         `json:"ls"`
	LastRoundDetail *CassinoScoreDetail `json:"ld"`
}

// cassinoMaxSliceLen caps slice sizes during deserialisation.
const cassinoMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (c *Cassino) MarshalJSON() ([]byte, error) {
	return json.Marshal(cassinoJSON{
		TrumpCards:      c.trumpCards,
		Players:         c.players,
		Config:          c.config,
		Phase:           c.round.phase,
		CurrentTurn:     c.round.currentTurn,
		TableCards:      c.round.tableCards,
		Builds:          c.round.builds,
		LastCaptureIdx:  c.round.lastCaptureIdx,
		HumanAction:     c.round.humanAction,
		CpuActions:      c.round.cpuActions,
		ActionLog:       c.round.actionLog,
		PacksDealt:      c.round.packsDealt,
		GameEndFlag:     c.round.gameEndFlag,
		RoundWinners:    c.round.roundWinners,
		LastRoundScores: c.round.lastRoundScores,
		LastRoundDetail: c.round.lastRoundDetail,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Cassino) UnmarshalJSON(data []byte) error {
	var j cassinoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > cassinoMaxSliceLen || len(j.TableCards) > cassinoMaxSliceLen ||
		len(j.Builds) > cassinoMaxSliceLen || len(j.CpuActions) > cassinoMaxSliceLen ||
		len(j.ActionLog) > cassinoMaxSliceLen {
		return fmt.Errorf("cassino: input array exceeds maximum allowed size")
	}
	if j.TrumpCards == nil {
		return fmt.Errorf("cassino: missing trump cards in state")
	}
	c.trumpCards = j.TrumpCards
	c.players = j.Players
	if c.players == nil {
		c.players = make([]*CassinoPlayer, 0)
	}
	c.config = j.Config
	c.round = cassinoRoundState{
		phase:           j.Phase,
		currentTurn:     j.CurrentTurn,
		tableCards:      j.TableCards,
		builds:          j.Builds,
		lastCaptureIdx:  j.LastCaptureIdx,
		humanAction:     j.HumanAction,
		cpuActions:      j.CpuActions,
		actionLogBase:   actionLogBase{actionLog: j.ActionLog},
		packsDealt:      j.PacksDealt,
		gameEndFlag:     j.GameEndFlag,
		roundWinners:    j.RoundWinners,
		lastRoundScores: j.LastRoundScores,
		lastRoundDetail: j.LastRoundDetail,
	}
	if c.round.actionLog == nil {
		c.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
