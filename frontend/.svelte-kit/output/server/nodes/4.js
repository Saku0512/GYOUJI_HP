import * as universal from '../entries/pages/login/_page.js';

export const index = 4;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/login/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/login/+page.js";
export const imports = ["_app/immutable/nodes/4.C7ZMP8xr.js","_app/immutable/chunks/BGq6rDGP.js","_app/immutable/chunks/CT_lE3Pl.js","_app/immutable/chunks/BqLxJNge.js","_app/immutable/chunks/DWufORZG.js","_app/immutable/chunks/DT5H3szw.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/DUpwANU7.js","_app/immutable/chunks/B7cxby2x.js","_app/immutable/chunks/CI-XvY1A.js","_app/immutable/chunks/B2k5sg_-.js"];
export const stylesheets = ["_app/immutable/assets/4.-o2-IrCg.css"];
export const fonts = [];
