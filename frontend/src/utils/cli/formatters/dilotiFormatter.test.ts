import { describe, expect, it } from 'vitest';
import { makeDilotiState } from '../../../test/stateFactories';
import { formatDilotiState } from './dilotiFormatter';

describe('formatDilotiState', () => {
  it('shows the round, the stock and the seats', () => {
    const out = formatDilotiState(makeDilotiState());
    expect(out).toContain('Diloti');
    expect(out).toContain('round: 1');
    expect(out).toContain('target: 61');
    expect(out).toContain('stock: 36');
    expect(out).toContain('xeri=0');
  });

  // **場札にも宣言にも番号が要る。** 取る対象はこの番号で指す。
  it('numbers the table cards and the declarations', () => {
    const out = formatDilotiState(makeDilotiState());
    expect(out).toMatch(/table: \[0\].*\[1\].*\[2\].*\[3\]/);
    expect(out).toContain('decl[0] 5');
    expect(out).toContain('plain');
  });

  it('marks a group declaration as unraisable', () => {
    const out = formatDilotiState(
      makeDilotiState({
        declarations: [
          {
            ownerIdx: 0,
            value: 6,
            groups: [[{ design: 'SPADE', value: 6, color: 'black' }], [{ design: 'HEART', value: 6, color: 'red' }]],
            isGroup: true,
          },
        ],
      }),
    );
    expect(out).toContain('group, cannot be raised');
    expect(out).not.toContain('plain');
  });

  it('says when the table is empty', () => {
    expect(formatDilotiState(makeDilotiState({ table: [] }))).toContain('table: (empty)');
  });

  // **どの札で何ができるかを出す。** 出さないと端末から総当たりで探すことになる。
  it('lists the takes and the declarations each card can make', () => {
    const out = formatDilotiState(makeDilotiState());
    expect(out).toContain('[0] can take: (0) (2 3)');
    expect(out).toContain('[1] can take: (d0)');
    expect(out).toContain('[1] can declare: 8:(0)');
  });

  it('omits the option lines for a card with no moves', () => {
    const out = formatDilotiState(makeDilotiState({ takeOptions: [[], [], []], declareOptions: [[], [], []] }));
    expect(out).not.toContain('can take');
    expect(out).not.toContain('can declare');
  });

  it('shows the round score, skipping the empty lines', () => {
    const out = formatDilotiState(
      makeDilotiState({
        lastResult: {
          lines: [
            { key: 'cards', points: [4, 0] },
            { key: 'aces', points: [0, 0] },
            { key: 'xeri', points: [10, 0] },
          ],
          totals: [14, 0],
          cardCounts: [30, 22],
          xeris: [1, 0],
        },
      }),
    );
    expect(out).toContain('cards: 4 - 0');
    expect(out).toContain('xeri: 10 - 0');
    expect(out).not.toContain('aces:');
    expect(out).toContain('round total: 14 - 0');
  });

  // **ヒントは訊いたときだけ出す。** 常時出すと盤面が助言で埋まる。
  it('shows the hint only when it was requested', () => {
    const hinted = makeDilotiState({ hintHandIdx: 1, hintAction: 'capture', hintReason: 'capture' });
    expect(formatDilotiState(hinted)).not.toContain('HINT:');
    expect(formatDilotiState({ ...hinted, messageCode: 'diloti.hintRequested' })).toContain(
      'HINT: play [1] and capture (capture)',
    );
  });

  it('shows the message and the winner', () => {
    expect(formatDilotiState(makeDilotiState({ message: 'hello' }))).toContain('hello');
    const over = formatDilotiState(makeDilotiState({ gameEndFlag: true, winnerIdx: 0 }));
    expect(over).toContain('Game Over!');
  });
});
