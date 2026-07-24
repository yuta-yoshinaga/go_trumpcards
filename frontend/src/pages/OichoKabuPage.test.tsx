import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { oichokabuApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, OichoKabuResponse } from '../types/card';
import { OichoKabuPhase } from '../types/phases';
import { OichoKabuPage } from './OichoKabuPage';

vi.mock('../api/gameApi', () => ({
  oichokabuApi: { exec: vi.fn() },
  actionLogApi: { oichokabu: vi.fn() },
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

const mockApi = vi.mocked(oichokabuApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const kabu = (value: number): Card => ({
  design: 'SPADE',
  value,
  glyph: String(value),
  label: String(value),
  color: 'black',
  deck: 'kabu',
});

const betState: OichoKabuResponse = {
  playerHand: [],
  bankerHand: [],
  playerRank: 0,
  bankerRank: 0,
  phase: OichoKabuPhase.BET,
  chips: 1000,
  bet: 0,
  result: 0,
  totalPayout: 0,
  message: '',
};

const drawState: OichoKabuResponse = {
  playerHand: [kabu(4), kabu(5)],
  bankerHand: [],
  playerRank: 9,
  bankerRank: 0,
  phase: OichoKabuPhase.DRAW,
  chips: 900,
  bet: 100,
  result: 0,
  totalPayout: 0,
  message: '',
};

const winState: OichoKabuResponse = {
  playerHand: [kabu(4), kabu(5)],
  bankerHand: [kabu(8), kabu(9)],
  playerRank: 9,
  bankerRank: 7,
  phase: OichoKabuPhase.END,
  chips: 1100,
  bet: 100,
  result: 1,
  totalPayout: 200,
  message: 'You win!',
};

const bankerDrewState: OichoKabuResponse = {
  ...winState,
  bankerHand: [kabu(2), kabu(1), kabu(6)],
  bankerRank: 9,
  result: -1,
  totalPayout: 0,
  message: 'You lose',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockApi.mockResolvedValue(betState);
});

describe('OichoKabuPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders bet phase with bet button', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ベット/ })).toBeInTheDocument());
  });

  it('triggers bet action with current amount', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<OichoKabuPage />);
    const betBtn = await screen.findByRole('button', { name: /^ベット$/ });
    mockApi.mockClear();
    fireEvent.click(betBtn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('renders draw phase with draw & stand buttons', async () => {
    mockApi.mockResolvedValue(drawState);
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /引く/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /勝負/ })).toBeInTheDocument();
  });

  it('triggers draw on draw click', async () => {
    mockApi.mockResolvedValue(drawState);
    renderWithProviders(<OichoKabuPage />);
    const btn = await screen.findByRole('button', { name: /引く/ });
    mockApi.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('draw'));
  });

  it('triggers stand on stand click', async () => {
    mockApi.mockResolvedValue(drawState);
    renderWithProviders(<OichoKabuPage />);
    const btn = await screen.findByRole('button', { name: /勝負/ });
    mockApi.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('stand'));
  });

  it('hides the banker hand during the draw phase', async () => {
    mockApi.mockResolvedValue(drawState);
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /引く/ })).toBeInTheDocument());
    // The child's hand and rank show, but the banker hand stays hidden until the end.
    expect(screen.getByText(/子（あなた）/)).toBeInTheDocument();
    expect(screen.queryByText(/親 — カブ/)).not.toBeInTheDocument();
  });

  it('renders end phase with both hands, payout and reset', async () => {
    mockApi.mockResolvedValue(winState);
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.getByText(/親 — カブ/)).toBeInTheDocument();
    expect(screen.getByText(/200/)).toBeInTheDocument();
  });

  it('exposes each hand rank by name and the result as a live region', async () => {
    mockApi.mockResolvedValue(winState); // playerRank 9 (Kabu), bankerRank 7
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    // Player rank 9 is named (Kabu, strongest); a plain rank (7) is not.
    expect(screen.getByRole('img', { name: '子（あなた）の目9（カブ、最強）' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: '親の目7' })).toBeInTheDocument();
    // The payout/result block is announced.
    expect(screen.getByTestId('payout-breakdown')).toHaveAttribute('role', 'status');
  });

  it('discloses the banker stand policy with its revealed rank at result', async () => {
    mockApi.mockResolvedValue(winState); // 2-card banker, rank 7 (> threshold 6) => stood
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    const policy = screen.getByTestId('dealer-policy');
    expect(policy).toHaveTextContent('親のドロー方針');
    expect(policy).toHaveTextContent('親は目7で規定値（6）を超えていたため引かなかった');
  });

  it('discloses the banker draw policy when the banker took a third card', async () => {
    mockApi.mockResolvedValue(bankerDrewState); // 3-card banker => drew
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.getByTestId('dealer-policy')).toHaveTextContent('親は目が規定値（6）以下だったため3枚目を引いた');
  });

  it('hides the banker draw policy before the result phase', async () => {
    mockApi.mockResolvedValue(drawState);
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /引く/ })).toBeInTheDocument());
    expect(screen.queryByTestId('dealer-policy')).not.toBeInTheDocument();
  });

  it('reads from useCliMode', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(mockUseCliMode).toHaveBeenCalledWith('oichokabu'));
  });

  it('renders the CLI terminal when useCliMode reports cliEnabled', async () => {
    mockUseCliMode.mockReturnValueOnce({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<OichoKabuPage />);
    await waitFor(() => expect(screen.queryAllByRole('button', { name: /ベット/ })).toHaveLength(0));
  });

  it('shows a Rebet button at end-phase after a bet has been placed', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<OichoKabuPage />);
    const betBtn = await screen.findByRole('button', { name: /^ベット$/ });
    mockApi.mockClear();
    mockApi.mockResolvedValue(winState);
    fireEvent.click(betBtn);
    await waitFor(() => expect(screen.getByTestId('ok-rebet-button')).toBeInTheDocument());

    mockApi.mockClear();
    mockApi.mockResolvedValueOnce(betState);
    mockApi.mockResolvedValueOnce(winState);
    fireEvent.click(screen.getByTestId('ok-rebet-button'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('does not show the Rebet button when chips are insufficient to replay', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<OichoKabuPage />);
    const betBtn = await screen.findByRole('button', { name: /^ベット$/ });
    mockApi.mockResolvedValue({ ...winState, chips: 50 });
    fireEvent.click(betBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.queryByTestId('ok-rebet-button')).not.toBeInTheDocument();
  });

  it("the 'd' keyboard shortcut draws during the draw phase", async () => {
    mockApi.mockResolvedValue(drawState);
    renderWithProviders(<OichoKabuPage />);
    await screen.findByRole('button', { name: /引く/ });
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('draw'));
  });

  it('suggests the previous bet in the bet phase and refills the input when clicked', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<OichoKabuPage />);
    const input = (await screen.findByLabelText(/賭け金/)) as HTMLInputElement;

    fireEvent.change(input, { target: { value: '300' } });
    fireEvent.click(screen.getByRole('button', { name: /^ベット$/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 300));

    fireEvent.change(input, { target: { value: '150' } });
    const suggest = await screen.findByTestId('ok-previous-bet');
    expect(suggest).toHaveTextContent('300');

    fireEvent.click(suggest);
    await waitFor(() => expect(input.value).toBe('300'));
    await waitFor(() => expect(screen.queryByTestId('ok-previous-bet')).not.toBeInTheDocument());
  });
});
