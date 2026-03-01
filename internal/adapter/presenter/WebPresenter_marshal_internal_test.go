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
				p := NewBlackJackWebPresenter()
				bj := domain.NewDefaultBlackJack()
				bj.Reset()
				return p.Output(bj, nil)
			}(),
		},
		{
			name: "PokerWebPresenter marshal error",
			output: func() string {
				p := NewPokerWebPresenter()
				poker := domain.NewPoker(domain.NewTrumpCards(0), domain.NewPokerPlayer(), domain.NewPokerPlayer())
				return p.Output(poker, nil)
			}(),
		},
		{
			name: "OldMaidWebPresenter marshal error",
			output: func() string {
				p := NewOldMaidWebPresenter()
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
				p := NewDaifugoWebPresenter()
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
				p := NewSevensWebPresenter()
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
				p := NewDoubtWebPresenter()
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
				p := NewHoldemWebPresenter()
				players := []*domain.HoldemPlayer{
					domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
					domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
					domain.NewHoldemPlayer(false, domain.HoldemStyleTAP),
					domain.NewHoldemPlayer(false, domain.HoldemStyleLAG),
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
