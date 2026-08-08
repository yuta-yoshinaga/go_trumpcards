//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// SevenBridgePlayerCnt セブンブリッジのプレイヤー数
const SevenBridgePlayerCnt = 2

// SevenBridgeHandSize 初期配布枚数
const SevenBridgeHandSize = 7

// SevenBridgeMeldMinSize メルドに必要な最小枚数
const SevenBridgeMeldMinSize = 3

// SevenBridgePivotRank 7（ピボットランク）
const SevenBridgePivotRank = 7

// SevenBridgeSevenPenalty 手札に残った 7 の 1 枚あたりペナルティ
const SevenBridgeSevenPenalty = 50

// SevenBridgePhase ゲームフェーズ
type SevenBridgePhase int

// SevenBridge のフェーズ定数
const (
	// SevenBridgePhaseDraw ドローフェーズ（山札から引く／ポン／チー）
	SevenBridgePhaseDraw SevenBridgePhase = 0
	// SevenBridgePhasePlay プレイフェーズ（メルド／レイオフ／ディスカード）
	SevenBridgePhasePlay SevenBridgePhase = 1
	// SevenBridgePhaseRoundEnd ラウンド終了フェーズ
	SevenBridgePhaseRoundEnd SevenBridgePhase = 2
	// SevenBridgePhaseGameEnd ゲーム終了フェーズ
	SevenBridgePhaseGameEnd SevenBridgePhase = 3
)

// SevenBridge セブンブリッジのゲームクラス
type SevenBridge struct {
	trumpCards       *TrumpCards
	players          []*SevenBridgePlayer
	config           SevenBridgeConfig
	phase            SevenBridgePhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
	roundWinnerIdx  int  // ラウンド勝者（上がったプレイヤー）。-1 は山切れ流局
	claimedThisTurn bool // 直前のターンで discard を pon/chi で取得したか
}

// NewSevenBridge コンストラクタ
func NewSevenBridge(trumpCards *TrumpCards, players []*SevenBridgePlayer, config SevenBridgeConfig) *SevenBridge {
	return &SevenBridge{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerIdx:      -1,
		roundNumber:    0,
		roundWinnerIdx: -1,
	}
}

// NewDefaultSevenBridge 標準構成（人間 1 / CPU 1）でコンストラクトする。
// CUI・Web・Worker のすべての起動経路から呼び出される SSoT。
func NewDefaultSevenBridge() *SevenBridge {
	players := []*SevenBridgePlayer{
		NewSevenBridgePlayer(true),
		NewSevenBridgePlayer(false),
	}
	return NewSevenBridge(NewTrumpCards(0), players, DefaultSevenBridgeConfig())
}

// Reset ゲームを初期化する
func (g *SevenBridge) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.roundWinnerIdx = -1
	g.claimedThisTurn = false

	for _, p := range g.players {
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.Reset()
		p.SetIsFinished(false)
		p.ClearMelds()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = SevenBridgePhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *SevenBridge) NextRound() {
	if g.phase != SevenBridgePhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	// 前ラウンドの勝者が先攻。勝者未確定なら 0 へ。
	if g.roundWinnerIdx >= 0 && g.roundWinnerIdx < len(g.players) {
		g.currentPlayerIdx = g.roundWinnerIdx
	} else {
		g.currentPlayerIdx = 0
	}
	g.roundWinnerIdx = -1
	g.claimedThisTurn = false

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = SevenBridgePhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに 7 枚、捨て札先頭には最初に見つかる 7 を配置
func (g *SevenBridge) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}

	rand.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	// 各プレイヤーに 7 枚配布
	for range SevenBridgeHandSize {
		for j := range SevenBridgePlayerCnt {
			if len(g.drawPile) == 0 {
				break
			}
			card := g.drawPile[len(g.drawPile)-1]
			g.drawPile = g.drawPile[:len(g.drawPile)-1]
			g.players[j].AddCard(card)
		}
	}

	// 最初の捨て札は必ず 7 から開始する（ルール上のピボット）。
	// まず残りの山札から 7 を探し、見つからなければプレイヤーの手札から 1 枚の 7 と
	// 山札の一番上のカードを交換して、必ず 7 を捨て札トップに置く。
	if idx := findFirstValueIdx(g.drawPile, SevenBridgePivotRank); idx >= 0 {
		first := g.drawPile[idx]
		g.drawPile = append(g.drawPile[:idx], g.drawPile[idx+1:]...)
		g.discardPile = append(g.discardPile, first)
		return
	}
	if len(g.drawPile) == 0 {
		return
	}
	swapTop := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	for _, p := range g.players {
		for i := 0; i < p.GetCardsSize(); i++ {
			if p.GetCard(i).GetValue() == SevenBridgePivotRank {
				seven := p.RemoveCard(i)
				p.AddCard(swapTop)
				g.discardPile = append(g.discardPile, seven)
				return
			}
		}
	}
	// 4 枚すべての 7 が消える事態は 52 枚デッキでは起こり得ないが、念のため
	// 保険として取り出した 1 枚を単純に捨て札へ戻す。
	g.discardPile = append(g.discardPile, swapTop)
}

