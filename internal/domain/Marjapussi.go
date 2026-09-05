//go:build !js || !wasm || extra5

// Package domain マルヤプッシ (Marjapussi) のドメインモデル。
//
// Marjapussi はフィンランドの 4 人用・2 対 2 チーム戦ポイント・トリックテイキングゲーム。
// 36 枚デッキ (6-A の 4 スート) から各自 8 枚を配り、残り 4 枚は伏せ札「ベリー袋 (pussi)」として
// 伏せる。競り (bid) や talon 交換は無く、配り後直ちにプレイフェーズへ進む。
//
// チーム戦:
// 席 0 と 席 2 がチーム 0、席 1 と 席 3 がチーム 1。得点はチーム単位で管理する。
//
// 結婚 (Marriage):
// リード時に同スートの K と Q を持っていれば、そのいずれかをリードすることで結婚を宣言できる。
// 結婚点は宣言時点の切り札スートと同じなら 40 点、違うなら 20 点。
// 違うスートを宣言した場合はそのスートが新しい切り札になる (ラウンド開始時は切り札なし)。
//
// フォロー規則 (裁定):
// マストフォロー。リードスートを持っていなければ、切り札を持っている場合は必ず切り札を出す。
// 切り札も無ければ任意の札を捨てられる。切り札が未決定の間はマストフォローのみで、フォローできなければ
// 任意の札を捨てられる (オーバートランプ義務なし)。
//
// ベリー袋 (pussi):
// 最後のトリック (第 8 トリック) を取ったプレイヤーのチームが pussi 4 枚を獲得し、そのカード点が
// そのチームのラウンド得点に加算される。
//
// 得点と終局:
// カード点: A=11, 10=10, K=4, Q=3, J=2, 9..6=0 (合計 120 点)。
// 毎ラウンド各チームは (獲得カード点 + 結婚点) を得点し、累積で PointLimit (既定 500 点) に
// 到達したチームがマッチの勝者となる。
// トリック強度: A > 10 > K > Q > J > 9 > 8 > 7 > 6。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// MarjapussiPlayerCnt プレイヤー数 (人間 1 + CPU 3 = 4 人)
const MarjapussiPlayerCnt = 4

// MarjapussiTeamCnt チーム数 (2 対 2)
const MarjapussiTeamCnt = 2

// MarjapussiHandSize 各プレイヤーの配り札枚数 (8 枚)
const MarjapussiHandSize = 8

// MarjapussiPussiSize ベリー袋 (pussi) の枚数 (4 枚)
const MarjapussiPussiSize = 4

// MarjapussiTrickCount 1 ラウンドのトリック数 (8 トリック)
const MarjapussiTrickCount = 8

// MarjapussiWinTarget マッチ勝利に必要な累積点 (デフォルト 500 点)
const MarjapussiWinTarget = 500

// MarjapussiPhase ゲームフェーズ
type MarjapussiPhase int

// Marjapussi のフェーズ定数 (Play(0) → TrickEnd(1) → RoundEnd(2) → GameEnd(3))
const (
	// MarjapussiPhasePlay トリックプレイフェーズ
	MarjapussiPhasePlay MarjapussiPhase = 0
	// MarjapussiPhaseTrickEnd トリック終了フェーズ
	MarjapussiPhaseTrickEnd MarjapussiPhase = 1
	// MarjapussiPhaseRoundEnd ラウンド終了フェーズ
	MarjapussiPhaseRoundEnd MarjapussiPhase = 2
	// MarjapussiPhaseGameEnd ゲーム終了フェーズ
	MarjapussiPhaseGameEnd MarjapussiPhase = 3
)

// MarjapussiPhaseMin フェーズ下限 (検証用)
const MarjapussiPhaseMin = int(MarjapussiPhasePlay)

// MarjapussiPhaseMax フェーズ上限 (検証用)
const MarjapussiPhaseMax = int(MarjapussiPhaseGameEnd)

// MarjapussiHint ヒント情報
type MarjapussiHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Reason      string // ヒント理由キー
}

