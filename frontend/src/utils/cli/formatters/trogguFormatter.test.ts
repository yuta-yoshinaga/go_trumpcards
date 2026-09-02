import { describe, expect, it } from 'vitest';
import { makeTrogguState } from '../../../test/stateFactories';
import { formatTrogguState } from './trogguFormatter';

describe('formatTrogguState', () => {
  it('formats a play-phase state with header, players, and contract', () => {
    const out = formatTrogguState(makeTrogguState());
    expect(out).toContain('Troggu');
    expect(out).toContain('phase: Bid');
    expect(out).toContain('contract: -');
    expect(out).toContain('P0 (you)');
  });

  it('displays the solo contract with soloTarget number in the contract line', () => {
    const out = formatTrogguState(
      makeTrogguState({
        phase: 1,
        contractName: 'solo',
        soloTarget: 92,
      }),
    );
    expect(out).toContain('contract: Solo (92+ pts)');
  });

  it('formats other contracts correctly', () => {
    expect(formatTrogguState(makeTrogguState({ contractName: 'trois' }))).toContain('contract: Trois (3 tricks)');
    expect(formatTrogguState(makeTrogguState({ contractName: 'piccolo' }))).toContain(
      'contract: Piccolo (exactly 1 trick)',
    );
    expect(formatTrogguState(makeTrogguState({ contractName: 'misere' }))).toContain('contract: Misere (no tricks)');
    expect(formatTrogguState(makeTrogguState({ contractName: 'unknown_contract' }))).toContain(
      'contract: unknown_contract',
    );
  });

  it('renders the current trick when cards are present', () => {
    const out = formatTrogguState(
      makeTrogguState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'SPADE', value: 14, glyph: '♠', label: 'K', color: 'black', deck: 'tarot' } },
          { playerIdx: 1, card: { design: 'CLOVER', value: 1, glyph: '♣', label: '1', color: 'black', deck: 'tarot' } },
        ],
      }),
    );
    expect(out).toContain('trick:');
    expect(out).toContain('P0:');
  });

  it('renders breakdown results in points for solo and tricks for others', () => {
    const soloOut = formatTrogguState(
      makeTrogguState({
        breakdown: {
          contract: 2,
          contractName: 'solo',
          declarerPoints: 100,
          declarerTricks: 10,
          target: 92,
          targetIsTricks: false,
          won: true,
          base: 20,
          seats: [60, -20, -20, -20],
        },
      }),
    );
    expect(soloOut).toContain('result: made — 100 pts (target 92)');

    const tricksOut = formatTrogguState(
      makeTrogguState({
        breakdown: {
          contract: 1,
          contractName: 'trois',
          declarerPoints: 20,
          declarerTricks: 3,
          target: 3,
          targetIsTricks: true,
          won: true,
          base: 10,
          seats: [30, -10, -10, -10],
        },
      }),
    );
    expect(tricksOut).toContain('result: made — 3 tricks (target 3)');
  });

  it('renders requested hint', () => {
    const out = formatTrogguState(
      makeTrogguState({
        messageCode: 'troggu.hintRequested',
        hint: { reason: 'bid_solo' },
      }),
    );
    expect(out).toContain('hint: bid_solo');
  });
});
