//go:build !js || !wasm || extra

// Package domain: Panguingue (Pan) implementation.
//
// Panguingue ("Pan") is a multi-deck rummy. This is a CLEAN, SIMPLIFIED and
// DETERMINISTIC interpretation — real Pan has intricate "valle"/condition and
// forced-meld rules; this implementation deliberately drops them:
//
//   - Deck: 8 decks × 40 cards = 320 cards. Each 40-card deck is a standard 52
//     minus the 8s, 9s, 10s (ranks A,2,3,4,5,6,7,J,Q,K per suit).
//   - Players: 3–6. Each player is dealt 10 cards.
//   - Melds are laid to the player's OWN table area (not shared):
//     a set = 3+ cards of the SAME rank (duplicate suits allowed on the big
//     multi-deck); a rope/run = 3+ consecutive cards of the SAME suit (Ace low
//     or high, no wrap). No wilds.
//   - Turn: draw one from stock OR take the discard top → optionally form new
//     melds and/or lay off onto existing melds → discard one card to end the
//     turn. Melding is OPTIONAL (NO forced-meld rule).
//   - Chip conditions (SIMPLIFIED): when a player lays down a qualifying NEW
//     meld, every other player immediately pays chips:
//   - a set of a "valle" rank (3, 5 or 7) → 1 chip each;
//   - a rope/run OR a set of 4+ cards      → 1 chip each.
//     (A meld may satisfy both — e.g. four 5s pays 2 chips each.)
//   - Win: the first player to have melded a total of 11 cards declares "Pan"
//     and wins the round. Each opponent then scores the pip-points still in
//     hand (Ace=1, faces=10, else pip). The Pan declarer scores 0. Cumulative
//     points are tracked; after TargetRounds the lowest cumulative wins (chips
//     are a secondary readout). Mirrors the Indian Rummy points model.
package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// PanHandSize 各プレイヤーの初期手札枚数
const PanHandSize = 10

// PanWinMeldCount この枚数を場に出したら「パン」あがり
const PanWinMeldCount = 11

// PanMeldMin メルド（セット／ラン）の最小枚数
const PanMeldMin = 3

// PanDefaultTargetRounds 既定のラウンド数
const PanDefaultTargetRounds = 3

// PanDeckCount 使用するデッキ数
const PanDeckCount = 8

// PanDeckSize 総カード枚数（8 デッキ × 40 枚）
const PanDeckSize = PanDeckCount * 40

// panDeckValues Pan の 40 枚デッキに含まれるランク（8,9,10 を抜いた A,2-7,J,Q,K）。
var panDeckValues = []int{1, 2, 3, 4, 5, 6, 7, 11, 12, 13}

// panValleRanks チップ条件となる「バジェ（valle）」ランク。
var panValleRanks = map[int]bool{3: true, 5: true, 7: true}

// PanPhase ゲームフェーズ
type PanPhase int

// Panguingue のフェーズ定数
const (
	// PanPhaseDraw ドローフェーズ（山札 or 捨て札トップから 1 枚引く）
	PanPhaseDraw PanPhase = 0
	// PanPhasePlay メルド／レイオフ／ディスカード可能フェーズ
	PanPhasePlay PanPhase = 1
	// PanPhaseRoundEnd ラウンド終了フェーズ
	PanPhaseRoundEnd PanPhase = 2
	// PanPhaseGameEnd ゲーム終了フェーズ
	PanPhaseGameEnd PanPhase = 3
)

// newPanDeck パングインゲ用 320 枚デッキ（8 デッキ × 40 枚）を構築する。
// 各 40 枚デッキは標準 52 枚から 8・9・10 を抜いた A,2-7,J,Q,K × 4 スート。
func newPanDeck() *TrumpCards {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	t := new(TrumpCards)
	t.deckCnt = PanDeckSize
	t.deck = make([]*Card, 0, PanDeckSize)
	for range PanDeckCount {
		for _, suit := range suits {
			for _, val := range panDeckValues {
				t.deck = append(t.deck, NewCard(suit, val, false))
			}
		}
	}
	t.deckInit()
	return t
}

