/**
 * Formats a numeric score delta with an explicit sign.
 *
 * Positive numbers are prefixed with "+", zero reads as "no change" ("±0"),
 * and negative numbers retain their leading minus sign.
 *
 * @param n - The numeric delta to format.
 * @returns The formatted delta string (e.g. "+5", "±0", "-3").
 */
export function formatSignedDelta(n: number): string {
  if (n > 0) return `+${n.toString()}`;
  if (n === 0) return '±0';
  return n.toString();
}
