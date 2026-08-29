//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ポンツーンのフェーズ定数
const (
	// PontoonPhaseBet ベットフェーズ
	PontoonPhaseBet = 1
	// PontoonPhasePlayerTurn 人間の手番（Stick / Twist / Buy / Split）
	PontoonPhasePlayerTurn = 2
	// PontoonPhaseBankerTurn 人間が親のときの引き止め判断
	PontoonPhaseBankerTurn = 3
	// PontoonPhaseEnd 精算済み
	PontoonPhaseEnd = 4
)

// ポンツーンの既定値
const (
	// PontoonDefaultChips 初期チップ
	PontoonDefaultChips = 1000
	// PontoonMinBet 最低ベット額
	PontoonMinBet = 10
	// PontoonMaxBet 最大ベット額
	PontoonMaxBet = 10000
	// PontoonSeatCnt 座席数（人間 1 + CPU 2）
	PontoonSeatCnt = 3
	// PontoonMaxCards 1 つの手に持てる上限。5 枚に達したら引けない。
	PontoonMaxCards = 5
	// PontoonStickMin これ未満では Stick できない。
	PontoonStickMin = 15
	// PontoonCpuStickMin CPU 席と CPU 親がここに達したら止める。
	//
	// **相手の停止ラインは人間の判断材料。**引き続けるかどうかは「相手がどこで
	// 止まるか」で決まるのに、この数字はどこにも出ていなかった (#5565)。案内は
	// この定数から書く。文言に 17 を焼き込むと、閾値を変えたとき案内だけが嘘になる。
	PontoonCpuStickMin = 17
	// PontoonTarget 21
	PontoonTarget = 21
	// PontoonMaxHands スプリットで作れる手の上限
	PontoonMaxHands = 4
	// pontoonMaxSliceLen JSON 復元時のスライス長上限
	pontoonMaxSliceLen = 1000
)

// PontoonRank 手の格。数値が大きいほど強い。
type PontoonRank int

// PontoonRank の定数。ポンツーン > ファイブカード・トリック > それ以外の点数。
const (
	// PontoonRankBust バースト（22 以上）。無価値。
	PontoonRankBust PontoonRank = iota
	// PontoonRankPoints 21 以下の通常の手。点数で比べる。
	PontoonRankPoints
	// PontoonRankFiveCard ちょうど 5 枚で 21 以下。
	PontoonRankFiveCard
	// PontoonRankPontoon A + 10 点札の 2 枚。最強。
	PontoonRankPontoon
)

// PontoonHand 1 つの手。スプリットすると 1 人が複数持つ。
type PontoonHand struct {
	cards []*Card
	bet   int
	// twisted は一度でも Twist したか。Buy は Twist 後には打てない。
	twisted bool
	stuck   bool
	// payout 精算後の増減（チップの差分）。
	payout int
}

// GetCards 手札を取得する
func (h *PontoonHand) GetCards() []*Card { return h.cards }

// GetBet この手に乗っている賭け金を取得する
func (h *PontoonHand) GetBet() int { return h.bet }

// IsTwisted 一度でも Twist したか
func (h *PontoonHand) IsTwisted() bool { return h.twisted }

// IsStuck Stick 済みか
func (h *PontoonHand) IsStuck() bool { return h.stuck }

// GetPayout 精算後の増減
func (h *PontoonHand) GetPayout() int { return h.payout }

// PontoonSeat 1 つの座席。人間は seat 0。
type PontoonSeat struct {
	name  string
	isCPU bool
	hands []*PontoonHand
}

// GetName 座席名
func (s *PontoonSeat) GetName() string { return s.name }

// IsCPU CPU 席か
func (s *PontoonSeat) IsCPU() bool { return s.isCPU }

// GetHands その席の手をすべて取得する
func (s *PontoonSeat) GetHands() []*PontoonHand { return s.hands }

