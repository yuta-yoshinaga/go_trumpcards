import type { badugiApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BadugiArgs = Parameters<typeof badugiApi.exec>;

const VALID_COMMANDS = [
  'f',
  'fold',
  'ck',
  'check',
  'c',
  'call',
  'b',
  'bet',
  'ra',
  'raise',
  'a',
  'allin',
  'ex',
  'exchange',
  'st',
  'stand',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Badugi CLI command into API exec arguments. */
export function parseBadugiCommand(input: string): CliParseResult<BadugiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'ck':
    case 'check':
      return { args: ['check'] };
    case 'c':
    case 'call':
      return { args: ['call'] };
    case 'a':
    case 'allin':
      return { args: ['allin'] };
    case 'b':
    case 'bet': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: b <amount>' };
      return { args: ['bet', undefined, parsed.value] };
    }
    case 'ra':
    case 'raise': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: ra <amount>' };
      return { args: ['raise', undefined, parsed.value] };
    }
    case 'ex':
    case 'exchange': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: ex <idx...> (e.g. ex 0 2 3)' };
      return { args: ['exchange', parsed.values] };
    }
    case 'st':
    case 'stand':
      return { args: ['stand'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Badugi CLI mode. */
export const BADUGI_HELP: string[] = [
  'f/fold         - Fold hand',
  'ck/check       - Check',
  'c/call         - Call current bet',
  'b <amount>     - Place bet',
  'ra <amount>    - Raise',
  'a/allin        - All-in',
  'ex <idx...>    - Exchange selected cards (e.g. ex 0 2)',
  'st/stand       - Stand pat (no exchange)',
  'r/reset        - Reset game',
  'h/hint         - Get a hint',
];
