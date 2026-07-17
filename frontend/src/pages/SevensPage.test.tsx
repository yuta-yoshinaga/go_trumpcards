import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, sevensApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevensResponse } from '../types/card';
import { SevensPage } from './SevensPage';

vi.mock('../api/gameApi', () => ({
  sevensApi: { exec: vi.fn() },
  actionLogApi: { sevens: vi.fn() },
}));

const mockExec = vi.mocked(sevensApi.exec);

// tableMinVals/tableMaxVals: index 0 unused; 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND
// tablePlaced: bitmask per suit; bit i = value i placed. 7 placed = 1<<7 = 128
// With all 7s placed: value 6 or 8 of any suit is playable
const defaultConfig = {
  tunnelEnabled: false,
  tunnelSkipWidth: 0,
  jokerCount: 0,
  cpuStrategy: 0,
  maxPasses: 5,
  noJokerFinish: false,
  jokerReclaimEnabled: false,
  endStopEnabled: false,
  jokerConsecutiveBanned: false,
};
const allSevensPlaced = [0, 128, 128, 128, 128]; // bit 7 set = 128

const humanTurnState: SevensResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      rank: -1,
      cardCount: 3,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
      cards: [
        { design: 'SPADE', value: 6 }, // playable: 6 == 7-1
        { design: 'HEART', value: 5 }, // not playable
        { design: 'CLOVER', value: 8 }, // playable: 8 == 7+1
      ],
    },
    {
      id: 1,
      isHuman: false,
      isFinished: false,
      rank: -1,
      cardCount: 4,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
      cards: [],
    },
    {
      id: 2,
      isHuman: false,
      isFinished: false,
      rank: -1,
      cardCount: 3,
      passesUsed: 1,
      maxPasses: 5,
      lastPlayedJoker: false,
      cards: [],
    },
    {
      id: 3,
      isHuman: false,
      isFinished: false,
      rank: -1,
      cardCount: 5,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
      cards: [],
    },
  ],
  currentTurn: 0,
  tableMinVals: [0, 7, 7, 7, 7],
  tableMaxVals: [0, 7, 7, 7, 7],
  tablePlaced: allSevensPlaced,
  config: defaultConfig,
  gameEndFlag: false,
  cpuActions: [],
  humanAction: null,
  message: '',
};

const cpuTurnState: SevensResponse = {
  ...humanTurnState,
  currentTurn: 1,
  humanAction: {
    playerIdx: 0,
    playedCard: { design: 'SPADE', value: 6 },
    targetSuit: 0,
    targetValue: 0,
    forcedPass: false,
  },
  tableMinVals: [0, 6, 7, 7, 7],
  tableMaxVals: [0, 7, 7, 7, 7],
  tablePlaced: [0, 128 | 64, 128, 128, 128], // spade 6+7 placed
};

const gameEndState: SevensResponse = {
  ...humanTurnState,
  gameEndFlag: true,
  message: 'ゲーム終了！ あなた:1位 CPU 1:2位 CPU 2:3位 CPU 3:4位',
  players: [
    { ...humanTurnState.players[0], isFinished: true, rank: 1, cardCount: 0, cards: [] },
    {
      id: 1,
      isHuman: false,
      isFinished: true,
      rank: 2,
      cardCount: 0,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
      cards: [],
    },
    {
      id: 2,
      isHuman: false,
      isFinished: true,
      rank: 3,
      cardCount: 0,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
      cards: [],
    },
    {
      id: 3,
      isHuman: false,
      isFinished: true,
      rank: 4,
      cardCount: 0,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
      cards: [],
    },
  ],
};

