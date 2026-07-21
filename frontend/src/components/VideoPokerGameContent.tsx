import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useLocalStorageToggle } from '../hooks/useLocalStorageToggle';
import { useVideoPokerStats } from '../hooks/useVideoPokerStats';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, VideoPokerResponse } from '../types/card';
import { VideoPokerPhase } from '../types/phases';
import type { CliGameConfig } from '../utils/cli/types';
import { getVideoPokerBaseHint } from '../utils/hints/videoPokerBaseHint';
import { evaluateJokerPokerMadeHand } from '../utils/jokerPokerMadeHand';
import {
  VIDEO_POKER_MAX_BET,
  videoPokerHandNameToRowKey,
  videoPokerPayoutCell,
  videoPokerPayoutRows,
} from '../utils/videoPokerPayout';
import { videoPokerNet, videoPokerWinRate } from '../utils/videoPokerStats';
import { ActionLogPanel } from './ActionLogPanel';
import { CliTerminal } from './cli/CliTerminal';
import { CliToggle } from './cli/CliToggle';
import { ErrorAlert } from './ErrorAlert';
import { GameFooter } from './GameFooter';
import { GameMessageBox } from './GameMessageBox';
import { GamePageShell } from './GamePageShell';
import { GameResetButton } from './GameResetButton';
import { HintTooltip } from './hint/HintTooltip';
import { AnimatedCard } from './motion/AnimatedCard';

/** Per-variant predicate identifying wild cards (Deuces Wild = twos, Joker Poker = jokers, default = none). */
const WILD_CARD_PREDICATE: Record<'videopoker' | 'deuceswild' | 'jokerpoker', (card: Card) => boolean> = {
  videopoker: () => false,
  deuceswild: (card) => card.value === 2,
  jokerpoker: (card) => card.design === 'JOKER',
};

/** Props for the VideoPokerGameContent shared component. */
export interface VideoPokerGameContentProps {
  /** Game identifier used for i18n and action log (e.g., "videopoker", "deuceswild") */
  gameName: 'videopoker' | 'deuceswild' | 'jokerpoker';
  /** i18n namespace (e.g., "videopoker", "deuceswild", "jokerpoker") */
  i18nNamespace: string;
  /** API exec function */
  apiExec: (
    command: 'reset' | 'bet' | 'hold' | 'log',
    amount?: number,
    indices?: number[],
  ) => Promise<VideoPokerResponse>;
  /** Route path for the game manual lookup (e.g., "/videopoker") */
  gamePath: string;
  /** CLI game configuration for CLI mode integration */
  cliGameConfig: Omit<CliGameConfig<VideoPokerResponse, Parameters<VideoPokerGameContentProps['apiExec']>>, 'gameName'>;
}

/** Payout table display component. Renders a bet-by-hand grid (columns = bet 1..5)
 * with the active bet column highlighted and, in the result phase, the winning hand
 * row highlighted. Expanded on the first visit so new players see the payouts; once
 * the player collapses it, the choice persists per variant. */
