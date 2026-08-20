import type { highcardflushApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type HighCardFlushArgs = Parameters<typeof highcardflushApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'ra',
  'raise',
  '1',
  '2',
  '3',
  'f',
  'fold',
  'log',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a raise multiplier (1-3) into the API exec arguments. */
function raiseArgs(multiplier: number): CliParseResult<HighCardFlushArgs> {
  if (multiplier < 1 || multiplier > 3) return { error: 'Usage: raise <1|2|3>' };
  return { args: ['raise', undefined, undefined, undefined, multiplier] };
}

/** Parse a High Card Flush CLI command into API exec arguments. */
export function parseHighcardflushCommand(input: string): CliParseResult<HighCardFlushArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const usage = 'Usage: b <ante> [flushBonus] [straightFlush]';
      const ante = parseIntArg(args, 0);
      if ('error' in ante) return { error: usage };
      if (args.length >= 3) {
        const flushBonus = parseIntArg(args, 1);
        const straightFlush = parseIntArg(args, 2);
        if ('error' in flushBonus || 'error' in straightFlush) return { error: usage };
        return { args: ['bet', ante.value, flushBonus.value, straightFlush.value] };
      }
      if (args.length >= 2) {
        const flushBonus = parseIntArg(args, 1);
        if ('error' in flushBonus) return { error: usage };
        return { args: ['bet', ante.value, flushBonus.value] };
      }
      return { args: ['bet', ante.value] };
    }
    case '1':
      return raiseArgs(1);
    case '2':
      return raiseArgs(2);
    case '3':
      return raiseArgs(3);
    case 'ra':
    case 'raise': {
      const multiplier = parseIntArg(args, 0);
      if ('error' in multiplier) return { error: 'Usage: raise <1|2|3>' };
      return raiseArgs(multiplier.value);
    }
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'log':
      return { args: ['log'] };
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

/** Help text for High Card Flush CLI mode. */
export const HIGHCARDFLUSH_HELP: string[] = [
  'b <ante> [fb] [sf] - Ante bet (optional Flush Bonus / Straight Flush side bets)',
  '1 / 2 / 3          - Raise 1x / 2x / 3x the ante',
  'ra/raise <1|2|3>   - Raise by multiplier',
  'f/fold             - Fold',
  'log                - Show action log',
  'r/reset            - Reset game',
  'h/hint             - Get a hint',
];
