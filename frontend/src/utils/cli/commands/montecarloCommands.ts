import type { montecarloApi } from '../../../api/games/montecarlo';
import type { MonteCarloResponse } from '../../../types/card';
import { MonteCarloPhase } from '../../../types/phases';
import { canMonteCarloDeal, canMonteCarloUndo, isMonteCarloAdjacent } from '../../monteCarloRules';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MonteCarloArgs = Parameters<typeof montecarloApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'remove',
  'd',
  'deal',
  'u',
  'undo',
  'g',
  'giveup',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Monte Carlo Solitaire command into API exec arguments. */
export function parseMonteCarloCommand(
  input: string,
  state?: MonteCarloResponse | null,
): CliParseResult<MonteCarloArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play':
    case 'remove': {
      if (args.length !== 2) return { error: 'Usage: p <idx1> <idx2> (0-24)' };
      const first = parseIntArg(args, 0);
      const second = parseIntArg(args, 1);
      if ('error' in first || 'error' in second) return { error: 'Usage: p <idx1> <idx2> (0-24)' };
      if ([first.value, second.value].some((value) => value < 0 || value > 24)) {
        return { error: 'Cell indices must be between 0 and 24' };
      }
      if (first.value === second.value) return { error: 'Choose two different cells' };
      if (state?.phase !== MonteCarloPhase.PLAYING)
        return { error: 'You can only remove cards while the game is in progress' };
      const a = { row: Math.floor(first.value / 5), col: first.value % 5 };
      const b = { row: Math.floor(second.value / 5), col: second.value % 5 };
      const firstCard = state.board[a.row]?.[a.col]?.card;
      const secondCard = state.board[b.row]?.[b.col]?.card;
      if (!firstCard || !secondCard) return { error: 'Both selected cells must contain a card' };
      if (!isMonteCarloAdjacent(a, b)) return { error: 'The selected cells must be adjacent' };
      if (firstCard.value !== secondCard.value) return { error: 'The selected cards must have the same rank' };
      return {
        args: ['remove', Math.floor(first.value / 5), first.value % 5, Math.floor(second.value / 5), second.value % 5],
      };
    }
    case 'd':
    case 'deal':
      if (!canMonteCarloDeal(state)) return { error: 'You can deal only while playing with cards left in the stock' };
      return { args: ['deal'] };
    case 'u':
    case 'undo':
      if (state?.phase !== MonteCarloPhase.PLAYING) return { error: 'You can only undo while the game is in progress' };
      if (!canMonteCarloUndo(state)) return { error: 'There is no action to undo' };
      return { args: ['undo'] };
    case 'g':
    case 'giveup':
      if (state?.phase !== MonteCarloPhase.PLAYING)
        return { error: 'You can only give up while the game is in progress' };
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      if (state?.phase !== MonteCarloPhase.PLAYING)
        return { error: 'Hints are available only while the game is in progress' };
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in Monte Carlo CLI mode. */
export const MONTECARLO_HELP = [
  'p <idx1> <idx2> - Remove an adjacent same-rank pair (board indices 0-24)',
  'd / deal - Deal five cards from the stock',
  'u / undo - Undo the last action',
  'g / giveup - Give up the game',
  'h / hint - Show a hint',
  'r / reset - Reset the game',
];
