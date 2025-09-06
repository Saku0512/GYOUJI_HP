import { T as head } from "../../../chunks/index2.js";
function _page($$payload) {
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Login - Tournament Management System</title>`;
  });
  $$payload.out.push(`<div class="login-container svelte-kn00yq"><h1 class="svelte-kn00yq">管理者ログイン</h1> <form><div class="form-group svelte-kn00yq"><label for="username" class="svelte-kn00yq">ユーザー名</label> <input type="text" id="username" name="username" required class="svelte-kn00yq"/></div> <div class="form-group svelte-kn00yq"><label for="password" class="svelte-kn00yq">パスワード</label> <input type="password" id="password" name="password" required class="svelte-kn00yq"/></div> <button type="submit" class="svelte-kn00yq">ログイン</button></form></div>`);
}
export {
  _page as default
};
