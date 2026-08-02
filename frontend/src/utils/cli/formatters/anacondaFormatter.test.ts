import { describe, expect, it } from 'vitest';
import { makeAnacondaState } from '../../../test/stateFactories';
import { formatAnacondaState } from './anacondaFormatter';

describe('formatAnacondaState', () => {
  it('includes the header, round, pot, ante and bet', () => {
    const out = formatAnacondaState(makeAnacondaState());
    expect(out).toContain('Anaconda');
    expect(out).toContain('round: 1');
    expect(out).toContain('pot: 40');
    expect(out).toContain('ante: 10');
    expect(out).toContain('bet: 0');
    expect(out).toContain('phase: Pass');
  });

  it('shows the remaining pass count during the pass phase', () => {
    const out = formatAnacondaState(makeAnacondaState({ phase: 0, passCount: 2 }));
    expect(out).toContain('pass: 2 card(s) left');
  });

  it('shows the reveal progress during the roll phase', () => {
    const out = formatAnacondaState(makeAnacondaState({ phase: 2, rollIndex: 3 }));
    expect(out).toContain('revealed: 3/5');
  });

  it('renders each player with chips, bet and status', () => {
    const out = formatAnacondaState(makeAnacondaState());
    expect(out).toContain('chips=190');
    expect(out).toContain('bet=10');
    expect(out).toContain('[to act]');
  });

  it('marks folded, winner and out players distinctly', () => {
    const state = makeAnacondaState({
      currentPlayer: 3,
      players: [
        {
          id: 0,
          isHuman: true,
          chips: 260,
          folded: false,
          out: false,
          roundBet: 10,
          streetBet: 0,
          cardCount: 5,
          cards: [],
          handName: 'flush',
          isWinner: true,
        },
        {
          id: 1,
          isHuman: false,
          chips: 150,
          folded: true,
          out: false,
          roundBet: 10,
          streetBet: 0,
          cardCount: 5,
          cards: [],
          isWinner: false,
        },
        {
          id: 2,
          isHuman: false,
          chips: 0,
          folded: false,
          out: true,
          roundBet: 0,
          streetBet: 0,
          cardCount: 0,
          cards: [],
          isWinner: false,
        },
        {
          id: 3,
          isHuman: false,
          chips: 150,
          folded: false,
          out: false,
          roundBet: 10,
          streetBet: 0,
          cardCount: 5,
          cards: [],
          isWinner: false,
        },
      ],
    });
    const out = formatAnacondaState(state);
    expect(out).toContain('[WINNER]');
    expect(out).toContain('[folded]');
    expect(out).toContain('[OUT]');
    expect(out).toContain('[to act]');
    expect(out).toContain('hand=flush');
  });

  it('shows the human hand cards when present', () => {
    const out = formatAnacondaState(makeAnacondaState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the backend hint line with card indices', () => {
    const out = formatAnacondaState(
      makeAnacondaState({
        messageCode: 'anaconda.hintRequested',
        hint: { action: 'pass', cardIndices: [4, 5, 6], reason: 'pass_weakest' },
      }),
    );
    expect(out).toContain('HINT: pass [4 5 6] (pass_weakest)');
  });

  it('renders a betting hint line without card indices', () => {
    const out = formatAnacondaState(
      makeAnacondaState({ hint: { action: 'raise', reason: 'strong_hand' }, messageCode: 'anaconda.hintRequested' }),
    );
    expect(out).toContain('HINT: raise (strong_hand)');
  });

  it('renders the game-over line with the match winner', () => {
    const out = formatAnacondaState(makeAnacondaState({ gameEndFlag: true, matchWinnerIdx: 0, phase: 3 }));
    expect(out).toContain('Game Over!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { action: 'pass', cardIndices: [4, 5, 6], reason: 'pass_weakest' };
    expect(formatAnacondaState(makeAnacondaState({ hint, messageCode: 'anaconda.hintRequested' }))).toContain('HINT');
    expect(formatAnacondaState(makeAnacondaState({ hint, messageCode: 'anaconda.playing' }))).not.toContain('HINT');
  });
});