// Marjapussi マルヤプッシのゲームクラス
type Marjapussi struct {
	trumpCards       *TrumpCards
	players          []*MarjapussiPlayer
	config           MarjapussiConfig
	phase            MarjapussiPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	trumpSuit        int                    // 切り札スート (0=未設定, CardDesignSpade..CardDesignDiamond)
	pussi            []*Card                // ベリー袋 (4 枚)
	teamScores       [MarjapussiTeamCnt]int // 累積チーム得点 (0: 席0+2, 1: 席1+3)
	roundCardPts     [MarjapussiTeamCnt]int // 現ラウンドのチーム別カード得点
	roundMarriage    [MarjapussiTeamCnt]int // 現ラウンドのチーム別結婚点
	lastTrickWinner  int                    // 最終トリック勝者 (-1=未確定)
	gameEndFlag      bool
	winnerPlayer     int // 勝利プレイヤー (-1=未確定)
	winnerTeam       int // 勝利チーム (-1=未確定)
	actionLogBase
}

// NewMarjapussi コンストラクタ
func NewMarjapussi(trumpCards *TrumpCards, players []*MarjapussiPlayer, config MarjapussiConfig) *Marjapussi {
	return &Marjapussi{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		winnerTeam:      -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultMarjapussi 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultMarjapussi() *Marjapussi {
	players := make([]*MarjapussiPlayer, MarjapussiPlayerCnt)
	players[0] = NewMarjapussiPlayer(true)
	for i := 1; i < MarjapussiPlayerCnt; i++ {
		players[i] = NewMarjapussiPlayer(false)
	}
	return NewMarjapussi(newMarjapussiDeck(), players, DefaultMarjapussiConfig())
}

// newMarjapussiDeck Marjapussi 用 36 枚デッキ (6-A × 4 スート) を生成する。
func newMarjapussiDeck() *TrumpCards {
	return NewTrumpCardsShortDeck()
}

// Reset ゲーム初期化
func (g *Marjapussi) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [MarjapussiTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Marjapussi) NextRound() {
	if g.phase != MarjapussiPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % MarjapussiPlayerCnt
	g.startRound()
}

// startRound 手札を配り、pussi を伏せ、プレイフェーズを開始する。
func (g *Marjapussi) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundCardPts = [MarjapussiTeamCnt]int{}
	g.roundMarriage = [MarjapussiTeamCnt]int{}
	g.lastTrickWinner = -1
	g.trumpSuit = 0
	g.pussi = nil
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	g.sortAllHands()

	// ディーラーの左隣からプレイ開始
	g.leadPlayerIdx = (g.dealerIdx + 1) % MarjapussiPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = MarjapussiPhasePlay
}

// deal 各プレイヤーへ 8 枚を配り、残り 4 枚を pussi (ベリー袋) にする。
func (g *Marjapussi) deal() {
	for i := 0; i < MarjapussiHandSize; i++ {
		for j := 0; j < MarjapussiPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % MarjapussiPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	g.pussi = make([]*Card, 0, MarjapussiPussiSize)
	for i := 0; i < MarjapussiPussiSize; i++ {
		if c := g.trumpCards.DrawCard(); c != nil {
			g.pussi = append(g.pussi, c)
		}
	}
}

// marjapussiPlayerTeam 席番号からチーム番号 (0 または 1) を返す。
// 席 0 と 2 がチーム 0、席 1 と 3 がチーム 1。
func marjapussiPlayerTeam(playerIdx int) int {
	return playerIdx % MarjapussiTeamCnt
}

// MarjapussiMarriageOption は宣言できる結婚 1 件 (スートとその結婚点)。
type MarjapussiMarriageOption struct {
	// Suit は K と Q を揃えているスート。
	Suit int
	// Points はそのスートの結婚点 (宣言時点の切り札なら 40、それ以外なら 20)。
	Points int
}

// GetMarriageOptions は playerIdx がいま宣言できる結婚を marjapussiSuits() の順で返す。
// 宣言には同スートの K と Q の両方が要る。範囲外の添字では nil を返す。
func (g *Marjapussi) GetMarriageOptions(playerIdx int) []MarjapussiMarriageOption {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	var opts []MarjapussiMarriageOption
	for _, suit := range marjapussiSuits() {
		if g.playerHasCard(playerIdx, suit, 13) && g.playerHasCard(playerIdx, suit, 12) {
			opts = append(opts, MarjapussiMarriageOption{Suit: suit, Points: g.marriagePoints(suit)})
		}
	}
	return opts
}

// marriagePoints 宣言時点の切り札スートと同じなら 40、違うなら 20 を返す。
func (g *Marjapussi) marriagePoints(suit int) int {
	return marjapussiMarriagePoints(g.trumpSuit, suit)
}

// --- Play ---

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Marjapussi) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MarjapussiPhasePlay {
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
	g.maybeDeclareMarriage(g.currentPlayerIdx, card)
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *Marjapussi) CpuPlay() {
	if g.gameEndFlag || g.phase != MarjapussiPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	card := g.players[idx].GetCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら GetCard(0) も RemoveCard(0) も nil を返す。ここは RemoveCard
	// より前に card を読むので、ガードも前に置かないと maybeDeclareMarriage が
	// nil を触って HTTP ハンドラごと落ちる (#4606)。
	if card == nil {
		return
	}
	g.maybeDeclareMarriage(idx, card)
	played := g.players[idx].RemoveCard(cardIdx)
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// maybeDeclareMarriage リード時に K/Q を出し同スートの相棒を持っていれば結婚を宣言し
// そのスートを切り札に設定してチームに結婚点を加算する。
// 宣言時点の切り札スートと同じなら 40 点、違うなら 20 点 (そのスートが新しい切り札になる)。
func (g *Marjapussi) maybeDeclareMarriage(playerIdx int, card *Card) {
	if len(g.currentTrick) != 0 {
		return // リード時のみ宣言可能
	}
	v := card.GetValue()
	if v != 13 && v != 12 {
		return
	}
	suit := card.GetDesign()
	partner := 12
	if v == 12 {
		partner = 13
	}
	if !g.playerHasCard(playerIdx, suit, partner) {
		return
	}
	pts := g.marriagePoints(suit)
	g.trumpSuit = suit
	teamIdx := marjapussiPlayerTeam(playerIdx)
	g.roundMarriage[teamIdx] += pts
	g.appendLog(playerIdx, "marriage",
		fmt.Sprintf("%s declares a %s marriage (+%d, trump=%s)",
			playerName(g.players, playerIdx), marjapussiSuitName(suit), pts, marjapussiSuitName(suit)), nil)
}

// playCard カードをプレイする共通処理。
func (g *Marjapussi) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == MarjapussiPlayerCnt {
		g.phase = MarjapussiPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % MarjapussiPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Marjapussi) ResolveTrick() {
	if g.phase != MarjapussiPhaseTrickEnd || len(g.currentTrick) != MarjapussiPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	winnerTeam := marjapussiPlayerTeam(winnerIdx)
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += marjapussiCardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.roundCardPts[winnerTeam] += pts
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d)", playerName(g.players, winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= MarjapussiTrickCount {
		g.lastTrickWinner = winnerIdx
		// 8 トリック目 (最終トリック) の勝者チームが pussi 4 枚を獲得し、その札点を加算
		pussiPts := 0
		for _, c := range g.pussi {
			pussiPts += marjapussiCardPoints(c)
		}
		g.roundCardPts[winnerTeam] += pussiPts
		g.appendLog(winnerIdx, "pussi_win",
			fmt.Sprintf("team %d wins the pussi (+%d)", winnerTeam, pussiPts), g.pussi)
		g.phase = MarjapussiPhaseRoundEnd
	} else {
		g.phase = MarjapussiPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Marjapussi) NextTrick() {
	if g.phase != MarjapussiPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = MarjapussiPhasePlay
}

// ScoreRound ラウンド結果を判定し、累積チーム得点へ加算してマッチ終了を判定する。
func (g *Marjapussi) ScoreRound() {
	if g.phase != MarjapussiPhaseRoundEnd {
		return
	}
	for t := 0; t < MarjapussiTeamCnt; t++ {
		g.teamScores[t] += g.roundCardPts[t] + g.roundMarriage[t]
	}
	pussiWinnerTeam := -1
	if g.lastTrickWinner >= 0 {
		pussiWinnerTeam = marjapussiPlayerTeam(g.lastTrickWinner)
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d scored: team 0=%d (+%d), team 1=%d (+%d) (pussi won by team %d)",
			g.roundNumber,
			g.teamScores[0], g.roundCardPts[0]+g.roundMarriage[0],
			g.teamScores[1], g.roundCardPts[1]+g.roundMarriage[1],
			pussiWinnerTeam), nil)
	g.checkGameEnd()
}

