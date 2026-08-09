//go:build !js || !wasm || casino

// Package domain トゥエンティナイン (Twenty-Nine / 29) のドメインモデル。
//
// 29 はインド・バングラデシュ発祥の 4 人 2 チーム入札トリックテイキング。32 枚デッキ
// (7〜A) を 8 枚ずつ配り、入札フェーズで各プレイヤーが 16〜28 点を 1 回宣言する。最高
// 入札者のチームが切り札を「非公開 (Hidden Trump)」で選び、誰かがフォローできず切り札/
// 別スートを出した瞬間に公開される (本実装では公開フラグのみ保持)。得点カードは
// J=3, 9=2, A=1, 10=1 (合計 28) で、最終トリック勝者に +1 (合計 29)。落札チームが宣言点
// 以上の得点を取れば 1 ゲーム点を獲得、失敗は相手チームが獲得し、先に目標 (既定 6) へ
// 達したチームが勝利する。
//
// トリックの強さ (切り札・非切り札共通): J > 9 > A > 10 > K > Q > 8 > 7。切り札は非切り札
// より常に強い。簡略化: 切り札は落札者の最長スートを自動選択する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// TwentyNinePlayerCnt プレイヤー数 (人間 1 + CPU 3)
const TwentyNinePlayerCnt = 4

// TwentyNineTeamCnt チーム数
const TwentyNineTeamCnt = 2

// TwentyNineHandSize 各プレイヤーの手札枚数
const TwentyNineHandSize = 8

// TwentyNineTrickCount 1 ラウンドのトリック数
const TwentyNineTrickCount = 8

// TwentyNineBid 入札種別 (値はそのまま宣言点; Pass=0)
type TwentyNineBid int

// 29 の入札定数
const (
	// TwentyNineBidPass パス
	TwentyNineBidPass TwentyNineBid = 0
	// TwentyNineBidSixteen 16 点宣言 (最低)
	TwentyNineBidSixteen TwentyNineBid = 16
	// TwentyNineBidTwenty 20 点宣言
	TwentyNineBidTwenty TwentyNineBid = 20
	// TwentyNineBidTwentyFour 24 点宣言
	TwentyNineBidTwentyFour TwentyNineBid = 24
	// TwentyNineBidTwentyEight 28 点宣言 (最高)
	TwentyNineBidTwentyEight TwentyNineBid = 28
)

// TwentyNinePhase ゲームフェーズ
type TwentyNinePhase int

// 29 のフェーズ定数
const (
	// TwentyNinePhaseBid 入札フェーズ
	TwentyNinePhaseBid TwentyNinePhase = 0
	// TwentyNinePhasePlay トリックプレイフェーズ
	TwentyNinePhasePlay TwentyNinePhase = 1
	// TwentyNinePhaseTrickEnd トリック終了フェーズ
	TwentyNinePhaseTrickEnd TwentyNinePhase = 2
	// TwentyNinePhaseRoundEnd ラウンド終了フェーズ
	TwentyNinePhaseRoundEnd TwentyNinePhase = 3
	// TwentyNinePhaseGameEnd ゲーム終了フェーズ
	TwentyNinePhaseGameEnd TwentyNinePhase = 4
)

// TwentyNineHint ヒント情報
type TwentyNineHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// TwentyNine トゥエンティナインのゲームクラス
type TwentyNine struct {
	trumpCards       *TrumpCards
	players          []*TwentyNinePlayer
	config           TwentyNineConfig
	phase            TwentyNinePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	bids             [TwentyNinePlayerCnt]TwentyNineBid
	bidDone          [TwentyNinePlayerCnt]bool
	declarerIdx      int // 落札者 (-1=未確定/全パス)
	contract         TwentyNineBid
	trumpSuit        int
	trumpRevealed    bool                   // 隠し切り札が公開済みか
	teamScores       [TwentyNineTeamCnt]int // 累積ゲーム点
	roundTeamPts     [TwentyNineTeamCnt]int // 現ラウンドのチーム別カード得点
	gameEndFlag      bool
	winnerTeam       int // -1=未確定
	actionLogBase
}

