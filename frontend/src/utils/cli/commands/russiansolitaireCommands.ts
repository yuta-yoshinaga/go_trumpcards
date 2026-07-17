import type { russianSolitaireApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RussianSolitaireArgs = Parameters<typeof russianSolitaireApi.exec>;

const VALID_COMMANDS = ['m', 'move', 'g', 'giveup', 'h', 'hint', 'ac', 'autocomplete', 'u', 'undo', 'r', 'reset'];

/** Parse a Russian Solitaire CLI command into API call arguments. */
export function parseRussianSolitaireCommand(input: string): CliParseResult<RussianSolitaireArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'm':
    case 'move': {
      // `m <fromCol> <toCol>`            → grab the whole face-up block (top card).
      // `m <fromCol> <cardIdx> <toCol>`  → grab from a specific face-up card in the column.
      if (args.length === 2) {
        const from = Number.parseInt(args[0], 10);
        const to = Number.parseInt(args[1], 10);
        if (Number.isNaN(from) || Number.isNaN(to)) return { error: 'Invalid column' };
        return { args: ['move', { zone: 'tableau', col: from, cardIndex: -1 }, { zone: 'tableau', col: to }] };
      }
      if (args.length === 3) {
        const from = Number.parseInt(args[0], 10);
        const cardIdx = Number.parseInt(args[1], 10);
        const to = Number.parseInt(args[2], 10);
        if (Number.isNaN(from) || Number.isNaN(to)) return { error: 'Invalid column' };
        if (Number.isNaN(cardIdx) || cardIdx < 0) return { error: 'Invalid card index' };
        return { args: ['move', { zone: 'tableau', col: from, cardIndex: cardIdx }, { zone: 'tableau', col: to }] };
      }
      return { error: 'Usage: m <fromCol> [cardIdx] <toCol>' };
    }
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Russian Solitaire CLI mode. */
export const RS_HELP: string[] = [
  'm <from> <to>          Move a tableau column block (top face-up card)',
  'm <from> <cardIdx> <to>  Move starting from a specific face-up card',
  'g              Give up',
  'h              Hint',
  'ac             Auto-complete',
  'u              Undo',
  'r              Reset',
];
