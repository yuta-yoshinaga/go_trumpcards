import type { VideoPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getVideoPokerBaseHint } from './videoPokerBaseHint';

/** Returns a frontend HintResult for Deuces Wild, or null if no suggestion available. */
export function getDeucesWildHint(state: VideoPokerResponse): HintResult | null {
  return getVideoPokerBaseHint(state, (card) => card.value === 2);
}
