// Package domain — Shithead (a.k.a. Karma, Palace) shedding card game.
//
// 4 player shedding game where each player holds three layers of cards:
// 3 face-down (blind), 3 face-up (visible), and 3 in hand. Players play
// hand cards first, drawing from stock to keep 3 in hand, then face-up,
// then face-down (blind). First to deplete all 9 cards finishes; the
// last player remaining with cards is the "shithead".
package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ShitheadPlayerCnt シットヘッドのプレイヤー数 (4人固定)
const ShitheadPlayerCnt = 4

// ShitheadInitialFaceDown / FaceUp / Hand: 各層の初期枚数
const (
	ShitheadInitialFaceDown = 3
	ShitheadInitialFaceUp   = 3
	ShitheadInitialHand     = 3
)

// シットヘッドのソース種別 (どこからカードを出すか)
const (
	// ShitheadSourceHand 手札から出す
	ShitheadSourceHand = "hand"
	// ShitheadSourceFaceUp 表向きの場札から出す
	ShitheadSourceFaceUp = "faceup"
	// ShitheadSourceFaceDown 裏向きの場札から出す (ブラインド)
	ShitheadSourceFaceDown = "facedown"
)

// ShitheadCpuAction CPUまたは人間の1ターン分の行動記録
type ShitheadCpuAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	Source      string  // 出した場所 (hand/faceup/facedown)。通常のpickupは ""、facedownのblind失敗時は "facedown"
	PlayedCards []*Card // 出したカード (pickup なら nil)
	Pickup      bool    // 場札を引き取った (出せずにパス) かどうか
	Burned      bool    // 場札焼却が発生したか
	Skipped     bool    // 次のプレイヤーをスキップしたか
}

// shitheadRoundState ゲームごとにリセットされる状態
type shitheadRoundState struct {
	currentTurn int                  // 現在の手番プレイヤーインデックス
	discardPile []*Card              // 場札 (末尾が最新)
	stockPile   []*Card              // 山札 (残り)
	skipNext    bool                 // 8効果: 次プレイヤースキップ
	sevenActive bool                 // 7効果: 次は7以下のカードしか出せない
	gameEndFlag bool                 // ゲーム終了フラグ
	cpuActions  []*ShitheadCpuAction // 人間ターン後のCPU行動履歴
	humanAction *ShitheadCpuAction   // 人間の最後の行動
	nextRank    int                  // 次に上がったプレイヤーに付与するランク
	actionLogBase
	turnNumber int // 内部ターン番号 (棋譜用)
}

// Shithead シットヘッドゲームクラス
type Shithead struct {
	trumpCards *TrumpCards
	players    []*ShitheadPlayer
	config     ShitheadConfig
	round      shitheadRoundState
}

// NewShithead コンストラクタ
func NewShithead(trumpCards *TrumpCards, players []*ShitheadPlayer, config ShitheadConfig) *Shithead {
	return &Shithead{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: shitheadRoundState{
			nextRank: 1,
		},
	}
}

// NewDefaultShithead returns a Shithead with the standard 4-player setup
// (1 human, 3 CPU) and DefaultShitheadConfig.
func NewDefaultShithead() *Shithead {
	config := DefaultShitheadConfig()
	players := make([]*ShitheadPlayer, ShitheadPlayerCnt)
	players[0] = NewShitheadPlayer(true)
	for i := 1; i < ShitheadPlayerCnt; i++ {
		players[i] = NewShitheadPlayer(false)
	}
	return NewShithead(NewTrumpCards(0), players, config)
}

