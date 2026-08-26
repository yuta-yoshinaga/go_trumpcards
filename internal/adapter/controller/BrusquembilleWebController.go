//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BrusquembilleWebInput ブリュスカンビーユWebインプット
type BrusquembilleWebInput struct {
	BaseWebInput
	CardIndex *int                    `json:"cardIndex,omitempty"`
	Config    *BrusquembilleWebConfig `json:"config,omitempty"`
}

// BrusquembilleWebConfig ブリュスカンビーユWeb設定
type BrusquembilleWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	// PlayerCnt 席数 (2-5)。省略時は既定の 2。
	PlayerCnt *int `json:"playerCnt,omitempty"`
}

// BrusquembilleWebOutputPlayer ブリュスカンビーユWebアウトプットプレイヤー
type BrusquembilleWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Points     int              `json:"points"`
	TrickCount int              `json:"trickCount"`
}

// BrusquembilleWebOutputHint ヒント出力
type BrusquembilleWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// BrusquembilleWebOutput ブリュスカンビーユWebアウトプット
type BrusquembilleWebOutput struct {
	Players          []*BrusquembilleWebOutputPlayer `json:"players"`
	Phase            int                             `json:"phase"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	TrumpSuit        int                             `json:"trumpSuit"`
	TrumpCard        *WebOutputCard                  `json:"trumpCard,omitempty"`
	DealerIdx        int                             `json:"dealerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	StockRemaining   int                             `json:"stockRemaining"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerIdx        int                             `json:"winnerIdx"`
	// ValidIndices は人間 (席 0) がいま合法に出せる手札の位置。
	//
	// **前半は全部、後半は追従できる札だけ。** クローン元のブリスコラは
	// いつでも何を出してもよいのでこの情報が要らなかったが、山札が尽きた
	// 後のブリュスカンビーユには非合法手がある。渡さないと、画面は押せる
	// ように見せておいて実行時にだけ拒否することになる。
	ValidIndices []int `json:"validIndices"`
	// FollowRequired は「いま追従義務があるか」。UI の説明文に使う。
	FollowRequired bool                        `json:"followRequired"`
	Hint           *BrusquembilleWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config BrusquembilleWebOutputConfig `json:"config"`
}

// BrusquembilleWebOutputConfig ブリュスカンビーユ設定アウトプット
type BrusquembilleWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	// PlayerCnt 席数 (2-5)
	PlayerCnt int `json:"playerCnt"`
}

// ToConfig builds a BrusquembilleConfig from the nested web config, applying bounds checking.
func (c *BrusquembilleWebConfig) ToConfig() domain.BrusquembilleConfig {
	cfg := domain.DefaultBrusquembilleConfig()
	cfg.CpuDifficulty = domain.BrusquembilleCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.BrusquembilleCpuDifficultyNormal), int(domain.BrusquembilleCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	// **席数も読む。** 読まないと、設定を送っても常に 2 人卓になる。
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.BrusquembilleMinPlayerCnt, domain.BrusquembilleMaxPlayerCnt, cfg.PlayerCnt)
	return cfg
}

// ToConfig builds a BrusquembilleConfig from the web input.
func (p BrusquembilleWebInput) ToConfig() domain.BrusquembilleConfig {
	return configOrDefault(p.Config, (*BrusquembilleWebConfig).ToConfig, domain.DefaultBrusquembilleConfig())
}

// BrusquembilleWebController ブリュスカンビーユWebコントローラークラス
type BrusquembilleWebController = GameWebController[usecase.BrusquembilleInteractorIF, BrusquembilleWebInput, *BrusquembilleWebOutput]

// NewBrusquembilleWebController and NewBrusquembilleWebControllerWithProvider are
// the standard and provider-backed constructors for BrusquembilleWebController.
var NewBrusquembilleWebController, NewBrusquembilleWebControllerWithProvider = webControllerPair[usecase.BrusquembilleInteractorIF, BrusquembilleWebInput, *BrusquembilleWebOutput](
	newBrusquembilleDefaultOutput, brusquembilleDispatch,
)

func newBrusquembilleDefaultOutput(msg string) *BrusquembilleWebOutput {
	return &BrusquembilleWebOutput{
		Players:       make([]*BrusquembilleWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func brusquembilleDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BrusquembilleInteractorIF, param BrusquembilleWebInput, newDefault func(string) *BrusquembilleWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextTrick())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
