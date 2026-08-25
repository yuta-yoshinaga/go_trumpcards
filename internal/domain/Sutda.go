//go:build !js || !wasm || extra2

// Package domain ソッタ (Sutda / 섯다) のドメインモデル。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// フェーズ。
const (
	// SutdaPhaseBet ベッティング中。
	SutdaPhaseBet = "bet"
	// SutdaPhaseShowdown ショーダウン (結果表示、次ハンド待ち)。
	SutdaPhaseShowdown = "showdown"
	// SutdaPhaseGameEnd ゲーム終了 (誰かのチップが尽きた)。
	SutdaPhaseGameEnd = "gameEnd"
)

// SutdaAction は人間が選べる行動。
const (
	// SutdaActionCall コール (現在のベットに合わせる。差額 0 ならチェック)。
	SutdaActionCall = "call"
	// SutdaActionRaise レイズ (1 単位上げる)。
	SutdaActionRaise = "raise"
	// SutdaActionFold フォールド (降りる)。
	SutdaActionFold = "fold"
)

// SutdaHandResult は 1 ハンドの結果。
type SutdaHandResult struct {
	// Winners は勝った席 (同点なら複数)。
	Winners []int
	// Pot はその席が分け合った総額。
	Pot int
	// Hands は席ごとの役 (降りた席も含む)。
	Hands []SutdaHand
	// Folded は席ごとに降りたか。
	Folded []bool
}

// Sutda はソッタの状態を保持する集約ルート。
type Sutda struct {
	deck          []*Card
	drawIdx       int
	players       []*SutdaPlayer
	config        SutdaConfig
	phase         string
	handNumber    int
	dealerIdx     int
	currentPlayer int
	pot           int
	currentBet    int
	raiseCount    int
	// actedThisRound は現在のベット額に対して行動を終えた席。
	actedThisRound []bool
	lastResult     *SutdaHandResult
	gameEndFlag    bool
	winnerIdx      int
	actionLogBase
}

// NewSutda はコンストラクタ。
func NewSutda(players []*SutdaPlayer, config SutdaConfig) *Sutda {
	return &Sutda{
		players:   players,
		config:    config,
		phase:     SutdaPhaseBet,
		winnerIdx: -1,
	}
}

// NewDefaultSutda は既定の設定でソッタを生成する。
func NewDefaultSutda() *Sutda {
	return NewSutdaWithConfig(DefaultSutdaConfig())
}

// NewSutdaWithConfig は設定から席を組み立ててソッタを生成する。
func NewSutdaWithConfig(cfg SutdaConfig) *Sutda {
	seats := cfg.Seats
	if seats < SutdaMinSeats || seats > SutdaMaxSeats {
		seats = SutdaDefaultSeats
	}
	players := make([]*SutdaPlayer, seats)
	players[0] = NewSutdaPlayer(true, cfg.StartChips)
	for i := 1; i < seats; i++ {
		players[i] = NewSutdaPlayer(false, cfg.StartChips)
	}
	return NewSutda(players, cfg)
}

// Reset は新しいゲームを開始する。
func (s *Sutda) Reset() {
	// **席数は設定で変わる。** 変えたのに席が据え置きだと、卓の人数と設定が
	// 食い違ったまま配ることになる。
	if len(s.players) != s.config.Seats {
		rebuilt := NewSutdaWithConfig(s.config)
		s.players = rebuilt.players
	}
	for _, p := range s.players {
		p.ResetHand()
		p.SetChips(s.config.StartChips)
	}
	s.handNumber = 1
	// **開幕の親は最後の席。** 親の左隣から始まるので、席 0 の人間が最初に動く。
	s.dealerIdx = len(s.players) - 1
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.lastResult = nil
	s.actionLog = make([]*ActionLogEntry, 0)
	s.startHand()
}

// NextHand は次のハンドを開始する。
func (s *Sutda) NextHand() {
	if s.gameEndFlag || s.phase != SutdaPhaseShowdown {
		return
	}
	if s.checkGameEnd() {
		return
	}
	s.handNumber++
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
	s.startHand()
}

