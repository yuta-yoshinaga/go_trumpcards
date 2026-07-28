//go:build !js || !wasm || classic

// Package domain セドマ (Sedma) のドメインモデル。
//
// Sedma はチェコ・スロバキア発祥の 4 人 2 チームの捕獲型トリックゲーム。切り札は
// なく、リードカードと同ランクのカードか 7 (ワイルド) を出したプレイヤーがトリックを
// 奪取する。各トリックは 1 人 1 枚ずつ計 4 枚、最後に奪取したプレイヤーが勝者となる。
// A と 10 がポイントカード (各 10 点)、最終 (8th) トリック勝者に +10 点。累積点が目標
// (既定 101) に達したチームが勝利する。フォロー義務はなく任意のカードを出せる。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// SedmaPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const SedmaPlayerCnt = 4

// SedmaTeamCnt チーム数
const SedmaTeamCnt = 2

// SedmaHandSize 各プレイヤーの手札枚数 (32 / 4)
const SedmaHandSize = 8

// SedmaTrickCount 1 ラウンドのトリック数
const SedmaTrickCount = 8

// SedmaLastTrickBonus 最終トリック勝者へのボーナス点
const SedmaLastTrickBonus = 10

// SedmaWildValue ワイルドカードのランク (7)
const SedmaWildValue = 7

// SedmaPhase ゲームフェーズ
type SedmaPhase int

// Sedma のフェーズ定数
const (
	// SedmaPhasePlay トリックプレイフェーズ
	SedmaPhasePlay SedmaPhase = 0
	// SedmaPhaseTrickEnd トリック終了フェーズ
	SedmaPhaseTrickEnd SedmaPhase = 1
	// SedmaPhaseRoundEnd ラウンド終了フェーズ
	SedmaPhaseRoundEnd SedmaPhase = 2
	// SedmaPhaseGameEnd ゲーム終了フェーズ
	SedmaPhaseGameEnd SedmaPhase = 3
)

// SedmaHint ヒント情報
type SedmaHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Sedma セドマのゲームクラス
type Sedma struct {
	trumpCards       *TrumpCards
	players          []*SedmaPlayer
	config           SedmaConfig
	phase            SedmaPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	teamScores       [SedmaTeamCnt]int // 累積点
	roundCardPts     [SedmaTeamCnt]int // 現ラウンドのカード得点
	gameEndFlag      bool
	winnerTeam       int // -1=未確定
	actionLog        []*ActionLogEntry
}

// NewSedma コンストラクタ
func NewSedma(trumpCards *TrumpCards, players []*SedmaPlayer, config SedmaConfig) *Sedma {
	return &Sedma{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1}
}

// NewDefaultSedma 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultSedma() *Sedma {
	players := make([]*SedmaPlayer, SedmaPlayerCnt)
	players[0] = NewSedmaPlayer(true)
	for i := 1; i < SedmaPlayerCnt; i++ {
		players[i] = NewSedmaPlayer(false)
	}
	return NewSedma(NewTrumpCardsBelote(), players, DefaultSedmaConfig())
}

// SedmaTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func SedmaTeamOf(playerIdx int) int { return playerIdx % SedmaTeamCnt }

// SedmaTeamName チーム番号を表示名 (A/B) に変換する (classic ワーカーで自己完結)。
func SedmaTeamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// Reset ゲーム初期化
func (g *Sedma) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [SedmaTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Sedma) NextRound() {
	if g.phase != SedmaPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % SedmaPlayerCnt
	g.startRound()
}

// startRound 手札を配り、プレイフェーズを開始する。
func (g *Sedma) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundCardPts = [SedmaTeamCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.leadPlayerIdx = (g.dealerIdx + 1) % SedmaPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = SedmaPhasePlay
}

