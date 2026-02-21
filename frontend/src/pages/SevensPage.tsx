import { useCallback, useEffect, useState } from 'react';
import { sevensApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import type { Card, CardDesign, SevensAction, SevensPlayerData, SevensResponse } from '../types/card';

// Design → suit index (matches Go backend: 1=SPADE, 2=CLOVER, 3=HEART, 4=DIAMOND)
const designToSuit: Record<CardDesign, number> = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
  JOKER: 0,
};

const SUITS = [
  { idx: 1, name: 'SPADE', label: '♠', color: '#e0e0e0' },
  { idx: 2, name: 'CLOVER', label: '♣', color: '#e0e0e0' },
  { idx: 3, name: 'HEART', label: '♥', color: '#f87171' },
  { idx: 4, name: 'DIAMOND', label: '♦', color: '#f87171' },
];

function valueName(v: number): string {
  if (v === 1) return 'A';
  if (v === 11) return 'J';
  if (v === 12) return 'Q';
  if (v === 13) return 'K';
  return String(v);
}

function isCardPlayable(card: Card, tableMinVals: number[], tableMaxVals: number[]): boolean {
  const suit = designToSuit[card.design];
  if (suit < 1 || suit > 4) return false;
  const v = card.value;
  const leftOk = v === tableMinVals[suit] - 1 && tableMinVals[suit] > 1;
  const rightOk = v === tableMaxVals[suit] + 1 && tableMaxVals[suit] < 13;
  return leftOk || rightOk;
}

function playerName(idx: number): string {
  return idx === 0 ? 'あなた' : `CPU ${idx}`;
}

function actionDesc(action: SevensAction): string {
  if (!action.playedCard) return `${playerName(action.playerIdx)}がパスしました`;
  const c = action.playedCard;
  return `${playerName(action.playerIdx)}が出しました: ${c.design} ${valueName(c.value)}`;
}

// ── styles ──────────────────────────────────────────────────────────────────

const playerAreaBaseClass =
  'bg-black/35 rounded-[10px] p-[10px] border-2 border-transparent flex-[1_1_180px] min-w-[150px]';
const btnPrimary =
  'px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnWarning =
  'px-3 py-1.5 text-sm font-medium text-gray-900 bg-yellow-400 rounded hover:bg-yellow-500 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';

// ── Board component ──────────────────────────────────────────────────────────

interface BoardProps {
  tableMinVals: number[];
  tableMaxVals: number[];
}

