import type { rummy500Api } from '../../../api/games/rummy500';
import type { Rummy500Response } from '../../../types/card';
import { Rummy500Phase } from '../../../types/phases';
import { classifyRummy500Meld } from '../../rummy500MeldValidator';
import { parseIntSlice, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Arguments accepted by the Rummy 500 API exec function. */
export type Rummy500CliArgs = Parameters<typeof rummy500Api.exec>;

function requirePhase(state: Rummy500Response, phase: number, action: string): string | null {
  if (state.phase === phase) return null;
  const phaseName = phase === Rummy500Phase.DRAW ? 'draw' : phase === Rummy500Phase.PLAY ? 'play' : 'round-end';
  return `Cannot ${action} during the ${state.phase === Rummy500Phase.DRAW ? 'draw' : state.phase === Rummy500Phase.PLAY ? 'play' : state.phase === Rummy500Phase.ROUND_END ? 'round-end' : 'game-end'} phase; this command is only available during the ${phaseName} phase.`;
}

const VALID_COMMANDS = [
  'd',
  'draw',
  'stock',
  't',
  'take',
  'm',
  'meld',
  'x',
  'discard',
  'p',
  'play',
  'lo',
  'layoff',
  'nr',
  'nextround',
  'r',
  'reset',
];

/** Parse a Rummy 500 terminal command into API exec arguments. */
export function parseRummy500Command(input: string, state: Rummy500Response): CliParseResult<Rummy500CliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'd':
    case 'draw':
    case 'stock': {
      const phaseError = requirePhase(state, Rummy500Phase.DRAW, 'draw');
      if (phaseError) return { error: phaseError };
      return { args: ['drawstock'] };
    }
    case 't':
    case 'take': {
      const phaseError = requirePhase(state, Rummy500Phase.DRAW, 'take from the discard pile');
      if (phaseError) return { error: phaseError };
      const idx = Number(args[0]);
      if (args.length === 0 || !Number.isInteger(idx) || idx < 0) return { error: 'Usage: take <discard-index>' };
      if (idx >= state.discardPile.length)
        return { error: `Discard index ${idx} is outside the pile (0-${state.discardPile.length - 1}).` };
      return { args: ['drawdiscard', undefined, undefined, undefined, idx] };
    }
    case 'm':
    case 'meld': {
      const phaseError = requirePhase(state, Rummy500Phase.PLAY, 'meld');
      if (phaseError) return { error: phaseError };
      const parsed = parseIntSlice(args);
      if ('error' in parsed || parsed.values.length < 3 || parsed.values.some((i) => i < 0))
        return { error: 'Usage: meld <idx> <idx> <idx> [idx ...]' };
      if (new Set(parsed.values).size !== parsed.values.length) return { error: 'Meld indices must be unique.' };
      const human = state.players.find((player) => player.isHuman);
      if (!human || parsed.values.some((i) => i >= human.cards.length))
        return { error: 'A meld hand index is outside your hand.' };
      if (!classifyRummy500Meld(parsed.values.map((i) => human.cards[i])).valid)
        return { error: 'Selected cards do not form a valid meld.' };
      return { args: ['meld', undefined, undefined, parsed.values] };
    }
    case 'x':
    case 'discard':
    case 'p':
    case 'play': {
      const phaseError = requirePhase(state, Rummy500Phase.PLAY, 'discard');
      if (phaseError) return { error: phaseError };
      const idx = Number(args[0]);
      if (args.length === 0 || !Number.isInteger(idx) || idx < 0) return { error: 'Usage: discard <idx>' };
      const human = state.players.find((player) => player.isHuman);
      if (!human || idx >= human.cards.length)
        return {
          error: human?.cards.length
            ? `Hand index ${idx} is outside your hand (0-${human.cards.length - 1}).`
            : 'Your hand is empty; there are no cards to discard.',
        };
      return { args: ['discard', idx] };
    }
    case 'lo':
    case 'layoff': {
      const phaseError = requirePhase(state, Rummy500Phase.PLAY, 'lay off a card');
      if (phaseError) return { error: phaseError };
      const nums = args.map(Number);
      if (args.length !== 3 || nums.some((n) => !Number.isInteger(n) || n < 0))
        return { error: 'Usage: layoff <hand-idx> <owner> <meld-idx>' };
      const human = state.players.find((player) => player.isHuman);
      if (!human || nums[0] >= human.cards.length) return { error: `Hand index ${nums[0]} is outside your hand.` };
      if (!state.layoffTargets?.[nums[0]]?.some((target) => target.owner === nums[1] && target.meldIdx === nums[2]))
        return { error: 'That meld is not a legal layoff target for this card.' };
      return {
        args: [
          'layoff',
          undefined,
          undefined,
          undefined,
          undefined,
          { cardIndex: nums[0], meldOwner: nums[1], meldIdx: nums[2] },
        ],
      };
    }
    case 'nr':
    case 'nextround':
      {
        const phaseError = requirePhase(state, Rummy500Phase.ROUND_END, 'start the next round');
        if (phaseError) return { error: phaseError };
      }
      return { args: ['nextround'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Rummy 500 terminal. */
export const RUMMY500_HELP = [
  'd / draw / stock  - Draw from stock',
  't / take <idx>    - Draw from discard at index and above',
  'm / meld <idx> ... - Meld selected hand cards (3 or more)',
  'x / discard <idx> - Discard a hand card',
  'p / play <idx>    - Discard a hand card',
  'lo / layoff <hand> <owner> <meld> - Lay off hand card 0 on player 1 meld 2',
  'nr / nextround    - Start the next round',
  'r / reset         - Reset game',
];
