import { describe, expect, it } from 'vitest';
import type { CardDesign, KlaberjassPlayer, KlaberjassResponse } from '../../../types/card';
import { formatKlaberjassState } from './klaberjassFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<KlaberjassPlayer>): KlaberjassPlayer {
  return {
    id,
    isHuman,
    cardCount: 2,
    cards: isHuman ? [card('SPADE', 11), card('HEART', 1)] : [],
    sequences: [],
    handPoints: 0,
    score: 0,
    isMaker: false,
    isDealer: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<KlaberjassResponse>): KlaberjassResponse {
  return {
    players: [seat(0, true, { isDealer: true }), seat(1, false, { isMaker: true })],
    phase: 3,
    dealNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 1,
    dealerIdx: 0,
    trumpSuit: 1,
    turnUpCard: null,
    makerIdx: 1,
    trick: [],
    trickLeaderIdx: 0,
    trickNumber: 0,
    validPlays: [0, 1],
    sequenceWinner: -1,
    lastTrickWinner: -1,
    lastTrickBonus: 10,
    belaHolder: -1,
    belaScored: false,
    dixUsed: false,
    bete: false,
    schmeissBy: -1,
    targetScore: 501,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatKlaberjassState', () => {
  it('shows the deal, trump and target', () => {
    const out = formatKlaberjassState(makeState());
    expect(out).toContain('deal: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('trump: S');
    expect(out).toContain('target: 501');
    expect(out).toContain('[dealer]');
    expect(out).toContain('[maker]');
  });

  // **出せる札を出さないと操作できない。**3つの強制ルールがあるので手札だけでは判らない。
  it('lists the playable indexes on the human turn', () => {
    const out = formatKlaberjassState(makeState());
    expect(out).toContain('playable: 0 1');
    expect(out).toContain('your turn');
  });

  it('hides the opponent hand but shows its size', () => {
    const out = formatKlaberjassState(makeState());
    expect(out).toContain('hidden (2)');
  });

  // 表向きカードはビッド中だけ意味を持つ。
  it('shows the turn-up only before trump is fixed', () => {
    const bidding = formatKlaberjassState(
      makeState({ phase: 0, bidPlayerIdx: 0, trumpSuit: 0, turnUpCard: card('HEART', 13) }),
    );
    expect(bidding).toContain('turn-up:');
    expect(bidding).toContain('your bid');

    const playing = formatKlaberjassState(makeState({ turnUpCard: card('HEART', 13) }));
    expect(playing).not.toContain('turn-up:');
  });

  it('prompts for each bidding round and the schmeiss', () => {
    expect(formatKlaberjassState(makeState({ phase: 1, bidPlayerIdx: 0 }))).toContain('c <1-4>');
    expect(formatKlaberjassState(makeState({ phase: 2, bidPlayerIdx: 0 }))).toContain('make them the maker');
  });

  // ベートは通常の精算と字面で見分けられる必要がある。
  it('tells a bete apart from an ordinary settlement', () => {
    expect(formatKlaberjassState(makeState({ phase: 4, bete: true }))).toContain('bete');
    expect(formatKlaberjassState(makeState({ phase: 4, bete: false }))).toContain('hand over');
  });

  it('shows the trick and announces the winner', () => {
    const out = formatKlaberjassState(makeState({ trick: [card('HEART', 10)] }));
    expect(out).toContain('trick:');

    const end = formatKlaberjassState(makeState({ phase: 5, gameEndFlag: true, winnerIdx: 1, message: 'done' }));
    expect(end).toContain('Game Over!');
    expect(end).toContain('done');
  });
});
