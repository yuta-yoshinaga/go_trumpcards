import type { omiApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type OmiArgs = Parameters<typeof omiApi.exec>;

/** Map trump suit name or number string to suit number (1=♠ 2=♣ 3=♥ 4=♦). */
const TRUMP_SUIT_MAP: Record<string, number> = {
  '1': 1,
  '2': 2,
  '3': 3,
  '4': 4,
  spade: 1,
  s: 1,
  clover: 2,
  club: 2,
  c: 2,
  heart: 3,
  h: 3,
  diamond: 4,
  d: 4,
};

const EXTRA_COMMANDS = ['t', 'trump', 'call', 'calltrump'];

/** Parse an Omi CLI command into API exec arguments.
 * Trump is declared with `t <1-4>` (1=♠ 2=♣ 3=♥ 4=♦), matching the CUI controller. */
export function parseOmiCommand(input: string): CliParseResult<OmiArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 't':
      case 'trump':
      case 'call':
      case 'calltrump': {
        if (args.length === 0) return { error: 'Usage: t <1-4>  (1=♠ 2=♣ 3=♥ 4=♦)' };
        const suit = TRUMP_SUIT_MAP[args[0].toLowerCase()];
        if (suit === undefined) return { error: 'Invalid suit. Use: 1=♠ 2=♣ 3=♥ 4=♦ (or spade/clover/heart/diamond)' };
        return { command: `calltrump:${suit}` };
      }
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };

  const cmd = result.command;
  if (cmd.startsWith('calltrump:')) {
    const suit = Number(cmd.split(':')[1]);
    return { args: ['calltrump', undefined, suit] };
  }
  return { args: [cmd as OmiArgs[0], result.cardIndex] };
}

/** Help text for Omi CLI mode. */
export const OMI_HELP: string[] = ['t <1-4>     - Call trump suit (1=♠ 2=♣ 3=♥ 4=♦)', ...TRICK_HELP];
