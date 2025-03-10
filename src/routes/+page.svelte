<script lang="ts"> 
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  
  let deepseekApiKey: string = '';
  let plexelsApiKeys: string[] = [];
  let currentPlexelsApiKey: string = '';
  let generateContent: string = '';
  let saveMessage: string = '';
  let errorMessage: string = '';
  let sessionId: string = ''; // Add this to store the random folder name
  let isGenerating: boolean = false;

  onMount(async () => {
    // Load saved API keys if they exist
    try {
      const response = await fetch('/api/config');
      if (response.ok) {
        const config = await response.json();
        deepseekApiKey = config.deepseekApiKey || '';
        plexelsApiKeys = config.plexelsApiKeys || [];
      }
    } catch (error) {
      console.error('Error loading config:', error);
    }
  });

  async function saveApiKeys() {
    try {
      const response = await fetch('/api/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ deepseekApiKey, plexelsApiKeys })
      });
      
      if (response.ok) {
        saveMessage = 'API keys saved successfully!';
        setTimeout(() => saveMessage = '', 3000);
      } else {
        saveMessage = 'Failed to save API keys.';
      }
    } catch (error) {
      console.error('Error saving API keys:', error);
      saveMessage = 'Error saving API keys.';
    }
  }

  function handlePlexelsKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && currentPlexelsApiKey.trim()) {
      event.preventDefault();
      addPlexelsApiKey();
    }
  }

  function addPlexelsApiKey() {
    const key = currentPlexelsApiKey.trim();
    if (key && !plexelsApiKeys.includes(key)) {
      plexelsApiKeys = [...plexelsApiKeys, key];
      currentPlexelsApiKey = '';
    }
  }

  function removePlexelsApiKey(key: String) {
    plexelsApiKeys = plexelsApiKeys.filter(k => k !== key);
  }

  function generateRandomId() {
    // Create a random ID with timestamp prefix for uniqueness
    const timestamp = Date.now().toString(36);
    const randomChars = Math.random().toString(36).substring(2, 8);
    return `${timestamp}-${randomChars}`;
  }

  async function handleGenerate() {
    // Validate input
    if (!generateContent.trim()) {
      errorMessage = 'Please enter content to generate';
      return;
    }
    
    try {
      isGenerating = true; // Set loading state
      // Generate a random folder name to use as session ID
      sessionId = generateRandomId();
      
      // Call the API endpoint that runs the scriptwriter.go script
      const response = await fetch('/api/scriptwriter', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          content: generateContent,
          sessionId: sessionId  // Pass the session ID to the API
        })
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || 'Failed to generate script');
      }
      
      // Navigate to the generator page with the session ID
      goto(`/generator?session=${sessionId}`);
    } catch (error) {
      console.error('Error during script generation:', error);
      errorMessage = error instanceof Error ? error.message : 'Error generating content. Please try again.';
    } finally {
      isGenerating = false; // Reset loading state
    }
  }

  function closeErrorPopup() {
    errorMessage = '';
  }
</script>

<main>
  <h1>ShortGen</h1>
  
  <section class="api-keys-section">
    <h2>API Settings</h2>
    <div class="form-group">
      <label for="deepseek">DeepSeek API Key</label>
      <input 
        id="deepseek" 
        type="password" 
        bind:value={deepseekApiKey} 
        placeholder="Enter your DeepSeek API key"
      />
    </div>

    <div class="form-group">
      <label for="plexels">Plexels API Keys</label>
      <input 
        id="plexels" 
        type="password" 
        bind:value={currentPlexelsApiKey} 
        on:keydown={handlePlexelsKeydown}
        placeholder="Enter your Plexels API key and press Enter"
      />
      
      {#if plexelsApiKeys.length > 0}
        <div class="api-keys-list">
          {#each plexelsApiKeys as key}
            <div class="api-key-item">
              <span class="api-key-text">{key.slice(0, 4)}•••</span>
              <button type="button" class="remove-key" on:click={() => removePlexelsApiKey(key)}>×</button>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <button type="button" class="save-button" on:click={saveApiKeys}>Save API Keys</button>
    {#if saveMessage}
      <p class="save-message">{saveMessage}</p>
    {/if}
  </section>
  
  <section class="content-section">
    <h2>Generate Content</h2>
    <form on:submit|preventDefault={handleGenerate}>
      <div class="form-group">
        <label for="content">What to Generate</label>
        <textarea 
          id="content" 
          bind:value={generateContent} 
          placeholder="Describe what you want to generate..."
          rows="5"
        ></textarea>
      </div>

      <button type="submit" disabled={isGenerating}>
        {#if isGenerating}
          Generating...
        {:else}
          Generate
        {/if}
      </button>
    </form>
  </section>

  {#if errorMessage}
    <div class="error-popup">
      <div class="error-content">
        <button type="button" class="close-error" on:click={closeErrorPopup}>×</button>
        <h3>Error</h3>
        <p>{errorMessage}</p>
        <button type="button" class="error-ok-button" on:click={closeErrorPopup}>OK</button>
      </div>
    </div>
  {/if}
</main>

<style>
  main {
    max-width: 600px;
    margin: 0 auto;
    padding: 2rem;
  }

  h1 {
    margin-bottom: 2rem;
    text-align: center;
  }

  h2 {
    margin-bottom: 1rem;
    font-size: 1.5rem;
  }

  section {
    margin-bottom: 2rem;
    padding: 1.5rem;
    border: 1px solid #eee;
    border-radius: 8px;
    background-color: #f9f9f9;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: bold;
  }

  input, textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 1rem;
  }

  button {
    display: block;
    width: 100%;
    padding: 0.75rem;
    background-color: #3b82f6;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  button:hover {
    background-color: #2563eb;
  }

  .save-button {
    background-color: #10b981;
  }

  .save-button:hover {
    background-color: #059669;
  }

  .save-message {
    margin-top: 0.75rem;
    text-align: center;
    font-weight: bold;
    color: #10b981;
  }

  .api-keys-list {
    margin-top: 1rem;
  }

  .api-key-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    margin-bottom: 0.5rem;
    background-color: #fff;
  }

  .api-key-text {
    font-family: monospace;
  }

  .remove-key {
    background: none;
    border: none;
    color: #f87171;
    font-size: 1.25rem;
    cursor: pointer;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    min-width: 24px;
  }

  .remove-key:hover {
    color: #ef4444;
  }

  .error-popup {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .error-content {
    background-color: white;
    padding: 2rem;
    border-radius: 8px;
    max-width: 500px;
    width: 90%;
    position: relative;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .close-error {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #666;
  }

  .error-content h3 {
    margin-top: 0;
    color: #e53e3e;
  }

  .error-ok-button {
    margin-top: 1rem;
    width: 100%;
    background-color: #e53e3e;
  }

  .error-ok-button:hover {
    background-color: #c53030;
  }
</style>
