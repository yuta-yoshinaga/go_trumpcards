//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
)

// LiteraturePlayerCnt はリテラチャーの人数。
const LiteraturePlayerCnt = 6

// LiteratureTeamCnt はチーム数。
const LiteratureTeamCnt = 2

// LiteratureHandSize は 1 人あたりの配り札。
const LiteratureHandSize = 8

// LiteratureDeckSize は使用する札数 (**8 を除いた 48 枚**)。
const LiteratureDeckSize = 48

// LiteratureHalfSuitCnt はハーフスートの総数。
//
// 4 スート × 2 (低位/高位) = 8 組。
const LiteratureHalfSuitCnt = 8

// LiteratureHalfSuitSize は 1 組の枚数。
const LiteratureHalfSuitSize = 6

// LiteratureWinThreshold は勝利に要るハーフスート数。
//
// **8 組の過半数なので 5 組。**issue の「4 組（半分）」では相手も 4 組になり
// 得るため決着しない。
const LiteratureWinThreshold = LiteratureHalfSuitCnt/2 + 1

// literatureLowRanks は低位ハーフスートの構成 (2-3-4-5-6-7)。
var literatureLowRanks = []int{2, 3, 4, 5, 6, 7}

// literatureHighRanks は高位ハーフスートの構成 (9-10-J-Q-K-A)。
//
// **8 は抜かれている**ので、低位と高位の境目に札が無い。
var literatureHighRanks = []int{9, 10, 11, 12, 13, 1}

// LiteratureHalfSuitOf は札が属するハーフスートの番号を返す (無ければ -1)。
//
// 番号は (スート - 1) * 2 + (高位なら 1)。
func LiteratureHalfSuitOf(c *Card) int {
	if c == nil || c.GetDesign() < CardDesignSpade || c.GetDesign() > CardDesignDiamond {
		return -1
	}
	base := (c.GetDesign() - CardDesignSpade) * 2
	for _, v := range literatureLowRanks {
		if c.GetValue() == v {
			return base
		}
	}
	for _, v := range literatureHighRanks {
		if c.GetValue() == v {
			return base + 1
		}
	}
	// **8 はデッキに無い。**どのハーフスートにも属さない。
	return -1
}

// LiteratureHalfSuitCards は指定ハーフスートの 6 枚を返す。
func LiteratureHalfSuitCards(half int) []*Card {
	if half < 0 || half >= LiteratureHalfSuitCnt {
		return nil
	}
	suit := CardDesignSpade + half/2
	ranks := literatureLowRanks
	if half%2 == 1 {
		ranks = literatureHighRanks
	}
	out := make([]*Card, 0, LiteratureHalfSuitSize)
	for _, v := range ranks {
		out = append(out, NewCard(suit, v, true))
	}
	return out
}

// LiteratureTeamOf は席のチームを返す (範囲外は -1)。
//
// **席は交互。**味方に要求できない規則があるので、席順は飾りではない。
func LiteratureTeamOf(seat int) int {
	if seat < 0 || seat >= LiteraturePlayerCnt {
		return -1
	}
	return seat % LiteratureTeamCnt
}

// LiteratureClaimOutcome は宣言の結果。
type LiteratureClaimOutcome int

// リテラチャーの宣言結果。**3 通りある。**
const (
	// LiteratureClaimWon 宣言側の獲得 (6 枚とも自チームにあり、所在も正しい)
	LiteratureClaimWon LiteratureClaimOutcome = iota
	// LiteratureClaimCancelled 無効。**どちらのチームも取らない。**
	//
	// 6 枚とも自チームにあるが、誰が持っているかの申告を間違えた場合。
	// 相手には渡らない。
	LiteratureClaimCancelled
	// LiteratureClaimLost 相手チームの獲得 (相手が 1 枚でも持っていた)
	LiteratureClaimLost
)

// LiteratureHalfSuitState はハーフスートの帰属。
type LiteratureHalfSuitState int

// ハーフスートの帰属状態。
const (
	// LiteratureHalfOpen まだ宣言されていない
	LiteratureHalfOpen LiteratureHalfSuitState = iota
	// LiteratureHalfTeam0 チーム 0 が獲得
	LiteratureHalfTeam0
	// LiteratureHalfTeam1 チーム 1 が獲得
	LiteratureHalfTeam1
	// LiteratureHalfCancelled 無効。**どちらのものにもならない。**
	LiteratureHalfCancelled
)

