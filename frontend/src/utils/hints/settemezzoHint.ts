import type { SetteEMezzoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SetteEMezzoPhase } from '../../types/phases';

/**
 * ここまで離れていればもう 1 枚引く、という残り半点。
 *
 * 絵札は半点なので、引いた 1 枚が足す量は 1〜14 半点に散る。残り 4 半点
 * (＝2 点) より近い所で引くと、外す目のほうが多くなる。
 */
const DRAW_WHILE_BEHIND_BY = 4;

/**
 * Returns a frontend {@link HintResult} for Sette e Mezzo, or null when no
 * suggestion is available.
 *
 * The target is read from `targetHalves` rather than hardcoded — the server
 * sends it precisely so 7.5 is not written down on both sides — and every
 * branch checks the server's `canHit` / `canStand` before naming an action, so
 * the hint never points at a control the page has disabled.
 */
export function getSetteEMezzoHint(state: SetteEMezzoResponse): HintResult | null {
  if (state.phase !== SetteEMezzoPhase.PLAYER_TURN || state.activeSeat !== 0) return null;

  const seat = state.seats[0];
  const hand = seat?.hand;
  // 伏せられている手は totalHalves が 0 で届く。読んで助言すると嘘になる。
  if (!seat || seat.isCpu || !hand || hand.hidden) return null;

  const remaining = state.targetHalves - hand.totalHalves;

  // **ちょうど 7.5 が最優先。**マッタは手を止めるまで付け替えられる
  // (SetteEMezzo.go:355-371) ので、合わせ終えた後も canSetMatta は立ったまま。
  // マッタを先に見ると「7.5 に合わせましょう」と言い続ける (#4612 のレビュー指摘)。
  if (remaining === 0) {
    return state.canStand
      ? { targetAction: 'stand', reason: 'frontendHint.settemezzoExact', confidence: 'strong' }
      : null;
  }

  // まだ届いていないなら、マッタの値でどこまで行けるかがこの一手を決める。
  if (state.canSetMatta && hand.hasMatta) {
    return { targetAction: 'matta', reason: 'frontendHint.settemezzoSetMatta', confidence: 'strong' };
  }

  if (remaining >= DRAW_WHILE_BEHIND_BY) {
    return state.canHit ? { targetAction: 'hit', reason: 'frontendHint.settemezzoHitLow', confidence: 'strong' } : null;
  }

  return state.canStand
    ? { targetAction: 'stand', reason: 'frontendHint.settemezzoStandClose', confidence: 'moderate' }
    : null;
}
