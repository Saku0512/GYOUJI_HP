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
		client: {start:"_app/immutable/entry/start.8Wdus3zD.js",app:"_app/immutable/entry/app.B6wej8r9.js",imports:["_app/immutable/entry/start.8Wdus3zD.js","_app/immutable/chunks/BGq6rDGP.js","_app/immutable/chunks/CT_lE3Pl.js","_app/immutable/chunks/BqLxJNge.js","_app/immutable/entry/app.B6wej8r9.js","_app/immutable/chunks/BqLxJNge.js","_app/immutable/chunks/CT_lE3Pl.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/B7cxby2x.js","_app/immutable/chunks/DXDn204r.js","_app/immutable/chunks/CCKrlfp6.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
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
