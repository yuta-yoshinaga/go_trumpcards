//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// ThreeThirteen（スリー・サーティーン）はアメリカ系のプログレッシブ・ラミー。
//
// ルール概要:
//   - 2〜4 人。2 デッキ（ジョーカーなし、合計 104 枚）を使用する。
//   - 全 11 ラウンド。ラウンド R(1..11) では各プレイヤーに R+2 枚配り、
//     値が R+2 のランクがそのラウンドのワイルドになる。
//     （ラウンド 1: 3 枚配り、ランク 3 がワイルド … ラウンド 11: 13 枚配り、ランク 13=K がワイルド）
//   - 手番: 山札トップ or 捨て札トップを 1 枚引く → 1 枚捨てる。
//   - 手札全体をメルド（同ランク 3 枚以上のセット／同スート連続 3 枚以上のラン、
//     ワイルド使用可）に分割でき、デッドウッドが 0 になるなら「ノック」してラウンドを終了できる。
//     （本実装ではノックは手札を完全にメルドできる場合＝デッドウッド 0 のときのみ許可する。）
//   - ノック後、他の各プレイヤーは最後の手番（ドロー＋ディスカード）を 1 回ずつ行い、
//     その後に全員の残り（デッドウッド）を集計してそのラウンドの得点とする。
//   - 山札が枯渇したラウンドは捨て札を再利用し、再利用できなければそのラウンドを終了する。
//   - 11 ラウンド終了後、累計得点が最も低いプレイヤーが勝者。
//
// カード値スキーム（threeThirteenCardValue）:
//   A=1, 2〜10=額面, J/Q/K=10。
//   デッドウッドに残ったワイルドは ThreeThirteenWildDeadwoodValue 点として数える。
//   この値はそのラウンドの得点（＝ペナルティ）にそのまま用いる。

// ThreeThirteenMeldMinSize メルド（セット／ラン）の最小枚数
const ThreeThirteenMeldMinSize = 3

// ThreeThirteenMinRound 最初のラウンド番号
const ThreeThirteenMinRound = 1

// ThreeThirteenMaxRound 最終ラウンド番号
const ThreeThirteenMaxRound = 11

// ThreeThirteenWildDeadwoodValue デッドウッドに残ったワイルドの点（ペナルティ計算用）
const ThreeThirteenWildDeadwoodValue = 20

// ThreeThirteenPhase ゲームフェーズ
type ThreeThirteenPhase int

// Three Thirteen のフェーズ定数
const (
	// ThreeThirteenPhaseDraw ドローフェーズ（山札 or 捨て札トップから 1 枚引く）
	ThreeThirteenPhaseDraw ThreeThirteenPhase = 0
	// ThreeThirteenPhaseDiscard ディスカードフェーズ（1 枚捨てる／ノックする）
	ThreeThirteenPhaseDiscard ThreeThirteenPhase = 1
	// ThreeThirteenPhaseRoundEnd ラウンド終了フェーズ
	ThreeThirteenPhaseRoundEnd ThreeThirteenPhase = 2
	// ThreeThirteenPhaseGameEnd ゲーム終了フェーズ
	ThreeThirteenPhaseGameEnd ThreeThirteenPhase = 3
)

// threeThirteenMaxTurns 全 CPU 戦などでの暴走防止用のターン上限
const threeThirteenMaxTurns = 5000

// ThreeThirteen スリー・サーティーンのゲームクラス。
type ThreeThirteen struct {
	trumpCards       *TrumpCards
	players          []*ThreeThirteenPlayer
	config           ThreeThirteenConfig
	phase            ThreeThirteenPhase
	round            int
	currentPlayerIdx int
	knockerIdx       int // ノックしたプレイヤー。-1 はまだ誰もノックしていない
	finalTurnsLeft   int // ノック後に残っている最終手番の回数
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	turnCount        int
	actionLogBase
}