// findFirstValueIdx returns the last index of a card whose value matches v,
// or -1 if none. Iterates from the top (tail) because dealInitialCards's
// draw-pile is a stack.
func findFirstValueIdx(pile []*Card, v int) int {
	for i := len(pile) - 1; i >= 0; i-- {
		if pile[i].GetValue() == v {
			return i
		}
	}
	return -1
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *SevenBridge) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SevenBridgePhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	return g.drawFromStock()
}

// drawFromStock 山札からカードを引く（人間・CPU 共通の内部処理）
func (g *SevenBridge) drawFromStock() error {
	if len(g.drawPile) == 0 {
		g.endRoundStockOut()
		return nil
	}
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
	g.claimedThisTurn = false
	g.phase = SevenBridgePhasePlay
	return nil
}

// PlayerClaimPon 人間プレイヤーがポンで捨て札を取得する
// cardIndices は、捨て札トップと同ランクの手札 2 枚のインデックス。
func (g *SevenBridge) PlayerClaimPon(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SevenBridgePhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyPonClaim(cardIndices)
}

// PlayerClaimChi 人間プレイヤーがチーで捨て札を取得する
// cardIndices は、捨て札トップと同スートで連続するランを作る手札 2 枚のインデックス。
func (g *SevenBridge) PlayerClaimChi(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SevenBridgePhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyChiClaim(cardIndices)
}

