//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MinibridgePhase はミニブリッジのフェーズ。
type MinibridgePhase int

const (
	// MinibridgePhaseContract は落札者が契約を選ぶフェーズ。
	MinibridgePhaseContract MinibridgePhase = iota
	// MinibridgePhasePlay はトリックプレイのフェーズ。
	MinibridgePhasePlay
	// MinibridgePhaseRoundEnd はディール終了。
	MinibridgePhaseRoundEnd
	// MinibridgePhaseGameEnd は終局。
	MinibridgePhaseGameEnd
)

// MinibridgePlayerCnt はプレイヤー数。
const MinibridgePlayerCnt = 4

// MinibridgeTeamCnt はチーム数。
const MinibridgeTeamCnt = 2

// MinibridgeHandSize は 1 人あたりの手札枚数。
const MinibridgeHandSize = 13

// MinibridgeTotalTricks は 1 ディールのトリック数。
const MinibridgeTotalTricks = MinibridgeHandSize

// MinibridgeTotalHcp は場に出るハイカードポイントの総和。
//
// **常にちょうど 40。** A4 + K3 + Q2 + J1 = 10 が 4 スートぶん。だから
// 「合計が高いペア」は 20-20 で決まらないことがあり（実測 8.1%）、
// issue が触れていないこの分岐にタイブレークが要ります。
const MinibridgeTotalHcp = 40

// MinibridgeMaxLevel は契約の最大レベル。
const MinibridgeMaxLevel = 7

// MinibridgeBookTricks は契約に上乗せされるブック。レベル n の契約は 6+n トリック。
const MinibridgeBookTricks = 6

// MinibridgeGameThreshold は game ボーナスの境目となる契約点。
const MinibridgeGameThreshold = 100

// 得点定数。**出典を写さず、全ケースが 1 つの式で一貫するように決めています。**
const (
	// minibridgeMinorPerTrick はマイナー（♣♦）1 トリックあたりの契約点。
	minibridgeMinorPerTrick = 20
	// minibridgeMajorPerTrick はメジャー（♠♥）1 トリックあたりの契約点。
	minibridgeMajorPerTrick = 30
	// minibridgeNoTrumpFirst はノートランプの 1 トリック目。
	minibridgeNoTrumpFirst = 40
	// minibridgeNoTrumpRest はノートランプの 2 トリック目以降。
	minibridgeNoTrumpRest = 30
	// minibridgeGameBonus は契約点 100 以上のボーナス。
	minibridgeGameBonus = 300
	// minibridgePartScoreBonus は契約点 100 未満のボーナス。
	minibridgePartScoreBonus = 50
	// minibridgeSmallSlamBonus はレベル 6 のボーナス。
	minibridgeSmallSlamBonus = 500
	// minibridgeGrandSlamBonus はレベル 7 のボーナス。
	minibridgeGrandSlamBonus = 1000
	// minibridgeUndertrickPenalty は不足 1 トリックあたりの相手得点。
	minibridgeUndertrickPenalty = 50
)

// minibridgeMaxSliceLen は復元時に受け付けるスライスの上限。
const minibridgeMaxSliceLen = 1000

