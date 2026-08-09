//go:build !js || !wasm || extra2

// Package domain アリュエット (Aluette) のドメインモデル。
//
// Aluette はフランス西部 (ヴァンデ/ブルターニュ) の 48 枚スペイン式デッキを
// 用いる 4 人・2 対 2 の固定チーム制トリックテイキングゲーム。
//
// # このゲームを他と分けているもの
//
//  1. **切り札が無い。**
//  2. **強さがランクではなく「特定の 1 枚」で決まる。**最強の 6 枚 (リュエット) は
//     スートとランクの組で名指しされた個別の札で、4 スートに散らばっている。
//  3. **フォロー義務が無い。**どの札もいつでも出せる。
//
// この 3 つは互いに支え合っている。リュエットは 4 スートに散らばっているので、
// フォロー義務を課すとほとんどの局面で最強札が出せなくなり、序列表そのものが
// 空文化する。issue #4412 の要件2は「リードスートに従う義務あり」としていたが、
// 同 issue が「追加するメリット」に挙げた「ランクではない独自序列」「切札が無い」
// と両立しないため、**実ゲームに合わせてフォロー義務なしを採用**した (issue の
// コメントに判断根拠を記載)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// AluettePlayerCnt プレイヤー数 (人間 1 + CPU 3)。
const AluettePlayerCnt = 4

// AluetteSuitCnt スート数。
const AluetteSuitCnt = 4

// aluetteValues 各スートに存在する 12 種のランク。
//
// **スペイン式 48 枚。**Tute などが使う 40 枚デッキ (8・9 抜き) とは違い、
// Aluette は 8 と 9 を含む。9 は 2 枚がリュエットになるので落とせない。
var aluetteValues = [...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13}

// AluetteDeckSize デッキ枚数 (4 スート × 12 種)。
const AluetteDeckSize = AluetteSuitCnt * len(aluetteValues)

// AluetteHandSize 各プレイヤーの手札枚数。
//
// **48 は 4 で割り切れる (12 枚ずつ) が、全部は配らない。**issue #4412 の要件は
// 「5 トリックを行う」「5 トリック中 3 勝したペアがそのセットの 1 点を獲得」と
// 明示しており、内部矛盾が無いのでそれに従う。残りの札はそのメーヌでは使わない。
// 史料によっては 9 枚配り説もあるため、**枚数はこの 1 定数に隔離**してある ——
// 変えるならここだけを触ればよい。
const AluetteHandSize = 5

// AluetteTrickCount 1 メーヌ (mène) のトリック数。
const AluetteTrickCount = AluetteHandSize

// AluetteTricksToWin メーヌを取るのに必要なトリック数 (5 戦 3 勝)。
const AluetteTricksToWin = AluetteTrickCount/2 + 1

// aluetteTeamCnt チーム数。
const aluetteTeamCnt = 2

// AluetteTeamOf 席のチーム番号を返す。対面同士 (0-2 / 1-3) が組む。
func AluetteTeamOf(seat int) int { return seat % 2 }

// --- 序列 ---

// aluetteLuette は名前を持つ最強札 1 枚の定義。
type aluetteLuette struct {
	design int
	value  int
	name   string
}

// aluetteLuettes は「リュエット」と呼ばれる 6 枚。**強い順**に並べる。
//
// ラテンスートと本実装の design の対応は慣用に従う:
// 1=Espadas(剣) / 2=Bastos(棍棒) / 3=Copas(聖杯) / 4=Oros(金貨)。
//
// **この 6 枚の同定と順序がゲームの中心。**ここを触ると強さの意味がすべて変わる
// ので、序列は必ずこの表を経由して引く (生の value で比較してはならない)。
var aluetteLuettes = [...]aluetteLuette{
	{design: 4, value: 3, name: "Monsieur"},  // 金貨の3
	{design: 3, value: 3, name: "Madame"},    // 聖杯の3
	{design: 3, value: 2, name: "Borgne"},    // 聖杯の2
	{design: 4, value: 2, name: "Vache"},     // 金貨の2
	{design: 3, value: 9, name: "GrandNeuf"}, // 聖杯の9
	{design: 4, value: 9, name: "PetitNeuf"}, // 金貨の9
}