// Pontoon ポンツーン（英国式ブラックジャック）ゲーム本体。
//
// 52 枚 1 組。**親（バンカー）を含む全員が 2 枚とも裏向き**で受け取るのが
// ブラックジャックとの一番の違いで、親のアップカードが無いぶん読み合いの質が変わる。
//
// 手の格は点数だけでは決まらない:
//
//	ポンツーン（A + 10 点札の 2 枚） > ファイブカード・トリック（ちょうど 5 枚で 21 以下）
//	> 21 以下の点数勝負
//
// 手番の選択肢は 3 つで、それぞれに制約がある:
//
//   - **Stick**: 合計 15 以上でなければ宣言できない。14 以下は引くしかない
//   - **Twist**: 表向きに 1 枚引く。賭け金は増えない
//   - **Buy**: 裏向きに 1 枚引く。賭け金を上乗せする。**一度 Twist したら以後 Buy は打てない**
//
// **同点は親の勝ち。** 親がファイブカード・トリックを作ると、ポンツーン以外の手は
// 点数に関係なく負ける。
//
// 親は固定ではない。**スプリットしていない手でポンツーンを出したプレイヤーは、
// 親がポンツーンでなければ次の局から親になる。** 人間が親になった局は、全員の手番の
// あとに人間自身が引き止めを決める（PontoonPhaseBankerTurn）。
//
// issue #4379 の仕様案とは 5 点異なり、いずれも実際の規則に合わせた:
//   - ファイブカード・トリックは「5 枚以上」ではなく**ちょうど 5 枚**。5 枚に達した
//     時点で引けなくなるので、6 枚の手は存在しない
//   - 親を継ぐのは「ディーラーに勝った中で最高役の者」ではなく、**ポンツーンを出した者**
//   - **15 未満では Stick できない**（issue は触れていない）
//   - **Twist のあとは Buy できない**、上限 5 枚（issue は触れていない）
//   - **同点は親の勝ち**、親のファイブカード・トリックはポンツーン以外に勝つ
//     （issue は「ディーラー未満の手は敗北」としか書いていない）
type Pontoon struct {
	trumpCards *TrumpCards
	seats      []*PontoonSeat
	// banker はバンカー席のインデックス。人間が親のこともある。
	banker int
	// bankerHand は親の手。親の席の hands とは別に持ち、親は賭けない。
	bankerHand *PontoonHand
	chips      ChipHolder
	// activeSeat / activeHand は手番の位置。親の席は飛ばす。
	activeSeat int
	activeHand int
	phase      int
	// nextBanker は次局の親。この局でポンツーンを出した最初のプレイヤー。
	nextBanker int
	lastResult string
	actionLog  []*ActionLogEntry
}

// pontoonOpeningBanker 最初の局の親。
//
// 人間 (seat 0) にしないのは、そこから始めると新規セッションがいきなり
// 「あなたが配ってください」で開き、このゲームの主要な流れ — 賭けて Stick /
// Twist / Buy を選ぶ — に入れないため。最初に誰が親を持つかは規則が定めて
// いないので (以後はポンツーンを出した者が取る)、遊びやすい方を選ぶ。
const pontoonOpeningBanker = 1

// NewPontoon コンストラクタ
func NewPontoon(trumpCards *TrumpCards) *Pontoon {
	p := &Pontoon{
		trumpCards: trumpCards,
		phase:      PontoonPhaseBet,
		banker:     pontoonOpeningBanker,
		// nextBanker must start at -1, not at its zero value. Reset applies it
		// with `>= 0`, so a zero would hand seat 0 -- the human -- the bank on
		// the very first deal and undo the line above.
		nextBanker: -1,
	}
	p.chips.SetChips(PontoonDefaultChips)
	return p
}

// NewDefaultPontoon returns Pontoon with a standard 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultPontoon() *Pontoon {
	return NewPontoon(NewTrumpCards(0))
}

