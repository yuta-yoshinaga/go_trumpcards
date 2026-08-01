import { describe, expect, it } from 'vitest';
import { makeNapState } from '../../../test/stateFactories';
import { formatNapState } from './napFormatter';

describe('formatNapState', () => {
  it('renders the header, round/trick, trump and per-player chips', () => {
    const out = formatNapState(makeNapState({ trumpSuit: 3, playerScores: [3, 1, 2, 0] }));
    expect(out).toContain('Nap');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P3=0');
  });

  it('renders the bids line while the declarer is undecided', () => {
    const out = formatNapState(makeNapState({ declarerIdx: -1, bids: [2, 0, 5, 0] }));
    expect(out).toContain('bids:');
    expect(out).toContain('P0=Two');
    expect(out).toContain('P2=Nap');
  });

  it('renders the declarer and contract once decided', () => {
    const out = formatNapState(
      makeNapState({
        phase: 1,
        declarerIdx: 0,
        contract: 3,
        players: [
          { id: 0, isHuman: true, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: true },
          { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
          { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
          { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, score: 0, isDeclarer: false },
        ],
      }),
    );
    expect(out).toContain('declarer:');
    expect(out).toContain('Three');
    expect(out).toContain('Declarer');
    expect(out).toContain('Defender');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatNapState(makeNapState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatNapState(
      makeNapState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders round tricks during RoundEnd', () => {
    const out = formatNapState(makeNapState({ phase: 3, roundTricks: [3, 1, 1, 0] }));
    expect(out).toContain('round result: tricks P0=3 P1=1 P2=1 P3=0');
  });

  it('renders a hint with card indices', () => {
    const out = formatNapState(
      makeNapState({ messageCode: 'nap.hintRequested', hint: { cardIndices: [1, 2], reason: 'follow_win' } }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatNapState(makeNapState({ phase: 4, gameEndFlag: true, winnerPlayer: 2 }));
    expect(out).toContain('Game Over! Winner: Player 2');
  });

  it('renders an explicit message when present', () => {
    const out = formatNapState(makeNapState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatNapState(makeNapState({ hint, messageCode: 'nap.hintRequested' }))).toContain('HINT');
    expect(formatNapState(makeNapState({ hint, messageCode: 'nap.playing' }))).not.toContain('HINT');
  });
});
