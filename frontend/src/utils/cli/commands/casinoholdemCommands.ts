import type { casinoholdemApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CasinoHoldemArgs = Parameters<typeof casinoholdemApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'c', 'call', 'f', 'fold', 'log', 'r', 'reset', 'h', 'hint', 'help', '?'];

/** Parse a Casino Hold'em CLI command into API exec arguments. */
export function parseCasinoholdemCommand(input: string): CliParseResult<CasinoHoldemArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount> [bonusBet]' };
      if (args.length >= 2) {
        const bonus = parseIntArg(args, 1);
        if ('error' in bonus) return { error: 'Usage: b <amount> [bonusBet]' };
        return { args: ['bet', amount.value, bonus.value] };
      }
      return { args: ['bet', amount.value] };
    }
    case 'c':
    case 'call':
      return { args: ['call'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
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

/** Help text for Casino Hold'em CLI mode. */
export const CASINOHOLDEM_HELP: string[] = [
  'b <amt> [bn] - Ante bet (optional AA bonus side bet)',
  'c/call       - Call 2x ante (reveals turn/river)',
  'f/fold       - Fold (forfeit ante)',
  'log          - Show action log',
  'r/reset      - Reset game',
  'h/hint       - Get a hint',
];
