//go:build !js || !wasm || classic

// Package domain マリアーシュ (Mariáš) のドメインモデル。
//
// Mariáš はチェコ・スロバキアの 3 人用トリックテイキングゲーム。32 枚デッキから
// 各自 10 枚を配り (残り 2 枚は talon として未使用)、ディーラーの左隣 (forehand) が
// その回の Soloist となり残り 2 人の Defender と対戦する。Soloist は手札の最長スートを
// 切り札として宣言する (本実装では自動)。同スートの K+Q を初手に持つ側は結婚点
// (通常 20 点・切り札 40 点) を得る。マストフォロー＋ボイド時は切り札強制。最終 (10th)
// トリックの勝者に +10 点。1 ラウンドの得点を比較し勝者へゲーム点を加算、累積が目標
// (既定 10) に達したプレイヤーが勝利する。
//
// カードポイント: A=11, 10=10, K=4, Q=3, J=2, 9/8/7=0 (合計 120 点)。
// トリック強度 (切り札・非切り札共通): A > 10 > K > Q > J > 9 > 8 > 7。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// MariasPlayerCnt プレイヤー数 (人間 1 + CPU 2)
const MariasPlayerCnt = 3

// MariasHandSize 各プレイヤーの手札枚数
const MariasHandSize = 10

// MariasTrickCount 1 ラウンドのトリック数
const MariasTrickCount = 10

// MariasLastTrickBonus 最終トリック勝者へのボーナス点
const MariasLastTrickBonus = 10

// MariasMarriagePoints 通常スートの結婚点
const MariasMarriagePoints = 20

// MariasTrumpMarriagePoints 切り札スートの結婚点
const MariasTrumpMarriagePoints = 40

// MariasPhase ゲームフェーズ
type MariasPhase int

// Mariáš のフェーズ定数
const (
	// MariasPhasePlay トリックプレイフェーズ
	MariasPhasePlay MariasPhase = 0
	// MariasPhaseTrickEnd トリック終了フェーズ
	MariasPhaseTrickEnd MariasPhase = 1
	// MariasPhaseRoundEnd ラウンド終了フェーズ
	MariasPhaseRoundEnd MariasPhase = 2
	// MariasPhaseGameEnd ゲーム終了フェーズ
	MariasPhaseGameEnd MariasPhase = 3
)

// MariasHint ヒント情報
type MariasHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Marias マリアーシュのゲームクラス
type Marias struct {
	trumpCards       *TrumpCards
	players          []*MariasPlayer
	config           MariasConfig
	phase            MariasPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	soloistIdx       int                  // その回の Soloist
	trumpSuit        int                  // 切り札スート
	playerScores     [MariasPlayerCnt]int // 累積ゲーム点
	roundCardPts     [MariasPlayerCnt]int // 現ラウンドのプレイヤー別カード得点
	roundMarriage    [MariasPlayerCnt]int // 現ラウンドのプレイヤー別結婚点
	lastTrickWinner  int                  // 最終トリック勝者 (-1=未確定)
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLog        []*ActionLogEntry
}

// NewMarias コンストラクタ
func NewMarias(trumpCards *TrumpCards, players []*MariasPlayer, config MariasConfig) *Marias {
	return &Marias{trumpCards: trumpCards, players: players, config: config, winnerPlayer: -1, lastTrickWinner: -1}
}

// NewDefaultMarias 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultMarias() *Marias {
	players := make([]*MariasPlayer, MariasPlayerCnt)
	players[0] = NewMariasPlayer(true)
	for i := 1; i < MariasPlayerCnt; i++ {
		players[i] = NewMariasPlayer(false)
	}
	return NewMarias(NewTrumpCardsBelote(), players, DefaultMariasConfig())
}

// Reset ゲーム初期化
func (g *Marias) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [MariasPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Marias) NextRound() {
	if g.phase != MariasPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % MariasPlayerCnt
	g.startRound()
}

