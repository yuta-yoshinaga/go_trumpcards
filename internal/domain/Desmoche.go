//go:build !js || !wasm || extra3

// Package domain — デスモチェ (Desmoche) のドメインモデル。
//
// ニカラグアのラミー。52 枚、2〜4 人。配札 9 枚で、**ちょうど 10 枚**
// (配札 9 + 引いた 1) をすべてメルドにできた人がポットを取る。
//
// # issue #4405 の仕様案との相違
//
// **issue の看板メカニクスが 2 つとも存在しない。**
//
//   - issue は「**ポーカーの役ランキング**で役の強さを比較し、最も強い者が
//     ポットを獲得」とするが、**ポーカーの役は一切使わない**。Desmoche は
//     Conquian や gin rummy に近いラミーで、勝敗は「先に出し切ったか」だけ
//   - issue は「9 枚全てを役に組み込めた時点で **desmoche を宣言**」とするが、
//     **desmoche は上がりの宣言ではない**。自分の場のメルドから札を抜いて別の
//     メルドに使い回す手のことで、残ったメルドが有効なままであることが条件
//   - 出すのは 9 枚ではなく**ちょうど 10 枚**。捨てる札が残らない状態になる
//   - 2〜5 人ではなく **2〜4 人**
//   - issue が触れていない: **山札が尽きるまで誰も出し切れなければ「勝者なし」**。
//     ポットは持ち越され、全員が同額を追加する
//   - メルドは進行中に**表向きで場に出す**
//
// したがって issue の「gin rummy のメルド判定 + poker の役ランキング計算」と
// いう方針は、後半が丸ごと不要になる。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// DesmochePlayerCnt はプレイヤー数 (4 人。原典は 2〜4 人)。
const DesmochePlayerCnt = 4

// DesmocheHandSize は配札枚数。
const DesmocheHandSize = 9

// DesmocheGoOutSize は上がりに必要なメルドの総枚数。
//
// **配札 9 枚 + 引いた 1 枚 = 10 枚。**9 枚ではない。捨てる札が残らないので、
// 上がった手番では discard が起きない。
const DesmocheGoOutSize = DesmocheHandSize + 1

// DesmocheMinMeldSize はメルドの最小枚数。
const DesmocheMinMeldSize = 3

// DesmocheAnte は 1 ラウンドあたりの掛け金。勝者なしだと持ち越される。
const DesmocheAnte = 10

// DesmocheMeldKind はメルドの種別。
type DesmocheMeldKind int

// Desmoche のメルド種別定数
const (
	// DesmocheMeldSet 同ランク
	DesmocheMeldSet DesmocheMeldKind = iota
	// DesmocheMeldRun 同スートの並び
	DesmocheMeldRun
)

// DesmochePhase はゲームフェーズ。
type DesmochePhase int

// Desmoche のフェーズ定数
const (
	// DesmochePhaseDraw 引く
	DesmochePhaseDraw DesmochePhase = iota
	// DesmochePhaseAct 出す・組み替える・捨てる
	DesmochePhaseAct
	// DesmochePhaseRoundEnd 1 ラウンド終了
	DesmochePhaseRoundEnd
	// DesmochePhaseGameEnd 決着
	DesmochePhaseGameEnd
)

// DesmocheRounds は決着までのラウンド数。
const DesmocheRounds = 5

// DesmocheMeld は場に出ているメルド。
type DesmocheMeld struct {
	Owner int
	Kind  DesmocheMeldKind
	Cards []*Card
}

// DesmocheValidateMeld は cards が正しいメルドかを判定し、種別を返す。
func DesmocheValidateMeld(cards []*Card) (DesmocheMeldKind, error) {
	if len(cards) < DesmocheMinMeldSize {
		return 0, fmt.Errorf("a meld needs at least %d cards", DesmocheMinMeldSize)
	}
	for _, c := range cards {
		if c == nil {
			return 0, fmt.Errorf("a meld cannot contain an empty card")
		}
	}
	if desmocheIsSet(cards) {
		return DesmocheMeldSet, nil
	}
	if desmocheIsRun(cards) {
		return DesmocheMeldRun, nil
	}
	return 0, fmt.Errorf("those cards form neither a set nor a run")
}

