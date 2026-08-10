package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// SixCardGolfPhase ゲームフェーズ
type SixCardGolfPhase int

// SixCardGolfのフェーズ定数
const (
	// SixCardGolfPhaseSetup 初期カード2枚めくりフェーズ
	SixCardGolfPhaseSetup SixCardGolfPhase = iota
	// SixCardGolfPhasePlayerTurn ドロー元選択待ち
	SixCardGolfPhasePlayerTurn
	// SixCardGolfPhaseDrawPending ドロー済み・交換/捨て選択待ち
	SixCardGolfPhaseDrawPending
	// SixCardGolfPhaseRoundOver ラウンド終了
	SixCardGolfPhaseRoundOver
	// SixCardGolfPhaseGameOver ゲーム終了
	SixCardGolfPhaseGameOver
)

// SixCardGolfGridSize 1プレイヤーのグリッドサイズ
const SixCardGolfGridSize = 6

// SixCardGolfDefaultRounds デフォルトラウンド数
const SixCardGolfDefaultRounds = 9

// SixCardGolfPlayerMin 最小プレイヤー数
const SixCardGolfPlayerMin = 2

// SixCardGolfPlayerMax 最大プレイヤー数
const SixCardGolfPlayerMax = 4

// SixCardGolfInitialFlips セットアップ時にめくるカード枚数
const SixCardGolfInitialFlips = 2

// SixCardGolfSlot グリッド上の1スロット
type SixCardGolfSlot struct {
	Card   *Card
	FaceUp bool
}

// SixCardGolfPlayer プレイヤー状態
type SixCardGolfPlayer struct {
	Grid            [SixCardGolfGridSize]SixCardGolfSlot
	IsCpu           bool
	RoundScore      int
	CumulativeScore int
}

// AllFaceUp 全カードが表向きか
func (p *SixCardGolfPlayer) AllFaceUp() bool {
	for _, s := range p.Grid {
		if !s.FaceUp {
			return false
		}
	}
	return true
}

// FaceUpCount 表向きカード枚数
func (p *SixCardGolfPlayer) FaceUpCount() int {
	cnt := 0
	for _, s := range p.Grid {
		if s.FaceUp {
			cnt++
		}
	}
	return cnt
}

// SixCardGolfCpuDifficulty CPU難易度
type SixCardGolfCpuDifficulty int

// CPU難易度定数
const (
	// SixCardGolfCpuEasy 低難易度
	SixCardGolfCpuEasy SixCardGolfCpuDifficulty = iota
	// SixCardGolfCpuNormal 中難易度
	SixCardGolfCpuNormal
	// SixCardGolfCpuHard 高難易度
	SixCardGolfCpuHard
)

// SixCardGolfConfig ゲーム設定
type SixCardGolfConfig struct {
	PlayerCount   int                      `json:"pc"`
	CpuDifficulty SixCardGolfCpuDifficulty `json:"cd"`
	Rounds        int                      `json:"rn"`
}

// DefaultSixCardGolfConfig デフォルト設定
func DefaultSixCardGolfConfig() SixCardGolfConfig {
	return SixCardGolfConfig{
		PlayerCount:   2,
		CpuDifficulty: SixCardGolfCpuNormal,
		Rounds:        SixCardGolfDefaultRounds,
	}
}

// Validate 設定バリデーション
func (c SixCardGolfConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, SixCardGolfPlayerMin, SixCardGolfPlayerMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SixCardGolfCpuEasy), int(SixCardGolfCpuHard)); err != nil {
		return err
	}
	if err := ValidateMin("rounds", c.Rounds, 1); err != nil {
		return err
	}
	return nil
}

// SixCardGolf シックスカードゴルフゲーム
type SixCardGolf struct {
	trumpCards       *TrumpCards
	players          []*SixCardGolfPlayer
	config           SixCardGolfConfig
	phase            SixCardGolfPhase
	currentPlayerIdx int
	drawPile         []*Card
	discardPile      []*Card
	drawnCard        *Card
	drawnFromDiscard bool
	canFlip          bool
	roundNumber      int
	finalTurnTrigger int
	finalTurnDone    []bool
	gameEndFlag      bool
	winnerIdx        int
	actionLog        []*ActionLogEntry
	rng              *rand.Rand
}

