import type { polignacApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PolignacArgs = Parameters<typeof polignacApi.exec>;

const VALID_COMMANDS = [
  'c',
  'capot',
  'pass',
  'p',
  'play',
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

/** Parse a Polignac CLI command into API exec arguments (indices are 0-based). */
export function parsePolignacCommand(input: string): CliParseResult<PolignacArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'capot':
      return { args: ['capot'] as PolignacArgs };
    case 'pass':
      return { args: ['pass'] as PolignacArgs };
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as PolignacArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as PolignacArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as PolignacArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as PolignacArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as PolignacArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as PolignacArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Polignac CLI mode. */
export const POLIGNAC_HELP: string[] = [
  'c/capot      - Declare capot (win every trick)',
  'pass         - Proceed without declaring',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
