//go:build !js || !wasm || extra

// Package domain: Indian Rummy (13-card) implementation.
//
// Indian Rummy is a draw-and-discard rummy played with 2 standard decks plus
// jokers (108 cards). Each player is dealt 13 cards and, on their turn, draws
// one card (from the stock or the discard top) then discards one — so a hand is
// held at 13 and momentarily rises to 14 during a turn.
//
// Wild joker rule (clean interpretation): after the deal one card is turned up;
// its rank becomes wild for the round, so all four suits of that rank act as
// jokers, together with the printed jokers. A wild-rank card is ALWAYS treated
// as a wild — it is never used at its natural rank. If the turned-up card is a
// printed joker, no additional rank is wild (only the printed jokers).
//
// Melds:
//   - Set: 3–4 cards of the same rank in different suits. A set may use at most
//     one wild.
//   - Sequence (run): 3+ consecutive cards of the same suit. A PURE sequence
//     uses no wild/joker; an IMPURE sequence uses one wild. A meld uses at most
//     one wild.
//
// Declaration validity: to declare, all 13 cards must form valid melds AND the
// hand must contain at least 2 sequences, at least one of which is a PURE
// sequence.
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
	"strconv"
)

// IndianRummyHandSize 各プレイヤーの手札枚数
const IndianRummyHandSize = 13

// IndianRummyDeadwoodCap 1 プレイヤーあたりのデッドウッド上限（標準の 80 点キャップ）
const IndianRummyDeadwoodCap = 80

// IndianRummyDefaultTargetRounds 既定のラウンド数
const IndianRummyDefaultTargetRounds = 3

// IndianRummySeqMin シーケンス（ラン）の最小枚数
const IndianRummySeqMin = 3

// IndianRummySetMin セットの最小枚数
const IndianRummySetMin = 3

// IndianRummyPhase ゲームフェーズ
type IndianRummyPhase int

// Indian Rummy のフェーズ定数
const (
	// IndianRummyPhaseDraw ドローフェーズ（山札 or 捨て札トップから 1 枚引く）
	IndianRummyPhaseDraw IndianRummyPhase = 0
	// IndianRummyPhaseDiscard ディスカードフェーズ（手札 14 枚から 1 枚捨てる or 宣言する）
	IndianRummyPhaseDiscard IndianRummyPhase = 1
	// IndianRummyPhaseRoundEnd ラウンド終了フェーズ
	IndianRummyPhaseRoundEnd IndianRummyPhase = 2
	// IndianRummyPhaseGameEnd ゲーム終了フェーズ
	IndianRummyPhaseGameEnd IndianRummyPhase = 3
)

// IndianRummyHint represents a hint for the human player.
type IndianRummyHint struct {
	// Action is the recommended action name.
	Action string
	// Reason is the identifier of the reason.
	Reason string
}

// newIndianRummyDeck インドラミー用 108 枚デッキ（標準 52 枚デッキ×2 + ジョーカー 4 枚）を構築する。
func newIndianRummyDeck() *TrumpCards {
	return NewTrumpCardsWithDecks(2, 4)
}

// buildIndianRummyPlayers n 人分のプレイヤー（席 0 が人間、残りが CPU）を生成する
func buildIndianRummyPlayers(n int) []*IndianRummyPlayer {
	if n < IndianRummyPlayerCountMin {
		n = IndianRummyPlayerCountMin
	}
	if n > IndianRummyPlayerCountMax {
		n = IndianRummyPlayerCountMax
	}
	players := make([]*IndianRummyPlayer, n)
	players[0] = NewIndianRummyPlayer(true)
	for i := 1; i < n; i++ {
		players[i] = NewIndianRummyPlayer(false)
	}
	return players
}

