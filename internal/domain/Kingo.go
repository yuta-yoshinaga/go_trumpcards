//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
)

// KingoPhase は進行の段階。
type KingoPhase int

const (
	// KingoPhaseBet は張りを待つ場面。
	KingoPhaseBet KingoPhase = iota
	// KingoPhaseResult はそのラウンドの決着を見せる場面。
	KingoPhaseResult
	// KingoPhaseGameEnd は終局。
	KingoPhaseGameEnd
)

// KingoPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const KingoPhaseMax = KingoPhaseGameEnd

// kingoMaxSliceLen は復元時に受け付けるスライスの上限。
const kingoMaxSliceLen = 512

// エラー値。
var (
	errKingoWrongPhase    = errors.New("kingo: not allowed in this phase")
	errKingoFinished      = errors.New("kingo: the game is over")
	errKingoBetAmount     = errors.New("kingo: bet out of range")
	errKingoBankerBets    = errors.New("kingo: the banker does not bet")
	errKingoRoundRange    = errors.New("kingo: round number out of range")
	errKingoBankerRange   = errors.New("kingo: banker seat out of range")
	errKingoPhaseRange    = errors.New("kingo: phase out of range")
	errKingoSliceTooLong  = errors.New("kingo: slice too long")
	errKingoSeatCount     = errors.New("kingo: seat count does not match the config")
	errKingoDeckTooLarge  = errors.New("kingo: the deck holds more cards than a kabufuda deck")
	errKingoRoundExceeded = errors.New("kingo: round number exceeds the configured rounds")
)

// KingoResult は 1 席のラウンド結果。
type KingoResult struct {
	PlayerIdx int
	Rank      KingoRank
	// MatchedValue はそろえた数字 (同じ役どうしの比較に使う)。
	MatchedValue int
	WonAmount    int
}

// Kingo はキンゴ (株札の勝負事) の卓。
//
// **おいちょかぶと同じ株札を使うが、競うものが違う。** おいちょかぶは合計の
// 下一桁で競うのに対し、キンゴは**同じ数字を何枚そろえたか**で決まる ──
// 合計を持ち込むと別のゲームになるので、役の判定に総和は一切出てこない。
//
// 親は席を順に回る。親は張らず、子の張りに対して 1 対 1 で受ける ── 総取りの
// 側が固定されないよう、ラウンド数は席数以上でなければならない (`Validate`)。
type Kingo struct {
	deck    []*Card
	players []*KingoPlayer
	config  KingoConfig

	phase KingoPhase
	// banker は親の席。ラウンドごとに 1 つずつ進む。
	banker int
	// roundNumber は何ラウンド目か (1 始まり)。
	roundNumber int

	results     []KingoResult
	gameEndFlag bool

	actionLog  []*ActionLogEntry
	turnNumber int
}

// NewKingo は Kingo を構築する。
func NewKingo(players []*KingoPlayer, config KingoConfig) *Kingo {
	return &Kingo{
		players:   players,
		config:    config,
		results:   make([]KingoResult, 0, len(players)),
		actionLog: make([]*ActionLogEntry, 0),
	}
}

// NewDefaultKingo は既定設定の卓を返す。
func NewDefaultKingo() *Kingo {
	cfg := DefaultKingoConfig()
	return NewKingo(NewKingoPlayersForTable(cfg.Seats, cfg.InitialChips), cfg)
}

// Reset はゲームを最初から始める。
func (g *Kingo) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultKingoConfig()
	}
	g.players = NewKingoPlayersForTable(g.config.Seats, g.config.InitialChips)
	g.banker = 0
	g.roundNumber = 0
	g.gameEndFlag = false
	g.actionLog = g.actionLog[:0]
	g.turnNumber = 0
	g.startRound()
}