// NewThreeThirteen コンストラクタ
func NewThreeThirteen(trumpCards *TrumpCards, players []*ThreeThirteenPlayer, config ThreeThirteenConfig) *ThreeThirteen {
	return &ThreeThirteen{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round:      ThreeThirteenMinRound,
		winnerIdx:  -1,
		knockerIdx: -1,
	}
}

// NewDefaultThreeThirteen 標準構成（人間 1 + CPU 3、2 デッキ・ジョーカーなし）でコンストラクトする SSoT。
func NewDefaultThreeThirteen() *ThreeThirteen {
	cfg := DefaultThreeThirteenConfig()
	return NewThreeThirteen(NewTrumpCardsWithDecks(2, 0), buildThreeThirteenPlayers(cfg.PlayerCount), cfg)
}

// buildThreeThirteenPlayers 人間 1 + CPU(n-1) のプレイヤースライスを作る。
func buildThreeThirteenPlayers(n int) []*ThreeThirteenPlayer {
	if n < ThreeThirteenMinPlayers {
		n = ThreeThirteenMinPlayers
	}
	if n > ThreeThirteenMaxPlayers {
		n = ThreeThirteenMaxPlayers
	}
	players := make([]*ThreeThirteenPlayer, 0, n)
	players = append(players, NewThreeThirteenPlayer(true))
	for i := 1; i < n; i++ {
		players = append(players, NewThreeThirteenPlayer(false))
	}
	return players
}

// WildRank そのラウンドのワイルドランクを返す（round+2）。
func ThreeThirteenWildRankFor(round int) int { return round + 2 }

// DealCount そのラウンドの 1 人あたり配布枚数を返す（round+2）。
func ThreeThirteenDealCountFor(round int) int { return round + 2 }

// Reset ゲームを最初のラウンドから初期化する
func (g *ThreeThirteen) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.round = ThreeThirteenMinRound
	g.currentPlayerIdx = 0
	g.knockerIdx = -1
	g.finalTurnsLeft = 0
	g.turnCount = 0
	g.actionLog = nil
	g.discardPile = nil
	g.drawPile = nil

	// プレイヤー数が設定と食い違う場合は作り直す。
	if len(g.players) != g.config.PlayerCount {
		g.players = buildThreeThirteenPlayers(g.config.PlayerCount)
	}
	for _, p := range g.players {
		p.SetCumulativeScore(0)
		p.ResetRound()
	}

	g.dealRound()
	g.phase = ThreeThirteenPhaseDraw
}

// NextRound 次のラウンドを開始する。ラウンド終了フェーズでのみ機能する。
// 11 ラウンド消化後はゲーム終了を確定する。
func (g *ThreeThirteen) NextRound() {
	if g.phase != ThreeThirteenPhaseRoundEnd {
		return
	}
	if g.round >= ThreeThirteenMaxRound {
		g.finalizeGameEnd()
		return
	}
	g.round++
	g.currentPlayerIdx = 0
	g.knockerIdx = -1
	g.finalTurnsLeft = 0
	for _, p := range g.players {
		p.SetRoundScore(0)
		p.Reset()
		p.SetIsFinished(false)
	}
	g.dealRound()
	g.phase = ThreeThirteenPhaseDraw
}

// dealRound 現在のラウンドの配布を行う。
func (g *ThreeThirteen) dealRound() {
	g.discardPile = nil
	g.trumpCards.Shuffle()

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

	dealCount := ThreeThirteenDealCountFor(g.round)
	for range dealCount {
		for j := range len(g.players) {
			if len(g.drawPile) == 0 {
				break
			}
			card := g.drawPile[len(g.drawPile)-1]
			g.drawPile = g.drawPile[:len(g.drawPile)-1]
			g.players[j].AddCard(card)
		}
	}
	if len(g.drawPile) > 0 {
		first := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, first)
	}
	g.sortAllHands()
}

// --- Human actions ---

