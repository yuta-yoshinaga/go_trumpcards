//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// Kalooki（カルーキ）はジャマイカ／英国系のジョーカーワイルド・ラミー。
//
// ルール概要:
//   - 2〜4 人。2 デッキ + ジョーカー 2 枚（合計 106 枚）を使用する。
//   - 各プレイヤーに 13 枚配り、残りを山札（stock）、1 枚を捨て札に表向きで置く。
//   - 手番: 山札トップ or 捨て札トップを 1 枚引く → 任意でメルド／レイオフ → 1 枚捨てる。
//   - メルドはセット（同ランク 3 枚以上）またはラン（同スート連続 3 枚以上）。
//   - オープニング要件: そのプレイヤーが最初に場へ出すメルド群の合計点が
//     OpeningThreshold（既定 51）点以上でなければならない。未満なら拒否される。
//   - オープン後はメルドの追加とレイオフ（自分・他者のメルドへ 1 枚足す）が自由になる。
//   - ジョーカーはワイルド（任意のカードの代用）。ジョーカーを含むメルドは基礎点の 1.5 倍
//     （floor）で得点する。
//   - 手札を最初に 0 枚にしたプレイヤーがラウンドに勝利。他者は残った手札（デッドウッド）の
//     合計点をペナルティとして累積する。1 ラウンドで決着し、ペナルティ最少が勝者。
//
// カード値スキーム（kalookiCardValue）:
//   A=15, K/Q/J/10=10, 2〜9=額面, ジョーカー=15。
//   これはオープニング合計点とデッドウッドペナルティの双方に用いる。

// KalookiHandSize 各プレイヤーへの初期配布枚数
const KalookiHandSize = 13

// KalookiMeldMinSize メルド（セット／ラン）の最小枚数
const KalookiMeldMinSize = 3

// KalookiJokerValue ジョーカー 1 枚の点（オープニング／デッドウッド計算用）
const KalookiJokerValue = 15

// KalookiPhase ゲームフェーズ
type KalookiPhase int

// Kalooki のフェーズ定数
const (
	// KalookiPhaseDraw ドローフェーズ（山札 or 捨て札トップから 1 枚引く）
	KalookiPhaseDraw KalookiPhase = 0
	// KalookiPhaseMeld メルド／レイオフ → ディスカードのフェーズ
	KalookiPhaseMeld KalookiPhase = 1
	// KalookiPhaseRoundEnd ラウンド終了フェーズ
	KalookiPhaseRoundEnd KalookiPhase = 2
	// KalookiPhaseGameEnd ゲーム終了フェーズ
	KalookiPhaseGameEnd KalookiPhase = 3
)

// kalookiMaxTurns 全 CPU 戦などでの暴走防止用のターン上限
const kalookiMaxTurns = 2000

// Kalooki カルーキのゲームクラス。
type Kalooki struct {
	trumpCards       *TrumpCards
	players          []*KalookiPlayer
	config           KalookiConfig
	phase            KalookiPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundWinnerIdx   int // 上がったプレイヤー。-1 は山切れ流局
	turnCount        int // 暴走防止用ターンカウンタ
	actionLogBase
}

// NewKalooki コンストラクタ
func NewKalooki(trumpCards *TrumpCards, players []*KalookiPlayer, config KalookiConfig) *Kalooki {
	return &Kalooki{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		roundWinnerIdx: -1,
	}
}

// NewDefaultKalooki 標準構成（人間 1 + CPU 3、2 デッキ + ジョーカー 2、デフォルト設定）でコンストラクトする SSoT。
func NewDefaultKalooki() *Kalooki {
	cfg := DefaultKalookiConfig()
	return NewKalooki(NewTrumpCardsWithDecks(2, 2), buildKalookiPlayers(cfg.PlayerCount), cfg)
}

