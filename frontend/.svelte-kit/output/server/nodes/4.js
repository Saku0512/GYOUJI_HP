import * as universal from '../entries/pages/login/_page.js';

export const index = 4;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/login/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/login/+page.js";
export const imports = ["_app/immutable/nodes/4.CCCqRHis.js","_app/immutable/chunks/Dp5EjsQM.js","_app/immutable/chunks/D2AUiy-P.js","_app/immutable/chunks/BRFgxvOq.js","_app/immutable/chunks/DiqqC9HZ.js","_app/immutable/chunks/Cm9PB-7n.js","_app/immutable/chunks/BB0NFQV6.js","_app/immutable/chunks/BOLYuTPu.js","_app/immutable/chunks/BDipnjnj.js","_app/immutable/chunks/DoIqPyr3.js"];
export const stylesheets = ["_app/immutable/assets/4.C8CwgfNd.css"];
export const fonts = [];
