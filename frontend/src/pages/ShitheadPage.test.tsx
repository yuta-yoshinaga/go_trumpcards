import { fireEvent, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { shitheadApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { ShitheadResponse } from '../types/card';
import { ShitheadPage } from './ShitheadPage';

vi.mock('../api/gameApi', () => ({
  shitheadApi: { exec: vi.fn() },
  actionLogApi: { shithead: vi.fn() },
}));

const mockExec = vi.mocked(shitheadApi.exec);

const baseConfig = {
  magicTwo: true,
  magicSeven: true,
  magicEight: true,
  magicTen: true,
  fourOfAKindBurn: true,
  cpuDifficulty: 1,
};

const humanTurnState: ShitheadResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      rank: -1,
      handCount: 3,
      handCards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 9 },
        { design: 'DIAMOND', value: 2 },
      ],
      faceUpCards: [
        { design: 'CLOVER', value: 11 },
        { design: 'CLOVER', value: 12 },
        { design: 'CLOVER', value: 13 },
      ],
      faceDownCount: 3,
    },
    {
      id: 1,
      isHuman: false,
      isFinished: false,
      rank: -1,
      handCount: 3,
      handCards: [],
      faceUpCards: [],
      faceDownCount: 3,
    },
  ],
  currentTurn: 0,
  currentSource: 'hand',
  discardPile: [{ design: 'HEART', value: 5 }],
  stockSize: 16,
  skipNext: false,
  sevenActive: false,
  gameEndFlag: false,
  config: baseConfig,
  cpuActions: [],
  message: '',
};

const cpuTurnState: ShitheadResponse = {
  ...humanTurnState,
  currentTurn: 1,
};

const gameEndState: ShitheadResponse = {
  ...humanTurnState,
  gameEndFlag: true,
  players: humanTurnState.players.map((p, i) => ({ ...p, isFinished: true, rank: i + 1 })),
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('ShitheadPage', () => {
  it('shows loading message before state arrives', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    expect(screen.getByText(/Loading|読み込み/i)).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: baseConfig,
      }),
    );
  });

  it('renders human player hand and play / pickup controls on human turn', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: /出す/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /場札を引き取る/ })).toBeInTheDocument();
  });

  it('play button disabled until at least one card selected', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: /出す/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /出す/ })).toBeDisabled();
  });

  it('dispatches play with selected indices', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: /出す/ })).toBeInTheDocument());
    // Click the first hand card (button labelled with suit + value)
    const cardButtons = screen.getAllByRole('button').filter((b) => /^♠|♥|♦|♣/.test(b.textContent ?? ''));
    expect(cardButtons.length).toBeGreaterThan(0);
    fireEvent.click(cardButtons[0]);
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: /出す/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { indices: [0] }));
  });

  it('pickup dispatches play with empty indices', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: /場札を引き取る/ })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: /場札を引き取る/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { indices: [] }));
  });

  it('hides player controls when CPU has the turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: /出す/ })).not.toBeInTheDocument();
  });

  it('shows game end labels when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getAllByText(/順位|シットヘッド/).length).toBeGreaterThan(0));
  });

  it('shows magic seven indicator when sevenActive is true', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, sevenActive: true });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/7: 次は7以下/)).toBeInTheDocument());
  });

  it('shows magic eight indicator when skipNext is true', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, skipNext: true });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/8: 次プレイヤーをスキップ/)).toBeInTheDocument());
  });

  it('shows blind play button when currentSource is facedown', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, currentSource: 'facedown' });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: /ブラインドで出す/ })).toBeInTheDocument());
  });

  it('renders magic-card badges for enabled magic ranks in the human hand', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('sh-magic-badge-2-2')).toBeInTheDocument());
    expect(screen.queryByTestId('sh-magic-badge-5-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sh-magic-badge-9-1')).not.toBeInTheDocument();
  });

  it('omits the magic badge when the rule is disabled in config', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...baseConfig, magicTwo: false },
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText('5')).toBeInTheDocument());
    expect(screen.queryByTestId('sh-magic-badge-2-2')).not.toBeInTheDocument();
  });

  it('shows ? for joker design in discard pile description', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      discardPile: [{ design: 'JOKER', value: 0 }],
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => {
      const matches = screen.getAllByText(/\?0/);
      expect(matches.length).toBeGreaterThan(0);
    });
  });
});
