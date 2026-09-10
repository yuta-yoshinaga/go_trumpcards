//go:build !js || !wasm || extra5

// Package domain: Marriage (21-card Nepali rummy) implementation.
//
// Marriage is a draw-and-discard rummy played with 3 standard 52-card decks
// plus 6 printed jokers (162 cards). Each player is dealt 21 cards and, on their turn, draws
// one card (from the stock or the discard top) then discards one — so a hand is
// held at 21 and momentarily rises to 22 during a turn.
//
// Wild joker rule: after the deal one card is turned up. The tiplu's rank, the
// ranks immediately above and below it, regardless of suit, and printed jokers
// are wild for the round. A wild-rank card is ALWAYS treated as a wild — it is
// never used at its natural rank. If the turned-up card is a printed joker, no
// additional rank is wild (only the printed jokers). Maal names are separate:
// they compare the tiplu's suit as well as its rank.
//
// Melds:
//   - Set: 3–4 cards of the same rank in different suits. A set may use at most
//     one wild.
//   - Sequence (run): 3+ consecutive cards of the same suit. A PURE sequence
//     uses no wild/joker; an IMPURE sequence uses one wild. A meld uses at most
//     one wild.
//
// Declaration validity: to declare, all 21 cards must form valid melds AND the
// hand must contain at least 3 pure sequences. Maal is enabled only when the
// hand contains at least 3 mutually disjoint pure sequences. This scoring is a
// house rule, not the sole historically authentic way to play Marriage.
//
// Maal kind       Points
// Tiplu               3
// Poplu               2
// Jhiplu              2
// Alter               1
// Printed joker       1
//
// Scoring on a declaration: the declarer scores 0 if the declaration is valid,
// otherwise the full cap (80). Each opponent scores the sum of their unmatched
// (deadwood) card points — Ace & face cards & 10 = 10, number cards = pip value,
// wild = 0 — capped at 80. A player with no pure sequence at all scores the full
// cap (80). Lowest cumulative score over TargetRounds wins.
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// MarriageHandSize 各プレイヤーの手札枚数
const MarriageHandSize = 21

// MarriageDeadwoodCap 1 プレイヤーあたりのデッドウッド上限（標準の 80 点キャップ）
const MarriageDeadwoodCap = 80

// MarriageDefaultTargetRounds 既定のラウンド数
const MarriageDefaultTargetRounds = 3

// MarriageSeqMin シーケンス（ラン）の最小枚数
const MarriageSeqMin = 3

// MarriageSetMin セットの最小枚数
const MarriageSetMin = 3

// MarriageMaalKind classifies a card's Marriage maal relative to the tiplu.
type MarriageMaalKind int

const (
	MarriageMaalNone MarriageMaalKind = iota
	MarriageMaalTiplu
	MarriageMaalPoplu
	MarriageMaalJhiplu
	MarriageMaalAlter
	MarriageMaalJoker
)

// These are the points for this repository's Marriage house rule.
const (
	MarriageMaalPointsTiplu  = 3
	MarriageMaalPointsPoplu  = 2
	MarriageMaalPointsJhiplu = 2
	MarriageMaalPointsAlter  = 1
	MarriageMaalPointsJoker  = 1
)

// MarriageMaalGateSequences is the number of pure sequences required to score maal.
const MarriageMaalGateSequences = 3

// marriageSearchCap は宣言/デッドウッド探索の最大反復回数。病的な手札での
// 指数的爆発を防ぐための保険（実手札では到達しない）。
const marriageSearchCap = 2000000

// marriagePureSequenceSearchCap は Worker の CPU 予算のため、純シーケンス
// 閾値探索に許可する最大反復回数。
const marriagePureSequenceSearchCap = 10000

// MarriagePhase ゲームフェーズ
type MarriagePhase int

// Marriage のフェーズ定数
const (
	// MarriagePhaseDraw ドローフェーズ（山札 or 捨て札トップから 1 枚引く）
	MarriagePhaseDraw MarriagePhase = 0
	// MarriagePhaseDiscard ディスカードフェーズ（手札 22 枚から 1 枚捨てる or 宣言する）
	MarriagePhaseDiscard MarriagePhase = 1
	// MarriagePhaseRoundEnd ラウンド終了フェーズ
	MarriagePhaseRoundEnd MarriagePhase = 2
	// MarriagePhaseGameEnd ゲーム終了フェーズ
	MarriagePhaseGameEnd MarriagePhase = 3
)

// newMarriageDeck Marriage 用 162 枚デッキ（標準 52 枚デッキ×3 + ジョーカー 6 枚）を構築する。
func newMarriageDeck() *TrumpCards {
	return NewTrumpCardsWithDecks(3, 6)
}

// buildMarriagePlayers n 人分のプレイヤー（席 0 が人間、残りが CPU）を生成する
func buildMarriagePlayers(n int) []*MarriagePlayer {
	if n < MarriagePlayerCountMin {
		n = MarriagePlayerCountMin
	}
	if n > MarriagePlayerCountMax {
		n = MarriagePlayerCountMax
	}
	players := make([]*MarriagePlayer, n)
	players[0] = NewMarriagePlayer(true)
	for i := 1; i < n; i++ {
		players[i] = NewMarriagePlayer(false)
	}
	return players
}

