import type { Card } from '../../types/card';

/**
 * The shallow "is this card material" test the rummy-family hints share, and the
 * discard choice built on it.
 *
 * Four hints had grown their own copy — Chinchón, Kalooki, Contract Rummy and
 * Three Thirteen — and the reviewer on #4639 flagged it before a fifth appeared.
 * Extracting it is not just deduplication: **each copy had already drifted**, and
 * the drift was invisible because every copy looked plausible on its own.
 *
 * - Contract Rummy and Kalooki treated the ace as adjacent only to the two, while
 *   both domains accept a high ace (J-Q-K-A). Fixed separately in #4646.
 * - Chinchón plays a 40-card deck where the seven and the jack are neighbours,
 *   so raw value arithmetic is wrong there. That copy already knew; the others
 *   would have inherited the bug had anyone copied it the other way.
 *
 * So the shared piece takes the adjacency rule as an argument rather than
 * assuming one. What is genuinely common is the *shape*: a card is material when
 * some other card shares its rank or sits next to it in suit, and the card to
 * throw is the heaviest that is material to nothing.
 *
 * **This is not a claim that a meld exists.** Melds need three cards; a pair is
 * only progress toward one. Where a hint asserts legality rather than progress —
 * Tonk's knock, which the server refuses above five points of deadwood — this
 * shape is the wrong tool and an exact search is required (see `tonkHint.ts`).
 */

/** Two ranks are adjacent for a run. Values are raw card values, ace = 1. */
export type Adjacency = (a: number, b: number) => boolean;

/** A(1) と K(13) の差。エースを高くも使えるゲームではランの端で隣り合う。 */
const ACE_TO_KING_GAP = 12;

/**
 * Ace-high adjacency: neighbouring ranks, plus the ace beside the king.
 *
 * Used by Kalooki (`Kalooki.go:822`), Three Thirteen (`ThreeThirteen.go:719`) and
 * Contract Rummy (`ContractRummy.go:930`), all of which accept A-2-3 and
 * J-Q-K-A while refusing to wrap around inside one run (K-A-2).
 */
export function aceHighAdjacent(a: number, b: number): boolean {
  const gap = Math.abs(a - b);
  return gap === 1 || gap === ACE_TO_KING_GAP;
}

/**
 * Whether `c` shares a rank with, or sits next to, some other card in `hand`.
 *
 * `hand` may include `c` itself; identity is not checked, so pass the rest of the
 * hand when asking about a specific position.
 */
export function isMaterial(c: Card, hand: Card[], adjacent: Adjacency): boolean {
  return hand.some((o) => o.value === c.value || (o.design === c.design && adjacent(o.value, c.value)));
}

/**
 * Index of the heaviest card that connects with nothing, or of the heaviest card
 * outright when every card connects.
 *
 * `skip` drops positions from consideration entirely — Three Thirteen uses it for
 * the round's wild rank, which should never be thrown. Returns -1 when `skip`
 * leaves nothing, which is a real state there (a hand of nothing but wilds).
 */
export function heaviestSpare(hand: Card[], adjacent: Adjacency, skip?: (c: Card) => boolean): number {
  const candidates = hand.map((_, i) => i).filter((i) => !skip?.(hand[i]));
  if (candidates.length === 0) return -1;

  const loose = candidates.filter(
    (i) =>
      !isMaterial(
        hand[i],
        hand.filter((_, j) => j !== i),
        adjacent,
      ),
  );
  // 全部が材料なら、材料の中から一番重いものを出すしかない。
  const pool = loose.length > 0 ? loose : candidates;
  let best = pool[0];
  for (const i of pool) {
    if (hand[i].value > hand[best].value) best = i;
  }
  return best;
}
