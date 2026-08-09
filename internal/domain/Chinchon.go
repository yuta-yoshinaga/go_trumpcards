//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// ChinchonHandSize 初期配布枚数 (各プレイヤー7枚)
const ChinchonHandSize = 7

// ChinchonPhase ゲームフェーズ
type ChinchonPhase int

// Chinchonのフェーズ定数
const (
	// ChinchonPhaseDraw ドローフェーズ (山札または捨て札から引く)
	ChinchonPhaseDraw ChinchonPhase = 0
	// ChinchonPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or ノック)
	ChinchonPhaseDiscard ChinchonPhase = 1
	// ChinchonPhaseLayoff レイオフフェーズ (他プレイヤーがノッカーのメルドに付ける)
	ChinchonPhaseLayoff ChinchonPhase = 2
	// ChinchonPhaseRoundEnd ラウンド終了フェーズ
	ChinchonPhaseRoundEnd ChinchonPhase = 3
	// ChinchonPhaseGameEnd ゲーム終了フェーズ
	ChinchonPhaseGameEnd ChinchonPhase = 4
)

// chinchonRankPosition は 8・9・10 を除いた40枚デッキでのランクの「隣接位置」を返す。
//
// Chinchón は 8/9/10 を抜いた40枚のラテンデッキを使うため、ラン (連続) の判定では
// A,2,3,4,5,6,7,J,Q,K を連続したランクとして扱う。すなわち 7 と J は隣接する。
// 戻り値: A=1,2=2,...,7=7,J=8,Q=9,K=10。デッキに存在しないランク(8/9/10/Joker)は 0。
func chinchonRankPosition(value int) int {
	switch {
	case value >= 1 && value <= 7: // A..7
		return value
	case value == 11: // J
		return 8
	case value == 12: // Q
		return 9
	case value == 13: // K
		return 10
	default: // 8,9,10,Joker など (デッキには存在しない)
		return 0
	}
}

// chinchonDeck は 8・9・10 を取り除いた40枚のデッキを構築する。
// 標準52枚デッキ (NewTrumpCards(0)) から rank 8/9/10 のカードを除外する。
// Conquian の newConquianDeck と同等だが、ゲーム間の結合を避けるため独自に実装する。
func chinchonDeck() []*Card {
	tc := NewTrumpCards(0)
	deck := make([]*Card, 0, 40)
	for {
		c := tc.DrawCard()
		if c == nil {
			break
		}
		v := c.GetValue()
		if v == 8 || v == 9 || v == 10 {
			continue
		}
		deck = append(deck, c)
	}
	return deck
}

// hasChinchon は手札がチンチョン (同一スートで連続する7枚) かを判定する。
//
// 7枚すべてが同じスートで、chinchonRankPosition が連続していればチンチョン。
// 例: ♠A-2-3-4-5-6-7 や ♥4-5-6-7-J-Q-K (7とJは隣接)。
func hasChinchon(cards []*Card) bool {
	if len(cards) != ChinchonHandSize {
		return false
	}
	suit := cards[0].GetDesign()
	positions := make([]int, 0, len(cards))
	for _, c := range cards {
		if c.GetDesign() != suit {
			return false
		}
		pos := chinchonRankPosition(c.GetValue())
		if pos == 0 {
			return false
		}
		positions = append(positions, pos)
	}
	sort.Ints(positions)
	for i := 1; i < len(positions); i++ {
		if positions[i] != positions[i-1]+1 {
			return false
		}
	}
	return true
}

// Chinchon チンチョンゲームクラス
type Chinchon struct {
	players          []*ChinchonPlayer
	config           ChinchonConfig
	phase            ChinchonPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int // マッチ勝者 (-1 = 未確定)
	roundNumber      int
	actionLogBase
	knockerIdx      int       // ノックしたプレイヤー (-1 = ノックなし)
	knockerMelds    [][]*Card // ノッカーのメルド (レイオフ用)
	knockerDeadwood []*Card   // ノッカーのデッドウッド
	layoffQueue     []int     // 残りのレイオフ対象プレイヤー (順番)
}

