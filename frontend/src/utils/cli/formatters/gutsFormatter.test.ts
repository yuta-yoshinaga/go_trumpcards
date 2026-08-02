import { describe, expect, it } from 'vitest';
import { makeGutsState } from '../../../test/stateFactories';
import { formatGutsState } from './gutsFormatter';

describe('formatGutsState', () => {
  it('includes the header, round, pot and ante', () => {
    const out = formatGutsState(makeGutsState());
    expect(out).toContain('Guts');
    expect(out).toContain('round: 1');
    expect(out).toContain('pot: 40');
    expect(out).toContain('ante: 10');
    expect(out).toContain('phase: Declare');
  });

  it('renders each player with chips, bet and status', () => {
    const out = formatGutsState(makeGutsState());
    expect(out).toContain('chips=200');
    expect(out).toContain('bet=10');
    expect(out).toContain('[waiting]');
  });

  it('marks in, winner, matched and out players distinctly', () => {
    const state = makeGutsState({
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 260,
          in: true,
          out: false,
          roundBet: 10,
          cardCount: 2,
          cards: [],
          isWinner: true,
          isMatcher: false,
        },
        {
          id: 1,
          isHuman: false,
          chips: 150,
          in: true,
          out: false,
          roundBet: 10,
          cardCount: 2,
          cards: [],
          isWinner: false,
          isMatcher: true,
        },
        {
          id: 2,
          isHuman: false,
          chips: 0,
          in: false,
          out: true,
          roundBet: 0,
          cardCount: 0,
          cards: [],
          isWinner: false,
          isMatcher: false,
        },
      ],
    });
    const out = formatGutsState(state);
    expect(out).toContain('[WINNER]');
    expect(out).toContain('[matched]');
    expect(out).toContain('[OUT]');
  });

  it('shows the human hand cards when present', () => {
    const out = formatGutsState(makeGutsState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the backend hint line', () => {
    const out = formatGutsState(
      makeGutsState({ hint: { declaration: 0, reason: 'weak_hand' }, messageCode: 'guts.hintRequested' }),
    );
    expect(out).toContain('HINT: out (weak_hand)');
  });

  it('renders the game-over line with the match winner', () => {
    const out = formatGutsState(makeGutsState({ gameEndFlag: true, matchWinnerIdx: 0, phase: 1 }));
    expect(out).toContain('Game Over!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { declaration: 0, reason: 'weak_hand' };
    expect(formatGutsState(makeGutsState({ hint, messageCode: 'guts.hintRequested' }))).toContain('HINT');
    expect(formatGutsState(makeGutsState({ hint, messageCode: 'guts.playing' }))).not.toContain('HINT');
  });
});
