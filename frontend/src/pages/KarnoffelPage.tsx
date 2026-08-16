import { useEffect, useMemo, useState } from 'react';
import { karnoffelApi } from '../api/gameApi';
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
import { btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KarnoffelResponse } from '../types/card';
import { KarnoffelPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { KARNOFFEL_HELP, parseKarnoffelCommand } from '../utils/cli/commands/karnoffelCommands';
import { formatKarnoffelState } from '../utils/cli/formatters/karnoffelFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { karnoffelRankKey } from '../utils/karnoffelRanks';

/** Karnöffel tutorial step definitions. */
const KARNOFFEL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="karnoffel-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="karnoffel-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KARNOFFEL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KarnoffelPhase.PLAY]: 'play',
  [KarnoffelPhase.HAND_END]: 'handEnd',
  [KarnoffelPhase.GAME_END]: 'gameEnd',
};

/** Renders the Karnöffel game page: the oldest named card game, with its irregular ranking. */
export const KarnoffelPage = withTutorial(KarnoffelPageContent, 'karnoffel', KARNOFFEL_TUTORIAL_STEPS);

/** Inner content of the Karnöffel page, wrapped by TutorialProvider. */
function KarnoffelPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('karnoffel');
  const { state, loading, error, exec, retry } = useGameApi(karnoffelApi.exec);

  const [selected, setSelected] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('karnoffel');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('karnoffel', state);
  const cliConfig: CliGameConfig<KarnoffelResponse, Parameters<typeof karnoffelApi.exec>> = useMemo(
    () => ({
      gameName: 'karnoffel',
      parseCommand: parseKarnoffelCommand,
      formatResponse: formatKarnoffelState,
      helpText: KARNOFFEL_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('karnoffel', KARNOFFEL_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="karnoffel" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const human = state.players.find((p) => p.isHuman);
  // Hoisted so a seat-less state exercises the empty case rather than leaving it
  // as an unreachable fallback inside the hand render.
  const humanCards = human?.cards ?? [];
  const isPlay = state.phase === KarnoffelPhase.PLAY;
  const isHandEnd = state.phase === KarnoffelPhase.HAND_END;
  const isGameEnd = state.phase === KarnoffelPhase.GAME_END || state.gameEndFlag;
  // 人間は席 0 = チーム 0。勝敗はチームで判定する。
  const humanWon = isGameEnd && state.winnerTeam === 0;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitLabel = (s: number): string => t(`suitName.${s}`);

  const canPlay = (i: number) => state.validPlays.includes(i);

  const handlePlay = () => {
    if (selected === null) return;
    exec('play', { cardIndex: selected });
    setSelected(null);
  };

  const handleManualReset = () => {
    hideActionLog();
    setSelected(null);
    exec('reset');
  };

  const result = state.lastResult;

  return (
    <GamePageShell
      title={tc('nav.karnoffel')}
      gameThemeBg={gameTheme.karnoffel.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanPlay}
      gamePath="/karnoffel"
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
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="karnoffel-info">
              <span className="mr-4">{t('hand', { n: state.handNumber })}</span>
              <span className="mr-4" data-testid="karnoffel-chosen">
                {t('chosenSuit')}: {suitLabel(state.chosenSuit)}
              </span>
            </div>

            {/* The lowest face-up card decides the suit — not a turn-up. */}
            <div className="mb-2 text-center text-ds-text-muted text-xs" data-testid="karnoffel-chosen-note">
              {t('chosenNote')}
            </div>

            {/* The irregular ranking, which is the whole point of the game. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="karnoffel-ladder">
              <div className="mb-1 text-ds-text-primary">{t('ladderTitle')}</div>
              <div className="text-ds-text-primary">{t('ladderLine')}</div>
              <div className="mt-1 text-ds-text-muted">{t('ladderNote')}</div>
            </div>

            {/* Score sheet */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="karnoffel-scores">
              <div>
                {t('team', { n: 0 })}: {t('handsWon', { n: state.handsWon[0] })} (
                {t('tricksWon', { n: state.teamTricks[0] })}) / {t('team', { n: 1 })}:{' '}
                {t('handsWon', { n: state.handsWon[1] })} ({t('tricksWon', { n: state.teamTricks[1] })})
              </div>
              <div className="text-xs text-ds-text-muted">
                {t('targetNote', { n: state.targetHands, need: state.tricksToWin })}
              </div>
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="karnoffel-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="karnoffel-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>({t('team', { n: p.team })})</span>
                  {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  {/* **表向きの札は全員ぶん見える。**切札の根拠がここにある。 */}
                  <span className="flex items-center gap-1">
                    {t('upCard')}:{p.upCard ? <CardImage card={p.upCard} width={cardWidth} /> : <span>-</span>}
                  </span>
                  <span>{t('tricksWon', { n: p.tricksWon })}</span>
                  {!p.isHuman && p.cards.length === 0 && <span>{t('hiddenHand', { count: p.cardCount })}</span>}
                </div>
              ))}
            </div>

            {/* Trick */}
            {state.trick.length > 0 && (
              <div className="mb-2 flex items-center gap-2" data-testid="karnoffel-trick">
                <span className="text-ds-text-muted text-sm">{t('trick')}</span>
                {state.trick.map((c, i) => (
                  <CardImage key={`trick-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
              </div>
            )}

            {/* Hand result */}
            {(isHandEnd || isGameEnd) && result && (
              <div
                className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
                data-testid="karnoffel-result"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('resultTitle')}</div>
                <div>
                  {result.winnerTeam >= 0
                    ? t('handWonLine', { team: result.winnerTeam, t0: result.tricks[0], t1: result.tricks[1] })
                    : t('handDrawnLine', { t0: result.tricks[0], t1: result.tricks[1] })}
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

          {/* Footer */}
          <GameFooter className={`${gameTheme.karnoffel.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="karnoffel-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {humanCards.map((c, i) => {
                  // The ladder text names the titled cards, but which card in hand
                  // holds a title depends on the suit chosen this deal (#4773).
                  const rankKey = karnoffelRankKey(c, state.chosenSuit);
                  return (
                    <button
                      key={`hand-${c.design}-${c.value}-${i}`}
                      type="button"
                      data-hint-action="play"
                      onClick={() => setSelected(i)}
                      disabled={loading || (isPlay && !canPlay(i))}
                      aria-label={rankKey ? `${cardAlt(c)} (${t(`rankBadge.${rankKey}`)})` : cardAlt(c)}
                      className={`relative rounded ${selected === i ? 'ring-2 ring-ds-accent' : ''} ${
                        isPlay && !canPlay(i) ? 'opacity-40' : ''
                      }`}
                    >
                      <CardImage card={c} width={cardWidth} />
                      {rankKey && (
                        <span
                          aria-hidden="true"
                          data-testid={`karnoffel-rank-${rankKey}`}
                          className={`absolute left-0 right-0 bottom-0 rounded-b px-0.5 text-[9px] font-bold text-center truncate ${badgeWarningColors}`}
                        >
                          {t(`rankBadge.${rankKey}`)}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanPlay && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="karnoffel-play-notice">
                {t('playNotice')}
              </div>
            )}

            <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="karnoffel-actions">
              {isHumanPlay && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={loading || selected === null}
                >
                  {t('playButton')}
                </button>
              )}

              {isHandEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextHand')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose')}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="karnoffel-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
