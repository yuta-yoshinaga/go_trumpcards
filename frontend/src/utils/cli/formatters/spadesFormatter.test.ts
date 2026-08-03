import { describe, expect, it } from 'vitest';
import { makeSpadesState } from '../../../test/stateFactories';
import { formatSpadesState } from './spadesFormatter';

describe('formatSpadesState', () => {
  it('renders the header, round, trick and whether spades are broken', () => {
    const out = formatSpadesState(makeSpadesState());
    expect(out).toContain('Spades');
    expect(out).toContain('round:');
    expect(out).toContain('trick:');
    expect(out).toContain('spades broken:');
  });

  it('reports broken spades', () => {
    expect(formatSpadesState(makeSpadesState({ spadesBroken: true }))).toContain('spades broken: yes');
    expect(formatSpadesState(makeSpadesState({ spadesBroken: false }))).toContain('spades broken: no');
  });

  it('renders a bid hint and a play hint differently', () => {
    const bid = formatSpadesState(
      makeSpadesState({ hint: { bid: 3, reason: 'strategic_bid' }, messageCode: 'spades.hintRequested' }),
    );
    expect(bid).toContain('HINT: bid 3');

    const play = formatSpadesState(
      makeSpadesState({ hint: { cardIndex: 2, reason: 'follow_suit' }, messageCode: 'spades.hintRequested' }),
    );
    expect(play).toContain('HINT: play [2]');
  });

  it('renders the message when present', () => {
    expect(formatSpadesState(makeSpadesState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 2, reason: 'follow_suit' };
    expect(formatSpadesState(makeSpadesState({ hint, messageCode: 'spades.hintRequested' }))).toContain('HINT');
    expect(formatSpadesState(makeSpadesState({ hint, messageCode: 'spades.playing' }))).not.toContain('HINT');
  });
});
