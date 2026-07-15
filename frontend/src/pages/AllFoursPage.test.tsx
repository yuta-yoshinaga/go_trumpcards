import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { allfoursApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AllFoursResponse } from '../types/card';
import { AllFoursPhase } from '../types/phases';
import { AllFoursPage } from './AllFoursPage';

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockUseCliMode = vi.mocked(useCliMode);

vi.mock('../api/gameApi', () => ({
  allfoursApi: { exec: vi.fn() },
  actionLogApi: { allfours: vi.fn() },
}));

const mockExec = vi.mocked(allfoursApi.exec);

const baseState: AllFoursResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 6,
      cards: [{ design: 'HEART', value: 5 }],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 6, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  phase: AllFoursPhase.BEG,
  roundNumber: 1,
  trickNumber: 0,
  dealerIdx: 1,
  nonDealerIdx: 0,
  currentPlayerIdx: 0,
  trumpSuit: 3,
  turnUp: { design: 'HEART', value: 7 },
  runCount: 0,
  currentTrick: [],
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: -1,
  validPlayIndices: [0],
  config: { cpuDifficulty: 1, pointLimit: 7 },
  message: '',
  messageCode: 'allfours.begPhase',
};

const playState: AllFoursResponse = {
  ...baseState,
  phase: AllFoursPhase.PLAY,
  trickNumber: 1,
  currentPlayerIdx: 0,
  validPlayIndices: [0],
};

const gameEndState: AllFoursResponse = {
  ...baseState,
  phase: AllFoursPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    { id: 0, isHuman: true, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 7, trickCount: 6 },
    { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 3, trickCount: 0 },
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('AllFoursPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows stand and beg buttons in beg phase', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベグ' })).toBeInTheDocument();
  });

  it('stand button calls exec with beg=false', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('beg', false));
  });

  it('beg button calls exec with beg=true', async () => {
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベグ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ベグ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('beg', true));
  });

  it('shows gift/run buttons when dealer is human in gift phase', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, phase: AllFoursPhase.GIFT, dealerIdx: 0, nonDealerIdx: 1 });
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ギフト/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /ラン/ })).toBeInTheDocument();
  });

  it('play button calls exec with play and card index', async () => {
    mockExec.mockResolvedValueOnce(playState);
    renderWithProviders(<AllFoursPage />);
    // Select the first (valid) card, then play.
    await waitFor(() => expect(screen.getByRole('button', { name: '♥5' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '♥5' }));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 0));
  });

  it('reads out the trump suit and turn-up by name', async () => {
    renderWithProviders(<AllFoursPage />); // trumpSuit 3 = ♥, turnUp ♥7
    expect(await screen.findByRole('img', { name: '切り札: ハート' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'めくり札: ハート7' })).toBeInTheDocument();
  });

  it('exposes the hint toggle as a labelled checkbox in the settings panel', async () => {
    renderWithProviders(<AllFoursPage />);
    const toggle = await screen.findByRole('checkbox', { name: /ヒント/ });
    expect(toggle).toBeInTheDocument();
  });

  it('reads the trump as unset and omits the turn-up before it is decided', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, trumpSuit: 0, turnUp: null });
    renderWithProviders(<AllFoursPage />);
    expect(await screen.findByRole('img', { name: '切り札: 未確定' })).toBeInTheDocument();
    // No turn-up card is shown before it is flipped.
    expect(screen.queryByRole('img', { name: /めくり札/ })).not.toBeInTheDocument();
  });

  it('shows winner message at game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<AllFoursPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝利！')).toBeInTheDocument());
  });
});
