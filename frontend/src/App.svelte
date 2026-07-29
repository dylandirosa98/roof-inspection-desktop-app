<script>
  import {
    AnalyzeProject,
    CreateProject,
    GetOriginalImageDataURL,
    GetProjects,
    PickDirectory,
    RetrieveAiImages,
  } from '../wailsjs/go/main/App.js'
  import { onMount } from 'svelte'
  import logo1 from './assets/images/spartan-logo-dark.png'
  let directory = ''
  let name = ''
  let project = null
  let images = []
  let projects = []
  let isAnalyzing = false
  let analysisError = ''
  let visibleImageCount = 20
  let jsonByImageID = {}
  let openJsonImageID = null
  let selectedImage = null
  let selectedImageSource = ''
  let selectedImageError = ''
  let isLoadingSelectedImage = false

  $: projectHasAnalysis = images.some((image) => image.AiImageID?.Valid)
  $: canAnalyzeProject = project && images.length > 0 && !projectHasAnalysis
  $: visibleImages = images.slice(0, visibleImageCount)

  async function createProject() {
    const result = await CreateProject(directory, name)
    project = result
    await loadProjectImages(result.ID)
    await getProjects()
  }

  function hasImageData(image) {
    return image?.ImagePreviewUrl?.Valid && image.ImagePreviewUrl.String.length > 0
  }

  function parseAnalysis(image) {
    if (!image.AnnotationsJson?.Valid) {
      return null
    }

    try {
      return JSON.parse(image.AnnotationsJson.String)
    } catch {
      return null
    }
  }

  function detectionCount(image) {
    return image.analysis?.OriginalImageBoxes?.length ?? 0
  }

  function previewBox(boundingBox, image) {
    const previewSize = 250
    const scale = Math.max(
      previewSize / image.ImageWidth.Int64,
      previewSize / image.ImageHeight.Int64,
    )
    const xOffset = (previewSize - image.ImageWidth.Int64 * scale) / 2
    const yOffset = (previewSize - image.ImageHeight.Int64 * scale) / 2

    return {
      left: boundingBox.Left * scale + xOffset,
      top: boundingBox.Top * scale + yOffset,
      width: (boundingBox.Right - boundingBox.Left) * scale,
      height: (boundingBox.Bottom - boundingBox.Top) * scale,
    }
  }

  async function loadProjectImages(projectID) {
    const rows = await RetrieveAiImages(projectID)
    images = rows
      .map((image) => ({ ...image, analysis: parseAnalysis(image) }))
      .sort((first, second) => detectionCount(second) - detectionCount(first))
    visibleImageCount = 20
    jsonByImageID = {}
    openJsonImageID = null
  }

  function toggleJson(image) {
    if (openJsonImageID === image.ImageID) {
      openJsonImageID = null
      return
    }

    if (!jsonByImageID[image.ImageID]) {
      jsonByImageID = {
        ...jsonByImageID,
        [image.ImageID]: JSON.stringify(image.analysis, null, 2),
      }
    }
    openJsonImageID = image.ImageID
  }

  async function openImage(image, showDetections) {
    selectedImage = { ...image, showDetections }
    selectedImageSource = ''
    selectedImageError = ''
    isLoadingSelectedImage = true

    try {
      const source = await GetOriginalImageDataURL(image.ImageID)
      if (selectedImage?.ImageID === image.ImageID) {
        selectedImageSource = source
      }
    } catch (error) {
      if (selectedImage?.ImageID === image.ImageID) {
        selectedImageError = error.message || String(error)
      }
    } finally {
      if (selectedImage?.ImageID === image.ImageID) {
        isLoadingSelectedImage = false
      }
    }
  }

  function closeImage() {
    selectedImage = null
    selectedImageSource = ''
    selectedImageError = ''
  }

  async function chooseDirectory() {
    const selected = await PickDirectory()
    if (selected) {
      directory = selected
    }
  }
  async function getProjects() {
    projects = await GetProjects()
  }

  onMount(() => {
    getProjects()
  })

  async function selectProject(selectedProject) {
    project = selectedProject
    analysisError = ''
    await loadProjectImages(selectedProject.ID)
  }

  async function analyzeProject() {
    isAnalyzing = true
    analysisError = ''

    try {
      await AnalyzeProject({ ID: project.ID, Name: project.Name, Directory: project.Directory })
      await loadProjectImages(project.ID)
    } catch (error) {
      analysisError = error.message || String(error)
    } finally {
      isAnalyzing = false
    }
  }

  function clearProject(){
    project = null
    images = []
    analysisError = ''
    visibleImageCount = 20
    jsonByImageID = {}
    openJsonImageID = null
  }
