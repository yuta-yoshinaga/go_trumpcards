import type { scartoApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ScartoArgs = Parameters<typeof scartoApi.exec>;

const VALID_COMMANDS = [
  's',
  'scarto',
  'd',
  'discard',
  'p',
  'play',
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

/** Parse a Scarto (スカルト) CLI command into API exec arguments. */
export function parseScartoCommand(input: string): CliParseResult<ScartoArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 's':
    case 'scarto':
    case 'd':
    case 'discard': {
      const indices: number[] = [];
      for (let i = 0; i < 3; i++) {
        const parsed = parseIntArg(args, i);
        if ('error' in parsed) return { error: 'Usage: scarto <i1> <i2> <i3>' };
        indices.push(parsed.value);
      }
      return { args: ['scarto', { cardIndices: indices }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
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

/** Help text for Scarto (スカルト) CLI mode. */
export const SCARTO_HELP: string[] = [
  's/scarto <i1> <i2> <i3>          - Bury 3 low pip cards (dealer only, Scarto phase)',
  'p <idx>                          - Play a card (Play phase, must follow suit / trump)',
  'n/next                           - Next trick',
  'nr/nextround                     - Next deal',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
