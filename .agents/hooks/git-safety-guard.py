#!/usr/bin/env python3
"""
PreToolUse Hook for Git Safety Guard.
Intercepts run_command tool calls to prevent destructive or unauthorized git operations.
"""

import json
import os
import shlex
import sys


def check_command_tokens(tokens):
    """Inspects tokens of a single command.

    Returns (is_blocked, reason).
    """
    if not tokens:
        return False, None

    # Skip variable assignments (e.g., VAR=val) and wrapper commands
    idx = 0
    wrapper_commands = {"sudo", "nohup", "time", "env", "exec"}
    while idx < len(tokens):
        tok = tokens[idx]
        if "=" in tok and not tok.startswith("-"):
            idx += 1
            continue
        if tok in wrapper_commands:
            idx += 1
            continue
        break

    if idx >= len(tokens):
        return False, None

    cmd = os.path.basename(tokens[idx])
    if cmd != "git":
        return False, None

    # Skip git global options (flags and flags with arguments)
    idx += 1
    flags_with_args = {
        "-C",
        "-c",
        "--git-dir",
        "--work-tree",
        "--namespace",
        "--super-prefix",
        "--exec-path",
    }
    subcmd = None
    subcmd_idx = None
    while idx < len(tokens):
        tok = tokens[idx]
        if tok in flags_with_args:
            idx += 2
            continue
        if any(tok.startswith(f"{f}=") for f in flags_with_args):
            idx += 1
            continue
        if tok.startswith("-"):
            idx += 1
            continue
        subcmd = tok
        subcmd_idx = idx
        break

    if not subcmd:
        return False, None

    subcmd_args = tokens[subcmd_idx + 1 :]

    # 1. git commit
    if subcmd == "commit":
        return (
            True,
            "git commit は禁止されています。コミットは呼び出し側が行うため、作業ツリーに変更を残したまま報告してください。",
        )

    # 2. git reset
    if subcmd == "reset":
        return (
            True,
            "git reset は禁止されています。未コミット変更を破棄せず、作業ツリーに残したまま報告してください。",
        )

    # 3. git restore
    if subcmd == "restore":
        return (
            True,
            "git restore は禁止されています。未コミット変更を破棄せず、作業ツリーに残したまま報告してください。",
        )

    # 4. git stash
    if subcmd == "stash":
        return (
            True,
            "git stash は禁止されています。未コミット変更を退避せず、作業ツリーに残したまま報告してください。",
        )

    # 5. git push
    if subcmd == "push":
        return (
            True,
            "git push は禁止されています。リモートへのプッシュは呼び出し側が行います。",
        )

    # 6. git checkout (only tree-discarding variants)
    if subcmd == "checkout":
        discard_flags = {
            "--",
            ".",
            "-f",
            "--force",
            "--ours",
            "--theirs",
            "--merge",
            "-m",
            "--patch",
            "-p",
        }
        for arg in subcmd_args:
            if arg in discard_flags:
                return (
                    True,
                    "git checkout による作業ツリーの復元・変更破棄は禁止されています。変更は破棄せず作業ツリーに残したまま報告してください。",
                )
        if len(subcmd_args) >= 2 and subcmd_args[0] == "HEAD":
            return (
                True,
                "git checkout による作業ツリーの復元・変更破棄は禁止されています。変更は破棄せず作業ツリーに残したまま報告してください。",
            )

    return False, None


def parse_and_check(command_line):
    """Splits command line across pipeline/chains and checks each command."""
    if not command_line or not command_line.strip():
        return False, None

    try:
        lexer = shlex.shlex(command_line, punctuation_chars=True)
        lexer.whitespace_split = True
        tokens = list(lexer)
    except Exception:
        tokens = command_line.strip().split()

    chain_delimiters = {";", "&&", "||", "|", "&", "\n"}
    current_cmd_tokens = []
    for tok in tokens:
        if tok in chain_delimiters:
            if current_cmd_tokens:
                blocked, reason = check_command_tokens(current_cmd_tokens)
                if blocked:
                    return True, reason
                current_cmd_tokens = []
        else:
            current_cmd_tokens.append(tok)

    if current_cmd_tokens:
        blocked, reason = check_command_tokens(current_cmd_tokens)
        if blocked:
            return True, reason

    return False, None


def main():
    try:
        input_data = sys.stdin.read()
        if not input_data.strip():
            print(json.dumps({"decision": "allow"}))
            return
        payload = json.loads(input_data)
    except Exception:
        print(json.dumps({"decision": "allow"}))
        return

    tool_call = payload.get("toolCall", {})
    if tool_call.get("name") != "run_command":
        print(json.dumps({"decision": "allow"}))
        return

    args = tool_call.get("args", {})
    command_line = (
        args.get("CommandLine")
        or args.get("commandLine")
        or args.get("command_line")
        or ""
    )

    blocked, reason = parse_and_check(command_line)
    if blocked:
        print(
            json.dumps(
                {"decision": "deny", "reason": reason}, ensure_ascii=False
            )
        )
    else:
        print(json.dumps({"decision": "allow"}, ensure_ascii=False))


if __name__ == "__main__":
    main()