// buildKalookiPlayers 人間 1 + CPU(n-1) のプレイヤースライスを作る。
func buildKalookiPlayers(n int) []*KalookiPlayer {
	if n < KalookiMinPlayers {
		n = KalookiMinPlayers
	}
	if n > KalookiMaxPlayers {
		n = KalookiMaxPlayers
	}
	players := make([]*KalookiPlayer, 0, n)
	players = append(players, NewKalookiPlayer(true))
	for i := 1; i < n; i++ {
		players = append(players, NewKalookiPlayer(false))
	}
	return players
}

// Reset ゲームを初期化する
func (g *Kalooki) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundWinnerIdx = -1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.turnCount = 0
	g.actionLog = nil

	// プレイヤー数が設定と食い違う場合は作り直す。
	if len(g.players) != g.config.PlayerCount {
		g.players = buildKalookiPlayers(g.config.PlayerCount)
	}
	for _, p := range g.players {
		p.SetCumulativeScore(0)
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = KalookiPhaseDraw
}

// NextRound 次のラウンドを開始する。Kalooki は 1 ラウンド完結なので、
// ラウンド終了後の呼び出しはゲーム終了を確定させる。
func (g *Kalooki) NextRound() {
	if g.phase != KalookiPhaseRoundEnd {
		return
	}
	g.finalizeGameEnd()
}

// dealInitialCards 各プレイヤーに KalookiHandSize 枚を配り、最初の 1 枚を捨て札トップに置く
func (g *Kalooki) dealInitialCards() {
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

	for range KalookiHandSize {
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

// PlayerDrawFromStock 人間プレイヤーが山札から引く
func (g *Kalooki) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KalookiPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromStock()
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札トップから引く
func (g *Kalooki) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KalookiPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromDiscard()
}

func (g *Kalooki) drawFromStock() error {
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
	g.phase = KalookiPhaseMeld
	return nil
}

func (g *Kalooki) drawFromDiscard() error {
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	g.phase = KalookiPhaseMeld
	return nil
}

// recycleDiscardIntoStock 山札が空のとき捨て札トップ 1 枚を残して残りを山札へ戻しシャッフルする。
func (g *Kalooki) recycleDiscardIntoStock() bool {
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

// PlayerMeld 人間プレイヤーがメルド群を場に出す。
// meldGroups[i] は i 番目のメルドに提出する手札インデックス群。
// まだオープンしていない場合、提出メルド群の合計点が OpeningThreshold 以上でなければ拒否する。
func (g *Kalooki) PlayerMeld(meldGroups [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KalookiPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyMeld(meldGroups)
}

func (g *Kalooki) applyMeld(meldGroups [][]int) error {
	player := g.players[g.currentPlayerIdx]
	if len(meldGroups) == 0 {
		return NewDomainError(ErrInvalidPlay, "メルドが指定されていません")
	}

	// 全インデックスの検証と（グループ間も含む）重複チェック
	allSeen := make(map[int]bool)
	groupCards := make([][]*Card, len(meldGroups))
	for gi, indices := range meldGroups {
		if len(indices) < KalookiMeldMinSize {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("メルドには最低 %d 枚必要です", KalookiMeldMinSize))
		}
		cards := make([]*Card, 0, len(indices))
		for _, idx := range indices {
			if idx < 0 || idx >= player.GetCardsSize() {
				return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
			}
			if allSeen[idx] {
				return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
			}
			allSeen[idx] = true
			cards = append(cards, player.GetCard(idx))
		}
		if !kalookiIsValidMeld(cards) {
			return NewDomainError(ErrInvalidPlay, "有効なメルド（セットまたはラン）ではありません")
		}
		groupCards[gi] = cards
	}

	// オープニング要件: まだオープンしていなければ合計点をチェックする
	if !player.HasOpened() {
		total := 0
		for _, cards := range groupCards {
			total += kalookiMeldValue(cards)
		}
		if total < g.config.OpeningThreshold {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("オープニングには合計 %d 点必要です（現在 %d 点）", g.config.OpeningThreshold, total))
		}
		player.SetHasOpened(true)
	}

	// メルドを場に追加
	for _, cards := range groupCards {
		meldCopy := make([]*Card, len(cards))
		copy(meldCopy, cards)
		sortCards(meldCopy)
		player.AppendMeld(meldCopy)
	}

	// インデックスを降順で削除
	allIndices := make([]int, 0, len(allSeen))
	for idx := range allSeen {
		allIndices = append(allIndices, idx)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(allIndices)))
	for _, idx := range allIndices {
		player.RemoveCard(idx)
	}

	g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d group(s)", playerName(g.players, g.currentPlayerIdx), len(groupCards)), nil)

	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerLayoff 人間プレイヤーが既存メルド（自分または他プレイヤー）にカード 1 枚を足す。
// オープン後でなければ実行できない。
func (g *Kalooki) PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KalookiPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyLayoff(targetPlayerIdx, meldIdx, cardIndex)
}