// PlayerDrawFromStock 人間プレイヤーが山札から引く
func (g *ThreeThirteen) PlayerDrawFromStock() error {
	if err := g.guardHumanDraw(); err != nil {
		return err
	}
	return g.drawFromStock()
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札トップから引く
func (g *ThreeThirteen) PlayerDrawFromDiscard() error {
	if err := g.guardHumanDraw(); err != nil {
		return err
	}
	return g.drawFromDiscard()
}

func (g *ThreeThirteen) guardHumanDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ThreeThirteenPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

func (g *ThreeThirteen) drawFromStock() error {
	if len(g.drawPile) == 0 {
		if !g.recycleDiscardIntoStock() {
			g.endRoundStockOut()
			return nil
		}
	}
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
	g.phase = ThreeThirteenPhaseDiscard
	return nil
}

func (g *ThreeThirteen) drawFromDiscard() error {
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	g.phase = ThreeThirteenPhaseDiscard
	return nil
}

// recycleDiscardIntoStock 山札が空のとき捨て札トップ 1 枚を残して残りを山札へ戻しシャッフルする。
func (g *ThreeThirteen) recycleDiscardIntoStock() bool {
	if len(g.discardPile) <= 1 {
		return false
	}
	top := g.discardPile[len(g.discardPile)-1]
	rest := g.discardPile[:len(g.discardPile)-1]
	g.discardPile = []*Card{top}
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	g.drawPile = append(g.drawPile, rest...)
	g.appendLog(-1, "recycle", fmt.Sprintf("Discard pile recycled into stock (%d cards)", len(rest)), nil)
	return true
}

// PlayerDiscard 人間プレイヤーが手札 1 枚を捨ててターン終了する
func (g *ThreeThirteen) PlayerDiscard(cardIndex int) error {
	if err := g.guardHumanDiscard(); err != nil {
		return err
	}
	return g.applyDiscard(cardIndex, false)
}

// PlayerKnock 人間プレイヤーが cardIndex を捨ててノックする。
// 残りの手札がデッドウッド 0（完全メルド）でなければ拒否する。
func (g *ThreeThirteen) PlayerKnock(cardIndex int) error {
	if err := g.guardHumanDiscard(); err != nil {
		return err
	}
	return g.applyDiscard(cardIndex, true)
}

func (g *ThreeThirteen) guardHumanDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ThreeThirteenPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// applyDiscard cardIndex を捨てる。knock=true なら捨てた後の手札が完全メルドであることを要求する。
func (g *ThreeThirteen) applyDiscard(cardIndex int, knock bool) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	if knock {
		if g.knockerIdx >= 0 {
			return NewDomainError(ErrInvalidPlay, "既にノックされています")
		}
		remaining := handWithout(player, cardIndex)
		_, deadwood := threeThirteenBestMelds(remaining, g.WildRank())
		if threeThirteenDeadwoodValue(deadwood, g.WildRank()) != 0 {
			return NewDomainError(ErrInvalidPlay, "手札を完全にメルドできないためノックできません")
		}
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	if knock {
		g.knockerIdx = g.currentPlayerIdx
		player.SetIsFinished(true)
		g.finalTurnsLeft = len(g.players) - 1
		g.appendLog(g.currentPlayerIdx, "knock", fmt.Sprintf("%s knocks!", playerName(g.players, g.currentPlayerIdx)), nil)
	}

	g.advanceTurn()
	return nil
}

// advanceTurn 次のプレイヤーへ。ノック後は最終手番のカウントを減らし、0 になったらラウンドを締める。
func (g *ThreeThirteen) advanceTurn() {
	if g.knockerIdx >= 0 {
		if g.finalTurnsLeft <= 0 {
			g.finishRound()
			return
		}
		g.finalTurnsLeft--
	}
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	if g.knockerIdx >= 0 && g.currentPlayerIdx == g.knockerIdx {
		// ノッカーへ戻ってきたらラウンド終了。
		g.finishRound()
		return
	}
	g.phase = ThreeThirteenPhaseDraw
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *ThreeThirteen) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// CpuPlay 現在の手番が CPU の場合にターンを実行する
func (g *ThreeThirteen) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	if g.phase == ThreeThirteenPhaseRoundEnd || g.phase == ThreeThirteenPhaseGameEnd {
		return
	}
	g.turnCount++
	if g.turnCount > threeThirteenMaxTurns {
		g.endRoundStockOut()
		return
	}

	switch g.phase {
	case ThreeThirteenPhaseDraw:
		g.cpuDraw()
	case ThreeThirteenPhaseDiscard:
		g.cpuDiscard()
	}
}

