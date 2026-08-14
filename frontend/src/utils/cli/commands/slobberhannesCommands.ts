import type { slobberhannesApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SlobberhannesArgs = Parameters<typeof slobberhannesApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Slobberhannes CLI command into API exec arguments (indices are 0-based). */
export function parseSlobberhannesCommand(input: string): CliParseResult<SlobberhannesArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as SlobberhannesArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as SlobberhannesArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as SlobberhannesArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as SlobberhannesArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as SlobberhannesArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as SlobberhannesArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Slobberhannes CLI mode. */
export const SLOBBERHANNES_HELP: string[] = [
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
