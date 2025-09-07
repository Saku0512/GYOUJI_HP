import * as universal from '../entries/pages/admin/_page.js';

export const index = 3;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/admin/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/admin/+page.js";
export const imports = ["_app/immutable/nodes/3.Cjk0WZ2g.js","_app/immutable/chunks/CmSD4Lw8.js","_app/immutable/chunks/D2AUiy-P.js","_app/immutable/chunks/VVabkdHj.js","_app/immutable/chunks/JJD-QFX9.js","_app/immutable/chunks/CTSyBIan.js","_app/immutable/chunks/Cm9PB-7n.js","_app/immutable/chunks/B-6MdqlE.js","_app/immutable/chunks/BSkGyvYi.js","_app/immutable/chunks/D8uSMR3S.js","_app/immutable/chunks/BH2mRkq2.js","_app/immutable/chunks/4-skfxEB.js","_app/immutable/chunks/DhbCgPnK.js"];
export const stylesheets = ["_app/immutable/assets/AnimatedTransition.D47Mopqw.css","_app/immutable/assets/StaggeredList.CmVwnfCe.css","_app/immutable/assets/ErrorBoundary.C8UzbSjG.css","_app/immutable/assets/3.DS8eZ74e.css"];
export const fonts = [];
