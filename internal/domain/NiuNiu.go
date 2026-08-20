//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// 闘牛（ニウニウ）のフェーズ定数
const (
	// NiuNiuPhaseBet ベットフェーズ
	NiuNiuPhaseBet = 1
	// NiuNiuPhaseEnd 精算済み
	NiuNiuPhaseEnd = 2
)

// 闘牛の既定値
const (
	// NiuNiuDefaultChips 初期チップ
	NiuNiuDefaultChips = 1000
	// NiuNiuMinBet 最低ベット額
	NiuNiuMinBet = 10
	// NiuNiuMaxBet 最大ベット額
	NiuNiuMaxBet = 10000
	// NiuNiuSeatCnt 座席数（人間 1 + CPU 3、うち 1 席が親）
	NiuNiuSeatCnt = 4
	// NiuNiuHandSize 1 人に配る枚数
	NiuNiuHandSize = 5
	// NiuNiuComboSize 牛を作るのに使う枚数
	NiuNiuComboSize = 3
	// niuNiuMaxSliceLen JSON 復元時のスライス長上限
	niuNiuMaxSliceLen = 1000
)

// NiuNiuMaxMultiplier 最大の配当倍率（牛牛の 3 倍）。
//
// 賭け金の検査に使う。このゲームは**賭けた額より多く取られる**（親が牛牛なら
// 3 倍）ので、賭け金だけを残高と比べても足りない。最悪の負けを賄えない額は
// 賭けさせない。
const NiuNiuMaxMultiplier = 3

// NiuNiuRank 手の格。数値が大きいほど強い。
//
// 無牛 (0) < 牛1 (1) < ... < 牛9 (9) < 牛牛 (10)。残り 2 枚の下 1 桁がそのまま
// 格になり、0 のときだけ最強の牛牛として扱う。
type NiuNiuRank int

// NiuNiuRank の定数
const (
	// NiuNiuRankNone 無牛。3 枚で 10 の倍数を作れない手。最弱。
	NiuNiuRankNone NiuNiuRank = 0
	// NiuNiuRankNiuNiu 牛牛。残り 2 枚も 10 の倍数。最強。
	NiuNiuRankNiuNiu NiuNiuRank = 10
)

// NiuNiuHand 1 人の手
type NiuNiuHand struct {
	cards []*Card
	bet   int
	// comboIdx は牛を作った 3 枚の位置。無牛なら nil。
	comboIdx []int
	rank     NiuNiuRank
	payout   int
}

// GetCards 手札を取得する
func (h *NiuNiuHand) GetCards() []*Card { return h.cards }

// GetBet 賭け金を取得する
func (h *NiuNiuHand) GetBet() int { return h.bet }

// GetComboIdx 牛を作った 3 枚の位置を取得する（無牛なら nil）
func (h *NiuNiuHand) GetComboIdx() []int { return h.comboIdx }

// GetRank 手の格を取得する
func (h *NiuNiuHand) GetRank() NiuNiuRank { return h.rank }

// GetPayout 精算後の増減を取得する
func (h *NiuNiuHand) GetPayout() int { return h.payout }

// NiuNiuSeat 1 つの座席。人間は seat 0。
type NiuNiuSeat struct {
	name  string
	isCPU bool
	hand  *NiuNiuHand
}

// GetName 席名
func (s *NiuNiuSeat) GetName() string { return s.name }

// IsCPU CPU 席か
func (s *NiuNiuSeat) IsCPU() bool { return s.isCPU }

// GetHand その席の手
func (s *NiuNiuSeat) GetHand() *NiuNiuHand { return s.hand }

