//go:build !js || !wasm || extra

// Package domain — 八八 (Hachi-Hachi / はちはち) のドメインモデル。
//
// 八八 は花札 (hanafuda) 48 枚 (12 か月 × 4 枚) を用いる 3 人用 (人間 vs CPU×2) の
// フィッシング系ゲーム。手札から 1 枚出して場の同月札を捕獲し、山札を 1 枚めくって
// 同様に捕獲する。全員の手札が尽きたら取り札の「素点」を数え、基準点 88 との差で
// 精算する。名前の由来はこの 88 点 (全 48 枚の素点合計 264 の 1/3)。
//
// # 札の符号化 (ADR-0033)
//
//   - design = 月 (1..12)
//   - value  = 月内インデックス (1..4)
//
// 花札には専用 PNG アートが無いため、各札は手続き的に描画される (glyph/label/color/
// deck="hanafuda")。カードの正体 (光/タネ/短冊/カス、短冊の色) は (月, インデックス) の
// 対応表 hachihachiCardTable から引く (こいこい/ゴーストップと同一のデッキ符号化)。
//
// # 素点 (card-point values)
//
//	光 (Bright)  = 20 点
//	タネ (Animal) = 10 点
//	短冊 (Ribbon) = 5 点
//	カス (Chaff)  = 1 点
//
// 全 48 枚の素点合計 = 5×20 + 9×10 + 10×5 + 24×1 = 264。3 人で割ると 264/3 = 88 が
// 基準点。各プレイヤーのラウンド差分 = (素点 + 出来役ボーナス − 88)。素点の精算部分
// Σ(素点 − 88) = 264 − 264 = 0 で厳密なゼロ和。
//
// # ターン進行
//
// 手番プレイヤーは手札を 1 枚出す。場に同月の札があれば両方を取り札にする (2 枚一致なら
// どちらを取るか選択、3 枚一致=場に既に 3 枚 → すべて取る)。一致が無ければ場に置く。
// 続いて山札の一番上をめくり、同様に同月捕獲 / 場に追加する。手番を次のプレイヤーへ回す。
//
// # ラウンド終了 / 精算
//
// 全員の手札が尽きたらラウンド終了。場に残った札は「最後に捕獲したプレイヤー」に
// まとめて渡す (これで常に 48 枚全てが分配され、素点の精算が厳密にゼロ和になる)。
// 各プレイヤーの差分 = 素点 + 出来役ボーナス − 88 を累計得点へ加算する。
// TargetRounds ラウンド戦った後、累計最高得点のプレイヤーが勝者 (同点は引き分け)。
//
// # 簡略化 (本実装の MVP)
//
//   - 親/子の祝儀 (しょうぎ) 倍加や親流し等のニュアンスは未実装。
//   - 出来役ボーナスは達成者に加点する追加報酬で、素点精算 (ゼロ和) とは別枠。
//     出来役が出た分だけ全体の合計はプラスになり得る (ドキュメント化した簡略ルール)。
//   - 手役 (配札時の役) は未実装。
package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// HachiHachiPlayerCnt は八八のプレイヤー数 (固定 3)。
const HachiHachiPlayerCnt = 3

// HachiHachiHandSize は各プレイヤーへ配る手札枚数。
const HachiHachiHandSize = 7

// HachiHachiFieldSize はゲーム開始時に場へ置くカード枚数。
const HachiHachiFieldSize = 6

// HachiHachiMonthCnt は月数 (= design の最大値)。
const HachiHachiMonthCnt = 12

// HachiHachiCardsPerMonth は 1 月あたりの札数 (= value の最大値)。
const HachiHachiCardsPerMonth = 4

// HachiHachiBaseline は精算の基準点 (全 48 枚の素点合計 264 ÷ 3 人 = 88)。
const HachiHachiBaseline = 88

// HachiHachiCategory は花札の札種。
type HachiHachiCategory int

// 札種定数
const (
	// HachiHachiChaff カス (最下位の札)
	HachiHachiChaff HachiHachiCategory = iota
	// HachiHachiRibbon 短冊 (たんざく)
	HachiHachiRibbon
	// HachiHachiAnimal タネ (動物/種)
	HachiHachiAnimal
	// HachiHachiBright 光 (最上位の札)
	HachiHachiBright
)

// HachiHachiRibbonColor は短冊の色。
type HachiHachiRibbonColor int

// 短冊色定数
const (
	// HachiHachiRibbonNone 短冊ではない
	HachiHachiRibbonNone HachiHachiRibbonColor = iota
	// HachiHachiRibbonRedPoetry 赤短 (書き入りの赤短冊: 1/2/3 月)
	HachiHachiRibbonRedPoetry
	// HachiHachiRibbonBlue 青短 (6/9/10 月)
	HachiHachiRibbonBlue
	// HachiHachiRibbonPlainRed 赤短冊 (書き無し: 4/5/7/11 月)
	HachiHachiRibbonPlainRed
)

