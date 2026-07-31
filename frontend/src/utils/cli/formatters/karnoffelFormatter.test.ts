import { describe, expect, it } from 'vitest';
import type { CardDesign, KarnoffelHandResult, KarnoffelPlayer, KarnoffelResponse } from '../../../types/card';
import { formatKarnoffelState } from './karnoffelFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<KarnoffelPlayer>): KarnoffelPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 2,
    cards: isHuman ? [card('SPADE', 13), card('HEART', 11)] : [],
    upCard: card('HEART', 3 + id),
    tricksWon: 0,
    isDealer: id === 3,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function result(overrides?: Partial<KarnoffelHandResult>): KarnoffelHandResult {
  return { winnerTeam: 0, tricks: [3, 1], chosenSuit: 3, ...overrides };
}

function makeState(overrides?: Partial<KarnoffelResponse>): KarnoffelResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: 0,
    handNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    chosenSuit: 3,
    trick: [],
    validPlays: [0, 1],
    trickLeaderIdx: 0,
    trickNumber: 0,
    teamTricks: [1, 0],
    handsWon: [0, 1],
    lastResult: null,
    tricksToWin: 3,
    handSize: 5,
    targetHands: 3,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0, targetHands: 3 },
    ...overrides,
  };
}

describe('formatKarnoffelState', () => {
  it('shows the header and how the suit was chosen', () => {
    const out = formatKarnoffelState(makeState());
    expect(out).toContain('Karnöffel');
    expect(out).toContain('phase: Play');
    expect(out).toContain('first to 3, a hand takes 3 tricks');
    // **切札は表向きの4枚のうち最も低い札が決める。**
    expect(out).toContain('chosen suit: H (the LOWEST face-up card picked it)');
  });

  // **表向きの札は全員ぶん見える。**手札そのものは自分だけ。
  it('shows every face-up card but hides the other hands', () => {
    const out = formatKarnoffelState(makeState());
    expect(out).toContain('up ');
    expect(out).toContain('[0]');
    expect(out.match(/hidden \(2\)/g)).toHaveLength(3);
    expect(out).toContain('[dealer]');
  });

  // **序列が普通と違う。**悪魔の特殊性を出す。
  it('lists the irregular ranking on the human turn', () => {
    const out = formatKarnoffelState(makeState());
    expect(out).toContain('J (Karnöffel) > 7 (devil, ONLY WHEN LED) > 6 (Pope)');
    expect(out).toContain('the 3 loses to kings, the 4 to kings and queens, the 5 to every face card');
    expect(out).toContain('playable: 0 1');
    expect(out).toContain('no need to follow suit, but the devil cannot lead the first trick');
  });

  it('shows the score sheet', () => {
    const out = formatKarnoffelState(makeState());
    expect(out).toContain('team0: 0 hands (1 tricks now)');
    expect(out).toContain('team1: 1 hands (0 tricks now)');
  });

  // **3トリックに届かなければ勝者なし。**
  it('reports the hand result, drawn hands included', () => {
    const won = formatKarnoffelState(makeState({ phase: 1, lastResult: result() }));
    expect(won).toContain('team 0 took the hand 3-1');

    const drawn = formatKarnoffelState(
      makeState({ phase: 1, lastResult: result({ winnerTeam: -1, tricks: [2, 2] }) }),
    );
    expect(drawn).toContain('neither side reached three tricks 2-2');
  });

  it('shows the trick, the message and the winner', () => {
    const out = formatKarnoffelState(
      makeState({
        trick: [card('SPADE', 9)],
        message: 'boom',
        gameEndFlag: true,
        winnerTeam: 0,
      }),
    );
    expect(out).toContain('trick:');
    expect(out).toContain('boom');
    expect(out).toContain('Winning team: 0');
  });

  it('survives an unknown phase, no chosen suit and a missing face-up card', () => {
    expect(formatKarnoffelState(makeState({ phase: 99 }))).toContain('phase: 99');
    expect(formatKarnoffelState(makeState({ chosenSuit: 0 }))).toContain('chosen suit: -');
    const noUp = formatKarnoffelState(
      makeState({ players: [seat(0, true, { upCard: null }), seat(1, false), seat(2, false), seat(3, false)] }),
    );
    expect(noUp).toContain('up -');
  });
});
