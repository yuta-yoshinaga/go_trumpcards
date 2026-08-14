import { describe, expect, it } from 'vitest';
import { makeHorseState } from '../../../test/stateFactories';
import { formatHorseState } from './horseFormatter';

describe('formatHorseState', () => {
  it('names the discipline, the hand counters and the pot', () => {
    const out = formatHorseState(makeHorseState());
    expect(out).toContain("H — Texas Hold'em");
    expect(out).toContain('hand 1/2');
    expect(out).toContain('pot: 30');
    expect(out).toContain('to call: 20');
  });

  it('lists every seat with its chips', () => {
    const out = formatHorseState(makeHorseState());
    for (const name of ['YOU', 'CPU1', 'CPU2', 'CPU3']) {
      expect(out).toContain(name);
    }
    expect(out).toContain('1000 chips');
  });

  // **見えている札だけを並べる。** 伏せ札はそもそも届かない。
  it('shows the cards a seat reveals and nothing else', () => {
    const out = formatHorseState(makeHorseState());
    expect(out).toMatch(/YOU \(you\): 1000 chips {2}\S+ \S+/);
    expect(out).toMatch(/CPU1: 1000 chips(\s+<-)?$/m);
  });

  it('shows the board once it is dealt', () => {
    const out = formatHorseState(
      makeHorseState({
        communityCards: [{ design: 'SPADE', value: 10, glyph: '♠', label: '10', color: 'black', deck: 'standard' }],
      }),
    );
    expect(out).toContain('board:');
  });

  it('omits the board in the stud disciplines', () => {
    const out = formatHorseState(makeHorseState({ disciplineName: 'razz', disciplineLetter: 'R' }));
    expect(out).toContain('Razz');
    expect(out).not.toContain('board:');
  });

  it('names the winner once the match is over', () => {
    const out = formatHorseState(makeHorseState({ phase: 2, gameEndFlag: true, winnerSeat: 1 }));
    expect(out).toContain('winner: CPU1');
  });
});