// minibridgeHcpValue は 1 枚のハイカードポイント。**A4 K3 Q2 J1、他は 0。**
func minibridgeHcpValue(c *Card) int {
	switch c.GetValue() {
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

// minibridgeRank は札の強さ。**A が最強。**
func minibridgeRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// minibridgeIsMinor はマイナースート（♣♦）かを返す。
func minibridgeIsMinor(suit int) bool {
	return suit == CardDesignClover || suit == CardDesignDiamond
}

// MinibridgeHint はミニブリッジの助言。
type MinibridgeHint struct {
	CardIndex *int
	Reason    string
	// Level / Suit は選ぶべき契約（プレイ中は 0）。
	Level int
	Suit  int
}

// Minibridge はミニブリッジのゲーム。
type Minibridge struct {
	trumpCards *TrumpCards
	players    []*MinibridgePlayer
	config     MinibridgeConfig
	phase      MinibridgePhase
	actionLogBase

	roundNumber int
	trickNumber int

	// contractLevel は 0 のあいだ未決定。
	contractLevel int
	// contractSuit は 0 がノートランプ。**「未決定」ではありません。**
	contractSuit int
	declarerIdx  int
	dummyIdx     int

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int

	teamScores [MinibridgeTeamCnt]int
	lastMade   bool
	lastTricks int

	gameEndFlag bool
	// winnerTeam は -1 のあいだ未確定。同点も -1。
	winnerTeam int
}

// NewMinibridge はコンストラクタ。
//
// **4 人ちょうどでなければ標準セットアップに差し替える** (#5312 のレビュー指摘と
// 同じ理由——席数は下流のどこでも固定の前提)。
func NewMinibridge(players []*MinibridgePlayer, config MinibridgeConfig) *Minibridge {
	if len(players) != MinibridgePlayerCnt {
		players = newMinibridgeSeats()
	}
	return &Minibridge{
		players:     players,
		config:      config,
		declarerIdx: -1,
		dummyIdx:    -1,
		winnerTeam:  -1,
	}
}

// newMinibridgeSeats は標準の席順（人間 1 + CPU 3、席 0/2 が同じチーム）を返す。
func newMinibridgeSeats() []*MinibridgePlayer {
	seats := make([]*MinibridgePlayer, 0, MinibridgePlayerCnt)
	for i := range MinibridgePlayerCnt {
		seats = append(seats, NewMinibridgePlayer(i == 0, i%MinibridgeTeamCnt))
	}
	return seats
}

// NewDefaultMinibridge は標準の 4 人セットアップを返す。
func NewDefaultMinibridge() *Minibridge {
	return NewMinibridge(newMinibridgeSeats(), DefaultMinibridgeConfig())
}

// Reset はゲームを初期化する。
func (m *Minibridge) Reset() {
	for _, p := range m.players {
		p.ResetGame()
	}
	m.teamScores = [MinibridgeTeamCnt]int{}
	m.roundNumber = 1
	m.dealerIdx = 0
	m.gameEndFlag = false
	m.winnerTeam = -1
	m.actionLog = nil
	m.addLog(-1, "start", "ミニブリッジを開始しました", nil)
	m.startRound()
}

// startRound は 1 ディールを配って契約フェーズに入る。
func (m *Minibridge) startRound() {
	for _, p := range m.players {
		p.ResetRound()
	}
	m.trumpCards = NewTrumpCards(0)
	m.trumpCards.Shuffle()
	// **親の左隣から配る。** 席順が親と一緒に回るので、同じ手が同じ席に来ない。
	for range MinibridgeHandSize {
		for i := range MinibridgePlayerCnt {
			idx := (m.dealerIdx + 1 + i) % MinibridgePlayerCnt
			if c := m.trumpCards.DrawCard(); c != nil {
				m.players[idx].AddCard(c)
			}
		}
	}
	m.sortAllHands()
	m.announceHcp()
	m.decideDeclarer()

	m.contractLevel, m.contractSuit = 0, 0
	m.trickNumber = 0
	m.currentTrick = nil
	m.phase = MinibridgePhaseContract
	m.currentPlayerIdx = m.declarerIdx
	m.leadPlayerIdx = m.nextSeat(m.declarerIdx)
}

// sortAllHands は手札をスート・ランク順に整える。
func (m *Minibridge) sortAllHands() {
	for _, p := range m.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return minibridgeRank(ci) < minibridgeRank(cj)
		})
	}
}

// announceHcp は各席の HCP を数えて公開する。
//
// **このゲームに競りは無く、これが唯一の公開情報。** 親から順に申告した順序を
// 棋譜に残します（盤面には全員ぶんが常時出ています）。
func (m *Minibridge) announceHcp() {
	total := 0
	for step := range MinibridgePlayerCnt {
		idx := (m.dealerIdx + step) % MinibridgePlayerCnt
		p := m.players[idx]
		hcp := 0
		for i := range p.GetCardsSize() {
			hcp += minibridgeHcpValue(p.GetCard(i))
		}
		p.SetHcp(hcp)
		total += hcp
		m.addLog(idx, "hcp", fmt.Sprintf("HCP %d を申告しました", hcp), nil)
	}
	// 配り切りなので必ず 40。崩れていたら配りのバグ。
	if total != MinibridgeTotalHcp {
		m.addLog(-1, "hcp", fmt.Sprintf("HCP 合計が %d です（想定 %d）", total, MinibridgeTotalHcp), nil)
	}
}

