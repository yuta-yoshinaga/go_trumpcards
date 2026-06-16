import { describe, expect, it } from 'vitest';
import { makeTwentyNineState } from '../../../test/stateFactories';
import { formatTwentyNineState } from './twentyNineFormatter';

describe('formatTwentyNineState', () => {
  it('renders the header, round/trick and game points', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ teamScores: [3, 1] }));
    expect(out).toContain('Twenty-Nine (29)');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('A=3');
    expect(out).toContain('B=1');
  });

  it('hides the trump until it is revealed', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ trumpSuit: 3, trumpRevealed: false }));
    expect(out).toContain('trump: hidden');
    expect(out).not.toContain('trump: ♥');
  });

  it('shows the trump symbol once revealed', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ trumpSuit: 3, trumpRevealed: true }));
    expect(out).toContain('trump: ♥');
  });

  it('renders the bids line while the declarer is undecided', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ declarerIdx: -1, bids: [16, 0, 20, 0] }));
    expect(out).toContain('bids:');
    expect(out).toContain('P0=16');
    expect(out).toContain('P2=20');
    expect(out).toContain('P1=Pass');
  });

  it('renders the declarer, team, and contract once decided', () => {
    const out = formatTwentyNineState(
      makeTwentyNineState({
        phase: 1,
        declarerIdx: 0,
        contract: 20,
        players: [
          { id: 0, isHuman: true, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: true },
          { id: 1, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
          { id: 2, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
          { id: 3, isHuman: false, cardCount: 8, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
        ],
      }),
    );
    expect(out).toContain('declarer:');
    expect(out).toContain('team 0');
    expect(out).toContain('20');
    expect(out).toContain('Declarer');
    expect(out).toContain('Defender');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatTwentyNineState(makeTwentyNineState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatTwentyNineState(
      makeTwentyNineState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders round team points during RoundEnd', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ phase: 3, roundTeamPoints: [18, 11] }));
    expect(out).toContain('round result: team A=18  team B=11');
  });

  it('renders a hint with card indices', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ hint: { cardIndices: [1, 2], reason: 'follow_win' } }));
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ phase: 4, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatTwentyNineState(makeTwentyNineState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
