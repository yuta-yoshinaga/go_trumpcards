//go:build !js || !wasm || extra3

// Package domain マドラッソ (Madrasso) のドメインモデル。
//
// Madrasso はイタリアの3大国民的カードゲームの一つで、切り札を持たない
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

// MadrassoPlayerCnt マドラッソのプレイヤー数
const MadrassoPlayerCnt = 4

// MadrassoHandSize 各プレイヤーの手札枚数 (40 / 4)
const MadrassoHandSize = 10

// MadrassoTrickCount 1ラウンドのトリック数
const MadrassoTrickCount = 10

// MadrassoTeamCnt チーム数
const MadrassoTeamCnt = 2

// MadrassoUltimaPoints 最終トリック勝者へのボーナス (整数点)
const MadrassoUltimaPoints = 1

// MadrassoRoundPoints 1ラウンドで奪い合う得点の総和 (整数点)。
//
// **1/3 点刻みではない。** クローン元のトレセッテは A=3, 絵札=1 の
// 1/3 点単位で 33 を積むが、こちらはブリスコラ系の整数点
// (A=11, 3=10, K=4, Q=3, J=2) で 1 スートあたり 30、4 スートで 120。
// これに最終トリックのボーナス 1 を足す。
// **手で置いた数ではなくデッキから数えた値** (テストが突き合わせる)。
const MadrassoRoundPoints = 121

// MadrassoPhase ゲームフェーズ
type MadrassoPhase int

// Madrasso のフェーズ定数
const (
	// MadrassoPhasePlay トリックプレイフェーズ
	MadrassoPhasePlay MadrassoPhase = 0
	// MadrassoPhaseTrickEnd トリック終了フェーズ (解決済み・次トリック待ち)
	MadrassoPhaseTrickEnd MadrassoPhase = 1
	// MadrassoPhaseRoundEnd ラウンド終了フェーズ
	MadrassoPhaseRoundEnd MadrassoPhase = 2
	// MadrassoPhaseGameEnd ゲーム終了フェーズ
	MadrassoPhaseGameEnd MadrassoPhase = 3
)

// MadrassoHint ヒント情報
type MadrassoHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Madrasso マドラッソのゲームクラス
type Madrasso struct {
	trumpCards       *TrumpCards
	players          []*MadrassoPlayer
	config           MadrassoConfig
	phase            MadrassoPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	teamScores       [MadrassoTeamCnt]int // 累積点 (整数点)
	teamRoundPoints  [MadrassoTeamCnt]int // 現ラウンドで獲得した整数点の合計
	// trumpSuit は配り終えた最後の 1 枚のスート (-1=未確定)。
	trumpSuit   int
	gameEndFlag bool
	winnerTeam  int
	actionLogBase
}

// NewMadrasso コンストラクタ
func NewMadrasso(trumpCards *TrumpCards, players []*MadrassoPlayer, config MadrassoConfig) *Madrasso {
	return &Madrasso{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 0,
	}
}

// NewDefaultMadrasso returns Madrasso with the standard 4-player setup
// (1 human, 3 CPU) and DefaultMadrassoConfig. Single source of truth for CUI,
// Web, and Worker construction.
func NewDefaultMadrasso() *Madrasso {
	players := []*MadrassoPlayer{
		NewMadrassoPlayer(true),
		NewMadrassoPlayer(false),
		NewMadrassoPlayer(false),
		NewMadrassoPlayer(false),
	}
	return NewMadrasso(NewTrumpCardsBriscola(), players, DefaultMadrassoConfig())
}

// MadrassoTeamOf プレイヤーインデックスが属するチーム (0 = 0&2, 1 = 1&3)
func MadrassoTeamOf(playerIdx int) int { return playerIdx % MadrassoTeamCnt }

// Reset ゲーム初期化: デッキをシャッフルして配り、最初のラウンドを開始する。
func (g *Madrasso) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.teamScores = [MadrassoTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Madrasso) NextRound() {
	if g.phase != MadrassoPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.startRound()
}

