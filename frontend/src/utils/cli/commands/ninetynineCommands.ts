import type { ninetyNineApi } from '../../../api/gameApi';
import { splitCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type NinetyNineArgs = Parameters<typeof ninetyNineApi.exec>;

const EXTRA_COMMANDS = ['bid'];

/** Parse a Ninety-Nine CLI command into API exec arguments. The bid command buries 3 cards. */
export function parseNinetynineCommand(input: string): CliParseResult<NinetyNineArgs> {
  const { cmd, args } = splitCommand(input);

  if (cmd === 'bid') {
    if (args.length !== 3) return { error: 'Usage: bid <i> <j> <k>' };
    const buryIndices: number[] = [];
    for (const a of args) {
      const n = Number.parseInt(a, 10);
      if (Number.isNaN(n) || n < 0) return { error: 'Usage: bid <i> <j> <k>' };
      buryIndices.push(n);
    }
    return { args: ['bid', buryIndices, undefined] };
  }

  const result = parseTrickCommand(input, EXTRA_COMMANDS);
  if ('error' in result) return { error: result.error };
  if (result.command === 'play') {
    return { args: ['play', undefined, result.cardIndex] };
  }
  return { args: [result.command as NinetyNineArgs[0], undefined, undefined] };
}

/** Help text for Ninety-Nine CLI mode. */
export const NINETYNINE_HELP: string[] = ['bid <i> <j> <k> - Bury 3 cards', ...TRICK_HELP];