func (g *Kalooki) applyLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
	current := g.players[g.currentPlayerIdx]
	if !current.HasOpened() {
		return NewDomainError(ErrInvalidPlay, "レイオフはオープン後にのみ可能です")
	}
	if targetPlayerIdx < 0 || targetPlayerIdx >= len(g.players) {
		return NewDomainError(ErrInvalidPlay, "対象プレイヤーが不正です")
	}
	target := g.players[targetPlayerIdx]
	if !target.HasOpened() {
		return NewDomainError(ErrInvalidPlay, "対象プレイヤーがまだオープンしていません")
	}
	if meldIdx < 0 || meldIdx >= target.GetMeldCount() {
		return NewDomainError(ErrInvalidPlay, "対象メルドが不正です")
	}
	if cardIndex < 0 || cardIndex >= current.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := current.GetCard(cardIndex)
	meld := target.GetMeld(meldIdx)
	if !canAddToKalookiMeld(meld, card) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%s はそのメルドに追加できません", cardStr(card)))
	}
	target.AddCardToMeld(meldIdx, card)
	current.RemoveCard(cardIndex)

	g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s on player %d's meld", playerName(g.players, g.currentPlayerIdx), cardStr(card), targetPlayerIdx), []*Card{card})
	if current.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerDiscard 人間プレイヤーが手札 1 枚を捨ててターン終了する
func (g *Kalooki) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KalookiPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDiscard(cardIndex)
}

func (g *Kalooki) applyDiscard(cardIndex int) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
		return nil
	}

	g.advanceTurn()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *Kalooki) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.phase = KalookiPhaseDraw
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Kalooki) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// CpuPlay 現在の手番が CPU の場合にターンを実行する
func (g *Kalooki) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	g.turnCount++
	if g.turnCount > kalookiMaxTurns {
		g.endRoundStockOut()
		return
	}

	switch g.phase {
	case KalookiPhaseDraw:
		g.cpuDraw()
	case KalookiPhaseMeld:
		g.cpuMeld()
	}
}

// cpuDraw CPU の引き処理。捨て札トップが手役を進めるなら拾い、そうでなければ山札から引く
func (g *Kalooki) cpuDraw() {
	top := g.GetDiscardTop()
	if top != nil && g.cpuShouldTakeDiscard(top) {
		_ = g.drawFromDiscard()
		return
	}
	_ = g.drawFromStock()
}

// cpuShouldTakeDiscard 捨て札トップを拾うべきかを返す
func (g *Kalooki) cpuShouldTakeDiscard(top *Card) bool {
	player := g.players[g.currentPlayerIdx]
	if player.HasOpened() {
		// オープン済 → レイオフ可能なら拾う
		if _, _, ok := g.locateLayoffTarget(top); ok {
			return true
		}
	}
	switch g.config.CpuDifficulty {
	case KalookiCpuDifficultyHard:
		return false
	case KalookiCpuDifficultyNormal:
		return rand.Intn(8) == 0
	default:
		return rand.Intn(3) == 0
	}
}

