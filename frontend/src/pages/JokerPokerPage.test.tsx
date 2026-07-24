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

// Four eights + a joker → Five of a Kind (a paying, wild-formed hand).
const jokerDrawState: VideoPokerResponse = {
  ...drawPhaseState,
  hand: [
    { design: 'HEART', value: 8 },
    { design: 'DIAMOND', value: 8 },
    { design: 'SPADE', value: 8 },
    { design: 'CLOVER', value: 8 },
    { design: 'JOKER', value: 0 },
  ],
};

// A junk hand below the Kings-or-Better minimum → no paying hand.
const noPayDrawState: VideoPokerResponse = {
  ...drawPhaseState,
  hand: [
    { design: 'HEART', value: 2 },
    { design: 'DIAMOND', value: 5 },
    { design: 'SPADE', value: 9 },
    { design: 'CLOVER', value: 11 },
    { design: 'HEART', value: 13 },
  ],
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

/** Drive the page from mount through deal to the DRAW phase with the given hand. */
async function playToDraw(draw: VideoPokerResponse) {
  mockExec.mockResolvedValueOnce(betPhaseState); // mount reset
  renderWithProviders(<JokerPokerPage />);
  await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());

  mockExec.mockResolvedValueOnce(draw);
  fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
  await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());
}

/** Drive the page from mount through deal + draw to a RESULT-phase state. */
async function playToResult(result: VideoPokerResponse) {
  await playToDraw(drawPhaseState);
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
    // The region remounts (key={resultNonce}) on each result, so query it fresh
    // after the announcement lands rather than holding a stale node reference.
    await waitFor(() =>
      expect(screen.getByTestId('vp-result-announce')).toHaveTextContent('役: フルハウス、配当 40 枚'),
    );
    const region = screen.getByTestId('vp-result-announce');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
  });

  it('announces the loss message when the payout is zero', async () => {
    await playToResult(loseResultState);
    await waitFor(() => expect(screen.getByTestId('vp-result-announce')).toHaveTextContent('役なし、ベット没収'));
  });

  it('does not show the made-hand readout before the deal (bet phase)', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<JokerPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());
    expect(screen.queryByTestId('vp-made-hand')).not.toBeInTheDocument();
  });

  it('shows the current made hand during the draw phase (full house)', async () => {
    await playToDraw(drawPhaseState);
    const made = screen.getByTestId('vp-made-hand');
    expect(made).toHaveTextContent('現在の役');
    expect(made).toHaveTextContent('フルハウス');
  });

  it('shows a wild-formed paying hand during the draw phase (five of a kind with a joker)', async () => {
    await playToDraw(jokerDrawState);
    expect(screen.getByTestId('vp-made-hand')).toHaveTextContent('ファイブカード');
  });

  it('shows the no-paying-hand label for a sub-minimum draw hand', async () => {
    await playToDraw(noPayDrawState);
    expect(screen.getByTestId('vp-made-hand')).toHaveTextContent('役なし（配当対象外）');
  });

  it('hides the made-hand readout after the draw (result phase)', async () => {
    await playToResult(winResultState);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByTestId('vp-made-hand')).not.toBeInTheDocument();
  });

  it('re-announces (nonce advances) even when two consecutive results are identical', async () => {
    await playToResult(winResultState);
    await waitFor(() => expect(screen.getByTestId('vp-result-announce')).toHaveTextContent('役: フルハウス'));
    const firstNonce = screen.getByTestId('vp-result-announce').getAttribute('data-nonce');

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

    expect(screen.getByTestId('vp-result-announce').getAttribute('data-nonce')).not.toBe(firstNonce);
  });
});
