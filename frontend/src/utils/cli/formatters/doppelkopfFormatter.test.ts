import { describe, expect, it } from 'vitest';
import { makeDoppelkopfState } from '../../../test/stateFactories';
import { formatDoppelkopfState } from './doppelkopfFormatter';

describe('formatDoppelkopfState', () => {
  it('renders the header', () => {
    expect(formatDoppelkopfState(makeDoppelkopfState())).toContain('Doppelkopf');
  });

  // 進行中の獲得点。GUI のパネルと CUI の 1 行が出しているのに、CLI モードだけが
  // 何も出していなかった (#6435)。値はレスポンス由来で、240 からの引き算ではない。
  it('shows the running card points while the round is in progress', () => {
    const out = formatDoppelkopfState(makeDoppelkopfState({ phase: 0, liveRePoints: 30, liveKontraPoints: 20 }));
    expect(out).toContain('card points: Re=30 Kontra=20');
    // 240 - 30 = 210 とは書かない。
    expect(out).not.toContain('210');
  });

  it('drops the running line once the round has ended', () => {
    const out = formatDoppelkopfState(
      makeDoppelkopfState({ phase: 2, liveRePoints: 130, liveKontraPoints: 110, roundRePoints: 130 }),
    );
    expect(out).not.toContain('card points:');
    expect(out).toContain('round result:');
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
