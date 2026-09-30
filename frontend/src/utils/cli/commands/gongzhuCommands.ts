import type { gongzhuApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GongZhuArgs = Parameters<typeof gongzhuApi.exec>;

const VALID_COMMANDS = ['expose', 'p', 'play', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse a Gong Zhu CLI command into API exec arguments. */
export function parseGongZhuCommand(input: string): CliParseResult<GongZhuArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'expose': {
      if (args.length === 0) return { args: ['expose', []] };
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: `Invalid indices: ${parsed.error}` };
      return { args: ['expose', parsed.values] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', undefined, parsed.value] };
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

/** Help text for Gong Zhu CLI mode. */
export const GONGZHU_HELP: string[] = [
  'expose [i...]    - Expose point cards (none = nothing)',
  'p <idx>          - Play a card',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
