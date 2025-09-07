import * as universal from '../entries/pages/admin/_page.js';

export const index = 3;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/admin/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/admin/+page.js";
export const imports = ["_app/immutable/nodes/3.nvbO9mXV.js","_app/immutable/chunks/DspYYv4h.js","_app/immutable/chunks/D0iwhpLH.js","_app/immutable/chunks/DnIs1vw8.js","_app/immutable/chunks/CuY-H75l.js","_app/immutable/chunks/CzS31Zl-.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/B1pWaKUj.js","_app/immutable/chunks/DWiXyezG.js","_app/immutable/chunks/CdAJyZzl.js","_app/immutable/chunks/BKXpaS3L.js","_app/immutable/chunks/DC-asg-G.js","_app/immutable/chunks/BjkiwGu8.js","_app/immutable/chunks/D43VgZms.js","_app/immutable/chunks/IxeuUGQt.js"];
export const stylesheets = ["_app/immutable/assets/LoadingSpinner.BxzunYK5.css","_app/immutable/assets/3.YvtEPsyZ.css"];
export const fonts = [];