// NewChinchon コンストラクタ
func NewChinchon(players []*ChinchonPlayer, config ChinchonConfig) *Chinchon {
	return &Chinchon{
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		knockerIdx:  -1,
	}
}

// NewDefaultChinchon は DefaultChinchonConfig (4人: 人間1 + CPU3) で Chinchon を生成する。
// CUI・Web・Worker の構築サイトの単一情報源。
func NewDefaultChinchon() *Chinchon {
	return newChinchonForConfig(DefaultChinchonConfig())
}

// newChinchonForConfig は config.PlayerCount に応じた人数 (seat 0 が人間、残りが CPU) を構築する。
func newChinchonForConfig(cfg ChinchonConfig) *Chinchon {
	n := cfg.PlayerCount
	if n < 2 || n > 4 {
		n = 4
		cfg.PlayerCount = 4
	}
	players := make([]*ChinchonPlayer, 0, n)
	players = append(players, NewChinchonPlayer(true))
	for i := 1; i < n; i++ {
		players = append(players, NewChinchonPlayer(false))
	}
	return NewChinchon(players, cfg)
}

// Reset ゲーム初期化
func (g *Chinchon) Reset() {
	// プレイヤー数が設定と一致しなければ作り直す (設定変更経由の Reset 対応)。
	if len(g.players) != g.config.PlayerCount {
		rebuilt := newChinchonForConfig(g.config)
		g.players = rebuilt.players
		g.config = rebuilt.config
	}

	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.resetRoundState()

	for _, p := range g.players {
		p.SetCumulativeScore(0)
		p.SetEliminated(false)
		p.ResetRound()
	}

	g.dealInitialCards()
	g.phase = ChinchonPhaseDraw
	g.checkChinchonForAll()
}

// NextRound 次のラウンドを開始する
func (g *Chinchon) NextRound() {
	if g.phase != ChinchonPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.resetRoundState()
	g.currentPlayerIdx = g.firstActivePlayer()

	for _, p := range g.players {
		p.ResetRound()
	}

	g.dealInitialCards()
	g.phase = ChinchonPhaseDraw
	g.checkChinchonForAll()
}

// resetRoundState はラウンド固有のゲーム状態をクリアする。
func (g *Chinchon) resetRoundState() {
	g.discardPile = nil
	g.drawPile = nil
	g.knockerIdx = -1
	g.knockerMelds = nil
	g.knockerDeadwood = nil
	g.layoffQueue = nil
}

// dealInitialCards 初期配布: アクティブな各プレイヤーに7枚、1枚を捨て札に、残りを山札へ。
func (g *Chinchon) dealInitialCards() {
	deck := chinchonDeck()
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	g.discardPile = make([]*Card, 0)
	g.drawPile = make([]*Card, 0, len(deck))

	idx := 0
	for i := 0; i < ChinchonHandSize; i++ {
		for j := range g.players {
			if g.players[j].GetEliminated() {
				continue
			}
			if idx < len(deck) {
				g.players[j].AddCard(deck[idx])
				idx++
			}
		}
	}
	// 1枚を捨て札に
	if idx < len(deck) {
		g.discardPile = append(g.discardPile, deck[idx])
		idx++
	}
	// 残りを山札へ
	for ; idx < len(deck); idx++ {
		g.drawPile = append(g.drawPile, deck[idx])
	}
	g.sortAllHands()
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Chinchon) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ChinchonPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return nil
	}

	g.doDrawStock(g.currentPlayerIdx)
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札の一番上を引く
func (g *Chinchon) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ChinchonPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札がありません")
	}

	g.doDrawDiscard(g.currentPlayerIdx)
	return nil
}

// doDrawStock 山札の一番上を idx の手札に加え、ディスカードフェーズへ移る。
func (g *Chinchon) doDrawStock(idx int) {
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[idx].AddCard(card)
	g.sortHand(idx)
	g.appendLog(idx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, idx)), nil)
	g.phase = ChinchonPhaseDiscard
}