// IndianRummy インドラミー（13 枚制ラミー）のゲームクラス。
type IndianRummy struct {
	trumpCards       *TrumpCards
	players          []*IndianRummyPlayer
	config           IndianRummyConfig
	phase            IndianRummyPhase
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

// NewIndianRummy コンストラクタ
func NewIndianRummy(trumpCards *TrumpCards, players []*IndianRummyPlayer, config IndianRummyConfig) *IndianRummy {
	return &IndianRummy{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		declarerIdx: -1,
	}
}

func (g *IndianRummy) appendLog(playerIdx int, actionType, detailCode string, detailParams map[string]string, cards []*Card) {
	g.appendLogCode(playerIdx, actionType, detailCode, detailParams, cards)
}

// NewDefaultIndianRummy 標準構成（人間 1 + CPU 3、108 枚デッキ、デフォルト設定）でコンストラクトする SSoT。
func NewDefaultIndianRummy() *IndianRummy {
	cfg := DefaultIndianRummyConfig()
	return NewIndianRummy(newIndianRummyDeck(), buildIndianRummyPlayers(cfg.PlayerCount), cfg)
}

// Reset ゲームを初期化する
func (g *IndianRummy) Reset() {
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
	g.players = buildIndianRummyPlayers(g.config.PlayerCount)
	g.currentPlayerIdx = (g.dealerIdx + 1) % len(g.players)

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = IndianRummyPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *IndianRummy) NextRound() {
	if g.phase != IndianRummyPhaseRoundEnd {
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

	g.phase = IndianRummyPhaseDraw
}

// dealInitialCards 各プレイヤーに 13 枚を配り、ワイルドジョーカーを開き、最初の 1 枚を捨て札に置く
func (g *IndianRummy) dealInitialCards() {
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

	for range IndianRummyHandSize {
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
		g.wildRank = indianRummyWildRankFromCard(wj)
	}

	// 最初の 1 枚を捨て札トップに置く。
	if len(g.drawPile) > 0 {
		first := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, first)
	}
}

// indianRummyWildRankFromCard 開かれたカードからワイルドランクを決める。
// 印刷ジョーカーなら 0（ランク指定なし）、それ以外はそのカードの値。
func indianRummyWildRankFromCard(c *Card) int {
	if c == nil || c.GetDesign() == CardDesignJoker {
		return 0
	}
	return c.GetValue()
}

// PlayerDrawFromStock 人間プレイヤーが山札から引く
func (g *IndianRummy) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != IndianRummyPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromStock()
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札トップから引く
func (g *IndianRummy) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != IndianRummyPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.drawFromDiscard()
}

func (g *IndianRummy) drawFromStock() error {
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

	g.appendLog(g.currentPlayerIdx, "draw_stock", "indianrummy.log.drawStock", map[string]string{"name": playerName(g.players, g.currentPlayerIdx)}, nil)
	g.phase = IndianRummyPhaseDiscard
	return nil
}

func (g *IndianRummy) drawFromDiscard() error {
	if len(g.discardPile) == 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "indianrummy.errDiscardPileEmpty", nil)
	}
	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", "indianrummy.log.drawDiscard", map[string]string{"name": playerName(g.players, g.currentPlayerIdx), "card": cardStr(card)}, []*Card{card})
	g.phase = IndianRummyPhaseDiscard
	return nil
}

// recycleDiscardIntoStock 山札が空のとき捨て札トップ 1 枚を残して残りを山札へ戻しシャッフルする。
func (g *IndianRummy) recycleDiscardIntoStock() bool {
	return recycleDiscardIntoStock(&g.discardPile, &g.drawPile, g, "indianrummy.log.recycle")
}

// PlayerDiscard 人間プレイヤーが手札 1 枚を捨ててターンを終了する
func (g *IndianRummy) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != IndianRummyPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDiscard(cardIndex)
}

func (g *IndianRummy) applyDiscard(cardIndex int) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "indianrummy.errDiscardCardIndexOutOfRange", nil)
	}
	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", "indianrummy.log.discard", map[string]string{"name": playerName(g.players, g.currentPlayerIdx), "card": cardStr(discarded)}, []*Card{discarded})
	g.advanceTurn()
	return nil
}

// PlayerDeclare 人間プレイヤーが宣言する。cardIndex は「フィニッシュスロット」に捨てる 14 枚目のインデックス。
// 残った 13 枚が有効なアレンジ（2+ シーケンス・うち 1 つ以上がピュア）であれば有効宣言。
func (g *IndianRummy) PlayerDeclare(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != IndianRummyPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDeclare(cardIndex)
}

