//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrenteEtQuaranteWebConfig はトラント・エ・カラント (Trente et Quarante) の Web 設定。
type TrenteEtQuaranteWebConfig struct {
	DefaultBet *int `json:"defaultBet,omitempty"`
}

// ToConfig は TrenteEtQuaranteWebConfig を domain.TrenteEtQuaranteConfig に変換する
// (境界チェック付き)。
func (c *TrenteEtQuaranteWebConfig) ToConfig() domain.TrenteEtQuaranteConfig {
	cfg := domain.DefaultTrenteEtQuaranteConfig()
	cfg.DefaultBet = domain.TrenteEtQuaranteBet(webutil.BoundedIntPtr(
		c.DefaultBet, int(domain.TrenteEtQuaranteBetNoir), int(domain.TrenteEtQuaranteBetInverse), int(cfg.DefaultBet)))
	return cfg
}

// TrenteEtQuaranteWebInput はトラント・エ・カラント Web インプット。
type TrenteEtQuaranteWebInput struct {
	BaseWebInput
	Bet    *int                       `json:"bet,omitempty"`
	Stake  *int                       `json:"stake,omitempty"`
	Config *TrenteEtQuaranteWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.TrenteEtQuaranteConfig を構築する。
func (p TrenteEtQuaranteWebInput) ToConfig() domain.TrenteEtQuaranteConfig {
	return configOrDefault(p.Config, (*TrenteEtQuaranteWebConfig).ToConfig, domain.DefaultTrenteEtQuaranteConfig())
}

// TrenteEtQuaranteWebOutputCard は 1 枚のカード出力。
type TrenteEtQuaranteWebOutput struct {
	Phase         int                            `json:"phase"`
	RoundNumber   int                            `json:"roundNumber"`
	Chips         int                            `json:"chips"`
	CurrentBet    int                            `json:"currentBet"`
	Stake         int                            `json:"stake"`
	NoirRow       []*WebOutputCard               `json:"noirRow"`
	RougeRow      []*WebOutputCard               `json:"rougeRow"`
	NoirTotal     int                            `json:"noirTotal"`
	RougeTotal    int                            `json:"rougeTotal"`
	WinningRow    int                            `json:"winningRow"`
	FirstCardRed  bool                           `json:"firstCardRed"`
	Refait        bool                           `json:"refait"`
	Result        int                            `json:"result"`
	Payout        int                            `json:"payout"`
	RemainingDeck int                            `json:"remainingDeck"`
	GameEndFlag   bool                           `json:"gameEndFlag"`
	Hint          *TrenteEtQuaranteWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config TrenteEtQuaranteWebConfigOutput `json:"config"`
}

// TrenteEtQuaranteWebConfigOutput は設定アウトプット。
type TrenteEtQuaranteWebConfigOutput struct {
	DefaultBet int `json:"defaultBet"`
}

// TrenteEtQuaranteWebOutputHint はヒント出力。
type TrenteEtQuaranteWebOutputHint struct {
	Bet    int    `json:"bet"`
	Reason string `json:"reason"`
}

// TrenteEtQuaranteWebController はトラント・エ・カラント Web コントローラークラス。
type TrenteEtQuaranteWebController = GameWebController[usecase.TrenteEtQuaranteInteractorIF, TrenteEtQuaranteWebInput, *TrenteEtQuaranteWebOutput]

// NewTrenteEtQuaranteWebController, NewTrenteEtQuaranteWebControllerWithProvider are the
// standard and provider-backed constructors for TrenteEtQuaranteWebController.
var NewTrenteEtQuaranteWebController, NewTrenteEtQuaranteWebControllerWithProvider = webControllerPair[usecase.TrenteEtQuaranteInteractorIF, TrenteEtQuaranteWebInput, *TrenteEtQuaranteWebOutput](
	newTrenteEtQuaranteDefaultOutput, trenteEtQuaranteDispatch,
)

func newTrenteEtQuaranteDefaultOutput(msg string) *TrenteEtQuaranteWebOutput {
	return &TrenteEtQuaranteWebOutput{
		NoirRow:       make([]*WebOutputCard, 0),
		RougeRow:      make([]*WebOutputCard, 0),
		WinningRow:    domain.TrenteEtQuaranteRowNone,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func trenteEtQuaranteDispatch(bc *baseController, w http.ResponseWriter, bi usecase.TrenteEtQuaranteInteractorIF, param TrenteEtQuaranteWebInput, newDefault func(string) *TrenteEtQuaranteWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "b", "bet":
		if !requireParam(bc, w, newDefault, param.Bet == nil, "param error: bet is required.") {
			return true
		}
		if !requireParam(bc, w, newDefault, param.Stake == nil, "param error: stake is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Bet(domain.TrenteEtQuaranteBet(*param.Bet), *param.Stake))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, bi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
