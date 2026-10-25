import type { TFunction } from 'i18next';
import type { ComponentType, ReactNode } from 'react';
import type { actionLogApi } from '../../api/gameApi';
import type { ActionBinding } from '../../hooks/useActionKeyboardNav';
import { useActionKeyboardNav } from '../../hooks/useActionKeyboardNav';
import { type UseCardKeyboardNavOptions, useCardKeyboardNav } from '../../hooks/useCardKeyboardNav';
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
import { CardNavShortcutsPanel } from '../CardNavShortcutsPanel';
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

/**
 * Common state and actions returned by a game hook.
 *
 * `A` defaults to `never` so `GamePageContext` can accept any game hook without
 * repeating its CLI argument tuple: `exec` is parameter-contravariant, so every
 * concrete `exec` is assignable to `(...args: never) => Promise<void>`.
 */
export interface GameHookBase<S, A extends unknown[] = never> {
  state: S | null;
  loading: boolean;
  error: string | null;
  hintError?: string | null;
  retry: ErrorAlertProps['onRetry'];
  handleReset?: () => void;
  handleGiveUp?: () => void;
  exec: (...args: A) => Promise<void>;
}

/** Values shared by the game-specific slots rendered by {@link createGamePage}. */
export interface GamePageContext<S, G extends GameHookBase<S, never>, V = undefined> {
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
  /** Renders the shared reset button; place it with `resetButton: 'custom'` so the factory does not add its own. */
  renderResetButton: (className?: string) => ReactNode;
}

