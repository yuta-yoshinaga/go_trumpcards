//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

// Rummy500PlayerCnt Rummy 500プレイヤー数（ヘッズアップ）
const Rummy500PlayerCnt = 2

// Rummy500HandSize 初期配布枚数（ヘッズアップ標準）
const Rummy500HandSize = 13

// Rummy500AceHighValue Q-K-Aランに含まれるエースの得点
const Rummy500AceHighValue = 15

// Rummy500Phase ゲームフェーズ
type Rummy500Phase int

// Rummy500のフェーズ定数
const (
	// Rummy500PhaseDraw ドローフェーズ (山札または捨て札から引く)
	Rummy500PhaseDraw Rummy500Phase = 0
	// Rummy500PhasePlay メルド/レイオフ/ディスカード可能フェーズ
	Rummy500PhasePlay Rummy500Phase = 1
	// Rummy500PhaseRoundEnd ラウンド終了フェーズ
	Rummy500PhaseRoundEnd Rummy500Phase = 2
	// Rummy500PhaseGameEnd ゲーム終了フェーズ
	Rummy500PhaseGameEnd Rummy500Phase = 3
)

// Rummy500 Rummy 500ゲームクラス
type Rummy500 struct {
	trumpCards       *TrumpCards
	players          []*Rummy500Player
	config           Rummy500Config
	phase            Rummy500Phase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
	roundEnderIdx int // ラウンド終了を引き起こしたプレイヤー（-1 = 山切れ）
}

// NewRummy500 コンストラクタ
func NewRummy500(trumpCards *TrumpCards, players []*Rummy500Player, config Rummy500Config) *Rummy500 {
	return &Rummy500{
		trumpCards:    trumpCards,
		players:       players,
		config:        config,
		winnerIdx:     -1,
		roundNumber:   0,
		roundEnderIdx: -1,
	}
}

// NewDefaultRummy500 returns Rummy500 with the standard 2-player setup (1 human, 1 CPU)
// and DefaultRummy500Config.
func NewDefaultRummy500() *Rummy500 {
	players := []*Rummy500Player{
		NewRummy500Player(true),
		NewRummy500Player(false),
	}
	return NewRummy500(NewTrumpCards(0), players, DefaultRummy500Config())
}

// Reset ゲーム初期化
func (g *Rummy500) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.roundEnderIdx = -1

	for _, p := range g.players {
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.Reset()
		p.SetLaidMelds(nil)
		p.SetIsFinished(false)
	}

	// trumpCards.Shuffle resets the drawn-card pointer via deckInit,
	// which dealInitialCards depends on to drain a full deck on every Reset.
	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = Rummy500PhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Rummy500) NextRound() {
	if g.phase != Rummy500PhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.roundEnderIdx = -1

	for _, p := range g.players {
		p.ResetRound()
	}

	// trumpCards.Shuffle resets the drawn-card pointer via deckInit,
	// which dealInitialCards depends on to drain a full deck on every round.
	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = Rummy500PhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに手札サイズ枚、1枚を捨て札に
func (g *Rummy500) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}

	rand.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	for i := 0; i < Rummy500HandSize; i++ {
		for j := 0; j < Rummy500PlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// 最初の1枚を捨て札に
	if len(g.drawPile) > 0 {
		firstCard := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, firstCard)
	}
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Rummy500) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != Rummy500PhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.drawPile) == 0 {
		g.endRoundStockEmpty()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	g.phase = Rummy500PhasePlay
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札からカードを引く
// idx は捨て札のインデックス（0が一番下、len-1が一番上）。
// idx から一番上までの全てのカードが手札に追加される（Rummy 500特有のルール）。
func (g *Rummy500) PlayerDrawFromDiscard(idx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != Rummy500PhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札がありません")
	}
	if idx < 0 || idx >= len(g.discardPile) {
		return NewDomainError(ErrInvalidCard, "捨て札インデックスが範囲外です")
	}

	taken := append([]*Card(nil), g.discardPile[idx:]...)
	g.discardPile = g.discardPile[:idx]

	for _, c := range taken {
		g.players[g.currentPlayerIdx].AddCard(c)
	}
	g.sortHand(g.currentPlayerIdx)

	first := taken[0]
	detail := fmt.Sprintf("%s draws %s from discard (+%d card(s))", playerName(g.players, g.currentPlayerIdx), cardStr(first), len(taken)-1)
	g.appendLog(g.currentPlayerIdx, "draw_discard", detail, taken)

	g.phase = Rummy500PhasePlay
	return nil
}