// Reset ゲーム初期化
func (s *Shithead) Reset() {
	s.round = shitheadRoundState{
		discardPile: make([]*Card, 0),
		stockPile:   make([]*Card, 0),
		nextRank:    1,
	}

	for _, p := range s.players {
		p.Reset()
	}

	s.trumpCards.Shuffle()

	// 各プレイヤーに 3 face-down + 3 face-up + 3 hand を配る
	for i := 0; i < ShitheadInitialFaceDown; i++ {
		for _, p := range s.players {
			c := s.trumpCards.DrawCard()
			if c == nil {
				continue
			}
			p.AddFaceDown(c)
		}
	}
	for i := 0; i < ShitheadInitialFaceUp; i++ {
		for _, p := range s.players {
			c := s.trumpCards.DrawCard()
			if c == nil {
				continue
			}
			p.AddFaceUp(c)
		}
	}
	for i := 0; i < ShitheadInitialHand; i++ {
		for _, p := range s.players {
			c := s.trumpCards.DrawCard()
			if c == nil {
				continue
			}
			p.AddCard(c)
		}
	}

	// 残りカードを山札にする
	for {
		c := s.trumpCards.DrawCard()
		if c == nil {
			break
		}
		s.round.stockPile = append(s.round.stockPile, c)
	}

	s.sortAllHands()

	s.round.currentTurn = 0
	s.round.turnNumber = 0
}

// sortAllHands 全プレイヤーの手札をソート
func (s *Shithead) sortAllHands() {
	for _, p := range s.players {
		sort.SliceStable(p.cards, func(i, j int) bool {
			return shitheadRank(p.cards[i]) < shitheadRank(p.cards[j])
		})
	}
}

// shitheadRank カードの強さ (Aは最強=14)
func shitheadRank(c *Card) int {
	if c == nil {
		return 0
	}
	return shitheadRankValue(c.GetValue())
}

