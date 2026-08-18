//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// ThirtyOnePlayerCnt ThirtyOne プレイヤー数 (人間 1 + CPU 3)
const ThirtyOnePlayerCnt = 4

// ThirtyOneHandSize 各プレイヤーの手札枚数
const ThirtyOneHandSize = 3

// CPU ノック閾値 (この点数以上の手札でノックを検討する)。
//
// **公開しているのは画面に出すため。**難易度の違いはこの数字がすべてなのに、
// UI からは Easy/Normal/Hard としか見えず、体験からしか学べなかった (#5623)。
// 説明文に数字を書き写すのではなく、ここを読ませる。
const (
	// ThirtyOneKnockThresholdEasy Easy の CPU がノックを検討する合計値
	ThirtyOneKnockThresholdEasy = 29
	// ThirtyOneKnockThresholdNormal Normal の CPU がノックを検討する合計値
	ThirtyOneKnockThresholdNormal = 27
	// ThirtyOneKnockThresholdHard Hard の CPU がノックを検討する合計値
	ThirtyOneKnockThresholdHard = 25
)

// ThirtyOnePhase ゲームフェーズ
type ThirtyOnePhase int

// ThirtyOne のフェーズ定数
const (
	// ThirtyOnePhaseDraw ドローフェーズ (山札/捨て札から引く、またはノック)
	ThirtyOnePhaseDraw ThirtyOnePhase = 0
	// ThirtyOnePhaseDiscard ディスカードフェーズ (手札から 1 枚捨てる)
	ThirtyOnePhaseDiscard ThirtyOnePhase = 1
	// ThirtyOnePhaseRoundEnd ラウンド終了フェーズ
	ThirtyOnePhaseRoundEnd ThirtyOnePhase = 2
	// ThirtyOnePhaseGameEnd ゲーム終了フェーズ
	ThirtyOnePhaseGameEnd ThirtyOnePhase = 3
)

// ThirtyOne ThirtyOne (サーティワン / Scat) ゲームクラス
type ThirtyOne struct {
	trumpCards       *TrumpCards
	players          []*ThirtyOnePlayer
	config           ThirtyOneConfig
	phase            ThirtyOnePhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	knockerIdx       int   // ノックしたプレイヤー (-1 = 未ノック)
	thirtyOneIdx     int   // 31 を達成したプレイヤー (-1 = なし)
	roundWinnerIdx   int   // 直近ラウンドの勝者 (-1 = 未確定)
	roundLosers      []int // 直近ラウンドでライフを失ったプレイヤー
	actionLogBase
	rng *rand.Rand
}

// NewThirtyOne コンストラクタ
func NewThirtyOne(trumpCards *TrumpCards, players []*ThirtyOnePlayer, config ThirtyOneConfig) *ThirtyOne {
	return &ThirtyOne{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		knockerIdx:     -1,
		thirtyOneIdx:   -1,
		roundWinnerIdx: -1,
		roundLosers:    make([]int, 0),
		rng:            rand.New(rand.NewSource(rand.Int63())),
	}
}

// SetRand テスト用に乱数源を差し替える
func (g *ThirtyOne) SetRand(r *rand.Rand) { g.rng = r }

// NewDefaultThirtyOne returns ThirtyOne with the standard 4-player setup
// (1 human, 3 CPU) and DefaultThirtyOneConfig. Single source of truth for the
// CUI, Web, and Worker construction sites.
func NewDefaultThirtyOne() *ThirtyOne {
	players := []*ThirtyOnePlayer{
		NewThirtyOnePlayer(true),
		NewThirtyOnePlayer(false),
		NewThirtyOnePlayer(false),
		NewThirtyOnePlayer(false),
	}
	return NewThirtyOne(NewTrumpCards(0), players, DefaultThirtyOneConfig())
}