func (g *IndianRummy) applyDeclare(cardIndex int) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "indianrummy.errDeclareCardIndexOutOfRange", nil)
	}
	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	g.declarerIdx = g.currentPlayerIdx
	cards := indianRummyCollectCards(player)
	g.declarationValid = IndianRummyValidateDeclaration(cards, g.wildRank)

	code := "indianrummy.log.declareValid"
	if !g.declarationValid {
		code = "indianrummy.log.declareInvalid"
	}
	g.appendLog(g.currentPlayerIdx, "declare", code, map[string]string{"name": playerName(g.players, g.currentPlayerIdx)}, nil)

	g.enterRoundEnd()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *IndianRummy) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % len(g.players)
	g.phase = IndianRummyPhaseDraw
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *IndianRummy) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// CpuPlay 現在の手番が CPU の場合にターンを実行する
func (g *IndianRummy) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	switch g.phase {
	case IndianRummyPhaseDraw:
		g.cpuDraw()
	case IndianRummyPhaseDiscard:
		g.cpuDiscardOrDeclare()
	}
}

// cpuDraw CPU の引き処理。捨て札トップが役を進めるなら拾い、そうでなければ山札から引く。
func (g *IndianRummy) cpuDraw() {
	cpuDrawTurn(g)
}

// cpuShouldTakeDiscard 捨て札トップを拾うべきかを返す
func (g *IndianRummy) cpuShouldTakeDiscard(top *Card) bool {
	if indianRummyIsWild(top, g.wildRank) {
		return true
	}
	player := g.players[g.currentPlayerIdx]
	cur := indianRummyCollectCards(player)
	without := indianRummyMinDeadwood(cur, g.wildRank)
	withTop := make([]*Card, 0, len(cur)+1)
	withTop = append(withTop, cur...)
	withTop = append(withTop, top)
	withVal := indianRummyMinDeadwood(withTop, g.wildRank)

	switch g.config.CpuDifficulty {
	case IndianRummyCpuDifficultyHard, IndianRummyCpuDifficultyNormal:
		return withVal < without
	default:
		return rand.Intn(3) == 0
	}
}

// cpuDiscardOrDeclare CPU のディスカード or 宣言処理
func (g *IndianRummy) cpuDiscardOrDeclare() {
	player := g.players[g.currentPlayerIdx]
	cards := indianRummyCollectCards(player)

	// 手役がほぼ完成しているときのみ宣言探索を行う（毎ターンの高コスト探索を回避）。
	if g.config.CpuDifficulty != IndianRummyCpuDifficultyEasy && indianRummyMinDeadwood(cards, g.wildRank) <= 2 {
		if f, ok := g.cpuFindDeclareCard(player); ok {
			_ = g.applyDeclare(f)
			return
		}
	}

	idx := g.chooseCpuDiscard(player)
	_ = g.applyDiscard(idx)
}

// cpuFindDeclareCard 14 枚のうち 1 枚をフィニッシュに回して残り 13 枚が有効宣言になるカードを探す。
func (g *IndianRummy) cpuFindDeclareCard(player *IndianRummyPlayer) (int, bool) {
	declarable := declarableIndianRummyDiscards(player, g.wildRank, true)
	if len(declarable) > 0 {
		return declarable[0], true
	}
	return 0, false
}

