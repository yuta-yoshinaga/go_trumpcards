//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// HokmPhase ホクムのゲームフェーズ
type HokmPhase int

// Hokm のフェーズ定数
const (
	// HokmPhaseTrump 親 (Hakem) が最初の 5 枚だけを見て切り札を宣言する
	HokmPhaseTrump HokmPhase = iota
	// HokmPhasePlay プレイ中
	HokmPhasePlay
	// HokmPhaseHandEnd ハンド終了
	HokmPhaseHandEnd
	// HokmPhaseGameEnd ゲーム終了
	HokmPhaseGameEnd
)

// HokmPlayerCnt プレイヤー数（4 人固定・2 対 2）
const HokmPlayerCnt = 4

// HokmTeamCnt チーム数
const HokmTeamCnt = 2

// HokmPeekSize 切り札を宣言する前に親へ配る枚数
const HokmPeekSize = 5

// HokmHandSize 各プレイヤーの最終的な手札枚数
const HokmHandSize = 13

// HokmTricksToWin ハンドを制するのに必要なトリック数
//
// **13 トリックを消化しない。** 7 つ取った時点でそのハンドは終わる。
const HokmTricksToWin = 7

// HokmKotPoints 相手を 1 トリックも取らせずに勝ったとき (Kot) のハンド勝ち点
const HokmKotPoints = 2

// HokmHandPoints 通常のハンド勝ち点
const HokmHandPoints = 1

// hokmMaxSliceLen caps slice sizes during deserialisation.
const hokmMaxSliceLen = 1000

// Hokm ホクム ゲームクラス。
//
// イランで最も広く遊ばれている切り札トリックテイキング。4 人 2 対 2（向かい
// 合う席が味方）、52 枚を 13 枚ずつ。
//
// **13 トリックを消化しない。** どちらかのチームが **7 トリック**を取った
// 時点でそのハンドは即座に終わり、残りの札は打たれない。既存の実装はどれも
// 全トリックを消化してから採点するので、この「早取り」がこのゲーム固有の形。
//
// 親 (**Hakem**) は最初の **5 枚だけ**を見て切り札を宣言し、そのあと残りが
// 配られて全員 13 枚になる。宣言に使えるのは 5 枚ぶんの情報だけで、13 枚
// 揃ってからではない——それが賭けになっている。
//
// **Kot**（相手を 1 トリックも取らせずに 7 取る）は 2 ハンド勝ち点。通常は 1。
// 先に規定ハンド（既定 7）を取ったチームの勝ち。
//
// **親は勝ち続けるかぎり交代しない。** 負けたときだけ左隣へ移る。切り札を
// 選べる立場が「勝っているあいだ続く」ので、Kot の重みが効いてくる。
type Hokm struct {
	trumpCards *TrumpCards
	players    []*HokmPlayer
	config     HokmConfig

	phase       HokmPhase
	handNumber  int
	trickNumber int
	trumpSuit   int
	// hakemIdx は切り札を宣言する親。負けたときだけ交代する。
	hakemIdx int

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int

	scores [HokmTeamCnt]int
	// lastHandKot は直前のハンドが Kot だったか（表示用）。
	lastHandKot bool
	// lastHandWinner は直前のハンドを制したチーム (-1: まだ無い)。
	lastHandWinner int

	gameEndFlag bool
	winnerTeam  int

	actionLogBase
}

// NewHokm コンストラクタ
func NewHokm(trumpCards *TrumpCards, players []*HokmPlayer, config HokmConfig) *Hokm {
	return &Hokm{trumpCards: trumpCards, players: players, config: config, winnerTeam: -1, lastHandWinner: -1}
}

// NewDefaultHokm 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultHokm() *Hokm {
	players := make([]*HokmPlayer, 0, HokmPlayerCnt)
	for i := range HokmPlayerCnt {
		players = append(players, NewHokmPlayer(i == 0))
	}
	return NewHokm(NewTrumpCards(0), players, DefaultHokmConfig())
}

