import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import type { DaifugoResponse } from '../../types/card';
import { DaifugoRulesBadges } from './DaifugoRulesBadges';

// Helper to simulate mobile viewport
function setViewportWidth(width: number) {
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: width });
  window.dispatchEvent(new Event('resize'));
}

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
    playableCardIndices: null,
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

  it('shows tooltip description on revolution badge', () => {
    render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
    const badge = screen.getByRole('button');
    expect(badge).toHaveAttribute('aria-label', expect.stringContaining('カードの強さが逆転しています'));
    expect(screen.getByRole('tooltip')).toHaveTextContent('カードの強さが逆転しています');
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

  it('shows multiple badges simultaneously with tooltips', () => {
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
    const tooltips = screen.getAllByRole('tooltip');
    expect(tooltips).toHaveLength(3);
  });

  it('badges are focusable buttons with aria-label for keyboard access', () => {
    render(<DaifugoRulesBadges state={makeState({ numberLocked: true })} />);
    const badge = screen.getByRole('button');
    expect(badge).toHaveAttribute('aria-label', expect.stringContaining('同じ数字のカードしか出せません'));
    expect(badge).toHaveClass('cursor-help');
  });

  describe('mobile collapsed view', () => {
    afterEach(() => {
      setViewportWidth(1024);
    });

    it('shows summary button instead of individual badges on mobile', () => {
      setViewportWidth(375);
      render(
        <DaifugoRulesBadges
          state={makeState({ revolutionActive: true, elevenBackActive: true, tableIsSequence: true })}
        />,
      );
      expect(screen.getByTestId('rules-summary-button')).toBeInTheDocument();
      expect(screen.getByTestId('rules-summary-button')).toHaveTextContent('特殊ルール発動中 (3)');
      // Individual badges should NOT be visible
      expect(screen.queryByText('革命中')).not.toBeInTheDocument();
    });

    it('opens modal when summary button is clicked', () => {
      setViewportWidth(375);
      render(<DaifugoRulesBadges state={makeState({ revolutionActive: true, elevenBackActive: true })} />);
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText('革命中')).toBeInTheDocument();
      expect(screen.getByText('11バック')).toBeInTheDocument();
    });

    it('closes modal when close button is clicked', () => {
      setViewportWidth(375);
      render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('closes modal on Escape key', () => {
      setViewportWidth(375);
      render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      fireEvent.keyDown(document, { key: 'Escape' });
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('closes modal on backdrop click', () => {
      setViewportWidth(375);
      render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      expect(screen.getByRole('dialog')).toBeInTheDocument();
      // Click the backdrop (presentation overlay)
      fireEvent.click(screen.getByRole('presentation'));
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('modal shows badge descriptions', () => {
      setViewportWidth(375);
      render(
        <DaifugoRulesBadges state={makeState({ revolutionActive: true, suitLocked: true, lockedSuit: 'SPADE' })} />,
      );
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      expect(screen.getByText('カードの強さが逆転しています（3が最強、2が最弱）')).toBeInTheDocument();
      expect(screen.getByText('同じスートのカードしか出せません')).toBeInTheDocument();
    });

    it('wraps focus to first element on Tab from last focusable element', () => {
      setViewportWidth(375);
      render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      const closeButton = screen.getByRole('button', { name: '閉じる' });
      closeButton.focus();
      fireEvent.keyDown(document, { key: 'Tab' });
      expect(document.activeElement).toBe(closeButton);
    });

    it('wraps focus to last element on Shift+Tab from first focusable element', () => {
      setViewportWidth(375);
      render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      const closeButton = screen.getByRole('button', { name: '閉じる' });
      closeButton.focus();
      fireEvent.keyDown(document, { key: 'Tab', shiftKey: true });
      expect(document.activeElement).toBe(closeButton);
    });

    it('ignores non-Escape non-Tab keys in modal keydown handler', () => {
      setViewportWidth(375);
      render(<DaifugoRulesBadges state={makeState({ revolutionActive: true })} />);
      fireEvent.click(screen.getByTestId('rules-summary-button'));
      fireEvent.keyDown(document, { key: 'Enter' });
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
  });
});
