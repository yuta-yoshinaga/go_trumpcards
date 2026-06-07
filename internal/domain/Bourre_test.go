package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBourreReset(t *testing.T) {
	b := domain.NewDefaultBourre()
	b.Reset()

	if b.GetPhase() != domain.BourrePhaseDecide {
		t.Fatalf("phase = %d, want Decide", b.GetPhase())
	}
	if b.GetHandNumber() != 1 {
		t.Errorf("handNumber = %d, want 1", b.GetHandNumber())
	}
	// 全員アンティ済み: pot == 5 players * ante
	wantPot := domain.BourrePlayerCnt * domain.BourreAnte
	if b.GetPot() != wantPot {
		t.Errorf("pot = %d, want %d", b.GetPot(), wantPot)
	}
	total := b.GetPot() + b.GetCarryPot()
	for i := 0; i < b.GetPlayerCnt(); i++ {
		p := b.GetPlayer(i)
		if p.GetCardsSize() != domain.BourreHandSize {
			t.Errorf("player %d has %d cards, want %d", i, p.GetCardsSize(), domain.BourreHandSize)
		}
		total += p.GetChips()
	}
	if total != domain.BourrePlayerCnt*domain.BourreInitChips {
		t.Errorf("chip conservation broken: total = %d", total)
	}
	if b.GetTrumpCard() == nil {
		t.Errorf("trump card should be set")
	}
}

// driveBourre plays a full game to completion, asserting the chip-conservation
// invariant at every step. humanPlays controls the human's decide choice.
func driveBourre(t *testing.T, b *domain.Bourre, humanPlays bool) {
	t.Helper()
	const wantTotal = domain.BourrePlayerCnt * domain.BourreInitChips
	humanIdx := 0
	for iter := 0; iter < 300000 && !b.GetGameEndFlag(); iter++ {
		switch b.GetPhase() {
		case domain.BourrePhaseRoundEnd:
			b.NextHand()
		default:
			if b.IsHumanTurn() {
				switch b.GetPhase() {
				case domain.BourrePhaseDecide:
					if err := b.PlayerDecide(humanPlays); err != nil {
						t.Fatalf("PlayerDecide: %v", err)
					}
				case domain.BourrePhaseDraw:
					if err := b.PlayerDraw([]int{0}); err != nil {
						t.Fatalf("PlayerDraw: %v", err)
					}
				case domain.BourrePhasePlay:
					valid := b.GetValidPlayIndices(humanIdx)
					if len(valid) == 0 {
						t.Fatalf("no valid plays for human in play phase")
					}
					if err := b.PlayerPlay(valid[0]); err != nil {
						t.Fatalf("PlayerPlay: %v", err)
					}
				}
			} else {
				b.CpuPlay()
			}
		}

		total := b.GetPot() + b.GetCarryPot()
		for i := 0; i < b.GetPlayerCnt(); i++ {
			total += b.GetPlayer(i).GetChips()
		}
		if total != wantTotal {
			t.Fatalf("chip conservation broken at iter %d: total = %d (phase %d)", iter, total, b.GetPhase())
		}
	}
	if !b.GetGameEndFlag() {
		t.Fatalf("game did not finish")
	}
	if b.GetWinnerIdx() < 0 || b.GetWinnerIdx() >= b.GetPlayerCnt() {
		t.Errorf("invalid winner idx %d", b.GetWinnerIdx())
	}
}

func TestBourreFullGame(t *testing.T) {
	cases := []struct {
		name       string
		difficulty domain.BourreCpuDifficulty
		humanPlays bool
	}{
		{"easy-play", domain.BourreDifficultyEasy, true},
		{"normal-play", domain.BourreDifficultyNormal, true},
		{"hard-play", domain.BourreDifficultyHard, true},
		{"normal-fold", domain.BourreDifficultyNormal, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 複数回繰り返してランダムな配札の多くの分岐 (ブーレ・引き分け・無参加) を踏む
			for rep := 0; rep < 25; rep++ {
				b := domain.NewDefaultBourre()
				b.SetConfig(domain.BourreConfig{CpuDifficulty: tc.difficulty})
				b.Reset()
				driveBourre(t, b, tc.humanPlays)
			}
		})
	}
}

