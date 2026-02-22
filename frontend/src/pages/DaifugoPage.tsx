import { useCallback, useEffect, useState } from 'react';
import { daifugoApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import type { Card, DaifugoAction, DaifugoPlayerData, DaifugoResponse } from '../types/card';

const btnPrimary =
  'px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnWarning =
  'px-3 py-1.5 text-sm font-medium text-gray-900 bg-yellow-400 rounded hover:bg-yellow-500 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnSuccess =
  'px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';

const playerAreaBaseClass =
  'bg-black/35 rounded-[10px] p-[10px] border-2 border-transparent flex-[1_1_180px] min-w-[150px]';

function playerName(idx: number): string {
  return idx === 0 ? 'あなた' : `CPU ${idx}`;
}

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
    default:
      return '';
  }
}

function cardLabel(card: Card | null): string {
  if (!card) return '';
  return `${card.design} ${card.value}`;
}

function actionDescription(action: DaifugoAction): string {
  if (!action.playedCards || action.playedCards.length === 0) {
    return `${playerName(action.playerIdx)}がパスしました`;
  }
  const cards = action.playedCards.map(cardLabel).join(', ');
  return `${playerName(action.playerIdx)}が出しました: ${cards}`;
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
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4 }}>
        {playerName(player.id)}
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
      {!player.isFinished && <div style={{ color: '#ccc', fontSize: '0.85em' }}>{player.cardCount}枚</div>}
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
    <div id="player-area-0" className={playerAreaBaseClass} style={conditionalStyle}>
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4 }}>
        {playerName(0)}
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
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
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

export function DaifugoPage() {
  const [state, setState] = useState<DaifugoResponse | null>(null);
  const [selectedIndices, setSelectedIndices] = useState<number[]>([]);

  const exec = useCallback(async (command: 'reset' | 'play', indices?: number[]) => {
    try {
      const res = await daifugoApi.exec(command, indices);
      setState(res);
      setSelectedIndices([]);
    } catch {
      console.error('daifugo request failed');
    }
  }, []);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  if (!state) return null;

  const isHumanTurn = !state.gameEndFlag && state.currentTurn === 0;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  const toggleCardSelection = (idx: number) => {
    setSelectedIndices((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0, background: '#1a5c1a' }}>
      {/* Scrollable: CPU rows + table cards + action logs + result */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '12px 16px 0' }}>
        {/* CPU row */}
        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap', marginBottom: 10 }}>
          {cpuPlayers.map((player) => (
            <CpuPlayerArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        {/* Table cards */}
        <div
          style={{
            background: 'rgba(0,0,0,0.3)',
            borderRadius: 10,
            padding: 10,
            margin: '8px 0',
          }}
        >
          <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 6 }}>場札</div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
            {!state.tableCards || state.tableCards.length === 0 ? (
              <span style={{ color: '#aaa' }}>（なし）</span>
            ) : (
              state.tableCards.map((card) => (
                <CardImage key={`${card.design}-${card.value}`} card={card} style={{ width: 52 }} />
              ))
            )}
          </div>
        </div>

        {/* Human action log */}
        {state.humanAction && (
          <div
            style={{
              background: 'rgba(0,0,0,0.4)',
              borderRadius: 8,
              color: '#cfc',
              padding: '8px 14px',
              margin: '8px 0',
              fontSize: '0.85em',
            }}
          >
            {actionDescription(state.humanAction)}
          </div>
        )}

        {/* CPU action log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div
            style={{
              background: 'rgba(0,0,0,0.4)',
              borderRadius: 8,
              color: '#ccc',
              padding: '8px 14px',
              margin: '8px 0',
              whiteSpace: 'pre-line',
              fontSize: '0.85em',
            }}
          >
            {['[CPUの行動]', ...state.cpuActions.map(actionDescription)].join('\n')}
          </div>
        )}

        {/* Result message */}
        {state.message && (
          <div
            style={{
              background: 'rgba(0,0,0,0.55)',
              borderRadius: 10,
              color: '#fff',
              textAlign: 'center',
              padding: '10px 16px',
              fontSize: '1.2em',
              fontWeight: 'bold',
              margin: '8px 0',
            }}
          >
            {state.message}
          </div>
        )}
      </div>

      {/* Sticky footer: human player hand + buttons */}
      <div
        style={{
          flexShrink: 0,
          background: '#163e16',
          borderTop: '1px solid rgba(255,255,255,0.2)',
          padding: '10px 16px',
          paddingBottom: 'calc(env(safe-area-inset-bottom) + 10px)',
        }}
      >
        {/* Human player */}
        {humanPlayer && (
          <div style={{ marginBottom: 8 }}>
            <HumanPlayerArea
              player={humanPlayer}
              selectedIndices={selectedIndices}
              onToggle={toggleCardSelection}
              isCurrentTurn={isHumanTurn}
            />
          </div>
        )}

        {/* Buttons */}
        <div className="text-center">
          <button type="button" className={`${btnPrimary} min-w-[90px]`} onClick={() => exec('reset')}>
            リセット
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[90px]`}
            disabled={!isHumanTurn || state.gameEndFlag}
            onClick={() => exec('play', [])}
          >
            パス
          </button>
          <button
            type="button"
            className={`${btnSuccess} min-w-[120px]`}
            disabled={!isHumanTurn || state.gameEndFlag || selectedIndices.length === 0}
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
