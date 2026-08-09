//go:build !js || !wasm || extra3

// Package domain — こいこい (Koi-Koi) のドメインモデル。
//
// Koi-Koi は花札 (hanafuda) 48 枚 (12 か月 × 4 枚) を用いる 2 人用 (人間 vs CPU) の
// フィッシング系ゲーム。手札から 1 枚出して場の同月札を捕獲し、取り札で「役 (やく)」を
// 作って得点する。役が成立した時点で「こいこい (続行して更に高い役を狙う)」か
// 「勝負 / あがり (そこで得点を確定)」かを選ぶ駆け引きが核。
//
// # 札の符号化 (ADR-0033)
//
//   - design = 月 (1..12)
//   - value  = 月内インデックス (1..4)
//
// 花札には専用 PNG アートが無いため、各札は手続き的に描画される (glyph/label/color/
// deck="hanafuda")。カードの正体 (光/タネ/短冊/カス、短冊の色) は (月, インデックス) の
// 対応表 koikoiCardTable から引く。
//
// # ターン進行
//
// 手番プレイヤーは手札を 1 枚出す。場に同月の札があれば両方を取り札にする (2 枚一致なら
// どちらを取るか選択、3 枚一致=場に既に 3 枚 → すべて取る)。一致が無ければ場に置く
// (捨て札)。続いて山札の一番上をめくり、同様に同月捕獲 / 場に追加する。ターン後、取り札が
// 新規または改善された役を形成していれば「こいこい決断」フェーズへ移り、続行 (こいこい) か
// 停止 (勝負) を選ぶ。
//
// # こいこいと倍率 (簡略化 — 本実装のルール)
//
// このラウンド中に誰かが 1 回でもこいこいを宣言していた場合、そのラウンドの勝者の得点を
// 2 倍にする (koikoiCount>=1 → ×2)。「こいこいすると相手が上がった時のリスクが上がる」
// 性質をこの単一倍率で表現する簡略ルールで、標準的な「各こいこいで倍加」よりも穏当。
//
// # ラウンド終了 / 終局
//
// いずれかが役で「勝負」を宣言 → その役合計 × 倍率を得点。双方の手札が尽きても勝者が
// いなければ引き分け (0 点)。累計得点が TargetScore に到達したプレイヤーが出るか、
// KoiKoiMaxRounds ラウンドに達したら終局。累計最高点が勝者 (同点は引き分け -1)。
//
// 手役 (配札時の手四/くっつき等) は簡略化のため未実装 (ドキュメント参照)。
package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// KoiKoiPlayerCnt はこいこいのプレイヤー数 (固定 2)。
const KoiKoiPlayerCnt = 2

// KoiKoiHandSize は各プレイヤーへ配る手札枚数。
const KoiKoiHandSize = 8

// KoiKoiFieldSize はゲーム開始時に場へ置くカード枚数。
const KoiKoiFieldSize = 8

// KoiKoiMonthCnt は月数 (= design の最大値)。
const KoiKoiMonthCnt = 12

// KoiKoiCardsPerMonth は 1 月あたりの札数 (= value の最大値)。
const KoiKoiCardsPerMonth = 4

// KoiKoiMaxRounds はゲームを打ち切る最大ラウンド数 (終局保証)。
const KoiKoiMaxRounds = 12

// KoiKoiCategory は花札の札種。
type KoiKoiCategory int

// 札種定数
const (
	// KoiKoiChaff カス (最下位の札)
	KoiKoiChaff KoiKoiCategory = iota
	// KoiKoiRibbon 短冊 (たんざく)
	KoiKoiRibbon
	// KoiKoiAnimal タネ (動物/種)
	KoiKoiAnimal
	// KoiKoiBright 光 (最上位の札)
	KoiKoiBright
)

// KoiKoiRibbonColor は短冊の色。
type KoiKoiRibbonColor int

// 短冊色定数
const (
	// KoiKoiRibbonNone 短冊ではない
	KoiKoiRibbonNone KoiKoiRibbonColor = iota
	// KoiKoiRibbonRedPoetry 赤短 (書き入りの赤短冊: 1/2/3 月)
	KoiKoiRibbonRedPoetry
	// KoiKoiRibbonBlue 青短 (6/9/10 月)
	KoiKoiRibbonBlue
	// KoiKoiRibbonPlainRed 赤短冊 (書き無し: 4/5/7/11 月)
	KoiKoiRibbonPlainRed
)

// koikoiCardInfo は 1 枚の花札の正体。
type koikoiCardInfo struct {
	category KoiKoiCategory
	ribbon   KoiKoiRibbonColor
	glyph    string // 手続き描画用の絵文字
	name     string // 短い識別名 (英語、デバッグ/ラベル補助)
}

