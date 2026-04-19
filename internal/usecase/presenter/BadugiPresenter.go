package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BadugiPresenter is the Badugi-specific presenter alias over the generic
// GamePresenter. Adapter-side implementations satisfy this via Output +
// ActionLogOutput.
type BadugiPresenter = GamePresenter[interfaces.BadugiGame]
