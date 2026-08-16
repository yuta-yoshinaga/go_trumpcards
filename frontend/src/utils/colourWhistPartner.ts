import { ColourWhistContract } from '../types/phases';

/**
 * Whether the given contract has a partner at all.
 *
 * Mirrors `ColourWhistHasPartner` in `internal/domain/ColourWhistConfig.go`:
 * only Samen (the declarer calls a partner) and Troel (forced by the deal, the
 * holder of the fourth ace becomes the partner) are two-against-two. Alleen and
 * Miserie are solo contracts.
 *
 * This matters because `partnerIdx` is `-1` for both "there is no partner" and
 * "the partner exists but is not revealed yet". Keying the UI on `partnerIdx`
 * alone told players a hidden ally existed in solo contracts (#5773).
 */
export function colourWhistHasPartner(contract: number): boolean {
  return contract === ColourWhistContract.SAMEN || contract === ColourWhistContract.TROEL;
}
