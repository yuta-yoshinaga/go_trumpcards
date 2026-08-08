//go:build !js || !wasm || extra3

// Package domain バスラ (Basra / Bastra) のドメインモデル。
//
// Basra はエジプト・レバント地方で親しまれる漁 (フィッシング/キャプチャ) 系カードゲーム。
// 標準 52 枚デッキを 4 人 (seat 0 が人間 + CPU 3 人, 個人戦) でプレイする。Pişti と
// Cassino の中間に位置し、同ランク捕獲・合計値捕獲・ジャックの一掃 (スイープ)・バスラ
// ボーナスを併せ持つ。
//
// # ルール概要
//
// 各プレイヤーへ BasraHandSize (4) 枚を配り、場に BasraInitialTableSize (4) 枚を表向きで
// 置く。手番では手札から 1 枚を出し、次のいずれかで場札を捕獲する。
//
//   - 出した数札 (A〜10) は、場札のうち「同ランク」または「合計値が出した札の値に等しい
//     組」をすべて捕獲できる (合計捕獲は複数の組を同時に取れる)。A は 1。
//   - Q / K は同ランク捕獲のみ (合計には参加しない)。
//   - ジャック (値 11) は場のジャック以外の全札を一掃 (スイープ) して捕獲する。場の別の
//     ジャックは残す。
//   - どの札も捕獲しなかった場合、その札は場に置かれる (トレイル)。
//
// # バスラ (ボーナス)
//
// ジャック以外の 1 枚で場札を「すべて」捕獲し場が空になった場合、それは "Basra" となり
// BasraScoreBasra 点のボーナスを得る (ジャックのスイープはバスラに含めない)。
//
// # 配り直し / 終局
//
// 全員の手札が尽きたら、山札から新たに 4 枚ずつ配る (場札は補充しない)。これを山札が
// 尽きるまで繰り返す。最後に場へ残った札は、最後に捕獲したプレイヤーが取る。
//
// # 得点 (calcFinalScore)
//
//   - 捕獲枚数が最多 (単独) のプレイヤー: +BasraScoreMostCards
//   - 7♦ を捕獲: +BasraScoreSevenDiamonds ("big" ボーナス)
//   - 10♦ を捕獲: +BasraScoreTenDiamonds
//   - A を捕獲: 1 枚につき +BasraScoreAce
//   - バスラ: 1 回につき +BasraScoreBasra
//
// 最高点が勝者。同点なら複数勝者 (GetWinners が全員を返す)。1 ゲームは山札を配り切る
// 1 セッションで完結し、NextRound は新規ゲームの開始 (Reset) と同義。
//
// 本実装は extra ワーカーから到達可能なよう捕獲/得点ロジックをすべてインラインで持ち、
// extra 到達可能な NewTrumpCards(0) で 52 枚デッキを生成する。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// BasraPlayerCnt はバスラのプレイヤー数 (固定 4, 個人戦)。
const BasraPlayerCnt = 4

// BasraHandSize は 1 回の配布で各プレイヤーへ配る手札枚数。
const BasraHandSize = 4

// BasraInitialTableSize はゲーム開始時に場へ置くカード枚数。
const BasraInitialTableSize = 4

// BasraJackValue はジャックのカード値 (スイープ札)。
const BasraJackValue = 11

// BasraPhase はゲームフェーズ。
type BasraPhase int

// Basra のフェーズ定数
const (
	// BasraPhasePlay プレイ中 (カードを場へ出す)
	BasraPhasePlay BasraPhase = 0
	// BasraPhaseGameEnd ゲーム終了 (山札を配り切り、最終得点を確定)
	BasraPhaseGameEnd BasraPhase = 1
)

// Basra の得点定数
const (
	// BasraScoreMostCards 捕獲枚数最多 (単独) のボーナス
	BasraScoreMostCards = 3
	// BasraScoreAce A 1 枚のボーナス
	BasraScoreAce = 1
	// BasraScoreSevenDiamonds 7♦ のボーナス ("big")
	BasraScoreSevenDiamonds = 3
	// BasraScoreTenDiamonds 10♦ のボーナス
	BasraScoreTenDiamonds = 2
	// BasraScoreBasra バスラ (単独札での場一掃) 1 回のボーナス
	BasraScoreBasra = 10
)

