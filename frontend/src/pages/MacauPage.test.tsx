import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { macauApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { MacauResponse } from '../types/card';
import { MacauPage } from './MacauPage';

vi.mock('../api/gameApi', () => ({
  macauApi: { exec: vi.fn() },
  actionLogApi: { macau: vi.fn() },
}));

const mockExec = vi.mocked(macauApi.exec);

const playPhaseState: MacauResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      hasDeclared: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 3, cumulativeScore: 10, hasDeclared: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], roundScore: 5, cumulativeScore: 20, hasDeclared: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 5, hasDeclared: false },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  chosenSuit: 0,
  penaltyDrawCount: 0,
  playableIndices: [],
  direction: 1,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

const chooseSuitState: MacauResponse = { ...playPhaseState, phase: 1 };
const mustDeclareState: MacauResponse = { ...playPhaseState, phase: 2 };
const roundEndState: MacauResponse = { ...playPhaseState, phase: 3 };
const gameEndState: MacauResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const cpuTurnState: MacauResponse = { ...playPhaseState, currentPlayerIdx: 1 };
const penaltyState: MacauResponse = { ...playPhaseState, penaltyDrawCount: 4 };
const reverseState: MacauResponse = { ...playPhaseState, direction: -1 };

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('MacauPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MacauPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows a role=status must-declare banner on the human MUST_DECLARE turn', async () => {
    mockExec.mockResolvedValue(mustDeclareState); // phase 2, human's turn
    renderWithProviders(<MacauPage />);
    const banner = await screen.findByTestId('macau-must-declare-banner');
    expect(banner).toHaveAttribute('role', 'status');
    expect(banner).toHaveTextContent('マカオ');
  });

  it('does not show the must-declare banner on a CPU turn', async () => {
    mockExec.mockResolvedValue({ ...mustDeclareState, currentPlayerIdx: 1 });
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByTestId('macau-must-declare-banner')).not.toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('renders human cards', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('renders play and draw buttons on human turn', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument();
    });
  });

  it('calls play when play button clicked', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('calls draw when draw button clicked', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('shows penalty banner and take-penalty draw label when penalty active', async () => {
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByText(/ドローペナルティ/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /引き受ける/ })).toBeInTheDocument();
  });

  it('shows a penalty count badge and danger-styled draw button when penalty active', async () => {
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<MacauPage />);
    const badge = await screen.findByTestId('penalty-badge');
    expect(badge).toHaveTextContent('4');
    const drawButton = screen.getByRole('button', { name: /引き受ける/ });
    expect(drawButton.className).toContain('bg-ds-error');
  });

  it('hides the penalty badge when no penalty is active', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    expect(screen.queryByTestId('penalty-badge')).not.toBeInTheDocument();
  });

  it('shows reverse direction indicator', async () => {
    mockExec.mockResolvedValue(reverseState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByText(/←/)).toBeInTheDocument());
  });

  it('renders choose suit phase with 4 suit buttons and calls suit', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♥ ハート' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 3));
  });

  it('renders declare and skip buttons in must-declare phase', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'マカオ！' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '宣言しない' })).toBeInTheDocument();
    });
  });

  it('calls declare when declare button clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'マカオ！' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'マカオ！' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('calls skipdeclare when skip button clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言しない' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言しない' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipdeclare'));
  });

  it('shows next round button and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('does not show play/draw on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows CPU player areas and score table', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*5枚/)).toBeInTheDocument();
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
    });
  });

  it('shows discard top card', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
    });
  });

  it('shows error alert on failed reset', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('round and direction info displayed', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
      expect(screen.getByText('山札: 30枚')).toBeInTheDocument();
      expect(screen.getByText(/→/)).toBeInTheDocument();
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<MacauPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders the magic card reference panel with domain-accurate effects', async () => {
    renderWithProviders(<MacauPage />);
    const panel = await screen.findByTestId('macau-magic-reference');
    expect(panel).toBeInTheDocument();
    // Panel content mirrors the Go domain magic-card effects (Macau.go).
    expect(panel).toHaveTextContent('マジックカード一覧');
    expect(panel).toHaveTextContent('次のプレイヤーが2枚ドロー（2を重ねてペナルティ累積可）');
    expect(panel).toHaveTextContent('次のプレイヤーをスキップ');
    expect(panel).toHaveTextContent('ワイルド — 好きなスートを指定、いつでも出せる');
    expect(panel).toHaveTextContent('プレイ方向を反転（リバース）');
  });

  // **CUI は出せる札を全部並べているのに、Web は都度クリックしてエラーで
  // 確かめるしかなかった (#4805)。**
  it('rings only the cards that can legally be played', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, playableIndices: [1] });
    renderWithProviders(<MacauPage />);

    await waitFor(() => expect(document.querySelectorAll('[data-playable="true"]')).toHaveLength(1));
  });

  // **「引くしかない」局面で全札が光ってはいけない。**手番中の空配列は
  // 「制限なし」ではなく「1 枚も出せない」(レビュー指摘 #5065)。
  it('rings nothing when it is the human turn and no card is legal', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, playableIndices: [] });
    renderWithProviders(<MacauPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(document.querySelectorAll('[data-playable="true"]')).toHaveLength(0);
  });

  it('rings every card while it is not the human turn', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, currentPlayerIdx: 1, playableIndices: [] });
    renderWithProviders(<MacauPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(document.querySelectorAll('[data-playable="true"]').length).toBeGreaterThan(1);
  });
});
