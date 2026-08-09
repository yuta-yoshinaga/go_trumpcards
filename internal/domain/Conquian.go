//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// ConquianPlayerCnt コンキャンのプレイヤー数 (2人)
const ConquianPlayerCnt = 2

// ConquianHandSize 初期配布枚数 (各プレイヤー10枚)
const ConquianHandSize = 10

// ConquianPhase ゲームフェーズ
type ConquianPhase int

// Conquianのフェーズ定数
const (
	// ConquianPhaseDraw ドローフェーズ (山札または捨て札から引く)
	ConquianPhaseDraw ConquianPhase = 0
	// ConquianPhaseMeld メルドフェーズ (メルドを並べる/付ける → 捨てる)
	ConquianPhaseMeld ConquianPhase = 1
	// ConquianPhaseRoundEnd ラウンド終了フェーズ
	ConquianPhaseRoundEnd ConquianPhase = 2
	// ConquianPhaseGameEnd ゲーム終了フェーズ
	ConquianPhaseGameEnd ConquianPhase = 3
)

// conquianRankPosition は 8・9・10 を除いた40枚デッキでのランクの「隣接位置」を返す。
//
// Conquian は 8/9/10 を抜いた40枚のラテンデッキを使うため、ラン (連続) の判定では
// A,2,3,4,5,6,7,J,Q,K を連続したランクとして扱う。すなわち 7 と J は隣接する。
// 戻り値: A=1,2=2,...,7=7,J=8,Q=9,K=10。デッキに存在しないランク(8/9/10/Joker)は 0。
func conquianRankPosition(value int) int {
	switch {
	case value >= 1 && value <= 7: // A..7
		return value
	case value == 11: // J
		return 8
	case value == 12: // Q
		return 9
	case value == 13: // K
		return 10
	default: // 8,9,10,Joker など (デッキには存在しない)
		return 0
	}
}

// newConquianDeck は 8・9・10 を取り除いた40枚のデッキを構築する。
// 標準52枚デッキ (NewTrumpCards(0)) から rank 8/9/10 のカードを除外する。
func newConquianDeck() []*Card {
	tc := NewTrumpCards(0)
	deck := make([]*Card, 0, 40)
	for {
		c := tc.DrawCard()
		if c == nil {
			break
		}
		v := c.GetValue()
		if v == 8 || v == 9 || v == 10 {
			continue
		}
		deck = append(deck, c)
	}
	return deck
}

// Conquian コンキャンゲームクラス
type Conquian struct {
	players          []*ConquianPlayer
	config           ConquianConfig
	phase            ConquianPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int // ラウンド勝者 (-1 = 未確定/引き分け)
	matchWinnerIdx   int // マッチ全体の勝者 (-1 = 未確定)
	roundNumber      int
	actionLogBase
	tookDiscard bool // 今ターン、捨て札を取って必ずメルドに使う必要があるか
	pendingCard *Card
}

// NewConquian コンストラクタ
func NewConquian(players []*ConquianPlayer, config ConquianConfig) *Conquian {
	return &Conquian{
		players:        players,
		config:         config,
		winnerIdx:      -1,
		matchWinnerIdx: -1,
		roundNumber:    0,
	}
}

// NewDefaultConquian は標準の2人 (人間1, CPU1) 構成と DefaultConquianConfig で
// Conquian を生成する。CUI・Web・Worker の構築サイトの単一情報源。
func NewDefaultConquian() *Conquian {
	players := []*ConquianPlayer{
		NewConquianPlayer(true),
		NewConquianPlayer(false),
	}
	return NewConquian(players, DefaultConquianConfig())
}

