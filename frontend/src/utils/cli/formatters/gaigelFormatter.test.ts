import { describe, expect, it } from 'vitest';
import { makeGaigelState } from '../../../test/stateFactories';
import { formatGaigelState } from './gaigelFormatter';

describe('formatGaigelState', () => {
  it('prints indexed human cards and game state', () => {
    const text = formatGaigelState(
      makeGaigelState({
        players: [
          { id: 0, isHuman: true, cardCount: 1, cards: [{ design: 'SPADE', value: 13 }], team: 0, trickCount: 0 },
        ],
      }),
    );
    expect(text).toContain('Gaigel');
    expect(text).toContain('round: 1');
    expect(text).toContain('[0]');
    expect(text).toContain('use p <idx> to play');
  });

  it('prints a completed trick, endgame and winner', () => {
    const text = formatGaigelState(
      makeGaigelState({
        phase: 3,
        isEndgame: true,
        gameEndFlag: true,
        winnerTeam: 1,
        currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 11 } }],
      }),
    );
    expect(text).toContain('phase: GameEnd');
    expect(text).toContain('(endgame)');
    expect(text).toContain('trick:');
    expect(text).toContain('Team 1 wins');
  });

  it('only prints requested hints, including marriage hints', () => {
    const hint = { cardIndex: 1, reason: 'marriage', isMarriage: true };
    expect(formatGaigelState(makeGaigelState({ hint })).toLowerCase()).not.toContain('hint:');
    expect(formatGaigelState(makeGaigelState({ hint, messageCode: 'gaigel.hintRequested' }))).toContain(
      'HINT: declare a marriage with card [1]',
    );
    expect(
      formatGaigelState(
        makeGaigelState({
          hint: { cardIndex: 0, reason: 'lead_low', isMarriage: false },
          messageCode: 'gaigel.hintRequested',
        }),
      ),
    ).toContain('HINT: play card [0]');
  });

  it('handles absent optional values and a draw', () => {
    const text = formatGaigelState(
      makeGaigelState({
        phase: 99,
        trumpSuit: 0,
        trumpCard: undefined,
        currentTrick: [],
        message: 'No move',
        gameEndFlag: true,
        winnerTeam: -1,
      }),
    );
    expect(text).toContain('phase: 99');
    expect(text).toContain('trump: none');
    expect(text).toContain('No move');
    expect(text).toContain('Game Over! Draw.');
  });
});