// LiteraturePhase はゲームフェーズ。
type LiteraturePhase int

// リテラチャーのフェーズ定数
const (
	// LiteraturePhasePlay 進行中
	LiteraturePhasePlay LiteraturePhase = iota
	// LiteraturePhaseGameEnd ゲーム終了
	LiteraturePhaseGameEnd
)

// LiteratureAsk は 1 件の要求の記録。
//
// **公開情報である。**誰が誰に何を訊いたか、そして通ったかどうかは全員が
// 見ているので、推理の材料になる。
type LiteratureAsk struct {
	From int
	To   int
	Card *Card
	// Success は札を得られたか。
	Success bool
}

// LiteratureClaimResult は 1 件の宣言の記録。
type LiteratureClaimResult struct {
	// Player は宣言した席。
	Player int
	// HalfSuit は宣言したハーフスート。
	HalfSuit int
	// Outcome は結果。
	Outcome LiteratureClaimOutcome
	// AwardedTeam は獲得したチーム (無効なら -1)。
	AwardedTeam int
}

// Literature はリテラチャーのゲームクラス。
//
// インド・カナダで遊ばれる推理型フィッシング。**8 を抜いた 48 枚**を 6 人
// 3 対 3 で 8 枚ずつ持ち、**8 組のハーフスート**を奪い合う。
type Literature struct {
	players []*LiteraturePlayer
	config  LiteratureConfig
	phase   LiteraturePhase
	// currentIdx は現在の手番。
	currentIdx int
	// halfSuits は各ハーフスートの帰属。
	halfSuits [LiteratureHalfSuitCnt]LiteratureHalfSuitState
	// asks は要求の履歴 (公開情報)。
	asks []*LiteratureAsk
	// claims は宣言の履歴。
	claims []*LiteratureClaimResult
	// lastAsk / lastClaim は直前の 1 件 (プレゼンター用)。
	lastAsk     *LiteratureAsk
	lastClaim   *LiteratureClaimResult
	gameEndFlag bool
	winnerTeam  int
	actionLogBase
}

// NewLiterature コンストラクタ
func NewLiterature(players []*LiteraturePlayer, config LiteratureConfig) *Literature {
	return &Literature{players: players, config: config, winnerTeam: -1}
}

// NewDefaultLiterature は人間 1 人 + CPU 5 体の卓を作る。
func NewDefaultLiterature() *Literature {
	players := make([]*LiteraturePlayer, 0, LiteraturePlayerCnt)
	for i := range LiteraturePlayerCnt {
		players = append(players, NewLiteraturePlayer(i == 0))
	}
	return NewLiterature(players, DefaultLiteratureConfig())
}

// Reset はゲームを初期化する。
func (l *Literature) Reset() {
	l.phase = LiteraturePhasePlay
	l.currentIdx = 0
	l.gameEndFlag = false
	l.winnerTeam = -1
	l.asks = make([]*LiteratureAsk, 0)
	l.claims = make([]*LiteratureClaimResult, 0)
	l.lastAsk = nil
	l.lastClaim = nil
	l.actionLog = make([]*ActionLogEntry, 0)
	for i := range LiteratureHalfSuitCnt {
		l.halfSuits[i] = LiteratureHalfOpen
	}
	for i := range LiteraturePlayerCnt {
		if p := l.GetPlayer(i); p != nil {
			p.ResetRound()
		}
	}
	l.dealRound()
	l.addLog(-1, "deal", "", nil)
}

// dealRound は 48 枚を 8 枚ずつ配る。
func (l *Literature) dealRound() {
	deck := newLiteratureDeck()
	literatureShuffle(deck)
	pos := 0
	for range LiteratureHandSize {
		for i := range LiteraturePlayerCnt {
			p := l.GetPlayer(i)
			if pos < len(deck) && p != nil {
				p.AddCard(deck[pos])
				pos++
			}
		}
	}
}

