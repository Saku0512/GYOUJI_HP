import { redirect } from "@sveltejs/kit";
import { g as getAuthToken } from "../../../chunks/storage.js";
async function load({ url }) {
  if (typeof window !== "undefined") {
    const token = getAuthToken();
    if (token) {
      const redirectTo = url.searchParams.get("redirect") || "/admin";
      throw redirect(302, redirectTo);
    }
  }
  return {};
}
export {
  load
};
