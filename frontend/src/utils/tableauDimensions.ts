/** Derive card dimensions from the width actually available to a tableau row. */
export function tableauDimensions(rowWidth: number, columns: number, gap: number, overlapRatio: number) {
  const cardWidth = (rowWidth - (columns - 1) * gap) / columns;
  return {
    cardWidth,
    cardHeight: Math.round(cardWidth * 1.5),
    cardOverlap: Math.round(cardWidth * overlapRatio),
  };
}
