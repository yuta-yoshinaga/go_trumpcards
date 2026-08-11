import type { tarabishApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TarabishArgs = Parameters<typeof tarabishApi.exec>;

const VALID_COMMANDS = [
  't',
  'take',
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

/** Parse a Tarabish CLI command into API exec arguments (indices are 0-based). */
export function parseTarabishCommand(input: string): CliParseResult<TarabishArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'take':
      return { args: ['take'] as TarabishArgs };
    case 'pass':
      return { args: ['pass'] as TarabishArgs };
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as TarabishArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as TarabishArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as TarabishArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as TarabishArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as TarabishArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as TarabishArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Tarabish CLI mode. */
export const TARABISH_HELP: string[] = [
  't/take       - Take the turned suit as trump',
  'pass         - Pass on trump (the dealer cannot)',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
