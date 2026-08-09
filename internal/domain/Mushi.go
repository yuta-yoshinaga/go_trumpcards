//go:build !js || !wasm || extra2

// Package domain — 虫 (Mushi) のドメインモデル。
//
// 大阪で遊ばれる花札のフィッシング系ゲーム。48 枚から**六月 (牡丹) と七月 (萩)** を
// 抜いた 40 枚で、2 人が場札合わせを行う。
//
// # issue #4418 の仕様案との相違
//
// issue の記載は 4 点が実際のルール (Hanafuda Hub / Fuda Wiki) と異なる。実ルール側を
// 採る:
//
//   - **抜く月は六月・七月**であって八月・九月ではない。八月・九月を抜くと光札 (芒に月)
//     が 1 枚減り、下記のスコア基準 115 が総点の半分でなくなる。
//   - **配札は手札 8 枚・場 4 枚**であって手札 6 枚・場 6 枚ではない。
//   - **「継続か勝負か」の選択は無い。** 虫は花合わせ系で、全札を取り切ってから精算する。
//     ここをこいこい型にすると別のゲームになる。
//   - **2 人専用。**
//
// issue が触れていない、このゲーム最大の特徴が**柳の雷札 (11-4) のワイルド**である。
//
// # 得点
//
// 札の点は 光 20 / 種 10 / 短冊 5 / カス 1。40 枚の内訳は 光 5・種 7・短冊 8・カス 20 で
// 合計 **230 点**、その半分が MushiScoreBaseline = 115 になる。ラウンドの得点は
//
//	自分の役 + 取り札点 − 相手の役 − 115
//
// で、「折半からどれだけ離れたか」を表す。この 115 が総点の半分であることは
// TestMushi_ScoreBaselineIsHalfTheDeck が数え上げで固定している。
//
// # 取り残される札 (未解決)
//
// 配札は手札 8+8・場 4 で山札 20 になる。手番ごとに手札 1 枚と山札 1 枚を処理するので、
// 手札が尽きるまでに山札は 16 枚しか消費されず、**平均して山札に約 4 枚・場に約 3〜4 枚が
// 残る**。これらは誰の得点にもならない。
//
// 出典は "play on until all cards have been captured" とだけ述べており、この余りの
// 扱いを書いていない。最後に取った側へ渡す実装も考えられるが、それは推測になるので
// 採らず、観測される挙動を TestMushi_LeftoverCardsScoreForNobody で明示的に固定して
// ある。資料が見つかればここだけ変えればよい。
//
// # 札の符号化 (ADR-0033)
//
//   - design = 月 (1..12、ただし 6 と 7 は使わない)
//   - value  = 月内インデックス (1..4)
//
// 花札は専用 PNG を持たないため手続き的に描画される。札の正体表 mushiCardTable は
// KoiKoi/GoStop/HachiHachi が各自持つものと同じ canonical な対応で、バケットが分かれる
// 以上そちらから参照できないため重複している (既存 3 ゲームと同じ扱い)。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// MushiPlayerCnt は虫のプレイヤー数 (固定 2)。
const MushiPlayerCnt = 2

// MushiHandSize は各プレイヤーへ配る手札枚数。
const MushiHandSize = 8

// MushiFieldSize はゲーム開始時に場へ置くカード枚数。
const MushiFieldSize = 4

// MushiCardsPerMonth は 1 か月あたりの札数。
const MushiCardsPerMonth = 4

// MushiDeckSize は虫で使う札の総数 (12 か月 − 2 か月) × 4。
const MushiDeckSize = 40

// MushiTotalCardPoints は 40 枚の札点の合計。
const MushiTotalCardPoints = 230

// MushiScoreBaseline はラウンド得点の基準。総点のちょうど半分で、これを引くことで
// 得点が「折半からの差分」になる。
const MushiScoreBaseline = MushiTotalCardPoints / 2

// MushiMaxRounds は既定の局数。
const MushiMaxRounds = 12

// mushiSkippedMonths は虫で使わない月 (六月=牡丹, 七月=萩)。
var mushiSkippedMonths = map[int]bool{6: true, 7: true}

