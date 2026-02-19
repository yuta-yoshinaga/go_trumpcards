import { useCallback, useEffect, useState } from 'react'
import { daifugoApi } from '../api/gameApi'
import { CardImage, CardBack } from '../components/CardImage'
import type { DaifugoResponse, DaifugoPlayerData, DaifugoAction, Card } from '../types/card'

const tableStyle: React.CSSProperties = {
  backgroundColor: '#1a5c1a',
  borderRadius: 16,
  padding: 20,
  margin: '10px auto',
  maxWidth: 960,
}

const playerAreaBase: React.CSSProperties = {
  background: 'rgba(0,0,0,0.35)',
  borderRadius: 10,
  padding: 10,
  border: '2px solid transparent',
  flex: '1 1 180px',
  minWidth: 150,
}

function playerName(idx: number): string {
  return idx === 0 ? 'あなた' : `CPU ${idx}`
}

function rankName(rank: number): string {
  switch (rank) {
    case 1: return '大富豪'
    case 2: return '富豪'
    case 3: return '平民'
    case 4: return '大貧民'
    default: return ''
  }
}

function cardLabel(card: Card | null): string {
  if (!card) return ''
  return `${card.design} ${card.value}`
}

function actionDescription(action: DaifugoAction): string {
  if (!action.playedCards || action.playedCards.length === 0) {
    return `${playerName(action.playerIdx)}がパスしました`
  }
  const cards = action.playedCards.map(cardLabel).join(', ')
  return `${playerName(action.playerIdx)}が出しました: ${cards}`
}

interface CpuPlayerAreaProps {
  player: DaifugoPlayerData
  isCurrentTurn: boolean
}

function CpuPlayerArea({ player, isCurrentTurn }: CpuPlayerAreaProps) {
  const areaStyle: React.CSSProperties = {
    ...playerAreaBase,
    ...(player.isFinished
      ? { opacity: 0.5 }
      : isCurrentTurn
      ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
      : {}),
  }
  const showCount = Math.min(player.cardCount, 10)
  return (
    <div id={`player-area-${player.id}`} style={areaStyle}>
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4 }}>
        {playerName(player.id)}
        {player.isFinished && (
          <span style={{
            background: '#5cb85c', color: '#fff', borderRadius: 6,
            padding: '1px 8px', marginLeft: 6, fontSize: '0.8em',
          }}>上がり ({rankName(player.rank)})</span>
        )}
        {isCurrentTurn && !player.isFinished && (
          <span style={{
            background: '#f0ad4e', color: '#222', borderRadius: 6,
            padding: '1px 8px', marginLeft: 6, fontSize: '0.8em', fontWeight: 'bold',
          }}>考え中...</span>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {player.cardCount}枚
        </div>
      )}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
        {player.isFinished ? null : (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              <CardBack key={i} style={{ width: 60 }} />
            ))}
            {player.cardCount > 10 && (
              <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 4 }}>
                +{player.cardCount - 10}
              </span>
            )}
          </>
        )}
      </div>
    </div>
  )
}

interface HumanPlayerAreaProps {
  player: DaifugoPlayerData
  selectedIndices: number[]
  onToggle: (idx: number) => void
  isCurrentTurn: boolean
}

function HumanPlayerArea({ player, selectedIndices, onToggle, isCurrentTurn }: HumanPlayerAreaProps) {
  const areaStyle: React.CSSProperties = {
    ...playerAreaBase,
    ...(player.isFinished
      ? { opacity: 0.5 }
      : isCurrentTurn
      ? { border: '2px solid #5cb85c', boxShadow: '0 0 12px #5cb85c' }
      : {}),
  }
  return (
    <div id="player-area-0" style={areaStyle}>
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4 }}>
        {playerName(0)}
        {player.isFinished && (
          <span style={{
            background: '#5cb85c', color: '#fff', borderRadius: 6,
            padding: '1px 8px', marginLeft: 6, fontSize: '0.8em',
          }}>上がり ({rankName(player.rank)})</span>
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
          <div
            key={i}
            onClick={() => isCurrentTurn && onToggle(i)}
            style={{
              cursor: isCurrentTurn ? 'pointer' : 'default',
              borderRadius: 8,
              border: selectedIndices.includes(i) ? '3px solid #f0ad4e' : '3px solid transparent',
              boxSizing: 'border-box',
            }}
          >
            <CardImage card={card} style={{ width: 60 }} />
          </div>
        ))}
      </div>
    </div>
  )
}

