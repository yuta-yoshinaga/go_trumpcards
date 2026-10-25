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

  it('marks marriage cards and names the trump suit', () => {
    const text = formatGaigelState(
      makeGaigelState({
        trumpSuit: 2,
        marriageIndices: [1],
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [
              { design: 'SPADE', value: 13 },
              { design: 'HEART', value: 12 },
            ],
            team: 0,
            trickCount: 0,
          },
        ],
      }),
    );
    expect(text).toContain('trump: ♣');
    expect(text).toContain('[1]*♥Q');
    expect(text).toContain('m <idx> to declare a marriage');
  });

  it('does not add marriage notation when there are no marriage indices', () => {
    const text = formatGaigelState(makeGaigelState({ marriageIndices: [] }));
    expect(text).not.toContain('*');
    expect(text).not.toContain('m <idx> to declare a marriage');
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

  it('does not print a hint for non-hint messages or gaigel.noHint', () => {
    const hint = { cardIndex: 1, reason: 'lead', isMarriage: false };
    expect(formatGaigelState(makeGaigelState({ hint, messageCode: 'gaigel.playPhase.lead' }))).not.toContain('HINT:');
    expect(formatGaigelState(makeGaigelState({ hint, messageCode: 'gaigel.noHint' }))).not.toContain('HINT:');
  });

  it('does not expose raw gaigel translation keys', () => {
    const text = formatGaigelState(
      makeGaigelState({ message: 'A translated message', messageCode: 'gaigel.hintRequested' }),
    );
    expect(text).not.toMatch(/gaigel\./);
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
