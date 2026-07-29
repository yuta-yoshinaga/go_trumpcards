//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// セッテ・エ・メッツォのフェーズ定数
const (
	// SetteEMezzoPhaseBet ベットフェーズ
	SetteEMezzoPhaseBet = 1
	// SetteEMezzoPhasePlayerTurn 人間の手番
	SetteEMezzoPhasePlayerTurn = 2
	// SetteEMezzoPhaseBankerTurn 人間が親のときの引き止め判断
	SetteEMezzoPhaseBankerTurn = 3
	// SetteEMezzoPhaseEnd 精算済み
	SetteEMezzoPhaseEnd = 4
)

// セッテ・エ・メッツォの既定値
const (
	// SetteEMezzoDefaultChips 初期チップ
	SetteEMezzoDefaultChips = 1000
	// SetteEMezzoMinBet 最低ベット額
	SetteEMezzoMinBet = 10
	// SetteEMezzoMaxBet 最大ベット額
	SetteEMezzoMaxBet = 10000
	// SetteEMezzoSeatCnt 座席数（人間 1 + CPU 2）
	SetteEMezzoSeatCnt = 3
	// SetteEMezzoDeckSize 40 枚（8・9・10 を抜いたイタリア式相当）
	SetteEMezzoDeckSize = 40
	// setteEMezzoMaxSliceLen JSON 復元時のスライス長上限
	setteEMezzoMaxSliceLen = 1000
)

// SetteEMezzoTargetHalves 目標点 7.5 を**半点単位**で表したもの。
//
// 絵札が 0.5 点なので合計は必ず 0.5 の倍数になる。float64 で持つと
// 0.5 の加算で誤差が出る計算ではないものの、比較と等値判定が値の表現に
// 依存する形になるため、全体を 2 倍した整数で扱う。
const SetteEMezzoTargetHalves = 15

// SetteEMezzoMattaDesign マッタ（コインの K）に対応するスート。
// フレンチスートではコイン＝ダイヤに対応させる。
const SetteEMezzoMattaDesign = CardDesignDiamond

// SetteEMezzoMattaValue マッタのランク（K）
const SetteEMezzoMattaValue = 13

// SetteEMezzoHand 1 人の手
type SetteEMezzoHand struct {
	cards []*Card
	bet   int
	// mattaHalves はマッタに割り当てた値（半点単位）。未指定なら 0。
	mattaHalves int
	stood       bool
	payout      int
}

// GetCards 手札を取得する
func (h *SetteEMezzoHand) GetCards() []*Card { return h.cards }

// GetBet 賭け金を取得する
func (h *SetteEMezzoHand) GetBet() int { return h.bet }

// GetMattaHalves マッタに割り当てた値（半点単位・未指定なら 0）
func (h *SetteEMezzoHand) GetMattaHalves() int { return h.mattaHalves }

// IsStood 引き止めたか
func (h *SetteEMezzoHand) IsStood() bool { return h.stood }

// GetPayout 精算後の増減
func (h *SetteEMezzoHand) GetPayout() int { return h.payout }

// HasMatta マッタを持っているか
func (h *SetteEMezzoHand) HasMatta() bool {
	return setteEMezzoIndexOfMatta(h.cards) >= 0
}

// SetteEMezzoSeat 1 つの座席。人間は seat 0。
type SetteEMezzoSeat struct {
	name  string
	isCPU bool
	hand  *SetteEMezzoHand
}

// GetName 席名
func (s *SetteEMezzoSeat) GetName() string { return s.name }

// IsCPU CPU 席か
func (s *SetteEMezzoSeat) IsCPU() bool { return s.isCPU }

// GetHand その席の手
func (s *SetteEMezzoSeat) GetHand() *SetteEMezzoHand { return s.hand }

