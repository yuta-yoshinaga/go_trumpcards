//go:build !js || !wasm || extra3

// Package domain — ヴィント (Vint) のドメインモデル。
//
// 帝政ロシアの「ロシア式ブリッジ」。52 枚・4 人 2 対 2、各自 13 枚。
// **ダミーを出さない**のがブリッジとの違い。
//
// # issue #4385 の仕様案との相違
//
//   - issue は「**コントラクト達成トリック数**と、7 トリックを超えた分の超過点を
//     集計する」とするが、**守備側も自分のトリックを得点する**。しかも達成/失敗に
//     関係なく、**両チームが取ったトリック全部**を線下に書く。「宣言側だけ」でも
//     「7 を超えた分だけ」でもない
//   - issue はトリック単価に触れていない。**スートとレベルの両方**で決まる
//     (♠4 ♣6 ♦8 ♥10 NT12、レベル +1 ごとに +10)
//   - issue はビッドのスート序列に触れていない。**♠ < ♣ < ♦ < ♥ < NT** で、
//     ブリッジ (♣<♦<♥<♠<NT) とは違い **♠ が最弱**
//   - issue は「オナー保有数に応じた追加ボーナス」とするが、**3 枚以上から**。
//     2 枚以下は 0 点
//   - issue は**エースの得点**に触れていない。オナーとは別勘定で、多く持つ側が
//     1 枚につき単価 × 10 を線上に得る。2 対 2 ならトリックの多い側が総取り
//   - issue は目標点を書いていない。**線下 500 点で 1 ゲーム**、先取 500 /
//     ラバー 1000 のボーナス
//   - issue は未達のペナルティに触れていない。**不足数 × レベル × 500**
//
// 得点表そのものは [VintScore.go] にある。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// VintPlayerCnt はプレイヤー数。
const VintPlayerCnt = 4

// VintTeamCnt はチーム数。
const VintTeamCnt = 2

// VintHandSize は各プレイヤーの手札枚数。
const VintHandSize = 13

// VintTeamOf は席のチームを返す (0/2 が team 0、1/3 が team 1)。
//
// **範囲外は -1。**Go の剰余は負の被除数で負を返すので、declarerIdx が未確定の
// -1 をそのまま渡すと -1 が返る。チーム添字として使う側が弾けるよう、範囲外は
// 明示的に -1 に揃える。
func VintTeamOf(seat int) int {
	if seat < 0 || seat >= VintPlayerCnt {
		return -1
	}
	return seat % VintTeamCnt
}

// VintPhase はゲームフェーズ。
type VintPhase int

// Vint のフェーズ定数
const (
	// VintPhaseBid ビッド
	VintPhaseBid VintPhase = iota
	// VintPhasePlay トリックプレイ
	VintPhasePlay
	// VintPhaseHandEnd 局終了 (精算済み)
	VintPhaseHandEnd
	// VintPhaseGameEnd ラバー終了
	VintPhaseGameEnd
)

// VintBid は 1 件の宣言。
type VintBid struct {
	Player int
	// Level は宣言レベル (0 ならパス)。
	Level int
	// Denom は宣言スート (VintDenom*)。
	Denom int
}

// VintHandResult は 1 局の精算結果。
type VintHandResult struct {
	// TrickPoints は各チームが線下に得た点。
	TrickPoints [VintTeamCnt]int
	// HonourPoints / AcePoints は各チームが線上に得た点。
	HonourPoints [VintTeamCnt]int
	AcePoints    [VintTeamCnt]int
	// Penalty は未達で相手が線上に得た点。
	Penalty [VintTeamCnt]int
	// Made は宣言側が達成したか。
	Made bool
	// DeclarerTricks は宣言側が取ったトリック数。
	DeclarerTricks int
	// TrickValue はこの局のトリック単価。
	TrickValue int
}

