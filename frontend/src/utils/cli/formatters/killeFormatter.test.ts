import { describe, expect, it } from 'vitest';
import type { KillePlayer, KilleResponse } from '../../../types/card';
import { formatKilleState } from './killeFormatter';

const card = (label: string) => ({
  design: 'JOKER' as const,
  value: 8,
  glyph: label,
  label,
  color: 'black',
  deck: 'kille',
});

function seat(id: number, isHuman: boolean, overrides?: Partial<KillePlayer>): KillePlayer {
  return {
    id,
    isHuman,
    card: isHuman ? card('5') : null,
    strength: isHuman ? 8 : 0,
    chips: 0,
    reentries: 0,
    reentryCost: 1,
    canReenter: true,
    isOut: false,
    knockedBy: '',
    isSatisfied: false,
    isFinished: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<KilleResponse>): KilleResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    stockCount: 38,
    pot: 4,
    events: [],
    loserIdxs: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatKilleState', () => {
  it('shows the round, the pot and the dealer', () => {
    const out = formatKilleState(makeState());
    expect(out).toContain('round: 1');
    expect(out).toContain('pot: 4');
    expect(out).toContain('stock: 38');
    expect(out).toContain('[dealer]');
    expect(out).toContain('phase: Exchange');
  });

  // 伏せた札を勝手に埋めない。
  it('marks a face-down seat as hidden', () => {
    const out = formatKilleState(makeState());
    expect(out).toContain('[hidden]');
  });

  it('names why each seat went out', () => {
    const out = formatKilleState(
      makeState({
        phase: 1,
        players: [
          seat(0, true, { isOut: true, knockedBy: 'hussar' }),
          seat(1, false, { isOut: true, knockedBy: 'pig' }),
          seat(2, false, { isOut: true, knockedBy: '' }),
          seat(3, false, { isSatisfied: true }),
        ],
      }),
    );
    expect(out).toContain('(out: Hussar)');
    expect(out).toContain('(out: Pig)');
    expect(out).toContain('(out: lowest)');
    expect(out).toContain('[stands pat]');
    expect(out).toContain('showdown');
  });

  // 山札との交換は相手席が -1 で来る。席番号として出してはいけない。
  it('renders a stock swap as the stock, not a seat', () => {
    const out = formatKilleState(makeState({ events: [{ kind: 'stock', actor: 3, target: -1 }] }));
    expect(out).toContain('the stock');
    expect(out).not.toContain('-> -1');
  });

  it('prompts on the human turn and announces the winner', () => {
    expect(formatKilleState(makeState())).toContain('your turn');
    const out = formatKilleState(makeState({ phase: 2, gameEndFlag: true, winnerIdx: 2, message: 'done' }));
    expect(out).toContain('Game Over!');
    expect(out).toContain('done');
  });

  it('marks an eliminated seat', () => {
    const out = formatKilleState(makeState({ players: [seat(0, true, { isFinished: true, isOut: true })] }));
    expect(out).toContain('(eliminated)');
  });
});
