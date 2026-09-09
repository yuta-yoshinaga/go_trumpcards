//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// カーンズのフェーズ定数
const (
	// QuinzePhaseBet ベットフェーズ
	QuinzePhaseBet = 1
	// QuinzePhasePlayerTurn 人間の手番
	QuinzePhasePlayerTurn = 2
	// QuinzePhaseBankerTurn 人間が親のときの引き止め判断
	QuinzePhaseBankerTurn = 3
	// QuinzePhaseEnd 精算済み
	QuinzePhaseEnd = 4
)

// Quinze の既定値
const (
	// QuinzeDefaultChips 初期チップ
	QuinzeDefaultChips = 1000
	// QuinzeMinBet 最低ベット額
	QuinzeMinBet = 10
	// QuinzeMaxBet 最大ベット額
	QuinzeMaxBet = 10000
	// QuinzeSeatCnt 座席数（人間 1 + CPU 2）
	QuinzeSeatCnt = 3
	// QuinzeDeckSize 標準デッキ（ジョーカーなし）の枚数
	QuinzeDeckSize = 52
	// quinzeMaxSliceLen JSON 復元時のスライス長上限
	quinzeMaxSliceLen = 1000
)

// QuinzeTarget 目標点。
const QuinzeTarget = 15

// QuinzeCpuStandPoints CPU 席と CPU 親がここに達したら止める。
//
// **相手の停止ラインは賭け続けるかの判断材料。**ブラックジャックの「17 で
// スタンド」に当たる数字なのに、どの画面にも出ていなかった (#5566)。案内は
// この定数を FormatPoints に通して書く。文言に数値を焼き込むと、閾値を
// 変えたとき案内だけが嘘になる。
const QuinzeCpuStandPoints = 12

// QuinzeHand 1 人の手
type QuinzeHand struct {
	cards  []*Card
	bet    int
	stood  bool
	payout int
}

// GetCards 手札を取得する
func (h *QuinzeHand) GetCards() []*Card { return h.cards }

// GetBet 賭け金を取得する
func (h *QuinzeHand) GetBet() int { return h.bet }

// IsStood 引き止めたか
func (h *QuinzeHand) IsStood() bool { return h.stood }

// GetPayout 精算後の増減
func (h *QuinzeHand) GetPayout() int { return h.payout }

// QuinzeSeat 1 つの座席。人間は seat 0。
type QuinzeSeat struct {
	name  string
	isCPU bool
	hand  *QuinzeHand
}

// GetName 席名
func (s *QuinzeSeat) GetName() string { return s.name }

// IsCPU CPU 席か
func (s *QuinzeSeat) IsCPU() bool { return s.isCPU }

// GetHand その席の手
func (s *QuinzeSeat) GetHand() *QuinzeHand { return s.hand }

// Quinze カーンズ（15を目指すフランスのバンキングゲーム）本体。
//
// 標準52枚デッキで15を目指す。Aは1点、2〜10は額面、J/Q/Kは10点。
//
// 親は固定ではない。**ちょうど15を出したプレイヤーが次の局の親になる。**
//
// **同点は親の勝ち。** したがってプレイヤーの 15 は「負けない手」ではなく、
// 「親が 15 でなければ勝つ手」でしかない。
//
// issue #4388 の仕様案とは 2 点異なり、いずれも実際の規則（pagat.com）に合わせた:
//   - 親は「次のディールで交代」ではなく、**ちょうど 15 を出した者だけ**が取る
//   - 15 は「即座に勝利」ではない。手番は終わって公開されるが、**親も 15 なら
//     同点で親の勝ち**
type Quinze struct {
	trumpCards *TrumpCards
	seats      []*QuinzeSeat
	banker     int
	bankerHand *QuinzeHand
	chips      ChipHolder
	activeSeat int
	phase      int
	// nextBanker はこの局で 15 を出した最初のプレイヤー（いなければ -1）。
	nextBanker int
	lastResult string
	actionLog  []*ActionLogEntry
}

// quinzeOpeningBanker 最初の局の親。
//
// 人間 (seat 0) にしないのは、そこから始めると新規セッションがいきなり「あなたが
// 配ってください」で開き、このゲームの主要な流れ — 賭けて引く — に入れないため。
// 誰が最初に持つかは規則が定めていないので、遊びやすい方を選ぶ。
const quinzeOpeningBanker = 1

