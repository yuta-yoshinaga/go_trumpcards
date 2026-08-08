//go:build !js || !wasm || extra

// Package domain: Machiavelli (マキャヴェッリ) implementation.
//
// Machiavelli is an Italian rummy — Rummikub with cards — where ALL melds live
// on a single SHARED TABLE. On your turn you may either draw one card from the
// stock (ending the turn) OR freely rebuild the entire table (moving cards
// between melds) as long as every resulting meld is valid and you add at least
// one card from your hand.
//
// Clean, deterministic interpretation implemented here:
//   - Deck: 2 standard 52-card decks = 104 cards, NO jokers/wilds.
//   - Players: 2–5 (default 4 = 1 human at seat 0 + CPUs). Each is dealt 13
//     cards; the remaining 52 cards form the stock. There is no discard pile —
//     you draw if you cannot or do not want to play.
//   - Melds:
//   - Set: 3+ cards of the same rank, all of DIFFERENT suits. Because suits
//     must be distinct within one set, a set holds at most 4 cards (suits
//     may repeat across separate melds in this two-deck game).
//   - Run: 3+ consecutive cards of the same suit (Ace low OR high, no wrap).
//   - Turn: draw one card (ends the turn, no meld) OR submit a proposed NEW
//     COMPLETE TABLE STATE plus the hand-card indices being added. The play is
//     accepted iff (a) every meld in the new table is a valid set/run of size
//     >=3, (b) the multiset of cards in the new table equals the previous
//     table's cards PLUS exactly the played hand cards (nothing invented or
//     dropped), and (c) at least one hand card was added.
//   - Win: the first player to empty their hand wins the round; every other
//     player scores the pip points still in hand (Ace=1, J/Q/K/10=10, else pip).
//     Lowest cumulative over TargetRounds wins.
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// MachiavelliHandSize 各プレイヤーの初期手札枚数
const MachiavelliHandSize = 13

// MachiavelliMeldMin メルド（セット／ラン）の最小枚数
const MachiavelliMeldMin = 3

// MachiavelliDefaultTargetRounds 既定のラウンド数
const MachiavelliDefaultTargetRounds = 3

// MachiavelliPhase ゲームフェーズ
type MachiavelliPhase int

// Machiavelli のフェーズ定数
const (
	// MachiavelliPhaseTurn 手番フェーズ（山札から引く or 場を再構築してプレイする）
	MachiavelliPhaseTurn MachiavelliPhase = 0
	// MachiavelliPhaseRoundEnd ラウンド終了フェーズ
	MachiavelliPhaseRoundEnd MachiavelliPhase = 1
	// MachiavelliPhaseGameEnd ゲーム終了フェーズ
	MachiavelliPhaseGameEnd MachiavelliPhase = 2
)

// MachiavelliCardRef は「新しい場」を指定するためのカード参照（デザイン＋数値）。
// 2 デッキ運用ではカードは (design, value) の値として等価に扱えるため、物理的な
// カードの同一性は不要で、多重集合の一致で保存則を検証できる。
type MachiavelliCardRef struct {
	// Design スート（1=Spade..4=Diamond）
	Design int `json:"design"`
	// Value ランク（1=Ace..13=King）
	Value int `json:"value"`
}

// newMachiavelliDeck マキャヴェッリ用 104 枚デッキ（標準 52 枚デッキ×2、ジョーカーなし）を構築する。
func newMachiavelliDeck() *TrumpCards {
	return NewTrumpCardsWithDecks(2, 0)
}

// buildMachiavelliPlayers n 人分のプレイヤー（席 0 が人間、残りが CPU）を生成する
func buildMachiavelliPlayers(n int) []*MachiavelliPlayer {
	if n < MachiavelliPlayerCountMin {
		n = MachiavelliPlayerCountMin
	}
	if n > MachiavelliPlayerCountMax {
		n = MachiavelliPlayerCountMax
	}
	players := make([]*MachiavelliPlayer, n)
	players[0] = NewMachiavelliPlayer(true)
	for i := 1; i < n; i++ {
		players[i] = NewMachiavelliPlayer(false)
	}
	return players
}

