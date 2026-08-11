//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GermanWhistPhase ジャーマンホイストのゲームフェーズ
type GermanWhistPhase int

// GermanWhistのフェーズ定数
const (
	// GermanWhistPhaseDraw 前半（山札を賭けて争う 13 トリック）
	GermanWhistPhaseDraw GermanWhistPhase = iota
	// GermanWhistPhaseScoring 後半（得点になる 13 トリック）
	GermanWhistPhaseScoring
	// GermanWhistPhaseGameEnd ゲーム終了
	GermanWhistPhaseGameEnd
)

// GermanWhistPlayerCnt プレイヤー数（2 人固定）
const GermanWhistPlayerCnt = 2

// GermanWhistHandSize 各プレイヤーの手札枚数
const GermanWhistHandSize = 13

// GermanWhistStageTricks 各フェーズのトリック数
const GermanWhistStageTricks = 13

// GermanWhistWinTricks 勝利に必要な後半トリック数（13 の過半数）
const GermanWhistWinTricks = GermanWhistStageTricks/2 + 1

// germanWhistMaxSliceLen caps slice sizes during deserialisation.
const germanWhistMaxSliceLen = 1000

// GermanWhist ジャーマンホイスト ゲームクラス。
//
// 標準 52 枚の 2 人用ホイスト。13 枚ずつ配り、残り 26 枚を山札にして
// **一番上を表向き**にする。この表向きの札の**スートがそのハンドの切り札**に
// なる。
//
// 前半 13 トリックは**得点にならない**。各トリックの勝者が表向きの札を取り、
// 敗者はその下の伏せ札を引く。そして次の 1 枚を表向きにする。つまり前半は
// 「見えている 1 枚を賭けた争奪戦」で、勝つべきトリックと捨てるべきトリックを
// 選ぶのが読みどころになる。
//
// 山札が尽きると後半 13 トリックに入り、**ここで取ったトリックだけが得点**。
// 7 トリック取った側が勝つ。
//
// issue #5232 の仕様案とは 1 点異なり、実際の規則に合わせた:
//   - **切り札がある。** 最初に表向きにした札のスートがハンドを通じての切り札に
//     なる。issue は切り札に一切触れておらず、それではただのフォロー義務ゲームに
//     なってしまう。なお切り札は最初の 1 枚で決まり、**その後めくられる札では
//     変わらない**
type GermanWhist struct {
	trumpCards  *TrumpCards
	players     []*GermanWhistPlayer
	phase       GermanWhistPhase
	trickNumber int
	// upCard は山札の一番上に表向きで置かれた札。前半のトリックはこれを賭けて争う。
	upCard           *Card
	stock            []*Card
	trumpSuit        int
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	gameEndFlag      bool
	winnerIdx        int
	actionLogBase
}

// NewGermanWhist コンストラクタ
func NewGermanWhist(trumpCards *TrumpCards, players []*GermanWhistPlayer) *GermanWhist {
	return &GermanWhist{trumpCards: trumpCards, players: players, winnerIdx: -1}
}

// NewDefaultGermanWhist returns GermanWhist with a standard 52-card deck and
// one human against one CPU. Used as the single source of truth for CUI, Web,
// and Worker construction sites.
func NewDefaultGermanWhist() *GermanWhist {
	return NewGermanWhist(NewTrumpCards(0), []*GermanWhistPlayer{
		NewGermanWhistPlayer(true),
		NewGermanWhistPlayer(false),
	})
}

// Reset ゲームリセット
func (g *GermanWhist) Reset() {
	g.phase = GermanWhistPhaseDraw
	g.trickNumber = 0
	g.currentTrick = nil
	g.stock = nil
	g.upCard = nil
	g.trumpSuit = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil
	g.leadPlayerIdx = 0
	g.currentPlayerIdx = 0

	for _, p := range g.players {
		p.ResetGame()
	}

	g.trumpCards.Shuffle()
	for range GermanWhistHandSize {
		for i := range GermanWhistPlayerCnt {
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[i].AddCard(c)
			}
		}
	}
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.stock = append(g.stock, card)
	}

	// **表向きの 1 枚がハンド全体の切り札を決める。** 以降めくられる札で
	// 切り札は変わらない。
	g.turnUpCard()
	if g.upCard != nil {
		g.trumpSuit = g.upCard.GetDesign()
		g.appendLog(-1, "trump", fmt.Sprintf("Trump: %s", cardStr(g.upCard)), []*Card{g.upCard})
	}
	g.sortAllHands()
}

