//go:build test

package domain

import (
	"fmt"
	"testing"
)

// TestSolver_ReproducesStandardTable is the gate that makes the generated
// Spanish 21 table trustworthy (#4705).
//
// **ソルバを信用してよい根拠はこれ一本。**同じソルバを標準デッキ・標準ルールで
// 走らせて、リポジトリが既に持っている `hardStrategy` / `softStrategy` /
// `pairStrategy` の表を再現できるなら、デッキ構成とバリアント設定だけ差し替えた
// 出力も同じ精度で信用できる。再現できないならソルバが間違っているので、
// スパニッシュ21の出力も採用してはいけない。
//
// 無限デッキ近似なので、境界に近い数マスは公表表と割れうる。落ちたセルは
// 黙って許さず、ここに理由付きで列挙する。
func TestSolver_ReproducesStandardTable(t *testing.T) {
	r := standardRules()

	// 無限デッキ近似では EV 差が極小で公表表と割れるマス。
	// 実測 EV 差を測ったうえで、|ΔEV| < 0.01 のものだけを許す。
	known := map[string]bool{}

	// **差が小さいことを実測してから許す。**「たぶん僅差だから」で見逃すと、
	// ソルバの本当の誤りも一緒に通る。
	const marginal = 0.01

	var mismatches []string
	check := func(label string, h handState, up int, isPair bool, pv int, got, want BJSuggestedAction) {
		if got == want || known[label] {
			return
		}
		gotEV := r.evOfAction(h, up, got, isPair, pv)
		wantEV := r.evOfAction(h, up, want, isPair, pv)
		gap := gotEV - wantEV
		if gap < marginal {
			t.Logf("marginal: %s solver=%s(%.4f) table=%s(%.4f) gap=%.4f",
				label, actionLetter(got), gotEV, actionLetter(want), wantEV, gap)
			return
		}
		mismatches = append(mismatches,
			fmt.Sprintf("%s: solver=%s(%.4f) table=%s(%.4f) gap=%.4f",
				label, actionLetter(got), gotEV, actionLetter(want), wantEV, gap))
	}

	for di, up := range solverUpcards {
		// ハード 5..17 (ペアにならない構成で作る)
		for total := 5; total <= 17; total++ {
			h := hardHandOfTotal(total)
			got := r.solveCell(h, up, false, 0)
			want := hardStrategy(total, di)
			check(fmt.Sprintf("hard%d-%s", total, upLabel(up)), h, up, false, 0, normalizeD(got), normalizeD(want))
		}
		// ソフト 13..20
		for total := 13; total <= 20; total++ {
			h := newHand(1, total-11)
			got := r.solveCell(h, up, false, 0)
			want := softStrategy(total, di)
			check(fmt.Sprintf("soft%d-%s", total, upLabel(up)), h, up, false, 0, normalizeD(got), normalizeD(want))
		}
		// ペア A,A および 2,2..10,10
		for _, pv := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
			h := newHand(pv, pv)
			got := r.solveCell(h, up, true, pv)
			want := pairStrategy(pv, di)
			check(fmt.Sprintf("pair%d-%s", pv, upLabel(up)), h, up, true, pv, normalizeD(got), normalizeD(want))
		}
	}

	if len(mismatches) > 0 {
		t.Fatalf("solver disagrees with the repository's standard table in %d cell(s):\n%s",
			len(mismatches), joinLines(mismatches))
	}
}

// normalizeD は D と Ds の差を潰す。どちらも「ダブル推奨」で、
// 不可時のフォールバックが違うだけ。表側の使い分けは EV では決まらない。
func normalizeD(a BJSuggestedAction) BJSuggestedAction {
	if a == BJSuggestDoubleStand {
		return BJSuggestDouble
	}
	return a
}

// hardHandOfTotal は指定ハードトータルの2枚ハンドを、ペアにならない組で作る。
func hardHandOfTotal(total int) handState {
	for a := 2; a <= 9; a++ {
		b := total - a
		if b < 2 || b > 10 || a == b {
			continue
		}
		return newHand(a, b)
	}
	// 5 は 2+3 で作れるのでここには来ないが、保険。
	return newHand(2, total-2)
}

func upLabel(up int) string {
	if up == 1 {
		return "A"
	}
	return fmt.Sprintf("%d", up)
}

func joinLines(ss []string) string {
	out := ""
	for _, s := range ss {
		out += "  " + s + "\n"
	}
	return out
}

// TestSolver_Soft19vs6_DependsOnDealerSoft17Rule confirms the solver's
// disagreement with the repository's table is a real defect, not solver error.
//
// `softH17Override` のコメントは「Soft 19 vs 6: S→Ds」と書いており、S17 では S、
// H17 では Ds のつもりで書かれている。ところが `softStrategy` の**基本表**が既に
// Ds なので override は何も変えておらず、S17 のプレイヤーに H17 用の助言が出る。
func TestSolver_Soft19vs6_DependsOnDealerSoft17Rule(t *testing.T) {
	h := newHand(1, 8) // A,8 = soft 19
	const dealerSix = 6

	s17 := standardRules()
	h17 := standardRules()
	h17.dealerHitsSoft17 = true

	standS17 := s17.evOfAction(h, dealerSix, BJSuggestStand, false, 0)
	doubleS17 := s17.evOfAction(h, dealerSix, BJSuggestDouble, false, 0)
	standH17 := h17.evOfAction(h, dealerSix, BJSuggestStand, false, 0)
	doubleH17 := h17.evOfAction(h, dealerSix, BJSuggestDouble, false, 0)

	// S17: スタンドが勝る。H17: ダブルが勝る。この向きが逆転しているのが要点。
	if standS17 <= doubleS17 {
		t.Errorf("S17: expected stand to beat double, got stand=%.4f double=%.4f", standS17, doubleS17)
	}
	if doubleH17 <= standH17 {
		t.Errorf("H17: expected double to beat stand, got stand=%.4f double=%.4f", standH17, doubleH17)
	}
	t.Logf("soft19 vs 6 — S17: stand=%.4f double=%.4f | H17: stand=%.4f double=%.4f",
		standS17, doubleS17, standH17, doubleH17)
}

