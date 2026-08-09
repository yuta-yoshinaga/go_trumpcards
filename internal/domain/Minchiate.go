//go:build !js || !wasm || solo

// Package domain ミンキアーテ (Minchiate) のドメインモデル。
//
// Minchiate は 16 世紀フィレンツェの 97 枚タロー系トリックテイキングゲーム。
// 4 人・2 対 2 の固定チーム戦 (対面同士がパートナー) で、入札は無い。
//
// # デッキ (97 枚)
//
//   - スート札 56 枚: design = 1..4 (4 スート)、value = 1..14
//     (1-10, Fante=11, Cavallo=12, Regina=13, Re=14)。
//   - 切り札 (tarocchi) 40 枚: design = MinchiateTrumpDesign (5)、value = 1..40。
//   - マット (Matto / 愚者) 1 枚: design = MinchiateMattoDesign (6)、value = 0。
//
// **新しいデッキ型は要らない。**Scarto / FrenchTarot / Tarocchini と同じ仮想
// デザイン (5 = 切札, 6 = 愚者) に乗るので、ADR-0033 の手続き描画パスがそのまま
// 使える。違いは切札が 21 枚ではなく 40 枚まで伸びることだけ。
//
// # 配札
//
// **97 は 4 で割り切れない。**各自 21 枚を配ると 84 枚で 13 枚余り、その 13 枚を
// ディーラーが拾って同数を捨てる (Scarto / Tarocchini と同じスカルト構造)。
// 21×4 + 13 = 97 と過不足なく一致する。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// MinchiatePlayerCnt プレイヤー数 (人間 1 + CPU 3)。
const MinchiatePlayerCnt = 4

// MinchiateSuitCnt スート数。
const MinchiateSuitCnt = 4

// MinchiateSuitMax 各スートの最大値 (Re)。
const MinchiateSuitMax = 14

// MinchiateTrumpDesign 切り札の仮想デザイン値。
const MinchiateTrumpDesign = 5

// MinchiateMattoDesign マット (Matto / 愚者) の仮想デザイン値。
const MinchiateMattoDesign = 6

// MinchiateMattoValue マットのカード値。
const MinchiateMattoValue = 0

// MinchiateMaxTrump 切り札の最大値。**21 ではなく 40。**
const MinchiateMaxTrump = 40

// MinchiateDeckSize デッキ枚数 (56 + 40 + 1)。
const MinchiateDeckSize = MinchiateSuitCnt*MinchiateSuitMax + MinchiateMaxTrump + 1

// MinchiateHandSize 各プレイヤーがトリックプレイで持つ札の枚数。
const MinchiateHandSize = 21

// MinchiateSurplus ディーラーが拾って捨てる余剰札の枚数。
//
// **97 は 4 で割り切れない。**21 枚ずつ配ると 13 枚余るので、ディーラーがそれを
// 拾って同数を捨てる。数値を直接書かずデッキ構成から導くことで、枚数を変えたとき
// に配札とスカルトが食い違わないようにしている。
const MinchiateSurplus = MinchiateDeckSize - MinchiatePlayerCnt*MinchiateHandSize

// MinchiateTrickCount 1 ラウンドのトリック数。
const MinchiateTrickCount = MinchiateHandSize

// minchiateTeamCnt チーム数。
const minchiateTeamCnt = 2

// MinchiateTeamOf 席のチーム番号を返す。対面同士 (0-2 / 1-3) が組む。
func MinchiateTeamOf(seat int) int { return seat % 2 }

