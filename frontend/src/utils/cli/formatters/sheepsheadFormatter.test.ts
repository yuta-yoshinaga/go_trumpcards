import { describe, expect, it } from 'vitest';
import { makeSheepsheadState } from '../../../test/stateFactories';
import { formatSheepsheadState } from './sheepsheadFormatter';

describe('formatSheepsheadState', () => {
  it('renders the header and the trick number', () => {
    const out = formatSheepsheadState(makeSheepsheadState());
    expect(out).toContain('Sheepshead');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], suit: 1, pick: false, reason: 'follow_suit' };
    expect(formatSheepsheadState(makeSheepsheadState({ hint, messageCode: 'sheepshead.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatSheepsheadState(makeSheepsheadState({ hint, messageCode: 'sheepshead.playing' }))).not.toContain(
      'HINT',
    );
  });
});
