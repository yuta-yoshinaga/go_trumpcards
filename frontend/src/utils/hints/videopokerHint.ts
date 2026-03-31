import type { VideoPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getVideoPokerBaseHint } from './videoPokerBaseHint';

/** Returns a frontend HintResult for Video Poker (Jacks or Better), or null if no suggestion available. */
export function getVideoPokerHint(state: VideoPokerResponse): HintResult | null {
  return getVideoPokerBaseHint(state, () => false);
}