// BasraScoreDetail は 1 ゲームの得点内訳。
type BasraScoreDetail struct {
	Cards            map[int]int // プレイヤー別捕獲枚数
	Aces             map[int]int // プレイヤー別エース数
	Basras           map[int]int // プレイヤー別バスラ数
	HasSevenDiamonds int         // 7♦ を捕獲したプレイヤー (-1 = なし)
	HasTenDiamonds   int         // 10♦ を捕獲したプレイヤー (-1 = なし)
	MostCards        int         // 捕獲枚数最多の単独プレイヤー (-1 = 同点/なし)
	Gained           map[int]int // プレイヤー別の得点
}

// BasraHint はヒント情報。
type BasraHint struct {
	CardIndices  []int  // 推奨手札インデックス
	TableIndices []int  // 推奨捕獲対象の場札インデックス
	Reason       string // ヒント理由キー
}

// basraState はゲーム進行状態。
type basraState struct {
	phase          BasraPhase
	currentTurn    int     // 現在の手番プレイヤー
	tableCards     []*Card // 場の札
	lastCaptureIdx int     // 最後に捕獲したプレイヤー (-1 = なし)
	packsDealt     int     // これまでに配ったパック数 (1 回の配布 = 4 枚/人)
	gameEndFlag    bool
	scored         bool // 最終得点を確定済みか (二重確定防止)
	winners        []int
	lastDealDetail *BasraScoreDetail
	actionLogBase
}

// Basra はバスラゲームの状態を保持する集約ルート。
type Basra struct {
	trumpCards *TrumpCards
	players    []*BasraPlayer
	config     BasraConfig
	state      basraState
}

// NewBasra はコンストラクタ。
func NewBasra(trumpCards *TrumpCards, players []*BasraPlayer, config BasraConfig) *Basra {
	return &Basra{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: basraState{
			phase:          BasraPhasePlay,
			lastCaptureIdx: -1,
		},
	}
}

// NewDefaultBasra は標準の 4 人構成 (1 human + 3 CPU) と DefaultBasraConfig で Basra を
// 生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultBasra() *Basra {
	players := make([]*BasraPlayer, BasraPlayerCnt)
	players[0] = NewBasraPlayer(true)
	for i := 1; i < BasraPlayerCnt; i++ {
		players[i] = NewBasraPlayer(false)
	}
	return NewBasra(newBasraDeck(), players, DefaultBasraConfig())
}

