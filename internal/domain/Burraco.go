package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// BurracoPlayerCnt ブラーコプレイヤー数
const BurracoPlayerCnt = 2

// BurracoHandSize 初期配布枚数 (2人制ブラーコ)
const BurracoHandSize = 11

// BurracoPozzettoSize ポゼット（予備手札）1山の枚数
const BurracoPozzettoSize = 11

// BurracoDefaultPointLimit デフォルトの目標スコア
const BurracoDefaultPointLimit = 2005

// ブラーコスコア定数
const (
	BurracoGoingOutBonus          = 100 // 上がりボーナス
	BurracoConcealedGoingOutBonus = 200 // コンシールド上がりボーナス (一度もメルドせずに上がる)
	BurracoNaturalBurracoBonus    = 200 // プーロ（清い）ブラーコボーナス (ワイルドなし)
	BurracoMixedBurracoBonus      = 100 // スポルコ（汚い）ブラーコボーナス (ワイルドあり)
	BurracoRed3Bonus              = 100 // 赤3のボーナス (1枚あたり)
	BurracoAllRed3Bonus           = 800 // 赤3全4枚ボーナス
)

// BurracoPhase ゲームフェーズ
type BurracoPhase int

// Burracoのフェーズ定数
const (
	// BurracoPhaseDraw ドローフェーズ (山札または捨て札から引く)
	BurracoPhaseDraw BurracoPhase = 0
	// BurracoPhaseMeld メルドフェーズ (メルドを出す/既存メルドに追加する)
	BurracoPhaseMeld BurracoPhase = 1
	// BurracoPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or 上がる)
	BurracoPhaseDiscard BurracoPhase = 2
	// BurracoPhaseRoundEnd ラウンド終了フェーズ
	BurracoPhaseRoundEnd BurracoPhase = 3
	// BurracoPhaseGameEnd ゲーム終了フェーズ
	BurracoPhaseGameEnd BurracoPhase = 4
)

// Burraco ブラーコゲームクラス
type Burraco struct {
	trumpCards       *TrumpCards
	players          []*BurracoPlayer
	config           BurracoConfig
	phase            BurracoPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	pozzetti         [][]*Card // 予備手札（ポゼット）の山。各11枚。最初に手札を出し切ったプレイヤーが1山獲得する
	isFrozen         bool      // 捨て札の山がフリーズ状態か
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLog        []*ActionLogEntry
	drewFromDiscard  bool  // 現在のターンで捨て札の山から引いたか
	drawnCard        *Card // 捨て札の山のトップカード (メルドバリデーション用)
}

