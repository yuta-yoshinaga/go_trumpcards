//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// HandAndFootPlayerCnt ハンドアンドフットのプレイヤー数 (4人/2チーム)
const HandAndFootPlayerCnt = 4

// HandAndFootTeamCnt チーム数
const HandAndFootTeamCnt = 2

// HandAndFootHandSize 初期配布の手札枚数
const HandAndFootHandSize = 22

// HandAndFootFootSize 初期配布のフット枚数
const HandAndFootFootSize = 13

// ハンドアンドフットのスコア定数
const (
	HandAndFootGoingOutBonus     = 100 // 上がりボーナス
	HandAndFootRedCanastaBonus   = 500 // 赤（クリーン）カナスタボーナス
	HandAndFootBlackCanastaBonus = 300 // 黒（ダーティ）カナスタボーナス
	HandAndFootRed3Bonus         = 100 // 赤3ボーナス (1枚あたり)
)

// HandAndFootTeamOf はシートインデックスからチームインデックス (0/1) を返す。
// シート {0,2} がチーム0、{1,3} がチーム1。
func HandAndFootTeamOf(seat int) int { return seat % 2 }

// HandAndFootPhase ゲームフェーズ
type HandAndFootPhase int

// HandAndFootのフェーズ定数 (Canasta と同じ列挙)
const (
	// HandAndFootPhaseDraw ドローフェーズ (山札または捨て札から引く)
	HandAndFootPhaseDraw HandAndFootPhase = 0
	// HandAndFootPhaseMeld メルドフェーズ (メルドを出す/既存メルドに追加する)
	HandAndFootPhaseMeld HandAndFootPhase = 1
	// HandAndFootPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or 上がる)
	HandAndFootPhaseDiscard HandAndFootPhase = 2
	// HandAndFootPhaseRoundEnd ラウンド終了フェーズ
	HandAndFootPhaseRoundEnd HandAndFootPhase = 3
	// HandAndFootPhaseGameEnd ゲーム終了フェーズ
	HandAndFootPhaseGameEnd HandAndFootPhase = 4
)

// HandAndFoot ハンドアンドフットゲームクラス。
//
// 4人2チームのカナスタ系ゲーム。Canasta の実装を踏襲しつつ、以下を変更:
//   - 216枚デッキ (4組 + 8ジョーカー)、手札22枚 + フット13枚を配布。
//   - メルドはチーム単位で共有 (teamMelds[2])。
//   - 山札からは毎ターン2枚引く。捨て札の山を取る場合は手札のナチュラルペア
//     ＋トップカードで最低3枚のメルドが作れること、トップ + 直下最大6枚 (計7枚)
//     を獲得する。黒3トップ・ワイルドトップは取得不可。
//   - 手札を出し切ると即座にフットを手札に取り込む (inFoot)。
//   - 上がり条件: フットに入っており、チームが必要数の赤・黒カナスタを保有し、
//     残り手札を出し切れること。
//
// 簡略化 (GoDoc に明記): 本来 2〜3 カナスタを要する上がり条件はデフォルト 1+1
// (設定変更可)。初回メルド最低点・コンシールド上がりは扱わない。
type HandAndFoot struct {
	trumpCards       *TrumpCards
	players          []*HandAndFootPlayer
	teamMelds        [HandAndFootTeamCnt][]*CanastaMeld
	teamRed3s        [HandAndFootTeamCnt][]*Card
	config           HandAndFootConfig
	phase            HandAndFootPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	isFrozen         bool
	gameEndFlag      bool
	winnerTeam       int
	roundNumber      int
	actionLogBase
	drewFromDiscard bool
	drawnCard       *Card
}

// NewHandAndFoot コンストラクタ
func NewHandAndFoot(trumpCards *TrumpCards, players []*HandAndFootPlayer, config HandAndFootConfig) *HandAndFoot {
	return &HandAndFoot{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 0,
	}
}

// NewDefaultHandAndFoot returns HandAndFoot with the standard 4-player setup
// (seat 0 human, seats 1-3 CPU) using a 4-deck pack with 8 jokers and
// DefaultHandAndFootConfig. Single source of truth for CUI/Web/Worker.
func NewDefaultHandAndFoot() *HandAndFoot {
	players := []*HandAndFootPlayer{
		NewHandAndFootPlayer(true),
		NewHandAndFootPlayer(false),
		NewHandAndFootPlayer(false),
		NewHandAndFootPlayer(false),
	}
	return NewHandAndFoot(NewTrumpCardsWithDecks(4, 8), players, DefaultHandAndFootConfig())
}