export function DaifugoPage() {
  const [state, setState] = useState<DaifugoResponse | null>(null)
  const [selectedIndices, setSelectedIndices] = useState<number[]>([])

  const exec = useCallback(async (command: 'reset' | 'play', indices?: number[]) => {
    try {
      const res = await daifugoApi.exec(command, indices)
      setState(res)
      setSelectedIndices([])
    } catch {
      console.error('daifugo request failed')
    }
  }, [])

  useEffect(() => { exec('reset') }, [exec])

  if (!state) return null

  const isHumanTurn = !state.gameEndFlag && state.currentTurn === 0
  const cpuPlayers = state.players.filter(p => !p.isHuman)
  const humanPlayer = state.players.find(p => p.isHuman)

  const toggleCardSelection = (idx: number) => {
    setSelectedIndices(prev =>
      prev.includes(idx) ? prev.filter(i => i !== idx) : [...prev, idx]
    )
  }

  return (
    <div style={tableStyle}>
      {/* CPU row */}
      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 12 }}>
        {cpuPlayers.map(player => (
          <CpuPlayerArea
            key={player.id}
            player={player}
            isCurrentTurn={state.currentTurn === player.id}
          />
        ))}
      </div>

      {/* Table cards */}
      <div style={{
        background: 'rgba(0,0,0,0.3)',
        borderRadius: 10,
        padding: 10,
        margin: '8px 0',
      }}>
        <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 6 }}>場札</div>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {!state.tableCards || state.tableCards.length === 0 ? (
            <span style={{ color: '#aaa' }}>（なし）</span>
          ) : state.tableCards.map((card, i) => (
            <CardImage key={i} card={card} style={{ width: 60 }} />
          ))}
        </div>
      </div>

      <div style={{ borderTop: '1px solid rgba(255,255,255,0.2)', margin: '12px 0' }} />

      {/* Human player */}
      {humanPlayer && (
        <HumanPlayerArea
          player={humanPlayer}
          selectedIndices={selectedIndices}
          onToggle={toggleCardSelection}
          isCurrentTurn={isHumanTurn}
        />
      )}

      {/* Human action log */}
      {state.humanAction && (
        <div style={{
          background: 'rgba(0,0,0,0.4)',
          borderRadius: 8,
          color: '#cfc',
          padding: '8px 14px',
          margin: '8px 0',
          fontSize: '0.85em',
        }}>
          {actionDescription(state.humanAction)}
        </div>
      )}

      {/* CPU action log */}
      {state.cpuActions && state.cpuActions.length > 0 && (
        <div style={{
          background: 'rgba(0,0,0,0.4)',
          borderRadius: 8,
          color: '#ccc',
          padding: '8px 14px',
          margin: '8px 0',
          whiteSpace: 'pre-line',
          fontSize: '0.85em',
        }}>
          {['[CPUの行動]', ...state.cpuActions.map(actionDescription)].join('\n')}
        </div>
      )}

      {/* Result message */}
      {state.message && (
        <div style={{
          background: 'rgba(0,0,0,0.55)',
          borderRadius: 10,
          color: '#fff',
          textAlign: 'center',
          padding: '12px 20px',
          fontSize: '1.3em',
          fontWeight: 'bold',
          margin: '10px 0',
        }}>
          {state.message}
        </div>
      )}

      {/* Buttons */}
      <div style={{ textAlign: 'center', margin: '14px 0 4px 0' }}>
        <button
          className="btn btn-primary btn-sm"
          style={{ margin: '0 6px', minWidth: 90 }}
          onClick={() => exec('reset')}
        >
          リセット
        </button>
        <button
          className="btn btn-warning btn-sm"
          style={{ margin: '0 6px', minWidth: 90 }}
          disabled={!isHumanTurn || state.gameEndFlag}
          onClick={() => exec('play', [])}
        >
          パス
        </button>
        <button
          className="btn btn-success btn-sm"
          style={{ margin: '0 6px', minWidth: 120 }}
          disabled={!isHumanTurn || state.gameEndFlag || selectedIndices.length === 0}
          onClick={() => exec('play', [...selectedIndices].sort((a, b) => a - b))}
        >
          選択して出す
        </button>
      </div>
    </div>
  )
}