// deal 各プレイヤーへ 8 枚を配る。
func (g *Sedma) deal() {
	for i := 0; i < SedmaHandSize; i++ {
		for j := 0; j < SedmaPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % SedmaPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Sedma) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SedmaPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *Sedma) CpuPlay() {
	if g.gameEndFlag || g.phase != SedmaPhasePlay {
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
func (g *Sedma) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == SedmaPlayerCnt {
		g.phase = SedmaPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SedmaPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Sedma) ResolveTrick() {
	if g.phase != SedmaPhaseTrickEnd || len(g.currentTrick) != SedmaPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += sedmaCardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	team := SedmaTeamOf(winnerIdx)
	g.roundCardPts[team] += pts
	bonus := ""
	if g.trickNumber >= SedmaTrickCount {
		g.roundCardPts[team] += SedmaLastTrickBonus
		bonus = fmt.Sprintf(" +%d last", SedmaLastTrickBonus)
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s captures trick %d (+%d%s)", g.playerName(winnerIdx), g.trickNumber, pts, bonus), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= SedmaTrickCount {
		g.phase = SedmaPhaseRoundEnd
	} else {
		g.phase = SedmaPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Sedma) NextTrick() {
	if g.phase != SedmaPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = SedmaPhasePlay
}

// ScoreRound ラウンドのカード得点を累積し、マッチ終了を判定する。
func (g *Sedma) ScoreRound() {
	if g.phase != SedmaPhaseRoundEnd {
		return
	}
	for t := 0; t < SedmaTeamCnt; t++ {
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
		g.phase = SedmaPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the match!", SedmaTeamName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// trickWinner トリックの勝者を決定する。リードと同ランクか 7 を出した最後の
// プレイヤーが奪取する。誰も奪取しなければリードプレイヤーが勝つ。
func (g *Sedma) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadRank := g.currentTrick[0].Card.GetValue()
	winnerIdx := g.currentTrick[0].PlayerIdx
	for _, tc := range g.currentTrick[1:] {
		v := tc.Card.GetValue()
		if v == leadRank || v == SedmaWildValue {
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// sedmaCardPoints カードポイント。A=10, 10=10, 他=0。
func sedmaCardPoints(card *Card) int {
	switch card.GetValue() {
	case 1, 10: // Ace, Ten
		return 10
	default:
		return 0
	}
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す (Sedma は任意)。
func (g *Sedma) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	valid := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		valid = append(valid, i)
	}
	return valid
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Sedma) sortAllHands() {
	for _, p := range g.players {
		sedmaSortHand(p)
	}
}

// sedmaSortHand 手札をスート→ランク順にソートする。
func sedmaSortHand(p *SedmaPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return sedmaSortRank(cards[i].GetValue()) < sedmaSortRank(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// sedmaSortRank 表示用の並び順 (7,8,9,10,J,Q,K,A)。
func sedmaSortRank(value int) int {
	if value == 1 {
		return 14 // Ace high for display
	}
	return value
}

// playerName プレイヤー名を返す。
func (g *Sedma) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Sedma) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// appendLog 棋譜にエントリを追加する。
func (g *Sedma) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
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
func (g *Sedma) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == SedmaCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 捕獲とチーム関係を意識した戦略プレイ。
func (g *Sedma) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	// リード時: 7 とポイント札を温存し、価値の低い札を出す。
	if len(g.currentTrick) == 0 {
		return g.minBy(player, valid, func(c *Card) int { return sedmaLeadCost(c) })
	}
	leadRank := g.currentTrick[0].Card.GetValue()
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += sedmaCardPoints(tc.Card)
	}
	currentWinner := g.trickWinner()
	partnerWinning := SedmaTeamOf(currentWinner) == SedmaTeamOf(playerIdx) && currentWinner != playerIdx
	captures := sedmaFilter(valid, func(idx int) bool {
		v := player.GetCard(idx).GetValue()
		return v == leadRank || v == SedmaWildValue
	})
	// 味方が奪取中ならポイント札を渡して温存。
	if partnerWinning {
		return g.maxBy(player, valid, func(c *Card) int { return sedmaCardPoints(c) })
	}
	// ポイントがあり奪取できるなら、まず同ランク、無ければ 7 で奪取。
	if trickPts > 0 && len(captures) > 0 {
		nonWild := sedmaFilter(captures, func(idx int) bool { return player.GetCard(idx).GetValue() != SedmaWildValue })
		if len(nonWild) > 0 {
			return nonWild[0]
		}
		return captures[0]
	}
	// 奪取しない: 価値の低い非ポイント・非 7 札を捨てる。
	return g.minBy(player, valid, func(c *Card) int { return sedmaLeadCost(c) })
}

// sedmaLeadCost リード/ディスカード時の温存コスト (低いほど捨ててよい)。
func sedmaLeadCost(c *Card) int {
	if c.GetValue() == SedmaWildValue {
		return 100 // 7 は最後まで温存
	}
	return sedmaCardPoints(c)*10 + sedmaSortRank(c.GetValue())
}

// minBy score が最小となるインデックスを返す。
func (g *Sedma) minBy(player *SedmaPlayer, indices []int, score func(*Card) int) int {
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
func (g *Sedma) maxBy(player *SedmaPlayer, indices []int, score func(*Card) int) int {
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

// sedmaFilter 述語を満たすインデックスを抽出する。
func sedmaFilter(indices []int, pred func(int) bool) []int {
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
func (g *Sedma) GetHint() *SedmaHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != SedmaPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &SedmaHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Sedma) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	leadRank := g.currentTrick[0].Card.GetValue()
	if card.GetValue() == leadRank || card.GetValue() == SedmaWildValue {
		return "capture"
	}
	return "discard_low"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Sedma) GetPhase() SedmaPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Sedma) SetPhase(phase SedmaPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Sedma) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Sedma) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Sedma) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Sedma) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Sedma) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Sedma) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Sedma) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Sedma) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Sedma) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Sedma) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Sedma) GetDealerIdx() int { return g.dealerIdx }

