import type { machiavelliApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MachiavelliArgs = Parameters<typeof machiavelliApi.exec>;

const VALID_COMMANDS = [
  'dr',
  'draw',
  'nm',
  'newmeld',
  'lo',
  'layoff',
  'ra',
  'rearrange',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const SUIT_LETTERS: Readonly<Record<string, number>> = { s: 1, c: 2, h: 3, d: 4 };
const RANK_LETTERS: Readonly<Record<string, number>> = { a: 1, j: 11, q: 12, k: 13 };
const RA_USAGE = 'Usage: ra <groups> / <hand indices>  (e.g. ra s5,h5,d5;c7,c8,c9 / 2,4)';

/** Parse one `<suit><rank>` token (e.g. `s5`, `hK`) into a card reference. */
function parseMachiavelliCard(token: string): { design: number; value: number } | null {
  const t = token.trim().toLowerCase();
  if (t.length < 2) return null;
  const design = SUIT_LETTERS[t[0]];
  if (design === undefined) return null;
  const rank = t.slice(1);
  const letter = RANK_LETTERS[rank];
  if (letter !== undefined) return { design, value: letter };
  if (!/^\d+$/.test(rank)) return null;
  const value = Number(rank);
  if (value < 1 || value > 13) return null;
  return { design, value };
}

/**
 * Parse the `ra` arguments: the whole new table as `;`-separated groups, then
 * `/`, then the hand indices to play. Mirrors `parseMachiavelliRearrange` in
 * `internal/adapter/controller/machiavelli_rearrange.go`.
 */
function parseMachiavelliRearrange(
  args: string[],
): { tableMelds: { design: number; value: number }[][]; handIndices: number[] } | { error: string } {
  const [left, right, ...rest] = args.join(' ').split('/');
  if (right === undefined || rest.length > 0) return { error: RA_USAGE };
  const tableMelds: { design: number; value: number }[][] = [];
  for (const group of left.split(';')) {
    if (group.trim() === '') continue;
    const meld: { design: number; value: number }[] = [];
    for (const token of group.split(',')) {
      if (token.trim() === '') continue;
      const card = parseMachiavelliCard(token);
      if (card === null) return { error: RA_USAGE };
      meld.push(card);
    }
    if (meld.length === 0) return { error: RA_USAGE };
    tableMelds.push(meld);
  }
  if (tableMelds.length === 0) return { error: RA_USAGE };
  const handIndices: number[] = [];
  for (const token of right.split(',')) {
    if (token.trim() === '') continue;
    if (!/^\d+$/.test(token.trim())) return { error: RA_USAGE };
    handIndices.push(Number(token.trim()));
  }
  if (handIndices.length === 0) return { error: RA_USAGE };
  return { tableMelds, handIndices };
}

/** Parse a Machiavelli CLI command into API exec arguments. */
export function parseMachiavelliCommand(input: string): CliParseResult<MachiavelliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'dr':
    case 'draw':
      return { args: ['draw'] };
    case 'nm':
    case 'newmeld': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: nm <i j k...> (3+ hand indices)' };
      if (parsed.values.length < 3) return { error: 'Usage: nm <i j k...> (a meld needs 3+ cards)' };
      return { args: ['newmeld', { handIndices: parsed.values }] };
    }
    case 'lo':
    case 'layoff': {
      const meld = parseIntArg(args, 0);
      const hand = parseIntArg(args, 1);
      if ('error' in meld || 'error' in hand) return { error: 'Usage: lo <meldIdx handIdx>' };
      return { args: ['layoff', { meldIdx: meld.value, handIndex: hand.value }] };
    }
    case 'ra':
    case 'rearrange': {
      // 場を組み替える手は「組み替え後の場の全体」と「出す手札」の両方が要る
      // (バックエンドの applyPlay が保存則を見るため)。CUI と同じ書式。
      const parsed = parseMachiavelliRearrange(args);
      if ('error' in parsed) return { error: parsed.error };
      return { args: ['play', { tableMelds: parsed.tableMelds, handIndices: parsed.handIndices }] };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
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

/** Help text for Machiavelli CLI mode. */
export const MACHIAVELLI_HELP: string[] = [
  'dr/draw          - Draw one card from stock (ends turn)',
  'nm <i j k...>    - Form a new meld from your hand',
  'lo <meld hand>   - Lay off a hand card onto a table meld',
  'ra <groups>/<i>  - Rebuild the table while playing from hand (ra s5,h5,d5;c7,c8,c9 / 2,4)',
  'nr/nextround     - Next round',
  'log              - Show action log',
  'r/reset          - Reset game',
];
