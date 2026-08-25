import { describe, expect, it } from 'vitest';
import { makeCostlyColoursState } from '../../../test/stateFactories';
import { formatCostlyColoursState } from './costlycoloursFormatter';

describe('formatCostlyColoursState', () => {
  it('shows the deal, the turn-up and the seats', () => {
    const out = formatCostlyColoursState(makeCostlyColoursState());
    expect(out).toContain('Costly Colours');
    expect(out).toContain('deal: 1');
    expect(out).toContain('target: 61');
    // **表の 1 枚は常に見せる。** ショーの色役も J / 2 の 4 点もこれ次第。
    expect(out).toMatch(/turn-up: .+/);
    expect(out).not.toContain('turn-up: (none)');
    expect(out).toContain('score=0');
  });

  it('says when there is no turn-up', () => {
    expect(formatCostlyColoursState(makeCostlyColoursState({ turnUp: null }))).toContain('turn-up: (none)');
  });

  it('shows the running count', () => {
    expect(formatCostlyColoursState(makeCostlyColoursState())).toContain('count: (nothing yet)  total 0');
    const out = formatCostlyColoursState(
      makeCostlyColoursState({
        phase: 'play',
        pile: [{ design: 'SPADE', value: 7, color: 'black' }],
        total: 7,
      }),
    );
    expect(out).toContain('total 7');
  });

  // **断ると相手に 1 点。** 選ぶ前にそう書く。
  it('spells out the cost of refusing the mog', () => {
    const out = formatCostlyColoursState(makeCostlyColoursState());
    expect(out).toContain('refusing pegs 1 for your opponent');
  });

  // **出せる札が無いなら「ゴー」。** 探させない。
  it('lists the playable indices, or says nothing fits', () => {
    expect(formatCostlyColoursState(makeCostlyColoursState({ phase: 'play', playableIdxs: [0, 2] }))).toContain(
      'playable: 0 2',
    );
    expect(formatCostlyColoursState(makeCostlyColoursState({ phase: 'play', playableIdxs: [] }))).toContain(
      'nothing fits under 31',
    );
  });

  it('omits the playable line when it is not the human turn', () => {
    const out = formatCostlyColoursState(
      makeCostlyColoursState({ phase: 'play', isHumanTurn: false, currentPlayerIdx: 1 }),
    );
    expect(out).not.toContain('playable:');
  });

  it('shows the played cards once they are out', () => {
    const out = formatCostlyColoursState(
      makeCostlyColoursState({
        phase: 'play',
        players: [
          {
            id: 0,
            isHuman: true,
            cards: [],
            cardCount: 0,
            played: [{ design: 'SPADE', value: 5, color: 'black' }],
            score: 0,
            isDealer: false,
            moggedIn: false,
          },
          { id: 1, isHuman: false, cards: [], cardCount: 3, played: [], score: 0, isDealer: true, moggedIn: false },
        ],
      }),
    );
    expect(out).toContain('played:');
  });

  // **どの色役が付いたのかを名指す。**
  it('shows the show, skipping empty lines and naming the combo', () => {
    const out = formatCostlyColoursState(
      makeCostlyColoursState({
        phase: 'show',
        lastResult: {
          lines: [
            { key: 'jackDeuce', points: [2, 0] },
            { key: 'rank', points: [0, 0] },
            { key: 'colour', points: [6, 0] },
          ],
          totals: [8, 0],
          combos: ['costlyColours', ''],
        },
      }),
    );
    expect(out).toContain('jackDeuce: 2 - 0');
    expect(out).toContain('colour: 6 - 0');
    expect(out).not.toContain('rank:');
    expect(out).toContain('costlyColours');
    expect(out).toContain('deal total: 8 - 0');
  });

  // **ヒントは訊いたときだけ、札を指さない場合も出す。**
  it('shows the hint only when requested, with or without a card', () => {
    const carded = makeCostlyColoursState({ phase: 'play', hintHandIdx: 1, hintReason: 'fifteen' });
    expect(formatCostlyColoursState(carded)).not.toContain('HINT:');
    expect(formatCostlyColoursState({ ...carded, messageCode: 'costlycolours.hintRequested' })).toContain(
      'HINT: [1] (fifteen)',
    );

    const mogHint = makeCostlyColoursState({
      hintHandIdx: -1,
      hintReason: 'mog_refuse',
      messageCode: 'costlycolours.hintRequested',
    });
    expect(formatCostlyColoursState(mogHint)).toContain('HINT: (mog_refuse)');
  });

  it('shows the message and the winner', () => {
    expect(formatCostlyColoursState(makeCostlyColoursState({ message: 'hello' }))).toContain('hello');
    expect(formatCostlyColoursState(makeCostlyColoursState({ gameEndFlag: true, winnerIdx: 0 }))).toContain(
      'Game Over!',
    );
  });
});
