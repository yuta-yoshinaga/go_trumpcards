import type { Card } from '../types/card';
import { TrexContract } from '../types/phases';

/**
 * Whether this card is a penalty card under the contract in play, mirroring the
 * `switch t.contract` in `Trex.cardPenalty` (`internal/domain/Trex.go`):
 *
 * - King of Hearts — the ♥K alone.
 * - Diamonds — every diamond.
 * - Queens — every queen, of any suit.
 * - Tricks / Trix / none — no individual card costs anything; the trick itself
 *   does, so nothing is marked.
 *
 * **Five contracts rotate within one kingdom**, so which cards are dangerous
 * changes deal to deal and cannot be learned once (#4911).
 * @param card - The card to test.
 * @param contract - The contract in play, from `TrexResponse.contract`.
 * @returns Whether taking this card costs points.
 */
export function trexIsPenaltyCard(card: Card | null | undefined, contract: number): boolean {
  if (!card) return false;
  switch (contract) {
    case TrexContract.KING_OF_HEARTS:
      return card.design === 'HEART' && card.value === 13;
    case TrexContract.DIAMONDS:
      return card.design === 'DIAMOND';
    case TrexContract.QUEENS:
      return card.value === 12;
    default:
      return false;
  }
}
