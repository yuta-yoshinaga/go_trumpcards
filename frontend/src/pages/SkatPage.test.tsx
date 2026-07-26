import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { skatApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SkatResponse } from '../types/card';
import { SkatGameType, SkatPhase } from '../types/phases';
import { SkatPage } from './SkatPage';

vi.mock('../api/gameApi', () => ({
  skatApi: { exec: vi.fn() },
  actionLogApi: { skat: vi.fn() },
}));

const mockExec = vi.mocked(skatApi.exec);

const basePlayer = (id: number, isHuman: boolean) => ({
  id,
  isHuman,
  cardCount: 10,
  cards: [] as Card[],
  bid: 0,
  isDeclarer: false,
  cardPoints: 0,
  roundsWon: 0,
  roundsLost: 0,
  roundScore: 0,
  cumulativeScore: 0,
  trickCount: 0,
});

const bidPhaseHumanTurn: SkatResponse = {
  players: [basePlayer(0, true), basePlayer(1, false), basePlayer(2, false)],
  phase: SkatPhase.BID,
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: -1,
  currentTrick: [],
  forehandIdx: 0,
  middlehandIdx: 1,
  rearhandIdx: 2,
  dealerIdx: 0,
  declarerIdx: -1,
  currentBid: 0,
  activeBidActorIdx: 0,
  gameType: SkatGameType.NONE,
  trumpSuit: 0,
  pickedSkat: false,
  declarerCardPoints: 0,
  defendersCardPoints: 0,
  winnerSide: -1,
  gameValue: 0,
  gameEndFlag: false,
  leadPlayerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, targetScore: 500 },
};

const bidPhaseCpuTurn: SkatResponse = {
  ...bidPhaseHumanTurn,
  activeBidActorIdx: 1,
};

const bidPhaseAcceptAt: SkatResponse = {
  ...bidPhaseHumanTurn,
  currentBid: 18,
};

const skatPickupPhase: SkatResponse = {
  ...bidPhaseHumanTurn,
  phase: SkatPhase.SKAT_PICKUP,
  declarerIdx: 0,
  currentBid: 18,
};

const discardPhase: SkatResponse = {
  ...bidPhaseHumanTurn,
  phase: SkatPhase.DISCARD,
  declarerIdx: 0,
  currentBid: 18,
  pickedSkat: true,
  players: [
    {
      ...basePlayer(0, true),
      cardCount: 12,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 13 },
      ],
    },
    basePlayer(1, false),
    basePlayer(2, false),
  ],
};

const gameDeclarationPhase: SkatResponse = {
  ...bidPhaseHumanTurn,
  phase: SkatPhase.GAME_DECLARATION,
  declarerIdx: 0,
  currentBid: 18,
};

const playPhaseHumanTurn: SkatResponse = {
  ...bidPhaseHumanTurn,
  phase: SkatPhase.PLAY,
  declarerIdx: 0,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentBid: 18,
  gameType: SkatGameType.SUIT,
  trumpSuit: 1, // Spade
  players: [
    {
      ...basePlayer(0, true),
      isDeclarer: true,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 13 },
      ],
    },
    basePlayer(1, false),
    basePlayer(2, false),
  ],
};

const playPhaseCpuTurn: SkatResponse = {
  ...playPhaseHumanTurn,
  currentPlayerIdx: 1,
};

const trickEndPhase: SkatResponse = {
  ...playPhaseHumanTurn,
  phase: SkatPhase.TRICK_END,
};

const roundEndPhase: SkatResponse = {
  ...playPhaseHumanTurn,
  phase: SkatPhase.ROUND_END,
  declarerCardPoints: 75,
  defendersCardPoints: 45,
  gameValue: 18,
  originalSkat: [
    { design: 'DIAMOND', value: 7 },
    { design: 'DIAMOND', value: 8 },
  ],
};

const gameEndPhase: SkatResponse = {
  ...playPhaseHumanTurn,
  phase: SkatPhase.GAME_END,
  gameEndFlag: true,
};

const grandGamePhase: SkatResponse = {
  ...playPhaseHumanTurn,
  gameType: SkatGameType.GRAND,
};

const nullGamePhase: SkatResponse = {
  ...playPhaseHumanTurn,
  gameType: SkatGameType.NULL,
};

const cpuDeclarerSuitGame: SkatResponse = {
  ...playPhaseHumanTurn,
  declarerIdx: 1,
};

beforeEach(() => {
  localStorage.clear();
  mockExec.mockResolvedValue(bidPhaseHumanTurn);
});