// doDrawDiscard 捨て札の一番上を idx の手札に加え、ディスカードフェーズへ移る。
func (g *Chinchon) doDrawDiscard(idx int) {
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[idx].AddCard(card)
	g.sortHand(idx)
	g.appendLog(idx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, idx), cardStr(card)), []*Card{card})
	g.phase = ChinchonPhaseDiscard
}

// PlayerDiscard 人間プレイヤーがカードを捨ててターンを終了する。
func (g *Chinchon) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ChinchonPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	// 捨てた後に7枚同スート連続が残ればチンチョン (即時ゲーム勝利)。
	if g.checkChinchon(g.currentPlayerIdx) {
		return nil
	}
	g.advanceTurn()
	return nil
}

// PlayerKnock 人間プレイヤーがノックする (1枚捨ててからノック)。
//
// 捨てた後の手札 (6枚) をメルド分割し、デッドウッドが KnockThreshold 以下のときのみ
// ノックできる。デッドウッド0なら「ジン」相当 (レイオフなしで即スコア)。
func (g *Chinchon) PlayerKnock(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ChinchonPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	testCards := cardsExcludingIndex(player, cardIndex)
	_, deadwood := chinchonFindBestMelds(testCards)
	deadwoodValue := CalcDeadwoodValue(deadwood)
	if deadwoodValue > g.config.KnockThreshold {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("デッドウッドが%d点以下でないとノックできません（現在%d点）", g.config.KnockThreshold, deadwoodValue))
	}

	g.executeKnock(g.currentPlayerIdx, cardIndex)
	return nil
}

// executeKnock はノックを実行し、レイオフフェーズへ移行する (デッドウッド0なら即スコア)。
func (g *Chinchon) executeKnock(idx, cardIndex int) {
	player := g.players[idx]
	testCards := cardsExcludingIndex(player, cardIndex)
	melds, deadwood := chinchonFindBestMelds(testCards)
	deadwoodValue := CalcDeadwoodValue(deadwood)

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	g.knockerIdx = idx
	g.knockerMelds = melds
	g.knockerDeadwood = deadwood
	g.appendLog(idx, "knock", fmt.Sprintf("%s knocks (deadwood: %d)", playerName(g.players, idx), deadwoodValue), []*Card{discarded})

	// レイオフ対象はノッカー以外のアクティブプレイヤー (席順)。
	g.layoffQueue = g.layoffQueue[:0]
	for i := 0; i < len(g.players); i++ {
		seat := (idx + 1 + i) % len(g.players)
		if seat == idx || g.players[seat].GetEliminated() {
			continue
		}
		g.layoffQueue = append(g.layoffQueue, seat)
	}

	g.phase = ChinchonPhaseLayoff
	g.advanceLayoff()
}

// PlayerLayoff 人間プレイヤーがレイオフするカードを選択する (空 indices でレイオフ終了)。
func (g *Chinchon) PlayerLayoff(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ChinchonPhaseLayoff {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]

	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
	}
	seen := make(map[int]bool)
	for _, idx := range cardIndices {
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
		}
		seen[idx] = true
	}
	for _, idx := range cardIndices {
		if !g.canLayoff(player.GetCard(idx)) {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%sはレイオフできません", cardStr(player.GetCard(idx))))
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(cardIndices)))
	for _, idx := range cardIndices {
		card := player.GetCard(idx)
		g.layoffCard(card)
		player.RemoveCard(idx)
		g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	}

	g.advanceLayoff()
	return nil
}

// advanceLayoff は次のレイオフ対象プレイヤーに手番を渡す。
// 残りがいなければスコアリングへ進む。
func (g *Chinchon) advanceLayoff() {
	if len(g.layoffQueue) == 0 {
		g.scoreRound()
		return
	}
	g.currentPlayerIdx = g.layoffQueue[0]
	g.layoffQueue = g.layoffQueue[1:]
}

// canLayoff カードがノッカーのメルドにレイオフ可能か
func (g *Chinchon) canLayoff(card *Card) bool {
	for _, meld := range g.knockerMelds {
		if chinchonCanAddToMeld(meld, card) {
			return true
		}
	}
	return false
}