// checkGameEnd 目標点 (既定 500 点) 到達でマッチ終了を判定する。
func (g *Marjapussi) checkGameEnd() {
	limit := g.config.pointLimit()
	t0 := g.teamScores[0]
	t1 := g.teamScores[1]
	if t0 >= limit || t1 >= limit {
		g.gameEndFlag = true
		g.phase = MarjapussiPhaseGameEnd
		if t0 > t1 {
			g.winnerTeam = 0
			g.winnerPlayer = 0
		} else if t1 > t0 {
			g.winnerTeam = 1
			g.winnerPlayer = 1
		} else {
			wt := 0
			if g.lastTrickWinner >= 0 {
				wt = marjapussiPlayerTeam(g.lastTrickWinner)
			}
			g.winnerTeam = wt
			g.winnerPlayer = wt
		}
		g.appendLog(-1, "game_end", fmt.Sprintf("team %d wins the match!", g.winnerTeam), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay フォロー規則を検証する。
//
// 裁定 (issue で「実装時に確定」とされたフォロー規則の決定事項):
//  1. マストフォロー: リードスートを持っていれば必ずリードスートを出さなければならない。
//  2. 切り札強制: リードスートを持っておらず、かつ切り札スートが決定済み (trumpSuit != 0) で
//     切り札を持っている場合は、必ず切り札を出さなければならない。
//  3. 自由放棄: リードスートも切り札も持っていない場合 (または切り札が未決定の場合)、任意の札を捨てられる。
//
// ※オーバートランプ義務 (場の最高切り札を上回る義務) は採用しない。
func (g *Marjapussi) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	hasLeadSuit := g.playerHasSuit(playerIdx, leadSuit)
	if hasLeadSuit {
		if card.GetDesign() != leadSuit {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
		return nil
	}
	// リードスートを持っていない場合
	if g.trumpSuit != 0 && g.playerHasSuit(playerIdx, g.trumpSuit) {
		if card.GetDesign() != g.trumpSuit {
			return NewDomainError(ErrInvalidPlay, "切り札を出してください")
		}
		return nil
	}
	// リードスートも切り札も持っていない (または切り札未決定): 任意の札を出せる
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Marjapussi) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(g.players[playerIdx], design)
}

// playerHasCard プレイヤーが指定スート・ランクの札を持っているか。
func (g *Marjapussi) playerHasCard(playerIdx, design, value int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && c.GetValue() == value {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *Marjapussi) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.marjapussiRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if g.trumpSuit != 0 && tc.Card.GetDesign() == g.trumpSuit {
			// 切り札は常に評価対象。
		} else if tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.marjapussiRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// marjapussiRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Marjapussi) marjapussiRank(card *Card) int {
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		return 100 + marjapussiStrength(card.GetValue())
	}
	return marjapussiStrength(card.GetValue())
}

// marjapussiStrength カード強度。A > 10 > K > Q > J > 9 > 8 > 7 > 6。
func marjapussiStrength(value int) int {
	switch value {
	case 1: // Ace
		return 9
	case 10: // Ten
		return 8
	case 13: // King
		return 7
	case 12: // Queen
		return 6
	case 11: // Jack
		return 5
	case 9:
		return 4
	case 8:
		return 3
	case 7:
		return 2
	case 6:
		return 1
	default:
		return 0
	}
}

// CardPoint カードポイント (A=11, 10=10, K=4, Q=3, J=2, 9..6=0)。
func (g *Marjapussi) CardPoint(card *Card) int {
	return marjapussiCardPoints(card)
}

// marjapussiCardPoints カードポイント。A=11, 10=10, K=4, Q=3, J=2, 9..6=0。
func marjapussiCardPoints(card *Card) int {
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

// marjapussiMarriagePoints 結婚点 (切り札スートなら 40、それ以外なら 20)。
func marjapussiMarriagePoints(currentTrump, suit int) int {
	if currentTrump != 0 && currentTrump == suit {
		return 40
	}
	return 20
}

// marjapussiSuits スート一覧を返す。
func marjapussiSuits() []int {
	return []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
}

// marjapussiSuitName スート表示名 (英語) を返す。
func marjapussiSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "Spades"
	case CardDesignClover:
		return "Clubs"
	case CardDesignHeart:
		return "Hearts"
	case CardDesignDiamond:
		return "Diamonds"
	default:
		return "None"
	}
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Marjapussi) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Marjapussi) sortAllHands() {
	for _, p := range g.players {
		marjapussiSortHand(p)
	}
}

