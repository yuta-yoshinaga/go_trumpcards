//go:build !js || !wasm || classic

// Package domain マニーユ (Manille) のドメインモデル。
//
// Manille はフランス・ベルギー発祥の 4 人 2 チームのトリックテイキングゲーム。
// 32 枚デッキで席 0&2 vs 1&3 が 8 トリックを戦う。最大の特徴は 10(Manille) が
// 最強・A(Manillon) が次点という逆転ランク体系。マストフォローだが、味方が
// 既にトリックを取っている場合は切り札を出す義務がない (温存できる)。
//
// ランクの強さ (切り札・非切り札共通): 10 > A > K > Q > J > 9 > 8 > 7
// カードポイント: 10=5, A=4, K=3, Q=2, J=1, 9/8/7=0 (合計 60 点)
// 切り札スートは配り切った最後の札のスートで決まる (簡略化実装)。累積点が目標
// (既定 101) に達したチームが勝利する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// ManillePlayerCnt プレイヤー数 (人間 1 + CPU 3)
const ManillePlayerCnt = 4

// ManilleTeamCnt チーム数
const ManilleTeamCnt = 2

// ManilleHandSize 各プレイヤーの手札枚数 (32 / 4)
const ManilleHandSize = 8

// ManilleTrickCount 1 ラウンドのトリック数
const ManilleTrickCount = 8

// ManillePhase ゲームフェーズ
type ManillePhase int

// Manille のフェーズ定数
const (
	// ManillePhasePlay トリックプレイフェーズ
	ManillePhasePlay ManillePhase = 0
	// ManillePhaseTrickEnd トリック終了フェーズ
	ManillePhaseTrickEnd ManillePhase = 1
	// ManillePhaseRoundEnd ラウンド終了フェーズ
	ManillePhaseRoundEnd ManillePhase = 2
	// ManillePhaseGameEnd ゲーム終了フェーズ
	ManillePhaseGameEnd ManillePhase = 3
)

// ManilleHint ヒント情報
type ManilleHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Manille マニーユのゲームクラス
type Manille struct {
	trumpCards       *TrumpCards
	players          []*ManillePlayer
	config           ManilleConfig
	phase            ManillePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int                 // 切り札スート
	teamScores       [ManilleTeamCnt]int // 累積点
	roundCardPts     [ManilleTeamCnt]int // 現ラウンドのカード得点
	gameEndFlag      bool
	winnerTeam       int // -1=未確定
	actionLogBase
}

// NewManille コンストラクタ
func NewManille(trumpCards *TrumpCards, players []*ManillePlayer, config ManilleConfig) *Manille {
	return &Manille{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1}
}

// NewDefaultManille 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultManille() *Manille {
	players := make([]*ManillePlayer, ManillePlayerCnt)
	players[0] = NewManillePlayer(true)
	for i := 1; i < ManillePlayerCnt; i++ {
		players[i] = NewManillePlayer(false)
	}
	return NewManille(NewTrumpCardsBelote(), players, DefaultManilleConfig())
}

// ManilleTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func ManilleTeamOf(playerIdx int) int { return playerIdx % ManilleTeamCnt }

// manilleTeamName チーム番号を表示名 (A/B) に変換する (classic ワーカーで自己完結)。
func manilleTeamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// Reset ゲーム初期化
func (g *Manille) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [ManilleTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Manille) NextRound() {
	if g.phase != ManillePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % ManillePlayerCnt
	g.startRound()
}

// startRound 手札を配り、切り札を決めてプレイフェーズを開始する。
func (g *Manille) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundCardPts = [ManilleTeamCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.leadPlayerIdx = (g.dealerIdx + 1) % ManillePlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = ManillePhasePlay
}

