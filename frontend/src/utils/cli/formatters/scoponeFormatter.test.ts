import { describe, expect, it } from 'vitest';
import { makeScoponeState } from '../../../test/stateFactories';
import { formatScoponeState } from './scoponeFormatter';

describe('formatScoponeState', () => {
  it('renders the header, phase, players, teams, and scores', () => {
    const out = formatScoponeState(
      makeScoponeState({
        players: [
          {
            id: 0,
            isHuman: true,
            team: 0,
            handCount: 2,
            cards: [
              { design: 'SPADE', value: 3 },
              { design: 'HEART', value: 5 },
            ],
            capturedCount: 4,
            scopaCount: 1,
          },
          { id: 1, isHuman: false, team: 1, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
          { id: 2, isHuman: false, team: 0, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
          { id: 3, isHuman: false, team: 1, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
        ],
        teamScores: [5, 3],
      }),
    );
    expect(out).toContain('Scopone');
    expect(out).toContain('team0: hand=2 captured=4 scopas=1');
    expect(out).toContain('team1: hand=3 captured=0 scopas=0');
    expect(out).toContain('Team0: 5');
    expect(out).toContain('Team1: 3');
  });

  it('shows the empty-table marker when there are no table cards', () => {
    const out = formatScoponeState(makeScoponeState({ tableCards: [] }));
    expect(out).toContain('Table: (empty)');
  });

  it('renders the table cards when present', () => {
    const out = formatScoponeState(makeScoponeState({ tableCards: [{ design: 'DIAMOND', value: 7 }] }));
    expect(out).toContain('Table:');
  });

  it('appends a message when present', () => {
    const out = formatScoponeState(makeScoponeState({ message: 'your turn' }));
    expect(out).toContain('your turn');
  });
});
