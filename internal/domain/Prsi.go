package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// PrsiPlayerCnt プルシープレイヤー数 (1 human + 3 CPU)
const PrsiPlayerCnt = 4

// PrsiHandSize 初期配布枚数
const PrsiHandSize = 5

// PrsiSevenValue 7のカード値 (次のプレイヤーに2枚引かせる/重ね可)
const PrsiSevenValue = 7

// PrsiAceValue Aceのカード値 (1: 次のプレイヤーをスキップ)
const PrsiAceValue = 1

// PrsiUnderValue Under(=Jack/spodek)のカード値 (11: 次のプレイヤーをスキップ)
const PrsiUnderValue = 11

// PrsiSevenDrawAmount 7を1枚出すごとに累積するペナルティ枚数
const PrsiSevenDrawAmount = 2

// PrsiPhase ゲームフェーズ
type PrsiPhase int

// Prsiのフェーズ定数
const (
	// PrsiPhasePlay 通常プレイフェーズ
	PrsiPhasePlay PrsiPhase = 0
	// PrsiPhaseGameEnd ゲーム終了フェーズ (誰かの手札が空になった)
	PrsiPhaseGameEnd PrsiPhase = 1
)

// Prsi プルシー(Prší / チェコ版クレイジーエイト/Mau Mau)ゲームクラス。
//
// 32枚デッキ (7,8,9,10,J,Q,K,A × 4スート)。捨て札のトップとスートまたは
// ランクが一致するカードを出す。出せなければ山札から1枚引いて手番終了。
// アクションカード:
//   - 7      : 次のプレイヤーに2枚引かせる。7は重ねられる (2枚で4枚…)。
//     ペナルティを受ける手番のプレイヤーは別の7を重ねて次へ回せる。
//   - Ace    : 次のプレイヤーをスキップ。Aceは連続して重ねられる。
//   - Under  : Jack(spodek)。スキップとして扱う (次のプレイヤーを飛ばす)。
//
// 手札を最初に空にしたプレイヤーが勝利 (即時終了、ラウンド/スコアなし)。
type Prsi struct {
	trumpCards       *TrumpCards
	players          []*PrsiPlayer
	config           PrsiConfig
	phase            PrsiPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	penaltyDrawCount int // 累積7ペナルティ枚数 (0 = ペナルティなし)
	pendingSkips     int // 累積スキップ数 (Ace/Under)
	gameEndFlag      bool
	winnerIdx        int
	actionLogBase
}

// NewPrsi コンストラクタ
func NewPrsi(trumpCards *TrumpCards, players []*PrsiPlayer, config PrsiConfig) *Prsi {
	return &Prsi{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
	}
}

// NewDefaultPrsi returns Prsi with the standard 4-player setup (1 human, 3 CPU)
// and DefaultPrsiConfig. Single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultPrsi() *Prsi {
	players := []*PrsiPlayer{
		NewPrsiPlayer(true),
		NewPrsiPlayer(false),
		NewPrsiPlayer(false),
		NewPrsiPlayer(false),
	}
	return NewPrsi(NewTrumpCardsPrsi(), players, DefaultPrsiConfig())
}

// Reset ゲーム初期化
func (g *Prsi) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.penaltyDrawCount = 0
	g.pendingSkips = 0
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.actionLog = nil

	for _, p := range g.players {
		p.Reset()
		p.SetIsFinished(false)
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = PrsiPhasePlay
}

// dealInitialCards 初期配布: 各プレイヤーに5枚、1枚を捨て札に。
// 最初の捨て札が7/Ace/Underでも初期のアクション効果は発動しない。
func (g *Prsi) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}
	// The deck was already shuffled by trumpCards.Shuffle() in Reset(), so the
	// drawPile is in random order here — no second shuffle needed.

	for i := 0; i < PrsiHandSize; i++ {
		for j := 0; j < PrsiPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	if len(g.drawPile) > 0 {
		firstCard := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, firstCard)
	}
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Prsi) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PrsiPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if !g.isValidPlay(card) {
		return NewDomainError(ErrInvalidPlay, "そのカードは出せません")
	}

	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// PlayerDraw 人間プレイヤーがカードを引く (ペナルティ中はスタックを引き受ける)
