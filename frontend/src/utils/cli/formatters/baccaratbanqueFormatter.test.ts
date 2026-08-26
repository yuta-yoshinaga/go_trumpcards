import { describe, expect, it } from 'vitest';
import { makeBaccaratBanqueState } from '../../../test/stateFactories';
import { formatBaccaratBanqueState } from './baccaratbanqueFormatter';

describe('formatBaccaratBanqueState', () => {
  it('names the phase and the coup', () => {
    const out = formatBaccaratBanqueState(makeBaccaratBanqueState());
    expect(out).toContain('Baccarat Banque');
    expect(out).toContain('Phase: BANKER DRAW');
    expect(out).toContain('Coup: 2');
  });

  // **負けても席が動かないことは残高から読めない。** 明示的に書く。
  it('says how long the bank has been held and that a loss does not end it', () => {
    const out = formatBaccaratBanqueState(makeBaccaratBanqueState({ bankHeld: 7 }));
    expect(out).toContain('Held this bank: 7 coup(s)');
    expect(out).toContain('a loss does not end it');
  });

  it('shows the shoe so the player can see the bank running out', () => {
    expect(formatBaccaratBanqueState(makeBaccaratBanqueState({ shoeRemaining: 12 }))).toContain(
      'Shoe: 12 card(s) left',
    );
  });

  it('renders all three seats with their totals and chips', () => {
    const out = formatBaccaratBanqueState(makeBaccaratBanqueState());
    expect(out).toContain('Banker (you)');
    expect(out).toContain('Right tableau');
    expect(out).toContain('Left tableau');
    expect(out).toContain('= 6');
    expect(out).toContain('chips 1000');
    expect(out).toContain('stake 50');
  });

  // **ナチュラルの印はなぜ 3 枚目が無いかの説明。**
  it('marks a natural, and does not mark an eight reached with three cards', () => {
    const natural = makeBaccaratBanqueState({
      players: makeBaccaratBanqueState().players.map((p) =>
        p.role === 'left' ? { ...p, cards: p.cards.slice(0, 2), total: 8 } : p,
      ),
    });
    expect(formatBaccaratBanqueState(natural)).toContain('*natural');

    const drawnToEight = makeBaccaratBanqueState({
      players: makeBaccaratBanqueState().players.map((p) => (p.role === 'right' ? { ...p, total: 8 } : p)),
    });
    const rightLine = formatBaccaratBanqueState(drawnToEight)
      .split('\n')
      .find((l) => l.startsWith('Right tableau')) as string;
    expect(rightLine).not.toContain('*natural');
  });

  // **左右は 1 行ずつ。** 差額だけだと、片方に払いもう片方から取ったクーが読めない。
  it('reports each tableau separately and then the bank net', () => {
    const out = formatBaccaratBanqueState(
      makeBaccaratBanqueState({
        phase: 'result',
        lastResult: {
          bankerTotal: 6,
          sides: [
            { seatIdx: 1, outcome: 'bankerWin', bet: 50, delta: 50 },
            { seatIdx: 2, outcome: 'punterWin', bet: 50, delta: -50 },
          ],
          bankerDelta: 0,
          bankerNatural: false,
        },
      }),
    );
    expect(out).toContain('Settlement (banker 6)');
    expect(out).toContain('Right tableau: banker wins (50)');
    expect(out).toContain('Left tableau: punter wins (-50)');
    expect(out).toContain('Banker net: 0');
  });

  it('distinguishes retiring from the bank ending any other way', () => {
    expect(formatBaccaratBanqueState(makeBaccaratBanqueState({ gameEndFlag: true, retired: true }))).toContain(
      'You gave up the bank.',
    );
    expect(formatBaccaratBanqueState(makeBaccaratBanqueState({ gameEndFlag: true, retired: false }))).toContain(
      'The bank has ended.',
    );
  });

  it('says nothing about the end while the bank is running', () => {
    const out = formatBaccaratBanqueState(makeBaccaratBanqueState());
    expect(out).not.toContain('gave up');
    expect(out).not.toContain('has ended');
  });
  it('shows a dash before any card is dealt', () => {
    const out = formatBaccaratBanqueState(
      makeBaccaratBanqueState({
        players: makeBaccaratBanqueState().players.map((p) => ({ ...p, cards: [], total: 0 })),
      }),
    );
    expect(out).toContain('Banker (you): — = 0');
    expect(out).not.toContain('stake 0');
  });

  it('signs a positive bank net so a win is not read as a loss', () => {
    const out = formatBaccaratBanqueState(
      makeBaccaratBanqueState({
        phase: 'result',
        lastResult: {
          bankerTotal: 8,
          sides: [
            { seatIdx: 1, outcome: 'bankerWin', bet: 50, delta: 50 },
            { seatIdx: 2, outcome: 'tie', bet: 50, delta: 0 },
          ],
          bankerDelta: 50,
          bankerNatural: true,
        },
      }),
    );
    expect(out).toContain('Left tableau: tie (0)');
    expect(out).toContain('Banker net: +50');
  });

  // 負のコントロール: 知らないフェーズを黙って別のフェーズに読み替えない。
  it('says UNKNOWN rather than guessing at an unrecognised phase', () => {
    expect(formatBaccaratBanqueState(makeBaccaratBanqueState({ phase: 'nonsense' }))).toContain('Phase: UNKNOWN');
  });
});
