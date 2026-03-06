import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { blackjackApi, daifugoApi, doubtApi, holdemApi, oldmaidApi, pokerApi, sessionId, sevensApi } from './gameApi';

describe('gameApi', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  function makeResponse(data: unknown, ok = true, status = 200) {
    return Promise.resolve({
      ok,
      status,
      json: () => Promise.resolve(data),
    });
  }

  describe('sessionId', () => {
    it('is a non-empty string', () => {
      expect(typeof sessionId).toBe('string');
      expect(sessionId.length).toBeGreaterThan(0);
    });
  });

  describe('blackjackApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        dealer: { score: 17, cards: [], chips: 1000 },
        player: { score: 15, cards: [], chips: 1000 },
        phase: 1,
        currentHandIdx: 0,
        insuranceBet: 0,
        insuranceAvailable: false,
        message: '',
      };
      mockFetch.mockReturnValue(makeResponse(payload));

      const result = await blackjackApi.exec('reset');

      expect(mockFetch).toHaveBeenCalledWith('/blackjack/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', amount: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with hit command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 20, cards: [], chips: 900 },
          phase: 4,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
        }),
      );
      await blackjackApi.exec('hit');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'hit', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with stand command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 18, cards: [], chips: 1000 },
          player: { score: 19, cards: [], chips: 1100 },
          phase: 5,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: 'win',
        }),
      );
      await blackjackApi.exec('stand');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'stand', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with bet command and amount', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 15, cards: [], chips: 900 },
          phase: 4,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
        }),
      );
      await blackjackApi.exec('bet', 100);
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'bet', amount: 100, sessionId }),
        }),
      );
    });

    it('calls with doubledown command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 18, cards: [], chips: 800 },
          phase: 5,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
        }),
      );
      await blackjackApi.exec('doubledown');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'doubledown', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with split command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 8, cards: [], chips: 800 },
          phase: 4,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
        }),
      );
      await blackjackApi.exec('split');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'split', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with insurance command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 15, cards: [], chips: 850 },
          phase: 4,
          currentHandIdx: 0,
          insuranceBet: 50,
          insuranceAvailable: true,
          message: '',
        }),
      );
      await blackjackApi.exec('insurance');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'insurance', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with declineinsurance command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 15, cards: [], chips: 900 },
          phase: 4,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
        }),
      );
      await blackjackApi.exec('declineinsurance');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'declineinsurance', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with togglesoft17 command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 0, cards: [], chips: 1000 },
          phase: 1,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          dealerHitsSoft17: true,
        }),
      );
      await blackjackApi.exec('togglesoft17');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'togglesoft17', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with togglecounting command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 0, cards: [], chips: 1000 },
          phase: 1,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          countingEnabled: true,
        }),
      );
      await blackjackApi.exec('togglecounting');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'togglecounting', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with reset command and config', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 0, cards: [], chips: 1000 },
          phase: 1,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          dealerHitsSoft17: true,
          cpuPlayerCount: 2,
          countingEnabled: true,
        }),
      );
      await blackjackApi.exec('reset', undefined, {
        dealerHitsSoft17: true,
        cpuPlayerCount: 2,
        countingEnabled: true,
      });
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            sessionId,
            dealerHitsSoft17: true,
            cpuPlayerCount: 2,
            countingEnabled: true,
          }),
        }),
      );
    });

    it('calls with bet command and side bets', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 15, cards: [], chips: 880 },
          phase: 4,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          perfectPairsBet: 10,
          twentyOnePlus3Bet: 20,
        }),
      );
      await blackjackApi.exec('bet', 100, undefined, { perfectPairsBet: 10, twentyOnePlus3Bet: 20 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'bet',
            amount: 100,
            sessionId,
            perfectPairsBet: 10,
            twentyOnePlus3Bet: 20,
          }),
        }),
      );
    });

    it('does not include side bet fields when sideBets is omitted', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 15, cards: [], chips: 900 },
          phase: 4,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
        }),
      );
      await blackjackApi.exec('bet', 100);
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'bet', amount: 100, sessionId }),
        }),
      );
    });

    it('calls with setcountingsystem command and amount', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 0, cards: [], chips: 1000 },
          phase: 1,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          countingSystem: 2,
        }),
      );
      await blackjackApi.exec('setcountingsystem', 2);
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'setcountingsystem', amount: 2, sessionId }),
        }),
      );
    });

    it('calls with setpenetration command and amount', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 0, cards: [], chips: 1000 },
          phase: 1,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          deckPenetration: 50,
        }),
      );
      await blackjackApi.exec('setpenetration', 50);
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'setpenetration', amount: 50, sessionId }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(blackjackApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('pokerApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        phase: 0,
        player: { cards: [], handRank: 0, handName: '', chips: 1000, bet: 0 },
        dealer: { cards: [], handRank: 0, handName: '', chips: 1000, bet: 0 },
        message: '',
        pot: 0,
        ante: 10,
      };
      mockFetch.mockReturnValue(makeResponse(payload));

      const result = await pokerApi.exec('reset');

      expect(mockFetch).toHaveBeenCalledWith('/poker/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', indices: undefined, amount: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with exchange command and indices', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          phase: 3,
          player: { cards: [], handRank: 1, handName: 'Pair', chips: 980, bet: 0 },
          dealer: { cards: [], handRank: 0, handName: 'High Card', chips: 980, bet: 0 },
          message: '',
          pot: 40,
          ante: 10,
        }),
      );
      await pokerApi.exec('exchange', [0, 2, 4]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'exchange', indices: [0, 2, 4], amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with bet command and amount', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          phase: 2,
          player: { cards: [], handRank: 0, handName: '', chips: 970, bet: 20 },
          dealer: { cards: [], handRank: 0, handName: '', chips: 970, bet: 20 },
          message: '',
          pot: 60,
          ante: 10,
        }),
      );
      await pokerApi.exec('bet', undefined, 20);
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'bet', indices: undefined, amount: 20, sessionId }),
        }),
      );
    });

    it('calls with call command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          phase: 2,
          player: { cards: [], handRank: 0, handName: '', chips: 980, bet: 10 },
          dealer: { cards: [], handRank: 0, handName: '', chips: 980, bet: 10 },
          message: '',
          pot: 40,
          ante: 10,
        }),
      );
      await pokerApi.exec('call');
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'call', indices: undefined, amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with fold command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          phase: 4,
          player: { cards: [], handRank: 0, handName: '', chips: 990, bet: 0 },
          dealer: { cards: [], handRank: 0, handName: '', chips: 1010, bet: 0 },
          message: 'You folded.',
          pot: 0,
          ante: 10,
        }),
      );
      await pokerApi.exec('fold');
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'fold', indices: undefined, amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with check command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          phase: 2,
          player: { cards: [], handRank: 0, handName: '', chips: 990, bet: 0 },
          dealer: { cards: [], handRank: 0, handName: '', chips: 990, bet: 0 },
          message: '',
          pot: 20,
          ante: 10,
        }),
      );
      await pokerApi.exec('check');
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'check', indices: undefined, amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with odds command and indices', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          phase: 2,
          odds: [{ handRank: 1, handName: 'One Pair', probability: 0.5, count: 5, total: 10 }],
          message: '',
        }),
      );
      await pokerApi.exec('odds', [0, 2]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'odds', indices: [0, 2], amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with reset command and bettingLimit', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          phase: 0,
          player: { cards: [], handRank: 0, handName: '', chips: 1000, bet: 0 },
          dealer: { cards: [], handRank: 0, handName: '', chips: 1000, bet: 0 },
          message: '',
          pot: 0,
          ante: 10,
        }),
      );
      await pokerApi.exec('reset', undefined, undefined, { bettingLimit: 1 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            indices: undefined,
            amount: undefined,
            bettingLimit: 1,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 503));
      await expect(pokerApi.exec('reset')).rejects.toThrow('HTTP error: 503');
    });
  });

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
      };
      mockFetch.mockReturnValue(makeResponse(payload));

      const result = await oldmaidApi.exec('reset');

      expect(mockFetch).toHaveBeenCalledWith('/oldmaid/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', drawIdx: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with draw command and drawIdx', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
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
        }),
      );
      await oldmaidApi.exec('draw', 2);
      expect(mockFetch).toHaveBeenCalledWith(
        '/oldmaid/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'draw', drawIdx: 2, sessionId }),
        }),
      );
    });

    it('calls with reset command and mode and cpuPlacementStrategy', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
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
        }),
      );
      await oldmaidApi.exec('reset', undefined, 1, true);
      expect(mockFetch).toHaveBeenCalledWith(
        '/oldmaid/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'reset', mode: 1, cpuPlacementStrategy: true, sessionId }),
        }),
      );
    });

    it('calls with shuffle command', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
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
        }),
      );
      await oldmaidApi.exec('shuffle');
      expect(mockFetch).toHaveBeenCalledWith(
        '/oldmaid/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'shuffle', sessionId }),
        }),
      );
    });

    it('calls with reorder command and reorderIndices', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
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
        }),
      );
      await oldmaidApi.exec('reorder', undefined, undefined, undefined, [2, 0, 1]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/oldmaid/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'reorder', reorderIndices: [2, 0, 1], sessionId }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 404));
      await expect(oldmaidApi.exec('reset')).rejects.toThrow('HTTP error: 404');
    });
  });

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
      };
      mockFetch.mockReturnValue(makeResponse(payload));

      const result = await daifugoApi.exec('reset');

      expect(mockFetch).toHaveBeenCalledWith('/daifugo/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          indices: undefined,
          config: undefined,
          sortMode: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with play command and indices', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 1,
          tableCards: [{ design: 'SPADE', value: 5 }],
          lastPlayPlayerIdx: 0,
          gameEndFlag: false,
          cpuActions: [],
          humanAction: { playerIdx: 0, playedCards: [{ design: 'SPADE', value: 5 }] },
          message: '',
        }),
      );
      await daifugoApi.exec('play', [0]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/daifugo/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'play', indices: [0], config: undefined, sortMode: undefined, sessionId }),
        }),
      );
    });

    it('calls with play command and empty indices for pass', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 1,
          tableCards: [],
          lastPlayPlayerIdx: -1,
          gameEndFlag: false,
          cpuActions: [],
          humanAction: { playerIdx: 0, playedCards: null },
          message: '',
        }),
      );
      await daifugoApi.exec('play', []);
      expect(mockFetch).toHaveBeenCalledWith(
        '/daifugo/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'play', indices: [], config: undefined, sortMode: undefined, sessionId }),
        }),
      );
    });

    it('calls with sort command and sortMode', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 0,
          tableCards: [],
          lastPlayPlayerIdx: -1,
          gameEndFlag: false,
          cpuActions: [],
          humanAction: null,
          message: '',
        }),
      );
      await daifugoApi.exec('sort', undefined, undefined, 1);
      expect(mockFetch).toHaveBeenCalledWith(
        '/daifugo/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'sort', indices: undefined, config: undefined, sortMode: 1, sessionId }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(daifugoApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('doubtApi.exec', () => {
    const payload = {
      players: [],
      currentTurn: 0,
      phase: 0,
      tableCardCount: 0,
      lastAction: null,
      cpuDoubters: [],
      cpuActions: [],
      humanAction: null,
      lastDoubtResult: null,
      gameEndFlag: false,
      winnerIdx: -1,
      message: '',
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await doubtApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/doubt/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndices: undefined,
          claimedValue: undefined,
          doubterIndices: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with play command and card indices and claimed value', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await doubtApi.exec('play', [0, 2], 5);
      expect(mockFetch).toHaveBeenCalledWith(
        '/doubt/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'play',
            cardIndices: [0, 2],
            claimedValue: 5,
            doubterIndices: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with doubt command and doubter indices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await doubtApi.exec('doubt', undefined, undefined, [0, 1]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/doubt/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'doubt',
            cardIndices: undefined,
            claimedValue: undefined,
            doubterIndices: [0, 1],
            sessionId,
          }),
        }),
      );
    });

    it('calls with skip command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await doubtApi.exec('skip');
      expect(mockFetch).toHaveBeenCalledWith(
        '/doubt/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'skip',
            cardIndices: undefined,
            claimedValue: undefined,
            doubterIndices: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('sends config fields when config is provided', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await doubtApi.exec('reset', undefined, undefined, undefined, { doubtWindowSec: 3, cpuMemoryLevel: 2 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/doubt/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            sessionId,
            doubtWindowSec: 3,
            cpuMemoryLevel: 2,
          }),
        }),
      );
    });

    it('omits config fields when config is not provided', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await doubtApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith(
        '/doubt/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(doubtApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('sevensApi.exec', () => {
    it('calls the correct URL with reset command', async () => {
      const payload = {
        players: [],
        currentTurn: 0,
        tableMinVals: [0, 7, 7, 7, 7],
        tableMaxVals: [0, 7, 7, 7, 7],
        tablePlaced: [0, 128, 128, 128, 128],
        config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: false, maxPasses: 5, noJokerFinish: false },
        gameEndFlag: false,
        cpuActions: [],
        humanAction: null,
        message: '',
      };
      mockFetch.mockReturnValue(makeResponse(payload));

      const result = await sevensApi.exec('reset');

      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', index: -1, jokerTargetSuit: 0, jokerTargetValue: 0, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with play command and card index', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 1,
          tableMinVals: [0, 6, 7, 7, 7],
          tableMaxVals: [0, 7, 7, 7, 7],
          tablePlaced: [0, 192, 128, 128, 128],
          config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: false, maxPasses: 5, noJokerFinish: false },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: { playerIdx: 0, playedCard: { design: 'SPADE', value: 6 }, targetSuit: 0, targetValue: 0 },
          message: '',
        }),
      );
      await sevensApi.exec('play', 2);
      expect(mockFetch).toHaveBeenCalledWith(
        '/sevens/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'play', index: 2, jokerTargetSuit: 0, jokerTargetValue: 0, sessionId }),
        }),
      );
    });

    it('calls with play command and -1 for pass', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 1,
          tableMinVals: [0, 7, 7, 7, 7],
          tableMaxVals: [0, 7, 7, 7, 7],
          tablePlaced: [0, 128, 128, 128, 128],
          config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: false, maxPasses: 5, noJokerFinish: false },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: { playerIdx: 0, playedCard: null, targetSuit: 0, targetValue: 0 },
          message: '',
        }),
      );
      await sevensApi.exec('play', -1);
      expect(mockFetch).toHaveBeenCalledWith(
        '/sevens/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'play', index: -1, jokerTargetSuit: 0, jokerTargetValue: 0, sessionId }),
        }),
      );
    });

    it('calls with joker command and target position', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 1,
          tableMinVals: [0, 6, 7, 7, 7],
          tableMaxVals: [0, 7, 7, 7, 7],
          tablePlaced: [0, 192, 128, 128, 128],
          config: { tunnelEnabled: false, jokerCount: 1, cpuStrategy: false, maxPasses: 5, noJokerFinish: false },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: {
            playerIdx: 0,
            playedCard: { design: 'JOKER', value: 0 },
            targetSuit: 1,
            targetValue: 6,
          },
          message: '',
        }),
      );
      await sevensApi.exec('joker', 0, 1, 6);
      expect(mockFetch).toHaveBeenCalledWith(
        '/sevens/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'joker', index: 0, jokerTargetSuit: 1, jokerTargetValue: 6, sessionId }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(sevensApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });

    it('sends config fields in body when config is provided', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 0,
          tableMinVals: [0, 7, 7, 7, 7],
          tableMaxVals: [0, 7, 7, 7, 7],
          tablePlaced: [0, 128, 128, 128, 128],
          config: { tunnelEnabled: true, jokerCount: 2, cpuStrategy: true, maxPasses: 5, noJokerFinish: true },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: null,
          message: '',
        }),
      );
      await sevensApi.exec('reset', -1, 0, 0, {
        tunnelEnabled: true,
        jokerCount: 2,
        cpuStrategy: true,
        noJokerFinish: true,
      });
      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          index: -1,
          jokerTargetSuit: 0,
          jokerTargetValue: 0,
          sessionId,
          tunnelEnabled: true,
          jokerCount: 2,
          cpuStrategy: true,
          noJokerFinish: true,
        }),
      });
    });

    it('sends maxPasses in config fields when config is provided', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 0,
          tableMinVals: [0, 7, 7, 7, 7],
          tableMaxVals: [0, 7, 7, 7, 7],
          tablePlaced: [0, 128, 128, 128, 128],
          config: { tunnelEnabled: true, jokerCount: 2, cpuStrategy: true, maxPasses: 3, noJokerFinish: false },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: null,
          message: '',
        }),
      );
      await sevensApi.exec('reset', -1, 0, 0, { tunnelEnabled: true, jokerCount: 2, cpuStrategy: true, maxPasses: 3 });
      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          index: -1,
          jokerTargetSuit: 0,
          jokerTargetValue: 0,
          sessionId,
          tunnelEnabled: true,
          jokerCount: 2,
          cpuStrategy: true,
          maxPasses: 3,
        }),
      });
    });

    it('does not send config fields when config is omitted', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 0,
          tableMinVals: [0, 7, 7, 7, 7],
          tableMaxVals: [0, 7, 7, 7, 7],
          tablePlaced: [0, 128, 128, 128, 128],
          config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: false, maxPasses: 5, noJokerFinish: false },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: null,
          message: '',
        }),
      );
      await sevensApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', index: -1, jokerTargetSuit: 0, jokerTargetValue: 0, sessionId }),
      });
    });
  });

  describe('holdemApi.exec', () => {
    const payload = {
      players: [],
      communityCards: [],
      pot: 0,
      sidePots: [],
      dealerIdx: 0,
      currentTurn: 0,
      phase: 1,
      gameEndFlag: false,
      lastBet: 0,
      minRaise: 0,
      roundResults: [],
      cpuActions: [],
      message: '',
      handCount: 0,
      smallBlind: 5,
      bigBlind: 10,
      tournamentMode: false,
      blindLevelHands: 10,
      blindMultiplier: 200,
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await holdemApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/holdem/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          amount: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with fold command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('fold');
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'fold',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with bet command and amount', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('bet', 50);
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'bet',
            amount: 50,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and custom blinds', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('reset', undefined, { smallBlind: 10, bigBlind: 20 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            sessionId,
            smallBlind: 10,
            bigBlind: 20,
          }),
        }),
      );
    });

    it('calls with reset and tournament config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('reset', undefined, {
        tournamentMode: true,
        blindLevelHands: 5,
        blindMultiplier: 200,
      });
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            sessionId,
            tournamentMode: true,
            blindLevelHands: 5,
            blindMultiplier: 200,
          }),
        }),
      );
    });

    it('calls with reset and bettingLimit config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('reset', undefined, { bettingLimit: 2 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            sessionId,
            bettingLimit: 2,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(holdemApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });
});
