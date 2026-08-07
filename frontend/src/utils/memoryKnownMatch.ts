import type { Card } from '../types/card';
import type { MemoryBoardCard } from '../types/games/memory';

/**
 * Position of a face-down card the player has already seen that matches the one
 * card currently face up, or `null` when there is no such position.
 *
 * `seen` holds the values the player genuinely saw on this board — the server
 * never re-sends a card once it flips back (that would be cheating), so this is
 * pure recall, not extra information (#4775).
 *
 * **表向きが1枚のときだけ答える。**2枚めくった後や0枚のときに答えると、
 * 「一致がある」という表示が実際には打てない手を指す。
 */
export function memoryKnownMatch(board: readonly MemoryBoardCard[], seen: ReadonlyMap<number, Card>): number | null {
  const faceUp = board.reduce<number[]>((acc, bc, idx) => {
    if (bc.faceUp && !bc.taken) acc.push(idx);
    return acc;
  }, []);
  if (faceUp.length !== 1) return null;

  const target = board[faceUp[0]]?.card;
  if (!target) return null;

  for (let idx = 0; idx < board.length; idx++) {
    const bc = board[idx];
    if (idx === faceUp[0] || bc.faceUp || bc.taken) continue;
    const remembered = seen.get(idx);
    if (remembered && remembered.design === target.design && remembered.value === target.value) return idx;
  }
  return null;
}