func TestBourreInvalidActions(t *testing.T) {
	b := domain.NewDefaultBourre()
	b.Reset()

	// プレイフェーズでない時の Play
	if err := b.PlayerPlay(0); err == nil {
		t.Errorf("PlayerPlay should fail outside play phase")
	}
	// ドローフェーズでない時の Draw
	if err := b.PlayerDraw([]int{0}); err == nil {
		t.Errorf("PlayerDraw should fail outside draw phase")
	}

	// CPU の決定手番では人間の Decide は拒否される (最初の手番は CPU)
	if b.IsHumanTurn() {
		// 人間が最初の手番のケースもあり得るが、デフォルトではディーラー0の左=CPU1
		t.Skip("human is first decider in this deal")
	}
	if err := b.PlayerDecide(true); err == nil {
		t.Errorf("PlayerDecide should fail when not human's turn")
	}
}

func TestBourreDrawValidation(t *testing.T) {
	b := domain.NewDefaultBourre()
	b.Reset()
	// 人間がドロー手番になるまで進める (決定手番では参加を選ぶ)
	reached := false
	for i := 0; i < 100000 && !b.GetGameEndFlag(); i++ {
		if b.IsHumanTurn() {
			if b.GetPhase() == domain.BourrePhaseDraw {
				reached = true
				break
			}
			if b.GetPhase() == domain.BourrePhaseDecide {
				if err := b.PlayerDecide(true); err != nil {
					t.Fatalf("PlayerDecide: %v", err)
				}
				continue
			}
			break // 人間のプレイ手番など (ドローに到達しないケース)
		}
		b.CpuPlay()
	}
	if !reached {
		t.Skip("could not reach human draw turn deterministically")
	}
	if err := b.PlayerDraw([]int{0, 0}); err == nil {
		t.Errorf("duplicate draw indices should be rejected")
	}
	if err := b.PlayerDraw([]int{99}); err == nil {
		t.Errorf("out-of-range draw index should be rejected")
	}
}

func TestBourreJSONRoundTrip(t *testing.T) {
	b := domain.NewDefaultBourre()
	b.Reset()
	// 数手進める
	for i := 0; i < 10 && !b.GetGameEndFlag(); i++ {
		if b.IsHumanTurn() {
			if b.GetPhase() == domain.BourrePhaseDecide {
				_ = b.PlayerDecide(true)
			} else if b.GetPhase() == domain.BourrePhaseDraw {
				_ = b.PlayerDraw(nil)
			} else if b.GetPhase() == domain.BourrePhasePlay {
				v := b.GetValidPlayIndices(0)
				if len(v) > 0 {
					_ = b.PlayerPlay(v[0])
				}
			}
		} else {
			b.CpuPlay()
		}
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored domain.Bourre
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetPhase() != b.GetPhase() {
		t.Errorf("phase mismatch: %d vs %d", restored.GetPhase(), b.GetPhase())
	}
	if restored.GetPot() != b.GetPot() {
		t.Errorf("pot mismatch: %d vs %d", restored.GetPot(), b.GetPot())
	}
	if restored.GetPlayerCnt() != b.GetPlayerCnt() {
		t.Errorf("player count mismatch")
	}
}

func TestBourreConfigValidate(t *testing.T) {
	if err := domain.DefaultBourreConfig().Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
	if err := (domain.BourreConfig{CpuDifficulty: 99}).Validate(); err == nil {
		t.Errorf("out-of-range difficulty should be invalid")
	}
}

func TestBourrePlayerJSONRoundTrip(t *testing.T) {
	p := domain.NewBourrePlayer(true)
	p.SetChips(42)
	p.SetFolded(true)
	p.SetDrawn(true)
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored domain.BourrePlayer
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetChips() != 42 || !restored.GetFolded() || !restored.GetDrawn() {
		t.Errorf("player fields not restored: chips=%d folded=%v drawn=%v",
			restored.GetChips(), restored.GetFolded(), restored.GetDrawn())
	}
	if restored.GetTrickCount() != 1 {
		t.Errorf("trick count = %d, want 1", restored.GetTrickCount())
	}
}
