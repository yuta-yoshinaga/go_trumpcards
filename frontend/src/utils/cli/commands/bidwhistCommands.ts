import type { bidWhistApi } from '../../../api/gameApi';
import type { BidWhistResponse } from '../../../types/card';
import { BidWhistDirection } from '../../../types/phases';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by bidWhistApi.exec. */
export type BidWhistCliArgs = Parameters<typeof bidWhistApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pa',
  'pass',
  't',
  'trump',
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

/** Parses a single CLI command line for the Bid Whist game. */
export function parseBidWhistCommand(input: string): CliParseResult<BidWhistCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const tricks = Number.parseInt(args[0] ?? '', 10);
      const dir = Number.parseInt(args[1] ?? '', 10);
      if (Number.isNaN(tricks) || Number.isNaN(dir))
        return { error: 'Usage: b <tricks> <dir> (dir 0=Uptown 1=Downtown 2=NoTrump)' };
      return { args: ['bid', { bidTricks: tricks, bidDirection: dir }] };
    }
    case 'pa':
    case 'pass':
      return { args: ['pass'] };
    case 't':
    case 'trump': {
      const suit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(suit)) return { error: 'Usage: t <suit> (1=S 2=C 3=H 4=D)' };
      return { args: ['trump', { trumpSuit: suit }] };
    }
    case 'e':
    case 'exchange': {
      const idxs = args.map((a) => Number.parseInt(a, 10));
      if (idxs.length !== 6 || idxs.some(Number.isNaN)) return { error: 'Usage: e <i1> <i2> <i3> <i4> <i5> <i6>' };
      return { args: ['exchange', { discardIndices: idxs }] };
    }
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx)) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: idx }] };
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

const DIRECTION_LABELS: Record<number, string> = {
  [BidWhistDirection.UPTOWN]: 'Uptown',
  [BidWhistDirection.DOWNTOWN]: 'Downtown',
  [BidWhistDirection.NO_TRUMP]: 'No Trump',
};

/** Renders a Bid Whist game state as CLI-friendly text. */
export function formatBidWhistState(s: BidWhistResponse): string {
  const lines: string[] = [];
  lines.push(`Round ${s.roundNumber} · Trick ${s.trickNumber}`);
  lines.push(`Scores: Team0 ${s.teamScores[0]} / Team1 ${s.teamScores[1]}`);
  if (s.declarerIdx >= 0) {
    lines.push(`Contract: ${s.contractTricks} ${DIRECTION_LABELS[s.contractDirection] ?? '?'}`);
  }
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

/** Help text shown in the CLI terminal for Bid Whist. */
export const BID_WHIST_HELP = [
  'b <tricks> <dir>   - Bid (dir 0=Uptown 1=Downtown 2=NoTrump)',
  'pa                 - Pass',
  't <suit>           - Declare trump (1=S 2=C 3=H 4=D)',
  'e <i1..i6>         - Exchange kitty (discard 6)',
  'p <idx>            - Play a card',
  'n / nr             - Next trick / next round',
  'r/reset            - Reset game',
  'h/hint             - Get a hint',
] as const;
