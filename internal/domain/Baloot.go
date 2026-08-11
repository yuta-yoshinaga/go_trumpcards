//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BalootMode バルートのモード。**同じハンドの中でトリック評価の規則そのものが変わる。**
type BalootMode int

// Balootのモード定数
const (
	// BalootModeNone 未宣言
	BalootModeNone BalootMode = iota
	// BalootModeSun 切り札なし。A が最強で、点数も強さも 1 通り。
	BalootModeSun
	// BalootModeHokom 切り札あり。切り札だけ J→9→A→10→K→Q→8→7 の序列になる。
	BalootModeHokom
)

// BalootPhase バルートのゲームフェーズ
type BalootPhase int

// Balootのフェーズ定数
const (
	// BalootPhaseDeclare モード宣言中（5 枚だけ配られている）
	BalootPhaseDeclare BalootPhase = iota
	// BalootPhasePlay プレイ中
	BalootPhasePlay
	// BalootPhaseRoundEnd ラウンド終了
	BalootPhaseRoundEnd
	// BalootPhaseGameEnd ゲーム終了
	BalootPhaseGameEnd
)

// BalootPlayerCnt プレイヤー数（4 人固定・2 対 2）
const BalootPlayerCnt = 4

// BalootTeamCnt チーム数
const BalootTeamCnt = 2

// BalootFirstDealSize モード宣言の前に配る枚数
const BalootFirstDealSize = 5

// BalootHandSize 各プレイヤーの最終的な手札枚数
const BalootHandSize = 8

// BalootTricksPerRound 1 ラウンドのトリック数
const BalootTricksPerRound = BalootHandSize

// BalootLastTrickBonus 最終トリックの加点
const BalootLastTrickBonus = 10

// BalootBonus Baloot（Hokom で切り札の K+Q）の加点
const BalootBonus = 20

// balootMaxSliceLen caps slice sizes during deserialisation.
const balootMaxSliceLen = 1000

// Baloot バルート ゲームクラス。
//
// サウジアラビアをはじめ湾岸諸国で最も遊ばれているトリックテイキング。フランスの
// ベロートから派生した 4 人 2 対 2（向かい合う席が味方）で、32 枚 (A,7〜K) を使う。
//
// **このゲームの肝は、モードによってトリック評価の規則そのものが入れ替わること。**
// 既存のどのゲームにも無い形で、同じ 32 枚が 2 通りの序列と 2 通りの点数表を持つ:
//
//	Sun   : 切り札なし。A>10>K>Q>J>9>8>7、点は A11/10=10/K4/Q3/J2 → 1 ラウンド 120 点
//	Hokom : 切り札あり。**切り札だけ** J>9>A>10>K>Q>8>7、点は J20/9=14/A11/10=10/K4/Q3
//	        非切り札は Sun と同じ → 1 ラウンド 152 点
//
// どちらも最終トリックに 10 点が乗る（Sun 130 / Hokom 162）。
//
// 配りは 2 段構え。**8 枚 × 4 人 = 32 枚はデッキ全部**なので、先に配り切ると
// 宣言のための情報を伏せたまま配り終えてしまう。まず 5 枚ずつ配って
// Sun / Hokom / パスを競り、決まってから残り 3 枚ずつを配る。
//
// **Baloot（役名そのもの）は Hokom のときだけ。** 切り札の K+Q を同じ手に持つと
// 20 点。Sun には切り札が無いので成立しない。
type Baloot struct {
	trumpCards *TrumpCards
	players    []*BalootPlayer
	config     BalootConfig

	phase       BalootPhase
	mode        BalootMode
	roundNumber int
	trickNumber int
	// trumpSuit は Hokom のときだけ意味を持つ。
	trumpSuit int
	// declarerIdx はモードを宣言したプレイヤー (-1: 未決定)
	declarerIdx int

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int

	scores      [BalootTeamCnt]int
	roundPoints [BalootTeamCnt]int

	gameEndFlag bool
	winnerTeam  int

	actionLogBase
}

// NewBaloot コンストラクタ
func NewBaloot(trumpCards *TrumpCards, players []*BalootPlayer, config BalootConfig) *Baloot {
	return &Baloot{trumpCards: trumpCards, players: players, config: config, declarerIdx: -1, winnerTeam: -1}
}

