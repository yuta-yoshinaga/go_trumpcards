import { useCallback, useEffect, useMemo, useState } from 'react';
import { pasurApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PasurResponse } from '../types/card';
import { PasurPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PASUR_HELP, parsePasurCommand } from '../utils/cli/commands/pasurCommands';
import { formatPasurState } from '../utils/cli/formatters/pasurFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guided tutorial steps (the 11 rule, face cards, soor, your hand). */
const PASUR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ps-rule"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ps-table"]', messageKey: 'tutorial.face', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ps-seats"]', messageKey: 'tutorial.soor', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ps-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
];

/**
 * Inner content for the Pasur page (wrapped by `withTutorial`).
 *
 * Iran's fishing game. Playing a card is two decisions — which card, and which
 * table cards it takes — so the page asks them in that order and **offers only
 * the combinations the server sent in `captureOptions`**. Rebuilding "which
 * subsets add to 11" here would drift from the domain.
 *
 * Capturing is compulsory, so the lay-down button disappears for any card that
 * has an option: the server would reject it.
 */
function PasurPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pasur');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<PasurResponse, Parameters<typeof pasurApi.exec>>(pasurApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('pasur', state);
  const [picked, setPicked] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pasur');
  const cliConfig: CliGameConfig<PasurResponse, Parameters<typeof pasurApi.exec>> = useMemo(
    () => ({
      gameName: 'pasur',
      parseCommand: parsePasurCommand,
      formatResponse: formatPasurState,
      helpText: PASUR_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    setPicked(null);
    void dispatch('reset');
  }, [dispatch, hideActionLog]);

  const handlePlay = useCallback(
    (cardIndex: number, table: number[]) => {
      setPicked(null);
      void dispatch('play', cardIndex, undefined, table);
    },
    [dispatch],
  );

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="pasur" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />;
  }

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === PasurPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentPlayerIdx]?.isHuman === true;

  const pickedOptions = picked === null ? [] : (state.captureOptions[picked] ?? []);
  // **取れる組み合わせがあるときは場に置けない。** サーバが必ず拒否する。
  const canTrail = picked !== null && pickedOptions.length === 0;

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winners.length === 1) {
      return state.winners[0] === 0
        ? t('result.you')
        : t('result.cpu', { name: t('header.cpu', { idx: String(state.winners[0]) }) });
    }
    return t('result.tie', { n: String(state.winners.length) });
  })();

  return (
    <GamePageShell
      title={tc('nav.pasur')}
      gameThemeBg={gameTheme.pasur.bg}
      phaseName={isGameEnd ? t('phase.gameEnd') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/pasur"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winners.length === 1 && state.winners[0] === 0}
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
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4" data-testid="ps-pack">
                {t('header.pack', { pack: String(state.packsDealt) })}
              </span>
              <span data-testid="ps-deck">{t('header.deck', { n: String(state.deckRemaining) })}</span>
            </div>

            {/* **11 の合計と絵札の扱いが規則そのもの。** 先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="ps-rule"
              data-tutorial="ps-rule"
            >
              {t('header.rule')}
            </div>

            <div className="mb-4" data-testid="ps-table" data-tutorial="ps-table">
              <div className="text-ds-text-muted text-sm mb-1">
                {state.table.length > 0 ? t('header.table') : t('header.tableEmpty')}
              </div>
              <div className="flex flex-wrap gap-2 justify-center">
                {state.table.map((card, idx) => (
                  <CardImage key={`table-${card.design}-${card.value}-${idx}`} card={card} width={cardWidth} />
                ))}
              </div>
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="ps-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`ps-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.id === state.lastCaptureIdx && (
                    <span className="ml-1 text-ds-accent">{t('header.lastCapture')}</span>
                  )}
                  {': '}
                  {t('header.captured', { n: String(p.capturedCount) })}
                  {' / '}
                  <span className="text-ds-accent">{t('header.soors', { n: String(p.soors) })}</span>
                  {' / '}
                  {t('header.score', { n: String(p.score) })}
                </div>
              ))}
            </div>

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="ps-result"
              >
                {resultBanner}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <ErrorAlert message={error} onRetry={retry} />

            {human && human.cards.length > 0 && (
              <div className="mt-4" data-tutorial="ps-hand">
                <div className="text-ds-text-muted text-sm mb-1">
                  {t('header.you')}: {human.cardCount}
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((card, idx) => (
                    <button
                      key={`${card.design}-${card.value}-${idx}`}
                      type="button"
                      onClick={() => setPicked(picked === idx ? null : idx)}
                      disabled={loading || !isHumanTurn}
                      aria-pressed={picked === idx}
                      aria-label={t('actions.selectAria', { card: cardAlt(card) })}
                      className={`disabled:opacity-50 ${
                        picked === idx
                          ? 'rounded-lg ring-2 ring-ds-warning'
                          : (state.captureOptions[idx]?.length ?? 0) > 0
                            ? 'rounded-lg ring-2 ring-ds-success'
                            : ''
                      }`}
                    >
                      <CardImage card={card} width={cardWidth} />
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* **候補はサーバが送ったものだけ。** ここで 11 の部分集合を作り直さない。 */}
            {picked !== null && isHumanTurn && (
              <div className="mt-3" data-testid="ps-options">
                <div className="text-ds-text-muted text-sm mb-1">{t('actions.pick')}</div>
                <div className="flex flex-wrap gap-2">
                  {pickedOptions.map((option) => {
                    // **スールは「取った結果、場が空になる」こと** (domain の
                    // takeCards と同じ条件)。倍化を狙うなら、どの選択肢がそれに
                    // 当たるかがボタンから読めないと選べない (#5762)。
                    const isSoor = option.length === state.table.length && state.table.length > 0;
                    return (
                      <button
                        key={`opt-${option.join('-')}`}
                        type="button"
                        className={isSoor ? `${btnSuccess} ring-2 ring-ds-success` : btnWarning}
                        onClick={() => handlePlay(picked, option)}
                        disabled={loading}
                        data-testid={`ps-take-${option.join('-')}-btn`}
                      >
                        {t('actions.take', {
                          cards: option.map((i) => cardAlt(state.table[i])).join(', '),
                        })}
                        {isSoor && (
                          <span className="ml-1 font-bold" data-testid={`ps-soor-${option.join('-')}`}>
                            <span aria-hidden="true">[{t('actions.takeSoor')}]</span>
                            <span className="sr-only">{t('actions.takeSoorAria')}</span>
                          </span>
                        )}
                      </button>
                    );
                  })}
                  {canTrail ? (
                    <button
                      type="button"
                      className={btnWarning}
                      onClick={() => handlePlay(picked, [])}
                      disabled={loading}
                      data-testid="ps-trail-btn"
                    >
                      {t('actions.trail')}
                    </button>
                  ) : (
                    <span className="text-ds-text-muted text-sm self-center" data-testid="ps-must-capture">
                      {t('actions.mustCapture')}
                    </span>
                  )}
                  <button type="button" className={btnPrimary} onClick={() => setPicked(null)} disabled={loading}>
                    {t('actions.cancel')}
                  </button>
                </div>
              </div>
            )}

            {hintEnabled && hint && (
              <div className="mt-3 flex justify-center">
                <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />
              </div>
            )}

            <div className="mt-4 flex flex-wrap gap-2" data-tutorial="ps-actions">
              <button
                type="button"
                className={btnPrimary}
                onClick={() => requestConfirm(handleReset)}
                disabled={loading}
              >
                {t('actions.reset')}
              </button>
              {!isGameEnd && (
                <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                  {t('actions.giveUp')}
                </button>
              )}
            </div>

            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, hintEnabled, setHintEnabled)] }]}
            />
          </div>

          <ActionLogSection
            isEndPhase={isGameEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />
        </>
      )}
    </GamePageShell>
  );
}

/** Pasur page wrapped with TutorialProvider. */
export const PasurPage = withTutorial(PasurPageContent, 'pasur', PASUR_TUTORIAL_STEPS);
