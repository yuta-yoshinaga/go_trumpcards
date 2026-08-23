import { describe, expect, it } from 'vitest';
import { makeSchafkopfState } from '../../../test/stateFactories';
import { formatSchafkopfState } from './schafkopfFormatter';

describe('formatSchafkopfState', () => {
  it('renders the header and the trick number', () => {
    const out = formatSchafkopfState(makeSchafkopfState());
    expect(out).toContain('Schafkopf');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1], suit: 1, pick: false, reason: 'follow_suit' };
    expect(formatSchafkopfState(makeSchafkopfState({ hint, messageCode: 'schafkopf.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatSchafkopfState(makeSchafkopfState({ hint, messageCode: 'schafkopf.playing' }))).not.toContain('HINT');
  });
});
