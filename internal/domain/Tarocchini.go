//go:build !js || !wasm || solo

// Package domain タロッキーニ (Tarocchini / Ottocento) のドメインモデル。
//
// Tarocchini はイタリア・ボローニャの 62 枚タロット (タロッコ・ボロニェーゼ) を用いる
// 4 人・2 対 2 の固定チーム制トリックテイキングゲーム。対面同士がパートナーになる。
//
// # デッキ (62 枚)
//
//   - スート札 40 枚: design = 1..4 (4 スート)、value ∈ {1, 6, 7, 8, 9, 10, 11, 12, 13, 14}。
//     **2..5 のピップは抜かれている**のがボロニェーゼ版の特徴で、各スート 10 枚になる。
//   - 切り札 (trumps) 21 枚: design = TarocchiniTrumpDesign (5)、value = 1..21。
//   - マット (Matto / Fool) 1 枚: design = TarocchiniMattoDesign (6)、value = 0。
//
// # このゲームを他と分けているもの
//
// **同格のパパ 4 枚は「後から出した方」が勝つ。**他のトリックテイカーの勝敗判定は
// 「より強い札が出たら勝者を更新する」(厳密な >) で書けるが、パパ同士は順位が等しい
// ため、同値でも更新する (>=) 必要がある。Scarto の `trickWinner` をそのまま流用すると
// 先に出したパパが勝ってしまうので、本ゲームは専用の判定を持つ。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// TarocchiniPlayerCnt プレイヤー数 (人間 1 + CPU 3)。
const TarocchiniPlayerCnt = 4

// TarocchiniSuitCnt スート数。
const TarocchiniSuitCnt = 4

// TarocchiniTrumpDesign 切り札の仮想デザイン値。1..4 はスート、5 が切り札。
//
// Scarto / FrenchTarot と同じ割り当てにしてある。**新しいデッキ型は要らない** ——
// ADR-0033 の手続き描画パスがこの仮想デザインをそのまま描ける。
const TarocchiniTrumpDesign = 5

// TarocchiniMattoDesign マット (Matto / Fool) の仮想デザイン値。
const TarocchiniMattoDesign = 6

// TarocchiniMattoValue マットのカード値。
const TarocchiniMattoValue = 0

// TarocchiniMaxTrump 切り札の最大値。
const TarocchiniMaxTrump = 21

// TarocchiniKingValue 王 (Re) の値。
const TarocchiniKingValue = 14

// tarocchiniSuitValues 各スートに残る 10 枚の値。
//
// **2..5 は入らない。**ボロニェーゼ版は低位ピップを抜いた 40 枚構成で、これを
// 52 枚デッキの感覚で 1..10 に読み替えると枚数が合わなくなる。
var tarocchiniSuitValues = [...]int{1, 6, 7, 8, 9, 10, 11, 12, 13, 14}

// TarocchiniDeckSize デッキ枚数 (40 + 21 + 1)。
const TarocchiniDeckSize = TarocchiniSuitCnt*len(tarocchiniSuitValues) + TarocchiniMaxTrump + 1

// TarocchiniHandSize 各プレイヤーがトリックプレイで持つ札の枚数。
const TarocchiniHandSize = 15

// TarocchiniSurplus ディーラーが拾って捨てる余剰札の枚数。
//
// **62 は 4 で割り切れない。**15 枚ずつ配ると 60 枚で 2 枚余るので、その 2 枚は
// ディーラーが拾い、同数を捨てて手札を 15 枚に戻す (Scarto のスカルトと同型)。
const TarocchiniSurplus = TarocchiniDeckSize - TarocchiniPlayerCnt*TarocchiniHandSize

// TarocchiniTrickCount 1 ラウンドのトリック数。
const TarocchiniTrickCount = TarocchiniHandSize

// tarocchiniPapiLow / tarocchiniPapiHigh は同格に扱う切り札 (パパ / モーリ) の範囲。
//
// **この 4 枚の同定は未確定。**タロッコ・ボロニェーゼの切札は「ベガート(1) → 同格の
// パパ 4 枚 → 番号付きの中位 → 最上位 4 枚 (アンジェロ/モンド/ソーレ/ルーナ)」という
// 構成で、"同格 4 枚" と "最上位 4 枚" は別のグループ。issue #4409 本文は両者を
// 混同している疑いがあるため、ここではベガートの直上 4 枚 (切札 2..5) を採る。
// 同定が変われば**この 2 つの定数だけ**を差し替えればよい。
const (
	tarocchiniPapiLow  = 2
	tarocchiniPapiHigh = 5
)

