//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func assertLiteralLogCodesInLocales(t *testing.T, codes []string) {
	t.Helper()
	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		for _, code := range codes {
			translated := i18n.Tf(code)
			require.NotEmpty(t, translated, "%s: %s", lang, code)
			require.NotEqual(t, code, translated, "%s", lang)
		}
	}
	i18n.SetLang("ja")
}

func assertUniqueLogCode(t *testing.T, seen map[string]bool, code string) {
	t.Helper()
	require.NotEmpty(t, code)
	require.False(t, seen[code], "duplicate log code: %s", code)
	seen[code] = true
}

func TestNewActionLogCodesRenderInBothLocales(t *testing.T) {
	codes := []string{
		"sjavs.log.deal", "sjavs.log.redeal", "sjavs.log.bid", "sjavs.log.pass", "sjavs.log.trump",
		"sjavs.log.play", "sjavs.log.trick", "sjavs.log.tie", "sjavs.log.score", "sjavs.log.rubber",
		"windmill.log.moveSailToCenter", "windmill.log.moveSailToCorner", "windmill.log.moveWasteToCenter",
		"windmill.log.moveWasteToCorner", "windmill.log.moveCornerToCenter",
		"napoleonssquare.log.moveWasteToTableau", "napoleonssquare.log.moveWasteToFoundation",
		"napoleonssquare.log.moveTableauToTableau", "napoleonssquare.log.moveTableauToFoundation",
	}
	assertLiteralLogCodesInLocales(t, codes)
}

func TestActionLogCodeFunctionsCoverAllCombinations(t *testing.T) {
	seen := make(map[string]bool)
	var codes []string
	add := func(code string) {
		assertUniqueLogCode(t, seen, code)
		codes = append(codes, code)
	}

	for contract := RikkenContractNone; contract <= RikkenContractMax; contract++ {
		add(rikkenBidLogCode(contract))
		add(rikkenPlayLogCode(contract))
		add(rikkenResultLogCode(contract, false))
		add(rikkenResultLogCode(contract, true))
	}
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		add(rikkenTrumpLogCode(suit))
	}
	add(rikkenTrumpLogCode(RikkenNoTrump))

	for kind := SixBidSoloMinBid; kind <= SixBidSoloMaxBid; kind++ {
		add(sixBidSoloBidLogCode(kind))
		for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
			add(sixBidSoloDeclareLogCode(kind, suit))
		}
	}

	for kind := GuandanComboSingle; kind <= GuandanComboJokerBomb; kind++ {
		add(guandanPlayLogCode(kind))
	}

	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for strength := 1; strength <= 2; strength++ {
			add(shengJiDeclareLogCode(suit, strength))
		}
	}

	for bet := TrenteEtQuaranteBetNoir; bet <= TrenteEtQuaranteBetInverse; bet++ {
		add(trenteEtQuaranteBetLogCode(bet))
		for _, row := range []int{TrenteEtQuaranteRowNoir, TrenteEtQuaranteRowRouge} {
			for _, result := range []TrenteEtQuaranteResult{TrenteEtQuaranteResultLose, TrenteEtQuaranteResultWin} {
				add(trenteEtQuaranteResultLogCode(row, bet, result))
			}
		}
	}

	for comp := PopeJoanAce; comp < PopeJoanCompartmentCount; comp++ {
		add(popeJoanAwardLogCode(comp, false))
		add(popeJoanAwardLogCode(comp, true))
	}

	assertLiteralLogCodesInLocales(t, codes)
}

