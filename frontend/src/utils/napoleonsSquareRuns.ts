import type { NapoleonsSquareTableauCard } from '../types/games/napoleonssquare';

/**
 * Returns whether the cards from `startIndex` through the top of a tableau
 * column form a same-suit descending run.
 *
 * domain_rule_quoted:
 * 	func napoleonsSquareIsRun(cards []*NapoleonsSquareTableauCard) bool {
 * 		for i := 1; i < len(cards); i++ {
 * 			prev, cur := cards[i-1].Card, cards[i].Card
 * 			if prev == nil || cur == nil { return false }
 * 			if prev.GetDesign() != cur.GetDesign() || cur.GetValue() != prev.GetValue()-1 {
 * 				return false
 * 			}
 * 		}
 * 		return true
 * 	}
 *
 * edge_cases:
 * - The column end is its last array element, which is the visible top card.
 * - A one-card suffix is always a legal run because the domain loop has no pair to inspect.
 * - A missing card or an out-of-range start index is not a movable run.
 */
export function isNapoleonsSquareRun(column: readonly NapoleonsSquareTableauCard[], startIndex: number): boolean {
  if (startIndex < 0 || startIndex >= column.length) return false;
  for (let i = startIndex + 1; i < column.length; i++) {
    const previous = column[i - 1]?.card;
    const current = column[i]?.card;
    if (!previous || !current || previous.design !== current.design || current.value !== previous.value - 1) {
      return false;
    }
  }
  return true;
}