// cpuDraw CPU の引き処理。捨て札トップが手役のデッドウッドを減らすなら拾い、そうでなければ山札から引く。
func (g *ThreeThirteen) cpuDraw() {
	top := g.GetDiscardTop()
	if top != nil && g.cpuShouldTakeDiscard(top) {
		_ = g.drawFromDiscard()
		return
	}
	_ = g.drawFromStock()
}

// cpuShouldTakeDiscard 捨て札トップを拾うべきかを返す。
func (g *ThreeThirteen) cpuShouldTakeDiscard(top *Card) bool {
	player := g.players[g.currentPlayerIdx]
	wild := g.WildRank()
	// ワイルドは常に拾う価値がある。
	if threeThirteenIsWild(top, wild) {
		return true
	}
	// 拾った後のデッドウッドが、捨てる候補を考慮しても改善するなら拾う。
	current := collectThreeThirteenCards(player)
	_, curDead := threeThirteenBestMelds(current, wild)
	curVal := threeThirteenDeadwoodValue(curDead, wild)

	withTop := append(append([]*Card{}, current...), top)
	bestAfter := threeThirteenBestDiscardValue(withTop, wild)
	if bestAfter < curVal {
		switch g.config.CpuDifficulty {
		case ThreeThirteenCpuDifficultyEasy:
			return rand.Intn(2) == 0
		default:
			return true
		}
	}
	return false
}

// cpuDiscard CPU のディスカード処理。完全メルドできるならノック、そうでなければデッドウッドを最小化する 1 枚を捨てる。
func (g *ThreeThirteen) cpuDiscard() {
	player := g.players[g.currentPlayerIdx]
	wild := g.WildRank()
	idx := g.chooseCpuDiscard(player)

	// ノック判定: そのカードを捨てた残りが完全メルドで、まだ誰もノックしていなければノック。
	if g.knockerIdx < 0 {
		remaining := handWithout(player, idx)
		_, dead := threeThirteenBestMelds(remaining, wild)
		if threeThirteenDeadwoodValue(dead, wild) == 0 {
			_ = g.applyDiscard(idx, true)
			return
		}
	}
	_ = g.applyDiscard(idx, false)
}

// chooseCpuDiscard CPU が捨てるカードを選ぶ（捨てた後のデッドウッドが最小になるカード）。
func (g *ThreeThirteen) chooseCpuDiscard(player *ThreeThirteenPlayer) int {
	wild := g.WildRank()
	n := player.GetCardsSize()
	if n == 0 {
		return 0
	}
	bestIdx := 0
	bestVal := -1
	for i := 0; i < n; i++ {
		remaining := handWithout(player, i)
		_, dead := threeThirteenBestMelds(remaining, wild)
		v := threeThirteenDeadwoodValue(dead, wild)
		if bestVal < 0 || v < bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

// finishRound ラウンドの集計を行い、各プレイヤーのデッドウッドを得点（ペナルティ）として加算する。
func (g *ThreeThirteen) finishRound() {
	if g.phase == ThreeThirteenPhaseRoundEnd || g.phase == ThreeThirteenPhaseGameEnd {
		return
	}
	wild := g.WildRank()
	for i := range g.players {
		cards := collectThreeThirteenCards(g.players[i])
		_, dead := threeThirteenBestMelds(cards, wild)
		score := threeThirteenDeadwoodValue(dead, wild)
		g.players[i].SetRoundScore(score)
		g.players[i].CommitRoundScore()
	}
	if g.knockerIdx >= 0 {
		g.appendLog(g.knockerIdx, "round_end", fmt.Sprintf("Round %d ends (%s knocked)", g.round, playerName(g.players, g.knockerIdx)), nil)
	} else {
		g.appendLog(-1, "round_end", fmt.Sprintf("Round %d ends (stock out)", g.round), nil)
	}
	g.phase = ThreeThirteenPhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了
func (g *ThreeThirteen) endRoundStockOut() {
	g.finishRound()
}

// finalizeGameEnd ゲーム終了処理（累計得点が最少のプレイヤーが勝者）
func (g *ThreeThirteen) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = ThreeThirteenPhaseGameEnd
	minScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() < minScore {
			minScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *ThreeThirteen) GetPhase() ThreeThirteenPhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *ThreeThirteen) SetPhase(p ThreeThirteenPhase) { g.phase = p }

// GetRound 現在のラウンド番号（1..11）
func (g *ThreeThirteen) GetRound() int { return g.round }

// SetRound ラウンド設定（テスト用）
func (g *ThreeThirteen) SetRound(r int) { g.round = r }

// WildRank そのラウンドのワイルドランク
func (g *ThreeThirteen) WildRank() int { return ThreeThirteenWildRankFor(g.round) }

// GetDealCount そのラウンドの 1 人あたり配布枚数
func (g *ThreeThirteen) GetDealCount() int { return ThreeThirteenDealCountFor(g.round) }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *ThreeThirteen) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *ThreeThirteen) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetKnockerIdx ノックしたプレイヤー（-1 = 未ノック）
func (g *ThreeThirteen) GetKnockerIdx() int { return g.knockerIdx }

