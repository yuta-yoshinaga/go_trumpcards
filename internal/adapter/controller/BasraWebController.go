//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BasraWebConfig はバスラ (Basra) の Web 設定。
type BasraWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// ToConfig は BasraWebConfig を domain.BasraConfig に変換する (境界チェック付き)。
func (c *BasraWebConfig) ToConfig() domain.BasraConfig {
	cfg := domain.DefaultBasraConfig()
	cfg.CpuDifficulty = domain.BasraCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty, int(domain.BasraCpuDifficultyEasy), int(domain.BasraCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// BasraWebInput はバスラ Web インプット。
type BasraWebInput struct {
	BaseWebInput
	CardIndex    *int            `json:"cardIndex,omitempty"`
	TableIndices []int           `json:"tableIndices,omitempty"`
	Config       *BasraWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.BasraConfig を構築する。
func (p BasraWebInput) ToConfig() domain.BasraConfig {
	return configOrDefault(p.Config, (*BasraWebConfig).ToConfig, domain.DefaultBasraConfig())
}

// BasraWebOutputPlayer はバスラ Web アウトプットプレイヤー。
type BasraWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	CardCount     int              `json:"cardCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	BasraCount    int              `json:"basraCount"`
	Score         int              `json:"score"`
}

// BasraWebOutputScoreDetail は 1 ゲームの得点内訳。
type BasraWebOutputScoreDetail struct {
	Cards            map[int]int `json:"cards"`
	Aces             map[int]int `json:"aces"`
	Basras           map[int]int `json:"basras"`
	HasSevenDiamonds int         `json:"hasSevenDiamonds"`
	HasTenDiamonds   int         `json:"hasTenDiamonds"`
	MostCards        int         `json:"mostCards"`
	Gained           map[int]int `json:"gained"`
}

// BasraWebOutputHint はヒント出力。
type BasraWebOutputHint struct {
	CardIndices  []int  `json:"cardIndices"`
	TableIndices []int  `json:"tableIndices"`
	Reason       string `json:"reason"`
}

// BasraWebOutput はバスラ Web アウトプット。
type BasraWebOutput struct {
	Players         []*BasraWebOutputPlayer    `json:"players"`
	Phase           int                        `json:"phase"`
	RoundNumber     int                        `json:"roundNumber"`
	CurrentTurn     int                        `json:"currentTurn"`
	TableCards      []*WebOutputCard           `json:"tableCards"`
	LastCaptureIdx  int                        `json:"lastCaptureIdx"`
	RemainingDeck   int                        `json:"remainingDeck"`
	PlayableIndices []int                      `json:"playableIndices"`
	CaptureOptions  map[int][]int              `json:"captureOptions"`
	Winners         []int                      `json:"winners"`
	GameEndFlag     bool                       `json:"gameEndFlag"`
	LastDealDetail  *BasraWebOutputScoreDetail `json:"lastDealDetail"`
	IsHumanTurn     bool                       `json:"isHumanTurn"`
	Hint            *BasraWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config BasraWebConfigOutput `json:"config"`
}

// BasraWebConfigOutput は設定アウトプット。
type BasraWebConfigOutput struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// BasraWebController はバスラ Web コントローラークラス。
type BasraWebController = GameWebController[usecase.BasraInteractorIF, BasraWebInput, *BasraWebOutput]

// NewBasraWebController, NewBasraWebControllerWithProvider are the standard and
// provider-backed constructors for BasraWebController.
var NewBasraWebController, NewBasraWebControllerWithProvider = webControllerPair[usecase.BasraInteractorIF, BasraWebInput, *BasraWebOutput](
	newBasraDefaultOutput, basraDispatch,
)

func newBasraDefaultOutput(msg string) *BasraWebOutput {
	return &BasraWebOutput{
		Players:         make([]*BasraWebOutputPlayer, 0),
		TableCards:      make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		CaptureOptions:  make(map[int][]int),
		Winners:         make([]int, 0),
		LastCaptureIdx:  -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func basraDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BasraInteractorIF, param BasraWebInput, newDefault func(string) *BasraWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex, param.TableIndices))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, bi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
