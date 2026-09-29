import type { RankAndFileMoveZone } from '../api/gameApi';
import type { RankAndFileResponse } from '../types/card';
import { rankAndFileFoundationTarget } from './rankAndFileFoundationTarget';

/**
 * Count legal tableau columns and the foundation destination for a selected source.
 * Mirrors RankAndFile.MoveWasteToTableau / MoveTableauToTableau and
 * MoveWasteToFoundation / MoveTableauToFoundation in internal/domain/RankAndFile.go.
 */
export function rankAndFileLegalTargetCount(state: RankAndFileResponse, source: RankAndFileMoveZone): number {
  const isWaste = source.zone === 'waste';
  const col = source.col ?? -1;
  const cardIndex = source.cardIndex ?? -1;
  const card = isWaste ? state.waste[state.waste.length - 1] : state.tableau[col]?.[cardIndex]?.card;
  if (!card) return 0;

  const sequenceIsMovable = isWaste || (state.sequenceStarts[col] ?? []).includes(cardIndex);
  if (!sequenceIsMovable) return 0;

  let count = 0;
  for (let targetCol = 0; targetCol < state.tableau.length; targetCol++) {
    if (!isWaste && targetCol === col) continue;
    const target = state.tableau[targetCol] ?? [];
    const top = target[target.length - 1];
    if (
      target.length === 0 ||
      (top?.faceUp && top.card && isOppositeColor(card.design, top.card.design) && card.value === top.card.value - 1)
    ) {
      count++;
    }
  }

  const isWasteTop = isWaste;
  const isTableauTop = !isWaste && cardIndex === (state.tableau[col]?.length ?? 0) - 1;
  if ((isWasteTop || isTableauTop) && rankAndFileFoundationTarget(card, state.foundation)) count++;
  return count;
}

function isOppositeColor(a: string, b: string): boolean {
  const red = (design: string) => design === 'HEART' || design === 'DIAMOND';
  return red(a) !== red(b);
}
