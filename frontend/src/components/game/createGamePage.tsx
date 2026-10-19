import type { TFunction } from 'i18next';
import { type ComponentType, useCallback, useMemo } from 'react';
import type { actionLogApi } from '../../api/gameApi';
import type { ActionBinding } from '../../hooks/useActionKeyboardNav';
import { useActionKeyboardNav } from '../../hooks/useActionKeyboardNav';
import { useCliGame } from '../../hooks/useCliGame';
import { useCliMode } from '../../hooks/useCliMode';
import { type HintGameName, type UseGameHintReturn, useGameHint } from '../../hooks/useGameHint';
import { useGamePageSetup } from '../../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../../hooks/useGiveUpConfirm';
import { type GameKey, gameTheme } from '../../styles/gameTheme';
import type { TutorialStep } from '../../types/tutorial';
import type { CliGameConfig } from '../../utils/cli/types';
import { hintCheckboxItem } from '../../utils/settingsItems';
import { ActionLogSection } from '../ActionLogSection';
import { ActionShortcutsPanel } from '../ActionShortcutsPanel';
import { CliTerminal } from '../cli/CliTerminal';
import { CliToggle } from '../cli/CliToggle';
import { type SettingsGroup, SettingsPanel } from '../common/SettingsPanel';
import type { ErrorAlertProps } from '../ErrorAlert';
import { ErrorAlert } from '../ErrorAlert';
import { GameFooter } from '../GameFooter';
import { GamePageShell } from '../GamePageShell';
import { GameResetButton } from '../GameResetButton';
import { LandscapeBanner } from '../LandscapeBanner';
import { GameSkeleton, type GameSkeletonProps } from '../skeleton/GameSkeleton';
import { withTutorial } from '../tutorial/withTutorial';

/** Common state and actions returned by a game hook. */
export interface GameHookBase<S> {
  state: S | null;
  loading: boolean;
  error: string | null;
  hintError?: string | null;
  retry: ErrorAlertProps['onRetry'];
  handleReset: () => void;
  handleGiveUp?: () => void;
  exec: (...args: never[]) => Promise<unknown>;
}

/** Values shared by the game-specific slots rendered by {@link createGamePage}. */
export interface GamePageContext<S, G extends GameHookBase<S>> {
  game: G;
  state: S;
  t: TFunction;
  tc: TFunction;
  busy: boolean;
  requestConfirm: (action: () => void) => void;
  confirmGiveUpAction: () => void;
  frontendHint: UseGameHintReturn;
  showActionLog: () => void;
  hideActionLog: () => void;
}

type GamePageKey = keyof typeof actionLogApi & HintGameName & keyof typeof gameTheme & GameKey;

/** Definition for the game-specific pieces and CLI configuration of a page. */
export interface GameDefinition<S, G extends GameHookBase<S>, A extends unknown[]> {
  key: GamePageKey;
  useGame: () => G;
  tutorial: TutorialStep[];
  skeleton: GameSkeletonProps['layout'];
  cli: Omit<CliGameConfig<S, A>, 'gameName'>;
  busy?: (game: G) => boolean;
  phase: (ctx: GamePageContext<S, G>) => { name: string; ended: boolean; won?: boolean };
  shortcuts?: (ctx: GamePageContext<S, G>) => ActionBinding[];
  shortcutsEnabled?: (ctx: GamePageContext<S, G>) => boolean;
  shortcutsTestId?: string;
  settings?: (ctx: GamePageContext<S, G>) => SettingsGroup[];
  resetTutorialId?: string;
  HeaderExtra?: ComponentType<GamePageContext<S, G>>;
  LiveRegion?: ComponentType<GamePageContext<S, G>>;
  Board: ComponentType<GamePageContext<S, G>>;
  FooterActions?: ComponentType<GamePageContext<S, G>>;
}