// startHand は参加料を集め、2 枚ずつ配ってベッティングに入る。
func (s *Sutda) startHand() {
	for _, p := range s.players {
		p.ResetHand()
	}
	s.pot = 0
	s.currentBet = 0
	s.raiseCount = 0
	s.actedThisRound = make([]bool, len(s.players))

	s.deck = buildSutdaDeck()
	rand.Shuffle(len(s.deck), func(i, j int) { s.deck[i], s.deck[j] = s.deck[j], s.deck[i] })
	s.drawIdx = 0

	for _, p := range s.players {
		ante := SutdaAnte
		if p.GetChips() < ante {
			ante = p.GetChips()
		}
		p.AddChips(-ante)
		p.AddBet(ante)
		s.pot += ante
	}
	s.currentBet = SutdaAnte

	for i := 0; i < SutdaHandSize; i++ {
		for j := 0; j < len(s.players); j++ {
			idx := (s.dealerIdx + 1 + j) % len(s.players)
			if card := s.draw(); card != nil {
				s.players[idx].AddCard(card)
			}
		}
	}
	for _, p := range s.players {
		sutdaSortHand(p)
	}

	s.currentPlayer = (s.dealerIdx + 1) % len(s.players)
	s.phase = SutdaPhaseBet
	s.appendLog(-1, "deal", fmt.Sprintf("hand %d: dealer=%d, ante=%d, pot=%d",
		s.handNumber, s.dealerIdx, SutdaAnte, s.pot), nil)
}

// draw は山から 1 枚引く。
func (s *Sutda) draw() *Card {
	if s.drawIdx >= len(s.deck) {
		return nil
	}
	c := s.deck[s.drawIdx]
	s.drawIdx++
	return c
}

// sutdaSortHand は月の昇順に並べる (光を先に置く)。
func sutdaSortHand(p *SutdaPlayer) {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// PlayerAction は人間が 1 手打つ。
func (s *Sutda) PlayerAction(action string) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if s.phase != SutdaPhaseBet {
		return ErrWrongPhase
	}
	if !s.players[s.currentPlayer].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return s.applyAction(s.currentPlayer, action)
}

// CpuAction は CPU が 1 手打つ。
func (s *Sutda) CpuAction() {
	if s.gameEndFlag || s.phase != SutdaPhaseBet {
		return
	}
	if s.players[s.currentPlayer].GetIsHuman() {
		return
	}
	_ = s.applyAction(s.currentPlayer, s.cpuChooseAction(s.currentPlayer))
}

// applyAction は 1 手を適用する。
func (s *Sutda) applyAction(playerIdx int, action string) error {
	p := s.players[playerIdx]
	switch action {
	case SutdaActionFold:
		p.SetFolded(true)
		s.appendLog(playerIdx, "fold", fmt.Sprintf("player %d folds", playerIdx), nil)
	case SutdaActionCall:
		s.putIn(playerIdx, s.currentBet-p.GetBet())
		s.appendLog(playerIdx, "call", fmt.Sprintf("player %d calls to %d", playerIdx, s.currentBet), nil)
	case SutdaActionRaise:
		if !s.CanRaise(playerIdx) {
			return NewDomainErrorCode(ErrInvalidPlay, "sutda.errCannotRaise", nil)
		}
		s.currentBet += SutdaBetUnit
		s.raiseCount++
		s.putIn(playerIdx, s.currentBet-p.GetBet())
		// **レイズは他の席の「行動済み」を取り消す。** 取り消さないと、
		// 上げられた側がコールもフォールドもしないままハンドが終わる。
		for i := range s.actedThisRound {
			s.actedThisRound[i] = false
		}
		s.appendLog(playerIdx, "raise", fmt.Sprintf("player %d raises to %d", playerIdx, s.currentBet), nil)
	default:
		return NewDomainErrorCode(ErrInvalidPlay, "sutda.errUnknownAction",
			map[string]string{"action": action})
	}
	s.actedThisRound[playerIdx] = true
	s.advanceTurn()
	return nil
}

// putIn は席からチップを場へ出す (足りなければ有り金すべて)。
func (s *Sutda) putIn(playerIdx, amount int) {
	if amount <= 0 {
		return
	}
	p := s.players[playerIdx]
	if amount > p.GetChips() {
		amount = p.GetChips()
	}
	p.AddChips(-amount)
	p.AddBet(amount)
	s.pot += amount
}