// Marriage ネパール／北インド式 21 枚制ラミーのゲームクラス。
type Marriage struct {
	trumpCards       *TrumpCards
	players          []*MarriagePlayer
	config           MarriageConfig
	phase            MarriagePhase
	currentPlayerIdx int
	dealerIdx        int
	discardPile      []*Card
	drawPile         []*Card
	wildJoker        *Card // 場に開かれたワイルドジョーカーカード（表示用）
	wildRank         int   // 当該ラウンドでワイルドとなるランク（0 = ランク指定なし・印刷ジョーカーのみワイルド）
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	scored           bool // ラウンド終了スコアリングが完了したか（フェーズ再入時の二重加算防止）
	declarerIdx      int  // 宣言したプレイヤー（-1 = 宣言なし／山切れ）
	declarationValid bool // 直近の宣言が有効だったか
	actionLogBase
}

// NewMarriage コンストラクタ
func NewMarriage(trumpCards *TrumpCards, players []*MarriagePlayer, config MarriageConfig) *Marriage {
	return &Marriage{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		declarerIdx: -1,
	}
}

// NewDefaultMarriage 標準構成（人間 1 + CPU 4、162 枚デッキ、デフォルト設定）でコンストラクトする SSoT。
func NewDefaultMarriage() *Marriage {
	cfg := DefaultMarriageConfig()
	return NewMarriage(newMarriageDeck(), buildMarriagePlayers(cfg.PlayerCount), cfg)
}

// Reset ゲームを初期化する
func (g *Marriage) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.dealerIdx = 0
	g.actionLog = nil
	g.scored = false
	g.declarerIdx = -1
	g.declarationValid = false
	g.wildJoker = nil
	g.wildRank = 0

	// 設定のプレイヤー数に合わせて席を再構築する（ResetWithConfig でも反映される）。
	g.players = buildMarriagePlayers(g.config.PlayerCount)
	g.currentPlayerIdx = (g.dealerIdx + 1) % len(g.players)

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = MarriagePhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Marriage) NextRound() {
	if g.phase != MarriagePhaseRoundEnd {
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
	g.declarerIdx = -1
	g.declarationValid = false
	g.wildJoker = nil
	g.wildRank = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = MarriagePhaseDraw
}

// dealInitialCards 各プレイヤーに 21 枚を配り、tiplu を表向きに開き、最初の 1 枚を捨て札に置く。
func (g *Marriage) dealInitialCards() {
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

	for range MarriageHandSize {
		for j := range len(g.players) {
			if len(g.drawPile) == 0 {
				break
			}
			card := g.drawPile[len(g.drawPile)-1]
			g.drawPile = g.drawPile[:len(g.drawPile)-1]
			g.players[j].AddCard(card)
		}
	}

	// ワイルドジョーカーを開く（そのランクが当該ラウンドのワイルドになる）。
	if len(g.drawPile) > 0 {
		wj := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.wildJoker = wj
		g.wildRank = marriageWildRankFromCard(wj)
	}

	// 最初の 1 枚を捨て札トップに置く。
	if len(g.drawPile) > 0 {
		first := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, first)
	}
}

// marriageWildRankFromCard 開かれたカードからワイルドランクを決める。
// 印刷ジョーカーなら 0（ランク指定なし）、それ以外はそのカードの値。
func marriageWildRankFromCard(c *Card) int {
	if c == nil || c.GetDesign() == CardDesignJoker {
		return 0
	}
	return c.GetValue()
}

// PlayerDrawFromStock 人間プレイヤーが山札から引く
func (g *Marriage) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MarriagePhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromStock()
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札トップから引く
func (g *Marriage) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MarriagePhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromDiscard()
}

func (g *Marriage) drawFromStock() error {
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
	g.phase = MarriagePhaseDiscard
	return nil
}

func (g *Marriage) drawFromDiscard() error {
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	g.phase = MarriagePhaseDiscard
	return nil
}

// recycleDiscardIntoStock 山札が空のとき捨て札トップ 1 枚を残して残りを山札へ戻しシャッフルする。
func (g *Marriage) recycleDiscardIntoStock() bool {
	return recycleDiscardIntoStock(&g.discardPile, &g.drawPile, g)
}

// PlayerDiscard 人間プレイヤーが手札 1 枚を捨ててターンを終了する
func (g *Marriage) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MarriagePhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDiscard(cardIndex)
}

