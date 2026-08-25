//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"fmt"
)

// 卓の形。
const (
	// DilotiPlayerCnt は席数 (人間 1 + CPU 1)。
	DilotiPlayerCnt = 2
	// DilotiHandSize は 1 回に配る枚数。
	DilotiHandSize = 6
	// DilotiTableSize は局の初めに場へ置く枚数。**配り直しでは足さない。**
	DilotiTableSize = 4
	// DilotiDeckSize は使う札の枚数。
	DilotiDeckSize = 52
)

// フェーズ。
const (
	// DilotiPhasePlay は打っている最中。
	DilotiPhasePlay = "play"
	// DilotiPhaseRoundEnd は 1 局の集計を見せている状態。
	DilotiPhaseRoundEnd = "roundEnd"
	// DilotiPhaseGameEnd は終局。
	DilotiPhaseGameEnd = "gameEnd"
)

// 手の種類。
const (
	// DilotiActionCapture は取る手。
	DilotiActionCapture = "capture"
	// DilotiActionDeclare は宣言する手。
	DilotiActionDeclare = "declare"
	// DilotiActionTrail は場に置く手。
	DilotiActionTrail = "trail"
)

// 得点。
//
// **クセリ 1 回で 10 点。** 最多枚数 4 点より重く、11 点しかない固定点の中で
// 決定的な差になる ── 場を 1 枚で払う手を探すのがこのゲームの中心にある。
const (
	// DilotiXeriPoints はクセリ 1 回の点。
	DilotiXeriPoints = 10
	// DilotiMostCardsPoints は最多枚数の点。
	DilotiMostCardsPoints = 4
	// DilotiTenOfDiamondsPoints は 10♦ の点。
	DilotiTenOfDiamondsPoints = 2
	// DilotiTwoOfClubsPoints は 2♣ の点。
	DilotiTwoOfClubsPoints = 1
	// DilotiAcePoints は A 1 枚の点。
	DilotiAcePoints = 1
	// DilotiHalfDeck は最多枚数が引き分けになる枚数。
	DilotiHalfDeck = DilotiDeckSize / 2
)

// 得点項目の識別子。
const (
	DilotiScoreCards         = "cards"
	DilotiScoreAces          = "aces"
	DilotiScoreTenOfDiamonds = "tenOfDiamonds"
	DilotiScoreTwoOfClubs    = "twoOfClubs"
	DilotiScoreXeri          = "xeri"
)

// DilotiScoreLine は 1 つの得点項目。
type DilotiScoreLine struct {
	// Key は項目の識別子。
	Key string
	// Points は席ごとの点。
	Points []int
}

// DilotiRoundResult は 1 局の集計。
type DilotiRoundResult struct {
	// Lines は項目別の内訳。
	Lines []DilotiScoreLine
	// Totals は席ごとのこの局の合計。
	Totals [DilotiPlayerCnt]int
	// CardCounts は席ごとの取り札枚数。
	CardCounts [DilotiPlayerCnt]int
	// Xeris は席ごとのクセリ回数。
	Xeris [DilotiPlayerCnt]int
}

// Diloti はディロティの状態を保持する集約ルート。
type Diloti struct {
	deck    []*Card
	drawIdx int
	players []*DilotiPlayer
	config  DilotiConfig
	// table は場の緩い札。
	table []*Card
	// decls は場に積まれた宣言。
	decls       []*DilotiDeclaration
	phase       string
	roundNumber int
	dealerIdx   int
	currentIdx  int
	// lastCapturer は最後に取った席。**山が尽きたときの場札はここへ渡る。**
	lastCapturer int
	// firstPlayDone はこの局の 1 手目が済んだか。
	//
	// **局の初手はクセリにならない。** 配った直後の 4 枚を 1 枚で払っても
	// 数えない ── 実力ではなく配りだからで、原典もそこだけ明示的に除いている。
	firstPlayDone bool
	lastResult    *DilotiRoundResult
	gameEndFlag   bool
	winnerIdx     int
	actionLogBase
}

