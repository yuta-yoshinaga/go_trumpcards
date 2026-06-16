import { describe, expect, it } from 'vitest';
import { makeCourtPieceState } from '../../../test/stateFactories';
import { formatCourtPieceState } from './courtPieceFormatter';

describe('formatCourtPieceState', () => {
  it('renders the header, round/trick and game points', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ teamScores: [3, 1] }));
    expect(out).toContain('Court Piece / Rang');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('A=3');
    expect(out).toContain('B=1');
  });

  it('shows an undeclared trump before declaration', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ trumpSuit: 0 }));
    expect(out).toContain('trump: undeclared');
  });

  it('shows the trump symbol once declared', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ trumpSuit: 3 }));
    expect(out).toContain('trump: ♥');
  });

  it('renders the caller line', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ callerIdx: 0 }));
    expect(out).toContain('caller:');
    expect(out).toContain('Caller');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatCourtPieceState(makeCourtPieceState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatCourtPieceState(
      makeCourtPieceState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders round trick totals during RoundEnd', () => {
    const players = [
      { id: 0, isHuman: true, team: 0, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 4 },
      { id: 1, isHuman: false, team: 1, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 2 },
      { id: 2, isHuman: false, team: 0, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 3 },
      { id: 3, isHuman: false, team: 1, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 4 },
    ];
    const out = formatCourtPieceState(makeCourtPieceState({ phase: 3, players, lastRoundCourt: true }));
    expect(out).toContain('round result: team A=7 tricks  team B=6 tricks');
    expect(out).toContain('Court bonus scored!');
  });

  it('renders a play hint with the card index', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ hint: { cardIndex: 2, reason: 'trump_cut' } }));
    expect(out).toContain('HINT: play card index [2]');
    expect(out).toContain('trump_cut');
  });

  it('renders a trump hint with the suit symbol', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ hint: { trumpSuit: 1, reason: 'trump_longest' } }));
    expect(out).toContain('HINT: declare trump ♠');
    expect(out).toContain('trump_longest');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ phase: 4, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatCourtPieceState(makeCourtPieceState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