// Reset 新しい局を始める。親は前局から引き継ぐ。
func (p *Pontoon) Reset() {
	if p.chips.GetChips() < PontoonMinBet {
		p.chips.SetChips(PontoonDefaultChips)
	}
	if p.seats == nil {
		p.seats = make([]*PontoonSeat, PontoonSeatCnt)
		for i := range PontoonSeatCnt {
			p.seats[i] = &PontoonSeat{name: pontoonSeatName(i), isCPU: i != 0}
		}
	}
	// 前局でポンツーンを出した者がいれば親を引き継ぐ。
	if p.nextBanker >= 0 && p.nextBanker < len(p.seats) {
		p.banker = p.nextBanker
	}
	p.nextBanker = -1

	p.trumpCards.Shuffle()
	for _, s := range p.seats {
		s.hands = nil
	}
	p.bankerHand = nil
	p.activeSeat = 0
	p.activeHand = 0
	p.phase = PontoonPhaseBet
	p.lastResult = ""
	p.actionLog = nil
}

// pontoonSeatName 席の既定名
func pontoonSeatName(i int) string {
	if i == 0 {
		return "あなた"
	}
	return fmt.Sprintf("CPU%d", i)
}

// PlaceBet 人間のベットを置き、その局を配る。
func (p *Pontoon) PlaceBet(bet int) error {
	if p.phase != PontoonPhaseBet {
		return NewDomainErrorCode(ErrWrongPhase, "pontoon.errNotBettingPhase", nil)
	}
	if p.IsHumanBanker() {
		// 親は賭けない。全員の賭けを受ける側なので、ベットを求めるのは誤り。
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errBankerDoesNotBet", nil)
	}
	if bet < PontoonMinBet || bet > PontoonMaxBet {
		return NewDomainErrorCode(ErrInvalidAmount, "pontoon.errBetOutOfRange", map[string]string{"Min": strconv.Itoa(PontoonMinBet), "Max": strconv.Itoa(PontoonMaxBet)})
	}
	if bet > p.chips.GetChips() {
		return NewDomainErrorCode(ErrInsufficientChips, "pontoon.errNotEnoughChips", nil)
	}
	p.chips.SetChips(p.chips.GetChips() - bet)
	p.deal(bet)
	return nil
}

// StartAsBanker 人間が親の局を配る。人間は賭けないのでベット額を取らない。
func (p *Pontoon) StartAsBanker() error {
	if p.phase != PontoonPhaseBet {
		return NewDomainErrorCode(ErrWrongPhase, "pontoon.errNotBettingPhase", nil)
	}
	if !p.IsHumanBanker() {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errNotBanker", nil)
	}
	p.deal(0)
	return nil
}

// deal 全員に裏向き 2 枚を配り、CPU の手番を進める。
func (p *Pontoon) deal(humanBet int) {
	for i, s := range p.seats {
		if i == p.banker {
			continue
		}
		bet := humanBet
		if s.isCPU {
			bet = pontoonCpuBet(p.chips.GetChips())
		}
		s.hands = []*PontoonHand{{cards: p.drawTwo(), bet: bet}}
	}
	p.bankerHand = &PontoonHand{cards: p.drawTwo()}
	p.appendLog("deal", "全員に裏向き2枚を配った", nil)

	// 親のポンツーンは即座に開かれ、その局は終わる。
	if pontoonRankOf(p.bankerHand.cards) == PontoonRankPontoon {
		p.appendLog("pontoon", "親がポンツーン", p.bankerHand.cards)
		p.settle()
		return
	}

	p.phase = PontoonPhasePlayerTurn
	p.activeSeat = 0
	p.activeHand = 0
	p.advanceToHuman()
}

// drawTwo 2 枚引く
func (p *Pontoon) drawTwo() []*Card {
	out := make([]*Card, 0, 2)
	for range 2 {
		if c := p.trumpCards.DrawCard(); c != nil {
			out = append(out, c)
		}
	}
	return out
}

// advanceToHuman 人間の番が来るまで CPU を自動で進める。
// 人間が親なら全員ぶん進めて親の手番へ移る。
func (p *Pontoon) advanceToHuman() {
	for p.phase == PontoonPhasePlayerTurn {
		if p.activeSeat >= len(p.seats) {
			p.startBankerTurn()
			return
		}
		s := p.seats[p.activeSeat]
		if p.activeSeat == p.banker || len(s.hands) == 0 {
			p.activeSeat++
			p.activeHand = 0
			continue
		}
		if !s.isCPU {
			// 人間の手が全部片付いていれば次の席へ。
			if p.activeHand >= len(s.hands) {
				p.activeSeat++
				p.activeHand = 0
				continue
			}
			return
		}
		p.playCpuSeat(s)
		p.activeSeat++
		p.activeHand = 0
	}
}

