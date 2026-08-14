import { describe, expect, it } from 'vitest';
import type { Card, SnapResponse } from '../../../types/card';
import { formatSnapState } from './snapFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const state = (over: Partial<SnapResponse> = {}): SnapResponse =>
  ({
    phase: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    currentTurnIdx: 0,
    isHumanTurn: true,
    snapAvailable: false,
    centerPileSize: 3,
    topCard: card('SPADE', 7),
    players: [
      { id: 0, isHuman: true, stockSize: 24 },
      { id: 1, isHuman: false, stockSize: 25 },
    ],
    playerCnt: 2,
    cpuDifficulty: 1,
    pendingKind: 0,
    pendingDeadlineMs: 0,
    lastEventKind: 0,
    lastEventPlayerIdx: 0,
    message: '',
    ...over,
  }) as unknown as SnapResponse;

describe('formatSnapState', () => {
  it('reports loading for a null state', () => {
    expect(formatSnapState(null)).toBe('Loading...');
  });

  it('shows the pile and the phase', () => {
    const out = formatSnapState(state());
    expect(out).toContain('pile 3');
    expect(out).toContain('PLAY');
    expect(out).toContain('top card:');
  });

  // **トリガーが動くことが規則そのもの。**
  it('states the rule every time', () => {
    expect(formatSnapState(state())).toContain('matches the one before it');
  });

  it('says when the pile is empty', () => {
    expect(formatSnapState(state({ centerPileSize: 0, topCard: undefined }))).toContain('pile: empty');
  });

  // **成立しているかは一目で分かる必要がある。**
  it('shouts only while a call would be correct', () => {
    expect(formatSnapState(state())).not.toContain('SNAP IS ON');
    expect(formatSnapState(state({ snapAvailable: true }))).toContain('SNAP IS ON');
  });

  it('shows every stock and marks the seat on turn', () => {
    const out = formatSnapState(state());
    expect(out).toContain('24 in stock');
    expect(out).toContain('25 in stock');
    expect(out).toMatch(/^>\S/m);
  });

  // **直近に何が起きたかを出す。** 盤面だけでは読めない。
  it.each([
    [1, /turned a card over/],
    [2, /called snap and took the pile/],
    [3, /called wrongly and paid a card/],
    [4, /has run out of stock/],
  ])('reports last event kind %s', (kind, expected) => {
    expect(formatSnapState(state({ lastEventKind: kind }))).toMatch(expected);
  });

  it('says nothing about an event when there is none', () => {
    const out = formatSnapState(state({ lastEventKind: 0 }));
    expect(out).not.toMatch(/turned a card over|took the pile|paid a card|run out of stock/);
  });

  it('reports the winner, and a game that could not continue', () => {
    expect(formatSnapState(state({ phase: 1, gameEndFlag: true, winnerIdx: 0 }))).toContain('wins');
    expect(formatSnapState(state({ phase: 1, gameEndFlag: true, winnerIdx: -1 }))).toContain('could not continue');
  });

  it('shows the server message', () => {
    expect(formatSnapState(state({ message: 'boom' }))).toContain('boom');
  });
});
