import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sevensApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevensResponse } from '../types/card';
import { SevensPage } from './SevensPage';

vi.mock('../api/gameApi', () => ({
  sevensApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(sevensApi.exec);

// tableMinVals/tableMaxVals: index 0 unused; 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND
// tablePlaced: bitmask per suit; bit i = value i placed. 7 placed = 1<<7 = 128
// With all 7s placed: value 6 or 8 of any suit is playable
const defaultConfig = {
  tunnelEnabled: false,
  jokerCount: 0,
  cpuStrategy: false,
  maxPasses: 5,
  noJokerFinish: false,
  jokerReclaimEnabled: false,
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
      cards: [
        { design: 'SPADE', value: 6 }, // playable: 6 == 7-1
        { design: 'HEART', value: 5 }, // not playable
        { design: 'CLOVER', value: 8 }, // playable: 8 == 7+1
      ],
    },
    { id: 1, isHuman: false, isFinished: false, rank: -1, cardCount: 4, passesUsed: 0, maxPasses: 5, cards: [] },
    { id: 2, isHuman: false, isFinished: false, rank: -1, cardCount: 3, passesUsed: 1, maxPasses: 5, cards: [] },
    { id: 3, isHuman: false, isFinished: false, rank: -1, cardCount: 5, passesUsed: 0, maxPasses: 5, cards: [] },
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
  message: '1位: あなた',
};

const passesExhaustedState: SevensResponse = {
  ...humanTurnState,
  players: [{ ...humanTurnState.players[0], passesUsed: 5, maxPasses: 5 }, ...humanTurnState.players.slice(1)],
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('SevensPage', () => {
  it('renders nothing before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    const { container } = renderWithProviders(<SevensPage />);
    expect(container.firstChild).toBeNull();
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

  it('pass button is enabled on human turn with passes remaining', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
  });

  it('pass button is disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('pass button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('pass button is disabled when passes are exhausted', async () => {
    mockExec.mockResolvedValue(passesExhaustedState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('calls play with -1 when pass button is clicked', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
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
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        jokerCount: 0,
        cpuStrategy: false,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: false,
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
        { playerIdx: 1, playedCard: { design: 'SPADE', value: 8 }, targetSuit: 0, targetValue: 0, forcedPass: false },
        { playerIdx: 2, playedCard: null, targetSuit: 0, targetValue: 0, forcedPass: false },
      ],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2がパスしました/)).toBeInTheDocument();
  });

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('1位: あなた')).toBeInTheDocument());
  });

  it('shows rank badge for finished players', async () => {
    const stateWithFinished: SevensResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, rank: 1, cardCount: 0, passesUsed: 0, maxPasses: 5, cards: [] },
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
        tunnelEnabled: true,
        jokerCount: 2,
        cpuStrategy: true,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaimEnabled: false,
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
    expect(screen.getByLabelText('CPU戦略')).not.toBeChecked();
    expect(screen.getByRole('combobox', { name: 'ジョーカー' })).toHaveValue('0');
    expect(screen.getByRole('combobox', { name: 'パス回数' })).toHaveValue('5');
  });

  it('sends config to API when reset button is clicked with config toggled', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByText('ルール設定 (リセット時に適用)')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('トンネル'));
    fireEvent.change(screen.getByRole('combobox', { name: 'ジョーカー' }), { target: { value: '1' } });
    fireEvent.click(screen.getByLabelText('CPU戦略'));

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...humanTurnState,
      config: {
        tunnelEnabled: true,
        jokerCount: 1,
        cpuStrategy: true,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaimEnabled: false,
      },
    });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: true,
        jokerCount: 1,
        cpuStrategy: true,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: false,
      }),
    );
  });

  it('syncs config state from server response', async () => {
    const configState: SevensResponse = {
      ...humanTurnState,
      config: {
        tunnelEnabled: true,
        jokerCount: 2,
        cpuStrategy: true,
        maxPasses: 3,
        noJokerFinish: false,
        jokerReclaimEnabled: false,
      },
    };
    mockExec.mockResolvedValue(configState);
    renderWithProviders(<SevensPage />);
    await waitFor(() => {
      expect(screen.getByLabelText('トンネル')).toBeChecked();
      expect(screen.getByLabelText('CPU戦略')).toBeChecked();
      expect(screen.getByRole('combobox', { name: 'ジョーカー' })).toHaveValue('2');
      expect(screen.getByRole('combobox', { name: 'パス回数' })).toHaveValue('3');
    });
  });

  it('disables action buttons while loading', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    let resolve!: (value: SevensResponse) => void;
    const slowPromise = new Promise<SevensResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));

    expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
  });

  it('shows error message when API call fails', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE)).toBeInTheDocument());
  }, 10000);

  it('clears error message on successful API call after failure', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE)).toBeInTheDocument());

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE)).not.toBeInTheDocument());
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
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        jokerCount: 0,
        cpuStrategy: false,
        maxPasses: 3,
        noJokerFinish: false,
        jokerReclaim: false,
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
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
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

  it('sets aria-busy and sr-only loading text while loading', async () => {
    renderWithProviders(<SevensPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'パス' }).closest('[aria-live]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
    expect(screen.queryByText('処理中...')).not.toBeInTheDocument();

    let resolve!: (value: SevensResponse) => void;
    const slowPromise = new Promise<SevensResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));

    expect(container).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByText('処理中...')).toBeInTheDocument();

    resolve(humanTurnState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
      expect(screen.queryByText('処理中...')).not.toBeInTheDocument();
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
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        jokerCount: 0,
        cpuStrategy: false,
        maxPasses: 5,
        noJokerFinish: true,
        jokerReclaim: false,
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
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', -1, 0, 0, {
        tunnelEnabled: false,
        jokerCount: 0,
        cpuStrategy: false,
        maxPasses: 5,
        noJokerFinish: false,
        jokerReclaim: true,
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
    // Find the one in SPADE row (first occurrence) that has tunnel highlight border
    const spadeACell = allACells.find((el) => el.style.borderColor !== '');
    expect(spadeACell).toBeDefined();
    expect(spadeACell).toHaveStyle({ borderColor: '#f59e0b' });
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
    // The SPADE K cell should have yellow border because A (1) is placed and tunnel is enabled
    const spadeKCell = allKCells.find((el) => el.style.borderColor !== '');
    expect(spadeKCell).toBeDefined();
    expect(spadeKCell).toHaveStyle({ borderColor: '#f59e0b' });
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
    // No A cell should have yellow border when tunnel is disabled
    const highlightedA = allACells.find((el) => el.style.borderColor !== '');
    expect(highlightedA).toBeUndefined();
  });
});
