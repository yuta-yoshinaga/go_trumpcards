import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { deuceswildApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { VideoPokerResponse } from '../types/card';
import { DeucesWildPage } from './DeucesWildPage';

vi.mock('../api/gameApi', () => ({
  deuceswildApi: { exec: vi.fn() },
  actionLogApi: { deuceswild: vi.fn() },
}));

const mockExec = vi.mocked(deuceswildApi.exec);

const betPhaseState: VideoPokerResponse = {
  hand: [],
  phase: 1,
  chips: 1000,
  betAmount: 0,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'deuceswild',
  message: '',
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('DeucesWildPage', () => {
  it('calls reset on mount and renders bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<DeucesWildPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/チップ.*1000/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument();
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<DeucesWildPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
