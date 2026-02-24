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

	const want = `{"error":"internal server error"}`

	t.Run("BlackJackWebPresenter marshal error", func(t *testing.T) {
		p := NewBlackJackWebPresenter()
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		assert.Equal(t, want, p.Output(bj, nil))
	})

	t.Run("PokerWebPresenter marshal error", func(t *testing.T) {
		p := NewPokerWebPresenter()
		poker := domain.NewPoker(domain.NewTrumpCards(0), domain.NewPokerPlayer(), domain.NewPokerPlayer())
		assert.Equal(t, want, p.Output(poker, nil))
	})

	t.Run("OldMaidWebPresenter marshal error", func(t *testing.T) {
		p := NewOldMaidWebPresenter()
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		om := domain.NewOldMaid(domain.NewTrumpCards(1), players)
		assert.Equal(t, want, p.Output(om, nil))
	})

	t.Run("DaifugoWebPresenter marshal error", func(t *testing.T) {
		p := NewDaifugoWebPresenter()
		players := []*domain.DaifugoPlayer{
			domain.NewDaifugoPlayer(true),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
		}
		dg := domain.NewDaifugo(domain.NewTrumpCards(0), players, domain.DefaultDaifugoConfig())
		assert.Equal(t, want, p.Output(dg, nil))
	})

	t.Run("SevensWebPresenter marshal error", func(t *testing.T) {
		p := NewSevensWebPresenter()
		players := []*domain.SevensPlayer{
			domain.NewSevensPlayer(true),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
		}
		s := domain.NewSevens(domain.NewTrumpCards(0), players, domain.DefaultSevensConfig())
		assert.Equal(t, want, p.Output(s, nil))
	})
}
