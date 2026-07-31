//go:build !js || !wasm || extra3

package domain

// BostonBidKind は宣言の型。
//
// **3 種類ある。**トリック数（宣言以上を取る）、ミゼール（1 つも取らない）、
// ピッコリッシモ（**ちょうど 1 つだけ**取る）。issue #4394 は 3 つ目に触れて
// いないうえ、ミゼールをトリック宣言の後段に置いているが、実際は序列が交互に
// 挟まる。
type BostonBidKind int

// Boston の宣言の型
const (
	// BostonKindPass パス
	BostonKindPass BostonBidKind = iota
	// BostonKindTricks トリック数の宣言 (宣言数以上を取る)
	BostonKindTricks
	// BostonKindMisere ミゼール (1 トリックも取らない)
	BostonKindMisere
	// BostonKindPiccolissimo ピッコリッシモ (ちょうど 1 トリック取る)
	BostonKindPiccolissimo
)

// BostonBidLevel は宣言の序列。値が大きいほど強い。
type BostonBidLevel int

// Boston の宣言序列。
//
// **トリック数の間にミゼールが割り込む。**Little Misère は 7 トリックより下、
// Grand Misère は 9 トリックより下である。issue のようにトリック宣言を並べた
// あとにミゼールを置くと、競りの意味そのものが変わる。
const (
	// BostonBidPass パス
	BostonBidPass BostonBidLevel = iota
	// BostonBidFive Simple Boston (5 トリック)
	BostonBidFive
	// BostonBidSix 6 トリック
	BostonBidSix
	// BostonBidLittleMisere リトル・ミゼール
	BostonBidLittleMisere
	// BostonBidSeven 7 トリック
	BostonBidSeven
	// BostonBidPiccolissimo ピッコリッシモ (ちょうど 1 トリック)
	BostonBidPiccolissimo
	// BostonBidEight 8 トリック
	BostonBidEight
	// BostonBidGrandMisere グランド・ミゼール
	BostonBidGrandMisere
	// BostonBidNine 9 トリック
	BostonBidNine
	// BostonBidLittleMisereTable リトル・ミゼール（手札公開）
	BostonBidLittleMisereTable
	// BostonBidTen 10 トリック
	BostonBidTen
	// BostonBidGrandMisereTable グランド・ミゼール（手札公開）
	BostonBidGrandMisereTable
	// BostonBidEleven 11 トリック
	BostonBidEleven
	// BostonBidTwelve 12 トリック
	BostonBidTwelve
	// BostonBidChelem Chelem / Grand Boston (13 トリック)
	BostonBidChelem
	// BostonBidChelemTable Chelem（手札公開）
	BostonBidChelemTable
	// BostonBidLevelCount 宣言の総数
	BostonBidLevelCount
)

// bostonBidSpec は 1 つの宣言の性質。
type bostonBidSpec struct {
	kind BostonBidKind
	// tricks は目標トリック数 (ミゼールは 0、ピッコリッシモは 1)。
	tricks int
	// exposed は第 1 トリックのあとに手札を公開するか。
	exposed bool
	// canCallPartner はパートナーを指名できるか。
	//
	// **トリック数の宣言だけ。**ミゼール系・ピッコリッシモ・スラムは単独固定。
	canCallPartner bool
	// payout は達成時に各相手から受け取る額（失敗時は同額を各相手に払う）。
	payout int
	// name は表示用の識別子。
	name string
}

