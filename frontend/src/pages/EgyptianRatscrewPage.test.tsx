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
