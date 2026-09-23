package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type adr0040ActionLogger interface {
	Reset()
	GetActionLog() []*ActionLogEntry
}

func assertADR0040DetailCode(t *testing.T, game adr0040ActionLogger, code string) *ActionLogEntry {
	t.Helper()
	for _, entry := range game.GetActionLog() {
		if entry.DetailCode == code {
			return entry
		}
	}
	t.Fatalf("action log entry %q not found", code)
	return nil
}

func TestADR0040NewlyMigratedGamesUseDetailCodes(t *testing.T) {
	tests := []struct {
		name   string
		game   adr0040ActionLogger
		code   string
		params []string
	}{
		{"HoneymoonBridge", NewDefaultHoneymoonBridge(), "honeymoonbridge.log.deal", []string{"round", "stock"}},
		{"Kaiser", NewDefaultKaiser(), "kaiser.log.deal", []string{"kitty"}},
		{"Kille", NewDefaultKille(), "kille.log.deal", []string{"pot"}},
		{"Klaberjass", NewDefaultKlaberjass(), "klaberjass.log.deal", nil},
		{"Loba", NewDefaultLoba(), "loba.log.deal", nil},
		{"Mushi", NewDefaultMushi(), "mushi.log.deal", []string{"round"}},
		{"NainJaune", NewDefaultNainJaune(), "nainjaune.log.deal", []string{"talon"}},
		{"Pig", NewDefaultPig(), "pig.log.start", []string{"players", "cards"}},
		{"Poch", NewDefaultPoch(), "poch.log.deal", []string{"suit"}},
		{"PopeJoan", NewDefaultPopeJoan(), "popejoan.log.deal", []string{"suit"}},
		{"Rikken", NewDefaultRikken(), "rikken.log.start", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.game.Reset()
			entry := assertADR0040DetailCode(t, tt.game, tt.code)
			for _, key := range tt.params {
				require.Contains(t, entry.DetailParams, key)
			}
		})
	}
}

func TestADR0040RemainingGamesUseDetailCodes(t *testing.T) {
	tests := []struct {
		name string
		game adr0040ActionLogger
		code string
	}{
		{"Bhabhi", NewDefaultBhabhi(), "bhabhi.log.deal"},
		{"BidEuchre", NewDefaultBidEuchre(), "bideuchre.log.deal"},
		{"Boston", NewDefaultBoston(), "boston.log.deal"},
		{"Botifarra", NewDefaultBotifarra(), "botifarra.log.start"},
		{"Bura", NewDefaultBura(), "bura.log.deal"},
		{"ChineseTen", NewDefaultChineseTen(), "chineseten.log.deal"},
		{"ColourWhist", NewDefaultColourWhist(), "colourwhist.log.start"},
		{"Cucumber", NewDefaultCucumber(), "cucumber.log.start"},
		{"Desmoche", NewDefaultDesmoche(), "desmoche.log.deal"},
		{"Goofspiel", NewDefaultGoofspiel(), "goofspiel.log.start"},
		{"Hasenpfeffer", NewDefaultHasenpfeffer(), "hasenpfeffer.log.deal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.game.Reset()
			require.NotEmpty(t, tt.game.GetActionLog())
			assertADR0040DetailCode(t, tt.game, tt.code)
		})
	}
}

func TestADR0040DelegatedGamesUseDetailCodes(t *testing.T) {
	tests := []struct {
		name   string
		game   adr0040ActionLogger
		code   string
		params []string
	}{
		{"RollingStone", NewDefaultRollingStone(), "rollingstone.log.start", []string{"players", "cards"}},
		{"SergeantMajor", NewDefaultSergeantMajor(), "sergeantmajor.log.deal", []string{"round", "kitty"}},
		{"Sjavs", NewDefaultSjavs(), "sjavs.log.deal", nil},
		{"Skitgubbe", NewDefaultSkitgubbe(), "skitgubbe.log.deal", nil},
		{"Snap", NewDefaultSnap(), "snap.log.start", []string{"players"}},
		{"StealingBundles", NewDefaultStealingBundles(), "stealingbundles.log.start", []string{"players"}},
		{"TeenDoPaanch", NewDefaultTeenDoPaanch(), "teendopaanch.log.deal", []string{"round"}},
		{"Toepen", NewDefaultToepen(), "toepen.log.deal", []string{"hand"}},
		{"Trex", NewDefaultTrex(), "trex.log.firstKingdom", nil},
		{"Vint", NewDefaultVint(), "vint.log.deal", []string{"cards"}},
		{"Zwicker", NewDefaultZwicker(), "zwicker.log.deal", []string{"cards"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.game.Reset()
			entry := assertADR0040DetailCode(t, tt.game, tt.code)
			for _, key := range tt.params {
				require.Contains(t, entry.DetailParams, key)
			}
		})
	}
}
