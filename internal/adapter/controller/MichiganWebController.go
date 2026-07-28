//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MichiganWebConfig はミシガン (Michigan) の Web 設定。
type MichiganWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ToConfig は MichiganWebConfig を domain.MichiganConfig に変換する (境界チェック付き)。
func (c *MichiganWebConfig) ToConfig() domain.MichiganConfig {
	cfg := domain.DefaultMichiganConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.MichiganMinPlayerCount, domain.MichiganMaxPlayerCount)
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, domain.MichiganMinAnte, domain.MichiganMaxAnte)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, domain.MichiganMinStartingChips, domain.MichiganMaxStartingChips)
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.MichiganMinTargetRounds, domain.MichiganMaxTargetRounds)
	return cfg
}

// MichiganWebInput はミシガン Web インプット。
type MichiganWebInput struct {
	BaseWebInput
	// BoodleBets は 4 つのブードルへの賭け分配 (bet コマンドで必須、合計 = ante)。
	BoodleBets *[]int `json:"boodleBets,omitempty"`
	// CardIndex は出すカードの手札インデックス (play コマンドで必須)。
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *MichiganWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.MichiganConfig を構築する。
func (p MichiganWebInput) ToConfig() domain.MichiganConfig {
	return configOrDefault(p.Config, (*MichiganWebConfig).ToConfig, domain.DefaultMichiganConfig())
}

// MichiganWebOutputPlayer は 1 プレイヤーの出力。
type MichiganWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Chips     int              `json:"chips"`
	RoundBet  int              `json:"roundBet"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	IsCurrent bool             `json:"isCurrent"`
	IsWinner  bool             `json:"isWinner"`
}

// MichiganWebOutputBoodle は 1 つのブードルの出力。
type MichiganWebOutputBoodle struct {
	Card      *WebOutputCard `json:"card"`
	Chips     int            `json:"chips"`
	ClaimedBy int            `json:"claimedBy"` // -1 = 未獲得
}

// MichiganWebOutputHint はヒント出力。
type MichiganWebOutputHint struct {
	CardIndex int    `json:"cardIndex"`
	Reason    string `json:"reason"`
}

// MichiganWebOutputConfig は設定アウトプット。
type MichiganWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
	TargetRounds  int `json:"targetRounds"`
}

// MichiganWebOutput はミシガン Web アウトプット。
type MichiganWebOutput struct {
	Players          []*MichiganWebOutputPlayer `json:"players"`
	Boodles          []*MichiganWebOutputBoodle `json:"boodles"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	Ante             int                        `json:"ante"`
	Chips            int                        `json:"chips"`
	BetBudget        int                        `json:"betBudget"`
	HumanBetPlaced   bool                       `json:"humanBetPlaced"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	SeqSuit          int                        `json:"seqSuit"`
	SeqSuitName      string                     `json:"seqSuitName"`
	SeqHighValue     int                        `json:"seqHighValue"`
	NeedNewSequence  bool                       `json:"needNewSequence"`
	DeadHandCount    int                        `json:"deadHandCount"`
	IsHumanTurn      bool                       `json:"isHumanTurn"`
	PlayableIndices  []int                      `json:"playableIndices"`
	WinnerIdx        int                        `json:"winnerIdx"`
	MatchWinnerIdx   int                        `json:"matchWinnerIdx"`
	Result           int                        `json:"result"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	Hint             *MichiganWebOutputHint     `json:"hint,omitempty"`
	Config           MichiganWebOutputConfig    `json:"config"`
	WebOutputBase
}

// MichiganWebController はミシガン Web コントローラークラス。
type MichiganWebController = GameWebController[usecase.MichiganInteractorIF, MichiganWebInput, *MichiganWebOutput]

// NewMichiganWebController, NewMichiganWebControllerWithProvider are the
// standard and provider-backed constructors for MichiganWebController.
var NewMichiganWebController, NewMichiganWebControllerWithProvider = webControllerPair[usecase.MichiganInteractorIF, MichiganWebInput, *MichiganWebOutput](
	newMichiganDefaultOutput, michiganDispatch,
)

func newMichiganDefaultOutput(msg string) *MichiganWebOutput {
	return &MichiganWebOutput{
		Players:         make([]*MichiganWebOutputPlayer, 0),
		Boodles:         make([]*MichiganWebOutputBoodle, 0),
		PlayableIndices: make([]int, 0),
		WinnerIdx:       -1,
		MatchWinnerIdx:  -1,
		SeqSuit:         0,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func michiganDispatch(bc *baseController, w http.ResponseWriter, ti usecase.MichiganInteractorIF, param MichiganWebInput, newDefault func(string) *MichiganWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "bet":
		if !requireParam(bc, w, newDefault, param.BoodleBets == nil, "param error: boodleBets is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Bet(*param.BoodleBets))
	case "play", "p":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