// buildPanPlayers n 人分のプレイヤー（席 0 が人間、残りが CPU）を生成する
func buildPanPlayers(n int) []*PanPlayer {
	if n < PanPlayerCountMin {
		n = PanPlayerCountMin
	}
	if n > PanPlayerCountMax {
		n = PanPlayerCountMax
	}
	players := make([]*PanPlayer, n)
	players[0] = NewPanPlayer(true)
	for i := 1; i < n; i++ {
		players[i] = NewPanPlayer(false)
	}
	return players
}

// Pan パングインゲのゲームクラス。
type Pan struct {
	trumpCards       *TrumpCards
	players          []*PanPlayer
	config           PanConfig
	phase            PanPhase
	currentPlayerIdx int
	dealerIdx        int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	scored           bool // ラウンド終了スコアリングが完了したか（フェーズ再入時の二重加算防止）
	panDeclarerIdx   int  // 「パン」を宣言したプレイヤー（-1 = 宣言なし／山切れ）
	actionLogBase
}

// NewPan コンストラクタ
func NewPan(trumpCards *TrumpCards, players []*PanPlayer, config PanConfig) *Pan {
	return &Pan{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		roundNumber:    0,
		panDeclarerIdx: -1,
	}
}

// NewDefaultPan 標準構成（人間 1 + CPU 3、320 枚デッキ、デフォルト設定）でコンストラクトする SSoT。
func NewDefaultPan() *Pan {
	cfg := DefaultPanConfig()
	return NewPan(newPanDeck(), buildPanPlayers(cfg.PlayerCount), cfg)
}

// Reset ゲームを初期化する
func (g *Pan) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.dealerIdx = 0
	g.actionLog = nil
	g.scored = false
	g.panDeclarerIdx = -1

	g.players = buildPanPlayers(g.config.PlayerCount)
	for _, p := range g.players {
		p.SetChips(0)
	}
	g.currentPlayerIdx = (g.dealerIdx + 1) % len(g.players)

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = PanPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Pan) NextRound() {
	if g.phase != PanPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finalizeGameEnd()
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.dealerIdx = (g.dealerIdx + 1) % len(g.players)
	g.currentPlayerIdx = (g.dealerIdx + 1) % len(g.players)
	g.scored = false
	g.panDeclarerIdx = -1

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = PanPhaseDraw
}

// dealInitialCards 各プレイヤーに 10 枚を配り、最初の 1 枚を捨て札に置く
func (g *Pan) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}
	// The deck is already shuffled by trumpCards.Shuffle() in Reset; a second
	// shuffle here only added non-determinism to tests.

	for range PanHandSize {
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
}

// --- Draw ---