const passesExhaustedState: SevensResponse = {
  ...humanTurnState,
  players: [{ ...humanTurnState.players[0], passesUsed: 5, maxPasses: 5 }, ...humanTurnState.players.slice(1)],
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('SevensPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SevensPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders human player area labeled あなた', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
  });

  it('renders CPU player areas with correct labels', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('renders the board section', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
  });

  it('shows human player face-up cards', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ 6')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 5')).toBeInTheDocument();
      expect(screen.getByAltText('♣ 8')).toBeInTheDocument();
    });
  });

  it('shows the number-key hint and per-card aria-keyshortcuts on the human turn', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('sevens-key-hints')).toBeInTheDocument());
    expect(screen.getByTestId('sevens-key-hints')).toHaveTextContent('数字キー: 対応する手札をプレイ');
    // ♠6 is hand index 0 → key "1"; ♣8 is index 2 → key "3" (♥5 at index 1 is not playable).
    expect(screen.getByAltText('♠ 6').closest('button')).toHaveAttribute('aria-keyshortcuts', '1');
    expect(screen.getByAltText('♣ 8').closest('button')).toHaveAttribute('aria-keyshortcuts', '3');
    expect(screen.getByAltText('♥ 5').closest('button')).not.toHaveAttribute('aria-keyshortcuts');
  });

  it('hides the number-key hint when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByTestId('sevens-key-hints')).not.toBeInTheDocument();
  });

  it('hides the number-key hint at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
    expect(screen.queryByTestId('sevens-key-hints')).not.toBeInTheDocument();
  });

  it('pass button is enabled on human turn with passes remaining', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled());
  });

  it('pass button is disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).toBeDisabled());
  });

  it('pass button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).toBeDisabled());
  });

  it('pass button is disabled when passes are exhausted', async () => {
    mockExec.mockResolvedValue(passesExhaustedState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).toBeDisabled());
  });

  it('calls play with -1 when pass button is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: /パス/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', -1));
  });

  it('calls play with card index when a playable card is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 6')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByAltText('♠ 6'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('does not call play when a non-playable card is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('♥ 5')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByAltText('♥ 5'));
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('calls reset when reset button is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        tunnelSkipWidth: 0,
        jokerCount: 0,
        cpuStrategy: 0,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: false,
        endStop: false,
        jokerConsecutiveBanned: false,
      }),
    );
  });

  it('shows human action log after play', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/あなたが出しました/)).toBeInTheDocument());
  });

  it('shows pass action log for human pass', async () => {
    const passState: SevensResponse = {
      ...cpuTurnState,
      humanAction: { playerIdx: 0, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: false },
    };
    mockExec.mockResolvedValue(passState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/あなたがパスしました/)).toBeInTheDocument());
  });

  it('shows CPU action log when cpuActions is non-empty', async () => {
    const stateWithCpuActions: SevensResponse = {
      ...humanTurnState,
      cpuActions: [
        { playerIdx: 1, playedCard: { design: 'SPADE', value: 8 }, targetSuit: 1, targetValue: 8, forcedPass: false },
        { playerIdx: 2, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: false },
      ],
      tablePlaced: [0, 128 | (1 << 8), 128, 128, 128], // SPADE: 7+8 placed
      tableMinVals: [0, 7, 7, 7, 7],
      tableMaxVals: [0, 8, 7, 7, 7],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    renderWithProviders(<SevensPage />);
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

  it('shows intermediate CPU action during replay animation', async () => {
    const stateWithCpuActions: SevensResponse = {
      ...humanTurnState,
      cpuActions: [
        { playerIdx: 1, playedCard: { design: 'SPADE', value: 8 }, targetSuit: 1, targetValue: 8, forcedPass: false },
        { playerIdx: 2, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: false },
      ],
      tablePlaced: [0, 128 | (1 << 8), 128, 128, 128],
      tableMinVals: [0, 7, 7, 7, 7],
      tableMaxVals: [0, 8, 7, 7, 7],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    renderWithProviders(<SevensPage />);
    // First intermediate state (CPU 1's action) appears immediately
    await waitFor(() => expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument());
    // After second animation step, CPU 2's action also appears
    await waitFor(() => expect(screen.getByText(/CPU 2がパスしました/)).toBeInTheDocument(), { timeout: 4000 });
  }, 10000);

  it('enables pass button after CPU replay animation completes', async () => {
    const stateWithCpuActions: SevensResponse = {
      ...humanTurnState,
      currentTurn: 0,
      cpuActions: [
        { playerIdx: 1, playedCard: { design: 'SPADE', value: 8 }, targetSuit: 1, targetValue: 8, forcedPass: false },
      ],
      tablePlaced: [0, 128 | (1 << 8), 128, 128, 128],
      tableMinVals: [0, 7, 7, 7, 7],
      tableMaxVals: [0, 8, 7, 7, 7],
    };
    // reset → humanTurnState, play → stateWithCpuActions
    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(stateWithCpuActions);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: /パス/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', -1));

    // Buttons stay disabled during animation
    expect(screen.getByRole('button', { name: /パス/ })).toBeDisabled();

    // After replay delay, human turn is restored and buttons re-enable
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled(), { timeout: 4000 });
  }, 10000);

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/ゲーム終了！ あなた:1位/)).toBeInTheDocument());
  });

  it('shows rank badge for finished players', async () => {
    const stateWithFinished: SevensResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        {
          id: 1,
          isHuman: false,
          isFinished: true,
          rank: 1,
          cardCount: 0,
          passesUsed: 0,
          maxPasses: 5,
          lastPlayedJoker: false,
          cards: [],
        },
        { ...humanTurnState.players[2] },
        { ...humanTurnState.players[3] },
      ],
    };
    mockExec.mockResolvedValue(stateWithFinished);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('1位')).toBeInTheDocument());
  });

  it('shows thinking indicator on current CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('考え中...')).toBeInTheDocument());
  });

  it('shows pass counts for CPU players', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      // CPU 2 has passesUsed=1
      expect(screen.getByText(/パス: 1\/5/)).toBeInTheDocument();
    });
  });

  it('shows rule header when config has features enabled', async () => {
    const stateWithConfig: SevensResponse = {
      ...humanTurnState,
      config: {
        ...defaultConfig,
        tunnelEnabled: true,
        jokerCount: 2,
        cpuStrategy: 1,
      },
    };
    mockExec.mockResolvedValue(stateWithConfig);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getAllByText(/\[トンネル\]/).length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText(/\[ジョーカー×2\]/)).toBeInTheDocument();
      expect(screen.getByText(/\[CPU戦略\]/)).toBeInTheDocument();
    });
  });

  it('shows harassment rule tag when cpuStrategy is 2', async () => {
    const stateWithHarassment: SevensResponse = {
      ...humanTurnState,
      config: {
        ...defaultConfig,
        cpuStrategy: 2,
      },
    };
    mockExec.mockResolvedValue(stateWithHarassment);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getByText(/\[嫌がらせ特化\]/)).toBeInTheDocument();
    });
    expect(screen.queryByText(/\[CPU戦略\]/)).not.toBeInTheDocument();
  });

  it('does not show rule header with default config', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    expect(screen.queryByText(/ルール:/)).not.toBeInTheDocument();
  });

  it('shows joker card as playable when board has open positions', async () => {
    const jokerState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1 },
      players: [
        {
          ...humanTurnState.players[0],
          cardCount: 2,
          cards: [
            { design: 'JOKER', value: 0 },
            { design: 'SPADE', value: 6 },
          ],
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(jokerState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      const jokerBtn = screen.getByAltText('ジョーカー').closest('button');
      expect(jokerBtn).not.toBeDisabled();
    });
  });

  it('shows joker target description in action log', async () => {
    const jokerActionState: SevensResponse = {
      ...humanTurnState,
      currentTurn: 1,
      humanAction: {
        playerIdx: 0,
        playedCard: { design: 'JOKER', value: 0 },
        targetSuit: 1,
        targetValue: 6,
        forcedPass: false,
      },
    };
    mockExec.mockResolvedValue(jokerActionState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/JOKER/)).toBeInTheDocument();
      expect(screen.getByText(/SPADE 6/)).toBeInTheDocument();
    });
  });

  it('shows tunnel indicator on board', async () => {
    const tunnelState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelEnabled: true },
    };
    mockExec.mockResolvedValue(tunnelState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('[トンネル]')).toBeInTheDocument());
  });

  it('renders config panel with default values', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());
    expect(screen.getByLabelText('トンネル')).not.toBeChecked();
    expect(screen.getByRole('combobox', { name: 'CPU戦略' })).toHaveValue('0');
    expect(screen.getByRole('combobox', { name: 'ジョーカー' })).toHaveValue('0');
    expect(screen.getByRole('combobox', { name: 'パス回数' })).toHaveValue('5');
  });

  it('sends config to API when reset button is clicked with config toggled', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('トンネル'));
    fireEvent.change(screen.getByRole('combobox', { name: 'ジョーカー' }), { target: { value: '1' } });
    fireEvent.change(screen.getByRole('combobox', { name: 'CPU戦略' }), { target: { value: '1' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: {
        ...defaultConfig,
        tunnelEnabled: true,
        jokerCount: 1,
        cpuStrategy: 1,
      },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: true,
        tunnelSkipWidth: 0,
        jokerCount: 1,
        cpuStrategy: 1,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: false,
        endStop: false,
        jokerConsecutiveBanned: false,
      }),
    );
  });

  it('syncs config state from server response', async () => {
    const configState: SevensResponse = {
      ...humanTurnState,
      config: {
        ...defaultConfig,
        tunnelEnabled: true,
        jokerCount: 2,
        cpuStrategy: 1,
        maxPasses: 3,
      },
    };
    mockExec.mockResolvedValue(configState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByLabelText('トンネル')).toBeChecked();
      expect(screen.getByRole('combobox', { name: 'CPU戦略' })).toHaveValue('1');
      expect(screen.getByRole('combobox', { name: 'ジョーカー' })).toHaveValue('2');
      expect(screen.getByRole('combobox', { name: 'パス回数' })).toHaveValue('3');
    });
  });

  it('disables action buttons while loading', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled());

    let resolve!: (value: SevensResponse) => void;
    const slowPromise = new Promise<SevensResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: /パス/ }));

    expect(screen.getByRole('button', { name: /パス/ })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled());
  });

  it('shows error message when API call fails', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  }, 10000);

  it('clears error message on successful API call after failure', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  }, 10000);

  it('A is playable via tunnel when K is placed', async () => {
    // SPADE: bit 7 (128) + bit 13 (8192) placed; A (bit 1) not placed
    const tunnelAState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelEnabled: true },
      tablePlaced: [0, 128 | 8192, 128, 128, 128],
      players: [
        { ...humanTurnState.players[0], cardCount: 1, cards: [{ design: 'SPADE', value: 1 }] },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(tunnelAState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const btn = screen.getByAltText('♠ A').closest('button');
    expect(btn).not.toBeDisabled();
  });

  it('K is playable via tunnel when A is placed', async () => {
    // SPADE: bit 7 (128) + bit 1 (2) placed; K (bit 13) not placed
    const tunnelKState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelEnabled: true },
      tablePlaced: [0, 128 | 2, 128, 128, 128],
      players: [
        { ...humanTurnState.players[0], cardCount: 1, cards: [{ design: 'SPADE', value: 13 }] },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(tunnelKState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());
    const btn = screen.getByAltText('♠ K').closest('button');
    expect(btn).not.toBeDisabled();
  });

  it('JOKER card is not playable when board has no open positions', async () => {
    // All 13 values placed per suit: bits 1-13 set = 16382
    const allPlaced = 0x3ffe; // bits 1-13 set = 16382
    const fullBoardState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1 },
      tablePlaced: [0, allPlaced, allPlaced, allPlaced, allPlaced],
      players: [
        { ...humanTurnState.players[0], cardCount: 1, cards: [{ design: 'JOKER', value: 0 }] },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(fullBoardState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('ジョーカー')).toBeInTheDocument());
    const jokerBtn = screen.getByAltText('ジョーカー').closest('button');
    expect(jokerBtn).toBeDisabled();
  });

  it('shows cancel button when joker card is selected for placement', async () => {
    const jokerHandState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1 },
      players: [
        {
          ...humanTurnState.players[0],
          cards: [
            { design: 'JOKER', value: 0 },
            { design: 'SPADE', value: 6 },
          ],
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(jokerHandState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('ジョーカー')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('ジョーカー'));

    expect(screen.getByRole('button', { name: 'キャンセル' })).toBeInTheDocument();
  });

  it('cancels joker selection when cancel button is clicked', async () => {
    const jokerHandState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1 },
      players: [
        {
          ...humanTurnState.players[0],
          cards: [
            { design: 'JOKER', value: 0 },
            { design: 'SPADE', value: 6 },
          ],
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(jokerHandState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('ジョーカー')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('ジョーカー'));
    expect(screen.getByRole('button', { name: 'キャンセル' })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('button', { name: 'キャンセル' })).not.toBeInTheDocument();
  });

  it('calls joker exec when clicking a valid board position in joker selection mode', async () => {
    const jokerHandState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1 },
      players: [
        { ...humanTurnState.players[0], cardCount: 1, cards: [{ design: 'JOKER', value: 0 }] },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(jokerHandState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('ジョーカー')).toBeInTheDocument());

    // Click JOKER card image to enter joker placement mode (same pattern as cancel test)
    fireEvent.click(screen.getByAltText('ジョーカー'));

    // After click, board position buttons for value '6' appear synchronously (7 is placed)
    const boardButtons6 = screen.getAllByRole('button', { name: /6 に配置/ });
    expect(boardButtons6.length).toBeGreaterThan(0);

    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(boardButtons6[0]);

    // exec is called as: sevensApi.exec('joker', 0, suit, 6)
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('joker', 0, expect.any(Number), 6));
  }, 10000);

  it('shows rank badge when human player finishes', async () => {
    const humanFinishedState: SevensResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0], isFinished: true, rank: 2, cardCount: 0, cards: [] },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(humanFinishedState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getAllByText('2位').length).toBeGreaterThan(0));
  });

  it('pass count dropdown renders with correct options', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());
    const passSelect = screen.getByRole('combobox', { name: 'パス回数' });
    const options = Array.from(passSelect.querySelectorAll('option'));
    expect(options.map((o) => o.textContent)).toEqual(['3', '5', '10', '無制限']);
    expect(options.map((o) => o.getAttribute('value'))).toEqual(['3', '5', '10', '0']);
  });

  it('pass count included in reset config', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());

    fireEvent.change(screen.getByRole('combobox', { name: 'パス回数' }), { target: { value: '3' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, maxPasses: 3 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        tunnelSkipWidth: 0,
        jokerCount: 0,
        cpuStrategy: 0,
        maxPasses: 3,
        noJokerFinish: false,
        jokerReclaim: false,
        endStop: false,
        jokerConsecutiveBanned: false,
      }),
    );
  });

  it('shows 0/∞ for unlimited passes (maxPasses=0)', async () => {
    const unlimitedState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, maxPasses: 0 },
      players: humanTurnState.players.map((p) => ({ ...p, maxPasses: 0 })),
    };
    mockExec.mockResolvedValue(unlimitedState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getAllByText(/0\/∞/).length).toBeGreaterThan(0));
  });

  it('shows 0/3 for custom passes (maxPasses=3)', async () => {
    const customState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, maxPasses: 3 },
      players: humanTurnState.players.map((p) => ({ ...p, maxPasses: 3 })),
    };
    mockExec.mockResolvedValue(customState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getAllByText(/0\/3/).length).toBeGreaterThan(0));
  });

  it('forced pass action shows warning text', async () => {
    const forcedPassState: SevensResponse = {
      ...humanTurnState,
      currentTurn: 1,
      humanAction: { playerIdx: 0, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: true },
    };
    mockExec.mockResolvedValue(forcedPassState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/⚠ 出せるカードなし!/)).toBeInTheDocument());
  });

  it('non-forced pass does NOT show warning text', async () => {
    const passState: SevensResponse = {
      ...cpuTurnState,
      humanAction: { playerIdx: 0, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: false },
    };
    mockExec.mockResolvedValue(passState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/あなたがパスしました/)).toBeInTheDocument());
    expect(screen.queryByText(/⚠ 出せるカードなし!/)).not.toBeInTheDocument();
  });

  it('rules banner shows [パス無制限] when maxPasses=0', async () => {
    const unlimitedRuleState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, maxPasses: 0 },
      players: humanTurnState.players.map((p) => ({ ...p, maxPasses: 0 })),
    };
    mockExec.mockResolvedValue(unlimitedRuleState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getByText(/\[パス無制限\]/)).toBeInTheDocument();
    });
  });

  it('rules banner shows [パス3回] when maxPasses=3', async () => {
    const customRuleState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, maxPasses: 3 },
      players: humanTurnState.players.map((p) => ({ ...p, maxPasses: 3 })),
    };
    mockExec.mockResolvedValue(customRuleState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getByText(/\[パス3回\]/)).toBeInTheDocument();
    });
  });

  it('canPass works with unlimited passes (maxPasses=0, passesUsed=5)', async () => {
    const unlimitedPassState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, maxPasses: 0 },
      players: [
        { ...humanTurnState.players[0], passesUsed: 5, maxPasses: 0 },
        ...humanTurnState.players.slice(1).map((p) => ({ ...p, maxPasses: 0 })),
      ],
    };
    mockExec.mockResolvedValue(unlimitedPassState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled());
  });

  it('forced pass human action has forced-pass testid', async () => {
    const forcedPassHumanState: SevensResponse = {
      ...humanTurnState,
      currentTurn: 1,
      humanAction: { playerIdx: 0, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: true },
    };
    mockExec.mockResolvedValue(forcedPassHumanState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('human-action-forced-pass')).toBeInTheDocument());
  });

  it('non-forced pass human action has normal testid', async () => {
    const normalPassHumanState: SevensResponse = {
      ...humanTurnState,
      currentTurn: 1,
      humanAction: { playerIdx: 0, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: false },
    };
    mockExec.mockResolvedValue(normalPassHumanState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('human-action')).toBeInTheDocument());
  });

  it('forced pass CPU action has forced-pass testid', async () => {
    const forcedPassCpuState: SevensResponse = {
      ...humanTurnState,
      cpuActions: [{ playerIdx: 1, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: true }],
    };
    mockExec.mockResolvedValue(forcedPassCpuState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('cpu-action-forced-pass-0')).toBeInTheDocument());
  });

  it('non-forced pass CPU action has normal testid', async () => {
    const normalPassCpuState: SevensResponse = {
      ...humanTurnState,
      cpuActions: [{ playerIdx: 1, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: false }],
    };
    mockExec.mockResolvedValue(normalPassCpuState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('cpu-action-0')).toBeInTheDocument());
  });

  it('sets aria-busy while loading', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: /パス/ }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: SevensResponse) => void;
    const slowPromise = new Promise<SevensResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: /パス/ }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(humanTurnState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
    });
  });

  it('renders noJokerFinish checkbox with default unchecked state', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());
    expect(screen.getByLabelText('ジョーカー上がり禁止')).not.toBeChecked();
  });

  it('sends noJokerFinish: true in config when checkbox is checked and reset is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('ジョーカー上がり禁止'));

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, noJokerFinish: true },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        tunnelSkipWidth: 0,
        jokerCount: 0,
        cpuStrategy: 0,
        maxPasses: 5,
        noJokerFinish: true,
        jokerReclaim: false,
        endStop: false,
        jokerConsecutiveBanned: false,
      }),
    );
  });

  it('shows rule badge [ジョーカー上がり禁止] when config has noJokerFinish: true', async () => {
    const noJokerFinishState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, noJokerFinish: true },
    };
    mockExec.mockResolvedValue(noJokerFinishState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getByText(/\[ジョーカー上がり禁止\]/)).toBeInTheDocument();
    });
  });

  it('joker card is not playable when noJokerFinish is true and player has only jokers', async () => {
    const noJokerFinishOnlyJokers: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1, noJokerFinish: true },
      players: [
        {
          ...humanTurnState.players[0],
          cardCount: 1,
          cards: [{ design: 'JOKER', value: 0 }],
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(noJokerFinishOnlyJokers);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByAltText('ジョーカー')).toBeInTheDocument());
    const jokerBtn = screen.getByAltText('ジョーカー').closest('button');
    expect(jokerBtn).toBeDisabled();
  });

  it('shows joker played without target info when targetSuit is 0', async () => {
    const jokerNoTargetState: SevensResponse = {
      ...humanTurnState,
      currentTurn: 1,
      humanAction: {
        playerIdx: 0,
        playedCard: { design: 'JOKER', value: 0 },
        targetSuit: 0,
        targetValue: 0,
        forcedPass: false,
      },
    };
    mockExec.mockResolvedValue(jokerNoTargetState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/あなたが出しました: JOKER/)).toBeInTheDocument());
    expect(screen.queryByText(/→/)).not.toBeInTheDocument();
  });

  it('renders ジョーカー回収 checkbox with default unchecked state', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());
    expect(screen.getByLabelText('ジョーカー回収')).not.toBeChecked();
  });

  it('sends jokerReclaim: true in config when checkbox is checked and reset is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('ジョーカー回収'));

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, jokerReclaimEnabled: true },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        tunnelSkipWidth: 0,
        jokerCount: 0,
        cpuStrategy: 0,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: true,
        endStop: false,
        jokerConsecutiveBanned: false,
      }),
    );
  });

  it('syncs jokerReclaim checkbox state from server response', async () => {
    const configState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerReclaimEnabled: true },
    };
    mockExec.mockResolvedValue(configState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByLabelText('ジョーカー回収')).toBeChecked());
  });

  it('shows rule badge [ジョーカー回収] when config has jokerReclaimEnabled: true', async () => {
    const jokerReclaimState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerReclaimEnabled: true },
    };
    mockExec.mockResolvedValue(jokerReclaimState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getByText(/\[ジョーカー回収\]/)).toBeInTheDocument();
    });
  });

  it('renders ↔ tunnel connector after value 13 in each suit row when tunnelEnabled', async () => {
    const tunnelState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelEnabled: true },
    };
    mockExec.mockResolvedValue(tunnelState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    const connectors = screen.getAllByLabelText('トンネル接続');
    // 4 suit rows, each has one ↔ connector
    expect(connectors).toHaveLength(4);
    for (const connector of connectors) {
      expect(connector.textContent).toBe('↔');
    }
  });

  it('does not render ↔ tunnel connector when tunnelEnabled is false', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    expect(screen.queryByLabelText('トンネル接続')).not.toBeInTheDocument();
  });

  it('shows yellow border on A cell when K is placed (tunnel highlight)', async () => {
    // SPADE: bit 7 (128) + bit 13 (8192) placed; A (bit 1) not placed
    const tunnelHighlightState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelEnabled: true },
      tablePlaced: [0, 128 | 8192, 128, 128, 128],
    };
    mockExec.mockResolvedValue(tunnelHighlightState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    // Find A cells in the SPADE row — the first suit row's value 1 (A)
    // Board renders values 1-13 as spans; A is valueName(1) = 'A'
    // The SPADE A cell should have yellow border because K (13) is placed and tunnel is enabled
    const allACells = screen.getAllByText('A');
    // Find the one in SPADE row (first occurrence) that has tunnel highlight border class
    const spadeACell = allACells.find((el) => el.classList.contains('border-ds-warning'));
    expect(spadeACell).toBeDefined();
    expect(spadeACell).toHaveClass('border-ds-warning');
  });

  it('shows yellow border on K cell when A is placed (tunnel highlight)', async () => {
    // SPADE: bit 7 (128) + bit 1 (2) placed; K (bit 13) not placed
    const tunnelHighlightKState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelEnabled: true },
      tablePlaced: [0, 128 | 2, 128, 128, 128],
    };
    mockExec.mockResolvedValue(tunnelHighlightKState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    const allKCells = screen.getAllByText('K');
    // The SPADE K cell should have yellow border class because A (1) is placed and tunnel is enabled
    const spadeKCell = allKCells.find((el) => el.classList.contains('border-ds-warning'));
    expect(spadeKCell).toBeDefined();
    expect(spadeKCell).toHaveClass('border-ds-warning');
  });

  it('does not show tunnel highlight when tunnelEnabled is false', async () => {
    // SPADE: bit 7 (128) + bit 13 (8192) placed; A not placed, but tunnel disabled
    const noTunnelState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelEnabled: false },
      tablePlaced: [0, 128 | 8192, 128, 128, 128],
    };
    mockExec.mockResolvedValue(noTunnelState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    const allACells = screen.getAllByText('A');
    // No A cell should have tunnel highlight class when tunnel is disabled
    const highlightedA = allACells.find((el) => el.classList.contains('border-ds-warning'));
    expect(highlightedA).toBeUndefined();
  });

  // ── TunnelSkipWidth tests ──────────────────────────────────────

  it('renders tunnelSkipWidth config select with default value 0 (off)', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    const select = screen.getByRole('combobox', { name: /トンネルスキップ幅/ });
    expect(select).toHaveValue('0');
  });

  it('shows rule badge [トンネルスキップ3] when config has tunnelSkipWidth >= 2', async () => {
    const skipState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelSkipWidth: 3 },
    };
    mockExec.mockResolvedValue(skipState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getByText(/\[トンネルスキップ3\]/)).toBeInTheDocument();
    });
  });

  it('syncs tunnelSkipWidth select state from server response', async () => {
    const skipState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, tunnelSkipWidth: 4 },
    };
    mockExec.mockResolvedValue(skipState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      const select = screen.getByRole('combobox', { name: /トンネルスキップ幅/ });
      expect(select).toHaveValue('4');
    });
  });

  it('sends tunnelSkipWidth in config when select is changed and reset is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());

    const select = screen.getByRole('combobox', { name: /トンネルスキップ幅/ });
    fireEvent.change(select, { target: { value: '3' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, tunnelSkipWidth: 3 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        tunnelSkipWidth: 3,
        jokerCount: 0,
        cpuStrategy: 0,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: false,
        endStop: false,
        jokerConsecutiveBanned: false,
      }),
    );
  });

  // ── EndStop tests ──────────────────────────────────────────────

  it('renders EndStop config checkbox unchecked by default', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());
    expect(screen.getByLabelText('片側ストップ')).not.toBeChecked();
  });

  it('sends endStop: true in config when checkbox is checked and reset is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('片側ストップ'));

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: true },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        tunnelSkipWidth: 0,
        jokerCount: 0,
        cpuStrategy: 0,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: false,
        endStop: true,
        jokerConsecutiveBanned: false,
      }),
    );
  });

  it('syncs endStop checkbox state from server response', async () => {
    const configState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: true },
    };
    mockExec.mockResolvedValue(configState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByLabelText('片側ストップ')).toBeChecked());
  });

  it('shows rule badge [片側ストップ] when config has endStopEnabled: true', async () => {
    const endStopState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: true },
    };
    mockExec.mockResolvedValue(endStopState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByText(/ルール:/)).toBeInTheDocument();
      expect(screen.getByText(/\[片側ストップ\]/)).toBeInTheDocument();
    });
  });

  it('blocks high side card when A is placed and EndStop enabled', async () => {
    // SPADE: bit 7 (128) + bit 1 (2) placed (A placed)
    const endStopBlockState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: true },
      tablePlaced: [0, 128 | 64 | 32 | 16 | 8 | 4 | 2, 128, 128, 128], // A through 6 + 7 placed
      players: [
        {
          ...humanTurnState.players[0],
          cards: [
            { design: 'SPADE', value: 8 }, // should be blocked (high side, A placed)
            { design: 'HEART', value: 6 }, // should be playable
          ],
          cardCount: 2,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(endStopBlockState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    const playableCards = screen.queryAllByTestId('playable-card');
    // Only Heart 6 should be playable, not Spade 8
    expect(playableCards).toHaveLength(1);
  });

  it('blocks low side card when K is placed and EndStop enabled', async () => {
    // SPADE: bit 7 (128) + bit 13 (8192) + bit 8..12 placed (K placed)
    const endStopBlockLowState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: true },
      tablePlaced: [0, 128 | 256 | 512 | 1024 | 2048 | 4096 | 8192, 128, 128, 128], // 7 through K placed
      players: [
        {
          ...humanTurnState.players[0],
          cards: [
            { design: 'SPADE', value: 6 }, // should be blocked (low side, K placed)
            { design: 'HEART', value: 8 }, // should be playable
          ],
          cardCount: 2,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(endStopBlockLowState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    const playableCards = screen.queryAllByTestId('playable-card');
    // Only Heart 8 should be playable, not Spade 6
    expect(playableCards).toHaveLength(1);
  });

  it('does not block high side when EndStop enabled but A not placed', async () => {
    // SPADE: only 7 placed (bit 7 = 128), EndStop enabled → Spade 8 should be playable
    const endStopNoBlockState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: true },
      tablePlaced: [0, 128, 128, 128, 128],
      players: [
        {
          ...humanTurnState.players[0],
          cards: [{ design: 'SPADE', value: 8 }],
          cardCount: 1,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(endStopNoBlockState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    expect(screen.queryAllByTestId('playable-card')).toHaveLength(1);
  });

  it('does not block low side when EndStop enabled but K not placed', async () => {
    // SPADE: only 7 placed (bit 7 = 128), EndStop enabled → Spade 6 should be playable
    const endStopNoBlockLowState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: true },
      tablePlaced: [0, 128, 128, 128, 128],
      players: [
        {
          ...humanTurnState.players[0],
          cards: [{ design: 'SPADE', value: 6 }],
          cardCount: 1,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(endStopNoBlockLowState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    expect(screen.queryAllByTestId('playable-card')).toHaveLength(1);
  });

  it('does not block cards when EndStop is disabled even if A is placed', async () => {
    // SPADE: A + 2-6 + 7 placed, EndStop disabled → Spade 8 should be playable
    const noEndStopState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, endStopEnabled: false },
      tablePlaced: [0, 128 | 64 | 32 | 16 | 8 | 4 | 2, 128, 128, 128],
      players: [
        {
          ...humanTurnState.players[0],
          cards: [{ design: 'SPADE', value: 8 }],
          cardCount: 1,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(noEndStopState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    const playableCards = screen.queryAllByTestId('playable-card');
    expect(playableCards).toHaveLength(1);
  });

  // ── JokerConsecutiveBanned ──

  it('renders jokerConsecutiveBanned checkbox with default unchecked state', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    expect(screen.getByLabelText('ジョーカー連続禁止')).not.toBeChecked();
  });

  it('sends jokerConsecutiveBanned: true in config when checkbox is checked and reset is clicked', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('ジョーカー連続禁止'));
    expect(screen.getByLabelText('ジョーカー連続禁止')).toBeChecked();

    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, jokerConsecutiveBanned: true },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        tunnelSkipWidth: 0,
        jokerCount: 0,
        cpuStrategy: 0,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: false,
        endStop: false,
        jokerConsecutiveBanned: true,
      }),
    );
  });

  it('syncs jokerConsecutiveBanned checkbox state from server response', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, jokerConsecutiveBanned: true },
    });
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByLabelText('ジョーカー連続禁止')).toBeChecked());
  });

  it('shows rule badge [ジョーカー連続禁止] when config has jokerConsecutiveBanned: true', async () => {
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: { ...defaultConfig, jokerConsecutiveBanned: true },
    });
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/\[ジョーカー連続禁止\]/)).toBeInTheDocument());
  });

  it('joker card is not playable when jokerConsecutiveBanned is true and lastPlayedJoker is true', async () => {
    const jcbState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1, jokerConsecutiveBanned: true },
      players: [
        {
          ...humanTurnState.players[0],
          cards: [
            { design: 'JOKER', value: 0 },
            { design: 'SPADE', value: 6 },
          ],
          cardCount: 2,
          lastPlayedJoker: true,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(jcbState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    // SPADE 6 is playable (adjacent to 7), but JOKER is blocked
    const playableCards = screen.queryAllByTestId('playable-card');
    expect(playableCards).toHaveLength(1);
  });

  it('joker card is playable when jokerConsecutiveBanned is true but lastPlayedJoker is false', async () => {
    const jcbState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1, jokerConsecutiveBanned: true },
      players: [
        {
          ...humanTurnState.players[0],
          cards: [
            { design: 'JOKER', value: 0 },
            { design: 'SPADE', value: 6 },
          ],
          cardCount: 2,
          lastPlayedJoker: false,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(jcbState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    // Both JOKER and SPADE 6 are playable
    const playableCards = screen.queryAllByTestId('playable-card');
    expect(playableCards).toHaveLength(2);
  });

  it('joker card is playable when jokerConsecutiveBanned is false even if lastPlayedJoker is true', async () => {
    const jcbState: SevensResponse = {
      ...humanTurnState,
      config: { ...defaultConfig, jokerCount: 1, jokerConsecutiveBanned: false },
      players: [
        {
          ...humanTurnState.players[0],
          cards: [
            { design: 'JOKER', value: 0 },
            { design: 'SPADE', value: 6 },
          ],
          cardCount: 2,
          lastPlayedJoker: true,
        },
        ...humanTurnState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(jcbState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
    // Both should be playable when rule is disabled
    const playableCards = screen.queryAllByTestId('playable-card');
    expect(playableCards).toHaveLength(2);
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      gameEndFlag: true,
      currentTurn: 0,
      players: [],
      playerIdx: 0,
      tablePlaced: {},
      config: {},
    } as unknown as SevensResponse);

    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.sevens).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.sevens).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // --- ConfirmDialog tests ---

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, expect.any(Object)));
  });

  // --- Keyboard navigation tests ---

  describe('keyboard navigation', () => {
    it('pressing number key directly plays the card at that index', async () => {
      renderWithProviders(<SevensPage />);
      await waitFor(() => expect(screen.getByAltText('♠ 6')).toBeInTheDocument());
      mockExec.mockClear();
      mockExec.mockResolvedValue(cpuTurnState);

      // Press '1' to play the first card (SPADE 6, index 0)
      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
    });

    it('pressing "3" plays the third card', async () => {
      renderWithProviders(<SevensPage />);
      await waitFor(() => expect(screen.getByAltText('♣ 8')).toBeInTheDocument());
      mockExec.mockClear();
      mockExec.mockResolvedValue(cpuTurnState);

      // Press '3' to play third card (CLOVER 8, index 2)
      await act(async () => {
        fireEvent.keyDown(document, { key: '3' });
      });
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
    });

    it('keyboard is disabled when not human turn', async () => {
      mockExec.mockResolvedValue(cpuTurnState);
      renderWithProviders(<SevensPage />);
      await waitFor(() => expect(screen.getByText('ボード')).toBeInTheDocument());
      mockExec.mockClear();

      // Press '1' - should not trigger play since it's CPU's turn
      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      expect(mockExec).not.toHaveBeenCalled();
    });

    it('keyboard is disabled when game has ended', async () => {
      mockExec.mockResolvedValue(gameEndState);
      renderWithProviders(<SevensPage />);
      await waitFor(() => expect(screen.getByText(/ゲーム終了/)).toBeInTheDocument());
      mockExec.mockClear();

      await act(async () => {
        fireEvent.keyDown(document, { key: '1' });
      });
      expect(mockExec).not.toHaveBeenCalled();
    });

    it('ignores key press for out-of-range index', async () => {
      renderWithProviders(<SevensPage />);
      await waitFor(() => expect(screen.getByAltText('♠ 6')).toBeInTheDocument());
      mockExec.mockClear();

      // Human has 3 cards, pressing '5' (index 4) should be ignored
      await act(async () => {
        fireEvent.keyDown(document, { key: '5' });
      });
      expect(mockExec).not.toHaveBeenCalled();
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // -- PhaseIndicator coverage --

  it('phase indicator shows your turn when human turn', async () => {
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('phase indicator shows end phase', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('終了'));
  });

  it('shows the remaining pass count on the pass button', async () => {
    // humanTurnState: passesUsed 0 / maxPasses 5 → 5 remaining.
    mockExec.mockResolvedValue(humanTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /残り5回/ })).toBeInTheDocument());
  });

  it('warns when only one pass remains', async () => {
    const oneLeft: SevensResponse = {
      ...humanTurnState,
      players: [{ ...humanTurnState.players[0], passesUsed: 4, maxPasses: 5 }, ...humanTurnState.players.slice(1)],
    };
    mockExec.mockResolvedValue(oneLeft);
    renderWithProviders(<SevensPage />);
    const btn = await screen.findByRole('button', { name: /残り1回/ });
    expect(btn.className).toContain('text-ds-warning');
  });

  it('omits the count on the pass button when passes are unlimited', async () => {
    const unlimited: SevensResponse = {
      ...humanTurnState,
      players: [{ ...humanTurnState.players[0], maxPasses: 0 }, ...humanTurnState.players.slice(1)],
    };
    mockExec.mockResolvedValue(unlimited);
    renderWithProviders(<SevensPage />);
    const btn = await screen.findByRole('button', { name: 'パス' });
    expect(btn.textContent).toBe('パス');
  });
});
