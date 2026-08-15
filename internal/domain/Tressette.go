//go:build !js || !wasm || casino

// Package domain トレセッテ (Tressette) のドメインモデル。
//
// Tressette はイタリアの3大国民的カードゲームの一つで、切り札を持たない
// 純粋なマストフォローのトリックテイキングゲーム。40枚デッキ (8,9,10 を除く)
// を 4 人 (2 対 2 のチーム戦) で全て配り切り、特定の得点札を奪い合う。
//
// カードの強さ:   3 > 2 > A > K > Q > J > 7 > 6 > 5 > 4
// カードの得点(1/3点単位): A=3, 2/3/J/Q/K=1, それ以外=0。最終トリックの勝者に
// ボーナス 1 (=1/3点)。1ラウンドの合計は 33 (=11点)。各ラウンド終了時に
// チームの「3分の1点」を 3 で割って (端数切り捨て) 累積点に加算し、目標点
// (既定 21 点) に先に到達したチームが勝者となる。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// TressettePlayerCnt トレセッテのプレイヤー数
const TressettePlayerCnt = 4

// TressetteHandSize 各プレイヤーの手札枚数 (40 / 4)
const TressetteHandSize = 10

// TressetteTrickCount 1ラウンドのトリック数
const TressetteTrickCount = 10

// TressetteTeamCnt チーム数
const TressetteTeamCnt = 2

// TressetteUltimaThirds 最終トリック勝者へのボーナス (1/3点 = 1)
const TressetteUltimaThirds = 1

// TressetteRoundThirds 1ラウンドで奪い合う得点の総和 (1/3点単位 = 11点)
const TressetteRoundThirds = 33

// TressettePhase ゲームフェーズ
type TressettePhase int

// Tressette のフェーズ定数
const (
	// TressettePhasePlay トリックプレイフェーズ
	TressettePhasePlay TressettePhase = 0
	// TressettePhaseTrickEnd トリック終了フェーズ (解決済み・次トリック待ち)
	TressettePhaseTrickEnd TressettePhase = 1
	// TressettePhaseRoundEnd ラウンド終了フェーズ
	TressettePhaseRoundEnd TressettePhase = 2
	// TressettePhaseGameEnd ゲーム終了フェーズ
	TressettePhaseGameEnd TressettePhase = 3
)

// TressetteHint ヒント情報
type TressetteHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Tressette トレセッテのゲームクラス
type Tressette struct {
	trumpCards       *TrumpCards
	players          []*TressettePlayer
	config           TressetteConfig
	phase            TressettePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	teamScores       [TressetteTeamCnt]int // 累積点 (整数点)
	teamRoundThirds  [TressetteTeamCnt]int // 現ラウンドで獲得した 1/3点 の合計
	gameEndFlag      bool
	winnerTeam       int
	actionLogBase
}

// NewTressette コンストラクタ
func NewTressette(trumpCards *TrumpCards, players []*TressettePlayer, config TressetteConfig) *Tressette {
	return &Tressette{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 0,
	}
}

// NewDefaultTressette returns Tressette with the standard 4-player setup
// (1 human, 3 CPU) and DefaultTressetteConfig. Single source of truth for CUI,
// Web, and Worker construction.
func NewDefaultTressette() *Tressette {
	players := []*TressettePlayer{
		NewTressettePlayer(true),
		NewTressettePlayer(false),
		NewTressettePlayer(false),
		NewTressettePlayer(false),
	}
	return NewTressette(NewTrumpCardsBriscola(), players, DefaultTressetteConfig())
}

// TressetteTeamOf プレイヤーインデックスが属するチーム (0 = 0&2, 1 = 1&3)
func TressetteTeamOf(playerIdx int) int { return playerIdx % TressetteTeamCnt }

// Reset ゲーム初期化: デッキをシャッフルして配り、最初のラウンドを開始する。
func (g *Tressette) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.teamScores = [TressetteTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Tressette) NextRound() {
	if g.phase != TressettePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.startRound()
}

