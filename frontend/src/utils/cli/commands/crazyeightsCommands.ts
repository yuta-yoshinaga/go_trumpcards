import type { crazyeightsApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CrazyEightsArgs = Parameters<typeof crazyeightsApi.exec>;

const SUIT_MAP: Record<string, number> = {
  spade: 1,
  spades: 1,
  s: 1,
  clover: 2,
  clubs: 2,
  c: 2,
  heart: 3,
  hearts: 3,
  h: 3,
  diamond: 4,
  diamonds: 4,
  d: 4,
};

const VALID_COMMANDS = ['p', 'play', 'd', 'draw', 'suit', 'nr', 'nextround', 'r', 'reset', 'help', '?'];

/** Parse a Crazy Eights CLI command into API exec arguments. */
export function parseCrazyeightsCommand(input: string): CliParseResult<CrazyEightsArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', parsed.value] };
    }
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'suit': {
      if (args.length === 0) return { error: 'Usage: suit <suit> (spade/clover/heart/diamond)' };
      const suit = SUIT_MAP[args[0].toLowerCase()];
      if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
      return { args: ['suit', undefined, suit] };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Crazy Eights CLI mode. */
export const CRAZYEIGHTS_HELP: string[] = [
  'p <idx>     - Play a card',
  'd/draw      - Draw from pile',
  'suit <suit> - Choose suit (after 8)',
  'nr/nextround- Next round',
  'r/reset     - Reset game',
];