// NewQuinze コンストラクタ
func NewQuinze(trumpCards *TrumpCards) *Quinze {
	s := &Quinze{
		trumpCards: trumpCards,
		phase:      QuinzePhaseBet,
		banker:     quinzeOpeningBanker,
		nextBanker: -1,
	}
	s.chips.SetChips(QuinzeDefaultChips)
	return s
}

// NewDefaultQuinze returns Quinze with a standard 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultQuinze() *Quinze {
	return NewQuinze(NewTrumpCards(0))
}

// Reset 新しい局を始める。親は前局から引き継ぐ。
func (s *Quinze) Reset() {
	if s.chips.GetChips() < QuinzeMinBet {
		s.chips.SetChips(QuinzeDefaultChips)
	}
	if s.seats == nil {
		s.seats = make([]*QuinzeSeat, QuinzeSeatCnt)
		for i := range QuinzeSeatCnt {
			s.seats[i] = &QuinzeSeat{name: quinzeSeatName(i), isCPU: i != 0}
		}
	}
	if s.nextBanker >= 0 && s.nextBanker < len(s.seats) {
		s.banker = s.nextBanker
	}
	s.nextBanker = -1

	s.trumpCards.Shuffle()
	for _, seat := range s.seats {
		seat.hand = nil
	}
	s.bankerHand = nil
	s.activeSeat = 0
	s.phase = QuinzePhaseBet
	s.lastResult = ""
	s.actionLog = nil
}

// quinzeSeatName 席の既定名
func quinzeSeatName(i int) string {
	if i == 0 {
		return "あなた"
	}
	return fmt.Sprintf("CPU%d", i)
}

// PlaceBet 人間のベットを置き、その局を配る。
func (s *Quinze) PlaceBet(bet int) error {
	if s.phase != QuinzePhaseBet {
		return errors.New("quinze: not in the betting phase")
	}
	if s.IsHumanBanker() {
		return errors.New("quinze: the banker does not place a bet")
	}
	if bet < QuinzeMinBet || bet > QuinzeMaxBet {
		return fmt.Errorf("quinze: bet must be between %d and %d", QuinzeMinBet, QuinzeMaxBet)
	}
	if bet > s.chips.GetChips() {
		return errors.New("quinze: not enough chips")
	}
	s.chips.SetChips(s.chips.GetChips() - bet)
	s.deal(bet)
	return nil
}

// StartAsBanker 人間が親の局を配る。親は賭けない。
func (s *Quinze) StartAsBanker() error {
	if s.phase != QuinzePhaseBet {
		return errors.New("quinze: not in the betting phase")
	}
	if !s.IsHumanBanker() {
		return errors.New("quinze: the human is not the banker")
	}
	s.deal(0)
	return nil
}

// deal 全員に 1 枚ずつ配り、CPU の手番を進める。
func (s *Quinze) deal(humanBet int) {
	for i, seat := range s.seats {
		if i == s.banker {
			continue
		}
		bet := humanBet
		if seat.isCPU {
			bet = QuinzeMinBet * 2
		}
		seat.hand = &QuinzeHand{cards: s.drawOne(), bet: bet}
	}
	s.bankerHand = &QuinzeHand{cards: s.drawOne()}
	s.appendLog("deal", "全員に1枚ずつ配った", nil)

	s.phase = QuinzePhasePlayerTurn
	s.activeSeat = 0
	s.advanceToHuman()
}

// drawOne 1 枚引いてスライスにする
func (s *Quinze) drawOne() []*Card {
	if c := s.trumpCards.DrawCard(); c != nil {
		return []*Card{c}
	}
	return nil
}

// advanceToHuman 人間の番が来るまで CPU を自動で進める。
func (s *Quinze) advanceToHuman() {
	for s.phase == QuinzePhasePlayerTurn {
		if s.activeSeat >= len(s.seats) {
			s.startBankerTurn()
			return
		}
		seat := s.seats[s.activeSeat]
		if s.activeSeat == s.banker || seat.hand == nil {
			s.activeSeat++
			continue
		}
		if !seat.isCPU {
			return
		}
		s.playCpuSeat(seat.hand)
		s.activeSeat++
	}
}