// Machiavelli マキャヴェッリ（共有テーブル式イタリアンラミー）のゲームクラス。
type Machiavelli struct {
	trumpCards       *TrumpCards
	players          []*MachiavelliPlayer
	config           MachiavelliConfig
	phase            MachiavelliPhase
	currentPlayerIdx int
	dealerIdx        int
	table            [][]*Card // 共有テーブル上のメルド群
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	scored           bool // ラウンド終了スコアリングが完了したか（フェーズ再入時の二重加算防止）
	roundWinnerIdx   int  // 直近ラウンドの勝者（-1 = 山切れ流局）
	actionLogBase
}

// NewMachiavelli コンストラクタ
func NewMachiavelli(trumpCards *TrumpCards, players []*MachiavelliPlayer, config MachiavelliConfig) *Machiavelli {
	return &Machiavelli{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		roundNumber:    0,
		roundWinnerIdx: -1,
	}
}

// NewDefaultMachiavelli 標準構成（人間 1 + CPU 3、104 枚デッキ、デフォルト設定）でコンストラクトする SSoT。
func NewDefaultMachiavelli() *Machiavelli {
	cfg := DefaultMachiavelliConfig()
	return NewMachiavelli(newMachiavelliDeck(), buildMachiavelliPlayers(cfg.PlayerCount), cfg)
}

// Reset ゲームを初期化する
func (g *Machiavelli) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.table = nil
	g.drawPile = nil
	g.dealerIdx = 0
	g.actionLog = nil
	g.scored = false
	g.roundWinnerIdx = -1

	g.players = buildMachiavelliPlayers(g.config.PlayerCount)
	g.currentPlayerIdx = (g.dealerIdx + 1) % len(g.players)

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = MachiavelliPhaseTurn
}

// NextRound 次のラウンドを開始する
func (g *Machiavelli) NextRound() {
	if g.phase != MachiavelliPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finalizeGameEnd()
		return
	}

	g.roundNumber++
	g.table = nil
	g.drawPile = nil
	g.dealerIdx = (g.dealerIdx + 1) % len(g.players)
	g.currentPlayerIdx = (g.dealerIdx + 1) % len(g.players)
	g.scored = false
	g.roundWinnerIdx = -1

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = MachiavelliPhaseTurn
}

// dealInitialCards 各プレイヤーに 13 枚を配り、残りを山札にする（捨て札・ワイルドなし）。
func (g *Machiavelli) dealInitialCards() {
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

	for range MachiavelliHandSize {
		for j := range len(g.players) {
			if len(g.drawPile) == 0 {
				break
			}
			card := g.drawPile[len(g.drawPile)-1]
			g.drawPile = g.drawPile[:len(g.drawPile)-1]
			g.players[j].AddCard(card)
		}
	}
}

// PlayerDraw 人間プレイヤーが山札から 1 枚引く（ターン終了）。
func (g *Machiavelli) PlayerDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MachiavelliPhaseTurn {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromStock()
}

func (g *Machiavelli) drawFromStock() error {
	if len(g.drawPile) == 0 {
		g.endRoundStockOut()
		return nil
	}
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw", fmt.Sprintf("%s draws from stock", g.playerName(g.currentPlayerIdx)), nil)
	g.advanceTurn()
	return nil
}

