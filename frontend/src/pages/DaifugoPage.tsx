import { useCallback, useEffect, useState } from 'react';
import { daifugoApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import type { Card, DaifugoAction, DaifugoExchangeAction, DaifugoPlayerData, DaifugoResponse } from '../types/card';
import { findPlayerName, playerName } from '../utils/playerUtils';

const playerAreaBaseClass =
  'bg-black/35 rounded-[10px] p-[10px] border-2 border-transparent flex-[1_1_180px] min-w-[150px]';

function rankName(rank: number): string {
  switch (rank) {
    case 1:
      return '大富豪';
    case 2:
      return '富豪';
    case 3:
      return '平民';
    case 4:
      return '大貧民';
    /* v8 ignore next 2 */
    default:
      return '';
  }
}

function cardLabel(card: Card): string {
  return `${card.design} ${card.value}`;
}

function actionDescription(players: { id: number; isHuman: boolean }[], action: DaifugoAction): string {
  if (!action.playedCards || action.playedCards.length === 0) {
    return `${findPlayerName(players, action.playerIdx)}がパスしました`;
  }
  const cards = action.playedCards.map(cardLabel).join(', ');
  return `${findPlayerName(players, action.playerIdx)}が出しました: ${cards}`;
}

interface CpuPlayerAreaProps {
  player: DaifugoPlayerData;
  isCurrentTurn: boolean;
}

function CpuPlayerArea({ player, isCurrentTurn }: CpuPlayerAreaProps) {
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
      : {};
  return (
    <div id={`player-area-${player.id}`} className={playerAreaBaseClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && (
          <span
            style={{
              background: '#5cb85c',
              color: '#fff',
              borderRadius: 6,
              padding: '1px 8px',
              marginLeft: 6,
              fontSize: '0.8em',
            }}
          >
            上がり ({rankName(player.rank)})
          </span>
        )}
        {isCurrentTurn && !player.isFinished && (
          <span
            style={{
              background: '#f0ad4e',
              color: '#222',
              borderRadius: 6,
              padding: '1px 8px',
              marginLeft: 6,
              fontSize: '0.8em',
              fontWeight: 'bold',
            }}
          >
            考え中...
          </span>
        )}
      </div>
      {!player.isFinished && <div className="text-[#ccc] text-[0.85em]">{player.cardCount}枚</div>}
    </div>
  );
}

interface HumanPlayerAreaProps {
  player: DaifugoPlayerData;
  selectedIndices: number[];
  onToggle: (idx: number) => void;
  isCurrentTurn: boolean;
}

function HumanPlayerArea({ player, selectedIndices, onToggle, isCurrentTurn }: HumanPlayerAreaProps) {
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {};
  return (
    <div id={`player-area-${player.id}`} className={playerAreaBaseClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && (
          <span
            style={{
              background: '#5cb85c',
              color: '#fff',
              borderRadius: 6,
              padding: '1px 8px',
              marginLeft: 6,
              fontSize: '0.8em',
            }}
          >
            上がり ({rankName(player.rank)})
          </span>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {player.cardCount}枚
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>カードをクリックして選択</span>}
        </div>
      )}
      <div className="flex flex-wrap gap-1">
        {player.cards?.map((card, i) => (
          <button
            key={`${card.design}-${card.value}`}
            type="button"
            disabled={!isCurrentTurn}
            onClick={() => onToggle(i)}
            style={{
              background: 'none',
              padding: 0,
              cursor: isCurrentTurn ? 'pointer' : 'default',
              borderRadius: 8,
              border: selectedIndices.includes(i) ? '3px solid #f0ad4e' : '3px solid transparent',
              boxSizing: 'border-box',
            }}
          >
            <CardImage card={card} width={52} />
          </button>
        ))}
      </div>
    </div>
  );
}

const badgeStyle: React.CSSProperties = {
  display: 'inline-block',
  borderRadius: 6,
  padding: '2px 10px',
  marginRight: 6,
  marginBottom: 4,
  fontSize: '0.8em',
  fontWeight: 'bold',
};