// HokmTeamOf 席のチーム番号。**向かい合う席が味方。**
func HokmTeamOf(playerIdx int) int { return playerIdx % HokmTeamCnt }

// Reset ゲーム全体を初期化する
func (h *Hokm) Reset() {
	h.handNumber = 1
	h.hakemIdx = 0
	h.gameEndFlag = false
	h.winnerTeam = -1
	h.lastHandWinner = -1
	h.lastHandKot = false
	h.scores = [HokmTeamCnt]int{}
	h.actionLog = nil
	for _, p := range h.players {
		p.ResetGame()
	}
	h.dealHand()
}

// dealHand 親に 5 枚だけ配り、切り札の宣言から始める
func (h *Hokm) dealHand() {
	h.phase = HokmPhaseTrump
	h.trickNumber = 0
	h.currentTrick = nil
	h.trumpSuit = 0
	for _, p := range h.players {
		p.ResetRound()
	}

	h.trumpCards = NewTrumpCards(0)
	h.trumpCards.Shuffle()
	// **親だけに 5 枚。** 宣言に使えるのはこの 5 枚ぶんの情報だけで、
	// 13 枚揃ってからではない。
	for range HokmPeekSize {
		if c := h.trumpCards.DrawCard(); c != nil {
			h.players[h.hakemIdx].AddCard(c)
		}
	}
	h.sortHand(h.hakemIdx)
	h.currentPlayerIdx = h.hakemIdx
	h.leadPlayerIdx = h.hakemIdx
	h.appendLog(-1, "deal", fmt.Sprintf("ハンド%d を開始（親: %d）", h.handNumber, h.hakemIdx), nil)
}

// dealRemaining 宣言後に残りを配り、全員 13 枚にそろえる
func (h *Hokm) dealRemaining() {
	for i := range HokmPlayerCnt {
		idx := (h.hakemIdx + i) % HokmPlayerCnt
		for h.players[idx].GetCardsSize() < HokmHandSize {
			c := h.trumpCards.DrawCard()
			if c == nil {
				break
			}
			h.players[idx].AddCard(c)
		}
		h.sortHand(idx)
	}
}

// sortHand 手札をスート・ランク順に並べ替える
func (h *Hokm) sortHand(idx int) {
	sortPlayerHand(h.players[idx], func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return hokmRank(ci) < hokmRank(cj)
	})
}

// hokmRank 札の強さ。A が最強、以下 K,Q,J,10..2。
func hokmRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// PlayerDeclareTrump 人間の親が切り札スートを宣言する
func (h *Hokm) PlayerDeclareTrump(suit int) error {
	if h.gameEndFlag {
		return errors.New("game has ended")
	}
	if h.phase != HokmPhaseTrump {
		return errors.New("not the trump-declaration phase")
	}
	if h.hakemIdx != 0 {
		return errors.New("only the hakem declares the trump suit")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", suit)
	}
	h.acceptTrump(suit)
	return nil
}

// CpuDeclareTrump CPU の親が切り札スートを宣言する
func (h *Hokm) CpuDeclareTrump() {
	if h.gameEndFlag || h.phase != HokmPhaseTrump || h.hakemIdx == 0 {
		return
	}
	h.acceptTrump(h.longestSuit(h.hakemIdx))
}

// acceptTrump 切り札を確定させ、残りを配ってプレイに入る
func (h *Hokm) acceptTrump(suit int) {
	h.trumpSuit = suit
	h.appendLog(h.hakemIdx, "trump", fmt.Sprintf("切り札を %d に宣言", suit), nil)
	h.dealRemaining()
	h.phase = HokmPhasePlay
	// リードは親から。
	h.leadPlayerIdx = h.hakemIdx
	h.currentPlayerIdx = h.hakemIdx
}

