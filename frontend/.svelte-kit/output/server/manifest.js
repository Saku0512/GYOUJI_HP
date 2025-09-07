export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set([]),
	mimeTypes: {},
	_: {
		client: {start:"_app/immutable/entry/start.CybkIFIy.js",app:"_app/immutable/entry/app.BIaFTpci.js",imports:["_app/immutable/entry/start.CybkIFIy.js","_app/immutable/chunks/BRsi0Z70.js","_app/immutable/chunks/VVabkdHj.js","_app/immutable/chunks/D2AUiy-P.js","_app/immutable/entry/app.BIaFTpci.js","_app/immutable/chunks/CTSyBIan.js","_app/immutable/chunks/VVabkdHj.js","_app/immutable/chunks/Cm9PB-7n.js","_app/immutable/chunks/BSkGyvYi.js","_app/immutable/chunks/4-skfxEB.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js')),
			__memo(() => import('./nodes/3.js')),
			__memo(() => import('./nodes/4.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/admin",
				pattern: /^\/admin\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/login",
				pattern: /^\/login\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
