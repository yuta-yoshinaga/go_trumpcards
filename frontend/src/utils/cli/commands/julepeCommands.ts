import type { julepeApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type JulepeArgs = Parameters<typeof julepeApi.exec>;

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

/** Parse a Julepe CLI command into API exec arguments (indices are 0-based). */
export function parseJulepeCommand(input: string): CliParseResult<JulepeArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'in':
    case 'play':
      return { args: ['in'] as JulepeArgs };
    case 'out':
    case 'pass':
      return { args: ['out'] as JulepeArgs };
    case 'c':
    case 'card': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: c <cardIdx>' };
      return { args: ['card', idx.value] as JulepeArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as JulepeArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as JulepeArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as JulepeArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as JulepeArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as JulepeArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Julepe CLI mode. */
export const JULEPE_HELP: string[] = [
  'in           - Play this round',
  'out          - Drop out of this round',
  'c <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
