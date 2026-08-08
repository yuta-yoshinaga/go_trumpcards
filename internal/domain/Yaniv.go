//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// YanivPlayerCnt Yaniv プレイヤー数 (人間 1 + CPU 3)
const YanivPlayerCnt = 4

// YanivHandSize 各プレイヤーの初期手札枚数
const YanivHandSize = 5

// YanivJokerCnt デッキに含めるジョーカー枚数
const YanivJokerCnt = 2

// YanivPhase ゲームフェーズ
type YanivPhase int

// Yaniv のフェーズ定数
const (
	// YanivPhaseDiscard ディスカードフェーズ (手札からカードを捨てる、または Yaniv 宣言)
	YanivPhaseDiscard YanivPhase = 0
	// YanivPhaseDraw ドローフェーズ (山札または直前の捨て札の端から 1 枚引く)
	YanivPhaseDraw YanivPhase = 1
	// YanivPhaseRoundEnd ラウンド終了フェーズ
	YanivPhaseRoundEnd YanivPhase = 2
	// YanivPhaseGameEnd ゲーム終了フェーズ
	YanivPhaseGameEnd YanivPhase = 3
)

// Yaniv Yaniv (ヤニブ) ゲームクラス
type Yaniv struct {
	trumpCards       *TrumpCards
	players          []*YanivPlayer
	config           YanivConfig
	phase            YanivPhase
	currentPlayerIdx int
	drawPile         []*Card // 山札
	pickupCards      []*Card // 直前のプレイヤーが捨てた束 (端のみ引ける)
	pendingDiscard   []*Card // 現プレイヤーが今ターンに捨てた束 (ドロー後に pickupCards となる)
	deadPile         []*Card // 引かれずに流れたカード (山札切れ時に再シャッフル)
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	callerIdx        int   // Yaniv を宣言したプレイヤー (-1 = なし)
	asafWinnerIdx    int   // アサフで宣言者を下回ったプレイヤー (-1 = なし)
	isAsaf           bool  // 直近の宣言がアサフだったか
	roundScores      []int // 直近ラウンドで各プレイヤーが加算された失点
	actionLogBase
	rng *rand.Rand
}

// NewYaniv コンストラクタ
func NewYaniv(trumpCards *TrumpCards, players []*YanivPlayer, config YanivConfig) *Yaniv {
	return &Yaniv{
		trumpCards:    trumpCards,
		players:       players,
		config:        config,
		winnerIdx:     -1,
		callerIdx:     -1,
		asafWinnerIdx: -1,
		roundScores:   make([]int, 0),
		rng:           rand.New(rand.NewSource(rand.Int63())),
	}
}

// SetRand テスト用に乱数源を差し替える
func (g *Yaniv) SetRand(r *rand.Rand) { g.rng = r }

// NewDefaultYaniv returns Yaniv with the standard 4-player setup (1 human, 3 CPU)
// and DefaultYanivConfig. Single source of truth for the CUI, Web, and Worker
// construction sites.
func NewDefaultYaniv() *Yaniv {
	players := []*YanivPlayer{
		NewYanivPlayer(true),
		NewYanivPlayer(false),
		NewYanivPlayer(false),
		NewYanivPlayer(false),
	}
	return NewYaniv(NewTrumpCards(YanivJokerCnt), players, DefaultYanivConfig())
}

// Reset ゲームを初期化する
func (g *Yaniv) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.actionLog = nil

	for _, p := range g.players {
		p.Reset()
		p.SetScore(0)
		p.SetEliminated(false)
	}

	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Yaniv) NextRound() {
	if g.phase != YanivPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.startRound()
}

// startRound 1 ラウンド分のセットアップを行う (脱落者を除いて配牌する)
func (g *Yaniv) startRound() {
	g.pickupCards = nil
	g.pendingDiscard = nil
	g.deadPile = nil
	g.callerIdx = -1
	g.asafWinnerIdx = -1
	g.isAsaf = false
	g.roundScores = make([]int, len(g.players))

	for _, p := range g.players {
		p.Reset()
	}

	g.dealRound()
	g.sortAllHands()

	g.currentPlayerIdx = g.firstActiveIdx()
	g.phase = YanivPhaseDiscard
}

