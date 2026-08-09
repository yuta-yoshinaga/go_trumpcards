//go:build !js || !wasm || classic

// Package domain ノックアウト・ホイスト (Knockout Whist) のドメインモデル。
//
// Knockout Whist はイギリスのサバイバル型トリックテイキング。52 枚デッキを使い、
// 第 1 ラウンドは 7 枚、以降ラウンドごとに 1 枚ずつ減らして配る (8 - ラウンド番号)。
// 前ラウンドの勝者 (最多トリック獲得者) が次ラウンドの切り札スートを選ぶ (本実装では
// 勝者の最長スートを自動選択)。マストフォローでトリックを消化し、1 トリックも取れな
// かったプレイヤーは Dogbone (猶予トークン、初期 1 枚) を消費して脱落を回避する。
// Dogbone を使い切った状態で再び 0 トリックなら脱落。最後に残った 1 人が優勝。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// KnockoutWhistPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const KnockoutWhistPlayerCnt = 4

// KnockoutWhistStartingDogbones 各プレイヤーの初期 Dogbone 数
const KnockoutWhistStartingDogbones = 1

// KnockoutWhistMaxRounds 最大ラウンド数 (7 枚→1 枚)
const KnockoutWhistMaxRounds = 7

// KnockoutWhistPhase ゲームフェーズ
type KnockoutWhistPhase int

// Knockout Whist のフェーズ定数
const (
	// KnockoutWhistPhasePlay トリックプレイフェーズ
	KnockoutWhistPhasePlay KnockoutWhistPhase = 0
	// KnockoutWhistPhaseTrickEnd トリック終了フェーズ
	KnockoutWhistPhaseTrickEnd KnockoutWhistPhase = 1
	// KnockoutWhistPhaseRoundEnd ラウンド終了フェーズ
	KnockoutWhistPhaseRoundEnd KnockoutWhistPhase = 2
	// KnockoutWhistPhaseGameEnd ゲーム終了フェーズ
	KnockoutWhistPhaseGameEnd KnockoutWhistPhase = 3
	// KnockoutWhistPhaseTrumpSelect 切り札選択フェーズ (人間のラウンド勝者が次ラウンドの切り札を選ぶ)
	KnockoutWhistPhaseTrumpSelect KnockoutWhistPhase = 4
)

// KnockoutWhistHint ヒント情報
type KnockoutWhistHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// KnockoutWhist ノックアウト・ホイストのゲームクラス
type KnockoutWhist struct {
	trumpCards       *TrumpCards
	players          []*KnockoutWhistPlayer
	config           KnockoutWhistConfig
	phase            KnockoutWhistPhase
	roundNumber      int
	handSize         int // 現ラウンドの配り枚数 (8 - roundNumber)
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int
	roundWinnerIdx   int // 直近ラウンドの勝者 (次ラウンドのリード/切り札決定者)
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLogBase
}

// NewKnockoutWhist コンストラクタ
func NewKnockoutWhist(trumpCards *TrumpCards, players []*KnockoutWhistPlayer, config KnockoutWhistConfig) *KnockoutWhist {
	return &KnockoutWhist{trumpCards: trumpCards, players: players, config: config, winnerPlayer: -1, roundWinnerIdx: -1}
}

// NewDefaultKnockoutWhist 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultKnockoutWhist() *KnockoutWhist {
	players := make([]*KnockoutWhistPlayer, KnockoutWhistPlayerCnt)
	players[0] = NewKnockoutWhistPlayer(true)
	for i := 1; i < KnockoutWhistPlayerCnt; i++ {
		players[i] = NewKnockoutWhistPlayer(false)
	}
	return NewKnockoutWhist(NewTrumpCards(0), players, DefaultKnockoutWhistConfig())
}

// activeCount 残存 (未脱落) プレイヤー数を返す。
func (g *KnockoutWhist) activeCount() int {
	n := 0
	for _, p := range g.players {
		if !p.GetEliminated() {
			n++
		}
	}
	return n
}

