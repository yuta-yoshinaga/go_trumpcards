import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';

/** Shared trick-taking command aliases used by Spades, Euchre, Napoleon, OhHell, Bridge, Pinochle. */
const TRICK_COMMANDS = ['p', 'play', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse shared trick-taking commands. Returns { command, cardIndex? } or error. */
export function parseTrickCommand(
  input: string,
  extraCommands?: string[],
  extraParser?: (
    cmd: string,
    args: string[],
  ) => { command: string; cardIndex?: number; bid?: number } | { error: string } | null,
): { command: string; cardIndex?: number; bid?: number } | { error: string } {
  const { cmd, args } = splitCommand(input);
  const allCommands = extraCommands ? [...TRICK_COMMANDS, ...extraCommands] : TRICK_COMMANDS;

  if (extraParser) {
    const result = extraParser(cmd, args);
    if (result) return result;
  }

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { command: 'play', cardIndex: parsed.value };
    }
    case 'n':
    case 'next':
      return { command: 'next' };
    case 'nr':
    case 'nextround':
      return { command: 'nextround' };
    case 'h':
    case 'hint':
      return { command: 'hint' };
    case 'r':
    case 'reset':
      return { command: 'reset' };
    default: {
      const suggestion = suggestCommand(cmd, allCommands);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Shared help text for trick-taking games. */
export const TRICK_HELP: string[] = [
  'p <idx>     - Play a card',
  'n/next      - Next trick',
  'nr/nextround- Next round',
  'h/hint      - Show hint',
  'r/reset     - Reset game',
];