// minchiateTrumpNames は切札 40 枚の呼び名 (index = 切札の値)。
//
// **この 40 枚が Minchiate の顔。**上位は星座 12・四大元素 4・美徳 3 を含み、
// 単なる番号ではない。画面に番号だけを出すと「35 と 36 のどちらが強いか」以外の
// 情報が失われるので、名前を持たせて表示する。
var minchiateTrumpNames = [MinchiateMaxTrump + 1]string{
	"",
	// 1..5: 低位の切札 (papi と呼ばれる群を含む)
	"Papa Uno", "Papa Due", "Papa Tre", "Papa Quattro", "Papa Cinque",
	// 6..10
	"Amore", "Temperanza", "Fortezza", "Giustizia", "Ruota",
	// 11..15
	"Carro", "Eremita", "Impiccato", "Morte", "Diavolo",
	// 16..19: 四大元素
	"Casa", "Acqua", "Terra", "Aria",
	// 20: 火 (四大元素の 4 枚目)
	"Fuoco",
	// 21..32: 黄道十二宮
	"Bilancia", "Vergine", "Scorpione", "Ariete", "Capricorno", "Sagittario",
	"Cancro", "Pesci", "Acquario", "Leone", "Toro", "Gemelli",
	// 33..40: 最上位
	"Stella", "Luna", "Sole", "Mondo", "Fama", "Tromba", "Angelo", "Papa Quaranta",
}

// MinchiateTrumpName は切札の呼び名を返す。範囲外は空文字。
func MinchiateTrumpName(value int) string {
	if value < 1 || value > MinchiateMaxTrump {
		return ""
	}
	return minchiateTrumpNames[value]
}

// minchiateIsTrump 切り札か。
func minchiateIsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == MinchiateTrumpDesign
}

// minchiateIsMatto マットか。
func minchiateIsMatto(c *Card) bool {
	return c != nil && c.GetDesign() == MinchiateMattoDesign
}

// buildMinchiateDeck 97 枚のデッキを組む。
func buildMinchiateDeck() []*Card {
	deck := make([]*Card, 0, MinchiateDeckSize)
	for suit := 1; suit <= MinchiateSuitCnt; suit++ {
		for val := 1; val <= MinchiateSuitMax; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= MinchiateMaxTrump; val++ {
		deck = append(deck, NewCard(MinchiateTrumpDesign, val, false))
	}
	deck = append(deck, NewCard(MinchiateMattoDesign, MinchiateMattoValue, false))
	return deck
}

// MinchiatePhase ゲームフェーズ。入札は無い。
type MinchiatePhase int

const (
	// MinchiatePhaseScarto ディーラーが余剰札を捨てるフェーズ。
	MinchiatePhaseScarto MinchiatePhase = iota
	// MinchiatePhasePlay トリックプレイ。
	MinchiatePhasePlay
	// MinchiatePhaseTrickEnd トリック終了 (結果表示待ち)。
	MinchiatePhaseTrickEnd
	// MinchiatePhaseRoundEnd ラウンド終了 (精算済み)。
	MinchiatePhaseRoundEnd
	// MinchiatePhaseGameEnd マッチ終了。
	MinchiatePhaseGameEnd
)

// minchiateWinRank トリック比較用のランクを返す。高いほど強い。
//
// マットはトリックを取らないので常に最弱。切り札はリードスートより常に強い。
func minchiateWinRank(c *Card, led int) int {
	switch {
	case c == nil || minchiateIsMatto(c):
		return -1
	case minchiateIsTrump(c):
		return 1000 + c.GetValue()
	case c.GetDesign() == led:
		return c.GetValue()
	default:
		return -1
	}
}

// minchiateTrickWinnerOf は与えられたトリックの勝者席を返す。
//
// **Tarocchini と違い、同格の札は無い。**切札 40 枚はすべて序列が異なるので、
// 判定は「厳密に強い札のみ勝者を更新する」通常形でよい。
func minchiateTrickWinnerOf(trick []*TrickCard, led int) int {
	if len(trick) == 0 {
		return 0
	}
	winIdx := trick[0].PlayerIdx
	winRank := -1
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		if r := minchiateWinRank(tc.Card, led); r > winRank {
			winRank, winIdx = r, tc.PlayerIdx
		}
	}
	return winIdx
}

