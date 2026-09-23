import type { MissMilliganTableauCard } from '../types/games/missmilligan';

/**
 * Returns whether the cards from `startIndex` through the top of a tableau
 * column form an alternating-colour descending run.
 *
 * domain_rule_quoted:
 * 	func missMilliganIsRun(cards []*MissMilliganTableauCard) bool {
 * 		for i := 1; i < len(cards); i++ {
 * 			prev, cur := cards[i-1].Card, cards[i].Card
 * 			if prev == nil || cur == nil { return false }
 * 			if cur.GetValue() != prev.GetValue()-1 { return false }
 * 			if missMilliganIsRed(prev.GetDesign()) == missMilliganIsRed(cur.GetDesign()) { return false }
 * 		}
 * 		return true
 * 	}
 *
 * edge_cases:
 * - The column end is its last array element, which is the visible top card.
 * - A one-card suffix is always a legal run because the domain loop has no pair to inspect.
 * - A missing card or an out-of-range start index is not a movable run.
 */
export function isMissMilliganRun(column: readonly MissMilliganTableauCard[], startIndex: number): boolean {
  if (startIndex < 0 || startIndex >= column.length) return false;
  for (let i = startIndex + 1; i < column.length; i++) {
    const previous = column[i - 1]?.card;
    const current = column[i]?.card;
    if (!previous || !current) return false;
    const previousIsRed = previous.design === 'HEART' || previous.design === 'DIAMOND';
    const currentIsRed = current.design === 'HEART' || current.design === 'DIAMOND';
    if (current.value !== previous.value - 1 || previousIsRed === currentIsRed) return false;
  }
  return true;
}
