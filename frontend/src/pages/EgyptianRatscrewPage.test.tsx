import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { egyptianRatscrewApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { EgyptianRatscrewResponse } from '../types/card';
import { EgyptianRatscrewEventKind, EgyptianRatscrewPhase, EgyptianRatscrewSlapReason } from '../types/phases';
import { EgyptianRatscrewPage } from './EgyptianRatscrewPage';

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

const mockUseCliMode = vi.mocked(useCliMode);

vi.mock('../api/gameApi', () => ({
  egyptianRatscrewApi: { exec: vi.fn() },
  actionLogApi: { egyptianratscrew: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = {
  playSound: mockPlaySound,
  muted: false,
  toggleMute: vi.fn(),
  claimExecSound: vi.fn(),
  consumeExecClaim: () => false,
};
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  // AnimatedCard と中央のタップ (useGameApi / GamePageShell) は useOptionalSound を
  // 使う。同じスパイに向けて、音の名前で絞って検証する。
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(egyptianRatscrewApi.exec);

const baseState: EgyptianRatscrewResponse = {
  phase: EgyptianRatscrewPhase.PLAY,
  gameEndFlag: false,
  winnerIdx: -1,
  currentTurnIdx: 0,
  isHumanTurn: true,
  isTopFaceCard: false,
  isSlappable: false,
  centerPileSize: 0,
  topCard: null,
  players: [
    { name: 'You', isHuman: true, stockSize: 26 },
    { name: 'CPU', isHuman: false, stockSize: 26 },
  ],
  cpuDifficulty: 1,
  chanceRemaining: 0,
  faceChances: { jack: 1, queen: 2, king: 3, ace: 4 },
  chanceFromIdx: -1,
  pendingKind: 0,
  pendingDeadlineMs: 0,
  lastEventKind: 0,
  lastEventPlayerIdx: 0,
  lastSlapReason: 0,
  message: '',
};

const slappableState: EgyptianRatscrewResponse = {
  ...baseState,
  isSlappable: true,
  centerPileSize: 2,
  topCard: { design: 'SPADE', value: 7 },
};

const chanceState: EgyptianRatscrewResponse = {
  ...baseState,
  isTopFaceCard: true,
  chanceRemaining: 2,
  chanceFromIdx: 0,
  centerPileSize: 1,
  topCard: { design: 'HEART', value: 12 },
  isHumanTurn: false,
  currentTurnIdx: 1,
};

const gameEndState: EgyptianRatscrewResponse = {
  ...baseState,
  phase: EgyptianRatscrewPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    { name: 'You', isHuman: true, stockSize: 52 },
    { name: 'CPU', isHuman: false, stockSize: 0 },
  ],
};

beforeEach(() => {
  // **音のスパイはテストごとに消す。**残ると、前のテストで鳴った音を
  // 「CPU のスラップで鳴った」と誤検知する。
  mockPlaySound.mockClear();
  mockExec.mockResolvedValue(baseState);
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

describe('EgyptianRatscrewPage', () => {
  it('renders the GameSkeleton while state is null and does not render raw Loading…', () => {
    // Keep exec pending so `state` stays null and the loading guard renders.
    mockExec.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<EgyptianRatscrewPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.queryByText('Loading…')).not.toBeInTheDocument();
    expect(screen.queryByTestId('step-button')).not.toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders stock counts after state loads', async () => {
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.getAllByText(/26/).length).toBeGreaterThan(0);
  });

  it('step button calls exec with step', async () => {
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('slap button calls exec with slap', async () => {
    mockExec.mockResolvedValueOnce(slappableState);
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('slap-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('slap'));
  });

  it('announces a correct sandwich slap to screen readers', async () => {
    mockExec.mockResolvedValue({
      ...slappableState,
      lastEventKind: EgyptianRatscrewEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 0,
      lastSlapReason: EgyptianRatscrewSlapReason.SANDWICH,
    });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => {
      const announce = screen.getByTestId('er-slap-announce');
      expect(announce).toHaveAttribute('aria-live', 'polite');
      expect(announce.textContent).toMatch(/スラップ成功/);
      expect(announce.textContent).toMatch(/サンドイッチ/);
    });
  });

  it('announces a wrong slap to screen readers', async () => {
    mockExec.mockResolvedValue({
      ...slappableState,
      lastEventKind: EgyptianRatscrewEventKind.SLAP_WRONG,
      lastEventPlayerIdx: 1,
    });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('er-slap-announce').textContent).toMatch(/スラップ失敗/));
  });

  it('disables slap button when pile is empty', async () => {
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeDisabled());
  });

  it('disables step and slap buttons on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
    expect(screen.getByTestId('slap-button')).toBeDisabled();
  });

  it('disables step button when it is not the human turn', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, isHumanTurn: false, currentTurnIdx: 1 });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
  });

  it('renders CLI terminal when enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.queryByTestId('step-button')).not.toBeInTheDocument());
  });

  it('renders the slappable callout and a flashing slap button', async () => {
    mockExec.mockResolvedValueOnce(slappableState);
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    const slap = screen.getByTestId('slap-button');
    expect(slap).not.toBeDisabled();
    expect(slap.className).toMatch(/animate-pulse/);
  });

  it('gives the step and slap buttons a 44x44px minimum tap target', async () => {
    mockExec.mockResolvedValueOnce(slappableState);
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    const slap = screen.getByTestId('slap-button');
    expect(slap.className).toContain('min-h-[44px]');
    expect(slap.className).toContain('min-w-[44px]');
    const step = screen.getByTestId('step-button');
    expect(step.className).toContain('min-h-[44px]');
    expect(step.className).toContain('min-w-[44px]');
  });

  it('shows a pair slap-reason badge while slappable', async () => {
    mockExec.mockResolvedValueOnce({ ...slappableState, lastSlapReason: EgyptianRatscrewSlapReason.PAIR });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('er-slap-reason')).toHaveTextContent('ペア'));
  });

  it('labels the slap-reason badge as a sandwich when applicable', async () => {
    mockExec.mockResolvedValueOnce({ ...slappableState, lastSlapReason: EgyptianRatscrewSlapReason.SANDWICH });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('er-slap-reason')).toHaveTextContent('サンドイッチ'));
  });

  it('does not show the slap-reason badge when the pile is not slappable', async () => {
    mockExec.mockResolvedValueOnce(baseState); // isSlappable false, lastSlapReason 0
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    expect(screen.queryByTestId('er-slap-reason')).not.toBeInTheDocument();
  });

  it('renders chance remaining indicator on face-card chance battle', async () => {
    mockExec.mockResolvedValueOnce(chanceState);
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    expect(screen.getAllByText(/2/).length).toBeGreaterThan(0);
  });

  it('renders a dot per remaining chance and names the responding player', async () => {
    // chanceRemaining 2, currentTurnIdx 1 (CPU) on this fixture.
    mockExec.mockResolvedValueOnce(chanceState);
    renderWithProviders(<EgyptianRatscrewPage />);
    const row = await screen.findByTestId('er-chance-row');
    // One decorative pip per remaining flip.
    expect(row.querySelectorAll('span.rounded-full')).toHaveLength(2);
    // Responder label identifies the CPU (currentTurnIdx 1, isHumanTurn false).
    expect(row).toHaveTextContent('応答: CPU 1');
  });

  it('names the human as responder when it is the human turn during a chance', async () => {
    mockExec.mockResolvedValueOnce({ ...chanceState, isHumanTurn: true, currentTurnIdx: 0, chanceRemaining: 1 });
    renderWithProviders(<EgyptianRatscrewPage />);
    const row = await screen.findByTestId('er-chance-row');
    expect(row.querySelectorAll('span.rounded-full')).toHaveLength(1);
    expect(row).toHaveTextContent('応答: あなた');
  });

  it('reset settings select fires reset with cpuDifficulty config', async () => {
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    const select = screen.getByLabelText(/CPU/i) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('renders Enter and Space keyboard badges on step/slap buttons', async () => {
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.getByTestId('step-button').textContent).toContain('Enter');
    expect(screen.getByTestId('slap-button').textContent).toContain('Space');
  });

  it('Space keypress triggers slap during play', async () => {
    mockExec.mockResolvedValueOnce(slappableState);
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: ' ' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('slap'));
  });

  it('Enter keypress triggers step during play', async () => {
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(window, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });
});

