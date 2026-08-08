//go:build !js || !wasm || casino

// Package domain トゥーテ (Tute) のドメインモデル。
//
// Tute はスペインで広くプレイされる 40 枚ラテンデッキの得点トリックテイキング
// ゲーム。切り札あり・マストフォローで 4 人 2 チーム (席 0&2 vs 1&3) が 10 トリックを
// 戦い、同スートの K と Q を揃えると「結婚宣言 (cante)」(非切り札=20点, 切り札=40点)
// ができる。4 枚の K または Q を揃えると「Tute」で即勝利。最終トリックに +10 点。
// 累積点が目標 (既定 121) に達したチームが勝利。
//
// カードの強さ (トリック): A > 3 > K > Q > J > 7 > 6 > 5 > 4 > 2
// カードポイント: A=11, 3=10, K=4, Q=3, J=2, それ以外=0 (1スート30点 × 4 = 120点)
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// TutePlayerCnt プレイヤー数 (人間 1 + CPU 3)
const TutePlayerCnt = 4

// TuteTeamCnt チーム数
const TuteTeamCnt = 2

// TuteHandSize 各プレイヤーの手札枚数 (40 / 4)
const TuteHandSize = 10

// TuteTrickCount 1 ラウンドのトリック数
const TuteTrickCount = 10

// TuteLastTrickBonus 最終トリック勝者へのボーナス点 (10 de últimas)
const TuteLastTrickBonus = 10

// TuteMarriagePlain 非切り札の結婚宣言点
const TuteMarriagePlain = 20

// TuteMarriageTrump 切り札の結婚宣言点
const TuteMarriageTrump = 40

// TutePhase ゲームフェーズ
type TutePhase int

// Tute のフェーズ定数
const (
	// TutePhasePlay トリックプレイフェーズ
	TutePhasePlay TutePhase = 0
	// TutePhaseTrickEnd トリック終了フェーズ (解決済み・次トリック待ち)
	TutePhaseTrickEnd TutePhase = 1
	// TutePhaseRoundEnd ラウンド終了フェーズ
	TutePhaseRoundEnd TutePhase = 2
	// TutePhaseGameEnd ゲーム終了フェーズ
	TutePhaseGameEnd TutePhase = 3
)

// TuteHint ヒント情報
type TuteHint struct {
	CardIndices []int  // 推奨カードインデックス
	Marriage    int    // 推奨結婚宣言スート (0=なし)
	Reason      string // ヒント理由キー
}

// Tute トゥーテのゲームクラス
type Tute struct {
	trumpCards       *TrumpCards
	players          []*TutePlayer
	config           TuteConfig
	phase            TutePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int                     // 切り札スート
	declaredSuits    [CardDesignMax + 1]bool // 結婚宣言済みスート
	teamScores       [TuteTeamCnt]int        // 累積点
	roundTeamPts     [TuteTeamCnt]int        // 現ラウンドの得点 (カード+結婚+最終)
	gameEndFlag      bool
	winnerTeam       int // -1=未確定
	actionLogBase
}

// NewTute コンストラクタ
func NewTute(trumpCards *TrumpCards, players []*TutePlayer, config TuteConfig) *Tute {
	return &Tute{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1}
}

// NewDefaultTute 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultTute() *Tute {
	players := make([]*TutePlayer, TutePlayerCnt)
	players[0] = NewTutePlayer(true)
	for i := 1; i < TutePlayerCnt; i++ {
		players[i] = NewTutePlayer(false)
	}
	return NewTute(NewTrumpCardsBriscola(), players, DefaultTuteConfig())
}

// TuteTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func TuteTeamOf(playerIdx int) int { return playerIdx % TuteTeamCnt }

// Reset ゲーム初期化
func (g *Tute) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [TuteTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Tute) NextRound() {
	if g.phase != TutePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % TutePlayerCnt
	g.startRound()
}

// startRound 手札を配り、切り札を決めてプレイフェーズを開始する。
func (g *Tute) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.declaredSuits = [CardDesignMax + 1]bool{}
	g.roundTeamPts = [TuteTeamCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.leadPlayerIdx = (g.dealerIdx + 1) % TutePlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = TutePhasePlay
}

