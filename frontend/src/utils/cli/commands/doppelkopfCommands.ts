import type { doppelkopfApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DoppelkopfArgs = Parameters<typeof doppelkopfApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'a',
  'announce',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Doppelkopf CLI command into API exec arguments. */
export function parseDoppelkopfCommand(input: string): CliParseResult<DoppelkopfArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'a':
    case 'announce':
      return { args: ['announce'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Doppelkopf CLI mode. */
export const DOPPELKOPF_HELP: string[] = [
  'p <idx>          - Play a card (Play phase, must follow suit)',
  'a/announce       - Announce Re/Kontra (first trick only)',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
