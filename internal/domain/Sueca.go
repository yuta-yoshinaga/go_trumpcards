//go:build !js || !wasm || casino

// Package domain スエカ (Sueca) のドメインモデル。
//
// Sueca はポルトガル・ブラジルで最も広くプレイされる 40 枚ラテンデッキの
// トリックテイキングゲーム。切り札あり・マストフォローで 4 人 2 チーム
// (席 0&2 vs 1&3) が 10 トリックを戦い、A=11・7=10 という独特の点配分で 120 点を
// 争う。各ラウンドで 61 点以上を取ったチームが、点差に応じて 1/2/4 ゲームポイントを
// 獲得し、目標 (既定 4) に先に達したチームがマッチ勝利。
//
// カードの強さ (トリック): A > 7 > K > J(Sota) > Q(Caballo) > 6 > 5 > 4 > 3 > 2
// カードポイント: A=11, 7=10, K=4, J=3, Q=2, それ以外=0 (1スート30点 × 4 = 120点)
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// SuecaPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const SuecaPlayerCnt = 4

// SuecaTeamCnt チーム数
const SuecaTeamCnt = 2

// SuecaHandSize 各プレイヤーの手札枚数 (40 / 4)
const SuecaHandSize = 10

// SuecaTrickCount 1 ラウンドのトリック数
const SuecaTrickCount = 10

// SuecaPhase ゲームフェーズ
type SuecaPhase int

// Sueca のフェーズ定数
const (
	// SuecaPhasePlay トリックプレイフェーズ
	SuecaPhasePlay SuecaPhase = 0
	// SuecaPhaseTrickEnd トリック終了フェーズ
	SuecaPhaseTrickEnd SuecaPhase = 1
	// SuecaPhaseRoundEnd ラウンド終了フェーズ
	SuecaPhaseRoundEnd SuecaPhase = 2
	// SuecaPhaseGameEnd ゲーム終了フェーズ
	SuecaPhaseGameEnd SuecaPhase = 3
)

// SuecaHint ヒント情報
type SuecaHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Sueca スエカのゲームクラス
type Sueca struct {
	trumpCards       *TrumpCards
	players          []*SuecaPlayer
	config           SuecaConfig
	phase            SuecaPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int               // 切り札スート (最後に配った札のスート)
	teamGamePts      [SuecaTeamCnt]int // 累積ゲームポイント (jogos)
	roundCardPts     [SuecaTeamCnt]int // 現ラウンドのカード得点
	roundWinnerTeam  int               // 直近ラウンドの勝者 (-1=引き分け/未確定)
	roundGamePts     int               // 直近ラウンドで勝者が得たゲームポイント
	gameEndFlag      bool
	winnerTeam       int // -1=未確定
	actionLog        []*ActionLogEntry
}

// NewSueca コンストラクタ
func NewSueca(trumpCards *TrumpCards, players []*SuecaPlayer, config SuecaConfig) *Sueca {
	return &Sueca{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1, roundWinnerTeam: -1}
}

// NewDefaultSueca 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultSueca() *Sueca {
	players := make([]*SuecaPlayer, SuecaPlayerCnt)
	players[0] = NewSuecaPlayer(true)
	for i := 1; i < SuecaPlayerCnt; i++ {
		players[i] = NewSuecaPlayer(false)
	}
	return NewSueca(NewTrumpCardsBriscola(), players, DefaultSuecaConfig())
}

// SuecaTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func SuecaTeamOf(playerIdx int) int { return playerIdx % SuecaTeamCnt }

// Reset ゲーム初期化
func (g *Sueca) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamGamePts = [SuecaTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Sueca) NextRound() {
	if g.phase != SuecaPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % SuecaPlayerCnt
	g.startRound()
}

// startRound 手札を配り、切り札を決めてプレイフェーズを開始する。
func (g *Sueca) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundCardPts = [SuecaTeamCnt]int{}
	g.roundWinnerTeam = -1
	g.roundGamePts = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.leadPlayerIdx = (g.dealerIdx + 1) % SuecaPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = SuecaPhasePlay
}