// TarocchiniIsPapa 同格に扱うパパか。
func TarocchiniIsPapa(c *Card) bool {
	return tarocchiniIsTrump(c) &&
		c.GetValue() >= tarocchiniPapiLow && c.GetValue() <= tarocchiniPapiHigh
}

// tarocchiniIsTrump 切り札か。
func tarocchiniIsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == TarocchiniTrumpDesign
}

// tarocchiniIsMatto マットか。
func tarocchiniIsMatto(c *Card) bool {
	return c != nil && c.GetDesign() == TarocchiniMattoDesign
}

// buildTarocchiniDeck 62 枚のデッキを組む。
func buildTarocchiniDeck() []*Card {
	deck := make([]*Card, 0, TarocchiniDeckSize)
	for suit := 1; suit <= TarocchiniSuitCnt; suit++ {
		for _, val := range tarocchiniSuitValues {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= TarocchiniMaxTrump; val++ {
		deck = append(deck, NewCard(TarocchiniTrumpDesign, val, false))
	}
	deck = append(deck, NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false))
	return deck
}

// TarocchiniTeamOf 席のチーム番号を返す。対面同士 (0-2 / 1-3) が組む。
func TarocchiniTeamOf(seat int) int { return seat % 2 }

// TarocchiniPhase ゲームフェーズ。入札は無い。
type TarocchiniPhase int

const (
	// TarocchiniPhaseScarto ディーラーが余剰 2 枚を捨てるフェーズ。
	TarocchiniPhaseScarto TarocchiniPhase = iota
	// TarocchiniPhasePlay トリックプレイ。
	TarocchiniPhasePlay
	// TarocchiniPhaseTrickEnd トリック終了 (結果表示待ち)。
	TarocchiniPhaseTrickEnd
	// TarocchiniPhaseRoundEnd ラウンド終了 (精算済み)。
	TarocchiniPhaseRoundEnd
	// TarocchiniPhaseGameEnd マッチ終了。
	TarocchiniPhaseGameEnd
)

// tarocchiniWinRank トリック比較用のランクを返す。高いほど強い。
//
// マットはトリックを取らないので常に最弱。切り札はリードスートより常に強い。
// **パパ同士の同格判定はここではしない** —— ランクだけでは「後出し優先」を
// 表現できないため、trickWinner 側で扱う。
func tarocchiniWinRank(c *Card, led int) int {
	switch {
	case c == nil || tarocchiniIsMatto(c):
		return -1
	case TarocchiniIsPapa(c):
		// **4 枚のパパは 1 つのランクに畳む。**素の値のままだと 2..5 が別ランクに
		// なり、同格であることも後出し優先も表現できない。
		return 1000 + tarocchiniPapiLow
	case tarocchiniIsTrump(c):
		return 1000 + c.GetValue()
	case c.GetDesign() == led:
		return c.GetValue()
	default:
		return -1
	}
}

// tarocchiniTrickWinnerOf は与えられたトリックの勝者席を返す。
//
// **同格のパパは後から出した方が勝つ。**そのため比較は「厳密に強い」ではなく、
// 「強い、または (両方パパで) 同じランク」で勝者を更新する。Scarto から
// `r > winRank` をそのまま持ってくると先に出したパパが勝ってしまう。
func tarocchiniTrickWinnerOf(trick []*TrickCard, led int) int {
	if len(trick) == 0 {
		return 0
	}
	winIdx := trick[0].PlayerIdx
	winRank := -1
	var winCard *Card
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		r := tarocchiniWinRank(tc.Card, led)
		tie := r == winRank && TarocchiniIsPapa(tc.Card) && TarocchiniIsPapa(winCard)
		if r > winRank || tie {
			winRank, winIdx, winCard = r, tc.PlayerIdx, tc.Card
		}
	}
	return winIdx
}

