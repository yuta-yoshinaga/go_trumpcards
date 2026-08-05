//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GutsWebConfig はガッツ (Guts) の Web 設定。
type GutsWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ToConfig は GutsWebConfig を domain.GutsConfig に変換する (境界チェック付き)。
func (c *GutsWebConfig) ToConfig() domain.GutsConfig {
	cfg := domain.DefaultGutsConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.GutsMinPlayerCount, domain.GutsMaxPlayerCount)
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, domain.GutsMinAnte, domain.GutsMaxAnte)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, domain.GutsMinStartingChips, domain.GutsMaxStartingChips)
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.GutsMinTargetRounds, domain.GutsMaxTargetRounds)
	return cfg
}

// GutsWebInput はガッツ Web インプット。
type GutsWebInput struct {
	BaseWebInput
	// Declaration は宣言種別 (0=out, 1=in)。declare コマンドで必須。
	Declaration *int           `json:"declaration,omitempty"`
	Config      *GutsWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.GutsConfig を構築する。
func (p GutsWebInput) ToConfig() domain.GutsConfig {
	return configOrDefault(p.Config, (*GutsWebConfig).ToConfig, domain.DefaultGutsConfig())
}

// GutsWebOutputPlayer は 1 プレイヤーの出力。
type GutsWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Chips     int              `json:"chips"`
	In        bool             `json:"in"`
	Out       bool             `json:"out"`
	RoundBet  int              `json:"roundBet"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// HandName は公開された手の役名キー (非公開時は空文字)。
	HandName string `json:"handName,omitempty"`
	IsWinner bool   `json:"isWinner"`
	// IsMatcher はこのラウンドでマッチした (負けたイン宣言者) かどうか。
	IsMatcher bool `json:"isMatcher"`
}

// GutsWebOutputHint はヒント出力。
type GutsWebOutputHint struct {
	Declaration int    `json:"declaration"`
	Reason      string `json:"reason"`
}

// GutsWebOutputConfig は設定アウトプット。
type GutsWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
	TargetRounds  int `json:"targetRounds"`
}

// GutsWebOutput はガッツ Web アウトプット。
type GutsWebOutput struct {
	Players     []*GutsWebOutputPlayer `json:"players"`
	Phase       int                    `json:"phase"`
	RoundNumber int                    `json:"roundNumber"`
	Pot         int                    `json:"pot"`
	CarryPot    int                    `json:"carryPot"`
	// CarryCount 連続で持ち越された回数 (CUI の guts.result.carry と同じ値)。
	CarryCount     int                 `json:"carryCount"`
	Ante           int                 `json:"ante"`
	Chips          int                 `json:"chips"`
	WinnerIdx      int                 `json:"winnerIdx"`
	MatchWinnerIdx int                 `json:"matchWinnerIdx"`
	Result         int                 `json:"result"`
	Matchers       []int               `json:"matchers"`
	GameEndFlag    bool                `json:"gameEndFlag"`
	Hint           *GutsWebOutputHint  `json:"hint,omitempty"`
	Config         GutsWebOutputConfig `json:"config"`
	WebOutputBase
}

// GutsWebController はガッツ Web コントローラークラス。
type GutsWebController = GameWebController[usecase.GutsInteractorIF, GutsWebInput, *GutsWebOutput]

// NewGutsWebController, NewGutsWebControllerWithProvider are the standard and
// provider-backed constructors for GutsWebController.
var NewGutsWebController, NewGutsWebControllerWithProvider = webControllerPair[usecase.GutsInteractorIF, GutsWebInput, *GutsWebOutput](
	newGutsDefaultOutput, gutsDispatch,
)

func newGutsDefaultOutput(msg string) *GutsWebOutput {
	return &GutsWebOutput{
		Players:        make([]*GutsWebOutputPlayer, 0),
		Matchers:       make([]int, 0),
		WinnerIdx:      -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func gutsDispatch(bc *baseController, w http.ResponseWriter, ti usecase.GutsInteractorIF, param GutsWebInput, newDefault func(string) *GutsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "declare":
		if !requireParam(bc, w, newDefault, param.Declaration == nil, "param error: declaration is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Declare(*param.Declaration == int(domain.GutsDeclarationIn)))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
