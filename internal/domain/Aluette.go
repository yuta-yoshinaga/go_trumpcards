//go:build !js || !wasm || solo

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
	winIdx, winRank := trick[0].PlayerIdx, -1
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		if r := AluetteRank(tc.Card); r > winRank {
			winRank, winIdx = r, tc.PlayerIdx
		}
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