// chooseCpuDiscard CPU が捨てるカードを選ぶ（デッドウッド点が最大のカード）。
func (g *IndianRummy) chooseCpuDiscard(player *IndianRummyPlayer) int {
	if player.GetCardsSize() == 0 {
		return 0
	}
	bestIdx := 0
	bestVal := indianRummyCardPoints(player.GetCard(0), g.wildRank)
	for i := 1; i < player.GetCardsSize(); i++ {
		v := indianRummyCardPoints(player.GetCard(i), g.wildRank)
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

// enterRoundEnd ラウンド終了処理をフェーズ再入で 1 度だけ実行する（scored ガード）。
func (g *IndianRummy) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.scoreRound()
	if g.roundNumber >= g.config.TargetRounds {
		g.finalizeGameEnd()
		return
	}
	g.phase = IndianRummyPhaseRoundEnd
}

// endRoundStockOut 山札枯渇によるラウンド終了（宣言なし・全員デッドウッド採点）。
func (g *IndianRummy) endRoundStockOut() {
	g.declarerIdx = -1
	g.appendLog(-1, "stock_out", "indianrummy.log.stockOut", nil, nil)
	g.enterRoundEnd()
}

// scoreRound ラウンドのスコアを確定する。
func (g *IndianRummy) scoreRound() {
	for i := range g.players {
		var s int
		if g.declarerIdx >= 0 && i == g.declarerIdx {
			if g.declarationValid {
				s = 0
			} else {
				s = IndianRummyDeadwoodCap
			}
		} else {
			s = IndianRummyDeadwoodScore(indianRummyCollectCards(g.players[i]), g.wildRank)
		}
		g.players[i].SetRoundScore(s)
	}

	if g.declarerIdx >= 0 && g.declarationValid {
		g.appendLog(g.declarerIdx, "round_win", "indianrummy.log.roundWin", map[string]string{"name": playerName(g.players, g.declarerIdx)}, nil)
	} else if g.declarerIdx >= 0 {
		g.appendLog(g.declarerIdx, "round_end", "indianrummy.log.roundEnd", map[string]string{"name": playerName(g.players, g.declarerIdx), "penalty": strconv.Itoa(IndianRummyDeadwoodCap)}, nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}
}

// finalizeGameEnd ゲーム終了処理（累計最少のプレイヤーが勝者）。
func (g *IndianRummy) finalizeGameEnd() {
	g.gameEndFlag = true
	g.phase = IndianRummyPhaseGameEnd

	minScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() < minScore {
			minScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", "indianrummy.log.gameEnd", map[string]string{"name": playerName(g.players, g.winnerIdx), "points": strconv.Itoa(minScore)}, nil)
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *IndianRummy) GetPhase() IndianRummyPhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *IndianRummy) SetPhase(p IndianRummyPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号
func (g *IndianRummy) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定（テスト用）
func (g *IndianRummy) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *IndianRummy) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *IndianRummy) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// GetDealerIdx ディーラーインデックスを取得
func (g *IndianRummy) GetDealerIdx() int { return g.dealerIdx }

// GetDiscardPile 捨て札の山
func (g *IndianRummy) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *IndianRummy) SetDiscardPile(p []*Card) { g.discardPile = p }

// GetDiscardTop 捨て札トップ
func (g *IndianRummy) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札残り枚数
func (g *IndianRummy) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *IndianRummy) SetDrawPile(p []*Card) { g.drawPile = p }

// GetWildJoker ワイルドジョーカーカード（表示用、nil の場合あり）
func (g *IndianRummy) GetWildJoker() *Card { return g.wildJoker }

// GetWildRank ワイルドランク（0 = ランク指定なし）
func (g *IndianRummy) GetWildRank() int { return g.wildRank }

// SetWildRank ワイルドランク設定（テスト用）
func (g *IndianRummy) SetWildRank(r int) { g.wildRank = r }

// GetGameEndFlag ゲーム終了フラグ
func (g *IndianRummy) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (g *IndianRummy) SetGameEndFlag(v bool) { g.gameEndFlag = v }

// GetWinnerIdx 勝者インデックス（-1 未確定）
func (g *IndianRummy) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *IndianRummy) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *IndianRummy) GetPlayer(i int) *IndianRummyPlayer {
	return getPlayer(g.players, i)
}

// GetConfig 設定取得
func (g *IndianRummy) GetConfig() IndianRummyConfig { return g.config }

// SetConfig 設定変更
func (g *IndianRummy) SetConfig(c IndianRummyConfig) { g.config = c }

// GetTargetRounds ゲーム終了までのラウンド数
func (g *IndianRummy) GetTargetRounds() int { return g.config.TargetRounds }

