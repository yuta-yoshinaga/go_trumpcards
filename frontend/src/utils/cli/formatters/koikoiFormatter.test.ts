import { describe, expect, it } from 'vitest';
import { makeKoiKoiState } from '../../../test/stateFactories';
import { formatKoiKoiState } from './koikoiFormatter';

describe('formatKoiKoiState', () => {
  it('formats the play phase with header, round line, field, and hands', () => {
    const out = formatKoiKoiState(makeKoiKoiState());
    expect(out).toContain('Koi-Koi');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('field:');
    expect(out).toContain('----------');
    // Human hand is indexed and rendered.
    expect(out).toContain('[0]');
  });

  it('renders (empty) when the field has no cards', () => {
    const out = formatKoiKoiState(makeKoiKoiState({ fieldCards: [] }));
    expect(out).toContain('field: (empty)');
  });

  it('falls back to the raw phase number for an unknown phase', () => {
    const out = formatKoiKoiState(makeKoiKoiState({ phase: 9 as unknown as 0 }));
    expect(out).toContain('phase: 9');
  });

  it('renders a captured yaku list for a player', () => {
    const state = makeKoiKoiState();
    state.players[0] = { ...state.players[0], yaku: [{ key: 'tane', points: 1 }] };
    const out = formatKoiKoiState(state);
    expect(out).toContain('yaku=[tane:1]');
  });

  it('renders the decision line during the KoiKoiDecision phase', () => {
    const out = formatKoiKoiState(
      makeKoiKoiState({ phase: 1, pendingYaku: [{ key: 'tane', points: 2 }], pendingPoints: 2 }),
    );
    expect(out).toContain('decision: [tane:2] = 2');
    expect(out).toContain('(koikoi / stop)');
  });

  it('renders the round result during the RoundEnd phase', () => {
    const out = formatKoiKoiState(
      makeKoiKoiState({
        phase: 2,
        lastRoundResult: {
          winner: 0,
          yaku: [{ key: 'tane', points: 1 }],
          basePoints: 1,
          multiplier: 2,
          total: 2,
          koikoiCount: 1,
        },
      }),
    );
    expect(out).toContain('result: winner=');
    expect(out).toContain('1×2=2');
  });

  it('renders a draw round result when the winner is negative', () => {
    const out = formatKoiKoiState(
      makeKoiKoiState({
        phase: 2,
        lastRoundResult: {
          winner: -1,
          yaku: [],
          basePoints: 0,
          multiplier: 1,
          total: 0,
          koikoiCount: 0,
        },
      }),
    );
    expect(out).toContain('winner=draw');
  });

  it('renders a hint line and a trailing message when present', () => {
    const out = formatKoiKoiState(
      makeKoiKoiState({
        hint: { cardIndex: 1, fieldIndex: 0, koikoi: 1, reason: 'koikoi_capture' },
        message: 'hello world',
      }),
    );
    expect(out).toContain('HINT: play 1 field 0 (koikoi_capture) / koikoi');
    expect(out).toContain('hello world');
  });

  it('renders the stop hint action when koikoi is not 1', () => {
    const out = formatKoiKoiState(
      makeKoiKoiState({ hint: { cardIndex: 0, fieldIndex: -1, koikoi: 0, reason: 'stop_secure' } }),
    );
    expect(out).toContain('/ stop');
  });
});
