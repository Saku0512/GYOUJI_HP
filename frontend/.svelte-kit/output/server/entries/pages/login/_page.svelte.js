import { y as head } from "../../../chunks/index.js";
function _page($$payload) {
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Login - Tournament Management System</title>`;
  });
  $$payload.out.push(`<div class="login-container svelte-1bvl3aa"><h1 class="svelte-1bvl3aa">管理者ログイン</h1> <form><div class="form-group svelte-1bvl3aa"><label for="username" class="svelte-1bvl3aa">ユーザー名</label> <input type="text" id="username" name="username" required class="svelte-1bvl3aa"/></div> <div class="form-group svelte-1bvl3aa"><label for="password" class="svelte-1bvl3aa">パスワード</label> <input type="password" id="password" name="password" required class="svelte-1bvl3aa"/></div> <button type="submit" class="svelte-1bvl3aa">ログイン</button></form></div>`);
}
export {
  _page as default
};