func (g *Marriage) applyDiscard(cardIndex int) error {
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

// PlayerDeclare 人間プレイヤーが宣言する。cardIndex は「フィニッシュスロット」に捨てる 22 枚目のインデックス。
// 残った 21 枚が有効なアレンジ（3 つ以上のピュアシーケンス）であれば有効宣言。
func (g *Marriage) PlayerDeclare(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MarriagePhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDeclare(cardIndex)
}

func (g *Marriage) applyDeclare(cardIndex int) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	g.declarerIdx = g.currentPlayerIdx
	cards := marriageCollectCards(player)
	g.declarationValid = MarriageValidateDeclaration(cards, g.wildRank)

	status := "valid"
	if !g.declarationValid {
		status = "invalid"
	}
	g.appendLog(g.currentPlayerIdx, "declare", fmt.Sprintf("%s declares (%s)", playerName(g.players, g.currentPlayerIdx), status), nil)

	g.enterRoundEnd()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *Marriage) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.phase = MarriagePhaseDraw
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Marriage) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// CpuPlay 現在の手番が CPU の場合にターンを実行する
func (g *Marriage) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	switch g.phase {
	case MarriagePhaseDraw:
		g.cpuDraw()
	case MarriagePhaseDiscard:
		g.cpuDiscardOrDeclare()
	}
}

// cpuDraw CPU の引き処理。捨て札トップが役を進めるなら拾い、そうでなければ山札から引く。
func (g *Marriage) cpuDraw() {
	top := g.GetDiscardTop()
	if top != nil && g.cpuShouldTakeDiscard(top) {
		_ = g.drawFromDiscard()
		return
	}
	_ = g.drawFromStock()
}

// cpuShouldTakeDiscard 捨て札トップを拾うべきかを返す
func (g *Marriage) cpuShouldTakeDiscard(top *Card) bool {
	if marriageIsWild(top, g.wildRank) {
		return true
	}
	player := g.players[g.currentPlayerIdx]
	cur := marriageCollectCards(player)
	without := marriageMinDeadwood(cur, g.wildRank)
	withTop := make([]*Card, 0, len(cur)+1)
	withTop = append(withTop, cur...)
	withTop = append(withTop, top)
	withVal := marriageMinDeadwood(withTop, g.wildRank)

	switch g.config.CpuDifficulty {
	case MarriageCpuDifficultyHard, MarriageCpuDifficultyNormal:
		return withVal < without
	default:
		return rand.Intn(3) == 0
	}
}

// cpuDiscardOrDeclare CPU のディスカード or 宣言処理
func (g *Marriage) cpuDiscardOrDeclare() {
	player := g.players[g.currentPlayerIdx]
	cards := marriageCollectCards(player)

	// 手役がほぼ完成しているときのみ宣言探索を行う（毎ターンの高コスト探索を回避）。
	if g.config.CpuDifficulty != MarriageCpuDifficultyEasy && marriageMinDeadwood(cards, g.wildRank) <= 2 {
		if f, ok := g.cpuFindDeclareCard(player); ok {
			_ = g.applyDeclare(f)
			return
		}
	}

	idx := g.chooseCpuDiscard(player)
	_ = g.applyDiscard(idx)
}

// cpuFindDeclareCard 22 枚のうち 1 枚をフィニッシュに回して残り 21 枚が有効宣言になるカードを探す。
func (g *Marriage) cpuFindDeclareCard(player *MarriagePlayer) (int, bool) {
	n := player.GetCardsSize()
	for f := 0; f < n; f++ {
		rem := make([]*Card, 0, n-1)
		for i := 0; i < n; i++ {
			if i != f {
				rem = append(rem, player.GetCard(i))
			}
		}
		if MarriageValidateDeclaration(rem, g.wildRank) {
			return f, true
		}
	}
	return 0, false
}

// chooseCpuDiscard CPU が捨てるカードを選ぶ（デッドウッド点が最大のカード）。
func (g *Marriage) chooseCpuDiscard(player *MarriagePlayer) int {
	if player.GetCardsSize() == 0 {
		return 0
	}
	bestIdx := 0
	bestVal := marriageCardPoints(player.GetCard(0), g.wildRank)
	for i := 1; i < player.GetCardsSize(); i++ {
		v := marriageCardPoints(player.GetCard(i), g.wildRank)
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

// enterRoundEnd ラウンド終了処理をフェーズ再入で 1 度だけ実行する（scored ガード）。
func (g *Marriage) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.scoreRound()
	if g.roundNumber >= g.config.TargetRounds {
		g.finalizeGameEnd()
		return
	}
	g.phase = MarriagePhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了（宣言なし・全員デッドウッド採点）。
func (g *Marriage) endRoundStockOut() {
	g.declarerIdx = -1
	g.appendLog(-1, "stock_out", "Round ends (stock exhausted)", nil)
	g.enterRoundEnd()
}

// scoreRound ラウンドのスコアを確定する。
func (g *Marriage) scoreRound() {
	for i := range g.players {
		var s int
		if g.declarerIdx >= 0 && i == g.declarerIdx {
			if g.declarationValid {
				s = 0
			} else {
				s = MarriageDeadwoodCap
			}
		} else {
			s = MarriageDeadwoodScore(marriageCollectCards(g.players[i]), g.wildRank)
		}
		maal := g.PlayerMaalValue(i)
		s -= maal
		g.players[i].SetRoundScore(s)
		if maal != 0 {
			g.appendLog(i, "maal", fmt.Sprintf("%s scores %d maal points", playerName(g.players, i), maal), nil)
		}
	}

	if g.declarerIdx >= 0 && g.declarationValid {
		g.appendLog(g.declarerIdx, "round_win", fmt.Sprintf("%s wins the round with a valid declaration", playerName(g.players, g.declarerIdx)), nil)
	} else if g.declarerIdx >= 0 {
		g.appendLog(g.declarerIdx, "round_end", fmt.Sprintf("%s made an invalid declaration (+%d penalty)", playerName(g.players, g.declarerIdx), MarriageDeadwoodCap), nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}
}

// finalizeGameEnd ゲーム終了処理（累計最少のプレイヤーが勝者）。
func (g *Marriage) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = MarriagePhaseGameEnd

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
func (g *Marriage) GetPhase() MarriagePhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *Marriage) SetPhase(p MarriagePhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号
func (g *Marriage) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定（テスト用）
func (g *Marriage) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *Marriage) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *Marriage) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetDealerIdx ディーラーインデックスを取得
func (g *Marriage) GetDealerIdx() int { return g.dealerIdx }

// GetDiscardPile 捨て札の山
func (g *Marriage) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *Marriage) SetDiscardPile(p []*Card) { g.discardPile = p }

