//go:build !js || !wasm || extra2

// Package domain — ロバ (Loba de Menos) のドメインモデル。
//
// アルゼンチンのラミー。**2 組 + ジョーカー 4 枚の 108 枚**を使い、手札を出し切って
// 相手に失点を残す。101 点に達した人から脱落し、最後に残った人が勝ち。
// 実装は pagat.com の Loba de Menos に従う。
//
// # issue #4414 の仕様案との相違
//
// **Loba はコンティンジェント・ラミーではない。**issue の中核の前提が成立しない。
//
//   - issue は「局ごとに定められた**コントラクト**を順に進行」「全コントラクトを
//     終えた時点の合計失点最少者が勝利」とするが、Loba に**固定契約は存在しない**。
//     局ごとに要求が変わるのは Carioca / Contract Rummy の構造。Loba は
//     **101 点に達した人が脱落**し、最後に残った人が勝つ
//   - issue は 3〜6 人とするが、**2〜5 人**
//   - 配札は **9 枚** (もう一方の Loba de Mas は 11 枚)
//
// issue が触れていない、Loba を Loba たらしめている規則:
//
//   - **ピエルナ (同ランク組) は「異なる 3 スート」でなければならない。**付け足せる
//     のもその 3 スートに限られる
//   - **ジョーカーはワイルドではない。**エスカレラ (同スートの並び) にしか置けず、
//     1 つのエスカレラに 1 枚まで。しかも通常は捨てられない
//   - 手札を一度も出さずに一気に上がると **-10**
//   - 手札の点は数札が額面、**ジョーカー・K・Q・J・A は 10 点**
//
// # 実装から外したもの
//
// 脱落者の復帰 (reengancharse) は入れていない。原典でも «can be reincorporated»
// と任意で、復帰の可否を誰がいつ決めるのかが規定されていないため、単独プレイの
// 対 CPU では判断の持ち主が決まらない。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// LobaPlayerCnt はプレイヤー数 (4 人。原典は 2〜5 人)。
const LobaPlayerCnt = 4

// LobaHandSize は配札枚数。
const LobaHandSize = 9

// LobaDeckSize は 108 枚 (52 × 2 + ジョーカー 4)。
const LobaDeckSize = 108

// LobaJokerCnt はジョーカーの枚数。
const LobaJokerCnt = 4

// LobaKnockOut は脱落する失点。
const LobaKnockOut = 101

// LobaGoOutCleanBonus は一度も出さずに上がったときの加点 (失点から引く)。
const LobaGoOutCleanBonus = 10

// LobaMinMeldSize はメルドの最小枚数。
const LobaMinMeldSize = 3

// LobaMeldKind はメルドの種別。
type LobaMeldKind int

// Loba のメルド種別定数
const (
	// LobaMeldPierna 同ランク・異なるスート
	LobaMeldPierna LobaMeldKind = iota
	// LobaMeldEscalera 同スートの並び
	LobaMeldEscalera
)

// LobaPhase はゲームフェーズ。
type LobaPhase int

// Loba のフェーズ定数
const (
	// LobaPhaseDraw 引く
	LobaPhaseDraw LobaPhase = iota
	// LobaPhaseAct 出す・付ける・捨てる
	LobaPhaseAct
	// LobaPhaseRoundEnd 1 ラウンド終了
	LobaPhaseRoundEnd
	// LobaPhaseGameEnd 決着
	LobaPhaseGameEnd
)

// LobaCardPoints は手札に残った 1 枚の失点を返す。
//
// **ジョーカーも 10 点。**ワイルドではないので使い道が限られ、抱えると重い。
func LobaCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == CardDesignJoker {
		return 10
	}
	switch v := c.GetValue(); {
	case v == 1 || v >= 11:
		return 10 // A, J, Q, K
	default:
		return v
	}
}

// LobaMeld は場に出ているメルド。
type LobaMeld struct {
	// Owner は最初に出した人。レイオフは誰のメルドにもできる。
	Owner int
	Kind  LobaMeldKind
	Cards []*Card
}

// lobaIsJoker は c がジョーカーかを返す。
func lobaIsJoker(c *Card) bool { return c != nil && c.GetDesign() == CardDesignJoker }

