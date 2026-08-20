import type { kaiserApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by kaiserApi.exec. */
export type KaiserCliArgs = Parameters<typeof kaiserApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'ps',
  'pass',
  't',
  'trump',
  'd',
  'discard',
  'p',
  'play',
  'n',
  'next',
  'l',
  'log',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Bid bounds (sync: `KaiserMinBid` / `KaiserMaxBid`). The floor is 7, not 6. */
const BID_MIN = 7;
const BID_MAX = 12;

/** Suit bounds on the wire (1=Spade … 4=Diamond). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Contract bounds: 0=with trump, 1=no trump, 2=low no trump. */
const CONTRACT_MAX = 2;

/** How many cards the declarer discards (sync: `KaiserKittySize`). */
const DISCARD_COUNT = 2;

/**
 * Parses a single CLI command line for the Kaiser game into
 * {@link kaiserApi}.exec arguments.
 *
 * `bid <7-12> [0-2]` bids a number of **points** — not tricks — with an
 * optional contract (0 = with trump, 1 = no trump, 2 = low no trump, where the
 * ranking reverses). `trump <1-4>` names the suit, `discard <i> <j>` throws the
 * two kitty cards back (never the ♥5 or ♠3), `play <i>` plays a card and
 * `next` deals again.
 */
export function parseKaiserCommand(input: string): CliParseResult<KaiserCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(bid) || bid < BID_MIN || bid > BID_MAX) {
        return { error: `Usage: b <${BID_MIN}-${BID_MAX}> [0-2] (points, not tricks)` };
      }
      let contract = 0;
      if (args.length > 1) {
        contract = Number.parseInt(args[1] ?? '', 10);
        if (Number.isNaN(contract) || contract < 0 || contract > CONTRACT_MAX) {
          return { error: 'Usage: contract is 0 (trump), 1 (no trump) or 2 (low no trump)' };
        }
      }
      return { args: ['bid', { bid, contract }] };
    }
    case 'ps':
    case 'pass':
      return { args: ['pass'] };
    case 't':
    case 'trump': {
      const suit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(suit) || suit < SUIT_MIN || suit > SUIT_MAX) {
        return { error: 'Usage: t <1-4> (1=S 2=C 3=H 4=D)' };
      }
      return { args: ['trump', { suit }] };
    }
    case 'd':
    case 'discard': {
      if (args.length < DISCARD_COUNT) return { error: `Usage: d <i> <j> (exactly ${DISCARD_COUNT} cards)` };
      const indices: number[] = [];
      for (const a of args.slice(0, DISCARD_COUNT)) {
        const v = Number.parseInt(a, 10);
        if (Number.isNaN(v) || v < 0) return { error: `Usage: d <i> <j> (exactly ${DISCARD_COUNT} cards)` };
        indices.push(v);
      }
      return { args: ['discard', { indices }] };
    }
    case 'p':
    case 'play': {
      const cardIndex = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(cardIndex) || cardIndex < 0) return { error: 'Usage: p <i>' };
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
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text shown in the CLI terminal for Kaiser. */
export const KAISER_HELP: string[] = [
  'b <7-12> [0-2]      - Bid POINTS (not tricks); 0=trump 1=NT 2=low NT',
  'ps / pass           - Pass',
  't <1-4>             - Name trump (1=S 2=C 3=H 4=D)',
  'd <i> <j>           - Discard two after taking the kitty (never the H5 or S3)',
  'p <i>               - Play a hand card',
  'n / next            - Deal the next hand',
  'l / log             - Show action log',
  'r / reset           - Reset game',
  'h/hint       - Get a hint',
];
