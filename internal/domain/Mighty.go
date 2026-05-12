package domain

// MightyPlayerCnt マイティのプレイヤー数 (5人固定)
const MightyPlayerCnt = 5

// MightyHandSize 各プレイヤーの手札枚数 (10枚)
const MightyHandSize = 10

// MightyTotalCards デッキ総枚数 (52 + Joker 1)
const MightyTotalCards = 53

// MightyKittySize 場札 (キティ) の枚数 (3枚)
const MightyKittySize = 3

// MightyTricksPerRound 1ラウンドあたりのトリック数
const MightyTricksPerRound = MightyHandSize

// MightyMaxPoints ラウンドあたりの得点札合計 (10/J/Q/K/A × 4スート = 20)
const MightyMaxPoints = 20

// MightyWinnerTeam 勝利チーム
const (
	// MightyWinnerUndecided 未確定
	MightyWinnerUndecided = -1
	// MightyWinnerDeclarer 宣言者 + パートナー (与党)
	MightyWinnerDeclarer = 0
	// MightyWinnerOpposition 野党 (3人)
	MightyWinnerOpposition = 1
)
