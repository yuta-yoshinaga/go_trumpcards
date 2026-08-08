//go:build !js || !wasm || extra3

// Package domain タブラネット (Tablanet / Tablić) のドメインモデル。
//
// Tablanet はバルカン半島 (セルビア・マケドニア等) で親しまれる漁 (フィッシング/キャプチャ)
// 系カードゲーム。標準 52 枚デッキを 4 人 (seat 0 が人間 + CPU 3 人, 個人戦) でプレイする。
// Cassino / Bastra に近く、同ランク捕獲・合計値捕獲・ジャックの一掃 (スイープ)・タブラ
// (Tabla) ボーナスを併せ持つ。
//
// # ルール概要
//
// 各プレイヤーへ TablanetHandSize (4) 枚を配り、場に TablanetInitialTableSize (4) 枚を表向きで
// 置く。手番では手札から 1 枚を出し、次のいずれかで場札を捕獲する。
//
//   - 出した数札 (A〜10) は、場札のうち「同ランク」または「合計値が出した札の値に等しい
//     組」をすべて捕獲できる (合計捕獲は複数の組を同時に取れる)。A は 1。
//   - Q / K は同ランク捕獲のみ (合計には参加しない)。
//   - ジャック (値 11) は場のジャック以外の全札を一掃 (スイープ) して捕獲する。場の別の
//     ジャックは残す。
//   - どの札も捕獲しなかった場合、その札は場に置かれる (トレイル)。
//
// # タブラ (Tabla ボーナス)
//
// ジャック以外の 1 枚で場札を「すべて」捕獲し場が空になった場合、それは "Tabla" となり
// TablanetScoreTabla 点のボーナスを得る (ジャックのスイープはタブラに含めない)。
//
// # 配り直し / 終局
//
// 全員の手札が尽きたら、山札から新たに 4 枚ずつ配る (場札は補充しない)。これを山札が
// 尽きるまで繰り返す。最後に場へ残った札は、最後に捕獲したプレイヤーが取る。
//
// # 得点 (calcFinalScore) — タブラネットの伝統的な配点
//
//   - 捕獲枚数が最多 (単独) のプレイヤー: +TablanetScoreMostCards (3)
//   - A を捕獲: 1 枚につき +TablanetScoreAce (1)
//   - J (ジャック) を捕獲: 1 枚につき +TablanetScoreJack (1)
//   - 10♦ を捕獲: +TablanetScoreTenDiamonds (2)
//   - 2♣ を捕獲: +TablanetScoreTwoClubs (1)
//   - タブラ (Tabla): 1 回につき +TablanetScoreTabla (1)
//
// (1 ディール当たりの基礎点 = A×4 + J×4 + 10♦ + 2♣ + 最多札 = 14 点。7♦ は Bastra 系の
// ボーナスでありタブラネットでは採点しない。)
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

// TablanetPlayerCnt はタブラネットのプレイヤー数 (固定 4, 個人戦)。
const TablanetPlayerCnt = 4

// TablanetHandSize は 1 回の配布で各プレイヤーへ配る手札枚数。
const TablanetHandSize = 4

// TablanetInitialTableSize はゲーム開始時に場へ置くカード枚数。
const TablanetInitialTableSize = 4

// TablanetJackValue はジャックのカード値 (スイープ札)。
const TablanetJackValue = 11

// TablanetPhase はゲームフェーズ。
type TablanetPhase int

// Tablanet のフェーズ定数
const (
	// TablanetPhasePlay プレイ中 (カードを場へ出す)
	TablanetPhasePlay TablanetPhase = 0
	// TablanetPhaseGameEnd ゲーム終了 (山札を配り切り、最終得点を確定)
	TablanetPhaseGameEnd TablanetPhase = 1
)