// dealRound 山札を作り直し、脱落していないプレイヤーに手札を配る
func (g *Yaniv) dealRound() {
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

	for range YanivHandSize {
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

	// 最初の 1 枚を場に出し、最初のプレイヤーの引き先 (端) とする
	if len(g.drawPile) > 0 {
		first := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.pickupCards = append(g.pickupCards, first)
	}
}

// --- Human actions ---

// PlayerDiscard 人間プレイヤーが手札からカードの組を捨てる
func (g *Yaniv) PlayerDiscard(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != YanivPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.discard(g.currentPlayerIdx, cardIndices)
}

// PlayerDeclareYaniv 人間プレイヤーが Yaniv を宣言してラウンドを締める
func (g *Yaniv) PlayerDeclareYaniv() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != YanivPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.players[g.currentPlayerIdx].HandTotal() > YanivCallThreshold {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("手札合計が%d点以下でないとYanivを宣言できません", YanivCallThreshold))
	}
	g.resolveYaniv(g.currentPlayerIdx)
	return nil
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Yaniv) PlayerDrawFromStock() error {
	if err := g.guardHumanDraw(); err != nil {
		return err
	}
	g.drawFromStock(g.currentPlayerIdx)
	return nil
}

// PlayerDrawFromPickup 人間プレイヤーが直前の捨て札の端から引く (end: 0=先頭, 1=末尾)
func (g *Yaniv) PlayerDrawFromPickup(end int) error {
	if err := g.guardHumanDraw(); err != nil {
		return err
	}
	if len(g.pickupCards) == 0 {
		return NewDomainError(ErrInvalidPlay, "引ける捨て札がありません")
	}
	g.drawFromPickup(g.currentPlayerIdx, end)
	return nil
}

// guardHumanDraw ドローフェーズの人間操作に共通する前提チェック
func (g *Yaniv) guardHumanDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != YanivPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// --- Shared move helpers ---

// discard カードの組を検証して捨て、ドローフェーズへ移行する
func (g *Yaniv) discard(idx int, cardIndices []int) error {
	p := g.players[idx]
	cards, ok := g.collectComboCards(p, cardIndices)
	if !ok {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if !YanivValidCombo(cards) {
		return NewDomainError(ErrInvalidPlay, "単札・同数の組・同スートの3枚以上の連番のみ捨てられます")
	}

	removed := p.RemoveCards(cardIndices)
	sortCardsForDiscard(removed)
	g.pendingDiscard = removed

	g.appendLog(idx, "discard", fmt.Sprintf("%s discards %s", g.playerName(idx), cardsStr(removed)), removed)
	g.phase = YanivPhaseDraw
	return nil
}

// collectComboCards インデックス列が有効な範囲・重複なしかを確認し、対応カードを返す
func (g *Yaniv) collectComboCards(p *YanivPlayer, cardIndices []int) ([]*Card, bool) {
	if len(cardIndices) == 0 {
		return nil, false
	}
	seen := make(map[int]bool, len(cardIndices))
	cards := make([]*Card, 0, len(cardIndices))
	for _, ci := range cardIndices {
		if ci < 0 || ci >= p.GetCardsSize() || seen[ci] {
			return nil, false
		}
		seen[ci] = true
		cards = append(cards, p.GetCard(ci))
	}
	return cards, true
}

// drawFromStock 山札から 1 枚引いてターンを確定する
func (g *Yaniv) drawFromStock(idx int) {
	if len(g.drawPile) == 0 {
		g.reshuffleDeadPile()
	}
	if len(g.drawPile) == 0 {
		g.endRoundNoContest()
		return
	}
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[idx].AddCard(card)
	g.sortHand(idx)
	g.appendLog(idx, "draw_stock", fmt.Sprintf("%s draws from stock", g.playerName(idx)), nil)
	g.finalizeDraw(-1)
}

// drawFromPickup 直前の捨て札の端 (0=先頭, 1=末尾) から 1 枚引いてターンを確定する
func (g *Yaniv) drawFromPickup(idx, end int) {
	takenIdx := 0
	if end != 0 {
		takenIdx = len(g.pickupCards) - 1
	}
	card := g.pickupCards[takenIdx]
	g.players[idx].AddCard(card)
	g.sortHand(idx)
	g.appendLog(idx, "draw_pickup", fmt.Sprintf("%s takes %s from the discard", g.playerName(idx), cardStr(card)), []*Card{card})
	g.finalizeDraw(takenIdx)
}

// finalizeDraw 引かれなかった捨て札を流し、今ターンの捨て札を次の引き先にしてターンを進める
func (g *Yaniv) finalizeDraw(takenFromPickupIdx int) {
	for i, c := range g.pickupCards {
		if i != takenFromPickupIdx {
			g.deadPile = append(g.deadPile, c)
		}
	}
	g.pickupCards = g.pendingDiscard
	g.pendingDiscard = nil
	g.advanceTurn()
}

// reshuffleDeadPile 流れたカードを山札に戻してシャッフルする
func (g *Yaniv) reshuffleDeadPile() {
	if len(g.deadPile) == 0 {
		return
	}
	g.drawPile = append(g.drawPile, g.deadPile...)
	g.deadPile = nil
	g.rng.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})
}

