//go:build !js || !wasm || classic

// Package domain クラヴァヤス (Klaverjas) のドメインモデル。
//
// Klaverjas はオランダの Jass 系トリックテイキングゲーム。32 枚デッキで 4 人 2 チーム
// (席 0&2 vs 1&3) が 8 トリックを戦う。切り札スートでは J(ジャス=20点) と 9(ネル=14点)
// が最強となる独特のランク体系を持ち、マストフォロー＋切り札強制追い越し
// (overtrump) の義務がある。配り手の最初の手札に含まれる連続札 (Roem) や 4 枚組で
// 追加点を得る。最終トリックに +10 点。累積点が目標 (既定 1501) に達したチームが勝利。
//
// 切り札の強さ: J > 9 > A > 10 > K > Q > 8 > 7
// 非切り札の強さ: A > 10 > K > Q > J > 9 > 8 > 7
// 切り札ポイント: J=20, 9=14, A=11, 10=10, K=4, Q=3, 8/7=0
// 非切り札ポイント: A=11, 10=10, K=4, Q=3, J=2, 9/8/7=0 (合計 152点 + 最終 10点)
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// KlaverjasPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const KlaverjasPlayerCnt = 4

// KlaverjasTeamCnt チーム数
const KlaverjasTeamCnt = 2

// KlaverjasHandSize 各プレイヤーの手札枚数 (32 / 4)
const KlaverjasHandSize = 8

// KlaverjasTrickCount 1 ラウンドのトリック数
const KlaverjasTrickCount = 8

// KlaverjasLastTrickBonus 最終トリック勝者へのボーナス点
const KlaverjasLastTrickBonus = 10

// KlaverjasPhase ゲームフェーズ
type KlaverjasPhase int

// Klaverjas のフェーズ定数
const (
	// KlaverjasPhasePlay トリックプレイフェーズ
	KlaverjasPhasePlay KlaverjasPhase = 0
	// KlaverjasPhaseTrickEnd トリック終了フェーズ
	KlaverjasPhaseTrickEnd KlaverjasPhase = 1
	// KlaverjasPhaseRoundEnd ラウンド終了フェーズ
	KlaverjasPhaseRoundEnd KlaverjasPhase = 2
	// KlaverjasPhaseGameEnd ゲーム終了フェーズ
	KlaverjasPhaseGameEnd KlaverjasPhase = 3
)

// KlaverjasHint ヒント情報
type KlaverjasHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Klaverjas クラヴァヤスのゲームクラス
type Klaverjas struct {
	trumpCards       *TrumpCards
	players          []*KlaverjasPlayer
	config           KlaverjasConfig
	phase            KlaverjasPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int                   // 切り札スート
	teamScores       [KlaverjasTeamCnt]int // 累積点
	roundCardPts     [KlaverjasTeamCnt]int // 現ラウンドのカード得点 (最終ボーナス含む)
	roundRoem        [KlaverjasTeamCnt]int // 現ラウンドの Roem 点
	gameEndFlag      bool
	winnerTeam       int // -1=未確定
	actionLogBase
}

// NewKlaverjas コンストラクタ
func NewKlaverjas(trumpCards *TrumpCards, players []*KlaverjasPlayer, config KlaverjasConfig) *Klaverjas {
	return &Klaverjas{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1}
}

// NewDefaultKlaverjas 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultKlaverjas() *Klaverjas {
	players := make([]*KlaverjasPlayer, KlaverjasPlayerCnt)
	players[0] = NewKlaverjasPlayer(true)
	for i := 1; i < KlaverjasPlayerCnt; i++ {
		players[i] = NewKlaverjasPlayer(false)
	}
	return NewKlaverjas(NewTrumpCardsBelote(), players, DefaultKlaverjasConfig())
}

// KlaverjasTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func KlaverjasTeamOf(playerIdx int) int { return playerIdx % KlaverjasTeamCnt }

// klaverjasTeamName チーム番号を表示名 (A/B) に変換する。
// (共有ヘルパー teamName は casino ワーカー専用ファイル定義のため、classic
// ワーカーでコンパイルできるよう Klaverjas 内に持つ)
func klaverjasTeamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// Reset ゲーム初期化
func (g *Klaverjas) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [KlaverjasTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Klaverjas) NextRound() {
	if g.phase != KlaverjasPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % KlaverjasPlayerCnt
	g.startRound()
}

// startRound 手札を配り、切り札と Roem を決めてプレイフェーズを開始する。
func (g *Klaverjas) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundCardPts = [KlaverjasTeamCnt]int{}
	g.roundRoem = [KlaverjasTeamCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()
	g.detectRoem()

	g.leadPlayerIdx = (g.dealerIdx + 1) % KlaverjasPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = KlaverjasPhasePlay
}