// Minchiate ミンキアーテのゲーム本体。
type Minchiate struct {
	players          []*MinchiatePlayer
	config           MinchiateConfig
	rng              *rand.Rand
	deck             []*Card
	deckDrawCnt      int
	phase            MinchiatePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	scarto           []*Card // ディーラーが捨てた札 (そのチームの獲得札に計上)
	lastTrickWinner  int
	teamScores       [minchiateTeamCnt]int
	roundTricks      [MinchiatePlayerCnt]int
	gameEndFlag      bool
	winnerTeam       int // -1 = 未確定 (同点)
	actionLog        []*ActionLogEntry
}

// NewMinchiate コンストラクタ。
func NewMinchiate(players []*MinchiatePlayer, config MinchiateConfig) *Minchiate {
	return &Minchiate{
		players: players,
		config:  config,
		// **本番の乱数源はここで入れる。**入れ忘れると山が並べ替わらず毎回同じ配りに
		// なる (Ganjifa #4661 で実際に起きた)。種は rand.Int63() から取る ——
		// time.Now().UnixNano() だと同じナノ秒に作った 2 局が同じ配りになりうる。
		rng:        rand.New(rand.NewSource(rand.Int63())),
		winnerTeam: -1,
	}
}

// NewDefaultMinchiate 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultMinchiate() *Minchiate {
	players := make([]*MinchiatePlayer, MinchiatePlayerCnt)
	players[0] = NewMinchiatePlayer(true)
	for i := 1; i < MinchiatePlayerCnt; i++ {
		players[i] = NewMinchiatePlayer(false)
	}
	return NewMinchiate(players, DefaultMinchiateConfig())
}

// SetRand 乱数源を差し替える (テスト用)。
func (g *Minchiate) SetRand(r *rand.Rand) { g.rng = r }

// Reset ゲームを初期化する。
func (g *Minchiate) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [minchiateTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *Minchiate) NextRound() {
	if g.phase != MinchiatePhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % MinchiatePlayerCnt
	g.startRound()
}

// startRound 手札を配り、スカルトフェーズを開始する。
func (g *Minchiate) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % MinchiatePlayerCnt
	g.lastTrickWinner = -1
	g.scarto = nil
	g.roundTricks = [MinchiatePlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.currentPlayerIdx = g.dealerIdx
	g.phase = MinchiatePhaseScarto
}

