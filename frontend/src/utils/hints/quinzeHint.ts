import type { QuinzeResponse } from '../../types/card';
import { QuinzePhase } from '../../types/games/quinze';
import type { HintResult } from '../../types/hint';

/**
 * ここまで離れていればもう 1 枚引く、という残り整数点。
 *
 * 絵札は10点なので、引いた 1 枚が足す量は 1〜14 整数点に散る。残り 4 整数点
 * (＝2 点) より近い所で引くと、外す目のほうが多くなる。
 */
const DRAW_WHILE_BEHIND_BY = 4;

/**
 * Returns a frontend {@link HintResult} for Quinze, or null when no
 * suggestion is available.
 *
 * The target is read from `targetPoints` rather than hardcoded — the server
 * sends it precisely so 15 is not written down on both sides — and every
 * branch checks the server's `canHit` / `canStand` before naming an action, so
 * the hint never points at a control the page has disabled.
 */
export function getQuinzeHint(state: QuinzeResponse): HintResult | null {
  if (state.phase !== QuinzePhase.PLAYER_TURN || state.activeSeat !== 0) return null;

  const seat = state.seats[0];
  const hand = seat?.hand;
  // 伏せられている手は totalPoints が 0 で届く。読んで助言すると嘘になる。
  if (!seat || seat.isCpu || !hand || hand.hidden) return null;

  const remaining = state.targetPoints - hand.totalPoints;

  if (hand.totalPoints === state.targetPoints) {
    return state.canStand ? { targetAction: 'stand', reason: 'frontendHint.quinzeExact', confidence: 'strong' } : null;
  }

  if (remaining >= DRAW_WHILE_BEHIND_BY) {
    return state.canHit ? { targetAction: 'hit', reason: 'frontendHint.quinzeHitLow', confidence: 'strong' } : null;
  }

  return state.canStand
    ? { targetAction: 'stand', reason: 'frontendHint.quinzeStandClose', confidence: 'moderate' }
    : null;
}