// playCpuSeat CPU 席を最後まで自動で打つ。
func (p *Pontoon) playCpuSeat(s *PontoonSeat) {
	for _, h := range s.hands {
		for {
			total := pontoonTotal(h.cards)
			if len(h.cards) >= PontoonMaxCards || total > PontoonTarget {
				break
			}
			// 15 未満は宣言できないので必ず引く。以降は 17 を目安に止める。
			if total >= PontoonStickMin && total >= PontoonCpuStickMin {
				h.stuck = true
				break
			}
			if total >= PontoonStickMin && len(h.cards) >= 4 {
				// 4 枚で 15〜16 なら、あと 1 枚でファイブカード・トリックが狙える。
				h.twisted = true
				p.hit(h)
				continue
			}
			if total >= PontoonStickMin {
				h.stuck = true
				break
			}
			h.twisted = true
			p.hit(h)
		}
	}
}

// hit 1 枚引いて手に加える
func (p *Pontoon) hit(h *PontoonHand) {
	if c := p.trumpCards.DrawCard(); c != nil {
		h.cards = append(h.cards, c)
	}
}

// pontoonCpuBet CPU の賭け金。人間の残高には影響しないので固定幅で十分。
func pontoonCpuBet(int) int { return PontoonMinBet * 2 }

// Stick 今の手を打ち止めにする。15 未満では宣言できない。
func (p *Pontoon) Stick() error {
	h, err := p.currentHand()
	if err != nil {
		return err
	}
	if pontoonTotal(h.cards) < PontoonStickMin {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errCannotStickBelowMin", map[string]string{"Min": strconv.Itoa(PontoonStickMin)})
	}
	h.stuck = true
	p.appendLog("stick", "スティック", h.cards)
	p.nextHand()
	return nil
}

// Twist 表向きに 1 枚引く。賭け金は増えない。
func (p *Pontoon) Twist() error {
	h, err := p.currentHand()
	if err != nil {
		return err
	}
	if len(h.cards) >= PontoonMaxCards {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errFiveCardTrickFull", nil)
	}
	if pontoonTotal(h.cards) >= PontoonTarget {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errCannotTwistOn21OrMore", nil)
	}
	h.twisted = true
	p.hit(h)
	p.appendLog("twist", "ツイスト", h.cards)
	if pontoonTotal(h.cards) > PontoonTarget || len(h.cards) >= PontoonMaxCards {
		p.nextHand()
	}
	return nil
}

// Buy 賭け金を上乗せして裏向きに 1 枚引く。
//
// 一度 Twist した手では打てない。これは「安く引いてから高く買い直す」ことを
// 封じる規則で、Twist を切るタイミングがこのゲームの読みどころになる。
func (p *Pontoon) Buy(extra int) error {
	h, err := p.currentHand()
	if err != nil {
		return err
	}
	if h.twisted {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errCannotBuyAfterTwisting", nil)
	}
	if len(h.cards) >= PontoonMaxCards {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errFiveCardTrickFull", nil)
	}
	if pontoonTotal(h.cards) >= PontoonTarget {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errCannotBuyOn21OrMore", nil)
	}
	if extra < PontoonMinBet || extra > h.bet*2 {
		return NewDomainErrorCode(ErrInvalidAmount, "pontoon.errBuyStakeOutOfRange", map[string]string{"Min": strconv.Itoa(PontoonMinBet)})
	}
	if extra > p.chips.GetChips() {
		return NewDomainErrorCode(ErrInsufficientChips, "pontoon.errNotEnoughChips", nil)
	}
	p.chips.SetChips(p.chips.GetChips() - extra)
	h.bet += extra
	p.hit(h)
	p.appendLog("buy", fmt.Sprintf("バイ（+%d）", extra), h.cards)
	if pontoonTotal(h.cards) > PontoonTarget || len(h.cards) >= PontoonMaxCards {
		p.nextHand()
	}
	return nil
}

