import type { indianpokerApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { BETTING_HELP, parseBettingCommand } from './sharedBettingCommands';

type IndianPokerArgs = Parameters<typeof indianpokerApi.exec>;

const EXTRA_COMMANDS = ['log'];

/** Parse an Indian Poker CLI command into API exec arguments. */
export function parseIndianpokerCommand(input: string): CliParseResult<IndianPokerArgs> {
  const result = parseBettingCommand(input, EXTRA_COMMANDS, (cmd, _args) => {
    if (cmd === 'log') return { command: 'log' };
    return null;
  });

  if ('error' in result) return { error: result.error };
  return { args: [result.command as IndianPokerArgs[0], result.amount] };
}

/** Help text for Indian Poker CLI mode. */
export const INDIANPOKER_HELP: string[] = [...BETTING_HELP, 'log         - Show action log'];
