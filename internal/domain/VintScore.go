//go:build !js || !wasm || extra3

package domain

// Vint の宣言スート。
//
// **ブリッジとは序列が違う。**♠ < ♣ < ♦ < ♥ < NT で、**♠ が最弱**である。
// ブリッジの ♣<♦<♥<♠<NT を持ち込むと競りが逆になる。
const (
	// VintDenomSpade ♠ (最弱)
	VintDenomSpade = iota
	// VintDenomClub ♣
	VintDenomClub
	// VintDenomDiamond ♦
	VintDenomDiamond
	// VintDenomHeart ♥
	VintDenomHeart
	// VintDenomNoTrump ノートランプ (最強)
	VintDenomNoTrump
	// VintDenomCount 宣言スートの総数
	VintDenomCount
)

// vintBaseTrickValue はレベル 1 のトリック単価。
//
// **スートとレベルの両方で決まる。**レベルが 1 上がるごとに +10。
var vintBaseTrickValue = [VintDenomCount]int{
	VintDenomSpade:   4,
	VintDenomClub:    6,
	VintDenomDiamond: 8,
	VintDenomHeart:   10,
	VintDenomNoTrump: 12,
}

// VintMinLevel / VintMaxLevel は宣言レベルの範囲。
const (
	VintMinLevel = 1
	VintMaxLevel = 7
)

// VintGameTarget は 1 ゲームに要る線下の点。
const VintGameTarget = 500

// VintFirstGameBonus / VintRubberBonus はラバーのボーナス。
const (
	// VintFirstGameBonus は先に 1 ゲーム取った側の線上ボーナス。
	VintFirstGameBonus = 500
	// VintRubberBonus は 2 ゲーム目 (ラバー) を取った側の線上ボーナス。
	VintRubberBonus = 1000
)

// VintUndertrickUnit は未達 1 トリックあたりの倍率。
//
// ペナルティは **不足数 × 宣言レベル × 500**。
const VintUndertrickUnit = 500

// VintHonourMin はオナーが得点になる最少枚数。
//
// **3 枚以上から。**2 枚以下は 0 点。issue の「保有数に応じた追加ボーナス」は
// この下限に触れていない。
const VintHonourMin = 3

// VintAceMultiplier はエース 1 枚あたりの倍率 (トリック単価 × これ)。
const VintAceMultiplier = 10

// VintTrickValue は宣言のトリック単価を返す。
//
// **スート単価 + (レベル - 1) × 10。**例: 2♠ = 4 + 10 = 14、3♠ = 24。
func VintTrickValue(denom, level int) int {
	if denom < 0 || denom >= VintDenomCount || level < VintMinLevel || level > VintMaxLevel {
		return 0
	}
	return vintBaseTrickValue[denom] + (level-1)*10
}

// VintBidRank は宣言の序列を返す。値が大きいほど強い。
func VintBidRank(denom, level int) int {
	if denom < 0 || denom >= VintDenomCount || level < VintMinLevel || level > VintMaxLevel {
		return 0
	}
	return level*VintDenomCount + denom
}

// VintDenomToSuit は宣言スートを Card の design 値に変換する (NT は 0)。
func VintDenomToSuit(denom int) int {
	switch denom {
	case VintDenomSpade:
		return CardDesignSpade
	case VintDenomClub:
		return CardDesignClover
	case VintDenomDiamond:
		return CardDesignDiamond
	case VintDenomHeart:
		return CardDesignHeart
	}
	return 0
}

// VintHonourBonus はオナー枚数に応じた線上の点を返す。
//
// **3 枚未満は 0。**3 枚で単価 × 20、4 枚で × 30、5 枚で × 40。
func VintHonourBonus(count, trickValue int) int {
	if count < VintHonourMin {
		return 0
	}
	switch {
	case count >= 5:
		return trickValue * 40
	case count == 4:
		return trickValue * 30
	default:
		return trickValue * 20
	}
}

// IsVintHonour は切札のオナー (A K Q J 10) かを返す。
func IsVintHonour(c *Card, trumpSuit int) bool {
	if c == nil || trumpSuit == 0 || c.GetDesign() != trumpSuit {
		return false
	}
	switch c.GetValue() {
	case 1, 13, 12, 11, 10:
		return true
	}
	return false
}

// IsVintAce はエースかを返す。
func IsVintAce(c *Card) bool { return c != nil && c.GetValue() == 1 }

// VintAceBonus はエースの枚数差から線上の点を返す。
//
// **多く持っている側が枚数ぶん全部を取る。**2 対 2 で分かれた場合は、トリックを
// 多く取った側が 4 枚ぶん取る (呼び出し側が tieToUs で渡す)。
// 戻り値は (自陣の点, 相手の点)。
func VintAceBonus(ours, theirs, trickValue int, tieToUs bool) (int, int) {
	total := (ours + theirs) * VintAceMultiplier * trickValue
	switch {
	case ours > theirs:
		return total, 0
	case theirs > ours:
		return 0, total
	}
	// 同数はトリックの多い側が総取り。
	if tieToUs {
		return total, 0
	}
	return 0, total
}
