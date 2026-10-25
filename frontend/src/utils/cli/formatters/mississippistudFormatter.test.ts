import { describe, expect, it } from 'vitest';
import { makeMississippiStudState } from '../../../test/stateFactories';
import { MississippiStudPhase, MississippiStudResult } from '../../../types/phases';
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
    expect(text).toContain('Action: play 1, 2, 3x or fold');
  });
  it('shows end results and fold status', () => {
    const text = formatMississippiStudState(
      makeMississippiStudState({
        phase: MississippiStudPhase.END,
        result: MississippiStudResult.LOSE,
        folded: true,
        totalPayout: 0,
        netChange: -100,
      }),
    );
    expect(text).toContain('phase: End');
    expect(text).toContain('Status: folded');
    expect(text).toContain('result: lose  payout: 0  net: -100');
    expect(text).toContain('hand rank:');
    expect(text).toContain('Action: reset');
  });
  it('shows a folded hand as a loss', () => {
    expect(
      formatMississippiStudState(
        makeMississippiStudState({ phase: MississippiStudPhase.END, result: MississippiStudResult.LOSE, folded: true }),
      ),
    ).toContain('result: lose');
  });
  it('shows unavailable antes and only affordable plays', () => {
    expect(
      formatMississippiStudState(makeMississippiStudState({ chips: 39, phase: MississippiStudPhase.ANTE })),
    ).toContain('Ante cannot be placed');
    expect(
      formatMississippiStudState(
        makeMississippiStudState({ chips: 150, anteAmount: 100, phase: MississippiStudPhase.THIRD_STREET }),
      ),
    ).toContain('Action: play 1x or fold');
  });
  it.each([
    [MississippiStudResult.WIN, 'win'],
    [MississippiStudResult.DRAW, 'push'],
    [MississippiStudResult.LOSE, 'lose'],
  ])('formats result %i as %s', (result, label) => {
    const text = formatMississippiStudState(makeMississippiStudState({ phase: MississippiStudPhase.END, result }));
    expect(text).toContain(`result: ${label}`);
  });
});
