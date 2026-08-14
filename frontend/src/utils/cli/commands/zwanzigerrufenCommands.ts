import type { zwanzigerrufenApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ZwanzigerrufenArgs = Parameters<typeof zwanzigerrufenApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
  'd',
  'discard',
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

/** Parse a Zwanzigerrufen CLI command into API exec arguments. */
export function parseZwanzigerrufenCommand(input: string): CliParseResult<ZwanzigerrufenArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase() ?? '';
      // **trischaken は宣言できない。** 全員パスの結果としてしか成立しない契約。
      if (arg !== 'rufer' && arg !== 'solo') return { error: 'Usage: bid rufer|solo' };
      return { args: ['bid', { bid: arg }] };
    }
    case 'pass':
      return { args: ['pass'] };
    case 'd':
    case 'discard': {
      const indices: number[] = [];
      for (let i = 0; i < 6; i++) {
        const parsed = parseIntArg(args, i);
        if ('error' in parsed) return { error: 'Usage: discard <i1> <i2> <i3> <i4> <i5> <i6>' };
        indices.push(parsed.value);
      }
      return { args: ['discard', { cardIndices: indices }] };
    }
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

/** Help text for Zwanzigerrufen (ツヴァンツィガールーフェン) CLI mode. */
export const ZWANZIGERRUFEN_HELP: string[] = [
  'bid rufer|solo                   - Declare a contract (Trischaken only arises when everyone passes)',
  'pass                             - Pass the auction',
  'discard <i1> ... <i6>            - Bury six cards (kings and trull cards are refused)',
  'p <idx>                          - Play the card at hand position idx',
  'n / next                         - Go to the next trick',
  'nr / nextround                   - Go to the next deal',
  'h / hint                         - Show a hint',
  'r / reset                        - Start a new match',
];