// LobaValidateMeld は cards が正しいメルドかを判定し、種別を返す。
//
// ピエルナは**異なる 3 スート**が要る。同じスートを 2 枚含めた 3 枚は、ランクが
// 揃っていてもメルドにならない。エスカレラはジョーカーを 1 枚まで含められる。
func LobaValidateMeld(cards []*Card) (LobaMeldKind, error) {
	if len(cards) < LobaMinMeldSize {
		return 0, fmt.Errorf("a meld needs at least %d cards", LobaMinMeldSize)
	}
	for _, c := range cards {
		if c == nil {
			return 0, fmt.Errorf("a meld cannot contain an empty card")
		}
	}
	if err := lobaValidatePierna(cards); err == nil {
		return LobaMeldPierna, nil
	}
	if err := lobaValidateEscalera(cards); err == nil {
		return LobaMeldEscalera, nil
	}
	return 0, fmt.Errorf("those cards form neither a pierna nor an escalera")
}

// lobaValidatePierna は同ランク・異なるスートかを判定する。
func lobaValidatePierna(cards []*Card) error {
	rank := cards[0].GetValue()
	suits := map[int]bool{}
	for _, c := range cards {
		if lobaIsJoker(c) {
			// **ジョーカーはピエルナに置けない。**ワイルドではないので
			// 「同ランク」を満たしようがない。
			return fmt.Errorf("a joker cannot be part of a pierna")
		}
		if c.GetValue() != rank {
			return fmt.Errorf("a pierna needs one rank")
		}
		suits[c.GetDesign()] = true
	}
	if len(suits) < LobaMinMeldSize {
		return fmt.Errorf("a pierna needs three different suits")
	}
	return nil
}

// lobaValidateEscalera は同スートの並びかを判定する。
//
// A は上か下のどちらか一方。K-A-2 のように跨ぐことはできない。
func lobaValidateEscalera(cards []*Card) error {
	jokers := 0
	suit := -1
	naturals := make([]int, 0, len(cards))
	for _, c := range cards {
		if lobaIsJoker(c) {
			jokers++
			continue
		}
		if suit == -1 {
			suit = c.GetDesign()
		} else if c.GetDesign() != suit {
			return fmt.Errorf("an escalera needs one suit")
		}
		naturals = append(naturals, c.GetValue())
	}
	if jokers > 1 {
		return fmt.Errorf("an escalera may contain at most one joker")
	}
	if len(naturals) == 0 {
		return fmt.Errorf("an escalera needs natural cards")
	}
	// A を下 (1) と上 (14) の両方で試す。跨ぎは片方でしか通らない。
	for _, aceHigh := range []bool{false, true} {
		if lobaSequenceFits(naturals, jokers, aceHigh) {
			return nil
		}
	}
	return fmt.Errorf("those cards are not in sequence")
}

// lobaSequenceFits は naturals が jokers 枚の穴埋めで連続になるかを返す。
func lobaSequenceFits(naturals []int, jokers int, aceHigh bool) bool {
	vals := make([]int, 0, len(naturals))
	for _, v := range naturals {
		if v == 1 && aceHigh {
			v = 14
		}
		vals = append(vals, v)
	}
	sort.Ints(vals)
	for i := 1; i < len(vals); i++ {
		if vals[i] == vals[i-1] {
			return false // 2 組デッキなので同じ札が 2 枚ありうるが、並びには使えない
		}
	}
	gaps := 0
	for i := 1; i < len(vals); i++ {
		gaps += vals[i] - vals[i-1] - 1
	}
	// 穴はジョーカーでちょうど埋まらなければならない。余ったジョーカーは端に
	// 置けるので、余りは許す。
	return gaps <= jokers
}