// startRound 手札を配り、リードプレイヤーを決めてプレイフェーズを開始する。
func (g *Madrasso) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.teamRoundPoints = [MadrassoTeamCnt]int{}

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	dealAllCards(g.trumpCards, g.players)
	// **切り札は配りで決まる。** 最後に配られた 1 枚のスートを採る ——
	// 誰も選ばないので入力面が増えず、盤面から説明でき、テストで固定できる。
	// (issue は「例: 最後に配られたカードのスート」とだけ書いている。)
	g.trumpSuit = madrassoLastDealtSuit(g.players)
	g.sortAllHands()

	g.leadPlayerIdx = (g.roundNumber - 1) % MadrassoPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = MadrassoPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Madrasso) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MadrassoPhasePlay {
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
func (g *Madrasso) CpuPlay() {
	if g.gameEndFlag || g.phase != MadrassoPhasePlay {
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
func (g *Madrasso) ResolveTrick() {
	if g.phase != MadrassoPhaseTrickEnd || len(g.currentTrick) != MadrassoPlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	thirds := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		thirds += madrassoPoints(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)

	team := MadrassoTeamOf(winnerIdx)
	g.teamRoundPoints[team] += thirds
	bonus := ""
	if g.trickNumber >= MadrassoTrickCount {
		g.teamRoundPoints[team] += MadrassoUltimaPoints
		bonus = " +ultima"
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d/3%s)", playerName(g.players, winnerIdx), g.trickNumber, thirds, bonus),
		trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= MadrassoTrickCount {
		g.phase = MadrassoPhaseRoundEnd
	} else {
		g.phase = MadrassoPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *Madrasso) NextTrick() {
	if g.phase != MadrassoPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = MadrassoPhasePlay
}

// ScoreRound ラウンドの得点を確定し、ゲーム終了判定を行う。1/3点を 3 で割って
// (端数切り捨て) 累積点へ加算する。
func (g *Madrasso) ScoreRound() {
	if g.phase != MadrassoPhaseRoundEnd {
		return
	}

	// **ディールを取ったチームに 1 点。** クローン元のトレセッテは 1/3 点を
	// 3 で割って積むが、こちらのカード点は 1 ディールで 120 点あり、
	// そのまま積むと 1 ディールで目標を越える。ブリスコラ系と同じく
	// 「過半 (61 点以上) を取った側がそのディールの勝ち」で数える。
	winner := -1
	switch {
	case g.teamRoundPoints[0]*2 > MadrassoRoundPoints-MadrassoUltimaPoints:
		winner = 0
	case g.teamRoundPoints[1]*2 > MadrassoRoundPoints-MadrassoUltimaPoints:
		winner = 1
	}
	if winner >= 0 {
		g.teamScores[winner]++
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: TeamA=%d pts, TeamB=%d pts -> deal to %s (scores %d-%d)",
			g.roundNumber, g.teamRoundPoints[0], g.teamRoundPoints[1],
			madrassoDealResultName(winner), g.teamScores[0], g.teamScores[1]), nil)

	leader, other := 0, 1
	if g.teamScores[1] > g.teamScores[0] {
		leader, other = 1, 0
	}
	if g.teamScores[leader] >= g.config.TargetPoints && g.teamScores[leader] > g.teamScores[other] {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = MadrassoPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the game!", madrassoTeamName(leader)), nil)
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Madrasso) GetPhase() MadrassoPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Madrasso) SetPhase(phase MadrassoPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Madrasso) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Madrasso) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Madrasso) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Madrasso) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Madrasso) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Madrasso) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Madrasso) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Madrasso) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Madrasso) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Madrasso) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetTeamScores チーム別累積点を取得
func (g *Madrasso) GetTeamScores() [MadrassoTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点を設定 (テスト用)
func (g *Madrasso) SetTeamScores(s [MadrassoTeamCnt]int) { g.teamScores = s }

// GetTeamRoundPoints チーム別の現ラウンド 1/3点 を取得
func (g *Madrasso) GetTeamRoundPoints() [MadrassoTeamCnt]int { return g.teamRoundPoints }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Madrasso) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Madrasso) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Madrasso) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Madrasso) GetPlayer(i int) *MadrassoPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Madrasso) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Madrasso) GetConfig() MadrassoConfig { return g.config }