// marjapussiSortHand 手札をスート→強さ順にソートする。
func marjapussiSortHand(p *MarjapussiPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return marjapussiStrength(cards[i].GetValue()) > marjapussiStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Marjapussi) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Marjapussi) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.marjapussiRank(g.currentTrick[idx].Card)
}

// --- CPU AI (play) ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Marjapussi) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == MarjapussiCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点を意識した戦略プレイ。リード時に結婚可能なら K を優先する。
func (g *Marjapussi) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		// 結婚可能なスートがあれば K をリードして宣言する。
		if idx := g.cpuMarriageLead(playerIdx, valid); idx >= 0 {
			return idx
		}
		return pickLowest(player, valid, func(c *Card) int {
			return marjapussiCardPoints(c)*100 + g.marjapussiRank(c)
		})
	}

	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	partnerIdx := (playerIdx + 2) % MarjapussiPlayerCnt
	isPartnerWinning := (winnerIdx == partnerIdx)

	// 味方が勝っている場合: 味方に勝ってしまう札を避ける
	if isPartnerWinning {
		return pickLowest(player, valid, func(c *Card) int {
			rank := g.marjapussiRank(c)
			pts := marjapussiCardPoints(c)
			if rank > topRank {
				// 味方より強い札を出してしまう場合はペナルティ (重ね勝ち防止)
				return 10000 + rank
			}
			// 味方が勝っているので点数札を投げ込む (pts が高いほど優先)
			return -pts*10 + rank
		})
	}

	// 相手が勝っている場合: 勝てる最小の札を探す
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += marjapussiCardPoints(tc.Card)
	}
	winners := filterIndices(valid, func(idx int) bool { return g.marjapussiRank(player.GetCard(idx)) > topRank })
	if len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.marjapussiRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int {
		return marjapussiCardPoints(c)*100 + g.marjapussiRank(c)
	})
}

