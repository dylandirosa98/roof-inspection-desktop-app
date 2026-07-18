<script>
  import {CreateProject} from '../wailsjs/go/main/App.js'
  import {PickDirectory} from "../wailsjs/go/main/App.js";
  import {RetrieveProject} from "../wailsjs/go/main/App.js";
  import {GetProjects} from "../wailsjs/go/main/App.js";
  import { onMount } from 'svelte'
  import logo1 from './assets/images/spartan-logo-dark.png'
  let resultText = "Please enter a directory below 👇"
  let directory = ''
  let name = ''
  let project = null
  let id = null
  let images = []
  let projects = []
  function createProject() {
    CreateProject(directory, name).then(result => {project = result
      resultText = JSON.stringify(result, null, 2)
      id = result.ID
      RetrieveProject(id).then(imageResults => {
        images = imageResults
      })
      getProjects()
    })
  }
  function hasImageData(image) {
    return image?.PreviewUrl.String && image?.PreviewUrl.String.length > 0;
  }

  async function chooseDirectory() {
    const selected = await PickDirectory()
    if (selected) {
      directory = selected
    }
  }
  async function getProjects() {
    GetProjects().then(results => {
      projects = results
    })
  }
  onMount(() => {
    getProjects()
  })

  function selectProject(selectedProject) {
    project = selectedProject
    id = selectedProject.ID
    RetrieveProject(id).then(imageResults => {images = imageResults})
  }
</script>
<div class="h-screen grid grid-cols-[22%_78%] bg-[#4B5563] text-white">
  <!-- Left Sidebar -->
  <aside class="border-r border-slate-800 p-4 bg-black p-4 flex flex-col">
    <div class="flex justify-center mt-[-40px]">
      <img src={logo1} alt="logo" class="w-[220px] h-[220px] object-contain"/>
    </div>
    <hr class="mt-[-10px]">
    <div>
      <h2 class="mt-[10px] mb-[8px] text-xl text-slate-400 mb">
        All Projects
      </h2>
      <!-- Scrollable project cards -->
      <div class="flex-1 overflow-y-auto space-y-3 pr-1">
        {#each projects as project}
          <button on:click={() => selectProject(project)} class="w-full text-left p-3 rounded-lg bg-[#1f2937] hover:bg[#374151]">
            <p class="font-semibold">{project.Name}</p>
            <p class="text-sm text-gray-500">{project.ImageCount} images</p>
          </button>
        {/each}
      </div>
    </div>
  </aside>
  <!-- main part of the dashboard -->
  <main style="background-color: #4B5563;">
    <div class="input-box" id="input">
      <div class="form-group">
        <label for="project-name">Project Name</label>
        <input
                autocomplete="off"
                bind:value={name}
                class="input text-black"
                id="project-name"
                type="text"
        />
      </div>

      <div class="form-group">
        <label for="project-directory">Directory</label>
        <div class="directory-row">
          <input
                  autocomplete="off"
                  bind:value={directory}
                  class="input text-black"
                  id="project-directory"
                  type="text"
                  readonly
          />
          <button class="btn" on:click={chooseDirectory}>Choose Folder</button>
        </div>
      </div>

      <div class="action-row">
        <button class="btn" on:click={createProject}>Create Project</button>
      </div>
    </div>
    {#if project}
      <div class="comparison-shell">
        <div class="image-section">
          <div class="tile-grid">
            {#each images as image}
              <div class="tile">
                {#if hasImageData(image)}
                  <img
                          src={image.PreviewUrl.String}
                          alt={image.Path}
                          class="tile-image"
                  />
                {:else}
                  <div class="tile-empty">
                    {image.Path}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>

        <div class="image-section">
          <div class="tile-grid">
            <!-- later -->
          </div>
        </div>
      </div>
    {/if}
  </main>
</div>

<style>

  .input-box .btn {
    width: auto;
    height: 30px;
    line-height: 30px;
    border-radius: 3px;
    border: none;
    margin: 0 0 0 20px;
    padding: 0 8px;
    cursor: pointer;
    background-color: white;
    color: black;
  }

  .input-box .btn:hover {
    background-image: linear-gradient(to top, #cfd9df 0%, #e2ebf0 100%);
    color: #333333;
  }

  .input-box .input {
    border: none;
    border-radius: 3px;
    outline: none;
    height: 30px;
    line-height: 30px;
    padding: 0 10px;
    background-color: rgba(240, 240, 240, 1);
    -webkit-font-smoothing: antialiased;
    color: black;
  }

  .input-box .input:hover {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  }

  .input-box .input:focus {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  }

  .tile-image {
    width: 100%;
    height: 100%;
    object-fit: contain;
    display: block;
  }

  .tile-empty {
    padding: 12px;
    text-align: center;
    color: #9ca3af;
    font-size: 14px;
    word-break: break-word;
  }

  .tile {
    aspect-ratio: 1 / 1;
    overflow: hidden;
    background-color: #0f172a;
    border: 1px solid #334155;
    display: flex;
    align-items: center;
    justify-content: center;

  }

  .comparison-shell {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    padding: 16px;
    align-items: start;
  }

  .tile-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 60px;
    padding: 16px;
  }

  .image-section {
    background: transparent;
    padding: 16px;
  }

  :global(body) {
    overflow: hidden;
  }

  aside {
    overflow-y: auto;
    min-height: 0;
  }

  main {
    overflow-y: auto;
    min-height: 0;
  }

  .input-box {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 20px;
    max-width: 700px;
    margin: 0 auto;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .directory-row {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .directory-row input {
    flex: 1;
  }

  .action-row {
    display: flex;
    justify-content: center;
  }
</style>
