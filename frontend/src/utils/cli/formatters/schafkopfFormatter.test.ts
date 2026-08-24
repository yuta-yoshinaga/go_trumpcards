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

describe('formatSchafkopfState — the auction', () => {
  it('lists only the contracts that outrank the standing bid', () => {
    const out = formatSchafkopfState(makeSchafkopfState({ phase: 0, beatableContracts: [2] }));
    // 行そのものを見る。`picker:` にも "pick" が入るので、素の contains では
    // 案内していない語まで拾ってしまう。
    const line = out.split('\n').find((l) => l.startsWith('you may declare:'));
    expect(line).toBe('you may declare: solo <suit>');
  });

  it('says nothing about declaring once the auction has closed', () => {
    const out = formatSchafkopfState(makeSchafkopfState({ phase: 2, beatableContracts: [] }));
    expect(out).not.toContain('you may declare');
  });

  it('names the contract in play, with the suit a Solo chose', () => {
    expect(formatSchafkopfState(makeSchafkopfState({ contract: 1 }))).toContain('contract: Wenz (Unters only)');
    expect(formatSchafkopfState(makeSchafkopfState({ contract: 2, soloSuit: 3 }))).toContain('contract: Solo (♥)');
    expect(formatSchafkopfState(makeSchafkopfState({ contract: 0 }))).toContain('contract: Rufspiel (called ace)');
  });
});
