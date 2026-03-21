import http from 'node:http';
import https from 'node:https';

import next from 'next';

const port = Number.parseInt(process.env.PORT ?? '3001', 10);
const hostname = process.env.HOSTNAME || '127.0.0.1';
const backendBaseUrl = new URL(
  (process.env.NEXT_DEV_BACKEND_URL || 'http://127.0.0.1:3000').replace(/\/+$/, ''),
);

const app = next({
  dev: true,
  hostname,
  port,
});

let handle;
let handleUpgrade;

function isApiRequest(url = '') {
  return url.startsWith('/api/');
}

function createProxyRequestOptions(req, targetUrl) {
  const headers = {
    ...req.headers,
    host: targetUrl.host,
  };

  if (headers.origin) {
    headers.origin = targetUrl.origin;
  }

  return {
    protocol: targetUrl.protocol,
    hostname: targetUrl.hostname,
    port: targetUrl.port || (targetUrl.protocol === 'https:' ? 443 : 80),
    method: req.method,
    path: `${targetUrl.pathname}${targetUrl.search}`,
    headers,
  };
}

function writeProxyError(res, error) {
  res.statusCode = 502;
  res.setHeader('Content-Type', 'application/json; charset=utf-8');
  res.end(
    JSON.stringify({
      success: false,
      message: `开发代理连接后端失败: ${error.message}`,
    }),
  );
}

function proxyHttpRequest(req, res) {
  const targetUrl = new URL(req.url || '/', backendBaseUrl);
  const requestImpl = targetUrl.protocol === 'https:' ? https.request : http.request;
  const proxyReq = requestImpl(createProxyRequestOptions(req, targetUrl), (proxyRes) => {
    res.writeHead(proxyRes.statusCode || 502, proxyRes.statusMessage, proxyRes.headers);
    proxyRes.pipe(res);
  });

  proxyReq.on('error', (error) => {
    if (!res.headersSent) {
      writeProxyError(res, error);
      return;
    }
    res.destroy(error);
  });

  req.pipe(proxyReq);
}

function formatUpgradeResponseHeaders(statusCode, statusMessage, headers) {
  const lines = [`HTTP/1.1 ${statusCode} ${statusMessage}`];
  for (const [key, value] of Object.entries(headers)) {
    if (value === undefined) {
      continue;
    }
    if (Array.isArray(value)) {
      for (const item of value) {
        lines.push(`${key}: ${item}`);
      }
      continue;
    }
    lines.push(`${key}: ${value}`);
  }
  lines.push('', '');
  return lines.join('\r\n');
}

function proxyWebSocketUpgrade(req, socket, head) {
  const targetUrl = new URL(req.url || '/', backendBaseUrl);
  const requestImpl = targetUrl.protocol === 'https:' ? https.request : http.request;
  const proxyReq = requestImpl(createProxyRequestOptions(req, targetUrl));

  proxyReq.on('upgrade', (proxyRes, proxySocket, proxyHead) => {
    socket.write(
      formatUpgradeResponseHeaders(
        proxyRes.statusCode || 101,
        proxyRes.statusMessage || 'Switching Protocols',
        proxyRes.headers,
      ),
    );

    if (head.length > 0) {
      proxySocket.write(head);
    }
    if (proxyHead.length > 0) {
      socket.write(proxyHead);
    }

    proxySocket.pipe(socket).pipe(proxySocket);

    proxySocket.on('error', () => socket.destroy());
    socket.on('error', () => proxySocket.destroy());
  });

  proxyReq.on('response', (proxyRes) => {
    socket.write(
      formatUpgradeResponseHeaders(
        proxyRes.statusCode || 502,
        proxyRes.statusMessage || 'Bad Gateway',
        proxyRes.headers,
      ),
    );
    proxyRes.pipe(socket);
  });

  proxyReq.on('error', () => {
    socket.destroy();
  });

  proxyReq.end();
}

await app.prepare();
handle = app.getRequestHandler();
handleUpgrade = app.getUpgradeHandler();

const server = http.createServer((req, res) => {
  if (isApiRequest(req.url)) {
    proxyHttpRequest(req, res);
    return;
  }
  handle(req, res);
});

server.on('upgrade', (req, socket, head) => {
  if (isApiRequest(req.url)) {
    proxyWebSocketUpgrade(req, socket, head);
    return;
  }
  handleUpgrade(req, socket, head);
});

server.listen(port, hostname, () => {
  console.log(`GinNextTemplate web dev server listening on http://${hostname}:${port}`);
  console.log(`Proxying /api/* and websocket upgrades to ${backendBaseUrl.href}`);
});