// shitheadRankValue converts a raw card value to its effective rank.
// Ace (value 1) is the strongest card and maps to rank 14.
func shitheadRankValue(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// GetPlayerCnt プレイヤー数を返す
func (s *Shithead) GetPlayerCnt() int { return len(s.players) }

// GetPlayer 指定インデックスのプレイヤーを返す
func (s *Shithead) GetPlayer(i int) *ShitheadPlayer {
	return getPlayer(s.players, i)
}

// GetCurrentTurn 現在の手番プレイヤーインデックス
func (s *Shithead) GetCurrentTurn() int { return s.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ
func (s *Shithead) GetGameEndFlag() bool { return s.round.gameEndFlag }

// GetDiscardPile 場札を返す (末尾が最新)
func (s *Shithead) GetDiscardPile() []*Card {
	out := make([]*Card, len(s.round.discardPile))
	copy(out, s.round.discardPile)
	return out
}

// GetTopCard 場札の一番上のカード (なければ nil)
func (s *Shithead) GetTopCard() *Card {
	if len(s.round.discardPile) == 0 {
		return nil
	}
	return s.round.discardPile[len(s.round.discardPile)-1]
}

// GetStockSize 山札の残り枚数
func (s *Shithead) GetStockSize() int { return len(s.round.stockPile) }

// GetCpuActions 人間ターン後のCPU行動履歴
func (s *Shithead) GetCpuActions() []*ShitheadCpuAction { return s.round.cpuActions }

// GetHumanAction 人間の最後の行動
func (s *Shithead) GetHumanAction() *ShitheadCpuAction { return s.round.humanAction }

// GetSkipNext 8効果が有効か
func (s *Shithead) GetSkipNext() bool { return s.round.skipNext }

// GetSevenActive 7効果が有効か
func (s *Shithead) GetSevenActive() bool { return s.round.sevenActive }

// GetConfig 現在の設定を返す
func (s *Shithead) GetConfig() ShitheadConfig { return s.config }

// SetConfig 設定をセット
func (s *Shithead) SetConfig(config ShitheadConfig) { s.config = config }

// GetActionLog 棋譜を返す
func (s *Shithead) GetActionLog() []*ActionLogEntry { return s.round.actionLog }

// IsHumanTurn 現在の手番が人間か
func (s *Shithead) IsHumanTurn() bool {
	if s.round.gameEndFlag || s.round.currentTurn < 0 || s.round.currentTurn >= len(s.players) {
		return false
	}
	return s.players[s.round.currentTurn].GetIsHuman()
}

// CurrentSource 現在のプレイヤーが出すべき場所を返す (hand/faceup/facedown)
func (s *Shithead) CurrentSource() string {
	return s.playerSource(s.round.currentTurn)
}

// playerSource 指定プレイヤーが現在出せる場所を返す
func (s *Shithead) playerSource(idx int) string {
	if idx < 0 || idx >= len(s.players) {
		return ShitheadSourceHand
	}
	p := s.players[idx]
	if p.GetCardsSize() > 0 {
		return ShitheadSourceHand
	}
	if p.GetFaceUpSize() > 0 {
		return ShitheadSourceFaceUp
	}
	return ShitheadSourceFaceDown
}

// PlayerPlay 人間がカードを出す。indices が空なら場札を引き取る (ピックアップ)。
// 裏向きの場札を出す場合は indices に1要素のみ指定。
func (s *Shithead) PlayerPlay(indices []int) error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if !s.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	s.round.cpuActions = nil
	s.round.humanAction = nil

	idx := s.round.currentTurn
	if len(indices) == 0 {
		return s.pickupAndAdvance(idx, true)
	}
	source := s.playerSource(idx)
	return s.playFromSource(idx, indices, source, true)
}

// CpuPlay CPUプレイヤーが1ターン実行する
func (s *Shithead) CpuPlay() {
	if s.round.gameEndFlag {
		return
	}
	idx := s.round.currentTurn
	if idx < 0 || idx >= len(s.players) {
		return
	}
	if s.players[idx].GetIsHuman() {
		return
	}
	source := s.playerSource(idx)
	indices, ok := s.cpuPickPlay(idx, source)
	if !ok {
		_ = s.pickupAndAdvance(idx, false)
		return
	}
	_ = s.playFromSource(idx, indices, source, false)
}

// cpuPickPlay returns indices to play; ok=false means pickup.
func (s *Shithead) cpuPickPlay(idx int, source string) ([]int, bool) {
	p := s.players[idx]
	switch source {
	case ShitheadSourceHand:
		return s.cpuPickFromList(p.cards)
	case ShitheadSourceFaceUp:
		return s.cpuPickFromList(p.faceUp)
	case ShitheadSourceFaceDown:
		// blind: pick first available index
		if p.GetFaceDownSize() == 0 {
			return nil, false
		}
		return []int{0}, true
	}
	return nil, false
}

// cpuPickFromList chooses indices of same-value cards to play. Prefers
// non-magic lowest playable groups; falls back to magic cards if no
// non-magic option exists.
func (s *Shithead) cpuPickFromList(cards []*Card) ([]int, bool) {
	if len(cards) == 0 {
		return nil, false
	}
	// Group indices by value
	groups := make(map[int][]int)
	for i, c := range cards {
		if c == nil {
			continue
		}
		groups[c.GetValue()] = append(groups[c.GetValue()], i)
	}
	// Sort values: prefer non-magic ascending, then magic ascending
	var nonMagic, magic []int
	for v := range groups {
		if s.isMagicValue(v) {
			magic = append(magic, v)
		} else {
			nonMagic = append(nonMagic, v)
		}
	}
	sortByRank := func(vals []int) {
		sort.Slice(vals, func(i, j int) bool {
			return shitheadRankValue(vals[i]) < shitheadRankValue(vals[j])
		})
	}
	sortByRank(nonMagic)
	sortByRank(magic)
	// Hard difficulty: only fall back to magic when no non-magic is playable
	candidates := append(append([]int{}, nonMagic...), magic...)
	if s.config.CpuDifficulty == ShitheadDifficultyEasy {
		// Easy plays the lowest rank regardless of magic status
		all := append([]int{}, nonMagic...)
		all = append(all, magic...)
		sortByRank(all)
		candidates = all
	}
	for _, v := range candidates {
		idxs := groups[v]
		// Try playing all same-value cards together (largest group first)
		// Then try fewer if multi-play is still valid.
		for size := len(idxs); size >= 1; size-- {
			sub := idxs[:size]
			cs := make([]*Card, size)
			for i, ix := range sub {
				cs[i] = cards[ix]
			}
			if s.isPlayable(cs) {
				return sub, true
			}
		}
	}
	return nil, false
}

// isMagicValue reports whether a card value triggers a magic effect under
// the current config.
func (s *Shithead) isMagicValue(v int) bool {
	switch v {
	case 2:
		return s.config.MagicTwo
	case 7:
		return s.config.MagicSeven
	case 8:
		return s.config.MagicEight
	case 10:
		return s.config.MagicTen
	}
	return false
}

// playFromSource validates and plays the chosen cards from the source.
func (s *Shithead) playFromSource(idx int, indices []int, source string, isHuman bool) error {
	indices = dedupSortedInts(indices)
	if len(indices) == 0 {
		return NewDomainError(ErrInvalidPlay, "no card selected")
	}
	p := s.players[idx]

	switch source {
	case ShitheadSourceHand:
		return s.playFromHandOrFaceUp(idx, p, indices, true, isHuman)
	case ShitheadSourceFaceUp:
		return s.playFromHandOrFaceUp(idx, p, indices, false, isHuman)
	case ShitheadSourceFaceDown:
		if len(indices) != 1 {
			return NewDomainError(ErrInvalidPlay, "face-down play requires exactly one index")
		}
		return s.playFromFaceDown(idx, p, indices[0], isHuman)
	}
	return NewDomainError(ErrInvalidPlay, fmt.Sprintf("unknown source %q", source))
}

// playFromHandOrFaceUp handles plays from the hand or face-up stack.
func (s *Shithead) playFromHandOrFaceUp(idx int, p *ShitheadPlayer, indices []int, fromHand bool, isHuman bool) error {
	cards, err := s.collectCardsFromList(p, indices, fromHand)
	if err != nil {
		return err
	}
	if !s.allSameValue(cards) {
		return NewDomainError(ErrInvalidPlay, "all selected cards must be the same value")
	}
	if !s.isPlayable(cards) {
		return NewDomainError(ErrInvalidPlay, "selected cards cannot be played")
	}
	// remove
	if fromHand {
		p.RemoveCards(indices)
	} else {
		removeFromFaceUp(p, indices)
	}
	source := ShitheadSourceHand
	if !fromHand {
		source = ShitheadSourceFaceUp
	}
	s.applyPlay(idx, cards, source, isHuman)
	return nil
}

// playFromFaceDown reveals one face-down card; if playable, plays it.
// Otherwise the player picks up the discard pile and advances.
func (s *Shithead) playFromFaceDown(idx int, p *ShitheadPlayer, fdIdx int, isHuman bool) error {
	c := p.GetFaceDownCard(fdIdx)
	if c == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("face-down index %d out of range", fdIdx))
	}
	cards := []*Card{c}
	if s.isPlayable(cards) {
		p.RemoveFaceDownCard(fdIdx)
		s.applyPlay(idx, cards, ShitheadSourceFaceDown, isHuman)
		return nil
	}
	// Not playable: take it back to hand along with the discard pile.
	p.RemoveFaceDownCard(fdIdx)
	p.AddCard(c)
	s.pickupDiscardForPlayer(idx)
	s.recordAction(idx, &ShitheadCpuAction{
		PlayerIdx: idx,
		Source:    ShitheadSourceFaceDown,
		Pickup:    true,
	}, isHuman)
	s.appendLog(idx, "facedown_pickup",
		fmt.Sprintf("face-down %s was not playable; picked up discard", cuiCardName(c)),
		[]*Card{c})
	s.advanceTurn()
	s.checkGameEnd()
	return nil
}

