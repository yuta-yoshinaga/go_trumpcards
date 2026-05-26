import type { russianpokerApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RussianPokerArgs = Parameters<typeof russianpokerApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'e',
  'exchange',
  '6',
  'buy6th',
  'sel',
  'select',
  'p',
  'play',
  'f',
  'fold',
  'force',
  'd',
  'decline',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Russian Poker CLI command into API exec arguments. */
export function parseRussianpokerCommand(input: string): CliParseResult<RussianPokerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 'e':
    case 'exchange': {
      const indices: number[] = [];
      for (let i = 0; i < args.length; i++) {
        const v = parseIntArg(args, i);
        if ('error' in v) return { error: 'Usage: e [idx ...]  (each idx in 0..4)' };
        if (v.value < 0 || v.value > 4) return { error: 'Indices must be 0..4' };
        indices.push(v.value);
      }
      return { args: ['exchange', undefined, indices] };
    }
    case '6':
    case 'buy6th':
      return { args: ['buy6th'] };
    case 'sel':
    case 'select': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: sel <idx>  (idx in 0..5)' };
      if (idx.value < 0 || idx.value > 5) return { error: 'Index must be 0..5' };
      return { args: ['select', undefined, undefined, idx.value] };
    }
    case 'p':
    case 'play':
      return { args: ['play'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'force':
      return { args: ['force'] };
    case 'd':
    case 'decline':
      return { args: ['decline'] };
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

/** Help text for Russian Poker CLI mode. */
export const RUSSIANPOKER_HELP: string[] = [
  'b <amt>       - Ante bet',
  'e [idx ...]   - Exchange cards at indices (fee = ante x count)',
  '6/buy6th      - Buy 6th card (fee = ante)',
  'sel <idx>     - Discard card at index from 6-card hand',
  'p/play        - Call (match 2x ante)',
  'f/fold        - Fold hand',
  'force         - Force exchange dealer highest card (fee = ante)',
  'd/decline     - Decline force exchange',
  'log           - Show action log',
  'r/reset       - Reset game',
];
