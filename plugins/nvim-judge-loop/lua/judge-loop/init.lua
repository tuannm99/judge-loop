local M = {}

M.config = {
  agent_url = "http://127.0.0.1:7070",
  auto_notify = true, -- show WARN on VimEnter if no practice yet today
}

function M.setup(opts)
  M.config = vim.tbl_deep_extend("force", M.config, opts or {})

  require("judge-loop.commands").register()

  -- On VimEnter: fetch today's status and remind the user if they haven't practiced.
  vim.api.nvim_create_autocmd("VimEnter", {
    once = true,
    callback = function()
      if not M.config.auto_notify then
        return
      end
      -- Defer so other plugins finish loading first.
      vim.defer_fn(function()
        require("judge-loop.agent").status_today(function(ok, data)
          if ok and not data.practiced then
            require("judge-loop.ui").warn(data.message or "No practice yet today. Start a session!")
          end
        end)
      end, 1000)
    end,
  })
end

return M
