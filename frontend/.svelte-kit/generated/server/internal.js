
import root from '../root.js';
import { set_building, set_prerendering } from '__sveltekit/environment';
import { set_assets } from '__sveltekit/paths';
import { set_manifest, set_read_implementation } from '__sveltekit/server';
import { set_private_env, set_public_env } from '../../../node_modules/@sveltejs/kit/src/runtime/shared-server.js';

export const options = {
	app_template_contains_nonce: false,
	csp: {"mode":"auto","directives":{"upgrade-insecure-requests":false,"block-all-mixed-content":false},"reportOnly":{"upgrade-insecure-requests":false,"block-all-mixed-content":false}},
	csrf_check_origin: true,
	csrf_trusted_origins: [],
	embedded: false,
	env_public_prefix: 'PUBLIC_',
	env_private_prefix: '',
	hash_routing: false,
	hooks: null, // added lazily, via `get_hooks`
	preload_strategy: "modulepreload",
	root,
	service_worker: false,
	service_worker_options: undefined,
	templates: {
		app: ({ head, body, assets, nonce, env }) => "<!doctype html>\n<html lang=\"ja\">\n  <head>\n    <meta charset=\"UTF-8\" />\n    <link rel=\"icon\" type=\"image/svg+xml\" href=\"" + assets + "/icon.svg\" />\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\" />\n    <meta\n      name=\"description\"\n      content=\"トーナメント管理システム - バレーボール、卓球、サッカーのトーナメント管理\"\n    />\n    \n    <!-- PWA Manifest -->\n    <link rel=\"manifest\" href=\"/manifest.json\" />\n    \n    <!-- Theme Color -->\n    <meta name=\"theme-color\" content=\"#007bff\" />\n    <meta name=\"msapplication-TileColor\" content=\"#007bff\" />\n    \n    <!-- Apple Touch Icons -->\n    <link rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/icon-192x192.png\" />\n    <meta name=\"apple-mobile-web-app-capable\" content=\"yes\" />\n    <meta name=\"apple-mobile-web-app-status-bar-style\" content=\"default\" />\n    <meta name=\"apple-mobile-web-app-title\" content=\"Tournament\" />\n    \n    <!-- Resource Hints -->\n    <link rel=\"dns-prefetch\" href=\"//fonts.googleapis.com\" />\n    <link rel=\"preconnect\" href=\"//fonts.gstatic.com\" crossorigin />\n    \n    <!-- Critical CSS inlined here for better performance -->\n    <style>\n      /* Critical CSS for initial render */\n      body {\n        margin: 0;\n        padding: 0;\n        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;\n        background-color: #f8f9fa;\n        color: #333;\n      }\n      \n      /* Loading spinner for initial load */\n      .initial-loading {\n        position: fixed;\n        top: 0;\n        left: 0;\n        width: 100%;\n        height: 100%;\n        background: #f8f9fa;\n        display: flex;\n        align-items: center;\n        justify-content: center;\n        z-index: 9999;\n      }\n      \n      .initial-loading .spinner {\n        width: 40px;\n        height: 40px;\n        border: 4px solid #e0e0e0;\n        border-top: 4px solid #007bff;\n        border-radius: 50%;\n        animation: spin 1s linear infinite;\n      }\n      \n      @keyframes spin {\n        0% { transform: rotate(0deg); }\n        100% { transform: rotate(360deg); }\n      }\n      \n      /* Hide loading when app is ready */\n      .app-ready .initial-loading {\n        display: none;\n      }\n    </style>\n    \n    <title>Tournament Management System</title>\n    " + head + "\n  </head>\n\n  <body data-sveltekit-preload-data=\"hover\">\n    <!-- Initial loading screen -->\n    <div class=\"initial-loading\">\n      <div class=\"spinner\"></div>\n    </div>\n    \n    <div style=\"display: contents\">" + body + "</div>\n    \n    <!-- Optimization initialization script -->\n    <script>\n      // Initialize optimizations as soon as possible\n      (function() {\n        // Mark app as ready when DOM is loaded\n        document.addEventListener('DOMContentLoaded', function() {\n          document.body.classList.add('app-ready');\n        });\n        \n        // Service Worker update notification\n        window.addEventListener('sw-update-available', function(event) {\n          if (confirm('新しいバージョンが利用可能です。更新しますか？')) {\n            if (window.__swManager) {\n              window.__swManager.skipWaiting();\n            }\n          }\n        });\n        \n        // Performance monitoring for critical metrics\n        if ('performance' in window && 'getEntriesByType' in performance) {\n          window.addEventListener('load', function() {\n            // Record initial load time\n            const navigation = performance.getEntriesByType('navigation')[0];\n            if (navigation) {\n              console.log('Page load time:', navigation.loadEventEnd - navigation.loadEventStart, 'ms');\n            }\n          });\n        }\n      })();\n    </script>\n  </body>\n</html>\n",
		error: ({ status, message }) => "<!doctype html>\n<html lang=\"en\">\n\t<head>\n\t\t<meta charset=\"utf-8\" />\n\t\t<title>" + message + "</title>\n\n\t\t<style>\n\t\t\tbody {\n\t\t\t\t--bg: white;\n\t\t\t\t--fg: #222;\n\t\t\t\t--divider: #ccc;\n\t\t\t\tbackground: var(--bg);\n\t\t\t\tcolor: var(--fg);\n\t\t\t\tfont-family:\n\t\t\t\t\tsystem-ui,\n\t\t\t\t\t-apple-system,\n\t\t\t\t\tBlinkMacSystemFont,\n\t\t\t\t\t'Segoe UI',\n\t\t\t\t\tRoboto,\n\t\t\t\t\tOxygen,\n\t\t\t\t\tUbuntu,\n\t\t\t\t\tCantarell,\n\t\t\t\t\t'Open Sans',\n\t\t\t\t\t'Helvetica Neue',\n\t\t\t\t\tsans-serif;\n\t\t\t\tdisplay: flex;\n\t\t\t\talign-items: center;\n\t\t\t\tjustify-content: center;\n\t\t\t\theight: 100vh;\n\t\t\t\tmargin: 0;\n\t\t\t}\n\n\t\t\t.error {\n\t\t\t\tdisplay: flex;\n\t\t\t\talign-items: center;\n\t\t\t\tmax-width: 32rem;\n\t\t\t\tmargin: 0 1rem;\n\t\t\t}\n\n\t\t\t.status {\n\t\t\t\tfont-weight: 200;\n\t\t\t\tfont-size: 3rem;\n\t\t\t\tline-height: 1;\n\t\t\t\tposition: relative;\n\t\t\t\ttop: -0.05rem;\n\t\t\t}\n\n\t\t\t.message {\n\t\t\t\tborder-left: 1px solid var(--divider);\n\t\t\t\tpadding: 0 0 0 1rem;\n\t\t\t\tmargin: 0 0 0 1rem;\n\t\t\t\tmin-height: 2.5rem;\n\t\t\t\tdisplay: flex;\n\t\t\t\talign-items: center;\n\t\t\t}\n\n\t\t\t.message h1 {\n\t\t\t\tfont-weight: 400;\n\t\t\t\tfont-size: 1em;\n\t\t\t\tmargin: 0;\n\t\t\t}\n\n\t\t\t@media (prefers-color-scheme: dark) {\n\t\t\t\tbody {\n\t\t\t\t\t--bg: #222;\n\t\t\t\t\t--fg: #ddd;\n\t\t\t\t\t--divider: #666;\n\t\t\t\t}\n\t\t\t}\n\t\t</style>\n\t</head>\n\t<body>\n\t\t<div class=\"error\">\n\t\t\t<span class=\"status\">" + status + "</span>\n\t\t\t<div class=\"message\">\n\t\t\t\t<h1>" + message + "</h1>\n\t\t\t</div>\n\t\t</div>\n\t</body>\n</html>\n"
	},
	version_hash: "1yw3vvi"
};

export async function get_hooks() {
	let handle;
	let handleFetch;
	let handleError;
	let handleValidationError;
	let init;
	

	let reroute;
	let transport;
	

	return {
		handle,
		handleFetch,
		handleError,
		handleValidationError,
		init,
		reroute,
		transport
	};
}

export { set_assets, set_building, set_manifest, set_prerendering, set_private_env, set_public_env, set_read_implementation };
