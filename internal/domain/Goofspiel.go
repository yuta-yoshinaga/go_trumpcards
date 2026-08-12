//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// GoofspielPhase はゲームフェーズ。
type GoofspielPhase int

// ゴフスピールのフェーズ定数
const (
	// GoofspielPhaseBid 賞札が公開され、全員が伏せて入札している状態
	GoofspielPhaseBid GoofspielPhase = 0
	// GoofspielPhaseReveal 入札が公開され、勝者が決まった状態
	GoofspielPhaseReveal GoofspielPhase = 1
	// GoofspielPhaseGameEnd ゲーム終了
	GoofspielPhaseGameEnd GoofspielPhase = 2
)

// GoofspielPhaseMin / Max はフェーズ列挙の範囲 (復元時の検証用)。
const (
	GoofspielPhaseMin = GoofspielPhaseBid
	GoofspielPhaseMax = GoofspielPhaseGameEnd
)

// goofspielMaxSliceLen は復元時に受け付けるスライス長の上限。
const goofspielMaxSliceLen = 2000

// GoofspielHint は人間への助言。
type GoofspielHint struct {
	// CardIndex は出すべき手札の位置。
	CardIndex *int
	// Reason は助言の理由。
	Reason string
}

// goofspielRank はランクを 1..13 で返す (賞札の得点でもある)。
func goofspielRank(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue()
}

// Goofspiel はゴフスピール (Goofspiel / GOPS) のゲームクラス。
//
// アメリカ発の数学的ゲームで、ゲーム理論の題材としても有名。**手番が無く、全員が
// 同時に伏せて入札します。**
//
// # 隠されているのは「今いくら出したか」だけ
//
// 各自の入札札は最初から最後まで**自分のスート 13 枚**で、使った札は場に残ります。
// つまり**相手の残り手札は完全に読めます**——隠れているのはこのラウンドの入札だけ。
// 情報が非対称なのではなく、**同時であること**だけが読み合いを生みます。
//
// # デッキ
//
// ダイヤを賞札、残りのスートを各自の入札札にします。**3 人卓は賞札 1 + 入札 3 で
// 4 スート要ります**——issue の「3 スートに分割」は 2 人卓の話で、3 人では足りません。
// ちょうど標準デッキに収まります。
//
// # 停止保証
//
// 賞札は 13 枚で、1 ラウンドに必ず 1 枚めくれて戻りません。**13 ラウンドちょうどで
// 終わります**。
type Goofspiel struct {
	players []*GoofspielPlayer
	config  GoofspielConfig

	phase GoofspielPhase
	// prizePile はまだめくっていない賞札 (末尾が次にめくられる)。
	prizePile []*Card
	// currentPrize はいま公開されている賞札。
	currentPrize *Card
	// carriedPrizes は同点で持ち越された賞札。
	//
	// **持ち越しは「今回の賞が増える」こと。** 消滅させる設定では常に空です。
	carriedPrizes []*Card
	// bids[i] は席 i が伏せた札 (nil = まだ出していない)。
	bids []*Card
	// revealedBids は直前のラウンドで公開された入札。
	revealedBids []*Card
	// lastWinnerIdx は直前のラウンドの勝者 (-1 = 同点で勝者なし)。
	lastWinnerIdx int
	// lastGained は直前のラウンドで動いた得点。
	lastGained  int
	roundNumber int
	gameEndFlag bool
	winnerIdx   int
	actionLogBase

	rng *rand.Rand
}

// NewGoofspiel はコンストラクタ。
func NewGoofspiel(players []*GoofspielPlayer, config GoofspielConfig) *Goofspiel {
	if config.Validate() != nil {
		config = DefaultGoofspielConfig()
	}
	if players == nil {
		players = newGoofspielSeats(config.PlayerCnt)
	}
	return &Goofspiel{
		players:       players,
		config:        config,
		lastWinnerIdx: -1,
		winnerIdx:     -1,
		rng:           rand.New(rand.NewSource(rand.Int63())),
	}
}

// newGoofspielSeats は席 0 を人間、以降を CPU とした座席を作る。
func newGoofspielSeats(n int) []*GoofspielPlayer {
	seats := make([]*GoofspielPlayer, n)
	for i := range seats {
		seats[i] = NewGoofspielPlayer(i == 0)
	}
	return seats
}

// NewDefaultGoofspiel は既定設定の Goofspiel を返す。
func NewDefaultGoofspiel() *Goofspiel {
	cfg := DefaultGoofspielConfig()
	return NewGoofspiel(newGoofspielSeats(cfg.PlayerCnt), cfg)
}