// NewDiloti はコンストラクタ。
func NewDiloti(players []*DilotiPlayer, config DilotiConfig) *Diloti {
	return &Diloti{
		players:      players,
		config:       config,
		phase:        DilotiPhasePlay,
		lastCapturer: -1,
		winnerIdx:    -1,
	}
}

// NewDefaultDiloti は既定の設定でディロティを生成する。
func NewDefaultDiloti() *Diloti {
	players := []*DilotiPlayer{NewDilotiPlayer(true), NewDilotiPlayer(false)}
	return NewDiloti(players, DefaultDilotiConfig())
}

// Reset はゲームを最初から始める。
//
// **親を席 1 にして席 0 から打たせる。** 非親が先に打つ規則なので、親を 0 に
// すると人間は最初の 4 枚に一度も手を出せないまま CPU が場を払う。
func (d *Diloti) Reset() {
	for _, p := range d.players {
		p.ResetRound()
		p.ResetScore()
	}
	d.roundNumber = 1
	d.dealerIdx = DilotiPlayerCnt - 1
	d.gameEndFlag = false
	d.winnerIdx = -1
	d.lastResult = nil
	d.actionLog = make([]*ActionLogEntry, 0)
	d.startRound()
}

// NextRound は次の局を始める。
func (d *Diloti) NextRound() {
	if d.gameEndFlag || d.phase != DilotiPhaseRoundEnd {
		return
	}
	d.roundNumber++
	d.dealerIdx = (d.dealerIdx + 1) % DilotiPlayerCnt
	d.startRound()
}

// startRound は山を作り、6 枚ずつ配って場に 4 枚置く。
func (d *Diloti) startRound() {
	for _, p := range d.players {
		p.ResetRound()
	}
	src := NewTrumpCards(0)
	src.Shuffle()
	d.deck = make([]*Card, 0, DilotiDeckSize)
	for {
		c := src.DrawCard()
		if c == nil {
			break
		}
		d.deck = append(d.deck, c)
	}
	d.drawIdx = 0
	d.decls = make([]*DilotiDeclaration, 0)
	d.table = make([]*Card, 0, DilotiTableSize)
	d.dealHands()
	// **場札を配るのは局の初めだけ。** 手札が尽きての配り直しでは足さない。
	for i := 0; i < DilotiTableSize; i++ {
		if c := d.draw(); c != nil {
			d.table = append(d.table, c)
		}
	}
	d.phase = DilotiPhasePlay
	d.lastCapturer = -1
	d.firstPlayDone = false
	d.currentIdx = (d.dealerIdx + 1) % DilotiPlayerCnt
	d.appendLog(-1, "deal", fmt.Sprintf("round %d: dealer=%d, table=%d",
		d.roundNumber, d.dealerIdx, len(d.table)), nil)
}

// dealHands は各席へ 6 枚ずつ配る。
func (d *Diloti) dealHands() {
	for i := 0; i < DilotiHandSize; i++ {
		for seat := 0; seat < DilotiPlayerCnt; seat++ {
			idx := (d.dealerIdx + 1 + seat) % DilotiPlayerCnt
			if c := d.draw(); c != nil {
				d.players[idx].AddCard(c)
			}
		}
	}
}

// draw は山から 1 枚引く。
func (d *Diloti) draw() *Card {
	if d.drawIdx >= len(d.deck) {
		return nil
	}
	c := d.deck[d.drawIdx]
	d.drawIdx++
	return c
}