// startRound 手札を配り、Soloist・切り札・結婚を決めてプレイフェーズを開始する。
func (g *Marias) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundCardPts = [MariasPlayerCnt]int{}
	g.roundMarriage = [MariasPlayerCnt]int{}
	g.lastTrickWinner = -1
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	// Soloist = ディーラーの左隣 (forehand)。切り札は Soloist の最長スート。
	g.soloistIdx = (g.dealerIdx + 1) % MariasPlayerCnt
	g.trumpSuit = g.longestSuit(g.soloistIdx)
	g.sortAllHands()
	g.detectMarriages()

	g.leadPlayerIdx = g.soloistIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = MariasPhasePlay
}

// deal 各プレイヤーへ 10 枚を配る (残りは talon として未使用)。
func (g *Marias) deal() {
	for i := 0; i < MariasHandSize; i++ {
		for j := 0; j < MariasPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % MariasPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// longestSuit プレイヤーが最も多く持つスートを返す (同数ならスート番号が小さい方)。
func (g *Marias) longestSuit(playerIdx int) int {
	counts := map[int]int{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	bestSuit, bestCnt := CardDesignSpade, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if counts[suit] > bestCnt {
			bestCnt = counts[suit]
			bestSuit = suit
		}
	}
	return bestSuit
}

// detectMarriages 各プレイヤーの初手から結婚 (同スート K+Q) を検出し加点する。
func (g *Marias) detectMarriages() {
	for i := range g.players {
		pts := 0
		for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
			if g.playerHasCard(i, suit, 13) && g.playerHasCard(i, suit, 12) {
				if suit == g.trumpSuit {
					pts += MariasTrumpMarriagePoints
				} else {
					pts += MariasMarriagePoints
				}
			}
		}
		if pts > 0 {
			g.roundMarriage[i] += pts
			g.appendLog(i, "marriage", fmt.Sprintf("%s declares marriages worth %d", g.playerName(i), pts), nil)
		}
	}
}

// playerHasCard プレイヤーが指定スート・ランクの札を持っているか。
func (g *Marias) playerHasCard(playerIdx, design, value int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && c.GetValue() == value {
			return true
		}
	}
	return false
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Marias) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MariasPhasePlay {
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
func (g *Marias) CpuPlay() {
	if g.gameEndFlag || g.phase != MariasPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *Marias) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == MariasPlayerCnt {
		g.phase = MariasPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % MariasPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Marias) ResolveTrick() {
	if g.phase != MariasPhaseTrickEnd || len(g.currentTrick) != MariasPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += mariasCardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.roundCardPts[winnerIdx] += pts
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d)", g.playerName(winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= MariasTrickCount {
		g.lastTrickWinner = winnerIdx
		g.roundCardPts[winnerIdx] += MariasLastTrickBonus
		g.phase = MariasPhaseRoundEnd
	} else {
		g.phase = MariasPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Marias) NextTrick() {
	if g.phase != MariasPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = MariasPhasePlay
}

// ScoreRound ラウンド結果を判定し、勝者へゲーム点を加算してマッチ終了を判定する。
func (g *Marias) ScoreRound() {
	if g.phase != MariasPhaseRoundEnd {
		return
	}
	soloistTotal, defenseTotal := g.sideTotals()
	soloistWon := soloistTotal > defenseTotal
	if soloistWon {
		// Soloist 勝利: Soloist へ 2 点。
		g.playerScores[g.soloistIdx] += 2
	} else {
		// Soloist 敗北: 各 Defender へ 2 点 (倍付け)。
		for i := 0; i < MariasPlayerCnt; i++ {
			if i != g.soloistIdx {
				g.playerScores[i] += 2
			}
		}
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: soloist(%s)=%d defense=%d -> %s",
			g.roundNumber, g.playerName(g.soloistIdx), soloistTotal, defenseTotal,
			map[bool]string{true: "soloist wins", false: "defense wins"}[soloistWon]), nil)

	g.checkGameEnd()
}

