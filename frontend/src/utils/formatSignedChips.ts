/** Formats a chip amount with an explicit sign, using ± for zero. */
export function formatSignedChips(n: number): string {
  return n > 0 ? `+${n}` : n < 0 ? `${n}` : '±0';
}
