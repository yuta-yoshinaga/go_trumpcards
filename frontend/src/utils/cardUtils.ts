/** Return display name for a card value (1→'A', 11→'J', 12→'Q', 13→'K', else string). */
export function valueName(v: number): string {
  if (v === 1) return 'A';
  if (v === 11) return 'J';
  if (v === 12) return 'Q';
  if (v === 13) return 'K';
  return String(v);
}
