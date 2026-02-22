import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pokerApi } from '../api/gameApi';
import type { PokerResponse } from '../types/card';
import { PokerPage } from './PokerPage';

vi.mock('../api/gameApi', () => ({
  pokerApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(pokerApi.exec);

const phase0State: PokerResponse = {
  phase: 0,
  player: { cards: [], handName: '' },
  dealer: { cards: [], handName: '' },
  message: 'リセットしました',
};

const phase1State: PokerResponse = {
  phase: 1,
  player: {
    cards: [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 10 },
      { design: 'CLOVER', value: 3 },
      { design: 'SPADE', value: 7 },
    ],
    handName: '',
  },
  dealer: { cards: [], handName: '' },
  message: '',
};

const phase2State: PokerResponse = {
  phase: 2,
  player: {
    cards: [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 10 },
      { design: 'CLOVER', value: 3 },
      { design: 'SPADE', value: 7 },
    ],
    handName: 'High Card',
  },
  dealer: {
    cards: [
      { design: 'HEART', value: 2 },
      { design: 'DIAMOND', value: 4 },
      { design: 'CLOVER', value: 6 },
      { design: 'SPADE', value: 8 },
      { design: 'HEART', value: 10 },
    ],
    handName: 'Pair',
  },
  message: 'あなたの負け',
};

beforeEach(() => {
  mockExec.mockResolvedValue(phase0State);
});

describe('PokerPage', () => {
  it('calls reset command on mount', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  it('renders dealer and player section labels', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー手札/)).toBeInTheDocument());
    expect(screen.getByText(/プレイヤー手札/)).toBeInTheDocument();
  });

  it('shows 5 card backs for dealer in phase 0', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー手札/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('card back');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('exchange button is disabled in phase 0', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー手札/)).toBeInTheDocument());
    expect(screen.getByText('交換')).toBeDisabled();
  });

  it('stand button is disabled in phase 0', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー手札/)).toBeInTheDocument());
    expect(screen.getByText('スタンド')).toBeDisabled();
  });

  it('exchange button is enabled in phase 1', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).not.toBeDisabled());
  });

  it('stand button is enabled in phase 1', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('スタンド')).not.toBeDisabled());
  });

  it('shows player hand name in phase 2', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('High Card')).toBeInTheDocument());
  });

  it('shows dealer hand name in phase 2', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('Pair')).toBeInTheDocument());
  });

  it('shows dealer cards (not card backs) in phase 2', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('High Card')).toBeInTheDocument());
    // In phase 2, dealer cards are shown as real cards not card backs
    const cardBacks = screen.queryAllByAltText('card back');
    expect(cardBacks).toHaveLength(0);
  });

  it('shows result message', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの負け')).toBeInTheDocument());
  });

  it('calls exchange with selected indices when exchange button clicked', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).not.toBeDisabled());

    // Click first two player card images to select them
    const playerCardImgs = screen.getAllByAltText(/SPADE 1|HEART 5/);
    fireEvent.click(playerCardImgs[0]);
    fireEvent.click(playerCardImgs[1]);

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...phase1State, phase: 2 });
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange', expect.arrayContaining([0, 1])));
  });

  it('calls stand command when stand button is clicked in phase 1', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('スタンド')).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(phase2State);
    fireEvent.click(screen.getByText('スタンド'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand', undefined));
  });

  it('calls reset command when reset button is clicked', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
    mockExec.mockClear();
    mockExec.mockResolvedValue(phase0State);
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  it('shows instruction text in phase 1', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/交換したいカードをクリックして選択/)).toBeInTheDocument());
  });
});
