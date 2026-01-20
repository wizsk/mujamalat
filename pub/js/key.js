// key.js

document.addEventListener("keydown", (e) => {
  // console.log(e.ctrlKey, e.shiftKey);

  if (e.code && e.code == "Escape") {
    try {
      if (document.activeElement === input) {
        e.preventDefault();
        input.blur();
        return;
      }
    } catch (e) {}

    // maybe defined or not who knows
    try {
      if (document.activeElement == customAfterInput) {
        e.preventDefault();
        customAfterInput.blur();
        return;
      }
    } catch (e) {}

    try {
      if (document.activeElement == noteTxtAr) {
        e.preventDefault();
        noteTxtAr.blur();
        return;
      }
    } catch (e) {}

    try {
      if (document.activeElement == customAfterInput) {
        e.preventDefault();
        customAfterInput.blur();
        return;
      }
    } catch (e) {}

    return;
  }

  if (!e.ctrlKey || e.shiftKey || e.altKey) return;

  switch (e.code) {
    case "KeyS":
      e.preventDefault();
      input.focus();
      input.select();
      break;
    case "KeyI":
      e.preventDefault();
      input.focus();
      input.setSelectionRange(input.value.length, input.value.length);
      break;

    case "Equal":
      e.preventDefault();
      fontSizeInc();
      break;
    case "Minus":
      e.preventDefault();
      fontSizeDec();
      break;
    case "Digit0":
      e.preventDefault();
      resetFontSize();
      break;

    case "KeyU":
      e.preventDefault();
      scroolToTop();
      break;

    case "KeyB":
      e.preventDefault();
      changeColor(!document.documentElement.classList.contains("dark"));
      break;

    case "KeyR":
      e.preventDefault();
      window.location.href = "/rd/";
      break;

    case "KeyM":
      e.preventDefault();
      try {
        if (readerMenu.open) {
          window.history.back();
        } else {
          readerMenuBtn.click();
        }
      } catch (e) {}
      break;

    case "KeyY":
      e.preventDefault();
      try {
        dictHighHiBtn.click();
      } catch (e) {}
      break;

    case "KeyO":
      e.preventDefault();
      try {
        if (!infoDialog.open) dictHighNoteBtn.click();
      } catch (e) {}
      break;
  }

  // ---- nav ----
  try {
    if (!nav.checkVisibility()) return;
  } catch (e) {
    return;
  }

  switch (e.code) {
    case "KeyH":
    case "ArrowDown":
      e.preventDefault();
      querySelectorNextPre(true, "sw-dict-item", "sw-dict-item-selected", scrollToEl);
      break;

    case "KeyL":
    case "ArrowUp":
      e.preventDefault();
      querySelectorNextPre(false, "sw-dict-item", "sw-dict-item-selected", scrollToEl);
      break;

    case "KeyJ":
      e.preventDefault();
      querySelectorNextPre(
        true,
        "querySelector-item",
        "querySelector-item-selected",
      );
      break;

    case "KeyK":
      e.preventDefault();
      querySelectorNextPre(
        false,
        "querySelector-item",
        "querySelector-item-selected",
      );
      break;
  }
});

/**
 * @param {boolean} next if true then goes to the next query otherwise preveious
 * @param {string}  className classname for the buttons
 * @param {string}  id is the button wich is selected for now
 */
function querySelectorNextPre(next, className, id, callb) {
  /** @type {NodeListOf<HTMLElement>} */
  const children = document.getElementsByClassName(className);
  if (children.length < 2) return;

  let curr = -1;
  for (let i = 0; i < children.length; i++) {
    if (children[i].id === id) {
      curr = i;
      break;
    }
  }

  if (curr < 0) return;

  if (next) curr = curr + 1 === children.length ? 0 : curr + 1;
  else curr = curr - 1 < 0 ? children.length - 1 : curr - 1;

  children[curr].click();

  if (callb) callb(children[curr]);
}