// deal 各プレイヤーへ 8 枚を配り、最後に配った札のスートを切り札とする。
func (g *Manille) deal() {
	var last *Card
	for i := 0; i < ManilleHandSize; i++ {
		for j := 0; j < ManillePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % ManillePlayerCnt
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
func (g *Manille) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ManillePhasePlay {
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
func (g *Manille) CpuPlay() {
	if g.gameEndFlag || g.phase != ManillePhasePlay {
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
func (g *Manille) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == ManillePlayerCnt {
		g.phase = ManillePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % ManillePlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Manille) ResolveTrick() {
	if g.phase != ManillePhaseTrickEnd || len(g.currentTrick) != ManillePlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += manilleCardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	team := ManilleTeamOf(winnerIdx)
	g.roundCardPts[team] += pts
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d)", playerName(g.players, winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= ManilleTrickCount {
		g.phase = ManillePhaseRoundEnd
	} else {
		g.phase = ManillePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Manille) NextTrick() {
	if g.phase != ManillePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = ManillePhasePlay
}

// ScoreRound ラウンドのカード得点を累積し、マッチ終了を判定する。
func (g *Manille) ScoreRound() {
	if g.phase != ManillePhaseRoundEnd {
		return
	}
	for t := 0; t < ManilleTeamCnt; t++ {
		g.teamScores[t] += g.roundCardPts[t]
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: A=%d (round %d), B=%d (round %d)",
			g.roundNumber, g.teamScores[0], g.roundCardPts[0],
			g.teamScores[1], g.roundCardPts[1]), nil)

	leader, other := 0, 1
	if g.teamScores[1] > g.teamScores[0] {
		leader, other = 1, 0
	}
	if g.teamScores[leader] >= g.config.TargetPoints && g.teamScores[leader] > g.teamScores[other] {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = ManillePhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the match!", manilleTeamName(leader)), nil)
	}
	// 加算済みのラウンド点をクリアして二重計上を防ぐ (冪等性)。
	g.roundCardPts = [ManilleTeamCnt]int{}
}

// --- Trick / play helpers ---

// validatePlay マストフォロー + 味方リード時の切り札温存ルールを検証する。
func (g *Manille) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	hasLeadSuit := g.playerHasSuit(playerIdx, leadSuit)
	// リードスートを持っていれば必ず従う。
	if hasLeadSuit && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	if hasLeadSuit {
		return nil
	}
	// リードスートのボイド。味方が現在トリックを取っているなら切り札温存可 (何でも捨てられる)。
	if g.partnerWinning(playerIdx) {
		return nil
	}
	// 味方が勝っていない: 切り札を持っていれば切り札を出す義務がある。
	if g.playerHasSuit(playerIdx, g.trumpSuit) && card.GetDesign() != g.trumpSuit {
		return NewDomainError(ErrInvalidPlay, "切り札を出してください")
	}
	return nil
}

// partnerWinning 現在のトリックを味方プレイヤーが取っているか。
func (g *Manille) partnerWinning(playerIdx int) bool {
	if len(g.currentTrick) == 0 {
		return false
	}
	winnerIdx := g.trickWinner()
	return ManilleTeamOf(winnerIdx) == ManilleTeamOf(playerIdx)
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Manille) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(g.players[playerIdx], design)
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札 (Manille ランク) が勝つ。
func (g *Manille) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.manilleRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.manilleRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// manilleRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Manille) manilleRank(card *Card) int {
	if card.GetDesign() == g.trumpSuit {
		return 100 + manilleStrength(card.GetValue())
	}
	return manilleStrength(card.GetValue())
}

// manilleStrength カード強度。10 > A > K > Q > J > 9 > 8 > 7。
func manilleStrength(value int) int {
	switch value {
	case 10: // Manille
		return 8
	case 1: // Ace (Manillon)
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

// manilleCardPoints カードポイント。10=5, A=4, K=3, Q=2, J=1, 他=0。
func manilleCardPoints(card *Card) int {
	switch card.GetValue() {
	case 10:
		return 5
	case 1:
		return 4
	case 13:
		return 3
	case 12:
		return 2
	case 11:
		return 1
	default:
		return 0
	}
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Manille) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Manille) sortAllHands() {
	for _, p := range g.players {
		manilleSortHand(p)
	}
}

// manilleSortHand 手札をスート→強さ順にソートする。
func manilleSortHand(p *ManillePlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return manilleStrength(cards[i].GetValue()) > manilleStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Manille) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Manille) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.manilleRank(g.currentTrick[idx].Card)
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Manille) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == ManilleCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *Manille) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int {
			return manilleCardPoints(c)*100 + g.manilleRank(c)
		})
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	partnerWinning := ManilleTeamOf(winnerIdx) == ManilleTeamOf(playerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += manilleCardPoints(tc.Card)
	}
	winners := manilleFilter(valid, func(idx int) bool { return g.manilleRank(player.GetCard(idx)) > topRank })
	if partnerWinning {
		// 味方が勝っている: 得点の高い札を渡す。
		nonWinners := manilleFilter(valid, func(idx int) bool { return g.manilleRank(player.GetCard(idx)) < topRank })
		if len(nonWinners) > 0 {
			return pickHighest(player, nonWinners, func(c *Card) int { return manilleCardPoints(c) })
		}
		return pickLowest(player, valid, func(c *Card) int { return g.manilleRank(c) })
	}
	if trickPts > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.manilleRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int {
		return manilleCardPoints(c)*100 + g.manilleRank(c)
	})
}

