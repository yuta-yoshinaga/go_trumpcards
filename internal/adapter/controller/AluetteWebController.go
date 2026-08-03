//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AluetteWebInput アリュエットのWebインプット
//
// **捨て札コマンドは無い。**アリュエットには余剰札を伏せる工程が無いので、
// タロー系から写した cardIndices は存在しない。
type AluetteWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *AluetteWebConfig `json:"config,omitempty"`
}

// AluetteWebConfig アリュエットのWeb設定
type AluetteWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// AluetteWebOutputPlayer アリュエットのWebアウトプットプレイヤー
type AluetteWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Team       int              `json:"team"`
	IsDealer   bool             `json:"isDealer"`
}

// AluetteWebOutput アリュエットのWebアウトプット
type AluetteWebOutput struct {
	Players          []*AluetteWebOutputPlayer    `json:"players"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	DealerIdx        int                          `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	TeamScores       [2]int                       `json:"teamScores"`
	RoundTricks      [domain.AluettePlayerCnt]int `json:"roundTricks"`
	LastTrickWinner  int                          `json:"lastTrickWinner"`
	PlayableIndices  []int                        `json:"playableIndices"`
	// Luettes 名前つき最強札 6 枚を**強い順**で返す。ドメインの序列表そのもの。
	Luettes     []AluetteWebOutputLuette `json:"luettes"`
	GameEndFlag bool                     `json:"gameEndFlag"`
	WinnerTeam  int                      `json:"winnerTeam"`
	IsHumanTurn bool                     `json:"isHumanTurn"`
	Hint        *WebOutputCardHint       `json:"hint,omitempty"`
	WebOutputBase
	Config AluetteWebOutputConfig `json:"config"`
}

// AluetteWebOutputLuette は名前つき最強札 1 枚の送出形。
//
// **スートは札と同じ表記で送る。**盤面の札は design を "DIAMOND" のような
// 文字列で受け取るので、序列表だけ数値で送ると画面側が両者を突き合わせる
// ためだけの対応表を持つことになる。
type AluetteWebOutputLuette struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
	Name   string `json:"name"`
}

// AluetteWebOutputConfig アリュエットの設定アウトプット
type AluetteWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds an AluetteConfig from the nested web config, applying bounds checking.
func (c *AluetteWebConfig) ToConfig() domain.AluetteConfig {
	cfg := domain.DefaultAluetteConfig()
	cfg.CpuDifficulty = domain.AluetteCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.AluetteCpuDifficultyEasy), int(domain.AluetteCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000)
	return cfg
}

// ToConfig builds an AluetteConfig from the web input.
func (p AluetteWebInput) ToConfig() domain.AluetteConfig {
	return configOrDefault(p.Config, (*AluetteWebConfig).ToConfig, domain.DefaultAluetteConfig())
}

// AluetteWebController アリュエットのWebコントローラークラス
type AluetteWebController = GameWebController[usecase.AluetteInteractorIF, AluetteWebInput, *AluetteWebOutput]

// NewAluetteWebController and NewAluetteWebControllerWithProvider are the standard
// and provider-backed constructors for AluetteWebController.
var NewAluetteWebController, NewAluetteWebControllerWithProvider = webControllerPair[usecase.AluetteInteractorIF, AluetteWebInput, *AluetteWebOutput](
	newAluetteDefaultOutput, aluetteDispatch,
)

func newAluetteDefaultOutput(msg string) *AluetteWebOutput {
	return &AluetteWebOutput{
		Players:         make([]*AluetteWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		// **エラー封筒には盤面が無い。**プレゼンターを通らない経路なので序列表も
		// 空で送り、null で画面側の find を落とさないことだけ担保する。
		Luettes:         make([]AluetteWebOutputLuette, 0),
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func aluetteDispatch(bc *baseController, w http.ResponseWriter, di usecase.AluetteInteractorIF, param AluetteWebInput, newDefault func(string) *AluetteWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