/** Builds a game page from shared wiring and a game-specific definition. */
export function createGamePage<S, G extends GameHookBase<S>, A extends unknown[]>(
  def: GameDefinition<S, G, A>,
): ComponentType {
  function Content() {
    const setup = useGamePageSetup(def.key);
    const game = def.useGame();
    const state = game.state;
    const frontendHint = useGameHint(def.key, state);
    const cliMode = useCliMode(def.key);
    const cliConfig = useMemo(() => ({ gameName: def.key, ...def.cli }), []);
    const { handleCommand } = useCliGame(game.exec as unknown as (...args: A) => Promise<void>, cliConfig, state, {
      addInput: cliMode.addInput,
      addOutput: cliMode.addOutput,
      addError: cliMode.addError,
      clearLog: cliMode.clearLog,
    });
    const confirmGiveUpAction = useGiveUpConfirm(game.handleGiveUp ?? (() => undefined), setup.requestGiveUpConfirm);
    const busy = def.busy ? def.busy(game) : game.loading;
    const context =
      state === null
        ? null
        : ({
            game,
            state,
            t: setup.t,
            tc: setup.tc,
            busy,
            requestConfirm: setup.requestConfirm,
            confirmGiveUpAction,
            frontendHint,
            showActionLog: setup.showActionLog,
            hideActionLog: setup.hideActionLog,
          } satisfies GamePageContext<S, G>);
    const shortcuts = context ? (def.shortcuts?.(context) ?? []) : [];
    const shortcutsEnabled = context ? (def.shortcutsEnabled?.(context) ?? true) : false;
    useActionKeyboardNav({ bindings: shortcuts, enabled: shortcutsEnabled && !busy });
    const handleManualReset = useCallback(() => {
      setup.hideActionLog();
      game.handleReset();
    }, [game.handleReset, setup.hideActionLog]);

    if (!state || !context) return <GameSkeleton gameKey={def.key} layout={def.skeleton} />;

    const phase = def.phase(context);
    const settings = def.settings?.(context) ?? [
      {
        items: [hintCheckboxItem(setup.tc, frontendHint.hintEnabled, frontendHint.setHintEnabled)],
      },
    ];
    const HeaderExtra = def.HeaderExtra;
    const LiveRegion = def.LiveRegion;
    const Board = def.Board;
    const FooterActions = def.FooterActions;
    return (
      <GamePageShell
        title={setup.tc(`nav.${def.key}`)}
        gameThemeBg={gameTheme[def.key].bg}
        phaseName={phase.name}
        gamePath={`/${def.key}`}
        gameEndFlag={phase.ended}
        winShow={phase.won}
        loading={game.loading}
        confirmOpen={setup.confirmOpen}
        confirmReset={setup.confirmReset}
        cancelReset={setup.cancelReset}
        giveUpConfirmOpen={setup.giveUpConfirmOpen}
        confirmGiveUp={setup.confirmGiveUp}
        cancelGiveUp={setup.cancelGiveUp}
        headerExtra={
          <>
            {HeaderExtra && <HeaderExtra {...context} />}
            <CliToggle cliEnabled={cliMode.cliEnabled} onToggle={cliMode.toggleCli} />
          </>
        }
      >
        {LiveRegion && <LiveRegion {...context} />}
        {cliMode.cliEnabled ? (
          <CliTerminal logEntries={cliMode.logEntries} onCommand={handleCommand} disabled={game.loading} />
        ) : (
          <>
            <LandscapeBanner message={setup.t('landscapeBanner')} />
            <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
              <Board {...context} />
              <ActionLogSection
                isEndPhase={phase.ended}
                actionLog={setup.actionLog}
                showActionLog={setup.showActionLog}
                hideActionLog={setup.hideActionLog}
              />
            </div>
            <SettingsPanel title={setup.tc('settings.title')} groups={settings} />
            <GameFooter className={`${gameTheme[def.key].footer} px-4 py-2.5`}>
              <ErrorAlert message={game.error ?? game.hintError ?? null} onRetry={game.retry} />
              <div className="flex gap-2 items-center flex-wrap">
                {FooterActions && <FooterActions {...context} />}
                <GameResetButton
                  isGameEnd={phase.ended}
                  onReset={handleManualReset}
                  requestConfirm={setup.requestConfirm}
                  loading={busy}
                  dataTutorial={def.resetTutorialId}
                />
              </div>
              <ActionShortcutsPanel bindings={shortcuts} data-testid={def.shortcutsTestId} />
            </GameFooter>
          </>
        )}
      </GamePageShell>
    );
  }
  return withTutorial(Content, def.key, def.tutorial);
}