// aluetteOrdinaryOrder はリュエット以外の札の強さ順 (**強い順**)。
//
// スートは一切見ない。3 > 2 > A > 王 > 騎 > 従 > 9 > 8 > 7 > 6 > 5 > 4。
var aluetteOrdinaryOrder = [...]int{3, 2, 1, 13, 12, 11, 9, 8, 7, 6, 5, 4}

// AluetteLuetteName はその札がリュエットならその名前を、違えば空文字を返す。
func AluetteLuetteName(c *Card) string {
	if c == nil {
		return ""
	}
	for _, l := range aluetteLuettes {
		if c.GetDesign() == l.design && c.GetValue() == l.value {
			return l.name
		}
	}
	return ""
}

// AluetteLuetteInfo は表示用に公開するリュエット 1 枚の定義。
type AluetteLuetteInfo struct {
	// Design スート (3=Copas, 4=Oros)
	Design int `json:"design"`
	// Value ランク
	Value int `json:"value"`
	// Name 呼び名 (Monsieur, Madame, ...)
	Name string `json:"name"`
}

// AluetteLuetteTable はリュエット 6 枚を**強い順**に返す。
//
// **UI 側で 6 枚を書き写させないための出口。**どの札がリュエットかを知らずに
// アリュエットは遊べない一方、序列表をフロントに複製すると必ずドメインとずれる。
// 表そのものを配ることで SSoT をドメインに残す。
func AluetteLuetteTable() []AluetteLuetteInfo {
	out := make([]AluetteLuetteInfo, 0, len(aluetteLuettes))
	for _, l := range aluetteLuettes {
		out = append(out, AluetteLuetteInfo{Design: l.design, Value: l.value, Name: l.name})
	}
	return out
}

// AluetteRank は札の強さを返す。**数値が大きいほど強い。**
//
// リュエット 6 枚が最上位を占め、残りは aluetteOrdinaryOrder の順。
// **生の value で比較してはならない** —— 金貨の3 (Monsieur) と剣の3 は同じ
// value 3 でも強さがまったく違う。
func AluetteRank(c *Card) int {
	if c == nil {
		return -1
	}
	// リュエットは上位を占める。表の先頭ほど強い。
	for i, l := range aluetteLuettes {
		if c.GetDesign() == l.design && c.GetValue() == l.value {
			return len(aluetteOrdinaryOrder) + len(aluetteLuettes) - i
		}
	}
	for i, v := range aluetteOrdinaryOrder {
		if c.GetValue() == v {
			return len(aluetteOrdinaryOrder) - i
		}
	}
	return -1
}

// aluetteTrickWinnerOf は与えられたトリックの勝者席を返す。
//
// **切り札もリードスートも見ない。**強さは AluetteRank だけで決まる。同ランクが
// 複数出た場合 (リュエット以外は同ランクが 4 枚ずつある) は**最初に出した側が勝つ**
// —— 後から同じ強さを重ねても奪えない。
func aluetteTrickWinnerOf(trick []*TrickCard) int {
	if len(trick) == 0 {
		return 0
	}
	winIdx, winRank := -1, -1
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		if winIdx < 0 {
			winIdx = tc.PlayerIdx
		}
		if r := AluetteRank(tc.Card); r > winRank {
			winRank, winIdx = r, tc.PlayerIdx
		}
	}
	if winIdx < 0 {
		return 0
	}
	return winIdx
}

