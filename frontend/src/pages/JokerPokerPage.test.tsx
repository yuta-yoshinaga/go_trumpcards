import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { jokerpokerApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { VideoPokerResponse } from '../types/card';
import { JokerPokerPage } from './JokerPokerPage';

vi.mock('../api/gameApi', () => ({
  jokerpokerApi: { exec: vi.fn() },
  actionLogApi: { jokerpoker: vi.fn() },
}));

const mockExec = vi.mocked(jokerpokerApi.exec);

const HAND: VideoPokerResponse['hand'] = [
  { design: 'HEART', value: 10 },
  { design: 'HEART', value: 10 },
  { design: 'SPADE', value: 10 },
  { design: 'CLOVER', value: 5 },
  { design: 'DIAMOND', value: 5 },
];

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
  variantName: 'jokerpoker',
  message: '',
};

const drawPhaseState: VideoPokerResponse = {
  ...betPhaseState,
  hand: HAND,
  phase: 2,
  betAmount: 1,
  heldIndices: [false, false, false, false, false],
};

const winResultState: VideoPokerResponse = {
  ...betPhaseState,
  hand: HAND,
  phase: 3,
  betAmount: 1,
  result: 1,
  payout: 40,
  handName: 'Full House',
  heldIndices: [true, true, true, true, true],
};

const loseResultState: VideoPokerResponse = {
  ...betPhaseState,
  hand: HAND,
  phase: 3,
  betAmount: 1,
  result: 0,
  payout: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
};

/** Drive the page from mount through deal + draw to a RESULT-phase state. */
async function playToResult(result: VideoPokerResponse) {
  mockExec.mockResolvedValueOnce(betPhaseState); // mount reset
  renderWithProviders(<JokerPokerPage />);
  await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());

  mockExec.mockResolvedValueOnce(drawPhaseState);
  fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
  await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

  mockExec.mockResolvedValueOnce(result);
  fireEvent.click(screen.getByRole('button', { name: /ドロー/ }));
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('JokerPokerPage', () => {
  it('calls reset on mount and renders bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<JokerPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/チップ.*1000/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument();
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<JokerPokerPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('announces the winning hand and payout in a polite live region on result', async () => {
    await playToResult(winResultState);
    const region = await screen.findByTestId('vp-result-announce');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    await waitFor(() => expect(region).toHaveTextContent('役: フルハウス、配当 40 枚'));
  });

  it('announces the loss message when the payout is zero', async () => {
    await playToResult(loseResultState);
    const region = await screen.findByTestId('vp-result-announce');
    await waitFor(() => expect(region).toHaveTextContent('役なし、ベット没収'));
  });

  it('re-announces (nonce advances) even when two consecutive results are identical', async () => {
    await playToResult(winResultState);
    await waitFor(() => expect(screen.getByTestId('vp-result-announce')).toHaveTextContent('役: フルハウス'));
    const firstNonce = screen.getByTestId('vp-result-nonce').textContent;

    // Next hand: reset → deal → draw → an identical winning result.
    mockExec.mockResolvedValueOnce(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(winResultState);
    fireEvent.click(screen.getByRole('button', { name: /ドロー/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.getByTestId('vp-result-nonce').textContent).not.toBe(firstNonce);
  });
});