// advanceTurn 次のアクティブプレイヤーへ移り、ディスカードフェーズに戻す
func (g *Yaniv) advanceTurn() {
	g.currentPlayerIdx = g.nextActiveIdx(g.currentPlayerIdx)
	g.phase = YanivPhaseDiscard
}

// --- CPU ---

// CpuPlay 現在の手番が CPU の場合に 1 アクション実行する
func (g *Yaniv) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.phase != YanivPhaseDiscard && g.phase != YanivPhaseDraw {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	switch g.phase {
	case YanivPhaseDiscard:
		g.cpuDiscard()
	case YanivPhaseDraw:
		g.cpuDraw()
	}
}

// cpuDiscard CPU の Yaniv 宣言 / ディスカード判断
func (g *Yaniv) cpuDiscard() {
	idx := g.currentPlayerIdx
	p := g.players[idx]
	if p.HandTotal() <= YanivCallThreshold && g.cpuShouldCall(p.HandTotal()) {
		g.resolveYaniv(idx)
		return
	}
	cards := g.handCards(idx)
	combo := bestYanivDiscard(cards)
	if err := g.discard(idx, combo); err != nil {
		// Defensive: bestYanivDiscard always yields a valid single-card combo, so
		// the fallback (discarding the first card) can never fail in practice.
		_ = g.discard(idx, []int{0})
	}
}

// cpuShouldCall 難易度に応じた Yaniv 宣言の判断
func (g *Yaniv) cpuShouldCall(handTotal int) bool {
	switch g.config.CpuDifficulty {
	case YanivCpuDifficultyEasy:
		return handTotal <= 2 || g.rng.Intn(2) == 0
	default:
		return true
	}
}

// cpuDraw CPU の引き先判断 (捨て札の端 or 山札)
func (g *Yaniv) cpuDraw() {
	idx := g.currentPlayerIdx
	if end := g.cpuPickupEnd(); end >= 0 {
		g.drawFromPickup(idx, end)
		return
	}
	g.drawFromStock(idx)
}

// cpuPickupEnd 捨て札の端から引くべきか (引く端 0/1、引かないなら -1) を返す
func (g *Yaniv) cpuPickupEnd() int {
	if len(g.pickupCards) == 0 {
		return -1
	}
	if g.config.CpuDifficulty == YanivCpuDifficultyEasy {
		if g.rng.Intn(3) == 0 {
			return 0
		}
		return -1
	}
	threshold := 2
	if g.config.CpuDifficulty == YanivCpuDifficultyHard {
		threshold = 3
	}
	bestEnd := -1
	bestVal := 99
	last := len(g.pickupCards) - 1
	for _, e := range []int{0, last} {
		v := yanivCardValue(g.pickupCards[e])
		if v <= threshold && v < bestVal {
			bestVal = v
			bestEnd = 0
			if e != 0 {
				bestEnd = 1
			}
		}
	}
	return bestEnd
}

