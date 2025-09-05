import { y as head } from "../../../chunks/index.js";
function _page($$payload) {
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Admin Dashboard - Tournament Management System</title>`;
  });
  $$payload.out.push(`<div class="admin-container svelte-1ll51xp"><h1 class="svelte-1ll51xp">管理者ダッシュボード</h1> <p class="svelte-1ll51xp">試合結果の入力とトーナメント管理</p></div>`);
}
export {
  _page as default
};
