//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SergeantMajorPhase はサージェントメジャーのゲームフェーズ。
type SergeantMajorPhase int

// SergeantMajor のフェーズ定数
const (
	// SergeantMajorPhaseTrump 親が切り札を宣言し、キティを取り込んで捨てる
	SergeantMajorPhaseTrump SergeantMajorPhase = iota
	// SergeantMajorPhaseDiscard 親がキティを取り込んで 4 枚捨てる
	SergeantMajorPhaseDiscard
	// SergeantMajorPhasePlay プレイ中
	SergeantMajorPhasePlay
	// SergeantMajorPhaseRoundEnd ラウンド終了
	SergeantMajorPhaseRoundEnd
	// SergeantMajorPhaseGameEnd ゲーム終了
	SergeantMajorPhaseGameEnd
)

// SergeantMajorPlayerCnt はプレイヤー数（3 人固定）。
const SergeantMajorPlayerCnt = 3

// SergeantMajorTargets は 3 つのノルマ。**席順で決まり、宣言しません。**
// 名前の「8-5-3」がそのままこの 3 つの数です。
var SergeantMajorTargets = [SergeantMajorPlayerCnt]int{8, 5, 3}

// SergeantMajorTricksPerRound は 1 ラウンドのトリック数。
//
// **8 + 5 + 3 = 16 で、トリックが余りも不足もしません。**
const SergeantMajorTricksPerRound = 8 + 5 + 3

// SergeantMajorHandSize は各プレイヤーの手札枚数。
const SergeantMajorHandSize = SergeantMajorTricksPerRound

// SergeantMajorKittySize は親が取り込む余り札の枚数。
//
// **issue の「52 枚を 3 人に均等に配り」は成立しません。** 52 は 3 で
// 割り切れないので、実際は **16 枚ずつ = 48 枚 + 余り 4 枚（キティ）**。
// 親がキティを取り込んで 4 枚捨てることで、全員 16 枚に戻ります。
const SergeantMajorKittySize = 52 - SergeantMajorPlayerCnt*SergeantMajorHandSize

// SergeantMajorDefaultRounds は既定のラウンド数。
const SergeantMajorDefaultRounds = 3

// sergeantMajorMaxSliceLen caps slice sizes during deserialisation.
const sergeantMajorMaxSliceLen = 1000

// SergeantMajor はサージェントメジャー（8-5-3）のゲームクラス。
//
// イギリス軍隊に由来するとされる 3 人専用のトリックテイキング。52 枚を
// 16 枚ずつ配り、**余った 4 枚（キティ）は親が取り込んで 4 枚捨てます**。
//
// **ノルマは席順で決まります。** 親が 8、その左隣が 5、右隣が 3——名前の
// 8-5-3 がそのままこの 3 つの数で、合計 16 はトリック数と一致します。宣言する
// 余地はありません。親はラウンドごとに交代するので、ノルマも回ります。
//
// **多く取れば取っただけ得です。** ノルマとの差がそのまま得点になり、届か
// なかったぶんは次のラウンドで**良い札を召し上げられる**罰に変わります。
type SergeantMajor struct {
	players     []*SergeantMajorPlayer
	config      SergeantMajorConfig
	phase       SergeantMajorPhase
	trumpCards  *TrumpCards
	trumpSuit   int
	roundNumber int
	trickNumber int
	// kitty は親が取り込む余り札（取り込むまで 4 枚、以降 0 枚）。
	kitty []*Card
	// absorbedKitty は取り込んだ 4 枚。捨て札を選ぶあいだだけ保持する。
	//
	// **取り込むと手札に紛れて見分けが付かなくなる** (#5759)。捨てる 4 枚を
	// 選ぶのに「元から持っていた札」と「今入ってきた札」の区別は要るので、
	// 捨て終わるまで覚えておく。手札は取り込み直後に並べ替えられるため、
	// 添字ではなく札そのもので持つ。
	absorbedKitty    []*Card
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int
	// surplus は前ラウンドの過不足（+ が超過、- が不足）。次の配りで札の
	// やり取りに使い、使い切ったら 0 に戻します。
	surplus []int
	// lastExchange は直前に動いた札の枚数（0 = やり取り無し）。
	lastExchange int
	gameEndFlag  bool
	// winnerIdx は勝者 (-1: 未確定/同点)。
	winnerIdx int
	actionLogBase
}

