import type { ultiApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type UltiArgs = Parameters<typeof ultiApi.exec>;

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

/** Maps a trump-suit letter to its numeric code (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_CODES: Readonly<Record<string, number>> = { s: 1, c: 2, h: 3, d: 4 };

/** Parse an Ulti (Ultimo) CLI command into API exec arguments. */
export function parseUltiCommand(input: string): CliParseResult<UltiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase();
      if (arg === 'party' || arg === 'p') {
        const suit = SUIT_CODES[args[1]?.toLowerCase() ?? ''];
        if (!suit) return { error: 'Usage: bid party <s|c|h|d>' };
        return { args: ['bid', { contract: 'party', trumpSuit: suit }] };
      }
      if (arg === 'betli' || arg === 'b') return { args: ['bid', { contract: 'betli' }] };
      if (arg === 'durchmarsch' || arg === 'd') return { args: ['bid', { contract: 'durchmarsch' }] };
      if (arg === 'ulti' || arg === 'u') {
        const suit = SUIT_CODES[args[1]?.toLowerCase() ?? ''];
        if (!suit) return { error: 'Usage: bid ulti <s|c|h|d>' };
        return { args: ['bid', { contract: 'ulti', trumpSuit: suit }] };
      }
      return { error: 'Usage: bid party <s|c|h|d> | bid betli | bid durchmarsch | bid ulti <s|c|h|d>' };
    }
    case 'd':
    case 'discard': {
      const a = parseIntArg(args, 0);
      const b = parseIntArg(args, 1);
      if ('error' in a || 'error' in b) return { error: 'Usage: discard <i> <j>' };
      return { args: ['discard', { cardIndices: [a.value, b.value] }] };
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

/** Help text for Ulti (Ultimo) CLI mode. */
export const ULTI_HELP: string[] = [
  'bid party <s|c|h|d>          - Declare Party with the chosen trump suit',
  'bid betli                    - Declare Betli (win no trick, no trump)',
  'bid durchmarsch              - Declare Durchmarsch (win every trick)',
  'bid ulti <s|c|h|d>           - Declare Ulti (win the last trick with the trump 7)',
  'discard <i> <j>              - Discard 2 talon cards (Discard phase)',
  'p <idx>                      - Play a card (Play phase, must follow the led suit)',
  'n/next                       - Next trick',
  'nr/nextround                 - Next deal',
  'h/hint                       - Show hint',
  'r/reset                      - Reset game',
];