// Reset ゲーム初期化
func (g *Conquian) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.matchWinnerIdx = -1
	g.roundNumber = 1
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.tookDiscard = false
	g.pendingCard = nil

	for _, p := range g.players {
		p.SetWins(0)
		p.ResetRound()
	}

	g.dealInitialCards()
	g.phase = ConquianPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *Conquian) NextRound() {
	if g.phase != ConquianPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.winnerIdx = -1
	g.currentPlayerIdx = 0
	g.tookDiscard = false
	g.pendingCard = nil

	for _, p := range g.players {
		p.ResetRound()
	}

	g.dealInitialCards()
	g.phase = ConquianPhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに10枚、残りを山札に。捨て札は空 (最初の手番は山札から引く)。
func (g *Conquian) dealInitialCards() {
	deck := newConquianDeck()
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	g.discardPile = make([]*Card, 0)
	g.drawPile = make([]*Card, 0, len(deck))

	// 各プレイヤーに10枚配布
	idx := 0
	for i := 0; i < ConquianHandSize; i++ {
		for j := 0; j < ConquianPlayerCnt; j++ {
			if idx < len(deck) {
				g.players[j].AddCard(deck[idx])
				idx++
			}
		}
	}
	// 残りを山札へ
	for ; idx < len(deck); idx++ {
		g.drawPile = append(g.drawPile, deck[idx])
	}
	g.sortAllHands()
}

// --- Meld validation ---

// conquianIsValidMeld はカード列が Conquian の有効なメルド (3枚以上のセットまたはラン) かを判定する。
//
//   - セット: 3枚以上、全て同ランク・全て異なるスート (最大4枚)。
//   - ラン: 3枚以上、同スートで conquianRankPosition が連続する。7とJは隣接 (8/9/10を除いたラテンデッキ)。
func conquianIsValidMeld(meld []*Card) bool {
	if len(meld) < 3 {
		return false
	}
	return conquianIsSet(meld) || conquianIsRun(meld)
}

// conquianIsSet はメルドが有効なセット (同ランク・異スート) かを判定する。
func conquianIsSet(meld []*Card) bool {
	if len(meld) < 3 || len(meld) > 4 {
		return false
	}
	rank := meld[0].GetValue()
	seenSuit := make(map[int]bool)
	for _, c := range meld {
		if c.GetValue() != rank {
			return false
		}
		if seenSuit[c.GetDesign()] {
			return false
		}
		seenSuit[c.GetDesign()] = true
	}
	return true
}

// conquianIsRun はメルドが有効なラン (同スート・連続位置) かを判定する。
func conquianIsRun(meld []*Card) bool {
	if len(meld) < 3 {
		return false
	}
	suit := meld[0].GetDesign()
	positions := make([]int, 0, len(meld))
	for _, c := range meld {
		if c.GetDesign() != suit {
			return false
		}
		pos := conquianRankPosition(c.GetValue())
		if pos == 0 {
			return false
		}
		positions = append(positions, pos)
	}
	sort.Ints(positions)
	for i := 1; i < len(positions); i++ {
		if positions[i] != positions[i-1]+1 {
			return false
		}
	}
	return true
}

// conquianCanExtendMeld は既存テーブルメルドに card を追加できるかを判定する。
func conquianCanExtendMeld(meld []*Card, card *Card) bool {
	if len(meld) == 0 {
		return false
	}
	candidate := append(append([]*Card{}, meld...), card)
	return conquianIsValidMeld(candidate)
}

// --- Actions ---

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Conquian) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ConquianPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if len(g.drawPile) == 0 {
		// 山札切れ → 進行不能なので引き分けでラウンド終了
		g.endRoundDraw()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)
	g.tookDiscard = false
	g.pendingCard = nil

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
	g.phase = ConquianPhaseMeld
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札の一番上を引く (強制使用ルール: 即座にメルドに使えなければ違法)
func (g *Conquian) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ConquianPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札がありません")
	}

	card := g.discardPile[len(g.discardPile)-1]
	if !g.canUseCardInMeld(g.currentPlayerIdx, card) {
		return NewDomainError(ErrInvalidPlay, "そのカードはメルドに使えないため捨て札から取れません")
	}

	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)
	g.tookDiscard = true
	g.pendingCard = card

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	g.phase = ConquianPhaseMeld
	return nil
}

