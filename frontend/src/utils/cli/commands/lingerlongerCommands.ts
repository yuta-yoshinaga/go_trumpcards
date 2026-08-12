import type { lingerlongerApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type LingerLongerArgs = Parameters<typeof lingerlongerApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Linger Longer CLI command into API exec arguments (indices are 0-based). */
export function parseLingerLongerCommand(input: string): CliParseResult<LingerLongerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as LingerLongerArgs };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] as LingerLongerArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as LingerLongerArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as LingerLongerArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as LingerLongerArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/**
 * Help text for Linger Longer CLI mode.
 *
 * **There is no draw command** — winning a trick refills your hand by itself.
 */
export const LINGERLONGER_HELP: string[] = [
  'p <cardIdx> - Play a card (following suit is compulsory)',
  'h/hint      - Show a hint',
  'g/giveup    - Give up',
  'log         - Show the action log',
  'r/reset     - Reset game',
];