// deal 各プレイヤーへ 8 枚を配り、最後に配った札のスートを切り札とする。
func (g *Klaverjas) deal() {
	var last *Card
	for i := 0; i < KlaverjasHandSize; i++ {
		for j := 0; j < KlaverjasPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % KlaverjasPlayerCnt
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

// detectRoem 各プレイヤーの初手から Roem (3+連続札・4枚組) を検出してチームへ加点する。
func (g *Klaverjas) detectRoem() {
	for i := range g.players {
		roem := klaverjasHandRoem(g.players[i])
		if roem > 0 {
			g.roundRoem[KlaverjasTeamOf(i)] += roem
			g.appendLog(i, "roem", fmt.Sprintf("%s scores %d roem", g.playerName(i), roem), nil)
		}
	}
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Klaverjas) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KlaverjasPhasePlay {
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
func (g *Klaverjas) CpuPlay() {
	if g.gameEndFlag || g.phase != KlaverjasPhasePlay {
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
func (g *Klaverjas) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == KlaverjasPlayerCnt {
		g.phase = KlaverjasPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % KlaverjasPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Klaverjas) ResolveTrick() {
	if g.phase != KlaverjasPhaseTrickEnd || len(g.currentTrick) != KlaverjasPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += g.cardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	team := KlaverjasTeamOf(winnerIdx)
	g.roundCardPts[team] += pts
	bonus := ""
	if g.trickNumber >= KlaverjasTrickCount {
		g.roundCardPts[team] += KlaverjasLastTrickBonus
		bonus = fmt.Sprintf(" +%d last", KlaverjasLastTrickBonus)
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d%s)", g.playerName(winnerIdx), g.trickNumber, pts, bonus), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= KlaverjasTrickCount {
		g.phase = KlaverjasPhaseRoundEnd
	} else {
		g.phase = KlaverjasPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Klaverjas) NextTrick() {
	if g.phase != KlaverjasPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = KlaverjasPhasePlay
}

// ScoreRound ラウンド得点 (カード + Roem) を累積し、マッチ終了を判定する。
func (g *Klaverjas) ScoreRound() {
	if g.phase != KlaverjasPhaseRoundEnd {
		return
	}
	for t := 0; t < KlaverjasTeamCnt; t++ {
		g.teamScores[t] += g.roundCardPts[t] + g.roundRoem[t]
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: A=%d (cards %d + roem %d), B=%d (cards %d + roem %d)",
			g.roundNumber, g.teamScores[0], g.roundCardPts[0], g.roundRoem[0],
			g.teamScores[1], g.roundCardPts[1], g.roundRoem[1]), nil)

	leader, other := 0, 1
	if g.teamScores[1] > g.teamScores[0] {
		leader, other = 1, 0
	}
	if g.teamScores[leader] >= g.config.TargetPoints && g.teamScores[leader] > g.teamScores[other] {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = KlaverjasPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the match!", klaverjasTeamName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォロー + 切り札強制追い越し (overtrump) を検証する。
func (g *Klaverjas) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	hasLeadSuit := g.playerHasSuit(playerIdx, leadSuit)
	// リードスートを持っていれば必ず従う。
	if hasLeadSuit && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	// リードスートのボイド: 切り札を持っていれば切り札を出す義務がある。
	if !hasLeadSuit && g.playerHasSuit(playerIdx, g.trumpSuit) && card.GetDesign() != g.trumpSuit {
		return NewDomainError(ErrInvalidPlay, "切り札を出してください")
	}
	// 切り札を出す場合、追い越せる切り札があるなら追い越す義務がある。
	// (リードスート自体が切り札のケースでも判定が漏れないよう、ここで一括検証する)
	highest := g.highestTrumpStrengthInTrick()
	if highest >= 0 && card.GetDesign() == g.trumpSuit {
		if g.trumpStrength(card.GetValue()) <= highest && g.canOvertrump(playerIdx, highest) {
			return NewDomainError(ErrInvalidPlay, "より強い切り札で追い越してください")
		}
	}
	return nil
}

// highestTrumpStrengthInTrick 現在のトリック中の最強切り札の強さ (-1=切り札なし)。
func (g *Klaverjas) highestTrumpStrengthInTrick() int {
	best := -1
	for _, tc := range g.currentTrick {
		if tc.Card.GetDesign() == g.trumpSuit {
			if s := g.trumpStrength(tc.Card.GetValue()); s > best {
				best = s
			}
		}
	}
	return best
}

// canOvertrump プレイヤーが strength を超える切り札を持っているか。
func (g *Klaverjas) canOvertrump(playerIdx, strength int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == g.trumpSuit && g.trumpStrength(c.GetValue()) > strength {
			return true
		}
	}
	return false
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Klaverjas) playerHasSuit(playerIdx, design int) bool {
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
func (g *Klaverjas) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.klaverjasRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.klaverjasRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// klaverjasRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Klaverjas) klaverjasRank(card *Card) int {
	if card.GetDesign() == g.trumpSuit {
		return 100 + g.trumpStrength(card.GetValue())
	}
	return klaverjasPlainStrength(card.GetValue())
}

// trumpStrength 切り札の強さ。J>9>A>10>K>Q>8>7。
func (g *Klaverjas) trumpStrength(value int) int {
	switch value {
	case 11: // Jack (Jas)
		return 8
	case 9: // Nel
		return 7
	case 1: // Ace
		return 6
	case 10:
		return 5
	case 13: // King
		return 4
	case 12: // Queen
		return 3
	case 8:
		return 2
	default: // 7
		return 1
	}
}

// klaverjasPlainStrength 非切り札の強さ。A>10>K>Q>J>9>8>7。
func klaverjasPlainStrength(value int) int {
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

// cardPoints カードポイント。切り札か否かで配点が異なる。
func (g *Klaverjas) cardPoints(card *Card) int {
	if card.GetDesign() == g.trumpSuit {
		switch card.GetValue() {
		case 11: // Jack
			return 20
		case 9: // Nel
			return 14
		case 1:
			return 11
		case 10:
			return 10
		case 13:
			return 4
		case 12:
			return 3
		default:
			return 0
		}
	}
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
func (g *Klaverjas) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Roem detection ---

// klaverjasSeqRank 連続札判定用のランク (A=8,K=7,Q=6,J=5,10=4,9=3,8=2,7=1)。
func klaverjasSeqRank(value int) int {
	switch value {
	case 1:
		return 8
	case 13:
		return 7
	case 12:
		return 6
	case 11:
		return 5
	case 10:
		return 4
	case 9:
		return 3
	case 8:
		return 2
	default: // 7
		return 1
	}
}

// klaverjasHandRoem 手札の Roem 点を計算する。スートごとの連続 3 枚=20、4 枚=50、
// 同位 4 枚組 (7/8 を除く) = 100。
func klaverjasHandRoem(p *KlaverjasPlayer) int {
	total := 0
	// 4 枚組 (4 of a kind)。
	rankCnt := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		rankCnt[p.GetCard(i).GetValue()]++
	}
	for v, c := range rankCnt {
		if c == 4 && v != 7 && v != 8 {
			total += 100
		}
	}
	// スートごとの連続札。
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		var seq []int
		for i := 0; i < p.GetCardsSize(); i++ {
			c := p.GetCard(i)
			if c.GetDesign() == suit {
				seq = append(seq, klaverjasSeqRank(c.GetValue()))
			}
		}
		total += klaverjasRunPoints(seq)
	}
	return total
}

// klaverjasRunPoints 連続ランク列から Roem 点を求める。同一スート内に独立した
// 連続列が複数あれば、それぞれを加算する (3 枚=20, 4 枚以上=50)。
func klaverjasRunPoints(seqRanks []int) int {
	if len(seqRanks) < 3 {
		return 0
	}
	s := append([]int(nil), seqRanks...)
	sort.Ints(s)
	total, run := 0, 1
	scoreRun := func(n int) {
		switch {
		case n >= 4:
			total += 50
		case n == 3:
			total += 20
		}
	}
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1]+1 {
			run++
		} else if s[i] != s[i-1] {
			scoreRun(run)
			run = 1
		}
	}
	scoreRun(run)
	return total
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Klaverjas) sortAllHands() {
	for _, p := range g.players {
		klaverjasSortHand(p)
	}
}

// klaverjasSortHand 手札をスート→自然ランク順にソートする。
func klaverjasSortHand(p *KlaverjasPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return klaverjasSeqRank(cards[i].GetValue()) > klaverjasSeqRank(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Klaverjas) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Klaverjas) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Klaverjas) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.klaverjasRank(g.currentTrick[idx].Card)
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Klaverjas) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Klaverjas) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == KlaverjasCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *Klaverjas) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int {
			return g.cardPoints(c)*100 + g.klaverjasRank(c)
		})
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	partnerWinning := KlaverjasTeamOf(winnerIdx) == KlaverjasTeamOf(playerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += g.cardPoints(tc.Card)
	}
	winners := klaverjasFilter(valid, func(idx int) bool { return g.klaverjasRank(player.GetCard(idx)) > topRank })
	if partnerWinning {
		// 味方が勝っている: 得点の高い札を渡す (勝ち札以外があればそれ)。
		nonWinners := klaverjasFilter(valid, func(idx int) bool { return g.klaverjasRank(player.GetCard(idx)) < topRank })
		if len(nonWinners) > 0 {
			return pickHighest(player, nonWinners, func(c *Card) int { return g.cardPoints(c) })
		}
		return pickLowest(player, valid, func(c *Card) int { return g.klaverjasRank(c) })
	}
	if trickPts > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.klaverjasRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int {
		return g.cardPoints(c)*100 + g.klaverjasRank(c)
	})
}

