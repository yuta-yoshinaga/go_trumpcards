//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TehonbikiCuiPresenter renders the number wager state.
type TehonbikiCuiPresenter struct{}

func (*TehonbikiCuiPresenter) Output(c interfaces.TehonbikiGame, e error) string {
	return buildCuiOutput(i18n.T("tehonbiki.outputTitle"), func(b *strings.Builder) {
		b.WriteString("phase: " + strconv.Itoa(int(c.GetPhase())) + " chips: " + strconv.Itoa(c.GetChips()) + "\n")
		if c.GetPhase() == domain.TehonbikiPhaseResult {
			b.WriteString("parent: " + strconv.Itoa(c.GetParentCard()) + " result: " + strconv.Itoa(int(c.GetResult())) + " payout: " + strconv.Itoa(c.GetPayout()) + "\n")
		}
		cuiErrorBlock(b, e)
	})
}
func (*TehonbikiCuiPresenter) ActionLogOutput(c interfaces.TehonbikiGame) string {
	return actionLogOutputText(c)
}
func (*TehonbikiCuiPresenter) HintOutput(c interfaces.TehonbikiGame) string {
	if h := c.GetHint(); h != nil {
		return h.Reason
	}
	return ""
}