// NewBurraco コンストラクタ
func NewBurraco(trumpCards *TrumpCards, players []*BurracoPlayer, config BurracoConfig) *Burraco {
	return &Burraco{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultBurraco returns Burraco with the standard 2-player setup (1 human, 1 CPU)
// using a 2-deck pack with 4 jokers and DefaultBurracoConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultBurraco() *Burraco {
	players := []*BurracoPlayer{
		NewBurracoPlayer(true),
		NewBurracoPlayer(false),
	}
	return NewBurraco(NewTrumpCardsWithDecks(2, 4), players, DefaultBurracoConfig())
}

// Reset ゲーム初期化
func (g *Burraco) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.pozzetti = nil
	g.isFrozen = false
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.drewFromDiscard = false
	g.drawnCard = nil

	for _, p := range g.players {
		p.roundScore = 0
		p.cumulativeScore = 0
		p.Reset()
		p.SetIsFinished(false)
		p.melds = make([]*BurracoMeld, 0)
		p.red3s = make([]*Card, 0)
		p.hasInitMeld = false
		p.tookPozzetto = false
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = BurracoPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Burraco) NextRound() {
	if g.phase != BurracoPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.pozzetti = nil
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

	g.phase = BurracoPhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに15枚、1枚を捨て札に
func (g *Burraco) dealInitialCards() {
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

	// 各プレイヤーに11枚配布
	for i := 0; i < BurracoHandSize; i++ {
		for j := 0; j < BurracoPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// ポゼット（予備手札）を2山、それぞれ11枚ずつ脇に取り分ける
	g.pozzetti = make([][]*Card, 0, BurracoPlayerCnt)
	for p := 0; p < BurracoPlayerCnt; p++ {
		pile := make([]*Card, 0, BurracoPozzettoSize)
		for i := 0; i < BurracoPozzettoSize && len(g.drawPile) > 0; i++ {
			card := g.drawPile[len(g.drawPile)-1]
			g.drawPile = g.drawPile[:len(g.drawPile)-1]
			pile = append(pile, card)
		}
		g.pozzetti = append(g.pozzetti, pile)
	}

	// 赤3の処理: 手札の赤3を自動的に場に出し、山札から補充
	for j := 0; j < BurracoPlayerCnt; j++ {
		g.autoLayRed3s(j)
	}

	// 最初の1枚を捨て札に (赤3やワイルドが出たら次のカードを引く)
	for len(g.drawPile) > 0 {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, card)
		if BurracoIsRed3(card) {
			// 赤3は捨て札にできない → 次のカードを引く
			continue
		}
		if BurracoIsWild(card) {
			g.isFrozen = true
		}
		break
	}
}

// autoLayRed3s プレイヤーの手札から赤3を自動的に場に出す
func (g *Burraco) autoLayRed3s(playerIdx int) {
	player := g.players[playerIdx]
	for {
		found := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if BurracoIsRed3(card) {
				player.RemoveCard(i)
				player.AddRed3(card)
				g.appendLog(playerIdx, "red3", fmt.Sprintf("%s lays down red 3: %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})
				// 山札から補充
				if len(g.drawPile) > 0 {
					replacement := g.drawPile[len(g.drawPile)-1]
					g.drawPile = g.drawPile[:len(g.drawPile)-1]
					player.AddCard(replacement)
				}
				found = true
				break // インデックスがずれるので最初からやり直す
			}
		}
		if !found {
			break
		}
	}
}

// takePozzetto プレイヤーが手札を出し切ったとき、ポゼット（予備手札）を1山獲得して手札に加える。
// 取得した場合 true を返す（その場合プレイヤーはまだ上がれない）。
func (g *Burraco) takePozzetto(playerIdx int) bool {
	player := g.players[playerIdx]
	if player.tookPozzetto || len(g.pozzetti) == 0 {
		return false
	}
	pile := g.pozzetti[len(g.pozzetti)-1]
	g.pozzetti = g.pozzetti[:len(g.pozzetti)-1]
	for _, c := range pile {
		player.AddCard(c)
	}
	player.tookPozzetto = true
	g.appendLog(playerIdx, "pozzetto", fmt.Sprintf("%s takes the pozzetto (%d cards)", g.playerName(playerIdx), len(pile)), nil)
	// 獲得した手札に赤3があれば自動的に場に出す
	g.autoLayRed3s(playerIdx)
	g.sortHand(playerIdx)
	return true
}

// canGoOut 上がり条件を満たすか（ポゼット獲得済み かつ ブラーコ完成済み）
func (g *Burraco) canGoOut(playerIdx int) bool {
	p := g.players[playerIdx]
	return p.tookPozzetto && p.HasBurraco()
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Burraco) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BurracoPhaseDraw {
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

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", g.playerName(g.currentPlayerIdx)), nil)

	// 赤3を引いた場合は自動的に場に出す (autoLayRed3s が補充まで処理する)
	if BurracoIsRed3(card) {
		g.autoLayRed3s(g.currentPlayerIdx)
		// autoLayRed3s で補充した結果、山札が空になった場合
		if len(g.drawPile) == 0 && g.players[g.currentPlayerIdx].GetCardsSize() == 0 {
			g.endRoundDraw()
			return nil
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)

	g.phase = BurracoPhaseMeld
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札の山を取る
func (g *Burraco) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BurracoPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札の山が空です")
	}

	topCard := g.discardPile[len(g.discardPile)-1]

	// 黒3がトップの場合は取れない
	if BurracoIsBlack3(topCard) {
		return NewDomainError(ErrInvalidPlay, "黒3がトップの場合は捨て札の山を取れません")
	}

	// ワイルドカードがトップの場合は取れない（通常はワイルドカードがトップにくると山がフリーズ）
	if BurracoIsWild(topCard) {
		return NewDomainError(ErrInvalidPlay, "ワイルドカードがトップの場合は捨て札の山を取れません")
	}

	// ナチュラルペアの検証
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

	// ペアは両方ナチュラルカードで、トップカードと同ランクでなければならない
	if BurracoIsWild(card0) || BurracoIsWild(card1) {
		return NewDomainError(ErrInvalidPlay, "ペアはナチュラルカード（ワイルドカード以外）でなければなりません")
	}
	if card0.GetValue() != topCard.GetValue() || card1.GetValue() != topCard.GetValue() {
		return NewDomainError(ErrInvalidPlay, "ペアのランクが捨て札のトップカードと一致しません")
	}

	// 初回メルド要件チェック: 捨て札の山を取る場合、初回メルドの最低点を満たすメルドが必要
	if !player.hasInitMeld {
		// トップカード + ペアで最低3枚のメルドが作れる: その点数をチェック
		meldValue := BurracoCardValue(topCard) + BurracoCardValue(card0) + BurracoCardValue(card1)
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if meldValue < minReq {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("初回メルドの最低点(%d)を満たしていません（現在%d点）", minReq, meldValue))
		}
	}

	// 捨て札の山を取る
	g.drawnCard = topCard
	g.drewFromDiscard = true

	// 捨て札の山のすべてのカードを手札に追加
	pileSize := len(g.discardPile)
	for _, c := range g.discardPile {
		player.AddCard(c)
	}
	g.discardPile = nil
	g.isFrozen = false

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s picks up the discard pile (%d cards)", g.playerName(g.currentPlayerIdx), pileSize), []*Card{topCard})

	// 手札に赤3があれば自動的に場に出す
	g.autoLayRed3s(g.currentPlayerIdx)
	g.sortHand(g.currentPlayerIdx)

	g.phase = BurracoPhaseMeld
	return nil
}

// PlayerMeld 人間プレイヤーがメルドを出す
func (g *Burraco) PlayerMeld(meldGroups [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BurracoPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(meldGroups) == 0 {
		// メルドなしでディスカードフェーズへ
		if g.drewFromDiscard {
			return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
		}
		g.phase = BurracoPhaseDiscard
		return nil
	}

	player := g.players[g.currentPlayerIdx]

	// すべてのインデックスを検証
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

	// 各グループのカードを取得
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

	// 初回メルドの場合、最低点を計算
	totalMeldValue := 0
	isInitialMeld := !player.hasInitMeld

	// 各グループを新規メルドまたは既存メルドへの追加として検証
	type meldAction struct {
		isNewMeld   bool
		existingIdx int // 既存メルドへの追加の場合のインデックス
		cards       []*Card
		cardIndices []int
	}
	actions := make([]meldAction, 0, len(groups))

	for _, grp := range groups {
		// 既存のメルドへの追加か、新規メルドかを判定
		existingIdx := g.findExistingMeldForCards(g.currentPlayerIdx, grp.cards)

		if existingIdx >= 0 {
			// 既存メルドへの追加
			existing := player.melds[existingIdx]
			if err := g.validateMeldAddition(existing, grp.cards); err != nil {
				return err
			}
			actions = append(actions, meldAction{
				isNewMeld:   false,
				existingIdx: existingIdx,
				cards:       grp.cards,
				cardIndices: grp.indices,
			})
		} else {
			// 新規メルド
			if err := g.validateNewMeld(grp.cards); err != nil {
				return err
			}
			actions = append(actions, meldAction{
				isNewMeld:   true,
				existingIdx: -1,
				cards:       grp.cards,
				cardIndices: grp.indices,
			})
		}

		if isInitialMeld {
			for _, c := range grp.cards {
				totalMeldValue += BurracoCardValue(c)
			}
		}
	}

	// 初回メルドの最低点チェック
	if isInitialMeld {
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if totalMeldValue < minReq {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("初回メルドの最低点(%d)を満たしていません（現在%d点）", minReq, totalMeldValue))
		}
	}

	// メルド実行 (インデックスを降順にソートして削除)
	allIdx := make([]int, 0, len(allIndices))
	for k := range allIndices {
		allIdx = append(allIdx, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(allIdx)))

	// アクション実行前にカードのポインタを保持
	for _, action := range actions {
		if action.isNewMeld {
			isNatural := true
			for _, c := range action.cards {
				if BurracoIsWild(c) {
					isNatural = false
					break
				}
			}
			meld := &BurracoMeld{
				Cards:     action.cards,
				IsNatural: isNatural,
			}
			player.AddMeld(meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d cards (rank %d)", g.playerName(g.currentPlayerIdx), len(action.cards), meld.GetRank()), action.cards)
		} else {
			existing := player.melds[action.existingIdx]
			for _, c := range action.cards {
				existing.Cards = append(existing.Cards, c)
				if BurracoIsWild(c) {
					existing.IsNatural = false
				}
			}
			g.appendLog(g.currentPlayerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to meld (rank %d)", g.playerName(g.currentPlayerIdx), len(action.cards), existing.GetRank()), action.cards)
		}
	}

	// 手札からカードを削除
	for _, idx := range allIdx {
		player.RemoveCard(idx)
	}

	player.hasInitMeld = true
	g.drewFromDiscard = false
	g.drawnCard = nil

	// ブラーコ完成チェック
	for _, m := range player.melds {
		if m.IsBurraco() {
			g.appendLog(g.currentPlayerIdx, "burraco", fmt.Sprintf("%s completes a %s burraco!", g.playerName(g.currentPlayerIdx), burracoTypeStr(m.IsNatural)), nil)
		}
	}

	// 手札を出し切った場合の処理
	if player.GetCardsSize() == 0 {
		if !player.tookPozzetto {
			// 初めて手札を出し切った → ポゼットを獲得して継続（まだ上がれない）
			g.takePozzetto(g.currentPlayerIdx)
		} else if player.HasBurraco() {
			// ポゼット獲得済み かつ ブラーコ完成 → 上がり
			g.goOut(g.currentPlayerIdx, false)
			return nil
		}
	}

	g.phase = BurracoPhaseDiscard
	return nil
}

// PlayerSkipMeld メルドフェーズをスキップする
func (g *Burraco) PlayerSkipMeld() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BurracoPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.drewFromDiscard {
		return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
	}

	g.phase = BurracoPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *Burraco) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BurracoPhaseDiscard {
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

	// 赤3は捨てられない
	if BurracoIsRed3(card) {
		return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	// ワイルドカードを捨てた場合、山をフリーズ
	if BurracoIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", g.playerName(g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	// 捨て札で手札を出し切り、まだポゼット未獲得なら獲得する
	if player.GetCardsSize() == 0 && !player.tookPozzetto {
		g.takePozzetto(g.currentPlayerIdx)
	}

	g.advanceTurn()
	return nil
}

// PlayerGoOut 人間プレイヤーが上がる
func (g *Burraco) PlayerGoOut() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BurracoPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]

	if !player.tookPozzetto {
		return NewDomainError(ErrInvalidPlay, "上がるにはポゼット（予備手札）を獲得している必要があります")
	}
	if !player.HasBurraco() {
		return NewDomainError(ErrInvalidPlay, "上がるには少なくとも1つのブラーコが必要です")
	}

	// 手札が1枚の場合、最後のカードを捨てて上がる
	if player.GetCardsSize() == 1 {
		card := player.GetCard(0)
		if BurracoIsRed3(card) {
			return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
		}
		discarded := player.RemoveCard(0)
		g.discardPile = append(g.discardPile, discarded)
		if BurracoIsWild(discarded) {
			g.isFrozen = true
		}
		g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", g.playerName(g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	} else if player.GetCardsSize() > 1 {
		return NewDomainError(ErrInvalidPlay, "上がるには手札が0枚または1枚でなければなりません")
	}

	concealed := !player.hasInitMeld
	g.goOut(g.currentPlayerIdx, concealed)
	return nil
}

// goOut 上がり処理
func (g *Burraco) goOut(playerIdx int, concealed bool) {
	bonus := BurracoGoingOutBonus
	if concealed {
		bonus = BurracoConcealedGoingOutBonus
	}
	g.appendLog(playerIdx, "go_out", fmt.Sprintf("%s goes out! (bonus: %d)", g.playerName(playerIdx), bonus), nil)
	g.scoreRound(playerIdx, bonus)
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *Burraco) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case BurracoPhaseDraw:
		g.cpuDraw()
	case BurracoPhaseMeld:
		g.cpuMeld()
	case BurracoPhaseDiscard:
		g.cpuDiscard()
	}
}