// deal 各プレイヤーへ 21 枚を配り、余剰 13 枚をディーラーへ渡す。
func (g *Minchiate) deal() {
	g.deck = buildMinchiateDeck()
	g.shuffle()
	g.deckDrawCnt = 0
	for i := 0; i < MinchiateHandSize; i++ {
		for j := 0; j < MinchiatePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % MinchiatePlayerCnt
			if c := g.drawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	for g.deckDrawCnt < len(g.deck) {
		if c := g.drawCard(); c != nil {
			g.players[g.dealerIdx].AddCard(c)
		}
	}
}

// shuffle 山を並べ替える。
func (g *Minchiate) shuffle() {
	// **nil は「並べ替えない」ではなく異常。**コンストラクタが必ず種を入れるので、
	// ここに来るのはゼロ値の Minchiate を直接組んだ場合だけ。
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	g.rng.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
}

// drawCard デッキから 1 枚配る (尽きたら nil)。
func (g *Minchiate) drawCard() *Card {
	return drawFromDeck(g.deck, &g.deckDrawCnt)
}

// sortAllHands 全員の手札をデザイン・値の順に整列する。
func (g *Minchiate) sortAllHands() {
	for _, p := range g.players {
		cards := make([]*Card, 0, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			cards = append(cards, p.GetCard(i))
		}
		sort.SliceStable(cards, func(i, j int) bool {
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
}

// finishMatch マッチを終え、得点の多いチームを勝者にする。
func (g *Minchiate) finishMatch() {
	g.gameEndFlag = true
	g.phase = MinchiatePhaseGameEnd
	switch {
	case g.teamScores[0] > g.teamScores[1]:
		g.winnerTeam = 0
	case g.teamScores[1] > g.teamScores[0]:
		g.winnerTeam = 1
	default:
		// **同点なら勝者なし。**席順で決めると片方のチームが常に得をする。
		g.winnerTeam = -1
	}
	g.appendLog(-1, "gameend", "マッチ終了", nil)
}

// appendLog 棋譜に 1 件追加する。
func (g *Minchiate) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.trickNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- スカルト (ディーラーの捨て札) ---

// minchiateCanDiscard スート札 (第一候補) か。
//
// **切札とマットは原則捨てられない。**捨札はそのまま獲得札に計上されるので、
// 高位の切札を伏せて確実に取れてしまうとトリックを戦う意味が薄れる。
func minchiateCanDiscard(c *Card) bool {
	return c != nil && !minchiateIsTrump(c) && !minchiateIsMatto(c)
}

// minchiateDiscardable は親が捨ててよい手札位置を弱い順に返す。
//
// **スート札だけでは足りない配りが実在する。**親は 97 枚中 34 枚を持ち、その
// うち捨てられないのは切札 40 + マット 1 の 41 枚。34 枚中スート札が 13 枚に
// 満たない裾は約 0.3% で起こる —— 「構造上ありえない」は誤りだった (#4665 の
// レビュー指摘)。そこで足りない分だけ切札を弱い順に開放する。Scarto が
// 「0.5 点札が足りない稀な場合に限り非ブーの切札を許す」のと同じ形。
//
// 捨てずに進める逃げ方は取らない。親の手札が 34 枚のまま 21 トリックを戦うと
// 13 枚が一度も場に出ず、どちらのチームにも計上されないまま次の配りで消える。
// 97 = 21x4 + 13 という閉じた勘定が崩れる。
func minchiateDiscardable(dealer *MinchiatePlayer) []int {
	var suits, trumps []int
	for i := 0; i < dealer.GetCardsSize(); i++ {
		c := dealer.GetCard(i)
		switch {
		case minchiateCanDiscard(c):
			suits = append(suits, i)
		case minchiateIsTrump(c):
			trumps = append(trumps, i)
		}
	}
	byValue := func(xs []int) {
		sort.SliceStable(xs, func(a, b int) bool {
			return dealer.GetCard(xs[a]).GetValue() < dealer.GetCard(xs[b]).GetValue()
		})
	}
	byValue(suits)
	if len(suits) >= MinchiateSurplus {
		return suits
	}
	// **マットは最後まで開放しない。**34 枚のうちマットは高々 1 枚なので、
	// スート札 + 切札で必ず MinchiateSurplus 枚に届く。
	byValue(trumps)
	return append(suits, trumps...)
}

// PlayerScarto 人間のディーラーが余剰札を伏せて捨てる。
func (g *Minchiate) PlayerScarto(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MinchiatePhaseScarto {
		return ErrWrongPhase
	}
	if !g.players[g.dealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if len(cardIndices) != MinchiateSurplus {
		return NewDomainError(ErrInvalidIndices,
			fmt.Sprintf("捨てる札は %d 枚選んでください", MinchiateSurplus))
	}
	dealer := g.players[g.dealerIdx]
	// **許可集合は毎回計算する。**スート札が足りない配りでは切札も開放される。
	allowed := minchiateDiscardable(dealer)
	seen := make(map[int]bool, len(cardIndices))
	for _, idx := range cardIndices {
		if idx < 0 || idx >= dealer.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidIndices, "同じ札を 2 回選べません")
		}
		seen[idx] = true
		if !minchiateContainsIdx(allowed, idx) {
			return NewDomainError(ErrInvalidPlay, "切札とマットは捨てられません")
		}
	}
	g.applyScarto(cardIndices)
	return nil
}

// applyScarto 選ばれた札を手札から抜き、獲得札として保持する。
func (g *Minchiate) applyScarto(cardIndices []int) {
	dealer := g.players[g.dealerIdx]
	// **降順に外す。**昇順で RemoveCard すると 2 枚目以降の位置がずれる。
	idx := append([]int(nil), cardIndices...)
	sort.Sort(sort.Reverse(sort.IntSlice(idx)))
	discarded := make([]*Card, 0, len(idx))
	for _, i := range idx {
		if c := dealer.RemoveCard(i); c != nil {
			discarded = append(discarded, c)
		}
	}
	g.scarto = discarded
	g.appendLog(g.dealerIdx, "scarto", fmt.Sprintf("%d 枚を捨てた", len(discarded)), discarded)
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = MinchiatePhasePlay
}

// CpuScarto CPU のディーラーが自動で捨てる。
func (g *Minchiate) CpuScarto() {
	if g.gameEndFlag || g.phase != MinchiatePhaseScarto {
		return
	}
	dealer := g.players[g.dealerIdx]
	if dealer.GetIsHuman() {
		return
	}
	order := minchiateDiscardable(dealer)
	if len(order) < MinchiateSurplus {
		// ここに来るのは手札が MinchiateSurplus 枚に満たないときだけ (テストが
		// 手札を差し替えた場合など)。持っている分だけ捨てて先へ進める。
		g.applyScarto(order)
		return
	}
	g.applyScarto(order[:MinchiateSurplus])
}

// --- トリックプレイ ---

// IsHumanTurn 人間のプレイ手番か。
func (g *Minchiate) IsHumanTurn() bool {
	return g.phase == MinchiatePhasePlay && g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanScartoTurn 人間のスカルト手番か。
func (g *Minchiate) IsHumanScartoTurn() bool {
	return g.phase == MinchiatePhaseScarto && g.players[g.dealerIdx].GetIsHuman()
}

// ledSuit 現在のトリックのリードデザインを返す (未リードなら 0)。
//
// **マットはスートを定めない。**マットはトリックを取らずフォロー義務も免れる札
// なので、これがリードされた場合のリードスートは「次に出た札」が決める。
// 先頭札のデザインをそのまま返すと 2 つ壊れる: 他の全札がリードスート外になって
// ランク -1 に落ちマットの席がトリックを取ってしまうこと、そして続く players に
// 切札を出す義務が生じて手札が丸ごと縛られること (Tarocchini #4662 で検出)。
func (g *Minchiate) ledSuit() int {
	for _, tc := range g.currentTrick {
		if tc == nil || tc.Card == nil || minchiateIsMatto(tc.Card) {
			continue
		}
		return tc.Card.GetDesign()
	}
	return 0
}

// GetValidPlayIndices 出せる手札の位置を返す。
//
// リードスートを持っていれば従う義務。持っていなければ切札を出す義務 (タロー系の
// 共通ルール)。マットはいつでも出せる。
func (g *Minchiate) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	p := g.players[playerIdx]
	all := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		all = append(all, i)
	}
	led := g.ledSuit()
	// **リードスートがまだ決まっていなければ義務は生じない。**トリックが空のとき
	// だけでなく、マットがリードされてまだ誰もスートを定めていないときも同じ。
	if len(g.currentTrick) == 0 || led == 0 {
		return all
	}
	var follow, trumps, matto []int
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		switch {
		case minchiateIsMatto(c):
			matto = append(matto, i)
		case c.GetDesign() == led:
			follow = append(follow, i)
		case minchiateIsTrump(c):
			trumps = append(trumps, i)
		}
	}
	if len(follow) > 0 {
		return append(follow, matto...)
	}
	if len(trumps) > 0 {
		return append(trumps, matto...)
	}
	return all
}

// GetPlayableIndices プレイ可能なカードの位置。
func (g *Minchiate) GetPlayableIndices(playerIdx int) []int {
	return g.GetValidPlayIndices(playerIdx)
}

// PlayerPlay 人間がカードを出す。
func (g *Minchiate) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MinchiatePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if !minchiateContains(g.GetValidPlayIndices(g.currentPlayerIdx), cardIndex) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	g.playCard(g.currentPlayerIdx, player.RemoveCard(cardIndex))
	return nil
}

