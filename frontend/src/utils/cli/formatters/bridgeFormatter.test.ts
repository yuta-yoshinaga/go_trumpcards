import { describe, expect, it } from 'vitest';
import type { BridgePlayerData, BridgeResponse } from '../../../types/card';
import { formatBridgeState } from './bridgeFormatter';

function makePlayer(overrides?: Partial<BridgePlayerData>): BridgePlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    team: 0,
    trickCount: 3,
    ...overrides,
  };
}

function makeState(overrides?: Partial<BridgeResponse>): BridgeResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [], team: 1 })],
    phase: 0,
    roundNumber: 1,
    trickNumber: 4,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    trumpSuit: 1,
    contractLevel: 3,
    contractSuit: 1,
    doubled: 0,
    declarerIdx: 0,
    dummyIdx: 1,
    bidHistory: [],
    vulnerability: [false, false],
    currentTrick: [],
    teamScores: [90, 40],
    gamesWon: [1, 0],
    belowLine: [90, 40],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    openingLeadDone: true,
    dummyHand: null,
    config: { cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

describe('formatBridgeState', () => {
  it('renders the header, round and trick', () => {
    const out = formatBridgeState(makeState());
    expect(out).toContain('Contract Bridge');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 4');
  });

  // **契約は競りが決まるまで無い。**level 0 のあいだ行ごと出ない。
  it('omits the contract until one is settled', () => {
    expect(formatBridgeState(makeState({ contractLevel: 0 }))).not.toContain('contract:');
    expect(formatBridgeState(makeState())).toContain('contract: 3');
  });

  // **ダブルとリダブルで表記が変わる。**0 なら何も付かない。
  it('marks a doubled and a redoubled contract', () => {
    expect(formatBridgeState(makeState({ doubled: 0 }))).not.toContain(' X');
    expect(formatBridgeState(makeState({ doubled: 1 }))).toContain(' X');
    expect(formatBridgeState(makeState({ doubled: 2 }))).toContain(' XX');
  });

  it('renders the rubber score and games won', () => {
    const out = formatBridgeState(makeState());
    expect(out).toContain('NS=90');
    expect(out).toContain('EW=40');
    expect(out).toContain('games: NS=1');
  });

  // **デクレアラーとダミーには印が付く。**
  it('marks the declarer and the dummy', () => {
    const out = formatBridgeState(makeState());
    expect(out).toContain('[Declarer]');
    expect(out).toContain('[Dummy]');
  });

  it('renders the message when present', () => {
    expect(formatBridgeState(makeState({ message: 'done' }))).toContain('done');
  });

  // **ダミーの手札はオープニングリードの後だけ見える。**null のうちは行ごと出ない。
  it('shows the dummy hand only once it is exposed', () => {
    expect(formatBridgeState(makeState())).not.toContain('dummy:');
    expect(formatBridgeState(makeState({ dummyHand: [{ design: 'HEART', value: 10 }] }))).toContain('dummy:');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatBridgeState(makeState({ hint, messageCode: 'bridge.hintRequested' }))).toContain('HINT');
    expect(formatBridgeState(makeState({ hint, messageCode: 'bridge.playing' }))).not.toContain('HINT');
  });
});
