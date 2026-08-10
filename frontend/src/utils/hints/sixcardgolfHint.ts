import type { SixCardGolfResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Card values considered "low" (K=13→0pts, A=1→1pt, 2→2pts, 3→3pts). */
const LOW_VALUES = new Set([13, 1, 2, 3]);
/** Card values considered "high" (J=11→10pts, Q=12→10pts, 8-10→face value). */
const HIGH_VALUES = new Set([8, 9, 10, 11, 12]);

/** Returns a frontend HintResult for Six Card Golf, or null if no suggestion. */
export function getSixcardgolfHint(state: SixCardGolfResponse): HintResult | null {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx === -1) return null;

  const human = state.players[humanIdx];

  // Phase 3 (roundOver) and phase 4 (gameOver): no hints
  if (state.phase >= 3) return null;

  // Phase 0 (setup): hint to flip cards
  if (state.phase === 0) {
    return null;
  }

  // Phase 1 (playerTurn): suggest drawing
  if (state.phase === 1) {
    if (state.currentPlayerIdx !== humanIdx) return null;

    // If canFlip, suggest flipping a face-down card
    if (state.canFlip) {
      return { targetAction: 'flip', reason: 'hintReason.flipCard', confidence: 'moderate' };
    }

    // Check discard top value
    if (state.discardTop && LOW_VALUES.has(state.discardTop.value)) {
      return { targetAction: 'drawDiscard', reason: 'hintReason.drawDiscardLow', confidence: 'strong' };
    }

    return { targetAction: 'drawStock', reason: 'hintReason.drawStock', confidence: 'moderate' };
  }

  // Phase 2 (drawPending): suggest swap or discard
  if (state.phase === 2) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    if (!state.drawnCard) return null;

    const drawnValue = state.drawnCard.value;

    // Check column match opportunity: if drawn card matches a visible card in the same column
    const columnMatchIdx = findColumnMatchSwap(human.grid, state.drawnCard.value);
    if (columnMatchIdx !== -1) {
      return {
        targetAction: 'swap',
        reason: 'hintReason.columnMatch',
        confidence: 'strong',
        targetPos: columnMatchIdx,
      };
    }

    // Low drawn card: suggest swapping with highest visible or face-down
    if (LOW_VALUES.has(drawnValue)) {
      const highVisibleIdx = findHighestVisibleCard(human.grid);
      if (highVisibleIdx !== -1) {
        return {
          targetAction: 'swap',
          reason: 'hintReason.swapHigh',
          confidence: 'strong',
          targetPos: highVisibleIdx,
        };
      }
      const faceDownIdx = human.grid.findIndex((s) => !s.faceUp);
      if (faceDownIdx !== -1) {
        return {
          targetAction: 'swap',
          reason: 'hintReason.swapFaceDown',
          confidence: 'moderate',
          targetPos: faceDownIdx,
        };
      }
    }

    // High drawn card: suggest discarding
    if (HIGH_VALUES.has(drawnValue)) {
      return { targetAction: 'discard', reason: 'hintReason.discardHigh', confidence: 'strong' };
    }

    // Medium value: suggest swapping with face-down if available
    const faceDownIdx = human.grid.findIndex((s) => !s.faceUp);
    if (faceDownIdx !== -1) {
      return {
        targetAction: 'swap',
        reason: 'hintReason.swapFaceDown',
        confidence: 'moderate',
        targetPos: faceDownIdx,
      };
    }

    return { targetAction: 'discard', reason: 'hintReason.discardHigh', confidence: 'moderate' };
  }

  return null;
}

/** Find a grid slot to swap that would create a column match (top/bottom with same rank).
 *  The grid is 6 slots: [0,1,2] top row, [3,4,5] bottom row. Columns: (0,3), (1,4), (2,5).
 *  Returns the index of the slot to swap into for a column match, or -1 if none. */
function findColumnMatchSwap(grid: SixCardGolfResponse['players'][0]['grid'], drawnValue: number): number {
  const columns = [
    [0, 3],
    [1, 4],
    [2, 5],
  ];
  for (const [top, bottom] of columns) {
    // If top is face up and matches drawn value, suggest swapping bottom
    if (grid[top].faceUp && grid[top].card && grid[top].card.value === drawnValue) {
      // Only suggest if bottom is not already matching
      if (!grid[bottom].faceUp || !grid[bottom].card || grid[bottom].card.value !== drawnValue) {
        return bottom;
      }
    }
    // If bottom is face up and matches drawn value, suggest swapping top
    if (grid[bottom].faceUp && grid[bottom].card && grid[bottom].card.value === drawnValue) {
      if (!grid[top].faceUp || !grid[top].card || grid[top].card.value !== drawnValue) {
        return top;
      }
    }
  }
  return -1;
}

/** Find the index of the highest-value visible card in the grid. Returns -1 if none visible. */
function findHighestVisibleCard(grid: SixCardGolfResponse['players'][0]['grid']): number {
  let bestIdx = -1;
  let bestValue = -1;
  for (let i = 0; i < grid.length; i++) {
    const slot = grid[i];
    if (slot.faceUp && slot.card) {
      // Score: K=0, A=1, 2-10=face, J=10, Q=10
      const score = cardScore(slot.card.value);
      if (score > bestValue) {
        bestValue = score;
        bestIdx = i;
      }
    }
  }
  return bestIdx;
}

/** Convert a card value to its Six Card Golf scoring value. */
function cardScore(value: number): number {
  if (value === 13) return 0; // K = 0
  if (value === 1) return 1; // A = 1
  if (value >= 11) return 10; // J, Q = 10
  return value; // 2-10 = face value
}
