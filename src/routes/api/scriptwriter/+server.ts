import { json } from '@sveltejs/kit';
import { exec } from 'child_process';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { promisify } from 'util';

const execPromise = promisify(exec);

// Get the directory paths
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const scriptsDir = path.resolve(__dirname, '../../../../static/scripts');
const baseOutputDir = path.resolve(scriptsDir, 'output');

export async function POST({ request }) {
  try {
    const { content, sessionId } = await request.json();
    
    // Validate content
    if (!content || content.trim() === '') {
      return json({ message: 'Content prompt cannot be empty' }, { status: 400 });
    }
    
    // Use sessionId to create a session-specific directory, fallback to 'default' if not provided
    const sessionFolder = sessionId || 'default';
    const tempOutputDir = path.resolve(baseOutputDir, sessionFolder);
    
    // Create the output directory if it doesn't exist
    if (!fs.existsSync(tempOutputDir)) {
      fs.mkdirSync(tempOutputDir, { recursive: true });
    }
    
    // Build the command to run the Go script
    const scriptPath = path.join(scriptsDir, 'ScriptWriter.go');
    const command = `cd "${scriptsDir}" && go run ScriptWriter.go -prompt "${content.replace(/"/g, '\\"')}" -output "${tempOutputDir}"`;
    
    console.log(`Executing command: ${command}`);
    
    // Execute the command
    const { stdout, stderr } = await execPromise(command);
    
    // Handle any errors from the script
    if (stderr && !stderr.includes('successfully generated')) {
      console.error('Error from ScriptWriter.go:', stderr);
      return json({ message: 'Script generation failed', details: stderr }, { status: 500 });
    }
    
    console.log('ScriptWriter output:', stdout);
    
    // Return success with paths to generated scripts with the session-specific directory
    return json({
      message: 'Scripts generated successfully',
      sessionId: sessionFolder,
      scriptPaths: {
        script1: path.join(tempOutputDir, 'script1.txt'),
        script2: path.join(tempOutputDir, 'script2.txt')
      }
    });
    
  } catch (error) {
    console.error('Error in scriptwriter API:', error);
    return json({ 
      message: 'Failed to generate script', 
      details: error instanceof Error ? error.message : 'Unknown error' 
    }, { status: 500 });
  }
}