// PlayerMeld 人間プレイヤーが手札のカードでメルドを場に出す
func (g *Rummy500) PlayerMeld(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != Rummy500PhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.executeMeld(g.currentPlayerIdx, cardIndices)
}

// executeMeld メルドを実行する（人間/CPU共通）
func (g *Rummy500) executeMeld(playerIdx int, cardIndices []int) error {
	player := g.players[playerIdx]

	if len(cardIndices) < 3 {
		return NewDomainError(ErrInvalidPlay, "メルドは3枚以上のカードが必要です")
	}

	seen := make(map[int]bool)
	for _, i := range cardIndices {
		if i < 0 || i >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[i] {
			return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
		}
		seen[i] = true
	}

	meld := make([]*Card, len(cardIndices))
	for i, idx := range cardIndices {
		meld[i] = player.GetCard(idx)
	}
	if !Rummy500IsValidMeld(meld) {
		return NewDomainError(ErrInvalidPlay, "有効なメルド（同ランク3枚以上 または 同スート連続3枚以上）ではありません")
	}

	// 降順にインデックスを並べてカードを取り除く
	sortedIdx := make([]int, len(cardIndices))
	copy(sortedIdx, cardIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(sortedIdx)))
	for _, idx := range sortedIdx {
		player.RemoveCard(idx)
	}

	player.AddLaidMeld(meld)

	cardsCopy := make([]*Card, len(meld))
	copy(cardsCopy, meld)
	g.appendLog(playerIdx, "meld",
		fmt.Sprintf("%s melds %s", playerName(g.players, playerIdx), formatCards(meld)), cardsCopy)

	return nil
}

// PlayerLayoff 人間プレイヤーがレイオフ（既存メルドにカードを追加）する
func (g *Rummy500) PlayerLayoff(meldOwner, meldIdx, cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != Rummy500PhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.executeLayoff(g.currentPlayerIdx, meldOwner, meldIdx, cardIndex)
}

