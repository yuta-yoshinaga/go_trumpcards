import type { rummy500Api } from '../../../api/games/rummy500';
import { parseIntSlice, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Arguments accepted by the Rummy 500 API exec function. */
export type Rummy500CliArgs = Parameters<typeof rummy500Api.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'stock',
  't',
  'take',
  'm',
  'meld',
  'x',
  'discard',
  'lo',
  'layoff',
  'nr',
  'nextround',
  'r',
  'reset',
];

/** Parse a Rummy 500 terminal command into API exec arguments. */
export function parseRummy500Command(input: string): CliParseResult<Rummy500CliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'd':
    case 'draw':
    case 'stock':
      return { args: ['drawstock'] };
    case 't':
    case 'take': {
      const idx = Number(args[0]);
      if (args.length === 0 || !Number.isInteger(idx) || idx < 0) return { error: 'Usage: take <discard-index>' };
      return { args: ['drawdiscard', undefined, undefined, undefined, idx] };
    }
    case 'm':
    case 'meld': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed || parsed.values.length < 3 || parsed.values.some((i) => i < 0))
        return { error: 'Usage: meld <idx> <idx> <idx> [idx ...]' };
      return { args: ['meld', undefined, undefined, parsed.values] };
    }
    case 'x':
    case 'discard':
    case 'p':
    case 'play': {
      const idx = Number(args[0]);
      if (args.length === 0 || !Number.isInteger(idx) || idx < 0) return { error: 'Usage: discard <idx>' };
      return { args: ['discard', idx] };
    }
    case 'lo':
    case 'layoff': {
      const nums = args.map(Number);
      if (args.length !== 3 || nums.some((n) => !Number.isInteger(n) || n < 0))
        return { error: 'Usage: layoff <hand-idx> <owner> <meld-idx>' };
      return {
        args: [
          'layoff',
          undefined,
          undefined,
          undefined,
          undefined,
          { cardIndex: nums[0], meldOwner: nums[1], meldIdx: nums[2] },
        ],
      };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Rummy 500 terminal. */
export const RUMMY500_HELP = [
  'h / hint          - Show a hint',
  'd / draw          - Draw from stock',
  'take <idx>        - Draw from discard at index and above',
  'meld <idx> ...    - Meld selected hand cards (3 or more)',
  'discard <idx>     - Discard a hand card',
  'layoff <hand> <owner> <meld> - Lay off a card on a meld',
  'nr / nextround    - Start the next round',
  'r / reset         - Reset game',
];