// Split 同ランク 2 枚を 2 つの手に分ける。
//
// 10 点札同士でも**ランクが同じ**でなければ割れない。Q と J は両方 10 点だが別ランク。
func (p *Pontoon) Split() error {
	h, err := p.currentHand()
	if err != nil {
		return err
	}
	s := p.seats[p.activeSeat]
	if len(s.hands) >= PontoonMaxHands {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errTooManyHands", map[string]string{"Max": strconv.Itoa(PontoonMaxHands)})
	}
	if len(h.cards) != 2 || h.cards[0].GetValue() != h.cards[1].GetValue() {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errSplitRequiresEqualRank", nil)
	}
	if h.bet > p.chips.GetChips() {
		return NewDomainErrorCode(ErrInsufficientChips, "pontoon.errNotEnoughChips", nil)
	}
	p.chips.SetChips(p.chips.GetChips() - h.bet)
	moved := h.cards[1]
	h.cards = h.cards[:1]
	p.hit(h)
	newHand := &PontoonHand{cards: []*Card{moved}, bet: h.bet}
	p.hit(newHand)
	// 分けた手は元の手のすぐ後ろに挿す。順に打てば席内の並びと一致する。
	s.hands = append(s.hands, nil)
	copy(s.hands[p.activeHand+2:], s.hands[p.activeHand+1:])
	s.hands[p.activeHand+1] = newHand
	p.appendLog("split", "スプリット", h.cards)
	return nil
}

// currentHand 手番の手を返す
func (p *Pontoon) currentHand() (*PontoonHand, error) {
	if p.phase != PontoonPhasePlayerTurn {
		return nil, NewDomainErrorCode(ErrWrongPhase, "pontoon.errNotPlayerTurn", nil)
	}
	if p.activeSeat != 0 || p.activeSeat == p.banker {
		return nil, NewDomainErrorCode(ErrInvalidPlay, "pontoon.errNotHumanTurn", nil)
	}
	s := p.seats[0]
	if p.activeHand < 0 || p.activeHand >= len(s.hands) {
		return nil, NewDomainErrorCode(ErrInvalidPlay, "pontoon.errNoHandInPlay", nil)
	}
	return s.hands[p.activeHand], nil
}

// nextHand 次の手へ進める。席の手が尽きたら次の席へ。
func (p *Pontoon) nextHand() {
	p.activeHand++
	if p.activeHand >= len(p.seats[p.activeSeat].hands) {
		p.activeSeat++
		p.activeHand = 0
	}
	p.advanceToHuman()
}

// startBankerTurn 親の手番に入る。CPU が親なら自動で打って精算まで進む。
func (p *Pontoon) startBankerTurn() {
	if p.IsHumanBanker() {
		p.phase = PontoonPhaseBankerTurn
		return
	}
	for {
		total := pontoonTotal(p.bankerHand.cards)
		if len(p.bankerHand.cards) >= PontoonMaxCards || total > PontoonTarget || total >= PontoonCpuStickMin {
			break
		}
		p.hit(p.bankerHand)
	}
	p.settle()
}

// BankerTwist 人間が親のときに 1 枚引く。
func (p *Pontoon) BankerTwist() error {
	if p.phase != PontoonPhaseBankerTurn {
		return NewDomainErrorCode(ErrWrongPhase, "pontoon.errNotBankerTurn", nil)
	}
	if len(p.bankerHand.cards) >= PontoonMaxCards {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errFiveCardTrickFull", nil)
	}
	if pontoonTotal(p.bankerHand.cards) > PontoonTarget {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errBankerAlreadyBust", nil)
	}
	p.hit(p.bankerHand)
	p.appendLog("bankerTwist", "親がカードを引いた", p.bankerHand.cards)
	if pontoonTotal(p.bankerHand.cards) > PontoonTarget || len(p.bankerHand.cards) >= PontoonMaxCards {
		p.settle()
	}
	return nil
}

