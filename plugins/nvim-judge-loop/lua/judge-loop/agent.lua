-- HTTP client for local-agent (127.0.0.1:7070).
-- Uses curl via vim.fn.jobstart (async, no UI blocking).
-- No external dependencies required.

local M = {}

local function get_url()
  return require("judge-loop").config.agent_url
end

local function curl_async(method, path, body, callback)
  local url = get_url() .. path
  local cmd = { "curl", "-s", "--max-time", "5", "-X", method, url, "-H", "Content-Type: application/json" }
  if body then
    table.insert(cmd, "-d")
    table.insert(cmd, vim.fn.json_encode(body))
  end

  local output = {}
  vim.fn.jobstart(cmd, {
    stdout_buffered = true,
    on_stdout = function(_, data)
      if data then
        for _, line in ipairs(data) do
          if line ~= "" then
            table.insert(output, line)
          end
        end
      end
    end,
    on_exit = function(_, code)
      if code ~= 0 then
        callback(false, nil)
        return
      end
      local raw = table.concat(output, "")
      if raw == "" then
        callback(false, nil)
        return
      end
      local ok, decoded = pcall(vim.fn.json_decode, raw)
      callback(ok, decoded)
    end,
  })
end

function M.status_today(callback)
  curl_async("GET", "/local/status/today", nil, callback)
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

return M
