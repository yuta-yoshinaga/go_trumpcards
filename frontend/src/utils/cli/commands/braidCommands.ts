import type { braidApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BraidArgs = Parameters<typeof braidApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'dir',
  'direction',
  'm',
  'move',
  'g',
  'giveup',
  'ac',
  'autocomplete',
  'u',
  'undo',
  'h',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Braid CLI command into API exec arguments. */
export function parseBraidCommand(input: string): CliParseResult<BraidArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'dir':
    case 'direction':
      return parseDirectionCommand(args);
    case 'm':
    case 'move':
      return parseMoveCommand(args);
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
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

function parseDirectionCommand(args: string[]): CliParseResult<BraidArgs> {
  const tok = (args[0] ?? '').toLowerCase();
  if (tok === 'a' || tok === 'asc' || tok === 'up') {
    return { args: ['dir', undefined, undefined, undefined, true] };
  }
  if (tok === 'd' || tok === 'desc' || tok === 'down') {
    return { args: ['dir', undefined, undefined, undefined, false] };
  }
  return { error: 'Usage: dir a (up) or dir d (down)' };
}

/** Parse a `fd2` / `hp5` slot token into its index. */
function parseSlotToken(tok: string, prefixLen: number): number | { error: string } {
  const body = tok.slice(prefixLen);
  const n = Number(body);
  if (body === '' || Number.isNaN(n)) return { error: 'Usage: fd<idx> or hp<idx>' };
  return n;
}

function parseMoveCommand(args: string[]): CliParseResult<BraidArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m br f, m fd2 f, m hp5 f, m w f, m w hp3)' };
  }
  const fromTok = args[0].toLowerCase();
  const toTok = args[1].toLowerCase();

  let from: { zone: string; col?: number };
  if (fromTok === 'br') {
    from = { zone: 'braid' };
  } else if (fromTok.startsWith('fd')) {
    const idx = parseSlotToken(fromTok, 2);
    if (typeof idx !== 'number') return idx;
    from = { zone: 'field', col: idx };
  } else if (fromTok.startsWith('hp')) {
    const idx = parseSlotToken(fromTok, 2);
    if (typeof idx !== 'number') return idx;
    from = { zone: 'helper', col: idx };
  } else if (fromTok === 'w') {
    from = { zone: 'waste' };
  } else {
    return { error: 'Invalid source: use br (braid), fd<idx>, hp<idx> or w (waste)' };
  }

  if (toTok === 'f') return { args: ['move', from, { zone: 'foundation' }] };
  // Only the waste can fill a helper -- every other slot is foundation-only.
  if (toTok.startsWith('hp')) {
    if (from.zone !== 'waste') return { error: 'Invalid move: only the waste can fill a helper' };
    const idx = parseSlotToken(toTok, 2);
    if (typeof idx !== 'number') return idx;
    return { args: ['move', from, { zone: 'helper', col: idx }] };
  }
  return { error: 'Invalid target: use f (foundation) or hp<idx> (helper, from the waste only)' };
}

/** Help text for Braid CLI mode. */
export const BRAID_HELP: string[] = [
  'd/draw          - Turn one card from the stock (redeal when empty, twice)',
  'dir a | dir d   - Fix the foundation direction (once per game)',
  'm br f          - Braid tail to a foundation',
  'm fd<i> f       - Braid field to a foundation (refills from the braid)',
  'm hp<i> f       - Helper to a foundation',
  'm w f           - Waste to a foundation',
  'm w hp<i>       - Waste into an empty helper',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
