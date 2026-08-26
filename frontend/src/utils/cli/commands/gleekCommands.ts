import type { gleekApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GleekArgs = Parameters<typeof gleekApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'd',
  'discard',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Gleek CLI command into API exec arguments. */
export function parseGleekCommand(input: string): CliParseResult<GleekArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase();
      if (arg === 'pass' || arg === 'p') return { args: ['bid', { bid: 0 }] };
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: bid <amount> | bid pass' };
      if (parsed.value < 0) return { error: 'Usage: bid <amount> | bid pass' };
      return { args: ['bid', { bid: parsed.value }] };
    }
    // **The only way out of the discard phase.** Without it the board freezes
    // right after the auction: `play` is rejected until the buyer is back down
    // to a full hand.
    case 'd':
    case 'discard': {
      const indices: number[] = [];
      for (const raw of args) {
        for (const part of raw.split(',')) {
          const token = part.trim();
          if (token === '') continue;
          const n = Number.parseInt(token, 10);
          if (Number.isNaN(n) || n < 0) return { error: 'Usage: d <idx> <idx> ... (seven cards)' };
          indices.push(n);
        }
      }
      if (indices.length === 0) return { error: 'Usage: d <idx> <idx> ... (seven cards)' };
      return { args: ['discard', { discardIndices: indices }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Gleek CLI mode. */
export const GLEEK_HELP: string[] = [
  'bid <amount>                 - Raise to that amount (steps of 2)',
  'bid pass                     - Drop out of the auction',
  'd/discard <idx> ...          - Throw the buyer’s seven cards (Discard phase)',
  'p <idx>                      - Play a card (Play phase, must follow the led suit)',
  'n/next                       - Next trick',
  'nr/nextround                 - Next deal',
  'h/hint                       - Show hint',
  'r/reset                      - Reset game',
];