// GetDiscardPile 捨て札の山
func (g *ThreeThirteen) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *ThreeThirteen) SetDiscardPile(p []*Card) { g.discardPile = p }

// GetDiscardTop 捨て札トップ
func (g *ThreeThirteen) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札残り枚数
func (g *ThreeThirteen) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *ThreeThirteen) SetDrawPile(p []*Card) { g.drawPile = p }

// GetGameEndFlag ゲーム終了フラグ
func (g *ThreeThirteen) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス
func (g *ThreeThirteen) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *ThreeThirteen) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *ThreeThirteen) GetPlayer(i int) *ThreeThirteenPlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *ThreeThirteen) GetConfig() ThreeThirteenConfig { return g.config }

// SetConfig 設定変更
func (g *ThreeThirteen) SetConfig(c ThreeThirteenConfig) { g.config = c }

// GetPlayerDeadwoodValue プレイヤーの最善メルド分割でのデッドウッド点を返す（プレゼンター向け）。
func (g *ThreeThirteen) GetPlayerDeadwoodValue(i int) int {
	p := g.GetPlayer(i)
	if p == nil {
		return 0
	}
	cards := collectThreeThirteenCards(p)
	_, dead := threeThirteenBestMelds(cards, g.WildRank())
	return threeThirteenDeadwoodValue(dead, g.WildRank())
}

// GetDeadwoodAfterDiscard returns the deadwood the player would be left with if
// they discarded the card at cardIndex, or -1 when the index is out of range.
//
// **捨てる前に分かる情報。**Web は 1 枚選ぶたびに予測デッドウッドを出しているのに、
// CUI は今の値しか出しておらず、どれを捨てると得かは実際に捨てるまで分からなかった
// (#4840)。ワイルドランクは現在のラウンドのものを使う。
func (g *ThreeThirteen) GetDeadwoodAfterDiscard(playerIdx, cardIndex int) int {
	p := g.GetPlayer(playerIdx)
	if p == nil || cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return -1
	}
	_, dead := threeThirteenBestMelds(handWithout(p, cardIndex), g.WildRank())
	return threeThirteenDeadwoodValue(dead, g.WildRank())
}

// --- Private helpers ---

func (g *ThreeThirteen) sortAllHands() {
	sortHands(len(g.players), g)
}

func (g *ThreeThirteen) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sortCards(cards)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// collectThreeThirteenCards プレイヤーの手札を []*Card で返す
func collectThreeThirteenCards(p *ThreeThirteenPlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// handWithout cardIndex を除いた手札を返す
func handWithout(p *ThreeThirteenPlayer, cardIndex int) []*Card {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if i == cardIndex {
			continue
		}
		cards = append(cards, p.GetCard(i))
	}
	return cards
}