// Vint はヴィントのゲームクラス。
type Vint struct {
	players []*VintPlayer
	config  VintConfig
	phase   VintPhase

	dealerIdx  int
	currentIdx int
	bidIdx     int
	bids       []*VintBid
	passCount  int

	highBid *VintBid
	// declarerIdx は落札者 (-1 なら未確定)。
	declarerIdx int
	trumpSuit   int

	trick       []*Card
	trickLeader int
	trickNumber int
	tricksWon   [VintPlayerCnt]int
	// takenCards は各チームが取った札 (オナー/エースの集計に要る)。
	takenCards [VintTeamCnt][]*Card

	// below / above は線下・線上の通算。
	below [VintTeamCnt]int
	above [VintTeamCnt]int
	// gamesWon は取ったゲーム数 (2 でラバー)。
	gamesWon [VintTeamCnt]int
	// lastResult は直前の局の精算。
	lastResult *VintHandResult

	handNumber  int
	gameEndFlag bool
	winnerTeam  int

	actionLogBase
}

// NewVint コンストラクタ
func NewVint(players []*VintPlayer, config VintConfig) *Vint {
	return &Vint{players: players, config: config, winnerTeam: -1, declarerIdx: -1}
}

// NewDefaultVint はデフォルト構成のゲームを返す。
func NewDefaultVint() *Vint {
	players := make([]*VintPlayer, VintPlayerCnt)
	for i := range players {
		players[i] = NewVintPlayer(i == 0)
	}
	return NewVint(players, DefaultVintConfig())
}

// ---- 進行 ----

// Reset ゲーム初期化
func (v *Vint) Reset() {
	v.gameEndFlag = false
	v.winnerTeam = -1
	v.below = [VintTeamCnt]int{}
	v.above = [VintTeamCnt]int{}
	v.gamesWon = [VintTeamCnt]int{}
	v.lastResult = nil
	v.handNumber = 0
	v.dealerIdx = 0
	v.actionLog = nil
	v.beginHand()
}

// beginHand は 1 局を配ってビッドへ入る。
func (v *Vint) beginHand() {
	v.handNumber++
	v.phase = VintPhaseBid
	v.bids = nil
	v.passCount = 0
	v.highBid = nil
	v.declarerIdx = -1
	v.trumpSuit = 0
	v.trick = nil
	v.trickNumber = 0
	v.trickLeader = -1
	v.tricksWon = [VintPlayerCnt]int{}
	v.takenCards = [VintTeamCnt][]*Card{}

	for _, p := range v.players {
		p.ResetRound()
	}

	deck := newVintDeck()
	vintShuffle(deck)
	pos := 0
	for range VintHandSize {
		for i := range VintPlayerCnt {
			v.players[i].AddCard(deck[pos])
			pos++
		}
	}

	v.bidIdx = (v.dealerIdx + 1) % VintPlayerCnt
	v.addLog(-1, "deal", fmt.Sprintf("%d cards each", VintHandSize), nil)
}

// newVintDeck は 52 枚のデッキを返す。
func newVintDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, 52)
	for _, s := range suits {
		for val := 1; val <= 13; val++ {
			deck = append(deck, NewCard(s, val, true))
		}
	}
	return deck
}

// vintShuffle は Fisher-Yates。
func vintShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// checkBidTurn は宣言できる状態かを確かめる。
func (v *Vint) checkBidTurn(player int) error {
	if v.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if v.phase != VintPhaseBid {
		return fmt.Errorf("bidding is not in progress")
	}
	if player != v.bidIdx {
		return fmt.Errorf("it is not player %d's turn to bid", player)
	}
	return nil
}

