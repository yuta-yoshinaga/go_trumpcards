//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Bourréのゲーム定数
const (
	// BourrePlayerCnt プレイヤー数 (人間1 + CPU4)
	BourrePlayerCnt = 5
	// BourreHandSize 各プレイヤーの手札枚数 / トリック数
	BourreHandSize = 5
	// BourreInitChips 各プレイヤーの初期チップ
	BourreInitChips = 100
	// BourreAnte 毎ハンドのアンティ
	BourreAnte = 5
)

// BourrePhase ゲームフェーズ
type BourrePhase int

// Bourréのフェーズ定数
const (
	// BourrePhaseDecide 参加/フォールド決定フェーズ
	BourrePhaseDecide BourrePhase = 0
	// BourrePhaseDraw 手札交換 (ドロー) フェーズ
	BourrePhaseDraw BourrePhase = 1
	// BourrePhasePlay トリックプレイフェーズ
	BourrePhasePlay BourrePhase = 2
	// BourrePhaseRoundEnd ハンド終了フェーズ (得点・ポット精算済み)
	BourrePhaseRoundEnd BourrePhase = 3
	// BourrePhaseGameEnd ゲーム終了フェーズ
	BourrePhaseGameEnd BourrePhase = 4
)

// BourreHandResult ハンド結果 (表示用)
type BourreHandResult struct {
	PlayerIdx int  `json:"pi"`
	Tricks    int  `json:"tk"`
	WonAmount int  `json:"wa"`
	Bourreed  bool `json:"bo"`
	Folded    bool `json:"fd"`
}

// Bourre ブーレゲームクラス
type Bourre struct {
	trumpCards       *TrumpCards
	players          []*BourrePlayer
	config           BourreConfig
	phase            BourrePhase
	pot              int // 現在のハンドのポット
	carryPot         int // 次ハンドへ繰り越すポット (ブーレ罰金 + 引き分け繰り越し)
	trumpSuit        int // 切り札スート (Card design)
	trumpCard        *Card
	dealerIdx        int
	currentPlayerIdx int
	trickNumber      int
	currentTrick     []*TrickCard
	lastTrick        []*TrickCard
	lastTrickWinner  int
	leadPlayerIdx    int
	handNumber       int
	gameEndFlag      bool
	winnerIdx        int
	lastResults      []*BourreHandResult
	actionLog        []*ActionLogEntry
}

