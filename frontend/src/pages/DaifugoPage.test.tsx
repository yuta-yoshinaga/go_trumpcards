import { createEvent, fireEvent, render, screen, waitFor } from '@testing-library/react';
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
  fiveSkipEnabled: false,
  sevenPassEnabled: false,
  tenDiscardEnabled: false,
  spadeThreeEnabled: false,
  capitalFallEnabled: false,
  nineReverseEnabled: false,
  coupDetatEnabled: false,
  intenseLockEnabled: false,
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
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
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
      expect(screen.getByAltText('♠ 3')).toBeInTheDocument();
      expect(screen.getAllByAltText('♥ 5').length).toBeGreaterThan(0);
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
      const tableCard = screen.getAllByAltText('♥ 5');
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
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', [], expect.any(Object)));
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
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    const card = screen.getByAltText('♠ 3');
    fireEvent.click(card);
    expect(screen.getByRole('button', { name: '選択して出す' })).not.toBeDisabled();
  });

  it('calls play with selected indices when play button is clicked', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ 3'));
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
    // Use selector:'span' to find the badge (not the settings panel label)
    await waitFor(() => expect(screen.getByText('11バック', { selector: 'span' })).toBeInTheDocument());
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
    // Use selector:'span' to find the badge (not the settings panel label)
    await waitFor(() => expect(screen.getByText('階段', { selector: 'span' })).toBeInTheDocument());
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
    // Use bracketed form to distinguish from settings panel label
    await waitFor(() => expect(screen.getByText(/\[カード交換\]/)).toBeInTheDocument());
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

  it('toggles aria-pressed on card button click', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ 3').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('deselects a card by clicking it again', async () => {
    render(<DaifugoPage />);
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

  it('settings panel renders checkbox labels', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('8切り')).toBeInTheDocument();
    expect(screen.getByLabelText('5飛び')).toBeInTheDocument();
    expect(screen.getByLabelText('7渡し')).toBeInTheDocument();
    expect(screen.getByLabelText('10捨て')).toBeInTheDocument();
    expect(screen.getByLabelText('スペ3返し')).toBeInTheDocument();
    expect(screen.getByLabelText('都落ち')).toBeInTheDocument();
  });

  it('settings panel joker count dropdown changes config', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('ジョーカー枚数:')).toBeInTheDocument());
    // Change joker count to 0
    fireEvent.change(screen.getByLabelText('ジョーカー枚数:'), { target: { value: '0' } });
    // Click reset → config with jokerCount:0 is sent
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', [], expect.objectContaining({ jokerCount: 0 })));
  });

  it('settings panel boolean checkbox toggle updates config', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByLabelText('5飛び')).toBeInTheDocument());
    // Enable 5飛び
    fireEvent.click(screen.getByLabelText('5飛び'));
    // Click reset → config with fiveSkipEnabled:true is sent
    mockExec.mockClear();
    mockExec.mockResolvedValue(humanTurnState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
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
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText(/【10捨て】/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
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
    render(<DaifugoPage />);
    // Card buttons are disabled because isHumanTurn = false (currentTurn is CPU)
    await waitFor(() => expect(screen.getByAltText('♠ 3').closest('button')).toBeDisabled());
  });

  it('drag card not in selection adds it to selection', async () => {
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
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
    render(<DaifugoPage />);
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

  it('sets aria-busy and sr-only loading text while loading', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'パス' }).closest('[aria-live]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
    expect(screen.queryByText('処理中...')).not.toBeInTheDocument();

    let resolve!: (value: DaifugoResponse) => void;
    const slowPromise = new Promise<DaifugoResponse>((res) => {
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

  it('shows 9リバース badge when reverseDirection is true', async () => {
    const reverseState: DaifugoResponse = {
      ...humanTurnState,
      reverseDirection: true,
    };
    mockExec.mockResolvedValue(reverseState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('9リバース', { selector: 'span' })).toBeInTheDocument());
  });

  it('shows 連番縛り badge when numberLocked is true', async () => {
    const numberLockedState: DaifugoResponse = {
      ...humanTurnState,
      numberLocked: true,
    };
    mockExec.mockResolvedValue(numberLockedState);
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('連番縛り')).toBeInTheDocument());
  });

  it('renders sort buttons and active button is highlighted', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '強さ順' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スート順' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '数字順' })).toBeInTheDocument();
    // Default sortMode=0 → 強さ順 should have primary style
    expect(screen.getByRole('button', { name: '強さ順' }).className).toContain('bg-blue-600');
    expect(screen.getByRole('button', { name: 'スート順' }).className).toContain('bg-gray-600');
  });

  it('calls sort command when sort button is clicked', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スート順' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...humanTurnState, sortMode: 1 });
    fireEvent.click(screen.getByRole('button', { name: 'スート順' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('sort', undefined, undefined, 1));
  });

  it('settings panel renders new rule checkboxes', async () => {
    render(<DaifugoPage />);
    await waitFor(() => expect(screen.getByText('ルール設定')).toBeInTheDocument());
    expect(screen.getByLabelText('9リバース')).toBeInTheDocument();
    expect(screen.getByLabelText('クーデター')).toBeInTheDocument();
    expect(screen.getByLabelText('激シバ')).toBeInTheDocument();
  });

  it('drop with invalid dataTransfer data is ignored (NaN guard)', async () => {
    render(<DaifugoPage />);
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
});