// SetConfig 設定変更
func (g *Madrasso) SetConfig(cfg MadrassoConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能な (マストフォローを満たす) カードのインデックス一覧を返す。
func (g *Madrasso) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- Private methods ---

// playCard カードをプレイする共通処理
func (g *Madrasso) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == MadrassoPlayerCnt {
		g.phase = MadrassoPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % MadrassoPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する (マストフォロー)
func (g *Madrasso) validatePlay(playerIdx int, card *Card) error {
	return validateFollowSuit(g.currentTrick, g.players, playerIdx, card)
}

// trickWinner トリックの勝者を決定する。切り札がないため、リードスートの最強札が勝つ。
func (g *Madrasso) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	// **切り札がある。** クローン元のトレセッテは切り札を持たないので
	// 「リードスートの最強札」だけで決まったが、こちらは配りで切り札が
	// 決まるため、切り札はどの平札にも勝つ。
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card

	for _, tc := range g.currentTrick[1:] {
		if madrassoBeats(tc.Card, winnerCard, leadSuit, g.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// madrassoBeats は challenger が best を上回るかを返す。
//
// 切り札 > リードスート > それ以外。同じ組の中では札位の強さで比べる。
func madrassoBeats(challenger, best *Card, leadSuit, trumpSuit int) bool {
	cTrump := challenger.GetDesign() == trumpSuit
	bTrump := best.GetDesign() == trumpSuit
	switch {
	case cTrump && !bTrump:
		return true
	case !cTrump && bTrump:
		return false
	case cTrump && bTrump:
		return madrassoStrength(challenger.GetValue()) > madrassoStrength(best.GetValue())
	}
	// どちらも平札。リードスートに従っていない札は勝てない。
	if challenger.GetDesign() != leadSuit {
		return false
	}
	if best.GetDesign() != leadSuit {
		return true
	}
	return madrassoStrength(challenger.GetValue()) > madrassoStrength(best.GetValue())
}

// MadrassoBeatsForTest は madrassoBeats を公開する (テスト用)。
func MadrassoBeatsForTest(challenger, best *Card, leadSuit, trumpSuit int) bool {
	return madrassoBeats(challenger, best, leadSuit, trumpSuit)
}

// madrassoLastDealtSuit は最後に配られた 1 枚のスートを返す。
//
// dealAllCards は席を順に回して配るので、**最後に配られるのは
// 最後の席の最後の 1 枚**。手札を並べ替える前に読むこと (sortAllHands の後だと
// 「最後に配られた札」ではなくなる)。
func madrassoLastDealtSuit(players []*MadrassoPlayer) int {
	last := players[len(players)-1]
	if n := last.GetCardsSize(); n > 0 {
		if c := last.GetCard(n - 1); c != nil {
			return c.GetDesign()
		}
	}
	return -1
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Madrasso) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *Madrasso) sortAllHands() {
	for _, p := range g.players {
		madrassoSortHand(p)
	}
}

// madrassoSortHand プレイヤーの手札をスート→強さの順にソートする
func madrassoSortHand(p *MadrassoPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return madrassoStrength(ci.GetValue()) < madrassoStrength(cj.GetValue())
	})
}

// madrassoTeamName チーム表示名 (0=A, 1=B)。
//
// **クローン元の teamName は共有できない。** あちら (Tressette.go) は casino
// タグ、こちらは extra3 なので、extra3 のビルドでは定義ごと消える。
func madrassoTeamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// --- Card helpers ---

// madrassoStrength トリックの強さ。3 が最強 (9)、4 が最弱 (0)。
//
//	3 > 2 > A > K > Q > J > 7 > 6 > 5 > 4
//
// madrassoStrength は札位の強さを返す。**順は A>3>K>Q>J>7>6>5>4>2。**
//
// **点の高い札ほど強い**というブリスコラ系の並びで、採用したカード点
// (A=11, 3=10, K=4, Q=3, J=2) と一致する。クローン元のトレセッテは
// 3>2>A>K>... で、2 が A より強いという別の並びを持つ ── 点の表と
// 食い違ったままにすると、一番高い札が一番弱いという盤面ができる。
func madrassoStrength(value int) int {
	switch value {
	case 1: // Asso
		return 9
	case 3: // Tre
		return 8
	case 13: // Re
		return 7
	case 12: // Cavallo
		return 6
	case 11: // Fante
		return 5
	case 7:
		return 4
	case 6:
		return 3
	case 5:
		return 2
	case 4:
		return 1
	default: // 2
		return 0
	}
}

// madrassoDealResultName はディールの結果をログ用の文字列にする。
func madrassoDealResultName(winner int) string {
	if winner < 0 {
		return "nobody (tied)"
	}
	return "Team " + madrassoTeamName(winner)
}

// SetTeamRoundPointsForTest はチームの現ラウンド獲得点を設定する (テスト用)。
func (g *Madrasso) SetTeamRoundPointsForTest(team, pts int) {
	if team >= 0 && team < MadrassoTeamCnt {
		g.teamRoundPoints[team] = pts
	}
}

// SetTrumpSuitForTest は切り札スートを設定する (テスト用)。
func (g *Madrasso) SetTrumpSuitForTest(suit int) { g.trumpSuit = suit }

// MadrassoStrengthForTest は札位の強さを返す (テスト用)。
func MadrassoStrengthForTest(value int) int { return madrassoStrength(value) }

// madrassoPoints カードの得点を 1/3点 単位で返す。A=3、2/3/J/Q/K=1、その他=0。
// madrassoPoints はカード点を**整数**で返す。
//
// **1/3 点刻みではない。** クローン元のトレセッテは A=3 / 絵札=1 の
// 1/3 点単位で数えるが、こちらはヴェネトのブリスコラ系と同じ整数点。
// 1 スートあたり 11+10+4+3+2 = 30 点、4 スートで 120 点。
func madrassoPoints(value int) int {
	switch value {
	case 1: // Asso
		return 11
	case 3: // Tre
		return 10
	case 13: // Re
		return 4
	case 12: // Cavallo
		return 3
	case 11: // Fante
		return 2
	default:
		return 0
	}
}

// MadrassoPointsForTest はカード点を返す (テスト用)。
func MadrassoPointsForTest(value int) int { return madrassoPoints(value) }

// GetTrumpSuit は配りで決まった切り札スートを返す (-1=未確定)。
func (g *Madrasso) GetTrumpSuit() int { return g.trumpSuit }

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨プレイを返す。
func (g *Madrasso) GetHint() *MadrassoHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != MadrassoPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &MadrassoHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason ヒント理由を判定する
func (g *Madrasso) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStrength := madrassoStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	if madrassoStrength(card.GetValue()) > topStrength {
		return "follow_win"
	}
	if MadrassoTeamOf(winnerIdx) == MadrassoTeamOf(playerIdx) {
		return "give_partner"
	}
	return "follow_duck"
}

