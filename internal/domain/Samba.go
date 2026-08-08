//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// SambaPlayerCnt サンバのプレイヤー数 (2チーム × 2人のパートナーシップ)
const SambaPlayerCnt = 4

// SambaTeamCnt サンバのチーム数 (席0・2 = チーム0, 席1・3 = チーム1)
const SambaTeamCnt = 2

// SambaHandSize 初期配布枚数
const SambaHandSize = 15

// SambaDefaultPointLimit デフォルトの目標スコア
const SambaDefaultPointLimit = 10000

// SambaGoOutRequiredMelds 上がりに必要な完成メルド（カナスタ/サンバ）の数
const SambaGoOutRequiredMelds = 2

// SambaRed3Count 3デッキ中の赤3の総数
const SambaRed3Count = 6

// サンバスコア定数
const (
	SambaGoingOutBonus       = 100  // 上がりボーナス
	SambaNaturalCanastaBonus = 500  // ナチュラルカナスタボーナス (セット・ワイルドなし)
	SambaMixedCanastaBonus   = 300  // ミックスカナスタボーナス (セット・ワイルドあり)
	SambaSambaBonus          = 1500 // サンバボーナス (7枚のシーケンス完成)
	SambaRed3Bonus           = 100  // 赤3のボーナス (1枚あたり)
	SambaAllRed3Bonus        = 800  // 赤3全枚ボーナス
)

// SambaPhase ゲームフェーズ
type SambaPhase int

// Sambaのフェーズ定数
const (
	// SambaPhaseDraw ドローフェーズ (山札または捨て札から引く)
	SambaPhaseDraw SambaPhase = 0
	// SambaPhaseMeld メルドフェーズ (メルドを出す/既存メルドに追加する)
	SambaPhaseMeld SambaPhase = 1
	// SambaPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or 上がる)
	SambaPhaseDiscard SambaPhase = 2
	// SambaPhaseRoundEnd ラウンド終了フェーズ
	SambaPhaseRoundEnd SambaPhase = 3
	// SambaPhaseGameEnd ゲーム終了フェーズ
	SambaPhaseGameEnd SambaPhase = 4
)

// Samba サンバゲームクラス。カナスタの派生ゲームで、3デッキ(162枚)を使い、
// 同ランクのセットメルドに加えて同スート連番のシーケンスメルド（サンバ）を作る。
// 4人のパートナーシップ (席0・2 vs 席1・3) でチーム対戦する。
type Samba struct {
	trumpCards       *TrumpCards
	players          []*SambaPlayer
	config           SambaConfig
	phase            SambaPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	isFrozen         bool
	gameEndFlag      bool
	winnerIdx        int // 勝利チームのインデックス (0 or 1), -1 = 未確定
	roundNumber      int
	teamScores       []int // チーム累積スコア (length SambaTeamCnt)
	actionLogBase
	drewFromDiscard bool  // 現在のターンで捨て札の山から引いたか
	drawnCard       *Card // 捨て札の山のトップカード (メルドバリデーション用)
}

// NewSamba コンストラクタ
func NewSamba(trumpCards *TrumpCards, players []*SambaPlayer, config SambaConfig) *Samba {
	return &Samba{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		teamScores:  make([]int, SambaTeamCnt),
	}
}

// newSambaDeck は3デッキ + 6ジョーカー = 162枚のパックを生成する
// (カナスタが NewTrumpCardsWithDecks(2, 4) を使うのを踏襲)。
func newSambaDeck() *TrumpCards {
	return NewTrumpCardsWithDecks(3, 6)
}

// NewDefaultSamba returns Samba with the standard 4-player partnership setup
// (seat 0 human, seats 1-3 CPU; teams 0 & 1) using a 3-deck pack with 6 jokers
// and DefaultSambaConfig. Single source of truth for CUI, Web, and Worker.
func NewDefaultSamba() *Samba {
	players := make([]*SambaPlayer, 0, SambaPlayerCnt)
	for i := 0; i < SambaPlayerCnt; i++ {
		players = append(players, NewSambaPlayer(i == 0, i%SambaTeamCnt))
	}
	return NewSamba(newSambaDeck(), players, DefaultSambaConfig())
}