// NewTwentyNine コンストラクタ
func NewTwentyNine(trumpCards *TrumpCards, players []*TwentyNinePlayer, config TwentyNineConfig) *TwentyNine {
	return &TwentyNine{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1, declarerIdx: -1}
}

// NewDefaultTwentyNine 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultTwentyNine() *TwentyNine {
	players := make([]*TwentyNinePlayer, TwentyNinePlayerCnt)
	players[0] = NewTwentyNinePlayer(true)
	for i := 1; i < TwentyNinePlayerCnt; i++ {
		players[i] = NewTwentyNinePlayer(false)
	}
	return NewTwentyNine(NewTrumpCardsBelote(), players, DefaultTwentyNineConfig())
}

// TwentyNineTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func TwentyNineTeamOf(playerIdx int) int { return playerIdx % TwentyNineTeamCnt }

// twentyNineTeamName チーム番号を表示名 (A/B) に変換する (casino ワーカーで自己完結)。
func twentyNineTeamName(team int) string {
	if team == 0 {
		return "A"
	}
	return "B"
}

// Reset ゲーム初期化
func (g *TwentyNine) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [TwentyNineTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *TwentyNine) NextRound() {
	if g.phase != TwentyNinePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % TwentyNinePlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *TwentyNine) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.bids = [TwentyNinePlayerCnt]TwentyNineBid{}
	g.bidDone = [TwentyNinePlayerCnt]bool{}
	g.declarerIdx = -1
	g.contract = TwentyNineBidPass
	g.trumpSuit = 0
	g.trumpRevealed = false
	g.roundTeamPts = [TwentyNineTeamCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	g.currentPlayerIdx = (g.dealerIdx + 1) % TwentyNinePlayerCnt // forehand bids first
	g.phase = TwentyNinePhaseBid
}