</script>
<div class="h-screen grid grid-cols-[22%_78%] bg-[#4B5563] text-white">
  <!-- Left Sidebar -->
  <aside class="border-r border-slate-800 p-4 bg-black p-4 flex flex-col">
    <div class="flex justify-center mt-[-40px]">
      <img src={logo1} alt="logo" class="w-[240px] h-[240px] object-contain"/>
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
    {#if project}
      <div class="current-project relative flex items-center justify-center p-4 mt-[10px] mb-[-10px]">
        <h2 class="font-sans text-3xl font-bold tracking-wide">{project.Name}</h2>

        <button
                class="absolute right-12 rounded-md bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700"
                on:click={clearProject}
        >
          Exit Project
        </button>
        {#if canAnalyzeProject}
          <button
                  class="absolute right-48 rounded-md bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700 disabled:opacity-50"
                  on:click={analyzeProject}
                  disabled={isAnalyzing}
          >
            {isAnalyzing ? 'Analyzing...' : 'Analyze Project'}
          </button>
        {/if}
      </div>
      {#if analysisError}
        <p class="analysis-error">{analysisError}</p>
      {/if}
    {:else}
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
      {/if}
      {#if project}
        <div class="comparison-shell">
          {#each visibleImages as image (image.ImageID)}
            <div class="comparison-row">
              <div class="image-section">
              <div class="tile">
                {#if hasImageData(image)}
                  <button class="image-button" type="button" on:click={() => openImage(image, false)}>
                    <img
                             src={image.ImagePreviewUrl.String}
                             alt={image.ImagePath}
                             class="tile-image"
                    />
                  </button>
                {:else}
                  <button class="image-button" type="button" on:click={() => openImage(image, false)}>
                    <div class="tile-empty">{image.ImagePath}</div>
                  </button>
                {/if}
              </div>
              </div>

              <div class="image-section">
              <div class="tile">
                <button class="image-button" type="button" on:click={() => openImage(image, true)}>
                  {#if image.analysis && detectionCount(image) > 0 && hasImageData(image)}
                    <svg
                            class="tile-image"
                            viewBox="0 0 250 250"
                            aria-label={`AI detections for ${image.ImagePath}`}
                    >
                      <image
                              href={image.ImagePreviewUrl.String}
                              width="250"
                              height="250"
                      />
                      {#each image.analysis.OriginalImageBoxes as detection}
                        <rect
                                x={previewBox(detection.BoundingBox, image).left}
                                y={previewBox(detection.BoundingBox, image).top}
                                width={previewBox(detection.BoundingBox, image).width}
                                height={previewBox(detection.BoundingBox, image).height}
                                class="detection-box"
                        />
                      {/each}
                    </svg>
                  {:else if image.analysis && detectionCount(image) === 0}
                    <div class="tile-empty">No detections</div>
                  {:else if image.analysis}
                    <div class="tile-empty">Unable to load image</div>
                  {:else}
                    <div class="tile-empty">Analyze project to view results</div>
                  {/if}
                </button>
              </div>
              </div>
              {#if image.analysis}
                <div class="result-actions">
                  <div class="result-toolbar">
                    <button class="json-toggle" type="button" on:click={() => toggleJson(image)}>
                      {openJsonImageID === image.ImageID ? 'Hide JSON' : 'View JSON'}
                    </button>
                    <button class="edit-button" type="button">Edit Boxes</button>
                  </div>
                  {#if openJsonImageID === image.ImageID}
                    <pre class="json-panel">{jsonByImageID[image.ImageID]}</pre>
                  {/if}
                </div>
              {/if}
            </div>
          {/each}
          {#if visibleImageCount < images.length}
            <button class="load-more" on:click={() => visibleImageCount += 20}>
              Load more images
            </button>
          {/if}
        </div>
      {/if}
  </main>
</div>

{#if selectedImage}
  <div class="image-modal" role="dialog" aria-modal="true" aria-label="Full image view">
    <button class="modal-close" type="button" on:click={closeImage}>Close</button>
    <div class="modal-content">
      {#if isLoadingSelectedImage}
        <div class="tile-empty">Loading original image...</div>
      {:else if selectedImageError}
        <div class="tile-empty">{selectedImageError}</div>
      {:else if selectedImageSource}
        <svg
                class="full-image"
                viewBox={`0 0 ${selectedImage.ImageWidth.Int64} ${selectedImage.ImageHeight.Int64}`}
                preserveAspectRatio="xMidYMid meet"
                aria-label={selectedImage.ImagePath}
        >
          <image
                  href={selectedImageSource}
                  width={selectedImage.ImageWidth.Int64}
                  height={selectedImage.ImageHeight.Int64}
          />
          {#if selectedImage.showDetections}
            {#each selectedImage.analysis?.OriginalImageBoxes || [] as detection}
              <rect
                      x={detection.BoundingBox.Left}
                      y={detection.BoundingBox.Top}
                      width={detection.BoundingBox.Right - detection.BoundingBox.Left}
                      height={detection.BoundingBox.Bottom - detection.BoundingBox.Top}
                      class="detection-box"
              />
            {/each}
          {/if}
        </svg>
      {/if}
    </div>
  </div>
{/if}

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

  .image-button {
    width: 100%;
    height: 100%;
    padding: 0;
    border: 0;
    background: transparent;
    color: inherit;
    cursor: zoom-in;
  }

  .tile-empty {
    padding: 12px;
    text-align: center;
    color: #9ca3af;
    font-size: 14px;
    word-break: break-word;
  }

  .detection-box {
    fill: transparent;
    stroke: #ef4444;
    stroke-width: 3;
    vector-effect: non-scaling-stroke;
  }

  .result-actions {
    grid-column: 1 / -1;
    margin: -1px 16px 0;
    border: 1px solid #334155;
    background: #1f2937;
  }

  .result-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 38px;
    padding: 0 10px;
  }

  .json-toggle,
  .edit-button {
    border: 0;
    border-radius: 3px;
    padding: 6px 10px;
    cursor: pointer;
    font-size: 12px;
  }

  .json-toggle {
    background: transparent;
    color: #d1d5db;
  }

  .edit-button {
    background: #dc2626;
    color: white;
  }

  .json-panel {
    max-height: 180px;
    overflow: auto;
    margin: 0;
    padding: 8px;
    border-top: 1px solid #334155;
    background: #0f172a;
    text-align: left;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .analysis-error {
    margin: 0 16px;
    color: #fecaca;
    text-align: center;
  }

  .load-more {
    align-self: center;
    padding: 8px 16px;
    border: 0;
    border-radius: 3px;
    background: white;
    color: black;
    cursor: pointer;
  }

  .image-modal {
    position: fixed;
    inset: 0;
    z-index: 10;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgba(0, 0, 0, 0.85);
  }

  .modal-close {
    align-self: flex-end;
    margin-bottom: 12px;
    border: 0;
    border-radius: 3px;
    padding: 8px 12px;
    background: #dc2626;
    color: white;
    cursor: pointer;
  }

  .modal-content {
    display: flex;
    width: min(1100px, 100%);
    height: min(80vh, 800px);
    align-items: center;
    justify-content: center;
    background: #0f172a;
  }

  .full-image {
    width: 100%;
    height: 100%;
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
    display: flex;
    flex-direction: column;
    gap: 60px;
    padding: 16px;
    align-items: start;
  }

  .comparison-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    column-gap: 16px;
    row-gap: 0;
    width: 100%;
    align-items: start;
  }

  .image-section {
    background: transparent;
    padding: 16px 16px 0;
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