// SetteEMezzo セッテ・エ・メッツォ（7 と 1/2）ゲーム本体。
//
// 40 枚（8・9・10 を抜いた構成）で 7.5 を目指す、イタリアのバンキングゲーム。
// **絵札が 0.5 点**なので、ブラックジャックのように「もう 1 枚で必ず大きく動く」
// ことがなく、半端点をどう埋めるかがそのまま駆け引きになる。
//
// **マッタ（コインの K）はワイルド**で、0.5 点または 1〜7 点のどれにでもできる。
// 引いた瞬間に値が決まるのではなく、手を止めるまで好きな値に付け替えられる。
//
// 親は固定ではない。**ちょうど 7.5 を出したプレイヤーが次の局の親になる。**
// 毎局まわるのではなく、7.5 を出した者だけが取る。
//
// **同点は親の勝ち。** したがってプレイヤーの 7.5 は「負けない手」ではなく、
// 「親が 7.5 でなければ勝つ手」でしかない。
//
// issue #4388 の仕様案とは 2 点異なり、いずれも実際の規則（pagat.com）に合わせた:
//   - 親は「次のディールで交代」ではなく、**ちょうど 7.5 を出した者だけ**が取る
//   - 7.5 は「即座に勝利」ではない。手番は終わって公開されるが、**親も 7.5 なら
//     同点で親の勝ち**
type SetteEMezzo struct {
	trumpCards *TrumpCards
	seats      []*SetteEMezzoSeat
	banker     int
	bankerHand *SetteEMezzoHand
	chips      ChipHolder
	activeSeat int
	phase      int
	// nextBanker はこの局で 7.5 を出した最初のプレイヤー（いなければ -1）。
	nextBanker int
	lastResult string
	actionLog  []*ActionLogEntry
}

// NewSetteEMezzo コンストラクタ
func NewSetteEMezzo(trumpCards *TrumpCards) *SetteEMezzo {
	s := &SetteEMezzo{trumpCards: trumpCards, phase: SetteEMezzoPhaseBet, nextBanker: -1}
	s.chips.SetChips(SetteEMezzoDefaultChips)
	return s
}

// NewDefaultSetteEMezzo returns SetteEMezzo with the 40-card Italian-style deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSetteEMezzo() *SetteEMezzo {
	return NewSetteEMezzo(NewTrumpCardsScopa())
}