// deal 各プレイヤーへ 8 枚を配る。
func (g *TwentyNine) deal() {
	for i := 0; i < TwentyNineHandSize; i++ {
		for j := 0; j < TwentyNinePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % TwentyNinePlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// IsHumanBidTurn 入札フェーズで人間の手番か。
func (g *TwentyNine) IsHumanBidTurn() bool {
	return g.phase == TwentyNinePhaseBid && g.currentPlayerIdx >= 0 &&
		g.currentPlayerIdx < len(g.players) && g.players[g.currentPlayerIdx].GetIsHuman()
}

// highestBid 現在の最高入札と入札者を返す (-1=なし)。
func (g *TwentyNine) highestBid() (TwentyNineBid, int) {
	best, bestIdx := TwentyNineBidPass, -1
	for i := 0; i < TwentyNinePlayerCnt; i++ {
		if g.bidDone[i] && g.bids[i] > best {
			best = g.bids[i]
			bestIdx = i
		}
	}
	return best, bestIdx
}

// PlayerBid 人間プレイヤーが入札する。
func (g *TwentyNine) PlayerBid(bid TwentyNineBid) error {
	if g.phase != TwentyNinePhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyBid(g.currentPlayerIdx, bid)
}

// CpuBid 入札フェーズで CPU が 1 件入札する。
func (g *TwentyNine) CpuBid() {
	if g.phase != TwentyNinePhaseBid {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	_ = g.applyBid(idx, g.cpuChooseBid(idx))
}

// twentyNineIsValidBid は bid が許可された入札定数のいずれかであるかを返す。
func twentyNineIsValidBid(bid TwentyNineBid) bool {
	return bid == TwentyNineBidPass || bid == TwentyNineBidSixteen || bid == TwentyNineBidTwenty ||
		bid == TwentyNineBidTwentyFour || bid == TwentyNineBidTwentyEight
}

// applyBid 入札を記録し、次の入札者へ進める。全員入札したら契約を確定する。
func (g *TwentyNine) applyBid(idx int, bid TwentyNineBid) error {
	if !twentyNineIsValidBid(bid) {
		return NewDomainError(ErrInvalidPlay, "入札値が不正です")
	}
	high, _ := g.highestBid()
	if bid != TwentyNineBidPass && bid <= high {
		return NewDomainError(ErrInvalidPlay, "現在の入札を上回る必要があります")
	}
	g.bids[idx] = bid
	g.bidDone[idx] = true
	if bid != TwentyNineBidPass {
		g.appendLog(idx, "bid", fmt.Sprintf("%s bids %d", playerName(g.players, idx), int(bid)), nil)
	} else {
		g.appendLog(idx, "bid", fmt.Sprintf("%s passes", playerName(g.players, idx)), nil)
	}
	for k := 1; k <= TwentyNinePlayerCnt; k++ {
		ni := (idx + k) % TwentyNinePlayerCnt
		if !g.bidDone[ni] {
			g.currentPlayerIdx = ni
			return nil
		}
	}
	g.resolveBidding()
	return nil
}

// resolveBidding 入札を締め、落札者・契約・隠し切り札を確定してプレイへ移る。
func (g *TwentyNine) resolveBidding() {
	bid, idx := g.highestBid()
	if idx < 0 || bid == TwentyNineBidPass {
		g.declarerIdx = -1
		g.phase = TwentyNinePhaseRoundEnd
		g.appendLog(-1, "passed_out", "all players passed; round is void", nil)
		return
	}
	g.declarerIdx = idx
	g.contract = bid
	g.trumpSuit = g.longestSuit(idx)
	g.trumpRevealed = false
	g.appendLog(idx, "contract",
		fmt.Sprintf("%s (team %s) bids %d with a hidden trump", playerName(g.players, idx), twentyNineTeamName(TwentyNineTeamOf(idx)), int(bid)), nil)
	g.leadPlayerIdx = idx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = TwentyNinePhasePlay
}

// longestSuit プレイヤーが最も多く持つスートを返す。
func (g *TwentyNine) longestSuit(playerIdx int) int {
	counts := map[int]int{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	bestSuit, bestCnt := CardDesignSpade, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if counts[suit] > bestCnt {
			bestCnt = counts[suit]
			bestSuit = suit
		}
	}
	return bestSuit
}

// cpuChooseBid CPU の入札を選ぶ。得点札と最長スートから見積もる。
func (g *TwentyNine) cpuChooseBid(idx int) TwentyNineBid {
	high, _ := g.highestBid()
	suit := g.longestSuit(idx)
	cnt, pts := 0, 0
	p := g.players[idx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == suit {
			cnt++
		}
		pts += twentyNineCardPoints(c)
	}
	want := TwentyNineBidPass
	switch {
	case cnt >= 5 && pts >= 8:
		want = TwentyNineBidTwentyFour
	case cnt >= 4 && pts >= 6:
		want = TwentyNineBidTwenty
	case cnt >= 4 || pts >= 6:
		want = TwentyNineBidSixteen
	}
	if want > high {
		return want
	}
	return TwentyNineBidPass
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *TwentyNine) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TwentyNinePhasePlay {
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
func (g *TwentyNine) CpuPlay() {
	if g.gameEndFlag || g.phase != TwentyNinePhasePlay {
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
func (g *TwentyNine) playCard(playerIdx int, card *Card) {
	// 隠し切り札の公開: リードスートに従えず別スート (切り札含む) が出された瞬間に公開。
	if len(g.currentTrick) > 0 && !g.trumpRevealed {
		leadSuit := g.currentTrick[0].Card.GetDesign()
		if card.GetDesign() != leadSuit {
			g.trumpRevealed = true
			g.appendLog(playerIdx, "reveal_trump", fmt.Sprintf("trump (%d) is revealed", g.trumpSuit), nil)
		}
	}
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == TwentyNinePlayerCnt {
		g.phase = TwentyNinePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TwentyNinePlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *TwentyNine) ResolveTrick() {
	if g.phase != TwentyNinePhaseTrickEnd || len(g.currentTrick) != TwentyNinePlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += twentyNineCardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	team := TwentyNineTeamOf(winnerIdx)
	g.roundTeamPts[team] += pts
	bonus := ""
	if g.trickNumber >= TwentyNineTrickCount {
		g.roundTeamPts[team]++ // 最終トリック +1
		bonus = " +1 last"
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d%s)", playerName(g.players, winnerIdx), g.trickNumber, pts, bonus), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= TwentyNineTrickCount {
		g.phase = TwentyNinePhaseRoundEnd
	} else {
		g.phase = TwentyNinePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *TwentyNine) NextTrick() {
	if g.phase != TwentyNinePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = TwentyNinePhasePlay
}

// ScoreRound 落札チームの達成可否を判定し、ゲーム点を加算してマッチ終了を判定する。
func (g *TwentyNine) ScoreRound() {
	if g.phase != TwentyNinePhaseRoundEnd {
		return
	}
	if g.declarerIdx >= 0 {
		bidTeam := TwentyNineTeamOf(g.declarerIdx)
		otherTeam := 1 - bidTeam
		made := g.roundTeamPts[bidTeam] >= int(g.contract)
		if made {
			g.teamScores[bidTeam]++
		} else {
			g.teamScores[otherTeam]++
		}
		g.appendLog(-1, "round_score",
			fmt.Sprintf("round %d: team %s bid %d, got %d -> %s",
				g.roundNumber, twentyNineTeamName(bidTeam), int(g.contract), g.roundTeamPts[bidTeam],
				map[bool]string{true: "made", false: "set"}[made]), nil)
		g.checkGameEnd()
	}
}

// checkGameEnd 目標ゲーム点到達でマッチ終了を判定する。
func (g *TwentyNine) checkGameEnd() {
	leader, best := -1, -1
	for t := 0; t < TwentyNineTeamCnt; t++ {
		if g.teamScores[t] > best {
			best = g.teamScores[t]
			leader = t
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerTeam = leader
		g.phase = TwentyNinePhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the match!", twentyNineTeamName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する。
func (g *TwentyNine) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if g.playerHasSuit(playerIdx, leadSuit) && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *TwentyNine) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(g.players[playerIdx], design)
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *TwentyNine) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.twentyNineRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.twentyNineRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// twentyNineRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *TwentyNine) twentyNineRank(card *Card) int {
	if card.GetDesign() == g.trumpSuit {
		return 100 + twentyNineStrength(card.GetValue())
	}
	return twentyNineStrength(card.GetValue())
}

// twentyNineStrength カード強度。J > 9 > A > 10 > K > Q > 8 > 7。
func twentyNineStrength(value int) int {
	switch value {
	case 11: // Jack
		return 8
	case 9:
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

// twentyNineCardPoints カードポイント。J=3, 9=2, A=1, 10=1, 他=0 (合計 28)。
func twentyNineCardPoints(card *Card) int {
	switch card.GetValue() {
	case 11: // Jack
		return 3
	case 9:
		return 2
	case 1, 10: // Ace, Ten
		return 1
	default:
		return 0
	}
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *TwentyNine) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *TwentyNine) sortAllHands() {
	for _, p := range g.players {
		twentyNineSortHand(p)
	}
}

// twentyNineSortHand 手札をスート→強さ順にソートする。
func twentyNineSortHand(p *TwentyNinePlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return twentyNineStrength(cards[i].GetValue()) > twentyNineStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *TwentyNine) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *TwentyNine) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.twentyNineRank(g.currentTrick[idx].Card)
}

// --- CPU AI ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *TwentyNine) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == TwentyNineCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *TwentyNine) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int { return twentyNineCardPoints(c)*100 + g.twentyNineRank(c) })
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	partnerWinning := TwentyNineTeamOf(winnerIdx) == TwentyNineTeamOf(playerIdx) && winnerIdx != playerIdx
	winners := twentyNineFilter(valid, func(idx int) bool { return g.twentyNineRank(player.GetCard(idx)) > topRank })
	if partnerWinning {
		// 味方が勝っている: 得点札を渡す。
		return pickHighest(player, valid, func(c *Card) int { return twentyNineCardPoints(c) })
	}
	if len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.twentyNineRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int { return twentyNineCardPoints(c)*100 + g.twentyNineRank(c) })
}

// twentyNineFilter 述語を満たすインデックスを抽出する。
func twentyNineFilter(indices []int, pred func(int) bool) []int {
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
func (g *TwentyNine) GetHint() *TwentyNineHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != TwentyNinePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &TwentyNineHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *TwentyNine) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.twentyNineRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *TwentyNine) GetPhase() TwentyNinePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *TwentyNine) SetPhase(phase TwentyNinePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *TwentyNine) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *TwentyNine) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *TwentyNine) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *TwentyNine) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *TwentyNine) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *TwentyNine) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *TwentyNine) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *TwentyNine) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *TwentyNine) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *TwentyNine) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *TwentyNine) GetDealerIdx() int { return g.dealerIdx }