// Bid は宣言する。
func (v *Vint) Bid(player, level, denom int) error {
	if err := v.checkBidTurn(player); err != nil {
		return err
	}
	if level < VintMinLevel || level > VintMaxLevel {
		return fmt.Errorf("a bid level must be between %d and %d", VintMinLevel, VintMaxLevel)
	}
	if denom < 0 || denom >= VintDenomCount {
		return fmt.Errorf("bad denomination: %d", denom)
	}
	if v.highBid != nil && VintBidRank(denom, level) <= VintBidRank(v.highBid.Denom, v.highBid.Level) {
		return fmt.Errorf("a bid must beat the standing %d", v.highBid.Level)
	}
	rec := &VintBid{Player: player, Level: level, Denom: denom}
	v.bids = append(v.bids, rec)
	v.highBid = rec
	v.passCount = 0
	v.addLog(player, "bid", fmt.Sprintf("bids %d of denomination %d", level, denom), nil)
	v.advanceBid()
	return nil
}

// PassBid は宣言を見送る。
func (v *Vint) PassBid(player int) error {
	if err := v.checkBidTurn(player); err != nil {
		return err
	}
	v.bids = append(v.bids, &VintBid{Player: player, Level: 0})
	v.passCount++
	v.addLog(player, "pass", "passes", nil)
	v.advanceBid()
	return nil
}

// advanceBid は次の宣言手番へ進め、決着していれば契約へ移る。
func (v *Vint) advanceBid() {
	if v.highBid != nil && v.passCount >= VintPlayerCnt-1 {
		v.settleBid()
		return
	}
	if v.highBid == nil && v.passCount >= VintPlayerCnt {
		// **全員パスなら配り直し。**
		v.addLog(-1, "redeal", "everybody passed", nil)
		v.handNumber--
		v.beginHand()
		return
	}
	v.bidIdx = (v.bidIdx + 1) % VintPlayerCnt
}

// settleBid は落札を確定してプレイへ入る。
func (v *Vint) settleBid() {
	v.declarerIdx = v.highBid.Player
	v.trumpSuit = VintDenomToSuit(v.highBid.Denom)
	v.phase = VintPhasePlay
	// **リードはディーラーの左隣。**ダミーが無いので落札者の左からではない。
	v.trickLeader = (v.dealerIdx + 1) % VintPlayerCnt
	v.currentIdx = v.trickLeader
	v.addLog(v.declarerIdx, "contract", fmt.Sprintf("takes %d of denomination %d", v.highBid.Level, v.highBid.Denom), nil)
}

// VintValidPlays は player が出せる手札インデックスを返す。
//
// 追随のみ強制。
func (v *Vint) VintValidPlays(player int) []int {
	p := v.GetPlayer(player)
	if p == nil {
		return nil
	}
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	if len(v.trick) == 0 || v.trick[0] == nil {
		return all
	}
	leadSuit := v.trick[0].GetDesign()
	same := make([]int, 0, len(all))
	for _, i := range all {
		if c := p.GetCard(i); c != nil && c.GetDesign() == leadSuit {
			same = append(same, i)
		}
	}
	if len(same) > 0 {
		return same
	}
	return all
}

// PlayCard は 1 枚出す。
func (v *Vint) PlayCard(player, idx int) error {
	if v.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if v.phase != VintPhasePlay {
		return fmt.Errorf("the play phase is not in progress")
	}
	if player != v.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := v.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return fmt.Errorf("bad card index: %d", idx)
	}
	if !vintContains(v.VintValidPlays(player), idx) {
		return fmt.Errorf("that card may not be played")
	}

	card := p.GetCard(idx)
	p.RemoveCard(idx)
	v.trick = append(v.trick, card)
	v.addLog(player, "play", "plays a card", []*Card{card})

	if len(v.trick) < VintPlayerCnt {
		v.currentIdx = (player + 1) % VintPlayerCnt
		return nil
	}
	v.resolveTrick()
	return nil
}