// SetRand はテスト用に乱数源を差し替える。
func (g *Goofspiel) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲームを初期化する。
func (g *Goofspiel) Reset() {
	for i, p := range g.players {
		p.ResetGame()
		// **入札札は自分のスート 13 枚で固定。** 配りに乱数は要りません。
		for v := 1; v <= GoofspielRounds; v++ {
			p.AddCard(NewCard(GoofspielBidSuit(i), v, false))
		}
	}

	// 賞札はダイヤ 13 枚を伏せて積む。
	g.prizePile = make([]*Card, 0, GoofspielRounds)
	for v := 1; v <= GoofspielRounds; v++ {
		g.prizePile = append(g.prizePile, NewCard(GoofspielPrizeSuit(), v, false))
	}
	g.rng.Shuffle(len(g.prizePile), func(i, j int) {
		g.prizePile[i], g.prizePile[j] = g.prizePile[j], g.prizePile[i]
	})

	g.phase = GoofspielPhaseBid
	g.currentPrize = nil
	g.carriedPrizes = nil
	g.bids = make([]*Card, len(g.players))
	g.revealedBids = nil
	g.lastWinnerIdx = -1
	g.lastGained = 0
	g.roundNumber = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil

	g.addLog(-1, "start", fmt.Sprintf("ゴフスピールを開始しました（%d 人）", g.config.PlayerCnt), nil)
	g.revealPrize()
}

// revealPrize は次の賞札をめくる。
func (g *Goofspiel) revealPrize() {
	if len(g.prizePile) == 0 {
		g.finish()
		return
	}
	g.currentPrize = g.prizePile[len(g.prizePile)-1]
	g.prizePile = g.prizePile[:len(g.prizePile)-1]
	g.roundNumber++
	g.phase = GoofspielPhaseBid
	g.bids = make([]*Card, len(g.players))
	g.addLog(-1, "prize", fmt.Sprintf("賞札 %d が公開されました", goofspielRank(g.currentPrize)),
		[]*Card{g.currentPrize})
}

// PrizeValue はいま懸かっている得点を返す (持ち越しを含む)。
func (g *Goofspiel) PrizeValue() int {
	total := goofspielRank(g.currentPrize)
	for _, c := range g.carriedPrizes {
		total += goofspielRank(c)
	}
	return total
}

// IsHumanTurn は人間の入力を待っているかを返す。
//
// **手番はありません。** 人間がまだ伏せていなければ真、というだけです。
func (g *Goofspiel) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case GoofspielPhaseBid:
		return g.bids[0] == nil
	case GoofspielPhaseReveal:
		// **公開された入札を読む時間。** 押されるまで次はめくりません。
		return true
	default:
		return false
	}
}

// GetValidBidIndices は入札できる手札の位置を返す。
func (g *Goofspiel) GetValidBidIndices(playerIdx int) []int {
	if g.phase != GoofspielPhaseBid || playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	if g.bids[playerIdx] != nil {
		return nil
	}
	p := g.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		out = append(out, i)
	}
	return out
}

// PlayerBid は人間が入札札を伏せる。
func (g *Goofspiel) PlayerBid(cardIndex int) error {
	if g.phase != GoofspielPhaseBid {
		return errors.New("いまは入札の場面ではありません")
	}
	if g.bids[0] != nil {
		return errors.New("すでに入札しています")
	}
	if err := g.bid(0, cardIndex); err != nil {
		return err
	}
	// **人間が伏せたら CPU も伏せて、同時に公開する。**
	g.cpuBidAll()
	g.resolveIfReady()
	return nil
}

// NextRound は公開状態から次の賞札をめくる。
func (g *Goofspiel) NextRound() error {
	if g.gameEndFlag {
		return errors.New("ゲームは終了しています")
	}
	if g.phase != GoofspielPhaseReveal {
		return errors.New("いまはラウンドの区切りではありません")
	}
	g.revealPrize()
	return nil
}

// bid は席 playerIdx に cardIndex を伏せさせる。
func (g *Goofspiel) bid(playerIdx, cardIndex int) error {
	p := g.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("手札の位置が範囲外です: %d", cardIndex)
	}
	g.bids[playerIdx] = p.RemoveCard(cardIndex)
	return nil
}

