/** Formats a non-negative number of thirds as points for a score announcement. */
export function formatThirdPoints(thirds: number): string {
  const points = Math.floor(thirds / 3);
  const remainder = thirds % 3;
  if (points > 0 && remainder > 0) return `${points}+${remainder}/3`;
  if (points > 0) return String(points);
  return `${remainder}/3`;
}
