import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, twoTenJackApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTwoTenJackState } from '../test/stateFactories';
import { TwoTenJackPage } from './TwoTenJackPage';

vi.mock('../api/gameApi', () => ({
  twoTenJackApi: { exec: vi.fn() },
  actionLogApi: { twotenjack: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockExec = vi.mocked(twoTenJackApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const playPhaseState = makeTwoTenJackState();

const declarePhaseState = makeTwoTenJackState({
  phase: 0,
  declarerIdx: 0,
  trumpSuit: -1,
});

const declarePhaseCpuState = makeTwoTenJackState({
  phase: 0,
  declarerIdx: 1,
  trumpSuit: -1,
});

const trickEndState = makeTwoTenJackState({
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
    { playerIdx: 2, card: { design: 'CLOVER', value: 7 } },
    { playerIdx: 3, card: { design: 'SPADE', value: 9 } },
  ],
});

const roundEndState = makeTwoTenJackState({ phase: 3 });

const gameEndState = makeTwoTenJackState({
  phase: 4,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
});

const cpuTurnState = makeTwoTenJackState({ currentPlayerIdx: 1 });

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

afterEach(() => {
  localStorage.clear();
});

describe('TwoTenJackPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TwoTenJackPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 50,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('gives the score table an sr-only caption describing its purpose', async () => {
    renderWithProviders(<TwoTenJackPage />);
    const captions = await screen.findAllByText('チーム別の得点集計');
    expect(captions.length).toBeGreaterThan(0);
    expect(captions[0].tagName).toBe('CAPTION');
  });

  it('shows four suit buttons during human declare phase', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30b9\u30da\u30fc\u30c9' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30af\u30e9\u30d6' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30cf\u30fc\u30c8' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30c0\u30a4\u30e4' })).toBeInTheDocument();
    });
  });

  it('shows a trump-selection prompt beside the suit buttons when the human declares', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByTestId('tt-declare-prompt')).toBeInTheDocument());
  });

  it('does not show the trump-selection prompt when a CPU declares', async () => {
    mockExec.mockResolvedValue(declarePhaseCpuState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.queryByTestId('skeleton')).not.toBeInTheDocument());
    expect(screen.queryByTestId('tt-declare-prompt')).not.toBeInTheDocument();
  });

  it('shows a CPU-declaring indicator while a CPU declares trump', async () => {
    mockExec.mockResolvedValue(declarePhaseCpuState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByTestId('tt-cpu-declaring')).toBeInTheDocument());
    expect(screen.getByTestId('tt-cpu-declaring')).toHaveTextContent('CPUが切り札を宣言中');
  });

  it('does not show the CPU-declaring indicator when the human declares', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByTestId('tt-declare-prompt')).toBeInTheDocument());
    expect(screen.queryByTestId('tt-cpu-declaring')).not.toBeInTheDocument();
  });

  it('dispatches declare command when a suit button is clicked', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30cf\u30fc\u30c8' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30cf\u30fc\u30c8' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 3));
  });

  it('does not show suit buttons during CPU declare turn', async () => {
    mockExec.mockResolvedValue(declarePhaseCpuState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '\u30cf\u30fc\u30c8' })).not.toBeInTheDocument();
  });

  it('play button disabled when no card selected', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u51fa\u3059' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '\u51fa\u3059' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0));
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );
  });

  it('plays the win celebration when the human team wins', async () => {
    mockExec.mockResolvedValue(gameEndState); // winnerTeam: 0 = human team
    renderWithProviders(<TwoTenJackPage />);
    expect(await screen.findByTestId('win-celebration')).toBeInTheDocument();
  });

  it('does not celebrate when the CPU team wins', async () => {
    mockExec.mockResolvedValue({ ...gameEndState, winnerTeam: 1 });
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByText('Game end!')).toBeInTheDocument());
    // Outwait the celebration's 400ms delay so a wrongly-fired overlay would be visible.
    await act(() => new Promise((resolve) => setTimeout(resolve, 600)));
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('reset confirm dispatches reset with current config', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 50,
      }),
    );
  });

  it('shows error alert on failed reset', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('ensures actionLogApi.twotenjack is registered', () => {
    expect(actionLogApi.twotenjack).toBeDefined();
  });

  it('renders CLI terminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('changing cpuDifficulty updates config passed to reset', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 2,
        pointLimit: 50,
      }),
    );
  });

  it('changing pointLimit updates config passed to reset', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '100' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
      }),
    );
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<TwoTenJackPage />);
    // wait for state to load (human cards appear in play phase)
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled and state has backend hint', async () => {
    localStorage.setItem('hint_enabled_twotenjack', 'true');
    // override state to include backend hint so getTwoTenJackHint returns non-null
    const hintState = makeTwoTenJackState({ hint: { reason: 'lead', cardIndex: 0 } });
    mockExec.mockResolvedValue(hintState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('fetches and displays a server play recommendation when the hint button is clicked', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const hintState = makeTwoTenJackState({ hint: { reason: 'lead', cardIndex: 0 } });
    mockExec.mockResolvedValue(hintState);
    fireEvent.click(screen.getByTestId('tt-hint-button'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    await waitFor(() => {
      const el = screen.getByTestId('tt-hint');
      expect(el).toHaveTextContent('推奨プレイ');
      expect(el).toHaveTextContent('[0]');
    });
  });

  it('fetches and displays a server trump recommendation during the declare phase', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByTestId('tt-declare-prompt')).toBeInTheDocument());

    const hintState = makeTwoTenJackState({
      phase: 0,
      declarerIdx: 0,
      trumpSuit: -1,
      hint: { reason: 'strategic_trump', trumpSuit: 1 },
    });
    mockExec.mockResolvedValue(hintState);
    fireEvent.click(screen.getByTestId('tt-hint-button'));

    await waitFor(() => {
      const el = screen.getByTestId('tt-hint');
      expect(el).toHaveTextContent('推奨トランプ');
      expect(el).toHaveTextContent('♠');
    });
  });

  it('shows the hint error in the ErrorAlert when the hint request fails', async () => {
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    mockExec.mockRejectedValueOnce(new Error('network'));
    fireEvent.click(screen.getByTestId('tt-hint-button'));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<TwoTenJackPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
      const allDetails = container.querySelectorAll('details');
      const cpuDetails = Array.from(allDetails).find((d) =>
        d.querySelector('summary')?.textContent?.includes('CPU対戦相手'),
      );
      expect(cpuDetails).toBeInTheDocument();
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders score table as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<TwoTenJackPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="tt-score-table"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('スコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // **同じページ内で非対称だった (#4741)。**プレイフェーズはヒント札を
  // 光らせているのに、宣言フェーズはテキストで「♠」と言うだけで、4つの
  // 宣言ボタンのどれが推奨か視覚的に分からなかった。
  it('highlights the suggested trump button during the declare phase', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByLabelText('スペード')).toBeInTheDocument());

    // ヒントを要求すると、推奨スートのボタンだけが強調される。
    mockExec.mockResolvedValue({ ...declarePhaseState, hint: { reason: 'declare', trumpSuit: 3 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(document.querySelectorAll('[data-hint-suggested="true"]')).toHaveLength(1));
    expect(screen.getByLabelText('ハート')).toHaveAttribute('data-hint-suggested', 'true');
  });

  // 逆側。ヒントを要求していなければ、どのボタンも強調しない。
  it('highlights no trump button before a hint is requested', async () => {
    mockExec.mockResolvedValue(declarePhaseState);
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(screen.getByLabelText('スペード')).toBeInTheDocument());

    expect(document.querySelectorAll('[data-hint-suggested="true"]')).toHaveLength(0);
  });
});