// applyPonClaim ポン成立処理（人間・CPU 共通）
func (g *SevenBridge) applyPonClaim(cardIndices []int) error {
	if len(cardIndices) != 2 {
		return NewDomainError(ErrInvalidPlay, "ポンには手札 2 枚のインデックスが必要です")
	}
	if err := validateIndexPair(cardIndices, g.players[g.currentPlayerIdx].GetCardsSize()); err != nil {
		return err
	}
	top := g.GetDiscardTop()
	if top == nil {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	player := g.players[g.currentPlayerIdx]
	c1 := player.GetCard(cardIndices[0])
	c2 := player.GetCard(cardIndices[1])
	if c1.GetValue() != top.GetValue() || c2.GetValue() != top.GetValue() {
		return NewDomainError(ErrInvalidPlay, "ポンの 2 枚は捨て札と同じランクでなければなりません")
	}

	// 捨て札トップを取り出し、手札 2 枚と合わせて新メルドを登録
	claimed := g.popDiscardTop()
	meld := []*Card{claimed, c1, c2}
	sort.Slice(meld, func(i, j int) bool { return meld[i].GetValue() < meld[j].GetValue() })
	player.AppendMeld(meld)

	indices := make([]int, len(cardIndices))
	copy(indices, cardIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	for _, idx := range indices {
		player.RemoveCard(idx)
	}

	g.appendLog(g.currentPlayerIdx, "pon", fmt.Sprintf("%s calls Pon on %s", playerName(g.players, g.currentPlayerIdx), cardStr(claimed)), []*Card{claimed, c1, c2})
	g.claimedThisTurn = true
	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
		return nil
	}
	g.phase = SevenBridgePhasePlay
	return nil
}

// applyChiClaim チー成立処理（人間・CPU 共通）
func (g *SevenBridge) applyChiClaim(cardIndices []int) error {
	if len(cardIndices) != 2 {
		return NewDomainError(ErrInvalidPlay, "チーには手札 2 枚のインデックスが必要です")
	}
	if err := validateIndexPair(cardIndices, g.players[g.currentPlayerIdx].GetCardsSize()); err != nil {
		return err
	}
	top := g.GetDiscardTop()
	if top == nil {
		return NewDomainError(ErrInvalidPlay, "捨て札が空です")
	}
	player := g.players[g.currentPlayerIdx]
	c1 := player.GetCard(cardIndices[0])
	c2 := player.GetCard(cardIndices[1])
	if c1.GetDesign() != top.GetDesign() || c2.GetDesign() != top.GetDesign() {
		return NewDomainError(ErrInvalidPlay, "チーの 2 枚は捨て札と同じスートでなければなりません")
	}

	combo := []int{top.GetValue(), c1.GetValue(), c2.GetValue()}
	sort.Ints(combo)
	if combo[0]+1 != combo[1] || combo[1]+1 != combo[2] {
		return NewDomainError(ErrInvalidPlay, "チーは捨て札と連続する 3 枚のランでなければなりません")
	}

	claimed := g.popDiscardTop()
	meld := []*Card{claimed, c1, c2}
	sort.Slice(meld, func(i, j int) bool { return meld[i].GetValue() < meld[j].GetValue() })
	player.AppendMeld(meld)

	indices := make([]int, len(cardIndices))
	copy(indices, cardIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	for _, idx := range indices {
		player.RemoveCard(idx)
	}

	g.appendLog(g.currentPlayerIdx, "chi", fmt.Sprintf("%s calls Chi on %s", playerName(g.players, g.currentPlayerIdx), cardStr(claimed)), []*Card{claimed, c1, c2})
	g.claimedThisTurn = true
	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
		return nil
	}
	g.phase = SevenBridgePhasePlay
	return nil
}

// PlayerMeld 人間プレイヤーが手札からメルドを場に出す
func (g *SevenBridge) PlayerMeld(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SevenBridgePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyMeld(cardIndices)
}

func (g *SevenBridge) applyMeld(cardIndices []int) error {
	if len(cardIndices) < SevenBridgeMeldMinSize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("メルドには最低 %d 枚必要です", SevenBridgeMeldMinSize))
	}
	player := g.players[g.currentPlayerIdx]
	if err := validateIndexList(cardIndices, player.GetCardsSize()); err != nil {
		return err
	}

	cards := make([]*Card, 0, len(cardIndices))
	for _, idx := range cardIndices {
		cards = append(cards, player.GetCard(idx))
	}
	if !IsSevenBridgeMeld(cards) {
		return NewDomainError(ErrInvalidPlay, "有効なメルド（セットまたはラン）ではありません")
	}

	// 既存のメルドに影響を与えないよう昇順ソートして保存
	meldCopy := make([]*Card, len(cards))
	copy(meldCopy, cards)
	sort.SliceStable(meldCopy, func(i, j int) bool {
		if meldCopy[i].GetDesign() != meldCopy[j].GetDesign() {
			return meldCopy[i].GetDesign() < meldCopy[j].GetDesign()
		}
		return meldCopy[i].GetValue() < meldCopy[j].GetValue()
	})
	player.AppendMeld(meldCopy)

	sorted := make([]int, len(cardIndices))
	copy(sorted, cardIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, idx := range sorted {
		player.RemoveCard(idx)
	}

	g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d cards", playerName(g.players, g.currentPlayerIdx), len(cards)), cards)
	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerLayoff 人間プレイヤーが既存メルドにカードを 1 枚足す
// targetPlayerIdx: 追加先のメルドを保持するプレイヤー
// meldIdx: そのプレイヤーのメルドインデックス
// cardIndex: 自分の手札のインデックス
func (g *SevenBridge) PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SevenBridgePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyLayoff(targetPlayerIdx, meldIdx, cardIndex)
}

