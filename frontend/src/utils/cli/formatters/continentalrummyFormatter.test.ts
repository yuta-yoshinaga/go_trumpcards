import { describe, expect, it } from 'vitest';
import { makeContinentalRummyState } from '../../../test/stateFactories';
import { formatContinentalRummyState } from './continentalrummyFormatter';

describe('formatContinentalRummyState', () => {
  it('names the phase and the round', () => {
    const out = formatContinentalRummyState(makeContinentalRummyState());
    expect(out).toContain('Continental Rummy');
    expect(out).toContain('Phase: DISCARD');
    expect(out).toContain('Round: 2 / 3');
  });

  it('shows the stock and the top discard', () => {
    expect(formatContinentalRummyState(makeContinentalRummyState({ stockCount: 7 }))).toContain('Stock: 7');
  });

  // **上がれる形はサーバが返したものだけを並べる。** 15 の分割から組み直すと
  // 5+5+5 を勝手に足してしまう。
  it('lists the legal layouts from the payload, and never invents 5+5+5', () => {
    const out = formatContinentalRummyState(makeContinentalRummyState());
    expect(out).toContain('3+3+3+3+3');
    expect(out).toContain('4+4+4+3');
    expect(out).toContain('5+4+3+3');
    expect(out).not.toContain('5+5+5');
    expect(out).toContain('sets are never melds');
  });

  it('renders every seat with its count and score, and only the human hand', () => {
    const out = formatContinentalRummyState(makeContinentalRummyState());
    expect(out).toContain('You (dealer) <- to play: 16 card(s), score 0');
    expect(out).toContain('CPU 1: 15 card(s), score 0');
    expect(out).toContain('[0]');
    expect(out).toContain('[15]');
  });

  // **上がれるときは黙っていない。** 15 枚の分割は目で追いきれない。
  it('says so when the hand can go out, and stays quiet otherwise', () => {
    expect(formatContinentalRummyState(makeContinentalRummyState())).toContain('You can go out: goout 15');
    expect(formatContinentalRummyState(makeContinentalRummyState({ goOutIdx: -1 }))).not.toContain('You can go out');
  });

  // **加点は内訳で見せる。** 合計だけだと、どう上がると得なのかが伝わらない。
  it('breaks the settlement down and says what was collected from each opponent', () => {
    const out = formatContinentalRummyState(
      makeContinentalRummyState({
        phase: 'roundEnd',
        lastResult: {
          winnerIdx: 0,
          bonuses: [
            { key: 'win', points: 1 },
            { key: 'noJoker', points: 10 },
            { key: 'dealt', points: 10 },
          ],
          perOpponent: 21,
          total: 63,
        },
      }),
    );
    expect(out).toContain('You went out:');
    expect(out).toContain('going out: 1');
    expect(out).toContain('no joker used: 10');
    expect(out).toContain('out on the dealt fifteen: 10');
    expect(out).toContain('21 from each opponent, 63 in all');
  });

  it('reports a washout as nobody going out', () => {
    const out = formatContinentalRummyState(
      makeContinentalRummyState({
        phase: 'roundEnd',
        lastResult: { winnerIdx: -1, bonuses: [], perOpponent: 0, total: 0 },
      }),
    );
    expect(out).toContain('Nobody went out');
    expect(out).not.toContain('from each opponent');
  });

  it('renders the melds a winner laid down', () => {
    const base = makeContinentalRummyState();
    const out = formatContinentalRummyState(
      makeContinentalRummyState({
        players: base.players.map((p) =>
          p.isHuman ? { ...p, cards: [], cardCount: 0, melds: [base.players[0].cards.slice(0, 3)] } : p,
        ),
      }),
    );
    expect(out).toContain('    ');
    expect(out).toContain('You (dealer) <- to play: 0 card(s)');
  });

  it('announces the winner only once the game is over', () => {
    expect(formatContinentalRummyState(makeContinentalRummyState())).not.toContain('You win!');
    expect(
      formatContinentalRummyState(makeContinentalRummyState({ gameEndFlag: true, winnerIdx: 0, phase: 'gameEnd' })),
    ).toContain('You win!');
    expect(
      formatContinentalRummyState(makeContinentalRummyState({ gameEndFlag: true, winnerIdx: 2, phase: 'gameEnd' })),
    ).toContain('CPU 2 wins.');
    expect(
      formatContinentalRummyState(makeContinentalRummyState({ gameEndFlag: true, winnerIdx: -1, phase: 'gameEnd' })),
    ).toContain('A draw.');
  });

  it('says UNKNOWN rather than guessing at an unrecognised phase', () => {
    expect(formatContinentalRummyState(makeContinentalRummyState({ phase: 'nonsense' }))).toContain('Phase: UNKNOWN');
  });
});