// deal 各プレイヤーへ 10 枚を配り、最後に配った札のスートを切り札とする。
func (g *Tute) deal() {
	var last *Card
	for i := 0; i < TuteHandSize; i++ {
		for j := 0; j < TutePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % TutePlayerCnt
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
func (g *Tute) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TutePhasePlay {
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

// PlayerDeclareMarriage 人間がリード時に同スートの K+Q で結婚宣言する。
func (g *Tute) PlayerDeclareMarriage(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TutePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.canDeclareMarriage(g.currentPlayerIdx, suit) {
		return NewDomainError(ErrInvalidPlay, "そのスートは結婚宣言できません")
	}
	g.applyMarriage(g.currentPlayerIdx, suit)
	return nil
}

// PlayerDeclareTute 人間が 4 枚の K または Q を揃えて Tute を宣言する（即勝利）。
func (g *Tute) PlayerDeclareTute() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TutePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.hasTute(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, "Tute を宣言できません")
	}
	g.applyTute(g.currentPlayerIdx)
	return nil
}

// canDeclareMarriage リード時、未宣言スートの K+Q を持っているか。
func (g *Tute) canDeclareMarriage(playerIdx, suit int) bool {
	if len(g.currentTrick) != 0 || g.currentPlayerIdx != playerIdx {
		return false
	}
	if suit < 1 || suit > CardDesignMax || g.declaredSuits[suit] {
		return false
	}
	return g.hasCard(playerIdx, suit, 13) && g.hasCard(playerIdx, suit, 12)
}

// applyMarriage 結婚宣言を反映し、得点を加算する。
func (g *Tute) applyMarriage(playerIdx, suit int) {
	g.declaredSuits[suit] = true
	pts := TuteMarriagePlain
	if suit == g.trumpSuit {
		pts = TuteMarriageTrump
	}
	team := TuteTeamOf(playerIdx)
	g.roundTeamPts[team] += pts
	g.appendLog(playerIdx, "marriage",
		fmt.Sprintf("%s declares a %s marriage (+%d)", g.playerName(playerIdx), suitStr(suit), pts), nil)
}

// hasTute プレイヤーが 4 枚の K または 4 枚の Q を持つか。
func (g *Tute) hasTute(playerIdx int) bool {
	return g.countRank(playerIdx, 13) == 4 || g.countRank(playerIdx, 12) == 4
}

// applyTute Tute 宣言で即時にそのチームを勝者とする。
func (g *Tute) applyTute(playerIdx int) {
	team := TuteTeamOf(playerIdx)
	g.gameEndFlag = true
	g.winnerTeam = team
	g.phase = TutePhaseGameEnd
	g.appendLog(playerIdx, "tute",
		fmt.Sprintf("%s declares TUTE! Team %s wins the game!", g.playerName(playerIdx), teamName(team)), nil)
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *Tute) CpuPlay() {
	if g.gameEndFlag || g.phase != TutePhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	// リード時: Tute → 結婚宣言を検討する。
	if len(g.currentTrick) == 0 {
		if g.hasTute(idx) {
			g.applyTute(idx)
			return
		}
		if suit := g.cpuMarriageSuit(idx); suit > 0 {
			g.applyMarriage(idx, suit)
		}
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
func (g *Tute) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == TutePlayerCnt {
		g.phase = TutePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TutePlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Tute) ResolveTrick() {
	if g.phase != TutePhaseTrickEnd || len(g.currentTrick) != TutePlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += tuteCardPoints(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)
	team := TuteTeamOf(winnerIdx)
	g.roundTeamPts[team] += pts
	bonus := ""
	if g.trickNumber >= TuteTrickCount {
		g.roundTeamPts[team] += TuteLastTrickBonus
		bonus = fmt.Sprintf(" +%d last", TuteLastTrickBonus)
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d%s)", g.playerName(winnerIdx), g.trickNumber, pts, bonus), trickCards)

	g.leadPlayerIdx = winnerIdx
	// Keep currentTrick intact through TrickEnd so the resolved trick stays
	// visible; NextTrick clears it before the next trick begins. (#2482 review)
	if g.trickNumber >= TuteTrickCount {
		g.phase = TutePhaseRoundEnd
	} else {
		g.phase = TutePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Tute) NextTrick() {
	if g.phase != TutePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = TutePhasePlay
}

// ScoreRound ラウンド得点を累積し、ゲーム終了判定を行う。
func (g *Tute) ScoreRound() {
	if g.phase != TutePhaseRoundEnd {
		return
	}
	for t := 0; t < TuteTeamCnt; t++ {
		g.teamScores[t] += g.roundTeamPts[t]
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: Team A=%d (+%d), Team B=%d (+%d)",
			g.roundNumber, g.teamScores[0], g.roundTeamPts[0], g.teamScores[1], g.roundTeamPts[1]), nil)

	leader, other := 0, 1
	if g.teamScores[1] > g.teamScores[0] {
		leader, other = 1, 0
	}
	if g.teamScores[leader] >= g.config.TargetPoints && g.teamScores[leader] > g.teamScores[other] {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = TutePhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the game!", teamName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォロー (リードスートに従う) を検証する。
func (g *Tute) validatePlay(playerIdx int, card *Card) error {
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
func (g *Tute) playerHasSuit(playerIdx, design int) bool {
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
func (g *Tute) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.tuteRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		// 勝負に絡むのは切り札またはリードスートの札のみ。
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.tuteRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// tuteRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Tute) tuteRank(card *Card) int {
	r := tuteStrength(card.GetValue())
	if card.GetDesign() == g.trumpSuit {
		r += 100
	}
	return r
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Tute) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Card helpers ---

// hasCard プレイヤーが指定スート・値のカードを持つか。
func (g *Tute) hasCard(playerIdx, design, value int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && c.GetValue() == value {
			return true
		}
	}
	return false
}

// countRank プレイヤーが持つ指定値のカード枚数。
func (g *Tute) countRank(playerIdx, value int) int {
	p := g.players[playerIdx]
	cnt := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == value {
			cnt++
		}
	}
	return cnt
}

// tuteStrength トリックの強さ。A>3>K>Q>J>7>6>5>4>2。
func tuteStrength(value int) int {
	switch value {
	case 1: // As
		return 10
	case 3:
		return 9
	case 13: // Rey
		return 8
	case 12: // Caballo(Q)
		return 7
	case 11: // Sota(J)
		return 6
	case 7:
		return 5
	case 6:
		return 4
	case 5:
		return 3
	case 4:
		return 2
	default: // 2
		return 1
	}
}

// tuteCardPoints カードポイント。A=11,3=10,K=4,Q=3,J=2,その他=0。
func tuteCardPoints(value int) int {
	switch value {
	case 1:
		return 11
	case 3:
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

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Tute) sortAllHands() {
	for _, p := range g.players {
		tuteSortHand(p)
	}
}

// tuteSortHand 手札をスート→強さ順にソートする。
func tuteSortHand(p *TutePlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return tuteStrength(cards[i].GetValue()) > tuteStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Tute) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Tute) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者 winnerIdx の札のランクを返す。見つからない場合は極小値。
func (g *Tute) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.tuteRank(g.currentTrick[idx].Card)
}

// --- CPU AI ---

// cpuMarriageSuit CPU がリード時に宣言する結婚スートを返す (0=なし)。切り札を優先。
func (g *Tute) cpuMarriageSuit(playerIdx int) int {
	best := 0
	for suit := 1; suit <= CardDesignMax; suit++ {
		if g.canDeclareMarriage(playerIdx, suit) {
			if suit == g.trumpSuit {
				return suit
			}
			best = suit
		}
	}
	return best
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Tute) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == TuteCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *Tute) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		// リード: 得点・強さの低い札を温存。
		return pickLowest(player, valid, func(c *Card) int {
			return tuteCardPoints(c.GetValue())*100 + tuteStrength(c.GetValue())
		})
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	partnerWinning := TuteTeamOf(winnerIdx) == TuteTeamOf(playerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += tuteCardPoints(tc.Card.GetValue())
	}

	var follows []int
	for _, idx := range valid {
		if player.GetCard(idx).GetDesign() == leadSuit {
			follows = append(follows, idx)
		}
	}
	if len(follows) == 0 {
		if partnerWinning {
			// 味方が勝っている: 得点札を渡す (切り札は温存)。
			return pickHighest(player, valid, func(c *Card) int {
				if c.GetDesign() == g.trumpSuit {
					return -tuteStrength(c.GetValue())
				}
				return tuteCardPoints(c.GetValue())*100 - tuteStrength(c.GetValue())
			})
		}
		return pickLowest(player, valid, func(c *Card) int {
			return tuteCardPoints(c.GetValue())*100 + tuteStrength(c.GetValue())
		})
	}
	winners := tuteFilter(follows, func(idx int) bool { return g.tuteRank(player.GetCard(idx)) > topRank })
	if partnerWinning {
		nonWinners := tuteFilter(follows, func(idx int) bool { return g.tuteRank(player.GetCard(idx)) < topRank })
		if len(nonWinners) > 0 {
			return pickHighest(player, nonWinners, func(c *Card) int {
				return tuteCardPoints(c.GetValue())*100 - tuteStrength(c.GetValue())
			})
		}
		return pickLowest(player, follows, func(c *Card) int { return tuteStrength(c.GetValue()) })
	}
	if trickPts > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return tuteStrength(c.GetValue()) })
	}
	return pickLowest(player, follows, func(c *Card) int {
		return tuteCardPoints(c.GetValue())*100 + tuteStrength(c.GetValue())
	})
}

