import { describe, expect, it } from 'vitest';
import type { SpoonsPlayer, SpoonsResponse } from '../../../types/card';
import { formatSpoonsState } from './spoonsFormatter';

function makePlayer(overrides: Partial<SpoonsPlayer> = {}): SpoonsPlayer {
  return {
    name: 'CPU',
    isHuman: false,
    handSize: 4,
    hand: [],
    letters: 0,
    eliminated: false,
    hasSpoon: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<SpoonsResponse> = {}): SpoonsResponse {
  return {
    phase: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    currentPlayerIdx: 0,
    feederIdx: 0,
    isHumanTurn: true,
    spoonsRemaining: 3,
    grabWindowOpen: false,
    firstGrabberIdx: -1,
    roundLoserIdx: -1,
    roundNumber: 1,
    drawPileSize: 36,
    cpuDifficulty: 1,
    message: '',
    players: [
      makePlayer({
        name: 'You',
        isHuman: true,
        hand: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 1 },
        ],
      }),
      makePlayer(),
      makePlayer(),
      makePlayer(),
    ],
    ...overrides,
  };
}

describe('formatSpoonsState', () => {
  it('includes the header, round, phase, spoons and draw pile', () => {
    const out = formatSpoonsState(makeState());
    expect(out).toContain('Spoons');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: Pass');
    expect(out).toContain('spoons left: 3');
    expect(out).toContain('draw pile: 36');
  });

  it("renders the human's hand with indices", () => {
    const out = formatSpoonsState(makeState());
    expect(out).toContain('[0]');
    expect(out).toContain('[1]');
  });

  it('shows the pass prompt on the human turn during the pass phase', () => {
    const out = formatSpoonsState(makeState());
    expect(out).toContain('pass a card');
  });

  it('shows the grab prompt when the grab window is open', () => {
    const out = formatSpoonsState(makeState({ phase: 1, grabWindowOpen: true }));
    expect(out).toContain('GRAB');
  });

  it('renders letters and status badges', () => {
    const out = formatSpoonsState(
      makeState({
        players: [
          makePlayer({ name: 'You', isHuman: true, letters: 2 }),
          makePlayer({ hasSpoon: true }),
          makePlayer({ eliminated: true, letters: 6 }),
          makePlayer(),
        ],
      }),
    );
    expect(out).toContain('letters=S-P');
    expect(out).toContain('[has spoon]');
    expect(out).toContain('[OUT]');
  });

  it('announces the winner on game end', () => {
    const out = formatSpoonsState(makeState({ phase: 3, gameEndFlag: true, winnerIdx: 0 }));
    expect(out).toContain('Game Over');
    expect(out).toContain('Winner');
  });
});