// buildAluetteDeck 48 枚のデッキを組む。
func buildAluetteDeck() []*Card {
	deck := make([]*Card, 0, AluetteDeckSize)
	for suit := 1; suit <= AluetteSuitCnt; suit++ {
		for _, val := range aluetteValues {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	return deck
}

// AluettePhase ゲームフェーズ。入札もスカルトも無い。
type AluettePhase int

const (
	// AluettePhasePlay トリックプレイ。
	AluettePhasePlay AluettePhase = iota
	// AluettePhaseTrickEnd トリック終了 (結果表示待ち)。
	AluettePhaseTrickEnd
	// AluettePhaseRoundEnd メーヌ終了 (精算済み)。
	AluettePhaseRoundEnd
	// AluettePhaseGameEnd マッチ終了。
	AluettePhaseGameEnd
)

// Aluette アリュエットのゲーム本体。
type Aluette struct {
	players          []*AluettePlayer
	config           AluetteConfig
	rng              *rand.Rand
	deck             []*Card
	deckDrawCnt      int
	phase            AluettePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	lastTrickWinner  int
	teamScores       [aluetteTeamCnt]int // メーヌ単位の獲得点
	roundTricks      [AluettePlayerCnt]int
	gameEndFlag      bool
	winnerTeam       int // -1 = 未確定 (同点)
	actionLog        []*ActionLogEntry
}

// NewAluette コンストラクタ。
func NewAluette(players []*AluettePlayer, config AluetteConfig) *Aluette {
	return &Aluette{
		players: players,
		config:  config,
		// **本番の乱数源はここで入れる。**入れ忘れると山が並べ替わらず毎回同じ配りに
		// なる (Ganjifa #4661)。種は rand.Int63() から取る。
		rng:        rand.New(rand.NewSource(rand.Int63())),
		winnerTeam: -1,
	}
}

// NewDefaultAluette 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultAluette() *Aluette {
	players := make([]*AluettePlayer, AluettePlayerCnt)
	players[0] = NewAluettePlayer(true)
	for i := 1; i < AluettePlayerCnt; i++ {
		players[i] = NewAluettePlayer(false)
	}
	return NewAluette(players, DefaultAluetteConfig())
}

// SetRand 乱数源を差し替える (テスト用)。
func (g *Aluette) SetRand(r *rand.Rand) { g.rng = r }

// Reset ゲームを初期化する。
func (g *Aluette) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [aluetteTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のメーヌを開始する。
func (g *Aluette) NextRound() {
	if g.phase != AluettePhaseRoundEnd {
		return
	}
	if g.matchDecided() {
		g.finishMatch()
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % AluettePlayerCnt
	g.startRound()
}

// startRound 手札を配り、プレイフェーズを開始する。
func (g *Aluette) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % AluettePlayerCnt
	g.lastTrickWinner = -1
	g.roundTricks = [AluettePlayerCnt]int{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = AluettePhasePlay
}

// deal 各プレイヤーへ AluetteHandSize 枚を配る。残りはそのメーヌでは使わない。
func (g *Aluette) deal() {
	g.deck = buildAluetteDeck()
	g.shuffle()
	g.deckDrawCnt = 0
	for i := 0; i < AluetteHandSize; i++ {
		for j := 0; j < AluettePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % AluettePlayerCnt
			if c := g.drawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// shuffle 山を並べ替える。
func (g *Aluette) shuffle() {
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	g.rng.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
}

// drawCard デッキから 1 枚配る (尽きたら nil)。
func (g *Aluette) drawCard() *Card {
	if g.deckDrawCnt >= len(g.deck) {
		return nil
	}
	card := g.deck[g.deckDrawCnt]
	card.SetDraw(true)
	g.deckDrawCnt++
	return card
}

// sortAllHands 手札を強い順に整列する。
//
// **値ではなく AluetteRank で並べる。**値順に並べると金貨の3 (Monsieur) が
// 剣の3 と同じ位置に来てしまい、手札を見て強さを判断できない。
func (g *Aluette) sortAllHands() {
	for _, p := range g.players {
		cards := make([]*Card, 0, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			cards = append(cards, p.GetCard(i))
		}
		sort.SliceStable(cards, func(i, j int) bool {
			return AluetteRank(cards[i]) > AluetteRank(cards[j])
		})
		p.Reset()
		for _, c := range cards {
			p.AddCard(c)
		}
	}
}

// matchDecided 規定点に到達したチームがあるか。
func (g *Aluette) matchDecided() bool {
	for _, s := range g.teamScores {
		if s >= g.config.TargetPoints {
			return true
		}
	}
	return false
}

// finishMatch マッチを終え、得点の多いチームを勝者にする。
func (g *Aluette) finishMatch() {
	g.gameEndFlag = true
	g.phase = AluettePhaseGameEnd
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
func (g *Aluette) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.trickNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- トリックプレイ ---

// IsHumanTurn 人間のプレイ手番か。
func (g *Aluette) IsHumanTurn() bool {
	return g.phase == AluettePhasePlay && g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetValidPlayIndices 出せる手札の位置を返す。
//
// **常に全部。**アリュエットにはフォロー義務も切り札義務も無い。ここで絞ると、
// 4 スートに散らばるリュエットが出せなくなり序列表が空文化する (#4412 の要件2は
// フォロー義務ありとしていたが、同 issue の「独自序列・切札なし」と両立しない)。
func (g *Aluette) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	p := g.players[playerIdx]
	all := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		all = append(all, i)
	}
	return all
}

// GetPlayableIndices プレイ可能なカードの位置。
func (g *Aluette) GetPlayableIndices(playerIdx int) []int {
	return g.GetValidPlayIndices(playerIdx)
}

// PlayerPlay 人間がカードを出す。
func (g *Aluette) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != AluettePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	g.playCard(g.currentPlayerIdx, player.RemoveCard(cardIndex))
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 枚出す。
func (g *Aluette) CpuPlay() {
	if g.gameEndFlag || g.phase != AluettePhasePlay {
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
	// それを playCard に渡すと nil デリファレンスで落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// cpuSelectPlayCard CPU が出す札の位置を選ぶ。
//
// **難易度の差は「リュエットをいつ使うか」に出る。**強さが絶対（切り札もフォロー
// 義務も無い）なので、リュエットは**いつ出しても勝つ**。つまり抱えておく損が無く、
// 使い所を選べる側が強い。
//
//   - Easy:   ランダム。
//   - Normal: 常に最強札を出す。取れる時は取るが、3 を出せば済む場面で
//     Monsieur を切ってしまう。
//   - Hard:   リードは最弱札で相手の強い札を吐かせ、フォローは**勝てる中で
//     最も弱い札**で取る。5 トリック中 3 勝すればよいので、確実に勝てる札を
//     終盤まで残せる側が有利。
func (g *Aluette) cpuSelectPlayCard(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	if g.config.CpuDifficulty == AluetteCpuDifficultyEasy {
		return valid[g.rng.Intn(len(valid))]
	}
	hard := g.config.CpuDifficulty == AluetteCpuDifficultyHard

	if len(g.currentTrick) == 0 {
		if hard {
			return g.lowestOf(playerIdx, valid)
		}
		return g.highestOf(playerIdx, valid)
	}
	// 味方が既に勝っているトリックは奪わない。
	winIdx := aluetteTrickWinnerOf(g.currentTrick)
	if AluetteTeamOf(winIdx) == AluetteTeamOf(playerIdx) {
		return g.lowestOf(playerIdx, valid)
	}
	if hard {
		if i, ok := g.cheapestWinnerOf(playerIdx, valid); ok {
			return i
		}
		return g.lowestOf(playerIdx, valid)
	}
	best := g.highestOf(playerIdx, valid)
	if g.winsTrick(playerIdx, p.GetCard(best)) {
		return best
	}
	return g.lowestOf(playerIdx, valid)
}

// winsTrick その札を今出したらトリックを取れるかを返す。
func (g *Aluette) winsTrick(playerIdx int, card *Card) bool {
	trial := append(append([]*TrickCard(nil), g.currentTrick...),
		&TrickCard{PlayerIdx: playerIdx, Card: card})
	return aluetteTrickWinnerOf(trial) == playerIdx
}

// cheapestWinnerOf 勝てる札のうち最も弱いものの位置を返す。第2返り値は勝てる札の有無。
//
// **勝つのに Monsieur は要らない。**リュエットはいつ出しても勝つので、3 で足りる
// トリックに最強札を使うと、その 1 枚ぶんだけ後半の勝てるトリックが減る。
func (g *Aluette) cheapestWinnerOf(playerIdx int, valid []int) (int, bool) {
	p := g.players[playerIdx]
	best, bestRank, found := 0, 1<<31-1, false
	for _, i := range valid {
		c := p.GetCard(i)
		if c == nil || !g.winsTrick(playerIdx, c) {
			continue
		}
		if r := AluetteRank(c); r < bestRank {
			best, bestRank, found = i, r, true
		}
	}
	return best, found
}

// highestOf 候補のうち最も強い札の位置を返す。
func (g *Aluette) highestOf(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	high, highRank := valid[0], -1
	for _, i := range valid {
		if r := AluetteRank(p.GetCard(i)); r > highRank {
			high, highRank = i, r
		}
	}
	return high
}

// lowestOf 候補のうち最も弱い札の位置を返す。
func (g *Aluette) lowestOf(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	low, lowRank := valid[0], 1<<31-1
	for _, i := range valid {
		if r := AluetteRank(p.GetCard(i)); r < lowRank {
			low, lowRank = i, r
		}
	}
	return low
}

// playCard 1 枚を場に出し、トリックが揃えばフェーズを進める。
func (g *Aluette) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", AluetteLuetteName(card), []*Card{card})
	if len(g.currentTrick) < AluettePlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % AluettePlayerCnt
		return
	}
	g.phase = AluettePhaseTrickEnd
}

// ResolveTrick トリックの勝者を決める。
func (g *Aluette) ResolveTrick() {
	if g.phase != AluettePhaseTrickEnd {
		return
	}
	winnerIdx := aluetteTrickWinnerOf(g.currentTrick)
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
	if g.trickNumber >= AluetteTrickCount {
		g.settleRound()
	}
}

// NextTrick 次のトリックを開始する。
func (g *Aluette) NextTrick() {
	if g.phase != AluettePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = AluettePhasePlay
}

// settleRound メーヌを精算する。
//
// **トリック数ではなく「メーヌを取ったか」で 1 点。**過半数を取ったチームに 1 点
// 入る方式なので、4-1 でも 3-2 でも得点は同じ 1 点。
func (g *Aluette) settleRound() {
	g.phase = AluettePhaseRoundEnd
	var teamTricks [aluetteTeamCnt]int
	for seat, n := range g.roundTricks {
		teamTricks[AluetteTeamOf(seat)] += n
	}
	switch {
	case teamTricks[0] >= AluetteTricksToWin:
		g.teamScores[0]++
	case teamTricks[1] >= AluetteTricksToWin:
		g.teamScores[1]++
	}
	g.appendLog(-1, "settle", fmt.Sprintf("メーヌ %d 終了", g.roundNumber), nil)
}

// ScoreRound メーヌを締め、規定点に達していればマッチを終える。
func (g *Aluette) ScoreRound() {
	if g.phase != AluettePhaseRoundEnd {
		return
	}
	if g.matchDecided() {
		g.finishMatch()
	}
}

// --- アクセサ ---

// GetPhase 現在のフェーズ。
func (g *Aluette) GetPhase() AluettePhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *Aluette) SetPhase(p AluettePhase) { g.phase = p }

// GetRoundNumber 現在のメーヌ番号。
func (g *Aluette) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号。
func (g *Aluette) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 現在の手番。
func (g *Aluette) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx 手番を設定する (テスト用)。
func (g *Aluette) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック。
func (g *Aluette) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLeadPlayerIdx リード席。
func (g *Aluette) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetDealerIdx ディーラー席。
func (g *Aluette) GetDealerIdx() int { return g.dealerIdx }

// GetLastTrickWinner 直前のトリックの勝者席 (-1 = 未確定)。
func (g *Aluette) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetTeamScores チーム別の累計メーヌ数。
func (g *Aluette) GetTeamScores() [aluetteTeamCnt]int { return g.teamScores }

// GetRoundTricks 現メーヌの席別獲得トリック数。
func (g *Aluette) GetRoundTricks() [AluettePlayerCnt]int { return g.roundTricks }

// GetGameEndFlag ゲーム終了フラグ。
func (g *Aluette) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1 = 未確定 / 同点)。
func (g *Aluette) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数。
func (g *Aluette) GetPlayerCnt() int { return AluettePlayerCnt }

// GetPlayer 指定席のプレイヤー。
func (g *Aluette) GetPlayer(i int) *AluettePlayer {
	return getPlayer(g.players, i)
}

// GetPlayers 全プレイヤー。
func (g *Aluette) GetPlayers() []*AluettePlayer { return g.players }

// GetConfig ゲーム設定。
func (g *Aluette) GetConfig() AluetteConfig { return g.config }

// SetConfig ゲーム設定を差し替える。
func (g *Aluette) SetConfig(c AluetteConfig) { g.config = c }

// GetActionLog 棋譜。
func (g *Aluette) GetActionLog() []*ActionLogEntry { return g.actionLog }

// AluetteHint ヒント情報。
type AluetteHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// GetHint 人間プレイヤーへのヒント。手番でなければ nil。
func (g *Aluette) GetHint() *AluetteHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.phase != AluettePhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.GetValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuSelectPlayCard(human, valid)
	return &AluetteHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason ヒント理由キーを判定する。
//
// **リュエットを勧めるときは専用の理由を返す。**名前を持つ 6 枚は手札の中で
// 見た目が他と変わらないので、「なぜその札か」を言わないと助言が読めない。
func (g *Aluette) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	switch {
	case card == nil:
		return "lead_low"
	case AluetteLuetteName(card) != "":
		return "play_luette"
	case len(g.currentTrick) == 0:
		return "lead_low"
	case aluetteTrickWinnerOf(g.currentTrick) >= 0 &&
		AluetteTeamOf(aluetteTrickWinnerOf(g.currentTrick)) == AluetteTeamOf(playerIdx):
		return "partner_winning"
	default:
		return "follow_low"
	}
}

// --- KV 永続化 ---

// aluetteJSON is the JSON wire format for Aluette.
//
// **全フィールドが非公開なので専用のコーデックが要る。**これを省くと Cloudflare
// Worker のセッション復元が空のゲームを返す (Ganjifa #4661 / Vira #4660)。
type aluetteJSON struct {
	Players          []*AluettePlayer      `json:"pl"`
	Config           AluetteConfig         `json:"cfg"`
	Phase            AluettePhase          `json:"ph"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	CurrentPlayerIdx int                   `json:"cpi"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	LeadPlayerIdx    int                   `json:"lpi"`
	DealerIdx        int                   `json:"di"`
	LastTrickWinner  int                   `json:"ltw"`
	TeamScores       [aluetteTeamCnt]int   `json:"ts"`
	RoundTricks      [AluettePlayerCnt]int `json:"rt"`
	GameEndFlag      bool                  `json:"gef"`
	WinnerTeam       int                   `json:"wt"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Aluette) MarshalJSON() ([]byte, error) {
	return json.Marshal(aluetteJSON{
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		LastTrickWinner:  g.lastTrickWinner,
		TeamScores:       g.teamScores,
		RoundTricks:      g.roundTricks,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// aluetteMaxSliceLen caps slice sizes during deserialisation.
const aluetteMaxSliceLen = 5000

var (
	errAluetteOversized      = errors.New("aluette: input array exceeds maximum allowed size")
	errAluetteInvalidPlayers = errors.New("aluette: invalid player count")
	errAluetteInvalidTrick   = errors.New("aluette: invalid trick card")
	errAluetteInvalidState   = errors.New("aluette: invalid state values in json")
)

// UnmarshalJSON implements json.Unmarshaler.
func (g *Aluette) UnmarshalJSON(data []byte) error {
	var j aluetteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > aluetteMaxSliceLen || len(j.CurrentTrick) > aluetteMaxSliceLen ||
		len(j.ActionLog) > aluetteMaxSliceLen {
		return errAluetteOversized
	}
	if len(j.Players) != AluettePlayerCnt {
		return errAluetteInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errAluetteInvalidPlayers
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= AluettePlayerCnt ||
		j.LeadPlayerIdx < 0 || j.LeadPlayerIdx >= AluettePlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= AluettePlayerCnt ||
		j.LastTrickWinner < -1 || j.LastTrickWinner >= AluettePlayerCnt ||
		j.WinnerTeam < -1 || j.WinnerTeam >= aluetteTeamCnt ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 1 || j.TrickNumber > AluetteTrickCount ||
		j.Phase < AluettePhasePlay || j.Phase > AluettePhaseGameEnd {
		return errAluetteInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= AluettePlayerCnt {
			return errAluetteInvalidTrick
		}
		if d := tc.Card.GetDesign(); d < 1 || d > AluetteSuitCnt {
			return errAluetteInvalidTrick
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
	g.lastTrickWinner = j.LastTrickWinner
	g.teamScores = j.TeamScores
	g.roundTricks = j.RoundTricks
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	// **復元したら必ず乱数源を張り直す。**Worker は毎リクエスト KV から組み直すので
	// SetRand は呼ばれない。これを落とすと Easy 難易度の乱択が nil で落ちる (#4663)。
	g.rng = rand.New(rand.NewSource(rand.Int63()))
	return nil
}