// playCpuSeat CPU の手を最後まで打つ。停止ライン未満なら引く。
func (s *Quinze) playCpuSeat(h *QuinzeHand) {
	for {
		total := s.handPoints(h)
		if total >= QuinzeTarget {
			break
		}
		if total >= QuinzeCpuStandPoints {
			h.stood = true
			break
		}
		if !s.hit(h) {
			break
		}
	}
	h.stood = true
}

// hit 1 枚引く。山が尽きたら false。
func (s *Quinze) hit(h *QuinzeHand) bool {
	c := s.trumpCards.DrawCard()
	if c == nil {
		return false
	}
	h.cards = append(h.cards, c)
	return true
}

// Hit 人間が 1 枚引く。
func (s *Quinze) Hit() error {
	h, err := s.currentHand()
	if err != nil {
		return err
	}
	if s.handPoints(h) >= QuinzeTarget {
		return errors.New("quinze: cannot draw on fifteen or more")
	}
	if !s.hit(h) {
		return errors.New("quinze: the deck is empty")
	}
	s.appendLog("hit", "1枚引いた", h.cards)
	// バーストしたか、ちょうど 15 に届いたらその手は終わり。15 は「見せて
	// 手番を終える」手であって、即座の勝ちではない。
	if s.handPoints(h) >= QuinzeTarget {
		h.stood = true
		s.nextSeat()
	}
	return nil
}

// Stand 人間が引き止める。
func (s *Quinze) Stand() error {
	h, err := s.currentHand()
	if err != nil {
		return err
	}
	h.stood = true
	s.appendLog("stand", "スタンド", h.cards)
	s.nextSeat()
	return nil
}

// currentHand 手番の手を返す
func (s *Quinze) currentHand() (*QuinzeHand, error) {
	if s.phase != QuinzePhasePlayerTurn {
		return nil, errors.New("quinze: not the player's turn")
	}
	if s.activeSeat != 0 || s.activeSeat == s.banker {
		return nil, errors.New("quinze: not the human's turn")
	}
	if s.seats[0].hand == nil {
		return nil, errors.New("quinze: no hand in play")
	}
	return s.seats[0].hand, nil
}

// nextSeat 次の席へ進める
func (s *Quinze) nextSeat() {
	s.activeSeat++
	s.advanceToHuman()
}

// startBankerTurn 親の手番へ。CPU が親なら自動で打って精算まで進む。
func (s *Quinze) startBankerTurn() {
	if s.IsHumanBanker() {
		s.phase = QuinzePhaseBankerTurn
		return
	}
	s.playCpuSeat(s.bankerHand)
	s.settle()
}

// BankerHit 人間が親のときに 1 枚引く。
func (s *Quinze) BankerHit() error {
	if s.phase != QuinzePhaseBankerTurn {
		return errors.New("quinze: not the banker's turn")
	}
	if s.handPoints(s.bankerHand) >= QuinzeTarget {
		return errors.New("quinze: cannot draw on fifteen or more")
	}
	if !s.hit(s.bankerHand) {
		return errors.New("quinze: the deck is empty")
	}
	s.appendLog("bankerHit", "親が1枚引いた", s.bankerHand.cards)
	if s.handPoints(s.bankerHand) >= QuinzeTarget {
		s.settle()
	}
	return nil
}

// BankerStand 人間が親のときに引き止めて精算する。
func (s *Quinze) BankerStand() error {
	if s.phase != QuinzePhaseBankerTurn {
		return errors.New("quinze: not the banker's turn")
	}
	s.settle()
	return nil
}

// settle 全席を精算する。
func (s *Quinze) settle() {
	s.bankerHand.stood = true
	bankerPoints := s.handPoints(s.bankerHand)
	bankerBust := bankerPoints > QuinzeTarget

	for i, seat := range s.seats {
		if i == s.banker || seat.hand == nil {
			continue
		}
		h := seat.hand
		h.payout = s.settleHand(h, bankerPoints, bankerBust)
		if s.IsHumanBanker() {
			s.chips.SetChips(s.chips.GetChips() - h.payout)
		} else if i == 0 {
			s.chips.SetChips(s.chips.GetChips() + h.bet + h.payout)
		}
		// ちょうど 15 を出した最初のプレイヤーが次の親になる。
		if s.handPoints(h) == QuinzeTarget && s.nextBanker < 0 {
			s.nextBanker = i
		}
	}
	s.phase = QuinzePhaseEnd
	s.lastResult = s.describeResult(bankerPoints, bankerBust)
	s.appendLog("result", s.lastResult, s.bankerHand.cards)
}

