import type { AmericanToadMoveZone } from '../types/games/americantoad';

/** Returns the server legal-target key for a source zone. */
export function americanToadLegalTargetKey(source: AmericanToadMoveZone | null | undefined): string | undefined {
  if (!source) return undefined;
  if (source.zone === 'reserve' || source.zone === 'waste') return source.zone;
  if (source.zone === 'tableau' && source.col !== undefined && source.cardIndex !== undefined) {
    return `tableau-${source.col}-${source.cardIndex}`;
  }
  return undefined;
}
