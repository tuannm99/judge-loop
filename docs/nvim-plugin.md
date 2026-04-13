# Neovim Plugin

## Overview

`nvim-judge-loop` is a Lua plugin for Neovim that integrates with the local agent to enforce daily practice.

Internally, async HTTP and picker flows use Lua coroutines with small `*_await()` helpers, so plugin code can stay sequential without adding dependencies.

## Installation

Using lazy.nvim:

```lua
{
  "tuannm99/nvim-judge-loop",
  config = function()
    require("judge-loop").setup({
      agent_url = "http://localhost:7070",
      auto_notify = true,
    })
  end
}
```

## Commands

| Command                     | Description                                        |
| --------------------------- | -------------------------------------------------- |
| `:JudgeUI`                  | Open the floating judge-loop UI                    |
| `:JudgeStatus`              | Show today's practice status                       |
| `:JudgeProblems`            | Browse problem list                                |
| `:JudgeSuggest`             | Pick a suggested problem                           |
| `:JudgeLanguage [language]` | Switch the solve buffer language                   |
| `:JudgeStart [problem_id]`  | Open a problem in the solve panel or start a timer |
| `:JudgeStop`                | Stop active timer                                  |
| `:JudgeSubmit`              | Submit current buffer                              |
| `:JudgeSync`                | Sync registry from server                          |

## Startup reminder

On `VimEnter`, the plugin calls `GET /local/status/today`.

If `practiced: false`, it displays a notification:

```
[judge-loop] No practice today yet! Run :JudgeStart to begin.
```

Uses `vim.notify` with level `WARN`. Compatible with nvim-notify.

## Timer display

When a timer is active, elapsed time is shown in the status line.

The plugin exposes a function for status line integration:

```lua
require("judge-loop").timer_statusline()
-- returns: " 12:34" or "" if no active timer
```

## Submit flow

`:JudgeSubmit`:

1. Reads current buffer content
2. Detects language from `&filetype`
3. Reads `problem_id` from buffer variable `b:judge_problem_id` (set by `:JudgeStart`)
4. POSTs to `/local/submit`
5. Shows notification with verdict

## Floating UI

`:JudgeUI` opens a dependency-free floating panel for the same core workflow as the web UI:

- check today's status and active timer
- browse problems
- accept a suggested problem
- open a configurable code-only side panel with starter code
- submit the current buffer
- stop the active timer
- sync the local registry

Solve buffers are cached on disk under `~/.judgeloopcache` by default. The first open uses problem starter code; later opens reuse the cached code for that problem and language.

API calls report structured failures in the UI. Network failures, non-2xx responses, and invalid JSON include the local-agent/api-server error message when available.

## Configuration

```lua
require("judge-loop").setup({
  agent_url = "http://localhost:7070", -- local agent URL
  auto_notify = true,                  -- check on VimEnter
  cache_dir = vim.fn.expand("~/.judgeloopcache"),
  editor = {
    side = "left", -- left or right
    width = 80,
  },
})
```

## File structure

```
nvim-judge-loop/
  plugin/
    judge-loop.lua       -- entry point, loaded by Neovim
  lua/
    judge-loop/
      init.lua           -- setup() function
      async.lua          -- coroutine helpers for await-style flows
      agent.lua          -- HTTP client for local agent
      commands.lua       -- command definitions
      statusline.lua     -- timer statusline component
      ui.lua             -- notifications and pickers
```
