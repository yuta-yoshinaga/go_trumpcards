//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustMightyOutputJSON(msg string) string {
	out := &controller.MightyWebOutput{
		Players:       []*controller.MightyWebOutputPlayer{},
		CurrentTrick:  []*controller.MightyWebOutputTrickCard{},
		WinnerTeam:    domain.MightyWinnerUndecided,
		DeclarerIdx:   -1,
		PartnerIdx:    -1,
		HighestBidder: -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMightyOutputJSON: %v", err))
	}
	return string(b)
}

// newMightyWebMock returns an interactor mock with the standard happy-path bindings.
func newMightyWebMock(mockOutput string) *usecase.MockMightyInteractor {
	m := new(usecase.MockMightyInteractor)
	m.On("ResetWithConfig", domain.DefaultMightyConfig()).Return(mockOutput)
	m.On("Bid", 14, false).Return(mockOutput)
	m.On("Bid", 15, true).Return(mockOutput)
	m.On("DeclareTrumpAndFriend", 1, 2, 13).Return(mockOutput)
	m.On("ExchangeKitty", []int{0, 1, 2}).Return(mockOutput)
	m.On("Play", 3).Return(mockOutput)
	m.On("PlayJokerLead", 5, 2).Return(mockOutput)
	m.On("NextTrick").Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)
	return m
}

func TestMightyWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"bidPlayerIdx":0,"currentTrick":[],"trumpSuit":0,"declarerIdx":-1,"partnerIdx":-1,"partnerRevealed":false,"highestBid":0,"highestBidder":-1,"winningBidNoTrump":false,"gameEndFlag":false,"winnerTeam":-1,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"minBid":0,"noTrumpExtra":0,"pointLimit":0}}`

	miMock := newMightyWebMock(mockOutput)
	factory := func() uc.MightyInteractorIF { return miMock }
	ctrl := controller.NewMightyWebController(factory)
	defer ctrl.Stop()

	t.Run("Exec q returns bye message", func(t *testing.T) {
		var input controller.MightyWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"sess-q"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mustMightyOutputJSON("bye."))
	})

	t.Run("Exec reset uses default config", func(t *testing.T) {
		var input controller.MightyWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"sess-r"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("Exec bid plain", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "sess-bid"},
			Bid:          intPtr(14),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("Exec bid no-trump (noTrump=true)", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "sess-bid-nt"},
			Bid:          intPtr(15),
			NoTrump:      boolPtr(true),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		miMock.AssertCalled(t, "Bid", 15, true)
	})

	t.Run("Exec trump declares partner triplet", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "trump", SessionID: "sess-t"},
			TrumpSuit:    intPtr(1),
			PartnerSuit:  intPtr(2),
			PartnerValue: intPtr(13),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		miMock.AssertCalled(t, "DeclareTrumpAndFriend", 1, 2, 13)
	})

	t.Run("Exec exchange with discardIndices", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "exchange", SessionID: "sess-e"},
			DiscardIndices: []int{0, 1, 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		miMock.AssertCalled(t, "ExchangeKitty", []int{0, 1, 2})
	})

	t.Run("Exec play", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "sess-p"},
			CardIndex:    intPtr(3),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("Exec jokerlead", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput:  controller.BaseWebInput{Command: "jl", SessionID: "sess-jl"},
			CardIndex:     intPtr(5),
			JokerLeadSuit: intPtr(2),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		miMock.AssertCalled(t, "PlayJokerLead", 5, 2)
	})

	t.Run("Exec next / nextround / log / hint", func(t *testing.T) {
		for _, cmd := range []string{"n", "next", "nr", "nextround", "log", "l", "h", "hint"} {
			var input controller.MightyWebInput
			body := fmt.Sprintf(`{"command":%q,"sessionId":"sess-x"}`, cmd)
			_ = json.Unmarshal([]byte(body), &input)
			rec := execRequest(t, ctrl.Exec, &input)
			rec.CodeIs(http.StatusOK)
			rec.BodyIs(mockOutput)
		}
	})

	// Error cases
	t.Run("Exec bid missing field", func(t *testing.T) {
		var input controller.MightyWebInput
		_ = json.Unmarshal([]byte(`{"command":"bid","sessionId":"err-bid"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		rec.BodyIs(mustMightyOutputJSON("param error: bid is required."))
	})

	t.Run("Exec trump missing partner triplet", func(t *testing.T) {
		var input controller.MightyWebInput
		_ = json.Unmarshal([]byte(`{"command":"trump","sessionId":"err-t"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		rec.BodyIs(mustMightyOutputJSON("param error: trumpSuit, partnerSuit, partnerValue are required."))
	})

	t.Run("Exec trump missing partnerValue only", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "trump", SessionID: "err-tv"},
			TrumpSuit:    intPtr(1),
			PartnerSuit:  intPtr(2),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	t.Run("Exec exchange missing discardIndices", func(t *testing.T) {
		var input controller.MightyWebInput
		_ = json.Unmarshal([]byte(`{"command":"exchange","sessionId":"err-e"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		rec.BodyIs(mustMightyOutputJSON("param error: discardIndices are required."))
	})

	t.Run("Exec play missing cardIndex", func(t *testing.T) {
		var input controller.MightyWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"err-p"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		rec.BodyIs(mustMightyOutputJSON("param error: cardIndex is required."))
	})

	t.Run("Exec jokerlead missing cardIndex", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput:  controller.BaseWebInput{Command: "jl", SessionID: "err-jl1"},
			JokerLeadSuit: intPtr(2),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
		rec.BodyIs(mustMightyOutputJSON("param error: cardIndex and jokerLeadSuit are required."))
	})

	t.Run("Exec jokerlead missing leadSuit", func(t *testing.T) {
		input := controller.MightyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "jl", SessionID: "err-jl2"},
			CardIndex:    intPtr(0),
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	t.Run("Exec unknown command", func(t *testing.T) {
		var input controller.MightyWebInput
		_ = json.Unmarshal([]byte(`{"command":"weird","sessionId":"err-u"}`), &input)
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestMightyWebController_ToConfigBounds(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	cases := []struct {
		name string
		cfg  *controller.MightyWebConfig
		want domain.MightyConfig
	}{
		{
			name: "all valid",
			cfg:  &controller.MightyWebConfig{CpuDifficulty: intPtr(2), MinBid: intPtr(15), NoTrumpExtra: intPtr(3), PointLimit: intPtr(200)},
			want: domain.MightyConfig{CpuDifficulty: domain.MightyCpuDifficultyHard, MinBid: 15, NoTrumpExtra: 3, PointLimit: 200},
		},
		{
			name: "cpuDifficulty above max clamps to default",
			cfg:  &controller.MightyWebConfig{CpuDifficulty: intPtr(99)},
			want: domain.DefaultMightyConfig(),
		},
		{
			name: "cpuDifficulty below min clamps to default",
			cfg:  &controller.MightyWebConfig{CpuDifficulty: intPtr(-1)},
			want: domain.DefaultMightyConfig(),
		},
		{
			name: "minBid below 1 ignored",
			cfg:  &controller.MightyWebConfig{MinBid: intPtr(0)},
			want: domain.DefaultMightyConfig(),
		},
		{
			name: "minBid above max ignored",
			cfg:  &controller.MightyWebConfig{MinBid: intPtr(99)},
			want: domain.DefaultMightyConfig(),
		},
		{
			name: "noTrumpExtra below 0 ignored",
			cfg:  &controller.MightyWebConfig{NoTrumpExtra: intPtr(-1)},
			want: domain.DefaultMightyConfig(),
		},
		{
			name: "noTrumpExtra above max ignored",
			cfg:  &controller.MightyWebConfig{NoTrumpExtra: intPtr(99)},
			want: domain.DefaultMightyConfig(),
		},
		{
			name: "pointLimit below 1 ignored",
			cfg:  &controller.MightyWebConfig{PointLimit: intPtr(0)},
			want: domain.DefaultMightyConfig(),
		},
		{
			name: "pointLimit above 1000 ignored",
			cfg:  &controller.MightyWebConfig{PointLimit: intPtr(1001)},
			want: domain.DefaultMightyConfig(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			miMock := new(usecase.MockMightyInteractor)
			miMock.On("ResetWithConfig", tc.want).Return(mockOutput)
			ctrl := controller.NewMightyWebController(func() uc.MightyInteractorIF { return miMock })
			defer ctrl.Stop()

			input := controller.MightyWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-" + tc.name},
				Config:       tc.cfg,
			}
			rec := execRequest(t, ctrl.Exec, &input)
			rec.CodeIs(http.StatusOK)
			miMock.AssertCalled(t, "ResetWithConfig", tc.want)
		})
	}

	t.Run("nil config uses defaults", func(t *testing.T) {
		miMock := new(usecase.MockMightyInteractor)
		miMock.On("ResetWithConfig", domain.DefaultMightyConfig()).Return(mockOutput)
		ctrl := controller.NewMightyWebController(func() uc.MightyInteractorIF { return miMock })
		defer ctrl.Stop()

		input := controller.MightyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-nil"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		miMock.AssertCalled(t, "ResetWithConfig", domain.DefaultMightyConfig())
	})
}
