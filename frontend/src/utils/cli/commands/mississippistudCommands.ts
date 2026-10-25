import type { mississippiStudApi } from '../../../api/gameApi';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by the Mississippi Stud API. */
export type MississippiStudCliArgs = Parameters<typeof mississippiStudApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'p', 'play', 'f', 'fold', 'r', 'reset', 'help', '?'];

/** Parse Mississippi Stud terminal commands into API calls. */
export function parseMississippiStudCommand(input: string): CliParseResult<MississippiStudCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = Number(args[0]);
      if (!args[0] || !Number.isInteger(amount) || amount <= 0) return { error: 'Usage: bet <amount>' };
      return { args: ['bet', amount] };
    }
    case 'p':
    case 'play': {
      const multiplier = Number(args[0]);
      if (!args[0] || !Number.isInteger(multiplier) || multiplier < 1 || multiplier > 3) {
        return { error: 'Usage: play <1-3>' };
      }
      return { args: ['play', undefined, multiplier] };
    }
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'r':
    case 'reset':
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
];
