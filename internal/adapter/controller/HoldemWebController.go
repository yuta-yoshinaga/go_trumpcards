package controller

import (
	"errors"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// HoldemWebInput テキサスホールデムWebインプット
type HoldemWebInput struct {
	BaseWebInput
	Amount           int   `json:"amount,omitempty"`
	HumanPlayMs      int   `json:"humanPlayMs,omitempty"`
	SmallBlind       *int  `json:"smallBlind,omitempty"`
	BigBlind         *int  `json:"bigBlind,omitempty"`
	TournamentMode   *bool `json:"tournamentMode,omitempty"`
	BlindLevelHands  *int  `json:"blindLevelHands,omitempty"`
	BlindMultiplier  *int  `json:"blindMultiplier,omitempty"`
	BettingLimit     *int  `json:"bettingLimit,omitempty"`
	TableSize        *int  `json:"tableSize,omitempty"`
	RebuyEnabled     *bool `json:"rebuyEnabled,omitempty"`
	RebuyMaxCount    *int  `json:"rebuyMaxCount,omitempty"`
	RebuyChips       *int  `json:"rebuyChips,omitempty"`
	RebuyPeriodHands *int  `json:"rebuyPeriodHands,omitempty"`
	AddonEnabled     *bool `json:"addonEnabled,omitempty"`
	AddonChips       *int  `json:"addonChips,omitempty"`
	AddonAfterHand   *int  `json:"addonAfterHand,omitempty"`
	CpuMetaAI        bool  `json:"cpuMetaAI,omitempty"`
}

// HoldemWebOutputPlayer テキサスホールデムWebアウトプットプレイヤー
type HoldemWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	Cards         []*WebOutputCard `json:"cards"`
	Chips         int              `json:"chips"`
	CurrentBet    int              `json:"currentBet"`
	Folded        bool             `json:"folded"`
	AllIn         bool             `json:"allIn"`
	HandRank      int              `json:"handRank"`
	HandName      string           `json:"handName"`
	BestHand      []*WebOutputCard `json:"bestHand"`
	PlayStyleName string           `json:"playStyleName"`
	TotalHands    int              `json:"totalHands"`
	VPIP          int              `json:"vpip"`
	PFR           int              `json:"pfr"`
	ThreeBet      int              `json:"threeBet"`
	AF            string           `json:"af"`
}

// HoldemWebOutputCpuAction テキサスホールデムCPU行動記録
type HoldemWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// HoldemWebOutputResult テキサスホールデムショーダウン結果
type HoldemWebOutputResult struct {
	PlayerIdx int              `json:"playerIdx"`
	HandRank  int              `json:"handRank"`
	HandName  string           `json:"handName"`
	Kickers   string           `json:"kickers"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
	Mucked    bool             `json:"mucked"`
}

// HoldemWebOutputSidePot テキサスホールデムサイドポット
type HoldemWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// HoldemWebOutputEquity テキサスホールデムエクイティ情報
type HoldemWebOutputEquity struct {
	WinProbability float64                    `json:"winProbability"`
	HandOdds       []*HoldemWebOutputHandOdds `json:"handOdds"`
}

// HoldemWebOutputHandOdds テキサスホールデムハンドオッズ
type HoldemWebOutputHandOdds struct {
	HandRank    int     `json:"handRank"`
	HandName    string  `json:"handName"`
	Probability float64 `json:"probability"`
}

// HoldemWebOutput テキサスホールデムWebアウトプット
type HoldemWebOutput struct {
	Players          []*HoldemWebOutputPlayer    `json:"players"`
	CommunityCards   []*WebOutputCard            `json:"communityCards"`
	Pot              int                         `json:"pot"`
	SidePots         []*HoldemWebOutputSidePot   `json:"sidePots"`
	DealerIdx        int                         `json:"dealerIdx"`
	CurrentTurn      int                         `json:"currentTurn"`
	Phase            int                         `json:"phase"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	LastBet          int                         `json:"lastBet"`
	MinRaise         int                         `json:"minRaise"`
	BettingLimit     int                         `json:"bettingLimit"`
	RaiseCount       int                         `json:"raiseCount"`
	MaxBetAmount     int                         `json:"maxBetAmount"`
	RoundResults     []*HoldemWebOutputResult    `json:"roundResults"`
	CpuActions       []*HoldemWebOutputCpuAction `json:"cpuActions"`
	HandCount        int                         `json:"handCount"`
	SmallBlind       int                         `json:"smallBlind"`
	BigBlind         int                         `json:"bigBlind"`
	TournamentMode   bool                        `json:"tournamentMode"`
	BlindLevelHands  int                         `json:"blindLevelHands"`
	BlindMultiplier  int                         `json:"blindMultiplier"`
	TableSize        int                         `json:"tableSize"`
	RebuyAvailable   bool                        `json:"rebuyAvailable"`
	AddonAvailable   bool                        `json:"addonAvailable"`
	RebuyCounts      []int                       `json:"rebuyCounts"`
	AddonUsed        []bool                      `json:"addonUsed"`
	RebuyEnabled     bool                        `json:"rebuyEnabled"`
	AddonEnabled     bool                        `json:"addonEnabled"`
	RebuyMaxCount    int                         `json:"rebuyMaxCount"`
	RebuyChips       int                         `json:"rebuyChips"`
	AddonChips       int                         `json:"addonChips"`
	RebuyPeriodHands int                         `json:"rebuyPeriodHands"`
	AddonAfterHand   int                         `json:"addonAfterHand"`
	RebuyPhaseType   int                         `json:"rebuyPhaseType"`
	MuckAvailable    bool                        `json:"muckAvailable"`
	Equity           *HoldemWebOutputEquity      `json:"equity,omitempty"`
	PotOdds          *float64                    `json:"potOdds,omitempty"`
	WebOutputBase
}