// collectCardsFromList returns the selected cards from hand or face-up.
func (s *Shithead) collectCardsFromList(p *ShitheadPlayer, indices []int, fromHand bool) ([]*Card, error) {
	out := make([]*Card, len(indices))
	for i, idx := range indices {
		var c *Card
		if fromHand {
			c = p.GetCard(idx)
		} else {
			c = p.GetFaceUpCard(idx)
		}
		if c == nil {
			return nil, NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		out[i] = c
	}
	return out, nil
}

// removeFromFaceUp removes the cards at the given face-up indices. Indices
// must be sorted ascending — RemoveFaceUpCard is called from the back so
// earlier indices stay valid.
func removeFromFaceUp(p *ShitheadPlayer, indices []int) {
	sorted := make([]int, len(indices))
	copy(sorted, indices)
	sort.Ints(sorted)
	for i := len(sorted) - 1; i >= 0; i-- {
		p.RemoveFaceUpCard(sorted[i])
	}
}

// allSameValue reports whether every card in cs has the same value.
func (s *Shithead) allSameValue(cs []*Card) bool {
	if len(cs) == 0 {
		return false
	}
	v := cs[0].GetValue()
	for _, c := range cs[1:] {
		if c == nil || c.GetValue() != v {
			return false
		}
	}
	return true
}

// isPlayable reports whether the given cards (all same value) can be
// played on the current discard pile.
func (s *Shithead) isPlayable(cs []*Card) bool {
	if len(cs) == 0 || cs[0] == nil {
		return false
	}
	if !s.allSameValue(cs) {
		return false
	}
	v := cs[0].GetValue()
	// Magic cards (2 / 10) are always playable
	if (v == 2 && s.config.MagicTwo) || (v == 10 && s.config.MagicTen) {
		return true
	}
	top := s.GetTopCard()
	if top == nil {
		return true
	}
	if s.round.sevenActive {
		// after a 7 only ≤7 may follow (plus 2/10 above)
		return shitheadRank(cs[0]) <= shitheadRank(top)
	}
	return shitheadRank(cs[0]) >= shitheadRank(top)
}

// applyPlay performs the post-play bookkeeping: discard, magic effects,
// hand refill, advance turn, end check.
func (s *Shithead) applyPlay(idx int, cards []*Card, source string, isHuman bool) {
	s.round.discardPile = append(s.round.discardPile, cards...)
	burned := false
	skipped := false

	v := cards[0].GetValue()

	// Reset magic state: 7 effect carries only one turn
	prevSeven := s.round.sevenActive
	s.round.sevenActive = false

	// 7: lock-low for next play
	if v == 7 && s.config.MagicSeven {
		s.round.sevenActive = true
	}
	// 2: clears the seven lock automatically (already cleared above) and lets anything come next
	_ = prevSeven

	// 10: burn pile
	if v == 10 && s.config.MagicTen {
		s.round.discardPile = make([]*Card, 0)
		burned = true
	}
	// Four-of-a-kind in pile: burn
	if !burned && s.config.FourOfAKindBurn && s.fourOfAKindOnTop() {
		s.round.discardPile = make([]*Card, 0)
		burned = true
	}

	// Refill hand from stock if source was hand
	if source == ShitheadSourceHand {
		s.refillHand(s.players[idx])
	}

	// Skip next player
	if v == 8 && s.config.MagicEight {
		s.round.skipNext = true
		skipped = true
	}

	s.recordAction(idx, &ShitheadCpuAction{
		PlayerIdx:   idx,
		Source:      source,
		PlayedCards: append([]*Card{}, cards...),
		Burned:      burned,
		Skipped:     skipped,
	}, isHuman)
	s.appendLog(idx, "play", buildPlayDetail(cards, burned, skipped, source), append([]*Card{}, cards...))

	// Check finish for this player
	s.maybeMarkFinished(idx)

	// Same player goes again on burn (10 or 4-of-a-kind), unless they finished
	if burned && !s.players[idx].GetIsFinished() {
		s.checkGameEnd()
		return
	}
	s.advanceTurn()
	s.checkGameEnd()
}

// fourOfAKindOnTop reports whether the top 4 discard cards have the same value.
func (s *Shithead) fourOfAKindOnTop() bool {
	n := len(s.round.discardPile)
	if n < 4 {
		return false
	}
	v := s.round.discardPile[n-1].GetValue()
	for i := n - 4; i < n-1; i++ {
		if s.round.discardPile[i].GetValue() != v {
			return false
		}
	}
	return true
}

// refillHand draws from stock until the player has 3 hand cards or the
// stock is exhausted. Only refills if the player still has cards somewhere
// (Shithead's "draw to 3" rule only applies while stock remains).
func (s *Shithead) refillHand(p *ShitheadPlayer) {
	for p.GetCardsSize() < ShitheadInitialHand && len(s.round.stockPile) > 0 {
		c := s.round.stockPile[0]
		s.round.stockPile = s.round.stockPile[1:]
		p.AddCard(c)
	}
	sort.SliceStable(p.cards, func(i, j int) bool {
		return shitheadRank(p.cards[i]) < shitheadRank(p.cards[j])
	})
}

// pickupAndAdvance the active player picks up the discard pile and turn advances.
func (s *Shithead) pickupAndAdvance(idx int, isHuman bool) error {
	s.pickupDiscardForPlayer(idx)
	s.round.sevenActive = false
	s.recordAction(idx, &ShitheadCpuAction{
		PlayerIdx: idx,
		Pickup:    true,
	}, isHuman)
	s.appendLog(idx, "pickup", "picked up the discard pile", nil)
	s.advanceTurn()
	s.checkGameEnd()
	return nil
}

// pickupDiscardForPlayer moves the entire discard pile into the player's hand.
func (s *Shithead) pickupDiscardForPlayer(idx int) {
	if idx < 0 || idx >= len(s.players) {
		return
	}
	p := s.players[idx]
	for _, c := range s.round.discardPile {
		p.AddCard(c)
	}
	s.round.discardPile = make([]*Card, 0)
	sort.SliceStable(p.cards, func(i, j int) bool {
		return shitheadRank(p.cards[i]) < shitheadRank(p.cards[j])
	})
}

// maybeMarkFinished marks the player as finished if all 9 cards are gone.
func (s *Shithead) maybeMarkFinished(idx int) {
	p := s.players[idx]
	if p.GetIsFinished() {
		return
	}
	if !p.HasAnyCards() {
		p.SetIsFinished(true)
		p.SetRank(s.round.nextRank)
		s.round.nextRank++
		s.appendLog(idx, "finish", fmt.Sprintf("finished at rank %d", p.GetRank()), nil)
	}
}

// advanceTurn moves to the next active player, applying skip-next.
// Skip skips the next *active* (unfinished) player, not a raw index
// offset. When only two active players remain, the skip effect is a
// no-op so the same player doesn't take two turns in a row, which would
// deadlock the CPU/human game loop after an 8 was just played.
func (s *Shithead) advanceTurn() {
	n := len(s.players)
	cur := s.round.currentTurn
	findNext := func(start int) int {
		for i := 1; i <= n; i++ {
			idx := (start + i) % n
			if idx == cur {
				continue
			}
			if s.players[idx] != nil && !s.players[idx].GetIsFinished() {
				return idx
			}
		}
		return start
	}
	next := findNext(cur)
	if s.round.skipNext {
		s.round.skipNext = false
		if skipped := findNext(next); skipped != cur {
			next = skipped
		}
	}
	s.round.currentTurn = next
	s.round.turnNumber++
}

// checkGameEnd marks the game ended when only one (or zero) active players remain.
func (s *Shithead) checkGameEnd() {
	active := 0
	for _, p := range s.players {
		if !p.GetIsFinished() {
			active++
		}
	}
	if active <= 1 {
		// last unfinished player gets the worst rank (the "shithead")
		for _, p := range s.players {
			if !p.GetIsFinished() {
				p.SetRank(s.round.nextRank)
				p.SetIsFinished(true)
				s.round.nextRank++
			}
		}
		s.round.gameEndFlag = true
	}
}

// recordAction stores the action as humanAction or appends to cpuActions.
func (s *Shithead) recordAction(_ int, action *ShitheadCpuAction, isHuman bool) {
	if isHuman {
		s.round.humanAction = action
	} else {
		s.round.cpuActions = append(s.round.cpuActions, action)
	}
}

// appendLog adds an entry to the action log.
func (s *Shithead) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.round.appendLog(playerIdx, actionType, detail, cards)
}

