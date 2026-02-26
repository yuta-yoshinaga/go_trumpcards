import { useCallback, useEffect, useRef, useState } from 'react';
import { doubtApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import type { Card, DoubtConfig, DoubtCpuAction, DoubtPlayerData, DoubtResponse } from '../types/card';
import { playerName } from '../utils/playerUtils';

function valueName(v: number): string {
  if (v === 1) return 'A';
  if (v === 11) return 'J';
  if (v === 12) return 'Q';
  if (v === 13) return 'K';
  return String(v);
}

// ── CPU player area ──────────────────────────────────────────────────────────

interface CpuAreaProps {
  player: DoubtPlayerData;
  isCurrentTurn: boolean;
}

function CpuArea({ player, isCurrentTurn }: CpuAreaProps) {
  const conditionalStyle: React.CSSProperties = isCurrentTurn
    ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
    : {};
  return (
    <div
      className="bg-black/35 rounded-[10px] p-[10px] border-2 border-transparent flex-[1_1_150px] min-w-[120px]"
      style={conditionalStyle}
    >
      <div className="text-white font-bold mb-1 text-sm">
        {playerName(player.id, player.isHuman)}
        {player.isFinished && (
          <span
            style={{
              background: '#5cb85c',
              color: '#fff',
              borderRadius: 6,
              padding: '1px 6px',
              marginLeft: 6,
              fontSize: '0.8em',
            }}
          >
            上がり
          </span>
        )}
        {isCurrentTurn && !player.isFinished && (
          <span
            style={{
              background: '#f0ad4e',
              color: '#222',
              borderRadius: 6,
              padding: '1px 6px',
              marginLeft: 6,
              fontSize: '0.8em',
              fontWeight: 'bold',
            }}
          >
            考え中...
          </span>
        )}
      </div>
      <div className="text-[#ccc] text-[0.85em]">{player.cardCount}枚</div>
    </div>
  );
}

// ── Human hand card ───────────────────────────────────────────────────────────

interface HandCardProps {
  card: Card;
  index: number;
  selected: boolean;
  selectable: boolean;
  onToggle: (idx: number) => void;
}

function HandCard({ card, index, selected, selectable, onToggle }: HandCardProps) {
  return (
    <button
      type="button"
      disabled={!selectable}
      onClick={() => onToggle(index)}
      style={{
        background: 'none',
        padding: 0,
        cursor: selectable ? 'pointer' : 'default',
        borderRadius: 8,
        border: selected ? '3px solid #5cb85c' : '3px solid transparent',
        transform: selected ? 'translateY(-8px)' : 'none',
        transition: 'transform 0.15s, border 0.15s',
        opacity: !selectable ? 0.5 : 1,
        boxSizing: 'border-box',
      }}
    >
      <CardImage card={card} width={52} />
    </button>
  );
}

// ── Action info ───────────────────────────────────────────────────────────────

function actionDesc(action: DoubtCpuAction, players: DoubtPlayerData[]): string {
  const p = players[action.playerIdx];
  const name = p ? playerName(p.id, p.isHuman) : `Player ${action.playerIdx}`;
  return `${name}が${action.cardCount}枚出しました (宣言: ${valueName(action.claimedValue)})`;
}

// ── Main page ────────────────────────────────────────────────────────────────

const DEFAULT_DOUBT_CONFIG: DoubtConfig = { doubtWindowSec: 10, cpuMemoryLevel: 1 };

const DOUBT_WINDOW_OPTIONS = [
  { value: 3, label: '3秒' },
  { value: 5, label: '5秒' },
  { value: 10, label: '10秒' },
] as const;

const CPU_MEMORY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

export function DoubtPage() {
  const [state, setState] = useState<DoubtResponse | null>(null);
  const [selectedCardIndices, setSelectedCardIndices] = useState<number[]>([]);
  const [claimedValue, setClaimedValue] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [countdown, setCountdown] = useState<number | null>(null);
  const [doubtConfig, setDoubtConfig] = useState<DoubtConfig>(DEFAULT_DOUBT_CONFIG);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const autoSkipRef = useRef(false);
  const cpuDoubtersRef = useRef<number[]>([]);

  // Keep cpuDoubters ref current so the auto-skip effect avoids stale state
  useEffect(() => {
    if (state) cpuDoubtersRef.current = state.cpuDoubters;
  }, [state]);

  const stopCountdown = useCallback(() => {
    autoSkipRef.current = false;
    if (countdownRef.current !== null) {
      clearInterval(countdownRef.current);
      countdownRef.current = null;
    }
    setCountdown(null);
  }, []);

  const startCountdown = useCallback(
    (sec: number) => {
      stopCountdown();
      setCountdown(sec);
      countdownRef.current = setInterval(() => {
        setCountdown((prev) => {
          const cur = prev as number;
          if (cur <= 1) {
            clearInterval(countdownRef.current as ReturnType<typeof setInterval>);
            countdownRef.current = null;
            autoSkipRef.current = true;
            return null;
          }
          return cur - 1;
        });
      }, 1000);
    },
    [stopCountdown],
  );

  const exec = useCallback(
    async (
      command: 'reset' | 'play' | 'doubt' | 'skip',
      cardIndices?: number[],
      cv?: number,
      doubterIndices?: number[],
      config?: DoubtConfig,
    ) => {
      setLoading(true);
      stopCountdown();
      try {
        setError(null);
        const res = await doubtApi.exec(command, cardIndices, cv, doubterIndices, config);
        setState(res);
        setSelectedCardIndices([]);
        setClaimedValue(1);
      } catch {
        setError('通信エラーが発生しました。もう一度お試しください。');
      } finally {
        setLoading(false);
      }
    },
    [stopCountdown],
  );

  const handleConfigChange = useCallback((key: keyof DoubtConfig, value: string) => {
    const parsed = Number(value);
    if (!Number.isNaN(parsed)) {
      setDoubtConfig((prev) => ({ ...prev, [key]: parsed }));
    }
  }, []);

  useEffect(() => {
    exec('reset', undefined, undefined, undefined, DEFAULT_DOUBT_CONFIG);
  }, [exec]);

  // Auto-skip when countdown timer expires without user interaction (mirrors CLI timeout behaviour)
  useEffect(() => {
    if (countdown !== null) return;
    if (!autoSkipRef.current) return;
    autoSkipRef.current = false;
    exec('skip', undefined, undefined, cpuDoubtersRef.current);
  }, [countdown, exec]);

  // Start countdown when it's the doubt phase and a CPU played (human needs to decide)
  useEffect(() => {
    if (!state) return;
    if (state.phase === 1 && state.lastAction !== null) {
      const lastActionPlayer = state.players[state.lastAction.playerIdx];
      if (lastActionPlayer && !lastActionPlayer.isHuman) {
        startCountdown(state.doubtWindowSec);
      }
    }
  }, [state, startCountdown]);

  if (!state) return null;

  const isHumanTurn = !state.gameEndFlag && state.players[state.currentTurn]?.isHuman === true;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const isDoubtPhase = state.phase === 1;
  const cpuPlayed = isDoubtPhase && state.lastAction !== null && !state.players[state.lastAction.playerIdx]?.isHuman;

  const toggleCard = (idx: number) => {
    setSelectedCardIndices((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  const handlePlay = () => {
    exec('play', selectedCardIndices, claimedValue);
  };

  const handleDoubt = () => {
    stopCountdown();
    exec('doubt', undefined, undefined, [0, ...state.cpuDoubters]);
  };

  const handleSkip = () => {
    stopCountdown();
    exec('skip', undefined, undefined, state.cpuDoubters);
  };

  const handleCpuDoubtConfirm = () => {
    exec('doubt', undefined, undefined, state.cpuDoubters);
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a2c5c]">
      {/* Settings panel */}
      <details className="px-4 pt-2">
        <summary className="text-white/70 text-xs cursor-pointer select-none">設定</summary>
        <div className="bg-black/30 rounded-lg p-3 mt-1 flex flex-wrap gap-4 text-sm text-white">
          <label className="flex items-center gap-2">
            ダウト時間:
            <select
              className="bg-black/50 text-white rounded px-2 py-1 border border-white/30"
              value={doubtConfig.doubtWindowSec}
              onChange={(e) => handleConfigChange('doubtWindowSec', e.target.value)}
            >
              {DOUBT_WINDOW_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </label>
          <label className="flex items-center gap-2">
            CPU記憶力:
            <select
              className="bg-black/50 text-white rounded px-2 py-1 border border-white/30"
              value={doubtConfig.cpuMemoryLevel}
              onChange={(e) => handleConfigChange('cpuMemoryLevel', e.target.value)}
            >
              {CPU_MEMORY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </label>
        </div>
      </details>

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* CPU player areas */}
        <div className="flex gap-2 flex-wrap mb-3">
          {cpuPlayers.map((player) => (
            <CpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
          ))}
        </div>

        {/* Table area */}
        <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
          <div className="text-white font-bold mb-1">テーブル</div>
          <div className="text-[#ccc] text-[0.9em]">場のカード: {state.tableCardCount}枚</div>
          {state.lastAction && (
            <div className="text-yellow-300 text-[0.85em] mt-1">{actionDesc(state.lastAction, state.players)}</div>
          )}
        </div>

        {/* Doubt/Skip UI */}
        {isDoubtPhase && !state.gameEndFlag && (
          <div className="bg-black/40 rounded-[10px] py-3 px-4 my-2">
            {cpuPlayed ? (
              <>
                <div className="text-white font-bold mb-2">ダウトしますか？</div>
                {countdown !== null && (
                  <div className="text-yellow-300 text-lg font-bold mb-2">残り {countdown} 秒</div>
                )}
                {state.cpuDoubters.length > 0 && (
                  <div className="text-[#ccc] text-[0.85em] mb-2">
                    ダウト宣言CPU: {state.cpuDoubters.map((idx) => playerName(idx, false)).join(', ')}
                  </div>
                )}
                <div className="flex gap-2">
                  <button type="button" className={btnDanger} disabled={loading} onClick={handleDoubt}>
                    ダウト！
                  </button>
                  <button type="button" className={btnWarning} disabled={loading} onClick={handleSkip}>
                    スルー
                  </button>
                </div>
              </>
            ) : (
              <>
                <div className="text-white font-bold mb-2">CPUがダウトを判定中...</div>
                {state.cpuDoubters.length > 0 && (
                  <div className="text-red-300 text-[0.9em] mb-2">
                    ダウト！ {state.cpuDoubters.map((idx) => playerName(idx, false)).join(', ')}
                  </div>
                )}
                <button type="button" className={btnPrimary} disabled={loading} onClick={handleCpuDoubtConfirm}>
                  確認
                </button>
              </>
            )}
          </div>
        )}

        {/* Last doubt result */}
        {state.lastDoubtResult && (
          <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-[0.85em]">
            <div className="text-white font-bold mb-1">ダウト結果</div>
            <div className={state.lastDoubtResult.wasLying ? 'text-red-300' : 'text-green-300'}>
              {state.lastDoubtResult.wasLying ? 'ウソでした！' : '本当でした！'}
            </div>
            <div className="text-[#ccc]">
              {playerName(
                state.players[state.lastDoubtResult.loserIdx]?.id ?? state.lastDoubtResult.loserIdx,
                state.players[state.lastDoubtResult.loserIdx]?.isHuman ?? false,
              )}
              が{state.lastDoubtResult.cardCount}枚引き取りました
            </div>
            {state.lastDoubtResult.revealedCards.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-1">
                {state.lastDoubtResult.revealedCards.map((card, i) => (
                  <CardImage key={`${card.design}-${card.value}-${i}`} card={card} width={36} />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Human/CPU action logs */}
        {state.humanAction && !isDoubtPhase && (
          <div className="bg-black/40 rounded-lg text-[#cfc] py-2 px-3.5 my-2 text-[0.85em]">
            {actionDesc(state.humanAction, state.players)}
          </div>
        )}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-[#ccc] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
            {['[CPUの行動]', ...state.cpuActions.map((a) => actionDesc(a, state.players))].join('\n')}
          </div>
        )}

        {/* Result message */}
        {state.message && (
          <div className="bg-black/55 rounded-[10px] text-white text-center py-2.5 px-4 text-[1.2em] font-bold my-2">
            {state.message}
          </div>
        )}
      </div>

      {/* Sticky footer: human player hand + action buttons */}
      <div
        className="shrink-0 bg-[#101c3a] border-t border-white/20 px-4 py-2.5"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 10px)' }}
      >
        {/* Human player info */}
        {humanPlayer && (
          <div className="mb-2">
            <div className="text-white font-bold text-sm mb-1">
              あなた ({humanPlayer.cardCount}枚)
              {isHumanTurn && state.phase === 0 && (
                <span className="text-green-400 text-xs ml-2">カードを選んで出してください</span>
              )}
            </div>
            {/* Human cards */}
            <div className="flex flex-wrap gap-1">
              {humanPlayer.cards?.map((card, i) => (
                <HandCard
                  key={`${card.design}-${card.value}-${i}`}
                  card={card}
                  index={i}
                  selected={selectedCardIndices.includes(i)}
                  selectable={isHumanTurn && state.phase === 0 && !loading}
                  onToggle={toggleCard}
                />
              ))}
            </div>

            {/* Claimed value input (shown when cards are selected) */}
            {selectedCardIndices.length > 0 && isHumanTurn && state.phase === 0 && (
              <div className="mt-2 flex items-center gap-2">
                <span className="text-white text-sm">宣言する値:</span>
                <input
                  type="number"
                  min={1}
                  max={13}
                  value={claimedValue}
                  onChange={(e) => {
                    const num = Number(e.target.value);
                    setClaimedValue(Math.max(1, Math.min(13, num)));
                  }}
                  className="bg-black/50 text-white rounded px-2 py-1 w-16 text-sm border border-white/30"
                />
                <span className="text-[#ccc] text-xs">({valueName(claimedValue)})</span>
              </div>
            )}
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Action buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => exec('reset', undefined, undefined, undefined, doubtConfig)}
          >
            リセット
          </button>
          {isHumanTurn && state.phase === 0 && (
            <button
              type="button"
              className={`${btnSuccess} min-w-[90px]`}
              disabled={loading || selectedCardIndices.length === 0}
              onClick={handlePlay}
            >
              出す
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
