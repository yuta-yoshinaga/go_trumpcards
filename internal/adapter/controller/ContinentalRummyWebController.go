//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ContinentalRummyWebInput はコンチネンタル・ラミーの Web インプット。
type ContinentalRummyWebInput struct {
	BaseWebInput
	// HandIndex は捨てる / 上がるときに手放す手札の位置。
	HandIndex *int `json:"handIndex,omitempty"`
	// Config はゲーム設定。
	Config *ContinentalRummyWebConfig `json:"config,omitempty"`
}

// ContinentalRummyWebConfig はコンチネンタル・ラミーの Web 設定。
type ContinentalRummyWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TotalRounds   *int `json:"totalRounds,omitempty"`
}

// ContinentalRummyWebOutputPlayer は 1 席ぶんの出力。
type ContinentalRummyWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Cards は手札。**自分の席だけ中身が入る。**
	Cards []*WebOutputCard `json:"cards"`
	// CardCount は手札の枚数。相手の席はこれだけ。
	CardCount int `json:"cardCount"`
	// Melds は上がったときに並べたシーケンス。上がっていなければ空。
	Melds [][]*WebOutputCard `json:"melds"`
	Score int                `json:"score"`
	// IsDealer は親か。
	IsDealer bool `json:"isDealer"`
}

// ContinentalRummyWebOutputBonus は加点の内訳 1 行。
type ContinentalRummyWebOutputBonus struct {
	Key    string `json:"key"`
	Points int    `json:"points"`
}

// ContinentalRummyWebOutputResult は 1 ラウンドの決着。
type ContinentalRummyWebOutputResult struct {
	// WinnerIdx は上がった席。誰も上がらず山が尽きたら -1。
	WinnerIdx int                              `json:"winnerIdx"`
	Bonuses   []ContinentalRummyWebOutputBonus `json:"bonuses"`
	// PerOpponent は相手 1 人あたりの取り立て額。
	PerOpponent int `json:"perOpponent"`
	// Total は上がった側が受け取った合計。
	Total int `json:"total"`
}

// ContinentalRummyWebOutput はコンチネンタル・ラミーの Web アウトプット。
type ContinentalRummyWebOutput struct {
	Players []*ContinentalRummyWebOutputPlayer `json:"players"`
	// Phase は "draw" | "discard" | "roundEnd" | "gameEnd"。
	Phase            string                           `json:"phase"`
	RoundNumber      int                              `json:"roundNumber"`
	TotalRounds      int                              `json:"totalRounds"`
	CurrentPlayerIdx int                              `json:"currentPlayerIdx"`
	DealerIdx        int                              `json:"dealerIdx"`
	StockCount       int                              `json:"stockCount"`
	DiscardTop       *WebOutputCard                   `json:"discardTop,omitempty"`
	Layouts          [][]int                          `json:"layouts"`
	LastResult       *ContinentalRummyWebOutputResult `json:"lastResult,omitempty"`
	GameEndFlag      bool                             `json:"gameEndFlag"`
	WinnerIdx        int                              `json:"winnerIdx"`
	IsHumanTurn      bool                             `json:"isHumanTurn"`
	// GoOutIdx は「これを捨てれば上がれる」1 枚。上がれないなら -1。
	//
	// **上がれるかはページ側で解き直さない。** 15 枚の分割問題なので、規則が
	// 2 か所に増えるとどこかで食い違う。
	GoOutIdx int `json:"goOutIdx"`
	// CanGoOutOnDeal は引かずに、配られた 15 枚のまま上がれるか。
	//
	// **こちらは札を捨てない上がり。** 引いたあとの上がりとは加点が違う
	// (10 点 vs 7 点) ので、別の入口として出す。
	CanGoOutOnDeal bool `json:"canGoOutOnDeal"`
	// HintDiscardIdx は捨てるとよい手札の位置。無ければ -1。
	HintDiscardIdx int `json:"hintDiscardIdx"`
	// HintReason は理由の識別子。
	HintReason string `json:"hintReason"`
	WebOutputBase
	Config ContinentalRummyWebOutputConfig `json:"config"`
}

// ContinentalRummyWebOutputConfig は設定アウトプット。
type ContinentalRummyWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TotalRounds   int `json:"totalRounds"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *ContinentalRummyWebConfig) ToConfig() domain.ContinentalRummyConfig {
	cfg := domain.DefaultContinentalRummyConfig()
	cfg.CpuDifficulty = domain.ContinentalRummyCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.ContinentalRummyCpuDifficultyEasy),
		int(domain.ContinentalRummyCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TotalRounds, c.TotalRounds,
		domain.ContinentalRummyMinRounds, domain.ContinentalRummyMaxRounds)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p ContinentalRummyWebInput) ToConfig() domain.ContinentalRummyConfig {
	return configOrDefault(p.Config, (*ContinentalRummyWebConfig).ToConfig,
		domain.DefaultContinentalRummyConfig())
}

// ContinentalRummyWebController はコンチネンタル・ラミーの Web コントローラー。
type ContinentalRummyWebController = GameWebController[usecase.ContinentalRummyInteractorIF, ContinentalRummyWebInput, *ContinentalRummyWebOutput]

// NewContinentalRummyWebController, NewContinentalRummyWebControllerWithProvider are
// the standard and provider-backed constructors.
var NewContinentalRummyWebController, NewContinentalRummyWebControllerWithProvider = webControllerPair[usecase.ContinentalRummyInteractorIF, ContinentalRummyWebInput, *ContinentalRummyWebOutput](
	newContinentalRummyDefaultOutput, continentalRummyDispatch,
)

func newContinentalRummyDefaultOutput(msg string) *ContinentalRummyWebOutput {
	return &ContinentalRummyWebOutput{
		Players:        make([]*ContinentalRummyWebOutputPlayer, 0),
		Layouts:        domain.ContinentalRummyLayouts(),
		WinnerIdx:      -1,
		GoOutIdx:       -1,
		HintDiscardIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func continentalRummyDispatch(bc *baseController, w http.ResponseWriter, di usecase.ContinentalRummyInteractorIF, param ContinentalRummyWebInput, def func(string) *ContinentalRummyWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	// **山と捨て札は別のコマンド。** 真偽値に畳むと、付け忘れた要求が
	// 黙ってどちらかに倒れる。
	case "ds", "stock":
		bc.writePresenterResponse(w, di.DrawStock())
	case "dd", "take":
		bc.writePresenterResponse(w, di.DrawDiscard())
	case "d", "discard":
		if !requireParam(bc, w, def, param.HandIndex == nil, "param error: handIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(*param.HandIndex))
	// **上がるときも捨てる 1 枚を名指す。** 15 枚を並べて 1 枚を手放すので、
	// どれを手放すかを決めずには上がれない。
	case "g", "goout":
		if !requireParam(bc, w, def, param.HandIndex == nil, "param error: handIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.GoOut(*param.HandIndex))
	// **引かずに上がるほうは札を捨てない。** 加点が違う (10 点 vs 7 点) ので
	// 別の命令にして、捨てる札を求めない。
	case "gd", "gooutdeal":
		bc.writePresenterResponse(w, di.GoOut(-1))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