// Loba はロバのゲームクラス。
type Loba struct {
	players []*LobaPlayer
	config  LobaConfig
	phase   LobaPhase

	stock   []*Card
	discard []*Card
	melds   []*LobaMeld
	// hasMelded[i] はこのラウンドで i が既に何か出したか。レイオフの前提。
	hasMelded []bool
	// meldedBefore[i] は i の**今の手番が始まった時点で**既に出していたか。
	// 「一度も出さずに一気に上がった」(-10) の判定にはこちらを使う。
	meldedBefore []bool

	currentIdx int
	dealerIdx  int
	roundNo    int

	scores []int
	// eliminated[i] は 101 点に達して脱落したか。
	eliminated []bool
	// roundWinner は直近ラウンドで上がった人 (-1: なし)。
	roundWinner int
	// roundClean は直近ラウンドの上がりが「一度も出さずに一気に」だったか。
	roundClean bool

	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewLoba はコンストラクタ。
func NewLoba(players []*LobaPlayer, config LobaConfig) *Loba {
	return &Loba{
		players:      players,
		config:       config,
		scores:       make([]int, len(players)),
		eliminated:   make([]bool, len(players)),
		hasMelded:    make([]bool, len(players)),
		meldedBefore: make([]bool, len(players)),
		roundWinner:  -1,
		winnerIdx:    -1,
	}
}

// NewDefaultLoba は標準の 4 人セットアップを返す。
func NewDefaultLoba() *Loba {
	players := make([]*LobaPlayer, 0, LobaPlayerCnt)
	players = append(players, NewLobaPlayer(true))
	for range LobaPlayerCnt - 1 {
		players = append(players, NewLobaPlayer(false))
	}
	return NewLoba(players, DefaultLobaConfig())
}

// newLobaDeck は 108 枚を生成する (シャッフル前)。
//
// **2 組ぶんなので同じ札が 2 枚ずつある。**単一デッキ前提で「その札は既に出た」と
// 判断すると必ず食い違う。
func newLobaDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, LobaDeckSize)
	for range 2 {
		for _, s := range suits {
			for v := 1; v <= 13; v++ {
				deck = append(deck, NewCard(s, v, true))
			}
		}
	}
	for range LobaJokerCnt {
		deck = append(deck, NewCard(CardDesignJoker, 0, true))
	}
	return deck
}

// lobaShuffle は Fisher-Yates。domain の shuffleCards は casino タグのファイルに
// あり extra2 ビルドから見えないため、専用名で持つ。
func lobaShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// Reset はゲーム全体を初期化する。
func (l *Loba) Reset() {
	l.scores = make([]int, len(l.players))
	l.eliminated = make([]bool, len(l.players))
	l.dealerIdx = 0
	l.roundNo = 0
	l.gameEndFlag = false
	l.winnerIdx = -1
	l.actionLog = nil
	l.dealRound()
}

// dealRound は 1 ラウンドを配る。
func (l *Loba) dealRound() {
	l.melds = nil
	l.hasMelded = make([]bool, len(l.players))
	l.meldedBefore = make([]bool, len(l.players))
	l.roundWinner = -1
	l.roundClean = false
	for _, p := range l.players {
		p.ResetGame()
	}

	deck := newLobaDeck()
	lobaShuffle(deck)
	pos := 0
	for range LobaHandSize {
		for i, p := range l.players {
			if l.eliminated[i] {
				continue
			}
			p.AddCard(deck[pos])
			pos++
		}
	}
	l.discard = []*Card{deck[pos]}
	pos++
	l.stock = append([]*Card(nil), deck[pos:]...)

	l.currentIdx = l.nextActive(l.dealerIdx)
	l.phase = LobaPhaseDraw
	l.addLog(-1, "deal", "cards dealt", nil)
}

// nextActive は idx の次の、脱落していないプレイヤーを返す。
func (l *Loba) nextActive(idx int) int {
	for i := 1; i <= len(l.players); i++ {
		n := (idx + i) % len(l.players)
		if !l.eliminated[n] {
			return n
		}
	}
	return idx
}

// DrawFromStock は山札から 1 枚引く。
func (l *Loba) DrawFromStock(player int) error {
	if err := l.checkDraw(player); err != nil {
		return err
	}
	if len(l.stock) == 0 {
		l.recycleDiscard()
	}
	if len(l.stock) == 0 {
		return fmt.Errorf("there is nothing left to draw")
	}
	card := l.stock[0]
	l.stock = l.stock[1:]
	l.GetPlayer(player).AddCard(card)
	l.beginAct(player)
	l.addLog(player, "draw", "draws from the stock", nil)
	return nil
}