// settleHand 1 つの手の増減（賭け金を除いた純増減）。同点は親の勝ち。
func (s *Quinze) settleHand(h *QuinzeHand, bankerPoints int, bankerBust bool) int {
	total := s.handPoints(h)
	if total > QuinzeTarget {
		// 自分がバーストしていれば、親のバーストは関係なく負け。
		return -h.bet
	}
	if bankerBust {
		return h.bet
	}
	if total > bankerPoints {
		return h.bet
	}
	return -h.bet
}

// describeResult 精算の要約
func (s *Quinze) describeResult(bankerPoints int, bankerBust bool) string {
	if bankerBust {
		return fmt.Sprintf("親がバースト（%s）", quinzeFormatPoints(bankerPoints))
	}
	return fmt.Sprintf("親は %s", quinzeFormatPoints(bankerPoints))
}

// quinzeCardPoints returns the card value used by Quinze.
func quinzeCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v >= 11 {
		return 10
	}
	return v
}

// quinzeBasePoints returns the sum of the cards in a hand.
func quinzeBasePoints(cards []*Card) int {
	total := 0
	for _, c := range cards {
		total += quinzeCardPoints(c)
	}
	return total
}

// handPoints returns the integer total of a hand.
func (s *Quinze) handPoints(h *QuinzeHand) int {
	if h == nil {
		return 0
	}
	return quinzeBasePoints(h.cards)
}

// quinzeFormatPoints formats an integer Quinze total.
func quinzeFormatPoints(points int) string {
	return fmt.Sprintf("%d", points)
}

// GetHandPoints returns the integer hand total for display.
func (s *Quinze) GetHandPoints(h *QuinzeHand) int { return s.handPoints(h) }

// FormatPoints formats an integer hand total.
func (s *Quinze) FormatPoints(points int) string { return quinzeFormatPoints(points) }

// GetPhase フェーズ取得
func (s *Quinze) GetPhase() int { return s.phase }

// GetChips 人間のチップ
func (s *Quinze) GetChips() int { return s.chips.GetChips() }

// GetSeats 全席を取得する
func (s *Quinze) GetSeats() []*QuinzeSeat { return s.seats }

// GetBankerIdx 親の席番号
func (s *Quinze) GetBankerIdx() int { return s.banker }

// IsHumanBanker 人間が親か
func (s *Quinze) IsHumanBanker() bool { return s.banker == 0 }

// GetBankerHand 親の手
func (s *Quinze) GetBankerHand() *QuinzeHand { return s.bankerHand }

// GetActiveSeat 手番の席
func (s *Quinze) GetActiveSeat() int { return s.activeSeat }

// GetNextBanker 次局の親（未定なら -1）
func (s *Quinze) GetNextBanker() int { return s.nextBanker }

// GetLastResult 直近の精算の要約
func (s *Quinze) GetLastResult() string { return s.lastResult }

// GetActionLog 棋譜取得
func (s *Quinze) GetActionLog() []*ActionLogEntry { return s.actionLog }

// GetGameEndFlag 局が終わっているか
func (s *Quinze) GetGameEndFlag() bool { return s.phase == QuinzePhaseEnd }

// CanHit 今の手で引けるか
func (s *Quinze) CanHit() bool {
	h, err := s.currentHand()
	return err == nil && s.handPoints(h) < QuinzeTarget
}

// CanStand 今の手で止められるか
func (s *Quinze) CanStand() bool {
	_, err := s.currentHand()
	return err == nil
}

// appendLog 棋譜エントリを追加
func (s *Quinze) appendLog(actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: len(s.actionLog),
		PlayerIdx:  s.activeSeat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      append([]*Card(nil), cards...),
	})
}

// quinzeHandJSON is the wire format for one hand. QuinzeHand's fields
// are unexported, so marshalling it directly would emit `{}` and every hand
// would come back from KV empty.
type quinzeHandJSON struct {
	Cards  []*Card `json:"cd"`
	Bet    int     `json:"bt"`
	Stood  bool    `json:"sd"`
	Payout int     `json:"po"`
}