// executeLayoff レイオフを実行する（人間/CPU共通）
func (g *Rummy500) executeLayoff(playerIdx, meldOwner, meldIdx, cardIndex int) error {
	if meldOwner < 0 || meldOwner >= len(g.players) {
		return NewDomainError(ErrInvalidPlay, "メルド所有者が不正です")
	}
	owner := g.players[meldOwner]
	melds := owner.GetLaidMelds()
	if meldIdx < 0 || meldIdx >= len(melds) {
		return NewDomainError(ErrInvalidPlay, "メルドインデックスが範囲外です")
	}
	player := g.players[playerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if !Rummy500CanLayoff(melds[meldIdx], card) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%sはレイオフできません", cardStr(card)))
	}

	owner.AppendToLaidMeld(meldIdx, card)
	player.RemoveCard(cardIndex)

	g.appendLog(playerIdx, "layoff",
		fmt.Sprintf("%s lays off %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨ててターンを終える
func (g *Rummy500) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != Rummy500PhasePlay {
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

	g.appendLog(g.currentPlayerIdx, "discard",
		fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	if player.GetCardsSize() == 0 {
		// あがり
		g.roundEnderIdx = g.currentPlayerIdx
		g.appendLog(g.currentPlayerIdx, "go_out",
			fmt.Sprintf("%s goes out!", playerName(g.players, g.currentPlayerIdx)), nil)
		g.scoreRound()
		return nil
	}

	g.advanceTurn()
	return nil
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *Rummy500) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case Rummy500PhaseDraw:
		g.cpuDraw()
	case Rummy500PhasePlay:
		g.cpuPlayMelds()
	}
}

// cpuDraw CPUがドローする
func (g *Rummy500) cpuDraw() {
	// 山札優先（簡易戦略）
	if len(g.drawPile) > 0 {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.players[g.currentPlayerIdx].AddCard(card)
		g.sortHand(g.currentPlayerIdx)
		g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
		g.phase = Rummy500PhasePlay
		return
	}
	// 山切れの場合は捨て札のトップから1枚だけ取る
	if len(g.discardPile) > 0 {
		top := g.discardPile[len(g.discardPile)-1]
		g.discardPile = g.discardPile[:len(g.discardPile)-1]
		g.players[g.currentPlayerIdx].AddCard(top)
		g.sortHand(g.currentPlayerIdx)
		g.appendLog(g.currentPlayerIdx, "draw_discard",
			fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(top)), []*Card{top})
		g.phase = Rummy500PhasePlay
		return
	}
	g.endRoundStockEmpty()
}

// cpuPlayMelds CPUがメルド/レイオフ/ディスカードを実行する
func (g *Rummy500) cpuPlayMelds() {
	idx := g.currentPlayerIdx

	// 可能な限りメルドする
	for {
		melded := g.cpuTryMeld(idx)
		if !melded {
			break
		}
	}

	// 可能な限りレイオフする
	for {
		laid := g.cpuTryLayoff(idx)
		if !laid {
			break
		}
	}

	player := g.players[idx]
	if player.GetCardsSize() == 0 {
		// メルド/レイオフで手札を使い切った場合、強制的にダミーディスカードはできない
		// → ラウンド終了（あがり扱い）
		g.roundEnderIdx = idx
		g.appendLog(idx, "go_out",
			fmt.Sprintf("%s goes out!", playerName(g.players, idx)), nil)
		g.scoreRound()
		return
	}

	// ディスカード: 最も高得点のカード（手札に残すと負担になる）を捨てる
	discardIdx := g.cpuChooseDiscard(idx)
	discarded := player.RemoveCard(discardIdx)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(idx, "discard",
		fmt.Sprintf("%s discards %s", playerName(g.players, idx), cardStr(discarded)), []*Card{discarded})

	if player.GetCardsSize() == 0 {
		g.roundEnderIdx = idx
		g.appendLog(idx, "go_out",
			fmt.Sprintf("%s goes out!", playerName(g.players, idx)), nil)
		g.scoreRound()
		return
	}

	g.advanceTurn()
}

// cpuTryMeld 手札から有効なメルドを1つ探して場に出す。出せたらtrue。
func (g *Rummy500) cpuTryMeld(playerIdx int) bool {
	player := g.players[playerIdx]
	n := player.GetCardsSize()
	cards := make([]*Card, n)
	for i := 0; i < n; i++ {
		cards[i] = player.GetCard(i)
	}

	candidates := findAllRummy500Melds(cards)
	if len(candidates) == 0 {
		return false
	}

	// 一番高得点のメルドを選ぶ
	bestIdx := 0
	bestScore := Rummy500MeldScore(candidates[0])
	for i := 1; i < len(candidates); i++ {
		s := Rummy500MeldScore(candidates[i])
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}

	chosen := candidates[bestIdx]
	indices := indicesFor(cards, chosen)
	if indices == nil {
		return false
	}
	return g.executeMeld(playerIdx, indices) == nil
}

// cpuTryLayoff 手札から1枚レイオフ可能なカードを探してレイオフする
func (g *Rummy500) cpuTryLayoff(playerIdx int) bool {
	player := g.players[playerIdx]
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		for owner := 0; owner < len(g.players); owner++ {
			melds := g.players[owner].GetLaidMelds()
			for mi, meld := range melds {
				if Rummy500CanLayoff(meld, card) {
					if g.executeLayoff(playerIdx, owner, mi, i) == nil {
						return true
					}
				}
			}
		}
	}
	return false
}

