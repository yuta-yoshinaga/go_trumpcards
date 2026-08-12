import type { rollingstoneApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RollingStoneArgs = Parameters<typeof rollingstoneApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'u', 'pickup', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Rolling Stone CLI command into API exec arguments (indices are 0-based). */
export function parseRollingStoneCommand(input: string): CliParseResult<RollingStoneArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as RollingStoneArgs };
    }
    case 'u':
    case 'pickup':
      // **引き取りは別のコマンド。** カード指定が無いのは省略ではない。
      return { args: ['pickup'] as RollingStoneArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as RollingStoneArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as RollingStoneArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as RollingStoneArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as RollingStoneArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Rolling Stone CLI mode. */
export const ROLLINGSTONE_HELP: string[] = [
  'p <cardIdx> - Play a card (following suit is compulsory)',
  'u/pickup    - Take the trick into your hand (only when you cannot follow)',
  'h/hint      - Show a hint',
  'g/giveup    - Give up',
  'log         - Show the action log',
  'r/reset     - Reset game',
];
