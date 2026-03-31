import type { ShortDeckResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getHoldemBaseHint } from './holdemBaseHint';

/** Returns a frontend HintResult for Short Deck Hold'em, or null if no suggestion available. */
export function getShortDeckHint(state: ShortDeckResponse): HintResult | null {
  return getHoldemBaseHint(state);
}