// cpuChooseDiscard CPUの捨てカード選択（一番高得点のカードを捨てる）
func (g *Rummy500) cpuChooseDiscard(playerIdx int) int {
	player := g.players[playerIdx]
	bestIdx := 0
	bestVal := rummy500CardScoreLow(player.GetCard(0))
	for i := 1; i < player.GetCardsSize(); i++ {
		v := rummy500CardScoreLow(player.GetCard(i))
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

// scoreRound ラウンドのスコアを確定する
func (g *Rummy500) scoreRound() {
	for i, p := range g.players {
		meldScore := 0
		for _, meld := range p.GetLaidMelds() {
			meldScore += Rummy500MeldScore(meld)
		}
		handPenalty := 0
		for j := 0; j < p.GetCardsSize(); j++ {
			handPenalty += rummy500CardScoreLow(p.GetCard(j))
		}
		round := meldScore - handPenalty
		p.SetRoundScore(round)
		g.appendLog(i, "score",
			fmt.Sprintf("%s scores %d (melds %d - hand %d)", playerName(g.players, i), round, meldScore, handPenalty), nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = Rummy500PhaseRoundEnd
	}
}

// endRoundStockEmpty 山札切れによるラウンド終了
func (g *Rummy500) endRoundStockEmpty() {
	g.appendLog(-1, "stock_empty", "Stock is empty, round ends", nil)
	g.roundEnderIdx = -1
	g.scoreRound()
}

// ScoreRound ラウンドの得点を確定する（インタフェース互換用）
func (g *Rummy500) ScoreRound() {
	// scoreRoundはあがり/山切れ時に既に呼ばれている
}

// advanceTurn 次のプレイヤーへ
func (g *Rummy500) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % Rummy500PlayerCnt
	g.phase = Rummy500PhaseDraw
}

// checkGameEnd ゲーム終了判定
func (g *Rummy500) checkGameEnd() {
	hasReached := false
	for i := 0; i < Rummy500PlayerCnt; i++ {
		if g.players[i].GetCumulativeScore() >= g.config.PointLimit {
			hasReached = true
			break
		}
	}
	if !hasReached {
		return
	}

	g.gameEndFlag = true
	g.phase = Rummy500PhaseGameEnd

	maxScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < Rummy500PlayerCnt; i++ {
		if g.players[i].GetCumulativeScore() > maxScore {
			maxScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Rummy500) GetPhase() Rummy500Phase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Rummy500) SetPhase(phase Rummy500Phase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Rummy500) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Rummy500) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Rummy500) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Rummy500) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Rummy500) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Rummy500) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Rummy500) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Rummy500) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Rummy500) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Rummy500) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Rummy500) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Rummy500) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Rummy500) GetPlayer(i int) *Rummy500Player {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Rummy500) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Rummy500) GetConfig() Rummy500Config { return g.config }

// SetConfig 設定変更
func (g *Rummy500) SetConfig(cfg Rummy500Config) { g.config = cfg }

// GetRoundEnderIdx ラウンドを終わらせたプレイヤーのインデックス（-1=山切れ）
func (g *Rummy500) GetRoundEnderIdx() int { return g.roundEnderIdx }

// --- Private helpers ---

func (g *Rummy500) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

func (g *Rummy500) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- Meld / Run / Set / Layoff helpers ---

// rummy500CardScoreLow A=1, J/Q/K=10, それ以外=フェイスバリュー
func rummy500CardScoreLow(c *Card) int {
	v := c.GetValue()
	if v == 1 {
		return 1
	}
	if v >= 10 {
		return 10
	}
	return v
}