// Tarocchini タロッキーニのゲーム本体。
type Tarocchini struct {
	players          []*TarocchiniPlayer
	config           TarocchiniConfig
	rng              *rand.Rand
	deck             []*Card
	deckDrawCnt      int
	phase            TarocchiniPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	scarto           []*Card // ディーラーが捨てた 2 枚 (そのチームの獲得札に計上)
	lastTrickWinner  int
	teamScores       [tarocchiniTeamCnt]int
	roundTricks      [TarocchiniPlayerCnt]int
	gameEndFlag      bool
	winnerTeam       int // -1 = 未確定 (同点)
	actionLog        []*ActionLogEntry
}

// tarocchiniTeamCnt チーム数。
const tarocchiniTeamCnt = 2

// NewTarocchini コンストラクタ。
func NewTarocchini(players []*TarocchiniPlayer, config TarocchiniConfig) *Tarocchini {
	return &Tarocchini{
		players: players,
		config:  config,
		// **本番の乱数源はここで入れる。**入れ忘れると山が並べ替わらず毎回同じ配りに
		// なる (Ganjifa #4661 で実際に起きた)。種は rand.Int63() から取る ——
		// time.Now().UnixNano() だと同じナノ秒に作った 2 局が同じ配りになりうる。
		rng:        rand.New(rand.NewSource(rand.Int63())),
		winnerTeam: -1,
	}
}

// NewDefaultTarocchini 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultTarocchini() *Tarocchini {
	players := make([]*TarocchiniPlayer, TarocchiniPlayerCnt)
	players[0] = NewTarocchiniPlayer(true)
	for i := 1; i < TarocchiniPlayerCnt; i++ {
		players[i] = NewTarocchiniPlayer(false)
	}
	return NewTarocchini(players, DefaultTarocchiniConfig())
}

// SetRand 乱数源を差し替える (テスト用)。
func (g *Tarocchini) SetRand(r *rand.Rand) { g.rng = r }

