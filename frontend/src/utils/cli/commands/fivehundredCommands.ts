import type { fiveHundredApi } from '../../../api/gameApi';
import type { FiveHundredResponse } from '../../../types/card';
import { FiveHundredContract } from '../../../types/phases';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by fiveHundredApi.exec. */
export type FiveHundredCliArgs = Parameters<typeof fiveHundredApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'bnt',
  'm',
  'misere',
  'om',
  'openmisere',
  'pa',
  'pass',
  'e',
  'exchange',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parses a single CLI command line for the 500 (Five Hundred) game. */
export function parseFiveHundredCommand(input: string): CliParseResult<FiveHundredCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const tricks = Number.parseInt(args[0] ?? '', 10);
      const suit = Number.parseInt(args[1] ?? '', 10);
      if (Number.isNaN(tricks) || Number.isNaN(suit))
        return { error: 'Usage: b <tricks> <suit> (suit 1=S 2=C 3=H 4=D)' };
      return { args: ['bid', { bidKind: FiveHundredContract.SUIT, bidTricks: tricks, bidSuit: suit }] };
    }
    case 'bnt': {
      const tricks = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(tricks)) return { error: 'Usage: bnt <tricks>' };
      return { args: ['bid', { bidKind: FiveHundredContract.NO_TRUMP, bidTricks: tricks }] };
    }
    case 'm':
    case 'misere':
      return { args: ['bid', { bidKind: FiveHundredContract.MISERE }] };
    case 'om':
    case 'openmisere':
      return { args: ['bid', { bidKind: FiveHundredContract.OPEN_MISERE }] };
    case 'pa':
    case 'pass':
      return { args: ['pass'] };
    case 'e':
    case 'exchange': {
      const idxs = args.map((a) => Number.parseInt(a, 10));
      if (idxs.length !== 3 || idxs.some(Number.isNaN)) return { error: 'Usage: e <i> <j> <k>' };
      return { args: ['exchange', { discardIndices: idxs }] };
    }
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx)) return { error: 'Usage: p <idx> [jokerSuit]' };
      const jokerSuit = args[1] !== undefined ? Number.parseInt(args[1], 10) : undefined;
      return { args: ['play', { cardIndex: idx, jokerSuit }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
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

/** Renders a 500 game state as CLI-friendly text. */
export function formatFiveHundredState(s: FiveHundredResponse): string {
  const lines: string[] = [];
  lines.push(`Round ${s.roundNumber} · Trick ${s.trickNumber}`);
  lines.push(`Scores: Team0 ${s.teamScores[0]} / Team1 ${s.teamScores[1]}`);
  for (const p of s.players) {
    const tag = p.isHuman ? 'You' : `CPU${p.id}`;
    const role = p.isDeclarer ? ' [declarer]' : p.passed ? ' [pass]' : '';
    lines.push(`${tag} (T${p.team}): ${p.cardCount} cards, ${p.trickCount} tricks${role}`);
  }
  if (s.currentTrick.length > 0) {
    lines.push(`Trick: ${s.currentTrick.map((tc) => `${tc.card.value}${tc.card.design}`).join(' ')}`);
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}

/** Help text shown in the CLI terminal for 500 (Five Hundred). */
export const FIVE_HUNDRED_HELP = [
  'b <tricks> <suit>  - Bid a suit (suit 1=S 2=C 3=H 4=D)',
  'bnt <tricks>       - Bid no-trump',
  'm / om             - Bid misere / open misere',
  'pa                 - Pass',
  'e <i> <j> <k>      - Exchange kitty (discard 3)',
  'p <idx> [suit]     - Play a card (suit = joker nomination in NT)',
  'n / nr             - Next trick / next round',
  'r/reset            - Reset game',
  'h/hint             - Get a hint',
] as const;