// CanRaise はその席がいまレイズできるかを返す。
func (s *Sutda) CanRaise(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return false
	}
	if s.raiseCount >= SutdaMaxRaises {
		return false
	}
	need := s.currentBet + SutdaBetUnit - s.players[playerIdx].GetBet()
	return s.players[playerIdx].GetChips() >= need
}

// advanceTurn は次の手番へ進める。全員が行動を終えたらショーダウン。
func (s *Sutda) advanceTurn() {
	if s.activeCount() <= 1 || s.everyoneActed() {
		s.showdown()
		return
	}
	for i := 1; i <= len(s.players); i++ {
		idx := (s.currentPlayer + i) % len(s.players)
		if !s.players[idx].IsFolded() {
			s.currentPlayer = idx
			return
		}
	}
	s.showdown()
}

// activeCount は降りていない席の数を返す。
func (s *Sutda) activeCount() int {
	n := 0
	for _, p := range s.players {
		if !p.IsFolded() {
			n++
		}
	}
	return n
}

// everyoneActed は残っている全員が現在のベット額に対して行動を終えたかを返す。
func (s *Sutda) everyoneActed() bool {
	for i, p := range s.players {
		if p.IsFolded() {
			continue
		}
		if !s.actedThisRound[i] {
			return false
		}
		// **有り金が尽きた席は追随できない。** その席を待つと止まる。
		if p.GetBet() < s.currentBet && p.GetChips() > 0 {
			return false
		}
	}
	return true
}

// showdown は役を比べてポットを分ける。
func (s *Sutda) showdown() {
	hands := make([]SutdaHand, len(s.players))
	folded := make([]bool, len(s.players))
	best := -1
	for i, p := range s.players {
		folded[i] = p.IsFolded()
		if p.IsFolded() || p.GetCardsSize() < SutdaHandSize {
			hands[i] = SutdaHand{Rank: 0, Name: "none", Kkeut: -1}
			continue
		}
		hands[i] = SutdaEvaluate(p.GetCard(0), p.GetCard(1))
		p.SetRevealed(true)
		if hands[i].Rank > best {
			best = hands[i].Rank
		}
	}

	winners := make([]int, 0, len(s.players))
	for i, p := range s.players {
		if !p.IsFolded() && hands[i].Rank == best {
			winners = append(winners, i)
		}
	}

	// **同点はポットを分ける。** 席順で決めると、同じ役なのに座席で負ける。
	if len(winners) > 0 {
		share := s.pot / len(winners)
		rest := s.pot - share*len(winners)
		for n, w := range winners {
			amount := share
			if n < rest {
				amount++ // 割り切れない端数は先に配る
			}
			s.players[w].AddChips(amount)
		}
	}

	s.lastResult = &SutdaHandResult{Winners: winners, Pot: s.pot, Hands: hands, Folded: folded}
	s.appendLog(-1, "showdown", fmt.Sprintf("hand %d: pot %d to %v", s.handNumber, s.pot, winners), nil)
	// **配り終えたポットは 0 に戻す。** 残したままだと、勝者のチップと場の
	// 両方に同じ額が乗って合計が増える ── 卓からチップが湧く。勝った額は
	// lastResult.Pot が持っている。
	s.pot = 0
	s.phase = SutdaPhaseShowdown
	s.checkGameEnd()
}

// checkGameEnd はチップが尽きた席があるかを見る。
//
// **終わるのは人間が破産したときと、CPU が全員破産したとき。** どちらか
// 一方でも生き残っていれば卓は続く。
func (s *Sutda) checkGameEnd() bool {
	human := findHumanIdx(s.players)
	if human >= 0 && s.players[human].GetChips() <= 0 {
		s.finishGame(s.richestSeat())
		return true
	}
	cpuAlive := false
	for i, p := range s.players {
		if i != human && p.GetChips() > 0 {
			cpuAlive = true
			break
		}
	}
	if !cpuAlive {
		s.finishGame(human)
		return true
	}
	return false
}

// richestSeat は最もチップの多い席を返す。
func (s *Sutda) richestSeat() int {
	best := 0
	for i, p := range s.players {
		if p.GetChips() > s.players[best].GetChips() {
			best = i
		}
	}
	return best
}

