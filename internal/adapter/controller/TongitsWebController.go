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
	CardIndex *int              `json:"cardIndex,omitempty"`
	Config    *TongitsWebConfig `json:"config,omitempty"`
}

// TongitsWebConfig Tongits Web設定
type TongitsWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// TongitsWebOutputPlayer Tongits Webアウトプットプレイヤー
type TongitsWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
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
	KnockerIdx       int                       `json:"knockerIdx"`
	KnockerMelds     []*TongitsWebOutputMeld   `json:"knockerMelds"`
	KnockerDeadwood  []*WebOutputCard          `json:"knockerDeadwood"`
	OpponentMelds    []*TongitsWebOutputMeld   `json:"opponentMelds"`
	OpponentDeadwood []*WebOutputCard          `json:"opponentDeadwood"`
	IsTongits        bool                      `json:"isTongits"`
	// UndercutRiskMax は「アンダーカットされうる」と警告する相手の残り枚数 (#5582)。
	// 閾値を画面に焼き込むと、変えたとき Web と CUI で警告の出る局面がずれる。
	UndercutRiskMax int  `json:"undercutRiskMax"`
	IsUndercut      bool `json:"isUndercut"`
	// BestDeadwood は1枚捨てて到達できる最小デッドウッド。CUI は毎ターン
	// これを閾値と比べて「ノック可能/不可」を出しているのに、Web は同じ判断を
	// プレイヤーの手計算に任せていた。人間のディスカードフェーズ以外は -1。
	BestDeadwood int `json:"bestDeadwood"`
	// KnockThreshold はノックできるデッドウッド上限 (domain.TongitsKnockThreshold)。
	// フロントに数値を写さず、判断の基準ごと送る。
	KnockThreshold int `json:"knockThreshold"`
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
		Players:          make([]*TongitsWebOutputPlayer, 0),
		WinnerIdx:        -1,
		KnockerIdx:       -1,
		KnockerMelds:     make([]*TongitsWebOutputMeld, 0),
		KnockerDeadwood:  make([]*WebOutputCard, 0),
		OpponentMelds:    make([]*TongitsWebOutputMeld, 0),
		OpponentDeadwood: make([]*WebOutputCard, 0),
		// 閾値は盤面が無くても規則なので、既定の応答にも乗せる。
		UndercutRiskMax: domain.TongitsUndercutRiskMax,
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
	case "k", "knock":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Knock(*param.CardIndex))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