// Reset ゲームを初期化する。
func (g *Tarocchini) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [tarocchiniTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する。
func (g *Tarocchini) NextRound() {
	if g.phase != TarocchiniPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % TarocchiniPlayerCnt
	g.startRound()
}

// startRound 手札を配り、スカルトフェーズを開始する。
func (g *Tarocchini) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % TarocchiniPlayerCnt
	g.lastTrickWinner = -1
	g.scarto = nil
	g.roundTricks = [TarocchiniPlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.currentPlayerIdx = g.dealerIdx
	g.phase = TarocchiniPhaseScarto
}

// deal 各プレイヤーへ 15 枚を配り、余剰 2 枚をディーラーへ渡す。
//
// **ディーラーは一時的に 17 枚を持つ。**62 は 4 で割り切れないので、余りを誰かが
// 引き受けるしかない。ディーラーがそれを拾って同数を捨てる (スカルト)。
func (g *Tarocchini) deal() {
	g.deck = buildTarocchiniDeck()
	g.shuffle()
	g.deckDrawCnt = 0
	for i := 0; i < TarocchiniHandSize; i++ {
		for j := 0; j < TarocchiniPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % TarocchiniPlayerCnt
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
func (g *Tarocchini) shuffle() {
	// **nil は「並べ替えない」ではなく異常。**コンストラクタが必ず種を入れるので、
	// ここに来るのはゼロ値の Tarocchini を直接組んだ場合だけ。
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	g.rng.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
}

// drawCard デッキから 1 枚配る (尽きたら nil)。
func (g *Tarocchini) drawCard() *Card {
	if g.deckDrawCnt >= len(g.deck) {
		return nil
	}
	card := g.deck[g.deckDrawCnt]
	card.SetDraw(true)
	g.deckDrawCnt++
	return card
}

// sortAllHands 全員の手札をデザイン・値の順に整列する。
func (g *Tarocchini) sortAllHands() {
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

// finishMatch マッチを終え、獲得トリック数の累計が最大のチームを勝者にする。
func (g *Tarocchini) finishMatch() {
	g.gameEndFlag = true
	g.phase = TarocchiniPhaseGameEnd
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
func (g *Tarocchini) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.trickNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- スカルト (ディーラーの捨て札) ---

// tarocchiniCanDiscard 捨てられる札か。
//
// **切り札とマットは捨てられない。**捨札はそのまま獲得札に計上されるので、
// 高得点の切り札を伏せて確実に取れてしまうと、トリックを戦う意味が薄れる。
func tarocchiniCanDiscard(c *Card) bool {
	return c != nil && !tarocchiniIsTrump(c) && !tarocchiniIsMatto(c)
}

// PlayerScarto 人間のディーラーが 2 枚を伏せて捨てる。
func (g *Tarocchini) PlayerScarto(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TarocchiniPhaseScarto {
		return ErrWrongPhase
	}
	if !g.players[g.dealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if len(cardIndices) != TarocchiniSurplus {
		return NewDomainError(ErrInvalidIndices,
			fmt.Sprintf("捨てる札は %d 枚選んでください", TarocchiniSurplus))
	}
	dealer := g.players[g.dealerIdx]
	seen := make(map[int]bool, len(cardIndices))
	for _, idx := range cardIndices {
		if idx < 0 || idx >= dealer.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidIndices, "同じ札を 2 回選べません")
		}
		seen[idx] = true
		if !tarocchiniCanDiscard(dealer.GetCard(idx)) {
			return NewDomainError(ErrInvalidPlay, "切り札とマットは捨てられません")
		}
	}
	g.applyScarto(cardIndices)
	return nil
}

// applyScarto 選ばれた札を手札から抜き、獲得札として保持する。
func (g *Tarocchini) applyScarto(cardIndices []int) {
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
	g.phase = TarocchiniPhasePlay
}

// CpuScarto CPU のディーラーが自動で 2 枚捨てる。
func (g *Tarocchini) CpuScarto() {
	if g.gameEndFlag || g.phase != TarocchiniPhaseScarto {
		return
	}
	dealer := g.players[g.dealerIdx]
	if dealer.GetIsHuman() {
		return
	}
	picked := make([]int, 0, TarocchiniSurplus)
	// 捨てられる札のうち値の低いものから。
	order := make([]int, 0, dealer.GetCardsSize())
	for i := 0; i < dealer.GetCardsSize(); i++ {
		if tarocchiniCanDiscard(dealer.GetCard(i)) {
			order = append(order, i)
		}
	}
	sort.SliceStable(order, func(a, b int) bool {
		return dealer.GetCard(order[a]).GetValue() < dealer.GetCard(order[b]).GetValue()
	})
	for _, i := range order {
		if len(picked) == TarocchiniSurplus {
			break
		}
		picked = append(picked, i)
	}
	// **捨てられる札が足りないことは構造上ありえない**が、起きたら捨てずに進める
	// 方が、手札枚数を壊してトリック数が合わなくなるより安全。
	if len(picked) < TarocchiniSurplus {
		g.currentPlayerIdx = g.leadPlayerIdx
		g.phase = TarocchiniPhasePlay
		return
	}
	g.applyScarto(picked)
}

// --- トリックプレイ ---

// IsHumanTurn 人間のプレイ手番か。
func (g *Tarocchini) IsHumanTurn() bool {
	return g.phase == TarocchiniPhasePlay && g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanScartoTurn 人間のスカルト手番か。
func (g *Tarocchini) IsHumanScartoTurn() bool {
	return g.phase == TarocchiniPhaseScarto && g.players[g.dealerIdx].GetIsHuman()
}

// ledSuit 現在のトリックのリードデザインを返す (未リードなら 0)。
//
// **マットはスートを定めない。**マットはトリックを取らずフォロー義務も免れる札
// なので、これがリードされた場合のリードスートは「次に出た札」が決める。
// 単純に先頭札のデザインを返すと 2 つ壊れる: (1) 他の全札がリードスート外に
// なってランク -1 に落ち、マットを出した席がトリックを取ってしまう。
// (2) 続く players が「リードスートを持たない」と判定され、切札を出す義務が
// 生じて手札が丸ごと縛られる。
func (g *Tarocchini) ledSuit() int {
	for _, tc := range g.currentTrick {
		if tc == nil || tc.Card == nil || tarocchiniIsMatto(tc.Card) {
			continue
		}
		return tc.Card.GetDesign()
	}
	return 0
}

// GetValidPlayIndices 出せる手札の位置を返す。
//
// リードスートを持っていればそれに従う義務。持っていなければ切り札を出す義務
// (タロー系の共通ルール)。マットはいつでも出せる。
func (g *Tarocchini) GetValidPlayIndices(playerIdx int) []int {
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
	// ここを取りこぼすと、マットの直後の席が切札を出す義務を負う。
	if len(g.currentTrick) == 0 || led == 0 {
		return all
	}
	var follow, trumps, matto []int
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		switch {
		case tarocchiniIsMatto(c):
			matto = append(matto, i)
		case c.GetDesign() == led:
			follow = append(follow, i)
		case tarocchiniIsTrump(c):
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
func (g *Tarocchini) GetPlayableIndices(playerIdx int) []int {
	return g.GetValidPlayIndices(playerIdx)
}

// PlayerPlay 人間がカードを出す。
func (g *Tarocchini) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TarocchiniPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if !tarocchiniContains(g.GetValidPlayIndices(g.currentPlayerIdx), cardIndex) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	g.playCard(g.currentPlayerIdx, player.RemoveCard(cardIndex))
	return nil
}

// tarocchiniContains slice に値が含まれるか。
func tarocchiniContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// CpuPlay 現在の手番が CPU の場合に 1 枚出す。
func (g *Tarocchini) CpuPlay() {
	if g.gameEndFlag || g.phase != TarocchiniPhasePlay {
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
func (g *Tarocchini) cpuSelectPlayCard(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	if g.config.CpuDifficulty == TarocchiniCpuDifficultyEasy {
		return valid[g.rng.Intn(len(valid))]
	}
	led := g.ledSuit()
	best, bestRank := valid[0], -1<<31
	for _, i := range valid {
		r := tarocchiniWinRank(p.GetCard(i), led)
		if r > bestRank {
			best, bestRank = i, r
		}
	}
	// リード時は最強札、フォロー時は取れるなら取り、取れないなら最弱札。
	// **マットだけがリードされた局面はまだ「誰も勝っていない」。**マットは
	// トリックを取らないので、リードスートが決まるまでは実質リードと同じ。
	// ここを follow として扱うと、味方がマットを出した局面で「味方が勝っている」
	// と誤判定し、取りに行くべきトリックをダックしてしまう。
	if len(g.currentTrick) == 0 || led == 0 {
		return best
	}
	winIdx := tarocchiniTrickWinnerOf(g.currentTrick, led)
	if TarocchiniTeamOf(winIdx) == TarocchiniTeamOf(playerIdx) {
		return g.lowestOf(playerIdx, valid, led)
	}
	trial := append(append([]*TrickCard(nil), g.currentTrick...),
		&TrickCard{PlayerIdx: playerIdx, Card: p.GetCard(best)})
	if tarocchiniTrickWinnerOf(trial, led) == playerIdx {
		return best
	}
	return g.lowestOf(playerIdx, valid, led)
}

// lowestOf 候補のうち最も弱い札の位置を返す。
func (g *Tarocchini) lowestOf(playerIdx int, valid []int, led int) int {
	p := g.players[playerIdx]
	low, lowRank := valid[0], 1<<31-1
	for _, i := range valid {
		if r := tarocchiniWinRank(p.GetCard(i), led); r < lowRank {
			low, lowRank = i, r
		}
	}
	return low
}

// playCard 1 枚を場に出し、トリックが揃えばフェーズを進める。
func (g *Tarocchini) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", "", []*Card{card})
	if len(g.currentTrick) < TarocchiniPlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TarocchiniPlayerCnt
		return
	}
	g.phase = TarocchiniPhaseTrickEnd
}

// ResolveTrick トリックの勝者を決める。
func (g *Tarocchini) ResolveTrick() {
	if g.phase != TarocchiniPhaseTrickEnd {
		return
	}
	winnerIdx := tarocchiniTrickWinnerOf(g.currentTrick, g.ledSuit())
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
	if g.trickNumber >= TarocchiniTrickCount {
		g.settleRound()
	}
}

// NextTrick 次のトリックを開始する。
func (g *Tarocchini) NextTrick() {
	if g.phase != TarocchiniPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = TarocchiniPhasePlay
}

// settleRound ラウンドを精算する。獲得トリック数をチーム得点に足す。
//
// **最終トリックにボーナスを与える。**タロッキーニは最後のトリックを取った側に
// 加点する慣習があり、終盤で切り札を温存する動機になる。
func (g *Tarocchini) settleRound() {
	g.phase = TarocchiniPhaseRoundEnd
	for seat, n := range g.roundTricks {
		g.teamScores[TarocchiniTeamOf(seat)] += n
	}
	if g.lastTrickWinner >= 0 {
		g.teamScores[TarocchiniTeamOf(g.lastTrickWinner)] += TarocchiniLastTrickBonus
	}
	// スカルトの 2 枚はディーラー側のチームの獲得札として 1 点扱い。
	if len(g.scarto) > 0 {
		g.teamScores[TarocchiniTeamOf(g.dealerIdx)] += len(g.scarto)
	}
	g.appendLog(-1, "settle", fmt.Sprintf("ラウンド %d 終了", g.roundNumber), nil)
}

// TarocchiniLastTrickBonus 最終トリック獲得ボーナス。
const TarocchiniLastTrickBonus = 2

// ScoreRound ラウンドを締め、規定局数ならマッチを終える。
func (g *Tarocchini) ScoreRound() {
	if g.phase != TarocchiniPhaseRoundEnd {
		return
	}
	if g.roundNumber >= g.config.TargetRounds {
		g.finishMatch()
	}
}

// --- アクセサ ---

// GetPhase 現在のフェーズ。
func (g *Tarocchini) GetPhase() TarocchiniPhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *Tarocchini) SetPhase(p TarocchiniPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号。
func (g *Tarocchini) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号。
func (g *Tarocchini) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 現在の手番。
func (g *Tarocchini) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番を設定する (テスト用)。
func (g *Tarocchini) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック。
func (g *Tarocchini) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLeadPlayerIdx リード席。
func (g *Tarocchini) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetDealerIdx ディーラー席。
func (g *Tarocchini) GetDealerIdx() int { return g.dealerIdx }

// GetLastTrickWinner 直前のトリックの勝者席 (-1 = 未確定)。
func (g *Tarocchini) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetScartoSize ディーラーが捨てた枚数。
func (g *Tarocchini) GetScartoSize() int { return len(g.scarto) }

// GetTeamScores チーム別の累計得点。
func (g *Tarocchini) GetTeamScores() [tarocchiniTeamCnt]int { return g.teamScores }

// GetRoundTricks 現ラウンドの席別獲得トリック数。
func (g *Tarocchini) GetRoundTricks() [TarocchiniPlayerCnt]int { return g.roundTricks }

// GetGameEndFlag ゲーム終了フラグ。
func (g *Tarocchini) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1 = 未確定 / 同点)。
func (g *Tarocchini) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数。
func (g *Tarocchini) GetPlayerCnt() int { return TarocchiniPlayerCnt }

// GetPlayer 指定席のプレイヤー。
func (g *Tarocchini) GetPlayer(i int) *TarocchiniPlayer {
	return getPlayer(g.players, i)
}

// GetPlayers 全プレイヤー。
func (g *Tarocchini) GetPlayers() []*TarocchiniPlayer { return g.players }

// GetConfig ゲーム設定。
func (g *Tarocchini) GetConfig() TarocchiniConfig { return g.config }

// SetConfig ゲーム設定を差し替える。
func (g *Tarocchini) SetConfig(c TarocchiniConfig) { g.config = c }

// GetActionLog 棋譜。
func (g *Tarocchini) GetActionLog() []*ActionLogEntry { return g.actionLog }

// --- KV 永続化 ---

// tarocchiniJSON is the JSON wire format for Tarocchini.
//
// **全フィールドが非公開なので専用のコーデックが要る。**これを省くと
// Cloudflare Worker のセッション復元が空のゲームを返し、リクエストのたびに
// 手札も得点も消える。Ganjifa (#4661) と Vira (#4660) で同じ穴を開けた。
// deck / rng は載せない —— 配り終えた後の deck は残り札を持たず、rng は
// 復元後に張り直せばよい。
type tarocchiniJSON struct {
	Players          []*TarocchiniPlayer      `json:"pl"`
	Config           TarocchiniConfig         `json:"cfg"`
	Phase            TarocchiniPhase          `json:"ph"`
	RoundNumber      int                      `json:"rn"`
	TrickNumber      int                      `json:"tn"`
	CurrentPlayerIdx int                      `json:"cpi"`
	CurrentTrick     []*TrickCard             `json:"ct"`
	LeadPlayerIdx    int                      `json:"lpi"`
	DealerIdx        int                      `json:"di"`
	Scarto           []*Card                  `json:"sc"`
	LastTrickWinner  int                      `json:"ltw"`
	TeamScores       [tarocchiniTeamCnt]int   `json:"ts"`
	RoundTricks      [TarocchiniPlayerCnt]int `json:"rt"`
	GameEndFlag      bool                     `json:"gef"`
	WinnerTeam       int                      `json:"wt"`
	ActionLog        []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Tarocchini) MarshalJSON() ([]byte, error) {
	return json.Marshal(tarocchiniJSON{
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

// tarocchiniMaxSliceLen caps slice sizes during deserialisation.
const tarocchiniMaxSliceLen = 5000

var (
	errTarocchiniOversized      = errors.New("tarocchini: input array exceeds maximum allowed size")
	errTarocchiniInvalidPlayers = errors.New("tarocchini: invalid player count")
	errTarocchiniInvalidTrick   = errors.New("tarocchini: invalid trick card")
	errTarocchiniInvalidState   = errors.New("tarocchini: invalid state values in json")
)

// UnmarshalJSON implements json.Unmarshaler.
//
// **デザインは 1..6 を許す。**52 枚デッキ用の 1..4 を持ち込むと、切り札 (5) と
// マット (6) を含むセッションが全て復元不能になる。
func (g *Tarocchini) UnmarshalJSON(data []byte) error {
	var j tarocchiniJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tarocchiniMaxSliceLen || len(j.CurrentTrick) > tarocchiniMaxSliceLen ||
		len(j.ActionLog) > tarocchiniMaxSliceLen || len(j.Scarto) > tarocchiniMaxSliceLen {
		return errTarocchiniOversized
	}
	if len(j.Players) != TarocchiniPlayerCnt {
		return errTarocchiniInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errTarocchiniInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= TarocchiniPlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= TarocchiniPlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= TarocchiniPlayerCnt ||
		j.LastTrickWinner < -1 || j.LastTrickWinner >= TarocchiniPlayerCnt ||
		j.WinnerTeam < -1 || j.WinnerTeam >= tarocchiniTeamCnt ||
		len(j.Scarto) > TarocchiniSurplus ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > TarocchiniTrickCount ||
		j.Phase < TarocchiniPhaseScarto || j.Phase > TarocchiniPhaseGameEnd {
		return errTarocchiniInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= TarocchiniPlayerCnt {
			return errTarocchiniInvalidTrick
		}
		if d := tc.Card.GetDesign(); d < 1 || d > TarocchiniMattoDesign {
			return errTarocchiniInvalidTrick
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
	// **復元したら必ず乱数源を張り直す。**Worker は毎リクエスト KV から組み直し、
	// SetRand は一度も呼ばれない。shuffle() だけが nil を遅延初期化していたので、
	// `cpuSelectPlayCard` の Easy 分岐が nil の *rand.Rand を触って落ちていた。
	// 個々の呼び出し側でガードするのではなく、ここで一度入れて構造的に安全にする。
	g.rng = rand.New(rand.NewSource(rand.Int63()))
	return nil
}

// TarocchiniHint ヒント情報。
type TarocchiniHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// GetHint 人間プレイヤーへのヒント。手番でなければ nil。
func (g *Tarocchini) GetHint() *TarocchiniHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != TarocchiniPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.GetValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuSelectPlayCard(human, valid)
	return &TarocchiniHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason ヒント理由キーを判定する。
//
// **パパを勧めるときは専用の理由を返す。**「後から出した方が勝つ」を知らないと
// 同じ助言が真逆の手に読めてしまうので、パパかどうかをキーに含める。
func (g *Tarocchini) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	switch {
	case card == nil:
		return "lead_low"
	case TarocchiniIsPapa(card):
		return "play_papa"
	case tarocchiniIsMatto(card):
		return "play_matto"
	case len(g.currentTrick) == 0 && tarocchiniIsTrump(card):
		return "lead_trump"
	case len(g.currentTrick) == 0:
		return "lead_low"
	case tarocchiniIsTrump(card):
		return "follow_trump"
	default:
		return "follow_low"
	}
}