function PayoutTable({
  t,
  gameName,
  betAmount,
  winningRowKey,
  handCounts,
}: {
  t: (key: string) => string;
  gameName: 'videopoker' | 'deuceswild' | 'jokerpoker';
  betAmount: number;
  winningRowKey: string | null;
  /** Per-hand win counts to append as a trailing column, or null to hide the column. */
  handCounts: Record<string, number> | null;
}) {
  const [open, setOpen] = useLocalStorageToggle(`paytable_open_${gameName}`, true);
  const rows = videoPokerPayoutRows(gameName);
  const bets = Array.from({ length: VIDEO_POKER_MAX_BET }, (_, i) => i + 1);
  return (
    <details className="mb-3 text-center" open={open} onToggle={(e) => setOpen(e.currentTarget.open)}>
      <summary className="text-ds-warning text-sm cursor-pointer lg:text-base">{t('payoutTable.title')}</summary>
      <table
        className="mt-1 mx-auto border-collapse text-ds-text-muted text-xs lg:text-sm"
        aria-label={t('payoutTable.title')}
      >
        <thead>
          <tr>
            <th className="px-1.5 py-0.5 text-left font-medium">{t('payoutTable.hand')}</th>
            {bets.map((b) => (
              <th
                key={b}
                className={`px-1.5 py-0.5 text-right ${b === betAmount ? 'text-ds-warning font-bold' : ''}`}
                aria-current={b === betAmount ? 'true' : undefined}
              >
                {b}
              </th>
            ))}
            {handCounts && <th className="px-1.5 py-0.5 text-right font-medium">{t('payoutTable.count')}</th>}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => {
            const isWin = row.key === winningRowKey;
            return (
              <tr
                key={row.key}
                data-testid={`vp-payout-row-${row.key}`}
                className={isWin ? 'bg-ds-success/20 text-ds-text-primary font-bold' : ''}
                aria-current={isWin ? 'true' : undefined}
              >
                <td className="px-1.5 py-0.5 text-left whitespace-nowrap">{t(`payoutTable.name.${row.key}`)}</td>
                {bets.map((b) => (
                  <td key={b} className={`px-1.5 py-0.5 text-right ${b === betAmount ? 'text-ds-warning' : ''}`}>
                    {videoPokerPayoutCell(row, b)}
                  </td>
                ))}
                {handCounts && (
                  <td
                    className="px-1.5 py-0.5 text-right text-ds-text-primary tabular-nums"
                    data-testid={`vp-payout-count-${row.key}`}
                  >
                    {handCounts[row.key] ?? 0}
                  </td>
                )}
              </tr>
            );
          })}
        </tbody>
      </table>
    </details>
  );
}