// NewSergeantMajor はコンストラクタ。
func NewSergeantMajor(players []*SergeantMajorPlayer, config SergeantMajorConfig) *SergeantMajor {
	return &SergeantMajor{
		players:   players,
		config:    config,
		surplus:   make([]int, SergeantMajorPlayerCnt),
		winnerIdx: -1,
	}
}

// NewDefaultSergeantMajor は標準の 3 人セットアップを返す。
func NewDefaultSergeantMajor() *SergeantMajor {
	players := make([]*SergeantMajorPlayer, 0, SergeantMajorPlayerCnt)
	players = append(players, NewSergeantMajorPlayer(true))
	for range SergeantMajorPlayerCnt - 1 {
		players = append(players, NewSergeantMajorPlayer(false))
	}
	return NewSergeantMajor(players, DefaultSergeantMajorConfig())
}

// sergeantMajorRank は札の強さ。**A が最強。**
func sergeantMajorRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// Reset はゲームを初期化する。
func (s *SergeantMajor) Reset() {
	s.roundNumber = 0
	s.dealerIdx = 0
	s.surplus = make([]int, SergeantMajorPlayerCnt)
	s.lastExchange = 0
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.actionLog = nil
	for _, p := range s.players {
		p.ResetGame()
	}
	s.startRound()
}

// startRound は 1 ラウンドを配り直す。
func (s *SergeantMajor) startRound() {
	s.phase = SergeantMajorPhaseTrump
	s.trumpSuit = 0
	s.trickNumber = 0
	s.currentTrick = nil
	s.lastExchange = 0
	for _, p := range s.players {
		p.ResetRound()
	}
	s.assignTargets()

	s.trumpCards = NewTrumpCards(0)
	s.trumpCards.Shuffle()
	for range SergeantMajorHandSize {
		for i := range SergeantMajorPlayerCnt {
			idx := (s.dealerIdx + 1 + i) % SergeantMajorPlayerCnt
			if c := s.trumpCards.DrawCard(); c != nil {
				s.players[idx].AddCard(c)
			}
		}
	}
	// **余りは 4 枚。** 親が取り込んで 4 枚捨てます。
	s.kitty = nil
	s.absorbedKitty = nil
	for range SergeantMajorKittySize {
		if c := s.trumpCards.DrawCard(); c != nil {
			s.kitty = append(s.kitty, c)
		}
	}
	s.sortAllHands()

	s.roundNumber++
	s.currentPlayerIdx = s.dealerIdx
	s.leadPlayerIdx = s.dealerIdx
	s.addLog(-1, "deal", fmt.Sprintf("ラウンド %d：16 枚ずつ配り、余り %d 枚は親へ",
		s.roundNumber, SergeantMajorKittySize), nil)
}

// assignTargets はノルマを席へ割り当てる。**親が 8、左隣が 5、右隣が 3。**
func (s *SergeantMajor) assignTargets() {
	for i := range SergeantMajorPlayerCnt {
		idx := (s.dealerIdx + i) % SergeantMajorPlayerCnt
		s.players[idx].SetTarget(SergeantMajorTargets[i])
	}
}

// IsHumanTrumpTurn は人間が切り札を宣言する番かを返す。
func (s *SergeantMajor) IsHumanTrumpTurn() bool {
	return !s.gameEndFlag && s.phase == SergeantMajorPhaseTrump && s.dealerIdx == 0
}

// IsHumanDiscardTurn は人間がキティを捨てる番かを返す。
func (s *SergeantMajor) IsHumanDiscardTurn() bool {
	return !s.gameEndFlag && s.phase == SergeantMajorPhaseDiscard && s.dealerIdx == 0
}

