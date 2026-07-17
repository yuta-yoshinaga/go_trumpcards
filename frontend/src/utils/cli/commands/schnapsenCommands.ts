import type { schnapsenApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SchnapsenArgs = Parameters<typeof schnapsenApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'm', 'marriage', 'n', 'next', 'h', 'hint', 'r', 'reset', 'log', 'l'];

/** Parse a Schnapsen CLI command into API exec arguments (indices are 0-based). */
export function parseSchnapsenCommand(input: string): CliParseResult<SchnapsenArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as SchnapsenArgs };
    }
    case 'm':
    case 'marriage': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: m <cardIdx>' };
      return { args: ['marriage', idx.value] as SchnapsenArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as SchnapsenArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as SchnapsenArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as SchnapsenArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as SchnapsenArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Schnapsen CLI mode. */
export const SCHNAPSEN_HELP: string[] = [
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'm <cardIdx>  - Declare a marriage with a card (marked M)',
  'n/next       - Advance to the next trick',
  'h/hint       - Show a hint',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