// Reset ゲーム初期化
func (g *Samba) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.isFrozen = false
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.drewFromDiscard = false
	g.drawnCard = nil
	g.teamScores = make([]int, SambaTeamCnt)

	for i, p := range g.players {
		p.team = i % SambaTeamCnt
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.Reset()
		p.SetIsFinished(false)
		p.melds = make([]*SambaMeld, 0)
		p.red3s = make([]*Card, 0)
		p.hasInitMeld = false
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = SambaPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Samba) NextRound() {
	if g.phase != SambaPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.isFrozen = false
	g.currentPlayerIdx = 0
	g.drewFromDiscard = false
	g.drawnCard = nil

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = SambaPhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに15枚、1枚を捨て札に
func (g *Samba) dealInitialCards() {
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

	for i := 0; i < SambaHandSize; i++ {
		for j := 0; j < SambaPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// 赤3の処理: 手札の赤3を自動的に場に出し、山札から補充
	for j := 0; j < SambaPlayerCnt; j++ {
		g.autoLayRed3s(j)
	}

	// 最初の1枚を捨て札に (赤3やワイルドが出たら次のカードを引く)
	for len(g.drawPile) > 0 {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, card)
		if SambaIsRed3(card) {
			continue
		}
		if SambaIsWild(card) {
			g.isFrozen = true
		}
		break
	}
}

// autoLayRed3s プレイヤーの手札から赤3を自動的に場に出す
func (g *Samba) autoLayRed3s(playerIdx int) {
	player := g.players[playerIdx]
	for {
		found := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if SambaIsRed3(card) {
				player.RemoveCard(i)
				player.AddRed3(card)
				g.appendLog(playerIdx, "red3", fmt.Sprintf("%s lays down red 3: %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})
				if len(g.drawPile) > 0 {
					replacement := g.drawPile[len(g.drawPile)-1]
					g.drawPile = g.drawPile[:len(g.drawPile)-1]
					player.AddCard(replacement)
				}
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
}

// teamCompletedCount チームの完成メルド（7枚以上のカナスタ/サンバ）の合計数
func (g *Samba) teamCompletedCount(team int) int {
	n := 0
	for _, p := range g.players {
		if p.team == team {
			n += p.CompletedMeldCount()
		}
	}
	return n
}

// canGoOut 上がり条件: チームが必要数の完成メルドを持っているか
func (g *Samba) canGoOut(playerIdx int) bool {
	return g.teamCompletedCount(g.players[playerIdx].team) >= SambaGoOutRequiredMelds
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Samba) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SambaPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	if SambaIsRed3(card) {
		g.autoLayRed3s(g.currentPlayerIdx)
		if len(g.drawPile) == 0 && g.players[g.currentPlayerIdx].GetCardsSize() == 0 {
			g.endRoundDraw()
			return nil
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)

	g.phase = SambaPhaseMeld
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札の山を取る。カナスタと同様に
// トップカードと同ランクのナチュラルカードのペアを手札から示す必要がある。
func (g *Samba) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SambaPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札の山が空です")
	}

	topCard := g.discardPile[len(g.discardPile)-1]

	if SambaIsBlack3(topCard) {
		return NewDomainError(ErrInvalidPlay, "黒3がトップの場合は捨て札の山を取れません")
	}
	if SambaIsWild(topCard) {
		return NewDomainError(ErrInvalidPlay, "ワイルドカードがトップの場合は捨て札の山を取れません")
	}

	if len(naturalPairIndices) != 2 {
		return NewDomainError(ErrInvalidPlay, "ナチュラルペア（同ランクの自然カード2枚）のインデックスを指定してください")
	}

	player := g.players[g.currentPlayerIdx]
	for _, idx := range naturalPairIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
	}
	if naturalPairIndices[0] == naturalPairIndices[1] {
		return NewDomainError(ErrInvalidCard, "同じカードは指定できません")
	}

	card0 := player.GetCard(naturalPairIndices[0])
	card1 := player.GetCard(naturalPairIndices[1])

	if SambaIsWild(card0) || SambaIsWild(card1) {
		return NewDomainError(ErrInvalidPlay, "ペアはナチュラルカード（ワイルドカード以外）でなければなりません")
	}
	if card0.GetValue() != topCard.GetValue() || card1.GetValue() != topCard.GetValue() {
		return NewDomainError(ErrInvalidPlay, "ペアのランクが捨て札のトップカードと一致しません")
	}

	if !player.hasInitMeld {
		meldValue := SambaCardValue(topCard) + SambaCardValue(card0) + SambaCardValue(card1)
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if meldValue < minReq {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("初回メルドの最低点(%d)を満たしていません（現在%d点）", minReq, meldValue))
		}
	}

	g.drawnCard = topCard
	g.drewFromDiscard = true

	pileSize := len(g.discardPile)
	for _, c := range g.discardPile {
		player.AddCard(c)
	}
	g.discardPile = nil
	g.isFrozen = false

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s picks up the discard pile (%d cards)", playerName(g.players, g.currentPlayerIdx), pileSize), []*Card{topCard})

	g.autoLayRed3s(g.currentPlayerIdx)
	g.sortHand(g.currentPlayerIdx)

	g.phase = SambaPhaseMeld
	return nil
}

// sambaMeldResolution はメルドグループの解決結果を表す。
type sambaMeldResolution struct {
	isNew       bool
	existingIdx int
	kind        SambaMeldKind
}

// PlayerMeld 人間プレイヤーがメルドを出す
func (g *Samba) PlayerMeld(meldGroups [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SambaPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(meldGroups) == 0 {
		if g.drewFromDiscard {
			return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
		}
		g.phase = SambaPhaseDiscard
		return nil
	}

	player := g.players[g.currentPlayerIdx]

	allIndices := make(map[int]bool)
	for _, group := range meldGroups {
		for _, idx := range group {
			if idx < 0 || idx >= player.GetCardsSize() {
				return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
			}
			if allIndices[idx] {
				return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
			}
			allIndices[idx] = true
		}
	}

	type meldGroupCards struct {
		indices []int
		cards   []*Card
	}
	groups := make([]meldGroupCards, len(meldGroups))
	for i, group := range meldGroups {
		cards := make([]*Card, len(group))
		for j, idx := range group {
			cards[j] = player.GetCard(idx)
		}
		groups[i] = meldGroupCards{indices: group, cards: cards}
	}

	// 捨て札の山を取った場合、トップカードが少なくとも1つのグループに含まれているか確認
	if g.drewFromDiscard && g.drawnCard != nil {
		found := false
		for _, grp := range groups {
			for _, c := range grp.cards {
				if c == g.drawnCard {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
		}
	}

	totalMeldValue := 0
	isInitialMeld := !player.hasInitMeld

	resolutions := make([]sambaMeldResolution, 0, len(groups))
	for _, grp := range groups {
		res, err := g.resolveMeldGroup(g.currentPlayerIdx, grp.cards)
		if err != nil {
			return err
		}
		resolutions = append(resolutions, res)
		if isInitialMeld {
			for _, c := range grp.cards {
				totalMeldValue += SambaCardValue(c)
			}
		}
	}

	if isInitialMeld {
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if totalMeldValue < minReq {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("初回メルドの最低点(%d)を満たしていません（現在%d点）", minReq, totalMeldValue))
		}
	}

	// メルド実行 (先にアクションを適用し、最後に手札を降順で削除)
	for i, grp := range groups {
		g.applyResolvedMeld(g.currentPlayerIdx, resolutions[i], grp.cards)
	}

	allIdx := make([]int, 0, len(allIndices))
	for k := range allIndices {
		allIdx = append(allIdx, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(allIdx)))
	for _, idx := range allIdx {
		player.RemoveCard(idx)
	}

	player.hasInitMeld = true
	g.drewFromDiscard = false
	g.drawnCard = nil

	g.logCompletedMelds(g.currentPlayerIdx)

	if player.GetCardsSize() == 0 && g.canGoOut(g.currentPlayerIdx) {
		g.goOut(g.currentPlayerIdx)
		return nil
	}

	g.phase = SambaPhaseDiscard
	return nil
}

// resolveMeldGroup はカードグループを新規メルドまたは既存メルドへの追加として
// 解決し、そのメルド種別を判定する。ワイルドはシーケンスに使えない。
func (g *Samba) resolveMeldGroup(playerIdx int, cards []*Card) (sambaMeldResolution, error) {
	player := g.players[playerIdx]

	if sambaGroupIsSetShaped(cards) {
		rank := sambaNaturalRank(cards)
		if rank != 0 {
			for i, m := range player.melds {
				if m.Kind == SambaMeldSet && m.GetRank() == rank {
					if err := g.validateSetAddition(m, cards); err != nil {
						return sambaMeldResolution{}, err
					}
					return sambaMeldResolution{isNew: false, existingIdx: i, kind: SambaMeldSet}, nil
				}
			}
			// セットとして出せないが、既存シーケンスを延長できるなら追加とみなす
			for i, m := range player.melds {
				if m.Kind == SambaMeldSequence && g.validateSequenceAddition(m, cards) == nil {
					return sambaMeldResolution{isNew: false, existingIdx: i, kind: SambaMeldSequence}, nil
				}
			}
		}
		if err := g.validateNewSet(cards); err != nil {
			return sambaMeldResolution{}, err
		}
		return sambaMeldResolution{isNew: true, existingIdx: -1, kind: SambaMeldSet}, nil
	}

	// シーケンス形状
	suit := sambaSequenceSuit(cards)
	if suit >= 0 {
		for i, m := range player.melds {
			if m.Kind == SambaMeldSequence && m.SuitDesign() == suit {
				if err := g.validateSequenceAddition(m, cards); err == nil {
					return sambaMeldResolution{isNew: false, existingIdx: i, kind: SambaMeldSequence}, nil
				}
			}
		}
	}
	if err := g.validateNewSequence(cards); err != nil {
		return sambaMeldResolution{}, err
	}
	return sambaMeldResolution{isNew: true, existingIdx: -1, kind: SambaMeldSequence}, nil
}

// applyResolvedMeld は解決済みメルドをプレイヤーの場に反映する（手札削除は呼び出し側）。
func (g *Samba) applyResolvedMeld(playerIdx int, res sambaMeldResolution, cards []*Card) {
	player := g.players[playerIdx]
	if res.isNew {
		isNatural := true
		for _, c := range cards {
			if SambaIsWild(c) {
				isNatural = false
				break
			}
		}
		meld := &SambaMeld{Cards: cards, Kind: res.kind, IsNatural: isNatural}
		player.AddMeld(meld)
		g.appendLog(playerIdx, "meld", fmt.Sprintf("%s melds a %s of %d cards", playerName(g.players, playerIdx), sambaMeldKindStr(res.kind), len(cards)), cards)
		return
	}
	existing := player.melds[res.existingIdx]
	for _, c := range cards {
		existing.Cards = append(existing.Cards, c)
		if SambaIsWild(c) {
			existing.IsNatural = false
		}
	}
	g.appendLog(playerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to a %s meld", playerName(g.players, playerIdx), len(cards), sambaMeldKindStr(existing.Kind)), cards)
}

// logCompletedMelds 完成したカナスタ/サンバをログに記録する
func (g *Samba) logCompletedMelds(playerIdx int) {
	player := g.players[playerIdx]
	for _, m := range player.melds {
		if m.IsSamba() {
			g.appendLog(playerIdx, "samba", fmt.Sprintf("%s completes a samba!", playerName(g.players, playerIdx)), nil)
		} else if m.IsCanasta() {
			g.appendLog(playerIdx, "canasta", fmt.Sprintf("%s completes a %s canasta!", playerName(g.players, playerIdx), sambaCanastaTypeStr(m.IsNatural)), nil)
		}
	}
}

// PlayerSkipMeld メルドフェーズをスキップする
func (g *Samba) PlayerSkipMeld() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SambaPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.drewFromDiscard {
		return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
	}

	g.phase = SambaPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *Samba) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SambaPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if SambaIsRed3(card) {
		return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	if SambaIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	g.advanceTurn()
	return nil
}

// PlayerGoOut 人間プレイヤーが上がる
func (g *Samba) PlayerGoOut() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SambaPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]

	if !g.canGoOut(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("上がるにはチームで%d個以上のカナスタ/サンバが必要です", SambaGoOutRequiredMelds))
	}

	if player.GetCardsSize() == 1 {
		card := player.GetCard(0)
		if SambaIsRed3(card) {
			return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
		}
		discarded := player.RemoveCard(0)
		g.discardPile = append(g.discardPile, discarded)
		if SambaIsWild(discarded) {
			g.isFrozen = true
		}
		g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	} else if player.GetCardsSize() > 1 {
		return NewDomainError(ErrInvalidPlay, "上がるには手札が0枚または1枚でなければなりません")
	}

	g.goOut(g.currentPlayerIdx)
	return nil
}

// goOut 上がり処理
func (g *Samba) goOut(playerIdx int) {
	bonus := SambaGoingOutBonus
	g.appendLog(playerIdx, "go_out", fmt.Sprintf("%s goes out! (bonus: %d)", playerName(g.players, playerIdx), bonus), nil)
	g.scoreRound(playerIdx, bonus)
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *Samba) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case SambaPhaseDraw:
		g.cpuDraw()
	case SambaPhaseMeld:
		g.cpuMeld()
	case SambaPhaseDiscard:
		g.cpuDiscard()
	}
}

