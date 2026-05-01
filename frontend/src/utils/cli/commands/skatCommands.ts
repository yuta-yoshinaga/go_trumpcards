import type { skatApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SkatArgs = Parameters<typeof skatApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pa',
  'pass',
  'ps',
  'pickskat',
  'sk',
  'skipskat',
  'd',
  'discard',
  'g',
  'game',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Skat CLI command into API exec arguments. */
export function parseSkatCommand(input: string): CliParseResult<SkatArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid':
      return { args: ['bid', { accept: true }] };
    case 'pa':
    case 'pass':
      return { args: ['bid', { accept: false }] };
    case 'ps':
    case 'pickskat':
      return { args: ['pickskat', { pickup: true }] };
    case 'sk':
    case 'skipskat':
      return { args: ['pickskat', { pickup: false }] };
    case 'd':
    case 'discard': {
      const a = parseIntArg(args, 0);
      const b = parseIntArg(args, 1);
      if ('error' in a || 'error' in b) return { error: 'Usage: d <idxA> <idxB>' };
      return { args: ['discard', { discardA: a.value, discardB: b.value }] };
    }
    case 'g':
    case 'game': {
      const gt = parseIntArg(args, 0);
      if ('error' in gt) return { error: 'Usage: g <gameType> [trumpSuit]  (1=suit 2=grand 3=null)' };
      if (gt.value < 1 || gt.value > 3) {
        return { error: 'gameType must be 1 (suit), 2 (grand), or 3 (null)' };
      }
      const ts = args.length >= 2 ? parseIntArg(args, 1) : null;
      if (ts && 'error' in ts) return { error: 'Invalid trump suit' };
      if (ts && !('error' in ts) && (ts.value < 1 || ts.value > 4)) {
        return { error: 'trumpSuit must be 1 (spade), 2 (clover), 3 (heart), or 4 (diamond)' };
      }
      return { args: ['game', { gameType: gt.value, trumpSuit: ts ? ts.value : undefined }] };
    }
    case 'p':
    case 'play': {
      const ci = parseIntArg(args, 0);
      if ('error' in ci) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', { cardIndex: ci.value }] };
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

/** Help text for Skat CLI mode. */
export const SKAT_HELP: string[] = [
  'b/bid         - Accept the bid',
  'pa/pass       - Pass the bid',
  'ps/pickskat   - Pick up the Skat',
  'sk/skipskat   - Play hand (skip Skat)',
  'd <a> <b>     - Discard 2 cards from hand by index',
  'g <gt> [ts]   - Declare game (1=suit 2=grand 3=null), optional trump suit',
  'p <idx>       - Play card from hand by index',
  'n/next        - Next trick',
  'nr/nextround  - Next round',
  'h/hint        - Show suggested move',
  'log           - Show action log',
  'r/reset       - Reset match',
];