// PlayerDrawFromStock 人間プレイヤーが山札から引く
func (g *Pan) PlayerDrawFromStock() error {
	if err := g.guardHumanAction(PanPhaseDraw); err != nil {
		return err
	}
	return g.drawFromStock()
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札トップから引く
func (g *Pan) PlayerDrawFromDiscard() error {
	if err := g.guardHumanAction(PanPhaseDraw); err != nil {
		return err
	}
	return g.drawFromDiscard()
}

func (g *Pan) drawFromStock() error {
	if len(g.drawPile) == 0 {
		g.endRoundStockOut()
		return nil
	}
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
	g.phase = PanPhasePlay
	return nil
}

func (g *Pan) drawFromDiscard() error {
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	g.phase = PanPhasePlay
	return nil
}

// --- Meld / Layoff ---

// PlayerMeld 人間プレイヤーが手札のカードで新しいメルドを場に出す
func (g *Pan) PlayerMeld(cardIndices []int) error {
	if err := g.guardHumanAction(PanPhasePlay); err != nil {
		return err
	}
	return g.executeMeld(g.currentPlayerIdx, cardIndices)
}

// executeMeld メルドを実行する（人間／CPU 共通）
func (g *Pan) executeMeld(playerIdx int, cardIndices []int) error {
	player := g.players[playerIdx]
	if len(cardIndices) < PanMeldMin {
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
	if !PanIsValidMeld(meld) {
		return NewDomainError(ErrInvalidPlay, "有効なメルド（同ランク3枚以上 または 同スート連続3枚以上）ではありません")
	}

	sortedIdx := make([]int, len(cardIndices))
	copy(sortedIdx, cardIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(sortedIdx)))
	for _, idx := range sortedIdx {
		player.RemoveCard(idx)
	}
	player.AddLaidMeld(meld)

	cardsCopy := make([]*Card, len(meld))
	copy(cardsCopy, meld)
	g.appendLog(playerIdx, "meld", fmt.Sprintf("%s melds %s", playerName(g.players, playerIdx), formatCards(meld)), cardsCopy)

	g.payChipConditions(playerIdx, meld)
	g.checkPanDeclaration(playerIdx)
	return nil
}

// PlayerLayoff 人間プレイヤーがレイオフ（既存メルドにカードを追加）する
func (g *Pan) PlayerLayoff(meldOwner, meldIdx, cardIndex int) error {
	if err := g.guardHumanAction(PanPhasePlay); err != nil {
		return err
	}
	return g.executeLayoff(g.currentPlayerIdx, meldOwner, meldIdx, cardIndex)
}

// executeLayoff レイオフを実行する（人間／CPU 共通）
func (g *Pan) executeLayoff(playerIdx, meldOwner, meldIdx, cardIndex int) error {
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
	if !PanCanLayoff(melds[meldIdx], card) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%sはレイオフできません", cardStr(card)))
	}

	owner.AppendToLaidMeld(meldIdx, card)
	player.RemoveCard(cardIndex)

	g.appendLog(playerIdx, "layoff", fmt.Sprintf("%s lays off %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	// レイオフ先の所有者が 11 枚に到達したらその所有者があがる。
	g.checkPanDeclaration(meldOwner)
	return nil
}

// payChipConditions 新しいメルドの条件付き役に応じて各対戦相手が melder にチップを支払う。
func (g *Pan) payChipConditions(playerIdx int, meld []*Card) {
	units := PanMeldChipUnits(meld)
	if units <= 0 {
		return
	}
	opponents := len(g.players) - 1
	if opponents <= 0 {
		return
	}
	for i := range g.players {
		if i != playerIdx {
			g.players[i].AddChips(-units)
		}
	}
	g.players[playerIdx].AddChips(units * opponents)
	g.appendLog(playerIdx, "chips", fmt.Sprintf("%s collects %d chip(s) from each opponent", playerName(g.players, playerIdx), units), nil)
}

// checkPanDeclaration playerIdx が 11 枚を場に出していれば「パン」あがりとしてラウンドを終える。
func (g *Pan) checkPanDeclaration(playerIdx int) {
	if g.panDeclarerIdx >= 0 {
		return
	}
	if g.players[playerIdx].GetMeldedCardCount() < PanWinMeldCount {
		return
	}
	g.panDeclarerIdx = playerIdx
	g.players[playerIdx].SetIsFinished(true)
	g.appendLog(playerIdx, "pan", fmt.Sprintf("%s declares Pan!", playerName(g.players, playerIdx)), nil)
	g.enterRoundEnd()
}

// --- Discard / turn ---

// PlayerDiscard 人間プレイヤーが手札 1 枚を捨ててターンを終了する
func (g *Pan) PlayerDiscard(cardIndex int) error {
	if err := g.guardHumanAction(PanPhasePlay); err != nil {
		return err
	}
	return g.applyDiscard(cardIndex)
}

func (g *Pan) applyDiscard(cardIndex int) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *Pan) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.phase = PanPhaseDraw
}

// guardHumanAction ゲーム終了／フェーズ不一致／非人間手番の共通ガード。
func (g *Pan) guardHumanAction(want PanPhase) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != want {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Pan) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// --- CPU ---

// CpuPlay 現在の手番が CPU の場合にターンを実行する
func (g *Pan) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	switch g.phase {
	case PanPhaseDraw:
		g.cpuDraw()
	case PanPhasePlay:
		g.cpuPlay()
	}
}

// cpuDraw CPU の引き処理。捨て札トップがすぐメルドできるなら拾い、そうでなければ山札から引く。
func (g *Pan) cpuDraw() {
	top := g.GetDiscardTop()
	if top != nil && g.config.CpuDifficulty != PanCpuDifficultyEasy && g.cpuShouldTakeDiscard(top) {
		_ = g.drawFromDiscard()
		return
	}
	_ = g.drawFromStock()
}