// vintCardRank は札の強さを返す (A が最強)。
func vintCardRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// resolveTrick はトリックを解決する。
func (v *Vint) resolveTrick() {
	lead := v.trick[0]
	leadSuit := lead.GetDesign()
	bestOffset := 0
	bestIsTrump := v.trumpSuit != 0 && leadSuit == v.trumpSuit
	bestRank := vintCardRank(lead)
	for i := 1; i < len(v.trick); i++ {
		c := v.trick[i]
		if c == nil {
			continue
		}
		isTrump := v.trumpSuit != 0 && c.GetDesign() == v.trumpSuit
		rank := vintCardRank(c)
		switch {
		case isTrump && !bestIsTrump:
			bestOffset, bestIsTrump, bestRank = i, true, rank
		case isTrump == bestIsTrump && c.GetDesign() == v.trick[bestOffset].GetDesign() && rank > bestRank:
			bestOffset, bestRank = i, rank
		}
	}
	winner := (v.trickLeader + bestOffset) % VintPlayerCnt
	team := VintTeamOf(winner)

	v.tricksWon[winner]++
	// **取った札を残す。**オナーとエースの集計に要る。
	v.takenCards[team] = append(v.takenCards[team], v.trick...)

	v.trickNumber++
	v.trick = nil
	v.trickLeader = winner
	v.currentIdx = winner
	v.addLog(winner, "trick", fmt.Sprintf("takes trick %d", v.trickNumber), nil)

	if v.trickNumber >= VintHandSize {
		v.finishHand()
	}
}

// VintTeamTricks はチームが取ったトリック数を返す。
func (v *Vint) VintTeamTricks(team int) int {
	if team < 0 || team >= VintTeamCnt {
		return 0
	}
	total := 0
	for i := range VintPlayerCnt {
		if VintTeamOf(i) == team {
			total += v.tricksWon[i]
		}
	}
	return total
}

// finishHand は局を精算する。
func (v *Vint) finishHand() {
	declTeam := VintTeamOf(v.declarerIdx)
	defTeam := 1 - declTeam
	level := v.highBid.Level
	value := VintTrickValue(v.highBid.Denom, level)

	res := &VintHandResult{TrickValue: value}

	// **両チームが取ったトリック全部を線下に書く。**達成/失敗に関係ない。
	// issue の「宣言側の達成トリック数だけ」は誤り。
	for team := range VintTeamCnt {
		res.TrickPoints[team] = v.VintTeamTricks(team) * value
		v.below[team] += res.TrickPoints[team]
	}

	// **オナーは切札の A K Q J 10 を 3 枚以上から。**ノートランプでは 0。
	for team := range VintTeamCnt {
		count := 0
		for _, c := range v.takenCards[team] {
			if IsVintHonour(c, v.trumpSuit) {
				count++
			}
		}
		res.HonourPoints[team] = VintHonourBonus(count, value)
		v.above[team] += res.HonourPoints[team]
	}

	// **エースは別勘定。**多く持つ側が総取り、2 対 2 はトリックの多い側。
	aces := [VintTeamCnt]int{}
	for team := range VintTeamCnt {
		for _, c := range v.takenCards[team] {
			if IsVintAce(c) {
				aces[team]++
			}
		}
	}
	tieToTeam0 := v.VintTeamTricks(0) >= v.VintTeamTricks(1)
	a0, a1 := VintAceBonus(aces[0], aces[1], value, tieToTeam0)
	res.AcePoints[0], res.AcePoints[1] = a0, a1
	v.above[0] += a0
	v.above[1] += a1

	// **未達なら不足数 × レベル × 500 を相手が線上に得る。**
	// 宣言レベル N は「6 + N トリック」を意味する。
	target := 6 + level
	res.DeclarerTricks = v.VintTeamTricks(declTeam)
	res.Made = res.DeclarerTricks >= target
	if !res.Made {
		short := target - res.DeclarerTricks
		res.Penalty[defTeam] = short * level * VintUndertrickUnit
		v.above[defTeam] += res.Penalty[defTeam]
		v.addLog(v.declarerIdx, "set", fmt.Sprintf("is %d short of %d", short, target), nil)
	} else {
		v.addLog(v.declarerIdx, "made", fmt.Sprintf("makes %d of %d", res.DeclarerTricks, target), nil)
	}

	v.lastResult = res
	v.phase = VintPhaseHandEnd
	v.checkGameWon()
}