// NewSixCardGolf コンストラクタ
func NewSixCardGolf(trumpCards *TrumpCards, config SixCardGolfConfig) *SixCardGolf {
	players := make([]*SixCardGolfPlayer, config.PlayerCount)
	players[0] = &SixCardGolfPlayer{IsCpu: false}
	for i := 1; i < config.PlayerCount; i++ {
		players[i] = &SixCardGolfPlayer{IsCpu: true}
	}
	return &SixCardGolf{
		trumpCards:       trumpCards,
		players:          players,
		config:           config,
		winnerIdx:        -1,
		finalTurnTrigger: -1,
		rng:              rand.New(rand.NewSource(rand.Int63())),
	}
}

// NewDefaultSixCardGolf 標準2人対戦を返す
func NewDefaultSixCardGolf() *SixCardGolf {
	return NewSixCardGolf(NewTrumpCards(0), DefaultSixCardGolfConfig())
}

// SetRand テスト用乱数源差し替え
func (g *SixCardGolf) SetRand(r *rand.Rand) { g.rng = r }

// Reset ゲーム初期化
func (g *SixCardGolf) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.actionLog = nil

	pc := g.config.PlayerCount
	if len(g.players) != pc {
		g.players = make([]*SixCardGolfPlayer, pc)
		g.players[0] = &SixCardGolfPlayer{IsCpu: false}
		for i := 1; i < pc; i++ {
			g.players[i] = &SixCardGolfPlayer{IsCpu: true}
		}
	}
	for _, p := range g.players {
		p.RoundScore = 0
		p.CumulativeScore = 0
	}

	g.startRound()
}

// NextRound 次のラウンドを開始
func (g *SixCardGolf) NextRound() {
	if g.phase != SixCardGolfPhaseRoundOver {
		return
	}
	g.roundNumber++
	g.startRound()
}

// startRound ラウンド開始共通処理
func (g *SixCardGolf) startRound() {
	g.trumpCards.Replenish()
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}
	g.rng.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	for _, p := range g.players {
		p.RoundScore = 0
		for j := range p.Grid {
			if len(g.drawPile) > 0 {
				c := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				p.Grid[j] = SixCardGolfSlot{Card: c, FaceUp: false}
			}
		}
	}

	g.discardPile = nil
	if len(g.drawPile) > 0 {
		top := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, top)
	}

	g.drawnCard = nil
	g.drawnFromDiscard = false
	g.canFlip = false
	g.currentPlayerIdx = 0
	g.finalTurnTrigger = -1
	g.finalTurnDone = make([]bool, len(g.players))
	g.phase = SixCardGolfPhaseSetup
}

// FlipInitial セットアップ時にカードをめくる
func (g *SixCardGolf) FlipInitial(pos int) error {
	if g.phase != SixCardGolfPhaseSetup {
		return ErrWrongPhase
	}
	p := g.players[g.currentPlayerIdx]
	if pos < 0 || pos >= SixCardGolfGridSize {
		return NewDomainError(ErrInvalidCard, "位置が範囲外です")
	}
	if p.Grid[pos].FaceUp {
		return NewDomainError(ErrInvalidPlay, "既に表向きです")
	}

	p.Grid[pos].FaceUp = true
	g.appendLog("flipInitial", fmt.Sprintf("プレイヤー%dが位置%dをめくった", g.currentPlayerIdx, pos), []*Card{p.Grid[pos].Card})

	if p.FaceUpCount() >= SixCardGolfInitialFlips {
		g.currentPlayerIdx++
		if g.currentPlayerIdx >= len(g.players) {
			g.currentPlayerIdx = 0
			g.phase = SixCardGolfPhasePlayerTurn
		}
	}
	return nil
}

