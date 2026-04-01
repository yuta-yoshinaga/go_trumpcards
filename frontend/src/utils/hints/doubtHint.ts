import type { DoubtResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { DoubtPhase } from '../../types/phases';

/** Returns a frontend HintResult for Doubt, or null if no suggestion. */
export function getDoubtHint(state: DoubtResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.isFinished) return null;
  if (state.gameEndFlag) return null;

  if (state.phase === DoubtPhase.DOUBT) {
    return getDoubtPhaseHint(state);
  }

  if (state.phase === DoubtPhase.PLAY) {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    if (state.currentTurn !== humanIdx) return null;
    return getPlayPhaseHint(human.cards);
  }

  return null;
}

/** Hint during doubt window: suggest doubt if opponent showed a tell. */
function getDoubtPhaseHint(state: DoubtResponse): HintResult | null {
  if (!state.lastAction) return null;

  if (state.lastAction.hasTell) {
    return { targetAction: 'doubt', reason: 'hint.doubtTell', confidence: 'strong' };
  }

  return { targetAction: 'skip', reason: 'hint.skipSafe', confidence: 'moderate' };
}

/** Hint during play phase: suggest playing truthfully if possible. */
function getPlayPhaseHint(cards: { value: number }[]): HintResult {
  const valueCounts = new Map<number, number>();
  for (const c of cards) {
    valueCounts.set(c.value, (valueCounts.get(c.value) ?? 0) + 1);
  }

  const hasPairs = [...valueCounts.values()].some((count) => count >= 2);
  if (hasPairs) {
    return { targetAction: 'play', reason: 'hint.playTruth', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.bluffCarefully', confidence: 'moderate' };
}
