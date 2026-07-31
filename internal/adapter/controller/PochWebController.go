//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PochWebInput ポッホWebインプット
type PochWebInput struct {
	BaseWebInput
	CardIndex *int           `json:"cardIndex,omitempty"`
	Config    *PochWebConfig `json:"config,omitempty"`
}

// PochWebConfig ポッホWeb設定
type PochWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// PochWebOutputPool は盤の 1 区画。
type PochWebOutputPool struct {
	// Name は "ace" / "marriage" / "sequence" / "pocher" / "centre" など。
	Name string `json:"name"`
	// Chips は残高。**取られなかった区画は持ち越される**ので、これは
	// 貯まっていく。
	Chips int `json:"chips"`
}

// PochWebOutputAward は第 1 段階で誰がどの区画を取ったかの記録。
type PochWebOutputAward struct {
	Pool   string `json:"pool"`
	Player int    `json:"player"`
	Chips  int    `json:"chips"`
}

// PochWebOutputPlayer ポッホWebアウトプットプレイヤー
type PochWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る。**残り 1 枚につき 1 チップ
	// 払う**ので、常に見えている必要がある。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	Chips     int              `json:"chips"`
	// Bet はこの pochen ラウンドで出した額。
	Bet    int  `json:"bet"`
	Folded bool `json:"folded"`
	Hidden bool `json:"hidden"`
}

// PochWebOutputHint ヒント出力
type PochWebOutputHint struct {
	// Action は "bet" / "fold" / "play"。
	Action    string `json:"action"`
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// PochWebOutput ポッホWebアウトプット
type PochWebOutput struct {
	Players          []*PochWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	Pools            []*PochWebOutputPool   `json:"pools"`
	// PaySuit は表向きの余り札のスート。第 1 段階はこれだけで決まる。
	PaySuit int `json:"paySuit"`
	// TurnUp は表向きにした余り札そのもの。
	TurnUp *WebOutputCard `json:"turnUp,omitempty"`
	// StakingAwards は第 1 段階の結果。自動で解決するので、何が起きたかは
	// これでしか読めない。
	StakingAwards []*PochWebOutputAward `json:"stakingAwards"`
	BetTarget     int                   `json:"betTarget"`
	// PochenWinner は組の比べ合いを制した席 (-1: 未決着)。**宣言ではない。**
	PochenWinner int `json:"pochenWinner"`
	PochenPot    int `json:"pochenPot"`
	// PlayedPile はストップスで出た札。
	PlayedPile []*WebOutputCard `json:"playedPile"`
	// StopsSuit は今の並びのスート (-1: 好きな札で始められる)。
	StopsSuit int `json:"stopsSuit"`
	// StopsRank は今の並びの最高ランク (A は 14 として数える)。
	StopsRank   int                `json:"stopsRank"`
	DealNo      int                `json:"dealNo"`
	TargetDeals int                `json:"targetDeals"`
	DealWinner  int                `json:"dealWinner"`
	GameEndFlag bool               `json:"gameEndFlag"`
	WinnerIdx   int                `json:"winnerIdx"`
	Hint        *PochWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config PochWebOutputConfig `json:"config"`
}

// PochWebOutputConfig ポッホ設定アウトプット
type PochWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig builds a PochConfig from the nested web config, applying bounds checking.
func (c *PochWebConfig) ToConfig() domain.PochConfig {
	cfg := domain.DefaultPochConfig()
	cfg.CpuDifficulty = domain.PochCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.PochCpuDifficultyNormal), int(domain.PochCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 100)
	return cfg
}

// ToConfig builds a PochConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *PochWebConfig and calling the method on it would
// dereference nil.
func (i PochWebInput) ToConfig() domain.PochConfig {
	return configOrDefault(i.Config, (*PochWebConfig).ToConfig, domain.DefaultPochConfig())
}

// PochWebController ポッホWebコントローラ
type PochWebController = GameWebController[usecase.PochInteractorIF, PochWebInput, *PochWebOutput]

// NewPochWebController and NewPochWebControllerWithProvider are the standard
// and provider-backed constructors for PochWebController.
var NewPochWebController, NewPochWebControllerWithProvider = webControllerPair[usecase.PochInteractorIF, PochWebInput, *PochWebOutput](
	newPochDefaultOutput, pochDispatch,
)

func newPochDefaultOutput(msg string) *PochWebOutput {
	return &PochWebOutput{
		Players:       make([]*PochWebOutputPlayer, 0),
		Pools:         make([]*PochWebOutputPool, 0),
		StakingAwards: make([]*PochWebOutputAward, 0),
		PlayedPile:    make([]*WebOutputCard, 0),
		TargetDeals:   domain.DefaultPochConfig().TargetDeals,
		StopsSuit:     -1,
		PochenWinner:  -1,
		DealWinner:    -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pochDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PochInteractorIF, param PochWebInput, newDefault func(string) *PochWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "b", "bet":
		bc.writePresenterResponse(w, pi.Bet())
	case "f", "fold":
		bc.writePresenterResponse(w, pi.Fold())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, pi.NextDeal())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}

// NewPochDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewPochDefaultOutputForTest(msg string) *PochWebOutput {
	return newPochDefaultOutput(msg)
}