func (g *SevenBridge) applyLayoff(targetPlayerIdx, meldIdx, cardIndex int) error {
	if targetPlayerIdx < 0 || targetPlayerIdx >= len(g.players) {
		return NewDomainError(ErrInvalidPlay, "対象プレイヤーが不正です")
	}
	target := g.players[targetPlayerIdx]
	if meldIdx < 0 || meldIdx >= target.GetMeldCount() {
		return NewDomainError(ErrInvalidPlay, "対象メルドが不正です")
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	meld := target.GetMeld(meldIdx)
	if !canAddToMeld(meld, card) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%s はそのメルドに追加できません", cardStr(card)))
	}
	target.AddCardToMeld(meldIdx, card)
	player.RemoveCard(cardIndex)

	g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	if player.GetCardsSize() == 0 {
		g.finishRound(g.currentPlayerIdx)
	}
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *SevenBridge) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SevenBridgePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyDiscard(cardIndex)
}

func (g *SevenBridge) applyDiscard(cardIndex int) error {
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	// メルド無しで最後の 1 枚を捨てて上がる不正ルートを人間プレイヤー側で禁止する。
	// CPU は chooseCpuDiscard の合法手優先と applyLayoff/applyMeld の上がり判定で
	// このゾンビ状態に入る前にターンを抜けるため対象外。
	if player.GetIsHuman() && player.GetCardsSize() == 1 && player.GetMeldCount() == 0 {
		return NewDomainError(ErrInvalidPlay, "上がりには最低 1 つのメルドが必要です")
	}
	card := player.GetCard(cardIndex)

	top := g.GetDiscardTop()
	if top != nil && !IsDiscardLegal(card, top) {
		// 合法的な捨て札が存在する場合のみ制限する（詰み回避）
		if g.hasAnyLegalDiscard(player, top) {
			return NewDomainError(ErrInvalidPlay, "7 と同じランク／±1 のランク、または 7 のみ捨てられます")
		}
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	// 上がり判定（手札 0）
	if player.GetCardsSize() == 0 && player.GetMeldCount() > 0 {
		g.finishRound(g.currentPlayerIdx)
		return nil
	}

	g.advanceTurn()
	return nil
}

// advanceTurn 次のプレイヤーへ
func (g *SevenBridge) advanceTurn() {
	g.currentPlayerIdx = 1 - g.currentPlayerIdx
	g.phase = SevenBridgePhaseDraw
	g.claimedThisTurn = false
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *SevenBridge) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// CpuPlay 現在の手番が CPU の場合にターンを実行
func (g *SevenBridge) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case SevenBridgePhaseDraw:
		g.cpuDraw()
	case SevenBridgePhasePlay:
		g.cpuPlay()
	}
}

// cpuDraw CPU が山札／ポン／チーのいずれかで手札を補充する
func (g *SevenBridge) cpuDraw() {
	top := g.GetDiscardTop()
	if top != nil {
		if indices, ok := g.findPonIndices(g.currentPlayerIdx, top); ok && g.shouldCpuClaim() {
			if err := g.applyPonClaim(indices); err == nil {
				return
			}
		}
		if indices, ok := g.findChiIndices(g.currentPlayerIdx, top); ok && g.shouldCpuClaim() {
			if err := g.applyChiClaim(indices); err == nil {
				return
			}
		}
	}
	_ = g.drawFromStock()
}

