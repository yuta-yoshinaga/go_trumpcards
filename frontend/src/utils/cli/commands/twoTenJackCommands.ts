import type { twoTenJackApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type TwoTenJackArgs = Parameters<typeof twoTenJackApi.exec>;

const EXTRA_COMMANDS = ['declare', 'd'];

/** Parse a Two Ten Jack CLI command into API exec arguments.
 *
 * Supported commands:
 * - `declare <suit>` / `d <suit>` — declare trump (suit: 1=spade, 2=club, 3=heart, 4=diamond)
 * - shared trick commands (`p`, `n`, `nr`, `h`, `r`)
 */
export function parseTwoTenJackCommand(input: string): CliParseResult<TwoTenJackArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    if (cmd === 'declare' || cmd === 'd') {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: declare <suit 1-4>' };
      if (parsed.value < 1 || parsed.value > 4) {
        return { error: 'Suit must be 1 (spade), 2 (club), 3 (heart), or 4 (diamond)' };
      }
      return { command: 'declare', bid: parsed.value };
    }
    return null;
  });

  if ('error' in result) return { error: result.error };
  if (result.command === 'declare') {
    return { args: ['declare', result.bid] };
  }
  return { args: [result.command as TwoTenJackArgs[0], undefined, result.cardIndex] };
}

/** Help text for Two Ten Jack CLI mode. */
export const TWOTENJACK_HELP: string[] = ['d/declare <s> - Declare trump suit (1=♠ 2=♣ 3=♥ 4=♦)', ...TRICK_HELP];