// startRound 手札を配り、リードプレイヤーを決めてプレイフェーズを開始する。
func (g *Tressette) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.teamRoundThirds = [TressetteTeamCnt]int{}

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	dealAllCards(g.trumpCards, g.players)
	g.sortAllHands()

	g.leadPlayerIdx = (g.roundNumber - 1) % TressettePlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = TressettePhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Tressette) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TressettePhasePlay {
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

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (g *Tressette) CpuPlay() {
	if g.gameEndFlag || g.phase != TressettePhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	player := g.players[g.currentPlayerIdx]
	cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(g.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Tressette) ResolveTrick() {
	if g.phase != TressettePhaseTrickEnd || len(g.currentTrick) != TressettePlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	thirds := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		thirds += tressetteThirds(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)

	team := TressetteTeamOf(winnerIdx)
	g.teamRoundThirds[team] += thirds
	bonus := ""
	if g.trickNumber >= TressetteTrickCount {
		g.teamRoundThirds[team] += TressetteUltimaThirds
		bonus = " +ultima"
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d/3%s)", playerName(g.players, winnerIdx), g.trickNumber, thirds, bonus),
		trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= TressetteTrickCount {
		g.phase = TressettePhaseRoundEnd
	} else {
		g.phase = TressettePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *Tressette) NextTrick() {
	if g.phase != TressettePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = TressettePhasePlay
}

// ScoreRound ラウンドの得点を確定し、ゲーム終了判定を行う。1/3点を 3 で割って
// (端数切り捨て) 累積点へ加算する。
func (g *Tressette) ScoreRound() {
	if g.phase != TressettePhaseRoundEnd {
		return
	}

	for t := 0; t < TressetteTeamCnt; t++ {
		g.teamScores[t] += g.teamRoundThirds[t] / 3
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: TeamA=%d (+%d/3), TeamB=%d (+%d/3)",
			g.roundNumber, g.teamScores[0], g.teamRoundThirds[0], g.teamScores[1], g.teamRoundThirds[1]), nil)

	leader, other := 0, 1
	if g.teamScores[1] > g.teamScores[0] {
		leader, other = 1, 0
	}
	if g.teamScores[leader] >= g.config.TargetPoints && g.teamScores[leader] > g.teamScores[other] {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = TressettePhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the game!", teamName(leader)), nil)
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Tressette) GetPhase() TressettePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Tressette) SetPhase(phase TressettePhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Tressette) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Tressette) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Tressette) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Tressette) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Tressette) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Tressette) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Tressette) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Tressette) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Tressette) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Tressette) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetTeamScores チーム別累積点を取得
func (g *Tressette) GetTeamScores() [TressetteTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点を設定 (テスト用)
func (g *Tressette) SetTeamScores(s [TressetteTeamCnt]int) { g.teamScores = s }

// GetTeamRoundThirds チーム別の現ラウンド 1/3点 を取得
func (g *Tressette) GetTeamRoundThirds() [TressetteTeamCnt]int { return g.teamRoundThirds }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Tressette) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Tressette) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Tressette) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Tressette) GetPlayer(i int) *TressettePlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Tressette) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Tressette) GetConfig() TressetteConfig { return g.config }

// SetConfig 設定変更
func (g *Tressette) SetConfig(cfg TressetteConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能な (マストフォローを満たす) カードのインデックス一覧を返す。
func (g *Tressette) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- Private methods ---

// playCard カードをプレイする共通処理
func (g *Tressette) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == TressettePlayerCnt {
		g.phase = TressettePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TressettePlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する (マストフォロー)
func (g *Tressette) validatePlay(playerIdx int, card *Card) error {
	return validateFollowSuit(g.currentTrick, g.players, playerIdx, card)
}

// trickWinner トリックの勝者を決定する。切り札がないため、リードスートの最強札が勝つ。
func (g *Tressette) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStrength := tressetteStrength(g.currentTrick[0].Card.GetValue())

	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() == leadSuit && tressetteStrength(tc.Card.GetValue()) > winnerStrength {
			winnerIdx = tc.PlayerIdx
			winnerStrength = tressetteStrength(tc.Card.GetValue())
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Tressette) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *Tressette) sortAllHands() {
	for _, p := range g.players {
		tressetteSortHand(p)
	}
}

// tressetteSortHand プレイヤーの手札をスート→強さの順にソートする
func tressetteSortHand(p *TressettePlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return tressetteStrength(ci.GetValue()) < tressetteStrength(cj.GetValue())
	})
}

// teamName チーム表示名 (0=A, 1=B)
func teamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// --- Card helpers ---

// tressetteStrength トリックの強さ。3 が最強 (9)、4 が最弱 (0)。
//
//	3 > 2 > A > K > Q > J > 7 > 6 > 5 > 4
func tressetteStrength(value int) int {
	switch value {
	case 3:
		return 9
	case 2:
		return 8
	case 1: // Ace
		return 7
	case 13: // King
		return 6
	case 12: // Queen
		return 5
	case 11: // Jack
		return 4
	case 7:
		return 3
	case 6:
		return 2
	case 5:
		return 1
	default: // 4
		return 0
	}
}