// cpuShouldTakeDiscard 捨て札トップを拾うと即メルドできるか（手札に同ランク 2 枚以上）。
func (g *Pan) cpuShouldTakeDiscard(top *Card) bool {
	player := g.players[g.currentPlayerIdx]
	same := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetValue() == top.GetValue() {
			same++
		}
	}
	return same >= 2
}

// cpuPlay CPU のメルド／レイオフ／ディスカード処理。
func (g *Pan) cpuPlay() {
	idx := g.currentPlayerIdx

	for g.panDeclarerIdx < 0 && g.cpuTryMeld(idx) {
	}
	for g.panDeclarerIdx < 0 && g.cpuTryLayoff(idx) {
	}
	if g.panDeclarerIdx >= 0 {
		return
	}

	player := g.players[idx]
	if player.GetCardsSize() == 0 {
		// 通常は手番開始時に手札 11 枚あるので発生しないが、安全側で山札から補充。
		if len(g.drawPile) > 0 {
			c := g.drawPile[len(g.drawPile)-1]
			g.drawPile = g.drawPile[:len(g.drawPile)-1]
			player.AddCard(c)
		} else {
			g.endRoundStockOut()
			return
		}
	}
	discardIdx := g.cpuChooseDiscard(idx)
	_ = g.applyDiscard(discardIdx)
}

// cpuTryMeld 手札から有効なメルドを 1 つ見つけて場に出す。出せたら true。
func (g *Pan) cpuTryMeld(playerIdx int) bool {
	player := g.players[playerIdx]
	n := player.GetCardsSize()
	cards := make([]*Card, n)
	for i := 0; i < n; i++ {
		cards[i] = player.GetCard(i)
	}
	candidate := findFirstPanMeld(cards)
	if candidate == nil {
		return false
	}
	indices := indicesFor(cards, candidate)
	if indices == nil {
		return false
	}
	return g.executeMeld(playerIdx, indices) == nil
}

// cpuTryLayoff 手札から 1 枚レイオフ可能なカードを探してレイオフする。
func (g *Pan) cpuTryLayoff(playerIdx int) bool {
	player := g.players[playerIdx]
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		for owner := 0; owner < len(g.players); owner++ {
			melds := g.players[owner].GetLaidMelds()
			for mi, meld := range melds {
				if PanCanLayoff(meld, card) {
					if g.executeLayoff(playerIdx, owner, mi, i) == nil {
						return true
					}
				}
			}
		}
	}
	return false
}

// cpuChooseDiscard CPU の捨てカード選択（手札に残る最大ピップ点のカード）。
func (g *Pan) cpuChooseDiscard(playerIdx int) int {
	player := g.players[playerIdx]
	if player.GetCardsSize() == 0 {
		return 0
	}
	bestIdx := 0
	bestVal := panCardPoints(player.GetCard(0))
	for i := 1; i < player.GetCardsSize(); i++ {
		v := panCardPoints(player.GetCard(i))
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

// --- Round / game end ---

// enterRoundEnd ラウンド終了処理をフェーズ再入で 1 度だけ実行する（scored ガード）。
func (g *Pan) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.scoreRound()
	if g.roundNumber >= g.config.TargetRounds {
		g.finalizeGameEnd()
		return
	}
	g.phase = PanPhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了（パンなし・全員手札点採点）。
func (g *Pan) endRoundStockOut() {
	g.panDeclarerIdx = -1
	g.appendLog(-1, "stock_out", "Round ends (stock exhausted)", nil)
	g.enterRoundEnd()
}

// scoreRound ラウンドのスコアを確定する。
func (g *Pan) scoreRound() {
	for i := range g.players {
		var s int
		if g.panDeclarerIdx >= 0 && i == g.panDeclarerIdx {
			s = 0
		} else {
			s = PanHandPoints(g.players[i])
		}
		g.players[i].SetRoundScore(s)
	}

	if g.panDeclarerIdx >= 0 {
		g.appendLog(g.panDeclarerIdx, "round_win", fmt.Sprintf("%s wins the round with Pan", playerName(g.players, g.panDeclarerIdx)), nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}
}

// finalizeGameEnd ゲーム終了処理（累計最少のプレイヤーが勝者）。
func (g *Pan) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = PanPhaseGameEnd

	minScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() < minScore {
			minScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game with %d points!", playerName(g.players, g.winnerIdx), minScore), nil)
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *Pan) GetPhase() PanPhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *Pan) SetPhase(p PanPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号
func (g *Pan) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定（テスト用）
func (g *Pan) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *Pan) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *Pan) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetDealerIdx ディーラーインデックスを取得
func (g *Pan) GetDealerIdx() int { return g.dealerIdx }

// GetDiscardPile 捨て札の山
func (g *Pan) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *Pan) SetDiscardPile(p []*Card) { g.discardPile = p }