// bostonBidTable は序列そのままの宣言表。
var bostonBidTable = [BostonBidLevelCount]bostonBidSpec{
	BostonBidPass:              {kind: BostonKindPass, name: "pass"},
	BostonBidFive:              {kind: BostonKindTricks, tricks: 5, canCallPartner: true, payout: 1, name: "five"},
	BostonBidSix:               {kind: BostonKindTricks, tricks: 6, canCallPartner: true, payout: 2, name: "six"},
	BostonBidLittleMisere:      {kind: BostonKindMisere, tricks: 0, payout: 3, name: "littleMisere"},
	BostonBidSeven:             {kind: BostonKindTricks, tricks: 7, canCallPartner: true, payout: 4, name: "seven"},
	BostonBidPiccolissimo:      {kind: BostonKindPiccolissimo, tricks: 1, payout: 5, name: "piccolissimo"},
	BostonBidEight:             {kind: BostonKindTricks, tricks: 8, canCallPartner: true, payout: 6, name: "eight"},
	BostonBidGrandMisere:       {kind: BostonKindMisere, tricks: 0, payout: 7, name: "grandMisere"},
	BostonBidNine:              {kind: BostonKindTricks, tricks: 9, canCallPartner: true, payout: 8, name: "nine"},
	BostonBidLittleMisereTable: {kind: BostonKindMisere, tricks: 0, exposed: true, payout: 9, name: "littleMisereTable"},
	BostonBidTen:               {kind: BostonKindTricks, tricks: 10, canCallPartner: true, payout: 10, name: "ten"},
	BostonBidGrandMisereTable:  {kind: BostonKindMisere, tricks: 0, exposed: true, payout: 12, name: "grandMisereTable"},
	BostonBidEleven:            {kind: BostonKindTricks, tricks: 11, payout: 14, name: "eleven"},
	BostonBidTwelve:            {kind: BostonKindTricks, tricks: 12, payout: 16, name: "twelve"},
	BostonBidChelem:            {kind: BostonKindTricks, tricks: 13, payout: 20, name: "chelem"},
	BostonBidChelemTable:       {kind: BostonKindTricks, tricks: 13, exposed: true, payout: 30, name: "chelemTable"},
}

// bostonSpecOf は宣言の性質を返す (範囲外はパス扱い)。
func bostonSpecOf(level BostonBidLevel) bostonBidSpec {
	if level < 0 || level >= BostonBidLevelCount {
		return bostonBidTable[BostonBidPass]
	}
	return bostonBidTable[level]
}

// BostonBidKindOf は宣言の型を返す。
func BostonBidKindOf(level BostonBidLevel) BostonBidKind { return bostonSpecOf(level).kind }

// BostonBidTricks は宣言の目標トリック数を返す。
func BostonBidTricks(level BostonBidLevel) int { return bostonSpecOf(level).tricks }

// BostonBidIsExposed は第 1 トリックのあと手札を公開する宣言かを返す。
//
// **公開はオプションではなく別の宣言。**issue の「ミゼール成立時は…オープン
// ハンドの場合あり」は誤りで、on the Table はそれ自体が 1 つ上の宣言である。
func BostonBidIsExposed(level BostonBidLevel) bool { return bostonSpecOf(level).exposed }

// BostonBidCanCallPartner はパートナーを指名できる宣言かを返す。
//
// **トリック数の宣言 (5〜10) だけ。**ミゼール系・ピッコリッシモ・スラムは
// 単独で 3 人を相手にする。issue の「各自個人戦」は正確ではない。
func BostonBidCanCallPartner(level BostonBidLevel) bool { return bostonSpecOf(level).canCallPartner }

// BostonBidPayout は達成時に各相手から受け取る額を返す。
func BostonBidPayout(level BostonBidLevel) int { return bostonSpecOf(level).payout }

// BostonBidName は表示用の識別子を返す。
func BostonBidName(level BostonBidLevel) string { return bostonSpecOf(level).name }

// BostonBidNeedsTrump は切札スートの指定が要る宣言かを返す。
//
// **ミゼール系とピッコリッシモは切札なし。**
func BostonBidNeedsTrump(level BostonBidLevel) bool {
	return bostonSpecOf(level).kind == BostonKindTricks
}

// BostonBidSucceeded は取ったトリック数で宣言が達成されたかを返す。
//
// **型ごとに条件が違う。**トリック宣言は「以上」、ミゼールは「0 ちょうど」、
// ピッコリッシモは「**1 ちょうど**」。ピッコリッシモは 0 でも失敗である。
func BostonBidSucceeded(level BostonBidLevel, won int) bool {
	spec := bostonSpecOf(level)
	switch spec.kind {
	case BostonKindTricks:
		return won >= spec.tricks
	case BostonKindMisere:
		return won == 0
	case BostonKindPiccolissimo:
		return won == 1
	}
	return false
}
