//go:build !js || !wasm || extra

// Package domain — ゴーストップ (Go-Stop / 고스톱; JA コッテ) のドメインモデル。
//
// Go-Stop は韓国花札 (Hwatu) 48 枚 (12 か月 × 4 枚) を用いる 2 人用 (人間 vs CPU) の
// フィッシング系ゲーム。Koi-Koi と同じ札・同じ捕獲ロジックを共有するが、得点は韓国式
// (光/띠/열끗/피) で、役が 3 点以上に達したら「ゴー (続行)」か「ストップ (確定)」を
// 選ぶ駆け引きと、ゴーの掛け金 (running bonus / 倍率) およびバク (박, 罰則倍率) が核。
//
// # 札の符号化 (ADR-0033)
//
//   - design = 月 (1..12)
//   - value  = 月内インデックス (1..4)
//
// 花札には専用 PNG アートが無いため、各札は手続き的に描画される (glyph/label/color/
// deck="hanafuda")。カードの正体 (光/열끗/띠/피、띠の色) は (月, インデックス) の
// 対応表 gostopCardTable から引く。表は Koi-Koi と同一の canonical 割り当てを複製する
// (未公開シンボルを跨いで import しない方針のため再定義)。
//
// # ターン進行 (Koi-Koi と同一)
//
// 手番プレイヤーは手札を 1 枚出す。場に同月の札があれば両方を取り札にする (2 枚一致なら
// どちらを取るか選択、3 枚一致=場に既に 3 枚 → すべて取る)。一致が無ければ場に置く。
// 続いて山札の一番上をめくり、同様に処理する。
//
// # 韓国式得点 (本実装のルール)
//
//   - 光(光): 3 光 (雨を除く) = 3 / 雨を含む 3 光 = 2 / 4 光 = 4 / 5 光 = 15
//   - 고도리 (Godori): 2 月鶯・4 月不如帰・8 月雁の 3 鳥 = 5
//   - 띠 (Tti/リボン): 홍단(1/2/3 の赤短) = 3 / 청단(6/9/10 の青短) = 3 /
//     초단(4/5/7 の草短) = 3 / 加えて 6 枚目以降 1 枚ごと +1 (合計 5 枚超過分)
//   - 열끗 (Yeol/動物): 5 枚 = 1、以降 1 枚ごと +1
//   - 피 (Pi/カス): 10 枚 = 1、以降 1 枚ごと +1 (MVP: すべてのカスを 1 피として数える。
//     특수 쌍피 (double-junk) は未モデル化)
//   - 菊の盃 (9 月열끗) は열끗としてのみ数える (双피扱いはしない)
//
// # ゴー / ストップと倍率 (本実装のルール)
//
// カテゴリ合計 (base) が 3 点以上に達した手番で、そのプレイヤーは「ゴー (続行)」か
// 「ストップ (確定)」を選ぶ。ゴー 1 回ごとに +1 点の掛け金 (running bonus) を積み、
// 3 回目以降のゴーで ×2 倍 (4 回目 ×4、…) の倍率を掛ける。すなわち
//
//	goScore(base, goCount) = (base + goCount) × 2^max(0, goCount-2)
//
// # バク (박; ストップ時に適用する ×2 の罰則。乗算的に累積)
//
//   - 光박 (Gwang-bak): 勝者が光の役で上がり (光点 > 0)、敗者の光札が 0 枚 → ×2
//   - 피박 (Pi-bak): 敗者の피が 5 枚未満 → ×2
//   - 고박 (Go-bak): 敗者がこのラウンドでゴーを宣言していた (goCount > 0) のに勝者に
//     先を越されて上がられた → ×2 (ゴーを打った側が罰を受ける)
//
// 最終得点 = goScore × (バク倍率の積)。これを勝者の累計得点に加算する。
//
// # ラウンド終了 / 終局
//
// いずれかがストップを宣言 → 上記の最終得点を加算。双方の手札が尽きても上がりが無ければ
// 引き分け (0 点)。累計得点が TargetScore に到達したプレイヤーが出るか、GoStopMaxRounds
// ラウンドに達したら終局。累計最高点が勝者 (同点は引き分け -1)。
package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// GoStopPlayerCnt はゴーストップのプレイヤー数 (固定 2)。
const GoStopPlayerCnt = 2

// GoStopHandSize は各プレイヤーへ配る手札枚数。
const GoStopHandSize = 10

// GoStopFieldSize はゲーム開始時に場へ置くカード枚数。
const GoStopFieldSize = 8

// GoStopMonthCnt は月数 (= design の最大値)。
const GoStopMonthCnt = 12

// GoStopCardsPerMonth は 1 月あたりの札数 (= value の最大値)。
const GoStopCardsPerMonth = 4

// GoStopMaxRounds はゲームを打ち切る最大ラウンド数 (終局保証)。
const GoStopMaxRounds = 12

// GoStopMinGoScore はゴー/ストップの判断が可能になる最低カテゴリ点。
const GoStopMinGoScore = 3

// GoStopCategory は花札の札種 (韓国式)。
type GoStopCategory int

// 札種定数
const (
	// GoStopPi 피 (カス。最下位の札)
	GoStopPi GoStopCategory = iota
	// GoStopTti 띠 (リボン/短冊)
	GoStopTti
	// GoStopYeol 열끗 (動物/種)
	GoStopYeol
	// GoStopGwang 光 (最上位の札)
	GoStopGwang
)

// GoStopRibbonColor は띠 (リボン) の色。
type GoStopRibbonColor int

