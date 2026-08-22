import type { DramahaResponse } from '../../../types/card';
import { DramahaPhase } from '../../../types/phases';
import { dramahaHands } from '../../dramahaBestFive';
import { formatGenericState } from './genericFormatter';

const PHASE_NAMES: Record<number, string> = {
  [DramahaPhase.INIT]: 'INIT',
  [DramahaPhase.PRE_FLOP]: 'PRE-FLOP',
  [DramahaPhase.FLOP]: 'FLOP',
  [DramahaPhase.DRAW]: 'DRAW',
  [DramahaPhase.TURN]: 'TURN',
  [DramahaPhase.RIVER]: 'RIVER',
  [DramahaPhase.SHOWDOWN]: 'SHOWDOWN',
  [DramahaPhase.END]: 'END',
  [DramahaPhase.REBUY]: 'REBUY',
};

/** Terminal-readable name for each hand-key the evaluator returns. */
const HAND_NAMES: Record<string, string> = {
  highCard: 'high card',
  onePair: 'one pair',
  twoPair: 'two pair',
  threeOfAKind: 'three of a kind',
  straight: 'straight',
  flush: 'flush',
  fullHouse: 'full house',
  fourOfAKind: 'four of a kind',
  straightFlush: 'straight flush',
  royalFlush: 'royal flush',
};

/**
 * Format a Dramaha game state as terminal text.
 *
 * Both halves of the split are printed every render. The pot always divides
 * between them, so showing only one (as the Hold'em formatter this was cloned
 * from does) hides half of what the player is playing for. The draw hand is
 * printed from the five hole cards alone — it does not read the board, so it
 * is there before the flop, when the Omaha hand does not yet exist.
 */
export function formatDramahaState(state: DramahaResponse): string {
  const customLines: string[] = [];
  if (state.tournamentMode)
    customLines.push(`blinds: ${state.smallBlind}/${state.bigBlind} (hand #${state.handCount})`);
  customLines.push('Pot always splits: Omaha hand (2 hole + 3 board) / draw hand (all five as dealt)');

  const human = state.players?.find((p) => p.isHuman);
  if (human && !human.folded) {
    const { omaha, draw } = dramahaHands(human.cards ?? [], state.communityCards ?? []);
    if (omaha) customLines.push(`Omaha hand: ${HAND_NAMES[omaha.key] ?? omaha.key}`);
    if (draw) customLines.push(`Draw hand: ${HAND_NAMES[draw.key] ?? draw.key}`);
  }

  if (state.phase === DramahaPhase.DRAW) customLines.push('Draw round: "d 1 3" to exchange, bare "d" to stand pat');
  if (state.rebuyAvailable) customLines.push('Rebuy available!');
  if (state.addonAvailable) customLines.push('Add-on available!');
  if (state.muckAvailable) customLines.push('Muck or Show?');

  return formatGenericState({
    title: 'Dramaha',
    players: state.players.map((p) => ({
      id: p.id,
      isHuman: p.isHuman,
      cards: p.cards,
      chips: p.chips,
      folded: p.folded,
      allIn: p.allIn,
      handName: state.gameEndFlag ? p.handName : undefined,
    })),
    phase: state.phase,
    phaseNames: PHASE_NAMES,
    currentTurn: state.currentTurn,
    pot: state.pot,
    communityCards: state.communityCards,
    gameEndFlag: state.gameEndFlag,
    message: state.message,
    scoreMode: 'chips',
    customLines,
  });
}
