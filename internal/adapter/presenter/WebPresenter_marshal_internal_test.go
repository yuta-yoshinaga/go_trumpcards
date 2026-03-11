package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestWebPresenters_MarshalError(t *testing.T) {
	orig := jsonMarshal
	defer func() { jsonMarshal = orig }()
	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("marshal error")
	}

	want := internalServerErrorJSON()

	testCases := []struct {
		name   string
		output string
	}{
		{
			name: "BlackJackWebPresenter marshal error",
			output: func() string {
				p := new(BlackJackWebPresenter)
				bj := domain.NewDefaultBlackJack()
				bj.Reset()
				return p.Output(bj, nil)
			}(),
		},
		{
			name: "PokerWebPresenter marshal error",
			output: func() string {
				p := new(PokerWebPresenter)
				players := []*domain.PokerPlayer{
					domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
					domain.NewPokerPlayer(false, domain.PokerStyleConservative),
				}
				poker := domain.NewPoker(domain.NewTrumpCards(0), players, domain.DefaultPokerConfig())
				return p.Output(poker, nil)
			}(),
		},
		{
			name: "PokerWebPresenter OutputWithOdds marshal error",
			output: func() string {
				p := new(PokerWebPresenter)
				players := []*domain.PokerPlayer{
					domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
					domain.NewPokerPlayer(false, domain.PokerStyleConservative),
				}
				poker := domain.NewPoker(domain.NewTrumpCards(0), players, domain.DefaultPokerConfig())
				odds := []domain.PokerDrawOdds{
					{HandRank: 0, HandName: "High Card", Probability: 1.0, Count: 1, Total: 1},
				}
				return p.OutputWithOdds(poker, nil, odds)
			}(),
		},
		{
			name: "OldMaidWebPresenter marshal error",
			output: func() string {
				p := new(OldMaidWebPresenter)
				players := []*domain.OldMaidPlayer{
					domain.NewOldMaidPlayer(true),
					domain.NewOldMaidPlayer(false),
					domain.NewOldMaidPlayer(false),
					domain.NewOldMaidPlayer(false),
				}
				om := domain.NewOldMaid(domain.NewTrumpCards(1), players)
				return p.Output(om, nil)
			}(),
		},
		{
			name: "DaifugoWebPresenter marshal error",
			output: func() string {
				p := new(DaifugoWebPresenter)
				players := []*domain.DaifugoPlayer{
					domain.NewDaifugoPlayer(true),
					domain.NewDaifugoPlayer(false),
					domain.NewDaifugoPlayer(false),
					domain.NewDaifugoPlayer(false),
				}
				dg := domain.NewDaifugo(domain.NewTrumpCards(0), players, domain.DefaultDaifugoConfig())
				return p.Output(dg, nil)
			}(),
		},
		{
			name: "SevensWebPresenter marshal error",
			output: func() string {
				p := new(SevensWebPresenter)
				players := []*domain.SevensPlayer{
					domain.NewSevensPlayer(true),
					domain.NewSevensPlayer(false),
					domain.NewSevensPlayer(false),
					domain.NewSevensPlayer(false),
				}
				s := domain.NewSevens(domain.NewTrumpCards(0), players, domain.DefaultSevensConfig())
				return p.Output(s, nil)
			}(),
		},
		{
			name: "DoubtWebPresenter marshal error",
			output: func() string {
				p := new(DoubtWebPresenter)
				players := []*domain.DoubtPlayer{
					domain.NewDoubtPlayer(true),
					domain.NewDoubtPlayer(false),
					domain.NewDoubtPlayer(false),
					domain.NewDoubtPlayer(false),
				}
				d := domain.NewDoubt(domain.NewTrumpCards(0), players)
				return p.Output(d, nil)
			}(),
		},
		{
			name: "HoldemWebPresenter marshal error",
			output: func() string {
				p := new(HoldemWebPresenter)
				players := []*domain.HoldemPlayer{
					domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
					domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
					domain.NewHoldemPlayer(false, domain.HoldemStyleTAP),
					domain.NewHoldemPlayer(false, domain.HoldemStyleGTO),
				}
				h := domain.NewHoldem(domain.NewTrumpCards(0), players, domain.DefaultHoldemConfig())
				return p.Output(h, nil)
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, want, tc.output)
		})
	}
}
