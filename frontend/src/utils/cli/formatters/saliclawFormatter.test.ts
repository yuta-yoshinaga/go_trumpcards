import { describe, expect, it } from 'vitest';
import type { SalicLawResponse } from '../../../types/card';
import { formatSalicLawState } from './saliclawFormatter';

function makeState(overrides?: Partial<SalicLawResponse>): SalicLawResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 95,
    queens: [{ design: 'SPADE', value: 12 }],
    openPiles: 1,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatSalicLawState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatSalicLawState(makeState());
    expect(result).toContain('Salic Law');
    expect(result).toContain('foundations:');
    expect(result).toContain('stock: 95');
    expect(result).toContain('t0: [not open]');
    expect(result).toContain('t7: [not open]');
  });

  // **未開放の列を「空き」と書かない。**このゲームの唯一の置き場所は
  // 「K だけの列」で、まだ K が出ていない列とは別物。
  it('calls an unopened column not-open, never empty', () => {
    const result = formatSalicLawState(makeState());
    expect(result).toContain('a king opens it');
    expect(result).not.toContain('king only');
  });

  // K だけの列には目印を付ける。他の列と同じ見た目だと探せない。
  it('marks a column holding just its king', () => {
    const result = formatSalicLawState(
      makeState({
        tableau: [
          [{ design: 'SPADE', value: 13 }],
          [
            { design: 'HEART', value: 13 },
            { design: 'CLOVER', value: 4 },
          ],
          ...Array.from({ length: 6 }, () => []),
        ],
        openPiles: 2,
      }),
    );
    // 印はちょうど 1 つ。全列に付いたら見分けにならない。
    expect(result.match(/king only/g)).toHaveLength(1);
  });

  // 退場した Q を出す。8 枚消えている理由は盤からは読めない。
  it('lists the queens that are out of play', () => {
    expect(formatSalicLawState(makeState())).toContain('queens out of play:');
  });

  it('renders cards in a pile', () => {
    const result = formatSalicLawState(
      makeState({ tableau: [[{ design: 'SPADE', value: 9 }], ...Array.from({ length: 7 }, () => [])] }),
    );
    expect(result).toContain('[0]');
  });

  it('shows a tableau hint with its pile', () => {
    const result = formatSalicLawState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'saliclaw.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3');
    expect(result).toContain('foundation2');
  });

  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatSalicLawState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 2 },
        messageCode: 'saliclaw.playing',
      }),
    );
    expect(result).not.toContain('HINT:');
  });

  // **配れのヒントに列は出ない。**移動の体裁に落とすと t-1 が漏れる。
  it('renders a deal hint without any index', () => {
    const result = formatSalicLawState(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, toZone: 'stock', toIdx: -1 },
        messageCode: 'saliclaw.hintAvailable',
      }),
    );
    expect(result).toContain('deal another card');
    expect(result).not.toContain('-1');
  });

  it('shows stalemate message', () => {
    expect(formatSalicLawState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatSalicLawState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatSalicLawState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