// shouldCpuClaim 難易度に応じて CPU がクレームすべきかを返す
func (g *SevenBridge) shouldCpuClaim() bool {
	switch g.config.CpuDifficulty {
	case SevenBridgeCpuDifficultyHard:
		return true
	case SevenBridgeCpuDifficultyNormal:
		return rand.Intn(4) != 0
	default:
		return rand.Intn(2) == 0
	}
}

// cpuPlay CPU がメルド・レイオフ・ディスカードを順に実行する
func (g *SevenBridge) cpuPlay() {
	player := g.players[g.currentPlayerIdx]

	// 可能なメルドをひたすら出す（3 枚単位で）
	for {
		indices, ok := g.findBestMeldIndices(player)
		if !ok {
			break
		}
		if err := g.applyMeld(indices); err != nil {
			break
		}
	}

	// 既存メルドへのレイオフ
	for {
		layoffDone := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if targetIdx, meldIdx, ok := g.findLayoffTarget(card); ok {
				if err := g.applyLayoff(targetIdx, meldIdx, i); err == nil {
					layoffDone = true
					break
				}
			}
		}
		if !layoffDone {
			break
		}
	}

	// 上がり確認
	if player.GetCardsSize() == 0 && player.GetMeldCount() > 0 {
		// 上がり時は discard 無しでラウンド終了
		g.finishRound(g.currentPlayerIdx)
		return
	}

	// ディスカード選択。chooseCpuDiscard は必ず有効なインデックスを返す。
	discardIdx := g.chooseCpuDiscard(player)
	_ = g.applyDiscard(discardIdx)
}

// chooseCpuDiscard CPU のディスカード戦略。
// - 捨て札制約を満たすカードから、最も高い得点（＝相手に有利になりにくい）の物を優先
// - 難易度 Hard では 7 をできるだけ温存する（=捨てない）
func (g *SevenBridge) chooseCpuDiscard(player *SevenBridgePlayer) int {
	top := g.GetDiscardTop()

	// 合法な候補を列挙
	legal := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if top == nil || IsDiscardLegal(c, top) {
			legal = append(legal, i)
		}
	}
	if len(legal) == 0 {
		// 合法手が無ければ任意に 0 番を返す
		return 0
	}

	// 難易度 Hard では 7 を極力避ける
	if g.config.CpuDifficulty == SevenBridgeCpuDifficultyHard {
		filtered := make([]int, 0, len(legal))
		for _, idx := range legal {
			if player.GetCard(idx).GetValue() != SevenBridgePivotRank {
				filtered = append(filtered, idx)
			}
		}
		if len(filtered) > 0 {
			legal = filtered
		}
	}

	// 高得点カードを捨てて自分のデッドウッドを減らす
	bestIdx := legal[0]
	bestVal := sevenBridgeCardPenalty(player.GetCard(bestIdx))
	for _, idx := range legal[1:] {
		v := sevenBridgeCardPenalty(player.GetCard(idx))
		if v > bestVal {
			bestVal = v
			bestIdx = idx
		}
	}
	return bestIdx
}

// findPonIndices player の手札から discard と同ランクの 2 枚を探す
func (g *SevenBridge) findPonIndices(playerIdx int, top *Card) ([]int, bool) {
	p := g.players[playerIdx]
	matches := make([]int, 0, 2)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == top.GetValue() {
			matches = append(matches, i)
			if len(matches) == 2 {
				return matches, true
			}
		}
	}
	return nil, false
}

// findChiIndices player の手札から discard と同スート連続の 2 枚を探す
func (g *SevenBridge) findChiIndices(playerIdx int, top *Card) ([]int, bool) {
	p := g.players[playerIdx]
	sameSuit := make(map[int]int) // value -> hand index
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == top.GetDesign() {
			sameSuit[c.GetValue()] = i
		}
	}
	// 3 枚連続となる { top-2, top-1, top }, { top-1, top, top+1 }, { top, top+1, top+2 } を試す
	tv := top.GetValue()
	candidates := [][2]int{{tv - 2, tv - 1}, {tv - 1, tv + 1}, {tv + 1, tv + 2}}
	for _, c := range candidates {
		i1, ok1 := sameSuit[c[0]]
		i2, ok2 := sameSuit[c[1]]
		if ok1 && ok2 {
			return []int{i1, i2}, true
		}
	}
	return nil, false
}

