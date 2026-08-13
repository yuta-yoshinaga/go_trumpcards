import type { SakuraResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SakuraPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Sakura, or null when there is
 * nothing to advise.
 *
 * **Which card to play comes from the server**, through `hint`. The page does
 * not re-derive the month matching: the server already publishes the legal
 * field cards per hand index in `captureOptions`, and a second opinion here
 * could disagree with the one that actually resolves the capture.
 */
export function getSakuraHint(state: SakuraResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === SakuraPhase.ROUND_END) {
    return { targetAction: 'next', reason: 'frontendHint.sakuraRoundIsOver', confidence: 'strong' };
  }
  if (!state.isHumanTurn) return null;

  // 合わせられる手札があるなら、捨てるより取るほうが必ず点になる。
  const hasCapture = Object.values(state.captureOptions ?? {}).some((opts) => opts.length > 0);
  if (hasCapture) {
    return { targetAction: 'play', reason: 'frontendHint.sakuraTakeTheMatch', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'frontendHint.sakuraDiscardCheapest', confidence: 'moderate' };
}