// --- Pure card-evaluation helpers (wild-aware) ---

// threeThirteenIsWild card がそのラウンドのワイルド（値 == wildRank）か。
func threeThirteenIsWild(card *Card, wildRank int) bool {
	return card != nil && card.GetValue() == wildRank
}

// threeThirteenCardValue 1 枚のカードのデッドウッド点（A=1, 2-10=額面, J/Q/K=10）。
// ワイルドは threeThirteenDeadwoodValue 側で別扱いするため、ここでは素の点を返す。
func threeThirteenCardValue(card *Card) int {
	v := card.GetValue()
	if v >= 10 {
		return 10
	}
	return v
}

// threeThirteenDeadwoodValue デッドウッドの合計点。ワイルドは ThreeThirteenWildDeadwoodValue 点として数える。
func threeThirteenDeadwoodValue(deadwood []*Card, wildRank int) int {
	total := 0
	for _, c := range deadwood {
		if threeThirteenIsWild(c, wildRank) {
			total += ThreeThirteenWildDeadwoodValue
			continue
		}
		total += threeThirteenCardValue(c)
	}
	return total
}

// threeThirteenIsValidMeld cards が有効なメルド（セット 3+ またはラン 3+）か。ワイルドは任意の役割を代用する。
func threeThirteenIsValidMeld(cards []*Card, wildRank int) bool {
	if len(cards) < ThreeThirteenMeldMinSize {
		return false
	}
	return threeThirteenIsSet(cards, wildRank) || threeThirteenIsRun(cards, wildRank)
}

// threeThirteenIsSet cards が同ランクのセットか（ワイルドは任意ランクを代用）。
func threeThirteenIsSet(cards []*Card, wildRank int) bool {
	rank := -1
	for _, c := range cards {
		if threeThirteenIsWild(c, wildRank) {
			continue
		}
		if rank == -1 {
			rank = c.GetValue()
		} else if c.GetValue() != rank {
			return false
		}
	}
	// 最低 1 枚は自然牌が必要。
	return rank != -1
}

// threeThirteenIsRun cards が同スートの連続するランか（ワイルドは隙間を埋める）。
// Ace は low (A-2-3) または high (Q-K-A) のいずれでも有効。
func threeThirteenIsRun(cards []*Card, wildRank int) bool {
	suit := -1
	wilds := 0
	naturals := make([]int, 0, len(cards))
	for _, c := range cards {
		if threeThirteenIsWild(c, wildRank) {
			wilds++
			continue
		}
		if suit == -1 {
			suit = c.GetDesign()
		} else if c.GetDesign() != suit {
			return false
		}
		naturals = append(naturals, c.GetValue())
	}
	if suit == -1 {
		// 全部ワイルド: ランとしては無効
		return false
	}
	sort.Ints(naturals)
	for _, variant := range aceVariants(naturals) {
		if threeThirteenRunFits(variant, wilds) {
			return true
		}
	}
	return false
}

// threeThirteenRunFits ソート済みの自然牌の値列に wilds 枚のワイルドを足して、
// 連続するラン（重複なし）を構成できるかを判定する。
func threeThirteenRunFits(sortedNaturals []int, wilds int) bool {
	if len(sortedNaturals) == 0 {
		return false
	}
	for i := 1; i < len(sortedNaturals); i++ {
		if sortedNaturals[i] == sortedNaturals[i-1] {
			return false // 同じ値はランに置けない
		}
	}
	// ランは 1 スート最大 13 枚（A〜K）。
	if len(sortedNaturals)+wilds > 13 {
		return false
	}
	span := sortedNaturals[len(sortedNaturals)-1] - sortedNaturals[0] + 1
	gaps := span - len(sortedNaturals)
	return gaps >= 0 && gaps <= wilds
}