// NiuNiu 闘牛（ニウニウ）ゲーム本体。
//
// 52 枚 1 組、全員に 5 枚ずつ配って親と 1 対 1 で比べる中国のギャンブルゲーム。
// プレイヤーに選択肢は無く、**配られた 5 枚から最良の役が一意に決まる**。
//
// 役の作り方:
//
//	5 枚のうち **3 枚の合計が 10 の倍数**になる組み合わせを探す（総当たり 10 通り）。
//	見つかれば「牛あり」で、**残り 2 枚の合計の下 1 桁**が格になる。
//	下 1 桁が 0 なら最強の**牛牛**、1〜9 ならそれぞれ牛1〜牛9。
//	どの 3 枚でも作れなければ**無牛**で最弱。
//
// 配当は格で変わる: 牛牛が 3 倍、牛7〜牛9 が 2 倍、それ以外は等倍。**勝った側の
// 手の倍率**が適用されるので、親が牛牛なら子は 3 倍取られる。
//
// 同点は捨てない。**最高ランクの札**（K > Q > J > 10 > … > A）で比べ、それも
// 同じならスート（♠ > ♥ > ♣ > ♦）で決める。10 通りしか格が無いので同点は頻繁に
// 起きる。
//
// 絵札の点数について: #4397 は「J/Q/K = 10 点」としており、pagat 系の資料は
// 「絵札 = 0 点」としている。**この 2 つは同値**で、判定はすべて mod 10 でしか
// 使われないため（10 ≡ 0）、3 枚が 10 の倍数になる組み合わせも残り 2 枚の下 1 桁も
// 完全に一致する。issue の記述で実装し、そのことをここに残しておく。
//
// #4397 が触れていないのは同点処理と配当倍率表で、どちらも実装には不可欠なため
// 上記のとおり定めた。
type NiuNiu struct {
	trumpCards *TrumpCards
	seats      []*NiuNiuSeat
	banker     int
	bankerHand *NiuNiuHand
	chips      ChipHolder
	phase      int
	lastResult string
	actionLog  []*ActionLogEntry
}

// NewNiuNiu コンストラクタ
func NewNiuNiu(trumpCards *TrumpCards) *NiuNiu {
	n := &NiuNiu{trumpCards: trumpCards, phase: NiuNiuPhaseBet, banker: niuNiuBankerSeat}
	n.chips.SetChips(NiuNiuDefaultChips)
	return n
}

// niuNiuBankerSeat 親の席。人間 (seat 0) にしないのは、そこから始めると新規
// セッションが賭ける前に終わってしまうため。闘牛の親は局ごとに移らない。
const niuNiuBankerSeat = 3

// NewDefaultNiuNiu returns NiuNiu with a standard 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultNiuNiu() *NiuNiu {
	return NewNiuNiu(NewTrumpCards(0))
}

// Reset 新しい局を始める。
func (n *NiuNiu) Reset() {
	// 最低額を賭けても最悪の負けを賄えない残高は、そもそも遊べない。最低額×最大
	// 倍率を下回ったら積み増す。ここを NiuNiuMinBet だけで見ていると、25 チップの
	// ような「賭けられるが払えない」残高で止まる。
	if n.chips.GetChips() < NiuNiuMinBet*NiuNiuMaxMultiplier {
		n.chips.SetChips(NiuNiuDefaultChips)
	}
	if n.seats == nil {
		n.seats = make([]*NiuNiuSeat, NiuNiuSeatCnt)
		for i := range NiuNiuSeatCnt {
			n.seats[i] = &NiuNiuSeat{name: niuNiuSeatName(i), isCPU: i != 0}
		}
	}
	n.trumpCards.Shuffle()
	for _, s := range n.seats {
		s.hand = nil
	}
	n.bankerHand = nil
	n.phase = NiuNiuPhaseBet
	n.lastResult = ""
	n.actionLog = nil
}

// niuNiuSeatName 席の既定名
func niuNiuSeatName(i int) string {
	if i == 0 {
		return "あなた"
	}
	if i == niuNiuBankerSeat {
		return "親"
	}
	return fmt.Sprintf("CPU%d", i)
}

// PlaceBet ベットを置き、配って精算まで一気に進める。
//
// プレイヤーに手番の選択肢が無いゲームなので、フェーズはベットと精算の 2 つしか
// 無い。引く／止めるを選ぶ余地は最初から存在しない。
func (n *NiuNiu) PlaceBet(bet int) error {
	if n.phase != NiuNiuPhaseBet {
		return errors.New("niuniu: not in the betting phase")
	}
	if bet < NiuNiuMinBet || bet > NiuNiuMaxBet {
		return fmt.Errorf("niuniu: bet must be between %d and %d", NiuNiuMinBet, NiuNiuMaxBet)
	}
	// 賭け金ではなく**最悪の負け**を残高と比べる。親が牛牛なら 3 倍取られるので、
	// bet だけ見ていると残高が負になる（chips=25 で 10 を賭け、親が牛牛だと -5）。
	if bet*NiuNiuMaxMultiplier > n.chips.GetChips() {
		return fmt.Errorf("niuniu: a stake of %d needs %d chips to cover the worst case",
			bet, bet*NiuNiuMaxMultiplier)
	}
	n.chips.SetChips(n.chips.GetChips() - bet)
	n.deal(bet)
	return nil
}

