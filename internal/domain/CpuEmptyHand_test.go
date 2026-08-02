//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// **手札の尽きた席に手番が回ってもサーバを落とさない。**
//
// 各ゲームの CPU セレクタは「出せる札が無い」とき 0 を返し、Player.RemoveCard は
// 範囲外で nil を返す (Player.go:89-96)。この 2 つが噛み合うと playCard が nil を
// 触り、**HTTP ハンドラごとパニックする**。E2E が GongZhu で実際にサーバを落とし
// (#4606)、同じ形が 52 ゲームにあった。
//
// **代表を選ばず全件を通す。**「同じ形だから同じはず」で括ったせいで探索が 3 回とも
// 少なく数えたのがこの修正の経緯で、実際 4 件だけ見ていた版では Briscola と Tysiac が
// **まだ落ちるのを見逃していた**。あの 2 つは CpuPlay ではなくセレクタの内部で
// 落ちていて、呼び出し側のガードでは届かない。
func TestCpuPlayWithEmptyHandDoesNotPanic(t *testing.T) {
	cases := []struct {
		name  string
		drive func()
	}{
		{"AllFours", func() {
			g := domain.NewDefaultAllFours()
			g.Reset()
			g.SetPhase(domain.AllFoursPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Belote", func() {
			g := domain.NewDefaultBelote()
			g.Reset()
			g.SetPhase(domain.BelotePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Bezique", func() {
			g := domain.NewDefaultBezique()
			g.Reset()
			g.SetPhase(domain.BeziquePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"BidWhist", func() {
			g := domain.NewDefaultBidWhist()
			g.Reset()
			g.SetPhase(domain.BidWhistPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Bridge", func() {
			g := domain.NewDefaultBridge()
			g.Reset()
			g.SetPhase(domain.BridgePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Briscola", func() {
			g := domain.NewDefaultBriscola()
			g.Reset()
			g.SetPhase(domain.BriscolaPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Calabresella", func() {
			g := domain.NewDefaultCalabresella()
			g.Reset()
			g.SetPhase(domain.CalabresellaPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"CallBreak", func() {
			g := domain.NewDefaultCallBreak()
			g.Reset()
			g.SetPhase(domain.CallBreakPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"CatchTen", func() {
			g := domain.NewDefaultCatchTen()
			g.Reset()
			g.SetPhase(domain.CatchTenPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Cego", func() {
			g := domain.NewDefaultCego()
			g.Reset()
			g.SetPhase(domain.CegoPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"CourtPiece", func() {
			g := domain.NewDefaultCourtPiece()
			g.Reset()
			g.SetPhase(domain.CourtPiecePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Doppelkopf", func() {
			g := domain.NewDefaultDoppelkopf()
			g.Reset()
			g.SetPhase(domain.DoppelkopfPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Ecarte", func() {
			g := domain.NewDefaultEcarte()
			g.Reset()
			g.SetPhase(domain.EcartePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Euchre", func() {
			g := domain.NewDefaultEuchre()
			g.Reset()
			g.SetPhase(domain.EuchrePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"FiveHundred", func() {
			g := domain.NewDefaultFiveHundred()
			g.Reset()
			g.SetPhase(domain.FiveHundredPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"FortyFives", func() {
			g := domain.NewDefaultFortyFives()
			g.Reset()
			g.SetPhase(domain.FortyFivesPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"FrenchTarot", func() {
			g := domain.NewDefaultFrenchTarot()
			g.Reset()
			g.SetPhase(domain.FrenchTarotPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Gaigel", func() {
			g := domain.NewDefaultGaigel()
			g.Reset()
			g.SetPhase(domain.GaigelPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Hearts", func() {
			g := domain.NewDefaultHearts()
			g.Reset()
			g.SetPhase(domain.HeartsPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Jass", func() {
			g := domain.NewDefaultJass()
			g.Reset()
			g.SetPhase(domain.JassPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Klaverjas", func() {
			g := domain.NewDefaultKlaverjas()
			g.Reset()
			g.SetPhase(domain.KlaverjasPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"KnockoutWhist", func() {
			g := domain.NewDefaultKnockoutWhist()
			g.Reset()
			g.SetPhase(domain.KnockoutWhistPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Koenigrufen", func() {
			g := domain.NewDefaultKoenigrufen()
			g.Reset()
			g.SetPhase(domain.KoenigrufenPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Manille", func() {
			g := domain.NewDefaultManille()
			g.Reset()
			g.SetPhase(domain.ManillePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Marias", func() {
			g := domain.NewDefaultMarias()
			g.Reset()
			g.SetPhase(domain.MariasPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Mighty", func() {
			g := domain.NewDefaultMighty()
			g.Reset()
			g.SetPhase(domain.MightyPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Nap", func() {
			g := domain.NewDefaultNap()
			g.Reset()
			g.SetPhase(domain.NapPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Napoleon", func() {
			g := domain.NewDefaultNapoleon()
			g.Reset()
			g.SetPhase(domain.NapoleonPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"NinetyNine", func() {
			g := domain.NewDefaultNinetyNine()
			g.Reset()
			g.SetPhase(domain.NinetyNinePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"OhHell", func() {
			g := domain.NewDefaultOhHell()
			g.Reset()
			g.SetPhase(domain.OhHellPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Ombre", func() {
			g := domain.NewDefaultOmbre()
			g.Reset()
			g.SetPhase(domain.OmbrePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Pitch", func() {
			g := domain.NewDefaultPitch()
			g.Reset()
			g.SetPhase(domain.PitchPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Preference", func() {
			g := domain.NewDefaultPreference()
			g.Reset()
			g.SetPhase(domain.PreferencePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Rook", func() {
			g := domain.NewDefaultRook()
			g.Reset()
			g.SetPhase(domain.RookPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Scarto", func() {
			g := domain.NewDefaultScarto()
			g.Reset()
			g.SetPhase(domain.ScartoPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Schnapsen", func() {
			g := domain.NewDefaultSchnapsen()
			g.Reset()
			g.SetPhase(domain.SchnapsenPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Sedma", func() {
			g := domain.NewDefaultSedma()
			g.Reset()
			g.SetPhase(domain.SedmaPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Skat", func() {
			g := domain.NewDefaultSkat()
			g.Reset()
			g.SetPhase(domain.SkatPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"SoloWhist", func() {
			g := domain.NewDefaultSoloWhist()
			g.Reset()
			g.SetPhase(domain.SoloWhistPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Spades", func() {
			g := domain.NewDefaultSpades()
			g.Reset()
			g.SetPhase(domain.SpadesPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"SpoilFive", func() {
			g := domain.NewDefaultSpoilFive()
			g.Reset()
			g.SetPhase(domain.SpoilFivePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Sueca", func() {
			g := domain.NewDefaultSueca()
			g.Reset()
			g.SetPhase(domain.SuecaPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Tarneeb", func() {
			g := domain.NewDefaultTarneeb()
			g.Reset()
			g.SetPhase(domain.TarneebPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Tressette", func() {
			g := domain.NewDefaultTressette()
			g.Reset()
			g.SetPhase(domain.TressettePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Tute", func() {
			g := domain.NewDefaultTute()
			g.Reset()
			g.SetPhase(domain.TutePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"TwentyNine", func() {
			g := domain.NewDefaultTwentyNine()
			g.Reset()
			g.SetPhase(domain.TwentyNinePhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"TwoTenJack", func() {
			g := domain.NewDefaultTwoTenJack()
			g.Reset()
			g.SetPhase(domain.TwoTenJackPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Tysiac", func() {
			g := domain.NewDefaultTysiac()
			g.Reset()
			g.SetPhase(domain.TysiacPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Ulti", func() {
			g := domain.NewDefaultUlti()
			g.Reset()
			g.SetPhase(domain.UltiPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Watten", func() {
			g := domain.NewDefaultWatten()
			g.Reset()
			g.SetPhase(domain.WattenPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Whist", func() {
			g := domain.NewDefaultWhist()
			g.Reset()
			g.SetPhase(domain.WhistPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Wizard", func() {
			g := domain.NewDefaultWizard()
			g.Reset()
			g.SetPhase(domain.WizardPhasePlay)
			p := g.GetPlayer(1)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("手札の無い CPU でパニックした: %v", r)
				}
			}()
			c.drive()
		})
	}
}
