import type { Card } from '../types/card';
import { TrexContract } from '../types/phases';
import goldenPenalties from './__fixtures__/trexPenalties.golden.json';

/**
 * Trex の失点定数。
 * 値は internal/domain/Trex.go の定数が正。
 * TestTrexPenalties_GoldenValues (Go 側) が定数と JSON の一致を検査するので、
 * 片方だけ変えるともう片方が落ちる。
 */
export const TREX_PENALTIES: {
  readonly kingOfHearts: number;
  readonly diamond: number;
  readonly queen: number;
} = goldenPenalties;

/**
 * Returns the penalty points for taking this card under the contract in play,
 * mirroring `TrexCardPenalty` in `internal/domain/Trex.go`:
 *
 * - King of Hearts — the ♥K alone costs -75.
 * - Diamonds — every diamond costs -10.
 * - Queens — every queen costs -25.
 * - Tricks / Trix / none — no individual card costs anything (returns 0).
 *
 * @param card - The card to test.
 * @param contract - The contract in play, from `TrexResponse.contract`.
 * @returns The penalty points (negative number) or 0 if not a penalty card.
 */
export function trexCardPenalty(card: Card | null | undefined, contract: number): number {
  if (!card) return 0;
  switch (contract) {
    case TrexContract.KING_OF_HEARTS:
      return card.design === 'HEART' && card.value === 13 ? TREX_PENALTIES.kingOfHearts : 0;
    case TrexContract.DIAMONDS:
      return card.design === 'DIAMOND' ? TREX_PENALTIES.diamond : 0;
    case TrexContract.QUEENS:
      return card.value === 12 ? TREX_PENALTIES.queen : 0;
    default:
      return 0;
  }
}

/**
 * Whether this card is a penalty card under the contract in play.
 * Implemented via `trexCardPenalty(card, contract) !== 0`.
 *
 * **Five contracts rotate within one kingdom**, so which cards are dangerous
 * changes deal to deal and cannot be learned once (#4911).
 * @param card - The card to test.
 * @param contract - The contract in play, from `TrexResponse.contract`.
 * @returns Whether taking this card costs points.
 */
export function trexIsPenaltyCard(card: Card | null | undefined, contract: number): boolean {
  return trexCardPenalty(card, contract) !== 0;
}
