//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// QuodlibetWebInput はクオドリベットの Web インプット。
type QuodlibetWebInput struct {
	BaseWebInput
	// CardIndex は出す札のインデックス (-1 = パス)。
	CardIndex *int `json:"cardIndex,omitempty"`
	// Contract は選択するコントラクト。
	Contract *int `json:"contract,omitempty"`
	// Config はゲーム設定。
	Config *QuodlibetWebConfig `json:"config,omitempty"`
}

// QuodlibetWebConfig はクオドリベットの Web 設定。
type QuodlibetWebConfig struct {
	CpuDifficulty      *int  `json:"cpuDifficulty,omitempty"`
	AutoSelectContract *bool `json:"autoSelectContract,omitempty"`
}

// QuodlibetWebOutputPlayer は 1 席ぶんの出力。
type QuodlibetWebOutputPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	CardCount int  `json:"cardCount"`
	// Cards は見えている手札。**第 3 の輪では見え方そのものが規則** なので、
	// 「自分のぶんだけ」とは限らない。
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	// Penalty は 12 ディール通算の罰点。**少ないほうが良い。**
	Penalty int `json:"penalty"`
	// DealPoints は直近ディールで負った罰点。
	DealPoints int  `json:"dealPoints"`
	OutRank    int  `json:"outRank"`
	IsDealer   bool `json:"isDealer"`
}

// QuodlibetWebOutputDeal は 1 ディールの罰点内訳。
type QuodlibetWebOutputDeal struct {
	Contract     int    `json:"contract"`
	ContractName string `json:"contractName"`
	Round        int    `json:"round"`
	DealerIdx    int    `json:"dealerIdx"`
	Points       []int  `json:"points"`
}

// QuodlibetWebOutput はクオドリベットの Web アウトプット。
type QuodlibetWebOutput struct {
	Players     []*QuodlibetWebOutputPlayer `json:"players"`
	Phase       string                      `json:"phase"`
	DealNumber  int                         `json:"dealNumber"`
	TotalDeals  int                         `json:"totalDeals"`
	RoundNumber int                         `json:"roundNumber"`
	DealerIdx   int                         `json:"dealerIdx"`
	// CurrentContract は打っているコントラクト (-1 = 未選択)。
	CurrentContract int `json:"currentContract"`
	// CurrentContractName は i18n キーに使う識別子。
	CurrentContractName string `json:"currentContractName"`
	// AvailableContracts はこの輪で残っているコントラクト。
	AvailableContracts []int `json:"availableContracts"`
	// AvailableContractNames は同じ並びの識別子。
	AvailableContractNames []string `json:"availableContractNames"`
	// IsShedding はトリックを取らないコントラクトか (四分 / 小食い)。
	IsShedding       bool                  `json:"isShedding"`
	TrickNumber      int                   `json:"trickNumber"`
	TrickCount       int                   `json:"trickCount"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                   `json:"leadPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard `json:"currentTrick"`
	LastTrick        []*WebOutputTrickCard `json:"lastTrick"`
	LastTrickWinner  int                   `json:"lastTrickWinner"`
	PlayableIndices  []int                 `json:"playableIndices"`
	// CanPass はいまパスできるか (シェディング系で合法手が無いとき)。
	CanPass bool `json:"canPass"`
	// TablePlaced は小食いの場 (スートごとに、置かれた位のインデックス)。
	TablePlaced [][]int `json:"tablePlaced"`
	// Stack は四分の現在の重ね。
	Stack           []*WebOutputCard          `json:"stack"`
	LastDeal        *QuodlibetWebOutputDeal   `json:"lastDeal,omitempty"`
	DealHistory     []*QuodlibetWebOutputDeal `json:"dealHistory"`
	Winners         []int                     `json:"winners"`
	GameEndFlag     bool                      `json:"gameEndFlag"`
	IsHumanTurn     bool                      `json:"isHumanTurn"`
	IsContractPhase bool                      `json:"isContractPhase"`
	Hint            *WebOutputCardHint        `json:"hint,omitempty"`
	// HintContract は選択フェーズで勧めるコントラクト (-1 = なし)。
	HintContract int `json:"hintContract"`
	WebOutputBase
	Config QuodlibetWebOutputConfig `json:"config"`
}

// QuodlibetWebOutputConfig は設定アウトプット。
type QuodlibetWebOutputConfig struct {
	CpuDifficulty      int  `json:"cpuDifficulty"`
	AutoSelectContract bool `json:"autoSelectContract"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *QuodlibetWebConfig) ToConfig() domain.QuodlibetConfig {
	cfg := domain.DefaultQuodlibetConfig()
	cfg.CpuDifficulty = domain.QuodlibetCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.QuodlibetCpuDifficultyEasy),
		int(domain.QuodlibetCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	if c.AutoSelectContract != nil {
		cfg.AutoSelectContract = *c.AutoSelectContract
	}
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p QuodlibetWebInput) ToConfig() domain.QuodlibetConfig {
	return configOrDefault(p.Config, (*QuodlibetWebConfig).ToConfig, domain.DefaultQuodlibetConfig())
}

// QuodlibetWebController はクオドリベットの Web コントローラー。
type QuodlibetWebController = GameWebController[usecase.QuodlibetInteractorIF, QuodlibetWebInput, *QuodlibetWebOutput]

// NewQuodlibetWebController, NewQuodlibetWebControllerWithProvider are the
// standard and provider-backed constructors.
var NewQuodlibetWebController, NewQuodlibetWebControllerWithProvider = webControllerPair[usecase.QuodlibetInteractorIF, QuodlibetWebInput, *QuodlibetWebOutput](
	newQuodlibetDefaultOutput, quodlibetDispatch,
)

func newQuodlibetDefaultOutput(msg string) *QuodlibetWebOutput {
	return &QuodlibetWebOutput{
		Players:                make([]*QuodlibetWebOutputPlayer, 0),
		AvailableContracts:     make([]int, 0),
		AvailableContractNames: make([]string, 0),
		CurrentTrick:           make([]*WebOutputTrickCard, 0),
		LastTrick:              make([]*WebOutputTrickCard, 0),
		PlayableIndices:        make([]int, 0),
		TablePlaced:            make([][]int, 0),
		Stack:                  make([]*WebOutputCard, 0),
		DealHistory:            make([]*QuodlibetWebOutputDeal, 0),
		Winners:                make([]int, 0),
		CurrentContract:        -1,
		LastTrickWinner:        -1,
		HintContract:           -1,
		TotalDeals:             domain.QuodlibetTotalDeals,
		TrickCount:             domain.QuodlibetHandSize,
		WebOutputBase:          WebOutputBase{Message: msg},
	}
}

func quodlibetDispatch(bc *baseController, w http.ResponseWriter, di usecase.QuodlibetInteractorIF, param QuodlibetWebInput, newDefault func(string) *QuodlibetWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "c", "contract":
		if !requireParam(bc, w, newDefault, param.Contract == nil, "param error: contract is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.SelectContract(*param.Contract))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "pass":
		// **パスは -1 のプレイ。** 別命令にせず同じ経路に載せることで、
		// 「パスできないのにパスした」の判定がドメイン 1 箇所で済む。
		bc.writePresenterResponse(w, di.Play(-1))
	case "nd", "nextdeal":
		bc.writePresenterResponse(w, di.NextDeal())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
