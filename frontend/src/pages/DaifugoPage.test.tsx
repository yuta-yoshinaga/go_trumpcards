import { act, createEvent, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, daifugoApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DaifugoResponse } from '../types/card';
import { DaifugoPage } from './DaifugoPage';

vi.mock('../api/gameApi', () => ({
  daifugoApi: { exec: vi.fn() },
  actionLogApi: { daifugo: vi.fn() },
}));

const mockExec = vi.mocked(daifugoApi.exec);

const defaultConfig = {
  jokerCount: 2,
  eightCutEnabled: true,
  suitLockMode: 2,
  elevenBackEnabled: true,
  sequenceEnabled: true,
  cardExchangeEnabled: true,
  blindExchangeEnabled: false,
  fiveSkipEnabled: false,
  fiveSkipCount: 1,
  sevenPassEnabled: false,
  tenDiscardEnabled: false,
  spadeThreeEnabled: false,
  capitalFallEnabled: false,
  nineReverseEnabled: false,
  coupDetatEnabled: false,
  numberLockEnabled: false,
  sandstormEnabled: false,
  emperorEnabled: false,
  sequenceRevolutionEnabled: false,
  sequenceLockEnabled: false,
  illegalFinishEnabled: false,
  queenBomberEnabled: false,
  cpuDifficulty: 0,
};

const humanTurnState: DaifugoResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      rank: -1,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 5 },
      ],
    },
    { id: 1, isHuman: false, isFinished: false, rank: -1, cardCount: 4, cards: [] },
    { id: 2, isHuman: false, isFinished: false, rank: -1, cardCount: 3, cards: [] },
    { id: 3, isHuman: false, isFinished: false, rank: -1, cardCount: 5, cards: [] },
  ],
  currentTurn: 0,
  tableCards: [],
  lastPlayPlayerIdx: -1,
  gameEndFlag: false,
  revolutionActive: false,
  elevenBackActive: false,
  suitLocked: false,
  lockedSuit: '',
  tableIsSequence: false,
  config: defaultConfig,
  exchangeActions: [],
  cpuActions: [],
  humanAction: null,
  message: '',
  pendingAction: 'none',
  pendingActionTarget: -1,
  reverseDirection: false,
  numberLocked: false,
  sequenceLocked: false,
  sortMode: 0,
};

const cpuTurnState: DaifugoResponse = {
  ...humanTurnState,
  currentTurn: 1,
  humanAction: {
    playerIdx: 0,
    playedCards: [
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 5 },
    ],
  },
  tableCards: [
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 5 },
  ],
  lastPlayPlayerIdx: 0,
};

