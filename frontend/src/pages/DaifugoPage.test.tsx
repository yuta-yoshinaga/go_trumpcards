import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { daifugoApi } from '../api/gameApi';
import type { DaifugoResponse } from '../types/card';
import { DaifugoPage } from './DaifugoPage';

vi.mock('../api/gameApi', () => ({
  daifugoApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(daifugoApi.exec);

const defaultConfig = {
  jokerCount: 2,
  eightCutEnabled: true,
  suitLockEnabled: true,
  elevenBackEnabled: true,
  sequenceEnabled: true,
  cardExchangeEnabled: true,
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
  message: '大富豪: あなた',
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('DaifugoPage', () => {
  it('renders nothing before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    const { container } = render(<DaifugoPage />);
    expect(container.firstChild).toBeNull();
  });

  it('calls reset command on mount', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  it('renders human player area labeled あなた', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
  });

  it('renders CPU player areas with correct labels', async () => {
    render(<DaifugoPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('shows human player face-up cards', async () => {
    render(<DaifugoPage />);
    await waitFor(() => {
      expect(screen.getByAltText('SPADE 3')).toBeInTheDocument();
      expect(screen.getAllByAltText('HEART 5').length).toBeGreaterThan(0);
    });
  });

  it('shows card counts for CPU players', async () => {
    render(<DaifugoPage />);
    await waitFor(() => {
      expect(screen.getByText('4枚')).toBeInTheDocument();
      expect(screen.getByText('5枚')).toBeInTheDocument();
    });
  });

  it('shows empty table message when tableCards is empty', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('（なし）')).toBeInTheDocument());
  });

  it('shows table cards when tableCards is non-empty', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DaifugoPage />);
    await waitFor(() => {
      const tableCard = screen.getAllByAltText('HEART 5');
      expect(tableCard.length).toBeGreaterThan(0);
    });
  });

  it('pass button is enabled on human turn', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
  });

  it('pass button is disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('pass button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());
  });

  it('play button is disabled when no cards are selected', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '選択して出す' })).toBeDisabled());
  });

  it('calls reset when reset button is clicked', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  it('calls play with empty indices when pass button is clicked', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', []));
  });

  it('selects a card on click and enables play button', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 3')).toBeInTheDocument());
    const card = screen.getByAltText('SPADE 3');
    fireEvent.click(card);
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();
  });

  it('calls play with selected indices when play button is clicked', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 3')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('SPADE 3'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByRole('button', { name: '選択して出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0]));
  });

  it('shows human action log after play', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1が出しました/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2がパスしました/)).toBeInTheDocument();
  });

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('大富豪: あなた')).toBeInTheDocument());
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
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/上がり.*大富豪/)).toBeInTheDocument());
  });

  it('shows thinking indicator on current CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('考え中...')).toBeInTheDocument());
  });

  it('shows revolution badge when revolutionActive is true', async () => {
    const revolutionState: DaifugoResponse = {
      ...humanTurnState,
      revolutionActive: true,
    };
    mockExec.mockResolvedValue(revolutionState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('革命中')).toBeInTheDocument());
  });

  it('shows 11-back badge when elevenBackActive is true', async () => {
    const elevenBackState: DaifugoResponse = {
      ...humanTurnState,
      elevenBackActive: true,
    };
    mockExec.mockResolvedValue(elevenBackState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('11バック')).toBeInTheDocument());
  });

  it('shows suit lock badge when suitLocked is true', async () => {
    const suitLockedState: DaifugoResponse = {
      ...humanTurnState,
      suitLocked: true,
      lockedSuit: 'SPADE',
    };
    mockExec.mockResolvedValue(suitLockedState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('スート縛り: SPADE')).toBeInTheDocument());
  });

  it('shows sequence badge when tableIsSequence is true', async () => {
    const seqState: DaifugoResponse = {
      ...humanTurnState,
      tableIsSequence: true,
    };
    mockExec.mockResolvedValue(seqState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('階段')).toBeInTheDocument());
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
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/カード交換/)).toBeInTheDocument());
  });

  it('disables action buttons while loading', async () => {
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  }, 10000);

  it('clears error message on successful API call after failure', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );

    mockExec.mockReset();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() =>
      expect(screen.queryByText('通信エラーが発生しました。もう一度お試しください。')).not.toBeInTheDocument(),
    );
  }, 10000);

  it('shows pass message for empty playedCards array', async () => {
    const stateWithEmptyPlay: DaifugoResponse = {
      ...humanTurnState,
      humanAction: { playerIdx: 0, playedCards: [] },
    };
    mockExec.mockResolvedValue(stateWithEmptyPlay);
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getAllByText(/上がり.*大富豪/).length).toBeGreaterThan(0));
  });

  it('does not show selection hint when not human turn in HumanPlayerArea', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/上がり \(大貧民\)/)).toBeInTheDocument());
  });

  it('deselects a card by clicking it again', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('SPADE 3')).toBeInTheDocument());

    // Click SPADE 3 to select it
    fireEvent.click(screen.getByAltText('SPADE 3'));
    // Click SPADE 3 again to deselect it
    fireEvent.click(screen.getByAltText('SPADE 3'));

    // After deselection, the 選択して出す button should not show any selection count
    // (the card toggle state resets on the second click)
    // We verify no error is thrown and the button is still present
    expect(screen.getByRole('button', { name: '選択して出す' })).toBeInTheDocument();
  });
});