// deal 全員に 5 枚ずつ配り、役を確定して精算する。
func (n *NiuNiu) deal(humanBet int) {
	for i, s := range n.seats {
		if i == n.banker {
			continue
		}
		bet := humanBet
		if s.isCPU {
			bet = NiuNiuMinBet * 2
		}
		s.hand = n.newHand(bet)
	}
	n.bankerHand = n.newHand(0)
	n.appendLog("deal", "全員に5枚ずつ配った", nil)
	n.settle()
}

// newHand 5 枚配って役を確定した手を作る
func (n *NiuNiu) newHand(bet int) *NiuNiuHand {
	h := &NiuNiuHand{bet: bet}
	for range NiuNiuHandSize {
		if c := n.trumpCards.DrawCard(); c != nil {
			h.cards = append(h.cards, c)
		}
	}
	h.comboIdx, h.rank = niuNiuEvaluate(h.cards)
	return h
}

// niuNiuCardPoints 1 枚の点数。A は 1、2〜10 は額面、絵札は 10。
//
// 絵札を 0 点とする流儀もあるが、判定はすべて mod 10 でしか使わないので
// (10 ≡ 0) どちらでも結果は変わらない。
func niuNiuCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v >= 10 {
		return 10
	}
	return v
}

// niuNiuEvaluate 5 枚から役を判定する。
//
// 3 枚の組み合わせは 10 通りしか無いので総当たりで足りる。複数の組み合わせが
// 10 の倍数になることはあるが、そのとき残り 2 枚の合計は必ず一致する
// (5 枚の合計が一定で、抜いた 3 枚が同じ剰余なら残りも同じ剰余になる) ので、
// どれを選んでも格は変わらない。最初に見つけたものを返す。
func niuNiuEvaluate(cards []*Card) ([]int, NiuNiuRank) {
	if len(cards) != NiuNiuHandSize {
		return nil, NiuNiuRankNone
	}
	pts := make([]int, len(cards))
	for i, c := range cards {
		pts[i] = niuNiuCardPoints(c)
	}
	total := 0
	for _, p := range pts {
		total += p
	}
	for i := range NiuNiuHandSize {
		for j := i + 1; j < NiuNiuHandSize; j++ {
			for k := j + 1; k < NiuNiuHandSize; k++ {
				if (pts[i]+pts[j]+pts[k])%10 != 0 {
					continue
				}
				rest := (total - pts[i] - pts[j] - pts[k]) % 10
				rank := NiuNiuRank(rest)
				if rest == 0 {
					rank = NiuNiuRankNiuNiu
				}
				return []int{i, j, k}, rank
			}
		}
	}
	return nil, NiuNiuRankNone
}

// niuNiuMultiplier 格ごとの配当倍率。牛牛が 3 倍、牛7〜牛9 が 2 倍、あとは等倍。
func niuNiuMultiplier(rank NiuNiuRank) int {
	switch {
	case rank == NiuNiuRankNiuNiu:
		return 3
	case rank >= 7:
		return 2
	default:
		return 1
	}
}

// niuNiuSuitRank スートの強さ。♠ > ♥ > ♣ > ♦。
//
// Card の定数は ♠=1 ♣=2 ♥=3 ♦=4 の順なので、そのままでは使えない。
func niuNiuSuitRank(design int) int {
	switch design {
	case CardDesignSpade:
		return 4
	case CardDesignHeart:
		return 3
	case CardDesignClover:
		return 2
	default:
		return 1
	}
}

// niuNiuCardStrength 同点を割るための 1 枚の強さ。K > Q > J > 10 > … > A。
//
// 点数ではなくランクで比べる。絵札はどれも 10 点だが、ここでは K > Q > J に
// 分かれる。
func niuNiuCardStrength(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v == 1 {
		// A は最弱。ランク 1 のままだと 2 より下にならないので明示的に下げる。
		v = 0
	}
	return v*10 + niuNiuSuitRank(c.GetDesign())
}

// niuNiuHighCard 手の中で最も強い 1 枚を返す
func niuNiuHighCard(cards []*Card) int {
	best := 0
	for _, c := range cards {
		if s := niuNiuCardStrength(c); s > best {
			best = s
		}
	}
	return best
}

// niuNiuBeats a が b に勝つか。格が同じなら最高ランクの札、それも同じならスート。
func niuNiuBeats(a, b *NiuNiuHand) bool {
	if a.rank != b.rank {
		return a.rank > b.rank
	}
	return niuNiuHighCard(a.cards) > niuNiuHighCard(b.cards)
}

