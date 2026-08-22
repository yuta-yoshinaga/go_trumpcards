import type { speculationApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SpeculationArgs = Parameters<typeof speculationApi.exec>;

const VALID_COMMANDS = [
  'f',
  'flip',
  'a',
  'accept',
  'd',
  'decline',
  'bid',
  'next',
  'h',
  'hint',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const BID_USAGE = 'Usage: bid <amount>';

/** CLI help text for Speculation. */
export const SPECULATION_CLI_HELP = [
  'flip (f)         turn up your next face-down card',
  'accept (a)       accept the standing offer (sell, or buy)',
  'decline (d)      decline the standing offer',
  'bid <amount>     raise the offer and buy',
  'next             start the next round',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Speculation CLI command into API exec arguments. */
export function parseSpeculationCommand(input: string): CliParseResult<SpeculationArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'f':
    case 'flip':
      return { args: ['flip'] };
    case 'a':
    case 'accept':
      return { args: ['accept'] };
    case 'd':
    case 'decline':
      return { args: ['decline'] };
    case 'bid': {
      // **額は省略できない。** 省略を 0 と読むと「0 で買う」が「断る」と
      // 区別できなくなる —— サーバも amount 無しの bid を拒否する。
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: BID_USAGE };
      if (amount.value <= 0) return { error: BID_USAGE };
      return { args: ['bid', { amount: amount.value }] };
    }
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'l':
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
