import { useEffect, useMemo, useState } from 'react';
import { faroApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { FaroResponse } from '../types/card';
import { FaroPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { valueName } from '../utils/cardUtils';
import { FARO_HELP, parseFaroCommand } from '../utils/cli/commands/faroCommands';
import { formatFaroState } from '../utils/cli/formatters/faroFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { FARO_RANK_COUNT, mergeSeenCards, remainingByRank } from '../utils/faroCaseKeeper';

/** Rank values on the Faro layout, A (1) through K (13). */
const RANKS = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13] as const;

/** Chip amounts the player may stake on a single click. */
const CHIP_AMOUNTS = [10, 50, 100] as const;

/** Maps a rank value (1..13) to its short display label (A, 2..10, J, Q, K). */
const RANK_LABELS: Readonly<Record<number, string>> = {
  1: 'A',
  11: 'J',
  12: 'Q',
  13: 'K',
};

/** Returns the short layout label for a rank value. */
function rankLabel(rank: number): string {
  return RANK_LABELS[rank] ?? String(rank);
}

/** Faro tutorial step definitions. */
const FARO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="faro-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="faro-layout"]',
    messageKey: 'tutorial.layout',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="faro-deal"]',
    messageKey: 'tutorial.deal',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="faro-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric Faro phases to i18n phase-label keys. */
const FARO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FaroPhase.BETTING]: 'betting',
  [FaroPhase.TURN]: 'turn',
  [FaroPhase.CALL]: 'call',
  [FaroPhase.ROUND_END]: 'roundEnd',
  [FaroPhase.GAME_END]: 'gameEnd',
};

/** Renders the Faro game page: a 19th-century single-player-vs-bank banking game. */
export const FaroPage = withTutorial(FaroPageContent, 'faro', FARO_TUTORIAL_STEPS);