// Reset ゲーム初期化
func (g *HandAndFoot) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.isFrozen = false
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.drewFromDiscard = false
	g.drawnCard = nil

	for t := 0; t < HandAndFootTeamCnt; t++ {
		g.teamMelds[t] = make([]*CanastaMeld, 0)
		g.teamRed3s[t] = make([]*Card, 0)
	}
	for _, p := range g.players {
		p.SetCumulativeScore(0)
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = HandAndFootPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *HandAndFoot) NextRound() {
	if g.phase != HandAndFootPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.isFrozen = false
	g.currentPlayerIdx = 0
	g.drewFromDiscard = false
	g.drawnCard = nil

	for t := 0; t < HandAndFootTeamCnt; t++ {
		g.teamMelds[t] = make([]*CanastaMeld, 0)
		g.teamRed3s[t] = make([]*Card, 0)
	}
	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = HandAndFootPhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに手札22枚 + フット13枚、1枚を捨て札に
func (g *HandAndFoot) dealInitialCards() {
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

	// 手札を配布
	for i := 0; i < HandAndFootHandSize; i++ {
		for j := 0; j < HandAndFootPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// フットを配布 (脇に置く)
	for i := 0; i < HandAndFootFootSize; i++ {
		for j := 0; j < HandAndFootPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddFootCard(card)
			}
		}
	}

	// 手札の赤3を自動的に場に出し、山札から補充
	for j := 0; j < HandAndFootPlayerCnt; j++ {
		g.autoLayRed3s(j)
	}

	// 最初の1枚を捨て札に (赤3やワイルドが出たら次のカードを引く)
	for len(g.drawPile) > 0 {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, card)
		if CanastaIsRed3(card) {
			continue
		}
		if CanastaIsWild(card) {
			g.isFrozen = true
		}
		break
	}
}

