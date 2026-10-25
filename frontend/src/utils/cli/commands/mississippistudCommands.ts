import type { mississippiStudApi } from '../../../api/gameApi';
import type { MississippiStudResponse } from '../../../types/card';
import { MississippiStudPhase } from '../../../types/phases';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by the Mississippi Stud API. */
export type MississippiStudCliArgs = Parameters<typeof mississippiStudApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'p', 'play', 'f', 'fold', 'r', 'reset', 'help', '?'];

/** Parse Mississippi Stud terminal commands into API calls. */
export function getMississippiStudAnteMax(chips: number): number {
  return Math.floor(chips / 4);
}

export function parseMississippiStudCommand(
  input: string,
  state?: MississippiStudResponse,
): CliParseResult<MississippiStudCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = Number(args[0]);
      if (!args[0] || !Number.isInteger(amount) || amount <= 0) return { error: 'Usage: bet <amount>' };
      const max = state ? getMississippiStudAnteMax(state.chips) : Number.POSITIVE_INFINITY;
      if (amount < 10 || amount % 10 !== 0 || amount > max)
        return { error: `Ante must be within the range 10–${max} in steps of 10` };
      if (state && state.phase !== MississippiStudPhase.ANTE)
        return { error: 'Bet is only available during the ante phase' };
      return { args: ['bet', amount] };
    }
    case 'p':
    case 'play': {
      const multiplier = Number(args[0]);
      if (!args[0] || !Number.isInteger(multiplier) || multiplier < 1 || multiplier > 3) {
        return { error: 'Usage: play <1-3>' };
      }
      if (
        state &&
        (!(
          [
            MississippiStudPhase.THIRD_STREET,
            MississippiStudPhase.FOURTH_STREET,
            MississippiStudPhase.FIFTH_STREET,
          ] as number[]
        ).includes(state.phase) ||
          state.anteAmount * multiplier > state.chips)
      )
        return { error: 'That play is unavailable in the current phase or with the current chips' };
      return { args: ['play', undefined, multiplier] };
    }
    case 'f':
    case 'fold':
      if (
        state &&
        !(
          [
            MississippiStudPhase.THIRD_STREET,
            MississippiStudPhase.FOURTH_STREET,
            MississippiStudPhase.FIFTH_STREET,
          ] as number[]
        ).includes(state.phase)
      )
        return { error: 'Fold is only available during a betting street' };
      return { args: ['fold'] };
    case 'r':
    case 'reset':
      if (state && state.phase !== MississippiStudPhase.END)
        return { error: 'Reset is only available after the hand ends' };
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Mississippi Stud terminal. */
export const MISSISSIPPISTUD_HELP: string[] = [
  'b / bet <amount> - Place the ante',
  'p / play <1-3>   - Bet 1x, 2x, or 3x on this street',
  'f / fold         - Fold the hand',
  'r / reset        - Reset the game',
  'h / hint         - Get a local hint',
];