// PlayerPlay は人間が 1 手打つ。
func (d *Diloti) PlayerPlay(handIdx int, action string, tableIdxs, declIdxs []int, declValue int) error {
	human := findHumanIdx(d.players)
	if human < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errNoHuman", nil)
	}
	if d.gameEndFlag {
		return NewDomainErrorCode(ErrGameEnded, "diloti.errGameEnded", nil)
	}
	if d.phase != DilotiPhasePlay {
		return NewDomainErrorCode(ErrWrongPhase, "diloti.errNotPlayPhase", nil)
	}
	if d.currentIdx != human {
		return NewDomainErrorCode(ErrNotHumanTurn, "diloti.errNotYourTurn", nil)
	}
	return d.applyPlay(human, handIdx, action, tableIdxs, declIdxs, declValue)
}

// applyPlay は 1 手を盤面へ反映する。
func (d *Diloti) applyPlay(seat, handIdx int, action string, tableIdxs, declIdxs []int, declValue int) error {
	p := d.players[seat]
	hand := p.GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return NewDomainErrorCode(ErrInvalidCard, "diloti.errCardRange",
			map[string]string{"idx": fmt.Sprint(handIdx)})
	}
	card := hand[handIdx]

	// **宣言した札は約束のために取っておく。** 使い切ってしまうと、宣言を
	// 抱えたまま取る手も置く手も無い ── 手番が回ってきても打てない盤面が
	// 生まれる。原典どおり「取り切るまで手放せない」を規則として持たせる。
	if err := d.guardBackingCard(seat, card, action, declIdxs); err != nil {
		return err
	}

	switch action {
	case DilotiActionCapture:
		if err := d.doCapture(seat, card, tableIdxs, declIdxs); err != nil {
			return err
		}
	case DilotiActionDeclare:
		if err := d.doDeclare(seat, card, tableIdxs, declIdxs, declValue); err != nil {
			return err
		}
	case DilotiActionTrail:
		if err := d.doTrail(seat, card); err != nil {
			return err
		}
	default:
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errUnknownAction",
			map[string]string{"action": action})
	}

	p.RemoveCard(handIdx)
	d.firstPlayDone = true
	d.advance()
	return nil
}

// doCapture は取る手を反映する。
func (d *Diloti) doCapture(seat int, card *Card, tableIdxs, declIdxs []int) error {
	if !IsValidDilotiCapture(card, d.table, d.decls, tableIdxs, declIdxs) {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errIllegalCapture", nil)
	}
	// **場を 1 枚で払い切ったらクセリ。** 判定は取る前の場の大きさで決まるので、
	// 取り除く前に「全部を選んでいるか」を見る。
	sweeps := len(tableIdxs) == len(d.table) && len(declIdxs) == len(d.decls) &&
		(len(d.table)+len(d.decls)) > 0

	taken := []*Card{card}
	for _, i := range tableIdxs {
		taken = append(taken, d.table[i])
	}
	for _, i := range declIdxs {
		taken = append(taken, d.decls[i].AllCards()...)
	}
	d.table = dilotiRemoveAt(d.table, tableIdxs)
	d.decls = dilotiRemoveDeclsAt(d.decls, declIdxs)
	d.players[seat].AddCaptured(taken)
	d.lastCapturer = seat
	d.appendLog(seat, "capture", fmt.Sprintf("player %d captures %d card(s)", seat, len(taken)-1), taken)

	// **局の初手はクセリにならない。**
	if sweeps && d.firstPlayDone {
		d.players[seat].AddXeri()
		d.appendLog(seat, "xeri", fmt.Sprintf("player %d sweeps the table", seat), nil)
	}
	return nil
}

// doTrail は場に置く手を反映する。
func (d *Diloti) doTrail(seat int, card *Card) error {
	// **宣言した側は取り切るまで置けない。** 置けてしまうと、宣言は取られない
	// まま場に残り続け、約束が意味を失う。
	if d.hasOutstandingDeclaration(seat) {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errMustAnswerDeclaration", nil)
	}
	if !CanTrailDiloti(card, d.table) {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errFaceMustCapture", nil)
	}
	d.table = append(d.table, card)
	d.appendLog(seat, "trail", fmt.Sprintf("player %d lays a card", seat), []*Card{card})
	return nil
}