// decideDeclarer は申告済みの HCP から落札者とダミーを決める。
//
// **issue はここに 2 つ穴を残している。** 総和が必ず 40 なので
// 「合計が高いペア」は 20-20 で決まらず（実測 8.1%）、ペアが決まっても
// 席が同点のことがある（4.4%）。どちらも決定的に解く:
//
//   - ペア同点 → **親の側**が宣言側（再配りは停止性の懸念があるので採らない）
//   - 席同点 → **親から見て先の席**
func (m *Minibridge) decideDeclarer() {
	teamHcp := [MinibridgeTeamCnt]int{}
	for _, p := range m.players {
		teamHcp[p.GetTeam()] += p.GetHcp()
	}

	declTeam := 0
	switch {
	case teamHcp[0] > teamHcp[1]:
		declTeam = 0
	case teamHcp[1] > teamHcp[0]:
		declTeam = 1
	default:
		declTeam = m.players[m.dealerIdx].GetTeam()
		m.addLog(-1, "declarer",
			fmt.Sprintf("HCP が %d-%d の同点のため、親の側が宣言側になります", teamHcp[0], teamHcp[1]), nil)
	}

	// 親から順に見るので、席が同点なら自然と先の席が残る。
	best := -1
	for step := range MinibridgePlayerCnt {
		idx := (m.dealerIdx + step) % MinibridgePlayerCnt
		if m.players[idx].GetTeam() != declTeam {
			continue
		}
		if best < 0 || m.players[idx].GetHcp() > m.players[best].GetHcp() {
			best = idx
		}
	}
	m.declarerIdx = best
	m.dummyIdx = (best + MinibridgeTeamCnt) % MinibridgePlayerCnt
	m.addLog(m.declarerIdx, "declarer",
		fmt.Sprintf("HCP %d でデクレアラーになりました（ダミーは席 %d）",
			m.players[m.declarerIdx].GetHcp(), m.dummyIdx), nil)
}

// nextSeat は次の席を返す。
func (m *Minibridge) nextSeat(i int) int { return (i + 1) % MinibridgePlayerCnt }

// PlayerSelectContract は人間が契約を選ぶ。
//
// **競りが無いので上回る必要はありません。** レベルと種別を 1 度だけ決めます。
func (m *Minibridge) PlayerSelectContract(level, suit int) error {
	if !m.IsHumanContractTurn() {
		return errors.New("not your turn to choose the contract")
	}
	return m.selectContractBy(m.declarerIdx, level, suit)
}

// CpuSelectContract は CPU の落札者が契約を選ぶ。
func (m *Minibridge) CpuSelectContract() {
	if m.gameEndFlag || m.phase != MinibridgePhaseContract || m.IsHumanContractTurn() {
		return
	}
	level, suit := m.chooseCpuContract(m.declarerIdx)
	_ = m.selectContractBy(m.declarerIdx, level, suit)
}

// selectContractBy は契約を確定してプレイフェーズに入る。
func (m *Minibridge) selectContractBy(playerIdx, level, suit int) error {
	if m.gameEndFlag {
		return ErrGameEnded
	}
	if m.phase != MinibridgePhaseContract {
		return ErrWrongPhase
	}
	if playerIdx != m.declarerIdx {
		return errors.New("only the declarer chooses the contract")
	}
	if level < 1 || level > MinibridgeMaxLevel {
		return fmt.Errorf("invalid contract level: %d", level)
	}
	// 0 はノートランプ。**「省略」ではなく 5 つ目の選択肢。**
	if suit < 0 || suit > CardDesignDiamond {
		return fmt.Errorf("invalid contract suit: %d", suit)
	}

	m.contractLevel, m.contractSuit = level, suit
	m.phase = MinibridgePhasePlay
	// **リードは落札者の左隣から。** ダミーは落札者の正面。
	m.leadPlayerIdx = m.nextSeat(m.declarerIdx)
	m.currentPlayerIdx = m.leadPlayerIdx
	m.addLog(playerIdx, "contract",
		fmt.Sprintf("契約 %d%s（必要 %d トリック）",
			level, minibridgeContractSuitStr(suit), m.RequiredTricks()), nil)
	return nil
}

// minibridgeContractSuitStr は契約の種別を短い表示にする。
func minibridgeContractSuitStr(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "S"
	case CardDesignClover:
		return "C"
	case CardDesignHeart:
		return "H"
	case CardDesignDiamond:
		return "D"
	default:
		return "NT"
	}
}

// RequiredTricks は契約に必要なトリック数を返す。**6 + レベル。**
func (m *Minibridge) RequiredTricks() int {
	if m.contractLevel == 0 {
		return 0
	}
	return MinibridgeBookTricks + m.contractLevel
}

// chooseCpuContract は CPU の落札者が選ぶ契約を返す。
//
// **自分とダミーの HCP は見えない前提で、自分の手だけで決めます。** 長いスートが
// あればそれを、無ければノートランプ。レベルはペアの HCP から素朴に見積もります。
func (m *Minibridge) chooseCpuContract(playerIdx int) (int, int) {
	p := m.players[playerIdx]
	counts := map[int]int{}
	for i := range p.GetCardsSize() {
		counts[p.GetCard(i).GetDesign()]++
	}
	bestSuit, bestLen := 0, 0
	// **スート番号の昇順で見る。** map の反復順は不定なので、そのまま回すと
	// 同じ手札で違う契約を返してテストが配りに依存する。
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestLen {
			bestSuit, bestLen = suit, counts[suit]
		}
	}
	if bestLen < minibridgeCpuSuitLen {
		bestSuit = 0
	}

	pairHcp := p.GetHcp() + m.players[m.dummyIdx].GetHcp()
	level := 1
	switch {
	case pairHcp >= minibridgeCpuSlamHcp:
		level = 6
	case pairHcp >= minibridgeCpuGameHcp:
		// game になる最小レベルを選ぶ。
		if bestSuit == 0 {
			level = 3
		} else if minibridgeIsMinor(bestSuit) {
			level = 5
		} else {
			level = 4
		}
	case pairHcp >= minibridgeCpuPartScoreHcp:
		level = 2
	}
	return level, bestSuit
}