// GetDiscardTop 捨て札トップ
func (g *Pan) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札残り枚数
func (g *Pan) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *Pan) SetDrawPile(p []*Card) { g.drawPile = p }

// GetGameEndFlag ゲーム終了フラグ
func (g *Pan) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス（-1 未確定）
func (g *Pan) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *Pan) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Pan) GetPlayer(i int) *PanPlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *Pan) GetConfig() PanConfig { return g.config }

// SetConfig 設定変更
func (g *Pan) SetConfig(c PanConfig) { g.config = c }

// GetTargetRounds ゲーム終了までのラウンド数
func (g *Pan) GetTargetRounds() int { return g.config.TargetRounds }

// GetPanDeclarerIdx 「パン」を宣言したプレイヤー（-1 = 宣言なし）
func (g *Pan) GetPanDeclarerIdx() int { return g.panDeclarerIdx }

// SetPanDeclarerIdx 宣言プレイヤー設定（テスト用）
func (g *Pan) SetPanDeclarerIdx(i int) { g.panDeclarerIdx = i }

// PlayerHandPoints プレイヤー i の手札ピップ点。
func (g *Pan) PlayerHandPoints(i int) int {
	p := g.GetPlayer(i)
	if p == nil {
		return 0
	}
	return PanHandPoints(p)
}

// PlayerMeldedCount プレイヤー i が場に出したカード枚数。
func (g *Pan) PlayerMeldedCount(i int) int {
	p := g.GetPlayer(i)
	if p == nil {
		return 0
	}
	return p.GetMeldedCardCount()
}

// --- Private helpers ---

func (g *Pan) sortAllHands() {
	sortHands(len(g.players), g)
}

func (g *Pan) sortHand(playerIdx int) {
	sortPlayerHand(g.players[playerIdx], bySuitThenValue)
}

// --- Scoring / meld helpers ---

// panCardPoints カードのピップ点（A=1、J/Q/K=10、2-7=フェイスバリュー）。
func panCardPoints(c *Card) int {
	v := c.GetValue()
	if v == 1 {
		return 1
	}
	if v >= 11 {
		return 10
	}
	return v
}

// PanHandPoints プレイヤーの手札に残ったピップ点の合計を返す。
func PanHandPoints(p *PanPlayer) int {
	total := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		total += panCardPoints(p.GetCard(i))
	}
	return total
}

// PanIsValidMeld 有効なメルド（セット または ラン）かどうか。
func PanIsValidMeld(cards []*Card) bool {
	if len(cards) < PanMeldMin {
		return false
	}
	return panIsValidSet(cards) || panIsValidRun(cards)
}

// panIsValidSet 同ランク 3 枚以上（マルチデッキのため同スート重複を許容）。
func panIsValidSet(cards []*Card) bool {
	if len(cards) < PanMeldMin {
		return false
	}
	v := cards[0].GetValue()
	for _, c := range cards[1:] {
		if c.GetValue() != v {
			return false
		}
	}
	return true
}

// panIsValidRun 同スートで連続する 3 枚以上（Ace は low または high、ラップ不可）。
// panRankOrder maps a card value to its position in Pan's reduced deck
// (ranks A,2,3,4,5,6,7,J,Q,K — the 8s, 9s and 10s are removed), so that 7 and
// J are adjacent. A run's consecutiveness must be judged on these indices, not
// on raw card values (7→11 looks like a gap but is a valid Pan sequence).
var panRankOrder = map[int]int{1: 0, 2: 1, 3: 2, 4: 3, 5: 4, 6: 5, 7: 6, 11: 7, 12: 8, 13: 9}

