//go:build test

package domain

import "testing"

func TestActionLogKeyHelpersCoverEveryBranch(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"german mussfrage", germanSoloBidKey(GermanSoloBidMussfrage), "germansolo.bidMussfrage"},
		{"german frage", germanSoloBidKey(GermanSoloBidFrage), "germansolo.bidFrage"},
		{"german solo", germanSoloBidKey(GermanSoloBidSolo), "germansolo.bidSolo"},
		{"german tout", germanSoloBidKey(GermanSoloBidTout), "germansolo.bidTout"},
		{"german none", germanSoloBidKey(GermanSoloBidNone), "germansolo.bidNone"},
		{"ombre entrar", ombreBidKey(OmbreBidEntrar), "ombre.bidEntrar"},
		{"ombre solo", ombreBidKey(OmbreBidSolo), "ombre.bidSolo"},
		{"ombre pass", ombreBidKey(OmbreBidNone), "ombre.bidPass"},
		{"ombre sacar", ombreOutcomeKey(OmbreOutcomeSacar), "ombre.outcomeSacar"},
		{"ombre puesta", ombreOutcomeKey(OmbreOutcomePuesta), "ombre.outcomePuesta"},
		{"ombre codille", ombreOutcomeKey(OmbreOutcomeCodille), "ombre.outcomeCodille"},
		{"ombre outcome none", ombreOutcomeKey(OmbreOutcomeNone), "ombre.outcomeNone"},
		{"quadrille entrar", quadrilleBidKey(QuadrilleBidEntrar), "quadrille.bidEntrar"},
		{"quadrille solo", quadrilleBidKey(QuadrilleBidSolo), "quadrille.bidSolo"},
		{"quadrille pass", quadrilleBidKey(QuadrilleBidNone), "quadrille.bidPass"},
		{"quadrille sacar", quadrilleOutcomeKey(QuadrilleOutcomeSacar), "quadrille.outcomeSacar"},
		{"quadrille puesta", quadrilleOutcomeKey(QuadrilleOutcomePuesta), "quadrille.outcomePuesta"},
		{"quadrille codille", quadrilleOutcomeKey(QuadrilleOutcomeCodille), "quadrille.outcomeCodille"},
		{"quadrille outcome none", quadrilleOutcomeKey(QuadrilleOutcomeNone), "quadrille.outcomeNone"},
		{"ulti party", ultiContractKey(UltiContractParty), "ulti.contractParty"},
		{"ulti betli", ultiContractKey(UltiContractBetli), "ulti.contractBetli"},
		{"ulti durchmarsch", ultiContractKey(UltiContractDurchmarsch), "ulti.contractDurchmarsch"},
		{"ulti ulti", ultiContractKey(UltiContractUlti), "ulti.contractUlti"},
		{"ulti none", ultiContractKey(UltiContractNone), "ulti.contractNone"},
		{"truco truco", trucoLevelKey(TrucoLevelTruco), "truco.levelTruco"},
		{"truco retruco", trucoLevelKey(TrucoLevelRetruco), "truco.levelRetruco"},
		{"truco vale cuatro", trucoLevelKey(TrucoLevelValeCuatro), "truco.levelValeCuatro"},
		{"truco none", trucoLevelKey(TrucoLevelNone), "truco.levelNone"},
		{"put put", putLevelKey(PutLevelPut), "put.levelPut"},
		{"put none", putLevelKey(PutLevelNone), "put.levelNone"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
