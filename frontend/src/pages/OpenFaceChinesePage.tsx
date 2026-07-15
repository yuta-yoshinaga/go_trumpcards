import { useEffect, useMemo } from 'react';
import { openfacechineseApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, OpenFaceChinesePlayer, OpenFaceChineseResponse } from '../types/card';
import { OpenFaceChinesePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { OPENFACECHINESE_HELP, parseOpenfacechineseCommand } from '../utils/cli/commands/openfacechineseCommands';
import { formatOpenfacechineseState } from '../utils/cli/formatters/openfacechineseFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Row indices accepted by the backend `place` command. */
const ROW_FRONT = 0;
const ROW_MIDDLE = 1;
const ROW_BACK = 2;

/** Open Face Chinese Poker (OFC) tutorial step definitions. */
const OFC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ofc-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ofc-rows"]',
    messageKey: 'tutorial.rows',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ofc-place"]',
    messageKey: 'tutorial.place',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ofc-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric OFC phases to i18n phase-label keys. */
const OFC_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OpenFaceChinesePhase.PLACING]: 'placing',
  [OpenFaceChinesePhase.ROUND_END]: 'roundEnd',
  [OpenFaceChinesePhase.GAME_END]: 'gameEnd',
};

/** Renders the Open Face Chinese Poker (OFC) game page. */
export const OpenFaceChinesePage = withTutorial(OpenFaceChinesePageContent, 'openfacechinese', OFC_TUTORIAL_STEPS);

