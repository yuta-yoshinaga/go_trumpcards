package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PineappleWebInput パイナ���プルポーカーWebインプット
type PineappleWebInput struct {
	BaseWebInput
	Amount           int             `json:"amount,omitempty"`
	HumanPlayMs      int             `json:"humanPlayMs,omitempty"`
	CardIdx          *int            `json:"cardIdx,omitempty"`
	SmallBlind       *int            `json:"smallBlind,omitempty"`
	BigBlind         *int            `json:"bigBlind,omitempty"`
	TournamentMode   *bool           `json:"tournamentMode,omitempty"`
	BlindLevelHands  *int            `json:"blindLevelHands,omitempty"`
	BlindMultiplier  *int            `json:"blindMultiplier,omitempty"`
	BettingLimit     *int            `json:"bettingLimit,omitempty"`
	TableSize        *int            `json:"tableSize,omitempty"`
	RebuyEnabled     *bool           `json:"rebuyEnabled,omitempty"`
	RebuyMaxCount    *int            `json:"rebuyMaxCount,omitempty"`
	RebuyChips       *int            `json:"rebuyChips,omitempty"`
	RebuyPeriodHands *int            `json:"rebuyPeriodHands,omitempty"`
	AddonEnabled     *bool           `json:"addonEnabled,omitempty"`
	AddonChips       *int            `json:"addonChips,omitempty"`
	AddonAfterHand   *int            `json:"addonAfterHand,omitempty"`
	CpuMetaAI        bool            `json:"cpuMetaAI,omitempty"`
	Profile          json.RawMessage `json:"profile,omitempty"`
}

// PineappleWebOutputPlayer パイナップルポーカーWebアウトプットプレイヤー
type PineappleWebOutputPlayer = HoldemWebOutputPlayer

// PineappleWebOutputCpuAction パイナップルポーカーCPU行動記録
type PineappleWebOutputCpuAction = HoldemWebOutputCpuAction

// PineappleWebOutputResult パイナップルポーカーショーダウン結果
type PineappleWebOutputResult = HoldemWebOutputResult

// PineappleWebOutputSidePot パイナップルポーカーサイドポット
type PineappleWebOutputSidePot = HoldemWebOutputSidePot

// PineappleWebOutputEquity パイナップルポーカーエクイティ情報
type PineappleWebOutputEquity = HoldemWebOutputEquity

// PineappleWebOutputMetaAI メタAI情報
type PineappleWebOutputMetaAI = HoldemWebOutputMetaAI

