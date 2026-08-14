//go:build !js || !wasm || solo

package domain

// 四色牌 (Tứ Sắc) のデッキ。
//
// **標準 52 枚とは構成そのものが違う。** 4 色 × 7 種 × 4 枚 = 112 枚で、
// スートは色、ランクは駒の種類にあたる ── 数字の大小という概念が無い。
//
// 新しいデッキ型は作らない。Ganjifa が 8 スートを design 1..8 で、
// FrenchTarot が切り札を design 5 で表しているのと同じで、**既存の `Card` に
// 新しい design 値を割り当てるだけ**で足りる (ADR-0033 の手続き CardFace 描画)。
// 型を増やすと保存形式が枝分かれし、`TrumpCards` の JSON を共有している
// 他の全ゲームに影響が出る。
const (
	// TuSacColorYellow は黄。
	TuSacColorYellow = 1
	// TuSacColorRed は赤。
	TuSacColorRed = 2
	// TuSacColorGreen は緑。
	TuSacColorGreen = 3
	// TuSacColorWhite は白。
	TuSacColorWhite = 4

	// TuSacColorMin は最小の色 design。
	TuSacColorMin = TuSacColorYellow
	// TuSacColorMax は最大の色 design。
	TuSacColorMax = TuSacColorWhite
	// TuSacColorCount は色数。
	TuSacColorCount = TuSacColorMax - TuSacColorMin + 1
)

// 駒の種類。**将棋の駒と同じ名前だが、強弱はない。**
//
// 順序は組み合わせの判定に使うだけで、大小の比較には使わない ── 数字の
// 大きいほうが強い、という規則がこのゲームには無い。
const (
	// TuSacPieceGeneral は 將 (将)。
	TuSacPieceGeneral = 1
	// TuSacPieceAdvisor は 士。
	TuSacPieceAdvisor = 2
	// TuSacPieceElephant は 象。
	TuSacPieceElephant = 3
	// TuSacPieceChariot は 車。
	TuSacPieceChariot = 4
	// TuSacPieceHorse は 馬。
	TuSacPieceHorse = 5
	// TuSacPieceCannon は 砲。
	TuSacPieceCannon = 6
	// TuSacPieceSoldier は 卒。
	TuSacPieceSoldier = 7

	// TuSacPieceMin は最小の駒。
	TuSacPieceMin = TuSacPieceGeneral
	// TuSacPieceMax は最大の駒。
	TuSacPieceMax = TuSacPieceSoldier
	// TuSacPieceCount は駒の種類数。
	TuSacPieceCount = TuSacPieceMax - TuSacPieceMin + 1
)

const (
	// TuSacCopies は同じ色・同じ駒の枚数。
	TuSacCopies = 4
	// TuSacDeckSize はデッキの総枚数 (4 色 × 7 種 × 4 枚)。
	TuSacDeckSize = TuSacColorCount * TuSacPieceCount * TuSacCopies
)

// TuSacColorName は色の識別子を返す (i18n キーの一部に使う)。
func TuSacColorName(design int) string {
	switch design {
	case TuSacColorYellow:
		return "yellow"
	case TuSacColorRed:
		return "red"
	case TuSacColorGreen:
		return "green"
	case TuSacColorWhite:
		return "white"
	default:
		return "unknown"
	}
}

// TuSacPieceName は駒の識別子を返す (i18n キーの一部に使う)。
func TuSacPieceName(value int) string {
	switch value {
	case TuSacPieceGeneral:
		return "general"
	case TuSacPieceAdvisor:
		return "advisor"
	case TuSacPieceElephant:
		return "elephant"
	case TuSacPieceChariot:
		return "chariot"
	case TuSacPieceHorse:
		return "horse"
	case TuSacPieceCannon:
		return "cannon"
	case TuSacPieceSoldier:
		return "soldier"
	default:
		return "unknown"
	}
}

// TuSacIsChariotHorseCannon は 車・馬・砲 のいずれかかを返す。
//
// **この 3 種だけが色をまたいだ組み合わせを作れる。** 他の駒は同色でしか
// そろわないので、ここを広げると別のゲームになる。
func TuSacIsChariotHorseCannon(value int) bool {
	return value == TuSacPieceChariot || value == TuSacPieceHorse || value == TuSacPieceCannon
}

// buildTuSacDeck は四色牌 112 枚を組む。
//
// **色ごとに 7 種を 4 枚ずつ。** 標準デッキの `NewTrumpCards` は 52 枚 (と
// ジョーカー) しか作れないので、ここで手続き的に組む。
func buildTuSacDeck() []*Card {
	deck := make([]*Card, 0, TuSacDeckSize)
	for range TuSacCopies {
		for design := TuSacColorMin; design <= TuSacColorMax; design++ {
			for piece := TuSacPieceMin; piece <= TuSacPieceMax; piece++ {
				deck = append(deck, NewCard(design, piece, false))
			}
		}
	}
	return deck
}