// threeThirteenBestMelds 手札を最善のメルド分割（デッドウッドを最小化）に分け、メルド群と残りを返す。
// GinRummy の findBestMeldsRecursive をワイルド対応の列挙・評価で置き換えたもの。
func threeThirteenBestMelds(cards []*Card, wildRank int) (melds [][]*Card, deadwood []*Card) {
	if len(cards) == 0 {
		return nil, nil
	}
	return threeThirteenBestMeldsRecursive(cards, nil, wildRank)
}

func threeThirteenBestMeldsRecursive(remaining []*Card, currentMelds [][]*Card, wildRank int) ([][]*Card, []*Card) {
	candidates := threeThirteenAllMelds(remaining, wildRank)
	if len(candidates) == 0 {
		return currentMelds, remaining
	}

	bestDeadwoodValue := threeThirteenDeadwoodValue(remaining, wildRank)
	bestMelds := currentMelds
	bestDeadwood := remaining

	for _, meld := range candidates {
		rest := excludeCards(remaining, meld)
		newMelds := append(copyMelds(currentMelds), meld)

		resultMelds, resultDeadwood := threeThirteenBestMeldsRecursive(rest, newMelds, wildRank)
		dv := threeThirteenDeadwoodValue(resultDeadwood, wildRank)
		if dv < bestDeadwoodValue {
			bestDeadwoodValue = dv
			bestMelds = resultMelds
			bestDeadwood = resultDeadwood
		}
		if bestDeadwoodValue == 0 {
			break // 完全メルド → 最適解
		}
	}
	return bestMelds, bestDeadwood
}

// threeThirteenAllMelds remaining から構成可能なメルド候補（3〜4 枚）を列挙する（ワイルド対応）。
func threeThirteenAllMelds(cards []*Card, wildRank int) [][]*Card {
	var melds [][]*Card
	wilds := make([]*Card, 0)
	naturals := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if threeThirteenIsWild(c, wildRank) {
			wilds = append(wilds, c)
		} else {
			naturals = append(naturals, c)
		}
	}

	// セット候補: 同ランクの自然牌 + 必要枚数のワイルドで 3 枚以上。
	byRank := make(map[int][]*Card)
	for _, c := range naturals {
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
	}
	for _, group := range byRank {
		for size := ThreeThirteenMeldMinSize; size <= ThreeThirteenMeldMinSize+1; size++ {
			needWild := size - len(group)
			if needWild < 0 {
				needWild = 0
			}
			natTake := size - needWild
			if natTake < 1 || natTake > len(group) || needWild > len(wilds) {
				continue
			}
			pick := make([]*Card, 0, size)
			pick = append(pick, group[:natTake]...)
			pick = append(pick, wilds[:needWild]...)
			if threeThirteenIsValidMeld(pick, wildRank) {
				melds = append(melds, pick)
			}
		}
	}

	// ラン候補: 同スートで連続 3〜4 枚（ワイルドで隙間を埋める）。
	bySuit := make(map[int]map[int]*Card)
	for _, c := range naturals {
		if bySuit[c.GetDesign()] == nil {
			bySuit[c.GetDesign()] = make(map[int]*Card)
		}
		if _, ok := bySuit[c.GetDesign()][c.GetValue()]; !ok {
			bySuit[c.GetDesign()][c.GetValue()] = c
		}
	}
	for _, byVal := range bySuit {
		melds = append(melds, threeThirteenRunsIn(byVal, wilds, wildRank)...)
	}
	return melds
}

// threeThirteenRunsIn 1 スート内の value→card マップから、ワイルドで隙間を埋めて
// 連続 3〜4 枚のランをすべて列挙する。Ace は high (14) としても扱う。
func threeThirteenRunsIn(byVal map[int]*Card, wilds []*Card, wildRank int) [][]*Card {
	var out [][]*Card
	// Ace-high 用に値 1 を 14 にもマップしたビューを作る。
	view := make(map[int]*Card, len(byVal)+1)
	for v, c := range byVal {
		view[v] = c
	}
	if c, ok := byVal[1]; ok {
		view[14] = c
	}
	for size := ThreeThirteenMeldMinSize; size <= ThreeThirteenMeldMinSize+1; size++ {
		for start := 1; start+size-1 <= 14; start++ {
			picked := make([]*Card, 0, size)
			wildsUsed := 0
			ok := true
			for v := start; v < start+size; v++ {
				if c, found := view[v]; found {
					picked = append(picked, c)
				} else if wildsUsed < len(wilds) {
					picked = append(picked, wilds[wildsUsed])
					wildsUsed++
				} else {
					ok = false
					break
				}
			}
			if ok && threeThirteenIsValidMeld(picked, wildRank) {
				out = append(out, picked)
			}
		}
	}
	return out
}

