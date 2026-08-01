import { describe, expect, it } from 'vitest';
import { makeBeziqueState } from '../../../test/stateFactories';
import { formatBeziqueState } from './beziqueFormatter';

describe('formatBeziqueState', () => {
  it('renders the header, deal/trick and match scores', () => {
    const out = formatBeziqueState(makeBeziqueState({ matchScore: [120, 40] }));
    expect(out).toContain('Bezique');
    expect(out).toContain('deal: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('P0=120');
    expect(out).toContain('P1=40');
  });

  it('shows the trump symbol and trump card', () => {
    const out = formatBeziqueState(makeBeziqueState({ trumpSuit: 3 }));
    expect(out).toContain('trump: ♥');
  });

  it('shows the stock count', () => {
    const out = formatBeziqueState(makeBeziqueState({ stockRemaining: 12 }));
    expect(out).toContain('stock: 12');
  });

  it('flags the endgame once the stock empties', () => {
    const out = formatBeziqueState(makeBeziqueState({ stockRemaining: 0, isEndgame: true }));
    expect(out).toContain('endgame');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatBeziqueState(makeBeziqueState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatBeziqueState(
      makeBeziqueState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'SPADE', value: 12 } },
          { playerIdx: 1, card: { design: 'HEART', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('lists available melds during the Meld phase', () => {
    const out = formatBeziqueState(
      makeBeziqueState({
        phase: 1,
        availableMelds: [
          { type: 0, suit: 1, points: 20 },
          { type: 2, suit: -1, points: 100 },
        ],
      }),
    );
    expect(out).toContain('available melds:');
    expect(out).toContain('marriage ♠ (20)');
    expect(out).toContain('four aces (100)');
  });

  it('renders a play hint with the card index', () => {
    const out = formatBeziqueState(
      makeBeziqueState({ hint: { cardIndex: 2, reason: 'follow_cut' }, messageCode: 'bezique.hintRequested' }),
    );
    expect(out).toContain('HINT: play card index [2]');
    expect(out).toContain('follow_cut');
  });

  it('renders a meld-declare hint with the meld index', () => {
    const out = formatBeziqueState(
      makeBeziqueState({ hint: { meldIndex: 1, reason: 'meld_declare' }, messageCode: 'bezique.hintRequested' }),
    );
    expect(out).toContain('HINT: declare meld index [1]');
    expect(out).toContain('meld_declare');
  });

  it('renders a meld-skip hint', () => {
    const out = formatBeziqueState(
      makeBeziqueState({ hint: { meldIndex: -1, reason: 'meld_skip' }, messageCode: 'bezique.hintRequested' }),
    );
    expect(out).toContain('HINT: skip the meld');
    expect(out).toContain('meld_skip');
  });

  it('renders the game-over banner with the winner', () => {
    const out = formatBeziqueState(makeBeziqueState({ phase: 3, gameEndFlag: true, winnerIdx: 0 }));
    expect(out).toContain('Game Over! Winner:');
  });

  it('renders an explicit message when present', () => {
    const out = formatBeziqueState(makeBeziqueState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  // このゲーム群は hintAvailable がラベルとして埋まっているため hintRequested。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 2, reason: 'follow_cut' };
    expect(formatBeziqueState(makeBeziqueState({ hint, messageCode: 'bezique.hintRequested' }))).toContain('HINT');
    expect(formatBeziqueState(makeBeziqueState({ hint, messageCode: 'bezique.playing' }))).not.toContain('HINT');
  });
});