// handCards プレイヤーの手札をスライスとして返す
func (g *Yaniv) handCards(idx int) []*Card {
	p := g.players[idx]
	cards := make([]*Card, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// --- Round / game resolution ---

// resolveYaniv Yaniv 宣言を精算する (アサフ判定と失点加算)
func (g *Yaniv) resolveYaniv(callerIdx int) {
	g.callerIdx = callerIdx
	callerTotal := g.players[callerIdx].HandTotal()

	minOpp := -1
	asafWinner := -1
	for i, p := range g.players {
		if i == callerIdx || p.IsEliminated() {
			continue
		}
		t := p.HandTotal()
		if minOpp < 0 || t < minOpp {
			minOpp = t
			asafWinner = i
		}
	}

	g.isAsaf = minOpp >= 0 && minOpp <= callerTotal
	g.asafWinnerIdx = -1
	scores := make([]int, len(g.players))

	if g.isAsaf {
		g.asafWinnerIdx = asafWinner
		scores[callerIdx] = YanivAsafPenalty
		g.appendLog(callerIdx, "asaf", fmt.Sprintf("%s calls Yaniv (%d) but is undercut by %s (%d)! +%d penalty",
			g.playerName(callerIdx), callerTotal, g.playerName(asafWinner), minOpp, YanivAsafPenalty), nil)
		for i, p := range g.players {
			if i == callerIdx || p.IsEliminated() || i == asafWinner {
				continue
			}
			scores[i] = p.HandTotal()
		}
	} else {
		scores[callerIdx] = 0
		g.appendLog(callerIdx, "yaniv", fmt.Sprintf("%s calls Yaniv with %d and wins the round!", g.playerName(callerIdx), callerTotal), nil)
		for i, p := range g.players {
			if i == callerIdx || p.IsEliminated() {
				continue
			}
			scores[i] = p.HandTotal()
		}
	}

	for i := range g.players {
		g.players[i].AddScore(scores[i])
	}
	g.roundScores = scores

	for i, p := range g.players {
		if !p.IsEliminated() && p.GetScore() > g.config.ScoreLimit {
			p.SetEliminated(true)
			g.appendLog(i, "eliminate", fmt.Sprintf("%s is eliminated (score: %d)", g.playerName(i), p.GetScore()), nil)
		}
	}

	g.finishRound()
}

// endRoundNoContest 山札も流し札も尽きた場合の不戦ラウンド終了 (失点なし)
func (g *Yaniv) endRoundNoContest() {
	g.callerIdx = -1
	g.roundScores = make([]int, len(g.players))
	g.appendLog(-1, "round_end", "Round ends with no Yaniv (deck exhausted)", nil)
	g.finishRound()
}

// finishRound ゲーム終了判定を行い、続行ならラウンド終了フェーズへ移る
func (g *Yaniv) finishRound() {
	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = YanivPhaseRoundEnd
	}
}

// checkGameEnd 残りプレイヤーが 1 人以下、または人間が脱落したらゲーム終了
func (g *Yaniv) checkGameEnd() {
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
	g.phase = YanivPhaseGameEnd
	g.winnerIdx = g.leaderIdx()
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", g.playerName(g.winnerIdx)), nil)
}

// leaderIdx 生存者のうち最も失点が少ないプレイヤー (同点は若いインデックス) を返す
func (g *Yaniv) leaderIdx() int {
	best := -1
	bestScore := 0
	for i, p := range g.players {
		if p.IsEliminated() {
			continue
		}
		if best < 0 || p.GetScore() < bestScore {
			best = i
			bestScore = p.GetScore()
		}
	}
	if best >= 0 {
		return best
	}
	// 全員脱落の保険: 失点最小のプレイヤー
	best = 0
	bestScore = g.players[0].GetScore()
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetScore() < bestScore {
			bestScore = g.players[i].GetScore()
			best = i
		}
	}
	return best
}

// firstActiveIdx 0 から見て最初のアクティブプレイヤーのインデックスを返す
func (g *Yaniv) firstActiveIdx() int {
	for i, p := range g.players {
		if !p.IsEliminated() {
			return i
		}
	}
	return 0
}

// nextActiveIdx from の次のアクティブプレイヤーのインデックスを返す
func (g *Yaniv) nextActiveIdx(from int) int {
	n := len(g.players)
	for step := 1; step <= n; step++ {
		idx := (from + step) % n
		if !g.players[idx].IsEliminated() {
			return idx
		}
	}
	return from
}

// humanIdx 人間プレイヤーのインデックスを返す (-1 = 不在)
func (g *Yaniv) humanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// --- Combo validation ---

// YanivValidCombo カードの組が捨て可能か (単札 / 同数の組 / 同スートの3枚以上の連番) を判定する
func YanivValidCombo(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}
	if len(cards) == 1 {
		return true
	}
	if yanivIsSameValueSet(cards) {
		return true
	}
	return yanivIsRun(cards)
}