// cpuDraw CPUがドローする
func (g *Burraco) cpuDraw() {
	player := g.players[g.currentPlayerIdx]

	// 捨て札の山を取るか判定
	if len(g.discardPile) > 0 {
		topCard := g.discardPile[len(g.discardPile)-1]
		if !BurracoIsBlack3(topCard) && !BurracoIsWild(topCard) {
			shouldPickUp := false
			switch g.config.CpuDifficulty {
			case BurracoCpuDifficultyHard:
				shouldPickUp = g.cpuShouldPickUpPileHard(player, topCard)
			case BurracoCpuDifficultyNormal:
				shouldPickUp = g.cpuShouldPickUpPileNormal(player, topCard)
			default:
				// Easy: ランダム
				shouldPickUp = rand.Intn(4) == 0 && g.cpuHasNaturalPair(player, topCard)
			}

			if shouldPickUp {
				pairIndices := g.cpuFindNaturalPair(player, topCard)
				if pairIndices != nil {
					// 初回メルド要件チェック
					canPickUp := true
					if !player.hasInitMeld {
						meldValue := BurracoCardValue(topCard) + BurracoCardValue(player.GetCard(pairIndices[0])) + BurracoCardValue(player.GetCard(pairIndices[1]))
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
						g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s picks up the discard pile (%d cards)", g.playerName(g.currentPlayerIdx), pileSize), []*Card{topCard})
						g.autoLayRed3s(g.currentPlayerIdx)
						g.sortHand(g.currentPlayerIdx)
						g.phase = BurracoPhaseMeld
						return
					}
				}
			}
		}
	}

	// 山札から引く
	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	player.AddCard(card)
	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", g.playerName(g.currentPlayerIdx)), nil)

	// 赤3の処理 (autoLayRed3s が補充まで処理する)
	if BurracoIsRed3(card) {
		g.autoLayRed3s(g.currentPlayerIdx)
		if len(g.drawPile) == 0 && player.GetCardsSize() == 0 {
			g.endRoundDraw()
			return
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)
	g.phase = BurracoPhaseMeld
}

// cpuMeld CPUがメルドする
func (g *Burraco) cpuMeld() {
	player := g.players[g.currentPlayerIdx]

	meldGroups := g.cpuFindMelds(player)
	if len(meldGroups) == 0 {
		g.drewFromDiscard = false
		g.drawnCard = nil
		g.phase = BurracoPhaseDiscard
		return
	}

	// 初回メルドの最低点チェック
	if !player.hasInitMeld {
		totalValue := 0
		for _, group := range meldGroups {
			for _, c := range group {
				totalValue += BurracoCardValue(c)
			}
		}
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if totalValue < minReq {
			g.drewFromDiscard = false
			g.drawnCard = nil
			g.phase = BurracoPhaseDiscard
			return
		}
	}

	// メルド実行
	for _, group := range meldGroups {
		existingIdx := g.findExistingMeldForCards(g.currentPlayerIdx, group)
		if existingIdx >= 0 {
			existing := player.melds[existingIdx]
			for _, c := range group {
				existing.Cards = append(existing.Cards, c)
				if BurracoIsWild(c) {
					existing.IsNatural = false
				}
			}
			g.appendLog(g.currentPlayerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to meld (rank %d)", g.playerName(g.currentPlayerIdx), len(group), existing.GetRank()), group)
		} else {
			isNatural := true
			for _, c := range group {
				if BurracoIsWild(c) {
					isNatural = false
					break
				}
			}
			meld := &BurracoMeld{Cards: group, IsNatural: isNatural}
			player.AddMeld(meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d cards (rank %d)", g.playerName(g.currentPlayerIdx), len(group), meld.GetRank()), group)
		}

		// メルドに使ったカードを手札から削除
		for _, c := range group {
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

	// 手札を出し切った場合の処理
	if player.GetCardsSize() == 0 {
		if !player.tookPozzetto {
			g.takePozzetto(g.currentPlayerIdx)
		} else if player.HasBurraco() {
			g.goOut(g.currentPlayerIdx, false)
			return
		}
	}

	g.phase = BurracoPhaseDiscard
}

// cpuDiscard CPUがディスカードする
func (g *Burraco) cpuDiscard() {
	player := g.players[g.currentPlayerIdx]

	// 手札が空でこのフェーズに来た場合（直前のメルドで出し切った）の処理。
	// ポゼット未獲得なら獲得して手札11枚で継続。獲得済みでブラーコ未完成なら
	// 捨てられないのでターンを進める（RemoveCard → nil のパニックを避ける）。
	if player.GetCardsSize() == 0 {
		if !player.tookPozzetto {
			g.takePozzetto(g.currentPlayerIdx)
		} else {
			g.advanceTurn()
			return
		}
	}

	// 上がれるかチェック
	if player.GetCardsSize() == 1 && g.canGoOut(g.currentPlayerIdx) {
		card := player.GetCard(0)
		if !BurracoIsRed3(card) {
			discarded := player.RemoveCard(0)
			g.discardPile = append(g.discardPile, discarded)
			if BurracoIsWild(discarded) {
				g.isFrozen = true
			}
			g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", g.playerName(g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
			g.goOut(g.currentPlayerIdx, false)
			return
		}
	}

	// 最適なディスカードを選択
	bestIdx := g.cpuBestDiscard(player)
	discarded := player.RemoveCard(bestIdx)
	g.discardPile = append(g.discardPile, discarded)

	if BurracoIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", g.playerName(g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// --- CPU AI helpers ---

// cpuShouldPickUpPileNormal Normal難易度: トップカードとペアがあれば取る
func (g *Burraco) cpuShouldPickUpPileNormal(player *BurracoPlayer, topCard *Card) bool {
	return g.cpuHasNaturalPair(player, topCard)
}

// cpuShouldPickUpPileHard Hard難易度: 戦略的に山を取る
func (g *Burraco) cpuShouldPickUpPileHard(player *BurracoPlayer, topCard *Card) bool {
	if !g.cpuHasNaturalPair(player, topCard) {
		return false
	}
	// 山が大きいほど取る価値がある
	return len(g.discardPile) >= 3
}

// cpuHasNaturalPair トップカードと同ランクのナチュラルカードのペアがあるか
func (g *Burraco) cpuHasNaturalPair(player *BurracoPlayer, topCard *Card) bool {
	return g.cpuFindNaturalPair(player, topCard) != nil
}

// cpuFindNaturalPair ナチュラルペアのインデックスを返す (見つからなければnil)
func (g *Burraco) cpuFindNaturalPair(player *BurracoPlayer, topCard *Card) []int {
	var indices []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !BurracoIsWild(c) && c.GetValue() == topCard.GetValue() {
			indices = append(indices, i)
			if len(indices) == 2 {
				return indices
			}
		}
	}
	return nil
}

// cpuFindMelds CPUのメルド候補を見つける
func (g *Burraco) cpuFindMelds(player *BurracoPlayer) [][]*Card {
	var melds [][]*Card

	// 手札をランクごとにグループ化
	byRank := make(map[int][]*Card)
	var wilds []*Card
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if BurracoIsWild(c) {
			wilds = append(wilds, c)
		} else if !BurracoIsRed3(c) && !BurracoIsBlack3(c) {
			byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
		}
	}

	// 既存メルドへの追加
	for _, m := range player.melds {
		rank := m.GetRank()
		if cards, ok := byRank[rank]; ok && len(cards) > 0 {
			melds = append(melds, cards)
			delete(byRank, rank)
		}
	}

	// 新規メルド: 同ランク3枚以上
	for _, cards := range byRank {
		if len(cards) >= 3 {
			melds = append(melds, cards[:3])
			// 残りは既存メルドへの追加として
			if len(cards) > 3 {
				melds = append(melds, cards[3:])
			}
		} else if len(cards) == 2 && len(wilds) > 0 {
			// ナチュラル2枚 + ワイルド1枚
			meld := []*Card{cards[0], cards[1], wilds[0]}
			wilds = wilds[1:]
			melds = append(melds, meld)
		}
	}

	return melds
}

// cpuBestDiscard CPUが最適なディスカードを選択する
func (g *Burraco) cpuBestDiscard(player *BurracoPlayer) int {
	// 黒3を優先的に捨てる（相手のピックアップをブロック）
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if BurracoIsBlack3(c) {
			return i
		}
	}

	switch g.config.CpuDifficulty {
	case BurracoCpuDifficultyHard:
		return g.cpuBestDiscardHard(player)
	case BurracoCpuDifficultyNormal:
		return g.cpuBestDiscardNormal(player)
	default:
		// Easy: ランダム (赤3とワイルド以外)
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if !BurracoIsRed3(c) && !BurracoIsWild(c) {
				return i
			}
		}
		return 0
	}
}

// cpuBestDiscardNormal Normal: 孤立カードを捨てる
func (g *Burraco) cpuBestDiscardNormal(player *BurracoPlayer) int {
	// ランクごとのカウントを計算
	rankCount := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !BurracoIsWild(c) && !BurracoIsRed3(c) {
			rankCount[c.GetValue()]++
		}
	}

	// 孤立カード (カウント1) のうち点数が低いものを捨てる
	bestIdx := -1
	bestValue := 1000
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if BurracoIsRed3(c) || BurracoIsWild(c) {
			continue
		}
		cnt := rankCount[c.GetValue()]
		val := BurracoCardValue(c)
		if cnt == 1 && val < bestValue {
			bestValue = val
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}

	// 孤立カードがない場合は最も点数が低い非ワイルドカードを捨てる
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if BurracoIsRed3(c) || BurracoIsWild(c) {
			continue
		}
		val := BurracoCardValue(c)
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

// cpuBestDiscardHard Hard: より戦略的なディスカード
func (g *Burraco) cpuBestDiscardHard(player *BurracoPlayer) int {
	// Normal のロジックをベースに、相手が取りにくいカードを捨てる
	return g.cpuBestDiscardNormal(player)
}

// --- Scoring ---

// scoreRound ラウンドのスコアを確定する
func (g *Burraco) scoreRound(goOutPlayerIdx int, goOutBonus int) {
	for i := 0; i < BurracoPlayerCnt; i++ {
		player := g.players[i]
		score := 0

		// メルドのカード点数
		for _, m := range player.melds {
			for _, c := range m.Cards {
				score += BurracoCardValue(c)
			}
			// ブラーコボーナス
			if m.IsBurraco() {
				if m.IsNatural {
					score += BurracoNaturalBurracoBonus
				} else {
					score += BurracoMixedBurracoBonus
				}
			}
		}

		// 赤3ボーナス
		red3Count := len(player.red3s)
		if red3Count == 4 {
			score += BurracoAllRed3Bonus
		} else {
			score += red3Count * BurracoRed3Bonus
		}

		// 上がりボーナス
		if i == goOutPlayerIdx {
			score += goOutBonus
		}

		// 手札のカード点数を減算
		for j := 0; j < player.GetCardsSize(); j++ {
			score -= BurracoCardValue(player.GetCard(j))
		}

		player.SetRoundScore(score)
		g.appendLog(i, "score", fmt.Sprintf("%s scores %d points this round", g.playerName(i), score), nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = BurracoPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによるラウンド終了
func (g *Burraco) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends (stock empty)", nil)
	g.scoreRound(-1, 0)
}

// advanceTurn 次のプレイヤーへ
func (g *Burraco) advanceTurn() {
	g.currentPlayerIdx = 1 - g.currentPlayerIdx
	g.drewFromDiscard = false
	g.drawnCard = nil

	// 山札が空で手札が全員あるなら引き分け
	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	g.phase = BurracoPhaseDraw
}

// checkGameEnd ゲーム終了判定
func (g *Burraco) checkGameEnd() {
	hasWinner := false
	for i := 0; i < BurracoPlayerCnt; i++ {
		if g.players[i].cumulativeScore >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = BurracoPhaseGameEnd

	maxScore := g.players[0].cumulativeScore
	g.winnerIdx = 0
	for i := 1; i < BurracoPlayerCnt; i++ {
		if g.players[i].cumulativeScore > maxScore {
			maxScore = g.players[i].cumulativeScore
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", g.playerName(g.winnerIdx)), nil)
}

// --- Meld Validation ---

// validateNewMeld 新規メルドの検証
func (g *Burraco) validateNewMeld(cards []*Card) error {
	if len(cards) < 3 {
		return NewDomainError(ErrInvalidPlay, "メルドには最低3枚のカードが必要です")
	}

	var naturalCount, wildCount int
	rank := 0
	for _, c := range cards {
		if BurracoIsWild(c) {
			wildCount++
		} else {
			naturalCount++
			if rank == 0 {
				rank = c.GetValue()
			} else if c.GetValue() != rank {
				return NewDomainError(ErrInvalidPlay, "メルドは同じランクのカードで構成する必要があります")
			}
		}
	}

	// 黒3はメルドできない
	for _, c := range cards {
		if BurracoIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
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

// validateMeldAddition 既存メルドへの追加の検証
func (g *Burraco) validateMeldAddition(existing *BurracoMeld, cards []*Card) error {
	rank := existing.GetRank()
	existingWildCount := 0
	for _, c := range existing.Cards {
		if BurracoIsWild(c) {
			existingWildCount++
		}
	}

	newWildCount := existingWildCount
	for _, c := range cards {
		if BurracoIsWild(c) {
			newWildCount++
		} else if c.GetValue() != rank {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ランク%dのメルドにランク%dのカードは追加できません", rank, c.GetValue()))
		}
		if BurracoIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
		}
	}

	if newWildCount > 3 {
		return NewDomainError(ErrInvalidPlay, "メルドにはワイルドカードは最大3枚までです")
	}

	return nil
}

// findExistingMeldForCards カードが追加できる既存メルドのインデックスを返す (-1 = なし)
func (g *Burraco) findExistingMeldForCards(playerIdx int, cards []*Card) int {
	player := g.players[playerIdx]
	// カードのランクを特定（ナチュラルカードから）
	rank := 0
	for _, c := range cards {
		if !BurracoIsWild(c) {
			rank = c.GetValue()
			break
		}
	}
	if rank == 0 {
		return -1 // すべてワイルド → 既存メルドへの追加は不可
	}

	for i, m := range player.melds {
		if m.GetRank() == rank {
			return i
		}
	}
	return -1
}

// minimumMeldValue 初回メルドの最低点を返す
func (g *Burraco) minimumMeldValue(playerIdx int) int {
	score := g.players[playerIdx].GetCumulativeScore()
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

// BurracoIsWild ワイルドカードかどうか (ジョーカーまたは2)
func BurracoIsWild(card *Card) bool {
	return card.GetDesign() == CardDesignJoker || card.GetValue() == 2
}

// BurracoIsRed3 赤3かどうか (ハート3またはダイヤ3)
func BurracoIsRed3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignHeart || card.GetDesign() == CardDesignDiamond)
}

// BurracoIsBlack3 黒3かどうか (スペード3またはクローバー3)
func BurracoIsBlack3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover)
}

// BurracoCardValue カードの点数を返す
func BurracoCardValue(card *Card) int {
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

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Burraco) GetPhase() BurracoPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Burraco) SetPhase(phase BurracoPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Burraco) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Burraco) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Burraco) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Burraco) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Burraco) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Burraco) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Burraco) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Burraco) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Burraco) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetDiscardPileCount 捨て札の枚数取得
func (g *Burraco) GetDiscardPileCount() int { return len(g.discardPile) }

