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
    expect(text).toContain('result: 0  payout: 0  net: -100');
    expect(text).toContain('Action: reset');
  });
  it('does not show any hint because the Mississippi Stud response has no hint payload', () => {
    const text = formatMississippiStudState(
      makeMississippiStudState({ message: 'No hint available', messageCode: 'mississippistud.noHint' }),
    );
    expect(text).not.toContain('HINT:');
  });
  it('only labels a response as a hint when the server marks it requested', () => {
    const requested = formatMississippiStudState(
      makeMississippiStudState({
        message: 'Stay in for the next street.',
        messageCode: 'mississippistud.hintRequested',
      }),
    );
    const passive = formatMississippiStudState(makeMississippiStudState({ message: 'Passive hint.' }));
    expect(requested).toContain('HINT: Stay in for the next street.');
    expect(passive).not.toContain('HINT:');
  });
});
