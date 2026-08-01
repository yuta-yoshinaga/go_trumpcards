import { describe, expect, it } from 'vitest';
import type { NapoleonPlayerData, NapoleonResponse } from '../../../types/card';
import { formatNapoleonState } from './napoleonFormatter';

function makePlayer(overrides?: Partial<NapoleonPlayerData>): NapoleonPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    bid: 13,
    isNapoleon: false,
    isAdjutant: false,
    adjutantRevealed: false,
    pictureCards: 2,
    roundScore: 0,
    cumulativeScore: 4,
    trickCount: 1,
    ...overrides,
  };
}

function makeState(overrides?: Partial<NapoleonResponse>): NapoleonResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [] })],
    phase: 0,
    roundNumber: 1,
    trickNumber: 2,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    adjutantCard: { design: 'HEART', value: 1 },
    napoleonIdx: 0,
    adjutantIdx: 1,
    adjutantRevealed: false,
    highestBid: 13,
    highestBidder: 0,
    kitty: [],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 1, minBid: 13, pointLimit: 3 },
    message: '',
    ...overrides,
  };
}

describe('formatNapoleonState', () => {
  it('renders the header, round and trick', () => {
    const out = formatNapoleonState(makeState());
    expect(out).toContain('Napoleon');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 2');
  });

  // **入札前は切り札も宣言額も副官も決まっていない。**どの行も出ない。
  it('omits the trump, bid and adjutant lines before they are decided', () => {
    const out = formatNapoleonState(makeState({ trumpSuit: 0, highestBid: 0, adjutantCard: null }));
    expect(out).not.toContain('trump:');
    expect(out).not.toContain('bid:');
    expect(out).not.toContain('adjutant:');
  });

  it('renders the trump, bid and adjutant once they are decided', () => {
    const out = formatNapoleonState(makeState());
    expect(out).toContain('trump:');
    expect(out).toContain('bid: 13');
    expect(out).toContain('adjutant:');
  });

  it('marks the Napoleon', () => {
    const out = formatNapoleonState(makeState({ players: [makePlayer({ isNapoleon: true }), makePlayer({ id: 1 })] }));
    expect(out).toContain('[Napoleon]');
  });

  // **副官は名乗り出るまで伏せられている。**明かされて初めて印が付く。
  it('marks the adjutant only once revealed', () => {
    const hidden = formatNapoleonState(
      makeState({ players: [makePlayer({ isAdjutant: true, adjutantRevealed: false })] }),
    );
    expect(hidden).not.toContain('[Adjutant]');

    const shown = formatNapoleonState(
      makeState({ players: [makePlayer({ isAdjutant: true, adjutantRevealed: true })] }),
    );
    expect(shown).toContain('[Adjutant]');
  });

  it('renders the message when present', () => {
    expect(formatNapoleonState(makeState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatNapoleonState(makeState({ hint, messageCode: 'napoleon.hintRequested' }))).toContain('HINT');
    expect(formatNapoleonState(makeState({ hint, messageCode: 'napoleon.playing' }))).not.toContain('HINT');
  });
});