// newBasraDeck はバスラ用 52 枚デッキを生成する。NewTrumpCards はビルドタグ無しの
// TrumpCards.go にあり extra ワーカーからも到達可能。
func newBasraDeck() *TrumpCards {
	return NewTrumpCards(0)
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。
func (g *Basra) Reset() {
	for _, p := range g.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
		p.ResetBasra()
		p.ResetScore()
	}
	g.trumpCards = newBasraDeck()
	g.trumpCards.Shuffle()
	g.state = basraState{
		phase:          BasraPhasePlay,
		currentTurn:    0,
		lastCaptureIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.dealHands()
	g.dealInitialTable()
	g.sortHumanHand()
}

// NextRound はバスラでは新規ゲーム開始 (Reset) と同義。1 ゲームは山札を配り切る 1
// セッションで完結するため、終局後の続行は新しいゲームの開始として扱う。
func (g *Basra) NextRound() {
	g.Reset()
}

// dealHands は各プレイヤーへ BasraHandSize 枚配る。山札が尽きたら途中で終わる。
func (g *Basra) dealHands() {
	for k := 0; k < BasraHandSize; k++ {
		for i := 0; i < len(g.players); i++ {
			card := g.trumpCards.DrawCard()
			if card == nil {
				g.state.packsDealt++
				return
			}
			g.players[i].AddCard(card)
		}
	}
	g.state.packsDealt++
}

// dealInitialTable はゲーム開始時に BasraInitialTableSize 枚を場へ置く。
func (g *Basra) dealInitialTable() {
	for i := 0; i < BasraInitialTableSize; i++ {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.state.tableCards = append(g.state.tableCards, card)
	}
	g.appendLog(-1, "deal", fmt.Sprintf("dealt %d table cards", len(g.state.tableCards)),
		append([]*Card(nil), g.state.tableCards...))
}

// allHandsEmpty は全員の手札が空かどうか。
func (g *Basra) allHandsEmpty() bool {
	for _, p := range g.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// --- 捕獲ロジック (インライン) ---

// basraCardValue はカードのキャプチャ用の値を返す (A=1 … K=13)。
func basraCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue()
}

// basraIsJack はジャック (値 11, スイープ札) かどうか。
func basraIsJack(c *Card) bool {
	return c != nil && c.GetValue() == BasraJackValue
}

// basraIsFace は絵札 (J/Q/K, 値 11-13) かどうか。
func basraIsFace(c *Card) bool {
	return c != nil && c.GetValue() >= 11
}

// basraCanPartition は values を「各グループの合計が target に等しい」ように過不足なく
// 分割できるかを返す。単独の同ランク札 (値 == target) も 1 枚のグループとして扱えるため、
// 同ランク捕獲と合計捕獲を統一的に検証できる。target を超える値が含まれる場合は不可。
func basraCanPartition(values []int, target int) bool {
	if len(values) == 0 || target <= 0 {
		return false
	}
	sum := 0
	for _, v := range values {
		if v <= 0 || v > target {
			return false
		}
		sum += v
	}
	if sum%target != 0 {
		return false
	}
	buckets := sum / target
	vs := append([]int(nil), values...)
	sort.Sort(sort.Reverse(sort.IntSlice(vs)))
	remaining := make([]int, buckets)
	for i := range remaining {
		remaining[i] = target
	}
	return basraFillBuckets(vs, 0, remaining)
}

// basraFillBuckets は k-partition のバックトラッキング補助。
func basraFillBuckets(vs []int, idx int, remaining []int) bool {
	if idx == len(vs) {
		return true
	}
	v := vs[idx]
	seen := make(map[int]bool)
	for b := 0; b < len(remaining); b++ {
		if remaining[b] < v || seen[remaining[b]] {
			continue
		}
		seen[remaining[b]] = true
		remaining[b] -= v
		if basraFillBuckets(vs, idx+1, remaining) {
			return true
		}
		remaining[b] += v
	}
	return false
}

// basraValidateSelection は playedCard で tableIdxs を捕獲する選択が合法かを検証する。
// ジャックは呼び出し側で別処理するため、ここでは数札 / Q / K のみを扱う。
func (g *Basra) basraValidateSelection(playedCard *Card, tableIdxs []int) error {
	if len(tableIdxs) == 0 {
		return NewDomainError(ErrInvalidPlay, "capture requires at least one table card")
	}
	seen := make(map[int]bool)
	selected := make([]*Card, 0, len(tableIdxs))
	for _, idx := range tableIdxs {
		if idx < 0 || idx >= len(g.state.tableCards) || seen[idx] {
			return NewDomainError(ErrInvalidPlay, "invalid table card selection")
		}
		seen[idx] = true
		selected = append(selected, g.state.tableCards[idx])
	}
	pv := basraCardValue(playedCard)
	if basraIsFace(playedCard) {
		// Q / K は同ランク捕獲のみ。
		for _, c := range selected {
			if basraCardValue(c) != pv {
				return NewDomainError(ErrInvalidPlay, "face cards capture only by matching rank")
			}
		}
		return nil
	}
	// 数札 (A〜10): 選択は target=pv の組に過不足なく分割できること。
	vals := make([]int, len(selected))
	for i, c := range selected {
		vals[i] = basraCardValue(c)
	}
	if !basraCanPartition(vals, pv) {
		return NewDomainError(ErrInvalidPlay, "selected table cards do not sum/match the played card")
	}
	return nil
}

// basraFindCaptures は playedCard が捕獲できる場札インデックスの最大集合を返す。
// CPU と、ジャック/自動捕獲経路で使用する。捕獲不能なら空スライスを返す。
func (g *Basra) basraFindCaptures(playedCard *Card) []int {
	if basraIsJack(playedCard) {
		// ジャック: 場のジャック以外を全て捕獲。
		var out []int
		for i, c := range g.state.tableCards {
			if !basraIsJack(c) {
				out = append(out, i)
			}
		}
		return out
	}
	pv := basraCardValue(playedCard)
	if basraIsFace(playedCard) {
		// Q / K: 同ランクのみ。
		var out []int
		for i, c := range g.state.tableCards {
			if basraCardValue(c) == pv {
				out = append(out, i)
			}
		}
		return out
	}
	// 数札: グループ (同ランク単独 + 合計) を貪欲に繰り返し抽出する。
	used := make([]bool, len(g.state.tableCards))
	captured := make([]int, 0)
	for {
		group := g.basraFindOneGroup(pv, used)
		if group == nil {
			break
		}
		for _, idx := range group {
			used[idx] = true
			captured = append(captured, idx)
		}
	}
	sort.Ints(captured)
	return captured
}

// basraFindOneGroup は未使用の数札の中から、合計が target になる 1 グループ (同ランク単独を
// 含む) を探して返す。見つからなければ nil。絵札 (値 11-13) は数札グループに含めない。
func (g *Basra) basraFindOneGroup(target int, used []bool) []int {
	// まず同ランク単独 (値 == target) を優先。
	for i, c := range g.state.tableCards {
		if used[i] || basraIsFace(c) {
			continue
		}
		if basraCardValue(c) == target {
			return []int{i}
		}
	}
	// 次に 2 枚以上の合計組を探索する。
	var avail []int
	for i, c := range g.state.tableCards {
		if used[i] || basraIsFace(c) {
			continue
		}
		if basraCardValue(c) < target {
			avail = append(avail, i)
		}
	}
	return g.basraSubsetSum(avail, 0, target, nil)
}

// basraSubsetSum は avail[start:] から合計 target になる部分集合を再帰的に探す。
func (g *Basra) basraSubsetSum(avail []int, start, target int, acc []int) []int {
	if target == 0 {
		if len(acc) >= 2 {
			return append([]int(nil), acc...)
		}
		return nil
	}
	for i := start; i < len(avail); i++ {
		v := basraCardValue(g.state.tableCards[avail[i]])
		if v > target {
			continue
		}
		if res := g.basraSubsetSum(avail, i+1, target-v, append(acc, avail[i])); res != nil {
			return res
		}
	}
	return nil
}

// --- Play ---

// PlayerPlay は人間プレイヤーが手札 handIdx を出す。
// tableIdxs が指定されていればその場札を捕獲する (合法性を検証)。tableIdxs が空の場合、
// ジャックは自動的に場を一掃し、それ以外はトレイル (場に置く)。
func (g *Basra) PlayerPlay(handIdx int, tableIdxs []int) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != BasraPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	if !g.players[g.state.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.state.currentTurn]
	if handIdx < 0 || handIdx >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	card := player.GetCard(handIdx)
	// ジャック以外で場札を指定した場合は選択を検証する。
	if len(tableIdxs) > 0 && !basraIsJack(card) {
		if err := g.basraValidateSelection(card, tableIdxs); err != nil {
			return err
		}
	}
	g.applyPlay(g.state.currentTurn, handIdx, tableIdxs)
	return nil
}

