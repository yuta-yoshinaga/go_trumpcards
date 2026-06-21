//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OpenFaceChineseWebInput オープンフェイス・チャイニーズポーカー (OFC) のWebインプット
type OpenFaceChineseWebInput struct {
	BaseWebInput
	// Row 置く段 (0=front,1=middle,2=back)
	Row *int `json:"row,omitempty"`
	// Config ゲーム設定
	Config *OpenFaceChineseWebConfig `json:"config,omitempty"`
}

// OpenFaceChineseWebConfig オープンフェイス・チャイニーズポーカー (OFC) のWeb設定
type OpenFaceChineseWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PlayerCount   *int `json:"playerCount,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// OpenFaceChineseWebOutputPlayer オープンフェイス・チャイニーズポーカー (OFC) のWebアウトプットプレイヤー
type OpenFaceChineseWebOutputPlayer struct {
	ID          int              `json:"id"`
	IsHuman     bool             `json:"isHuman"`
	Front       []*WebOutputCard `json:"front"`
	Middle      []*WebOutputCard `json:"middle"`
	Back        []*WebOutputCard `json:"back"`
	Pending     []*WebOutputCard `json:"pending"`
	RoundScore  int              `json:"roundScore"`
	Royalty     int              `json:"royalty"`
	Fouled      bool             `json:"fouled"`
	Fantasyland bool             `json:"fantasyland"`
	TotalScore  int              `json:"totalScore"`
}

// OpenFaceChineseWebOutputHint ヒント出力
type OpenFaceChineseWebOutputHint struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

// OpenFaceChineseWebOutput オープンフェイス・チャイニーズポーカー (OFC) のWebアウトプット
type OpenFaceChineseWebOutput struct {
	Players          []*OpenFaceChineseWebOutputPlayer `json:"players"`
	Phase            int                               `json:"phase"`
	RoundNumber      int                               `json:"roundNumber"`
	CurrentPlayerIdx int                               `json:"currentPlayerIdx"`
	DealerIdx        int                               `json:"dealerIdx"`
	CurrentCard      *WebOutputCard                    `json:"currentCard,omitempty"`
	GameEndFlag      bool                              `json:"gameEndFlag"`
	WinnerIdx        int                               `json:"winnerIdx"`
	IsHumanTurn      bool                              `json:"isHumanTurn"`
	Hint             *OpenFaceChineseWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config OpenFaceChineseWebOutputConfig `json:"config"`
}

// OpenFaceChineseWebOutputConfig オープンフェイス・チャイニーズポーカー (OFC) の設定アウトプット
type OpenFaceChineseWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PlayerCount   int `json:"playerCount"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds an OpenFaceChineseConfig from the nested web config, applying bounds checking.
func (c *OpenFaceChineseWebConfig) ToConfig() domain.OpenFaceChineseConfig {
	cfg := domain.DefaultOpenFaceChineseConfig()
	cfg.CpuDifficulty = domain.OpenFaceChineseCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.OpenFaceChineseCpuDifficultyEasy), int(domain.OpenFaceChineseCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.OpenFaceChinesePlayerMin, domain.OpenFaceChinesePlayerMax)
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 1000000)
	return cfg
}

// ToConfig builds an OpenFaceChineseConfig from the web input.
func (p OpenFaceChineseWebInput) ToConfig() domain.OpenFaceChineseConfig {
	return configOrDefault(p.Config, (*OpenFaceChineseWebConfig).ToConfig, domain.DefaultOpenFaceChineseConfig())
}

// OpenFaceChineseWebController オープンフェイス・チャイニーズポーカー (OFC) のWebコントローラークラス
type OpenFaceChineseWebController = GameWebController[usecase.OpenFaceChineseInteractorIF, OpenFaceChineseWebInput, *OpenFaceChineseWebOutput]

// NewOpenFaceChineseWebController and NewOpenFaceChineseWebControllerWithProvider are
// the standard and provider-backed constructors for OpenFaceChineseWebController.
var NewOpenFaceChineseWebController, NewOpenFaceChineseWebControllerWithProvider = webControllerPair[usecase.OpenFaceChineseInteractorIF, OpenFaceChineseWebInput, *OpenFaceChineseWebOutput](
	newOpenFaceChineseDefaultOutput, openFaceChineseDispatch,
)

func newOpenFaceChineseDefaultOutput(msg string) *OpenFaceChineseWebOutput {
	return &OpenFaceChineseWebOutput{
		Players:       make([]*OpenFaceChineseWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func openFaceChineseDispatch(bc *baseController, w http.ResponseWriter, di usecase.OpenFaceChineseInteractorIF, param OpenFaceChineseWebInput, newDefault func(string) *OpenFaceChineseWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "place":
		if !requireParam(bc, w, newDefault, param.Row == nil, "param error: row is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Place(*param.Row))
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
