

export const index = 3;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/admin/_page.svelte.js')).default;
export const imports = ["_app/immutable/nodes/3.CscQ1e2m.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/DUpwANU7.js","_app/immutable/chunks/BqLxJNge.js"];
export const stylesheets = ["_app/immutable/assets/3.hoyyCCFb.css"];
export const fonts = [];
