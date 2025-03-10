<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  
  // Get session ID from URL query parameters
  let sessionId: string | null = null;
  
  // Content for option boxes
  let option1Content = "Loading...";
  let option2Content = "Loading...";
  let isLoading = true;
  let loadError = false;
  
  async function loadContentFiles() {
    if (!sessionId) return;
    
    try {
      // Fetch content for both script files
      const [script1Response, script2Response] = await Promise.all([
        fetch(`/api/content/${sessionId}/script1.txt`),
        fetch(`/api/content/${sessionId}/script2.txt`)
      ]);
      
      if (!script1Response.ok || !script2Response.ok) {
        throw new Error('Failed to load one or more content files');
      }
      
      option1Content = await script1Response.text();
      option2Content = await script2Response.text();
      isLoading = false;
    } catch (error) {
      console.error('Error loading content files:', error);
      loadError = true;
      isLoading = false;
    }
  }
  
  onMount(() => {
    // Extract sessionId from URL query parameters
    sessionId = $page.url.searchParams.get('sessionId');
    console.log('Session ID:', sessionId);
    
    if (sessionId) {
      loadContentFiles();
    }
  });
  
  // State to track which box is selected
  let selectedBox: 'left' | 'right' | null = null;
  
  function selectBox(box: 'left' | 'right') {
    selectedBox = box;
  }
  
  function goBackToHome() {
    window.history.back();
  }
</script>

<main>
  <h1>Generated Content</h1>
  
  <div class="content-container">
    <div 
      class="content-box {selectedBox === 'left' ? 'selected' : ''}"
      role="button"
      tabindex="0"
      on:click={() => selectBox('left')}
      on:keydown={(e) => e.key === 'Enter' || e.key === ' ' ? selectBox('left') : null}
    >
      {#if isLoading}
        <div class="placeholder-content">
          <p>Loading content...</p>
        </div>
      {:else if loadError}
        <div class="placeholder-content error">
          <p>Failed to load content</p>
        </div>
      {:else}
        <div class="content">
          <h3>Option 1</h3>
          <div class="script-content">{@html option1Content.replace(/\n/g, '<br>')}</div>
        </div>
      {/if}
    </div>
    
    <div 
      class="content-box {selectedBox === 'right' ? 'selected' : ''}" 
      role="button"
      tabindex="0"
      on:click={() => selectBox('right')}
      on:keydown={(e) => e.key === 'Enter' || e.key === ' ' ? selectBox('right') : null}
    >
      {#if isLoading}
        <div class="placeholder-content">
          <p>Loading content...</p>
        </div>
      {:else if loadError}
        <div class="placeholder-content error">
          <p>Failed to load content</p>
        </div>
      {:else}
        <div class="content">
          <h3>Option 2</h3>
          <div class="script-content">{@html option2Content.replace(/\n/g, '<br>')}</div>
        </div>
      {/if}
    </div>
  </div>
  
  <div class="actions">
    <button type="button" on:click={goBackToHome} class="back-button">
      Back
    </button>
    
    <button type="button" class="use-selected" disabled={!selectedBox}>
      Use Selected
    </button>
  </div>
  
  {#if sessionId}
    <div class="session-info">Using session: {sessionId}</div>
  {/if}
</main>

<style>
  main {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
  }
  
  h1 {
    margin-bottom: 2rem;
    text-align: center;
  }
  
  .content-container {
    display: flex;
    flex-direction: row;
    gap: 2rem;
    margin-bottom: 2rem;
  }
  
  @media (max-width: 768px) {
    .content-container {
      flex-direction: column;
    }
  }
  
  .content-box {
    flex: 1;
    min-height: 300px;
    border: 2px solid #ddd;
    border-radius: 8px;
    padding: 1.5rem;
    cursor: pointer;
    transition: all 0.2s ease;
    background-color: #f9f9f9;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .content-box:hover {
    border-color: #bbb;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  }
  
  .content-box.selected {
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
    background-color: #eef2ff;
  }
  
  .placeholder-content {
    text-align: center;
  }
  
  .placeholder-content h3 {
    margin-bottom: 0.5rem;
  }
  
  .actions {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
  }
  
  button {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .back-button {
    background-color: #e5e7eb;
    color: #374151;
  }
  
  .back-button:hover {
    background-color: #d1d5db;
  }
  
  .use-selected {
    background-color: #3b82f6;
    color: white;
  }
  
  .use-selected:hover:not([disabled]) {
    background-color: #2563eb;
  }
  
  .use-selected[disabled] {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .session-info {
    margin-top: 1rem;
    font-size: 0.875rem;
    color: #6b7280;
    text-align: center;
  }
  
  .content {
    width: 100%;
  }
  
  .script-content {
    text-align: left;
    white-space: pre-wrap;
    font-family: inherit;
    line-height: 1.5;
    max-height: 400px;
    overflow-y: auto;
  }
  
  .error {
    color: #ef4444;
  }
</style>
