import type { HoldemResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getHoldemBaseHint } from './holdemBaseHint';

/** Returns a frontend HintResult for Texas Hold'em, or null if no suggestion available. */
export function getHoldemHint(state: HoldemResponse): HintResult | null {
  return getHoldemBaseHint(state);
}