// startRound は次のラウンドを始める。
func (g *Kingo) startRound() {
	g.deck = buildOichoKabuDeck()
	rand.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
	g.roundNumber++
	g.phase = KingoPhaseBet
	g.results = g.results[:0]
	for _, p := range g.players {
		p.ResetForRound()
	}
	g.appendLog(g.banker, "banker", "takes the bank", nil)

	// **張れない人間がいたらここで終える。** 精算のときだけ見ていると、
	// 張れないのに張りを待つ盤面で止まる ── 人間は張ることも進むことも
	// できず、ゲームが動かなくなる。
	if g.humanIsBroke() && !g.IsHumanBanker() {
		g.finish()
		return
	}

	// **子の張りは先に決める。** 人間が親の回でも、子の張りは出そろう。
	for i, p := range g.players {
		if i == g.banker || p.GetIsHuman() {
			continue
		}
		p.SetBet(g.cpuBet(p))
	}
}

// cpuBet は CPU の張り額を決める。
func (g *Kingo) cpuBet(p *KingoPlayer) int {
	bet := g.config.MinBet
	if p.GetChips() >= g.config.MinBet*3 {
		bet = g.config.MinBet * (1 + rand.Intn(2))
	}
	return min(bet, p.GetChips())
}

// PlaceBet は人間の張りを処理する。
func (g *Kingo) PlaceBet(amount int) error {
	if g.gameEndFlag {
		return errKingoFinished
	}
	if g.phase != KingoPhaseBet {
		return errKingoWrongPhase
	}
	human := g.HumanSeat()
	if human == g.banker {
		// **親は張らない。** 受ける側なので、張らせると二重に払うことになる。
		return errKingoBankerBets
	}
	p := g.players[human]
	if amount < g.config.MinBet || amount > KingoMaxBet ||
		amount%KingoBetStep != 0 || amount > p.GetChips() {
		return errKingoBetAmount
	}
	p.SetBet(amount)
	g.appendLog(human, "bet", "places a bet", nil)
	g.deal()
	g.resolve()
	return nil
}

// Deal は人間が親のとき、子の張りを見てから配る。
//
// **親の回も一手を挟む。** 張りを受けてすぐ決着まで走らせると、親の側に
// 見せ場が無く、子の張りを見た記憶のないまま結果だけが出る ── 親は受ける
// 側とはいえ、卓を進めるのは親の仕事。
func (g *Kingo) Deal() error {
	if g.gameEndFlag {
		return errKingoFinished
	}
	if g.phase != KingoPhaseBet {
		return errKingoWrongPhase
	}
	if !g.IsHumanBanker() {
		// 子の回は張りが配る合図なので、`PlaceBet` を通す。
		return errKingoWrongPhase
	}
	g.deal()
	g.resolve()
	return nil
}

// deal は全員に手札を配る。
func (g *Kingo) deal() {
	for range KingoHandSize {
		for _, p := range g.players {
			p.AddCard(g.drawCard())
		}
	}
	for i, p := range g.players {
		g.appendLog(i, "deal", KingoRankName(p.GetRank()), p.GetCards())
	}
}

// drawCard は山から 1 枚引く。
func (g *Kingo) drawCard() *Card {
	if len(g.deck) == 0 {
		return nil
	}
	c := g.deck[0]
	g.deck = g.deck[1:]
	return c
}