// cpuMarriageLead 結婚を宣言できる K のインデックスを返す (-1=なし)。
func (g *Marjapussi) cpuMarriageLead(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	best := -1
	bestPts := 0
	for _, idx := range valid {
		c := player.GetCard(idx)
		if c.GetValue() != 13 && c.GetValue() != 12 {
			continue
		}
		suit := c.GetDesign()
		partnerVal := 12
		if c.GetValue() == 12 {
			partnerVal = 13
		}
		if !g.playerHasCard(playerIdx, suit, partnerVal) {
			continue
		}
		pts := g.marriagePoints(suit)
		if pts > bestPts || (pts == bestPts && c.GetValue() == 13) {
			bestPts = pts
			best = idx
		}
	}
	return best
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Marjapussi) GetHint() *MarjapussiHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	if g.phase != MarjapussiPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &MarjapussiHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Marjapussi) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if card.GetValue() == 13 && g.playerHasCard(playerIdx, card.GetDesign(), 12) {
			return "lead_marriage"
		}
		return "lead_low"
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.marjapussiRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Marjapussi) GetPhase() MarjapussiPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Marjapussi) SetPhase(phase MarjapussiPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Marjapussi) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Marjapussi) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Marjapussi) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Marjapussi) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Marjapussi) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Marjapussi) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Marjapussi) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Marjapussi) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Marjapussi) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Marjapussi) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Marjapussi) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Marjapussi) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetTrumpSuit 切り札スート取得 (0=未設定)
func (g *Marjapussi) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Marjapussi) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetPussi ベリー袋 (pussi) のカード一覧を返す。
func (g *Marjapussi) GetPussi() []*Card { return g.pussi }