// PlayerPlay 人間プレイヤーが「新しい場（メルド群）」と追加する手札インデックスを提出する。
func (g *Machiavelli) PlayerPlay(refs [][]MachiavelliCardRef, handIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MachiavelliPhaseTurn {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	newTable, err := machiavelliRefsToTable(refs)
	if err != nil {
		return err
	}
	return g.applyPlay(newTable, handIndices)
}

// PlayerNewMeld 人間プレイヤーが手札インデックスから新しいメルドを 1 つ場に出す（再構築なし）。
// PlayerPlay の簡易版で、CUI から使いやすい。
func (g *Machiavelli) PlayerNewMeld(handIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MachiavelliPhaseTurn {
		return ErrWrongPhase
	}
	player := g.players[g.currentPlayerIdx]
	if !player.GetIsHuman() {
		return ErrNotHumanTurn
	}
	if err := validateIndexList(handIndices, player.GetCardsSize()); err != nil {
		return err
	}
	meld := make([]*Card, len(handIndices))
	for i, idx := range handIndices {
		meld[i] = player.GetCard(idx)
	}
	newTable := machiavelliCloneTable(g.table)
	newTable = append(newTable, meld)
	return g.applyPlay(newTable, handIndices)
}

// PlayerLayoff 人間プレイヤーが手札 1 枚を既存メルドに追加する（再構築なし）。
func (g *Machiavelli) PlayerLayoff(meldIdx, handIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MachiavelliPhaseTurn {
		return ErrWrongPhase
	}
	player := g.players[g.currentPlayerIdx]
	if !player.GetIsHuman() {
		return ErrNotHumanTurn
	}
	if meldIdx < 0 || meldIdx >= len(g.table) {
		return NewDomainError(ErrInvalidPlay, "対象メルドが不正です")
	}
	if handIndex < 0 || handIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	newTable := machiavelliCloneTable(g.table)
	newTable[meldIdx] = append(newTable[meldIdx], player.GetCard(handIndex))
	return g.applyPlay(newTable, []int{handIndex})
}

// applyPlay 提出された新しい場を検証し、合格すれば適用する（共通コア）。
func (g *Machiavelli) applyPlay(newTable [][]*Card, handIndices []int) error {
	player := g.players[g.currentPlayerIdx]

	// (c) 少なくとも 1 枚は手札から出す
	if len(handIndices) < 1 {
		return NewDomainError(ErrInvalidPlay, "手札から少なくとも1枚出す必要があります")
	}
	if err := validateIndexList(handIndices, player.GetCardsSize()); err != nil {
		return err
	}

	// (a) 各メルドが有効なセット／ラン (>=3)
	for _, meld := range newTable {
		if !machiavelliIsValidMeld(meld) {
			return NewDomainError(ErrInvalidPlay, "無効なメルドが含まれています")
		}
	}

	// (b) 保存則: newTable == oldTable + 出した手札
	playedCards := make([]*Card, len(handIndices))
	for i, idx := range handIndices {
		playedCards[i] = player.GetCard(idx)
	}
	if !machiavelliConserves(g.table, playedCards, newTable) {
		return NewDomainError(ErrInvalidPlay, "場のカード構成が一致しません")
	}

	// 適用: 出した手札を降順で除去し、場を差し替える
	idxDesc := append([]int(nil), handIndices...)
	sort.Sort(sort.Reverse(sort.IntSlice(idxDesc)))
	for _, idx := range idxDesc {
		player.RemoveCard(idx)
	}
	g.table = newTable

	g.appendLog(g.currentPlayerIdx, "play", fmt.Sprintf("%s plays %d card(s) to the table", g.playerName(g.currentPlayerIdx), len(handIndices)), playedCards)

	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
		return nil
	}
	g.advanceTurn()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *Machiavelli) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.phase = MachiavelliPhaseTurn
}

