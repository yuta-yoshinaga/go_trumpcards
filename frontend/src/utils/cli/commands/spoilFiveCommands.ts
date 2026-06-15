import type { spoilFiveApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SpoilFiveArgs = Parameters<typeof spoilFiveApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse a Spoil Five CLI command into API exec arguments. */
export function parseSpoilFiveCommand(input: string): CliParseResult<SpoilFiveArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
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

/** Help text for Spoil Five CLI mode. */
export const SPOIL_FIVE_HELP: string[] = [
  'p <idx>          - Play a card (must follow the lead suit; top trumps 5/J/♥A may be held back — Reneging)',
  'n/next           - Next trick',
  'nr/nextround     - Next round (win 3 tricks to take the pot; otherwise a Spoil carries the pot over)',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