// BankerStay 人間が親のときに引くのをやめて精算する。
func (p *Pontoon) BankerStay() error {
	if p.phase != PontoonPhaseBankerTurn {
		return NewDomainErrorCode(ErrWrongPhase, "pontoon.errNotBankerTurn", nil)
	}
	p.settle()
	return nil
}

// settle 全席を精算する。
func (p *Pontoon) settle() {
	bankerRank, bankerTotal := pontoonRankOf(p.bankerHand.cards), pontoonTotal(p.bankerHand.cards)
	for i, s := range p.seats {
		if i == p.banker {
			continue
		}
		for _, h := range s.hands {
			h.payout = p.settleHand(h, bankerRank, bankerTotal)
			// 親が人間なら、プレイヤーの損得はそのまま親の得損になる。
			if p.IsHumanBanker() {
				p.chips.SetChips(p.chips.GetChips() - h.payout)
			} else if i == 0 {
				p.chips.SetChips(p.chips.GetChips() + h.bet + h.payout)
			}
			// スプリットしていない手のポンツーンだけが親の座を取れる。
			if i != p.banker && len(s.hands) == 1 &&
				pontoonRankOf(h.cards) == PontoonRankPontoon &&
				bankerRank != PontoonRankPontoon && p.nextBanker < 0 {
				p.nextBanker = i
			}
		}
	}
	p.phase = PontoonPhaseEnd
	p.lastResult = p.describeResult(bankerRank, bankerTotal)
	p.appendLog("result", p.lastResult, p.bankerHand.cards)
}

// settleHand 1 つの手の増減を返す（賭け金を除いた純増減）。
//
// 支払いは 2 倍と等倍の 2 段階しかない。ポンツーンとファイブカード・トリックが 2 倍で、
// 親がファイブカード・トリックを作った局は**ポンツーン以外が 2 倍で負ける**。
func (p *Pontoon) settleHand(h *PontoonHand, bankerRank PontoonRank, bankerTotal int) int {
	rank := pontoonRankOf(h.cards)
	if rank == PontoonRankBust {
		return -h.bet
	}
	if bankerRank == PontoonRankPontoon {
		return -h.bet * 2
	}
	if rank == PontoonRankPontoon {
		return h.bet * 2
	}
	if bankerRank == PontoonRankFiveCard {
		return -h.bet * 2
	}
	if rank == PontoonRankFiveCard {
		return h.bet * 2
	}
	if bankerRank == PontoonRankBust {
		return h.bet
	}
	// 同点は親の勝ち。
	if pontoonTotal(h.cards) > bankerTotal {
		return h.bet
	}
	return -h.bet
}

// describeResult 精算の要約
func (p *Pontoon) describeResult(bankerRank PontoonRank, bankerTotal int) string {
	switch bankerRank {
	case PontoonRankPontoon:
		return "親がポンツーン"
	case PontoonRankFiveCard:
		return "親がファイブカード・トリック"
	case PontoonRankBust:
		return fmt.Sprintf("親がバースト（%d）", bankerTotal)
	default:
		return fmt.Sprintf("親は %d", bankerTotal)
	}
}

