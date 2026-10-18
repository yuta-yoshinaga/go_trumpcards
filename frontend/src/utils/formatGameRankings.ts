/** Formats structured seat ranks with caller supplied localized labels. */
export function formatGameRankings(
  rankings: { seat: number; rank: number; shithead?: boolean }[] | undefined,
  seatName: (seat: number) => string,
  rankLabel: (rank: number, shithead: boolean) => string,
  formatEntry: (seat: string, rank: string) => string,
  separator: string,
): string {
  return (rankings ?? [])
    .map(({ seat, rank, shithead }) => formatEntry(seatName(seat), rankLabel(rank, !!shithead)))
    .join(separator);
}
