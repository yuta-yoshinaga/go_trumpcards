import { describe, expect, it } from 'vitest';
import { makeThreeCardBragState } from '../../../test/stateFactories';
import { formatThreeCardBragState } from './threeCardBragFormatter';

describe('formatThreeCardBragState', () => {
  it('includes the header, deal, pot and stake', () => {
    const out = formatThreeCardBragState(makeThreeCardBragState());
    expect(out).toContain('Three Card Brag');
    expect(out).toContain('deal: 1');
    expect(out).toContain('pot: 4');
    expect(out).toContain('stake: 1');
    expect(out).toContain('phase: Betting');
  });

  it('renders each player with chips, bet and status', () => {
    const out = formatThreeCardBragState(makeThreeCardBragState());
    expect(out).toContain('chips=100');
    expect(out).toContain('bet=1');
    expect(out).toContain('[blind]');
  });

  it('marks a seen player and a folded player distinctly', () => {
    const state = makeThreeCardBragState({
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 80,
          seen: true,
          folded: false,
          out: false,
          roundBet: 2,
          cardCount: 3,
          cards: [],
        },
        {
          id: 1,
          isHuman: false,
          chips: 90,
          seen: false,
          folded: true,
          out: false,
          roundBet: 1,
          cardCount: 3,
          cards: [],
        },
        {
          id: 2,
          isHuman: false,
          chips: 0,
          seen: false,
          folded: false,
          out: true,
          roundBet: 0,
          cardCount: 0,
          cards: [],
        },
        {
          id: 3,
          isHuman: false,
          chips: 100,
          seen: false,
          folded: false,
          out: false,
          roundBet: 1,
          cardCount: 3,
          cards: [],
        },
      ],
    });
    const out = formatThreeCardBragState(state);
    expect(out).toContain('[seen]');
    expect(out).toContain('[folded]');
    expect(out).toContain('[OUT]');
  });

  it('shows the human hand cards when present', () => {
    const out = formatThreeCardBragState(makeThreeCardBragState());
    // The human's three spades are indexed cards in the default factory.
    expect(out).toMatch(/\[0\]/);
  });

  it('notes when a Show is available', () => {
    const out = formatThreeCardBragState(makeThreeCardBragState({ canShow: true }));
    expect(out.toLowerCase()).toContain('show');
  });

  it('renders the backend hint line', () => {
    const out = formatThreeCardBragState(makeThreeCardBragState({ hint: { action: 'fold', reason: 'fold' } }));
    expect(out).toContain('HINT: fold');
  });

  it('renders the game-over line with the match winner', () => {
    const out = formatThreeCardBragState(makeThreeCardBragState({ gameEndFlag: true, matchWinnerIdx: 0, phase: 3 }));
    expect(out).toContain('Game Over!');
  });
});
