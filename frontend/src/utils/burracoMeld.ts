import type { Card } from '../types/common';
import type { BurracoMeldData } from '../types/games/burraco';

/**
 * Reports whether a Burraco card selection can extend an existing meld or form a new one.
 * Validation order mirrors `validateNewMeld`, `validateMeldAddition`, and
 * `findExistingMeldForCards` in `internal/domain/Canasta.go` (around lines 1305, 1352, and 1386).
 */
export function burracoMeldSelectionStatus(
  selected: readonly Card[],
  playerMelds: readonly BurracoMeldData[],
): 'select' | 'valid' | 'invalid' {
  if (selected.length === 0) return 'select';

  // findExistingMeldForCards: take the rank from the first natural card and
  // choose the first meld with that rank; all-wild selections cannot extend.
  const selectedRank = selected.find((card) => !isWild(card))?.value;
  const existing = selectedRank === undefined ? undefined : playerMelds.find((meld) => meld.rank === selectedRank);
  if (existing) {
    // validateMeldAddition: rank mismatch (while checking black threes), then
    // combined wild-card limit. A wild-only selection bypasses rank matching.
    let combinedWildCount = existing.cards.filter(isWild).length;
    for (const card of selected) {
      if (isWild(card)) {
        combinedWildCount++;
      } else if (card.value !== existing.rank) {
        return 'invalid';
      }
      if (isBlackThree(card)) return 'invalid';
    }
    return combinedWildCount > 3 ? 'invalid' : 'valid';
  }

  // validateNewMeld: minimum size, same natural rank, black three, at least
  // two naturals, at most three wilds, and no more wilds than naturals.
  if (selected.length < 3) return 'invalid';
  let naturalCount = 0;
  let wildCount = 0;
  let rank = 0;
  for (const card of selected) {
    if (isWild(card)) {
      wildCount++;
    } else {
      naturalCount++;
      if (rank === 0) rank = card.value;
      else if (card.value !== rank) return 'invalid';
    }
  }
  if (selected.some(isBlackThree)) return 'invalid';
  if (naturalCount < 2) return 'invalid';
  if (wildCount > 3) return 'invalid';
  if (wildCount > naturalCount) return 'invalid';
  return 'valid';
}

function isWild(card: Card): boolean {
  return card.design === 'JOKER' || card.value === 2;
}

function isBlackThree(card: Card): boolean {
  return card.value === 3 && (card.design === 'SPADE' || card.design === 'CLOVER');
}
