import { describe, expect, it } from 'vitest';
import { formatGameRankings } from './formatGameRankings';

describe('formatGameRankings', () => {
  const entry = (seat: string, rank: string) => `${seat} [${rank}]`;
  const rankLabel = (rank: number, shithead: boolean) => `#${rank}${shithead ? ' loser' : ''}`;

  it('formats seat and rank through the supplied localized entry template', () => {
    expect(
      formatGameRankings(
        [
          { seat: 0, rank: 1 },
          { seat: 2, rank: 3 },
        ],
        (seat) => `P${seat}`,
        rankLabel,
        entry,
        ' / ',
      ),
    ).toBe('P0 [#1] / P2 [#3]');
  });

  it('marks the Shithead and returns empty text when no rankings exist', () => {
    expect(
      formatGameRankings([{ seat: 1, rank: 4, shithead: true }], (seat) => `P${seat}`, rankLabel, entry, ' / '),
    ).toBe('P1 [#4 loser]');
    expect(formatGameRankings(undefined, () => '', rankLabel, entry, ' / ')).toBe('');
  });
});
