//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// CanastaPlayerCnt カナスタプレイヤー数
const CanastaPlayerCnt = 2

// CanastaHandSize 初期配布枚数 (2人制カナスタ)
const CanastaHandSize = 15

// CanastaBurracoHandSize Burraco モードの初期配布枚数 (2人制)
const CanastaBurracoHandSize = 11

// CanastaPozzettoSize Burraco モードのポゼット（予備手札）1山の枚数
const CanastaPozzettoSize = 11

// CanastaDefaultPointLimit デフォルトの目標スコア
const CanastaDefaultPointLimit = 5000

// カナスタスコア定数
const (
	CanastaGoingOutBonus          = 100 // 上がりボーナス
	CanastaConcealedGoingOutBonus = 200 // コンシールド上がりボーナス (一度もメルドせずに上がる)
	CanastaNaturalCanastaBonus    = 500 // ナチュラルカナスタボーナス (ワイルドなし)
	CanastaMixedCanastaBonus      = 300 // ミックスカナスタボーナス (ワイルドあり)
	CanastaRed3Bonus              = 100 // 赤3のボーナス (1枚あたり)
	CanastaAllRed3Bonus           = 800 // 赤3全4枚ボーナス
)

// CanastaPhase ゲームフェーズ
type CanastaPhase int

// Canastaのフェーズ定数
const (
	// CanastaPhaseDraw ドローフェーズ (山札または捨て札から引く)
	CanastaPhaseDraw CanastaPhase = 0
	// CanastaPhaseMeld メルドフェーズ (メルドを出す/既存メルドに追加する)
	CanastaPhaseMeld CanastaPhase = 1
	// CanastaPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or 上がる)
	CanastaPhaseDiscard CanastaPhase = 2
	// CanastaPhaseRoundEnd ラウンド終了フェーズ
	CanastaPhaseRoundEnd CanastaPhase = 3
	// CanastaPhaseGameEnd ゲーム終了フェーズ
	CanastaPhaseGameEnd CanastaPhase = 4
)

// Canasta カナスタゲームクラス
type Canasta struct {
	trumpCards       *TrumpCards
	players          []*CanastaPlayer
	config           CanastaConfig
	phase            CanastaPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	pozzetti         [][]*Card // Burraco モードの予備手札（ポゼット）。各11枚×2山
	isFrozen         bool      // 捨て札の山がフリーズ状態か
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
	drewFromDiscard bool  // 現在のターンで捨て札の山から引いたか
	drawnCard       *Card // 捨て札の山のトップカード (メルドバリデーション用)
}