// resolve は親と各子を突き合わせて精算する。
//
// **負けた子から集めてから、勝った子に払う。** 逆順にすると、3 人が同時に
// 勝ったときに親の手持ちを超えて払い出し、親の残高が負になる。
func (g *Kingo) resolve() {
	g.phase = KingoPhaseResult
	g.results = g.results[:0]
	for range g.players {
		g.results = append(g.results, KingoResult{})
	}

	banker := g.players[g.banker]
	for i, p := range g.players {
		g.results[i].PlayerIdx = i
		g.results[i].Rank = p.GetRank()
		g.results[i].MatchedValue = KingoMatchedValue(p.GetCards())
	}

	// 1) 負けた子から集める。
	winners := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if i == g.banker || p.GetBet() <= 0 {
			continue
		}
		cmp := KingoCompare(p.GetCards(), banker.GetCards())
		switch {
		case cmp < 0:
			// **払う額は勝った側の役で決まる。** 親が嵐なら 3 倍取る。
			loss := min(p.GetBet()*KingoPayout(banker.GetRank()), p.GetChips())
			p.SubtractChips(loss)
			banker.AddChips(loss)
			p.SetWonAmount(-loss)
			g.results[i].WonAmount = -loss
			g.results[g.banker].WonAmount += loss
		case cmp > 0:
			winners = append(winners, i)
		default:
			// 引き分け。張りは戻る (動かさない)。
			p.SetWonAmount(0)
		}
	}

	// 2) 勝った子へ払う。親の手持ちを超えないよう、若い席から順に払う。
	for _, i := range winners {
		p := g.players[i]
		win := min(p.GetBet()*KingoPayout(p.GetRank()), banker.GetChips())
		banker.SubtractChips(win)
		p.AddChips(win)
		p.SetWonAmount(win)
		g.results[i].WonAmount = win
		g.results[g.banker].WonAmount -= win
	}
	banker.SetWonAmount(g.results[g.banker].WonAmount)

	if g.roundNumber >= g.config.Rounds || g.humanIsBroke() || g.aliveSeats() <= 1 {
		g.finish()
	}
}

// NextRound は次のラウンドを始める。
func (g *Kingo) NextRound() error {
	if g.gameEndFlag {
		return errKingoFinished
	}
	if g.phase != KingoPhaseResult {
		return errKingoWrongPhase
	}
	// **親は交代してから配る。** 決着の画面はそのラウンドの親を映すので、
	// 先に回すと「勝った親」と表示上の親がずれる。
	g.banker = (g.banker + 1) % len(g.players)
	g.startRound()
	return nil
}

// finish は終局にする。
func (g *Kingo) finish() {
	g.gameEndFlag = true
	g.phase = KingoPhaseGameEnd
}

// humanIsBroke は人間が張れなくなったかを返す。
func (g *Kingo) humanIsBroke() bool {
	return g.players[g.HumanSeat()].GetChips() < g.config.MinBet
}

// aliveSeats はチップの残っている席の数を返す。
func (g *Kingo) aliveSeats() int {
	n := 0
	for _, p := range g.players {
		if p.GetChips() > 0 {
			n++
		}
	}
	return n
}

// HumanSeat は人間の席を返す。
func (g *Kingo) HumanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// IsHumanBanker は人間が親かを返す。
func (g *Kingo) IsHumanBanker() bool { return g.HumanSeat() == g.banker }

// IsHumanTurn は人間の入力を待っているかを返す。
//
// **親の回も人間の手番。** 子なら張り、親なら配る合図を待っている。
func (g *Kingo) IsHumanTurn() bool {
	return g.phase == KingoPhaseBet && !g.gameEndFlag
}

// WinnerSeat はいちばんチップの多い席を返す。
func (g *Kingo) WinnerSeat() int {
	best, seat := -1, 0
	for i, p := range g.players {
		if p.GetChips() > best {
			best, seat = p.GetChips(), i
		}
	}
	return seat
}

// GetConfig は設定を返す。
func (g *Kingo) GetConfig() KingoConfig { return g.config }

// SetConfig は設定を差し替える。
func (g *Kingo) SetConfig(c KingoConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *Kingo) GetPhase() KingoPhase { return g.phase }

// GetGameEndFlag は終局かを返す。
func (g *Kingo) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayers は席の一覧を返す。
func (g *Kingo) GetPlayers() []*KingoPlayer { return g.players }

// GetBankerSeat は親の席を返す。
func (g *Kingo) GetBankerSeat() int { return g.banker }

// GetRoundNumber は何ラウンド目かを返す。
func (g *Kingo) GetRoundNumber() int { return g.roundNumber }

// GetResults は直近のラウンド結果を返す。
func (g *Kingo) GetResults() []KingoResult { return g.results }