// GetDeclarerIdx 宣言したプレイヤー（-1 = 宣言なし）
func (g *IndianRummy) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 宣言プレイヤー設定（テスト用）
func (g *IndianRummy) SetDeclarerIdx(i int) { g.declarerIdx = i }

// GetDeclarationValid 直近の宣言が有効だったか
func (g *IndianRummy) GetDeclarationValid() bool { return g.declarationValid }

// PlayerDeadwoodValue プレイヤー i のデッドウッド採点値（キャップ・ピュアシーケンス規則を含む）。
func (g *IndianRummy) PlayerDeadwoodValue(i int) int {
	p := g.GetPlayer(i)
	if p == nil {
		return 0
	}
	return IndianRummyDeadwoodScore(indianRummyCollectCards(p), g.wildRank)
}

// PlayerHasPureSequence プレイヤー i の手札にピュアシーケンスがあるか。
func (g *IndianRummy) PlayerHasPureSequence(i int) bool {
	p := g.GetPlayer(i)
	if p == nil {
		return false
	}
	return IndianRummyHasPureSequence(indianRummyCollectCards(p), g.wildRank)
}

// GetDeclarableDiscards returns the hand indices whose discard leaves a valid declaration.
func (g *IndianRummy) GetDeclarableDiscards() []int {
	if g.phase != IndianRummyPhaseDiscard {
		return []int{}
	}
	p := g.GetPlayer(g.currentPlayerIdx)
	if p == nil || !p.GetIsHuman() {
		return []int{}
	}
	return declarableIndianRummyDiscards(p, g.wildRank, false)
}

func declarableIndianRummyDiscards(player *IndianRummyPlayer, wildRank int, firstOnly bool) []int {
	n := player.GetCardsSize()
	declarable := make([]int, 0)
	for f := 0; f < n; f++ {
		rem := make([]*Card, 0, n-1)
		for i := 0; i < n; i++ {
			if i != f {
				rem = append(rem, player.GetCard(i))
			}
		}
		if IndianRummyValidateDeclaration(rem, wildRank) {
			declarable = append(declarable, f)
			if firstOnly {
				break
			}
		}
	}
	return declarable
}

// indianRummyFitsWithHand checks if card fits with hand to form a potential meld.
// Sync: frontend/src/utils/hints/indianRummyHint.ts (fitsWithHand)
func indianRummyFitsWithHand(card *Card, hand []*Card, wildRank int) bool {
	if card == nil {
		return false
	}
	var natural []*Card
	for _, c := range hand {
		if !indianRummyIsWild(c, wildRank) {
			natural = append(natural, c)
		}
	}

	// Set: another card of the same value.
	for _, c := range natural {
		if c.GetValue() == card.GetValue() {
			return true
		}
	}

	// Run: same suit within two ranks (adjacent or one-gap).
	for _, c := range natural {
		if c.GetDesign() == card.GetDesign() {
			diff := c.GetValue() - card.GetValue()
			if diff < 0 {
				diff = -diff
			}
			if diff == 1 || diff == 2 {
				return true
			}
		}
	}
	return false
}

