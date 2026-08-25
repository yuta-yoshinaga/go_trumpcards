import { describe, expect, it } from 'vitest';
import { makeCometState } from '../../../test/stateFactories';
import { formatCometState } from './cometFormatter';

describe('formatCometState', () => {
  it('shows the round, the dead hand and the seats', () => {
    const out = formatCometState(makeCometState());
    expect(out).toContain('Comet');
    expect(out).toContain('round: 1');
    expect(out).toContain('target: 100');
    // **死に手の枚数は見せる。** ここに眠った札で連なりが止まる。
    expect(out).toContain('dead hand: 3');
    expect(out).toContain('score=0');
  });

  // **スートは問わない。** 数字だけで昇ることを毎回書く。
  it('names the rank needed and says suit does not matter', () => {
    expect(formatCometState(makeCometState())).toContain('need: lead any card');
    const out = formatCometState(makeCometState({ need: 8 }));
    expect(out).toContain('need: rank 8 (any suit)');
    expect(out).not.toContain('lead any card');
  });

  it('shows the sequence, empty at first', () => {
    expect(formatCometState(makeCometState())).toContain('sequence: (nothing yet)');
    const out = formatCometState(
      makeCometState({
        pile: [
          { design: 'SPADE', value: 5, color: 'black' },
          { design: 'HEART', value: 6, color: 'red' },
        ],
      }),
    );
    expect(out).toMatch(/sequence: .*->.*/);
  });

  // **連なりが長くなっても末尾だけ見せる。** 全部並べると読めない。
  it('trims a long sequence to its tail', () => {
    const pile = Array.from({ length: 20 }, (_, i) => ({
      design: 'SPADE' as const,
      value: (i % 13) + 1,
      color: 'black' as const,
    }));
    const out = formatCometState(makeCometState({ pile }));
    const line = out.split('\n').find((l) => l.startsWith('sequence:')) ?? '';
    expect(line.split('->')).toHaveLength(8);
  });

  // **出せる札が無いならパスしかない、と書く。** 探させない。
  it('lists the playable indices, or says to pass', () => {
    expect(formatCometState(makeCometState())).toContain('playable: 0 1 2');
    expect(formatCometState(makeCometState({ playableIdxs: [] }))).toContain('you must pass');
  });

  it('omits the playable line when it is not the human turn', () => {
    const out = formatCometState(makeCometState({ isHumanTurn: false, currentPlayerIdx: 1 }));
    expect(out).not.toContain('playable:');
  });

  it('shows the round result including the unplayed kings', () => {
    const out = formatCometState(
      makeCometState({
        lastResult: { winnerIdx: 0, cardsLeft: [0, 2, 3, 1], gained: [13, 0, 0, 0], unplayedKings: 2, heldWildIdx: 2 },
      }),
    );
    expect(out).toContain('went out for 13');
    expect(out).toContain('kings never played: 2');
    expect(out).toContain('held the Comet: -1');
  });

  it('omits the comet line when nobody held it', () => {
    const out = formatCometState(
      makeCometState({
        lastResult: { winnerIdx: 0, cardsLeft: [0, 1, 1, 1], gained: [4, 0, 0, 0], unplayedKings: 0, heldWildIdx: -1 },
      }),
    );
    expect(out).not.toContain('held the Comet');
  });

  // **ヒントは訊いたときだけ出す。**
  it('shows the hint only when it was requested', () => {
    const hinted = makeCometState({ hintHandIdx: 2, hintReason: 'comet' });
    expect(formatCometState(hinted)).not.toContain('HINT:');
    expect(formatCometState({ ...hinted, messageCode: 'comet.hintRequested' })).toContain('HINT: play [2] (comet)');
  });

  it('shows the message and the winner', () => {
    expect(formatCometState(makeCometState({ message: 'hello' }))).toContain('hello');
    expect(formatCometState(makeCometState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('Game Over!');
  });
});