// DeclareTrump は親が切り札を宣言し、キティを取り込む。
func (s *SergeantMajor) DeclareTrump(suit int) error {
	if s.gameEndFlag {
		return errors.New("game is over")
	}
	if s.phase != SergeantMajorPhaseTrump {
		return errors.New("trump has already been declared")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid suit: %d", suit)
	}
	s.trumpSuit = suit
	// **キティは親の手に入る。** 4 枚捨てて 16 枚に戻します。
	s.absorbedKitty = append([]*Card(nil), s.kitty...)
	for _, c := range s.kitty {
		s.players[s.dealerIdx].AddCard(c)
	}
	s.kitty = nil
	s.sortAllHands()
	s.phase = SergeantMajorPhaseDiscard
	s.currentPlayerIdx = s.dealerIdx
	s.addLog(s.dealerIdx, "trump", fmt.Sprintf("切り札は %s、キティを取り込みました", suitStr(suit)), nil)
	return nil
}

// PlayerDeclareTrump は人間（親）が切り札を宣言する。
func (s *SergeantMajor) PlayerDeclareTrump(suit int) error {
	if !s.IsHumanTrumpTurn() {
		return errors.New("not your call")
	}
	return s.DeclareTrump(suit)
}

// CpuDeclareTrump は CPU の親が切り札を宣言する。
func (s *SergeantMajor) CpuDeclareTrump() {
	if s.gameEndFlag || s.phase != SergeantMajorPhaseTrump || s.IsHumanTrumpTurn() {
		return
	}
	_ = s.DeclareTrump(s.chooseCpuTrump(s.dealerIdx))
}

// chooseCpuTrump は CPU の切り札選び。**枚数と強さの合計で決めます。**
func (s *SergeantMajor) chooseCpuTrump(playerIdx int) int {
	p := s.players[playerIdx]
	best, bestScore := CardDesignSpade, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		score := 0
		for i := range p.GetCardsSize() {
			if c := p.GetCard(i); c.GetDesign() == suit {
				score += 2 + sergeantMajorRank(c)/4
			}
		}
		if score > bestScore {
			best, bestScore = suit, score
		}
	}
	return best
}

// PlayerDiscard は人間（親）がキティのぶんを捨てる。
func (s *SergeantMajor) PlayerDiscard(indices []int) error {
	if !s.IsHumanDiscardTurn() {
		return errors.New("not your discard")
	}
	return s.discardBy(0, indices)
}

// CpuDiscard は CPU の親がキティのぶんを捨てる。
func (s *SergeantMajor) CpuDiscard() {
	if s.gameEndFlag || s.phase != SergeantMajorPhaseDiscard || s.IsHumanDiscardTurn() {
		return
	}
	_ = s.discardBy(s.dealerIdx, s.chooseCpuDiscard(s.dealerIdx))
}

// discardBy は親に 4 枚捨てさせ、プレイフェーズへ進む。
func (s *SergeantMajor) discardBy(playerIdx int, indices []int) error {
	if s.phase != SergeantMajorPhaseDiscard {
		return errors.New("not the discard phase")
	}
	if playerIdx != s.dealerIdx {
		return errors.New("only the dealer discards")
	}
	if len(indices) != SergeantMajorKittySize {
		return fmt.Errorf("must discard exactly %d cards", SergeantMajorKittySize)
	}
	p := s.players[playerIdx]
	seen := map[int]bool{}
	for _, i := range indices {
		if i < 0 || i >= p.GetCardsSize() {
			return fmt.Errorf("invalid card index: %d", i)
		}
		if seen[i] {
			return fmt.Errorf("duplicate card index: %d", i)
		}
		seen[i] = true
	}
	// **後ろから消す。** 前から消すと残りのインデックスがずれる。
	sorted := append([]int(nil), indices...)
	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] > sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for _, i := range sorted {
		p.RemoveCard(i)
	}
	// 捨て終われば印は消える。次のラウンドや通常プレイに持ち越さない。
	s.absorbedKitty = nil

	s.exchangeCards()
	s.phase = SergeantMajorPhasePlay
	// **リードは親の左隣。**
	s.leadPlayerIdx = (s.dealerIdx + 1) % SergeantMajorPlayerCnt
	s.currentPlayerIdx = s.leadPlayerIdx
	s.sortAllHands()
	s.addLog(playerIdx, "discard", fmt.Sprintf("%d 枚捨てました", SergeantMajorKittySize), nil)
	return nil
}