/** Inner content of the Faro page, wrapped by TutorialProvider. */
function FaroPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('faro');
  const { state, loading, error, exec, retry } = useGameApi(faroApi.exec);

  const [chipAmount, setChipAmount] = useState<number>(CHIP_AMOUNTS[0]);
  const [copper, setCopper] = useState(false);
  const [callOrder, setCallOrder] = useState<number[]>([]);

  // Case keeper: the server sends only the current turn's cards, so accumulate
  // every revealed card into a local set of unique keys and derive the
  // per-rank remaining counts from it. Reset each round (reset / next).
  const [seenKeys, setSeenKeys] = useState<Set<string>>(() => new Set());

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // Merge each turn's newly revealed cards (soda, last turn, and any call cards)
  // into the running set. Only new keys grow the set, so returning the previous
  // reference when nothing changed avoids a redundant re-render.
  useEffect(() => {
    if (!state) return;
    setSeenKeys((prev) => {
      const next = mergeSeenCards(prev, [state.soda, state.losingCard, state.winningCard, ...state.callCards]);
      return next.size === prev.size ? prev : next;
    });
  }, [state]);

  const remaining = useMemo(() => remainingByRank(seenKeys), [seenKeys]);

  const phaseNames = usePhaseNames('faro', FARO_PHASE_KEYS);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('faro');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('faro', state);
  const cliConfig: CliGameConfig<FaroResponse, Parameters<typeof faroApi.exec>> = useMemo(
    () => ({
      gameName: 'faro',
      parseCommand: parseFaroCommand,
      formatResponse: formatFaroState,
      helpText: FARO_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();

  if (!state)
    return <GameSkeleton gameKey="faro" layout={{ kind: 'casino-table', sections: [2], footerStyle: 'bet' }} />;

  const isGameEnd = state.phase === FaroPhase.GAME_END || state.gameEndFlag;
  const phaseName = phaseNames[state.phase] ?? '';
  const isBetting = state.phase === FaroPhase.BETTING && !isGameEnd;
  const isCall = state.phase === FaroPhase.CALL && !isGameEnd;
  const isRoundEnd = state.phase === FaroPhase.ROUND_END && !isGameEnd;
  const humanWon = isRoundEnd && state.totalPayout > 0;

  const betFor = (rank: number) => state.bets.find((b) => b.rank === rank);

  const handleReset = () => {
    hideActionLog();
    setCallOrder([]);
    setSeenKeys(new Set());
    exec('reset');
  };

  const handleDeal = () => exec('deal');

  const handleNext = () => {
    setCallOrder([]);
    setSeenKeys(new Set());
    exec('next');
  };

  /** Toggle a call card (tracked by its index) into / out of the predicted order.
   * Tracking by index — not rank — keeps selection correct when the three
   * remaining cards share a rank (e.g. two 5s and a 9). */
  const toggleCallCard = (idx: number) => {
    setCallOrder((prev) => (prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]));
  };

  const submitCall = () => {
    // Map the picked card indices to their ranks (the order the backend expects).
    exec('call', { order: callOrder.map((i) => state.callCards[i].value) });
    setCallOrder([]);
  };

  const skipCall = () => {
    exec('call', { order: [] });
    setCallOrder([]);
  };

  return (
    <GamePageShell
      title={tc('nav.faro')}
      gameThemeBg={gameTheme.faro.bg}
      phaseName={phaseName}
      gamePath="/faro"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Chips / payout / turns */}
            <div
              className="mb-3 p-2 rounded bg-black/30 flex flex-wrap justify-center gap-x-6 gap-y-1 text-sm"
              data-tutorial="faro-info"
            >
              <span className="font-semibold text-ds-warning">{t('chips', { count: state.chips })}</span>
              <span className="text-ds-text-primary">{t('payout', { amount: state.totalPayout })}</span>
              <span className="text-ds-text-muted">
                {t('turns', { played: state.turnsPlayed, total: state.turnsTotal })}
              </span>
              <span className="text-ds-text-muted">{t('remaining', { count: state.remaining })}</span>
            </div>

            {/* Soda */}
            {state.soda && (
              <div className="mb-2 text-center text-ds-text-muted text-xs flex items-center justify-center gap-2">
                <span>{t('soda')}</span>
                <CardImage card={state.soda} width={Math.round(cardWidth * 0.7)} />
              </div>
            )}

            {/* 13-rank betting layout */}
            <div className="mb-3 p-3 rounded bg-black/20" data-tutorial="faro-layout">
              <div className="text-ds-text-muted text-xs mb-2 text-center">{t('layoutTitle')}</div>
              {isBetting && copper && (
                <div
                  className="mb-2 text-center text-xs font-semibold text-ds-accent rounded bg-ds-accent/15 py-1"
                  data-testid="faro-copper-mode"
                >
                  🪙 {t('copperModeHint')}
                </div>
              )}
              <div className="grid grid-cols-7 gap-2 justify-items-center">
                {RANKS.map((rank) => {
                  const bet = betFor(rank);
                  return (
                    <button
                      key={`rank-${rank}`}
                      type="button"
                      onClick={() => isBetting && exec('bet', { rank, amount: chipAmount, copper })}
                      disabled={!isBetting || loading}
                      className={`relative w-12 h-14 rounded border text-lg font-bold transition-all ${
                        bet?.copper
                          ? 'border-ds-accent bg-ds-accent/20 text-ds-accent'
                          : bet
                            ? badgeWarningColors
                            : 'border-white/30 bg-black/30 text-ds-text-primary'
                      } ${isBetting ? 'cursor-pointer hover:border-ds-warning hover:-translate-y-0.5' : 'cursor-default opacity-80'}`}
                      data-testid={`rank-${rank}`}
                      aria-label={`${t('rankName')} ${rankLabel(rank)}${bet?.copper ? ` (${t('copperTag')})` : ''}`}
                    >
                      {/* Copper bets (betting on the rank to lose) are marked with a coin icon + accent
                          colour so they read differently from normal win bets at a glance (#2695). */}
                      {bet?.copper && (
                        <span className="absolute -top-1 -right-1 text-[10px]" aria-hidden="true">
                          🪙
                        </span>
                      )}
                      {rankLabel(rank)}
                      {bet && (
                        <span className="absolute -bottom-1 left-1/2 -translate-x-1/2 text-[10px] whitespace-nowrap">
                          {bet.amount}
                          {bet.copper ? ` ${t('copperTag')}` : ''}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Case keeper: remaining cards per rank, accumulated from revealed cards */}
            <div className="mb-3 p-3 rounded bg-black/20" data-testid="case-keeper">
              <div className="text-ds-text-muted text-xs mb-2 text-center">{t('caseKeeperTitle')}</div>
              <div className="grid grid-cols-7 gap-1.5 justify-items-center">
                {RANKS.map((rank) => {
                  const left = remaining[rank] ?? FARO_RANK_COUNT;
                  const depleted = left === 0;
                  return (
                    <div
                      key={`case-${rank}`}
                      className={`flex flex-col items-center rounded border px-1 py-0.5 text-center leading-tight ${
                        depleted ? 'border-white/10 bg-black/30 opacity-50' : 'border-white/25 bg-black/20'
                      }`}
                      data-testid={`case-keeper-rank-${rank}`}
                      role="img"
                      aria-label={t('caseKeeperCell', { rank: valueName(rank), count: left })}
                    >
                      <span className="text-[11px] font-bold text-ds-text-primary" aria-hidden="true">
                        {valueName(rank)}
                      </span>
                      <span
                        className={`text-[13px] font-semibold tabular-nums ${
                          depleted ? 'text-ds-text-muted' : 'text-ds-warning'
                        }`}
                      >
                        {left}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Current bets list */}
            <div className="mb-3 p-2 rounded bg-black/20">
              <div className="text-ds-text-muted text-xs mb-1">{t('betsTitle')}</div>
              {state.bets.length > 0 ? (
                <div className="flex flex-col gap-1">
                  {state.bets.map((b) => (
                    <div key={`bet-${b.rank}`} className="flex items-center gap-2 text-sm">
                      <span className="text-ds-text-primary">
                        {t('betLine', { rank: rankLabel(b.rank), amount: b.amount })}
                      </span>
                      {b.copper && <span className="text-ds-accent text-xs font-semibold">{t('copperTag')}</span>}
                      {isBetting && (
                        <button
                          type="button"
                          className="text-ds-text-muted text-xs underline"
                          onClick={() => exec('clearBet', { rank: b.rank })}
                          disabled={loading}
                          data-testid={`clear-bet-${b.rank}`}
                        >
                          {t('clearBet')}
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-ds-text-muted text-sm">{t('noBets')}</div>
              )}
            </div>

            {/* Last turn reveal */}
            {(state.losingCard || state.winningCard) && (
              <div className="mb-3 p-3 rounded bg-black/20 text-center">
                <div className="text-ds-text-muted text-xs mb-2">{t('lastTurnTitle')}</div>
                <div className="flex justify-center gap-6">
                  <div className="flex flex-col items-center gap-1">
                    <span className="text-ds-error text-xs">{t('losing')}</span>
                    {state.losingCard && <CardImage card={state.losingCard} width={cardWidth} />}
                  </div>
                  <div className="flex flex-col items-center gap-1">
                    <span className="text-ds-success text-xs">{t('winning')}</span>
                    {state.winningCard && <CardImage card={state.winningCard} width={cardWidth} />}
                  </div>
                </div>
                {state.split && <div className="mt-2 text-ds-warning text-sm font-semibold">{t('split')}</div>}
              </div>
            )}

            <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center cursor-pointer min-h-[44px]">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {/* Call phase order picker */}
            {isCall && (
              <div className="mb-3 p-3 rounded bg-black/30 text-center">
                <div className="text-ds-text-primary text-sm mb-1">{t('callTitle')}</div>
                <div className="text-ds-text-muted text-xs mb-2">{t('callHint')}</div>
                <div className="flex flex-wrap justify-center gap-3 mb-2">
                  {state.callCards.map((card, i) => {
                    const pos = callOrder.indexOf(i);
                    const isSelected = pos >= 0;
                    return (
                      <button
                        key={`call-card-${card.value}-${i}`}
                        type="button"
                        onClick={() => toggleCallCard(i)}
                        disabled={loading}
                        aria-pressed={isSelected}
                        className={`relative rounded-md border-2 transition-all ${focusRingWhite} ${
                          isSelected
                            ? 'border-ds-warning -translate-y-1'
                            : 'border-transparent hover:border-white/40 hover:-translate-y-0.5'
                        } ${loading ? 'cursor-default' : 'cursor-pointer'}`}
                        style={{ background: 'none', padding: 0, lineHeight: 0 }}
                        data-testid={`call-card-${card.value}-${i}`}
                        aria-label={`${t('rankName')} ${rankLabel(card.value)}${
                          isSelected ? `, ${t('callSlot', { n: pos + 1 })}` : ''
                        }`}
                      >
                        <CardImage card={card} width={cardWidth} />
                        {isSelected && (
                          <span
                            className="absolute -top-2 -right-2 flex h-6 w-6 items-center justify-center rounded-full bg-ds-warning text-black text-xs font-bold shadow"
                            aria-hidden="true"
                            data-testid={`call-order-badge-${card.value}-${i}`}
                          >
                            {pos + 1}
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>
                <div className="flex justify-center gap-2">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={submitCall}
                    disabled={loading || callOrder.length !== state.callCards.length}
                  >
                    {t('callButton')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={skipCall} disabled={loading}>
                    {t('skipCall')}
                  </button>
                </div>
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer: chip selector, copper toggle, deal / next / reset */}
          <GameFooter className={`${gameTheme.faro.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />

            {isBetting && (
              <div className="mb-2 flex flex-wrap items-center gap-2">
                <span className="text-ds-text-muted text-xs">{t('betAmountLabel')}</span>
                {CHIP_AMOUNTS.map((amt) => (
                  <button
                    key={`chip-${amt}`}
                    type="button"
                    onClick={() => setChipAmount(amt)}
                    className={`px-3 py-1 rounded text-sm font-semibold transition-all ${
                      chipAmount === amt ? 'bg-ds-warning text-black' : 'bg-black/30 text-ds-text-primary'
                    }`}
                    data-testid={`chip-${amt}`}
                  >
                    {amt}
                  </button>
                ))}
                <button
                  type="button"
                  onClick={() => setCopper((c) => !c)}
                  className={`px-3 py-1 rounded text-sm font-semibold transition-all ${
                    copper ? 'bg-ds-accent text-black' : 'bg-black/30 text-ds-text-primary'
                  }`}
                  data-testid="copper-toggle"
                  aria-pressed={copper}
                >
                  {copper ? t('copperOn') : t('copperOff')}
                </button>
                <button
                  type="button"
                  className="text-ds-text-muted text-xs underline ml-1"
                  onClick={() => exec('clearAll')}
                  disabled={loading || state.bets.length === 0}
                  data-testid="clear-all"
                >
                  {t('clearAll')}
                </button>
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center">
              {isGameEnd && <span className="text-ds-text-primary text-sm font-semibold mr-1">{t('gameEnd')}</span>}

              {(isBetting || state.phase === FaroPhase.TURN) && !isGameEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleDeal}
                  disabled={loading}
                  data-tutorial="faro-deal"
                  data-testid="deal-button"
                >
                  {t('deal')}
                </button>
              )}

              {isRoundEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNext}
                  disabled={loading}
                  data-testid="next-button"
                >
                  {t('next')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="faro-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
