//go:build !js || !wasm || extra2

// Package domain トラッポラ (Trappola) のドメインモデル。
//
// Trappola はイタリアの3大国民的カードゲームの一つで、切り札を持たない
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

// TrappolaPlayerCnt トラッポラのプレイヤー数
const TrappolaPlayerCnt = 4

// TrappolaHandSize 各プレイヤーの手札枚数 (36 / 4)
//
// **36 枚を 4 人で配り切る。** クローン元のトレセッテは 40 枚 10 枚ずつ。
const TrappolaHandSize = 9

// TrappolaTrickCount 1ラウンドのトリック数。
//
// **手札枚数そのもの** (36 / 4 = 9)。クローン元のトレセッテは 40 枚 10 枚ずつ
// なので 10 だった。ここが手札より大きいと、最終トリックのボーナスが
// **永久に発火しない**。
const TrappolaTrickCount = TrappolaHandSize

// TrappolaTeamCnt チーム数
const TrappolaTeamCnt = 2

// TrappolaUltimaThirds 最終トリック勝者へのボーナス (1/3点 = 1)
const TrappolaUltimaThirds = 1

// TrappolaRoundThirds 1ラウンドで奪い合う得点の総和 (1/3点単位)。
//
// A が 4 枚 × 3 + 絵札 12 枚 × 1 = 24、これに最終トリックのボーナス 1 を
// 足して 25。**デッキから数えた値**で、手で置いた数ではない
// (TestTrappolaRoundThirdsMatchesTheDeck が突き合わせる)。
const TrappolaRoundThirds = 25

// TrappolaPhase ゲームフェーズ
type TrappolaPhase int

// Trappola のフェーズ定数
const (
	// TrappolaPhasePlay トリックプレイフェーズ
	TrappolaPhasePlay TrappolaPhase = 0
	// TrappolaPhaseTrickEnd トリック終了フェーズ (解決済み・次トリック待ち)
	TrappolaPhaseTrickEnd TrappolaPhase = 1
	// TrappolaPhaseRoundEnd ラウンド終了フェーズ
	TrappolaPhaseRoundEnd TrappolaPhase = 2
	// TrappolaPhaseGameEnd ゲーム終了フェーズ
	TrappolaPhaseGameEnd TrappolaPhase = 3
)

// TrappolaHint ヒント情報
type TrappolaHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Trappola トラッポラのゲームクラス
type Trappola struct {
	trumpCards       *TrumpCards
	players          []*TrappolaPlayer
	config           TrappolaConfig
	phase            TrappolaPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	teamScores       [TrappolaTeamCnt]int // 累積点 (整数点)
	teamRoundThirds  [TrappolaTeamCnt]int // 現ラウンドで獲得した 1/3点 の合計
	// declarations は配った時点で成立した役。**盤面の一部**なので保存する。
	declarations []TrappolaDeclaration
	gameEndFlag  bool
	winnerTeam   int
	actionLogBase
}

// NewTrappola コンストラクタ
func NewTrappola(trumpCards *TrumpCards, players []*TrappolaPlayer, config TrappolaConfig) *Trappola {
	return &Trappola{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 0,
	}
}

// NewDefaultTrappola returns Trappola with the standard 4-player setup
// (1 human, 3 CPU) and DefaultTrappolaConfig. Single source of truth for CUI,
// Web, and Worker construction.
func NewDefaultTrappola() *Trappola {
	players := []*TrappolaPlayer{
		NewTrappolaPlayer(true),
		NewTrappolaPlayer(false),
		NewTrappolaPlayer(false),
		NewTrappolaPlayer(false),
	}
	return NewTrappola(NewTrumpCardsTrappola(), players, DefaultTrappolaConfig())
}

// TrappolaTeamOf プレイヤーインデックスが属するチーム (0 = 0&2, 1 = 1&3)
func TrappolaTeamOf(playerIdx int) int { return playerIdx % TrappolaTeamCnt }

