import { D as getContext, E as fallback, F as attr_class, G as escape_html, I as bind_props, B as pop, z as push, J as stringify, K as store_get, M as ensure_array_like, N as attr, O as slot, P as unsubscribe_stores } from "../../chunks/index2.js";
import "@sveltejs/kit/internal";
import "../../chunks/exports.js";
import "../../chunks/utils.js";
import "clsx";
import "../../chunks/state.svelte.js";
import { a as authStore } from "../../chunks/auth.js";
import { L as LoadingSpinner, u as uiStore } from "../../chunks/LoadingSpinner.js";
const getStores = () => {
  const stores$1 = getContext("__svelte__");
  return {
    /** @type {typeof page} */
    page: {
      subscribe: stores$1.page.subscribe
    },
    /** @type {typeof navigating} */
    navigating: {
      subscribe: stores$1.navigating.subscribe
    },
    /** @type {typeof updated} */
    updated: stores$1.updated
  };
};
const page = {
  subscribe(fn) {
    const store = getStores().page;
    return store.subscribe(fn);
  }
};
function NotificationToast($$payload, $$props) {
  push();
  let message = fallback($$props["message"], "");
  let type = fallback(
    $$props["type"],
    "info"
    // 'success', 'error', 'warning', 'info'
  );
  let duration = fallback(
    $$props["duration"],
    5e3
    // 自動消去時間（ミリ秒）
  );
  let dismissible = fallback(
    $$props["dismissible"],
    true
    // 手動で閉じることができるか
  );
  let visible = true;
  let timeoutId;
  if (duration > 0) {
    timeoutId = setTimeout(
      () => {
        close();
      },
      duration
    );
  }
  function close() {
    visible = false;
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
  }
  if (visible) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div${attr_class(`toast toast-${stringify(type)}`, "svelte-trunga")} role="alert" aria-live="polite" aria-atomic="true"><div class="toast-content svelte-trunga"><div class="toast-icon svelte-trunga">`);
    if (type === "success") {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path></svg>`);
    } else {
      $$payload.out.push("<!--[!-->");
      if (type === "error") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path></svg>`);
      } else {
        $$payload.out.push("<!--[!-->");
        if (type === "warning") {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path></svg>`);
        } else {
          $$payload.out.push("<!--[!-->");
          $$payload.out.push(`<svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path></svg>`);
        }
        $$payload.out.push(`<!--]-->`);
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]--></div> <div class="toast-message svelte-trunga">${escape_html(message)}</div> `);
    if (dismissible) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<button class="toast-close svelte-trunga" aria-label="通知を閉じる" type="button"><svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path></svg></button>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  bind_props($$props, { message, type, duration, dismissible });
  pop();
}
function _layout($$payload, $$props) {
  push();
  var $$store_subs;
  let auth, ui;
  let mobileMenuOpen = false;
  function isActivePage(path) {
    return store_get($$store_subs ??= {}, "$page", page).url.pathname === path;
  }
  auth = store_get($$store_subs ??= {}, "$authStore", authStore);
  ui = store_get($$store_subs ??= {}, "$uiStore", uiStore);
  const each_array = ensure_array_like(ui.notifications);
  $$payload.out.push(`<div class="app-layout svelte-ikcr39"><header class="header svelte-ikcr39"><div class="container"><nav class="navbar svelte-ikcr39"><div class="navbar-brand svelte-ikcr39"><a href="/" class="brand-link svelte-ikcr39"><h1 class="brand-title svelte-ikcr39">トーナメント管理</h1></a></div> <div class="navbar-nav desktop-nav svelte-ikcr39"><a href="/"${attr_class("nav-link svelte-ikcr39", void 0, { "active": isActivePage("/") })}>ホーム</a> `);
  if (auth.isAuthenticated) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<a href="/admin"${attr_class("nav-link svelte-ikcr39", void 0, { "active": isActivePage("/admin") })}>管理ダッシュボード</a> <button class="nav-button logout-button svelte-ikcr39"${attr("disabled", auth.loading, true)}>ログアウト</button>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<a href="/login"${attr_class("nav-link login-link svelte-ikcr39", void 0, { "active": isActivePage("/login") })}>管理者ログイン</a>`);
  }
  $$payload.out.push(`<!--]--></div> <button class="mobile-menu-button svelte-ikcr39" aria-label="メニューを開く"${attr("aria-expanded", mobileMenuOpen)}><span${attr_class("hamburger-line svelte-ikcr39", void 0, { "open": mobileMenuOpen })}></span> <span${attr_class("hamburger-line svelte-ikcr39", void 0, { "open": mobileMenuOpen })}></span> <span${attr_class("hamburger-line svelte-ikcr39", void 0, { "open": mobileMenuOpen })}></span></button></nav></div> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></header> <main class="main-content svelte-ikcr39"><!---->`);
  slot($$payload, $$props, "default", {});
  $$payload.out.push(`<!----></main> <footer class="footer svelte-ikcr39"><div class="container"><div class="footer-content svelte-ikcr39"><div class="footer-section svelte-ikcr39"><h3 class="footer-title svelte-ikcr39">トーナメント管理システム</h3> <p class="footer-description svelte-ikcr39">バレーボール、卓球、8人制サッカーのトーナメント管理</p></div> <div class="footer-section svelte-ikcr39"><h4 class="footer-subtitle svelte-ikcr39">リンク</h4> <ul class="footer-links svelte-ikcr39"><li class="svelte-ikcr39"><a href="/" class="svelte-ikcr39">ホーム</a></li> `);
  if (auth.isAuthenticated) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<li class="svelte-ikcr39"><a href="/admin" class="svelte-ikcr39">管理ダッシュボード</a></li>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<li class="svelte-ikcr39"><a href="/login" class="svelte-ikcr39">管理者ログイン</a></li>`);
  }
  $$payload.out.push(`<!--]--></ul></div></div> <div class="footer-bottom svelte-ikcr39"><p>© 2024 トーナメント管理システム. All rights reserved.</p></div></div></footer> `);
  if (ui.loading || auth.loading) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="loading-overlay svelte-ikcr39"><div class="loading-content svelte-ikcr39">`);
    LoadingSpinner($$payload, { size: "large" });
    $$payload.out.push(`<!----> <p class="loading-text svelte-ikcr39">処理中...</p></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <div class="notifications-container svelte-ikcr39"><!--[-->`);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let notification = each_array[$$index];
    NotificationToast($$payload, {
      message: notification.message,
      type: notification.type,
      duration: 0
    });
  }
  $$payload.out.push(`<!--]--></div></div>`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _layout as default
};