// chooseCpuDiscard は CPU が捨てる 4 枚。**切り札でない最弱札から。**
func (s *SergeantMajor) chooseCpuDiscard(playerIdx int) []int {
	p := s.players[playerIdx]
	type scored struct{ idx, rank int }
	all := make([]scored, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		r := sergeantMajorRank(c)
		if c.GetDesign() == s.trumpSuit {
			r += 100 // 切り札は捨てにくくする
		}
		all = append(all, scored{i, r})
	}
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			if all[j].rank < all[i].rank {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	out := make([]int, 0, SergeantMajorKittySize)
	for i := 0; i < SergeantMajorKittySize && i < len(all); i++ {
		out = append(out, all[i].idx)
	}
	return out
}

// exchangeCards は前ラウンドの過不足で札をやり取りする。
//
// **ノルマに届かなかった人が、超えた人へ良い札を差し出します。** 不足ぶんだけ
// 最強札を渡し、代わりに相手の最弱札を受け取ります。
func (s *SergeantMajor) exchangeCards() {
	moved := 0
	for taker := range SergeantMajorPlayerCnt {
		for s.surplus[taker] > 0 {
			giver := s.nextDeficit()
			if giver < 0 {
				break
			}
			// **動かなかったら数えない（レビュー指摘 PR #5311）。** 数えると
			// GetLastExchange が「動いた」と報告するのに 1 枚も動いていない、
			// という表示になる。いまは全員 16 枚固定で到達しないが、
			// 枚数の前提が変わったときに黙って壊れる形にしない。
			if !s.moveBestCard(giver, taker) {
				break
			}
			s.surplus[taker]--
			s.surplus[giver]++
			moved++
		}
	}
	s.surplus = make([]int, SergeantMajorPlayerCnt)
	s.lastExchange = moved
	if moved > 0 {
		s.sortAllHands()
		s.addLog(-1, "exchange", fmt.Sprintf("前ラウンドの過不足で %d 枚を移しました", moved), nil)
	}
}

// nextDeficit はまだ不足ぶんが残っている席を返す (-1 = 無し)。
func (s *SergeantMajor) nextDeficit() int {
	for i := range SergeantMajorPlayerCnt {
		if s.surplus[i] < 0 {
			return i
		}
	}
	return -1
}

// moveBestCard は giver の最強札を taker に渡し、taker の最弱札を返す。
//
// 実際に動かせたかを返す。どちらかの手札が空なら何もせず false。
func (s *SergeantMajor) moveBestCard(giver, taker int) bool {
	gp, tp := s.players[giver], s.players[taker]
	if gp.GetCardsSize() == 0 || tp.GetCardsSize() == 0 {
		return false
	}
	bi, brank := 0, -1
	for i := range gp.GetCardsSize() {
		if r := sergeantMajorRank(gp.GetCard(i)); r > brank {
			bi, brank = i, r
		}
	}
	wi, wrank := 0, 1<<30
	for i := range tp.GetCardsSize() {
		if r := sergeantMajorRank(tp.GetCard(i)); r < wrank {
			wi, wrank = i, r
		}
	}
	best := gp.RemoveCard(bi)
	worst := tp.RemoveCard(wi)
	if best != nil {
		tp.AddCard(best)
	}
	if worst != nil {
		gp.AddCard(worst)
	}
	return true
}

// sortAllHands は手札をスート・ランク順に整える。
func (s *SergeantMajor) sortAllHands() {
	for _, p := range s.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return sergeantMajorRank(ci) < sergeantMajorRank(cj)
		})
	}
}

// --- プレイ -----------------------------------------------------------------