// deal 各プレイヤーへ 10 枚を配り、最後に配った札のスートを切り札とする。
func (g *Sueca) deal() {
	var last *Card
	for i := 0; i < SuecaHandSize; i++ {
		for j := 0; j < SuecaPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % SuecaPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
				last = c
			}
		}
	}
	if last != nil {
		g.trumpSuit = last.GetDesign()
	}
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Sueca) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SuecaPhasePlay {
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
func (g *Sueca) CpuPlay() {
	if g.gameEndFlag || g.phase != SuecaPhasePlay {
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
func (g *Sueca) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == SuecaPlayerCnt {
		g.phase = SuecaPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SuecaPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Sueca) ResolveTrick() {
	if g.phase != SuecaPhaseTrickEnd || len(g.currentTrick) != SuecaPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += suecaCardPoints(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.roundCardPts[SuecaTeamOf(winnerIdx)] += pts
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d)", g.playerName(winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	// Keep currentTrick intact through TrickEnd so the resolved trick stays
	// visible; NextTrick clears it before the next trick begins. (#2483 review)
	if g.trickNumber >= SuecaTrickCount {
		g.phase = SuecaPhaseRoundEnd
	} else {
		g.phase = SuecaPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Sueca) NextTrick() {
	if g.phase != SuecaPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = SuecaPhasePlay
}

// ScoreRound ラウンド得点 (120点) を集計し、ゲームポイントを付与してマッチ終了を判定する。
// 勝者は 61-90 点で 1、91-119 点で 2、120 点 (全取り) で 4 ゲームポイント。60-60 は引き分け。
func (g *Sueca) ScoreRound() {
	if g.phase != SuecaPhaseRoundEnd {
		return
	}
	a, b := g.roundCardPts[0], g.roundCardPts[1]
	switch {
	case a > b:
		g.roundWinnerTeam = 0
		g.roundGamePts = suecaGamePoints(a)
	case b > a:
		g.roundWinnerTeam = 1
		g.roundGamePts = suecaGamePoints(b)
	default:
		g.roundWinnerTeam = -1 // 60-60 引き分け
		g.roundGamePts = 0
	}
	if g.roundWinnerTeam >= 0 {
		g.teamGamePts[g.roundWinnerTeam] += g.roundGamePts
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: cards A=%d B=%d → Team %s +%d game pts (total A=%d B=%d)",
			g.roundNumber, a, b, suecaTeamLabel(g.roundWinnerTeam), g.roundGamePts, g.teamGamePts[0], g.teamGamePts[1]), nil)

	leader, other := 0, 1
	if g.teamGamePts[1] > g.teamGamePts[0] {
		leader, other = 1, 0
	}
	if g.teamGamePts[leader] >= g.config.TargetGamePoints && g.teamGamePts[leader] > g.teamGamePts[other] {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = SuecaPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the match!", teamName(leader)), nil)
	}
}

// suecaGamePoints カード得点からゲームポイントを求める。61-90=1, 91-119=2, 120=4。
func suecaGamePoints(cardPts int) int {
	switch {
	case cardPts >= 120:
		return 4
	case cardPts >= 91:
		return 2
	default: // 61-90
		return 1
	}
}

// suecaTeamLabel チーム表示ラベル (-1=Draw)。
func suecaTeamLabel(team int) string {
	if team < 0 {
		return "Draw"
	}
	return teamName(team)
}

// --- Trick / play helpers ---

// validatePlay マストフォロー (リードスートに従う) を検証する。
func (g *Sueca) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && g.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Sueca) playerHasSuit(playerIdx, design int) bool {
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
func (g *Sueca) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.suecaRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		// 勝負に絡むのは切り札またはリードスートの札のみ。
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.suecaRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// suecaRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Sueca) suecaRank(card *Card) int {
	r := suecaStrength(card.GetValue())
	if card.GetDesign() == g.trumpSuit {
		r += 100
	}
	return r
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Sueca) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Card helpers ---

// suecaStrength トリックの強さ。A>7>K>J>Q>6>5>4>3>2。
func suecaStrength(value int) int {
	switch value {
	case 1: // As
		return 10
	case 7:
		return 9
	case 13: // Rei (K)
		return 8
	case 11: // Valete (J/Sota)
		return 7
	case 12: // Dama (Q/Caballo)
		return 6
	case 6:
		return 5
	case 5:
		return 4
	case 4:
		return 3
	case 3:
		return 2
	default: // 2
		return 1
	}
}

// suecaCardPoints カードポイント。A=11,7=10,K=4,J=3,Q=2,その他=0。
func suecaCardPoints(value int) int {
	switch value {
	case 1:
		return 11
	case 7:
		return 10
	case 13:
		return 4
	case 11:
		return 3
	case 12:
		return 2
	default:
		return 0
	}
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Sueca) sortAllHands() {
	for _, p := range g.players {
		suecaSortHand(p)
	}
}

// suecaSortHand 手札をスート→強さ順にソートする。
func suecaSortHand(p *SuecaPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return suecaStrength(cards[i].GetValue()) > suecaStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Sueca) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Sueca) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Sueca) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.suecaRank(g.currentTrick[idx].Card)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Sueca) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// appendLog 棋譜にエントリを追加する。
func (g *Sueca) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
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
func (g *Sueca) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == SuecaCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *Sueca) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return g.minBy(player, valid, func(c *Card) int {
			return suecaCardPoints(c.GetValue())*100 + suecaStrength(c.GetValue())
		})
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	partnerWinning := SuecaTeamOf(winnerIdx) == SuecaTeamOf(playerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += suecaCardPoints(tc.Card.GetValue())
	}
	var follows []int
	for _, idx := range valid {
		if player.GetCard(idx).GetDesign() == leadSuit {
			follows = append(follows, idx)
		}
	}
	if len(follows) == 0 {
		if partnerWinning {
			return g.maxBy(player, valid, func(c *Card) int {
				if c.GetDesign() == g.trumpSuit {
					return -suecaStrength(c.GetValue())
				}
				return suecaCardPoints(c.GetValue())*100 - suecaStrength(c.GetValue())
			})
		}
		return g.minBy(player, valid, func(c *Card) int {
			return suecaCardPoints(c.GetValue())*100 + suecaStrength(c.GetValue())
		})
	}
	winners := suecaFilter(follows, func(idx int) bool { return g.suecaRank(player.GetCard(idx)) > topRank })
	if partnerWinning {
		nonWinners := suecaFilter(follows, func(idx int) bool { return g.suecaRank(player.GetCard(idx)) < topRank })
		if len(nonWinners) > 0 {
			return g.maxBy(player, nonWinners, func(c *Card) int {
				return suecaCardPoints(c.GetValue())*100 - suecaStrength(c.GetValue())
			})
		}
		return g.minBy(player, follows, func(c *Card) int { return suecaStrength(c.GetValue()) })
	}
	if trickPts > 0 && len(winners) > 0 {
		return g.minBy(player, winners, func(c *Card) int { return suecaStrength(c.GetValue()) })
	}
	return g.minBy(player, follows, func(c *Card) int {
		return suecaCardPoints(c.GetValue())*100 + suecaStrength(c.GetValue())
	})
}

