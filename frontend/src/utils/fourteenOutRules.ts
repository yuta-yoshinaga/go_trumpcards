import type { FourteenOutResponse } from '../types/card';

/** Returns the exposed top card of a column, or null when the column is empty. */
export function fourteenOutTail(col: FourteenOutResponse['columns'][number] | undefined) {
  return col?.at(-1)?.card ?? null;
}

/** Returns the columns that can be removed with the selected column. */
export function fourteenOutPartners(state: FourteenOutResponse, col: number): number[] {
  return state.legalTargets[`tableau-${col}`]?.tableau ?? [];
}

/** Reports whether the current game state allows undo. */
export function fourteenOutCanUndo(state: FourteenOutResponse): boolean {
  return state.canUndo;
}

/** Lists every removable column pair once in numeric column order. */
export function fourteenOutRemovablePairs(state: FourteenOutResponse): Array<[number, number]> {
  const pairs = new Map<string, [number, number]>();
  for (let from = 0; from < state.columns.length; from++) {
    for (const to of fourteenOutPartners(state, from)) {
      if (from < to) pairs.set(`${from}:${to}`, [from, to]);
    }
  }
  return [...pairs.values()].sort(([a1, a2], [b1, b2]) => a1 - b1 || a2 - b2);
}
