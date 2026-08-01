import { describe, expect, it } from 'vitest';
import type { OsmosisResponse } from '../../../types/card';
import { formatOsmosisState } from './osmosisFormatter';

function makeState(overrides?: Partial<OsmosisResponse>): OsmosisResponse {
  return {
    reserve: [[], [], [], []],
    stockCount: 0,
    waste: [],
    foundation: [[], [], [], []],
    baseRank: 1,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('formatOsmosisState', () => {
  it('renders header and empty board', () => {
    const result = formatOsmosisState(makeState());
    expect(result).toContain('Osmosis');
    expect(result).toContain('stock: 0');
    expect(result).toContain('reserve0: [empty]');
    expect(result).toContain('foundation0: [  ]');
    expect(result).toContain('base: 1');
  });

  it('renders waste top card', () => {
    const result = formatOsmosisState(makeState({ waste: [{ design: 'SPADE', value: 5 }] }));
    expect(result).toContain('5');
  });

  it('renders reserve and foundation top cards with counts', () => {
    const result = formatOsmosisState(
      makeState({
        reserve: [[{ design: 'SPADE', value: 9 }], [], [], []],
        foundation: [[{ design: 'HEART', value: 1 }], [], [], []],
      }),
    );
    expect(result).toContain('reserve0:');
    expect(result).toContain('foundation0:');
    expect(result).toContain('(1)');
  });

  it('shows reserve hint when present', () => {
    const result = formatOsmosisState(
      makeState({ hint: { fromZone: 'reserve', fromCol: 2, toCol: 1 }, messageCode: 'osmosis.hintAvailable' }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('reserve2');
    expect(result).toContain('foundation1');
  });

  it('shows waste hint when present', () => {
    const result = formatOsmosisState(
      makeState({ hint: { fromZone: 'waste', fromCol: -1, toCol: 0 }, messageCode: 'osmosis.hintAvailable' }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('waste');
  });

  it('shows congrats on win phase', () => {
    expect(formatOsmosisState(makeState({ phase: 1 }))).toContain('Congratulations');
  });

  it('renders message when present', () => {
    expect(formatOsmosisState(makeState({ message: 'No moves available' }))).toContain('No moves available');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'reserve', fromCol: 2, toCol: 1 };
    expect(formatOsmosisState(makeState({ hint, messageCode: 'osmosis.hintAvailable' }))).toContain('HINT');
    expect(formatOsmosisState(makeState({ hint, messageCode: 'osmosis.playing' }))).not.toContain('HINT');
  });
});