// cpuMeld CPU のメルド・レイオフ・ディスカード処理
func (g *Kalooki) cpuMeld() {
	player := g.players[g.currentPlayerIdx]

	// オープン未達なら、しきい値を満たすメルド群を探して一気に出す
	if !player.HasOpened() {
		if groups := g.findOpeningMeld(player); groups != nil {
			_ = g.applyMeld(groups)
		}
	}

	// オープン済なら、追加メルド → レイオフを試みる
	if player.HasOpened() {
		for {
			cards := collectKalookiCards(player)
			extra := findKalookiMeld(cards)
			if extra == nil {
				break
			}
			handIdx := mapKalookiSelection(player, extra)
			if handIdx == nil {
				break
			}
			if err := g.applyMeld([][]int{handIdx}); err != nil {
				break
			}
		}
		for {
			done := false
			for i := 0; i < player.GetCardsSize(); i++ {
				card := player.GetCard(i)
				if pi, mi, ok := g.locateLayoffTarget(card); ok {
					if err := g.applyLayoff(pi, mi, i); err == nil {
						done = true
						break
					}
				}
			}
			if !done {
				break
			}
		}
		if player.GetCardsSize() == 0 {
			g.finishRound(g.currentPlayerIdx)
			return
		}
	}

	// ディスカード（最も高得点のカードを捨てる）
	idx := g.chooseCpuDiscard(player)
	_ = g.applyDiscard(idx)
}

// findOpeningMeld 手札からしきい値を満たすメルド群を貪欲に探す。見つからなければ nil。
func (g *Kalooki) findOpeningMeld(player *KalookiPlayer) [][]int {
	used := make([]bool, player.GetCardsSize())
	groups := make([][]int, 0)
	total := 0
	for {
		cards := collectUnusedKalookiCards(player, used)
		meld := findKalookiMeld(cards)
		if meld == nil {
			break
		}
		idx := mapKalookiSelectionMasked(player, meld, used)
		if idx == nil {
			break
		}
		for _, i := range idx {
			used[i] = true
		}
		groups = append(groups, idx)
		total += kalookiMeldValue(meld)
		if total >= g.config.OpeningThreshold {
			return groups
		}
	}
	return nil
}

// chooseCpuDiscard CPU が捨てるカードを選ぶ（最も高得点のカード）
func (g *Kalooki) chooseCpuDiscard(player *KalookiPlayer) int {
	if player.GetCardsSize() == 0 {
		return 0
	}
	bestIdx := 0
	bestVal := kalookiCardValue(player.GetCard(0))
	for i := 1; i < player.GetCardsSize(); i++ {
		v := kalookiCardValue(player.GetCard(i))
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

// locateLayoffTarget card のレイオフ先を返す（オープン済プレイヤーのメルドのみ対象）
func (g *Kalooki) locateLayoffTarget(card *Card) (int, int, bool) {
	for pi := range g.players {
		if !g.players[pi].HasOpened() {
			continue
		}
		for mi := 0; mi < g.players[pi].GetMeldCount(); mi++ {
			if canAddToKalookiMeld(g.players[pi].GetMeld(mi), card) {
				return pi, mi, true
			}
		}
	}
	return 0, 0, false
}

// finishRound 上がり／山切れの最終スコアリング。Kalooki は 1 ラウンド完結。
func (g *Kalooki) finishRound(winnerIdx int) {
	if g.phase == KalookiPhaseRoundEnd || g.phase == KalookiPhaseGameEnd {
		return
	}
	g.roundWinnerIdx = winnerIdx

	for i := range g.players {
		penalty := 0
		if winnerIdx < 0 || i != winnerIdx {
			for k := 0; k < g.players[i].GetCardsSize(); k++ {
				penalty += kalookiCardValue(g.players[i].GetCard(k))
			}
		}
		g.players[i].SetRoundScore(penalty)
		g.players[i].CommitRoundScore()
	}

	if winnerIdx >= 0 {
		g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s goes out!", playerName(g.players, winnerIdx)), nil)
	} else {
		g.appendLog(-1, "draw", "Round ends in a draw (stock empty)", nil)
	}

	g.phase = KalookiPhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了
func (g *Kalooki) endRoundStockOut() {
	g.finishRound(-1)
}

// finalizeGameEnd ゲーム終了処理（最少累計ペナルティのプレイヤーが勝者）
func (g *Kalooki) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = KalookiPhaseGameEnd

	// ラウンド勝者がいればそのまま勝者。流局時はペナルティ最少。
	if g.roundWinnerIdx >= 0 {
		g.winnerIdx = g.roundWinnerIdx
	} else {
		minScore := g.players[0].GetCumulativeScore()
		g.winnerIdx = 0
		for i := 1; i < len(g.players); i++ {
			if g.players[i].GetCumulativeScore() < minScore {
				minScore = g.players[i].GetCumulativeScore()
				g.winnerIdx = i
			}
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *Kalooki) GetPhase() KalookiPhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *Kalooki) SetPhase(p KalookiPhase) { g.phase = p }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *Kalooki) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *Kalooki) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetDiscardPile 捨て札の山
func (g *Kalooki) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *Kalooki) SetDiscardPile(p []*Card) { g.discardPile = p }