// GetValidPlayIndices は playerIdx が出せる手札の添字を返す。
//
// **フォロー義務あり。**
func (s *SergeantMajor) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= SergeantMajorPlayerCnt || s.gameEndFlag {
		return []int{}
	}
	if s.phase != SergeantMajorPhasePlay {
		return []int{}
	}
	p := s.players[playerIdx]
	leadSuit := 0
	if len(s.currentTrick) > 0 {
		leadSuit = s.currentTrick[0].Card.GetDesign()
	}
	all := make([]int, 0, p.GetCardsSize())
	follow := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
		if leadSuit != 0 && p.GetCard(i).GetDesign() == leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// IsHumanTurn は現在の手番が人間かを返す。
func (s *SergeantMajor) IsHumanTurn() bool {
	if s.gameEndFlag || s.phase != SergeantMajorPhasePlay {
		return false
	}
	return s.players[s.currentPlayerIdx].GetIsHuman()
}

// PlayerPlay は人間が 1 枚出す。
func (s *SergeantMajor) PlayerPlay(cardIndex int) error {
	if !s.IsHumanTurn() {
		return errors.New("not your turn")
	}
	return s.play(0, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (s *SergeantMajor) CpuPlay() {
	if s.gameEndFlag || s.phase != SergeantMajorPhasePlay || s.IsHumanTurn() {
		return
	}
	_ = s.play(s.currentPlayerIdx, s.chooseCpuCard(s.currentPlayerIdx))
}

// play は playerIdx に手札の cardIndex を出させる。
func (s *SergeantMajor) play(playerIdx, cardIndex int) error {
	if s.gameEndFlag {
		return errors.New("game is over")
	}
	if s.phase != SergeantMajorPhasePlay {
		return errors.New("not the play phase")
	}
	if playerIdx != s.currentPlayerIdx {
		return fmt.Errorf("not player %d's turn", playerIdx)
	}
	p := s.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	if !sergeantMajorContains(s.GetValidPlayIndices(playerIdx), cardIndex) {
		return errors.New("must follow the led suit")
	}

	card := p.RemoveCard(cardIndex)
	s.currentTrick = append(s.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	s.addLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(s.currentTrick) < SergeantMajorPlayerCnt {
		s.currentPlayerIdx = (s.currentPlayerIdx + 1) % SergeantMajorPlayerCnt
		return nil
	}
	s.resolveTrick()
	return nil
}

// resolveTrick はトリックを解決する。
func (s *SergeantMajor) resolveTrick() {
	winner := s.trickWinner()
	cards := make([]*Card, 0, SergeantMajorPlayerCnt)
	for _, tc := range s.currentTrick {
		cards = append(cards, tc.Card)
	}
	s.players[winner].AddTrick(cards)
	s.currentTrick = nil
	s.trickNumber++
	s.leadPlayerIdx = winner
	s.currentPlayerIdx = winner
	s.addLog(winner, "trick", fmt.Sprintf("トリック %d を取りました", s.trickNumber), nil)

	if s.trickNumber >= SergeantMajorTricksPerRound {
		s.finishRound()
	}
}

// trickWinner は切り札 > リードスートの順で最強札を出した人を返す。
func (s *SergeantMajor) trickWinner() int {
	if len(s.currentTrick) == 0 {
		return s.leadPlayerIdx
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	best, bestRank, bestTrump := s.currentTrick[0].PlayerIdx, -1, false
	for _, tc := range s.currentTrick {
		suit, rank := tc.Card.GetDesign(), sergeantMajorRank(tc.Card)
		isTrump := s.trumpSuit != 0 && suit == s.trumpSuit
		switch {
		case isTrump && !bestTrump:
			best, bestRank, bestTrump = tc.PlayerIdx, rank, true
		case isTrump == bestTrump && suit == leadSuit && !bestTrump && rank > bestRank:
			best, bestRank = tc.PlayerIdx, rank
		case isTrump && bestTrump && rank > bestRank:
			best, bestRank = tc.PlayerIdx, rank
		}
	}
	return best
}

// finishRound はラウンドを精算する。
//
// **ノルマとの差がそのまま得点。** 多く取れば取っただけ得で、届かなかった
// ぶんは次のラウンドで良い札を召し上げられる罰に変わります。
func (s *SergeantMajor) finishRound() {
	s.phase = SergeantMajorPhaseRoundEnd
	for i, p := range s.players {
		diff := p.GetTrickCount() - p.GetTarget()
		s.surplus[i] = diff
		p.AddScore(diff)
		s.addLog(i, "score", fmt.Sprintf("ノルマ %d に対し %d トリック（%+d）",
			p.GetTarget(), p.GetTrickCount(), diff), nil)
	}
	if s.roundNumber >= s.config.Rounds {
		s.finishGame()
	}
}

// NextRound は次のラウンドを開始する。
func (s *SergeantMajor) NextRound() {
	if s.gameEndFlag || s.phase != SergeantMajorPhaseRoundEnd {
		return
	}
	// **親は交代し、ノルマも一緒に回ります。**
	s.dealerIdx = (s.dealerIdx + 1) % SergeantMajorPlayerCnt
	s.startRound()
}

// finishGame は終局処理。**得点がいちばん高い人の勝ち。**
func (s *SergeantMajor) finishGame() {
	s.phase = SergeantMajorPhaseGameEnd
	s.gameEndFlag = true
	best, bestScore, tie := -1, -1<<30, false
	for i, p := range s.players {
		switch {
		case p.GetScore() > bestScore:
			best, bestScore, tie = i, p.GetScore(), false
		case p.GetScore() == bestScore:
			tie = true
		}
	}
	if tie {
		best = -1
	}
	s.winnerIdx = best
	s.addLog(-1, "result", "ゲーム終了", nil)
}

// GiveUp は投了する。
func (s *SergeantMajor) GiveUp() {
	if s.gameEndFlag {
		return
	}
	s.phase = SergeantMajorPhaseGameEnd
	s.gameEndFlag = true
	s.winnerIdx = -1
	s.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の手。
//
// **多く取れば取っただけ得なので、基本は取りにいきます。**
func (s *SergeantMajor) chooseCpuCard(playerIdx int) int {
	valid := s.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := s.players[playerIdx]
	pick, pickRank := valid[0], sergeantMajorRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if r := sergeantMajorRank(p.GetCard(i)); r > pickRank {
			pick, pickRank = i, r
		}
	}
	return pick
}

// SergeantMajorHint はサージェントメジャーの助言。
type SergeantMajorHint struct {
	CardIndex *int
	Reason    string
	// Suit は切り札に勧めるスート（プレイ中は 0）。
	Suit int
	// Indices は捨てるべき札（捨て札フェーズ以外は空）。
	Indices []int
}

// GetHint は人間への助言を返す。
func (s *SergeantMajor) GetHint() *SergeantMajorHint {
	if s.gameEndFlag {
		return nil
	}
	switch {
	case s.IsHumanTrumpTurn():
		return &SergeantMajorHint{Reason: "sergeantmajorSelectTrump", Suit: s.chooseCpuTrump(0)}
	case s.IsHumanDiscardTurn():
		return &SergeantMajorHint{Reason: "sergeantmajorDiscard", Indices: s.chooseCpuDiscard(0)}
	case s.IsHumanTurn():
		valid := s.GetValidPlayIndices(0)
		if len(valid) == 0 {
			return nil
		}
		idx := s.chooseCpuCard(0)
		reason := "sergeantmajorWinTrick"
		if s.players[0].GetTrickCount() >= s.players[0].GetTarget() {
			reason = "sergeantmajorPressOn"
		}
		return &SergeantMajorHint{CardIndex: &idx, Reason: reason}
	default:
		return nil
	}
}

// sergeantMajorContains は xs が v を含むかを返す。
func sergeantMajorContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// addLog は棋譜に 1 行足す。
func (s *SergeantMajor) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (s *SergeantMajor) GetConfig() SergeantMajorConfig { return s.config }

// SetConfig はゲーム設定を設定する。
func (s *SergeantMajor) SetConfig(cfg SergeantMajorConfig) { s.config = cfg }

// GetPhase は現在のフェーズを返す。
func (s *SergeantMajor) GetPhase() SergeantMajorPhase { return s.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (s *SergeantMajor) GetGameEndFlag() bool { return s.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (s *SergeantMajor) GetRoundNumber() int { return s.roundNumber }

// GetTrickNumber は現在のトリック番号を返す。
func (s *SergeantMajor) GetTrickNumber() int { return s.trickNumber }

// GetTrumpSuit は切り札のスートを返す（0 = 未宣言）。
func (s *SergeantMajor) GetTrumpSuit() int { return s.trumpSuit }

// GetKittySize はキティの枚数を返す（取り込み後は 0）。
func (s *SergeantMajor) GetKittySize() int { return len(s.kitty) }

// IsAbsorbedKittyCard は、その札が今回のキティ由来かを返す。
//
// **札そのもので見分ける。**取り込み直後に手札が並べ替えられるので、添字では
// 追えない。52 枚の中でスートとランクの組は一意なので、これで足りる (#5759)。
func (s *SergeantMajor) IsAbsorbedKittyCard(c *Card) bool {
	if c == nil {
		return false
	}
	for _, k := range s.absorbedKitty {
		if k != nil && k.GetDesign() == c.GetDesign() && k.GetValue() == c.GetValue() {
			return true
		}
	}
	return false
}

// GetDiscardCount は親が捨てる枚数を返す。
func (s *SergeantMajor) GetDiscardCount() int { return SergeantMajorKittySize }

// GetCurrentPlayerIdx は現在の手番を返す。
func (s *SergeantMajor) GetCurrentPlayerIdx() int { return s.currentPlayerIdx }

// GetLeadPlayerIdx はリードプレイヤーを返す。
func (s *SergeantMajor) GetLeadPlayerIdx() int { return s.leadPlayerIdx }

// GetDealerIdx はディーラーを返す。**この席がノルマ 8。**
func (s *SergeantMajor) GetDealerIdx() int { return s.dealerIdx }

// GetCurrentTrick は現在のトリックを返す。
func (s *SergeantMajor) GetCurrentTrick() []*TrickCard { return s.currentTrick }

// GetLastExchange は直前のラウンド間で動いた札の枚数を返す。
func (s *SergeantMajor) GetLastExchange() int { return s.lastExchange }

// GetPlayerCnt はプレイヤー数を返す。
func (s *SergeantMajor) GetPlayerCnt() int { return SergeantMajorPlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (s *SergeantMajor) GetPlayer(i int) *SergeantMajorPlayer {
	if i < 0 || i >= len(s.players) {
		return nil
	}
	return s.players[i]
}

// GetWinnerIdx は勝者を返す（-1 = 未確定/同点）。
func (s *SergeantMajor) GetWinnerIdx() int { return s.winnerIdx }

// GetActionLog は棋譜を返す。
func (s *SergeantMajor) GetActionLog() []*ActionLogEntry { return s.actionLog }

// sergeantMajorJSON は KV スナップショットの表現。
type sergeantMajorJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*SergeantMajorPlayer `json:"pl"`
	Config           SergeantMajorConfig    `json:"cf"`
	Phase            SergeantMajorPhase     `json:"ph"`
	TrumpSuit        int                    `json:"ts"`
	RoundNumber      int                    `json:"rn"`
	TrickNumber      int                    `json:"tn"`
	Kitty            []*Card                `json:"ki"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	CurrentPlayerIdx int                    `json:"ci"`
	LeadPlayerIdx    int                    `json:"li"`
	DealerIdx        int                    `json:"dl"`
	Surplus          []int                  `json:"sp"`
	LastExchange     int                    `json:"le"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerIdx        int                    `json:"wi"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (s *SergeantMajor) MarshalJSON() ([]byte, error) {
	return json.Marshal(&sergeantMajorJSON{
		TrumpCards: s.trumpCards, Players: s.players, Config: s.config, Phase: s.phase,
		TrumpSuit: s.trumpSuit, RoundNumber: s.roundNumber, TrickNumber: s.trickNumber,
		Kitty: s.kitty, CurrentTrick: s.currentTrick, CurrentPlayerIdx: s.currentPlayerIdx,
		LeadPlayerIdx: s.leadPlayerIdx, DealerIdx: s.dealerIdx, Surplus: s.surplus,
		LastExchange: s.lastExchange, GameEndFlag: s.gameEndFlag, WinnerIdx: s.winnerIdx,
		ActionLog: s.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
func (s *SergeantMajor) UnmarshalJSON(data []byte) error {
	var j sergeantMajorJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < SergeantMajorPhaseTrump || j.Phase > SergeantMajorPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **切り札は宣言フェーズのあいだだけ 0。** 素通しすると trickWinner が
	// どの札も切り札と見なさなくなり、勝敗が黙って変わる (#5302〜#5305)。
	if j.Phase == SergeantMajorPhaseTrump {
		if j.TrumpSuit != 0 {
			return fmt.Errorf("trump suit %d before it was declared", j.TrumpSuit)
		}
	} else if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
	}
	// **キティは宣言前だけ 4 枚。** 親が取り込んだら消える。
	wantKitty := 0
	if j.Phase == SergeantMajorPhaseTrump {
		wantKitty = SergeantMajorKittySize
	}
	if len(j.Kitty) != wantKitty {
		return fmt.Errorf("kitty holds %d cards in phase %d", len(j.Kitty), j.Phase)
	}
	if j.RoundNumber < 1 || j.RoundNumber > SergeantMajorRoundsMax {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > SergeantMajorTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if len(j.CurrentTrick) > SergeantMajorPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	// **枚数だけでなく中身も見る（#5310 で踏んだ panic の再発防止）。**
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= SergeantMajorPlayerCnt {
			return errors.New("invalid current trick entry")
		}
	}
	if len(j.ActionLog) > sergeantMajorMaxSliceLen {
		return errors.New("sergeantmajor: input array exceeds maximum allowed size")
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= SergeantMajorPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= SergeantMajorPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if !j.GameEndFlag && j.WinnerIdx != -1 {
		return fmt.Errorf("winner %d before the game ended", j.WinnerIdx)
	}
	// **過不足は合計 0。** 誰かの超過は必ず誰かの不足。
	if len(j.Surplus) != SergeantMajorPlayerCnt {
		return fmt.Errorf("surplus holds %d entries", len(j.Surplus))
	}
	sum := 0
	for _, v := range j.Surplus {
		if v < -SergeantMajorTricksPerRound || v > SergeantMajorTricksPerRound {
			return fmt.Errorf("invalid surplus: %d", v)
		}
		sum += v
	}
	if sum != 0 {
		return fmt.Errorf("surplus sums to %d, not 0", sum)
	}
	if j.LastExchange < 0 || j.LastExchange > SergeantMajorTricksPerRound {
		return fmt.Errorf("invalid exchange size: %d", j.LastExchange)
	}

	if j.TrumpCards != nil {
		s.trumpCards = j.TrumpCards
	}
	if len(j.Players) == SergeantMajorPlayerCnt {
		s.players = j.Players
	}
	s.config, s.phase, s.trumpSuit = j.Config, j.Phase, j.TrumpSuit
	s.roundNumber, s.trickNumber, s.kitty = j.RoundNumber, j.TrickNumber, j.Kitty
	s.currentTrick, s.currentPlayerIdx = j.CurrentTrick, j.CurrentPlayerIdx
	s.leadPlayerIdx, s.dealerIdx, s.surplus = j.LeadPlayerIdx, j.DealerIdx, j.Surplus
	s.lastExchange, s.gameEndFlag, s.winnerIdx = j.LastExchange, j.GameEndFlag, j.WinnerIdx
	s.actionLog = j.ActionLog
	return nil
}