// autoLayRed3s プレイヤーの手札から赤3を自動的にチームの赤3置場に出す
func (g *HandAndFoot) autoLayRed3s(playerIdx int) {
	player := g.players[playerIdx]
	team := HandAndFootTeamOf(playerIdx)
	for {
		found := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if CanastaIsRed3(card) {
				player.RemoveCard(i)
				g.teamRed3s[team] = append(g.teamRed3s[team], card)
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

// enterFootIfEmpty 手札を出し切ったらフットを取り込む。取り込んだら true。
func (g *HandAndFoot) enterFootIfEmpty(playerIdx int) bool {
	player := g.players[playerIdx]
	if player.GetCardsSize() != 0 || player.GetInFoot() {
		return false
	}
	foot := player.GetFoot()
	for _, c := range foot {
		player.AddCard(c)
	}
	player.SetFoot(make([]*Card, 0))
	player.SetInFoot(true)
	g.appendLog(playerIdx, "foot", fmt.Sprintf("%s picks up the foot (%d cards)", playerName(g.players, playerIdx), len(foot)), nil)
	g.autoLayRed3s(playerIdx)
	g.sortHand(playerIdx)
	return true
}

// teamCanastaCounts はチームの赤・黒カナスタ数を返す。
func (g *HandAndFoot) teamCanastaCounts(team int) (red, black int) {
	for _, m := range g.teamMelds[team] {
		if !m.IsCanasta() {
			continue
		}
		if m.IsNatural {
			red++
		} else {
			black++
		}
	}
	return red, black
}

// CanastaMinMeld はキャナスタ系の初回メルド最低点を累積得点から返す。
// Web の canastaMinMeld と同じ帯 (マイナス 15 / 1500 未満 50 / 3000 未満 90 /
// それ以上 120)。表示専用で、ドメインはこの最低点を強制しない。
func CanastaMinMeld(cumulativeScore int) int {
	switch {
	case cumulativeScore < 0:
		return 15
	case cumulativeScore < 1500:
		return 50
	case cumulativeScore < 3000:
		return 90
	default:
		return 120
	}
}

// HandAndFootGoOutStatus は上がり条件の充足状況。UI が「なぜ今上がれないか」を
// 説明するために使う (#4836)。
type HandAndFootGoOutStatus struct {
	InFoot       bool
	RedCanastas  int
	RedRequired  int
	BlackCanasta int
	BlackReq     int
}

// CanGoOut は 3 条件をすべて満たしているかを返す。
func (s HandAndFootGoOutStatus) CanGoOut() bool {
	return s.InFoot && s.RedCanastas >= s.RedRequired && s.BlackCanasta >= s.BlackReq
}

// GetGoOutStatus は指定プレイヤーの上がり条件の内訳を返す。canGoOut と同じ判定を
// 使うので、「条件を満たしている」表示と実際に上がれるかがずれない。
func (g *HandAndFoot) GetGoOutStatus(playerIdx int) HandAndFootGoOutStatus {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return HandAndFootGoOutStatus{}
	}
	red, black := g.teamCanastaCounts(HandAndFootTeamOf(playerIdx))
	return HandAndFootGoOutStatus{
		InFoot:       g.players[playerIdx].GetInFoot(),
		RedCanastas:  red,
		RedRequired:  g.config.RedCanastasToGoOut,
		BlackCanasta: black,
		BlackReq:     g.config.BlackCanastasToGoOut,
	}
}

// canGoOut 上がり条件を満たすか。表示側と同じ材料で判定するため
// GetGoOutStatus に委譲する — 2 つの実装が並ぶと「上がれます」と出しながら
// サーバーに拒否される状態が作れてしまう。
func (g *HandAndFoot) canGoOut(playerIdx int) bool {
	return g.GetGoOutStatus(playerIdx).CanGoOut()
}

// PlayerDrawFromStock 人間プレイヤーが山札から2枚引く
func (g *HandAndFoot) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != HandAndFootPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return nil
	}

	player := g.players[g.currentPlayerIdx]
	// 2枚引く (山札が足りなければあるだけ)
	for i := 0; i < 2 && len(g.drawPile) > 0; i++ {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		player.AddCard(card)
		if CanastaIsRed3(card) {
			g.autoLayRed3s(g.currentPlayerIdx)
		}
	}
	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws 2 from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)

	g.phase = HandAndFootPhaseMeld
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札の山を取る。
// トップ + 直下最大6枚 (計7枚) を獲得する。
func (g *HandAndFoot) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != HandAndFootPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札の山が空です")
	}

	topCard := g.discardPile[len(g.discardPile)-1]
	if CanastaIsBlack3(topCard) {
		return NewDomainError(ErrInvalidPlay, "黒3がトップの場合は捨て札の山を取れません")
	}
	if CanastaIsWild(topCard) {
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
	if CanastaIsWild(card0) || CanastaIsWild(card1) {
		return NewDomainError(ErrInvalidPlay, "ペアはナチュラルカード（ワイルドカード以外）でなければなりません")
	}
	if card0.GetValue() != topCard.GetValue() || card1.GetValue() != topCard.GetValue() {
		return NewDomainError(ErrInvalidPlay, "ペアのランクが捨て札のトップカードと一致しません")
	}

	// トップ + 直下最大6枚 (計7枚) を獲得
	g.drawnCard = topCard
	g.drewFromDiscard = true
	taken := g.takeDiscardTop(7)
	for _, c := range taken {
		player.AddCard(c)
	}
	g.isFrozen = false

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s picks up %d cards from the discard pile", playerName(g.players, g.currentPlayerIdx), len(taken)), []*Card{topCard})

	g.autoLayRed3s(g.currentPlayerIdx)
	g.sortHand(g.currentPlayerIdx)

	g.phase = HandAndFootPhaseMeld
	return nil
}

// takeDiscardTop は捨て札の山のトップから最大 n 枚を取り出して返す (上に積まれた順)。
func (g *HandAndFoot) takeDiscardTop(n int) []*Card {
	if n > len(g.discardPile) {
		n = len(g.discardPile)
	}
	start := len(g.discardPile) - n
	taken := make([]*Card, n)
	copy(taken, g.discardPile[start:])
	g.discardPile = g.discardPile[:start]
	return taken
}