// minchiateContainsIdx 許可集合に位置が含まれるか。
func minchiateContainsIdx(xs []int, v int) bool { return minchiateContains(xs, v) }

// minchiateContains slice に値が含まれるか。
func minchiateContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// CpuPlay 現在の手番が CPU の場合に 1 枚出す。
func (g *Minchiate) CpuPlay() {
	if g.gameEndFlag || g.phase != MinchiatePhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	valid := g.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return
	}
	played := g.players[idx].RemoveCard(g.cpuSelectPlayCard(idx, valid))
	// **出せる札が無ければ何もしない。**RemoveCard は手札が空なら nil を返し、
	// それを playCard に渡すと nil デリファレンスでハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// cpuSelectPlayCard CPU が出す札の位置を選ぶ。
func (g *Minchiate) cpuSelectPlayCard(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	if g.config.CpuDifficulty == MinchiateCpuDifficultyEasy {
		return valid[g.rng.Intn(len(valid))]
	}
	led := g.ledSuit()
	best, bestRank := valid[0], -1<<31
	for _, i := range valid {
		if r := minchiateWinRank(p.GetCard(i), led); r > bestRank {
			best, bestRank = i, r
		}
	}
	// **マットだけがリードされた局面はまだ「誰も勝っていない」。**follow 扱いに
	// すると、味方がマットを出した局面で「味方が勝っている」と誤判定する。
	if len(g.currentTrick) == 0 || led == 0 {
		return best
	}
	// 味方が既に勝っているトリックを奪わない。
	winIdx := minchiateTrickWinnerOf(g.currentTrick, led)
	if MinchiateTeamOf(winIdx) == MinchiateTeamOf(playerIdx) {
		return g.lowestOf(playerIdx, valid, led)
	}
	trial := append(append([]*TrickCard(nil), g.currentTrick...),
		&TrickCard{PlayerIdx: playerIdx, Card: p.GetCard(best)})
	if minchiateTrickWinnerOf(trial, led) == playerIdx {
		return best
	}
	return g.lowestOf(playerIdx, valid, led)
}