// panAceHighRank is the Ace's index when it ranks above the King (Q-K-A).
const panAceHighRank = 10

// panRankIndex returns the Pan rank index for a card value, or -1 if the value
// is not part of the Pan deck (i.e. an 8, 9 or 10).
func panRankIndex(value int) int {
	if idx, ok := panRankOrder[value]; ok {
		return idx
	}
	return -1
}

func panIsValidRun(cards []*Card) bool {
	if len(cards) < PanMeldMin {
		return false
	}
	suit := cards[0].GetDesign()
	for _, c := range cards[1:] {
		if c.GetDesign() != suit {
			return false
		}
	}
	idx := make([]int, len(cards))
	for i, c := range cards {
		ri := panRankIndex(c.GetValue())
		if ri < 0 {
			return false
		}
		idx[i] = ri
	}
	sort.Ints(idx)
	for i := 1; i < len(idx); i++ {
		if idx[i] == idx[i-1] {
			return false // 連続ランに重複ランクは不可
		}
	}
	if isConsecutive(idx) {
		return true
	}
	// Ace を high(K の上) として再評価（Q-K-A など）。
	if idx[0] == 0 {
		high := make([]int, len(idx))
		copy(high, idx)
		high[0] = panAceHighRank
		sort.Ints(high)
		if isConsecutive(high) {
			return true
		}
	}
	return false
}

// PanCanLayoff 既存メルドに card をレイオフできるかどうか。
func PanCanLayoff(meld []*Card, card *Card) bool {
	if len(meld) == 0 {
		return false
	}
	candidate := append([]*Card(nil), meld...)
	candidate = append(candidate, card)
	if panIsValidSet(meld) {
		return panIsValidSet(candidate)
	}
	return panIsValidRun(candidate)
}

// PanIsValleMeld reports whether the meld is a valle (バジェ): a set of 3s, 5s
// or 7s. Laying one pays every player a chip, but neither UI said which meld on
// the table caused the chip column to move (#4853).
func PanIsValleMeld(meld []*Card) bool {
	// panIsValidSet が真なら PanIsValidMeld も真 (set か run の OR) なので重ねない。
	return panIsValidSet(meld) && panValleRanks[meld[0].GetValue()]
}

// PanMeldChipUnits メルドが満たすチップ条件の数を返す（0〜2）。
//   - バジェランク（3,5,7）のセット → +1
//   - ラン、または 4 枚以上のセット   → +1
func PanMeldChipUnits(meld []*Card) int {
	if !PanIsValidMeld(meld) {
		return 0
	}
	units := 0
	isSetMeld := panIsValidSet(meld)
	if PanIsValleMeld(meld) {
		units++
	}
	if !isSetMeld || len(meld) >= 4 {
		// ラン、または 4 枚以上のセット
		units++
	}
	return units
}

// findFirstPanMeld 手札から最初に見つかった有効メルド（セット優先、次いでラン）を返す。無ければ nil。
func findFirstPanMeld(cards []*Card) []*Card {
	// セット候補（同ランク 3 枚）
	byRank := make(map[int][]*Card)
	order := make([]int, 0)
	for _, c := range cards {
		v := c.GetValue()
		if _, ok := byRank[v]; !ok {
			order = append(order, v)
		}
		byRank[v] = append(byRank[v], c)
	}
	sort.Ints(order)
	for _, v := range order {
		if len(byRank[v]) >= PanMeldMin {
			return append([]*Card(nil), byRank[v][:PanMeldMin]...)
		}
	}

	// ラン候補（同スート連続 3 枚）
	bySuit := make(map[int][]*Card)
	suitOrder := make([]int, 0)
	for _, c := range cards {
		s := c.GetDesign()
		if _, ok := bySuit[s]; !ok {
			suitOrder = append(suitOrder, s)
		}
		bySuit[s] = append(bySuit[s], c)
	}
	sort.Ints(suitOrder)
	for _, s := range suitOrder {
		group := bySuit[s]
		sorted := make([]*Card, len(group))
		copy(sorted, group)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].GetValue() < sorted[j].GetValue() })
		// 値の重複を除いた昇順列で連続 3 枚を探す。
		uniq := make([]*Card, 0, len(sorted))
		for i, c := range sorted {
			if i == 0 || c.GetValue() != sorted[i-1].GetValue() {
				uniq = append(uniq, c)
			}
		}
		for i := 0; i+PanMeldMin <= len(uniq); i++ {
			run := []*Card{uniq[i]}
			for j := i + 1; j < len(uniq); j++ {
				if panRankIndex(uniq[j].GetValue()) == panRankIndex(run[len(run)-1].GetValue())+1 {
					run = append(run, uniq[j])
				} else {
					break
				}
			}
			if len(run) >= PanMeldMin {
				return append([]*Card(nil), run[:PanMeldMin]...)
			}
		}
	}
	return nil
}

