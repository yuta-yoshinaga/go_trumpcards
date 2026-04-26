package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShitheadWebConfig ローカルルール設定 (入力・出力共用)
type ShitheadWebConfig struct {
	MagicTwo        bool `json:"magicTwo"`
	MagicSeven      bool `json:"magicSeven"`
	MagicEight      bool `json:"magicEight"`
	MagicTen        bool `json:"magicTen"`
	FourOfAKindBurn bool `json:"fourOfAKindBurn"`
	CpuDifficulty   int  `json:"cpuDifficulty"`
}

// ToConfig converts ShitheadWebConfig to domain.ShitheadConfig.
func (c ShitheadWebConfig) ToConfig() domain.ShitheadConfig {
	return domain.ShitheadConfig{
		MagicTwo:        c.MagicTwo,
		MagicSeven:      c.MagicSeven,
		MagicEight:      c.MagicEight,
		MagicTen:        c.MagicTen,
		FourOfAKindBurn: c.FourOfAKindBurn,
		CpuDifficulty:   domain.ShitheadCpuDifficulty(c.CpuDifficulty),
	}
}

// ShitheadWebInput シットヘッドWebインプット
type ShitheadWebInput struct {
	BaseWebInput
	Indices []int              `json:"indices"` // 出すカードのインデックス。play コマンド用。空はピックアップ。
	Config  *ShitheadWebConfig `json:"config"`  // リセット時のローカルルール設定 (省略可)
}

// ShitheadWebOutputPlayer シットヘッドWebアウトプットプレイヤー
type ShitheadWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	IsFinished    bool             `json:"isFinished"`
	Rank          int              `json:"rank"`
	HandCount     int              `json:"handCount"`
	HandCards     []*WebOutputCard `json:"handCards"`
	FaceUpCards   []*WebOutputCard `json:"faceUpCards"`
	FaceDownCount int              `json:"faceDownCount"`
}

// ShitheadWebOutputAction プレイヤー行動記録
type ShitheadWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	Source      string           `json:"source"`
	PlayedCards []*WebOutputCard `json:"playedCards"` // pickup なら空
	Pickup      bool             `json:"pickup"`
	Burned      bool             `json:"burned"`
	Skipped     bool             `json:"skipped"`
}

// ShitheadWebOutput シットヘッドWebアウトプット
type ShitheadWebOutput struct {
	Players       []*ShitheadWebOutputPlayer `json:"players"`
	CurrentTurn   int                        `json:"currentTurn"`
	CurrentSource string                     `json:"currentSource"`
	DiscardPile   []*WebOutputCard           `json:"discardPile"`
	StockSize     int                        `json:"stockSize"`
	SkipNext      bool                       `json:"skipNext"`
	SevenActive   bool                       `json:"sevenActive"`
	GameEndFlag   bool                       `json:"gameEndFlag"`
	Config        ShitheadWebConfig          `json:"config"`
	CpuActions    []*ShitheadWebOutputAction `json:"cpuActions"`
	HumanAction   *ShitheadWebOutputAction   `json:"humanAction"`
	WebOutputBase
}

// ShitheadWebController シットヘッドWebコントローラークラス
type ShitheadWebController = GameWebController[usecase.ShitheadInteractorIF, ShitheadWebInput, *ShitheadWebOutput]

// NewShitheadWebController, NewShitheadWebControllerWithProvider are the
// standard and provider-backed constructors for ShitheadWebController.
var NewShitheadWebController, NewShitheadWebControllerWithProvider = webControllerPair[usecase.ShitheadInteractorIF, ShitheadWebInput, *ShitheadWebOutput](
	newShitheadDefaultOutput, shitheadDispatch,
)

func newShitheadDefaultOutput(msg string) *ShitheadWebOutput {
	return &ShitheadWebOutput{
		Players:       make([]*ShitheadWebOutputPlayer, 0),
		DiscardPile:   make([]*WebOutputCard, 0),
		CpuActions:    make([]*ShitheadWebOutputAction, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func shitheadDispatch(bc *baseController, w http.ResponseWriter, si usecase.ShitheadInteractorIF, param ShitheadWebInput, _ func(string) *ShitheadWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, si.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, si.Play(indices))
	default:
		return dispatchLog(param.Command, bc, w, si.ActionLog)
	}
	return true
}