function Board({ tableMinVals, tableMaxVals }: BoardProps) {
  return (
    <div
      style={{
        background: 'rgba(0,0,0,0.3)',
        borderRadius: 10,
        padding: '10px 14px',
        margin: '8px 0',
      }}
    >
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 8 }}>ボード</div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
        {SUITS.map(({ idx, name, label, color }) => {
          const min = tableMinVals[idx] ?? 7;
          const max = tableMaxVals[idx] ?? 7;
          return (
            <div
              key={name}
              style={{
                background: 'rgba(255,255,255,0.08)',
                borderRadius: 8,
                padding: '6px 10px',
                display: 'flex',
                alignItems: 'center',
                gap: 8,
              }}
            >
              <span style={{ color, fontWeight: 'bold', fontSize: '1.1em', minWidth: 18 }}>{label}</span>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 3 }}>
                {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => {
                  const placed = v >= min && v <= max;
                  const isCenter = v === 7;
                  return (
                    <span
                      key={v}
                      style={{
                        display: 'inline-block',
                        width: 22,
                        height: 22,
                        lineHeight: '22px',
                        textAlign: 'center',
                        borderRadius: 4,
                        fontSize: '0.7em',
                        fontWeight: isCenter ? 'bold' : 'normal',
                        background: placed ? (isCenter ? '#f0ad4e' : '#5cb85c') : 'rgba(255,255,255,0.1)',
                        color: placed ? '#000' : '#555',
                      }}
                    >
                      {valueName(v)}
                    </span>
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ── CPU player area ──────────────────────────────────────────────────────────

interface CpuAreaProps {
  player: SevensPlayerData;
  isCurrentTurn: boolean;
}

function CpuArea({ player, isCurrentTurn }: CpuAreaProps) {
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
      : {};
  const showCount = Math.min(player.cardCount, 10);
  return (
    <div className={playerAreaBaseClass} style={conditionalStyle}>
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
            {player.rank}位
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
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {player.cardCount}枚　パス: {player.passesUsed}/{player.maxPasses}
        </div>
      )}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
        {!player.isFinished &&
          Array.from({ length: showCount }).map((_, i) => <CardBack key={i} style={{ width: 50 }} />)}
        {player.cardCount > 10 && (
          <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 4 }}>+{player.cardCount - 10}</span>
        )}
      </div>
    </div>
  );
}

// ── Human player area ────────────────────────────────────────────────────────

interface HumanAreaProps {
  player: SevensPlayerData;
  isCurrentTurn: boolean;
  tableMinVals: number[];
  tableMaxVals: number[];
  onPlay: (idx: number) => void;
}

function HumanArea({ player, isCurrentTurn, tableMinVals, tableMaxVals, onPlay }: HumanAreaProps) {
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {};
  return (
    <div className={playerAreaBaseClass} style={conditionalStyle}>
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
            {player.rank}位
          </span>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {player.cardCount}枚　パス: {player.passesUsed}/{player.maxPasses}
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>出せるカードをクリック</span>}
        </div>
      )}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
        {player.cards?.map((card, i) => {
          const playable = isCurrentTurn && isCardPlayable(card, tableMinVals, tableMaxVals);
          return (
            <button
              key={i}
              type="button"
              disabled={!playable}
              onClick={() => onPlay(i)}
              title={playable ? `出す: ${card.design} ${valueName(card.value)}` : undefined}
              style={{
                background: 'none',
                padding: 0,
                cursor: playable ? 'pointer' : 'default',
                borderRadius: 8,
                border: playable ? '3px solid #5cb85c' : '3px solid transparent',
                opacity: isCurrentTurn && !playable ? 0.5 : 1,
                boxSizing: 'border-box',
              }}
            >
              <CardImage card={card} style={{ width: 60 }} />
            </button>
          );
        })}
      </div>
    </div>
  );
}

// ── Main page ────────────────────────────────────────────────────────────────

export function SevensPage() {
  const [state, setState] = useState<SevensResponse | null>(null);

  const exec = useCallback(async (command: 'reset' | 'play', index = -1) => {
    try {
      const res = await sevensApi.exec(command, index);
      setState(res);
    } catch {
      console.error('sevens request failed');
    }
  }, []);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  if (!state) return null;

  const isHumanTurn = !state.gameEndFlag && state.currentTurn === 0;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const canPass = isHumanTurn && (humanPlayer?.passesUsed ?? 0) < (humanPlayer?.maxPasses ?? 5);

  return (
    <div className="bg-[#1a5c1a] rounded-2xl p-5 my-2.5 mx-auto max-w-[1000px]">
      {/* CPU row */}
      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 12 }}>
        {cpuPlayers.map((player) => (
          <CpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
        ))}
      </div>

      {/* Board */}
      <Board tableMinVals={state.tableMinVals} tableMaxVals={state.tableMaxVals} />

      <div style={{ borderTop: '1px solid rgba(255,255,255,0.2)', margin: '12px 0' }} />

      {/* Human player */}
      {humanPlayer && (
        <HumanArea
          player={humanPlayer}
          isCurrentTurn={isHumanTurn}
          tableMinVals={state.tableMinVals}
          tableMaxVals={state.tableMaxVals}
          onPlay={(idx) => exec('play', idx)}
        />
      )}

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
          {actionDesc(state.humanAction)}
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
          {['[CPUの行動]', ...state.cpuActions.map(actionDesc)].join('\n')}
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
            padding: '12px 20px',
            fontSize: '1.3em',
            fontWeight: 'bold',
            margin: '10px 0',
          }}
        >
          {state.message}
        </div>
      )}

      {/* Buttons */}
      <div className="text-center mt-3.5 mb-1">
        <button type="button" className={`${btnPrimary} min-w-[90px]`} onClick={() => exec('reset')}>
          リセット
        </button>
        <button
          type="button"
          className={`${btnWarning} min-w-[90px]`}
          disabled={!canPass}
          onClick={() => exec('play', -1)}
        >
          パス
        </button>
      </div>
    </div>
  );
}
