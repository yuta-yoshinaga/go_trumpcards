import type { VideoPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getVideoPokerBaseHint } from './videoPokerBaseHint';

/** Returns a frontend HintResult for Deuces Wild, or null if no suggestion available. */
export function getDeucesWildHint(state: VideoPokerResponse): HintResult | null {
  // Deuces Wild pays nothing below three of a kind: its paytable has no pair
  // row at all, so no pair is ever worth holding for its own sake (#6301).
  return getVideoPokerBaseHint(state, (card) => card.value === 2, null);
}
