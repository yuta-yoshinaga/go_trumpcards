import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { reddogApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, RedDogResponse } from '../types/card';
import { RedDogPhase } from '../types/phases';
import { RedDogPage } from './RedDogPage';

vi.mock('../api/gameApi', () => ({
  reddogApi: { exec: vi.fn() },
  actionLogApi: { reddog: vi.fn() },
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

const mockApi = vi.mocked(reddogApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betState: RedDogResponse = {
  initialCards: [],
  phase: RedDogPhase.BET,
  chips: 1000,
  ante: 0,
  raise: 0,
  spread: 0,
  result: 0,
  totalPayout: 0,
  message: '',
};

const spreadState: RedDogResponse = {
  initialCards: [card('SPADE', 5), card('HEART', 10)],
  phase: RedDogPhase.SPREAD_DECISION,
  chips: 900,
  ante: 100,
  raise: 0,
  spread: 4,
  result: 0,
  totalPayout: 0,
  message: '',
};

const winState: RedDogResponse = {
  initialCards: [card('SPADE', 5), card('HEART', 10)],
  thirdCard: card('CLOVER', 7),
  phase: RedDogPhase.END,
  chips: 1100,
  ante: 100,
  raise: 0,
  spread: 4,
  result: 1,
  totalPayout: 200,
  message: 'You win!',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockApi.mockResolvedValue(betState);
});

describe('RedDogPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('falls back to the initial-dealt phase label', async () => {
    mockApi.mockResolvedValue({
      ...betState,
      phase: RedDogPhase.INITIAL_DEALT,
      initialCards: [card('SPADE', 5), card('HEART', 10)],
      ante: 100,
    });
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('初手公開'));
  });

  it('renders bet phase with bet button', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ベット/ })).toBeInTheDocument());
  });

  it('triggers bet action with current amount', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<RedDogPage />);
    const betBtn = await screen.findByRole('button', { name: /ベット/ });
    mockApi.mockClear();
    fireEvent.click(betBtn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('renders spread decision with raise & stay buttons', async () => {
    mockApi.mockResolvedValue(spreadState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /レイズ/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /ステイ/ })).toBeInTheDocument();
  });

  it('triggers stay on stay click', async () => {
    mockApi.mockResolvedValue(spreadState);
    renderWithProviders(<RedDogPage />);
    const stayBtn = await screen.findByRole('button', { name: /ステイ/ });
    mockApi.mockClear();
    fireEvent.click(stayBtn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('stay'));
  });

  it('triggers raise on raise click', async () => {
    mockApi.mockResolvedValue(spreadState);
    renderWithProviders(<RedDogPage />);
    const raiseBtn = await screen.findByRole('button', { name: /レイズ/ });
    mockApi.mockClear();
    fireEvent.click(raiseBtn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('raise', 100));
  });

  it('disables the raise button when chips fall below the minimum bet', async () => {
    mockApi.mockResolvedValue({ ...spreadState, chips: 5 });
    renderWithProviders(<RedDogPage />);
    const raiseBtn = await screen.findByRole('button', { name: /レイズ/ });
    expect(raiseBtn).toBeDisabled();
  });

  it('renders end phase with reset and payout', async () => {
    mockApi.mockResolvedValue(winState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.getByText(/200/)).toBeInTheDocument();
  });

  it('does not render an empty settings panel', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ベット/ })).toBeInTheDocument());
    expect(screen.queryByText('設定')).not.toBeInTheDocument();
  });

  it('reads from useCliMode', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(mockUseCliMode).toHaveBeenCalledWith('reddog'));
  });

  it('shows ghost rank chips for winning ranks during spread decision', async () => {
    mockApi.mockResolvedValue(spreadState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByTestId('reddog-ghost-ranks')).toBeInTheDocument());
    for (const r of [6, 7, 8, 9]) {
      expect(screen.getByTestId(`reddog-ghost-${r}`)).toBeInTheDocument();
    }
    // Text summary mirrors the ghost chips for at-a-glance / non-visual reading.
    expect(screen.getByTestId('reddog-winners-text')).toHaveTextContent('6, 7, 8, 9');
  });

  it('marks the hit ghost chip in end phase', async () => {
    mockApi.mockResolvedValue(winState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByTestId('reddog-ghost-7')).toBeInTheDocument());
    const hitChip = screen.getByTestId('reddog-ghost-7');
    expect(hitChip.className).toContain('bg-ds-success');
  });
});

// --- keyboard shortcut execution (#4429) ---
// Red Dog binds only three keys, each gated on a different phase.
const kbdCases: [string, unknown[], RedDogResponse][] = [
  ['b', ['bet', 100], betState],
  ['s', ['stay'], spreadState],
  ['r', ['reset'], winState],
];

describe('RedDogPage keyboard shortcuts', () => {
  it.each(kbdCases)('pressing %s dispatches %j', async (key, expected, state) => {
    mockApi.mockResolvedValue(state);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    mockApi.mockResolvedValue(state);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith(...expected));
  });

  it('ignores a key whose phase gate is closed', async () => {
    // 's' (stay) is only offered once the spread has been revealed.
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<RedDogPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 's' });
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalled();
  });

  it('renders the i18n skeleton instead of a hardcoded Loading label before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<RedDogPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
  });
});