// Tablanet の得点定数 (伝統的なタブラネット/タブリッチの配点)
const (
	// TablanetScoreMostCards 捕獲枚数最多 (単独) のボーナス
	TablanetScoreMostCards = 3
	// TablanetScoreAce A 1 枚のボーナス
	TablanetScoreAce = 1
	// TablanetScoreJack J (ジャック) 1 枚のボーナス
	TablanetScoreJack = 1
	// TablanetScoreTenDiamonds 10♦ のボーナス
	TablanetScoreTenDiamonds = 2
	// TablanetScoreTwoClubs 2♣ のボーナス
	TablanetScoreTwoClubs = 1
	// TablanetScoreTabla タブラ (単独札での場一掃) 1 回のボーナス
	TablanetScoreTabla = 1
)

// TablanetScoreDetail は 1 ゲームの得点内訳。
type TablanetScoreDetail struct {
	Cards          map[int]int // プレイヤー別捕獲枚数
	Aces           map[int]int // プレイヤー別エース数
	Jacks          map[int]int // プレイヤー別ジャック数
	Tablas         map[int]int // プレイヤー別タブラ数
	HasTenDiamonds int         // 10♦ を捕獲したプレイヤー (-1 = なし)
	HasTwoClubs    int         // 2♣ を捕獲したプレイヤー (-1 = なし)
	MostCards      int         // 捕獲枚数最多の単独プレイヤー (-1 = 同点/なし)
	Gained         map[int]int // プレイヤー別の得点
}

// TablanetHint はヒント情報。
type TablanetHint struct {
	CardIndices  []int  // 推奨手札インデックス
	TableIndices []int  // 推奨捕獲対象の場札インデックス
	Reason       string // ヒント理由キー
}

// tablanetState はゲーム進行状態。
type tablanetState struct {
	phase          TablanetPhase
	currentTurn    int     // 現在の手番プレイヤー
	tableCards     []*Card // 場の札
	lastCaptureIdx int     // 最後に捕獲したプレイヤー (-1 = なし)
	packsDealt     int     // これまでに配ったパック数 (1 回の配布 = 4 枚/人)
	gameEndFlag    bool
	scored         bool // 最終得点を確定済みか (二重確定防止)
	winners        []int
	lastDealDetail *TablanetScoreDetail
	actionLogBase
}

// Tablanet はタブラネットゲームの状態を保持する集約ルート。
type Tablanet struct {
	trumpCards *TrumpCards
	players    []*TablanetPlayer
	config     TablanetConfig
	state      tablanetState
}

// NewTablanet はコンストラクタ。
func NewTablanet(trumpCards *TrumpCards, players []*TablanetPlayer, config TablanetConfig) *Tablanet {
	return &Tablanet{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: tablanetState{
			phase:          TablanetPhasePlay,
			lastCaptureIdx: -1,
		},
	}
}

// NewDefaultTablanet は標準の 4 人構成 (1 human + 3 CPU) と DefaultTablanetConfig で Tablanet を
// 生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultTablanet() *Tablanet {
	players := make([]*TablanetPlayer, TablanetPlayerCnt)
	players[0] = NewTablanetPlayer(true)
	for i := 1; i < TablanetPlayerCnt; i++ {
		players[i] = NewTablanetPlayer(false)
	}
	return NewTablanet(newTablanetDeck(), players, DefaultTablanetConfig())
}

