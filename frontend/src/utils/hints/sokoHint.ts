import type { FiveCardStudResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getFiveCardStudHint } from './fivecardstudHint';

/**
 * Returns a Soko frontend hint.
 *
 * Soko's betting decisions are Five Card Stud's — the streets, the bring-in and
 * the pot odds are identical — so the stud hint applies unchanged. What differs
 * is the showdown ranking, and the server has already resolved that into
 * `handName`/`handRank` before the page sees it, so there is nothing extra for
 * the frontend to compute. This delegates rather than forking a near-identical
 * copy that would drift.
 */
export function getSokoHint(state: FiveCardStudResponse): HintResult | null {
  return getFiveCardStudHint(state);
}
