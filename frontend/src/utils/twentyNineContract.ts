/**
 * Points contested in one Twenty-Nine round: 7 per suit (J=3, 9=2, A=1, 10=1)
 * across four suits, plus 1 for taking the last trick. Mirrors
 * `TwentyNineTotalPoints` in internal/domain/TwentyNine.go.
 */
export const TWENTYNINE_TOTAL_POINTS = 29;

/** Whether the declaring team has made its contract, cannot, or still might. */
export type TwentyNineContractStatus = 'made' | 'failed' | 'needMore';

/** How the declaring team stands against its contract. */
export interface TwentyNineContractProgress {
  /** Declaring team index (0 or 1). */
  declarerTeam: number;
  /** Points the declaring team holds so far. */
  points: number;
  /** Points the contract requires. */
  contract: number;
  /** Points still needed; 0 once the contract is made. */
  remaining: number;
  /** made / failed / needMore. */
  status: TwentyNineContractStatus;
}

/**
 * Evaluate the declaring team's contract against the points already taken.
 *
 * "Cannot be made" is settled by the points **still in play** — the round total
 * minus what both teams have taken. Twenty-Nine tricks are not worth a fixed
 * amount (a jack alone is 3), so counting remaining tricks would be wrong.
 *
 * Mirrors `GetContractProgress` in internal/domain/TwentyNine.go.
 *
 * @param declarerIdx - Seat that won the bidding, or -1 when nobody did.
 * @param contract - The winning bid (0 when there is none).
 * @param roundTeamPoints - Points taken this round, indexed by team.
 * @returns The progress, or null when no contract exists yet.
 */
export function twentyNineContractProgress(
  declarerIdx: number,
  contract: number,
  roundTeamPoints: readonly number[],
): TwentyNineContractProgress | null {
  if (declarerIdx < 0 || contract <= 0) return null;
  const declarerTeam = declarerIdx % 2;
  const points = roundTeamPoints[declarerTeam] ?? 0;
  const taken = roundTeamPoints.reduce((sum, p) => sum + p, 0);
  const available = Math.max(0, TWENTYNINE_TOTAL_POINTS - taken);
  const status: TwentyNineContractStatus =
    points >= contract ? 'made' : points + available < contract ? 'failed' : 'needMore';
  return { declarerTeam, points, contract, remaining: Math.max(0, contract - points), status };
}
