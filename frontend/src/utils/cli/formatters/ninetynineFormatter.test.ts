import { describe, expect, it } from 'vitest';
import type { NinetyNinePlayerData, NinetyNineResponse } from '../../../types/card';
import { formatNinetynineState } from './ninetynineFormatter';

function makePlayer(overrides?: Partial<NinetyNinePlayerData>): NinetyNinePlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    bid: 2,
    roundScore: 0,
    cumulativeScore: 15,
    trickCount: 1,
    buriedCount: 3,
    ...overrides,
  };
}

function makeState(overrides?: Partial<NinetyNineResponse>): NinetyNineResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [] })],
    phase: 0,
    dealNumber: 2,
    targetScore: 100,
    handSize: 12,
    trickNumber: 3,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    trumpSuit: 1,
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1, targetScore: 100 },
    message: '',
    ...overrides,
  };
}

describe('formatNinetynineState', () => {
  it('renders the header, deal, trick and hand size', () => {
    const out = formatNinetynineState(makeState());
    expect(out).toContain('Ninety-Nine');
    expect(out).toContain('deal: 2');
    expect(out).toContain('trick: 3');
    expect(out).toContain('hand size: 12');
  });

  it('renders the trump and the target score', () => {
    const out = formatNinetynineState(makeState());
    expect(out).toContain('trump:');
    expect(out).toContain('target: 100');
  });

  it('renders the current trick when one is under way', () => {
    const out = formatNinetynineState(
      makeState({ currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 9 } }] }),
    );
    expect(out).toContain('trick:');
  });

  // **伏せ札とプレイでヒント行が変わる。**
  it('renders a bury hint and a play hint differently', () => {
    const bury = formatNinetynineState(
      makeState({ hint: { buryIndices: [0, 2], reason: 'bury_low' }, messageCode: 'ninetynine.hintRequested' }),
    );
    expect(bury).toContain('HINT: bury [0, 2]');

    const play = formatNinetynineState(
      makeState({ hint: { cardIndex: 1, reason: 'follow_suit' }, messageCode: 'ninetynine.hintRequested' }),
    );
    expect(play).toContain('HINT: play [1]');
  });

  it('renders the message when present', () => {
    expect(formatNinetynineState(makeState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 1, reason: 'follow_suit' };
    expect(formatNinetynineState(makeState({ hint, messageCode: 'ninetynine.hintRequested' }))).toContain('HINT');
    expect(formatNinetynineState(makeState({ hint, messageCode: 'ninetynine.playing' }))).not.toContain('HINT');
  });
});
