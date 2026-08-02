import type { BidWhistResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Bid Whist, or null when no
 * suggestion is available.
 *
 * The hint is computed by the Go backend (`BidWhist.GetHint`) and, since #4483,
 * arrives on every response rather than only the `hint` command's. This adapter
 * re-maps it into the frontend shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it.
 *
 * It used to be a stub returning null with a comment saying the reasoning lived
 * server-side — which was true, and stopped being a reason to return nothing
 * once the server started sending the answer.
 *
 * The hint carries **five mutually exclusive shapes**, one per decision point,
 * and each is checked with `!== undefined` because 0 is a legal value for every
 * one of them: bidding zero-direction, naming spades, and playing the first
 * card in hand.
 */
export function getBidWhistHint(state: BidWhistResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  if (hint.cardIndex !== undefined) {
    return { targetAction: `card-${hint.cardIndex}`, reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  if (hint.discardIndices !== undefined) {
    return { targetAction: 'discard', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  if (hint.trumpSuit !== undefined) {
    return { targetAction: 'trump', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  if (hint.pass === true) {
    return { targetAction: 'pass', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  if (hint.bidTricks !== undefined) {
    return { targetAction: 'bid', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }
  return null;
}