// checkGameWon は線下 500 点に届いたチームがあればゲームを締める。
func (v *Vint) checkGameWon() {
	for team := range VintTeamCnt {
		if v.below[team] < VintGameTarget {
			continue
		}
		v.gamesWon[team]++
		// **ゲームを取ったら線下をリセットする。**次のゲームは 0 から。
		v.below = [VintTeamCnt]int{}
		if v.gamesWon[team] >= 2 {
			v.above[team] += VintRubberBonus
			v.winnerTeam = team
			v.gameEndFlag = true
			v.phase = VintPhaseGameEnd
			v.addLog(-1, "rubber", fmt.Sprintf("team %d takes the rubber", team), nil)
			return
		}
		v.above[team] += VintFirstGameBonus
		v.addLog(-1, "game", fmt.Sprintf("team %d takes a game", team), nil)
		return
	}
}

// NextHand は次の局を配る。
func (v *Vint) NextHand() error {
	if v.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if v.phase != VintPhaseHandEnd {
		return fmt.Errorf("the hand is still in progress")
	}
	v.dealerIdx = (v.dealerIdx + 1) % VintPlayerCnt
	v.beginHand()
	return nil
}

// vintContains は s に val が含まれるかを返す。
func vintContains(s []int, val int) bool {
	for _, x := range s {
		if x == val {
			return true
		}
	}
	return false
}

// ---- CPU ----

// VintCpuBid は CPU の宣言を決める。
func (v *Vint) VintCpuBid(idx int) (level, denom int) {
	p := v.GetPlayer(idx)
	if p == nil {
		return 0, 0
	}
	counts := map[int]int{}
	honours := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		counts[c.GetDesign()]++
		if val := c.GetValue(); val == 1 || val >= 11 {
			honours[c.GetDesign()]++
		}
	}
	bestSuit, best := CardDesignSpade, -1
	for s := CardDesignSpade; s <= CardDesignDiamond; s++ {
		if val := counts[s] + honours[s]*2; val > best {
			bestSuit, best = s, val
		}
	}
	// 見込みトリック数から宣言レベルを粗く決める。
	estimate := counts[bestSuit] + honours[bestSuit]
	level = estimate - 6
	if level < VintMinLevel {
		return 0, 0
	}
	if level > VintMaxLevel {
		level = VintMaxLevel
	}
	denom = vintSuitToDenom(bestSuit)
	if v.highBid != nil && VintBidRank(denom, level) <= VintBidRank(v.highBid.Denom, v.highBid.Level) {
		return 0, 0
	}
	return level, denom
}

// vintSuitToDenom は Card の design 値を宣言スートに変換する。
func vintSuitToDenom(suit int) int {
	switch suit {
	case CardDesignSpade:
		return VintDenomSpade
	case CardDesignClover:
		return VintDenomClub
	case CardDesignDiamond:
		return VintDenomDiamond
	case CardDesignHeart:
		return VintDenomHeart
	}
	return VintDenomNoTrump
}

// VintCpuPlay は CPU が出す手札インデックスを返す。
func (v *Vint) VintCpuPlay(idx int) int {
	valid := v.VintValidPlays(idx)
	if len(valid) == 0 {
		return -1
	}
	p := v.GetPlayer(idx)
	if p == nil {
		return valid[0]
	}
	if len(v.trick) == 0 {
		return vintHighest(p, valid)
	}
	lead := v.trick[0]
	winning, winRank := -1, -1
	for _, i := range valid {
		c := p.GetCard(i)
		if vintBeats(c, lead, v.trumpSuit) && vintCardRank(c) > winRank {
			winning, winRank = i, vintCardRank(c)
		}
	}
	if winning >= 0 {
		return winning
	}
	return vintLowest(p, valid)
}