// DrawStock 山札から引く
func (g *SixCardGolf) DrawStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SixCardGolfPhasePlayerTurn {
		return ErrWrongPhase
	}
	if g.players[g.currentPlayerIdx].IsCpu {
		return ErrNotHumanTurn
	}
	if len(g.drawPile) == 0 {
		g.refillDrawPile()
	}
	if len(g.drawPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "山札がありません")
	}

	g.drawnCard = g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.drawnFromDiscard = false
	g.phase = SixCardGolfPhaseDrawPending
	g.appendLog("drawStock", fmt.Sprintf("プレイヤー%dが山札から引いた", g.currentPlayerIdx), []*Card{g.drawnCard})
	return nil
}

// DrawDiscard 捨て札から引く
func (g *SixCardGolf) DrawDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SixCardGolfPhasePlayerTurn {
		return ErrWrongPhase
	}
	if g.players[g.currentPlayerIdx].IsCpu {
		return ErrNotHumanTurn
	}
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札がありません")
	}

	g.drawnCard = g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.drawnFromDiscard = true
	g.phase = SixCardGolfPhaseDrawPending
	g.appendLog("drawDiscard", fmt.Sprintf("プレイヤー%dが捨て札から引いた", g.currentPlayerIdx), []*Card{g.drawnCard})
	return nil
}

// SwapCard 引いたカードとグリッド位置を交換
func (g *SixCardGolf) SwapCard(pos int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SixCardGolfPhaseDrawPending {
		return ErrWrongPhase
	}
	if g.players[g.currentPlayerIdx].IsCpu {
		return ErrNotHumanTurn
	}
	if pos < 0 || pos >= SixCardGolfGridSize {
		return NewDomainError(ErrInvalidCard, "位置が範囲外です")
	}

	p := g.players[g.currentPlayerIdx]
	old := p.Grid[pos].Card
	p.Grid[pos].Card = g.drawnCard
	p.Grid[pos].FaceUp = true
	g.discardPile = append(g.discardPile, old)
	g.appendLog("swap", fmt.Sprintf("プレイヤー%dが位置%dを交換", g.currentPlayerIdx, pos), []*Card{g.drawnCard, old})

	g.drawnCard = nil
	g.canFlip = false
	g.advanceTurn()
	return nil
}

// DiscardDrawn 引いたカードを捨てる
func (g *SixCardGolf) DiscardDrawn() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SixCardGolfPhaseDrawPending {
		return ErrWrongPhase
	}
	if g.players[g.currentPlayerIdx].IsCpu {
		return ErrNotHumanTurn
	}

	g.discardPile = append(g.discardPile, g.drawnCard)
	g.appendLog("discard", fmt.Sprintf("プレイヤー%dが引いたカードを捨てた", g.currentPlayerIdx), []*Card{g.drawnCard})
	g.drawnCard = nil

	if !g.drawnFromDiscard {
		p := g.players[g.currentPlayerIdx]
		hasFaceDown := false
		for _, s := range p.Grid {
			if !s.FaceUp {
				hasFaceDown = true
				break
			}
		}
		if hasFaceDown {
			g.canFlip = true
			g.phase = SixCardGolfPhasePlayerTurn
			return nil
		}
	}

	g.canFlip = false
	g.advanceTurn()
	return nil
}

// FlipCard 捨て後に伏せ札をめくる
func (g *SixCardGolf) FlipCard(pos int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if !g.canFlip {
		return NewDomainError(ErrInvalidPlay, "めくれません")
	}
	if g.players[g.currentPlayerIdx].IsCpu {
		return ErrNotHumanTurn
	}
	if pos < 0 || pos >= SixCardGolfGridSize {
		return NewDomainError(ErrInvalidCard, "位置が範囲外です")
	}

	p := g.players[g.currentPlayerIdx]
	if p.Grid[pos].FaceUp {
		return NewDomainError(ErrInvalidPlay, "既に表向きです")
	}

	p.Grid[pos].FaceUp = true
	g.appendLog("flip", fmt.Sprintf("プレイヤー%dが位置%dをめくった", g.currentPlayerIdx, pos), []*Card{p.Grid[pos].Card})

	g.canFlip = false
	g.advanceTurn()
	return nil
}

