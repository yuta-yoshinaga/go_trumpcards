import type { montebankApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MonteBankArgs = Parameters<typeof montebankApi.exec>;

const VALID_COMMANDS = ['bet', 'b', 'next', 'hint', 'log', 'r', 'reset', 'help', '?'];

/** The layout is always four cards. */
const LAYOUT_SIZE = 4;

const BET_USAGE = 'Usage: bet <1-4> <amount>';

/** CLI help text for Monte Bank. */
export const MONTEBANK_CLI_HELP = [
  'bet <n> <amount> back layout card n (1-4) and turn the gate',
  'next             deal the next layout',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Monte Bank CLI command into API exec arguments. */
export function parseMonteBankCommand(input: string): CliParseResult<MonteBankArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      // **画面は 1 始まり、ワイヤは 0 始まり。** 変換はここ 1 か所だけ。
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: BET_USAGE };
      if (idx.value < 1 || idx.value > LAYOUT_SIZE) return { error: BET_USAGE };
      const bet = parseIntArg(args, 1);
      if ('error' in bet) return { error: BET_USAGE };
      return { args: ['bet', { idx: idx.value - 1, bet: bet.value }] };
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