// settle 全席を親と比べて精算する。
func (n *NiuNiu) settle() {
	for i, s := range n.seats {
		if i == n.banker || s.hand == nil {
			continue
		}
		h := s.hand
		h.payout = n.settleHand(h)
		if i == 0 {
			n.chips.SetChips(n.chips.GetChips() + h.bet + h.payout)
		}
	}
	n.phase = NiuNiuPhaseEnd
	n.lastResult = fmt.Sprintf("親: %s", NiuNiuRankLabel(n.bankerHand.rank))
	n.appendLog("result", n.lastResult, n.bankerHand.cards)
}

// settleHand 1 つの手の増減（賭け金を除いた純増減）。
//
// 倍率は**勝った側の手**のものを使う。親が牛牛なら子は 3 倍取られ、子が牛牛なら
// 3 倍もらう。
func (n *NiuNiu) settleHand(h *NiuNiuHand) int {
	if niuNiuBeats(h, n.bankerHand) {
		return h.bet * niuNiuMultiplier(h.rank)
	}
	return -h.bet * niuNiuMultiplier(n.bankerHand.rank)
}

// NiuNiuRankLabel 格の表示名
func NiuNiuRankLabel(rank NiuNiuRank) string {
	if rank == NiuNiuRankNone {
		return "無牛"
	}
	if rank == NiuNiuRankNiuNiu {
		return "牛牛"
	}
	return fmt.Sprintf("牛%d", int(rank))
}

// NiuNiuRankKey は格をロケール非依存の識別子で返す。
//
// NiuNiuRankLabel は "牛牛" のような**表示文字列**で、settle() がそれを
// "親: %s" に埋めて presenter がそのまま送っていたため、英語ロケールでも日本語が
// 出ていた (#5567)。文言の組み立ては presenter の i18n に任せ、ドメインは
// どの格かだけを伝える。
func NiuNiuRankKey(rank NiuNiuRank) string {
	switch rank {
	case NiuNiuRankNone:
		return "none"
	case NiuNiuRankNiuNiu:
		return "niuniu"
	default:
		return fmt.Sprintf("n%d", int(rank))
	}
}

// GetBankerRankKey は親の格をロケール非依存の識別子で返す。役が未確定なら空。
func (n *NiuNiu) GetBankerRankKey() string {
	if n.bankerHand == nil {
		return ""
	}
	return NiuNiuRankKey(n.bankerHand.rank)
}

// GetMultiplier 格の配当倍率を返す（表示用）
func (n *NiuNiu) GetMultiplier(rank NiuNiuRank) int { return niuNiuMultiplier(rank) }

// GetPhase フェーズ取得
func (n *NiuNiu) GetPhase() int { return n.phase }

// GetChips 人間のチップ
func (n *NiuNiu) GetChips() int { return n.chips.GetChips() }

// GetMaxMultiplier 最大の配当倍率。賭けられる上限は残高をこれで割った額になる。
func (n *NiuNiu) GetMaxMultiplier() int { return NiuNiuMaxMultiplier }

// GetSeats 全席を取得する
func (n *NiuNiu) GetSeats() []*NiuNiuSeat { return n.seats }

// GetBankerIdx 親の席番号
func (n *NiuNiu) GetBankerIdx() int { return n.banker }

// GetBankerHand 親の手
func (n *NiuNiu) GetBankerHand() *NiuNiuHand { return n.bankerHand }

// GetLastResult 直近の精算の要約
func (n *NiuNiu) GetLastResult() string { return n.lastResult }

// GetActionLog 棋譜取得
func (n *NiuNiu) GetActionLog() []*ActionLogEntry { return n.actionLog }

// GetGameEndFlag 局が終わっているか
func (n *NiuNiu) GetGameEndFlag() bool { return n.phase == NiuNiuPhaseEnd }

// appendLog 棋譜エントリを追加
func (n *NiuNiu) appendLog(actionType, detail string, cards []*Card) {
	n.actionLog = append(n.actionLog, &ActionLogEntry{
		TurnNumber: len(n.actionLog),
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      append([]*Card(nil), cards...),
	})
}

// niuNiuHandJSON is the wire format for one hand. NiuNiuHand's fields are
// unexported, so marshalling it directly would emit `{}` and every hand would
// come back from KV empty.
type niuNiuHandJSON struct {
	Cards    []*Card    `json:"cd"`
	Bet      int        `json:"bt"`
	ComboIdx []int      `json:"ci,omitempty"`
	Rank     NiuNiuRank `json:"rk"`
	Payout   int        `json:"po"`
}

