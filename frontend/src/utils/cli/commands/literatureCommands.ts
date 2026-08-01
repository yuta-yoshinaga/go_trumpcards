import type { literatureApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by literatureApi.exec. */
export type LiteratureCliArgs = Parameters<typeof literatureApi.exec>;

const VALID_COMMANDS = ['a', 'ask', 'c', 'claim', 'l', 'log', 'r', 'reset', 'help', '?'];

/** Seats at the table (sync: `LiteraturePlayerCnt`). */
const SEAT_MAX = 5;

/** Suit bounds: 1=Spade 2=Clover 3=Heart 4=Diamond. */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Rank bounds. */
const RANK_MIN = 1;
const RANK_MAX = 13;

/** Half-suits (sync: `LiteratureHalfSuitCnt`). */
const HALF_MAX = 7;

/** Cards in a half-suit (sync: `LiteratureHalfSuitSize`). */
const HALF_SIZE = 6;

/**
 * Parses a single CLI command line for the Literature game into
 * {@link literatureApi}.exec arguments.
 *
 * `ask <seat> <suit> <rank>` takes all three, because a card is only identified
 * by suit *and* rank; the legality checks — **opponents only**, you must hold
 * the half-suit, and you may not ask for a card you hold — belong to the server
 * because they depend on hands the client cannot see.
 *
 * `claim <half> <six seats>` needs **all six placements**: a claim states where
 * every card is, and getting one wrong is what makes the difference between
 * taking the half-suit and having it cancelled.
 */
export function parseLiteratureCommand(input: string): CliParseResult<LiteratureCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'a':
    case 'ask': {
      if (args.length < 3) {
        return { error: 'Usage: a <seat> <suit 1-4> <rank 1-13>' };
      }
      const target = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(target) || target < 0 || target > SEAT_MAX) {
        return { error: `Usage: the seat is 0-${SEAT_MAX}` };
      }
      const suit = Number.parseInt(args[1] ?? '', 10);
      if (Number.isNaN(suit) || suit < SUIT_MIN || suit > SUIT_MAX) {
        return { error: 'Usage: the suit is 1-4 (1=S 2=C 3=H 4=D)' };
      }
      const value = Number.parseInt(args[2] ?? '', 10);
      if (Number.isNaN(value) || value < RANK_MIN || value > RANK_MAX) {
        return { error: `Usage: the rank is ${RANK_MIN}-${RANK_MAX}` };
      }
      return { args: ['ask', { target, suit, value }] };
    }
    case 'c':
    case 'claim': {
      if (args.length < 1 + HALF_SIZE) {
        return { error: `Usage: c <half 0-${HALF_MAX}> followed by all ${HALF_SIZE} holders` };
      }
      const halfSuit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(halfSuit) || halfSuit < 0 || halfSuit > HALF_MAX) {
        return { error: `Usage: the half-suit is 0-${HALF_MAX}` };
      }
      const holders: number[] = [];
      for (let i = 0; i < HALF_SIZE; i++) {
        const seat = Number.parseInt(args[1 + i] ?? '', 10);
        if (Number.isNaN(seat) || seat < 0 || seat > SEAT_MAX) {
          return { error: `Usage: every holder is a seat 0-${SEAT_MAX}` };
        }
        holders.push(seat);
      }
      return { args: ['claim', { halfSuit, holders }] };
    }
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

/** Help text shown in the CLI terminal for Literature. */
export const LITERATURE_HELP: string[] = [
  'a <seat> <1-4> <1-13> - Ask an OPPONENT for a card (you must hold that half-suit, and not that card)',
  'c <half> <seat x6>    - Claim a half-suit, placing all six cards',
  'l / log               - Show action log',
  'r / reset             - Reset game',
];
