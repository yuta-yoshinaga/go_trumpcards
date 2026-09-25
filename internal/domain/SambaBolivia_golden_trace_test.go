//go:build test

package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const sambaBoliviaGoldenPath = "testdata/samba_bolivia_golden.json"

type sambaBoliviaTrace struct {
	Starts []sambaBoliviaGameTrace `json:"games"`
}

type sambaBoliviaGameTrace struct {
	Game  string             `json:"game"`
	Deals []sambaBoliviaDeal `json:"deals"`
}

type sambaBoliviaDeal struct {
	Start  string   `json:"start"`
	Hashes []string `json:"hashes"`
	Stop   string   `json:"stop"`
}

type sambaBoliviaGame interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	CpuPlay()
	GetGameEndFlag() bool
	GetRoundNumber() int
	GetPhase() int
}

type sambaTraceAdapter struct{ *Samba }

func (g sambaTraceAdapter) GetPhase() int { return int(g.Samba.GetPhase()) }

type boliviaTraceAdapter struct{ *Bolivia }

func (g boliviaTraceAdapter) GetPhase() int { return int(g.Bolivia.GetPhase()) }

func traceHumanSamba(g *Samba) error {
	switch g.GetPhase() {
	case SambaPhaseDraw:
		return g.PlayerDrawFromStock()
	case SambaPhaseMeld:
		return g.PlayerSkipMeld()
	case SambaPhaseDiscard:
		for i := 0; i < g.GetPlayer(g.GetCurrentPlayerIdx()).GetCardsSize(); i++ {
			if err := g.PlayerDiscard(i); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("no legal Samba action in phase %d", g.GetPhase())
}

func traceHumanBolivia(g *Bolivia) error {
	switch g.GetPhase() {
	case BoliviaPhaseDraw:
		return g.PlayerDrawFromStock()
	case BoliviaPhaseMeld:
		return g.PlayerSkipMeld()
	case BoliviaPhaseDiscard:
		for i := 0; i < g.GetPlayer(g.GetCurrentPlayerIdx()).GetCardsSize(); i++ {
			if err := g.PlayerDiscard(i); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("no legal Bolivia action in phase %d", g.GetPhase())
}

func advanceSambaBolivia(g sambaBoliviaGame, human func() error, initialRound int) (string, string, error) {
	if g.GetGameEndFlag() || g.GetRoundNumber() != initialRound {
		return "", "round_or_game_end", nil
	}
	// An empty stock would reshuffle through rand, which a replay cannot reproduce.
	if g.GetPhase() == int(SambaPhaseDraw) && g.(interface{ GetDrawPileCount() int }).GetDrawPileCount() == 0 {
		return "", "stock_exhausted", nil
	}
	if g.(interface{ IsHumanTurn() bool }).IsHumanTurn() {
		if err := human(); err != nil {
			return "", "", err
		}
	} else {
		g.CpuPlay()
	}
	state, err := g.MarshalJSON()
	if err != nil {
		return "", "", err
	}
	return hashState(state), "", nil
}

func hashState(state []byte) string { sum := sha256.Sum256(state); return hex.EncodeToString(sum[:]) }

func TestSambaBoliviaGoldenTrace(t *testing.T) {
	var golden sambaBoliviaTrace
	if os.Getenv("GOLDEN_UPDATE") == "1" {
		golden = generateSambaBoliviaGolden(t)
		data, err := json.MarshalIndent(golden, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(sambaBoliviaGoldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(sambaBoliviaGoldenPath, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	data, err := os.ReadFile(sambaBoliviaGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	for _, game := range golden.Starts {
		t.Run(game.Game, func(t *testing.T) {
			steps := 0
			for dealIndex, deal := range game.Deals {
				var gameState sambaBoliviaGame
				var human func() error
				if game.Game == "Samba" {
					instance := &Samba{}
					gameState = sambaTraceAdapter{instance}
					human = func() error { return traceHumanSamba(instance) }
				} else {
					instance := &Bolivia{}
					gameState = boliviaTraceAdapter{instance}
					human = func() error { return traceHumanBolivia(instance) }
				}
				if err := gameState.UnmarshalJSON([]byte(deal.Start)); err != nil {
					t.Fatalf("deal %d restore: %v", dealIndex, err)
				}
				startRound := gameState.GetRoundNumber()
				for i, want := range deal.Hashes {
					got, stop, err := advanceSambaBolivia(gameState, human, startRound)
					if err != nil {
						t.Fatalf("deal %d step %d: %v", dealIndex, i+1, err)
					}
					if stop != "" {
						t.Fatalf("deal %d stopped at step %d (%s), golden has %d steps", dealIndex, i+1, stop, len(deal.Hashes))
					}
					if got != want {
						t.Fatalf("deal %d step %d hash mismatch: got %s want %s", dealIndex, i+1, got, want)
					}
					steps++
				}
				if deal.Stop != "" {
					continue
				}
			}
			if steps == 0 {
				t.Fatal("golden trace contains no replayed steps")
			}
		})
	}
}

func generateSambaBoliviaGolden(t *testing.T) sambaBoliviaTrace {
	var result sambaBoliviaTrace
	for _, name := range []string{"Samba", "Bolivia"} {
		entry := sambaBoliviaGameTrace{Game: name}
		for deal := 0; deal < 12; deal++ {
			var instance sambaBoliviaGame
			var human func() error
			if name == "Samba" {
				g := NewDefaultSamba()
				cfg := g.GetConfig()
				cfg.CpuDifficulty = SambaCpuDifficultyHard
				g.SetConfig(cfg)
				g.Reset()
				instance = sambaTraceAdapter{g}
				human = func() error { return traceHumanSamba(g) }
			} else {
				g := NewDefaultBolivia()
				cfg := g.GetConfig()
				cfg.CpuDifficulty = BoliviaCpuDifficultyHard
				g.SetConfig(cfg)
				g.Reset()
				instance = boliviaTraceAdapter{g}
				human = func() error { return traceHumanBolivia(g) }
			}
			start, err := instance.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			record := sambaBoliviaDeal{Start: string(start)}
			initialRound := instance.GetRoundNumber()
			for len(record.Hashes) < 40 {
				if instance.GetGameEndFlag() || instance.GetRoundNumber() != initialRound {
					record.Stop = "round_or_game_end"
					break
				}
				if instance.GetPhase() == 0 && instance.(interface{ GetDrawPileCount() int }).GetDrawPileCount() == 0 {
					record.Stop = "stock_exhausted"
					break
				}
				hash, _, err := advanceSambaBolivia(instance, human, initialRound)
				if err != nil {
					t.Fatalf("%s deal %d step %d: %v", name, deal, len(record.Hashes)+1, err)
				}
				record.Hashes = append(record.Hashes, hash)
			}
			entry.Deals = append(entry.Deals, record)
		}
		result.Starts = append(result.Starts, entry)
	}
	return result
}