// buildPlayDetail constructs a human-readable description of a play.
func buildPlayDetail(cards []*Card, burned, skipped bool, source string) string {
	var s string
	if len(cards) == 1 {
		s = fmt.Sprintf("played %s from %s", cuiCardName(cards[0]), source)
	} else {
		s = fmt.Sprintf("played %d %ss from %s", len(cards), valueName(cards[0]), source)
	}
	if burned {
		s += " (burned the pile)"
	}
	if skipped {
		s += " (next player skipped)"
	}
	return s
}

// cuiCardName returns "<suit> <value>" like "SPADE 5". Used only for action
// log detail messages.
func cuiCardName(c *Card) string {
	if c == nil {
		return "??"
	}
	suits := []string{"JOKER", "SPADE", "CLOVER", "HEART", "DIAMOND"}
	d := c.GetDesign()
	if d < 0 || d >= len(suits) {
		d = 0
	}
	return fmt.Sprintf("%s %d", suits[d], c.GetValue())
}

// valueName returns the card value as text for action log.
func valueName(c *Card) string {
	if c == nil {
		return "?"
	}
	switch c.GetValue() {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	}
	return fmt.Sprintf("%d", c.GetValue())
}

// shitheadJSON is the JSON wire format for Shithead.
type shitheadJSON struct {
	TrumpCards *TrumpCards          `json:"tc"`
	Players    []*ShitheadPlayer    `json:"ps"`
	Config     ShitheadConfig       `json:"cf"`
	Round      shitheadRoundStateJS `json:"rd"`
}