// 띠色定数
const (
	// GoStopRibbonNone 띠ではない
	GoStopRibbonNone GoStopRibbonColor = iota
	// GoStopRibbonRedPoetry 홍단 (書き入りの赤短: 1/2/3 月)
	GoStopRibbonRedPoetry
	// GoStopRibbonBlue 청단 (6/9/10 月)
	GoStopRibbonBlue
	// GoStopRibbonPlainRed 초단ほか (書き無しの赤短: 4/5/7/11 月)
	GoStopRibbonPlainRed
)

// GoStopResult は人間視点のゲーム結果 (casino 版 GameResult は extra ビルドから到達不能な
// ため、ゲーム固有型として定義する)。
type GoStopResult int

// ゲーム結果定数
const (
	// GoStopResultNone 未決 (進行中)
	GoStopResultNone GoStopResult = iota
	// GoStopResultWin 人間の勝ち
	GoStopResultWin
	// GoStopResultLose 人間の負け
	GoStopResultLose
	// GoStopResultDraw 引き分け
	GoStopResultDraw
)

// gostopCardInfo は 1 枚の花札の正体。
type gostopCardInfo struct {
	category GoStopCategory
	ribbon   GoStopRibbonColor
	glyph    string // 手続き描画用の絵文字
	name     string // 短い識別名 (英語、デバッグ/ラベル補助)
}

