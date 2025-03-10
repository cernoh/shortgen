import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import toml from '@iarna/toml';
import { json } from '@sveltejs/kit';

// Get the directory path for the config file
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const configDir = path.resolve(__dirname, '../../../../static/scripts');
const configPath = path.join(configDir, 'config.toml');

export async function GET() {
  try {
    // Check if config file exists
    if (fs.existsSync(configPath)) {
      const configFile = fs.readFileSync(configPath, 'utf8');
      const config = toml.parse(configFile);
      
      // Ensure plexelsApiKeys is always an array
      if (config.plexelsApiKeys && !Array.isArray(config.plexelsApiKeys)) {
        config.plexelsApiKeys = [String(config.plexelsApiKeys)];
      } else if (!config.plexelsApiKeys) {
        config.plexelsApiKeys = [];
      }
      
      return json(config);
    } else {
      return json({ deepseekApiKey: '', plexelsApiKeys: [] });
    }
  } catch (error) {
    console.error('Error reading config file:', error);
    return json({ error: 'Failed to read config' }, { status: 500 });
  }
}

export async function POST({ request }) {
  try {
    const { deepseekApiKey, plexelsApiKeys } = await request.json();
    
    // Create config object
    const config = {
      deepseekApiKey,
      plexelsApiKeys: Array.isArray(plexelsApiKeys) ? plexelsApiKeys : []
    };
    
    // Ensure the directory exists
    if (!fs.existsSync(configDir)) {
      fs.mkdirSync(configDir, { recursive: true });
    }
    
    // Convert to TOML and write to file
    const tomlContent = toml.stringify(config);
    fs.writeFileSync(configPath, tomlContent, 'utf8');
    
    return json({ success: true });
  } catch (error) {
    console.error('Error writing config file:', error);
    return json({ error: 'Failed to write config' }, { status: 500 });
  }
}