// nextActive from の次の未脱落プレイヤーのインデックスを返す (from 自身は除く)。
func (g *KnockoutWhist) nextActive(from int) int {
	for k := 1; k <= KnockoutWhistPlayerCnt; k++ {
		ni := (from + k) % KnockoutWhistPlayerCnt
		if !g.players[ni].GetEliminated() {
			return ni
		}
	}
	return from
}

// firstActiveFrom from を含めて最初の未脱落プレイヤーを返す。
func (g *KnockoutWhist) firstActiveFrom(from int) int {
	for k := 0; k < KnockoutWhistPlayerCnt; k++ {
		ni := (from + k) % KnockoutWhistPlayerCnt
		if !g.players[ni].GetEliminated() {
			return ni
		}
	}
	return from
}

// Reset ゲーム初期化
func (g *KnockoutWhist) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	for _, p := range g.players {
		p.SetEliminated(false)
		p.SetDogbones(KnockoutWhistStartingDogbones)
	}
	g.actionLog = nil
	// 第 1 ラウンドのリードはディーラーの左隣。
	g.leadPlayerIdx = g.firstActiveFrom((g.dealerIdx + 1) % KnockoutWhistPlayerCnt)
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *KnockoutWhist) NextRound() {
	if g.phase != KnockoutWhistPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % KnockoutWhistPlayerCnt
	// 次ラウンドのリードは直近ラウンドの勝者 (未脱落なら)。
	if g.roundWinnerIdx >= 0 && !g.players[g.roundWinnerIdx].GetEliminated() {
		g.leadPlayerIdx = g.roundWinnerIdx
	} else {
		g.leadPlayerIdx = g.firstActiveFrom((g.dealerIdx + 1) % KnockoutWhistPlayerCnt)
	}
	g.startRound()
}

// startRound 手札を配り、切り札を決めてプレイフェーズを開始する。
func (g *KnockoutWhist) startRound() {
	g.handSize = KnockoutWhistMaxRounds + 1 - g.roundNumber // round1=7 ... round7=1
	if g.handSize < 1 {
		g.handSize = 1
	}
	g.trickNumber = 1
	g.currentTrick = nil
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.currentPlayerIdx = g.leadPlayerIdx

	// 切り札はラウンド勝者 (= リードプレイヤー) が選ぶ。人間勝者は TrumpSelect フェーズで
	// 対話的に選択する。CPU 勝者と第 1 ラウンド (前ラウンド勝者なし) は最長スートを自動選択。
	if g.roundNumber > 1 && g.players[g.leadPlayerIdx].GetIsHuman() {
		g.phase = KnockoutWhistPhaseTrumpSelect
		g.appendLog(g.leadPlayerIdx, "trump_select",
			fmt.Sprintf("round %d: %d cards each, %s chooses trump",
				g.roundNumber, g.handSize, playerName(g.players, g.leadPlayerIdx)), nil)
		return
	}

	g.trumpSuit = g.longestSuit(g.leadPlayerIdx)
	g.phase = KnockoutWhistPhasePlay
	g.appendLog(g.leadPlayerIdx, "round_start",
		fmt.Sprintf("round %d: %d cards each, trump %d, %s leads",
			g.roundNumber, g.handSize, g.trumpSuit, playerName(g.players, g.leadPlayerIdx)), nil)
}

// PlayerSelectTrump 人間のラウンド勝者が次ラウンドの切り札スートを選択する (1-4)。
func (g *KnockoutWhist) PlayerSelectTrump(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KnockoutWhistPhaseTrumpSelect {
		return ErrWrongPhase
	}
	if !g.players[g.leadPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "切り札スートが範囲外です")
	}
	g.trumpSuit = suit
	g.phase = KnockoutWhistPhasePlay
	g.appendLog(g.leadPlayerIdx, "round_start",
		fmt.Sprintf("round %d: %d cards each, trump %d, %s leads",
			g.roundNumber, g.handSize, g.trumpSuit, playerName(g.players, g.leadPlayerIdx)), nil)
	return nil
}