/** Shared Video Poker game content used by all variants. */
export function VideoPokerGameContent({
  gameName,
  i18nNamespace,
  apiExec,
  gamePath,
  cliGameConfig,
}: VideoPokerGameContentProps) {
  const { t: tNs } = useTranslation(i18nNamespace);
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup(gameName);
  const { playSound } = useSound();

  const [betAmount, setBetAmount] = useState(1);
  const [heldCards, setHeldCards] = useState<boolean[]>([false, false, false, false, false]);
  // Screen-reader announcement for a hold toggle (aria-pressed alone is not
  // reliably re-announced by every AT when toggled via the keyboard).
  const [holdAnnounce, setHoldAnnounce] = useState('');
  // Screen-reader announcement summarising the draw result (winning hand +
  // payout, or the loss message). The payout text and the payout-table row
  // highlight are static, so this sr-only live region is the only channel
  // that tells a non-visual user the outcome. `resultNonce` is an invisible
  // counter that forces re-announcement even when two consecutive hands
  // produce identical text.
  const [resultAnnounce, setResultAnnounce] = useState('');
  const [resultNonce, setResultNonce] = useState(0);

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(apiExec);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode(gameName);
  const cliConfig: CliGameConfig<VideoPokerResponse, Parameters<typeof apiExec>> = useMemo(
    () => ({ gameName, ...cliGameConfig }),
    [gameName, cliGameConfig],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { hint, hintEnabled, setHintEnabled } = useGameHint(gameName, state);
  // Auto-hold: when the deal completes, pre-select the cards the hint engine
  // recommends so the player can hit Draw without re-discovering the optimal
  // hold themselves. Default ON to match real-machine behaviour; persisted
  // per variant so each game (videopoker / deuceswild / jokerpoker) keeps
  // its own toggle.
  const [autoHoldEnabled, setAutoHoldEnabled] = useLocalStorageToggle(`auto_hold_${gameName}`, true);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const isBetPhase = state?.phase === VideoPokerPhase.BET;
  const isDrawPhase = state?.phase === VideoPokerPhase.DRAW;
  const isResultPhase = state?.phase === VideoPokerPhase.RESULT;

  // Session statistics (hands / win rate / net) are scoped to the base
  // Jacks-or-Better variant so the shared component leaves Deuces Wild and
  // Joker Poker untouched. Storage is still keyed per variant, so enabling the
  // others later needs no data migration.
  const statsEnabled = gameName === 'videopoker';
  const { stats, clear: clearStats } = useVideoPokerStats(gameName, state, isResultPhase, statsEnabled);

  // Announce the draw outcome once per hand. Keying on the state reference (a
  // fresh object on every API result) re-fires the effect even for an identical
  // hand, so a repeated result is read aloud again.
  useEffect(() => {
    if (!isResultPhase || !state) {
      return;
    }
    const rowKey = videoPokerHandNameToRowKey(state.handName);
    const msg =
      state.payout > 0
        ? tNs('resultAnnounce.win', {
            handName: rowKey ? tNs(`payoutTable.name.${rowKey}`) : state.handName,
            payout: state.payout,
          })
        : tNs('resultAnnounce.lose');
    setResultAnnounce(msg);
    setResultNonce((n) => n + 1);
  }, [isResultPhase, state, tNs]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: re-apply auto-hold only on the phase transition into DRAW (a fresh hand). Adding autoHoldEnabled / state / gameName to the deps would overwrite the player's manual hold edits whenever any of those change mid-draw.
  useEffect(() => {
    if (!isDrawPhase) return;
    // Default: nothing held. When auto-hold is enabled, compute the
    // recommendation directly via getVideoPokerBaseHint so it works
    // regardless of whether the player has the hint UI toggled on.
    const next: boolean[] = [false, false, false, false, false];
    if (autoHoldEnabled && state) {
      const autoHint = getVideoPokerBaseHint(state, WILD_CARD_PREDICATE[gameName]);
      if (autoHint?.targetAction.startsWith('hold:')) {
        const csv = autoHint.targetAction.slice('hold:'.length);
        if (csv.length > 0) {
          for (const raw of csv.split(',')) {
            const idx = Number.parseInt(raw, 10);
            if (Number.isInteger(idx) && idx >= 0 && idx < next.length) {
              next[idx] = true;
            }
          }
        }
      }
    }
    setHeldCards(next);
    // Clear any stale hold announcement carried over from the previous hand.
    setHoldAnnounce('');
  }, [isDrawPhase]);

  const toggleHold = useCallback(
    (index: number) => {
      if (!isDrawPhase) return;
      const willHold = !heldCards[index];
      setHeldCards((prev) => {
        const next = [...prev];
        next[index] = !next[index];
        return next;
      });
      setHoldAnnounce(tNs(willHold ? 'a11y.holdOn' : 'a11y.holdOff', { index: index + 1 }));
    },
    [isDrawPhase, heldCards, tNs],
  );

  const handleDeal = useCallback(() => {
    execApi('bet', betAmount);
  }, [execApi, betAmount]);
  // Bet the table maximum (5) and deal in one action. Pass 5 explicitly rather
  // than relying on the async setBetAmount so the deal never uses a stale value.
  const handleBetMax = useCallback(() => {
    setBetAmount(5);
    execApi('bet', 5);
  }, [execApi]);

  const handleDraw = useCallback(() => {
    const indices = heldCards.reduce<number[]>((acc, held, i) => {
      if (held) acc.push(i);
      return acc;
    }, []);
    execApi('hold', undefined, indices);
  }, [execApi, heldCards]);

  const handleReset = useCallback(() => {
    execApi('reset');
  }, [execApi]);

  const phaseName = useMemo(() => {
    if (isBetPhase) return t('phase.bet');
    if (isDrawPhase) return t('phase.draw');
    return t('phase.result');
  }, [isBetPhase, isDrawPhase, t]);

  // Joker Poker only: evaluate the current 5 cards during the draw phase so the
  // player sees whether they already hold a paying hand (Kings or Better+). The
  // readout depends solely on the dealt hand, so toggling holds never changes
  // it, and it disappears once the phase leaves DRAW. `rowKey === null` means
  // the hand does not reach the pay minimum.
  const madeHand = useMemo(() => {
    if (gameName !== 'jokerpoker' || !isDrawPhase || !state || state.hand.length !== 5) return null;
    return evaluateJokerPokerMadeHand(state.hand);
  }, [gameName, isDrawPhase, state]);

  const actionBindings = useMemo(
    () => [
      { key: 'b', action: handleDeal, enabled: isBetPhase },
      { key: 'd', action: handleDraw, enabled: isDrawPhase },
      { key: 'r', action: handleReset, enabled: isResultPhase },
      // Number keys 1-5 toggle HOLD on the matching card (DRAW phase only),
      // mirroring the physical hold buttons on a real video-poker machine.
      ...[0, 1, 2, 3, 4].map((i) => ({
        key: String(i + 1),
        action: () => toggleHold(i),
        enabled: isDrawPhase,
      })),
    ],
    [handleDeal, handleDraw, handleReset, toggleHold, isBetPhase, isDrawPhase, isResultPhase],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  if (!state) return null;

  const displayHeld = isDrawPhase ? heldCards : (state.heldIndices ?? []);

  return (
    <GamePageShell
      title={tc(`nav.${gameName}`)}
      gameThemeBg={gameTheme[gameName].bg}
      phaseName={phaseName}
      isHumanTurn={isBetPhase || isDrawPhase}
      gamePath={gamePath}
      gameEndFlag={isResultPhase}
      winShow={isResultPhase && state.result === 1}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>{t('label.chips', { chips: state.chips })}</span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div
            className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint} ${isBetPhase ? 'flex flex-col' : ''}`}
          >
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="sr-only" role="status" aria-live="polite" data-testid="vp-hold-announce">
              {holdAnnounce}
            </div>

            {/* Keying the live region on resultNonce remounts it on each result, so
                assistive tech re-announces even two identical consecutive hands (an
                aria-hidden counter would be invisible to the accessibility tree and
                never trigger a re-read). data-nonce exposes the counter to tests. */}
            <div
              key={resultNonce}
              className="sr-only"
              role="status"
              aria-live="polite"
              data-testid="vp-result-announce"
              data-nonce={resultNonce}
            >
              {resultAnnounce}
            </div>

            {madeHand && (
              <div className="text-center mb-2" data-testid="vp-made-hand" aria-live="polite" aria-atomic="true">
                <span className="text-ds-text-muted text-xs mr-1">{tNs('madeHand.label')}:</span>
                {madeHand.rowKey ? (
                  <span className="text-ds-success text-sm font-bold">
                    {tNs(`payoutTable.name.${madeHand.rowKey}`)}
                  </span>
                ) : (
                  <span className="text-ds-text-muted text-sm">{tNs('madeHand.none')}</span>
                )}
              </div>
            )}

            {state.hand.length > 0 && (
              <div className="mb-4" data-tutorial="vp-hand">
                <div className="flex justify-center gap-2">
                  {state.hand.map((card, i) => {
                    const isWild = WILD_CARD_PREDICATE[gameName](card);
                    return (
                      <div key={`vp-${card.design}-${card.value}-${i}`} className="flex flex-col items-center">
                        <button
                          type="button"
                          onClick={() => toggleHold(i)}
                          disabled={!isDrawPhase}
                          className={`relative rounded transition-transform ${
                            displayHeld[i] ? 'ring-4 ring-ds-warning -translate-y-2 motion-safe:animate-card-lock' : ''
                          }`}
                          aria-label={`${displayHeld[i] ? `${tNs('hold')} ${i}` : tNs('card', { index: i })}${isWild ? ` ${tNs('wild')}` : ''}`}
                          aria-pressed={displayHeld[i] ?? false}
                          data-held={displayHeld[i] ? 'true' : undefined}
                        >
                          <AnimatedCard card={card} width={cardWidth} />
                          {displayHeld[i] && (
                            <span
                              aria-hidden="true"
                              className="absolute inset-x-0 bottom-1 mx-auto w-fit px-2 py-0.5 rounded bg-ds-warning text-ds-text-on-accent text-[10px] font-extrabold tracking-wider shadow-md pointer-events-none"
                              data-testid={`vp-hold-badge-${i}`}
                            >
                              {tNs('hold')}
                            </span>
                          )}
                          {isWild && (
                            <span
                              aria-hidden="true"
                              className="absolute top-1 right-1 px-1.5 py-0.5 rounded bg-ds-info text-ds-text-on-accent text-[9px] font-extrabold tracking-wider shadow-md pointer-events-none"
                              data-testid={`vp-wild-badge-${i}`}
                            >
                              {tNs('wild')}
                            </span>
                          )}
                        </button>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {isResultPhase && state.payout > 0 && (
              <div className="text-ds-text-primary text-center font-bold mb-2">
                {t('label.payout', { payout: state.payout })}
              </div>
            )}

            {isBetPhase && (
              <div className="flex-1 flex flex-col items-center justify-center" data-tutorial="vp-bet-controls">
                <ErrorAlert message={error} onRetry={retry} />
                <div className="flex items-center gap-2">
                  <label htmlFor="vp-bet-amount" className="text-ds-text-primary text-sm">
                    {t('label.betAmount')}
                  </label>
                  <select
                    id="vp-bet-amount"
                    value={betAmount}
                    onChange={(e) => setBetAmount(Number(e.target.value))}
                    className="px-2 py-1 rounded text-sm"
                  >
                    {[1, 2, 3, 4, 5].map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="mt-2 flex gap-2">
                  <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
                    {t('button.deal')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleBetMax} disabled={loading}>
                    {t('label.betMax')}
                  </button>
                </div>
                <p className="text-ds-text-muted text-xs mt-2">{tNs('dealGuide')}</p>
              </div>
            )}

            <PayoutTable
              t={tNs}
              gameName={gameName}
              betAmount={betAmount}
              winningRowKey={isResultPhase ? videoPokerHandNameToRowKey(state.handName) : null}
              handCounts={statsEnabled ? stats.handCounts : null}
            />

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme[gameName].footer} px-4 pt-3`}>
            {!isBetPhase && <ErrorAlert message={error} onRetry={retry} />}
            {hintEnabled && hint && <HintTooltip reason={tNs(hint.reason)} confidence={hint.confidence} />}
            {isDrawPhase && (
              <div className="flex flex-col items-center gap-1 pb-2" data-tutorial="vp-draw-button">
                <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                  {t('button.draw')}
                </button>
                <p className="text-ds-text-muted text-xs">{tNs('a11y.kbdHint')}</p>
              </div>
            )}
            {isResultPhase && (
              <div className="flex justify-center gap-2 pb-2">
                <GameResetButton
                  isGameEnd={isResultPhase}
                  onReset={handleReset}
                  requestConfirm={requestConfirm}
                  loading={loading}
                  dataTutorial="vp-reset-button"
                />
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('actionLog.view')}
                </button>
              </div>
            )}
            {statsEnabled && (
              <div
                className="flex flex-wrap items-center justify-center gap-x-3 gap-y-1 pb-2 text-ds-text-muted text-xs lg:text-sm"
                data-testid="vp-session-stats"
              >
                <span data-testid="vp-stats-summary">
                  {stats.hands === 0
                    ? tNs('stats.empty')
                    : tNs('stats.summary', {
                        hands: stats.hands,
                        winRate: Math.round(videoPokerWinRate(stats) * 100),
                        net: `${videoPokerNet(stats) >= 0 ? '+' : ''}${videoPokerNet(stats)}`,
                      })}
                </span>
                {stats.hands > 0 && (
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={clearStats}
                    disabled={loading}
                    data-testid="vp-stats-clear"
                  >
                    {tNs('stats.clear')}
                  </button>
                )}
              </div>
            )}
            <div className="flex flex-wrap justify-center gap-x-4 pb-2">
              <label className="text-ds-text-primary text-sm flex items-center gap-2 min-h-[44px]">
                <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <label className="text-ds-text-primary text-sm flex items-center gap-2 min-h-[44px]">
                <input
                  type="checkbox"
                  checked={autoHoldEnabled}
                  onChange={(e) => setAutoHoldEnabled(e.target.checked)}
                  data-testid="vp-auto-hold-toggle"
                />
                {t('label.autoHold')}
              </label>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
