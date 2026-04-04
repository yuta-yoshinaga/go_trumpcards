import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';

const VP_COMMANDS = ['b', 'bet', 'hold', 'r', 'reset', 'help', '?'];

/** Parse video poker commands (shared by videopoker, deuceswild, jokerpoker). */
export function parseVideoPokerCommand(
  input: string,
): { command: string; amount?: number; indices?: number[] } | { error: string } {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: b <amount>' };
      return { command: 'bet', amount: parsed.value };
    }
    case 'hold': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: hold <idx...>' };
      return { command: 'hold', indices: parsed.values };
    }
    case 'r':
    case 'reset':
      return { command: 'reset' };
    default: {
      const suggestion = suggestCommand(cmd, VP_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Shared help text for video poker games. */
export const VIDEO_POKER_HELP: string[] = [
  'b <amount>  - Place bet',
  'hold <idx...>- Hold cards (e.g., hold 0 2 4)',
  'r/reset     - Reset game',
];
