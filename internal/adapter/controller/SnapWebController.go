//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SnapWebInput スナップ Web 入力
type SnapWebInput struct {
	BaseWebInput
	Config *SnapWebConfig `json:"config,omitempty"`
}

// SnapWebConfig スナップの設定リクエスト
type SnapWebConfig struct {
	PlayerCnt     *int `json:"playerCnt,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// SnapWebPlayer プレイヤー出力
type SnapWebPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	StockSize int  `json:"stockSize"`
}

// SnapWebOutputHint ヒント出力
type SnapWebOutputHint struct {
	// Snap が true なら「いま宣言すべき」。
	Snap   bool   `json:"snap"`
	Reason string `json:"reason"`
}

// SnapWebOutput スナップ Web 出力
type SnapWebOutput struct {
	Phase       int  `json:"phase"`
	GameEndFlag bool `json:"gameEndFlag"`
	// WinnerIdx は -1 のあいだ未確定（決着なしも -1）。
	WinnerIdx      int  `json:"winnerIdx"`
	CurrentTurnIdx int  `json:"currentTurnIdx"`
	IsHumanTurn    bool `json:"isHumanTurn"`
	// SnapAvailable はいま宣言が正しいか。
	//
	// **上 2 枚が同ランクのときだけ真。** 場札 1 枚では決して立ちません。
	SnapAvailable      bool               `json:"snapAvailable"`
	CenterPileSize     int                `json:"centerPileSize"`
	TopCard            *WebOutputCard     `json:"topCard,omitempty"`
	Players            []*SnapWebPlayer   `json:"players"`
	PlayerCnt          int                `json:"playerCnt"`
	CpuDifficulty      int                `json:"cpuDifficulty"`
	PendingKind        int                `json:"pendingKind"`
	PendingDeadlineMs  int64              `json:"pendingDeadlineMs"`
	LastEventKind      int                `json:"lastEventKind"`
	LastEventPlayerIdx int                `json:"lastEventPlayerIdx"`
	Hint               *SnapWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// snapConfigFromInput は現在の設定に入力を重ねる。
func snapConfigFromInput(cur domain.SnapConfig, in *SnapWebConfig) domain.SnapConfig {
	if in == nil {
		return cur
	}
	cfg := cur
	cfg.PlayerCnt = webutil.BoundedIntPtr(in.PlayerCnt,
		domain.SnapPlayerCntMin, domain.SnapPlayerCntMax, cur.PlayerCnt)
	cfg.CpuDifficulty = domain.SnapCpuDifficulty(webutil.BoundedIntPtr(in.CpuDifficulty,
		int(domain.SnapCpuEasy), int(domain.SnapCpuHard), int(cur.CpuDifficulty)))
	return cfg
}

// SnapWebController スナップ Web コントローラー
type SnapWebController = GameWebController[usecase.SnapInteractorIF, SnapWebInput, *SnapWebOutput]

// NewSnapWebController and NewSnapWebControllerWithProvider are the
// standard and provider-backed constructors for SnapWebController.
var NewSnapWebController, NewSnapWebControllerWithProvider = webControllerPair[usecase.SnapInteractorIF, SnapWebInput, *SnapWebOutput](
	newSnapDefaultOutput, snapDispatch,
)

func newSnapDefaultOutput(msg string) *SnapWebOutput {
	cfg := domain.DefaultSnapConfig()
	return &SnapWebOutput{
		WinnerIdx:     -1,
		PlayerCnt:     cfg.PlayerCnt,
		CpuDifficulty: int(cfg.CpuDifficulty),
		Players:       make([]*SnapWebPlayer, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func snapDispatch(bc *baseController, w http.ResponseWriter, si usecase.SnapInteractorIF, param SnapWebInput, _ func(string) *SnapWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, si.ResetWithConfig(snapConfigFromInput(si.GetConfig(), param.Config)))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
	case "s", "step":
		bc.writePresenterResponse(w, si.Step())
	case "n", "snap":
		// **人間 (席 0) の宣言専用。** クライアントが席を指定できると、
		// CPU に強制的に誤宣言させられます。
		bc.writePresenterResponse(w, si.Snap())
	case "t", "tick":
		bc.writePresenterResponse(w, si.Tick())
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
