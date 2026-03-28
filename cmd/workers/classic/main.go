//go:build js && wasm

package main

import (
	"net/http"

	"github.com/syumai/workers/cloudflare"

	"github.com/syumai/workers"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	corsmw "github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/cors"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func main() {
	mux := http.NewServeMux()

	// Hearts
	htc := controller.NewHeartsWebController(func() usecase.HeartsInteractorIF {
		config := domain.DefaultHeartsConfig()
		players := []*domain.HeartsPlayer{
			domain.NewHeartsPlayer(true),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
		}
		hearts := domain.NewHearts(domain.NewTrumpCards(0), players, config)
		return usecase.NewHeartsInteractor(hearts, new(presenter.HeartsWebPresenter))
	})
	mux.HandleFunc("/hearts/exec", htc.Exec)

	// Spades
	spc := controller.NewSpadesWebController(func() usecase.SpadesInteractorIF {
		config := domain.DefaultSpadesConfig()
		players := []*domain.SpadesPlayer{
			domain.NewSpadesPlayer(true),
			domain.NewSpadesPlayer(false),
			domain.NewSpadesPlayer(false),
			domain.NewSpadesPlayer(false),
		}
		spades := domain.NewSpades(domain.NewTrumpCards(0), players, config)
		return usecase.NewSpadesInteractor(spades, new(presenter.SpadesWebPresenter))
	})
	mux.HandleFunc("/spades/exec", spc.Exec)

	// Euchre
	euc := controller.NewEuchreWebController(func() usecase.EuchreInteractorIF {
		config := domain.DefaultEuchreConfig()
		players := []*domain.EuchrePlayer{
			domain.NewEuchrePlayer(true, 0),
			domain.NewEuchrePlayer(false, 1),
			domain.NewEuchrePlayer(false, 0),
			domain.NewEuchrePlayer(false, 1),
		}
		euchre := domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, config)
		return usecase.NewEuchreInteractor(euchre, new(presenter.EuchreWebPresenter))
	})
	mux.HandleFunc("/euchre/exec", euc.Exec)

	// Napoleon
	npc := controller.NewNapoleonWebController(func() usecase.NapoleonInteractorIF {
		config := domain.DefaultNapoleonConfig()
		players := []*domain.NapoleonPlayer{
			domain.NewNapoleonPlayer(true),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
			domain.NewNapoleonPlayer(false),
		}
		napoleon := domain.NewNapoleon(domain.NewTrumpCards(1), players, config)
		return usecase.NewNapoleonInteractor(napoleon, new(presenter.NapoleonWebPresenter))
	})
	mux.HandleFunc("/napoleon/exec", npc.Exec)

	// Old Maid
	omc := controller.NewOldMaidWebController(func() usecase.OldMaidInteractorIF {
		players := []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}
		oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
		return usecase.NewOldMaidInteractor(oldMaid, new(presenter.OldMaidWebPresenter))
	})
	mux.HandleFunc("/oldmaid/exec", omc.Exec)

	// Doubt
	dwc := controller.NewDoubtWebController(func() usecase.DoubtInteractorIF {
		players := []*domain.DoubtPlayer{
			domain.NewDoubtPlayer(true),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
		}
		doubt := domain.NewDoubt(domain.NewTrumpCards(0), players)
		return usecase.NewDoubtInteractor(doubt, new(presenter.DoubtWebPresenter))
	})
	mux.HandleFunc("/doubt/exec", dwc.Exec)

	// Daifugo
	dgc := controller.NewDaifugoWebController(func() usecase.DaifugoInteractorIF {
		config := domain.DefaultDaifugoConfig()
		players := []*domain.DaifugoPlayer{
			domain.NewDaifugoPlayer(true),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
		}
		daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewDaifugoInteractor(daifugo, new(presenter.DaifugoWebPresenter))
	})
	mux.HandleFunc("/daifugo/exec", dgc.Exec)

	// Sevens
	sgc := controller.NewSevensWebController(func() usecase.SevensInteractorIF {
		config := domain.DefaultSevensConfig()
		players := []*domain.SevensPlayer{
			domain.NewSevensPlayer(true),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
		}
		sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
		return usecase.NewSevensInteractor(sevens, new(presenter.SevensWebPresenter))
	})
	mux.HandleFunc("/sevens/exec", sgc.Exec)

	// Crazy Eights
	cec := controller.NewCrazyEightsWebController(func() usecase.CrazyEightsInteractorIF {
		config := domain.DefaultCrazyEightsConfig()
		players := []*domain.CrazyEightsPlayer{
			domain.NewCrazyEightsPlayer(true),
			domain.NewCrazyEightsPlayer(false),
			domain.NewCrazyEightsPlayer(false),
			domain.NewCrazyEightsPlayer(false),
		}
		ce := domain.NewCrazyEights(domain.NewTrumpCards(0), players, config)
		return usecase.NewCrazyEightsInteractor(ce, new(presenter.CrazyEightsWebPresenter))
	})
	mux.HandleFunc("/crazyeights/exec", cec.Exec)

	var handler http.Handler = mux
	if origins := corsmw.ParseOrigins(cloudflare.Getenv("CORS_ALLOWED_ORIGINS")); origins != nil {
		handler = corsmw.Middleware(origins, mux)
	}
	workers.Serve(handler)
}
