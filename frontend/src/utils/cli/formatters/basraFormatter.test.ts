import { describe, expect, it } from 'vitest';
import { makeBasraState } from '../../../test/stateFactories';
import { formatBasraState } from './basraFormatter';

describe('formatBasraState', () => {
  it('renders header, deal/deck line and the table', () => {
    const out = formatBasraState(makeBasraState());
    expect(out).toContain('Basra');
    expect(out).toContain('deal: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('table:');
  });

  it('renders "(empty)" when the table has no cards', () => {
    const out = formatBasraState(makeBasraState({ tableCards: [] }));
    expect(out).toContain('(empty)');
  });

  it('renders the human hand and per-player capture/basra counts', () => {
    const out = formatBasraState(makeBasraState());
    expect(out).toContain('captured=0');
    expect(out).toContain('basra=0');
  });

  it('renders final scores and winner at game end', () => {
    const out = formatBasraState(
      makeBasraState({
        phase: 1,
        gameEndFlag: true,
        winners: [0],
        lastDealDetail: {
          cards: { 0: 30, 1: 22 },
          aces: { 0: 2 },
          basras: { 0: 1 },
          hasSevenDiamonds: 0,
          hasTenDiamonds: 0,
          mostCards: 0,
          gained: { 0: 16, 1: 4 },
        },
      }),
    );
    expect(out).toContain('final:');
    expect(out).toContain('winner:');
  });

  it('renders a hint line when a hint is present', () => {
    const out = formatBasraState(
      makeBasraState({
        messageCode: 'basra.hintRequested',
        hint: { cardIndices: [0], tableIndices: [0], reason: 'basra_sweep' },
      }),
    );
    expect(out).toContain('HINT:');
    expect(out).toContain('basra_sweep');
  });

  it('appends the server message when present', () => {
    const out = formatBasraState(makeBasraState({ message: 'Game over.' }));
    expect(out).toContain('Game over.');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [0], tableIndices: [0], reason: 'basra_sweep' };
    expect(formatBasraState(makeBasraState({ hint, messageCode: 'basra.hintRequested' }))).toContain('HINT');
    expect(formatBasraState(makeBasraState({ hint, messageCode: 'basra.playing' }))).not.toContain('HINT');
  });
});
