import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/api': {
				target: 'http://host.docker.internal:3000',
				changeOrigin: true,
				configure: (proxy, options) => {
					proxy.on('error', (err, req, res) => {
						console.error('Proxy error:', err);
					});
					proxy.on('proxyReq', (proxyReq, req, res) => {
						console.log(`Proxying request from ${req.originalUrl} to ${options.target}${proxyReq.path}`);
					});
				}
			}
		}
	}
});
