import type { trogguApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TrogguArgs = Parameters<typeof trogguApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** The four contracts, by the exact name the API accepts. */
const CONTRACTS = ['trois', 'solo', 'piccolo', 'misere'] as const;

/** Parse a Troggu (トロッグ) CLI command into API exec arguments. */
export function parseTrogguCommand(input: string): CliParseResult<TrogguArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = (args[0]?.toLowerCase() ?? '') as (typeof CONTRACTS)[number];
      // **pass は入札ではない。** 別コマンドなので、契約名としては受け付けない。
      if (!CONTRACTS.includes(arg)) return { error: 'Usage: bid trois|solo|piccolo|misere' };
      return { args: ['bid', { bid: arg }] };
    }
    case 'pass':
      return { args: ['pass'] };
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
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
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Troggu (トロッグ) CLI mode. */
export const TROGGU_HELP: string[] = [
  'bid trois                        - Take three tricks',
  'bid solo                         - Take more than half the card points',
  'bid piccolo                      - Take exactly one trick',
  'bid misere                       - Take no tricks at all',
  'pass                             - Pass the auction (all pass throws the deal in)',
  'p <idx>                          - Play the card at hand position idx',
  'n / next                         - Go to the next trick',
  'nr / nextround                   - Go to the next deal',
  'h / hint                         - Show a hint',
  'r / reset                        - Start a new match',
];
