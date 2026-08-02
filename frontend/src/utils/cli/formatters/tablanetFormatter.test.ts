import { describe, expect, it } from 'vitest';
import { makeTablanetState } from '../../../test/stateFactories';
import { formatTablanetState } from './tablanetFormatter';

describe('formatTablanetState', () => {
  it('renders header, deal/deck line and the table', () => {
    const out = formatTablanetState(makeTablanetState());
    expect(out).toContain('Tablanet');
    expect(out).toContain('deal: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('table:');
  });

  it('renders "(empty)" when the table has no cards', () => {
    const out = formatTablanetState(makeTablanetState({ tableCards: [] }));
    expect(out).toContain('(empty)');
  });

  it('renders the human hand and per-player capture/tabla counts', () => {
    const out = formatTablanetState(makeTablanetState());
    expect(out).toContain('captured=0');
    expect(out).toContain('tabla=0');
  });

  it('renders final scores and winner at game end', () => {
    const out = formatTablanetState(
      makeTablanetState({
        phase: 1,
        gameEndFlag: true,
        winners: [0],
        lastDealDetail: {
          cards: { 0: 30, 1: 22 },
          aces: { 0: 2 },
          jacks: { 0: 1 },
          tablas: { 0: 1 },
          hasTenDiamonds: 0,
          hasTwoClubs: 0,
          mostCards: 0,
          gained: { 0: 16, 1: 4 },
        },
      }),
    );
    expect(out).toContain('final:');
    expect(out).toContain('winner:');
  });

  it('renders a hint line when a hint is present', () => {
    const out = formatTablanetState(
      makeTablanetState({
        messageCode: 'tablanet.hintRequested',
        hint: { cardIndices: [0], tableIndices: [0], reason: 'tabla_sweep' },
      }),
    );
    expect(out).toContain('HINT:');
    expect(out).toContain('tabla_sweep');
  });

  it('appends the server message when present', () => {
    const out = formatTablanetState(makeTablanetState({ message: 'Game over.' }));
    expect(out).toContain('Game over.');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [0], tableIndices: [0], reason: 'tabla_sweep' };
    expect(formatTablanetState(makeTablanetState({ hint, messageCode: 'tablanet.hintRequested' }))).toContain('HINT');
    expect(formatTablanetState(makeTablanetState({ hint, messageCode: 'tablanet.playing' }))).not.toContain('HINT');
  });
});