// newLiteratureDeck は **8 を除いた 48 枚**を作る。
func newLiteratureDeck() []*Card {
	cards := make([]*Card, 0, LiteratureDeckSize)
	for half := range LiteratureHalfSuitCnt {
		cards = append(cards, LiteratureHalfSuitCards(half)...)
	}
	return cards
}

// literatureShuffle は札をシャッフルする。
func literatureShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) //nolint:gosec // ゲームのシャッフルに暗号強度は要らない
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// literatureIndexOf は手札の中の札の位置を返す (無ければ -1)。
func literatureIndexOf(p *LiteraturePlayer, c *Card) int {
	if p == nil || c == nil {
		return -1
	}
	for i := range p.GetCardsSize() {
		h := p.GetCard(i)
		if h != nil && h.GetDesign() == c.GetDesign() && h.GetValue() == c.GetValue() {
			return i
		}
	}
	return -1
}

// LiteratureHoldsHalfSuit は席がそのハーフスートの札を 1 枚以上持つかを返す。
func (l *Literature) LiteratureHoldsHalfSuit(seat, half int) bool {
	p := l.GetPlayer(seat)
	if p == nil {
		return false
	}
	for i := range p.GetCardsSize() {
		if LiteratureHalfSuitOf(p.GetCard(i)) == half {
			return true
		}
	}
	return false
}

// LiteratureCanAsk は要求が成立するかを返す。
//
// **条件は 4 つ。**相手チームであること、要求先に手札が残っていること、
// 自分がそのハーフスートを 1 枚以上持っていること、**自分が持っていない札**で
// あること。
func (l *Literature) LiteratureCanAsk(from, to int, c *Card) error {
	if l.gameEndFlag {
		return errors.New("the game is over")
	}
	if from != l.currentIdx {
		return errors.New("it is not your turn")
	}
	fp, tp := l.GetPlayer(from), l.GetPlayer(to)
	if fp == nil || tp == nil {
		return errors.New("there is no such seat")
	}
	// **味方には要求できない。**
	if LiteratureTeamOf(from) == LiteratureTeamOf(to) {
		return errors.New("you may only ask an opponent")
	}
	if tp.GetCardsSize() == 0 {
		return errors.New("that player has no cards left")
	}
	half := LiteratureHalfSuitOf(c)
	if half < 0 {
		return errors.New("that card is not in this pack")
	}
	if l.halfSuits[half] != LiteratureHalfOpen {
		return errors.New("that half-suit has already been settled")
	}
	// **自分がそのハーフスートを持っていなければ訊けない。**
	if !l.LiteratureHoldsHalfSuit(from, half) {
		return errors.New("you must hold a card of that half-suit")
	}
	// **自分が持っている札は訊けない。**
	if literatureIndexOf(fp, c) >= 0 {
		return errors.New("you already hold that card")
	}
	return nil
}

// Ask は札を要求する。
func (l *Literature) Ask(from, to int, c *Card) error {
	if err := l.LiteratureCanAsk(from, to, c); err != nil {
		return err
	}
	fp, tp := l.GetPlayer(from), l.GetPlayer(to)
	idx := literatureIndexOf(tp, c)
	got := idx >= 0
	if got {
		tp.RemoveCard(idx)
		fp.AddCard(NewCard(c.GetDesign(), c.GetValue(), true))
	}
	ask := &LiteratureAsk{From: from, To: to, Card: NewCard(c.GetDesign(), c.GetValue(), true), Success: got}
	l.asks = append(l.asks, ask)
	l.lastAsk = ask
	l.addLog(from, "ask", strconv.Itoa(to)+" "+literatureCardName(c)+" "+strconv.FormatBool(got), []*Card{c})

	// **的中なら手番継続、外れれば手番は要求先へ移る。**
	if !got {
		l.currentIdx = to
	}
	l.settleTurnIfHandEmpty()
	return nil
}

