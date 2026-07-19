/** Cardinal direction for 2D grid keyboard navigation. */
export type GridDir = 'left' | 'right' | 'up' | 'down';

/**
 * Step one cell from `index` in `dir` within a `cols`-wide grid of `total`
 * cells. Returns the neighbour index, or `null` when the move would leave the
 * grid (past a row/column edge or beyond the last cell).
 */
function step(index: number, dir: GridDir, cols: number, total: number): number | null {
  const col = index % cols;
  switch (dir) {
    case 'left':
      return col > 0 ? index - 1 : null;
    case 'right': {
      const next = index + 1;
      return col < cols - 1 && next < total ? next : null;
    }
    case 'up': {
      const next = index - cols;
      return next >= 0 ? next : null;
    }
    case 'down': {
      const next = index + cols;
      return next < total ? next : null;
    }
  }
}

/**
 * Move focus across a 2D card grid, clamped at the edges and skipping cells for
 * which `isSkipped` returns true (e.g. taken or face-up cards). Advances one
 * step at a time in `dir`, hopping over skipped cells, and returns the new focus
 * index — or the original `index` when no landable cell exists in that
 * direction.
 *
 * @param index Current focus index (0-based).
 * @param dir Direction to move.
 * @param cols Number of columns in the grid.
 * @param total Total number of cells.
 * @param isSkipped Predicate marking non-focusable cells; defaults to none.
 */
export function moveFocus(
  index: number,
  dir: GridDir,
  cols: number,
  total: number,
  isSkipped: (i: number) => boolean = () => false,
): number {
  if (cols <= 0 || total <= 0) return index;
  let cur = index;
  for (;;) {
    const next = step(cur, dir, cols, total);
    if (next === null) return index;
    if (!isSkipped(next)) return next;
    cur = next;
  }
}