// yanivIsSameValueSet 全カードが同じ数値 (ジョーカー不可) かを判定する
func yanivIsSameValueSet(cards []*Card) bool {
	v := cards[0].GetValue()
	for _, c := range cards {
		if c.GetDesign() == CardDesignJoker || c.GetValue() != v {
			return false
		}
	}
	return true
}

// yanivIsRun 全カードが同スートかつ連続する数値の3枚以上の連番かを判定する
func yanivIsRun(cards []*Card) bool {
	if len(cards) < 3 {
		return false
	}
	suit := cards[0].GetDesign()
	if suit == CardDesignJoker {
		return false
	}
	values := make([]int, 0, len(cards))
	for _, c := range cards {
		if c.GetDesign() != suit {
			return false
		}
		values = append(values, c.GetValue())
	}
	sort.Ints(values)
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}
	return true
}

// sortCardsForDiscard 捨て札を数値→スート順に並べ替える (引ける端を決定論的にする)
func sortCardsForDiscard(cards []*Card) {
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetValue() != cards[j].GetValue() {
			return cards[i].GetValue() < cards[j].GetValue()
		}
		return cards[i].GetDesign() < cards[j].GetDesign()
	})
}

// bestYanivDiscard 最も失点を多く捨てられる有効な組のインデックス列を返す
func bestYanivDiscard(cards []*Card) []int {
	best := []int{0}
	bestVal := -1
	bestLen := 0

	consider := func(indices []int) {
		picked := make([]*Card, 0, len(indices))
		for _, i := range indices {
			picked = append(picked, cards[i])
		}
		if !YanivValidCombo(picked) {
			return
		}
		val := 0
		for _, c := range picked {
			val += yanivCardValue(c)
		}
		if val > bestVal || (val == bestVal && len(indices) > bestLen) {
			bestVal = val
			bestLen = len(indices)
			best = append([]int{}, indices...)
		}
	}

	// 単札
	for i := range cards {
		consider([]int{i})
	}
	// 同数の組
	byValue := map[int][]int{}
	for i, c := range cards {
		if c.GetDesign() == CardDesignJoker {
			continue
		}
		byValue[c.GetValue()] = append(byValue[c.GetValue()], i)
	}
	for _, idxs := range byValue {
		if len(idxs) >= 2 {
			consider(idxs)
		}
	}
	// 同スートの連番
	bySuit := map[int][]int{}
	for i, c := range cards {
		if c.GetDesign() == CardDesignJoker {
			continue
		}
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], i)
	}
	for _, idxs := range bySuit {
		sort.Slice(idxs, func(a, b int) bool { return cards[idxs[a]].GetValue() < cards[idxs[b]].GetValue() })
		run := []int{}
		for k, i := range idxs {
			if k > 0 && cards[i].GetValue() == cards[idxs[k-1]].GetValue()+1 {
				run = append(run, i)
			} else {
				if len(run) >= 3 {
					consider(run)
				}
				run = []int{i}
			}
		}
		if len(run) >= 3 {
			consider(run)
		}
	}
	return best
}

// --- Getters ---

