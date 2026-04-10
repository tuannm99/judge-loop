local M = {}

M.config = {
	agent_url = "http://127.0.0.1:7070",
	auto_notify = true, -- show WARN on VimEnter if no practice yet today
	cache_dir = vim.fn.expand("~/.judgeloopcache"),
	editor = {
		side = "left",
		width = 80,
	},
}

function M.setup(opts)
	M.config = vim.tbl_deep_extend("force", M.config, opts or {})

	require("judge-loop.commands").register()

	vim.api.nvim_create_autocmd("BufLeave", {
		callback = function()
			if vim.b.judge_cache_path then
				require("judge-loop.ui").save_current_code()
			end
		end,
	})

	-- On VimEnter: fetch today's status and remind the user if they haven't practiced.
	vim.api.nvim_create_autocmd("VimEnter", {
		once = true,
		callback = function()
			if not M.config.auto_notify then
				return
			end
			-- Defer so other plugins finish loading first.
			vim.defer_fn(function()
				require("judge-loop.agent").status_today(function(ok, data, err)
					if ok and not data.practiced then
						require("judge-loop.ui").warn(data.message or "No practice yet today. Start a session!")
					elseif not ok and err and err.message and err.message ~= "" then
						require("judge-loop.ui").warn("Status check failed: " .. err.message)
					end
				end)
			end, 1000)
		end,
	})
end

function M.timer_statusline()
	return require("judge-loop.statusline").timer_statusline()
end

return M