// GetHint returns the recommended action for the human player.
// Sync: frontend/src/utils/hints/indianRummyHint.ts (getIndianRummyHint)
func (g *IndianRummy) GetHint() *IndianRummyHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.gameEndFlag || g.currentPlayerIdx != human {
		return &IndianRummyHint{Reason: "none"}
	}
	player := g.players[human]
	if player.GetCardsSize() == 0 {
		return &IndianRummyHint{Reason: "none"}
	}

	switch g.phase {
	case IndianRummyPhaseDraw:
		// Sync: frontend/src/utils/hints/indianRummyHint.ts (getDrawHint)
		cards := indianRummyCollectCards(player)
		top := g.GetDiscardTop()
		if top != nil && !indianRummyIsWild(top, g.wildRank) && indianRummyFitsWithHand(top, cards, g.wildRank) {
			return &IndianRummyHint{Action: "drawDiscard", Reason: "draw_discard"}
		}
		return &IndianRummyHint{Action: "drawStock", Reason: "draw_stock"}

	case IndianRummyPhaseDiscard:
		// 判断は frontend/src/utils/hints/indianRummyHint.ts (getDiscardHint) と同じ —
		// **ただし写しではない。** あちらは calcDeadwood が 0 になるかしか見ておらず、
		// 「シーケンス 2 つ以上・うち 1 つはピュア」の宣言条件を検査していない。
		// こちらは IndianRummyValidateDeclaration を通す本物の検査なので、
		// **無効な宣言を勧めない。** フロント側の穴は #7736。
		if _, canDeclare := g.cpuFindDeclareCard(player); canDeclare {
			return &IndianRummyHint{Action: "declare", Reason: "declare_now"}
		}
		return &IndianRummyHint{Action: "discard", Reason: "discard_deadwood"}

	default:
		return &IndianRummyHint{Reason: "none"}
	}
}

// --- Private helpers ---

func (g *IndianRummy) sortAllHands() {
	sortHands(len(g.players), g)
}

func (g *IndianRummy) sortHand(playerIdx int) {
	sortHandInPlace(g.players[playerIdx], sortCards)
}