// newTablanetDeck はタブラネット用 52 枚デッキを生成する。NewTrumpCards はビルドタグ無しの
// TrumpCards.go にあり extra ワーカーからも到達可能。
func newTablanetDeck() *TrumpCards {
	return NewTrumpCards(0)
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。
func (g *Tablanet) Reset() {
	for _, p := range g.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
		p.ResetTabla()
		p.ResetScore()
	}
	g.trumpCards = newTablanetDeck()
	g.trumpCards.Shuffle()
	g.state = tablanetState{
		phase:          TablanetPhasePlay,
		currentTurn:    0,
		lastCaptureIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.dealHands()
	g.dealInitialTable()
	g.sortHumanHand()
}

// NextRound はタブラネットでは新規ゲーム開始 (Reset) と同義。1 ゲームは山札を配り切る 1
// セッションで完結するため、終局後の続行は新しいゲームの開始として扱う。
func (g *Tablanet) NextRound() {
	g.Reset()
}

// dealHands は各プレイヤーへ TablanetHandSize 枚配る。山札が尽きたら途中で終わる。
func (g *Tablanet) dealHands() {
	for k := 0; k < TablanetHandSize; k++ {
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

// dealInitialTable はゲーム開始時に TablanetInitialTableSize 枚を場へ置く。
func (g *Tablanet) dealInitialTable() {
	for i := 0; i < TablanetInitialTableSize; i++ {
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
func (g *Tablanet) allHandsEmpty() bool {
	for _, p := range g.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// --- 捕獲ロジック (インライン) ---

// tablanetCardValue はカードのキャプチャ用の値を返す (A=1 … K=13)。
func tablanetCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue()
}

// tablanetIsJack はジャック (値 11, スイープ札) かどうか。
func tablanetIsJack(c *Card) bool {
	return c != nil && c.GetValue() == TablanetJackValue
}

// tablanetIsFace は絵札 (J/Q/K, 値 11-13) かどうか。
func tablanetIsFace(c *Card) bool {
	return c != nil && c.GetValue() >= 11
}

// tablanetCanPartition は values を「各グループの合計が target に等しい」ように過不足なく
// 分割できるかを返す。単独の同ランク札 (値 == target) も 1 枚のグループとして扱えるため、
// 同ランク捕獲と合計捕獲を統一的に検証できる。target を超える値が含まれる場合は不可。
func tablanetCanPartition(values []int, target int) bool {
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
	return tablanetFillBuckets(vs, 0, remaining)
}

// tablanetFillBuckets は k-partition のバックトラッキング補助。
func tablanetFillBuckets(vs []int, idx int, remaining []int) bool {
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
		if tablanetFillBuckets(vs, idx+1, remaining) {
			return true
		}
		remaining[b] += v
	}
	return false
}

// tablanetValidateSelection は playedCard で tableIdxs を捕獲する選択が合法かを検証する。
// ジャックは呼び出し側で別処理するため、ここでは数札 / Q / K のみを扱う。
func (g *Tablanet) tablanetValidateSelection(playedCard *Card, tableIdxs []int) error {
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
	pv := tablanetCardValue(playedCard)
	if tablanetIsFace(playedCard) {
		// Q / K は同ランク捕獲のみ。
		for _, c := range selected {
			if tablanetCardValue(c) != pv {
				return NewDomainError(ErrInvalidPlay, "face cards capture only by matching rank")
			}
		}
		return nil
	}
	// 数札 (A〜10): 選択は target=pv の組に過不足なく分割できること。
	vals := make([]int, len(selected))
	for i, c := range selected {
		vals[i] = tablanetCardValue(c)
	}
	if !tablanetCanPartition(vals, pv) {
		return NewDomainError(ErrInvalidPlay, "selected table cards do not sum/match the played card")
	}
	return nil
}

// tablanetFindCaptures は playedCard が捕獲できる場札インデックスの最大集合を返す。
// CPU と、ジャック/自動捕獲経路で使用する。捕獲不能なら空スライスを返す。
func (g *Tablanet) tablanetFindCaptures(playedCard *Card) []int {
	if tablanetIsJack(playedCard) {
		// ジャック: 場のジャック以外を全て捕獲。
		var out []int
		for i, c := range g.state.tableCards {
			if !tablanetIsJack(c) {
				out = append(out, i)
			}
		}
		return out
	}
	pv := tablanetCardValue(playedCard)
	if tablanetIsFace(playedCard) {
		// Q / K: 同ランクのみ。
		var out []int
		for i, c := range g.state.tableCards {
			if tablanetCardValue(c) == pv {
				out = append(out, i)
			}
		}
		return out
	}
	// 数札: グループ (同ランク単独 + 合計) を貪欲に繰り返し抽出する。
	used := make([]bool, len(g.state.tableCards))
	captured := make([]int, 0)
	for {
		group := g.tablanetFindOneGroup(pv, used)
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

// tablanetFindOneGroup は未使用の数札の中から、合計が target になる 1 グループ (同ランク単独を
// 含む) を探して返す。見つからなければ nil。絵札 (値 11-13) は数札グループに含めない。
func (g *Tablanet) tablanetFindOneGroup(target int, used []bool) []int {
	// まず同ランク単独 (値 == target) を優先。
	for i, c := range g.state.tableCards {
		if used[i] || tablanetIsFace(c) {
			continue
		}
		if tablanetCardValue(c) == target {
			return []int{i}
		}
	}
	// 次に 2 枚以上の合計組を探索する。
	var avail []int
	for i, c := range g.state.tableCards {
		if used[i] || tablanetIsFace(c) {
			continue
		}
		if tablanetCardValue(c) < target {
			avail = append(avail, i)
		}
	}
	return g.tablanetSubsetSum(avail, 0, target, nil)
}

// tablanetSubsetSum は avail[start:] から合計 target になる部分集合を再帰的に探す。
func (g *Tablanet) tablanetSubsetSum(avail []int, start, target int, acc []int) []int {
	if target == 0 {
		if len(acc) >= 2 {
			return append([]int(nil), acc...)
		}
		return nil
	}
	for i := start; i < len(avail); i++ {
		v := tablanetCardValue(g.state.tableCards[avail[i]])
		if v > target {
			continue
		}
		if res := g.tablanetSubsetSum(avail, i+1, target-v, append(acc, avail[i])); res != nil {
			return res
		}
	}
	return nil
}

// --- Play ---

// PlayerPlay は人間プレイヤーが手札 handIdx を出す。
// tableIdxs が指定されていればその場札を捕獲する (合法性を検証)。tableIdxs が空の場合、
// ジャックは自動的に場を一掃し、それ以外はトレイル (場に置く)。
func (g *Tablanet) PlayerPlay(handIdx int, tableIdxs []int) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != TablanetPhasePlay {
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
	if len(tableIdxs) > 0 && !tablanetIsJack(card) {
		if err := g.tablanetValidateSelection(card, tableIdxs); err != nil {
			return err
		}
	}
	g.applyPlay(g.state.currentTurn, handIdx, tableIdxs)
	return nil
}

// CpuPlay は CPU のターンを 1 回進める。
func (g *Tablanet) CpuPlay() {
	if g.state.gameEndFlag || g.state.phase != TablanetPhasePlay {
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
func (g *Tablanet) applyPlay(playerIdx, handIdx int, tableIdxs []int) {
	player := g.players[playerIdx]
	card := player.RemoveCard(handIdx)
	if card == nil {
		return
	}

	// ジャックは常に場のジャック以外を一掃 (選択は無視)。
	if tablanetIsJack(card) {
		tableIdxs = g.tablanetFindCaptures(card)
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

	// タブラ判定: ジャック以外の 1 枚で場を空にした場合。
	isTabla := !tablanetIsJack(card) && len(g.state.tableCards) == 0 && tableBefore > 0
	if isTabla {
		player.IncrementTabla()
		g.appendLog(playerIdx, "tabla",
			fmt.Sprintf("%s scores a Tabla! captured %d card(s)", g.playerName(playerIdx), len(captured)-1),
			captured)
	} else if tablanetIsJack(card) {
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
func (g *Tablanet) advanceTurn() {
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
func (g *Tablanet) finishGame() {
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
	g.state.phase = TablanetPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended (top score %d)", maxScore), nil)
}

// calcFinalScore は各プレイヤーの最終得点内訳を計算する。
func (g *Tablanet) calcFinalScore() *TablanetScoreDetail {
	det := &TablanetScoreDetail{
		Cards:          make(map[int]int),
		Aces:           make(map[int]int),
		Jacks:          make(map[int]int),
		Tablas:         make(map[int]int),
		Gained:         make(map[int]int),
		HasTenDiamonds: -1,
		HasTwoClubs:    -1,
		MostCards:      -1,
	}
	for i, p := range g.players {
		det.Cards[i] = p.CapturedCount()
		det.Tablas[i] = p.GetTablaCount()
		for _, c := range p.GetCapturedCards() {
			if c.GetValue() == 1 {
				det.Aces[i]++
			}
			if c.GetValue() == TablanetJackValue {
				det.Jacks[i]++
			}
			if c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond {
				det.HasTenDiamonds = i
			}
			if c.GetValue() == 2 && c.GetDesign() == CardDesignClover {
				det.HasTwoClubs = i
			}
		}
	}
	det.MostCards = tablanetUniqueMaxIndex(det.Cards)
	for i := range g.players {
		score := 0
		if i == det.MostCards {
			score += TablanetScoreMostCards
		}
		if i == det.HasTenDiamonds {
			score += TablanetScoreTenDiamonds
		}
		if i == det.HasTwoClubs {
			score += TablanetScoreTwoClubs
		}
		score += det.Aces[i] * TablanetScoreAce
		score += det.Jacks[i] * TablanetScoreJack
		score += det.Tablas[i] * TablanetScoreTabla
		det.Gained[i] = score
	}
	return det
}

// tablanetUniqueMaxIndex はマップ中で最大値かつ単独のキーを返す。同点または空 (最大 0) なら -1。
func tablanetUniqueMaxIndex(m map[int]int) int {
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
func (g *Tablanet) removeTableCardsByIndex(idxs []int) {
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
//   - Hard   : タブラネット / 高得点札の捕獲を最優先し、無ければ Normal 相当。
func (g *Tablanet) chooseCpuPlay(playerIdx int) (int, []int) {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0, nil
	}
	if g.config.CpuDifficulty == TablanetCpuDifficultyEasy {
		idx := rand.Intn(size)
		return idx, g.tablanetFindCaptures(player.GetCard(idx))
	}

	bestIdx := -1
	var bestCaps []int
	bestScore := -1
	for i := 0; i < size; i++ {
		card := player.GetCard(i)
		caps := g.tablanetFindCaptures(card)
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

// cpuCaptureScore は (card, caps) の望ましさを評価する。捕獲枚数を基本点とし、タブラ・
// 高得点札 (A/J/10♦/2♣) の捕獲を加点する。Hard はタブラをさらに重視する。
func (g *Tablanet) cpuCaptureScore(card *Card, caps []int) int {
	if len(caps) == 0 {
		return 0
	}
	score := len(caps)
	// タブラ (ジャック以外で場を空にする)。
	if !tablanetIsJack(card) && len(caps) == len(g.state.tableCards) && len(g.state.tableCards) > 0 {
		if g.config.CpuDifficulty == TablanetCpuDifficultyHard {
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
			score += 2 // Ace
		}
		if c.GetValue() == TablanetJackValue {
			score += 2 // Jack (each worth a point)
		}
		if c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond {
			score += 3 // 10♦
		}
		if c.GetValue() == 2 && c.GetDesign() == CardDesignClover {
			score += 2 // 2♣
		}
	}
	return score
}

// lowestValueCardIdx は捨てるのに最も無難な (得点価値の低い) 手札インデックスを返す。
// ジャックは強力な捕獲札なので温存する。
func (g *Tablanet) lowestValueCardIdx(player *TablanetPlayer) int {
	best := 0
	bestPts := 1 << 30
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		pts := 0
		if c.GetValue() == 1 {
			pts += 2
		}
		if tablanetIsJack(c) {
			pts += 5 // ジャックは温存 (強力な捕獲札 + 各 1 点)
		}
		if c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond {
			pts += 3 // 10♦
		}
		if c.GetValue() == 2 && c.GetDesign() == CardDesignClover {
			pts += 2 // 2♣
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
func (g *Tablanet) GetHint() *TablanetHint {
	if g.state.gameEndFlag || g.state.phase != TablanetPhasePlay {
		return nil
	}
	human := g.findHumanIdx()
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
		caps := g.tablanetFindCaptures(card)
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
		return &TablanetHint{CardIndices: []int{g.lowestValueCardIdx(player)}, Reason: "trail_low"}
	}
	reason := "capture"
	if tablanetIsJack(card) {
		reason = "jack_sweep"
	} else if len(bestCaps) == len(g.state.tableCards) {
		reason = "tabla_sweep"
	}
	return &TablanetHint{CardIndices: []int{bestIdx}, TableIndices: bestCaps, Reason: reason}
}

// --- ヘルパー ---

func (g *Tablanet) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// sortHumanHand は人間の手札を表示用にスート→ランク順で並べ替える。
func (g *Tablanet) sortHumanHand() {
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

func (g *Tablanet) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

func (g *Tablanet) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn は現在の手番が人間かどうかを返す。
func (g *Tablanet) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	return g.players[g.state.currentTurn].GetIsHuman()
}

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (g *Tablanet) GetCurrentTurn() int { return g.state.currentTurn }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *Tablanet) SetCurrentTurn(idx int) { g.state.currentTurn = idx }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Tablanet) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *Tablanet) GetPhase() TablanetPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Tablanet) SetPhase(p TablanetPhase) { g.state.phase = p }

// GetTableCards は場の札を返す。
func (g *Tablanet) GetTableCards() []*Card { return g.state.tableCards }

// SetTableCards は場の札を設定する (テスト用)。
func (g *Tablanet) SetTableCards(cards []*Card) { g.state.tableCards = cards }

// GetLastCaptureIdx は最後に捕獲したプレイヤーを返す (-1 = なし)。
func (g *Tablanet) GetLastCaptureIdx() int { return g.state.lastCaptureIdx }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Tablanet) GetPlayer(i int) *TablanetPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetPlayerCnt はプレイヤー数を返す。
func (g *Tablanet) GetPlayerCnt() int { return len(g.players) }

// GetRemainingDeck は山札の残り枚数を返す。
func (g *Tablanet) GetRemainingDeck() int { return g.trumpCards.GetRemainingCount() }

// GetRoundNumber はこれまでに配布されたパック数 (配り直し回数) を返す。
func (g *Tablanet) GetRoundNumber() int { return g.state.packsDealt }

// GetConfig はローカルルール設定を返す。
func (g *Tablanet) GetConfig() TablanetConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Tablanet) SetConfig(cfg TablanetConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *Tablanet) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// GetWinners はゲーム終了時の勝者シートのリストを返す (同点なら複数)。
func (g *Tablanet) GetWinners() []int { return g.state.winners }

// GetLastDealDetail は直前ゲームの得点内訳を返す (nil の場合もある)。
func (g *Tablanet) GetLastDealDetail() *TablanetScoreDetail { return g.state.lastDealDetail }

// GetPlayableIndices はプレイフェーズで人間がプレイできる手札インデックス (全札) を返す。
// タブラネットでは常に任意の手札を出せる (捕獲できなければトレイル)。
func (g *Tablanet) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != TablanetPhasePlay {
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
func (g *Tablanet) GetCaptureOptions(playerIdx int) map[int][]int {
	out := make(map[int][]int)
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != TablanetPhasePlay {
		return out
	}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		caps := g.tablanetFindCaptures(p.GetCard(i))
		if len(caps) > 0 {
			out[i] = caps
		}
	}
	return out
}

// --- JSON Serialization ---

// tablanetJSON is the JSON wire format for Tablanet.
type tablanetJSON struct {
	TrumpCards     *TrumpCards          `json:"tc"`
	Players        []*TablanetPlayer    `json:"pl"`
	Config         TablanetConfig       `json:"cf"`
	Phase          TablanetPhase        `json:"ph"`
	CurrentTurn    int                  `json:"ct"`
	TableCards     []*Card              `json:"tb"`
	LastCaptureIdx int                  `json:"lc"`
	PacksDealt     int                  `json:"pd"`
	GameEndFlag    bool                 `json:"ge"`
	Scored         bool                 `json:"sc"`
	Winners        []int                `json:"wn"`
	LastDealDetail *TablanetScoreDetail `json:"ld"`
	ActionLog      []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Tablanet) MarshalJSON() ([]byte, error) {
	return json.Marshal(tablanetJSON{
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

// tablanetMaxSliceLen caps slice sizes during deserialisation to prevent excessive
// memory allocation from malformed input.
const tablanetMaxSliceLen = 1000

// tablanetValidPhase は有効なフェーズかどうか。
func tablanetValidPhase(p TablanetPhase) bool {
	return p == TablanetPhasePlay || p == TablanetPhaseGameEnd
}

// tablanetValidateCards は復元したカードスライスに nil が無いか検証する。
func tablanetValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("tablanet: nil card in state")
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う。
func (g *Tablanet) UnmarshalJSON(data []byte) error {
	var j tablanetJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tablanetMaxSliceLen || len(j.TableCards) > tablanetMaxSliceLen ||
		len(j.ActionLog) > tablanetMaxSliceLen || len(j.Winners) > tablanetMaxSliceLen {
		return fmt.Errorf("tablanet: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("tablanet: invalid config: %w", err)
	}
	if j.TrumpCards == nil {
		return fmt.Errorf("tablanet: missing trump cards in state")
	}
	if len(j.Players) != TablanetPlayerCnt {
		return fmt.Errorf("tablanet: invalid player count %d, expected %d", len(j.Players), TablanetPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("tablanet: nil player in state")
		}
	}
	if !tablanetValidPhase(j.Phase) {
		return fmt.Errorf("tablanet: invalid phase %d", j.Phase)
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= len(j.Players) {
		return fmt.Errorf("tablanet: current turn out of range")
	}
	if j.LastCaptureIdx < -1 || j.LastCaptureIdx >= len(j.Players) {
		return fmt.Errorf("tablanet: last capture index out of range")
	}
	if err := tablanetValidateCards(j.TableCards); err != nil {
		return err
	}
	for _, w := range j.Winners {
		if w < 0 || w >= len(j.Players) {
			return fmt.Errorf("tablanet: winner index out of range")
		}
	}

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.state = tablanetState{
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

// tablanetScoreDetailJSON is the JSON wire format for TablanetScoreDetail.
type tablanetScoreDetailJSON struct {
	Cards          map[int]int `json:"cd"`
	Aces           map[int]int `json:"ac"`
	Jacks          map[int]int `json:"jk"`
	Tablas         map[int]int `json:"tb"`
	HasTenDiamonds int         `json:"t0"`
	HasTwoClubs    int         `json:"c2"`
	MostCards      int         `json:"mc"`
	Gained         map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *TablanetScoreDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(tablanetScoreDetailJSON{
		Cards:          d.Cards,
		Aces:           d.Aces,
		Jacks:          d.Jacks,
		Tablas:         d.Tablas,
		HasTenDiamonds: d.HasTenDiamonds,
		HasTwoClubs:    d.HasTwoClubs,
		MostCards:      d.MostCards,
		Gained:         d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *TablanetScoreDetail) UnmarshalJSON(data []byte) error {
	var j tablanetScoreDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Cards = j.Cards
	d.Aces = j.Aces
	d.Jacks = j.Jacks
	d.Tablas = j.Tablas
	d.HasTenDiamonds = j.HasTenDiamonds
	d.HasTwoClubs = j.HasTwoClubs
	d.MostCards = j.MostCards
	d.Gained = j.Gained
	return nil
}
