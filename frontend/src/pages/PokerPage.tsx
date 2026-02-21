import { useCallback, useEffect, useState } from 'react'
import { pokerApi } from '../api/gameApi'
import { CardImage, CardBack } from '../components/CardImage'
import type { PokerResponse } from '../types/card'

const PHASE_INIT = 0
const PHASE_DEAL = 1
const PHASE_END = 2

const btnPrimary = 'px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5'
const btnWarning = 'px-3 py-1.5 text-sm font-medium text-gray-900 bg-yellow-400 rounded hover:bg-yellow-500 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5'
const btnSuccess = 'px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed mx-1.5'

const cardWrapBase: React.CSSProperties = {
  position: 'relative',
  cursor: 'pointer',
  transition: 'transform 0.15s',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
}

export function PokerPage() {
  const [state, setState] = useState<PokerResponse | null>(null)
  const [selected, setSelected] = useState<number[]>([])

  const exec = useCallback(async (command: 'reset' | 'exchange' | 'stand', indices?: number[]) => {
    try {
      const res = await pokerApi.exec(command, indices)
      setState(res)
      setSelected([])
    } catch {
      console.error('poker request failed')
    }
  }, [])

  useEffect(() => { exec('reset') }, [exec])

  const phase = state?.phase ?? PHASE_INIT

  const toggleSelect = (idx: number) => {
    if (phase !== PHASE_DEAL) return
    setSelected(prev =>
      prev.includes(idx) ? prev.filter(i => i !== idx) : [...prev, idx]
    )
  }

  return (
    <div className="bg-[#1a6b1a] rounded-2xl p-5 my-2.5 mx-auto max-w-[900px]">
      {/* Dealer area */}
      <div style={{ marginBottom: 8 }}>
        <div style={{ color: '#fff', fontSize: '1.1em', marginBottom: 6 }}>
          ディーラー手札
          {phase === PHASE_END && state?.dealer?.handName && (
            <span style={{
              display: 'inline-block',
              background: '#f0ad4e',
              color: '#222',
              fontWeight: 'bold',
              borderRadius: 8,
              padding: '2px 12px',
              marginLeft: 8,
              fontSize: '0.95em',
            }}>
              {state.dealer.handName}
            </span>
          )}
        </div>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 10 }}>
          {phase === PHASE_END && state?.dealer?.cards?.length
            ? state.dealer.cards.map((card, i) => (
              <div key={i} style={{ ...cardWrapBase, cursor: 'default' }}>
                <CardImage card={card} style={{ border: '3px solid transparent' }} />
              </div>
            ))
            : Array.from({ length: 5 }).map((_, i) => (
              <CardBack key={i} />
            ))
          }
        </div>
      </div>

      <div style={{ borderTop: '1px solid rgba(255,255,255,0.2)', margin: '14px 0' }} />

      {/* Player area */}
      <div>
        <div style={{ color: '#fff', fontSize: '1.1em', marginBottom: 6 }}>
          プレイヤー手札
          {phase === PHASE_END && state?.player?.handName && (
            <span style={{
              display: 'inline-block',
              background: '#f0ad4e',
              color: '#222',
              fontWeight: 'bold',
              borderRadius: 8,
              padding: '2px 12px',
              marginLeft: 8,
              fontSize: '0.95em',
            }}>
              {state.player.handName}
            </span>
          )}
        </div>
        {phase === PHASE_DEAL && (
          <div style={{ color: '#cfc', fontSize: '0.9em', marginBottom: 4 }}>
            交換したいカードをクリックして選択し、「交換」または「スタンド」を押してください。
          </div>
        )}
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 10 }}>
          {state?.player?.cards?.map((card, i) => {
            const isSelected = selected.includes(i)
            return (
              <div
                key={i}
                onClick={() => toggleSelect(i)}
                style={{
                  ...cardWrapBase,
                  cursor: phase === PHASE_DEAL ? 'pointer' : 'default',
                }}
              >
                <img
                  src={`/images/${cardPrefix(card.design)}${String(card.value).padStart(2, '0')}.png`}
                  alt={`${card.design} ${card.value}`}
                  style={{
                    width: 80,
                    borderRadius: 6,
                    border: isSelected ? '3px solid #f0ad4e' : '3px solid transparent',
                    transform: isSelected ? 'translateY(-12px)' : undefined,
                    display: 'block',
                    transition: 'transform 0.15s',
                  }}
                />
                <div style={{
                  color: '#f0ad4e',
                  fontSize: '0.75em',
                  fontWeight: 'bold',
                  visibility: isSelected ? 'visible' : 'hidden',
                }}>
                  交換
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Result */}
      <div style={{
        background: 'rgba(0,0,0,0.55)',
        borderRadius: 10,
        color: '#fff',
        textAlign: 'center',
        padding: '12px 20px',
        fontSize: '1.3em',
        fontWeight: 'bold',
        margin: '10px 0',
        minHeight: 48,
      }}>
        {state?.message ?? ''}
      </div>

      {/* Buttons */}
      <div className="text-center mt-3.5 mb-1">
        <button
          className={`${btnPrimary} min-w-[90px]`}
          onClick={() => exec('reset')}
        >
          リセット
        </button>
        <button
          className={`${btnWarning} min-w-[90px]`}
          disabled={phase !== PHASE_DEAL}
          onClick={() => exec('exchange', selected)}
        >
          交換
        </button>
        <button
          className={`${btnSuccess} min-w-[90px]`}
          disabled={phase !== PHASE_DEAL}
          onClick={() => exec('stand')}
        >
          スタンド
        </button>
      </div>
    </div>
  )
}

function cardPrefix(design: string): string {
  const map: Record<string, string> = {
    SPADE: 's', CLOVER: 'c', HEART: 'h', DIAMOND: 'd', JOKER: 'x',
  }
  return map[design] ?? 'x'
}
