//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GoStopWebConfig はゴーストップ (Go-Stop) の Web 設定。
type GoStopWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// ToConfig は GoStopWebConfig を domain.GoStopConfig に変換する (境界チェック付き)。
func (c *GoStopWebConfig) ToConfig() domain.GoStopConfig {
	cfg := domain.DefaultGoStopConfig()
	cfg.CpuDifficulty = domain.GoStopCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty, int(domain.GoStopCpuDifficultyEasy), int(domain.GoStopCpuDifficultyHard), int(cfg.CpuDifficulty)))
	cfg.TargetScore = webutil.BoundedIntPtr(
		c.TargetScore, domain.GoStopTargetScoreMin, domain.GoStopTargetScoreMax, cfg.TargetScore)
	return cfg
}

// GoStopWebInput はゴーストップ Web インプット。
type GoStopWebInput struct {
	BaseWebInput
	CardIndex  *int             `json:"cardIndex,omitempty"`
	FieldIndex *int             `json:"fieldIndex,omitempty"`
	Config     *GoStopWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.GoStopConfig を構築する。
func (p GoStopWebInput) ToConfig() domain.GoStopConfig {
	return configOrDefault(p.Config, (*GoStopWebConfig).ToConfig, domain.DefaultGoStopConfig())
}

// GoStopWebOutputBreakdown は得点内訳。
type GoStopWebOutputBreakdown struct {
	Gwang       int `json:"gwang"`
	Godori      int `json:"godori"`
	Tti         int `json:"tti"`
	Yeol        int `json:"yeol"`
	Pi          int `json:"pi"`
	Base        int `json:"base"`
	GoCount     int `json:"goCount"`
	GoMult      int `json:"goMult"`
	GoScore     int `json:"goScore"`
	BrightCount int `json:"brightCount"`
	RibbonCount int `json:"ribbonCount"`
	AnimalCount int `json:"animalCount"`
	PiCount     int `json:"piCount"`
}

// GoStopWebOutputPlayer はゴーストップ Web アウトプットプレイヤー。
type GoStopWebOutputPlayer struct {
	ID            int                       `json:"id"`
	IsHuman       bool                      `json:"isHuman"`
	CardCount     int                       `json:"cardCount"`
	Cards         []*WebOutputCard          `json:"cards"`
	Captured      []*WebOutputCard          `json:"captured"`
	CapturedCount int                       `json:"capturedCount"`
	Score         int                       `json:"score"`
	GoCount       int                       `json:"goCount"`
	Breakdown     *GoStopWebOutputBreakdown `json:"breakdown"`
	Points        int                       `json:"points"`
}

// GoStopWebOutputRoundResult は 1 ラウンドの結果。
type GoStopWebOutputRoundResult struct {
	Winner     int                       `json:"winner"`
	Breakdown  *GoStopWebOutputBreakdown `json:"breakdown"`
	BasePoints int                       `json:"basePoints"`
	GoScore    int                       `json:"goScore"`
	BakMult    int                       `json:"bakMult"`
	Total      int                       `json:"total"`
	GwangBak   bool                      `json:"gwangBak"`
	PiBak      bool                      `json:"piBak"`
	GoBak      bool                      `json:"goBak"`
	GoCount    int                       `json:"goCount"`
}

// GoStopWebOutputHint はヒント出力。
type GoStopWebOutputHint struct {
	CardIndex  int    `json:"cardIndex"`
	FieldIndex int    `json:"fieldIndex"`
	Go         int    `json:"go"`
	Reason     string `json:"reason"`
}

// GoStopWebOutput はゴーストップ Web アウトプット。
type GoStopWebOutput struct {
	Players          []*GoStopWebOutputPlayer    `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	CurrentTurn      int                         `json:"currentTurn"`
	FieldCards       []*WebOutputCard            `json:"fieldCards"`
	RemainingDeck    int                         `json:"remainingDeck"`
	PlayableIndices  []int                       `json:"playableIndices"`
	CaptureOptions   map[int][]int               `json:"captureOptions"`
	PendingBreakdown *GoStopWebOutputBreakdown   `json:"pendingBreakdown"`
	PendingPoints    int                         `json:"pendingPoints"`
	RoundWinner      int                         `json:"roundWinner"`
	Winner           int                         `json:"winner"`
	Result           int                         `json:"result"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	IsHumanTurn      bool                        `json:"isHumanTurn"`
	LastRoundResult  *GoStopWebOutputRoundResult `json:"lastRoundResult"`
	Hint             *GoStopWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config GoStopWebConfigOutput `json:"config"`
}

// GoStopWebConfigOutput は設定アウトプット。
type GoStopWebConfigOutput struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// GoStopWebController はゴーストップ Web コントローラークラス。
type GoStopWebController = GameWebController[usecase.GoStopInteractorIF, GoStopWebInput, *GoStopWebOutput]

// NewGoStopWebController, NewGoStopWebControllerWithProvider are the standard and
// provider-backed constructors for GoStopWebController.
var NewGoStopWebController, NewGoStopWebControllerWithProvider = webControllerPair[usecase.GoStopInteractorIF, GoStopWebInput, *GoStopWebOutput](
	newGoStopDefaultOutput, gostopDispatch,
)

func newGoStopDefaultOutput(msg string) *GoStopWebOutput {
	return &GoStopWebOutput{
		Players:         make([]*GoStopWebOutputPlayer, 0),
		FieldCards:      make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		CaptureOptions:  make(map[int][]int),
		RoundWinner:     -1,
		Winner:          -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func gostopDispatch(bc *baseController, w http.ResponseWriter, ki usecase.GoStopInteractorIF, param GoStopWebInput, newDefault func(string) *GoStopWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		fieldIdx := -1
		if param.FieldIndex != nil {
			fieldIdx = *param.FieldIndex
		}
		bc.writePresenterResponse(w, ki.Play(*param.CardIndex, fieldIdx))
	case "go":
		bc.writePresenterResponse(w, ki.Decide(true))
	case "st", "stop":
		bc.writePresenterResponse(w, ki.Decide(false))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ki.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ki.Hint, ki.ActionLog)
	}
	return true
}