// lowestOf 候補のうち最も弱い札の位置を返す。
func (g *Minchiate) lowestOf(playerIdx int, valid []int, led int) int {
	p := g.players[playerIdx]
	low, lowRank := valid[0], 1<<31-1
	for _, i := range valid {
		if r := minchiateWinRank(p.GetCard(i), led); r < lowRank {
			low, lowRank = i, r
		}
	}
	return low
}

// playCard 1 枚を場に出し、トリックが揃えばフェーズを進める。
func (g *Minchiate) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", "", []*Card{card})
	if len(g.currentTrick) < MinchiatePlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % MinchiatePlayerCnt
		return
	}
	g.phase = MinchiatePhaseTrickEnd
}

// ResolveTrick トリックの勝者を決める。
func (g *Minchiate) ResolveTrick() {
	if g.phase != MinchiatePhaseTrickEnd {
		return
	}
	winnerIdx := minchiateTrickWinnerOf(g.currentTrick, g.ledSuit())
	cards := make([]*Card, 0, len(g.currentTrick))
	for _, tc := range g.currentTrick {
		cards = append(cards, tc.Card)
	}
	g.players[winnerIdx].AddTrick(cards)
	g.roundTricks[winnerIdx]++
	g.lastTrickWinner = winnerIdx
	g.appendLog(winnerIdx, "trickwin", fmt.Sprintf("トリック %d を獲得", g.trickNumber), cards)

	g.leadPlayerIdx = winnerIdx
	g.currentPlayerIdx = winnerIdx
	if g.trickNumber >= MinchiateTrickCount {
		g.settleRound()
	}
}

// NextTrick 次のトリックを開始する。
func (g *Minchiate) NextTrick() {
	if g.phase != MinchiatePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = MinchiatePhasePlay
}

// MinchiateLastTrickBonus 最終トリック獲得ボーナス。
const MinchiateLastTrickBonus = 3