// cpuDraw CPUがドローする
func (g *Samba) cpuDraw() {
	player := g.players[g.currentPlayerIdx]

	if len(g.discardPile) > 0 {
		topCard := g.discardPile[len(g.discardPile)-1]
		if !SambaIsBlack3(topCard) && !SambaIsWild(topCard) {
			shouldPickUp := false
			switch g.config.CpuDifficulty {
			case SambaCpuDifficultyHard:
				shouldPickUp = g.cpuHasNaturalPair(player, topCard) && len(g.discardPile) >= 3
			case SambaCpuDifficultyNormal:
				shouldPickUp = g.cpuHasNaturalPair(player, topCard)
			default:
				shouldPickUp = rand.Intn(4) == 0 && g.cpuHasNaturalPair(player, topCard)
			}

			if shouldPickUp {
				pairIndices := g.cpuFindNaturalPair(player, topCard)
				if pairIndices != nil {
					canPickUp := true
					if !player.hasInitMeld {
						meldValue := SambaCardValue(topCard) + SambaCardValue(player.GetCard(pairIndices[0])) + SambaCardValue(player.GetCard(pairIndices[1]))
						minReq := g.minimumMeldValue(g.currentPlayerIdx)
						canPickUp = meldValue >= minReq
					}
					if canPickUp {
						g.drawnCard = topCard
						g.drewFromDiscard = true
						for _, c := range g.discardPile {
							player.AddCard(c)
						}
						pileSize := len(g.discardPile)
						g.discardPile = nil
						g.isFrozen = false
						g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s picks up the discard pile (%d cards)", playerName(g.players, g.currentPlayerIdx), pileSize), []*Card{topCard})
						g.autoLayRed3s(g.currentPlayerIdx)
						g.sortHand(g.currentPlayerIdx)
						g.phase = SambaPhaseMeld
						return
					}
				}
			}
		}
	}

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	player.AddCard(card)
	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	if SambaIsRed3(card) {
		g.autoLayRed3s(g.currentPlayerIdx)
		if len(g.drawPile) == 0 && player.GetCardsSize() == 0 {
			g.endRoundDraw()
			return
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)
	g.phase = SambaPhaseMeld
}

// sambaCpuGroup はCPUのメルド候補（種別と既存メルドへの追加先を含む）。
type sambaCpuGroup struct {
	cards       []*Card
	kind        SambaMeldKind
	existingIdx int // -1 = 新規メルド
}

// cpuMeld CPUがメルドする
func (g *Samba) cpuMeld() {
	player := g.players[g.currentPlayerIdx]

	groups := g.cpuFindMelds(player)
	if len(groups) == 0 {
		g.drewFromDiscard = false
		g.drawnCard = nil
		g.phase = SambaPhaseDiscard
		return
	}

	if !player.hasInitMeld {
		totalValue := 0
		for _, grp := range groups {
			for _, c := range grp.cards {
				totalValue += SambaCardValue(c)
			}
		}
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if totalValue < minReq {
			g.drewFromDiscard = false
			g.drawnCard = nil
			g.phase = SambaPhaseDiscard
			return
		}
	}

	for _, grp := range groups {
		if grp.existingIdx >= 0 && grp.existingIdx < len(player.melds) {
			existing := player.melds[grp.existingIdx]
			for _, c := range grp.cards {
				existing.Cards = append(existing.Cards, c)
				if SambaIsWild(c) {
					existing.IsNatural = false
				}
			}
			g.appendLog(g.currentPlayerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to a %s meld", playerName(g.players, g.currentPlayerIdx), len(grp.cards), sambaMeldKindStr(existing.Kind)), grp.cards)
		} else {
			isNatural := true
			for _, c := range grp.cards {
				if SambaIsWild(c) {
					isNatural = false
					break
				}
			}
			meld := &SambaMeld{Cards: grp.cards, Kind: grp.kind, IsNatural: isNatural}
			player.AddMeld(meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds a %s of %d cards", playerName(g.players, g.currentPlayerIdx), sambaMeldKindStr(grp.kind), len(grp.cards)), grp.cards)
		}

		for _, c := range grp.cards {
			for i := 0; i < player.GetCardsSize(); i++ {
				if player.GetCard(i) == c {
					player.RemoveCard(i)
					break
				}
			}
		}
	}

	player.hasInitMeld = true
	g.drewFromDiscard = false
	g.drawnCard = nil

	g.logCompletedMelds(g.currentPlayerIdx)

	if player.GetCardsSize() == 0 && g.canGoOut(g.currentPlayerIdx) {
		g.goOut(g.currentPlayerIdx)
		return
	}

	g.phase = SambaPhaseDiscard
}

// cpuDiscard CPUがディスカードする
func (g *Samba) cpuDiscard() {
	player := g.players[g.currentPlayerIdx]

	if player.GetCardsSize() == 0 {
		if g.canGoOut(g.currentPlayerIdx) {
			g.goOut(g.currentPlayerIdx)
		} else {
			g.advanceTurn()
		}
		return
	}

	if player.GetCardsSize() == 1 && g.canGoOut(g.currentPlayerIdx) {
		card := player.GetCard(0)
		if !SambaIsRed3(card) {
			discarded := player.RemoveCard(0)
			g.discardPile = append(g.discardPile, discarded)
			if SambaIsWild(discarded) {
				g.isFrozen = true
			}
			g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
			g.goOut(g.currentPlayerIdx)
			return
		}
	}

	bestIdx := g.cpuBestDiscard(player)
	discarded := player.RemoveCard(bestIdx)
	g.discardPile = append(g.discardPile, discarded)

	if SambaIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// --- CPU AI helpers ---

// cpuHasNaturalPair トップカードと同ランクのナチュラルカードのペアがあるか
func (g *Samba) cpuHasNaturalPair(player *SambaPlayer, topCard *Card) bool {
	return g.cpuFindNaturalPair(player, topCard) != nil
}

// cpuFindNaturalPair ナチュラルペアのインデックスを返す (見つからなければnil)
func (g *Samba) cpuFindNaturalPair(player *SambaPlayer, topCard *Card) []int {
	var indices []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !SambaIsWild(c) && c.GetValue() == topCard.GetValue() {
			indices = append(indices, i)
			if len(indices) == 2 {
				return indices
			}
		}
	}
	return nil
}

// cpuFindMelds CPUのメルド候補を見つける。セット（同ランク）とシーケンス
// （同スート連番＝サンバ）の両方を探す。返すグループのカードは互いに重複しない。
func (g *Samba) cpuFindMelds(player *SambaPlayer) []sambaCpuGroup {
	var groups []sambaCpuGroup
	used := make(map[*Card]bool)

	byRank := make(map[int][]*Card)
	var wilds []*Card
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if SambaIsWild(c) {
			wilds = append(wilds, c)
		} else if !SambaIsRed3(c) && !SambaIsBlack3(c) {
			byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
		}
	}

	// 既存セットメルドへの追加
	for idx, m := range player.melds {
		if m.Kind != SambaMeldSet {
			continue
		}
		rank := m.GetRank()
		if cards, ok := byRank[rank]; ok && len(cards) > 0 {
			groups = append(groups, sambaCpuGroup{cards: cards, kind: SambaMeldSet, existingIdx: idx})
			for _, c := range cards {
				used[c] = true
			}
			delete(byRank, rank)
		}
	}

	// 新規セットメルド (同ランク3枚以上、または2枚+ワイルド)
	for _, cards := range byRank {
		if len(cards) >= 3 {
			groups = append(groups, sambaCpuGroup{cards: cards[:3], kind: SambaMeldSet, existingIdx: -1})
			for _, c := range cards[:3] {
				used[c] = true
			}
			if len(cards) > 3 {
				groups = append(groups, sambaCpuGroup{cards: cards[3:], kind: SambaMeldSet, existingIdx: -1})
				for _, c := range cards[3:] {
					used[c] = true
				}
			}
		} else if len(cards) == 2 && len(wilds) > 0 {
			meld := []*Card{cards[0], cards[1], wilds[0]}
			wilds = wilds[1:]
			groups = append(groups, sambaCpuGroup{cards: meld, kind: SambaMeldSet, existingIdx: -1})
			for _, c := range meld {
				used[c] = true
			}
		}
	}

	// シーケンス: 未使用のナチュラルカードをスートごとに集めて連番を探す
	groups = append(groups, g.cpuFindSequenceGroups(player, used)...)

	return groups
}

// cpuFindSequenceGroups は未使用カードから同スート連番のシーケンス候補を探す。
// 既存シーケンスメルドの延長と新規シーケンスの両方を返す。
func (g *Samba) cpuFindSequenceGroups(player *SambaPlayer, used map[*Card]bool) []sambaCpuGroup {
	var groups []sambaCpuGroup

	// スートごとに未使用カードを集める (ワイルド・3は除外)
	bySuit := make(map[int][]*Card)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if used[c] || SambaIsWild(c) || c.GetValue() == 3 {
			continue
		}
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c)
	}

	// 既存シーケンスメルドの延長
	for idx, m := range player.melds {
		if m.Kind != SambaMeldSequence {
			continue
		}
		suit := m.SuitDesign()
		pool := bySuit[suit]
		if len(pool) == 0 {
			continue
		}
		vals := sambaSequenceValues(m.Cards)
		if len(vals) == 0 {
			continue
		}
		sort.Ints(vals)
		low, high := vals[0], vals[len(vals)-1]
		var added []*Card
		remaining := pool[:0:0]
		remaining = append(remaining, pool...)
		extended := true
		for extended {
			extended = false
			for k := 0; k < len(remaining); k++ {
				c := remaining[k]
				if used[c] {
					continue
				}
				v := sambaSequenceValue(c)
				if v == high+1 && v <= 14 {
					added = append(added, c)
					used[c] = true
					high = v
					extended = true
				} else if v == low-1 && v >= 4 {
					added = append(added, c)
					used[c] = true
					low = v
					extended = true
				}
			}
		}
		if len(added) > 0 {
			groups = append(groups, sambaCpuGroup{cards: added, kind: SambaMeldSequence, existingIdx: idx})
		}
		bySuit[suit] = filterUnused(pool, used)
	}

	// 新規シーケンス (未使用カードから長さ3以上の連番を作る)
	for _, pool := range bySuit {
		cards := filterUnused(pool, used)
		sort.Slice(cards, func(i, j int) bool {
			return sambaSequenceValue(cards[i]) < sambaSequenceValue(cards[j])
		})
		run := make([]*Card, 0, len(cards))
		flush := func() {
			if len(run) >= 3 {
				grp := make([]*Card, len(run))
				copy(grp, run)
				groups = append(groups, sambaCpuGroup{cards: grp, kind: SambaMeldSequence, existingIdx: -1})
				for _, c := range grp {
					used[c] = true
				}
			}
			run = run[:0]
		}
		for i, c := range cards {
			if i == 0 {
				run = append(run, c)
				continue
			}
			prev := sambaSequenceValue(cards[i-1])
			cur := sambaSequenceValue(c)
			if cur == prev+1 {
				run = append(run, c)
			} else if cur == prev {
				// 重複ランク（別デッキの同カード）はシーケンスに使えないのでスキップ
				continue
			} else {
				flush()
				run = append(run, c)
			}
		}
		flush()
	}

	return groups
}