// sideTotals Soloist 側と Defender 側のラウンド合計点を返す。
func (g *Marias) sideTotals() (int, int) {
	soloist, defense := 0, 0
	for i := 0; i < MariasPlayerCnt; i++ {
		total := g.roundCardPts[i] + g.roundMarriage[i]
		if i == g.soloistIdx {
			soloist += total
		} else {
			defense += total
		}
	}
	return soloist, defense
}

// checkGameEnd 目標点到達でマッチ終了を判定する。
func (g *Marias) checkGameEnd() {
	leader, best := -1, -1
	for i := 0; i < MariasPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerPlayer = leader
		g.phase = MariasPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", g.playerName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォロー + ボイド時の切り札強制を検証する。
func (g *Marias) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	hasLeadSuit := g.playerHasSuit(playerIdx, leadSuit)
	if hasLeadSuit && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	if !hasLeadSuit && g.playerHasSuit(playerIdx, g.trumpSuit) && card.GetDesign() != g.trumpSuit {
		return NewDomainError(ErrInvalidPlay, "切り札を出してください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Marias) playerHasSuit(playerIdx, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *Marias) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.mariasRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.mariasRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// mariasRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Marias) mariasRank(card *Card) int {
	if card.GetDesign() == g.trumpSuit {
		return 100 + mariasStrength(card.GetValue())
	}
	return mariasStrength(card.GetValue())
}

// mariasStrength カード強度。A > 10 > K > Q > J > 9 > 8 > 7。
func mariasStrength(value int) int {
	switch value {
	case 1: // Ace
		return 8
	case 10:
		return 7
	case 13: // King
		return 6
	case 12: // Queen
		return 5
	case 11: // Jack
		return 4
	case 9:
		return 3
	case 8:
		return 2
	default: // 7
		return 1
	}
}

// mariasCardPoints カードポイント。A=11, 10=10, K=4, Q=3, J=2, 他=0。
func mariasCardPoints(card *Card) int {
	switch card.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	default:
		return 0
	}
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Marias) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Marias) sortAllHands() {
	for _, p := range g.players {
		mariasSortHand(p)
	}
}

// mariasSortHand 手札をスート→強さ順にソートする。
func mariasSortHand(p *MariasPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return mariasStrength(cards[i].GetValue()) > mariasStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Marias) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Marias) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Marias) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.mariasRank(g.currentTrick[idx].Card)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Marias) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// appendLog 棋譜にエントリを追加する。
func (g *Marias) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Marias) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == MariasCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点を意識した戦略プレイ。
func (g *Marias) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return g.minBy(player, valid, func(c *Card) int {
			return mariasCardPoints(c)*100 + g.mariasRank(c)
		})
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += mariasCardPoints(tc.Card)
	}
	winners := mariasFilter(valid, func(idx int) bool { return g.mariasRank(player.GetCard(idx)) > topRank })
	if trickPts > 0 && len(winners) > 0 {
		return g.minBy(player, winners, func(c *Card) int { return g.mariasRank(c) })
	}
	return g.minBy(player, valid, func(c *Card) int {
		return mariasCardPoints(c)*100 + g.mariasRank(c)
	})
}

// minBy score が最小となるインデックスを返す。
func (g *Marias) minBy(player *MariasPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// mariasFilter 述語を満たすインデックスを抽出する。
func mariasFilter(indices []int, pred func(int) bool) []int {
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
func (g *Marias) GetHint() *MariasHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != MariasPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &MariasHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Marias) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.mariasRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Marias) GetPhase() MariasPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Marias) SetPhase(phase MariasPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Marias) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Marias) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Marias) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Marias) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Marias) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Marias) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Marias) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Marias) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Marias) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Marias) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Marias) GetDealerIdx() int { return g.dealerIdx }

// GetSoloistIdx Soloist インデックス取得
func (g *Marias) GetSoloistIdx() int { return g.soloistIdx }