// Reset ゲームを初期化する
func (g *ThirtyOne) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.actionLog = nil

	for _, p := range g.players {
		p.Reset()
		p.SetLives(g.config.InitialLives)
	}

	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *ThirtyOne) NextRound() {
	if g.phase != ThirtyOnePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.startRound()
}

// startRound 1 ラウンド分のセットアップを行う (脱落者を除いて配牌する)
func (g *ThirtyOne) startRound() {
	g.discardPile = nil
	g.drawPile = nil
	g.knockerIdx = -1
	g.thirtyOneIdx = -1
	g.roundWinnerIdx = -1
	g.roundLosers = make([]int, 0)

	for _, p := range g.players {
		p.Reset()
	}

	g.dealRound()

	g.currentPlayerIdx = g.firstActiveIdx()
	g.phase = ThirtyOnePhaseDraw
	g.checkBlitzOnDeal()
}

// dealRound 山札を作り直し、脱落していないプレイヤーに手札を配る
func (g *ThirtyOne) dealRound() {
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

	for range ThirtyOneHandSize {
		for _, p := range g.players {
			if p.IsEliminated() {
				continue
			}
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				p.AddCard(card)
			}
		}
	}

	if len(g.drawPile) > 0 {
		first := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, first)
	}
}

// checkBlitzOnDeal 配牌時にちょうど 31 を持つプレイヤーがいれば即ラウンド勝利
func (g *ThirtyOne) checkBlitzOnDeal() {
	for i, p := range g.players {
		if p.IsEliminated() {
			continue
		}
		if p.BestSuitScore() == ThirtyOneTarget {
			g.appendLog(i, "blitz", fmt.Sprintf("%s is dealt 31!", playerName(g.players, i)), nil)
			g.declareThirtyOne(i)
			return
		}
	}
}

// --- Human actions ---

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *ThirtyOne) PlayerDrawFromStock() error {
	if err := g.guardHumanDraw(); err != nil {
		return err
	}
	if len(g.drawPile) == 0 {
		g.endRound("stock empty")
		return nil
	}
	g.drawFromStock(g.currentPlayerIdx)
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札からカードを引く
func (g *ThirtyOne) PlayerDrawFromDiscard() error {
	if err := g.guardHumanDraw(); err != nil {
		return err
	}
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札がありません")
	}
	g.drawFromDiscard(g.currentPlayerIdx)
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *ThirtyOne) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ThirtyOnePhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	p := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	g.discardAndResolve(g.currentPlayerIdx, cardIndex)
	return nil
}

// PlayerKnock 人間プレイヤーがノックする (手札を引かずに勝負を締める)
func (g *ThirtyOne) PlayerKnock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ThirtyOnePhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.knockerIdx >= 0 {
		return NewDomainError(ErrInvalidPlay, "既にノックされています")
	}
	g.knock(g.currentPlayerIdx)
	return nil
}

// guardHumanDraw ドローフェーズの人間操作に共通する前提チェック
func (g *ThirtyOne) guardHumanDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ThirtyOnePhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// --- Shared move helpers ---

// drawFromStock 山札からカードを引いてディスカードフェーズへ移行する
func (g *ThirtyOne) drawFromStock(idx int) {
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[idx].AddCard(card)
	g.appendLog(idx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, idx)), nil)
	g.phase = ThirtyOnePhaseDiscard
}

// drawFromDiscard 捨て札の一番上を引いてディスカードフェーズへ移行する
func (g *ThirtyOne) drawFromDiscard(idx int) {
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[idx].AddCard(card)
	g.appendLog(idx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, idx), cardStr(card)), []*Card{card})
	g.phase = ThirtyOnePhaseDiscard
}

// discardAndResolve カードを捨て、31 達成判定の上でターンを進める
func (g *ThirtyOne) discardAndResolve(idx, cardIndex int) {
	discarded := g.players[idx].RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(idx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, idx), cardStr(discarded)), []*Card{discarded})

	if g.players[idx].BestSuitScore() == ThirtyOneTarget {
		g.appendLog(idx, "thirty_one", fmt.Sprintf("%s reaches 31!", playerName(g.players, idx)), nil)
		g.declareThirtyOne(idx)
		return
	}
	g.advanceTurn()
}

