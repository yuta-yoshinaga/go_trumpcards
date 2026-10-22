import type { DuchessMoveZone } from '../types/card';

/** Builds the server legal-target key for a selected Duchess move source. */
export function duchessSourceKey(source: DuchessMoveZone | null): string | undefined {
  if (!source) return undefined;
  if (source.zone === 'tableau') {
    if (source.col === undefined || source.cardIndex === undefined) return undefined;
    return `tableau-${source.col}-${source.cardIndex}`;
  }
  if (source.zone === 'reserve') {
    if (source.col === undefined) return undefined;
    return `reserve-${source.col}`;
  }
  if (source.zone === 'waste') return 'waste';
  return undefined;
}
