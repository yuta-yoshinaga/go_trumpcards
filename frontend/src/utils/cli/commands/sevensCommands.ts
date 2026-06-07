import type { sevensApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import { STANDARD_SUIT_MAP as SUIT_MAP } from '../suitMaps';
import type { CliParseResult } from '../types';

type SevensArgs = Parameters<typeof sevensApi.exec>;

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
