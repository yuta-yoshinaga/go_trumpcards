import type { freebetApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type FreeBetArgs = Parameters<typeof freebetApi.exec>;

const VALID_COMMANDS = [
  'bet',
  'b',
  'hit',
  'h',
  'stand',
  's',
  'freedouble',
  'fd',
  'freesplit',
  'fs',
  'next',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const BET_USAGE = 'Usage: bet <ante>';

/** CLI help text for Free Bet Blackjack. */
export const FREEBET_CLI_HELP = [
  'bet <ante>       place the ante and deal',
  'hit (h)          take a card',
  'stand (s)        stand pat',
  'freedouble (fd)  double on a hard 9-11; the house pays the raise',
  'freesplit (fs)   split a pair (not ten-valued); the house pays the raise',
  'next             deal the next hand',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Free Bet Blackjack CLI command into API exec arguments. */
export function parseFreeBetCommand(input: string): CliParseResult<FreeBetArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const ante = parseIntArg(args, 0);
      if ('error' in ante) return { error: BET_USAGE };
      return { args: ['bet', { ante: ante.value }] };
    }
    case 'h':
    case 'hit':
      return { args: ['hit'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
    // **無料の操作に額は無い。** 上乗せぶんはハウスが決めるので、引数を取ると
    // 「いくら払うのか」という存在しない問いを画面に持ち込むことになる。
    case 'fd':
    case 'freedouble':
      return { args: ['freedouble'] };
    case 'fs':
    case 'freesplit':
      return { args: ['freesplit'] };
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
