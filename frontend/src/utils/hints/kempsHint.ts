import type { Card, KempsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { KempsPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Kemps, or null when no suggestion
 * is available.
 *
 * Kemps exposes no hint from the backend, so this is computed from what the
 * human can see — which is exactly the right basis here, because the game is
 * about reading a partner's secret signal. `partnerSignaling` and
 * `opponentSignaling` are human-only cues the server already decides to show,
 * so acting on them tells the player nothing they were not shown.
 */
export function getKempsHint(state: KempsResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.hand.length === 0) return null;

  if (state.phase === KempsPhase.DECLARE) {
    // **相方の合図が最優先。**取り損ねると相手にカウンターで持っていかれる。
    // 相手も動いていそうな場合でも、外すと −1 のカウンターより確実。
    if (state.partnerSignaling) {
      return { targetAction: 'kemps', reason: 'frontendHint.kempsPartnerSignalled', confidence: 'strong' };
    }
    if (state.opponentSignaling) {
      return { targetAction: 'counter', reason: 'frontendHint.kempsOpponentSignalled', confidence: 'moderate' };
    }
    return { targetAction: 'pass', reason: 'frontendHint.kempsNoSignal', confidence: 'moderate' };
  }

  if (state.phase !== KempsPhase.EXCHANGE || !state.isHumanTurn) return null;

  // 揃った後に交換すると崩れる。合図に専念する局面。
  if (human.hasFourOfAKind) {
    return { targetAction: 'signal', reason: 'frontendHint.kempsSignalNow', confidence: 'strong' };
  }

  const oddIdx = oddCardOut(human.hand);
  if (oddIdx < 0) return null;
  return { targetAction: `hand-${oddIdx}`, reason: 'frontendHint.kempsSwapOdd', confidence: 'moderate' };
}

/**
 * 最も枚数の多いランクから外れた札のうち、最初のものの位置を返す。
 *
 * 4 枚とも同じランクなら -1（揃っている＝交換する理由がない）。同数で並んだ
 * 場合は先に現れたランクを残す。どちらを残しても等価なので、選び方が安定して
 * いることのほうが大事。
 */
function oddCardOut(hand: Card[]): number {
  const counts = new Map<number, number>();
  for (const c of hand) counts.set(c.value, (counts.get(c.value) ?? 0) + 1);

  // Map の要素を数える。`counts.get(c.value) ?? 0` で引き直すと、必ず存在する
  // キーの `?? 0` 側が到達不能な分岐として残る。
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
