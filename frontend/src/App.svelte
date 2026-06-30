<script>
  import {CreateProject} from '../wailsjs/go/main/App.js'
  import logo1 from './assets/images/Hail_Scan_Logo_Dark_Mode-removebg-preview.png'
  let resultText = "Please enter a directory below 👇"
  let directory = ''

  function createProject() {
    CreateProject(directory).then(result => {resultText = JSON.stringify(result, null, 2)})
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
  <main>
    <div class="input-box" id="input">
      <input autocomplete="off" bind:value={directory} class="input text-black mt-[20px]" id="name" type="text"/>
      <button class="btn mt-[20px]" on:click={createProject}>Create Project</button>
    </div>
    <div class="result" id="result">{resultText}</div>
  </main>
</div>

<style>

  .result {
    height: 20px;
    line-height: 20px;
    margin: 1.5rem auto;
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

</style>