// klaverjasFilter 述語を満たすインデックスを抽出する。
func klaverjasFilter(indices []int, pred func(int) bool) []int {
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
func (g *Klaverjas) GetHint() *KlaverjasHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != KlaverjasPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &KlaverjasHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Klaverjas) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.klaverjasRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Klaverjas) GetPhase() KlaverjasPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Klaverjas) SetPhase(phase KlaverjasPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Klaverjas) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Klaverjas) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Klaverjas) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Klaverjas) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Klaverjas) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Klaverjas) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Klaverjas) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Klaverjas) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Klaverjas) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Klaverjas) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Klaverjas) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit 切り札スート取得
func (g *Klaverjas) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Klaverjas) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTeamScores チーム別累積点取得
func (g *Klaverjas) GetTeamScores() [KlaverjasTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点設定 (テスト用)
func (g *Klaverjas) SetTeamScores(s [KlaverjasTeamCnt]int) { g.teamScores = s }

// GetRoundCardPoints 現ラウンドのカード得点取得
func (g *Klaverjas) GetRoundCardPoints() [KlaverjasTeamCnt]int { return g.roundCardPts }

// SetRoundCardPoints 現ラウンドのカード得点設定 (テスト用)
func (g *Klaverjas) SetRoundCardPoints(s [KlaverjasTeamCnt]int) { g.roundCardPts = s }

// GetRoundRoem 現ラウンドの Roem 点取得
func (g *Klaverjas) GetRoundRoem() [KlaverjasTeamCnt]int { return g.roundRoem }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Klaverjas) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *Klaverjas) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Klaverjas) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Klaverjas) GetPlayer(i int) *KlaverjasPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間か。
func (g *Klaverjas) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Klaverjas) GetConfig() KlaverjasConfig { return g.config }

