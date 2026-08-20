//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrexWebInput トリックスWebインプット
type TrexWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// Contract は王が選ぶ契約 (0=♥K, 1=♦, 2=Q, 3=トリック, 4=ドミノ)。
	Contract *int           `json:"contract,omitempty"`
	Config   *TrexWebConfig `json:"config,omitempty"`
}

// TrexWebConfig トリックスWeb設定
type TrexWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// TrexWebOutputPlayer トリックスWebアウトプットプレイヤー
type TrexWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Score は累計得点。ドミノで加点されるので正にもなる。
	Score int `json:"score"`
	// DealScore は今ディールの得点。
	DealScore int `json:"dealScore"`
	// TricksWon は今ディールのトリック数。
	TricksWon int  `json:"tricksWon"`
	Hidden    bool `json:"hidden"`
}

// TrexWebOutputTrickCard トリックに出された1枚
type TrexWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// TrexWebOutputRun ドミノの1スートの伸び
type TrexWebOutputRun struct {
	Suit    int  `json:"suit"`
	Started bool `json:"started"`
	// Low/High は場に出ている範囲 (A=14)。J=11 が起点。
	Low  int `json:"low"`
	High int `json:"high"`
}

// TrexWebOutputHint ヒント出力
type TrexWebOutputHint struct {
	CardIndex *int `json:"cardIndex,omitempty"`
	// Contract は選択フェーズで推奨する契約。
	Contract *int `json:"contract,omitempty"`
	// Pass が真ならドミノでパスするのが推奨手。
	Pass   bool   `json:"pass"`
	Reason string `json:"reason"`
}

// TrexWebOutput トリックスWebアウトプット
type TrexWebOutput struct {
	Players          []*TrexWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	KingIdx          int                    `json:"kingIdx"`
	// Contract は現在の契約 (5=未選択)。
	Contract int `json:"contract"`
	// AvailableContracts は王がまだ選んでいない契約。同じ契約は 1 王国に 1 度だけ。
	AvailableContracts []int `json:"availableContracts"`
	// IsTrix が真ならドミノ契約。トリックではなく列を伸ばす。
	IsTrix bool `json:"isTrix"`
	// DealNo は完了したディール数。TotalDeals に達すると終局。
	DealNo     int `json:"dealNo"`
	TotalDeals int `json:"totalDeals"`
	// Trick は現在のトリック (ドミノでは空)。
	Trick   []*TrexWebOutputTrickCard `json:"trick"`
	TrickNo int                       `json:"trickNo"`
	// Runs はドミノの4列。J=11 起点で上下に伸びる。
	Runs []*TrexWebOutputRun `json:"runs"`
	// FinishOrder はドミノの上がり順。
	FinishOrder []int `json:"finishOrder"`
	// ValidIndices は出せる手札の添字。契約ごとに違う規則をクライアントに
	// 再実装させない。
	ValidIndices []int `json:"validIndices"`
	// CanPass が真ならドミノでパスできる (出せる札が 1 枚も無いとき)。
	CanPass     bool               `json:"canPass"`
	GameEndFlag bool               `json:"gameEndFlag"`
	WinnerIdx   int                `json:"winnerIdx"`
	Hint        *TrexWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config TrexWebOutputConfig `json:"config"`
}

// TrexWebOutputConfig トリックス設定アウトプット
type TrexWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a TrexConfig from the nested web config, applying bounds checking.
func (c *TrexWebConfig) ToConfig() domain.TrexConfig {
	cfg := domain.DefaultTrexConfig()
	cfg.CpuDifficulty = domain.TrexCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.TrexCpuDifficultyNormal), int(domain.TrexCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a TrexConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *TrexWebConfig and calling the method on it would
// dereference nil.
func (i TrexWebInput) ToConfig() domain.TrexConfig {
	return configOrDefault(i.Config, (*TrexWebConfig).ToConfig, domain.DefaultTrexConfig())
}

// TrexWebController トリックスWebコントローラ
type TrexWebController = GameWebController[usecase.TrexInteractorIF, TrexWebInput, *TrexWebOutput]

// NewTrexWebController and NewTrexWebControllerWithProvider are the standard
// and provider-backed constructors for TrexWebController.
var NewTrexWebController, NewTrexWebControllerWithProvider = webControllerPair[usecase.TrexInteractorIF, TrexWebInput, *TrexWebOutput](
	newTrexDefaultOutput, trexDispatch,
)

func newTrexDefaultOutput(msg string) *TrexWebOutput {
	return &TrexWebOutput{
		Players:            make([]*TrexWebOutputPlayer, 0),
		AvailableContracts: make([]int, 0),
		Trick:              make([]*TrexWebOutputTrickCard, 0),
		Runs:               make([]*TrexWebOutputRun, 0),
		FinishOrder:        make([]int, 0),
		ValidIndices:       make([]int, 0),
		Contract:           int(domain.TrexContractNone),
		TotalDeals:         domain.TrexTotalDeals,
		WinnerIdx:          -1,
		WebOutputBase:      WebOutputBase{Message: msg},
	}
}

func trexDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TrexInteractorIF, param TrexWebInput, newDefault func(string) *TrexWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "c", "choose":
		if !requireParam(bc, w, newDefault, param.Contract == nil, "param error: contract is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Choose(*param.Contract))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "s", "pass":
		bc.writePresenterResponse(w, ti.Pass())
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextDeal())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
