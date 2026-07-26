import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spideretteApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SpideretteResponse, SpideretteTableauCard } from '../types/card';
import { SpiderettePage } from './SpiderettePage';

vi.mock('../api/gameApi', () => ({
  spideretteApi: { exec: vi.fn() },
  actionLogApi: { spiderette: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const playSoundMock = vi.fn();
vi.mock('../providers/SoundProvider', async () => {
  const actual = await vi.importActual<typeof import('../providers/SoundProvider')>('../providers/SoundProvider');
  return {
    ...actual,
    useSound: () => ({ playSound: playSoundMock, muted: false, toggleMute: vi.fn() }),
  };
});

const mockSend = vi.mocked(spideretteApi.exec);

function makeTableau(cols: SpideretteTableauCard[][]): SpideretteTableauCard[][] {
  const result: SpideretteTableauCard[][] = [];
  for (let i = 0; i < 7; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: SpideretteResponse = {
  tableau: makeTableau([
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 5), faceUp: true },
    ],
    [],
    [],
    [],
    [],
    [],
  ]),
  stockCount: 24,
  completedSuits: 0,
  score: 500,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: SpideretteResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'spiderette.gameClear',
  messageParams: { moveCount: '42', score: '500' },
};

beforeEach(() => {
  mockSend.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  playSoundMock.mockClear();
});

describe('SpiderettePage', () => {
  it('renders skeleton when no state', () => {
    mockSend.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpiderettePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('highlights the hint source card and target column after requesting a hint', async () => {
    renderWithProviders(<SpiderettePage />);
    await screen.findByTestId('spdt-card-1-1');
    // The Hint button fetches a hint: move HEART 5 (col 1, idx 1) onto column 0.
    mockSend.mockResolvedValue({ ...playingState, hint: { fromCol: 1, cardIndex: 1, toCol: 0 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Hint-suggested source uses ring-ds-info (distinct from user-selected ring-ds-warning).
    await waitFor(() => expect(screen.getByTestId('spdt-card-1-1').className).toContain('ring-ds-info'));
    expect(screen.getByTestId('spdt-card-1-1').className).not.toContain('ring-ds-warning');
    expect(screen.getByTestId('spdt-col-0').className).toContain('ring-ds-success');
    // A non-target column is not highlighted.
    expect(screen.getByTestId('spdt-col-2').className).not.toContain('ring-ds-success');
    // The hint text is exposed to screen readers via an aria-live status region.
    const status = screen.getByRole('status');
    expect(status).toHaveAttribute('aria-live', 'polite');
    expect(status.textContent).toContain('場札');
  });

  it('hides the frontend hint tooltip when hints are disabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.buildSameSuit', confidence: 'strong' },
      hintEnabled: false,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SpiderettePage />);
    await screen.findByTestId('spdt-card-1-1');
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('shows the frontend hint tooltip when hints are enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.buildSameSuit', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SpiderettePage />);
    const tooltip = await screen.findByTestId('hint-tooltip');
    expect(tooltip).toHaveTextContent('同スートで積み重ねられるカードがあります');
  });

  it('announces the empty-column deal guard to screen readers', async () => {
    renderWithProviders(<SpiderettePage />);
    // playingState has empty columns and stock remaining, so a deal is guarded.
    const dealBtn = (await screen.findAllByRole('button', { name: '配る' }))[0];
    fireEvent.click(dealBtn);
    await waitFor(() => {
      const warn = screen.getByText('空の列をすべて埋めないと配れません');
      expect(warn).toHaveAttribute('role', 'status');
      expect(warn).toHaveAttribute('aria-live', 'assertive');
    });
  });

  it('renders stock count', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(24\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders completed suits 0/4', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/完成: 0\/4/));
  });

  it('shows game clear phase label', async () => {
    mockSend.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockSend.mockClear();
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockSend).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockSend).toHaveBeenCalledWith('giveup'));
  });
});
