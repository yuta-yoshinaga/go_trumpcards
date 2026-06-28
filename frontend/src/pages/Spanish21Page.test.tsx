import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spanish21Api } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BlackJackResponse } from '../types/card';
import { Spanish21Page } from './Spanish21Page';

vi.mock('../api/gameApi', () => ({
  spanish21Api: { exec: vi.fn() },
  actionLogApi: { spanish21: vi.fn() },
}));

const mockExec = vi.mocked(spanish21Api.exec);

const betPhaseState: BlackJackResponse = {
  dealer: { chips: 1000 },
  player: { chips: 1000 },
  phase: 1,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: '',
  hintEnabled: false,
  suggestedAction: 0,
  deckCount: 1,
  dealerHitsSoft17: false,
  countingEnabled: false,
  cpuPlayerCount: 0,
  runningCount: 0,
  trueCount: 0,
  perfectPairsBet: 0,
  twentyOnePlus3Bet: 0,
  doubleAfterSplit: true,
  countingSystem: 0,
  deckPenetration: 75,
  multiHandCount: 0,
  surrenderRule: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(betPhaseState);
});

describe('Spanish21Page', () => {
  it('resets via the Spanish 21 API on mount', async () => {
    renderWithProviders(<Spanish21Page />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('opens a Spanish 21 specific tutorial step describing the 48-card deck and bonuses', async () => {
    renderWithProviders(<Spanish21Page />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());

    // Steps: betControls → betButton → (Spanish 21) payout reference.
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toHaveTextContent('48枚デッキ'));
  });
});
