import { describe, expect, it } from 'vitest';
import { makeKingState } from '../../../test/stateFactories';
import { formatKingState } from './kingFormatter';

describe('formatKingState', () => {
  it('renders the header, deal/trick, contract, trump and per-player scores', () => {
    const out = formatKingState(
      makeKingState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 13,
            cards: [{ design: 'HEART', value: 12 }],
            trickCount: 1,
            totalScore: -20,
          },
          { id: 1, isHuman: false, cardCount: 13, cards: [], trickCount: 0, totalScore: 0 },
          { id: 2, isHuman: false, cardCount: 13, cards: [], trickCount: 0, totalScore: 0 },
          { id: 3, isHuman: false, cardCount: 13, cards: [], trickCount: 0, totalScore: 0 },
        ],
      }),
    );
    expect(out).toContain('King');
    expect(out).toContain('deal: 1/7');
    expect(out).toContain('trick: 1');
    expect(out).toContain('contract: No Tricks');
    expect(out).toContain('score=-20');
  });

  it('renders a dash for an unset trump and an unset contract', () => {
    const out = formatKingState(makeKingState({ phase: 'selectContract', currentContract: -1, trumpSuit: -1 }));
    expect(out).toContain('contract: -');
    expect(out).toContain('trump: -');
  });

  it('renders the King (Trump) contract with its trump symbol', () => {
    const out = formatKingState(makeKingState({ currentContract: 6, trumpSuit: 3 }));
    expect(out).toContain('contract: King (Trump)');
    expect(out).toContain('trump: ♥');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatKingState(makeKingState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatKingState(
      makeKingState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders a hint with card indices', () => {
    const out = formatKingState(
      makeKingState({ hint: { cardIndices: [1, 2], reason: 'avoid_low' }, messageCode: 'king.hintRequested' }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('avoid_low');
  });

  it('renders the game-over banner with the winning player(s)', () => {
    const out = formatKingState(makeKingState({ phase: 'gameEnd', gameEndFlag: true, roundWinners: [1] }));
    expect(out).toContain('Game Over! Winner(s): Player 1');
  });

  it('renders an explicit message when present', () => {
    const out = formatKingState(makeKingState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'avoid_low' };
    expect(formatKingState(makeKingState({ hint, messageCode: 'king.hintRequested' }))).toContain('HINT');
    expect(formatKingState(makeKingState({ hint, messageCode: 'king.playing' }))).not.toContain('HINT');
  });
});
