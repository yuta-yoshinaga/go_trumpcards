import type { Card, KalookiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * フェーズ番号 (sync: internal/domain/Kalooki.go)。
 *
 * `KalookiPhase` は `types/phases.ts` に無く、`KalookiPage` がページ内で
 * 同じ表を持っている。ここから import すると循環するので、同期先を明記して
 * 持ち直している。
 */
const PHASE_DRAW = 0;
const PHASE_MELD = 1;

/**
 * Returns a frontend {@link HintResult} for Kalooki, or null when no suggestion
 * is available.
 *
 * `melds` on a player is what they have **laid down**, not what their hand could
 * make, so this does not try to prove a meld in hand. It uses the same shallow
 * "connects with something" test as Chinchón — same rank, or the neighbouring
 * rank in the same suit — which is enough for the two decisions taken every
 * turn without re-implementing the domain's meld detection.
 *
 * Before opening, the useful thing to say is different: the threshold has to be
 * met in one go, so the hint says to keep building rather than to throw the
 * heaviest card away.
 */
export function getKalookiHint(state: KalookiResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === PHASE_DRAW) {
    const top = state.discardTop;
    return top && connects(top, human.cards)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.kalookiTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.kalookiDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== PHASE_MELD) return null;

  // **開いていないうちは崩さない。**規定点を一度に満たす必要があるので、
  // 重い札こそ組の材料になる。
  if (!human.hasOpened) {
    return { targetAction: 'meld', reason: 'frontendHint.kalookiOpenFirst', confidence: 'moderate' };
  }

  const idx = heaviestLoose(human.cards);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.kalookiDiscardHeavy', confidence: 'moderate' };
}

/** 同じランクがあるか、同じスートで隣のランクがあるか。メルドの証明ではない。 */
function connects(c: Card, hand: Card[]): boolean {
  return hand.some((o) => o.value === c.value || (o.design === c.design && Math.abs(o.value - c.value) === 1));
}

/**
 * 繋がっていない札のうち一番重いものの位置。全部繋がっていれば一番重い札。
 *
 * 呼び出し側が手札の非空を確かめているので、必ず有効な位置を返す。
 */
function heaviestLoose(hand: Card[]): number {
  const loose: number[] = [];
  for (let i = 0; i < hand.length; i += 1) {
    const rest = hand.filter((_, j) => j !== i);
    if (!connects(hand[i], rest)) loose.push(i);
  }
  const pool = loose.length > 0 ? loose : hand.map((_, i) => i);
  let best = pool[0];
  for (const i of pool) {
    if (hand[i].value > hand[best].value) best = i;
  }
  return best;
}
