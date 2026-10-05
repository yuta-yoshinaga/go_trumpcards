import type { BaseballPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Baseball Poker, or null when there
 * is nothing to advise.
 *
 * The server evaluates the actual hand and supplies its recommendation.
 */
export function getBaseballpokerHint(state: BaseballPokerResponse): HintResult | null {
  const { serverHint } = state;
  if (!serverHint) return null;
  const reason = `frontendHint.baseball${serverHint.reason.charAt(0).toUpperCase()}${serverHint.reason.slice(1)}`;
  return {
    targetAction: serverHint.action,
    reason,
    confidence: serverHint.reason === 'seeAnotherCard' ? 'strong' : 'moderate',
  };
}

/** True when the card value is wild for this game state. */
export function isBaseballWild(state: BaseballPokerResponse, value: number): boolean {
  return state.wildValues.includes(value);
}