// Reset 新しい局を始める。親は前局から引き継ぐ。
func (s *SetteEMezzo) Reset() {
	if s.chips.GetChips() < SetteEMezzoMinBet {
		s.chips.SetChips(SetteEMezzoDefaultChips)
	}
	if s.seats == nil {
		s.seats = make([]*SetteEMezzoSeat, SetteEMezzoSeatCnt)
		for i := range SetteEMezzoSeatCnt {
			s.seats[i] = &SetteEMezzoSeat{name: setteEMezzoSeatName(i), isCPU: i != 0}
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
	s.phase = SetteEMezzoPhaseBet
	s.lastResult = ""
	s.actionLog = nil
}

// setteEMezzoSeatName 席の既定名
func setteEMezzoSeatName(i int) string {
	if i == 0 {
		return "あなた"
	}
	return fmt.Sprintf("CPU%d", i)
}

// PlaceBet 人間のベットを置き、その局を配る。
func (s *SetteEMezzo) PlaceBet(bet int) error {
	if s.phase != SetteEMezzoPhaseBet {
		return errors.New("settemezzo: not in the betting phase")
	}
	if s.IsHumanBanker() {
		return errors.New("settemezzo: the banker does not place a bet")
	}
	if bet < SetteEMezzoMinBet || bet > SetteEMezzoMaxBet {
		return fmt.Errorf("settemezzo: bet must be between %d and %d", SetteEMezzoMinBet, SetteEMezzoMaxBet)
	}
	if bet > s.chips.GetChips() {
		return errors.New("settemezzo: not enough chips")
	}
	s.chips.SetChips(s.chips.GetChips() - bet)
	s.deal(bet)
	return nil
}

// StartAsBanker 人間が親の局を配る。親は賭けない。
func (s *SetteEMezzo) StartAsBanker() error {
	if s.phase != SetteEMezzoPhaseBet {
		return errors.New("settemezzo: not in the betting phase")
	}
	if !s.IsHumanBanker() {
		return errors.New("settemezzo: the human is not the banker")
	}
	s.deal(0)
	return nil
}

// deal 全員に 1 枚ずつ配り、CPU の手番を進める。
func (s *SetteEMezzo) deal(humanBet int) {
	for i, seat := range s.seats {
		if i == s.banker {
			continue
		}
		bet := humanBet
		if seat.isCPU {
			bet = SetteEMezzoMinBet * 2
		}
		seat.hand = &SetteEMezzoHand{cards: s.drawOne(), bet: bet}
	}
	s.bankerHand = &SetteEMezzoHand{cards: s.drawOne()}
	s.appendLog("deal", "全員に1枚ずつ配った", nil)

	s.phase = SetteEMezzoPhasePlayerTurn
	s.activeSeat = 0
	s.advanceToHuman()
}

// drawOne 1 枚引いてスライスにする
func (s *SetteEMezzo) drawOne() []*Card {
	if c := s.trumpCards.DrawCard(); c != nil {
		return []*Card{c}
	}
	return nil
}

// advanceToHuman 人間の番が来るまで CPU を自動で進める。
func (s *SetteEMezzo) advanceToHuman() {
	for s.phase == SetteEMezzoPhasePlayerTurn {
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

// playCpuSeat CPU の手を最後まで打つ。5.5 未満なら引く、という素直な方針。
func (s *SetteEMezzo) playCpuSeat(h *SetteEMezzoHand) {
	for {
		s.autoAssignMatta(h)
		total := s.handHalves(h)
		if total >= SetteEMezzoTargetHalves {
			break
		}
		// 11 半点 = 5.5 点。ここを超えたら止める。
		if total >= 11 {
			h.stood = true
			break
		}
		if !s.hit(h) {
			break
		}
	}
	h.stood = true
	s.autoAssignMatta(h)
}

// hit 1 枚引く。山が尽きたら false。
func (s *SetteEMezzo) hit(h *SetteEMezzoHand) bool {
	c := s.trumpCards.DrawCard()
	if c == nil {
		return false
	}
	h.cards = append(h.cards, c)
	return true
}

// autoAssignMatta マッタの値を、バーストしない範囲で最大になるよう自動で決める。
// 人間は SetMattaValue で上書きできる。
func (s *SetteEMezzo) autoAssignMatta(h *SetteEMezzoHand) {
	if setteEMezzoIndexOfMatta(h.cards) < 0 {
		h.mattaHalves = 0
		return
	}
	base := setteEMezzoBaseHalves(h.cards)
	best := 1 // 0.5 点。マッタが取りうる最小値。
	for v := 2; v <= 14; v += 2 {
		if base+v <= SetteEMezzoTargetHalves {
			best = v
		}
	}
	h.mattaHalves = best
}

// Hit 人間が 1 枚引く。
func (s *SetteEMezzo) Hit() error {
	h, err := s.currentHand()
	if err != nil {
		return err
	}
	if s.handHalves(h) >= SetteEMezzoTargetHalves {
		return errors.New("settemezzo: cannot draw on seven and a half or more")
	}
	if !s.hit(h) {
		return errors.New("settemezzo: the deck is empty")
	}
	s.autoAssignMatta(h)
	s.appendLog("hit", "1枚引いた", h.cards)
	// バーストしたか、ちょうど 7.5 に届いたらその手は終わり。7.5 は「見せて
	// 手番を終える」手であって、即座の勝ちではない。
	if s.handHalves(h) >= SetteEMezzoTargetHalves {
		h.stood = true
		s.nextSeat()
	}
	return nil
}

// Stand 人間が引き止める。
func (s *SetteEMezzo) Stand() error {
	h, err := s.currentHand()
	if err != nil {
		return err
	}
	h.stood = true
	s.appendLog("stand", "スタンド", h.cards)
	s.nextSeat()
	return nil
}

// SetMattaValue マッタに割り当てる値を選ぶ（半点単位で 1、または 2〜14 の偶数）。
//
// 引いた瞬間に固定されないのがマッタの要点で、手を止めるまで付け替えられる。
func (s *SetteEMezzo) SetMattaValue(halves int) error {
	h, err := s.currentHand()
	if err != nil {
		return err
	}
	if setteEMezzoIndexOfMatta(h.cards) < 0 {
		return errors.New("settemezzo: this hand holds no matta")
	}
	if halves != 1 && (halves < 2 || halves > 14 || halves%2 != 0) {
		return errors.New("settemezzo: the matta is worth 0.5 or a whole number from 1 to 7")
	}
	h.mattaHalves = halves
	s.appendLog("matta", fmt.Sprintf("マッタを %s 点にした", setteEMezzoFormatHalves(halves)), h.cards)
	return nil
}

// currentHand 手番の手を返す
func (s *SetteEMezzo) currentHand() (*SetteEMezzoHand, error) {
	if s.phase != SetteEMezzoPhasePlayerTurn {
		return nil, errors.New("settemezzo: not the player's turn")
	}
	if s.activeSeat != 0 || s.activeSeat == s.banker {
		return nil, errors.New("settemezzo: not the human's turn")
	}
	if s.seats[0].hand == nil {
		return nil, errors.New("settemezzo: no hand in play")
	}
	return s.seats[0].hand, nil
}

// nextSeat 次の席へ進める
func (s *SetteEMezzo) nextSeat() {
	s.activeSeat++
	s.advanceToHuman()
}

// startBankerTurn 親の手番へ。CPU が親なら自動で打って精算まで進む。
func (s *SetteEMezzo) startBankerTurn() {
	if s.IsHumanBanker() {
		s.phase = SetteEMezzoPhaseBankerTurn
		s.autoAssignMatta(s.bankerHand)
		return
	}
	s.playCpuSeat(s.bankerHand)
	s.settle()
}

// BankerHit 人間が親のときに 1 枚引く。
func (s *SetteEMezzo) BankerHit() error {
	if s.phase != SetteEMezzoPhaseBankerTurn {
		return errors.New("settemezzo: not the banker's turn")
	}
	if s.handHalves(s.bankerHand) >= SetteEMezzoTargetHalves {
		return errors.New("settemezzo: cannot draw on seven and a half or more")
	}
	if !s.hit(s.bankerHand) {
		return errors.New("settemezzo: the deck is empty")
	}
	s.autoAssignMatta(s.bankerHand)
	s.appendLog("bankerHit", "親が1枚引いた", s.bankerHand.cards)
	if s.handHalves(s.bankerHand) >= SetteEMezzoTargetHalves {
		s.settle()
	}
	return nil
}

// BankerStand 人間が親のときに引き止めて精算する。
func (s *SetteEMezzo) BankerStand() error {
	if s.phase != SetteEMezzoPhaseBankerTurn {
		return errors.New("settemezzo: not the banker's turn")
	}
	s.settle()
	return nil
}

// settle 全席を精算する。
func (s *SetteEMezzo) settle() {
	s.bankerHand.stood = true
	bankerHalves := s.handHalves(s.bankerHand)
	bankerBust := bankerHalves > SetteEMezzoTargetHalves

	for i, seat := range s.seats {
		if i == s.banker || seat.hand == nil {
			continue
		}
		h := seat.hand
		h.payout = s.settleHand(h, bankerHalves, bankerBust)
		if s.IsHumanBanker() {
			s.chips.SetChips(s.chips.GetChips() - h.payout)
		} else if i == 0 {
			s.chips.SetChips(s.chips.GetChips() + h.bet + h.payout)
		}
		// ちょうど 7.5 を出した最初のプレイヤーが次の親になる。
		if s.handHalves(h) == SetteEMezzoTargetHalves && s.nextBanker < 0 {
			s.nextBanker = i
		}
	}
	s.phase = SetteEMezzoPhaseEnd
	s.lastResult = s.describeResult(bankerHalves, bankerBust)
	s.appendLog("result", s.lastResult, s.bankerHand.cards)
}

// settleHand 1 つの手の増減（賭け金を除いた純増減）。同点は親の勝ち。
func (s *SetteEMezzo) settleHand(h *SetteEMezzoHand, bankerHalves int, bankerBust bool) int {
	total := s.handHalves(h)
	if total > SetteEMezzoTargetHalves {
		// 自分がバーストしていれば、親のバーストは関係なく負け。
		return -h.bet
	}
	if bankerBust {
		return h.bet
	}
	if total > bankerHalves {
		return h.bet
	}
	return -h.bet
}

// describeResult 精算の要約
func (s *SetteEMezzo) describeResult(bankerHalves int, bankerBust bool) string {
	if bankerBust {
		return fmt.Sprintf("親がバースト（%s）", setteEMezzoFormatHalves(bankerHalves))
	}
	return fmt.Sprintf("親は %s", setteEMezzoFormatHalves(bankerHalves))
}

// setteEMezzoIndexOfMatta マッタの位置（無ければ -1）
func setteEMezzoIndexOfMatta(cards []*Card) int {
	for i, c := range cards {
		if c != nil && c.GetDesign() == SetteEMezzoMattaDesign && c.GetValue() == SetteEMezzoMattaValue {
			return i
		}
	}
	return -1
}

// setteEMezzoCardHalves 1 枚の点数（半点単位）。A は 1 点、2〜7 は数どおり、
// 絵札は 0.5 点。マッタはここでは数えない。
func setteEMezzoCardHalves(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v >= 11 {
		return 1
	}
	return v * 2
}

// setteEMezzoBaseHalves マッタを除いた合計（半点単位）
func setteEMezzoBaseHalves(cards []*Card) int {
	mattaIdx := setteEMezzoIndexOfMatta(cards)
	total := 0
	for i, c := range cards {
		if i == mattaIdx {
			continue
		}
		total += setteEMezzoCardHalves(c)
	}
	return total
}

// handHalves 手の合計（半点単位）。マッタは割り当て値で数え、未指定なら 0.5 点。
func (s *SetteEMezzo) handHalves(h *SetteEMezzoHand) int {
	if h == nil {
		return 0
	}
	total := setteEMezzoBaseHalves(h.cards)
	if setteEMezzoIndexOfMatta(h.cards) >= 0 {
		if h.mattaHalves > 0 {
			total += h.mattaHalves
		} else {
			total++
		}
	}
	return total
}

// setteEMezzoFormatHalves 半点単位を「7.5」「3」のような表示にする
func setteEMezzoFormatHalves(halves int) string {
	if halves%2 == 0 {
		return fmt.Sprintf("%d", halves/2)
	}
	return fmt.Sprintf("%d.5", halves/2)
}

// GetHandHalves 手の合計を半点単位で取得する（表示用）
func (s *SetteEMezzo) GetHandHalves(h *SetteEMezzoHand) int { return s.handHalves(h) }

// FormatHalves 半点単位を表示文字列にする
func (s *SetteEMezzo) FormatHalves(halves int) string { return setteEMezzoFormatHalves(halves) }

// GetPhase フェーズ取得
func (s *SetteEMezzo) GetPhase() int { return s.phase }

// GetChips 人間のチップ
func (s *SetteEMezzo) GetChips() int { return s.chips.GetChips() }

// GetSeats 全席を取得する
func (s *SetteEMezzo) GetSeats() []*SetteEMezzoSeat { return s.seats }

// GetBankerIdx 親の席番号
func (s *SetteEMezzo) GetBankerIdx() int { return s.banker }

// IsHumanBanker 人間が親か
func (s *SetteEMezzo) IsHumanBanker() bool { return s.banker == 0 }

// GetBankerHand 親の手
func (s *SetteEMezzo) GetBankerHand() *SetteEMezzoHand { return s.bankerHand }

// GetActiveSeat 手番の席
func (s *SetteEMezzo) GetActiveSeat() int { return s.activeSeat }

// GetNextBanker 次局の親（未定なら -1）
func (s *SetteEMezzo) GetNextBanker() int { return s.nextBanker }

// GetLastResult 直近の精算の要約
func (s *SetteEMezzo) GetLastResult() string { return s.lastResult }

// GetActionLog 棋譜取得
func (s *SetteEMezzo) GetActionLog() []*ActionLogEntry { return s.actionLog }

// GetGameEndFlag 局が終わっているか
func (s *SetteEMezzo) GetGameEndFlag() bool { return s.phase == SetteEMezzoPhaseEnd }

// CanHit 今の手で引けるか
func (s *SetteEMezzo) CanHit() bool {
	h, err := s.currentHand()
	return err == nil && s.handHalves(h) < SetteEMezzoTargetHalves
}

// CanStand 今の手で止められるか
func (s *SetteEMezzo) CanStand() bool {
	_, err := s.currentHand()
	return err == nil
}

// CanSetMatta マッタの値を選べるか
func (s *SetteEMezzo) CanSetMatta() bool {
	h, err := s.currentHand()
	return err == nil && setteEMezzoIndexOfMatta(h.cards) >= 0
}

// appendLog 棋譜エントリを追加
func (s *SetteEMezzo) appendLog(actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: len(s.actionLog),
		PlayerIdx:  s.activeSeat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      append([]*Card(nil), cards...),
	})
}

// setteEMezzoHandJSON is the wire format for one hand. SetteEMezzoHand's fields
// are unexported, so marshalling it directly would emit `{}` and every hand
// would come back from KV empty.
type setteEMezzoHandJSON struct {
	Cards       []*Card `json:"cd"`
	Bet         int     `json:"bt"`
	MattaHalves int     `json:"mh"`
	Stood       bool    `json:"sd"`
	Payout      int     `json:"po"`
}

// MarshalJSON implements json.Marshaler for SetteEMezzoHand.
func (h *SetteEMezzoHand) MarshalJSON() ([]byte, error) {
	return json.Marshal(setteEMezzoHandJSON{
		Cards:       h.cards,
		Bet:         h.bet,
		MattaHalves: h.mattaHalves,
		Stood:       h.stood,
		Payout:      h.payout,
	})
}

// UnmarshalJSON implements json.Unmarshaler for SetteEMezzoHand.
func (h *SetteEMezzoHand) UnmarshalJSON(data []byte) error {
	var j setteEMezzoHandJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Cards) > setteEMezzoMaxSliceLen {
		return errors.New("settemezzo: hand exceeds maximum allowed size")
	}
	if j.MattaHalves < 0 || j.MattaHalves > 14 {
		return fmt.Errorf("settemezzo: invalid matta value: %d", j.MattaHalves)
	}
	h.cards = j.Cards
	h.bet = j.Bet
	h.mattaHalves = j.MattaHalves
	h.stood = j.Stood
	h.payout = j.Payout
	return nil
}

// setteEMezzoSeatJSON is the wire format for one seat.
type setteEMezzoSeatJSON struct {
	Name  string           `json:"nm"`
	IsCPU bool             `json:"cp"`
	Hand  *SetteEMezzoHand `json:"hd,omitempty"`
}

// MarshalJSON implements json.Marshaler for SetteEMezzoSeat.
func (s *SetteEMezzoSeat) MarshalJSON() ([]byte, error) {
	return json.Marshal(setteEMezzoSeatJSON{Name: s.name, IsCPU: s.isCPU, Hand: s.hand})
}

// UnmarshalJSON implements json.Unmarshaler for SetteEMezzoSeat.
func (s *SetteEMezzoSeat) UnmarshalJSON(data []byte) error {
	var j setteEMezzoSeatJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	s.name = j.Name
	s.isCPU = j.IsCPU
	s.hand = j.Hand
	return nil
}

// setteEMezzoJSON is the JSON wire format for SetteEMezzo.
type setteEMezzoJSON struct {
	TrumpCards *TrumpCards        `json:"tc"`
	Seats      []*SetteEMezzoSeat `json:"st"`
	Banker     int                `json:"bk"`
	BankerHand *SetteEMezzoHand   `json:"bh"`
	Chips      int                `json:"ch"`
	ActiveSeat int                `json:"as"`
	Phase      int                `json:"ph"`
	NextBanker int                `json:"nb"`
	LastResult string             `json:"lr"`
	ActionLog  []*ActionLogEntry  `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (s *SetteEMezzo) MarshalJSON() ([]byte, error) {
	return json.Marshal(setteEMezzoJSON{
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
func (s *SetteEMezzo) UnmarshalJSON(data []byte) error {
	var j setteEMezzoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < SetteEMezzoPhaseBet || j.Phase > SetteEMezzoPhaseEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if len(j.Seats) > SetteEMezzoSeatCnt {
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
	if len(j.ActionLog) > setteEMezzoMaxSliceLen {
		return errors.New("settemezzo: action log exceeds maximum allowed size")
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
