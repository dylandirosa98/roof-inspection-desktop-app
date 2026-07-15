<script>
  import {CreateProject} from '../wailsjs/go/main/App.js'
  import {PickDirectory} from "../wailsjs/go/main/App.js";
  import {RetrieveProject} from "../wailsjs/go/main/App.js";
  import logo1 from './assets/images/Hail_Scan_Logo_Dark_Mode-removebg-preview.png'
  let resultText = "Please enter a directory below 👇"
  let directory = ''
  let name = ''
  let project = null
  let id = null
  let images = []
  function createProject() {
    CreateProject(directory, name).then(result => {project = result
      resultText = JSON.stringify(result, null, 2)
      id = result.ID
      RetrieveProject(id).then(imageResults => {
        images = imageResults
      })
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
</script>
<div class="h-screen grid grid-cols-[22%_78%] bg-[#4B5563] text-white">
  <!-- Left Sidebar -->
  <aside class="border-r border-slate-800 p-4 bg-black p-4 flex flex-col">
    <div class="flex justify-center mt-[-80px]">
      <img src={logo1} alt="logo" class="w-[290px] h-[290px] flex flex-col items-center gap-2"/>
    </div>
    <hr class="mt-[-50px]">
    <div>
      <h2 class="mt-[10px] mb-[8px] text-xl text-slate-400 mb">
        All Projects
      </h2>
      <!-- Scrollable project cards -->
      <div class="flex-1 overflow-y-auto space-y-3 pr-1">
        <button class="w-full text-left p-3 rounded-lg bg-[#1f2937] hover:bg[#374151]">
          <p class="font-semibold">Example House</p>
          <p class="text-sm text-gray-500">48 images</p>
        </button>

        <button class="w-full text-left p-3 rounded-lg bg-[#1f2937] hover:bg[#374151]">
          <p class="font-semibold">Example House 2</p>
          <p class="text-sm text-gray-500">27 images</p>
        </button>
      </div>
    </div>
  </aside>
  <!-- main part of the dashboard -->
  <main style="background-color: #4B5563;">
    <div class="input-box" id="input">
      <label for="project-name">Project Name</label>
      <input autocomplete="off" bind:value={name} class="input text-black mt-[20px]" id="project-name" type="text"/>

      <label for="project-directory" class="mt-[10px] block">Directory</label>
      <div class="flex gap-2 mt-[8px]">
        <input autocomplete="off" bind:value={directory} class="input text-black" id="project-directory" type="text" readonly/>
        <button class="btn" on:click={chooseDirectory}>Choose Folder</button>
      </div>

      <button class="btn mt-[20px]" on:click={createProject}>Create Project</button>
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
</style>