// hachihachiCardInfo は 1 枚の花札の正体。
type hachihachiCardInfo struct {
	category HachiHachiCategory
	ribbon   HachiHachiRibbonColor
	glyph    string // 手続き描画用の絵文字
	name     string // 短い識別名 (英語、デバッグ/ラベル補助)
}

// hachihachiCardTable[month][index] (month 1..12, index 1..4) が札の正体を返す。
// インデックス 0 は未使用。こいこい/ゴーストップと同一の canonical な割り当て。
var hachihachiCardTable = [HachiHachiMonthCnt + 1][HachiHachiCardsPerMonth + 1]hachihachiCardInfo{
	1: { // 松 (松に鶴)
		1: {HachiHachiBright, HachiHachiRibbonNone, "🦢", "Crane"},
		2: {HachiHachiRibbon, HachiHachiRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	2: { // 梅 (梅に鶯)
		1: {HachiHachiAnimal, HachiHachiRibbonNone, "🐦", "Warbler"},
		2: {HachiHachiRibbon, HachiHachiRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	3: { // 桜 (桜に幕)
		1: {HachiHachiBright, HachiHachiRibbonNone, "🌸", "Curtain"},
		2: {HachiHachiRibbon, HachiHachiRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	4: { // 藤 (藤に不如帰)
		1: {HachiHachiAnimal, HachiHachiRibbonNone, "🐦", "Cuckoo"},
		2: {HachiHachiRibbon, HachiHachiRibbonPlainRed, "🎴", "RedRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	5: { // 菖蒲 (菖蒲に八橋)
		1: {HachiHachiAnimal, HachiHachiRibbonNone, "🌉", "Bridge"},
		2: {HachiHachiRibbon, HachiHachiRibbonPlainRed, "🎴", "RedRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	6: { // 牡丹 (牡丹に蝶)
		1: {HachiHachiAnimal, HachiHachiRibbonNone, "🦋", "Butterfly"},
		2: {HachiHachiRibbon, HachiHachiRibbonBlue, "🎴", "BlueRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	7: { // 萩 (萩に猪)
		1: {HachiHachiAnimal, HachiHachiRibbonNone, "🐗", "Boar"},
		2: {HachiHachiRibbon, HachiHachiRibbonPlainRed, "🎴", "RedRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	8: { // 芒 (芒に月 / 芒に雁)
		1: {HachiHachiBright, HachiHachiRibbonNone, "🌕", "Moon"},
		2: {HachiHachiAnimal, HachiHachiRibbonNone, "🦆", "Geese"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	9: { // 菊 (菊に盃)
		1: {HachiHachiAnimal, HachiHachiRibbonNone, "🍶", "SakeCup"},
		2: {HachiHachiRibbon, HachiHachiRibbonBlue, "🎴", "BlueRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	10: { // 紅葉 (紅葉に鹿)
		1: {HachiHachiAnimal, HachiHachiRibbonNone, "🦌", "Deer"},
		2: {HachiHachiRibbon, HachiHachiRibbonBlue, "🎴", "BlueRibbon"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
	11: { // 柳 (柳に小野道風 / 燕 / 短冊 / 雷)
		1: {HachiHachiBright, HachiHachiRibbonNone, "☂️", "Rainman"},
		2: {HachiHachiAnimal, HachiHachiRibbonNone, "🐦", "Swallow"},
		3: {HachiHachiRibbon, HachiHachiRibbonPlainRed, "🎴", "RedRibbon"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "⚡", "Lightning"},
	},
	12: { // 桐 (桐に鳳凰)
		1: {HachiHachiBright, HachiHachiRibbonNone, "🦅", "Phoenix"},
		2: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		3: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
		4: {HachiHachiChaff, HachiHachiRibbonNone, "🍂", "Chaff"},
	},
}

// hachihachiInfo は札の正体を返す。範囲外の札はカス相当を返す (防御的)。
func hachihachiInfo(c *Card) hachihachiCardInfo {
	if c == nil {
		return hachihachiCardInfo{category: HachiHachiChaff, glyph: "🍂", name: "Chaff"}
	}
	m, i := c.GetDesign(), c.GetValue()
	if m < 1 || m > HachiHachiMonthCnt || i < 1 || i > HachiHachiCardsPerMonth {
		return hachihachiCardInfo{category: HachiHachiChaff, glyph: "🍂", name: "Chaff"}
	}
	return hachihachiCardTable[m][i]
}

// hachihachiSameMonth は 2 枚が同月かどうか。
func hachihachiSameMonth(a, b *Card) bool {
	return a != nil && b != nil && a.GetDesign() == b.GetDesign()
}

// --- 素点 (card-point values) ---
const (
	hachihachiPointsBright = 20 // 光
	hachihachiPointsAnimal = 10 // タネ
	hachihachiPointsRibbon = 5  // 短冊
	hachihachiPointsChaff  = 1  // カス
)

// hachihachiCardPoints は 1 枚の素点を返す。
func hachihachiCardPoints(c *Card) int {
	switch hachihachiInfo(c).category {
	case HachiHachiBright:
		return hachihachiPointsBright
	case HachiHachiAnimal:
		return hachihachiPointsAnimal
	case HachiHachiRibbon:
		return hachihachiPointsRibbon
	default:
		return hachihachiPointsChaff
	}
}

// --- 出来役ボーナス (素点精算とは別枠の追加報酬) ---
const (
	hachihachiBonusGoko        = 100 // 五光 (光 5 枚)
	hachihachiBonusShiko       = 80  // 四光 (光 4 枚, 雨札なし)
	hachihachiBonusAmeShiko    = 60  // 雨四光 (光 4 枚, 雨札あり)
	hachihachiBonusSanko       = 40  // 三光 (光 3 枚, 雨札なし)
	hachihachiBonusInoshikacho = 50  // 猪鹿蝶
	hachihachiBonusAkatan      = 40  // 赤短
	hachihachiBonusAotan       = 40  // 青短
)

// HachiHachiYaku は成立した 1 出来役。
type HachiHachiYaku struct {
	Key    string `json:"key"`    // 役キー (例 "goko")
	Points int    `json:"points"` // ボーナス点
}

// HachiHachiEvaluateScore は取り札から素点合計・成立出来役・ボーナス合計を返す純粋関数。
// 光の役 (五光/雨四光/四光/三光) は排他で最上位のみ採用する。猪鹿蝶/赤短/青短は独立に加点。
func HachiHachiEvaluateScore(captured []*Card) (rawScore int, yaku []HachiHachiYaku, bonus int) {
	brights := 0
	hasRain := false
	hasBoar, hasDeer, hasButterfly := false, false, false
	redPoetry, blue := 0, 0

	for _, c := range captured {
		if c == nil {
			continue
		}
		rawScore += hachihachiCardPoints(c)
		info := hachihachiInfo(c)
		month := c.GetDesign()
		switch info.category {
		case HachiHachiBright:
			brights++
			if month == 11 {
				hasRain = true
			}
		case HachiHachiAnimal:
			switch month {
			case 7:
				hasBoar = true
			case 10:
				hasDeer = true
			case 6:
				hasButterfly = true
			}
		case HachiHachiRibbon:
			switch info.ribbon {
			case HachiHachiRibbonRedPoetry:
				redPoetry++
			case HachiHachiRibbonBlue:
				blue++
			}
		}
	}

	yaku = make([]HachiHachiYaku, 0, 4)
	add := func(key string, pts int) { yaku = append(yaku, HachiHachiYaku{Key: key, Points: pts}) }

	// 光 (排他)。
	switch {
	case brights >= 5:
		add("goko", hachihachiBonusGoko)
	case brights == 4 && hasRain:
		add("ameshiko", hachihachiBonusAmeShiko)
	case brights == 4:
		add("shiko", hachihachiBonusShiko)
	case brights == 3 && !hasRain:
		add("sanko", hachihachiBonusSanko)
	}
	// 猪鹿蝶。
	if hasBoar && hasDeer && hasButterfly {
		add("inoshikacho", hachihachiBonusInoshikacho)
	}
	// 赤短 / 青短 (それぞれ固有の 3 枚。>=3 は全数一致を意味する)。
	if redPoetry >= 3 {
		add("akatan", hachihachiBonusAkatan)
	}
	if blue >= 3 {
		add("aotan", hachihachiBonusAotan)
	}

	for _, y := range yaku {
		bonus += y.Points
	}
	return rawScore, yaku, bonus
}

// HachiHachiPhase はゲームフェーズ。
type HachiHachiPhase int

// Hachi-Hachi のフェーズ定数
const (
	// HachiHachiPhasePlay プレイ中 (手札を 1 枚出す)
	HachiHachiPhasePlay HachiHachiPhase = 0
	// HachiHachiPhaseRoundEnd ラウンド終了 (結果表示。次ラウンド待ち)
	HachiHachiPhaseRoundEnd HachiHachiPhase = 1
	// HachiHachiPhaseGameEnd 終局
	HachiHachiPhaseGameEnd HachiHachiPhase = 2
)

// HachiHachiResult は人間視点のゲーム結果。
type HachiHachiResult int

// Hachi-Hachi の結果定数
const (
	// HachiHachiResultNone 未決 (対戦中)
	HachiHachiResultNone HachiHachiResult = iota
	// HachiHachiResultWin 人間の勝ち
	HachiHachiResultWin
	// HachiHachiResultLose 人間の負け
	HachiHachiResultLose
	// HachiHachiResultDraw 引き分け
	HachiHachiResultDraw
)

// HachiHachiPlayerScore は 1 プレイヤーのラウンド精算内訳。
type HachiHachiPlayerScore struct {
	PlayerIdx int              `json:"playerIdx"` // プレイヤーインデックス
	RawScore  int              `json:"rawScore"`  // 素点合計
	Yaku      []HachiHachiYaku `json:"yaku"`      // 成立出来役
	Bonus     int              `json:"bonus"`     // 出来役ボーナス合計
	Delta     int              `json:"delta"`     // 符号付き差分 (素点+ボーナス−88)
}

// HachiHachiRoundResult は 1 ラウンドの結果。
type HachiHachiRoundResult struct {
	Scores []HachiHachiPlayerScore `json:"scores"` // 各プレイヤーの内訳
	Best   int                     `json:"best"`   // このラウンド最高差分のプレイヤー (-1=引き分け)
}

// HachiHachiHint はヒント情報。
type HachiHachiHint struct {
	CardIndex  int    `json:"cardIndex"`  // 推奨手札インデックス (-1 = なし)
	FieldIndex int    `json:"fieldIndex"` // 推奨捕獲場札インデックス (-1 = なし/自動)
	Reason     string `json:"reason"`     // 理由キー
}

// hachihachiState はゲーム進行状態。
type hachihachiState struct {
	phase           HachiHachiPhase
	currentTurn     int
	fieldCards      []*Card
	drawPile        []*Card
	roundNumber     int
	lastCapturer    int // 直近に捕獲したプレイヤー (ラウンド終了時の場札掃き寄せ先)
	gameEndFlag     bool
	winner          int // 終局時の勝者 (-1 = 引き分け)
	lastRoundResult *HachiHachiRoundResult
	actionLogBase
}

// HachiHachi は八八ゲームの状態を保持する集約ルート。
type HachiHachi struct {
	players []*HachiHachiPlayer
	config  HachiHachiConfig
	state   hachihachiState
}

// NewHachiHachi はコンストラクタ。
func NewHachiHachi(players []*HachiHachiPlayer, config HachiHachiConfig) *HachiHachi {
	return &HachiHachi{
		players: players,
		config:  config,
		state: hachihachiState{
			phase:        HachiHachiPhasePlay,
			lastCapturer: -1,
			winner:       -1,
		},
	}
}

// NewDefaultHachiHachi は標準の 3 人構成 (1 human + 2 CPU) で HachiHachi を生成する。
func NewDefaultHachiHachi() *HachiHachi {
	players := make([]*HachiHachiPlayer, HachiHachiPlayerCnt)
	players[0] = NewHachiHachiPlayer(true)
	players[1] = NewHachiHachiPlayer(false)
	players[2] = NewHachiHachiPlayer(false)
	return NewHachiHachi(players, DefaultHachiHachiConfig())
}

// buildHachiHachiDeck は花札 48 枚を design=月(1..12)/value=index(1..4) で直接生成する。
func buildHachiHachiDeck() []*Card {
	deck := make([]*Card, 0, HachiHachiMonthCnt*HachiHachiCardsPerMonth)
	for m := 1; m <= HachiHachiMonthCnt; m++ {
		for i := 1; i <= HachiHachiCardsPerMonth; i++ {
			deck = append(deck, NewCard(m, i, false))
		}
	}
	return deck
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。
func (g *HachiHachi) Reset() {
	for _, p := range g.players {
		p.Reset()
		p.ResetRound()
		p.ResetScore()
	}
	g.state = hachihachiState{
		phase:         HachiHachiPhasePlay,
		lastCapturer:  -1,
		winner:        -1,
		roundNumber:   1,
		actionLogBase: actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound はラウンド終了後に次ラウンドを開始する。
func (g *HachiHachi) NextRound() {
	if g.state.gameEndFlag || g.state.phase != HachiHachiPhaseRoundEnd {
		return
	}
	g.state.roundNumber++
	g.startRound()
}

// startRound はデッキ生成・配札・場札配置を行い、プレイフェーズを開始する。
// ラウンドごとに先手を交代する ((roundNumber-1) % 3)。
func (g *HachiHachi) startRound() {
	deck := buildHachiHachiDeck()
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	for _, p := range g.players {
		p.Reset()
		p.ResetRound()
	}
	g.state.fieldCards = make([]*Card, 0, HachiHachiFieldSize)
	g.state.lastCapturer = -1
	g.state.phase = HachiHachiPhasePlay
	g.state.currentTurn = (g.state.roundNumber - 1) % HachiHachiPlayerCnt

	// 交互配り: 各プレイヤーへ 7 枚、場へ 6 枚、残り 21 枚を山札に。
	pos := 0
	for k := 0; k < HachiHachiHandSize; k++ {
		for _, p := range g.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	for k := 0; k < HachiHachiFieldSize; k++ {
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
func (g *HachiHachi) allHandsEmpty() bool {
	for _, p := range g.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// --- 捕獲ロジック ---

// hachihachiFieldMatches は場札のうち card と同月のインデックスを返す。
func (g *HachiHachi) hachihachiFieldMatches(card *Card) []int {
	var out []int
	for i, c := range g.state.fieldCards {
		if hachihachiSameMonth(c, card) {
			out = append(out, i)
		}
	}
	return out
}

// hachihachiBestFieldMatch は複数一致のうち最も価値の高い場札インデックスを返す。
func (g *HachiHachi) hachihachiBestFieldMatch(matches []int) int {
	best := -1
	bestVal := -1
	for _, idx := range matches {
		if idx < 0 || idx >= len(g.state.fieldCards) {
			continue
		}
		v := hachihachiCardPoints(g.state.fieldCards[idx])
		if v > bestVal {
			bestVal = v
			best = idx
		}
	}
	return best
}

// hachihachiPlaceCard は 1 枚 (手札またはめくり札) を場と突き合わせて解決する。
//   - 一致なし: 場に置く。
//   - 一致 1 枚: その札と共に捕獲。
//   - 一致 2 枚: chosen が一致札なら chosen を、そうでなければ最良の 1 枚を捕獲。
//   - 一致 3 枚: すべて捕獲 (場の同月 4 枚目)。
//
// 捕獲した場合は lastCapturer を playerIdx に更新する。
func (g *HachiHachi) hachihachiPlaceCard(playerIdx int, card *Card, chosen int) {
	matches := g.hachihachiFieldMatches(card)
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
			sel = g.hachihachiBestFieldMatch(matches)
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
	g.state.lastCapturer = playerIdx
}

// removeFieldByIndex は降順に並べ替えてから場札を削除する。
func (g *HachiHachi) removeFieldByIndex(idxs []int) {
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
func (g *HachiHachi) PlayerPlay(handIdx, fieldIdx int) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != HachiHachiPhasePlay {
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
	if fieldIdx >= 0 {
		if fieldIdx >= len(g.state.fieldCards) || !hachihachiSameMonth(g.state.fieldCards[fieldIdx], card) {
			return NewDomainError(ErrInvalidPlay, "chosen field card does not match the played card's month")
		}
	}
	g.applyTurn(g.state.currentTurn, handIdx, fieldIdx)
	return nil
}

// CpuPlay は CPU のプレイ手番を 1 回進める。
func (g *HachiHachi) CpuPlay() {
	if g.state.gameEndFlag || g.state.phase != HachiHachiPhasePlay {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() || p.GetCardsSize() == 0 {
		return
	}
	handIdx, fieldIdx := g.chooseCpuPlay(g.state.currentTurn)
	g.applyTurn(g.state.currentTurn, handIdx, fieldIdx)
}

// applyTurn は手札を出し→めくり札を処理→手番を進める共通処理。
func (g *HachiHachi) applyTurn(playerIdx, handIdx, fieldIdx int) {
	player := g.players[playerIdx]
	card := player.RemoveCard(handIdx)
	if card == nil {
		return
	}
	beforeField := len(g.state.fieldCards)
	g.hachihachiPlaceCard(playerIdx, card, fieldIdx)
	handCaptured := len(g.state.fieldCards) <= beforeField
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s (%s)",
		g.playerName(playerIdx), hachihachiCardStr(card), hachihachiCapturedWord(handCaptured)), []*Card{card})

	// めくり札。
	if len(g.state.drawPile) > 0 {
		drawn := g.state.drawPile[0]
		g.state.drawPile = g.state.drawPile[1:]
		before2 := len(g.state.fieldCards)
		g.hachihachiPlaceCard(playerIdx, drawn, -1)
		drawCaptured := len(g.state.fieldCards) <= before2
		g.appendLog(playerIdx, "draw", fmt.Sprintf("%s draws %s (%s)",
			g.playerName(playerIdx), hachihachiCardStr(drawn), hachihachiCapturedWord(drawCaptured)), []*Card{drawn})
	}

	g.advanceTurn()
}

// advanceTurn は手番を次へ進め、全員の手札が尽きたらラウンドを精算する。
func (g *HachiHachi) advanceTurn() {
	if g.allHandsEmpty() {
		g.endRound()
		return
	}
	g.state.currentTurn = (g.state.currentTurn + 1) % HachiHachiPlayerCnt
}

// endRound はラウンドを精算する。場に残った札を最後の捕獲者へ掃き寄せ、各プレイヤーの
// 差分 (素点 + 出来役ボーナス − 88) を累計得点に加算する。
func (g *HachiHachi) endRound() {
	g.sweepFieldToLastCapturer()

	scores := make([]HachiHachiPlayerScore, 0, len(g.players))
	best, bestDelta := -1, math.MinInt
	tie := false
	for i, p := range g.players {
		raw, yaku, bonus := HachiHachiEvaluateScore(p.GetCapturedCards())
		delta := raw + bonus - HachiHachiBaseline
		p.AddScore(delta)
		p.SetRoundDelta(delta)
		scores = append(scores, HachiHachiPlayerScore{
			PlayerIdx: i,
			RawScore:  raw,
			Yaku:      yaku,
			Bonus:     bonus,
			Delta:     delta,
		})
		switch {
		case delta > bestDelta:
			bestDelta = delta
			best = i
			tie = false
		case delta == bestDelta:
			tie = true
		}
	}
	if tie {
		best = -1
	}
	g.state.lastRoundResult = &HachiHachiRoundResult{Scores: scores, Best: best}
	g.appendLog(best, "roundEnd",
		fmt.Sprintf("round %d settled (best %d)", g.state.roundNumber, best), nil)

	if g.state.roundNumber >= g.config.TargetRounds {
		g.finishGame()
		return
	}
	g.state.phase = HachiHachiPhaseRoundEnd
}

// sweepFieldToLastCapturer は場に残った札を最後の捕獲者へ渡す。捕獲が一度も無かった
// (lastCapturer < 0) 場合は現在の手番プレイヤーへ渡す。これで常に 48 枚全てが分配され、
// 素点精算が厳密にゼロ和になる。
func (g *HachiHachi) sweepFieldToLastCapturer() {
	if len(g.state.fieldCards) == 0 {
		return
	}
	target := g.state.lastCapturer
	if target < 0 || target >= len(g.players) {
		target = g.state.currentTurn
	}
	g.players[target].AddCaptured(append([]*Card(nil), g.state.fieldCards...))
	g.appendLog(target, "sweep",
		fmt.Sprintf("%s sweeps %d leftover field card(s)", g.playerName(target), len(g.state.fieldCards)),
		append([]*Card(nil), g.state.fieldCards...))
	g.state.fieldCards = make([]*Card, 0)
}

// finishGame は終局処理: 累計最高点を勝者にする (同点は引き分け -1)。
func (g *HachiHachi) finishGame() {
	best := -1
	bestScore := math.MinInt
	tie := false
	for i, p := range g.players {
		s := p.GetScore()
		switch {
		case s > bestScore:
			best = i
			bestScore = s
			tie = false
		case s == bestScore:
			tie = true
		}
	}
	if tie {
		best = -1
	}
	g.state.winner = best
	g.state.gameEndFlag = true
	g.state.phase = HachiHachiPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended (winner %d)", best), nil)
}

// --- CPU AI ---

// chooseCpuPlay は CPU の手札インデックスと捕獲場札を選ぶ。
func (g *HachiHachi) chooseCpuPlay(playerIdx int) (int, int) {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0, -1
	}
	if g.config.CpuDifficulty == HachiHachiCpuDifficultyEasy {
		idx := rand.Intn(size)
		return idx, g.cpuFieldChoice(player.GetCard(idx))
	}
	bestIdx := 0
	bestScore := math.MinInt
	for i := 0; i < size; i++ {
		card := player.GetCard(i)
		matches := g.hachihachiFieldMatches(card)
		s := 0
		if len(matches) > 0 {
			// 捕獲価値 = 出す札 + 取れる最良の場札の素点。
			bestMatch := g.hachihachiBestFieldMatch(matches)
			if bestMatch >= 0 {
				s = hachihachiCardPoints(card) + hachihachiCardPoints(g.state.fieldCards[bestMatch])
			}
			if len(matches) >= 3 {
				s += 20
			}
		} else {
			// 捨てる場合は素点の低い札ほど良い。
			s = -hachihachiCardPoints(card)
		}
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx, g.cpuFieldChoice(player.GetCard(bestIdx))
}

// cpuFieldChoice は 2 枚一致時に取る場札を選ぶ (最良札)。一致 0/1/3 は -1 (自動)。
func (g *HachiHachi) cpuFieldChoice(card *Card) int {
	matches := g.hachihachiFieldMatches(card)
	if len(matches) == 2 {
		return g.hachihachiBestFieldMatch(matches)
	}
	return -1
}

// --- Hint ---

// GetHint は人間の手番における推奨手を返す。
func (g *HachiHachi) GetHint() *HachiHachiHint {
	if g.state.gameEndFlag {
		return nil
	}
	human := findHumanIdx(g.players)
	if human < 0 || g.state.currentTurn != human || g.state.phase != HachiHachiPhasePlay {
		return nil
	}
	if g.players[human].GetCardsSize() == 0 {
		return nil
	}
	idx, field := g.chooseCpuPlay(human)
	reason := "capture"
	if len(g.hachihachiFieldMatches(g.players[human].GetCard(idx))) == 0 {
		reason = "discard_low"
	}
	return &HachiHachiHint{CardIndex: idx, FieldIndex: field, Reason: reason}
}

// --- ヘルパー ---

// sortHumanHand は人間の手札を月→インデックス順に並べ替える。
func (g *HachiHachi) sortHumanHand() {
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

func (g *HachiHachi) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU%d", idx)
}

func (g *HachiHachi) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// hachihachiCardStr は札を "松·光" のように表す (ログ/デバッグ用)。
func hachihachiCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	return hachihachiMonthKanji(c.GetDesign()) + "·" + hachihachiCategoryShort(hachihachiInfo(c).category)
}

func hachihachiCapturedWord(captured bool) string {
	if captured {
		return "captures"
	}
	return "to field"
}

// hachihachiMonthKanji は月番号を月札の代表漢字にする。
func hachihachiMonthKanji(month int) string {
	kanji := []string{"?", "松", "梅", "桜", "藤", "菖", "牡", "萩", "芒", "菊", "紅", "柳", "桐"}
	if month >= 1 && month <= HachiHachiMonthCnt {
		return kanji[month]
	}
	return "?"
}

// hachihachiCategoryShort は札種の短い日本語表記。
func hachihachiCategoryShort(cat HachiHachiCategory) string {
	switch cat {
	case HachiHachiBright:
		return "光"
	case HachiHachiAnimal:
		return "タネ"
	case HachiHachiRibbon:
		return "短"
	default:
		return "カス"
	}
}

// --- 描画用エクスポートアクセサ (adapter/presenter が参照) ---

// HachiHachiCardGlyph は札の手続き描画用グリフ (絵文字) を返す。
func HachiHachiCardGlyph(c *Card) string { return hachihachiInfo(c).glyph }

// HachiHachiCardCategory は札種 (光/タネ/短/カス) を返す。
func HachiHachiCardCategory(c *Card) HachiHachiCategory { return hachihachiInfo(c).category }

// HachiHachiCardRibbonColor は短冊札の色を返す (短冊でなければ None)。
func HachiHachiCardRibbonColor(c *Card) HachiHachiRibbonColor { return hachihachiInfo(c).ribbon }

// HachiHachiCardLabel は札の短い日本語ラベル ("松·光" 等) を返す。
func HachiHachiCardLabel(c *Card) string { return hachihachiCardStr(c) }

// HachiHachiCardPoints は札の素点を返す。
func HachiHachiCardPoints(c *Card) int { return hachihachiCardPoints(c) }

// --- 状態アクセサ ---

// IsHumanTurn は現在プレイの手番が人間かどうかを返す。
func (g *HachiHachi) IsHumanTurn() bool {
	if g.state.gameEndFlag || g.state.phase != HachiHachiPhasePlay {
		return false
	}
	return g.players[g.state.currentTurn].GetIsHuman()
}

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *HachiHachi) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *HachiHachi) GetPhase() HachiHachiPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *HachiHachi) SetPhase(p HachiHachiPhase) { g.state.phase = p }

// GetCurrentTurn は現在の手番を返す。
func (g *HachiHachi) GetCurrentTurn() int { return g.state.currentTurn }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *HachiHachi) SetCurrentTurn(idx int) { g.state.currentTurn = idx }

// GetFieldCards は場札を返す。
func (g *HachiHachi) GetFieldCards() []*Card { return g.state.fieldCards }

// SetFieldCards は場札を設定する (テスト用)。
func (g *HachiHachi) SetFieldCards(cards []*Card) { g.state.fieldCards = cards }

// GetRemainingDeck は山札の残り枚数を返す。
func (g *HachiHachi) GetRemainingDeck() int { return len(g.state.drawPile) }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *HachiHachi) GetRoundNumber() int { return g.state.roundNumber }

// GetLastRoundResult は直近ラウンド結果を返す (nil の場合もある)。
func (g *HachiHachi) GetLastRoundResult() *HachiHachiRoundResult { return g.state.lastRoundResult }

// GetWinner は終局時の勝者を返す (-1 = 引き分け/未決)。
func (g *HachiHachi) GetWinner() int { return g.state.winner }

// GetResult は人間視点のゲーム結果を返す。
func (g *HachiHachi) GetResult() HachiHachiResult {
	if !g.state.gameEndFlag {
		return HachiHachiResultNone
	}
	human := findHumanIdx(g.players)
	if g.state.winner < 0 {
		return HachiHachiResultDraw
	}
	if g.state.winner == human {
		return HachiHachiResultWin
	}
	return HachiHachiResultLose
}

// GetPlayerCnt はプレイヤー数を返す。
func (g *HachiHachi) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *HachiHachi) GetPlayer(i int) *HachiHachiPlayer {
	return getPlayer(g.players, i)
}

// GetConfig はローカルルール設定を返す。
func (g *HachiHachi) GetConfig() HachiHachiConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *HachiHachi) SetConfig(cfg HachiHachiConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *HachiHachi) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// GetYaku は指定プレイヤーの現在の取り札で成立している出来役と素点合計を返す (UI 補助)。
func (g *HachiHachi) GetYaku(playerIdx int) ([]HachiHachiYaku, int) {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil, 0
	}
	raw, yaku, _ := HachiHachiEvaluateScore(g.players[playerIdx].GetCapturedCards())
	return yaku, raw
}

// GetPlayableIndices はプレイフェーズで人間がプレイできる手札インデックス (全札) を返す。
func (g *HachiHachi) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != HachiHachiPhasePlay {
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
func (g *HachiHachi) GetCaptureOptions(playerIdx int) map[int][]int {
	out := make(map[int][]int)
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != HachiHachiPhasePlay {
		return out
	}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if m := g.hachihachiFieldMatches(p.GetCard(i)); len(m) > 0 {
			out[i] = m
		}
	}
	return out
}

// --- JSON Serialization ---

// hachihachiJSON is the JSON wire format for HachiHachi.
type hachihachiJSON struct {
	Players         []*HachiHachiPlayer    `json:"pl"`
	Config          HachiHachiConfig       `json:"cf"`
	Phase           HachiHachiPhase        `json:"ph"`
	CurrentTurn     int                    `json:"ct"`
	FieldCards      []*Card                `json:"fd"`
	DrawPile        []*Card                `json:"dp"`
	RoundNumber     int                    `json:"rn"`
	LastCapturer    int                    `json:"lc"`
	GameEndFlag     bool                   `json:"ge"`
	Winner          int                    `json:"wn"`
	LastRoundResult *HachiHachiRoundResult `json:"lr"`
	ActionLog       []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *HachiHachi) MarshalJSON() ([]byte, error) {
	return json.Marshal(hachihachiJSON{
		Players:         g.players,
		Config:          g.config,
		Phase:           g.state.phase,
		CurrentTurn:     g.state.currentTurn,
		FieldCards:      g.state.fieldCards,
		DrawPile:        g.state.drawPile,
		RoundNumber:     g.state.roundNumber,
		LastCapturer:    g.state.lastCapturer,
		GameEndFlag:     g.state.gameEndFlag,
		Winner:          g.state.winner,
		LastRoundResult: g.state.lastRoundResult,
		ActionLog:       g.state.actionLog,
	})
}

// hachihachiMaxSliceLen caps slice sizes during deserialisation to prevent excessive
// memory allocation from malformed input.
const hachihachiMaxSliceLen = 1000

// hachihachiValidPhase は有効なフェーズかどうか。
func hachihachiValidPhase(p HachiHachiPhase) bool {
	return p >= HachiHachiPhasePlay && p <= HachiHachiPhaseGameEnd
}

// hachihachiValidateCards は復元したカードスライスに nil や範囲外の月/インデックスが
// 無いか検証する。
func hachihachiValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("hachihachi: nil card in state")
		}
		m, i := c.GetDesign(), c.GetValue()
		if m < 1 || m > HachiHachiMonthCnt || i < 1 || i > HachiHachiCardsPerMonth {
			return fmt.Errorf("hachihachi: card out of range (month %d, index %d)", m, i)
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否する。
func (g *HachiHachi) UnmarshalJSON(data []byte) error {
	var j hachihachiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > hachihachiMaxSliceLen || len(j.FieldCards) > hachihachiMaxSliceLen ||
		len(j.DrawPile) > hachihachiMaxSliceLen || len(j.ActionLog) > hachihachiMaxSliceLen {
		return fmt.Errorf("hachihachi: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("hachihachi: invalid config: %w", err)
	}
	if len(j.Players) != HachiHachiPlayerCnt {
		return fmt.Errorf("hachihachi: invalid player count %d, expected %d", len(j.Players), HachiHachiPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("hachihachi: nil player in state")
		}
		if err := hachihachiValidateCards(p.GetCapturedCards()); err != nil {
			return fmt.Errorf("hachihachi: invalid captured cards: %w", err)
		}
		hand := make([]*Card, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			hand[i] = p.GetCard(i)
		}
		if err := hachihachiValidateCards(hand); err != nil {
			return fmt.Errorf("hachihachi: invalid hand cards: %w", err)
		}
	}
	if j.LastRoundResult != nil {
		if len(j.LastRoundResult.Scores) > hachihachiMaxSliceLen {
			return fmt.Errorf("hachihachi: last round result scores exceeds maximum allowed size")
		}
		if j.LastRoundResult.Best < -1 || j.LastRoundResult.Best >= len(j.Players) {
			return fmt.Errorf("hachihachi: last round result best out of range")
		}
		for _, s := range j.LastRoundResult.Scores {
			if s.PlayerIdx < 0 || s.PlayerIdx >= len(j.Players) {
				return fmt.Errorf("hachihachi: last round result player index out of range")
			}
			for _, y := range s.Yaku {
				if y.Key == "" {
					return fmt.Errorf("hachihachi: empty yaku key in last round result")
				}
			}
		}
	}
	if !hachihachiValidPhase(j.Phase) {
		return fmt.Errorf("hachihachi: invalid phase %d", j.Phase)
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= len(j.Players) {
		return fmt.Errorf("hachihachi: current turn out of range")
	}
	if j.LastCapturer < -1 || j.LastCapturer >= len(j.Players) {
		return fmt.Errorf("hachihachi: last capturer out of range")
	}
	if j.Winner < -1 || j.Winner >= len(j.Players) {
		return fmt.Errorf("hachihachi: winner out of range")
	}
	if err := hachihachiValidateCards(j.FieldCards); err != nil {
		return err
	}
	if err := hachihachiValidateCards(j.DrawPile); err != nil {
		return err
	}

	g.players = j.Players
	g.config = j.Config
	g.state = hachihachiState{
		phase:           j.Phase,
		currentTurn:     j.CurrentTurn,
		fieldCards:      j.FieldCards,
		drawPile:        j.DrawPile,
		roundNumber:     j.RoundNumber,
		lastCapturer:    j.LastCapturer,
		gameEndFlag:     j.GameEndFlag,
		winner:          j.Winner,
		lastRoundResult: j.LastRoundResult,
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
