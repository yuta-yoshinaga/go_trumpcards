import { describe, expect, it } from 'vitest';
import type { EuchrePlayerData, EuchreResponse } from '../../../types/card';
import { formatEuchreState } from './euchreFormatter';

function makePlayer(overrides?: Partial<EuchrePlayerData>): EuchrePlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    team: 0,
    trickCount: 2,
    ...overrides,
  };
}

function makeState(overrides?: Partial<EuchreResponse>): EuchreResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [], team: 1 })],
    phase: 0,
    roundNumber: 1,
    trickNumber: 3,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    trumpSuit: 1,
    faceUpCard: { design: 'HEART', value: 11 },
    makerTeam: 0,
    goingAlone: false,
    goingAlonePlayerIdx: -1,
    currentTrick: [],
    teamScores: [2, 1],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1, pointLimit: 10 },
    message: '',
    ...overrides,
  };
}

describe('formatEuchreState', () => {
  it('renders the header, round, trick and team scores', () => {
    const out = formatEuchreState(makeState());
    expect(out).toContain('Euchre');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 3');
    expect(out).toContain('Team0=2');
    expect(out).toContain('Team1=1');
  });

  // **切り札は入札が決まるまで無い。**suit 0 のあいだ行ごと出ない。
  it('names the trump only once it is decided', () => {
    expect(formatEuchreState(makeState())).toContain('trump:');
    expect(formatEuchreState(makeState({ trumpSuit: 0 }))).not.toContain('trump:');
  });

  // **一人打ちは宣言されたときだけ告知する。**
  it('announces going alone only when someone did', () => {
    expect(formatEuchreState(makeState())).not.toContain('Going alone');
    expect(formatEuchreState(makeState({ goingAlone: true }))).toContain('Going alone!');
  });

  it("renders each player's team and trick count", () => {
    const out = formatEuchreState(makeState());
    expect(out).toContain('team=0');
    expect(out).toContain('tricks=2');
  });

  it('renders a hint line', () => {
    expect(
      formatEuchreState(
        makeState({ hint: { cardIndex: 0, reason: 'follow_suit' }, messageCode: 'euchre.hintRequested' }),
      ),
    ).toContain('HINT');
  });

  it('renders the message when present', () => {
    expect(formatEuchreState(makeState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatEuchreState(makeState({ hint, messageCode: 'euchre.hintRequested' }))).toContain('HINT');
    expect(formatEuchreState(makeState({ hint, messageCode: 'euchre.playing' }))).not.toContain('HINT');
  });
});