// SetConfig 設定変更
func (g *Klaverjas) SetConfig(cfg KlaverjasConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Klaverjas) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != KlaverjasPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// klaverjasJSON is the JSON wire format for Klaverjas.
type klaverjasJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*KlaverjasPlayer    `json:"ps"`
	Config           KlaverjasConfig       `json:"cf"`
	Phase            KlaverjasPhase        `json:"ph"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	CurrentPlayerIdx int                   `json:"ci"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	LeadPlayerIdx    int                   `json:"li"`
	DealerIdx        int                   `json:"di"`
	TrumpSuit        int                   `json:"ts"`
	TeamScores       [KlaverjasTeamCnt]int `json:"sc"`
	RoundCardPts     [KlaverjasTeamCnt]int `json:"rp"`
	RoundRoem        [KlaverjasTeamCnt]int `json:"rr"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerTeam       int                   `json:"wt"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Klaverjas) MarshalJSON() ([]byte, error) {
	return json.Marshal(klaverjasJSON{
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
		RoundRoem:        g.roundRoem,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// klaverjasMaxSliceLen caps slice sizes during deserialisation.
const klaverjasMaxSliceLen = 5000

// errKlaverjasOversized is the single sentinel error for oversized input arrays.
var errKlaverjasOversized = errors.New("klaverjas: input array exceeds maximum allowed size")

// errKlaverjasInvalidPlayers is returned when restored state lacks exactly KlaverjasPlayerCnt players.
var errKlaverjasInvalidPlayers = errors.New("klaverjas: invalid player count")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Klaverjas) UnmarshalJSON(data []byte) error {
	var j klaverjasJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > klaverjasMaxSliceLen || len(j.CurrentTrick) > klaverjasMaxSliceLen ||
		len(j.ActionLog) > klaverjasMaxSliceLen {
		return errKlaverjasOversized
	}
	if len(j.Players) != KlaverjasPlayerCnt {
		return errKlaverjasInvalidPlayers
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
	g.roundRoem = j.RoundRoem
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