// Rummy500MeldScore メルドの得点（A-K含むランではAce=15、A-2-3ランではAce=1）
func Rummy500MeldScore(meld []*Card) int {
	if isSet(meld) {
		total := 0
		for _, c := range meld {
			total += rummy500CardScoreLow(c)
		}
		return total
	}
	// ラン: Q-K-Aを含む場合はAceを高(15)として扱う
	hasAce := false
	hasKing := false
	for _, c := range meld {
		if c.GetValue() == 1 {
			hasAce = true
		}
		if c.GetValue() == 13 {
			hasKing = true
		}
	}
	total := 0
	for _, c := range meld {
		if hasAce && hasKing && c.GetValue() == 1 {
			total += Rummy500AceHighValue
		} else {
			total += rummy500CardScoreLow(c)
		}
	}
	return total
}

// Rummy500IsValidMeld 有効なメルド（セット または ラン）かどうか
func Rummy500IsValidMeld(cards []*Card) bool {
	if len(cards) < 3 {
		return false
	}
	if rummy500IsValidSet(cards) {
		return true
	}
	if rummy500IsValidRun(cards) {
		return true
	}
	return false
}

// rummy500IsValidSet 同ランクで全て別スート（最大4枚）
func rummy500IsValidSet(cards []*Card) bool {
	if len(cards) < 3 || len(cards) > 4 {
		return false
	}
	v := cards[0].GetValue()
	seen := make(map[int]bool)
	for _, c := range cards {
		if c.GetValue() != v {
			return false
		}
		if seen[c.GetDesign()] {
			return false
		}
		seen[c.GetDesign()] = true
	}
	return true
}

// rummy500IsValidRun 同スートで連続したランクのカード（Aceは1または14として）
func rummy500IsValidRun(cards []*Card) bool {
	if len(cards) < 3 {
		return false
	}
	suit := cards[0].GetDesign()
	for _, c := range cards[1:] {
		if c.GetDesign() != suit {
			return false
		}
	}
	values := make([]int, len(cards))
	for i, c := range cards {
		values[i] = c.GetValue()
	}
	sort.Ints(values)
	// 重複は不可
	for i := 1; i < len(values); i++ {
		if values[i] == values[i-1] {
			return false
		}
	}
	// 連続チェック（Ace = 1 low）
	if isConsecutive(values) {
		return true
	}
	// Aceがあれば14 high として再評価
	if values[0] == 1 {
		highValues := make([]int, len(values))
		copy(highValues, values)
		highValues[0] = 14
		sort.Ints(highValues)
		if isConsecutive(highValues) {
			return true
		}
	}
	return false
}

// Rummy500CanLayoff 既存メルドにカードをレイオフできるかどうか
func Rummy500CanLayoff(meld []*Card, card *Card) bool {
	if len(meld) == 0 {
		return false
	}
	candidate := append([]*Card(nil), meld...)
	candidate = append(candidate, card)
	if isSet(meld) {
		// セット拡張: 同ランクかつスート未使用、4枚以下
		if len(meld) >= 4 {
			return false
		}
		if card.GetValue() != meld[0].GetValue() {
			return false
		}
		for _, m := range meld {
			if m.GetDesign() == card.GetDesign() {
				return false
			}
		}
		return rummy500IsValidSet(candidate)
	}
	// ラン拡張: 同スートで連続
	return rummy500IsValidRun(candidate)
}