// vintLowest は valid のうち一番弱い札の索引を返す。
func vintLowest(p *VintPlayer, valid []int) int {
	best, bestRank := valid[0], 1<<30
	for _, i := range valid {
		if r := vintCardRank(p.GetCard(i)); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// vintHighest は valid のうち一番強い札の索引を返す。
func vintHighest(p *VintPlayer, valid []int) int {
	best, bestRank := valid[0], -1
	for _, i := range valid {
		if r := vintCardRank(p.GetCard(i)); r > bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// vintBeats は c が lead に勝つかを返す。
func vintBeats(c, lead *Card, trumpSuit int) bool {
	if c == nil || lead == nil {
		return false
	}
	if trumpSuit != 0 && c.GetDesign() == trumpSuit && lead.GetDesign() != trumpSuit {
		return true
	}
	if c.GetDesign() != lead.GetDesign() {
		return false
	}
	return vintCardRank(c) > vintCardRank(lead)
}

// IsHumanTurn は今が人間の手番かを返す。
func (v *Vint) IsHumanTurn() bool {
	if v.gameEndFlag {
		return false
	}
	switch v.phase {
	case VintPhaseBid:
		p := v.GetPlayer(v.bidIdx)
		return p != nil && p.GetIsHuman()
	case VintPhasePlay:
		p := v.GetPlayer(v.currentIdx)
		return p != nil && p.GetIsHuman()
	}
	return false
}

// CpuPlay は今の手番の CPU に 1 手打たせる。
func (v *Vint) CpuPlay() {
	if v.gameEndFlag {
		return
	}
	switch v.phase {
	case VintPhaseBid:
		idx := v.bidIdx
		if p := v.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		level, denom := v.VintCpuBid(idx)
		if level < VintMinLevel || v.Bid(idx, level, denom) != nil {
			_ = v.PassBid(idx)
		}
	case VintPhasePlay:
		idx := v.currentIdx
		if p := v.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		if i := v.VintCpuPlay(idx); i >= 0 {
			_ = v.PlayCard(idx, i)
		}
	}
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (v *Vint) GetPlayers() []*VintPlayer { return v.players }

// GetPlayer は idx のプレイヤーを返す。
func (v *Vint) GetPlayer(idx int) *VintPlayer {
	return getPlayer(v.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (v *Vint) GetPhase() VintPhase { return v.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (v *Vint) GetCurrentPlayerIdx() int { return v.currentIdx }

// GetBidPlayerIdx は宣言中の手番を返す。
func (v *Vint) GetBidPlayerIdx() int { return v.bidIdx }

// GetDealerIdx はディーラーを返す。
func (v *Vint) GetDealerIdx() int { return v.dealerIdx }

// GetBids はこの局の宣言履歴を返す。
func (v *Vint) GetBids() []*VintBid { return v.bids }

// GetHighBid は現在の最高宣言を返す (未落札なら nil)。
func (v *Vint) GetHighBid() *VintBid { return v.highBid }

// GetDeclarerIdx は落札者を返す (-1 なら未確定)。
func (v *Vint) GetDeclarerIdx() int { return v.declarerIdx }

// GetTrumpSuit は切札を返す (0 ならノートランプ)。
func (v *Vint) GetTrumpSuit() int { return v.trumpSuit }

// GetTrick は場に出ている札を返す。
func (v *Vint) GetTrick() []*Card { return v.trick }

// GetTrickLeaderIdx はこのトリックのリード席を返す。
func (v *Vint) GetTrickLeaderIdx() int { return v.trickLeader }

// GetTrickNumber は済んだトリック数を返す。
func (v *Vint) GetTrickNumber() int { return v.trickNumber }

// GetTricksWon は席が取ったトリック数を返す。
func (v *Vint) GetTricksWon(idx int) int {
	if idx < 0 || idx >= VintPlayerCnt {
		return 0
	}
	return v.tricksWon[idx]
}

// GetBelow はチームの線下の点を返す。
func (v *Vint) GetBelow(team int) int {
	if team < 0 || team >= VintTeamCnt {
		return 0
	}
	return v.below[team]
}

// GetAbove はチームの線上の点を返す。
func (v *Vint) GetAbove(team int) int {
	if team < 0 || team >= VintTeamCnt {
		return 0
	}
	return v.above[team]
}

// GetGamesWon はチームが取ったゲーム数を返す。
func (v *Vint) GetGamesWon(team int) int {
	if team < 0 || team >= VintTeamCnt {
		return 0
	}
	return v.gamesWon[team]
}

// GetLastResult は直前の局の精算を返す (まだ無ければ nil)。
func (v *Vint) GetLastResult() *VintHandResult { return v.lastResult }

// GetHandNumber は現在の局番号を返す。
func (v *Vint) GetHandNumber() int { return v.handNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (v *Vint) GetGameEndFlag() bool { return v.gameEndFlag }

// GetWinnerTeam は勝利チームを返す (-1 なら未決)。
func (v *Vint) GetWinnerTeam() int { return v.winnerTeam }

// GetConfig は設定を返す。
func (v *Vint) GetConfig() VintConfig { return v.config }

// SetConfig は設定をセットする。
func (v *Vint) SetConfig(c VintConfig) { v.config = c }

// addLog は棋譜を 1 行足す。
func (v *Vint) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	v.appendLog(playerIdx, actionType, detail, cards)
}

// SetPhaseForTest はテスト用にフェーズを設定する。
func (v *Vint) SetPhaseForTest(p VintPhase) { v.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (v *Vint) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(v.GetPlayer(idx), cards)
}

// SetContractForTest はテスト用に契約を設定する。
func (v *Vint) SetContractForTest(declarer, level, denom int) {
	v.declarerIdx = declarer
	v.highBid = &VintBid{Player: declarer, Level: level, Denom: denom}
	v.trumpSuit = VintDenomToSuit(denom)
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (v *Vint) SetCurrentPlayerForTest(idx int) { v.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (v *Vint) SetTrickLeaderForTest(idx int) { v.trickLeader = idx }

// SetTricksWonForTest はテスト用に取得トリック数を設定する。
func (v *Vint) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < VintPlayerCnt {
		v.tricksWon[idx] = n
	}
}

// SetTakenForTest はテスト用にチームが取った札を設定する。
func (v *Vint) SetTakenForTest(team int, cards []*Card) {
	if team >= 0 && team < VintTeamCnt {
		v.takenCards[team] = cards
	}
}

// SetBelowForTest はテスト用に線下の点を設定する。
func (v *Vint) SetBelowForTest(team, n int) {
	if team >= 0 && team < VintTeamCnt {
		v.below[team] = n
	}
}

// SetGamesWonForTest はテスト用に取ったゲーム数を設定する。
func (v *Vint) SetGamesWonForTest(team, n int) {
	if team >= 0 && team < VintTeamCnt {
		v.gamesWon[team] = n
	}
}

// FinishHandForTest はテスト用に精算を走らせる。
func (v *Vint) FinishHandForTest() { v.finishHand() }

// vintJSON is the JSON wire format for Vint.
type vintJSON struct {
	Players     []*VintPlayer        `json:"pl"`
	Config      VintConfig           `json:"cf"`
	Phase       VintPhase            `json:"ph"`
	DealerIdx   int                  `json:"di"`
	CurrentIdx  int                  `json:"ci"`
	BidIdx      int                  `json:"bi"`
	Bids        []*VintBid           `json:"bd"`
	PassCount   int                  `json:"pc"`
	HighBid     *VintBid             `json:"hb"`
	DeclarerIdx int                  `json:"de"`
	TrumpSuit   int                  `json:"ts"`
	Trick       []*Card              `json:"tk"`
	TrickLeader int                  `json:"tl"`
	TrickNumber int                  `json:"tn"`
	TricksWon   [VintPlayerCnt]int   `json:"tw"`
	TakenCards  [VintTeamCnt][]*Card `json:"tc"`
	Below       [VintTeamCnt]int     `json:"bl"`
	Above       [VintTeamCnt]int     `json:"ab"`
	GamesWon    [VintTeamCnt]int     `json:"gw"`
	LastResult  *VintHandResult      `json:"lr"`
	HandNumber  int                  `json:"hn"`
	GameEndFlag bool                 `json:"ge"`
	WinnerTeam  int                  `json:"wt"`
	ActionLog   []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (v *Vint) MarshalJSON() ([]byte, error) {
	return json.Marshal(vintJSON{
		Players: v.players, Config: v.config, Phase: v.phase,
		DealerIdx: v.dealerIdx, CurrentIdx: v.currentIdx, BidIdx: v.bidIdx,
		Bids: v.bids, PassCount: v.passCount, HighBid: v.highBid,
		DeclarerIdx: v.declarerIdx, TrumpSuit: v.trumpSuit,
		Trick: v.trick, TrickLeader: v.trickLeader, TrickNumber: v.trickNumber,
		TricksWon: v.tricksWon, TakenCards: v.takenCards,
		Below: v.below, Above: v.above, GamesWon: v.gamesWon, LastResult: v.lastResult,
		HandNumber: v.handNumber, GameEndFlag: v.gameEndFlag, WinnerTeam: v.winnerTeam,
		ActionLog: v.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元でしか入らない値を弾く。**KV から戻ってきた壊れた状態でプレイが
// 詰まないよう、席番号・宣言・スートを検証する。
func (v *Vint) UnmarshalJSON(data []byte) error {
	var j vintJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != VintPlayerCnt {
		return fmt.Errorf("bad player count: %d", len(j.Players))
	}
	if j.Phase < VintPhaseBid || j.Phase > VintPhaseGameEnd {
		return fmt.Errorf("bad phase: %d", j.Phase)
	}
	for name, val := range map[string]int{"dealer": j.DealerIdx, "current": j.CurrentIdx, "bid": j.BidIdx} {
		if val < 0 || val >= VintPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, val)
		}
	}
	for name, val := range map[string]int{"declarer": j.DeclarerIdx, "trick leader": j.TrickLeader} {
		if val < -1 || val >= VintPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, val)
		}
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= VintTeamCnt {
		return fmt.Errorf("bad winner team: %d", j.WinnerTeam)
	}
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("bad trump suit: %d", j.TrumpSuit)
	}
	if len(j.Trick) > VintPlayerCnt {
		return fmt.Errorf("bad trick size: %d", len(j.Trick))
	}
	if j.HighBid != nil && j.HighBid.Level != 0 {
		if j.HighBid.Level < VintMinLevel || j.HighBid.Level > VintMaxLevel {
			return fmt.Errorf("bad high bid level: %d", j.HighBid.Level)
		}
		if j.HighBid.Denom < 0 || j.HighBid.Denom >= VintDenomCount {
			return fmt.Errorf("bad high bid denomination: %d", j.HighBid.Denom)
		}
	}

	v.players = j.Players
	v.config = j.Config
	v.phase = j.Phase
	v.dealerIdx = j.DealerIdx
	v.currentIdx = j.CurrentIdx
	v.bidIdx = j.BidIdx
	v.bids = j.Bids
	v.passCount = j.PassCount
	v.highBid = j.HighBid
	v.declarerIdx = j.DeclarerIdx
	v.trumpSuit = j.TrumpSuit
	v.trick = j.Trick
	v.trickLeader = j.TrickLeader
	v.trickNumber = j.TrickNumber
	v.tricksWon = j.TricksWon
	v.takenCards = j.TakenCards
	v.below = j.Below
	v.above = j.Above
	v.gamesWon = j.GamesWon
	v.lastResult = j.LastResult
	v.handNumber = j.HandNumber
	v.gameEndFlag = j.GameEndFlag
	v.winnerTeam = j.WinnerTeam
	v.actionLog = j.ActionLog
	return nil
}