// settleRound ラウンドを精算する。
//
// **最終トリックにボーナスを与える。**ミンキアーテは最後のトリックを取った側に
// 加点する慣習があり、終盤に高位の切札を温存する動機になる。
func (g *Minchiate) settleRound() {
	g.phase = MinchiatePhaseRoundEnd
	for seat, n := range g.roundTricks {
		g.teamScores[MinchiateTeamOf(seat)] += n
	}
	if g.lastTrickWinner >= 0 {
		g.teamScores[MinchiateTeamOf(g.lastTrickWinner)] += MinchiateLastTrickBonus
	}
	if len(g.scarto) > 0 {
		g.teamScores[MinchiateTeamOf(g.dealerIdx)] += len(g.scarto)
	}
	g.appendLog(-1, "settle", fmt.Sprintf("ラウンド %d 終了", g.roundNumber), nil)
}

// ScoreRound ラウンドを締め、規定局数ならマッチを終える。
func (g *Minchiate) ScoreRound() {
	if g.phase != MinchiatePhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
	}
}

// --- アクセサ ---

// GetPhase 現在のフェーズ。
func (g *Minchiate) GetPhase() MinchiatePhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *Minchiate) SetPhase(p MinchiatePhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号。
func (g *Minchiate) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号。
func (g *Minchiate) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 現在の手番。
func (g *Minchiate) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番を設定する (テスト用)。
func (g *Minchiate) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック。
func (g *Minchiate) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLeadPlayerIdx リード席。
func (g *Minchiate) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetDealerIdx ディーラー席。
func (g *Minchiate) GetDealerIdx() int { return g.dealerIdx }

// GetLastTrickWinner 直前のトリックの勝者席 (-1 = 未確定)。
func (g *Minchiate) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetScartoSize ディーラーが捨てた枚数。
func (g *Minchiate) GetScartoSize() int { return len(g.scarto) }

// GetTeamScores チーム別の累計得点。
func (g *Minchiate) GetTeamScores() [minchiateTeamCnt]int { return g.teamScores }

// GetRoundTricks 現ラウンドの席別獲得トリック数。
func (g *Minchiate) GetRoundTricks() [MinchiatePlayerCnt]int { return g.roundTricks }

// GetGameEndFlag ゲーム終了フラグ。
func (g *Minchiate) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1 = 未確定 / 同点)。
func (g *Minchiate) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数。
func (g *Minchiate) GetPlayerCnt() int { return MinchiatePlayerCnt }

// GetPlayer 指定席のプレイヤー。
func (g *Minchiate) GetPlayer(i int) *MinchiatePlayer {
	return getPlayer(g.players, i)
}

// GetPlayers 全プレイヤー。
func (g *Minchiate) GetPlayers() []*MinchiatePlayer { return g.players }

// GetConfig ゲーム設定。
func (g *Minchiate) GetConfig() MinchiateConfig { return g.config }

// SetConfig ゲーム設定を差し替える。
func (g *Minchiate) SetConfig(c MinchiateConfig) { g.config = c }

// GetActionLog 棋譜。
func (g *Minchiate) GetActionLog() []*ActionLogEntry { return g.actionLog }

// MinchiateHint ヒント情報。
type MinchiateHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// GetHint 人間プレイヤーへのヒント。手番でなければ nil。
func (g *Minchiate) GetHint() *MinchiateHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != MinchiatePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.GetValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuSelectPlayCard(human, valid)
	return &MinchiateHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason ヒント理由キーを判定する。
func (g *Minchiate) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	switch {
	case card == nil:
		return "lead_low"
	case minchiateIsMatto(card):
		return "play_matto"
	case len(g.currentTrick) == 0 && minchiateIsTrump(card):
		return "lead_trump"
	case len(g.currentTrick) == 0:
		return "lead_low"
	case minchiateIsTrump(card):
		return "follow_trump"
	default:
		return "follow_low"
	}
}

// --- KV 永続化 ---

