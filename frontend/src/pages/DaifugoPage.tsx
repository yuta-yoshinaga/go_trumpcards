import { useCallback, useEffect, useState } from 'react'
import { daifugoApi } from '../api/gameApi'
import { CardImage, CardBack } from '../components/CardImage'
import type { DaifugoResponse, DaifugoPlayerData } from '../types/card'

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
    case 0: return '大富豪'
    case 1: return '富豪'
    case 2: return '平民'
    case 3: return '貧民'
    case 4: return '大貧民'
    default: return ''
  }
}

interface PlayerAreaProps {
  player: DaifugoPlayerData
  isTurn: boolean
  gameEndFlag: boolean
  selectedIndices: number[]
  onToggleCard: (idx: number) => void
}

function PlayerArea({ player, isTurn, gameEndFlag, selectedIndices, onToggleCard }: PlayerAreaProps) {
  const areaStyle: React.CSSProperties = {
    ...playerAreaBase,
    ...(player.isFinished
      ? { opacity: 0.5 }
      : isTurn && !gameEndFlag
      ? { border: '2px solid #f0ad4e', boxShadow: '0 0 12px #f0ad4e' }
      : {}),
  }

  const showCount = Math.min(player.cardCount, 15)

  return (
    <div id={`player-area-${player.id}`} style={areaStyle}>
      <div style={{ color: '#fff', fontWeight: 'bold', marginBottom: 4 }}>
        {playerName(player.id)}
        {player.isFinished && player.rank >= 0 && (
          <span style={{
            background: '#5cb85c', color: '#fff', borderRadius: 6,
            padding: '1px 8px', marginLeft: 6, fontSize: '0.8em',
          }}>{rankName(player.rank)}</span>
        )}
        {isTurn && !gameEndFlag && (
          <span style={{
            background: '#f0ad4e', color: '#222', borderRadius: 6,
            padding: '1px 8px', marginLeft: 6, fontSize: '0.8em', fontWeight: 'bold',
          }}>TURN</span>
        )}
      </div>
      {!player.isFinished && (
        <div style={{ color: '#ccc', fontSize: '0.85em', marginBottom: 4 }}>
          {player.cardCount}枚
        </div>
      )}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
        {player.isFinished ? null : player.isHuman ? (
          player.cards?.map((card, i) => {
            const isSelected = selectedIndices.includes(i)
            return (
              <div key={i} onClick={() => onToggleCard(i)} style={{ cursor: 'pointer' }}>
                <CardImage
                  card={card}
                  style={{ 
                    width: 60, 
                    transform: isSelected ? 'translateY(-10px)' : 'none',
                    border: isSelected ? '2px solid #5cb85c' : 'none',
                    borderRadius: 6,
                    transition: 'transform 0.2s'
                  }}
                />
              </div>
            )
          })
        ) : (
          <>
            {Array.from({ length: showCount }).map((_, i) => (
              <CardBack key={i} style={{ width: 40 }} />
            ))}
            {player.cardCount > 15 && (
              <span style={{ color: '#fff', alignSelf: 'center', marginLeft: 4 }}>
                +{player.cardCount - 15}
              </span>
            )}
          </>
        )}
      </div>
    </div>
  )
}

export function DaifugoPage() {
  const [state, setState] = useState<DaifugoResponse | null>(null)
  const [selectedIndices, setSelectedIndices] = useState<number[]>([])

  const exec = useCallback(async (command: 'reset' | 'play' | 'pass', cardIndices?: number[]) => {
    try {
      const res = await daifugoApi.exec(command, cardIndices)
      setState(res)
      setSelectedIndices([])
    } catch {
      console.error('daifugo request failed')
    }
  }, [])

  useEffect(() => { exec('reset') }, [exec])

  const toggleCard = (idx: number) => {
    setSelectedIndices(prev => 
      prev.includes(idx) ? prev.filter(i => i !== idx) : [...prev, idx]
    )
  }

  if (!state) return null

  const isHumanTurn = !state.gameEndFlag && state.currentTurn === 0
  const cpuPlayers = state.players.filter(p => !p.isHuman)
  const humanPlayer = state.players.find(p => p.isHuman)

  return (
    <div style={tableStyle}>
      {/* CPU row */}
      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 12 }}>
        {cpuPlayers.map(player => (
          <PlayerArea
            key={player.id}
            player={player}
            isTurn={state.currentTurn === player.id}
            gameEndFlag={state.gameEndFlag}
            selectedIndices={[]}
            onToggleCard={() => {}}
          />
        ))}
      </div>

      <div style={{ borderTop: '1px solid rgba(255,255,255,0.2)', margin: '12px 0' }} />

      {/* Field */}
      <div style={{ 
        height: 120, 
        background: 'rgba(0,0,0,0.2)', 
        borderRadius: 8, 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'center',
        gap: 8,
        position: 'relative',
        marginBottom: 12
      }}>
        {state.lastPlay && state.lastPlay.length > 0 ? (
          state.lastPlay.map((card, i) => (
            <CardImage key={i} card={card} style={{ width: 80 }} />
          ))
        ) : (
          <span style={{ color: 'rgba(255,255,255,0.3)', fontSize: '1.2em' }}>場にカードはありません</span>
        )}
        {state.isRevolution && (
          <div style={{
            position: 'absolute',
            top: 10,
            right: 10,
            background: '#d9534f',
            color: '#fff',
            padding: '2px 10px',
            borderRadius: 4,
            fontWeight: 'bold',
            fontSize: '0.8em'
          }}>革命中</div>
        )}
      </div>

      {/* Human player */}
      {humanPlayer && (
        <PlayerArea
          player={humanPlayer}
          isTurn={state.currentTurn === 0}
          gameEndFlag={state.gameEndFlag}
          selectedIndices={selectedIndices}
          onToggleCard={toggleCard}
        />
      )}

      {/* Result */}
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
        <button className="btn btn-primary btn-sm" style={{ margin: '0 6px', minWidth: 90 }}
          onClick={() => exec('reset')}>
          リセット
        </button>
        <button className="btn btn-success btn-sm" style={{ margin: '0 6px', minWidth: 90 }}
          disabled={!isHumanTurn || state.gameEndFlag || selectedIndices.length === 0}
          onClick={() => exec('play', selectedIndices)}>
          出す
        </button>
        <button className="btn btn-warning btn-sm" style={{ margin: '0 6px', minWidth: 90 }}
          disabled={!isHumanTurn || state.gameEndFlag}
          onClick={() => exec('pass')}>
          パス
        </button>
      </div>
    </div>
  )
}
