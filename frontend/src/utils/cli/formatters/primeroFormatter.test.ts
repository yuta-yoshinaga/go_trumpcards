import { describe, expect, it } from 'vitest';
import { makePrimeroState } from '../../../test/stateFactories';
import { formatPrimeroState } from './primeroFormatter';

describe('formatPrimeroState', () => {
  it('includes the header, round, pot, ante and current bet', () => {
    const out = formatPrimeroState(makePrimeroState());
    expect(out).toContain('Primero');
    expect(out).toContain('round: 1');
    expect(out).toContain('pot: 40');
    expect(out).toContain('ante: 10');
    expect(out).toContain('bet: 10');
    expect(out).toContain('phase: Betting');
  });

  it('renders each player with chips, bet and status', () => {
    const out = formatPrimeroState(makePrimeroState());
    expect(out).toContain('chips=190');
    expect(out).toContain('bet=10');
    expect(out).toContain('[active]');
  });

  it('marks winner, folded and out players distinctly', () => {
    const state = makePrimeroState({
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 260,
          roundBet: 40,
          folded: false,
          out: false,
          cardCount: 4,
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
          cardCount: 4,
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
    const out = formatPrimeroState(state);
    expect(out).toContain('[WINNER]');
    expect(out).toContain('[folded]');
    expect(out).toContain('[OUT]');
  });

  it('shows the human hand cards when present', () => {
    const out = formatPrimeroState(makePrimeroState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the backend hint line', () => {
    const out = formatPrimeroState(
      makePrimeroState({ hint: { action: 'fold', reason: 'weak_hand' }, messageCode: 'primero.hintRequested' }),
    );
    expect(out).toContain('HINT: fold (weak_hand)');
  });

  it('renders the game-over line with the match winner', () => {
    const out = formatPrimeroState(makePrimeroState({ gameEndFlag: true, matchWinnerIdx: 0, phase: 1 }));
    expect(out).toContain('Game Over!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { action: 'fold', reason: 'weak_hand' };
    expect(formatPrimeroState(makePrimeroState({ hint, messageCode: 'primero.hintRequested' }))).toContain('HINT');
    expect(formatPrimeroState(makePrimeroState({ hint, messageCode: 'primero.playing' }))).not.toContain('HINT');
  });
});