// filterUnused はまだ使われていないカードのみを返す。
func filterUnused(cards []*Card, used map[*Card]bool) []*Card {
	out := cards[:0:0]
	for _, c := range cards {
		if !used[c] {
			out = append(out, c)
		}
	}
	return out
}

// cpuBestDiscard CPUが最適なディスカードを選択する
func (g *Samba) cpuBestDiscard(player *SambaPlayer) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if SambaIsBlack3(c) {
			return i
		}
	}

	if g.config.CpuDifficulty == SambaCpuDifficultyEasy {
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if !SambaIsRed3(c) && !SambaIsWild(c) {
				return i
			}
		}
		return 0
	}

	return g.cpuBestDiscardSmart(player)
}

// cpuBestDiscardSmart Normal/Hard: 孤立した低得点カードを捨てる
func (g *Samba) cpuBestDiscardSmart(player *SambaPlayer) int {
	rankCount := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !SambaIsWild(c) && !SambaIsRed3(c) {
			rankCount[c.GetValue()]++
		}
	}

	bestIdx := -1
	bestValue := 1000
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if SambaIsRed3(c) || SambaIsWild(c) {
			continue
		}
		if rankCount[c.GetValue()] == 1 {
			val := SambaCardValue(c)
			if val < bestValue {
				bestValue = val
				bestIdx = i
			}
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}

	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if SambaIsRed3(c) || SambaIsWild(c) {
			continue
		}
		val := SambaCardValue(c)
		if val < bestValue {
			bestValue = val
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	return 0
}

