//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupNiuNiuCuiMockDefaults(g *interfaces.MockNiuNiuGame) {
	g.On("GetPhase").Return(domain.NiuNiuPhaseBet).Maybe()
	g.On("GetChips").Return(900).Maybe()
	g.On("GetMaxMultiplier").Return(domain.NiuNiuMaxMultiplier).Maybe()
	g.On("GetBankerIdx").Return(3).Maybe()
	g.On("GetLastResult").Return("親: 牛牛").Maybe()
	g.On("GetBankerRankKey").Return("niuniu").Maybe()
	g.On("GetGameEndFlag").Return(false).Maybe()
	g.On("GetMultiplier", mock.Anything).Return(3).Maybe()
	g.On("GetSeats").Return([]*domain.NiuNiuSeat{
		nnSeatFromJSON(`{"nm":"あなた","cp":false,"hd":{"cd":` + nnFiveCards + `,"bt":100,"ci":[0,1,2],"rk":10}}`),
		nnSeatFromJSON(`{"nm":"CPU1","cp":true,"hd":{"cd":` + nnFiveCards + `,"bt":20,"ci":[0,1,2],"rk":10}}`),
		nnSeatFromJSON(`{"nm":"CPU2","cp":true}`),
		nnSeatFromJSON(`{"nm":"親","cp":true}`),
	}).Maybe()
	g.On("GetBankerHand").Return(
		nnHandFromJSON(`{"cd":` + nnFiveCards + `,"ci":[0,1,2],"rk":10}`)).Maybe()
}

func TestNiuNiuCuiPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")

	t.Run("chips lead the view", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)

		out := new(NiuNiuCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "900")
		assert.Contains(t, out, i18n.T("niuniu.placeBet"))
	})

	t.Run("the banker's hand stays hidden until the round settles", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)

		out := new(NiuNiuCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("niuniu.faceDown"))
	})

	t.Run("a settled round reveals every hand", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetGameEndFlag").Return(true)
		g.On("GetPhase").Return(domain.NiuNiuPhaseEnd)

		out := new(NiuNiuCuiPresenter).Output(g, nil)
		assert.NotContains(t, out, i18n.T("niuniu.faceDown"))
		assert.Contains(t, out, "親: 牛牛")
	})

	// Marking the three cards that made the bull is what lets a player see WHY
	// the hand scored what it did -- five cards with no marks are unreadable.
	t.Run("the three cards forming the bull are marked", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)

		out := new(NiuNiuCuiPresenter).Output(g, nil)
		// 自分の手は常に見えるので、そこに * が 3 つ付く。
		if got := strings.Count(out, "*"); got < domain.NiuNiuComboSize {
			t.Errorf("found %d marks, want at least %d", got, domain.NiuNiuComboSize)
		}
	})

	t.Run("the multiplier is shown when it is above even money", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)

		assert.Contains(t, new(NiuNiuCuiPresenter).Output(g, nil), "x3")
	})

	t.Run("an even-money rank shows no multiplier", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetMultiplier")
		g.On("GetMultiplier", mock.Anything).Return(1)

		assert.NotContains(t, new(NiuNiuCuiPresenter).Output(g, nil), "x1")
	})

	t.Run("payouts appear once the round settles", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetSeats")
		g.On("GetGameEndFlag").Return(true)
		g.On("GetPhase").Return(domain.NiuNiuPhaseEnd)
		g.On("GetSeats").Return([]*domain.NiuNiuSeat{
			nnSeatFromJSON(`{"nm":"あなた","cp":false,"hd":{"cd":` + nnFiveCards + `,"bt":100,"ci":[0,1,2],"rk":10,"po":300}}`),
			nnSeatFromJSON(`{"nm":"CPU1","cp":true}`),
			nnSeatFromJSON(`{"nm":"CPU2","cp":true}`),
			nnSeatFromJSON(`{"nm":"親","cp":true}`),
		})

		assert.Contains(t, new(NiuNiuCuiPresenter).Output(g, nil), "300")
	})

	t.Run("error block", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)
		assert.Contains(t, new(NiuNiuCuiPresenter).Output(g, assertError{}), "boom")
	})
}

func TestNiuNiuCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("mid-round hides the log", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		g.On("GetGameEndFlag").Return(false)
		assert.NotContains(t, new(NiuNiuCuiPresenter).ActionLogOutput(g), "deal")
	})

	t.Run("a settled round shows the log", func(t *testing.T) {
		g := new(interfaces.MockNiuNiuGame)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "deal", Detail: "test detail"},
		})
		assert.Contains(t, new(NiuNiuCuiPresenter).ActionLogOutput(g), "test detail")
	})
}

// ラウンド終了の見出しがロケールに追随すること。
//
// 以前は GetLastResult() の "親: 牛牛" をそのまま出しており、英語ロケールでも
// 日本語のままだった (#5567)。
func TestNiuNiuCuiPresenter_BankerResultFollowsTheLocale(t *testing.T) {
	render := func(rankKey string) string {
		g := new(interfaces.MockNiuNiuGame)
		setupNiuNiuCuiMockDefaults(g)
		// ラウンド終了行を出すのは GetGameEndFlag() であって GetPhase() ではない。
		// フェーズだけ差し替えると賭けの画面が返り、Contains が素通りする。
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetGameEndFlag")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBankerRankKey")
		g.On("GetGameEndFlag").Return(true)
		g.On("GetBankerRankKey").Return(rankKey)
		return new(NiuNiuCuiPresenter).Output(g, nil)
	}

	t.Run("ja", func(t *testing.T) {
		assert.Contains(t, render("niuniu"), "親: 牛牛")
		assert.Contains(t, render("none"), "親: 無牛")
		assert.Contains(t, render("n7"), "親: 牛7")
	})

	t.Run("en", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		for _, c := range []struct{ key, want string }{
			{"niuniu", "Banker: Niu Niu"},
			{"none", "Banker: No Niu"},
			{"n7", "Banker: Niu 7"},
		} {
			out := render(c.key)
			assert.Contains(t, out, c.want)
			// 日本語が残っていたら、それはロケールを無視している経路。
			assert.NotContains(t, out, "親:")
			assert.NotContains(t, out, "牛牛")
		}
	})
}