// SetPussi ベリー袋 (pussi) を設定する (テスト用)。
func (g *Marjapussi) SetPussi(pussi []*Card) { g.pussi = pussi }

// GetTeamScores チーム別累積点取得
func (g *Marjapussi) GetTeamScores() [MarjapussiTeamCnt]int { return g.teamScores }

// SetTeamScores チーム別累積点設定 (テスト用)
func (g *Marjapussi) SetTeamScores(s [MarjapussiTeamCnt]int) { g.teamScores = s }

// GetPlayerScores プレイヤー別累積点取得 (席 0+2 がチーム 0、席 1+3 がチーム 1)
func (g *Marjapussi) GetPlayerScores() [MarjapussiPlayerCnt]int {
	return [MarjapussiPlayerCnt]int{
		g.teamScores[0],
		g.teamScores[1],
		g.teamScores[0],
		g.teamScores[1],
	}
}

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Marjapussi) SetPlayerScores(s [MarjapussiPlayerCnt]int) {
	g.teamScores[0] = s[0]
	g.teamScores[1] = s[1]
}

// GetRoundCardPoints 現ラウンドのチーム別カード得点取得
func (g *Marjapussi) GetRoundCardPoints() [MarjapussiTeamCnt]int { return g.roundCardPts }

// SetRoundCardPoints 現ラウンドのチーム別カード得点設定 (テスト用)
func (g *Marjapussi) SetRoundCardPoints(s [MarjapussiTeamCnt]int) { g.roundCardPts = s }

// GetRoundMarriage 現ラウンドのチーム別結婚点取得
func (g *Marjapussi) GetRoundMarriage() [MarjapussiTeamCnt]int { return g.roundMarriage }

// SetRoundMarriage 現ラウンドのチーム別結婚点設定 (テスト用)
func (g *Marjapussi) SetRoundMarriage(s [MarjapussiTeamCnt]int) { g.roundMarriage = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Marjapussi) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (g *Marjapussi) SetGameEndFlag(f bool) { g.gameEndFlag = f }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Marjapussi) GetWinnerPlayer() int { return g.winnerPlayer }

// SetWinnerPlayer 勝利プレイヤー設定 (テスト用)
func (g *Marjapussi) SetWinnerPlayer(p int) { g.winnerPlayer = p }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *Marjapussi) GetWinnerTeam() int { return g.winnerTeam }

// SetWinnerTeam 勝利チーム設定 (テスト用)
func (g *Marjapussi) SetWinnerTeam(t int) { g.winnerTeam = t }

// GetPlayerCnt プレイヤー数取得
func (g *Marjapussi) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Marjapussi) GetPlayer(i int) *MarjapussiPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (プレイ) が人間か。
func (g *Marjapussi) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Marjapussi) GetConfig() MarjapussiConfig { return g.config }