// knock ノックを記録してターンを進める
func (g *ThirtyOne) knock(idx int) {
	g.knockerIdx = idx
	g.appendLog(idx, "knock", fmt.Sprintf("%s knocks (score: %d)", playerName(g.players, idx), g.players[idx].BestSuitScore()), nil)
	g.advanceTurn()
}

// --- CPU ---

// CpuPlay 現在の手番が CPU の場合に 1 アクション実行する
func (g *ThirtyOne) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.phase != ThirtyOnePhaseDraw && g.phase != ThirtyOnePhaseDiscard {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	switch g.phase {
	case ThirtyOnePhaseDraw:
		g.cpuDraw()
	case ThirtyOnePhaseDiscard:
		g.cpuDiscard()
	}
}

// cpuDraw CPU のドロー/ノック判断
func (g *ThirtyOne) cpuDraw() {
	idx := g.currentPlayerIdx
	p := g.players[idx]

	if g.knockerIdx < 0 && p.BestSuitScore() >= g.cpuKnockThreshold() {
		g.knock(idx)
		return
	}

	if len(g.discardPile) > 0 && g.cpuWantsDiscard(p) {
		g.drawFromDiscard(idx)
		return
	}

	if len(g.drawPile) == 0 {
		g.endRound("stock empty")
		return
	}
	g.drawFromStock(idx)
}

// ThirtyOneHint は人間プレイヤーへの推奨手。
type ThirtyOneHint struct {
	// Action は "draw_stock" / "draw_discard" / "discard" / "knock" のいずれか。
	Action string
	// CardIndex は discard のときの推奨インデックス (それ以外は -1)。
	CardIndex int
	// Reason は i18n キーの末尾。
	Reason string
}

// GetHint は現在の局面での推奨手を返す (人間の手番でなければ nil)。
//
// **CPU と同じ材料で判断する。**ドローは cpuWantsDiscard、捨て札は bestDropIndex、
// ノックは難易度ごとの閾値をそのまま使う。別の計算を書くと、CPU には有利と見える
// 手を人間には勧めない、という食い違いが出る (#4806)。
func (g *ThirtyOne) GetHint() *ThirtyOneHint {
	if g.gameEndFlag || !g.players[g.currentPlayerIdx].GetIsHuman() {
		return nil
	}
	p := g.players[g.currentPlayerIdx]
	switch g.phase {
	case ThirtyOnePhaseDraw:
		// ノックできる点数に届いているなら、まずそれを勧める。
		if g.knockerIdx < 0 && p.BestSuitScore() >= g.cpuKnockThreshold() {
			return &ThirtyOneHint{Action: "knock", CardIndex: -1, Reason: "knock_ready"}
		}
		if len(g.discardPile) > 0 && g.cpuWantsDiscard(p) {
			return &ThirtyOneHint{Action: "draw_discard", CardIndex: -1, Reason: "discard_improves"}
		}
		return &ThirtyOneHint{Action: "draw_stock", CardIndex: -1, Reason: "draw_stock"}
	case ThirtyOnePhaseDiscard:
		cards := g.handCards(g.currentPlayerIdx)
		if len(cards) == 0 {
			return nil
		}
		return &ThirtyOneHint{Action: "discard", CardIndex: bestDropIndex(cards), Reason: "drop_weakest"}
	}
	return nil
}

