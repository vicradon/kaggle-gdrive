const { OAuth2Client } = require('google-auth-library');
const http = require('http');
const url = require('url');
const fs = require('fs');

const { client_id, client_secret } = JSON.parse(fs.readFileSync('./client_secret.json', 'utf8')).installed;
const REDIRECT_URI = 'http://localhost:3000/callback';

const client = new OAuth2Client(client_id, client_secret, REDIRECT_URI);

const authUrl = client.generateAuthUrl({
  access_type: 'offline',
  scope: ['https://www.googleapis.com/auth/drive'],
  prompt: 'consent',
});

console.log('Visit this URL:\n', authUrl);

const server = http.createServer(async (req, res) => {
  const code = new url.URL(req.url, 'http://localhost:3000').searchParams.get('code');
  if (!code) return;
  const { tokens } = await client.getToken(code);
  console.log('\nClient ID:      ', client_id);
  console.log('Client Secret:  ', client_secret);
  console.log('Refresh Token:  ', tokens.refresh_token);
  res.end('Done! You can close this tab.');
  server.close();
}).listen(3000);