// SuggestPon は playerIdx が捨て札の上札でポンできる手札インデックスを返す。
// できなければ nil。**判定は claim 経路と同じ findPonIndices を通す** — 別に
// 書くと、案内した組が拒否されることになる (#4904)。
func (g *SevenBridge) SuggestPon(playerIdx int) []int {
	top := g.GetDiscardTop()
	if top == nil || playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	idx, ok := g.findPonIndices(playerIdx, top)
	if !ok {
		return nil
	}
	return idx
}

// SuggestChi は playerIdx が捨て札の上札でチーできる手札インデックスを返す。
// できなければ nil。
func (g *SevenBridge) SuggestChi(playerIdx int) []int {
	top := g.GetDiscardTop()
	if top == nil || playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	idx, ok := g.findChiIndices(playerIdx, top)
	if !ok {
		return nil
	}
	return idx
}

// findBestMeldIndices プレイヤーの手札から最大のメルド（3 枚以上）を探す
// SuggestMeld は playerIdx の最善メルド (手札インデックス) を返す。メルドできる組が
// 無ければ nil。CUI ヒント用に findBestMeldIndices を公開する薄いラッパー。
func (g *SevenBridge) SuggestMeld(playerIdx int) []int {
	p := g.GetPlayer(playerIdx)
	if p == nil {
		return nil
	}
	if idxs, ok := g.findBestMeldIndices(p); ok {
		return idxs
	}
	return nil
}

// SuggestDiscard は playerIdx の推奨ディスカード手札インデックスを返す。手札が無ければ -1。
// CUI ヒント用に chooseCpuDiscard を公開する薄いラッパー。
func (g *SevenBridge) SuggestDiscard(playerIdx int) int {
	p := g.GetPlayer(playerIdx)
	if p == nil || p.GetCardsSize() == 0 {
		return -1
	}
	return g.chooseCpuDiscard(p)
}

func (g *SevenBridge) findBestMeldIndices(p *SevenBridgePlayer) ([]int, bool) {
	n := p.GetCardsSize()
	// セット候補
	byRank := make(map[int][]int)
	for i := range n {
		v := p.GetCard(i).GetValue()
		byRank[v] = append(byRank[v], i)
	}
	for _, group := range byRank {
		if len(group) >= SevenBridgeMeldMinSize {
			return group[:SevenBridgeMeldMinSize], true
		}
	}
	// ラン候補
	bySuit := make(map[int][]int)
	for i := range n {
		bySuit[p.GetCard(i).GetDesign()] = append(bySuit[p.GetCard(i).GetDesign()], i)
	}
	for _, group := range bySuit {
		if len(group) < SevenBridgeMeldMinSize {
			continue
		}
		sorted := make([]int, len(group))
		copy(sorted, group)
		sort.Slice(sorted, func(i, j int) bool {
			return p.GetCard(sorted[i]).GetValue() < p.GetCard(sorted[j]).GetValue()
		})
		for i := 0; i+SevenBridgeMeldMinSize <= len(sorted); i++ {
			run := []int{sorted[i]}
			for j := i + 1; j < len(sorted); j++ {
				if p.GetCard(sorted[j]).GetValue() == p.GetCard(run[len(run)-1]).GetValue()+1 {
					run = append(run, sorted[j])
				} else {
					break
				}
			}
			if len(run) >= SevenBridgeMeldMinSize {
				return run[:SevenBridgeMeldMinSize], true
			}
		}
	}
	return nil, false
}

