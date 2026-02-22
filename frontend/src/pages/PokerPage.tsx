import { useCallback, useEffect, useState } from 'react';
import { pokerApi } from '../api/gameApi';
import { CardBack, CardImage } from '../components/CardImage';
import type { PokerResponse } from '../types/card';

const PHASE_INIT = 0;
const PHASE_DEAL = 1;
const PHASE_END = 2;

const btnPrimary =
  'px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnWarning =
  'px-3 py-1.5 text-sm font-medium text-gray-900 bg-yellow-400 rounded hover:bg-yellow-500 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';
const btnSuccess =
  'px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5';

const cardWrapBase: React.CSSProperties = {
  position: 'relative',
  cursor: 'pointer',
  transition: 'transform 0.15s',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
};

export function PokerPage() {
  const [state, setState] = useState<PokerResponse | null>(null);
  const [selected, setSelected] = useState<number[]>([]);

  const exec = useCallback(async (command: 'reset' | 'exchange' | 'stand', indices?: number[]) => {
    try {
      const res = await pokerApi.exec(command, indices);
      setState(res);
      setSelected([]);
    } catch {
      console.error('poker request failed');
    }
  }, []);

  useEffect(() => {
    exec('reset');
  }, [exec]);

  const phase = state?.phase ?? PHASE_INIT;

  const toggleSelect = (idx: number) => {
    if (phase !== PHASE_DEAL) return;
    setSelected((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a6b1a]">
      {/* Scrollable: dealer area */}
      <div className="flex-1 overflow-y-auto pt-4 px-5">
        <div className="mb-2">
          <div className="text-white text-[1.1em] mb-1.5">
            ディーラー手札
            {phase === PHASE_END && state?.dealer?.handName && (
              <span
                style={{
                  display: 'inline-block',
                  background: '#f0ad4e',
                  color: '#222',
                  fontWeight: 'bold',
                  borderRadius: 8,
                  padding: '2px 12px',
                  marginLeft: 8,
                  fontSize: '0.95em',
                }}
              >
                {state.dealer.handName}
              </span>
            )}
          </div>
          <div className="flex flex-wrap gap-2 mb-2.5">
            {phase === PHASE_END && state?.dealer?.cards?.length
              ? state.dealer.cards.map((card) => (
                  <div key={`${card.design}-${card.value}`} style={{ ...cardWrapBase, cursor: 'default' }}>
                    <CardImage card={card} width={60} style={{ border: '3px solid transparent' }} />
                  </div>
                ))
              : Array.from({ length: 5 }).map((_, i) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: placeholder array with no card identity
                  <CardBack key={i} width={60} />
                ))}
          </div>
        </div>
      </div>

      {/* Sticky footer: player hand + result + buttons */}
      <div
        className="shrink-0 bg-[#155715] border-t border-white/20 px-5 py-3"
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      >
        {/* Player area */}
        <div>
          <div className="text-white text-[1.1em] mb-1">
            プレイヤー手札
            {phase === PHASE_END && state?.player?.handName && (
              <span
                style={{
                  display: 'inline-block',
                  background: '#f0ad4e',
                  color: '#222',
                  fontWeight: 'bold',
                  borderRadius: 8,
                  padding: '2px 12px',
                  marginLeft: 8,
                  fontSize: '0.95em',
                }}
              >
                {state.player.handName}
              </span>
            )}
          </div>
          {phase === PHASE_DEAL && (
            <div className="text-[#cfc] text-[0.85em] mb-1">
              交換したいカードをクリックして選択し、「交換」または「スタンド」を押してください。
            </div>
          )}
          <div className="flex flex-wrap gap-1.5 mb-2">
            {state?.player?.cards?.map((card, i) => {
              const isSelected = selected.includes(i);
              return (
                <button
                  key={`${card.design}-${card.value}`}
                  type="button"
                  onClick={() => toggleSelect(i)}
                  style={{
                    background: 'none',
                    border: 'none',
                    padding: 0,
                    ...cardWrapBase,
                    cursor: phase === PHASE_DEAL ? 'pointer' : 'default',
                  }}
                >
                  <CardImage
                    card={card}
                    width={60}
                    style={{
                      border: isSelected ? '3px solid #f0ad4e' : '3px solid transparent',
                      transform: isSelected ? 'translateY(-10px)' : undefined,
                      transition: 'transform 0.15s',
                    }}
                  />
                  <div
                    style={{
                      color: '#f0ad4e',
                      fontSize: '0.75em',
                      fontWeight: 'bold',
                      visibility: isSelected ? 'visible' : 'hidden',
                    }}
                  >
                    交換
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Result */}
        <div className="bg-black/55 rounded-lg text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2 min-h-[36px]">
          {state?.message ?? ''}
        </div>

        {/* Buttons */}
        <div className="text-center">
          <button type="button" className={`${btnPrimary} min-w-[90px]`} onClick={() => exec('reset')}>
            リセット
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[90px]`}
            disabled={phase !== PHASE_DEAL}
            onClick={() => exec('exchange', selected)}
          >
            交換
          </button>
          <button
            type="button"
            className={`${btnSuccess} min-w-[90px]`}
            disabled={phase !== PHASE_DEAL}
            onClick={() => exec('stand')}
          >
            スタンド
          </button>
        </div>
      </div>
    </div>
  );
}