// --- Scoring ---

// scoreRound ラウンドのスコアをチーム単位で確定する
func (g *Samba) scoreRound(goOutPlayerIdx int, goOutBonus int) {
	teamRound := make([]int, SambaTeamCnt)

	for i, player := range g.players {
		score := 0
		for _, m := range player.melds {
			for _, c := range m.Cards {
				score += SambaCardValue(c)
			}
			if m.IsCanasta() {
				if m.IsNatural {
					score += SambaNaturalCanastaBonus
				} else {
					score += SambaMixedCanastaBonus
				}
			}
			if m.IsSamba() {
				score += SambaSambaBonus
			}
		}

		// 赤3はチームが初回メルドを完了している場合のみプラス。ラウンド終了時に
		// チームが一度もメルドしていなければ赤3の点は減算される (サンバ/カナスタの標準ルール)。
		teamMelded := false
		for _, tp := range g.players {
			if tp.team == player.team && tp.hasInitMeld {
				teamMelded = true
				break
			}
		}
		red3Count := len(player.red3s)
		red3Score := red3Count * SambaRed3Bonus
		if red3Count >= SambaRed3Count {
			red3Score = SambaAllRed3Bonus
		}
		if teamMelded {
			score += red3Score
		} else {
			score -= red3Score
		}

		if i == goOutPlayerIdx {
			score += goOutBonus
		}

		for j := 0; j < player.GetCardsSize(); j++ {
			score -= SambaCardValue(player.GetCard(j))
		}

		team := player.team
		if team < 0 || team >= SambaTeamCnt {
			team = i % SambaTeamCnt
		}
		teamRound[team] += score
		g.appendLog(i, "score", fmt.Sprintf("%s contributes %d points to team %d", playerName(g.players, i), score, team), nil)
	}

	for t := 0; t < SambaTeamCnt; t++ {
		g.teamScores[t] += teamRound[t]
	}

	// チーム合計を各プレイヤーの表示スコアに反映する
	for _, player := range g.players {
		team := player.team
		if team < 0 || team >= SambaTeamCnt {
			continue
		}
		player.SetRoundScore(teamRound[team])
		player.SetCumulativeScore(g.teamScores[team])
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = SambaPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによるラウンド終了
func (g *Samba) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends (stock empty)", nil)
	g.scoreRound(-1, 0)
}

// advanceTurn 次のプレイヤーへ
func (g *Samba) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SambaPlayerCnt
	g.drewFromDiscard = false
	g.drawnCard = nil

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	g.phase = SambaPhaseDraw
}

