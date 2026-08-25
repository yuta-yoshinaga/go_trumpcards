//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// BoliviaPlayerCnt ボリビアのプレイヤー数 (2チーム × 2人のパートナーシップ)
const BoliviaPlayerCnt = 4

// BoliviaTeamCnt ボリビアのチーム数 (席0・2 = チーム0, 席1・3 = チーム1)
const BoliviaTeamCnt = 2

// BoliviaHandSize 初期配布枚数
const BoliviaHandSize = 15

// BoliviaDefaultPointLimit デフォルトの目標スコア。
//
// **サンバの 10000 ではなく 15000。** ボリビアとエスカレラの加点が大きいぶん、
// 1 ラウンドで動く点が跳ね上がるので、区切りも上に置かれている。
const BoliviaDefaultPointLimit = 15000

// BoliviaGoOutRequiredMelds 上がりに必要な完成メルドの数。
const BoliviaGoOutRequiredMelds = 2

// BoliviaCanastaSize は 1 つのメルドが完成する枚数。
const BoliviaCanastaSize = 7

// BoliviaWildMeldMin はワイルドだけのメルドに要る最小枚数。
const BoliviaWildMeldMin = 3

// BoliviaRed3Count 3デッキ中の赤3の総数
const BoliviaRed3Count = 6

// ボリビアスコア定数
const (
	BoliviaGoingOutBonus       = 100 // 上がりボーナス
	BoliviaNaturalCanastaBonus = 500 // ナチュラルカナスタボーナス (セット・ワイルドなし)
	BoliviaMixedCanastaBonus   = 300 // ミックスカナスタボーナス (セット・ワイルドあり)
	// BoliviaEscaleraBonus は **エスカレラ** (ワイルド無しの同スート 7 枚連番)。
	// スペイン語の「梯子」で、**3 とは何の関係も無い** ── #5465 はこれを
	// 「3 だけで作るメルド」と取り違えている。
	BoliviaEscaleraBonus = 1500
	// BoliviaBoliviaBonus は **ボリビア** (ワイルドだけ 7 枚)。ゲーム名の由来で、
	// この形式にしか無い。
	BoliviaBoliviaBonus = 2500
	BoliviaRed3Bonus    = 100 // 赤3のボーナス (1枚あたり)
	BoliviaAllRed3Bonus = 800 // 赤3全枚ボーナス
)

// BoliviaPhase ゲームフェーズ
type BoliviaPhase int

// Boliviaのフェーズ定数
const (
	// BoliviaPhaseDraw ドローフェーズ (山札または捨て札から引く)
	BoliviaPhaseDraw BoliviaPhase = 0
	// BoliviaPhaseMeld メルドフェーズ (メルドを出す/既存メルドに追加する)
	BoliviaPhaseMeld BoliviaPhase = 1
	// BoliviaPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or 上がる)
	BoliviaPhaseDiscard BoliviaPhase = 2
	// BoliviaPhaseRoundEnd ラウンド終了フェーズ
	BoliviaPhaseRoundEnd BoliviaPhase = 3
	// BoliviaPhaseGameEnd ゲーム終了フェーズ
	BoliviaPhaseGameEnd BoliviaPhase = 4
)

// Bolivia ボリビアゲームクラス。カナスタの派生ゲームで、3デッキ(162枚)を使い、
// 同ランクのセットメルドに加えて同スート連番のシーケンスメルド（ボリビア）を作る。
// 4人のパートナーシップ (席0・2 vs 席1・3) でチーム対戦する。
type Bolivia struct {
	trumpCards       *TrumpCards
	players          []*BoliviaPlayer
	config           BoliviaConfig
	phase            BoliviaPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	isFrozen         bool
	gameEndFlag      bool
	winnerIdx        int // 勝利チームのインデックス (0 or 1), -1 = 未確定
	roundNumber      int
	teamScores       []int // チーム累積スコア (length BoliviaTeamCnt)
	actionLogBase
	drewFromDiscard bool  // 現在のターンで捨て札の山から引いたか
	drawnCard       *Card // 捨て札の山のトップカード (メルドバリデーション用)
}

// NewBolivia コンストラクタ
func NewBolivia(trumpCards *TrumpCards, players []*BoliviaPlayer, config BoliviaConfig) *Bolivia {
	return &Bolivia{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		teamScores:  make([]int, BoliviaTeamCnt),
	}
}

// newBoliviaDeck は3デッキ + 6ジョーカー = 162枚のパックを生成する
// (カナスタが NewTrumpCardsWithDecks(2, 4) を使うのを踏襲)。
func newBoliviaDeck() *TrumpCards {
	return NewTrumpCardsWithDecks(3, 6)
}

// NewDefaultBolivia returns Bolivia with the standard 4-player partnership setup
// (seat 0 human, seats 1-3 CPU; teams 0 & 1) using a 3-deck pack with 6 jokers
// and DefaultBoliviaConfig. Single source of truth for CUI, Web, and Worker.
func NewDefaultBolivia() *Bolivia {
	players := make([]*BoliviaPlayer, 0, BoliviaPlayerCnt)
	for i := 0; i < BoliviaPlayerCnt; i++ {
		players = append(players, NewBoliviaPlayer(i == 0, i%BoliviaTeamCnt))
	}
	return NewBolivia(newBoliviaDeck(), players, DefaultBoliviaConfig())
}