// desmocheIsSet は同ランクかを返す。1 組デッキなので同じ札は 1 枚しかない。
func desmocheIsSet(cards []*Card) bool {
	rank := cards[0].GetValue()
	suits := map[int]bool{}
	for _, c := range cards {
		if c.GetValue() != rank || suits[c.GetDesign()] {
			return false
		}
		suits[c.GetDesign()] = true
	}
	return true
}

// desmocheIsRun は同スートの並びかを返す。A は下 (A-2-3) か上 (Q-K-A) の一方。
func desmocheIsRun(cards []*Card) bool {
	suit := cards[0].GetDesign()
	for _, c := range cards {
		if c.GetDesign() != suit {
			return false
		}
	}
	for _, aceHigh := range []bool{false, true} {
		vals := make([]int, 0, len(cards))
		for _, c := range cards {
			v := c.GetValue()
			if v == 1 && aceHigh {
				v = 14
			}
			vals = append(vals, v)
		}
		sort.Ints(vals)
		ok := true
		for i := 1; i < len(vals); i++ {
			if vals[i] != vals[i-1]+1 {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// newDesmocheDeck は 52 枚を生成する (シャッフル前)。
func newDesmocheDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, 52)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// desmocheShuffle は Fisher-Yates。domain の shuffleCards は casino タグの
// ファイルにあり extra3 ビルドから見えないため、専用名で持つ。
func desmocheShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// Desmoche はデスモチェのゲームクラス。
type Desmoche struct {
	players []*DesmochePlayer
	config  DesmocheConfig
	phase   DesmochePhase

	stock   []*Card
	discard []*Card
	melds   []*DesmocheMeld

	currentIdx int
	dealerIdx  int
	roundNo    int

	// pot は場の掛け金。勝者なしのラウンドで持ち越される。
	pot    int
	scores []int
	// roundWinner は直近ラウンドの勝者 (-1: 勝者なし)。
	roundWinner int
	// roundExhausted は山札が尽きて勝者なしで終わったか。
	roundExhausted bool

	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewDesmoche はコンストラクタ。
func NewDesmoche(players []*DesmochePlayer, config DesmocheConfig) *Desmoche {
	return &Desmoche{
		players:     players,
		config:      config,
		scores:      make([]int, len(players)),
		roundWinner: -1,
		winnerIdx:   -1,
	}
}

// NewDefaultDesmoche は標準の 4 人セットアップを返す。
func NewDefaultDesmoche() *Desmoche {
	players := make([]*DesmochePlayer, 0, DesmochePlayerCnt)
	players = append(players, NewDesmochePlayer(true))
	for range DesmochePlayerCnt - 1 {
		players = append(players, NewDesmochePlayer(false))
	}
	return NewDesmoche(players, DefaultDesmocheConfig())
}

// Reset はゲーム全体を初期化する。
func (d *Desmoche) Reset() {
	d.scores = make([]int, len(d.players))
	d.dealerIdx = 0
	d.roundNo = 0
	d.pot = 0
	d.gameEndFlag = false
	d.winnerIdx = -1
	d.actionLog = nil
	d.dealRound()
}

// dealRound は 1 ラウンドを配る。
//
// **ポットは持ち越す。**前のラウンドが勝者なしで終わっていれば、その分に
// 全員の新しい掛け金が積み上がる。
func (d *Desmoche) dealRound() {
	d.melds = nil
	d.roundWinner = -1
	d.roundExhausted = false
	for _, p := range d.players {
		p.ResetRound()
	}
	d.pot += DesmocheAnte * len(d.players)
	for i := range d.players {
		d.scores[i] -= DesmocheAnte
	}

	deck := newDesmocheDeck()
	desmocheShuffle(deck)
	pos := 0
	for range DesmocheHandSize {
		for _, p := range d.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	d.discard = []*Card{deck[pos]}
	pos++
	d.stock = append([]*Card(nil), deck[pos:]...)

	d.currentIdx = (d.dealerIdx + 1) % len(d.players)
	d.phase = DesmochePhaseDraw
	d.addLog(-1, "deal", fmt.Sprintf("cards dealt, pot is %d", d.pot), nil)
}

// DrawFromStock は山札から 1 枚引く。
func (d *Desmoche) DrawFromStock(player int) error {
	if err := d.checkDraw(player); err != nil {
		return err
	}
	if len(d.stock) == 0 {
		// **山札が尽きたら勝者なし。**ポットは持ち越される。
		d.finishRound(-1)
		return nil
	}
	card := d.stock[0]
	d.stock = d.stock[1:]
	d.GetPlayer(player).AddCard(card)
	d.phase = DesmochePhaseAct
	d.addLog(player, "draw", "draws from the stock", nil)
	return nil
}

// DrawFromDiscard は捨て札の一番上を取る。
func (d *Desmoche) DrawFromDiscard(player int) error {
	if err := d.checkDraw(player); err != nil {
		return err
	}
	if len(d.discard) == 0 {
		return fmt.Errorf("the discard pile is empty")
	}
	card := d.discard[len(d.discard)-1]
	d.discard = d.discard[:len(d.discard)-1]
	d.GetPlayer(player).AddCard(card)
	d.phase = DesmochePhaseAct
	d.addLog(player, "draw", "takes the discard", []*Card{card})
	return nil
}

// checkDraw は引ける状態かを確かめる。
func (d *Desmoche) checkDraw(player int) error {
	if d.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if d.phase != DesmochePhaseDraw {
		return fmt.Errorf("it is not the draw step")
	}
	if player != d.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	return nil
}

// Meld は手札の添字集合をメルドとして場に出す。
func (d *Desmoche) Meld(player int, handIdxs []int) error {
	if err := d.checkAct(player); err != nil {
		return err
	}
	cards, err := d.peekHand(player, handIdxs)
	if err != nil {
		return err
	}
	kind, err := DesmocheValidateMeld(cards)
	if err != nil {
		return err
	}
	d.removeFromHand(player, handIdxs)
	d.melds = append(d.melds, &DesmocheMeld{Owner: player, Kind: kind, Cards: cards})
	d.addLog(player, "meld", fmt.Sprintf("puts down %d card(s)", len(cards)), cards)
	d.checkGoneOut(player)
	return nil
}

// Desmoche は自分の場のメルドから 1 枚を抜き、別のメルドへ移す。
//
// **これが「desmoche」の本来の意味。**上がりの宣言ではなく、場のメルドを
// 組み替える手である。抜いた側のメルドが 3 枚未満になったり、種別として
// 壊れたりする場合は認められない。
func (d *Desmoche) Desmoche(player, fromMeldIdx, cardIdx, toMeldIdx int) error {
	if err := d.checkAct(player); err != nil {
		return err
	}
	if fromMeldIdx == toMeldIdx {
		return fmt.Errorf("a card cannot move to the meld it came from")
	}
	from, err := d.meldAt(fromMeldIdx)
	if err != nil {
		return err
	}
	to, err := d.meldAt(toMeldIdx)
	if err != nil {
		return err
	}
	if from.Owner != player {
		return fmt.Errorf("you may only take from your own melds")
	}
	if cardIdx < 0 || cardIdx >= len(from.Cards) {
		return fmt.Errorf("card index %d out of range", cardIdx)
	}

	card := from.Cards[cardIdx]
	rest := make([]*Card, 0, len(from.Cards)-1)
	rest = append(rest, from.Cards[:cardIdx]...)
	rest = append(rest, from.Cards[cardIdx+1:]...)
	if _, err := DesmocheValidateMeld(rest); err != nil {
		return fmt.Errorf("the meld you took from would no longer be valid: %w", err)
	}
	grown := make([]*Card, 0, len(to.Cards)+1)
	grown = append(grown, to.Cards...)
	grown = append(grown, card)
	kind, err := DesmocheValidateMeld(grown)
	if err != nil {
		return fmt.Errorf("that card does not fit the other meld: %w", err)
	}

	from.Cards = rest
	to.Cards = grown
	to.Kind = kind
	d.addLog(player, "desmoche", fmt.Sprintf("moves a card from meld %d to meld %d", fromMeldIdx, toMeldIdx), []*Card{card})
	return nil
}

// LayOff は手札 1 枚を既存のメルドに付ける。
func (d *Desmoche) LayOff(player, handIdx, meldIdx int) error {
	if err := d.checkAct(player); err != nil {
		return err
	}
	meld, err := d.meldAt(meldIdx)
	if err != nil {
		return err
	}
	p := d.GetPlayer(player)
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := p.GetCard(handIdx)
	grown := make([]*Card, 0, len(meld.Cards)+1)
	grown = append(grown, meld.Cards...)
	grown = append(grown, card)
	kind, err := DesmocheValidateMeld(grown)
	if err != nil {
		return fmt.Errorf("that card does not fit the meld: %w", err)
	}

	p.RemoveCard(handIdx)
	meld.Cards = grown
	meld.Kind = kind
	d.addLog(player, "layoff", fmt.Sprintf("adds to meld %d", meldIdx), []*Card{card})
	d.checkGoneOut(player)
	return nil
}

// Discard は手札 1 枚を捨てて手番を終える。
func (d *Desmoche) Discard(player, handIdx int) error {
	if err := d.checkAct(player); err != nil {
		return err
	}
	p := d.GetPlayer(player)
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := p.RemoveCard(handIdx)
	d.discard = append(d.discard, card)
	d.addLog(player, "discard", "discards", []*Card{card})

	d.currentIdx = (d.currentIdx + 1) % len(d.players)
	d.phase = DesmochePhaseDraw
	return nil
}

// checkAct は出す・組み替える・捨てるができる状態かを確かめる。
func (d *Desmoche) checkAct(player int) error {
	if d.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if d.phase != DesmochePhaseAct {
		return fmt.Errorf("you must draw first")
	}
	if player != d.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	return nil
}

// meldAt は添字のメルドを返す。
func (d *Desmoche) meldAt(idx int) (*DesmocheMeld, error) {
	if idx < 0 || idx >= len(d.melds) {
		return nil, fmt.Errorf("no such meld: %d", idx)
	}
	return d.melds[idx], nil
}

// peekHand は添字集合に対応する札を (取り除かずに) 返す。
func (d *Desmoche) peekHand(player int, idxs []int) ([]*Card, error) {
	p := d.GetPlayer(player)
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
func (d *Desmoche) removeFromHand(player int, idxs []int) {
	sorted := append([]int(nil), idxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	p := d.GetPlayer(player)
	for _, i := range sorted {
		p.RemoveCard(i)
	}
}

// MeldedCount は player が場に出している総枚数を返す。
func (d *Desmoche) MeldedCount(player int) int {
	n := 0
	for _, m := range d.melds {
		if m.Owner == player {
			n += len(m.Cards)
		}
	}
	return n
}

// checkGoneOut は出し切ったかを見てラウンドを締めるか、手番を渡す。
//
// 手札が空でもメルド総数が 10 に届かないことがある (3+3+3 で出し切って捨てた
// 場合)。そのときは**捨てる札が無いので Discard を呼べない**ため、ここで手番を
// 進めておかないと状態機械が詰まる。
func (d *Desmoche) checkGoneOut(player int) {
	if d.GetPlayer(player).GetCardsSize() != 0 {
		return
	}
	if d.MeldedCount(player) >= DesmocheGoOutSize {
		d.finishRound(player)
		return
	}
	d.currentIdx = (d.currentIdx + 1) % len(d.players)
	d.phase = DesmochePhaseDraw
}

// finishRound はラウンドを締める。winner が -1 なら勝者なしでポット持ち越し。
func (d *Desmoche) finishRound(winner int) {
	d.roundWinner = winner
	if winner >= 0 {
		d.scores[winner] += d.pot
		d.addLog(winner, "round_end", fmt.Sprintf("takes the pot of %d", d.pot), nil)
		d.pot = 0
	} else {
		// **勝者なし。**ポットはそのまま次のラウンドへ持ち越す。
		d.roundExhausted = true
		d.addLog(-1, "round_end", fmt.Sprintf("nobody went out; %d carries over", d.pot), nil)
	}

	d.roundNo++
	d.phase = DesmochePhaseRoundEnd
	if d.roundNo >= DesmocheRounds {
		d.finishGame()
	}
}

// finishGame は最終集計する。
func (d *Desmoche) finishGame() {
	best := 0
	for i := 1; i < len(d.scores); i++ {
		if d.scores[i] > d.scores[best] {
			best = i
		}
	}
	d.winnerIdx = best
	d.gameEndFlag = true
	d.phase = DesmochePhaseGameEnd
	d.addLog(best, "game_end", "finishes ahead", nil)
}

// NextRound は次のラウンドを配る。
func (d *Desmoche) NextRound() error {
	if d.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if d.phase != DesmochePhaseRoundEnd {
		return fmt.Errorf("the round is still in progress")
	}
	d.dealerIdx = (d.dealerIdx + 1) % len(d.players)
	d.dealRound()
	return nil
}

// ---- CPU ----

// DesmocheCpuAction は CPU が選んだ手。
type DesmocheCpuAction struct {
	// MeldIdxs は出すメルドの手札添字 (無ければ nil)。
	MeldIdxs []int
	// DiscardIdx は捨てる手札の添字 (-1: 捨てない)。
	DiscardIdx int
}

// DesmocheCpuDecide は idx の CPU が取る手を決める。
func (d *Desmoche) DesmocheCpuDecide(idx int) DesmocheCpuAction {
	if d.phase == DesmochePhaseDraw {
		return DesmocheCpuAction{DiscardIdx: -1}
	}
	if meld := d.findMeld(idx); meld != nil {
		return DesmocheCpuAction{MeldIdxs: meld, DiscardIdx: -1}
	}
	return DesmocheCpuAction{DiscardIdx: d.pickDiscard(idx)}
}

// findMeld は手札から出せるメルドを 1 つ探す。
func (d *Desmoche) findMeld(idx int) []int {
	p := d.GetPlayer(idx)
	if p == nil || p.GetCardsSize() < DesmocheMinMeldSize {
		return nil
	}
	n := p.GetCardsSize()
	for a := range n {
		for b := a + 1; b < n; b++ {
			for c := b + 1; c < n; c++ {
				cards := []*Card{p.GetCard(a), p.GetCard(b), p.GetCard(c)}
				if _, err := DesmocheValidateMeld(cards); err == nil {
					return []int{a, b, c}
				}
			}
		}
	}
	return nil
}

// pickDiscard は捨てる札を選ぶ。組みかけの札を残し、繋がらない札から捨てる。
func (d *Desmoche) pickDiscard(idx int) int {
	p := d.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return -1
	}
	best, bestUse := 0, 99
	for i := range p.GetCardsSize() {
		if use := d.usefulness(idx, i); use < bestUse {
			best, bestUse = i, use
		}
	}
	return best
}

// usefulness は手札 i の札が他の札とどれだけ繋がっているかを返す。
func (d *Desmoche) usefulness(idx, i int) int {
	p := d.GetPlayer(idx)
	c := p.GetCard(i)
	if c == nil {
		return 99
	}
	n := 0
	for j := range p.GetCardsSize() {
		if j == i {
			continue
		}
		o := p.GetCard(j)
		if o == nil {
			continue
		}
		switch {
		case o.GetValue() == c.GetValue():
			n++
		case o.GetDesign() == c.GetDesign() && desmocheNear(o.GetValue(), c.GetValue()):
			n++
		}
	}
	return n
}

// desmocheNear は 2 つのランクが並びに育ちうる距離かを返す。
func desmocheNear(a, b int) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= 2
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (d *Desmoche) GetPlayers() []*DesmochePlayer { return d.players }

// GetPlayer は idx のプレイヤーを返す。
func (d *Desmoche) GetPlayer(idx int) *DesmochePlayer {
	return getPlayer(d.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (d *Desmoche) GetPhase() DesmochePhase { return d.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (d *Desmoche) GetCurrentPlayerIdx() int { return d.currentIdx }

// GetStockCount は山札の残り枚数を返す。
func (d *Desmoche) GetStockCount() int { return len(d.stock) }

// GetDiscardTop は捨て札の一番上を返す (無ければ nil)。
func (d *Desmoche) GetDiscardTop() *Card {
	if len(d.discard) == 0 {
		return nil
	}
	return d.discard[len(d.discard)-1]
}

// GetMelds は場のメルドを返す。
func (d *Desmoche) GetMelds() []*DesmocheMeld { return d.melds }

// GetPot は場の掛け金を返す。
func (d *Desmoche) GetPot() int { return d.pot }

// GetScore は idx の収支を返す。
func (d *Desmoche) GetScore(idx int) int {
	return elemAt(d.scores, idx)
}

// GetRoundNumber は完了したラウンド数を返す。
func (d *Desmoche) GetRoundNumber() int { return d.roundNo }

// GetRoundWinner は直近ラウンドの勝者を返す (-1: 勝者なし)。
func (d *Desmoche) GetRoundWinner() int { return d.roundWinner }

// IsRoundExhausted は山札が尽きて勝者なしで終わったかを返す。
func (d *Desmoche) IsRoundExhausted() bool { return d.roundExhausted }

// GetGameEndFlag は決着しているかを返す。
func (d *Desmoche) GetGameEndFlag() bool { return d.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す (-1: 未決着)。
func (d *Desmoche) GetWinnerIdx() int { return d.winnerIdx }

// GetConfig はゲーム設定を返す。
func (d *Desmoche) GetConfig() DesmocheConfig { return d.config }

// SetConfig はゲーム設定をセットする。
func (d *Desmoche) SetConfig(c DesmocheConfig) { d.config = c }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (d *Desmoche) SetPhaseForTest(p DesmochePhase) { d.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (d *Desmoche) SetCurrentPlayerForTest(idx int) { d.currentIdx = idx }

// SetStockForTest はテスト用に山札を差し替える。
func (d *Desmoche) SetStockForTest(cards []*Card) { d.stock = cards }

// SetDiscardForTest はテスト用に捨て札を差し替える。
func (d *Desmoche) SetDiscardForTest(cards []*Card) { d.discard = cards }

// SetRoundNumberForTest はテスト用にラウンド数を差し替える。
func (d *Desmoche) SetRoundNumberForTest(n int) { d.roundNo = n }

// addLog は棋譜に 1 件追加する。
func (d *Desmoche) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	d.appendLog(playerIdx, actionType, detail, cards)
}

// desmocheJSON is the JSON wire format for Desmoche.
type desmocheJSON struct {
	Players        []*DesmochePlayer `json:"pl"`
	Config         DesmocheConfig    `json:"cfg"`
	Phase          DesmochePhase     `json:"ph"`
	Stock          []*Card           `json:"st"`
	Discard        []*Card           `json:"di"`
	Melds          []*DesmocheMeld   `json:"me"`
	Current        int               `json:"cur"`
	Dealer         int               `json:"dl"`
	RoundNo        int               `json:"rn"`
	Pot            int               `json:"pt"`
	Scores         []int             `json:"sc"`
	RoundWinner    int               `json:"rw"`
	RoundExhausted bool              `json:"re"`
	GameEnd        bool              `json:"ge"`
	WinnerIdx      int               `json:"wi"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (d *Desmoche) MarshalJSON() ([]byte, error) {
	return json.Marshal(desmocheJSON{
		Players: d.players, Config: d.config, Phase: d.phase, Stock: d.stock,
		Discard: d.discard, Melds: d.melds, Current: d.currentIdx, Dealer: d.dealerIdx,
		RoundNo: d.roundNo, Pot: d.pot, Scores: d.scores, RoundWinner: d.roundWinner,
		RoundExhausted: d.roundExhausted, GameEnd: d.gameEndFlag, WinnerIdx: d.winnerIdx,
		ActionLog: d.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を検証する。
// **pot が落ちると持ち越しが消える**ので、そのまま復元する。
func (d *Desmoche) UnmarshalJSON(data []byte) error {
	var raw desmocheJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != DesmochePlayerCnt {
		return fmt.Errorf("expected %d players, got %d", DesmochePlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < DesmochePhaseDraw || raw.Phase > DesmochePhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}

	d.players = raw.Players
	d.config = raw.Config
	d.phase = raw.Phase
	d.stock = raw.Stock
	d.discard = raw.Discard
	d.roundNo = raw.RoundNo
	d.pot = raw.Pot
	d.roundExhausted = raw.RoundExhausted
	d.gameEndFlag = raw.GameEnd
	d.actionLog = raw.ActionLog

	d.currentIdx = clampDesmocheIdx(raw.Current, len(d.players))
	d.dealerIdx = clampDesmocheIdx(raw.Dealer, len(d.players))
	d.roundWinner = raw.RoundWinner
	if d.roundWinner < -1 || d.roundWinner >= len(d.players) {
		d.roundWinner = -1
	}
	d.winnerIdx = raw.WinnerIdx
	if d.winnerIdx < -1 || d.winnerIdx >= len(d.players) {
		d.winnerIdx = -1
	}

	d.scores = make([]int, len(d.players))
	copy(d.scores, raw.Scores)

	d.melds = make([]*DesmocheMeld, 0, len(raw.Melds))
	for _, m := range raw.Melds {
		if m == nil || len(m.Cards) < DesmocheMinMeldSize {
			continue
		}
		if m.Owner < 0 || m.Owner >= len(d.players) {
			continue
		}
		d.melds = append(d.melds, m)
	}
	return nil
}

// clampDesmocheIdx は席番号を 0..n-1 に収める。
func clampDesmocheIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}