// DrawFromDiscard は捨て札の一番上を取る。
func (l *Loba) DrawFromDiscard(player int) error {
	if err := l.checkDraw(player); err != nil {
		return err
	}
	if len(l.discard) == 0 {
		return fmt.Errorf("the discard pile is empty")
	}
	card := l.discard[len(l.discard)-1]
	l.discard = l.discard[:len(l.discard)-1]
	l.GetPlayer(player).AddCard(card)
	l.beginAct(player)
	l.addLog(player, "draw", "takes the discard", []*Card{card})
	return nil
}

// beginAct は引いた直後に呼ばれ、その手番が始まった時点で player が既に場に
// 出していたかを控える。
//
// **これを控えておかないと「一気に上がった」判定ができない。**Meld した瞬間に
// hasMelded が立つので、上がった後に見ても常に true になってしまう。
func (l *Loba) beginAct(player int) {
	l.phase = LobaPhaseAct
	if player >= 0 && player < len(l.meldedBefore) {
		l.meldedBefore[player] = l.hasMelded[player]
	}
}

// checkDraw は引ける状態かを確かめる。
func (l *Loba) checkDraw(player int) error {
	if l.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if l.phase != LobaPhaseDraw {
		return fmt.Errorf("it is not the draw step")
	}
	if player != l.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	return nil
}

// recycleDiscard は捨て札を裏返して山札に戻す (一番上は残す)。
func (l *Loba) recycleDiscard() {
	if len(l.discard) <= 1 {
		return
	}
	top := l.discard[len(l.discard)-1]
	rest := l.discard[:len(l.discard)-1]
	l.stock = append([]*Card(nil), rest...)
	lobaShuffle(l.stock)
	l.discard = []*Card{top}
	l.addLog(-1, "recycle", "the discard pile is turned back into the stock", nil)
}

// Meld は手札の添字集合をメルドとして場に出す。
func (l *Loba) Meld(player int, handIdxs []int) error {
	if err := l.checkAct(player); err != nil {
		return err
	}
	cards, err := l.takeFromHand(player, handIdxs)
	if err != nil {
		return err
	}
	kind, err := LobaValidateMeld(cards)
	if err != nil {
		return err
	}
	l.removeFromHand(player, handIdxs)
	l.melds = append(l.melds, &LobaMeld{Owner: player, Kind: kind, Cards: cards})
	l.hasMelded[player] = true
	l.addLog(player, "meld", fmt.Sprintf("puts down %d card(s)", len(cards)), cards)
	l.checkGoneOut(player)
	return nil
}

// LayOff は手札 1 枚を既存のメルドに付ける。
//
// **他家のメルドにも付けられる** (de Menos の規則)。ただし自分がこのラウンドで
// 一度も出していないうちは付けられない。
func (l *Loba) LayOff(player, handIdx, meldIdx int) error {
	if err := l.checkAct(player); err != nil {
		return err
	}
	if !l.hasMelded[player] {
		return fmt.Errorf("you must put down a meld of your own first")
	}
	if meldIdx < 0 || meldIdx >= len(l.melds) {
		return fmt.Errorf("no such meld: %d", meldIdx)
	}
	p := l.GetPlayer(player)
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := p.GetCard(handIdx)
	meld := l.melds[meldIdx]
	if !l.layOffFits(meld, card) {
		return fmt.Errorf("that card does not fit the meld")
	}

	p.RemoveCard(handIdx)
	meld.Cards = append(meld.Cards, card)
	l.addLog(player, "layoff", fmt.Sprintf("adds to meld %d", meldIdx), []*Card{card})
	l.checkGoneOut(player)
	return nil
}