// minBy score が最小となるインデックスを返す。
func (g *Sueca) minBy(player *SuecaPlayer, indices []int, score func(*Card) int) int {
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

// maxBy score が最大となるインデックスを返す。
func (g *Sueca) maxBy(player *SuecaPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// suecaFilter 述語を満たすインデックスを抽出する。
func suecaFilter(indices []int, pred func(int) bool) []int {
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
func (g *Sueca) GetHint() *SuecaHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != SuecaPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &SuecaHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Sueca) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.suecaRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Sueca) GetPhase() SuecaPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Sueca) SetPhase(phase SuecaPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Sueca) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Sueca) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Sueca) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Sueca) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Sueca) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Sueca) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Sueca) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Sueca) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Sueca) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Sueca) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Sueca) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit 切り札スート取得
func (g *Sueca) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Sueca) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTeamGamePoints チーム別累積ゲームポイント取得
func (g *Sueca) GetTeamGamePoints() [SuecaTeamCnt]int { return g.teamGamePts }

// SetTeamGamePoints チーム別累積ゲームポイント設定 (テスト用)
func (g *Sueca) SetTeamGamePoints(s [SuecaTeamCnt]int) { g.teamGamePts = s }

// GetRoundCardPoints 現ラウンドのチーム別カード得点取得
func (g *Sueca) GetRoundCardPoints() [SuecaTeamCnt]int { return g.roundCardPts }

