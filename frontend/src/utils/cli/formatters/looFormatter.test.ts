import { describe, expect, it } from 'vitest';
import { makeLooState } from '../../../test/stateFactories';
import { formatLooState } from './looFormatter';

describe('formatLooState', () => {
  it('renders the header, the pot and the trump', () => {
    const out = formatLooState(makeLooState());
    expect(out).toContain('Loo');
    expect(out).toContain('pot:');
    expect(out).toContain('trump:');
  });

  it("renders each player's tricks and chips", () => {
    const out = formatLooState(makeLooState());
    expect(out).toContain('tricks=');
    expect(out).toContain('chips=');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [0], reason: 'lead_low' };
    expect(formatLooState(makeLooState({ hint, messageCode: 'loo.hintRequested' }))).toContain('HINT');
    expect(formatLooState(makeLooState({ hint, messageCode: 'loo.playing' }))).not.toContain('HINT');
  });
});
