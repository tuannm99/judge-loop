-- judge-loop Neovim plugin entry point.
-- Loaded automatically by Neovim when the plugin is installed.

if vim.g.loaded_judge_loop then
  return
end
vim.g.loaded_judge_loop = true

require("judge-loop").setup()
