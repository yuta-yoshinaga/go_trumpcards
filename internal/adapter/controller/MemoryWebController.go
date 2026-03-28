package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MemoryWebInput 神経衰弱Webインプット
type MemoryWebInput struct {
	BaseWebInput
	Position *int             `json:"position,omitempty"`
	Config   *MemoryWebConfig `json:"config,omitempty"`
}

// MemoryWebConfig 神経衰弱Web設定
type MemoryWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// MemoryWebOutputPlayer 神経衰弱Webアウトプットプレイヤー
type MemoryWebOutputPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	PairCount int  `json:"pairCount"`
}

// MemoryWebOutputBoardCard 神経衰弱Webアウトプットボードカード
type MemoryWebOutputBoardCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
	Taken  bool           `json:"taken"`
}

// MemoryWebOutput 神経衰弱Webアウトプット
type MemoryWebOutput struct {
	Players          []*MemoryWebOutputPlayer    `json:"players"`
	Board            []*MemoryWebOutputBoardCard `json:"board"`
	Phase            int                         `json:"phase"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	FirstFlipPos     int                         `json:"firstFlipPos"`
	SecondFlipPos    int                         `json:"secondFlipPos"`
	LastMatchResult  bool                        `json:"lastMatchResult"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerIdx        int                         `json:"winnerIdx"`
	TurnNumber       int                         `json:"turnNumber"`
	WebOutputBase
	Config MemoryWebOutputConfig `json:"config"`
}

// MemoryWebOutputConfig 神経衰弱設定アウトプット
type MemoryWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a MemoryConfig from the nested web config, applying bounds checking.
func (c *MemoryWebConfig) ToConfig() domain.MemoryConfig {
	cfg := domain.DefaultMemoryConfig()
	cfg.CpuDifficulty = domain.MemoryCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MemoryCpuDifficultyEasy), int(domain.MemoryCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a MemoryConfig from the web input.
func (p MemoryWebInput) ToConfig() domain.MemoryConfig {
	return configOrDefault(p.Config, (*MemoryWebConfig).ToConfig, domain.DefaultMemoryConfig())
}

// MemoryWebController 神経衰弱Webコントローラークラス
type MemoryWebController = GameWebController[usecase.MemoryInteractorIF, MemoryWebInput, *MemoryWebOutput]

// NewMemoryWebController コンストラクタ
func NewMemoryWebController(factory func() usecase.MemoryInteractorIF) *MemoryWebController {
	return NewGameWebController(factory, newMemoryDefaultOutput, memoryDispatch)
}

// NewMemoryWebControllerWithProvider creates a MemoryWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewMemoryWebControllerWithProvider(
	provider SessionProvider[usecase.MemoryInteractorIF],
	factory func() usecase.MemoryInteractorIF,
) *MemoryWebController {
	return NewGameWebControllerWithProvider(provider, factory, newMemoryDefaultOutput, memoryDispatch)
}

func newMemoryDefaultOutput(msg string) *MemoryWebOutput {
	return &MemoryWebOutput{
		Players:       make([]*MemoryWebOutputPlayer, 0),
		Board:         make([]*MemoryWebOutputBoardCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func memoryDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MemoryInteractorIF, param MemoryWebInput, newDefault func(string) *MemoryWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, mi.ResetWithConfig(param.ToConfig()))
	case "f", "flip":
		if param.Position == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: position is required."))
			return true
		}
		bc.writePresenterResponse(w, mi.Flip(*param.Position))
	case "n", "next":
		bc.writePresenterResponse(w, mi.Next())
	case "log", "l":
		bc.writePresenterResponse(w, mi.ActionLog())
	default:
		return false
	}
	return true
}
