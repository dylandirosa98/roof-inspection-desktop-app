<script>
  import {
    AnalyzeProject,
    ApproveImageReview,
    CreateInspectionReport,
    CreateProject,
    DeleteProject,
    GenerateInspectionReport,
    GetInspectionReportByProjectID,
    GetOriginalImageDataURL,
    GetProjects,
    OpenLastGeneratedInspectionReport,
    PickDirectory,
    RetrieveAiImages,
    SaveEditedAnnotations,
    UpdateInspectionReport,
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
  let projectError = ''
  let isChoosingDirectory = false
  let isCreatingProject = false
  let isDeletingProjectID = null
  let projectListError = ''
  let visibleImageCount = 20
  let jsonByImageID = {}
  let openJsonImageID = null
  let openAnnotationImageID = null
  let annotationNotesByImageID = {}
  let isSavingAnnotationImageID = null
  let annotationError = ''
  let isApprovingImageID = null
  let reviewErrorImageID = null
  let reviewError = ''
  let selectedImage = null
  let selectedImageSource = ''
  let selectedImageError = ''
  let isLoadingSelectedImage = false
  let editedDetectionsByImageID = {}
  let selectedDetectionIndex = null
  let editPointer = null
  let fullImageSvg
  let projectView = 'photos'
  let report = null
  let reportDraft = null
  let reportState = 'idle'
  let reportProjectID = null
  let reportError = ''
  let reportMessage = ''
  let isSavingReport = false
  let isGeneratingReport = false
  let mainContent
  let showScrollToTop = false

  const reportTitle = 'Roof Inspection and Photo Documentation Report'
  const editableReportFields = [
    'ReportNumber',
    'ReportTitle',
    'CustomerName',
    'PropertyAddress',
    'PropertyCity',
    'PropertyState',
    'PropertyZip',
    'InspectorName',
    'InspectionDate',
    'InsuranceCarrier',
    'ClaimNumber',
    'DateOfLoss',
    'Summary',
  ]

  $: projectHasAnalysis = images.some((image) => image.AiImageID?.Valid)
  $: canAnalyzeProject = project && images.length > 0 && !projectHasAnalysis
  $: visibleImages = images.slice(0, visibleImageCount)
  $: approvedReviewCount = images.filter(isReviewApproved).length
  $: approvedDamagePhotoCount = images.filter(
    (image) => isReviewApproved(image) && finalDetectionCount(image) > 0,
  ).length
  $: approvedNoDamageCount = approvedReviewCount - approvedDamagePhotoCount
  $: unreviewedPhotoCount = images.length - approvedReviewCount
  $: reportIsDirty = Boolean(report && reportDraft && editableReportFields.some(
    (field) => (reportDraft[field] || '') !== (report[field] || ''),
  ))

  async function createProject() {
    if (!name.trim()) {
      projectError = 'Enter a project name.'
      return
    }
    if (!directory) {
      projectError = 'Choose the folder containing the roof photos.'
      return
    }

    isCreatingProject = true
    projectError = ''
    try {
      const result = await CreateProject(directory, name.trim())
      project = result
      projectView = 'photos'
      resetReportState()
      await loadProjectImages(result.ID)
      await getProjects()
    } catch (error) {
      projectError = errorText(error)
    } finally {
      isCreatingProject = false
    }
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

  function parseEditedAnnotations(image) {
    if (!image.EditedAnnotationsJson?.Valid) {
      return null
    }

    try {
      return JSON.parse(image.EditedAnnotationsJson.String)
    } catch {
      return null
    }
  }

  function detectionCount(image) {
    return detectionsFor(image).length
  }

  function isReviewApproved(image) {
    if (image.ReviewApproved?.Valid) {
      return image.ReviewApproved.Int64 === 1
    }
    return image.ReviewApproved === 1 || image.ReviewApproved === true
  }

  function finalDetectionCount(image) {
    return image.editedAnnotations?.detections?.length || 0
  }

  function detectionsFor(image) {
    return editedDetectionsByImageID[image.ImageID]
      || image.editedAnnotations?.detections
      || image.analysis?.OriginalImageBoxes
      || []
  }

  function cloneDetections(detections) {
    return JSON.parse(JSON.stringify(detections))
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
      .map((image) => ({
        ...image,
        analysis: parseAnalysis(image),
        editedAnnotations: parseEditedAnnotations(image),
      }))
      .sort((first, second) => detectionCount(second) - detectionCount(first))
    visibleImageCount = 20
    jsonByImageID = {}
    openJsonImageID = null
    openAnnotationImageID = null
    editedDetectionsByImageID = {}
    isApprovingImageID = null
    reviewErrorImageID = null
    reviewError = ''
    annotationNotesByImageID = Object.fromEntries(images.map((image) => [
      image.ImageID,
      image.editedAnnotations?.notes || '',
    ]))
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

  function toggleAnnotations(image) {
    openAnnotationImageID = openAnnotationImageID === image.ImageID ? null : image.ImageID
  }

  async function saveAnnotations(image) {
    const annotations = {
      detections: detectionsFor(image),
      notes: annotationNotesByImageID[image.ImageID] || '',
    }

    isSavingAnnotationImageID = image.ImageID
    annotationError = ''
    try {
      await SaveEditedAnnotations(image.ImageID, JSON.stringify(annotations))
      images = images.map((currentImage) => currentImage.ImageID === image.ImageID
        ? {
          ...currentImage,
          editedAnnotations: annotations,
          ReviewApproved: { Int64: 0, Valid: true },
          ReviewedAt: { String: '', Valid: false },
        }
        : currentImage)
      if (selectedImage?.ImageID === image.ImageID) {
        selectedImage = {
          ...selectedImage,
          editedAnnotations: annotations,
          ReviewApproved: { Int64: 0, Valid: true },
          ReviewedAt: { String: '', Valid: false },
        }
      }
      return true
    } catch (error) {
      annotationError = error.message || String(error)
      return false
    } finally {
      isSavingAnnotationImageID = null
    }
  }

  async function approveReview(image) {
    const annotations = {
      detections: detectionsFor(image),
      notes: annotationNotesByImageID[image.ImageID] || '',
    }

    isApprovingImageID = image.ImageID
    reviewErrorImageID = null
    reviewError = ''
    try {
      await ApproveImageReview(image.ImageID, JSON.stringify(annotations))
      const reviewedAt = { String: new Date().toISOString(), Valid: true }
      images = images.map((currentImage) => currentImage.ImageID === image.ImageID
        ? {
          ...currentImage,
          editedAnnotations: annotations,
          ReviewApproved: { Int64: 1, Valid: true },
          ReviewedAt: reviewedAt,
        }
        : currentImage)
      if (selectedImage?.ImageID === image.ImageID) {
        selectedImage = {
          ...selectedImage,
          editedAnnotations: annotations,
          ReviewApproved: { Int64: 1, Valid: true },
          ReviewedAt: reviewedAt,
        }
      }
    } catch (error) {
      reviewErrorImageID = image.ImageID
      reviewError = errorText(error)
    } finally {
      isApprovingImageID = null
    }
  }

  async function openImage(image, showDetections, editing = false) {
    selectedImage = { ...image, showDetections, editing }
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
    if (selectedImage?.editing) {
      cancelEdits()
    }
    selectedImage = null
    selectedImageSource = ''
    selectedImageError = ''
    selectedDetectionIndex = null
    editPointer = null
  }

  function startEditing(image) {
    if (!editedDetectionsByImageID[image.ImageID]) {
      editedDetectionsByImageID = {
        ...editedDetectionsByImageID,
        [image.ImageID]: cloneDetections(detectionsFor(image)),
      }
    }
    selectedDetectionIndex = null
    openImage(image, true, true)
  }

  function stopEditing() {
    if (selectedImage) {
      selectedImage = { ...selectedImage, editing: false }
    }
    selectedDetectionIndex = null
    editPointer = null
  }

  function cancelEdits() {
    const detections = cloneDetections(
      selectedImage.editedAnnotations?.detections || selectedImage.analysis?.OriginalImageBoxes || [],
    )
    editedDetectionsByImageID = {
      ...editedDetectionsByImageID,
      [selectedImage.ImageID]: detections,
    }
    images = images.map((image) => image.ImageID === selectedImage.ImageID
      ? { ...image, editedAnnotations: { ...image.editedAnnotations, detections } }
      : image)
    stopEditing()
  }

  function imagePoint(event) {
    if (!fullImageSvg || !selectedImage) {
      return null
    }

    const bounds = fullImageSvg.getBoundingClientRect()
    const imageWidth = selectedImage.ImageWidth.Int64
    const imageHeight = selectedImage.ImageHeight.Int64
    const scale = Math.min(bounds.width / imageWidth, bounds.height / imageHeight)
    const offsetX = (bounds.width - imageWidth * scale) / 2
    const offsetY = (bounds.height - imageHeight * scale) / 2

    return {
      x: Math.max(0, Math.min(imageWidth, (event.clientX - bounds.left - offsetX) / scale)),
      y: Math.max(0, Math.min(imageHeight, (event.clientY - bounds.top - offsetY) / scale)),
    }
  }

  function updateDetection(index, detection) {
    const detections = [...detectionsFor(selectedImage)]
    detections[index] = detection
    editedDetectionsByImageID = {
      ...editedDetectionsByImageID,
      [selectedImage.ImageID]: detections,
    }
    images = images.map((image) => image.ImageID === selectedImage.ImageID
      ? { ...image, editedAnnotations: { ...image.editedAnnotations, detections } }
      : image)
    selectedImage = { ...selectedImage }
  }

  function createDetection(left, top, right, bottom) {
    return {
      BoundingBox: { Left: left, Top: top, Right: right, Bottom: bottom },
      Detection: {
        Class: 'manual-annotation',
        Confidence: 1,
        X: (left + right) / 2,
        Y: (top + bottom) / 2,
        Width: right - left,
        Height: bottom - top,
      },
    }
  }

  function startMove(event, index) {
    const point = imagePoint(event)
    if (!point) return

    event.preventDefault()
    selectedDetectionIndex = index
    editPointer = {
      action: 'move',
      index,
      point,
      box: { ...detectionsFor(selectedImage)[index].BoundingBox },
    }
  }

  function startResize(event, index, corner) {
    const point = imagePoint(event)
    if (!point) return

    event.preventDefault()
    selectedDetectionIndex = index
    editPointer = {
      action: 'resize',
      index,
      corner,
      point,
      box: { ...detectionsFor(selectedImage)[index].BoundingBox },
    }
  }

  function startDraw(event) {
    if (!selectedImage?.editing) return

    const point = imagePoint(event)
    if (!point) return

    event.preventDefault()
    const index = detectionsFor(selectedImage).length
    updateDetection(index, createDetection(point.x, point.y, point.x, point.y))
    selectedDetectionIndex = index
    editPointer = { action: 'draw', index, point }
  }

  function movePointer(event) {
    if (!editPointer) return

    const point = imagePoint(event)
    if (!point) return

    if (editPointer.action === 'draw') {
      updateDetection(editPointer.index, createDetection(
        Math.min(editPointer.point.x, point.x),
        Math.min(editPointer.point.y, point.y),
        Math.max(editPointer.point.x, point.x),
        Math.max(editPointer.point.y, point.y),
      ))
      return
    }

    const { box } = editPointer
    let nextBox
    if (editPointer.action === 'move') {
      const width = box.Right - box.Left
      const height = box.Bottom - box.Top
      const left = Math.max(0, Math.min(selectedImage.ImageWidth.Int64 - width, box.Left + point.x - editPointer.point.x))
      const top = Math.max(0, Math.min(selectedImage.ImageHeight.Int64 - height, box.Top + point.y - editPointer.point.y))
      nextBox = { Left: left, Top: top, Right: left + width, Bottom: top + height }
    } else {
      let left = editPointer.corner.includes('w') ? point.x : box.Left
      let right = editPointer.corner.includes('e') ? point.x : box.Right
      let top = editPointer.corner.includes('n') ? point.y : box.Top
      let bottom = editPointer.corner.includes('s') ? point.y : box.Bottom
      nextBox = {
        Left: Math.min(left, right),
        Top: Math.min(top, bottom),
        Right: Math.max(left, right),
        Bottom: Math.max(top, bottom),
      }
    }

    updateDetection(editPointer.index, {
      ...detectionsFor(selectedImage)[editPointer.index],
      BoundingBox: nextBox,
    })
  }

  function endPointer(event) {
    if (!editPointer) return
    editPointer = null
  }

  function deleteSelectedDetection() {
    if (selectedDetectionIndex === null) return

    const detections = [...detectionsFor(selectedImage)]
    detections.splice(selectedDetectionIndex, 1)
    editedDetectionsByImageID = {
      ...editedDetectionsByImageID,
      [selectedImage.ImageID]: detections,
    }
    selectedDetectionIndex = null
  }

  async function saveEdits() {
    if (await saveAnnotations(selectedImage)) {
      stopEditing()
    }
  }

  async function chooseDirectory() {
    isChoosingDirectory = true
    projectError = ''
    try {
      const selected = await PickDirectory()
      if (selected) {
        directory = selected
      }
    } catch (error) {
      projectError = errorText(error)
    } finally {
      isChoosingDirectory = false
    }
  }
  async function getProjects() {
    projects = await GetProjects()
  }

  onMount(() => {
    getProjects()
  })

  function errorText(error) {
    return error?.message || String(error)
  }

  function resetReportState() {
    report = null
    reportDraft = null
    reportState = 'idle'
    reportProjectID = null
    reportError = ''
    reportMessage = ''
    isSavingReport = false
    isGeneratingReport = false
  }

  function setReport(currentReport) {
    report = currentReport
    reportDraft = {
      ...currentReport,
      ReportTitle: currentReport.ReportTitle || reportTitle,
    }
    reportState = 'ready'
    reportProjectID = currentReport.ProjectID
  }

  function canLeaveReport() {
    return !reportIsDirty || window.confirm('Discard unsaved report changes?')
  }

  async function loadInspectionReport(projectID, force = false) {
    if (!force && reportProjectID === projectID && reportState !== 'idle') {
      return
    }

    reportState = 'loading'
    reportProjectID = projectID
    reportError = ''
    reportMessage = ''
    try {
      setReport(await GetInspectionReportByProjectID(projectID))
    } catch (error) {
      const message = errorText(error)
      report = null
      reportDraft = null
      if (message.toLowerCase().includes('no rows')) {
        reportState = 'missing'
      } else {
        reportState = 'error'
        reportError = message
      }
    }
  }

  async function showProjectView(view) {
    if (view === projectView) return
    if (projectView === 'report' && !canLeaveReport()) return

    projectView = view
    scrollMainToTop('auto')
    reportError = ''
    reportMessage = ''
    if (view === 'report') {
      await loadInspectionReport(project.ID)
    }
  }

  async function createReport() {
    reportState = 'loading'
    reportError = ''
    reportMessage = ''
    try {
      const created = await CreateInspectionReport(project.ID, `SP-${String(project.ID).padStart(4, '0')}`)
      setReport(created)
      reportMessage = 'Report created. Add the inspection details below.'
    } catch (error) {
      reportState = 'missing'
      reportError = errorText(error)
    }
  }

  function reportUpdateParams() {
    return {
      ID: reportDraft.ID,
      ReportNumber: reportDraft.ReportNumber.trim(),
      ReportTitle: reportTitle,
      CustomerName: reportDraft.CustomerName.trim(),
      PropertyAddress: reportDraft.PropertyAddress.trim(),
      PropertyCity: reportDraft.PropertyCity.trim(),
      PropertyState: reportDraft.PropertyState.trim().toUpperCase(),
      PropertyZip: reportDraft.PropertyZip.trim(),
      InspectorName: reportDraft.InspectorName.trim(),
      InspectionDate: reportDraft.InspectionDate,
      InsuranceCarrier: reportDraft.InsuranceCarrier.trim(),
      ClaimNumber: reportDraft.ClaimNumber.trim(),
      DateOfLoss: reportDraft.DateOfLoss,
      Summary: reportDraft.Summary.trim(),
      Notes: '',
    }
  }

  async function saveReport(showConfirmation = true) {
    if (!reportDraft.ReportNumber.trim()) {
      reportError = 'Report number is required.'
      return null
    }

    isSavingReport = true
    reportError = ''
    reportMessage = ''
    try {
      const saved = await UpdateInspectionReport(reportUpdateParams())
      setReport(saved)
      if (showConfirmation) {
        reportMessage = 'Report changes saved.'
      }
      return saved
    } catch (error) {
      reportError = errorText(error)
      return null
    } finally {
      isSavingReport = false
    }
  }

  function safeFilenamePart(value) {
    return value.trim().replace(/[^a-z0-9._-]+/gi, '_').replace(/^_+|_+$/g, '') || 'report'
  }

  function reportOutputPath(savedReport) {
    const directoryPath = project.Directory.replace(/[\\/]+$/, '')
    const projectName = safeFilenamePart(project.Name)
    const reportNumber = safeFilenamePart(savedReport.ReportNumber)
    return `${directoryPath}/reports/Roof_Inspection_Report_${projectName}_${reportNumber}.pdf`
  }

  function validateReportForGeneration() {
    const missingFields = []
    if (!reportDraft.PropertyAddress.trim()) missingFields.push('property address')
    if (!reportDraft.InspectorName.trim()) missingFields.push('inspector')
    if (!reportDraft.InspectionDate) missingFields.push('inspection date')
    if (missingFields.length > 0) {
      reportError = `Complete the ${missingFields.join(', ')} before generating the PDF.`
      return false
    }
    if (approvedDamagePhotoCount === 0) {
      reportError = 'Approve at least one image with a final damage box before generating the PDF.'
      return false
    }
    return true
  }

  async function generateAndOpenReport() {
    reportError = ''
    reportMessage = ''
    if (!validateReportForGeneration()) return

    isGeneratingReport = true
    const saved = await saveReport(false)
    if (!saved) {
      isGeneratingReport = false
      return
    }

    try {
      const result = await GenerateInspectionReport(saved.ID, reportOutputPath(saved))
      await loadInspectionReport(project.ID, true)
      reportMessage = `Generated ${result.ImageCount} approved damage ${result.ImageCount === 1 ? 'photo' : 'photos'} across ${result.PageCount} ${result.PageCount === 1 ? 'page' : 'pages'}.`
      try {
        await OpenLastGeneratedInspectionReport(saved.ID)
      } catch (error) {
        reportError = `The PDF was generated, but it could not be opened: ${errorText(error)}`
      }
    } catch (error) {
      reportError = errorText(error)
    } finally {
      isGeneratingReport = false
    }
  }

  async function openLastReport() {
    reportError = ''
    reportMessage = ''
    try {
      await OpenLastGeneratedInspectionReport(report.ID)
    } catch (error) {
      reportError = errorText(error)
    }
  }

  function handleMainScroll() {
    showScrollToTop = (mainContent?.scrollTop || 0) > 500
  }

  function scrollMainToTop(behavior = 'smooth') {
    mainContent?.scrollTo({ top: 0, behavior })
    showScrollToTop = false
  }

  async function selectProject(selectedProject) {
    if (projectView === 'report' && !canLeaveReport()) return
    project = selectedProject
    projectView = 'photos'
    resetReportState()
    scrollMainToTop('auto')
    analysisError = ''
    await loadProjectImages(selectedProject.ID)
  }

  async function deleteProject(selectedProject) {
    const hasUnsavedReport = project?.ID === selectedProject.ID && projectView === 'report' && reportIsDirty
    const warning = hasUnsavedReport
      ? `Delete "${selectedProject.Name}" and discard its unsaved report changes? This removes the project, reviews, and report data from the app. Original photo files will not be deleted.`
      : `Delete "${selectedProject.Name}"? This removes the project, reviews, and report data from the app. Original photo files will not be deleted.`
    if (!window.confirm(warning)) return

    isDeletingProjectID = selectedProject.ID
    projectListError = ''
    try {
      await DeleteProject(selectedProject.ID)
      if (project?.ID === selectedProject.ID) {
        resetSelectedProject()
      }
      await getProjects()
    } catch (error) {
      projectListError = errorText(error)
    } finally {
      isDeletingProjectID = null
    }
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

  function resetSelectedProject() {
    project = null
    projectView = 'photos'
    resetReportState()
    scrollMainToTop('auto')
    images = []
    analysisError = ''
    visibleImageCount = 20
    jsonByImageID = {}
    openJsonImageID = null
    openAnnotationImageID = null
    annotationNotesByImageID = {}
    editedDetectionsByImageID = {}
    isApprovingImageID = null
    reviewErrorImageID = null
    reviewError = ''
  }

  function clearProject(){
    if (projectView === 'report' && !canLeaveReport()) return
    resetSelectedProject()
  }
</script>
<svelte:window on:mousemove={movePointer} on:mouseup={endPointer} />
<div class="app-shell h-screen grid grid-cols-[22%_78%] bg-[#4B5563] text-white">
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
        {#each projects as savedProject}
          <div class="project-list-item" class:selected={project?.ID === savedProject.ID}>
            <button class="project-select" on:click={() => selectProject(savedProject)}>
              <p class="font-semibold">{savedProject.Name}</p>
              <p class="text-sm text-gray-500">{savedProject.ImageCount} images</p>
            </button>
            <button
                    class="project-delete"
                    type="button"
                    title={`Delete ${savedProject.Name}`}
                    aria-label={`Delete ${savedProject.Name}`}
                    disabled={isDeletingProjectID === savedProject.ID}
                    on:click={() => deleteProject(savedProject)}
            >
              {isDeletingProjectID === savedProject.ID ? '...' : 'Delete'}
            </button>
          </div>
        {/each}
      </div>
      {#if projectListError}<p class="project-list-error">{projectListError}</p>{/if}
    </div>
  </aside>
  <!-- main part of the dashboard -->
  <main bind:this={mainContent} on:scroll={handleMainScroll} style="background-color: #4B5563;">
    {#if project}
      <div class="current-project relative flex items-center justify-center p-4 mt-[10px] mb-[-10px]">
        <h2 class="font-sans text-3xl font-bold tracking-wide">{project.Name}</h2>

        <button
                class="absolute right-12 rounded-md bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700"
                on:click={clearProject}
        >
          Exit Project
        </button>
        {#if projectView === 'photos' && canAnalyzeProject}
          <button
                  class="absolute right-48 rounded-md bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700 disabled:opacity-50"
                  on:click={analyzeProject}
                  disabled={isAnalyzing}
          >
            {isAnalyzing ? 'Analyzing...' : 'Analyze Project'}
          </button>
        {/if}
      </div>
      <div class="project-tabs" role="tablist" aria-label="Project views">
        <button
                type="button"
                role="tab"
                aria-selected={projectView === 'photos'}
                class:active={projectView === 'photos'}
                on:click={() => showProjectView('photos')}
        >
          Photo Review
        </button>
        <button
                type="button"
                role="tab"
                aria-selected={projectView === 'report'}
                class:active={projectView === 'report'}
                on:click={() => showProjectView('report')}
        >
          Report
          <span class="reviewed-count">{approvedReviewCount}/{images.length}</span>
        </button>
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
              <button class="btn" on:click={chooseDirectory} disabled={isChoosingDirectory}>
                {isChoosingDirectory ? 'Opening...' : 'Choose Folder'}
              </button>
            </div>
          </div>

          {#if projectError}
            <p class="project-error" role="alert">{projectError}</p>
          {/if}

          <div class="action-row">
            <button class="btn" on:click={createProject} disabled={isCreatingProject}>
              {isCreatingProject ? 'Creating...' : 'Create Project'}
            </button>
        </div>
      </div>
      {/if}
      {#if project && projectView === 'photos'}
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
                  {#if image.analysis && detectionsFor(image).length > 0 && hasImageData(image)}
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
                      {#each detectionsFor(image) as detection}
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
                    <div class="result-status">
                      <button class="json-toggle" type="button" on:click={() => toggleJson(image)}>
                        {openJsonImageID === image.ImageID ? 'Hide JSON' : 'View JSON'}
                      </button>
                      <span class={`review-badge ${isReviewApproved(image) ? (finalDetectionCount(image) > 0 ? 'damage' : 'no-damage') : ''}`}>
                        {#if isReviewApproved(image)}
                          Approved - {finalDetectionCount(image) > 0 ? 'Damage' : 'No damage'}
                        {:else}
                          Unreviewed
                        {/if}
                      </span>
                    </div>
                    <div class="toolbar-actions">
                      <button class="annotation-toggle" type="button" on:click={() => toggleAnnotations(image)}>
                        {openAnnotationImageID === image.ImageID ? 'Hide Notes' : 'Photo Notes'}
                      </button>
                      <button class="edit-button" type="button" on:click={() => startEditing(image)}>Edit Boxes</button>
                      <button
                              class="approve-button"
                              type="button"
                              on:click={() => approveReview(image)}
                              disabled={isReviewApproved(image) || isApprovingImageID === image.ImageID}
                      >
                        {isReviewApproved(image) ? 'Approved' : isApprovingImageID === image.ImageID ? 'Approving...' : 'Approve Review'}
                      </button>
                    </div>
                  </div>
                  {#if reviewErrorImageID === image.ImageID}
                    <p class="review-error">{reviewError}</p>
                  {/if}
                  {#if openJsonImageID === image.ImageID}
                    <pre class="json-panel">{jsonByImageID[image.ImageID]}</pre>
                  {/if}
                  {#if openAnnotationImageID === image.ImageID}
                    <div class="annotation-panel">
                      <label for={`annotation-${image.ImageID}`}>Photo notes</label>
                      <textarea
                              id={`annotation-${image.ImageID}`}
                              bind:value={annotationNotesByImageID[image.ImageID]}
                              maxlength="180"
                              placeholder="Optional note shown below this photo in the PDF"
                      />
                      <small>{(annotationNotesByImageID[image.ImageID] || '').length}/180</small>
                      {#if isReviewApproved(image)}
                        <p class="approval-reset-note">Saving a changed note returns this image to unreviewed.</p>
                      {/if}
                      {#if annotationError}
                        <p class="annotation-error">{annotationError}</p>
                      {/if}
                      <button
                              class="save-annotations"
                              type="button"
                              on:click={() => saveAnnotations(image)}
                              disabled={isSavingAnnotationImageID === image.ImageID}
                      >
                        {isSavingAnnotationImageID === image.ImageID ? 'Saving...' : 'Save Notes'}
                      </button>
                    </div>
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
      {:else if project && projectView === 'report'}
        <section class="report-workspace" aria-labelledby="report-heading">
          <div class="report-heading-row">
            <div>
              <p class="report-eyebrow">Insurance documentation</p>
              <h3 id="report-heading">Roof Inspection Report</h3>
              <p>Enter the report details, then generate a PDF from the final reviewed damage photos.</p>
            </div>
            <div class:ready={unreviewedPhotoCount === 0 && images.length > 0} class="review-status">
              <strong>{approvedReviewCount}/{images.length}</strong>
              <span>images reviewed</span>
              <small>{approvedDamagePhotoCount} damage / {approvedNoDamageCount} no damage</small>
            </div>
          </div>

          {#if reportMessage}
            <p class="report-message" role="status">{reportMessage}</p>
          {/if}
          {#if reportError}
            <p class="report-error" role="alert">{reportError}</p>
          {/if}

          {#if reportState === 'loading'}
            <div class="report-state-card">Loading report...</div>
          {:else if reportState === 'error'}
            <div class="report-state-card">
              <p>The report could not be loaded.</p>
              <button type="button" on:click={() => loadInspectionReport(project.ID, true)}>Try Again</button>
            </div>
          {:else if reportState === 'missing'}
            <div class="report-state-card">
              <h4>No report created yet</h4>
              <p>Create one report for this project, then add the inspection and claim details.</p>
              <button type="button" on:click={createReport}>Create Report</button>
            </div>
          {:else if reportState === 'ready' && reportDraft}
            <form class="report-form" on:submit|preventDefault={() => saveReport()}>
              <div class="report-field">
                <label for="report-number">Report number</label>
                <input id="report-number" maxlength="40" bind:value={reportDraft.ReportNumber} required />
              </div>
              <div class="report-field">
                <label for="customer-name">Customer / contact</label>
                <input id="customer-name" maxlength="60" bind:value={reportDraft.CustomerName} />
              </div>
              <div class="report-field report-field-wide">
                <label for="property-address">Property address</label>
                <input id="property-address" maxlength="80" bind:value={reportDraft.PropertyAddress} required />
              </div>
              <div class="report-address-fields report-field-wide">
                <div class="report-field">
                  <label for="property-city">City</label>
                  <input id="property-city" maxlength="60" bind:value={reportDraft.PropertyCity} />
                </div>
                <div class="report-field">
                  <label for="property-state">State</label>
                  <input
                          id="property-state"
                          maxlength="2"
                          bind:value={reportDraft.PropertyState}
                          on:input={() => reportDraft.PropertyState = reportDraft.PropertyState.toUpperCase()}
                  />
                </div>
                <div class="report-field">
                  <label for="property-zip">ZIP code</label>
                  <input id="property-zip" maxlength="10" inputmode="numeric" bind:value={reportDraft.PropertyZip} />
                </div>
              </div>
              <div class="report-field">
                <label for="inspector-name">Inspector</label>
                <input id="inspector-name" maxlength="60" bind:value={reportDraft.InspectorName} required />
              </div>
              <div class="report-field">
                <label for="inspection-date">Inspection date</label>
                <input id="inspection-date" type="date" bind:value={reportDraft.InspectionDate} required />
              </div>
              <div class="report-field">
                <label for="loss-date">Date of loss <span>Optional</span></label>
                <input id="loss-date" type="date" bind:value={reportDraft.DateOfLoss} />
              </div>
              <div class="report-field">
                <label for="insurance-carrier">Insurance carrier <span>Optional</span></label>
                <input id="insurance-carrier" maxlength="60" bind:value={reportDraft.InsuranceCarrier} />
              </div>
              <div class="report-field">
                <label for="claim-number">Claim number <span>Optional</span></label>
                <input id="claim-number" maxlength="40" bind:value={reportDraft.ClaimNumber} />
              </div>
              <div class="report-field report-field-wide">
                <label for="report-summary">Inspection summary</label>
                <textarea id="report-summary" maxlength="350" rows="4" bind:value={reportDraft.Summary}></textarea>
                <small>{reportDraft.Summary.length}/350</small>
              </div>
              {#if unreviewedPhotoCount > 0}
                <div class="report-review-warning report-field-wide">
                  <div>
                    <strong>{unreviewedPhotoCount} {unreviewedPhotoCount === 1 ? 'image is' : 'images are'} still unreviewed</strong>
                    <p>{approvedDamagePhotoCount > 0 ? 'You can generate now; unapproved images will be excluded from the PDF.' : 'Unapproved images will be excluded; approve any damage image before generating.'}</p>
                  </div>
                  <button type="button" on:click={() => showProjectView('photos')}>Review Images</button>
                </div>
              {/if}

              {#if approvedDamagePhotoCount === 0}
                <div class="report-readiness report-field-wide">
                  <div>
                    <strong>No approved damage photos yet</strong>
                    <p>Approve at least one image containing a final damage box before generating the report.</p>
                  </div>
                  <button type="button" on:click={() => showProjectView('photos')}>Go to Photo Review</button>
                </div>
              {/if}

              <div class="report-actions report-field-wide">
                <button
                        class="secondary-report-button"
                        type="submit"
                        disabled={isSavingReport || isGeneratingReport || !reportIsDirty}
                >
                  {isSavingReport ? 'Saving...' : 'Save Changes'}
                </button>
                {#if report.LastGeneratedPdfPath}
                  <button
                          class="secondary-report-button"
                          type="button"
                          on:click={openLastReport}
                          disabled={isSavingReport || isGeneratingReport}
                  >
                    Open Last PDF
                  </button>
                {/if}
                <button
                        class="primary-report-button"
                        type="button"
                        on:click={generateAndOpenReport}
                        disabled={isSavingReport || isGeneratingReport || approvedDamagePhotoCount === 0}
                >
                  {isGeneratingReport ? 'Generating...' : 'Generate & Open PDF'}
                </button>
              </div>
            </form>

            {#if report.LastGeneratedPdfPath}
              <div class="last-report">
                <div>
                  <span>Last generated PDF</span>
                  <strong>{report.LastGeneratedAt || 'Generated'}</strong>
                </div>
                <code>{report.LastGeneratedPdfPath}</code>
              </div>
            {/if}
          {/if}
        </section>
      {/if}
      {#if showScrollToTop}
        <button class="scroll-to-top" type="button" on:click={() => scrollMainToTop()}>
          Back to top
        </button>
      {/if}
  </main>
</div>

{#if selectedImage}
  <div class="image-modal" role="dialog" aria-modal="true" aria-label="Full image view">
    <div class="modal-actions">
      {#if selectedImage.editing}
        <span class="editor-hint">Drag empty space to add a box</span>
        <button class="editor-button" type="button" on:click={deleteSelectedDetection} disabled={selectedDetectionIndex === null}>
          Delete Selected
        </button>
        <button class="editor-button" type="button" on:click={saveEdits} disabled={isSavingAnnotationImageID === selectedImage.ImageID}>
          {isSavingAnnotationImageID === selectedImage.ImageID ? 'Saving...' : 'Save Changes'}
        </button>
        <button class="editor-button" type="button" on:click={cancelEdits}>Cancel Edit</button>
      {/if}
      <button class="modal-close" type="button" on:click={closeImage}>Close</button>
    </div>
    <div class="modal-content">
      {#if isLoadingSelectedImage}
        <div class="tile-empty">Loading original image...</div>
      {:else if selectedImageError}
        <div class="tile-empty">{selectedImageError}</div>
      {:else if selectedImageSource}
        <svg
                bind:this={fullImageSvg}
                class="full-image"
                class:editing-image={selectedImage.editing}
                viewBox={`0 0 ${selectedImage.ImageWidth.Int64} ${selectedImage.ImageHeight.Int64}`}
                preserveAspectRatio="xMidYMid meet"
                aria-label={selectedImage.ImagePath}
                on:mousedown={startDraw}
        >
          <image
                  href={selectedImageSource}
                  width={selectedImage.ImageWidth.Int64}
                  height={selectedImage.ImageHeight.Int64}
          />
          {#if selectedImage.showDetections}
            {#each detectionsFor(selectedImage) as detection, index}
              <rect
                      x={detection.BoundingBox.Left}
                      y={detection.BoundingBox.Top}
                      width={detection.BoundingBox.Right - detection.BoundingBox.Left}
                      height={detection.BoundingBox.Bottom - detection.BoundingBox.Top}
                      class:selected-detection={selectedImage.editing && selectedDetectionIndex === index}
                      class="detection-box"
                      on:mousedown|stopPropagation={(event) => selectedImage.editing && startMove(event, index)}
              />
              {#if selectedImage.editing && selectedDetectionIndex === index}
                <rect
                        x={detection.BoundingBox.Left - 8}
                        y={detection.BoundingBox.Top - 8}
                        width="16"
                        height="16"
                        class="resize-handle"
                        on:mousedown|stopPropagation={(event) => startResize(event, index, 'nw')}
                />
                <rect
                        x={detection.BoundingBox.Right - 8}
                        y={detection.BoundingBox.Top - 8}
                        width="16"
                        height="16"
                        class="resize-handle"
                        on:mousedown|stopPropagation={(event) => startResize(event, index, 'ne')}
                />
                <rect
                        x={detection.BoundingBox.Left - 8}
                        y={detection.BoundingBox.Bottom - 8}
                        width="16"
                        height="16"
                        class="resize-handle"
                        on:mousedown|stopPropagation={(event) => startResize(event, index, 'sw')}
                />
                <rect
                        x={detection.BoundingBox.Right - 8}
                        y={detection.BoundingBox.Bottom - 8}
                        width="16"
                        height="16"
                        class="resize-handle"
                        on:mousedown|stopPropagation={(event) => startResize(event, index, 'se')}
                />
              {/if}
            {/each}
          {/if}
        </svg>
      {/if}
    </div>
  </div>
{/if}

<style>

  .project-tabs {
    display: flex;
    gap: 6px;
    margin: 12px 32px 0;
    border-bottom: 1px solid #6b7280;
  }

  .project-tabs button {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border: 0;
    border-bottom: 3px solid transparent;
    padding: 10px 14px;
    background: transparent;
    color: #d1d5db;
    cursor: pointer;
    font-weight: 700;
  }

  .project-tabs button:hover,
  .project-tabs button.active {
    border-bottom-color: #dc2626;
    color: white;
  }

  .reviewed-count {
    min-width: 22px;
    border-radius: 999px;
    padding: 2px 7px;
    background: #374151;
    color: #f3f4f6;
    font-size: 11px;
    text-align: center;
  }

  .project-tabs button.active .reviewed-count {
    background: #dc2626;
  }

  .report-workspace {
    width: min(980px, calc(100% - 32px));
    margin: 20px auto 48px;
    color: #f8fafc;
  }

  .report-heading-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 28px;
    margin-bottom: 18px;
  }

  .report-eyebrow {
    margin: 0 0 5px;
    color: #fca5a5;
    font-size: 11px;
    font-weight: 800;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  .report-heading-row h3 {
    margin: 0;
    font-size: 25px;
    line-height: 1.2;
  }

  .report-heading-row p:not(.report-eyebrow) {
    max-width: 650px;
    margin: 7px 0 0;
    color: #d1d5db;
  }

  .review-status {
    display: flex;
    min-width: 126px;
    flex-direction: column;
    border: 1px solid #6b7280;
    border-radius: 6px;
    padding: 10px 14px;
    background: #374151;
    color: #d1d5db;
  }

  .review-status.ready {
    border-color: #16a34a;
    background: #14532d;
    color: #dcfce7;
  }

  .review-status strong {
    font-size: 22px;
    line-height: 1;
  }

  .review-status span {
    margin-top: 5px;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
  }

  .review-status small {
    margin-top: 5px;
    color: inherit;
    font-size: 11px;
  }

  .report-message,
  .report-error {
    margin: 0 0 14px;
    border-radius: 4px;
    padding: 10px 12px;
  }

  .report-message {
    border: 1px solid #16a34a;
    background: #14532d;
    color: #dcfce7;
  }

  .report-error {
    border: 1px solid #dc2626;
    background: #7f1d1d;
    color: #fee2e2;
  }

  .report-state-card {
    display: flex;
    min-height: 210px;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    border: 1px solid #6b7280;
    border-radius: 7px;
    padding: 30px;
    background: #1f2937;
    text-align: center;
  }

  .report-state-card h4,
  .report-state-card p {
    margin: 0;
  }

  .report-state-card p {
    color: #d1d5db;
  }

  .report-state-card button,
  .report-readiness button,
  .report-review-warning button {
    border: 0;
    border-radius: 4px;
    padding: 8px 12px;
    background: white;
    color: #111827;
    cursor: pointer;
    font-weight: 700;
  }

  .report-form {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px 18px;
    border: 1px solid #6b7280;
    border-radius: 7px;
    padding: 22px;
    background: #1f2937;
  }

  .report-field {
    position: relative;
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 6px;
  }

  .report-field-wide {
    grid-column: 1 / -1;
  }

  .report-address-fields {
    display: grid;
    grid-template-columns: minmax(0, 2fr) minmax(80px, 0.6fr) minmax(110px, 1fr);
    gap: 18px;
  }

  .report-field label {
    color: #f3f4f6;
    font-size: 12px;
    font-weight: 800;
  }

  .report-field label span {
    color: #9ca3af;
    font-weight: 500;
  }

  .report-field input,
  .report-field textarea {
    width: 100%;
    border: 1px solid #9ca3af;
    border-radius: 4px;
    padding: 9px 10px;
    outline: none;
    background: #f9fafb;
    color: #111827;
    font: inherit;
  }

  .report-field textarea {
    resize: vertical;
    line-height: 1.45;
  }

  .report-field input:focus,
  .report-field textarea:focus {
    border-color: #dc2626;
    box-shadow: 0 0 0 2px rgba(220, 38, 38, 0.2);
  }

  .report-field small {
    align-self: flex-end;
    color: #9ca3af;
    font-size: 10px;
  }

  .report-readiness,
  .report-review-warning {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    border: 1px solid #d97706;
    border-radius: 5px;
    padding: 12px 14px;
    background: #78350f;
    color: #fffbeb;
  }

  .report-readiness p {
    margin: 3px 0 0;
    color: #fde68a;
    font-size: 12px;
  }

  .report-review-warning p {
    margin: 3px 0 0;
    color: #fde68a;
    font-size: 12px;
  }

  .report-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    border-top: 1px solid #4b5563;
    padding-top: 18px;
  }

  .report-actions button {
    border: 0;
    border-radius: 4px;
    padding: 9px 14px;
    cursor: pointer;
    font-weight: 800;
  }

  .secondary-report-button {
    background: #f3f4f6;
    color: #111827;
  }

  .primary-report-button {
    background: #dc2626;
    color: white;
  }

  .report-actions button:disabled {
    cursor: default;
    opacity: 0.45;
  }

  .last-report {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: 18px;
    margin-top: 14px;
    border: 1px solid #6b7280;
    border-radius: 5px;
    padding: 12px 14px;
    background: #111827;
  }

  .last-report div {
    display: flex;
    flex-direction: column;
  }

  .last-report span {
    color: #9ca3af;
    font-size: 10px;
    font-weight: 800;
    text-transform: uppercase;
  }

  .last-report code {
    overflow-wrap: anywhere;
    color: #d1d5db;
    font-size: 11px;
  }

  .scroll-to-top {
    position: fixed;
    right: 24px;
    bottom: 24px;
    z-index: 5;
    border: 1px solid #fca5a5;
    border-radius: 999px;
    padding: 9px 14px;
    background: #991b1b;
    color: white;
    cursor: pointer;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
    font-weight: 800;
  }

  .scroll-to-top:hover {
    background: #dc2626;
  }

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

  .input-box .btn:disabled {
    cursor: default;
    opacity: 0.55;
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
    pointer-events: all;
  }

  .selected-detection {
    stroke: #facc15;
  }

  .resize-handle {
    fill: white;
    stroke: #dc2626;
    stroke-width: 2;
    vector-effect: non-scaling-stroke;
    cursor: nwse-resize;
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
    gap: 10px;
    min-height: 38px;
    padding: 0 10px;
  }

  .result-status {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .review-badge {
    border: 1px solid #d97706;
    border-radius: 999px;
    padding: 3px 8px;
    background: #78350f;
    color: #fde68a;
    font-size: 10px;
    font-weight: 800;
    text-transform: uppercase;
  }

  .review-badge.damage {
    border-color: #ef4444;
    background: #7f1d1d;
    color: #fee2e2;
  }

  .review-badge.no-damage {
    border-color: #22c55e;
    background: #14532d;
    color: #dcfce7;
  }

  .toolbar-actions {
    display: flex;
    gap: 8px;
  }

  .json-toggle,
  .annotation-toggle,
  .edit-button,
  .approve-button {
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

  .annotation-toggle {
    background: #334155;
    color: #d1d5db;
  }

  .edit-button {
    background: #dc2626;
    color: white;
  }

  .approve-button {
    background: #16a34a;
    color: white;
  }

  .approve-button:disabled {
    cursor: default;
    opacity: 0.55;
  }

  .review-error {
    margin: 0;
    border-top: 1px solid #334155;
    padding: 7px 10px;
    color: #fecaca;
    font-size: 12px;
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

  .annotation-panel {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px;
    border-top: 1px solid #334155;
    color: #d1d5db;
    font-size: 12px;
  }

  .annotation-panel textarea {
    min-height: 90px;
    resize: vertical;
    border: 1px solid #475569;
    border-radius: 3px;
    padding: 8px;
    background: #0f172a;
    color: white;
    font: inherit;
  }

  .annotation-panel small {
    align-self: flex-end;
    color: #9ca3af;
  }

  .approval-reset-note {
    margin: 0;
    color: #fde68a;
  }

  .save-annotations {
    align-self: flex-end;
    border: 0;
    border-radius: 3px;
    padding: 6px 10px;
    background: #dc2626;
    color: white;
    cursor: pointer;
    font-size: 12px;
  }

  .save-annotations:disabled {
    cursor: default;
    opacity: 0.5;
  }

  .annotation-error {
    margin: 0;
    color: #fecaca;
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

  .modal-actions {
    display: flex;
    width: min(1100px, 100%);
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    margin-bottom: 12px;
  }

  .editor-hint {
    margin-right: auto;
    color: #d1d5db;
    font-size: 12px;
  }

  .modal-close,
  .editor-button {
    border: 0;
    border-radius: 3px;
    padding: 8px 12px;
    color: white;
    cursor: pointer;
  }

  .modal-close {
    background: #dc2626;
  }

  .editor-button {
    background: #334155;
  }

  .editor-button:disabled {
    cursor: default;
    opacity: 0.5;
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

  .editing-image {
    cursor: crosshair;
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

  .project-list-item {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: stretch;
    overflow: hidden;
    border: 1px solid #334155;
    border-radius: 8px;
    background: #1f2937;
  }

  .project-list-item:hover,
  .project-list-item.selected {
    border-color: #64748b;
    background: #273449;
  }

  .project-select {
    min-width: 0;
    padding: 12px;
    background: transparent;
    text-align: left;
  }

  .project-delete {
    align-self: stretch;
    border-left: 1px solid #334155;
    padding: 0 10px;
    background: transparent;
    color: #fca5a5;
    font-size: 0.72rem;
    font-weight: 700;
  }

  .project-delete:hover:not(:disabled) {
    background: #7f1d1d;
    color: white;
  }

  .project-delete:disabled {
    cursor: wait;
    opacity: 0.55;
  }

  .project-list-error {
    margin-top: 12px;
    color: #fecaca;
    font-size: 0.8rem;
    line-height: 1.4;
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

  .project-error {
    margin: 0;
    color: #fecaca;
    text-align: center;
  }

  @media (max-width: 760px) {
    .app-shell {
      grid-template-columns: 1fr;
      grid-template-rows: auto minmax(0, 1fr);
    }

    aside {
      max-height: 220px;
    }

    .report-heading-row,
    .report-readiness,
    .report-review-warning {
      flex-direction: column;
      align-items: stretch;
    }

    .result-toolbar,
    .toolbar-actions {
      flex-wrap: wrap;
    }

    .report-form {
      grid-template-columns: 1fr;
    }

    .report-address-fields {
      grid-template-columns: 1fr;
    }

    .report-field-wide {
      grid-column: auto;
    }

    .report-actions {
      align-items: stretch;
      flex-direction: column;
    }

    .last-report {
      grid-template-columns: 1fr;
    }
  }
</style>
