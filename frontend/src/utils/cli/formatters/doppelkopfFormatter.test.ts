import { describe, expect, it } from 'vitest';
import { makeDoppelkopfState } from '../../../test/stateFactories';
import { formatDoppelkopfState } from './doppelkopfFormatter';

describe('formatDoppelkopfState', () => {
  it('renders the header', () => {
    expect(formatDoppelkopfState(makeDoppelkopfState())).toContain('Doppelkopf');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], reason: 'follow_suit' };
    expect(formatDoppelkopfState(makeDoppelkopfState({ hint, messageCode: 'doppelkopf.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatDoppelkopfState(makeDoppelkopfState({ hint, messageCode: 'doppelkopf.playing' }))).not.toContain(
      'HINT',
    );
  });
});
