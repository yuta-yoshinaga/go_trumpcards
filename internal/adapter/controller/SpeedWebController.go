package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpeedWebInput スピードWebインプット
type SpeedWebInput struct {
	BaseWebInput
	CardIndex     *int `json:"cardIndex,omitempty"`
	PileIndex     *int `json:"pileIndex,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// ToConfig インプットからSpeedConfigを生成する
func (in SpeedWebInput) ToConfig() domain.SpeedConfig {
	cfg := domain.DefaultSpeedConfig()
	if in.CpuDifficulty != nil {
		cfg.CpuDifficulty = domain.SpeedCpuDifficulty(webutil.BoundedIntPtr(in.CpuDifficulty, int(domain.SpeedCpuDifficultyEasy), int(domain.SpeedCpuDifficultyHard), int(cfg.CpuDifficulty)))
	}
	return cfg
}

// SpeedWebOutputPlayer スピードWebアウトプットプレイヤー
type SpeedWebOutputPlayer struct {
	ID           int              `json:"id"`
	IsHuman      bool             `json:"isHuman"`
	CardCount    int              `json:"cardCount"`
	Cards        []*WebOutputCard `json:"cards"`
	DrawPileSize int              `json:"drawPileSize"`
}

// SpeedWebOutputCpuAction CPU行動記録
type SpeedWebOutputCpuAction struct {
	CardIndex int `json:"cardIndex"`
	PileIndex int `json:"pileIndex"`
}

// SpeedWebOutputHint ヒント情報
type SpeedWebOutputHint struct {
	CardIndex int  `json:"cardIndex"`
	PileIndex int  `json:"pileIndex"`
	Found     bool `json:"found"`
}

// SpeedWebOutputConfig 設定情報
type SpeedWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// SpeedWebOutput スピードWebアウトプット
type SpeedWebOutput struct {
	Players     []*SpeedWebOutputPlayer    `json:"players"`
	CenterPiles []*WebOutputCard           `json:"centerPiles"`
	Phase       int                        `json:"phase"`
	GameEndFlag bool                       `json:"gameEndFlag"`
	WinnerIdx   int                        `json:"winnerIdx"`
	CpuActions  []*SpeedWebOutputCpuAction `json:"cpuActions,omitempty"`
	Hint        *SpeedWebOutputHint        `json:"hint,omitempty"`
	Config      SpeedWebOutputConfig       `json:"config"`
	WebOutputBase
}

// SpeedWebController スピードWebコントローラー型
type SpeedWebController = GameWebController[usecase.SpeedInteractorIF, SpeedWebInput, *SpeedWebOutput]

// NewSpeedWebController コンストラクタ
func NewSpeedWebController(factory func() usecase.SpeedInteractorIF) *SpeedWebController {
	return NewGameWebController(factory, newSpeedDefaultOutput, speedDispatch)
}

// NewSpeedWebControllerWithProvider KVプロバイダ付きコンストラクタ
func NewSpeedWebControllerWithProvider(
	provider SessionProvider[usecase.SpeedInteractorIF],
	factory func() usecase.SpeedInteractorIF,
) *SpeedWebController {
	return NewGameWebControllerWithProvider(provider, factory, newSpeedDefaultOutput, speedDispatch)
}

func newSpeedDefaultOutput(msg string) *SpeedWebOutput {
	o := new(SpeedWebOutput)
	o.Message = msg
	return o
}

func speedDispatch(
	bc *baseController,
	w http.ResponseWriter,
	si usecase.SpeedInteractorIF,
	param SpeedWebInput,
	newDefault func(string) *SpeedWebOutput,
) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if param.CardIndex == nil || param.PileIndex == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error"))
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex, *param.PileIndex))
	case "f", "flip":
		bc.writePresenterResponse(w, si.Flip())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
