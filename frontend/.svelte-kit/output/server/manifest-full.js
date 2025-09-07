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
		client: {start:"_app/immutable/entry/start.Bo3SUo_F.js",app:"_app/immutable/entry/app.DHvSQza2.js",imports:["_app/immutable/entry/start.Bo3SUo_F.js","_app/immutable/chunks/CRuF0lUP.js","_app/immutable/chunks/CuY-H75l.js","_app/immutable/chunks/D0iwhpLH.js","_app/immutable/entry/app.DHvSQza2.js","_app/immutable/chunks/CuY-H75l.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/DWiXyezG.js","_app/immutable/chunks/IxeuUGQt.js","_app/immutable/chunks/BKXpaS3L.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
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
