//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EscobaWebConfig ローカルルール設定 (入力・出力共用)
type EscobaWebConfig struct {
	TargetScore   int `json:"targetScore"`
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts EscobaWebConfig to domain.EscobaConfig.
func (c EscobaWebConfig) ToConfig() domain.EscobaConfig {
	return domain.EscobaConfig{
		TargetScore:   c.TargetScore,
		CpuDifficulty: domain.EscobaCpuDifficulty(c.CpuDifficulty),
	}
}

// EscobaWebInput エスコバWebインプット
type EscobaWebInput struct {
	BaseWebInput
	HandIndex    int              `json:"handIndex"`
	TableIndices []int            `json:"tableIndices"`
	Config       *EscobaWebConfig `json:"config"`
}

// EscobaWebOutputPlayer エスコバWebアウトプットプレイヤー
type EscobaWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	HandCount     int              `json:"handCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	// CapturedCards は獲得した捕獲札の中身。人間プレイヤーのみに公開し、CPU は
	// 記憶ゲーム性を保つため空スライスとする (枚数は CapturedCount で確認可能)。
	CapturedCards []*WebOutputCard `json:"capturedCards"`
	EscobaCount   int              `json:"escobaCount"`
	Score         int              `json:"score"`
}

// EscobaWebOutputScoreDetail 得点内訳 (プレイヤー単位)
type EscobaWebOutputScoreDetail struct {
	Cards      [domain.EscobaPlayerCnt]int `json:"cards"`
	Espadas    [domain.EscobaPlayerCnt]int `json:"espadas"`
	Sevens     [domain.EscobaPlayerCnt]int `json:"sevens"`
	Oros       [domain.EscobaPlayerCnt]int `json:"oros"`
	Escobas    [domain.EscobaPlayerCnt]int `json:"escobas"`
	Gained     [domain.EscobaPlayerCnt]int `json:"gained"`
	AceEspada  int                         `json:"aceEspada"`
	SeteEspada int                         `json:"seteEspada"`
}

// EscobaWebOutput エスコバWebアウトプット
type EscobaWebOutput struct {
	Players         []*EscobaWebOutputPlayer    `json:"players"`
	TableCards      []*WebOutputCard            `json:"tableCards"`
	Phase           string                      `json:"phase"`
	RoundNumber     int                         `json:"roundNumber"`
	CurrentTurn     int                         `json:"currentTurn"`
	DealerIdx       int                         `json:"dealerIdx"`
	StockRemaining  int                         `json:"stockRemaining"`
	LastCaptureIdx  int                         `json:"lastCaptureIdx"`
	WinnerIdx       int                         `json:"winnerIdx"`
	GameEndFlag     bool                        `json:"gameEndFlag"`
	IsHumanTurn     bool                        `json:"isHumanTurn"`
	HandCaptures    [][][]int                   `json:"handCaptures"`
	LastRoundDetail *EscobaWebOutputScoreDetail `json:"lastRoundDetail"`
	Config          EscobaWebConfig             `json:"config"`
	WebOutputBase
}

// EscobaWebController エスコバWebコントローラークラス
type EscobaWebController = GameWebController[usecase.EscobaInteractorIF, EscobaWebInput, *EscobaWebOutput]

// NewEscobaWebController, NewEscobaWebControllerWithProvider are the standard and
// provider-backed constructors for EscobaWebController.
var NewEscobaWebController, NewEscobaWebControllerWithProvider = webControllerPair[usecase.EscobaInteractorIF, EscobaWebInput, *EscobaWebOutput](
	newEscobaDefaultOutput, escobaDispatch,
)

func newEscobaDefaultOutput(msg string) *EscobaWebOutput {
	return &EscobaWebOutput{
		Players:       make([]*EscobaWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		HandCaptures:  make([][][]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

// escobaMaxIndices caps client-supplied table-index slices for the play command.
const escobaMaxIndices = 40

func escobaDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EscobaInteractorIF, param EscobaWebInput, defaultOutput func(string) *EscobaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, ei.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, ei.Reset())
		}
	case "n", "next":
		bc.writePresenterResponse(w, ei.NextRound())
	case "p", "play":
		if len(param.TableIndices) > escobaMaxIndices {
			bc.writeJsonResponse(w, http.StatusBadRequest, defaultOutput("too many indices"))
			return true
		}
		bc.writePresenterResponse(w, ei.Play(param.HandIndex, param.TableIndices))
	case "config":
		if param.Config != nil {
			bc.writePresenterResponse(w, ei.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, ei.Reset())
		}
	default:
		return dispatchHintAndLog(param.Command, bc, w, ei.Hint, ei.ActionLog)
	}
	return true
}