// NewBourre コンストラクタ
func NewBourre(trumpCards *TrumpCards, players []*BourrePlayer, config BourreConfig) *Bourre {
	return &Bourre{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerIdx:       -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultBourre returns Bourre with the standard 5-player setup (1 human, 4 CPU)
// and DefaultBourreConfig. Used as the single source of truth for CUI, Web, and
// Worker construction sites.
func NewDefaultBourre() *Bourre {
	players := make([]*BourrePlayer, BourrePlayerCnt)
	players[0] = NewBourrePlayer(true)
	for i := 1; i < BourrePlayerCnt; i++ {
		players[i] = NewBourrePlayer(false)
	}
	return NewBourre(NewTrumpCards(0), players, DefaultBourreConfig())
}

// Reset ゲーム全体を初期化し、最初のハンドを開始する
func (b *Bourre) Reset() {
	b.gameEndFlag = false
	b.winnerIdx = -1
	b.handNumber = 0
	b.carryPot = 0
	b.actionLog = nil
	b.dealerIdx = 0
	b.lastResults = nil

	for _, p := range b.players {
		p.SetChips(BourreInitChips)
		p.SetIsFinished(false)
		p.ResetHand()
	}

	b.startHand()
}

// startHand 1ハンド分の準備 (アンティ・配札・切り札決定) を行う
func (b *Bourre) startHand() {
	n := len(b.players)
	if n == 0 {
		return
	}
	b.handNumber++
	b.trickNumber = 0
	b.currentTrick = nil
	b.lastTrick = nil
	b.lastTrickWinner = -1
	b.leadPlayerIdx = -1
	b.lastResults = nil
	b.trumpCard = nil

	for _, p := range b.players {
		p.ResetHand()
		p.SetIsFinished(p.GetChips() <= 0)
	}

	// ポット繰り越し + アンティ
	b.pot = b.carryPot
	b.carryPot = 0
	for i := 0; i < n; i++ {
		p := b.players[i]
		if p.GetIsFinished() {
			continue
		}
		ante := min(BourreAnte, p.GetChips())
		p.SubtractChips(ante)
		b.pot += ante
		b.appendLog(i, "ante", fmt.Sprintf("%s antes %d", b.playerName(i), ante), nil)
	}

	// 配札 (ディーラーの左から、ディーラーが最後)
	b.trumpCards.Shuffle()
	order := b.participantsFrom((b.dealerIdx + 1) % n)
	var lastDealt *Card
	for r := 0; r < BourreHandSize; r++ {
		for _, idx := range order {
			c := b.trumpCards.DrawCard()
			if c != nil {
				b.players[idx].AddCard(c)
				lastDealt = c
			}
		}
	}
	// ディーラーの最後の1枚を表向きにして切り札スートを決定
	if lastDealt != nil {
		b.trumpCard = lastDealt
		b.trumpSuit = lastDealt.GetDesign()
	}

	for _, idx := range order {
		b.sortHand(b.players[idx])
	}

	b.appendLog(-1, "trump", fmt.Sprintf("Trump suit: %s", suitNameOf(b.trumpSuit)), trumpCardSlice(b.trumpCard))

	b.phase = BourrePhaseDecide
	if len(order) > 0 {
		b.currentPlayerIdx = order[0]
	}
}

// NextHand ハンド終了後に次のハンドを開始する
func (b *Bourre) NextHand() {
	if b.gameEndFlag || b.phase != BourrePhaseRoundEnd {
		return
	}
	b.startHand()
}

// --- 参加/フォールド決定フェーズ ---

// PlayerDecide 人間プレイヤーが参加(true)かフォールド(false)を決める
func (b *Bourre) PlayerDecide(play bool) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BourrePhaseDecide {
		return ErrWrongPhase
	}
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= len(b.players) {
		return ErrNotHumanTurn
	}
	p := b.players[idx]
	if p == nil || !p.GetIsHuman() || p.GetIsFinished() || p.GetDecided() {
		return ErrNotHumanTurn
	}
	b.applyDecide(idx, play)
	return nil
}

// CpuDecide 現在の決定手番がCPUなら参加可否を決める
func (b *Bourre) cpuDecide() {
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= len(b.players) {
		return
	}
	p := b.players[idx]
	if p == nil || p.GetIsHuman() || p.GetIsFinished() || p.GetDecided() {
		return
	}
	b.applyDecide(idx, b.cpuShouldPlay(idx))
}

// applyDecide 参加可否を適用して次へ進める
func (b *Bourre) applyDecide(idx int, play bool) {
	p := b.players[idx]
	p.SetDecided(true)
	p.SetFolded(!play)
	action := "play"
	detail := fmt.Sprintf("%s plays", b.playerName(idx))
	if !play {
		action = "fold"
		detail = fmt.Sprintf("%s folds", b.playerName(idx))
	}
	b.appendLog(idx, action, detail, nil)
	b.advanceDecide()
}

// advanceDecide 次の未決定プレイヤーへ進む。全員決定済みなら次フェーズへ。
func (b *Bourre) advanceDecide() {
	n := len(b.players)
	for i := 1; i <= n; i++ {
		idx := (b.currentPlayerIdx + i) % n
		if !b.players[idx].GetIsFinished() && !b.players[idx].GetDecided() {
			b.currentPlayerIdx = idx
			return
		}
	}
	b.finishDecide()
}

// finishDecide 全員の決定が済んだあとの遷移
func (b *Bourre) finishDecide() {
	if b.activeCount() <= 1 {
		b.resolveNoContest()
		return
	}
	b.phase = BourrePhaseDraw
	b.currentPlayerIdx = b.firstActiveFrom((b.dealerIdx + 1) % len(b.players))
}

// --- ドローフェーズ ---

// PlayerDraw 人間プレイヤーが不要なカードを捨てて同数ドローする
func (b *Bourre) PlayerDraw(indices []int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BourrePhaseDraw {
		return ErrWrongPhase
	}
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= len(b.players) {
		return ErrNotHumanTurn
	}
	p := b.players[idx]
	if p == nil || !p.GetIsHuman() || !b.isActive(idx) || p.GetDrawn() {
		return ErrNotHumanTurn
	}
	if err := b.validateDrawIndices(p, indices); err != nil {
		return err
	}
	b.applyDraw(idx, indices)
	return nil
}

// cpuDraw 現在のドロー手番がCPUなら手札交換する
func (b *Bourre) cpuDraw() {
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= len(b.players) {
		return
	}
	p := b.players[idx]
	if p == nil || p.GetIsHuman() || !b.isActive(idx) || p.GetDrawn() {
		return
	}
	b.applyDraw(idx, b.cpuSelectDiscards(idx))
}

// validateDrawIndices ドローインデックスの妥当性を検証する
func (b *Bourre) validateDrawIndices(p *BourrePlayer, indices []int) error {
	seen := make(map[int]bool, len(indices))
	for _, i := range indices {
		if i < 0 || i >= p.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "card index out of range")
		}
		if seen[i] {
			return NewDomainError(ErrInvalidPlay, "duplicate card index")
		}
		seen[i] = true
	}
	return nil
}