// GetDiscardTop 捨て札トップ
func (g *Marriage) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札残り枚数
func (g *Marriage) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *Marriage) SetDrawPile(p []*Card) { g.drawPile = p }

// GetWildJoker ワイルドジョーカーカード（表示用、nil の場合あり）
func (g *Marriage) GetWildJoker() *Card { return g.wildJoker }

// SetWildJoker ワイルドジョーカーカード設定（テスト用）
func (g *Marriage) SetWildJoker(c *Card) { g.wildJoker = c }

// GetWildRank ワイルドランク（0 = ランク指定なし）
func (g *Marriage) GetWildRank() int { return g.wildRank }

// SetWildRank ワイルドランク設定（テスト用）
func (g *Marriage) SetWildRank(r int) { g.wildRank = r }

// GetGameEndFlag ゲーム終了フラグ
func (g *Marriage) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス（-1 未確定）
func (g *Marriage) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *Marriage) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Marriage) GetPlayer(i int) *MarriagePlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *Marriage) GetConfig() MarriageConfig { return g.config }

// SetConfig 設定変更
func (g *Marriage) SetConfig(c MarriageConfig) { g.config = c }

// GetTargetRounds ゲーム終了までのラウンド数
func (g *Marriage) GetTargetRounds() int { return g.config.TargetRounds }

// GetDeclarerIdx 宣言したプレイヤー（-1 = 宣言なし）
func (g *Marriage) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 宣言プレイヤー設定（テスト用）
func (g *Marriage) SetDeclarerIdx(i int) { g.declarerIdx = i }

// GetDeclarationValid 直近の宣言が有効だったか
func (g *Marriage) GetDeclarationValid() bool { return g.declarationValid }

// PlayerDeadwoodValue プレイヤー i のデッドウッド採点値（キャップ・ピュアシーケンス規則を含む）。
func (g *Marriage) PlayerDeadwoodValue(i int) int {
	p := g.GetPlayer(i)
	if p == nil {
		return 0
	}
	return MarriageDeadwoodScore(marriageCollectCards(p), g.wildRank)
}

// PlayerMaalValue returns player i's maal deduction when the pure-sequence gate is met.
func (g *Marriage) PlayerMaalValue(i int) int {
	p := g.GetPlayer(i)
	if p == nil || g.wildJoker == nil {
		return 0
	}
	cards := marriageCollectCards(p)
	if !MarriageHasPureSequences(cards, g.wildRank, MarriageMaalGateSequences) {
		return 0
	}
	return MarriageMaalTotal(cards, g.wildJoker)
}

// PlayerHasPureSequence プレイヤー i の手札にピュアシーケンスがあるか。
func (g *Marriage) PlayerHasPureSequence(i int) bool {
	p := g.GetPlayer(i)
	if p == nil {
		return false
	}
	return MarriageHasPureSequence(marriageCollectCards(p), g.wildRank)
}

// --- Private helpers ---

func (g *Marriage) sortAllHands() {
	sortHands(len(g.players), g)
}

func (g *Marriage) sortHand(playerIdx int) {
	sortHandInPlace(g.players[playerIdx], marriageSortCards)
}

// marriageSortCards sorts a hand by suit and then rank for stable display.
func marriageSortCards(cards []*Card) {
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
}