func TestActionLogLiteralCodesPreserveAllExistingKeys(t *testing.T) {
	expected := []string{
		"rikken.log.bid.none", "rikken.log.play.none", "rikken.log.result.noneFailed", "rikken.log.result.noneMade",
		"rikken.log.bid.rik", "rikken.log.play.rik", "rikken.log.result.rikFailed", "rikken.log.result.rikMade",
		"rikken.log.bid.misere", "rikken.log.play.misere", "rikken.log.result.misereFailed", "rikken.log.result.misereMade",
		"rikken.log.bid.solo", "rikken.log.play.solo", "rikken.log.result.soloFailed", "rikken.log.result.soloMade",
		"rikken.log.bid.openMisere", "rikken.log.play.openMisere", "rikken.log.result.openMisereFailed", "rikken.log.result.openMisereMade",
		"rikken.log.trump.spade", "rikken.log.trump.clover", "rikken.log.trump.heart", "rikken.log.trump.diamond", "rikken.log.trump.notrump",
		"sixbidsolo.log.bidSolo", "sixbidsolo.log.bidHeartSolo", "sixbidsolo.log.bidMisere", "sixbidsolo.log.bidGuarantee", "sixbidsolo.log.bidSpreadMisere", "sixbidsolo.log.bidCallSolo",
		"sixbidsolo.log.declareSoloS", "sixbidsolo.log.declareSoloC", "sixbidsolo.log.declareSoloH", "sixbidsolo.log.declareSoloD",
		"sixbidsolo.log.declareHeartSoloS", "sixbidsolo.log.declareHeartSoloC", "sixbidsolo.log.declareHeartSoloH", "sixbidsolo.log.declareHeartSoloD",
		"sixbidsolo.log.declareMisereS", "sixbidsolo.log.declareMisereC", "sixbidsolo.log.declareMisereH", "sixbidsolo.log.declareMisereD",
		"sixbidsolo.log.declareGuaranteeS", "sixbidsolo.log.declareGuaranteeC", "sixbidsolo.log.declareGuaranteeH", "sixbidsolo.log.declareGuaranteeD",
		"sixbidsolo.log.declareSpreadMisereS", "sixbidsolo.log.declareSpreadMisereC", "sixbidsolo.log.declareSpreadMisereH", "sixbidsolo.log.declareSpreadMisereD",
		"sixbidsolo.log.declareCallSoloS", "sixbidsolo.log.declareCallSoloC", "sixbidsolo.log.declareCallSoloH", "sixbidsolo.log.declareCallSoloD",
		"guandan.log.playSingle", "guandan.log.playPair", "guandan.log.playTriple", "guandan.log.playFullHouse", "guandan.log.playStraight", "guandan.log.playPlate", "guandan.log.playTube", "guandan.log.playBomb", "guandan.log.playStraightFlush", "guandan.log.playJokerBomb",
		"shengji.log.declareSpadeSingle", "shengji.log.declareSpadePair", "shengji.log.declareCloverSingle", "shengji.log.declareCloverPair", "shengji.log.declareHeartSingle", "shengji.log.declareHeartPair", "shengji.log.declareDiamondSingle", "shengji.log.declareDiamondPair",
		"trenteetquarante.log.resultNoirNoirLose", "trenteetquarante.log.resultNoirNoirWin", "trenteetquarante.log.resultNoirRougeLose", "trenteetquarante.log.resultNoirRougeWin", "trenteetquarante.log.resultNoirCouleurLose", "trenteetquarante.log.resultNoirCouleurWin", "trenteetquarante.log.resultNoirInverseLose", "trenteetquarante.log.resultNoirInverseWin",
		"trenteetquarante.log.resultRougeNoirLose", "trenteetquarante.log.resultRougeNoirWin", "trenteetquarante.log.resultRougeRougeLose", "trenteetquarante.log.resultRougeRougeWin", "trenteetquarante.log.resultRougeCouleurLose", "trenteetquarante.log.resultRougeCouleurWin", "trenteetquarante.log.resultRougeInverseLose", "trenteetquarante.log.resultRougeInverseWin",
	}
	var codes []string
	for _, contract := range []int{RikkenContractNone, RikkenContractRik, RikkenContractMisere, RikkenContractSolo, RikkenContractOpenMisere} {
		codes = append(codes, rikkenBidLogCode(contract), rikkenPlayLogCode(contract))
		codes = append(codes, rikkenResultLogCode(contract, false), rikkenResultLogCode(contract, true))
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond, RikkenNoTrump} {
		codes = append(codes, rikkenTrumpLogCode(suit))
	}

	for kind := SixBidSoloMinBid; kind <= SixBidSoloMaxBid; kind++ {
		codes = append(codes, sixBidSoloBidLogCode(kind))
		for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
			codes = append(codes, sixBidSoloDeclareLogCode(kind, suit))
		}
	}

	for kind := GuandanComboSingle; kind <= GuandanComboJokerBomb; kind++ {
		codes = append(codes, guandanPlayLogCode(kind))
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, strength := range []int{1, 2} {
			codes = append(codes, shengJiDeclareLogCode(suit, strength))
		}
	}
	for _, row := range []int{TrenteEtQuaranteRowNoir, TrenteEtQuaranteRowRouge} {
		for bet := TrenteEtQuaranteBetNoir; bet <= TrenteEtQuaranteBetInverse; bet++ {
			codes = append(codes,
				trenteEtQuaranteResultLogCode(row, bet, TrenteEtQuaranteResultLose),
				trenteEtQuaranteResultLogCode(row, bet, TrenteEtQuaranteResultWin))
		}
	}

	require.Len(t, codes, 89)
	require.ElementsMatch(t, expected, codes)
	assertLiteralLogCodesInLocales(t, codes)
}

