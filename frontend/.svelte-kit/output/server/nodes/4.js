import * as universal from '../entries/pages/login/_page.js';

export const index = 4;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/login/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/login/+page.js";
export const imports = ["_app/immutable/nodes/4.Ccz_Og3Y.js","_app/immutable/chunks/DspYYv4h.js","_app/immutable/chunks/D0iwhpLH.js","_app/immutable/chunks/DnIs1vw8.js","_app/immutable/chunks/CuY-H75l.js","_app/immutable/chunks/CzS31Zl-.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/B1pWaKUj.js","_app/immutable/chunks/DWiXyezG.js","_app/immutable/chunks/BjkiwGu8.js","_app/immutable/chunks/CRuF0lUP.js"];
export const stylesheets = ["_app/immutable/assets/4.B6BljH5-.css"];
export const fonts = [];