// Reset ゲーム初期化: デッキをシャッフルして配り、最初のラウンドを開始する。
func (g *Trappola) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.teamScores = [TrappolaTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Trappola) NextRound() {
	if g.phase != TrappolaPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.startRound()
}

// startRound 手札を配り、リードプレイヤーを決めてプレイフェーズを開始する。
func (g *Trappola) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.teamRoundThirds = [TrappolaTeamCnt]int{}

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	dealAllCards(g.trumpCards, g.players)
	g.sortAllHands()

	g.scoreDeclarations()

	g.leadPlayerIdx = (g.roundNumber - 1) % TrappolaPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = TrappolaPhasePlay
}

// TrappolaDeclaration は配った時点で成立する役。
type TrappolaDeclaration struct {
	// PlayerIdx は役を持っていた席。
	PlayerIdx int `json:"pi"`
	// Kind は役の種類。
	Kind TrappolaDeclarationKind `json:"kd"`
	// Value は四つ揃い / 三つ揃いの札位、トラッポラのスート。
	Value int `json:"vl"`
	// Thirds は加算された 1/3 点。
	Thirds int `json:"th"`
}

// TrappolaDeclarationKind 役の種類
type TrappolaDeclarationKind int

const (
	// TrappolaDeclarationFour 四つ揃い (同ランク 4 枚)
	TrappolaDeclarationFour TrappolaDeclarationKind = iota
	// TrappolaDeclarationTrappola トラッポラ (同スートの A+K+Q)
	TrappolaDeclarationTrappola
	// TrappolaDeclarationThree 三つ揃い (同ランク 3 枚)
	TrappolaDeclarationThree
)

// 役の点 (1/3 点単位)。**カード点と同じ単位**にしてある ——
// 1 ラウンドで奪い合うカード点が 25 thirds しかないので、
// 別単位の大きな点を混ぜると役だけで勝負が決まる。
const (
	// TrappolaFourThirds 四つ揃いの点 (= 2 点)
	TrappolaFourThirds = 6
	// TrappolaTrappolaThirds トラッポラの点
	TrappolaTrappolaThirds = 4
	// TrappolaThreeThirds 三つ揃いの点 (= 1 点)
	TrappolaThreeThirds = 3
)

// scoreDeclarations は配った手札から役を判定してチーム点に足す。
//
// **プレイヤーに訊かない。** 役の申告にはコストも情報漏洩も無いので、
// 「宣言しない」選択に意味が無く、訊けば偽の選択肢になる。配り終えた時点で
// 全席ぶん自動で評価し、棋譜に残す。
func (g *Trappola) scoreDeclarations() {
	g.declarations = nil
	for idx, p := range g.players {
		for _, d := range trappolaFindDeclarations(idx, p) {
			g.declarations = append(g.declarations, d)
			g.teamRoundThirds[TrappolaTeamOf(idx)] += d.Thirds
			g.appendLog(idx, "declaration",
				fmt.Sprintf("%s declares %s (+%d thirds)",
					playerName(g.players, idx), trappolaDeclarationName(d), d.Thirds), nil)
		}
	}
}

// trappolaFindDeclarations は 1 席の手札から成立する役を返す。
//
// 同じランクで四つ揃いと三つ揃いを二重に数えない (四つ揃いが優先)。
func trappolaFindDeclarations(playerIdx int, p *TrappolaPlayer) []TrappolaDeclaration {
	byValue := map[int]int{}
	bySuit := map[int]map[int]bool{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		byValue[c.GetValue()]++
		if bySuit[c.GetDesign()] == nil {
			bySuit[c.GetDesign()] = map[int]bool{}
		}
		bySuit[c.GetDesign()][c.GetValue()] = true
	}

	out := make([]TrappolaDeclaration, 0, 4)
	// 四つ揃い / 三つ揃いは A,K,Q,J のみ。平札の揃いは役にしない。
	for _, v := range []int{1, 13, 12, 11} {
		switch {
		case byValue[v] >= 4:
			out = append(out, TrappolaDeclaration{playerIdx, TrappolaDeclarationFour, v, TrappolaFourThirds})
		case byValue[v] == 3:
			out = append(out, TrappolaDeclaration{playerIdx, TrappolaDeclarationThree, v, TrappolaThreeThirds})
		}
	}
	// トラッポラ: 同スートの A+K+Q。
	for suit := CardDesignSpade; suit <= CardDesignMax; suit++ {
		s := bySuit[suit]
		if s[1] && s[13] && s[12] {
			out = append(out, TrappolaDeclaration{playerIdx, TrappolaDeclarationTrappola, suit, TrappolaTrappolaThirds})
		}
	}
	return out
}