// GetTeamScores チーム別累積点取得
func (g *Sedma) GetTeamScores() [SedmaTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点設定 (テスト用)
func (g *Sedma) SetTeamScores(s [SedmaTeamCnt]int) { g.teamScores = s }

// GetRoundCardPoints 現ラウンドのカード得点取得
func (g *Sedma) GetRoundCardPoints() [SedmaTeamCnt]int { return g.roundCardPts }

// SetRoundCardPoints 現ラウンドのカード得点設定 (テスト用)
func (g *Sedma) SetRoundCardPoints(s [SedmaTeamCnt]int) { g.roundCardPts = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Sedma) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *Sedma) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Sedma) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Sedma) GetPlayer(i int) *SedmaPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か。
func (g *Sedma) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Sedma) GetConfig() SedmaConfig { return g.config }

// SetConfig 設定変更
func (g *Sedma) SetConfig(cfg SedmaConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Sedma) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Sedma) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != SedmaPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// sedmaJSON is the JSON wire format for Sedma.
type sedmaJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*SedmaPlayer    `json:"ps"`
	Config           SedmaConfig       `json:"cf"`
	Phase            SedmaPhase        `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LeadPlayerIdx    int               `json:"li"`
	DealerIdx        int               `json:"di"`
	TeamScores       [SedmaTeamCnt]int `json:"sc"`
	RoundCardPts     [SedmaTeamCnt]int `json:"rp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerTeam       int               `json:"wt"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Sedma) MarshalJSON() ([]byte, error) {
	return json.Marshal(sedmaJSON{
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
		TeamScores:       g.teamScores,
		RoundCardPts:     g.roundCardPts,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// sedmaMaxSliceLen caps slice sizes during deserialisation.
const sedmaMaxSliceLen = 5000

// errSedmaOversized is the single sentinel error for oversized input arrays.
var errSedmaOversized = errors.New("sedma: input array exceeds maximum allowed size")

// errSedmaInvalidPlayers is returned when restored state lacks exactly SedmaPlayerCnt players.
var errSedmaInvalidPlayers = errors.New("sedma: invalid player count")

// errSedmaInvalidTrick is returned when a restored trick card or its card is nil.
var errSedmaInvalidTrick = errors.New("sedma: invalid trick card")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Sedma) UnmarshalJSON(data []byte) error {
	var j sedmaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > sedmaMaxSliceLen || len(j.CurrentTrick) > sedmaMaxSliceLen ||
		len(j.ActionLog) > sedmaMaxSliceLen {
		return errSedmaOversized
	}
	if len(j.Players) != SedmaPlayerCnt {
		return errSedmaInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errSedmaInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errSedmaInvalidTrick
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
	g.teamScores = j.TeamScores
	g.roundCardPts = j.RoundCardPts
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