// GetDeclarerIdx 落札者インデックス取得 (-1=未確定)
func (g *TwentyNine) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 落札者インデックス設定 (テスト用)
func (g *TwentyNine) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract 確定した契約を取得
func (g *TwentyNine) GetContract() TwentyNineBid { return g.contract }

// SetContract 契約設定 (テスト用)
func (g *TwentyNine) SetContract(b TwentyNineBid) { g.contract = b }

// GetBids 各プレイヤーの入札を取得
func (g *TwentyNine) GetBids() [TwentyNinePlayerCnt]TwentyNineBid { return g.bids }

// GetTrumpSuit 切り札スート取得 (0=未確定)
func (g *TwentyNine) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *TwentyNine) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTrumpRevealed 隠し切り札が公開済みか取得
func (g *TwentyNine) GetTrumpRevealed() bool { return g.trumpRevealed }

// SetTrumpRevealed 切り札公開フラグ設定 (テスト用)
func (g *TwentyNine) SetTrumpRevealed(v bool) { g.trumpRevealed = v }

// GetTeamScores チーム別累積ゲーム点取得
func (g *TwentyNine) GetTeamScores() [TwentyNineTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積ゲーム点設定 (テスト用)
func (g *TwentyNine) SetTeamScores(s [TwentyNineTeamCnt]int) { g.teamScores = s }

// GetRoundTeamPoints 現ラウンドのチーム別カード得点取得
func (g *TwentyNine) GetRoundTeamPoints() [TwentyNineTeamCnt]int { return g.roundTeamPts }

// SetRoundTeamPoints 現ラウンドのチーム別カード得点設定 (テスト用)
func (g *TwentyNine) SetRoundTeamPoints(s [TwentyNineTeamCnt]int) { g.roundTeamPts = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *TwentyNine) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *TwentyNine) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *TwentyNine) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *TwentyNine) GetPlayer(i int) *TwentyNinePlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間か (プレイフェーズ)。
func (g *TwentyNine) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *TwentyNine) GetConfig() TwentyNineConfig { return g.config }