// finishGame は終局する。
func (s *Sutda) finishGame(winner int) {
	s.gameEndFlag = true
	s.winnerIdx = winner
	s.phase = SutdaPhaseGameEnd
	s.appendLog(-1, "gameEnd", fmt.Sprintf("player %d takes the table", winner), nil)
}

// IsHumanTurn は人間の手番かを返す。
func (s *Sutda) IsHumanTurn() bool {
	if s.gameEndFlag || s.phase != SutdaPhaseBet {
		return false
	}
	if s.currentPlayer < 0 || s.currentPlayer >= len(s.players) {
		return false
	}
	return s.players[s.currentPlayer].GetIsHuman()
}

// GetCallAmount はその席がコールに必要な額を返す (0 ならチェック)。
func (s *Sutda) GetCallAmount(playerIdx int) int {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return 0
	}
	need := s.currentBet - s.players[playerIdx].GetBet()
	if need < 0 {
		return 0
	}
	return need
}

// SutdaHint は人間への推奨手。
type SutdaHint struct {
	// Action は勧める行動。
	Action string
	// Reason は理由の識別子。
	Reason string
}

// GetHint は人間への推奨手を返す。
func (s *Sutda) GetHint() *SutdaHint {
	human := findHumanIdx(s.players)
	if human < 0 || s.gameEndFlag {
		return &SutdaHint{Reason: "none"}
	}
	switch s.phase {
	case SutdaPhaseBet:
		if s.currentPlayer != human {
			return &SutdaHint{Reason: "none"}
		}
		// **助言は CPU の難易度に引きずらせない。** cpuChooseAction は Easy だと
		// ほぼコールを返すので、そのまま使うと Easy を選んだ人にだけ雑な助言が出る。
		action := s.smartActionFor(human)
		return &SutdaHint{Action: action, Reason: sutdaHintReason(action)}
	case SutdaPhaseShowdown:
		return &SutdaHint{Reason: "next_hand"}
	default:
		return &SutdaHint{Reason: "none"}
	}
}

// sutdaHintReason は行動に対応する理由の識別子を返す。
func sutdaHintReason(action string) string {
	switch action {
	case SutdaActionRaise:
		return "strong_hand"
	case SutdaActionFold:
		return "weak_hand"
	default:
		return "stay_in"
	}
}

// GetHandOf は席の役を返す (2 枚揃っていなければ none)。
func (s *Sutda) GetHandOf(playerIdx int) SutdaHand {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return SutdaHand{Rank: 0, Name: "none", Kkeut: -1}
	}
	p := s.players[playerIdx]
	if p.GetCardsSize() < SutdaHandSize {
		return SutdaHand{Rank: 0, Name: "none", Kkeut: -1}
	}
	return SutdaEvaluate(p.GetCard(0), p.GetCard(1))
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (s *Sutda) GetConfig() SutdaConfig { return s.config }

// SetConfig はゲーム設定を設定する。
func (s *Sutda) SetConfig(c SutdaConfig) { s.config = c }

// GetPhase は現在のフェーズを返す。
func (s *Sutda) GetPhase() string { return s.phase }

// GetPlayerCnt は席数を返す。
func (s *Sutda) GetPlayerCnt() int { return len(s.players) }

// GetPlayer は席のプレイヤーを返す。
func (s *Sutda) GetPlayer(i int) *SutdaPlayer {
	if i < 0 || i >= len(s.players) {
		return nil
	}
	return s.players[i]
}

// GetPlayers は全プレイヤーを返す。
func (s *Sutda) GetPlayers() []*SutdaPlayer { return s.players }

// GetHandNumber は現在のハンド番号を返す。
func (s *Sutda) GetHandNumber() int { return s.handNumber }

// GetDealerIdx は親の席を返す。
func (s *Sutda) GetDealerIdx() int { return s.dealerIdx }

// GetCurrentPlayerIdx は手番の席を返す。
func (s *Sutda) GetCurrentPlayerIdx() int { return s.currentPlayer }

// GetPot は場のチップを返す。
func (s *Sutda) GetPot() int { return s.pot }

// GetCurrentBet は現在のベット額を返す。
func (s *Sutda) GetCurrentBet() int { return s.currentBet }