// PlayerMeld 人間プレイヤーが手札からメルドを並べる/既存メルドに付ける。
//
// meldGroups は手札インデックスのグループ列。各グループが
//   - 新しいメルド (3枚以上のセット/ラン)、または
//   - 1枚で既存テーブルメルドへの追加 (extend)
//
// のいずれかとして解釈される。空スライスを渡すと「メルドなし」として扱われる。
func (g *Conquian) PlayerMeld(meldGroups [][]int) error {
	return g.PlayerMeldWithTargets(meldGroups, nil)
}

// PlayerMeldWithTargets は PlayerMeld に「1 枚グループをどのメルドへ足すか」の
// 指定を加えたもの。extendTargets[i] は meldGroups[i] の延長先メルド番号で、
// 負値または範囲外なら従来どおり最初に延長できるメルドを選ぶ。
//
// **延長先が一意とは限らない。**♠5 は「5 のセット」も「♠6-7-8 のラン」も延長できる。
// 先頭一致で決め打つと、プレイヤーが意図した側を選べない (#4837)。
func (g *Conquian) PlayerMeldWithTargets(meldGroups [][]int, extendTargets []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ConquianPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(meldGroups) == 0 {
		return g.finishMeldStep()
	}

	player := g.players[g.currentPlayerIdx]

	// 全インデックスの妥当性 (範囲・重複) を先に検証する。
	seen := make(map[int]bool)
	for _, group := range meldGroups {
		if len(group) == 0 {
			return NewDomainError(ErrInvalidPlay, "空のメルドグループは指定できません")
		}
		for _, idx := range group {
			if idx < 0 || idx >= player.GetCardsSize() {
				return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
			}
			if seen[idx] {
				return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
			}
			seen[idx] = true
		}
	}

	// 各グループが有効か (新メルド or extend) を検証する。
	type pendingMeld struct {
		cards     []*Card
		extendIdx int // -1 = 新メルド、>=0 = 既存メルドへの追加
	}
	var pending []pendingMeld
	for gi, group := range meldGroups {
		cards := make([]*Card, 0, len(group))
		for _, idx := range group {
			cards = append(cards, player.GetCard(idx))
		}
		if len(cards) >= 3 && conquianIsValidMeld(cards) {
			pending = append(pending, pendingMeld{cards: cards, extendIdx: -1})
			continue
		}
		if len(cards) == 1 {
			ext := g.findExtendableMeld(g.currentPlayerIdx, cards[0])
			if t := conquianTargetAt(extendTargets, gi); t >= 0 &&
				t < len(g.players[g.currentPlayerIdx].melds) &&
				conquianCanExtendMeld(g.players[g.currentPlayerIdx].melds[t], cards[0]) {
				ext = t
			}
			if ext >= 0 {
				pending = append(pending, pendingMeld{cards: cards, extendIdx: ext})
				continue
			}
		}
		return NewDomainError(ErrInvalidPlay, "無効なメルドです")
	}

	// 強制使用ルール: 捨て札を取った場合、そのカードがどれかのメルドに含まれていなければならない。
	if g.tookDiscard && g.pendingCard != nil {
		used := false
		for _, pm := range pending {
			for _, c := range pm.cards {
				if c == g.pendingCard {
					used = true
				}
			}
		}
		if !used {
			return NewDomainError(ErrInvalidPlay, "捨て札から取ったカードはメルドに使う必要があります")
		}
	}

	// 検証完了 → 実行する。降順にカードを除去して安全に。
	indicesToRemove := make([]int, 0, len(seen))
	for idx := range seen {
		indicesToRemove = append(indicesToRemove, idx)
	}

	for _, pm := range pending {
		if pm.extendIdx >= 0 {
			player.melds[pm.extendIdx] = append(player.melds[pm.extendIdx], pm.cards[0])
			g.appendLog(g.currentPlayerIdx, "extend", fmt.Sprintf("%s extends a meld with %s", playerName(g.players, g.currentPlayerIdx), cardStr(pm.cards[0])), pm.cards)
		} else {
			meldCopy := append([]*Card{}, pm.cards...)
			player.AddMeld(meldCopy)
			g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s lays a meld", playerName(g.players, g.currentPlayerIdx)), meldCopy)
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(indicesToRemove)))
	for _, idx := range indicesToRemove {
		player.RemoveCard(idx)
	}
	g.tookDiscard = false
	g.pendingCard = nil

	// 手札を使い切ったら勝利。
	if player.GetCardsSize() == 0 {
		g.winRound(g.currentPlayerIdx)
		return nil
	}

	return nil
}

