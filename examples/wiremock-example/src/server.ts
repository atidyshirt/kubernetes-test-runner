import http from 'http';

const WIREMOCK_HOST = process.env.WIREMOCK_HOST || 'wiremock';
const WIREMOCK_PORT = parseInt(process.env.WIREMOCK_PORT || '8080');

const server = http.createServer((req, res) => {
  if (req.url === '/api/data' && req.method === 'GET') {
    const options = {
      hostname: WIREMOCK_HOST,
      port: WIREMOCK_PORT,
      path: '/api/data',
      method: 'GET',
    };

    const proxyReq = http.request(options, (proxyRes) => {
      res.writeHead(proxyRes.statusCode || 200, { 'Content-Type': 'application/json' });
      proxyRes.pipe(res);
    });

    proxyReq.on('error', (error) => {
      console.error('Error calling wiremock:', error);
      res.writeHead(500, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Failed to call wiremock' }));
    });

    req.pipe(proxyReq);
  } else {
    res.writeHead(404, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ error: 'Not found' }));
  }
});

const PORT = parseInt(process.env.PORT || '3000');
server.listen(PORT, () => {
  console.log(`Example service listening on port ${PORT}`);
});
