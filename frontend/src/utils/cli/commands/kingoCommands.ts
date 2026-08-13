import type { kingoApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type KingoArgs = Parameters<typeof kingoApi.exec>;

const VALID_COMMANDS = ['bet', 'b', 'deal', 'd', 'next', 'hint', 'log', 'r', 'reset', 'help', '?'];

const AMOUNT_USAGE = 'Usage: bet <amount>';

/** CLI help text for Kingo. */
export const KINGO_CLI_HELP = [
  'bet <amount> (b) bet as a child',
  'deal (d)         deal as the banker',
  'next             start the next round',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Kingo CLI command into API exec arguments. */
export function parseKingoCommand(input: string): CliParseResult<KingoArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      // **額が要るのは張りだけ。** 配るのに額は要らない。
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: AMOUNT_USAGE };
      return { args: ['bet', { amount: amount.value }] };
    }
    // **張ると配るは別のコマンド。** 親と子で求められる手が違う。
    case 'd':
    case 'deal':
      return { args: ['deal'] };
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