// PlayerDiscard 人間プレイヤーが手札から1枚捨ててターンを終了する。
func (g *Conquian) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ConquianPhaseMeld {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if g.tookDiscard {
		return NewDomainError(ErrInvalidPlay, "捨て札から取ったカードを先にメルドに使ってください")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	g.advanceTurn()
	return nil
}

// finishMeldStep はメルドを並べずにメルドフェーズを終える際の検証を行う。
func (g *Conquian) finishMeldStep() error {
	if g.tookDiscard {
		return NewDomainError(ErrInvalidPlay, "捨て札から取ったカードはメルドに使う必要があります")
	}
	// 何もしない (続いて discard が呼ばれる)
	return nil
}

// canUseCardInMeld は card を現在の手札と組み合わせて新メルドにできる、
// または既存テーブルメルドに付けられるかを判定する (強制使用ルールの判定)。
func (g *Conquian) canUseCardInMeld(playerIdx int, card *Card) bool {
	if g.findExtendableMeld(playerIdx, card) >= 0 {
		return true
	}
	player := g.players[playerIdx]
	hand := make([]*Card, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		hand = append(hand, player.GetCard(i))
	}
	return conquianCardFormsMeld(card, hand)
}

// conquianCardFormsMeld は card と hand 内の2枚以上を使って有効なメルドが作れるかを判定する。
func conquianCardFormsMeld(card *Card, hand []*Card) bool {
	// セット: 同ランクが手札に2枚以上 (異スート) あるか
	sameRankSuits := map[int]bool{card.GetDesign(): true}
	for _, c := range hand {
		if c.GetValue() == card.GetValue() && !sameRankSuits[c.GetDesign()] {
			sameRankSuits[c.GetDesign()] = true
		}
	}
	if len(sameRankSuits) >= 3 {
		return true
	}

	// ラン: 同スートで連続位置が card を含む3枚を作れるか
	suit := card.GetDesign()
	posSet := map[int]bool{}
	cp := conquianRankPosition(card.GetValue())
	if cp == 0 {
		return false
	}
	posSet[cp] = true
	for _, c := range hand {
		if c.GetDesign() == suit {
			p := conquianRankPosition(c.GetValue())
			if p != 0 {
				posSet[p] = true
			}
		}
	}
	// card 位置を含む連続3枚 (cp-2..cp, cp-1..cp+1, cp..cp+2) のいずれかが存在するか
	windows := [][]int{
		{cp - 2, cp - 1, cp},
		{cp - 1, cp, cp + 1},
		{cp, cp + 1, cp + 2},
	}
	for _, w := range windows {
		ok := true
		for _, p := range w {
			if p < 1 || p > 10 || !posSet[p] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// findExtendableMeld は card を付けられる現在プレイヤーのテーブルメルドのインデックスを返す (なければ -1)。
func (g *Conquian) findExtendableMeld(playerIdx int, card *Card) int {
	for i, meld := range g.players[playerIdx].melds {
		if conquianCanExtendMeld(meld, card) {
			return i
		}
	}
	return -1
}

// conquianTargetAt は extendTargets の i 番目を返す (無ければ -1)。
func conquianTargetAt(targets []int, i int) int {
	if i < 0 || i >= len(targets) {
		return -1
	}
	return targets[i]
}

// GetExtendableMeldIndices は指定プレイヤーの盤面で、その札を足せるメルドの番号を
// すべて返す。UI が延長先の候補を示すために使う (#4837)。
func (g *Conquian) GetExtendableMeldIndices(playerIdx int, card *Card) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || card == nil {
		return nil
	}
	var out []int
	for i, meld := range g.players[playerIdx].melds {
		if conquianCanExtendMeld(meld, card) {
			out = append(out, i)
		}
	}
	return out
}

// CpuPlay 現在の手番がCPUの場合にターンを実行する
func (g *Conquian) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	switch g.phase {
	case ConquianPhaseDraw:
		g.cpuDraw()
	case ConquianPhaseMeld:
		g.cpuMeldAndDiscard()
	}
}

// cpuDraw CPUがドローする
func (g *Conquian) cpuDraw() {
	idx := g.currentPlayerIdx
	// 捨て札の一番上がメルドに使えるなら取る (強制使用ルールを満たす)
	if len(g.discardPile) > 0 {
		top := g.discardPile[len(g.discardPile)-1]
		if g.canUseCardInMeld(idx, top) {
			g.discardPile = g.discardPile[:len(g.discardPile)-1]
			g.players[idx].AddCard(top)
			g.sortHand(idx)
			g.tookDiscard = true
			g.pendingCard = top
			g.appendLog(idx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, idx), cardStr(top)), []*Card{top})
			g.phase = ConquianPhaseMeld
			return
		}
	}
	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}
	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[idx].AddCard(card)
	g.sortHand(idx)
	g.tookDiscard = false
	g.pendingCard = nil
	g.appendLog(idx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, idx)), nil)
	g.phase = ConquianPhaseMeld
}