type shitheadRoundStateJS struct {
	CurrentTurn int                  `json:"ct"`
	DiscardPile []*Card              `json:"dp"`
	StockPile   []*Card              `json:"sp"`
	SkipNext    bool                 `json:"sn"`
	SevenActive bool                 `json:"sa"`
	GameEndFlag bool                 `json:"ge"`
	CpuActions  []*ShitheadCpuAction `json:"ca"`
	HumanAction *ShitheadCpuAction   `json:"ha"`
	NextRank    int                  `json:"nr"`
	ActionLog   []*ActionLogEntry    `json:"al"`
	TurnNumber  int                  `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (s *Shithead) MarshalJSON() ([]byte, error) {
	return json.Marshal(shitheadJSON{
		TrumpCards: s.trumpCards,
		Players:    s.players,
		Config:     s.config,
		Round: shitheadRoundStateJS{
			CurrentTurn: s.round.currentTurn,
			DiscardPile: s.round.discardPile,
			StockPile:   s.round.stockPile,
			SkipNext:    s.round.skipNext,
			SevenActive: s.round.sevenActive,
			GameEndFlag: s.round.gameEndFlag,
			CpuActions:  s.round.cpuActions,
			HumanAction: s.round.humanAction,
			NextRank:    s.round.nextRank,
			ActionLog:   s.round.actionLog,
			TurnNumber:  s.round.turnNumber,
		},
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Shithead) UnmarshalJSON(data []byte) error {
	var j shitheadJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.players = j.Players
	s.config = j.Config
	s.round = shitheadRoundState{
		currentTurn:   j.Round.CurrentTurn,
		discardPile:   j.Round.DiscardPile,
		stockPile:     j.Round.StockPile,
		skipNext:      j.Round.SkipNext,
		sevenActive:   j.Round.SevenActive,
		gameEndFlag:   j.Round.GameEndFlag,
		cpuActions:    j.Round.CpuActions,
		humanAction:   j.Round.HumanAction,
		nextRank:      j.Round.NextRank,
		actionLogBase: actionLogBase{actionLog: j.Round.ActionLog},
		turnNumber:    j.Round.TurnNumber,
	}
	if s.round.discardPile == nil {
		s.round.discardPile = make([]*Card, 0)
	}
	if s.round.stockPile == nil {
		s.round.stockPile = make([]*Card, 0)
	}
	return nil
}
