//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MendikotPhase メンディコットのゲームフェーズ
type MendikotPhase int

// Mendikot のフェーズ定数
const (
	// MendikotPhasePlay プレイ中（切り札は未定のまま始まる）
	//
	// **切り札を決めるための専用フェーズは無い。** フォローできなかった人が
	// そのとき出した札のスートがそのまま切り札になるので、プレイの流れが
	// 止まらない。
	MendikotPhasePlay MendikotPhase = iota
	// MendikotPhaseHandEnd ハンド終了
	MendikotPhaseHandEnd
	// MendikotPhaseGameEnd ゲーム終了
	MendikotPhaseGameEnd
)

// MendikotPlayerCnt プレイヤー数（4 人固定・2 対 2）
const MendikotPlayerCnt = 4

// MendikotTeamCnt チーム数
const MendikotTeamCnt = 2

// MendikotHandSize 各プレイヤーの手札枚数
const MendikotHandSize = 13

// MendikotTricksPerRound 1 ハンドのトリック数
const MendikotTricksPerRound = MendikotHandSize

// MendikotTensInDeck デッキに入っている 10 の枚数
const MendikotTensInDeck = 4

// MendikotWinPoints 通常の勝ちのハンド勝ち点
const MendikotWinPoints = 1

// MendikotMendikotPoints 10 を 4 枚とも取った (Mendikot) ときのハンド勝ち点
const MendikotMendikotPoints = 2

// MendikotWhitewashPoints 全 13 トリックを取った (Whitewash) ときのハンド勝ち点
const MendikotWhitewashPoints = 3

// mendikotMaxSliceLen caps slice sizes during deserialisation.
const mendikotMaxSliceLen = 1000

// Mendikot メンディコット ゲームクラス。
//
// インド・マハーラーシュトラ州のトリックテイキング。4 人 2 対 2（向かい合う席が
// 味方）、52 枚を 13 枚ずつ。
//
// **勝敗を決めるのはデッキに 4 枚しかない「10」。** カード点も契約も無く、
// トリック数ですらない——10 をどちらのチームが多く取ったかがすべて。
//
// **切り札は配られた時点では存在しない。** 最初にリードのスートをフォロー
// できなかったプレイヤーが、そこで自分の手札から 1 枚を選んでそのスートを
// 切り札にする。決まるまでは切り札なしで進む。
//
// **issue の勝敗条件は成立しない。** 「10 を 3 枚かつ 7 トリック以上」と
// 書かれているが、それだと *3 枚 + 6 トリック* のハンドで誰も勝たない。
// 本来の規則はこう:
//
//	10 が 3 枚以上   : そのチームの勝ち（トリック数は関係ない）
//	10 が 2 枚ずつ   : トリックの多いほうの勝ち（13 は奇数なので必ず決まる）
//	10 が 4 枚       : Mendikot（2 点）
//	全 13 トリック   : Whitewash（3 点）
//
// これなら引き分けが起こりえない。
type Mendikot struct {
	trumpCards *TrumpCards
	players    []*MendikotPlayer
	config     MendikotConfig

	phase       MendikotPhase
	handNumber  int
	trickNumber int
	// trumpSuit は切り札 (0: まだ決まっていない)。
	trumpSuit int
	// trumpChooserIdx は切り札を決めた／決めるプレイヤー (-1: まだ居ない)。
	trumpChooserIdx int

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int

	scores [MendikotTeamCnt]int
	// lastHandWinner / lastHandKind は直前のハンドの結末（表示用）。
	lastHandWinner int
	lastHandKind   string

	gameEndFlag bool
	winnerTeam  int

	actionLogBase
}

// NewMendikot コンストラクタ
func NewMendikot(trumpCards *TrumpCards, players []*MendikotPlayer, config MendikotConfig) *Mendikot {
	return &Mendikot{
		trumpCards: trumpCards, players: players, config: config,
		trumpChooserIdx: -1, lastHandWinner: -1, winnerTeam: -1,
	}
}

// NewDefaultMendikot 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultMendikot() *Mendikot {
	players := make([]*MendikotPlayer, 0, MendikotPlayerCnt)
	for i := range MendikotPlayerCnt {
		players = append(players, NewMendikotPlayer(i == 0))
	}
	return NewMendikot(NewTrumpCards(0), players, DefaultMendikotConfig())
}

