//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSetteEMezzoPresenter セッテ・エ・メッツォ プレゼンターモック
type MockSetteEMezzoPresenter struct {
	MockGamePresenter[interfaces.SetteEMezzoGame]
}
