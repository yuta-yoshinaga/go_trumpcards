/**
 * Minimum stake for Pontoon. Mirrors the backend `PontoonMinBet` constant in
 * `internal/domain/Pontoon.go`, so the buy control and the server agree on the
 * floor without either silently drifting.
 */
export const PONTOON_MIN_BET = 10;

/**
 * Largest stake `Pontoon.Buy` will accept: twice the current bet, never below
 * the floor. Mirrors the guard `extra < PontoonMinBet || extra > h.bet*2`.
 * @param currentBet - The stake on the hand being played.
 * @returns The highest legal buy.
 */
export function pontoonMaxBuy(currentBet: number): number {
  return Math.max(PONTOON_MIN_BET, currentBet * 2);
}

/**
 * Every legal buy stake for the current bet, in floor-sized steps.
 * @param currentBet - The stake on the hand being played.
 * @returns Ascending stakes from the floor to {@link pontoonMaxBuy}.
 */
export function pontoonBuyChoices(currentBet: number): number[] {
  const max = pontoonMaxBuy(currentBet);
  const out: number[] = [];
  for (let v = PONTOON_MIN_BET; v <= max; v += PONTOON_MIN_BET) out.push(v);
  return out;
}

/**
 * Bring a chosen stake inside the legal range for the current bet, so a choice
 * made on one hand cannot become illegal when the next hand's stake differs.
 * @param chosen - The player's chosen stake, or null to follow the current bet.
 * @param currentBet - The stake on the hand being played.
 * @returns A stake the server will accept.
 */
export function pontoonClampBuy(chosen: number | null, currentBet: number): number {
  return Math.min(Math.max(chosen ?? currentBet, PONTOON_MIN_BET), pontoonMaxBuy(currentBet));
}