// findAllRummy500Melds 与えられたカードから有効なメルド候補をすべて列挙する
func findAllRummy500Melds(cards []*Card) [][]*Card {
	var melds [][]*Card

	// セット候補
	byRank := make(map[int][]*Card)
	for _, c := range cards {
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
	}
	for _, group := range byRank {
		if len(group) >= 3 {
			melds = append(melds, append([]*Card{}, group[:3]...))
		}
		if len(group) >= 4 {
			melds = append(melds, append([]*Card{}, group[:4]...))
		}
	}

	// ラン候補
	bySuit := make(map[int][]*Card)
	for _, c := range cards {
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c)
	}
	for _, group := range bySuit {
		if len(group) < 3 {
			continue
		}
		sorted := make([]*Card, len(group))
		copy(sorted, group)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].GetValue() < sorted[j].GetValue()
		})
		for i := 0; i < len(sorted); i++ {
			run := []*Card{sorted[i]}
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].GetValue() == run[len(run)-1].GetValue()+1 {
					run = append(run, sorted[j])
					if len(run) >= 3 {
						runCopy := make([]*Card, len(run))
						copy(runCopy, run)
						melds = append(melds, runCopy)
					}
				} else {
					break
				}
			}
		}

		// Ace-highラン候補: Q-K-A や J-Q-K-A など、Aceを14として扱うラン。
		// 上の low-Ace パスでは Ace=1 が常に先頭に出るため検出できない。
		var ace *Card
		for _, c := range sorted {
			if c.GetValue() == 1 {
				ace = c
				break
			}
		}
		if ace == nil {
			continue
		}
		// Aceを除いた高位部分(>=2)を逆順に走査して連続ランを探す。
		highPart := make([]*Card, 0, len(sorted))
		for _, c := range sorted {
			if c.GetValue() != 1 {
				highPart = append(highPart, c)
			}
		}
		// highPartはすでに昇順
		if len(highPart) >= 2 && highPart[len(highPart)-1].GetValue() == 13 {
			// 末尾からKに連続するシーケンスを取り出す
			tail := []*Card{highPart[len(highPart)-1]}
			for k := len(highPart) - 2; k >= 0; k-- {
				if highPart[k].GetValue() == tail[0].GetValue()-1 {
					tail = append([]*Card{highPart[k]}, tail...)
				} else {
					break
				}
			}
			if len(tail) >= 2 {
				// tail + Ace は長さ 3 以上のラン
				runWithAce := append([]*Card{}, tail...)
				runWithAce = append(runWithAce, ace)
				melds = append(melds, runWithAce)
			}
		}
	}

	return melds
}

// indicesFor returns the positions of cards in source. Returns nil if any
// pointer is not found exactly once.
func indicesFor(source []*Card, cards []*Card) []int {
	indices := make([]int, 0, len(cards))
	used := make(map[int]bool)
	for _, c := range cards {
		found := -1
		for i, s := range source {
			if used[i] {
				continue
			}
			if s == c {
				found = i
				break
			}
		}
		if found < 0 {
			return nil
		}
		used[found] = true
		indices = append(indices, found)
	}
	return indices
}

// formatCards human-readable string for a list of cards.
func formatCards(cards []*Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cardStr(c))
	}
	return strings.Join(parts, ",")
}

// --- JSON wire format ---

type rummy500JSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*Rummy500Player `json:"pl"`
	Config           Rummy500Config    `json:"cf"`
	Phase            Rummy500Phase     `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
	RoundEnderIdx    int               `json:"re"`
}

// MarshalJSON implements json.Marshaler.
func (g *Rummy500) MarshalJSON() ([]byte, error) {
	return json.Marshal(rummy500JSON{
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
		ActionLog:        g.actionLog,
		RoundEnderIdx:    g.roundEnderIdx,
	})
}

const rummy500MaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Rummy500) UnmarshalJSON(data []byte) error {
	var j rummy500JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > rummy500MaxSliceLen || len(j.DiscardPile) > rummy500MaxSliceLen ||
		len(j.DrawPile) > rummy500MaxSliceLen || len(j.ActionLog) > rummy500MaxSliceLen {
		return fmt.Errorf("rummy500: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*Rummy500Player, 0)
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
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.roundEnderIdx = j.RoundEnderIdx
	return nil
}
