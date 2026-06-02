/**
 * AutoNotes Search Page Logic
 */

let searchIndex = [];
let fuseInstance = null;
let currentFilter = "all";

// Inline Image Crops Sizing & Extraction
function initializeImageCrops(container) {
  const MIN_RESCALE_WIDTH = 0.25;
  const MAX_RESCALE_WIDTH = 1.5;

  container.querySelectorAll(".inline-image-crop").forEach((cropContainer) => {
    const src = cropContainer.getAttribute("data-src");
    const right = parseFloat(cropContainer.getAttribute("data-right"));
    const left = parseFloat(cropContainer.getAttribute("data-left"));
    const bottom = parseFloat(cropContainer.getAttribute("data-bottom"));
    const top = parseFloat(cropContainer.getAttribute("data-top"));

    const rightVal = isNaN(right) ? 1000 : right;
    const leftVal = isNaN(left) ? 0 : left;
    const topVal = isNaN(top) ? 0 : top;
    const bottomVal = isNaN(bottom) ? 1000 : bottom;

    const img = new Image();
    img.src = src;
    img.onload = () => {
      const natW = img.naturalWidth;
      const natH = img.naturalHeight;

      const sx = (leftVal / 1000) * natW;
      const sy = (topVal / 1000) * natH;
      const sWidth = Math.max(1, ((rightVal - leftVal) / 1000) * natW);
      const sHeight = Math.max(1, ((bottomVal - topVal) / 1000) * natH);

      const canvas = document.createElement("canvas");
      canvas.width = sWidth;
      canvas.height = sHeight;

      const ctx = canvas.getContext("2d");
      ctx.drawImage(img, sx, sy, sWidth, sHeight, 0, 0, sWidth, sHeight);

      const croppedImg = new Image();
      croppedImg.src = canvas.toDataURL();
      croppedImg.style.width = "100%";
      croppedImg.style.display = "block";

      const cropRatio = Math.max(0.001, (rightVal - leftVal) / 1000);
      const mappedRatio =
        MIN_RESCALE_WIDTH + (MAX_RESCALE_WIDTH - MIN_RESCALE_WIDTH) * cropRatio;
      const finalWidthPercent = Math.min(100, Math.max(0, mappedRatio * 100));
      cropContainer.style.setProperty("--crop-ratio", cropRatio);
      cropContainer.style.width = finalWidthPercent + "%";
      cropContainer.innerHTML = "";
      cropContainer.appendChild(croppedImg);
    };
  });
}

// Render all results at load
function renderAllResults(entries) {
  const container = document.getElementById("search-results");
  if (!container) return;

  container.innerHTML = "";

  entries.forEach((entry) => {
    const card = document.createElement("a");
    card.className = "search-card";
    card.href = entry.lessonLink;
    card.target = "_blank";
    card.rel = "noopener noreferrer";
    card.dataset.type = entry.type;
    card.dataset.id = entry.id;

    card.innerHTML = `
      <div class="card-content">
        <div class="card-header">
          <span class="card-type-badge ${entry.type}">${entry.type}</span>
          <span class="card-lesson-link" title="Open lesson ${entry.lessonTitle}">
            <i data-lucide="external-link" size="11"></i>
          </span>
        </div>
        <div class="card-body">
          ${entry.contentHtml}
        </div>
        <div class="card-footer">
          <span class="card-course" title="${entry.lessonTitle}">${entry.course}</span>
          <span class="card-date">${entry.date}</span>
        </div>
      </div>
    `;
    container.appendChild(card);
  });

  // Initialize Icons & Image Crops
  if (window.lucide) {
    window.lucide.createIcons();
  }
  initializeImageCrops(container);

  // Typeset MathJax
  if (window.MathJax && window.MathJax.typesetPromise) {
    const mathElements = container.querySelectorAll(
      "span.math:not([data-rendered])",
    );
    mathElements.forEach((el) => {
      let content = el.textContent;
      // Strip out "larger than normal" size directives
      content = content.replace(/\\(large|Large|LARGE|huge|Huge)\b\s*/g, "");
      // Render all math as inline
      el.textContent = `\\(${content}\\)`;
      el.dataset.rendered = "true";
    });

    window.MathJax.typesetPromise([container]);
  }
}