// threeThirteenBestDiscardValue cards から 1 枚捨てたときに到達できる最小のデッドウッド点を返す。
func threeThirteenBestDiscardValue(cards []*Card, wildRank int) int {
	if len(cards) == 0 {
		return 0
	}
	best := -1
	for i := range cards {
		rest := make([]*Card, 0, len(cards)-1)
		rest = append(rest, cards[:i]...)
		rest = append(rest, cards[i+1:]...)
		_, dead := threeThirteenBestMelds(rest, wildRank)
		v := threeThirteenDeadwoodValue(dead, wildRank)
		if best < 0 || v < best {
			best = v
		}
	}
	return best
}

// --- JSON ---

// threeThirteenJSON は ThreeThirteen の JSON 表現
type threeThirteenJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*ThreeThirteenPlayer `json:"pl"`
	Config           ThreeThirteenConfig    `json:"cf"`
	Phase            ThreeThirteenPhase     `json:"ps"`
	Round            int                    `json:"rd"`
	CurrentPlayerIdx int                    `json:"ci"`
	KnockerIdx       int                    `json:"kn"`
	FinalTurnsLeft   int                    `json:"ft"`
	DiscardPile      []*Card                `json:"dp"`
	DrawPile         []*Card                `json:"wp"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerIdx        int                    `json:"wi"`
	TurnCount        int                    `json:"tn"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *ThreeThirteen) MarshalJSON() ([]byte, error) {
	return json.Marshal(threeThirteenJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		Round:            g.round,
		CurrentPlayerIdx: g.currentPlayerIdx,
		KnockerIdx:       g.knockerIdx,
		FinalTurnsLeft:   g.finalTurnsLeft,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		TurnCount:        g.turnCount,
		ActionLog:        g.actionLog,
	})
}

const threeThirteenMaxSliceLen = 1000

// errThreeThirteenInvalidState は不正な復元データを表す共有センチネルエラー。
var errThreeThirteenInvalidState = fmt.Errorf("threethirteen: invalid game state")

// UnmarshalJSON implements json.Unmarshaler.
func (g *ThreeThirteen) UnmarshalJSON(data []byte) error {
	var j threeThirteenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > threeThirteenMaxSliceLen || len(j.DiscardPile) > threeThirteenMaxSliceLen ||
		len(j.DrawPile) > threeThirteenMaxSliceLen || len(j.ActionLog) > threeThirteenMaxSliceLen {
		return errThreeThirteenInvalidState
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) < ThreeThirteenMinPlayers || len(j.Players) > ThreeThirteenMaxPlayers {
		return errThreeThirteenInvalidState
	}
	for _, p := range j.Players {
		if p == nil {
			return errThreeThirteenInvalidState
		}
	}
	if j.Phase < ThreeThirteenPhaseDraw || j.Phase > ThreeThirteenPhaseGameEnd {
		return errThreeThirteenInvalidState
	}
	if j.Round < ThreeThirteenMinRound || j.Round > ThreeThirteenMaxRound {
		return errThreeThirteenInvalidState
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= len(j.Players) {
		return errThreeThirteenInvalidState
	}
	if j.KnockerIdx < -1 || j.KnockerIdx >= len(j.Players) {
		return errThreeThirteenInvalidState
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.round = j.Round
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.knockerIdx = j.KnockerIdx
	g.finalTurnsLeft = j.FinalTurnsLeft
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
	g.turnCount = j.TurnCount
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
