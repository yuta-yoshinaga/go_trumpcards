/**
 * Returns whether a Crazy Quilt cell contains a vertically oriented card.
 *
 * domain_rule_quoted:
 * 	if idx < 0 || idx >= CrazyQuiltCells {
 * 		return false
 * 	}
 * 	return (idx/CrazyQuiltGridSize+idx%CrazyQuiltGridSize)%2 == 0
 *
 * The Go domain rule alternates orientation by the parity of the row and
 * column sum, with out-of-range indexes treated as horizontal.
 */
export function isCrazyQuiltVertical(idx: number): boolean {
  if (idx < 0 || idx >= 64) return false;
  return (Math.floor(idx / 8) + (idx % 8)) % 2 === 0;
}
