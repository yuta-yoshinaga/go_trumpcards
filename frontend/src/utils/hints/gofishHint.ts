import type { Card, GoFishCpuAction, GoFishLastAsk, GoFishResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GoFishPhase } from '../../types/phases';
import { valueName } from '../cardUtils';

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
    return {
      targetAction: 'ask',
      reason: 'hint.askKnownRank',
      confidence: 'strong',
      ...pointAtRank(human.cards, knownOpponentRank),
    };
  }

  return {
    targetAction: 'ask',
    reason: 'hint.askMostCopies',
    confidence: 'moderate',
    ...pointAtRank(human.cards, mostCopiesRank(human.cards)),
  };
}

/**
 * Name a rank and point at every card of it in the hand.
 *
 * Both reasons promise a specific rank -- "the rank you hold most copies of",
 * "the rank an opponent recently took" -- and the hint used to return neither
 * the rank nor the cards, leaving the player to count the hand again (#5518).
 */
function pointAtRank(cards: Card[], rank: number): Pick<HintResult, 'reasonParams' | 'targetIndices'> {
  const targetIndices = cards.map((c, i) => (c.value === rank ? i : -1)).filter((i) => i >= 0);
  return { reasonParams: { rank: valueName(rank) }, targetIndices };
}

/**
 * The rank held in the most copies, ties broken by the lowest rank.
 *
 * **The tie-break has to be on the rank, not on hand order.** Picking the first
 * one encountered makes the same board recommend a different rank once the hand
 * is re-sorted.
 */
function mostCopiesRank(cards: Card[]): number {
  const counts = new Map<number, number>();
  for (const c of cards) {
    counts.set(c.value, (counts.get(c.value) ?? 0) + 1);
  }
  let best = cards[0].value;
  let bestCount = 0;
  for (const [rank, count] of [...counts].sort((a, b) => a[0] - b[0])) {
    if (count > bestCount) {
      best = rank;
      bestCount = count;
    }
  }
  return best;
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

  // 若い順に見る。手札の並び順に任せると、同じ盤面でも並べ替えただけで
  // 勧めるランクが変わる。
  const matches = [...humanRanks].filter((rank) => knownRanks.has(rank));
  return matches.length > 0 ? Math.min(...matches) : null;
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