// Claim はハーフスートを宣言する。
//
// holders[i] は LiteratureHalfSuitCards(half)[i] を持っていると申告する席。
func (l *Literature) Claim(player, half int, holders []int) error {
	if l.gameEndFlag {
		return errors.New("the game is over")
	}
	if player != l.currentIdx {
		return errors.New("it is not your turn")
	}
	if half < 0 || half >= LiteratureHalfSuitCnt {
		return errors.New("there is no such half-suit")
	}
	if l.halfSuits[half] != LiteratureHalfOpen {
		return errors.New("that half-suit has already been settled")
	}
	if len(holders) != LiteratureHalfSuitSize {
		return errors.New("a claim must place all six cards")
	}
	team := LiteratureTeamOf(player)
	for _, seat := range holders {
		if LiteratureTeamOf(seat) != team {
			return errors.New("a claim may only place cards with your own team")
		}
	}

	cards := LiteratureHalfSuitCards(half)
	// **相手が 1 枚でも持っていれば相手のものになる。**
	opponentHolds := false
	correct := true
	for i, c := range cards {
		owner := l.literatureOwnerOf(c)
		if owner < 0 {
			continue
		}
		if LiteratureTeamOf(owner) != team {
			opponentHolds = true
			continue
		}
		if owner != holders[i] {
			// **自チーム内で所在を言い間違えた。**相手には渡らない。
			correct = false
		}
	}

	res := &LiteratureClaimResult{Player: player, HalfSuit: half, AwardedTeam: -1}
	switch {
	case opponentHolds:
		res.Outcome = LiteratureClaimLost
		res.AwardedTeam = 1 - team
		l.halfSuits[half] = literatureStateForTeam(1 - team)
	case !correct:
		// **無効。**どちらのチームも取らない。
		res.Outcome = LiteratureClaimCancelled
		l.halfSuits[half] = LiteratureHalfCancelled
	default:
		res.Outcome = LiteratureClaimWon
		res.AwardedTeam = team
		l.halfSuits[half] = literatureStateForTeam(team)
	}
	l.claims = append(l.claims, res)
	l.lastClaim = res
	l.addLog(player, "claim", strconv.Itoa(half)+" "+strconv.Itoa(int(res.Outcome)), nil)

	// 宣言されたハーフスートの札は場から抜ける。
	l.literatureRemoveHalfSuit(half)
	l.settleTurnIfHandEmpty()
	l.checkGameEnd()
	return nil
}

// literatureStateForTeam はチーム番号を帰属状態へ変換する。
func literatureStateForTeam(team int) LiteratureHalfSuitState {
	if team == 0 {
		return LiteratureHalfTeam0
	}
	return LiteratureHalfTeam1
}

// literatureOwnerOf は札を持っている席を返す (誰も持っていなければ -1)。
func (l *Literature) literatureOwnerOf(c *Card) int {
	for i := range LiteraturePlayerCnt {
		if literatureIndexOf(l.GetPlayer(i), c) >= 0 {
			return i
		}
	}
	return -1
}

// literatureRemoveHalfSuit は決着したハーフスートの札を全員の手札から抜く。
func (l *Literature) literatureRemoveHalfSuit(half int) {
	for i := range LiteraturePlayerCnt {
		p := l.GetPlayer(i)
		if p == nil {
			continue
		}
		for j := p.GetCardsSize() - 1; j >= 0; j-- {
			if LiteratureHalfSuitOf(p.GetCard(j)) == half {
				p.RemoveCard(j)
			}
		}
	}
}

// settleTurnIfHandEmpty は手番の席の手札が尽きていたら手番を渡す。
//
// **まだ手札のある味方へ渡す。**味方も全員空なら相手チームへ渡す。
func (l *Literature) settleTurnIfHandEmpty() {
	p := l.GetPlayer(l.currentIdx)
	if p != nil && p.GetCardsSize() > 0 {
		return
	}
	team := LiteratureTeamOf(l.currentIdx)
	for i := 1; i <= LiteraturePlayerCnt; i++ {
		seat := (l.currentIdx + i) % LiteraturePlayerCnt
		q := l.GetPlayer(seat)
		if q != nil && q.GetCardsSize() > 0 && LiteratureTeamOf(seat) == team {
			l.currentIdx = seat
			return
		}
	}
	for i := 1; i <= LiteraturePlayerCnt; i++ {
		seat := (l.currentIdx + i) % LiteraturePlayerCnt
		q := l.GetPlayer(seat)
		if q != nil && q.GetCardsSize() > 0 {
			l.currentIdx = seat
			return
		}
	}
}

