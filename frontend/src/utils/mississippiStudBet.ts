import { MississippiStudPhase } from '../types/phases';

/** Play multipliers accepted on each Mississippi Stud betting street. */
export const MISSISSIPPI_STUD_PLAY_MULTIPLIERS = [1, 2, 3] as const;

/** Minimum ante accepted by Mississippi Stud. */
export const MISSISSIPPI_STUD_ANTE_MIN = 10;
/** Increment accepted for Mississippi Stud antes. */
export const MISSISSIPPI_STUD_ANTE_STEP = 10;

/** Returns the maximum ante allowed by the player's chip balance. */
export function anteMax(chips: number): number {
  return Math.floor(chips / 4);
}

/** Returns whether a street play multiplier fits within the player's chips. */
export function canPlay(multiplier: number, chips: number, anteAmount: number): boolean {
  return anteAmount * multiplier <= chips;
}

/** Returns whether the phase is one of Mississippi Stud's three betting streets. */
export function isBettingStreet(phase: number): boolean {
  return (
    phase === MississippiStudPhase.THIRD_STREET ||
    phase === MississippiStudPhase.FOURTH_STREET ||
    phase === MississippiStudPhase.FIFTH_STREET
  );
}

/** Returns the chips committed by a play multiplier. */
export function playAmount(multiplier: number, anteAmount: number): number {
  return anteAmount * multiplier;
}