// MendikotTeamOf 席のチーム番号。**向かい合う席が味方。**
func MendikotTeamOf(playerIdx int) int { return playerIdx % MendikotTeamCnt }

// Reset ゲーム全体を初期化する
func (m *Mendikot) Reset() {
	m.handNumber = 1
	m.dealerIdx = 0
	m.gameEndFlag = false
	m.winnerTeam = -1
	m.lastHandWinner = -1
	m.lastHandKind = ""
	m.scores = [MendikotTeamCnt]int{}
	m.actionLog = nil
	for _, p := range m.players {
		p.ResetGame()
	}
	m.dealHand()
}

// dealHand 13 枚ずつ配る。**切り札はまだ無い。**
func (m *Mendikot) dealHand() {
	m.phase = MendikotPhasePlay
	m.trickNumber = 0
	m.currentTrick = nil
	m.trumpSuit = 0
	m.trumpChooserIdx = -1
	for _, p := range m.players {
		p.ResetRound()
	}

	m.trumpCards = NewTrumpCards(0)
	m.trumpCards.Shuffle()
	for range MendikotHandSize {
		for i := range MendikotPlayerCnt {
			idx := (m.dealerIdx + 1 + i) % MendikotPlayerCnt
			if c := m.trumpCards.DrawCard(); c != nil {
				m.players[idx].AddCard(c)
			}
		}
	}
	m.sortAllHands()
	m.leadPlayerIdx = (m.dealerIdx + 1) % MendikotPlayerCnt
	m.currentPlayerIdx = m.leadPlayerIdx
	m.appendLog(-1, "deal", fmt.Sprintf("ハンド%d を開始（切り札は未定）", m.handNumber), nil)
}

// sortAllHands 手札をスート・ランク順に並べ替える
func (m *Mendikot) sortAllHands() {
	for _, p := range m.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return mendikotRank(ci) < mendikotRank(cj)
		})
	}
}

