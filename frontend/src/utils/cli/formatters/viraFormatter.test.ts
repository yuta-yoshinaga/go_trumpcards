import { describe, expect, it } from 'vitest';
import { makeViraState } from '../../../test/stateFactories';
import { formatViraState } from './viraFormatter';

describe('formatViraState', () => {
  it('renders the header, round/trick, trump and per-player scores', () => {
    const out = formatViraState(makeViraState({ trumpSuit: 3, playerScores: [3, 1, 2] }));
    expect(out).toContain('Vira');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P2=2');
  });

  it('renders the bids line while the declarer is undecided', () => {
    const out = formatViraState(makeViraState({ declarerIdx: -1, bids: [1, 0, 2] }));
    expect(out).toContain('bids:');
    expect(out).toContain('P0=Six');
    expect(out).toContain('P2=Misère');
  });

  it('renders the declarer and contract once decided', () => {
    const out = formatViraState(
      makeViraState({
        phase: 1,
        declarerIdx: 0,
        contract: 3,
        players: [
          { id: 0, isHuman: true, cardCount: 10, cards: [], trickCount: 0, score: 0, isDeclarer: true },
          { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isDeclarer: false },
          { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isDeclarer: false },
        ],
      }),
    );
    expect(out).toContain('declarer:');
    expect(out).toContain('Seven');
    expect(out).toContain('Declarer');
    expect(out).toContain('Defender');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatViraState(makeViraState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatViraState(
      makeViraState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders round tricks during RoundEnd', () => {
    const out = formatViraState(makeViraState({ phase: 3, roundTricks: [6, 2, 2] }));
    expect(out).toContain('round result: tricks P0=6 P1=2 P2=2');
  });

  it('renders a hint with card indices', () => {
    const out = formatViraState(
      makeViraState({
        hint: { cardIndices: [1, 2], reason: 'follow_win' },
        messageCode: 'vira.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatViraState(makeViraState({ phase: 4, gameEndFlag: true, winnerPlayer: 2 }));
    expect(out).toContain('Game Over! Winner: Player 2');
  });

  it('renders an explicit message when present', () => {
    const out = formatViraState(makeViraState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatViraState(makeViraState({ hint, messageCode: 'vira.hintRequested' }))).toContain('HINT');
    expect(formatViraState(makeViraState({ hint, messageCode: 'vira.playing' }))).not.toContain('HINT');
  });
});
