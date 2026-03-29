# Neovim Plugin

## Overview

`nvim-judge-loop` is a Lua plugin for Neovim that integrates with the local agent to enforce daily practice.

## Installation

Using lazy.nvim:
```lua
{
  "tuannm99/nvim-judge-loop",
  config = function()
    require("judge-loop").setup({
      agent_url = "http://localhost:7070",
      auto_start_agent = true,
      remind_on_open = true,
    })
  end
}
```

## Commands

| Command | Description |
|---------|-------------|
| `:JudgeStatus` | Show today's practice status |
| `:JudgeStart [problem_slug]` | Start a timed session |
| `:JudgeStop` | Stop active timer |
| `:JudgeSubmit` | Submit current buffer |
| `:JudgeMission` | Show daily mission |
| `:JudgeProblems` | Browse problem list |
| `:JudgeSync` | Sync registry from server |

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

## Configuration

```lua
require("judge-loop").setup({
  agent_url = "http://localhost:7070",  -- local agent URL
  auto_start_agent = true,             -- spawn agent if not running
  remind_on_open = true,               -- check on VimEnter
  notify_level = vim.log.levels.WARN,  -- reminder log level
  statusline = true,                   -- enable timer in statusline
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
      agent.lua          -- HTTP client for local agent
      commands.lua       -- command definitions
      statusline.lua     -- timer statusline component
      ui.lua             -- notifications and pickers
```
