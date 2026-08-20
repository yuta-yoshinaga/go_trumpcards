import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { deuceswildApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, VideoPokerResponse } from '../types/card';
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

const card = (design: CardDesign, value: number): Card => ({ design, value });

const drawPhaseState: VideoPokerResponse = {
  hand: [card('SPADE', 1), card('HEART', 11), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 13)],
  phase: 2,
  chips: 997,
  betAmount: 3,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'deuceswild',
  message: '',
};

const resultPhaseState: VideoPokerResponse = {
  hand: [card('SPADE', 2), card('CLOVER', 2), card('HEART', 7), card('DIAMOND', 9), card('SPADE', 11)],
  phase: 3,
  chips: 999,
  betAmount: 1,
  result: -1,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [true, true, false, false, false],
  variantName: 'deuceswild',
  message: '',
};

/** Query the 5 hold-toggle card buttons (the only ones exposing aria-pressed). */
function holdButtons(): HTMLElement[] {
  return screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
}

/** Drive the page from mount into the draw phase with auto-hold disabled so
 * every card starts unheld and hold state is driven purely by the test. */
async function enterDrawPhaseWithAutoHoldOff() {
  renderWithProviders(<DeucesWildPage />);
  await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  fireEvent.click(screen.getByTestId('vp-auto-hold-toggle'));
  fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
  await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());
}

beforeEach(() => {
  vi.clearAllMocks();
  // Clear the persisted per-variant auto-hold/hint toggles so each case starts
  // from the hardcoded defaults instead of leaking state across tests.
  localStorage.clear();
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

  it('toggles hold on card 0 when the "1" key is pressed in draw phase', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    await enterDrawPhaseWithAutoHoldOff();

    const cards = holdButtons();
    expect(cards[0]).toHaveAttribute('aria-pressed', 'false');
    fireEvent.keyDown(document.body, { key: '1' });
    expect(cards[0]).toHaveAttribute('aria-pressed', 'true');
    // A second press releases the hold again.
    fireEvent.keyDown(document.body, { key: '1' });
    expect(cards[0]).toHaveAttribute('aria-pressed', 'false');
  });

  it('toggles hold on card 4 when the "5" key is pressed in draw phase', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    await enterDrawPhaseWithAutoHoldOff();

    const cards = holdButtons();
    expect(cards[4]).toHaveAttribute('aria-pressed', 'false');
    fireEvent.keyDown(document.body, { key: '5' });
    expect(cards[4]).toHaveAttribute('aria-pressed', 'true');
  });

  it('ignores the number keys outside the draw phase', async () => {
    mockExec.mockResolvedValue(resultPhaseState);
    renderWithProviders(<DeucesWildPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());

    // Result phase reflects the server-sent held indices [true, true, false, ...].
    expect(screen.getByTestId('vp-hold-badge-0')).toBeInTheDocument();
    expect(screen.queryByTestId('vp-hold-badge-2')).not.toBeInTheDocument();
    // Number keys are gated to draw phase, so nothing changes here.
    fireEvent.keyDown(document.body, { key: '1' });
    fireEvent.keyDown(document.body, { key: '3' });
    expect(screen.getByTestId('vp-hold-badge-0')).toBeInTheDocument();
    expect(screen.queryByTestId('vp-hold-badge-2')).not.toBeInTheDocument();
  });
});

// #5507: madeHand は jokerpoker でしか出ておらず、Deuces Wild では常に無効だった。
// スリーカード以上でしか配当が出ない=分散が大きい変種なので、ドロー前に配当対象かを
// 知れる価値はむしろ大きい。
describe('DeucesWildPage made-hand readout', () => {
  const drawTo = async (hand: VideoPokerResponse['hand']) => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce({ ...drawPhaseState, hand });
    renderWithProviders(<DeucesWildPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());
  };

  it('counts a two as wild when naming the hand', async () => {
    await drawTo([card('SPADE', 5), card('HEART', 5), card('CLOVER', 2), card('DIAMOND', 9), card('SPADE', 12)]);
    const made = screen.getByTestId('vp-made-hand');
    expect(made).toHaveTextContent('現在の役');
    expect(made).toHaveTextContent('スリーカード');
  });

  // **配当の下限はスリーカード。** ツーペアでは「役なし」。
  it('says nothing pays below three of a kind', async () => {
    await drawTo([card('SPADE', 5), card('HEART', 5), card('CLOVER', 9), card('DIAMOND', 9), card('SPADE', 12)]);
    expect(screen.getByTestId('vp-made-hand')).toHaveTextContent('役なし');
  });

  it('names four deuces on its own row', async () => {
    await drawTo([card('SPADE', 2), card('HEART', 2), card('CLOVER', 2), card('DIAMOND', 2), card('SPADE', 9)]);
    expect(screen.getByTestId('vp-made-hand')).toHaveTextContent('フォーデューシーズ');
  });
});
