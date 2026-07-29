import { describe, expect, it } from 'vitest';
import type { WindmillResponse } from '../../../types/card';
import { formatWindmillState } from './windmillFormatter';

function makeState(overrides?: Partial<WindmillResponse>): WindmillResponse {
  return {
    sails: Array.from({ length: 8 }, () => null),
    center: [],
    corners: Array.from({ length: 4 }, () => []),
    stockCount: 95,
    waste: [],
    transferBlocked: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatWindmillState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatWindmillState(makeState());
    expect(result).toContain('Windmill');
    expect(result).toContain('center: [  ] (0/52)');
    expect(result).toContain('k0:[  ]');
    expect(result).toContain('stock: 95');
    expect(result).toContain('s0: [empty]');
    expect(result).toContain('s7: [empty]');
  });

  it('renders the centre and corner tops with their depth', () => {
    const result = formatWindmillState(
      makeState({
        center: [{ design: 'SPADE', value: 1 }],
        corners: [[{ design: 'HEART', value: 13 }], [], [], []],
      }),
    );
    expect(result).toContain('(1/52)');
    expect(result).toContain('(1/13)');
  });

  it('renders the waste top', () => {
    expect(formatWindmillState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('renders a filled sail', () => {
    const sails: WindmillResponse['sails'] = Array.from({ length: 8 }, () => null);
    sails[0] = { design: 'SPADE', value: 9 };
    expect(formatWindmillState(makeState({ sails }))).not.toContain('s0: [empty]');
  });

  // The block is invisible in the layout, so it belongs on the status line.
  it('states the transfer block', () => {
    expect(formatWindmillState(makeState({ transferBlocked: true }))).toContain('must come from a sail');
  });

  it('shows a corner hint with its index', () => {
    const result = formatWindmillState(
      makeState({ hint: { fromZone: 'sail', fromIdx: 3, toZone: 'corner', toIdx: 1 } }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('sail3');
    expect(result).toContain('corner1');
  });

  // A draw has no source or destination index.
  it('renders a draw hint without indices', () => {
    const result = formatWindmillState(
      makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatWindmillState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatWindmillState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatWindmillState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