// layoffCard カードをノッカーのメルドに追加する
func (g *Chinchon) layoffCard(card *Card) {
	for i, meld := range g.knockerMelds {
		if chinchonCanAddToMeld(meld, card) {
			g.knockerMelds[i] = append(meld, card)
			return
		}
	}
}

// scoreRound ラウンドのスコアを確定する。
//
// 各プレイヤーは残りデッドウッドを累積点に加算する (ノッカーは自身のデッドウッドのみ)。
// 累積点が EliminationLimit を超えたプレイヤーは脱落。残り1人になればマッチ終了。
func (g *Chinchon) scoreRound() {
	for i, p := range g.players {
		if p.GetEliminated() {
			continue
		}
		var deadwoodValue int
		if i == g.knockerIdx {
			deadwoodValue = CalcDeadwoodValue(g.knockerDeadwood)
		} else {
			cards := handCards(p)
			_, dw := chinchonFindBestMelds(cards)
			deadwoodValue = CalcDeadwoodValue(dw)
		}
		p.SetRoundScore(deadwoodValue)
		p.CommitRoundScore()
		g.appendLog(i, "score", fmt.Sprintf("%s scores %d (total %d)", playerName(g.players, i), deadwoodValue, p.GetCumulativeScore()), nil)
	}

	// 脱落判定
	for i, p := range g.players {
		if !p.GetEliminated() && p.GetCumulativeScore() > g.config.EliminationLimit {
			p.SetEliminated(true)
			g.appendLog(i, "eliminate", fmt.Sprintf("%s is eliminated (%d)", playerName(g.players, i), p.GetCumulativeScore()), nil)
		}
	}

	g.checkMatchEnd()
	if !g.gameEndFlag {
		g.phase = ChinchonPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによる引き分け (デッドウッドは加算する)。
func (g *Chinchon) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends (stock exhausted)", nil)
	g.knockerIdx = -1
	g.knockerDeadwood = nil
	g.scoreRound()
}

// checkChinchonForAll は全プレイヤーのチンチョンを判定する (配布直後用)。
func (g *Chinchon) checkChinchonForAll() {
	for i := range g.players {
		if g.checkChinchon(i) {
			return
		}
	}
}

// checkChinchon は idx の手札がチンチョンなら即時ゲーム勝利として終了させ true を返す。
func (g *Chinchon) checkChinchon(idx int) bool {
	if g.gameEndFlag {
		return false
	}
	p := g.players[idx]
	if p.GetEliminated() {
		return false
	}
	if hasChinchon(handCards(p)) {
		g.appendLog(idx, "chinchon", fmt.Sprintf("%s declares Chinchón and wins the game!", playerName(g.players, idx)), nil)
		g.winnerIdx = idx
		g.gameEndFlag = true
		g.phase = ChinchonPhaseGameEnd
		return true
	}
	return false
}

// checkMatchEnd はマッチ終了判定を行う。アクティブ (非脱落) が1人になればそのプレイヤーが勝者。
func (g *Chinchon) checkMatchEnd() {
	active := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if !p.GetEliminated() {
			active = append(active, i)
		}
	}
	if len(active) <= 1 {
		g.gameEndFlag = true
		g.phase = ChinchonPhaseGameEnd
		if len(active) == 1 {
			g.winnerIdx = active[0]
			g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, g.winnerIdx)), nil)
		} else {
			g.winnerIdx = -1
			g.appendLog(-1, "game_end", "Match ends with no survivor", nil)
		}
	}
}

// advanceTurn 次のアクティブプレイヤーへ手番を進める。
func (g *Chinchon) advanceTurn() {
	g.currentPlayerIdx = g.nextActivePlayer(g.currentPlayerIdx)
	g.phase = ChinchonPhaseDraw
}

// nextActivePlayer は from の次の非脱落プレイヤーの席を返す。
func (g *Chinchon) nextActivePlayer(from int) int {
	for i := 1; i <= len(g.players); i++ {
		seat := (from + i) % len(g.players)
		if !g.players[seat].GetEliminated() {
			return seat
		}
	}
	return from
}