// SetSoloistIdx Soloist インデックス設定 (テスト用)
func (g *Marias) SetSoloistIdx(idx int) { g.soloistIdx = idx }

// GetTrumpSuit 切り札スート取得
func (g *Marias) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Marias) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetPlayerScores プレイヤー別累積点取得
func (g *Marias) GetPlayerScores() [MariasPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Marias) SetPlayerScores(s [MariasPlayerCnt]int) { g.playerScores = s }

// GetRoundCardPoints 現ラウンドのカード得点取得
func (g *Marias) GetRoundCardPoints() [MariasPlayerCnt]int { return g.roundCardPts }

// SetRoundCardPoints 現ラウンドのカード得点設定 (テスト用)
func (g *Marias) SetRoundCardPoints(s [MariasPlayerCnt]int) { g.roundCardPts = s }

// GetRoundMarriage 現ラウンドの結婚点取得
func (g *Marias) GetRoundMarriage() [MariasPlayerCnt]int { return g.roundMarriage }

// SetRoundMarriage 現ラウンドの結婚点設定 (テスト用)
func (g *Marias) SetRoundMarriage(s [MariasPlayerCnt]int) { g.roundMarriage = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Marias) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Marias) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Marias) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Marias) GetPlayer(i int) *MariasPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か。
func (g *Marias) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Marias) GetConfig() MariasConfig { return g.config }

// SetConfig 設定変更
func (g *Marias) SetConfig(cfg MariasConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Marias) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Marias) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != MariasPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// mariasJSON is the JSON wire format for Marias.
type mariasJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*MariasPlayer      `json:"ps"`
	Config           MariasConfig         `json:"cf"`
	Phase            MariasPhase          `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	LeadPlayerIdx    int                  `json:"li"`
	DealerIdx        int                  `json:"di"`
	SoloistIdx       int                  `json:"so"`
	TrumpSuit        int                  `json:"ts"`
	PlayerScores     [MariasPlayerCnt]int `json:"sc"`
	RoundCardPts     [MariasPlayerCnt]int `json:"rp"`
	RoundMarriage    [MariasPlayerCnt]int `json:"rm"`
	LastTrickWinner  int                  `json:"lt"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerPlayer     int                  `json:"wp"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Marias) MarshalJSON() ([]byte, error) {
	return json.Marshal(mariasJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		SoloistIdx:       g.soloistIdx,
		TrumpSuit:        g.trumpSuit,
		PlayerScores:     g.playerScores,
		RoundCardPts:     g.roundCardPts,
		RoundMarriage:    g.roundMarriage,
		LastTrickWinner:  g.lastTrickWinner,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// mariasMaxSliceLen caps slice sizes during deserialisation.
const mariasMaxSliceLen = 5000

// errMariasOversized is the single sentinel error for oversized input arrays.
var errMariasOversized = errors.New("marias: input array exceeds maximum allowed size")

// errMariasInvalidPlayers is returned when restored state lacks exactly MariasPlayerCnt players.
var errMariasInvalidPlayers = errors.New("marias: invalid player count")

// errMariasInvalidTrick is returned when a restored trick card or its card is nil.
var errMariasInvalidTrick = errors.New("marias: invalid trick card")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Marias) UnmarshalJSON(data []byte) error {
	var j mariasJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > mariasMaxSliceLen || len(j.CurrentTrick) > mariasMaxSliceLen ||
		len(j.ActionLog) > mariasMaxSliceLen {
		return errMariasOversized
	}
	if len(j.Players) != MariasPlayerCnt {
		return errMariasInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errMariasInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errMariasInvalidTrick
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBelote()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.soloistIdx = j.SoloistIdx
	g.trumpSuit = j.TrumpSuit
	g.playerScores = j.PlayerScores
	g.roundCardPts = j.RoundCardPts
	g.roundMarriage = j.RoundMarriage
	g.lastTrickWinner = j.LastTrickWinner
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
