import type { brusquembilleApi } from '../../../api/games/brusquembille';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

export type BrusquembilleCliArgs = Parameters<typeof brusquembilleApi.exec>;
const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'h', 'hint', 'l', 'log', 'r', 'reset'];

export function parseBrusquembilleCommand(input: string): CliParseResult<BrusquembilleCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx) || idx < 0) return { error: 'Usage: p <idx>' };
      return { args: ['play', idx] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

export const BRUSQUEMBILLE_HELP = [
  'p <idx>          - Play a card by its hand index (0-based)',
  'n / next         - Complete the current trick',
  'h / hint         - Request a hint',
  'l / log          - Show action log',
  'r / reset        - Reset game',
];
