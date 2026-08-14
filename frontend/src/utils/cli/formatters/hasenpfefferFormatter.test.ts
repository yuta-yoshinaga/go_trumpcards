import { describe, expect, it } from 'vitest';
import type { Card, HasenpfefferResponse } from '../../../types/card';
import { formatHasenpfefferState } from './hasenpfefferFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 1)] : [],
  bid: -1,
  trickCount: 0,
  ...over,
});

const state = (over: Partial<HasenpfefferResponse> = {}): HasenpfefferResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 2,
    handNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    declarerIdx: 1,
    contract: 4,
    minBid: 5,
    mustBid: false,
    blindSize: 0,
    scores: [3, 5],
    teamTricks: [2, 1],
    lastHandEuchred: false,
    lastHandTricks: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 3,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 10 },
    message: '',
    ...over,
  }) as unknown as HasenpfefferResponse;

describe('formatHasenpfefferState', () => {
  it('reports loading for a null state', () => {
    expect(formatHasenpfefferState(null)).toBe('Loading...');
  });

  it('shows the hand, trick, target and phase', () => {
    const out = formatHasenpfefferState(state());
    expect(out).toContain('hand 2');
    expect(out).toContain('trick 4/6');
    expect(out).toContain('first to 10');
    expect(out).toContain('PLAY');
  });

  // **ジョーカーが最強という序列を毎回書く。**
  it('states the joker ranking', () => {
    expect(formatHasenpfefferState(state())).toContain('the joker is the highest trump of all');
  });

  // 伏せ札・未宣言・確定の 3 状態を踏む。
  it('shows the blind while the auction runs', () => {
    const out = formatHasenpfefferState(state({ phase: 0, trumpSuit: 0, blindSize: 1, declarerIdx: -1 }));
    expect(out).toContain('blind: 1 card');
  });

  it('says trump is not named yet once the blind is taken', () => {
    expect(formatHasenpfefferState(state({ phase: 1, trumpSuit: 0, blindSize: 0 }))).toContain('trump: not named yet');
  });

  it('names the suit and the contract once trump is declared', () => {
    const out = formatHasenpfefferState(state());
    expect(out).toContain('trump: ♥');
    expect(out).toContain('bid 4');
    expect(formatHasenpfefferState(state({ trumpSuit: 9 }))).toContain('trump: ?');
  });

  // **親は降りられないことがある。** 選択肢が無い場面を明示する。
  it('says when the dealer cannot pass', () => {
    expect(formatHasenpfefferState(state({ phase: 0, mustBid: true }))).toContain('you cannot pass');
  });

  it('says when the maximum bid is already standing', () => {
    const out = formatHasenpfefferState(state({ phase: 0, mustBid: false, minBid: 0 }));
    expect(out).toContain('you can only pass');
  });

  // **宣言の状態は 3 通り。** 未宣言 / 降り / 数字を取り違えない。
  it('renders every bid state', () => {
    const out = formatHasenpfefferState(
      state({
        players: [seat(0, { bid: 4 }), seat(1, { bid: 0 }), seat(2), seat(3)],
      } as Partial<HasenpfefferResponse>),
    );
    expect(out).toContain('bid 4');
    expect(out).toContain('bid passed');
    expect(out).toContain('bid -');
  });

  it('marks the declarer and the dealer', () => {
    const out = formatHasenpfefferState(state());
    expect(out).toContain('[declarer]');
    expect(out).toContain('[dealer]');
  });

  it('marks the playable cards in your hand', () => {
    expect(formatHasenpfefferState(state())).toMatch(/your hand: \[0\].*\*/);
  });

  it('says the hand is empty rather than printing nothing', () => {
    const out = formatHasenpfefferState(
      state({ players: [{ ...seat(0), cards: [] }, seat(1), seat(2), seat(3)] } as Partial<HasenpfefferResponse>),
    );
    expect(out).toContain('your hand: (empty)');
  });

  it('omits the hand block when no seat is human', () => {
    expect(formatHasenpfefferState(state({ players: [seat(1), seat(2), seat(3)] }))).not.toContain('your hand:');
  });

  it('shows the current trick when one is in progress', () => {
    expect(formatHasenpfefferState(state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 9) }] }))).toContain(
      'trick: ',
    );
  });

  // **落としたのか達成したのかは盤面から読めない。** 両側を踏む。
  it.each([
    [false, 'made'],
    [true, 'euchred'],
  ])('reports the hand outcome (euchred=%s)', (lastHandEuchred, expected) => {
    const out = formatHasenpfefferState(state({ phase: 3, lastHandEuchred, lastHandTricks: 3 }));
    expect(out).toContain(expected);
  });

  it.each([
    [0, 'team 0 wins'],
    [1, 'team 1 wins'],
    [-1, 'tie'],
  ])('reports the game result for winnerTeam %s', (winnerTeam, expected) => {
    expect(formatHasenpfefferState(state({ phase: 4, gameEndFlag: true, winnerTeam }))).toContain(expected);
  });

  it('falls back to the raw phase for an unknown value', () => {
    expect(formatHasenpfefferState(state({ phase: 99 }))).toContain('99');
  });

  it('echoes a server message', () => {
    expect(formatHasenpfefferState(state({ message: 'must follow the led suit' }))).toContain(
      'must follow the led suit',
    );
  });
});
