import type { VideoPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getVideoPokerBaseHint } from './videoPokerBaseHint';

/** Returns a frontend HintResult for Joker Poker, or null if no suggestion available. */
export function getJokerPokerHint(state: VideoPokerResponse): HintResult | null {
  return getVideoPokerBaseHint(state, (card) => card.design === 'JOKER');
}