// NewCanasta コンストラクタ
func NewCanasta(trumpCards *TrumpCards, players []*CanastaPlayer, config CanastaConfig) *Canasta {
	return &Canasta{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultCanasta returns Canasta with the standard 2-player setup (1 human, 1 CPU)
// using a 2-deck pack with 4 jokers and DefaultCanastaConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCanasta() *Canasta {
	players := []*CanastaPlayer{
		NewCanastaPlayer(true),
		NewCanastaPlayer(false),
	}
	return NewCanasta(NewTrumpCardsWithDecks(2, 4), players, DefaultCanastaConfig())
}

// Reset ゲーム初期化
func (g *Canasta) Reset() {
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
		p.melds = make([]*CanastaMeld, 0)
		p.red3s = make([]*Card, 0)
		p.hasInitMeld = false
		p.tookPozzetto = false
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = CanastaPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Canasta) NextRound() {
	if g.phase != CanastaPhaseRoundEnd {
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

	g.phase = CanastaPhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに15枚、1枚を捨て札に
func (g *Canasta) dealInitialCards() {
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

	// 各プレイヤーに配布 (Canasta=15枚, Burraco=11枚)
	handSize := CanastaHandSize
	if g.config.UsePozzetto {
		handSize = CanastaBurracoHandSize
	}
	for i := 0; i < handSize; i++ {
		for j := 0; j < CanastaPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// Burraco モード: ポゼット（予備手札）を2山、各11枚ずつ脇に取り分ける
	if g.config.UsePozzetto {
		g.pozzetti = make([][]*Card, 0, CanastaPlayerCnt)
		for p := 0; p < CanastaPlayerCnt; p++ {
			pile := make([]*Card, 0, CanastaPozzettoSize)
			for i := 0; i < CanastaPozzettoSize && len(g.drawPile) > 0; i++ {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				pile = append(pile, card)
			}
			g.pozzetti = append(g.pozzetti, pile)
		}
	}

	// 赤3の処理: 手札の赤3を自動的に場に出し、山札から補充
	for j := 0; j < CanastaPlayerCnt; j++ {
		g.autoLayRed3s(j)
	}

	// 最初の1枚を捨て札に (赤3やワイルドが出たら次のカードを引く)
	for len(g.drawPile) > 0 {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, card)
		if CanastaIsRed3(card) {
			// 赤3は捨て札にできない → 次のカードを引く
			continue
		}
		if CanastaIsWild(card) {
			g.isFrozen = true
		}
		break
	}
}

// autoLayRed3s プレイヤーの手札から赤3を自動的に場に出す
func (g *Canasta) autoLayRed3s(playerIdx int) {
	player := g.players[playerIdx]
	for {
		found := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if CanastaIsRed3(card) {
				player.RemoveCard(i)
				player.AddRed3(card)
				g.appendLog(playerIdx, "red3", fmt.Sprintf("%s lays down red 3: %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})
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

// takePozzetto Burraco モードで、プレイヤーが手札を出し切ったときポゼット
// （予備手札）を1山獲得して手札に加える。取得した場合 true を返す。
func (g *Canasta) takePozzetto(playerIdx int) bool {
	player := g.players[playerIdx]
	if !g.config.UsePozzetto || player.tookPozzetto || len(g.pozzetti) == 0 {
		return false
	}
	pile := g.pozzetti[len(g.pozzetti)-1]
	g.pozzetti = g.pozzetti[:len(g.pozzetti)-1]
	for _, c := range pile {
		player.AddCard(c)
	}
	player.tookPozzetto = true
	g.appendLog(playerIdx, "pozzetto", fmt.Sprintf("%s takes the pozzetto (%d cards)", playerName(g.players, playerIdx), len(pile)), nil)
	// 獲得した手札に赤3があれば自動的に場に出す
	g.autoLayRed3s(playerIdx)
	g.sortHand(playerIdx)
	return true
}

// canGoOut 上がり条件を満たすか。Burraco モードではポゼット獲得済みかつ
// カナスタ（ブラーコ）完成が必要。Canasta モードではカナスタ完成のみ。
func (g *Canasta) canGoOut(playerIdx int) bool {
	p := g.players[playerIdx]
	if g.config.UsePozzetto && !p.tookPozzetto {
		return false
	}
	return p.HasCanasta()
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Canasta) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CanastaPhaseDraw {
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

	// 赤3を引いた場合は自動的に場に出す (autoLayRed3s が補充まで処理する)
	if CanastaIsRed3(card) {
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

	g.phase = CanastaPhaseMeld
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札の山を取る
func (g *Canasta) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CanastaPhaseDraw {
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
	if CanastaIsBlack3(topCard) {
		return NewDomainError(ErrInvalidPlay, "黒3がトップの場合は捨て札の山を取れません")
	}

	// ワイルドカードがトップの場合は取れない（通常はワイルドカードがトップにくると山がフリーズ）
	if CanastaIsWild(topCard) {
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
	if CanastaIsWild(card0) || CanastaIsWild(card1) {
		return NewDomainError(ErrInvalidPlay, "ペアはナチュラルカード（ワイルドカード以外）でなければなりません")
	}
	if card0.GetValue() != topCard.GetValue() || card1.GetValue() != topCard.GetValue() {
		return NewDomainError(ErrInvalidPlay, "ペアのランクが捨て札のトップカードと一致しません")
	}

	// 初回メルド要件チェック: 捨て札の山を取る場合、初回メルドの最低点を満たすメルドが必要
	if !player.hasInitMeld {
		// トップカード + ペアで最低3枚のメルドが作れる: その点数をチェック
		meldValue := CanastaCardValue(topCard) + CanastaCardValue(card0) + CanastaCardValue(card1)
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

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s picks up the discard pile (%d cards)", playerName(g.players, g.currentPlayerIdx), pileSize), []*Card{topCard})

	// 手札に赤3があれば自動的に場に出す
	g.autoLayRed3s(g.currentPlayerIdx)
	g.sortHand(g.currentPlayerIdx)

	g.phase = CanastaPhaseMeld
	return nil
}

// PlayerMeld 人間プレイヤーがメルドを出す
func (g *Canasta) PlayerMeld(meldGroups [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CanastaPhaseMeld {
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
		g.phase = CanastaPhaseDiscard
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
				totalMeldValue += CanastaCardValue(c)
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
				if CanastaIsWild(c) {
					isNatural = false
					break
				}
			}
			meld := &CanastaMeld{
				Cards:     action.cards,
				IsNatural: isNatural,
			}
			player.AddMeld(meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d cards (rank %d)", playerName(g.players, g.currentPlayerIdx), len(action.cards), meld.GetRank()), action.cards)
		} else {
			existing := player.melds[action.existingIdx]
			for _, c := range action.cards {
				existing.Cards = append(existing.Cards, c)
				if CanastaIsWild(c) {
					existing.IsNatural = false
				}
			}
			g.appendLog(g.currentPlayerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to meld (rank %d)", playerName(g.players, g.currentPlayerIdx), len(action.cards), existing.GetRank()), action.cards)
		}
	}

	// 手札からカードを削除
	for _, idx := range allIdx {
		player.RemoveCard(idx)
	}

	player.hasInitMeld = true
	g.drewFromDiscard = false
	g.drawnCard = nil

	// カナスタ完成チェック
	for _, m := range player.melds {
		if m.IsCanasta() {
			g.appendLog(g.currentPlayerIdx, "canasta", fmt.Sprintf("%s completes a %s canasta!", playerName(g.players, g.currentPlayerIdx), canastaTypeStr(m.IsNatural)), nil)
		}
	}

	// 手札を出し切った場合の処理
	if player.GetCardsSize() == 0 {
		if g.config.UsePozzetto && !player.tookPozzetto {
			// Burraco: 初めて手札を出し切った → ポゼットを獲得して継続
			g.takePozzetto(g.currentPlayerIdx)
		} else if player.HasCanasta() {
			g.goOut(g.currentPlayerIdx, false)
			return nil
		}
	}

	g.phase = CanastaPhaseDiscard
	return nil
}

// PlayerSkipMeld メルドフェーズをスキップする
func (g *Canasta) PlayerSkipMeld() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CanastaPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.drewFromDiscard {
		return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
	}

	g.phase = CanastaPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *Canasta) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CanastaPhaseDiscard {
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
	if CanastaIsRed3(card) {
		return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	// ワイルドカードを捨てた場合、山をフリーズ
	if CanastaIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	// Burraco: 捨て札で手札を出し切り、まだポゼット未獲得なら獲得する
	if g.config.UsePozzetto && player.GetCardsSize() == 0 && !player.tookPozzetto {
		g.takePozzetto(g.currentPlayerIdx)
	}

	g.advanceTurn()
	return nil
}

// PlayerGoOut 人間プレイヤーが上がる
func (g *Canasta) PlayerGoOut() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CanastaPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]

	if g.config.UsePozzetto && !player.tookPozzetto {
		return NewDomainError(ErrInvalidPlay, "上がるにはポゼット（予備手札）を獲得している必要があります")
	}
	if !player.HasCanasta() {
		return NewDomainError(ErrInvalidPlay, "上がるには少なくとも1つのカナスタが必要です")
	}

	// 手札が1枚の場合、最後のカードを捨てて上がる
	if player.GetCardsSize() == 1 {
		card := player.GetCard(0)
		if CanastaIsRed3(card) {
			return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
		}
		discarded := player.RemoveCard(0)
		g.discardPile = append(g.discardPile, discarded)
		if CanastaIsWild(discarded) {
			g.isFrozen = true
		}
		g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	} else if player.GetCardsSize() > 1 {
		return NewDomainError(ErrInvalidPlay, "上がるには手札が0枚または1枚でなければなりません")
	}

	concealed := !player.hasInitMeld
	g.goOut(g.currentPlayerIdx, concealed)
	return nil
}

// goOut 上がり処理
func (g *Canasta) goOut(playerIdx int, concealed bool) {
	bonus := CanastaGoingOutBonus
	if concealed {
		bonus = CanastaConcealedGoingOutBonus
	}
	g.appendLog(playerIdx, "go_out", fmt.Sprintf("%s goes out! (bonus: %d)", playerName(g.players, playerIdx), bonus), nil)
	g.scoreRound(playerIdx, bonus)
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *Canasta) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case CanastaPhaseDraw:
		g.cpuDraw()
	case CanastaPhaseMeld:
		g.cpuMeld()
	case CanastaPhaseDiscard:
		g.cpuDiscard()
	}
}

// cpuDraw CPUがドローする
func (g *Canasta) cpuDraw() {
	player := g.players[g.currentPlayerIdx]

	// 捨て札の山を取るか判定
	if len(g.discardPile) > 0 {
		topCard := g.discardPile[len(g.discardPile)-1]
		if !CanastaIsBlack3(topCard) && !CanastaIsWild(topCard) {
			shouldPickUp := false
			switch g.config.CpuDifficulty {
			case CanastaCpuDifficultyHard:
				shouldPickUp = g.cpuShouldPickUpPileHard(player, topCard)
			case CanastaCpuDifficultyNormal:
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
						meldValue := CanastaCardValue(topCard) + CanastaCardValue(player.GetCard(pairIndices[0])) + CanastaCardValue(player.GetCard(pairIndices[1]))
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
						g.phase = CanastaPhaseMeld
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
	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	// 赤3の処理 (autoLayRed3s が補充まで処理する)
	if CanastaIsRed3(card) {
		g.autoLayRed3s(g.currentPlayerIdx)
		if len(g.drawPile) == 0 && player.GetCardsSize() == 0 {
			g.endRoundDraw()
			return
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)
	g.phase = CanastaPhaseMeld
}

// cpuMeld CPUがメルドする
func (g *Canasta) cpuMeld() {
	player := g.players[g.currentPlayerIdx]

	meldGroups := g.cpuFindMelds(player)
	if len(meldGroups) == 0 {
		g.drewFromDiscard = false
		g.drawnCard = nil
		g.phase = CanastaPhaseDiscard
		return
	}

	// 初回メルドの最低点チェック
	if !player.hasInitMeld {
		totalValue := 0
		for _, group := range meldGroups {
			for _, c := range group {
				totalValue += CanastaCardValue(c)
			}
		}
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if totalValue < minReq {
			g.drewFromDiscard = false
			g.drawnCard = nil
			g.phase = CanastaPhaseDiscard
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
				if CanastaIsWild(c) {
					existing.IsNatural = false
				}
			}
			g.appendLog(g.currentPlayerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to meld (rank %d)", playerName(g.players, g.currentPlayerIdx), len(group), existing.GetRank()), group)
		} else {
			isNatural := true
			for _, c := range group {
				if CanastaIsWild(c) {
					isNatural = false
					break
				}
			}
			meld := &CanastaMeld{Cards: group, IsNatural: isNatural}
			player.AddMeld(meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d cards (rank %d)", playerName(g.players, g.currentPlayerIdx), len(group), meld.GetRank()), group)
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
		if g.config.UsePozzetto && !player.tookPozzetto {
			g.takePozzetto(g.currentPlayerIdx)
		} else if player.HasCanasta() {
			g.goOut(g.currentPlayerIdx, false)
			return
		}
	}

	g.phase = CanastaPhaseDiscard
}

// cpuDiscard CPUがディスカードする
func (g *Canasta) cpuDiscard() {
	player := g.players[g.currentPlayerIdx]

	// Defensive: if the CPU reached the discard phase with no cards (e.g. a
	// prior meld emptied the hand but HasCanasta() was false so goOut didn't
	// fire), do not try to discard — that path panics via RemoveCard → nil.
	if player.GetCardsSize() == 0 {
		if g.config.UsePozzetto && !player.tookPozzetto {
			// Burraco: 手札を出し切ったがポゼット未獲得 → 獲得して継続
			g.takePozzetto(g.currentPlayerIdx)
		} else if g.config.UsePozzetto {
			// Burraco: ポゼット獲得済みだがブラーコ未完成 → 捨てられないので進む
			g.advanceTurn()
			return
		} else {
			g.goOut(g.currentPlayerIdx, false)
			return
		}
	}

	// 上がれるかチェック
	if player.GetCardsSize() == 1 && g.canGoOut(g.currentPlayerIdx) {
		card := player.GetCard(0)
		if !CanastaIsRed3(card) {
			discarded := player.RemoveCard(0)
			g.discardPile = append(g.discardPile, discarded)
			if CanastaIsWild(discarded) {
				g.isFrozen = true
			}
			g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
			g.goOut(g.currentPlayerIdx, false)
			return
		}
	}

	// 最適なディスカードを選択
	bestIdx := g.cpuBestDiscard(player)
	discarded := player.RemoveCard(bestIdx)
	g.discardPile = append(g.discardPile, discarded)

	if CanastaIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// --- CPU AI helpers ---

// cpuShouldPickUpPileNormal Normal難易度: トップカードとペアがあれば取る
func (g *Canasta) cpuShouldPickUpPileNormal(player *CanastaPlayer, topCard *Card) bool {
	return g.cpuHasNaturalPair(player, topCard)
}

// cpuShouldPickUpPileHard Hard難易度: 戦略的に山を取る
func (g *Canasta) cpuShouldPickUpPileHard(player *CanastaPlayer, topCard *Card) bool {
	if !g.cpuHasNaturalPair(player, topCard) {
		return false
	}
	// 山が大きいほど取る価値がある
	return len(g.discardPile) >= 3
}

// cpuHasNaturalPair トップカードと同ランクのナチュラルカードのペアがあるか
func (g *Canasta) cpuHasNaturalPair(player *CanastaPlayer, topCard *Card) bool {
	return g.cpuFindNaturalPair(player, topCard) != nil
}

// cpuFindNaturalPair ナチュラルペアのインデックスを返す (見つからなければnil)
func (g *Canasta) cpuFindNaturalPair(player *CanastaPlayer, topCard *Card) []int {
	var indices []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !CanastaIsWild(c) && c.GetValue() == topCard.GetValue() {
			indices = append(indices, i)
			if len(indices) == 2 {
				return indices
			}
		}
	}
	return nil
}

// cpuFindMelds CPUのメルド候補を見つける
func (g *Canasta) cpuFindMelds(player *CanastaPlayer) [][]*Card {
	var melds [][]*Card

	// 手札をランクごとにグループ化
	byRank := make(map[int][]*Card)
	var wilds []*Card
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if CanastaIsWild(c) {
			wilds = append(wilds, c)
		} else if !CanastaIsRed3(c) && !CanastaIsBlack3(c) {
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
func (g *Canasta) cpuBestDiscard(player *CanastaPlayer) int {
	// 黒3を優先的に捨てる（相手のピックアップをブロック）
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if CanastaIsBlack3(c) {
			return i
		}
	}

	switch g.config.CpuDifficulty {
	case CanastaCpuDifficultyHard:
		return g.cpuBestDiscardHard(player)
	case CanastaCpuDifficultyNormal:
		return g.cpuBestDiscardNormal(player)
	default:
		// Easy: ランダム (赤3とワイルド以外)
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if !CanastaIsRed3(c) && !CanastaIsWild(c) {
				return i
			}
		}
		return 0
	}
}

// cpuBestDiscardNormal Normal: 孤立カードを捨てる
func (g *Canasta) cpuBestDiscardNormal(player *CanastaPlayer) int {
	// ランクごとのカウントを計算
	rankCount := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !CanastaIsWild(c) && !CanastaIsRed3(c) {
			rankCount[c.GetValue()]++
		}
	}

	// 孤立カード (カウント1) のうち点数が低いものを捨てる
	bestIdx := -1
	bestValue := 1000
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if CanastaIsRed3(c) || CanastaIsWild(c) {
			continue
		}
		cnt := rankCount[c.GetValue()]
		val := CanastaCardValue(c)
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
		if CanastaIsRed3(c) || CanastaIsWild(c) {
			continue
		}
		val := CanastaCardValue(c)
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
func (g *Canasta) cpuBestDiscardHard(player *CanastaPlayer) int {
	// Normal のロジックをベースに、相手が取りにくいカードを捨てる
	return g.cpuBestDiscardNormal(player)
}

// --- Scoring ---

// scoreRound ラウンドのスコアを確定する
func (g *Canasta) scoreRound(goOutPlayerIdx int, goOutBonus int) {
	for i := 0; i < CanastaPlayerCnt; i++ {
		player := g.players[i]
		score := 0

		// メルドのカード点数
		for _, m := range player.melds {
			for _, c := range m.Cards {
				score += CanastaCardValue(c)
			}
			// カナスタボーナス
			if m.IsCanasta() {
				if m.IsNatural {
					score += CanastaNaturalCanastaBonus
				} else {
					score += CanastaMixedCanastaBonus
				}
			}
		}

		// 赤3ボーナス
		red3Count := len(player.red3s)
		if red3Count == 4 {
			score += CanastaAllRed3Bonus
		} else {
			score += red3Count * CanastaRed3Bonus
		}

		// 上がりボーナス
		if i == goOutPlayerIdx {
			score += goOutBonus
		}

		// 手札のカード点数を減算
		for j := 0; j < player.GetCardsSize(); j++ {
			score -= CanastaCardValue(player.GetCard(j))
		}

		player.SetRoundScore(score)
		g.appendLog(i, "score", fmt.Sprintf("%s scores %d points this round", playerName(g.players, i), score), nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = CanastaPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによるラウンド終了
func (g *Canasta) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends (stock empty)", nil)
	g.scoreRound(-1, 0)
}

// advanceTurn 次のプレイヤーへ
func (g *Canasta) advanceTurn() {
	g.currentPlayerIdx = 1 - g.currentPlayerIdx
	g.drewFromDiscard = false
	g.drawnCard = nil

	// 山札が空で手札が全員あるなら引き分け
	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	g.phase = CanastaPhaseDraw
}

// checkGameEnd ゲーム終了判定
func (g *Canasta) checkGameEnd() {
	hasWinner := false
	for i := 0; i < CanastaPlayerCnt; i++ {
		if g.players[i].cumulativeScore >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = CanastaPhaseGameEnd

	maxScore := g.players[0].cumulativeScore
	g.winnerIdx = 0
	for i := 1; i < CanastaPlayerCnt; i++ {
		if g.players[i].cumulativeScore > maxScore {
			maxScore = g.players[i].cumulativeScore
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// --- Meld Validation ---

// validateNewMeld 新規メルドの検証
func (g *Canasta) validateNewMeld(cards []*Card) error {
	if len(cards) < 3 {
		return NewDomainError(ErrInvalidPlay, "メルドには最低3枚のカードが必要です")
	}

	var naturalCount, wildCount int
	rank := 0
	for _, c := range cards {
		if CanastaIsWild(c) {
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
		if CanastaIsBlack3(c) {
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
func (g *Canasta) validateMeldAddition(existing *CanastaMeld, cards []*Card) error {
	rank := existing.GetRank()
	existingWildCount := 0
	for _, c := range existing.Cards {
		if CanastaIsWild(c) {
			existingWildCount++
		}
	}

	newWildCount := existingWildCount
	for _, c := range cards {
		if CanastaIsWild(c) {
			newWildCount++
		} else if c.GetValue() != rank {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ランク%dのメルドにランク%dのカードは追加できません", rank, c.GetValue()))
		}
		if CanastaIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
		}
	}

	if newWildCount > 3 {
		return NewDomainError(ErrInvalidPlay, "メルドにはワイルドカードは最大3枚までです")
	}

	return nil
}

// findExistingMeldForCards カードが追加できる既存メルドのインデックスを返す (-1 = なし)
func (g *Canasta) findExistingMeldForCards(playerIdx int, cards []*Card) int {
	player := g.players[playerIdx]
	// カードのランクを特定（ナチュラルカードから）
	rank := 0
	for _, c := range cards {
		if !CanastaIsWild(c) {
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
func (g *Canasta) minimumMeldValue(playerIdx int) int {
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

// CanastaIsWild ワイルドカードかどうか (ジョーカーまたは2)
func CanastaIsWild(card *Card) bool {
	return card.GetDesign() == CardDesignJoker || card.GetValue() == 2
}

// CanastaIsRed3 赤3かどうか (ハート3またはダイヤ3)
func CanastaIsRed3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignHeart || card.GetDesign() == CardDesignDiamond)
}

// CanastaIsBlack3 黒3かどうか (スペード3またはクローバー3)
func CanastaIsBlack3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover)
}

// CanastaCardValue カードの点数を返す
func CanastaCardValue(card *Card) int {
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
func (g *Canasta) GetPhase() CanastaPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Canasta) SetPhase(phase CanastaPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Canasta) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Canasta) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Canasta) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Canasta) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Canasta) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Canasta) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Canasta) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Canasta) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Canasta) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetDiscardPileCount 捨て札の枚数取得
func (g *Canasta) GetDiscardPileCount() int { return len(g.discardPile) }

// GetPozzettoCount 残っているポゼット（予備手札）の山の数を取得 (Burraco モード)
func (g *Canasta) GetPozzettoCount() int { return len(g.pozzetti) }

// GetPozzettoCardCount 残っているポゼットの総カード枚数を取得
func (g *Canasta) GetPozzettoCardCount() int {
	n := 0
	for _, pile := range g.pozzetti {
		n += len(pile)
	}
	return n
}

// SetPozzetti ポゼットを設定 (テスト用)
func (g *Canasta) SetPozzetti(piles [][]*Card) { g.pozzetti = piles }

// GetIsFrozen 捨て札の山がフリーズ状態か取得
func (g *Canasta) GetIsFrozen() bool { return g.isFrozen }

// SetIsFrozen フリーズ状態設定 (テスト用)
func (g *Canasta) SetIsFrozen(v bool) { g.isFrozen = v }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Canasta) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (g *Canasta) SetGameEndFlag(v bool) { g.gameEndFlag = v }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Canasta) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Canasta) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Canasta) GetPlayer(i int) *CanastaPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Canasta) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Canasta) GetConfig() CanastaConfig { return g.config }

// SetConfig 設定変更
func (g *Canasta) SetConfig(cfg CanastaConfig) { g.config = cfg }

// GetDrewFromDiscard 捨て札から引いたか取得
func (g *Canasta) GetDrewFromDiscard() bool { return g.drewFromDiscard }

// --- Hint ---

// CanastaHint はカナスタ / ブラーコの現在手番に対する推奨アクション。
type CanastaHint struct {
	// Action は推奨アクション種別。
	// "draw_stock" / "draw_discard" / "meld" / "skip_meld" / "discard" のいずれか。
	Action string
	// Indices は推奨アクションの対象カードの手札インデックス。
	// draw_discard: 捨て札トップと同ランクのナチュラルペア、meld: メルド候補、
	// discard: 捨てる1枚。draw_stock / skip_meld では空。
	Indices []int
	// Reason はヒント理由の i18n キー(接頭辞なし)。
	Reason string
}

// GetHint は現在の人間の手番フェーズに応じた推奨アクションを返す。CPU の手番や
// ヒント不要なフェーズでは nil を返す。判定には既存の CPU AI ヘルパーを再利用する。
func (g *Canasta) GetHint() *CanastaHint {
	if !g.IsHumanTurn() {
		return nil
	}
	player := g.players[g.currentPlayerIdx]
	switch g.phase {
	case CanastaPhaseDraw:
		if top := g.GetDiscardTop(); top != nil {
			if pair := g.cpuFindNaturalPair(player, top); pair != nil {
				return &CanastaHint{Action: "draw_discard", Indices: pair, Reason: "draw_discard_pair"}
			}
		}
		return &CanastaHint{Action: "draw_stock", Reason: "draw_stock_safe"}
	case CanastaPhaseMeld:
		if melds := g.cpuFindMelds(player); len(melds) > 0 {
			return &CanastaHint{Action: "meld", Indices: g.handIndicesOf(player, melds[0]), Reason: "meld_available"}
		}
		return &CanastaHint{Action: "skip_meld", Reason: "no_meld"}
	case CanastaPhaseDiscard:
		if player.GetCardsSize() == 0 {
			return nil
		}
		return &CanastaHint{Action: "discard", Indices: []int{g.cpuBestDiscard(player)}, Reason: "discard_safe"}
	default:
		return nil
	}
}

// handIndicesOf は指定カード群の手札インデックスを返す(見つからないカードは除外)。
func (g *Canasta) handIndicesOf(player *CanastaPlayer, cards []*Card) []int {
	idxs := make([]int, 0, len(cards))
	for _, c := range cards {
		for i := 0; i < player.GetCardsSize(); i++ {
			if player.GetCard(i) == c {
				idxs = append(idxs, i)
				break
			}
		}
	}
	return idxs
}

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Canasta) sortAllHands() {
	sortHands(len(g.players), g)
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Canasta) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// canastaTypeStr カナスタの種別文字列を返す
func canastaTypeStr(isNatural bool) string {
	if isNatural {
		return "natural"
	}
	return "mixed"
}

// --- JSON serialization ---

// canastaJSON is the JSON wire format for Canasta.
type canastaJSON struct {
	TrumpCards       *TrumpCards      `json:"tc"`
	Players          []*CanastaPlayer `json:"pl"`
	Config           CanastaConfig    `json:"cf"`
	Phase            CanastaPhase     `json:"ps"`
	CurrentPlayerIdx int              `json:"ci"`
	DiscardPile      []*Card          `json:"dp"`
	DrawPile         []*Card          `json:"wp"`
	// Pozzetto piles are serialised as two flat []*Card fields rather than a
	// [][]*Card. The slice-of-slice form would make TinyGo emit a dedicated
	// encoder that ships in every Cloudflare Worker WASM binary (the classic
	// worker is at the 1 MB gzip limit); the flat []*Card encoder already exists.
	Pozzetto1       []*Card           `json:"p1,omitempty"`
	Pozzetto2       []*Card           `json:"p2,omitempty"`
	IsFrozen        bool              `json:"fr"`
	GameEndFlag     bool              `json:"ge"`
	WinnerIdx       int               `json:"wi"`
	RoundNumber     int               `json:"rn"`
	ActionLog       []*ActionLogEntry `json:"al"`
	DrewFromDiscard bool              `json:"dd"`
}

// MarshalJSON implements json.Marshaler.
func (g *Canasta) MarshalJSON() ([]byte, error) {
	var pz1, pz2 []*Card
	if len(g.pozzetti) > 0 {
		pz1 = g.pozzetti[0]
	}
	if len(g.pozzetti) > 1 {
		pz2 = g.pozzetti[1]
	}
	return json.Marshal(canastaJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		Pozzetto1:        pz1,
		Pozzetto2:        pz2,
		IsFrozen:         g.isFrozen,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		DrewFromDiscard:  g.drewFromDiscard,
	})
}

// canastaMaxSliceLen caps slice sizes during deserialisation.
const canastaMaxSliceLen = 2000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Canasta) UnmarshalJSON(data []byte) error {
	var j canastaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > canastaMaxSliceLen || len(j.DiscardPile) > canastaMaxSliceLen ||
		len(j.DrawPile) > canastaMaxSliceLen || len(j.ActionLog) > canastaMaxSliceLen {
		return fmt.Errorf("canasta: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsWithDecks(2, 4)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*CanastaPlayer, 0)
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
	g.pozzetti = nil
	if len(j.Pozzetto1) > 0 {
		g.pozzetti = append(g.pozzetti, j.Pozzetto1)
	}
	if len(j.Pozzetto2) > 0 {
		g.pozzetti = append(g.pozzetti, j.Pozzetto2)
	}
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
