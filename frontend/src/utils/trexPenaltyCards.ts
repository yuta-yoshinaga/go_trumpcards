import type { Card } from '../types/card';

/**
 * Trex contracts, matching `domain.TrexContract`. Each names a different set of
 * cards as the ones that cost points.
 */
export const TREX_CONTRACT = {
  KingOfHearts: 0,
  Diamonds: 1,
  Queens: 2,
  Tricks: 3,
  Trix: 4,
  None: 5,
} as const;

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
    case TREX_CONTRACT.KingOfHearts:
      return card.design === 'HEART' && card.value === 13;
    case TREX_CONTRACT.Diamonds:
      return card.design === 'DIAMOND';
    case TREX_CONTRACT.Queens:
      return card.value === 12;
    default:
      return false;
  }
}
