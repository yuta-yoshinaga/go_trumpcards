import { describe, expect, it } from 'vitest';
import type { BanLuckResponse, Card } from '../../../types/card';
import { BAN_LUCK_RANK } from '../../../types/games/banluck';
import { BanLuckPhase } from '../../../types/phases';
import { formatBanLuckState } from './banluckFormatter';

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<BanLuckResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 50,
    cards: [card(10), card(7)],
    score: 17,
    rank: BAN_LUCK_RANK.point,
    outcome: 1,
    roundBet: 50,
    delta: 0,
    busted: false,
    stood: false,
    isBanker: false,
    isTurn: false,
    ...over,
  }) as BanLuckResponse['seats'][number];

const base = {
  phase: BanLuckPhase.PLAY,
  seats: [seat({ name: 'YOU', isBanker: true }), seat({ name: 'CPU1', isHuman: false })],
  bankerSeat: 0,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  mustHit: false,
  roundNumber: 3,
  remainingCards: 40,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
} as unknown as BanLuckResponse;

const at = (over: Partial<BanLuckResponse>) => ({ ...base, ...over }) as BanLuckResponse;

describe('formatBanLuckState', () => {
  it('フェーズとラウンドを出す', () => {
    const out = formatBanLuckState(base);
    expect(out).toContain('PLAY');
    expect(out).toContain('Round: 3');
  });

  // **親が誰かは常に見えていないといけない。** 役割が回るゲームなので。
  it('親を名指しし、席にも印を付ける', () => {
    const out = formatBanLuckState(base);
    expect(out).toContain('Banker: YOU');
    expect(out).toContain('YOU (banker)');
    expect(out).not.toContain('CPU1 (banker)');
  });

  it('操作中の席に印を付ける', () => {
    const out = formatBanLuckState(at({ turnSeat: 1 }));
    const lines = out.split('\n').filter((l) => l.includes('chips'));
    expect(lines[0]?.startsWith(' ')).toBe(true);
    expect(lines[1]?.startsWith('*')).toBe(true);
  });

  // **義務は名指しする。** 出す / 出さないの両側を踏む。
  it('親の義務を出す', () => {
    expect(formatBanLuckState(at({ mustHit: true }))).toContain('cannot stand below 15');
    expect(formatBanLuckState(base)).not.toContain('cannot stand below 15');
  });

  it('決着で役と収支を出す', () => {
    const out = formatBanLuckState(
      at({
        phase: BanLuckPhase.ROUND_END,
        seats: [
          seat({ name: 'YOU', rank: BAN_LUCK_RANK.banBan, delta: 150 }),
          seat({ name: 'CPU1', rank: BAN_LUCK_RANK.fiveDragon, delta: -150 }),
        ],
      }),
    );
    expect(out).toContain('YOU: ban ban -> 150');
    expect(out).toContain('CPU1: five dragon -> -150');
  });

  it('終局で勝者を出す', () => {
    const out = formatBanLuckState(at({ phase: BanLuckPhase.GAME_END, gameEndFlag: true, winnerSeat: 1 }));
    expect(out).toContain('Winner: CPU1');
  });

  it('配る前は点数を出さない', () => {
    const out = formatBanLuckState(at({ phase: BanLuckPhase.BET, seats: [seat({ cards: [], score: 0, bet: 0 })] }));
    expect(out).not.toContain('= 0');
  });
});
