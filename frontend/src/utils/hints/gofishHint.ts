import type { GoFishCpuAction, GoFishLastAsk, GoFishResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GoFishPhase } from '../../types/phases';

/** Returns a frontend HintResult for Go Fish, or null if no suggestion. */
export function getGoFishHint(state: GoFishResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.phase !== GoFishPhase.PLAY) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0 || state.currentTurn !== humanIdx) return null;

  const human = state.players[humanIdx];
  if (!human || human.cards.length === 0) return null;

  const knownOpponentRank = findOpponentKnownRank(state, humanIdx);
  if (knownOpponentRank !== null) {
    return { targetAction: 'ask', reason: 'hint.askKnownRank', confidence: 'strong' };
  }

  return { targetAction: 'ask', reason: 'hint.askMostCopies', confidence: 'moderate' };
}

/**
 * Look for a rank the human holds that a previous action revealed an opponent also holds.
 * When a CPU asked the human for rank R and failed, or when a CPU drew from the deck and did not
 * form a book, we cannot be sure; but when a CPU successfully took rank R from another player we
 * know that CPU held R recently.
 */
function findOpponentKnownRank(state: GoFishResponse, humanIdx: number): number | null {
  const human = state.players[humanIdx];
  if (!human) return null;
  const humanRanks = new Set(human.cards.map((c) => c.value));

  const knownRanks = new Set<number>();
  collectFromLastAsk(state.lastAsk, humanIdx, knownRanks);
  for (const action of state.cpuActions) {
    collectFromCpuAction(action, humanIdx, knownRanks);
  }

  for (const rank of humanRanks) {
    if (knownRanks.has(rank)) return rank;
  }
  return null;
}

function collectFromLastAsk(lastAsk: GoFishLastAsk | null, humanIdx: number, out: Set<number>): void {
  if (!lastAsk) return;
  if (lastAsk.playerIdx === humanIdx) return;
  if (lastAsk.success && lastAsk.targetIdx !== humanIdx) {
    out.add(lastAsk.rank);
  }
}

function collectFromCpuAction(action: GoFishCpuAction, humanIdx: number, out: Set<number>): void {
  if (action.askPlayerIdx === humanIdx) return;
  if (action.success && action.askTargetIdx !== humanIdx) {
    out.add(action.askRank);
  }
}
