import { describe, expect, it } from 'vitest';
import { makePiedmonteseTarotState } from '../../../test/stateFactories';
import { formatPiedmonteseTarotState } from './piedmonteseTarotFormatter';

describe('formatPiedmonteseTarotState', () => {
  it('shows the deal, the trick counters and the seats', () => {
    const out = formatPiedmonteseTarotState(makePiedmonteseTarotState());
    expect(out).toContain('Tarocco Piemontese');
    expect(out).toContain('deal: 1');
    expect(out).toContain('trick: 1/19');
    expect(out).toContain('P0=0');
    // 自分の手札だけ番号付きで並ぶ。
    expect(out).toMatch(/\[0\]/);
  });

  // **捨てる枚数を出す。** 卓の大きさで 2 枚と 3 枚が変わるので、書かないと
  // 端末からはどちらを打てばよいのか分からない。
  it('says how many cards the scarto takes', () => {
    const four = formatPiedmonteseTarotState(
      makePiedmonteseTarotState({ phase: 0, isHumanScarto: true, talonSize: 2 }),
    );
    expect(four).toContain('bury 2 card(s)');

    const three = formatPiedmonteseTarotState(
      makePiedmonteseTarotState({ phase: 0, isHumanScarto: true, talonSize: 3 }),
    );
    expect(three).toContain('bury 3 card(s)');
  });

  it('omits the scarto line when a CPU deals', () => {
    const out = formatPiedmonteseTarotState(makePiedmonteseTarotState({ phase: 0, isHumanScarto: false }));
    expect(out).not.toContain('bury');
  });

  it('shows the trick in play', () => {
    const state = makePiedmonteseTarotState({
      currentTrick: [
        { playerIdx: 1, card: { design: 'HEART', value: 5, glyph: '♥', label: '5', color: 'red', deck: 'tarot' } },
      ],
    });
    expect(formatPiedmonteseTarotState(state)).toContain('trick:');
  });

  it('shows the settlement at the end of a deal', () => {
    const out = formatPiedmonteseTarotState(
      makePiedmonteseTarotState({ phase: 3, outcome: 1, dealScores: [12, -4, -4, -4] }),
    );
    expect(out).toContain('Above average');
    expect(out).toContain('P0=+12');
    expect(out).toContain('P1=-4');
  });

  it('announces the winner', () => {
    const out = formatPiedmonteseTarotState(makePiedmonteseTarotState({ gameEndFlag: true, winnerPlayer: 0 }));
    expect(out).toContain('Winner: Player 0');
    const draw = formatPiedmonteseTarotState(makePiedmonteseTarotState({ gameEndFlag: true, winnerPlayer: -1 }));
    expect(draw).toContain('Draw');
  });

  // 頼んでいないヒントは出さない。
  it('prints the hint only when it was requested', () => {
    const quiet = formatPiedmonteseTarotState(
      makePiedmonteseTarotState({ hint: { cardIndices: [0], reason: 'lead_low' } }),
    );
    expect(quiet).not.toContain('HINT');

    const asked = formatPiedmonteseTarotState(
      makePiedmonteseTarotState({
        hint: { cardIndices: [0], reason: 'lead_low' },
        messageCode: 'piedmontesetarot.hintRequested',
      }),
    );
    expect(asked).toContain('HINT: card indices [0] (lead_low)');
  });
});
