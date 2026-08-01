import { describe, expect, it } from 'vitest';
import type { CardDesign, LiteraturePlayer, LiteratureResponse } from '../../../types/card';
import { formatLiteratureState } from './literatureFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function halfSuitCards(half: number) {
  const suits: CardDesign[] = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'];
  const design = suits[Math.floor(half / 2)] ?? 'SPADE';
  const ranks = half % 2 === 0 ? [2, 3, 4, 5, 6, 7] : [9, 10, 11, 12, 13, 1];
  return ranks.map((v) => card(design, v));
}

function seat(id: number, isHuman: boolean, overrides?: Partial<LiteraturePlayer>): LiteraturePlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 8,
    cards: isHuman ? [card('SPADE', 2)] : [],
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<LiteratureResponse>): LiteratureResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false), seat(4, false), seat(5, false)],
    phase: 0,
    currentPlayerIdx: 0,
    halfSuits: [0, 0, 0, 0, 0, 0, 0, 0],
    halfSuitCards: Array.from({ length: 8 }, (_, h) => halfSuitCards(h)),
    asks: [],
    claims: [],
    lastAsk: null,
    lastClaim: null,
    teamHalfSuits: [0, 0],
    cancelledCount: 0,
    openCount: 8,
    winThreshold: 5,
    halfSuitCnt: 8,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

describe('formatLiteratureState', () => {
  // **勝利には 5 組。**4 組では決着しない。
  it('states the real threshold', () => {
    const out = formatLiteratureState(makeState());
    expect(out).toContain('Literature');
    expect(out).toContain('winning takes 5 of 8 — a MAJORITY, so four decides nothing');
  });

  // **無効は別枠。**合計が 8 になるとは限らない。
  it('counts cancelled half-suits separately', () => {
    const out = formatLiteratureState(
      makeState({ teamHalfSuits: [3, 2], cancelledCount: 1, openCount: 2, halfSuits: [1, 1, 1, 2, 2, 3, 0, 0] }),
    );
    expect(out).toContain('team0: 3 | team1: 2 | cancelled: 1 | open: 2');
    expect(out).toContain('[5] H high (9-A): CANCELLED');
    expect(out).toContain('[0] S low (2-7): team 0');
    expect(out).toContain('[6] D low (2-7): open');
  });

  // **終局まで味方の手札も見えない。**
  it('hides every hand but the human own', () => {
    const out = formatLiteratureState(makeState());
    expect(out.match(/hidden \(8\)/g)).toHaveLength(5);
    expect(out).toContain('<- turn');
  });

  // **要求の履歴は公開情報。**
  it('shows the ask history with hits and misses', () => {
    const out = formatLiteratureState(
      makeState({
        asks: [
          { from: 0, to: 1, card: card('SPADE', 3), success: true },
          { from: 1, to: 0, card: card('HEART', 9), success: false },
          { from: 2, to: 3, card: null, success: false },
        ],
      }),
    );
    expect(out).toContain('everyone sees these');
    expect(out).toContain('... hit');
    expect(out).toContain('... miss');
    // card が無くても落ちない。
    expect(out).toContain('seat 2 -> seat 3: ?');
  });

  // **宣言の結末は 3 通り。**無効は「相手に渡る」とは違う。
  it('distinguishes the three claim outcomes', () => {
    const won = formatLiteratureState(makeState({ lastClaim: { player: 0, halfSuit: 0, outcome: 0, awardedTeam: 0 } }));
    expect(won).toContain('went to team 0');

    const cancelled = formatLiteratureState(
      makeState({ lastClaim: { player: 0, halfSuit: 0, outcome: 1, awardedTeam: -1 } }),
    );
    expect(cancelled).toContain('was CANCELLED');
    expect(cancelled).toContain('does NOT go to the opponents');

    const lost = formatLiteratureState(
      makeState({ lastClaim: { player: 0, halfSuit: 0, outcome: 2, awardedTeam: 1 } }),
    );
    expect(lost).toContain('went to the opponents, team 1');
  });

  // **要求の 3 条件を手番のときに書く。**
  it('states the ask conditions on the human turn', () => {
    const out = formatLiteratureState(makeState());
    expect(out).toContain('ONLY an opponent');
    expect(out).toContain('only for a half-suit you hold');
    expect(out).toContain('only for a card you do NOT hold');
  });

  it('shows the message and the outcome, a level finish included', () => {
    const won = formatLiteratureState(makeState({ message: 'boom', gameEndFlag: true, winnerTeam: 0 }));
    expect(won).toContain('boom');
    expect(won).toContain('Winning team: 0');

    const drawn = formatLiteratureState(makeState({ gameEndFlag: true, winnerTeam: -1 }));
    expect(drawn).toContain('It ended level');
  });

  it('survives an unknown phase', () => {
    expect(formatLiteratureState(makeState({ phase: 99 }))).toContain('phase: 99');
  });
});
