package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GongZhuWebInput 拱猪Webインプット
type GongZhuWebInput struct {
	BaseWebInput
	CardIndices []int             `json:"cardIndices,omitempty"`
	CardIndex   *int              `json:"cardIndex,omitempty"`
	Config      *GongZhuWebConfig `json:"config,omitempty"`
}

// GongZhuWebConfig 拱猪Web設定
type GongZhuWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// GongZhuWebOutputPlayer 拱猪Webアウトプットプレイヤー
type GongZhuWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// CapturedPointCards はこのプレイヤーがこれまでのトリックで獲得した
	// 得点札（ハート各札・♠Q=豚・♦J=羊・♣10=倍化）の一覧。トリック獲得時に
	// 公開される公開情報のため、全プレイヤー分を送出する。
	CapturedPointCards []*WebOutputCard `json:"capturedPointCards"`
	RoundScore         int              `json:"roundScore"`
	CumulativeScore    int              `json:"cumulativeScore"`
	TrickCount         int              `json:"trickCount"`
}

// GongZhuWebOutputBreakdown はラウンド得点の内訳。
type GongZhuWebOutputBreakdown struct {
	HeartCount        int  `json:"heartCount"`
	HeartsSum         int  `json:"heartsSum"`
	AllHearts         bool `json:"allHearts"`
	AceExposed        bool `json:"aceExposed"`
	HasPig            bool `json:"hasPig"`
	PigExposed        bool `json:"pigExposed"`
	HasSheep          bool `json:"hasSheep"`
	SheepExposed      bool `json:"sheepExposed"`
	HasDoubler        bool `json:"hasDoubler"`
	DoublerMultiplier int  `json:"doublerMultiplier"`
	DoublerStandalone int  `json:"doublerStandalone"`
	Subtotal          int  `json:"subtotal"`
	Total             int  `json:"total"`
}

// GongZhuWebOutputExposure 公開されたポイントカード
type GongZhuWebOutputExposure struct {
	Pig     bool `json:"pig"`
	Sheep   bool `json:"sheep"`
	Ace     bool `json:"ace"`
	Doubler bool `json:"doubler"`
}

// GongZhuWebOutput 拱猪Webアウトプット
type GongZhuWebOutput struct {
	Players          []*GongZhuWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	TrickNumber      int                       `json:"trickNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard     `json:"currentTrick"`
	// PlayableIndices はいま出せる手札の位置。マストフォローの可視化に使う (#4812)。
	PlayableIndices  []int                    `json:"playableIndices"`
	HeartsBroken     bool                     `json:"heartsBroken"`
	Exposed          GongZhuWebOutputExposure `json:"exposed"`
	ExposableIndices []int                    `json:"exposableIndices"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	Hint             *WebOutputCardHint       `json:"hint,omitempty"`
	// ScoreBreakdowns はラウンド終了時の得点内訳 (プレイヤー順)。
	// **ドメインの計算そのもの**を運ぶので、画面の説明と実際の点が食い違わない (#5630)。
	ScoreBreakdowns []*GongZhuWebOutputBreakdown `json:"scoreBreakdowns,omitempty"`
	WebOutputBase
	Config GongZhuWebOutputConfig `json:"config"`
}

// GongZhuWebOutputConfig 拱猪設定アウトプット
type GongZhuWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a GongZhuConfig from the nested web config, applying bounds checking.
func (c *GongZhuWebConfig) ToConfig() domain.GongZhuConfig {
	cfg := domain.DefaultGongZhuConfig()
	cfg.CpuDifficulty = domain.GongZhuCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.GongZhuCpuDifficultyEasy), int(domain.GongZhuCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 100000)
	return cfg
}

// ToConfig builds a GongZhuConfig from the web input.
func (p GongZhuWebInput) ToConfig() domain.GongZhuConfig {
	return configOrDefault(p.Config, (*GongZhuWebConfig).ToConfig, domain.DefaultGongZhuConfig())
}

// GongZhuWebController 拱猪Webコントローラークラス
type GongZhuWebController = GameWebController[usecase.GongZhuInteractorIF, GongZhuWebInput, *GongZhuWebOutput]

// NewGongZhuWebController and NewGongZhuWebControllerWithProvider are
// the standard and provider-backed constructors for GongZhuWebController.
var NewGongZhuWebController, NewGongZhuWebControllerWithProvider = webControllerPair[usecase.GongZhuInteractorIF, GongZhuWebInput, *GongZhuWebOutput](
	newGongZhuDefaultOutput, gongZhuDispatch,
)

func newGongZhuDefaultOutput(msg string) *GongZhuWebOutput {
	return &GongZhuWebOutput{
		Players:          make([]*GongZhuWebOutputPlayer, 0),
		CurrentTrick:     make([]*WebOutputTrickCard, 0),
		ExposableIndices: make([]int, 0),
		WinnerIdx:        -1,
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func gongZhuDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GongZhuInteractorIF, param GongZhuWebInput, newDefault func(string) *GongZhuWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, gi.ResetWithConfig(param.ToConfig()))
	case "expose":
		bc.writePresenterResponse(w, gi.Expose(param.CardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, gi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, gi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, gi.Hint, gi.ActionLog)
	}
	return true
}
