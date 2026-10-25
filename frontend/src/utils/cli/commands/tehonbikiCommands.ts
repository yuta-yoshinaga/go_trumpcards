import type { tehonbikiApi } from '../../../api/games/tehonbiki';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type Args = Parameters<typeof tehonbikiApi.exec>;
const VALID_COMMANDS = ['bet', 'b', 'half', 'next', 'hint', 'h', 'log', 'reset', 'r'];
const MIN_BET = 10;
const MAX_BET = 500;
/** CLI help for Tehonbiki. */
export const TEHONBIKI_CLI_HELP = [
  'bet (b) <type> <numbers...> <amount> (10..500)',
  'half <numbers...> <amount> (10..500)',
  'next',
  'hint (h)',
  'log',
  'reset (r)',
];
/** Parses a Tehonbiki CLI command. */
export function parseTehonbikiCommand(input: string): CliParseResult<Args> {
  const { cmd, args } = splitCommand(input);
  if (cmd === 'next' || cmd === 'hint' || cmd === 'h' || cmd === 'log' || cmd === 'reset' || cmd === 'r')
    return { args: [cmd === 'r' ? 'reset' : cmd === 'h' ? 'hint' : cmd] };
  if (cmd === 'bet' || cmd === 'b' || cmd === 'half') {
    const betArgs = cmd === 'half' ? ['half', ...args] : args;
    if (betArgs.length < 3) return { error: TEHONBIKI_CLI_HELP[0] };
    const bet = parseIntArg(betArgs, betArgs.length - 1);
    if ('error' in bet) return { error: TEHONBIKI_CLI_HELP[0] };
    const numbers = betArgs.slice(1, -1).map((_, i) => parseIntArg(betArgs, i + 1));
    if (numbers.some((n) => 'error' in n)) return { error: TEHONBIKI_CLI_HELP[0] };
    const values = numbers.map((n) => (n as { value: number }).value);
    if (values.some((n) => n < 1 || n > 6) || bet.value < MIN_BET || bet.value > MAX_BET)
      return { error: TEHONBIKI_CLI_HELP[0] };
    const betType = betArgs[0].toLowerCase();
    const expected =
      betType === 'single' ? 1 : betType === 'double' ? 2 : betType === 'triple' || betType === 'half' ? 3 : 0;
    if (expected === 0 || values.length !== expected || new Set(values).size !== values.length)
      return { error: TEHONBIKI_CLI_HELP[0] };
    if (
      betType === 'half' &&
      !([1, 2, 3].every((n) => values.includes(n)) || [4, 5, 6].every((n) => values.includes(n)))
    )
      return { error: TEHONBIKI_CLI_HELP[0] };
    return {
      args: ['bet', { betType, numbers: values, bet: bet.value }],
    };
  }
  return unknownCommand(cmd, VALID_COMMANDS);
}
