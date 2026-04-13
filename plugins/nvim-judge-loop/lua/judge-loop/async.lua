local M = {}

local pack = table.pack or function(...)
	return { n = select("#", ...), ... }
end
local unpack = table.unpack or unpack

local function report_error(thread, err)
	local message = tostring(err)
	if debug and debug.traceback then
		message = debug.traceback(thread, message)
	end
	vim.schedule(function()
		error(message)
	end)
end

local function resume_thread(thread, ...)
	local ok, err = coroutine.resume(thread, ...)
	if not ok then
		report_error(thread, err)
	end
end

function M.run(fn)
	local thread = coroutine.create(fn)
	resume_thread(thread)
end

function M.await(register)
	local thread = coroutine.running()
	assert(thread, "judge-loop.async.await() must be called inside judge-loop.async.run()")

	local done = false
	local waiting = false
	local values = nil

	local function resolve(...)
		if done then
			return
		end
		done = true
		values = pack(...)
		if waiting then
			waiting = false
			resume_thread(thread, unpack(values, 1, values.n))
		end
	end

	register(resolve)

	if done then
		return unpack(values, 1, values.n)
	end

	waiting = true
	return coroutine.yield()
end

function M.sleep(ms)
	return M.await(function(resolve)
		vim.defer_fn(resolve, ms)
	end)
end

return M
