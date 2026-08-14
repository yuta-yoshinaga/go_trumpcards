import { describe, expect, it } from 'vitest';
import type { Card, PasurResponse } from '../../../types/card';
import { formatPasurState } from './pasurFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 4), card('SPADE', 8)] : [],
  capturedCount: 0,
  soors: 0,
  score: 0,
  ...over,
});

const state = (over: Partial<PasurResponse> = {}): PasurResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    table: [card('SPADE', 7), card('CLOVER', 13)],
    captureOptions: [[[0]], []],
    deckRemaining: 32,
    packsDealt: 1,
    lastCaptureIdx: -1,
    currentPlayerIdx: 0,
    gameEndFlag: false,
    winners: [],
    config: { playerCnt: 4 },
    message: '',
    ...over,
  }) as unknown as PasurResponse;

describe('formatPasurState', () => {
  it('reports loading for a null state', () => {
    expect(formatPasurState(null)).toBe('Loading...');
  });

  it('shows the pack, deck and phase', () => {
    const out = formatPasurState(state());
    expect(out).toContain('pack 1');
    expect(out).toContain('deck 32 left');
    expect(out).toContain('PLAY');
  });

  // **11 の合計と絵札の扱いが規則そのもの。**
  it('states the capture rule every time', () => {
    const out = formatPasurState(state());
    expect(out).toContain('add to 11');
    expect(out).toContain('same rank only');
  });

  // **場の札には番号を振る。** これが `p <i> <t...>` の引数。
  it('numbers the table, and says when it is empty', () => {
    expect(formatPasurState(state())).toMatch(/table: \[0\]\S+\s+\[1\]/);
    expect(formatPasurState(state({ table: [] }))).toContain('table: empty');
  });

  // **取れる札に印を付ける。** 取れるときは場に置けない。
  it('marks the cards that can capture', () => {
    const out = formatPasurState(state());
    expect(out).toMatch(/\[0\]\S+\*/);
    expect(out).not.toMatch(/\[1\]\S+\*/);
    expect(out).toContain('capturing is compulsory');
  });

  it('shows captures and soors per seat', () => {
    const out = formatPasurState(state({ players: [seat(0, { capturedCount: 6, soors: 2, score: 9 })] }));
    expect(out).toContain('6 taken, 2 soor');
    expect(out).toContain('score 9');
  });

  // **場に残った札の行き先が読めること。**
  it('marks the last capturer only once someone has captured', () => {
    expect(formatPasurState(state())).not.toContain('[last capture]');
    expect(formatPasurState(state({ lastCaptureIdx: 2 }))).toContain('[last capture]');
  });

  it('marks the seat on turn', () => {
    expect(formatPasurState(state())).toMatch(/^>\S/m);
  });

  it('reports a single winner, and a tie', () => {
    expect(formatPasurState(state({ phase: 1, gameEndFlag: true, winners: [0] }))).toContain('wins on points');
    expect(formatPasurState(state({ phase: 1, gameEndFlag: true, winners: [1, 2] }))).toContain('2 players tie');
  });

  it('shows the server message', () => {
    expect(formatPasurState(state({ message: 'boom' }))).toContain('boom');
  });
});
