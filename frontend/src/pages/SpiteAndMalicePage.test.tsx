import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spiteAndMaliceApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SpiteAndMaliceResponse } from '../types/card';
import { SpiteAndMalicePage } from './SpiteAndMalicePage';

vi.mock('../api/gameApi', () => ({
  spiteAndMaliceApi: { exec: vi.fn() },
  actionLogApi: { spiteandmalice: vi.fn() },
}));

const mockExec = vi.mocked(spiteAndMaliceApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseState: SpiteAndMaliceResponse = {
  phase: 0,
  current: 0,
  players: [
    {
      hand: [card('SPADE', 5), card('HEART', 8), card('DIAMOND', 11), card('CLOVER', 13), card('SPADE', 2)],
      goalTop: card('HEART', 9),
      goalSize: 20,
      sides: [[], [], [], []],
      isCpu: false,
    },
    {
      hand: [],
      goalTop: card('CLOVER', 7),
      goalSize: 20,
      sides: [[], [], [], []],
      isCpu: true,
    },
  ],
  foundations: [[], [], [], []],
  foundationTops: [0, 0, 0, 0],
  stockSize: 60,
  completedSize: 0,
  moveCount: 0,
  winner: -1,
  goalSize: 20,
  cpuDifficulty: 1,
  canAutoComplete: false,
  message: '',
  messageCode: 'spiteandmalice.playing',
};

const cpuTurnState: SpiteAndMaliceResponse = { ...baseState, current: 1 };

const winState: SpiteAndMaliceResponse = {
  ...baseState,
  phase: 1,
  winner: 0,
  moveCount: 42,
  messageCode: 'spiteandmalice.win',
  messageParams: { moveCount: '42' },
};

const loseState: SpiteAndMaliceResponse = {
  ...baseState,
  phase: 1,
  winner: 1,
  moveCount: 42,
  messageCode: 'spiteandmalice.lose',
  messageParams: { moveCount: '42' },
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(baseState);
});

