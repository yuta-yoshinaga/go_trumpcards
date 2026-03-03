import { useCallback, useEffect, useState } from 'react';
import { sevensApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CpuTurnArea } from '../components/CpuTurnArea';
import { ErrorAlert } from '../components/ErrorAlert';
import { StatusBadge } from '../components/StatusBadge';
import { useGameApi } from '../hooks/useGameApi';
import { btnPrimary, btnWarning } from '../styles/buttonStyles';
import { playerAreaBase } from '../styles/gameStyles';
import type { Card, CardDesign, SevensAction, SevensPlayerData, SevensResponse } from '../types/card';
import { suitName, valueName } from '../utils/cardUtils';
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

function isPositionPlaced(tablePlaced: number[], suit: number, value: number): boolean {
  return (tablePlaced[suit] & (1 << value)) !== 0;
}

function isPositionPlayable(tablePlaced: number[], suit: number, value: number, tunnelEnabled: boolean): boolean {
  if (isPositionPlaced(tablePlaced, suit, value)) return false;
  if (isPositionPlaced(tablePlaced, suit, value + 1)) return true;
  if (isPositionPlaced(tablePlaced, suit, value - 1)) return true;
  if (tunnelEnabled) {
    if (value === 1 && isPositionPlaced(tablePlaced, suit, 13)) return true;
    if (value === 13 && isPositionPlaced(tablePlaced, suit, 1)) return true;
  }
  return false;
}

function hasAnyPlayablePosition(tablePlaced: number[], tunnelEnabled: boolean): boolean {
  for (let suit = 1; suit <= 4; suit++) {
    for (let v = 1; v <= 13; v++) {
      if (isPositionPlayable(tablePlaced, suit, v, tunnelEnabled)) return true;
    }
  }
  return false;
}

function hasOnlyJokers(cards: Card[]): boolean {
  return cards.length > 0 && cards.every((c) => c.design === 'JOKER');
}

function isCardPlayable(
  card: Card,
  tablePlaced: number[],
  tunnelEnabled: boolean,
  noJokerFinish: boolean,
  allCards: Card[],
): boolean {
  if (card.design === 'JOKER') {
    if (noJokerFinish && hasOnlyJokers(allCards)) return false;
    return hasAnyPlayablePosition(tablePlaced, tunnelEnabled);
  }
  const suit = designToSuit[card.design];
  return isPositionPlayable(tablePlaced, suit, card.value, tunnelEnabled);
}

function actionDesc(players: { id: number; isHuman: boolean }[], action: SevensAction): string {
  if (!action.playedCard) {
    const base = `${findPlayerName(players, action.playerIdx)}がパスしました`;
    return action.forcedPass ? `${base} ⚠ 出せるカードなし!` : base;
  }
  const c = action.playedCard;
  let desc = `${findPlayerName(players, action.playerIdx)}が出しました: ${c.design} ${valueName(c.value)}`;
  if (c.design === 'JOKER' && action.targetSuit > 0) {
    desc += ` → ${suitName(action.targetSuit)} ${valueName(action.targetValue)}`;
  }
  return desc;
}

// ── styles ──────────────────────────────────────────────────────────────────