// gostopCardTable[month][index] (month 1..12, index 1..4) が札の正体を返す。
// インデックス 0 は未使用。Koi-Koi と同一の canonical 割り当てを複製する。
var gostopCardTable = [GoStopMonthCnt + 1][GoStopCardsPerMonth + 1]gostopCardInfo{
	1: { // 松 (松に鶴)
		1: {GoStopGwang, GoStopRibbonNone, "🦢", "Crane"},
		2: {GoStopTti, GoStopRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	2: { // 梅 (梅に鶯)
		1: {GoStopYeol, GoStopRibbonNone, "🐦", "Warbler"},
		2: {GoStopTti, GoStopRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	3: { // 桜 (桜に幕)
		1: {GoStopGwang, GoStopRibbonNone, "🌸", "Curtain"},
		2: {GoStopTti, GoStopRibbonRedPoetry, "🎴", "RedPoetryRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	4: { // 藤 (藤に不如帰)
		1: {GoStopYeol, GoStopRibbonNone, "🐦", "Cuckoo"},
		2: {GoStopTti, GoStopRibbonPlainRed, "🎴", "RedRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	5: { // 菖蒲 (菖蒲に八橋)
		1: {GoStopYeol, GoStopRibbonNone, "🌉", "Bridge"},
		2: {GoStopTti, GoStopRibbonPlainRed, "🎴", "RedRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	6: { // 牡丹 (牡丹に蝶)
		1: {GoStopYeol, GoStopRibbonNone, "🦋", "Butterfly"},
		2: {GoStopTti, GoStopRibbonBlue, "🎴", "BlueRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	7: { // 萩 (萩に猪)
		1: {GoStopYeol, GoStopRibbonNone, "🐗", "Boar"},
		2: {GoStopTti, GoStopRibbonPlainRed, "🎴", "RedRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	8: { // 芒 (芒に月 / 芒に雁)
		1: {GoStopGwang, GoStopRibbonNone, "🌕", "Moon"},
		2: {GoStopYeol, GoStopRibbonNone, "🦆", "Geese"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	9: { // 菊 (菊に盃)
		1: {GoStopYeol, GoStopRibbonNone, "🍶", "SakeCup"},
		2: {GoStopTti, GoStopRibbonBlue, "🎴", "BlueRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	10: { // 紅葉 (紅葉に鹿)
		1: {GoStopYeol, GoStopRibbonNone, "🦌", "Deer"},
		2: {GoStopTti, GoStopRibbonBlue, "🎴", "BlueRibbon"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
	11: { // 柳 (柳に小野道風 / 燕 / 短冊 / 雷)
		1: {GoStopGwang, GoStopRibbonNone, "☂️", "Rainman"},
		2: {GoStopYeol, GoStopRibbonNone, "🐦", "Swallow"},
		3: {GoStopTti, GoStopRibbonPlainRed, "🎴", "RedRibbon"},
		4: {GoStopPi, GoStopRibbonNone, "⚡", "Lightning"},
	},
	12: { // 桐 (桐に鳳凰)
		1: {GoStopGwang, GoStopRibbonNone, "🦅", "Phoenix"},
		2: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		3: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
		4: {GoStopPi, GoStopRibbonNone, "🍂", "Chaff"},
	},
}

// gostopInfo は札の正体を返す。範囲外の札はカス相当を返す (防御的)。
func gostopInfo(c *Card) gostopCardInfo {
	if c == nil {
		return gostopCardInfo{category: GoStopPi, glyph: "🍂", name: "Chaff"}
	}
	m, i := c.GetDesign(), c.GetValue()
	if m < 1 || m > GoStopMonthCnt || i < 1 || i > GoStopCardsPerMonth {
		return gostopCardInfo{category: GoStopPi, glyph: "🍂", name: "Chaff"}
	}
	return gostopCardTable[m][i]
}

// gostopSameMonth は 2 枚が同月かどうか。
func gostopSameMonth(a, b *Card) bool {
	return a != nil && b != nil && a.GetDesign() == b.GetDesign()
}

// --- 得点 ---

const (
	gostopPointsGoko5    = 15 // 5 光
	gostopPointsGoko4    = 4  // 4 光
	gostopPointsGoko3    = 3  // 3 光 (雨無し)
	gostopPointsGoko3Ame = 2  // 3 光 (雨含む)
	gostopPointsGodori   = 5  // 고도리
	gostopPointsHongdan  = 3  // 홍단
	gostopPointsChongdan = 3  // 청단
	gostopPointsChodan   = 3  // 초단
	gostopTtiThreshold   = 5  // 띠は 5 枚超過分で加点
	gostopYeolThreshold  = 5  // 열끗 5 枚で 1 点
	gostopPiThreshold    = 10 // 피 10 枚で 1 点
)

// GoStopBreakdown はカテゴリ別の得点内訳と、ゴー適用後の点数を保持する。
type GoStopBreakdown struct {
	Gwang       int `json:"gwang"`       // 光の点
	Godori      int `json:"godori"`      // 고도리の点
	Tti         int `json:"tti"`         // 띠の点
	Yeol        int `json:"yeol"`        // 열끗の点
	Pi          int `json:"pi"`          // 피の点
	Base        int `json:"base"`        // 上記 5 カテゴリの合計 (ゴー適用前)
	GoCount     int `json:"goCount"`     // このプレイヤーのゴー宣言回数
	GoMult      int `json:"goMult"`      // ゴー倍率 (3 回目以降 ×2^(n-2))
	GoScore     int `json:"goScore"`     // ゴー適用後の点 = (Base+GoCount)×GoMult
	BrightCount int `json:"brightCount"` // 光札枚数 (バク判定用)
	RibbonCount int `json:"ribbonCount"` // 띠枚数
	AnimalCount int `json:"animalCount"` // 열끗枚数
	PiCount     int `json:"piCount"`     // 피枚数 (バク判定用)
}

// gostopScoreAfterGo はカテゴリ合計 base とゴー回数 goCount から掛け金・倍率を適用する。
//
//	(base + goCount) × 2^max(0, goCount-2)
func gostopScoreAfterGo(base, goCount int) (score, mult int) {
	mult = 1
	if goCount >= 3 {
		mult = 1 << (goCount - 2)
	}
	return (base + goCount) * mult, mult
}

// gostopEvaluateScore は取り札から韓国式のカテゴリ得点内訳を計算し、ゴー回数を適用した
// 最終点 (goScore) を返す純粋関数。バク倍率はここでは適用しない (相手の取り札が必要な
// ため endRound で計算する)。
func gostopEvaluateScore(captured []*Card, goCount int) (*GoStopBreakdown, int) {
	brights := 0
	hasRain := false
	animals, piCount, ribbons := 0, 0, 0
	hasWarbler, hasCuckoo, hasGeese := false, false, false
	// 띠の月別存在フラグ。
	ribbonMonth := [GoStopMonthCnt + 1]bool{}

	for _, c := range captured {
		if c == nil {
			continue
		}
		info := gostopInfo(c)
		month := c.GetDesign()
		switch info.category {
		case GoStopGwang:
			brights++
			if month == 11 {
				hasRain = true
			}
		case GoStopYeol:
			animals++
			switch month {
			case 2:
				hasWarbler = true
			case 4:
				hasCuckoo = true
			case 8:
				hasGeese = true
			}
		case GoStopTti:
			ribbons++
			if month >= 1 && month <= GoStopMonthCnt {
				ribbonMonth[month] = true
			}
		case GoStopPi:
			piCount++
		}
	}

	b := &GoStopBreakdown{
		GoCount:     goCount,
		BrightCount: brights,
		RibbonCount: ribbons,
		AnimalCount: animals,
		PiCount:     piCount,
	}

	// 光。
	switch {
	case brights >= 5:
		b.Gwang = gostopPointsGoko5
	case brights == 4:
		b.Gwang = gostopPointsGoko4
	case brights == 3 && hasRain:
		b.Gwang = gostopPointsGoko3Ame
	case brights == 3:
		b.Gwang = gostopPointsGoko3
	}

	// 고도리。
	if hasWarbler && hasCuckoo && hasGeese {
		b.Godori = gostopPointsGodori
	}

	// 띠 (홍단/청단/초단 + 5 枚超過分)。
	if ribbonMonth[1] && ribbonMonth[2] && ribbonMonth[3] {
		b.Tti += gostopPointsHongdan
	}
	if ribbonMonth[6] && ribbonMonth[9] && ribbonMonth[10] {
		b.Tti += gostopPointsChongdan
	}
	if ribbonMonth[4] && ribbonMonth[5] && ribbonMonth[7] {
		b.Tti += gostopPointsChodan
	}
	if ribbons >= gostopTtiThreshold {
		b.Tti += 1 + (ribbons - gostopTtiThreshold)
	}

	// 열끗。
	if animals >= gostopYeolThreshold {
		b.Yeol = 1 + (animals - gostopYeolThreshold)
	}

	// 피。
	if piCount >= gostopPiThreshold {
		b.Pi = 1 + (piCount - gostopPiThreshold)
	}

	b.Base = b.Gwang + b.Godori + b.Tti + b.Yeol + b.Pi
	b.GoScore, b.GoMult = gostopScoreAfterGo(b.Base, goCount)
	return b, b.GoScore
}

// GoStopPhase はゲームフェーズ。
type GoStopPhase int

// Go-Stop のフェーズ定数
const (
	// GoStopPhasePlay プレイ中 (手札を 1 枚出す)
	GoStopPhasePlay GoStopPhase = 0
	// GoStopPhaseGoDecision ゴー/ストップ決断中
	GoStopPhaseGoDecision GoStopPhase = 1
	// GoStopPhaseRoundEnd ラウンド終了 (結果表示。次ラウンド待ち)
	GoStopPhaseRoundEnd GoStopPhase = 2
	// GoStopPhaseGameEnd 終局
	GoStopPhaseGameEnd GoStopPhase = 3
)

// GoStopRoundResult は 1 ラウンドの結果。
type GoStopRoundResult struct {
	Winner     int              `json:"winner"`     // 勝者インデックス (-1 = 引き分け)
	Breakdown  *GoStopBreakdown `json:"breakdown"`  // 勝者の得点内訳
	BasePoints int              `json:"basePoints"` // カテゴリ合計 (ゴー適用前)
	GoScore    int              `json:"goScore"`    // ゴー適用後の点
	BakMult    int              `json:"bakMult"`    // バク倍率の積
	Total      int              `json:"total"`      // 実際に加算された得点 = GoScore×BakMult
	GwangBak   bool             `json:"gwangBak"`   // 光박 成立
	PiBak      bool             `json:"piBak"`      // 피박 成立
	GoBak      bool             `json:"goBak"`      // 고박 成立
	GoCount    int              `json:"goCount"`    // 勝者のゴー宣言回数
}

// GoStopHint はヒント情報。
type GoStopHint struct {
	CardIndex  int    `json:"cardIndex"`  // 推奨手札インデックス (-1 = なし)
	FieldIndex int    `json:"fieldIndex"` // 推奨捕獲場札インデックス (-1 = なし/自動)
	Go         int    `json:"go"`         // 決断ヒント (1=ゴー, 0=ストップ, -1=非該当)
	Reason     string `json:"reason"`     // 理由キー
}

// gostopState はゲーム進行状態。
type gostopState struct {
	phase            GoStopPhase
	currentTurn      int
	fieldCards       []*Card
	drawPile         []*Card
	roundNumber      int
	roundWinner      int // -1 = 未決/引き分け
	gameEndFlag      bool
	winner           int // 終局時の勝者 (-1 = 引き分け)
	lastRoundResult  *GoStopRoundResult
	pendingBreakdown *GoStopBreakdown
	pendingPoints    int // 決断フェーズのカテゴリ合計
	actionLogBase
}

// GoStop はゴーストップゲームの状態を保持する集約ルート。
type GoStop struct {
	players []*GoStopPlayer
	config  GoStopConfig
	state   gostopState
}

// NewGoStop はコンストラクタ。
func NewGoStop(players []*GoStopPlayer, config GoStopConfig) *GoStop {
	return &GoStop{
		players: players,
		config:  config,
		state: gostopState{
			phase:       GoStopPhasePlay,
			roundWinner: -1,
			winner:      -1,
		},
	}
}

// NewDefaultGoStop は標準の 2 人構成 (1 human + 1 CPU) で GoStop を生成する。
func NewDefaultGoStop() *GoStop {
	players := make([]*GoStopPlayer, GoStopPlayerCnt)
	players[0] = NewGoStopPlayer(true)
	players[1] = NewGoStopPlayer(false)
	return NewGoStop(players, DefaultGoStopConfig())
}

// buildGoStopDeck は花札 48 枚を design=月(1..12)/value=index(1..4) で直接生成する。
func buildGoStopDeck() []*Card {
	deck := make([]*Card, 0, GoStopMonthCnt*GoStopCardsPerMonth)
	for m := 1; m <= GoStopMonthCnt; m++ {
		for i := 1; i <= GoStopCardsPerMonth; i++ {
			deck = append(deck, NewCard(m, i, false))
		}
	}
	return deck
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。
func (g *GoStop) Reset() {
	for _, p := range g.players {
		p.Reset()
		p.ResetRound()
		p.ResetScore()
	}
	g.state = gostopState{
		phase:         GoStopPhasePlay,
		roundWinner:   -1,
		winner:        -1,
		roundNumber:   1,
		actionLogBase: actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound はラウンド終了後に次ラウンドを開始する。
func (g *GoStop) NextRound() {
	if g.state.gameEndFlag || g.state.phase != GoStopPhaseRoundEnd {
		return
	}
	g.state.roundNumber++
	g.startRound()
}

// startRound はデッキ生成・配札・場札配置を行い、プレイフェーズを開始する。
// ラウンドごとに先手を交代する ((roundNumber-1) % 2)。
func (g *GoStop) startRound() {
	deck := buildGoStopDeck()
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	for _, p := range g.players {
		p.Reset()
		p.ResetRound()
	}
	g.state.fieldCards = make([]*Card, 0, GoStopFieldSize)
	g.state.roundWinner = -1
	g.state.pendingBreakdown = nil
	g.state.pendingPoints = 0
	g.state.phase = GoStopPhasePlay
	g.state.currentTurn = (g.state.roundNumber - 1) % GoStopPlayerCnt

	// 交互配り: 各プレイヤーへ 10 枚、場へ 8 枚。
	pos := 0
	for k := 0; k < GoStopHandSize; k++ {
		for _, p := range g.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	for k := 0; k < GoStopFieldSize; k++ {
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
func (g *GoStop) allHandsEmpty() bool {
	for _, p := range g.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// --- 捕獲ロジック (Koi-Koi と同一) ---

// gostopFieldMatches は場札のうち card と同月のインデックスを返す。
func (g *GoStop) gostopFieldMatches(card *Card) []int {
	var out []int
	for i, c := range g.state.fieldCards {
		if gostopSameMonth(c, card) {
			out = append(out, i)
		}
	}
	return out
}

// gostopBestFieldMatch は複数一致のうち最も価値の高い場札インデックスを返す。
func (g *GoStop) gostopBestFieldMatch(matches []int) int {
	best := -1
	bestVal := -1
	for _, idx := range matches {
		if idx < 0 || idx >= len(g.state.fieldCards) {
			continue
		}
		v := gostopCardWeight(g.state.fieldCards[idx])
		if v > bestVal {
			bestVal = v
			best = idx
		}
	}
	return best
}

// gostopCardWeight は札の概算価値 (光>열끗>띠>피)。捕獲選択と CPU AI で使用。
func gostopCardWeight(c *Card) int {
	switch gostopInfo(c).category {
	case GoStopGwang:
		return 5
	case GoStopYeol:
		return 3
	case GoStopTti:
		return 2
	default:
		return 1
	}
}

// gostopPlaceCard は 1 枚 (手札またはめくり札) を場と突き合わせて解決する。
//   - 一致なし: 場に置く (捨て札)。
//   - 一致 1 枚: その札と共に捕獲。
//   - 一致 2 枚: chosen が一致札なら chosen を、そうでなければ最良の 1 枚を捕獲。
//   - 一致 3 枚: すべて捕獲 (場の同月 4 枚目)。
func (g *GoStop) gostopPlaceCard(playerIdx int, card *Card, chosen int) {
	matches := g.gostopFieldMatches(card)
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
			sel = g.gostopBestFieldMatch(matches)
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
func (g *GoStop) removeFieldByIndex(idxs []int) {
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
func (g *GoStop) PlayerPlay(handIdx, fieldIdx int) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != GoStopPhasePlay {
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
		if fieldIdx >= len(g.state.fieldCards) || !gostopSameMonth(g.state.fieldCards[fieldIdx], card) {
			return NewDomainError(ErrInvalidPlay, "chosen field card does not match the played card's month")
		}
	}
	g.applyTurn(g.state.currentTurn, handIdx, fieldIdx)
	return nil
}

// CpuPlay は CPU のプレイ手番を 1 回進める。
func (g *GoStop) CpuPlay() {
	if g.state.gameEndFlag || g.state.phase != GoStopPhasePlay {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() || p.GetCardsSize() == 0 {
		return
	}
	handIdx, fieldIdx := g.chooseCpuPlay(g.state.currentTurn)
	g.applyTurn(g.state.currentTurn, handIdx, fieldIdx)
}

// applyTurn は手札を出し→めくり札を処理→得点判定を行う共通処理。カテゴリ合計が 3 点
// 以上かつ前回決断時より増えていればゴー/ストップ決断フェーズへ移り、そうでなければ
// 手番を進める。
func (g *GoStop) applyTurn(playerIdx, handIdx, fieldIdx int) {
	player := g.players[playerIdx]
	card := player.RemoveCard(handIdx)
	if card == nil {
		return
	}
	beforeField := len(g.state.fieldCards)
	g.gostopPlaceCard(playerIdx, card, fieldIdx)
	handCaptured := len(g.state.fieldCards) <= beforeField
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s (%s)",
		g.playerName(playerIdx), gostopCardStr(card), gostopCapturedWord(handCaptured)), []*Card{card})

	// めくり札。
	if len(g.state.drawPile) > 0 {
		drawn := g.state.drawPile[0]
		g.state.drawPile = g.state.drawPile[1:]
		before2 := len(g.state.fieldCards)
		g.gostopPlaceCard(playerIdx, drawn, -1)
		drawCaptured := len(g.state.fieldCards) <= before2
		g.appendLog(playerIdx, "draw", fmt.Sprintf("%s draws %s (%s)",
			g.playerName(playerIdx), gostopCardStr(drawn), gostopCapturedWord(drawCaptured)), []*Card{drawn})
	}

	// 得点判定。
	bd, _ := gostopEvaluateScore(player.GetCapturedCards(), player.GetGoCount())
	if bd.Base >= GoStopMinGoScore && bd.Base > player.GetLastScorePoints() {
		g.state.pendingBreakdown = bd
		g.state.pendingPoints = bd.Base
		g.state.phase = GoStopPhaseGoDecision
		g.appendLog(playerIdx, "score",
			fmt.Sprintf("%s reaches %d points", g.playerName(playerIdx), bd.Base), nil)
		return
	}
	g.advanceTurn()
}

// PlayerDecide は人間のゴー/ストップ決断 (goDecision=true でゴー、false でストップ)。
func (g *GoStop) PlayerDecide(goDecision bool) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != GoStopPhaseGoDecision {
		return NewDomainError(ErrWrongPhase, "not in go/stop decision phase")
	}
	if !g.players[g.state.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyDecision(g.state.currentTurn, goDecision)
	return nil
}

// CpuDecide は CPU のゴー/ストップ決断を 1 回進める。
func (g *GoStop) CpuDecide() {
	if g.state.gameEndFlag || g.state.phase != GoStopPhaseGoDecision {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() {
		return
	}
	g.applyDecision(g.state.currentTurn, g.chooseCpuDecision(g.state.currentTurn))
}

// applyDecision はゴー/ストップ決断を適用する。ゴーなら掛け金カウンタを増やして手番を
// 進め、ストップなら現在の得点でラウンドを終える。
func (g *GoStop) applyDecision(playerIdx int, goDecision bool) {
	player := g.players[playerIdx]
	if goDecision && player.GetCardsSize() > 0 {
		player.IncGoCount()
		player.SetCalledGo(true)
		player.SetLastScorePoints(g.state.pendingPoints)
		g.appendLog(playerIdx, "go", fmt.Sprintf("%s calls Go (x%d)", g.playerName(playerIdx), player.GetGoCount()), nil)
		g.state.pendingBreakdown = nil
		g.state.pendingPoints = 0
		g.state.phase = GoStopPhasePlay
		g.advanceTurn()
		return
	}
	// ストップ (あがり)。手札が無い場合も強制的にここで確定する。
	g.appendLog(playerIdx, "stop", fmt.Sprintf("%s stops", g.playerName(playerIdx)), nil)
	g.endRound(playerIdx)
}

// advanceTurn は手番を次へ進め、双方の手札が尽きたら引き分けでラウンドを終える。
func (g *GoStop) advanceTurn() {
	if g.allHandsEmpty() {
		g.endRound(-1)
		return
	}
	g.state.currentTurn = (g.state.currentTurn + 1) % GoStopPlayerCnt
}

// gostopBakMultiplier は勝者と敗者の状態からバク倍率とフラグを計算する。
func gostopBakMultiplier(winner *GoStopBreakdown, loser *GoStopBreakdown, loserGoCount int) (mult int, gwangBak, piBak, goBak bool) {
	mult = 1
	// 光박: 勝者が光で上がり (光点 > 0) かつ敗者の光札 0 枚。
	if winner.Gwang > 0 && loser.BrightCount == 0 {
		mult *= 2
		gwangBak = true
	}
	// 피박: 勝者が피で得点 (피点 > 0) しており、かつ敗者の피が 5 枚未満。
	if winner.Pi > 0 && loser.PiCount < 5 {
		mult *= 2
		piBak = true
	}
	// 고박: 敗者がゴーを宣言していた。
	if loserGoCount > 0 {
		mult *= 2
		goBak = true
	}
	return mult, gwangBak, piBak, goBak
}

// endRound はラウンドを終える。winnerIdx>=0 なら得点を加点、-1 は引き分け。
func (g *GoStop) endRound(winnerIdx int) {
	g.state.roundWinner = winnerIdx
	g.state.pendingBreakdown = nil
	g.state.pendingPoints = 0

	result := &GoStopRoundResult{Winner: winnerIdx, BakMult: 1}
	if winnerIdx >= 0 {
		loserIdx := (winnerIdx + 1) % GoStopPlayerCnt
		winner := g.players[winnerIdx]
		loser := g.players[loserIdx]
		wbd, goScore := gostopEvaluateScore(winner.GetCapturedCards(), winner.GetGoCount())
		lbd, _ := gostopEvaluateScore(loser.GetCapturedCards(), loser.GetGoCount())
		bakMult, gwangBak, piBak, goBak := gostopBakMultiplier(wbd, lbd, loser.GetGoCount())
		total := goScore * bakMult
		result.Breakdown = wbd
		result.BasePoints = wbd.Base
		result.GoScore = goScore
		result.BakMult = bakMult
		result.Total = total
		result.GwangBak = gwangBak
		result.PiBak = piBak
		result.GoBak = goBak
		result.GoCount = winner.GetGoCount()
		winner.AddScore(total)
		g.appendLog(winnerIdx, "roundWin",
			fmt.Sprintf("%s wins round with %d points (bak x%d)", g.playerName(winnerIdx), total, bakMult), nil)
	} else {
		g.appendLog(-1, "draw", "round drawn (no winner)", nil)
	}
	g.state.lastRoundResult = result

	// 終局判定。
	if g.reachedTarget() || g.state.roundNumber >= GoStopMaxRounds {
		g.finishGame()
		return
	}
	g.state.phase = GoStopPhaseRoundEnd
}

// reachedTarget は目標得点に到達したプレイヤーがいるか。
func (g *GoStop) reachedTarget() bool {
	for _, p := range g.players {
		if p.GetScore() >= g.config.TargetScore {
			return true
		}
	}
	return false
}

// finishGame は終局処理: 累計最高点を勝者にする (同点は引き分け -1)。
func (g *GoStop) finishGame() {
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
	g.state.phase = GoStopPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended (winner %d)", best), nil)
}

// --- CPU AI ---

// chooseCpuPlay は CPU の手札インデックスと捕獲場札を選ぶ。
func (g *GoStop) chooseCpuPlay(playerIdx int) (int, int) {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0, -1
	}
	if g.config.CpuDifficulty == GoStopCpuDifficultyEasy {
		idx := rand.Intn(size)
		return idx, g.cpuFieldChoice(player.GetCard(idx))
	}
	bestIdx := 0
	bestScore := math.MinInt
	for i := 0; i < size; i++ {
		card := player.GetCard(i)
		matches := g.gostopFieldMatches(card)
		s := 0
		if len(matches) > 0 {
			s = gostopCardWeight(card) + gostopCardWeight(g.state.fieldCards[g.gostopBestFieldMatch(matches)])
			if len(matches) >= 3 {
				s += 4
			}
		} else {
			s = -gostopCardWeight(card)
		}
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx, g.cpuFieldChoice(player.GetCard(bestIdx))
}

// cpuFieldChoice は 2 枚一致時に取る場札を選ぶ (最良札)。一致 0/1/3 は -1 (自動)。
func (g *GoStop) cpuFieldChoice(card *Card) int {
	matches := g.gostopFieldMatches(card)
	if len(matches) == 2 {
		return g.gostopBestFieldMatch(matches)
	}
	return -1
}

// chooseCpuDecision は CPU がゴー (true) かストップ (false) かを決める。
func (g *GoStop) chooseCpuDecision(playerIdx int) bool {
	player := g.players[playerIdx]
	total := g.state.pendingPoints
	handsLeft := player.GetCardsSize()
	switch g.config.CpuDifficulty {
	case GoStopCpuDifficultyEasy:
		return false // 常にストップ (保守的)
	case GoStopCpuDifficultyHard:
		return total < 7 && handsLeft >= 2
	default: // Normal
		return total < 4 && handsLeft >= 4
	}
}

// --- Hint ---

// GetHint は人間の手番における推奨手を返す。
func (g *GoStop) GetHint() *GoStopHint {
	if g.state.gameEndFlag {
		return nil
	}
	human := g.findHumanIdx()
	if human < 0 || g.state.currentTurn != human {
		return nil
	}
	switch g.state.phase {
	case GoStopPhaseGoDecision:
		gg := 0
		reason := "stop_secure"
		if g.state.pendingPoints < 4 && g.players[human].GetCardsSize() >= 4 {
			gg = 1
			reason = "go_lowscore"
		}
		return &GoStopHint{CardIndex: -1, FieldIndex: -1, Go: gg, Reason: reason}
	case GoStopPhasePlay:
		idx, field := g.chooseCpuPlay(human)
		reason := "capture"
		if len(g.gostopFieldMatches(g.players[human].GetCard(idx))) == 0 {
			reason = "discard_low"
		}
		return &GoStopHint{CardIndex: idx, FieldIndex: field, Go: -1, Reason: reason}
	default:
		return nil
	}
}

// --- ヘルパー ---

func (g *GoStop) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// sortHumanHand は人間の手札を月→インデックス順に並べ替える。
func (g *GoStop) sortHumanHand() {
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

func (g *GoStop) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return "CPU"
}

func (g *GoStop) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// gostopCardStr は札を "松·光" のように表す (ログ/デバッグ用)。
func gostopCardStr(c *Card) string {
	if c == nil {
		return "??"
	}
	return gostopMonthKanji(c.GetDesign()) + "·" + gostopCategoryShort(gostopInfo(c).category)
}

func gostopCapturedWord(captured bool) string {
	if captured {
		return "captures"
	}
	return "to field"
}

// gostopMonthKanji は月番号を月札の代表漢字にする。
func gostopMonthKanji(month int) string {
	kanji := []string{"?", "松", "梅", "桜", "藤", "菖", "牡", "萩", "芒", "菊", "紅", "柳", "桐"}
	if month >= 1 && month <= GoStopMonthCnt {
		return kanji[month]
	}
	return "?"
}

// gostopCategoryShort は札種の短い日本語表記。
func gostopCategoryShort(cat GoStopCategory) string {
	switch cat {
	case GoStopGwang:
		return "光"
	case GoStopYeol:
		return "열"
	case GoStopTti:
		return "띠"
	default:
		return "피"
	}
}

// --- 描画用エクスポートアクセサ (adapter/presenter が参照) ---

// GoStopCardGlyph は札の手続き描画用グリフ (絵文字) を返す。
func GoStopCardGlyph(c *Card) string { return gostopInfo(c).glyph }

// GoStopCardCategory は札種 (光/열끗/띠/피) を返す。
func GoStopCardCategory(c *Card) GoStopCategory { return gostopInfo(c).category }

// GoStopCardRibbonColor は띠札の色を返す (띠でなければ None)。
func GoStopCardRibbonColor(c *Card) GoStopRibbonColor { return gostopInfo(c).ribbon }

// GoStopCardLabel は札の短い日本語ラベル ("松·光" 等) を返す。
func GoStopCardLabel(c *Card) string { return gostopCardStr(c) }

// --- 状態アクセサ ---

// IsHumanTurn は現在プレイ/決断の手番が人間かどうかを返す。
func (g *GoStop) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	if g.state.phase != GoStopPhasePlay && g.state.phase != GoStopPhaseGoDecision {
		return false
	}
	return g.players[g.state.currentTurn].GetIsHuman()
}

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *GoStop) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *GoStop) GetPhase() GoStopPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *GoStop) SetPhase(p GoStopPhase) { g.state.phase = p }

// GetCurrentTurn は現在の手番を返す。
func (g *GoStop) GetCurrentTurn() int { return g.state.currentTurn }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *GoStop) SetCurrentTurn(idx int) { g.state.currentTurn = idx }

// GetFieldCards は場札を返す。
func (g *GoStop) GetFieldCards() []*Card { return g.state.fieldCards }

// SetFieldCards は場札を設定する (テスト用)。
func (g *GoStop) SetFieldCards(cards []*Card) { g.state.fieldCards = cards }

// GetRemainingDeck は山札の残り枚数を返す。
func (g *GoStop) GetRemainingDeck() int { return len(g.state.drawPile) }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *GoStop) GetRoundNumber() int { return g.state.roundNumber }

// GetRoundWinner は直近ラウンドの勝者を返す (-1 = 引き分け/未決)。
func (g *GoStop) GetRoundWinner() int { return g.state.roundWinner }

// GetLastRoundResult は直近ラウンド結果を返す (nil の場合もある)。
func (g *GoStop) GetLastRoundResult() *GoStopRoundResult { return g.state.lastRoundResult }

// GetPendingBreakdown は決断フェーズで表示する得点内訳を返す。
func (g *GoStop) GetPendingBreakdown() *GoStopBreakdown { return g.state.pendingBreakdown }

// GetPendingPoints は決断フェーズのカテゴリ合計点を返す。
func (g *GoStop) GetPendingPoints() int { return g.state.pendingPoints }

// GetWinner は終局時の勝者を返す (-1 = 引き分け/未決)。
func (g *GoStop) GetWinner() int { return g.state.winner }

// GetResult は人間視点のゲーム結果を返す (終局していなければ None)。
func (g *GoStop) GetResult() GoStopResult {
	if !g.state.gameEndFlag {
		return GoStopResultNone
	}
	human := g.findHumanIdx()
	switch {
	case g.state.winner < 0:
		return GoStopResultDraw
	case g.state.winner == human:
		return GoStopResultWin
	default:
		return GoStopResultLose
	}
}

// GetPlayerCnt はプレイヤー数を返す。
func (g *GoStop) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *GoStop) GetPlayer(i int) *GoStopPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetConfig はローカルルール設定を返す。
func (g *GoStop) GetConfig() GoStopConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *GoStop) SetConfig(cfg GoStopConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *GoStop) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// GetScore は指定プレイヤーの現在の取り札の得点内訳と最終点を返す (UI 補助)。
func (g *GoStop) GetScore(playerIdx int) (*GoStopBreakdown, int) {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil, 0
	}
	return gostopEvaluateScore(g.players[playerIdx].GetCapturedCards(), g.players[playerIdx].GetGoCount())
}

// GetPlayableIndices はプレイフェーズで人間がプレイできる手札インデックス (全札) を返す。
func (g *GoStop) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != GoStopPhasePlay {
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
func (g *GoStop) GetCaptureOptions(playerIdx int) map[int][]int {
	out := make(map[int][]int)
	if playerIdx < 0 || playerIdx >= len(g.players) || g.state.phase != GoStopPhasePlay {
		return out
	}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if m := g.gostopFieldMatches(p.GetCard(i)); len(m) > 0 {
			out[i] = m
		}
	}
	return out
}

// --- JSON Serialization ---

// gostopJSON is the JSON wire format for GoStop.
type gostopJSON struct {
	Players          []*GoStopPlayer    `json:"pl"`
	Config           GoStopConfig       `json:"cf"`
	Phase            GoStopPhase        `json:"ph"`
	CurrentTurn      int                `json:"ct"`
	FieldCards       []*Card            `json:"fd"`
	DrawPile         []*Card            `json:"dp"`
	RoundNumber      int                `json:"rn"`
	RoundWinner      int                `json:"rw"`
	GameEndFlag      bool               `json:"ge"`
	Winner           int                `json:"wn"`
	LastRoundResult  *GoStopRoundResult `json:"lr"`
	PendingBreakdown *GoStopBreakdown   `json:"pb"`
	PendingPoints    int                `json:"pp"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *GoStop) MarshalJSON() ([]byte, error) {
	return json.Marshal(gostopJSON{
		Players:          g.players,
		Config:           g.config,
		Phase:            g.state.phase,
		CurrentTurn:      g.state.currentTurn,
		FieldCards:       g.state.fieldCards,
		DrawPile:         g.state.drawPile,
		RoundNumber:      g.state.roundNumber,
		RoundWinner:      g.state.roundWinner,
		GameEndFlag:      g.state.gameEndFlag,
		Winner:           g.state.winner,
		LastRoundResult:  g.state.lastRoundResult,
		PendingBreakdown: g.state.pendingBreakdown,
		PendingPoints:    g.state.pendingPoints,
		ActionLog:        g.state.actionLog,
	})
}

// gostopMaxSliceLen caps slice sizes during deserialisation to prevent excessive
// memory allocation from malformed input.
const gostopMaxSliceLen = 1000

// gostopValidPhase は有効なフェーズかどうか。
func gostopValidPhase(p GoStopPhase) bool {
	return p >= GoStopPhasePlay && p <= GoStopPhaseGameEnd
}

// gostopValidateCards は復元したカードスライスに nil や範囲外の月/インデックスが
// 無いか検証する。
func gostopValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("gostop: nil card in state")
		}
		m, i := c.GetDesign(), c.GetValue()
		if m < 1 || m > GoStopMonthCnt || i < 1 || i > GoStopCardsPerMonth {
			return fmt.Errorf("gostop: card out of range (month %d, index %d)", m, i)
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否する。
func (g *GoStop) UnmarshalJSON(data []byte) error {
	var j gostopJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > gostopMaxSliceLen || len(j.FieldCards) > gostopMaxSliceLen ||
		len(j.DrawPile) > gostopMaxSliceLen || len(j.ActionLog) > gostopMaxSliceLen {
		return fmt.Errorf("gostop: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("gostop: invalid config: %w", err)
	}
	if len(j.Players) != GoStopPlayerCnt {
		return fmt.Errorf("gostop: invalid player count %d, expected %d", len(j.Players), GoStopPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("gostop: nil player in state")
		}
	}
	if !gostopValidPhase(j.Phase) {
		return fmt.Errorf("gostop: invalid phase %d", j.Phase)
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= len(j.Players) {
		return fmt.Errorf("gostop: current turn out of range")
	}
	if j.RoundWinner < -1 || j.RoundWinner >= len(j.Players) {
		return fmt.Errorf("gostop: round winner out of range")
	}
	if j.Winner < -1 || j.Winner >= len(j.Players) {
		return fmt.Errorf("gostop: winner out of range")
	}
	if err := gostopValidateCards(j.FieldCards); err != nil {
		return err
	}
	if err := gostopValidateCards(j.DrawPile); err != nil {
		return err
	}

	g.players = j.Players
	g.config = j.Config
	g.state = gostopState{
		phase:            j.Phase,
		currentTurn:      j.CurrentTurn,
		fieldCards:       j.FieldCards,
		drawPile:         j.DrawPile,
		roundNumber:      j.RoundNumber,
		roundWinner:      j.RoundWinner,
		gameEndFlag:      j.GameEndFlag,
		winner:           j.Winner,
		lastRoundResult:  j.LastRoundResult,
		pendingBreakdown: j.PendingBreakdown,
		pendingPoints:    j.PendingPoints,
		actionLogBase:    actionLogBase{actionLog: j.ActionLog},
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
