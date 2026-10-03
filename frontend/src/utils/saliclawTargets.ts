import type { SalicLawResponse } from '../types/card';

/**
 * Mirrors SalicLaw.canPlaceOnFoundation (SalicLaw.go:452) and
 * SalicLaw.canPlaceOnTableau (SalicLaw.go:445).
 */
export function isSalicLawTarget(
  state: SalicLawResponse,
  source: number,
  zone: 'foundation' | 'tableau',
  target: number,
): boolean {
  const sourcePile = state.tableau[source];
  if (!sourcePile || sourcePile.length <= 1) return false;
  const value = sourcePile[sourcePile.length - 1]?.value;
  if (value === undefined) return false;
  if (zone === 'tableau') return target !== source && state.tableau[target]?.length === 1;
  if (target < 0 || target >= state.openPiles) return false;
  const foundation = state.foundation[target];
  if (!foundation) return false;
  if (foundation.length === 0) return value === 1;
  if (foundation.length >= 11) return false;
  const top = foundation.at(-1);
  return top !== undefined && value === top.value + 1;
}

/** List legal foundation and bare-king destinations for a selected column. */
export function listSalicLawTargets(
  state: SalicLawResponse,
  source: number,
): { zone: 'foundation' | 'tableau'; idx: number }[] {
  const targets: { zone: 'foundation' | 'tableau'; idx: number }[] = [];
  for (let i = 0; i < state.foundation.length; i++) {
    if (isSalicLawTarget(state, source, 'foundation', i)) targets.push({ zone: 'foundation', idx: i });
  }
  for (let i = 0; i < state.tableau.length; i++) {
    if (isSalicLawTarget(state, source, 'tableau', i)) targets.push({ zone: 'tableau', idx: i });
  }
  return targets;
}
