import { error } from '@sveltejs/kit';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

export async function GET({ params }) {
  const { sessionId, filename } = params;
  
  // Validate filename to prevent directory traversal attacks
  if (!/^script[12]\.txt$/.test(filename)) {
    throw error(400, 'Invalid filename');
  }
  
  // Set base directory path for content files
  const __dirname = path.dirname(fileURLToPath(import.meta.url));
  const contentDir = path.resolve(__dirname, '../../../../../out', sessionId);
  const filePath = path.join(contentDir, filename);
  
  try {
    // Check if file exists
    if (!fs.existsSync(filePath)) {
      throw error(404, 'File not found');
    }
    
    // Read file content
    const content = fs.readFileSync(filePath, 'utf-8');
    
    // Return content with appropriate headers
    return new Response(content, {
      headers: {
        'Content-Type': 'text/plain'
      }
    });
  } catch (err) {
    if (err instanceof Error && 'status' in err && err.status === 404) {
      throw err;
    }
    console.error(`Error serving file ${filePath}:`, err);
    throw error(500, 'Internal server error');
  }
}