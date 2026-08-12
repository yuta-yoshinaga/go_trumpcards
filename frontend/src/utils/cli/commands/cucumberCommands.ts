import type { cucumberApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CucumberArgs = Parameters<typeof cucumberApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Cucumber CLI command into API exec arguments (indices are 0-based). */
export function parseCucumberCommand(input: string): CliParseResult<CucumberArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as CucumberArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as CucumberArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as CucumberArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as CucumberArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as CucumberArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as CucumberArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Cucumber CLI mode. */
export const CUCUMBER_HELP: string[] = [
  'p <cardIdx> - Play a card (you must beat the highest, or dump your lowest)',
  'n/next      - Deal the next round',
  'h/hint      - Show a hint',
  'g/giveup    - Give up',
  'log         - Show the action log',
  'r/reset     - Reset game',
];