// tuteFilter 述語を満たすインデックスを抽出する。
func tuteFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Tute) GetHint() *TuteHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != TutePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	// リード時の結婚宣言を提案。
	if len(g.currentTrick) == 0 {
		if g.hasTute(human) {
			return &TuteHint{Reason: "declare_tute"}
		}
		if suit := g.cpuMarriageSuit(human); suit > 0 {
			return &TuteHint{Marriage: suit, Reason: "declare_marriage"}
		}
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &TuteHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Tute) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.tuteRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Tute) GetPhase() TutePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Tute) SetPhase(phase TutePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Tute) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Tute) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Tute) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Tute) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Tute) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Tute) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Tute) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Tute) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Tute) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Tute) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Tute) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit 切り札スート取得
func (g *Tute) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Tute) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// IsSuitDeclared 指定スートが結婚宣言済みか
func (g *Tute) IsSuitDeclared(suit int) bool {
	if suit < 0 || suit > CardDesignMax {
		return false
	}
	return g.declaredSuits[suit]
}

// GetTeamScores チーム別累積点取得
func (g *Tute) GetTeamScores() [TuteTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点設定 (テスト用)
func (g *Tute) SetTeamScores(s [TuteTeamCnt]int) { g.teamScores = s }

// GetRoundTeamPoints 現ラウンドのチーム別得点取得
func (g *Tute) GetRoundTeamPoints() [TuteTeamCnt]int { return g.roundTeamPts }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Tute) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *Tute) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Tute) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Tute) GetPlayer(i int) *TutePlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か。
func (g *Tute) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// CanHumanDeclareMarriage 人間が今いずれかのスートで結婚宣言できるか。
func (g *Tute) CanHumanDeclareMarriage() bool {
	human := findHumanIdx(g.players)
	if human < 0 || g.currentPlayerIdx != human {
		return false
	}
	return g.cpuMarriageSuit(human) > 0
}

