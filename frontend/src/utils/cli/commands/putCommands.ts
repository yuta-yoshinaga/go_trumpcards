import type { putApi } from '../../../api/games/put';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by putApi.exec. */
export type PutCliArgs = Parameters<typeof putApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'put', 'accept', 'decline', 'n', 'next', 'h', 'hint', 'r', 'reset'];

/** Parse a Put CLI command into putApi.exec arguments. */
export function parsePutCommand(input: string): CliParseResult<PutCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number(args[0]);
      if (!Number.isInteger(idx) || idx < 0 || idx > 2) return { error: 'Usage: p <idx> (0-2)' };
      return { args: ['play', idx] };
    }
    case 'put':
      return { args: ['put'] };
    case 'accept':
      return { args: ['accept'] };
    case 'decline':
      return { args: ['decline'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Put CLI terminal. */
export const PUT_HELP = [
  'p <idx>          - Play a card from your hand (0-2)',
  'put              - Declare Put',
  'accept           - Accept a Put call',
  'decline          - Decline a Put call',
  'n / next         - Advance to the next trick or hand',
  'h / hint         - Show a hint',
  'r / reset        - Reset the game',
];
