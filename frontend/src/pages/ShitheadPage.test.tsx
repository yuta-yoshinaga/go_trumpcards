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
    // Selectable hand cards are the buttons that expose aria-pressed.
    const cardButtons = screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
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

  it('shows the current-source banner on the human turn and reflects the source', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, currentSource: 'facedown' });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    const banner = await screen.findByTestId('sh-source-banner');
    expect(banner).toHaveTextContent(/裏向き/);
  });

  it('hides the source banner when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('sh-source-banner')).not.toBeInTheDocument();
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
    await waitFor(() => expect(screen.getByRole('button', { name: /出す/ })).toBeInTheDocument());
    expect(screen.queryByTestId('sh-magic-badge-2-2')).not.toBeInTheDocument();
  });

  it('renders magic-card badges on face-up cards too', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      players: [
        {
          ...humanTurnState.players[0],
          faceUpCards: [
            { design: 'HEART', value: 10 }, // burn — magic
            { design: 'CLOVER', value: 11 }, // J — not magic
            { design: 'CLOVER', value: 12 }, // Q — not magic
          ],
        },
        humanTurnState.players[1],
      ],
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('sh-magic-badge-10-0')).toBeInTheDocument());
    expect(screen.queryByTestId('sh-magic-badge-11-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sh-magic-badge-12-2')).not.toBeInTheDocument();
  });

  it('exposes the magic-card effect to assistive tech via an sr-only span', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    // humanTurnState.handCards[2] is the 2 of diamonds; its sr-only text describes the reset effect.
    await waitFor(() => expect(screen.getByTestId('sh-magic-badge-2-2')).toBeInTheDocument());
    expect(screen.getByText('リセット: 次のプレイヤーは何でも出せます')).toBeInTheDocument();
  });

  it('renders face-down cards as card-back images rather than "?" buttons', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    // faceDownCount is 3 for the human player.
    await waitFor(() => expect(screen.getByTestId('sh-facedown-0')).toBeInTheDocument());
    expect(screen.getByTestId('sh-facedown-1')).toBeInTheDocument();
    expect(screen.getByTestId('sh-facedown-2')).toBeInTheDocument();
    // Each face-down slot renders a card-back image, not a literal "?".
    expect(screen.getAllByTestId('animated-card-back').length).toBeGreaterThanOrEqual(3);
    expect(screen.getByTestId('sh-facedown-0')).not.toHaveTextContent('?');
  });

  it('makes face-down cards selectable with a ring when currentSource is facedown', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, currentSource: 'facedown' });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    const first = await screen.findByTestId('sh-facedown-0');
    expect(first).toBeEnabled();
    expect(first).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(first);
    await waitFor(() => expect(screen.getByTestId('sh-facedown-0')).toHaveAttribute('aria-pressed', 'true'));
    expect(screen.getByTestId('sh-facedown-0').className).toContain('ring-ds-warning');
  });

  it('disables face-down cards when the current source is not facedown', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    // Default humanTurnState currentSource is 'hand'.
    const first = await screen.findByTestId('sh-facedown-0');
    expect(first).toBeDisabled();
    expect(first).not.toHaveAttribute('aria-pressed');
  });

  it('renders a joker discard top as a card image with an accessible label', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      discardPile: [{ design: 'JOKER', value: 0 }],
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/shithead']}>
        <ShitheadPage />
      </MemoryRouter>,
    );
    // The discard top is now a card image; the joker img carries an accessible 'ジョーカー' alt.
    await waitFor(() => expect(screen.getAllByAltText('ジョーカー').length).toBeGreaterThan(0));
    expect(screen.queryByText(/\?0/)).not.toBeInTheDocument();
  });
});
