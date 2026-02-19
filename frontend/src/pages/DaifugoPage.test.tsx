import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { DaifugoPage } from './DaifugoPage'
import { daifugoApi } from '../api/gameApi'
import type { DaifugoResponse } from '../types/card'

vi.mock('../api/gameApi', () => ({
  daifugoApi: { exec: vi.fn() },
}))

const mockExec = vi.mocked(daifugoApi.exec)

const humanTurnState: DaifugoResponse = {
  players: [
    { id: 0, isHuman: true, isFinished: false, cardCount: 3, cards: [
      { design: 'SPADE', value: 3 },
      { design: 'HEART', value: 3 },
      { design: 'DIAMOND', value: 4 },
    ], rank: -1 },
    { id: 1, isHuman: false, isFinished: false, cardCount: 5, cards: [], rank: -1 },
    { id: 2, isHuman: false, isFinished: false, cardCount: 5, cards: [], rank: -1 },
    { id: 3, isHuman: false, isFinished: false, cardCount: 5, cards: [], rank: -1 },
  ],
  currentTurn: 0,
  lastPlay: [],
  isRevolution: false,
  gameEndFlag: false,
  message: '',
}

const fieldWithCardsState: DaifugoResponse = {
  ...humanTurnState,
  lastPlay: [{ design: 'CLOVER', value: 3 }],
}

beforeEach(() => {
  mockExec.mockResolvedValue(humanTurnState)
})

describe('DaifugoPage', () => {
  it('calls reset command on mount', async () => {
    render(<DaifugoPage />)
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined))
  })

  it('renders player labels', async () => {
    render(<DaifugoPage />)
    await waitFor(() => {
      expect(screen.getByText('あなた')).toBeInTheDocument()
      expect(screen.getByText('CPU 1')).toBeInTheDocument()
    })
  })

  it('shows human player cards', async () => {
    render(<DaifugoPage />)
    await waitFor(() => {
      expect(screen.getByAltText('SPADE 3')).toBeInTheDocument()
      expect(screen.getByAltText('HEART 3')).toBeInTheDocument()
      expect(screen.getByAltText('DIAMOND 4')).toBeInTheDocument()
    })
  })

  it('toggles card selection when clicked', async () => {
    render(<DaifugoPage />)
    await waitFor(() => expect(screen.getByAltText('SPADE 3')).toBeInTheDocument())
    
    const card = screen.getByAltText('SPADE 3')
    fireEvent.click(card)
    // Use data-selected as a more reliable indicator for testing than exact style string matching
    expect(card).toHaveAttribute('data-selected', 'true')
    
    fireEvent.click(card)
    expect(card).toHaveAttribute('data-selected', 'false')
  })

  it('calls play with selected indices when "出す" is clicked', async () => {
    render(<DaifugoPage />)
    await waitFor(() => expect(screen.getByAltText('SPADE 3')).toBeInTheDocument())
    
    fireEvent.click(screen.getByAltText('SPADE 3'))
    fireEvent.click(screen.getByAltText('HEART 3'))
    
    const playBtn = screen.getByText('出す')
    expect(playBtn).not.toBeDisabled()
    
    mockExec.mockClear()
    mockExec.mockResolvedValue(humanTurnState)
    fireEvent.click(playBtn)
    
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', [0, 1]))
  })

  it('calls pass when "パス" is clicked', async () => {
    render(<DaifugoPage />)
    await waitFor(() => expect(screen.getByText('パス')).toBeInTheDocument())
    
    mockExec.mockClear()
    mockExec.mockResolvedValue(humanTurnState)
    fireEvent.click(screen.getByText('パス'))
    
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', undefined))
  })

  it('shows cards in the field', async () => {
    mockExec.mockResolvedValue(fieldWithCardsState)
    render(<DaifugoPage />)
    await waitFor(() => {
      expect(screen.getByAltText('CLOVER 3')).toBeInTheDocument()
    })
  })

  it('shows "革命中" badge when isRevolution is true', async () => {
    const revolutionState: DaifugoResponse = {
      ...humanTurnState,
      isRevolution: true,
    }
    mockExec.mockResolvedValue(revolutionState)
    render(<DaifugoPage />)
    await waitFor(() => expect(screen.getByText('革命中')).toBeInTheDocument())
  })

  it('shows player rank when finished', async () => {
    const finishedState: DaifugoResponse = {
      ...humanTurnState,
      players: [
        { ...humanTurnState.players[0], isFinished: true, rank: 0 },
        ...humanTurnState.players.slice(1),
      ],
    }
    mockExec.mockResolvedValue(finishedState)
    render(<DaifugoPage />)
    await waitFor(() => expect(screen.getByText('大富豪')).toBeInTheDocument())
  })
})