// turnUpCard 山札の一番上を表向きにする
func (g *GermanWhist) turnUpCard() {
	if len(g.stock) == 0 {
		g.upCard = nil
		return
	}
	g.upCard = g.stock[0]
	g.stock = g.stock[1:]
}

// sortAllHands 手札をスート（切り札は最後）→ ランク で並べ替える
func (g *GermanWhist) sortAllHands() {
	sortEachHand(g.players, g.sortHand)
}

// sortHand プレイヤー 1 人の手札を並べ替える
func (g *GermanWhist) sortHand(p *GermanWhistPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == g.trumpSuit
		jTrump := cj.GetDesign() == g.trumpSuit
		if iTrump != jTrump {
			return !iTrump
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return germanWhistRank(ci) < germanWhistRank(cj)
	})
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (g *GermanWhist) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return errors.New("game has ended")
	}
	if g.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return g.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (g *GermanWhist) CpuPlay() {
	if g.gameEndFlag || g.currentPlayerIdx == 0 {
		return
	}
	idx := g.chooseCpuCard(g.currentPlayerIdx)
	_ = g.play(g.currentPlayerIdx, idx)
}

// play 指定プレイヤーが 1 枚出す
func (g *GermanWhist) play(playerIdx, cardIndex int) error {
	p := g.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !g.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(g.currentTrick) < GermanWhistPlayerCnt {
		g.currentPlayerIdx = (playerIdx + 1) % GermanWhistPlayerCnt
		return nil
	}
	g.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか。リードなら何でも出せる。
func (g *GermanWhist) canPlay(playerIdx int, card *Card) bool {
	if len(g.currentTrick) == 0 {
		return true
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	// リードのスートを持っているなら、それを出さなければならない。
	p := g.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// resolveTrick トリックを解決する
func (g *GermanWhist) resolveTrick() {
	winner := g.trickWinner()
	cards := make([]*Card, 0, len(g.currentTrick))
	for _, tc := range g.currentTrick {
		cards = append(cards, tc.Card)
	}
	g.players[winner].AddTrick(cards)

	if g.phase == GermanWhistPhaseDraw {
		g.drawAfterTrick(winner)
	} else {
		// **後半のトリックだけが得点になる。**
		g.players[winner].AddScoringTrick()
	}

	g.trickNumber++
	g.currentTrick = nil
	g.leadPlayerIdx = winner
	g.currentPlayerIdx = winner
	g.appendLog(winner, "trick", fmt.Sprintf("トリック%d を獲得", g.trickNumber), cards)

	g.advancePhase()
}

// drawAfterTrick 前半の補充。**勝者が表向きの札、敗者が次の伏せ札**を引く。
//
// 敗者が引く札は伏せたままなので、勝者にも見えない。前半の駆け引きは
// 「見えている 1 枚を取りに行くか、見えない 1 枚に賭けるか」になる。
func (g *GermanWhist) drawAfterTrick(winner int) {
	loser := (winner + 1) % GermanWhistPlayerCnt
	if g.upCard != nil {
		g.players[winner].AddCard(g.upCard)
		g.upCard = nil
	}
	if len(g.stock) > 0 {
		g.players[loser].AddCard(g.stock[0])
		g.stock = g.stock[1:]
	}
	g.turnUpCard()
	g.sortAllHands()
}

// advancePhase 前半が終われば後半へ、全トリック終了ならゲーム終了へ
func (g *GermanWhist) advancePhase() {
	if g.phase == GermanWhistPhaseDraw && g.trickNumber >= GermanWhistStageTricks {
		g.phase = GermanWhistPhaseScoring
		g.appendLog(-1, "phase", "後半（得点になるトリック）開始", nil)
		return
	}
	if g.phase == GermanWhistPhaseScoring && g.trickNumber >= GermanWhistStageTricks*2 {
		g.finish()
	}
}

// finish 勝敗を決めて終了する
func (g *GermanWhist) finish() {
	g.phase = GermanWhistPhaseGameEnd
	g.gameEndFlag = true
	a, b := g.players[0].GetScoringTricks(), g.players[1].GetScoringTricks()
	switch {
	case a > b:
		g.winnerIdx = 0
	case b > a:
		g.winnerIdx = 1
	default:
		// 13 は奇数なので後半で引き分けは起きないが、復元した壊れた状態でも
		// 落ちないように -1 を残す。
		g.winnerIdx = -1
	}
	g.appendLog(-1, "result", fmt.Sprintf("後半トリック %d - %d", a, b), nil)
}

// trickWinner 現在のトリックの勝者
func (g *GermanWhist) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return g.leadPlayerIdx
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card
	for _, tc := range g.currentTrick[1:] {
		if germanWhistBeats(tc.Card, winnerCard, leadSuit, g.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// germanWhistBeats challenger が currentBest を上回るか。
// 切り札は非切り札に勝ち、同スートなら A が最強（A > K > ... > 2）。
func germanWhistBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit
	switch {
	case cIsTrump && bIsTrump:
		return germanWhistRank(challenger) > germanWhistRank(currentBest)
	case cIsTrump:
		return true
	case bIsTrump:
		return false
	}
	if challenger.GetDesign() != leadSuit {
		return false
	}
	return germanWhistRank(challenger) > germanWhistRank(currentBest)
}

// germanWhistRank ホイストの強さ。A が最強なので 1 を 14 に読み替える。
func germanWhistRank(c *Card) int {
	if c.GetValue() == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// chooseCpuCard CPU の手。
//
// 前半は**表向きの札が欲しいかどうか**で勝ちに行くか降りるかを決める。
// 後半は素直に勝ちに行く。これがこのゲームの読みどころそのものなので、
// 単に一番強い札を出すだけにはしない。
func (g *GermanWhist) chooseCpuCard(playerIdx int) int {
	p := g.players[playerIdx]
	legal := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if g.canPlay(playerIdx, p.GetCard(i)) {
			legal = append(legal, i)
		}
	}
	if len(legal) == 0 {
		return 0
	}
	if g.phase != GermanWhistPhaseDraw || g.upCardIsWorthTaking() {
		return g.pickLegal(p, legal, true)
	}
	return g.pickLegal(p, legal, false)
}

// upCardIsWorthTaking 表向きの札を取りに行く価値があるか。
// 切り札か、絵札級（10 以上）か、A なら取りに行く。
func (g *GermanWhist) upCardIsWorthTaking() bool {
	if g.upCard == nil {
		return false
	}
	if g.upCard.GetDesign() == g.trumpSuit {
		return true
	}
	return germanWhistRank(g.upCard) >= 10
}

// pickLegal 合法手のうち最も強い（strongest=true）か最も弱い札の添字を返す
func (g *GermanWhist) pickLegal(p *GermanWhistPlayer, legal []int, strongest bool) int {
	best := legal[0]
	for _, i := range legal[1:] {
		a := germanWhistCardStrength(p.GetCard(i), g.trumpSuit)
		b := germanWhistCardStrength(p.GetCard(best), g.trumpSuit)
		if (strongest && a > b) || (!strongest && a < b) {
			best = i
		}
	}
	return best
}

// germanWhistCardStrength 切り札を上乗せした比較用の強さ
func germanWhistCardStrength(c *Card, trumpSuit int) int {
	s := germanWhistRank(c)
	if c.GetDesign() == trumpSuit {
		s += 100
	}
	return s
}

// --- Getters ---

// GetPhase フェーズ取得
func (g *GermanWhist) GetPhase() GermanWhistPhase { return g.phase }

// GetTrickNumber 完了したトリック数
func (g *GermanWhist) GetTrickNumber() int { return g.trickNumber }

// GetUpCard 表向きの札（前半のみ、尽きたら nil）
func (g *GermanWhist) GetUpCard() *Card { return g.upCard }

// GetStockCount 伏せ札の残り枚数
func (g *GermanWhist) GetStockCount() int { return len(g.stock) }

// GetTrumpSuit 切り札スート
func (g *GermanWhist) GetTrumpSuit() int { return g.trumpSuit }

// GetCurrentTrick 場に出ている札
func (g *GermanWhist) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetCurrentPlayerIdx 手番のプレイヤー
func (g *GermanWhist) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetLeadPlayerIdx リードしたプレイヤー
func (g *GermanWhist) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetPlayerCnt プレイヤー数
func (g *GermanWhist) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤーを取得
func (g *GermanWhist) GetPlayer(i int) *GermanWhistPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (g *GermanWhist) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者（未確定なら -1）
func (g *GermanWhist) GetWinnerIdx() int { return g.winnerIdx }

// GermanWhistHint ヒント情報
type GermanWhistHint struct {
	// CardIndex 推奨する手札のインデックス
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
}

// GetValidPlayIndices 出せる手札のインデックスを返す。
//
// リードなら全部。フォローできるならリードのスートだけ。
func (g *GermanWhist) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	p := g.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if g.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// IsHumanTurn 人間の手番かを返す
func (g *GermanWhist) IsHumanTurn() bool {
	return !g.gameEndFlag && g.currentPlayerIdx == 0
}

// GetHint 人間プレイヤーへの推奨手を返す。手番でなければ nil。
//
// 理由キーは前半と後半で意味が反転する。前半は表向きの札が欲しいときだけ
// 取りに行き、要らなければ**わざと負ける**のが定石なので、同じ「1 枚出す」
// でも狙いが逆になる。後半は取ったトリックがそのまま得点なので常に取りに行く。
func (g *GermanWhist) GetHint() *GermanWhistHint {
	if !g.IsHumanTurn() || g.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := g.chooseCpuCard(0)
	return &GermanWhistHint{CardIndex: &idx, Reason: g.hintReason()}
}

// hintReason 現在の狙いを表す理由キーを返す
func (g *GermanWhist) hintReason() string {
	if g.phase != GermanWhistPhaseDraw {
		return "germanWhistWinTrick"
	}
	if g.upCardIsWorthTaking() {
		return "germanWhistTakeUpCard"
	}
	return "germanWhistDuck"
}

// GiveUp ギブアップ
func (g *GermanWhist) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.phase = GermanWhistPhaseGameEnd
	g.gameEndFlag = true
	g.winnerIdx = 1
	g.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (g *GermanWhist) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLogAt(g.trickNumber, playerIdx, actionType, detail, cards)
}

// germanWhistJSON is the JSON wire format for GermanWhist.
type germanWhistJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*GermanWhistPlayer `json:"pl"`
	Phase            GermanWhistPhase     `json:"ph"`
	TrickNumber      int                  `json:"tn"`
	UpCard           *Card                `json:"uc"`
	Stock            []*Card              `json:"st"`
	TrumpSuit        int                  `json:"ts"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	CurrentPlayerIdx int                  `json:"cp"`
	LeadPlayerIdx    int                  `json:"lp"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerIdx        int                  `json:"wi"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (g *GermanWhist) MarshalJSON() ([]byte, error) {
	return json.Marshal(&germanWhistJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Phase:            g.phase,
		TrickNumber:      g.trickNumber,
		UpCard:           g.upCard,
		Stock:            g.stock,
		TrumpSuit:        g.trumpSuit,
		CurrentTrick:     g.currentTrick,
		CurrentPlayerIdx: g.currentPlayerIdx,
		LeadPlayerIdx:    g.leadPlayerIdx,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (g *GermanWhist) UnmarshalJSON(data []byte) error {
	var j germanWhistJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < GermanWhistPhaseDraw || j.Phase > GermanWhistPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.TrickNumber < 0 || j.TrickNumber > GermanWhistStageTricks*2 {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if len(j.Stock) > CardCnt || len(j.ActionLog) > germanWhistMaxSliceLen {
		return errors.New("germanwhist: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > GermanWhistPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= GermanWhistPlayerCnt {
		return fmt.Errorf("invalid current player: %d", j.CurrentPlayerIdx)
	}
	if j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= GermanWhistPlayerCnt {
		return fmt.Errorf("invalid lead player: %d", j.LeadPlayerIdx)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= GermanWhistPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.TrumpCards != nil {
		g.trumpCards = j.TrumpCards
	}
	if len(j.Players) == GermanWhistPlayerCnt {
		g.players = j.Players
	}
	g.phase = j.Phase
	g.trickNumber = j.TrickNumber
	g.upCard = j.UpCard
	g.stock = j.Stock
	g.trumpSuit = j.TrumpSuit
	g.currentTrick = j.CurrentTrick
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	return nil
}