// SetConfig 設定変更
func (g *TwentyNine) SetConfig(cfg TwentyNineConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *TwentyNine) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != TwentyNinePhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// twentyNineJSON is the JSON wire format for TwentyNine.
type twentyNineJSON struct {
	TrumpCards       *TrumpCards                        `json:"tc"`
	Players          []*TwentyNinePlayer                `json:"ps"`
	Config           TwentyNineConfig                   `json:"cf"`
	Phase            TwentyNinePhase                    `json:"ph"`
	RoundNumber      int                                `json:"rn"`
	TrickNumber      int                                `json:"tn"`
	CurrentPlayerIdx int                                `json:"ci"`
	CurrentTrick     []*TrickCard                       `json:"ct"`
	LeadPlayerIdx    int                                `json:"li"`
	DealerIdx        int                                `json:"di"`
	Bids             [TwentyNinePlayerCnt]TwentyNineBid `json:"bd"`
	BidDone          [TwentyNinePlayerCnt]bool          `json:"bf"`
	DeclarerIdx      int                                `json:"dc"`
	Contract         TwentyNineBid                      `json:"co"`
	TrumpSuit        int                                `json:"ts"`
	TrumpRevealed    bool                               `json:"tr"`
	TeamScores       [TwentyNineTeamCnt]int             `json:"sc"`
	RoundTeamPts     [TwentyNineTeamCnt]int             `json:"rp"`
	GameEndFlag      bool                               `json:"ge"`
	WinnerTeam       int                                `json:"wt"`
	ActionLog        []*ActionLogEntry                  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *TwentyNine) MarshalJSON() ([]byte, error) {
	return json.Marshal(twentyNineJSON{
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
		Bids:             g.bids,
		BidDone:          g.bidDone,
		DeclarerIdx:      g.declarerIdx,
		Contract:         g.contract,
		TrumpSuit:        g.trumpSuit,
		TrumpRevealed:    g.trumpRevealed,
		TeamScores:       g.teamScores,
		RoundTeamPts:     g.roundTeamPts,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// twentyNineMaxSliceLen caps slice sizes during deserialisation.
const twentyNineMaxSliceLen = 5000

// errTwentyNineOversized is the single sentinel error for oversized input arrays.
var errTwentyNineOversized = errors.New("twentynine: input array exceeds maximum allowed size")

// errTwentyNineInvalidPlayers is returned when restored state lacks exactly TwentyNinePlayerCnt players.
var errTwentyNineInvalidPlayers = errors.New("twentynine: invalid player count")

// errTwentyNineInvalidTrick is returned when a restored trick card is nil/out of range.
var errTwentyNineInvalidTrick = errors.New("twentynine: invalid trick card")

// errTwentyNineInvalidState is returned when a restored index/state field is out of range.
var errTwentyNineInvalidState = errors.New("twentynine: invalid state values in json")

// UnmarshalJSON implements json.Unmarshaler.
func (g *TwentyNine) UnmarshalJSON(data []byte) error {
	var j twentyNineJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > twentyNineMaxSliceLen || len(j.CurrentTrick) > twentyNineMaxSliceLen ||
		len(j.ActionLog) > twentyNineMaxSliceLen {
		return errTwentyNineOversized
	}
	if len(j.Players) != TwentyNinePlayerCnt {
		return errTwentyNineInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errTwentyNineInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= TwentyNinePlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= TwentyNinePlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= TwentyNinePlayerCnt ||
		j.DeclarerIdx < -1 || j.DeclarerIdx >= TwentyNinePlayerCnt ||
		j.WinnerTeam < -1 || j.WinnerTeam >= TwentyNineTeamCnt ||
		j.TrumpSuit < 0 || j.TrumpSuit > 4 ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > TwentyNineTrickCount ||
		j.Phase < TwentyNinePhaseBid || j.Phase > TwentyNinePhaseGameEnd {
		return errTwentyNineInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= TwentyNinePlayerCnt {
			return errTwentyNineInvalidTrick
		}
	}
	// Bids is a fixed-size array, so no DoS size cap is needed; validate each
	// element (and the resolved contract) against the allowed bid constants so
	// a tampered payload cannot inject an out-of-range bid that bypasses the
	// scoring rules.
	if !twentyNineIsValidBid(j.Contract) {
		return errTwentyNineInvalidState
	}
	for _, b := range j.Bids {
		if !twentyNineIsValidBid(b) {
			return errTwentyNineInvalidState
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
	g.bids = j.Bids
	g.bidDone = j.BidDone
	g.declarerIdx = j.DeclarerIdx
	g.contract = j.Contract
	g.trumpSuit = j.TrumpSuit
	g.trumpRevealed = j.TrumpRevealed
	g.teamScores = j.TeamScores
	g.roundTeamPts = j.RoundTeamPts
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
