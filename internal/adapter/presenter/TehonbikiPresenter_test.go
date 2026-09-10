//go:build test

package presenter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTehonbikiPresenterGame(t *testing.T) *domain.Tehonbiki {
	t.Helper()
	g := domain.NewDefaultTehonbiki()
	require.NoError(t, g.SetParentCard(4))
	return g
}

func TestTehonbikiWebPresenterHidesParentUntilResult(t *testing.T) {
	p := new(TehonbikiWebPresenter)
	var before map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(newTehonbikiPresenterGame(t), nil)), &before))
	if _, ok := before["parentCard"]; ok {
		t.Fatal("parent card leaked before the bet")
	}

	g := newTehonbikiPresenterGame(t)
	require.NoError(t, g.PlaceBet([]int{4}, domain.TehonbikiBetSingle, 50))
	var after struct {
		ParentCard *int `json:"parentCard"`
		PayoutNum  int  `json:"payoutNum"`
		PayoutDen  int  `json:"payoutDen"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &after))
	require.NotNil(t, after.ParentCard)
	if *after.ParentCard != 4 {
		t.Fatalf("parentCard=%d", *after.ParentCard)
	}
	if after.PayoutNum != 9 || after.PayoutDen != 2 {
		t.Fatalf("payout ratio=%d/%d", after.PayoutNum, after.PayoutDen)
	}
}

func TestTehonbikiCuiPresenterShowsResultAndErrors(t *testing.T) {
	p := new(TehonbikiCuiPresenter)
	g := newTehonbikiPresenterGame(t)
	require.NoError(t, g.PlaceBet([]int{1, 2}, domain.TehonbikiBetDouble, 50))
	if got := p.Output(g, nil); got == "" || !containsAll(got, "parent:", "payout:") {
		t.Fatalf("output=%q", got)
	}
	if got := p.Output(g, errPresenterTest); !containsAll(got, "boom") {
		t.Fatalf("error output=%q", got)
	}
}

func TestTehonbikiCuiPresenterShowsResultAtGameEnd(t *testing.T) {
	p := new(TehonbikiCuiPresenter)
	g := newTehonbikiPresenterGame(t)
	g.SetChips(10)
	require.NoError(t, g.SetParentCard(6))
	require.NoError(t, g.PlaceBet([]int{1}, domain.TehonbikiBetSingle, 10))
	require.NoError(t, g.NextRound())
	got := p.Output(g, nil)
	if !containsAll(got, "parent: 6", "result:", "payout:") {
		t.Fatalf("game-end output=%q", got)
	}
}

var errPresenterTest = presenterTestError("boom")

type presenterTestError string

func (e presenterTestError) Error() string { return string(e) }
func containsAll(s string, xs ...string) bool {
	for _, x := range xs {
		if !strings.Contains(s, x) {
			return false
		}
	}
	return true
}
