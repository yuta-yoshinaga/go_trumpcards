import { describe, expect, it } from 'vitest';
import { makeBouillotteState } from '../../../test/stateFactories';
import { formatBouillotteState } from './bouillotteFormatter';

describe('formatBouillotteState', () => {
  it('includes the header, round, pot, ante and current bet', () => {
    const out = formatBouillotteState(makeBouillotteState());
    expect(out).toContain('Bouillotte');
    expect(out).toContain('round: 1');
    expect(out).toContain('pot: 40');
    expect(out).toContain('ante: 10');
    expect(out).toContain('bet: 10');
    expect(out).toContain('phase: Betting');
  });

  it('renders the shared retourne card', () => {
    const out = formatBouillotteState(makeBouillotteState());
    expect(out).toContain('retourne:');
  });

  it('renders each player with chips, bet and status', () => {
    const out = formatBouillotteState(makeBouillotteState());
    expect(out).toContain('chips=190');
    expect(out).toContain('bet=10');
    expect(out).toContain('[active]');
  });

  it('marks winner, folded and out players distinctly', () => {
    const state = makeBouillotteState({
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 260,
          roundBet: 40,
          folded: false,
          out: false,
          cardCount: 3,
          cards: [],
          isWinner: true,
        },
        {
          id: 1,
          isHuman: false,
          chips: 150,
          roundBet: 10,
          folded: true,
          out: false,
          cardCount: 3,
          cards: [],
          isWinner: false,
        },
        {
          id: 2,
          isHuman: false,
          chips: 0,
          roundBet: 0,
          folded: false,
          out: true,
          cardCount: 0,
          cards: [],
          isWinner: false,
        },
      ],
    });
    const out = formatBouillotteState(state);
    expect(out).toContain('[WINNER]');
    expect(out).toContain('[folded]');
    expect(out).toContain('[OUT]');
  });

  it('shows the human hand cards when present', () => {
    const out = formatBouillotteState(makeBouillotteState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the backend hint line', () => {
    const out = formatBouillotteState(
      makeBouillotteState({ messageCode: 'bouillotte.hintRequested', hint: { action: 'fold', reason: 'weak_hand' } }),
    );
    expect(out).toContain('HINT: fold (weak_hand)');
  });

  it('renders the game-over line with the match winner', () => {
    const out = formatBouillotteState(makeBouillotteState({ gameEndFlag: true, matchWinnerIdx: 0, phase: 1 }));
    expect(out).toContain('Game Over!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { action: 'fold', reason: 'weak_hand' };
    expect(formatBouillotteState(makeBouillotteState({ hint, messageCode: 'bouillotte.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatBouillotteState(makeBouillotteState({ hint, messageCode: 'bouillotte.playing' }))).not.toContain(
      'HINT',
    );
  });
});
