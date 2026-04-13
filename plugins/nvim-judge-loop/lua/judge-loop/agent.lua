-- HTTP client for local-agent (127.0.0.1:7070).
-- Uses curl via vim.fn.jobstart (async, no UI blocking).
-- Public requests remain callback-based, with coroutine-friendly *_await helpers.
-- No external dependencies required.

local M = {}
local async = require("judge-loop.async")
local pack = table.pack or function(...)
	return { n = select("#", ...), ... }
end
local unpack = table.unpack or unpack

local function get_url()
	return require("judge-loop").config.agent_url
end

local function curl_async(method, path, body, callback)
	local url = get_url() .. path
	local cmd = {
		"curl",
		"-sS",
		"--max-time",
		"5",
		"-X",
		method,
		url,
		"-H",
		"Content-Type: application/json",
		"-w",
		"\n%{http_code}",
	}
	if body then
		table.insert(cmd, "-d")
		table.insert(cmd, vim.fn.json_encode(body))
	end

	local output = {}
	local stderr = {}
	local job_id = vim.fn.jobstart(cmd, {
		stdout_buffered = true,
		stderr_buffered = true,
		on_stdout = function(_, data)
			if data then
				for _, line in ipairs(data) do
					table.insert(output, line)
				end
			end
		end,
		on_stderr = function(_, data)
			if data then
				for _, line in ipairs(data) do
					if line ~= "" then
						table.insert(stderr, line)
					end
				end
			end
		end,
		on_exit = function(_, code)
			local raw = table.concat(output, "\n")
			local body_raw, status_raw = raw:match("^(.*)\n(%d%d%d)%s*$")
			local status = tonumber(status_raw or "0") or 0

			if code ~= 0 then
				local message = table.concat(stderr, "\n")
				if message == "" then
					message = "curl exited with code " .. tostring(code)
				end
				vim.schedule(function()
					callback(false, nil, {
						message = message,
						code = code,
						status = status,
						body = body_raw,
					})
				end)
				return
			end

			if not body_raw then
				vim.schedule(function()
					callback(false, nil, {
						message = "missing HTTP status from curl response",
						status = status,
						body = raw,
					})
				end)
				return
			end

			local decoded = nil
			if body_raw ~= "" then
				local decode_ok, value = pcall(vim.fn.json_decode, body_raw)
				if not decode_ok then
					vim.schedule(function()
						callback(false, nil, {
							message = "invalid JSON response",
							status = status,
							body = body_raw,
						})
					end)
					return
				end
				decoded = value
			end

			if status >= 400 then
				local message = "HTTP " .. tostring(status)
				if type(decoded) == "table" and decoded.error then
					message = decoded.error
				elseif type(decoded) == "table" and decoded.message then
					message = decoded.message
				elseif body_raw ~= "" then
					message = body_raw
				end
				vim.schedule(function()
					callback(false, decoded, {
						message = message,
						status = status,
						body = body_raw,
					})
				end)
				return
			end

			vim.schedule(function()
				callback(true, decoded, nil)
			end)
		end,
	})

	if job_id <= 0 then
		vim.schedule(function()
			callback(false, nil, { message = "failed to start curl job" })
		end)
	end
end

local function encode_query(params)
	local query = {}
	for key, value in pairs(params or {}) do
		if value and value ~= "" then
			table.insert(query, vim.uri_encode(key) .. "=" .. vim.uri_encode(tostring(value)))
		end
	end
	if #query == 0 then
		return ""
	end
	return "?" .. table.concat(query, "&")
end

function M.status_today(callback)
	curl_async("GET", "/local/status/today", nil, callback)
end

function M.problems(params, callback)
	curl_async("GET", "/local/problems" .. encode_query(params), nil, callback)
end

function M.problem_get(id, callback)
	curl_async("GET", "/local/problems/" .. id, nil, callback)
end

function M.problem_suggest(callback)
	curl_async("GET", "/local/problems/suggest", nil, callback)
end

function M.timer_current(callback)
	curl_async("GET", "/local/timer/current", nil, callback)
end

function M.timer_start(problem_id, callback)
	curl_async("POST", "/local/timer/start", { problem_id = problem_id or "" }, callback)
end

function M.timer_stop(callback)
	curl_async("POST", "/local/timer/stop", {}, callback)
end

function M.submit(problem_id, language, code, callback)
	curl_async("POST", "/local/submit", {
		problem_id = problem_id,
		language = language,
		code = code,
	}, callback)
end

function M.sync(callback)
	curl_async("POST", "/local/sync", {}, callback)
end

-- Poll verdict for a submission by ID.
function M.submission_get(id, callback)
	curl_async("GET", "/local/submissions/" .. id, nil, callback)
end

local function await_call(fn, ...)
	local args = pack(...)
	return async.await(function(callback)
		args.n = args.n + 1
		args[args.n] = callback
		fn(unpack(args, 1, args.n))
	end)
end

for _, name in ipairs({
	"status_today",
	"problems",
	"problem_get",
	"problem_suggest",
	"timer_current",
	"timer_start",
	"timer_stop",
	"submit",
	"sync",
	"submission_get",
}) do
	M[name .. "_await"] = function(...)
		return await_call(M[name], ...)
	end
end

return M