// PlayerMeld 人間プレイヤーがメルドを出す (チームのメルドへ)
func (g *HandAndFoot) PlayerMeld(meldGroups [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != HandAndFootPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(meldGroups) == 0 {
		if g.drewFromDiscard {
			return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
		}
		g.phase = HandAndFootPhaseDiscard
		return nil
	}

	player := g.players[g.currentPlayerIdx]
	team := HandAndFootTeamOf(g.currentPlayerIdx)

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

	// 捨て札の山を取った場合、トップカードが少なくとも1つのグループに含まれているか
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

	type meldAction struct {
		isNewMeld   bool
		existingIdx int
		cards       []*Card
	}
	actions := make([]meldAction, 0, len(groups))

	for _, grp := range groups {
		existingIdx := g.findExistingTeamMeldForCards(team, grp.cards)
		if existingIdx >= 0 {
			existing := g.teamMelds[team][existingIdx]
			if err := g.validateMeldAddition(existing, grp.cards); err != nil {
				return err
			}
			actions = append(actions, meldAction{isNewMeld: false, existingIdx: existingIdx, cards: grp.cards})
		} else {
			if err := g.validateNewMeld(grp.cards); err != nil {
				return err
			}
			actions = append(actions, meldAction{isNewMeld: true, existingIdx: -1, cards: grp.cards})
		}
	}

	// メルド実行
	allIdx := make([]int, 0, len(allIndices))
	for k := range allIndices {
		allIdx = append(allIdx, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(allIdx)))

	for _, action := range actions {
		if action.isNewMeld {
			isNatural := true
			for _, c := range action.cards {
				if CanastaIsWild(c) {
					isNatural = false
					break
				}
			}
			meld := &CanastaMeld{Cards: action.cards, IsNatural: isNatural}
			g.teamMelds[team] = append(g.teamMelds[team], meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d cards (rank %d)", playerName(g.players, g.currentPlayerIdx), len(action.cards), meld.GetRank()), action.cards)
		} else {
			existing := g.teamMelds[team][action.existingIdx]
			for _, c := range action.cards {
				existing.Cards = append(existing.Cards, c)
				if CanastaIsWild(c) {
					existing.IsNatural = false
				}
			}
			g.appendLog(g.currentPlayerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to meld (rank %d)", playerName(g.players, g.currentPlayerIdx), len(action.cards), existing.GetRank()), action.cards)
		}
	}

	for _, idx := range allIdx {
		player.RemoveCard(idx)
	}

	g.drewFromDiscard = false
	g.drawnCard = nil

	for _, m := range g.teamMelds[team] {
		if m.IsCanasta() {
			g.appendLog(g.currentPlayerIdx, "canasta", fmt.Sprintf("%s completes a %s canasta!", playerName(g.players, g.currentPlayerIdx), canastaTypeStr(m.IsNatural)), nil)
		}
	}

	// 手札を出し切ったらフットを取り込む
	if g.enterFootIfEmpty(g.currentPlayerIdx) {
		return nil
	}

	g.phase = HandAndFootPhaseDiscard
	return nil
}

// PlayerSkipMeld メルドフェーズをスキップする
func (g *HandAndFoot) PlayerSkipMeld() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != HandAndFootPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.drewFromDiscard {
		return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
	}

	g.phase = HandAndFootPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *HandAndFoot) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != HandAndFootPhaseDiscard {
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
	if CanastaIsRed3(card) {
		return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	if CanastaIsWild(discarded) {
		g.isFrozen = true
	}
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	// 捨て札で手札を出し切ったらフットを取り込む
	g.enterFootIfEmpty(g.currentPlayerIdx)

	g.advanceTurn()
	return nil
}

// PlayerGoOut 人間プレイヤーが上がる
func (g *HandAndFoot) PlayerGoOut() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != HandAndFootPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]
	if !g.canGoOut(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, "上がり条件（フット取り込み・必要なカナスタ）を満たしていません")
	}

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

	g.goOut(g.currentPlayerIdx)
	return nil
}

// goOut 上がり処理
func (g *HandAndFoot) goOut(playerIdx int) {
	g.appendLog(playerIdx, "go_out", fmt.Sprintf("%s goes out! (bonus: %d)", playerName(g.players, playerIdx), HandAndFootGoingOutBonus), nil)
	g.scoreRound(HandAndFootTeamOf(playerIdx))
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *HandAndFoot) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case HandAndFootPhaseDraw:
		g.cpuDraw()
	case HandAndFootPhaseMeld:
		g.cpuMeld()
	case HandAndFootPhaseDiscard:
		g.cpuDiscard()
	}
}