// checkGameEnd ゲーム終了判定 (チームスコアで判定)
func (g *Samba) checkGameEnd() {
	hasWinner := false
	for t := 0; t < SambaTeamCnt; t++ {
		if g.teamScores[t] >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}
	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = SambaPhaseGameEnd

	maxScore := g.teamScores[0]
	g.winnerIdx = 0
	for t := 1; t < SambaTeamCnt; t++ {
		if g.teamScores[t] > maxScore {
			maxScore = g.teamScores[t]
			g.winnerIdx = t
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", g.winnerIdx), nil)
}

// --- Meld Validation ---

// validateNewSet 新規セットメルドの検証
func (g *Samba) validateNewSet(cards []*Card) error {
	if len(cards) < 3 {
		return NewDomainError(ErrInvalidPlay, "メルドには最低3枚のカードが必要です")
	}

	var naturalCount, wildCount int
	rank := 0
	for _, c := range cards {
		if SambaIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
		}
		if SambaIsWild(c) {
			wildCount++
			continue
		}
		naturalCount++
		if rank == 0 {
			rank = c.GetValue()
		} else if c.GetValue() != rank {
			return NewDomainError(ErrInvalidPlay, "セットメルドは同じランクのカードで構成する必要があります")
		}
	}

	if naturalCount < 2 {
		return NewDomainError(ErrInvalidPlay, "メルドにはナチュラルカードが最低2枚必要です")
	}
	if wildCount > 3 {
		return NewDomainError(ErrInvalidPlay, "メルドにはワイルドカードは最大3枚までです")
	}
	if wildCount > naturalCount {
		return NewDomainError(ErrInvalidPlay, "ワイルドカードの数がナチュラルカードの数を超えてはいけません")
	}
	return nil
}

// validateSetAddition 既存セットメルドへの追加の検証
func (g *Samba) validateSetAddition(existing *SambaMeld, cards []*Card) error {
	rank := existing.GetRank()
	wildCount := 0
	for _, c := range existing.Cards {
		if SambaIsWild(c) {
			wildCount++
		}
	}
	for _, c := range cards {
		if SambaIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
		}
		if SambaIsWild(c) {
			wildCount++
		} else if c.GetValue() != rank {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ランク%dのメルドにランク%dのカードは追加できません", rank, c.GetValue()))
		}
	}
	if wildCount > 3 {
		return NewDomainError(ErrInvalidPlay, "メルドにはワイルドカードは最大3枚までです")
	}
	return nil
}

