import * as universal from '../entries/pages/login/_page.js';

export const index = 4;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/login/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/login/+page.js";
export const imports = ["_app/immutable/nodes/4.B6RydUJ2.js","_app/immutable/chunks/CmSD4Lw8.js","_app/immutable/chunks/D2AUiy-P.js","_app/immutable/chunks/VVabkdHj.js","_app/immutable/chunks/JJD-QFX9.js","_app/immutable/chunks/Cm9PB-7n.js","_app/immutable/chunks/B-6MdqlE.js","_app/immutable/chunks/BSkGyvYi.js","_app/immutable/chunks/BCfgoF3A.js","_app/immutable/chunks/BRsi0Z70.js","_app/immutable/chunks/KE2xTBgs.js","_app/immutable/chunks/DoIqPyr3.js"];
export const stylesheets = ["_app/immutable/assets/4.C8CwgfNd.css"];
export const fonts = [];