// layOffFits は card を meld に付けられるかを返す。
func (l *Loba) layOffFits(meld *LobaMeld, card *Card) bool {
	if meld == nil || card == nil {
		return false
	}
	candidate := make([]*Card, 0, len(meld.Cards)+1)
	candidate = append(candidate, meld.Cards...)
	candidate = append(candidate, card)

	if meld.Kind == LobaMeldPierna {
		// **付け足せるのは最初の 3 スートに限られる。**4 つ目のスートは
		// 「異なる 3 スート」の枠を壊す。
		if lobaIsJoker(card) {
			return false
		}
		if card.GetValue() != meld.Cards[0].GetValue() {
			return false
		}
		allowed := map[int]bool{}
		for _, c := range meld.Cards {
			allowed[c.GetDesign()] = true
		}
		return allowed[card.GetDesign()]
	}
	return lobaValidateEscalera(candidate) == nil
}

// Discard は手札 1 枚を捨てて手番を終える。
func (l *Loba) Discard(player, handIdx int) error {
	if err := l.checkAct(player); err != nil {
		return err
	}
	p := l.GetPlayer(player)
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := p.GetCard(handIdx)
	// **ジョーカーは通常は捨てられない。**手札が 1 枚だけのときだけ例外。
	if lobaIsJoker(card) && p.GetCardsSize() > 1 {
		return fmt.Errorf("a joker cannot be discarded")
	}

	p.RemoveCard(handIdx)
	l.discard = append(l.discard, card)
	l.addLog(player, "discard", "discards", []*Card{card})

	if p.GetCardsSize() == 0 {
		l.finishRound(player)
		return nil
	}
	l.currentIdx = l.nextActive(l.currentIdx)
	l.phase = LobaPhaseDraw
	return nil
}

// checkAct は出す・付ける・捨てるができる状態かを確かめる。
func (l *Loba) checkAct(player int) error {
	if l.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if l.phase != LobaPhaseAct {
		return fmt.Errorf("you must draw first")
	}
	if player != l.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	return nil
}

// takeFromHand は添字集合に対応する札を (取り除かずに) 返す。
func (l *Loba) takeFromHand(player int, idxs []int) ([]*Card, error) {
	p := l.GetPlayer(player)
	if p == nil {
		return nil, fmt.Errorf("no such player: %d", player)
	}
	seen := map[int]bool{}
	cards := make([]*Card, 0, len(idxs))
	for _, i := range idxs {
		if i < 0 || i >= p.GetCardsSize() {
			return nil, fmt.Errorf("card index %d out of range", i)
		}
		if seen[i] {
			return nil, fmt.Errorf("card index %d is listed twice", i)
		}
		seen[i] = true
		cards = append(cards, p.GetCard(i))
	}
	return cards, nil
}