// GetPozzettoCount 残っているポゼット（予備手札）の山の数を取得
func (g *Burraco) GetPozzettoCount() int { return len(g.pozzetti) }

// GetPozzettoCardCount 残っているポゼットの総カード枚数を取得
func (g *Burraco) GetPozzettoCardCount() int {
	n := 0
	for _, pile := range g.pozzetti {
		n += len(pile)
	}
	return n
}

// SetPozzetti ポゼットを設定 (テスト用)
func (g *Burraco) SetPozzetti(piles [][]*Card) { g.pozzetti = piles }

// GetIsFrozen 捨て札の山がフリーズ状態か取得
func (g *Burraco) GetIsFrozen() bool { return g.isFrozen }

// SetIsFrozen フリーズ状態設定 (テスト用)
func (g *Burraco) SetIsFrozen(v bool) { g.isFrozen = v }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Burraco) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (g *Burraco) SetGameEndFlag(v bool) { g.gameEndFlag = v }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Burraco) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Burraco) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Burraco) GetPlayer(i int) *BurracoPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Burraco) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Burraco) GetConfig() BurracoConfig { return g.config }

// SetConfig 設定変更
func (g *Burraco) SetConfig(cfg BurracoConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Burraco) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetDrewFromDiscard 捨て札から引いたか取得
func (g *Burraco) GetDrewFromDiscard() bool { return g.drewFromDiscard }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Burraco) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Burraco) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す
func (g *Burraco) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (g *Burraco) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// burracoTypeStr ブラーコの種別文字列を返す
func burracoTypeStr(isNatural bool) string {
	if isNatural {
		return "natural"
	}
	return "mixed"
}

