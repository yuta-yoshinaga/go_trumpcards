import { describe, expect, it } from 'vitest';
import { makeMississippiStudState } from '../../../test/stateFactories';
import { MississippiStudPhase } from '../../../types/phases';
import { formatMississippiStudState } from './mississippistudFormatter';

describe('formatMississippiStudState', () => {
  it('shows ante decisions and chip balance', () => {
    const text = formatMississippiStudState(makeMississippiStudState({ chips: 800, phase: MississippiStudPhase.ANTE }));
    expect(text).toContain('phase: Ante');
    expect(text).toContain('chips: 800');
    expect(text).toContain('Action: bet <amount>');
    expect(text).toContain('Ante range: min 10, step 10, max 200');
  });
  it('shows street, indexed hand cards, masked community cards, and available actions', () => {
    const text = formatMississippiStudState(
      makeMississippiStudState({
        phase: MississippiStudPhase.FOURTH_STREET,
        playerHand: [
          { design: 'SPADE', value: 13 },
          { design: 'HEART', value: 1 },
        ],
        communityCards: [
          { design: '', value: 0 },
          { design: 'DIAMOND', value: 3 },
          { design: '', value: 0 },
        ],
      }),
    );
    expect(text).toContain('phase: FourthStreet');
    expect(text).toContain('[0]♠K');
    expect(text).toContain('[1]♥A');
    expect(text).toContain('community: [hidden] ♦3 [hidden]');
    expect(text).toContain('Action: play <1-3> or fold');
  });
  it('shows end results and fold status', () => {
    const text = formatMississippiStudState(
      makeMississippiStudState({ phase: MississippiStudPhase.END, folded: true, totalPayout: 0, netChange: -100 }),
    );
    expect(text).toContain('phase: End');
    expect(text).toContain('Status: folded');
    expect(text).toContain('result: lose  payout: 0  net: -100');
    expect(text).toContain('hand rank:');
    expect(text).toContain('Action: reset');
  });
  it.each([
    [1, 'win'],
    [-1, 'push'],
    [0, 'lose'],
  ])('formats result %i as %s', (result, label) => {
    const text = formatMississippiStudState(makeMississippiStudState({ phase: MississippiStudPhase.END, result }));
    expect(text).toContain(`result: ${label}`);
  });
});