// NewDefaultBaloot 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultBaloot() *Baloot {
	players := make([]*BalootPlayer, 0, BalootPlayerCnt)
	for i := range BalootPlayerCnt {
		players = append(players, NewBalootPlayer(i == 0))
	}
	return NewBaloot(NewTrumpCardsBelote(), players, DefaultBalootConfig())
}

// BalootTeamOf 席のチーム番号。**向かい合う席が味方。**
func BalootTeamOf(playerIdx int) int { return playerIdx % BalootTeamCnt }

// Reset ゲーム全体を初期化する
func (b *Baloot) Reset() {
	b.roundNumber = 1
	b.dealerIdx = 0
	b.gameEndFlag = false
	b.winnerTeam = -1
	b.scores = [BalootTeamCnt]int{}
	b.actionLog = nil
	for _, p := range b.players {
		p.ResetGame()
	}
	b.dealRound()
}

// dealRound 5 枚ずつ配って宣言フェーズに入る
func (b *Baloot) dealRound() {
	b.phase = BalootPhaseDeclare
	b.mode = BalootModeNone
	b.trickNumber = 0
	b.currentTrick = nil
	b.roundPoints = [BalootTeamCnt]int{}
	b.declarerIdx = -1
	b.trumpSuit = 0
	for _, p := range b.players {
		p.ResetRound()
	}

	b.trumpCards = NewTrumpCardsBelote()
	b.trumpCards.Shuffle()
	// **まず 5 枚だけ。** 8 枚 × 4 人 = 32 枚はデッキ全部なので、配り切ってから
	// 宣言させると「何を見て決めるか」が無くなる。
	for range BalootFirstDealSize {
		for i := range BalootPlayerCnt {
			idx := (b.dealerIdx + 1 + i) % BalootPlayerCnt
			if c := b.trumpCards.DrawCard(); c != nil {
				b.players[idx].AddCard(c)
			}
		}
	}
	b.leadPlayerIdx = (b.dealerIdx + 1) % BalootPlayerCnt
	b.currentPlayerIdx = b.leadPlayerIdx
	b.sortAllHands()
	b.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d を開始", b.roundNumber), nil)
}

// sortAllHands 手札を並べ替える。**モードで強さが変わるので並びも変わる。**
func (b *Baloot) sortAllHands() {
	for _, p := range b.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return b.rankOf(ci) < b.rankOf(cj)
		})
	}
}

// DeclareSun 人間プレイヤーが Sun（切り札なし）を宣言する
func (b *Baloot) DeclareSun() error {
	if err := b.guardDeclare(); err != nil {
		return err
	}
	b.accept(0, BalootModeSun, 0)
	return nil
}

// DeclareHokom 人間プレイヤーが Hokom（切り札あり）を宣言する
func (b *Baloot) DeclareHokom(suit int) error {
	if err := b.guardDeclare(); err != nil {
		return err
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", suit)
	}
	b.accept(0, BalootModeHokom, suit)
	return nil
}

// PassDeclaration 人間プレイヤーが宣言を見送る
func (b *Baloot) PassDeclaration() error {
	if err := b.guardDeclare(); err != nil {
		return err
	}
	// **親は見送れない。** 全員が見送ると誰もモードを決めないまま進めなくなる。
	if b.dealerIdx == 0 {
		return errors.New("the dealer must declare")
	}
	b.players[0].SetDeclared(true)
	b.appendLog(0, "pass", "宣言を見送った", nil)
	b.advanceDeclare()
	return nil
}

// guardDeclare 宣言できる状態かを確かめる
func (b *Baloot) guardDeclare() error {
	if b.gameEndFlag {
		return errors.New("game has ended")
	}
	if b.phase != BalootPhaseDeclare {
		return errors.New("not the declaration phase")
	}
	if b.currentPlayerIdx != 0 {
		return errors.New("not your turn to declare")
	}
	return nil
}