// mendikotRank 札の強さ。A が最強、以下 K,Q,J,10..2。
//
// **10 は強さの上では普通の札。** 勝敗を決めるのは強さではなく獲得枚数。
func mendikotRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (m *Mendikot) PlayerPlay(cardIndex int) error {
	if m.gameEndFlag {
		return errors.New("game has ended")
	}
	if m.phase != MendikotPhasePlay {
		return errors.New("not the play phase")
	}
	if m.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return m.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (m *Mendikot) CpuPlay() {
	if m.gameEndFlag || m.phase != MendikotPhasePlay || m.currentPlayerIdx == 0 {
		return
	}
	_ = m.play(m.currentPlayerIdx, m.chooseCpuCard(m.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (m *Mendikot) play(playerIdx, cardIndex int) error {
	p := m.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !m.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}

	// **切り札が未定で、この人が初めてフォローできなかったなら、ここで決まる。**
	// 出した札のスートがそのまま切り札になる。
	needsTrump := m.trumpSuit == 0 && len(m.currentTrick) > 0 &&
		card.GetDesign() != m.currentTrick[0].Card.GetDesign()

	p.RemoveCard(cardIndex)
	m.currentTrick = append(m.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	m.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if needsTrump {
		m.trumpSuit = card.GetDesign()
		m.trumpChooserIdx = playerIdx
		m.appendLog(playerIdx, "trump",
			fmt.Sprintf("フォローできず、切り札を %d に決めた", m.trumpSuit), []*Card{card})
	}

	if len(m.currentTrick) < MendikotPlayerCnt {
		m.currentPlayerIdx = (playerIdx + 1) % MendikotPlayerCnt
		return nil
	}
	m.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか
func (m *Mendikot) canPlay(playerIdx int, card *Card) bool {
	if len(m.currentTrick) == 0 {
		return true
	}
	leadSuit := m.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := m.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (m *Mendikot) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(m.players) {
		return nil
	}
	p := m.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if m.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決し、10 の枚数を数える
func (m *Mendikot) resolveTrick() {
	winner := m.trickWinner()
	cards := make([]*Card, 0, len(m.currentTrick))
	tens := 0
	for _, tc := range m.currentTrick {
		cards = append(cards, tc.Card)
		if tc.Card.GetValue() == 10 {
			tens++
		}
	}
	m.players[winner].AddTrick(cards)
	if tens > 0 {
		// **10 は勝敗そのもの。** 誰が取ったかを数える。
		m.players[winner].AddTens(tens)
		m.appendLog(winner, "ten", fmt.Sprintf("10 を %d 枚獲得", tens), nil)
	}

	m.trickNumber++
	m.currentTrick = nil
	m.leadPlayerIdx = winner
	m.currentPlayerIdx = winner

	if m.trickNumber >= MendikotTricksPerRound {
		m.finishHand()
	}
}

// trickWinner 現在のトリックの勝者
func (m *Mendikot) trickWinner() int {
	if len(m.currentTrick) == 0 {
		return m.leadPlayerIdx
	}
	leadSuit := m.currentTrick[0].Card.GetDesign()
	bestIdx, best := m.currentTrick[0].PlayerIdx, m.currentTrick[0].Card
	for _, tc := range m.currentTrick[1:] {
		if m.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// beats challenger が currentBest に勝つか。
//
// **切り札が未定のあいだはリードのスートだけが勝ちうる。**
func (m *Mendikot) beats(challenger, currentBest *Card, leadSuit int) bool {
	if m.trumpSuit != 0 {
		cTrump := challenger.GetDesign() == m.trumpSuit
		bTrump := currentBest.GetDesign() == m.trumpSuit
		if cTrump != bTrump {
			return cTrump
		}
	}
	if challenger.GetDesign() != currentBest.GetDesign() {
		return challenger.GetDesign() == leadSuit
	}
	return mendikotRank(challenger) > mendikotRank(currentBest)
}

// TeamTens チームが獲得した 10 の枚数
func (m *Mendikot) TeamTens(team int) int {
	if team < 0 || team >= MendikotTeamCnt {
		return 0
	}
	n := 0
	for i, p := range m.players {
		if MendikotTeamOf(i) == team {
			n += p.GetTens()
		}
	}
	return n
}

// TeamTricks チームの獲得トリック数
func (m *Mendikot) TeamTricks(team int) int {
	if team < 0 || team >= MendikotTeamCnt {
		return 0
	}
	n := 0
	for i, p := range m.players {
		if MendikotTeamOf(i) == team {
			n += p.GetTrickCount()
		}
	}
	return n
}

// MendikotHandResult ハンドの結末
type MendikotHandResult struct {
	// WinnerTeam 勝ったチーム
	WinnerTeam int
	// Points ハンド勝ち点
	Points int
	// Kind 結末の種類 ("whitewash" / "mendikot" / "tens" / "tricks")
	Kind string
}

// MendikotResultFor 10 の枚数とトリック数からハンドの結末を決める。
//
// **引き分けは起こりえない。** 10 が 2 枚ずつのときはトリックで決まり、
// 13 は奇数なので必ず片方が 7 以上になる。
func MendikotResultFor(tens0, tens1, tricks0, tricks1 int) MendikotHandResult {
	switch {
	case tricks0 == MendikotTricksPerRound:
		return MendikotHandResult{WinnerTeam: 0, Points: MendikotWhitewashPoints, Kind: "whitewash"}
	case tricks1 == MendikotTricksPerRound:
		return MendikotHandResult{WinnerTeam: 1, Points: MendikotWhitewashPoints, Kind: "whitewash"}
	case tens0 == MendikotTensInDeck:
		return MendikotHandResult{WinnerTeam: 0, Points: MendikotMendikotPoints, Kind: "mendikot"}
	case tens1 == MendikotTensInDeck:
		return MendikotHandResult{WinnerTeam: 1, Points: MendikotMendikotPoints, Kind: "mendikot"}
	case tens0 > tens1:
		return MendikotHandResult{WinnerTeam: 0, Points: MendikotWinPoints, Kind: "tens"}
	case tens1 > tens0:
		return MendikotHandResult{WinnerTeam: 1, Points: MendikotWinPoints, Kind: "tens"}
	case tricks0 > tricks1:
		return MendikotHandResult{WinnerTeam: 0, Points: MendikotWinPoints, Kind: "tricks"}
	default:
		return MendikotHandResult{WinnerTeam: 1, Points: MendikotWinPoints, Kind: "tricks"}
	}
}

// finishHand ハンドの勝ち点を確定させる
func (m *Mendikot) finishHand() {
	res := MendikotResultFor(m.TeamTens(0), m.TeamTens(1), m.TeamTricks(0), m.TeamTricks(1))
	m.scores[res.WinnerTeam] += res.Points
	m.lastHandWinner = res.WinnerTeam
	m.lastHandKind = res.Kind
	m.appendLog(-1, res.Kind,
		fmt.Sprintf("チーム%d が %s で +%d（10: %d-%d / トリック: %d-%d）",
			res.WinnerTeam, res.Kind, res.Points,
			m.TeamTens(0), m.TeamTens(1), m.TeamTricks(0), m.TeamTricks(1)), nil)

	// **負けたチームの席へ親が移る。** 勝ったチームは親を守る。
	if MendikotTeamOf(m.dealerIdx) == res.WinnerTeam {
		m.dealerIdx = (m.dealerIdx + 1) % MendikotPlayerCnt
	}

	if m.scores[res.WinnerTeam] >= m.config.Target {
		m.finishGame()
		return
	}
	m.phase = MendikotPhaseHandEnd
}

// NextHand 次のハンドを開始する
func (m *Mendikot) NextHand() {
	if m.gameEndFlag || m.phase != MendikotPhaseHandEnd {
		return
	}
	m.handNumber++
	m.dealHand()
}

// finishGame 規定点に達したチームの勝ち
func (m *Mendikot) finishGame() {
	m.phase = MendikotPhaseGameEnd
	m.gameEndFlag = true
	switch {
	case m.scores[0] > m.scores[1]:
		m.winnerTeam = 0
	case m.scores[1] > m.scores[0]:
		m.winnerTeam = 1
	default:
		m.winnerTeam = -1
	}
	m.appendLog(-1, "result", fmt.Sprintf("最終得点 %d - %d", m.scores[0], m.scores[1]), nil)
}

// chooseCpuCard CPU の手。**10 を取らせない／取ることを最優先にする。**
func (m *Mendikot) chooseCpuCard(playerIdx int) int {
	valid := m.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := m.players[playerIdx]

	if len(m.currentTrick) == 0 {
		bestIdx, bestRank := valid[0], mendikotRank(p.GetCard(valid[0]))
		for _, i := range valid[1:] {
			if r := mendikotRank(p.GetCard(i)); r > bestRank {
				bestIdx, bestRank = i, r
			}
		}
		return bestIdx
	}

	// **10 が場に出ているトリックは取りに行く。** 勝敗そのものなので。
	if m.trickHasTen() {
		if idx, ok := m.pickCheapestWinner(p, valid); ok {
			return idx
		}
	}
	if m.partnerIsWinning(playerIdx) {
		// 味方が勝っているなら 10 を乗せてよい。
		for _, i := range valid {
			if p.GetCard(i).GetValue() == 10 && !m.wouldWin(p.GetCard(i)) {
				return i
			}
		}
		return m.pickLowestNonTen(p, valid)
	}
	if idx, ok := m.pickCheapestWinner(p, valid); ok {
		return idx
	}
	// **取れないなら 10 を捨てない。** 相手に渡してしまう。
	return m.pickLowestNonTen(p, valid)
}

// trickHasTen 現在のトリックに 10 が出ているか
func (m *Mendikot) trickHasTen() bool {
	for _, tc := range m.currentTrick {
		if tc.Card.GetValue() == 10 {
			return true
		}
	}
	return false
}

// pickLowestNonTen 10 以外でいちばん弱い札。無ければいちばん弱い札。
func (m *Mendikot) pickLowestNonTen(p *MendikotPlayer, valid []int) int {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if c.GetValue() == 10 {
			continue
		}
		if r := mendikotRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	bestIdx, bestRank = valid[0], mendikotRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if r := mendikotRank(p.GetCard(i)); r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx
}

// partnerIsWinning 現時点で味方がトリックを取っているか
func (m *Mendikot) partnerIsWinning(playerIdx int) bool {
	if len(m.currentTrick) == 0 {
		return false
	}
	leadSuit := m.currentTrick[0].Card.GetDesign()
	best, bestIdx := m.currentTrick[0].Card, m.currentTrick[0].PlayerIdx
	for _, tc := range m.currentTrick[1:] {
		if m.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx != playerIdx && MendikotTeamOf(bestIdx) == MendikotTeamOf(playerIdx)
}

// pickCheapestWinner トリックを取れる札のうち一番弱いもの
func (m *Mendikot) pickCheapestWinner(p *MendikotPlayer, valid []int) (int, bool) {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !m.wouldWin(c) {
			continue
		}
		if r := mendikotRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx, bestIdx >= 0
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (m *Mendikot) wouldWin(c *Card) bool {
	if c == nil || len(m.currentTrick) == 0 {
		return true
	}
	leadSuit := m.currentTrick[0].Card.GetDesign()
	best := m.currentTrick[0].Card
	for _, tc := range m.currentTrick[1:] {
		if m.beats(tc.Card, best, leadSuit) {
			best = tc.Card
		}
	}
	return m.beats(c, best, leadSuit)
}

// MendikotHint ヒント情報
type MendikotHint struct {
	// CardIndex 推奨する手札のインデックス
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
}

// GetHint 人間プレイヤーへの推奨手を返す
func (m *Mendikot) GetHint() *MendikotHint {
	if m.gameEndFlag || !m.IsHumanTurn() || m.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := m.chooseCpuCard(0)
	reason := "mendikotDuck"
	switch {
	case m.trickHasTen():
		reason = "mendikotChaseTen"
	case m.partnerIsWinning(0):
		reason = "mendikotFeedPartner"
	}
	return &MendikotHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (m *Mendikot) GetPhase() MendikotPhase { return m.phase }

// GetConfig 現在の設定
func (m *Mendikot) GetConfig() MendikotConfig { return m.config }

// SetConfig 設定を差し替える
func (m *Mendikot) SetConfig(c MendikotConfig) { m.config = c }

// GetHandNumber 現在のハンド番号（1 起点）
func (m *Mendikot) GetHandNumber() int { return m.handNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (m *Mendikot) GetTrickNumber() int { return m.trickNumber }

// GetTrumpSuit 切り札のスート（未決定は 0）
func (m *Mendikot) GetTrumpSuit() int { return m.trumpSuit }

// GetTrumpChooserIdx 切り札を決めたプレイヤー (-1: 未決定)
func (m *Mendikot) GetTrumpChooserIdx() int { return m.trumpChooserIdx }

// GetScore チームのハンド勝ち点
func (m *Mendikot) GetScore(team int) int {
	if team < 0 || team >= MendikotTeamCnt {
		return 0
	}
	return m.scores[team]
}

// SetScoreForTestUse チームのハンド勝ち点を設定する（復元・テスト用）
func (m *Mendikot) SetScoreForTestUse(team, n int) {
	if team >= 0 && team < MendikotTeamCnt {
		m.scores[team] = n
	}
}

// GetLastHandWinner 直前のハンドを制したチーム (-1: まだ無い)
func (m *Mendikot) GetLastHandWinner() int { return m.lastHandWinner }

// GetLastHandKind 直前のハンドの結末の種類
func (m *Mendikot) GetLastHandKind() string { return m.lastHandKind }

// GetCurrentTrick 現在のトリック
func (m *Mendikot) GetCurrentTrick() []*TrickCard { return m.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (m *Mendikot) GetCurrentPlayerIdx() int { return m.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (m *Mendikot) GetLeadPlayerIdx() int { return m.leadPlayerIdx }

// GetDealerIdx ディーラー
func (m *Mendikot) GetDealerIdx() int { return m.dealerIdx }

// GetPlayerCnt プレイヤー数
func (m *Mendikot) GetPlayerCnt() int { return len(m.players) }

// GetPlayer 指定インデックスのプレイヤー
func (m *Mendikot) GetPlayer(i int) *MendikotPlayer {
	if i < 0 || i >= len(m.players) {
		return nil
	}
	return m.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (m *Mendikot) GetGameEndFlag() bool { return m.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1: 未確定/同点)
func (m *Mendikot) GetWinnerTeam() int { return m.winnerTeam }

// IsHumanTurn 人間の手番か
func (m *Mendikot) IsHumanTurn() bool {
	return !m.gameEndFlag && m.phase == MendikotPhasePlay && m.currentPlayerIdx == 0
}

// GiveUp 投了する
func (m *Mendikot) GiveUp() {
	if m.gameEndFlag {
		return
	}
	m.phase = MendikotPhaseGameEnd
	m.gameEndFlag = true
	m.winnerTeam = 1
	m.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (m *Mendikot) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	m.appendLogAt(m.trickNumber, playerIdx, actionType, detail, cards)
}

// mendikotJSON is the KV snapshot format for Mendikot.
type mendikotJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*MendikotPlayer    `json:"pl"`
	Config           MendikotConfig       `json:"cf"`
	Phase            MendikotPhase        `json:"ph"`
	HandNumber       int                  `json:"hn"`
	TrickNumber      int                  `json:"tn"`
	TrumpSuit        int                  `json:"ts"`
	TrumpChooserIdx  int                  `json:"tx"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	CurrentPlayerIdx int                  `json:"cp"`
	LeadPlayerIdx    int                  `json:"lp"`
	DealerIdx        int                  `json:"di"`
	Scores           [MendikotTeamCnt]int `json:"sc"`
	LastHandWinner   int                  `json:"lw"`
	LastHandKind     string               `json:"lk"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerTeam       int                  `json:"wt"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (m *Mendikot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&mendikotJSON{
		TrumpCards:       m.trumpCards,
		Players:          m.players,
		Config:           m.config,
		Phase:            m.phase,
		HandNumber:       m.handNumber,
		TrickNumber:      m.trickNumber,
		TrumpSuit:        m.trumpSuit,
		TrumpChooserIdx:  m.trumpChooserIdx,
		CurrentTrick:     m.currentTrick,
		CurrentPlayerIdx: m.currentPlayerIdx,
		LeadPlayerIdx:    m.leadPlayerIdx,
		DealerIdx:        m.dealerIdx,
		Scores:           m.scores,
		LastHandWinner:   m.lastHandWinner,
		LastHandKind:     m.lastHandKind,
		GameEndFlag:      m.gameEndFlag,
		WinnerTeam:       m.winnerTeam,
		ActionLog:        m.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。値域を検証する。
func (m *Mendikot) UnmarshalJSON(data []byte) error {
	var j mendikotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < MendikotPhasePlay || j.Phase > MendikotPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **切り札は 0（未決定）か実在するスート。** 素通しすると beats() が
	// どの札も切り札と見なさなくなり、勝敗が黙って変わる (#5302〜#5305)。
	// このゲームは 0 が正当な状態なので、フェーズではなく決定者と突き合わせる。
	if j.TrumpSuit == 0 {
		if j.TrumpChooserIdx != -1 {
			return fmt.Errorf("trump chooser %d without a trump suit", j.TrumpChooserIdx)
		}
	} else {
		if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
			return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
		}
		if j.TrumpChooserIdx < 0 || j.TrumpChooserIdx >= MendikotPlayerCnt {
			return fmt.Errorf("trump suit %d without a chooser", j.TrumpSuit)
		}
	}
	if j.TrickNumber < 0 || j.TrickNumber > MendikotTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.HandNumber < 1 {
		return fmt.Errorf("invalid hand number: %d", j.HandNumber)
	}
	if len(j.ActionLog) > mendikotMaxSliceLen {
		return errors.New("mendikot: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > MendikotPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= MendikotPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	for name, team := range map[string]int{
		"winner team":      j.WinnerTeam,
		"last hand winner": j.LastHandWinner,
	} {
		if team < -1 || team >= MendikotTeamCnt {
			return fmt.Errorf("invalid %s: %d", name, team)
		}
	}
	if j.TrumpCards != nil {
		m.trumpCards = j.TrumpCards
	}
	if len(j.Players) == MendikotPlayerCnt {
		m.players = j.Players
	}
	m.config = j.Config
	m.phase = j.Phase
	m.handNumber = j.HandNumber
	m.trickNumber = j.TrickNumber
	m.trumpSuit = j.TrumpSuit
	m.trumpChooserIdx = j.TrumpChooserIdx
	m.currentTrick = j.CurrentTrick
	m.currentPlayerIdx = j.CurrentPlayerIdx
	m.leadPlayerIdx = j.LeadPlayerIdx
	m.dealerIdx = j.DealerIdx
	m.scores = j.Scores
	m.lastHandWinner = j.LastHandWinner
	m.lastHandKind = j.LastHandKind
	m.gameEndFlag = j.GameEndFlag
	m.winnerTeam = j.WinnerTeam
	m.actionLog = j.ActionLog
	return nil
}
