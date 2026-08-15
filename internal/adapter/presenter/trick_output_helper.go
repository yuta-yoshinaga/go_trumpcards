package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"

// cardHint builds the web-output hint from a domain hint's parts.
//
// Consolidates the block repeated verbatim in 93 presenter methods across the
// trick-taking games. Each game's GetHint returns its own type
// (domain.AluetteHint, domain.FortyFivesHint, …) with no shared interface, so
// the caller still does the nil check and unpacks the fields; only the
// construction is shared. See issue #5368.
//
// A plain function rather than a method on the output types: giving all 24
// *WebOutput structs a SetCardHint method would add more API surface than the
// duplication costs, and this keeps the assignment visible at the call site.
//
// Passive hints are filled in on Output, not only on HintOutput: HintOutput is
// the `command: "hint"` response and is never merged into the page state, so
// without it the frontend's `state.hint` stays undefined and every branch that
// reads it is dead (#4483).
func cardHint(cardIndices []int, reason string) *controller.WebOutputCardHint {
	return &controller.WebOutputCardHint{CardIndices: cardIndices, Reason: reason}
}