// doDeclare は宣言する手を反映する。declIdxs が空なら新しい単一宣言、
// 1 つ指していればその宣言を上げる (値が変わる) かグループにする (値が同じ)。
func (d *Diloti) doDeclare(seat int, card *Card, tableIdxs, declIdxs []int, value int) error {
	if DilotiIsFaceCard(card) {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errFaceCannotDeclare", nil)
	}
	if value < DilotiMinDeclaration || value > DilotiMaxDeclaration {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errDeclarationRange",
			map[string]string{"val": fmt.Sprint(value)})
	}
	if !dilotiIdxsInRange(tableIdxs, len(d.table)) || !dilotiIdxsInRange(declIdxs, len(d.decls)) {
		return NewDomainErrorCode(ErrInvalidIndices, "diloti.errIndexRange", nil)
	}
	if len(declIdxs) > 1 {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errOneDeclarationAtATime", nil)
	}
	// **裏付けの札を残していない宣言はできない。**
	if !dilotiHoldsValue(d.players[seat].GetHand(), d.handIndexOf(seat, card), value) {
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errNoBackingCard",
			map[string]string{"val": fmt.Sprint(value)})
	}

	sum := DilotiCardValue(card)
	picked := make([]*Card, 0, len(tableIdxs)+1)
	picked = append(picked, card)
	for _, i := range tableIdxs {
		c := d.table[i]
		if DilotiIsFaceCard(c) {
			return NewDomainErrorCode(ErrInvalidPlay, "diloti.errFaceInDeclaration", nil)
		}
		sum += DilotiCardValue(c)
		picked = append(picked, c)
	}

	if len(declIdxs) == 0 {
		if sum != value {
			return NewDomainErrorCode(ErrInvalidPlay, "diloti.errDeclarationSum",
				map[string]string{"sum": fmt.Sprint(sum), "val": fmt.Sprint(value)})
		}
		d.table = dilotiRemoveAt(d.table, tableIdxs)
		d.decls = append(d.decls, NewDilotiDeclaration(seat, value, picked))
		d.appendLog(seat, "declare", fmt.Sprintf("player %d declares %d", seat, value), picked)
		return nil
	}

	target := d.decls[declIdxs[0]]
	switch {
	case value == target.Value:
		// グループ宣言: 同じ値の束を足す。**上げられなくなり、丸ごとしか取れない。**
		if sum != value {
			return NewDomainErrorCode(ErrInvalidPlay, "diloti.errDeclarationSum",
				map[string]string{"sum": fmt.Sprint(sum), "val": fmt.Sprint(value)})
		}
		d.table = dilotiRemoveAt(d.table, tableIdxs)
		target.AddGroup(picked)
		target.OwnerIdx = seat
		d.appendLog(seat, "group", fmt.Sprintf("player %d groups %d", seat, value), picked)
		return nil
	case target.IsGroup:
		// **グループ宣言は上げられない。**
		return NewDomainErrorCode(ErrInvalidPlay, "diloti.errGroupCannotBeRaised", nil)
	default:
		// 上げる: 既存の束へ札を足し、合計が新しい宣言値になる。
		if value <= target.Value {
			return NewDomainErrorCode(ErrInvalidPlay, "diloti.errRaiseMustGoUp",
				map[string]string{"val": fmt.Sprint(target.Value)})
		}
		if sum+target.Value != value {
			return NewDomainErrorCode(ErrInvalidPlay, "diloti.errDeclarationSum",
				map[string]string{"sum": fmt.Sprint(sum + target.Value), "val": fmt.Sprint(value)})
		}
		d.table = dilotiRemoveAt(d.table, tableIdxs)
		target.Groups[0] = append(target.Groups[0], picked...)
		target.Value = value
		target.OwnerIdx = seat
		d.appendLog(seat, "raise", fmt.Sprintf("player %d raises to %d", seat, value), picked)
		return nil
	}
}

