import type { pageoneApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PageOneArgs = Parameters<typeof pageoneApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'd',
  'draw',
  'dc',
  'declare',
  'sk',
  'skip',
  'nr',
  'nextround',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Page One CLI command into API arguments. */
export function parsePageoneCommand(input: string): CliParseResult<PageOneArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', parsed.value] };
    }
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'dc':
    case 'declare':
      return { args: ['declare'] };
    case 'sk':
    case 'skip':
      return { args: ['skip'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Page One CLI mode. */
export const PAGEONE_HELP: string[] = [
  'p <idx>      - Play a card',
  'd/draw       - Draw from pile',
  'dc/declare   - Declare Page One!',
  'sk/skip      - Skip declaration (penalty: 2 cards)',
  'nr/nextround - Next round',
  'r/reset      - Reset game',
  'h/hint       - Get a hint',
];
