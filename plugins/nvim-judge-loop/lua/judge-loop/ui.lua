local M = {}
local async = require("judge-loop.async")
local agent = require("judge-loop.agent")

local levels = vim.log.levels
local state = { win = nil, buf = nil }

function M.info(msg)
	vim.schedule(function()
		vim.notify("[judge-loop] " .. msg, levels.INFO)
	end)
end

function M.warn(msg)
	vim.schedule(function()
		vim.notify("[judge-loop] " .. msg, levels.WARN)
	end)
end

function M.error(msg)
	vim.schedule(function()
		vim.notify("[judge-loop] " .. msg, levels.ERROR)
	end)
end

function M.api_error(prefix, err)
	local msg = prefix
	if err and err.message and err.message ~= "" then
		msg = msg .. ": " .. err.message
	elseif err and err.status and err.status > 0 then
		msg = msg .. ": HTTP " .. tostring(err.status)
	end
	M.error(msg)
end

local verdict_labels = {
	accepted = "✓ Accepted",
	wrong_answer = "✗ Wrong Answer",
	time_limit_exceeded = "⏱ Time Limit Exceeded",
	runtime_error = "💥 Runtime Error",
	compile_error = "🔨 Compile Error",
	pending = "⏳ Pending…",
}

local default_code = {
	python = "# Write your solution here\n\n",
	go = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println()\n}\n',
}

local language_ext = {
	python = "py",
	go = "go",
}

local function config()
	return require("judge-loop").config
end

local function problem_key(problem)
	return problem.slug or problem.id
end

local function problem_cache_dir(problem)
	return config().cache_dir .. "/" .. problem_key(problem)
end

local function cache_path(problem, language)
	local ext = language_ext[language] or language
	return problem_cache_dir(problem) .. "/solution." .. ext
end

local function meta_path(problem)
	return problem_cache_dir(problem) .. "/problem.json"
end

local function language_pool(problem)
	local starter = problem.starter_code or {}
	local pool = {}
	if starter.python or not starter.go then
		table.insert(pool, "python")
	end
	if starter.go or #pool == 0 then
		table.insert(pool, "go")
	end
	return pool
end

local function write_cache_file(path, code)
	if vim.fn.filereadable(path) == 1 then
		return
	end
	vim.fn.mkdir(vim.fn.fnamemodify(path, ":h"), "p")
	vim.fn.writefile(vim.split(code or "", "\n", { plain = true }), path)
end

local function cache_metadata(problem)
	vim.fn.mkdir(problem_cache_dir(problem), "p")
	vim.fn.writefile({ vim.fn.json_encode(problem) }, meta_path(problem))
end

local function read_metadata(dir)
	local path = dir .. "/problem.json"
	if vim.fn.filereadable(path) ~= 1 then
		return nil
	end
	local raw = table.concat(vim.fn.readfile(path), "\n")
	local ok, decoded = pcall(vim.fn.json_decode, raw)
	if ok then
		return decoded
	end
	return nil
end

local function save_current_code()
	if not vim.b.judge_cache_path or vim.bo.buftype ~= "" then
		return
	end
	local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
	vim.fn.mkdir(vim.fn.fnamemodify(vim.b.judge_cache_path, ":h"), "p")
	vim.fn.writefile(lines, vim.b.judge_cache_path)
	vim.bo.modified = false
end

local function open_code_panel(path)
	local editor = config().editor or {}
	local side = editor.side or "left"
	local width = tonumber(editor.width) or 80
	local placement = side == "right" and "botright" or "topleft"

	vim.cmd(placement .. " vertical " .. width .. "new")
	vim.cmd("noswapfile edit " .. vim.fn.fnameescape(path))
	vim.cmd("vertical resize " .. width)
end

local function close_window()
	if state.win and vim.api.nvim_win_is_valid(state.win) then
		vim.api.nvim_win_close(state.win, true)
	end
	state.win = nil
	state.buf = nil
end