// firstActivePlayer は席順で最初の非脱落プレイヤーを返す。
func (g *Chinchon) firstActivePlayer() int {
	for i := range g.players {
		if !g.players[i].GetEliminated() {
			return i
		}
	}
	return 0
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行する。
func (g *Chinchon) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	switch g.phase {
	case ChinchonPhaseDraw:
		g.cpuDraw()
	case ChinchonPhaseDiscard:
		g.cpuDiscardOrKnock()
	case ChinchonPhaseLayoff:
		g.cpuLayoff()
	}
}

// cpuDraw CPUがドローする (捨て札がデッドウッドを減らすなら取る)。
func (g *Chinchon) cpuDraw() {
	idx := g.currentPlayerIdx
	if len(g.discardPile) > 0 {
		top := g.discardPile[len(g.discardPile)-1]
		hand := handCards(g.players[idx])
		_, dwWithout := chinchonFindBestMelds(hand)
		_, dwWith := chinchonFindBestMelds(append(append([]*Card{}, hand...), top))
		if CalcDeadwoodValue(dwWith) < CalcDeadwoodValue(dwWithout) {
			g.doDrawDiscard(idx)
			return
		}
	}
	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}
	g.doDrawStock(idx)
}

// cpuDiscardOrKnock CPUが最適な捨て札を選び、ノック可能ならノックする。
func (g *Chinchon) cpuDiscardOrKnock() {
	idx := g.currentPlayerIdx
	player := g.players[idx]

	bestDiscardIdx := 0
	bestDeadwood := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		_, dw := chinchonFindBestMelds(cardsExcludingIndex(player, i))
		dwVal := CalcDeadwoodValue(dw)
		if bestDeadwood < 0 || dwVal < bestDeadwood {
			bestDeadwood = dwVal
			bestDiscardIdx = i
		}
	}

	if bestDeadwood <= g.config.KnockThreshold {
		g.executeKnock(idx, bestDiscardIdx)
		return
	}

	discarded := player.RemoveCard(bestDiscardIdx)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(idx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, idx), cardStr(discarded)), []*Card{discarded})
	if g.checkChinchon(idx) {
		return
	}
	g.advanceTurn()
}

// cpuLayoff CPUがレイオフ可能なカードをすべて付け、次の手番へ進める。
func (g *Chinchon) cpuLayoff() {
	player := g.players[g.currentPlayerIdx]
	for {
		found := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if g.canLayoff(card) {
				g.layoffCard(card)
				player.RemoveCard(i)
				g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	g.advanceLayoff()
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Chinchon) GetPhase() ChinchonPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Chinchon) SetPhase(phase ChinchonPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Chinchon) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Chinchon) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Chinchon) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Chinchon) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Chinchon) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Chinchon) SetDiscardPile(pile []*Card) {
	if pile == nil {
		pile = make([]*Card, 0)
	}
	g.discardPile = pile
}

// GetDiscardTop 捨て札の一番上を取得
func (g *Chinchon) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Chinchon) GetDrawPileCount() int { return len(g.drawPile) }

// SetStock 山札を設定 (テスト用)
func (g *Chinchon) SetStock(pile []*Card) {
	if pile == nil {
		pile = make([]*Card, 0)
	}
	g.drawPile = pile
}

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Chinchon) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx マッチ勝者インデックス取得 (-1 = 未確定/勝者なし)
func (g *Chinchon) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Chinchon) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Chinchon) GetPlayer(i int) *ChinchonPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Chinchon) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetPlayerDeadwoodValue はプレイヤーの手札を最善メルド分割したときのデッドウッド点を返す（プレゼンター向け）。
func (g *Chinchon) GetPlayerDeadwoodValue(i int) int {
	p := g.GetPlayer(i)
	if p == nil {
		return 0
	}
	_, dead := chinchonFindBestMelds(handCards(p))
	return CalcDeadwoodValue(dead)
}

