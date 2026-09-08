//go:build !js || !wasm || extra5

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TongitsWebInput Tongits Webインプット
type TongitsWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// Indices は meld で公開する手札の位置。sapaw ではなく meld 専用。
	Indices []int `json:"indices,omitempty"`
	// TargetPlayerIdx / MeldIdx は sapaw の宛先。**他家のメルドも指せる**ので
	// プレイヤー番号が要る -- ここが Tonk のノックと決定的に違う点。
	TargetPlayerIdx *int `json:"targetPlayerIdx,omitempty"`
	MeldIdx         *int `json:"meldIdx,omitempty"`
	// Agreed は challenge に他家が応じたか。要素数はプレイヤー数-1。
	Agreed []bool            `json:"agreed,omitempty"`
	Config *TongitsWebConfig `json:"config,omitempty"`
}

// TongitsWebConfig Tongits Web設定
type TongitsWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// TongitsWebOutputPlayer Tongits Webアウトプットプレイヤー
type TongitsWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Melds はそのプレイヤーが場に公開しているメルド。**Tongits のメルドは
	// 出した時点で全員に見える**ので、手札と違って伏せない。sapaw の宛先を
	// 選ぶにもこれが要る。
	Melds           []*TongitsWebOutputMeld `json:"melds"`
	RoundScore      int                     `json:"roundScore"`
	CumulativeScore int                     `json:"cumulativeScore"`
}

// TongitsWebOutputMeld メルドのアウトプット
type TongitsWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// TongitsWebOutput Tongits Webアウトプット
type TongitsWebOutput struct {
	Players          []*TongitsWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard            `json:"discardTop"`
	DrawPileCount    int                       `json:"drawPileCount"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerIdx        int                       `json:"winnerIdx"`
	IsTongits        bool                      `json:"isTongits"`
	// RemainingPoints は手番のプレイヤーの残り点 (challenge の判定材料)。
	// 手番でないときは -1。
	RemainingPoints int `json:"remainingPoints"`
	// これを閾値と比べて「ノック可能/不可」を出しているのに、Web は同じ判断を
	// プレイヤーの手計算に任せていた。人間のディスカードフェーズ以外は -1。
	// フロントに数値を写さず、判断の基準ごと送る。
	WebOutputBase
	Config TongitsWebOutputConfig `json:"config"`
}

// TongitsWebOutputConfig Tongits設定アウトプット
type TongitsWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a TongitsConfig from the nested web config, applying bounds checking.
func (c *TongitsWebConfig) ToConfig() domain.TongitsConfig {
	cfg := domain.DefaultTongitsConfig()
	cfg.CpuDifficulty = domain.TongitsCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TongitsCpuDifficultyEasy), int(domain.TongitsCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a TongitsConfig from the web input.
func (p TongitsWebInput) ToConfig() domain.TongitsConfig {
	return configOrDefault(p.Config, (*TongitsWebConfig).ToConfig, domain.DefaultTongitsConfig())
}

// TongitsWebController Tongits Webコントローラークラス
type TongitsWebController = GameWebController[usecase.TongitsInteractorIF, TongitsWebInput, *TongitsWebOutput]

// NewTongitsWebController and NewTongitsWebControllerWithProvider are
// the standard and provider-backed constructors for TongitsWebController.
var NewTongitsWebController, NewTongitsWebControllerWithProvider = webControllerPair[usecase.TongitsInteractorIF, TongitsWebInput, *TongitsWebOutput](
	newTongitsDefaultOutput, tongitsDispatch,
)

func newTongitsDefaultOutput(msg string) *TongitsWebOutput {
	return &TongitsWebOutput{
		Players:         make([]*TongitsWebOutputPlayer, 0),
		WinnerIdx:       -1,
		RemainingPoints: -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func tongitsDispatch(bc *baseController, w http.ResponseWriter, ci usecase.TongitsInteractorIF, param TongitsWebInput, newDefault func(string) *TongitsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Discard(*param.CardIndex))
	case "m", "meld":
		if !requireParam(bc, w, newDefault, len(param.Indices) == 0, "param error: indices is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Meld(param.Indices))
	case "sp", "sapaw":
		if !requireParam(bc, w, newDefault,
			param.TargetPlayerIdx == nil || param.MeldIdx == nil || param.CardIndex == nil,
			"param error: targetPlayerIdx, meldIdx and cardIndex are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Sapaw(*param.TargetPlayerIdx, *param.MeldIdx, *param.CardIndex))
	case "c", "challenge":
		if !requireParam(bc, w, newDefault, len(param.Agreed) == 0, "param error: agreed is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Challenge(param.Agreed))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
