import { describe, expect, it } from 'vitest';
import { makeEscobaState } from '../../../test/stateFactories';
import { formatEscobaState } from './escobaFormatter';

describe('formatEscobaState', () => {
  it('renders the header, phase, players, and scores', () => {
    const out = formatEscobaState(
      makeEscobaState({
        players: [
          {
            id: 0,
            isHuman: true,
            handCount: 2,
            cards: [
              { design: 'SPADE', value: 3 },
              { design: 'HEART', value: 5 },
            ],
            capturedCount: 4,
            capturedCards: [],
            escobaCount: 1,
            score: 5,
          },
          {
            id: 1,
            isHuman: false,
            handCount: 3,
            cards: [],
            capturedCount: 0,
            capturedCards: [],
            escobaCount: 0,
            score: 3,
          },
          {
            id: 2,
            isHuman: false,
            handCount: 3,
            cards: [],
            capturedCount: 0,
            capturedCards: [],
            escobaCount: 0,
            score: 0,
          },
          {
            id: 3,
            isHuman: false,
            handCount: 3,
            cards: [],
            capturedCount: 0,
            capturedCards: [],
            escobaCount: 0,
            score: 0,
          },
        ],
      }),
    );
    expect(out).toContain('Escoba');
    expect(out).toContain('hand=2 captured=4 escobas=1 score=5');
    expect(out).toContain('hand=3 captured=0 escobas=0 score=3');
  });

  it('shows the empty-table marker when there are no table cards', () => {
    const out = formatEscobaState(makeEscobaState({ tableCards: [] }));
    expect(out).toContain('Table: (empty)');
  });

  it('renders the table cards when present', () => {
    const out = formatEscobaState(makeEscobaState({ tableCards: [{ design: 'DIAMOND', value: 7 }] }));
    expect(out).toContain('Table:');
  });

  it('renders the stock remaining', () => {
    const out = formatEscobaState(makeEscobaState({ stockRemaining: 12 }));
    expect(out).toContain('Stock: 12');
  });

  it('appends a message when present', () => {
    const out = formatEscobaState(makeEscobaState({ message: 'your turn' }));
    expect(out).toContain('your turn');
  });
});
