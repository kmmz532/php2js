// Package worker generates Cloudflare Workers entry point and config.
package worker

import (
	"os"
	"path/filepath"
	"strings"
)

// Generate creates the Workers entry point and wrangler config.
func Generate(outputDir string, projectName string) error {
	srcDir := filepath.Join(outputDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	// Write worker entry point
	if err := os.WriteFile(filepath.Join(srcDir, "index.js"), []byte(workerEntry), 0644); err != nil {
		return err
	}

	// Write wrangler.toml
	toml := strings.ReplaceAll(wranglerTemplate, "{{PROJECT_NAME}}", projectName)
	if err := os.WriteFile(filepath.Join(outputDir, "wrangler.toml"), []byte(toml), 0644); err != nil {
		return err
	}

	// Write package.json
	pkg := strings.ReplaceAll(packageJSON, "{{PROJECT_NAME}}", projectName)
	if err := os.WriteFile(filepath.Join(outputDir, "package.json"), []byte(pkg), 0644); err != nil {
		return err
	}

	return nil
}

const workerEntry = `// Auto-generated Cloudflare Workers entry point
import * as __runtime from './runtime/index.js';

export default {
  async fetch(request, env, ctx) {
    // Initialize the PHP runtime with CF bindings
    __runtime.init({
      env,
      request,
      ctx,
      r2: env.R2_BUCKET,
      d1: env.DB,
    });

    try {
      // Parse the request
      const url = new URL(request.url);
      const method = request.method;

      // Set up PHP superglobals
      __runtime.SERVER['REQUEST_METHOD'] = method;
      __runtime.SERVER['REQUEST_URI'] = url.pathname + url.search;
      __runtime.SERVER['QUERY_STRING'] = url.search.slice(1);
      __runtime.SERVER['HTTP_HOST'] = url.hostname;
      __runtime.SERVER['SCRIPT_NAME'] = '/index.php';
      __runtime.SERVER['SERVER_NAME'] = url.hostname;
      __runtime.SERVER['SERVER_PORT'] = url.port || (url.protocol === 'https:' ? '443' : '80');
      __runtime.SERVER['HTTPS'] = url.protocol === 'https:' ? 'on' : 'off';
      __runtime.SERVER['REMOTE_ADDR'] = request.headers.get('cf-connecting-ip') || '127.0.0.1';
      __runtime.SERVER['HTTP_USER_AGENT'] = request.headers.get('user-agent') || '';
      __runtime.SERVER['HTTP_REFERER'] = request.headers.get('referer') || '';
      __runtime.SERVER['CONTENT_TYPE'] = request.headers.get('content-type') || '';

      // Parse GET parameters
      for (const [key, value] of url.searchParams.entries()) {
        __runtime.GET[key] = value;
      }

      // Parse POST body
      if (method === 'POST') {
        const contentType = request.headers.get('content-type') || '';
        if (contentType.includes('application/x-www-form-urlencoded')) {
          const text = await request.text();
          const params = new URLSearchParams(text);
          for (const [key, value] of params.entries()) {
            __runtime.POST[key] = value;
          }
        } else if (contentType.includes('multipart/form-data')) {
          try {
            const formData = await request.formData();
            for (const [key, value] of formData.entries()) {
              if (value instanceof File) {
                __runtime.FILES[key] = {
                  name: value.name,
                  type: value.type,
                  size: value.size,
                  tmp_name: key,
                  error: 0,
                  _file: value,
                };
              } else {
                __runtime.POST[key] = value;
              }
            }
          } catch (e) {
            // Ignore multipart parse errors
          }
        } else if (contentType.includes('application/json')) {
          try {
            __runtime.POST = await request.json();
          } catch (e) {
            // Ignore JSON parse errors
          }
        }
      }

      // Parse cookies
      const cookieHeader = request.headers.get('cookie') || '';
      for (const cookie of cookieHeader.split(';')) {
        const [key, ...rest] = cookie.trim().split('=');
        if (key) {
          __runtime.COOKIE[key] = decodeURIComponent(rest.join('='));
        }
      }

      // Merge GET and POST into REQUEST
      Object.assign(__runtime.REQUEST, __runtime.GET, __runtime.POST);

      // Set request headers into SERVER
      for (const [key, value] of request.headers.entries()) {
        const serverKey = 'HTTP_' + key.toUpperCase().replace(/-/g, '_');
        __runtime.SERVER[serverKey] = value;
      }

      // Import and execute the transpiled index
      const app = await import('./transpiled/index.js');
      if (typeof app.default === 'function') {
        await app.default(__runtime);
      } else if (typeof app.main === 'function') {
        await app.main(__runtime);
      }

      // Build response
      const body = __runtime.getOutput();
      const status = __runtime.getStatusCode();
      const headers = __runtime.getHeaders();

      return new Response(body, { status, headers });
    } catch (error) {
      console.error('Worker error:', error);
      return new Response(
        ` + "`" + `<!DOCTYPE html><html><body><h1>500 Internal Server Error</h1><pre>${error.stack || error.message}</pre></body></html>` + "`" + `,
        { status: 500, headers: { 'Content-Type': 'text/html; charset=utf-8' } }
      );
    } finally {
      __runtime.reset();
    }
  },
};
`

const wranglerTemplate = `name = "{{PROJECT_NAME}}"
main = "src/index.js"
compatibility_date = "2024-12-01"

[[r2_buckets]]
binding = "R2_BUCKET"
bucket_name = "{{PROJECT_NAME}}-storage"

[[d1_databases]]
binding = "DB"
database_name = "{{PROJECT_NAME}}-db"
database_id = "REPLACE_WITH_YOUR_DATABASE_ID"
`

const packageJSON = `{
  "name": "{{PROJECT_NAME}}",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "dev": "wrangler dev",
    "deploy": "wrangler deploy"
  },
  "devDependencies": {
    "wrangler": "^3.0.0"
  }
}
`