// GetDiscardTop 捨て札トップ
func (g *Kalooki) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札残り枚数
func (g *Kalooki) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *Kalooki) SetDrawPile(p []*Card) { g.drawPile = p }

// GetGameEndFlag ゲーム終了フラグ
func (g *Kalooki) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス
func (g *Kalooki) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *Kalooki) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Kalooki) GetPlayer(i int) *KalookiPlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *Kalooki) GetConfig() KalookiConfig { return g.config }

// SetConfig 設定変更
func (g *Kalooki) SetConfig(c KalookiConfig) { g.config = c }

// GetRoundWinnerIdx 直近ラウンドの勝者
func (g *Kalooki) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetOpeningThreshold オープニング要件のしきい値
func (g *Kalooki) GetOpeningThreshold() int { return g.config.OpeningThreshold }

// --- Private helpers ---

func (g *Kalooki) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

func (g *Kalooki) sortHand(playerIdx int) {
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

// --- Pure card-evaluation helpers ---

// kalookiIsJoker card がジョーカーか（Kalooki では 2 はワイルドではない）
func kalookiIsJoker(card *Card) bool {
	return card != nil && card.GetDesign() == CardDesignJoker
}

// kalookiCardValue 1 枚のカードの点（A=15, K/Q/J/10=10, 2-9=額面, ジョーカー=15）。
func kalookiCardValue(card *Card) int {
	if kalookiIsJoker(card) {
		return KalookiJokerValue
	}
	v := card.GetValue()
	if v == 1 {
		return 15
	}
	if v >= 10 {
		return 10
	}
	return v
}

// kalookiMeldValue メルドの得点。基礎点はカード値の合計。ジョーカーを含むなら 1.5 倍（floor）。
func kalookiMeldValue(cards []*Card) int {
	base := 0
	hasJoker := false
	for _, c := range cards {
		base += kalookiCardValue(c)
		if kalookiIsJoker(c) {
			hasJoker = true
		}
	}
	if hasJoker {
		return base * 3 / 2 // 1.5 倍を整数演算（floor）で算出
	}
	return base
}

// kalookiIsValidMeld cards が有効なメルド（セット 3+ またはラン 3+）か判定する。
// ジョーカーはワイルドとして任意のカードを代用できる。
func kalookiIsValidMeld(cards []*Card) bool {
	if len(cards) < KalookiMeldMinSize {
		return false
	}
	return kalookiIsSet(cards) || kalookiIsRun(cards)
}

// kalookiIsSet cards が同ランクのセットか（ジョーカーは任意ランクを代用）。
func kalookiIsSet(cards []*Card) bool {
	rank := -1
	jokers := 0
	for _, c := range cards {
		if kalookiIsJoker(c) {
			jokers++
			continue
		}
		if rank == -1 {
			rank = c.GetValue()
		} else if c.GetValue() != rank {
			return false
		}
	}
	// 全部ジョーカーは（実用上）不可。最低 1 枚は自然牌が必要。
	return rank != -1
}

// kalookiIsRun cards が同スートの連続するランか（ジョーカーは欠けた位置を埋める）。
// Ace は low (A-2-3) または high (Q-K-A) のいずれでも有効。
func kalookiIsRun(cards []*Card) bool {
	suit := -1
	jokers := 0
	naturals := make([]int, 0, len(cards))
	for _, c := range cards {
		if kalookiIsJoker(c) {
			jokers++
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
		// 全部ジョーカー: ランとしては無効
		return false
	}
	// aceVariants 内でソートされるが、kalookiRunFits のソート済み前提を明示的に
	// 担保するためここでもソートしておく（レイオフで末尾追加された場合に備える）。
	sort.Ints(naturals)
	for _, variant := range aceVariants(naturals) {
		if kalookiRunFits(variant, jokers) {
			return true
		}
	}
	return false
}

// kalookiRunFits ソート済みの自然牌の値列に jokers 枚のワイルドを足して、
// 連続するラン（重複なし）を構成できるかを判定する。
func kalookiRunFits(sortedNaturals []int, jokers int) bool {
	for i := 1; i < len(sortedNaturals); i++ {
		if sortedNaturals[i] == sortedNaturals[i-1] {
			return false // 同じ値はランに置けない
		}
	}
	if len(sortedNaturals) == 0 {
		return false
	}
	// ランは1スート最大13枚（A〜K）。自然牌＋ジョーカーがこれを超える組は不正。
	if len(sortedNaturals)+jokers > 13 {
		return false
	}
	// 端から端までを埋めるのに必要なジョーカー数
	span := sortedNaturals[len(sortedNaturals)-1] - sortedNaturals[0] + 1
	gaps := span - len(sortedNaturals)
	if gaps < 0 || gaps > jokers {
		return false
	}
	// 余ったジョーカーは両端の延長に使える（妥当な値域に収まる範囲で）。常に true。
	return true
}

// canAddToKalookiMeld card を既存メルドにレイオフできるか（ジョーカー対応）。
func canAddToKalookiMeld(meld []*Card, card *Card) bool {
	if len(meld) == 0 || card == nil {
		return false
	}
	extended := make([]*Card, 0, len(meld)+1)
	extended = append(extended, meld...)
	extended = append(extended, card)
	return kalookiIsValidMeld(extended)
}

// --- CPU helpers ---

// collectKalookiCards プレイヤーの手札を []*Card で返す
func collectKalookiCards(p *KalookiPlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// collectUnusedKalookiCards used[i]==false の手札のみを返す
func collectUnusedKalookiCards(p *KalookiPlayer, used []bool) []*Card {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if i < len(used) && used[i] {
			continue
		}
		cards = append(cards, p.GetCard(i))
	}
	return cards
}

// findKalookiMeld 手札からセット (3) またはラン (3) を 1 つ見つけて返す（ジョーカー対応）。
func findKalookiMeld(cards []*Card) []*Card {
	jokers := make([]*Card, 0)
	naturals := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if kalookiIsJoker(c) {
			jokers = append(jokers, c)
		} else {
			naturals = append(naturals, c)
		}
	}

	// セット: 同ランクが 3 枚、または 2 枚+ジョーカー1、または 1 枚+ジョーカー2
	byRank := make(map[int][]*Card)
	for _, c := range naturals {
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
	}
	for _, group := range byRank {
		need := KalookiMeldMinSize - len(group)
		if need <= 0 {
			return []*Card{group[0], group[1], group[2]}
		}
		if need <= len(jokers) {
			pick := make([]*Card, 0, KalookiMeldMinSize)
			pick = append(pick, group...)
			pick = append(pick, jokers[:need]...)
			if kalookiIsValidMeld(pick) {
				return pick
			}
		}
	}

	// ラン: 同スートで連続 3 枚（ジョーカーで隙間を埋める）
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
		if pick := findKalookiRunIn(byVal, jokers); pick != nil {
			return pick
		}
	}
	return nil
}

