import type { quadrilleApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type QuadrilleArgs = Parameters<typeof quadrilleApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'k',
  'king',
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

/** Parse a Quadrille CLI command into API exec arguments. */
export function parseQuadrilleCommand(input: string): CliParseResult<QuadrilleArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase();
      if (arg === 'pass' || arg === 'p') return { args: ['bid', { bid: 0 }] };
      if (arg === 'entrar' || arg === 'e' || arg === 'solo' || arg === 's') {
        const suit = SUIT_CODES[args[1]?.toLowerCase() ?? ''];
        if (!suit) return { error: 'Usage: bid entrar|solo <s|c|h|d>' };
        const bid = arg === 'solo' || arg === 's' ? 2 : 1;
        return { args: ['bid', { bid, trumpSuit: suit }] };
      }
      return { error: 'Usage: bid pass | bid entrar <s|c|h|d> | bid solo <s|c|h|d>' };
    }
    // **落札した直後は王呼びフェーズ。** ここを解釈できないと、王を呼ぶまで
    // play はフェーズ違いで弾かれ続け、CLI モードから先へ進めなくなる。
    case 'k':
    case 'king': {
      const suit = SUIT_CODES[args[0]?.toLowerCase() ?? ''];
      if (!suit) return { error: 'Usage: king <s|c|h|d>' };
      return { args: ['king', { kingSuit: suit }] };
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

/** Help text for Quadrille CLI mode. */
export const QUADRILLE_HELP: string[] = [
  'bid pass                     - Pass in the auction (Bid phase)',
  'bid entrar <s|c|h|d>         - Declare entrar with the chosen trump suit',
  'bid solo <s|c|h|d>           - Declare solo with the chosen trump suit',
  'king <s|c|h|d>               - Call a king you do not hold (KingCall phase)',
  'p <idx>                      - Play a card (Play phase, must follow the led suit)',
  'n/next                       - Next trick',
  'nr/nextround                 - Next round',
  'h/hint                       - Show hint',
  'r/reset                      - Reset game',
];