// SkipFlip めくりをスキップしてターンを進める
func (g *SixCardGolf) SkipFlip() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if !g.canFlip {
		return NewDomainError(ErrInvalidPlay, "スキップ不可")
	}
	g.canFlip = false
	g.advanceTurn()
	return nil
}

// CpuPlay CPUの1ターン実行
func (g *SixCardGolf) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	p := g.players[g.currentPlayerIdx]
	if !p.IsCpu {
		return
	}

	switch g.phase {
	case SixCardGolfPhaseSetup:
		g.cpuSetup()
	case SixCardGolfPhasePlayerTurn:
		if g.canFlip {
			g.cpuFlipAfterDiscard()
			return
		}
		g.cpuDraw()
		if g.phase == SixCardGolfPhaseDrawPending {
			g.cpuSwapOrDiscard()
		}
	case SixCardGolfPhaseDrawPending:
		g.cpuSwapOrDiscard()
	}
}

// cpuSetup CPUのセットアップフェーズ処理
func (g *SixCardGolf) cpuSetup() {
	p := g.players[g.currentPlayerIdx]
	needed := SixCardGolfInitialFlips - p.FaceUpCount()
	flipped := 0
	for i := range p.Grid {
		if flipped >= needed {
			break
		}
		if !p.Grid[i].FaceUp {
			p.Grid[i].FaceUp = true
			g.appendLog("flipInitial", fmt.Sprintf("CPU%dが位置%dをめくった", g.currentPlayerIdx, i), []*Card{p.Grid[i].Card})
			flipped++
		}
	}
	g.currentPlayerIdx++
	if g.currentPlayerIdx >= len(g.players) {
		g.currentPlayerIdx = 0
		g.phase = SixCardGolfPhasePlayerTurn
	}
}

// cpuDraw CPUのドロー判断
func (g *SixCardGolf) cpuDraw() {
	discardTop := g.GetDiscardTop()
	if discardTop != nil && g.sixCardGolfCardScore(discardTop) <= 3 {
		g.drawnCard = g.discardPile[len(g.discardPile)-1]
		g.discardPile = g.discardPile[:len(g.discardPile)-1]
		g.drawnFromDiscard = true
		g.appendLog("drawDiscard", fmt.Sprintf("CPU%dが捨て札から引いた", g.currentPlayerIdx), []*Card{g.drawnCard})
	} else {
		if len(g.drawPile) == 0 {
			g.refillDrawPile()
		}
		if len(g.drawPile) == 0 {
			return
		}
		g.drawnCard = g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.drawnFromDiscard = false
		g.appendLog("drawStock", fmt.Sprintf("CPU%dが山札から引いた", g.currentPlayerIdx), []*Card{g.drawnCard})
	}
	g.phase = SixCardGolfPhaseDrawPending
}