// indianRummyCollectCards プレイヤーの手札を []*Card で返す
func indianRummyCollectCards(p *IndianRummyPlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// --- Wild / points ---

// indianRummyIsWild card がワイルド（印刷ジョーカー or ワイルドランク）かどうか。
func indianRummyIsWild(card *Card, wildRank int) bool {
	if card == nil {
		return false
	}
	if card.GetDesign() == CardDesignJoker {
		return true
	}
	return wildRank != 0 && card.GetValue() == wildRank
}

// indianRummyCardPoints デッドウッド点（ワイルド=0、A=10、2-9=face、10/J/Q/K=10）。
func indianRummyCardPoints(card *Card, wildRank int) int {
	if indianRummyIsWild(card, wildRank) {
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

// IndianRummyCardPoints はデッドウッド計算に使うカード点を返す（外部公開用）。
func IndianRummyCardPoints(card *Card, wildRank int) int {
	return indianRummyCardPoints(card, wildRank)
}

// IndianRummyValidateDeclaration は全カード被覆とシーケンス条件を検証する。
func IndianRummyValidateDeclaration(cards []*Card, wildRank int) bool {
	if len(cards) != IndianRummyHandSize {
		return false
	}
	return rummyValidate(cards, indianRummyRules(wildRank), func(seq, pure int) bool { return seq >= 2 && pure >= 1 })
}

// IndianRummyHasPureSequence はピュアシーケンスの有無を返す。
func IndianRummyHasPureSequence(cards []*Card, wildRank int) bool {
	return rummyHasPureSequence(cards, indianRummyRules(wildRank))
}

// IndianRummyDeadwoodScore はデッドウッド採点値を返す。
func IndianRummyDeadwoodScore(cards []*Card, wildRank int) int {
	if !IndianRummyHasPureSequence(cards, wildRank) {
		return IndianRummyDeadwoodCap
	}
	dw := indianRummyMinDeadwood(cards, wildRank)
	if dw > IndianRummyDeadwoodCap {
		dw = IndianRummyDeadwoodCap
	}
	return dw
}
func indianRummyMinDeadwood(cards []*Card, wildRank int) int {
	v, _ := rummyMinDeadwood(cards, indianRummyRules(wildRank))
	return v
}
func indianRummyRules(wildRank int) rummyMeldRules {
	return rummyMeldRules{isWild: func(c *Card) bool { return indianRummyIsWild(c, wildRank) }, points: func(c *Card) int { return indianRummyCardPoints(c, wildRank) }}
}

// --- JSON ---

// indianRummyJSON は IndianRummy の JSON 表現。
type indianRummyJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*IndianRummyPlayer `json:"pl"`
	Config           IndianRummyConfig    `json:"cf"`
	Phase            IndianRummyPhase     `json:"ps"`
	CurrentPlayerIdx int                  `json:"ci"`
	DealerIdx        int                  `json:"di"`
	DiscardPile      []*Card              `json:"dp"`
	DrawPile         []*Card              `json:"wp"`
	WildJoker        *Card                `json:"wj"`
	WildRank         int                  `json:"wr"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerIdx        int                  `json:"wi"`
	RoundNumber      int                  `json:"rn"`
	Scored           bool                 `json:"sc"`
	DeclarerIdx      int                  `json:"de"`
	DeclarationValid bool                 `json:"dv"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *IndianRummy) MarshalJSON() ([]byte, error) {
	return json.Marshal(indianRummyJSON{
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

const indianRummyMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. KV 復元時に全インデックス・要素を検証し、
// 範囲外の値や nil 要素が後段でパニックすることを防ぐ。
func (g *IndianRummy) UnmarshalJSON(data []byte) error {
	var j indianRummyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > indianRummyMaxSliceLen || len(j.DiscardPile) > indianRummyMaxSliceLen ||
		len(j.DrawPile) > indianRummyMaxSliceLen || len(j.ActionLog) > indianRummyMaxSliceLen {
		return fmt.Errorf("indianrummy: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newIndianRummyDeck()
	}

	// プレイヤー要素の nil を拒否し、人数を検証する。
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*IndianRummyPlayer, 0)
	}
	for i, p := range g.players {
		if p == nil {
			return fmt.Errorf("indianrummy: player %d is nil", i)
		}
	}
	n := len(g.players)
	if n < IndianRummyPlayerCountMin || n > IndianRummyPlayerCountMax {
		return fmt.Errorf("indianrummy: invalid player count %d", n)
	}

	g.config = j.Config
	if g.config.PlayerCount <= 0 {
		g.config.PlayerCount = n
	}
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("indianrummy: invalid config: %w", err)
	}

	if j.Phase < IndianRummyPhaseDraw || j.Phase > IndianRummyPhaseGameEnd {
		return fmt.Errorf("indianrummy: invalid phase %d", j.Phase)
	}
	g.phase = j.Phase

	if j.RoundNumber < 0 {
		return fmt.Errorf("indianrummy: invalid round number %d", j.RoundNumber)
	}
	g.roundNumber = j.RoundNumber

	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= n {
		return fmt.Errorf("indianrummy: currentPlayerIdx %d out of range", j.CurrentPlayerIdx)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= n {
		return fmt.Errorf("indianrummy: dealerIdx %d out of range", j.DealerIdx)
	}
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.dealerIdx = j.DealerIdx

	if err := indianRummyValidateSentinelIdx("winnerIdx", j.WinnerIdx, n); err != nil {
		return err
	}
	if err := indianRummyValidateSentinelIdx("declarerIdx", j.DeclarerIdx, n); err != nil {
		return err
	}
	g.winnerIdx = j.WinnerIdx
	g.declarerIdx = j.DeclarerIdx

	if j.WildRank < 0 || j.WildRank > CardValueMax {
		return fmt.Errorf("indianrummy: invalid wild rank %d", j.WildRank)
	}
	g.wildRank = j.WildRank
	g.wildJoker = j.WildJoker

	g.gameEndFlag = j.GameEndFlag
	g.scored = j.Scored
	g.declarationValid = j.DeclarationValid

	// カードスライスから nil 要素を除去する（後段のイテレーションでの nil デリファレンス防止）。
	g.discardPile = indianRummyFilterNilCards(j.DiscardPile)
	g.drawPile = indianRummyFilterNilCards(j.DrawPile)

	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// indianRummyValidateSentinelIdx は -1（未確定センチネル）または [0, n) の範囲を許容する。
func indianRummyValidateSentinelIdx(name string, idx, n int) error {
	if idx == -1 {
		return nil
	}
	if idx < 0 || idx >= n {
		return fmt.Errorf("indianrummy: %s %d out of range", name, idx)
	}
	return nil
}

// indianRummyFilterNilCards nil 要素を除いた新しいスライスを返す（nil 入力には空スライス）。
func indianRummyFilterNilCards(cards []*Card) []*Card {
	out := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}
