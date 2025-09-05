

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_layout.svelte.js')).default;
export const imports = ["_app/immutable/nodes/0.DNc5Akbx.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/DKpEfF5J.js","_app/immutable/chunks/BkOiC1l7.js"];
export const stylesheets = ["_app/immutable/assets/0.CEgD6ITX.css"];
export const fonts = [];
