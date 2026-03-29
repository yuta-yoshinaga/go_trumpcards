package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NapoleonWebInput ナポレオンWebインプット
type NapoleonWebInput struct {
	BaseWebInput
	Bid           *int               `json:"bid,omitempty"`
	CardIndex     *int               `json:"cardIndex,omitempty"`
	TrumpSuit     *int               `json:"trumpSuit,omitempty"`
	AdjutantSuit  *int               `json:"adjutantSuit,omitempty"`
	AdjutantValue *int               `json:"adjutantValue,omitempty"`
	DiscardIndex  *int               `json:"discardIndex,omitempty"`
	Config        *NapoleonWebConfig `json:"config,omitempty"`
}

// NapoleonWebConfig ナポレオンWeb設定
type NapoleonWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	MinBid        *int `json:"minBid,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// NapoleonWebOutputPlayer ナポレオンWebアウトプットプレイヤー
type NapoleonWebOutputPlayer struct {
	ID               int              `json:"id"`
	IsHuman          bool             `json:"isHuman"`
	CardCount        int              `json:"cardCount"`
	Cards            []*WebOutputCard `json:"cards"`
	Bid              int              `json:"bid"`
	IsNapoleon       bool             `json:"isNapoleon"`
	IsAdjutant       bool             `json:"isAdjutant"`
	AdjutantRevealed bool             `json:"adjutantRevealed"`
	PictureCards     int              `json:"pictureCards"`
	RoundScore       int              `json:"roundScore"`
	CumulativeScore  int              `json:"cumulativeScore"`
	TrickCount       int              `json:"trickCount"`
}

// NapoleonWebOutputTrickCard トリック中の1枚
type NapoleonWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// NapoleonWebOutputHint ヒント出力
type NapoleonWebOutputHint struct {
	CardIndex     *int   `json:"cardIndex,omitempty"`
	Bid           *int   `json:"bid,omitempty"`
	TrumpSuit     *int   `json:"trumpSuit,omitempty"`
	AdjutantSuit  *int   `json:"adjutantSuit,omitempty"`
	AdjutantValue *int   `json:"adjutantValue,omitempty"`
	DiscardIndex  *int   `json:"discardIndex,omitempty"`
	Reason        string `json:"reason"`
}

// NapoleonWebOutput ナポレオンWebアウトプット
type NapoleonWebOutput struct {
	Players          []*NapoleonWebOutputPlayer    `json:"players"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	TrickNumber      int                           `json:"trickNumber"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	BidPlayerIdx     int                           `json:"bidPlayerIdx"`
	CurrentTrick     []*NapoleonWebOutputTrickCard `json:"currentTrick"`
	TrumpSuit        int                           `json:"trumpSuit"`
	AdjutantCard     *WebOutputCard                `json:"adjutantCard,omitempty"`
	NapoleonIdx      int                           `json:"napoleonIdx"`
	AdjutantIdx      int                           `json:"adjutantIdx"`
	AdjutantRevealed bool                          `json:"adjutantRevealed"`
	HighestBid       int                           `json:"highestBid"`
	HighestBidder    int                           `json:"highestBidder"`
	Kitty            []*WebOutputCard              `json:"kitty,omitempty"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerTeam       int                           `json:"winnerTeam"`
	LeadPlayerIdx    int                           `json:"leadPlayerIdx"`
	Hint             *NapoleonWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config NapoleonWebOutputConfig `json:"config"`
}

// NapoleonWebOutputConfig ナポレオン設定アウトプット
type NapoleonWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	MinBid        int `json:"minBid"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a NapoleonConfig from the nested web config, applying bounds checking.
func (c *NapoleonWebConfig) ToConfig() domain.NapoleonConfig {
	cfg := domain.DefaultNapoleonConfig()
	cfg.CpuDifficulty = domain.NapoleonCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.NapoleonCpuDifficultyEasy), int(domain.NapoleonCpuDifficultyHard), int(cfg.CpuDifficulty)))
	cfg.MinBid = webutil.BoundedIntPtr(c.MinBid, 1, domain.NapoleonMaxPictureCards, cfg.MinBid)
	cfg.PointLimit = webutil.BoundedIntPtr(c.PointLimit, 1, 1000, cfg.PointLimit)
	return cfg
}

// ToConfig builds a NapoleonConfig from the web input.
func (p NapoleonWebInput) ToConfig() domain.NapoleonConfig {
	return configOrDefault(p.Config, (*NapoleonWebConfig).ToConfig, domain.DefaultNapoleonConfig())
}

// NapoleonWebController ナポレオンWebコントローラークラス
type NapoleonWebController = GameWebController[usecase.NapoleonInteractorIF, NapoleonWebInput, *NapoleonWebOutput]

// NewNapoleonWebController コンストラクタ
func NewNapoleonWebController(factory func() usecase.NapoleonInteractorIF) *NapoleonWebController {
	return NewGameWebController(factory, newNapoleonDefaultOutput, napoleonDispatch)
}

// NewNapoleonWebControllerWithProvider creates a NapoleonWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewNapoleonWebControllerWithProvider(
	provider SessionProvider[usecase.NapoleonInteractorIF],
	factory func() usecase.NapoleonInteractorIF,
) *NapoleonWebController {
	return NewGameWebControllerWithProvider(provider, factory, newNapoleonDefaultOutput, napoleonDispatch)
}

func newNapoleonDefaultOutput(msg string) *NapoleonWebOutput {
	return &NapoleonWebOutput{
		Players:       make([]*NapoleonWebOutputPlayer, 0),
		CurrentTrick:  make([]*NapoleonWebOutputTrickCard, 0),
		WinnerTeam:    domain.NapoleonWinnerUndecided,
		NapoleonIdx:   -1,
		AdjutantIdx:   -1,
		HighestBidder: -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func napoleonDispatch(bc *baseController, w http.ResponseWriter, ni usecase.NapoleonInteractorIF, param NapoleonWebInput, newDefault func(string) *NapoleonWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ni.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if param.Bid == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: bid is required."))
			return true
		}
		bc.writePresenterResponse(w, ni.Bid(*param.Bid))
	case "t", "trump":
		if param.TrumpSuit == nil || param.AdjutantSuit == nil || param.AdjutantValue == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: trumpSuit, adjutantSuit, adjutantValue are required."))
			return true
		}
		bc.writePresenterResponse(w, ni.DeclareTrump(*param.TrumpSuit, *param.AdjutantSuit, *param.AdjutantValue))
	case "e", "exchange":
		if param.DiscardIndex == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: discardIndex is required."))
			return true
		}
		bc.writePresenterResponse(w, ni.ExchangeKitty(*param.DiscardIndex))
	case "p", "play":
		if param.CardIndex == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: cardIndex is required."))
			return true
		}
		bc.writePresenterResponse(w, ni.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ni.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ni.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ni.Hint, ni.ActionLog)
	}
	return true
}