// PineappleWebOutput パイナップルポーカーWebアウトプット
type PineappleWebOutput struct {
	Players          []*PineappleWebOutputPlayer     `json:"players"`
	CommunityCards   []*WebOutputCard                `json:"communityCards"`
	Pot              int                             `json:"pot"`
	SidePots         []*PineappleWebOutputSidePot    `json:"sidePots"`
	DealerIdx        int                             `json:"dealerIdx"`
	CurrentTurn      int                             `json:"currentTurn"`
	Phase            int                             `json:"phase"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	LastBet          int                             `json:"lastBet"`
	MinRaise         int                             `json:"minRaise"`
	BettingLimit     int                             `json:"bettingLimit"`
	RaiseCount       int                             `json:"raiseCount"`
	MaxBetAmount     int                             `json:"maxBetAmount"`
	RoundResults     []*PineappleWebOutputResult     `json:"roundResults"`
	CpuActions       []*PineappleWebOutputCpuAction  `json:"cpuActions"`
	HandCount        int                             `json:"handCount"`
	SmallBlind       int                             `json:"smallBlind"`
	BigBlind         int                             `json:"bigBlind"`
	TournamentMode   bool                            `json:"tournamentMode"`
	BlindLevelHands  int                             `json:"blindLevelHands"`
	BlindMultiplier  int                             `json:"blindMultiplier"`
	TableSize        int                             `json:"tableSize"`
	RebuyAvailable   bool                            `json:"rebuyAvailable"`
	AddonAvailable   bool                            `json:"addonAvailable"`
	RebuyCounts      []int                           `json:"rebuyCounts"`
	AddonUsed        []bool                          `json:"addonUsed"`
	RebuyEnabled     bool                            `json:"rebuyEnabled"`
	AddonEnabled     bool                            `json:"addonEnabled"`
	RebuyMaxCount    int                             `json:"rebuyMaxCount"`
	RebuyChips       int                             `json:"rebuyChips"`
	AddonChips       int                             `json:"addonChips"`
	RebuyPeriodHands int                             `json:"rebuyPeriodHands"`
	AddonAfterHand   int                             `json:"addonAfterHand"`
	RebuyPhaseType   int                             `json:"rebuyPhaseType"`
	MuckAvailable    bool                            `json:"muckAvailable"`
	IsDiscardPhase   bool                            `json:"isDiscardPhase"`
	DiscardDone      []bool                          `json:"discardDone"`
	Equity           *PineappleWebOutputEquity       `json:"equity,omitempty"`
	PotOdds          *float64                        `json:"potOdds,omitempty"`
	MetaAI           *PineappleWebOutputMetaAI       `json:"metaAI,omitempty"`
	Profile          *domain.BettingHumanProfileData `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a PineappleConfig from the web input.
func (p PineappleWebInput) ToConfig() (domain.PineappleConfig, error) {
	cfg := domain.DefaultPineappleConfig()
	sb, bb := cfg.SmallBlind, cfg.BigBlind
	sbProvided := p.SmallBlind != nil && *p.SmallBlind >= 1
	bbProvided := p.BigBlind != nil && *p.BigBlind >= 1
	if sbProvided {
		sb = *p.SmallBlind
	}
	if bbProvided {
		bb = *p.BigBlind
	}
	if sbProvided && !bbProvided && sb >= cfg.BigBlind {
		bb = sb * 2
	} else if bbProvided && !sbProvided && bb > 1 {
		sb = bb / 2
	}
	if sb >= bb {
		return domain.PineappleConfig{}, errors.New("param error: smallBlind must be less than bigBlind")
	}
	cfg.SmallBlind = sb
	cfg.BigBlind = bb
	if p.TournamentMode != nil {
		cfg.TournamentMode = *p.TournamentMode
	}
	if p.BlindLevelHands != nil && *p.BlindLevelHands >= 1 {
		cfg.BlindLevelHands = *p.BlindLevelHands
	}
	if p.BlindMultiplier != nil && *p.BlindMultiplier >= 101 {
		cfg.BlindMultiplier = *p.BlindMultiplier
	}
	if p.BettingLimit != nil {
		bl := *p.BettingLimit
		if bl < 0 {
			bl = 0
		} else if bl > 2 {
			bl = 2
		}
		cfg.BettingLimit = domain.BettingLimitType(bl)
	}
	if p.TableSize != nil {
		ts := *p.TableSize
		if !domain.IsValidHoldemTableSize(ts) {
			return domain.PineappleConfig{}, errors.New("param error: tableSize must be 4, 6, or 9")
		}
		cfg.TableSize = ts
	}
	if p.RebuyEnabled != nil {
		cfg.RebuyEnabled = *p.RebuyEnabled
	}
	if p.RebuyMaxCount != nil && *p.RebuyMaxCount >= 1 {
		cfg.RebuyMaxCount = *p.RebuyMaxCount
	}
	if p.RebuyChips != nil && *p.RebuyChips >= 1 {
		cfg.RebuyChips = *p.RebuyChips
	}
	if p.RebuyPeriodHands != nil && *p.RebuyPeriodHands >= 1 {
		cfg.RebuyPeriodHands = *p.RebuyPeriodHands
	}
	if p.AddonEnabled != nil {
		cfg.AddonEnabled = *p.AddonEnabled
	}
	if p.AddonChips != nil && *p.AddonChips >= 1 {
		cfg.AddonChips = *p.AddonChips
	}
	if p.AddonAfterHand != nil && *p.AddonAfterHand >= 1 {
		cfg.AddonAfterHand = *p.AddonAfterHand
	}
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg, nil
}

// PineappleWebController パイナップルポーカーWebコントローラークラス
type PineappleWebController = GameWebController[usecase.PineappleInteractorIF, PineappleWebInput, *PineappleWebOutput]

// NewPineappleWebController and NewPineappleWebControllerWithProvider are
// the standard and provider-backed constructors for PineappleWebController.
var NewPineappleWebController, NewPineappleWebControllerWithProvider = WebControllerPair[usecase.PineappleInteractorIF, PineappleWebInput, *PineappleWebOutput](
	newPineappleDefaultOutput, pineappleDispatch,
)

func newPineappleDefaultOutput(msg string) *PineappleWebOutput {
	return &PineappleWebOutput{
		Players:        make([]*PineappleWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*PineappleWebOutputSidePot, 0),
		RoundResults:   make([]*PineappleWebOutputResult, 0),
		CpuActions:     make([]*PineappleWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func pineappleDispatch(bc *baseController, w http.ResponseWriter, pgi usecase.PineappleInteractorIF, param PineappleWebInput, newDefault func(string) *PineappleWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, pgi.ResetWithConfig(cfg, param.Profile))
	case "f", "fold":
		bc.writePresenterResponse(w, pgi.Action(domain.PineappleActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, pgi.Action(domain.PineappleActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, pgi.Action(domain.PineappleActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, pgi.Action(domain.PineappleActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, pgi.Action(domain.PineappleActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, pgi.Action(domain.PineappleActionAllIn, 0, param.HumanPlayMs))
	case "d", "discard":
		if param.CardIdx == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: cardIdx is required for discard"))
			return true
		}
		bc.writePresenterResponse(w, pgi.Discard(*param.CardIdx))
	case "rb", "rebuy":
		bc.writePresenterResponse(w, pgi.Rebuy())
	case "sr", "skiprebuy":
		bc.writePresenterResponse(w, pgi.SkipRebuy())
	case "ad", "addon":
		bc.writePresenterResponse(w, pgi.Addon())
	case "sa", "skipaddon":
		bc.writePresenterResponse(w, pgi.SkipAddon())
	case "m", "muck":
		bc.writePresenterResponse(w, pgi.Muck())
	case "sh", "show":
		bc.writePresenterResponse(w, pgi.ShowHand())
	default:
		return dispatchLog(param.Command, bc, w, pgi.ActionLog)
	}
	return true
}