// cpuSwapOrDiscard CPUの交換/捨て判断
func (g *SixCardGolf) cpuSwapOrDiscard() {
	p := g.players[g.currentPlayerIdx]
	drawnScore := g.sixCardGolfCardScore(g.drawnCard)

	bestPos := -1
	bestGain := 0

	for i := 0; i < SixCardGolfGridSize; i++ {
		s := p.Grid[i]
		if s.FaceUp {
			currentScore := g.sixCardGolfCardScore(s.Card)
			gain := currentScore - drawnScore
			if g.wouldColumnMatch(g.currentPlayerIdx, i, g.drawnCard) {
				gain = currentScore + drawnScore
			}
			if gain > bestGain {
				bestGain = gain
				bestPos = i
			}
		} else {
			if drawnScore <= 3 {
				gain := 5 - drawnScore
				if g.wouldColumnMatch(g.currentPlayerIdx, i, g.drawnCard) {
					gain = 10
				}
				if gain > bestGain {
					bestGain = gain
					bestPos = i
				}
			}
		}
	}

	if bestPos >= 0 {
		old := p.Grid[bestPos].Card
		p.Grid[bestPos].Card = g.drawnCard
		p.Grid[bestPos].FaceUp = true
		g.discardPile = append(g.discardPile, old)
		g.appendLog("swap", fmt.Sprintf("CPU%dが位置%dを交換", g.currentPlayerIdx, bestPos), []*Card{g.drawnCard, old})
		g.drawnCard = nil
		g.canFlip = false
		g.advanceTurn()
	} else {
		g.discardPile = append(g.discardPile, g.drawnCard)
		g.appendLog("discard", fmt.Sprintf("CPU%dが引いたカードを捨てた", g.currentPlayerIdx), []*Card{g.drawnCard})
		g.drawnCard = nil
		if !g.drawnFromDiscard {
			g.canFlip = true
			g.cpuFlipAfterDiscard()
		} else {
			g.canFlip = false
			g.advanceTurn()
		}
	}
}

// cpuFlipAfterDiscard CPUの捨て後めくり
func (g *SixCardGolf) cpuFlipAfterDiscard() {
	p := g.players[g.currentPlayerIdx]
	for i := 0; i < SixCardGolfGridSize; i++ {
		if !p.Grid[i].FaceUp {
			p.Grid[i].FaceUp = true
			g.appendLog("flip", fmt.Sprintf("CPU%dが位置%dをめくった", g.currentPlayerIdx, i), []*Card{p.Grid[i].Card})
			break
		}
	}
	g.canFlip = false
	g.advanceTurn()
}

// wouldColumnMatch 指定位置にカードを置いた場合に列一致するか
func (g *SixCardGolf) wouldColumnMatch(playerIdx, pos int, card *Card) bool {
	if card == nil {
		return false
	}
	col := pos % 3
	pairPos := col
	if pos < 3 {
		pairPos = col + 3
	}
	p := g.players[playerIdx]
	if p == nil {
		return false
	}
	if !p.Grid[pairPos].FaceUp || p.Grid[pairPos].Card == nil {
		return false
	}
	return p.Grid[pairPos].Card.GetValue() == card.GetValue()
}

// advanceTurn ターン進行
func (g *SixCardGolf) advanceTurn() {
	p := g.players[g.currentPlayerIdx]
	if p.AllFaceUp() && g.finalTurnTrigger < 0 {
		g.finalTurnTrigger = g.currentPlayerIdx
		g.appendLog("trigger", fmt.Sprintf("プレイヤー%dが全カードを公開！最終ターン開始", g.currentPlayerIdx), nil)
	}

	if g.finalTurnTrigger >= 0 {
		g.finalTurnDone[g.currentPlayerIdx] = true
	}

	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.drawnFromDiscard = false
	g.canFlip = false

	if g.finalTurnTrigger >= 0 {
		allDone := true
		for i, done := range g.finalTurnDone {
			if i != g.finalTurnTrigger && !done {
				allDone = false
				break
			}
		}
		if allDone {
			g.scoreRound()
			return
		}
		if g.currentPlayerIdx == g.finalTurnTrigger {
			g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
		}
	}

	g.phase = SixCardGolfPhasePlayerTurn
}

// scoreRound ラウンドスコア計算
func (g *SixCardGolf) scoreRound() {
	for _, p := range g.players {
		for j := range p.Grid {
			p.Grid[j].FaceUp = true
		}
	}

	for i, p := range g.players {
		p.RoundScore = g.ScorePlayer(i)
		p.CumulativeScore += p.RoundScore
	}

	if g.roundNumber >= g.config.Rounds {
		g.gameEndFlag = true
		g.winnerIdx = g.findWinner()
		g.phase = SixCardGolfPhaseGameOver
		g.appendLog("gameOver", fmt.Sprintf("ゲーム終了！プレイヤー%dの勝利", g.winnerIdx), nil)
	} else {
		g.phase = SixCardGolfPhaseRoundOver
		g.appendLog("roundOver", fmt.Sprintf("ラウンド%d終了", g.roundNumber), nil)
	}
}