// MarshalJSON implements json.Marshaler for QuinzeHand.
func (h *QuinzeHand) MarshalJSON() ([]byte, error) {
	return json.Marshal(quinzeHandJSON{
		Cards:  h.cards,
		Bet:    h.bet,
		Stood:  h.stood,
		Payout: h.payout,
	})
}

// UnmarshalJSON implements json.Unmarshaler for QuinzeHand.
func (h *QuinzeHand) UnmarshalJSON(data []byte) error {
	var j quinzeHandJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Cards) > quinzeMaxSliceLen {
		return errors.New("quinze: hand exceeds maximum allowed size")
	}
	h.cards = j.Cards
	h.bet = j.Bet
	h.stood = j.Stood
	h.payout = j.Payout
	return nil
}

// quinzeSeatJSON is the wire format for one seat.
type quinzeSeatJSON struct {
	Name  string      `json:"nm"`
	IsCPU bool        `json:"cp"`
	Hand  *QuinzeHand `json:"hd,omitempty"`
}

// MarshalJSON implements json.Marshaler for QuinzeSeat.
func (s *QuinzeSeat) MarshalJSON() ([]byte, error) {
	return json.Marshal(quinzeSeatJSON{Name: s.name, IsCPU: s.isCPU, Hand: s.hand})
}

// UnmarshalJSON implements json.Unmarshaler for QuinzeSeat.
func (s *QuinzeSeat) UnmarshalJSON(data []byte) error {
	var j quinzeSeatJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	s.name = j.Name
	s.isCPU = j.IsCPU
	s.hand = j.Hand
	return nil
}

// quinzeJSON is the JSON wire format for Quinze.
type quinzeJSON struct {
	TrumpCards *TrumpCards       `json:"tc"`
	Seats      []*QuinzeSeat     `json:"st"`
	Banker     int               `json:"bk"`
	BankerHand *QuinzeHand       `json:"bh"`
	Chips      int               `json:"ch"`
	ActiveSeat int               `json:"as"`
	Phase      int               `json:"ph"`
	NextBanker int               `json:"nb"`
	LastResult string            `json:"lr"`
	ActionLog  []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (s *Quinze) MarshalJSON() ([]byte, error) {
	return json.Marshal(quinzeJSON{
		TrumpCards: s.trumpCards,
		Seats:      s.seats,
		Banker:     s.banker,
		BankerHand: s.bankerHand,
		Chips:      s.chips.GetChips(),
		ActiveSeat: s.activeSeat,
		Phase:      s.phase,
		NextBanker: s.nextBanker,
		LastResult: s.lastResult,
		ActionLog:  s.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (s *Quinze) UnmarshalJSON(data []byte) error {
	var j quinzeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < QuinzePhaseBet || j.Phase > QuinzePhaseEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if len(j.Seats) > QuinzeSeatCnt {
		return fmt.Errorf("invalid seat count: %d", len(j.Seats))
	}
	if j.Banker < 0 || j.Banker >= max(len(j.Seats), 1) {
		return fmt.Errorf("invalid banker: %d", j.Banker)
	}
	if j.NextBanker < -1 || j.NextBanker >= max(len(j.Seats), 1) {
		return fmt.Errorf("invalid next banker: %d", j.NextBanker)
	}
	if j.ActiveSeat < 0 || j.ActiveSeat > len(j.Seats) {
		return fmt.Errorf("invalid active seat: %d", j.ActiveSeat)
	}
	if j.Chips < 0 {
		return fmt.Errorf("invalid chips: %d", j.Chips)
	}
	if len(j.ActionLog) > quinzeMaxSliceLen {
		return errors.New("quinze: action log exceeds maximum allowed size")
	}
	if j.TrumpCards != nil {
		s.trumpCards = j.TrumpCards
	}
	s.seats = j.Seats
	s.banker = j.Banker
	s.bankerHand = j.BankerHand
	s.chips.SetChips(j.Chips)
	s.activeSeat = j.ActiveSeat
	s.phase = j.Phase
	s.nextBanker = j.NextBanker
	s.lastResult = j.LastResult
	s.actionLog = j.ActionLog
	return nil
}