// GetHumanDeclarableMarriageSuits returns the suits (1..CardDesignMax) for which
// the human may currently declare a K+Q marriage — the human leads and holds an
// unclaimed suit's King and Queen. Empty when no declaration is possible now.
func (g *Tute) GetHumanDeclarableMarriageSuits() []int {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	var suits []int
	for suit := 1; suit <= CardDesignMax; suit++ {
		if g.canDeclareMarriage(human, suit) {
			suits = append(suits, suit)
		}
	}
	return suits
}

// CanHumanDeclareTute 人間が今 Tute を宣言できるか。
func (g *Tute) CanHumanDeclareTute() bool {
	human := findHumanIdx(g.players)
	return human >= 0 && g.currentPlayerIdx == human && len(g.currentTrick) == 0 && g.hasTute(human)
}

// GetConfig 設定取得
func (g *Tute) GetConfig() TuteConfig { return g.config }

// SetConfig 設定変更
func (g *Tute) SetConfig(cfg TuteConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Tute) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != TutePhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// tuteJSON is the JSON wire format for Tute.
type tuteJSON struct {
	TrumpCards       *TrumpCards             `json:"tc"`
	Players          []*TutePlayer           `json:"ps"`
	Config           TuteConfig              `json:"cf"`
	Phase            TutePhase               `json:"ph"`
	RoundNumber      int                     `json:"rn"`
	TrickNumber      int                     `json:"tn"`
	CurrentPlayerIdx int                     `json:"ci"`
	CurrentTrick     []*TrickCard            `json:"ct"`
	LeadPlayerIdx    int                     `json:"li"`
	DealerIdx        int                     `json:"di"`
	TrumpSuit        int                     `json:"ts"`
	DeclaredSuits    [CardDesignMax + 1]bool `json:"ds"`
	TeamScores       [TuteTeamCnt]int        `json:"sc"`
	RoundTeamPts     [TuteTeamCnt]int        `json:"rp"`
	GameEndFlag      bool                    `json:"ge"`
	WinnerTeam       int                     `json:"wt"`
	ActionLog        []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Tute) MarshalJSON() ([]byte, error) {
	return json.Marshal(tuteJSON{
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
		DeclaredSuits:    g.declaredSuits,
		TeamScores:       g.teamScores,
		RoundTeamPts:     g.roundTeamPts,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// tuteMaxSliceLen caps slice sizes during deserialisation.
const tuteMaxSliceLen = 5000

// errTuteOversized is the single sentinel error for oversized input arrays.
var errTuteOversized = errors.New("tute: input array exceeds maximum allowed size")

// errTuteInvalidPlayers is returned when restored state lacks exactly TutePlayerCnt players.
var errTuteInvalidPlayers = errors.New("tute: invalid player count")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Tute) UnmarshalJSON(data []byte) error {
	var j tuteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tuteMaxSliceLen || len(j.CurrentTrick) > tuteMaxSliceLen ||
		len(j.ActionLog) > tuteMaxSliceLen {
		return errTuteOversized
	}
	if len(j.Players) != TutePlayerCnt {
		return errTuteInvalidPlayers
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
	g.declaredSuits = j.DeclaredSuits
	g.teamScores = j.TeamScores
	g.roundTeamPts = j.RoundTeamPts
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