// LiteratureTeamHalfSuits はチームが取ったハーフスート数を返す。
func (l *Literature) LiteratureTeamHalfSuits(team int) int {
	want := literatureStateForTeam(team)
	if team < 0 || team >= LiteratureTeamCnt {
		return 0
	}
	n := 0
	for _, st := range l.halfSuits {
		if st == want {
			n++
		}
	}
	return n
}

// LiteratureCancelledCount は無効になったハーフスート数を返す。
//
// **合計が 8 にならないのはこれがあるため。**
func (l *Literature) LiteratureCancelledCount() int {
	n := 0
	for _, st := range l.halfSuits {
		if st == LiteratureHalfCancelled {
			n++
		}
	}
	return n
}

// LiteratureOpenCount はまだ決着していないハーフスート数を返す。
func (l *Literature) LiteratureOpenCount() int {
	n := 0
	for _, st := range l.halfSuits {
		if st == LiteratureHalfOpen {
			n++
		}
	}
	return n
}

// checkGameEnd はゲーム終了かを見る。
//
// **過半数 (5 組) を取れば即決着。**そこまで行かなくても、全組が決着したら
// 多いほうの勝ち。無効があるので合計は 8 に満たないことがある。
func (l *Literature) checkGameEnd() {
	for team := range LiteratureTeamCnt {
		if l.LiteratureTeamHalfSuits(team) >= LiteratureWinThreshold {
			l.finish(team)
			return
		}
	}
	if l.LiteratureOpenCount() > 0 {
		return
	}
	a, b := l.LiteratureTeamHalfSuits(0), l.LiteratureTeamHalfSuits(1)
	switch {
	case a > b:
		l.finish(0)
	case b > a:
		l.finish(1)
	default:
		// 無効が絡んで同数になることがある。勝者なしで終了。
		l.finish(-1)
	}
}

// finish はゲームを終了させる。
func (l *Literature) finish(team int) {
	l.gameEndFlag = true
	l.phase = LiteraturePhaseGameEnd
	l.winnerTeam = team
	l.addLog(-1, "gameEnd", strconv.Itoa(team), nil)
}

// ---- CPU ----

// literatureAskAlreadyFailed は (相手, 札) の組が既に空振り済みかを返す。
//
// **同じ相手に同じ札を訊き直しても必ず外れる。**外れた事実は公開情報なので、
// 使わない手はない。使わないと CPU 同士が同じ札を訊き合って無限に回る。
//
// ただし**その後その札が誰かに渡っていれば無効**になるので、最後の成功より
// あとの空振りだけを見る。
func (l *Literature) literatureAskAlreadyFailed(to int, c *Card) bool {
	if c == nil {
		return false
	}
	failed := false
	for _, a := range l.asks {
		if a.Card == nil || a.Card.GetDesign() != c.GetDesign() || a.Card.GetValue() != c.GetValue() {
			continue
		}
		if a.Success {
			// 札が動いたので、それ以前の「持っていない」は当てにならない。
			failed = false
			continue
		}
		if a.To == to {
			failed = true
		}
	}
	return failed
}

