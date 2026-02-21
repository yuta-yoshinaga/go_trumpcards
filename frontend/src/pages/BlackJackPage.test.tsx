import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { blackjackApi } from '../api/gameApi';
import type { BlackJackResponse } from '../types/card';
import { BlackJackPage } from './BlackJackPage';

vi.mock('../api/gameApi', () => ({
  blackjackApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(blackjackApi.exec);

const defaultState: BlackJackResponse = {
  dealer: { score: 17, cards: [{ design: 'SPADE', value: 1 }] },
  player: {
    score: 15,
    cards: [
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 10 },
    ],
  },
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('BlackJackPage', () => {
  it('calls reset command on mount', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders dealer and player section headings', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('ディーラー手札')).toBeInTheDocument());
    expect(screen.getByText('プレイヤー手札')).toBeInTheDocument();
  });

  it('displays dealer score when non-zero', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 17/)).toBeInTheDocument());
  });

  it('hides dealer score when zero (face-down)', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      dealer: { score: 0, cards: [{ design: 'SPADE', value: 1 }] },
    });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('ディーラー手札')).toBeInTheDocument());
    // Score 0 should not appear in the dealer score heading
    const headings = screen.getAllByText(/スコア/);
    expect(headings[0].textContent).toBe('スコア ');
  });

  it('shows card back image when dealer score is zero', async () => {
    mockExec.mockResolvedValue({
      ...defaultState,
      dealer: { score: 0, cards: [{ design: 'SPADE', value: 1 }] },
    });
    render(<BlackJackPage />);
    await waitFor(() => {
      const imgs = screen.getAllByRole('img');
      const cardBackImg = imgs.find((img) => img.getAttribute('src') === '/images/z01.png');
      expect(cardBackImg).toBeInTheDocument();
    });
  });

  it('displays player score', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 15/)).toBeInTheDocument());
  });

  it('renders Hit button', async () => {
    render(<BlackJackPage />);
    expect(screen.getByText('ヒット')).toBeInTheDocument();
  });

  it('renders Stand button', async () => {
    render(<BlackJackPage />);
    expect(screen.getByText('スタンド')).toBeInTheDocument();
  });

  it('renders Reset button', async () => {
    render(<BlackJackPage />);
    expect(screen.getByText('リセット')).toBeInTheDocument();
  });

  it('calls hit command when Hit button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...defaultState, player: { score: 20, cards: defaultState.player.cards } });
    fireEvent.click(screen.getByText('ヒット'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hit'));
  });

  it('calls stand command when Stand button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...defaultState, message: 'win' });
    fireEvent.click(screen.getByText('スタンド'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('calls reset command when Reset button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows message overlay when message is non-empty', async () => {
    mockExec.mockResolvedValue({ ...defaultState, message: 'あなたの勝ち！' });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show message overlay when message is empty', async () => {
    mockExec.mockResolvedValue({ ...defaultState, message: '' });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('ディーラー手札')).toBeInTheDocument());
    expect(screen.queryByText('あなたの勝ち！')).not.toBeInTheDocument();
  });
});