// cpuDraw CPUがドローする
func (g *HandAndFoot) cpuDraw() {
	player := g.players[g.currentPlayerIdx]

	// 捨て札の山を取るか判定
	if len(g.discardPile) > 0 {
		topCard := g.discardPile[len(g.discardPile)-1]
		if !CanastaIsBlack3(topCard) && !CanastaIsWild(topCard) {
			shouldPickUp := false
			switch g.config.CpuDifficulty {
			case HandAndFootCpuDifficultyHard:
				shouldPickUp = g.cpuHasNaturalPair(player, topCard) && len(g.discardPile) >= 3
			case HandAndFootCpuDifficultyNormal:
				shouldPickUp = g.cpuHasNaturalPair(player, topCard)
			default:
				shouldPickUp = rand.Intn(4) == 0 && g.cpuHasNaturalPair(player, topCard)
			}
			if shouldPickUp && g.cpuFindNaturalPair(player, topCard) != nil {
				g.drawnCard = topCard
				g.drewFromDiscard = true
				taken := g.takeDiscardTop(7)
				for _, c := range taken {
					player.AddCard(c)
				}
				g.isFrozen = false
				g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s picks up %d cards from the discard pile", playerName(g.players, g.currentPlayerIdx), len(taken)), []*Card{topCard})
				g.autoLayRed3s(g.currentPlayerIdx)
				g.sortHand(g.currentPlayerIdx)
				g.phase = HandAndFootPhaseMeld
				return
			}
		}
	}

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	for i := 0; i < 2 && len(g.drawPile) > 0; i++ {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		player.AddCard(card)
		if CanastaIsRed3(card) {
			g.autoLayRed3s(g.currentPlayerIdx)
		}
	}
	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws 2 from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)
	g.phase = HandAndFootPhaseMeld
}

// cpuMeld CPUがメルドする
func (g *HandAndFoot) cpuMeld() {
	player := g.players[g.currentPlayerIdx]
	team := HandAndFootTeamOf(g.currentPlayerIdx)

	meldGroups := g.cpuFindMelds(player, team)
	if len(meldGroups) == 0 {
		g.drewFromDiscard = false
		g.drawnCard = nil
		g.phase = HandAndFootPhaseDiscard
		return
	}

	for _, group := range meldGroups {
		existingIdx := g.findExistingTeamMeldForCards(team, group)
		if existingIdx >= 0 {
			existing := g.teamMelds[team][existingIdx]
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
			g.teamMelds[team] = append(g.teamMelds[team], meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds %d cards (rank %d)", playerName(g.players, g.currentPlayerIdx), len(group), meld.GetRank()), group)
		}

		for _, c := range group {
			for i := 0; i < player.GetCardsSize(); i++ {
				if player.GetCard(i) == c {
					player.RemoveCard(i)
					break
				}
			}
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil

	if g.enterFootIfEmpty(g.currentPlayerIdx) {
		return
	}

	g.phase = HandAndFootPhaseDiscard
}

// cpuDiscard CPUがディスカードする
func (g *HandAndFoot) cpuDiscard() {
	player := g.players[g.currentPlayerIdx]

	// 防御的: 手札が0枚 (メルドで出し切ったがフット取り込み未発生など)
	if player.GetCardsSize() == 0 {
		if g.enterFootIfEmpty(g.currentPlayerIdx) {
			g.phase = HandAndFootPhaseDiscard
			return
		}
		if g.canGoOut(g.currentPlayerIdx) {
			g.goOut(g.currentPlayerIdx)
			return
		}
		g.advanceTurn()
		return
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
			g.goOut(g.currentPlayerIdx)
			return
		}
	}

	bestIdx := g.cpuBestDiscard(player)
	discarded := player.RemoveCard(bestIdx)
	g.discardPile = append(g.discardPile, discarded)
	if CanastaIsWild(discarded) {
		g.isFrozen = true
	}
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	g.enterFootIfEmpty(g.currentPlayerIdx)
	g.advanceTurn()
}

// --- CPU AI helpers ---

// cpuHasNaturalPair トップカードと同ランクのナチュラルカードのペアがあるか
func (g *HandAndFoot) cpuHasNaturalPair(player *HandAndFootPlayer, topCard *Card) bool {
	return g.cpuFindNaturalPair(player, topCard) != nil
}

// cpuFindNaturalPair ナチュラルペアのインデックスを返す (見つからなければnil)
func (g *HandAndFoot) cpuFindNaturalPair(player *HandAndFootPlayer, topCard *Card) []int {
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
// SuggestMelds は playerIdx がいま作れるメルド候補 (同ランク 3 枚以上、ワイルド補完込み・
// 既存チームメルドへの追加を含む) をカード群のリストで返す。無ければ nil。CUI ヒント用に
// cpuFindMelds を公開する薄いラッパー。
func (g *HandAndFoot) SuggestMelds(playerIdx int) [][]*Card {
	player := g.GetPlayer(playerIdx)
	if player == nil {
		return nil
	}
	return g.cpuFindMelds(player, HandAndFootTeamOf(playerIdx))
}

func (g *HandAndFoot) cpuFindMelds(player *HandAndFootPlayer, team int) [][]*Card {
	var melds [][]*Card

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
	for _, m := range g.teamMelds[team] {
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
			if len(cards) > 3 {
				melds = append(melds, cards[3:])
			}
		} else if len(cards) == 2 && len(wilds) > 0 {
			meld := []*Card{cards[0], cards[1], wilds[0]}
			wilds = wilds[1:]
			melds = append(melds, meld)
		}
	}

	return melds
}

// cpuBestDiscard CPUが最適なディスカードを選択する
func (g *HandAndFoot) cpuBestDiscard(player *HandAndFootPlayer) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if CanastaIsBlack3(c) {
			return i
		}
	}

	rankCount := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !CanastaIsWild(c) && !CanastaIsRed3(c) {
			rankCount[c.GetValue()]++
		}
	}

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

