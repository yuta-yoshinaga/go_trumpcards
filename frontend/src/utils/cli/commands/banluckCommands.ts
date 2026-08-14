import type { banluckApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BanLuckArgs = Parameters<typeof banluckApi.exec>;

const VALID_COMMANDS = ['bet', 'b', 'hit', 'h', 'stand', 's', 'next', 'hint', 'log', 'r', 'reset', 'help', '?'];

const BET_USAGE = 'Usage: bet <amount> (use 0 on your banker round)';

/** CLI help text for Ban Luck. */
export const BANLUCK_CLI_HELP = [
  'bet <amount>     place your stake and deal; use 0 on your banker round',
  'hit (h)          draw one card',
  'stand (s)        stand; the banker cannot stand below 15',
  'next             start the next round',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Ban Luck CLI command into API exec arguments. */
export function parseBanLuckCommand(input: string): CliParseResult<BanLuckArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      // **0 は「親なので賭けない」で、正当な入力。** 弾かないこと。
      const bet = parseIntArg(args, 0);
      if ('error' in bet) return { error: BET_USAGE };
      return { args: ['bet', { bet: bet.value }] };
    }
    case 'h':
    case 'hit':
      return { args: ['hit'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
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