// IsHumanTurn 現在の手番が人間かどうか（かつゲーム継続中）。
func (g *Machiavelli) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	if g.phase != MachiavelliPhaseTurn {
		return false
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// CpuPlay 現在の手番が CPU の場合にターンを実行する。
func (g *Machiavelli) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.phase != MachiavelliPhaseTurn {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	newTable, handIdx, ok := g.cpuBuildPlay()
	if ok {
		_ = g.applyPlay(newTable, handIdx)
		return
	}
	_ = g.drawFromStock()
}

// cpuBuildPlay CPU の保守的なプレイを構築する（他プレイヤーのカードを並べ替えず、
// 既存メルドへのレイオフと手札からの新規メルドのみ）。追加できた手札がなければ ok=false。
func (g *Machiavelli) cpuBuildPlay() ([][]*Card, []int, bool) {
	player := g.players[g.currentPlayerIdx]
	hand := machiavelliCollectCards(player)
	used := make([]bool, len(hand))
	newTable := machiavelliCloneTable(g.table)
	added := make([]int, 0, len(hand))

	// 1. レイオフ: 各手札を既存メルドに追加できれば追加する（安定するまで反復）。
	changed := true
	for changed {
		changed = false
		for i := range hand {
			if used[i] {
				continue
			}
			for m := range newTable {
				trial := append(append([]*Card(nil), newTable[m]...), hand[i])
				if machiavelliIsValidMeld(trial) {
					newTable[m] = trial
					used[i] = true
					added = append(added, i)
					changed = true
					break
				}
			}
		}
	}

	// 2. 手札から新しいセットを作る（同ランク別スート 3 枚以上）。
	byRank := make(map[int][]int)
	for i, c := range hand {
		if !used[i] {
			byRank[c.GetValue()] = append(byRank[c.GetValue()], i)
		}
	}
	for _, idxs := range byRank {
		distinct := machiavelliDistinctSuitIdx(hand, idxs)
		if len(distinct) >= MachiavelliMeldMin {
			meld := make([]*Card, 0, len(distinct))
			for _, i := range distinct {
				meld = append(meld, hand[i])
				used[i] = true
				added = append(added, i)
			}
			newTable = append(newTable, meld)
		}
	}

	// 3. 手札から新しいランを作る（同スート連続 3 枚以上）。
	for _, meldIdx := range g.cpuFindRuns(hand, used) {
		meld := make([]*Card, 0, len(meldIdx))
		for _, i := range meldIdx {
			meld = append(meld, hand[i])
			used[i] = true
			added = append(added, i)
		}
		newTable = append(newTable, meld)
	}

	if len(added) == 0 {
		return nil, nil, false
	}
	return newTable, added, true
}

// machiavelliDistinctSuitIdx idxs の中からスートが相異なるインデックスを最大 4 枚選ぶ。
func machiavelliDistinctSuitIdx(hand []*Card, idxs []int) []int {
	seen := make(map[int]bool)
	out := make([]int, 0, 4)
	for _, i := range idxs {
		s := hand[i].GetDesign()
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, i)
	}
	return out
}