// LiteratureCpuAsk は CPU の要求を決める (訊けなければ nil)。
//
// **相手の手札は見ない。**要求履歴だけから、相手が持っていそうな札を選ぶ。
func (l *Literature) LiteratureCpuAsk(seat int) (int, *Card) {
	p := l.GetPlayer(seat)
	if p == nil {
		return -1, nil
	}
	// 自分が持っているハーフスートを集める。
	holds := map[int]bool{}
	for i := range p.GetCardsSize() {
		if h := LiteratureHalfSuitOf(p.GetCard(i)); h >= 0 && l.halfSuits[h] == LiteratureHalfOpen {
			holds[h] = true
		}
	}
	if len(holds) == 0 {
		return -1, nil
	}

	// **履歴で「持っていそう」と分かった札を優先する。**
	// 相手がそのハーフスートを訊いたということは、その相手も同じ組を集めている。
	for half := range LiteratureHalfSuitCnt {
		if !holds[half] {
			continue
		}
		for _, c := range LiteratureHalfSuitCards(half) {
			if literatureIndexOf(p, c) >= 0 {
				continue
			}
			// **相手がその組を訊いたということは、その組を集めている。**
			// ただし訊いた本人はその札を持っていないことが確定しているので、
			// 「同じ組の別の札」を狙う。空振り済みの組は避ける。
			for _, ask := range l.asks {
				if ask.Card == nil || LiteratureHalfSuitOf(ask.Card) != half {
					continue
				}
				if LiteratureTeamOf(ask.From) == LiteratureTeamOf(seat) {
					continue
				}
				if l.literatureAskAlreadyFailed(ask.From, c) {
					continue
				}
				if l.LiteratureCanAsk(seat, ask.From, c) == nil {
					return ask.From, c
				}
			}
		}
	}

	// 手掛かりが無ければ、まだ試していない (相手, 札) を順に当たる。
	// **空振り済みを飛ばすので、選択肢は必ず減る。**訊く手が尽きれば宣言に回り、
	// 決着していない組が必ず 1 つ減る。
	for half := range LiteratureHalfSuitCnt {
		if !holds[half] {
			continue
		}
		for _, c := range LiteratureHalfSuitCards(half) {
			for i := range LiteraturePlayerCnt {
				if l.literatureAskAlreadyFailed(i, c) {
					continue
				}
				if l.LiteratureCanAsk(seat, i, c) == nil {
					return i, c
				}
			}
		}
	}
	return -1, nil
}

// LiteratureCpuClaim は CPU が宣言できるハーフスートを探す。
//
// **自チームで 6 枚そろっていると確信できるときだけ宣言する。**確信の根拠は
// 自分の手札と、味方が要求して通った履歴のみ。相手の手札は見ない。
func (l *Literature) LiteratureCpuClaim(seat int) (int, []int) {
	p := l.GetPlayer(seat)
	if p == nil {
		return -1, nil
	}
	team := LiteratureTeamOf(seat)
	for half := range LiteratureHalfSuitCnt {
		if l.halfSuits[half] != LiteratureHalfOpen {
			continue
		}
		cards := LiteratureHalfSuitCards(half)
		holders := make([]int, LiteratureHalfSuitSize)
		known := true
		for i, c := range cards {
			if literatureIndexOf(p, c) >= 0 {
				holders[i] = seat
				continue
			}
			// 味方が要求して通った札は、その味方が持っている。
			owner := -1
			for _, ask := range l.asks {
				if ask.Card == nil || !ask.Success {
					continue
				}
				if ask.Card.GetDesign() == c.GetDesign() && ask.Card.GetValue() == c.GetValue() {
					owner = ask.From
				}
			}
			if owner < 0 || LiteratureTeamOf(owner) != team {
				known = false
				break
			}
			holders[i] = owner
		}
		if known {
			return half, holders
		}
	}
	return -1, nil
}

// LiteratureCpuForcedClaim は「訊けず、確信も無い」ときの当て推量の宣言を返す。
//
// **実際のリテラチャーに「パス」は無い。**訊けないなら宣言するしかないので、
// CPU も必ずどちらかを選ぶ。ここが無いと、決着していない組が残ったまま
// 全員が訊けなくなり、ゲームが終わらない。
//
// 所在は分かるぶんだけ埋め、残りは自分に置く。外れれば無効か相手の獲得になる
// が、それがこの規則の帰結である。
func (l *Literature) LiteratureCpuForcedClaim(seat int) (int, []int) {
	p := l.GetPlayer(seat)
	if p == nil {
		return -1, nil
	}
	team := LiteratureTeamOf(seat)

	// 自分が 1 枚でも持っている組を優先する。
	pick := -1
	for half := range LiteratureHalfSuitCnt {
		if l.halfSuits[half] != LiteratureHalfOpen {
			continue
		}
		if pick < 0 {
			pick = half
		}
		if l.LiteratureHoldsHalfSuit(seat, half) {
			pick = half
			break
		}
	}
	if pick < 0 {
		return -1, nil
	}

	holders := make([]int, LiteratureHalfSuitSize)
	for i, c := range LiteratureHalfSuitCards(pick) {
		holders[i] = seat
		if literatureIndexOf(p, c) >= 0 {
			continue
		}
		// 味方が要求して通った札は、その味方が持っている。
		for _, ask := range l.asks {
			if ask.Card == nil || !ask.Success {
				continue
			}
			if ask.Card.GetDesign() == c.GetDesign() && ask.Card.GetValue() == c.GetValue() &&
				LiteratureTeamOf(ask.From) == team {
				holders[i] = ask.From
			}
		}
	}
	return pick, holders
}