// CpuPlay は CPU のターンを 1 回進める。
func (g *Basra) CpuPlay() {
	if g.state.gameEndFlag || g.state.phase != BasraPhasePlay {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() || p.GetCardsSize() == 0 {
		return
	}
	handIdx, tableIdxs := g.chooseCpuPlay(g.state.currentTurn)
	g.applyPlay(g.state.currentTurn, handIdx, tableIdxs)
}

// applyPlay は共通のプレイ処理: 手札を出して捕獲/トレイルを解決し、手番を進める。
// tableIdxs が nil/空の場合、ジャックは自動一掃、数札/絵札はトレイル。tableIdxs が
// 指定済みならその場札を捕獲する (呼び出し側で検証済み前提)。
func (g *Basra) applyPlay(playerIdx, handIdx int, tableIdxs []int) {
	player := g.players[playerIdx]
	card := player.RemoveCard(handIdx)
	if card == nil {
		return
	}

	// ジャックは常に場のジャック以外を一掃 (選択は無視)。
	if basraIsJack(card) {
		tableIdxs = g.basraFindCaptures(card)
	}

	if len(tableIdxs) == 0 {
		// トレイル: 場に置く。
		g.state.tableCards = append(g.state.tableCards, card)
		g.appendLog(playerIdx, "trail", fmt.Sprintf("%s trails %s", g.playerName(playerIdx), cardStr(card)),
			[]*Card{card})
		g.advanceTurn()
		return
	}

	tableBefore := len(g.state.tableCards)
	captured := make([]*Card, 0, len(tableIdxs)+1)
	captured = append(captured, card)
	for _, idx := range tableIdxs {
		captured = append(captured, g.state.tableCards[idx])
	}
	g.removeTableCardsByIndex(tableIdxs)
	player.AddCaptured(captured)
	g.state.lastCaptureIdx = playerIdx

	// バスラ判定: ジャック以外の 1 枚で場を空にした場合。
	isBasra := !basraIsJack(card) && len(g.state.tableCards) == 0 && tableBefore > 0
	if isBasra {
		player.IncrementBasra()
		g.appendLog(playerIdx, "basra",
			fmt.Sprintf("%s scores a Basra! captured %d card(s)", g.playerName(playerIdx), len(captured)-1),
			captured)
	} else if basraIsJack(card) {
		g.appendLog(playerIdx, "sweep",
			fmt.Sprintf("%s sweeps %d card(s) with a Jack", g.playerName(playerIdx), len(captured)-1),
			captured)
	} else {
		g.appendLog(playerIdx, "capture",
			fmt.Sprintf("%s captures %d card(s)", g.playerName(playerIdx), len(captured)-1),
			captured)
	}
	g.advanceTurn()
}

// advanceTurn は手番を次に進め、必要なら配り直し・終局処理を行う。
func (g *Basra) advanceTurn() {
	g.state.currentTurn = (g.state.currentTurn + 1) % len(g.players)
	if !g.allHandsEmpty() {
		return
	}
	if g.trumpCards.GetRemainingCount() > 0 {
		g.dealHands()
		g.sortHumanHand()
		return
	}
	g.finishGame()
}

// finishGame は終局処理: 残りの場札を最後の捕獲者へ渡し、最終得点を確定する。
// scored フラグで二重確定を防ぐ。lastDealDetail は本メソッドで即座に構築するため、
// フロントエンドは gameEnd フェーズ入場時点で結果画面を描画できる。
func (g *Basra) finishGame() {
	if g.state.scored {
		return
	}
	g.state.scored = true
	if g.state.lastCaptureIdx >= 0 && len(g.state.tableCards) > 0 {
		leftover := append([]*Card(nil), g.state.tableCards...)
		g.players[g.state.lastCaptureIdx].AddCaptured(leftover)
		g.appendLog(g.state.lastCaptureIdx, "lastTake",
			fmt.Sprintf("last-take: %d card(s)", len(leftover)), leftover)
	}
	g.state.tableCards = nil

	detail := g.calcFinalScore()
	g.state.lastDealDetail = detail
	for i, p := range g.players {
		p.SetScore(detail.Gained[i])
	}
	// 勝者判定 (最高点, 同点なら複数)。
	maxScore := 0
	for _, p := range g.players {
		if p.GetScore() > maxScore {
			maxScore = p.GetScore()
		}
	}
	winners := make([]int, 0)
	for i, p := range g.players {
		if p.GetScore() == maxScore {
			winners = append(winners, i)
		}
	}
	g.state.winners = winners
	g.state.gameEndFlag = true
	g.state.phase = BasraPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended (top score %d)", maxScore), nil)
}

// calcFinalScore は各プレイヤーの最終得点内訳を計算する。
func (g *Basra) calcFinalScore() *BasraScoreDetail {
	det := &BasraScoreDetail{
		Cards:            make(map[int]int),
		Aces:             make(map[int]int),
		Basras:           make(map[int]int),
		Gained:           make(map[int]int),
		HasSevenDiamonds: -1,
		HasTenDiamonds:   -1,
		MostCards:        -1,
	}
	for i, p := range g.players {
		det.Cards[i] = p.CapturedCount()
		det.Basras[i] = p.GetBasraCount()
		for _, c := range p.GetCapturedCards() {
			if c.GetValue() == 1 {
				det.Aces[i]++
			}
			if c.GetValue() == 7 && c.GetDesign() == CardDesignDiamond {
				det.HasSevenDiamonds = i
			}
			if c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond {
				det.HasTenDiamonds = i
			}
		}
	}
	det.MostCards = basraUniqueMaxIndex(det.Cards)
	for i := range g.players {
		score := 0
		if i == det.MostCards {
			score += BasraScoreMostCards
		}
		if i == det.HasSevenDiamonds {
			score += BasraScoreSevenDiamonds
		}
		if i == det.HasTenDiamonds {
			score += BasraScoreTenDiamonds
		}
		score += det.Aces[i] * BasraScoreAce
		score += det.Basras[i] * BasraScoreBasra
		det.Gained[i] = score
	}
	return det
}

// basraUniqueMaxIndex はマップ中で最大値かつ単独のキーを返す。同点または空 (最大 0) なら -1。
func basraUniqueMaxIndex(m map[int]int) int {
	best := -1
	bestVal := 0
	tie := false
	for k := 0; k < len(m); k++ {
		v := m[k]
		if best == -1 || v > bestVal {
			best = k
			bestVal = v
			tie = false
		} else if v == bestVal {
			tie = true
		}
	}
	if tie || bestVal == 0 {
		return -1
	}
	return best
}

// removeTableCardsByIndex は降順に並び替えてから tableCards を削除する。
func (g *Basra) removeTableCardsByIndex(idxs []int) {
	if len(idxs) == 0 {
		return
	}
	sorted := append([]int(nil), idxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, idx := range sorted {
		if idx >= 0 && idx < len(g.state.tableCards) {
			g.state.tableCards = append(g.state.tableCards[:idx], g.state.tableCards[idx+1:]...)
		}
	}
}

// --- CPU AI ---

// chooseCpuPlay は CPU の手番で出す手札インデックスと捕獲対象を選ぶ。
//   - Easy   : 合法手からランダム (捕獲できれば捕獲、無ければトレイル)。
//   - Normal : 捕獲を優先し、捕獲枚数が最大の札を選ぶ。
//   - Hard   : バスラ / 高得点札の捕獲を最優先し、無ければ Normal 相当。
func (g *Basra) chooseCpuPlay(playerIdx int) (int, []int) {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0, nil
	}
	if g.config.CpuDifficulty == BasraCpuDifficultyEasy {
		idx := rand.Intn(size)
		return idx, g.basraFindCaptures(player.GetCard(idx))
	}

	bestIdx := -1
	var bestCaps []int
	bestScore := -1
	for i := 0; i < size; i++ {
		card := player.GetCard(i)
		caps := g.basraFindCaptures(card)
		s := g.cpuCaptureScore(card, caps)
		if s > bestScore {
			bestScore = s
			bestIdx = i
			bestCaps = caps
		}
	}
	if bestIdx < 0 {
		return 0, nil
	}
	// 捕獲できないなら最も価値の低い札を捨てる (トレイル)。
	if len(bestCaps) == 0 {
		return g.lowestValueCardIdx(player), nil
	}
	return bestIdx, bestCaps
}

// cpuCaptureScore は (card, caps) の望ましさを評価する。捕獲枚数を基本点とし、バスラ・
// 高得点札 (7♦/10♦/A) の捕獲を加点する。Hard はバスラをさらに重視する。
func (g *Basra) cpuCaptureScore(card *Card, caps []int) int {
	if len(caps) == 0 {
		return 0
	}
	score := len(caps)
	// バスラ (ジャック以外で場を空にする)。
	if !basraIsJack(card) && len(caps) == len(g.state.tableCards) && len(g.state.tableCards) > 0 {
		if g.config.CpuDifficulty == BasraCpuDifficultyHard {
			score += 20
		} else {
			score += 8
		}
	}
	for _, idx := range caps {
		if idx < 0 || idx >= len(g.state.tableCards) {
			continue
		}
		c := g.state.tableCards[idx]
		if c.GetValue() == 1 {
			score += 2
		}
		if c.GetValue() == 7 && c.GetDesign() == CardDesignDiamond {
			score += 4
		}
		if c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond {
			score += 3
		}
	}
	return score
}

// lowestValueCardIdx は捨てるのに最も無難な (得点価値の低い) 手札インデックスを返す。
// ジャックは強力な捕獲札なので温存する。
func (g *Basra) lowestValueCardIdx(player *BasraPlayer) int {
	best := 0
	bestPts := 1 << 30
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		pts := 0
		if c.GetValue() == 1 {
			pts += 2
		}
		if basraIsJack(c) {
			pts += 5 // ジャックは温存
		}
		if c.GetValue() == 7 && c.GetDesign() == CardDesignDiamond {
			pts += 4
		}
		if c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond {
			pts += 3
		}
		if pts < bestPts {
			bestPts = pts
			best = i
		}
	}
	return best
}

