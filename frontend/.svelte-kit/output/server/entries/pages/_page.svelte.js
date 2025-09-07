import { E as fallback, F as attr_class, Q as clsx, G as escape_html, I as bind_props, B as pop, z as push, M as ensure_array_like, N as attr, J as stringify, K as store_get, R as head, P as unsubscribe_stores } from "../../chunks/index2.js";
import { o as onDestroy, t as tournamentStore } from "../../chunks/tournament.js";
import { L as LoadingSpinner, u as uiStore, a as uiActions } from "../../chunks/LoadingSpinner.js";
function formatMatchStatus(status) {
  const statusNames = {
    pending: "未実施",
    in_progress: "進行中",
    completed: "完了",
    cancelled: "キャンセル"
  };
  return statusNames[status] || status;
}
function formatScore(score1, score2) {
  if (score1 === null || score1 === void 0 || score2 === null || score2 === void 0) {
    return "未実施";
  }
  return `${score1} - ${score2}`;
}
function formatTeamName(teamName, maxLength = 10) {
  if (!teamName) return "";
  if (teamName.length <= maxLength) {
    return teamName;
  }
  return teamName.substring(0, maxLength - 3) + "...";
}
function MatchCard($$payload, $$props) {
  push();
  let winner, cardClasses, team1Classes, team2Classes;
  let match = fallback($$props["match"], () => ({}), true);
  let editable = fallback($$props["editable"], false);
  let compact = fallback($$props["compact"], false);
  match.score1 || "";
  match.score2 || "";
  function getWinner(match2) {
    if (match2.winner) return match2.winner;
    if (match2.score1 !== null && match2.score2 !== null && match2.score1 !== void 0 && match2.score2 !== void 0) {
      if (match2.score1 > match2.score2) return match2.team1;
      if (match2.score2 > match2.score1) return match2.team2;
      return "draw";
    }
    return null;
  }
  winner = getWinner(match);
  cardClasses = [
    "match-card",
    compact ? "compact" : "",
    editable ? "editable" : "",
    ""
  ].filter(Boolean).join(" ");
  team1Classes = ["team", winner === match.team1 ? "winner" : ""].filter(Boolean).join(" ");
  team2Classes = ["team", winner === match.team2 ? "winner" : ""].filter(Boolean).join(" ");
  $$payload.out.push(`<div${attr_class(clsx(cardClasses), "svelte-l0munw")} data-testid="match-card"><div class="match-header svelte-l0munw">`);
  if (match.round) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="round svelte-l0munw" data-testid="match-round">${escape_html(match.round)}</div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (match.scheduled_at) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="schedule svelte-l0munw" data-testid="match-schedule">${escape_html(new Date(match.scheduled_at).toLocaleDateString("ja-JP", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit"
    }))}</div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div> <div class="teams svelte-l0munw"><div${attr_class(clsx(team1Classes), "svelte-l0munw")} data-testid="team1"><span class="team-name svelte-l0munw">${escape_html(formatTeamName(match.team1 || "Team A", compact ? 8 : 15))}</span> `);
  if (winner === match.team1) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="winner-badge svelte-l0munw">🏆</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div> <div class="vs-section svelte-l0munw"><span class="vs svelte-l0munw">vs</span></div> <div${attr_class(clsx(team2Classes), "svelte-l0munw")} data-testid="team2">`);
  if (winner === match.team2) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="winner-badge svelte-l0munw">🏆</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <span class="team-name svelte-l0munw">${escape_html(formatTeamName(match.team2 || "Team B", compact ? 8 : 15))}</span></div></div> <div class="score-section svelte-l0munw">`);
  {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div class="score-display svelte-l0munw" data-testid="score-display">`);
    if (match.score1 !== null && match.score1 !== void 0 && match.score2 !== null && match.score2 !== void 0) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="score svelte-l0munw">${escape_html(formatScore(match.score1, match.score2))}</div>`);
    } else {
      $$payload.out.push("<!--[!-->");
      $$payload.out.push(`<div class="status svelte-l0munw" data-testid="match-status">${escape_html(formatMatchStatus(match.status || "pending"))}</div>`);
    }
    $$payload.out.push(`<!--]--></div> `);
    if (editable && match.status !== "completed") {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<button class="edit-btn svelte-l0munw" data-testid="edit-match-btn" aria-label="試合結果を編集">✏️ 編集</button>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]--></div> `);
  if (!compact && (match.completed_at || match.status === "completed")) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="match-details svelte-l0munw"><div class="completion-time svelte-l0munw" data-testid="completion-time">完了: ${escape_html(new Date(match.completed_at).toLocaleDateString("ja-JP", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit"
    }))}</div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  bind_props($$props, { match, editable, compact });
  pop();
}
function TournamentBracket($$payload, $$props) {
  push();
  let groupedMatches, roundOrder;
  let sport = fallback($$props["sport"], "volleyball");
  let matches = fallback($$props["matches"], () => [], true);
  let isAdmin = fallback($$props["isAdmin"], false);
  let onEditMatch = fallback($$props["onEditMatch"], null);
  const sportNames = { volleyball: "バレーボール", table_tennis: "卓球", soccer: "サッカー" };
  const roundNames = {
    "round_1": "1回戦",
    "round_2": "2回戦",
    "quarterfinal": "準々決勝",
    "semifinal": "準決勝",
    "final": "決勝",
    "third_place": "3位決定戦"
  };
  function groupMatchesByRound(matches2) {
    if (!Array.isArray(matches2)) {
      return {};
    }
    return matches2.reduce(
      (groups, match) => {
        if (!match || typeof match !== "object") {
          return groups;
        }
        const round = match.round || "unknown";
        if (!groups[round]) {
          groups[round] = [];
        }
        groups[round].push(match);
        return groups;
      },
      {}
    );
  }
  function getRoundOrder(groupedMatches2) {
    const availableRounds = Object.keys(groupedMatches2);
    const standardOrder = [
      "round_1",
      "round_2",
      "quarterfinal",
      "semifinal",
      "third_place",
      "final"
    ];
    return standardOrder.filter((round) => availableRounds.includes(round)).concat(availableRounds.filter((round) => !standardOrder.includes(round)));
  }
  function getMatchWinner(match) {
    if (match.winner) {
      return match.winner;
    }
    if (match.score1 !== void 0 && match.score2 !== void 0) {
      if (match.score1 > match.score2) {
        return match.team1;
      } else if (match.score2 > match.score1) {
        return match.team2;
      }
    }
    return null;
  }
  function getMatchStatus(match) {
    if (match.status === "completed" || match.score1 !== void 0 && match.score2 !== void 0) {
      return "completed";
    } else if (match.status === "in_progress") {
      return "in_progress";
    } else {
      return "pending";
    }
  }
  function getResponsiveClass() {
    if (typeof window !== "undefined") {
      const width = window.innerWidth;
      if (width < 768) return "mobile";
      if (width < 1024) return "tablet";
      return "desktop";
    }
    return "desktop";
  }
  let responsiveClass = "desktop";
  if (typeof window !== "undefined") {
    responsiveClass = getResponsiveClass();
    window.addEventListener("resize", () => {
      responsiveClass = getResponsiveClass();
    });
  }
  groupedMatches = groupMatchesByRound(matches);
  roundOrder = getRoundOrder(groupedMatches);
  $$payload.out.push(`<div${attr_class(`tournament-bracket ${stringify(responsiveClass)}`, "svelte-10kswq8")}><div class="bracket-header svelte-10kswq8"><h2 class="tournament-title svelte-10kswq8">${escape_html(sportNames[sport] || sport)} トーナメント</h2> `);
  if (matches.length === 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<p class="no-matches svelte-10kswq8">試合データがありません</p>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div> `);
  if (matches.length > 0) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(roundOrder);
    $$payload.out.push(`<div class="bracket-container svelte-10kswq8"><!--[-->`);
    for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
      let round = each_array[$$index_1];
      const each_array_1 = ensure_array_like(groupedMatches[round]);
      $$payload.out.push(`<div class="round-column svelte-10kswq8"${attr("data-round", round)}><h3 class="round-title svelte-10kswq8">${escape_html(roundNames[round] || round)}</h3> <div class="matches-container svelte-10kswq8"><!--[-->`);
      for (let $$index = 0, $$length2 = each_array_1.length; $$index < $$length2; $$index++) {
        let match = each_array_1[$$index];
        $$payload.out.push(`<div class="match-wrapper svelte-10kswq8"${attr("data-status", getMatchStatus(match))}><div class="match-card-container svelte-10kswq8">`);
        MatchCard($$payload, { match, editable: isAdmin });
        $$payload.out.push(`<!----> `);
        if (getMatchWinner(match)) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<div class="winner-indicator svelte-10kswq8">勝者: ${escape_html(getMatchWinner(match))}</div>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--> `);
        if (isAdmin && getMatchStatus(match) === "pending") {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<button class="edit-button svelte-10kswq8" aria-label="試合結果を編集">結果入力</button>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--></div> `);
        if (round !== "final" && round !== "third_place") {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<div class="connection-line svelte-10kswq8" aria-hidden="true"></div>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--></div>`);
      }
      $$payload.out.push(`<!--]--></div></div>`);
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  bind_props($$props, { sport, matches, isAdmin, onEditMatch });
  pop();
}
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  let tournament, ui, currentTournamentData, matches;
  const sports = [
    { key: "volleyball", name: "バレーボール", icon: "🏐" },
    { key: "table_tennis", name: "卓球", icon: "🏓" },
    { key: "soccer", name: "サッカー", icon: "⚽" }
  ];
  onDestroy(() => {
  });
  async function loadTournamentData(sport, showLoading = true) {
    try {
      if (showLoading) {
        uiActions.setLoading(true);
      }
      const result = await tournamentStore.fetchTournaments(sport);
      if (!result.success) {
        uiActions.showNotification(result.message || "データの取得に失敗しました", "error");
      }
      return result;
    } catch (error) {
      console.error("Load tournament data error:", error);
      uiActions.showNotification("データの読み込みでエラーが発生しました", "error");
      return { success: false, message: error.message };
    } finally {
      if (showLoading) {
        uiActions.setLoading(false);
      }
    }
  }
  function getSportName(sportKey) {
    const sport = sports.find((s) => s.key === sportKey);
    return sport ? sport.name : sportKey;
  }
  function handleVisibilityChange() {
    if (typeof document !== "undefined") {
      if (!document.hidden) {
        loadTournamentData(tournament.currentSport, false);
      }
    }
  }
  if (typeof document !== "undefined") {
    document.addEventListener("visibilitychange", handleVisibilityChange);
  }
  tournament = store_get($$store_subs ??= {}, "$tournamentStore", tournamentStore);
  ui = store_get($$store_subs ??= {}, "$uiStore", uiStore);
  currentTournamentData = tournament.tournaments[tournament.currentSport] || null;
  matches = currentTournamentData?.matches || [];
  const each_array = ensure_array_like(sports);
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>トーナメント管理システム - ${escape_html(getSportName(tournament.currentSport))}</title>`;
    $$payload2.out.push(`<meta name="description" content="バレーボール、卓球、サッカーのトーナメント結果をリアルタイムで確認できます"/>`);
  });
  $$payload.out.push(`<div class="homepage svelte-srgitj"><div class="container svelte-srgitj"><div class="page-header svelte-srgitj"><h1 class="page-title svelte-srgitj">トーナメント管理システム</h1> <p class="page-description svelte-srgitj">リアルタイムでトーナメントの進行状況を確認できます</p> <div class="header-actions svelte-srgitj"><button class="refresh-button svelte-srgitj"${attr("disabled", tournament.loading || ui.loading, true)} aria-label="データを更新"><span${attr_class("refresh-icon svelte-srgitj", void 0, { "spinning": tournament.loading || ui.loading })}>🔄</span> 更新</button> `);
  if (tournament.lastUpdated) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="last-updated svelte-srgitj">最終更新: ${escape_html(new Date(tournament.lastUpdated).toLocaleTimeString("ja-JP"))}</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></div> <div class="sports-tabs svelte-srgitj"><div class="tabs-container svelte-srgitj"><!--[-->`);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let sport = each_array[$$index];
    $$payload.out.push(`<button${attr_class("sport-tab svelte-srgitj", void 0, { "active": tournament.currentSport === sport.key })}${attr("disabled", tournament.loading, true)}${attr("aria-label", `${stringify(sport.name)}のトーナメントを表示`)}><span class="sport-icon svelte-srgitj">${escape_html(sport.icon)}</span> <span class="sport-name svelte-srgitj">${escape_html(sport.name)}</span></button>`);
  }
  $$payload.out.push(`<!--]--></div></div> <div class="main-content svelte-srgitj">`);
  if (tournament.error) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="error-container svelte-srgitj"><div class="error-message svelte-srgitj"><h3 class="svelte-srgitj">エラーが発生しました</h3> <p class="svelte-srgitj">${escape_html(tournament.error)}</p> <button class="retry-button svelte-srgitj">再試行</button></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    if (tournament.loading && !currentTournamentData) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="loading-container svelte-srgitj">`);
      LoadingSpinner($$payload, { size: "large" });
      $$payload.out.push(`<!----> <p class="loading-text svelte-srgitj">トーナメントデータを読み込み中...</p></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
      if (currentTournamentData && matches.length > 0) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<div class="tournament-container svelte-srgitj">`);
        TournamentBracket($$payload, { sport: tournament.currentSport, matches, isAdmin: false });
        $$payload.out.push(`<!----></div>`);
      } else {
        $$payload.out.push("<!--[!-->");
        $$payload.out.push(`<div class="no-data-container svelte-srgitj"><div class="no-data-message svelte-srgitj"><h3 class="svelte-srgitj">トーナメントデータがありません</h3> <p class="svelte-srgitj">${escape_html(getSportName(tournament.currentSport))}のトーナメントはまだ開始されていません。</p> <button class="refresh-button svelte-srgitj">データを確認</button></div></div>`);
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]--></div></div></div>`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