// ToConfig builds a HoldemConfig from the web input.
// Returns an error if the blind configuration is invalid or tableSize is invalid.
func (p HoldemWebInput) ToConfig() (domain.HoldemConfig, error) {
	cfg := domain.DefaultHoldemConfig()
	sb, bb := cfg.SmallBlind, cfg.BigBlind
	sbProvided := p.SmallBlind != nil && *p.SmallBlind >= 1
	bbProvided := p.BigBlind != nil && *p.BigBlind >= 1
	if sbProvided {
		sb = *p.SmallBlind
	}
	if bbProvided {
		bb = *p.BigBlind
	}
	// 片方のみ指定された場合、もう片方を自動調整
	if sbProvided && !bbProvided && sb >= cfg.BigBlind {
		bb = sb * 2
	} else if bbProvided && !sbProvided && bb > 1 {
		sb = bb / 2
	}
	if sb >= bb {
		return domain.HoldemConfig{}, errors.New("param error: smallBlind must be less than bigBlind")
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
			return domain.HoldemConfig{}, errors.New("param error: tableSize must be 4, 6, or 9")
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

// HoldemWebController テキサスホールデムWebコントローラークラス
type HoldemWebController = GameWebController[usecase.HoldemInteractorIF, HoldemWebInput, *HoldemWebOutput]

// NewHoldemWebController コンストラクタ
func NewHoldemWebController(factory func() usecase.HoldemInteractorIF) *HoldemWebController {
	return NewGameWebController(factory, newHoldemDefaultOutput, holdemDispatch)
}

func newHoldemDefaultOutput(msg string) *HoldemWebOutput {
	return &HoldemWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func holdemDispatch(bc *baseController, w rest.ResponseWriter, hgi usecase.HoldemInteractorIF, param HoldemWebInput, newDefault func(string) *HoldemWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, hgi.ResetWithConfig(cfg))
	case "f", "fold":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionAllIn, 0, param.HumanPlayMs))
	case "rb", "rebuy":
		bc.writePresenterResponse(w, hgi.Rebuy())
	case "sr", "skiprebuy":
		bc.writePresenterResponse(w, hgi.SkipRebuy())
	case "ad", "addon":
		bc.writePresenterResponse(w, hgi.Addon())
	case "sa", "skipaddon":
		bc.writePresenterResponse(w, hgi.SkipAddon())
	case "m", "muck":
		bc.writePresenterResponse(w, hgi.Muck())
	case "sh", "show":
		bc.writePresenterResponse(w, hgi.ShowHand())
	case "log", "l":
		bc.writePresenterResponse(w, hgi.ActionLog())
	default:
		return false
	}
	return true
}