// Reset ゲーム初期化
func (g *Bolivia) Reset() {
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
	g.teamScores = make([]int, BoliviaTeamCnt)

	for i, p := range g.players {
		p.team = i % BoliviaTeamCnt
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.Reset()
		p.SetIsFinished(false)
		p.melds = make([]*BoliviaMeld, 0)
		p.red3s = make([]*Card, 0)
		p.hasInitMeld = false
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = BoliviaPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Bolivia) NextRound() {
	if g.phase != BoliviaPhaseRoundEnd {
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

	g.phase = BoliviaPhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに15枚、1枚を捨て札に
func (g *Bolivia) dealInitialCards() {
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

	for i := 0; i < BoliviaHandSize; i++ {
		for j := 0; j < BoliviaPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// 赤3の処理: 手札の赤3を自動的に場に出し、山札から補充
	for j := 0; j < BoliviaPlayerCnt; j++ {
		g.autoLayRed3s(j)
	}

	// 最初の1枚を捨て札に (赤3やワイルドが出たら次のカードを引く)
	for len(g.drawPile) > 0 {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, card)
		if BoliviaIsRed3(card) {
			continue
		}
		if BoliviaIsWild(card) {
			g.isFrozen = true
		}
		break
	}
}

// autoLayRed3s プレイヤーの手札から赤3を自動的に場に出す
func (g *Bolivia) autoLayRed3s(playerIdx int) {
	player := g.players[playerIdx]
	for {
		found := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if BoliviaIsRed3(card) {
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

// teamCompletedCount チームの完成メルド（7枚以上のカナスタ/ボリビア）の合計数
func (g *Bolivia) teamCompletedCount(team int) int {
	n := 0
	for _, p := range g.players {
		if p.team == team {
			n += p.CompletedMeldCount()
		}
	}
	return n
}

// canGoOut 上がり条件。
//
// **完成メルド 2 つでは足りない ── うち最低 1 つがエスカレラであること。**
// サンバは「サンバ 2 つ / カナスタ 2 つ / 1 つずつ」のどれでもよいが、
// ボリビアはワイルド無しの同スート 7 枚連番を必ず 1 本要求する。#5465 は
// この条件を「規定数のカナスタ」としか書いておらず、いちばん効く縛りを
// 落としている。
func (g *Bolivia) canGoOut(playerIdx int) bool {
	team := g.players[playerIdx].team
	return g.teamCompletedCount(team) >= BoliviaGoOutRequiredMelds && g.teamHasEscalera(team)
}

// teamHasEscalera はチームが完成したエスカレラを持っているかを返す。
func (g *Bolivia) teamHasEscalera(team int) bool {
	for _, p := range g.players {
		if p.team != team {
			continue
		}
		for _, m := range p.GetMelds() {
			if m != nil && m.IsEscalera() {
				return true
			}
		}
	}
	return false
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Bolivia) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BoliviaPhaseDraw {
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

	if BoliviaIsRed3(card) {
		g.autoLayRed3s(g.currentPlayerIdx)
		if len(g.drawPile) == 0 && g.players[g.currentPlayerIdx].GetCardsSize() == 0 {
			g.endRoundDraw()
			return nil
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)

	g.phase = BoliviaPhaseMeld
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札の山を取る。カナスタと同様に
// トップカードと同ランクのナチュラルカードのペアを手札から示す必要がある。
func (g *Bolivia) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BoliviaPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札の山が空です")
	}

	topCard := g.discardPile[len(g.discardPile)-1]

	if BoliviaIsBlack3(topCard) {
		return NewDomainError(ErrInvalidPlay, "黒3がトップの場合は捨て札の山を取れません")
	}
	if BoliviaIsWild(topCard) {
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

	if BoliviaIsWild(card0) || BoliviaIsWild(card1) {
		return NewDomainError(ErrInvalidPlay, "ペアはナチュラルカード（ワイルドカード以外）でなければなりません")
	}
	if card0.GetValue() != topCard.GetValue() || card1.GetValue() != topCard.GetValue() {
		return NewDomainError(ErrInvalidPlay, "ペアのランクが捨て札のトップカードと一致しません")
	}

	if !player.hasInitMeld {
		meldValue := BoliviaCardValue(topCard) + BoliviaCardValue(card0) + BoliviaCardValue(card1)
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

	g.phase = BoliviaPhaseMeld
	return nil
}

// boliviaMeldResolution はメルドグループの解決結果を表す。
type boliviaMeldResolution struct {
	isNew       bool
	existingIdx int
	kind        BoliviaMeldKind
}

// PlayerMeld 人間プレイヤーがメルドを出す
func (g *Bolivia) PlayerMeld(meldGroups [][]int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BoliviaPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(meldGroups) == 0 {
		if g.drewFromDiscard {
			return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
		}
		g.phase = BoliviaPhaseDiscard
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

	resolutions := make([]boliviaMeldResolution, 0, len(groups))
	for _, grp := range groups {
		res, err := g.resolveMeldGroup(g.currentPlayerIdx, grp.cards)
		if err != nil {
			return err
		}
		resolutions = append(resolutions, res)
		if isInitialMeld {
			for _, c := range grp.cards {
				totalMeldValue += BoliviaCardValue(c)
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

	g.phase = BoliviaPhaseDiscard
	return nil
}

// resolveMeldGroup はカードグループを新規メルドまたは既存メルドへの追加として
// 解決し、そのメルド種別を判定する。ワイルドはシーケンスに使えない。
func (g *Bolivia) resolveMeldGroup(playerIdx int, cards []*Card) (boliviaMeldResolution, error) {
	player := g.players[playerIdx]

	// **ワイルドだけの組はいちばん先に見る。** 後回しにすると
	// boliviaGroupIsSetShaped がワイルドを読み飛ばして「セット」と判定し、
	// 7 枚揃ってもボリビア (2500) ではなくミックスカナスタ (300) として
	// 数えられてしまう。
	if BoliviaIsWildOnly(cards) {
		for i, m := range player.melds {
			if m.Kind != BoliviaMeldWild {
				continue
			}
			// **完成したボリビアには足せない。** 7 枚でそれきり。
			if m.IsBoliviaCanasta() {
				return boliviaMeldResolution{}, NewDomainError(ErrInvalidPlay,
					"完成したボリビアにカードは追加できません")
			}
			return boliviaMeldResolution{isNew: false, existingIdx: i, kind: BoliviaMeldWild}, nil
		}
		if len(cards) < 3 {
			return boliviaMeldResolution{}, NewDomainError(ErrInvalidPlay,
				"メルドには最低3枚のカードが必要です")
		}
		return boliviaMeldResolution{isNew: true, existingIdx: -1, kind: BoliviaMeldWild}, nil
	}

	if boliviaGroupIsSetShaped(cards) {
		rank := boliviaNaturalRank(cards)
		if rank != 0 {
			for i, m := range player.melds {
				if m.Kind == BoliviaMeldSet && m.GetRank() == rank {
					if err := g.validateSetAddition(m, cards); err != nil {
						return boliviaMeldResolution{}, err
					}
					return boliviaMeldResolution{isNew: false, existingIdx: i, kind: BoliviaMeldSet}, nil
				}
			}
			// セットとして出せないが、既存シーケンスを延長できるなら追加とみなす
			for i, m := range player.melds {
				if m.Kind == BoliviaMeldEscalera && g.validateEscaleraAddition(m, cards) == nil {
					return boliviaMeldResolution{isNew: false, existingIdx: i, kind: BoliviaMeldEscalera}, nil
				}
			}
		}
		if err := g.validateNewSet(cards); err != nil {
			return boliviaMeldResolution{}, err
		}
		return boliviaMeldResolution{isNew: true, existingIdx: -1, kind: BoliviaMeldSet}, nil
	}

	// シーケンス形状
	suit := boliviaEscaleraSuit(cards)
	if suit >= 0 {
		for i, m := range player.melds {
			if m.Kind == BoliviaMeldEscalera && m.SuitDesign() == suit {
				if err := g.validateEscaleraAddition(m, cards); err == nil {
					return boliviaMeldResolution{isNew: false, existingIdx: i, kind: BoliviaMeldEscalera}, nil
				}
			}
		}
	}
	if err := g.validateNewEscalera(cards); err != nil {
		return boliviaMeldResolution{}, err
	}
	return boliviaMeldResolution{isNew: true, existingIdx: -1, kind: BoliviaMeldEscalera}, nil
}

// applyResolvedMeld は解決済みメルドをプレイヤーの場に反映する（手札削除は呼び出し側）。
func (g *Bolivia) applyResolvedMeld(playerIdx int, res boliviaMeldResolution, cards []*Card) {
	player := g.players[playerIdx]
	if res.isNew {
		isNatural := true
		for _, c := range cards {
			if BoliviaIsWild(c) {
				isNatural = false
				break
			}
		}
		meld := &BoliviaMeld{Cards: cards, Kind: res.kind, IsNatural: isNatural}
		player.AddMeld(meld)
		g.appendLog(playerIdx, "meld", fmt.Sprintf("%s melds a %s of %d cards", playerName(g.players, playerIdx), boliviaMeldKindStr(res.kind), len(cards)), cards)
		return
	}
	existing := player.melds[res.existingIdx]
	for _, c := range cards {
		existing.Cards = append(existing.Cards, c)
		if BoliviaIsWild(c) {
			existing.IsNatural = false
		}
	}
	g.appendLog(playerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to a %s meld", playerName(g.players, playerIdx), len(cards), boliviaMeldKindStr(existing.Kind)), cards)
}

// logCompletedMelds 完成したカナスタ/ボリビアをログに記録する
func (g *Bolivia) logCompletedMelds(playerIdx int) {
	player := g.players[playerIdx]
	for _, m := range player.melds {
		if m.IsEscalera() {
			g.appendLog(playerIdx, "bolivia", fmt.Sprintf("%s completes a bolivia!", playerName(g.players, playerIdx)), nil)
		} else if m.IsCanasta() {
			g.appendLog(playerIdx, "canasta", fmt.Sprintf("%s completes a %s canasta!", playerName(g.players, playerIdx), boliviaCanastaTypeStr(m.IsNatural)), nil)
		}
	}
}

// PlayerSkipMeld メルドフェーズをスキップする
func (g *Bolivia) PlayerSkipMeld() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BoliviaPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.drewFromDiscard {
		return NewDomainError(ErrInvalidPlay, "捨て札の山を取った場合はトップカードをメルドに含める必要があります")
	}

	g.phase = BoliviaPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *Bolivia) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BoliviaPhaseDiscard {
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
	if BoliviaIsRed3(card) {
		return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	if BoliviaIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	g.advanceTurn()
	return nil
}

// PlayerGoOut 人間プレイヤーが上がる
func (g *Bolivia) PlayerGoOut() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != BoliviaPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]

	if !g.canGoOut(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("上がるにはチームで%d個以上のカナスタ/ボリビアが必要です", BoliviaGoOutRequiredMelds))
	}

	if player.GetCardsSize() == 1 {
		card := player.GetCard(0)
		if BoliviaIsRed3(card) {
			return NewDomainError(ErrInvalidPlay, "赤3は捨てられません")
		}
		discarded := player.RemoveCard(0)
		g.discardPile = append(g.discardPile, discarded)
		if BoliviaIsWild(discarded) {
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
func (g *Bolivia) goOut(playerIdx int) {
	bonus := BoliviaGoingOutBonus
	g.appendLog(playerIdx, "go_out", fmt.Sprintf("%s goes out! (bonus: %d)", playerName(g.players, playerIdx), bonus), nil)
	g.scoreRound(playerIdx, bonus)
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *Bolivia) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case BoliviaPhaseDraw:
		g.cpuDraw()
	case BoliviaPhaseMeld:
		g.cpuMeld()
	case BoliviaPhaseDiscard:
		g.cpuDiscard()
	}
}

// cpuDraw CPUがドローする
func (g *Bolivia) cpuDraw() {
	player := g.players[g.currentPlayerIdx]

	if len(g.discardPile) > 0 {
		topCard := g.discardPile[len(g.discardPile)-1]
		if !BoliviaIsBlack3(topCard) && !BoliviaIsWild(topCard) {
			shouldPickUp := false
			switch g.config.CpuDifficulty {
			case BoliviaCpuDifficultyHard:
				shouldPickUp = g.cpuHasNaturalPair(player, topCard) && len(g.discardPile) >= 3
			case BoliviaCpuDifficultyNormal:
				shouldPickUp = g.cpuHasNaturalPair(player, topCard)
			default:
				shouldPickUp = rand.Intn(4) == 0 && g.cpuHasNaturalPair(player, topCard)
			}

			if shouldPickUp {
				pairIndices := g.cpuFindNaturalPair(player, topCard)
				if pairIndices != nil {
					canPickUp := true
					if !player.hasInitMeld {
						meldValue := BoliviaCardValue(topCard) + BoliviaCardValue(player.GetCard(pairIndices[0])) + BoliviaCardValue(player.GetCard(pairIndices[1]))
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
						g.phase = BoliviaPhaseMeld
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

	if BoliviaIsRed3(card) {
		g.autoLayRed3s(g.currentPlayerIdx)
		if len(g.drawPile) == 0 && player.GetCardsSize() == 0 {
			g.endRoundDraw()
			return
		}
	}

	g.drewFromDiscard = false
	g.drawnCard = nil
	g.sortHand(g.currentPlayerIdx)
	g.phase = BoliviaPhaseMeld
}

// boliviaCpuGroup はCPUのメルド候補（種別と既存メルドへの追加先を含む）。
type boliviaCpuGroup struct {
	cards       []*Card
	kind        BoliviaMeldKind
	existingIdx int // -1 = 新規メルド
}

// cpuMeld CPUがメルドする
func (g *Bolivia) cpuMeld() {
	player := g.players[g.currentPlayerIdx]

	groups := g.cpuFindMelds(player)
	if len(groups) == 0 {
		g.drewFromDiscard = false
		g.drawnCard = nil
		g.phase = BoliviaPhaseDiscard
		return
	}

	if !player.hasInitMeld {
		totalValue := 0
		for _, grp := range groups {
			for _, c := range grp.cards {
				totalValue += BoliviaCardValue(c)
			}
		}
		minReq := g.minimumMeldValue(g.currentPlayerIdx)
		if totalValue < minReq {
			g.drewFromDiscard = false
			g.drawnCard = nil
			g.phase = BoliviaPhaseDiscard
			return
		}
	}

	for _, grp := range groups {
		if grp.existingIdx >= 0 && grp.existingIdx < len(player.melds) {
			existing := player.melds[grp.existingIdx]
			for _, c := range grp.cards {
				existing.Cards = append(existing.Cards, c)
				if BoliviaIsWild(c) {
					existing.IsNatural = false
				}
			}
			g.appendLog(g.currentPlayerIdx, "meld_add", fmt.Sprintf("%s adds %d cards to a %s meld", playerName(g.players, g.currentPlayerIdx), len(grp.cards), boliviaMeldKindStr(existing.Kind)), grp.cards)
		} else {
			isNatural := true
			for _, c := range grp.cards {
				if BoliviaIsWild(c) {
					isNatural = false
					break
				}
			}
			meld := &BoliviaMeld{Cards: grp.cards, Kind: grp.kind, IsNatural: isNatural}
			player.AddMeld(meld)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s melds a %s of %d cards", playerName(g.players, g.currentPlayerIdx), boliviaMeldKindStr(grp.kind), len(grp.cards)), grp.cards)
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

	g.phase = BoliviaPhaseDiscard
}

// cpuDiscard CPUがディスカードする
func (g *Bolivia) cpuDiscard() {
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
		if !BoliviaIsRed3(card) {
			discarded := player.RemoveCard(0)
			g.discardPile = append(g.discardPile, discarded)
			if BoliviaIsWild(discarded) {
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

	if BoliviaIsWild(discarded) {
		g.isFrozen = true
	}

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// --- CPU AI helpers ---

// cpuHasNaturalPair トップカードと同ランクのナチュラルカードのペアがあるか
func (g *Bolivia) cpuHasNaturalPair(player *BoliviaPlayer, topCard *Card) bool {
	return g.cpuFindNaturalPair(player, topCard) != nil
}

// cpuFindNaturalPair ナチュラルペアのインデックスを返す (見つからなければnil)
func (g *Bolivia) cpuFindNaturalPair(player *BoliviaPlayer, topCard *Card) []int {
	var indices []int
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !BoliviaIsWild(c) && c.GetValue() == topCard.GetValue() {
			indices = append(indices, i)
			if len(indices) == 2 {
				return indices
			}
		}
	}
	return nil
}

// cpuFindMelds CPUのメルド候補を見つける。セット（同ランク）とシーケンス
// （同スート連番＝ボリビア）の両方を探す。返すグループのカードは互いに重複しない。
func (g *Bolivia) cpuFindMelds(player *BoliviaPlayer) []boliviaCpuGroup {
	var groups []boliviaCpuGroup
	used := make(map[*Card]bool)

	// **エスカレラを先に取る。** 上がるには最低 1 本要るので、連番になる札を
	// セットに食わせてしまうとチームが永久に上がれない ── 実測で 60 局すべてが
	// 「誰も上がれないまま山切れ」で終わっていた。
	groups = append(groups, g.cpuFindEscaleraGroups(player, used)...)
	for _, grp := range groups {
		for _, c := range grp.cards {
			used[c] = true
		}
	}

	byRank := make(map[int][]*Card)
	var wilds []*Card
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if used[c] {
			continue
		}
		if BoliviaIsWild(c) {
			wilds = append(wilds, c)
		} else if !BoliviaIsRed3(c) && !BoliviaIsBlack3(c) {
			byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
		}
	}

	// **ワイルドが余っているならワイルドだけのメルドを作る。** これが 7 枚で
	// 「ボリビア」= 2500 点、この形式で最も重い加点。セットの穴埋めに 1 枚ずつ
	// 配ってしまうと、名前になっている役に一生届かない。
	if len(wilds) >= BoliviaWildMeldMin {
		groups = append(groups, boliviaCpuGroup{cards: wilds, kind: BoliviaMeldWild, existingIdx: -1})
		for _, c := range wilds {
			used[c] = true
		}
		wilds = nil
	}
	for idx, m := range player.melds {
		if m.Kind == BoliviaMeldWild && !m.IsBoliviaCanasta() && len(wilds) > 0 {
			groups = append(groups, boliviaCpuGroup{cards: wilds, kind: BoliviaMeldWild, existingIdx: idx})
			for _, c := range wilds {
				used[c] = true
			}
			wilds = nil
		}
	}

	// 既存セットメルドへの追加
	for idx, m := range player.melds {
		if m.Kind != BoliviaMeldSet {
			continue
		}
		rank := m.GetRank()
		if cards, ok := byRank[rank]; ok && len(cards) > 0 {
			groups = append(groups, boliviaCpuGroup{cards: cards, kind: BoliviaMeldSet, existingIdx: idx})
			for _, c := range cards {
				used[c] = true
			}
			delete(byRank, rank)
		}
	}

	// 新規セットメルド (同ランク3枚以上、または2枚+ワイルド)
	for _, cards := range byRank {
		if len(cards) >= 3 {
			// **同ランクは 1 つの組にまとめて出す。** 3 枚とあまりに割ると、
			// あまりが 1〜2 枚の「新規メルド」として提案され、3 枚未満の
			// メルドが場に残る (Samba から引き継いだ不具合。あちらでも
			// 1405 個中 138 個が 3 枚未満だった)。
			groups = append(groups, boliviaCpuGroup{cards: cards, kind: BoliviaMeldSet, existingIdx: -1})
			for _, c := range cards {
				used[c] = true
			}
		} else if len(cards) == 2 && len(wilds) > 0 {
			meld := []*Card{cards[0], cards[1], wilds[0]}
			wilds = wilds[1:]
			groups = append(groups, boliviaCpuGroup{cards: meld, kind: BoliviaMeldSet, existingIdx: -1})
			for _, c := range meld {
				used[c] = true
			}
		}
	}

	return groups
}

// cpuFindEscaleraGroups は未使用カードから同スート連番のシーケンス候補を探す。
// 既存シーケンスメルドの延長と新規シーケンスの両方を返す。
func (g *Bolivia) cpuFindEscaleraGroups(player *BoliviaPlayer, used map[*Card]bool) []boliviaCpuGroup {
	var groups []boliviaCpuGroup

	// スートごとに未使用カードを集める (ワイルド・3は除外)
	bySuit := make(map[int][]*Card)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if used[c] || BoliviaIsWild(c) || c.GetValue() == 3 {
			continue
		}
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c)
	}

	// 既存シーケンスメルドの延長
	for idx, m := range player.melds {
		if m.Kind != BoliviaMeldEscalera {
			continue
		}
		suit := m.SuitDesign()
		pool := bySuit[suit]
		if len(pool) == 0 {
			continue
		}
		vals := boliviaEscaleraValues(m.Cards)
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
				v := boliviaEscaleraValue(c)
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
			groups = append(groups, boliviaCpuGroup{cards: added, kind: BoliviaMeldEscalera, existingIdx: idx})
		}
		bySuit[suit] = boliviaFilterUnused(pool, used)
	}

	// 新規シーケンス (未使用カードから長さ3以上の連番を作る)
	for _, pool := range bySuit {
		cards := boliviaFilterUnused(pool, used)
		sort.Slice(cards, func(i, j int) bool {
			return boliviaEscaleraValue(cards[i]) < boliviaEscaleraValue(cards[j])
		})
		run := make([]*Card, 0, len(cards))
		flush := func() {
			if len(run) >= 3 {
				grp := make([]*Card, len(run))
				copy(grp, run)
				groups = append(groups, boliviaCpuGroup{cards: grp, kind: BoliviaMeldEscalera, existingIdx: -1})
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
			prev := boliviaEscaleraValue(cards[i-1])
			cur := boliviaEscaleraValue(c)
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

// boliviaFilterUnused はまだ使われていないカードのみを返す。
func boliviaFilterUnused(cards []*Card, used map[*Card]bool) []*Card {
	out := cards[:0:0]
	for _, c := range cards {
		if !used[c] {
			out = append(out, c)
		}
	}
	return out
}

// cpuBestDiscard CPUが最適なディスカードを選択する
func (g *Bolivia) cpuBestDiscard(player *BoliviaPlayer) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if BoliviaIsBlack3(c) {
			return i
		}
	}

	if g.config.CpuDifficulty == BoliviaCpuDifficultyEasy {
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if !BoliviaIsRed3(c) && !BoliviaIsWild(c) {
				return i
			}
		}
		return 0
	}

	return g.cpuBestDiscardSmart(player)
}

// cpuBestDiscardSmart Normal/Hard: 孤立した低得点カードを捨てる
func (g *Bolivia) cpuBestDiscardSmart(player *BoliviaPlayer) int {
	rankCount := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if !BoliviaIsWild(c) && !BoliviaIsRed3(c) {
			rankCount[c.GetValue()]++
		}
	}

	bestIdx := -1
	bestValue := 1000
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if BoliviaIsRed3(c) || BoliviaIsWild(c) {
			continue
		}
		if rankCount[c.GetValue()] == 1 {
			val := BoliviaCardValue(c)
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
		if BoliviaIsRed3(c) || BoliviaIsWild(c) {
			continue
		}
		val := BoliviaCardValue(c)
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
func (g *Bolivia) scoreRound(goOutPlayerIdx int, goOutBonus int) {
	teamRound := make([]int, BoliviaTeamCnt)

	for i, player := range g.players {
		score := 0
		for _, m := range player.melds {
			for _, c := range m.Cards {
				score += BoliviaCardValue(c)
			}
			if m.IsCanasta() {
				if m.IsNatural {
					score += BoliviaNaturalCanastaBonus
				} else {
					score += BoliviaMixedCanastaBonus
				}
			}
			if m.IsEscalera() {
				score += BoliviaEscaleraBonus
			}
			// **ボリビアはこの形式で最も重い加点。** ワイルド 7 枚を貯める
			// のは手札を丸ごと 1 つの賭けに使うことなので、それに見合う。
			if m.IsBoliviaCanasta() {
				score += BoliviaBoliviaBonus
			}
		}

		// 赤3はチームが初回メルドを完了している場合のみプラス。ラウンド終了時に
		// チームが一度もメルドしていなければ赤3の点は減算される (ボリビア/カナスタの標準ルール)。
		teamMelded := false
		for _, tp := range g.players {
			if tp.team == player.team && tp.hasInitMeld {
				teamMelded = true
				break
			}
		}
		red3Count := len(player.red3s)
		red3Score := red3Count * BoliviaRed3Bonus
		if red3Count >= BoliviaRed3Count {
			red3Score = BoliviaAllRed3Bonus
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
			score -= BoliviaCardValue(player.GetCard(j))
		}

		team := player.team
		if team < 0 || team >= BoliviaTeamCnt {
			team = i % BoliviaTeamCnt
		}
		teamRound[team] += score
		g.appendLog(i, "score", fmt.Sprintf("%s contributes %d points to team %d", playerName(g.players, i), score, team), nil)
	}

	for t := 0; t < BoliviaTeamCnt; t++ {
		g.teamScores[t] += teamRound[t]
	}

	// チーム合計を各プレイヤーの表示スコアに反映する
	for _, player := range g.players {
		team := player.team
		if team < 0 || team >= BoliviaTeamCnt {
			continue
		}
		player.SetRoundScore(teamRound[team])
		player.SetCumulativeScore(g.teamScores[team])
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = BoliviaPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによるラウンド終了
func (g *Bolivia) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends (stock empty)", nil)
	g.scoreRound(-1, 0)
}

// advanceTurn 次のプレイヤーへ
func (g *Bolivia) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % BoliviaPlayerCnt
	g.drewFromDiscard = false
	g.drawnCard = nil

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	g.phase = BoliviaPhaseDraw
}

// checkGameEnd ゲーム終了判定 (チームスコアで判定)
func (g *Bolivia) checkGameEnd() {
	hasWinner := false
	for t := 0; t < BoliviaTeamCnt; t++ {
		if g.teamScores[t] >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}
	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = BoliviaPhaseGameEnd

	maxScore := g.teamScores[0]
	g.winnerIdx = 0
	for t := 1; t < BoliviaTeamCnt; t++ {
		if g.teamScores[t] > maxScore {
			maxScore = g.teamScores[t]
			g.winnerIdx = t
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", g.winnerIdx), nil)
}

// --- Meld Validation ---

// validateNewSet 新規セットメルドの検証
func (g *Bolivia) validateNewSet(cards []*Card) error {
	if len(cards) < 3 {
		return NewDomainError(ErrInvalidPlay, "メルドには最低3枚のカードが必要です")
	}

	var naturalCount, wildCount int
	rank := 0
	for _, c := range cards {
		if BoliviaIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
		}
		if BoliviaIsWild(c) {
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

	// **ワイルドだけのメルドはこの形式でだけ通る。** サンバなら「ナチュラルが
	// 最低 2 枚」で弾かれるところで、ボリビアは 2 とジョーカーだけの組を認め、
	// 7 枚揃うとゲーム名のとおり「ボリビア」になる。
	if naturalCount == 0 {
		return nil
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

// BoliviaIsWildOnly はカードがすべてワイルド (2 / ジョーカー) かを返す。
// 空の組は false ── 何も無いものはメルドではない。
func BoliviaIsWildOnly(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}
	for _, c := range cards {
		if !BoliviaIsWild(c) {
			return false
		}
	}
	return true
}

// validateSetAddition 既存セットメルドへの追加の検証
func (g *Bolivia) validateSetAddition(existing *BoliviaMeld, cards []*Card) error {
	rank := existing.GetRank()
	wildCount := 0
	for _, c := range existing.Cards {
		if BoliviaIsWild(c) {
			wildCount++
		}
	}
	for _, c := range cards {
		if BoliviaIsBlack3(c) {
			return NewDomainError(ErrInvalidPlay, "黒3はメルドできません")
		}
		if BoliviaIsWild(c) {
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

// validateNewEscalera 新規シーケンスメルドの検証 (同スート連番、ワイルド不可)
func (g *Bolivia) validateNewEscalera(cards []*Card) error {
	return boliviaValidateEscaleraCards(cards)
}

// validateEscaleraAddition 既存シーケンスメルドへの追加の検証
func (g *Bolivia) validateEscaleraAddition(existing *BoliviaMeld, cards []*Card) error {
	combined := make([]*Card, 0, len(existing.Cards)+len(cards))
	combined = append(combined, existing.Cards...)
	combined = append(combined, cards...)
	return boliviaValidateEscaleraCards(combined)
}

// boliviaValidateEscaleraCards は一連のカードが有効なシーケンス（同スート連番、
// ワイルド・3を含まない、重複なし）かどうかを検証する。スパンを測る前に
// 必ず値をソートする。
func boliviaValidateEscaleraCards(cards []*Card) error {
	if len(cards) < 3 {
		return NewDomainError(ErrInvalidPlay, "シーケンスには最低3枚のカードが必要です")
	}
	design := -1
	vals := make([]int, 0, len(cards))
	for _, c := range cards {
		if BoliviaIsWild(c) {
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
		vals = append(vals, boliviaEscaleraValue(c))
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
func (g *Bolivia) minimumMeldValue(playerIdx int) int {
	team := g.players[playerIdx].team
	score := 0
	if team >= 0 && team < len(g.teamScores) {
		score = g.teamScores[team]
	}
	return BoliviaMinimumMeldValue(score)
}

// BoliviaMinimumMeldValue はチーム累積点から初回メルドの最低点を返す。
// Web 側は frontend/src/utils/boliviaScore.ts の boliviaMinMeld で同じ表を持っており、
// 両者の一致は internal/infrastructure/games のガードが見る。
func BoliviaMinimumMeldValue(cumulativeScore int) int {
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

// GetMinimumMeldValue は playerIdx のチームに課される初回メルドの最低点を返す。
func (g *Bolivia) GetMinimumMeldValue(playerIdx int) int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return 0
	}
	return g.minimumMeldValue(playerIdx)
}

// --- Card type helpers ---

// BoliviaIsWild ワイルドカードかどうか (ジョーカーまたは2)
func BoliviaIsWild(card *Card) bool {
	return card.GetDesign() == CardDesignJoker || card.GetValue() == 2
}

// BoliviaIsRed3 赤3かどうか (ハート3またはダイヤ3)
func BoliviaIsRed3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignHeart || card.GetDesign() == CardDesignDiamond)
}

// BoliviaIsBlack3 黒3かどうか (スペード3またはクローバー3)
func BoliviaIsBlack3(card *Card) bool {
	return card.GetValue() == 3 && (card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover)
}

// BoliviaCardValue カードの点数を返す
func BoliviaCardValue(card *Card) int {
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

// boliviaEscaleraValue はシーケンス判定用のカード値を返す。エースは高位(14)扱いで
// ラップアラウンドはしない。
func boliviaEscaleraValue(card *Card) int {
	if card.GetValue() == 1 {
		return 14
	}
	return card.GetValue()
}

// boliviaEscaleraValues はカード列のシーケンス値のスライスを返す。
func boliviaEscaleraValues(cards []*Card) []int {
	out := make([]int, 0, len(cards))
	for _, c := range cards {
		if BoliviaIsWild(c) {
			continue
		}
		out = append(out, boliviaEscaleraValue(c))
	}
	return out
}

// boliviaGroupIsSetShaped はグループの全ナチュラルカードが同ランクかどうかを返す。
func boliviaGroupIsSetShaped(cards []*Card) bool {
	rank := 0
	for _, c := range cards {
		if BoliviaIsWild(c) {
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

// boliviaNaturalRank はグループ内の最初のナチュラルカードのランクを返す (なければ0)。
func boliviaNaturalRank(cards []*Card) int {
	for _, c := range cards {
		if !BoliviaIsWild(c) {
			return c.GetValue()
		}
	}
	return 0
}

// boliviaEscaleraSuit はシーケンス形状グループのスートを返す (なければ-1)。
func boliviaEscaleraSuit(cards []*Card) int {
	for _, c := range cards {
		if !BoliviaIsWild(c) {
			return c.GetDesign()
		}
	}
	return -1
}

// boliviaMeldKindStr はメルド種別の文字列を返す。
func boliviaMeldKindStr(kind BoliviaMeldKind) string {
	if kind == BoliviaMeldEscalera {
		return "sequence"
	}
	return "set"
}

// boliviaCanastaTypeStr はカナスタの種別文字列を返す。
func boliviaCanastaTypeStr(isNatural bool) string {
	if isNatural {
		return "natural"
	}
	return "mixed"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Bolivia) GetPhase() BoliviaPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Bolivia) SetPhase(phase BoliviaPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Bolivia) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Bolivia) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Bolivia) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Bolivia) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Bolivia) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Bolivia) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Bolivia) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Bolivia) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Bolivia) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetDiscardPileCount 捨て札の枚数取得
func (g *Bolivia) GetDiscardPileCount() int { return len(g.discardPile) }

// GetIsFrozen 捨て札の山がフリーズ状態か取得
func (g *Bolivia) GetIsFrozen() bool { return g.isFrozen }

// SetIsFrozen フリーズ状態設定 (テスト用)
func (g *Bolivia) SetIsFrozen(v bool) { g.isFrozen = v }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Bolivia) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (g *Bolivia) SetGameEndFlag(v bool) { g.gameEndFlag = v }

// GetWinnerIdx 勝利チームインデックス取得 (-1 = 未確定)
func (g *Bolivia) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Bolivia) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Bolivia) GetPlayer(i int) *BoliviaPlayer {
	return getPlayer(g.players, i)
}

// GetTeamCount チーム数取得
func (g *Bolivia) GetTeamCount() int { return BoliviaTeamCnt }

// GetTeamScore チームの累積スコアを取得
func (g *Bolivia) GetTeamScore(team int) int {
	if team < 0 || team >= len(g.teamScores) {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームの累積スコアを設定 (テスト用)
func (g *Bolivia) SetTeamScore(team, score int) {
	if team >= 0 && team < len(g.teamScores) {
		g.teamScores[team] = score
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Bolivia) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Bolivia) GetConfig() BoliviaConfig { return g.config }

// SetConfig 設定変更
func (g *Bolivia) SetConfig(cfg BoliviaConfig) { g.config = cfg }

// GetDrewFromDiscard 捨て札から引いたか取得
func (g *Bolivia) GetDrewFromDiscard() bool { return g.drewFromDiscard }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Bolivia) sortAllHands() {
	sortHands(len(g.players), g)
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Bolivia) sortHand(playerIdx int) {
	sortPlayerHand(g.players[playerIdx], bySuitThenValue)
}

// --- JSON serialization ---

// boliviaJSON is the JSON wire format for Bolivia.
type boliviaJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*BoliviaPlayer  `json:"pl"`
	Config           BoliviaConfig     `json:"cf"`
	Phase            BoliviaPhase      `json:"ps"`
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
func (g *Bolivia) MarshalJSON() ([]byte, error) {
	return json.Marshal(boliviaJSON{
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

// boliviaMaxSliceLen caps slice sizes during deserialisation.
const boliviaMaxSliceLen = 3000

// UnmarshalJSON implements json.Unmarshaler with defensive validation of all
// indices, team values, phase and meld kinds so a corrupt KV blob cannot panic
// scoring (which indexes teamScores by a restored player's team) or a getter.
func (g *Bolivia) UnmarshalJSON(data []byte) error {
	var j boliviaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > boliviaMaxSliceLen || len(j.DiscardPile) > boliviaMaxSliceLen ||
		len(j.DrawPile) > boliviaMaxSliceLen || len(j.ActionLog) > boliviaMaxSliceLen {
		return fmt.Errorf("bolivia: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newBoliviaDeck()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*BoliviaPlayer, 0)
	}
	for i, p := range g.players {
		if p == nil {
			return fmt.Errorf("bolivia: player %d is nil", i)
		}
	}
	g.config = j.Config

	// フェーズは既知の値のみ許可する
	if j.Phase < BoliviaPhaseDraw || j.Phase > BoliviaPhaseGameEnd {
		g.phase = BoliviaPhaseDraw
	} else {
		g.phase = j.Phase
	}

	// チームスコアは常に長さ BoliviaTeamCnt に正規化する
	g.teamScores = make([]int, BoliviaTeamCnt)
	for t := 0; t < BoliviaTeamCnt && t < len(j.TeamScores); t++ {
		g.teamScores[t] = j.TeamScores[t]
	}

	// 各プレイヤーのチームは [0, BoliviaTeamCnt) に収める。teamScores を
	// インデックスするため、範囲外の値は席順から導出し直す (security)。
	for i, p := range g.players {
		if p == nil {
			continue
		}
		if p.team < 0 || p.team >= BoliviaTeamCnt {
			p.SetTeam(i % BoliviaTeamCnt)
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

	// winnerIdx は -1 (未確定) または [0, BoliviaTeamCnt) のみ許可する
	if j.WinnerIdx < 0 || j.WinnerIdx >= BoliviaTeamCnt {
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