// IsHumanTurn は現在の手番が人間かを返す。
func (l *Literature) IsHumanTurn() bool {
	if l.gameEndFlag {
		return false
	}
	p := l.GetPlayer(l.currentIdx)
	return p != nil && p.GetIsHuman()
}

// CpuPlay は CPU が 1 アクション実行する。
func (l *Literature) CpuPlay() {
	if l.gameEndFlag || l.IsHumanTurn() {
		return
	}
	seat := l.currentIdx
	// **宣言できるなら宣言を優先する。**
	if half, holders := l.LiteratureCpuClaim(seat); half >= 0 {
		if l.Claim(seat, half, holders) == nil {
			return
		}
	}
	if to, c := l.LiteratureCpuAsk(seat); c != nil {
		if l.Ask(seat, to, c) == nil {
			return
		}
	}
	// **訊けないなら宣言するしかない。**パスは規則に無く、認めるとゲームが
	// 終わらなくなる。
	if half, holders := l.LiteratureCpuForcedClaim(seat); half >= 0 {
		if l.Claim(seat, half, holders) == nil {
			return
		}
	}
	l.passTurn()
}

// passTurn は手札のある次の席へ手番を渡す。
func (l *Literature) passTurn() {
	for i := 1; i <= LiteraturePlayerCnt; i++ {
		seat := (l.currentIdx + i) % LiteraturePlayerCnt
		if q := l.GetPlayer(seat); q != nil && q.GetCardsSize() > 0 {
			l.currentIdx = seat
			return
		}
	}
	// 誰の手札も残っていない。決着させる。
	l.checkGameEnd()
	if !l.gameEndFlag {
		l.finish(-1)
	}
}

