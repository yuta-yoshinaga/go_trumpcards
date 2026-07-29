import type { niuniuApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type NiuNiuArgs = Parameters<typeof niuniuApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Niu Niu CLI command into API exec arguments. */
export function parseNiuNiuCommand(input: string): CliParseResult<NiuNiuArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      if (args.length === 0) return { error: 'Usage: b <amount>' };
      const amount = Number(args[0]);
      if (!Number.isFinite(amount) || amount <= 0) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount] };
    }
    case 'log':
      return { args: ['log'] };
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

/** Help text for Niu Niu CLI mode. */
export const NIUNIU_HELP: string[] = [
  'b <amount>      - Bet, deal and settle in one step',
  'log             - Show action log',
  'r/reset         - Next round',
];