// cpuBidAll はまだ伏せていない CPU に入札させる。
func (g *Goofspiel) cpuBidAll() {
	for i := range g.players {
		if g.players[i].GetIsHuman() || g.bids[i] != nil {
			continue
		}
		_ = g.bid(i, g.chooseCpuCard(i))
	}
}

// CpuPlay は CPU の入札を進める。
func (g *Goofspiel) CpuPlay() {
	if g.gameEndFlag || g.phase != GoofspielPhaseBid {
		return
	}
	g.cpuBidAll()
	g.resolveIfReady()
}

// resolveIfReady は全員が伏せていれば同時に公開する。
func (g *Goofspiel) resolveIfReady() {
	for _, b := range g.bids {
		if b == nil {
			return
		}
	}
	g.resolve()
}

// resolve は入札を公開して賞札の行き先を決める。
func (g *Goofspiel) resolve() {
	g.revealedBids = make([]*Card, len(g.bids))
	copy(g.revealedBids, g.bids)

	winner, tied := g.highestBidder()
	prize := g.PrizeValue()
	switch {
	case tied:
		// **同点は誰も取らない。** 設定によって消えるか、次の賞に上乗せされます。
		g.lastWinnerIdx = -1
		g.lastGained = 0
		if g.config.TieRule == GoofspielTieCarryOver {
			g.carriedPrizes = append(g.carriedPrizes, g.currentPrize)
			g.addLog(-1, "tie", fmt.Sprintf("同点。賞札 %d は次に持ち越します",
				goofspielRank(g.currentPrize)), nil)
		} else {
			g.addLog(-1, "tie", fmt.Sprintf("同点。賞札 %d は流れました",
				goofspielRank(g.currentPrize)), nil)
		}
	default:
		g.players[winner].AddScore(prize)
		g.lastWinnerIdx = winner
		g.lastGained = prize
		g.carriedPrizes = nil
		g.addLog(winner, "win", fmt.Sprintf("%d 点を獲得しました", prize), nil)
	}

	g.currentPrize = nil
	g.bids = make([]*Card, len(g.players))
	g.phase = GoofspielPhaseReveal

	if len(g.prizePile) == 0 {
		g.finish()
	}
}

// highestBidder は最高額を出した席を返す。同額が複数なら tied が真。
func (g *Goofspiel) highestBidder() (int, bool) {
	best, bestRank, tied := -1, 0, false
	for i, b := range g.revealedBids {
		r := goofspielRank(b)
		switch {
		case r > bestRank:
			best, bestRank, tied = i, r, false
		case r == bestRank:
			tied = true
		}
	}
	return best, tied
}

// finish は得点のいちばん高い席を勝ちにして終局する。
func (g *Goofspiel) finish() {
	g.phase = GoofspielPhaseGameEnd
	g.gameEndFlag = true
	g.winnerIdx = g.leaderIdx()
	g.addLog(g.winnerIdx, "result",
		fmt.Sprintf("%d 点でいちばん多く取りました", g.players[g.winnerIdx].GetScore()), nil)
}

// leaderIdx は得点のいちばん高い席を返す (同点なら若い席)。
func (g *Goofspiel) leaderIdx() int {
	best := 0
	for i, p := range g.players {
		if p.GetScore() > g.players[best].GetScore() {
			best = i
		}
	}
	return best
}

// GiveUp は投了する。
func (g *Goofspiel) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.phase = GoofspielPhaseGameEnd
	g.gameEndFlag = true
	best := -1
	for i := 1; i < len(g.players); i++ {
		if best < 0 || g.players[i].GetScore() > g.players[best].GetScore() {
			best = i
		}
	}
	if best < 0 {
		best = 0
	}
	g.winnerIdx = best
	g.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の入札を選ぶ。
//
// **賞札の価値に見合った札を出します。** 残っている賞札の中での順位を、自分の
// 残り手札の順位に対応させる——ゲーム理論でよく使われる素直な戦略です。
func (g *Goofspiel) chooseCpuCard(playerIdx int) int {
	p := g.players[playerIdx]
	if p.GetCardsSize() == 0 {
		return 0
	}
	prize := goofspielRank(g.currentPrize)

	// 残っている賞札 (いま公開中の分を含む) の中で、この賞は上から何番目か。
	higher := 0
	for _, c := range g.prizePile {
		if goofspielRank(c) > prize {
			higher++
		}
	}
	// 同じ順位の入札札を出す。手札は昇順なので、上から higher 番目。
	idx := p.GetCardsSize() - 1 - higher
	if idx < 0 {
		idx = 0
	}
	if idx >= p.GetCardsSize() {
		idx = p.GetCardsSize() - 1
	}
	return idx
}