// guardBackingCard は宣言の裏付け札を守る。
//
// 宣言を抱えている席が、その宣言値と同じ札を「宣言を取る手」以外に使おうと
// したとき、同じ値の札が他に残っていなければ弾く。
func (d *Diloti) guardBackingCard(seat int, card *Card, action string, declIdxs []int) error {
	owed := -1
	owedIdx := -1
	for i, x := range d.decls {
		if x != nil && x.OwnerIdx == seat {
			owed, owedIdx = x.Value, i
			break
		}
	}
	if owed < 0 || DilotiCardValue(card) != owed {
		return nil
	}
	// 抱えた宣言を取る手なら、まさにそのための札なので通す。
	if action == DilotiActionCapture {
		for _, i := range declIdxs {
			if i == owedIdx {
				return nil
			}
		}
	}
	// 同じ値の札がもう 1 枚あるなら手放してよい。
	if dilotiHoldsValue(d.players[seat].GetHand(), d.handIndexOf(seat, card), owed) {
		return nil
	}
	return NewDomainErrorCode(ErrInvalidPlay, "diloti.errKeepBackingCard",
		map[string]string{"val": fmt.Sprint(owed)})
}

// handIndexOf は席 seat の手札における card の位置を返す (無ければ -1)。
func (d *Diloti) handIndexOf(seat int, card *Card) int {
	for i, c := range d.players[seat].GetHand() {
		if c == card {
			return i
		}
	}
	return -1
}

// hasOutstandingDeclaration は席 seat が取り切っていない宣言を持つかを返す。
func (d *Diloti) hasOutstandingDeclaration(seat int) bool {
	for _, x := range d.decls {
		if x != nil && x.OwnerIdx == seat {
			return true
		}
	}
	return false
}

// dilotiRemoveAt は idxs の位置を取り除いた新しいスライスを返す。
func dilotiRemoveAt(cards []*Card, idxs []int) []*Card {
	drop := make(map[int]struct{}, len(idxs))
	for _, i := range idxs {
		drop[i] = struct{}{}
	}
	out := make([]*Card, 0, len(cards))
	for i, c := range cards {
		if _, gone := drop[i]; gone {
			continue
		}
		out = append(out, c)
	}
	return out
}

// dilotiRemoveDeclsAt は idxs の位置の宣言を取り除く。
func dilotiRemoveDeclsAt(decls []*DilotiDeclaration, idxs []int) []*DilotiDeclaration {
	drop := make(map[int]struct{}, len(idxs))
	for _, i := range idxs {
		drop[i] = struct{}{}
	}
	out := make([]*DilotiDeclaration, 0, len(decls))
	for i, x := range decls {
		if _, gone := drop[i]; gone {
			continue
		}
		out = append(out, x)
	}
	return out
}

// advance は手番を進め、必要なら配り直し・局の終了を行う。
func (d *Diloti) advance() {
	if d.handsEmpty() {
		if d.drawIdx < len(d.deck) {
			// **配り直しでは場札を足さない。** 場は打ち手が積んだものだけ。
			d.dealHands()
			d.currentIdx = (d.currentIdx + 1) % DilotiPlayerCnt
			return
		}
		d.finishRound()
		return
	}
	d.currentIdx = (d.currentIdx + 1) % DilotiPlayerCnt
}

