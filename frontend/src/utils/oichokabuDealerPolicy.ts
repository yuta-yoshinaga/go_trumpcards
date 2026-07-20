/**
 * Oicho-Kabu banker (dealer / 親) draw-policy disclosure.
 *
 * The banker follows a fixed house rule (see the domain constant
 * `OichoKabuBankerDrawThreshold` in `internal/domain/OichoKabu.go`): starting
 * from a two-card hand, it draws a third card only when its rank (目) is at or
 * below the threshold, and stands otherwise. Because both sides are always
 * dealt exactly two cards, a three-card banker hand means the banker drew, and
 * a two-card banker hand means it stood — in which case the revealed rank is
 * the same rank the banker used to make its decision.
 */

/**
 * Banker draw threshold — mirrors the domain constant
 * `OichoKabuBankerDrawThreshold` (the banker draws on rank 6 or below).
 */
export const OICHOKABU_BANKER_DRAW_THRESHOLD = 6;

/** An i18n key plus its interpolation params describing the banker's decision. */
export interface OichokabuDealerPolicy {
  /** i18n key under the `dealerPolicy` namespace. */
  i18nKey: string;
  /** Interpolation params for the key. */
  params: Record<string, number>;
}

/**
 * Derive the banker's draw-policy disclosure from the revealed result state.
 *
 * @param bankerHandCount - Number of cards in the banker's final hand (2 = stood, 3 = drew).
 * @param bankerRank - The banker's revealed rank (目). Only meaningful for the stood case.
 * @returns The i18n key + params explaining why the banker drew or stood.
 */
export function oichokabuDealerPolicy(bankerHandCount: number, bankerRank: number): OichokabuDealerPolicy {
  const drew = bankerHandCount > 2;
  return drew
    ? { i18nKey: 'dealerPolicy.drew', params: { threshold: OICHOKABU_BANKER_DRAW_THRESHOLD } }
    : { i18nKey: 'dealerPolicy.stood', params: { rank: bankerRank, threshold: OICHOKABU_BANKER_DRAW_THRESHOLD } };
}