// cpuFindRuns 未使用の手札から同スート連続 3+ 枚のランを見つけてインデックス集合を返す。
func (g *Machiavelli) cpuFindRuns(hand []*Card, used []bool) [][]int {
	bySuit := make(map[int]map[int]int) // suit → value → hand index
	for i, c := range hand {
		if used[i] {
			continue
		}
		s := c.GetDesign()
		if bySuit[s] == nil {
			bySuit[s] = make(map[int]int)
		}
		if _, ok := bySuit[s][c.GetValue()]; !ok {
			bySuit[s][c.GetValue()] = i
		}
	}
	runs := make([][]int, 0)
	consumed := make(map[int]bool)
	for _, byVal := range bySuit {
		values := make([]int, 0, len(byVal))
		for v := range byVal {
			values = append(values, v)
		}
		sort.Ints(values)
		run := make([]int, 0)
		flush := func() {
			if len(run) >= MachiavelliMeldMin {
				runs = append(runs, append([]int(nil), run...))
			}
			run = run[:0]
		}
		for i, v := range values {
			if i > 0 && v != values[i-1]+1 {
				flush()
			}
			run = append(run, byVal[v])
		}
		flush()
	}
	// 二重使用しないよう連続ラン内のインデックスを記録する（マップは決定的でないため保険）。
	final := make([][]int, 0, len(runs))
	for _, r := range runs {
		ok := true
		for _, i := range r {
			if consumed[i] {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, i := range r {
			consumed[i] = true
		}
		final = append(final, r)
	}
	return final
}

// finishRound 上がり／山切れの最終スコアリング（scored ガードで 1 度だけ）。
func (g *Machiavelli) finishRound(winnerIdx int) {
	if g.scored {
		return
	}
	g.scored = true
	g.roundWinnerIdx = winnerIdx

	for i := range g.players {
		s := 0
		if winnerIdx < 0 || i != winnerIdx {
			s = machiavelliDeadwood(g.players[i])
		}
		g.players[i].SetRoundScore(s)
	}

	if winnerIdx >= 0 {
		g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s goes out (round %d)", g.playerName(winnerIdx), g.roundNumber), nil)
	} else {
		g.appendLog(-1, "draw", "Round ends (stock exhausted)", nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	if g.roundNumber >= g.config.TargetRounds {
		g.finalizeGameEnd()
		return
	}
	g.phase = MachiavelliPhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了（勝者なし・全員デッドウッド採点）。
func (g *Machiavelli) endRoundStockOut() {
	g.finishRound(-1)
}

// finalizeGameEnd ゲーム終了処理（累計最少のプレイヤーが勝者）。
func (g *Machiavelli) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = MachiavelliPhaseGameEnd

	minScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() < minScore {
			minScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game with %d points!", g.playerName(g.winnerIdx), minScore), nil)
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *Machiavelli) GetPhase() MachiavelliPhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *Machiavelli) SetPhase(p MachiavelliPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号
func (g *Machiavelli) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定（テスト用）
func (g *Machiavelli) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *Machiavelli) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *Machiavelli) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetDealerIdx ディーラーインデックス
func (g *Machiavelli) GetDealerIdx() int { return g.dealerIdx }

// GetTable 共有テーブル上のメルド群を取得
func (g *Machiavelli) GetTable() [][]*Card { return g.table }

// SetTable テーブル設定（テスト用）
func (g *Machiavelli) SetTable(t [][]*Card) { g.table = t }

// GetDrawPileCount 山札残り枚数
func (g *Machiavelli) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *Machiavelli) SetDrawPile(p []*Card) { g.drawPile = p }

// GetGameEndFlag ゲーム終了フラグ
func (g *Machiavelli) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス（-1 未確定）
func (g *Machiavelli) GetWinnerIdx() int { return g.winnerIdx }

// GetRoundWinnerIdx 直近ラウンドの勝者（-1 = 山切れ）
func (g *Machiavelli) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetPlayerCnt プレイヤー数
func (g *Machiavelli) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Machiavelli) GetPlayer(i int) *MachiavelliPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetConfig 設定取得
func (g *Machiavelli) GetConfig() MachiavelliConfig { return g.config }

// SetConfig 設定変更
func (g *Machiavelli) SetConfig(c MachiavelliConfig) { g.config = c }

// GetTargetRounds ゲーム終了までのラウンド数
func (g *Machiavelli) GetTargetRounds() int { return g.config.TargetRounds }

// PlayerDeadwoodValue プレイヤー i の手札デッドウッド点。
func (g *Machiavelli) PlayerDeadwoodValue(i int) int {
	p := g.GetPlayer(i)
	if p == nil {
		return 0
	}
	return machiavelliDeadwood(p)
}

// --- Private helpers ---

func (g *Machiavelli) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

func (g *Machiavelli) sortHand(playerIdx int) {
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

func (g *Machiavelli) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// machiavelliCollectCards プレイヤーの手札を []*Card で返す
func machiavelliCollectCards(p *MachiavelliPlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// machiavelliDeadwood プレイヤーの手札残りのピップ点合計。
func machiavelliDeadwood(p *MachiavelliPlayer) int {
	total := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		total += machiavelliCardPoints(p.GetCard(i))
	}
	return total
}

// machiavelliCardPoints カードのピップ点（Ace=1、10/J/Q/K=10、それ以外は数値）。
func machiavelliCardPoints(card *Card) int {
	if card == nil {
		return 0
	}
	v := card.GetValue()
	if v >= 10 {
		return 10
	}
	return v
}

// MachiavelliCardPoints はデッドウッド計算に使うカード点を返す（外部公開用）。
func MachiavelliCardPoints(card *Card) int {
	return machiavelliCardPoints(card)
}

// machiavelliCloneTable テーブルを（メルドとスライスを）浅い要素コピーで複製する。
func machiavelliCloneTable(table [][]*Card) [][]*Card {
	out := make([][]*Card, len(table))
	for i, meld := range table {
		out[i] = append([]*Card(nil), meld...)
	}
	return out
}

// machiavelliRefsToTable カード参照の集合を []*Card のテーブルへ変換する。
func machiavelliRefsToTable(refs [][]MachiavelliCardRef) ([][]*Card, error) {
	table := make([][]*Card, len(refs))
	for i, meldRefs := range refs {
		if len(meldRefs) == 0 {
			return nil, NewDomainError(ErrInvalidPlay, "空のメルドは指定できません")
		}
		meld := make([]*Card, len(meldRefs))
		for j, r := range meldRefs {
			if r.Design < CardDesignSpade || r.Design > CardDesignDiamond || r.Value < 1 || r.Value > CardValueMax {
				return nil, NewDomainError(ErrInvalidCard, "無効なカードが指定されました")
			}
			meld[j] = NewCard(r.Design, r.Value, false)
		}
		table[i] = meld
	}
	return table, nil
}

// machiavelliCardKey は (design, value) を一意な整数キーにする（多重集合用）。
func machiavelliCardKey(c *Card) int {
	return c.GetDesign()*100 + c.GetValue()
}

// machiavelliConserves は保存則を検証する:
// newTable の多重集合 == oldTable の多重集合 + played の多重集合。
func machiavelliConserves(oldTable [][]*Card, played []*Card, newTable [][]*Card) bool {
	want := make(map[int]int)
	for _, meld := range oldTable {
		for _, c := range meld {
			want[machiavelliCardKey(c)]++
		}
	}
	for _, c := range played {
		want[machiavelliCardKey(c)]++
	}
	got := make(map[int]int)
	for _, meld := range newTable {
		for _, c := range meld {
			got[machiavelliCardKey(c)]++
		}
	}
	if len(want) != len(got) {
		return false
	}
	for k, v := range want {
		if got[k] != v {
			return false
		}
	}
	return true
}

// --- Meld validators (joker-less) ---

// MachiavelliIsValidMeld は cards が有効なメルド（セットまたはラン、3+ 枚）か判定する（外部公開）。
func MachiavelliIsValidMeld(cards []*Card) bool {
	return machiavelliIsValidMeld(cards)
}

func machiavelliIsValidMeld(cards []*Card) bool {
	if len(cards) < MachiavelliMeldMin {
		return false
	}
	return machiavelliIsSet(cards) || machiavelliIsRun(cards)
}

// machiavelliIsSet は cards が有効なセット（同ランク・全スート相異、3+ 枚）か判定する。
func machiavelliIsSet(cards []*Card) bool {
	if len(cards) < MachiavelliMeldMin {
		return false
	}
	rank := -1
	seenSuit := make(map[int]bool)
	for _, c := range cards {
		if c == nil || c.GetDesign() == CardDesignJoker {
			return false
		}
		if rank == -1 {
			rank = c.GetValue()
		} else if c.GetValue() != rank {
			return false
		}
		s := c.GetDesign()
		if seenSuit[s] {
			return false // スート重複はセット内では不可
		}
		seenSuit[s] = true
	}
	return true
}

// machiavelliIsRun は cards が有効なラン（同スート連続、3+ 枚、Ace low/high、ラップなし）か判定する。
func machiavelliIsRun(cards []*Card) bool {
	if len(cards) < MachiavelliMeldMin {
		return false
	}
	suit := -1
	values := make([]int, 0, len(cards))
	seen := make(map[int]bool)
	for _, c := range cards {
		if c == nil || c.GetDesign() == CardDesignJoker {
			return false
		}
		if suit == -1 {
			suit = c.GetDesign()
		} else if c.GetDesign() != suit {
			return false
		}
		v := c.GetValue()
		if seen[v] {
			return false
		}
		seen[v] = true
		values = append(values, v)
	}
	// aceVariants は値をコピーしてソートするので、span/連続判定の前に必ずソートされる。
	for _, variant := range aceVariants(values) {
		if isConsecutive(variant) {
			return true
		}
	}
	return false
}

// --- JSON ---

// machiavelliJSON は Machiavelli の JSON 表現。
type machiavelliJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*MachiavelliPlayer `json:"pl"`
	Config           MachiavelliConfig    `json:"cf"`
	Phase            MachiavelliPhase     `json:"ps"`
	CurrentPlayerIdx int                  `json:"ci"`
	DealerIdx        int                  `json:"di"`
	Table            [][]*Card            `json:"tb"`
	DrawPile         []*Card              `json:"wp"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerIdx        int                  `json:"wi"`
	RoundNumber      int                  `json:"rn"`
	Scored           bool                 `json:"sc"`
	RoundWinnerIdx   int                  `json:"rw"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Machiavelli) MarshalJSON() ([]byte, error) {
	return json.Marshal(machiavelliJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DealerIdx:        g.dealerIdx,
		Table:            g.table,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		Scored:           g.scored,
		RoundWinnerIdx:   g.roundWinnerIdx,
		ActionLog:        g.actionLog,
	})
}

const machiavelliMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. KV 復元時に全インデックス・要素を検証し、
// 範囲外の値や nil 要素が後段でパニックすることを防ぐ。
func (g *Machiavelli) UnmarshalJSON(data []byte) error {
	var j machiavelliJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > machiavelliMaxSliceLen || len(j.Table) > machiavelliMaxSliceLen ||
		len(j.DrawPile) > machiavelliMaxSliceLen || len(j.ActionLog) > machiavelliMaxSliceLen {
		return fmt.Errorf("machiavelli: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newMachiavelliDeck()
	}

	// プレイヤー要素の nil を拒否し、人数を検証する。
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*MachiavelliPlayer, 0)
	}
	for i, p := range g.players {
		if p == nil {
			return fmt.Errorf("machiavelli: player %d is nil", i)
		}
	}
	n := len(g.players)
	if n < MachiavelliPlayerCountMin || n > MachiavelliPlayerCountMax {
		return fmt.Errorf("machiavelli: invalid player count %d", n)
	}

	g.config = j.Config
	if g.config.PlayerCount <= 0 {
		g.config.PlayerCount = n
	}
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("machiavelli: invalid config: %w", err)
	}

	if j.Phase < MachiavelliPhaseTurn || j.Phase > MachiavelliPhaseGameEnd {
		return fmt.Errorf("machiavelli: invalid phase %d", j.Phase)
	}
	g.phase = j.Phase

	if j.RoundNumber < 0 {
		return fmt.Errorf("machiavelli: invalid round number %d", j.RoundNumber)
	}
	g.roundNumber = j.RoundNumber

	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= n {
		return fmt.Errorf("machiavelli: currentPlayerIdx %d out of range", j.CurrentPlayerIdx)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n {
		return fmt.Errorf("machiavelli: dealerIdx %d out of range", j.DealerIdx)
	}
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.dealerIdx = j.DealerIdx

	if err := machiavelliValidateSentinelIdx("winnerIdx", j.WinnerIdx, n); err != nil {
		return err
	}
	if err := machiavelliValidateSentinelIdx("roundWinnerIdx", j.RoundWinnerIdx, n); err != nil {
		return err
	}
	g.winnerIdx = j.WinnerIdx
	g.roundWinnerIdx = j.RoundWinnerIdx

	g.gameEndFlag = j.GameEndFlag
	g.scored = j.Scored

	// テーブルの各メルドから nil カードを除去し、nil メルドを捨てる（後段の nil デリファレンス防止）。
	g.table = machiavelliFilterTable(j.Table)
	g.drawPile = machiavelliFilterNilCards(j.DrawPile)

	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// machiavelliValidateSentinelIdx は -1（未確定センチネル）または [0, n) の範囲を許容する。
func machiavelliValidateSentinelIdx(name string, idx, n int) error {
	if idx == -1 {
		return nil
	}
	if idx < 0 || idx >= n {
		return fmt.Errorf("machiavelli: %s %d out of range", name, idx)
	}
	return nil
}

// machiavelliFilterNilCards nil 要素を除いた新しいスライスを返す（nil 入力には空スライス）。
func machiavelliFilterNilCards(cards []*Card) []*Card {
	out := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}

// machiavelliFilterTable 各メルドから nil カードを除き、空になったメルドを捨てる。
func machiavelliFilterTable(table [][]*Card) [][]*Card {
	out := make([][]*Card, 0, len(table))
	for _, meld := range table {
		if meld == nil {
			continue
		}
		cleaned := machiavelliFilterNilCards(meld)
		if len(cleaned) > 0 {
			out = append(out, cleaned)
		}
	}
	return out
}
