import type { TFunction } from 'i18next';
import { type ComponentType, useCallback } from 'react';
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
import { GamePageShell, type GamePageShellProps, type GiveUpConfirmProps } from '../GamePageShell';
import { GameResetButton } from '../GameResetButton';
import { LandscapeBanner } from '../LandscapeBanner';
import { GameSkeleton, type GameSkeletonProps } from '../skeleton/GameSkeleton';
import { withTutorial } from '../tutorial/withTutorial';

/** Common state and actions returned by a game hook. */
export interface GameHookBase<S, A extends unknown[] = any[]> {
  state: S | null;
  loading: boolean;
  error: string | null;
  hintError?: string | null;
  retry: ErrorAlertProps['onRetry'];
  handleReset: () => void;
  handleGiveUp?: () => void;
  exec: (...args: A) => Promise<void>;
}

/** Values shared by the game-specific slots rendered by {@link createGamePage}. */
export interface GamePageContext<S, G extends GameHookBase<S, any>, V = undefined> {
  game: G;
  state: S;
  t: TFunction;
  tc: TFunction;
  busy: boolean;
  requestConfirm: (action: () => void) => void;
  confirmGiveUpAction: (() => void) | undefined;
  frontendHint: UseGameHintReturn;
  showActionLog: () => void;
  hideActionLog: () => void;
  view: V;
}

type GamePageKey = keyof typeof actionLogApi & HintGameName & keyof typeof gameTheme & GameKey;

/**
 * Definition for the game-specific pieces and CLI configuration of a page.
 *
 * Board, FooterActions, and HeaderExtra unmount in CLI mode, so state that must
 * persist belongs in useGame or useView. phase, shortcuts, settings, cli, and shell
 * are plain functions that must not call hooks.
 */
export interface GameDefinition<S, G extends GameHookBase<S, A>, A extends unknown[], V = undefined> {
  key: GamePageKey;
  useGame: () => G;
  /**
   * Optional view hook for state that must survive CLI toggling.
   * Called unconditionally on every render before the early return on null state.
   */
  useView?: (input: { game: G; t: TFunction; tc: TFunction }) => V;
  tutorial: TutorialStep[];
  skeleton: GameSkeletonProps['layout'];
  /** Plain function returning CLI configuration and optional success callback. Must not call hooks. */
  cli: (ctx: { game: G; state: S | null; view: V }) => Omit<CliGameConfig<S, A>, 'gameName'> & {
    onCommandSuccess?: (args: A) => void;
  };
  busy?: (game: G) => boolean;
  /** Plain function computing the current phase display. Must not call hooks. */
  phase: (ctx: GamePageContext<S, G, V>) => { name: string; ended: boolean; won?: boolean };
  /** Plain function returning action shortcuts. Must not call hooks. */
  shortcuts?: (ctx: GamePageContext<S, G, V>) => ActionBinding[];
  /** Plain function returning whether shortcuts are enabled. Defaults to true. Must not call hooks. */
  shortcutsEnabled?: (ctx: GamePageContext<S, G, V>) => boolean;
  shortcutsTestId?: string;
  /** Plain function returning settings groups. Must not call hooks. */
  settings?: (ctx: GamePageContext<S, G, V>) => SettingsGroup[];
  resetTutorialId?: string;
  HeaderExtra?: ComponentType<GamePageContext<S, G, V>>;
  LiveRegion?: ComponentType<GamePageContext<S, G, V>>;
  Board: ComponentType<GamePageContext<S, G, V>>;
  FooterActions?: ComponentType<GamePageContext<S, G, V>>;
  /** Plain function returning additional shell props spread into GamePageShell. Must not call hooks. */
  shell?: (
    ctx: GamePageContext<S, G, V>,
  ) => Pick<GamePageShellProps, 'isHumanTurn' | 'onCelebrate' | 'lossShow' | 'headerEnd' | 'outerKey'>;
  /** Plain function replacing game.handleReset on manual reset. Must not call hooks. */
  reset?: (ctx: GamePageContext<S, G, V>) => void;
  /** Replacement class for the main scrollable content area. Defaults to "pt-3 px-4 lg:px-8". */
  contentClassName?: string;
}

/** Builds a game page from shared wiring and a game-specific definition. */
export function createGamePage<S, G extends GameHookBase<S, A>, A extends unknown[], V = undefined>(
  def: GameDefinition<S, G, A, V>,
): ComponentType {
  const useView = def.useView ?? (() => undefined as V);

  function Content() {
    const setup = useGamePageSetup(def.key);
    const game = def.useGame();
    const state = game.state;
    const view = useView({ game, t: setup.t, tc: setup.tc });
    const frontendHint = useGameHint(def.key, state);
    const cliMode = useCliMode(def.key);
    const { onCommandSuccess, ...rawCliConfig } = def.cli({ game, state, view });
    const cliConfig = { gameName: def.key, ...rawCliConfig };
    const { handleCommand } = useCliGame(
      game.exec,
      cliConfig,
      state,
      {
        addInput: cliMode.addInput,
        addOutput: cliMode.addOutput,
        addError: cliMode.addError,
        clearLog: cliMode.clearLog,
      },
      { onCommandSuccess },
    );
    const rawConfirmGiveUpAction = useGiveUpConfirm(game.handleGiveUp ?? (() => undefined), setup.requestGiveUpConfirm);
    const confirmGiveUpAction = game.handleGiveUp !== undefined ? rawConfirmGiveUpAction : undefined;
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
            view,
          } satisfies GamePageContext<S, G, V>);
    const shortcuts = context ? (def.shortcuts?.(context) ?? []) : [];
    const shortcutsEnabled = context ? (def.shortcutsEnabled?.(context) ?? true) : false;
    useActionKeyboardNav({ bindings: shortcuts, enabled: shortcutsEnabled && !busy });
    const handleManualReset = useCallback(() => {
      setup.hideActionLog();
      if (def.reset && context) {
        def.reset(context);
      } else {
        game.handleReset();
      }
    }, [context, game.handleReset, setup.hideActionLog]);

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
    const shellProps = def.shell?.(context);
    const giveUpProps: GiveUpConfirmProps =
      game.handleGiveUp !== undefined
        ? {
            giveUpConfirmOpen: setup.giveUpConfirmOpen,
            confirmGiveUp: setup.confirmGiveUp,
            cancelGiveUp: setup.cancelGiveUp,
          }
        : {};

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
        {...giveUpProps}
        headerExtra={
          <>
            {HeaderExtra && <HeaderExtra {...context} />}
            <CliToggle cliEnabled={cliMode.cliEnabled} onToggle={cliMode.toggleCli} />
          </>
        }
        {...shellProps}
      >
        {LiveRegion && <LiveRegion {...context} />}
        {cliMode.cliEnabled ? (
          <CliTerminal logEntries={cliMode.logEntries} onCommand={handleCommand} disabled={game.loading} />
        ) : (
          <>
            <LandscapeBanner message={setup.t('landscapeBanner')} />
            <div className={`flex-1 overflow-y-auto ${def.contentClassName ?? 'pt-3 px-4 lg:px-8'}`}>
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
  Content.displayName = `${def.key}PageContent`;
  return withTutorial(Content, def.key, def.tutorial);
}