// applyDraw 手札交換を適用して次へ進める
func (b *Bourre) applyDraw(idx int, indices []int) {
	p := b.players[idx]
	removed := p.RemoveCards(indices)
	for range removed {
		c := b.trumpCards.DrawCard()
		if c != nil {
			p.AddCard(c)
		}
	}
	b.sortHand(p)
	p.SetDrawn(true)
	b.appendLog(idx, "draw", fmt.Sprintf("%s draws %d", b.playerName(idx), len(removed)), nil)
	b.advanceDraw()
}

// advanceDraw 次の未ドローのアクティブプレイヤーへ進む。全員済みならプレイ開始。
func (b *Bourre) advanceDraw() {
	n := len(b.players)
	for i := 1; i <= n; i++ {
		idx := (b.currentPlayerIdx + i) % n
		if b.isActive(idx) && !b.players[idx].GetDrawn() {
			b.currentPlayerIdx = idx
			return
		}
	}
	b.startPlay()
}

// startPlay トリックプレイフェーズを開始する
func (b *Bourre) startPlay() {
	b.phase = BourrePhasePlay
	b.trickNumber = 1
	b.currentTrick = nil
	b.lastTrick = nil
	b.leadPlayerIdx = b.firstActiveFrom((b.dealerIdx + 1) % len(b.players))
	b.currentPlayerIdx = b.leadPlayerIdx
}

// --- トリックプレイフェーズ ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (b *Bourre) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BourrePhasePlay {
		return ErrWrongPhase
	}
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= len(b.players) {
		return ErrNotHumanTurn
	}
	p := b.players[idx]
	if p == nil || !p.GetIsHuman() || !b.isActive(idx) {
		return ErrNotHumanTurn
	}
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "card index out of range")
	}
	if !b.isLegalPlay(idx, p.GetCard(cardIndex)) {
		return NewDomainError(ErrInvalidPlay, "must follow suit, trump when void, and beat the trick if able")
	}
	played := p.RemoveCard(cardIndex)
	b.playCard(idx, played)
	return nil
}

// cpuPlayCard 現在のプレイ手番がCPUなら1枚プレイする
func (b *Bourre) cpuPlayCard() {
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= len(b.players) {
		return
	}
	p := b.players[idx]
	if p == nil || p.GetIsHuman() || !b.isActive(idx) || p.GetCardsSize() == 0 {
		return
	}
	cardIdx := b.cpuSelectPlay(idx)
	played := p.RemoveCard(cardIdx)
	if played == nil {
		return
	}
	b.playCard(idx, played)
}