// koikoiCardTable[month][index] (month 1..12, index 1..4) が札の正体を返す。
// インデックス 0 は未使用。canonical な割り当て (課題の指定に準拠)。
var koikoiCardTable = [KoiKoiMonthCnt + 1][KoiKoiCardsPerMonth + 1]koikoiCardInfo{
	1: { // 松 (松に鶴)
		1: {KoiKoiBright, KoiKoiRibbonNone, "🦢", "Crane"},
		2: {KoiKoiRibbon, KoiKoiRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	2: { // 梅 (梅に鶯)
		1: {KoiKoiAnimal, KoiKoiRibbonNone, "🐦", "Warbler"},
		2: {KoiKoiRibbon, KoiKoiRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	3: { // 桜 (桜に幕)
		1: {KoiKoiBright, KoiKoiRibbonNone, "🌸", "Curtain"},
		2: {KoiKoiRibbon, KoiKoiRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	4: { // 藤 (藤に不如帰)
		1: {KoiKoiAnimal, KoiKoiRibbonNone, "🐦", "Cuckoo"},
		2: {KoiKoiRibbon, KoiKoiRibbonPlainRed, "🎴", "RedRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	5: { // 菖蒲 (菖蒲に八橋)
		1: {KoiKoiAnimal, KoiKoiRibbonNone, "🌉", "Bridge"},
		2: {KoiKoiRibbon, KoiKoiRibbonPlainRed, "🎴", "RedRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	6: { // 牡丹 (牡丹に蝶)
		1: {KoiKoiAnimal, KoiKoiRibbonNone, "🦋", "Butterfly"},
		2: {KoiKoiRibbon, KoiKoiRibbonBlue, "🎴", "BlueRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	7: { // 萩 (萩に猪)
		1: {KoiKoiAnimal, KoiKoiRibbonNone, "🐗", "Boar"},
		2: {KoiKoiRibbon, KoiKoiRibbonPlainRed, "🎴", "RedRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	8: { // 芒 (芒に月 / 芒に雁)
		1: {KoiKoiBright, KoiKoiRibbonNone, "🌕", "Moon"},
		2: {KoiKoiAnimal, KoiKoiRibbonNone, "🦆", "Geese"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	9: { // 菊 (菊に盃)
		1: {KoiKoiAnimal, KoiKoiRibbonNone, "🍶", "SakeCup"},
		2: {KoiKoiRibbon, KoiKoiRibbonBlue, "🎴", "BlueRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	10: { // 紅葉 (紅葉に鹿)
		1: {KoiKoiAnimal, KoiKoiRibbonNone, "🦌", "Deer"},
		2: {KoiKoiRibbon, KoiKoiRibbonBlue, "🎴", "BlueRibbon"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
	11: { // 柳 (柳に小野道風 / 燕 / 短冊 / 雷)
		1: {KoiKoiBright, KoiKoiRibbonNone, "☂️", "Rainman"},
		2: {KoiKoiAnimal, KoiKoiRibbonNone, "🐦", "Swallow"},
		3: {KoiKoiRibbon, KoiKoiRibbonPlainRed, "🎴", "RedRibbon"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "⚡", "Lightning"},
	},
	12: { // 桐 (桐に鳳凰)
		1: {KoiKoiBright, KoiKoiRibbonNone, "🦅", "Phoenix"},
		2: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		3: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
		4: {KoiKoiChaff, KoiKoiRibbonNone, "🍂", "Chaff"},
	},
}

// koikoiInfo は札の正体を返す。範囲外の札はカス相当を返す (防御的)。
func koikoiInfo(c *Card) koikoiCardInfo {
	if c == nil {
		return koikoiCardInfo{category: KoiKoiChaff, glyph: "🍂", name: "Chaff"}
	}
	m, i := c.GetDesign(), c.GetValue()
	if m < 1 || m > KoiKoiMonthCnt || i < 1 || i > KoiKoiCardsPerMonth {
		return koikoiCardInfo{category: KoiKoiChaff, glyph: "🍂", name: "Chaff"}
	}
	return koikoiCardTable[m][i]
}

// koikoiSameMonth は 2 枚が同月かどうか。
func koikoiSameMonth(a, b *Card) bool {
	return a != nil && b != nil && a.GetDesign() == b.GetDesign()
}

// --- 役の点数 ---
const (
	koikoiPointsGoko        = 10 // 五光
	koikoiPointsAmeShiko    = 7  // 雨四光
	koikoiPointsShiko       = 8  // 四光
	koikoiPointsSanko       = 5  // 三光
	koikoiPointsInoshikacho = 5  // 猪鹿蝶
	koikoiPointsAkatan      = 5  // 赤短
	koikoiPointsAotan       = 5  // 青短
	koikoiPointsHanami      = 5  // 花見酒
	koikoiPointsTsukimi     = 5  // 月見酒
)

// KoiKoiYaku は成立した 1 役。
type KoiKoiYaku struct {
	Key    string `json:"key"`    // 役キー (例 "goko", "tane")
	Points int    `json:"points"` // 点数
}

// koikoiEvaluateYaku は取り札から成立している役の一覧と合計点を返す純粋関数。
// 光の役 (五光/雨四光/四光/三光) は排他で最上位のみ採用する。短冊/タネ/カスは枚数
// 閾値 (5/5/10) を超えると +1 ずつ加点する。赤短/青短/花見/月見は独立に加点。
func koikoiEvaluateYaku(captured []*Card) ([]KoiKoiYaku, int) {
	brights := 0
	hasRain, hasMoon, hasCurtain := false, false, false
	hasBoar, hasDeer, hasButterfly, hasSakeCup := false, false, false, false
	animals, ribbons, redPoetry, blue, chaff := 0, 0, 0, 0, 0

	for _, c := range captured {
		if c == nil {
			continue
		}
		info := koikoiInfo(c)
		month := c.GetDesign()
		switch info.category {
		case KoiKoiBright:
			brights++
			switch month {
			case 11:
				hasRain = true
			case 8:
				hasMoon = true
			case 3:
				hasCurtain = true
			}
		case KoiKoiAnimal:
			animals++
			switch month {
			case 7:
				hasBoar = true
			case 10:
				hasDeer = true
			case 6:
				hasButterfly = true
			case 9:
				hasSakeCup = true
			}
		case KoiKoiRibbon:
			ribbons++
			switch info.ribbon {
			case KoiKoiRibbonRedPoetry:
				redPoetry++
			case KoiKoiRibbonBlue:
				blue++
			}
		case KoiKoiChaff:
			chaff++
		}
	}

	yakus := make([]KoiKoiYaku, 0, 6)
	add := func(key string, pts int) { yakus = append(yakus, KoiKoiYaku{Key: key, Points: pts}) }

	// 光 (排他)。
	switch {
	case brights >= 5:
		add("goko", koikoiPointsGoko)
	case brights == 4 && hasRain:
		add("ameshiko", koikoiPointsAmeShiko)
	case brights == 4:
		add("shiko", koikoiPointsShiko)
	case brights == 3 && !hasRain:
		add("sanko", koikoiPointsSanko)
	}
	// 猪鹿蝶。
	if hasBoar && hasDeer && hasButterfly {
		add("inoshikacho", koikoiPointsInoshikacho)
	}
	// 赤短 / 青短 (それぞれ固有の 3 枚。>=3 は全数一致を意味する)。
	if redPoetry >= 3 {
		add("akatan", koikoiPointsAkatan)
	}
	if blue >= 3 {
		add("aotan", koikoiPointsAotan)
	}
	// 短冊 / タネ / カス (枚数閾値 + 超過 1 枚ごと +1)。
	if ribbons >= 5 {
		add("tanzaku", 1+(ribbons-5))
	}
	if animals >= 5 {
		add("tane", 1+(animals-5))
	}
	if chaff >= 10 {
		add("kasu", 1+(chaff-10))
	}
	// 花見酒 / 月見酒。
	if hasCurtain && hasSakeCup {
		add("hanami", koikoiPointsHanami)
	}
	if hasMoon && hasSakeCup {
		add("tsukimi", koikoiPointsTsukimi)
	}

	total := 0
	for _, y := range yakus {
		total += y.Points
	}
	return yakus, total
}

// KoiKoiPhase はゲームフェーズ。
type KoiKoiPhase int

// Koi-Koi のフェーズ定数
const (
	// KoiKoiPhasePlay プレイ中 (手札を 1 枚出す)
	KoiKoiPhasePlay KoiKoiPhase = 0
	// KoiKoiPhaseKoiKoiDecision こいこい決断中 (続行/勝負を選ぶ)
	KoiKoiPhaseKoiKoiDecision KoiKoiPhase = 1
	// KoiKoiPhaseRoundEnd ラウンド終了 (結果表示。次ラウンド待ち)
	KoiKoiPhaseRoundEnd KoiKoiPhase = 2
	// KoiKoiPhaseGameEnd 終局
	KoiKoiPhaseGameEnd KoiKoiPhase = 3
)

// KoiKoiRoundResult は 1 ラウンドの結果。
type KoiKoiRoundResult struct {
	Winner      int          `json:"winner"`      // 勝者インデックス (-1 = 引き分け)
	Yaku        []KoiKoiYaku `json:"yaku"`        // 成立役 (勝者のもの)
	BasePoints  int          `json:"basePoints"`  // 役合計 (倍率適用前)
	Multiplier  int          `json:"multiplier"`  // 倍率 (こいこい発生で 2)
	Total       int          `json:"total"`       // 実際に加算された得点
	KoikoiCount int          `json:"koikoiCount"` // このラウンドのこいこい宣言回数
}

// KoiKoiHint はヒント情報。
type KoiKoiHint struct {
	CardIndex  int    `json:"cardIndex"`  // 推奨手札インデックス (-1 = なし)
	FieldIndex int    `json:"fieldIndex"` // 推奨捕獲場札インデックス (-1 = なし/自動)
	KoiKoi     int    `json:"koikoi"`     // 決断ヒント (1=こいこい, 0=勝負, -1=非該当)
	Reason     string `json:"reason"`     // 理由キー
}

// koikoiState はゲーム進行状態。
type koikoiState struct {
	phase           KoiKoiPhase
	currentTurn     int
	fieldCards      []*Card
	drawPile        []*Card
	roundNumber     int
	koikoiCount     int // このラウンドのこいこい宣言回数
	roundWinner     int // -1 = 未決/引き分け
	gameEndFlag     bool
	winner          int // 終局時の勝者 (-1 = 引き分け)
	lastRoundResult *KoiKoiRoundResult
	pendingYaku     []KoiKoiYaku // 決断フェーズで表示する役
	pendingPoints   int          // 決断フェーズの役合計
	actionLogBase
}

// KoiKoi はこいこいゲームの状態を保持する集約ルート。
type KoiKoi struct {
	players []*KoiKoiPlayer
	config  KoiKoiConfig
	state   koikoiState
}

// NewKoiKoi はコンストラクタ。
func NewKoiKoi(players []*KoiKoiPlayer, config KoiKoiConfig) *KoiKoi {
	return &KoiKoi{
		players: players,
		config:  config,
		state: koikoiState{
			phase:       KoiKoiPhasePlay,
			roundWinner: -1,
			winner:      -1,
		},
	}
}

// NewDefaultKoiKoi は標準の 2 人構成 (1 human + 1 CPU) で KoiKoi を生成する。
func NewDefaultKoiKoi() *KoiKoi {
	players := make([]*KoiKoiPlayer, KoiKoiPlayerCnt)
	players[0] = NewKoiKoiPlayer(true)
	players[1] = NewKoiKoiPlayer(false)
	return NewKoiKoi(players, DefaultKoiKoiConfig())
}

// buildKoiKoiDeck は花札 48 枚を design=月(1..12)/value=index(1..4) で直接生成する。
func buildKoiKoiDeck() []*Card {
	deck := make([]*Card, 0, KoiKoiMonthCnt*KoiKoiCardsPerMonth)
	for m := 1; m <= KoiKoiMonthCnt; m++ {
		for i := 1; i <= KoiKoiCardsPerMonth; i++ {
			deck = append(deck, NewCard(m, i, false))
		}
	}
	return deck
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。
func (g *KoiKoi) Reset() {
	for _, p := range g.players {
		p.Reset()
		p.ResetRound()
		p.ResetScore()
	}
	g.state = koikoiState{
		phase:         KoiKoiPhasePlay,
		roundWinner:   -1,
		winner:        -1,
		roundNumber:   1,
		actionLogBase: actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound はラウンド終了後に次ラウンドを開始する。
func (g *KoiKoi) NextRound() {
	if g.state.gameEndFlag || g.state.phase != KoiKoiPhaseRoundEnd {
		return
	}
	g.state.roundNumber++
	g.startRound()
}

// startRound はデッキ生成・配札・場札配置を行い、プレイフェーズを開始する。
// ラウンドごとに先手を交代する ((roundNumber-1) % 2)。
func (g *KoiKoi) startRound() {
	deck := buildKoiKoiDeck()
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	for _, p := range g.players {
		p.Reset()
		p.ResetRound()
	}
	g.state.fieldCards = make([]*Card, 0, KoiKoiFieldSize)
	g.state.koikoiCount = 0
	g.state.roundWinner = -1
	g.state.pendingYaku = nil
	g.state.pendingPoints = 0
	g.state.phase = KoiKoiPhasePlay
	g.state.currentTurn = (g.state.roundNumber - 1) % KoiKoiPlayerCnt

	// 交互配り: 各プレイヤーへ 8 枚、場へ 8 枚。
	pos := 0
	for k := 0; k < KoiKoiHandSize; k++ {
		for _, p := range g.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	for k := 0; k < KoiKoiFieldSize; k++ {
		g.state.fieldCards = append(g.state.fieldCards, deck[pos])
		pos++
	}
	g.state.drawPile = append([]*Card(nil), deck[pos:]...)

	g.sortHumanHand()
	g.appendLog(-1, "deal", fmt.Sprintf("round %d dealt (field %d, draw %d)",
		g.state.roundNumber, len(g.state.fieldCards), len(g.state.drawPile)),
		append([]*Card(nil), g.state.fieldCards...))
}

// allHandsEmpty は全員の手札が空かどうか。
func (g *KoiKoi) allHandsEmpty() bool {
	for _, p := range g.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// --- 捕獲ロジック ---

// koikoiFieldMatches は場札のうち card と同月のインデックスを返す。
func (g *KoiKoi) koikoiFieldMatches(card *Card) []int {
	var out []int
	for i, c := range g.state.fieldCards {
		if koikoiSameMonth(c, card) {
			out = append(out, i)
		}
	}
	return out
}

// koikoiBestFieldMatch は複数一致のうち最も価値の高い場札インデックスを返す。
func (g *KoiKoi) koikoiBestFieldMatch(matches []int) int {
	best := -1
	bestVal := -1
	for _, idx := range matches {
		if idx < 0 || idx >= len(g.state.fieldCards) {
			continue
		}
		v := koikoiCardWeight(g.state.fieldCards[idx])
		if v > bestVal {
			bestVal = v
			best = idx
		}
	}
	return best
}

// koikoiCardWeight は札の概算価値 (光>タネ>短冊>カス)。捕獲選択と CPU AI で使用。
func koikoiCardWeight(c *Card) int {
	switch koikoiInfo(c).category {
	case KoiKoiBright:
		return 5
	case KoiKoiAnimal:
		return 3
	case KoiKoiRibbon:
		return 2
	default:
		return 1
	}
}

// koikoiPlaceCard は 1 枚 (手札またはめくり札) を場と突き合わせて解決する。
//   - 一致なし: 場に置く (捨て札)。
//   - 一致 1 枚: その札と共に捕獲。
//   - 一致 2 枚: chosen が一致札なら chosen を、そうでなければ最良の 1 枚を捕獲。
//   - 一致 3 枚: すべて捕獲 (場の同月 4 枚目)。
//
// captured へ取り札を積み、場から取り除く。
func (g *KoiKoi) koikoiPlaceCard(playerIdx int, card *Card, chosen int) {
	matches := g.koikoiFieldMatches(card)
	if len(matches) == 0 {
		g.state.fieldCards = append(g.state.fieldCards, card)
		return
	}
	var take []int
	switch {
	case len(matches) >= 3:
		take = matches
	case len(matches) == 2:
		sel := -1
		for _, idx := range matches {
			if idx == chosen {
				sel = idx
			}
		}
		if sel < 0 {
			sel = g.koikoiBestFieldMatch(matches)
		}
		take = []int{sel}
	default:
		take = []int{matches[0]}
	}
	captured := make([]*Card, 0, len(take)+1)
	captured = append(captured, card)
	for _, idx := range take {
		captured = append(captured, g.state.fieldCards[idx])
	}
	g.removeFieldByIndex(take)
	g.players[playerIdx].AddCaptured(captured)
}

// removeFieldByIndex は降順に並べ替えてから場札を削除する。
func (g *KoiKoi) removeFieldByIndex(idxs []int) {
	sorted := append([]int(nil), idxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, idx := range sorted {
		if idx >= 0 && idx < len(g.state.fieldCards) {
			g.state.fieldCards = append(g.state.fieldCards[:idx], g.state.fieldCards[idx+1:]...)
		}
	}
}

// --- Play ---

// PlayerPlay は人間が手札 handIdx を出す。fieldIdx は同月札が 2 枚ある場合に
// どちらを取るかの場札インデックス (不要なら -1)。
func (g *KoiKoi) PlayerPlay(handIdx, fieldIdx int) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != KoiKoiPhasePlay {
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
	// fieldIdx が指定された場合、同月の場札であることを検証する。
	if fieldIdx >= 0 {
		if fieldIdx >= len(g.state.fieldCards) || !koikoiSameMonth(g.state.fieldCards[fieldIdx], card) {
			return NewDomainError(ErrInvalidPlay, "chosen field card does not match the played card's month")
		}
	}
	g.applyTurn(g.state.currentTurn, handIdx, fieldIdx)
	return nil
}

// CpuPlay は CPU のプレイ手番を 1 回進める。
func (g *KoiKoi) CpuPlay() {
	if g.state.gameEndFlag || g.state.phase != KoiKoiPhasePlay {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() || p.GetCardsSize() == 0 {
		return
	}
	handIdx, fieldIdx := g.chooseCpuPlay(g.state.currentTurn)
	g.applyTurn(g.state.currentTurn, handIdx, fieldIdx)
}

// applyTurn は手札を出し→めくり札を処理→役判定を行う共通処理。役が改善されたら
// こいこい決断フェーズへ移り、そうでなければ手番を進める。
func (g *KoiKoi) applyTurn(playerIdx, handIdx, fieldIdx int) {
	player := g.players[playerIdx]
	card := player.RemoveCard(handIdx)
	if card == nil {
		return
	}
	beforeField := len(g.state.fieldCards)
	g.koikoiPlaceCard(playerIdx, card, fieldIdx)
	handCaptured := len(g.state.fieldCards) <= beforeField
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s (%s)",
		g.playerName(playerIdx), koikoiCardStr(card), koikoiCapturedWord(handCaptured)), []*Card{card})

	// めくり札。
	if len(g.state.drawPile) > 0 {
		drawn := g.state.drawPile[0]
		g.state.drawPile = g.state.drawPile[1:]
		before2 := len(g.state.fieldCards)
		g.koikoiPlaceCard(playerIdx, drawn, -1)
		drawCaptured := len(g.state.fieldCards) <= before2
		g.appendLog(playerIdx, "draw", fmt.Sprintf("%s draws %s (%s)",
			g.playerName(playerIdx), koikoiCardStr(drawn), koikoiCapturedWord(drawCaptured)), []*Card{drawn})
	}

	// 役判定。
	_, total := koikoiEvaluateYaku(player.GetCapturedCards())
	if total > player.GetLastYakuPoints() {
		yakus, _ := koikoiEvaluateYaku(player.GetCapturedCards())
		g.state.pendingYaku = yakus
		g.state.pendingPoints = total
		g.state.phase = KoiKoiPhaseKoiKoiDecision
		g.appendLog(playerIdx, "yaku",
			fmt.Sprintf("%s forms a yaku worth %d", g.playerName(playerIdx), total), nil)
		return
	}
	g.advanceTurn()
}

// PlayerDecide は人間のこいこい決断 (koikoi=true で続行、false で勝負/あがり)。
func (g *KoiKoi) PlayerDecide(koikoi bool) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != KoiKoiPhaseKoiKoiDecision {
		return NewDomainError(ErrWrongPhase, "not in koi-koi decision phase")
	}
	if !g.players[g.state.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyDecision(g.state.currentTurn, koikoi)
	return nil
}

// CpuDecide は CPU のこいこい決断を 1 回進める。
func (g *KoiKoi) CpuDecide() {
	if g.state.gameEndFlag || g.state.phase != KoiKoiPhaseKoiKoiDecision {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() {
		return
	}
	g.applyDecision(g.state.currentTurn, g.chooseCpuDecision(g.state.currentTurn))
}

// applyDecision はこいこい決断を適用する。続行なら基準役点を更新して手番を進め、
// 勝負なら現在の役でラウンドを終える。
func (g *KoiKoi) applyDecision(playerIdx int, koikoi bool) {
	player := g.players[playerIdx]
	if koikoi && player.GetCardsSize() > 0 {
		player.SetCalledKoiKoi(true)
		player.SetLastYakuPoints(g.state.pendingPoints)
		g.state.koikoiCount++
		g.appendLog(playerIdx, "koikoi", fmt.Sprintf("%s calls Koi-Koi", g.playerName(playerIdx)), nil)
		g.state.pendingYaku = nil
		g.state.pendingPoints = 0
		g.state.phase = KoiKoiPhasePlay
		g.advanceTurn()
		return
	}
	// 勝負 (あがり)。手札が無い場合も強制的にここで確定する。
	g.appendLog(playerIdx, "shobu", fmt.Sprintf("%s stops (Shobu)", g.playerName(playerIdx)), nil)
	g.endRound(playerIdx)
}

// advanceTurn は手番を次へ進め、双方の手札が尽きたら引き分けでラウンドを終える。
func (g *KoiKoi) advanceTurn() {
	if g.allHandsEmpty() {
		g.endRound(-1)
		return
	}
	g.state.currentTurn = (g.state.currentTurn + 1) % KoiKoiPlayerCnt
}

// endRound はラウンドを終える。winnerIdx>=0 なら役 × 倍率を加点、-1 は引き分け。
func (g *KoiKoi) endRound(winnerIdx int) {
	g.state.roundWinner = winnerIdx
	g.state.pendingYaku = nil
	g.state.pendingPoints = 0

	multiplier := 1
	if g.state.koikoiCount >= 1 {
		multiplier = 2
	}
	result := &KoiKoiRoundResult{
		Winner:      winnerIdx,
		Multiplier:  multiplier,
		KoikoiCount: g.state.koikoiCount,
	}
	if winnerIdx >= 0 {
		yakus, base := koikoiEvaluateYaku(g.players[winnerIdx].GetCapturedCards())
		total := base * multiplier
		result.Yaku = yakus
		result.BasePoints = base
		result.Total = total
		g.players[winnerIdx].AddScore(total)
		g.appendLog(winnerIdx, "roundWin",
			fmt.Sprintf("%s wins round with %d points (x%d)", g.playerName(winnerIdx), total, multiplier), nil)
	} else {
		result.Yaku = make([]KoiKoiYaku, 0)
		g.appendLog(-1, "draw", "round drawn (no winner)", nil)
	}
	g.state.lastRoundResult = result

	// 終局判定。
	if g.reachedTarget() || g.state.roundNumber >= KoiKoiMaxRounds {
		g.finishGame()
		return
	}
	g.state.phase = KoiKoiPhaseRoundEnd
}

// reachedTarget は目標得点に到達したプレイヤーがいるか。
func (g *KoiKoi) reachedTarget() bool {
	for _, p := range g.players {
		if p.GetScore() >= g.config.TargetScore {
			return true
		}
	}
	return false
}

// finishGame は終局処理: 累計最高点を勝者にする (同点は引き分け -1)。
func (g *KoiKoi) finishGame() {
	best := -1
	bestScore := -1
	tie := false
	for i, p := range g.players {
		s := p.GetScore()
		if s > bestScore {
			best = i
			bestScore = s
			tie = false
		} else if s == bestScore {
			tie = true
		}
	}
	if tie {
		best = -1
	}
	g.state.winner = best
	g.state.gameEndFlag = true
	g.state.phase = KoiKoiPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended (winner %d)", best), nil)
}

// --- CPU AI ---

// chooseCpuPlay は CPU の手札インデックスと捕獲場札を選ぶ。
func (g *KoiKoi) chooseCpuPlay(playerIdx int) (int, int) {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0, -1
	}
	if g.config.CpuDifficulty == KoiKoiCpuDifficultyEasy {
		idx := rand.Intn(size)
		return idx, g.cpuFieldChoice(player.GetCard(idx))
	}
	bestIdx := 0
	// 下限で初期化する。捨て札スコアは -koikoiCardWeight (=-1..-5) で、以前の
	// 初期値 -1 は「カス捨て」の最大値と同値だったため、全札が捨て札の手番では
	// 厳密比較 (>) が一度も成立せず bestIdx が 0 のまま = hand[0] を無条件に捨てる
	// バグになっていた (最安の札を捨てる意図を無効化)。
	bestScore := math.MinInt
	for i := 0; i < size; i++ {
		card := player.GetCard(i)
		matches := g.koikoiFieldMatches(card)
		s := 0
		if len(matches) > 0 {
			// 捕獲価値 = 出す札 + 取れる最良の場札。
			s = koikoiCardWeight(card) + koikoiCardWeight(g.state.fieldCards[g.koikoiBestFieldMatch(matches)])
			if len(matches) >= 3 {
				s += 4
			}
		} else {
			// 捨てる場合は価値の低い札ほど良い (score を負に寄せる)。
			s = -koikoiCardWeight(card)
		}
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx, g.cpuFieldChoice(player.GetCard(bestIdx))
}

// cpuFieldChoice は 2 枚一致時に取る場札を選ぶ (最良札)。一致 0/1/3 は -1 (自動)。
func (g *KoiKoi) cpuFieldChoice(card *Card) int {
	matches := g.koikoiFieldMatches(card)
	if len(matches) == 2 {
		return g.koikoiBestFieldMatch(matches)
	}
	return -1
}

// chooseCpuDecision は CPU がこいこい (true) か勝負 (false) かを決める。
func (g *KoiKoi) chooseCpuDecision(playerIdx int) bool {
	player := g.players[playerIdx]
	total := g.state.pendingPoints
	handsLeft := player.GetCardsSize()
	switch g.config.CpuDifficulty {
	case KoiKoiCpuDifficultyEasy:
		return false // 常に勝負 (保守的)
	case KoiKoiCpuDifficultyHard:
		return total < 7 && handsLeft >= 2
	default: // Normal
		return total < 4 && handsLeft >= 4
	}
}

// --- Hint ---

// GetHint は人間の手番における推奨手を返す。
func (g *KoiKoi) GetHint() *KoiKoiHint {
	if g.state.gameEndFlag {
		return nil
	}
	human := findHumanIdx(g.players)
	if human < 0 || g.state.currentTurn != human {
		return nil
	}
	switch g.state.phase {
	case KoiKoiPhaseKoiKoiDecision:
		kk := 0
		reason := "stop_secure"
		if g.state.pendingPoints < 4 && g.players[human].GetCardsSize() >= 4 {
			kk = 1
			reason = "koikoi_lowyaku"
		}
		return &KoiKoiHint{CardIndex: -1, FieldIndex: -1, KoiKoi: kk, Reason: reason}
	case KoiKoiPhasePlay:
		idx, field := g.chooseCpuPlay(human)
		reason := "capture"
		if len(g.koikoiFieldMatches(g.players[human].GetCard(idx))) == 0 {
			reason = "discard_low"
		}
		return &KoiKoiHint{CardIndex: idx, FieldIndex: field, KoiKoi: -1, Reason: reason}
	default:
		return nil
	}
}

// --- ヘルパー ---

// sortHumanHand は人間の手札を月→インデックス順に並べ替える。
func (g *KoiKoi) sortHumanHand() {
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

func (g *KoiKoi) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return "CPU"
}

func (g *KoiKoi) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// koikoiCardStr は札を "松·光" のように表す (ログ/デバッグ用)。
func koikoiCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	return koikoiMonthKanji(c.GetDesign()) + "·" + koikoiCategoryShort(koikoiInfo(c).category)
}

func koikoiCapturedWord(captured bool) string {
	if captured {
		return "captures"
	}
	return "to field"
}

// koikoiMonthKanji は月番号を月札の代表漢字にする。
func koikoiMonthKanji(month int) string {
	kanji := []string{"?", "松", "梅", "桜", "藤", "菖", "牡", "萩", "芒", "菊", "紅", "柳", "桐"}
	if month >= 1 && month <= KoiKoiMonthCnt {
		return kanji[month]
	}
	return "?"
}

// koikoiCategoryShort は札種の短い日本語表記。
func koikoiCategoryShort(cat KoiKoiCategory) string {
	switch cat {
	case KoiKoiBright:
		return "光"
	case KoiKoiAnimal:
		return "タネ"
	case KoiKoiRibbon:
		return "短"
	default:
		return "カス"
	}
}

// --- 描画用エクスポートアクセサ (adapter/presenter が参照) ---

// KoiKoiCardGlyph は札の手続き描画用グリフ (絵文字) を返す。
func KoiKoiCardGlyph(c *Card) string { return koikoiInfo(c).glyph }

// KoiKoiCardCategory は札種 (光/タネ/短/カス) を返す。
func KoiKoiCardCategory(c *Card) KoiKoiCategory { return koikoiInfo(c).category }

// KoiKoiCardRibbonColor は短冊札の色を返す (短冊でなければ None)。
func KoiKoiCardRibbonColor(c *Card) KoiKoiRibbonColor { return koikoiInfo(c).ribbon }

// KoiKoiCardLabel は札の短い日本語ラベル ("松·光" 等) を返す。
func KoiKoiCardLabel(c *Card) string { return koikoiCardStr(c) }

// --- 状態アクセサ ---

// IsHumanTurn は現在プレイ/決断の手番が人間かどうかを返す。
func (g *KoiKoi) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	if g.state.phase != KoiKoiPhasePlay && g.state.phase != KoiKoiPhaseKoiKoiDecision {
		return false
	}
	return g.players[g.state.currentTurn].GetIsHuman()
}

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *KoiKoi) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *KoiKoi) GetPhase() KoiKoiPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *KoiKoi) SetPhase(p KoiKoiPhase) { g.state.phase = p }

// GetCurrentTurn は現在の手番を返す。
func (g *KoiKoi) GetCurrentTurn() int { return g.state.currentTurn }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *KoiKoi) SetCurrentTurn(idx int) { g.state.currentTurn = idx }

// SetKoikoiCount はこのラウンドのこいこい宣言回数を設定する (テスト用)。
func (g *KoiKoi) SetKoikoiCount(n int) { g.state.koikoiCount = n }

// GetFieldCards は場札を返す。
func (g *KoiKoi) GetFieldCards() []*Card { return g.state.fieldCards }

// SetFieldCards は場札を設定する (テスト用)。
func (g *KoiKoi) SetFieldCards(cards []*Card) { g.state.fieldCards = cards }

// GetRemainingDeck は山札の残り枚数を返す。
func (g *KoiKoi) GetRemainingDeck() int { return len(g.state.drawPile) }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *KoiKoi) GetRoundNumber() int { return g.state.roundNumber }

// GetKoikoiCount はこのラウンドのこいこい宣言回数を返す。
func (g *KoiKoi) GetKoikoiCount() int { return g.state.koikoiCount }

// GetRoundWinner は直近ラウンドの勝者を返す (-1 = 引き分け/未決)。
func (g *KoiKoi) GetRoundWinner() int { return g.state.roundWinner }

// GetLastRoundResult は直近ラウンド結果を返す (nil の場合もある)。
func (g *KoiKoi) GetLastRoundResult() *KoiKoiRoundResult { return g.state.lastRoundResult }

// GetPendingYaku は決断フェーズで表示する成立役を返す。
func (g *KoiKoi) GetPendingYaku() []KoiKoiYaku { return g.state.pendingYaku }

// GetPendingPoints は決断フェーズの役合計点を返す。
func (g *KoiKoi) GetPendingPoints() int { return g.state.pendingPoints }

// GetWinner は終局時の勝者を返す (-1 = 引き分け/未決)。
func (g *KoiKoi) GetWinner() int { return g.state.winner }

// GetPlayerCnt はプレイヤー数を返す。
func (g *KoiKoi) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *KoiKoi) GetPlayer(i int) *KoiKoiPlayer {
	return getPlayer(g.players, i)
}

// GetConfig はローカルルール設定を返す。
func (g *KoiKoi) GetConfig() KoiKoiConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *KoiKoi) SetConfig(cfg KoiKoiConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *KoiKoi) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// GetYaku は指定プレイヤーの現在の取り札で成立している役を返す (UI 補助)。
func (g *KoiKoi) GetYaku(playerIdx int) ([]KoiKoiYaku, int) {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil, 0
	}
	return koikoiEvaluateYaku(g.players[playerIdx].GetCapturedCards())
}

// GetPlayableIndices はプレイフェーズで人間がプレイできる手札インデックス (全札) を返す。
func (g *KoiKoi) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != KoiKoiPhasePlay {
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
func (g *KoiKoi) GetCaptureOptions(playerIdx int) map[int][]int {
	out := make(map[int][]int)
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != KoiKoiPhasePlay {
		return out
	}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if m := g.koikoiFieldMatches(p.GetCard(i)); len(m) > 0 {
			out[i] = m
		}
	}
	return out
}

// --- JSON Serialization ---

// koikoiJSON is the JSON wire format for KoiKoi.
type koikoiJSON struct {
	Players         []*KoiKoiPlayer    `json:"pl"`
	Config          KoiKoiConfig       `json:"cf"`
	Phase           KoiKoiPhase        `json:"ph"`
	CurrentTurn     int                `json:"ct"`
	FieldCards      []*Card            `json:"fd"`
	DrawPile        []*Card            `json:"dp"`
	RoundNumber     int                `json:"rn"`
	KoikoiCount     int                `json:"kc"`
	RoundWinner     int                `json:"rw"`
	GameEndFlag     bool               `json:"ge"`
	Winner          int                `json:"wn"`
	LastRoundResult *KoiKoiRoundResult `json:"lr"`
	PendingYaku     []KoiKoiYaku       `json:"py"`
	PendingPoints   int                `json:"pp"`
	ActionLog       []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *KoiKoi) MarshalJSON() ([]byte, error) {
	return json.Marshal(koikoiJSON{
		Players:         g.players,
		Config:          g.config,
		Phase:           g.state.phase,
		CurrentTurn:     g.state.currentTurn,
		FieldCards:      g.state.fieldCards,
		DrawPile:        g.state.drawPile,
		RoundNumber:     g.state.roundNumber,
		KoikoiCount:     g.state.koikoiCount,
		RoundWinner:     g.state.roundWinner,
		GameEndFlag:     g.state.gameEndFlag,
		Winner:          g.state.winner,
		LastRoundResult: g.state.lastRoundResult,
		PendingYaku:     g.state.pendingYaku,
		PendingPoints:   g.state.pendingPoints,
		ActionLog:       g.state.actionLog,
	})
}

// koikoiMaxSliceLen caps slice sizes during deserialisation to prevent excessive
// memory allocation from malformed input.
const koikoiMaxSliceLen = 1000

// koikoiValidPhase は有効なフェーズかどうか。
func koikoiValidPhase(p KoiKoiPhase) bool {
	return p >= KoiKoiPhasePlay && p <= KoiKoiPhaseGameEnd
}

// koikoiValidateCards は復元したカードスライスに nil や範囲外の月/インデックスが
// 無いか検証する。
func koikoiValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("koikoi: nil card in state")
		}
		m, i := c.GetDesign(), c.GetValue()
		if m < 1 || m > KoiKoiMonthCnt || i < 1 || i > KoiKoiCardsPerMonth {
			return fmt.Errorf("koikoi: card out of range (month %d, index %d)", m, i)
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否する。
func (g *KoiKoi) UnmarshalJSON(data []byte) error {
	var j koikoiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > koikoiMaxSliceLen || len(j.FieldCards) > koikoiMaxSliceLen ||
		len(j.DrawPile) > koikoiMaxSliceLen || len(j.ActionLog) > koikoiMaxSliceLen ||
		len(j.PendingYaku) > koikoiMaxSliceLen {
		return fmt.Errorf("koikoi: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("koikoi: invalid config: %w", err)
	}
	if len(j.Players) != KoiKoiPlayerCnt {
		return fmt.Errorf("koikoi: invalid player count %d, expected %d", len(j.Players), KoiKoiPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("koikoi: nil player in state")
		}
	}
	if !koikoiValidPhase(j.Phase) {
		return fmt.Errorf("koikoi: invalid phase %d", j.Phase)
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= len(j.Players) {
		return fmt.Errorf("koikoi: current turn out of range")
	}
	if j.RoundWinner < -1 || j.RoundWinner >= len(j.Players) {
		return fmt.Errorf("koikoi: round winner out of range")
	}
	if j.Winner < -1 || j.Winner >= len(j.Players) {
		return fmt.Errorf("koikoi: winner out of range")
	}
	if err := koikoiValidateCards(j.FieldCards); err != nil {
		return err
	}
	if err := koikoiValidateCards(j.DrawPile); err != nil {
		return err
	}

	g.players = j.Players
	g.config = j.Config
	g.state = koikoiState{
		phase:           j.Phase,
		currentTurn:     j.CurrentTurn,
		fieldCards:      j.FieldCards,
		drawPile:        j.DrawPile,
		roundNumber:     j.RoundNumber,
		koikoiCount:     j.KoikoiCount,
		roundWinner:     j.RoundWinner,
		gameEndFlag:     j.GameEndFlag,
		winner:          j.Winner,
		lastRoundResult: j.LastRoundResult,
		pendingYaku:     j.PendingYaku,
		pendingPoints:   j.PendingPoints,
		actionLogBase:   actionLogBase{actionLog: j.ActionLog},
	}
	if g.state.fieldCards == nil {
		g.state.fieldCards = make([]*Card, 0)
	}
	if g.state.drawPile == nil {
		g.state.drawPile = make([]*Card, 0)
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