// --- Scoring ---

// scoreRound ラウンドのスコアを確定する。goOutTeam == -1 のときは上がりなし。
func (g *HandAndFoot) scoreRound(goOutTeam int) {
	teamScores := [HandAndFootTeamCnt]int{}
	for t := 0; t < HandAndFootTeamCnt; t++ {
		score := 0
		for _, m := range g.teamMelds[t] {
			for _, c := range m.Cards {
				score += CanastaCardValue(c)
			}
			if m.IsCanasta() {
				if m.IsNatural {
					score += HandAndFootRedCanastaBonus
				} else {
					score += HandAndFootBlackCanastaBonus
				}
			}
		}
		score += len(g.teamRed3s[t]) * HandAndFootRed3Bonus
		if t == goOutTeam {
			score += HandAndFootGoingOutBonus
		}
		teamScores[t] = score
	}

	// 各メンバーの手札・フットの残りカードを減算
	for i := 0; i < HandAndFootPlayerCnt; i++ {
		player := g.players[i]
		team := HandAndFootTeamOf(i)
		for j := 0; j < player.GetCardsSize(); j++ {
			teamScores[team] -= CanastaCardValue(player.GetCard(j))
		}
		for _, c := range player.GetFoot() {
			teamScores[team] -= CanastaCardValue(c)
		}
	}

	// チームスコアを各メンバーに反映 (累積はチーム共通 = 各メンバー同値)
	for i := 0; i < HandAndFootPlayerCnt; i++ {
		team := HandAndFootTeamOf(i)
		g.players[i].SetRoundScore(teamScores[team])
	}
	for t := 0; t < HandAndFootTeamCnt; t++ {
		g.appendLog(-1, "score", fmt.Sprintf("Team %d scores %d points this round", t, teamScores[t]), nil)
	}
	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = HandAndFootPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによるラウンド終了
func (g *HandAndFoot) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends (stock empty)", nil)
	g.scoreRound(-1)
}

// advanceTurn 次のプレイヤーへ
func (g *HandAndFoot) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % HandAndFootPlayerCnt
	g.drewFromDiscard = false
	g.drawnCard = nil

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	g.phase = HandAndFootPhaseDraw
}