// CpuDeclare 手番の CPU が宣言する
func (b *Baloot) CpuDeclare() {
	if b.gameEndFlag || b.phase != BalootPhaseDeclare || b.currentPlayerIdx == 0 {
		return
	}
	idx := b.currentPlayerIdx
	if mode, suit, ok := b.cpuChooseMode(idx); ok {
		b.accept(idx, mode, suit)
		return
	}
	b.players[idx].SetDeclared(true)
	b.appendLog(idx, "pass", "宣言を見送った", nil)
	b.advanceDeclare()
}

// advanceDeclare 次の席へ宣言を回す。
//
// **回ってきた先が CPU の親なら、その場で宣言させる。** 宣言は親の左隣から
// 始まって親で終わるので、全員が見送ると最後は必ず親になる。人間が親のときは
// 代わりに決めず、その席で止めて選ばせる（PassDeclaration が拒否するので、
// 押せるのは Sun か Hokom だけになる）。
func (b *Baloot) advanceDeclare() {
	b.currentPlayerIdx = (b.currentPlayerIdx + 1) % BalootPlayerCnt
	if b.currentPlayerIdx == b.dealerIdx && b.dealerIdx != 0 {
		mode, suit, ok := b.cpuChooseMode(b.dealerIdx)
		if !ok {
			// 親は降りられないので、いちばん長いスートで Hokom を引き受ける。
			mode, suit = BalootModeHokom, b.longestSuit(b.dealerIdx)
		}
		b.accept(b.dealerIdx, mode, suit)
	}
}

// accept モードを確定させ、残りを配ってプレイに入る
func (b *Baloot) accept(idx int, mode BalootMode, suit int) {
	b.mode = mode
	b.declarerIdx = idx
	b.players[idx].SetDeclared(true)
	if mode == BalootModeHokom {
		b.trumpSuit = suit
		b.appendLog(idx, "declare", fmt.Sprintf("Hokom を宣言（切り札 %d）", suit), nil)
	} else {
		b.trumpSuit = 0
		b.appendLog(idx, "declare", "Sun を宣言", nil)
	}
	b.completeDeal()
	b.markBaloot()
	b.sortAllHands()
	b.phase = BalootPhasePlay
	b.leadPlayerIdx = (b.dealerIdx + 1) % BalootPlayerCnt
	b.currentPlayerIdx = b.leadPlayerIdx
}

// completeDeal 残り 3 枚ずつを配る。5×4 + 3×4 = 32 でちょうど配り切る。
func (b *Baloot) completeDeal() {
	for i := range BalootPlayerCnt {
		idx := (b.dealerIdx + 1 + i) % BalootPlayerCnt
		for b.players[idx].GetCardsSize() < BalootHandSize {
			c := b.trumpCards.DrawCard()
			if c == nil {
				break
			}
			b.players[idx].AddCard(c)
		}
	}
}

// markBaloot Hokom のとき、切り札の K+Q を持つプレイヤーに印を付ける
func (b *Baloot) markBaloot() {
	if b.mode != BalootModeHokom {
		return
	}
	for i, p := range b.players {
		k, q := false, false
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c.GetDesign() != b.trumpSuit {
				continue
			}
			switch c.GetValue() {
			case 13:
				k = true
			case 12:
				q = true
			}
		}
		if k && q {
			p.SetHasBaloot(true)
			b.appendLog(i, "baloot", fmt.Sprintf("Baloot（切り札のK+Q）+%d", BalootBonus), nil)
		}
	}
}

// cpuChooseMode CPU の宣言。**Sun は A と 10 の多さ、Hokom は 1 スートの厚みで測る。**
func (b *Baloot) cpuChooseMode(idx int) (BalootMode, int, bool) {
	p := b.players[idx]

	sun := 0
	for i := range p.GetCardsSize() {
		switch p.GetCard(i).GetValue() {
		case 1:
			sun += 3
		case 10:
			sun += 2
		case 13:
			sun++
		}
	}

	bestSuit, bestScore := 0, 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		score := 0
		for i := range p.GetCardsSize() {
			c := p.GetCard(i)
			if c.GetDesign() != suit {
				continue
			}
			switch c.GetValue() {
			case 11: // J
				score += 4
			case 9:
				score += 3
			case 1:
				score += 2
			default:
				score++
			}
		}
		if score > bestScore {
			bestSuit, bestScore = suit, score
		}
	}

	switch {
	case bestScore >= 7:
		return BalootModeHokom, bestSuit, true
	case sun >= 7:
		return BalootModeSun, 0, true
	default:
		return BalootModeNone, 0, false
	}
}