// --- Hint ---

// GetHint は人間プレイヤーの手番における推奨プレイを返す。
func (g *Basra) GetHint() *BasraHint {
	if g.state.gameEndFlag || g.state.phase != BasraPhasePlay {
		return nil
	}
	human := findHumanIdx(g.players)
	if human < 0 || g.state.currentTurn != human {
		return nil
	}
	player := g.players[human]
	if player.GetCardsSize() == 0 {
		return nil
	}
	bestIdx := -1
	var bestCaps []int
	bestScore := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		caps := g.basraFindCaptures(card)
		s := g.cpuCaptureScore(card, caps)
		if s > bestScore {
			bestScore = s
			bestIdx = i
			bestCaps = caps
		}
	}
	if bestIdx < 0 {
		return nil
	}
	card := player.GetCard(bestIdx)
	if len(bestCaps) == 0 {
		return &BasraHint{CardIndices: []int{g.lowestValueCardIdx(player)}, Reason: "trail_low"}
	}
	reason := "capture"
	if basraIsJack(card) {
		reason = "jack_sweep"
	} else if len(bestCaps) == len(g.state.tableCards) {
		reason = "basra_sweep"
	}
	return &BasraHint{CardIndices: []int{bestIdx}, TableIndices: bestCaps, Reason: reason}
}

