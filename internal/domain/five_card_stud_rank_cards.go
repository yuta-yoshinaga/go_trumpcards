//go:build !js || !wasm || casino

package domain

// FiveCardStudRankCards は、評価済みの5枚から役を構成するカードだけを返す。
// Soko モードでは4枚ストレートと4枚フラッシュも扱う。
func FiveCardStudRankCards(best []*Card, rank int, soko bool) []*Card {
	if len(best) != 5 {
		return []*Card{}
	}
	rankGroups := make(map[int][]*Card, 5)
	suitGroups := make(map[int][]*Card, 4)
	for _, card := range best {
		rankGroups[card.GetValue()] = append(rankGroups[card.GetValue()], card)
		suitGroups[card.GetDesign()] = append(suitGroups[card.GetDesign()], card)
	}
	selectGroups := func(minimum int) []*Card {
		out := make([]*Card, 0, 4)
		for _, card := range best {
			if len(rankGroups[card.GetValue()]) >= minimum {
				out = append(out, card)
			}
		}
		return out
	}
	if soko {
		switch rank {
		case SokoHandHighCard:
			return highestCard(best)
		case SokoHandOnePair:
			return selectGroups(2)
		case SokoHandFourStraight:
			return fourStraightCards(best)
		case SokoHandFourFlush:
			// best の順に見て、4 枚そろったスートの札だけを返す (map を走査しない)。
			for _, card := range best {
				if len(suitGroups[card.GetDesign()]) == 4 {
					out := make([]*Card, 0, 4)
					for _, c := range best {
						if c.GetDesign() == card.GetDesign() {
							out = append(out, c)
						}
					}
					return out
				}
			}
		case SokoHandTwoPair:
			return selectGroups(2)
		case SokoHandThreeOfAKind:
			return selectGroups(3)
		case SokoHandFourOfAKind:
			return selectGroups(4)
		default:
			return append([]*Card(nil), best...)
		}
		return []*Card{}
	}
	switch rank {
	case PokerHandHighCard:
		return highestCard(best)
	case PokerHandOnePair:
		return selectGroups(2)
	case PokerHandTwoPair:
		return selectGroups(2)
	case PokerHandThreeOfAKind:
		return selectGroups(3)
	case PokerHandFourOfAKind:
		return selectGroups(4)
	default:
		return append([]*Card(nil), best...)
	}
}

func highestCard(cards []*Card) []*Card {
	best := cards[0]
	for _, card := range cards[1:] {
		if card.GetValue() == 1 && best.GetValue() != 1 || card.GetValue() > best.GetValue() && best.GetValue() != 1 {
			best = card
		}
	}
	return []*Card{best}
}

func fourStraightCards(cards []*Card) []*Card {
	present := make(map[int]*Card, 6)
	for _, card := range cards {
		v := card.GetValue()
		present[v] = card
		if v == 1 {
			present[CardValueMax+1] = card
		}
	}
	for start := 1; start <= CardValueMax+1-3; start++ {
		out := make([]*Card, 0, 4)
		for i := range 4 {
			card := present[start+i]
			if card == nil {
				break
			}
			out = append(out, card)
		}
		if len(out) == 4 {
			return out
		}
	}
	return []*Card{}
}
