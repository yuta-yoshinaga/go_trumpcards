import type { ramsApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RamsArgs = Parameters<typeof ramsApi.exec>;

const VALID_COMMANDS = [
  'in',
  'play',
  'out',
  'pass',
  'c',
  'card',
  'n',
  'next',
  'h',
  'hint',
  'g',
  'giveup',
  'r',
  'reset',
  'log',
  'l',
];

/** Parse a Rams CLI command into API exec arguments (indices are 0-based). */
export function parseRamsCommand(input: string): CliParseResult<RamsArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'in':
    case 'play':
      return { args: ['in'] as RamsArgs };
    case 'out':
    case 'pass':
      return { args: ['out'] as RamsArgs };
    case 'c':
    case 'card': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: c <cardIdx>' };
      return { args: ['card', idx.value] as RamsArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as RamsArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as RamsArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as RamsArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as RamsArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as RamsArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Rams CLI mode. */
export const RAMS_HELP: string[] = [
  'in           - Play this round',
  'out          - Drop out of this round',
  'c <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