// **姉妹ゲームの Slapjack には正誤の音があるのに、ERS は完全に無音だった
// (#4749)。**視覚 (SlapBurst) と読み上げ (aria-live) はほぼ同じ実装なのに、
// 音のフィードバックだけ欠けていた。
describe('EgyptianRatscrewPage slap sounds', () => {
  it('plays a fanfare on the human’s correct slap', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      lastEventKind: EgyptianRatscrewEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 0,
      lastSlapReason: EgyptianRatscrewSlapReason.PAIR,
    });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('winFanfare'));
  });

  it('plays an error buzz on the human’s false slap', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      lastEventKind: EgyptianRatscrewEventKind.SLAP_WRONG,
      lastEventPlayerIdx: 0,
    });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz'));
  });

  // **CPU のスラップでは鳴らさない。**CPU の成功でファンファーレが鳴ったり、
  // CPU のミスでブザーが鳴って人間が責められたように感じたりしないため。
  it('stays silent for a CPU slap', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      lastEventKind: EgyptianRatscrewEventKind.SLAP_CORRECT,
      lastEventPlayerIdx: 1,
      lastSlapReason: EgyptianRatscrewSlapReason.PAIR,
    });
    renderWithProviders(<EgyptianRatscrewPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(mockPlaySound).not.toHaveBeenCalledWith('winFanfare');
    expect(mockPlaySound).not.toHaveBeenCalledWith('errorBuzz');
  });

  // #5580: チャンスもスラップ条件も、成立してから初めて画面に現れていた。
  // 初見の人は仕組みを推測するしかなかった。
  describe('permanent rule line', () => {
    it('states the chances and both slap conditions from the start', async () => {
      renderWithProviders(<EgyptianRatscrewPage />);
      const rule = await screen.findByTestId('er-rule');
      expect(rule).toHaveTextContent('ペア');
      expect(rule).toHaveTextContent('サンドイッチ');
      // 回数はサーバから来た値。
      expect(rule).toHaveTextContent('J=1');
      expect(rule).toHaveTextContent('A=4');
    });

    // **回数を訳文に焼き込んでいない証拠。**別の回数を返せばそのまま出る。
    it('renders whatever chance counts the server sends', async () => {
      mockExec.mockResolvedValue({ ...baseState, faceChances: { jack: 5, queen: 6, king: 7, ace: 8 } });
      renderWithProviders(<EgyptianRatscrewPage />);
      const rule = await screen.findByTestId('er-rule');
      expect(rule).toHaveTextContent('J=5');
      expect(rule).toHaveTextContent('A=8');
      expect(rule).not.toHaveTextContent('J=1');
    });

    // 動的な表示と重複しないこと (受け入れ条件3)。規則の行はチャンス中も
    // 変わらず、残り回数は既存のドットが受け持つ。
    it('stays put during a chance rather than restating the count left', async () => {
      mockExec.mockResolvedValue(chanceState);
      renderWithProviders(<EgyptianRatscrewPage />);
      const rule = await screen.findByTestId('er-rule');
      expect(rule).toHaveTextContent('J=1');
    });
  });
});