function RulesBadges({ state }: { state: DaifugoResponse }) {
  const badges: { label: string; bg: string; color: string }[] = [];
  if (state.revolutionActive) {
    badges.push({ label: '革命中', bg: '#d9534f', color: '#fff' });
  }
  if (state.elevenBackActive) {
    badges.push({ label: '11バック', bg: '#f0ad4e', color: '#222' });
  }
  if (state.suitLocked) {
    badges.push({ label: `スート縛り: ${state.lockedSuit}`, bg: '#5bc0de', color: '#222' });
  }
  if (state.tableIsSequence) {
    badges.push({ label: '階段', bg: '#9b59b6', color: '#fff' });
  }
  if (badges.length === 0) return null;
  return (
    <div className="my-1 px-1">
      {badges.map((b) => (
        <span key={b.label} style={{ ...badgeStyle, background: b.bg, color: b.color }}>
          {b.label}
        </span>
      ))}
    </div>
  );
}

function exchangeDescription(players: { id: number; isHuman: boolean }[], action: DaifugoExchangeAction): string {
  const from = findPlayerName(players, action.fromPlayerIdx);
  const to = findPlayerName(players, action.toPlayerIdx);
  const cards = action.cards.map(cardLabel).join(', ');
  return `${from} → ${to}: ${cards}`;
}

function ExchangeLog({
  players,
  actions,
}: {
  players: { id: number; isHuman: boolean }[];
  actions: DaifugoExchangeAction[];
}) {
  return (
    <div className="bg-black/40 rounded-lg text-[#ffd] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
      {['[カード交換]', ...actions.map((a) => exchangeDescription(players, a))].join('\n')}
    </div>
  );
}

export function DaifugoPage() {
  const [state, setState] = useState<DaifugoResponse | null>(null);
  const [selectedIndices, setSelectedIndices] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const exec = useCallback(async (command: 'reset' | 'play', indices?: number[]) => {
    setLoading(true);
    try {
      setError(null);
      const res = await daifugoApi.exec(command, indices);
      setState(res);
      setSelectedIndices([]);
    } catch {
      setError('通信エラーが発生しました。もう一度お試しください。');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  if (!state) return null;

  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  const toggleCardSelection = (idx: number) => {
    setSelectedIndices((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a5c1a]">
      {/* Scrollable: CPU rows + table cards + action logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* CPU row */}
        <div className="flex gap-2.5 flex-wrap mb-2.5">
          {cpuPlayers.map((player) => (
            <CpuPlayerArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        {/* Table cards */}
        <div className="bg-black/30 rounded-[10px] p-2.5 my-2">
          <div className="text-white font-bold mb-1.5">場札</div>
          <div className="flex flex-wrap gap-1">
            {!state.tableCards || state.tableCards.length === 0 ? (
              <span style={{ color: '#aaa' }}>（なし）</span>
            ) : (
              state.tableCards.map((card) => <CardImage key={`${card.design}-${card.value}`} card={card} width={52} />)
            )}
          </div>
        </div>

        {/* Local rules status badges */}
        <RulesBadges state={state} />

        {/* Card exchange actions */}
        {state.exchangeActions && state.exchangeActions.length > 0 && (
          <ExchangeLog players={state.players} actions={state.exchangeActions} />
        )}

        {/* Human action log */}
        {state.humanAction && (
          <div className="bg-black/40 rounded-lg text-[#cfc] py-2 px-3.5 my-2 text-[0.85em]">
            {actionDescription(state.players, state.humanAction)}
          </div>
        )}

        {/* CPU action log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-[#ccc] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
            {['[CPUの行動]', ...state.cpuActions.map((a) => actionDescription(state.players, a))].join('\n')}
          </div>
        )}

        {/* Result message */}
        {state.message && (
          <div className="bg-black/55 rounded-[10px] text-white text-center py-2.5 px-4 text-[1.2em] font-bold my-2">
            {state.message}
          </div>
        )}
      </div>

      {/* Sticky footer: human player hand + buttons */}
      <div
        className="shrink-0 bg-[#163e16] border-t border-white/20 px-4 py-2.5"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 10px)' }}
      >
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
            <HumanPlayerArea
              player={humanPlayer}
              selectedIndices={selectedIndices}
              onToggle={toggleCardSelection}
              isCurrentTurn={isHumanTurn}
            />
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => exec('reset')}
          >
            リセット
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[90px]`}
            disabled={loading || !isHumanTurn || state.gameEndFlag}
            onClick={() => exec('play', [])}
          >
            パス
          </button>
          <button
            type="button"
            className={`${btnSuccess} min-w-[120px]`}
            disabled={loading || !isHumanTurn || state.gameEndFlag || selectedIndices.length === 0}
            onClick={() =>
              exec(
                'play',
                [...selectedIndices].sort((a, b) => a - b),
              )
            }
          >
            選択して出す
          </button>
        </div>
      </div>
    </div>
  );
}
