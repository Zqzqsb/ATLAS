/**
 * ATLAS Frontend SSR Server
 * 
 * This server:
 * 1. Pre-fetches initial data from backend on page load (SSR)
 * 2. Proxies API requests to backend container
 * 3. Serves static assets
 */

const express = require('express');
const axios = require('axios');
const path = require('path');
const fs = require('fs');

const app = express();
const PORT = process.env.PORT || 80;
const BACKEND_URL = process.env.BACKEND_URL || 'http://backend:8080';

// Parse JSON body - MUST be before API routes
app.use(express.json());

// Read the index.html template
const distPath = path.join(__dirname, 'dist');
const indexHtmlPath = path.join(distPath, 'index.html');

// Function to read index.html - reads fresh copy each time to avoid stale cache
function getIndexHtml() {
  return fs.readFileSync(indexHtmlPath, 'utf-8');
}

// Pre-fetch initial data from backend
async function fetchInitialData() {
  const data = {
    connections: null,
    databases: null,
    spiderDatabases: null,
    systemInfo: null,
    error: null
  };

  try {
    // Fetch all initial data in parallel
    const [connectionsRes, databasesRes, spiderDbRes, systemInfoRes] = await Promise.allSettled([
      axios.get(`${BACKEND_URL}/api/v1/connections`, { timeout: 5000 }),
      axios.get(`${BACKEND_URL}/api/v1/databases`, { timeout: 5000 }),
      axios.get(`${BACKEND_URL}/api/v1/spider/databases?source=spider_sqlite`, { timeout: 5000 }),
      axios.get(`${BACKEND_URL}/api/v1/system/info`, { timeout: 5000 })
    ]);

    if (connectionsRes.status === 'fulfilled') {
      data.connections = connectionsRes.value.data;
    }
    if (databasesRes.status === 'fulfilled') {
      data.databases = databasesRes.value.data;
    }
    if (spiderDbRes.status === 'fulfilled') {
      data.spiderDatabases = spiderDbRes.value.data;
    }
    if (systemInfoRes.status === 'fulfilled') {
      data.systemInfo = systemInfoRes.value.data;
    }
  } catch (err) {
    console.error('Error fetching initial data:', err.message);
    data.error = err.message;
  }

  return data;
}

// Inject SSR data into HTML
function injectSSRData(html, data) {
  const script = `<script>window.__SSR_DATA__ = ${JSON.stringify(data)};</script>`;
  // Insert before closing </head> tag
  return html.replace('</head>', `${script}</head>`);
}

// Manual API proxy - forward all /api requests to backend
app.use('/api', async (req, res) => {
  const targetUrl = `${BACKEND_URL}${req.originalUrl}`;

  try {
    // Check if this is a streaming request
    const isStreaming = req.originalUrl.includes('/stream');

    const axiosConfig = {
      method: req.method,
      url: targetUrl,
      headers: {
        ...req.headers,
        host: new URL(BACKEND_URL).host
      },
      data: req.method !== 'GET' ? req.body : undefined,
      timeout: 120000, // 2 minutes for LLM calls
      responseType: isStreaming ? 'stream' : 'json',
      validateStatus: () => true // Don't throw on any status
    };

    // Remove headers that shouldn't be proxied
    delete axiosConfig.headers['content-length'];
    delete axiosConfig.headers['connection'];

    const response = await axios(axiosConfig);

    // Copy response headers
    Object.entries(response.headers).forEach(([key, value]) => {
      if (key !== 'transfer-encoding') {
        res.setHeader(key, value);
      }
    });

    res.status(response.status);

    if (isStreaming && response.data.pipe) {
      // Stream response
      response.data.pipe(res);
    } else {
      res.json(response.data);
    }
  } catch (err) {
    console.error(`API proxy error for ${targetUrl}:`, err.message);
    res.status(502).json({ error: 'Backend unavailable', message: err.message });
  }
});

// Health check endpoint
app.get('/health', async (req, res) => {
  try {
    const response = await axios.get(`${BACKEND_URL}/health`, { timeout: 3000 });
    res.json({ status: 'ok', backend: response.data });
  } catch (err) {
    res.status(503).json({ status: 'error', message: err.message });
  }
});

// Serve static assets - disable cache to ensure fresh assets after rebuild
app.use('/assets', express.static(path.join(distPath, 'assets'), {
  maxAge: 0,
  etag: false,
  lastModified: false,
  setHeaders: (res) => {
    res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
    res.setHeader('Pragma', 'no-cache');
    res.setHeader('Expires', '0');
  }
}));

// Serve other static files (favicon, etc.)
app.use(express.static(distPath, {
  index: false,
  maxAge: 0,
  etag: false,
  setHeaders: (res) => {
    res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
  }
}));

// SSR route - all other routes get the index.html with pre-fetched data
app.use(async (req, res, next) => {
  // Only handle GET requests for HTML pages
  if (req.method !== 'GET') {
    return next();
  }
  try {
    // Read fresh index.html for each request (avoids SSR cache issues)
    const indexHtml = getIndexHtml();

    // Fetch initial data from backend
    const initialData = await fetchInitialData();

    // Inject data into HTML
    const html = injectSSRData(indexHtml, initialData);

    res.setHeader('Content-Type', 'text/html');
    res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
    res.send(html);
  } catch (err) {
    console.error('SSR error:', err);
    // Fallback to original HTML without SSR data
    const indexHtml = getIndexHtml();
    res.setHeader('Content-Type', 'text/html');
    res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
    res.send(indexHtml);
  }
});

// Start server
app.listen(PORT, '0.0.0.0', () => {
  console.log(`🚀 ATLAS Frontend Server running on port ${PORT}`);
  console.log(`   Backend URL: ${BACKEND_URL}`);
  console.log(`   Static files: ${distPath}`);
});
