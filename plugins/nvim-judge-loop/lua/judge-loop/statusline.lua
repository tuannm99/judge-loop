-- Statusline component for lualine or a plain statusline.
--
-- Usage with lualine:
--   require("lualine").setup({
--     sections = {
--       lualine_x = { require("judge-loop.statusline").timer_statusline },
--     },
--   })
--
-- Usage in a plain statusline:
--   set statusline+=%{luaeval('require("judge-loop.statusline").timer_statusline()')}

local M = {}

-- Simple TTL cache: avoids firing a curl job on every statusline redraw.
local _cache = { text = "", updated_at = 0, pending = false }
local CACHE_TTL_SECS = 5

function M.timer_statusline()
	local now = os.time()
	if now - _cache.updated_at < CACHE_TTL_SECS or _cache.pending then
		return _cache.text
	end

	_cache.pending = true
	require("judge-loop.agent").timer_current(function(ok, data)
		_cache.pending = false
		_cache.updated_at = os.time()
		if not ok or not data or not data.active then
			_cache.text = ""
			return
		end
		local secs = data.elapsed_seconds or 0
		local mins = math.floor(secs / 60)
		_cache.text = string.format(" %d:%02d", mins, secs % 60)
	end)

	return _cache.text
end

return M