// deal 未脱落プレイヤーへ handSize 枚ずつ配る。
func (g *KnockoutWhist) deal() {
	for i := 0; i < g.handSize; i++ {
		idx := g.leadPlayerIdx
		for j := 0; j < KnockoutWhistPlayerCnt; j++ {
			if !g.players[idx].GetEliminated() {
				if c := g.trumpCards.DrawCard(); c != nil {
					g.players[idx].AddCard(c)
				}
			}
			idx = (idx + 1) % KnockoutWhistPlayerCnt
		}
	}
}

// longestSuit プレイヤーが最も多く持つスートを返す。
func (g *KnockoutWhist) longestSuit(playerIdx int) int {
	return longestSuit(g.players[playerIdx])
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *KnockoutWhist) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KnockoutWhistPhasePlay {
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
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *KnockoutWhist) CpuPlay() {
	if g.gameEndFlag || g.phase != KnockoutWhistPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *KnockoutWhist) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) >= g.activeCount() {
		g.phase = KnockoutWhistPhaseTrickEnd
	} else {
		g.currentPlayerIdx = g.nextActive(g.currentPlayerIdx)
	}
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *KnockoutWhist) ResolveTrick() {
	if g.phase != KnockoutWhistPhaseTrickEnd || len(g.currentTrick) < g.activeCount() {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.players[winnerIdx].IncRoundTricks()
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= g.handSize {
		g.phase = KnockoutWhistPhaseRoundEnd
	} else {
		g.phase = KnockoutWhistPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *KnockoutWhist) NextTrick() {
	if g.phase != KnockoutWhistPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = KnockoutWhistPhasePlay
}

// ScoreRound ラウンド結果を判定し、0 トリックのプレイヤーを脱落/Dogbone 処理して
// マッチ終了を判定する。
func (g *KnockoutWhist) ScoreRound() {
	if g.phase != KnockoutWhistPhaseRoundEnd {
		return
	}
	// ラウンド勝者 = 最多トリック獲得者 (次ラウンドのリード/切り札決定者)。
	g.roundWinnerIdx = g.mostTricksPlayer()

	// 0 トリックの未脱落プレイヤーを処理。
	for i, p := range g.players {
		if p.GetEliminated() {
			continue
		}
		if p.GetRoundTricks() == 0 {
			if p.GetDogbones() > 0 {
				p.SetDogbones(p.GetDogbones() - 1)
				g.appendLog(i, "dogbone", fmt.Sprintf("%s spends a dogbone to survive", playerName(g.players, i)), nil)
			} else {
				p.SetEliminated(true)
				g.appendLog(i, "eliminated", fmt.Sprintf("%s is knocked out", playerName(g.players, i)), nil)
			}
		}
	}

	// 終了判定: 残り 1 人以下、または最終 (1 枚) ラウンド消化。
	if g.activeCount() <= 1 || g.roundNumber >= KnockoutWhistMaxRounds {
		g.gameEndFlag = true
		g.phase = KnockoutWhistPhaseGameEnd
		if g.activeCount() == 1 {
			g.winnerPlayer = g.firstActiveFrom(0)
		} else {
			// 全滅 (同時 0 トリック) や最終ラウンド到達時はラウンド勝者を優勝とする。
			g.winnerPlayer = g.roundWinnerIdx
		}
		g.appendLog(g.winnerPlayer, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, g.winnerPlayer)), nil)
	}
}