// manilleFilter 述語を満たすインデックスを抽出する。
func manilleFilter(indices []int, pred func(int) bool) []int {
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
func (g *Manille) GetHint() *ManilleHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != ManillePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &ManilleHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Manille) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.manilleRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Manille) GetPhase() ManillePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Manille) SetPhase(phase ManillePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Manille) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Manille) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Manille) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Manille) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Manille) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Manille) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Manille) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Manille) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Manille) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Manille) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Manille) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit 切り札スート取得
func (g *Manille) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Manille) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTeamScores チーム別累積点取得
func (g *Manille) GetTeamScores() [ManilleTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点設定 (テスト用)
func (g *Manille) SetTeamScores(s [ManilleTeamCnt]int) { g.teamScores = s }

// GetRoundCardPoints 現ラウンドのカード得点取得
func (g *Manille) GetRoundCardPoints() [ManilleTeamCnt]int { return g.roundCardPts }

// SetRoundCardPoints 現ラウンドのカード得点設定 (テスト用)
func (g *Manille) SetRoundCardPoints(s [ManilleTeamCnt]int) { g.roundCardPts = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Manille) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *Manille) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Manille) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Manille) GetPlayer(i int) *ManillePlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間か。
func (g *Manille) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Manille) GetConfig() ManilleConfig { return g.config }

// SetConfig 設定変更
func (g *Manille) SetConfig(cfg ManilleConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Manille) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != ManillePhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// manilleJSON is the JSON wire format for Manille.
type manilleJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*ManillePlayer    `json:"ps"`
	Config           ManilleConfig       `json:"cf"`
	Phase            ManillePhase        `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	LeadPlayerIdx    int                 `json:"li"`
	DealerIdx        int                 `json:"di"`
	TrumpSuit        int                 `json:"ts"`
	TeamScores       [ManilleTeamCnt]int `json:"sc"`
	RoundCardPts     [ManilleTeamCnt]int `json:"rp"`
	GameEndFlag      bool                `json:"ge"`
	WinnerTeam       int                 `json:"wt"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Manille) MarshalJSON() ([]byte, error) {
	return json.Marshal(manilleJSON{
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
		TeamScores:       g.teamScores,
		RoundCardPts:     g.roundCardPts,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// manilleMaxSliceLen caps slice sizes during deserialisation.
const manilleMaxSliceLen = 5000

// errManilleOversized is the single sentinel error for oversized input arrays.
var errManilleOversized = errors.New("manille: input array exceeds maximum allowed size")

// errManilleInvalidPlayers is returned when restored state lacks exactly ManillePlayerCnt players.
var errManilleInvalidPlayers = errors.New("manille: invalid player count")

// errManilleInvalidTrick is returned when a restored trick card or its card is nil.
var errManilleInvalidTrick = errors.New("manille: invalid trick card")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Manille) UnmarshalJSON(data []byte) error {
	var j manilleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > manilleMaxSliceLen || len(j.CurrentTrick) > manilleMaxSliceLen ||
		len(j.ActionLog) > manilleMaxSliceLen {
		return errManilleOversized
	}
	if len(j.Players) != ManillePlayerCnt {
		return errManilleInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errManilleInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errManilleInvalidTrick
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
	g.trumpSuit = j.TrumpSuit
	g.teamScores = j.TeamScores
	g.roundCardPts = j.RoundCardPts
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
