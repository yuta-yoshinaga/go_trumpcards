import type { pokerApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PokerArgs = Parameters<typeof pokerApi.exec>;

const VALID_COMMANDS = [
  'e',
  'exchange',
  's',
  'stand',
  'b',
  'bet',
  'c',
  'call',
  'ra',
  'raise',
  'f',
  'fold',
  'ck',
  'check',
  'a',
  'allin',
  'o',
  'odds',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Poker CLI command into API exec arguments. */
export function parsePokerCommand(input: string): CliParseResult<PokerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'e':
    case 'exchange': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: `Invalid indices: ${parsed.error}` };
      return { args: ['exchange', parsed.values] };
    }
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'b':
    case 'bet': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: b <amount>' };
      return { args: ['bet', undefined, parsed.value] };
    }
    case 'c':
    case 'call':
      return { args: ['call'] };
    case 'ra':
    case 'raise': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: ra <amount>' };
      return { args: ['raise', undefined, parsed.value] };
    }
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'ck':
    case 'check':
      return { args: ['check'] };
    case 'a':
    case 'allin':
      return { args: ['allin'] };
    case 'o':
    case 'odds': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: `Invalid indices: ${parsed.error}` };
      return { args: ['odds', parsed.values] };
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

/** Help text for Poker CLI mode. */
export const POKER_HELP: string[] = [
  'e <idx...>  - Exchange cards (e.g., e 0 2 4)',
  's/stand     - Keep hand',
  'b <amount>  - Place bet',
  'c/call      - Call current bet',
  'ra <amount> - Raise',
  'f/fold      - Fold hand',
  'ck/check    - Check',
  'a/allin     - All-in',
  'o <idx...>  - Calculate odds',
  'r/reset     - Reset game',
];