/** Card navigation used when a page has no `cardNav` or no state yet; the hook still runs to keep hook order. */
const NO_CARD_NAV: UseCardKeyboardNavOptions = {
  cardCount: 0,
  onToggle: () => undefined,
  onConfirm: () => undefined,
  onClear: () => undefined,
  enabled: false,
};

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
  /**
   * Plain function returning CLI configuration and optional success callback. Must not call hooks.
   * `frontendHint` is the same hint state the board sees, so CLI commands honour the hint toggle.
   */
  cli: (ctx: { game: G; state: S | null; view: V; frontendHint: UseGameHintReturn }) => Omit<
    CliGameConfig<S, A>,
    'gameName'
  > & {
    onCommandSuccess?: (args: A) => void;
  };
  busy?: (game: G) => boolean;
  /** Plain function computing the current phase display. Must not call hooks. */
  phase: (ctx: GamePageContext<S, G, V>) => { name: string; ended: boolean; won?: boolean };
  /** Plain function returning action shortcuts. Must not call hooks. */
  shortcuts?: (ctx: GamePageContext<S, G, V>) => ActionBinding[];
  /** Plain function returning pinned primary footer actions. Must not call hooks. Its presence selects the pinned footer layout. Return null (not undefined) on turns with no actions. */
  primaryActions?: (ctx: GamePageContext<S, G, V>) => ReactNode;
  /** Plain function returning persistent footer guidance. Must not call hooks. */
  message?: (ctx: GamePageContext<S, G, V>) => ReactNode;
  /** Plain function returning whether shortcuts are enabled. Defaults to true. Must not call hooks. */
  shortcutsEnabled?: (ctx: GamePageContext<S, G, V>) => boolean;
  shortcutsTestId?: string;
  /**
   * Plain function returning card keyboard navigation handlers. Must not call hooks.
   * When set, `shortcutsTestId` goes to the card-nav panel; an action panel added by
   * `shortcuts` alongside it has no test id.
   */
  cardNav?: (ctx: GamePageContext<S, G, V>) => UseCardKeyboardNavOptions;
  /** Plain function returning settings groups. Must not call hooks. */
  settings?: (ctx: GamePageContext<S, G, V>) => SettingsGroup[];
  /** Controls whether settings appear before or after the scrollable game content. Defaults to bottom. */
  settingsPlacement?: 'top' | 'bottom';
  /** Whether to render the landscape recommendation banner. Defaults to false. */
  landscapeBanner?: boolean;
  resetTutorialId?: string;
  /** Tutorial identifier on the fixed footer action row. Use with `primaryActions`; when omitted, no attribute is rendered. */
  footerTutorialId?: string;
  /** Tutorial identifier applied to the footer actions row. */
  actionsTutorialId?: string;
  /** Replacement classes for the footer actions row. Defaults to "flex gap-2 items-center flex-wrap". */
  actionsClassName?: string;
  /**
   * Where the reset button goes. `'end'` (default) appends it to the actions row; `'custom'` leaves it
   * to the page, which places it with `renderResetButton` from the context.
   */
  resetButton?: 'end' | 'custom';
  HeaderExtra?: ComponentType<GamePageContext<S, G, V>>;
  LiveRegion?: ComponentType<GamePageContext<S, G, V>>;
  /** Optional content rendered after top-placed settings and before the scrolling game content, so it does not scroll. */
  AboveContent?: ComponentType<GamePageContext<S, G, V>>;
  Board: ComponentType<GamePageContext<S, G, V>>;
  /** Optional hand content rendered first in the game footer. */
  FooterHand?: ComponentType<GamePageContext<S, G, V>>;
  /** Optional status content rendered after errors and before footer actions. */
  FooterStatus?: ComponentType<GamePageContext<S, G, V>>;
  FooterActions?: ComponentType<GamePageContext<S, G, V>>;
  /** Plain function returning additional shell props spread into GamePageShell. Must not call hooks. */
  shell?: (
    ctx: GamePageContext<S, G, V>,
  ) => Pick<GamePageShellProps, 'isHumanTurn' | 'onCelebrate' | 'lossShow' | 'headerEnd' | 'outerKey'> &
    Partial<Pick<GamePageShellProps, 'gameThemeBg'>>;
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
    const { onCommandSuccess, ...rawCliConfig } = def.cli({ game, state, view, frontendHint });
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
    // Plain closures: the reset button and confirm dialog read `context` and the phase at click/render time.
    let context: GamePageContext<S, G, V> | null = null;
    const handleManualReset = () => {
      setup.hideActionLog();
      if (def.reset && context) {
        def.reset(context);
      } else {
        game.handleReset?.();
      }
    };
    const renderResetButton = (className?: string) =>
      context ? (
        <GameResetButton
          isGameEnd={def.phase(context).ended}
          onReset={handleManualReset}
          requestConfirm={setup.requestConfirm}
          loading={busy}
          dataTutorial={def.resetTutorialId}
          className={className}
        />
      ) : null;
    context =
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
            renderResetButton,
          } satisfies GamePageContext<S, G, V>);
    const shortcuts = context ? (def.shortcuts?.(context) ?? []) : [];
    const shortcutsEnabled = context ? (def.shortcutsEnabled?.(context) ?? true) : false;
    useActionKeyboardNav({ bindings: shortcuts, enabled: shortcutsEnabled && !busy });
    const cardNav = (context && def.cardNav?.(context)) || NO_CARD_NAV;
    useCardKeyboardNav({ ...cardNav, enabled: cardNav.enabled && !busy });
    if (!state || !context) return <GameSkeleton gameKey={def.key} layout={def.skeleton} />;

    const phase = def.phase(context);
    const settings = def.settings?.(context) ?? [
      {
        items: [hintCheckboxItem(setup.tc, frontendHint.hintEnabled, frontendHint.setHintEnabled)],
      },
    ];
    const HeaderExtra = def.HeaderExtra;
    const LiveRegion = def.LiveRegion;
    const AboveContent = def.AboveContent;
    const Board = def.Board;
    const FooterHand = def.FooterHand;
    const FooterStatus = def.FooterStatus;
    const FooterActions = def.FooterActions;
    const primaryActions = def.primaryActions?.(context);
    const message = def.message?.(context);
    const shellProps = def.shell?.(context);
    // The error alert, status, action row and shortcuts sit inside the footer in the legacy layout and
    // at the end of the scrolling content when `primaryActions` pins the footer's action row.
    const footerControls = (
      <>
        <ErrorAlert message={game.error ?? game.hintError ?? null} onRetry={game.retry} />
        {FooterStatus && <FooterStatus {...context} />}
        <div
          className={def.actionsClassName ?? 'flex gap-2 items-center flex-wrap'}
          data-tutorial={def.actionsTutorialId}
        >
          {FooterActions && <FooterActions {...context} />}
          {def.resetButton !== 'custom' && renderResetButton()}
        </div>
        {def.cardNav ? (
          <CardNavShortcutsPanel data-testid={def.shortcutsTestId} />
        ) : (
          <ActionShortcutsPanel bindings={shortcuts} data-testid={def.shortcutsTestId} />
        )}
        {def.cardNav && def.shortcuts && <ActionShortcutsPanel bindings={shortcuts} />}
      </>
    );
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
            {def.landscapeBanner && <LandscapeBanner message={setup.t('landscapeBanner')} />}
            {def.settingsPlacement === 'top' && <SettingsPanel title={setup.tc('settings.title')} groups={settings} />}
            {AboveContent && <AboveContent {...context} />}
            <div className={`flex-1 overflow-y-auto ${def.contentClassName ?? 'pt-3 px-4 lg:px-8'}`}>
              <Board {...context} />
              <ActionLogSection
                isEndPhase={phase.ended}
                actionLog={setup.actionLog}
                showActionLog={setup.showActionLog}
                hideActionLog={setup.hideActionLog}
              />
              {def.primaryActions && footerControls}
            </div>
            {def.settingsPlacement !== 'top' && <SettingsPanel title={setup.tc('settings.title')} groups={settings} />}
            {def.primaryActions ? (
              <GameFooter
                className={`${gameTheme[def.key].footer} px-4 py-2.5`}
                actions={primaryActions ?? null}
                message={message}
                dataTutorial={def.footerTutorialId}
              >
                {FooterHand && <FooterHand {...context} />}
              </GameFooter>
            ) : (
              <GameFooter className={`${gameTheme[def.key].footer} px-4 py-2.5`} message={message}>
                {FooterHand && <FooterHand {...context} />}
                {footerControls}
              </GameFooter>
            )}
          </>
        )}
      </GamePageShell>
    );
  }
  Content.displayName = `${def.key}PageContent`;
  return withTutorial(Content, def.key, def.tutorial);
}