// CPU の契約見積もりのしきい値。
const (
	// minibridgeCpuSuitLen はトランプに選ぶ最低枚数。
	minibridgeCpuSuitLen = 5
	// minibridgeCpuPartScoreHcp はレベル 2 を狙う下限。
	minibridgeCpuPartScoreHcp = 20
	// minibridgeCpuGameHcp は game を狙う下限。
	minibridgeCpuGameHcp = 25
	// minibridgeCpuSlamHcp はスラムを狙う下限。
	minibridgeCpuSlamHcp = 33
)

// GetValidPlayIndices はプレイ可能な手札インデックスを返す。**フォロー義務あり。**
func (m *Minibridge) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= MinibridgePlayerCnt {
		return nil
	}
	p := m.players[playerIdx]
	all := make([]int, 0, p.GetCardsSize())
	follow := make([]int, 0, p.GetCardsSize())
	lead := m.leadSuit()
	for i := range p.GetCardsSize() {
		all = append(all, i)
		if lead > 0 && p.GetCard(i).GetDesign() == lead {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// leadSuit はこのトリックのリードスートを返す（誰も出していなければ 0）。
func (m *Minibridge) leadSuit() int {
	if len(m.currentTrick) == 0 {
		return 0
	}
	return m.currentTrick[0].Card.GetDesign()
}

// IsHumanContractTurn は人間が契約を選ぶ番かを返す。
func (m *Minibridge) IsHumanContractTurn() bool {
	if m.gameEndFlag || m.phase != MinibridgePhaseContract || m.declarerIdx < 0 {
		return false
	}
	return m.players[m.declarerIdx].GetIsHuman()
}

// IsHumanTurn は人間が札を出す番かを返す。
//
// **ダミーの席かどうかを先に見る（レビュー指摘 PR #5313）。** ダミーは常に
// 落札者の相方なので、落札者が席 2 ならダミーは席 0 ——**人間自身の席**です。
// 「その席が人間か」を先に見ると、CPU が落札したダミーまで人間の番と判定され、
// `advance()` が CPU を進めないまま人間の入力待ちで止まります。
// ダミーの手番を握るのは席の持ち主ではなく**落札者**です。
func (m *Minibridge) IsHumanTurn() bool {
	if m.gameEndFlag || m.phase != MinibridgePhasePlay {
		return false
	}
	if m.currentPlayerIdx == m.dummyIdx && m.declarerIdx >= 0 {
		return m.players[m.declarerIdx].GetIsHuman()
	}
	return m.players[m.currentPlayerIdx].GetIsHuman()
}

// PlayerPlay は人間が札を出す。
func (m *Minibridge) PlayerPlay(cardIndex int) error {
	if !m.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return m.play(m.currentPlayerIdx, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (m *Minibridge) CpuPlay() {
	if m.gameEndFlag || m.phase != MinibridgePhasePlay || m.IsHumanTurn() {
		return
	}
	_ = m.play(m.currentPlayerIdx, m.chooseCpuCard(m.currentPlayerIdx))
}

// play は 1 枚出す共通処理。
func (m *Minibridge) play(playerIdx, cardIndex int) error {
	if m.gameEndFlag {
		return ErrGameEnded
	}
	if m.phase != MinibridgePhasePlay {
		return ErrWrongPhase
	}
	if playerIdx != m.currentPlayerIdx {
		return ErrNotHumanTurn
	}
	p := m.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if !minibridgeContains(m.GetValidPlayIndices(playerIdx), cardIndex) {
		return errors.New("must follow the led suit")
	}

	card := p.RemoveCard(cardIndex)
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードがありません")
	}
	m.currentTrick = append(m.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	m.addLog(playerIdx, "play", "カードを出しました", []*Card{card})

	if len(m.currentTrick) < MinibridgePlayerCnt {
		m.currentPlayerIdx = m.nextSeat(playerIdx)
		return nil
	}
	m.resolveTrick()
	return nil
}

// resolveTrick はトリックを解決する。
func (m *Minibridge) resolveTrick() {
	winner := m.trickWinner()
	cards := make([]*Card, 0, len(m.currentTrick))
	for _, tc := range m.currentTrick {
		cards = append(cards, tc.Card)
	}
	m.players[winner].AddTrick(cards)
	m.addLog(winner, "trick", "トリックを取りました", cards)

	m.currentTrick = nil
	m.trickNumber++
	m.leadPlayerIdx = winner
	m.currentPlayerIdx = winner

	if m.trickNumber >= MinibridgeTotalTricks {
		m.finishRound()
	}
}

// trickWinner はこのトリックの勝者を返す。
func (m *Minibridge) trickWinner() int {
	if len(m.currentTrick) == 0 {
		return m.leadPlayerIdx
	}
	best := m.currentTrick[0]
	for _, tc := range m.currentTrick[1:] {
		if m.beats(tc.Card, best.Card) {
			best = tc
		}
	}
	return best.PlayerIdx
}

// beats は challenger が champion に勝つかを返す。
//
// **ノートランプ契約では contractSuit が 0 で、どの札の design とも一致しません**
// （CardDesign は 1 から始まる）。切り札の分岐がそのまま無効化されます。
func (m *Minibridge) beats(challenger, champion *Card) bool {
	cs, hs := challenger.GetDesign(), champion.GetDesign()
	switch {
	case cs == hs:
		return minibridgeRank(challenger) > minibridgeRank(champion)
	case cs == m.contractSuit:
		return true
	case hs == m.contractSuit:
		return false
	default:
		// champion は必ずリードスートか切り札なので、ここに来る challenger は負け。
		return false
	}
}

// minibridgeContains は xs が v を含むかを返す。
func minibridgeContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// chooseCpuCard は CPU の手。**取れるなら取り、取れないなら安く出す。**
//
// **どれも勝てないときに最強札を投げない（レビュー指摘 PR #5313）。** リードの
// スートを持っていないと `GetValidPlayIndices` は手札全部を返すので、素朴に
// 最強を選ぶと切り札でもないエースを捨て札にしてしまいます。
func (m *Minibridge) chooseCpuCard(playerIdx int) int {
	valid := m.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := m.players[playerIdx]

	// いま場の最強札に勝てる手があるかを先に見る。
	canWin := false
	for _, i := range valid {
		if m.winsTrick(p.GetCard(i)) {
			canWin = true
			break
		}
	}

	pick := valid[0]
	pickRank := minibridgeRank(p.GetCard(pick))
	for _, i := range valid[1:] {
		r := minibridgeRank(p.GetCard(i))
		switch {
		case canWin && m.winsTrick(p.GetCard(i)) && (!m.winsTrick(p.GetCard(pick)) || r > pickRank):
			pick, pickRank = i, r
		case !canWin && r < pickRank:
			pick, pickRank = i, r
		}
	}
	return pick
}

// winsTrick は card がいまの場を取れるかを返す（場が空ならリードなので取れる扱い）。
func (m *Minibridge) winsTrick(card *Card) bool {
	if len(m.currentTrick) == 0 {
		return true
	}
	best := m.currentTrick[0].Card
	for _, tc := range m.currentTrick[1:] {
		if m.beats(tc.Card, best) {
			best = tc.Card
		}
	}
	return m.beats(card, best)
}

// finishRound はディールを精算する。
func (m *Minibridge) finishRound() {
	m.phase = MinibridgePhaseRoundEnd
	declTeam := m.players[m.declarerIdx].GetTeam()
	took := 0
	for _, p := range m.players {
		if p.GetTeam() == declTeam {
			took += p.GetTrickCount()
		}
	}
	need := m.RequiredTricks()
	m.lastTricks = took

	if took >= need {
		points := m.contractScore(took - need)
		m.teamScores[declTeam] += points
		m.lastMade = true
		m.addLog(m.declarerIdx, "score",
			fmt.Sprintf("契約 %d に対し %d トリック：成立 (+%d)", need, took, points), nil)
	} else {
		points := (need - took) * minibridgeUndertrickPenalty
		m.teamScores[1-declTeam] += points
		m.lastMade = false
		m.addLog(m.declarerIdx, "score",
			fmt.Sprintf("契約 %d に対し %d トリック：失敗、相手に +%d", need, took, points), nil)
	}

	if m.roundNumber >= m.config.Rounds {
		m.finishGame()
	}
}

// contractScore は成立時の得点を返す。
//
// **契約点で game / partscore が分かれます。** 契約点 100 が境目で、
// ちょうど 3NT・4♥/♠・5♣/♦ が game になります——実際のブリッジの game 契約と
// 一致するので、この表が正しいことの傍証になります。
func (m *Minibridge) contractScore(overtricks int) int {
	contract := m.contractPoints()
	score := contract
	if contract >= MinibridgeGameThreshold {
		score += minibridgeGameBonus
	} else {
		score += minibridgePartScoreBonus
	}
	switch m.contractLevel {
	case 6:
		score += minibridgeSmallSlamBonus
	case MinibridgeMaxLevel:
		score += minibridgeGrandSlamBonus
	}
	// **オーバートリックは 1 トリックあたりの契約点と同じ率。**
	score += overtricks * m.perTrickPoints()
	return score
}

// contractPoints は契約そのものの点を返す（ボーナス抜き）。
func (m *Minibridge) contractPoints() int {
	if m.contractLevel == 0 {
		return 0
	}
	if m.contractSuit == 0 {
		return minibridgeNoTrumpFirst + minibridgeNoTrumpRest*(m.contractLevel-1)
	}
	return m.perTrickPoints() * m.contractLevel
}

// perTrickPoints は 1 トリックあたりの契約点を返す。
func (m *Minibridge) perTrickPoints() int {
	switch {
	case m.contractSuit == 0:
		return minibridgeNoTrumpRest
	case minibridgeIsMinor(m.contractSuit):
		return minibridgeMinorPerTrick
	default:
		return minibridgeMajorPerTrick
	}
}

// NextRound は次のディールを開始する。
func (m *Minibridge) NextRound() {
	if m.gameEndFlag || m.phase != MinibridgePhaseRoundEnd {
		return
	}
	m.roundNumber++
	m.dealerIdx = m.nextSeat(m.dealerIdx)
	m.startRound()
}

// finishGame は終局処理。
func (m *Minibridge) finishGame() {
	m.phase = MinibridgePhaseGameEnd
	m.gameEndFlag = true
	switch {
	case m.teamScores[0] > m.teamScores[1]:
		m.winnerTeam = 0
	case m.teamScores[1] > m.teamScores[0]:
		m.winnerTeam = 1
	default:
		m.winnerTeam = -1
	}
	m.addLog(-1, "result",
		fmt.Sprintf("最終得点 %d - %d", m.teamScores[0], m.teamScores[1]), nil)
}

// GiveUp は投了する。
func (m *Minibridge) GiveUp() {
	if m.gameEndFlag {
		return
	}
	m.phase = MinibridgePhaseGameEnd
	m.gameEndFlag = true
	// 人間は席 0 = チーム 0。
	m.winnerTeam = 1
	m.addLog(0, "giveup", "投了しました", nil)
}

// GetHint は人間への助言を返す。
func (m *Minibridge) GetHint() *MinibridgeHint {
	if m.gameEndFlag {
		return nil
	}
	if m.IsHumanContractTurn() {
		level, suit := m.chooseCpuContract(m.declarerIdx)
		return &MinibridgeHint{Reason: "minibridgeContract", Level: level, Suit: suit}
	}
	if !m.IsHumanTurn() {
		return nil
	}
	valid := m.GetValidPlayIndices(m.currentPlayerIdx)
	if len(valid) == 0 {
		return nil
	}
	idx := m.chooseCpuCard(m.currentPlayerIdx)
	reason := "minibridgeWinTrick"
	if m.currentPlayerIdx == m.dummyIdx {
		// **ダミーの手番も人間が操作します。** 何を動かしているかを言い分ける。
		reason = "minibridgeDummy"
	}
	return &MinibridgeHint{CardIndex: &idx, Reason: reason}
}

// addLog は棋譜に 1 行足す。
func (m *Minibridge) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	m.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (m *Minibridge) GetConfig() MinibridgeConfig { return m.config }

// SetConfig はゲーム設定を設定する。
func (m *Minibridge) SetConfig(cfg MinibridgeConfig) { m.config = cfg }

// GetPhase は現在のフェーズを返す。
func (m *Minibridge) GetPhase() MinibridgePhase { return m.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (m *Minibridge) GetGameEndFlag() bool { return m.gameEndFlag }

// GetRoundNumber は現在のディール番号を返す。
func (m *Minibridge) GetRoundNumber() int { return m.roundNumber }

// GetTrickNumber は現在のトリック番号を返す。
func (m *Minibridge) GetTrickNumber() int { return m.trickNumber }

// GetContractLevel は契約レベルを返す（0: 未決定）。
func (m *Minibridge) GetContractLevel() int { return m.contractLevel }

// GetContractSuit は契約の種別を返す（0: ノートランプ）。
func (m *Minibridge) GetContractSuit() int { return m.contractSuit }

// GetDeclarerIdx は落札者を返す（-1: 未定）。
func (m *Minibridge) GetDeclarerIdx() int { return m.declarerIdx }

// GetDummyIdx はダミーを返す（-1: 未定）。
func (m *Minibridge) GetDummyIdx() int { return m.dummyIdx }

// GetDummyHand はダミーの手札を返す。**契約が決まってから公開されます。**
func (m *Minibridge) GetDummyHand() []*Card {
	if m.dummyIdx < 0 || m.phase == MinibridgePhaseContract {
		return nil
	}
	p := m.players[m.dummyIdx]
	out := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		out = append(out, p.GetCard(i))
	}
	return out
}

// GetCurrentPlayerIdx は現在の手番を返す。
func (m *Minibridge) GetCurrentPlayerIdx() int { return m.currentPlayerIdx }

// GetLeadPlayerIdx はリードプレイヤーを返す。
func (m *Minibridge) GetLeadPlayerIdx() int { return m.leadPlayerIdx }

// GetDealerIdx は親を返す。
func (m *Minibridge) GetDealerIdx() int { return m.dealerIdx }

// GetCurrentTrick は現在のトリックを返す。
func (m *Minibridge) GetCurrentTrick() []*TrickCard { return m.currentTrick }

// GetLastMade は直前のディールで契約が成立したかを返す。
func (m *Minibridge) GetLastMade() bool { return m.lastMade }

// GetLastTricks は直前のディールで宣言側が取ったトリック数を返す。
func (m *Minibridge) GetLastTricks() int { return m.lastTricks }

// GetTeamScore はチームの累計得点を返す。
func (m *Minibridge) GetTeamScore(team int) int {
	if team < 0 || team >= MinibridgeTeamCnt {
		return 0
	}
	return m.teamScores[team]
}

// SetTeamScore はチームの累計得点を設定する（復元・テスト用）。
func (m *Minibridge) SetTeamScore(team, score int) {
	if team < 0 || team >= MinibridgeTeamCnt {
		return
	}
	m.teamScores[team] = score
}

// GetPlayerCnt はプレイヤー数を返す。
func (m *Minibridge) GetPlayerCnt() int { return MinibridgePlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (m *Minibridge) GetPlayer(i int) *MinibridgePlayer {
	if i < 0 || i >= len(m.players) {
		return nil
	}
	return m.players[i]
}

// GetWinnerTeam は勝ったチームを返す（-1: 未確定/同点）。
func (m *Minibridge) GetWinnerTeam() int { return m.winnerTeam }

// minibridgeJSON は KV スナップショットの表現。
type minibridgeJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*MinibridgePlayer    `json:"pl"`
	Config           MinibridgeConfig       `json:"cf"`
	Phase            MinibridgePhase        `json:"ph"`
	RoundNumber      int                    `json:"rn"`
	TrickNumber      int                    `json:"tn"`
	ContractLevel    int                    `json:"cl"`
	ContractSuit     int                    `json:"cs"`
	DeclarerIdx      int                    `json:"di"`
	DummyIdx         int                    `json:"dm"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	CurrentPlayerIdx int                    `json:"ci"`
	LeadPlayerIdx    int                    `json:"li"`
	DealerIdx        int                    `json:"dl"`
	TeamScores       [MinibridgeTeamCnt]int `json:"ts"`
	LastMade         bool                   `json:"lm"`
	LastTricks       int                    `json:"lt"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerTeam       int                    `json:"wt"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (m *Minibridge) MarshalJSON() ([]byte, error) {
	return json.Marshal(&minibridgeJSON{
		TrumpCards: m.trumpCards, Players: m.players, Config: m.config, Phase: m.phase,
		RoundNumber: m.roundNumber, TrickNumber: m.trickNumber,
		ContractLevel: m.contractLevel, ContractSuit: m.contractSuit,
		DeclarerIdx: m.declarerIdx, DummyIdx: m.dummyIdx,
		CurrentTrick: m.currentTrick, CurrentPlayerIdx: m.currentPlayerIdx,
		LeadPlayerIdx: m.leadPlayerIdx, DealerIdx: m.dealerIdx,
		TeamScores: m.teamScores, LastMade: m.lastMade, LastTricks: m.lastTricks,
		GameEndFlag: m.gameEndFlag, WinnerTeam: m.winnerTeam, ActionLog: m.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
//
// **6 PR 連続で「個々のフィールドは範囲内だが組み合わせがあり得ない」を通していた**
// ので、ここはフェーズ × 各フィールドの表として書いています (#5302〜#5312)。
func (m *Minibridge) UnmarshalJSON(data []byte) error {
	var j minibridgeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < MinibridgePhaseContract || j.Phase > MinibridgePhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= MinibridgePlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	// **落札者とダミーは対で決まる。** 片方だけ立っている状態は無い。
	if j.DeclarerIdx < -1 || j.DeclarerIdx >= MinibridgePlayerCnt {
		return fmt.Errorf("invalid declarer: %d", j.DeclarerIdx)
	}
	if j.DummyIdx < -1 || j.DummyIdx >= MinibridgePlayerCnt {
		return fmt.Errorf("invalid dummy: %d", j.DummyIdx)
	}
	if (j.DeclarerIdx < 0) != (j.DummyIdx < 0) {
		return fmt.Errorf("declarer %d and dummy %d disagree", j.DeclarerIdx, j.DummyIdx)
	}
	if j.DeclarerIdx >= 0 && j.DummyIdx != (j.DeclarerIdx+MinibridgeTeamCnt)%MinibridgePlayerCnt {
		return fmt.Errorf("dummy %d is not the declarer's partner", j.DummyIdx)
	}
	// **配ったあとは必ず落札者がいる。** 契約フェーズでも決まっています——
	// 決まっていないと「誰が契約を選ぶのか」が無く、advance が止まらなくなる。
	if j.DeclarerIdx < 0 {
		return errors.New("no declarer")
	}
	if j.ContractLevel < 0 || j.ContractLevel > MinibridgeMaxLevel {
		return fmt.Errorf("invalid contract level: %d", j.ContractLevel)
	}
	if j.ContractSuit < 0 || j.ContractSuit > CardDesignDiamond {
		return fmt.Errorf("invalid contract suit: %d", j.ContractSuit)
	}
	if j.ContractLevel == 0 && j.ContractSuit != 0 {
		return fmt.Errorf("contract suit %d without a level", j.ContractSuit)
	}
	// **契約フェーズは未決定、それ以降は決定済み。** #5312 で踏んだ形。
	if j.Phase == MinibridgePhaseContract && j.ContractLevel != 0 {
		return fmt.Errorf("contract %d already chosen during the contract phase", j.ContractLevel)
	}
	if j.Phase == MinibridgePhasePlay && j.ContractLevel == 0 {
		return errors.New("play phase without a contract")
	}
	if j.RoundNumber < 1 || j.RoundNumber > j.Config.Rounds {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > MinibridgeTotalTricks {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.LastTricks < 0 || j.LastTricks > MinibridgeTotalTricks {
		return fmt.Errorf("invalid last tricks: %d", j.LastTricks)
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= MinibridgeTeamCnt {
		return fmt.Errorf("invalid winner team: %d", j.WinnerTeam)
	}
	if !j.GameEndFlag && j.WinnerTeam != -1 {
		return fmt.Errorf("winner team %d before the game ended", j.WinnerTeam)
	}
	// **終了フラグとフェーズは対（レビュー指摘 PR #5313）。** `finishGame` は必ず
	// 両方を立てるので、片方だけ立った盤面は存在しません。通してしまうと
	// すべての入口が `gameEndFlag` で早期 return する一方フェーズは進まず、
	// **投了でも復旧できない恒久デッドロック**になります。
	if j.GameEndFlag != (j.Phase == MinibridgePhaseGameEnd) {
		return fmt.Errorf("game end flag %v disagrees with phase %d", j.GameEndFlag, j.Phase)
	}
	if len(j.CurrentTrick) > MinibridgePlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	// **枚数だけでなく中身も見る (#5310 で踏んだ panic の再発防止)。**
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= MinibridgePlayerCnt {
			return errors.New("invalid current trick entry")
		}
	}
	if len(j.ActionLog) > minibridgeMaxSliceLen {
		return errors.New("minibridge: input array exceeds maximum allowed size")
	}
	// **HCP の総和は必ず 40。** 席ごとの範囲検査だけでは通ってしまう組み合わせを
	// ここで弾きます——崩れていると落札者の決定そのものが別物になる。
	if len(j.Players) == MinibridgePlayerCnt {
		total := 0
		for _, p := range j.Players {
			if p == nil {
				return errors.New("nil player")
			}
			total += p.GetHcp()
		}
		if total != MinibridgeTotalHcp {
			return fmt.Errorf("hcp total is %d, want %d", total, MinibridgeTotalHcp)
		}
		m.players = j.Players
	}

	if j.TrumpCards != nil {
		m.trumpCards = j.TrumpCards
	}
	m.config, m.phase = j.Config, j.Phase
	m.roundNumber, m.trickNumber = j.RoundNumber, j.TrickNumber
	m.contractLevel, m.contractSuit = j.ContractLevel, j.ContractSuit
	m.declarerIdx, m.dummyIdx = j.DeclarerIdx, j.DummyIdx
	m.currentTrick, m.currentPlayerIdx = j.CurrentTrick, j.CurrentPlayerIdx
	m.leadPlayerIdx, m.dealerIdx = j.LeadPlayerIdx, j.DealerIdx
	m.teamScores, m.lastMade, m.lastTricks = j.TeamScores, j.LastMade, j.LastTricks
	m.gameEndFlag, m.winnerTeam, m.actionLog = j.GameEndFlag, j.WinnerTeam, j.ActionLog
	return nil
}
