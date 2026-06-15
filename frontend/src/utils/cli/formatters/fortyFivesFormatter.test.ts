import { describe, expect, it } from 'vitest';
import { makeFortyFivesState } from '../../../test/stateFactories';
import { formatFortyFivesState } from './fortyFivesFormatter';

describe('formatFortyFivesState', () => {
  it('renders the header, round/trick, trump and team scores', () => {
    const out = formatFortyFivesState(makeFortyFivesState({ trumpSuit: 3, teamScores: [20, 10] }));
    expect(out).toContain('Auction Forty-Fives');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♥');
    expect(out).toContain('A=20');
    expect(out).toContain('B=10');
  });

  it('renders the bids line while the declarer is undecided', () => {
    const out = formatFortyFivesState(makeFortyFivesState({ declarerIdx: -1, bids: [15, 0, 20, 0] }));
    expect(out).toContain('bids:');
    expect(out).toContain('P0=15');
    expect(out).toContain('P2=20');
    expect(out).toContain('P1=Pass');
  });

  it('renders the declarer, team, and contract once decided', () => {
    const out = formatFortyFivesState(
      makeFortyFivesState({
        phase: 1,
        declarerIdx: 0,
        contract: 20,
        players: [
          { id: 0, isHuman: true, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: true },
          { id: 1, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
          { id: 2, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
          { id: 3, isHuman: false, cardCount: 5, cards: [], trickCount: 0, teamScore: 0, isDeclarer: false },
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
    const out = formatFortyFivesState(makeFortyFivesState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatFortyFivesState(
      makeFortyFivesState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders round team points during RoundEnd', () => {
    const out = formatFortyFivesState(makeFortyFivesState({ phase: 3, roundTeamPoints: [15, 10] }));
    expect(out).toContain('round result: team A=15  team B=10');
  });

  it('renders a hint with card indices', () => {
    const out = formatFortyFivesState(makeFortyFivesState({ hint: { cardIndices: [1, 2], reason: 'take_trick' } }));
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('take_trick');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatFortyFivesState(makeFortyFivesState({ phase: 4, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatFortyFivesState(makeFortyFivesState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
