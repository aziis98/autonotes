/**
 * AutoNotes - Interaction Logic (Vanilla JS Module)
 * Neo-Brutalist Edition - No Animations, Direct Snapping.
 */

const imgEl = document.getElementById("current-image");
const highlight = document.getElementById("highlight");
const lensContainer = document.getElementById("lens-container");
const lensScaler = document.getElementById("lens-scaler");
const lensImg = document.getElementById("lens-image");

/**
 * Update the high-resolution crop preview (The Lens)
 */
export function updateLens(imgSrc, top, right, bottom, left) {
  if (!lensContainer || !imgSrc || isNaN(top)) {
    if (lensContainer) lensContainer.classList.add("hidden");
    return;
  }

  // Horizontal padding (10%) to provide more context
  const padY = (bottom - top) * 0.1;
  const pTop = Math.max(0, top - padY);
  const pBottom = Math.min(1000, bottom + padY);

  lensImg.src = imgSrc;

  const ratioW = right - left;
  const ratioH = pBottom - pTop;

  if (ratioW <= 0 || ratioH <= 0) return;

  // Correctly incorporate the image's physical aspect ratio
  const natW = imgEl ? imgEl.naturalWidth : 1000;
  const natH = imgEl ? imgEl.naturalHeight : 1000;
  const aspectRatio = (ratioW / ratioH) * (natW / natH);

  lensContainer.classList.remove("hidden");
  lensContainer.style.display = "flex"; // Ensure it's rendered to get width

  const columnEl = document.getElementById("left-column");
  if (!columnEl) return;

  const maxWidth = Math.min(1200, columnEl.clientWidth - 80); // Capped at 1200px, minus padding
  const maxHeight = window.innerHeight * 0.45;

  let finalW = maxWidth;
  let finalH = finalW / aspectRatio;

  if (finalH > maxHeight) {
    finalH = maxHeight;
    finalW = finalH * aspectRatio;
  }

  lensScaler.style.width = Math.floor(finalW) + "px";
  lensScaler.style.height = Math.floor(finalH) + "px";

  const mappedImgW = finalW * (1000 / ratioW);
  const mappedImgH = finalH * (1000 / ratioH);

  lensImg.style.width = mappedImgW + "px";
  lensImg.style.height = mappedImgH + "px";

  const mappedLeft = -(left / 1000) * mappedImgW;
  const mappedTop = -(pTop / 1000) * mappedImgH;

  lensImg.style.left = mappedLeft + "px";
  lensImg.style.top = mappedTop + "px";
}

/**
 * Update the red/bold highlight box in the main image viewer
 */
export function updateHighlight(top, right, bottom, left) {
  if (!highlight || !imgEl) return;

  const rectW = imgEl.clientWidth;
  const rectH = imgEl.clientHeight;

  const hlTop = (top / 1000) * rectH;
  const hlLeft = (left / 1000) * rectW;
  const hlWidth = ((right - left) / 1000) * rectW;
  const hlHeight = ((bottom - top) / 1000) * rectH;

  highlight.style.top = hlTop + "px";
  highlight.style.left = hlLeft + "px";
  highlight.style.width = hlWidth + "px";
  highlight.style.height = hlHeight + "px";
  highlight.style.display = "block";
}

export function setHighlight(imgSrc, top, right, bottom, left) {
  if (!imgEl) return;

  if (imgSrc && imgSrc !== "" && imgEl.getAttribute("src") !== imgSrc) {
    imgEl.src = imgSrc;
    imgEl.style.display = "block";
    if (highlight) highlight.style.display = "none";

    imgEl.onload = () => {
      if (!isNaN(top)) {
        updateHighlight(top, right, bottom, left);
        updateLens(imgSrc, top, right, bottom, left);
      }
    };
  } else if (imgSrc && !isNaN(top)) {
    updateHighlight(top, right, bottom, left);
    updateLens(imgSrc, top, right, bottom, left);
  } else {
    clearHighlight();
  }
}

export function clearHighlight() {
  if (highlight) highlight.style.display = "none";
  if (lensContainer) {
    lensContainer.classList.add("hidden");
    lensContainer.style.display = "none";
  }
}

