import { z as push, E as fallback, N as attr, F as attr_class, Q as clsx, O as slot, I as bind_props, B as pop, G as escape_html, M as ensure_array_like, T as maybe_selected, J as stringify, K as store_get, U as copy_payload, V as assign_payload, P as unsubscribe_stores, R as head } from "../../../chunks/index2.js";
import { o as onDestroy, t as tournamentStore } from "../../../chunks/tournament.js";
import { a as authStore } from "../../../chunks/auth.js";
import "clsx";
import "../../../chunks/client.js";
/* empty css                                                           */
function Button($$payload, $$props) {
  push();
  let classes;
  let variant = fallback($$props["variant"], "primary");
  let size = fallback($$props["size"], "medium");
  let disabled = fallback($$props["disabled"], false);
  let loading = fallback($$props["loading"], false);
  let type = fallback($$props["type"], "button");
  let href = fallback($$props["href"], null);
  let target = fallback($$props["target"], null);
  let fullWidth = fallback($$props["fullWidth"], false);
  let outline = fallback($$props["outline"], false);
  classes = [
    "btn",
    `btn-${variant}`,
    `btn-${size}`,
    outline ? "btn-outline" : "",
    fullWidth ? "btn-full-width" : "",
    disabled ? "btn-disabled" : "",
    loading ? "btn-loading" : ""
  ].filter(Boolean).join(" ");
  if (href) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<a${attr("href", href)}${attr("target", target)}${attr_class(clsx(classes), "svelte-qfafqz", { "disabled": disabled })} role="button"${attr("tabindex", disabled ? -1 : 0)}${attr("aria-disabled", disabled)}>`);
    if (loading) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="btn-spinner svelte-qfafqz" aria-hidden="true"></span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> <!---->`);
    slot($$payload, $$props, "default", {});
    $$payload.out.push(`<!----></a>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<button${attr("type", type)}${attr("disabled", disabled, true)}${attr_class(clsx(classes), "svelte-qfafqz")}${attr("aria-disabled", disabled)}>`);
    if (loading) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="btn-spinner svelte-qfafqz" aria-hidden="true"></span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> <!---->`);
    slot($$payload, $$props, "default", {});
    $$payload.out.push(`<!----></button>`);
  }
  $$payload.out.push(`<!--]-->`);
  bind_props($$props, {
    variant,
    size,
    disabled,
    loading,
    type,
    href,
    target,
    fullWidth,
    outline
  });
  pop();
}
function Select($$payload, $$props) {
  push();
  let selectId, validationState, classes;
  let value = fallback($$props["value"], "");
  let options = fallback($$props["options"], () => [], true);
  let placeholder = fallback($$props["placeholder"], "選択してください");
  let disabled = fallback($$props["disabled"], false);
  let required = fallback($$props["required"], false);
  let size = fallback($$props["size"], "medium");
  let variant = fallback($$props["variant"], "default");
  let label = fallback($$props["label"], "");
  let helperText = fallback($$props["helperText"], "");
  let errorMessage = fallback($$props["errorMessage"], "");
  let id = fallback($$props["id"], "");
  let name = fallback($$props["name"], "");
  let multiple = fallback($$props["multiple"], false);
  let fullWidth = fallback($$props["fullWidth"], false);
  function focus() {
  }
  function blur() {
  }
  selectId = id || `select-${Math.random().toString(36).substr(2, 9)}`;
  validationState = errorMessage ? "error" : variant;
  classes = [
    "select",
    `select-${size}`,
    `select-${validationState}`,
    fullWidth ? "select-full-width" : "",
    "",
    disabled ? "select-disabled" : ""
  ].filter(Boolean).join(" ");
  $$payload.out.push(`<div${attr_class("select-container svelte-170cbq6", void 0, { "select-container-full-width": fullWidth })}>`);
  if (label) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<label${attr("for", selectId)}${attr_class("select-label svelte-170cbq6", void 0, { "select-label-required": required })}>${escape_html(label)} `);
    if (required) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="select-required-mark svelte-170cbq6" aria-label="必須">*</span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></label>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <div class="select-wrapper svelte-170cbq6">`);
  if (multiple) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(options);
    $$payload.out.push(`<select${attr("id", selectId)}${attr("name", name)}${attr("disabled", disabled, true)}${attr("required", required, true)} multiple${attr_class(clsx(classes), "svelte-170cbq6")}${attr("aria-describedby", helperText || errorMessage ? `${selectId}-help` : void 0)}${attr("aria-invalid", errorMessage ? "true" : "false")}>`);
    $$payload.select_value = value;
    $$payload.out.push(`<!--[-->`);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let option = each_array[$$index];
      $$payload.out.push(`<option${attr("value", option.value)}${maybe_selected($$payload, option.value)}${attr("disabled", option.disabled || false, true)} class="svelte-170cbq6">${escape_html(option.label)}</option>`);
    }
    $$payload.out.push(`<!--]-->`);
    $$payload.select_value = void 0;
    $$payload.out.push(`</select>`);
  } else {
    $$payload.out.push("<!--[!-->");
    const each_array_1 = ensure_array_like(options);
    $$payload.out.push(`<select${attr("id", selectId)}${attr("name", name)}${attr("disabled", disabled, true)}${attr("required", required, true)}${attr_class(clsx(classes), "svelte-170cbq6")}${attr("aria-describedby", helperText || errorMessage ? `${selectId}-help` : void 0)}${attr("aria-invalid", errorMessage ? "true" : "false")}>`);
    $$payload.select_value = value;
    if (placeholder) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<option value=""${maybe_selected($$payload, "")} disabled${attr("selected", !value, true)} class="svelte-170cbq6">${escape_html(placeholder)}</option>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--><!--[-->`);
    for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
      let option = each_array_1[$$index_1];
      $$payload.out.push(`<option${attr("value", option.value)}${maybe_selected($$payload, option.value)}${attr("disabled", option.disabled || false, true)} class="svelte-170cbq6">${escape_html(option.label)}</option>`);
    }
    $$payload.out.push(`<!--]-->`);
    $$payload.select_value = void 0;
    $$payload.out.push(`</select>`);
  }
  $$payload.out.push(`<!--]--> `);
  if (!multiple) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="select-arrow svelte-170cbq6" aria-hidden="true"><svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd"></path></svg></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div> `);
  if (helperText && !errorMessage) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div${attr("id", `${stringify(selectId)}-help`)} class="select-helper-text svelte-170cbq6">${escape_html(helperText)}</div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (errorMessage) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div${attr("id", `${stringify(selectId)}-help`)} class="select-error-message svelte-170cbq6" role="alert">${escape_html(errorMessage)}</div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  bind_props($$props, {
    value,
    options,
    placeholder,
    disabled,
    required,
    size,
    variant,
    label,
    helperText,
    errorMessage,
    id,
    name,
    multiple,
    fullWidth,
    focus,
    blur
  });
  pop();
}
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  let currentSport = "volleyball";
  let pendingMatches = [];
  let availableFormats = [];
  let currentFormat = "";
  let isLoading = false;
  let isUpdatingFormat = false;
  let showMatchForm = false;
  const sportOptions = [
    { value: "volleyball", label: "バレーボール" },
    { value: "table_tennis", label: "卓球" },
    { value: "soccer", label: "8人制サッカー" }
  ];
  onDestroy(() => {
  });
  function getSportLabel(sport) {
    const option = sportOptions.find((opt) => opt.value === sport);
    return option ? option.label : sport;
  }
  function getStatusLabel(status) {
    const statusMap = {
      "pending": "未実施",
      "in_progress": "進行中",
      "completed": "完了",
      "cancelled": "キャンセル"
    };
    return statusMap[status] || status;
  }
  function getStatusClass(status) {
    const classMap = {
      "pending": "status-pending",
      "in_progress": "status-progress",
      "completed": "status-completed",
      "cancelled": "status-cancelled"
    };
    return classMap[status] || "status-default";
  }
  store_get($$store_subs ??= {}, "$authStore", authStore);
  store_get($$store_subs ??= {}, "$tournamentStore", tournamentStore);
  let $$settled = true;
  let $$inner_payload;
  function $$render_inner($$payload2) {
    head($$payload2, ($$payload3) => {
      $$payload3.title = `<title>管理者ダッシュボード - Tournament Management System</title>`;
    });
    $$payload2.out.push(`<div class="admin-container svelte-kc3jxr"><header class="admin-header svelte-kc3jxr"><div class="header-content svelte-kc3jxr"><h1 class="svelte-kc3jxr">管理者ダッシュボード</h1> <div class="header-actions svelte-kc3jxr">`);
    Button($$payload2, {
      variant: "outline",
      size: "small",
      disabled: isLoading,
      children: ($$payload3) => {
        $$payload3.out.push(`<!---->更新`);
      },
      $$slots: { default: true }
    });
    $$payload2.out.push(`<!----> `);
    Button($$payload2, {
      variant: "secondary",
      size: "small",
      children: ($$payload3) => {
        $$payload3.out.push(`<!---->ログアウト`);
      },
      $$slots: { default: true }
    });
    $$payload2.out.push(`<!----></div></div> <p class="header-description svelte-kc3jxr">試合結果の入力とトーナメント管理</p></header> <section class="control-panel svelte-kc3jxr"><div class="control-group svelte-kc3jxr"><label for="sport-select" class="svelte-kc3jxr">スポーツ選択</label> `);
    Select($$payload2, {
      id: "sport-select",
      options: sportOptions,
      disabled: isLoading,
      get value() {
        return currentSport;
      },
      set value($$value) {
        currentSport = $$value;
        $$settled = false;
      }
    });
    $$payload2.out.push(`<!----></div> `);
    if (availableFormats.length > 0) {
      $$payload2.out.push("<!--[-->");
      $$payload2.out.push(`<div class="control-group svelte-kc3jxr"><label for="format-select" class="svelte-kc3jxr">トーナメント形式 `);
      {
        $$payload2.out.push("<!--[!-->");
      }
      $$payload2.out.push(`<!--]--></label> `);
      Select($$payload2, {
        id: "format-select",
        options: availableFormats.map((format) => ({ value: format, label: format })),
        disabled: isUpdatingFormat,
        get value() {
          return currentFormat;
        },
        set value($$value) {
          currentFormat = $$value;
          $$settled = false;
        }
      });
      $$payload2.out.push(`<!----> `);
      {
        $$payload2.out.push("<!--[!-->");
      }
      $$payload2.out.push(`<!--]--></div>`);
    } else {
      $$payload2.out.push("<!--[!-->");
    }
    $$payload2.out.push(`<!--]--></section> <section class="matches-section"><div class="section-header svelte-kc3jxr"><h2 class="svelte-kc3jxr">${escape_html(getSportLabel(currentSport))}の未完了試合</h2> `);
    {
      $$payload2.out.push("<!--[!-->");
    }
    $$payload2.out.push(`<!--]--></div> `);
    {
      $$payload2.out.push("<!--[!-->");
      if (pendingMatches.length === 0) {
        $$payload2.out.push("<!--[-->");
        $$payload2.out.push(`<div class="empty-state svelte-kc3jxr"><p class="svelte-kc3jxr">未完了の試合はありません</p> `);
        Button($$payload2, {
          variant: "outline",
          children: ($$payload3) => {
            $$payload3.out.push(`<!---->データを更新`);
          },
          $$slots: { default: true }
        });
        $$payload2.out.push(`<!----></div>`);
      } else {
        $$payload2.out.push("<!--[!-->");
        const each_array = ensure_array_like(pendingMatches);
        $$payload2.out.push(`<div class="matches-grid svelte-kc3jxr"><!--[-->`);
        for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
          let match = each_array[$$index];
          $$payload2.out.push(`<div class="match-card svelte-kc3jxr"><div class="match-header svelte-kc3jxr"><span class="round-label svelte-kc3jxr">${escape_html(match.round)}</span> <span${attr_class(`status-badge ${stringify(getStatusClass(match.status))}`, "svelte-kc3jxr")}>${escape_html(getStatusLabel(match.status))}</span></div> <div class="match-teams svelte-kc3jxr"><div class="team svelte-kc3jxr"><span class="team-name svelte-kc3jxr">${escape_html(match.team1)}</span> `);
          if (match.score1 !== null) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<span class="team-score svelte-kc3jxr">${escape_html(match.score1)}</span>`);
          } else {
            $$payload2.out.push("<!--[!-->");
          }
          $$payload2.out.push(`<!--]--></div> <div class="vs-divider svelte-kc3jxr">vs</div> <div class="team svelte-kc3jxr"><span class="team-name svelte-kc3jxr">${escape_html(match.team2)}</span> `);
          if (match.score2 !== null) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<span class="team-score svelte-kc3jxr">${escape_html(match.score2)}</span>`);
          } else {
            $$payload2.out.push("<!--[!-->");
          }
          $$payload2.out.push(`<!--]--></div></div> `);
          if (match.winner) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<div class="winner-info svelte-kc3jxr"><span class="winner-label svelte-kc3jxr">勝者:</span> <span class="winner-name svelte-kc3jxr">${escape_html(match.winner)}</span></div>`);
          } else {
            $$payload2.out.push("<!--[!-->");
          }
          $$payload2.out.push(`<!--]--> `);
          if (match.scheduled_at) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<div class="schedule-info svelte-kc3jxr"><span class="schedule-label svelte-kc3jxr">予定:</span> <span class="schedule-time">${escape_html(new Date(match.scheduled_at).toLocaleString("ja-JP"))}</span></div>`);
          } else {
            $$payload2.out.push("<!--[!-->");
          }
          $$payload2.out.push(`<!--]--> <div class="match-actions svelte-kc3jxr">`);
          Button($$payload2, {
            variant: "primary",
            size: "small",
            disabled: showMatchForm,
            children: ($$payload3) => {
              $$payload3.out.push(`<!---->結果入力`);
            },
            $$slots: { default: true }
          });
          $$payload2.out.push(`<!----></div></div>`);
        }
        $$payload2.out.push(`<!--]--></div>`);
      }
      $$payload2.out.push(`<!--]-->`);
    }
    $$payload2.out.push(`<!--]--></section> `);
    {
      $$payload2.out.push("<!--[!-->");
    }
    $$payload2.out.push(`<!--]--></div>`);
  }
  do {
    $$settled = true;
    $$inner_payload = copy_payload($$payload);
    $$render_inner($$inner_payload);
  } while (!$$settled);
  assign_payload($$payload, $$inner_payload);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
