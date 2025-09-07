import { redirect } from "@sveltejs/kit";
import { g as getAuthToken } from "../../../chunks/storage.js";
async function load({ url, fetch }) {
  if (typeof window !== "undefined") {
    const token = getAuthToken();
    if (!token) {
      throw redirect(302, "/login?redirect=" + encodeURIComponent(url.pathname));
    }
    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      const currentTime = Math.floor(Date.now() / 1e3);
      if (payload.exp && payload.exp < currentTime) {
        throw redirect(302, "/login?expired=true&redirect=" + encodeURIComponent(url.pathname));
      }
    } catch (error) {
      console.error("Token validation error:", error);
      throw redirect(302, "/login?invalid=true&redirect=" + encodeURIComponent(url.pathname));
    }
  }
  return {
    // ページに必要な初期データがあればここで取得
    title: "管理者ダッシュボード"
  };
}
export {
  load
};