func TestPopeJoanAwardLiteralCodesPreserveAllExistingKeys(t *testing.T) {
	codes := make([]string, 0, PopeJoanCompartmentCount*2)
	for comp := PopeJoanAce; comp < PopeJoanCompartmentCount; comp++ {
		codes = append(codes, popeJoanAwardLogCode(comp, false), popeJoanAwardLogCode(comp, true))
	}
	require.Len(t, codes, 16)
	assertLiteralLogCodesInLocales(t, codes)
}

func TestActionLogCodeFunctionsAreUsedByProductionPaths(t *testing.T) {
	assertLastCode := func(t *testing.T, logs []*ActionLogEntry, want string) {
		t.Helper()
		require.Len(t, logs, 1)
		require.Equal(t, want, logs[0].DetailCode)
	}

	t.Run("Guandan", func(t *testing.T) {
		g := &Guandan{}
		g.addGuandanPlayLog(0, GuandanComboBomb, nil)
		assertLastCode(t, g.GetActionLog(), "guandan.log.playBomb")
	})
	t.Run("Rikken", func(t *testing.T) {
		g := &Rikken{}
		g.addRikkenBidLog(0, RikkenContractSolo)
		assertLastCode(t, g.GetActionLog(), "rikken.log.bid.solo")
		g.actionLog = nil
		g.addRikkenTrumpLog(0, CardDesignHeart)
		assertLastCode(t, g.GetActionLog(), "rikken.log.trump.heart")
		g.actionLog = nil
		g.addRikkenPlayLog(RikkenContractMisere)
		assertLastCode(t, g.GetActionLog(), "rikken.log.play.misere")
	})
	t.Run("ShengJi", func(t *testing.T) {
		g := &ShengJi{}
		g.addShengJiDeclareLog(0, CardDesignClover, 2)
		assertLastCode(t, g.GetActionLog(), "shengji.log.declareCloverPair")
	})
	t.Run("SixBidSolo", func(t *testing.T) {
		g := &SixBidSolo{}
		g.addSixBidSoloBidLog(0, SixBidSoloBidGuarantee)
		assertLastCode(t, g.GetActionLog(), "sixbidsolo.log.bidGuarantee")
		g.actionLog = nil
		g.addSixBidSoloDeclareLog(0, SixBidSoloBidGuarantee, CardDesignDiamond)
		assertLastCode(t, g.GetActionLog(), "sixbidsolo.log.declareGuaranteeD")
	})
	t.Run("PopeJoan", func(t *testing.T) {
		g := &PopeJoan{}
		g.addPopeJoanAwardLog(0, PopeJoanPope, true, 3)
		assertLastCode(t, g.GetActionLog(), "popejoan.log.award.popeFromTurnUp")
	})
	t.Run("TrenteEtQuarante", func(t *testing.T) {
		g := &TrenteEtQuarante{}
		g.addTrenteEtQuaranteResultLog(TrenteEtQuaranteRowRouge, TrenteEtQuaranteBetNoir, TrenteEtQuaranteResultLose, 0)
		assertLastCode(t, g.GetActionLog(), "trenteetquarante.log.resultRougeNoirLose")
	})
}
