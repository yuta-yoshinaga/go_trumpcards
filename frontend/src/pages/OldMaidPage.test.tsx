import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { oldmaidApi } from '../api/gameApi';
import type { OldMaidResponse } from '../types/card';
import { OldMaidPage } from './OldMaidPage';

vi.mock('../api/gameApi', () => ({
  oldmaidApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(oldmaidApi.exec);

const humanTurnState: OldMaidResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 2 },
        { design: 'JOKER', value: 0 },
      ],
    },
    { id: 1, isHuman: false, isFinished: false, cardCount: 4, cards: [] },
    { id: 2, isHuman: false, isFinished: false, cardCount: 2, cards: [] },
  ],
  currentTurn: 0,
  nextDrawTargetIdx: 1,
  gameEndFlag: false,
  hasDrawn: false,
  lastDrawPlayerIdx: 0,
  lastDrawFromIdx: 0,
  lastDrawCard: null,
  lastDiscardedPairs: 0,
  cpuActions: [],
  humanAction: null,
  message: '',
};

const cpuTurnState: OldMaidResponse = {
  ...humanTurnState,
  currentTurn: 1,
  nextDrawTargetIdx: 0,
  hasDrawn: true,
  lastDrawPlayerIdx: 1,
  lastDrawFromIdx: 2,
  lastDrawCard: { design: 'HEART', value: 5 },
  lastDiscardedPairs: 1,
};

const gameEndState: OldMaidResponse = {
  ...humanTurnState,
  gameEndFlag: true,
  message: 'あなたが負けました',
};

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState);
});

