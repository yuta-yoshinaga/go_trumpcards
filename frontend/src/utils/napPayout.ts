/**
 * Chips a Nap contract is worth: `make` goes to the declarer when the contract
 * is made, `fail` goes to **each opponent** when it is not.
 *
 * Nap (5 tricks) is the asymmetric one — 10 to make, 5 apiece to beat — which
 * is what makes declaring it a real decision. Every other contract moves its
 * own trick count either way. Mirrors `NapBidPayout` in internal/domain/Nap.go.
 */
export interface NapPayout {
  /** Chips the declarer gains on success. */
  make: number;
  /** Chips each opponent gains on failure. */
  fail: number;
}

/** Contract value that means "5 tricks" (Nap). */
const NAP_CONTRACT = 5;

/**
 * Payout for a contract, or null for Pass (which stakes nothing).
 *
 * @param contract - The bid value (0=Pass, otherwise the trick count).
 * @returns The chips at stake, or null when nothing is staked.
 */
export function napPayout(contract: number): NapPayout | null {
  if (contract <= 0) return null;
  if (contract === NAP_CONTRACT) return { make: 10, fail: 5 };
  return { make: contract, fail: contract };
}
