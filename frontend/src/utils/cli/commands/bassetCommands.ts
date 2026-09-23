import type { bassetApi } from '../../../api/games/basset';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by the Basset API command. */
export type BassetCliArgs = Parameters<typeof bassetApi.exec>;
const VALID_COMMANDS = ['b', 'bet', 'deal', 'd', 'take', 'paroli', 'next', 'n', 'log', 'reset', 'r'];

/** Parses a Basset terminal command. */
export function parseBassetCommand(input: string): CliParseResult<BassetCliArgs> {
  const { cmd, args } = splitCommand(input);
  if (cmd === 'b' || cmd === 'bet') {
    const rank = Number.parseInt(args[0] ?? '', 10);
    const amount = Number.parseInt(args[1] ?? '', 10);
    if (Number.isNaN(rank) || rank < 1 || rank > 13 || Number.isNaN(amount) || amount <= 0)
      return { error: 'Usage: b <rank 1-13> <amount>' };
    return { args: ['bet', { rank, amount }] };
  }
  if (cmd === 'd' || cmd === 'deal') return { args: ['deal'] };
  if (cmd === 'take') return { args: ['take'] };
  if (cmd === 'paroli') return { args: ['paroli'] };
  if (cmd === 'n' || cmd === 'next') return { args: ['next'] };
  if (cmd === 'r' || cmd === 'reset') return { args: ['reset'] };
  if (cmd === 'log') return { args: ['log'] };
  const suggestion = suggestCommand(cmd, VALID_COMMANDS);
  return { error: `Unknown command: ${cmd}${suggestion ? `. Did you mean ${suggestion}?` : ''}` };
}

/** Help text shown for Basset. */
export const BASSET_HELP = [
  'b <rank> <amount> - place a wager',
  'deal - deal banker and player cards',
  'take - receive the current winnings',
  'paroli - leave the wager and increase the payout stage',
  'next - start the next deal',
  'reset - reset the game',
];
