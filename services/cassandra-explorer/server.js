const express = require('express');
const cors = require('cors');
const cassandra = require('cassandra-driver');

const app = express();
app.use(cors());
app.use(express.json());

let client;
let connected = false;

async function initCassandra() {
  try {
    client = new cassandra.Client({
      contactPoints: ['cassandra'],
      localDataCenter: 'datacenter1',
      keyspace: 'message_service',
      protocolVersion: 3,
      ioOptions: {
        coreConnectionsPerHost: { '0': 1, '1': 1, '2': 1 },
        maxConnectionsPerHost: { '0': 4, '1': 4, '2': 4 }
      }
    });
    
    await client.connect();
    connected = true;
    console.log('Cassandra connected');
    return true;
  } catch (err) {
    console.error('Cassandra connection failed:', err.message);
    return false;
  }
}

// Health check endpoint  
app.get('/health', (req, res) => {
  if (connected) {
    res.status(200).json({ status: 'healthy', cassandra: 'connected' });
  } else {
    res.status(503).json({ status: 'unhealthy', cassandra: 'disconnected' });
  }
});

// API Routes
app.get('/api/tables', async (req, res) => {
  if (!connected) {
    return res.status(503).json({ error: 'Cassandra not connected' });
  }
  
  try {
    const result = await client.execute(`
      SELECT table_name FROM system_schema.tables 
      WHERE keyspace_name = 'message_service'
    `);
    res.json(result.rows.map(r => r.table_name));
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/api/data/:table', async (req, res) => {
  if (!connected) {
    return res.status(503).json({ error: 'Cassandra not connected' });
  }
  
  try {
    const limit = req.query.limit || 100;
    const result = await client.execute(`SELECT * FROM ${req.params.table} LIMIT ?`, [limit], { prepare: true });
    res.json(result.rows);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/api/schema/:table', async (req, res) => {
  if (!connected) {
    return res.status(503).json({ error: 'Cassandra not connected' });
  }
  
  try {
    const result = await client.execute(`
      SELECT column_name, type FROM system_schema.columns 
      WHERE keyspace_name = 'message_service' AND table_name = ?
    `, [req.params.table], { prepare: true });
    res.json(result.rows);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.post('/api/query', async (req, res) => {
  if (!connected) {
    return res.status(503).json({ error: 'Cassandra not connected' });
  }
  
  try {
    const { query } = req.body;
    if (!query) return res.status(400).json({ error: 'Query required' });
    
    const result = await client.execute(query);
    res.json(result.rows);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// Dashboard HTML
const html = `
<!DOCTYPE html>
<html>
<head>
  <title>Cassandra Explorer</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: 'Segoe UI', sans-serif; background: #f5f5f5; }
    .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; }
    .header h1 { font-size: 28px; margin-bottom: 5px; }
    .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
    .tabs { display: flex; gap: 10px; margin-bottom: 20px; border-bottom: 2px solid #ddd; }
    .tab-btn { background: white; border: none; padding: 12px 20px; cursor: pointer; border-bottom: 3px solid transparent; }
    .tab-btn.active { border-bottom-color: #667eea; color: #667eea; }
    .tab-content { display: none; }
    .tab-content.active { display: block; }
    table { width: 100%; border-collapse: collapse; background: white; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    th { background: #667eea; color: white; padding: 12px; text-align: left; }
    td { padding: 10px 12px; border-bottom: 1px solid #eee; }
    tr:hover { background: #f9f9f9; }
    .select { padding: 10px; margin-bottom: 20px; }
    select { padding: 8px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
    textarea { padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
    button { margin-top: 10px; padding: 10px 20px; background: #667eea; color: white; border: none; border-radius: 4px; cursor: pointer; }
    button:hover { background: #764ba2; }
    .loading { text-align: center; padding: 40px; color: #667eea; }
  </style>
</head>
<body>
<div class="header">
  <h1>Cassandra Explorer</h1>
  <p>Interface Web pour explorer votre base de données Cassandra</p>
</div>

<div class="container">
  <div class="tabs">
    <button class="tab-btn active" onclick="switchTab('tables')">Tables</button>
    <button class="tab-btn" onclick="switchTab('query')">Requête</button>
  </div>

  <div id="tables" class="tab-content active">
    <div class="select">
      <select id="tableSelect" onchange="loadTableData()">
        <option value="">Choisir une table...</option>
      </select>
    </div>
    <div id="tableContent"></div>
  </div>

  <div id="query" class="tab-content">
    <div class="select">
      <textarea id="queryInput" placeholder="SELECT * FROM message_service.conversations LIMIT 10;"></textarea>
      <button onclick="executeQuery()">Exécuter</button>
    </div>
    <div id="queryContent"></div>
  </div>
</div>

<script>
  function switchTab(tab) {
    document.querySelectorAll('.tab-content').forEach(el => el.classList.remove('active'));
    document.querySelectorAll('.tab-btn').forEach(el => el.classList.remove('active'));
    document.getElementById(tab).classList.add('active');
    event.target.classList.add('active');
    
    if (tab === 'tables') loadTables();
  }

  async function loadTables() {
    try {
      const res = await fetch('/api/tables');
      const tables = await res.json();
      const select = document.getElementById('tableSelect');
      select.innerHTML = '<option value="">Choisir une table...</option>';
      tables.forEach(t => {
        const opt = document.createElement('option');
        opt.value = t;
        opt.textContent = t;
        select.appendChild(opt);
      });
    } catch (e) {
      alert('Erreur: ' + e.message);
    }
  }

  async function loadTableData() {
    const table = document.getElementById('tableSelect').value;
    if (!table) return;
    
    document.getElementById('tableContent').innerHTML = '<div class="loading">Chargement...</div>';
    
    try {
      const [schemaRes, dataRes] = await Promise.all([
        fetch('/api/schema/' + table),
        fetch('/api/data/' + table)
      ]);
      
      const schema = await schemaRes.json();
      const data = await dataRes.json();
      
      let html = '<table><tr>';
      schema.forEach(col => {
        html += '<th>' + col.column_name + '<br><small>' + col.type + '</small></th>';
      });
      html += '</tr>';
      
      data.forEach(row => {
        html += '<tr>';
        Object.values(row).forEach(val => {
          html += '<td>' + (val === null ? 'NULL' : String(val).substring(0, 100)) + '</td>';
        });
        html += '</tr>';
      });
      html += '</table>';
      
      document.getElementById('tableContent').innerHTML = html;
    } catch (e) {
      document.getElementById('tableContent').innerHTML = '<div style="color: red;">Erreur: ' + e.message + '</div>';
    }
  }

  async function executeQuery() {
    const query = document.getElementById('queryInput').value;
    if (!query) return alert('Entrez une requête');
    
    document.getElementById('queryContent').innerHTML = '<div class="loading">Exécution...</div>';
    
    try {
      const res = await fetch('/api/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query })
      });
      const data = await res.json();
      
      if (!Array.isArray(data) || data.length === 0) {
        document.getElementById('queryContent').innerHTML = '<p>Zéro résultat</p>';
        return;
      }
      
      const keys = Object.keys(data[0]);
      let html = '<table><tr>' + keys.map(k => '<th>' + k + '</th>').join('') + '</tr>';
      data.forEach(row => {
        html += '<tr>' + keys.map(k => '<td>' + (row[k] === null ? 'NULL' : String(row[k]).substring(0, 100)) + '</td>').join('') + '</tr>';
      });
      html += '</table>';
      
      document.getElementById('queryContent').innerHTML = html;
    } catch (e) {
      document.getElementById('queryContent').innerHTML = '<div style="color: red;">Erreur: ' + e.message + '</div>';
    }
  }

  loadTables();
</script>
</body>
</html>
`;

app.get('/', (req, res) => {
  res.setHeader('Content-Type', 'text/html');
  res.send(html);
});

const PORT = process.env.PORT || 3000;

// Start server and try to connect to Cassandra
async function startServer() {
  app.listen(PORT, () => {
    console.log('Cassandra Explorer running on port ' + PORT);
  });
  
  // Try to connect to Cassandra every 2 seconds
  setInterval(async () => {
    if (!connected) {
      await initCassandra();
    }
  }, 2000);
}

startServer();