// SetRoundCardPoints 現ラウンドのカード得点設定 (テスト用)
func (g *Sueca) SetRoundCardPoints(s [SuecaTeamCnt]int) { g.roundCardPts = s }

// GetRoundWinnerTeam 直近ラウンドの勝者チーム取得 (-1=引き分け/未確定)
func (g *Sueca) GetRoundWinnerTeam() int { return g.roundWinnerTeam }

// GetRoundGamePoints 直近ラウンドで勝者が得たゲームポイント取得
func (g *Sueca) GetRoundGamePoints() int { return g.roundGamePts }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Sueca) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *Sueca) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Sueca) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Sueca) GetPlayer(i int) *SuecaPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か。
func (g *Sueca) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Sueca) GetConfig() SuecaConfig { return g.config }

// SetConfig 設定変更
func (g *Sueca) SetConfig(cfg SuecaConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Sueca) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Sueca) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != SuecaPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// suecaJSON is the JSON wire format for Sueca.
type suecaJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*SuecaPlayer    `json:"ps"`
	Config           SuecaConfig       `json:"cf"`
	Phase            SuecaPhase        `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LeadPlayerIdx    int               `json:"li"`
	DealerIdx        int               `json:"di"`
	TrumpSuit        int               `json:"ts"`
	TeamGamePts      [SuecaTeamCnt]int `json:"tg"`
	RoundCardPts     [SuecaTeamCnt]int `json:"rp"`
	RoundWinnerTeam  int               `json:"rw"`
	RoundGamePts     int               `json:"rg"`
	GameEndFlag      bool              `json:"ge"`
	WinnerTeam       int               `json:"wt"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Sueca) MarshalJSON() ([]byte, error) {
	return json.Marshal(suecaJSON{
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
		TrumpSuit:        g.trumpSuit,
		TeamGamePts:      g.teamGamePts,
		RoundCardPts:     g.roundCardPts,
		RoundWinnerTeam:  g.roundWinnerTeam,
		RoundGamePts:     g.roundGamePts,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// suecaMaxSliceLen caps slice sizes during deserialisation.
const suecaMaxSliceLen = 5000

// errSuecaOversized is the single sentinel error for oversized input arrays.
var errSuecaOversized = errors.New("sueca: input array exceeds maximum allowed size")

// errSuecaInvalidPlayers is returned when restored state lacks exactly SuecaPlayerCnt players.
var errSuecaInvalidPlayers = errors.New("sueca: invalid player count")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Sueca) UnmarshalJSON(data []byte) error {
	var j suecaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > suecaMaxSliceLen || len(j.CurrentTrick) > suecaMaxSliceLen ||
		len(j.ActionLog) > suecaMaxSliceLen {
		return errSuecaOversized
	}
	if len(j.Players) != SuecaPlayerCnt {
		return errSuecaInvalidPlayers
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBriscola()
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
	g.trumpSuit = j.TrumpSuit
	g.teamGamePts = j.TeamGamePts
	g.roundCardPts = j.RoundCardPts
	g.roundWinnerTeam = j.RoundWinnerTeam
	g.roundGamePts = j.RoundGamePts
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