// MarshalJSON implements json.Marshaler for NiuNiuHand.
func (h *NiuNiuHand) MarshalJSON() ([]byte, error) {
	return json.Marshal(niuNiuHandJSON{
		Cards:    h.cards,
		Bet:      h.bet,
		ComboIdx: h.comboIdx,
		Rank:     h.rank,
		Payout:   h.payout,
	})
}

// UnmarshalJSON implements json.Unmarshaler for NiuNiuHand.
func (h *NiuNiuHand) UnmarshalJSON(data []byte) error {
	var j niuNiuHandJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Cards) > niuNiuMaxSliceLen {
		return errors.New("niuniu: hand exceeds maximum allowed size")
	}
	if len(j.ComboIdx) > NiuNiuComboSize {
		return fmt.Errorf("niuniu: combo holds %d cards", len(j.ComboIdx))
	}
	for _, idx := range j.ComboIdx {
		if idx < 0 || idx >= NiuNiuHandSize {
			return fmt.Errorf("niuniu: invalid combo index: %d", idx)
		}
	}
	if j.Rank < NiuNiuRankNone || j.Rank > NiuNiuRankNiuNiu {
		return fmt.Errorf("niuniu: invalid rank: %d", j.Rank)
	}
	h.cards = j.Cards
	h.bet = j.Bet
	h.comboIdx = j.ComboIdx
	h.rank = j.Rank
	h.payout = j.Payout
	return nil
}

// niuNiuSeatJSON is the wire format for one seat.
type niuNiuSeatJSON struct {
	Name  string      `json:"nm"`
	IsCPU bool        `json:"cp"`
	Hand  *NiuNiuHand `json:"hd,omitempty"`
}

// MarshalJSON implements json.Marshaler for NiuNiuSeat.
func (s *NiuNiuSeat) MarshalJSON() ([]byte, error) {
	return json.Marshal(niuNiuSeatJSON{Name: s.name, IsCPU: s.isCPU, Hand: s.hand})
}

// UnmarshalJSON implements json.Unmarshaler for NiuNiuSeat.
func (s *NiuNiuSeat) UnmarshalJSON(data []byte) error {
	var j niuNiuSeatJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	s.name = j.Name
	s.isCPU = j.IsCPU
	s.hand = j.Hand
	return nil
}

// niuNiuJSON is the JSON wire format for NiuNiu.
type niuNiuJSON struct {
	TrumpCards *TrumpCards       `json:"tc"`
	Seats      []*NiuNiuSeat     `json:"st"`
	Banker     int               `json:"bk"`
	BankerHand *NiuNiuHand       `json:"bh"`
	Chips      int               `json:"ch"`
	Phase      int               `json:"ph"`
	LastResult string            `json:"lr"`
	ActionLog  []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (n *NiuNiu) MarshalJSON() ([]byte, error) {
	return json.Marshal(niuNiuJSON{
		TrumpCards: n.trumpCards,
		Seats:      n.seats,
		Banker:     n.banker,
		BankerHand: n.bankerHand,
		Chips:      n.chips.GetChips(),
		Phase:      n.phase,
		LastResult: n.lastResult,
		ActionLog:  n.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (n *NiuNiu) UnmarshalJSON(data []byte) error {
	var j niuNiuJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < NiuNiuPhaseBet || j.Phase > NiuNiuPhaseEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if len(j.Seats) > NiuNiuSeatCnt {
		return fmt.Errorf("invalid seat count: %d", len(j.Seats))
	}
	if j.Banker < 0 || j.Banker >= max(len(j.Seats), 1) {
		return fmt.Errorf("invalid banker: %d", j.Banker)
	}
	if j.Chips < 0 {
		return fmt.Errorf("invalid chips: %d", j.Chips)
	}
	if len(j.ActionLog) > niuNiuMaxSliceLen {
		return errors.New("niuniu: action log exceeds maximum allowed size")
	}
	if j.TrumpCards != nil {
		n.trumpCards = j.TrumpCards
	}
	n.seats = j.Seats
	n.banker = j.Banker
	n.bankerHand = j.BankerHand
	n.chips.SetChips(j.Chips)
	n.phase = j.Phase
	n.lastResult = j.LastResult
	n.actionLog = j.ActionLog
	return nil
}