// cpuMeldAndDiscard CPUが可能なメルドを並べ、1枚捨てる (または上がる)
func (g *Conquian) cpuMeldAndDiscard() {
	idx := g.currentPlayerIdx
	player := g.players[idx]

	// 可能な限りメルドを並べる/付ける
	g.cpuLayMelds(idx)

	// 上がり判定
	if player.GetCardsSize() == 0 {
		g.winRound(idx)
		return
	}

	g.tookDiscard = false
	g.pendingCard = nil

	// 捨てる: 任意の手札 (デッドウッドが最も役に立たないものを優先せず単純に末尾)
	discardIdx := g.cpuChooseDiscard(idx)
	if discardIdx < 0 {
		// 手札がない (理論上 winRound 済み) ガード
		g.advanceTurn()
		return
	}
	discarded := player.RemoveCard(discardIdx)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(idx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, idx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// cpuLayMelds はCPUが手札から作れる新メルドを並べ、既存メルドへ付けられるカードを付ける。
func (g *Conquian) cpuLayMelds(idx int) {
	player := g.players[idx]
	// 新メルドを貪欲に並べる
	for {
		hand := make([]*Card, 0, player.GetCardsSize())
		for i := 0; i < player.GetCardsSize(); i++ {
			hand = append(hand, player.GetCard(i))
		}
		meld := conquianFindMeld(hand)
		if meld == nil {
			break
		}
		// meld のカードをインデックスで特定して除去
		removeIdx := make([]int, 0, len(meld))
		for _, mc := range meld {
			for i := 0; i < player.GetCardsSize(); i++ {
				if player.GetCard(i) == mc {
					removeIdx = append(removeIdx, i)
					break
				}
			}
		}
		player.AddMeld(append([]*Card{}, meld...))
		g.appendLog(idx, "meld", fmt.Sprintf("%s lays a meld", playerName(g.players, idx)), meld)
		sort.Sort(sort.Reverse(sort.IntSlice(removeIdx)))
		for _, ri := range removeIdx {
			player.RemoveCard(ri)
		}
	}
	// 既存メルドへ extend
	for {
		extended := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			ext := g.findExtendableMeld(idx, card)
			if ext >= 0 {
				player.melds[ext] = append(player.melds[ext], card)
				g.appendLog(idx, "extend", fmt.Sprintf("%s extends a meld with %s", playerName(g.players, idx), cardStr(card)), []*Card{card})
				player.RemoveCard(i)
				extended = true
				break
			}
		}
		if !extended {
			break
		}
	}
}

// cpuChooseDiscard は捨てるカードのインデックスを選ぶ (メルドに絡まないカードを優先)。
func (g *Conquian) cpuChooseDiscard(idx int) int {
	player := g.players[idx]
	if player.GetCardsSize() == 0 {
		return -1
	}
	// メルドを作れないカードを優先して捨てる
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		hand := make([]*Card, 0, player.GetCardsSize()-1)
		for j := 0; j < player.GetCardsSize(); j++ {
			if j != i {
				hand = append(hand, player.GetCard(j))
			}
		}
		if !conquianCardFormsMeld(card, hand) {
			return i
		}
	}
	return player.GetCardsSize() - 1
}