describe('SkatPage', () => {
  it('shows loading message before state arrives', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    expect(screen.getByText(/Loading|読み込み/i)).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetScore: 500 } }),
    );
  });

  // The phase key map must hold bare keys; usePhaseNames adds the `phase.`
  // prefix itself, so Skat's namespace-qualified keys resolved to literals
  // like "phase.skat.phase.skatPickup" on screen. The trick phases also had
  // no locale entry at all. See issue #4374.
  it.each([
    [bidPhaseHumanTurn, 'ビッドフェーズ'],
    [skatPickupPhase, 'スカート拾い'],
    [discardPhase, '捨て札'],
    [gameDeclarationPhase, 'ゲーム宣言'],
    [playPhaseHumanTurn, 'プレイ'],
    [trickEndPhase, 'トリック終了'],
    [roundEndPhase, 'ラウンド終了'],
  ])('renders the translated phase name, not the raw i18n key (%#)', async (state, expected) => {
    mockExec.mockResolvedValue(state);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(expected));
    expect(screen.getByTestId('phase-indicator')).not.toHaveTextContent('phase.');
  });

  it('renders the CPU difficulty selector in the settings panel', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '18でコールする' })).toBeInTheDocument());
    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/) as HTMLSelectElement;
    expect(select).toBeInTheDocument();
    expect(select.value).toBe('1');
  });

  it('reflects the selected CPU difficulty in the reset API call', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '18でコールする' })).toBeInTheDocument());

    const select = screen.getByLabelText(/CPU難易度|CPU Difficulty/);
    fireEvent.change(select, { target: { value: '2' } });

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'reset',
        expect.objectContaining({ config: expect.objectContaining({ cpuDifficulty: 2 }) }),
      ),
    );
  });

  it('resets via the API instead of reloading the page when the reset button is clicked', async () => {
    // Guard against a regression to window.location.reload(): spy on it and
    // assert it is never called, while the reset command drives a fresh game.
    const reloadSpy = vi.fn();
    const originalLocation = window.location;
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...originalLocation, reload: reloadSpy },
    });
    try {
      renderWithProviders(
        <MemoryRouter initialEntries={['/skat']}>
          <SkatPage />
        </MemoryRouter>,
      );
      await waitFor(() => expect(screen.getByRole('button', { name: '18でコールする' })).toBeInTheDocument());

      mockExec.mockClear();
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      fireEvent.click(screen.getByRole('button', { name: '確認' }));
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetScore: 500 } }),
      );
      expect(reloadSpy).not.toHaveBeenCalled();
    } finally {
      Object.defineProperty(window, 'location', {
        configurable: true,
        value: originalLocation,
      });
    }
  });

  it('renders bid controls with "call 18" button when no current bid', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '18でコールする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('renders accept-at button when there is an active bid', async () => {
    mockExec.mockResolvedValue(bidPhaseAcceptAt);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: /受ける\s*18/ })).toBeInTheDocument());
  });

  it('hides bid controls when active bidder is CPU', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuTurn);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getAllByText(/ラウンド/i).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '18でコールする' })).not.toBeInTheDocument();
  });

  it('dispatches bid command on accept click', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '18でコールする' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(skatPickupPhase);
    fireEvent.click(screen.getByRole('button', { name: '18でコールする' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { accept: true }));
  });

  it('dispatches bid pass on pass click', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseHumanTurn);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { accept: false }));
  });

  it('renders pickup buttons in skat-pickup phase', async () => {
    mockExec.mockResolvedValue(skatPickupPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'スカートを拾う' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ハンドゲーム' })).toBeInTheDocument();
  });

  it('dispatches pickskat on click', async () => {
    mockExec.mockResolvedValue(skatPickupPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'スカートを拾う' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhase);
    fireEvent.click(screen.getByRole('button', { name: 'スカートを拾う' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pickskat', { pickup: true }));
  });

  it('dispatches handGame (pickup=false) when hand-game button clicked', async () => {
    mockExec.mockResolvedValue(skatPickupPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'ハンドゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameDeclarationPhase);
    fireEvent.click(screen.getByRole('button', { name: 'ハンドゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pickskat', { pickup: false }));
  });

  it('discard button is disabled until two cards are selected', async () => {
    mockExec.mockResolvedValue(discardPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '選択した2枚を伏せる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '選択した2枚を伏せる' })).toBeDisabled();
  });

  it('renders all three game-declaration buttons in declaration phase', async () => {
    mockExec.mockResolvedValue(gameDeclarationPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'スートゲームを宣言' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'グランドを宣言' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ヌルを宣言' })).toBeInTheDocument();
  });

  it('dispatches game declaration with selected trump suit', async () => {
    mockExec.mockResolvedValue(gameDeclarationPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'スートゲームを宣言' })).toBeInTheDocument());
    const select = screen.getByLabelText('切り札スート');
    fireEvent.change(select, { target: { value: '3' } }); // hearts
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    fireEvent.click(screen.getByRole('button', { name: 'スートゲームを宣言' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('game', { gameType: SkatGameType.SUIT, trumpSuit: 3 }));
  });

  it('dispatches grand declaration without trump suit', async () => {
    mockExec.mockResolvedValue(gameDeclarationPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'グランドを宣言' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    fireEvent.click(screen.getByRole('button', { name: 'グランドを宣言' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('game', { gameType: SkatGameType.GRAND, trumpSuit: undefined }),
    );
  });

  it('dispatches null declaration without trump suit', async () => {
    mockExec.mockResolvedValue(gameDeclarationPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヌルを宣言' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    fireEvent.click(screen.getByRole('button', { name: 'ヌルを宣言' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('game', { gameType: SkatGameType.NULL, trumpSuit: undefined }),
    );
  });

  it('renders human cards during play phase', async () => {
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByAltText('♥ K')).toBeInTheDocument();
  });

  it('shows declarer info with suit-game label and trump symbol', async () => {
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/スートゲーム/)).toBeInTheDocument());
    expect(screen.getByText(/♠/)).toBeInTheDocument();
  });

  it('shows grand-game label when gameType is GRAND', async () => {
    mockExec.mockResolvedValue(grandGamePhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/グランド/)).toBeInTheDocument());
  });

  it('shows null-game label when gameType is NULL', async () => {
    mockExec.mockResolvedValue(nullGamePhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/ヌル/)).toBeInTheDocument());
  });

  it('shows CPU declarer label when declarer is CPU', async () => {
    mockExec.mockResolvedValue(cpuDeclarerSuitGame);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    // The string "CPU 1" appears both in the declarer header and in the
    // per-player score row, so use getAllByText.
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThanOrEqual(1));
  });

  it('play button disabled until exactly one card is selected', async () => {
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: 'カードを出す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'カードを出す' })).toBeDisabled();
  });

  it('play button enabled and dispatches when a card is selected', async () => {
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: 'カードを出す' })).not.toBeDisabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseCpuTurn);
    fireEvent.click(screen.getByRole('button', { name: 'カードを出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('hides play button when CPU is on turn', async () => {
    mockExec.mockResolvedValue(playPhaseCpuTurn);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getAllByText(/ラウンド/i).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: 'カードを出す' })).not.toBeInTheDocument();
  });

  it('shows next-trick button at trick end and dispatches', async () => {
    mockExec.mockResolvedValue(trickEndPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseHumanTurn);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows next-round button at round end and dispatches', async () => {
    mockExec.mockResolvedValue(roundEndPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseHumanTurn);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders skat (face-up) in round-end state when originalSkat is present', async () => {
    mockExec.mockResolvedValue(roundEndPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/スカート \(場札\)/)).toBeInTheDocument());
    // The two skat cards now render as real card images, not raw "DIAMOND 7" text.
    const reveal = screen.getByTestId('skat-reveal');
    expect(within(reveal).getAllByRole('img')).toHaveLength(2);
    expect(reveal).not.toHaveTextContent('DIAMOND');
  });

  it('hides next-round button when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndPhase);
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getAllByText(/ラウンド/i).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '次のラウンド' })).not.toBeInTheDocument();
  });

  it('renders bid-estimate during bid phase without an exceeds warning when the current bid is 0', async () => {
    mockExec.mockResolvedValue({
      ...bidPhaseHumanTurn,
      players: [
        {
          ...basePlayer(0, true),
          cards: [{ design: 'CLOVER', value: 11 } as Card],
        },
        basePlayer(1, false),
        basePlayer(2, false),
      ],
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    const estimate = await screen.findByTestId('bid-estimate');
    expect(estimate).not.toHaveAttribute('data-exceeds');
    // J♣ alone → Grand "with 1" → value 48; the label should mention that number.
    expect(estimate.textContent).toContain('48');
  });

  it('flags the bid-estimate as exceeded when the current bid is above the hand ceiling', async () => {
    mockExec.mockResolvedValue({
      ...bidPhaseHumanTurn,
      currentBid: 999,
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/skat']}>
        <SkatPage />
      </MemoryRouter>,
    );
    const estimate = await screen.findByTestId('bid-estimate');
    expect(estimate).toHaveAttribute('data-exceeds', 'true');
    expect(estimate.textContent).toContain('⚠️');
  });
});
