//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FollowTheQueenWebInput フォロー・ザ・クイーンWebインプット
type FollowTheQueenWebInput struct {
	BaseWebInput
	PokerCommonInput
	Amount         int             `json:"amount,omitempty"`
	HumanPlayMs    int             `json:"humanPlayMs,omitempty"`
	Ante           *int            `json:"ante,omitempty"`
	BringIn        *int            `json:"bringIn,omitempty"`
	SmallBet       *int            `json:"smallBet,omitempty"`
	BigBet         *int            `json:"bigBet,omitempty"`
	AnteLevelHands *int            `json:"anteLevelHands,omitempty"`
	AnteMultiplier *int            `json:"anteMultiplier,omitempty"`
	CpuMetaAI      bool            `json:"cpuMetaAI,omitempty"`
	Profile        json.RawMessage `json:"profile,omitempty"`
}

// FollowTheQueenWebOutputPlayer フォロー・ザ・クイーンWebアウトプットプレイヤー
type FollowTheQueenWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	HoleCards     []*WebOutputCard `json:"holeCards"`
	DoorCards     []*WebOutputCard `json:"doorCards"`
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

// FollowTheQueenWebOutputCpuAction フォロー・ザ・クイーンCPU行動記録
type FollowTheQueenWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// FollowTheQueenWebOutputResult フォロー・ザ・クイーンショーダウン結果
type FollowTheQueenWebOutputResult struct {
	PlayerIdx int              `json:"playerIdx"`
	HandRank  int              `json:"handRank"`
	HandName  string           `json:"handName"`
	Kickers   string           `json:"kickers"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
	Mucked    bool             `json:"mucked"`
}

// FollowTheQueenWebOutputSidePot フォロー・ザ・クイーンサイドポット
type FollowTheQueenWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// FollowTheQueenWebOutputMetaAI メタAI情報
type FollowTheQueenWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// FollowTheQueenWebOutput フォロー・ザ・クイーンWebアウトプット
type FollowTheQueenWebOutput struct {
	Players       []*FollowTheQueenWebOutputPlayer  `json:"players"`
	CommunityCard *WebOutputCard                    `json:"communityCard"`
	Pot           int                               `json:"pot"`
	SidePots      []*FollowTheQueenWebOutputSidePot `json:"sidePots"`
	DealerIdx     int                               `json:"dealerIdx"`
	CurrentTurn   int                               `json:"currentTurn"`
	// WildRank は現在ワイルドのランク (0 = ワイルド無し)。**画面に出すために送る** ——
	// 表向きのクイーンが出た次のカードのランクが全員のワイルドになるので、
	// いま何がワイルドかが分からないと自分の役すら読めない。
	WildRank         int                                 `json:"wildRank"`
	Phase            int                                 `json:"phase"`
	GameEndFlag      bool                                `json:"gameEndFlag"`
	LastBet          int                                 `json:"lastBet"`
	MinRaise         int                                 `json:"minRaise"`
	BettingLimit     int                                 `json:"bettingLimit"`
	RaiseCount       int                                 `json:"raiseCount"`
	MaxBetAmount     int                                 `json:"maxBetAmount"`
	RoundResults     []*FollowTheQueenWebOutputResult    `json:"roundResults"`
	CpuActions       []*FollowTheQueenWebOutputCpuAction `json:"cpuActions"`
	HandCount        int                                 `json:"handCount"`
	Ante             int                                 `json:"ante"`
	BringIn          int                                 `json:"bringIn"`
	SmallBet         int                                 `json:"smallBet"`
	BigBet           int                                 `json:"bigBet"`
	TournamentMode   bool                                `json:"tournamentMode"`
	AnteLevelHands   int                                 `json:"anteLevelHands"`
	AnteMultiplier   int                                 `json:"anteMultiplier"`
	TableSize        int                                 `json:"tableSize"`
	BringInPlayerIdx int                                 `json:"bringInPlayerIdx"`
	RebuyAvailable   bool                                `json:"rebuyAvailable"`
	AddonAvailable   bool                                `json:"addonAvailable"`
	RebuyCounts      []int                               `json:"rebuyCounts"`
	AddonUsed        []bool                              `json:"addonUsed"`
	RebuyEnabled     bool                                `json:"rebuyEnabled"`
	AddonEnabled     bool                                `json:"addonEnabled"`
	RebuyMaxCount    int                                 `json:"rebuyMaxCount"`
	RebuyChips       int                                 `json:"rebuyChips"`
	AddonChips       int                                 `json:"addonChips"`
	RebuyPeriodHands int                                 `json:"rebuyPeriodHands"`
	AddonAfterHand   int                                 `json:"addonAfterHand"`
	RebuyPhaseType   int                                 `json:"rebuyPhaseType"`
	MuckAvailable    bool                                `json:"muckAvailable"`
	MetaAI           *FollowTheQueenWebOutputMetaAI      `json:"metaAI,omitempty"`
	Profile          *domain.BettingHumanProfileData     `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a FollowTheQueenConfig from the web input.
// Returns an error if the table size is invalid.
func (p FollowTheQueenWebInput) ToConfig() (domain.FollowTheQueenConfig, error) {
	cfg := domain.DefaultFollowTheQueenConfig()
	applyIntIfGte(&cfg.Ante, p.Ante, 1)
	applyIntIfGte(&cfg.BringIn, p.BringIn, 1)
	applyIntIfGte(&cfg.SmallBet, p.SmallBet, 1)
	applyIntIfGte(&cfg.BigBet, p.BigBet, 1)
	applyBool(&cfg.TournamentMode, p.TournamentMode)
	applyIntIfGte(&cfg.AnteLevelHands, p.AnteLevelHands, 1)
	applyIntIfGte(&cfg.AnteMultiplier, p.AnteMultiplier, 101)
	applyBettingLimit(&cfg.BettingLimit, p.BettingLimit)
	if err := applyTableSize(&cfg.TableSize, p.TableSize, domain.IsValidFollowTheQueenTableSize, "param error: tableSize must be 2-7"); err != nil {
		return domain.FollowTheQueenConfig{}, err
	}
	applyRebuyConfig(&cfg.RebuyEnabled, &cfg.RebuyMaxCount, &cfg.RebuyChips, &cfg.RebuyPeriodHands,
		p.RebuyEnabled, p.RebuyMaxCount, p.RebuyChips, p.RebuyPeriodHands)
	applyAddonConfig(&cfg.AddonEnabled, &cfg.AddonChips, &cfg.AddonAfterHand,
		p.AddonEnabled, p.AddonChips, p.AddonAfterHand)
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg, nil
}

// FollowTheQueenWebController フォロー・ザ・クイーンWebコントローラークラス
type FollowTheQueenWebController = GameWebController[usecase.FollowTheQueenInteractorIF, FollowTheQueenWebInput, *FollowTheQueenWebOutput]

// NewFollowTheQueenWebController and NewFollowTheQueenWebControllerWithProvider are
// the standard and provider-backed constructors for FollowTheQueenWebController.
var NewFollowTheQueenWebController, NewFollowTheQueenWebControllerWithProvider = webControllerPair[usecase.FollowTheQueenInteractorIF, FollowTheQueenWebInput, *FollowTheQueenWebOutput](
	newFollowTheQueenDefaultOutput, followTheQueenDispatch,
)

func newFollowTheQueenDefaultOutput(msg string) *FollowTheQueenWebOutput {
	return &FollowTheQueenWebOutput{
		Players:       make([]*FollowTheQueenWebOutputPlayer, 0),
		SidePots:      make([]*FollowTheQueenWebOutputSidePot, 0),
		RoundResults:  make([]*FollowTheQueenWebOutputResult, 0),
		CpuActions:    make([]*FollowTheQueenWebOutputCpuAction, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func followTheQueenDispatch(bc *baseController, w http.ResponseWriter, sgi usecase.FollowTheQueenInteractorIF, param FollowTheQueenWebInput, newDefault func(string) *FollowTheQueenWebOutput) bool {
	if dispatchPokerAction(bc, w, sgi, param.Command, param.Amount, param.HumanPlayMs) {
		return true
	}
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, sgi.ResetWithConfig(cfg, param.Profile))
	default:
		return dispatchHintAndLog(param.Command, bc, w, sgi.Hint, sgi.ActionLog)
	}
	return true
}
