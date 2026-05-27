import type { PineappleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getPineappleHint } from './pineappleHint';

/** Returns a frontend HintResult for Irish Poker, or null if no suggestion
 * available. Delegates to the shared Pineapple hint logic. */
export function getIrishPokerHint(state: PineappleResponse): HintResult | null {
  return getPineappleHint(state);
}
