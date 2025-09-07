import * as universal from '../entries/pages/admin/_page.js';

export const index = 3;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/admin/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/admin/+page.js";
export const imports = ["_app/immutable/nodes/3.BvKJEfpM.js","_app/immutable/chunks/Dp5EjsQM.js","_app/immutable/chunks/D2AUiy-P.js","_app/immutable/chunks/BRFgxvOq.js","_app/immutable/chunks/DiqqC9HZ.js","_app/immutable/chunks/CTSyBIan.js","_app/immutable/chunks/Cm9PB-7n.js","_app/immutable/chunks/BB0NFQV6.js","_app/immutable/chunks/Ce4CX5Aa.js","_app/immutable/chunks/CZAXbs8e.js","_app/immutable/chunks/B3wANwLS.js"];
export const stylesheets = ["_app/immutable/assets/StaggeredList.CmVwnfCe.css","_app/immutable/assets/AnimatedTransition.D47Mopqw.css","_app/immutable/assets/ErrorBoundary.C8UzbSjG.css","_app/immutable/assets/3.DS8eZ74e.css"];
export const fonts = [];