// removeFromHand は添字集合を手札から取り除く。大きい添字から消す。
func (l *Loba) removeFromHand(player int, idxs []int) {
	sorted := append([]int(nil), idxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	p := l.GetPlayer(player)
	for _, i := range sorted {
		p.RemoveCard(i)
	}
}

// checkGoneOut は手札が空になっていればラウンドを締める。
func (l *Loba) checkGoneOut(player int) {
	if l.GetPlayer(player).GetCardsSize() == 0 {
		l.finishRound(player)
	}
}

// finishRound はラウンドを精算する。
func (l *Loba) finishRound(winner int) {
	l.roundWinner = winner
	// **一度も出さずに一気に上がると -10。**
	l.roundClean = l.wentOutInOneGo(winner)
	if l.roundClean {
		l.scores[winner] -= LobaGoOutCleanBonus
		l.addLog(winner, "clean", "goes out in one go", nil)
	}

	for i, p := range l.players {
		if l.eliminated[i] || i == winner {
			continue
		}
		pts := 0
		for j := range p.GetCardsSize() {
			pts += LobaCardPoints(p.GetCard(j))
		}
		l.scores[i] += pts
	}
	l.addLog(winner, "round_end", fmt.Sprintf("wins round %d", l.roundNo+1), nil)

	for i := range l.players {
		if !l.eliminated[i] && l.scores[i] >= LobaKnockOut {
			l.eliminated[i] = true
			l.addLog(i, "eliminated", fmt.Sprintf("reaches %d and is out", l.scores[i]), nil)
		}
	}

	l.roundNo++
	l.phase = LobaPhaseRoundEnd
	l.checkGameEnd()
}

// wentOutInOneGo は winner が上がりの手番で初めて場に出したかを返す。
//
// **メルドの「数」で数えてはいけない。**9 枚を一気に出す形は 3+3+3 が普通で、
// 1 手番で 3 つのメルドになる。数で見ると最も典型的な一気上がりが弾かれ、逆に
// 前の手番で 1 つ出しておいて残りを他人のメルドへレイオフした人が通ってしまう。
// 見るべきは手番が始まった時点で出していたかどうかだけである。
func (l *Loba) wentOutInOneGo(winner int) bool {
	if winner < 0 || winner >= len(l.meldedBefore) {
		return false
	}
	return !l.meldedBefore[winner]
}

// checkGameEnd は残り 1 人になっていれば決着させる。
func (l *Loba) checkGameEnd() {
	alive, last := 0, -1
	for i := range l.players {
		if !l.eliminated[i] {
			alive++
			last = i
		}
	}
	if alive <= 1 {
		l.winnerIdx = last
		l.gameEndFlag = true
		l.phase = LobaPhaseGameEnd
		l.addLog(last, "game_end", "is the last player standing", nil)
	}
}

// NextRound は次のラウンドを配る。
func (l *Loba) NextRound() error {
	if l.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if l.phase != LobaPhaseRoundEnd {
		return fmt.Errorf("the round is still in progress")
	}
	l.dealerIdx = l.nextActive(l.dealerIdx)
	l.dealRound()
	return nil
}

// ---- CPU ----

// LobaCpuAction は CPU が選んだ手。
type LobaCpuAction struct {
	// TakeDiscard が真なら捨て札を取る。
	TakeDiscard bool
	// MeldIdxs は出すメルドの手札添字 (無ければ nil)。
	MeldIdxs []int
	// DiscardIdx は捨てる手札の添字 (-1: 捨てない)。
	DiscardIdx int
}

// LobaCpuDecide は idx の CPU が取る手を決める。
//
// 引く段階では山札を引く。行動段階では、出せるメルドがあれば出し、最も点の高い
// 札を捨てる。**ジョーカーは捨てられない**ので候補から外す。
func (l *Loba) LobaCpuDecide(idx int) LobaCpuAction {
	if l.phase == LobaPhaseDraw {
		return LobaCpuAction{DiscardIdx: -1}
	}
	if meld := l.findMeld(idx); meld != nil {
		return LobaCpuAction{MeldIdxs: meld, DiscardIdx: -1}
	}
	return LobaCpuAction{DiscardIdx: l.pickDiscard(idx)}
}

// findMeld は手札から出せるメルドを 1 つ探す。無ければ nil。
func (l *Loba) findMeld(idx int) []int {
	p := l.GetPlayer(idx)
	if p == nil || p.GetCardsSize() < LobaMinMeldSize {
		return nil
	}
	n := p.GetCardsSize()
	// 3 枚の組み合わせだけを見る。4 枚以上は次の手番で付ければよい。
	for a := range n {
		for b := a + 1; b < n; b++ {
			for c := b + 1; c < n; c++ {
				cards := []*Card{p.GetCard(a), p.GetCard(b), p.GetCard(c)}
				if _, err := LobaValidateMeld(cards); err == nil {
					return []int{a, b, c}
				}
			}
		}
	}
	return nil
}

// pickDiscard は捨てる札を選ぶ。ジョーカーは (手札 1 枚のときを除き) 捨てない。
//
// **単純に「最も点の高い札」を捨てると、手札が一生メルドに育たない。**引いて
// 捨てるだけで枚数が減らないので、誰も上がらずラウンドが終わらなくなる
// (実際に -count=200 で再現した)。組みかけの札を残し、どこにも繋がらない札から
// 捨てる。
func (l *Loba) pickDiscard(idx int) int {
	p := l.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return -1
	}
	best, bestUse, bestPts := -1, 99, -1
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if lobaIsJoker(c) && p.GetCardsSize() > 1 {
			continue
		}
		use := l.discardUsefulness(idx, i)
		pts := LobaCardPoints(c)
		// 使い道が少ない方を優先し、同じなら点の高い方を捨てる。
		if best == -1 || use < bestUse || (use == bestUse && pts > bestPts) {
			best, bestUse, bestPts = i, use, pts
		}
	}
	if best < 0 {
		return 0
	}
	return best
}

