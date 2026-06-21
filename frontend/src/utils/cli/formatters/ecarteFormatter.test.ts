import { describe, expect, it } from 'vitest';
import { makeEcarteState } from '../../../test/stateFactories';
import { formatEcarteState } from './ecarteFormatter';

describe('formatEcarteState', () => {
  it('renders the header, deal/trick and match scores', () => {
    const out = formatEcarteState(makeEcarteState({ matchScore: [3, 1] }));
    expect(out).toContain('Écarté');
    expect(out).toContain('deal: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
  });

  it('shows the negotiation sub-step during the Exchange phase', () => {
    const out = formatEcarteState(makeEcarteState({ phase: 0, negStep: 1 }));
    expect(out).toContain('negotiation: DealerRespond');
  });

  it('shows the trump symbol', () => {
    const out = formatEcarteState(makeEcarteState({ trumpSuit: 3 }));
    expect(out).toContain('trump: ♥');
  });

  it('shows the stock count', () => {
    const out = formatEcarteState(makeEcarteState({ stockRemaining: 12 }));
    expect(out).toContain('stock: 12');
  });

  it('flags a dealer refusal', () => {
    const out = formatEcarteState(makeEcarteState({ refusalByDealer: true }));
    expect(out).toContain('dealer refused');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatEcarteState(makeEcarteState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatEcarteState(
      makeEcarteState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'SPADE', value: 13 } },
          { playerIdx: 1, card: { design: 'HEART', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders a play hint with the card index', () => {
    const out = formatEcarteState(makeEcarteState({ phase: 1, hint: { cardIndex: 2, reason: 'follow_cut' } }));
    expect(out).toContain('HINT: play card index [2]');
    expect(out).toContain('follow_cut');
  });

  it('renders an exchange-action hint', () => {
    const out = formatEcarteState(makeEcarteState({ hint: { action: 'propose', reason: 'propose' } }));
    expect(out).toContain('HINT: propose');
  });

  it('renders the game-over banner with the winner', () => {
    const out = formatEcarteState(makeEcarteState({ phase: 3, gameEndFlag: true, winnerIdx: 0 }));
    expect(out).toContain('Game Over! Winner:');
  });

  it('renders an explicit message when present', () => {
    const out = formatEcarteState(makeEcarteState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });
});