// --- ヘルパー ---

// sortHumanHand は人間の手札を表示用にスート→ランク順で並べ替える。
func (g *Basra) sortHumanHand() {
	for _, p := range g.players {
		if !p.GetIsHuman() {
			continue
		}
		cards := make([]*Card, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			cards[i] = p.GetCard(i)
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

func (g *Basra) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

func (g *Basra) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn は現在の手番が人間かどうかを返す。
func (g *Basra) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	return g.players[g.state.currentTurn].GetIsHuman()
}

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (g *Basra) GetCurrentTurn() int { return g.state.currentTurn }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *Basra) SetCurrentTurn(idx int) { g.state.currentTurn = idx }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Basra) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *Basra) GetPhase() BasraPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Basra) SetPhase(p BasraPhase) { g.state.phase = p }

// GetTableCards は場の札を返す。
func (g *Basra) GetTableCards() []*Card { return g.state.tableCards }

// SetTableCards は場の札を設定する (テスト用)。
func (g *Basra) SetTableCards(cards []*Card) { g.state.tableCards = cards }

// GetLastCaptureIdx は最後に捕獲したプレイヤーを返す (-1 = なし)。
func (g *Basra) GetLastCaptureIdx() int { return g.state.lastCaptureIdx }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Basra) GetPlayer(i int) *BasraPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetPlayerCnt はプレイヤー数を返す。
func (g *Basra) GetPlayerCnt() int { return len(g.players) }

