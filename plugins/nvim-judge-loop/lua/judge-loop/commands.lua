local M = {}

-- Statuses that mean the judge has finished — stop polling.
local TERMINAL_STATUSES = {
	accepted = true,
	wrong_answer = true,
	compile_error = true,
	runtime_error = true,
	time_limit_exceeded = true,
}

-- Poll GET /local/submissions/:id every 1.5 s until a terminal status arrives
-- or we exhaust the retry budget (~30 s total).
local function poll_verdict(submission_id, remaining)
	if remaining <= 0 then
		require("judge-loop.ui").warn("Verdict timeout — check submission " .. submission_id)
		return
	end
	require("judge-loop.agent").submission_get(submission_id, function(ok, data, err)
		if not ok or not data then
			if err and err.message and err.message ~= "" and remaining % 5 == 0 then
				require("judge-loop.ui").warn("Polling verdict failed: " .. err.message)
			end
			vim.defer_fn(function()
				poll_verdict(submission_id, remaining - 1)
			end, 2000)
			return
		end
		local status = data.status or "pending"
		if TERMINAL_STATUSES[status] then
			require("judge-loop.ui").verdict(status, submission_id)
		else
			vim.defer_fn(function()
				poll_verdict(submission_id, remaining - 1)
			end, 1500)
		end
	end)
end

function M.register()
	vim.api.nvim_create_user_command("JudgeUI", function()
		require("judge-loop.ui").dashboard()
	end, { desc = "Open judge-loop UI" })

	vim.api.nvim_create_user_command("JudgeProblems", function()
		require("judge-loop.ui").pick_problem()
	end, { desc = "Browse judge-loop problems" })

	vim.api.nvim_create_user_command("JudgeSuggest", function()
		require("judge-loop.ui").suggest_problem()
	end, { desc = "Suggest a judge-loop problem" })

	vim.api.nvim_create_user_command("JudgeLanguage", function(opts)
		require("judge-loop.ui").switch_language(opts.args ~= "" and opts.args or nil)
	end, { nargs = "?", desc = "Switch judge-loop solve buffer language" })

	-- :JudgeStatus — print today's practice summary
	vim.api.nvim_create_user_command("JudgeStatus", function()
		require("judge-loop.agent").status_today(function(ok, data, err)
			if not ok then
				require("judge-loop.ui").api_error("Could not load status", err)
				return
			end
			require("judge-loop.ui").info(data.message or "OK")
		end)
	end, { desc = "Show today's practice status" })

	-- :JudgeStart [problem_id] — start in-memory timer
	vim.api.nvim_create_user_command("JudgeStart", function(opts)
		local problem_id = opts.args ~= "" and opts.args or nil
		if problem_id then
			require("judge-loop.agent").problem_get(problem_id, function(ok, problem, err)
				if ok and problem then
					require("judge-loop.ui").start_problem(problem)
					return
				end
				require("judge-loop.ui").api_error("Problem not found: " .. problem_id, err)
			end)
			return
		end
		require("judge-loop.agent").timer_start(problem_id, function(ok, data, err)
			if not ok then
				require("judge-loop.ui").api_error("Failed to start timer", err)
				return
			end
			require("judge-loop.ui").info("Timer started at " .. (data.started_at or "?"))
		end)
	end, { nargs = "?", desc = "Start practice timer [problem_id]" })

	-- :JudgeStop — stop active timer
	vim.api.nvim_create_user_command("JudgeStop", function()
		require("judge-loop.agent").timer_stop(function(ok, data, err)
			if not ok then
				require("judge-loop.ui").api_error("Failed to stop timer", err)
				return
			end
			local secs = data.elapsed_seconds or 0
			local mins = math.floor(secs / 60)
			require("judge-loop.ui").info(string.format("Timer stopped. Elapsed: %d min %d sec", mins, secs % 60))
		end)
	end, { desc = "Stop active practice timer" })

	-- :JudgeSubmit — submit current buffer to the judge
	vim.api.nvim_create_user_command("JudgeSubmit", function()
		require("judge-loop.ui").save_current_code()

		local problem_id = vim.b.judge_problem_id
		if not problem_id or problem_id == "" then
			require("judge-loop.ui").error("No problem set. Run :JudgeStart <problem_id> first.")
			return
		end
		local lang = vim.bo.filetype
		local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
		local code = table.concat(lines, "\n")
		require("judge-loop.ui").info("Submitting…")
		require("judge-loop.agent").submit(problem_id, lang, code, function(ok, data, err)
			if not ok then
				require("judge-loop.ui").api_error("Submission failed", err)
				return
			end
			local sub_id = data.submission_id or data.id or ""
			require("judge-loop.ui").info("Queued — waiting for verdict…")
			-- Start polling after a short delay; judge-worker needs time to pick up the job.
			vim.defer_fn(function()
				poll_verdict(sub_id, 20) -- 20 attempts × 1.5 s ≈ 30 s budget
			end, 800)
		end)
	end, { desc = "Submit current buffer to judge" })

	-- :JudgeSync — trigger registry sync (stub until Milestone 7)
	vim.api.nvim_create_user_command("JudgeSync", function()
		require("judge-loop.ui").sync_registry()
	end, { desc = "Sync problem registry" })
end

return M
