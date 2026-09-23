import { describe, expect, it } from 'vitest';
import type { BinokelPlayerData, BinokelResponse } from '../../../types/card';
import { formatBinokelState } from './binokelFormatter';

function makePlayer(overrides?: Partial<BinokelPlayerData>): BinokelPlayerData {
  return {
    id: 0,
    isHuman: true,
    cardCount: 1,
    cards: [{ design: 'SPADE', value: 1 }],
    score: 120,
    trickCount: 2,
    bid: 0,
    hasPassed: false,
    meldScore: 40,
    trickPoints: 12,
    ...overrides,
  };
}

function makeState(overrides?: Partial<BinokelResponse>): BinokelResponse {
  return {
    players: [
      makePlayer({ id: 0, isHuman: true, score: 120 }),
      makePlayer({ id: 1, isHuman: false, cards: [], score: 80 }),
      makePlayer({ id: 2, isHuman: false, cards: [], score: 50 }),
    ],
    phase: 0,
    roundNumber: 2,
    trickNumber: 5,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 2,
    trumpSuit: 1,
    highestBid: 250,
    highestBidder: 0,
    currentTrick: [],
    scores: [120, 80, 50],
    gameEndFlag: false,
    winnerPlayer: -1,
    leadPlayerIdx: 0,
    playerMelds: [[], [], []],
    dabb: [],
    dabbDiscarded: [],
    config: { cpuDifficulty: 1, pointLimit: 1500 },
    message: '',
    ...overrides,
  };
}

describe('formatBinokelState', () => {
  it('renders the header, round, trick and player individual scores', () => {
    const out = formatBinokelState(makeState());
    expect(out).toContain('Binokel');
    expect(out).toContain('round: 2');
    expect(out).toContain('trick: 5');
    expect(out).toContain('scores: あなた=120  CPU 1=80  CPU 2=50');
    expect(out).not.toContain('Team');
  });

  // **切り札も宣言額も競りが終わるまで無い。**どちらも行ごと出ない。
  it('omits the trump and the bid until the auction settles them', () => {
    const out = formatBinokelState(makeState({ trumpSuit: 0, highestBid: 0, highestBidder: -1 }));
    expect(out).not.toContain('trump:');
    expect(out).not.toContain('highest bid:');
    expect(formatBinokelState(makeState())).toContain('highest bid: 250 (あなた)');
  });

  it('renders declarer information with trump and highest bid', () => {
    const out = formatBinokelState(makeState());
    expect(out).toContain('trump: Spade (declarer: あなた)');
    expect(out).toContain('highest bid: 250 (あなた)');
  });

  it("renders each player's meld and trick points and declarer tag", () => {
    const out = formatBinokelState(makeState());
    expect(out).toContain('meld=40');
    expect(out).toContain('tricks=2T/12pts');
    expect(out).toContain('[Declarer]');
  });

  it('differentiates unbid, passed, and declared bids', () => {
    const state = makeState({
      highestBidder: -1,
      players: [
        makePlayer({ id: 0, isHuman: true, bid: 0, hasPassed: false }),
        makePlayer({ id: 1, isHuman: false, cards: [], bid: 0, hasPassed: true }),
        makePlayer({ id: 2, isHuman: false, cards: [], bid: 160, hasPassed: false }),
      ],
    });
    const out = formatBinokelState(state);
    expect(out).toContain('あなた: score=120 bid=-');
    expect(out).toContain('CPU 1: score=120 bid=[Passed]');
    expect(out).toContain('CPU 2: score=120 bid=160');
  });

  it('announces the winner once the game ends', () => {
    const out = formatBinokelState(makeState({ gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: CPU 1');
    expect(out).not.toContain('Team');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 0, reason: 'follow_suit' };
    expect(formatBinokelState(makeState({ hint, messageCode: 'binokel.hintRequested' }))).toContain('HINT');
    expect(formatBinokelState(makeState({ hint, messageCode: 'binokel.playing' }))).not.toContain('HINT');
  });

  it('renders revealed Dabb cards if present', () => {
    const out = formatBinokelState(
      makeState({
        dabb: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 10 },
          { design: 'DIAMOND', value: 12 },
        ],
      }),
    );
    expect(out).toContain('Dabb: ♠A, ♥10, ♦Q');
  });

  it('renders current trick cards', () => {
    const out = formatBinokelState(
      makeState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'SPADE', value: 1 } },
          { playerIdx: 1, card: { design: 'HEART', value: 10 } },
        ],
      }),
    );
    expect(out).toContain('trick: あなた=♠A, CPU 1=♥10');
  });

  it('renders player melds if present', () => {
    const out = formatBinokelState(
      makeState({
        playerMelds: [
          [
            {
              type: 1,
              points: 20,
              cards: [
                { design: 'HEART', value: 13 },
                { design: 'HEART', value: 12 },
              ],
            },
          ],
          [],
          [],
        ],
      }),
    );
    expect(out).toContain('あなた melds: 20pts(♥K, ♥Q)');
  });

  it('renders trump and highest bid without declarer if highestBidder is -1, and handles fallback suit', () => {
    const out = formatBinokelState(
      makeState({
        trumpSuit: 99,
        highestBid: 150,
        highestBidder: -1,
      }),
    );
    expect(out).toContain('trump: ?');
    expect(out).toContain('highest bid: 150');
    expect(out).not.toContain('declarer:');
  });

  it('falls back to p.score when state.scores is undefined', () => {
    const state = makeState({ scores: undefined as unknown as number[] });
    const out = formatBinokelState(state);
    expect(out).toContain('scores: あなた=120  CPU 1=80  CPU 2=50');
  });

  it('renders message if present', () => {
    const out = formatBinokelState(makeState({ message: 'Round started' }));
    expect(out).toContain('Round started');
  });

  it('announces human winner at game end', () => {
    const out = formatBinokelState(makeState({ gameEndFlag: true, winnerPlayer: 0 }));
    expect(out).toContain('Game Over! Winner: あなた');
  });
});