// GetHint は人間への助言を返す。
func (g *Goofspiel) GetHint() *GoofspielHint {
	if g.gameEndFlag || g.phase != GoofspielPhaseBid || g.bids[0] != nil {
		return nil
	}
	if g.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := g.chooseCpuCard(0)
	reason := "goofspielMatch"
	switch {
	case len(g.carriedPrizes) > 0:
		reason = "goofspielCarried"
	case goofspielRank(g.currentPrize) >= 11:
		reason = "goofspielHighPrize"
	case goofspielRank(g.currentPrize) <= 3:
		reason = "goofspielLowPrize"
	}
	return &GoofspielHint{CardIndex: &idx, Reason: reason}
}

// addLog は棋譜に 1 行足す。
func (g *Goofspiel) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLog(playerIdx, actionType, detail, cards)
}

// GetConfig は設定を返す。
func (g *Goofspiel) GetConfig() GoofspielConfig { return g.config }

// SetConfig は設定を更新する。
func (g *Goofspiel) SetConfig(cfg GoofspielConfig) {
	if err := cfg.Validate(); err != nil {
		return
	}
	if cfg.PlayerCnt != g.config.PlayerCnt {
		g.players = newGoofspielSeats(cfg.PlayerCnt)
	}
	g.config = cfg
}

// GetPhase は現在のフェーズを返す。
func (g *Goofspiel) GetPhase() GoofspielPhase { return g.phase }

// GetGameEndFlag は終局フラグを返す。
func (g *Goofspiel) GetGameEndFlag() bool { return g.gameEndFlag }

// GetCurrentPrize はいま公開されている賞札を返す (nil = 公開中でない)。
func (g *Goofspiel) GetCurrentPrize() *Card { return g.currentPrize }

// GetCarriedPrizes は持ち越された賞札を返す。
func (g *Goofspiel) GetCarriedPrizes() []*Card { return g.carriedPrizes }

// GetPrizeRemaining はまだめくっていない賞札の枚数を返す。
func (g *Goofspiel) GetPrizeRemaining() int { return len(g.prizePile) }

// HasBid は席 i が伏せ終えたかを返す。
func (g *Goofspiel) HasBid(i int) bool {
	if i < 0 || i >= len(g.bids) {
		return false
	}
	return g.bids[i] != nil
}

// GetRevealedBids は直前に公開された入札を返す。
func (g *Goofspiel) GetRevealedBids() []*Card { return g.revealedBids }

// GetLastWinnerIdx は直前のラウンドの勝者を返す (-1 = 同点)。
func (g *Goofspiel) GetLastWinnerIdx() int { return g.lastWinnerIdx }

// GetLastGained は直前のラウンドで動いた得点を返す。
func (g *Goofspiel) GetLastGained() int { return g.lastGained }

// GetRoundNumber はラウンド数を返す。
func (g *Goofspiel) GetRoundNumber() int { return g.roundNumber }

// GetPlayerCnt は人数を返す。
func (g *Goofspiel) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は席 i のプレイヤーを返す。
func (g *Goofspiel) GetPlayer(i int) *GoofspielPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetWinnerIdx は勝者の席を返す (-1 = 未確定)。
func (g *Goofspiel) GetWinnerIdx() int { return g.winnerIdx }