func (g *Prsi) PlayerDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PrsiPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	return g.drawCard(g.currentPlayerIdx)
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (g *Prsi) CpuPlay() {
	if g.gameEndFlag || g.phase != PrsiPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	if cardIdx >= 0 {
		player := g.players[g.currentPlayerIdx]
		played := player.RemoveCard(cardIdx)
		// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
		// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
		// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
		if played == nil {
			return
		}
		g.playCard(g.currentPlayerIdx, played)
	} else {
		// drawCard always returns nil today; the error is ignored intentionally.
		_ = g.drawCard(g.currentPlayerIdx)
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Prsi) GetPhase() PrsiPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Prsi) SetPhase(phase PrsiPhase) { g.phase = phase }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Prsi) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Prsi) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Prsi) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Prsi) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Prsi) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Prsi) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Prsi) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetPenaltyDrawCount 累積7ペナルティ枚数取得 (0 = ペナルティなし)
func (g *Prsi) GetPenaltyDrawCount() int { return g.penaltyDrawCount }

// SetPenaltyDrawCount 累積7ペナルティ枚数設定 (テスト用)
func (g *Prsi) SetPenaltyDrawCount(n int) { g.penaltyDrawCount = n }

// GetPendingSkips 累積スキップ数取得 (Ace/Under)
func (g *Prsi) GetPendingSkips() int { return g.pendingSkips }

// SetPendingSkips 累積スキップ数設定 (テスト用)
func (g *Prsi) SetPendingSkips(n int) { g.pendingSkips = n }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Prsi) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Prsi) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Prsi) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Prsi) GetPlayer(i int) *PrsiPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Prsi) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Prsi) GetConfig() PrsiConfig { return g.config }

// SetConfig 設定変更
func (g *Prsi) SetConfig(cfg PrsiConfig) { g.config = cfg }

// --- Private methods ---

// isValidPlay カードがプレイ可能か判定
func (g *Prsi) isValidPlay(card *Card) bool {
	// 7ペナルティ中は7のみ重ねられる
	if g.penaltyDrawCount > 0 {
		return card.GetValue() == PrsiSevenValue
	}

	top := g.GetDiscardTop()
	if top == nil {
		return true
	}

	// スートまたはランクが一致
	return card.GetDesign() == top.GetDesign() || card.GetValue() == top.GetValue()
}

// playCard カードをプレイする共通処理
func (g *Prsi) playCard(playerIdx int, card *Card) {
	g.discardPile = append(g.discardPile, card)

	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	// アクションカードの状態更新
	switch card.GetValue() {
	case PrsiSevenValue:
		g.penaltyDrawCount += PrsiSevenDrawAmount
		g.appendLog(playerIdx, "seven", fmt.Sprintf("Draw stack is now %d", g.penaltyDrawCount), nil)
	case PrsiAceValue:
		g.pendingSkips++
		g.appendLog(playerIdx, "skip", "Next player is skipped (Ace)", nil)
	case PrsiUnderValue:
		g.pendingSkips++
		g.appendLog(playerIdx, "skip", "Next player is skipped (Under)", nil)
	}

	// 手札が空になったら勝利 (即時終了)
	if g.players[playerIdx].GetCardsSize() == 0 {
		g.players[playerIdx].SetIsFinished(true)
		g.winnerIdx = playerIdx
		g.gameEndFlag = true
		g.phase = PrsiPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, playerIdx)), nil)
		return
	}

	g.advanceTurn()
}

// advanceTurn 次のプレイヤーへ (スキップを反映)
func (g *Prsi) advanceTurn() {
	steps := 1 + g.pendingSkips
	g.pendingSkips = 0
	g.currentPlayerIdx = (g.currentPlayerIdx + steps) % PrsiPlayerCnt
	g.phase = PrsiPhasePlay
}