// --- JSON serialization ---

// burracoJSON is the JSON wire format for Burraco.
type burracoJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*BurracoPlayer  `json:"pl"`
	Config           BurracoConfig     `json:"cf"`
	Phase            BurracoPhase      `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	Pozzetti         [][]*Card         `json:"pz"`
	IsFrozen         bool              `json:"fr"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
	DrewFromDiscard  bool              `json:"dd"`
}

// MarshalJSON implements json.Marshaler.
func (g *Burraco) MarshalJSON() ([]byte, error) {
	return json.Marshal(burracoJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		Pozzetti:         g.pozzetti,
		IsFrozen:         g.isFrozen,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		DrewFromDiscard:  g.drewFromDiscard,
	})
}

// burracoMaxSliceLen caps slice sizes during deserialisation.
const burracoMaxSliceLen = 2000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Burraco) UnmarshalJSON(data []byte) error {
	var j burracoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > burracoMaxSliceLen || len(j.DiscardPile) > burracoMaxSliceLen ||
		len(j.DrawPile) > burracoMaxSliceLen || len(j.ActionLog) > burracoMaxSliceLen {
		return fmt.Errorf("burraco: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsWithDecks(2, 4)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*BurracoPlayer, 0)
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
	g.pozzetti = j.Pozzetti
	g.isFrozen = j.IsFrozen
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.drewFromDiscard = j.DrewFromDiscard
	return nil
}