// #5527: 読み上げの「+N」に獲得点札の生の合計を渡していた。契約を落とした
// ラウンドでは実際の加点が 0 でも点札は取っているので、読み上げた数字と
// 実際に累計へ足された数字が食い違う。
describe('TwoTenJackPage round announcement', () => {
  // ヒント用の常設ライブ領域が入って role=status が2つになった (#6663)。
  // 読みたいのはメッセージ箱の方なので、ヒント領域を除いて拾う。
  const announce = () =>
    screen
      .getAllByRole('status')
      .filter((el) => el.getAttribute('data-testid') !== 'twotenjack-hint-live')
      .map((el) => el.textContent ?? '')
      .join(' ');

  const stateWith = (fields: { captured: number[]; round: number[]; cumulative: number[] }) =>
    makeTwoTenJackState({
      phase: 3,
      players: makeTwoTenJackState().players.map((p, i) => ({
        ...p,
        capturedPoints: fields.captured[i],
        roundScore: fields.round[i],
        cumulativeScore: fields.cumulative[i],
      })),
    });

  it('announces the score that was actually added, not the points captured', async () => {
    // 宣言チーム(0)が 20 点しか取れず契約を落としたラウンド。
    // 加点はチーム1に 6 点、チーム0 は 0 点。
    mockExec.mockResolvedValue(
      stateWith({ captured: [12, 30, 8, 20], round: [0, 6, 0, 6], cumulative: [10, 16, 10, 16] }),
    );
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(announce()).toContain('ラウンド終了'));

    // チーム0 は点札を 20 枚分取っているが加点は 0。
    expect(announce()).toContain('+0');
    // チーム1 は 2 人分の roundScore の合計 = 累計の増分。
    expect(announce()).toContain('+12');
    // 生の獲得点札 (20 / 50) は読み上げに出てこない。
    expect(announce()).not.toContain('+20');
    expect(announce()).not.toContain('+50');
  });

  // **席が欠けたレスポンスでも落ちない。**合計は `?? 0` で埋めるので、
  // その分岐も一度は通しておく (codecov は ?. と ?? を別の枝に数える)。
  it('treats a missing seat as zero rather than crashing', async () => {
    const short = makeTwoTenJackState({ phase: 3 });
    mockExec.mockResolvedValue({ ...short, players: short.players.slice(0, 2) });
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(announce()).toContain('ラウンド終了'));
    expect(announce()).toContain('+0');
  });

  // 席が1つも無いレスポンス: 4本の `?.` がすべて短絡する側を通す。
  it('renders with no players at all', async () => {
    const short = makeTwoTenJackState({ phase: 3 });
    mockExec.mockResolvedValue({ ...short, players: [] });
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(announce()).toContain('ラウンド終了'));
    expect(announce()).toContain('+0');
  });

  it('still shows the captured points in the score table', async () => {
    mockExec.mockResolvedValue(
      stateWith({ captured: [12, 30, 8, 20], round: [0, 6, 0, 6], cumulative: [10, 16, 10, 16] }),
    );
    renderWithProviders(<TwoTenJackPage />);
    await waitFor(() => expect(announce()).toContain('ラウンド終了'));
    // 点札の合計 (12+8=20 / 30+20=50) は表に残す。
    expect(screen.getAllByText('20').length).toBeGreaterThan(0);
    expect(screen.getAllByText('50').length).toBeGreaterThan(0);
  });
});