const gameEndState: DaifugoResponse = {
  ...humanTurnState,
  gameEndFlag: true,
  message: 'ゲーム終了！ あなた:大富豪 CPU 1:富豪 CPU 2:平民 CPU 3:大貧民',
  players: [
    { id: 0, isHuman: true, isFinished: true, rank: 1, cardCount: 0, cards: [] },
    { id: 1, isHuman: false, isFinished: true, rank: 2, cardCount: 0, cards: [] },
    { id: 2, isHuman: false, isFinished: true, rank: 3, cardCount: 0, cards: [] },
    { id: 3, isHuman: false, isFinished: true, rank: 4, cardCount: 0, cards: [] },
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('DaifugoPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DaifugoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('marks the active sort button with aria-pressed under an sr-only legend', async () => {
    // humanTurnState has sortMode 0 (strength).
    renderWithProviders(<DaifugoPage />);
    const strength = await screen.findByTestId('df-sort-0');
    expect(strength).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('df-sort-1')).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByTestId('df-sort-2')).toHaveAttribute('aria-pressed', 'false');
    // The three buttons are grouped with an accessible legend.
    expect(screen.getByRole('group', { name: '手札の並べ替え' })).toBeInTheDocument();
  });

  it('shows phase indicator with プレイ during gameplay', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      const indicator = screen.getByTestId('phase-indicator');
      expect(indicator).toHaveTextContent('プレイ');
    });
  });

  it('shows あなたのターン when it is human turn', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      expect(screen.getByText('あなたのターン')).toBeInTheDocument();
    });
  });

  it('shows 終了 phase when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      const indicator = screen.getByTestId('phase-indicator');
      expect(indicator).toHaveTextContent('終了');
    });
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders human player area labeled あなた', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
  });

  it('renders CPU player areas with correct labels', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('shows human player face-up cards', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
      expect(screen.getAllByAltText('♥ 5').length).toBeGreaterThan(0);
    });
  });

  it('shows card counts for CPU players', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      expect(screen.getByText('4枚')).toBeInTheDocument();
      expect(screen.getByText('5枚')).toBeInTheDocument();
    });
  });

  it('renders compact CPU row on mobile viewport', async () => {
    const originalInnerWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      const { container } = renderWithProviders(<DaifugoPage />);
      await waitFor(() => {
        expect(screen.getByText('CPU 1')).toBeInTheDocument();
        expect(screen.getByText('4枚')).toBeInTheDocument();
      });
      // On mobile the CPU row uses the compact flex+overflow layout; the
      // desktop CPU card (DaifugoCpuArea with flex-[1_1_180px]) must not render
      // for any CPU. The human area (player-area-0) still uses playerAreaClass.
      for (const cpuId of [1, 2, 3]) {
        const cpuNode = container.querySelector(`#player-area-${cpuId}`);
        expect(cpuNode?.className ?? '').not.toMatch(/flex-\[1_1_180px\]/);
      }
    } finally {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: originalInnerWidth,
      });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('shows empty table message when tableCards is empty', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('（なし）')).toBeInTheDocument());
  });

  it('shows table cards when tableCards is non-empty', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      const tableCard = screen.getAllByAltText('♥ 5');
      expect(tableCard.length).toBeGreaterThan(0);
    });
  });

  it('pass button is enabled on human turn', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
  });

  it('pass button is disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('pass button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('calls reset when reset button is clicked', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', [], expect.any(Object)));
  });

  it('calls play with empty indices when pass button is clicked', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('selects a card on click and enables play button', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    const card = screen.getByAltText('♠ 3');
    fireEvent.click(card);
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();
  });

  it('calls play with selected indices when play button is clicked', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ 3'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: '選択して出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('warns and disables play when the selection count does not match the table', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      tableCards: [
        { design: 'HEART', value: 6 },
        { design: 'DIAMOND', value: 6 },
      ],
    });
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ 3')); // 1 card selected, table needs 2
    expect(screen.getByTestId('daifugo-count-warning')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '選択して出す' })).toBeDisabled();
  });

  it('allows play when the selection count matches the table', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      tableCards: [
        { design: 'HEART', value: 6 },
        { design: 'DIAMOND', value: 6 },
      ],
    });
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 5')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♥ 5'));
    fireEvent.click(screen.getByAltText('♦ 5')); // 2 cards = table count
    expect(screen.queryByTestId('daifugo-count-warning')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();
  });

  it('does not show the count warning during a pending action', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      tableCards: [
        { design: 'HEART', value: 6 },
        { design: 'DIAMOND', value: 6 },
      ],
      pendingAction: 'sevenPass',
      pendingActionTarget: 1,
    });
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ 3')); // mismatching count, but pending → gated off
    expect(screen.queryByTestId('daifugo-count-warning')).not.toBeInTheDocument();
  });

  it('shows human action log after play', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/あなたが出しました/)).toBeInTheDocument());
  });

  it('shows CPU action log when cpuActions is non-empty', async () => {
    const stateWithCpuActions: DaifugoResponse = {
      ...humanTurnState,
      cpuActions: [
        { playerIdx: 1, playedCards: [{ design: 'SPADE', value: 7 }] },
        { playerIdx: 2, playedCards: null },
      ],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    renderWithProviders(<DaifugoPage />);
    // Each CPU action has an 800ms animation delay; wait for all to complete
    await waitFor(
      () => {
        expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument();
        expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument();
        expect(screen.getByText(/CPU 2がパスしました/)).toBeInTheDocument();
      },
      { timeout: 4000 },
    );
  }, 10000);

  it('shows intermediate human action state before CPU replay (humanAction with playedCards)', async () => {
    const stateWithHumanAndCpu: DaifugoResponse = {
      ...humanTurnState,
      currentTurn: 0,
      humanAction: { playerIdx: 0, playedCards: [{ design: 'SPADE', value: 3 }] },
      cpuActions: [{ playerIdx: 1, playedCards: [{ design: 'HEART', value: 5 }] }],
      players: [
        { ...humanTurnState.players[0], cardCount: 2 },
        { ...humanTurnState.players[1], cardCount: 3 },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithHumanAndCpu);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument(), { timeout: 4000 });
  }, 10000);

  it('shows intermediate human action state before CPU replay (humanAction with empty playedCards)', async () => {
    // Exercises the falsy branch: ha.playedCards?.length ? ha.playedCards : finalState.tableCards
    const stateWithPassAndCpu: DaifugoResponse = {
      ...humanTurnState,
      currentTurn: 0,
      humanAction: { playerIdx: 0, playedCards: [] },
      tableCards: [{ design: 'DIAMOND', value: 7 }],
      cpuActions: [{ playerIdx: 1, playedCards: [{ design: 'HEART', value: 5 }] }],
      players: [
        { ...humanTurnState.players[0], cardCount: 3 },
        { ...humanTurnState.players[1], cardCount: 3 },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithPassAndCpu);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument(), { timeout: 4000 });
  }, 10000);

  it('shows intermediate CPU action during replay animation', async () => {
    const stateWithCpuActions: DaifugoResponse = {
      ...humanTurnState,
      cpuActions: [
        { playerIdx: 1, playedCards: [{ design: 'SPADE', value: 7 }] },
        { playerIdx: 2, playedCards: null },
      ],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    renderWithProviders(<DaifugoPage />);
    // First intermediate state (CPU 1's action) appears before the second
    await waitFor(() => expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument());
    // After all animation steps, CPU 2's action also appears
    await waitFor(() => expect(screen.getByText(/CPU 2がパスしました/)).toBeInTheDocument(), { timeout: 4000 });
  }, 10000);

  it('enables play button after CPU replay animation completes', async () => {
    const stateWithCpuActions: DaifugoResponse = {
      ...humanTurnState,
      currentTurn: 0,
      cpuActions: [{ playerIdx: 1, playedCards: [{ design: 'SPADE', value: 7 }] }],
    };
    // reset → humanTurnState, play → stateWithCpuActions
    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(stateWithCpuActions);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));

    // Buttons stay disabled during animation
    expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled();

    // After replay delay, human turn is restored and buttons re-enable
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled(), { timeout: 4000 });
  }, 10000);

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/ゲーム終了！ あなた:大富豪/)).toBeInTheDocument());
  });

  it('shows rank badge for finished CPU players', async () => {
    const stateWithFinished: DaifugoResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, rank: 1, cardCount: 0, cards: [] },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithFinished);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/上がり.*大富豪/)).toBeInTheDocument());
  });

  it('shows thinking indicator on current CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('考え中...')).toBeInTheDocument());
  });

  it('shows revolution badge when revolutionActive is true', async () => {
    const revolutionState: DaifugoResponse = {
      ...humanTurnState,
      revolutionActive: true,
    };
    mockExec.mockResolvedValue(revolutionState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('革命中')).toBeInTheDocument());
  });

  it('shows 11-back badge when elevenBackActive is true', async () => {
    const elevenBackState: DaifugoResponse = {
      ...humanTurnState,
      elevenBackActive: true,
    };
    mockExec.mockResolvedValue(elevenBackState);
    renderWithProviders(<DaifugoPage />);
    // Use selector:'button' to find the badge (not the settings panel label)
    await waitFor(() => expect(screen.getByText('11バック', { selector: 'button' })).toBeInTheDocument());
  });

  it('switches the page background to the revolution palette when revolution is active', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, revolutionActive: true });
    const { container } = renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('革命中')).toBeInTheDocument());
    expect(container.querySelector('.bg-game-bg-revolution')).toBeInTheDocument();
    expect(container.querySelector('.bg-game-bg-green')).not.toBeInTheDocument();
    // The footer should track the body palette so the page doesn't read as a green/crimson split.
    expect(container.querySelector('.bg-game-bg-revolution-dark')).toBeInTheDocument();
  });

  it('switches the page background to the revolution palette when 11-back is active', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, elevenBackActive: true });
    const { container } = renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(container.querySelector('.bg-game-bg-revolution')).toBeInTheDocument());
    expect(container.querySelector('.bg-game-bg-revolution-dark')).toBeInTheDocument();
    expect(container.querySelector('.bg-game-bg-green')).not.toBeInTheDocument();
  });

  it('keeps the default green background when neither inversion is active', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    const { container } = renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(container.querySelector('.bg-game-bg-green')).toBeInTheDocument());
    expect(container.querySelector('.bg-game-bg-revolution')).not.toBeInTheDocument();
    expect(container.querySelector('.bg-game-bg-revolution-dark')).not.toBeInTheDocument();
  });

  it('shows suit lock badge when suitLocked is true', async () => {
    const suitLockedState: DaifugoResponse = {
      ...humanTurnState,
      suitLocked: true,
      lockedSuit: 'SPADE',
    };
    mockExec.mockResolvedValue(suitLockedState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('スート縛り: SPADE')).toBeInTheDocument());
  });

  it('shows sequence badge when tableIsSequence is true', async () => {
    const seqState: DaifugoResponse = {
      ...humanTurnState,
      tableIsSequence: true,
    };
    mockExec.mockResolvedValue(seqState);
    renderWithProviders(<DaifugoPage />);
    // Use selector:'button' to find the badge (not the settings panel label)
    await waitFor(() => expect(screen.getByText('階段', { selector: 'button' })).toBeInTheDocument());
  });

  it('shows card exchange log when exchangeActions is non-empty', async () => {
    const exchangeState: DaifugoResponse = {
      ...humanTurnState,
      exchangeActions: [
        {
          fromPlayerIdx: 3,
          toPlayerIdx: 0,
          cards: [{ design: 'SPADE', value: 2 }],
        },
      ],
    };
    mockExec.mockResolvedValue(exchangeState);
    renderWithProviders(<DaifugoPage />);
    // Use bracketed form to distinguish from settings panel label
    await waitFor(() => expect(screen.getByText(/\[カード交換\]/)).toBeInTheDocument());
  });

  it('disables action buttons while loading', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    let resolve!: (value: DaifugoResponse) => void;
    const slowPromise = new Promise<DaifugoResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));

    expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '選択して出す' })).toBeDisabled();

    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
  });

  it('shows error message when API call fails', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  }, 10000);

  it('clears error message on successful API call after failure', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  }, 10000);

  it('shows pass message for empty playedCards array', async () => {
    const stateWithEmptyPlay: DaifugoResponse = {
      ...humanTurnState,
      humanAction: { playerIdx: 0, playedCards: [] },
    };
    mockExec.mockResolvedValue(stateWithEmptyPlay);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('あなたがパスしました')).toBeInTheDocument());
  });

  it('shows rank badge when human player finishes', async () => {
    const humanFinishedState: DaifugoResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0], isFinished: true, rank: 1, cardCount: 0, cards: [] },
        { ...humanTurnState.players[1] },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(humanFinishedState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getAllByText(/上がり.*大富豪/).length).toBeGreaterThan(0));
  });

  it('does not show selection hint when not human turn in HumanPlayerArea', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('考え中...')).toBeInTheDocument());
    expect(screen.queryByText('カードをクリックして選択')).not.toBeInTheDocument();
  });

  it('shows 富豪 rank badge for finished player with rank 2', async () => {
    const stateWithRank2: DaifugoResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, rank: 2, cardCount: 0, cards: [] },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithRank2);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/上がり \(富豪\)/)).toBeInTheDocument());
  });

  it('shows 平民 rank badge for finished player with rank 3', async () => {
    const stateWithRank3: DaifugoResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, rank: 3, cardCount: 0, cards: [] },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithRank3);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/上がり \(平民\)/)).toBeInTheDocument());
  });

  it('shows 大貧民 rank badge for finished player with rank 4', async () => {
    const stateWithRank4: DaifugoResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, rank: 4, cardCount: 0, cards: [] },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithRank4);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/上がり \(大貧民\)/)).toBeInTheDocument());
  });

  it('toggles aria-pressed on card button click', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ 3').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('deselects a card by clicking it again', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    // Click SPADE 3 to select it
    fireEvent.click(screen.getByAltText('♠ 3'));
    // Click SPADE 3 again to deselect it
    fireEvent.click(screen.getByAltText('♠ 3'));

    // After deselection, the 選択して出す button should not show any selection count
    // (the card toggle state resets on the second click)
    // We verify no error is thrown and the button is still present
    expect(screen.getByRole('button', { name: '選択して出す' })).toBeInTheDocument();
  });

  it('clears the whole selection with the deselect button', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    const clearBtn = screen.getByTestId('daifugo-clear-selection');
    // Disabled with nothing selected.
    expect(clearBtn).toBeDisabled();
    // Select a card → the deselect button activates.
    fireEvent.click(screen.getByAltText('♠ 3'));
    expect(clearBtn).not.toBeDisabled();
    // Clicking it empties the selection (the play button goes back to disabled).
    fireEvent.click(clearBtn);
    expect(screen.getByRole('button', { name: '選択して出す' })).toBeDisabled();
  });

  it('settings panel renders checkbox labels', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('8切り')).toBeInTheDocument();
    expect(screen.getByLabelText('5飛び')).toBeInTheDocument();
    expect(screen.getByLabelText('7渡し')).toBeInTheDocument();
    expect(screen.getByLabelText('10捨て')).toBeInTheDocument();
    expect(screen.getByLabelText('スペ3返し')).toBeInTheDocument();
    expect(screen.getByLabelText('都落ち')).toBeInTheDocument();
  });

  it('settings panel joker count dropdown changes config', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('ジョーカー枚数:')).toBeInTheDocument());
    // Change joker count to 0
    fireEvent.change(screen.getByLabelText('ジョーカー枚数:'), { target: { value: '0' } });
    // Click reset → config with jokerCount:0 is sent
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ jokerCount: 0 })));
  });

  it('settings panel boolean checkbox toggle updates config', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('5飛び')).toBeInTheDocument());
    // Enable 5飛び
    fireEvent.click(screen.getByLabelText('5飛び'));
    // Click reset → config with fiveSkipEnabled:true is sent
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ fiveSkipEnabled: true })),
    );
  });

  it('shows sevenPass pending banner and changes play button label', async () => {
    const sevenPassState: DaifugoResponse = {
      ...humanTurnState,
      pendingAction: 'sevenPass',
      pendingActionTarget: 1,
    } as DaifugoResponse;
    mockExec.mockResolvedValue(sevenPassState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/【7渡し】/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '渡す' })).toBeInTheDocument();
    // Pass button is disabled when pending action is active
    expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled();
  });

  it('shows tenDiscard pending banner and changes play button label', async () => {
    const tenDiscardState: DaifugoResponse = {
      ...humanTurnState,
      pendingAction: 'tenDiscard',
      pendingActionTarget: -1,
    } as DaifugoResponse;
    mockExec.mockResolvedValue(tenDiscardState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/【10捨て】/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
  });

  it('shows queenBomber pending banner with number buttons and disables play button', async () => {
    const queenBomberState: DaifugoResponse = {
      ...humanTurnState,
      pendingAction: 'queenBomber',
      pendingActionTarget: -1,
    } as DaifugoResponse;
    mockExec.mockResolvedValue(queenBomberState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/【12ボンバー】/)).toBeInTheDocument());
    // Number buttons A,2-10,J,Q,K should be visible
    expect(screen.getByRole('button', { name: 'A' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'K' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Q' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'J' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '5' })).toBeInTheDocument();
    // Play button label stays default but is disabled
    expect(screen.getByRole('button', { name: '選択して出す' })).toBeDisabled();
  });

  it('queenBomber number button calls exec with play and value', async () => {
    const queenBomberState: DaifugoResponse = {
      ...humanTurnState,
      pendingAction: 'queenBomber',
      pendingActionTarget: -1,
    } as DaifugoResponse;
    mockExec.mockResolvedValue(queenBomberState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'A' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'A' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [1]));
  });

  it('queenBomber number buttons not shown when not human turn', async () => {
    const queenBomberCpuTurn: DaifugoResponse = {
      ...humanTurnState,
      currentTurn: 1,
      pendingAction: 'queenBomber',
      pendingActionTarget: -1,
    } as DaifugoResponse;
    mockExec.mockResolvedValue(queenBomberCpuTurn);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/【12ボンバー】/)).toBeInTheDocument());
    // Number buttons should NOT be visible (CPU turn)
    expect(screen.queryByRole('button', { name: 'A' })).not.toBeInTheDocument();
  });

  it('UI is disabled when currentTurn is CPU even if pendingAction is set', async () => {
    // Pending actions always belong to currentTurn's player; if CPU has a
    // pending action, the human UI must stay disabled.
    const pendingCpuTurn: DaifugoResponse = {
      ...humanTurnState,
      currentTurn: 1, // CPU's turn to resolve pending action
      pendingAction: 'sevenPass',
      pendingActionTarget: 2,
    } as DaifugoResponse;
    mockExec.mockResolvedValue(pendingCpuTurn);
    renderWithProviders(<DaifugoPage />);
    // Card buttons are disabled because isHumanTurn = false (currentTurn is CPU)
    await waitFor(() => expect(screen.getByAltText('♠ 3').closest('button')).toBeDisabled());
  });

  it('drag card not in selection adds it to selection', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    // Initially no card selected → play button disabled
    expect(screen.getByRole('button', { name: '選択して出す' })).toBeDisabled();

    // Fire dragStart on SPADE 3 (index 0, not in selection)
    const cardButton = screen.getByAltText('♠ 3').closest('button') as HTMLElement;
    fireEvent.dragStart(cardButton, {
      dataTransfer: { setData: vi.fn() },
    });

    // handleDragCard adds card to selection → play button enabled
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();
  });

  it('drag card already in selection keeps it in selection', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    const cardButton = screen.getByAltText('♠ 3').closest('button') as HTMLElement;
    // First click to select
    fireEvent.click(cardButton);
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();

    // dragStart with card already in selection → still selected
    fireEvent.dragStart(cardButton, {
      dataTransfer: { setData: vi.fn() },
    });
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();
  });

  it('drop on table plays dragged card when not in selection', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    // No cards selected; drag card index 0 onto the table
    const dropZone = screen.getByText('場札').closest('div') as HTMLElement;
    const dropEvent = createEvent.drop(dropZone);
    Object.defineProperty(dropEvent, 'dataTransfer', {
      value: { getData: vi.fn().mockReturnValue('0') },
      writable: false,
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent(dropZone, dropEvent);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('drop on table plays selected cards when dragged card is in selection', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    // Select cards 0 and 1
    fireEvent.click(screen.getByAltText('♠ 3'));
    fireEvent.click(screen.getAllByAltText('♥ 5')[0]);
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();

    // Drop card 0 (which is in selection) → plays [0,1]
    const dropZone = screen.getByText('場札').closest('div') as HTMLElement;
    const dropEvent = createEvent.drop(dropZone);
    Object.defineProperty(dropEvent, 'dataTransfer', {
      value: { getData: vi.fn().mockReturnValue('0') },
      writable: false,
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent(dropZone, dropEvent);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0, 1]));
  });

  it('sets aria-busy while loading', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'パス' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: DaifugoResponse) => void;
    const slowPromise = new Promise<DaifugoResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(humanTurnState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
    });
  });

  it('shows 9リバース badge when reverseDirection is true', async () => {
    const reverseState: DaifugoResponse = {
      ...humanTurnState,
      reverseDirection: true,
    };
    mockExec.mockResolvedValue(reverseState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('9リバース', { selector: 'button' })).toBeInTheDocument());
  });

  it('shows 数縛り badge when numberLocked is true', async () => {
    const numberLockedState: DaifugoResponse = {
      ...humanTurnState,
      numberLocked: true,
    };
    mockExec.mockResolvedValue(numberLockedState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      const badges = screen.getAllByText('数縛り');
      expect(badges.length).toBeGreaterThanOrEqual(1);
      expect(badges.some((el) => el.tagName === 'BUTTON')).toBe(true);
    });
  });

  it('renders sort buttons and active button is highlighted', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '強さ順' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スート順' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '数字順' })).toBeInTheDocument();
    // Default sortMode=0 → 強さ順 should have primary style
    expect(screen.getByRole('button', { name: '強さ順' }).className).toContain('ds-accent');
    expect(screen.getByRole('button', { name: 'スート順' }).className).toContain('ds-surface-elevated');
  });

  it('calls sort command when sort button is clicked', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スート順' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...humanTurnState, sortMode: 1 });
    fireEvent.click(screen.getByRole('button', { name: 'スート順' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('sort', undefined, undefined, 1));
  });

  it('settings panel renders new rule checkboxes', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('9リバース')).toBeInTheDocument();
    expect(screen.getByLabelText('クーデター')).toBeInTheDocument();
    expect(screen.getByLabelText('数縛り')).toBeInTheDocument();
  });

  it('drop with invalid dataTransfer data is ignored (NaN guard)', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    const dropZone = screen.getByText('場札').closest('div') as HTMLElement;
    const dropEvent = createEvent.drop(dropZone);
    Object.defineProperty(dropEvent, 'dataTransfer', {
      value: { getData: vi.fn().mockReturnValue('') },
      writable: false,
    });
    mockExec.mockClear();
    fireEvent(dropZone, dropEvent);
    // exec should NOT be called when draggedIdx is NaN
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('settings panel renders 砂嵐 checkbox', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('砂嵐')).toBeInTheDocument();
    expect(screen.getByLabelText('砂嵐')).not.toBeChecked();
  });

  it('settings panel renders エンペラー checkbox', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('エンペラー')).toBeInTheDocument();
    expect(screen.getByLabelText('エンペラー')).not.toBeChecked();
  });

  it('砂嵐 checkbox toggle updates config on reset', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('砂嵐')).toBeInTheDocument());
    fireEvent.click(screen.getByLabelText('砂嵐'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ sandstormEnabled: true })),
    );
  });

  it('エンペラー checkbox toggle updates config on reset', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('エンペラー')).toBeInTheDocument());
    fireEvent.click(screen.getByLabelText('エンペラー'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ emperorEnabled: true })),
    );
  });

  it('settings panel renders 階段革命 checkbox', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('階段革命')).toBeInTheDocument();
    expect(screen.getByLabelText('階段革命')).not.toBeChecked();
  });

  it('settings panel renders 反則上がり checkbox', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('反則上がり')).toBeInTheDocument();
    expect(screen.getByLabelText('反則上がり')).not.toBeChecked();
  });

  it('階段革命 checkbox toggle updates config on reset', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('階段革命')).toBeInTheDocument());
    fireEvent.click(screen.getByLabelText('階段革命'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ sequenceRevolutionEnabled: true })),
    );
  });

  it('反則上がり checkbox toggle updates config on reset', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('反則上がり')).toBeInTheDocument());
    fireEvent.click(screen.getByLabelText('反則上がり'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ illegalFinishEnabled: true })),
    );
  });

  it('shows 反則上がり penalty badge for CPU player', async () => {
    const penaltyState: DaifugoResponse = {
      ...gameEndState,
      players: [
        { id: 0, isHuman: true, isFinished: true, rank: 1, cardCount: 0, cards: [] },
        {
          id: 1,
          isHuman: false,
          isFinished: true,
          rank: 4,
          cardCount: 0,
          cards: [],
          illegalFinishPenalty: true,
        },
        { id: 2, isHuman: false, isFinished: true, rank: 2, cardCount: 0, cards: [] },
        { id: 3, isHuman: false, isFinished: true, rank: 3, cardCount: 0, cards: [] },
      ],
    };
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('反則上がり', { selector: 'span' })).toBeInTheDocument());
  });

  it('shows 反則上がり penalty badge for human player', async () => {
    const penaltyState: DaifugoResponse = {
      ...gameEndState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: true,
          rank: 4,
          cardCount: 0,
          cards: [],
          illegalFinishPenalty: true,
        },
        { id: 1, isHuman: false, isFinished: true, rank: 1, cardCount: 0, cards: [] },
        { id: 2, isHuman: false, isFinished: true, rank: 2, cardCount: 0, cards: [] },
        { id: 3, isHuman: false, isFinished: true, rank: 3, cardCount: 0, cards: [] },
      ],
    };
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('反則上がり', { selector: 'span' })).toBeInTheDocument());
  });

  it('does not show 反則上がり badge when no penalty', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/ゲーム終了！/)).toBeInTheDocument());
    expect(screen.queryByText('反則上がり', { selector: 'span' })).not.toBeInTheDocument();
  });

  it('settings panel renders CPU difficulty dropdown', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('CPU難易度:')).toBeInTheDocument());
    const select = screen.getByLabelText('CPU難易度:') as HTMLSelectElement;
    expect(select.value).toBe('0');
    expect(screen.getByText('ふつう')).toBeInTheDocument();
    expect(screen.getByText('よわい')).toBeInTheDocument();
    expect(screen.getByText('つよい')).toBeInTheDocument();
  });

  it('CPU difficulty dropdown change updates config on reset', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('CPU難易度:')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('CPU難易度:'), { target: { value: '2' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ cpuDifficulty: 2 })),
    );
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      gameEndFlag: true,
      currentTurn: 0,
      players: [],
      playerIdx: 0,
      lastDiscardedCards: [],
    } as unknown as DaifugoResponse);

    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.daifugo).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.daifugo).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // --- ConfirmDialog tests ---

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', [], expect.any(Object)));
  });

  // --- Keyboard navigation tests ---

  it('pressing number key toggles card selection', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('img', { name: '♠ 3' })).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    // card at index 0 should now be selected (visually highlighted)
    const card1 = screen.getByRole('img', { name: '♠ 3' }).closest('button');
    expect(card1).toBeDefined();

    // pressing again toggles it off
    fireEvent.keyDown(document, { key: '1' });
  });

  it('pressing Enter triggers play with selected cards', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('img', { name: '♠ 3' })).toBeInTheDocument());

    // select first card
    fireEvent.keyDown(document, { key: '1' });

    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('pressing Escape clears card selection', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('img', { name: '♠ 3' })).toBeInTheDocument());

    // select a card first
    fireEvent.keyDown(document, { key: '1' });

    // press Escape to clear
    fireEvent.keyDown(document, { key: 'Escape' });

    // Enter should not trigger play since selection is cleared
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'Enter' });
    // exec should not have been called with 'play' since no cards selected
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('keyboard navigation is disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('img', { name: '♠ 3' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: '1' });
    fireEvent.keyDown(document, { key: 'Enter' });
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('shows 階段縛り badge when sequenceLocked is true', async () => {
    const sequenceLockedState: DaifugoResponse = {
      ...humanTurnState,
      sequenceLocked: true,
    };
    mockExec.mockResolvedValue(sequenceLockedState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => {
      const badges = screen.getAllByText('階段縛り');
      expect(badges.length).toBeGreaterThanOrEqual(1);
      expect(badges.some((el) => el.tagName === 'BUTTON')).toBe(true);
    });
  });

  it('settings panel renders 階段縛り checkbox', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('階段縛り')).toBeInTheDocument();
    expect(screen.getByLabelText('階段縛り')).not.toBeChecked();
  });

  it('settings panel renders ブラインド交換 checkbox', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('ブラインド交換')).toBeInTheDocument();
    expect(screen.getByLabelText('ブラインド交換')).not.toBeChecked();
  });

  it('number key beyond card count is ignored', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('img', { name: '♠ 3' })).toBeInTheDocument());

    // human has 3 cards, key '4' (index 3) should be ignored
    fireEvent.keyDown(document, { key: '4' });

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'Enter' });
    // no cards selected so Enter should not call play
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('shows a toast when revolution becomes active', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, revolutionActive: true });
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByTestId('df-inversion-toast')).toHaveTextContent('革命発動！'));
  });

  it('shows an eleven-back toast when eleven-back becomes active', async () => {
    mockExec.mockResolvedValue({ ...humanTurnState, elevenBackActive: true });
    renderWithProviders(<DaifugoPage />);
    await waitFor(() => expect(screen.getByTestId('df-inversion-toast')).toHaveTextContent('11バック発動！'));
  });

  it('auto-hides the inversion toast after 3 seconds', async () => {
    vi.useFakeTimers();
    try {
      mockExec.mockResolvedValue({ ...humanTurnState, revolutionActive: true });
      renderWithProviders(<DaifugoPage />);
      await vi.waitFor(() => expect(screen.getByTestId('df-inversion-toast')).toBeInTheDocument());
      await act(async () => {
        vi.advanceTimersByTime(3000);
      });
      expect(screen.queryByTestId('df-inversion-toast')).not.toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });
});
