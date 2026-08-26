import { describe, expect, it } from 'vitest';
import { makeGleekState } from '../../../test/stateFactories';
import { formatGleekState } from './gleekFormatter';

describe('formatGleekState', () => {
  it('renders the header, round/trick, trump and per-player scores', () => {
    const out = formatGleekState(makeGleekState({ playerScores: [3, 1, -4] }));
    expect(out).toContain('Gleek');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: heart');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P2=-4');
  });

  // **フェーズ名は phase 値で引く。** Discard を並びから落とすと、以降が
  // 1 つずつずれて Play が「Discard」と表示される。
  it('names every phase by its own value', () => {
    expect(formatGleekState(makeGleekState({ phase: 1 }))).toContain('phase: Discard');
    expect(formatGleekState(makeGleekState({ phase: 2 }))).toContain('phase: Play');
    expect(formatGleekState(makeGleekState({ phase: 3 }))).toContain('phase: TrickEnd');
  });

  it('shows the auction while the stock is unsold, and the buyer once it is', () => {
    const unsold = formatGleekState(makeGleekState({ buyerIdx: -1, highestBid: 12, nextBidAmount: 14 }));
    expect(unsold).toContain('stock: unsold');
    expect(unsold).toContain('highest 12');
    const sold = formatGleekState(makeGleekState({ buyerIdx: 0, winningBid: 14 }));
    expect(sold).toContain('bought by P0 for 14');
  });

  // **段階の点は出さないと見えない。** ラフとメルドで動いた点が画面に無いと、
  // 累積点だけが理由なく動いているように見える。
  it('renders the ruff and both meld kinds', () => {
    const out = formatGleekState(
      makeGleekState({
        ruffWinnerIdx: 0,
        melds: [
          { playerIdx: 0, rank: 13, count: 3, value: 3 },
          { playerIdx: 1, rank: 11, count: 4, value: 2 },
        ],
      }),
    );
    expect(out).toContain('ruff: P0 with 31 in heart');
    expect(out).toContain('gleek: P0 shows 3 kings for 3 each');
    expect(out).toContain('mournival: P1 shows 4 jacks for 2 each');
  });

  it('omits the ruff line before it is scored', () => {
    expect(formatGleekState(makeGleekState({ ruffWinnerIdx: -1 }))).not.toContain('ruff:');
  });

  it('renders the human hand with indices', () => {
    expect(formatGleekState(makeGleekState())).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatGleekState(
      makeGleekState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the deal total and par at the end of a deal', () => {
    const out = formatGleekState(makeGleekState({ phase: 4, dealPoints: 78, par: 26 }));
    expect(out).toContain('deal points: 78');
    expect(out).toContain('par: 26');
  });

  it('renders a hint with card indices only when it was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatGleekState(makeGleekState({ hint, messageCode: 'gleek.hintRequested' }))).toContain(
      'HINT: card indices [1, 2]',
    );
    expect(formatGleekState(makeGleekState({ hint, messageCode: 'gleek.playPhase.lead' }))).not.toContain('HINT');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatGleekState(makeGleekState({ phase: 5, gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: Player 1');
  });

  it('renders an explicit message when present', () => {
    expect(formatGleekState(makeGleekState({ message: 'hello world' }))).toContain('hello world');
  });

  // **切り札は配った時点で決まるが、盤が壊れて届かないこともある。** その場合に
  // 未定義を出さず、ダッシュに落とす。
  it('falls back to a dash for an unset trump or a missing turn-up', () => {
    const out = formatGleekState(makeGleekState({ trumpSuit: -1, turnUp: null }));
    expect(out).toContain('trump: -');
    expect(out).toContain('turn-up: -');
    expect(out).not.toContain('undefined');
  });

  it('falls back for a ruff winner with no recorded suit', () => {
    const out = formatGleekState(
      makeGleekState({
        ruffWinnerIdx: 1,
        players: makeGleekState().players.map((p) => (p.id === 1 ? { ...p, ruff: 0, ruffSuit: -1 } : p)),
      }),
    );
    expect(out).toContain('ruff: P1 with 0 in -');
  });

  it('prints an unknown meld rank as its number rather than undefined', () => {
    const out = formatGleekState(makeGleekState({ melds: [{ playerIdx: 0, rank: 9, count: 3, value: 1 }] }));
    expect(out).toContain('shows 3 9 for 1 each');
    expect(out).not.toContain('undefined');
  });

  it('labels a trick card played by a seat that is not in the players list', () => {
    const out = formatGleekState(
      makeGleekState({ currentTrick: [{ playerIdx: 9, card: { design: 'HEART', value: 12 } }] }),
    );
    expect(out).toContain('trick:');
    expect(out).not.toContain('undefined');
  });

  it('renders a requested hint that carries no card indices', () => {
    const out = formatGleekState(
      makeGleekState({
        hint: { cardIndices: undefined as unknown as number[], reason: 'bid_raise' },
        messageCode: 'gleek.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices []');
    expect(out).toContain('bid_raise');
  });

  it('omits the hand line for a human seat holding no cards', () => {
    const out = formatGleekState(
      makeGleekState({ players: makeGleekState().players.map((p) => (p.isHuman ? { ...p, cards: [] } : p)) }),
    );
    expect(out).not.toMatch(/\[0\]/);
  });
});