// Perform Fuse.js query and toggle display
function executeSearch() {
  const query = document.getElementById("search-input").value.trim();
  let results = searchIndex;

  if (query !== "") {
    const fuseResults = fuseInstance.search(query);
    results = fuseResults.map((r) => r.item);
  }

  // Filter by active pill
  if (currentFilter !== "all") {
    results = results.filter((item) => item.type === currentFilter);
  }

  const totalMatches = results.length;
  // Only slice to top 20 when there is an active search query
  const slicedResults = query !== "" ? results.slice(0, 20) : results;

  // Build a Map of visible IDs to their index for sorting
  const visibleOrderMap = new Map();
  slicedResults.forEach((item, index) => {
    visibleOrderMap.set(item.id, index);
  });

  // Toggle card visibility and set order
  const container = document.getElementById("search-results");
  if (container) {
    container.querySelectorAll(".search-card").forEach((card) => {
      const id = card.dataset.id;
      if (visibleOrderMap.has(id)) {
        card.style.display = ""; // default layout (grid)
        card.style.order = visibleOrderMap.get(id); // Sort based on fuse relevance score
      } else {
        card.style.display = "none"; // hide
        card.style.order = ""; // reset
      }
    });
  }

  // Update stats
  const stats = document.getElementById("search-stats");
  if (stats) {
    if (totalMatches === 0) {
      stats.textContent = "No matching items found.";
    } else {
      if (query !== "" && totalMatches > 20) {
        stats.textContent = `Showing top 20 of ${totalMatches} items.`;
      } else {
        stats.textContent = `Found ${totalMatches} item${totalMatches === 1 ? "" : "s"}.`;
      }
    }
  }
}

// Main Page Initialization
document.addEventListener("DOMContentLoaded", () => {
  const searchInput = document.getElementById("search-input");
  const stats = document.getElementById("search-stats");

  // Fetch search data
  fetch("data/search.json")
    .then((res) => {
      if (!res.ok) throw new Error("Search index not found");
      return res.json();
    })
    .then((data) => {
      searchIndex = data;

      // Initialize Fuse
      fuseInstance = new Fuse(searchIndex, {
        keys: [
          { name: "lessonTitle", weight: 0.4 },
          { name: "type", weight: 0.5 },
          { name: "course", weight: 0.3 },
          { name: "contentText", weight: 1.0 },
        ],
        threshold: 0.4,
        ignoreLocation: true,
      });

      // Render all cards once
      renderAllResults(searchIndex);

      // Apply initial search filter
      executeSearch();
    })
    .catch((err) => {
      console.error(err);
      stats.textContent =
        "Error loading search index. Please build the project first.";
    });

  // Search Input Listener
  if (searchInput) {
    searchInput.addEventListener("input", executeSearch);
  }

  // Filter Pill Listeners
  const filterContainer = document.getElementById("search-filters");
  if (filterContainer) {
    filterContainer.addEventListener("click", (e) => {
      const pill = e.target.closest(".filter-pill");
      if (!pill) return;

      filterContainer.querySelectorAll(".filter-pill").forEach((btn) => {
        btn.classList.remove("active");
      });
      pill.classList.add("active");

      currentFilter = pill.getAttribute("data-type");
      executeSearch();
    });
  }

  // Theme Management
  function applyTheme(theme) {
    const root = document.documentElement;
    const systemDark = window.matchMedia(
      "(prefers-color-scheme: dark)",
    ).matches;

    let actualTheme = theme;
    if (theme === "system") {
      actualTheme = systemDark ? "dark" : "light";
    }

    if (actualTheme === "dark") {
      root.setAttribute("data-theme", "dark");
    } else {
      root.removeAttribute("data-theme");
    }

    const themeToggle = document.getElementById("theme-toggle");
    if (themeToggle) {
      let iconName = "sun-moon";
      if (theme === "light") iconName = "sun";
      else if (theme === "dark") iconName = "moon";

      themeToggle.innerHTML = `<i data-lucide="${iconName}" size="18"></i>`;
      if (window.lucide) window.lucide.createIcons();
    }

    localStorage.setItem("theme-pref", theme);
  }

  const savedTheme = localStorage.getItem("theme-pref") || "system";
  applyTheme(savedTheme);

  const themeToggle = document.getElementById("theme-toggle");
  if (themeToggle) {
    themeToggle.addEventListener("click", () => {
      const current = localStorage.getItem("theme-pref") || "system";
      let next;
      if (current === "system") next = "light";
      else if (current === "light") next = "dark";
      else next = "system";
      applyTheme(next);
    });
  }

  // Listen for system theme changes if set to system
  window
    .matchMedia("(prefers-color-scheme: dark)")
    .addEventListener("change", () => {
      if (localStorage.getItem("theme-pref") === "system") {
        applyTheme("system");
      }
    });

  // Initialize Lucide Icons
  if (window.lucide) {
    window.lucide.createIcons();
  }
});