// discardUsefulness は手札 i の札が他の札とどれだけ繋がっているかを返す。
//
// ピエルナは**異なるスート**が要るので、同じスートの同ランクは数えない。
// エスカレラは同スートで隣接していれば数える。
func (l *Loba) discardUsefulness(idx, i int) int {
	p := l.GetPlayer(idx)
	c := p.GetCard(i)
	if c == nil || lobaIsJoker(c) {
		return 99
	}
	n := 0
	for j := range p.GetCardsSize() {
		if j == i {
			continue
		}
		o := p.GetCard(j)
		if o == nil || lobaIsJoker(o) {
			continue
		}
		switch {
		case o.GetValue() == c.GetValue() && o.GetDesign() != c.GetDesign():
			n++
		case o.GetDesign() == c.GetDesign() && lobaNear(o.GetValue(), c.GetValue()):
			n++
		}
	}
	return n
}

// lobaNear は 2 つのランクがエスカレラに育ちうる距離かを返す。
// domain には既に別ゲームの abs があるため、専用名で持つ。
func lobaNear(a, b int) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 2
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (l *Loba) GetPlayers() []*LobaPlayer { return l.players }

// GetPlayer は idx のプレイヤーを返す。
func (l *Loba) GetPlayer(idx int) *LobaPlayer {
	return getPlayer(l.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (l *Loba) GetPhase() LobaPhase { return l.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (l *Loba) GetCurrentPlayerIdx() int { return l.currentIdx }

// GetStockCount は山札の残り枚数を返す。
func (l *Loba) GetStockCount() int { return len(l.stock) }

// GetDiscardTop は捨て札の一番上を返す (無ければ nil)。
func (l *Loba) GetDiscardTop() *Card {
	return discardTop(l.discard)
}

// GetMelds は場のメルドを返す。
func (l *Loba) GetMelds() []*LobaMeld { return l.melds }

// HasMelded は idx がこのラウンドで既に出したかを返す。
func (l *Loba) HasMelded(idx int) bool {
	if idx < 0 || idx >= len(l.hasMelded) {
		return false
	}
	return l.hasMelded[idx]
}

// GetScore は idx の累計失点を返す。
func (l *Loba) GetScore(idx int) int {
	return elemAt(l.scores, idx)
}

// IsEliminated は idx が脱落しているかを返す。
func (l *Loba) IsEliminated(idx int) bool {
	if idx < 0 || idx >= len(l.eliminated) {
		return false
	}
	return l.eliminated[idx]
}

// GetRoundNumber は完了したラウンド数を返す。
func (l *Loba) GetRoundNumber() int { return l.roundNo }

// GetRoundWinner は直近ラウンドで上がった人を返す (-1: なし)。
func (l *Loba) GetRoundWinner() int { return l.roundWinner }

// IsRoundClean は直近の上がりが「一気に」だったかを返す。
func (l *Loba) IsRoundClean() bool { return l.roundClean }

// GetGameEndFlag は決着しているかを返す。
func (l *Loba) GetGameEndFlag() bool { return l.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す (-1: 未決着)。
func (l *Loba) GetWinnerIdx() int { return l.winnerIdx }

// GetConfig はゲーム設定を返す。
func (l *Loba) GetConfig() LobaConfig { return l.config }

// SetConfig はゲーム設定をセットする。
func (l *Loba) SetConfig(c LobaConfig) { l.config = c }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (l *Loba) SetPhaseForTest(p LobaPhase) { l.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (l *Loba) SetCurrentPlayerForTest(idx int) { l.currentIdx = idx }

// SetStockForTest はテスト用に山札を差し替える。
func (l *Loba) SetStockForTest(cards []*Card) { l.stock = cards }

// SetDiscardForTest はテスト用に捨て札を差し替える。
func (l *Loba) SetDiscardForTest(cards []*Card) { l.discard = cards }

// SetScoreForTest はテスト用に失点を差し替える。
func (l *Loba) SetScoreForTest(idx, score int) { l.scores[idx] = score }

// SetHasMeldedForTest はテスト用にメルド済みフラグを差し替える。
func (l *Loba) SetHasMeldedForTest(idx int, v bool) { l.hasMelded[idx] = v }

// addLog は棋譜に 1 件追加する。
func (l *Loba) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	l.appendLog(playerIdx, actionType, detail, cards)
}

// lobaJSON is the JSON wire format for Loba.
type lobaJSON struct {
	Players   []*LobaPlayer `json:"pl"`
	Config    LobaConfig    `json:"cfg"`
	Phase     LobaPhase     `json:"ph"`
	Stock     []*Card       `json:"st"`
	Discard   []*Card       `json:"di"`
	Melds     []*LobaMeld   `json:"me"`
	HasMelded []bool        `json:"hm"`
	// MeldedBefore は Worker で手番をまたいでも -10 判定が壊れないように残す。
	MeldedBefore []bool            `json:"mb"`
	Current      int               `json:"cur"`
	Dealer       int               `json:"dl"`
	RoundNo      int               `json:"rn"`
	Scores       []int             `json:"sc"`
	Eliminated   []bool            `json:"el"`
	RoundWinner  int               `json:"rw"`
	RoundClean   bool              `json:"rc"`
	GameEnd      bool              `json:"ge"`
	WinnerIdx    int               `json:"wi"`
	ActionLog    []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (l *Loba) MarshalJSON() ([]byte, error) {
	return json.Marshal(lobaJSON{
		Players: l.players, Config: l.config, Phase: l.phase, Stock: l.stock,
		Discard: l.discard, Melds: l.melds, HasMelded: l.hasMelded,
		MeldedBefore: l.meldedBefore, Current: l.currentIdx,
		Dealer: l.dealerIdx, RoundNo: l.roundNo, Scores: l.scores, Eliminated: l.eliminated,
		RoundWinner: l.roundWinner, RoundClean: l.roundClean, GameEnd: l.gameEndFlag,
		WinnerIdx: l.winnerIdx, ActionLog: l.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を検証する。
// **eliminated と scores が落ちると脱落者が復活する**ので、長さも固定する。
func (l *Loba) UnmarshalJSON(data []byte) error {
	var raw lobaJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != LobaPlayerCnt {
		return fmt.Errorf("expected %d players, got %d", LobaPlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < LobaPhaseDraw || raw.Phase > LobaPhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}

	l.players = raw.Players
	l.config = raw.Config
	l.phase = raw.Phase
	l.stock = raw.Stock
	l.discard = raw.Discard
	l.roundNo = raw.RoundNo
	l.roundClean = raw.RoundClean
	l.gameEndFlag = raw.GameEnd
	l.actionLog = raw.ActionLog

	l.currentIdx = clampLobaIdx(raw.Current, len(l.players))
	l.dealerIdx = clampLobaIdx(raw.Dealer, len(l.players))
	l.roundWinner = raw.RoundWinner
	if l.roundWinner < -1 || l.roundWinner >= len(l.players) {
		l.roundWinner = -1
	}
	l.winnerIdx = raw.WinnerIdx
	if l.winnerIdx < -1 || l.winnerIdx >= len(l.players) {
		l.winnerIdx = -1
	}

	l.scores = make([]int, len(l.players))
	copy(l.scores, raw.Scores)
	l.eliminated = make([]bool, len(l.players))
	copy(l.eliminated, raw.Eliminated)
	l.hasMelded = make([]bool, len(l.players))
	copy(l.hasMelded, raw.HasMelded)
	l.meldedBefore = make([]bool, len(l.players))
	copy(l.meldedBefore, raw.MeldedBefore)

	l.melds = make([]*LobaMeld, 0, len(raw.Melds))
	for _, m := range raw.Melds {
		if m == nil || len(m.Cards) < LobaMinMeldSize {
			continue
		}
		if m.Owner < 0 || m.Owner >= len(l.players) {
			continue
		}
		l.melds = append(l.melds, m)
	}
	return nil
}

// clampLobaIdx は席番号を 0..n-1 に収める。
func clampLobaIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}