// ScorePlayer プレイヤーのスコア計算（列一致0点ルール適用）
func (g *SixCardGolf) ScorePlayer(playerIdx int) int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return 0
	}
	p := g.players[playerIdx]
	total := 0
	scored := [SixCardGolfGridSize]bool{}

	for col := 0; col < 3; col++ {
		top := col
		bot := col + 3
		if p.Grid[top].FaceUp && p.Grid[bot].FaceUp &&
			p.Grid[top].Card != nil && p.Grid[bot].Card != nil &&
			p.Grid[top].Card.GetValue() == p.Grid[bot].Card.GetValue() {
			scored[top] = true
			scored[bot] = true
		}
	}

	for i := 0; i < SixCardGolfGridSize; i++ {
		if scored[i] {
			continue
		}
		if p.Grid[i].FaceUp {
			total += g.sixCardGolfCardScore(p.Grid[i].Card)
		}
	}
	return total
}

// ShouldDrawFromDiscard は捨て札トップを引くべきか (スコア <= 3 の低いカード) を返す。
// CPU の cpuDraw と同じ基準で、PlayerTurn の CUI ヒントが利用する。
func (g *SixCardGolf) ShouldDrawFromDiscard() bool {
	top := g.GetDiscardTop()
	return top != nil && g.sixCardGolfCardScore(top) <= 3
}

// RecommendedSwap は引いたカードを交換すべき最良のグリッド位置と、それが列ペアを
// 完成させるかを返す。交換より捨てが良ければ pos = -1。DrawPending 以外・引いた
// カードなしでも pos = -1。CPU の cpuSwapOrDiscard と同じ評価ロジックを用いる。
func (g *SixCardGolf) RecommendedSwap() (pos int, formsPair bool) {
	if g.phase != SixCardGolfPhaseDrawPending || g.drawnCard == nil {
		return -1, false
	}
	p := g.players[g.currentPlayerIdx]
	if p == nil {
		return -1, false
	}
	drawnScore := g.sixCardGolfCardScore(g.drawnCard)
	bestPos, bestGain, bestPair := -1, 0, false
	for i := 0; i < SixCardGolfGridSize; i++ {
		s := p.Grid[i]
		pair := g.wouldColumnMatch(g.currentPlayerIdx, i, g.drawnCard)
		var gain int
		if s.FaceUp {
			gain = g.sixCardGolfCardScore(s.Card) - drawnScore
			if pair {
				gain = g.sixCardGolfCardScore(s.Card) + drawnScore
			}
		} else if drawnScore <= 3 {
			gain = 5 - drawnScore
			if pair {
				gain = 10
			}
		} else {
			continue
		}
		if gain > bestGain {
			bestGain, bestPos, bestPair = gain, i, pair
		}
	}
	return bestPos, bestPair
}

// sixCardGolfCardScore カード1枚のスコア
func (g *SixCardGolf) sixCardGolfCardScore(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	switch v {
	case 13:
		return 0
	case 1:
		return 1
	case 11, 12:
		return 10
	default:
		return v
	}
}

// findWinner 累積スコア最低プレイヤーを返す
func (g *SixCardGolf) findWinner() int {
	best := -1
	bestScore := 1<<31 - 1
	for i, p := range g.players {
		if p.CumulativeScore < bestScore {
			bestScore = p.CumulativeScore
			best = i
		}
	}
	return best
}

// refillDrawPile 山札補充（捨て札トップ以外をシャッフル）
func (g *SixCardGolf) refillDrawPile() {
	if len(g.discardPile) <= 1 {
		return
	}
	top := g.discardPile[len(g.discardPile)-1]
	rest := make([]*Card, len(g.discardPile)-1)
	copy(rest, g.discardPile[:len(g.discardPile)-1])
	g.rng.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	g.drawPile = append(g.drawPile, rest...)
	g.discardPile = []*Card{top}
}

