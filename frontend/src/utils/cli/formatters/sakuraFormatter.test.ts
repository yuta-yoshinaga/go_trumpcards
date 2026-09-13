import { describe, expect, it } from 'vitest';
import { makeSakuraState } from '../../../test/stateFactories';
import { formatSakuraState } from './sakuraFormatter';

describe('formatSakuraState', () => {
  it('formats the field, players, and human hand', () => {
    const out = formatSakuraState(makeSakuraState());
    expect(out).toContain('Sakura');
    expect(out).toContain('round: 1/3');
    expect(out).toContain('field:');
    expect(out).toContain('[0]');
  });
  it('formats empty field, results, hint, and message', () => {
    const out = formatSakuraState(
      makeSakuraState({
        fieldCards: [],
        phase: 1,
        lastResult: { round: 1, winner: -1, seats: [] },
        hint: { cardIndex: 0, fieldIndex: -1, reason: 'discard' },
        message: 'hello',
      }),
    );
    expect(out).toContain('field: (empty)');
    expect(out).toContain('winner=draw');
    expect(out).toContain('HINT: play 0 (discard)');
    expect(out).toContain('hello');
  });

  it('formats a round winner and a human with an empty hand', () => {
    const out = formatSakuraState(
      makeSakuraState({
        phase: 1,
        players: [{ ...makeSakuraState().players[0], cards: [] }, ...makeSakuraState().players.slice(1)],
        lastResult: { round: 1, winner: 0, seats: [] },
      }),
    );
    expect(out).toContain('winner=あなた');
    expect(out).not.toContain('[0]');
  });

  it('formats a game winner and a hint that includes a field index', () => {
    const out = formatSakuraState(
      makeSakuraState({ phase: 2, winner: 1, hint: { cardIndex: 2, fieldIndex: 0, reason: 'capture' } }),
    );
    expect(out).toContain('winner: CPU 1');
    expect(out).toContain('HINT: play 2 field 0 (capture)');
  });

  it('falls back to draw when the game winner is negative', () => {
    expect(formatSakuraState(makeSakuraState({ phase: 2, winner: -1 }))).toContain('winner: draw');
  });
});
