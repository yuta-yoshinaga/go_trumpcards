import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  actionLogApi,
  agnesApi,
  baccaratApi,
  bidWhistApi,
  bigOApi,
  bigOHiLoApi,
  blackjackApi,
  blackjackswitchApi,
  canfieldApi,
  casinoholdemApi,
  chinchonApi,
  conquianApi,
  crazyeightsApi,
  daifugoApi,
  doubtApi,
  dramahaApi,
  ginrummyApi,
  golfApi,
  gongzhuApi,
  heartsApi,
  holdemApi,
  indianpokerApi,
  indianRummyApi,
  klondikeApi,
  letitrideApi,
  machiavelliApi,
  memoryApi,
  montecarloApi,
  oldmaidApi,
  omahaApi,
  panApi,
  pokerApi,
  pyramidApi,
  sessionId,
  sevensApi,
  shitheadApi,
  slapjackApi,
  spadesApi,
  spiderApi,
  threethirteenApi,
  tripeaksApi,
  twoTenJackApi,
} from './gameApi';

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
            dealerHitsSoft17: true,
            cpuPlayerCount: 2,
            countingEnabled: true,
            sessionId,
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
            perfectPairsBet: 10,
            twentyOnePlus3Bet: 20,
            sessionId,
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

    it('calls with earlysurrender command', async () => {
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
      await blackjackApi.exec('earlysurrender');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'earlysurrender', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with declineearlysurrender command', async () => {
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
      await blackjackApi.exec('declineearlysurrender');
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'declineearlysurrender', amount: undefined, sessionId }),
        }),
      );
    });

    it('calls with setsurrenderrule command and amount', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 0, cards: [], chips: 1000 },
          phase: 1,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          surrenderRule: 1,
        }),
      );
      await blackjackApi.exec('setsurrenderrule', 1);
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'setsurrenderrule', amount: 1, sessionId }),
        }),
      );
    });

    it('calls with reset command and surrenderRule in config', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          dealer: { score: 0, cards: [], chips: 1000 },
          player: { score: 0, cards: [], chips: 1000 },
          phase: 1,
          currentHandIdx: 0,
          insuranceBet: 0,
          insuranceAvailable: false,
          message: '',
          surrenderRule: 1,
        }),
      );
      await blackjackApi.exec('reset', undefined, { surrenderRule: 1 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjack/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            surrenderRule: 1,
            sessionId,
          }),
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

    it('calls with reset command and isLowball', async () => {
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
      await pokerApi.exec('reset', undefined, undefined, { isLowball: true });
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            indices: undefined,
            amount: undefined,
            isLowball: true,
            sessionId,
          }),
        }),
      );
    });

    it('calls with bet command and humanPlayMs', async () => {
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
      await pokerApi.exec('bet', undefined, 20, undefined, 500);
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'bet', indices: undefined, amount: 20, humanPlayMs: 500, sessionId }),
        }),
      );
    });

    it('calls with reset command and cpuMetaAI config', async () => {
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
      await pokerApi.exec('reset', undefined, undefined, { cpuMetaAI: true });
      expect(mockFetch).toHaveBeenCalledWith(
        '/poker/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            indices: undefined,
            amount: undefined,
            cpuMetaAI: true,
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

    it('calls with reset command and cpuMemoryAI', async () => {
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
      await oldmaidApi.exec('reset', undefined, 1, true, undefined, true);
      expect(mockFetch).toHaveBeenCalledWith(
        '/oldmaid/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            mode: 1,
            cpuPlacementStrategy: true,
            cpuMemoryAI: true,
            sessionId,
          }),
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
      await doubtApi.exec('reset', undefined, undefined, undefined, {
        doubtWindowSec: 3,
        cpuMemoryLevel: 2,
        penaltyDrawLimit: 5,
        cpuHesitationEnabled: true,
        cpuMetaAI: false,
      });
      expect(mockFetch).toHaveBeenCalledWith(
        '/doubt/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            doubtWindowSec: 3,
            cpuMemoryLevel: 2,
            penaltyDrawLimit: 5,
            cpuHesitationEnabled: true,
            cpuMetaAI: false,
            sessionId,
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

    it('sends humanPlayMs when provided', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await doubtApi.exec('play', [0], 5, undefined, undefined, 1500);
      expect(mockFetch).toHaveBeenCalledWith(
        '/doubt/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'play',
            cardIndices: [0],
            claimedValue: 5,
            humanPlayMs: 1500,
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
        config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: 0, maxPasses: 5, noJokerFinish: false },
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
          config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: 0, maxPasses: 5, noJokerFinish: false },
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
          config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: 0, maxPasses: 5, noJokerFinish: false },
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
          config: { tunnelEnabled: false, jokerCount: 1, cpuStrategy: 0, maxPasses: 5, noJokerFinish: false },
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
          config: { tunnelEnabled: true, jokerCount: 2, cpuStrategy: 1, maxPasses: 5, noJokerFinish: true },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: null,
          message: '',
        }),
      );
      await sevensApi.exec('reset', -1, 0, 0, {
        tunnelEnabled: true,
        jokerCount: 2,
        cpuStrategy: 1,
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
          tunnelEnabled: true,
          jokerCount: 2,
          cpuStrategy: 1,
          noJokerFinish: true,
          sessionId,
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
          config: { tunnelEnabled: true, jokerCount: 2, cpuStrategy: 1, maxPasses: 3, noJokerFinish: false },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: null,
          message: '',
        }),
      );
      await sevensApi.exec('reset', -1, 0, 0, { tunnelEnabled: true, jokerCount: 2, cpuStrategy: 1, maxPasses: 3 });
      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          index: -1,
          jokerTargetSuit: 0,
          jokerTargetValue: 0,
          tunnelEnabled: true,
          jokerCount: 2,
          cpuStrategy: 1,
          maxPasses: 3,
          sessionId,
        }),
      });
    });

    it('sends jokerConsecutiveBanned in config fields when config is provided', async () => {
      mockFetch.mockReturnValue(
        makeResponse({
          players: [],
          currentTurn: 0,
          tableMinVals: [0, 7, 7, 7, 7],
          tableMaxVals: [0, 7, 7, 7, 7],
          tablePlaced: [0, 128, 128, 128, 128],
          config: {
            tunnelEnabled: false,
            jokerCount: 1,
            cpuStrategy: 0,
            maxPasses: 5,
            noJokerFinish: false,
            jokerConsecutiveBanned: true,
          },
          gameEndFlag: false,
          cpuActions: [],
          humanAction: null,
          message: '',
        }),
      );
      await sevensApi.exec('reset', -1, 0, 0, { jokerCount: 1, jokerConsecutiveBanned: true });
      expect(mockFetch).toHaveBeenCalledWith('/sevens/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          index: -1,
          jokerTargetSuit: 0,
          jokerTargetValue: 0,
          jokerCount: 1,
          jokerConsecutiveBanned: true,
          sessionId,
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
          config: { tunnelEnabled: false, jokerCount: 0, cpuStrategy: 0, maxPasses: 5, noJokerFinish: false },
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
      tableSize: 4,
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
            smallBlind: 10,
            bigBlind: 20,
            sessionId,
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
            tournamentMode: true,
            blindLevelHands: 5,
            blindMultiplier: 200,
            sessionId,
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
            bettingLimit: 2,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and tableSize config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('reset', undefined, { tableSize: 6 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            tableSize: 6,
            sessionId,
          }),
        }),
      );
    });

    it('calls with rebuy command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('rebuy');
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'rebuy',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with skipaddon command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('skipaddon');
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'skipaddon',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and rebuy/addon config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('reset', undefined, { rebuyEnabled: true, addonEnabled: true, rebuyMaxCount: 5 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            rebuyEnabled: true,
            addonEnabled: true,
            rebuyMaxCount: 5,
            sessionId,
          }),
        }),
      );
    });

    it('calls with muck command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('muck');
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'muck',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with show command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('show');
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'show',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with fold command and humanPlayMs', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('fold', undefined, undefined, 300);
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'fold',
            amount: undefined,
            humanPlayMs: 300,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and cpuMetaAI config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await holdemApi.exec('reset', undefined, { cpuMetaAI: true });
      expect(mockFetch).toHaveBeenCalledWith(
        '/holdem/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            cpuMetaAI: true,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(holdemApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('omahaApi', () => {
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
      tableSize: 4,
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await omahaApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/omaha/exec', {
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
      await omahaApi.exec('fold');
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
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
      await omahaApi.exec('bet', 50);
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
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
      await omahaApi.exec('reset', undefined, { smallBlind: 10, bigBlind: 20 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            smallBlind: 10,
            bigBlind: 20,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and tournament config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('reset', undefined, {
        tournamentMode: true,
        blindLevelHands: 5,
        blindMultiplier: 200,
      });
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            tournamentMode: true,
            blindLevelHands: 5,
            blindMultiplier: 200,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and bettingLimit config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('reset', undefined, { bettingLimit: 2 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            bettingLimit: 2,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and tableSize config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('reset', undefined, { tableSize: 6 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            tableSize: 6,
            sessionId,
          }),
        }),
      );
    });

    it('calls with rebuy command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('rebuy');
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'rebuy',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with skipaddon command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('skipaddon');
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'skipaddon',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and rebuy/addon config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('reset', undefined, { rebuyEnabled: true, addonEnabled: true, rebuyMaxCount: 5 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            rebuyEnabled: true,
            addonEnabled: true,
            rebuyMaxCount: 5,
            sessionId,
          }),
        }),
      );
    });

    it('calls with muck command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('muck');
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'muck',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with show command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('show');
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'show',
            amount: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with fold command and humanPlayMs', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('fold', undefined, undefined, 300);
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'fold',
            amount: undefined,
            humanPlayMs: 300,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and cpuMetaAI config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await omahaApi.exec('reset', undefined, { cpuMetaAI: true });
      expect(mockFetch).toHaveBeenCalledWith(
        '/omaha/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            amount: undefined,
            cpuMetaAI: true,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(omahaApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('dramahaApi', () => {
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
      tableSize: 4,
    };

    /** The JSON body of the last fetch, parsed. */
    function lastBody(): Record<string, unknown> {
      const init = mockFetch.mock.calls.at(-1)?.[1] as { body: string };
      return JSON.parse(init.body);
    }

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await dramahaApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/dramaha/exec', expect.objectContaining({ method: 'POST' }));
      expect(lastBody()).toEqual({ command: 'reset', sessionId });
      expect(result).toEqual(payload);
    });

    it('shares the Hold-em betting commands', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await dramahaApi.exec('raise', 80);
      expect(lastBody()).toEqual({ command: 'raise', amount: 80, sessionId });
    });

    it('puts the draw indices at the top level of the body, where DramahaWebInput reads them', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await dramahaApi.exec('draw', undefined, { indices: [0, 2] });
      expect(mockFetch).toHaveBeenCalledWith('/dramaha/exec', expect.objectContaining({ method: 'POST' }));
      // NOT nested under `config`: the Go input has `indices` as a sibling of `command`.
      expect(lastBody()).toEqual({ command: 'draw', indices: [0, 2], sessionId });
    });

    it('sends an empty list when the player stands pat', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await dramahaApi.exec('draw', undefined, { indices: [] });
      expect(lastBody()).toEqual({ command: 'draw', indices: [], sessionId });
    });

    it('carries the table config through unchanged', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await dramahaApi.exec('reset', undefined, { smallBlind: 10, bigBlind: 20, cpuMetaAI: true });
      expect(lastBody()).toEqual({
        command: 'reset',
        smallBlind: 10,
        bigBlind: 20,
        cpuMetaAI: true,
        sessionId,
      });
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(dramahaApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('bigOApi / bigOHiLoApi', () => {
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
      tableSize: 4,
    };

    it('bigOApi posts to /bigo/exec', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await bigOApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/bigo/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', amount: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('bigOHiLoApi posts to /bigohilo/exec', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await bigOHiLoApi.exec('call', 20);
      expect(mockFetch).toHaveBeenCalledWith('/bigohilo/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'call', amount: 20, sessionId }),
      });
    });

    it('bigOApi throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(bigOApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('heartsApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      trickNumber: 1,
      currentPlayerIdx: 0,
      currentTrick: [],
      heartsBroken: false,
      passDirection: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      leadPlayerIdx: 0,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 100 },
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await heartsApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/hearts/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndices: undefined,
          cardIndex: undefined,
          config: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with play command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await heartsApi.exec('play', undefined, 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/hearts/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'play',
            cardIndices: undefined,
            cardIndex: 3,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with pass command and cardIndices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await heartsApi.exec('pass', [0, 1, 2]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/hearts/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'pass',
            cardIndices: [0, 1, 2],
            cardIndex: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await heartsApi.exec('reset', undefined, undefined, { cpuDifficulty: 2, pointLimit: 50 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/hearts/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            cardIndices: undefined,
            cardIndex: undefined,
            config: { cpuDifficulty: 2, pointLimit: 50 },
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(heartsApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('bidWhistApi.exec', () => {
    it('posts to /bidwhist/exec with the bid params', async () => {
      mockFetch.mockReturnValue(makeResponse({ players: [] }));
      await bidWhistApi.exec('bid', { bidTricks: 4, bidDirection: 0 });
      expect(mockFetch).toHaveBeenCalledWith('/bidwhist/exec', expect.objectContaining({ method: 'POST' }));
      const body = JSON.parse(mockFetch.mock.calls.at(-1)?.[1].body);
      expect(body.command).toBe('bid');
      expect(body.bidTricks).toBe(4);
      expect(body.bidDirection).toBe(0);
    });

    it('posts trump and exchange commands', async () => {
      mockFetch.mockReturnValue(makeResponse({ players: [] }));
      await bidWhistApi.exec('t', { trumpSuit: 1 });
      let body = JSON.parse(mockFetch.mock.calls.at(-1)?.[1].body);
      expect(body).toMatchObject({ command: 't', trumpSuit: 1 });

      await bidWhistApi.exec('e', { discardIndices: [0, 1, 2, 3, 4, 5] });
      body = JSON.parse(mockFetch.mock.calls.at(-1)?.[1].body);
      expect(body).toMatchObject({ command: 'e', discardIndices: [0, 1, 2, 3, 4, 5] });
    });
  });

  describe('gongzhuApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      trickNumber: 0,
      currentPlayerIdx: 0,
      currentTrick: [],
      heartsBroken: false,
      exposed: { pig: false, sheep: false, ace: false, doubler: false },
      exposableIndices: [],
      gameEndFlag: false,
      winnerIdx: -1,
      leadPlayerIdx: 0,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 1000 },
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await gongzhuApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/gongzhu/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndices: undefined,
          cardIndex: undefined,
          config: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with expose command and cardIndices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await gongzhuApi.exec('expose', [0, 1]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/gongzhu/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'expose',
            cardIndices: [0, 1],
            cardIndex: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with play command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await gongzhuApi.exec('play', undefined, 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/gongzhu/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'play',
            cardIndices: undefined,
            cardIndex: 3,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await gongzhuApi.exec('reset', undefined, undefined, { cpuDifficulty: 2, pointLimit: 500 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/gongzhu/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            cardIndices: undefined,
            cardIndex: undefined,
            config: { cpuDifficulty: 2, pointLimit: 500 },
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(gongzhuApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('spadesApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      trickNumber: 1,
      currentPlayerIdx: 0,
      bidPlayerIdx: 0,
      currentTrick: [],
      spadesBroken: false,
      gameEndFlag: false,
      winnerIdx: -1,
      leadPlayerIdx: 0,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 500, nilBonus: 100, bagPenaltyThreshold: 10 },
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await spadesApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/spades/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          bid: undefined,
          cardIndex: undefined,
          config: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with play command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await spadesApi.exec('play', undefined, 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/spades/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'play',
            bid: undefined,
            cardIndex: 3,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with bid command and bid value', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await spadesApi.exec('bid', 5);
      expect(mockFetch).toHaveBeenCalledWith(
        '/spades/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'bid',
            bid: 5,
            cardIndex: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await spadesApi.exec('reset', undefined, undefined, { cpuDifficulty: 2, pointLimit: 300 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/spades/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            bid: undefined,
            cardIndex: undefined,
            config: { cpuDifficulty: 2, pointLimit: 300 },
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(spadesApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('memoryApi.exec', () => {
    const payload = {
      players: [],
      board: [],
      phase: 0,
      currentPlayerIdx: 0,
      firstFlipPos: -1,
      secondFlipPos: -1,
      lastMatchResult: false,
      gameEndFlag: false,
      winnerIdx: -1,
      turnNumber: 0,
      message: '',
      config: { cpuDifficulty: 1 },
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await memoryApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/memory/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          position: undefined,
          config: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with flip command and position', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await memoryApi.exec('flip', 5);
      expect(mockFetch).toHaveBeenCalledWith(
        '/memory/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'flip',
            position: 5,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await memoryApi.exec('reset', undefined, { cpuDifficulty: 2 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/memory/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            position: undefined,
            config: { cpuDifficulty: 2 },
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(memoryApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('klondikeApi.exec', () => {
    const payload: {
      tableau: { card: { design: string; value: number } | null; faceUp: boolean }[][];
      stockCount: number;
      waste: { design: string; value: number }[];
      foundation: { design: string; value: number }[][];
      phase: number;
      moveCount: number;
      message: string;
    } = {
      tableau: [[{ card: { design: 'SPADE', value: 13 }, faceUp: true }]],
      stockCount: 20,
      waste: [{ design: 'CLOVER', value: 3 }],
      foundation: [[], [], [], []],
      phase: 0,
      moveCount: 0,
      message: '',
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await klondikeApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/klondike/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          from: undefined,
          to: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with draw command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await klondikeApi.exec('draw');
      expect(mockFetch).toHaveBeenCalledWith(
        '/klondike/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'draw',
            from: undefined,
            to: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with move command and from/to zones', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await klondikeApi.exec('move', { zone: 'waste' }, { zone: 'tableau', col: 3 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/klondike/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'move',
            from: { zone: 'waste' },
            to: { zone: 'tableau', col: 3 },
            sessionId,
          }),
        }),
      );
    });

    it('calls with move command with cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await klondikeApi.exec('move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/klondike/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'move',
            from: { zone: 'tableau', col: 0, cardIndex: 2 },
            to: { zone: 'tableau', col: 3 },
            sessionId,
          }),
        }),
      );
    });

    it('calls with giveup command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await klondikeApi.exec('giveup');
      expect(mockFetch).toHaveBeenCalledWith(
        '/klondike/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'giveup',
            from: undefined,
            to: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with hint command', async () => {
      const hintPayload = {
        ...payload,
        hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
      };
      mockFetch.mockReturnValue(makeResponse(hintPayload));
      const result = await klondikeApi.exec('hint');
      expect(mockFetch).toHaveBeenCalledWith(
        '/klondike/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'hint',
            from: undefined,
            to: undefined,
            sessionId,
          }),
        }),
      );
      expect(result.hint).toEqual(hintPayload.hint);
    });

    it('calls with autocomplete command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await klondikeApi.exec('autocomplete');
      expect(mockFetch).toHaveBeenCalledWith(
        '/klondike/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'autocomplete',
            from: undefined,
            to: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with log command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await klondikeApi.exec('log');
      expect(mockFetch).toHaveBeenCalledWith(
        '/klondike/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'log',
            from: undefined,
            to: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(klondikeApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('crazyeightsApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      currentPlayerIdx: 0,
      discardTop: null,
      drawPileCount: 0,
      chosenSuit: -1,
      gameEndFlag: false,
      winnerIdx: -1,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 200 },
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await crazyeightsApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/crazyeights/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndex: undefined,
          suit: undefined,
          config: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with play command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await crazyeightsApi.exec('play', 2);
      expect(mockFetch).toHaveBeenCalledWith(
        '/crazyeights/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'play',
            cardIndex: 2,
            suit: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with suit command and suit value', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await crazyeightsApi.exec('suit', undefined, 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/crazyeights/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'suit',
            cardIndex: undefined,
            suit: 3,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await crazyeightsApi.exec('reset', undefined, undefined, { cpuDifficulty: 2, pointLimit: 100 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/crazyeights/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            cardIndex: undefined,
            suit: undefined,
            config: { cpuDifficulty: 2, pointLimit: 100 },
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(crazyeightsApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('ginrummyApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      currentPlayerIdx: 0,
      discardTop: null,
      drawPileCount: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      knockerIdx: -1,
      knockerMelds: [],
      knockerDeadwood: [],
      isGin: false,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 100 },
    };

    it('calls the correct URL with reset command and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await ginrummyApi.exec('reset', undefined, { cpuDifficulty: 1, pointLimit: 100 });
      expect(mockFetch).toHaveBeenCalledWith('/ginrummy/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndex: undefined,
          cardIndices: undefined,
          config: { cpuDifficulty: 1, pointLimit: 100 },
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with discard command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await ginrummyApi.exec('discard', 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/ginrummy/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'discard',
            cardIndex: 3,
            cardIndices: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with layoff command and cardIndices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await ginrummyApi.exec('layoff', undefined, undefined, [0, 2]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/ginrummy/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'layoff',
            cardIndex: undefined,
            cardIndices: [0, 2],
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with log command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await ginrummyApi.exec('log');
      expect(mockFetch).toHaveBeenCalledWith(
        '/ginrummy/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'log',
            cardIndex: undefined,
            cardIndices: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(ginrummyApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('indianRummyApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      targetRounds: 3,
      currentPlayerIdx: 0,
      dealerIdx: 0,
      discardTop: null,
      drawPileCount: 0,
      wildJoker: null,
      wildRank: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      declarerIdx: -1,
      declarationValid: false,
      message: '',
      config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
    };

    it('calls the correct URL with reset command and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await indianRummyApi.exec('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      });
      expect(mockFetch).toHaveBeenCalledWith('/indianrummy/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndex: undefined,
          config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with discard command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await indianRummyApi.exec('discard', 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/indianrummy/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'discard',
            cardIndex: 3,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with declare command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await indianRummyApi.exec('declare', 13);
      expect(mockFetch).toHaveBeenCalledWith(
        '/indianrummy/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'declare',
            cardIndex: 13,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with log command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await indianRummyApi.exec('log');
      expect(mockFetch).toHaveBeenCalledWith(
        '/indianrummy/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'log',
            cardIndex: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(indianRummyApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('panApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      targetRounds: 3,
      currentPlayerIdx: 0,
      dealerIdx: 0,
      discardTop: null,
      drawPileCount: 0,
      deckSize: 320,
      winMeldCount: 11,
      gameEndFlag: false,
      winnerIdx: -1,
      panDeclarerIdx: -1,
      message: '',
      config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
    };

    it('calls the correct URL with reset command and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await panApi.exec('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      });
      expect(mockFetch).toHaveBeenCalledWith('/pan/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndices: undefined,
          cardIndex: undefined,
          meldOwner: undefined,
          meldIdx: undefined,
          config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with drawstock command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await panApi.exec('drawstock');
      expect(mockFetch).toHaveBeenCalledWith(
        '/pan/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'drawstock',
            cardIndices: undefined,
            cardIndex: undefined,
            meldOwner: undefined,
            meldIdx: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with meld command and cardIndices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await panApi.exec('meld', { cardIndices: [0, 1, 2] });
      expect(mockFetch).toHaveBeenCalledWith(
        '/pan/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'meld',
            cardIndices: [0, 1, 2],
            cardIndex: undefined,
            meldOwner: undefined,
            meldIdx: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with layoff command and meld target', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await panApi.exec('layoff', { meldOwner: 1, meldIdx: 0, cardIndex: 3 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/pan/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'layoff',
            cardIndices: undefined,
            cardIndex: 3,
            meldOwner: 1,
            meldIdx: 0,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with discard command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await panApi.exec('discard', { cardIndex: 5 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/pan/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'discard',
            cardIndices: undefined,
            cardIndex: 5,
            meldOwner: undefined,
            meldIdx: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(panApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('machiavelliApi.exec', () => {
    const payload = {
      players: [],
      table: [],
      phase: 0,
      roundNumber: 1,
      targetRounds: 3,
      currentPlayerIdx: 0,
      dealerIdx: 0,
      drawPileCount: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      roundWinnerIdx: -1,
      message: '',
      config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
    };

    it('calls the correct URL with reset command and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await machiavelliApi.exec('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      });
      expect(mockFetch).toHaveBeenCalledWith('/machiavelli/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with draw command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await machiavelliApi.exec('draw');
      expect(mockFetch).toHaveBeenCalledWith(
        '/machiavelli/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'draw', sessionId }),
        }),
      );
    });

    it('calls with newmeld command and hand indices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await machiavelliApi.exec('newmeld', { handIndices: [0, 1, 2] });
      expect(mockFetch).toHaveBeenCalledWith(
        '/machiavelli/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'newmeld', handIndices: [0, 1, 2], sessionId }),
        }),
      );
    });

    it('calls with layoff command and meld/hand indices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await machiavelliApi.exec('layoff', { meldIdx: 1, handIndex: 4 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/machiavelli/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'layoff', meldIdx: 1, handIndex: 4, sessionId }),
        }),
      );
    });

    it('calls with log command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await machiavelliApi.exec('log');
      expect(mockFetch).toHaveBeenCalledWith(
        '/machiavelli/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'log', sessionId }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(machiavelliApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('conquianApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      currentPlayerIdx: 0,
      discardTop: null,
      drawPileCount: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      roundWinnerIdx: -1,
      tookDiscard: false,
      message: '',
      config: { cpuDifficulty: 1, targetWins: 3 },
    };

    it('calls the correct URL with reset command and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await conquianApi.exec('reset', undefined, { cpuDifficulty: 1, targetWins: 3 });
      expect(mockFetch).toHaveBeenCalledWith('/conquian/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndex: undefined,
          config: { cpuDifficulty: 1, targetWins: 3 },
          meldGroups: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with discard command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await conquianApi.exec('discard', 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/conquian/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'discard',
            cardIndex: 3,
            config: undefined,
            meldGroups: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with meld command and meldGroups', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await conquianApi.exec('meld', undefined, undefined, [[0, 1, 2]]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/conquian/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'meld',
            cardIndex: undefined,
            config: undefined,
            meldGroups: [[0, 1, 2]],
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(conquianApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('chinchonApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      currentPlayerIdx: 0,
      discardTop: null,
      drawPileCount: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      knockerIdx: -1,
      knockerMelds: [],
      message: '',
      config: { cpuDifficulty: 1, playerCount: 2, knockThreshold: 5, eliminationLimit: 100 },
    };

    it('calls the correct URL with reset command and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await chinchonApi.exec('reset', undefined, {
        cpuDifficulty: 1,
        playerCount: 2,
        knockThreshold: 5,
        eliminationLimit: 100,
      });
      expect(mockFetch).toHaveBeenCalledWith('/chinchon/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndex: undefined,
          cardIndices: undefined,
          config: { cpuDifficulty: 1, playerCount: 2, knockThreshold: 5, eliminationLimit: 100 },
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with knock command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await chinchonApi.exec('knock', 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/chinchon/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'knock',
            cardIndex: 3,
            cardIndices: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with layoff command and cardIndices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await chinchonApi.exec('layoff', undefined, undefined, [0, 2]);
      expect(mockFetch).toHaveBeenCalledWith(
        '/chinchon/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'layoff',
            cardIndex: undefined,
            cardIndices: [0, 2],
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(chinchonApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('threethirteenApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      round: 1,
      wildRank: 3,
      dealCount: 3,
      currentPlayerIdx: 0,
      knockerIdx: -1,
      discardTop: null,
      drawPileCount: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      message: '',
      config: { cpuDifficulty: 1, playerCount: 2 },
    };

    it('calls the correct URL with reset command and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await threethirteenApi.exec('reset', undefined, { cpuDifficulty: 1, playerCount: 2 });
      expect(mockFetch).toHaveBeenCalledWith('/threethirteen/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          cardIndex: undefined,
          config: { cpuDifficulty: 1, playerCount: 2 },
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with knock command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await threethirteenApi.exec('knock', 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/threethirteen/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'knock',
            cardIndex: 3,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(threethirteenApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('baccaratApi.exec', () => {
    it('sends reset command', async () => {
      const mockResponse = { phase: 1, chips: 1000 };
      mockFetch.mockReturnValue(makeResponse(mockResponse));
      const result = await baccaratApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/baccarat/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          amount: undefined,
          betType: undefined,
          playerPairBet: undefined,
          bankerPairBet: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(mockResponse);
    });

    it('sends bet command with amount and betType', async () => {
      mockFetch.mockReturnValue(makeResponse({ phase: 2 }));
      await baccaratApi.exec('bet', 100, 0);
      expect(mockFetch).toHaveBeenCalledWith('/baccarat/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'bet',
          amount: 100,
          betType: 0,
          playerPairBet: undefined,
          bankerPairBet: undefined,
          sessionId,
        }),
      });
    });

    it('sends bet command with banker betType', async () => {
      mockFetch.mockReturnValue(makeResponse({ phase: 2 }));
      await baccaratApi.exec('bet', 100, 1);
      expect(mockFetch).toHaveBeenCalledWith('/baccarat/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'bet',
          amount: 100,
          betType: 1,
          playerPairBet: undefined,
          bankerPairBet: undefined,
          sessionId,
        }),
      });
    });

    it('sends bet command with side bets', async () => {
      mockFetch.mockReturnValue(makeResponse({ phase: 2 }));
      await baccaratApi.exec('bet', 100, 0, 10, 20);
      expect(mockFetch).toHaveBeenCalledWith('/baccarat/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'bet',
          amount: 100,
          betType: 0,
          playerPairBet: 10,
          bankerPairBet: 20,
          sessionId,
        }),
      });
    });

    it('sends log command', async () => {
      mockFetch.mockReturnValue(makeResponse({ entries: [] }));
      await baccaratApi.exec('log');
      expect(mockFetch).toHaveBeenCalledWith('/baccarat/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'log',
          amount: undefined,
          betType: undefined,
          playerPairBet: undefined,
          bankerPairBet: undefined,
          sessionId,
        }),
      });
    });

    it('sends clearhistory command', async () => {
      mockFetch.mockReturnValue(makeResponse({ phase: 1 }));
      await baccaratApi.exec('clearhistory');
      expect(mockFetch).toHaveBeenCalledWith('/baccarat/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'clearhistory',
          amount: undefined,
          betType: undefined,
          playerPairBet: undefined,
          bankerPairBet: undefined,
          sessionId,
        }),
      });
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(baccaratApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('actionLogApi', () => {
    const logPayload = { entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'hit card', cards: [] }] };

    describe.each([
      ['blackjack', actionLogApi.blackjack],
      ['poker', actionLogApi.poker],
      ['oldmaid', actionLogApi.oldmaid],
      ['daifugo', actionLogApi.daifugo],
      ['sevens', actionLogApi.sevens],
      ['doubt', actionLogApi.doubt],
      ['holdem', actionLogApi.holdem],
      ['omaha', actionLogApi.omaha],
      ['hearts', actionLogApi.hearts],
      ['spades', actionLogApi.spades],
      ['memory', actionLogApi.memory],
      ['klondike', actionLogApi.klondike],
      ['baccarat', actionLogApi.baccarat],
      ['crazyeights', actionLogApi.crazyeights],
      ['ginrummy', actionLogApi.ginrummy],
      ['indianrummy', actionLogApi.indianrummy],
      ['pan', actionLogApi.pan],
      ['spider', actionLogApi.spider],
      ['indianpoker', actionLogApi.indianpoker],
    ])('actionLogApi.%s', (gameName, apiFn) => {
      it(`calls /${gameName}/exec with log command`, async () => {
        mockFetch.mockReturnValue(makeResponse(logPayload));
        const result = await apiFn();
        expect(mockFetch).toHaveBeenCalledWith(`/${gameName}/exec`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ command: 'log', sessionId }),
        });
        expect(result).toEqual(logPayload);
      });
    });
  });

  describe('spiderApi.exec', () => {
    const payload = {
      tableau: [[{ card: { design: 'SPADE', value: 5 }, faceUp: true }]],
      stockCount: 50,
      completedSuits: 0,
      phase: 0,
      moveCount: 0,
      score: 500,
      difficulty: 1,
      canUndo: false,
      isStalemate: false,
      message: '',
    };

    it('calls with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await spiderApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/spider/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', from: undefined, to: undefined, config: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with deal command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await spiderApi.exec('deal');
      expect(mockFetch).toHaveBeenCalledWith(
        '/spider/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'deal', from: undefined, to: undefined, config: undefined, sessionId }),
        }),
      );
    });

    it('calls with move command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await spiderApi.exec('move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/spider/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'move',
            from: { zone: 'tableau', col: 0, cardIndex: 2 },
            to: { zone: 'tableau', col: 3 },
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await spiderApi.exec('reset', undefined, undefined, { difficulty: 2 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/spider/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            from: undefined,
            to: undefined,
            config: { difficulty: 2 },
            sessionId,
          }),
        }),
      );
    });
  });

  describe('indianpokerApi.exec', () => {
    const payload = {
      players: [],
      pot: 0,
      sidePots: [],
      dealerIdx: 0,
      currentTurn: 0,
      phase: 0,
      gameEndFlag: false,
      lastBet: 0,
      minRaise: 20,
      bettingLimit: 2,
      raiseCount: 0,
      maxBetAmount: 0,
      roundResults: [],
      cpuActions: [],
      handCount: 0,
      ante: 10,
      message: '',
    };

    it('calls /indianpoker/exec with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));

      const result = await indianpokerApi.exec('reset');

      expect(mockFetch).toHaveBeenCalledWith('/indianpoker/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          amount: undefined,
          humanPlayMs: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls /indianpoker/exec with bet command and amount', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));

      await indianpokerApi.exec('bet', 40);

      expect(mockFetch).toHaveBeenCalledWith('/indianpoker/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'bet',
          amount: 40,
          humanPlayMs: undefined,
          sessionId,
        }),
      });
    });

    it('calls /indianpoker/exec with config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));

      await indianpokerApi.exec('reset', undefined, { ante: 20, bettingLimit: 1, cpuMetaAI: false });

      expect(mockFetch).toHaveBeenCalledWith('/indianpoker/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          amount: undefined,
          humanPlayMs: undefined,
          ante: 20,
          bettingLimit: 1,
          cpuMetaAI: false,
          sessionId,
        }),
      });
    });

    it('calls /indianpoker/exec with humanPlayMs', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));

      await indianpokerApi.exec('call', undefined, undefined, 1500);

      expect(mockFetch).toHaveBeenCalledWith('/indianpoker/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'call',
          amount: undefined,
          humanPlayMs: 1500,
          sessionId,
        }),
      });
    });
  });

  describe('pyramidApi.exec', () => {
    const payload = {
      pyramid: [],
      stockCount: 20,
      waste: [],
      phase: 0,
      moveCount: 0,
      canUndo: false,
      isStalemate: false,
      message: '',
    };

    it('calls with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await pyramidApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/pyramid/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', card1: undefined, card2: undefined, n: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with undo_n command and n value', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await pyramidApi.exec('undo_n', undefined, undefined, 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/pyramid/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'undo_n', card1: undefined, card2: undefined, n: 3, sessionId }),
        }),
      );
    });
  });

  describe('tripeaksApi.exec', () => {
    const payload = {
      layout: [],
      stockCount: 20,
      waste: [],
      phase: 0,
      moveCount: 0,
      canUndo: false,
      isStalemate: false,
      message: '',
    };

    it('calls with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await tripeaksApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/tripeaks/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', row: undefined, col: undefined, n: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with undo_n command and n value', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await tripeaksApi.exec('undo_n', undefined, undefined, 5);
      expect(mockFetch).toHaveBeenCalledWith(
        '/tripeaks/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'undo_n', row: undefined, col: undefined, n: 5, sessionId }),
        }),
      );
    });
  });

  describe('golfApi.exec', () => {
    const payload = {
      layout: [],
      stockCount: 20,
      waste: [],
      phase: 0,
      moveCount: 0,
      canUndo: false,
      isStalemate: false,
      message: '',
    };

    it('calls with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await golfApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/golf/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', col: undefined, n: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with undo_n command and n value', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await golfApi.exec('undo_n', undefined, 4);
      expect(mockFetch).toHaveBeenCalledWith(
        '/golf/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'undo_n', col: undefined, n: 4, sessionId }),
        }),
      );
    });
  });

  describe('twoTenJackApi.exec', () => {
    const payload = {
      players: [],
      phase: 0,
      roundNumber: 1,
      trickNumber: 1,
      currentPlayerIdx: 0,
      declarerIdx: 0,
      trumpSuit: -1,
      currentTrick: [],
      gameEndFlag: false,
      winnerTeam: -1,
      leadPlayerIdx: 0,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 50 },
    };

    it('calls the correct URL with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await twoTenJackApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/twotenjack/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          trumpSuit: undefined,
          cardIndex: undefined,
          config: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with declare command and trumpSuit', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await twoTenJackApi.exec('declare', 3);
      expect(mockFetch).toHaveBeenCalledWith(
        '/twotenjack/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'declare',
            trumpSuit: 3,
            cardIndex: undefined,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with play command and cardIndex', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await twoTenJackApi.exec('play', undefined, 2);
      expect(mockFetch).toHaveBeenCalledWith(
        '/twotenjack/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'play',
            trumpSuit: undefined,
            cardIndex: 2,
            config: undefined,
            sessionId,
          }),
        }),
      );
    });

    it('calls with reset and config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await twoTenJackApi.exec('reset', undefined, undefined, { cpuDifficulty: 2, pointLimit: 100 });
      expect(mockFetch).toHaveBeenCalledWith(
        '/twotenjack/exec',
        expect.objectContaining({
          body: JSON.stringify({
            command: 'reset',
            trumpSuit: undefined,
            cardIndex: undefined,
            config: { cpuDifficulty: 2, pointLimit: 100 },
            sessionId,
          }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(twoTenJackApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('letitrideApi.exec', () => {
    it('sends reset command', async () => {
      const mockResponse = { phase: 0, chips: 1000 };
      mockFetch.mockReturnValue(makeResponse(mockResponse));
      const result = await letitrideApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/letitride/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', amount: undefined, sessionId }),
      });
      expect(result).toEqual(mockResponse);
    });

    it('sends bet command with amount', async () => {
      mockFetch.mockReturnValue(makeResponse({ phase: 1 }));
      await letitrideApi.exec('bet', 100);
      expect(mockFetch).toHaveBeenCalledWith('/letitride/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'bet', amount: 100, sessionId }),
      });
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(letitrideApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  // canfieldApi covers the createSolitaireMoveApi factory body — the seven
  // games that use it (Canfield, FreeCell, Yukon, Scorpion, Accordion,
  // FortyThieves, Calculation) all share the same wire shape, so one
  // representative test pins the factory output for the whole group.
  describe('canfieldApi.exec (createSolitaireMoveApi factory smoke test)', () => {
    it('sends reset command with no zones', async () => {
      const mockResponse = { tableau: [], stockCount: 0, phase: 0, message: '' };
      mockFetch.mockReturnValue(makeResponse(mockResponse));
      const result = await canfieldApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/canfield/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', from: undefined, to: undefined, sessionId }),
      });
      expect(result).toEqual(mockResponse);
    });

    it('sends undo_n command with batch count', async () => {
      mockFetch.mockReturnValue(makeResponse({ phase: 1 }));
      await canfieldApi.exec('undo_n', undefined, undefined, 3);
      expect(mockFetch).toHaveBeenCalledWith('/canfield/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'undo_n', from: undefined, to: undefined, n: 3, sessionId }),
      });
    });
  });

  describe('agnesApi.exec', () => {
    it('sends deal command with no zones', async () => {
      const mockResponse = { tableau: [], stockCount: 0, foundation: [], baseRank: 0, phase: 0, message: '' };
      mockFetch.mockReturnValue(makeResponse(mockResponse));
      const result = await agnesApi.exec('deal');
      expect(mockFetch).toHaveBeenCalledWith('/agnes/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'deal', from: undefined, to: undefined, sessionId }),
      });
      expect(result).toEqual(mockResponse);
    });

    it('sends move command with from and to zones', async () => {
      mockFetch.mockReturnValue(makeResponse({ phase: 0 }));
      await agnesApi.exec('move', { zone: 'tableau', col: 0, cardIndex: 0 }, { zone: 'foundation' });
      expect(mockFetch).toHaveBeenCalledWith('/agnes/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'move',
          from: { zone: 'tableau', col: 0, cardIndex: 0 },
          to: { zone: 'foundation' },
          sessionId,
        }),
      });
    });
  });

  describe('shitheadApi.exec', () => {
    const payload = {
      players: [],
      currentTurn: 0,
      currentSource: 'hand',
      discardPile: [],
      stockSize: 0,
      skipNext: false,
      sevenActive: false,
      gameEndFlag: false,
      config: {},
      cpuActions: [],
      message: '',
    };

    it('calls with reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await shitheadApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/shithead/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('calls with play command and indices', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await shitheadApi.exec('play', { indices: [0, 1] });
      expect(mockFetch).toHaveBeenCalledWith(
        '/shithead/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'play', indices: [0, 1], sessionId }),
        }),
      );
    });

    it('calls with reset command and config', async () => {
      const config = {
        magicTwo: true,
        magicSeven: false,
        magicEight: true,
        magicTen: true,
        fourOfAKindBurn: true,
        cpuDifficulty: 2,
      };
      mockFetch.mockReturnValue(makeResponse(payload));
      await shitheadApi.exec('reset', { config });
      expect(mockFetch).toHaveBeenCalledWith(
        '/shithead/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'reset', config, sessionId }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(shitheadApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('slapjackApi', () => {
    const payload = {
      phase: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      currentTurnIdx: 0,
      isHumanTurn: true,
      isTopJack: false,
      centerPileSize: 0,
      topCard: null,
      players: [],
      cpuDifficulty: 1,
      pendingKind: 0,
      pendingDeadlineMs: 0,
      lastEventKind: 0,
      lastEventPlayerIdx: 0,
      message: '',
    };

    it('sends reset command', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await slapjackApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/slapjack/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('sends step / slap / tick / log commands', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      for (const cmd of ['step', 'slap', 'tick', 'log'] as const) {
        await slapjackApi.exec(cmd);
        expect(mockFetch).toHaveBeenCalledWith(
          '/slapjack/exec',
          expect.objectContaining({
            body: JSON.stringify({ command: cmd, sessionId }),
          }),
        );
      }
    });

    it('sends reset with config', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await slapjackApi.exec('reset', { config: { cpuDifficulty: 2 } });
      expect(mockFetch).toHaveBeenCalledWith(
        '/slapjack/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2 }, sessionId }),
        }),
      );
    });

    it('throws on HTTP error', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(slapjackApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });

  describe('blackjackswitchApi.exec', () => {
    const payload = {
      hands: [],
      dealerCards: [],
      dealerScore: 0,
      phase: 1,
      currentHandIdx: 0,
      chips: 1000,
      switched: false,
      dealerPushed22: false,
      overallResult: 0,
      totalPayout: 0,
      message: '',
    };

    it('calls reset with the correct URL and body', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await blackjackswitchApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/blackjackswitch/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: 'reset', amount: undefined, sessionId }),
      });
      expect(result).toEqual(payload);
    });

    it('forwards bet amount in the body', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      await blackjackswitchApi.exec('bet', 100);
      expect(mockFetch).toHaveBeenCalledWith(
        '/blackjackswitch/exec',
        expect.objectContaining({
          body: JSON.stringify({ command: 'bet', amount: 100, sessionId }),
        }),
      );
    });
  });

  describe('montecarloApi.exec', () => {
    const payload = {
      board: Array.from({ length: 5 }, () => Array.from({ length: 5 }, () => ({ card: null }))),
      phase: 0,
      stockCount: 27,
      removedCount: 0,
      dealCount: 0,
      canUndo: false,
      isStalemate: false,
      message: '',
    };

    it('calls reset with the correct URL and body', async () => {
      mockFetch.mockReturnValue(makeResponse(payload));
      const result = await montecarloApi.exec('reset');
      expect(mockFetch).toHaveBeenCalledWith('/montecarlo/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          command: 'reset',
          fromR: undefined,
          fromC: undefined,
          toR: undefined,
          toC: undefined,
          sessionId,
        }),
      });
      expect(result).toEqual(payload);
    });

    it('throws on non-OK responses', async () => {
      mockFetch.mockReturnValue(makeResponse(null, false, 500));
      await expect(montecarloApi.exec('reset')).rejects.toThrow('HTTP error: 500');
    });
  });
});