// MushiWildMonth / MushiWildIndex は柳の雷札 (ワイルド) の位置。
const (
	MushiWildMonth = 11
	MushiWildIndex = 4
)

// MushiCategory は札の種別。
type MushiCategory int

// 札種別定数
const (
	// MushiChaff カス (1 点)
	MushiChaff MushiCategory = iota
	// MushiRibbon 短冊 (5 点)
	MushiRibbon
	// MushiAnimal 種 (10 点)
	MushiAnimal
	// MushiBright 光 (20 点)
	MushiBright
)

// mushiCategoryPoints は種別ごとの札点。
var mushiCategoryPoints = map[MushiCategory]int{
	MushiChaff:  1,
	MushiRibbon: 5,
	MushiAnimal: 10,
	MushiBright: 20,
}

// mushiCardInfo は 1 枚の花札の正体。
type mushiCardInfo struct {
	category MushiCategory
	glyph    string
	name     string
}

// mushiCardTable[month][index] (month 1..12, index 1..4) が札の正体を返す。
// 六月・七月の行は虫では使わないが、月番号を詰めると既存の花札描画パス
// (design=月) と食い違うため、表は 12 か月ぶん持ったままにする。
var mushiCardTable = [13][MushiCardsPerMonth + 1]mushiCardInfo{
	1: {
		1: {MushiBright, "🦢", "Crane"},
		2: {MushiRibbon, "🎴", "RedPoetryRibbon"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	2: {
		1: {MushiAnimal, "🐦", "Warbler"},
		2: {MushiRibbon, "🎴", "RedPoetryRibbon"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	3: {
		1: {MushiBright, "🌸", "Curtain"},
		2: {MushiRibbon, "🎴", "RedPoetryRibbon"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	4: {
		1: {MushiAnimal, "🐦", "Cuckoo"},
		2: {MushiRibbon, "🎴", "RedRibbon"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	5: {
		1: {MushiAnimal, "🌉", "Bridge"},
		2: {MushiRibbon, "🎴", "RedRibbon"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	// 6 (牡丹) と 7 (萩) は虫では使わない。
	8: {
		1: {MushiBright, "🌕", "Moon"},
		2: {MushiAnimal, "🦆", "Geese"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	9: {
		1: {MushiAnimal, "🍶", "SakeCup"},
		2: {MushiRibbon, "🎴", "BlueRibbon"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	10: {
		1: {MushiAnimal, "🦌", "Deer"},
		2: {MushiRibbon, "🎴", "BlueRibbon"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
	11: {
		1: {MushiBright, "☂️", "Rainman"},
		2: {MushiAnimal, "🐦", "Swallow"},
		3: {MushiRibbon, "🎴", "RedRibbon"},
		4: {MushiChaff, "⚡", "Lightning"}, // ワイルド
	},
	12: {
		1: {MushiBright, "🦅", "Phoenix"},
		2: {MushiChaff, "🍂", "Chaff"},
		3: {MushiChaff, "🍂", "Chaff"},
		4: {MushiChaff, "🍂", "Chaff"},
	},
}

// MushiCardCategory は札の種別を返す。範囲外は MushiChaff。
func MushiCardCategory(c *Card) MushiCategory {
	if c == nil {
		return MushiChaff
	}
	m, i := c.GetDesign(), c.GetValue()
	if m < 1 || m > 12 || i < 1 || i > MushiCardsPerMonth {
		return MushiChaff
	}
	return mushiCardTable[m][i].category
}

// MushiCardPoints は札の点を返す。nil は 0 点。
//
// nil を明示的に弾くこと。MushiCardCategory(nil) はカス (iota の 0) を返し、カスは
// 1 点なので、素通しにすると存在しない札が 1 点を生む。
func MushiCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	return mushiCategoryPoints[MushiCardCategory(c)]
}

// MushiIsWild は柳の雷札かを返す。この 1 枚だけが任意の札を取れる。
func MushiIsWild(c *Card) bool {
	return c != nil && c.GetDesign() == MushiWildMonth && c.GetValue() == MushiWildIndex
}

// MushiYaku は成立した役 1 件。
type MushiYaku struct {
	Key    string `json:"key"`
	Points int    `json:"points"`
}

// 役の点数
const (
	// MushiPointsGoko 五光 (光 5 枚すべて)
	MushiPointsGoko = 30
	// MushiPointsSanko 三光
	MushiPointsSanko = 25
	// MushiPointsKiriShima 桐島 (十二月の全札)
	MushiPointsKiriShima = 10
	// MushiPointsFujiShima 藤島 (四月の全札)
	MushiPointsFujiShima = 10
)

// mushiSankoCards は三光の構成札。
//
// **一月の光・三月の光に加えて「二月の種 (梅に鶯)」を含む。** 光札 3 枚ではないので、
// 標準的な花札の三光とは別物であることに注意。この構成は虫を扱う資料
// (Hanafuda Hub) に拠るもので、日本語の一般的な花札資料は虫の役を扱っていない。
// 単一出典なので、異論があればここだけ差し替えれば済むよう表に切り出してある。
var mushiSankoCards = [][2]int{{1, 1}, {3, 1}, {2, 1}}

// MushiDetectYaku は取り札から成立している役を返す。
//
// 五光と三光は同時に成立しうるが、五光が成立していれば三光は数えない。三光は五光の
// 下位互換ではない (二月の種を含む) が、両方数えると光を集めた側が二重取りになる。
func MushiDetectYaku(captured []*Card) []MushiYaku {
	has := make(map[[2]int]bool, len(captured))
	brights := 0
	monthCount := map[int]int{}
	for _, c := range captured {
		if c == nil {
			continue
		}
		key := [2]int{c.GetDesign(), c.GetValue()}
		if has[key] {
			continue
		}
		has[key] = true
		monthCount[c.GetDesign()]++
		if MushiCardCategory(c) == MushiBright {
			brights++
		}
	}

	var yaku []MushiYaku
	switch brights {
	case 5:
		yaku = append(yaku, MushiYaku{Key: "goko", Points: MushiPointsGoko})
	default:
		sanko := true
		for _, mi := range mushiSankoCards {
			if !has[[2]int{mi[0], mi[1]}] {
				sanko = false
				break
			}
		}
		if sanko {
			yaku = append(yaku, MushiYaku{Key: "sanko", Points: MushiPointsSanko})
		}
	}
	if monthCount[12] == MushiCardsPerMonth {
		yaku = append(yaku, MushiYaku{Key: "kirishima", Points: MushiPointsKiriShima})
	}
	if monthCount[4] == MushiCardsPerMonth {
		yaku = append(yaku, MushiYaku{Key: "fujishima", Points: MushiPointsFujiShima})
	}
	return yaku
}

// MushiYakuPoints は役の合計点を返す。
func MushiYakuPoints(yaku []MushiYaku) int {
	total := 0
	for _, y := range yaku {
		total += y.Points
	}
	return total
}

// MushiPhase はゲームフェーズ。
type MushiPhase int

// Mushiのフェーズ定数
const (
	// MushiPhasePlay 札を出すフェーズ
	MushiPhasePlay MushiPhase = iota
	// MushiPhaseSelect 場に同月が 2 枚あり、どちらを取るか選ぶフェーズ
	MushiPhaseSelect
	// MushiPhaseWildSelect ワイルドを出し、取る札を選ぶフェーズ
	MushiPhaseWildSelect
	// MushiPhaseRoundEnd ラウンド終了
	MushiPhaseRoundEnd
	// MushiPhaseGameEnd 終局
	MushiPhaseGameEnd
)

// newMushiDeck は虫の 40 枚を生成する (シャッフル前)。
func newMushiDeck() []*Card {
	deck := make([]*Card, 0, MushiDeckSize)
	for month := 1; month <= 12; month++ {
		if mushiSkippedMonths[month] {
			continue
		}
		for i := 1; i <= MushiCardsPerMonth; i++ {
			deck = append(deck, NewCard(month, i, true))
		}
	}
	return deck
}

// Mushi は虫のゲームクラス。
type Mushi struct {
	players      []*MushiPlayer
	config       MushiConfig
	phase        MushiPhase
	stock        []*Card
	field        []*Card
	captured     [][]*Card
	currentIdx   int
	dealerIdx    int
	roundNumber  int
	scores       []int
	pending      *Card // 選択待ちの、出した/めくった札
	pendingFlip  bool  // pending が山札めくりに由来するか
	roundResults []int
	gameEndFlag  bool
	winnerIdx    int
	actionLogBase
}

// NewMushi はコンストラクタ。
func NewMushi(players []*MushiPlayer, config MushiConfig) *Mushi {
	return &Mushi{
		players:   players,
		config:    config,
		captured:  make([][]*Card, len(players)),
		scores:    make([]int, len(players)),
		winnerIdx: -1,
	}
}

// NewDefaultMushi は標準の 2 人セットアップを返す。
func NewDefaultMushi() *Mushi {
	return NewMushi([]*MushiPlayer{NewMushiPlayer(true), NewMushiPlayer(false)}, DefaultMushiConfig())
}

// Reset はゲーム全体を初期化する。
func (m *Mushi) Reset() {
	m.roundNumber = 0
	m.dealerIdx = 0
	m.scores = make([]int, len(m.players))
	m.roundResults = nil
	m.gameEndFlag = false
	m.winnerIdx = -1
	m.actionLog = nil
	m.startRound()
}

// startRound は 1 ラウンドを配り直す。
func (m *Mushi) startRound() {
	m.phase = MushiPhasePlay
	m.pending = nil
	m.pendingFlip = false
	m.field = nil
	m.captured = make([][]*Card, len(m.players))
	for i := range m.captured {
		m.captured[i] = make([]*Card, 0, MushiDeckSize)
	}
	for _, p := range m.players {
		p.ResetGame()
	}

	deck := newMushiDeck()
	mushiShuffle(deck)

	pos := 0
	for range MushiHandSize {
		for _, p := range m.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	m.field = append(m.field, deck[pos:pos+MushiFieldSize]...)
	pos += MushiFieldSize
	m.stock = append([]*Card(nil), deck[pos:]...)

	// 初期場札に雷が出ていたら親が必ず取る。取れる相手が場に無くても、
	// ワイルドは単独で親の取り札になる。
	for i, c := range m.field {
		if MushiIsWild(c) {
			m.field = append(m.field[:i], m.field[i+1:]...)
			m.captured[m.dealerIdx] = append(m.captured[m.dealerIdx], c)
			m.addLog(m.dealerIdx, "wild", "dealer claims the lightning card from the field", []*Card{c})
			break
		}
	}

	m.currentIdx = m.dealerIdx
	m.roundNumber++
	m.addLog(-1, "deal", fmt.Sprintf("round %d dealt", m.roundNumber), nil)
}

// mushiShuffle は Fisher-Yates。TrumpCards を通さないのは虫の 40 枚が標準デッキの
// 部分集合ではなく月別のサブセットだからで、専用名にしてあるのは domain 内の
// shuffleCards が casino タグのファイル (HoldemEquity.go) にあり、extra2 ビルドでは
// 見えないうえ非 WASM ビルドでは名前が衝突するため。
func mushiShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// PlayCard は手札 idx の札を出す。
//
// 場に同月が 1 枚あれば即座に両方を取り、2 枚以上あれば選択フェーズへ移る。
// 一致が無ければ場に置く。ワイルドを出した場合は取る札の選択フェーズへ移る。
func (m *Mushi) PlayCard(player, handIdx int) error {
	if m.gameEndFlag || m.phase == MushiPhaseRoundEnd {
		return fmt.Errorf("the round is over")
	}
	if m.phase != MushiPhasePlay {
		return fmt.Errorf("a selection is pending")
	}
	if player != m.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := m.GetPlayer(player)
	if p == nil {
		return fmt.Errorf("no such player: %d", player)
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}

	card := p.RemoveCard(handIdx)
	if card == nil {
		return fmt.Errorf("card index %d is empty", handIdx)
	}
	m.resolvePlay(player, card, false)
	return nil
}

// resolvePlay は 1 枚出した (またはめくった) 結果を場へ反映する。
func (m *Mushi) resolvePlay(player int, card *Card, fromFlip bool) {
	if MushiIsWild(card) && !fromFlip && len(m.wildTargets()) > 0 {
		// ワイルドは取る相手を選ぶ。めくりでは選択させず通常の月合わせにする
		// (自動処理の途中で入力を求めると、CPU の手番が人間の入力待ちになる)。
		m.pending = card
		m.pendingFlip = false
		m.phase = MushiPhaseWildSelect
		return
	}

	matches := m.matchIndices(card.GetDesign())
	switch len(matches) {
	case 0:
		m.field = append(m.field, card)
		m.addLog(player, "place", "no match; card goes to the field", []*Card{card})
	case 1:
		m.capture(player, card, []int{matches[0]})
	case 3:
		// 場に 3 枚あれば 4 枚すべてが揃うので選択の余地はない。
		m.capture(player, card, matches)
	default:
		m.pending = card
		m.pendingFlip = fromFlip
		m.phase = MushiPhaseSelect
		return
	}
	m.afterResolve(player, fromFlip)
}

// SelectCapture は選択フェーズで場札 fieldIdx を取る。
func (m *Mushi) SelectCapture(player, fieldIdx int) error {
	if m.phase != MushiPhaseSelect && m.phase != MushiPhaseWildSelect {
		return fmt.Errorf("no selection is pending")
	}
	if player != m.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if fieldIdx < 0 || fieldIdx >= len(m.field) {
		return fmt.Errorf("field index %d out of range", fieldIdx)
	}
	target := m.field[fieldIdx]
	if m.phase == MushiPhaseWildSelect {
		if target.GetDesign() == MushiWildMonth {
			return fmt.Errorf("the lightning card cannot take another willow card")
		}
	} else if target.GetDesign() != m.pending.GetDesign() {
		return fmt.Errorf("field card %d is not the same month", fieldIdx)
	}

	card, fromFlip := m.pending, m.pendingFlip
	m.pending, m.pendingFlip = nil, false
	m.phase = MushiPhasePlay
	m.capture(player, card, []int{fieldIdx})
	m.afterResolve(player, fromFlip)
	return nil
}

// wildTargets はワイルドで取れる場札の添字を返す (柳の札は取れない)。
func (m *Mushi) wildTargets() []int {
	var out []int
	for i, c := range m.field {
		if c != nil && c.GetDesign() != MushiWildMonth {
			out = append(out, i)
		}
	}
	return out
}

// matchIndices は月 month に一致する場札の添字を返す。
func (m *Mushi) matchIndices(month int) []int {
	var out []int
	for i, c := range m.field {
		if c != nil && c.GetDesign() == month {
			out = append(out, i)
		}
	}
	return out
}

// capture は出した札と指定の場札を取り札へ移す。
func (m *Mushi) capture(player int, card *Card, fieldIdxs []int) {
	taken := []*Card{card}
	sorted := append([]int(nil), fieldIdxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, i := range sorted {
		if i < 0 || i >= len(m.field) {
			continue
		}
		taken = append(taken, m.field[i])
		m.field = append(m.field[:i], m.field[i+1:]...)
	}
	m.captured[player] = append(m.captured[player], taken...)
	m.addLog(player, "capture", fmt.Sprintf("captures %d card(s)", len(taken)), taken)
}

// afterResolve は 1 枚ぶんの処理が終わった後の進行 (山札めくり → 手番交代)。
func (m *Mushi) afterResolve(player int, wasFlip bool) {
	if !wasFlip {
		if len(m.stock) > 0 {
			flip := m.stock[0]
			m.stock = m.stock[1:]
			m.resolvePlay(player, flip, true)
			return
		}
	}
	if m.phase == MushiPhaseSelect || m.phase == MushiPhaseWildSelect {
		return
	}
	m.currentIdx = (player + 1) % len(m.players)
	if m.handsEmpty() {
		m.finishRound()
	}
}

// handsEmpty は全員の手札が尽きたかを返す。
func (m *Mushi) handsEmpty() bool {
	for _, p := range m.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishRound はラウンドを精算する。
//
// 得点は「自分の役 + 取り札点 − 相手の役 − 115」。115 は 40 枚の総点 230 の半分なので、
// これは折半からの差分を表す。役も取り札点も両者ぶん数えるため、2 人の得点は必ず
// 符号が反転した同じ絶対値になる。
func (m *Mushi) finishRound() {
	m.phase = MushiPhaseRoundEnd

	yakuPts := make([]int, len(m.players))
	cardPts := make([]int, len(m.players))
	for i := range m.players {
		yakuPts[i] = MushiYakuPoints(MushiDetectYaku(m.captured[i]))
		for _, c := range m.captured[i] {
			cardPts[i] += MushiCardPoints(c)
		}
	}

	m.roundResults = make([]int, len(m.players))
	for i := range m.players {
		opp := (i + 1) % len(m.players)
		delta := yakuPts[i] + cardPts[i] - yakuPts[opp] - MushiScoreBaseline
		m.roundResults[i] = delta
		m.scores[i] += delta
	}
	m.addLog(-1, "round", fmt.Sprintf("round %d settled", m.roundNumber), nil)

	if m.roundNumber >= m.config.TargetRounds {
		m.finishGame()
	}
}

// NextRound は次のラウンドを開始する。
func (m *Mushi) NextRound() error {
	if m.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if m.phase != MushiPhaseRoundEnd {
		return fmt.Errorf("the round is still in progress")
	}
	m.dealerIdx = (m.dealerIdx + 1) % len(m.players)
	m.startRound()
	return nil
}

// finishGame は終局処理。
func (m *Mushi) finishGame() {
	m.gameEndFlag = true
	m.phase = MushiPhaseGameEnd
	best, tie := 0, false
	for i := 1; i < len(m.scores); i++ {
		switch {
		case m.scores[i] > m.scores[best]:
			best, tie = i, false
		case m.scores[i] == m.scores[best]:
			tie = true
		}
	}
	if tie {
		m.winnerIdx = -1
	} else {
		m.winnerIdx = best
	}
	m.addLog(-1, "game", "game over", nil)
}

// addLog は棋譜へ 1 行追加する。
func (m *Mushi) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	m.appendLog(playerIdx, actionType, detail, cards)
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (m *Mushi) GetPlayers() []*MushiPlayer { return m.players }

// GetPlayer は idx のプレイヤーを返す。範囲外は nil。
func (m *Mushi) GetPlayer(idx int) *MushiPlayer {
	return getPlayer(m.players, idx)
}

// GetField は場札を返す。
func (m *Mushi) GetField() []*Card { return m.field }

// GetStockCount は山札の残り枚数を返す。
func (m *Mushi) GetStockCount() int { return len(m.stock) }

// GetCaptured は idx の取り札を返す。
func (m *Mushi) GetCaptured(idx int) []*Card {
	if idx < 0 || idx >= len(m.captured) {
		return nil
	}
	return m.captured[idx]
}

// GetPhase は現在のフェーズを返す。
func (m *Mushi) GetPhase() MushiPhase { return m.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (m *Mushi) GetCurrentPlayerIdx() int { return m.currentIdx }

// GetDealerIdx は親の添字を返す。
func (m *Mushi) GetDealerIdx() int { return m.dealerIdx }

// GetRoundNumber は現在のラウンド番号 (1 始まり) を返す。
func (m *Mushi) GetRoundNumber() int { return m.roundNumber }

// GetScore は idx の累計得点を返す。
func (m *Mushi) GetScore(idx int) int {
	return elemAt(m.scores, idx)
}

// SetScore は idx の累計得点を設定する (テスト用)。
func (m *Mushi) SetScore(idx, score int) {
	if idx < 0 || idx >= len(m.scores) {
		return
	}
	m.scores[idx] = score
}

// GetRoundResult は直前のラウンドでの idx の増減を返す。
func (m *Mushi) GetRoundResult(idx int) int {
	if idx < 0 || idx >= len(m.roundResults) {
		return 0
	}
	return m.roundResults[idx]
}

// GetPendingCard は選択待ちの札を返す (無ければ nil)。
func (m *Mushi) GetPendingCard() *Card { return m.pending }

// GetSelectableIndices は選択フェーズで取れる場札の添字を返す。
func (m *Mushi) GetSelectableIndices() []int {
	switch m.phase {
	case MushiPhaseWildSelect:
		return m.wildTargets()
	case MushiPhaseSelect:
		if m.pending == nil {
			return nil
		}
		return m.matchIndices(m.pending.GetDesign())
	default:
		return nil
	}
}

// GetGameEndFlag は終局しているかを返す。
func (m *Mushi) GetGameEndFlag() bool { return m.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す。未確定・引き分けは -1。
func (m *Mushi) GetWinnerIdx() int { return m.winnerIdx }

// GetConfig はゲーム設定を返す。
func (m *Mushi) GetConfig() MushiConfig { return m.config }

// SetConfig はゲーム設定を差し替える。
func (m *Mushi) SetConfig(c MushiConfig) { m.config = c }

// ---- JSON ----

// mushiJSON は KV のワイヤ形式。Worker は毎リクエストここから組み直すので、
// ここに無いものは次のリクエストでは存在しない。
type mushiJSON struct {
	Players      []*MushiPlayer    `json:"pl"`
	Config       MushiConfig       `json:"cf"`
	Phase        MushiPhase        `json:"ph"`
	Stock        []*Card           `json:"st"`
	Field        []*Card           `json:"fd"`
	Captured     [][]*Card         `json:"cp"`
	CurrentIdx   int               `json:"ci"`
	DealerIdx    int               `json:"di"`
	RoundNumber  int               `json:"rn"`
	Scores       []int             `json:"sc"`
	Pending      *Card             `json:"pd"`
	PendingFlip  bool              `json:"pf"`
	RoundResults []int             `json:"rr"`
	GameEndFlag  bool              `json:"ge"`
	WinnerIdx    int               `json:"wi"`
	ActionLog    []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (m *Mushi) MarshalJSON() ([]byte, error) {
	return json.Marshal(mushiJSON{
		Players:      m.players,
		Config:       m.config,
		Phase:        m.phase,
		Stock:        m.stock,
		Field:        m.field,
		Captured:     m.captured,
		CurrentIdx:   m.currentIdx,
		DealerIdx:    m.dealerIdx,
		RoundNumber:  m.roundNumber,
		Scores:       m.scores,
		Pending:      m.pending,
		PendingFlip:  m.pendingFlip,
		RoundResults: m.roundResults,
		GameEndFlag:  m.gameEndFlag,
		WinnerIdx:    m.winnerIdx,
		ActionLog:    m.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// Worker はこれを KV の未検証バイト列に対して毎リクエスト実行する。添字は信用せず
// 丸める。壊れた currentIdx はそのままだとプレイヤー配列の外を指す。
func (m *Mushi) UnmarshalJSON(data []byte) error {
	var j mushiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) == 0 {
		return fmt.Errorf("mushi: no players in snapshot")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("mushi: %w", err)
	}
	m.players = j.Players
	m.config = j.Config
	m.phase = j.Phase
	m.stock = j.Stock
	m.field = j.Field
	m.roundNumber = j.RoundNumber
	m.pending = j.Pending
	m.pendingFlip = j.PendingFlip
	m.gameEndFlag = j.GameEndFlag
	m.actionLog = j.ActionLog

	n := len(m.players)
	m.currentIdx = mushiClampSeat(j.CurrentIdx, n)
	m.dealerIdx = mushiClampSeat(j.DealerIdx, n)
	if m.dealerIdx < 0 {
		m.dealerIdx = 0
	}
	m.winnerIdx = mushiClampSeat(j.WinnerIdx, n)

	m.scores = make([]int, n)
	copy(m.scores, j.Scores)
	m.roundResults = make([]int, n)
	copy(m.roundResults, j.RoundResults)

	m.captured = make([][]*Card, n)
	for i := range m.captured {
		if i < len(j.Captured) && j.Captured[i] != nil {
			m.captured[i] = j.Captured[i]
		} else {
			m.captured[i] = make([]*Card, 0, MushiDeckSize)
		}
	}
	return nil
}

// ---- CPU ----

// MushiCpuAction は CPU が選んだ手。
type MushiCpuAction struct {
	// HandIdx は出す手札の添字 (選択フェーズでは -1)。
	HandIdx int
	// FieldIdx は選択フェーズで取る場札の添字 (それ以外では -1)。
	FieldIdx int
}

// MushiCpuDecide は idx の CPU が取る手を決める。
//
// 選択フェーズなら最も点の高い場札を取り、そうでなければ最も点の高い捕獲になる
// 手札を出す。捕獲できる手が無ければ、場に置いても相手に取られにくい札
// (点が低いもの) を捨てる。
func (m *Mushi) MushiCpuDecide(idx int) MushiCpuAction {
	if m.phase == MushiPhaseSelect || m.phase == MushiPhaseWildSelect {
		best, bestPts := -1, -1
		for _, i := range m.GetSelectableIndices() {
			if pts := MushiCardPoints(m.field[i]); pts > bestPts {
				best, bestPts = i, pts
			}
		}
		return MushiCpuAction{HandIdx: -1, FieldIdx: best}
	}

	p := m.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return MushiCpuAction{HandIdx: -1, FieldIdx: -1}
	}

	bestIdx, bestGain := -1, -1
	worstIdx, worstPts := 0, -1
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		gain := 0
		if MushiIsWild(c) {
			// ワイルドは場の最高点を取れる。
			for _, t := range m.wildTargets() {
				if pts := MushiCardPoints(m.field[t]); pts > gain {
					gain = pts
				}
			}
			if gain > 0 {
				gain += MushiCardPoints(c)
			}
		} else if matches := m.matchIndices(c.GetDesign()); len(matches) > 0 {
			gain = MushiCardPoints(c)
			for _, t := range matches {
				gain += MushiCardPoints(m.field[t])
			}
		}
		if gain > bestGain {
			bestIdx, bestGain = i, gain
		}
		if pts := MushiCardPoints(c); worstPts == -1 || pts < worstPts {
			worstIdx, worstPts = i, pts
		}
	}
	if bestGain > 0 {
		return MushiCpuAction{HandIdx: bestIdx, FieldIdx: -1}
	}
	return MushiCpuAction{HandIdx: worstIdx, FieldIdx: -1}
}

// mushiClampSeat は範囲外のプレイヤー添字を -1 (未確定) に丸める。
//
// domain には同じ働きの clampPlayerIdx があるが、あれは Bura.go すなわち extra3 タグの
// ファイルにあり、extra2 ビルドからは見えない。`go build ./...` は !js||!wasm 側を
// コンパイルするため両方が存在してしまい、この種の取り違えを検出できない
// (GOOS=js GOARCH=wasm -tags extra2 のビルドで初めて落ちる)。
func mushiClampSeat(idx, n int) int {
	if idx < 0 || idx >= n {
		return -1
	}
	return idx
}

// MushiCardGlyph は手続き描画用の絵文字を返す (ADR-0033)。
func MushiCardGlyph(c *Card) string {
	if c == nil {
		return ""
	}
	m, i := c.GetDesign(), c.GetValue()
	if m < 1 || m > 12 || i < 1 || i > MushiCardsPerMonth {
		return ""
	}
	return mushiCardTable[m][i].glyph
}

// MushiCardName は札の短い識別名を返す。
func MushiCardName(c *Card) string {
	if c == nil {
		return ""
	}
	m, i := c.GetDesign(), c.GetValue()
	if m < 1 || m > 12 || i < 1 || i > MushiCardsPerMonth {
		return ""
	}
	return mushiCardTable[m][i].name
}