// CpuPlay フェーズに応じてCPUの1アクションを実行する
func (b *Bourre) CpuPlay() {
	if b.gameEndFlag {
		return
	}
	switch b.phase {
	case BourrePhaseDecide:
		b.cpuDecide()
	case BourrePhaseDraw:
		b.cpuDraw()
	case BourrePhasePlay:
		b.cpuPlayCard()
	}
}

// playCard カードをトリックに加える共通処理
func (b *Bourre) playCard(playerIdx int, card *Card) {
	b.currentTrick = append(b.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	b.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", b.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(b.currentTrick) == b.activeCount() {
		b.resolveTrick()
	} else {
		b.currentPlayerIdx = b.nextActive(playerIdx)
	}
}

// resolveTrick トリックを解決し、勝者にトリックを付与する
func (b *Bourre) resolveTrick() {
	winner := b.trickWinner()
	trickCards := make([]*Card, len(b.currentTrick))
	for i, tc := range b.currentTrick {
		trickCards[i] = tc.Card
	}
	b.players[winner].AddTrick(trickCards)
	b.appendLog(winner, "trick_win", fmt.Sprintf("%s wins trick %d", b.playerName(winner), b.trickNumber), trickCards)

	b.leadPlayerIdx = winner
	b.lastTrick = b.currentTrick
	b.lastTrickWinner = winner
	b.currentTrick = nil

	if b.trickNumber >= BourreHandSize {
		b.scoreHand()
	} else {
		b.trickNumber++
		b.currentPlayerIdx = winner
	}
}

// trickWinner 現在のトリックの勝者インデックスを返す
func (b *Bourre) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	bestIdx := b.currentTrick[0].PlayerIdx
	best := b.currentTrick[0].Card
	for _, tc := range b.currentTrick[1:] {
		if b.beats(tc.Card, best, leadSuit) {
			best = tc.Card
			bestIdx = tc.PlayerIdx
		}
	}
	return bestIdx
}

// beats c が現在の勝ち札 cur を上回るか (leadSuit はトリックのリードスート)
func (b *Bourre) beats(c, cur *Card, leadSuit int) bool {
	cTrump := c.GetDesign() == b.trumpSuit
	curTrump := cur.GetDesign() == b.trumpSuit
	switch {
	case cTrump && !curTrump:
		return true
	case !cTrump && curTrump:
		return false
	case cTrump && curTrump:
		return bourreRank(c) > bourreRank(cur)
	default:
		// どちらも非トランプ: リードスート同士のみ比較
		if c.GetDesign() == leadSuit && cur.GetDesign() == leadSuit {
			return bourreRank(c) > bourreRank(cur)
		}
		// cur は常にリードスートか切り札なので、ここに来る c は無効スート
		return false
	}
}

// --- 得点・精算 ---

// resolveNoContest 参加者が1人以下のときの精算
func (b *Bourre) resolveNoContest() {
	potValue := b.pot
	soleWinner := -1
	for i := range b.players {
		if b.isActive(i) {
			soleWinner = i
			break
		}
	}
	if soleWinner >= 0 {
		b.players[soleWinner].AddChips(potValue)
		b.appendLog(soleWinner, "pot_win", fmt.Sprintf("%s takes the pot (%d) uncontested", b.playerName(soleWinner), potValue), nil)
	} else {
		b.carryPot += potValue
		b.appendLog(-1, "pot_carry", fmt.Sprintf("All folded — pot of %d carries over", potValue), nil)
	}
	b.pot = 0
	b.buildResults(potValue, soleWinner)
	b.finishHand()
}

