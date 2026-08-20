//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RussianBankWebInput ロシアンバンク (クラペット) のWebインプット。
type RussianBankWebInput struct {
	BaseWebInput
	// Zone 移動元ゾーン (0=reserve,1=waste,2=tableau)。
	Zone *int `json:"zone,omitempty"`
	// FromOpp 移動元が相手の山か。
	FromOpp *bool `json:"fromOpp,omitempty"`
	// Col 移動元タブロー列 (Zone=tableau のとき)。
	Col *int `json:"col,omitempty"`
	// ToCol 移動先タブロー列 (mt のとき)。
	ToCol *int `json:"toCol,omitempty"`
	// Config ゲーム設定。
	Config *RussianBankWebConfig `json:"config,omitempty"`
}

// RussianBankWebConfig ロシアンバンク (クラペット) のWeb設定。
type RussianBankWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// RussianBankWebOutputPlayer ロシアンバンク (クラペット) のWebアウトプットプレイヤー。
type RussianBankWebOutputPlayer struct {
	ID           int            `json:"id"`
	IsHuman      bool           `json:"isHuman"`
	ReserveCount int            `json:"reserveCount"`
	ReserveTop   *WebOutputCard `json:"reserveTop,omitempty"`
	HandCount    int            `json:"handCount"`
	WasteCount   int            `json:"wasteCount"`
	WasteTop     *WebOutputCard `json:"wasteTop,omitempty"`
	StopPoints   int            `json:"stopPoints"`
}

// RussianBankWebOutputHint ヒント出力。
type RussianBankWebOutputHint struct {
	Zone         int  `json:"zone"`
	FromOpponent bool `json:"fromOpponent"`
	Col          int  `json:"col"`
	ToFoundation bool `json:"toFoundation"`
	ToCol        int  `json:"toCol"`
}

// RussianBankWebOutput ロシアンバンク (クラペット) のWebアウトプット。
type RussianBankWebOutput struct {
	Players          []*RussianBankWebOutputPlayer `json:"players"`
	Tableau          [][]*WebOutputCard            `json:"tableau"`
	Foundations      [][]*WebOutputCard            `json:"foundations"`
	Phase            int                           `json:"phase"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerIdx        int                           `json:"winnerIdx"`
	IsHumanTurn      bool                          `json:"isHumanTurn"`
	CanCallStop      bool                          `json:"canCallStop"`
	CanUndo          bool                          `json:"canUndo"`
	MoveCount        int                           `json:"moveCount"`
	Hint             *RussianBankWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config RussianBankWebOutputConfig `json:"config"`
}

// RussianBankWebOutputConfig ロシアンバンク (クラペット) の設定アウトプット。
type RussianBankWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a RussianBankConfig from the nested web config, applying bounds checking.
func (c *RussianBankWebConfig) ToConfig() domain.RussianBankConfig {
	cfg := domain.DefaultRussianBankConfig()
	cfg.CpuDifficulty = domain.RussianBankCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.RussianBankCpuDifficultyEasy), int(domain.RussianBankCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a RussianBankConfig from the web input.
func (p RussianBankWebInput) ToConfig() domain.RussianBankConfig {
	return configOrDefault(p.Config, (*RussianBankWebConfig).ToConfig, domain.DefaultRussianBankConfig())
}

// RussianBankWebController ロシアンバンク (クラペット) のWebコントローラークラス。
type RussianBankWebController = GameWebController[usecase.RussianBankInteractorIF, RussianBankWebInput, *RussianBankWebOutput]

// NewRussianBankWebController and NewRussianBankWebControllerWithProvider are the
// standard and provider-backed constructors for RussianBankWebController.
var NewRussianBankWebController, NewRussianBankWebControllerWithProvider = webControllerPair[usecase.RussianBankInteractorIF, RussianBankWebInput, *RussianBankWebOutput](
	newRussianBankDefaultOutput, russianBankDispatch,
)

func newRussianBankDefaultOutput(msg string) *RussianBankWebOutput {
	return &RussianBankWebOutput{
		Players:       make([]*RussianBankWebOutputPlayer, 0),
		Tableau:       make([][]*WebOutputCard, 0),
		Foundations:   make([][]*WebOutputCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func russianBankDispatch(bc *baseController, w http.ResponseWriter, di usecase.RussianBankInteractorIF, param RussianBankWebInput, newDefault func(string) *RussianBankWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "pf":
		if !requireParam(bc, w, newDefault, param.Zone == nil, "param error: zone is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveToFoundation(*param.Zone, derefBool(param.FromOpp), derefInt(param.Col)))
	case "mt":
		if !requireParam(bc, w, newDefault, param.Zone == nil || param.ToCol == nil, "param error: zone and toCol are required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveToTableau(*param.Zone, derefBool(param.FromOpp), derefInt(param.Col), *param.ToCol))
	case "d", "discard":
		bc.writePresenterResponse(w, di.Discard())
	case "s", "stop":
		bc.writePresenterResponse(w, di.CallStop())
	case "u", "undo":
		bc.writePresenterResponse(w, di.Undo())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}

// derefInt は *int を 0 既定で参照外しする。
func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// derefBool は *bool を false 既定で参照外しする。
func derefBool(p *bool) bool {
	return p != nil && *p
}