// tressetteThirds カードの得点を 1/3点 単位で返す。A=3、2/3/J/Q/K=1、その他=0。
func tressetteThirds(value int) int {
	switch value {
	case 1: // Ace = 1 point = 3 thirds
		return 3
	case 2, 3, 11, 12, 13: // 2,3,J,Q,K = 1/3 point each
		return 1
	default:
		return 0
	}
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨プレイを返す。
func (g *Tressette) GetHint() *TressetteHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != TressettePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &TressetteHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason ヒント理由を判定する
func (g *Tressette) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStrength := tressetteStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	if tressetteStrength(card.GetValue()) > topStrength {
		return "follow_win"
	}
	if TressetteTeamOf(winnerIdx) == TressetteTeamOf(playerIdx) {
		return "give_partner"
	}
	return "follow_duck"
}

// indexOfPlayerInTrick currentTrick 内で playerIdx が出したカードの位置を返す (-1=なし)
func (g *Tressette) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (g *Tressette) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == TressetteCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点を意識した戦略プレイ
func (g *Tressette) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	// リード: 得点・強さの低いカードを出して温存する。
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int {
			return tressetteThirds(c.GetValue())*100 + tressetteStrength(c.GetValue())
		})
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topStrength := tressetteStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	partnerWinning := TressetteTeamOf(winnerIdx) == TressetteTeamOf(playerIdx)
	trickThirds := 0
	for _, tc := range g.currentTrick {
		trickThirds += tressetteThirds(tc.Card.GetValue())
	}

	var follows []int
	for _, idx := range valid {
		if player.GetCard(idx).GetDesign() == leadSuit {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		// ボイド: 得点・強さの低いカードを捨てて温存する。
		return pickLowest(player, valid, func(c *Card) int {
			return tressetteThirds(c.GetValue())*100 + tressetteStrength(c.GetValue())
		})
	}

	winners := filterIndices(follows, func(idx int) bool {
		return tressetteStrength(player.GetCard(idx).GetValue()) > topStrength
	})

	if partnerWinning {
		// 味方が勝っている: 得点札を渡しつつ、無駄に上書きしない。
		nonWinners := filterIndices(follows, func(idx int) bool {
			return tressetteStrength(player.GetCard(idx).GetValue()) < topStrength
		})
		if len(nonWinners) > 0 {
			// 勝ちを取らない範囲で最も得点の高い札を渡す。
			return pickHighest(player, nonWinners, func(c *Card) int {
				return tressetteThirds(c.GetValue())*100 - tressetteStrength(c.GetValue())
			})
		}
		// 上書きせざるを得ない場合は最弱札で被害を抑える。
		return pickLowest(player, follows, func(c *Card) int { return tressetteStrength(c.GetValue()) })
	}

	// 相手が勝っている: 得点があり勝てるなら最小限の札で取りに行く。
	if trickThirds > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return tressetteStrength(c.GetValue()) })
	}
	// 取れない/取る価値がない: 得点・強さの低い札でダックする。
	return pickLowest(player, follows, func(c *Card) int {
		return tressetteThirds(c.GetValue())*100 + tressetteStrength(c.GetValue())
	})
}

// --- JSON ---

// tressetteJSON is the JSON wire format for Tressette.
type tressetteJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*TressettePlayer    `json:"ps"`
	Config           TressetteConfig       `json:"cf"`
	Phase            TressettePhase        `json:"ph"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	CurrentPlayerIdx int                   `json:"ci"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	LeadPlayerIdx    int                   `json:"li"`
	TeamScores       [TressetteTeamCnt]int `json:"ts"`
	TeamRoundThirds  [TressetteTeamCnt]int `json:"tr"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerTeam       int                   `json:"wt"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Tressette) MarshalJSON() ([]byte, error) {
	return json.Marshal(tressetteJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		TeamScores:       g.teamScores,
		TeamRoundThirds:  g.teamRoundThirds,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// tressetteMaxSliceLen caps slice sizes during deserialisation.
const tressetteMaxSliceLen = 1000

// errTressetteOversized is the single sentinel error for oversized input arrays.
var errTressetteOversized = errors.New("tressette: input array exceeds maximum allowed size")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Tressette) UnmarshalJSON(data []byte) error {
	var j tressetteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tressetteMaxSliceLen || len(j.CurrentTrick) > tressetteMaxSliceLen ||
		len(j.ActionLog) > tressetteMaxSliceLen {
		return errTressetteOversized
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBriscola()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*TressettePlayer, 0)
	}
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
	g.teamScores = j.TeamScores
	g.teamRoundThirds = j.TeamRoundThirds
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
