//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CoincheWebInput コワンシュWebインプット
type CoincheWebInput struct {
	BaseWebInput
	// Points 宣言する目標点 (bid コマンド; 80..180 と 250=Capot)
	Points *int `json:"points,omitempty"`
	// Suit 宣言する切り札スート (bid コマンド; 1=♠ 2=♣ 3=♥ 4=♦)
	Suit      *int              `json:"suit,omitempty"`
	CardIndex *int              `json:"cardIndex,omitempty"`
	Config    *CoincheWebConfig `json:"config,omitempty"`
}

// CoincheWebConfig コワンシュWeb設定
type CoincheWebConfig struct {
	CpuDifficulty        *int  `json:"cpuDifficulty,omitempty"`
	TargetScore          *int  `json:"targetScore,omitempty"`
	DixDeDer             *int  `json:"dixDeDer,omitempty"`
	EnableBeloteRebelote *bool `json:"enableBeloteRebelote,omitempty"`
}

// CoincheWebOutputPlayer コワンシュWebアウトプットプレイヤー
type CoincheWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// CoincheWebOutputHint ヒント出力
type CoincheWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	Reason    string `json:"reason"`
}

// CoincheWebOutput コワンシュWebアウトプット
type CoincheWebOutput struct {
	Players          []*CoincheWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	TrickNumber      int                       `json:"trickNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	BidPlayerIdx     int                       `json:"bidPlayerIdx"`
	DealerIdx        int                       `json:"dealerIdx"`
	TrumpSuit        int                       `json:"trumpSuit"`
	// ContractPoints 落札された目標点 (0 = 未落札)
	ContractPoints int `json:"contractPoints"`
	// Multiplier 得点倍率 (1 / 2 / 4)。精算の係数そのもの。
	Multiplier int `json:"multiplier"`
	// Double 倍率の状態 (0=なし 1=コワンシュ 2=シュルコワンシュ)
	Double int `json:"double"`
	// BiddablePoints 今この席が宣言できる目標点。押せるのに必ず拒否される
	// ボタンを描かないために要る。
	BiddablePoints   []int                 `json:"biddablePoints"`
	MakerTeam        int                   `json:"makerTeam"`
	MakerPlayerIdx   int                   `json:"makerPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard `json:"currentTrick"`
	TeamScores       [2]int                `json:"teamScores"`
	RoundPoints      [2]int                `json:"roundPoints"`
	RoundBeloteBonus [2]int                `json:"roundBeloteBonus"`
	GameEndFlag      bool                  `json:"gameEndFlag"`
	WinnerTeam       int                   `json:"winnerTeam"`
	LeadPlayerIdx    int                   `json:"leadPlayerIdx"`
	Hint             *CoincheWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config CoincheWebOutputConfig `json:"config"`
}

// CoincheWebOutputConfig コワンシュ設定アウトプット
type CoincheWebOutputConfig struct {
	CpuDifficulty        int  `json:"cpuDifficulty"`
	TargetScore          int  `json:"targetScore"`
	DixDeDer             int  `json:"dixDeDer"`
	EnableBeloteRebelote bool `json:"enableBeloteRebelote"`
}

// ToConfig builds a CoincheConfig from the nested web config, applying bounds checking.
func (c *CoincheWebConfig) ToConfig() domain.CoincheConfig {
	cfg := domain.DefaultCoincheConfig()
	cfg.CpuDifficulty = domain.CoincheCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CoincheCpuDifficultyEasy), int(domain.CoincheCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 10000)
	webutil.ApplyBoundedInt(&cfg.DixDeDer, c.DixDeDer, 0, 100)
	if c.EnableBeloteRebelote != nil {
		cfg.EnableBeloteRebelote = *c.EnableBeloteRebelote
	}
	return cfg
}

// ToConfig builds a CoincheConfig from the web input.
func (p CoincheWebInput) ToConfig() domain.CoincheConfig {
	return configOrDefault(p.Config, (*CoincheWebConfig).ToConfig, domain.DefaultCoincheConfig())
}

// CoincheWebController コワンシュWebコントローラークラス
type CoincheWebController = GameWebController[usecase.CoincheInteractorIF, CoincheWebInput, *CoincheWebOutput]

// NewCoincheWebController and NewCoincheWebControllerWithProvider are
// the standard and provider-backed constructors for CoincheWebController.
var NewCoincheWebController, NewCoincheWebControllerWithProvider = webControllerPair[usecase.CoincheInteractorIF, CoincheWebInput, *CoincheWebOutput](
	newCoincheDefaultOutput, coincheDispatch,
)

func newCoincheDefaultOutput(msg string) *CoincheWebOutput {
	return &CoincheWebOutput{
		Players:       make([]*CoincheWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func coincheDispatch(bc *baseController, w http.ResponseWriter, bi usecase.CoincheInteractorIF, param CoincheWebInput, newDefault func(string) *CoincheWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		// **点とスートは 2 つで 1 つの宣言。** 片方だけ来た要求を通すと、
		// 残りに既定値が入って別の契約になる。
		if !requireParam(bc, w, newDefault, param.Points == nil, "param error: points is required.") {
			return true
		}
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Bid(*param.Points, *param.Suit))
	case "pa", "pass":
		bc.writePresenterResponse(w, bi.Pass())
	case "co", "coinche":
		bc.writePresenterResponse(w, bi.Coinche())
	case "su", "surcoinche":
		bc.writePresenterResponse(w, bi.Surcoinche())
	case "ok", "decline":
		bc.writePresenterResponse(w, bi.DeclineDouble())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, bi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
