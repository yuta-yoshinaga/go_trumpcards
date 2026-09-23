import type { tehonbikiApi } from '../../../api/games/tehonbiki';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type Args = Parameters<typeof tehonbikiApi.exec>;
/** CLI help for Tehonbiki. */
export const TEHONBIKI_CLI_HELP = ['bet <type> <numbers...> <amount>', 'next', 'hint', 'log', 'reset (r)'];
/** Parses a Tehonbiki CLI command. */
export function parseTehonbikiCommand(input: string): CliParseResult<Args> {
  const { cmd, args } = splitCommand(input);
  if (cmd === 'next' || cmd === 'hint' || cmd === 'log' || cmd === 'reset' || cmd === 'r')
    return { args: [cmd === 'r' ? 'reset' : cmd] };
  if (cmd === 'bet' || cmd === 'b') {
    if (args.length < 3) return { error: TEHONBIKI_CLI_HELP[0] };
    const bet = parseIntArg(args, args.length - 1);
    if ('error' in bet) return { error: TEHONBIKI_CLI_HELP[0] };
    const numbers = args.slice(1, -1).map((_, i) => parseIntArg(args, i + 1));
    if (numbers.some((n) => 'error' in n)) return { error: TEHONBIKI_CLI_HELP[0] };
    return {
      args: ['bet', { betType: args[0], numbers: numbers.map((n) => (n as { value: number }).value), bet: bet.value }],
    };
  }
  const s = suggestCommand(cmd, ['bet', 'next', 'hint', 'log', 'reset']);
  return { error: `Unknown command: ${cmd}${s ? `. Did you mean ${s}?` : ''}` };
}