// Initial Bootstrap
document.addEventListener("DOMContentLoaded", () => {
  const body = document.body;
  const toggle = document.getElementById("toggle-preview");

  // Preview persistence
  if (localStorage.getItem("preview-hidden") === "true") {
    body.classList.add("preview-hidden");
  }

  if (toggle) {
    toggle.addEventListener("click", () => {
      body.classList.toggle("preview-hidden");
      localStorage.setItem(
        "preview-hidden",
        body.classList.contains("preview-hidden"),
      );
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

    // Update toggle button icon
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

  // Transcription Toggle
  const toggleTranscription = document.getElementById("toggle-transcription");
  if (toggleTranscription) {
    function updateTranscriptionIcon(hidden) {
      const iconName = hidden ? "eye-off" : "eye";
      toggleTranscription.innerHTML = `<i data-lucide="${iconName}" size="18"></i>`;
      if (window.lucide) window.lucide.createIcons();
    }

    const isHidden = localStorage.getItem("transcription-hidden") === "true";
    if (isHidden) {
      body.classList.add("transcription-hidden");
    }
    updateTranscriptionIcon(isHidden);

    toggleTranscription.addEventListener("click", () => {
      const nowHidden = body.classList.toggle("transcription-hidden");
      localStorage.setItem("transcription-hidden", nowHidden);
      updateTranscriptionIcon(nowHidden);
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

  // Delegation for Hover Events
  body.addEventListener("mouseover", (e) => {
    const box = e.target.closest(".box-text, .inline-image-crop");
    if (box) {
      const img = box.getAttribute("data-img");
      const t = parseFloat(box.getAttribute("data-top"));
      const r = parseFloat(box.getAttribute("data-right"));
      const b = parseFloat(box.getAttribute("data-bottom"));
      const l = parseFloat(box.getAttribute("data-left"));
      setHighlight(img, t, r, b, l);
    }
  });

  body.addEventListener("mouseout", (e) => {
    const box = e.target.closest(".box-text, .inline-image-crop");
    if (box) {
      const related = e.relatedTarget
        ? e.relatedTarget.closest(".box-text, .inline-image-crop")
        : null;
      if (related !== box) {
        clearHighlight();
      }
    }
  });

  // Reword Reference Hovering
  body.addEventListener("mouseover", (e) => {
    const reword = e.target.closest(".reword");
    if (reword) {
      const refs = reword.getAttribute("data-ref");
      if (window.DEBUG) console.log(`DEBUG: Reword hover, data-ref='${refs}'`);
      if (refs) {
        const ids = refs.split(" ");
        if (window.DEBUG) console.log(`DEBUG:   IDs to highlight:`, ids);
        let minT = 1000,
          maxR = 0,
          maxB = 0,
          minL = 1000;
        let commonImg = null;
        let foundAny = false;

        ids.forEach((id) => {
          const target = document.getElementById(id);
          if (target) {
            target.classList.add("box-highlight-green");

            const img =
              target.getAttribute("data-img") || target.getAttribute("src");
            const t = parseFloat(target.getAttribute("data-top"));
            const r = parseFloat(target.getAttribute("data-right"));
            const b = parseFloat(target.getAttribute("data-bottom"));
            const l = parseFloat(target.getAttribute("data-left"));

            if (img && !isNaN(t)) {
              if (!commonImg) commonImg = img;
              // Only gather boxes from the same image
              if (img === commonImg) {
                minT = Math.min(minT, t);
                maxR = Math.max(maxR, r);
                maxB = Math.max(maxB, b);
                minL = Math.min(minL, l);
                foundAny = true;
              }
            }
          }
        });

        if (foundAny) {
          setHighlight(commonImg, minT, maxR, maxB, minL);
        }
      }
    }
  });

  body.addEventListener("mouseout", (e) => {
    const reword = e.target.closest(".reword");
    if (reword) {
      const related = e.relatedTarget
        ? e.relatedTarget.closest(".reword")
        : null;
      if (related !== reword) {
        const refs = reword.getAttribute("data-ref");
        if (refs) {
          refs.split(" ").forEach((id) => {
            const target = document.getElementById(id);
            if (target) {
              target.classList.remove("box-highlight-green");
            }
          });
        }
        clearHighlight();
      }
    }
  });

  // MathJax Autorender
  function renderMath() {
    if (window.MathJax && window.MathJax.typesetPromise) {
      const mathElements = document.querySelectorAll(
        "span.math:not([data-rendered])",
      );
      if (mathElements.length === 0) return;

      mathElements.forEach((el) => {
        const display = el.getAttribute("data-display") === "true";
        const content = el.textContent;
        el.textContent = display ? `\\[${content}\\]` : `\\(${content}\\)`;
        el.dataset.rendered = "true";
      });

      window.MathJax.typesetPromise();
    } else {
      // If MathJax is not yet available, try again shortly
      setTimeout(renderMath, 100);
    }
  }

  // Call renderMath on load
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", renderMath);
  } else {
    renderMath();
  }

  // Inline Image Crops
  document.querySelectorAll(".inline-image-crop").forEach((container) => {
    const src = container.getAttribute("data-src");
    const right = parseFloat(container.getAttribute("data-right"));
    const left = parseFloat(container.getAttribute("data-left"));
    const bottom = parseFloat(container.getAttribute("data-bottom"));
    const top = parseFloat(container.getAttribute("data-top"));

    const img = new Image();
    img.src = src;
    img.onload = () => {
      const natW = img.naturalWidth;
      const natH = img.naturalHeight;
      const cropW = ((right - left) / 1000) * natW;
      const cropH = ((bottom - top) / 1000) * natH;
      const aspectRatio = cropW / cropH;

      const finalW = container.clientWidth;
      const finalH = finalW / aspectRatio;

      container.style.height = finalH + "px";

      const innerImg = document.createElement("img");
      innerImg.src = src;
      innerImg.style.position = "absolute";
      innerImg.style.width = (1000 / (right - left)) * 100 + "%";
      innerImg.style.left = -((left / (right - left)) * 100) + "%";
      innerImg.style.top = -((top / (bottom - top)) * 100) + "%";
      innerImg.style.height = (1000 / (bottom - top)) * 100 + "%";
      innerImg.style.maxWidth = "none";

      container.appendChild(innerImg);
    };
  });

  // Set Default Image for Right Column
  const firstBox = document.querySelector(".box-text");
  if (firstBox) {
    const firstImg = firstBox.getAttribute("data-img");
    if (firstImg) {
      if (imgEl) {
        imgEl.src = firstImg;
        imgEl.style.display = "block";
      }
    }
  }

  // Scroll Tracking: Auto-select nearest image based on viewport position
  let lastScrollUpdate = 0;
  window.addEventListener("scroll", () => {
    if (imgEl) {
      if (body.classList.contains("inspect-mode")) return; // Disable scroll sync in inspect mode

      clearHighlight();
      const now = Date.now();
      if (now - lastScrollUpdate < 150) return; // Throttle to 150ms
      lastScrollUpdate = now;

      const elements = document.querySelectorAll(".box-text, .reword");
      let closest = null;
      let minDistance = Infinity;

      // We target 150px from the top as the "focus line"
      const focusY = 150;

      elements.forEach((el) => {
        const rect = el.getBoundingClientRect();
        const distance = Math.abs(rect.top - focusY);

        if (distance < minDistance) {
          minDistance = distance;
          closest = el;
        }
      });

      if (closest) {
        let imgSrc = closest.getAttribute("data-img");

        // If it's a reword block, try to find the image from its references
        if (!imgSrc && closest.classList.contains("reword")) {
          const refs = closest.getAttribute("data-ref");
          if (refs) {
            const firstId = refs.split(" ")[0];
            const target = document.getElementById(firstId);
            if (target) {
              imgSrc =
                target.getAttribute("data-img") || target.getAttribute("src");
            }
          }
        }

        // Compare normalized href if src is a full URL
        if (imgSrc && !imgEl.src.endsWith(imgSrc)) {
          clearHighlight();
          imgEl.src = imgSrc;
        }
      }
    }
  });

  // Inspector Mode
  const toggleInspect = document.getElementById("toggle-inspect");
  const inspectorOverlay = document.getElementById("inspector-overlay");
  const inspectorSelection = document.getElementById("inspector-selection");

  if (toggleInspect && inspectorOverlay && inspectorSelection) {
    let isDragging = false;
    let startX = 0;
    let startY = 0;

    toggleInspect.addEventListener("click", () => {
      const active = body.classList.toggle("inspect-mode");
      toggleInspect.classList.toggle("active", active);
      inspectorOverlay.classList.toggle("hidden", !active);

      if (!active) {
        inspectorSelection.style.display = "none";
      }
    });

    inspectorOverlay.addEventListener("mousedown", (e) => {
      isDragging = true;
      const rect = inspectorOverlay.getBoundingClientRect();
      startX = e.clientX - rect.left;
      startY = e.clientY - rect.top;

      inspectorSelection.style.left = startX + "px";
      inspectorSelection.style.top = startY + "px";
      inspectorSelection.style.width = "0px";
      inspectorSelection.style.height = "0px";
      inspectorSelection.style.display = "block";
    });

    window.addEventListener("mousemove", (e) => {
      if (!isDragging) return;

      const rect = inspectorOverlay.getBoundingClientRect();
      const currentX = Math.max(0, Math.min(rect.width, e.clientX - rect.left));
      const currentY = Math.max(0, Math.min(rect.height, e.clientY - rect.top));

      const x = Math.min(startX, currentX);
      const y = Math.min(startY, currentY);
      const w = Math.abs(currentX - startX);
      const h = Math.abs(currentY - startY);

      inspectorSelection.style.left = x + "px";
      inspectorSelection.style.top = y + "px";
      inspectorSelection.style.width = w + "px";
      inspectorSelection.style.height = h + "px";
    });

    window.addEventListener("mouseup", () => {
      if (!isDragging) return;
      isDragging = false;

      const rect = inspectorOverlay.getBoundingClientRect();
      const selRect = inspectorSelection.getBoundingClientRect();

      const top = Math.round(((selRect.top - rect.top) / rect.height) * 1000);
      const left = Math.round(((selRect.left - rect.left) / rect.width) * 1000);
      const bottom = Math.round(
        ((selRect.bottom - rect.top) / rect.height) * 1000,
      );
      const right = Math.round(
        ((selRect.right - rect.left) / rect.width) * 1000,
      );

      const imgSrc = imgEl.src.split("/").pop();
      const tag = `<image src="${imgSrc}" top=${top} right=${right} bottom=${bottom} left=${left} />`;

      navigator.clipboard.writeText(tag).then(() => {
        // Feedback
        const originalHTML = toggleInspect.innerHTML;
        toggleInspect.innerHTML = `<i data-lucide="check" size="18"></i>`;
        if (window.lucide) window.lucide.createIcons();

        setTimeout(() => {
          toggleInspect.innerHTML = originalHTML;
          if (window.lucide) window.lucide.createIcons();

          // Auto-disable inspect mode after a short delay
          body.classList.remove("inspect-mode");
          toggleInspect.classList.remove("active");
          inspectorOverlay.classList.add("hidden");
          inspectorSelection.style.display = "none";
        }, 500);
      });
    });
  }

  // File Counters Logic
  document.querySelectorAll(".entry-counter").forEach((counter) => {
    const path = counter.getAttribute("data-path");
    const fullKey = `counter:${window.location.pathname}:${path}`;
    const valueEl = counter.querySelector(".counter-value");
    const incBtn = counter.querySelector(".inc");
    const decBtn = counter.querySelector(".dec");

    // Load from localStorage
    const saved = localStorage.getItem(fullKey) || "0";
    valueEl.textContent = saved;

    const update = (delta) => {
      const current = parseInt(valueEl.textContent, 10);
      const next = current + delta;
      valueEl.textContent = next;
      localStorage.setItem(fullKey, next);
    };

    incBtn.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      update(1);
    });

    decBtn.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      update(-1);
    });
  });
});
