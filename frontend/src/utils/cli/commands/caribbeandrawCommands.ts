import type { caribbeandrawApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CaribbeanDrawArgs = Parameters<typeof caribbeandrawApi.exec>;

/** Hand size, and therefore the largest card number the player may name. */
const HAND_SIZE = 5;

/** Most cards that may be exchanged in one draw (sync: internal/domain/CaribbeanDraw.go). */
const MAX_EXCHANGE = 2;

const VALID_COMMANDS = [
  'b',
  'bet',
  'd',
  'draw',
  'p',
  'play',
  'f',
  'fold',
  'log',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Caribbean Draw Poker CLI command into API exec arguments. */
export function parseCaribbeandrawCommand(input: string): CliParseResult<CaribbeanDrawArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount> [jackpotBet]' };
      if (args.length >= 2) {
        const jpBet = parseIntArg(args, 1);
        if ('error' in jpBet) return { error: 'Usage: b <amount> [jackpotBet]' };
        return { args: ['bet', amount.value, jpBet.value] };
      }
      return { args: ['bet', amount.value] };
    }
    case 'd':
    case 'draw':
      return parseDraw(args);
    case 'p':
    case 'play':
      return { args: ['play'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
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

/**
 * Parse the arguments of `d`/`draw` into 0-based hand indices.
 *
 * The player types the card **numbers** shown on screen, which start at 1; the
 * API takes 0-based positions. Skipping the conversion is silent — every index
 * still lands inside the hand, so the wrong card is discarded and nothing
 * reports an error. A bare `d` stands pat.
 */
function parseDraw(args: string[]): CliParseResult<CaribbeanDrawArgs> {
  if (args.length === 0) return { args: ['draw', undefined, undefined, []] };
  if (args.length > MAX_EXCHANGE) return { error: `Usage: d [n...] - at most ${MAX_EXCHANGE} cards` };

  const parsed = parseIntSlice(args);
  if ('error' in parsed) return { error: `Usage: d [n...] - card numbers 1-${HAND_SIZE}` };

  const seen = new Set<number>();
  for (const n of parsed.values) {
    if (n < 1 || n > HAND_SIZE) return { error: `Usage: d [n...] - card numbers 1-${HAND_SIZE}` };
    if (seen.has(n)) return { error: `Usage: d [n...] - card ${n} named twice` };
    seen.add(n);
  }

  return { args: ['draw', undefined, undefined, parsed.values.map((n) => n - 1)] };
}

/** Help text for Caribbean Draw Poker CLI mode. */
export const CARIBBEANDRAW_HELP: string[] = [
  'b <amt> [jp] - Ante bet (optional jackpot side bet)',
  'd [n...]     - Exchange up to 2 cards by number (fee = ante); bare d stands pat',
  'p/play       - Call (match 2x ante)',
  'f/fold       - Fold hand',
  'log          - Show action log',
  'r/reset      - Reset game',
  'h/hint       - Get a hint',
];