// GetPhase 現在のフェーズを取得する
func (g *Yaniv) GetPhase() YanivPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)
func (g *Yaniv) SetPhase(phase YanivPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号を取得する
func (g *Yaniv) GetRoundNumber() int { return g.roundNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
func (g *Yaniv) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックスを設定する (テスト用)
func (g *Yaniv) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetPickupCards 直前のプレイヤーが捨てた束 (端のみ引ける) を取得する
func (g *Yaniv) GetPickupCards() []*Card { return g.pickupCards }

// SetPickupCards 引ける捨て札を設定する (テスト用)
func (g *Yaniv) SetPickupCards(pile []*Card) { g.pickupCards = pile }

// GetPendingDiscard 現プレイヤーが今ターンに捨てた束を取得する
func (g *Yaniv) GetPendingDiscard() []*Card { return g.pendingDiscard }

// GetDiscardTop 引ける捨て札の末尾 (空なら nil) を取得する
func (g *Yaniv) GetDiscardTop() *Card {
	if len(g.pickupCards) == 0 {
		return nil
	}
	return g.pickupCards[len(g.pickupCards)-1]
}

// GetDrawPileCount 山札の残り枚数を取得する
func (g *Yaniv) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定する (テスト用)
func (g *Yaniv) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグを取得する
func (g *Yaniv) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックスを取得する (-1 = 未確定)
func (g *Yaniv) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数を取得する
func (g *Yaniv) GetPlayerCnt() int { return len(g.players) }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (g *Yaniv) GetPlayer(i int) *YanivPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間かを返す
func (g *Yaniv) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetCallerIdx Yaniv を宣言したプレイヤーインデックスを取得する (-1 = なし)
func (g *Yaniv) GetCallerIdx() int { return g.callerIdx }

// SetCallerIdx 宣言者を設定する (テスト用)
func (g *Yaniv) SetCallerIdx(idx int) { g.callerIdx = idx }

// GetAsafWinnerIdx アサフで宣言者を下回ったプレイヤーインデックスを取得する (-1 = なし)
func (g *Yaniv) GetAsafWinnerIdx() int { return g.asafWinnerIdx }

// GetIsAsaf 直近の宣言がアサフだったかを取得する
func (g *Yaniv) GetIsAsaf() bool { return g.isAsaf }

// GetRoundScores 直近ラウンドで各プレイヤーが加算された失点を取得する
func (g *Yaniv) GetRoundScores() []int { return g.roundScores }

// GetConfig ゲーム設定を取得する
func (g *Yaniv) GetConfig() YanivConfig { return g.config }

// SetConfig ゲーム設定を設定する
func (g *Yaniv) SetConfig(cfg YanivConfig) { g.config = cfg }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Yaniv) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→数値順にソートする
func (g *Yaniv) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	cards := make([]*Card, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す
func (g *Yaniv) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// cardsStr 複数カードを空白区切りの文字列にする
func cardsStr(cards []*Card) string {
	if len(cards) == 0 {
		return "-"
	}
	s := cardStr(cards[0])
	for _, c := range cards[1:] {
		s += " " + cardStr(c)
	}
	return s
}

// --- JSON ---

// yanivJSON is the JSON wire format for Yaniv.
type yanivJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*YanivPlayer    `json:"pl"`
	Config           YanivConfig       `json:"cf"`
	Phase            YanivPhase        `json:"ph"`
	CurrentPlayerIdx int               `json:"ci"`
	DrawPile         []*Card           `json:"wp"`
	PickupCards      []*Card           `json:"pk"`
	PendingDiscard   []*Card           `json:"pd"`
	DeadPile         []*Card           `json:"dd"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	CallerIdx        int               `json:"kc"`
	AsafWinnerIdx    int               `json:"aw"`
	IsAsaf           bool              `json:"ia"`
	RoundScores      []int             `json:"rs"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// yanivMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const yanivMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (g *Yaniv) MarshalJSON() ([]byte, error) {
	return json.Marshal(yanivJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DrawPile:         g.drawPile,
		PickupCards:      g.pickupCards,
		PendingDiscard:   g.pendingDiscard,
		DeadPile:         g.deadPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		CallerIdx:        g.callerIdx,
		AsafWinnerIdx:    g.asafWinnerIdx,
		IsAsaf:           g.isAsaf,
		RoundScores:      g.roundScores,
		ActionLog:        g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Yaniv) UnmarshalJSON(data []byte) error {
	var j yanivJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > yanivMaxSliceLen || len(j.DrawPile) > yanivMaxSliceLen ||
		len(j.PickupCards) > yanivMaxSliceLen || len(j.PendingDiscard) > yanivMaxSliceLen ||
		len(j.DeadPile) > yanivMaxSliceLen || len(j.RoundScores) > yanivMaxSliceLen ||
		len(j.ActionLog) > yanivMaxSliceLen {
		return fmt.Errorf("yaniv: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(YanivJokerCnt)
	}
	g.players = j.Players
	if len(g.players) != YanivPlayerCnt {
		return fmt.Errorf("yaniv: invalid player count: %d", len(g.players))
	}
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.drawPile = sliceOrEmpty(j.DrawPile)
	g.pickupCards = sliceOrEmpty(j.PickupCards)
	g.pendingDiscard = sliceOrEmpty(j.PendingDiscard)
	g.deadPile = sliceOrEmpty(j.DeadPile)
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.callerIdx = j.CallerIdx
	g.asafWinnerIdx = j.AsafWinnerIdx
	g.isAsaf = j.IsAsaf
	g.roundScores = j.RoundScores
	if g.roundScores == nil {
		g.roundScores = make([]int, 0)
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