// cpuDiscard CPU が最も得点に貢献しないカードを捨てる
func (g *ThirtyOne) cpuDiscard() {
	idx := g.currentPlayerIdx
	cards := g.handCards(idx)
	if len(cards) == 0 {
		// Defensive: only reachable from a corrupt deserialized state where a
		// player in the discard phase has no cards. Avoid an Intn(0) panic.
		g.advanceTurn()
		return
	}

	dropIdx := 0
	if g.config.CpuDifficulty == ThirtyOneCpuDifficultyEasy {
		dropIdx = g.rng.Intn(len(cards))
	} else {
		dropIdx = bestDropIndex(cards)
	}
	g.discardAndResolve(idx, dropIdx)
}

// cpuWantsDiscard 捨て札を引くと最高スート得点が改善するかを判定する
func (g *ThirtyOne) cpuWantsDiscard(p *ThirtyOnePlayer) bool {
	if g.config.CpuDifficulty == ThirtyOneCpuDifficultyEasy {
		return g.rng.Intn(2) == 0
	}
	top := g.discardPile[len(g.discardPile)-1]
	current := p.BestSuitScore()
	cards := make([]*Card, 0, p.GetCardsSize()+1)
	for i := range p.GetCardsSize() {
		cards = append(cards, p.GetCard(i))
	}
	cards = append(cards, top)
	return bestScoreAfterDrop(cards) > current
}

// cpuKnockThreshold 難易度ごとのノック閾値
func (g *ThirtyOne) cpuKnockThreshold() int {
	switch g.config.CpuDifficulty {
	case ThirtyOneCpuDifficultyEasy:
		return ThirtyOneKnockThresholdEasy
	case ThirtyOneCpuDifficultyHard:
		return ThirtyOneKnockThresholdHard
	default:
		return ThirtyOneKnockThresholdNormal
	}
}

