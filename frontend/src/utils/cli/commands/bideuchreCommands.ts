import type { bidEuchreApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by bidEuchreApi.exec. */
export type BidEuchreCliArgs = Parameters<typeof bidEuchreApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'ps',
  'pass',
  't',
  'trump',
  'p',
  'play',
  'n',
  'next',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Bid bounds (sync: `BidEuchreMinBid` / `BidEuchreMaxBid`). **Three is the floor.** */
const BID_MIN = 3;
const BID_MAX = 6;

/** Trump declaration bounds: 0=Spade … 4=NT high, 5=**NT LOW**. */
const TRUMP_MIN = 0;
const TRUMP_MAX = 5;

/** Cards per hand (sync: `BidEuchreHandSize`). */
const HAND_MAX_INDEX = 5;

/**
 * Parses a single CLI command line for the Bid Euchre game into
 * {@link bidEuchreApi}.exec arguments.
 *
 * `bid <value>` names a trick count; **three is the minimum**, and every bid
 * must beat the standing one — except the dealer's, which may equal it (the
 * server decides that, so this parser only bounds the range). `trump <n>` then
 * names **0=♠ 1=♣ 2=♦ 3=♥ 4=NT high 5=NT LOW**; there are two no-trump forms
 * because at no trump low the ranking reverses and the nine is highest.
 */
export function parseBidEuchreCommand(input: string): CliParseResult<BidEuchreCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const value = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(value) || value < BID_MIN || value > BID_MAX) {
        return { error: `Usage: b <${BID_MIN}-${BID_MAX}> (three tricks is the floor)` };
      }
      return { args: ['bid', { value }] };
    }
    case 'ps':
    case 'pass':
      return { args: ['pass'] };
    case 't':
    case 'trump': {
      const trump = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(trump) || trump < TRUMP_MIN || trump > TRUMP_MAX) {
        return { error: 'Usage: t <0-5> (0=S 1=C 2=D 3=H 4=NT high 5=NT LOW, nine highest)' };
      }
      return { args: ['trump', { trump }] };
    }
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

/** Help text shown in the CLI terminal for Bid Euchre. */
export const BIDEUCHRE_HELP: string[] = [
  'b <3-6>             - Bid a trick count (three is the floor; the DEALER may EQUAL the standing bid)',
  'ps / pass           - Pass',
  't <0-5>             - Name trump (0=S 1=C 2=D 3=H 4=NT high 5=NT LOW, nine highest)',
  'p <0-5>             - Play a hand card',
  'n / next            - Deal the next hand',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