// pontoonTotal 手の合計。A は 11 として数え、バーストするぶんだけ 1 に落とす。
func pontoonTotal(cards []*Card) int {
	total, aces := 0, 0
	for _, c := range cards {
		if c == nil {
			continue
		}
		v := c.GetValue()
		switch {
		case v == 1:
			aces++
			total += 11
		case v >= 10:
			total += 10
		default:
			total += v
		}
	}
	for total > PontoonTarget && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

// pontoonRankOf 手の格を判定する。
func pontoonRankOf(cards []*Card) PontoonRank {
	total := pontoonTotal(cards)
	if total > PontoonTarget {
		return PontoonRankBust
	}
	if len(cards) == 2 && total == PontoonTarget {
		return PontoonRankPontoon
	}
	// ちょうど 5 枚。5 枚に達すると引けなくなるので 6 枚以上は存在しない。
	if len(cards) == PontoonMaxCards {
		return PontoonRankFiveCard
	}
	return PontoonRankPoints
}

// GetPhase フェーズ取得
func (p *Pontoon) GetPhase() int { return p.phase }

// GetChips 人間のチップ
func (p *Pontoon) GetChips() int { return p.chips.GetChips() }

// GetSeats 全席を取得する
func (p *Pontoon) GetSeats() []*PontoonSeat { return p.seats }

// GetBankerIdx 親の席番号
func (p *Pontoon) GetBankerIdx() int { return p.banker }

// IsHumanBanker 人間が親か
func (p *Pontoon) IsHumanBanker() bool { return p.banker == 0 }

// GetBankerHand 親の手
func (p *Pontoon) GetBankerHand() *PontoonHand { return p.bankerHand }

// GetActiveSeat 手番の席
func (p *Pontoon) GetActiveSeat() int { return p.activeSeat }

// GetActiveHand 手番の手
func (p *Pontoon) GetActiveHand() int { return p.activeHand }

// GetNextBanker 次局の親（未定なら -1）
func (p *Pontoon) GetNextBanker() int { return p.nextBanker }

// GetLastResult 直近の精算の要約
func (p *Pontoon) GetLastResult() string { return p.lastResult }

// GetActionLog 棋譜取得
func (p *Pontoon) GetActionLog() []*ActionLogEntry { return p.actionLog }

// GetGameEndFlag 局が終わっているか
func (p *Pontoon) GetGameEndFlag() bool { return p.phase == PontoonPhaseEnd }

// GetHandTotal 手の合計を返す（表示用）
func (p *Pontoon) GetHandTotal(cards []*Card) int { return pontoonTotal(cards) }

// GetHandRank 手の格を返す（表示用）
func (p *Pontoon) GetHandRank(cards []*Card) PontoonRank { return pontoonRankOf(cards) }

// CanStick 今の手で Stick を宣言できるか
func (p *Pontoon) CanStick() bool {
	h, err := p.currentHand()
	return err == nil && pontoonTotal(h.cards) >= PontoonStickMin
}

// CanTwist 今の手で Twist できるか
func (p *Pontoon) CanTwist() bool {
	h, err := p.currentHand()
	return err == nil && len(h.cards) < PontoonMaxCards && pontoonTotal(h.cards) < PontoonTarget
}

// CanBuy 今の手で Buy できるか
func (p *Pontoon) CanBuy() bool {
	h, err := p.currentHand()
	return err == nil && !h.twisted && len(h.cards) < PontoonMaxCards &&
		pontoonTotal(h.cards) < PontoonTarget && p.chips.GetChips() >= PontoonMinBet
}

// CanSplit 今の手を割れるか
func (p *Pontoon) CanSplit() bool {
	h, err := p.currentHand()
	if err != nil {
		return false
	}
	return len(p.seats[0].hands) < PontoonMaxHands && len(h.cards) == 2 &&
		h.cards[0].GetValue() == h.cards[1].GetValue() && p.chips.GetChips() >= h.bet
}

// appendLog 棋譜エントリを追加
func (p *Pontoon) appendLog(actionType, detail string, cards []*Card) {
	p.actionLog = append(p.actionLog, &ActionLogEntry{
		TurnNumber: len(p.actionLog),
		PlayerIdx:  p.activeSeat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      append([]*Card(nil), cards...),
	})
}

// pontoonHandJSON is the wire format for one hand. PontoonHand's fields are
// unexported, so marshalling it directly would emit `{}` and every hand would
// come back from KV empty.
type pontoonHandJSON struct {
	Cards   []*Card `json:"cd"`
	Bet     int     `json:"bt"`
	Twisted bool    `json:"tw"`
	Stuck   bool    `json:"sk"`
	Payout  int     `json:"po"`
}

// MarshalJSON implements json.Marshaler for PontoonHand.
func (h *PontoonHand) MarshalJSON() ([]byte, error) {
	return json.Marshal(pontoonHandJSON{
		Cards:   h.cards,
		Bet:     h.bet,
		Twisted: h.twisted,
		Stuck:   h.stuck,
		Payout:  h.payout,
	})
}

// UnmarshalJSON implements json.Unmarshaler for PontoonHand.
func (h *PontoonHand) UnmarshalJSON(data []byte) error {
	var j pontoonHandJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Cards) > pontoonMaxSliceLen {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errHandTooLarge", nil)
	}
	h.cards = j.Cards
	h.bet = j.Bet
	h.twisted = j.Twisted
	h.stuck = j.Stuck
	h.payout = j.Payout
	return nil
}