// appendLog 棋譜追加
func (g *SixCardGolf) appendLog(actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.roundNumber,
		PlayerIdx:  g.currentPlayerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetPhase フェーズ取得
func (g *SixCardGolf) GetPhase() SixCardGolfPhase { return g.phase }

// SetPhase テスト用
func (g *SixCardGolf) SetPhase(ph SixCardGolfPhase) { g.phase = ph }

// GetConfig 設定取得
func (g *SixCardGolf) GetConfig() SixCardGolfConfig { return g.config }

// SetConfig 設定変更
func (g *SixCardGolf) SetConfig(cfg SixCardGolfConfig) { g.config = cfg }

// GetGameEndFlag ゲーム終了フラグ
func (g *SixCardGolf) GetGameEndFlag() bool { return g.gameEndFlag }

// IsHumanTurn 人間ターンか
func (g *SixCardGolf) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return !g.players[g.currentPlayerIdx].IsCpu
}

// GetRoundNumber ラウンド番号
func (g *SixCardGolf) GetRoundNumber() int { return g.roundNumber }

// GetCurrentPlayerIdx 現在プレイヤー
func (g *SixCardGolf) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx テスト用
func (g *SixCardGolf) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardTop 捨て札トップ
func (g *SixCardGolf) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札枚数
func (g *SixCardGolf) GetDrawPileCount() int { return len(g.drawPile) }

// GetWinnerIdx 勝者
func (g *SixCardGolf) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *SixCardGolf) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *SixCardGolf) GetPlayer(i int) *SixCardGolfPlayer {
	return getPlayer(g.players, i)
}

// GetDrawnCard 引いたカード
func (g *SixCardGolf) GetDrawnCard() *Card { return g.drawnCard }

// SetDrawnCard テスト用
func (g *SixCardGolf) SetDrawnCard(c *Card) { g.drawnCard = c }

// GetDrawnFromDiscard 捨て札から引いたか
func (g *SixCardGolf) GetDrawnFromDiscard() bool { return g.drawnFromDiscard }

// SetDrawnFromDiscard テスト用
func (g *SixCardGolf) SetDrawnFromDiscard(v bool) { g.drawnFromDiscard = v }

// GetCanFlip めくり可能か
func (g *SixCardGolf) GetCanFlip() bool { return g.canFlip }

// SetCanFlip テスト用
func (g *SixCardGolf) SetCanFlip(v bool) { g.canFlip = v }

// GetFinalTurnTrigger 最終ターントリガー
func (g *SixCardGolf) GetFinalTurnTrigger() int { return g.finalTurnTrigger }

// SetFinalTurnTrigger テスト用
func (g *SixCardGolf) SetFinalTurnTrigger(idx int) { g.finalTurnTrigger = idx }

// GetFinalTurnDone 最終ターン完了状態
func (g *SixCardGolf) GetFinalTurnDone() []bool {
	out := make([]bool, len(g.finalTurnDone))
	copy(out, g.finalTurnDone)
	return out
}

// SetFinalTurnDone テスト用
func (g *SixCardGolf) SetFinalTurnDone(done []bool) {
	g.finalTurnDone = make([]bool, len(done))
	copy(g.finalTurnDone, done)
}

// GetActionLog 棋譜取得
func (g *SixCardGolf) GetActionLog() []*ActionLogEntry { return g.actionLog }

// SetDrawPile テスト用
func (g *SixCardGolf) SetDrawPile(cards []*Card) {
	g.drawPile = append(g.drawPile[:0], cards...)
}

// SetDiscardPile テスト用
func (g *SixCardGolf) SetDiscardPile(cards []*Card) {
	g.discardPile = append(g.discardPile[:0], cards...)
}

// SetPlayerGrid テスト用
func (g *SixCardGolf) SetPlayerGrid(idx int, grid [SixCardGolfGridSize]SixCardGolfSlot) {
	if idx >= 0 && idx < len(g.players) {
		g.players[idx].Grid = grid
	}
}

