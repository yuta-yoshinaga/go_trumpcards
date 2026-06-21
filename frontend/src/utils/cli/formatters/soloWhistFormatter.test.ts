import { describe, expect, it } from 'vitest';
import { makeSoloWhistState } from '../../../test/stateFactories';
import { formatSoloWhistState } from './soloWhistFormatter';

describe('formatSoloWhistState', () => {
  it('renders the header, round/trick, trump and per-player scores', () => {
    const out = formatSoloWhistState(makeSoloWhistState({ trumpSuit: 3, playerScores: [3, 1, 2, 0] }));
    expect(out).toContain('Solo Whist');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P3=0');
  });

  it('renders the bids line while the declarer is undecided', () => {
    const out = formatSoloWhistState(makeSoloWhistState({ declarerIdx: -1, bids: [1, 0, 2, 0] }));
    expect(out).toContain('bids:');
    expect(out).toContain('P0=Solo');
    expect(out).toContain('P2=Misère');
  });

  it('renders the declarer and contract once decided', () => {
    const out = formatSoloWhistState(
      makeSoloWhistState({
        phase: 1,
        declarerIdx: 0,
        contract: 1,
        players: [
          { id: 0, isHuman: true, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: true },
          { id: 1, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
          { id: 2, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
          { id: 3, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
        ],
      }),
    );
    expect(out).toContain('declarer:');
    expect(out).toContain('Solo');
    expect(out).toContain('Declarer');
    expect(out).toContain('Defender');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatSoloWhistState(makeSoloWhistState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatSoloWhistState(
      makeSoloWhistState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders round tricks during RoundEnd', () => {
    const out = formatSoloWhistState(makeSoloWhistState({ phase: 3, roundTricks: [8, 2, 2, 1] }));
    expect(out).toContain('round result: tricks P0=8 P1=2 P2=2 P3=1');
  });

  it('renders a hint with card indices', () => {
    const out = formatSoloWhistState(makeSoloWhistState({ hint: { cardIndices: [1, 2], reason: 'follow_win' } }));
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatSoloWhistState(makeSoloWhistState({ phase: 4, gameEndFlag: true, winnerPlayer: 2 }));
    expect(out).toContain('Game Over! Winner: Player 2');
  });

  it('renders an explicit message when present', () => {
    const out = formatSoloWhistState(makeSoloWhistState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
