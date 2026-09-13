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
});