// ---- アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (l *Literature) GetPlayers() []*LiteraturePlayer { return l.players }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (l *Literature) GetPlayer(idx int) *LiteraturePlayer {
	return getPlayer(l.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (l *Literature) GetPhase() LiteraturePhase { return l.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (l *Literature) GetCurrentPlayerIdx() int { return l.currentIdx }

// GetHalfSuitState はハーフスートの帰属を返す。
func (l *Literature) GetHalfSuitState(half int) LiteratureHalfSuitState {
	if half < 0 || half >= LiteratureHalfSuitCnt {
		return LiteratureHalfOpen
	}
	return l.halfSuits[half]
}

// GetAsks は要求の履歴を返す。**公開情報。**
func (l *Literature) GetAsks() []*LiteratureAsk { return l.asks }

// GetClaims は宣言の履歴を返す。
func (l *Literature) GetClaims() []*LiteratureClaimResult { return l.claims }

// GetLastAsk は直前の要求を返す。
func (l *Literature) GetLastAsk() *LiteratureAsk { return l.lastAsk }

// GetLastClaim は直前の宣言を返す。
func (l *Literature) GetLastClaim() *LiteratureClaimResult { return l.lastClaim }

// GetGameEndFlag はゲーム終了フラグを返す。
func (l *Literature) GetGameEndFlag() bool { return l.gameEndFlag }

// GetWinnerTeam は勝利チームを返す (未確定・引き分けなら -1)。
func (l *Literature) GetWinnerTeam() int { return l.winnerTeam }

// GetConfig はゲーム設定を返す。
func (l *Literature) GetConfig() LiteratureConfig { return l.config }

// SetConfig はゲーム設定をセットする。
func (l *Literature) SetConfig(c LiteratureConfig) { l.config = c }

// GetActionLog は棋譜を返す。
func (l *Literature) GetActionLog() []*ActionLogEntry { return l.actionLog }

// addLog は棋譜を 1 件追加する。
func (l *Literature) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	l.appendLogAt(0, playerIdx, actionType, detail, cards)
}

// literatureCardName は札の内部名を返す (棋譜用)。
func literatureCardName(c *Card) string {
	if c == nil {
		return "-"
	}
	return strconv.Itoa(c.GetDesign()) + "-" + strconv.Itoa(c.GetValue())
}

// ---- テスト用 ----

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (l *Literature) SetPhaseForTest(p LiteraturePhase) { l.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (l *Literature) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(l.GetPlayer(idx), cards)
}

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (l *Literature) SetCurrentPlayerForTest(idx int) { l.currentIdx = idx }

// SetHalfSuitForTest はハーフスートの帰属を差し替える (テスト専用)。
func (l *Literature) SetHalfSuitForTest(half int, st LiteratureHalfSuitState) {
	if half >= 0 && half < LiteratureHalfSuitCnt {
		l.halfSuits[half] = st
	}
}

// CheckGameEndForTest は決着判定を走らせる (テスト専用)。
func (l *Literature) CheckGameEndForTest() { l.checkGameEnd() }

// ---- JSON ----

// literatureJSON is the KV wire format for Literature.
type literatureJSON struct {
	Players     []*LiteraturePlayer                            `json:"pl"`
	Config      LiteratureConfig                               `json:"cf"`
	Phase       LiteraturePhase                                `json:"ph"`
	CurrentIdx  int                                            `json:"ci"`
	HalfSuits   [LiteratureHalfSuitCnt]LiteratureHalfSuitState `json:"hs"`
	Asks        []*LiteratureAsk                               `json:"as"`
	Claims      []*LiteratureClaimResult                       `json:"cl"`
	LastAsk     *LiteratureAsk                                 `json:"la"`
	LastClaim   *LiteratureClaimResult                         `json:"lc"`
	GameEndFlag bool                                           `json:"ge"`
	WinnerTeam  int                                            `json:"wt"`
	ActionLog   []*ActionLogEntry                              `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (l *Literature) MarshalJSON() ([]byte, error) {
	return json.Marshal(literatureJSON{
		Players: l.players, Config: l.config, Phase: l.phase,
		CurrentIdx: l.currentIdx, HalfSuits: l.halfSuits,
		Asks: l.asks, Claims: l.claims, LastAsk: l.lastAsk, LastClaim: l.lastClaim,
		GameEndFlag: l.gameEndFlag, WinnerTeam: l.winnerTeam, ActionLog: l.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **KV から戻る値なので範囲を検査する。**壊れた状態をそのまま受け入れると
// 添字で落ちる。
func (l *Literature) UnmarshalJSON(data []byte) error {
	var j literatureJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != LiteraturePlayerCnt {
		return errors.New("literature needs exactly six seats")
	}
	if j.Phase < LiteraturePhasePlay || j.Phase > LiteraturePhaseGameEnd {
		return errors.New("unknown phase")
	}
	if j.CurrentIdx < 0 || j.CurrentIdx >= LiteraturePlayerCnt {
		return errors.New("bad current seat")
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= LiteratureTeamCnt {
		return errors.New("bad winner team")
	}
	for _, st := range j.HalfSuits {
		if st < LiteratureHalfOpen || st > LiteratureHalfCancelled {
			return errors.New("bad half-suit state")
		}
	}
	for _, cl := range j.Claims {
		if cl == nil {
			continue
		}
		if cl.HalfSuit < 0 || cl.HalfSuit >= LiteratureHalfSuitCnt {
			return errors.New("bad claimed half-suit")
		}
		if cl.Outcome < LiteratureClaimWon || cl.Outcome > LiteratureClaimLost {
			return errors.New("bad claim outcome")
		}
	}
	for _, a := range j.Asks {
		if a == nil {
			continue
		}
		if a.From < 0 || a.From >= LiteraturePlayerCnt || a.To < 0 || a.To >= LiteraturePlayerCnt {
			return errors.New("bad ask seat")
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}

	l.players = j.Players
	l.config = j.Config
	l.phase = j.Phase
	l.currentIdx = j.CurrentIdx
	l.halfSuits = j.HalfSuits
	l.asks = j.Asks
	l.claims = j.Claims
	l.lastAsk = j.LastAsk
	l.lastClaim = j.LastClaim
	l.gameEndFlag = j.GameEndFlag
	l.winnerTeam = j.WinnerTeam
	l.actionLog = j.ActionLog
	return nil
}