// indexOfPlayerInTrick currentTrick 内で playerIdx が出したカードの位置を返す (-1=なし)
func (g *Madrasso) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (g *Madrasso) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == MadrassoCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点を意識した戦略プレイ
func (g *Madrasso) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	// リード: 得点・強さの低いカードを出して温存する。
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int {
			return madrassoPoints(c.GetValue())*100 + madrassoStrength(c.GetValue())
		})
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topStrength := madrassoStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	partnerWinning := MadrassoTeamOf(winnerIdx) == MadrassoTeamOf(playerIdx)
	trickPoints := 0
	for _, tc := range g.currentTrick {
		trickPoints += madrassoPoints(tc.Card.GetValue())
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
			return madrassoPoints(c.GetValue())*100 + madrassoStrength(c.GetValue())
		})
	}

	winners := filterIndices(follows, func(idx int) bool {
		return madrassoStrength(player.GetCard(idx).GetValue()) > topStrength
	})

	if partnerWinning {
		// 味方が勝っている: 得点札を渡しつつ、無駄に上書きしない。
		nonWinners := filterIndices(follows, func(idx int) bool {
			return madrassoStrength(player.GetCard(idx).GetValue()) < topStrength
		})
		if len(nonWinners) > 0 {
			// 勝ちを取らない範囲で最も得点の高い札を渡す。
			return pickHighest(player, nonWinners, func(c *Card) int {
				return madrassoPoints(c.GetValue())*100 - madrassoStrength(c.GetValue())
			})
		}
		// 上書きせざるを得ない場合は最弱札で被害を抑える。
		return pickLowest(player, follows, func(c *Card) int { return madrassoStrength(c.GetValue()) })
	}

	// 相手が勝っている: 得点があり勝てるなら最小限の札で取りに行く。
	if trickPoints > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return madrassoStrength(c.GetValue()) })
	}
	// 取れない/取る価値がない: 得点・強さの低い札でダックする。
	return pickLowest(player, follows, func(c *Card) int {
		return madrassoPoints(c.GetValue())*100 + madrassoStrength(c.GetValue())
	})
}

// --- JSON ---

// madrassoJSON is the JSON wire format for Madrasso.
type madrassoJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*MadrassoPlayer    `json:"ps"`
	Config           MadrassoConfig       `json:"cf"`
	Phase            MadrassoPhase        `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	LeadPlayerIdx    int                  `json:"li"`
	TeamScores       [MadrassoTeamCnt]int `json:"ts"`
	TeamRoundPoints  [MadrassoTeamCnt]int `json:"tr"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerTeam       int                  `json:"wt"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Madrasso) MarshalJSON() ([]byte, error) {
	return json.Marshal(madrassoJSON{
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
		TeamRoundPoints:  g.teamRoundPoints,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// madrassoMaxSliceLen caps slice sizes during deserialisation.
const madrassoMaxSliceLen = 1000

// errMadrassoOversized is the single sentinel error for oversized input arrays.
var errMadrassoOversized = errors.New("madrasso: input array exceeds maximum allowed size")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Madrasso) UnmarshalJSON(data []byte) error {
	var j madrassoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > madrassoMaxSliceLen || len(j.CurrentTrick) > madrassoMaxSliceLen ||
		len(j.ActionLog) > madrassoMaxSliceLen {
		return errMadrassoOversized
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBriscola()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*MadrassoPlayer, 0)
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
	g.teamRoundPoints = j.TeamRoundPoints
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
