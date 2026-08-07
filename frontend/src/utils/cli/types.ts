/** A single entry in the CLI terminal log. */
export interface CliLogEntry {
  /** Entry type: input (user command), output (game state), or error. */
  type: 'input' | 'output' | 'error';
  /** Text content to display. */
  text: string;
  /** Unique identifier for React key. */
  id: number;
}

/** Result of parsing a CLI command — either exec args or an error message. */
export type CliParseResult<TArgs extends unknown[]> = { args: TArgs } | { error: string };

/** Configuration for CLI mode in a specific game. */
export interface CliGameConfig<TState, TArgs extends unknown[]> {
  /** Game identifier (e.g., 'blackjack'). */
  gameName: string;
  /** Parse user input into API exec arguments or an error. */
  parseCommand: (input: string) => CliParseResult<TArgs>;
  /** Format the game state into terminal text. */
  formatResponse: (state: TState) => string;
  /** Help text lines shown when the user types "help". */
  helpText: string[];
  /**
   * Answer a command locally, without calling the API. Return `null` to fall
   * through to {@link parseCommand}.
   *
   * Some commands only report on state the page already holds — Wasp's `legal`
   * lists the columns a card may move onto, which `waspLegalTargets` derives
   * client-side (#4792). Routing those through the server would add an API
   * action that computes nothing new.
   */
  localCommand?: (input: string) => string | null;
}
