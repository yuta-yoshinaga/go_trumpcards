import { describe, expect, it } from 'vitest';
import { makeKoenigrufenState } from '../../../test/stateFactories';
import { formatKoenigrufenState } from './koenigrufenFormatter';

describe('formatKoenigrufenState', () => {
  it('renders the header, the contract and the highest bid', () => {
    const out = formatKoenigrufenState(makeKoenigrufenState());
    expect(out).toContain('Königrufen');
    expect(out).toContain('contract:');
    expect(out).toContain('highestBid:');
  });

  // **呼ばれた王は宣言されるまで無い。**0 のあいだは行ごと出ない。
  it('names the called King only once one is called', () => {
    expect(formatKoenigrufenState(makeKoenigrufenState({ calledKing: 0 }))).not.toContain('called King:');
    expect(formatKoenigrufenState(makeKoenigrufenState({ calledKing: 1 }))).toContain('called King:');
  });

  it('renders every player score on the summary line', () => {
    expect(formatKoenigrufenState(makeKoenigrufenState())).toContain('P0=');
  });

  it("renders each player's cards, tricks and points", () => {
    const out = formatKoenigrufenState(makeKoenigrufenState());
    expect(out).toContain('cards=');
    expect(out).toContain('tricks=');
    expect(out).toContain('pts=');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], reason: 'lead_low' };
    expect(formatKoenigrufenState(makeKoenigrufenState({ hint, messageCode: 'koenigrufen.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatKoenigrufenState(makeKoenigrufenState({ hint, messageCode: 'koenigrufen.playing' }))).not.toContain(
      'HINT',
    );
  });
});
