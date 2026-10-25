import type { ultimatetexasholdemApi } from '../../../api/gameApi';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Page-derived constraints used to keep CLI actions aligned with enabled controls. */
export interface UltimateTexasHoldemCliConstraints {
  maxAnte: number;
  maxTrips: number;
  boundsForAnte?: (ante: number, trips: number) => { maxTrips: number; valid: boolean };
  phase?: number;
  canPlay4x?: boolean;
  canPlay3x?: boolean;
  canPlay2x?: boolean;
  canPlay1x?: boolean;
}

/** Args accepted by ultimatetexasholdemApi.exec. */
export type UltimateTexasHoldemCliArgs = Parameters<typeof ultimatetexasholdemApi.exec>;

const VALID_COMMANDS = ['b', 'bet', '4', '3', '2', '1', 'play', 'c', 'check', 'f', 'fold', 'r', 'reset', 'help', '?'];

/** Parse a CLI command into Ultimate Texas Hold'em API arguments. */
export function parseUltimateTexasHoldemCommand(
  input: string,
  constraints?: UltimateTexasHoldemCliConstraints,
): CliParseResult<UltimateTexasHoldemCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'b':
    case 'bet': {
      const ante = Number(args[0]);
      const trips = args[1] === undefined ? 0 : Number(args[1]);
      if (!Number.isInteger(ante) || !Number.isInteger(trips) || ante <= 0 || trips < 0) {
        return { error: 'Usage: bet <ante> [trips]' };
      }
      if (ante < 10 || ante % 10 !== 0 || trips % 10 !== 0)
        return { error: 'Bet is outside the allowed range (ante minimum 10, step 10; trips minimum 0, step 10).' };
      const enteredBounds = constraints?.boundsForAnte?.(ante, trips);
      if (
        constraints &&
        (ante > constraints.maxAnte ||
          trips > (enteredBounds?.maxTrips ?? constraints.maxTrips) ||
          enteredBounds?.valid === false)
      )
        return {
          error: `Bet is outside the allowed range (ante 10–${constraints.maxAnte}, trips 0–${enteredBounds?.maxTrips ?? constraints.maxTrips}, step 10).`,
        };
      if (constraints?.phase !== undefined && constraints.phase !== 1)
        return { error: 'Bet is unavailable in this phase.' };
      return { args: ['bet', ante, trips] };
    }
    case 'play': {
      const multiplier = Number(args[0]);
      if (!Number.isInteger(multiplier) || multiplier < 1 || multiplier > 4) return { error: 'Usage: play <1|2|3|4>' };
      if (constraints?.phase !== undefined && !playAllowed(multiplier, constraints))
        return { error: 'Play is unavailable in this phase or with the current chips.' };
      return { args: ['play', undefined, undefined, multiplier] };
    }
    case '1':
    case '2':
    case '3':
    case '4':
      if (constraints?.phase !== undefined && !playAllowed(Number(cmd), constraints))
        return { error: 'Play is unavailable in this phase or with the current chips.' };
      return { args: ['play', undefined, undefined, Number(cmd)] };
    case 'c':
    case 'check':
      if (constraints?.phase !== undefined && constraints.phase !== 2 && constraints.phase !== 3)
        return { error: 'Check is unavailable in this phase.' };
      return { args: ['check'] };
    case 'f':
    case 'fold':
      if (constraints?.phase !== undefined && constraints.phase !== 4)
        return { error: 'Fold is unavailable in this phase.' };
      return { args: ['fold'] };
    case 'r':
    case 'reset':
      if (constraints?.phase !== undefined && constraints.phase !== 5)
        return { error: 'Reset is unavailable in this phase.' };
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

function playAllowed(multiplier: number, constraints: UltimateTexasHoldemCliConstraints): boolean {
  return multiplier === 4
    ? constraints.phase === 2 && constraints.canPlay4x !== false
    : multiplier === 3
      ? constraints.phase === 2 && constraints.canPlay3x !== false
      : multiplier === 2
        ? constraints.phase === 3 && constraints.canPlay2x !== false
        : constraints.phase === 4 && constraints.canPlay1x !== false;
}

/** Help text shown in the Ultimate Texas Hold'em terminal. */
export const ULTIMATE_TEXAS_HOLDEM_HELP = [
  'bet <ante> [trips] - Place Ante and optional Trips bets',
  '4 / 3 / 2 / 1      - Play the available multiplier',
  'c / check          - Check',
  'f / fold           - Fold',
  'h / hint           - Request a hint',
  'r / reset          - Reset game',
];