// scoreHand 5トリック終了後の得点・ポット精算
func (b *Bourre) scoreHand() {
	potValue := b.pot
	maxTricks := -1
	for i := range b.players {
		if b.isActive(i) {
			if t := b.players[i].GetTrickCount(); t > maxTricks {
				maxTricks = t
			}
		}
	}
	winners := make([]int, 0, len(b.players))
	for i := range b.players {
		if b.isActive(i) && b.players[i].GetTrickCount() == maxTricks {
			winners = append(winners, i)
		}
	}

	// ブーレ罰金: 参加して0トリックのプレイヤーはポットと同額を次ポットへ
	for i := range b.players {
		if b.isActive(i) && b.players[i].GetTrickCount() == 0 {
			b.players[i].SetBourreed(true)
			pen := min(potValue, b.players[i].GetChips())
			b.players[i].SubtractChips(pen)
			b.carryPot += pen
			b.appendLog(i, "bourre", fmt.Sprintf("%s is bourréd! pays %d penalty", b.playerName(i), pen), nil)
		}
	}

	winnerIdx := -1
	if len(winners) == 1 {
		winnerIdx = winners[0]
		b.players[winnerIdx].AddChips(potValue)
		b.appendLog(winnerIdx, "pot_win", fmt.Sprintf("%s wins the pot (%d) with %d tricks", b.playerName(winnerIdx), potValue, maxTricks), nil)
	} else {
		b.carryPot += potValue
		b.appendLog(-1, "pot_carry", fmt.Sprintf("Tie for most tricks — pot of %d carries over", potValue), nil)
	}
	b.pot = 0
	b.buildResults(potValue, winnerIdx)
	b.finishHand()
}

// buildResults 表示用のハンド結果を構築する
func (b *Bourre) buildResults(potValue, winnerIdx int) {
	b.lastResults = make([]*BourreHandResult, 0, len(b.players))
	for i := range b.players {
		p := b.players[i]
		if p.GetIsFinished() && !p.GetDecided() {
			continue
		}
		won := 0
		if i == winnerIdx {
			won = potValue
		}
		b.lastResults = append(b.lastResults, &BourreHandResult{
			PlayerIdx: i,
			Tricks:    p.GetTrickCount(),
			WonAmount: won,
			Bourreed:  p.GetBourreed(),
			Folded:    p.GetFolded(),
		})
	}
}

// finishHand ハンド終了の共通処理 (ディーラー移動・ゲーム終了判定)
func (b *Bourre) finishHand() {
	b.phase = BourrePhaseRoundEnd
	if n := len(b.players); n > 0 {
		b.dealerIdx = (b.dealerIdx + 1) % n
	}
	b.checkGameEnd()
}

// checkGameEnd ゲーム終了判定 (生存者1人以下、または人間が破産)
func (b *Bourre) checkGameEnd() {
	if len(b.players) == 0 {
		return
	}
	solvent := 0
	for _, p := range b.players {
		if p.GetChips() > 0 {
			solvent++
		}
	}
	humanIdx := b.findHumanIdx()
	humanBroke := humanIdx >= 0 && b.players[humanIdx].GetChips() <= 0
	if solvent > 1 && !humanBroke {
		return
	}
	b.gameEndFlag = true
	b.phase = BourrePhaseGameEnd
	b.winnerIdx = 0
	for i := 1; i < len(b.players); i++ {
		if b.players[i].GetChips() > b.players[b.winnerIdx].GetChips() {
			b.winnerIdx = i
		}
	}
	b.appendLog(b.winnerIdx, "game_end", fmt.Sprintf("%s wins the game with %d chips!", b.playerName(b.winnerIdx), b.players[b.winnerIdx].GetChips()), nil)
}

// --- 補助 ---

// participantsFrom start から時計回りに参加者 (チップあり) のインデックスを返す
func (b *Bourre) participantsFrom(start int) []int {
	n := len(b.players)
	order := make([]int, 0, n)
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		if !b.players[idx].GetIsFinished() {
			order = append(order, idx)
		}
	}
	return order
}

// firstActiveFrom start から時計回りで最初のアクティブ (参加 & 非フォールド) プレイヤー
func (b *Bourre) firstActiveFrom(start int) int {
	n := len(b.players)
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		if b.isActive(idx) {
			return idx
		}
	}
	return start
}

// nextActive from の次のアクティブプレイヤー
func (b *Bourre) nextActive(from int) int {
	n := len(b.players)
	for i := 1; i <= n; i++ {
		idx := (from + i) % n
		if b.isActive(idx) {
			return idx
		}
	}
	return from
}

