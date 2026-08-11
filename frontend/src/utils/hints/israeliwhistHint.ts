import type { IsraeliWhistResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Israeli Whist, or null when no
 * suggestion is available.
 *
 * Neither bidding round names a card. An auction hint has to carry both the
 * number and the suit — either alone is not a legal bid — while a calling hint
 * carries only the number. Ducking once you already have your call is close to
 * automatic; chasing a trick you still need is a judgement call.
 */
export function getIsraeliWhistHint(state: IsraeliWhistResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  if (hint.cardIndex === undefined) {
    // **入札は数とスートの両方で 1 つの意思決定。** 片方だけでは動けない。
    if (hint.reason === 'israeliwhistAuctionBid') {
      return {
        targetAction: `auction-${hint.value}-${hint.suit}`,
        reason: `hint.${hint.reason}`,
        confidence: 'moderate',
      };
    }
    if (hint.reason === 'israeliwhistAuctionPass') {
      return { targetAction: 'auction-pass', reason: `hint.${hint.reason}`, confidence: 'moderate' };
    }
    return {
      targetAction: `bid-${hint.value}`,
      reason: `hint.${hint.reason}`,
      // ノルマは守らなければ宣言そのものが通らないので、迷う余地がない。
      confidence: hint.reason === 'israeliwhistMeetQuota' ? 'strong' : 'moderate',
    };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'israeliwhistDuck' ? 'strong' : 'moderate',
  };
}