// minchiateJSON is the JSON wire format for Minchiate.
//
// **全フィールドが非公開なので専用のコーデックが要る。**これを省くと Cloudflare
// Worker のセッション復元が空のゲームを返し、リクエストのたびに手札も得点も消える
// (Ganjifa #4661 / Vira #4660 で実際に起きた)。
type minchiateJSON struct {
	Players          []*MinchiatePlayer      `json:"pl"`
	Config           MinchiateConfig         `json:"cfg"`
	Phase            MinchiatePhase          `json:"ph"`
	RoundNumber      int                     `json:"rn"`
	TrickNumber      int                     `json:"tn"`
	CurrentPlayerIdx int                     `json:"cpi"`
	CurrentTrick     []*TrickCard            `json:"ct"`
	LeadPlayerIdx    int                     `json:"lpi"`
	DealerIdx        int                     `json:"di"`
	Scarto           []*Card                 `json:"sc"`
	LastTrickWinner  int                     `json:"ltw"`
	TeamScores       [minchiateTeamCnt]int   `json:"ts"`
	RoundTricks      [MinchiatePlayerCnt]int `json:"rt"`
	GameEndFlag      bool                    `json:"gef"`
	WinnerTeam       int                     `json:"wt"`
	ActionLog        []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Minchiate) MarshalJSON() ([]byte, error) {
	return json.Marshal(minchiateJSON{
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		Scarto:           g.scarto,
		LastTrickWinner:  g.lastTrickWinner,
		TeamScores:       g.teamScores,
		RoundTricks:      g.roundTricks,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// minchiateMaxSliceLen caps slice sizes during deserialisation.
const minchiateMaxSliceLen = 5000

var (
	errMinchiateOversized      = errors.New("minchiate: input array exceeds maximum allowed size")
	errMinchiateInvalidPlayers = errors.New("minchiate: invalid player count")
	errMinchiateInvalidTrick   = errors.New("minchiate: invalid trick card")
	errMinchiateInvalidState   = errors.New("minchiate: invalid state values in json")
)

// UnmarshalJSON implements json.Unmarshaler.
//
// **デザインは 1..6 を許す。**52 枚デッキ用の 1..4 を持ち込むと、切札 (5) と
// マット (6) を含むセッションが全て復元不能になる。
func (g *Minchiate) UnmarshalJSON(data []byte) error {
	var j minchiateJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > minchiateMaxSliceLen || len(j.CurrentTrick) > minchiateMaxSliceLen ||
		len(j.ActionLog) > minchiateMaxSliceLen || len(j.Scarto) > minchiateMaxSliceLen {
		return errMinchiateOversized
	}
	if len(j.Players) != MinchiatePlayerCnt {
		return errMinchiateInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errMinchiateInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= MinchiatePlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= MinchiatePlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= MinchiatePlayerCnt ||
		j.LastTrickWinner < -1 || j.LastTrickWinner >= MinchiatePlayerCnt ||
		j.WinnerTeam < -1 || j.WinnerTeam >= minchiateTeamCnt ||
		len(j.Scarto) > MinchiateSurplus ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > MinchiateTrickCount ||
		j.Phase < MinchiatePhaseScarto || j.Phase > MinchiatePhaseGameEnd {
		return errMinchiateInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= MinchiatePlayerCnt {
			return errMinchiateInvalidTrick
		}
		if d := tc.Card.GetDesign(); d < 1 || d > MinchiateMattoDesign {
			return errMinchiateInvalidTrick
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.scarto = j.Scarto
	g.lastTrickWinner = j.LastTrickWinner
	g.teamScores = j.TeamScores
	g.roundTricks = j.RoundTricks
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	// **復元したら必ず乱数源を張り直す。**Worker は毎リクエスト KV から組み直すので
	// SetRand は呼ばれない。rng を nil のままにすると cpuSelectPlayCard の Easy
	// 分岐が nil デリファレンスで落ちる (Tarocchini #4662 のレビューで検出)。
	g.rng = rand.New(rand.NewSource(rand.Int63()))
	return nil
}
