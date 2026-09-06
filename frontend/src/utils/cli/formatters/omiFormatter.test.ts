import { describe, expect, it } from 'vitest';
import type { OmiPlayerData, OmiResponse } from '../../../types/card';
import { formatOmiState } from './omiFormatter';

function makePlayer(overrides?: Partial<OmiPlayerData>): OmiPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 8,
    cards: [{ design: 'SPADE', value: 1 }],
    team: 0,
    trickCount: 2,
    ...overrides,
  };
}

function makeState(overrides?: Partial<OmiResponse>): OmiResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [], team: 1 })],
    phase: 1, // OmiPhase.PLAY
    roundNumber: 1,
    trickNumber: 3,
    currentPlayerIdx: 0,
    trumpCallerIdx: 1,
    bidPlayerIdx: 1,
    dealerIdx: 1,
    trumpSuit: 1,
    dealStage: 2,
    faceUpCard: null,
    makerTeam: 1,
    goingAlone: false,
    goingAlonePlayerIdx: -1,
    currentTrick: [],
    teamScores: [2, 1],
    teamTricks: [3, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1, pointLimit: 10 },
    message: '',
    ...overrides,
  };
}

describe('formatOmiState', () => {
  it('renders the header, round, trick and team scores', () => {
    const out = formatOmiState(makeState());
    expect(out).toContain('Omi');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 3');
    expect(out).toContain('Team0=2');
    expect(out).toContain('Team1=1');
  });

  // Trump is shown only once called; caller name is included
  it('names the trump and caller once trump is decided', () => {
    const out = formatOmiState(makeState());
    expect(out).toContain('trump:');
    expect(out).toContain('caller:');
  });

  it('omits trump line before trump is called (suit 0)', () => {
    expect(formatOmiState(makeState({ trumpSuit: 0 }))).not.toContain('trump:');
  });

  // Team tricks are shown
  it('renders team trick counts for this round', () => {
    const out = formatOmiState(makeState());
    expect(out).toContain('tricks: Team0=3');
    expect(out).toContain('Team1=2');
  });

  // Scoring rules should appear after trump is set
  it('shows Omi scoring rules after trump is decided', () => {
    const out = formatOmiState(makeState());
    // 5 tricks = 1 point
    expect(out).toContain('5+');
    // All 8 = 2 points
    expect(out).toContain('Omi!');
    // 4-4 = 0 points
    expect(out).toContain('4-4');
  });

  it('does not show scoring rules when trump not yet called', () => {
    const out = formatOmiState(makeState({ trumpSuit: 0 }));
    expect(out).not.toContain('scoring');
  });

  // Deal stage annotation
  it('annotates hand when dealStage=1 (4 cards dealt, awaiting trump)', () => {
    const out = formatOmiState(makeState({ dealStage: 1 }));
    expect(out).toContain('awaiting trump');
  });

  it('does not annotate hand for dealStage=2', () => {
    const out = formatOmiState(makeState({ dealStage: 2 }));
    expect(out).not.toContain('awaiting trump');
  });

  it("renders each player's team and trick count", () => {
    const out = formatOmiState(makeState());
    expect(out).toContain('team=0');
    expect(out).toContain('tricks=2');
  });

  it('renders a hint line', () => {
    expect(
      formatOmiState(makeState({ hint: { cardIndex: 0, reason: 'follow_suit' }, messageCode: 'omi.hintRequested' })),
    ).toContain('HINT');
  });

  it('renders the message when present', () => {
    expect(formatOmiState(makeState({ message: 'done' }))).toContain('done');
  });

  // HINT line only when hint was requested
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatOmiState(makeState({ hint, messageCode: 'omi.hintRequested' }))).toContain('HINT');
    expect(formatOmiState(makeState({ hint, messageCode: 'omi.playing' }))).not.toContain('HINT');
  });

  // No Euchre content
  it('does not contain Euchre or bower references', () => {
    const out = formatOmiState(makeState());
    expect(out).not.toMatch(/[Ee]uchre/);
    expect(out).not.toMatch(/[Bb]ower/);
    expect(out).not.toMatch(/バウアー/);
    expect(out).not.toMatch(/pickup|pick_up/i);
  });
});
