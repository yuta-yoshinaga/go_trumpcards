package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TonkWebInput Tonk Webインプット
type TonkWebInput struct {
	BaseWebInput
	CardIndex *int           `json:"cardIndex,omitempty"`
	Config    *TonkWebConfig `json:"config,omitempty"`
}

// TonkWebConfig Tonk Web設定
type TonkWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// TonkWebOutputPlayer Tonk Webアウトプットプレイヤー
type TonkWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
}

// TonkWebOutputMeld メルドのアウトプット
type TonkWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// TonkWebOutput Tonk Webアウトプット
type TonkWebOutput struct {
	Players          []*TonkWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	RoundNumber      int                    `json:"roundNumber"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard         `json:"discardTop"`
	DrawPileCount    int                    `json:"drawPileCount"`
	GameEndFlag      bool                   `json:"gameEndFlag"`
	WinnerIdx        int                    `json:"winnerIdx"`
	KnockerIdx       int                    `json:"knockerIdx"`
	KnockerMelds     []*TonkWebOutputMeld   `json:"knockerMelds"`
	KnockerDeadwood  []*WebOutputCard       `json:"knockerDeadwood"`
	OpponentMelds    []*TonkWebOutputMeld   `json:"opponentMelds"`
	OpponentDeadwood []*WebOutputCard       `json:"opponentDeadwood"`
	IsTonk           bool                   `json:"isTonk"`
	IsUndercut       bool                   `json:"isUndercut"`
	// BestDeadwood は1枚捨てて到達できる最小デッドウッド。CUI は毎ターン
	// これを閾値と比べて「ノック可能/不可」を出しているのに、Web は同じ判断を
	// プレイヤーの手計算に任せていた。人間のディスカードフェーズ以外は -1。
	BestDeadwood int `json:"bestDeadwood"`
	// KnockThreshold はノックできるデッドウッド上限 (domain.TonkKnockThreshold)。
	// フロントに数値を写さず、判断の基準ごと送る。
	KnockThreshold int `json:"knockThreshold"`
	WebOutputBase
	Config TonkWebOutputConfig `json:"config"`
}

// TonkWebOutputConfig Tonk設定アウトプット
type TonkWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a TonkConfig from the nested web config, applying bounds checking.
func (c *TonkWebConfig) ToConfig() domain.TonkConfig {
	cfg := domain.DefaultTonkConfig()
	cfg.CpuDifficulty = domain.TonkCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TonkCpuDifficultyEasy), int(domain.TonkCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a TonkConfig from the web input.
func (p TonkWebInput) ToConfig() domain.TonkConfig {
	return configOrDefault(p.Config, (*TonkWebConfig).ToConfig, domain.DefaultTonkConfig())
}

// TonkWebController Tonk Webコントローラークラス
type TonkWebController = GameWebController[usecase.TonkInteractorIF, TonkWebInput, *TonkWebOutput]

// NewTonkWebController and NewTonkWebControllerWithProvider are
// the standard and provider-backed constructors for TonkWebController.
var NewTonkWebController, NewTonkWebControllerWithProvider = webControllerPair[usecase.TonkInteractorIF, TonkWebInput, *TonkWebOutput](
	newTonkDefaultOutput, tonkDispatch,
)

func newTonkDefaultOutput(msg string) *TonkWebOutput {
	return &TonkWebOutput{
		Players:          make([]*TonkWebOutputPlayer, 0),
		WinnerIdx:        -1,
		KnockerIdx:       -1,
		KnockerMelds:     make([]*TonkWebOutputMeld, 0),
		KnockerDeadwood:  make([]*WebOutputCard, 0),
		OpponentMelds:    make([]*TonkWebOutputMeld, 0),
		OpponentDeadwood: make([]*WebOutputCard, 0),
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func tonkDispatch(bc *baseController, w http.ResponseWriter, ci usecase.TonkInteractorIF, param TonkWebInput, newDefault func(string) *TonkWebOutput) bool {
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
