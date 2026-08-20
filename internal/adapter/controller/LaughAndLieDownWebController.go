//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LaughAndLieDownWebInput ラフ・アンド・ライダウンWebインプット
type LaughAndLieDownWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// TakeCount は場から取る枚数 (1 か 3)。省略時は 1。
	TakeCount *int                      `json:"takeCount,omitempty"`
	Config    *LaughAndLieDownWebConfig `json:"config,omitempty"`
}

// LaughAndLieDownWebConfig ラフ・アンド・ライダウンWeb設定
type LaughAndLieDownWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// LaughAndLieDownWebOutputPlayer ラフ・アンド・ライダウンWebアウトプットプレイヤー
type LaughAndLieDownWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る -- 公開情報。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// WonCount は取得枚数。8 との差がそのまま精算になるので、常に見えている必要がある。
	WonCount int `json:"wonCount"`
	// LaidDown は「取れなくなって手札を場に置いた」か。
	LaidDown bool `json:"laidDown"`
	// Score は収支。終局まで 0。
	Score  int  `json:"score"`
	Hidden bool `json:"hidden"`
}

// LaughAndLieDownWebOutputHint ヒント出力
type LaughAndLieDownWebOutputHint struct {
	CardIndex *int `json:"cardIndex,omitempty"`
	// TakeCount は推奨する取得枚数 (1 か 3)。
	TakeCount int    `json:"takeCount"`
	Reason    string `json:"reason"`
}

// LaughAndLieDownWebOutput ラフ・アンド・ライダウンWebアウトプット
type LaughAndLieDownWebOutput struct {
	Players []*LaughAndLieDownWebOutputPlayer `json:"players"`
	// Layout は表向きの場札。降りた人の手札もここに積まれる。伏せた山札は無い。
	Layout           []*WebOutputCard `json:"layout"`
	Phase            int              `json:"phase"`
	CurrentPlayerIdx int              `json:"currentPlayerIdx"`
	// ValidIndices は出せる手札の添字。場のどこかと同ランクなら出せる。
	ValidIndices []int `json:"validIndices"`
	// ThreeTakeIndices は「3 枚取り」も選べる手札の添字。ValidIndices の部分集合。
	ThreeTakeIndices []int `json:"threeTakeIndices"`
	DealerIdx        int   `json:"dealerIdx"`
	// LastInIdx は最後まで手札が残っていた人 (-1: 未決着/該当なし)。
	LastInIdx int `json:"lastInIdx"`
	// LastInBonus は上の人が受け取る額 (#5576)。CUI は精算行で額まで出しているので、
	// Web も同じ粒度で出せるよう渡す。訳文に数字を書くと、額を変えたとき片方だけ嘘になる。
	LastInBonus int `json:"lastInBonus"`
	// Pot はポットの総額。精算の内訳がこれに一致することが規則の裏取りになる。
	Pot         int                           `json:"pot"`
	GameEndFlag bool                          `json:"gameEndFlag"`
	Hint        *LaughAndLieDownWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config LaughAndLieDownWebOutputConfig `json:"config"`
}

// LaughAndLieDownWebOutputConfig ラフ・アンド・ライダウン設定アウトプット
type LaughAndLieDownWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a LaughAndLieDownConfig from the nested web config, applying bounds checking.
func (c *LaughAndLieDownWebConfig) ToConfig() domain.LaughAndLieDownConfig {
	cfg := domain.DefaultLaughAndLieDownConfig()
	cfg.CpuDifficulty = domain.LaughAndLieDownCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.LaughAndLieDownCpuDifficultyNormal), int(domain.LaughAndLieDownCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a LaughAndLieDownConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *LaughAndLieDownWebConfig and calling the method on
// it would dereference nil.
func (i LaughAndLieDownWebInput) ToConfig() domain.LaughAndLieDownConfig {
	return configOrDefault(i.Config, (*LaughAndLieDownWebConfig).ToConfig, domain.DefaultLaughAndLieDownConfig())
}

// TakeCountOrOne returns the requested take count, defaulting to one.
//
// `takeCount` is optional so an ordinary single capture needs no extra field;
// the domain rejects anything that is not one or three.
func (i LaughAndLieDownWebInput) TakeCountOrOne() int {
	if i.TakeCount == nil {
		return 1
	}
	return *i.TakeCount
}

// LaughAndLieDownWebController ラフ・アンド・ライダウンWebコントローラ
type LaughAndLieDownWebController = GameWebController[usecase.LaughAndLieDownInteractorIF, LaughAndLieDownWebInput, *LaughAndLieDownWebOutput]

// NewLaughAndLieDownWebController and NewLaughAndLieDownWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewLaughAndLieDownWebController, NewLaughAndLieDownWebControllerWithProvider = webControllerPair[usecase.LaughAndLieDownInteractorIF, LaughAndLieDownWebInput, *LaughAndLieDownWebOutput](
	newLaughAndLieDownDefaultOutput, laughAndLieDownDispatch,
)

func newLaughAndLieDownDefaultOutput(msg string) *LaughAndLieDownWebOutput {
	return &LaughAndLieDownWebOutput{
		Players:          make([]*LaughAndLieDownWebOutputPlayer, 0),
		Layout:           make([]*WebOutputCard, 0),
		ValidIndices:     make([]int, 0),
		ThreeTakeIndices: make([]int, 0),
		LastInIdx:        -1,
		LastInBonus:      domain.LaughAndLieDownLastInBonus,
		Pot:              domain.LaughAndLieDownPot,
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func laughAndLieDownDispatch(bc *baseController, w http.ResponseWriter, li usecase.LaughAndLieDownInteractorIF, param LaughAndLieDownWebInput, newDefault func(string) *LaughAndLieDownWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, li.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.Play(*param.CardIndex, param.TakeCountOrOne()))
	default:
		return dispatchHintAndLog(param.Command, bc, w, li.Hint, li.ActionLog)
	}
	return true
}
