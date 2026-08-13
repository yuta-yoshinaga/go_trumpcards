import type { baseballpokerApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BaseballPokerArgs = Parameters<typeof baseballpokerApi.exec>;

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
  'pay',
  'p',
  'buyfold',
  'next',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const AMOUNT_USAGE = 'Usage: bet <amount> / raise <amount>';

/** CLI help text for Baseball Poker. */
export const BASEBALLPOKER_CLI_HELP = [
  'fold (f)         fold this hand',
  'check (k)        pass without betting',
  'call (c)         match the current bet',
  'bet <amount>     bet when nobody has',
  'raise <amount>   raise (three per round)',
  'pay (p)          buy the pot and stay in',
  'buyfold          fold instead of buying',
  'next             deal the next hand',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Baseball Poker CLI command into API exec arguments. */
export function parseBaseballPokerCommand(input: string): CliParseResult<BaseballPokerArgs> {
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
      // **額が要る手だけ引数を取る。**
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: AMOUNT_USAGE };
      return { args: [cmd === 'raise' ? 'raise' : 'bet', { amount: amount.value }] };
    }
    // **買い増しの返事は別々のコマンド。** ひとつの語に数値を添えさせると、
    // 添え忘れが「0 番の返事」= 支払いに化ける。
    case 'p':
    case 'pay':
      return { args: ['pay'] };
    case 'buyfold':
      return { args: ['buyfold'] };
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
