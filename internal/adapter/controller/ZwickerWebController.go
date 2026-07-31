//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ZwickerWebInput ツヴィッカーWebインプット
type ZwickerWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// PlayedValue は出す札をどの値で扱うか。**A と絵札は 2 択を持つ**ので、
	// 札だけでは捕獲を決められない。
	PlayedValue *int `json:"playedValue,omitempty"`
	// TableIndices は取る (またはビルドに使う) 場札の添字。
	TableIndices []int `json:"tableIndices,omitempty"`
	// BuildIndices は取る場のビルドの添字。
	BuildIndices []int `json:"buildIndices,omitempty"`
	// DeclaredValue はビルドの宣言値。
	DeclaredValue *int              `json:"declaredValue,omitempty"`
	Config        *ZwickerWebConfig `json:"config,omitempty"`
}

// ZwickerWebConfig ツヴィッカーWeb設定
type ZwickerWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// ZwickerWebOutputCard は場札・手札 1 枚と、その札が取りうるマッチ値。
type ZwickerWebOutputCard struct {
	*WebOutputCard
	// Values は取りうるマッチ値。A/絵札は 2 つ、それ以外は 1 つ。
	// クライアントに値表を持たせるとサーバーとずれるので毎回送る。
	Values []int `json:"values"`
}

// ZwickerWebOutputPlayer ツヴィッカーWebアウトプットプレイヤー
type ZwickerWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0 か 1。向かい合わせが味方。
	Team int `json:"team"`
	// CardCount は手札の枚数。伏せている間も送る。
	CardCount int                     `json:"cardCount"`
	Cards     []*ZwickerWebOutputCard `json:"cards"`
	// CapturedCount はこのディールで取った枚数。枚数最多の 3 点に効く。
	CapturedCount int `json:"capturedCount"`
	// Zwicks は場を空にした回数。1 回 1 点。
	Zwicks int  `json:"zwicks"`
	Hidden bool `json:"hidden"`
}

// ZwickerWebOutputBuild 場のビルド
type ZwickerWebOutputBuild struct {
	Owner int              `json:"owner"`
	Value int              `json:"value"`
	Cards []*WebOutputCard `json:"cards"`
}

// ZwickerWebOutputRoundScore は直近ディールの内訳。
type ZwickerWebOutputRoundScore struct {
	CardPoints [2]int `json:"cardPoints"`
	Cards      [2]int `json:"cards"`
	// MajorityTeam は枚数最多のチーム (-1 = 同数で誰も取らない)。
	MajorityTeam int    `json:"majorityTeam"`
	Zwicks       [2]int `json:"zwicks"`
	Total        [2]int `json:"total"`
}

// ZwickerWebOutputHint ヒント出力
type ZwickerWebOutputHint struct {
	// Take が真なら取るのが推奨手。偽なら捨てる。
	Take      bool   `json:"take"`
	CardIndex *int   `json:"cardIndex,omitempty"`
	Value     int    `json:"value,omitempty"`
	TableIdxs []int  `json:"tableIndices,omitempty"`
	Reason    string `json:"reason"`
}

// ZwickerWebOutput ツヴィッカーWebアウトプット
type ZwickerWebOutput struct {
	Players          []*ZwickerWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	StockCount       int                       `json:"stockCount"`
	TableCards       []*ZwickerWebOutputCard   `json:"tableCards"`
	Builds           []*ZwickerWebOutputBuild  `json:"builds"`
	// TeamScores はチーム別の累計得点。
	TeamScores [2]int `json:"teamScores"`
	// TargetScore はこれに達したチームが勝つ点。クライアントに書き写させない。
	TargetScore int `json:"targetScore"`
	// LastRound は直近ディールの内訳 (未精算なら null)。
	LastRound   *ZwickerWebOutputRoundScore `json:"lastRound,omitempty"`
	GameEndFlag bool                        `json:"gameEndFlag"`
	// WinnerTeam は勝ったチーム (-1: 未決着)。
	WinnerTeam int                   `json:"winnerTeam"`
	Hint       *ZwickerWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config ZwickerWebOutputConfig `json:"config"`
}

// ZwickerWebOutputConfig ツヴィッカー設定アウトプット
type ZwickerWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a ZwickerConfig from the nested web config, applying bounds checking.
func (c *ZwickerWebConfig) ToConfig() domain.ZwickerConfig {
	cfg := domain.DefaultZwickerConfig()
	cfg.CpuDifficulty = domain.ZwickerCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.ZwickerCpuDifficultyNormal), int(domain.ZwickerCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 1000)
	return cfg
}

// ToConfig builds a ZwickerConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *ZwickerWebConfig and calling the method on it would
// dereference nil.
func (i ZwickerWebInput) ToConfig() domain.ZwickerConfig {
	return configOrDefault(i.Config, (*ZwickerWebConfig).ToConfig, domain.DefaultZwickerConfig())
}

// ZwickerWebController ツヴィッカーWebコントローラ
type ZwickerWebController = GameWebController[usecase.ZwickerInteractorIF, ZwickerWebInput, *ZwickerWebOutput]

// NewZwickerWebController and NewZwickerWebControllerWithProvider are the
// standard and provider-backed constructors for ZwickerWebController.
var NewZwickerWebController, NewZwickerWebControllerWithProvider = webControllerPair[usecase.ZwickerInteractorIF, ZwickerWebInput, *ZwickerWebOutput](
	newZwickerDefaultOutput, zwickerDispatch,
)

func newZwickerDefaultOutput(msg string) *ZwickerWebOutput {
	return &ZwickerWebOutput{
		Players:       make([]*ZwickerWebOutputPlayer, 0),
		TableCards:    make([]*ZwickerWebOutputCard, 0),
		Builds:        make([]*ZwickerWebOutputBuild, 0),
		TargetScore:   domain.DefaultZwickerConfig().TargetScore,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func zwickerDispatch(bc *baseController, w http.ResponseWriter, zi usecase.ZwickerInteractorIF, param ZwickerWebInput, newDefault func(string) *ZwickerWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, zi.ResetWithConfig(param.ToConfig()))
	case "t", "take":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil || param.PlayedValue == nil,
			"param error: cardIndex and playedValue are required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Take(*param.CardIndex, *param.PlayedValue, param.TableIndices, param.BuildIndices))
	case "b", "build":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil || param.DeclaredValue == nil,
			"param error: cardIndex and declaredValue are required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Build(*param.CardIndex, param.TableIndices, *param.DeclaredValue))
	// **"l" は使えない。**dispatchHintAndLog が棋譜のエイリアスとして先に
	// 押さえているので、ここで奪うと log が死ぬ。
	case "tr", "trail":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, zi.Trail(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, zi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, zi.Hint, zi.ActionLog)
	}
	return true
}

// NewZwickerDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewZwickerDefaultOutputForTest(msg string) *ZwickerWebOutput {
	return newZwickerDefaultOutput(msg)
}