// longestSuit いちばん枚数の多いスート（親が降りられないときの逃げ道）
func (b *Baloot) longestSuit(idx int) int {
	p := b.players[idx]
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

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (b *Baloot) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return errors.New("game has ended")
	}
	if b.phase != BalootPhasePlay {
		return errors.New("not the play phase")
	}
	if b.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return b.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (b *Baloot) CpuPlay() {
	if b.gameEndFlag || b.phase != BalootPhasePlay || b.currentPlayerIdx == 0 {
		return
	}
	_ = b.play(b.currentPlayerIdx, b.chooseCpuCard(b.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (b *Baloot) play(playerIdx, cardIndex int) error {
	p := b.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !b.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	b.currentTrick = append(b.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	b.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(b.currentTrick) < BalootPlayerCnt {
		b.currentPlayerIdx = (playerIdx + 1) % BalootPlayerCnt
		return nil
	}
	b.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか
func (b *Baloot) canPlay(playerIdx int, card *Card) bool {
	if len(b.currentTrick) == 0 {
		return true
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := b.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (b *Baloot) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(b.players) {
		return nil
	}
	p := b.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if b.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決し、カード点を勝者チームに加える
func (b *Baloot) resolveTrick() {
	winner := b.trickWinner()
	cards := make([]*Card, 0, len(b.currentTrick))
	pts := 0
	for _, tc := range b.currentTrick {
		cards = append(cards, tc.Card)
		pts += b.CardPoints(tc.Card)
	}
	b.players[winner].AddTrick(cards)
	b.roundPoints[BalootTeamOf(winner)] += pts

	b.trickNumber++
	b.currentTrick = nil
	b.leadPlayerIdx = winner
	b.currentPlayerIdx = winner

	if b.trickNumber >= BalootTricksPerRound {
		b.roundPoints[BalootTeamOf(winner)] += BalootLastTrickBonus
		b.appendLog(winner, "last", fmt.Sprintf("最終トリック +%d", BalootLastTrickBonus), nil)
		b.finishRound()
	}
}

// CardPoints 札の点数。**モードで表が変わる。**
func (b *Baloot) CardPoints(c *Card) int {
	return BalootCardPoints(c, b.mode, b.trumpSuit)
}

// BalootCardPoints モードごとの札の点数。
//
//	Sun          : A=11, 10=10, K=4, Q=3, J=2, 他=0
//	Hokom の切り札: J=20, 9=14, A=11, 10=10, K=4, Q=3, 他=0
//	Hokom の非切り札: Sun と同じ
func BalootCardPoints(c *Card, mode BalootMode, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if mode == BalootModeHokom && c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11:
			return 20
		case 9:
			return 14
		case 1:
			return 11
		case 10:
			return 10
		case 13:
			return 4
		case 12:
			return 3
		}
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	}
	return 0
}

// finishRound カード点と Baloot ボーナスを足してチーム得点を確定させる
func (b *Baloot) finishRound() {
	for i, p := range b.players {
		if p.GetHasBaloot() {
			b.roundPoints[BalootTeamOf(i)] += BalootBonus
		}
	}
	for team := range BalootTeamCnt {
		b.scores[team] += b.roundPoints[team]
		b.appendLog(-1, "score", fmt.Sprintf("チーム%d に %d 点", team, b.roundPoints[team]), nil)
	}

	if b.scores[0] >= b.config.Target || b.scores[1] >= b.config.Target {
		b.finishGame()
		return
	}
	b.phase = BalootPhaseRoundEnd
}

// NextRound 次のラウンドを開始する
func (b *Baloot) NextRound() {
	if b.gameEndFlag || b.phase != BalootPhaseRoundEnd {
		return
	}
	b.roundNumber++
	b.dealerIdx = (b.dealerIdx + 1) % BalootPlayerCnt
	b.dealRound()
}

// finishGame 目標点に達したチームの勝ち
func (b *Baloot) finishGame() {
	b.phase = BalootPhaseGameEnd
	b.gameEndFlag = true
	switch {
	case b.scores[0] > b.scores[1]:
		b.winnerTeam = 0
	case b.scores[1] > b.scores[0]:
		b.winnerTeam = 1
	default:
		b.winnerTeam = -1
	}
	b.appendLog(-1, "result", fmt.Sprintf("最終得点 %d - %d", b.scores[0], b.scores[1]), nil)
}

// trickWinner 現在のトリックの勝者
func (b *Baloot) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return b.leadPlayerIdx
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	bestIdx, best := b.currentTrick[0].PlayerIdx, b.currentTrick[0].Card
	for _, tc := range b.currentTrick[1:] {
		if b.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// beats challenger が currentBest に勝つか。**Sun には切り札が無い。**
func (b *Baloot) beats(challenger, currentBest *Card, leadSuit int) bool {
	if b.mode == BalootModeHokom {
		cTrump := challenger.GetDesign() == b.trumpSuit
		bTrump := currentBest.GetDesign() == b.trumpSuit
		if cTrump != bTrump {
			return cTrump
		}
	}
	if challenger.GetDesign() != currentBest.GetDesign() {
		// スートが違えば、リードのスートだけが勝ちうる。
		return challenger.GetDesign() == leadSuit
	}
	return b.rankOf(challenger) > b.rankOf(currentBest)
}

// rankOf 札の強さ。**モードで序列が入れ替わる。**
func (b *Baloot) rankOf(c *Card) int {
	if b.mode == BalootModeHokom && c.GetDesign() == b.trumpSuit {
		return balootHokomRank(c)
	}
	return balootSunRank(c)
}

// balootSunRank Sun の序列。A>10>K>Q>J>9>8>7。
func balootSunRank(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // A
		return 8
	case 10:
		return 7
	case 13: // K
		return 6
	case 12: // Q
		return 5
	case 11: // J
		return 4
	case 9:
		return 3
	case 8:
		return 2
	}
	return 1 // 7
}

// balootHokomRank Hokom の切り札の序列。J>9>A>10>K>Q>8>7。
func balootHokomRank(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 11: // J
		return 8
	case 9:
		return 7
	case 1: // A
		return 6
	case 10:
		return 5
	case 13: // K
		return 4
	case 12: // Q
		return 3
	case 8:
		return 2
	}
	return 1 // 7
}

// chooseCpuCard CPU の手。味方が勝っていれば点を乗せ、そうでなければ取りに行く。
func (b *Baloot) chooseCpuCard(playerIdx int) int {
	valid := b.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := b.players[playerIdx]

	if len(b.currentTrick) == 0 {
		bestIdx, bestRank := valid[0], b.rankOf(p.GetCard(valid[0]))
		for _, i := range valid[1:] {
			if r := b.rankOf(p.GetCard(i)); r > bestRank {
				bestIdx, bestRank = i, r
			}
		}
		return bestIdx
	}

	if b.partnerIsWinning(playerIdx) {
		bestIdx, bestPts := -1, -1
		for _, i := range valid {
			c := p.GetCard(i)
			if b.wouldWin(c) {
				continue
			}
			if pts := b.CardPoints(c); pts > bestPts {
				bestIdx, bestPts = i, pts
			}
		}
		if bestIdx >= 0 {
			return bestIdx
		}
	}
	if idx, ok := b.pickWinning(p, valid); ok {
		return idx
	}
	bestIdx, bestPts := valid[0], b.CardPoints(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if pts := b.CardPoints(p.GetCard(i)); pts < bestPts {
			bestIdx, bestPts = i, pts
		}
	}
	return bestIdx
}

// partnerIsWinning 現時点で味方がトリックを取っているか
func (b *Baloot) partnerIsWinning(playerIdx int) bool {
	if len(b.currentTrick) == 0 {
		return false
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	best, bestIdx := b.currentTrick[0].Card, b.currentTrick[0].PlayerIdx
	for _, tc := range b.currentTrick[1:] {
		if b.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx != playerIdx && BalootTeamOf(bestIdx) == BalootTeamOf(playerIdx)
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (b *Baloot) wouldWin(c *Card) bool {
	if c == nil || len(b.currentTrick) == 0 {
		return true
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	best := b.currentTrick[0].Card
	for _, tc := range b.currentTrick[1:] {
		if b.beats(tc.Card, best, leadSuit) {
			best = tc.Card
		}
	}
	return b.beats(c, best, leadSuit)
}

// pickWinning トリックを取れる札のうち一番安いもの
func (b *Baloot) pickWinning(p *BalootPlayer, valid []int) (int, bool) {
	bestIdx, bestPts := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !b.wouldWin(c) {
			continue
		}
		if pts := b.CardPoints(c); bestIdx < 0 || pts < bestPts {
			bestIdx, bestPts = i, pts
		}
	}
	return bestIdx, bestIdx >= 0
}

// BalootHint ヒント情報
type BalootHint struct {
	// CardIndex 推奨する手札のインデックス（宣言中は nil）
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
	// Suit Hokom を勧めるときの切り札スート（それ以外は 0）
	Suit int
}

// GetHint 人間プレイヤーへの推奨手を返す
func (b *Baloot) GetHint() *BalootHint {
	if b.gameEndFlag {
		return nil
	}
	if b.phase == BalootPhaseDeclare && b.currentPlayerIdx == 0 {
		mode, suit, ok := b.cpuChooseMode(0)
		switch {
		case ok && mode == BalootModeHokom:
			return &BalootHint{Reason: "balootDeclareHokom", Suit: suit}
		case ok:
			return &BalootHint{Reason: "balootDeclareSun"}
		default:
			return &BalootHint{Reason: "balootPassDeclare"}
		}
	}
	if !b.IsHumanTurn() || b.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := b.chooseCpuCard(0)
	reason := "balootWinTrick"
	if b.partnerIsWinning(0) {
		reason = "balootFeedPartner"
	}
	return &BalootHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (b *Baloot) GetPhase() BalootPhase { return b.phase }

// GetMode 現在のモード
func (b *Baloot) GetMode() BalootMode { return b.mode }

// GetConfig 現在の設定
func (b *Baloot) GetConfig() BalootConfig { return b.config }

// SetConfig 設定を差し替える
func (b *Baloot) SetConfig(c BalootConfig) { b.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (b *Baloot) GetRoundNumber() int { return b.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (b *Baloot) GetTrickNumber() int { return b.trickNumber }

// GetTrumpSuit 切り札のスート（Sun では 0）
func (b *Baloot) GetTrumpSuit() int { return b.trumpSuit }

// GetDeclarerIdx モードを宣言したプレイヤー (-1: 未決定)
func (b *Baloot) GetDeclarerIdx() int { return b.declarerIdx }

// GetScore チームの累計得点
func (b *Baloot) GetScore(team int) int {
	if team < 0 || team >= BalootTeamCnt {
		return 0
	}
	return b.scores[team]
}

// SetScoreForTestUse チームの累計得点を設定する（復元・テスト用）
func (b *Baloot) SetScoreForTestUse(team, n int) {
	if team >= 0 && team < BalootTeamCnt {
		b.scores[team] = n
	}
}

// GetRoundPoints チームの現ラウンド点
func (b *Baloot) GetRoundPoints(team int) int {
	if team < 0 || team >= BalootTeamCnt {
		return 0
	}
	return b.roundPoints[team]
}

// GetCurrentTrick 現在のトリック
func (b *Baloot) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (b *Baloot) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (b *Baloot) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// GetDealerIdx ディーラー
func (b *Baloot) GetDealerIdx() int { return b.dealerIdx }

// GetPlayerCnt プレイヤー数
func (b *Baloot) GetPlayerCnt() int { return len(b.players) }

// GetPlayer 指定インデックスのプレイヤー
func (b *Baloot) GetPlayer(i int) *BalootPlayer {
	if i < 0 || i >= len(b.players) {
		return nil
	}
	return b.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (b *Baloot) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1: 未確定/同点)
func (b *Baloot) GetWinnerTeam() int { return b.winnerTeam }

// IsHumanTurn 人間の手番か
func (b *Baloot) IsHumanTurn() bool {
	return !b.gameEndFlag && b.phase == BalootPhasePlay && b.currentPlayerIdx == 0
}

// IsHumanDeclareTurn 人間がモードを宣言する番か
func (b *Baloot) IsHumanDeclareTurn() bool {
	return !b.gameEndFlag && b.phase == BalootPhaseDeclare && b.currentPlayerIdx == 0
}

// GiveUp 投了する
func (b *Baloot) GiveUp() {
	if b.gameEndFlag {
		return
	}
	b.phase = BalootPhaseGameEnd
	b.gameEndFlag = true
	b.winnerTeam = 1
	b.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (b *Baloot) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.appendLogAt(b.trickNumber, playerIdx, actionType, detail, cards)
}

// balootJSON is the KV snapshot format for Baloot.
type balootJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*BalootPlayer    `json:"pl"`
	Config           BalootConfig       `json:"cf"`
	Phase            BalootPhase        `json:"ph"`
	Mode             BalootMode         `json:"md"`
	RoundNumber      int                `json:"rn"`
	TrickNumber      int                `json:"tn"`
	TrumpSuit        int                `json:"ts"`
	DeclarerIdx      int                `json:"dc"`
	CurrentTrick     []*TrickCard       `json:"ct"`
	CurrentPlayerIdx int                `json:"cp"`
	LeadPlayerIdx    int                `json:"lp"`
	DealerIdx        int                `json:"di"`
	Scores           [BalootTeamCnt]int `json:"sc"`
	RoundPoints      [BalootTeamCnt]int `json:"rp"`
	GameEndFlag      bool               `json:"ge"`
	WinnerTeam       int                `json:"wt"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (b *Baloot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&balootJSON{
		TrumpCards:       b.trumpCards,
		Players:          b.players,
		Config:           b.config,
		Phase:            b.phase,
		Mode:             b.mode,
		RoundNumber:      b.roundNumber,
		TrickNumber:      b.trickNumber,
		TrumpSuit:        b.trumpSuit,
		DeclarerIdx:      b.declarerIdx,
		CurrentTrick:     b.currentTrick,
		CurrentPlayerIdx: b.currentPlayerIdx,
		LeadPlayerIdx:    b.leadPlayerIdx,
		DealerIdx:        b.dealerIdx,
		Scores:           b.scores,
		RoundPoints:      b.roundPoints,
		GameEndFlag:      b.gameEndFlag,
		WinnerTeam:       b.winnerTeam,
		ActionLog:        b.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。値域を検証する。
func (b *Baloot) UnmarshalJSON(data []byte) error {
	var j balootJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < BalootPhaseDeclare || j.Phase > BalootPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **モードは序列と点数表の両方を選ぶ。** 壊れた値を通すと札の強さが変わる。
	if j.Mode < BalootModeNone || j.Mode > BalootModeHokom {
		return fmt.Errorf("invalid mode: %d", j.Mode)
	}
	if j.TrickNumber < 0 || j.TrickNumber > BalootTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if len(j.ActionLog) > balootMaxSliceLen {
		return errors.New("baloot: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > BalootPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= BalootPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.DeclarerIdx < -1 || j.DeclarerIdx >= BalootPlayerCnt {
		return fmt.Errorf("invalid declarer: %d", j.DeclarerIdx)
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= BalootTeamCnt {
		return fmt.Errorf("invalid winner team: %d", j.WinnerTeam)
	}
	if j.TrumpCards != nil {
		b.trumpCards = j.TrumpCards
	}
	if len(j.Players) == BalootPlayerCnt {
		b.players = j.Players
	}
	b.config = j.Config
	b.phase = j.Phase
	b.mode = j.Mode
	b.roundNumber = j.RoundNumber
	b.trickNumber = j.TrickNumber
	b.trumpSuit = j.TrumpSuit
	b.declarerIdx = j.DeclarerIdx
	b.currentTrick = j.CurrentTrick
	b.currentPlayerIdx = j.CurrentPlayerIdx
	b.leadPlayerIdx = j.LeadPlayerIdx
	b.dealerIdx = j.DealerIdx
	b.scores = j.Scores
	b.roundPoints = j.RoundPoints
	b.gameEndFlag = j.GameEndFlag
	b.winnerTeam = j.WinnerTeam
	b.actionLog = j.ActionLog
	return nil
}
