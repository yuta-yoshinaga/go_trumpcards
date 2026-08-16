import type { trexApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TrexArgs = Parameters<typeof trexApi.exec>;

const VALID_COMMANDS = ['c', 'choose', 'p', 'play', 's', 'pass', 'n', 'next', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Trex CLI command into API exec arguments. */
export function parseTrexCommand(input: string): CliParseResult<TrexArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'choose': {
      if (args.length === 0) return { error: `Usage: ${cmd} <contract>` };
      const n = Number(args[0]);
      // Contract 0 is the king of hearts, so zero is valid input.
      if (!Number.isInteger(n) || n < 0 || n > 4) return { error: `Invalid contract: ${args[0]}` };
      return { args: ['choose', n] };
    }
    case 'p':
    case 'play': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index>` };
      const n = Number(args[0]);
      if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${args[0]}` };
      // The hand index goes in the THIRD position so it can never be read as a
      // contract number.
      return { args: ['play', undefined, n] };
    }
    case 's':
    case 'pass':
      return { args: ['pass'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'log':
      return { args: ['log'] };
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

/** Help text for Trex CLI mode. */
export const TREX_HELP: string[] = [
  'c <n>           - Choose contract n (0=KingOfHearts 1=Diamonds 2=Queens 3=Tricks 4=Dominoes)',
  'p <i>           - Play the hand card at index i',
  's/pass          - Pass in the dominoes (only with no legal play)',
  'n/next          - Deal the next hand',
  'log             - Show action log',
  'r/reset         - New game',
  'h/hint      - Get a hint',
];
