//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TablanetWebConfig はタブラネット (Tablanet) の Web 設定。
type TablanetWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// ToConfig は TablanetWebConfig を domain.TablanetConfig に変換する (境界チェック付き)。
func (c *TablanetWebConfig) ToConfig() domain.TablanetConfig {
	cfg := domain.DefaultTablanetConfig()
	cfg.CpuDifficulty = domain.TablanetCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty, int(domain.TablanetCpuDifficultyEasy), int(domain.TablanetCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// TablanetWebInput はタブラネット Web インプット。
type TablanetWebInput struct {
	BaseWebInput
	CardIndex    *int               `json:"cardIndex,omitempty"`
	TableIndices []int              `json:"tableIndices,omitempty"`
	Config       *TablanetWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.TablanetConfig を構築する。
func (p TablanetWebInput) ToConfig() domain.TablanetConfig {
	return configOrDefault(p.Config, (*TablanetWebConfig).ToConfig, domain.DefaultTablanetConfig())
}

// TablanetWebOutputPlayer はタブラネット Web アウトプットプレイヤー。
type TablanetWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	CardCount     int              `json:"cardCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	TablaCount    int              `json:"tablaCount"`
	Score         int              `json:"score"`
}

// TablanetWebOutputScoreDetail は 1 ゲームの得点内訳。
type TablanetWebOutputScoreDetail struct {
	Cards          map[int]int `json:"cards"`
	Aces           map[int]int `json:"aces"`
	Jacks          map[int]int `json:"jacks"`
	Tablas         map[int]int `json:"tablas"`
	HasTenDiamonds int         `json:"hasTenDiamonds"`
	HasTwoClubs    int         `json:"hasTwoClubs"`
	MostCards      int         `json:"mostCards"`
	Gained         map[int]int `json:"gained"`
}

// TablanetWebOutputHint はヒント出力。
type TablanetWebOutputHint struct {
	CardIndices  []int  `json:"cardIndices"`
	TableIndices []int  `json:"tableIndices"`
	Reason       string `json:"reason"`
}

// TablanetWebOutput はタブラネット Web アウトプット。
type TablanetWebOutput struct {
	Players         []*TablanetWebOutputPlayer    `json:"players"`
	Phase           int                           `json:"phase"`
	RoundNumber     int                           `json:"roundNumber"`
	CurrentTurn     int                           `json:"currentTurn"`
	TableCards      []*WebOutputCard              `json:"tableCards"`
	LastCaptureIdx  int                           `json:"lastCaptureIdx"`
	RemainingDeck   int                           `json:"remainingDeck"`
	PlayableIndices []int                         `json:"playableIndices"`
	CaptureOptions  map[int][]int                 `json:"captureOptions"`
	Winners         []int                         `json:"winners"`
	GameEndFlag     bool                          `json:"gameEndFlag"`
	LastDealDetail  *TablanetWebOutputScoreDetail `json:"lastDealDetail"`
	IsHumanTurn     bool                          `json:"isHumanTurn"`
	Hint            *TablanetWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config TablanetWebConfigOutput `json:"config"`
}

// TablanetWebConfigOutput は設定アウトプット。
type TablanetWebConfigOutput struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// TablanetWebController はタブラネット Web コントローラークラス。
type TablanetWebController = GameWebController[usecase.TablanetInteractorIF, TablanetWebInput, *TablanetWebOutput]

// NewTablanetWebController, NewTablanetWebControllerWithProvider are the standard and
// provider-backed constructors for TablanetWebController.
var NewTablanetWebController, NewTablanetWebControllerWithProvider = webControllerPair[usecase.TablanetInteractorIF, TablanetWebInput, *TablanetWebOutput](
	newTablanetDefaultOutput, tablanetDispatch,
)

func newTablanetDefaultOutput(msg string) *TablanetWebOutput {
	return &TablanetWebOutput{
		Players:         make([]*TablanetWebOutputPlayer, 0),
		TableCards:      make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		CaptureOptions:  make(map[int][]int),
		Winners:         make([]int, 0),
		LastCaptureIdx:  -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func tablanetDispatch(bc *baseController, w http.ResponseWriter, bi usecase.TablanetInteractorIF, param TablanetWebInput, newDefault func(string) *TablanetWebOutput) bool {
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