/** Inner content of the OFC page, wrapped by TutorialProvider. */
function OpenFaceChinesePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('openfacechinese');
  const { state, loading, error, exec, retry } = useGameApi(openfacechineseApi.exec);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('openfacechinese');
  const cliConfig: CliGameConfig<OpenFaceChineseResponse, Parameters<typeof openfacechineseApi.exec>> = useMemo(
    () => ({
      gameName: 'openfacechinese',
      parseCommand: parseOpenfacechineseCommand,
      formatResponse: formatOpenfacechineseState,
      helpText: OPENFACECHINESE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const phaseNames = usePhaseNames('openfacechinese', OFC_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();

  if (!state)
    return (
      <GameSkeleton gameKey="openfacechinese" layout={{ kind: 'casino-table', sections: [2], footerStyle: 'bet' }} />
    );

  const isGameEnd = state.phase === OpenFaceChinesePhase.GAME_END || state.gameEndFlag;
  const isPlacing = state.phase === OpenFaceChinesePhase.PLACING && !isGameEnd;
  const isRoundEnd = state.phase === OpenFaceChinesePhase.ROUND_END && !isGameEnd;
  const phaseName = phaseNames[state.phase] ?? '';

  const human = state.players.find((p) => p.isHuman) ?? null;
  const humanWon = isGameEnd && human !== null && state.winnerIdx === human.id;
  const canPlace = isPlacing && state.isHumanTurn && Boolean(state.currentCard);

  const handleReset = () => {
    hideActionLog();
    exec('reset');
  };

  const handlePlace = (row: number) => exec('place', { row });
  const handleNext = () => exec('nextround');

  const playerName = (player: OpenFaceChinesePlayer) => (player.isHuman ? t('you') : t('cpu', { n: player.id }));

  /** Renders a single row of a player's board, padding empty slots up to `capacity`. */
  const renderRow = (label: string, cards: Card[], capacity: number, keyPrefix: string) => {
    // A fieldset+legend names the whole row for SR (count + card contents), since
    // the empty slots are decorative and biome forbids role="group" on a div.
    const cardNames = cards.map(cardAlt).join(', ');
    const rowAria = cardNames
      ? t('rowAriaFilled', { label, count: cards.length, cards: cardNames })
      : t('rowAriaEmpty', { label, count: cards.length });
    return (
      <fieldset className="flex flex-col gap-1 border-0 p-0 m-0" data-testid={`ofc-row-${keyPrefix}`}>
        <legend className="sr-only">{rowAria}</legend>
        <span className="text-ds-text-muted text-[11px]" aria-hidden="true">
          {label}
        </span>
        <div className="flex gap-1 flex-wrap">
          {cards.map((c, i) => (
            <CardImage key={`${keyPrefix}-${i}`} card={c} width={Math.round(cardWidth * 0.62)} />
          ))}
          {Array.from({ length: Math.max(0, capacity - cards.length) }).map((_, i) => (
            <div
              key={`${keyPrefix}-empty-${i}`}
              className="rounded border border-dashed border-white/20 bg-black/20"
              style={{ width: Math.round(cardWidth * 0.62), height: Math.round(cardWidth * 0.62 * 1.4) }}
              aria-hidden="true"
            />
          ))}
        </div>
      </fieldset>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.openfacechinese')}
      gameThemeBg={gameTheme.openfacechinese.bg}
      phaseName={phaseName}
      gamePath="/openfacechinese"
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
            {/* Round info */}
            <div
              className="mb-3 p-2 rounded bg-black/30 flex flex-wrap justify-center gap-x-6 gap-y-1 text-sm"
              data-tutorial="ofc-info"
            >
              <span className="font-semibold text-ds-warning">{t('round', { round: state.roundNumber })}</span>
              {human && <span className="text-ds-text-muted">{t('totalScore', { score: human.totalScore })}</span>}
            </div>

            {/* Announce the pending card by name whenever it changes (persistent region). */}
            <span className="sr-only" role="status" aria-live="polite" data-testid="ofc-pending-announce">
              {canPlace && state.currentCard ? t('pendingAnnounce', { card: cardAlt(state.currentCard) }) : ''}
            </span>

            {/* Pending card to place */}
            {canPlace && state.currentCard && (
              <div
                className="mb-3 p-3 rounded bg-black/30 text-center flex flex-col items-center gap-2"
                data-tutorial="ofc-place"
              >
                <span className="text-ds-text-muted text-xs">{t('pendingTitle')}</span>
                <CardImage card={state.currentCard} width={cardWidth} />
                <span className="text-ds-text-primary text-sm">{t('placePrompt')}</span>
                <div className="flex flex-wrap justify-center gap-2 mt-1">
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => handlePlace(ROW_FRONT)}
                    disabled={loading}
                    data-testid="place-front"
                  >
                    {t('place.front')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => handlePlace(ROW_MIDDLE)}
                    disabled={loading}
                    data-testid="place-middle"
                  >
                    {t('place.middle')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => handlePlace(ROW_BACK)}
                    disabled={loading}
                    data-testid="place-back"
                  >
                    {t('place.back')}
                  </button>
                </div>
              </div>
            )}

            {/* Player boards */}
            <div className="grid gap-3 sm:grid-cols-2" data-tutorial="ofc-rows">
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  className={`p-3 rounded bg-black/20 ${p.id === state.currentPlayerIdx && isPlacing ? 'ring-1 ring-ds-warning' : ''}`}
                  data-testid={`player-${p.id}`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-ds-text-primary text-sm font-semibold">{playerName(p)}</span>
                    <div className="flex items-center gap-2 text-xs">
                      {p.fouled && (
                        <span className="text-ds-error font-semibold" role="status">
                          {t('fouled')}
                        </span>
                      )}
                      {p.fantasyland && (
                        <span className="text-ds-accent font-semibold" role="status">
                          {t('fantasyland')}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex flex-col gap-2">
                    {renderRow(t('rows.front'), p.front, 3, `front-${p.id}`)}
                    {renderRow(t('rows.middle'), p.middle, 5, `middle-${p.id}`)}
                    {renderRow(t('rows.back'), p.back, 5, `back-${p.id}`)}
                  </div>
                  {isRoundEnd && (
                    <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs">
                      <span className="text-ds-text-primary" data-testid={`round-score-${p.id}`}>
                        {t('roundScore', { score: p.roundScore })}
                      </span>
                      {p.royalty > 0 && <span className="text-ds-success">{t('royalty', { points: p.royalty })}</span>}
                      <span className="text-ds-text-muted">{t('totalScore', { score: p.totalScore })}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>

            {isRoundEnd && (
              <div className="mt-3 text-center text-ds-text-primary text-sm font-semibold">{t('roundEndTitle')}</div>
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

          {/* Footer: next round / reset */}
          <GameFooter className={`${gameTheme.openfacechinese.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex flex-wrap gap-2 items-center">
              {isGameEnd && <span className="text-ds-text-primary text-sm font-semibold mr-1">{t('gameEnd')}</span>}

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
                dataTutorial="ofc-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