// findLayoffTarget card を追加可能な既存メルドを返す（targetPlayerIdx, meldIdx）
func (g *SevenBridge) findLayoffTarget(card *Card) (int, int, bool) {
	for pi := range g.players {
		for mi := 0; mi < g.players[pi].GetMeldCount(); mi++ {
			if canAddToMeld(g.players[pi].GetMeld(mi), card) {
				return pi, mi, true
			}
		}
	}
	return 0, 0, false
}

// finishRound 上がり／山切れの最終スコアリング
func (g *SevenBridge) finishRound(winnerIdx int) {
	if g.phase == SevenBridgePhaseRoundEnd || g.phase == SevenBridgePhaseGameEnd {
		return
	}
	g.roundWinnerIdx = winnerIdx
	loserTotal := 0
	if winnerIdx >= 0 {
		loserIdx := 1 - winnerIdx
		loser := g.players[loserIdx]
		for i := 0; i < loser.GetCardsSize(); i++ {
			loserTotal += sevenBridgeCardPenalty(loser.GetCard(i))
		}
		// 勝者のラウンドスコア = 相手のペナルティ
		g.players[winnerIdx].SetRoundScore(loserTotal)
		g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s goes out! Opponent penalty: %d", playerName(g.players, winnerIdx), loserTotal), nil)
	} else {
		g.appendLog(-1, "draw", "Round ends in a draw (stock empty)", nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = SevenBridgePhaseRoundEnd
	}
}

// endRoundStockOut 山札切れでラウンド終了
func (g *SevenBridge) endRoundStockOut() {
	g.finishRound(-1)
}

// checkGameEnd ゲーム終了判定
func (g *SevenBridge) checkGameEnd() {
	hasWinner := false
	for i := 0; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}
	if !hasWinner {
		return
	}
	g.gameEndFlag = true
	g.phase = SevenBridgePhaseGameEnd

	maxScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetCumulativeScore() > maxScore {
			maxScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// hasAnyLegalDiscard 手札に top に対して合法な捨て札があるか
func (g *SevenBridge) hasAnyLegalDiscard(player *SevenBridgePlayer, top *Card) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if IsDiscardLegal(player.GetCard(i), top) {
			return true
		}
	}
	return false
}

// --- Getters / Setters ---

// GetPhase 現在のフェーズを取得
func (g *SevenBridge) GetPhase() SevenBridgePhase { return g.phase }

// SetPhase フェーズ設定（テスト用）
func (g *SevenBridge) SetPhase(phase SevenBridgePhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号
func (g *SevenBridge) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定（テスト用）
func (g *SevenBridge) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在の手番プレイヤー
func (g *SevenBridge) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番プレイヤー設定（テスト用）
func (g *SevenBridge) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山
func (g *SevenBridge) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定（テスト用）
func (g *SevenBridge) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札トップ
func (g *SevenBridge) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// popDiscardTop 捨て札トップを取り出して返す
func (g *SevenBridge) popDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	top := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	return top
}

// GetDrawPileCount 山札残り枚数
func (g *SevenBridge) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札設定（テスト用）
func (g *SevenBridge) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグ
func (g *SevenBridge) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス（-1 未確定）
func (g *SevenBridge) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数
func (g *SevenBridge) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *SevenBridge) GetPlayer(i int) *SevenBridgePlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetConfig 設定取得
func (g *SevenBridge) GetConfig() SevenBridgeConfig { return g.config }

// SetConfig 設定変更
func (g *SevenBridge) SetConfig(cfg SevenBridgeConfig) { g.config = cfg }

// GetRoundWinnerIdx 直近ラウンドの勝者
func (g *SevenBridge) GetRoundWinnerIdx() int { return g.roundWinnerIdx }

// GetClaimedThisTurn 直前ターンで claim されたか
func (g *SevenBridge) GetClaimedThisTurn() bool { return g.claimedThisTurn }

// --- Private helpers ---

func (g *SevenBridge) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

func (g *SevenBridge) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- Pure helpers ---