// GetRemainingCards は山の残り枚数を返す。
func (g *Kingo) GetRemainingCards() int { return len(g.deck) }

// GetActionLog は棋譜を返す。
func (g *Kingo) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *Kingo) appendLog(seat int, actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  seat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- 助言 ---

// KingoHint は人間への助言。
type KingoHint struct {
	// Action は薦める操作 ("bet" / "next")。
	Action string
	// Amount は薦める張り額 (Action が "bet" のとき)。
	Amount int
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// GetHint は人間への助言を返す。
//
// **配る前に手札は見えない。** 張りは完全に情報の無い賭けなので、助言できるのは
// 「手持ちに対して重すぎないか」だけ ── ここで役の話をすると、見えていない
// ものを見たふりをすることになる。
func (g *Kingo) GetHint() *KingoHint {
	if g.gameEndFlag {
		return nil
	}
	if g.phase == KingoPhaseResult {
		return &KingoHint{Action: "next", Reason: "roundIsOver"}
	}
	if g.IsHumanBanker() {
		return &KingoHint{Action: "deal", Reason: "youAreTheBanker"}
	}
	p := g.players[g.HumanSeat()]
	if p.GetChips() < g.config.MinBet*5 {
		return &KingoHint{Action: "bet", Amount: g.config.MinBet, Reason: "stackIsShort"}
	}
	return &KingoHint{Action: "bet", Amount: g.config.MinBet, Reason: "noInformationYet"}
}

// --- 永続化 ---

// kingoJSON is the JSON wire format for Kingo.
type kingoJSON struct {
	Deck        []*Card           `json:"dk"`
	Players     []*KingoPlayer    `json:"pl"`
	Config      KingoConfig       `json:"cf"`
	Phase       int               `json:"ph"`
	Banker      int               `json:"bk"`
	RoundNumber int               `json:"rn"`
	Results     []KingoResult     `json:"rs"`
	GameEndFlag bool              `json:"ge"`
	ActionLog   []*ActionLogEntry `json:"al"`
	TurnNumber  int               `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *Kingo) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingoJSON{
		Deck: g.deck, Players: g.players, Config: g.config,
		Phase: int(g.phase), Banker: g.banker, RoundNumber: g.roundNumber,
		Results: g.results, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Kingo) UnmarshalJSON(data []byte) error {
	var j kingoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := kingoValidate(&j); err != nil {
		return err
	}

	g.deck = j.Deck
	g.players = j.Players
	g.config = j.Config
	g.phase = KingoPhase(j.Phase)
	g.banker = j.Banker
	g.roundNumber = j.RoundNumber
	g.results = j.Results
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber

	if g.results == nil {
		g.results = make([]KingoResult, 0, len(g.players))
	}
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// kingoValidate は復元した盤面が破綻していないかを見る。
func kingoValidate(j *kingoJSON) error {
	if len(j.Players) > kingoMaxSliceLen || len(j.ActionLog) > kingoMaxSliceLen ||
		len(j.Results) > kingoMaxSliceLen {
		return errKingoSliceTooLong
	}
	if len(j.Deck) > KingoDeckSize {
		return errKingoDeckTooLarge
	}
	if len(j.Players) != j.Config.Seats {
		return errKingoSeatCount
	}
	if j.Phase < 0 || KingoPhase(j.Phase) > KingoPhaseMax {
		return errKingoPhaseRange
	}
	if j.Banker < 0 || j.Banker >= len(j.Players) {
		return errKingoBankerRange
	}
	if j.RoundNumber < 1 {
		return errKingoRoundRange
	}
	// **ラウンド数は設定を超えられない。** 超えた保存は「終わっているのに
	// 続く卓」で、親の回りも設定の意味も壊れる ── 範囲検査だけでは通る。
	if j.RoundNumber > j.Config.Rounds {
		return errKingoRoundExceeded
	}
	return nil
}