// validateNewSequence 新規シーケンスメルドの検証 (同スート連番、ワイルド不可)
func (g *Samba) validateNewSequence(cards []*Card) error {
	return sambaValidateSequenceCards(cards)
}

// validateSequenceAddition 既存シーケンスメルドへの追加の検証
func (g *Samba) validateSequenceAddition(existing *SambaMeld, cards []*Card) error {
	combined := make([]*Card, 0, len(existing.Cards)+len(cards))
	combined = append(combined, existing.Cards...)
	combined = append(combined, cards...)
	return sambaValidateSequenceCards(combined)
}

// sambaValidateSequenceCards は一連のカードが有効なシーケンス（同スート連番、
// ワイルド・3を含まない、重複なし）かどうかを検証する。スパンを測る前に
// 必ず値をソートする。
func sambaValidateSequenceCards(cards []*Card) error {
	if len(cards) < 3 {
		return NewDomainError(ErrInvalidPlay, "シーケンスには最低3枚のカードが必要です")
	}
	design := -1
	vals := make([]int, 0, len(cards))
	for _, c := range cards {
		if SambaIsWild(c) {
			return NewDomainError(ErrInvalidPlay, "シーケンスメルドにワイルドカードは使えません")
		}
		if c.GetValue() == 3 {
			return NewDomainError(ErrInvalidPlay, "3はシーケンスメルドに使えません")
		}
		if design == -1 {
			design = c.GetDesign()
		} else if c.GetDesign() != design {
			return NewDomainError(ErrInvalidPlay, "シーケンスメルドは同じスートで構成する必要があります")
		}
		vals = append(vals, sambaSequenceValue(c))
	}
	sort.Ints(vals)
	for i := 1; i < len(vals); i++ {
		if vals[i] == vals[i-1] {
			return NewDomainError(ErrInvalidPlay, "シーケンスメルドに同じカードは含められません")
		}
		if vals[i] != vals[i-1]+1 {
			return NewDomainError(ErrInvalidPlay, "シーケンスメルドは連番でなければなりません")
		}
	}
	return nil
}

// minimumMeldValue 初回メルドの最低点を返す (チーム累積スコアに基づく)
func (g *Samba) minimumMeldValue(playerIdx int) int {
	team := g.players[playerIdx].team
	score := 0
	if team >= 0 && team < len(g.teamScores) {
		score = g.teamScores[team]
	}
	switch {
	case score < 0:
		return 15
	case score < 1500:
		return 50
	case score < 3000:
		return 90
	default:
		return 120
	}
}

// --- Card type helpers ---

// SambaIsWild ワイルドカードかどうか (ジョーカーまたは2)
func SambaIsWild(card *Card) bool {
	return card.GetDesign() == CardDesignJoker || card.GetValue() == 2
}

// SambaIsRed3 赤3かどうか (ハート3またはダイヤ3)
func SambaIsRed3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignHeart || card.GetDesign() == CardDesignDiamond)
}

// SambaIsBlack3 黒3かどうか (スペード3またはクローバー3)
func SambaIsBlack3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover)
}

// SambaCardValue カードの点数を返す
func SambaCardValue(card *Card) int {
	if card.GetDesign() == CardDesignJoker {
		return 50
	}
	v := card.GetValue()
	if v == 2 {
		return 20
	}
	if v == 1 { // Ace
		return 20
	}
	if v == 3 && (card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover) {
		return 5 // 黒3
	}
	if v >= 8 {
		return 10
	}
	return 5
}

// sambaSequenceValue はシーケンス判定用のカード値を返す。エースは高位(14)扱いで
// ラップアラウンドはしない。
func sambaSequenceValue(card *Card) int {
	if card.GetValue() == 1 {
		return 14
	}
	return card.GetValue()
}

// sambaSequenceValues はカード列のシーケンス値のスライスを返す。
func sambaSequenceValues(cards []*Card) []int {
	out := make([]int, 0, len(cards))
	for _, c := range cards {
		if SambaIsWild(c) {
			continue
		}
		out = append(out, sambaSequenceValue(c))
	}
	return out
}

// sambaGroupIsSetShaped はグループの全ナチュラルカードが同ランクかどうかを返す。
func sambaGroupIsSetShaped(cards []*Card) bool {
	rank := 0
	for _, c := range cards {
		if SambaIsWild(c) {
			continue
		}
		if rank == 0 {
			rank = c.GetValue()
		} else if c.GetValue() != rank {
			return false
		}
	}
	return true
}

// sambaNaturalRank はグループ内の最初のナチュラルカードのランクを返す (なければ0)。
func sambaNaturalRank(cards []*Card) int {
	for _, c := range cards {
		if !SambaIsWild(c) {
			return c.GetValue()
		}
	}
	return 0
}

