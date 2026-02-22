import { useCallback, useEffect, useState } from 'react';
import { sevensApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import type { Card, CardDesign, SevensAction, SevensPlayerData, SevensResponse } from '../types/card';
import { findPlayerName, playerName } from '../utils/playerUtils';

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

function actionDesc(players: { id: number; isHuman: boolean }[], action: SevensAction): string {
  if (!action.playedCard) return `${findPlayerName(players, action.playerIdx)}がパスしました`;
  const c = action.playedCard;
  return `${findPlayerName(players, action.playerIdx)}が出しました: ${c.design} ${valueName(c.value)}`;
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
    <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
      <div className="text-white font-bold mb-2">ボード</div>
      <div className="grid grid-cols-2 gap-2">
        {SUITS.map(({ idx, name, label, color }) => {
          const min = tableMinVals[idx] ?? 7;
          const max = tableMaxVals[idx] ?? 7;
          return (
            <div key={name} className="bg-white/[0.08] rounded-lg py-1.5 px-2.5 flex items-center gap-2">
              <span style={{ color, fontWeight: 'bold', fontSize: '1.1em', minWidth: 18 }}>{label}</span>
              <div className="flex flex-wrap gap-[3px]">
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
  return (
    <div className={playerAreaBaseClass} style={conditionalStyle}>
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
        <div className="text-[#ccc] text-[0.85em]">
          {player.cardCount}枚　パス: {player.passesUsed}/{player.maxPasses}
        </div>
      )}
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
            {player.rank}位
          </span>
        )}
      </div>
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em] mb-1">
          {player.cardCount}枚　パス: {player.passesUsed}/{player.maxPasses}
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>出せるカードをクリック</span>}
        </div>
      )}
      <div className="flex flex-wrap gap-1">
        {player.cards?.map((card, i) => {
          const playable = isCurrentTurn && isCardPlayable(card, tableMinVals, tableMaxVals);
          return (
            <button
              key={`${card.design}-${card.value}`}
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
              <CardImage card={card} width={52} />
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

  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const canPass = isHumanTurn && (humanPlayer?.passesUsed ?? 0) < (humanPlayer?.maxPasses ?? 5);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a5c1a]">
      {/* Scrollable: CPU rows + board + action logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* CPU row */}
        <div className="flex gap-2.5 flex-wrap mb-2.5">
          {cpuPlayers.map((player) => (
            <CpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        {/* Board */}
        <Board tableMinVals={state.tableMinVals} tableMaxVals={state.tableMaxVals} />

        {/* Human action log */}
        {state.humanAction && (
          <div className="bg-black/40 rounded-lg text-[#cfc] py-2 px-3.5 my-2 text-[0.85em]">
            {actionDesc(state.players, state.humanAction)}
          </div>
        )}

        {/* CPU action log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-[#ccc] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
            {['[CPUの行動]', ...state.cpuActions.map((a) => actionDesc(state.players, a))].join('\n')}
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
            <HumanArea
              player={humanPlayer}
              isCurrentTurn={isHumanTurn}
              tableMinVals={state.tableMinVals}
              tableMaxVals={state.tableMaxVals}
              onPlay={(idx) => exec('play', idx)}
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
            disabled={!canPass}
            onClick={() => exec('play', -1)}
          >
            パス
          </button>
        </div>
      </div>
    </div>
  );
}
