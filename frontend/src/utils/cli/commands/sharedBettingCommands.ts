import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';

/** Shared betting command aliases used by Holdem, Omaha, ShortDeck, Pineapple, IndianPoker. */
const BETTING_COMMANDS = [
  'f',
  'fold',
  'ck',
  'check',
  'c',
  'call',
  'b',
  'bet',
  'ra',
  'raise',
  'a',
  'allin',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse shared betting commands. Returns [command, amount?] or error. */
export function parseBettingCommand(
  input: string,
  extraCommands?: string[],
  extraParser?: (cmd: string, args: string[]) => { command: string; amount?: number } | { error: string } | null,
): { command: string; amount?: number } | { error: string } {
  const { cmd, args } = splitCommand(input);
  const allCommands = extraCommands ? [...BETTING_COMMANDS, ...extraCommands] : BETTING_COMMANDS;

  // Try extra parser first
  if (extraParser) {
    const result = extraParser(cmd, args);
    if (result) return result;
  }

  switch (cmd) {
    case 'f':
    case 'fold':
      return { command: 'fold' };
    case 'ck':
    case 'check':
      return { command: 'check' };
    case 'c':
    case 'call':
      return { command: 'call' };
    case 'a':
    case 'allin':
      return { command: 'allin' };
    case 'b':
    case 'bet': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: b <amount>' };
      return { command: 'bet', amount: parsed.value };
    }
    case 'ra':
    case 'raise': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: ra <amount>' };
      return { command: 'raise', amount: parsed.value };
    }
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

/** Shared help text for betting games. */
export const BETTING_HELP: string[] = [
  'f/fold      - Fold hand',
  'ck/check    - Check',
  'c/call      - Call current bet',
  'b <amount>  - Place bet',
  'ra <amount> - Raise',
  'a/allin     - All-in',
  'r/reset     - Reset game',
];
