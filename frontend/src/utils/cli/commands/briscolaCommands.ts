import type { briscolaApi } from '../../../api/gameApi';
import type { BriscolaResponse } from '../../../types/card';
import { BriscolaPhase } from '../../../types/phases';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by briscolaApi.exec. */
export type BriscolaCliArgs = Parameters<typeof briscolaApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'h', 'hint', 'r', 'reset'];

/** Parse a Briscola command into arguments accepted by briscolaApi.exec. */
export function parseBriscolaCommand(input: string, state?: BriscolaResponse | null): CliParseResult<BriscolaCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number(args[0]);
      const handSize = state?.players.find((player) => player.isHuman)?.cards.length ?? 3;
      if (!Number.isInteger(idx) || idx < 0 || idx >= handSize) return { error: `Usage: p <idx> (0-${handSize - 1})` };
      if (state && (state.phase !== BriscolaPhase.PLAY || state.players[state.currentPlayerIdx]?.isHuman !== true))
        return { error: 'You can only play a card during your turn.' };
      return { args: ['play', idx] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Briscola CLI terminal. */
export const BRISCOLA_HELP = [
  'p <idx>          - Play a card from your hand',
  'n / next         - Continue after a trick',
  'h / hint         - Request a hint',
  'r / reset        - Reset game',
];