// pontoonSeatJSON is the wire format for one seat.
type pontoonSeatJSON struct {
	Name  string         `json:"nm"`
	IsCPU bool           `json:"cp"`
	Hands []*PontoonHand `json:"hd"`
}

// MarshalJSON implements json.Marshaler for PontoonSeat.
func (s *PontoonSeat) MarshalJSON() ([]byte, error) {
	return json.Marshal(pontoonSeatJSON{Name: s.name, IsCPU: s.isCPU, Hands: s.hands})
}

// UnmarshalJSON implements json.Unmarshaler for PontoonSeat.
func (s *PontoonSeat) UnmarshalJSON(data []byte) error {
	var j pontoonSeatJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Hands) > PontoonMaxHands {
		return fmt.Errorf("pontoon: a seat holds %d hands", len(j.Hands))
	}
	s.name = j.Name
	s.isCPU = j.IsCPU
	s.hands = j.Hands
	return nil
}

// pontoonJSON is the JSON wire format for Pontoon.
type pontoonJSON struct {
	TrumpCards *TrumpCards       `json:"tc"`
	Seats      []*PontoonSeat    `json:"st"`
	Banker     int               `json:"bk"`
	BankerHand *PontoonHand      `json:"bh"`
	Chips      int               `json:"ch"`
	ActiveSeat int               `json:"as"`
	ActiveHand int               `json:"ah"`
	Phase      int               `json:"ph"`
	NextBanker int               `json:"nb"`
	LastResult string            `json:"lr"`
	ActionLog  []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (p *Pontoon) MarshalJSON() ([]byte, error) {
	return json.Marshal(pontoonJSON{
		TrumpCards: p.trumpCards,
		Seats:      p.seats,
		Banker:     p.banker,
		BankerHand: p.bankerHand,
		Chips:      p.chips.GetChips(),
		ActiveSeat: p.activeSeat,
		ActiveHand: p.activeHand,
		Phase:      p.phase,
		NextBanker: p.nextBanker,
		LastResult: p.lastResult,
		ActionLog:  p.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (p *Pontoon) UnmarshalJSON(data []byte) error {
	var j pontoonJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < PontoonPhaseBet || j.Phase > PontoonPhaseEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if len(j.Seats) > PontoonSeatCnt {
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
	if j.ActiveHand < 0 || j.ActiveHand > PontoonMaxHands {
		return fmt.Errorf("invalid active hand: %d", j.ActiveHand)
	}
	if j.Chips < 0 {
		return NewDomainErrorCode(ErrInvalidAmount, "pontoon.errInvalidChips", map[string]string{"Chips": strconv.Itoa(j.Chips)})
	}
	if len(j.ActionLog) > pontoonMaxSliceLen {
		return NewDomainErrorCode(ErrInvalidPlay, "pontoon.errActionLogTooLarge", nil)
	}
	if j.TrumpCards != nil {
		p.trumpCards = j.TrumpCards
	}
	p.seats = j.Seats
	p.banker = j.Banker
	p.bankerHand = j.BankerHand
	p.chips.SetChips(j.Chips)
	p.activeSeat = j.ActiveSeat
	p.activeHand = j.ActiveHand
	p.phase = j.Phase
	p.nextBanker = j.NextBanker
	p.lastResult = j.LastResult
	p.actionLog = j.ActionLog
	return nil
}