// conquianFindMeld は cards から作れる有効なメルド (最小サイズ3) を1つ返す (なければ nil)。
func conquianFindMeld(cards []*Card) []*Card {
	// セット
	byRank := make(map[int][]*Card)
	for _, c := range cards {
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
	}
	for _, group := range byRank {
		// 異スートのみ抽出
		seen := make(map[int]bool)
		uniq := make([]*Card, 0, len(group))
		for _, c := range group {
			if !seen[c.GetDesign()] {
				seen[c.GetDesign()] = true
				uniq = append(uniq, c)
			}
		}
		if len(uniq) >= 3 {
			return uniq[:3]
		}
	}
	// ラン
	bySuit := make(map[int][]*Card)
	for _, c := range cards {
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c)
	}
	for _, group := range bySuit {
		sorted := append([]*Card{}, group...)
		sort.Slice(sorted, func(i, j int) bool {
			return conquianRankPosition(sorted[i].GetValue()) < conquianRankPosition(sorted[j].GetValue())
		})
		for i := 0; i+2 < len(sorted); i++ {
			run := []*Card{sorted[i]}
			for j := i + 1; j < len(sorted); j++ {
				prev := conquianRankPosition(run[len(run)-1].GetValue())
				cur := conquianRankPosition(sorted[j].GetValue())
				if cur == prev+1 {
					run = append(run, sorted[j])
					if len(run) >= 3 {
						return append([]*Card{}, run[:3]...)
					}
				} else if cur != prev {
					break
				}
			}
		}
	}
	return nil
}

// winRound はプレイヤーがラウンドに勝利した処理を行う。
func (g *Conquian) winRound(playerIdx int) {
	g.winnerIdx = playerIdx
	g.players[playerIdx].AddWin()
	g.players[playerIdx].SetIsFinished(true)
	g.appendLog(playerIdx, "round_win", fmt.Sprintf("%s goes out and wins the round!", playerName(g.players, playerIdx)), nil)
	g.checkMatchEnd()
	if !g.gameEndFlag {
		g.phase = ConquianPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによる引き分け (勝者なし)
func (g *Conquian) endRoundDraw() {
	g.winnerIdx = -1
	g.appendLog(-1, "draw", "Round ends in a draw (stock exhausted)", nil)
	// 引き分けはマッチ勝利数に影響しないが、ゲームは必ず終了させる必要がある。
	// 引き分けが起きたらマッチを終了する (累積勝利数が最大のプレイヤーが勝者、同数なら引き分け)。
	g.endMatchOnDraw()
}

// endMatchOnDraw は山札切れの引き分けでマッチを終了させる。
func (g *Conquian) endMatchOnDraw() {
	g.gameEndFlag = true
	g.phase = ConquianPhaseGameEnd
	// 累積勝利数で勝者を決める。同数なら勝者なし (-1)。
	g.matchWinnerIdx = -1
	best := -1
	tie := false
	for i := 0; i < ConquianPlayerCnt; i++ {
		w := g.players[i].GetWins()
		if w > best {
			best = w
			g.matchWinnerIdx = i
			tie = false
		} else if w == best {
			tie = true
		}
	}
	if tie || best <= 0 {
		g.matchWinnerIdx = -1
	}
	g.appendLog(-1, "game_end", "Match ends (stock exhausted)", nil)
}

// checkMatchEnd はマッチ全体の終了判定を行う (TargetWins に到達したプレイヤーが勝者)。
func (g *Conquian) checkMatchEnd() {
	for i := 0; i < ConquianPlayerCnt; i++ {
		if g.players[i].GetWins() >= g.config.TargetWins {
			g.gameEndFlag = true
			g.matchWinnerIdx = i
			g.phase = ConquianPhaseGameEnd
			g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, i)), nil)
			return
		}
	}
}