// GetRaiseCount はこのハンドのレイズ回数を返す。
func (s *Sutda) GetRaiseCount() int { return s.raiseCount }

// GetLastResult は直前ハンドの結果を返す。
func (s *Sutda) GetLastResult() *SutdaHandResult { return s.lastResult }

// GetGameEndFlag は終局フラグを返す。
func (s *Sutda) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerIdx は最終的な勝者の席を返す (-1 = 未決)。
func (s *Sutda) GetWinnerIdx() int { return s.winnerIdx }

// --- 永続化 ---

// sutdaJSON is the JSON wire format for Sutda.
type sutdaJSON struct {
	Deck           []*Card           `json:"dk"`
	DrawIdx        int               `json:"dx"`
	Players        []*SutdaPlayer    `json:"pl"`
	Config         SutdaConfig       `json:"cf"`
	Phase          string            `json:"ph"`
	HandNumber     int               `json:"hn"`
	DealerIdx      int               `json:"di"`
	CurrentPlayer  int               `json:"cp"`
	Pot            int               `json:"pt"`
	CurrentBet     int               `json:"cb"`
	RaiseCount     int               `json:"rc"`
	ActedThisRound []bool            `json:"ar"`
	LastResult     *SutdaHandResult  `json:"lr"`
	GameEndFlag    bool              `json:"ge"`
	WinnerIdx      int               `json:"wi"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// sutdaMaxSliceLen は復元時に受け入れるスライスの上限。
const sutdaMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。**
func (s *Sutda) MarshalJSON() ([]byte, error) {
	return json.Marshal(sutdaJSON{
		Deck:           s.deck,
		DrawIdx:        s.drawIdx,
		Players:        s.players,
		Config:         s.config,
		Phase:          s.phase,
		HandNumber:     s.handNumber,
		DealerIdx:      s.dealerIdx,
		CurrentPlayer:  s.currentPlayer,
		Pot:            s.pot,
		CurrentBet:     s.currentBet,
		RaiseCount:     s.raiseCount,
		ActedThisRound: s.actedThisRound,
		LastResult:     s.lastResult,
		GameEndFlag:    s.gameEndFlag,
		WinnerIdx:      s.winnerIdx,
		ActionLog:      s.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Sutda) UnmarshalJSON(data []byte) error {
	var j sutdaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Deck) > sutdaMaxSliceLen || len(j.Players) > sutdaMaxSliceLen ||
		len(j.ActionLog) > sutdaMaxSliceLen {
		return fmt.Errorf("sutda: input array exceeds maximum allowed size")
	}
	if len(j.Players) < SutdaMinSeats || len(j.Players) > SutdaMaxSeats {
		return fmt.Errorf("sutda: invalid player count %d", len(j.Players))
	}
	s.deck = j.Deck
	s.drawIdx = j.DrawIdx
	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.handNumber = j.HandNumber
	s.dealerIdx = j.DealerIdx
	s.currentPlayer = j.CurrentPlayer
	s.pot = j.Pot
	s.currentBet = j.CurrentBet
	s.raiseCount = j.RaiseCount
	s.actedThisRound = j.ActedThisRound
	if len(s.actedThisRound) != len(s.players) {
		s.actedThisRound = make([]bool, len(s.players))
	}
	s.lastResult = j.LastResult
	s.gameEndFlag = j.GameEndFlag
	s.winnerIdx = j.WinnerIdx
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// sutdaHandResultJSON is the JSON wire format for SutdaHandResult.
type sutdaHandResultJSON struct {
	Winners []int       `json:"w"`
	Pot     int         `json:"p"`
	Hands   []SutdaHand `json:"h"`
	Folded  []bool      `json:"f"`
}

// MarshalJSON implements json.Marshaler.
func (r *SutdaHandResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(sutdaHandResultJSON{Winners: r.Winners, Pot: r.Pot, Hands: r.Hands, Folded: r.Folded})
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *SutdaHandResult) UnmarshalJSON(data []byte) error {
	var j sutdaHandResultJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	r.Winners = j.Winners
	r.Pot = j.Pot
	r.Hands = j.Hands
	r.Folded = j.Folded
	return nil
}

// sutdaRandIntn は rand.Intn の薄いラッパ (n <= 0 を握りつぶす)。
func sutdaRandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}
