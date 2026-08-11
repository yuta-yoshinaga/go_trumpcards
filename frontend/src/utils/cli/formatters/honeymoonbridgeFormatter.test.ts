import { describe, expect, it } from 'vitest';
import type { Card, HoneymoonBridgeResponse } from '../../../types/card';
import { formatHoneymoonBridgeState } from './honeymoonbridgeFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 8)] : [],
  bidLevel: 0,
  bidSuit: 0,
  trickCount: 0,
  score: 0,
  ...over,
});

const state = (over: Partial<HoneymoonBridgeResponse> = {}): HoneymoonBridgeResponse =>
  ({
    players: [seat(0), seat(1)],
    phase: 2,
    roundNumber: 2,
    trickNumber: 3,
    stockSize: 0,
    trumpSuit: 3,
    declarerIdx: 0,
    contractLevel: 2,
    requiredTricks: 8,
    minBidLevel: 0,
    minBidSuit: 0,
    lastMade: false,
    lastTricks: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { target: 100 },
    message: '',
    ...over,
  }) as unknown as HoneymoonBridgeResponse;

describe('formatHoneymoonBridgeState', () => {
  it('reports loading for a null state', () => {
    expect(formatHoneymoonBridgeState(null)).toBe('Loading...');
  });

  it('shows the deal, trick, target and phase', () => {
    const out = formatHoneymoonBridgeState(state());
    expect(out).toContain('deal 2');
    expect(out).toContain('trick 4/13');
    expect(out).toContain('target 100');
    expect(out).toContain('PLAY');
  });

  // **前半のトリックは得点にならない。** 山札の残りだけが意味を持つ。
  it('says the draw half does not score, and only there', () => {
    const drawing = formatHoneymoonBridgeState(state({ phase: 0, stockSize: 26, contractLevel: 0, declarerIdx: -1 }));
    expect(drawing).toContain('stock: 26 left');
    expect(drawing).toContain('do not score');
    expect(drawing).toContain('DRAW');

    expect(formatHoneymoonBridgeState(state())).not.toContain('stock:');
  });

  // 契約は未決定と確定の両側を踏む。
  it('shows the contract and the tricks it needs', () => {
    expect(formatHoneymoonBridgeState(state({ contractLevel: 0, declarerIdx: -1 }))).toContain(
      'contract: not yet decided',
    );

    const out = formatHoneymoonBridgeState(state());
    expect(out).toContain('contract: 2♥');
    expect(out).toContain('needs 8 tricks');
  });

  // **通る最小の宣言を出す。** これが無いと拒否される値を打ち込むことになる。
  it('names the lowest bid that outbids, and says when there is none', () => {
    const open = formatHoneymoonBridgeState(state({ phase: 1, minBidLevel: 3, minBidSuit: 0 }));
    expect(open).toContain('lowest bid that outbids: 3NT');

    const capped = formatHoneymoonBridgeState(state({ phase: 1, minBidLevel: 0, minBidSuit: 0 }));
    expect(capped).toContain('pass is the only move');
  });

  // **ノートランプは NT と書く。** 数字の 0 では読めない。
  it('writes no-trump as NT', () => {
    expect(formatHoneymoonBridgeState(state({ trumpSuit: 0 }))).toContain('contract: 2NT');
  });

  it('shows each seat, marking the declarer and their bid', () => {
    const out = formatHoneymoonBridgeState(
      state({ declarerIdx: 1, players: [seat(0), seat(1, { bidLevel: 2, bidSuit: 3, trickCount: 4, score: 30 })] }),
    );
    expect(out).toMatch(/\[declarer\]: bid 2♥, took 4 \| total 30/);
    expect(out).toContain('no bid');
  });

  it('marks the seat on turn', () => {
    expect(formatHoneymoonBridgeState(state())).toMatch(/^>\S/m);
  });

  it('marks the playable cards in your hand', () => {
    const out = formatHoneymoonBridgeState(state());
    expect(out).toMatch(/\[0\]\S+\*/);
    expect(out).not.toMatch(/\[1\]\S+\*/);
  });

  it('shows the current trick when there is one', () => {
    const out = formatHoneymoonBridgeState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<HoneymoonBridgeResponse>),
    );
    expect(out).toContain('trick:');
  });

  // ディール終了の 3 通りすべて。
  it('reports the deal result', () => {
    expect(formatHoneymoonBridgeState(state({ phase: 3, lastMade: true, lastTricks: 9 }))).toContain(
      'contract made: 9 of 8 tricks',
    );
    expect(formatHoneymoonBridgeState(state({ phase: 3, lastMade: false, lastTricks: 6 }))).toContain(
      'contract down: 6 of 8 tricks',
    );
    // 流局は契約が無いので成立/失敗を書かない。
    const passedOut = formatHoneymoonBridgeState(state({ phase: 3, contractLevel: 0, declarerIdx: -1 }));
    expect(passedOut).not.toContain('contract made');
    expect(passedOut).not.toContain('contract down');
  });

  it('reports the winner, and a tie', () => {
    expect(formatHoneymoonBridgeState(state({ phase: 4, gameEndFlag: true, winnerIdx: 0 }))).toContain(
      'wins on points',
    );
    expect(formatHoneymoonBridgeState(state({ phase: 4, gameEndFlag: true, winnerIdx: -1 }))).toContain(
      'game over — tie',
    );
  });

  it('shows the server message', () => {
    expect(formatHoneymoonBridgeState(state({ message: 'boom' }))).toContain('boom');
  });
});