// handCards プレイヤーの手札をスライスとして返す
func (g *ThirtyOne) handCards(idx int) []*Card {
	p := g.players[idx]
	cards := make([]*Card, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// thirtyOneScoreOf 任意のカード集合の最高スート得点を返す
func thirtyOneScoreOf(cards []*Card) int {
	scores := map[int]int{}
	for _, c := range cards {
		scores[c.GetDesign()] += thirtyOneCardScore(c)
	}
	best := 0
	for _, s := range scores {
		if s > best {
			best = s
		}
	}
	return best
}

// bestScoreAfterDrop 1 枚捨てた後に得られる最高スート得点を返す
func bestScoreAfterDrop(cards []*Card) int {
	best := 0
	for i := range cards {
		remaining := make([]*Card, 0, len(cards)-1)
		for j, c := range cards {
			if j != i {
				remaining = append(remaining, c)
			}
		}
		if s := thirtyOneScoreOf(remaining); s > best {
			best = s
		}
	}
	return best
}

// bestDropIndex 捨てると残りの最高スート得点が最大になるカードのインデックスを返す
func bestDropIndex(cards []*Card) int {
	bestIdx := 0
	bestScore := -1
	for i := range cards {
		remaining := make([]*Card, 0, len(cards)-1)
		for j, c := range cards {
			if j != i {
				remaining = append(remaining, c)
			}
		}
		if s := thirtyOneScoreOf(remaining); s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx
}

// --- Round / game resolution ---

// declareThirtyOne 31 達成者を勝者とし、他のアクティブプレイヤーがライフを失う
func (g *ThirtyOne) declareThirtyOne(idx int) {
	g.thirtyOneIdx = idx
	g.roundWinnerIdx = idx
	g.roundLosers = make([]int, 0)
	for i, p := range g.players {
		if i == idx || p.IsEliminated() {
			continue
		}
		p.LoseLife()
		g.roundLosers = append(g.roundLosers, i)
	}
	g.appendLog(idx, "round_win", fmt.Sprintf("%s wins the round with 31", playerName(g.players, idx)), nil)
	g.finishRound()
}

// endRound ノック後または山札切れでラウンドを精算し、最低点のプレイヤーがライフを失う
func (g *ThirtyOne) endRound(reason string) {
	if reason != "" {
		g.appendLog(-1, "round_end", fmt.Sprintf("Round ends (%s)", reason), nil)
	}

	minScore := -1
	maxScore := -1
	for i, p := range g.players {
		if p.IsEliminated() {
			continue
		}
		s := p.BestSuitScore()
		if minScore < 0 || s < minScore {
			minScore = s
		}
		if s > maxScore {
			maxScore = s
			g.roundWinnerIdx = i
		}
	}

	g.roundLosers = make([]int, 0)
	for i, p := range g.players {
		if p.IsEliminated() {
			continue
		}
		if p.BestSuitScore() == minScore {
			p.LoseLife()
			g.roundLosers = append(g.roundLosers, i)
			g.appendLog(i, "lose_life", fmt.Sprintf("%s loses a life (score: %d)", playerName(g.players, i), minScore), nil)
		}
	}
	g.finishRound()
}

// finishRound ゲーム終了判定を行い、続行ならラウンド終了フェーズへ移る
func (g *ThirtyOne) finishRound() {
	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = ThirtyOnePhaseRoundEnd
	}
}

// checkGameEnd 残りプレイヤーが 1 人以下、または人間が脱落したらゲーム終了
func (g *ThirtyOne) checkGameEnd() {
	activeCount := 0
	for _, p := range g.players {
		if !p.IsEliminated() {
			activeCount++
		}
	}
	humanIdx := g.humanIdx()
	humanOut := humanIdx >= 0 && g.players[humanIdx].IsEliminated()

	if activeCount > 1 && !humanOut {
		return
	}

	g.gameEndFlag = true
	g.phase = ThirtyOnePhaseGameEnd
	g.winnerIdx = g.leaderIdx()
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// leaderIdx 最もライフが多いプレイヤー (同点は若いインデックス) を返す
func (g *ThirtyOne) leaderIdx() int {
	best := 0
	bestLives := g.players[0].GetLives()
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetLives() > bestLives {
			bestLives = g.players[i].GetLives()
			best = i
		}
	}
	return best
}

// advanceTurn 次のアクティブプレイヤーへ。ノック後に一巡したらラウンド精算
func (g *ThirtyOne) advanceTurn() {
	next := g.nextActiveIdx(g.currentPlayerIdx)
	if g.knockerIdx >= 0 && next == g.knockerIdx {
		g.endRound("")
		return
	}
	g.currentPlayerIdx = next
	g.phase = ThirtyOnePhaseDraw
}

// firstActiveIdx 0 から見て最初のアクティブプレイヤーのインデックスを返す
func (g *ThirtyOne) firstActiveIdx() int {
	for i, p := range g.players {
		if !p.IsEliminated() {
			return i
		}
	}
	return 0
}

// nextActiveIdx from の次のアクティブプレイヤーのインデックスを返す
func (g *ThirtyOne) nextActiveIdx(from int) int {
	return nextIndexWhere(g.players, from, func(p *ThirtyOnePlayer) bool { return !p.IsEliminated() })
}

// humanIdx 人間プレイヤーのインデックスを返す (-1 = 不在)
func (g *ThirtyOne) humanIdx() int {
	return findHumanIdx(g.players)
}

// --- Getters ---

// GetPhase 現在のフェーズを取得する
// GetCpuKnockThreshold は現在の難易度で CPU がノックを検討する合計値を返す。
func (g *ThirtyOne) GetCpuKnockThreshold() int { return g.cpuKnockThreshold() }

func (g *ThirtyOne) GetPhase() ThirtyOnePhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)
func (g *ThirtyOne) SetPhase(phase ThirtyOnePhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号を取得する
func (g *ThirtyOne) GetRoundNumber() int { return g.roundNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
func (g *ThirtyOne) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックスを設定する (テスト用)
func (g *ThirtyOne) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardTop 捨て札の一番上を取得する (空なら nil)
func (g *ThirtyOne) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// SetDiscardPile 捨て札を設定する (テスト用)
func (g *ThirtyOne) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDrawPileCount 山札の残り枚数を取得する
func (g *ThirtyOne) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定する (テスト用)
func (g *ThirtyOne) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグを取得する
func (g *ThirtyOne) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックスを取得する (-1 = 未確定)
func (g *ThirtyOne) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数を取得する
func (g *ThirtyOne) GetPlayerCnt() int { return len(g.players) }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (g *ThirtyOne) GetPlayer(i int) *ThirtyOnePlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かを返す
func (g *ThirtyOne) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetKnockerIdx ノックしたプレイヤーインデックスを取得する (-1 = 未ノック)
func (g *ThirtyOne) GetKnockerIdx() int { return g.knockerIdx }

// SetKnockerIdx ノッカーを設定する (テスト用)
func (g *ThirtyOne) SetKnockerIdx(idx int) { g.knockerIdx = idx }

// GetThirtyOneIdx 31 を達成したプレイヤーインデックスを取得する (-1 = なし)
func (g *ThirtyOne) GetThirtyOneIdx() int { return g.thirtyOneIdx }

// GetRoundWinnerIdx 直近ラウンドの勝者インデックスを取得する (-1 = 未確定)
func (g *ThirtyOne) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetRoundLosers 直近ラウンドでライフを失ったプレイヤーのインデックス一覧を取得する
func (g *ThirtyOne) GetRoundLosers() []int { return g.roundLosers }

// GetConfig ゲーム設定を取得する
func (g *ThirtyOne) GetConfig() ThirtyOneConfig { return g.config }

// SetConfig ゲーム設定を設定する
func (g *ThirtyOne) SetConfig(cfg ThirtyOneConfig) { g.config = cfg }

// --- JSON ---

// thirtyOneJSON is the JSON wire format for ThirtyOne.
type thirtyOneJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*ThirtyOnePlayer `json:"pl"`
	Config           ThirtyOneConfig    `json:"cf"`
	Phase            ThirtyOnePhase     `json:"ph"`
	CurrentPlayerIdx int                `json:"ci"`
	DiscardPile      []*Card            `json:"dp"`
	DrawPile         []*Card            `json:"wp"`
	GameEndFlag      bool               `json:"ge"`
	WinnerIdx        int                `json:"wi"`
	RoundNumber      int                `json:"rn"`
	KnockerIdx       int                `json:"ki"`
	ThirtyOneIdx     int                `json:"oi"`
	RoundWinnerIdx   int                `json:"rw"`
	RoundLosers      []int              `json:"rl"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// thirtyOneMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const thirtyOneMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (g *ThirtyOne) MarshalJSON() ([]byte, error) {
	return json.Marshal(thirtyOneJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		KnockerIdx:       g.knockerIdx,
		ThirtyOneIdx:     g.thirtyOneIdx,
		RoundWinnerIdx:   g.roundWinnerIdx,
		RoundLosers:      g.roundLosers,
		ActionLog:        g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *ThirtyOne) UnmarshalJSON(data []byte) error {
	var j thirtyOneJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > thirtyOneMaxSliceLen || len(j.DiscardPile) > thirtyOneMaxSliceLen ||
		len(j.DrawPile) > thirtyOneMaxSliceLen || len(j.ActionLog) > thirtyOneMaxSliceLen ||
		len(j.RoundLosers) > thirtyOneMaxSliceLen {
		return fmt.Errorf("thirtyone: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if len(g.players) != ThirtyOnePlayerCnt {
		return fmt.Errorf("thirtyone: invalid player count: %d", len(g.players))
	}
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.knockerIdx = j.KnockerIdx
	g.thirtyOneIdx = j.ThirtyOneIdx
	g.roundWinnerIdx = j.RoundWinnerIdx
	g.roundLosers = j.RoundLosers
	if g.roundLosers == nil {
		g.roundLosers = make([]int, 0)
	}
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