// GetPlayerMeldSplit はプレイヤーの手札を最善のメルド群と残りのデッドウッドに
// 分けて返す。GetPlayerDeadwoodValue と同じ分割なので、デッドウッドの合計は
// 必ずその値に一致する。
//
// **どの札が成立しているかは捨て札選びの前提。**Web は緑/破線で色分けし内訳まで
// 出しているのに、CUI は合計点しか出していなかった (#4838)。
func (g *Chinchon) GetPlayerMeldSplit(i int) (melds [][]*Card, deadwood []*Card) {
	p := g.GetPlayer(i)
	if p == nil {
		return nil, nil
	}
	return chinchonFindBestMelds(handCards(p))
}

// GetKnockThreshold はノック可能なデッドウッド点の上限（この値以下でノック可）を返す。
func (g *Chinchon) GetKnockThreshold() int { return g.config.KnockThreshold }

// GetConfig 設定取得
func (g *Chinchon) GetConfig() ChinchonConfig { return g.config }

// SetConfig 設定変更
func (g *Chinchon) SetConfig(cfg ChinchonConfig) { g.config = cfg }

// GetKnockerIdx ノッカーのインデックス取得 (-1 = ノックなし)
func (g *Chinchon) GetKnockerIdx() int { return g.knockerIdx }

// SetKnockerIdx ノッカーのインデックス設定 (テスト用)
func (g *Chinchon) SetKnockerIdx(idx int) { g.knockerIdx = idx }

// GetKnockerMelds ノッカーのメルド取得
func (g *Chinchon) GetKnockerMelds() [][]*Card { return g.knockerMelds }

// SetKnockerMelds ノッカーのメルド設定 (テスト用)
func (g *Chinchon) SetKnockerMelds(melds [][]*Card) { g.knockerMelds = melds }

// GetKnockerDeadwood ノッカーのデッドウッド取得
func (g *Chinchon) GetKnockerDeadwood() []*Card { return g.knockerDeadwood }

// SetKnockerDeadwood ノッカーのデッドウッド設定 (テスト用)
func (g *Chinchon) SetKnockerDeadwood(deadwood []*Card) { g.knockerDeadwood = deadwood }

// --- Private helpers ---

// cardsExcludingIndex は player の手札から cardIndex を除いたカード列を返す。
func cardsExcludingIndex(player *ChinchonPlayer, cardIndex int) []*Card {
	out := make([]*Card, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		if i != cardIndex {
			out = append(out, player.GetCard(i))
		}
	}
	return out
}