// GetRemainingDeck は山札の残り枚数を返す。
func (g *Basra) GetRemainingDeck() int { return g.trumpCards.GetRemainingCount() }

// GetRoundNumber はこれまでに配布されたパック数 (配り直し回数) を返す。
func (g *Basra) GetRoundNumber() int { return g.state.packsDealt }

// GetConfig はローカルルール設定を返す。
func (g *Basra) GetConfig() BasraConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Basra) SetConfig(cfg BasraConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *Basra) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// GetWinners はゲーム終了時の勝者シートのリストを返す (同点なら複数)。
func (g *Basra) GetWinners() []int { return g.state.winners }

// GetLastDealDetail は直前ゲームの得点内訳を返す (nil の場合もある)。
func (g *Basra) GetLastDealDetail() *BasraScoreDetail { return g.state.lastDealDetail }

// GetPlayableIndices はプレイフェーズで人間がプレイできる手札インデックス (全札) を返す。
// バスラでは常に任意の手札を出せる (捕獲できなければトレイル)。
func (g *Basra) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != BasraPhasePlay {
		return nil
	}
	p := g.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		out = append(out, i)
	}
	return out
}

// GetCaptureOptions は playerIdx の各手札が捕獲できる場札インデックスを返す
// (キー = 手札インデックス)。捕獲できない手札は含めない。UI 補助用。
func (g *Basra) GetCaptureOptions(playerIdx int) map[int][]int {
	out := make(map[int][]int)
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != BasraPhasePlay {
		return out
	}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		caps := g.basraFindCaptures(p.GetCard(i))
		if len(caps) > 0 {
			out[i] = caps
		}
	}
	return out
}

// --- JSON Serialization ---

