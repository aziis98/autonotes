/**
 * AutoNotes - Interaction Logic (Vanilla JS Module)
 * Neo-Brutalist Edition - No Animations, Direct Snapping.
 */

// Initialize DEBUG from script tag
const debugEl = document.getElementById("debug-data");
window.DEBUG = debugEl ? JSON.parse(debugEl.textContent) : false;

const imgEl = document.getElementById("current-image");
const highlight = document.getElementById("highlight");
const lensContainer = document.getElementById("lens-container");
const lensScaler = document.getElementById("lens-scaler");
const lensImg = document.getElementById("lens-image");
const allBoxesOverlay = document.getElementById("all-boxes-overlay");

/**
 * Update the high-resolution crop preview (The Lens)
 */
export function updateLens(imgSrc, top, right, bottom, left) {
  if (document.body.classList.contains("lens-hidden")) {
    if (lensContainer) {
      lensContainer.classList.add("hidden");
      lensContainer.style.display = "none";
    }
    return;
  }

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

function getBoxTarget(id) {
  if (!id) return null;
  const query = `.box-text[id="${id}"], .inline-image-crop[id="${id}"]`;
  const boxTarget = document.querySelector(query);
  if (boxTarget) return boxTarget;
  return document.getElementById(id);
}

// Initial Bootstrap
document.addEventListener("DOMContentLoaded", () => {
  const body = document.body;
  const toggle = document.getElementById("toggle-preview");

  // Preview persistence (defaults to hidden)
  if (localStorage.getItem("preview-hidden") !== "false") {
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

    const isHidden = localStorage.getItem("transcription-hidden") !== "false";
    if (isHidden) {
      body.classList.add("transcription-hidden");
    }
    toggleTranscription.classList.toggle("active", !isHidden);
    updateTranscriptionIcon(isHidden);

    toggleTranscription.addEventListener("click", () => {
      const nowHidden = body.classList.toggle("transcription-hidden");
      localStorage.setItem("transcription-hidden", nowHidden);
      toggleTranscription.classList.toggle("active", !nowHidden);
      updateTranscriptionIcon(nowHidden);
    });
  }

  // Lens Toggle
  const toggleLens = document.getElementById("toggle-lens");
  if (toggleLens) {
    function updateLensIcon(hidden) {
      const iconName = hidden ? "panel-bottom-open" : "panel-bottom-close";
      toggleLens.innerHTML = `<i data-lucide="${iconName}" size="18"></i>`;
      if (window.lucide) window.lucide.createIcons();
    }

    const isHidden = localStorage.getItem("lens-hidden") === "true";
    if (isHidden) {
      body.classList.add("lens-hidden");
    }
    toggleLens.classList.toggle("active", !isHidden);
    updateLensIcon(isHidden);

    toggleLens.addEventListener("click", () => {
      const nowHidden = body.classList.toggle("lens-hidden");
      localStorage.setItem("lens-hidden", nowHidden);
      toggleLens.classList.toggle("active", !nowHidden);
      updateLensIcon(nowHidden);
      if (nowHidden && lensContainer) {
        lensContainer.classList.add("hidden");
        lensContainer.style.display = "none";
      }
    });
  }

  // Reword Focus Toggle
  const toggleRewordsFocus = document.getElementById("toggle-rewords-focus");
  if (toggleRewordsFocus) {
    const isActive = localStorage.getItem("rewords-focus") === "true";
    if (isActive) {
      body.classList.add("rewords-focus");
      setTimeout(focusElementAtCenter, 200);
    }
    toggleRewordsFocus.classList.toggle("active", isActive);

    toggleRewordsFocus.addEventListener("click", () => {
      const nowActive = body.classList.toggle("rewords-focus");
      localStorage.setItem("rewords-focus", nowActive);
      toggleRewordsFocus.classList.toggle("active", nowActive);
      if (nowActive) {
        focusElementAtCenter();
      } else {
        clearRewordsFocus();
      }
    });
  }

  function focusElementAtCenter() {
    const readingElements = Array.from(
      document.querySelectorAll(
        "#left-column .reword-paragraph, #left-column li",
      ),
    );
    if (readingElements.length === 0) return;

    const centerY = window.innerHeight / 2;
    let closestElement = null;
    let minDistance = Infinity;

    readingElements.forEach((el) => {
      const rect = el.getBoundingClientRect();
      const elementCenterY = rect.top + rect.height / 2;
      const distance = Math.abs(elementCenterY - centerY);
      if (distance < minDistance) {
        minDistance = distance;
        closestElement = el;
      }
    });

    if (closestElement) {
      applyFocusFading(closestElement);
      updateIndicatorPosition();
    }
  }

  function updateIndicatorPosition() {
    const focused = document.querySelector(".focused-reading-element");
    const indicator = document.getElementById("reword-focus-indicator");
    if (!focused || !indicator) {
      if (indicator) indicator.classList.remove("visible");
      return;
    }
    const leftColumn = document.getElementById("left-column");
    const targetRect = focused.getBoundingClientRect();
    const columnRect = leftColumn.getBoundingClientRect();

    const top = targetRect.top - columnRect.top + leftColumn.scrollTop;
    // Center vertically with the line of text/box (indicator height is 32px)
    const heightOffset = (targetRect.height - 32) / 2;

    // Align horizontally with the main content wrapper's left edge
    const wrapper = document.querySelector(".content-wrapper");
    const wrapperRect = wrapper ? wrapper.getBoundingClientRect() : targetRect;
    const left = wrapperRect.left - columnRect.left - 45; // 45px to the left of content wrapper

    indicator.style.top = top + heightOffset + "px";
    indicator.style.left = left + "px";
    indicator.classList.add("visible");
  }

  function clearRewordsFocus() {
    document.querySelectorAll(".focused-reading-element").forEach((el) => {
      el.classList.remove("focused-reading-element");
    });
    document.querySelectorAll(".faded-reading-element").forEach((el) => {
      el.classList.remove("faded-reading-element");
    });
    const indicator = document.getElementById("reword-focus-indicator");
    if (indicator) {
      indicator.classList.remove("visible");
    }
  }

  function applyFocusFading(target) {
    document.querySelectorAll(".faded-reading-element").forEach((el) => {
      el.classList.remove("faded-reading-element");
    });
    document.querySelectorAll(".focused-reading-element").forEach((el) => {
      el.classList.remove("focused-reading-element");
    });

    if (!target) return;

    target.classList.add("focused-reading-element");

    const candidates = Array.from(
      document.querySelectorAll(
        "#left-column .reword-paragraph, #left-column li, #left-column .theorem, #left-column .lemma, #left-column .definition, #left-column .proposition, #left-column .corollary, #left-column .dim, #left-column .fact, #left-column .inline-image-crop, #left-column .box-text",
      ),
    );

    candidates.forEach((el) => {
      if (el === target) return;

      const position = target.compareDocumentPosition(el);
      if (position & Node.DOCUMENT_POSITION_FOLLOWING) {
        if (el.contains(target)) return;
        el.classList.add("faded-reading-element");
      }
    });
  }

  const leftColumn = document.getElementById("left-column");
  if (leftColumn) {
    leftColumn.addEventListener("click", (e) => {
      if (!body.classList.contains("rewords-focus")) return;

      // Don't intercept clicks on links
      if (e.target.closest("a")) return;

      const target = e.target.closest(".reword-paragraph, #left-column li");
      if (!target) return;

      e.stopPropagation();

      if (target.classList.contains("focused-reading-element")) {
        return;
      }
      applyFocusFading(target);
      updateIndicatorPosition();
    });
  }

  window.addEventListener("resize", updateIndicatorPosition);

  // Keyboard Navigation for Reword Focus
  window.addEventListener("keydown", (e) => {
    if (!body.classList.contains("rewords-focus")) return;

    if (e.key === "Escape") {
      e.preventDefault();
      body.classList.remove("rewords-focus");
      localStorage.setItem("rewords-focus", "false");
      const toggleRewordsFocus = document.getElementById(
        "toggle-rewords-focus",
      );
      if (toggleRewordsFocus) {
        toggleRewordsFocus.classList.remove("active");
      }
      clearRewordsFocus();
      return;
    }

    if (e.key !== "ArrowUp" && e.key !== "ArrowDown") return;

    // Ignore if typing in inputs/textareas
    if (
      document.activeElement &&
      (document.activeElement.tagName === "INPUT" ||
        document.activeElement.tagName === "TEXTAREA" ||
        document.activeElement.isContentEditable)
    ) {
      return;
    }

    const readingElements = Array.from(
      document.querySelectorAll(
        "#left-column .reword-paragraph, #left-column li",
      ),
    );
    if (readingElements.length === 0) return;

    e.preventDefault();

    const focused = document.querySelector(".focused-reading-element");
    let nextIndex = 0;

    if (focused) {
      const currentIndex = readingElements.indexOf(focused);
      if (e.key === "ArrowUp") {
        nextIndex = currentIndex - 1;
      } else {
        nextIndex = currentIndex + 1;
      }
    } else {
      if (e.key === "ArrowUp") {
        nextIndex = readingElements.length - 1;
      } else {
        nextIndex = 0;
      }
    }

    // Boundary check
    if (nextIndex >= 0 && nextIndex < readingElements.length) {
      const newTarget = readingElements[nextIndex];

      // Open parent details if closed
      let parent = newTarget.parentElement;
      while (parent) {
        if (parent.tagName === "DETAILS" && !parent.open) {
          parent.open = true;
        }
        parent = parent.parentElement;
      }

      applyFocusFading(newTarget);
      updateIndicatorPosition();
      newTarget.scrollIntoView({ block: "center", behavior: "smooth" });
    }
  });

  // Listen for system theme changes if set to system
  window
    .matchMedia("(prefers-color-scheme: dark)")
    .addEventListener("change", () => {
      if (localStorage.getItem("theme-pref") === "system") {
        applyTheme("system");
      }
    });

  // Highlight All Boxes Logic
  const toggleBoxes = document.getElementById("toggle-boxes");
  let showAllBoxes = localStorage.getItem("show-all-boxes") === "true";

  function updateAllBoxesHighlights() {
    if (!allBoxesOverlay || !imgEl) return;
    allBoxesOverlay.innerHTML = "";
    if (!showAllBoxes) {
      allBoxesOverlay.classList.add("hidden");
      return;
    }
    allBoxesOverlay.classList.remove("hidden");

    const currentSrc = imgEl.src;
    if (!currentSrc) return;

    const rectW = imgEl.clientWidth;
    const rectH = imgEl.clientHeight;
    if (rectW === 0 || rectH === 0) return;

    const boxes = document.querySelectorAll(".box-text, .inline-image-crop");
    boxes.forEach((box) => {
      const boxImg =
        box.getAttribute("data-img") || box.getAttribute("data-src");
      if (boxImg && currentSrc.endsWith(boxImg)) {
        const t = parseFloat(box.getAttribute("data-top"));
        const r = parseFloat(box.getAttribute("data-right"));
        const b = parseFloat(box.getAttribute("data-bottom"));
        const l = parseFloat(box.getAttribute("data-left"));

        if (!isNaN(t) && !isNaN(r) && !isNaN(b) && !isNaN(l)) {
          const hl = document.createElement("div");
          hl.className = "box-highlight";
          hl.style.top = (t / 1000) * rectH + "px";
          hl.style.left = (l / 1000) * rectW + "px";
          hl.style.width = ((r - l) / 1000) * rectW + "px";
          hl.style.height = ((b - t) / 1000) * rectH + "px";
          hl.style.pointerEvents = "auto";
          hl.style.cursor = "alias";

          hl.addEventListener("click", () => {
            let targetEl = box;
            const ocrOff = document.body.classList.contains(
              "transcription-hidden",
            );
            if (ocrOff && box.classList.contains("box-text")) {
              // Find the first reword referencing this box's ID
              const boxId = box.id;
              let reword = null;
              if (boxId) {
                reword = Array.from(document.querySelectorAll(".reword")).find(
                  (rw) => {
                    const refs = rw.getAttribute("data-ref");
                    return refs && refs.split(" ").includes(boxId);
                  },
                );
              }
              // If not found, look for the first reword inside the parent container of the box
              if (!reword) {
                const parent = box.parentElement;
                if (parent) {
                  reword = parent.querySelector(".reword");
                }
              }
              if (reword) {
                targetEl = reword;
              }
            }
            if (targetEl) {
              targetEl.scrollIntoView({ block: "center", behavior: "smooth" });
            }
          });

          allBoxesOverlay.appendChild(hl);
        }
      }
    });
  }

  if (toggleBoxes) {
    if (showAllBoxes) toggleBoxes.classList.add("active");
    updateAllBoxesHighlights();

    toggleBoxes.addEventListener("click", () => {
      showAllBoxes = !showAllBoxes;
      toggleBoxes.classList.toggle("active", showAllBoxes);
      localStorage.setItem("show-all-boxes", showAllBoxes);
      updateAllBoxesHighlights();
    });
  }

  if (imgEl) {
    imgEl.addEventListener("load", updateAllBoxesHighlights);

    imgEl.addEventListener("click", () => {
      // Don't open if we're in inspect mode
      if (body.classList.contains("inspect-mode")) return;

      const modalOverlay = document.createElement("div");
      modalOverlay.className = "image-modal-overlay";

      const modalContainer = document.createElement("div");
      modalContainer.className = "image-modal-container";

      const imgWrapper = document.createElement("div");
      imgWrapper.className = "image-modal-img-wrapper";

      const modalImg = document.createElement("img");
      modalImg.className = "image-modal-img";
      modalImg.src = imgEl.src;

      const modalInspectOverlay = document.createElement("div");
      modalInspectOverlay.className = "image-modal-inspector-overlay hidden";

      const modalInspectSelection = document.createElement("div");
      modalInspectSelection.className = "image-modal-inspector-selection";

      modalInspectOverlay.appendChild(modalInspectSelection);
      imgWrapper.appendChild(modalImg);
      imgWrapper.appendChild(modalInspectOverlay);

      const closeBtn = document.createElement("button");
      closeBtn.className = "brutal-button icon modal-close-btn";
      closeBtn.innerHTML = `<i data-lucide="x" size="18"></i>`;
      closeBtn.title = "Close Image";

      const inspectBtn = document.createElement("button");
      inspectBtn.className = "brutal-button icon modal-inspect-btn";
      inspectBtn.innerHTML = `<i data-lucide="crosshair" size="18"></i>`;
      inspectBtn.title = "Inspect Region";

      const actionsBar = document.createElement("div");
      actionsBar.className = "image-modal-actions-bar";
      actionsBar.appendChild(closeBtn);
      actionsBar.appendChild(inspectBtn);

      const contentWrapper = document.createElement("div");
      contentWrapper.className = "image-modal-content-wrapper";
      contentWrapper.appendChild(imgWrapper);
      contentWrapper.appendChild(actionsBar);

      modalContainer.appendChild(contentWrapper);
      modalOverlay.appendChild(modalContainer);
      document.body.appendChild(modalOverlay);

      // Initialize Lucide icons
      if (window.lucide) window.lucide.createIcons();

      // Prevent body scrolling while modal is open
      document.body.style.overflow = "hidden";

      const closeModal = () => {
        modalOverlay.remove();
        document.body.style.overflow = "";
      };

      modalOverlay.addEventListener("click", (e) => {
        if (e.target === modalOverlay || e.target === modalContainer) {
          closeModal();
        }
      });

      closeBtn.addEventListener("click", (e) => {
        e.stopPropagation();
        closeModal();
      });

      // Drag/Inspect Logic
      let modalIsDragging = false;
      let modalStartX = 0;
      let modalStartY = 0;

      inspectBtn.addEventListener("click", (e) => {
        e.stopPropagation();
        const active = modalInspectOverlay.classList.toggle("hidden");
        inspectBtn.classList.toggle("active", !active);
        if (active) {
          modalInspectSelection.style.display = "none";
        }
      });

      modalInspectOverlay.addEventListener("mousedown", (e) => {
        e.stopPropagation();
        modalIsDragging = true;
        const rect = modalInspectOverlay.getBoundingClientRect();
        modalStartX = e.clientX - rect.left;
        modalStartY = e.clientY - rect.top;

        modalInspectSelection.style.left = modalStartX + "px";
        modalInspectSelection.style.top = modalStartY + "px";
        modalInspectSelection.style.width = "0px";
        modalInspectSelection.style.height = "0px";
        modalInspectSelection.style.display = "block";
      });

      modalInspectOverlay.addEventListener("mousemove", (e) => {
        if (!modalIsDragging) return;
        e.stopPropagation();

        const rect = modalInspectOverlay.getBoundingClientRect();
        const currentX = Math.max(
          0,
          Math.min(rect.width, e.clientX - rect.left),
        );
        const currentY = Math.max(
          0,
          Math.min(rect.height, e.clientY - rect.top),
        );

        const x = Math.min(modalStartX, currentX);
        const y = Math.min(modalStartY, currentY);
        const w = Math.abs(currentX - modalStartX);
        const h = Math.abs(currentY - modalStartY);

        modalInspectSelection.style.left = x + "px";
        modalInspectSelection.style.top = y + "px";
        modalInspectSelection.style.width = w + "px";
        modalInspectSelection.style.height = h + "px";
      });

      modalInspectOverlay.addEventListener("mouseup", (e) => {
        if (!modalIsDragging) return;
        e.stopPropagation();
        modalIsDragging = false;

        const rect = modalInspectOverlay.getBoundingClientRect();
        const selRect = modalInspectSelection.getBoundingClientRect();

        const top = Math.round(((selRect.top - rect.top) / rect.height) * 1000);
        const left = Math.round(
          ((selRect.left - rect.left) / rect.width) * 1000,
        );
        const bottom = Math.round(
          ((selRect.bottom - rect.top) / rect.height) * 1000,
        );
        const right = Math.round(
          ((selRect.right - rect.left) / rect.width) * 1000,
        );

        const imgSrc = imgEl.src.split("/").pop();
        const tag = `<image src="${imgSrc}" top=${top} right=${right} bottom=${bottom} left=${left} />`;

        navigator.clipboard.writeText(tag).then(() => {
          const originalHTML = inspectBtn.innerHTML;
          inspectBtn.innerHTML = `<i data-lucide="check" size="18"></i>`;
          if (window.lucide) window.lucide.createIcons();

          setTimeout(() => {
            inspectBtn.innerHTML = originalHTML;
            if (window.lucide) window.lucide.createIcons();

            modalInspectOverlay.classList.add("hidden");
            inspectBtn.classList.remove("active");
            modalInspectSelection.style.display = "none";
          }, 500);
        });
      });

      // Add keyboard close (Escape)
      const escHandler = (e) => {
        if (e.key === "Escape") {
          closeModal();
          document.removeEventListener("keydown", escHandler);
        }
      };
      document.addEventListener("keydown", escHandler);
    });
  }
  window.addEventListener("resize", updateAllBoxesHighlights);

  // Add icon to content links
  document.querySelectorAll("#left-column .note-content a").forEach((a) => {
    const href = a.getAttribute("href") || "";
    if (!a.querySelector(".link-icon")) {
      const isExternal =
        href.startsWith("http://") || href.startsWith("https://");
      const iconName = isExternal ? "external-link" : "link";
      const icon = document.createElement("span");
      icon.className = "link-icon";
      icon.innerHTML = `<i data-lucide="${iconName}" size="12"></i>`;
      a.prepend(icon);
    }
  });

  // Open Source File button logic (if dev server /cwd endpoint is available)
  fetch("/cwd")
    .then((res) => {
      if (res.ok) {
        return res.text();
      }
      throw new Error("Not connected to dev server");
    })
    .then((cwd) => {
      const openSourceBtn = document.getElementById("open-source");
      if (openSourceBtn) {
        openSourceBtn.style.display = "inline-flex";
        if (window.lucide) window.lucide.createIcons();
        openSourceBtn.addEventListener("click", () => {
          const sourcePath =
            document.body.getAttribute("data-source-path") || "";
          const url = `antigravity-ide://file/${cwd.trim()}/${sourcePath}`;
          window.location.href = url;
        });
      }
    })
    .catch((err) => {
      if (window.DEBUG) {
        console.log("CWD endpoint not available:", err);
      }
    });

  // Initialize Lucide Icons
  if (window.lucide) {
    window.lucide.createIcons();
  }

  // Spoiler +/- Lucide Icon Management
  document.querySelectorAll(".spoiler").forEach((spoiler) => {
    const summary = spoiler.querySelector(".spoiler-summary");
    if (!summary) return;

    let iconContainer = summary.querySelector(".spoiler-icon");
    if (!iconContainer) {
      iconContainer = document.createElement("span");
      iconContainer.className = "spoiler-icon";
      summary.prepend(iconContainer);
    }

    const updateIcon = () => {
      const isOpen = spoiler.classList.contains("open");
      const iconName = isOpen ? "square-minus" : "square-plus";
      iconContainer.innerHTML = `<i data-lucide="${iconName}" size="14"></i>`;
      if (window.lucide) window.lucide.createIcons();
    };

    updateIcon();
    
    summary.addEventListener("click", () => {
      spoiler.classList.toggle("open");
      updateIcon();
    });
  });

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
          const target = getBoxTarget(id);
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

        if (foundAny && !e.target.closest(".box-text, .inline-image-crop")) {
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
            const target = getBoxTarget(id);
            if (target) {
              target.classList.remove("box-highlight-green");
            }
          });
        }
        clearHighlight();
      }
    }
  });

  function scrollToHash(behavior = "auto") {
    const hash = window.location.hash;
    if (hash) {
      try {
        const target = document.querySelector(hash);
        if (target) {
          target.scrollIntoView({ block: "center", behavior: behavior });
        }
      } catch (err) {
        console.error("Failed to scroll to hash:", err);
      }
    }
  }

  // MathJax Autorender
  function renderMath() {
    if (window.MathJax && window.MathJax.typesetPromise) {
      const mathElements = document.querySelectorAll(
        "span.math:not([data-rendered])",
      );
      if (mathElements.length === 0) {
        scrollToHash("auto");
        return;
      }

      mathElements.forEach((el) => {
        const display = el.getAttribute("data-display") === "true";
        const content = el.textContent;
        el.textContent = display ? `\\[${content}\\]` : `\\(${content}\\)`;
        el.dataset.rendered = "true";
      });

      window.MathJax.typesetPromise().then(() => {
        scrollToHash("auto");
      });
    } else {
      // If MathJax is not yet available, try again shortly
      setTimeout(renderMath, 100);
    }
  }

  // Call renderMath on load
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
      renderMath();
      scrollToHash("auto");
    });
  } else {
    renderMath();
    scrollToHash("auto");
  }

  window.addEventListener("hashchange", () => {
    scrollToHash("smooth");
  });

  window.addEventListener("load", () => {
    scrollToHash("auto");
  });

  // Inline Image Crops
  // Sizing configuration constants:
  // - MIN_RESCALE_WIDTH: Relative width (0.0 to 1.0) of container when crop width is 0% of page.
  // - MAX_RESCALE_WIDTH: Relative width (0.0 to 1.0) of container when crop width is 100% of page.
  const MIN_RESCALE_WIDTH = 0.0;
  const MAX_RESCALE_WIDTH = 1.5;

  document.querySelectorAll(".inline-image-crop").forEach((container) => {
    const src = container.getAttribute("data-src");
    const right = parseFloat(container.getAttribute("data-right"));
    const left = parseFloat(container.getAttribute("data-left"));
    const bottom = parseFloat(container.getAttribute("data-bottom"));
    const top = parseFloat(container.getAttribute("data-top"));

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
      container.style.setProperty("--crop-ratio", cropRatio);
      container.style.width = finalWidthPercent + "%";
      container.innerHTML = "";
      container.appendChild(croppedImg);
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

      const ocrOff = body.classList.contains("transcription-hidden");
      const selector = ocrOff ? ".reword" : ".box-text, .reword";
      const elements = document.querySelectorAll(selector);
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
            const target = getBoxTarget(firstId);
            if (target) {
              imgSrc =
                target.getAttribute("data-img") || target.getAttribute("src");
            }
          }
        }

        // Compare normalized href if src is a full URL
        if (imgSrc && !imgEl.src.endsWith(imgSrc)) {
          clearHighlight();
          if (allBoxesOverlay) allBoxesOverlay.innerHTML = "";
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