// sambaSequenceSuit はシーケンス形状グループのスートを返す (なければ-1)。
func sambaSequenceSuit(cards []*Card) int {
	for _, c := range cards {
		if !SambaIsWild(c) {
			return c.GetDesign()
		}
	}
	return -1
}

// sambaMeldKindStr はメルド種別の文字列を返す。
func sambaMeldKindStr(kind SambaMeldKind) string {
	if kind == SambaMeldSequence {
		return "sequence"
	}
	return "set"
}

// sambaCanastaTypeStr はカナスタの種別文字列を返す。
func sambaCanastaTypeStr(isNatural bool) string {
	if isNatural {
		return "natural"
	}
	return "mixed"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Samba) GetPhase() SambaPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Samba) SetPhase(phase SambaPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Samba) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Samba) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Samba) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Samba) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Samba) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Samba) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Samba) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Samba) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Samba) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetDiscardPileCount 捨て札の枚数取得
func (g *Samba) GetDiscardPileCount() int { return len(g.discardPile) }

// GetIsFrozen 捨て札の山がフリーズ状態か取得
func (g *Samba) GetIsFrozen() bool { return g.isFrozen }

// SetIsFrozen フリーズ状態設定 (テスト用)
func (g *Samba) SetIsFrozen(v bool) { g.isFrozen = v }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Samba) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (g *Samba) SetGameEndFlag(v bool) { g.gameEndFlag = v }

// GetWinnerIdx 勝利チームインデックス取得 (-1 = 未確定)
func (g *Samba) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Samba) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Samba) GetPlayer(i int) *SambaPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetTeamCount チーム数取得
func (g *Samba) GetTeamCount() int { return SambaTeamCnt }

// GetTeamScore チームの累積スコアを取得
func (g *Samba) GetTeamScore(team int) int {
	if team < 0 || team >= len(g.teamScores) {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームの累積スコアを設定 (テスト用)
func (g *Samba) SetTeamScore(team, score int) {
	if team >= 0 && team < len(g.teamScores) {
		g.teamScores[team] = score
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Samba) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Samba) GetConfig() SambaConfig { return g.config }

// SetConfig 設定変更
func (g *Samba) SetConfig(cfg SambaConfig) { g.config = cfg }

// GetDrewFromDiscard 捨て札から引いたか取得
func (g *Samba) GetDrewFromDiscard() bool { return g.drewFromDiscard }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Samba) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Samba) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- JSON serialization ---

// sambaJSON is the JSON wire format for Samba.
type sambaJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*SambaPlayer    `json:"pl"`
	Config           SambaConfig       `json:"cf"`
	Phase            SambaPhase        `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	IsFrozen         bool              `json:"fr"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	TeamScores       []int             `json:"ts"`
	ActionLog        []*ActionLogEntry `json:"al"`
	DrewFromDiscard  bool              `json:"dd"`
}

// MarshalJSON implements json.Marshaler.
func (g *Samba) MarshalJSON() ([]byte, error) {
	return json.Marshal(sambaJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		IsFrozen:         g.isFrozen,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		TeamScores:       g.teamScores,
		ActionLog:        g.actionLog,
		DrewFromDiscard:  g.drewFromDiscard,
	})
}

// sambaMaxSliceLen caps slice sizes during deserialisation.
const sambaMaxSliceLen = 3000

// UnmarshalJSON implements json.Unmarshaler with defensive validation of all
// indices, team values, phase and meld kinds so a corrupt KV blob cannot panic
// scoring (which indexes teamScores by a restored player's team) or a getter.
func (g *Samba) UnmarshalJSON(data []byte) error {
	var j sambaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > sambaMaxSliceLen || len(j.DiscardPile) > sambaMaxSliceLen ||
		len(j.DrawPile) > sambaMaxSliceLen || len(j.ActionLog) > sambaMaxSliceLen {
		return fmt.Errorf("samba: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newSambaDeck()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*SambaPlayer, 0)
	}
	for i, p := range g.players {
		if p == nil {
			return fmt.Errorf("samba: player %d is nil", i)
		}
	}
	g.config = j.Config

	// フェーズは既知の値のみ許可する
	if j.Phase < SambaPhaseDraw || j.Phase > SambaPhaseGameEnd {
		g.phase = SambaPhaseDraw
	} else {
		g.phase = j.Phase
	}

	// チームスコアは常に長さ SambaTeamCnt に正規化する
	g.teamScores = make([]int, SambaTeamCnt)
	for t := 0; t < SambaTeamCnt && t < len(j.TeamScores); t++ {
		g.teamScores[t] = j.TeamScores[t]
	}

	// 各プレイヤーのチームは [0, SambaTeamCnt) に収める。teamScores を
	// インデックスするため、範囲外の値は席順から導出し直す (security)。
	for i, p := range g.players {
		if p == nil {
			continue
		}
		if p.team < 0 || p.team >= SambaTeamCnt {
			p.SetTeam(i % SambaTeamCnt)
		}
	}

	// currentPlayerIdx はプレイヤー数の範囲に収める
	if len(g.players) > 0 && (j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= len(g.players)) {
		g.currentPlayerIdx = 0
	} else {
		g.currentPlayerIdx = j.CurrentPlayerIdx
	}

	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.isFrozen = j.IsFrozen
	g.gameEndFlag = j.GameEndFlag

	// winnerIdx は -1 (未確定) または [0, SambaTeamCnt) のみ許可する
	if j.WinnerIdx < 0 || j.WinnerIdx >= SambaTeamCnt {
		g.winnerIdx = -1
	} else {
		g.winnerIdx = j.WinnerIdx
	}

	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.drewFromDiscard = j.DrewFromDiscard
	return nil
}