// SetConfig 設定変更
func (g *Marjapussi) SetConfig(cfg MarjapussiConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Marjapussi) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != MarjapussiPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// marjapussiJSON is the JSON wire format for Marjapussi.
type marjapussiJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*MarjapussiPlayer    `json:"ps"`
	Config           MarjapussiConfig       `json:"cf"`
	Phase            MarjapussiPhase        `json:"ph"`
	RoundNumber      int                    `json:"rn"`
	TrickNumber      int                    `json:"tn"`
	CurrentPlayerIdx int                    `json:"ci"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	LeadPlayerIdx    int                    `json:"li"`
	DealerIdx        int                    `json:"di"`
	TrumpSuit        int                    `json:"ts"`
	Pussi            []*Card                `json:"pu"`
	TeamScores       [MarjapussiTeamCnt]int `json:"sc"`
	RoundCardPts     [MarjapussiTeamCnt]int `json:"rp"`
	RoundMarriage    [MarjapussiTeamCnt]int `json:"rm"`
	LastTrickWinner  int                    `json:"lt"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerPlayer     int                    `json:"wp"`
	WinnerTeam       int                    `json:"wt"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Marjapussi) MarshalJSON() ([]byte, error) {
	return json.Marshal(marjapussiJSON{
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
		Pussi:            g.pussi,
		TeamScores:       g.teamScores,
		RoundCardPts:     g.roundCardPts,
		RoundMarriage:    g.roundMarriage,
		LastTrickWinner:  g.lastTrickWinner,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// marjapussiMaxSliceLen caps slice sizes during deserialisation.
const marjapussiMaxSliceLen = 5000

// errMarjapussiOversized is the single sentinel error for oversized input arrays.
var errMarjapussiOversized = errors.New("marjapussi: input array exceeds maximum allowed size")

// errMarjapussiInvalidPlayers is returned when restored state lacks exactly MarjapussiPlayerCnt players.
var errMarjapussiInvalidPlayers = errors.New("marjapussi: invalid player count")

// errMarjapussiInvalidTrick is returned when a restored trick card or its card is nil.
var errMarjapussiInvalidTrick = errors.New("marjapussi: invalid trick card")

// errMarjapussiInvalidPussi is returned when a restored pussi card is nil.
var errMarjapussiInvalidPussi = errors.New("marjapussi: invalid pussi card")

// errMarjapussiInvalidIndex is returned when a restored index field is out of range.
var errMarjapussiInvalidIndex = errors.New("marjapussi: index field out of range")

// errMarjapussiInvalidPhase is returned when a restored phase is out of range.
var errMarjapussiInvalidPhase = errors.New("marjapussi: phase out of range")

// errMarjapussiInvalidTrump is returned when a restored trump suit is out of range.
var errMarjapussiInvalidTrump = errors.New("marjapussi: trump suit out of range")

// marjapussiInRange reports whether v is in [0, MarjapussiPlayerCnt).
func marjapussiInRange(v int) bool { return v >= 0 && v < MarjapussiPlayerCnt }

// marjapussiInRangeOrUnset reports whether v is -1 (unset) or in [0, MarjapussiPlayerCnt).
func marjapussiInRangeOrUnset(v int) bool { return v == -1 || marjapussiInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Marjapussi) UnmarshalJSON(data []byte) error {
	var j marjapussiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > marjapussiMaxSliceLen || len(j.CurrentTrick) > marjapussiMaxSliceLen ||
		len(j.ActionLog) > marjapussiMaxSliceLen || len(j.Pussi) > marjapussiMaxSliceLen {
		return errMarjapussiOversized
	}
	if len(j.Players) != MarjapussiPlayerCnt {
		return errMarjapussiInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errMarjapussiInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errMarjapussiInvalidTrick
		}
		if !marjapussiInRange(tc.PlayerIdx) {
			return errMarjapussiInvalidTrick
		}
	}
	for _, c := range j.Pussi {
		if c == nil {
			return errMarjapussiInvalidPussi
		}
	}
	// 範囲必須のインデックス (0..PlayerCnt)。
	if !marjapussiInRange(j.CurrentPlayerIdx) || !marjapussiInRange(j.DealerIdx) {
		return errMarjapussiInvalidIndex
	}
	// -1 (未設定) 許容のインデックス。
	if !marjapussiInRangeOrUnset(j.LeadPlayerIdx) ||
		!marjapussiInRangeOrUnset(j.LastTrickWinner) || !marjapussiInRangeOrUnset(j.WinnerPlayer) {
		return errMarjapussiInvalidIndex
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= MarjapussiTeamCnt {
		return errMarjapussiInvalidIndex
	}
	if int(j.Phase) < MarjapussiPhaseMin || int(j.Phase) > MarjapussiPhaseMax {
		return errMarjapussiInvalidPhase
	}
	// trumpSuit: 0=未設定 許容、それ以外は [Spade, Diamond]。
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return errMarjapussiInvalidTrump
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newMarjapussiDeck()
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
	g.pussi = j.Pussi
	if g.pussi == nil {
		g.pussi = make([]*Card, 0)
	}
	g.teamScores = j.TeamScores
	g.roundCardPts = j.RoundCardPts
	g.roundMarriage = j.RoundMarriage
	g.lastTrickWinner = j.LastTrickWinner
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
