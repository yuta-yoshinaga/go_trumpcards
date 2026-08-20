import type { paigowApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PaiGowArgs = Parameters<typeof paigowApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 's', 'set', 'log', 'r', 'reset', 'h', 'hint', 'help', '?'];

/** Parse a Pai Gow Poker CLI command into API exec arguments. */
export function parsePaigowCommand(input: string): CliParseResult<PaiGowArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 's':
    case 'set': {
      const low0 = parseIntArg(args, 0);
      if ('error' in low0) return { error: 'Usage: s <low0> <low1>' };
      const low1 = parseIntArg(args, 1);
      if ('error' in low1) return { error: 'Usage: s <low0> <low1>' };
      return { args: ['set', undefined, low0.value, low1.value] };
    }
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

/** Help text for Pai Gow Poker CLI mode. */
export const PAIGOW_HELP: string[] = [
  'b <amt>       - Place bet',
  's <i0> <i1>   - Set low hand (2 card indices)',
  'log            - Show action log',
  'r/reset        - Reset game',
  'h/hint         - Get a hint',
];
