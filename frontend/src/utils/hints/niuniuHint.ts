import type { NiuNiuResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { NiuNiuPhase } from '../../types/phases';

/**
 * ページが並べる賭け額 (sync: `BET_OPTIONS`, `frontend/src/pages/NiuNiuPage.tsx:30`)。
 * ここに無い額は押せないので、勧める額もこの中から選ぶ。
 */
const BET_OPTIONS: readonly number[] = [10, 50, 100, 500];

/**
 * Returns a frontend {@link HintResult} for Niu Niu, or null when no suggestion
 * is available.
 *
 * The only decision in this game is the stake, so that is all the hint talks
 * about. The thing worth saying is the one the page had to write a comment about
 * next to its `disabled` rule: **the exposure is the stake times
 * `maxMultiplier`, not the stake.** A banker's niu niu pays three times, so a
 * 500 stake needs 1,500 behind it. Reading a button as unaffordable when the
 * chips clearly exceed its face value is the confusion this addresses.
 *
 * It therefore recommends the largest option the stack actually covers, and says
 * so in terms of the multiplier when nothing is affordable — which is a real
 * state, not a hypothetical: with the multiplier at 3 a stack under 30 cannot
 * back even the smallest button, while still showing a positive chip count.
 */
export function getNiuNiuHint(state: NiuNiuResponse): HintResult | null {
  if (state.phase !== NiuNiuPhase.BET) return null;

  // **倍率が 0 で届くことがある**（配り直しの直後など）。割り算ではなく
  // 掛け算で比べているので 0 でも壊れないが、その場合は上限が無いことになり
  // 一番大きい額を勧めてしまう。倍率が立つまで黙る。
  if (state.maxMultiplier <= 0) return null;

  const affordable = BET_OPTIONS.filter((amount) => amount * state.maxMultiplier <= state.chips);

  if (affordable.length === 0) {
    return { targetAction: 'reset', reason: 'frontendHint.niuniuCannotCover', confidence: 'strong' };
  }

  const largest = affordable[affordable.length - 1];
  return {
    targetAction: `bet-${largest}`,
    reason:
      largest === BET_OPTIONS[BET_OPTIONS.length - 1] ? 'frontendHint.niuniuBetMax' : 'frontendHint.niuniuBetCapped',
    confidence: 'moderate',
  };
}
