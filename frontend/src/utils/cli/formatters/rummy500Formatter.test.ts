import { describe, expect, it } from 'vitest';
import { makeRummy500State } from '../../../test/stateFactories';
import { formatRummy500State } from './rummy500Formatter';

describe('formatRummy500State', () => {
  it('shows named phase, indexed selectable hand and available actions', () => {
    const state = makeRummy500State({
      phase: 1,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 8 },
          ],
          roundScore: 3,
          cumulativeScore: 12,
          laidMelds: [],
        },
        { id: 1, isHuman: false, cardCount: 4, cards: [], roundScore: 0, cumulativeScore: 2, laidMelds: [] },
      ],
    });
    const text = formatRummy500State(state);
    expect(text).toContain('phase: Play');
    expect(text).toContain('[0]♠7');
    expect(text).toContain('[1]♥8');
    expect(text).toContain('discard <idx>');
  });

  it('shows discard choices and round end action', () => {
    const text = formatRummy500State(makeRummy500State({ phase: 2, discardPile: [{ design: 'DIAMOND', value: 5 }] }));
    expect(text).toContain('phase: RoundEnd');
    expect(text).toContain('[0]♦5');
    expect(text).toContain('Use nr');
  });

  it('handles draw, unknown phases, optional winner and messages', () => {
    const draw = formatRummy500State(
      makeRummy500State({ phase: 0, message: 'Your turn', gameEndFlag: true, winnerIdx: 9 }),
    );
    expect(draw).toContain('phase: Draw');
    expect(draw).toContain('Use d to draw');
    expect(draw).toContain('Your turn');
    expect(draw).toContain('wins!');
    const unknown = formatRummy500State(makeRummy500State({ phase: 99 }));
    expect(unknown).toContain('phase: 99');
    expect(unknown).toContain('Game complete.');
  });

  it('prints a hint only for a requested hint message code', () => {
    const state = makeRummy500State({ message: 'Take the top discard card.', messageCode: 'rummy500.hintRequested' });
    expect(formatRummy500State(state)).toContain('HINT: Take the top discard card.');
    expect(formatRummy500State(makeRummy500State({ message: 'Passive message' }))).not.toContain('HINT');
  });
});
