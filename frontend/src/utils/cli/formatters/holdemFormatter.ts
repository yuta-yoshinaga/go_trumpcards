import type { HoldemResponse } from '../../../types/card';
import { HoldemPhase } from '../../../types/phases';
import { formatGenericState } from './genericFormatter';

const PHASE_NAMES: Record<number, string> = {
  [HoldemPhase.INIT]: 'INIT',
  [HoldemPhase.PRE_FLOP]: 'PRE-FLOP',
  [HoldemPhase.FLOP]: 'FLOP',
  [HoldemPhase.TURN]: 'TURN',
  [HoldemPhase.RIVER]: 'RIVER',
  [HoldemPhase.SHOWDOWN]: 'SHOWDOWN',
  [HoldemPhase.END]: 'END',
  [HoldemPhase.REBUY]: 'REBUY',
};

/** Format a Texas Hold'em game state as terminal text. */
export function formatHoldemState(state: HoldemResponse): string {
  const customLines: string[] = [];
  if (state.tournamentMode)
    customLines.push(`blinds: ${state.smallBlind}/${state.bigBlind} (hand #${state.handCount})`);
  if (state.rebuyAvailable) customLines.push('Rebuy available!');
  if (state.addonAvailable) customLines.push('Add-on available!');
  if (state.muckAvailable) customLines.push('Muck or Show?');

  return formatGenericState({
    title: "Texas Hold'em",
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