// marriageCollectCards プレイヤーの手札を []*Card で返す
func marriageCollectCards(p *MarriagePlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// --- Wild / points ---

// MarriageIsTiplu reports whether card is the tiplu.
func MarriageIsTiplu(card, tiplu *Card) bool {
	return MarriageMaalOf(card, tiplu) == MarriageMaalTiplu
}

// MarriageIsPoplu は tiplu の次ランク（K の次は A）かを返す。
func MarriageIsPoplu(card, tiplu *Card) bool {
	return MarriageMaalOf(card, tiplu) == MarriageMaalPoplu
}

// MarriageIsJhiplu は tiplu の前ランク（A の前は K）かを返す。
func MarriageIsJhiplu(card, tiplu *Card) bool {
	return MarriageMaalOf(card, tiplu) == MarriageMaalJhiplu
}

// MarriageIsAlter は tiplu と同じランクで別スートのカードかを返す。
func MarriageIsAlter(card, tiplu *Card) bool {
	return MarriageMaalOf(card, tiplu) == MarriageMaalAlter
}

// MarriageMaalOf is the sole maal classifier. Maal names are distinct from
// the wild set: wildness uses rank only, while maal also uses suit.
func MarriageMaalOf(card, tiplu *Card) MarriageMaalKind {
	if card == nil || tiplu == nil {
		return MarriageMaalNone
	}
	if card.GetDesign() == CardDesignJoker {
		return MarriageMaalJoker
	}
	if card.GetDesign() == tiplu.GetDesign() && card.GetValue() == tiplu.GetValue() {
		return MarriageMaalTiplu
	}
	if card.GetDesign() == tiplu.GetDesign() && card.GetValue() == marriageNextRank(tiplu.GetValue()) {
		return MarriageMaalPoplu
	}
	if card.GetDesign() == tiplu.GetDesign() && card.GetValue() == marriagePreviousRank(tiplu.GetValue()) {
		return MarriageMaalJhiplu
	}
	if card.GetDesign() != tiplu.GetDesign() && card.GetValue() == tiplu.GetValue() {
		return MarriageMaalAlter
	}
	return MarriageMaalNone
}

// MarriageMaalPoints returns maal points for card relative to tiplu.
func MarriageMaalPoints(card, tiplu *Card) int {
	switch MarriageMaalOf(card, tiplu) {
	case MarriageMaalTiplu:
		return MarriageMaalPointsTiplu
	case MarriageMaalPoplu:
		return MarriageMaalPointsPoplu
	case MarriageMaalJhiplu:
		return MarriageMaalPointsJhiplu
	case MarriageMaalAlter:
		return MarriageMaalPointsAlter
	case MarriageMaalJoker:
		return MarriageMaalPointsJoker
	default:
		return 0
	}
}

// MarriageMaalTotal returns the total maal points in cards relative to tiplu.
func MarriageMaalTotal(cards []*Card, tiplu *Card) int {
	total := 0
	for _, card := range cards {
		total += MarriageMaalPoints(card, tiplu)
	}
	return total
}

// marriageNextRank はランクを循環させる。Marriage の house rule では K の次を A とする。
func marriageNextRank(rank int) int { return rank%CardValueMax + 1 }

// marriagePreviousRank はランクを循環させる。Marriage の house rule では A の前を K とする。
func marriagePreviousRank(rank int) int { return (rank+CardValueMax-2)%CardValueMax + 1 }

// marriageIsWild card がワイルド（tiplu/poplu/jhiplu/alter/印刷ジョーカー）かどうか。
func marriageIsWild(card *Card, wildRank int) bool {
	if card == nil {
		return false
	}
	if card.GetDesign() == CardDesignJoker {
		return true
	}
	if wildRank == 0 {
		return false
	}
	return card.GetValue() == wildRank || card.GetValue() == marriageNextRank(wildRank) || card.GetValue() == marriagePreviousRank(wildRank)
}

// marriageCardPoints デッドウッド点（ワイルド=0、A=10、2-9=face、10/J/Q/K=10）。
func marriageCardPoints(card *Card, wildRank int) int {
	if marriageIsWild(card, wildRank) {
		return 0
	}
	v := card.GetValue()
	if v == 1 { // Ace
		return 10
	}
	if v >= 10 { // 10, J, Q, K
		return 10
	}
	return v
}

// MarriageCardPoints はデッドウッド計算に使うカード点を返す（外部公開用）。
func MarriageCardPoints(card *Card, wildRank int) int {
	return marriageCardPoints(card, wildRank)
}

// --- Meld generation (joker-aware) ---

// marriageMeld は候補メルド。idx は cards スライス内のインデックス集合。
type marriageMeld struct {
	idx  []int
	seq  bool // シーケンス（ラン）か
	pure bool // ピュアシーケンス（ワイルド未使用）か
}

// marriageGenerateMelds cards から有効なセット／シーケンス候補を全列挙する。
func marriageGenerateMelds(cards []*Card, wildRank int) []marriageMeld {
	wildIdxs := make([]int, 0)
	for i, c := range cards {
		if marriageIsWild(c, wildRank) {
			wildIdxs = append(wildIdxs, i)
		}
	}
	melds := marriageGenerateSets(cards, wildRank, wildIdxs)
	melds = append(melds, marriageGenerateRuns(cards, wildRank, wildIdxs)...)
	return melds
}

// marriageDistinctSuitCombos idxs から k 枚を、全てスートが異なるように選ぶ組み合わせを返す。
func marriageDistinctSuitCombos(idxs []int, cards []*Card, k int) [][]int {
	bySuit := make(map[int][]int)
	order := make([]int, 0)
	for _, i := range idxs {
		s := cards[i].GetDesign()
		if _, ok := bySuit[s]; !ok {
			order = append(order, s)
		}
		bySuit[s] = append(bySuit[s], i)
	}
	res := make([][]int, 0)
	var rec func(pos int, cur []int)
	rec = func(pos int, cur []int) {
		if len(cur) == k {
			res = append(res, append([]int(nil), cur...))
			return
		}
		if pos >= len(order) || len(order)-pos < k-len(cur) {
			return
		}
		// このスートを使わない
		rec(pos+1, cur)
		// このスートの 1 枚を使う
		for _, i := range bySuit[order[pos]] {
			rec(pos+1, append(cur, i))
		}
	}
	rec(0, nil)
	return res
}

// marriageGenerateSets セット候補（同ランク別スート 3-4 枚、ワイルド最大 1 枚）を列挙する。
func marriageGenerateSets(cards []*Card, wildRank int, wildIdxs []int) []marriageMeld {
	byRank := make(map[int][]int)
	for i, c := range cards {
		if marriageIsWild(c, wildRank) {
			continue
		}
		byRank[c.GetValue()] = append(byRank[c.GetValue()], i)
	}
	melds := make([]marriageMeld, 0)
	for _, idxs := range byRank {
		// ピュアセット（ワイルドなし）3 枚・4 枚
		for _, k := range []int{3, 4} {
			for _, combo := range marriageDistinctSuitCombos(idxs, cards, k) {
				melds = append(melds, marriageMeld{idx: combo, seq: false, pure: false})
			}
		}
		// ワイルド 1 枚を含むセット（合計 3 枚・4 枚）
		for _, total := range []int{3, 4} {
			naturals := total - 1
			for _, combo := range marriageDistinctSuitCombos(idxs, cards, naturals) {
				for _, w := range wildIdxs {
					m := append(append([]int(nil), combo...), w)
					melds = append(melds, marriageMeld{idx: m, seq: false, pure: false})
				}
			}
		}
	}
	return melds
}

// marriageGenerateRuns シーケンス候補（同スート連続 3+ 枚、ワイルド最大 1 枚）を列挙する。
// Ace は low(A-2-3) / high(Q-K-A) の双方を許容する。
func marriageGenerateRuns(cards []*Card, wildRank int, wildIdxs []int) []marriageMeld {
	bySuit := make(map[int]map[int][]int)
	for i, c := range cards {
		if marriageIsWild(c, wildRank) {
			continue
		}
		s := c.GetDesign()
		if bySuit[s] == nil {
			bySuit[s] = make(map[int][]int)
		}
		v := c.GetValue()
		bySuit[s][v] = append(bySuit[s][v], i)
	}
	melds := make([]marriageMeld, 0)
	for _, byVal := range bySuit {
		for start := 1; start <= 13; start++ {
			for length := MarriageSeqMin; start+length-1 <= 14; length++ {
				m, ok := marriageBuildRunWindow(byVal, start, length, wildIdxs)
				if ok {
					melds = append(melds, m...)
				}
			}
		}
	}
	return melds
}

// marriageBuildRunWindow 単一スートの value→idxs から窓 [start, start+length-1] のラン候補を作る。
func marriageBuildRunWindow(byVal map[int][]int, start, length int, wildIdxs []int) ([]marriageMeld, bool) {
	present := make([][]int, 0, length)
	missing := 0
	seen := make(map[int]bool)
	for v := start; v < start+length; v++ {
		lv := v
		if lv == 14 {
			lv = 1 // Ace-high
		}
		if seen[lv] {
			return nil, false // 同じランクを二度参照する窓（ラップアラウンド）は不可
		}
		seen[lv] = true
		opts := byVal[lv]
		if len(opts) == 0 {
			missing++
			present = append(present, nil)
		} else {
			present = append(present, opts)
		}
	}
	switch {
	case missing == 0:
		combos := marriageCartesian(present)
		out := make([]marriageMeld, 0, len(combos))
		for _, combo := range combos {
			out = append(out, marriageMeld{idx: combo, seq: true, pure: true})
		}
		return out, true
	case missing == 1 && len(wildIdxs) > 0:
		base := make([][]int, 0, length-1)
		for _, opts := range present {
			if len(opts) > 0 {
				base = append(base, opts)
			}
		}
		combos := marriageCartesian(base)
		out := make([]marriageMeld, 0, len(combos)*len(wildIdxs))
		for _, combo := range combos {
			for _, w := range wildIdxs {
				m := append(append([]int(nil), combo...), w)
				out = append(out, marriageMeld{idx: m, seq: true, pure: false})
			}
		}
		return out, true
	default:
		return nil, false
	}
}

// marriageCartesian 各値の候補インデックスから 1 つずつ選ぶ直積を返す（爆発防止に上限あり）。
func marriageCartesian(lists [][]int) [][]int {
	const maxCartesian = 256
	res := [][]int{{}}
	for _, opts := range lists {
		if len(opts) == 0 {
			continue
		}
		next := make([][]int, 0, len(res)*len(opts))
		for _, prefix := range res {
			for _, o := range opts {
				next = append(next, append(append([]int(nil), prefix...), o))
			}
		}
		if len(next) > maxCartesian {
			next = next[:maxCartesian]
		}
		res = next
	}
	return res
}

// --- Declaration / deadwood search ---

// marriageCovering 各カードインデックスを覆うメルドのインデックス一覧を返す。
func marriageCovering(n int, melds []marriageMeld) [][]int {
	covering := make([][]int, n)
	for mi, m := range melds {
		for _, ci := range m.idx {
			covering[ci] = append(covering[ci], mi)
		}
	}
	return covering
}

// MarriageValidateDeclaration cards（21 枚）が有効宣言か。
// 全カードがメルドに収まり、ピュアシーケンスが 3 つ以上あれば true。
func MarriageValidateDeclaration(cards []*Card, wildRank int) bool {
	n := len(cards)
	if n != MarriageHandSize {
		return false
	}
	melds := marriageGenerateMelds(cards, wildRank)
	covering := marriageCovering(n, melds)
	decided := make([]bool, n)
	iter := 0

	var dfs func(seq, pure int) bool
	dfs = func(seq, pure int) bool {
		iter++
		if iter > marriageSearchCap {
			return false
		}
		i := marriageFirstUndecided(decided)
		if i == -1 {
			return pure >= 3
		}
		for _, mi := range covering[i] {
			m := melds[mi]
			if !marriageAllUndecided(decided, m.idx) {
				continue
			}
			marriageSetDecided(decided, m.idx, true)
			si, pi := 0, 0
			if m.seq {
				si = 1
				if m.pure {
					pi = 1
				}
			}
			if dfs(seq+si, pure+pi) {
				return true
			}
			marriageSetDecided(decided, m.idx, false)
		}
		return false
	}
	return dfs(0, 0)
}

// MarriageHasPureSequence cards にピュアシーケンス（ワイルド未使用の同スート連続 3+ 枚）が存在するか。
func MarriageHasPureSequence(cards []*Card, wildRank int) bool {
	for _, m := range marriageGenerateMelds(cards, wildRank) {
		if m.seq && m.pure {
			return true
		}
	}
	return false
}

// MarriageHasPureSequences reports whether cards contain n mutually disjoint
// pure sequences. It does not require covering the hand.
func MarriageHasPureSequences(cards []*Card, wildRank, n int) bool {
	if n <= 0 {
		return true
	}
	pure := make([]marriageMeld, 0)
	for _, m := range marriageGenerateMelds(cards, wildRank) {
		if m.seq && m.pure {
			pure = append(pure, m)
		}
	}
	used := make([]bool, len(cards))
	iter := 0
	var dfs func(int, int) bool
	dfs = func(pos, count int) bool {
		iter++
		if iter > marriagePureSequenceSearchCap {
			return false
		}
		if count >= n {
			return true
		}
		for i := pos; i < len(pure); i++ {
			m := pure[i]
			if !marriageAllUndecided(used, m.idx) {
				continue
			}
			marriageSetDecided(used, m.idx, true)
			if dfs(i+1, count+1) {
				return true
			}
			marriageSetDecided(used, m.idx, false)
		}
		return false
	}
	return dfs(0, 0)
}

// MarriageDeadwoodScore デッドウッド採点値を返す。
// ピュアシーケンスが無ければ 80（フルキャップ）。あれば最小デッドウッド点を 80 で頭打ちにする。
func MarriageDeadwoodScore(cards []*Card, wildRank int) int {
	if !MarriageHasPureSequence(cards, wildRank) {
		return MarriageDeadwoodCap
	}
	dw := marriageMinDeadwood(cards, wildRank)
	if dw > MarriageDeadwoodCap {
		dw = MarriageDeadwoodCap
	}
	return dw
}

// marriageMinDeadwood cards を互いに素なメルドで覆ったときの最小デッドウッド点を返す。
func marriageMinDeadwood(cards []*Card, wildRank int) int {
	n := len(cards)
	if n == 0 {
		return 0
	}
	melds := marriageGenerateMelds(cards, wildRank)
	covering := marriageCovering(n, melds)
	points := make([]int, n)
	for i, c := range cards {
		points[i] = marriageCardPoints(c, wildRank)
	}
	decided := make([]bool, n)
	iter := 0

	var dfs func() int
	dfs = func() int {
		iter++
		if iter > marriageSearchCap {
			s := 0
			for k := 0; k < n; k++ {
				if !decided[k] {
					s += points[k]
				}
			}
			return s
		}
		i := marriageFirstUndecided(decided)
		if i == -1 {
			return 0
		}
		// 選択肢 A: カード i をデッドウッドにする
		decided[i] = true
		best := points[i] + dfs()
		decided[i] = false
		// 選択肢 B: i を覆うメルドを使う
		for _, mi := range covering[i] {
			m := melds[mi]
			if !marriageAllUndecided(decided, m.idx) {
				continue
			}
			marriageSetDecided(decided, m.idx, true)
			c := dfs()
			marriageSetDecided(decided, m.idx, false)
			if c < best {
				best = c
			}
		}
		return best
	}
	return dfs()
}

func marriageFirstUndecided(decided []bool) int {
	for i := 0; i < len(decided); i++ {
		if !decided[i] {
			return i
		}
	}
	return -1
}

func marriageAllUndecided(decided []bool, idx []int) bool {
	for _, ci := range idx {
		if decided[ci] {
			return false
		}
	}
	return true
}

func marriageSetDecided(decided []bool, idx []int, v bool) {
	for _, ci := range idx {
		decided[ci] = v
	}
}

// --- JSON ---

// marriageJSON は Marriage の JSON 表現。
type marriageJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*MarriagePlayer `json:"pl"`
	Config           MarriageConfig    `json:"cf"`
	Phase            MarriagePhase     `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DealerIdx        int               `json:"di"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	WildJoker        *Card             `json:"wj"`
	WildRank         int               `json:"wr"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	Scored           bool              `json:"sc"`
	DeclarerIdx      int               `json:"de"`
	DeclarationValid bool              `json:"dv"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Marriage) MarshalJSON() ([]byte, error) {
	return json.Marshal(marriageJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DealerIdx:        g.dealerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		WildJoker:        g.wildJoker,
		WildRank:         g.wildRank,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		Scored:           g.scored,
		DeclarerIdx:      g.declarerIdx,
		DeclarationValid: g.declarationValid,
		ActionLog:        g.actionLog,
	})
}

const marriageMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. KV 復元時に全インデックス・要素を検証し、
// 範囲外の値や nil 要素が後段でパニックすることを防ぐ。
func (g *Marriage) UnmarshalJSON(data []byte) error {
	var j marriageJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > marriageMaxSliceLen || len(j.DiscardPile) > marriageMaxSliceLen ||
		len(j.DrawPile) > marriageMaxSliceLen || len(j.ActionLog) > marriageMaxSliceLen {
		return fmt.Errorf("marriage: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newMarriageDeck()
	}

	// プレイヤー要素の nil を拒否し、人数を検証する。
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*MarriagePlayer, 0)
	}
	for i, p := range g.players {
		if p == nil {
			return fmt.Errorf("marriage: player %d is nil", i)
		}
	}
	n := len(g.players)
	if n < MarriagePlayerCountMin || n > MarriagePlayerCountMax {
		return fmt.Errorf("marriage: invalid player count %d", n)
	}

	g.config = j.Config
	if g.config.PlayerCount <= 0 {
		g.config.PlayerCount = n
	}
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("marriage: invalid config: %w", err)
	}

	if j.Phase < MarriagePhaseDraw || j.Phase > MarriagePhaseGameEnd {
		return fmt.Errorf("marriage: invalid phase %d", j.Phase)
	}
	g.phase = j.Phase

	if j.RoundNumber < 0 {
		return fmt.Errorf("marriage: invalid round number %d", j.RoundNumber)
	}
	g.roundNumber = j.RoundNumber

	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= n {
		return fmt.Errorf("marriage: currentPlayerIdx %d out of range", j.CurrentPlayerIdx)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n {
		return fmt.Errorf("marriage: dealerIdx %d out of range", j.DealerIdx)
	}
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.dealerIdx = j.DealerIdx

	if err := marriageValidateSentinelIdx("winnerIdx", j.WinnerIdx, n); err != nil {
		return err
	}
	if err := marriageValidateSentinelIdx("declarerIdx", j.DeclarerIdx, n); err != nil {
		return err
	}
	g.winnerIdx = j.WinnerIdx
	g.declarerIdx = j.DeclarerIdx

	if j.WildRank < 0 || j.WildRank > CardValueMax {
		return fmt.Errorf("marriage: invalid wild rank %d", j.WildRank)
	}
	g.wildRank = j.WildRank
	g.wildJoker = j.WildJoker

	g.gameEndFlag = j.GameEndFlag
	g.scored = j.Scored
	g.declarationValid = j.DeclarationValid

	// カードスライスから nil 要素を除去する（後段のイテレーションでの nil デリファレンス防止）。
	g.discardPile = marriageFilterNilCards(j.DiscardPile)
	g.drawPile = marriageFilterNilCards(j.DrawPile)

	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// marriageValidateSentinelIdx は -1（未確定センチネル）または [0, n) の範囲を許容する。
func marriageValidateSentinelIdx(name string, idx, n int) error {
	if idx == -1 {
		return nil
	}
	if idx < 0 || idx >= n {
		return fmt.Errorf("marriage: %s %d out of range", name, idx)
	}
	return nil
}

// marriageFilterNilCards nil 要素を除いた新しいスライスを返す（nil 入力には空スライス）。
func marriageFilterNilCards(cards []*Card) []*Card {
	out := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}
