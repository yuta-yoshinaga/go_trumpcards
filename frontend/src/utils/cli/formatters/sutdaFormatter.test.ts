import { describe, expect, it } from 'vitest';
import { makeSutdaState } from '../../../test/stateFactories';
import { formatSutdaState } from './sutdaFormatter';

describe('formatSutdaState', () => {
  it('prints the hand, phase and pot', () => {
    const out = formatSutdaState(makeSutdaState());
    expect(out).toContain('hand: 1  phase: bet');
    expect(out).toContain('pot: 30  current bet: 10');
  });

  it('lists each seat with its chips and bet', () => {
    const out = formatSutdaState(makeSutdaState());
    expect(out).toContain('chips=990 bet=10');
    expect(out).toContain('(Dealer)');
  });

  // **伏せているうちは自分のぶんだけ。** 相手の役が見えると賭ける意味が無い。
  it('prints only the hands that are visible', () => {
    const out = formatSutdaState(makeSutdaState());
    expect(out).toContain('gwang38');
    expect(out.match(/gwang38/g)).toHaveLength(1);
  });

  it('prints a revealed opponent hand', () => {
    const base = makeSutdaState();
    const out = formatSutdaState(
      makeSutdaState({
        players: base.players.map((p, i) =>
          i === 1 ? { ...p, revealed: true, cards: base.players[0].cards, handName: 'mangtong' } : p,
        ),
      }),
    );
    expect(out).toContain('mangtong');
  });

  it('marks a folded seat', () => {
    const base = makeSutdaState();
    const out = formatSutdaState(
      makeSutdaState({ players: base.players.map((p, i) => (i === 1 ? { ...p, folded: true } : p)) }),
    );
    expect(out).toContain('(folded)');
  });

  it('says whether you owe anything', () => {
    expect(formatSutdaState(makeSutdaState())).toContain('nothing to call - you may check');
    expect(formatSutdaState(makeSutdaState({ callAmount: 20 }))).toContain('to call: 20');
  });

  it('reports how many raises are left', () => {
    expect(formatSutdaState(makeSutdaState({ raiseCount: 1 }))).toContain('raise: 20 per step (2 left)');
    expect(formatSutdaState(makeSutdaState({ canRaise: false }))).not.toContain('raise:');
  });

  it('says nothing about betting when it is not your turn', () => {
    const out = formatSutdaState(makeSutdaState({ isHumanTurn: false }));
    expect(out).not.toContain('to call');
    expect(out).not.toContain('raise:');
  });

  it('prints the showdown result', () => {
    const out = formatSutdaState(
      makeSutdaState({
        lastResult: {
          winners: [0, 1],
          pot: 70,
          handNames: ['gwang38', 'gwang38', 'kkeut5'],
          folded: [false, false, true],
        },
      }),
    );
    expect(out).toContain('showdown: ');
    expect(out).toContain('70');
  });

  // 頼んでいないヒントは出さない。
  it('gates the hint on the request', () => {
    expect(
      formatSutdaState(makeSutdaState({ hintAction: 'raise', hintReason: 'strong_hand', messageCode: '' })),
    ).not.toContain('HINT:');
    expect(
      formatSutdaState(
        makeSutdaState({ hintAction: 'raise', hintReason: 'strong_hand', messageCode: 'sutda.hintRequested' }),
      ),
    ).toContain('HINT: raise (strong_hand)');
  });

  it('announces the winner at the end', () => {
    expect(formatSutdaState(makeSutdaState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('Game Over!');
  });

  it('prints the server message', () => {
    expect(formatSutdaState(makeSutdaState({ message: '行動を選んでください。' }))).toContain('行動を選んでください。');
  });
});