// handsEmpty は全席の手札が尽きたかを返す。
func (d *Diloti) handsEmpty() bool {
	for _, p := range d.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishRound は取り残しを回収して集計する。
func (d *Diloti) finishRound() {
	// **山が尽きたときの場札は最後に取った側へ。** ただしこれはクセリではない。
	if leftover := len(d.table) + len(d.decls); leftover > 0 && d.lastCapturer >= 0 {
		rest := append([]*Card(nil), d.table...)
		for _, x := range d.decls {
			rest = append(rest, x.AllCards()...)
		}
		d.players[d.lastCapturer].AddCaptured(rest)
		d.appendLog(d.lastCapturer, "sweep",
			fmt.Sprintf("player %d takes the %d card(s) left on the table", d.lastCapturer, len(rest)), rest)
		d.table = make([]*Card, 0)
		d.decls = make([]*DilotiDeclaration, 0)
	}
	d.lastResult = d.scoreRound()
	for i := range d.players {
		d.players[i].AddScore(d.lastResult.Totals[i])
	}
	d.appendLog(-1, "roundEnd", fmt.Sprintf("round %d: %d - %d",
		d.roundNumber, d.lastResult.Totals[0], d.lastResult.Totals[1]), nil)
	d.phase = DilotiPhaseRoundEnd
	d.checkGameEnd()
}

// scoreRound は 1 局を集計する。
//
// **点は 11 点しか固定で無く、残りはすべてクセリ。** 最多枚数 4・A 各 1・
// 10♦ 2・2♣ 1 で合計 11 点、そこへ 1 回 10 点のクセリが乗る。
func (d *Diloti) scoreRound() *DilotiRoundResult {
	res := &DilotiRoundResult{}
	counts := [DilotiPlayerCnt]int{}
	for i, p := range d.players {
		counts[i] = len(p.GetCaptured())
		res.CardCounts[i] = counts[i]
		res.Xeris[i] = p.GetXeri()
	}

	// 最多枚数。**26 対 26 はどちらにも入らない。**
	// 原典は「分ける」と「どちらにも入らない」で割れているが、4 点を 2 点ずつに
	// 割る読みは 61 点への到達を早めるだけで、引き分けを引き分けとして扱う
	// もうひとつの読みのほうが素直なのでこちらを採る。
	cards := make([]int, DilotiPlayerCnt)
	switch {
	case counts[0] > counts[1]:
		cards[0] = DilotiMostCardsPoints
	case counts[1] > counts[0]:
		cards[1] = DilotiMostCardsPoints
	}
	res.Lines = append(res.Lines, DilotiScoreLine{Key: DilotiScoreCards, Points: cards})

	aces := make([]int, DilotiPlayerCnt)
	ten := make([]int, DilotiPlayerCnt)
	two := make([]int, DilotiPlayerCnt)
	for i, p := range d.players {
		for _, c := range p.GetCaptured() {
			switch {
			case c.GetValue() == 1:
				aces[i] += DilotiAcePoints
			case c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond:
				ten[i] += DilotiTenOfDiamondsPoints
			case c.GetValue() == 2 && c.GetDesign() == CardDesignClover:
				two[i] += DilotiTwoOfClubsPoints
			}
		}
	}
	res.Lines = append(res.Lines,
		DilotiScoreLine{Key: DilotiScoreAces, Points: aces},
		DilotiScoreLine{Key: DilotiScoreTenOfDiamonds, Points: ten},
		DilotiScoreLine{Key: DilotiScoreTwoOfClubs, Points: two})

	xeri := make([]int, DilotiPlayerCnt)
	for i, p := range d.players {
		xeri[i] = p.GetXeri() * DilotiXeriPoints
	}
	res.Lines = append(res.Lines, DilotiScoreLine{Key: DilotiScoreXeri, Points: xeri})

	for _, line := range res.Lines {
		for i, pt := range line.Points {
			res.Totals[i] += pt
		}
	}
	return res
}

// checkGameEnd は目標点に届いたかを見る。
//
// **同点では終わらない。** 両者が同時に届いたときは高いほうが勝ち、
// 並んだままなら次の局へ持ち越す。
func (d *Diloti) checkGameEnd() {
	best, bestIdx, tied := -1, -1, false
	for i, p := range d.players {
		switch {
		case p.GetScore() > best:
			best, bestIdx, tied = p.GetScore(), i, false
		case p.GetScore() == best:
			tied = true
		}
	}
	if best < d.config.TargetScore || tied {
		return
	}
	d.gameEndFlag = true
	d.winnerIdx = bestIdx
	d.phase = DilotiPhaseGameEnd
	d.appendLog(-1, "gameEnd", fmt.Sprintf("player %d wins the match", bestIdx), nil)
}

// CpuPlay は CPU が 1 手打つ。
func (d *Diloti) CpuPlay() {
	if d.gameEndFlag || d.phase != DilotiPhasePlay {
		return
	}
	seat := d.currentIdx
	if d.players[seat].GetIsHuman() {
		return
	}
	move := d.chooseCpuMove(seat)
	if move == nil {
		return
	}
	_ = d.applyPlay(seat, move.HandIdx, move.Action, move.TableIdxs, move.DeclIdxs, move.Value)
}

// IsHumanTurn は人間の手番かを返す。
func (d *Diloti) IsHumanTurn() bool {
	human := findHumanIdx(d.players)
	return human >= 0 && d.currentIdx == human && d.phase == DilotiPhasePlay && !d.gameEndFlag
}

// GetConfig はゲーム設定を返す。
func (d *Diloti) GetConfig() DilotiConfig { return d.config }

// SetConfig はゲーム設定を差し替える。
func (d *Diloti) SetConfig(cfg DilotiConfig) { d.config = cfg }

// GetGameEndFlag は終局フラグを返す。
func (d *Diloti) GetGameEndFlag() bool { return d.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (d *Diloti) GetPhase() string { return d.phase }

// GetTable は場の緩い札を返す。
func (d *Diloti) GetTable() []*Card { return d.table }

// GetDeclarations は場に積まれた宣言を返す。
func (d *Diloti) GetDeclarations() []*DilotiDeclaration { return d.decls }

// GetRoundNumber は現在の局を返す。
func (d *Diloti) GetRoundNumber() int { return d.roundNumber }

// GetDealerIdx は親の席を返す。
func (d *Diloti) GetDealerIdx() int { return d.dealerIdx }

// GetCurrentPlayerIdx は手番の席を返す。
func (d *Diloti) GetCurrentPlayerIdx() int { return d.currentIdx }

// GetLastCapturer は最後に取った席を返す (-1 = まだ無い)。
func (d *Diloti) GetLastCapturer() int { return d.lastCapturer }

// GetDeckRemaining は山の残り枚数を返す。
func (d *Diloti) GetDeckRemaining() int { return len(d.deck) - d.drawIdx }

// GetPlayerCnt は席数を返す。
func (d *Diloti) GetPlayerCnt() int { return len(d.players) }

// GetPlayer は指定席のプレイヤーを返す。
func (d *Diloti) GetPlayer(i int) *DilotiPlayer {
	if i < 0 || i >= len(d.players) {
		return nil
	}
	return d.players[i]
}

// GetLastResult は直前の局の集計を返す。
func (d *Diloti) GetLastResult() *DilotiRoundResult { return d.lastResult }

// GetWinnerIdx は勝者の席を返す (-1 = 未決)。
func (d *Diloti) GetWinnerIdx() int { return d.winnerIdx }

// GetTakeOptions は席 seat が手札 handIdx を出したときの取り手を返す。
func (d *Diloti) GetTakeOptions(seat, handIdx int) []DilotiTake {
	if seat < 0 || seat >= len(d.players) {
		return nil
	}
	hand := d.players[seat].GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return nil
	}
	return EnumerateDilotiTakes(hand[handIdx], d.table, d.decls)
}

// GetDeclareOptions は席 seat が手札 handIdx で作れる新しい単一宣言を返す。
func (d *Diloti) GetDeclareOptions(seat, handIdx int) []DilotiDeclCandidate {
	if seat < 0 || seat >= len(d.players) {
		return nil
	}
	hand := d.players[seat].GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return nil
	}
	// **宣言を抱えている間は新しい宣言をしない。**
	if d.hasOutstandingDeclaration(seat) {
		return nil
	}
	return EnumerateDilotiDeclarations(hand[handIdx], handIdx, hand, d.table)
}