// isActive 参加 (チップあり) かつ非フォールドか
func (b *Bourre) isActive(idx int) bool {
	if idx < 0 || idx >= len(b.players) {
		return false
	}
	p := b.players[idx]
	if p == nil {
		return false
	}
	return !p.GetIsFinished() && !p.GetFolded()
}

// activeCount アクティブプレイヤー数
func (b *Bourre) activeCount() int {
	cnt := 0
	for i := range b.players {
		if b.isActive(i) {
			cnt++
		}
	}
	return cnt
}

// findHumanIdx 人間プレイヤーのインデックス (-1 = なし)
func (b *Bourre) findHumanIdx() int {
	for i, p := range b.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// isLegalPlay card のプレイが合法か (legalPlays への所属で判定)
func (b *Bourre) isLegalPlay(idx int, card *Card) bool {
	for _, i := range b.legalPlays(idx) {
		if b.players[idx].GetCard(i) == card {
			return true
		}
	}
	return false
}

// legalPlays マストフォロー・マストトランプ・マストウィンを満たす合法手のインデックス集合
func (b *Bourre) legalPlays(idx int) []int {
	p := b.players[idx]
	n := p.GetCardsSize()
	allIdx := func() []int {
		r := make([]int, n)
		for i := range r {
			r[i] = i
		}
		return r
	}
	if len(b.currentTrick) == 0 {
		return allIdx() // リードは任意
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()

	hasTrump := false
	highTrump := 0
	highLead := 0
	for _, tc := range b.currentTrick {
		if tc.Card.GetDesign() == b.trumpSuit {
			hasTrump = true
			if r := bourreRank(tc.Card); r > highTrump {
				highTrump = r
			}
		} else if tc.Card.GetDesign() == leadSuit {
			if r := bourreRank(tc.Card); r > highLead {
				highLead = r
			}
		}
	}

	leadIdx := make([]int, 0, n)
	trumpIdx := make([]int, 0, n)
	for i := 0; i < n; i++ {
		c := p.GetCard(i)
		if c.GetDesign() == leadSuit {
			leadIdx = append(leadIdx, i)
		}
		if c.GetDesign() == b.trumpSuit {
			trumpIdx = append(trumpIdx, i)
		}
	}

	// リードスートを持っている (リードが切り札以外)
	if len(leadIdx) > 0 && leadSuit != b.trumpSuit {
		if !hasTrump {
			if w := b.higherThan(p, leadIdx, highLead); len(w) > 0 {
				return w // 勝てるなら勝たねばならない
			}
		}
		return leadIdx
	}
	// リードスートが切り札 (= 切り札フォロー)
	if len(leadIdx) > 0 && leadSuit == b.trumpSuit {
		if w := b.higherThan(p, leadIdx, highTrump); len(w) > 0 {
			return w
		}
		return leadIdx
	}
	// リードスート無し → 切り札を出さねばならない
	if len(trumpIdx) > 0 {
		if hasTrump {
			if w := b.higherThan(p, trumpIdx, highTrump); len(w) > 0 {
				return w // オーバートランプできるならする
			}
		}
		return trumpIdx
	}
	// リードも切り札も無い → 任意のカードを捨てる
	return allIdx()
}

// higherThan cand の中で rank が threshold を上回るインデックス集合
func (b *Bourre) higherThan(p *BourrePlayer, cand []int, threshold int) []int {
	w := make([]int, 0, len(cand))
	for _, i := range cand {
		if bourreRank(p.GetCard(i)) > threshold {
			w = append(w, i)
		}
	}
	return w
}

// sortHand 手札をスート→ランク (エース高) でソートする
func (b *Bourre) sortHand(p *BourrePlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return bourreRank(cards[i]) < bourreRank(cards[j])
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー表示名
func (b *Bourre) playerName(idx int) string {
	if idx < 0 || idx >= len(b.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if b.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (b *Bourre) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: len(b.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// bourreRank カードのランクを返す (エース=14、それ以外は額面値)
func bourreRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// suitNameOf design 定数のスート名 (英語)
func suitNameOf(design int) string {
	switch design {
	case CardDesignSpade:
		return "Spades"
	case CardDesignClover:
		return "Clubs"
	case CardDesignHeart:
		return "Hearts"
	case CardDesignDiamond:
		return "Diamonds"
	default:
		return "?"
	}
}

// trumpCardSlice trumpCard を []*Card に変換する (nil は空)
func trumpCardSlice(c *Card) []*Card {
	if c == nil {
		return nil
	}
	return []*Card{c}
}

// --- ゲッター ---

// GetPhase 現在のフェーズ取得
func (b *Bourre) GetPhase() BourrePhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Bourre) SetPhase(p BourrePhase) { b.phase = p }

// GetGameEndFlag ゲーム終了フラグ取得
func (b *Bourre) GetGameEndFlag() bool { return b.gameEndFlag }

// GetPlayerCnt プレイヤー数取得
func (b *Bourre) GetPlayerCnt() int { return len(b.players) }

// GetPlayer プレイヤー取得
func (b *Bourre) GetPlayer(i int) *BourrePlayer {
	if i < 0 || i >= len(b.players) {
		return nil
	}
	return b.players[i]
}

// GetCurrentPlayerIdx 現在の手番プレイヤーインデックス取得
func (b *Bourre) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// SetCurrentPlayerIdx 現在の手番プレイヤーインデックス設定 (テスト用)
func (b *Bourre) SetCurrentPlayerIdx(i int) { b.currentPlayerIdx = i }

// GetPot 現在のポット取得
func (b *Bourre) GetPot() int { return b.pot }

// GetCarryPot 繰り越しポット取得
func (b *Bourre) GetCarryPot() int { return b.carryPot }

// GetTrumpSuit 切り札スート (Card design) 取得
func (b *Bourre) GetTrumpSuit() int { return b.trumpSuit }

// GetTrumpCard 切り札カード取得
func (b *Bourre) GetTrumpCard() *Card { return b.trumpCard }

// GetDealerIdx ディーラーインデックス取得
func (b *Bourre) GetDealerIdx() int { return b.dealerIdx }

// GetTrickNumber 現在のトリック番号取得
func (b *Bourre) GetTrickNumber() int { return b.trickNumber }

// GetCurrentTrick 現在進行中のトリック取得
func (b *Bourre) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// GetLastTrick 直前に解決されたトリック取得
func (b *Bourre) GetLastTrick() []*TrickCard { return b.lastTrick }

// GetLastTrickWinner 直前トリックの勝者インデックス取得 (-1 = なし)
func (b *Bourre) GetLastTrickWinner() int { return b.lastTrickWinner }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (b *Bourre) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// GetHandNumber 現在のハンド番号取得
func (b *Bourre) GetHandNumber() int { return b.handNumber }

// GetWinnerIdx ゲーム勝者インデックス取得 (-1 = 未確定)
func (b *Bourre) GetWinnerIdx() int { return b.winnerIdx }

// GetLastResults 直前ハンドの結果一覧取得
func (b *Bourre) GetLastResults() []*BourreHandResult { return b.lastResults }

// GetConfig 設定取得
func (b *Bourre) GetConfig() BourreConfig { return b.config }

// SetConfig 設定変更
func (b *Bourre) SetConfig(cfg BourreConfig) { b.config = cfg }

// GetActionLog 棋譜取得
func (b *Bourre) GetActionLog() []*ActionLogEntry { return b.actionLog }

// IsHumanTurn 現在の手番が人間かどうか
func (b *Bourre) IsHumanTurn() bool {
	if b.gameEndFlag {
		return false
	}
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= len(b.players) {
		return false
	}
	p := b.players[idx]
	if !p.GetIsHuman() {
		return false
	}
	switch b.phase {
	case BourrePhaseDecide:
		return !p.GetIsFinished() && !p.GetDecided()
	case BourrePhaseDraw:
		return b.isActive(idx) && !p.GetDrawn()
	case BourrePhasePlay:
		return b.isActive(idx)
	default:
		return false
	}
}

// GetValidPlayIndices 人間がプレイ可能なカードのインデックス一覧 (Web/CUI用)
func (b *Bourre) GetValidPlayIndices(idx int) []int {
	if b.phase != BourrePhasePlay || !b.isActive(idx) {
		return nil
	}
	return b.legalPlays(idx)
}

// bourreJSON is the JSON wire format for Bourre.
type bourreJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*BourrePlayer     `json:"ps"`
	Config           BourreConfig        `json:"cf"`
	Phase            BourrePhase         `json:"ph"`
	Pot              int                 `json:"pt"`
	CarryPot         int                 `json:"cp"`
	TrumpSuit        int                 `json:"ts"`
	TrumpCard        *Card               `json:"td"`
	DealerIdx        int                 `json:"di"`
	CurrentPlayerIdx int                 `json:"ci"`
	TrickNumber      int                 `json:"tn"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	LastTrick        []*TrickCard        `json:"lt"`
	LastTrickWinner  int                 `json:"lw"`
	LeadPlayerIdx    int                 `json:"li"`
	HandNumber       int                 `json:"hn"`
	GameEndFlag      bool                `json:"ge"`
	WinnerIdx        int                 `json:"wi"`
	LastResults      []*BourreHandResult `json:"lr"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// bourreMaxSliceLen caps slice sizes during deserialisation.
const bourreMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (b *Bourre) MarshalJSON() ([]byte, error) {
	return json.Marshal(bourreJSON{
		TrumpCards:       b.trumpCards,
		Players:          b.players,
		Config:           b.config,
		Phase:            b.phase,
		Pot:              b.pot,
		CarryPot:         b.carryPot,
		TrumpSuit:        b.trumpSuit,
		TrumpCard:        b.trumpCard,
		DealerIdx:        b.dealerIdx,
		CurrentPlayerIdx: b.currentPlayerIdx,
		TrickNumber:      b.trickNumber,
		CurrentTrick:     b.currentTrick,
		LastTrick:        b.lastTrick,
		LastTrickWinner:  b.lastTrickWinner,
		LeadPlayerIdx:    b.leadPlayerIdx,
		HandNumber:       b.handNumber,
		GameEndFlag:      b.gameEndFlag,
		WinnerIdx:        b.winnerIdx,
		LastResults:      b.lastResults,
		ActionLog:        b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bourre) UnmarshalJSON(data []byte) error {
	var j bourreJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > bourreMaxSliceLen || len(j.CurrentTrick) > bourreMaxSliceLen ||
		len(j.LastTrick) > bourreMaxSliceLen || len(j.LastResults) > bourreMaxSliceLen ||
		len(j.ActionLog) > bourreMaxSliceLen {
		return fmt.Errorf("bourre: input array exceeds maximum allowed size")
	}
	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCards(0)
	}
	b.players = j.Players
	if b.players == nil {
		b.players = make([]*BourrePlayer, 0)
	}
	b.config = j.Config
	b.phase = j.Phase
	b.pot = j.Pot
	b.carryPot = j.CarryPot
	b.trumpSuit = j.TrumpSuit
	b.trumpCard = j.TrumpCard
	b.dealerIdx = j.DealerIdx
	b.currentPlayerIdx = j.CurrentPlayerIdx
	b.trickNumber = j.TrickNumber
	b.currentTrick = j.CurrentTrick
	if b.currentTrick == nil {
		b.currentTrick = make([]*TrickCard, 0)
	}
	b.lastTrick = j.LastTrick
	if b.lastTrick == nil {
		b.lastTrick = make([]*TrickCard, 0)
	}
	b.lastTrickWinner = j.LastTrickWinner
	b.leadPlayerIdx = j.LeadPlayerIdx
	b.handNumber = j.HandNumber
	b.gameEndFlag = j.GameEndFlag
	b.winnerIdx = j.WinnerIdx
	b.lastResults = j.LastResults
	if b.lastResults == nil {
		b.lastResults = make([]*BourreHandResult, 0)
	}
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
