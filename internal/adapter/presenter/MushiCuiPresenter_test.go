//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestMushiCuiPresenter_ShowsBothCapturesButOnlyTheHumanHand(t *testing.T) {
	m := domain.NewDefaultMushi()
	m.Reset()
	out := new(MushiCuiPresenter).Output(m, nil)

	// Two "captured" lines -- one per seat, because captures are public.
	assert.Equal(t, 2, strings.Count(out, "取り札:"), "both seats' captures are shown")
	// One indexed hand: the human's. The CPU's hand must not be printed.
	assert.Equal(t, 1, strings.Count(out, "[0]"+mushiCardStrForTest(m)), "")
}

// mushiCardStrForTest returns the rendering of the human's first hand card, so
// the assertion above counts a real card rather than a substring that might
// appear anywhere.
func mushiCardStrForTest(m *domain.Mushi) string {
	return mushiCardStr(m.GetPlayer(0).GetCard(0))
}

func TestMushiCuiPresenter_HidesTheCpuHand(t *testing.T) {
	m := domain.NewDefaultMushi()
	m.Reset()
	out := new(MushiCuiPresenter).Output(m, nil)

	// Every CPU hand card, if printed, would appear verbatim.
	cpu := m.GetPlayer(1)
	for i := range cpu.GetCardsSize() {
		assert.NotContains(t, out, "[0]"+mushiCardStr(cpu.GetCard(i)),
			"the CPU's hand must not be printed")
	}
}

func TestMushiCuiPresenter_MarksTheWildCard(t *testing.T) {
	assert.True(t, strings.HasSuffix(mushiCardStr(domain.NewCard(11, 4, true)), "*"))
	assert.False(t, strings.HasSuffix(mushiCardStr(domain.NewCard(11, 1, true)), "*"))
	assert.Equal(t, "--", mushiCardStr(nil))
	assert.Equal(t, "-", mushiCardListStr(nil, false))
}

func TestMushiCuiPresenter_RendersEachPhase(t *testing.T) {
	t.Run("an error", func(t *testing.T) {
		m := domain.NewDefaultMushi()
		m.Reset()
		assert.Contains(t, new(MushiCuiPresenter).Output(m, assert.AnError), assert.AnError.Error())
	})

	t.Run("round end and game end", func(t *testing.T) {
		m := domain.NewDefaultMushi()
		cfg := m.GetConfig()
		cfg.TargetRounds = 1
		m.SetConfig(cfg)
		m.Reset()
		for range 400 {
			if m.GetGameEndFlag() {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			action := m.MushiCpuDecide(idx)
			phase := m.GetPhase()
			if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
				require.NoError(t, m.SelectCapture(idx, action.FieldIdx))
				continue
			}
			if phase != domain.MushiPhasePlay {
				break
			}
			require.NoError(t, m.PlayCard(idx, action.HandIdx))
		}
		assert.NotEmpty(t, new(MushiCuiPresenter).Output(m, nil))
		assert.NotEmpty(t, new(MushiCuiPresenter).ActionLogOutput(m))
	})

	t.Run("a pending selection", func(t *testing.T) {
		for range 300 {
			m := domain.NewDefaultMushi()
			m.Reset()
			for range 40 {
				phase := m.GetPhase()
				if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
					assert.NotEmpty(t, new(MushiCuiPresenter).Output(m, nil))
					return
				}
				if phase != domain.MushiPhasePlay {
					break
				}
				idx := m.GetCurrentPlayerIdx()
				if m.GetPlayer(idx).GetCardsSize() == 0 || m.PlayCard(idx, 0) != nil {
					break
				}
			}
		}
		t.Skip("no selection phase occurred")
	})
}

func TestMushiCuiPresenter_HintResolvesItsReasonKey(t *testing.T) {
	// The Web presenter ships the identifier and the frontend looks it up; the
	// CUI must resolve it or it prints the raw key at the player.
	for range 50 {
		m := domain.NewDefaultMushi()
		m.Reset()
		out := new(MushiCuiPresenter).HintOutput(m)
		assert.NotContains(t, out, "mushi.hint.", "the reason must be translated, not printed raw")
		assert.NotEmpty(t, strings.TrimSpace(out))
	}
}

func TestMushiCuiPresenter_HintRendersEachShape(t *testing.T) {
	// The CUI hint has three shapes -- a hand index, a field index, and a bare
	// reason -- and each formats differently. A missing branch prints the wrong
	// line or the raw key.
	t.Run("a field-index suggestion", func(t *testing.T) {
		for range 300 {
			m := domain.NewDefaultMushi()
			m.Reset()
			for range 40 {
				phase := m.GetPhase()
				if (phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect) &&
					m.GetCurrentPlayerIdx() == 0 {
					out := new(MushiCuiPresenter).HintOutput(m)
					assert.Contains(t, out, "場札[")
					return
				}
				if phase != domain.MushiPhasePlay {
					break
				}
				idx := m.GetCurrentPlayerIdx()
				if m.GetPlayer(idx).GetCardsSize() == 0 || m.PlayCard(idx, 0) != nil {
					break
				}
			}
		}
		t.Skip("no human selection phase occurred")
	})

	t.Run("a bare reason once the game is over", func(t *testing.T) {
		m := domain.NewDefaultMushi()
		cfg := m.GetConfig()
		cfg.TargetRounds = 1
		m.SetConfig(cfg)
		m.Reset()
		for range 400 {
			if m.GetGameEndFlag() {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			action := m.MushiCpuDecide(idx)
			phase := m.GetPhase()
			if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
				require.NoError(t, m.SelectCapture(idx, action.FieldIdx))
				continue
			}
			if phase != domain.MushiPhasePlay {
				break
			}
			require.NoError(t, m.PlayCard(idx, action.HandIdx))
		}
		out := new(MushiCuiPresenter).HintOutput(m)
		// Assert on the index markers, NOT on a bare "[": ANSI colour escapes
		// contain one, so `NotContains(out, "[")` can never pass.
		assert.NotContains(t, out, "手札[")
		assert.NotContains(t, out, "場札[")
		assert.NotContains(t, out, "mushi.hint.")
	})
}
