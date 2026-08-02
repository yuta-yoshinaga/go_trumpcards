import type { NiuNiuResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { NiuNiuPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Niu Niu, or null when no suggestion
 * is available.
 *
 * The only decision the player makes is the stake, and the one thing worth
 * saying about it is the cap. A banker's Niu Niu pays three times, so a stake
 * is legal only while the stack covers `stake * maxMultiplier` — the limit is
 * `chips / maxMultiplier`, not `chips`, which is exactly the trap a player who
 * reads their chip count walks into.
 *
 * `maxMultiplier` is read from the response rather than hardcoded; the type
 * says it is sent so the figure is not written down twice.
 */
export function getNiuNiuHint(state: NiuNiuResponse): HintResult | null {
  if (state.phase !== NiuNiuPhase.BET) return null;

  // 賭けられる上限。倍率が届いていなければ計算しない。
  if (state.maxMultiplier <= 0) return null;
  const cap = Math.floor(state.chips / state.maxMultiplier);

  return cap <= 0
    ? { targetAction: 'bet', reason: 'frontendHint.niuniuStackTooShort', confidence: 'strong' }
    : { targetAction: 'bet', reason: 'frontendHint.niuniuStakeCap', confidence: 'moderate' };
}