// mostTricksPlayer 現ラウンドで最多トリックを取った未脱落プレイヤーを返す。
func (g *KnockoutWhist) mostTricksPlayer() int {
	best, bestIdx := -1, -1
	for i, p := range g.players {
		if p.GetEliminated() {
			continue
		}
		if p.GetRoundTricks() > best {
			best = p.GetRoundTricks()
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return g.firstActiveFrom(0)
	}
	return bestIdx
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する。
func (g *KnockoutWhist) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if g.playerHasSuit(playerIdx, leadSuit) && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *KnockoutWhist) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(g.players[playerIdx], design)
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *KnockoutWhist) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.knockoutRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.knockoutRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// knockoutRank トリック比較用ランク。切り札は非切り札より常に強い。A 高。
func (g *KnockoutWhist) knockoutRank(card *Card) int {
	base := knockoutCardStrength(card.GetValue())
	if card.GetDesign() == g.trumpSuit {
		return 100 + base
	}
	return base
}

// knockoutCardStrength カード強度 (A 高: A>K>Q>J>10>...>2)。
func knockoutCardStrength(value int) int {
	if value == 1 {
		return 14
	}
	return value
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *KnockoutWhist) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *KnockoutWhist) sortAllHands() {
	for _, p := range g.players {
		knockoutSortHand(p)
	}
}

// knockoutSortHand 手札をスート→強さ順にソートする。
func knockoutSortHand(p *KnockoutWhistPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return knockoutCardStrength(cards[i].GetValue()) > knockoutCardStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *KnockoutWhist) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *KnockoutWhist) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.knockoutRank(g.currentTrick[idx].Card)
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *KnockoutWhist) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == KnockoutWhistCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart トリックを取りに行く戦略プレイ (生存のためトリックは貴重)。
func (g *KnockoutWhist) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		// リード: 強い札で主導権を取る。
		return pickHighest(player, valid, func(c *Card) int { return g.knockoutRank(c) })
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	winners := knockoutFilter(valid, func(idx int) bool { return g.knockoutRank(player.GetCard(idx)) > topRank })
	if len(winners) > 0 {
		// 最小コストで勝てる札。
		return pickLowest(player, winners, func(c *Card) int { return g.knockoutRank(c) })
	}
	// 勝てない: 最弱札を捨てる。
	return pickLowest(player, valid, func(c *Card) int { return g.knockoutRank(c) })
}

// knockoutFilter 述語を満たすインデックスを抽出する。
func knockoutFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨プレイを返す。
func (g *KnockoutWhist) GetHint() *KnockoutWhistHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != KnockoutWhistPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &KnockoutWhistHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *KnockoutWhist) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_high"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.knockoutRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *KnockoutWhist) GetPhase() KnockoutWhistPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *KnockoutWhist) SetPhase(phase KnockoutWhistPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *KnockoutWhist) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *KnockoutWhist) SetRoundNumber(n int) { g.roundNumber = n }

// GetHandSize 現ラウンドの配り枚数取得
func (g *KnockoutWhist) GetHandSize() int { return g.handSize }

// SetHandSize 配り枚数設定 (テスト用)
func (g *KnockoutWhist) SetHandSize(n int) { g.handSize = n }

// GetTrickNumber トリック番号取得
func (g *KnockoutWhist) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *KnockoutWhist) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *KnockoutWhist) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *KnockoutWhist) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *KnockoutWhist) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *KnockoutWhist) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *KnockoutWhist) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *KnockoutWhist) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *KnockoutWhist) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit 切り札スート取得
func (g *KnockoutWhist) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *KnockoutWhist) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetRoundWinnerIdx 直近ラウンド勝者取得 (-1=未確定)
func (g *KnockoutWhist) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *KnockoutWhist) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *KnockoutWhist) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *KnockoutWhist) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *KnockoutWhist) GetPlayer(i int) *KnockoutWhistPlayer {
	return getPlayer(g.players, i)
}

// GetActiveCount 残存プレイヤー数を返す。
func (g *KnockoutWhist) GetActiveCount() int { return g.activeCount() }

// IsHumanTurn 現在の手番が人間か。
func (g *KnockoutWhist) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *KnockoutWhist) GetConfig() KnockoutWhistConfig { return g.config }

