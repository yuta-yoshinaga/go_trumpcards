import { describe, expect, it } from 'vitest';
import { makeCirullaState } from '../../../test/stateFactories';
import { formatCirullaState } from './cirullaFormatter';

describe('formatCirullaState', () => {
  it('prints the round, phase, target and stock', () => {
    const out = formatCirullaState(makeCirullaState());
    expect(out).toContain('round: 1  phase: play  target: 51');
    expect(out).toContain('stock: 30');
  });

  it('lists each seat with its takings', () => {
    const out = formatCirullaState(makeCirullaState());
    expect(out).toContain('taken=0 denari=0 scopa=0 score=0');
    expect(out).toContain('(Dealer)');
  });

  it('indexes the human hand and the table', () => {
    const out = formatCirullaState(makeCirullaState());
    expect(out).toContain('[0]');
    expect(out).toContain('table: ');
  });

  it('says when the table is empty', () => {
    expect(formatCirullaState(makeCirullaState({ table: [] }))).toContain('table: (empty)');
  });

  // **どの札で何が取れるかを出す。** 3 つの規則が混ざるので、出さないと
  // 端末から総当たりで探すことになる。
  it('lists the capture groups per hand card', () => {
    const out = formatCirullaState(makeCirullaState());
    expect(out).toContain('[0] can take: (2)');
    expect(out).toContain('[2] can take: (0 1 2 3)');
    // 取れない札の行は出さない。
    expect(out).not.toContain('[1] can take');
  });

  it('surfaces a deal bonus', () => {
    const base = makeCirullaState();
    const out = formatCirullaState(
      makeCirullaState({ players: base.players.map((p, i) => (i === 0 ? { ...p, lastBonus: 'barsegon' } : p)) }),
    );
    expect(out).toContain('bonus: barsegon');
  });

  it('breaks the round scoring down, skipping empty lines', () => {
    const out = formatCirullaState(
      makeCirullaState({
        lastResult: {
          lines: [
            { key: 'cards', points: [1, 0] },
            { key: 'denari', points: [0, 0] },
          ],
          totals: [1, 0],
          sweptDenari: -1,
        },
      }),
    );
    expect(out).toContain('cards: 1 - 0');
    expect(out).not.toContain('denari: 0 - 0');
    expect(out).toContain('round total: 1 - 0');
    expect(out).not.toContain('all denari');
  });

  it('calls out a denari sweep', () => {
    const out = formatCirullaState(makeCirullaState({ lastResult: { lines: [], totals: [1, 0], sweptDenari: 0 } }));
    expect(out).toContain('all denari to');
  });

  // 頼んでいないヒントは出さない。
  it('gates the hint on the request', () => {
    const hinted = { hintHandIdx: 0, hintCaptureIdxs: [2], hintReason: 'capture' };
    expect(formatCirullaState(makeCirullaState({ ...hinted, messageCode: '' }))).not.toContain('HINT:');
    expect(formatCirullaState(makeCirullaState({ ...hinted, messageCode: 'cirulla.hintRequested' }))).toContain(
      'HINT: play [0] and take (2) (capture)',
    );
  });

  it('says lay off when the hint captures nothing', () => {
    expect(
      formatCirullaState(
        makeCirullaState({
          hintHandIdx: 1,
          hintCaptureIdxs: [],
          hintReason: 'lay_off',
          messageCode: 'cirulla.hintRequested',
        }),
      ),
    ).toContain('HINT: play [1] and lay off (lay_off)');
  });

  it('announces the winner at the end', () => {
    expect(formatCirullaState(makeCirullaState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('Game Over!');
  });

  it('prints the server message', () => {
    expect(formatCirullaState(makeCirullaState({ message: '札を出してください。' }))).toContain('札を出してください。');
  });
});
