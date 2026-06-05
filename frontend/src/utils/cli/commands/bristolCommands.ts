import type { bristolApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BristolArgs = Parameters<typeof bristolApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
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

/** Parse a Bristol CLI command into API exec arguments. */
export function parseBristolCommand(input: string): CliParseResult<BristolArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
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

const MOVE_USAGE = 'Usage: m t<col> t<col> | m t<col> f | m n<fan> t<col> | m n<fan> f';

/**
 * Resolve a source/index token. Accepts compact (`t0`, `n1`) and spaced
 * (`t 0`, `n 1`) forms. Returns the matched zone, index, and how many tokens
 * were consumed, or undefined on a parse error.
 */
function resolveZone(
  tokens: string[],
  letter: string,
  zone: string,
): { zone: string; col: number; consumed: number } | undefined {
  const first = tokens[0].toLowerCase();
  if (first === letter) {
    if (tokens.length < 2) return undefined;
    const n = Number(tokens[1]);
    return Number.isNaN(n) ? undefined : { zone, col: n, consumed: 2 };
  }
  const n = Number(first.slice(1));
  return Number.isNaN(n) ? undefined : { zone, col: n, consumed: 1 };
}

function parseMoveCommand(args: string[]): CliParseResult<BristolArgs> {
  if (args.length < 2) return { error: MOVE_USAGE };

  const srcChar = args[0][0].toLowerCase();
  let source: { zone: string; col: number; consumed: number } | undefined;
  if (srcChar === 't') {
    source = resolveZone(args, 't', 'tableau');
  } else if (srcChar === 'n') {
    source = resolveZone(args, 'n', 'fan');
  }
  if (!source) return { error: MOVE_USAGE };

  const rest = args.slice(source.consumed);
  if (rest.length === 0) return { error: MOVE_USAGE };

  const destChar = rest[0][0].toLowerCase();
  if (destChar === 'f') {
    return { args: ['move', { zone: source.zone, col: source.col }, { zone: 'foundation' }] };
  }
  if (destChar === 't') {
    const dest = resolveZone(rest, 't', 'tableau');
    if (!dest) return { error: MOVE_USAGE };
    return {
      args: ['move', { zone: source.zone, col: source.col }, { zone: 'tableau', col: dest.col }],
    };
  }
  return { error: MOVE_USAGE };
}

/** Help text for Bristol CLI mode. */
export const BRISTOL_HELP: string[] = [
  'd/draw          - Deal 1 card to each of the 3 fans',
  'm t<col> t<col> - Move tableau column to tableau column',
  'm t<col> f      - Move tableau column to foundation',
  'm n<fan> t<col> - Move fan to tableau column',
  'm n<fan> f      - Move fan to foundation',
  'ac/autocomplete - Auto-complete to foundations',
  'u/undo          - Undo last move',
  'h/hint          - Get a hint',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