describe('casinoholdemApi', () => {
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

  it('sends reset command', async () => {
    const mockResponse = { phase: 1, chips: 1000 };
    mockFetch.mockReturnValue(makeResponse(mockResponse));
    const result = await casinoholdemApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/casinoholdem/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', amount: undefined, bonusBet: undefined, sessionId }),
    });
    expect(result).toEqual(mockResponse);
  });

  it('sends bet command with ante and AA bonus', async () => {
    mockFetch.mockReturnValue(makeResponse({ phase: 2 }));
    await casinoholdemApi.exec('bet', 100, 10);
    expect(mockFetch).toHaveBeenCalledWith('/casinoholdem/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'bet', amount: 100, bonusBet: 10, sessionId }),
    });
  });

  it('sends call command', async () => {
    mockFetch.mockReturnValue(makeResponse({ phase: 3 }));
    await casinoholdemApi.exec('call');
    expect(mockFetch).toHaveBeenCalledWith('/casinoholdem/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'call', amount: undefined, bonusBet: undefined, sessionId }),
    });
  });

  it('sends fold command', async () => {
    mockFetch.mockReturnValue(makeResponse({ phase: 3 }));
    await casinoholdemApi.exec('fold');
    expect(mockFetch).toHaveBeenCalledWith('/casinoholdem/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'fold', amount: undefined, bonusBet: undefined, sessionId }),
    });
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(makeResponse(null, false, 500));
    await expect(casinoholdemApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
