import { w as slot } from "../../chunks/index.js";
function _layout($$payload, $$props) {
  $$payload.out.push(`<main class="svelte-theyqk"><!---->`);
  slot($$payload, $$props, "default", {});
  $$payload.out.push(`<!----></main>`);
}
export {
  _layout as default
};
