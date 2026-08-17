import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { videopokerApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, VideoPokerResponse } from '../types/card';
import { VideoPokerPage } from './VideoPokerPage';

vi.mock('../api/gameApi', () => ({
  videopokerApi: { exec: vi.fn() },
  actionLogApi: { videopoker: vi.fn() },
}));

const mockExec = vi.mocked(videopokerApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

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
  variantName: 'jacksorbetter',
  message: '',
};

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
  variantName: 'jacksorbetter',
  message: '',
};

const resultPhaseWin: VideoPokerResponse = {
  hand: [card('SPADE', 11), card('CLOVER', 11), card('HEART', 3), card('DIAMOND', 5), card('SPADE', 9)],
  phase: 3,
  chips: 1001,
  betAmount: 1,
  result: 1,
  payout: 1,
  handRank: 1,
  handName: 'Jacks or Better',
  heldIndices: [true, true, false, false, false],
  message: 'Jacks or Better! You win!',
  messageCode: 'videopoker.result.win',
  variantName: 'jacksorbetter',
  messageParams: { handName: 'Jacks or Better', payout: '1' },
};

const resultPhaseLose: VideoPokerResponse = {
  hand: [card('SPADE', 2), card('CLOVER', 5), card('HEART', 7), card('DIAMOND', 9), card('SPADE', 11)],
  phase: 3,
  chips: 999,
  betAmount: 1,
  result: -1,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  message: 'No winning hand.',
  variantName: 'jacksorbetter',
  messageCode: 'videopoker.result.lose',
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('VideoPokerPage', () => {
  it('calls reset on mount and renders bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/チップ.*1000/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument();
  });

  it('renders draw phase with 5 cards', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());
  });

  it('draw phase: number key 1 toggles hold on the first card and announces it', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await screen.findByRole('button', { name: /ドロー/ });

    expect(screen.queryByTestId('vp-hold-badge-0')).not.toBeInTheDocument();
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(screen.getByTestId('vp-hold-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('vp-hold-announce').textContent).toMatch(/カード1をホールド/);

    // Pressing it again releases the hold.
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(screen.queryByTestId('vp-hold-badge-0')).not.toBeInTheDocument());
    expect(screen.getByTestId('vp-hold-announce').textContent).toMatch(/カード1のホールドを解除/);
  });

  it('number keys do not toggle hold outside the draw phase', async () => {
    // Result phase: held cards come from the server; local keyboard toggles are ignored.
    mockExec.mockResolvedValue(resultPhaseWin);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByText(/次のゲーム/)).toBeInTheDocument());
    // Card index 2 is not held in resultPhaseWin.
    expect(screen.queryByTestId('vp-hold-badge-2')).not.toBeInTheDocument();
    fireEvent.keyDown(document.body, { key: '3' });
    expect(screen.queryByTestId('vp-hold-badge-2')).not.toBeInTheDocument();
    expect(screen.getByTestId('vp-hold-announce').textContent).toBe('');
  });

  it('renders result phase with win message', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByText(/次のゲーム/)).toBeInTheDocument());
  });

  it('renders result phase with lose message', async () => {
    mockExec.mockResolvedValue(resultPhaseLose);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByText(/次のゲーム/)).toBeInTheDocument());
  });

  it('next game button fires reset directly without confirm dialog', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByText(/次のゲーム/)).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /次のゲーム/ }));
    expect(screen.queryByText(/リセットしますか/)).not.toBeInTheDocument();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});

// #5506: madeHand は gameName !== 'jokerpoker' で早期 null を返しており、
// **ワイルドで救われない Jacks or Better では常に非表示**だった。
describe('VideoPokerPage made-hand readout', () => {
  const drawTo = async (state: VideoPokerResponse) => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(state);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());
  };

  it('names a paying hand during the draw phase', async () => {
    await drawTo({
      ...drawPhaseState,
      hand: [card('SPADE', 11), card('HEART', 11), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 3)],
    });
    const made = screen.getByTestId('vp-made-hand');
    expect(made).toHaveTextContent('現在の役');
    expect(made).toHaveTextContent('ジャックス・オア・ベター');
  });

  // **最低ラインに届かない手は「役なし」と出す。** 表示が消えると、評価されて
  // いないのか届いていないのか区別できない。
  it('says so when the hand pays nothing', async () => {
    await drawTo({
      ...drawPhaseState,
      hand: [card('SPADE', 10), card('HEART', 10), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 3)],
    });
    expect(screen.getByTestId('vp-made-hand')).toHaveTextContent('役なし');
  });

  it('stays hidden before the deal', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<VideoPokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());
    expect(screen.queryByTestId('vp-made-hand')).not.toBeInTheDocument();
  });
});
