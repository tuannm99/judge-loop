local M = {}

local levels = vim.log.levels

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

local verdict_labels = {
  accepted = "✓ Accepted",
  wrong_answer = "✗ Wrong Answer",
  time_limit_exceeded = "⏱ Time Limit Exceeded",
  runtime_error = "💥 Runtime Error",
  compile_error = "🔨 Compile Error",
  pending = "⏳ Pending…",
}

function M.verdict(status, submission_id)
  local label = verdict_labels[status] or status
  local msg = string.format("%s  (id: %s)", label, submission_id or "?")
  local level = (status == "accepted") and levels.INFO or levels.WARN
  vim.schedule(function()
    vim.notify("[judge-loop] " .. msg, level)
  end)
end

return M