// trappolaDeclarationName は棋譜に出す役名を組み立てる。
func trappolaDeclarationName(d TrappolaDeclaration) string {
	switch d.Kind {
	case TrappolaDeclarationTrappola:
		return fmt.Sprintf("trappola in %s", suitStr(d.Value))
	case TrappolaDeclarationFour:
		return fmt.Sprintf("four %ds", d.Value)
	default:
		return fmt.Sprintf("three %ds", d.Value)
	}
}

// TrappolaFindDeclarationsForTest は 1 席の手札から成立する役を返す (テスト用)。
func TrappolaFindDeclarationsForTest(playerIdx int, p *TrappolaPlayer) []TrappolaDeclaration {
	return trappolaFindDeclarations(playerIdx, p)
}

// GetDeclarations は現ラウンドで成立した役を返す。
func (g *Trappola) GetDeclarations() []TrappolaDeclaration { return g.declarations }

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Trappola) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TrappolaPhasePlay {
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
func (g *Trappola) CpuPlay() {
	if g.gameEndFlag || g.phase != TrappolaPhasePlay {
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
func (g *Trappola) ResolveTrick() {
	if g.phase != TrappolaPhaseTrickEnd || len(g.currentTrick) != TrappolaPlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	thirds := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		thirds += trappolaThirds(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)

	team := TrappolaTeamOf(winnerIdx)
	g.teamRoundThirds[team] += thirds
	bonus := ""
	if g.trickNumber >= TrappolaTrickCount {
		g.teamRoundThirds[team] += TrappolaUltimaThirds
		bonus = " +ultima"
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d/3%s)", playerName(g.players, winnerIdx), g.trickNumber, thirds, bonus),
		trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= TrappolaTrickCount {
		g.phase = TrappolaPhaseRoundEnd
	} else {
		g.phase = TrappolaPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *Trappola) NextTrick() {
	if g.phase != TrappolaPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = TrappolaPhasePlay
}

// ScoreRound ラウンドの得点を確定し、ゲーム終了判定を行う。1/3点を 3 で割って
// (端数切り捨て) 累積点へ加算する。
func (g *Trappola) ScoreRound() {
	if g.phase != TrappolaPhaseRoundEnd {
		return
	}

	for t := 0; t < TrappolaTeamCnt; t++ {
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
		g.phase = TrappolaPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the game!", trappolaTeamName(leader)), nil)
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Trappola) GetPhase() TrappolaPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Trappola) SetPhase(phase TrappolaPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Trappola) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Trappola) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Trappola) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Trappola) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Trappola) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Trappola) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Trappola) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Trappola) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Trappola) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Trappola) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetTeamScores チーム別累積点を取得
func (g *Trappola) GetTeamScores() [TrappolaTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点を設定 (テスト用)
func (g *Trappola) SetTeamScores(s [TrappolaTeamCnt]int) { g.teamScores = s }

// GetTeamRoundThirds チーム別の現ラウンド 1/3点 を取得
func (g *Trappola) GetTeamRoundThirds() [TrappolaTeamCnt]int { return g.teamRoundThirds }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Trappola) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Trappola) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Trappola) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Trappola) GetPlayer(i int) *TrappolaPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Trappola) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Trappola) GetConfig() TrappolaConfig { return g.config }

