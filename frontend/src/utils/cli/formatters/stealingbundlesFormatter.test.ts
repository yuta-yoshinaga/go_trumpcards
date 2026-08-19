import { describe, expect, it } from 'vitest';
import type { Card, StealingBundlesResponse } from '../../../types/card';
import { formatStealingBundlesState } from './stealingbundlesFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? [card('SPADE', 7), card('HEART', 9)] : [],
  bundleSize: 0,
  ...over,
});

const state = (over: Partial<StealingBundlesResponse> = {}): StealingBundlesResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    tableCards: [card('CLOVER', 7)],
    tableMatches: { '0': [0] },
    stealTargets: {},
    canCapture: true,
    deckRemaining: 32,
    lastCaptureKind: '',
    lastCaptureVictimIdx: -1,
    lastCaptureIdx: -1,
    currentPlayerIdx: 0,
    turnNumber: 2,
    packsDealt: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4 },
    message: '',
    ...over,
  }) as unknown as StealingBundlesResponse;

describe('formatStealingBundlesState', () => {
  it('reports loading for a null state', () => {
    expect(formatStealingBundlesState(null)).toBe('Loading...');
  });

  it('shows the turn, deck and the rule', () => {
    const out = formatStealingBundlesState(state());
    expect(out).toContain('turn 3');
    expect(out).toContain('deck 32');
    expect(out).toMatch(/rival bundle goes whole/);
  });

  // **空の場も情報。** 行が消えると見落としと区別が付きません。
  it('shows the table, empty or not', () => {
    expect(formatStealingBundlesState(state())).toMatch(/table: /);
    expect(formatStealingBundlesState(state({ tableCards: [] }))).toMatch(/table: \(empty\)/);
  });

  // **束の一番上は全員に見えます。**
  it('shows every bundle size and top card', () => {
    const out = formatStealingBundlesState(
      state({ players: [seat(0), seat(1, { bundleSize: 5, bundleTop: card('DIAMOND', 9) })] }),
    );
    expect(out).toMatch(/bundle 5, top /);
    expect(out).toMatch(/bundle 0, top none/);
  });

  it('marks who captured last, and how', () => {
    expect(formatStealingBundlesState(state({ lastCaptureIdx: 2, lastCaptureKind: 'take' }))).toContain(
      'CPU 2[took from the table]',
    );
    // **盗みは別の出来事** (#5767)。被害者まで出す。
    expect(
      formatStealingBundlesState(state({ lastCaptureIdx: 2, lastCaptureKind: 'steal', lastCaptureVictimIdx: 0 })),
    ).toContain("CPU 2[stole あなた's bundle]");
    expect(formatStealingBundlesState(state())).not.toContain('[took from the table]');
  });

  // **どの札で何ができるかは盤面から読み切れません。**
  it('marks which hand cards capture and which steal', () => {
    const out = formatStealingBundlesState(state({ tableMatches: { '0': [0] }, stealTargets: { '1': [2, 3] } }));
    expect(out).toMatch(/\[0\]\S+\*/);
    expect(out).toMatch(/\[1\]\S+!23/);
    expect(out).toMatch(/captures from the table/);
  });

  // **取れるときは置けません。** 言わないと trail が弾かれる理由が読めません。
  it('says whether a capture is compulsory', () => {
    expect(formatStealingBundlesState(state())).toMatch(/cannot place a card/);
    expect(formatStealingBundlesState(state({ canCapture: false }))).toMatch(/nothing can be captured/);
  });

  it('names the winner once the game ends', () => {
    const out = formatStealingBundlesState(
      state({ phase: 1, gameEndFlag: true, winnerIdx: 0, players: [seat(0, { bundleSize: 30 }), seat(1)] }),
    );
    expect(out).toMatch(/game over — あなた collected the most \(30\)/);
    // 終局後は促さない。
    expect(out).not.toMatch(/cannot place a card/);
  });

  it('appends the server message', () => {
    expect(formatStealingBundlesState(state({ message: 'boom' }))).toContain('boom');
  });
});
