import type { Card, SpoonsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SpoonsPhase } from '../../types/phases';

/** 手札の枚数。4 枚揃えばスプーンを取りに行ける。 */
const HAND_SIZE = 4;

/**
 * Returns a frontend {@link HintResult} for Spoons, or null when no suggestion
 * is available.
 *
 * Spoons exposes no hint from the backend, so this is computed from the human's
 * own hand and the race state. There is not much to reason about — that is the
 * game — so the hint mostly makes sure the player does not miss the two moments
 * that decide a round: the grab window opening, and their own fourth card
 * arriving.
 */
export function getSpoonsHint(state: SpoonsResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.eliminated || human.hand.length === 0) return null;

  if (state.phase === SpoonsPhase.GRAB) {
    // 既に持っているなら急ぐことはない。窓が閉じていれば取れない。
    if (!state.grabWindowOpen || human.hasSpoon) return null;
    return { targetAction: 'grab', reason: 'frontendHint.spoonsGrabNow', confidence: 'strong' };
  }

  if (state.phase !== SpoonsPhase.PASS || !state.isHumanTurn) return null;

  // **4 枚揃ったら自分から始める。**待つ理由がない。
  const oddIdx = oddCardOut(human.hand);
  if (oddIdx < 0) {
    return human.hand.length === HAND_SIZE
      ? { targetAction: 'grab', reason: 'frontendHint.spoonsFourOfAKind', confidence: 'strong' }
      : null;
  }
  return { targetAction: `hand-${oddIdx}`, reason: 'frontendHint.spoonsPassOdd', confidence: 'moderate' };
}

/**
 * 最も枚数の多いランクから外れた札のうち、最初のものの位置を返す。全部同じ
 * ランクなら -1。
 *
 * 集計を数え直さず Map の要素を回すのは、`counts.get(...) ?? 0` の `?? 0` 側が
 * どんな入力でも通らない分岐として残るため。
 */
function oddCardOut(hand: Card[]): number {
  const counts = new Map<number, number>();
  for (const c of hand) counts.set(c.value, (counts.get(c.value) ?? 0) + 1);

  let keep = hand[0].value;
  let best = 0;
  for (const [value, n] of counts) {
    if (n > best) {
      best = n;
      keep = value;
    }
  }
  return hand.findIndex((c) => c.value !== keep);
}