// advanceTurn 次のプレイヤーへ
func (g *Conquian) advanceTurn() {
	g.currentPlayerIdx = 1 - g.currentPlayerIdx
	g.tookDiscard = false
	g.pendingCard = nil
	g.phase = ConquianPhaseDraw
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Conquian) GetPhase() ConquianPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Conquian) SetPhase(phase ConquianPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Conquian) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Conquian) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Conquian) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Conquian) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Conquian) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Conquian) SetDiscardPile(pile []*Card) {
	if pile == nil {
		pile = make([]*Card, 0)
	}
	g.discardPile = pile
}

// GetDiscardTop 捨て札の一番上を取得
func (g *Conquian) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Conquian) GetDrawPileCount() int { return len(g.drawPile) }

// SetStock 山札を設定 (テスト用)
func (g *Conquian) SetStock(pile []*Card) {
	if pile == nil {
		pile = make([]*Card, 0)
	}
	g.drawPile = pile
}

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Conquian) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx マッチ勝者インデックス取得 (-1 = 未確定/引き分け)
func (g *Conquian) GetWinnerIdx() int { return g.matchWinnerIdx }

// GetRoundWinnerIdx 直近ラウンド勝者インデックス取得 (-1 = 未確定/引き分け)
func (g *Conquian) GetRoundWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Conquian) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Conquian) GetPlayer(i int) *ConquianPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Conquian) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Conquian) GetConfig() ConquianConfig { return g.config }

// SetConfig 設定変更
func (g *Conquian) SetConfig(cfg ConquianConfig) { g.config = cfg }

// GetTookDiscard 今ターンに捨て札を取ったか (強制使用待ち) を返す
func (g *Conquian) GetTookDiscard() bool { return g.tookDiscard }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Conquian) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→ラン位置の順にソートする
func (g *Conquian) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return conquianRankPosition(cards[i].GetValue()) < conquianRankPosition(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- JSON ---

// conquianJSON is the JSON wire format for Conquian.
type conquianJSON struct {
	Players          []*ConquianPlayer `json:"pl"`
	Config           ConquianConfig    `json:"cf"`
	Phase            ConquianPhase     `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	MatchWinnerIdx   int               `json:"mw"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
	TookDiscard      bool              `json:"td"`
}

// MarshalJSON implements json.Marshaler.
func (g *Conquian) MarshalJSON() ([]byte, error) {
	return json.Marshal(conquianJSON{
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		MatchWinnerIdx:   g.matchWinnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		TookDiscard:      g.tookDiscard,
	})
}

// conquianMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const conquianMaxSliceLen = 1000

// errConquianInvalidState は不正なデシリアライズ入力を表す共有センチネルエラー。
var errConquianInvalidState = fmt.Errorf("conquian: invalid serialized state")

// UnmarshalJSON implements json.Unmarshaler.
//
// 入力を厳格に検証する: プレイヤー数は ConquianPlayerCnt と一致 (nil要素不可)、
// フェーズ・インデックスは範囲内、設定は Validate() を通過、スライス長は上限以内。
func (g *Conquian) UnmarshalJSON(data []byte) error {
	var j conquianJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	if len(j.Players) != ConquianPlayerCnt {
		return errConquianInvalidState
	}
	for _, p := range j.Players {
		if p == nil {
			return errConquianInvalidState
		}
	}
	if len(j.DiscardPile) > conquianMaxSliceLen || len(j.DrawPile) > conquianMaxSliceLen ||
		len(j.ActionLog) > conquianMaxSliceLen {
		return errConquianInvalidState
	}
	if j.Phase < ConquianPhaseDraw || j.Phase > ConquianPhaseGameEnd {
		return errConquianInvalidState
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= ConquianPlayerCnt {
		return errConquianInvalidState
	}
	if err := j.Config.Validate(); err != nil {
		return errConquianInvalidState
	}

	g.players = j.Players
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
	g.matchWinnerIdx = j.MatchWinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.tookDiscard = j.TookDiscard
	return nil
}
