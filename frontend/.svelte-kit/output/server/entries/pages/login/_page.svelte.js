import { R as head, F as attr_class, N as attr, G as escape_html, B as pop, z as push } from "../../../chunks/index2.js";
import "@sveltejs/kit/internal";
import "../../../chunks/exports.js";
import "../../../chunks/utils.js";
import "../../../chunks/state.svelte.js";
import { a as authStore } from "../../../chunks/auth.js";
function _page($$payload, $$props) {
  push();
  let formData = { username: "", password: "" };
  let validationErrors = {};
  let isSubmitting = false;
  let authState = {};
  authStore.subscribe((state) => {
    authState = state;
  });
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>管理者ログイン - Tournament Management System</title>`;
  });
  $$payload.out.push(`<div class="login-container svelte-1pxb3um"><div class="login-card svelte-1pxb3um"><h1 class="svelte-1pxb3um">管理者ログイン</h1> <p class="login-description svelte-1pxb3um">トーナメント管理システムの管理者ダッシュボードにアクセスするには、認証情報を入力してください。</p> <form novalidate class="svelte-1pxb3um"><div class="form-group svelte-1pxb3um"><label for="username"${attr_class("svelte-1pxb3um", void 0, { "error": validationErrors.username })}>ユーザー名</label> <input type="text" id="username" name="username"${attr("value", formData.username)}${attr("disabled", isSubmitting, true)} autocomplete="username" data-testid="username" required${attr_class("svelte-1pxb3um", void 0, { "error": validationErrors.username })}/> `);
  if (validationErrors.username) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="error-message svelte-1pxb3um">${escape_html(validationErrors.username)}</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div> <div class="form-group svelte-1pxb3um"><label for="password"${attr_class("svelte-1pxb3um", void 0, { "error": validationErrors.password })}>パスワード</label> <div class="password-input-container svelte-1pxb3um"><input${attr("type", "password")} id="password" name="password"${attr("value", formData.password)}${attr("disabled", isSubmitting, true)} autocomplete="current-password" data-testid="password" required${attr_class("svelte-1pxb3um", void 0, { "error": validationErrors.password })}/> <button type="button" class="password-toggle svelte-1pxb3um"${attr("disabled", isSubmitting, true)}${attr("aria-label", "パスワードを表示")}>${escape_html("👁️")}</button></div> `);
  if (validationErrors.password) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="error-message svelte-1pxb3um">${escape_html(validationErrors.password)}</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div> <button type="submit" class="login-button svelte-1pxb3um"${attr("disabled", authState.loading, true)} data-testid="login-button">`);
  if (authState.loading) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="loading-spinner svelte-1pxb3um"></span> ログイン中...`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`ログイン`);
  }
  $$payload.out.push(`<!--]--></button></form> `);
  if (authState.loading) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="loading-overlay svelte-1pxb3um"><div class="loading-content svelte-1pxb3um"><span class="loading-spinner large svelte-1pxb3um"></span> <p class="svelte-1pxb3um">認証中...</p></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></div>`);
  pop();
}
export {
  _page as default
};