// IsSevenBridgeMeld cards が有効なメルド（3 枚以上のセット or ラン）か判定する
func IsSevenBridgeMeld(cards []*Card) bool {
	if len(cards) < SevenBridgeMeldMinSize {
		return false
	}
	if isSet(cards) {
		// 同ランク（4 種のスートまでなので最大 4 枚）
		return len(cards) <= 4 && allDistinctSuits(cards)
	}
	// ラン: 同スート＋連続
	suit := cards[0].GetDesign()
	values := make([]int, 0, len(cards))
	seen := make(map[int]bool)
	for _, c := range cards {
		if c.GetDesign() != suit {
			return false
		}
		if seen[c.GetValue()] {
			return false
		}
		seen[c.GetValue()] = true
		values = append(values, c.GetValue())
	}
	sort.Ints(values)
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}
	return true
}

// allDistinctSuits cards のスートが全て異なる場合 true
func allDistinctSuits(cards []*Card) bool {
	seen := make(map[int]bool)
	for _, c := range cards {
		if seen[c.GetDesign()] {
			return false
		}
		seen[c.GetDesign()] = true
	}
	return true
}

// IsDiscardLegal card が top に対して合法な捨て札か
//   - 初手（top == nil）は 7 のみ
//   - それ以降は：7 を含む / 同ランク / ランク ±1
func IsDiscardLegal(card, top *Card) bool {
	if card == nil {
		return false
	}
	if top == nil {
		return card.GetValue() == SevenBridgePivotRank
	}
	if card.GetValue() == SevenBridgePivotRank {
		return true
	}
	if card.GetValue() == top.GetValue() {
		return true
	}
	if card.GetValue() == top.GetValue()+1 || card.GetValue() == top.GetValue()-1 {
		return true
	}
	return false
}

// sevenBridgeCardPenalty 手札ペナルティ計算
//   - A = 1, 2-6 / 8-9 = face value, 10 / J / Q / K = 10, 7 = SevenBridgeSevenPenalty
func sevenBridgeCardPenalty(card *Card) int {
	v := card.GetValue()
	if v == SevenBridgePivotRank {
		return SevenBridgeSevenPenalty
	}
	if v == 1 {
		return 1
	}
	if v >= 10 {
		return 10
	}
	return v
}

// validateIndexPair validates a two-element index list. Callers must ensure
// len(indices) == 2 before calling.
func validateIndexPair(indices []int, size int) error {
	return validateIndexList(indices, size)
}

// --- JSON ---

// sevenBridgeJSON is the JSON wire format for SevenBridge.
type sevenBridgeJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*SevenBridgePlayer `json:"pl"`
	Config           SevenBridgeConfig    `json:"cf"`
	Phase            SevenBridgePhase     `json:"ps"`
	CurrentPlayerIdx int                  `json:"ci"`
	DiscardPile      []*Card              `json:"dp"`
	DrawPile         []*Card              `json:"wp"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerIdx        int                  `json:"wi"`
	RoundNumber      int                  `json:"rn"`
	ActionLog        []*ActionLogEntry    `json:"al"`
	RoundWinnerIdx   int                  `json:"rw"`
	ClaimedThisTurn  bool                 `json:"ct"`
}

// MarshalJSON implements json.Marshaler.
func (g *SevenBridge) MarshalJSON() ([]byte, error) {
	return json.Marshal(sevenBridgeJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		RoundWinnerIdx:   g.roundWinnerIdx,
		ClaimedThisTurn:  g.claimedThisTurn,
	})
}

const sevenBridgeMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *SevenBridge) UnmarshalJSON(data []byte) error {
	var j sevenBridgeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > sevenBridgeMaxSliceLen || len(j.DiscardPile) > sevenBridgeMaxSliceLen ||
		len(j.DrawPile) > sevenBridgeMaxSliceLen || len(j.ActionLog) > sevenBridgeMaxSliceLen {
		return fmt.Errorf("sevenbridge: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*SevenBridgePlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.roundWinnerIdx = j.RoundWinnerIdx
	g.claimedThisTurn = j.ClaimedThisTurn
	return nil
}