// --- JSON ---

type sixCardGolfJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*sixCardGolfPlayerJS `json:"pl"`
	Config           SixCardGolfConfig      `json:"cf"`
	Phase            SixCardGolfPhase       `json:"ph"`
	CurrentPlayerIdx int                    `json:"ci"`
	DrawPile         []*Card                `json:"dp"`
	DiscardPile      []*Card                `json:"di"`
	DrawnCard        *Card                  `json:"dc"`
	DrawnFromDiscard bool                   `json:"df"`
	CanFlip          bool                   `json:"fl"`
	RoundNumber      int                    `json:"rn"`
	FinalTurnTrigger int                    `json:"ft"`
	FinalTurnDone    []bool                 `json:"fd"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerIdx        int                    `json:"wi"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

type sixCardGolfPlayerJS struct {
	Grid            [SixCardGolfGridSize]sixCardGolfSlotJS `json:"gr"`
	IsCpu           bool                                   `json:"ic"`
	RoundScore      int                                    `json:"rs"`
	CumulativeScore int                                    `json:"cs"`
}

type sixCardGolfSlotJS struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

const sixCardGolfMaxSliceLen = 200

// MarshalJSON implements json.Marshaler.
func (g *SixCardGolf) MarshalJSON() ([]byte, error) {
	j := sixCardGolfJSON{
		TrumpCards:       g.trumpCards,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DrawPile:         g.drawPile,
		DiscardPile:      g.discardPile,
		DrawnCard:        g.drawnCard,
		DrawnFromDiscard: g.drawnFromDiscard,
		CanFlip:          g.canFlip,
		RoundNumber:      g.roundNumber,
		FinalTurnTrigger: g.finalTurnTrigger,
		FinalTurnDone:    g.finalTurnDone,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	}
	j.Players = make([]*sixCardGolfPlayerJS, len(g.players))
	for i, p := range g.players {
		pj := &sixCardGolfPlayerJS{
			IsCpu:           p.IsCpu,
			RoundScore:      p.RoundScore,
			CumulativeScore: p.CumulativeScore,
		}
		for k, s := range p.Grid {
			pj.Grid[k] = sixCardGolfSlotJS(s)
		}
		j.Players[i] = pj
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *SixCardGolf) UnmarshalJSON(data []byte) error {
	var j sixCardGolfJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.DrawPile) > sixCardGolfMaxSliceLen || len(j.DiscardPile) > sixCardGolfMaxSliceLen || len(j.ActionLog) > sixCardGolfMaxSliceLen {
		return fmt.Errorf("sixcardgolf: input array exceeds maximum allowed size")
	}
	if len(j.Players) < SixCardGolfPlayerMin || len(j.Players) > SixCardGolfPlayerMax {
		return fmt.Errorf("sixcardgolf: invalid player count")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawnCard = j.DrawnCard
	g.drawnFromDiscard = j.DrawnFromDiscard
	g.canFlip = j.CanFlip
	g.roundNumber = j.RoundNumber
	g.finalTurnTrigger = j.FinalTurnTrigger
	g.finalTurnDone = j.FinalTurnDone
	if g.finalTurnDone == nil || len(g.finalTurnDone) != len(j.Players) {
		g.finalTurnDone = make([]bool, len(j.Players))
	}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	if g.phase != SixCardGolfPhaseGameOver {
		g.winnerIdx = -1
	}
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}

	g.players = make([]*SixCardGolfPlayer, len(j.Players))
	for i, pj := range j.Players {
		if pj == nil {
			return fmt.Errorf("sixcardgolf: player data cannot be null")
		}
		p := &SixCardGolfPlayer{
			IsCpu:           pj.IsCpu,
			RoundScore:      pj.RoundScore,
			CumulativeScore: pj.CumulativeScore,
		}
		for k, s := range pj.Grid {
			p.Grid[k] = SixCardGolfSlot(s)
		}
		g.players[i] = p
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