describe('SpiteAndMalicePage', () => {
  it('renders heading', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count label', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getByText(/手数|Moves/)).toBeInTheDocument());
  });

  it('explains why discard is disabled until a hand card is selected', async () => {
    mockExec.mockResolvedValue(baseState);
    renderWithProviders(<SpiteAndMalicePage />);
    const discardBtns = await screen.findAllByRole('button', { name: 'ディスカード' });
    expect(discardBtns[0]).toHaveAttribute('aria-describedby', 'sam-discard-hint');
    // Nothing selected → the shared hint states the requirement and discard is disabled.
    expect(discardBtns[0]).toBeDisabled();
    expect(screen.getByTestId('sam-discard-hint')).toHaveTextContent('先に手札');
    // Selecting a hand card (♠5) flips the hint to "ready" and enables discard.
    fireEvent.click(screen.getByRole('button', { name: /♠ 5/ }));
    await waitFor(() => expect(screen.getByTestId('sam-discard-hint')).toHaveTextContent('ディスカードできます'));
    expect(screen.getAllByRole('button', { name: 'ディスカード' })[0]).not.toBeDisabled();
  });

  it('localizes the hand heading, card aria-labels, and CPU goal (no raw English)', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Hand card aria-labels are localized ("手札 N: ..."), not "Hand N: ...".
    expect(screen.getByLabelText(/^手札 1:/)).toBeInTheDocument();
    expect(screen.queryByLabelText(/^Hand 1:/)).not.toBeInTheDocument();
    // CPU goal count is localized, not the hardcoded "CPU goal x20".
    expect(screen.getByText(/CPUゴール: 20/)).toBeInTheDocument();
    expect(screen.queryByText(/CPU goal/)).not.toBeInTheDocument();
  });

  it('localizes a hidden (face-down) hand card and a zero CPU goal', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [
        { ...baseState.players[0], hand: [card('SPADE', 5), null] },
        { ...baseState.players[1], goalTop: undefined, goalSize: 0 },
      ],
    });
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Null hand card uses the localized hidden label.
    expect(screen.getByLabelText(/手札 2: 非公開/)).toBeInTheDocument();
    // Empty goal pile renders the localized zero-count label.
    expect(screen.getByText(/CPUゴール: 0/)).toBeInTheDocument();
  });

  it('renders the CPU side piles with a count badge on non-empty piles', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      players: [
        baseState.players[0],
        { ...baseState.players[1], sides: [[card('SPADE', 4), card('HEART', 3)], [], [], []] },
      ],
    });
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getByTestId('sam-cpu-sides')).toBeInTheDocument());
    // Pile 0 has 2 cards → top card image + count badge "2".
    expect(within(screen.getByTestId('sam-cpu-side-0')).getByText('2')).toBeInTheDocument();
    // The other three piles render an empty dashed placeholder (no count badge).
    expect(within(screen.getByTestId('sam-cpu-side-1')).queryByText(/\d/)).not.toBeInTheDocument();
  });

  it('drives CPU turn automatically', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('cpu'), { timeout: 2000 });
  });

  it('shows win phase label', async () => {
    mockExec.mockResolvedValue(winState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー|Game Over/).length).toBeGreaterThan(0));
  });

  it('shows lose phase label', async () => {
    mockExec.mockResolvedValue(loseState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー|Game Over/).length).toBeGreaterThan(0));
  });

  it('renders the autocomplete button on human turn and disables it when canAutoComplete=false', async () => {
    mockExec.mockResolvedValue(baseState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getByTestId('sam-autocomplete-btn')).toBeInTheDocument());
    expect(screen.getByTestId('sam-autocomplete-btn')).toBeDisabled();
  });

  it('enables the autocomplete button when canAutoComplete=true and dispatches the command on click', async () => {
    const playableState: SpiteAndMaliceResponse = { ...baseState, canAutoComplete: true };
    mockExec.mockResolvedValue(playableState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getByTestId('sam-autocomplete-btn')).toBeEnabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playableState);
    fireEvent.click(screen.getByTestId('sam-autocomplete-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('marks the human goal pile as playable when goalTop fits a foundation', async () => {
    const playableGoalState: SpiteAndMaliceResponse = {
      ...baseState,
      players: [{ ...baseState.players[0], goalTop: card('SPADE', 1) }, baseState.players[1]],
    };
    mockExec.mockResolvedValue(playableGoalState);
    renderWithProviders(<SpiteAndMalicePage />);
    const goalBtn = await screen.findByRole('button', { name: /ゴール|Goal/ });
    expect(goalBtn.dataset.goalPlayable).toBe('true');
    expect(goalBtn.className).toContain('ring-ds-warning');
    expect(goalBtn.className).toContain('motion-safe:animate-pulse');
  });

  it('does not pulse the goal pile when goalTop cannot be played', async () => {
    const idleGoalState: SpiteAndMaliceResponse = {
      ...baseState,
      players: [{ ...baseState.players[0], goalTop: card('SPADE', 5) }, baseState.players[1]],
      foundationTops: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(idleGoalState);
    renderWithProviders(<SpiteAndMalicePage />);
    const goalBtn = await screen.findByRole('button', { name: /ゴール|Goal/ });
    expect(goalBtn.dataset.goalPlayable).toBe('false');
    expect(goalBtn.className).not.toContain('ring-ds-warning');
  });

  it('does not pulse the goal pile on CPU turn even if it would be playable', async () => {
    const playableButCpu: SpiteAndMaliceResponse = {
      ...cpuTurnState,
      players: [{ ...cpuTurnState.players[0], goalTop: card('SPADE', 1) }, cpuTurnState.players[1]],
    };
    mockExec.mockResolvedValue(playableButCpu);
    renderWithProviders(<SpiteAndMalicePage />);
    const goalBtn = await screen.findByRole('button', { name: /ゴール|Goal/ });
    expect(goalBtn.dataset.goalPlayable).toBe('false');
  });

  it('does not pulse the goal pile at game over even if goalTop is playable', async () => {
    // winState has phase=GAME_OVER but current=0 (still the human's turn),
    // so a winning hand with a playable goalTop must not keep pulsing on the
    // game-over screen.
    const gameOverPlayable: SpiteAndMaliceResponse = {
      ...winState,
      players: [{ ...winState.players[0], goalTop: card('SPADE', 1) }, winState.players[1]],
      foundationTops: [0, 0, 0, 0],
    };
    mockExec.mockResolvedValue(gameOverPlayable);
    renderWithProviders(<SpiteAndMalicePage />);
    const goalBtn = await screen.findByRole('button', { name: /ゴール|Goal/ });
    expect(goalBtn.dataset.goalPlayable).toBe('false');
    expect(goalBtn.className).not.toContain('ring-ds-warning');
  });

  it('hides the autocomplete button on CPU turn and at game over', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    const { unmount } = renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('sam-autocomplete-btn')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(winState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー|Game Over/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('sam-autocomplete-btn')).not.toBeInTheDocument();
  });

  it('shows the HintTooltip reason when hints are enabled and a hint is available', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      hint: { source: 'goal', index: 0, foundationIdx: 0, discard: false },
    });
    renderWithProviders(<SpiteAndMalicePage />);
    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    if (!(toggle as HTMLInputElement).checked) fireEvent.click(toggle);
    const tooltip = await screen.findByTestId('hint-tooltip');
    expect(tooltip).toHaveTextContent('ゴールパイルのトップをファウンデーションに出せます');
  });

  it('defaults the CPU speed select to normal when unset', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    const select = (await screen.findByTestId('sam-cpu-speed-select')) as HTMLSelectElement;
    expect(select.value).toBe('normal');
  });

  it('loads the persisted CPU speed from localStorage on mount', async () => {
    localStorage.setItem('spiteandmalice:cpuSpeed', 'fast');
    renderWithProviders(<SpiteAndMalicePage />);
    const select = (await screen.findByTestId('sam-cpu-speed-select')) as HTMLSelectElement;
    expect(select.value).toBe('fast');
  });

  it('persists the chosen CPU speed to localStorage', async () => {
    renderWithProviders(<SpiteAndMalicePage />);
    const select = (await screen.findByTestId('sam-cpu-speed-select')) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: 'slow' } });
    expect(select.value).toBe('slow');
    expect(localStorage.getItem('spiteandmalice:cpuSpeed')).toBe('slow');
  });

  it('uses the selected speed as the CPU turn wait', async () => {
    vi.useFakeTimers();
    try {
      // 'fast' → 200ms delay. Verify the CPU turn fires only after the boundary.
      localStorage.setItem('spiteandmalice:cpuSpeed', 'fast');
      mockExec.mockResolvedValue(cpuTurnState);
      renderWithProviders(<SpiteAndMalicePage />);
      // Flush the mount reset + initial render so the CPU-turn effect subscribes.
      await vi.advanceTimersByTimeAsync(0);
      mockExec.mockClear();
      // Just before the 200ms fast delay no CPU turn has fired yet.
      await vi.advanceTimersByTimeAsync(199);
      expect(mockExec).not.toHaveBeenCalledWith('cpu');
      // Crossing 200ms fires the fast CPU turn.
      await vi.advanceTimersByTimeAsync(1);
      expect(mockExec).toHaveBeenCalledWith('cpu');
    } finally {
      vi.useRealTimers();
    }
  });
});

// #5560: K はどの基礎札にも出せるワイルドだが、その規則が表示にも読み上げにも
// 出ておらず、初見では気付けなかった。
describe('SpiteAndMalicePage wild king', () => {
  it('badges the King in hand and says so in the label', async () => {
    mockExec.mockResolvedValue(baseState); // 手札に ♣K がある
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getAllByTestId('sam-wild-badge').length).toBeGreaterThan(0));
    // 読み上げにも入る。
    expect(screen.getByRole('button', { name: /♣ K.*ワイルド/ })).toBeInTheDocument();
  });

  // **K 以外には付かない。**全部に付いたら区別にならない。
  it('leaves other ranks alone', async () => {
    mockExec.mockResolvedValue(baseState);
    renderWithProviders(<SpiteAndMalicePage />);
    await waitFor(() => expect(screen.getAllByTestId('sam-wild-badge')).toHaveLength(1));
    expect(screen.queryByRole('button', { name: /♠ 5.*ワイルド/ })).not.toBeInTheDocument();
  });
});
