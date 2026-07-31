import type { vintApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by vintApi.exec. */
export type VintCliArgs = Parameters<typeof vintApi.exec>;

const VALID_COMMANDS = ['b', 'bid', 'ps', 'pass', 'p', 'play', 'n', 'next', 'l', 'log', 'r', 'reset', 'help', '?'];

/** Bid level bounds (sync: `VintMinLevel` / `VintMaxLevel`). */
const LEVEL_MIN = 1;
const LEVEL_MAX = 7;

/** Denomination bounds: 0=Spade … 4=NoTrump. **Spades are LOWEST.** */
const DENOM_MIN = 0;
const DENOM_MAX = 4;

/** Cards per hand (sync: `VintHandSize`). */
const HAND_MAX_INDEX = 12;

/**
 * Parses a single CLI command line for the Vint game into {@link vintApi}.exec
 * arguments.
 *
 * `bid <level> <denom>` names a level (contracting for 6 + level tricks) and a
 * denomination, where **0=♠ 1=♣ 2=♦ 3=♥ 4=NT and spades are the LOWEST** — the
 * reverse of bridge, which is why the denomination is given by number rather
 * than by suit letter.
 */
export function parseVintCommand(input: string): CliParseResult<VintCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < LEVEL_MIN || level > LEVEL_MAX) {
        return { error: `Usage: b <${LEVEL_MIN}-${LEVEL_MAX}> <${DENOM_MIN}-${DENOM_MAX}>` };
      }
      const denom = Number.parseInt(args[1] ?? '', 10);
      if (Number.isNaN(denom) || denom < DENOM_MIN || denom > DENOM_MAX) {
        return { error: 'Usage: denomination is 0-4 (0=S 1=C 2=D 3=H 4=NT; spades are LOWEST)' };
      }
      return { args: ['bid', { level, denom }] };
    }
    case 'ps':
    case 'pass':
      return { args: ['pass'] };
    case 'p':
    case 'play': {
      const cardIndex = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(cardIndex) || cardIndex < 0 || cardIndex > HAND_MAX_INDEX) {
        return { error: `Usage: p <0-${HAND_MAX_INDEX}>` };
      }
      return { args: ['play', { cardIndex }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'l':
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

/** Help text shown in the CLI terminal for Vint. */
export const VINT_HELP: string[] = [
  'b <1-7> <0-4>       - Bid a level and denomination (0=S 1=C 2=D 3=H 4=NT; spades LOWEST)',
  'ps / pass           - Pass',
  'p <0-12>            - Play a hand card',
  'n / next            - Deal the next hand',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