// handCards は player の全手札をスライスで返す。
func handCards(player *ChinchonPlayer) []*Card {
	out := make([]*Card, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		out[i] = player.GetCard(i)
	}
	return out
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *Chinchon) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→ラン位置の順にソートする
func (g *Chinchon) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	cards := handCards(p)
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return chinchonRankPosition(cards[i].GetValue()) < chinchonRankPosition(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- JSON ---

// chinchonJSON is the JSON wire format for Chinchon.
type chinchonJSON struct {
	Players          []*ChinchonPlayer `json:"pl"`
	Config           ChinchonConfig    `json:"cf"`
	Phase            ChinchonPhase     `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
	KnockerIdx       int               `json:"ki"`
	KnockerMelds     [][]*Card         `json:"km"`
	KnockerDeadwood  []*Card           `json:"kd"`
	LayoffQueue      []int             `json:"lq"`
}

// MarshalJSON implements json.Marshaler.
func (g *Chinchon) MarshalJSON() ([]byte, error) {
	return json.Marshal(chinchonJSON{
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		KnockerIdx:       g.knockerIdx,
		KnockerMelds:     g.knockerMelds,
		KnockerDeadwood:  g.knockerDeadwood,
		LayoffQueue:      g.layoffQueue,
	})
}

// chinchonMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const chinchonMaxSliceLen = 1000

// errChinchonInvalidState は不正なデシリアライズ入力を表す共有センチネルエラー。
var errChinchonInvalidState = fmt.Errorf("chinchon: invalid serialized state")

// UnmarshalJSON implements json.Unmarshaler.
//
// 入力を厳格に検証する: プレイヤー数は設定の PlayerCount と一致 (2-4、nil要素不可)、
// フェーズ・インデックスは範囲内、設定は Validate() を通過、スライス長は上限以内。
func (g *Chinchon) UnmarshalJSON(data []byte) error {
	var j chinchonJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	if err := j.Config.Validate(); err != nil {
		return errChinchonInvalidState
	}
	if len(j.Players) != j.Config.PlayerCount {
		return errChinchonInvalidState
	}
	for _, p := range j.Players {
		if p == nil {
			return errChinchonInvalidState
		}
	}
	if len(j.DiscardPile) > chinchonMaxSliceLen || len(j.DrawPile) > chinchonMaxSliceLen ||
		len(j.ActionLog) > chinchonMaxSliceLen || len(j.KnockerMelds) > chinchonMaxSliceLen ||
		len(j.KnockerDeadwood) > chinchonMaxSliceLen || len(j.LayoffQueue) > chinchonMaxSliceLen {
		return errChinchonInvalidState
	}
	if j.Phase < ChinchonPhaseDraw || j.Phase > ChinchonPhaseGameEnd {
		return errChinchonInvalidState
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= len(j.Players) {
		return errChinchonInvalidState
	}
	if j.KnockerIdx < -1 || j.KnockerIdx >= len(j.Players) {
		return errChinchonInvalidState
	}
	for _, seat := range j.LayoffQueue {
		if seat < 0 || seat >= len(j.Players) {
			return errChinchonInvalidState
		}
	}

	g.players = j.Players
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
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.knockerIdx = j.KnockerIdx
	g.knockerMelds = j.KnockerMelds
	if g.knockerMelds == nil {
		g.knockerMelds = make([][]*Card, 0)
	}
	g.knockerDeadwood = j.KnockerDeadwood
	if g.knockerDeadwood == nil {
		g.knockerDeadwood = make([]*Card, 0)
	}
	g.layoffQueue = j.LayoffQueue
	if g.layoffQueue == nil {
		g.layoffQueue = make([]int, 0)
	}
	return nil
}

// chinchonRemapValue は 8/9/10 を除いた40枚デッキのランクを連続位置
// (A=1..7=7, J=8, Q=9, K=10) に写像したカードを返す。Gin Rummy の値ベースの
// ラン検出が ♠7-♠J を隣接として扱えるようにするための内部表現。
func chinchonRemapValue(c *Card) *Card {
	return NewCard(c.GetDesign(), chinchonRankPosition(c.GetValue()), false)
}

// chinchonFindBestMelds は FindBestMelds を 40枚デッキのランク隣接 (7-J 隣接)
// で実行するラッパー。各カードを連続位置に写像してから Gin Rummy のメルド探索を
// 呼び出し、結果のメルド・デッドウッドを元のカードに写し戻す。デッドウッドの
// 点数計算は呼び出し側が元のカードに対して行うため正しい点数が得られる。
func chinchonFindBestMelds(cards []*Card) (melds [][]*Card, deadwood []*Card) {
	remap := make([]*Card, len(cards))
	back := make(map[*Card]*Card, len(cards))
	for i, c := range cards {
		rc := chinchonRemapValue(c)
		remap[i] = rc
		back[rc] = c
	}
	rMelds, rDead := FindBestMelds(remap)
	melds = make([][]*Card, len(rMelds))
	for i, m := range rMelds {
		grp := make([]*Card, len(m))
		for j, rc := range m {
			grp[j] = back[rc]
		}
		melds[i] = grp
	}
	deadwood = make([]*Card, len(rDead))
	for i, rc := range rDead {
		deadwood[i] = back[rc]
	}
	return melds, deadwood
}

// chinchonCanAddToMeld は canAddToMeld を 40枚デッキのランク隣接で実行する
// ラッパー。メルドと追加カードを連続位置に写像してから判定する。
func chinchonCanAddToMeld(meld []*Card, card *Card) bool {
	rMeld := make([]*Card, len(meld))
	for i, c := range meld {
		rMeld[i] = chinchonRemapValue(c)
	}
	return canAddToMeld(rMeld, chinchonRemapValue(card))
}