// SetConfig 設定変更
func (g *Trappola) SetConfig(cfg TrappolaConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能な (マストフォローを満たす) カードのインデックス一覧を返す。
func (g *Trappola) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- Private methods ---

// playCard カードをプレイする共通処理
func (g *Trappola) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == TrappolaPlayerCnt {
		g.phase = TrappolaPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TrappolaPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する (マストフォロー)
func (g *Trappola) validatePlay(playerIdx int, card *Card) error {
	return validateFollowSuit(g.currentTrick, g.players, playerIdx, card)
}

// trickWinner トリックの勝者を決定する。切り札がないため、リードスートの最強札が勝つ。
func (g *Trappola) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStrength := trappolaStrength(g.currentTrick[0].Card.GetValue())

	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() == leadSuit && trappolaStrength(tc.Card.GetValue()) > winnerStrength {
			winnerIdx = tc.PlayerIdx
			winnerStrength = trappolaStrength(tc.Card.GetValue())
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Trappola) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *Trappola) sortAllHands() {
	for _, p := range g.players {
		trappolaSortHand(p)
	}
}

// trappolaSortHand プレイヤーの手札をスート→強さの順にソートする
func trappolaSortHand(p *TrappolaPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return trappolaStrength(ci.GetValue()) < trappolaStrength(cj.GetValue())
	})
}

// --- Card helpers ---

// trappolaTeamName チーム表示名 (0=A, 1=B)。
//
// **クローン元の teamName は共有できない。** あちら (Tressette.go) は casino
// タグ、こちらは extra2 なので、extra2 のビルドでは定義ごと消えて
// stranded symbol になる。名前を分けて自前で持つ。
func trappolaTeamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// trappolaStrength は札位の強さを返す。**順は A-K-Q-J-7-6-5-4-3。**
//
// クローン元のトレセッテは 3-2-A-K-Q-J-7-6-5-4 で、**3 と 2 が最強**という
// 独特の順を持つ。トラッポラのデッキに 2 は無く、3 は最弱。切り札は無いので
// この 1 本がトリックの勝敗を全部決める。
func trappolaStrength(value int) int {
	switch value {
	case 1: // Ace — 最強
		return 8
	case 13: // King
		return 7
	case 12: // Queen
		return 6
	case 11: // Jack
		return 5
	case 7:
		return 4
	case 6:
		return 3
	case 5:
		return 2
	case 4:
		return 1
	default: // 3 — 最弱
		return 0
	}
}

// trappolaThirds カードの得点を 1/3点 単位で返す。A=3、2/3/J/Q/K=1、その他=0。
// GetDeckForTest は山札を返す (テスト用)。
func (g *Trappola) GetDeckForTest() *TrumpCards { return g.trumpCards }

// TrappolaStrengthForTest は札位の強さを返す (テスト用)。
func TrappolaStrengthForTest(value int) int { return trappolaStrength(value) }

// TrappolaThirdsForTest はカード点 (1/3 点単位) を返す (テスト用)。
func TrappolaThirdsForTest(value int) int { return trappolaThirds(value) }

// trappolaThirds はカード点を 1/3 点単位で返す。
//
// **2 はこのデッキに無い。** クローン元のトレセッテは 2 と 3 を絵札と同じ
// 1/3 点として数えるが、トラッポラのデッキ (A,K,Q,J,7,6,5,4,3) に 2 は
// 入っておらず、3 は最弱の平札。点を持つのは A と絵札だけ。
func trappolaThirds(value int) int {
	switch value {
	case 1: // Ace = 1 point = 3 thirds
		return 3
	case 11, 12, 13: // J,Q,K = 1/3 point each
		return 1
	default:
		return 0
	}
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨プレイを返す。
func (g *Trappola) GetHint() *TrappolaHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != TrappolaPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &TrappolaHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason ヒント理由を判定する
func (g *Trappola) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStrength := trappolaStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	if trappolaStrength(card.GetValue()) > topStrength {
		return "follow_win"
	}
	if TrappolaTeamOf(winnerIdx) == TrappolaTeamOf(playerIdx) {
		return "give_partner"
	}
	return "follow_duck"
}

