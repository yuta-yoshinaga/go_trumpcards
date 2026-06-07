//go:build !js || !wasm || casino

package domain

// shortDeckStyleParamsMap ショートデック用スタイルごとのパラメータ
// ショートデックは36枚デッキのため平均ハンド強度が高く、閾値を引き上げている
var shortDeckStyleParamsMap = map[HoldemPlayStyle]cpuStyleParams{
	HoldemStyleTAG: {
		aggressive: true, bluffRate: 15,
		preFlopFoldThreshold: 55, preFlopFoldCompound: false,
		preFlopRaiseThreshold: 75, preFlopRaisePotPct: 75,
		postFlopRaiseRank: ShortDeckHandThreeOfAKind, postFlopRaisePotPct: 66,
		postFlopCondCallRank: ShortDeckHandOnePair, postFlopFallbackFold: true,
	},
	HoldemStyleLAP: {
		aggressive: false, bluffRate: 5,
		preFlopFoldThreshold: 25, preFlopFoldCompound: true, preFlopFoldCallMult: 2,
		preFlopBluffPotPct:   50,
		postFlopPassFoldRank: ShortDeckHandHighCard, postFlopPassFoldMult: 3,
		postFlopBluffPotPct: 33,
	},
	HoldemStyleTAP: {
		aggressive: false, bluffRate: 5,
		preFlopFoldThreshold: 45, preFlopFoldCompound: false,
		preFlopBluffPotPct:   50,
		postFlopPassFoldRank: ShortDeckHandHighCard, postFlopPassFoldMult: -1,
		postFlopBluffPotPct: 33,
	},
	HoldemStyleLAG: {
		aggressive: true, bluffRate: 30,
		preFlopFoldThreshold: 25, preFlopFoldCompound: true, preFlopFoldCallMult: 3,
		preFlopRaiseThreshold: 55, preFlopRaisePotPct: 100,
		postFlopRaiseRank: ShortDeckHandOnePair, postFlopRaisePotPct: 100,
		postFlopFallbackFold: false,
		postFlopAggrFoldRank: ShortDeckHandHighCard, postFlopAggrFoldMult: 4,
	},
}
