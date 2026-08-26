import { describe, expect, it } from 'vitest';
import { makeQuodlibetState } from '../../../test/stateFactories';
import { formatQuodlibetState } from './quodlibetFormatter';

const playState = makeQuodlibetState({
  phase: 'play',
  isContractPhase: false,
  currentContract: 1,
  currentContractName: 'minus',
  currentPlayerIdx: 0,
  playableIndices: [0, 1, 2, 3],
});

describe('formatQuodlibetState', () => {
  it('prints the deal, wheel and phase', () => {
    const out = formatQuodlibetState(makeQuodlibetState());
    expect(out).toContain('deal: 1/12  wheel: 1  phase: selectContract');
  });

  it('prints the contract and trick once a deal is under way', () => {
    const out = formatQuodlibetState(playState);
    expect(out).toContain('contract: minus');
    expect(out).toContain('trick: 1/8');
  });

  // **点は罰点。** "score" と書くと多いほうが良いように読める。
  it('labels the running total as penalty', () => {
    const out = formatQuodlibetState(
      makeQuodlibetState({ players: makeQuodlibetState().players.map((p) => ({ ...p, penalty: 40 })) }),
    );
    expect(out).toContain('penalty=40');
    expect(out).not.toContain('score=');
  });

  it('indexes the human hand', () => {
    expect(formatQuodlibetState(playState)).toContain('[0]');
  });

  // **選べるのはこの輪の残りだけ。**
  it('lists the choices while the dealer is choosing', () => {
    const out = formatQuodlibetState(makeQuodlibetState());
    expect(out).toContain('choose: [0]plus  [1]minus  [2]badNeighbour  [3]alarich');
  });

  it('does not list choices once the contract is settled', () => {
    expect(formatQuodlibetState(playState)).not.toContain('choose:');
  });

  it('prints the cards already in the trick', () => {
    const base = makeQuodlibetState();
    const out = formatQuodlibetState(
      makeQuodlibetState({ ...playState, currentTrick: [{ playerIdx: 1, card: base.players[0].cards[0] }] }),
    );
    expect(out).toContain('trick: CPU 1=');
  });

  // **四分と小食いはトリックではない。** 場と重ねを出す。
  it('prints the stack and the table for shedding contracts', () => {
    const out = formatQuodlibetState(
      makeQuodlibetState({
        ...playState,
        currentContract: 10,
        currentContractName: 'quadrature',
        isShedding: true,
        stack: [{ design: 'SPADE' as const, value: 7, color: 'black' }],
        tablePlaced: [[4], [], [], []],
      }),
    );
    expect(out).toContain('stack: ');
    expect(out).toContain('table: s1=[4]');
    // シェディングではトリック番号を出さない。
    expect(out).not.toContain('trick: 1/8');
  });

  it('says when nothing can be played', () => {
    expect(formatQuodlibetState(makeQuodlibetState({ ...playState, isShedding: true, canPass: true }))).toContain(
      'nothing playable: use pass',
    );
    expect(formatQuodlibetState(playState)).not.toContain('nothing playable');
  });

  it('prints the last deal breakdown', () => {
    const out = formatQuodlibetState(
      makeQuodlibetState({
        lastDeal: { contract: 1, contractName: 'minus', round: 0, dealerIdx: 0, points: [30, 0, 20, 30] },
      }),
    );
    expect(out).toContain('last deal (minus): 30 / 0 / 20 / 30');
  });

  // 頼んでいないヒントは出さない。
  it('gates the hint on the request', () => {
    const hint = { cardIndices: [2], reason: 'avoid_penalty' };
    expect(formatQuodlibetState(makeQuodlibetState({ ...playState, hint, messageCode: '' }))).not.toContain('HINT:');
    expect(
      formatQuodlibetState(makeQuodlibetState({ ...playState, hint, messageCode: 'quodlibet.hintRequested' })),
    ).toContain('HINT: card indices [2] (avoid_penalty)');
  });

  it('gates the contract hint on the request too', () => {
    expect(
      formatQuodlibetState(makeQuodlibetState({ hintContract: 2, messageCode: 'quodlibet.hintRequested' })),
    ).toContain('HINT: contract [2]');
    expect(formatQuodlibetState(makeQuodlibetState({ hintContract: 2, messageCode: '' }))).not.toContain('HINT:');
  });

  // **勝つのは罰点が最少の人。**
  it('names the seats on the fewest penalty points at the end', () => {
    const out = formatQuodlibetState(makeQuodlibetState({ gameEndFlag: true, winners: [0, 2] }));
    expect(out).toContain('Game Over! Fewest penalty: ');
  });

  it('prints the server message', () => {
    expect(formatQuodlibetState(makeQuodlibetState({ message: '札を出してください。' }))).toContain(
      '札を出してください。',
    );
  });
});