// longestSuit いちばん枚数の多いスート。同数なら強い札の多いほう。
func (h *Hokm) longestSuit(idx int) int {
	p := h.players[idx]
	counts := map[int]int{}
	strength := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		counts[c.GetDesign()]++
		strength[c.GetDesign()] += hokmRank(c)
	}
	best, bestN, bestS := CardDesignSpade, -1, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestN || (counts[suit] == bestN && strength[suit] > bestS) {
			best, bestN, bestS = suit, counts[suit], strength[suit]
		}
	}
	return best
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (h *Hokm) PlayerPlay(cardIndex int) error {
	if h.gameEndFlag {
		return errors.New("game has ended")
	}
	if h.phase != HokmPhasePlay {
		return errors.New("not the play phase")
	}
	if h.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return h.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (h *Hokm) CpuPlay() {
	if h.gameEndFlag || h.phase != HokmPhasePlay || h.currentPlayerIdx == 0 {
		return
	}
	_ = h.play(h.currentPlayerIdx, h.chooseCpuCard(h.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (h *Hokm) play(playerIdx, cardIndex int) error {
	p := h.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !h.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	h.currentTrick = append(h.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	h.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(h.currentTrick) < HokmPlayerCnt {
		h.currentPlayerIdx = (playerIdx + 1) % HokmPlayerCnt
		return nil
	}
	h.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか
func (h *Hokm) canPlay(playerIdx int, card *Card) bool {
	if len(h.currentTrick) == 0 {
		return true
	}
	leadSuit := h.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := h.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (h *Hokm) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(h.players) {
		return nil
	}
	p := h.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if h.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決し、7 トリックに達していればハンドを終える
func (h *Hokm) resolveTrick() {
	winner := h.trickWinner()
	cards := make([]*Card, 0, len(h.currentTrick))
	for _, tc := range h.currentTrick {
		cards = append(cards, tc.Card)
	}
	h.players[winner].AddTrick(cards)

	h.trickNumber++
	h.currentTrick = nil
	h.leadPlayerIdx = winner
	h.currentPlayerIdx = winner

	// **7 トリック取った時点でハンドは終わる。** 残りの札は打たれない。
	if team, ok := h.teamReachedTarget(); ok {
		h.finishHand(team)
		return
	}
	// 保険: 13 トリック消化しても決着しない配りは存在しないが、
	// 手札が尽きたら終える（7+7 = 14 > 13 なので必ず先に上で終わる）。
	if h.trickNumber >= HokmHandSize {
		h.finishHand(h.leadingTeam())
	}
}

// teamReachedTarget 7 トリックに達したチームを返す
func (h *Hokm) teamReachedTarget() (int, bool) {
	for team := range HokmTeamCnt {
		if h.TeamTricks(team) >= HokmTricksToWin {
			return team, true
		}
	}
	return -1, false
}

// leadingTeam 現時点でトリック数の多いチーム（同数なら 0）
func (h *Hokm) leadingTeam() int {
	if h.TeamTricks(1) > h.TeamTricks(0) {
		return 1
	}
	return 0
}

// TeamTricks チームの獲得トリック数
func (h *Hokm) TeamTricks(team int) int {
	if team < 0 || team >= HokmTeamCnt {
		return 0
	}
	n := 0
	for i, p := range h.players {
		if HokmTeamOf(i) == team {
			n += p.GetTrickCount()
		}
	}
	return n
}

// finishHand ハンドの勝ち点を確定させる。
//
// **相手が 1 トリックも取れていなければ Kot で 2 点。** 通常は 1 点。
func (h *Hokm) finishHand(winnerTeam int) {
	loser := 1 - winnerTeam
	kot := h.TeamTricks(loser) == 0
	points := HokmHandPoints
	if kot {
		points = HokmKotPoints
	}
	h.scores[winnerTeam] += points
	h.lastHandKot = kot
	h.lastHandWinner = winnerTeam

	if kot {
		h.appendLog(-1, "kot", fmt.Sprintf("チーム%d が Kot（相手 0 トリック）で +%d", winnerTeam, points), nil)
	} else {
		h.appendLog(-1, "hand", fmt.Sprintf("チーム%d がハンドを制して +%d", winnerTeam, points), nil)
	}

	// **親は負けたときだけ交代する。** 勝っているあいだは切り札を選び続ける。
	if HokmTeamOf(h.hakemIdx) != winnerTeam {
		h.hakemIdx = (h.hakemIdx + 1) % HokmPlayerCnt
		h.appendLog(-1, "hakem", fmt.Sprintf("親が %d へ移った", h.hakemIdx), nil)
	}

	if h.scores[winnerTeam] >= h.config.Target {
		h.finishGame()
		return
	}
	h.phase = HokmPhaseHandEnd
}

// NextHand 次のハンドを開始する
func (h *Hokm) NextHand() {
	if h.gameEndFlag || h.phase != HokmPhaseHandEnd {
		return
	}
	h.handNumber++
	h.dealHand()
}

// finishGame 規定ハンドに達したチームの勝ち
func (h *Hokm) finishGame() {
	h.phase = HokmPhaseGameEnd
	h.gameEndFlag = true
	switch {
	case h.scores[0] > h.scores[1]:
		h.winnerTeam = 0
	case h.scores[1] > h.scores[0]:
		h.winnerTeam = 1
	default:
		h.winnerTeam = -1
	}
	h.appendLog(-1, "result", fmt.Sprintf("最終得点 %d - %d", h.scores[0], h.scores[1]), nil)
}

// trickWinner 現在のトリックの勝者
func (h *Hokm) trickWinner() int {
	if len(h.currentTrick) == 0 {
		return h.leadPlayerIdx
	}
	leadSuit := h.currentTrick[0].Card.GetDesign()
	bestIdx, best := h.currentTrick[0].PlayerIdx, h.currentTrick[0].Card
	for _, tc := range h.currentTrick[1:] {
		if h.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// beats challenger が currentBest に勝つか
func (h *Hokm) beats(challenger, currentBest *Card, leadSuit int) bool {
	cTrump := challenger.GetDesign() == h.trumpSuit
	bTrump := currentBest.GetDesign() == h.trumpSuit
	if cTrump != bTrump {
		return cTrump
	}
	if challenger.GetDesign() != currentBest.GetDesign() {
		return challenger.GetDesign() == leadSuit
	}
	return hokmRank(challenger) > hokmRank(currentBest)
}

// chooseCpuCard CPU の手。味方が勝っていれば安く、そうでなければ取りに行く。
func (h *Hokm) chooseCpuCard(playerIdx int) int {
	valid := h.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := h.players[playerIdx]

	if len(h.currentTrick) == 0 {
		return h.pickExtreme(p, valid, true)
	}
	if h.partnerIsWinning(playerIdx) {
		return h.pickExtreme(p, valid, false)
	}
	if idx, ok := h.pickCheapestWinner(p, valid); ok {
		return idx
	}
	return h.pickExtreme(p, valid, false)
}

// partnerIsWinning 現時点で味方がトリックを取っているか
func (h *Hokm) partnerIsWinning(playerIdx int) bool {
	if len(h.currentTrick) == 0 {
		return false
	}
	leadSuit := h.currentTrick[0].Card.GetDesign()
	best, bestIdx := h.currentTrick[0].Card, h.currentTrick[0].PlayerIdx
	for _, tc := range h.currentTrick[1:] {
		if h.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx != playerIdx && HokmTeamOf(bestIdx) == HokmTeamOf(playerIdx)
}

// pickExtreme valid のうち最強 (high) または最弱の札を選ぶ
func (h *Hokm) pickExtreme(p *HokmPlayer, valid []int, high bool) int {
	bestIdx, bestRank := valid[0], hokmRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := hokmRank(p.GetCard(i))
		if (high && r > bestRank) || (!high && r < bestRank) {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx
}

// pickCheapestWinner トリックを取れる札のうち一番弱いもの
func (h *Hokm) pickCheapestWinner(p *HokmPlayer, valid []int) (int, bool) {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !h.wouldWin(c) {
			continue
		}
		if r := hokmRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx, bestIdx >= 0
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (h *Hokm) wouldWin(c *Card) bool {
	if c == nil || len(h.currentTrick) == 0 {
		return true
	}
	leadSuit := h.currentTrick[0].Card.GetDesign()
	best := h.currentTrick[0].Card
	for _, tc := range h.currentTrick[1:] {
		if h.beats(tc.Card, best, leadSuit) {
			best = tc.Card
		}
	}
	return h.beats(c, best, leadSuit)
}

// HokmHint ヒント情報
type HokmHint struct {
	// CardIndex 推奨する手札のインデックス（宣言中は nil）
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
	// Suit 宣言を勧める切り札スート（それ以外は 0）
	Suit int
}

// GetHint 人間プレイヤーへの推奨手を返す
func (h *Hokm) GetHint() *HokmHint {
	if h.gameEndFlag {
		return nil
	}
	if h.phase == HokmPhaseTrump && h.hakemIdx == 0 {
		return &HokmHint{Reason: "hokmDeclareTrump", Suit: h.longestSuit(0)}
	}
	if !h.IsHumanTurn() || h.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := h.chooseCpuCard(0)
	reason := "hokmWinTrick"
	if h.partnerIsWinning(0) {
		reason = "hokmSaveCards"
	}
	return &HokmHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (h *Hokm) GetPhase() HokmPhase { return h.phase }

// GetConfig 現在の設定
func (h *Hokm) GetConfig() HokmConfig { return h.config }

// SetConfig 設定を差し替える
func (h *Hokm) SetConfig(c HokmConfig) { h.config = c }

// GetHandNumber 現在のハンド番号（1 起点）
func (h *Hokm) GetHandNumber() int { return h.handNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (h *Hokm) GetTrickNumber() int { return h.trickNumber }

// GetTrumpSuit 切り札のスート（宣言前は 0）
func (h *Hokm) GetTrumpSuit() int { return h.trumpSuit }

// GetHakemIdx 親 (Hakem) のインデックス
func (h *Hokm) GetHakemIdx() int { return h.hakemIdx }

// GetScore チームのハンド勝ち点
func (h *Hokm) GetScore(team int) int {
	if team < 0 || team >= HokmTeamCnt {
		return 0
	}
	return h.scores[team]
}

// SetScoreForTestUse チームのハンド勝ち点を設定する（復元・テスト用）
func (h *Hokm) SetScoreForTestUse(team, n int) {
	if team >= 0 && team < HokmTeamCnt {
		h.scores[team] = n
	}
}

// GetLastHandKot 直前のハンドが Kot だったか
func (h *Hokm) GetLastHandKot() bool { return h.lastHandKot }

// GetLastHandWinner 直前のハンドを制したチーム (-1: まだ無い)
func (h *Hokm) GetLastHandWinner() int { return h.lastHandWinner }

// GetCurrentTrick 現在のトリック
func (h *Hokm) GetCurrentTrick() []*TrickCard { return h.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (h *Hokm) GetCurrentPlayerIdx() int { return h.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (h *Hokm) GetLeadPlayerIdx() int { return h.leadPlayerIdx }

// GetPlayerCnt プレイヤー数
func (h *Hokm) GetPlayerCnt() int { return len(h.players) }

// GetPlayer 指定インデックスのプレイヤー
func (h *Hokm) GetPlayer(i int) *HokmPlayer {
	if i < 0 || i >= len(h.players) {
		return nil
	}
	return h.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (h *Hokm) GetGameEndFlag() bool { return h.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1: 未確定/同点)
func (h *Hokm) GetWinnerTeam() int { return h.winnerTeam }

// IsHumanTurn 人間の手番か
func (h *Hokm) IsHumanTurn() bool {
	return !h.gameEndFlag && h.phase == HokmPhasePlay && h.currentPlayerIdx == 0
}

// IsHumanTrumpTurn 人間が切り札を宣言する番か
func (h *Hokm) IsHumanTrumpTurn() bool {
	return !h.gameEndFlag && h.phase == HokmPhaseTrump && h.hakemIdx == 0
}

// GiveUp 投了する
func (h *Hokm) GiveUp() {
	if h.gameEndFlag {
		return
	}
	h.phase = HokmPhaseGameEnd
	h.gameEndFlag = true
	h.winnerTeam = 1
	h.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (h *Hokm) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	h.appendLogAt(h.trickNumber, playerIdx, actionType, detail, cards)
}

// hokmJSON is the KV snapshot format for Hokm.
type hokmJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*HokmPlayer     `json:"pl"`
	Config           HokmConfig        `json:"cf"`
	Phase            HokmPhase         `json:"ph"`
	HandNumber       int               `json:"hn"`
	TrickNumber      int               `json:"tn"`
	TrumpSuit        int               `json:"ts"`
	HakemIdx         int               `json:"hk"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	CurrentPlayerIdx int               `json:"cp"`
	LeadPlayerIdx    int               `json:"lp"`
	Scores           [HokmTeamCnt]int  `json:"sc"`
	LastHandKot      bool              `json:"lk"`
	LastHandWinner   int               `json:"lw"`
	GameEndFlag      bool              `json:"ge"`
	WinnerTeam       int               `json:"wt"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (h *Hokm) MarshalJSON() ([]byte, error) {
	return json.Marshal(&hokmJSON{
		TrumpCards:       h.trumpCards,
		Players:          h.players,
		Config:           h.config,
		Phase:            h.phase,
		HandNumber:       h.handNumber,
		TrickNumber:      h.trickNumber,
		TrumpSuit:        h.trumpSuit,
		HakemIdx:         h.hakemIdx,
		CurrentTrick:     h.currentTrick,
		CurrentPlayerIdx: h.currentPlayerIdx,
		LeadPlayerIdx:    h.leadPlayerIdx,
		Scores:           h.scores,
		LastHandKot:      h.lastHandKot,
		LastHandWinner:   h.lastHandWinner,
		GameEndFlag:      h.gameEndFlag,
		WinnerTeam:       h.winnerTeam,
		ActionLog:        h.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。値域を検証する。
func (h *Hokm) UnmarshalJSON(data []byte) error {
	var j hokmJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < HokmPhaseTrump || j.Phase > HokmPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **切り札はフェーズと整合していなければならない。** 宣言前はまだ 0、
	// 宣言後は実在するスート。素通しすると beats() がどの札も切り札と
	// 見なさなくなり、トリックの勝敗が黙って変わる (#5302 / #5303 と同じ穴)。
	if j.Phase == HokmPhaseTrump {
		if j.TrumpSuit != 0 {
			return fmt.Errorf("trump suit %d before it was declared", j.TrumpSuit)
		}
	} else if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
	}
	if j.TrickNumber < 0 || j.TrickNumber > HokmHandSize {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.HandNumber < 1 {
		return fmt.Errorf("invalid hand number: %d", j.HandNumber)
	}
	if len(j.ActionLog) > hokmMaxSliceLen {
		return errors.New("hokm: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > HokmPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"hakem":          j.HakemIdx,
	} {
		if idx < 0 || idx >= HokmPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	for name, team := range map[string]int{
		"winner team":      j.WinnerTeam,
		"last hand winner": j.LastHandWinner,
	} {
		if team < -1 || team >= HokmTeamCnt {
			return fmt.Errorf("invalid %s: %d", name, team)
		}
	}
	if j.TrumpCards != nil {
		h.trumpCards = j.TrumpCards
	}
	if len(j.Players) == HokmPlayerCnt {
		h.players = j.Players
	}
	h.config = j.Config
	h.phase = j.Phase
	h.handNumber = j.HandNumber
	h.trickNumber = j.TrickNumber
	h.trumpSuit = j.TrumpSuit
	h.hakemIdx = j.HakemIdx
	h.currentTrick = j.CurrentTrick
	h.currentPlayerIdx = j.CurrentPlayerIdx
	h.leadPlayerIdx = j.LeadPlayerIdx
	h.scores = j.Scores
	h.lastHandKot = j.LastHandKot
	h.lastHandWinner = j.LastHandWinner
	h.gameEndFlag = j.GameEndFlag
	h.winnerTeam = j.WinnerTeam
	h.actionLog = j.ActionLog
	return nil
}