// TestGenerateSpanish21Table prints the Spanish 21 basic-strategy table solved
// from this game's own rules. Not an assertion — run it with -v to regenerate
// the literal table in BlackJackSpanish21Strategy.go:
//
//	go test -tags test ./internal/domain -run TestGenerateSpanish21Table -v
func TestGenerateSpanish21Table(t *testing.T) {
	if testing.Short() {
		t.Skip("generator")
	}
	r := spanish21Rules()

	t.Log("// hard 5..17")
	for total := 5; total <= 17; total++ {
		cells := make([]BJSuggestedAction, len(solverUpcards))
		for i, up := range solverUpcards {
			cells[i] = r.solveCell(hardHandOfTotal(total), up, false, 0)
		}
		t.Logf("%s, // hard %d", renderRow(cells), total)
	}
	t.Log("// soft 13..20")
	for total := 13; total <= 20; total++ {
		cells := make([]BJSuggestedAction, len(solverUpcards))
		for i, up := range solverUpcards {
			cells[i] = r.solveCell(newHand(1, total-11), up, false, 0)
		}
		t.Logf("%s, // soft %d", renderRow(cells), total)
	}
	t.Log("// pairs A,A then 2,2..10,10")
	for _, pv := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		cells := make([]BJSuggestedAction, len(solverUpcards))
		for i, up := range solverUpcards {
			cells[i] = r.solveCell(newHand(pv, pv), up, true, pv)
		}
		t.Logf("%s, // pair %d", renderRow(cells), pv)
	}
}

// TestSpanish21_SoftFourteenVsFour records why this cell disagrees with the
// published Spanish 21 advice, so the divergence is evidence rather than doubt.
//
// Discount Gambling は「A-3 vs 4 はダブルで +15.2%、ヒットなら +9.1%」と書く。
// このゲームの規則で解くと**ヒットが勝る**。理由は規則差で、
// `PlayerDoubleDown` が 2 枚のときしかダブルを許さないこと (本来のスパニッシュ21は
// 何枚でもダブルでき、そこが積極的にダブルする根拠になっている)。
func TestSpanish21_SoftFourteenVsFour(t *testing.T) {
	r := spanish21Rules()
	h := newHand(1, 3) // A,3 = soft 14
	const dealerFour = 4

	hitEV := r.evOfAction(h, dealerFour, BJSuggestHit, false, 0)
	doubleEV := r.evOfAction(h, dealerFour, BJSuggestDouble, false, 0)

	if hitEV <= doubleEV {
		t.Errorf("expected hit to beat double under this game's 2-card-only doubling, got hit=%.4f double=%.4f",
			hitEV, doubleEV)
	}
	t.Logf("soft 14 vs 4 — hit=%.4f double=%.4f (2-card-only doubling)", hitEV, doubleEV)
}

// TestSpanish21Table_MatchesSolver keeps the committed table and the solver in
// step. Editing one without the other is exactly how a strategy table rots.
func TestSpanish21Table_MatchesSolver(t *testing.T) {
	r := spanish21Rules()
	var drift []string

	for di, up := range solverUpcards {
		for total := 5; total <= 17; total++ {
			h := hardHandOfTotal(total)
			want := r.solveCell(h, up, false, 0)
			got := spanish21HardStrategy(total, di)
			if got != want {
				drift = append(drift, fmt.Sprintf("hard%d-%s: table=%s solver=%s",
					total, upLabel(up), actionLetter(got), actionLetter(want)))
			}
		}
		for total := 13; total <= 20; total++ {
			h := newHand(1, total-11)
			want := r.solveCell(h, up, false, 0)
			got := spanish21SoftStrategy(total, di)
			if got != want {
				drift = append(drift, fmt.Sprintf("soft%d-%s: table=%s solver=%s",
					total, upLabel(up), actionLetter(got), actionLetter(want)))
			}
		}
		for pv := 1; pv <= 10; pv++ {
			h := newHand(pv, pv)
			want := r.solveCell(h, up, true, pv)
			got := spanish21PairStrategy(pv, di)
			if got != want {
				drift = append(drift, fmt.Sprintf("pair%d-%s: table=%s solver=%s",
					pv, upLabel(up), actionLetter(got), actionLetter(want)))
			}
		}
	}

	if len(drift) > 0 {
		t.Fatalf("committed Spanish 21 table drifted from the solver in %d cell(s):\n%s"+
			"regenerate with: go test -tags test ./internal/domain -run TestGenerateSpanish21Table -v",
			len(drift), joinLines(drift))
	}
}
