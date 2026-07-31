import type { sixBidSoloApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by sixBidSoloApi.exec. */
export type SixBidSoloCliArgs = Parameters<typeof sixBidSoloApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'ps',
  'pass',
  'd',
  'declare',
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

/** Bid bounds (sync: `SixBidSoloMinBid` / `SixBidSoloMaxBid`). */
const BID_MIN = 1;
const BID_MAX = 6;

/** Suit bounds: 1=Spade 2=Clover 3=Heart 4=Diamond. */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Rank bounds for a called card. */
const RANK_MIN = 1;
const RANK_MAX = 13;

/** Cards per hand (sync: `SixBidSoloHandSize`). **Eleven, not twelve.** */
const HAND_MAX_INDEX = 10;

/**
 * Parses a single CLI command line for the Six-Bid Solo game into
 * {@link sixBidSoloApi}.exec arguments.
 *
 * `bid <n>` names one of the six bids in ascending order — 1 = Solo,
 * 2 = Heart Solo, 3 = Misère, 4 = Guarantee Solo, 5 = Spread Misère,
 * 6 = Call Solo — and only a higher bid stands. `declare <suit>` then names the
 * trump; **a call solo continues with the card it names**, whose holder must
 * exchange it, so that form takes three arguments rather than one.
 */
export function parseSixBidSoloCommand(input: string): CliParseResult<SixBidSoloCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(bid) || bid < BID_MIN || bid > BID_MAX) {
        return {
          error: `Usage: b <${BID_MIN}-${BID_MAX}> (1=solo 2=heart solo 3=misere 4=guarantee 5=spread 6=call)`,
        };
      }
      return { args: ['bid', { bid }] };
    }
    case 'ps':
    case 'pass':
      return { args: ['pass'] };
    case 'd':
    case 'declare': {
      const suit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(suit) || suit < SUIT_MIN || suit > SUIT_MAX) {
        return { error: `Usage: d <${SUIT_MIN}-${SUIT_MAX}> (1=S 2=C 3=H 4=D)` };
      }
      if (args.length === 1) return { args: ['declare', { suit }] };
      // **指名札はスートとランクの両方が要る。**片方だけでは札が決まらない。
      if (args.length < 3) {
        return { error: 'Usage: a call solo needs both the called suit and its rank (e.g. d 1 3 13)' };
      }
      const calledSuit = Number.parseInt(args[1] ?? '', 10);
      const calledValue = Number.parseInt(args[2] ?? '', 10);
      if (Number.isNaN(calledSuit) || calledSuit < SUIT_MIN || calledSuit > SUIT_MAX) {
        return { error: 'Usage: the called suit is 1-4 (1=S 2=C 3=H 4=D)' };
      }
      if (Number.isNaN(calledValue) || calledValue < RANK_MIN || calledValue > RANK_MAX) {
        return { error: `Usage: the called rank is ${RANK_MIN}-${RANK_MAX}` };
      }
      return { args: ['declare', { suit, calledSuit, calledValue }] };
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

/** Help text shown in the CLI terminal for Six-Bid Solo. */
export const SIXBIDSOLO_HELP: string[] = [
  'b <1-6>             - Bid (1=solo 2=heart solo 3=misere 4=guarantee 5=spread 6=call); only a HIGHER bid stands',
  'ps / pass           - Pass',
  'd <1-4>             - Name trump (1=S 2=C 3=H 4=D)',
  'd <1-4> <1-4> <1-13>- Call solo: name trump, then the card whose holder must exchange it',
  'p <0-10>            - Play a hand card (eleven each, plus a three-card widow)',
  'n / next            - Deal the next hand',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