// CanTrail は席 seat が手札 handIdx を場に置けるかを返す。
func (d *Diloti) CanTrail(seat, handIdx int) bool {
	if seat < 0 || seat >= len(d.players) || d.hasOutstandingDeclaration(seat) {
		return false
	}
	hand := d.players[seat].GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return false
	}
	return CanTrailDiloti(hand[handIdx], d.table)
}

// dilotiJSON is the JSON wire format for Diloti.
type dilotiJSON struct {
	Deck          []*Card              `json:"dk"`
	DrawIdx       int                  `json:"di"`
	Players       []*DilotiPlayer      `json:"pl"`
	Config        DilotiConfig         `json:"cf"`
	Table         []*Card              `json:"tb"`
	Decls         []*DilotiDeclaration `json:"dc"`
	Phase         string               `json:"ph"`
	RoundNumber   int                  `json:"rn"`
	DealerIdx     int                  `json:"dl"`
	CurrentIdx    int                  `json:"cu"`
	LastCapturer  int                  `json:"lc"`
	FirstPlayDone bool                 `json:"fp"`
	LastResult    *DilotiRoundResult   `json:"lr"`
	GameEndFlag   bool                 `json:"ge"`
	WinnerIdx     int                  `json:"wi"`
	ActionLog     []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** 保存した
// 盤で打ち続けられなくなるので、全部の欄を明示する。
func (d *Diloti) MarshalJSON() ([]byte, error) {
	return json.Marshal(dilotiJSON{
		Deck: d.deck, DrawIdx: d.drawIdx, Players: d.players, Config: d.config,
		Table: d.table, Decls: d.decls, Phase: d.phase, RoundNumber: d.roundNumber,
		DealerIdx: d.dealerIdx, CurrentIdx: d.currentIdx, LastCapturer: d.lastCapturer,
		FirstPlayDone: d.firstPlayDone, LastResult: d.lastResult,
		GameEndFlag: d.gameEndFlag, WinnerIdx: d.winnerIdx, ActionLog: d.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Diloti) UnmarshalJSON(data []byte) error {
	var j dilotiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.deck, d.drawIdx, d.config = j.Deck, j.DrawIdx, j.Config
	d.players = j.Players
	if len(d.players) != DilotiPlayerCnt {
		return fmt.Errorf("diloti: expected %d players, got %d", DilotiPlayerCnt, len(d.players))
	}
	d.table = j.Table
	if d.table == nil {
		d.table = make([]*Card, 0)
	}
	d.decls = j.Decls
	if d.decls == nil {
		d.decls = make([]*DilotiDeclaration, 0)
	}
	d.phase, d.roundNumber = j.Phase, j.RoundNumber
	d.dealerIdx, d.currentIdx, d.lastCapturer = j.DealerIdx, j.CurrentIdx, j.LastCapturer
	d.firstPlayDone, d.lastResult = j.FirstPlayDone, j.LastResult
	d.gameEndFlag, d.winnerIdx = j.GameEndFlag, j.WinnerIdx
	d.actionLog = j.ActionLog
	if d.actionLog == nil {
		d.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// SetTableForTest は場札と宣言を差し替える (テスト用)。
//
// **場は配りで決まるので、狙った盤面は組めない。** 捕獲規則や宣言の見え方を
// 確かめるにはここで固定するしかない。
func (d *Diloti) SetTableForTest(cards []*Card, decls []*DilotiDeclaration) {
	d.table = cards
	if decls == nil {
		decls = make([]*DilotiDeclaration, 0)
	}
	d.decls = decls
}

// SetFirstPlayDoneForTest は局の初手が済んだ印を立てる (テスト用)。
func (d *Diloti) SetFirstPlayDoneForTest(v bool) { d.firstPlayDone = v }
