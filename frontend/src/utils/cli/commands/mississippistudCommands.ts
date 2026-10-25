import type { mississippiStudApi } from '../../../api/gameApi';
import type { MississippiStudResponse } from '../../../types/card';
import { MississippiStudPhase } from '../../../types/phases';
import {
  anteMax,
  canPlay,
  isBettingStreet,
  MISSISSIPPI_STUD_ANTE_MIN,
  MISSISSIPPI_STUD_ANTE_STEP,
  MISSISSIPPI_STUD_PLAY_MULTIPLIERS,
} from '../../mississippiStudBet';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by the Mississippi Stud API. */
export type MississippiStudCliArgs = Parameters<typeof mississippiStudApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'p', 'play', 'f', 'fold', 'r', 'reset', 'help', '?'];

/** Parse Mississippi Stud terminal commands into API calls. */
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
      const max = state ? anteMax(state.chips) : undefined;
      if (state && state.phase !== MississippiStudPhase.ANTE)
        return { error: 'Bet is only available during the ante phase' };
      if (
        amount < MISSISSIPPI_STUD_ANTE_MIN ||
        amount % MISSISSIPPI_STUD_ANTE_STEP !== 0 ||
        (max !== undefined && amount > max)
      ) {
        if (max !== undefined && max < MISSISSIPPI_STUD_ANTE_MIN)
          return { error: 'Ante cannot be placed with the current chips' };
        return {
          error:
            max === undefined
              ? `Ante must be at least ${MISSISSIPPI_STUD_ANTE_MIN} in steps of ${MISSISSIPPI_STUD_ANTE_STEP}`
              : `Ante must be within the range ${MISSISSIPPI_STUD_ANTE_MIN}–${max} in steps of ${MISSISSIPPI_STUD_ANTE_STEP}`,
        };
      }
      return { args: ['bet', amount] };
    }
    case 'p':
    case 'play': {
      const multiplier = Number(args[0]);
      if (
        !args[0] ||
        !Number.isInteger(multiplier) ||
        !MISSISSIPPI_STUD_PLAY_MULTIPLIERS.includes(multiplier as 1 | 2 | 3)
      ) {
        return {
          error: `Usage: play <${MISSISSIPPI_STUD_PLAY_MULTIPLIERS[0]}-${MISSISSIPPI_STUD_PLAY_MULTIPLIERS.at(-1)}>`,
        };
      }
      if (state && !isBettingStreet(state.phase)) return { error: 'Play is only available during a betting street' };
      if (state && !canPlay(multiplier, state.chips, state.anteAmount))
        return { error: 'That play is unaffordable with the current chips' };
      return { args: ['play', undefined, multiplier] };
    }
    case 'f':
    case 'fold':
      if (state && !isBettingStreet(state.phase)) return { error: 'Fold is only available during a betting street' };
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