describe('OldMaidPage', () => {
  it('renders nothing before first API response', () => {
    // Simulate pending promise that never resolves during this check
    mockExec.mockReturnValue(new Promise(() => undefined));
    const { container } = render(<OldMaidPage />);
    expect(container.firstChild).toBeNull();
  });

  it('calls reset command on mount', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  it('renders human player area labeled あなた', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('あなた')).toBeInTheDocument());
  });

  it('renders CPU player areas with correct labels', async () => {
    render(<OldMaidPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
    });
  });

  it('shows human player cards', async () => {
    render(<OldMaidPage />);
    await waitFor(() => {
      // Human player cards: SPADE 1, HEART 2, JOKER 0
      expect(screen.getByAltText('SPADE 1')).toBeInTheDocument();
      expect(screen.getByAltText('HEART 2')).toBeInTheDocument();
      expect(screen.getByAltText('JOKER 0')).toBeInTheDocument();
    });
  });

  it('shows card counts for CPU players', async () => {
    render(<OldMaidPage />);
    await waitFor(() => {
      expect(screen.getByText('4枚')).toBeInTheDocument();
      expect(screen.getByText('2枚')).toBeInTheDocument();
    });
  });

  it('shows target player badge (← 引く相手) on human turn', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('← 引く相手')).toBeInTheDocument());
  });

  it('shows instruction to click target player on human turn', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText(/あなたの番！/)).toBeInTheDocument());
  });

  it('random draw button is enabled on human turn', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());
  });

  it('random draw button is disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('ランダムに引く')).toBeDisabled());
  });

  it('random draw button is disabled when game has ended', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('ランダムに引く')).toBeDisabled());
  });

  it('calls reset when reset button is clicked', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('リセット')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  it('calls draw when random draw button is clicked', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', undefined));
  });

  it('shows status message when hasDrawn is true', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1がCPU 2から1枚引きました/)).toBeInTheDocument());
  });

  it('shows discarded pairs in status message', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText(/1組捨てました/)).toBeInTheDocument());
  });

  it('shows finished badge for completed players', async () => {
    const stateWithFinished: OldMaidResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0] },
        { id: 1, isHuman: false, isFinished: true, cardCount: 0, cards: [] },
        { ...humanTurnState.players[2] },
      ],
    };
    mockExec.mockResolvedValue(stateWithFinished);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('上がり')).toBeInTheDocument());
  });

  it('shows game result message when game ends', async () => {
    mockExec.mockResolvedValue(gameEndState);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('あなたが負けました')).toBeInTheDocument());
  });

  it('shows CPU actions log when cpuActions is non-empty', async () => {
    const stateWithCpuActions: OldMaidResponse = {
      ...humanTurnState,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0 }],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1がCPU 2から1枚引きました/)).toBeInTheDocument();
  });

  it('does not show drawn card in CPU actions log', async () => {
    const stateWithCpuActions: OldMaidResponse = {
      ...humanTurnState,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0 }],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument());
    expect(screen.queryByText(/SPADE 3/)).not.toBeInTheDocument();
  });

  it('calls draw with drawIdx when a target player card back is clicked', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('← 引く相手')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(cpuTurnState);
    // CPU 1 is target player: click its first card back
    const cardBacks = screen.getAllByAltText('card back');
    fireEvent.click(cardBacks[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', 0));
  });

  it('shows stacked discarded cards when lastDiscardedCards has a pair', async () => {
    const stateWithDiscarded: OldMaidResponse = {
      ...humanTurnState,
      hasDrawn: true,
      lastDrawPlayerIdx: 0,
      lastDrawFromIdx: 1,
      lastDrawCard: { design: 'SPADE', value: 5 },
      lastDiscardedPairs: 1,
      lastDiscardedCards: [
        { design: 'SPADE', value: 5 },
        { design: 'CLOVER', value: 5 },
      ],
    };
    mockExec.mockResolvedValue(stateWithDiscarded);
    render(<OldMaidPage />);
    await waitFor(() => {
      expect(screen.getByAltText('SPADE 5')).toBeInTheDocument();
      expect(screen.getByAltText('CLOVER 5')).toBeInTheDocument();
    });
  });

  it('shows human draw status before CPU replay and eventually shows CPU log', async () => {
    const stateWithCpuActions: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
            { design: 'JOKER', value: 0 },
            { design: 'DIAMOND', value: 7 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 1, cards: [] },
      ],
      currentTurn: 0,
      nextDrawTargetIdx: 1,
      hasDrawn: true,
      lastDrawPlayerIdx: 1,
      lastDrawFromIdx: 2,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0, discardedCards: [] }],
      humanAction: {
        drawPlayerIdx: 0,
        drawFromIdx: 1,
        drawnCard: { design: 'HEART', value: 3 },
        discardedPairs: 0,
        discardedCards: [],
      },
    };

    // reset returns humanTurnState, draw returns stateWithCpuActions
    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(stateWithCpuActions);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    // Trigger draw
    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', undefined));

    // After all replay delays, final state CPU log should be visible
    // Each delay is 800ms; humanAction + 1 cpuAction = 2 delays = 1600ms max
    await waitFor(() => expect(screen.getByText(/\[CPUの行動\]/)).toBeInTheDocument(), { timeout: 4000 });
  }, 10000);

  it('disables buttons while loading', async () => {
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ランダムに引く' })).not.toBeDisabled());

    let resolve!: (value: OldMaidResponse) => void;
    const slowPromise = new Promise<OldMaidResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ランダムに引く' }));

    expect(screen.getByRole('button', { name: 'ランダムに引く' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(humanTurnState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ランダムに引く' })).not.toBeDisabled());
  });

  it('enables random draw button after CPU replay animation completes (Bug 1 regression)', async () => {
    // Final server state has currentTurn=0 (human's turn)
    const finalState: OldMaidResponse = {
      ...humanTurnState,
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 2 },
            { design: 'JOKER', value: 0 },
            { design: 'DIAMOND', value: 7 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
        { id: 2, isHuman: false, isFinished: false, cardCount: 1, cards: [] },
      ],
      currentTurn: 0,
      nextDrawTargetIdx: 1,
      hasDrawn: true,
      lastDrawPlayerIdx: 1,
      lastDrawFromIdx: 2,
      lastDrawCard: null,
      lastDiscardedPairs: 0,
      cpuActions: [{ drawPlayerIdx: 1, drawFromIdx: 2, drawnCard: null, discardedPairs: 0, discardedCards: [] }],
      humanAction: {
        drawPlayerIdx: 0,
        drawFromIdx: 1,
        drawnCard: { design: 'HEART', value: 3 },
        discardedPairs: 0,
        discardedCards: [],
      },
    };

    mockExec.mockResolvedValueOnce(humanTurnState).mockResolvedValueOnce(finalState);
    render(<OldMaidPage />);
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled());

    fireEvent.click(screen.getByText('ランダムに引く'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw', undefined));

    // After all replay delays, the final state (currentTurn=0) must be applied
    // so the button is re-enabled for the human's next turn
    await waitFor(() => expect(screen.getByText('ランダムに引く')).not.toBeDisabled(), { timeout: 4000 });
  }, 10000);
});