// checkGameEnd ゲーム終了判定
func (g *HandAndFoot) checkGameEnd() {
	hasWinner := false
	for t := 0; t < HandAndFootTeamCnt; t++ {
		if g.teamCumulativeScore(t) >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}
	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = HandAndFootPhaseGameEnd

	maxScore := g.teamCumulativeScore(0)
	g.winnerTeam = 0
	for t := 1; t < HandAndFootTeamCnt; t++ {
		if g.teamCumulativeScore(t) > maxScore {
			maxScore = g.teamCumulativeScore(t)
			g.winnerTeam = t
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", g.winnerTeam), nil)
}

// teamCumulativeScore チームの累積スコア (チームの最初のメンバーで代表)
func (g *HandAndFoot) teamCumulativeScore(team int) int {
	for i := 0; i < HandAndFootPlayerCnt; i++ {
		if HandAndFootTeamOf(i) == team {
			return g.players[i].GetCumulativeScore()
		}
	}
	return 0
}

// --- Meld Validation ---

// validateNewMeld 新規メルドの検証
func (g *HandAndFoot) validateNewMeld(cards []*Card) error {
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
func (g *HandAndFoot) validateMeldAddition(existing *CanastaMeld, cards []*Card) error {
	rank := existing.GetRank()
	existingWildCount := 0
	existingNaturalCount := 0
	for _, c := range existing.Cards {
		if CanastaIsWild(c) {
			existingWildCount++
		} else {
			existingNaturalCount++
		}
	}

	newWildCount := existingWildCount
	newNaturalCount := existingNaturalCount
	for _, c := range cards {
		if CanastaIsWild(c) {
			newWildCount++
		} else if c.GetValue() != rank {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ランク%dのメルドにランク%dのカードは追加できません", rank, c.GetValue()))
		} else {
			newNaturalCount++
		}
		if CanastaIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
		}
	}

	// 7枚到達 (カナスタ完成) 時はワイルド最大3枚。それ以外はナチュラル多数を維持。
	if newWildCount > 3 {
		return NewDomainError(ErrInvalidPlay, "メルドにはワイルドカードは最大3枚までです")
	}
	if newWildCount > newNaturalCount {
		return NewDomainError(ErrInvalidPlay, "ワイルドカードの数がナチュラルカードの数を超えてはいけません")
	}

	return nil
}

// findExistingTeamMeldForCards カードが追加できるチームの既存メルドのインデックスを返す (-1 = なし)
func (g *HandAndFoot) findExistingTeamMeldForCards(team int, cards []*Card) int {
	rank := 0
	for _, c := range cards {
		if !CanastaIsWild(c) {
			rank = c.GetValue()
			break
		}
	}
	if rank == 0 {
		return -1
	}
	for i, m := range g.teamMelds[team] {
		if m.GetRank() == rank {
			return i
		}
	}
	return -1
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *HandAndFoot) GetPhase() HandAndFootPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *HandAndFoot) SetPhase(phase HandAndFootPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *HandAndFoot) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *HandAndFoot) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *HandAndFoot) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *HandAndFoot) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *HandAndFoot) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *HandAndFoot) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *HandAndFoot) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *HandAndFoot) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *HandAndFoot) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetDiscardPileCount 捨て札の枚数取得
func (g *HandAndFoot) GetDiscardPileCount() int { return len(g.discardPile) }

// GetIsFrozen 捨て札の山がフリーズ状態か取得
func (g *HandAndFoot) GetIsFrozen() bool { return g.isFrozen }

// SetIsFrozen フリーズ状態設定 (テスト用)
func (g *HandAndFoot) SetIsFrozen(v bool) { g.isFrozen = v }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *HandAndFoot) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (g *HandAndFoot) SetGameEndFlag(v bool) { g.gameEndFlag = v }

// GetWinnerTeam 勝利チームインデックス取得 (-1 = 未確定)
func (g *HandAndFoot) GetWinnerTeam() int { return g.winnerTeam }

// GetWinnerIdx 勝利チームの代表プレイヤーインデックスを取得 (-1 = 未確定)。
// プレゼンターの汎用処理 (勝者名表示) のために提供する。
func (g *HandAndFoot) GetWinnerIdx() int {
	if g.winnerTeam < 0 {
		return -1
	}
	for i := 0; i < HandAndFootPlayerCnt; i++ {
		if HandAndFootTeamOf(i) == g.winnerTeam {
			return i
		}
	}
	return -1
}

// GetPlayerCnt プレイヤー数取得
func (g *HandAndFoot) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *HandAndFoot) GetPlayer(i int) *HandAndFootPlayer {
	return getPlayer(g.players, i)
}

// GetTeamMelds 指定チームのメルドを取得
func (g *HandAndFoot) GetTeamMelds(team int) []*CanastaMeld {
	if team < 0 || team >= HandAndFootTeamCnt {
		return nil
	}
	return g.teamMelds[team]
}

// SetTeamMelds 指定チームのメルドを設定 (テスト用)
func (g *HandAndFoot) SetTeamMelds(team int, melds []*CanastaMeld) {
	if team < 0 || team >= HandAndFootTeamCnt {
		return
	}
	g.teamMelds[team] = melds
}

// GetTeamRed3s 指定チームの赤3を取得
func (g *HandAndFoot) GetTeamRed3s(team int) []*Card {
	if team < 0 || team >= HandAndFootTeamCnt {
		return nil
	}
	return g.teamRed3s[team]
}

