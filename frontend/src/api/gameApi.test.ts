import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { blackjackApi, pokerApi, oldmaidApi } from './gameApi'

describe('gameApi', () => {
  const mockFetch = vi.fn()

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.clearAllMocks()
  })

  function makeResponse(data: unknown, ok = true, status = 200) {
    return Promise.resolve({
      ok,
      status,
      json: () => Promise.resolve(data),
    })
  }

  describe('blackjackApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        dealer: { score: 17, cards: [] },
        player: { score: 15, cards: [] },
        message: '',
      }
      mockFetch.mockReturnValue(makeResponse(payload))

      const result = await blackjackApi.exec('reset')

      expect(mockFetch).toHaveBeenCalledWith('/blackjac/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset' }),
      })
      expect(result).toEqual(payload)
    })

    it('calls with hit command', async () => {
      mockFetch.mockReturnValue(makeResponse({ dealer: { score: 0, cards: [] }, player: { score: 20, cards: [] }, message: '' }))
      await blackjackApi.exec('hit')
      expect(mockFetch).toHaveBeenCalledWith('/blackjac/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'hit' }),
      }))
    })

    it('calls with stand command', async () => {
      mockFetch.mockReturnValue(makeResponse({ dealer: { score: 18, cards: [] }, player: { score: 19, cards: [] }, message: 'win' }))
      await blackjackApi.exec('stand')
      expect(mockFetch).toHaveBeenCalledWith('/blackjac/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'stand' }),
      }))
    })

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500))
      await expect(blackjackApi.exec('reset')).rejects.toThrow('HTTP error: 500')
    })
  })

  describe('pokerApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        phase: 0,
        player: { cards: [], handName: '' },
        dealer: { cards: [], handName: '' },
        message: '',
      }
      mockFetch.mockReturnValue(makeResponse(payload))

      const result = await pokerApi.exec('reset')

      expect(mockFetch).toHaveBeenCalledWith('/poker/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', indices: undefined }),
      })
      expect(result).toEqual(payload)
    })

    it('calls with exchange command and indices', async () => {
      mockFetch.mockReturnValue(makeResponse({
        phase: 2,
        player: { cards: [], handName: 'Pair' },
        dealer: { cards: [], handName: 'High Card' },
        message: 'win',
      }))
      await pokerApi.exec('exchange', [0, 2, 4])
      expect(mockFetch).toHaveBeenCalledWith('/poker/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'exchange', indices: [0, 2, 4] }),
      }))
    })

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 503))
      await expect(pokerApi.exec('reset')).rejects.toThrow('HTTP error: 503')
    })
  })

  describe('oldmaidApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        players: [],
        currentTurn: 0,
        nextDrawTargetIdx: 1,
        gameEndFlag: false,
        hasDrawn: false,
        lastDrawPlayerIdx: 0,
        lastDrawFromIdx: 0,
        lastDrawCard: null,
        lastDiscardedPairs: 0,
        cpuActions: [],
        message: '',
      }
      mockFetch.mockReturnValue(makeResponse(payload))

      const result = await oldmaidApi.exec('reset')

      expect(mockFetch).toHaveBeenCalledWith('/oldmaid/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', drawIdx: undefined }),
      })
      expect(result).toEqual(payload)
    })

    it('calls with draw command and drawIdx', async () => {
      mockFetch.mockReturnValue(makeResponse({
        players: [],
        currentTurn: 1,
        nextDrawTargetIdx: 0,
        gameEndFlag: false,
        hasDrawn: true,
        lastDrawPlayerIdx: 1,
        lastDrawFromIdx: 0,
        lastDrawCard: { design: 'SPADE', value: 3 },
        lastDiscardedPairs: 0,
        cpuActions: [],
        message: '',
      }))
      await oldmaidApi.exec('draw', 2)
      expect(mockFetch).toHaveBeenCalledWith('/oldmaid/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'draw', drawIdx: 2 }),
      }))
    })

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 404))
      await expect(oldmaidApi.exec('reset')).rejects.toThrow('HTTP error: 404')
    })
  })
})
