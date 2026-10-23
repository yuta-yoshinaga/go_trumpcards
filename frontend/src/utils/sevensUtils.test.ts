import { describe, expect, it } from 'vitest';
import { actionDesc } from './sevensUtils';

describe('actionDesc joker reclaim', () => {
  const players = [
    { id: 0, isHuman: true },
    { id: 1, isHuman: false },
  ];
  // Echo the key and its params so the assertion reads what was actually asked
  // for, rather than a translated string that could resolve to anything.
  const t = (key: string, opts?: Record<string, unknown>) => `${key}${opts ? `(${Object.keys(opts).join(',')})` : ''}`;

  it('mentions the reclaim, which is otherwise a silent extra card', () => {
    const desc = actionDesc(
      players,
      {
        playerIdx: 0,
        playedCard: { design: 'SPADE', value: 6 },
        targetSuit: 0,
        targetValue: 0,
        forcedPass: false,
        jokerReclaimed: true,
      },
      t,
    );
    expect(desc).toContain('actionPlayed');
    expect(desc).toContain('actionReclaimedJoker');
  });

  it('says nothing about a reclaim that did not happen', () => {
    const desc = actionDesc(
      players,
      {
        playerIdx: 0,
        playedCard: { design: 'SPADE', value: 6 },
        targetSuit: 0,
        targetValue: 0,
        forcedPass: false,
        jokerReclaimed: false,
      },
      t,
    );
    expect(desc).not.toContain('actionReclaimedJoker');
  });
});
