import type { reversisApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ReversisArgs = Parameters<typeof reversisApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Reversis CLI command into API exec arguments (indices are 0-based). */
export function parseReversisCommand(input: string): CliParseResult<ReversisArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as ReversisArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as ReversisArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as ReversisArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as ReversisArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as ReversisArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as ReversisArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Reversis CLI mode. */
export const REVERSIS_HELP: string[] = [
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
