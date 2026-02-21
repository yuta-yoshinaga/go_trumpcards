import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { blackjackApi, pokerApi, oldmaidApi, daifugoApi, sevensApi, sessionId } from './gameApi'

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

  describe('sessionId', () => {
    it('is a non-empty string', () => {
      expect(typeof sessionId).toBe('string')
      expect(sessionId.length).toBeGreaterThan(0)
    })
  })

  describe('blackjackApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        dealer: { score: 17, cards: [] },
        player: { score: 15, cards: [] },
        message: '',
      }
      mockFetch.mockReturnValue(makeResponse(payload))

      const result = await blackjackApi.exec('reset')

      expect(mockFetch).toHaveBeenCalledWith('/blackjack/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', sessionId }),
      })
      expect(result).toEqual(payload)
    })

    it('calls with hit command', async () => {
      mockFetch.mockReturnValue(makeResponse({ dealer: { score: 0, cards: [] }, player: { score: 20, cards: [] }, message: '' }))
      await blackjackApi.exec('hit')
      expect(mockFetch).toHaveBeenCalledWith('/blackjack/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'hit', sessionId }),
      }))
    })

    it('calls with stand command', async () => {
      mockFetch.mockReturnValue(makeResponse({ dealer: { score: 18, cards: [] }, player: { score: 19, cards: [] }, message: 'win' }))
      await blackjackApi.exec('stand')
      expect(mockFetch).toHaveBeenCalledWith('/blackjack/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'stand', sessionId }),
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
        body: JSON.stringify({ command: 'reset', indices: undefined, sessionId }),
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
        body: JSON.stringify({ command: 'exchange', indices: [0, 2, 4], sessionId }),
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
        body: JSON.stringify({ command: 'reset', drawIdx: undefined, sessionId }),
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
        body: JSON.stringify({ command: 'draw', drawIdx: 2, sessionId }),
      }))
    })

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 404))
      await expect(oldmaidApi.exec('reset')).rejects.toThrow('HTTP error: 404')
    })
  })

  describe('daifugoApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        players: [],
        currentTurn: 0,
        tableCards: [],
        lastPlayPlayerIdx: -1,
        gameEndFlag: false,
        cpuActions: [],
        humanAction: null,
        message: '',
      }
      mockFetch.mockReturnValue(makeResponse(payload))

      const result = await daifugoApi.exec('reset')

      expect(mockFetch).toHaveBeenCalledWith('/daifugo/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', indices: undefined, sessionId }),
      })
      expect(result).toEqual(payload)
    })

    it('calls with play command and indices', async () => {
      mockFetch.mockReturnValue(makeResponse({
        players: [],
        currentTurn: 1,
        tableCards: [{ design: 'SPADE', value: 5 }],
        lastPlayPlayerIdx: 0,
        gameEndFlag: false,
        cpuActions: [],
        humanAction: { playerIdx: 0, playedCards: [{ design: 'SPADE', value: 5 }] },
        message: '',
      }))
      await daifugoApi.exec('play', [0])
      expect(mockFetch).toHaveBeenCalledWith('/daifugo/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'play', indices: [0], sessionId }),
      }))
    })

    it('calls with play command and empty indices for pass', async () => {
      mockFetch.mockReturnValue(makeResponse({
        players: [],
        currentTurn: 1,
        tableCards: [],
        lastPlayPlayerIdx: -1,
        gameEndFlag: false,
        cpuActions: [],
        humanAction: { playerIdx: 0, playedCards: null },
        message: '',
      }))
      await daifugoApi.exec('play', [])
      expect(mockFetch).toHaveBeenCalledWith('/daifugo/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'play', indices: [], sessionId }),
      }))
    })

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500))
      await expect(daifugoApi.exec('reset')).rejects.toThrow('HTTP error: 500')
    })
  })

  describe('sevensApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        players: [],
        currentTurn: 0,
        tableMinVals: [0, 7, 7, 7, 7],
        tableMaxVals: [0, 7, 7, 7, 7],
        gameEndFlag: false,
        cpuActions: [],
        humanAction: null,
        message: '',
      }
      mockFetch.mockReturnValue(makeResponse(payload))

      const result = await sevensApi.exec('reset')

      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', index: -1, sessionId }),
      })
      expect(result).toEqual(payload)
    })

    it('calls with play command and card index', async () => {
      mockFetch.mockReturnValue(makeResponse({
        players: [],
        currentTurn: 1,
        tableMinVals: [0, 6, 7, 7, 7],
        tableMaxVals: [0, 7, 7, 7, 7],
        gameEndFlag: false,
        cpuActions: [],
        humanAction: { playerIdx: 0, playedCard: { design: 'SPADE', value: 6 } },
        message: '',
      }))
      await sevensApi.exec('play', 2)
      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'play', index: 2, sessionId }),
      }))
    })

    it('calls with play command and -1 for pass', async () => {
      mockFetch.mockReturnValue(makeResponse({
        players: [],
        currentTurn: 1,
        tableMinVals: [0, 7, 7, 7, 7],
        tableMaxVals: [0, 7, 7, 7, 7],
        gameEndFlag: false,
        cpuActions: [],
        humanAction: { playerIdx: 0, playedCard: null },
        message: '',
      }))
      await sevensApi.exec('play', -1)
      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', expect.objectContaining({
        body: JSON.stringify({ command: 'play', index: -1, sessionId }),
      }))
    })

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500))
      await expect(sevensApi.exec('reset')).rejects.toThrow('HTTP error: 500')
    })
  })
})