// findKalookiRunIn 1 スート内の value→card マップから、ジョーカーで隙間を埋めて
// 連続 3 枚のランを 1 つ探す。
func findKalookiRunIn(byVal map[int]*Card, jokers []*Card) []*Card {
	for start := 1; start <= 12; start++ {
		picked := make([]*Card, 0, KalookiMeldMinSize)
		jokersUsed := 0
		for v := start; v < start+KalookiMeldMinSize; v++ {
			if c, ok := byVal[v]; ok {
				picked = append(picked, c)
			} else if jokersUsed < len(jokers) {
				picked = append(picked, jokers[jokersUsed])
				jokersUsed++
			} else {
				picked = nil
				break
			}
		}
		if picked != nil && kalookiIsValidMeld(picked) {
			return picked
		}
	}
	return nil
}

// mapKalookiSelection selection をプレイヤーの手札インデックスへマップする
func mapKalookiSelection(p *KalookiPlayer, selection []*Card) []int {
	used := make([]bool, p.GetCardsSize())
	return mapKalookiSelectionMasked(p, selection, used)
}

// mapKalookiSelectionMasked used[i]==true を避けて selection を手札インデックスへマップする
func mapKalookiSelectionMasked(p *KalookiPlayer, selection []*Card, used []bool) []int {
	local := make([]bool, p.GetCardsSize())
	result := make([]int, 0, len(selection))
	for _, c := range selection {
		found := false
		for i := 0; i < p.GetCardsSize(); i++ {
			if (i < len(used) && used[i]) || local[i] {
				continue
			}
			if p.GetCard(i) == c {
				result = append(result, i)
				local[i] = true
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return result
}

// --- JSON ---

// kalookiJSON は Kalooki の JSON 表現
type kalookiJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*KalookiPlayer  `json:"pl"`
	Config           KalookiConfig     `json:"cf"`
	Phase            KalookiPhase      `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundWinnerIdx   int               `json:"rw"`
	TurnCount        int               `json:"tn"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Kalooki) MarshalJSON() ([]byte, error) {
	return json.Marshal(kalookiJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundWinnerIdx:   g.roundWinnerIdx,
		TurnCount:        g.turnCount,
		ActionLog:        g.actionLog,
	})
}

const kalookiMaxSliceLen = 1000

// errKalookiInvalidState は不正な復元データを表す共有センチネルエラー。
var errKalookiInvalidState = fmt.Errorf("kalooki: invalid game state")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Kalooki) UnmarshalJSON(data []byte) error {
	var j kalookiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > kalookiMaxSliceLen || len(j.DiscardPile) > kalookiMaxSliceLen ||
		len(j.DrawPile) > kalookiMaxSliceLen || len(j.ActionLog) > kalookiMaxSliceLen {
		return errKalookiInvalidState
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) < KalookiMinPlayers || len(j.Players) > KalookiMaxPlayers {
		return errKalookiInvalidState
	}
	for _, p := range j.Players {
		if p == nil {
			return errKalookiInvalidState
		}
	}
	if j.Phase < KalookiPhaseDraw || j.Phase > KalookiPhaseGameEnd {
		return errKalookiInvalidState
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= len(j.Players) {
		return errKalookiInvalidState
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsWithDecks(2, 2)
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
	g.roundWinnerIdx = j.RoundWinnerIdx
	g.turnCount = j.TurnCount
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