// indexOfPlayerInTrick currentTrick 内で playerIdx が出したカードの位置を返す (-1=なし)
func (g *Trappola) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (g *Trappola) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == TrappolaCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点を意識した戦略プレイ
func (g *Trappola) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	// リード: 得点・強さの低いカードを出して温存する。
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int {
			return trappolaThirds(c.GetValue())*100 + trappolaStrength(c.GetValue())
		})
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topStrength := trappolaStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	partnerWinning := TrappolaTeamOf(winnerIdx) == TrappolaTeamOf(playerIdx)
	trickThirds := 0
	for _, tc := range g.currentTrick {
		trickThirds += trappolaThirds(tc.Card.GetValue())
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
			return trappolaThirds(c.GetValue())*100 + trappolaStrength(c.GetValue())
		})
	}

	winners := filterIndices(follows, func(idx int) bool {
		return trappolaStrength(player.GetCard(idx).GetValue()) > topStrength
	})

	if partnerWinning {
		// 味方が勝っている: 得点札を渡しつつ、無駄に上書きしない。
		nonWinners := filterIndices(follows, func(idx int) bool {
			return trappolaStrength(player.GetCard(idx).GetValue()) < topStrength
		})
		if len(nonWinners) > 0 {
			// 勝ちを取らない範囲で最も得点の高い札を渡す。
			return pickHighest(player, nonWinners, func(c *Card) int {
				return trappolaThirds(c.GetValue())*100 - trappolaStrength(c.GetValue())
			})
		}
		// 上書きせざるを得ない場合は最弱札で被害を抑える。
		return pickLowest(player, follows, func(c *Card) int { return trappolaStrength(c.GetValue()) })
	}

	// 相手が勝っている: 得点があり勝てるなら最小限の札で取りに行く。
	if trickThirds > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return trappolaStrength(c.GetValue()) })
	}
	// 取れない/取る価値がない: 得点・強さの低い札でダックする。
	return pickLowest(player, follows, func(c *Card) int {
		return trappolaThirds(c.GetValue())*100 + trappolaStrength(c.GetValue())
	})
}

// --- JSON ---

// trappolaJSON is the JSON wire format for Trappola.
type trappolaJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*TrappolaPlayer    `json:"ps"`
	Config           TrappolaConfig       `json:"cf"`
	Phase            TrappolaPhase        `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	LeadPlayerIdx    int                  `json:"li"`
	TeamScores       [TrappolaTeamCnt]int `json:"ts"`
	TeamRoundThirds  [TrappolaTeamCnt]int `json:"tr"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerTeam       int                  `json:"wt"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Trappola) MarshalJSON() ([]byte, error) {
	return json.Marshal(trappolaJSON{
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

// trappolaMaxSliceLen caps slice sizes during deserialisation.
const trappolaMaxSliceLen = 1000

// errTrappolaOversized is the single sentinel error for oversized input arrays.
var errTrappolaOversized = errors.New("trappola: input array exceeds maximum allowed size")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Trappola) UnmarshalJSON(data []byte) error {
	var j trappolaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > trappolaMaxSliceLen || len(j.CurrentTrick) > trappolaMaxSliceLen ||
		len(j.ActionLog) > trappolaMaxSliceLen {
		return errTrappolaOversized
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		// **コンストラクタと同じデッキでないといけない。** クローン元は 40 枚の
		// ブリスコラデッキで、そのままだと 2 が混じる —— この 36 枚デッキに
		// 2 は無いので、強さも点も既定の枝に落ちて静かにずれる。
		g.trumpCards = NewTrumpCardsTrappola()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*TrappolaPlayer, 0)
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
