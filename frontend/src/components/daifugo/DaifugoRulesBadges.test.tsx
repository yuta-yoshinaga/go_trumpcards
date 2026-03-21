import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import type { DaifugoResponse } from '../../types/card';
import { DaifugoRulesBadges } from './DaifugoRulesBadges';

function makeState(overrides: Partial<DaifugoResponse> = {}): DaifugoResponse {
  return {
    players: [],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    revolutionActive: false,
    elevenBackActive: false,
    suitLocked: false,
    lockedSuit: '',
    tableIsSequence: false,
    config: {
      jokerCount: 2,
      eightCutEnabled: true,
      suitLockMode: 2,
      elevenBackEnabled: true,
      sequenceEnabled: true,
      cardExchangeEnabled: true,
      blindExchangeEnabled: false,
      fiveSkipEnabled: false,
      fiveSkipCount: 1,
      sevenPassEnabled: false,
      tenDiscardEnabled: false,
      spadeThreeEnabled: false,
      capitalFallEnabled: false,
      nineReverseEnabled: false,
      coupDetatEnabled: false,
      numberLockEnabled: false,
      sandstormEnabled: false,
      emperorEnabled: false,
      sequenceRevolutionEnabled: false,
      sequenceLockEnabled: false,
      illegalFinishEnabled: false,
      queenBomberEnabled: false,
      cpuDifficulty: 1,
    },
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    message: '',
    pendingAction: 'none',
    pendingActionTarget: -1,
    reverseDirection: false,
    numberLocked: false,
    sequenceLocked: false,
    sortMode: 0,
    ...overrides,
  };
}

describe('DaifugoRulesBadges', () => {
  it('returns null when no badges are active', () => {
    const { container } = render(<DaifugoRulesBadges state={makeState()} />);
    expect(container.innerHTML).toBe('');
  });

  it('shows revolution badge when revolutionActive is true', () => {
    render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
    expect(screen.getByText('革命中')).toBeInTheDocument();
  });

  it('shows elevenBack badge when elevenBackActive is true', () => {
    render(<DaifugoRulesBadges state={makeState({ elevenBackActive: true })} />);
    expect(screen.getByText('11バック')).toBeInTheDocument();
  });

  it('shows suitLock badge without partial suffix when suitLockMode is 2', () => {
    render(<DaifugoRulesBadges state={makeState({ suitLocked: true, lockedSuit: 'SPADE' })} />);
    const badge = screen.getByText('スート縛り: SPADE');
    expect(badge).toBeInTheDocument();
    expect(badge.textContent).not.toContain('片縛り');
  });

  it('shows suitLock badge with partial suffix when suitLockMode is 1', () => {
    render(
      <DaifugoRulesBadges
        state={makeState({
          suitLocked: true,
          lockedSuit: 'HEART',
          config: { ...makeState().config, suitLockMode: 1 },
        })}
      />,
    );
    expect(screen.getByText('スート縛り: HEART (片縛り)')).toBeInTheDocument();
  });

  it('shows sequence badge when tableIsSequence is true', () => {
    render(<DaifugoRulesBadges state={makeState({ tableIsSequence: true })} />);
    expect(screen.getByText('階段')).toBeInTheDocument();
  });

  it('shows nineReverse badge when reverseDirection is true', () => {
    render(<DaifugoRulesBadges state={makeState({ reverseDirection: true })} />);
    expect(screen.getByText('9リバース')).toBeInTheDocument();
  });

  it('shows numberLock badge when numberLocked is true', () => {
    render(<DaifugoRulesBadges state={makeState({ numberLocked: true })} />);
    expect(screen.getByText('数縛り')).toBeInTheDocument();
  });

  it('shows sequenceLock badge when sequenceLocked is true', () => {
    render(<DaifugoRulesBadges state={makeState({ sequenceLocked: true })} />);
    expect(screen.getByText('階段縛り')).toBeInTheDocument();
  });

  it('shows multiple badges simultaneously', () => {
    render(
      <DaifugoRulesBadges
        state={makeState({
          revolutionActive: true,
          elevenBackActive: true,
          tableIsSequence: true,
        })}
      />,
    );
    expect(screen.getByText('革命中')).toBeInTheDocument();
    expect(screen.getByText('11バック')).toBeInTheDocument();
    expect(screen.getByText('階段')).toBeInTheDocument();
  });
});
