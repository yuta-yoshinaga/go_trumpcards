import type { sevensApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SevensArgs = Parameters<typeof sevensApi.exec>;

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

const VALID_COMMANDS = ['p', 'play', 'j', 'joker', 'pass', 'r', 'reset', 'help', '?'];

/** Parse a Sevens CLI command into API exec arguments. */
export function parseSevensCommand(input: string): CliParseResult<SevensArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', parsed.value] };
    }
    case 'pass':
      return { args: ['play', -1] };
    case 'j':
    case 'joker': {
      if (args.length < 2) return { error: 'Usage: j <suit> <value> (suit: spade/clover/heart/diamond)' };
      const suit = SUIT_MAP[args[0].toLowerCase()];
      if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
      const val = parseIntArg(args, 1);
      if ('error' in val) return { error: 'Usage: j <suit> <value>' };
      return { args: ['joker', -1, suit, val.value] };
    }
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

/** Help text for Sevens CLI mode. */
export const SEVENS_HELP: string[] = [
  'p <idx>     - Play a card',
  'pass        - Pass turn',
  'j <suit> <v>- Play joker (e.g., j spade 6)',
  'r/reset     - Reset game',
];
