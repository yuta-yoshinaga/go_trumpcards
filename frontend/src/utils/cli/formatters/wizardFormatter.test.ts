import { describe, expect, it } from 'vitest';
import type { WizardPlayerData, WizardResponse } from '../../../types/card';
import { formatWizardState } from './wizardFormatter';

function makePlayer(overrides?: Partial<WizardPlayerData>): WizardPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    bid: 2,
    roundScore: 20,
    cumulativeScore: 60,
    trickCount: 2,
    ...overrides,
  };
}

function makeState(overrides?: Partial<WizardResponse>): WizardResponse {
  return {
    players: [makePlayer(), makePlayer({ id: 1, isHuman: false, cards: [] })],
    phase: 0,
    roundNumber: 3,
    totalRounds: 20,
    handSize: 3,
    trickNumber: 2,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    currentTrick: [],
    trumpCard: { design: 'HEART', value: 11 },
    trumpSuit: 3,
    restrictedBid: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

describe('formatWizardState', () => {
  it('renders the header, round out of total, trick and hand size', () => {
    const out = formatWizardState(makeState());
    expect(out).toContain('Wizard');
    expect(out).toContain('round: 3/20');
    expect(out).toContain('trick: 2');
    expect(out).toContain('hand size: 3');
  });

  // **切り札なしのラウンドがある。**最終ラウンドは山が尽きて trumpCard が null。
  it('names the trump card only when there is one', () => {
    expect(formatWizardState(makeState())).toContain('trump:');
    expect(formatWizardState(makeState({ trumpCard: null }))).not.toContain('trump:');
  });

  it("renders each player's score, bid and tricks", () => {
    const out = formatWizardState(makeState());
    expect(out).toContain('total=60');
    expect(out).toContain('bid=2');
    expect(out).toContain('tricks=2');
  });

  it('renders the current trick when one is under way', () => {
    const out = formatWizardState(
      makeState({ currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 5 } }] }),
    );
    expect(out).toContain('trick:');
  });

  // **入札とプレイでヒント行が変わる。**
  it('renders a bid hint and a play hint differently', () => {
    expect(
      formatWizardState(makeState({ hint: { bid: 1, reason: 'strategic_bid' }, messageCode: 'wizard.hintRequested' })),
    ).toContain('HINT: bid 1');
    expect(
      formatWizardState(
        makeState({ hint: { cardIndex: 0, reason: 'follow_suit' }, messageCode: 'wizard.hintRequested' }),
      ),
    ).toContain('HINT: play [0]');
  });

  it('renders the message when present', () => {
    expect(formatWizardState(makeState({ message: 'done' }))).toContain('done');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatWizardState(makeState({ hint, messageCode: 'wizard.hintRequested' }))).toContain('HINT');
    expect(formatWizardState(makeState({ hint, messageCode: 'wizard.playing' }))).not.toContain('HINT');
  });
});
