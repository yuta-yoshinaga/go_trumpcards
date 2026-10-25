import type { ultimatetexasholdemApi } from '../../../api/gameApi';
import { UltimateTexasHoldemPhase } from '../../../types/phases';
import {
  anteMax,
  canPlay,
  maxTrips,
  UTH_ANTE_MIN,
  UTH_ANTE_STEP,
  UTH_TRIPS_MIN,
  UTH_TRIPS_STEP,
  utHoldemBetBounds,
} from '../../utHoldemBet';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Page-derived constraints used to keep CLI actions aligned with enabled controls. */
export interface UltimateTexasHoldemCliConstraints {
  chips: number;
  anteBet?: number;
  boundsForAnte?: (ante: number, trips: number) => { maxTrips: number; valid: boolean };
  phase?: number;
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
      if (!Number.isInteger(ante) || !Number.isInteger(trips) || ante <= 0 || trips < UTH_TRIPS_MIN) {
        return { error: 'Usage: bet <ante> [trips]' };
      }
      if (ante < UTH_ANTE_MIN || ante % UTH_ANTE_STEP !== 0 || trips % UTH_TRIPS_STEP !== 0)
        return {
          error: `Bet is outside the allowed range (ante minimum ${UTH_ANTE_MIN}, step ${UTH_ANTE_STEP}; trips minimum ${UTH_TRIPS_MIN}, step ${UTH_TRIPS_STEP}).`,
        };
      const bounds =
        constraints?.boundsForAnte?.(ante, trips) ??
        (constraints ? utHoldemBetBounds(ante, trips, constraints.chips) : undefined);
      if (
        constraints &&
        (ante > anteMax(constraints.chips) ||
          trips > (bounds?.maxTrips ?? maxTrips(ante, constraints.chips)) ||
          bounds?.valid === false)
      )
        return {
          error: `Bet is outside the allowed range (ante ${UTH_ANTE_MIN}–${anteMax(constraints.chips)}, trips ${UTH_TRIPS_MIN}–${bounds?.maxTrips ?? maxTrips(ante, constraints.chips)}, step ${UTH_TRIPS_STEP}).`,
        };
      if (constraints?.phase !== undefined && constraints.phase !== UltimateTexasHoldemPhase.BET)
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
      if (
        constraints?.phase !== undefined &&
        constraints.phase !== UltimateTexasHoldemPhase.PRE_FLOP &&
        constraints.phase !== UltimateTexasHoldemPhase.FLOP
      )
        return { error: 'Check is unavailable in this phase.' };
      return { args: ['check'] };
    case 'f':
    case 'fold':
      if (constraints?.phase !== undefined && constraints.phase !== UltimateTexasHoldemPhase.RIVER)
        return { error: 'Fold is unavailable in this phase.' };
      return { args: ['fold'] };
    case 'r':
    case 'reset':
      if (constraints?.phase !== undefined && constraints.phase !== UltimateTexasHoldemPhase.END)
        return { error: 'Reset is unavailable in this phase.' };
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

function playAllowed(multiplier: number, constraints: UltimateTexasHoldemCliConstraints): boolean {
  const phase =
    multiplier >= 3
      ? UltimateTexasHoldemPhase.PRE_FLOP
      : multiplier === 2
        ? UltimateTexasHoldemPhase.FLOP
        : UltimateTexasHoldemPhase.RIVER;
  return constraints.phase === phase && canPlay(multiplier, constraints.chips, constraints.anteBet ?? 0);
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
