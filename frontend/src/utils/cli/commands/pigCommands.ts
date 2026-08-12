import type { pigApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PigArgs = Parameters<typeof pigApi.exec>;

const VALID_COMMANDS = ['p', 'pass', 's', 'signal', 'n', 'next', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Pig CLI command into API exec arguments (indices are 0-based). */
export function parsePigCommand(input: string): CliParseResult<PigArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'pass': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['pass', idx.value] as PigArgs };
    }
    case 's':
    case 'signal':
      // **合図は別の行動。** カード指定が無いのは省略ではない。
      return { args: ['signal'] as PigArgs };
    case 'n':
    case 'next':
      return { args: ['next'] as PigArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as PigArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as PigArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as PigArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as PigArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Pig CLI mode. */
export const PIG_HELP: string[] = [
  'p <cardIdx> - Pass that card to your left (everyone passes at once)',
  's/signal    - Say you noticed the signal (only the last to react takes a letter)',
  'n/next      - Deal the next round',
  'h/hint      - Show a hint',
  'g/giveup    - Give up',
  'log         - Show the action log',
  'r/reset     - Reset game',
];