const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_180px] min-w-[150px]`;

// ── Board component ──────────────────────────────────────────────────────────

interface BoardProps {
  tablePlaced: number[];
  tunnelEnabled: boolean;
  jokerSelecting: boolean;
  onJokerPlace?: (suit: number, value: number) => void;
}

function Board({ tablePlaced, tunnelEnabled, jokerSelecting, onJokerPlace }: BoardProps) {
  return (
    <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
      <div className="text-white font-bold mb-2">
        ボード
        {tunnelEnabled && <span className="text-yellow-400 text-xs ml-2">[トンネル]</span>}
        {jokerSelecting && <span className="text-green-400 text-xs ml-2">配置先を選択してください</span>}
      </div>
      <div className="grid grid-cols-2 gap-2">
        {SUITS.map(({ idx, name, label, color }) => (
          <div key={name} className="bg-white/[0.08] rounded-lg py-1.5 px-2.5 flex items-center gap-2">
            <span style={{ color, fontWeight: 'bold', fontSize: '1.1em', minWidth: 18 }}>{label}</span>
            <div className="flex flex-wrap gap-[3px]">
              {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => {
                const placed = isPositionPlaced(tablePlaced, idx, v);
                const isCenter = v === 7;
                const canPlace = jokerSelecting && isPositionPlayable(tablePlaced, idx, v, tunnelEnabled);
                const cellStyle: React.CSSProperties = {
                  display: 'inline-block',
                  width: 22,
                  height: 22,
                  lineHeight: '22px',
                  textAlign: 'center',
                  borderRadius: 4,
                  fontSize: '0.7em',
                  fontWeight: isCenter ? 'bold' : 'normal',
                  background: canPlace
                    ? '#3b82f6'
                    : placed
                      ? isCenter
                        ? '#f0ad4e'
                        : '#5cb85c'
                      : 'rgba(255,255,255,0.1)',
                  color: canPlace ? '#fff' : placed ? '#000' : '#555',
                };
                if (canPlace) {
                  return (
                    <button
                      key={v}
                      type="button"
                      onClick={() => onJokerPlace?.(idx, v)}
                      aria-label={`${suitName(idx)} ${valueName(v)} に配置`}
                      style={{ ...cellStyle, border: '1px solid #60a5fa', cursor: 'pointer', padding: 0 }}
                    >
                      {valueName(v)}
                    </button>
                  );
                }
                return (
                  <span key={v} style={cellStyle}>
                    {valueName(v)}
                  </span>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ── CPU player area ──────────────────────────────────────────────────────────

function SevCpuArea({ player, isCurrentTurn }: { player: SevensPlayerData; isCurrentTurn: boolean }) {
  return (
    <CpuTurnArea
      playerId={player.id}
      isHuman={player.isHuman}
      isCurrentTurn={isCurrentTurn}
      isFinished={player.isFinished}
      finishedLabel={player.isFinished ? `${player.rank}位` : undefined}
      className={playerAreaClass}
    >
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em]">
          {player.cardCount}枚　パス: {player.passesUsed}/{player.maxPasses === 0 ? '∞' : player.maxPasses}
        </div>
      )}
    </CpuTurnArea>
  );
}

// ── Human player area ────────────────────────────────────────────────────────

interface HumanAreaProps {
  player: SevensPlayerData;
  isCurrentTurn: boolean;
  tablePlaced: number[];
  tunnelEnabled: boolean;
  noJokerFinish: boolean;
  loading: boolean;
  onPlay: (idx: number) => void;
}

function HumanArea({
  player,
  isCurrentTurn,
  tablePlaced,
  tunnelEnabled,
  noJokerFinish,
  loading,
  onPlay,
}: HumanAreaProps) {
  const conditionalStyle: React.CSSProperties = player.isFinished
    ? { opacity: 0.5 }
    : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {};
  return (
    <div className={playerAreaClass} style={conditionalStyle}>
      <div className="text-white font-bold mb-1">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && <StatusBadge variant="success">{player.rank}位</StatusBadge>}
      </div>
      {!player.isFinished && (
        <div className="text-[#ccc] text-[0.85em] mb-1">
          {player.cardCount}枚　パス: {player.passesUsed}/{player.maxPasses === 0 ? '∞' : player.maxPasses}
          {isCurrentTurn && <span style={{ marginLeft: 8, color: '#cfc' }}>出せるカードをクリック</span>}
        </div>
      )}
      <div className="flex flex-wrap gap-1">
        {player.cards?.map((card, i) => {
          const playable =
            isCurrentTurn && !loading && isCardPlayable(card, tablePlaced, tunnelEnabled, noJokerFinish, player.cards);
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
  const [jokerCardIdx, setJokerCardIdx] = useState<number | null>(null);
  const [cfgTunnel, setCfgTunnel] = useState(false);
  const [cfgJokerCount, setCfgJokerCount] = useState(0);
  const [cfgCpuStrategy, setCfgCpuStrategy] = useState(false);
  const [cfgMaxPasses, setCfgMaxPasses] = useState(5);
  const [cfgNoJokerFinish, setCfgNoJokerFinish] = useState(false);

  const onSuccess = useCallback((res: SevensResponse) => {
    setJokerCardIdx(null);
    setCfgTunnel(res.config.tunnelEnabled);
    setCfgJokerCount(res.config.jokerCount);
    setCfgCpuStrategy(res.config.cpuStrategy);
    setCfgMaxPasses(res.config.maxPasses);
    setCfgNoJokerFinish(res.config.noJokerFinish);
  }, []);
  const { state, loading, error, exec } = useGameApi(sevensApi.exec, { onSuccess });

  useEffect(() => {
    exec('reset');
  }, [exec]);

  if (!state) return null;

  const tablePlaced = state.tablePlaced;
  const tunnelEnabled = state.config.tunnelEnabled;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const canPass =
    isHumanTurn &&
    humanPlayer != null &&
    (humanPlayer.maxPasses === 0 || humanPlayer.passesUsed < humanPlayer.maxPasses);

  const handleCardPlay = (idx: number) => {
    const card = humanPlayer?.cards?.[idx];
    if (card?.design === 'JOKER') {
      setJokerCardIdx(idx);
    } else {
      exec('play', idx);
    }
  };

  const handleJokerPlace = (suit: number, value: number) => {
    exec('joker', jokerCardIdx as number, suit, value);
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a5c1a]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">処理中...</span>}
      {/* Scrollable: CPU rows + board + action logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Config rules */}
        {state.config &&
          (state.config.tunnelEnabled ||
            state.config.jokerCount > 0 ||
            state.config.cpuStrategy ||
            state.config.maxPasses !== 5 ||
            state.config.noJokerFinish) && (
            <div className="bg-black/30 rounded-lg text-yellow-300 py-1.5 px-3 mb-2 text-[0.85em]">
              ルール:
              {state.config.tunnelEnabled && ' [トンネル]'}
              {state.config.jokerCount > 0 && ` [ジョーカー×${state.config.jokerCount}]`}
              {state.config.cpuStrategy && ' [CPU戦略]'}
              {state.config.maxPasses === 0 && ' [パス無制限]'}
              {state.config.maxPasses !== 5 && state.config.maxPasses !== 0 && ` [パス${state.config.maxPasses}回]`}
              {state.config.noJokerFinish && ' [ジョーカー上がり禁止]'}
            </div>
          )}

        {/* CPU row */}
        <div className="flex gap-2.5 flex-wrap mb-2.5">
          {cpuPlayers.map((player) => (
            <SevCpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        {/* Board */}
        <Board
          tablePlaced={tablePlaced}
          tunnelEnabled={tunnelEnabled}
          jokerSelecting={jokerCardIdx !== null}
          onJokerPlace={handleJokerPlace}
        />

        {/* Human action log */}
        {state.humanAction && (
          <div
            data-testid={state.humanAction.forcedPass ? 'human-action-forced-pass' : 'human-action'}
            className={`rounded-lg py-2 px-3.5 my-2 text-[0.85em] ${state.humanAction.forcedPass ? 'bg-red-900/50 text-[#fca] border border-red-500/50' : 'bg-black/40 text-[#cfc]'}`}
          >
            {actionDesc(state.players, state.humanAction)}
          </div>
        )}

        {/* CPU action log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-[0.85em]">
            <span className="text-[#ccc]">[CPUの行動]</span>
            {state.cpuActions.map((a, i) => (
              <div
                key={`cpu-action-${a.playerIdx}-${i}`}
                data-testid={a.forcedPass ? `cpu-action-forced-pass-${i}` : `cpu-action-${i}`}
                className={a.forcedPass ? 'text-[#fca]' : 'text-[#ccc]'}
              >
                {actionDesc(state.players, a)}
              </div>
            ))}
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
              tablePlaced={tablePlaced}
              tunnelEnabled={tunnelEnabled}
              noJokerFinish={state.config.noJokerFinish}
              loading={loading}
              onPlay={handleCardPlay}
            />
          </div>
        )}

        {/* Config panel */}
        <div className="bg-black/30 rounded-lg py-1.5 px-3 mb-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-[0.85em] text-white/80">
          <span className="text-yellow-300 font-bold">ルール設定 (リセット時に適用)</span>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgTunnel} onChange={(e) => setCfgTunnel(e.target.checked)} />
            トンネル
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            ジョーカー
            <select
              value={cfgJokerCount}
              onChange={(e) => setCfgJokerCount(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={0}>0</option>
              <option value={1}>1</option>
              <option value={2}>2</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgCpuStrategy} onChange={(e) => setCfgCpuStrategy(e.target.checked)} />
            CPU戦略
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            パス回数
            <select
              value={cfgMaxPasses}
              onChange={(e) => setCfgMaxPasses(Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1 py-0.5"
            >
              <option value={3}>3</option>
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={0}>無制限</option>
            </select>
          </label>
          <label className="flex items-center gap-1 cursor-pointer">
            <input type="checkbox" checked={cfgNoJokerFinish} onChange={(e) => setCfgNoJokerFinish(e.target.checked)} />
            ジョーカー上がり禁止
          </label>
        </div>

        <ErrorAlert message={error} />

        {/* Buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() =>
              exec('reset', -1, 0, 0, {
                tunnelEnabled: cfgTunnel,
                jokerCount: cfgJokerCount,
                cpuStrategy: cfgCpuStrategy,
                maxPasses: cfgMaxPasses,
                noJokerFinish: cfgNoJokerFinish,
              })
            }
          >
            リセット
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[90px]`}
            disabled={loading || !canPass}
            onClick={() => exec('play', -1)}
          >
            パス
          </button>
          {jokerCardIdx !== null && (
            <button type="button" className={`${btnWarning} min-w-[90px]`} onClick={() => setJokerCardIdx(null)}>
              キャンセル
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