// --- JSON ---

// panJSON は Pan の JSON 表現。
type panJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*PanPlayer      `json:"pl"`
	Config           PanConfig         `json:"cf"`
	Phase            PanPhase          `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DealerIdx        int               `json:"di"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	Scored           bool              `json:"sc"`
	PanDeclarerIdx   int               `json:"pd"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Pan) MarshalJSON() ([]byte, error) {
	return json.Marshal(panJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DealerIdx:        g.dealerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		Scored:           g.scored,
		PanDeclarerIdx:   g.panDeclarerIdx,
		ActionLog:        g.actionLog,
	})
}

const panMaxSliceLen = 2000

// UnmarshalJSON implements json.Unmarshaler. KV 復元時に全インデックス・要素を検証し、
// 範囲外の値や nil 要素が後段でパニックすることを防ぐ。
func (g *Pan) UnmarshalJSON(data []byte) error {
	var j panJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > panMaxSliceLen || len(j.DiscardPile) > panMaxSliceLen ||
		len(j.DrawPile) > panMaxSliceLen || len(j.ActionLog) > panMaxSliceLen {
		return fmt.Errorf("pan: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newPanDeck()
	}

	g.players = j.Players
	if g.players == nil {
		g.players = make([]*PanPlayer, 0)
	}
	for i, p := range g.players {
		if p == nil {
			return fmt.Errorf("pan: player %d is nil", i)
		}
	}
	n := len(g.players)
	if n < PanPlayerCountMin || n > PanPlayerCountMax {
		return fmt.Errorf("pan: invalid player count %d", n)
	}

	g.config = j.Config
	if g.config.PlayerCount <= 0 {
		g.config.PlayerCount = n
	}
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("pan: invalid config: %w", err)
	}

	if j.Phase < PanPhaseDraw || j.Phase > PanPhaseGameEnd {
		return fmt.Errorf("pan: invalid phase %d", j.Phase)
	}
	g.phase = j.Phase

	if j.RoundNumber < 0 {
		return fmt.Errorf("pan: invalid round number %d", j.RoundNumber)
	}
	g.roundNumber = j.RoundNumber

	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= n {
		return fmt.Errorf("pan: currentPlayerIdx %d out of range", j.CurrentPlayerIdx)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n {
		return fmt.Errorf("pan: dealerIdx %d out of range", j.DealerIdx)
	}
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.dealerIdx = j.DealerIdx

	if err := panValidateSentinelIdx("winnerIdx", j.WinnerIdx, n); err != nil {
		return err
	}
	if err := panValidateSentinelIdx("panDeclarerIdx", j.PanDeclarerIdx, n); err != nil {
		return err
	}
	g.winnerIdx = j.WinnerIdx
	g.panDeclarerIdx = j.PanDeclarerIdx

	g.gameEndFlag = j.GameEndFlag
	g.scored = j.Scored

	g.discardPile = panFilterNilCards(j.DiscardPile)
	g.drawPile = panFilterNilCards(j.DrawPile)

	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// panValidateSentinelIdx は -1（未確定センチネル）または [0, n) の範囲を許容する。
func panValidateSentinelIdx(name string, idx, n int) error {
	if idx == -1 {
		return nil
	}
	if idx < 0 || idx >= n {
		return fmt.Errorf("pan: %s %d out of range", name, idx)
	}
	return nil
}

// panFilterNilCards nil 要素を除いた新しいスライスを返す（nil 入力には空スライス）。
func panFilterNilCards(cards []*Card) []*Card {
	out := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}