// goofspielJSON is the JSON wire format for Goofspiel.
type goofspielJSON struct {
	Players       []*GoofspielPlayer `json:"pl"`
	Config        GoofspielConfig    `json:"cf"`
	Phase         GoofspielPhase     `json:"ph"`
	PrizePile     []*Card            `json:"pp"`
	CurrentPrize  *Card              `json:"cp"`
	CarriedPrizes []*Card            `json:"cr"`
	Bids          []*Card            `json:"bd"`
	RevealedBids  []*Card            `json:"rb"`
	LastWinnerIdx int                `json:"lw"`
	LastGained    int                `json:"lg"`
	RoundNumber   int                `json:"rn"`
	GameEndFlag   bool               `json:"ge"`
	WinnerIdx     int                `json:"wi"`
	ActionLog     []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Goofspiel) MarshalJSON() ([]byte, error) {
	return json.Marshal(goofspielJSON{
		Players:       g.players,
		Config:        g.config,
		Phase:         g.phase,
		PrizePile:     g.prizePile,
		CurrentPrize:  g.currentPrize,
		CarriedPrizes: g.carriedPrizes,
		Bids:          g.bids,
		RevealedBids:  g.revealedBids,
		LastWinnerIdx: g.lastWinnerIdx,
		LastGained:    g.lastGained,
		RoundNumber:   g.roundNumber,
		GameEndFlag:   g.gameEndFlag,
		WinnerIdx:     g.winnerIdx,
		ActionLog:     g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Goofspiel) UnmarshalJSON(data []byte) error {
	var j goofspielJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("seat count %d does not match the configured %d", len(j.Players), j.Config.PlayerCnt)
	}
	if j.Phase < GoofspielPhaseMin || j.Phase > GoofspielPhaseMax {
		return fmt.Errorf("phase out of range: %d", j.Phase)
	}
	if len(j.ActionLog) > goofspielMaxSliceLen {
		return fmt.Errorf("action log too long: %d", len(j.ActionLog))
	}
	if len(j.Bids) != 0 && len(j.Bids) != len(j.Players) {
		return fmt.Errorf("bids has %d slots for %d seats", len(j.Bids), len(j.Players))
	}
	if len(j.RevealedBids) != 0 && len(j.RevealedBids) != len(j.Players) {
		return fmt.Errorf("revealed bids has %d entries for %d seats", len(j.RevealedBids), len(j.Players))
	}
	if j.LastWinnerIdx < -1 || j.LastWinnerIdx >= len(j.Players) {
		return fmt.Errorf("last winner index out of range: %d", j.LastWinnerIdx)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= len(j.Players) {
		return fmt.Errorf("winner index out of range: %d", j.WinnerIdx)
	}
	if j.GameEndFlag != (j.Phase == GoofspielPhaseGameEnd) {
		return fmt.Errorf("the game-end flag and the phase disagree (flag=%v, phase=%d)", j.GameEndFlag, j.Phase)
	}
	if j.GameEndFlag != (j.WinnerIdx >= 0) {
		return fmt.Errorf("a finished game has a winner and an unfinished one does not (flag=%v, winner=%d)",
			j.GameEndFlag, j.WinnerIdx)
	}
	if j.RoundNumber < 0 || j.RoundNumber > GoofspielRounds {
		return fmt.Errorf("round number out of range: %d", j.RoundNumber)
	}
	if j.LastGained < 0 {
		return fmt.Errorf("the last gain cannot be negative: %d", j.LastGained)
	}
	// **同点なら誰も取らない。** 勝者が居ないのに得点が動いた記録は矛盾です。
	if (j.LastWinnerIdx < 0) && j.LastGained != 0 {
		return fmt.Errorf("nobody won the round but %d points moved", j.LastGained)
	}

	// **賞札 13 枚は増えも減りもしません。** 山 + 公開中 + 持ち越し + 決着済み。
	settled := GoofspielRounds - len(j.PrizePile) - len(j.CarriedPrizes)
	if j.CurrentPrize != nil {
		settled--
	}
	if settled < 0 || settled > GoofspielRounds {
		return fmt.Errorf("the prizes do not add up to %d (pile=%d, carried=%d, current=%v)",
			GoofspielRounds, len(j.PrizePile), len(j.CarriedPrizes), j.CurrentPrize != nil)
	}
	// **入札の場面には賞札が出ています。**
	if j.Phase == GoofspielPhaseBid && j.CurrentPrize == nil {
		return errors.New("the bidding phase needs a prize on the table")
	}
	if j.Phase == GoofspielPhaseReveal && j.CurrentPrize != nil {
		return errors.New("the prize is settled once the bids are revealed")
	}

	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("seat %d is missing", i)
		}
		// **入札札は 13 枚から使ったぶんだけ減ります。**
		used := j.RoundNumber
		if j.Phase == GoofspielPhaseBid && len(j.Bids) > i && j.Bids[i] == nil {
			used--
		}
		if want := GoofspielRounds - used; p.GetCardsSize() != want {
			return fmt.Errorf("seat %d holds %d bid cards in round %d, want %d",
				i, p.GetCardsSize(), j.RoundNumber, want)
		}
	}

	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.prizePile = j.PrizePile
	g.currentPrize = j.CurrentPrize
	g.carriedPrizes = j.CarriedPrizes
	g.bids = j.Bids
	if len(g.bids) == 0 {
		g.bids = make([]*Card, len(j.Players))
	}
	g.revealedBids = j.RevealedBids
	g.lastWinnerIdx = j.LastWinnerIdx
	g.lastGained = j.LastGained
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