// SetTeamRed3s 指定チームの赤3を設定 (テスト用)
func (g *HandAndFoot) SetTeamRed3s(team int, red3s []*Card) {
	if team < 0 || team >= HandAndFootTeamCnt {
		return
	}
	g.teamRed3s[team] = red3s
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *HandAndFoot) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *HandAndFoot) GetConfig() HandAndFootConfig { return g.config }

// SetConfig 設定変更
func (g *HandAndFoot) SetConfig(cfg HandAndFootConfig) { g.config = cfg }

// GetDrewFromDiscard 捨て札から引いたか取得
func (g *HandAndFoot) GetDrewFromDiscard() bool { return g.drewFromDiscard }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *HandAndFoot) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *HandAndFoot) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- JSON serialization ---

// handAndFootJSON is the JSON wire format for HandAndFoot.
type handAndFootJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*HandAndFootPlayer `json:"pl"`
	Team0Melds       []*CanastaMeld       `json:"m0,omitempty"`
	Team1Melds       []*CanastaMeld       `json:"m1,omitempty"`
	Team0Red3s       []*Card              `json:"r0,omitempty"`
	Team1Red3s       []*Card              `json:"r1,omitempty"`
	Config           HandAndFootConfig    `json:"cf"`
	Phase            HandAndFootPhase     `json:"ps"`
	CurrentPlayerIdx int                  `json:"ci"`
	DiscardPile      []*Card              `json:"dp"`
	DrawPile         []*Card              `json:"wp"`
	IsFrozen         bool                 `json:"fr"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerTeam       int                  `json:"wt"`
	RoundNumber      int                  `json:"rn"`
	ActionLog        []*ActionLogEntry    `json:"al"`
	DrewFromDiscard  bool                 `json:"dd"`
}

// MarshalJSON implements json.Marshaler.
func (g *HandAndFoot) MarshalJSON() ([]byte, error) {
	return json.Marshal(handAndFootJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Team0Melds:       g.teamMelds[0],
		Team1Melds:       g.teamMelds[1],
		Team0Red3s:       g.teamRed3s[0],
		Team1Red3s:       g.teamRed3s[1],
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		IsFrozen:         g.isFrozen,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		DrewFromDiscard:  g.drewFromDiscard,
	})
}

// handAndFootMaxSliceLen caps slice sizes during deserialisation.
const handAndFootMaxSliceLen = 2000

// UnmarshalJSON implements json.Unmarshaler with bounds/state validation.
func (g *HandAndFoot) UnmarshalJSON(data []byte) error {
	var j handAndFootJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > handAndFootMaxSliceLen || len(j.DiscardPile) > handAndFootMaxSliceLen ||
		len(j.DrawPile) > handAndFootMaxSliceLen || len(j.ActionLog) > handAndFootMaxSliceLen ||
		len(j.Team0Melds) > handAndFootMaxSliceLen || len(j.Team1Melds) > handAndFootMaxSliceLen {
		return fmt.Errorf("handandfoot: input array exceeds maximum allowed size")
	}
	if j.Phase < HandAndFootPhaseDraw || j.Phase > HandAndFootPhaseGameEnd {
		return fmt.Errorf("handandfoot: invalid phase")
	}
	if len(j.Players) != HandAndFootPlayerCnt {
		return fmt.Errorf("handandfoot: invalid player count")
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("handandfoot: player cannot be nil")
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= len(j.Players) {
		return fmt.Errorf("handandfoot: current player index out of range")
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= HandAndFootTeamCnt {
		return fmt.Errorf("handandfoot: invalid winner team")
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsWithDecks(4, 8)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*HandAndFootPlayer, 0)
	}
	g.teamMelds[0] = j.Team0Melds
	if g.teamMelds[0] == nil {
		g.teamMelds[0] = make([]*CanastaMeld, 0)
	}
	g.teamMelds[1] = j.Team1Melds
	if g.teamMelds[1] == nil {
		g.teamMelds[1] = make([]*CanastaMeld, 0)
	}
	g.teamRed3s[0] = j.Team0Red3s
	if g.teamRed3s[0] == nil {
		g.teamRed3s[0] = make([]*Card, 0)
	}
	g.teamRed3s[1] = j.Team1Red3s
	if g.teamRed3s[1] == nil {
		g.teamRed3s[1] = make([]*Card, 0)
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
	g.isFrozen = j.IsFrozen
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.drewFromDiscard = j.DrewFromDiscard
	return nil
}