// drawCard カードを引く (ペナルティ中はスタックを引き受けて手番終了)
func (g *Prsi) drawCard(playerIdx int) error {
	if g.penaltyDrawCount > 0 {
		drawn := g.drawCards(playerIdx, g.penaltyDrawCount)
		g.appendLog(playerIdx, "take_penalty", fmt.Sprintf("%s takes %d penalty cards", playerName(g.players, playerIdx), drawn), nil)
		g.penaltyDrawCount = 0
		g.sortHand(playerIdx)
		g.advanceTurn()
		return nil
	}

	if len(g.drawPile) == 0 {
		g.recycleDrawPile()
	}

	if len(g.drawPile) == 0 {
		// 引けるカードがない→パス
		g.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes (no cards to draw)", playerName(g.players, playerIdx)), nil)
		g.advanceTurn()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[playerIdx].AddCard(card)
	g.sortHand(playerIdx)

	g.appendLog(playerIdx, "draw", fmt.Sprintf("%s draws a card", playerName(g.players, playerIdx)), nil)

	// プルシーでは引いたら手番終了 (引いたカードを即座に出すことはできない)
	g.advanceTurn()

	return nil
}

// drawCards 指定枚数を引く (山札が尽きたら捨て札を再利用)。実際に引けた枚数を返す。
func (g *Prsi) drawCards(playerIdx, n int) int {
	drawn := 0
	for i := 0; i < n; i++ {
		if len(g.drawPile) == 0 {
			g.recycleDrawPile()
		}
		if len(g.drawPile) == 0 {
			break
		}
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.players[playerIdx].AddCard(card)
		drawn++
	}
	return drawn
}

// recycleDrawPile 捨て札から山札を再構築する
func (g *Prsi) recycleDrawPile() {
	if len(g.discardPile) <= 1 {
		return
	}

	top := g.discardPile[len(g.discardPile)-1]
	recycled := g.discardPile[:len(g.discardPile)-1]
	g.discardPile = []*Card{top}

	rand.Shuffle(len(recycled), func(i, j int) {
		recycled[i], recycled[j] = recycled[j], recycled[i]
	})

	g.drawPile = recycled
}

// hasPlayableCard プレイヤーが出せるカードを持っているか
func (g *Prsi) hasPlayableCard(playerIdx int) bool {
	return handHasAny(g.players[playerIdx], g.isValidPlay)
}

// HasPlayableCard プレイヤーが出せるカードを持っているか (Web/ヒント用)
func (g *Prsi) HasPlayableCard(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return false
	}
	return g.hasPlayableCard(playerIdx)
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *Prsi) sortAllHands() {
	sortHands(len(g.players), g)
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Prsi) sortHand(playerIdx int) {
	sortPlayerHand(g.players[playerIdx], bySuitThenValue)
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する (-1 = プレイ不可)
func (g *Prsi) cpuSelectPlayCard(playerIdx int) int {
	validIndices := g.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return -1
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch g.config.CpuDifficulty {
	case PrsiCpuDifficultyHard:
		return g.cpuPlayHard(playerIdx, validIndices)
	case PrsiCpuDifficultyNormal:
		return g.cpuPlayNormal(playerIdx, validIndices)
	default:
		return g.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (g *Prsi) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 最も多いスートを優先して出す
func (g *Prsi) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]
	suitCount := g.countSuits(playerIdx)
	bestIdx := validIndices[0]
	bestCount := suitCount[player.GetCard(validIndices[0]).GetDesign()]
	for _, idx := range validIndices[1:] {
		sc := suitCount[player.GetCard(idx).GetDesign()]
		if sc > bestCount {
			bestCount = sc
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 戦略的プレイ: 相手の手札が少ないときはアクションカード(7/Ace/Under)を優先する
func (g *Prsi) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	// 次のプレイヤーの手札が少ないなら、攻撃カードを優先
	nextIdx := (playerIdx + 1) % PrsiPlayerCnt
	if g.players[nextIdx].GetCardsSize() <= 2 {
		for _, idx := range validIndices {
			if g.isActionCard(player.GetCard(idx)) {
				return idx
			}
		}
	}

	// それ以外は最も多いスートを優先 (アクションカードは温存)
	suitCount := g.countSuits(playerIdx)
	bestIdx := validIndices[0]
	bestScore := g.cpuCardPriority(player.GetCard(validIndices[0]), suitCount)
	for _, idx := range validIndices[1:] {
		score := g.cpuCardPriority(player.GetCard(idx), suitCount)
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// isActionCard 7/Ace/Underかどうか
func (g *Prsi) isActionCard(card *Card) bool {
	v := card.GetValue()
	return v == PrsiSevenValue || v == PrsiAceValue || v == PrsiUnderValue
}

// cpuCardPriority カードの優先度スコアを計算 (大きいほど先に出したい)。
// 多いスートを優先しつつ、アクションカードは温存する。
func (g *Prsi) cpuCardPriority(card *Card, suitCount map[int]int) int {
	score := suitCount[card.GetDesign()]
	if g.isActionCard(card) {
		score -= 5 // 温存
	}
	return score
}

// countSuits プレイヤーの手札のスート別枚数をカウント
func (g *Prsi) countSuits(playerIdx int) map[int]int {
	player := g.players[playerIdx]
	counts := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		counts[player.GetCard(i).GetDesign()]++
	}
	return counts
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Prsi) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.isValidPlay(c) })
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *Prsi) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// prsiJSON is the JSON wire format for Prsi.
type prsiJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*PrsiPlayer     `json:"pl"`
	Config           PrsiConfig        `json:"cf"`
	Phase            PrsiPhase         `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	PenaltyDrawCount int               `json:"pd"`
	PendingSkips     int               `json:"sk"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Prsi) MarshalJSON() ([]byte, error) {
	return json.Marshal(prsiJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		PenaltyDrawCount: g.penaltyDrawCount,
		PendingSkips:     g.pendingSkips,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// prsiMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const prsiMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. Hardens against malformed input by
// bounding slice lengths, validating the config and phase range, the
// (fixed) player count, the current-player index, and rejecting nil player or
// card elements that would otherwise panic during play.
func (g *Prsi) UnmarshalJSON(data []byte) error {
	var j prsiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > prsiMaxSliceLen || len(j.DiscardPile) > prsiMaxSliceLen ||
		len(j.DrawPile) > prsiMaxSliceLen || len(j.ActionLog) > prsiMaxSliceLen {
		return fmt.Errorf("prsi: input array exceeds maximum allowed size")
	}
	// Prsi is strictly a 4-player game; reject malformed states that would
	// otherwise cause out-of-bounds panics during play.
	if len(j.Players) != PrsiPlayerCnt {
		return fmt.Errorf("prsi: invalid player count: expected %d, got %d", PrsiPlayerCnt, len(j.Players))
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("prsi: nil player at index %d", i)
		}
	}
	for i, c := range j.DiscardPile {
		if c == nil {
			return fmt.Errorf("prsi: nil discard card at index %d", i)
		}
	}
	for i, c := range j.DrawPile {
		if c == nil {
			return fmt.Errorf("prsi: nil draw card at index %d", i)
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= PrsiPlayerCnt {
		return fmt.Errorf("prsi: currentPlayerIdx %d out of range [0, %d)", j.CurrentPlayerIdx, PrsiPlayerCnt)
	}
	if j.Phase < PrsiPhasePlay || j.Phase > PrsiPhaseGameEnd {
		return fmt.Errorf("prsi: invalid phase: %d", j.Phase)
	}
	// When the game is over the winner must be a real seat; otherwise it must be
	// the -1 sentinel. This keeps presenters/logs from rendering "Player 99".
	if j.GameEndFlag {
		if j.WinnerIdx < 0 || j.WinnerIdx >= PrsiPlayerCnt {
			return fmt.Errorf("prsi: winnerIdx %d out of range [0, %d) with game ended", j.WinnerIdx, PrsiPlayerCnt)
		}
	} else if j.WinnerIdx != -1 {
		return fmt.Errorf("prsi: winnerIdx must be -1 while the game is in progress, got %d", j.WinnerIdx)
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("prsi: invalid config: %w", err)
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsPrsi()
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
	g.penaltyDrawCount = j.PenaltyDrawCount
	if g.penaltyDrawCount < 0 {
		g.penaltyDrawCount = 0
	} else if g.penaltyDrawCount > prsiMaxSliceLen {
		g.penaltyDrawCount = prsiMaxSliceLen
	}
	g.pendingSkips = j.PendingSkips
	if g.pendingSkips < 0 {
		g.pendingSkips = 0
	} else if g.pendingSkips > PrsiPlayerCnt {
		g.pendingSkips = PrsiPlayerCnt
	}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