local function open_window(title, lines, mappings)
	close_window()

	local width = math.min(math.max(72, math.floor(vim.o.columns * 0.7)), vim.o.columns - 4)
	local height = math.min(math.max(14, #lines + 4), vim.o.lines - 4)
	local row = math.floor((vim.o.lines - height) / 2)
	local col = math.floor((vim.o.columns - width) / 2)

	local buf = vim.api.nvim_create_buf(false, true)
	vim.api.nvim_buf_set_option(buf, "buftype", "nofile")
	vim.api.nvim_buf_set_option(buf, "bufhidden", "wipe")
	vim.api.nvim_buf_set_option(buf, "swapfile", false)
	vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
	vim.api.nvim_buf_set_option(buf, "modifiable", false)

	local win = vim.api.nvim_open_win(buf, true, {
		relative = "editor",
		width = width,
		height = height,
		row = row,
		col = col,
		style = "minimal",
		border = "rounded",
		title = " " .. title .. " ",
		title_pos = "center",
	})

	state.buf = buf
	state.win = win

	vim.keymap.set("n", "q", close_window, { buffer = buf, nowait = true, silent = true })
	if mappings then
		for lhs, rhs in pairs(mappings) do
			vim.keymap.set("n", lhs, rhs, { buffer = buf, nowait = true, silent = true })
		end
	end
end

local function fmt_problem(problem)
	local tags = table.concat(problem.tags or {}, ", ")
	local patterns = table.concat(problem.pattern_tags or {}, ", ")
	return string.format(
		"%s [%s] %s | %s",
		problem.title or problem.slug or problem.id,
		problem.difficulty or "?",
		tags,
		patterns
	)
end

local function open_url(url)
	if not url or url == "" then
		M.warn("No source URL for this problem")
		return
	end
	if vim.ui and vim.ui.open then
		vim.ui.open(url)
	else
		vim.fn.setreg("+", url)
		M.info("Copied source URL to clipboard")
	end
end

local function select_async(items, opts)
	return async.await(function(resolve)
		vim.ui.select(items, opts, resolve)
	end)
end

local function start_problem_async(problem)
	local language = select_async(language_pool(problem), { prompt = "Language" })
	if not language then
		return
	end

	local starter = problem.starter_code or {}
	local code = starter[language] or default_code[language] or ""
	local path = cache_path(problem, language)

	cache_metadata(problem)
	write_cache_file(path, code)
	open_code_panel(path)

	vim.bo.swapfile = false
	vim.bo.filetype = language
	vim.b.judge_problem_id = problem.id
	vim.b.judge_problem_slug = problem.slug
	vim.b.judge_problem_title = problem.title
	vim.b.judge_problem_languages = language_pool(problem)
	vim.b.judge_problem_metadata = problem
	vim.b.judge_cache_path = path

	local ok, data, err = agent.timer_start_await(problem.id)
	if not ok then
		local detail = err and err.message and (": " .. err.message) or ""
		M.warn("Opened solve buffer, but timer did not start" .. detail)
		return
	end
	M.info("Solving " .. (problem.title or problem.slug or problem.id) .. " from " .. (data.started_at or "now"))
end

function M.start_problem(problem)
	async.run(function()
		start_problem_async(problem)
	end)
end

function M.switch_language(language)
	local problem = vim.b.judge_problem_metadata
	if not problem and vim.b.judge_cache_path then
		problem = read_metadata(vim.fn.fnamemodify(vim.b.judge_cache_path, ":h"))
	end
	if not problem then
		M.error("No judge-loop problem in this buffer")
		return
	end

	local function open_language(next_language)
		if not next_language or next_language == "" then
			return
		end
		save_current_code()

		local starter = problem.starter_code or {}
		local code = starter[next_language] or default_code[next_language] or ""
		local path = cache_path(problem, next_language)
		write_cache_file(path, code)

		vim.cmd("noswapfile edit " .. vim.fn.fnameescape(path))
		vim.bo.swapfile = false
		vim.bo.filetype = next_language
		vim.b.judge_problem_id = problem.id
		vim.b.judge_problem_slug = problem.slug
		vim.b.judge_problem_title = problem.title
		vim.b.judge_problem_languages = language_pool(problem)
		vim.b.judge_problem_metadata = problem
		vim.b.judge_cache_path = path
		M.info("Language: " .. next_language)
	end

	if language then
		open_language(language)
		return
	end
	async.run(function()
		open_language(select_async(language_pool(problem), { prompt = "Language" }))
	end)
end

function M.save_current_code()
	save_current_code()
end

function M.show_problem(problem)
	local lines = {
		problem.title or problem.slug or problem.id,
		"",
		"Difficulty: " .. (problem.difficulty or "?"),
		"Provider:   " .. (problem.provider or "?"),
		"Estimate:   " .. tostring(problem.estimated_time or "?") .. "m",
		"Tags:       " .. table.concat(problem.tags or {}, ", "),
		"Patterns:   " .. table.concat(problem.pattern_tags or {}, ", "),
		"Source:     " .. (problem.source_url or ""),
		"",
		"s start solving",
		"o open source",
		"q close",
	}
	open_window("Problem", lines, {
		s = function()
			close_window()
			M.start_problem(problem)
		end,
		o = function()
			open_url(problem.source_url)
		end,
	})
end

local function problem_actions_async(problem)
	local choice = select_async({ "Start solving", "Details", "Open source" }, {
		prompt = problem.title or problem.slug or "Problem",
	})
	if choice == "Start solving" then
		start_problem_async(problem)
	elseif choice == "Details" then
		M.show_problem(problem)
	elseif choice == "Open source" then
		open_url(problem.source_url)
	end
end

function M.problem_actions(problem)
	async.run(function()
		problem_actions_async(problem)
	end)
end

function M.pick_problem()
	async.run(function()
		local ok, data, err = agent.problems_await({ limit = 200 })
		if not ok or not data or not data.problems then
			M.api_error("Could not load problems", err)
			return
		end
		local problem = select_async(data.problems, {
			prompt = "Problems",
			format_item = fmt_problem,
		})
		if problem then
			problem_actions_async(problem)
		end
	end)
end

function M.suggest_problem()
	async.run(function()
		local ok, problem, err = agent.problem_suggest_await()
		if not ok or not problem then
			M.api_error("No suggested problem available", err)
			return
		end
		problem_actions_async(problem)
	end)
end

function M.sync_registry()
	async.run(function()
		local ok, data, err = agent.sync_await()
		if not ok then
			M.api_error("Sync request failed", err)
			return
		end
		M.info(data.message or "Sync done")
	end)
end

local function timer_line(timer)
	if not timer or not timer.active then
		return "Timer: inactive"
	end
	local secs = timer.elapsed_seconds or 0
	return string.format("Timer: active %d:%02d", math.floor(secs / 60), secs % 60)
end

function M.dashboard()
	async.run(function()
		local status_ok, status, status_err = agent.status_today_await()
		local _, timer, timer_err = agent.timer_current_await()
		local status_line = status_ok and (status.message or "Status loaded") or "Local agent unreachable"
		if status_err and status_err.message and status_err.message ~= "" then
			status_line = status_line .. ": " .. status_err.message
		end
		if timer_err and timer_err.message and timer_err.message ~= "" then
			status_line = status_line .. " | timer: " .. timer_err.message
		end
		local lines = {
			"judge-loop",
			"",
			status_line,
			"Solved today: " .. tostring((status and status.solved_count) or 0),
			"Streak:       " .. tostring((status and status.streak) or 0),
			timer_line(timer),
			"",
			"p problems",
			"g suggest",
			"s submit current buffer",
			"x sync registry",
			"t stop timer",
			"r refresh",
			"q close",
		}
		open_window("judge-loop", lines, {
			p = M.pick_problem,
			g = M.suggest_problem,
			s = function()
				close_window()
				vim.cmd("JudgeSubmit")
			end,
			x = M.sync_registry,
			t = function()
				async.run(function()
					local ok, data, err = agent.timer_stop_await()
					if not ok then
						M.api_error("Failed to stop timer", err)
						return
					end
					M.info("Timer stopped. Elapsed: " .. tostring(data.elapsed_seconds or 0) .. "s")
					M.dashboard()
				end)
			end,
			r = M.dashboard,
		})
	end)
end

function M.verdict(status, submission_id)
	local label = verdict_labels[status] or status
	local msg = string.format("%s  (id: %s)", label, submission_id or "?")
	local level = (status == "accepted") and levels.INFO or levels.WARN
	vim.schedule(function()
		vim.notify("[judge-loop] " .. msg, level)
	end)
end

return M
