import type { gaigelApi } from '../../../api/gameApi';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by gaigelApi.exec. */
export type GaigelCliArgs = Parameters<typeof gaigelApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'm', 'marriage', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'r', 'reset'];

/** Parse a CLI command into arguments accepted by gaigelApi.exec. */
export function parseGaigelCommand(input: string): CliParseResult<GaigelCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play':
    case 'm':
    case 'marriage': {
      const idx = Number(args[0]);
      if (!Number.isInteger(idx) || idx < 0 || idx > 4) {
        return { error: cmd === 'p' || cmd === 'play' ? 'Usage: p <idx> (0-4)' : 'Usage: m <idx> (0-4)' };
      }
      return { args: [cmd === 'p' || cmd === 'play' ? 'play' : 'marriage', undefined, idx] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
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

/** Help text shown in the Gaigel CLI terminal. */
export const GAIGEL_HELP = [
  'p <idx>          - Play a card from your hand (0-4)',
  'm <idx>          - Declare a marriage with a card (0-4)',
  'n / next         - Continue after a trick',
  'nr / nextround   - Start the next round',
  'h / hint         - Request a hint',
  'r / reset        - Reset game',
];
