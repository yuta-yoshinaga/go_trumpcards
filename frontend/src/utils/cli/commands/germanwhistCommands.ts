import type { germanwhistApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GermanWhistArgs = Parameters<typeof germanwhistApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a German Whist CLI command into API exec arguments (indices are 0-based). */
export function parseGermanWhistCommand(input: string): CliParseResult<GermanWhistArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as GermanWhistArgs };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] as GermanWhistArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as GermanWhistArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as GermanWhistArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as GermanWhistArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for German Whist CLI mode. */
export const GERMANWHIST_HELP: string[] = [
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
