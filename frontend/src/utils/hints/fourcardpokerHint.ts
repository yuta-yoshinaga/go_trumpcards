import type { FourCardPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Four Card Poker hint stub: returns `null` because the optimal Play vs Fold
 * decision in 4CP is highly dependent on the dealer's face-up card and the
 * paytable variance — surfacing an opinionated single hint would mislead
 * beginners more than help. Reserved for a future strategy engine.
 */
export function getFourCardPokerHint(_state: FourCardPokerResponse): HintResult | null {
  return null;
}
