//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MightyWebInput マイティWebインプット
type MightyWebInput struct {
	BaseWebInput
	Bid            *int             `json:"bid,omitempty"`
	NoTrump        *bool            `json:"noTrump,omitempty"`
	CardIndex      *int             `json:"cardIndex,omitempty"`
	TrumpSuit      *int             `json:"trumpSuit,omitempty"`
	PartnerSuit    *int             `json:"partnerSuit,omitempty"`
	PartnerValue   *int             `json:"partnerValue,omitempty"`
	DiscardIndices []int            `json:"discardIndices,omitempty"`
	JokerLeadSuit  *int             `json:"jokerLeadSuit,omitempty"`
	Config         *MightyWebConfig `json:"config,omitempty"`
}

// MightyWebConfig マイティWeb設定
type MightyWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	MinBid        *int `json:"minBid,omitempty"`
	NoTrumpExtra  *int `json:"noTrumpExtra,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// MightyWebOutputPlayer マイティWebアウトプットプレイヤー
type MightyWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	BidNoTrump      bool             `json:"bidNoTrump"`
	IsDeclarer      bool             `json:"isDeclarer"`
	IsPartner       bool             `json:"isPartner"`
	PartnerRevealed bool             `json:"partnerRevealed"`
	PointCards      int              `json:"pointCards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// MightyWebOutputTrickCard トリック中の1枚
type MightyWebOutputTrickCard struct {
	PlayerIdx      int            `json:"playerIdx"`
	Card           *WebOutputCard `json:"card"`
	IsJokerLead    bool           `json:"isJokerLead,omitempty"`
	LeadDemandSuit int            `json:"leadDemandSuit,omitempty"`
}

// MightyWebOutputHint ヒント出力
type MightyWebOutputHint struct {
	CardIndex      *int   `json:"cardIndex,omitempty"`
	Bid            *int   `json:"bid,omitempty"`
	BidNoTrump     *bool  `json:"bidNoTrump,omitempty"`
	TrumpSuit      *int   `json:"trumpSuit,omitempty"`
	PartnerSuit    *int   `json:"partnerSuit,omitempty"`
	PartnerValue   *int   `json:"partnerValue,omitempty"`
	DiscardIndices []int  `json:"discardIndices,omitempty"`
	JokerLeadSuit  *int   `json:"jokerLeadSuit,omitempty"`
	Reason         string `json:"reason"`
}

// MightyWebOutput マイティWebアウトプット
type MightyWebOutput struct {
	Players           []*MightyWebOutputPlayer    `json:"players"`
	Phase             int                         `json:"phase"`
	RoundNumber       int                         `json:"roundNumber"`
	TrickNumber       int                         `json:"trickNumber"`
	CurrentPlayerIdx  int                         `json:"currentPlayerIdx"`
	BidPlayerIdx      int                         `json:"bidPlayerIdx"`
	CurrentTrick      []*MightyWebOutputTrickCard `json:"currentTrick"`
	TrumpSuit         int                         `json:"trumpSuit"`
	PartnerCard       *WebOutputCard              `json:"partnerCard,omitempty"`
	DeclarerIdx       int                         `json:"declarerIdx"`
	PartnerIdx        int                         `json:"partnerIdx"`
	PartnerRevealed   bool                        `json:"partnerRevealed"`
	HighestBid        int                         `json:"highestBid"`
	HighestBidder     int                         `json:"highestBidder"`
	WinningBidNoTrump bool                        `json:"winningBidNoTrump"`
	Kitty             []*WebOutputCard            `json:"kitty,omitempty"`
	GameEndFlag       bool                        `json:"gameEndFlag"`
	WinnerTeam        int                         `json:"winnerTeam"`
	LeadPlayerIdx     int                         `json:"leadPlayerIdx"`
	Hint              *MightyWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config MightyWebOutputConfig `json:"config"`
}

// MightyWebOutputConfig マイティ設定アウトプット
type MightyWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	MinBid        int `json:"minBid"`
	NoTrumpExtra  int `json:"noTrumpExtra"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a MightyConfig from the nested web config, applying bounds checking.
func (c *MightyWebConfig) ToConfig() domain.MightyConfig {
	cfg := domain.DefaultMightyConfig()
	cfg.CpuDifficulty = domain.MightyCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MightyCpuDifficultyEasy), int(domain.MightyCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.MinBid, c.MinBid, 1, domain.MightyMaxPoints)
	webutil.ApplyBoundedInt(&cfg.NoTrumpExtra, c.NoTrumpExtra, 0, domain.MightyMaxPoints-1)
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a MightyConfig from the web input.
func (p MightyWebInput) ToConfig() domain.MightyConfig {
	return configOrDefault(p.Config, (*MightyWebConfig).ToConfig, domain.DefaultMightyConfig())
}

// MightyWebController マイティWebコントローラークラス
type MightyWebController = GameWebController[usecase.MightyInteractorIF, MightyWebInput, *MightyWebOutput]

// NewMightyWebController and NewMightyWebControllerWithProvider are
// the standard and provider-backed constructors for MightyWebController.
var NewMightyWebController, NewMightyWebControllerWithProvider = webControllerPair[usecase.MightyInteractorIF, MightyWebInput, *MightyWebOutput](
	newMightyDefaultOutput, mightyDispatch,
)

func newMightyDefaultOutput(msg string) *MightyWebOutput {
	return &MightyWebOutput{
		Players:       make([]*MightyWebOutputPlayer, 0),
		CurrentTrick:  make([]*MightyWebOutputTrickCard, 0),
		WinnerTeam:    domain.MightyWinnerUndecided,
		DeclarerIdx:   -1,
		PartnerIdx:    -1,
		HighestBidder: -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func mightyDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MightyInteractorIF, param MightyWebInput, newDefault func(string) *MightyWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, mi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		noTrump := param.NoTrump != nil && *param.NoTrump
		bc.writePresenterResponse(w, mi.Bid(*param.Bid, noTrump))
	case "t", "trump":
		if !requireParam(bc, w, newDefault, param.TrumpSuit == nil || param.PartnerSuit == nil || param.PartnerValue == nil, "param error: trumpSuit, partnerSuit, partnerValue are required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.DeclareTrumpAndFriend(*param.TrumpSuit, *param.PartnerSuit, *param.PartnerValue))
	case "e", "exchange":
		if !requireParam(bc, w, newDefault, len(param.DiscardIndices) == 0, "param error: discardIndices are required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.ExchangeKitty(param.DiscardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Play(*param.CardIndex))
	case "jl", "jokerlead":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil || param.JokerLeadSuit == nil, "param error: cardIndex and jokerLeadSuit are required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.PlayJokerLead(*param.CardIndex, *param.JokerLeadSuit))
	case "n", "next":
		bc.writePresenterResponse(w, mi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, mi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, mi.Hint, mi.ActionLog)
	}
	return true
}
