import type { cincinnatiApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CincinnatiArgs = Parameters<typeof cincinnatiApi.exec>;

const VALID_COMMANDS = [
  'fold',
  'f',
  'check',
  'k',
  'call',
  'c',
  'bet',
  'b',
  'raise',
  'next',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const AMOUNT_USAGE = 'Usage: bet <amount> / raise <amount>';

/** CLI help text for Cincinnati. */
export const CINCINNATI_CLI_HELP = [
  'fold (f)         fold this hand',
  'check (k)        pass without betting',
  'call (c)         match the current bet',
  'bet <amount>     bet when nobody has',
  'raise <amount>   raise (three per round)',
  'next             deal the next hand',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Cincinnati CLI command into API exec arguments. */
export function parseCincinnatiCommand(input: string): CliParseResult<CincinnatiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'k':
    case 'check':
      return { args: ['check'] };
    case 'c':
    case 'call':
      return { args: ['call'] };
    case 'b':
    case 'bet':
    case 'raise': {
      // **額が要る手だけ引数を取る。** 降りる・チェック・コールは取らない。
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: AMOUNT_USAGE };
      return { args: [cmd === 'raise' ? 'raise' : 'bet', { amount: amount.value }] };
    }
    case 'next':
      return { args: ['next'] };
    case 'hint':
      return { args: ['hint'] };
    case 'log':
      return { args: ['log'] };
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
