//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CostlyColoursWebInput はコストリー・カラーズの Web インプット。
type CostlyColoursWebInput struct {
	BaseWebInput
	// HandIndex は出す手札のインデックス。
	HandIndex *int `json:"handIndex,omitempty"`
	// Accept は交換に応じるか (mog コマンドで使う)。
	Accept *bool `json:"accept,omitempty"`
	// Config はゲーム設定。
	Config *CostlyColoursWebConfig `json:"config,omitempty"`
}

// CostlyColoursWebConfig はコストリー・カラーズの Web 設定。
type CostlyColoursWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// CostlyColoursWebOutputPlayer は 1 席ぶんの出力。
type CostlyColoursWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Cards は手札 (人間のみ公開)。
	Cards     []*WebOutputCard `json:"cards"`
	CardCount int              `json:"cardCount"`
	// Played はこのディールで出した札 (ショーで数える 3 枚の一部)。
	Played   []*WebOutputCard `json:"played"`
	Score    int              `json:"score"`
	IsDealer bool             `json:"isDealer"`
	MoggedIn bool             `json:"moggedIn"`
}

// CostlyColoursWebOutputScoreLine は 1 つの得点項目。
type CostlyColoursWebOutputScoreLine struct {
	Key    string `json:"key"`
	Points []int  `json:"points"`
}

// CostlyColoursWebOutputResult は 1 ディールの集計結果。
type CostlyColoursWebOutputResult struct {
	Lines  []*CostlyColoursWebOutputScoreLine `json:"lines"`
	Totals []int                              `json:"totals"`
	// Combos は席ごとの色とスートの役の識別子。
	Combos []string `json:"combos"`
}

// CostlyColoursWebOutput はコストリー・カラーズの Web アウトプット。
type CostlyColoursWebOutput struct {
	Players    []*CostlyColoursWebOutputPlayer `json:"players"`
	Phase      string                          `json:"phase"`
	DealNumber int                             `json:"dealNumber"`
	DealerIdx  int                             `json:"dealerIdx"`
	// CurrentPlayerIdx は手番の席。
	CurrentPlayerIdx int `json:"currentPlayerIdx"`
	// TurnUp は表に返した 1 枚。**ショーでも数える。**
	TurnUp *WebOutputCard `json:"turnUp,omitempty"`
	// Pile は今の数え上げに出た札。
	Pile []*WebOutputCard `json:"pile"`
	// Total は今の数え上げの累計。**31 を超えられない。**
	Total int `json:"total"`
	// WentOut は「ゴー」を宣言した席 (-1 = なし)。
	WentOut int `json:"wentOut"`
	// PlayableIdxs は人間が出せる手札の位置。
	PlayableIdxs []int                         `json:"playableIdxs"`
	LastResult   *CostlyColoursWebOutputResult `json:"lastResult,omitempty"`
	GameEndFlag  bool                          `json:"gameEndFlag"`
	WinnerIdx    int                           `json:"winnerIdx"`
	IsHumanTurn  bool                          `json:"isHumanTurn"`
	// HintHandIdx は勧める手札 (-1 = なし)。
	HintHandIdx int `json:"hintHandIdx"`
	// HintAcceptMog は交換に応じるべきか。
	HintAcceptMog bool `json:"hintAcceptMog"`
	// HintReason は理由の識別子。
	HintReason string `json:"hintReason"`
	WebOutputBase
	Config CostlyColoursWebOutputConfig `json:"config"`
}

// CostlyColoursWebOutputConfig は設定アウトプット。
type CostlyColoursWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *CostlyColoursWebConfig) ToConfig() domain.CostlyColoursConfig {
	cfg := domain.DefaultCostlyColoursConfig()
	cfg.CpuDifficulty = domain.CostlyColoursCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.CostlyColoursCpuDifficultyEasy),
		int(domain.CostlyColoursCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore,
		domain.CostlyColoursMinTarget, domain.CostlyColoursMaxTarget)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p CostlyColoursWebInput) ToConfig() domain.CostlyColoursConfig {
	return configOrDefault(p.Config, (*CostlyColoursWebConfig).ToConfig,
		domain.DefaultCostlyColoursConfig())
}

// CostlyColoursWebController はコストリー・カラーズの Web コントローラー。
type CostlyColoursWebController = GameWebController[usecase.CostlyColoursInteractorIF, CostlyColoursWebInput, *CostlyColoursWebOutput]

// NewCostlyColoursWebController, NewCostlyColoursWebControllerWithProvider are the
// standard and provider-backed constructors.
var NewCostlyColoursWebController, NewCostlyColoursWebControllerWithProvider = webControllerPair[usecase.CostlyColoursInteractorIF, CostlyColoursWebInput, *CostlyColoursWebOutput](
	newCostlyColoursDefaultOutput, costlyColoursDispatch,
)

func newCostlyColoursDefaultOutput(msg string) *CostlyColoursWebOutput {
	return &CostlyColoursWebOutput{
		Players:       make([]*CostlyColoursWebOutputPlayer, 0),
		Pile:          make([]*WebOutputCard, 0),
		PlayableIdxs:  make([]int, 0),
		WentOut:       -1,
		WinnerIdx:     -1,
		HintHandIdx:   -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func costlyColoursDispatch(bc *baseController, w http.ResponseWriter, di usecase.CostlyColoursInteractorIF, param CostlyColoursWebInput, newDefault func(string) *CostlyColoursWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "mog":
		// **応じるか断るかは要求に乗せる。** 既定を「応じる」にすると、
		// 断るつもりの要求が黙って交換になる。
		if !requireParam(bc, w, newDefault, param.Accept == nil, "param error: accept is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Mog(*param.Accept))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.HandIndex == nil, "param error: handIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.HandIndex))
	case "nd", "nextdeal":
		bc.writePresenterResponse(w, di.NextDeal())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