// basraJSON is the JSON wire format for Basra.
type basraJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	Players        []*BasraPlayer    `json:"pl"`
	Config         BasraConfig       `json:"cf"`
	Phase          BasraPhase        `json:"ph"`
	CurrentTurn    int               `json:"ct"`
	TableCards     []*Card           `json:"tb"`
	LastCaptureIdx int               `json:"lc"`
	PacksDealt     int               `json:"pd"`
	GameEndFlag    bool              `json:"ge"`
	Scored         bool              `json:"sc"`
	Winners        []int             `json:"wn"`
	LastDealDetail *BasraScoreDetail `json:"ld"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Basra) MarshalJSON() ([]byte, error) {
	return json.Marshal(basraJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		Config:         g.config,
		Phase:          g.state.phase,
		CurrentTurn:    g.state.currentTurn,
		TableCards:     g.state.tableCards,
		LastCaptureIdx: g.state.lastCaptureIdx,
		PacksDealt:     g.state.packsDealt,
		GameEndFlag:    g.state.gameEndFlag,
		Scored:         g.state.scored,
		Winners:        g.state.winners,
		LastDealDetail: g.state.lastDealDetail,
		ActionLog:      g.state.actionLog,
	})
}

// basraMaxSliceLen caps slice sizes during deserialisation to prevent excessive
// memory allocation from malformed input.
const basraMaxSliceLen = 1000

// basraValidPhase は有効なフェーズかどうか。
func basraValidPhase(p BasraPhase) bool {
	return p == BasraPhasePlay || p == BasraPhaseGameEnd
}

// basraValidateCards は復元したカードスライスに nil が無いか検証する。
func basraValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("basra: nil card in state")
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う。
func (g *Basra) UnmarshalJSON(data []byte) error {
	var j basraJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > basraMaxSliceLen || len(j.TableCards) > basraMaxSliceLen ||
		len(j.ActionLog) > basraMaxSliceLen || len(j.Winners) > basraMaxSliceLen {
		return fmt.Errorf("basra: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("basra: invalid config: %w", err)
	}
	if j.TrumpCards == nil {
		return fmt.Errorf("basra: missing trump cards in state")
	}
	if len(j.Players) != BasraPlayerCnt {
		return fmt.Errorf("basra: invalid player count %d, expected %d", len(j.Players), BasraPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("basra: nil player in state")
		}
	}
	if !basraValidPhase(j.Phase) {
		return fmt.Errorf("basra: invalid phase %d", j.Phase)
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= len(j.Players) {
		return fmt.Errorf("basra: current turn out of range")
	}
	if j.LastCaptureIdx < -1 || j.LastCaptureIdx >= len(j.Players) {
		return fmt.Errorf("basra: last capture index out of range")
	}
	if err := basraValidateCards(j.TableCards); err != nil {
		return err
	}
	for _, w := range j.Winners {
		if w < 0 || w >= len(j.Players) {
			return fmt.Errorf("basra: winner index out of range")
		}
	}

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.state = basraState{
		phase:          j.Phase,
		currentTurn:    j.CurrentTurn,
		tableCards:     j.TableCards,
		lastCaptureIdx: j.LastCaptureIdx,
		packsDealt:     j.PacksDealt,
		gameEndFlag:    j.GameEndFlag,
		scored:         j.Scored,
		winners:        j.Winners,
		lastDealDetail: j.LastDealDetail,
		actionLogBase:  actionLogBase{actionLog: j.ActionLog},
	}
	if g.state.tableCards == nil {
		g.state.tableCards = make([]*Card, 0)
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// basraScoreDetailJSON is the JSON wire format for BasraScoreDetail.
type basraScoreDetailJSON struct {
	Cards            map[int]int `json:"cd"`
	Aces             map[int]int `json:"ac"`
	Basras           map[int]int `json:"br"`
	HasSevenDiamonds int         `json:"s7"`
	HasTenDiamonds   int         `json:"t0"`
	MostCards        int         `json:"mc"`
	Gained           map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *BasraScoreDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(basraScoreDetailJSON{
		Cards:            d.Cards,
		Aces:             d.Aces,
		Basras:           d.Basras,
		HasSevenDiamonds: d.HasSevenDiamonds,
		HasTenDiamonds:   d.HasTenDiamonds,
		MostCards:        d.MostCards,
		Gained:           d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *BasraScoreDetail) UnmarshalJSON(data []byte) error {
	var j basraScoreDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Cards = j.Cards
	d.Aces = j.Aces
	d.Basras = j.Basras
	d.HasSevenDiamonds = j.HasSevenDiamonds
	d.HasTenDiamonds = j.HasTenDiamonds
	d.MostCards = j.MostCards
	d.Gained = j.Gained
	return nil
}