// SetConfig 設定変更
func (g *KnockoutWhist) SetConfig(cfg KnockoutWhistConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *KnockoutWhist) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != KnockoutWhistPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// knockoutWhistJSON is the JSON wire format for KnockoutWhist.
type knockoutWhistJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*KnockoutWhistPlayer `json:"ps"`
	Config           KnockoutWhistConfig    `json:"cf"`
	Phase            KnockoutWhistPhase     `json:"ph"`
	RoundNumber      int                    `json:"rn"`
	HandSize         int                    `json:"hs"`
	TrickNumber      int                    `json:"tn"`
	CurrentPlayerIdx int                    `json:"ci"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	LeadPlayerIdx    int                    `json:"li"`
	DealerIdx        int                    `json:"di"`
	TrumpSuit        int                    `json:"ts"`
	RoundWinnerIdx   int                    `json:"rw"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerPlayer     int                    `json:"wp"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *KnockoutWhist) MarshalJSON() ([]byte, error) {
	return json.Marshal(knockoutWhistJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		HandSize:         g.handSize,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		TrumpSuit:        g.trumpSuit,
		RoundWinnerIdx:   g.roundWinnerIdx,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// knockoutWhistMaxSliceLen caps slice sizes during deserialisation.
const knockoutWhistMaxSliceLen = 5000

// errKnockoutWhistOversized is the single sentinel error for oversized input arrays.
var errKnockoutWhistOversized = errors.New("knockoutwhist: input array exceeds maximum allowed size")

// errKnockoutWhistInvalidPlayers is returned when restored state lacks exactly KnockoutWhistPlayerCnt players.
var errKnockoutWhistInvalidPlayers = errors.New("knockoutwhist: invalid player count")

// errKnockoutWhistInvalidTrick is returned when a restored trick card or its card is nil.
var errKnockoutWhistInvalidTrick = errors.New("knockoutwhist: invalid trick card")

// errKnockoutWhistInvalidState is returned when a restored index/state field is out of range.
var errKnockoutWhistInvalidState = errors.New("knockoutwhist: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
func (g *KnockoutWhist) UnmarshalJSON(data []byte) error {
	var j knockoutWhistJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > knockoutWhistMaxSliceLen || len(j.CurrentTrick) > knockoutWhistMaxSliceLen ||
		len(j.ActionLog) > knockoutWhistMaxSliceLen {
		return errKnockoutWhistOversized
	}
	if len(j.Players) != KnockoutWhistPlayerCnt {
		return errKnockoutWhistInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errKnockoutWhistInvalidPlayers
		}
	}
	// Bounds-check restored indices/state so corrupted or tampered KV data
	// cannot trigger out-of-range panics or negative-modulo turn advancement.
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= KnockoutWhistPlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= KnockoutWhistPlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= KnockoutWhistPlayerCnt ||
		j.RoundWinnerIdx < -1 || j.RoundWinnerIdx >= KnockoutWhistPlayerCnt ||
		j.WinnerPlayer < -1 || j.WinnerPlayer >= KnockoutWhistPlayerCnt ||
		j.TrumpSuit < 1 || j.TrumpSuit > 4 ||
		j.RoundNumber < 1 || j.RoundNumber > KnockoutWhistMaxRounds ||
		j.HandSize < 1 || j.HandSize > KnockoutWhistMaxRounds ||
		j.TrickNumber < 1 || j.TrickNumber > KnockoutWhistMaxRounds ||
		j.Phase < KnockoutWhistPhasePlay || j.Phase > KnockoutWhistPhaseTrumpSelect {
		return errKnockoutWhistInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= KnockoutWhistPlayerCnt {
			return errKnockoutWhistInvalidTrick
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.handSize = j.HandSize
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.trumpSuit = j.TrumpSuit
	g.roundWinnerIdx = j.RoundWinnerIdx
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
